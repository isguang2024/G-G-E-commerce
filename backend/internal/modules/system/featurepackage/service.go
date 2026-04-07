package featurepackage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/appscope"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacefeaturebinding"
)

var (
	ErrFeaturePackageNotFound = errors.New("feature package not found")
	ErrFeaturePackageExists   = errors.New("feature package already exists")
	ErrFeaturePackageBuiltin  = errors.New("feature package is builtin")
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
		return nil, 0, errors.New("invalid request")
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
		return nil, errors.New("invalid request")
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
					"(app_key = ? OR COALESCE(jsonb_array_length(app_keys), 0) = 0 OR app_keys ? ?)",
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

func (s *service) GetCollaborationWorkspacePackages(collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, []user.FeaturePackage, error) {
	packageIDs, err := appscope.PackageIDsByCollaborationWorkspace(s.db, collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.packageRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, nil, err
	}
	return filterPackagesForApp(packageIDs, items, normalizeAppKey(appKey))
}

func (s *service) GetPackageCollaborationWorkspaces(id uuid.UUID, appKey string) ([]uuid.UUID, error) {
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
	workspaceCollaborationWorkspaceIDs, err := s.getWorkspaceCollaborationWorkspaceIDsByPackageID(id, appKey)
	if err != nil {
		return nil, err
	}
	legacyCollaborationWorkspaceIDs, err := s.collaborationWorkspaceFeaturePackageRepo.GetCollaborationWorkspaceIDsByPackageID(id)
	if err != nil {
		return nil, err
	}
	return mergeUUIDSlice(workspaceCollaborationWorkspaceIDs, legacyCollaborationWorkspaceIDs), nil
}

func (s *service) SetPackageCollaborationWorkspaces(id uuid.UUID, collaborationWorkspaceIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error) {
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
	currentCollaborationWorkspaceIDs, err := s.GetPackageCollaborationWorkspaces(id, appKey)
	if err != nil {
		return nil, err
	}
	desired := make(map[uuid.UUID]struct{}, len(collaborationWorkspaceIDs))
	affected := make(map[uuid.UUID]struct{}, len(currentCollaborationWorkspaceIDs)+len(collaborationWorkspaceIDs))
	for _, collaborationWorkspaceID := range currentCollaborationWorkspaceIDs {
		affected[collaborationWorkspaceID] = struct{}{}
	}
	collaborationWorkspaceMap, err := s.getCollaborationWorkspaceMap(collaborationWorkspaceIDs)
	if err != nil {
		return nil, err
	}
	for _, collaborationWorkspaceID := range collaborationWorkspaceIDs {
		if _, ok := collaborationWorkspaceMap[collaborationWorkspaceID]; !ok {
			return nil, errors.New("存在无效的协作空间")
		}
		desired[collaborationWorkspaceID] = struct{}{}
		affected[collaborationWorkspaceID] = struct{}{}
	}
	for collaborationWorkspaceID := range affected {
		packageIDs, err := appscope.PackageIDsByCollaborationWorkspace(s.db, collaborationWorkspaceID, item.AppKey)
		if err != nil {
			return nil, err
		}
		nextPackageIDs := make([]uuid.UUID, 0, len(packageIDs)+1)
		seen := make(map[uuid.UUID]struct{}, len(packageIDs)+1)
		for _, packageID := range packageIDs {
			if packageID == id {
				continue
			}
			if _, ok := seen[packageID]; ok {
				continue
			}
			seen[packageID] = struct{}{}
			nextPackageIDs = append(nextPackageIDs, packageID)
		}
		if _, ok := desired[collaborationWorkspaceID]; ok {
			if _, exists := seen[id]; !exists {
				nextPackageIDs = append(nextPackageIDs, id)
			}
		}
		if err := appscope.ReplaceCollaborationWorkspacePackagesInApp(s.db, collaborationWorkspaceID, item.AppKey, nextPackageIDs, grantedBy); err != nil {
			return nil, err
		}
		if err := s.replaceWorkspacePackagesForCollaborationWorkspaceID(collaborationWorkspaceID, item.AppKey, nextPackageIDs); err != nil {
			return nil, err
		}
		if s.refresher != nil {
			if err := s.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := s.boundaryService.RefreshSnapshot(collaborationWorkspaceID, item.AppKey); err != nil {
			return nil, err
		}
	}
	stats := permissionrefresh.RefreshStats{
		CollaborationWorkspaceCount: len(affected),
	}
	_ = s.saveVersionSnapshot(id, "set_collaboration_workspaces", grantedBy, "")
	_ = s.recordRiskAudit("feature_package", id.String(), "set_collaboration_workspaces", nil, ginLikeMap("collaboration_workspace_ids", collaborationWorkspaceIDs), refreshStatsSummary(stats), grantedBy, "")
	return &stats, nil
}

func (s *service) SetCollaborationWorkspacePackages(collaborationWorkspaceID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	packageMap, err := s.getPackageMap(packageIDs)
	if err != nil {
		return nil, err
	}
	for _, packageID := range packageIDs {
		item, ok := packageMap[packageID]
		if !ok {
			return nil, ErrFeaturePackageNotFound
		}
		if !packageBelongsToApp(&item, normalizedAppKey) {
			return nil, errors.New("功能包不在当前 App 的适用范围内")
		}
	}
	if err := appscope.ReplaceCollaborationWorkspacePackagesInApp(s.db, collaborationWorkspaceID, normalizedAppKey, packageIDs, grantedBy); err != nil {
		return nil, err
	}
	if err := s.replaceWorkspacePackagesForCollaborationWorkspaceID(collaborationWorkspaceID, normalizedAppKey, packageIDs); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			return nil, err
		}
		stats := permissionrefresh.RefreshStats{CollaborationWorkspaceCount: 1}
		_ = s.recordRiskAudit("collaboration_workspace_feature_package", collaborationWorkspaceID.String(), "set_collaboration_workspace_packages", nil, ginLikeMap("package_ids", packageIDs), refreshStatsSummary(stats), grantedBy, "")
		return &stats, nil
	}
	_, err = s.boundaryService.RefreshSnapshot(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	stats := permissionrefresh.RefreshStats{CollaborationWorkspaceCount: 1}
	_ = s.recordRiskAudit("collaboration_workspace_feature_package", collaborationWorkspaceID.String(), "set_collaboration_workspace_packages", nil, ginLikeMap("package_ids", packageIDs), refreshStatsSummary(stats), grantedBy, "")
	return &stats, nil
}

func (s *service) getWorkspacePackageIDsByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var workspace models.Workspace
	if err := s.db.
		Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, collaborationWorkspaceID).
		First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}

	var packageIDs []uuid.UUID
	if err := s.db.Model(&models.WorkspaceFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
		Where("workspace_feature_packages.workspace_id = ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", workspace.ID, true).
		Where("feature_packages.deleted_at IS NULL").
		Where(func(tx *gorm.DB) *gorm.DB {
			requestedAppKey := normalizeAppKey(appKey)
			if requestedAppKey == "" {
				return tx
			}
			return tx.Where(
				"feature_packages.app_key = ? OR COALESCE(jsonb_array_length(feature_packages.app_keys), 0) = 0 OR feature_packages.app_keys ? ?",
				requestedAppKey,
				requestedAppKey,
			)
		}).
		Distinct("workspace_feature_packages.package_id").
		Pluck("package_id", &packageIDs).Error; err != nil {
		return nil, err
	}
	return packageIDs, nil
}

