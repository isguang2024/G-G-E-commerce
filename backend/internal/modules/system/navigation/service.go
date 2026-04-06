package navigation

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	menupkg "github.com/gg-ecommerce/backend/internal/modules/system/menu"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	pagepkg "github.com/gg-ecommerce/backend/internal/modules/system/page"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

type Manifest struct {
	CurrentApp      *apppkg.CurrentResponse   `json:"current_app"`
	CurrentSpace    *spacepkg.CurrentResponse `json:"current_space"`
	Context         gin.H                     `json:"context"`
	MenuTree        []gin.H                   `json:"menu_tree"`
	EntryRoutes     []gin.H                   `json:"entry_routes"`
	ManagedPages    []gin.H                   `json:"managed_pages"`
	VersionStamp    string                    `json:"version_stamp"`
	LegacyMenusTree []gin.H                   `json:"legacy_menus_tree,omitempty"`
	LegacyPages     []gin.H                   `json:"legacy_managed_pages,omitempty"`
}

type Compiler interface {
	Compile(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*Manifest, error)
}

type menuTreeProvider interface {
	GetTree(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error)
}

type managedPageProvider interface {
	ResolveCompiledAccessContext(appKey, spaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*pagepkg.CompiledAccessContext, error)
	ListRuntimeWithAccess(appKey, spaceKey string, accessCtx *pagepkg.CompiledAccessContext) ([]pagepkg.Record, error)
}

type currentSpaceProvider interface {
	GetCurrent(appKey string, host string, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*spacepkg.CurrentResponse, error)
}

type service struct {
	db          *gorm.DB
	appService  apppkg.Service
	menuService menuTreeProvider
	pageService managedPageProvider
	spaceSvc    currentSpaceProvider
}

func NewService(db *gorm.DB, appService apppkg.Service, menuService menupkg.MenuService, pageService pagepkg.Service, spaceSvc spacepkg.Service) Compiler {
	return &service{
		db:          db,
		appService:  appService,
		menuService: menuService,
		pageService: pageService,
		spaceSvc:    spaceSvc,
	}
}

func (s *service) Compile(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*Manifest, error) {
	normalizedAppKey := apppkg.NormalizeAppKey(appKey)
	var currentApp *apppkg.CurrentResponse
	var err error
	if s.appService != nil {
		currentApp, err = s.appService.GetCurrent(host, normalizedAppKey)
		if err != nil {
			return nil, err
		}
	}
	current, err := s.spaceSvc.GetCurrent(normalizedAppKey, host, requestedSpaceKey, userID, tenantID)
	if err != nil {
		return nil, err
	}
	resolvedSpaceKey := models.DefaultMenuSpaceKey
	if current != nil {
		resolvedSpaceKey = spacepkg.NormalizeSpaceKey(current.Space.SpaceKey)
	}

	accessCtx, err := s.pageService.ResolveCompiledAccessContext(normalizedAppKey, resolvedSpaceKey, userID, tenantID)
	if err != nil {
		return nil, err
	}

	visibleMenuIDs := accessCtx.VisibleMenuIDList()
	menuTree, err := s.menuService.GetTree(false, visibleMenuIDs, normalizedAppKey, resolvedSpaceKey)
	if err != nil {
		return nil, err
	}
	managedPages, err := s.pageService.ListRuntimeWithAccess(normalizedAppKey, resolvedSpaceKey, accessCtx)
	if err != nil {
		return nil, err
	}

	manifest := &Manifest{
		CurrentApp:   currentApp,
		CurrentSpace: current,
		Context: gin.H{
			"app_key":             normalizedAppKey,
			"space_key":           resolvedSpaceKey,
			"request_host":        strings.TrimSpace(host),
			"requested_space_key": strings.TrimSpace(requestedSpaceKey),
			"authenticated":       accessCtx != nil && accessCtx.Authenticated,
			"super_admin":         accessCtx != nil && accessCtx.SuperAdmin,
			"visible_menu_count":  len(visibleMenuIDs),
			"managed_page_count":  len(managedPages),
			"action_key_count":    accessContextActionKeyCount(accessCtx),
		},
		// MenuTree / ManagedPages 都来自同一份已编译访问上下文，前端不需要再重复做显隐裁剪。
		MenuTree:     menupkg.BuildRuntimeTreeMaps(menuTree),
		EntryRoutes:  extractEntryRouteMaps(menuTree),
		ManagedPages: pagepkg.BuildRuntimePageMaps(managedPages),
		VersionStamp: s.buildVersionStamp(normalizedAppKey, resolvedSpaceKey, len(menuTree), len(managedPages)),
	}
	if userID != nil {
		manifest.Context["user_id"] = userID.String()
	}
	if tenantID != nil {
		manifest.Context["collaboration_workspace_id"] = tenantID.String()
	}
	return manifest, nil
}

func accessContextActionKeyCount(ctx *pagepkg.CompiledAccessContext) int {
	if ctx == nil {
		return 0
	}
	return len(ctx.ActionKeys)
}

func extractEntryRouteMaps(tree []*user.Menu) []gin.H {
	if len(tree) == 0 {
		return []gin.H{}
	}
	result := make([]gin.H, 0)
	var walk func(nodes []*user.Menu)
	walk = func(nodes []*user.Menu) {
		for _, node := range nodes {
			if node == nil {
				continue
			}
			if strings.TrimSpace(node.Kind) == models.MenuKindEntry {
				result = append(result, menupkg.BuildRuntimeTreeMaps([]*user.Menu{node})...)
			}
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}
	walk(tree)
	return result
}

func (s *service) buildVersionStamp(appKey, spaceKey string, menuCount, managedPageCount int) string {
	if s == nil || s.db == nil {
		return fmt.Sprintf("%s:%s:%d:%d", appKey, spaceKey, menuCount, managedPageCount)
	}
	type stampRow struct {
		UpdatedAt *time.Time
	}

	candidates := make([]time.Time, 0, 4)
	loadMax := func(model interface{}, query string, args ...interface{}) {
		row := stampRow{}
		db := s.db.Model(model)
		if strings.TrimSpace(query) != "" {
			db = db.Where(query, args...)
		}
		if err := db.Select("MAX(updated_at) AS updated_at").Scan(&row).Error; err == nil && row.UpdatedAt != nil {
			candidates = append(candidates, row.UpdatedAt.UTC())
		}
	}

	loadMax(&models.App{}, "app_key = ?", appKey)
	loadMax(&models.MenuDefinition{}, "app_key = ?", appKey)
	loadMax(&models.SpaceMenuPlacement{}, "app_key = ? AND space_key = ?", appKey, spaceKey)
	loadMax(&models.UIPage{}, "app_key = ?", appKey)
	loadMax(&models.PageSpaceBinding{}, "app_key = ?", appKey)
	loadMax(&models.MenuSpace{}, "app_key = ? AND space_key = ?", appKey, spaceKey)
	loadMax(&models.AppHostBinding{}, "app_key = ? AND default_space_key = ?", appKey, spaceKey)

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].After(candidates[j])
	})
	if len(candidates) == 0 {
		return fmt.Sprintf("%s:%s:%d:%d", appKey, spaceKey, menuCount, managedPageCount)
	}
	return fmt.Sprintf("%s:%s:%d:%d:%d", appKey, spaceKey, menuCount, managedPageCount, candidates[0].Unix())
}
