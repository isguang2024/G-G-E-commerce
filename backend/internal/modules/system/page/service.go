package page

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	spaceutil "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
)

var (
	ErrPageNotFound        = errors.New("page not found")
	ErrPageKeyExists       = errors.New("page key already exists")
	ErrRouteNameExists     = errors.New("route name already exists")
	ErrParentMenuInvalid   = errors.New("parent menu invalid")
	ErrParentPageInvalid   = errors.New("parent page invalid")
	ErrDisplayGroupInvalid = errors.New("display group invalid")
	ErrPageHasChildren     = errors.New("page has children")
	ErrPageValidation      = errors.New("page validation failed")
)

type ListRequest struct {
	Current      int
	Size         int
	Keyword      string
	AppKey       string `form:"app_key"`
	SpaceKey     string `form:"space_key"`
	PageType     string
	ModuleKey    string
	ParentMenuID string
	AccessMode   string
	Source       string
	Status       string
}

type SaveRequest struct {
	AppKey            string                 `json:"app_key"`
	PageKey           string                 `json:"page_key"`
	Name              string                 `json:"name"`
	RouteName         string                 `json:"route_name"`
	RoutePath         string                 `json:"route_path"`
	Component         string                 `json:"component"`
	SpaceKey          string                 `json:"space_key"`
	SpaceKeys         []string               `json:"space_keys"`
	PageType          string                 `json:"page_type"`
	Source            string                 `json:"source"`
	ModuleKey         string                 `json:"module_key"`
	SortOrder         int                    `json:"sort_order"`
	ParentMenuID      string                 `json:"parent_menu_id"`
	ParentPageKey     string                 `json:"parent_page_key"`
	DisplayGroupKey   string                 `json:"display_group_key"`
	ActiveMenuPath    string                 `json:"active_menu_path"`
	BreadcrumbMode    string                 `json:"breadcrumb_mode"`
	AccessMode        string                 `json:"access_mode"`
	PermissionKey     string                 `json:"permission_key"`
	InheritPermission *bool                  `json:"inherit_permission"`
	KeepAlive         *bool                  `json:"keep_alive"`
	IsFullPage        *bool                  `json:"is_full_page"`
	Status            string                 `json:"status"`
	Meta              map[string]interface{} `json:"meta"`
}

type Record struct {
	models.UIPage
	ParentMenuName   string `json:"parent_menu_name"`
	ParentPageName   string `json:"parent_page_name"`
	DisplayGroupName string `json:"display_group_name"`
}

type MenuOption struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Title    string       `json:"title"`
	Path     string       `json:"path"`
	Children []MenuOption `json:"children,omitempty"`
}

type UnregisteredRecord struct {
	FilePath       string `json:"file_path"`
	Component      string `json:"component"`
	PageKey        string `json:"page_key"`
	Name           string `json:"name"`
	RouteName      string `json:"route_name"`
	RoutePath      string `json:"route_path"`
	PageType       string `json:"page_type"`
	ModuleKey      string `json:"module_key"`
	ParentMenuID   string `json:"parent_menu_id"`
	ParentMenuName string `json:"parent_menu_name"`
	ActiveMenuPath string `json:"active_menu_path"`
}

type SyncResult struct {
	CreatedCount int      `json:"created_count"`
	SkippedCount int      `json:"skipped_count"`
	CreatedKeys  []string `json:"created_keys"`
}

type runtimeMenuNode struct {
	Menu     models.Menu
	FullPath string
}

type scannedViewPage struct {
	FilePath  string
	Component string
}

type BreadcrumbPreviewItem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Path    string `json:"path"`
	PageKey string `json:"page_key,omitempty"`
}

type AccessTraceRequest struct {
	AppKey    string `form:"app_key"`
	UserID    string `form:"user_id"`
	TenantID  string `form:"tenant_id"`
	PageKey   string `form:"page_key"`
	SpaceKey  string `form:"space_key"`
	PageKeys  string `form:"page_keys"`
	RoutePath string `form:"route_path"`
}

type AccessTraceRoleItem struct {
	RoleID   string `json:"role_id"`
	RoleCode string `json:"role_code"`
	RoleName string `json:"role_name"`
	Status   string `json:"status"`
}

type AccessTracePageItem struct {
	PageKey          string   `json:"page_key"`
	PageName         string   `json:"page_name"`
	RoutePath        string   `json:"route_path"`
	AccessMode       string   `json:"access_mode"`
	PermissionKey    string   `json:"permission_key"`
	ParentPageKey    string   `json:"parent_page_key"`
	ParentMenuID     string   `json:"parent_menu_id"`
	ActiveMenuPath   string   `json:"active_menu_path"`
	Visible          bool     `json:"visible"`
	Reason           string   `json:"reason"`
	MatchedActionKey string   `json:"matched_action_key"`
	EffectiveChain   []string `json:"effective_chain"`
}

type AccessTraceResult struct {
	UserID         string                `json:"user_id"`
	TenantID       string                `json:"tenant_id,omitempty"`
	SpaceKey       string                `json:"space_key"`
	Authenticated  bool                  `json:"authenticated"`
	SuperAdmin     bool                  `json:"super_admin"`
	ActionKeyCount int                   `json:"action_key_count"`
	VisibleMenuIDs []string              `json:"visible_menu_ids"`
	Roles          []AccessTraceRoleItem `json:"roles"`
	Pages          []AccessTracePageItem `json:"pages"`
}

type Service interface {
	List(req *ListRequest) ([]Record, int64, error)
	ListOptions(appKey, spaceKey string) ([]models.UIPage, error)
	ListRuntime(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) ([]Record, error)
	ListRuntimePublic(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) ([]Record, error)
	ResolveCompiledAccessContext(appKey, spaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*CompiledAccessContext, error)
	GetAccessTrace(appKey string, req *AccessTraceRequest) (*AccessTraceResult, error)
	ListRuntimeWithAccess(appKey, spaceKey string, accessCtx *CompiledAccessContext) ([]Record, error)
	ListUnregistered(appKey string) ([]UnregisteredRecord, error)
	Sync(appKey string) (*SyncResult, error)
	PreviewBreadcrumb(id uuid.UUID, appKey string) ([]BreadcrumbPreviewItem, error)
	Get(id uuid.UUID, appKey string) (*Record, error)
	Create(req *SaveRequest) (*Record, error)
	Update(id uuid.UUID, req *SaveRequest) (*Record, error)
	Delete(id uuid.UUID, appKey string) error
	ListMenuOptions(appKey, spaceKey string) ([]MenuOption, error)
}

type service struct {
	db       *gorm.DB
	menuRepo user.MenuRepository
}

func NewService(db *gorm.DB, menuRepo user.MenuRepository) Service {
	return &service{db: db, menuRepo: menuRepo}
}