func (s *service) getWorkspaceCollaborationWorkspaceIDsByPackageID(packageID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	type row struct {
		SourceCollaborationWorkspaceID *uuid.UUID `gorm:"column:collaboration_workspace_id"`
	}
	rows := make([]row, 0)
	if err := s.db.Model(&models.WorkspaceFeaturePackage{}).
		Select("workspaces.collaboration_workspace_id").
		Joins("JOIN workspaces ON workspaces.id = workspace_feature_packages.workspace_id").
		Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
		Where("workspace_feature_packages.package_id = ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", packageID, true).
		Where("workspaces.workspace_type = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypeCollaboration).
		Where("feature_packages.deleted_at IS NULL").
		Where(func(tx *gorm.DB) *gorm.DB {
			requestedAppKey := normalizeAppKey(appKey)
			if requestedAppKey == "" {
				return tx
			}
			return tx.Where(
				"feature_packages.app_key = ? OR COALESCE(jsonb_array_length(feature_packages.app_keys), 0) = 0 OR feature_packages.app_keys ? ?",
				requestedAppKey,
				requestedAppKey,
			)
		}).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	collaborationWorkspaceIDs := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, item := range rows {
		if item.SourceCollaborationWorkspaceID == nil || *item.SourceCollaborationWorkspaceID == uuid.Nil {
			continue
		}
		if _, exists := seen[*item.SourceCollaborationWorkspaceID]; exists {
			continue
		}
		seen[*item.SourceCollaborationWorkspaceID] = struct{}{}
		collaborationWorkspaceIDs = append(collaborationWorkspaceIDs, *item.SourceCollaborationWorkspaceID)
	}
	return collaborationWorkspaceIDs, nil
}

func (s *service) replaceWorkspacePackagesForCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID, appKey string, packageIDs []uuid.UUID) error {
	var workspace models.Workspace
	if err := s.db.
		Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, collaborationWorkspaceID).
		First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	scopedIDs, err := appscope.FilterPackageIDs(s.db, appKey, packageIDs)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workspace_id = ?", workspace.ID).Delete(&models.WorkspaceFeaturePackage{}).Error; err != nil {
			return err
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		items := make([]models.WorkspaceFeaturePackage, 0, len(scopedIDs))
		for _, packageID := range scopedIDs {
			items = append(items, models.WorkspaceFeaturePackage{
				WorkspaceID: workspace.ID,
				PackageID:   packageID,
				Enabled:     true,
			})
		}
		return tx.Create(&items).Error
	})
}

