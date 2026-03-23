package role

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
)

var (
	ErrRoleNotFound            = errors.New("role not found")
	ErrRoleCodeExists          = errors.New("role code already exists")
	ErrSystemRoleCannotDelete  = errors.New("system role cannot be deleted")
	ErrTeamRoleActionReadonly  = errors.New("team role actions are managed by team boundary")
	ErrTenantRoleManagedByTeam = errors.New("tenant role is managed in team context")
)

type RoleService interface {
	List(req *dto.RoleListRequest) ([]user.Role, int64, error)
	Get(id uuid.UUID) (*user.Role, error)
	Create(req *dto.RoleCreateRequest) (*user.Role, error)
	Update(id uuid.UUID, req *dto.RoleUpdateRequest) error
	Delete(id uuid.UUID) error
	GetRolePackages(roleID uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error)
	SetRolePackages(roleID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
	GetRoleMenuBoundary(roleID uuid.UUID) (*RoleMenuBoundary, error)
	GetRoleMenuIDs(roleID uuid.UUID) ([]uuid.UUID, error)
	SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error
	GetRoleActionBoundary(roleID uuid.UUID) (*RoleActionBoundary, error)
	GetRoleActions(roleID uuid.UUID) ([]user.RoleActionPermission, error)
	SetRoleActions(roleID uuid.UUID, actions []user.RoleActionPermission) error
	GetRoleDataPermissions(roleID uuid.UUID) ([]user.RoleDataPermission, []string, []DataPermissionScopeOption, error)
	SetRoleDataPermissions(roleID uuid.UUID, permissions []user.RoleDataPermission) error
}

type RoleMenuBoundary struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	AvailableMenuIDs   []uuid.UUID
	HiddenMenuIDs      []uuid.UUID
	EffectiveMenuIDs   []uuid.UUID
	MenuSourceMap      map[uuid.UUID][]uuid.UUID
}

type RoleActionBoundary struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	AvailableActionIDs []uuid.UUID
	DisabledActionIDs  []uuid.UUID
	EffectiveActionIDs []uuid.UUID
	ActionSourceMap    map[uuid.UUID][]uuid.UUID
}

type DataPermissionScopeOption struct {
	Code string
	Name string
}

type roleService struct {
	roleRepo               user.RoleRepository
	rolePackageRepo        user.RoleFeaturePackageRepository
	featurePkgRepo         user.FeaturePackageRepository
	packageActionRepo      user.FeaturePackageActionRepository
	packageMenuRepo        user.FeaturePackageMenuRepository
	packageBundleRepo      user.FeaturePackageBundleRepository
	roleHiddenMenuRepo     user.RoleHiddenMenuRepository
	roleDisabledActionRepo user.RoleDisabledActionRepository
	roleDataRepo           user.RoleDataPermissionRepository
	actionRepo             user.PermissionActionRepository
	roleSnapshotService    platformroleaccess.Service
	refresher              permissionrefresh.Service
	logger                 *zap.Logger
}

func NewRoleService(
	roleRepo user.RoleRepository,
	rolePackageRepo user.RoleFeaturePackageRepository,
	featurePkgRepo user.FeaturePackageRepository,
	packageActionRepo user.FeaturePackageActionRepository,
	packageMenuRepo user.FeaturePackageMenuRepository,
	packageBundleRepo user.FeaturePackageBundleRepository,
	roleHiddenMenuRepo user.RoleHiddenMenuRepository,
	roleDisabledActionRepo user.RoleDisabledActionRepository,
	roleDataRepo user.RoleDataPermissionRepository,
	actionRepo user.PermissionActionRepository,
	roleSnapshotService platformroleaccess.Service,
	refresher permissionrefresh.Service,
	logger *zap.Logger,
) RoleService {
	return &roleService{
		roleRepo:               roleRepo,
		rolePackageRepo:        rolePackageRepo,
		featurePkgRepo:         featurePkgRepo,
		packageActionRepo:      packageActionRepo,
		packageMenuRepo:        packageMenuRepo,
		packageBundleRepo:      packageBundleRepo,
		roleHiddenMenuRepo:     roleHiddenMenuRepo,
		roleDisabledActionRepo: roleDisabledActionRepo,
		roleDataRepo:           roleDataRepo,
		actionRepo:             actionRepo,
		roleSnapshotService:    roleSnapshotService,
		refresher:              refresher,
		logger:                 logger,
	}
}

func (s *roleService) List(req *dto.RoleListRequest) ([]user.Role, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.roleRepo.ListByPage(
		offset,
		req.Size,
		req.RoleCode,
		req.RoleName,
		req.Description,
		req.StartTime,
		req.EndTime,
		req.Enabled,
	)
}