func (s *service) List(req *ListRequest) ([]Record, int64, error) {
	appKey := strings.TrimSpace(req.AppKey)
	if appKey == "" {
		return nil, 0, fmt.Errorf("%w: app_key is required", ErrPageValidation)
	}
	appKey = normalizeAppKey(appKey)
	query := s.db.Model(&models.UIPage{}).Where("app_key = ?", appKey)
	if req != nil {
		if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where(
				"name ILIKE ? OR page_key ILIKE ? OR route_name ILIKE ? OR route_path ILIKE ? OR component ILIKE ? OR module_key ILIKE ? OR display_group_key ILIKE ?",
				like, like, like, like, like, like, like,
			)
		}
		if pageType := normalizePageType(req.PageType); pageType != "" {
			query = query.Where("page_type = ?", pageType)
		}
		if moduleKey := strings.TrimSpace(req.ModuleKey); moduleKey != "" {
			query = query.Where("module_key = ?", moduleKey)
		}
		if accessMode := normalizeAccessMode(req.AccessMode); accessMode != "" {
			query = query.Where("access_mode = ?", accessMode)
		}
		if source := normalizeSource(req.Source); source != "" {
			query = query.Where("source = ?", source)
		}
		if status := normalizeStatus(req.Status); status != "" {
			query = query.Where("status = ?", status)
		}
		if parentMenuID, err := parseOptionalUUID(req.ParentMenuID); err == nil && parentMenuID != nil {
			query = query.Where("parent_menu_id = ?", *parentMenuID)
		}
	}

	var items []models.UIPage
	if err := query.Order("sort_order ASC, created_at ASC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	records, err := s.decorateRecords(items)
	if err != nil {
		return nil, 0, err
	}
	menuMap, err := s.loadMenuMap(appKey)
	if err != nil {
		return nil, 0, err
	}
	pageMap, err := s.loadPageMap(appKey)
	if err != nil {
		return nil, 0, err
	}
	bindingMap, err := loadPageSpaceBindingMap(s.db, appKey)
	if err != nil {
		return nil, 0, err
	}
	records = s.applyManagedPageModel(records, req.SpaceKey, menuMap, pageMap, bindingMap)
	total := int64(len(records))
	current, size := normalizePageAndSize(req)
	start := (current - 1) * size
	if start >= len(records) {
		return []Record{}, total, nil
	}
	end := start + size
	if end > len(records) {
		end = len(records)
	}
	return records[start:end], total, nil
}

func (s *service) ListOptions(appKey, spaceKey string) ([]models.UIPage, error) {
	if strings.TrimSpace(appKey) == "" {
		return nil, fmt.Errorf("%w: app_key is required", ErrPageValidation)
	}
	items := make([]models.UIPage, 0)
	query := s.db.Model(&models.UIPage{}).Where("app_key = ?", normalizeAppKey(appKey))
	err := query.
		Select(
			"id",
			"page_key",
			"name",
			"route_name",
			"route_path",
			"component",
			"space_key",
			"page_type",
			"module_key",
			"parent_menu_id",
			"parent_page_key",
			"display_group_key",
			"status",
		).
		Order("sort_order ASC, created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	records, err := s.decorateRecords(items)
	if err != nil {
		return nil, err
	}
	menuMap, err := s.loadMenuMap(normalizeAppKey(appKey))
	if err != nil {
		return nil, err
	}
	pageMap, err := s.loadPageMap(normalizeAppKey(appKey))
	if err != nil {
		return nil, err
	}
	bindingMap, err := loadPageSpaceBindingMap(s.db, normalizeAppKey(appKey))
	if err != nil {
		return nil, err
	}
	records = s.applyManagedPageModel(records, spaceKey, menuMap, pageMap, bindingMap)
	result := make([]models.UIPage, 0, len(records))
	for _, item := range records {
		result = append(result, item.UIPage)
	}
	return result, nil
}

func (s *service) ListRuntime(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) ([]Record, error) {
	return s.loadRuntimeRecords(normalizeAppKey(appKey), host, requestedSpaceKey, userID, tenantID)
}

func (s *service) ListRuntimePublic(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) ([]Record, error) {
	return s.loadPublicRuntimeRecords(normalizeAppKey(appKey), host, requestedSpaceKey, userID, tenantID)
}

func (s *service) ResolveCompiledAccessContext(appKey, spaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*CompiledAccessContext, error) {
	return s.buildCompiledAccessContextForSpace(normalizeAppKey(appKey), spaceKey, userID, tenantID)
}

func (s *service) GetAccessTrace(appKey string, req *AccessTraceRequest) (*AccessTraceResult, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is required", ErrPageValidation)
	}
	reqAppKey := strings.TrimSpace(req.AppKey)
	if reqAppKey == "" {
		return nil, fmt.Errorf("%w: app_key is required", ErrPageValidation)
	}
	appKey = normalizeAppKey(reqAppKey)
	userID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		return nil, fmt.Errorf("%w: user_id is invalid", ErrPageValidation)
	}
	var tenantID *uuid.UUID
	if rawTenantID := strings.TrimSpace(req.TenantID); rawTenantID != "" {
		parsed, parseErr := uuid.Parse(rawTenantID)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: tenant_id is invalid", ErrPageValidation)
		}
		tenantID = &parsed
	}
	spaceKey := normalizeSpaceKey(req.SpaceKey)
	if spaceKey == "" {
		spaceKey = spaceutil.DefaultMenuSpaceKey
	}
	accessCtx, err := s.ResolveCompiledAccessContext(appKey, spaceKey, &userID, tenantID)
	if err != nil {
		return nil, err
	}
	records, err := s.ListRuntimeWithAccess(appKey, spaceKey, accessCtx)
	if err != nil {
		return nil, err
	}
	visiblePages := make(map[string]Record, len(records))
	for _, item := range records {
		visiblePages[strings.TrimSpace(item.PageKey)] = item
	}

	requestedPageKeys := resolveRequestedPageKeys(req, records)
	pageItems := make([]AccessTracePageItem, 0, len(requestedPageKeys))
	for _, pageKey := range requestedPageKeys {
		page, getErr := s.findPageByKey(pageKey, appKey)
		if getErr != nil {
			continue
		}
		pageItems = append(pageItems, buildAccessTracePageItem(page, visiblePages, accessCtx))
	}

	roleItems, err := s.loadAccessTraceRoles(userID, tenantID)
	if err != nil {
		return nil, err
	}
	visibleMenuIDs := accessCtx.VisibleMenuIDList()
	sort.Slice(visibleMenuIDs, func(i, j int) bool {
		return visibleMenuIDs[i].String() < visibleMenuIDs[j].String()
	})
	visibleMenuIDStrings := make([]string, 0, len(visibleMenuIDs))
	for _, id := range visibleMenuIDs {
		visibleMenuIDStrings = append(visibleMenuIDStrings, id.String())
	}

	result := &AccessTraceResult{
		UserID:         userID.String(),
		SpaceKey:       spaceKey,
		Authenticated:  accessCtx.Authenticated,
		SuperAdmin:     accessCtx.SuperAdmin,
		ActionKeyCount: len(accessCtx.ActionKeys),
		VisibleMenuIDs: visibleMenuIDStrings,
		Roles:          roleItems,
		Pages:          pageItems,
	}
	if tenantID != nil {
		result.TenantID = tenantID.String()
	}
	return result, nil
}