func (s *service) GetRelationTree(workspaceScope, keyword string) (*FeaturePackageRelationTree, error) {
	packages, err := s.ListOptions(&dto.FeaturePackageListRequest{
		WorkspaceScope: normalizeWorkspaceScope(workspaceScope),
		Keyword:        strings.TrimSpace(keyword),
	})
	if err != nil {
		return nil, err
	}
	nodes := make(map[uuid.UUID]FeaturePackageRelationNode, len(packages))
	for _, pkg := range packages {
		nodes[pkg.ID] = FeaturePackageRelationNode{
			ID:             pkg.ID,
			PackageKey:     pkg.PackageKey,
			Name:           pkg.Name,
			PackageType:    pkg.PackageType,
			WorkspaceScope: pkg.WorkspaceScope,
			AppKeys:        pkg.AppKeys,
			Status:         pkg.Status,
		}
	}
	if len(nodes) == 0 {
		return &FeaturePackageRelationTree{
			Roots:             []FeaturePackageRelationNode{},
			CycleDependencies: [][]string{},
			IsolatedBaseKeys:  []string{},
		}, nil
	}

	type relationRow struct {
		PackageID      uuid.UUID
		ChildPackageID uuid.UUID
	}
	rows := make([]relationRow, 0)
	packageIDs := make([]uuid.UUID, 0, len(nodes))
	for id := range nodes {
		packageIDs = append(packageIDs, id)
	}
	if err := s.db.Model(&user.FeaturePackageBundle{}).
		Select("package_id", "child_package_id").
		Where("package_id IN ? AND child_package_id IN ?", packageIDs, packageIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	childrenMap := make(map[uuid.UUID][]uuid.UUID, len(rows))
	parentCount := make(map[uuid.UUID]int, len(nodes))
	for _, row := range rows {
		childrenMap[row.PackageID] = append(childrenMap[row.PackageID], row.ChildPackageID)
		parentCount[row.ChildPackageID]++
	}
	for id, node := range nodes {
		node.ReferenceCount = parentCount[id]
		nodes[id] = node
	}

	visited := make(map[uuid.UUID]bool, len(nodes))
	pathMark := make(map[uuid.UUID]bool, len(nodes))
	cycleSet := make(map[string]struct{})
	cycles := make([][]string, 0)
	var buildTree func(id uuid.UUID, path []uuid.UUID) FeaturePackageRelationNode
	buildTree = func(id uuid.UUID, path []uuid.UUID) FeaturePackageRelationNode {
		node := nodes[id]
		if pathMark[id] {
			cycle := append(path, id)
			keys := make([]string, 0, len(cycle))
			for _, item := range cycle {
				if value, ok := nodes[item]; ok {
					keys = append(keys, value.PackageKey)
				}
			}
			signature := strings.Join(keys, " -> ")
			if signature != "" {
				if _, exists := cycleSet[signature]; !exists {
					cycleSet[signature] = struct{}{}
					cycles = append(cycles, keys)
				}
			}
			return node
		}
		pathMark[id] = true
		visited[id] = true
		children := childrenMap[id]
		if len(children) > 0 {
			node.Children = make([]FeaturePackageRelationNode, 0, len(children))
			for _, childID := range children {
				if _, ok := nodes[childID]; !ok {
					continue
				}
				node.Children = append(node.Children, buildTree(childID, append(path, id)))
			}
		}
		pathMark[id] = false
		return node
	}

	roots := make([]FeaturePackageRelationNode, 0)
	for id := range nodes {
		if parentCount[id] == 0 {
			roots = append(roots, buildTree(id, nil))
		}
	}
	for id := range nodes {
		if visited[id] {
			continue
		}
		roots = append(roots, buildTree(id, nil))
	}

	isolatedBaseKeys := make([]string, 0)
	for id, node := range nodes {
		if node.PackageType != "base" {
			continue
		}
		if len(childrenMap[id]) > 0 {
			continue
		}
		if parentCount[id] > 0 {
			continue
		}
		isolatedBaseKeys = append(isolatedBaseKeys, node.PackageKey)
	}

	return &FeaturePackageRelationTree{
		Roots:             roots,
		CycleDependencies: cycles,
		IsolatedBaseKeys:  isolatedBaseKeys,
	}, nil
}

func (s *service) GetImpactPreview(id uuid.UUID) (*FeaturePackageImpactPreview, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}

	result := &FeaturePackageImpactPreview{PackageID: id}
	type countRow struct {
		Total int64
	}

	var row countRow
	if err := s.db.Model(&user.RoleFeaturePackage{}).Where("package_id = ? AND enabled = ?", id, true).Distinct("role_id").Count(&row.Total).Error; err != nil {
		return nil, err
	}
	result.RoleCount = row.Total
	workspaceCollaborationWorkspaceIDs, err := workspacefeaturebinding.ListCollaborationWorkspaceIDsByPackageIDs(s.db, []uuid.UUID{id}, item.AppKey)
	if err != nil {
		return nil, err
	}
	var legacyCollaborationWorkspaceIDs []uuid.UUID
	if err := s.db.Model(&user.CollaborationWorkspaceFeaturePackage{}).Where("package_id = ? AND enabled = ?", id, true).Distinct("collaboration_workspace_id").Pluck("collaboration_workspace_id", &legacyCollaborationWorkspaceIDs).Error; err != nil {
		return nil, err
	}
	result.CollaborationWorkspaceCount = int64(len(mergeUUIDSlice(workspaceCollaborationWorkspaceIDs, legacyCollaborationWorkspaceIDs)))
	workspaceUserIDs, err := workspacefeaturebinding.ListPlatformUserIDsByPackageIDs(s.db, []uuid.UUID{id}, item.AppKey)
	if err != nil {
		return nil, err
	}
	var legacyUserIDs []uuid.UUID
	if err := s.db.Model(&user.UserFeaturePackage{}).Where("package_id = ? AND enabled = ?", id, true).Distinct("user_id").Pluck("user_id", &legacyUserIDs).Error; err != nil {
		return nil, err
	}
	result.UserCount = int64(len(mergeUUIDSlice(workspaceUserIDs, legacyUserIDs)))
	if err := s.db.Model(&user.FeaturePackageMenu{}).Where("package_id = ?", id).Count(&row.Total).Error; err != nil {
		return nil, err
	}
	result.MenuCount = row.Total
	if err := s.db.Model(&user.FeaturePackageKey{}).Where("package_id = ?", id).Count(&row.Total).Error; err != nil {
		return nil, err
	}
	result.ActionCount = row.Total
	return result, nil
}

