package page

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	spaceutil "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

var (
	ErrPageNotFound        = errors.New("页面不存在")
	ErrPageKeyExists       = errors.New("页面 key 已存在")
	ErrRouteNameExists     = errors.New("路由名称已存在")
	ErrRoutePathExists     = errors.New("路由路径已存在")
	ErrParentMenuInvalid   = errors.New("无效的上级菜单")
	ErrParentPageInvalid   = errors.New("无效的上级页面")
	ErrDisplayGroupInvalid = errors.New("无效的显示分组")
	ErrPageHasChildren     = errors.New("页面下存在子页面，无法删除")
	ErrPageValidation      = errors.New("页面数据校验失败")
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
	SpaceKeys         []string               `json:"space_keys"`
	PageType          string                 `json:"page_type"`
	VisibilityScope   string                 `json:"visibility_scope"`
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
	RemoteBinding     *RemoteBinding         `json:"remote_binding"`
	Meta              map[string]interface{} `json:"meta"`
}

type RemoteBinding struct {
	ManifestURL      string `json:"manifest_url"`
	RemoteAppKey     string `json:"remote_app_key"`
	RemotePageKey    string `json:"remote_page_key"`
	RemoteEntryURL   string `json:"remote_entry_url"`
	RemoteRoutePath  string `json:"remote_route_path"`
	RemoteModule     string `json:"remote_module"`
	RemoteModuleName string `json:"remote_module_name"`
	RemoteURL        string `json:"remote_url"`
	RuntimeVersion   string `json:"runtime_version"`
	HealthCheckURL   string `json:"health_check_url"`
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
	FilePath        string `json:"file_path"`
	Component       string `json:"component"`
	PageKey         string `json:"page_key"`
	Name            string `json:"name"`
	RouteName       string `json:"route_name"`
	RoutePath       string `json:"route_path"`
	PageType        string `json:"page_type"`
	VisibilityScope string `json:"visibility_scope"`
	ModuleKey       string `json:"module_key"`
	ParentMenuID    string `json:"parent_menu_id"`
	ParentMenuName  string `json:"parent_menu_name"`
	ActiveMenuPath  string `json:"active_menu_path"`
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
	AppKey                   string `form:"app_key"`
	UserID                   string `form:"user_id"`
	CollaborationWorkspaceID string `form:"collaboration_workspace_id"`
	PageKey                  string `form:"page_key"`
	SpaceKey                 string `form:"space_key"`
	PageKeys                 string `form:"page_keys"`
	RoutePath                string `form:"route_path"`
}

type AccessTraceRoleItem struct {
	RoleID   string `json:"role_id"`
	RoleCode string `json:"role_code"`
	RoleName string `json:"role_name"`
	Status   string `json:"status"`
}

type AccessTraceMenuItem struct {
	ID        string `json:"id"`
	ParentID  string `json:"parent_id"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	FullPath  string `json:"full_path"`
	Kind      string `json:"kind"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	Hidden    bool   `json:"hidden"`
	Visible   bool   `json:"visible"`
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
	UserID                   string                `json:"user_id"`
	CollaborationWorkspaceID string                `json:"collaboration_workspace_id,omitempty"`
	SpaceKey                 string                `json:"space_key"`
	Authenticated            bool                  `json:"authenticated"`
	SuperAdmin               bool                  `json:"super_admin"`
	ActionKeyCount           int                   `json:"action_key_count"`
	VisibleMenuIDs           []string              `json:"visible_menu_ids"`
	Menus                    []AccessTraceMenuItem `json:"menus"`
	Roles                    []AccessTraceRoleItem `json:"roles"`
	Pages                    []AccessTracePageItem `json:"pages"`
}