func resolveRequestedPageKeys(req *AccessTraceRequest, runtimeRecords []Record) []string {
	normalized := make([]string, 0)
	seen := map[string]struct{}{}
	appendKey := func(value string) {
		key := strings.TrimSpace(value)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	appendKey(req.PageKey)
	for _, item := range strings.Split(req.PageKeys, ",") {
		appendKey(item)
	}
	routePath := normalizeRoutePath(req.RoutePath)
	if routePath != "" {
		for _, item := range runtimeRecords {
			if normalizeRoutePath(item.RoutePath) == routePath {
				appendKey(item.PageKey)
			}
		}
	}
	if len(normalized) > 0 {
		return normalized
	}
	for _, item := range runtimeRecords {
		appendKey(item.PageKey)
	}
	return normalized
}

func buildAccessTracePageItem(page *models.UIPage, visiblePages map[string]Record, accessCtx *CompiledAccessContext) AccessTracePageItem {
	pageKey := strings.TrimSpace(page.PageKey)
	_, visible := visiblePages[pageKey]
	reason := "denied_or_not_in_runtime"
	if visible {
		reason = "visible_in_runtime"
	}
	matchedActionKey := ""
	if permissionKey := strings.TrimSpace(page.PermissionKey); permissionKey != "" {
		normalizedPermissionKey := permissionkey.Normalize(permissionKey)
		if accessCtx != nil {
			if _, ok := accessCtx.ActionKeys[normalizedPermissionKey]; ok {
				matchedActionKey = normalizedPermissionKey
			}
		}
	}
	return AccessTracePageItem{
		PageKey:          pageKey,
		PageName:         strings.TrimSpace(page.Name),
		RoutePath:        normalizeRoutePath(page.RoutePath),
		AccessMode:       normalizeAccessMode(page.AccessMode),
		PermissionKey:    strings.TrimSpace(page.PermissionKey),
		ParentPageKey:    strings.TrimSpace(page.ParentPageKey),
		ActiveMenuPath:   normalizeRoutePath(page.ActiveMenuPath),
		Visible:          visible,
		Reason:           reason,
		MatchedActionKey: matchedActionKey,
		EffectiveChain:   buildAccessTraceChain(page),
	}
}

func buildAccessTraceChain(page *models.UIPage) []string {
	chain := make([]string, 0, 3)
	if key := strings.TrimSpace(page.PageKey); key != "" {
		chain = append(chain, "page:"+key)
	}
	if key := strings.TrimSpace(page.ParentPageKey); key != "" {
		chain = append(chain, "parent_page:"+key)
	}
	if path := normalizeRoutePath(page.ActiveMenuPath); path != "" {
		chain = append(chain, "active_menu:"+path)
	}
	return chain
}

func (s *service) loadAccessTraceRoles(userID uuid.UUID, tenantID *uuid.UUID) ([]AccessTraceRoleItem, error) {
	roles := make([]models.Role, 0)
	if tenantID != nil {
		effective, err := s.loadRuntimeEffectiveActiveRoles(userID, *tenantID)
		if err != nil {
			return nil, err
		}
		roles = effective
	} else {
		if err := s.db.Model(&models.Role{}).
			Joins("JOIN user_roles ON user_roles.role_id = roles.id").
			Where("user_roles.user_id = ?", userID).
			Where("roles.status = ?", "normal").
			Distinct("roles.*").
			Find(&roles).Error; err != nil {
			return nil, err
		}
	}
	sort.Slice(roles, func(i, j int) bool {
		if roles[i].SortOrder == roles[j].SortOrder {
			return roles[i].Code < roles[j].Code
		}
		return roles[i].SortOrder < roles[j].SortOrder
	})
	items := make([]AccessTraceRoleItem, 0, len(roles))
	for _, role := range roles {
		items = append(items, AccessTraceRoleItem{
			RoleID:   role.ID.String(),
			RoleCode: strings.TrimSpace(role.Code),
			RoleName: strings.TrimSpace(role.Name),
			Status:   strings.TrimSpace(role.Status),
		})
	}
	return items, nil
}

func (s *service) ListRuntimeWithAccess(appKey, spaceKey string, accessCtx *CompiledAccessContext) ([]Record, error) {
	return s.loadRuntimeRecordsWithAccess(normalizeAppKey(appKey), spaceKey, accessCtx)
}

func (s *service) buildRuntimeRecords(appKey, spaceKey string) ([]Record, map[uuid.UUID]runtimeMenuNode, error) {
	var items []models.UIPage
	if err := s.db.Where("app_key = ?", normalizeAppKey(appKey)).Where("status = ? AND page_type <> ?", "normal", "display_group").
		Order("sort_order ASC, created_at ASC").
		Find(&items).Error; err != nil {
		return nil, nil, err
	}
	records, err := s.decorateRecords(items)
	if err != nil {
		return nil, nil, err
	}
	menuMap, err := s.loadMenuMap(normalizeAppKey(appKey))
	if err != nil {
		return nil, nil, err
	}
	pageMap, err := s.loadPageMap(normalizeAppKey(appKey))
	if err != nil {
		return nil, nil, err
	}
	bindingMap, err := loadPageSpaceBindingMap(s.db, normalizeAppKey(appKey))
	if err != nil {
		return nil, nil, err
	}
	filtered := s.applyManagedPageModel(records, spaceKey, menuMap, pageMap, bindingMap)
	return filtered, menuMap, nil
}

func (s *service) applyManagedPageModel(
	records []Record,
	spaceKey string,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
	bindingMap map[uuid.UUID][]string,
) []Record {
	filtered := make([]Record, 0, len(records))
	for index := range records {
		record := records[index]
		if isMenuBackedEntryPage(record.UIPage, menuMap) {
			continue
		}
		resolvedSpaceKeys := resolvePageSpaceKeys(record.UIPage, pageMap, menuMap, bindingMap, map[string]struct{}{})
		applyResolvedPageSpace(&record.UIPage, resolvedSpaceKeys)
		record.ActiveMenuPath = s.resolveActiveMenuPath(
			&record.UIPage,
			menuMap,
			pageMap,
			map[string]struct{}{},
		)
		if !isPageVisibleInSpace(record.UIPage, spaceKey) {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}

func (s *service) ListUnregistered(appKey string) ([]UnregisteredRecord, error) {
	return s.buildUnregisteredRecords(normalizeAppKey(appKey))
}

func (s *service) Sync(appKey string) (*SyncResult, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	items, err := s.buildUnregisteredRecords(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	result := &SyncResult{
		CreatedKeys: make([]string, 0, len(items)),
	}
	for _, item := range items {
		req := &SaveRequest{
			AppKey:            normalizedAppKey,
			PageKey:           item.PageKey,
			Name:              item.Name,
			RouteName:         item.RouteName,
			RoutePath:         item.RoutePath,
			Component:         item.Component,
			PageType:          item.PageType,
			Source:            "sync",
			ModuleKey:         item.ModuleKey,
			ParentMenuID:      item.ParentMenuID,
			ActiveMenuPath:    item.ActiveMenuPath,
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        deriveSyncedPageAccessMode(item.PageType),
			InheritPermission: boolPtr(item.PageType == "inner"),
			KeepAlive:         boolPtr(false),
			IsFullPage:        boolPtr(false),
			Status:            "normal",
			Meta:              map[string]interface{}{},
		}
		if _, err := s.Create(req); err != nil {
			return nil, err
		}
		result.CreatedCount++
		result.CreatedKeys = append(result.CreatedKeys, item.PageKey)
	}
	result.SkippedCount = len(items) - result.CreatedCount
	return result, nil
}

func (s *service) PreviewBreadcrumb(id uuid.UUID, appKey string) ([]BreadcrumbPreviewItem, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	page, err := s.findPageByID(id, normalizedAppKey)
	if err != nil {
		return nil, err
	}
	menuMap, err := s.loadMenuMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	pageMap, err := s.loadPageMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	chain, err := s.resolveBreadcrumbChain(page, menuMap, pageMap)
	if err != nil {
		return nil, err
	}
	result := make([]BreadcrumbPreviewItem, 0, len(chain)+1)
	result = append(result, chain...)
	result = append(result, BreadcrumbPreviewItem{
		Type:    "page",
		Title:   page.Name,
		Path:    strings.TrimSpace(page.RoutePath),
		PageKey: page.PageKey,
	})
	return result, nil
}

func (s *service) Get(id uuid.UUID, appKey string) (*Record, error) {
	var item models.UIPage
	normalizedAppKey := normalizeAppKey(appKey)
	if err := s.db.Where("id = ? AND app_key = ?", id, normalizedAppKey).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	records, err := s.decorateRecords([]models.UIPage{item})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, ErrPageNotFound
	}
	if err := s.hydrateManagedRecord(&records[0]); err != nil {
		return nil, err
	}
	return &records[0], nil
}

func (s *service) Create(req *SaveRequest) (*Record, error) {
	item, err := s.buildModel(nil, req)
	if err != nil {
		return nil, err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		return syncPageSpaceBindings(tx, item.AppKey, item.ID, req.SpaceKey, req.SpaceKeys, item.ParentMenuID, item.ParentPageKey)
	}); err != nil {
		return nil, err
	}
	InvalidateRuntimeCache()
	return s.Get(item.ID, item.AppKey)
}

func (s *service) Update(id uuid.UUID, req *SaveRequest) (*Record, error) {
	var existing models.UIPage
	normalizedAppKey := strings.TrimSpace(req.AppKey)
	if normalizedAppKey == "" {
		return nil, fmt.Errorf("%w: app_key is required", ErrPageValidation)
	}
	if err := s.db.Where("id = ? AND app_key = ?", id, normalizedAppKey).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	item, err := s.buildModel(&existing, req)
	if err != nil {
		return nil, err
	}
	item.ID = existing.ID
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if existing.PageKey != item.PageKey {
			if err := tx.Model(&models.UIPage{}).
				Where("app_key = ? AND parent_page_key = ?", existing.AppKey, existing.PageKey).
				Update("parent_page_key", item.PageKey).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.UIPage{}).
				Where("app_key = ? AND display_group_key = ?", existing.AppKey, existing.PageKey).
				Update("display_group_key", item.PageKey).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&existing).Updates(pageToUpdateMap(item)).Error; err != nil {
			return err
		}
		return syncPageSpaceBindings(tx, existing.AppKey, existing.ID, req.SpaceKey, req.SpaceKeys, item.ParentMenuID, item.ParentPageKey)
	}); err != nil {
		return nil, err
	}
	InvalidateRuntimeCache()
	return s.Get(id, existing.AppKey)
}

