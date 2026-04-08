package featurepackage

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
)

var (
	ErrFeaturePackageNotFound = errors.New("功能包不存在")
	ErrFeaturePackageExists   = errors.New("功能包已存在")
	ErrFeaturePackageBuiltin  = errors.New("内置功能包不可修改")
)

type Service interface {
	List(req *dto.FeaturePackageListRequest) ([]user.FeaturePackage, int64, error)
	ListOptions(req *dto.FeaturePackageListRequest) ([]user.FeaturePackage, error)
	GetPackageStats(packageIDs []uuid.UUID) (map[uuid.UUID]int64, map[uuid.UUID]int64, map[uuid.UUID]int64, error)
	GetRelationTree(workspaceScope, keyword string) (*FeaturePackageRelationTree, error)
	Get(id uuid.UUID) (*user.FeaturePackage, error)
	Create(req *dto.FeaturePackageCreateRequest) (*user.FeaturePackage, error)
	Update(id uuid.UUID, req *dto.FeaturePackageUpdateRequest) (*permissionrefresh.RefreshStats, error)
	Delete(id uuid.UUID) (*permissionrefresh.RefreshStats, error)
	GetPackageChildren(id uuid.UUID, appKey string) ([]uuid.UUID, []user.FeaturePackage, error)
	SetPackageChildren(id uuid.UUID, childPackageIDs []uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error)
	GetPackageKeys(id uuid.UUID, appKey string) ([]uuid.UUID, []user.PermissionKey, error)
	SetPackageKeys(id uuid.UUID, actionIDs []uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error)
	GetPackageMenus(id uuid.UUID, appKey string) ([]uuid.UUID, []user.Menu, error)
	SetPackageMenus(id uuid.UUID, menuIDs []uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error)
	GetPackageCollaborationWorkspaces(id uuid.UUID, appKey string) ([]uuid.UUID, error)
	SetPackageCollaborationWorkspaces(id uuid.UUID, collaborationWorkspaceIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error)
	GetCollaborationWorkspacePackages(collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, []user.FeaturePackage, error)
	SetCollaborationWorkspacePackages(collaborationWorkspaceID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error)
	GetImpactPreview(id uuid.UUID) (*FeaturePackageImpactPreview, error)
	ListVersions(id uuid.UUID, current, size int) ([]user.FeaturePackageVersion, int64, error)
	Rollback(id uuid.UUID, versionID uuid.UUID, operatorID *uuid.UUID, requestID string) (*permissionrefresh.RefreshStats, error)
	ListRiskAudits(id uuid.UUID, current, size int) ([]user.RiskOperationAudit, int64, error)
}

type FeaturePackageRelationNode struct {
	ID             uuid.UUID                    `json:"id"`
	PackageKey     string                       `json:"package_key"`
	Name           string                       `json:"name"`
	PackageType    string                       `json:"package_type"`
	WorkspaceScope string                       `json:"workspace_scope"`
	AppKeys        []string                     `json:"app_keys,omitempty"`
	Status         string                       `json:"status"`
	ReferenceCount int                          `json:"reference_count"`
	Children       []FeaturePackageRelationNode `json:"children,omitempty"`
}

type FeaturePackageRelationTree struct {
	Roots             []FeaturePackageRelationNode `json:"roots"`
	CycleDependencies [][]string                   `json:"cycle_dependencies"`
	IsolatedBaseKeys  []string                     `json:"isolated_base_keys"`
}

type FeaturePackageImpactPreview struct {
	PackageID                   uuid.UUID `json:"package_id"`
	RoleCount                   int64     `json:"role_count"`
	CollaborationWorkspaceCount int64     `json:"collaboration_workspace_count"`
	UserCount                   int64     `json:"user_count"`
	MenuCount                   int64     `json:"menu_count"`
	ActionCount                 int64     `json:"action_count"`
}