func (s *service) ListVersions(id uuid.UUID, current, size int) ([]user.FeaturePackageVersion, int64, error) {
	if _, err := s.packageRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrFeaturePackageNotFound
		}
		return nil, 0, err
	}
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	query := s.db.Model(&user.FeaturePackageVersion{}).Where("package_id = ?", id)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]user.FeaturePackageVersion, 0)
	if err := query.Order("version_no DESC").Offset((current - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) Rollback(id uuid.UUID, versionID uuid.UUID, operatorID *uuid.UUID, requestID string) (*permissionrefresh.RefreshStats, error) {
	pkg, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	var version user.FeaturePackageVersion
	if err := s.db.Where("id = ? AND package_id = ?", versionID, id).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("版本不存在")
		}
		return nil, err
	}
	snapshot := version.Snapshot
	parseUUIDList := func(key string) []uuid.UUID {
		raw, ok := snapshot[key]
		if !ok {
			return []uuid.UUID{}
		}
		switch values := raw.(type) {
		case []interface{}:
			result := make([]uuid.UUID, 0, len(values))
			for _, item := range values {
				parsed, parseErr := uuid.Parse(strings.TrimSpace(fmt.Sprint(item)))
				if parseErr == nil {
					result = append(result, parsed)
				}
			}
			return result
		case []string:
			result := make([]uuid.UUID, 0, len(values))
			for _, item := range values {
				parsed, parseErr := uuid.Parse(strings.TrimSpace(item))
				if parseErr == nil {
					result = append(result, parsed)
				}
			}
			return result
		default:
			return []uuid.UUID{}
		}
	}

	childIDs := parseUUIDList("child_package_ids")
	actionIDs := parseUUIDList("action_ids")
	menuIDs := parseUUIDList("menu_ids")
	collaborationWorkspaceIDs := parseUUIDList("collaboration_workspace_ids")
	updates := map[string]interface{}{
		"package_key":     strings.TrimSpace(fmt.Sprint(snapshot["package_key"])),
		"package_type":    normalizePackageTypeDefault(fmt.Sprint(snapshot["package_type"]), pkg.PackageType),
		"name":            strings.TrimSpace(fmt.Sprint(snapshot["name"])),
		"description":     strings.TrimSpace(fmt.Sprint(snapshot["description"])),
		"workspace_scope": normalizeWorkspaceScopeDefault(fmt.Sprint(snapshot["workspace_scope"]), pkg.WorkspaceScope),
		"context_type":    normalizeContextTypeDefault(fmt.Sprint(snapshot["context_type"]), pkg.ContextType),
		"app_keys":        snapshot["app_keys"],
		"status":          normalizeStatus(fmt.Sprint(snapshot["status"])),
		"sort_order":      intFromSnapshot(snapshot["sort_order"], pkg.SortOrder),
		"updated_at":      time.Now(),
	}
	if err := s.packageRepo.UpdateWithMap(id, updates); err != nil {
		return nil, err
	}
	if err := s.packageBundleRepo.ReplaceChildPackages(id, childIDs); err != nil {
		return nil, err
	}
	if err := s.packageActionRepo.ReplacePackageKeys(id, actionIDs); err != nil {
		return nil, err
	}
	if err := s.packageMenuRepo.ReplacePackageMenus(id, menuIDs); err != nil {
		return nil, err
	}
	if err := s.syncPackageCollaborationWorkspacesBySet(id, collaborationWorkspaceIDs, operatorID); err != nil {
		return nil, err
	}

	var stats permissionrefresh.RefreshStats
	if s.refresher != nil {
		ref, refreshErr := s.refresher.RefreshByPackageWithStats(id)
		if refreshErr != nil {
			return nil, refreshErr
		}
		stats = ref
	}
	_ = s.saveVersionSnapshot(id, "rollback", operatorID, requestID)
	_ = s.recordRiskAudit("feature_package", id.String(), "rollback", nil, map[string]interface{}{
		"rollback_version_id": versionID.String(),
		"rollback_version_no": version.VersionNo,
	}, refreshStatsSummary(stats), operatorID, requestID)
	return &stats, nil
}