func (s *service) Delete(id uuid.UUID, appKey string) error {
	var existing models.UIPage
	normalizedAppKey := normalizeAppKey(appKey)
	if err := s.db.Where("id = ? AND app_key = ?", id, normalizedAppKey).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPageNotFound
		}
		return err
	}
	var childCount int64
	if err := s.db.Model(&models.UIPage{}).
		Where("app_key = ? AND (parent_page_key = ? OR display_group_key = ?)", existing.AppKey, existing.PageKey, existing.PageKey).
		Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return ErrPageHasChildren
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("app_key = ? AND page_id = ?", existing.AppKey, id).Delete(&models.PageSpaceBinding{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&models.UIPage{}, "id = ? AND app_key = ?", id, existing.AppKey)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrPageNotFound
		}
		return nil
	}); err != nil {
		return err
	}
	InvalidateRuntimeCache()
	return nil
}

func (s *service) ListMenuOptions(appKey, spaceKey string) ([]MenuOption, error) {
	menus, err := s.menuRepo.ListByAppAndSpace(normalizeAppKey(appKey), normalizeSpaceKey(spaceKey))
	if err != nil {
		return nil, err
	}

	childrenMap := make(map[string][]user.Menu)
	roots := make([]user.Menu, 0)
	for _, menu := range menus {
		if resolveMenuKind(menu) == models.MenuKindExternal {
			continue
		}
		if menu.ParentID == nil {
			roots = append(roots, menu)
			continue
		}
		childrenMap[menu.ParentID.String()] = append(childrenMap[menu.ParentID.String()], menu)
	}

	var build func(item user.Menu) MenuOption
	build = func(item user.Menu) MenuOption {
		children := childrenMap[item.ID.String()]
		node := MenuOption{
			ID:    item.ID.String(),
			Name:  item.Name,
			Title: item.Title,
			Path:  item.Path,
		}
		if len(children) > 0 {
			node.Children = make([]MenuOption, 0, len(children))
			for _, child := range children {
				node.Children = append(node.Children, build(child))
			}
		}
		return node
	}

	result := make([]MenuOption, 0, len(roots))
	for _, root := range roots {
		result = append(result, build(root))
	}
	return result, nil
}

func (s *service) buildModel(existing *models.UIPage, req *SaveRequest) (*models.UIPage, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: 请求体不能为空", ErrPageValidation)
	}
	appKey := strings.TrimSpace(req.AppKey)
	if appKey == "" {
		return nil, fmt.Errorf("%w: app_key is required", ErrPageValidation)
	}
	appKey = normalizeAppKey(appKey)

	pageType := normalizePageType(req.PageType)
	if pageType == "" {
		pageType = "inner"
	}
	source := normalizeSource(req.Source)
	if source == "" {
		source = "manual"
	}
	breadcrumbMode := normalizeBreadcrumbMode(req.BreadcrumbMode)
	if breadcrumbMode == "" {
		breadcrumbMode = "inherit_menu"
	}
	accessMode := normalizeAccessMode(req.AccessMode)
	if accessMode == "" {
		accessMode = "inherit"
	}
	status := normalizeStatus(req.Status)
	if status == "" {
		status = "normal"
	}

	pageKey := strings.TrimSpace(req.PageKey)
	name := strings.TrimSpace(req.Name)
	routeName := strings.TrimSpace(req.RouteName)
	routePath := strings.TrimSpace(req.RoutePath)
	component := strings.TrimSpace(req.Component)
	moduleKey := strings.TrimSpace(req.ModuleKey)
	parentPageKey := strings.TrimSpace(req.ParentPageKey)
	displayGroupKey := strings.TrimSpace(req.DisplayGroupKey)
	activeMenuPath := strings.TrimSpace(req.ActiveMenuPath)
	permissionKey := strings.TrimSpace(req.PermissionKey)

	if pageType == "group" && pageKey == "" {
		if existing != nil && strings.TrimSpace(existing.PageKey) != "" {
			pageKey = strings.TrimSpace(existing.PageKey)
		} else {
			pageKey = s.generateGroupPageKey()
		}
	}
	if pageType == "display_group" && pageKey == "" {
		if existing != nil && strings.TrimSpace(existing.PageKey) != "" {
			pageKey = strings.TrimSpace(existing.PageKey)
		} else {
			pageKey = s.generateDisplayGroupPageKey()
		}
	}
	if isRoutelessPageType(pageType) && routeName == "" {
		routeName = pageKey
	}
	if name == "" {
		return nil, fmt.Errorf("%w: 页面名称不能为空", ErrPageValidation)
	}
	if pageKey == "" {
		return nil, fmt.Errorf("%w: 页面标识不能为空", ErrPageValidation)
	}
	if !isRoutelessPageType(pageType) && (routeName == "" || routePath == "" || component == "") {
		return nil, fmt.Errorf("%w: 页面节点必须填写路由名称、路由路径和组件路径", ErrPageValidation)
	}
	if accessMode == "permission" && permissionKey == "" {
		return nil, fmt.Errorf("%w: 权限访问模式必须指定权限键", ErrPageValidation)
	}
	if parentPageKey != "" {
		displayGroupKey = ""
	}
	if pageType == "display_group" {
		parentPageKey = ""
		displayGroupKey = ""
		moduleKey = ""
		routePath = ""
		component = ""
		activeMenuPath = ""
		breadcrumbMode = "inherit_menu"
		accessMode = "inherit"
		permissionKey = ""
	}

	inheritPermission := true
	if existing != nil {
		inheritPermission = existing.InheritPermission
	}
	if req.InheritPermission != nil {
		inheritPermission = *req.InheritPermission
	}

	keepAlive := false
	if existing != nil {
		keepAlive = existing.KeepAlive
	}
	if req.KeepAlive != nil {
		keepAlive = *req.KeepAlive
	}

	isFullPage := false
	if existing != nil {
		isFullPage = existing.IsFullPage
	}
	if req.IsFullPage != nil {
		isFullPage = *req.IsFullPage
	}

	parentMenuID, err := parseOptionalUUID(req.ParentMenuID)
	if err != nil {
		return nil, ErrParentMenuInvalid
	}
	if pageType == "display_group" || pageType == "global" {
		parentMenuID = nil
	}
	if pageType == "global" {
		parentPageKey = ""
		activeMenuPath = ""
	}
	if parentMenuID != nil {
		parentMenu, err := s.menuRepo.GetByID(*parentMenuID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrParentMenuInvalid
			}
			return nil, err
		}
		if normalizeAppKey(parentMenu.AppKey) != appKey {
			return nil, ErrParentMenuInvalid
		}
	}

	if parentPageKey != "" {
		parentPage, err := s.findPageByKey(parentPageKey, appKey)
		if err != nil {
			return nil, err
		}
		if normalizePageType(parentPage.PageType) == "display_group" {
			return nil, ErrParentPageInvalid
		}
		if parentPage.PageKey == pageKey {
			return nil, fmt.Errorf("%w: 上级页面不能指向自身", ErrPageValidation)
		}
		if existing != nil && parentPage.ID == existing.ID {
			return nil, fmt.Errorf("%w: 上级页面不能指向自身", ErrPageValidation)
		}
		if existing != nil {
			if err := s.ensureNoPageCycle(pageKey, parentPage.PageKey, appKey); err != nil {
				return nil, err
			}
		}
	}
	if displayGroupKey != "" {
		displayGroup, err := s.findDisplayGroupByKey(displayGroupKey, appKey)
		if err != nil {
			return nil, err
		}
		if displayGroup.PageKey == pageKey {
			return nil, fmt.Errorf("%w: 普通分组不能指向自身", ErrPageValidation)
		}
	}

	checkID := uuid.Nil
	if existing != nil {
		checkID = existing.ID
	}
	if err := s.ensureUnique("page_key", pageKey, checkID, appKey); err != nil {
		return nil, err
	}
	if err := s.ensureUnique("route_name", routeName, checkID, appKey); err != nil {
		return nil, err
	}

	meta := models.MetaJSON(req.Meta)
	if meta == nil {
		meta = models.MetaJSON{}
	}

	item := &models.UIPage{
		AppKey:            appKey,
		PageKey:           pageKey,
		Name:              name,
		RouteName:         routeName,
		RoutePath:         routePath,
		Component:         component,
		SpaceKey:          normalizeStandalonePageSpaceKey(req.SpaceKey, parentMenuID, parentPageKey),
		PageType:          pageType,
		Source:            source,
		ModuleKey:         moduleKey,
		SortOrder:         req.SortOrder,
		ParentMenuID:      parentMenuID,
		ParentPageKey:     parentPageKey,
		DisplayGroupKey:   displayGroupKey,
		ActiveMenuPath:    activeMenuPath,
		BreadcrumbMode:    breadcrumbMode,
		AccessMode:        accessMode,
		PermissionKey:     permissionKey,
		InheritPermission: inheritPermission,
		KeepAlive:         keepAlive,
		IsFullPage:        isFullPage,
		Status:            status,
		Meta:              meta,
	}
	return item, nil
}