type service struct {
	db                                       *gorm.DB
	packageRepo                              user.FeaturePackageRepository
	packageBundleRepo                        user.FeaturePackageBundleRepository
	packageActionRepo                        user.FeaturePackageKeyRepository
	packageMenuRepo                          user.FeaturePackageMenuRepository
	collaborationWorkspaceFeaturePackageRepo user.CollaborationWorkspaceFeaturePackageRepository
	rolePackageRepo                          user.RoleFeaturePackageRepository
	actionRepo                               user.PermissionKeyRepository
	menuRepo                                 user.MenuRepository
	collaborationWorkspaceRepo               user.CollaborationWorkspaceRepository
	boundaryService                          collaborationworkspaceboundary.Service
	refresher                                permissionrefresh.Service
}

func NewService(
	db *gorm.DB,
	packageRepo user.FeaturePackageRepository,
	packageBundleRepo user.FeaturePackageBundleRepository,
	packageActionRepo user.FeaturePackageKeyRepository,
	packageMenuRepo user.FeaturePackageMenuRepository,
	collaborationWorkspaceFeaturePackageRepo user.CollaborationWorkspaceFeaturePackageRepository,
	rolePackageRepo user.RoleFeaturePackageRepository,
	actionRepo user.PermissionKeyRepository,
	menuRepo user.MenuRepository,
	collaborationWorkspaceRepo user.CollaborationWorkspaceRepository,
	boundaryService collaborationworkspaceboundary.Service,
	refresher permissionrefresh.Service,
) Service {
	return &service{
		db:                                       db,
		packageRepo:                              packageRepo,
		packageBundleRepo:                        packageBundleRepo,
		packageActionRepo:                        packageActionRepo,
		packageMenuRepo:                          packageMenuRepo,
		collaborationWorkspaceFeaturePackageRepo: collaborationWorkspaceFeaturePackageRepo,
		rolePackageRepo:                          rolePackageRepo,
		actionRepo:                               actionRepo,
		menuRepo:                                 menuRepo,
		collaborationWorkspaceRepo:               collaborationWorkspaceRepo,
		boundaryService:                          boundaryService,
		refresher:                                refresher,
	}
}

func (s *service) List(req *dto.FeaturePackageListRequest) ([]user.FeaturePackage, int64, error) {
	if req == nil {
		return nil, 0, errors.New("无效的请求参数")
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	return s.packageRepo.List((req.Current-1)*req.Size, req.Size, &user.FeaturePackageListParams{
		Keyword:        strings.TrimSpace(req.Keyword),
		AppKey:         normalizeAppKey(req.AppKey),
		PackageKey:     strings.TrimSpace(req.PackageKey),
		PackageType:    normalizePackageType(req.PackageType),
		Name:           strings.TrimSpace(req.Name),
		WorkspaceScope: normalizeWorkspaceScope(req.WorkspaceScope),
		Status:         strings.TrimSpace(req.Status),
	})
}

func (s *service) ListOptions(req *dto.FeaturePackageListRequest) ([]user.FeaturePackage, error) {
	if req == nil {
		return nil, errors.New("无效的请求参数")
	}
	query := s.db.Model(&user.FeaturePackage{})
	if req != nil {
		if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("(package_key LIKE ? OR name LIKE ? OR description LIKE ?)", like, like, like)
		}
		if packageKey := strings.TrimSpace(req.PackageKey); packageKey != "" {
			query = query.Where("package_key LIKE ?", "%"+packageKey+"%")
		}
		if packageType := normalizePackageType(req.PackageType); packageType != "" {
			query = query.Where("package_type = ?", packageType)
		}
		if appKey := normalizeAppKey(req.AppKey); appKey != "" {
			if s.db.Migrator().HasColumn(&user.FeaturePackage{}, "app_keys") {
				query = query.Where(
					"(app_key = ? OR COALESCE(jsonb_array_length(app_keys), 0) = 0 OR jsonb_exists(app_keys, ?))",
					appKey,
					appKey,
				)
			} else {
				query = query.Where("app_key = ?", appKey)
			}
		}
		if name := strings.TrimSpace(req.Name); name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if workspaceScope := normalizeWorkspaceScope(req.WorkspaceScope); workspaceScope != "" {
			switch workspaceScope {
			case "all":
				// 不过滤
			case "personal", "collaboration", "common":
				query = query.Where("(workspace_scope = ? OR workspace_scope = ?)", workspaceScope, "all")
			default:
				query = query.Where("workspace_scope = ?", workspaceScope)
			}
		}
		if status := strings.TrimSpace(req.Status); status != "" {
			query = query.Where("status = ?", status)
		}
	}

	items := make([]user.FeaturePackage, 0)
	err := query.
		Select("id", "app_key", "app_keys", "package_key", "package_type", "name", "description", "workspace_scope", "context_type", "is_builtin", "status", "sort_order", "created_at", "updated_at").
		Order("sort_order ASC, created_at DESC").
		Find(&items).Error
	return items, err
}