func (s *service) ListRiskAudits(id uuid.UUID, current, size int) ([]user.RiskOperationAudit, int64, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	query := s.db.Model(&user.RiskOperationAudit{}).
		Where("object_type = ? AND object_id = ?", "feature_package", id.String())
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]user.RiskOperationAudit, 0)
	if err := query.Order("created_at DESC").Offset((current - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) saveVersionSnapshot(packageID uuid.UUID, changeType string, operatorID *uuid.UUID, requestID string) error {
	pkg, err := s.packageRepo.GetByID(packageID)
	if err != nil {
		return err
	}
	childIDs, err := s.packageBundleRepo.GetChildPackageIDs(packageID)
	if err != nil {
		return err
	}
	actionIDs, err := s.packageActionRepo.GetKeyIDsByPackageID(packageID)
	if err != nil {
		return err
	}
	menuIDs, err := s.packageMenuRepo.GetMenuIDsByPackageID(packageID)
	if err != nil {
		return err
	}
	collaborationWorkspaceIDs, err := s.collaborationWorkspaceFeaturePackageRepo.GetCollaborationWorkspaceIDsByPackageID(packageID)
	if err != nil {
		return err
	}

	var maxVersion int64
	if err := s.db.Model(&user.FeaturePackageVersion{}).Where("package_id = ?", packageID).Select("COALESCE(MAX(version_no), 0)").Scan(&maxVersion).Error; err != nil {
		return err
	}
	item := &user.FeaturePackageVersion{
		PackageID:  packageID,
		VersionNo:  int(maxVersion) + 1,
		ChangeType: strings.TrimSpace(changeType),
		Snapshot: map[string]interface{}{
			"package_id":                  packageID.String(),
			"package_key":                 pkg.PackageKey,
			"package_type":                pkg.PackageType,
			"name":                        pkg.Name,
			"description":                 pkg.Description,
			"workspace_scope":             pkg.WorkspaceScope,
			"context_type":                pkg.ContextType,
			"app_keys":                    pkg.AppKeys,
			"status":                      pkg.Status,
			"sort_order":                  pkg.SortOrder,
			"child_package_ids":           uuidSliceToStrings(childIDs),
			"action_ids":                  uuidSliceToStrings(actionIDs),
			"menu_ids":                    uuidSliceToStrings(menuIDs),
			"collaboration_workspace_ids": uuidSliceToStrings(collaborationWorkspaceIDs),
			"snapshot_createdAt":          time.Now().Format(time.RFC3339),
		},
		OperatorID: operatorID,
		RequestID:  strings.TrimSpace(requestID),
	}
	return s.db.Create(item).Error
}

func (s *service) syncPackageCollaborationWorkspacesBySet(id uuid.UUID, collaborationWorkspaceIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	currentCollaborationWorkspaceIDs, err := s.collaborationWorkspaceFeaturePackageRepo.GetCollaborationWorkspaceIDsByPackageID(id)
	if err != nil {
		return err
	}
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		return err
	}
	desired := make(map[uuid.UUID]struct{}, len(collaborationWorkspaceIDs))
	affected := make(map[uuid.UUID]struct{}, len(currentCollaborationWorkspaceIDs)+len(collaborationWorkspaceIDs))
	for _, collaborationWorkspaceID := range currentCollaborationWorkspaceIDs {
		affected[collaborationWorkspaceID] = struct{}{}
	}
	for _, collaborationWorkspaceID := range collaborationWorkspaceIDs {
		desired[collaborationWorkspaceID] = struct{}{}
		affected[collaborationWorkspaceID] = struct{}{}
	}
	for collaborationWorkspaceID := range affected {
		packageIDs, packageErr := appscope.PackageIDsByCollaborationWorkspace(s.db, collaborationWorkspaceID, item.AppKey)
		if packageErr != nil {
			return packageErr
		}
		nextPackageIDs := make([]uuid.UUID, 0, len(packageIDs)+1)
		seen := make(map[uuid.UUID]struct{}, len(packageIDs)+1)
		for _, packageID := range packageIDs {
			if packageID == id {
				continue
			}
			if _, ok := seen[packageID]; ok {
				continue
			}
			seen[packageID] = struct{}{}
			nextPackageIDs = append(nextPackageIDs, packageID)
		}
		if _, ok := desired[collaborationWorkspaceID]; ok {
			nextPackageIDs = append(nextPackageIDs, id)
		}
		if err := appscope.ReplaceCollaborationWorkspacePackagesInApp(s.db, collaborationWorkspaceID, item.AppKey, nextPackageIDs, grantedBy); err != nil {
			return err
		}
	}
	return nil
}