func (s *service) generateGroupPageKey() string {
	return "group." + uuid.NewString()
}

func (s *service) generateDisplayGroupPageKey() string {
	return "display-group." + uuid.NewString()
}

func (s *service) ensureUnique(field, value string, excludeID uuid.UUID, appKey string) error {
	query := s.db.Model(&models.UIPage{}).Where("app_key = ?", normalizeAppKey(appKey)).Where(field+" = ?", value)
	if excludeID != uuid.Nil {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	switch field {
	case "page_key":
		return ErrPageKeyExists
	case "route_name":
		return ErrRouteNameExists
	default:
		return ErrPageValidation
	}
}

func (s *service) decorateRecords(items []models.UIPage) ([]Record, error) {
	if len(items) == 0 {
		return []Record{}, nil
	}
	appKey := normalizeAppKey(items[0].AppKey)

	menuIDs := make([]uuid.UUID, 0, len(items))
	parentPageKeys := make([]string, 0, len(items))
	displayGroupKeys := make([]string, 0, len(items))
	menuSeen := map[uuid.UUID]struct{}{}
	pageSeen := map[string]struct{}{}
	displayGroupSeen := map[string]struct{}{}
	for _, rawItem := range items {
		item := normalizeLegacyGlobalPage(rawItem)
		if item.ParentMenuID != nil {
			if _, ok := menuSeen[*item.ParentMenuID]; !ok {
				menuSeen[*item.ParentMenuID] = struct{}{}
				menuIDs = append(menuIDs, *item.ParentMenuID)
			}
		}
		if key := strings.TrimSpace(item.ParentPageKey); key != "" {
			if _, ok := pageSeen[key]; !ok {
				pageSeen[key] = struct{}{}
				parentPageKeys = append(parentPageKeys, key)
			}
		}
		if key := strings.TrimSpace(item.DisplayGroupKey); key != "" {
			if _, ok := displayGroupSeen[key]; !ok {
				displayGroupSeen[key] = struct{}{}
				displayGroupKeys = append(displayGroupKeys, key)
			}
		}
	}

	menuNameMap := map[uuid.UUID]string{}
	if len(menuIDs) > 0 {
		menus, err := s.menuRepo.GetByIDs(menuIDs)
		if err != nil {
			return nil, err
		}
		for _, menu := range menus {
			menuNameMap[menu.ID] = menu.Name
		}
	}

	parentPageNameMap := map[string]string{}
	if len(parentPageKeys) > 0 {
		var pages []models.UIPage
		if err := s.db.Select("page_key", "name").Where("app_key = ? AND page_key IN ?", appKey, parentPageKeys).Find(&pages).Error; err != nil {
			return nil, err
		}
		for _, page := range pages {
			parentPageNameMap[page.PageKey] = page.Name
		}
	}

	displayGroupNameMap := map[string]string{}
	if len(displayGroupKeys) > 0 {
		var groups []models.UIPage
		if err := s.db.Select("page_key", "name").Where("app_key = ? AND page_key IN ?", appKey, displayGroupKeys).Find(&groups).Error; err != nil {
			return nil, err
		}
		for _, group := range groups {
			displayGroupNameMap[group.PageKey] = group.Name
		}
	}

	records := make([]Record, 0, len(items))
	for _, rawItem := range items {
		item := normalizeLegacyGlobalPage(rawItem)
		record := Record{UIPage: item}
		if item.ParentMenuID != nil {
			record.ParentMenuName = menuNameMap[*item.ParentMenuID]
		}
		if item.ParentPageKey != "" {
			record.ParentPageName = parentPageNameMap[item.ParentPageKey]
		}
		if item.DisplayGroupKey != "" {
			record.DisplayGroupName = displayGroupNameMap[item.DisplayGroupKey]
		}
		records = append(records, record)
	}
	return records, nil
}

func pageToUpdateMap(item *models.UIPage) map[string]interface{} {
	return map[string]interface{}{
		"app_key":            item.AppKey,
		"page_key":           item.PageKey,
		"name":               item.Name,
		"route_name":         item.RouteName,
		"route_path":         item.RoutePath,
		"component":          item.Component,
		"page_type":          item.PageType,
		"source":             item.Source,
		"module_key":         item.ModuleKey,
		"sort_order":         item.SortOrder,
		"parent_menu_id":     item.ParentMenuID,
		"parent_page_key":    item.ParentPageKey,
		"display_group_key":  item.DisplayGroupKey,
		"active_menu_path":   item.ActiveMenuPath,
		"breadcrumb_mode":    item.BreadcrumbMode,
		"access_mode":        item.AccessMode,
		"permission_key":     item.PermissionKey,
		"inherit_permission": item.InheritPermission,
		"keep_alive":         item.KeepAlive,
		"is_full_page":       item.IsFullPage,
		"status":             item.Status,
		"meta":               item.Meta,
	}
}

func (s *service) hydrateManagedRecord(record *Record) error {
	if record == nil {
		return nil
	}
	menuMap, err := s.loadMenuMap(record.AppKey)
	if err != nil {
		return err
	}
	pageMap, err := s.loadPageMap(record.AppKey)
	if err != nil {
		return err
	}
	bindingMap, err := loadPageSpaceBindingMap(s.db, record.AppKey)
	if err != nil {
		return err
	}
	resolvedSpaceKeys := resolvePageSpaceKeys(record.UIPage, pageMap, menuMap, bindingMap, map[string]struct{}{})
	applyResolvedPageSpace(&record.UIPage, resolvedSpaceKeys)
	record.ActiveMenuPath = s.resolveActiveMenuPath(
		&record.UIPage,
		menuMap,
		pageMap,
		map[string]struct{}{},
	)
	return nil
}

func syncPageSpaceBindings(
	tx *gorm.DB,
	appKey string,
	pageID uuid.UUID,
	spaceKey string,
	spaceKeys []string,
	parentMenuID *uuid.UUID,
	parentPageKey string,
) error {
	if tx == nil || pageID == uuid.Nil {
		return nil
	}
	if err := tx.Where("app_key = ? AND page_id = ?", normalizeAppKey(appKey), pageID).Delete(&models.PageSpaceBinding{}).Error; err != nil {
		return err
	}
	bindingKeys := normalizeStandalonePageBindingKeys(spaceKey, spaceKeys, parentMenuID, parentPageKey)
	if len(bindingKeys) == 0 {
		return nil
	}
	rows := make([]models.PageSpaceBinding, 0, len(bindingKeys))
	for _, key := range bindingKeys {
		rows = append(rows, models.PageSpaceBinding{
			ID:       uuid.New(),
			AppKey:   normalizeAppKey(appKey),
			PageID:   pageID,
			SpaceKey: key,
		})
	}
	return tx.Create(&rows).Error
}

func (s *service) findPageByID(id uuid.UUID, appKey string) (*models.UIPage, error) {
	var item models.UIPage
	if err := s.db.Where("id = ? AND app_key = ?", id, normalizeAppKey(appKey)).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	normalized := normalizeLegacyGlobalPage(item)
	return &normalized, nil
}

func (s *service) loadPageMap(appKey string) (map[string]models.UIPage, error) {
	var items []models.UIPage
	if err := s.db.Where("app_key = ?", normalizeAppKey(appKey)).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make(map[string]models.UIPage, len(items))
	for _, rawItem := range items {
		item := normalizeLegacyGlobalPage(rawItem)
		result[item.PageKey] = item
	}
	return result, nil
}

func (s *service) loadMenuMap(appKey string) (map[uuid.UUID]runtimeMenuNode, error) {
	menus, err := s.menuRepo.ListByAppAndSpace(normalizeAppKey(appKey), "")
	if err != nil {
		return nil, err
	}
	menuMap := make(map[uuid.UUID]models.Menu, len(menus))
	for _, item := range menus {
		menuMap[item.ID] = item
	}

	result := make(map[uuid.UUID]runtimeMenuNode, len(menus))
	var resolveFullPath func(menu models.Menu) string
	resolveFullPath = func(menu models.Menu) string {
		if cached, ok := result[menu.ID]; ok && strings.TrimSpace(cached.FullPath) != "" {
			return cached.FullPath
		}

		parentPath := ""
		if menu.ParentID != nil {
			if parent, ok := menuMap[*menu.ParentID]; ok {
				parentPath = resolveFullPath(parent)
			}
		}

		fullPath := buildMenuFullPath(strings.TrimSpace(menu.Path), parentPath)
		result[menu.ID] = runtimeMenuNode{
			Menu:     menu,
			FullPath: fullPath,
		}
		return fullPath
	}

	for _, menu := range menus {
		resolveFullPath(menu)
	}
	return result, nil
}

func (s *service) resolveBreadcrumbChain(
	page *models.UIPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
) ([]BreadcrumbPreviewItem, error) {
	if page == nil {
		return []BreadcrumbPreviewItem{}, nil
	}
	switch normalizeBreadcrumbMode(page.BreadcrumbMode) {
	case "inherit_page":
		chain, err := s.resolveParentPageBreadcrumbChain(page, menuMap, pageMap, map[string]struct{}{})
		if err != nil {
			return nil, err
		}
		if len(chain) > 0 {
			return chain, nil
		}
		fallthrough
	case "custom":
		fallthrough
	default:
		activePath := s.resolveActiveMenuPath(page, menuMap, pageMap, map[string]struct{}{})
		return resolveMenuBreadcrumbChain(activePath, menuMap), nil
	}
}

func (s *service) resolveParentPageBreadcrumbChain(
	page *models.UIPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
	seen map[string]struct{},
) ([]BreadcrumbPreviewItem, error) {
	if page == nil {
		return []BreadcrumbPreviewItem{}, nil
	}
	parentKey := strings.TrimSpace(page.ParentPageKey)
	if parentKey == "" {
		activePath := s.resolveActiveMenuPath(page, menuMap, pageMap, seen)
		return resolveMenuBreadcrumbChain(activePath, menuMap), nil
	}
	if _, ok := seen[parentKey]; ok {
		return nil, fmt.Errorf("%w: 页面面包屑存在循环引用", ErrPageValidation)
	}
	parentPage, ok := pageMap[parentKey]
	if !ok {
		activePath := s.resolveActiveMenuPath(page, menuMap, pageMap, seen)
		return resolveMenuBreadcrumbChain(activePath, menuMap), nil
	}
	seen[parentKey] = struct{}{}
	parentChain, err := s.resolveBreadcrumbChain(&parentPage, menuMap, pageMap)
	delete(seen, parentKey)
	if err != nil {
		return nil, err
	}
	if isRoutelessPageType(parentPage.PageType) || strings.TrimSpace(parentPage.RoutePath) == "" {
		return parentChain, nil
	}
	return append(parentChain, BreadcrumbPreviewItem{
		Type:    "page",
		Title:   parentPage.Name,
		Path:    normalizeRoutePath(parentPage.RoutePath),
		PageKey: parentPage.PageKey,
	}), nil
}

func (s *service) resolveActiveMenuPath(
	page *models.UIPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
	seen map[string]struct{},
) string {
	if page == nil {
		return ""
	}
	if activePath := normalizeRoutePath(page.ActiveMenuPath); activePath != "" {
		return activePath
	}
	if page.ParentMenuID != nil {
		if node, ok := menuMap[*page.ParentMenuID]; ok {
			return node.FullPath
		}
	}
	parentKey := strings.TrimSpace(page.ParentPageKey)
	if parentKey == "" {
		return ""
	}
	if _, ok := seen[parentKey]; ok {
		return ""
	}
	parentPage, ok := pageMap[parentKey]
	if !ok {
		return ""
	}
	seen[parentKey] = struct{}{}
	defer delete(seen, parentKey)
	return s.resolveActiveMenuPath(&parentPage, menuMap, pageMap, seen)
}

func resolveMenuBreadcrumbChain(activePath string, menuMap map[uuid.UUID]runtimeMenuNode) []BreadcrumbPreviewItem {
	targetPath := normalizeRoutePath(activePath)
	if targetPath == "" {
		return []BreadcrumbPreviewItem{}
	}
	var target *runtimeMenuNode
	for _, node := range menuMap {
		if normalizeRoutePath(node.FullPath) == targetPath {
			item := node
			target = &item
			break
		}
	}
	if target == nil {
		return []BreadcrumbPreviewItem{}
	}

	chain := make([]BreadcrumbPreviewItem, 0, 4)
	current := target
	for current != nil {
		title := strings.TrimSpace(current.Menu.Title)
		if title == "" {
			title = strings.TrimSpace(current.Menu.Name)
		}
		chain = append(chain, BreadcrumbPreviewItem{
			Type:  "menu",
			Title: title,
			Path:  normalizeRoutePath(current.FullPath),
		})
		if current.Menu.ParentID == nil {
			break
		}
		parent, ok := menuMap[*current.Menu.ParentID]
		if !ok {
			break
		}
		parentCopy := parent
		current = &parentCopy
	}
	reverseBreadcrumbItems(chain)
	return chain
}

func (s *service) buildUnregisteredRecords(appKey string) ([]UnregisteredRecord, error) {
	viewPages, err := enumerateManagedViewPages()
	if err != nil {
		return nil, err
	}

	var existingPages []models.UIPage
	if err := s.db.Where("app_key = ?", normalizeAppKey(appKey)).Find(&existingPages).Error; err != nil {
		return nil, err
	}
	pageComponentSet := make(map[string]struct{}, len(existingPages))
	pageKeySet := make(map[string]struct{}, len(existingPages))
	routeNameSet := make(map[string]struct{}, len(existingPages))
	for _, item := range existingPages {
		pageComponentSet[strings.TrimSpace(item.Component)] = struct{}{}
		pageKeySet[strings.TrimSpace(item.PageKey)] = struct{}{}
		routeNameSet[strings.TrimSpace(item.RouteName)] = struct{}{}
	}

	menuMap, err := s.loadMenuMap(appKey)
	if err != nil {
		return nil, err
	}
	menuComponentSet := make(map[string]struct{}, len(menuMap))
	for _, node := range menuMap {
		component := strings.TrimSpace(node.Menu.Component)
		if component != "" {
			menuComponentSet[component] = struct{}{}
		}
	}

	result := make([]UnregisteredRecord, 0, len(viewPages))
	routePathSet := make(map[string]struct{})
	for _, page := range viewPages {
		component := strings.TrimSpace(page.Component)
		if component == "" {
			continue
		}
		if _, ok := pageComponentSet[component]; ok {
			continue
		}
		if _, ok := menuComponentSet[component]; ok {
			continue
		}

		candidate := deriveUnregisteredRecord(page, menuMap, pageKeySet, routeNameSet, routePathSet)
		pageKeySet[candidate.PageKey] = struct{}{}
		routeNameSet[candidate.RouteName] = struct{}{}
		routePathSet[candidate.RoutePath] = struct{}{}
		result = append(result, candidate)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Component < result[j].Component
	})
	return result, nil
}