func (s *service) GetPackageStats(packageIDs []uuid.UUID) (map[uuid.UUID]int64, map[uuid.UUID]int64, map[uuid.UUID]int64, error) {
	actionCounts, err := s.packageActionRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	menuCounts, err := s.packageMenuRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	collaborationWorkspaceCounts, err := s.collaborationWorkspaceFeaturePackageRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	return actionCounts, menuCounts, collaborationWorkspaceCounts, nil
}

func (s *service) Get(id uuid.UUID) (*user.FeaturePackage, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *service) Create(req *dto.FeaturePackageCreateRequest) (*user.FeaturePackage, error) {
	packageKey := strings.TrimSpace(req.PackageKey)
	if packageKey == "" {
		return nil, errors.New("package_key 不能为空")
	}
	if _, err := s.packageRepo.GetByPackageKey(packageKey); err == nil {
		return nil, ErrFeaturePackageExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	appKeys := normalizeAppKeys(append(req.AppKeys, req.AppKey))
	workspaceScope := normalizeWorkspaceScope(req.WorkspaceScope)
	if workspaceScope == "" {
		workspaceScope = "all"
	}
	contextType := workspaceScopeToContextType(workspaceScope)

	item := &user.FeaturePackage{
		AppKey:         firstOrDefault(appKeys, systemmodels.DefaultAppKey),
		AppKeys:        appKeys,
		PackageKey:     packageKey,
		PackageType:    normalizePackageTypeDefault(req.PackageType, "base"),
		Name:           strings.TrimSpace(req.Name),
		Description:    strings.TrimSpace(req.Description),
		WorkspaceScope: workspaceScope,
		ContextType:    contextType,
		Status:         normalizeStatus(req.Status),
		SortOrder:      req.SortOrder,
	}
	if err := s.packageRepo.Create(item); err != nil {
		return nil, err
	}
	return s.packageRepo.GetByID(item.ID)
}

func (s *service) Update(id uuid.UUID, req *dto.FeaturePackageUpdateRequest) (*permissionrefresh.RefreshStats, error) {
	current, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	updates := map[string]interface{}{
		"updated_at": time.Now(),
		"sort_order": req.SortOrder,
	}
	if packageKey := strings.TrimSpace(req.PackageKey); packageKey != "" && packageKey != current.PackageKey {
		existing, getErr := s.packageRepo.GetByPackageKey(packageKey)
		if getErr == nil && existing != nil && existing.ID != id {
			return nil, ErrFeaturePackageExists
		}
		if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		}
		updates["package_key"] = packageKey
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		updates["name"] = name
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if packageType := normalizePackageType(req.PackageType); packageType != "" {
		if packageType != current.PackageType {
			if packageType == "bundle" {
				actionIDs, getActionErr := s.packageActionRepo.GetKeyIDsByPackageID(id)
				if getActionErr != nil {
					return nil, getActionErr
				}
				if len(actionIDs) > 0 {
					return nil, errors.New("存在已绑定功能权限，请先清空后再改为组合包")
				}
				menuIDs, getMenuErr := s.packageMenuRepo.GetMenuIDsByPackageID(id)
				if getMenuErr != nil {
					return nil, getMenuErr
				}
				if len(menuIDs) > 0 {
					return nil, errors.New("存在已绑定菜单，请先清空后再改为组合包")
				}
			}
			if packageType == "base" {
				childPackageIDs, childErr := s.packageBundleRepo.GetChildPackageIDs(id)
				if childErr != nil {
					return nil, childErr
				}
				if len(childPackageIDs) > 0 {
					return nil, errors.New("组合包仍包含基础包，请先清空组合关系后再改为基础包")
				}
			}
		}
		updates["package_type"] = packageType
	}
	if req.AppKeys != nil || strings.TrimSpace(req.AppKey) != "" {
		appKeys := normalizeAppKeys(append(req.AppKeys, req.AppKey))
		if len(appKeys) > 0 {
			updates["app_key"] = firstOrDefault(appKeys, current.AppKey)
		}
		updates["app_keys"] = appKeys
	}
	if workspaceScope := normalizeWorkspaceScope(req.WorkspaceScope); workspaceScope != "" {
		updates["workspace_scope"] = workspaceScope
		updates["context_type"] = workspaceScopeToContextType(workspaceScope)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		updates["status"] = normalizeStatus(status)
	}
	if err := s.packageRepo.UpdateWithMap(id, updates); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		stats, refreshErr := s.refresher.RefreshByPackageWithStats(id)
		if refreshErr != nil {
			return nil, refreshErr
		}
		_ = s.saveVersionSnapshot(id, "update", nil, "")
		_ = s.recordRiskAudit("feature_package", id.String(), "update", packageSummary(current), packageSummaryFromUpdates(current, updates), refreshStatsSummary(stats), nil, "")
		return &stats, nil
	}
	_ = s.saveVersionSnapshot(id, "update", nil, "")
	_ = s.recordRiskAudit("feature_package", id.String(), "update", packageSummary(current), packageSummaryFromUpdates(current, updates), nil, nil, "")
	return nil, nil
}