type Service interface {
	List(req *ListRequest) ([]Record, int64, error)
	ListOptions(appKey, spaceKey string) ([]models.UIPage, error)
	ListRuntime(appKey, host, requestedSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]Record, error)
	ListRuntimePublic(appKey, host, requestedSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]Record, error)
	ResolveCompiledAccessContext(appKey, spaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) (*CompiledAccessContext, error)
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
	menuMap, err := s.loadMenuMap(appKey, req.SpaceKey)
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
			"visibility_scope",
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
	menuMap, err := s.loadMenuMap(normalizeAppKey(appKey), spaceKey)
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
	var collaborationWorkspaceID *uuid.UUID
	if rawCW := strings.TrimSpace(req.CollaborationWorkspaceID); rawCW != "" {
		parsed, parseErr := uuid.Parse(rawCW)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: collaboration_workspace_id is invalid", ErrPageValidation)
		}
		collaborationWorkspaceID = &parsed
	}
	spaceKey := normalizeSpaceKey(req.SpaceKey)
	if spaceKey == "" {
		resolvedSpaceKey, _, resolveErr := spaceutil.ResolveCurrentSpaceKey(s.db, appKey, "", "", nil, nil)
		if resolveErr != nil {
			return nil, resolveErr
		}
		spaceKey = normalizeSpaceKey(resolvedSpaceKey)
		if spaceKey == "" {
			spaceKey = spaceutil.DefaultMenuSpaceKey
		}
	}
	accessCtx, err := s.ResolveCompiledAccessContext(appKey, spaceKey, &userID, collaborationWorkspaceID)
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

	roleItems, err := s.loadAccessTraceRoles(userID, collaborationWorkspaceID)
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

	menuItems, err := s.buildAccessTraceMenus(appKey, spaceKey, accessCtx)
	if err != nil {
		return nil, err
	}

	result := &AccessTraceResult{
		UserID:         userID.String(),
		SpaceKey:       spaceKey,
		Authenticated:  accessCtx.Authenticated,
		SuperAdmin:     accessCtx.SuperAdmin,
		ActionKeyCount: len(accessCtx.ActionKeys),
		VisibleMenuIDs: visibleMenuIDStrings,
		Menus:          menuItems,
		Roles:          roleItems,
		Pages:          pageItems,
	}
	if collaborationWorkspaceID != nil {
		result.CollaborationWorkspaceID = collaborationWorkspaceID.String()
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
	if permKey := strings.TrimSpace(page.PermissionKey); permKey != "" {
		normalizedPermissionKey := permissionkey.Normalize(permKey)
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

func (s *service) buildAccessTraceMenus(appKey, spaceKey string, accessCtx *CompiledAccessContext) ([]AccessTraceMenuItem, error) {
	menuMap, err := s.loadMenuMap(appKey, spaceKey)
	if err != nil {
		return nil, err
	}
	visible := map[uuid.UUID]struct{}{}
	if accessCtx != nil {
		visible = accessCtx.VisibleMenuIDs
	}
	items := make([]AccessTraceMenuItem, 0, len(menuMap))
	for _, node := range menuMap {
		menu := node.Menu
		parentID := ""
		if menu.ParentID != nil {
			parentID = menu.ParentID.String()
		}
		_, isVisible := visible[menu.ID]
		if accessCtx != nil && accessCtx.SuperAdmin {
			isVisible = true
		}
		items = append(items, AccessTraceMenuItem{
			ID:        menu.ID.String(),
			ParentID:  parentID,
			Name:      strings.TrimSpace(menu.Name),
			Title:     strings.TrimSpace(menu.Title),
			Path:      strings.TrimSpace(menu.Path),
			FullPath:  strings.TrimSpace(node.FullPath),
			Kind:      strings.TrimSpace(menu.Kind),
			Icon:      strings.TrimSpace(menu.Icon),
			SortOrder: menu.SortOrder,
			Hidden:    menu.Hidden,
			Visible:   isVisible,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].Name < items[j].Name
	})
	return items, nil
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

func (s *service) loadAccessTraceRoles(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]AccessTraceRoleItem, error) {
	roles := make([]models.Role, 0)
	if collaborationWorkspaceID != nil {
		effective, err := s.loadRuntimeEffectiveActiveRoles(userID, *collaborationWorkspaceID)
		if err != nil {
			return nil, err
		}
		roles = effective
	} else {
		roleIDs, err := workspacerolebinding.ListPersonalRoleIDsByUserID(s.db, userID, true)
		if err != nil {
			return nil, err
		}
		if len(roleIDs) > 0 {
			if err := s.db.Model(&models.Role{}).
				Where("id IN ?", roleIDs).
				Where("collaboration_workspace_id IS NULL").
				Where("status = ?", "normal").
				Find(&roles).Error; err != nil {
				return nil, err
			}
		} else {
			if err := s.db.Model(&models.Role{}).
				Joins("JOIN user_roles ON user_roles.role_id = roles.id").
				Where("user_roles.user_id = ?", userID).
				Where("user_roles.collaboration_workspace_id IS NULL").
				Where("roles.collaboration_workspace_id IS NULL").
				Where("roles.status = ?", "normal").
				Where("roles.deleted_at IS NULL").
				Distinct("roles.*").
				Find(&roles).Error; err != nil {
				return nil, err
			}
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
		return syncPageSpaceBindings(tx, item.AppKey, item.ID, req.SpaceKeys, item.PageType, item.VisibilityScope, item.ParentMenuID, item.ParentPageKey)
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
		return syncPageSpaceBindings(tx, existing.AppKey, existing.ID, req.SpaceKeys, item.PageType, item.VisibilityScope, item.ParentMenuID, item.ParentPageKey)
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

// ---------------------------------------------------------------------------
// private helpers
// ---------------------------------------------------------------------------

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
		pageType = "standalone"
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
	permKey := strings.TrimSpace(req.PermissionKey)

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
	if accessMode == "permission" {
		if permKey == "" {
			return nil, fmt.Errorf("%w: 权限访问模式必须指定权限键", ErrPageValidation)
		}
	} else {
		// 仅 permission 模式下 permKey 有语义，其他模式强制清空避免脏数据
		permKey = ""
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
		permKey = ""
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
	if pageType == "display_group" {
		parentMenuID = nil
	}
	if pageType == models.PageTypeInner && parentMenuID == nil && parentPageKey == "" {
		return nil, fmt.Errorf("%w: 内页必须挂到菜单或上级页面", ErrPageValidation)
	}
	if pageType == models.PageTypeStandalone && (parentMenuID != nil || parentPageKey != "") {
		return nil, fmt.Errorf("%w: 独立页不能挂到菜单或上级页面", ErrPageValidation)
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
	// 非 routeless 页面（非 group / display_group）需要确保 route_path 在 App 内唯一，
	// 否则运行时前端路由注册会冲突。
	if !isRoutelessPageType(pageType) && routePath != "" {
		if err := s.ensureUnique("route_path", routePath, checkID, appKey); err != nil {
			return nil, err
		}
	}

	meta := models.MetaJSON(req.Meta)
	if meta == nil {
		meta = models.MetaJSON{}
	}
	applyRemoteBindingContract(meta, req.RemoteBinding)

	item := &models.UIPage{
		AppKey:            appKey,
		PageKey:           pageKey,
		Name:              name,
		RouteName:         routeName,
		RoutePath:         routePath,
		Component:         component,
		SpaceKey:          "",
		PageType:          pageType,
		VisibilityScope:   normalizePageVisibilityScopeForSave(pageType, req.VisibilityScope, parentMenuID, parentPageKey),
		Source:            source,
		ModuleKey:         moduleKey,
		SortOrder:         req.SortOrder,
		ParentMenuID:      parentMenuID,
		ParentPageKey:     parentPageKey,
		DisplayGroupKey:   displayGroupKey,
		ActiveMenuPath:    activeMenuPath,
		BreadcrumbMode:    breadcrumbMode,
		AccessMode:        accessMode,
		PermissionKey:     permKey,
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
	case "route_path":
		return ErrRoutePathExists
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
		item := rawItem
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
		item := rawItem
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
		"space_key":          "",
		"page_type":          item.PageType,
		"visibility_scope":   item.VisibilityScope,
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

func applyRemoteBindingContract(meta models.MetaJSON, binding *RemoteBinding) {
	if meta == nil {
		return
	}
	if binding == nil {
		return
	}
	writeRemoteMetaString(meta, "manifest_url", binding.ManifestURL)
	writeRemoteMetaString(meta, "remote_app_key", binding.RemoteAppKey)
	writeRemoteMetaString(meta, "remote_page_key", binding.RemotePageKey)
	writeRemoteMetaString(meta, "remote_entry_url", binding.RemoteEntryURL)
	writeRemoteMetaString(meta, "remote_route_path", binding.RemoteRoutePath)
	writeRemoteMetaString(meta, "remote_module", binding.RemoteModule)
	writeRemoteMetaString(meta, "remote_module_name", binding.RemoteModuleName)
	writeRemoteMetaString(meta, "remote_url", binding.RemoteURL)
	writeRemoteMetaString(meta, "runtime_version", binding.RuntimeVersion)
	writeRemoteMetaString(meta, "health_check_url", binding.HealthCheckURL)
	for _, key := range []string{
		"manifestUrl",
		"remoteAppKey",
		"remotePageKey",
		"remoteEntryUrl",
		"remoteRoutePath",
		"remoteModule",
		"remoteModuleName",
		"remoteUrl",
		"runtimeVersion",
		"version",
		"healthCheckUrl",
	} {
		delete(meta, key)
	}
}

func writeRemoteMetaString(meta models.MetaJSON, key string, value string) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		delete(meta, key)
		return
	}
	meta[key] = normalized
}

func (s *service) hydrateManagedRecord(record *Record) error {
	if record == nil {
		return nil
	}
	menuMap, err := s.loadMenuMap(record.AppKey, firstResolvedPageSpaceKey(record.UIPage))
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
	spaceKeys []string,
	pageType string,
	visibilityScope string,
	parentMenuID *uuid.UUID,
	parentPageKey string,
) error {
	if tx == nil || pageID == uuid.Nil {
		return nil
	}
	if err := tx.Where("app_key = ? AND page_id = ?", normalizeAppKey(appKey), pageID).Delete(&models.PageSpaceBinding{}).Error; err != nil {
		return err
	}
	bindingKeys := normalizeStandalonePageBindingKeys(spaceKeys, pageType, visibilityScope, parentMenuID, parentPageKey)
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
	return &item, nil
}

func (s *service) loadPageMap(appKey string) (map[string]models.UIPage, error) {
	var items []models.UIPage
	if err := s.db.Where("app_key = ?", normalizeAppKey(appKey)).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make(map[string]models.UIPage, len(items))
	for _, rawItem := range items {
		item := rawItem
		result[item.PageKey] = item
	}
	return result, nil
}

func (s *service) loadMenuMap(appKey string, spaceKey string) (map[uuid.UUID]runtimeMenuNode, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	normalizedSpaceKey := normalizeSpaceKey(spaceKey)
	if strings.TrimSpace(spaceKey) == "" {
		if resolvedSpaceKey, _, err := spaceutil.ResolveSpaceKeyByHost(s.db, normalizedAppKey, ""); err == nil {
			normalizedSpaceKey = normalizeSpaceKey(resolvedSpaceKey)
		}
	}
	menus, err := s.menuRepo.ListByAppAndSpace(normalizedAppKey, normalizedSpaceKey)
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
		result[menu.ID] = runtimeMenuNode{Menu: menu, FullPath: fullPath}
		return fullPath
	}

	for _, menu := range menus {
		resolveFullPath(menu)
	}
	return result, nil
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

// ---------------------------------------------------------------------------
// normalisation helpers
// ---------------------------------------------------------------------------

func firstResolvedPageSpaceKey(item models.UIPage) string {
	keys := readPageSpaceKeys(item)
	if len(keys) > 0 {
		return keys[0]
	}
	return ""
}

func boolPtr(value bool) *bool {
	return &value
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
	case "group", "display_group", "inner", "standalone":
		return strings.TrimSpace(value)
	default:
		return ""
	}
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

func normalizePageVisibilityScopeForSave(pageType string, value string, parentMenuID *uuid.UUID, parentPageKey string) string {
	switch normalizePageType(pageType) {
	case models.PageTypeInner:
		return pageVisibilityScopeInherit
	case models.PageTypeStandalone, "group", "display_group":
		if parentMenuID != nil || strings.TrimSpace(parentPageKey) != "" {
			return pageVisibilityScopeInherit
		}
		if strings.TrimSpace(value) == pageVisibilityScopeSpaces {
			return pageVisibilityScopeSpaces
		}
		return pageVisibilityScopeApp
	default:
		return pageVisibilityScopeApp
	}
}

func normalizeStandalonePageBindingKeys(values []string, pageType string, visibilityScope string, parentMenuID *uuid.UUID, parentPageKey string) []string {
	if normalizePageVisibilityScopeForSave(pageType, visibilityScope, parentMenuID, parentPageKey) != pageVisibilityScopeSpaces {
		return []string{}
	}
	candidates := make([]string, 0, len(values))
	for _, item := range values {
		if target := normalizeSpaceKey(item); target != "" {
			candidates = append(candidates, target)
		}
	}
	return uniqueSortedStrings(candidates)
}

func normalizeAppKey(value string) string {
	return apppkg.NormalizeAppKey(value)
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