func enumerateManagedViewPages() ([]scannedViewPage, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, err
	}
	viewsDir := filepath.Join(projectRoot, "frontend", "src", "views")
	items := make([]scannedViewPage, 0, 128)

	err = filepath.WalkDir(viewsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".vue" {
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		filePath := "/" + filepath.ToSlash(rel)
		component := toComponentPath(filePath)
		if !isManagedViewComponent(component) {
			return nil
		}
		items = append(items, scannedViewPage{
			FilePath:  filePath,
			Component: component,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func deriveUnregisteredRecord(
	item scannedViewPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageKeySet map[string]struct{},
	routeNameSet map[string]struct{},
	routePathSet map[string]struct{},
) UnregisteredRecord {
	routePath := normalizeRoutePath(item.Component)
	moduleKey := deriveModuleKey(item.Component)
	name := derivePageDisplayName(item.Component)
	pageKey := ensureUniqueValue(derivePageKey(item.Component), pageKeySet)
	routeName := ensureUniqueValue(deriveRouteName(item.Component), routeNameSet)
	if _, ok := routePathSet[routePath]; ok {
		routePath = ensureUniqueRoutePath(routePath, routePathSet)
	}

	parentMenuID, parentMenuName, activeMenuPath := guessParentMenu(routePath, menuMap)
	pageType := "global"
	if parentMenuID != "" {
		pageType = "inner"
	}

	return UnregisteredRecord{
		FilePath:       item.FilePath,
		Component:      item.Component,
		PageKey:        pageKey,
		Name:           name,
		RouteName:      routeName,
		RoutePath:      routePath,
		PageType:       pageType,
		ModuleKey:      moduleKey,
		ParentMenuID:   parentMenuID,
		ParentMenuName: parentMenuName,
		ActiveMenuPath: activeMenuPath,
	}
}

func guessParentMenu(routePath string, menuMap map[uuid.UUID]runtimeMenuNode) (string, string, string) {
	targetPath := normalizeRoutePath(routePath)
	bestDepth := -1
	bestID := ""
	bestName := ""
	bestPath := ""
	for _, node := range menuMap {
		menuPath := normalizeRoutePath(node.FullPath)
		if menuPath == "" || menuPath == "/" {
			continue
		}
		if !strings.HasPrefix(targetPath, menuPath) {
			continue
		}
		if targetPath != menuPath && !strings.HasPrefix(targetPath, menuPath+"/") {
			continue
		}
		depth := strings.Count(menuPath, "/")
		if depth <= bestDepth {
			continue
		}
		bestDepth = depth
		bestID = node.Menu.ID.String()
		bestName = firstNonEmpty(strings.TrimSpace(node.Menu.Title), strings.TrimSpace(node.Menu.Name))
		bestPath = menuPath
	}
	return bestID, bestName, bestPath
}

func deriveSyncedPageAccessMode(pageType string) string {
	if normalizePageType(pageType) == "inner" {
		return "inherit"
	}
	return "jwt"
}

func boolPtr(value bool) *bool {
	return &value
}

func ensureUniqueValue(base string, used map[string]struct{}) string {
	target := strings.TrimSpace(base)
	if target == "" {
		target = "page"
	}
	if _, ok := used[target]; !ok {
		return target
	}
	for idx := 2; idx < 10000; idx++ {
		candidate := target + "_" + strconv.Itoa(idx)
		if _, ok := used[candidate]; !ok {
			return candidate
		}
	}
	return target + "_" + uuid.NewString()[:8]
}

func ensureUniqueRoutePath(base string, used map[string]struct{}) string {
	target := normalizeRoutePath(base)
	if _, ok := used[target]; !ok {
		return target
	}
	for idx := 2; idx < 10000; idx++ {
		candidate := normalizeRoutePath(target + "-" + strconv.Itoa(idx))
		if _, ok := used[candidate]; !ok {
			return candidate
		}
	}
	return normalizeRoutePath(target + "-" + uuid.NewString()[:8])
}

func derivePageKey(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return "page"
	}
	normalized := make([]string, 0, len(segments))
	for _, segment := range segments {
		normalized = append(normalized, sanitizeSegment(segment))
	}
	return strings.Join(normalized, ".")
}

func deriveRouteName(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return "Page"
	}
	builder := strings.Builder{}
	for _, segment := range segments {
		builder.WriteString(toPascalCase(segment))
	}
	result := builder.String()
	if result == "" {
		return "Page"
	}
	return result
}

func derivePageDisplayName(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return "未命名页面"
	}
	return humanizeSegment(segments[len(segments)-1])
}

func deriveModuleKey(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return ""
	}
	return sanitizeSegment(segments[0])
}