func (s *service) Delete(id uuid.UUID) (*permissionrefresh.RefreshStats, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	if item.IsBuiltin {
		return nil, ErrFeaturePackageBuiltin
	}
	collaborationWorkspaceIDs, err := s.collaborationWorkspaceFeaturePackageRepo.GetCollaborationWorkspaceIDsByPackageID(id)
	if err != nil {
		return nil, err
	}
	parentPackageIDs, err := s.packageBundleRepo.GetParentPackageIDs(id)
	if err != nil {
		return nil, err
	}
	if err := s.packageActionRepo.DeleteByPackageID(id); err != nil {
		return nil, err
	}
	if err := s.packageMenuRepo.DeleteByPackageID(id); err != nil {
		return nil, err
	}
	if err := s.collaborationWorkspaceFeaturePackageRepo.DeleteByPackageID(id); err != nil {
		return nil, err
	}
	if err := s.db.Where("package_id = ?", id).Delete(&models.WorkspaceFeaturePackage{}).Error; err != nil {
		return nil, err
	}
	if err := s.rolePackageRepo.DeleteByPackageID(id); err != nil {
		return nil, err
	}
	if err := s.packageBundleRepo.DeleteByPackageID(id); err != nil {
		return nil, err
	}
	if err := s.packageBundleRepo.DeleteByChildPackageID(id); err != nil {
		return nil, err
	}
	if err := s.packageRepo.Delete(id); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		acc := permissionrefresh.RefreshStats{}
		for _, packageID := range parentPackageIDs {
			stats, refreshErr := s.refresher.RefreshByPackageWithStats(packageID)
			if refreshErr != nil {
				return nil, refreshErr
			}
			acc = mergeRefreshStats(acc, stats)
		}
		if refreshErr := s.refresher.RefreshCollaborationWorkspaces(collaborationWorkspaceIDs); refreshErr != nil {
			return nil, refreshErr
		}
		_ = s.recordRiskAudit("feature_package", id.String(), "delete", packageSummary(item), nil, refreshStatsSummary(acc), nil, "")
		return &acc, nil
	}
	for _, collaborationWorkspaceID := range collaborationWorkspaceIDs {
		if _, err := s.boundaryService.RefreshSnapshot(collaborationWorkspaceID); err != nil {
			return nil, err
		}
	}
	_ = s.recordRiskAudit("feature_package", id.String(), "delete", packageSummary(item), nil, nil, nil, "")
	return nil, nil
}