func (s *roleService) Get(id uuid.UUID) (*user.Role, error) {
	return s.roleRepo.GetByID(id)
}

func (s *roleService) Create(req *dto.RoleCreateRequest) (*user.Role, error) {
	if _, err := s.roleRepo.GetByCode(req.Code); err == nil {
		return nil, ErrRoleCodeExists
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	status := "normal"
	if req.Status != "" {
		status = req.Status
	}
	role := &user.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Priority:    req.Priority,
		Status:      status,
	}
	if err := database.DB.Create(role).Error; err != nil {
		return nil, err
	}
	return s.roleRepo.GetByID(role.ID)
}

func (s *roleService) Update(id uuid.UUID, req *dto.RoleUpdateRequest) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.TenantID != nil {
		return ErrTenantRoleManagedByTeam
	}
	updates := make(map[string]interface{})
	if req.Code != "" && req.Code != role.Code {
		existingRole, err := s.roleRepo.GetByCode(req.Code)
		if err == nil && existingRole != nil && existingRole.ID != id {
			return ErrRoleCodeExists
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		updates["code"] = req.Code
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["sort_order"] = req.SortOrder
	updates["priority"] = req.Priority
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if len(updates) == 0 {
		s.logger.Info("无需更新角色", zap.String("roleId", id.String()))
		return nil
	}
	updates["updated_at"] = time.Now()
	s.logger.Info("更新角色字段", zap.String("roleId", id.String()), zap.Any("updates", updates))
	if err := s.roleRepo.UpdateWithMap(id, updates); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPlatformRole(id)
	}
	return nil
}

func (s *roleService) Delete(id uuid.UUID) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.TenantID != nil {
		return ErrTenantRoleManagedByTeam
	}
	if role.Code == "admin" || role.Code == "team_admin" || role.Code == "team_member" {
		return ErrSystemRoleCannotDelete
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&user.RoleFeaturePackage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&user.RoleHiddenMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&user.RoleDisabledAction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&user.RoleDataPermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user.Role{}, id).Error
	})
}

func (s *roleService) GetRolePackages(roleID uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, ErrRoleNotFound
		}
		return nil, nil, err
	}
	if role.TenantID != nil {
		return nil, nil, ErrTenantRoleManagedByTeam
	}
	packageIDs, err := s.rolePackageRepo.GetPackageIDsByRoleID(roleID)
	if err != nil {
		return nil, nil, err
	}
	packages, err := s.featurePkgRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, nil, err
	}
	return packageIDs, packages, nil
}

func (s *roleService) SetRolePackages(roleID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.TenantID != nil {
		return ErrTenantRoleManagedByTeam
	}
	if len(packageIDs) > 0 {
		packages, err := s.featurePkgRepo.GetByIDs(packageIDs)
		if err != nil {
			return err
		}
		if len(packages) != len(packageIDs) {
			return errors.New("存在无效的功能包")
		}
		for _, item := range packages {
			if !supportsPlatformPackage(item.ContextType) {
				return errors.New("仅支持为平台角色绑定平台功能包")
			}
		}
	}
	if err := s.rolePackageRepo.ReplaceRolePackages(roleID, packageIDs, grantedBy); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPlatformRole(roleID)
	}
	return nil
}

func (s *roleService) GetRoleMenuIDs(roleID uuid.UUID) ([]uuid.UUID, error) {
	boundary, err := s.GetRoleMenuBoundary(roleID)
	if err != nil {
		return nil, err
	}
	return boundary.EffectiveMenuIDs, nil
}

func (s *roleService) SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.TenantID != nil {
		return ErrTenantRoleManagedByTeam
	}
	boundary, err := s.GetRoleMenuBoundary(roleID)
	if err != nil {
		return err
	}
	selectedMenuIDs, err := ensureSubsetUUIDs(menuIDs, boundary.AvailableMenuIDs, "role menus exceed bound feature packages")
	if err != nil {
		return err
	}
	hiddenMenuIDs := subtractUUIDs(boundary.AvailableMenuIDs, selectedMenuIDs)
	if err := s.roleHiddenMenuRepo.ReplaceRoleHiddenMenus(roleID, hiddenMenuIDs); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPlatformRole(roleID)
	}
	return nil
}

func (s *roleService) GetRoleActions(roleID uuid.UUID) ([]user.RoleActionPermission, error) {
	boundary, err := s.GetRoleActionBoundary(roleID)
	if err != nil {
		return nil, err
	}
	return buildRoleActionPermissions(roleID, boundary.EffectiveActionIDs), nil
}