func splitComponentSegments(component string) []string {
	target := strings.Trim(component, "/")
	if target == "" {
		return []string{}
	}
	segments := strings.Split(target, "/")
	result := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" || segment == "index" {
			continue
		}
		result = append(result, segment)
	}
	return result
}

func humanizeSegment(segment string) string {
	parts := splitWords(segment)
	if len(parts) == 0 {
		return "未命名页面"
	}
	for idx, part := range parts {
		parts[idx] = strings.Title(part)
	}
	return strings.Join(parts, " ")
}

func toPascalCase(segment string) string {
	parts := splitWords(segment)
	if len(parts) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

func splitWords(value string) []string {
	replacer := strings.NewReplacer("-", " ", "_", " ", ".", " ")
	normalized := replacer.Replace(strings.TrimSpace(value))
	return strings.Fields(normalized)
}

func sanitizeSegment(segment string) string {
	target := strings.TrimSpace(segment)
	if target == "" {
		return "page"
	}
	var builder strings.Builder
	for _, r := range target {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(unicode.ToLower(r))
		case r == '-' || r == '_' || r == '.':
			builder.WriteRune('_')
		}
	}
	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "page"
	}
	return result
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := wd
	for {
		frontendPath := filepath.Join(current, "frontend")
		backendPath := filepath.Join(current, "backend")
		if _, err := os.Stat(frontendPath); err == nil {
			if _, err := os.Stat(backendPath); err == nil {
				return current, nil
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return wd, nil
}

func toComponentPath(filePath string) string {
	withoutPrefix := strings.TrimPrefix(filePath, "/frontend/src/views")
	withoutExt := strings.TrimSuffix(withoutPrefix, ".vue")
	normalized := strings.TrimSuffix(withoutExt, "/index")
	normalized = strings.ReplaceAll(normalized, "//", "/")
	if normalized == "" {
		return "/"
	}
	if !strings.HasPrefix(normalized, "/") {
		return "/" + normalized
	}
	return normalized
}

func isManagedViewComponent(component string) bool {
	target := normalizeRoutePath(component)
	switch {
	case target == "", target == "/", target == "/index", target == "/outside/Iframe":
		return false
	case strings.Contains(target, "/modules/"):
		return false
	case strings.HasPrefix(target, "/auth/"),
		strings.HasPrefix(target, "/exception/"),
		strings.HasPrefix(target, "/result/"):
		return false
	default:
		return true
	}
}

func buildMenuFullPath(path, parentPath string) string {
	target := strings.TrimSpace(path)
	if target == "" {
		return normalizeRoutePath(parentPath)
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	if strings.HasPrefix(target, "/") {
		return normalizeRoutePath(target)
	}
	parent := normalizeRoutePath(parentPath)
	if parent == "" || parent == "/" {
		return normalizeRoutePath("/" + target)
	}
	return normalizeRoutePath(strings.TrimRight(parent, "/") + "/" + strings.TrimLeft(target, "/"))
}

func normalizeRoutePath(path string) string {
	target := strings.TrimSpace(path)
	if target == "" {
		return ""
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	normalized := "/" + strings.TrimLeft(target, "/")
	normalized = strings.ReplaceAll(normalized, "//", "/")
	if normalized != "/" {
		normalized = strings.TrimRight(normalized, "/")
	}
	return normalized
}

func reverseBreadcrumbItems(items []BreadcrumbPreviewItem) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

func normalizeAppKey(value string) string {
	return apppkg.NormalizeAppKey(value)
}

func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	target := strings.TrimSpace(raw)
	if target == "" {
		return nil, nil
	}
	id, err := uuid.Parse(target)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func normalizePageAndSize(req *ListRequest) (int, int) {
	current := 1
	size := 20
	if req == nil {
		return current, size
	}
	if req.Current > 0 {
		current = req.Current
	}
	if req.Size > 0 && req.Size <= 1000 {
		size = req.Size
	}
	return current, size
}

func normalizePageType(value string) string {
	switch strings.TrimSpace(value) {
	case "group", "display_group", "inner", "global":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeLegacyGlobalPage(item models.UIPage) models.UIPage {
	if normalizePageType(item.PageType) != "global" {
		return item
	}
	item.ParentMenuID = nil
	item.ParentPageKey = ""
	item.ActiveMenuPath = ""
	return item
}

func isRoutelessPageType(pageType string) bool {
	switch normalizePageType(pageType) {
	case "group", "display_group":
		return true
	default:
		return false
	}
}

func normalizeSource(value string) string {
	switch strings.TrimSpace(value) {
	case "seed", "sync", "manual":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeBreadcrumbMode(value string) string {
	switch strings.TrimSpace(value) {
	case "inherit_menu", "inherit_page", "custom":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeAccessMode(value string) string {
	switch strings.TrimSpace(value) {
	case "inherit", "public", "jwt", "permission":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "normal", "suspended":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeStandalonePageSpaceKey(value string, parentMenuID *uuid.UUID, parentPageKey string) string {
	// 新模型下页面主表不再承担空间归属语义：
	// - 常规页面从菜单或父页面继承空间
	// - 少量独立页通过 page_space_bindings 控制暴露范围
	// 因此新写入统一清空主表 space_key，仅保留旧数据兼容读取。
	_ = value
	if parentMenuID != nil || strings.TrimSpace(parentPageKey) != "" {
		return ""
	}
	return ""
}

func normalizeStandalonePageBindingKeys(value string, values []string, parentMenuID *uuid.UUID, parentPageKey string) []string {
	if parentMenuID != nil || strings.TrimSpace(parentPageKey) != "" {
		return []string{}
	}
	if values != nil {
		candidates := make([]string, 0, len(values))
		for _, item := range values {
			if target := normalizeSpaceKey(item); target != "" {
				candidates = append(candidates, target)
			}
		}
		return uniqueSortedStrings(candidates)
	}
	candidates := make([]string, 0, len(values)+1)
	for _, item := range values {
		if target := normalizeSpaceKey(item); target != "" {
			candidates = append(candidates, target)
		}
	}
	if len(candidates) == 0 {
		if target := normalizeSpaceKey(value); target != "" {
			candidates = append(candidates, target)
		}
	}
	return uniqueSortedStrings(candidates)
}

func (s *service) findPageByKey(pageKey string, appKey string) (*models.UIPage, error) {
	var parentPage models.UIPage
	err := s.db.Where("app_key = ? AND page_key = ?", normalizeAppKey(appKey), pageKey).First(&parentPage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrParentPageInvalid
		}
		return nil, err
	}
	return &parentPage, nil
}

func (s *service) findDisplayGroupByKey(pageKey string, appKey string) (*models.UIPage, error) {
	group, err := s.findPageByKey(pageKey, appKey)
	if err != nil {
		if errors.Is(err, ErrParentPageInvalid) {
			return nil, ErrDisplayGroupInvalid
		}
		return nil, err
	}
	if normalizePageType(group.PageType) != "display_group" {
		return nil, ErrDisplayGroupInvalid
	}
	return group, nil
}

func (s *service) ensureNoPageCycle(currentPageKey, candidateParentKey string, appKey string) error {
	seen := map[string]struct{}{currentPageKey: {}}
	nextKey := strings.TrimSpace(candidateParentKey)
	for nextKey != "" {
		if _, exists := seen[nextKey]; exists {
			return fmt.Errorf("%w: 页面分组形成循环引用", ErrPageValidation)
		}
		seen[nextKey] = struct{}{}
		parent, err := s.findPageByKey(nextKey, appKey)
		if err != nil {
			return err
		}
		nextKey = strings.TrimSpace(parent.ParentPageKey)
	}
	return nil
}