func (s *service) GetPackageChildren(id uuid.UUID, appKey string) ([]uuid.UUID, []user.FeaturePackage, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrFeaturePackageNotFound
		}
		return nil, nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, nil, ErrFeaturePackageNotFound
	}
	if item.PackageType != "bundle" {
		return []uuid.UUID{}, []user.FeaturePackage{}, nil
	}
	childPackageIDs, err := s.packageBundleRepo.GetChildPackageIDs(id)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.packageRepo.GetByIDs(childPackageIDs)
	if err != nil {
		return nil, nil, err
	}
	return filterPackagesForApp(childPackageIDs, items, normalizeAppKey(appKey))
}

func (s *service) SetPackageChildren(id uuid.UUID, childPackageIDs []uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, ErrFeaturePackageNotFound
	}
	if item.PackageType != "bundle" {
		return nil, errors.New("仅组合包允许配置基础包集合")
	}
	childMap, err := s.getPackageMap(childPackageIDs)
	if err != nil {
		return nil, err
	}
	for _, childPackageID := range childPackageIDs {
		if childPackageID == id {
			return nil, errors.New("组合包不能包含自己")
		}
		child, ok := childMap[childPackageID]
		if !ok {
			return nil, ErrFeaturePackageNotFound
		}
		if !packageBelongsToApp(&child, appKey) {
			return nil, errors.New("组合包与基础包必须属于同一应用")
		}
		if child.PackageType != "base" {
			return nil, errors.New("组合包只能包含基础包")
		}
	}
	if err := s.packageBundleRepo.ReplaceChildPackages(id, childPackageIDs); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		stats, refreshErr := s.refresher.RefreshByPackageWithStats(id)
		if refreshErr != nil {
			return nil, refreshErr
		}
		_ = s.saveVersionSnapshot(id, "set_children", nil, "")
		_ = s.recordRiskAudit("feature_package", id.String(), "set_children", nil, ginLikeMap("child_package_ids", childPackageIDs), refreshStatsSummary(stats), nil, "")
		return &stats, nil
	}
	_ = s.saveVersionSnapshot(id, "set_children", nil, "")
	return nil, nil
}

func (s *service) GetPackageKeys(id uuid.UUID, appKey string) ([]uuid.UUID, []user.PermissionKey, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrFeaturePackageNotFound
		}
		return nil, nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, nil, ErrFeaturePackageNotFound
	}
	actionIDs, err := s.packageActionRepo.GetKeyIDsByPackageID(id)
	if err != nil {
		return nil, nil, err
	}
	actions, err := s.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		return nil, nil, err
	}
	return actionIDs, actions, nil
}

func (s *service) SetPackageKeys(id uuid.UUID, actionIDs []uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, ErrFeaturePackageNotFound
	}
	if item.PackageType == "bundle" {
		return nil, errors.New("组合包不允许直接绑定功能权限")
	}
	if len(actionIDs) > 0 {
		actions, getErr := s.actionRepo.GetByIDs(actionIDs)
		if getErr != nil {
			return nil, getErr
		}
		if len(actions) != len(actionIDs) {
			return nil, errors.New("存在无效的功能权限")
		}
		for range actions {
		}
	}
	if err := s.packageActionRepo.ReplacePackageKeys(id, actionIDs); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		stats, refreshErr := s.refresher.RefreshByPackageWithStats(id)
		if refreshErr != nil {
			return nil, refreshErr
		}
		_ = s.saveVersionSnapshot(id, "set_actions", nil, "")
		_ = s.recordRiskAudit("feature_package", id.String(), "set_actions", nil, ginLikeMap("action_ids", actionIDs), refreshStatsSummary(stats), nil, "")
		return &stats, nil
	}
	_ = s.saveVersionSnapshot(id, "set_actions", nil, "")
	return nil, nil
}