func mergeRefreshStats(left, right permissionrefresh.RefreshStats) permissionrefresh.RefreshStats {
	result := permissionrefresh.RefreshStats{
		RequestedPackageCount:       left.RequestedPackageCount + right.RequestedPackageCount,
		ImpactedPackageCount:        left.ImpactedPackageCount + right.ImpactedPackageCount,
		RoleCount:                   left.RoleCount + right.RoleCount,
		CollaborationWorkspaceCount: left.CollaborationWorkspaceCount + right.CollaborationWorkspaceCount,
		UserCount:                   left.UserCount + right.UserCount,
		ElapsedMilliseconds:         left.ElapsedMilliseconds + right.ElapsedMilliseconds,
	}
	if right.FinishedAt.After(left.FinishedAt) {
		result.FinishedAt = right.FinishedAt
	} else {
		result.FinishedAt = left.FinishedAt
	}
	return result
}

func packageSummary(item *user.FeaturePackage) map[string]interface{} {
	if item == nil {
		return nil
	}
	return map[string]interface{}{
		"id":              item.ID.String(),
		"package_key":     item.PackageKey,
		"package_type":    item.PackageType,
		"name":            item.Name,
		"workspace_scope": item.WorkspaceScope,
		"context_type":    item.ContextType,
		"app_keys":        item.AppKeys,
		"status":          item.Status,
		"sort_order":      item.SortOrder,
	}
}