func (s *roleService) SetRoleActions(roleID uuid.UUID, actions []user.RoleActionPermission) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.TenantID != nil {
		return ErrTenantRoleManagedByTeam
	}
	if role.Code == "team_admin" || role.Code == "team_member" {
		return ErrTeamRoleActionReadonly
	}
	actionIDs := make([]uuid.UUID, 0, len(actions))
	for _, item := range actions {
		actionIDs = append(actionIDs, item.ActionID)
	}
	permissionActions, err := s.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		return err
	}
	permissionActionByID := make(map[uuid.UUID]struct{}, len(permissionActions))
	for _, action := range permissionActions {
		permissionActionByID[action.ID] = struct{}{}
	}
	for _, item := range actions {
		if _, ok := permissionActionByID[item.ActionID]; !ok {
			return errors.New("功能权限不存在")
		}
	}
	boundary, err := s.GetRoleActionBoundary(roleID)
	if err != nil {
		return err
	}
	selectedActionIDs, err := ensureSubsetUUIDs(actionIDs, boundary.AvailableActionIDs, "role actions exceed bound feature packages")
	if err != nil {
		return err
	}
	disabledActionIDs := subtractUUIDs(boundary.AvailableActionIDs, selectedActionIDs)
	if err := s.roleDisabledActionRepo.ReplaceRoleDisabledActions(roleID, disabledActionIDs); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPlatformRole(roleID)
	}
	return nil
}

func (s *roleService) GetRoleDataPermissions(roleID uuid.UUID) ([]user.RoleDataPermission, []string, []DataPermissionScopeOption, error) {
	if _, err := s.roleRepo.GetByID(roleID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil, ErrRoleNotFound
		}
		return nil, nil, nil, err
	}
	records, err := s.roleDataRepo.GetByRoleID(roleID)
	if err != nil {
		return nil, nil, nil, err
	}
	resourceCodes, err := s.actionRepo.ListDistinctResourceCodes()
	if err != nil {
		return nil, nil, nil, err
	}
	return records, resourceCodes, buildAvailableDataScopeOptions(), nil
}

func (s *roleService) SetRoleDataPermissions(roleID uuid.UUID, permissions []user.RoleDataPermission) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.TenantID != nil {
		return ErrTenantRoleManagedByTeam
	}
	resourceCodes, err := s.actionRepo.ListDistinctResourceCodes()
	if err != nil {
		return err
	}
	allowedResources := make(map[string]struct{}, len(resourceCodes))
	for _, resourceCode := range resourceCodes {
		allowedResources[resourceCode] = struct{}{}
	}
	allowedScopes := make(map[string]struct{})
	for _, option := range buildAvailableDataScopeOptions() {
		allowedScopes[option.Code] = struct{}{}
	}

	normalized := make([]user.RoleDataPermission, 0, len(permissions))
	seenResources := make(map[string]struct{}, len(permissions))
	for _, item := range permissions {
		resourceCode := strings.TrimSpace(item.ResourceCode)
		dataScope := strings.TrimSpace(item.DataScope)
		if resourceCode == "" || dataScope == "" {
			return errors.New("资源编码和数据范围不能为空")
		}
		if _, ok := allowedResources[resourceCode]; !ok {
			return errors.New("存在未注册的数据权限资源")
		}
		if _, ok := allowedScopes[dataScope]; !ok {
			return errors.New("存在无效的数据权限范围")
		}
		if _, ok := seenResources[resourceCode]; ok {
			return errors.New("同一资源只能配置一条数据权限")
		}
		seenResources[resourceCode] = struct{}{}
		normalized = append(normalized, user.RoleDataPermission{
			RoleID:       roleID,
			ResourceCode: resourceCode,
			DataScope:    dataScope,
		})
	}
	return s.roleDataRepo.ReplaceRoleDataPermissions(roleID, normalized)
}

func buildAvailableDataScopeOptions() []DataPermissionScopeOption {
	return []DataPermissionScopeOption{
		{Code: "self", Name: "仅自己"},
		{Code: "team", Name: "当前团队"},
		{Code: "all", Name: "全部数据"},
	}
}

func supportsPlatformPackage(contextType string) bool {
	return contextType == "" || contextType == "platform" || contextType == "platform,team"
}