func (s *service) GetPackageMenus(id uuid.UUID, appKey string) ([]uuid.UUID, []user.Menu, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrFeaturePackageNotFound
		}
		return nil, nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, nil, ErrFeaturePackageNotFound
	}
	menuIDs, err := s.packageMenuRepo.GetMenuIDsByPackageID(id)
	if err != nil {
		return nil, nil, err
	}
	menus, err := s.menuRepo.GetByIDs(menuIDs)
	if err != nil {
		return nil, nil, err
	}
	filteredMenus := make([]user.Menu, 0, len(menus))
	allowedIDs := make(map[uuid.UUID]struct{}, len(menus))
	normalizedAppKey := normalizeAppKey(appKey)
	allowedAppKeys := normalizeAppKeys(item.AppKeys)
	for _, menu := range menus {
		menuAppKey := normalizeAppKey(menu.AppKey)
		if normalizedAppKey != "" && menuAppKey != normalizedAppKey {
			continue
		}
		if normalizedAppKey == "" && len(allowedAppKeys) > 0 && !containsString(allowedAppKeys, menuAppKey) {
			continue
		}
		filteredMenus = append(filteredMenus, menu)
		allowedIDs[menu.ID] = struct{}{}
	}
	filteredMenuIDs := make([]uuid.UUID, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		if _, ok := allowedIDs[menuID]; ok {
			filteredMenuIDs = append(filteredMenuIDs, menuID)
		}
	}
	return filteredMenuIDs, filteredMenus, nil
}

func (s *service) SetPackageMenus(id uuid.UUID, menuIDs []uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, ErrFeaturePackageNotFound
	}
	if item.PackageType == "bundle" {
		return nil, errors.New("组合包不允许直接绑定菜单")
	}
	if len(menuIDs) > 0 {
		menus, getErr := s.menuRepo.GetByIDs(menuIDs)
		if getErr != nil {
			return nil, getErr
		}
		if len(menus) != len(menuIDs) {
			return nil, errors.New("存在无效的菜单")
		}
		allowedAppKeys := normalizeAppKeys(item.AppKeys)
		normalizedAppKey := normalizeAppKey(appKey)
		for _, menu := range menus {
			menuAppKey := normalizeAppKey(menu.AppKey)
			if normalizedAppKey != "" && menuAppKey != normalizedAppKey {
				return nil, errors.New("功能包与菜单必须属于同一应用")
			}
			if normalizedAppKey == "" && len(allowedAppKeys) > 0 && !containsString(allowedAppKeys, menuAppKey) {
				return nil, errors.New("功能包与菜单必须属于同一应用")
			}
		}
	}
	if err := s.packageMenuRepo.ReplacePackageMenus(id, menuIDs); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		stats, refreshErr := s.refresher.RefreshByPackageWithStats(id)
		if refreshErr != nil {
			return nil, refreshErr
		}
		_ = s.saveVersionSnapshot(id, "set_menus", nil, "")
		_ = s.recordRiskAudit("feature_package", id.String(), "set_menus", nil, ginLikeMap("menu_ids", menuIDs), refreshStatsSummary(stats), nil, "")
		return &stats, nil
	}
	_ = s.saveVersionSnapshot(id, "set_menus", nil, "")
	return nil, nil
}

func (s *service) getPackageMap(packageIDs []uuid.UUID) (map[uuid.UUID]user.FeaturePackage, error) {
	items, err := s.packageRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]user.FeaturePackage, len(items))
	for _, item := range items {
		result[item.ID] = item
	}
	return result, nil
}