func packageSummaryFromUpdates(base *user.FeaturePackage, updates map[string]interface{}) map[string]interface{} {
	if base == nil {
		return nil
	}
	result := packageSummary(base)
	for key, value := range updates {
		if key == "updated_at" {
			continue
		}
		result[key] = value
	}
	return result
}

func refreshStatsSummary(stats permissionrefresh.RefreshStats) map[string]interface{} {
	return map[string]interface{}{
		"requested_package_count":       stats.RequestedPackageCount,
		"impacted_package_count":        stats.ImpactedPackageCount,
		"role_count":                    stats.RoleCount,
		"collaboration_workspace_count": stats.CollaborationWorkspaceCount,
		"user_count":                    stats.UserCount,
		"elapsed_milliseconds":          stats.ElapsedMilliseconds,
	}
}

func uuidSliceToStrings(items []uuid.UUID) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.String())
	}
	return result
}

func ginLikeMap(key string, values []uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		key: uuidSliceToStrings(values),
	}
}

func intFromSnapshot(value interface{}, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		parsed := strings.TrimSpace(v)
		if parsed == "" {
			return fallback
		}
		var num int
		_, err := fmt.Sscanf(parsed, "%d", &num)
		if err != nil {
			return fallback
		}
		return num
	default:
		return fallback
	}
}

func (s *service) recordRiskAudit(
	objectType string,
	objectID string,
	operationType string,
	beforeSummary map[string]interface{},
	afterSummary map[string]interface{},
	impactSummary map[string]interface{},
	operatorID *uuid.UUID,
	requestID string,
) error {
	item := &user.RiskOperationAudit{
		ObjectType:    strings.TrimSpace(objectType),
		ObjectID:      strings.TrimSpace(objectID),
		OperationType: strings.TrimSpace(operationType),
		BeforeSummary: beforeSummary,
		AfterSummary:  afterSummary,
		ImpactSummary: impactSummary,
		OperatorID:    operatorID,
		RequestID:     strings.TrimSpace(requestID),
	}
	return s.db.Create(item).Error
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

func (s *service) getCollaborationWorkspaceMap(collaborationWorkspaceIDs []uuid.UUID) (map[uuid.UUID]struct{}, error) {
	items, err := s.collaborationWorkspaceRepo.GetByIDs(collaborationWorkspaceIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		result[item.ID] = struct{}{}
	}
	return result, nil
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