func (s *roleService) GetRoleMenuBoundary(roleID uuid.UUID) (*RoleMenuBoundary, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	if role.TenantID != nil {
		return nil, ErrTenantRoleManagedByTeam
	}
	if s.roleSnapshotService != nil {
		snapshot, snapshotErr := s.roleSnapshotService.GetSnapshot(roleID)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		return &RoleMenuBoundary{
			PackageIDs:         dedupeUUIDs(snapshot.PackageIDs),
			ExpandedPackageIDs: dedupeUUIDs(snapshot.ExpandedPackageIDs),
			AvailableMenuIDs:   dedupeUUIDs(snapshot.AvailableMenuIDs),
			HiddenMenuIDs:      dedupeUUIDs(snapshot.HiddenMenuIDs),
			EffectiveMenuIDs:   dedupeUUIDs(snapshot.EffectiveMenuIDs),
			MenuSourceMap:      filterUUIDSourceMap(snapshot.MenuSourceMap, snapshot.AvailableMenuIDs),
		}, nil
	}
	return &RoleMenuBoundary{PackageIDs: []uuid.UUID{}, ExpandedPackageIDs: []uuid.UUID{}, AvailableMenuIDs: []uuid.UUID{}, HiddenMenuIDs: []uuid.UUID{}, EffectiveMenuIDs: []uuid.UUID{}, MenuSourceMap: map[uuid.UUID][]uuid.UUID{}}, nil
}

func (s *roleService) GetRoleActionBoundary(roleID uuid.UUID) (*RoleActionBoundary, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	if role.TenantID != nil {
		return nil, ErrTenantRoleManagedByTeam
	}
	if s.roleSnapshotService != nil {
		snapshot, snapshotErr := s.roleSnapshotService.GetSnapshot(roleID)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		return &RoleActionBoundary{
			PackageIDs:         dedupeUUIDs(snapshot.PackageIDs),
			ExpandedPackageIDs: dedupeUUIDs(snapshot.ExpandedPackageIDs),
			AvailableActionIDs: dedupeUUIDs(snapshot.AvailableActionIDs),
			DisabledActionIDs:  dedupeUUIDs(snapshot.DisabledActionIDs),
			EffectiveActionIDs: dedupeUUIDs(snapshot.EffectiveActionIDs),
			ActionSourceMap:    filterUUIDSourceMap(snapshot.ActionSourceMap, snapshot.AvailableActionIDs),
		}, nil
	}
	return &RoleActionBoundary{PackageIDs: []uuid.UUID{}, ExpandedPackageIDs: []uuid.UUID{}, AvailableActionIDs: []uuid.UUID{}, DisabledActionIDs: []uuid.UUID{}, EffectiveActionIDs: []uuid.UUID{}, ActionSourceMap: map[uuid.UUID][]uuid.UUID{}}, nil
}

func ensureSubsetUUIDs(selected []uuid.UUID, allowed []uuid.UUID, errMsg string) ([]uuid.UUID, error) {
	allowedSet := make(map[uuid.UUID]struct{}, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(selected))
	seen := make(map[uuid.UUID]struct{}, len(selected))
	for _, id := range selected {
		if _, ok := allowedSet[id]; !ok {
			return nil, errors.New(errMsg)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func buildRoleActionPermissions(roleID uuid.UUID, actionIDs []uuid.UUID) []user.RoleActionPermission {
	result := make([]user.RoleActionPermission, 0, len(actionIDs))
	for _, actionID := range actionIDs {
		result = append(result, user.RoleActionPermission{
			RoleID:   roleID,
			ActionID: actionID,
		})
	}
	return result
}

func subtractUUIDs(source []uuid.UUID, blocked []uuid.UUID) []uuid.UUID {
	if len(source) == 0 {
		return []uuid.UUID{}
	}
	if len(blocked) == 0 {
		return append([]uuid.UUID{}, source...)
	}
	blockedSet := make(map[uuid.UUID]struct{}, len(blocked))
	for _, id := range blocked {
		blockedSet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(source))
	seen := make(map[uuid.UUID]struct{}, len(source))
	for _, id := range source {
		if _, ok := blockedSet[id]; ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func dedupeUUIDs(items []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, id := range items {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func appendUUIDIfMissing(items []uuid.UUID, value uuid.UUID) []uuid.UUID {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func filterUUIDSourceMap(sourceMap map[uuid.UUID][]uuid.UUID, allowed []uuid.UUID) map[uuid.UUID][]uuid.UUID {
	if len(sourceMap) == 0 {
		return map[uuid.UUID][]uuid.UUID{}
	}
	allowedSet := make(map[uuid.UUID]struct{}, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = struct{}{}
	}
	result := make(map[uuid.UUID][]uuid.UUID)
	for id, packageIDs := range sourceMap {
		if _, ok := allowedSet[id]; !ok {
			continue
		}
		result[id] = dedupeUUIDs(packageIDs)
	}
	return result
}