func filterPackagesForApp(packageIDs []uuid.UUID, items []user.FeaturePackage, appKey string) ([]uuid.UUID, []user.FeaturePackage, error) {
	filteredItems := make([]user.FeaturePackage, 0, len(items))
	allowedIDs := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if !packageBelongsToApp(&item, appKey) {
			continue
		}
		filteredItems = append(filteredItems, item)
		allowedIDs[item.ID] = struct{}{}
	}
	filteredIDs := make([]uuid.UUID, 0, len(packageIDs))
	for _, packageID := range packageIDs {
		if _, ok := allowedIDs[packageID]; ok {
			filteredIDs = append(filteredIDs, packageID)
		}
	}
	return filteredIDs, filteredItems, nil
}

func packageBelongsToApp(item *user.FeaturePackage, appKey string) bool {
	if item == nil {
		return false
	}
	if len(normalizeAppKeys(item.AppKeys)) == 0 {
		return true
	}
	requested := normalizeAppKey(appKey)
	if requested == "" {
		return true
	}
	appKeys := normalizeAppKeys(item.AppKeys)
	if len(appKeys) > 0 {
		for _, key := range appKeys {
			if key == requested {
				return true
			}
		}
		return false
	}
	return normalizeAppKey(item.AppKey) == requested
}

func normalizeAppKey(value string) string {
	return apppkg.NormalizeAppKey(value)
}

func normalizeAppKeys(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		key := normalizeAppKey(value)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}

func firstOrDefault(values []string, fallback string) string {
	for _, value := range values {
		if key := normalizeAppKey(value); key != "" {
			return key
		}
	}
	return normalizeAppKey(fallback)
}

func normalizeContextType(value string) string {
	switch strings.ReplaceAll(strings.TrimSpace(value), " ", "") {
	case "personal", "collaboration", "common":
		return strings.ReplaceAll(strings.TrimSpace(value), " ", "")
	default:
		return ""
	}
}

func normalizeContextTypeDefault(value, fallback string) string {
	if normalized := normalizeContextType(value); normalized != "" {
		return normalized
	}
	return fallback
}

func normalizeWorkspaceScope(value string) string {
	switch strings.ReplaceAll(strings.TrimSpace(value), " ", "") {
	case "all", "personal", "collaboration":
		return strings.ReplaceAll(strings.TrimSpace(value), " ", "")
	case "common":
		return "all"
	default:
		return ""
	}
}

func normalizeWorkspaceScopeDefault(value, fallback string) string {
	if normalized := normalizeWorkspaceScope(value); normalized != "" {
		return normalized
	}
	switch strings.ReplaceAll(strings.TrimSpace(fallback), " ", "") {
	case "personal", "collaboration", "all":
		return strings.ReplaceAll(strings.TrimSpace(fallback), " ", "")
	case "common":
		return "all"
	default:
		return ""
	}
}

func workspaceScopeFromContextType(contextType string) string {
	switch normalizeContextType(contextType) {
	case "personal":
		return "personal"
	case "collaboration":
		return "collaboration"
	case "common":
		return "all"
	default:
		return ""
	}
}

func workspaceScopeToContextType(workspaceScope string) string {
	switch normalizeWorkspaceScope(workspaceScope) {
	case "personal":
		return "personal"
	case "collaboration":
		return "collaboration"
	case "all":
		return "common"
	default:
		return ""
	}
}

func normalizeStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "disabled":
		return "disabled"
	default:
		return "normal"
	}
}

func normalizePackageType(value string) string {
	switch strings.TrimSpace(value) {
	case "base", "bundle":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizePackageTypeDefault(value, fallback string) string {
	if normalized := normalizePackageType(value); normalized != "" {
		return normalized
	}
	return fallback
}

func supportsCollaborationWorkspaceContext(contextType string) bool {
	return true
}

func contextSupportsAction(packageContextType, actionContextType string) bool {
	return true
}

func contextSupportsChildPackage(bundleContextType, childContextType string) bool {
	return true
}

func mergeUUIDSlice(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, id := range group {
			if id == uuid.Nil {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}
