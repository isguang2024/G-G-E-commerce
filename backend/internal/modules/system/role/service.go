package role

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/appscope"
	"github.com/maben/backend/internal/pkg/database"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
	"github.com/maben/backend/internal/pkg/platformroleaccess"
)

var (
	ErrRoleNotFound                          = errors.New("角色不存在")
	ErrRoleCodeExists                        = errors.New("角色编码已存在")
	ErrRoleAppScopeMismatch                  = errors.New("角色在当前应用中不生效")
	ErrSystemRoleCannotDelete                = errors.New("系统角色不可删除")
	ErrCollaborationWorkspaceRoleKeyReadonly = errors.New("协作空间角色的权限键由空间边界统一管理，不可在此修改")
	ErrCollaborationWorkspaceRoleManaged     = errors.New("该角色在协作空间上下文中管理")
)

type RoleService interface {
	List(req *dto.RoleListRequest) ([]user.Role, int64, error)
	ListOptions() ([]user.Role, error)
	Get(id uuid.UUID) (*user.Role, error)
	Create(req *dto.RoleCreateRequest) (*user.Role, error)
	Update(id uuid.UUID, req *dto.RoleUpdateRequest) error
	Delete(id uuid.UUID) error
	GetRolePackages(roleID uuid.UUID, appKey string) ([]uuid.UUID, []user.FeaturePackage, error)
	SetRolePackages(roleID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) error
	GetRoleMenuBoundary(roleID uuid.UUID, appKey string) (*RoleMenuBoundary, error)
	GetRoleMenuIDs(roleID uuid.UUID, appKey string) ([]uuid.UUID, error)
	SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID, appKey string) error
	GetRoleKeyBoundary(roleID uuid.UUID, appKey string) (*RoleKeyBoundary, error)
	GetRoleKeys(roleID uuid.UUID, appKey string) ([]user.RoleKeyPermission, error)
	SetRoleKeys(roleID uuid.UUID, keys []user.RoleKeyPermission, appKey string) error
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

type RoleKeyBoundary struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	AvailableKeyIDs    []uuid.UUID
	DisabledKeyIDs     []uuid.UUID
	EffectiveKeyIDs    []uuid.UUID
	KeySourceMap       map[uuid.UUID][]uuid.UUID
}

type DataPermissionScopeOption struct {
	Code string
	Name string
}

type roleService struct {
	db                     *gorm.DB
	roleRepo               user.RoleRepository
	rolePackageRepo        user.RoleFeaturePackageRepository
	featurePkgRepo         user.FeaturePackageRepository
	packageKeyRepo         user.FeaturePackageKeyRepository
	packageMenuRepo        user.FeaturePackageMenuRepository
	packageBundleRepo      user.FeaturePackageBundleRepository
	roleHiddenMenuRepo     user.RoleHiddenMenuRepository
	roleDisabledActionRepo user.RoleDisabledActionRepository
	roleDataRepo           user.RoleDataPermissionRepository
	keyRepo                user.PermissionKeyRepository
	roleSnapshotService    platformroleaccess.Service
	refresher              permissionrefresh.Service
	logger                 *zap.Logger
}

func NewRoleService(
	db *gorm.DB,
	roleRepo user.RoleRepository,
	rolePackageRepo user.RoleFeaturePackageRepository,
	featurePkgRepo user.FeaturePackageRepository,
	packageKeyRepo user.FeaturePackageKeyRepository,
	packageMenuRepo user.FeaturePackageMenuRepository,
	packageBundleRepo user.FeaturePackageBundleRepository,
	roleHiddenMenuRepo user.RoleHiddenMenuRepository,
	roleDisabledActionRepo user.RoleDisabledActionRepository,
	roleDataRepo user.RoleDataPermissionRepository,
	keyRepo user.PermissionKeyRepository,
	roleSnapshotService platformroleaccess.Service,
	refresher permissionrefresh.Service,
	logger *zap.Logger,
) RoleService {
	return &roleService{
		db:                     db,
		roleRepo:               roleRepo,
		rolePackageRepo:        rolePackageRepo,
		featurePkgRepo:         featurePkgRepo,
		packageKeyRepo:         packageKeyRepo,
		packageMenuRepo:        packageMenuRepo,
		packageBundleRepo:      packageBundleRepo,
		roleHiddenMenuRepo:     roleHiddenMenuRepo,
		roleDisabledActionRepo: roleDisabledActionRepo,
		roleDataRepo:           roleDataRepo,
		keyRepo:                keyRepo,
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
		req.AppKey,
		req.StartTime,
		req.EndTime,
		req.Enabled,
		req.GlobalOnly,
	)
}

func (s *roleService) ListOptions() ([]user.Role, error) {
	return s.roleRepo.List()
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
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		SortOrder:    req.SortOrder,
		CustomParams: user.MetaJSON(req.CustomParams),
		Status:       status,
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		return replaceRoleAppScopesTx(tx, role.ID, req.AppKeys)
	}); err != nil {
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
	if role.CollaborationWorkspaceID != nil {
		return ErrCollaborationWorkspaceRoleManaged
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
	updates["custom_params"] = user.MetaJSON(req.CustomParams)
	if req.Status != "" {
		updates["status"] = req.Status
	}
	appScopesChanged := req.AppKeys != nil && !sameStringSet(role.AppKeys, req.AppKeys)
	if len(updates) == 0 && !appScopesChanged {
		s.logger.Info("无需更新角色", zap.String("roleId", id.String()))
		return nil
	}
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
	}
	s.logger.Info("更新角色字段", zap.String("roleId", id.String()), zap.Any("updates", updates), zap.Any("appKeys", req.AppKeys))
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&user.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}
		if appScopesChanged {
			if err := replaceRoleAppScopesTx(tx, id, req.AppKeys); err != nil {
				return err
			}
			if err := cleanupRoleAppArtifactsTx(tx, id, req.AppKeys); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPersonalWorkspaceRole(id)
	}
	return nil
}

func cleanupRoleAppArtifactsTx(tx *gorm.DB, roleID uuid.UUID, appKeys []string) error {
	if len(appKeys) == 0 {
		return nil
	}
	normalized := normalizeAppKeys(appKeys)
	if len(normalized) == 0 {
		return nil
	}
	if err := tx.Where("role_id = ? AND app_key NOT IN ?", roleID, normalized).Delete(&user.RoleFeaturePackage{}).Error; err != nil {
		return err
	}
	if err := tx.Where("role_id = ? AND app_key NOT IN ?", roleID, normalized).Delete(&user.RoleHiddenMenu{}).Error; err != nil {
		return err
	}
	return tx.Where("role_id = ? AND app_key NOT IN ?", roleID, normalized).Delete(&user.RoleDisabledAction{}).Error
}

func replaceRoleAppScopesTx(tx *gorm.DB, roleID uuid.UUID, appKeys []string) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&user.RoleAppScope{}).Error; err != nil {
		return err
	}
	normalized := normalizeAppKeys(appKeys)
	if len(normalized) == 0 {
		return nil
	}
	items := make([]user.RoleAppScope, 0, len(normalized))
	for _, item := range normalized {
		items = append(items, user.RoleAppScope{RoleID: roleID, AppKey: item})
	}
	return tx.Create(&items).Error
}

func (s *roleService) ensureRoleEffectiveInApp(role *user.Role, appKey string) error {
	if role == nil {
		return ErrRoleNotFound
	}
	if len(role.AppKeys) == 0 {
		return nil
	}
	normalizedAppKey := appscope.Normalize(appKey)
	for _, item := range role.AppKeys {
		if appscope.Normalize(item) == normalizedAppKey {
			return nil
		}
	}
	return ErrRoleAppScopeMismatch
}

func normalizeAppKeys(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		target := appscope.Normalize(item)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		result = append(result, target)
	}
	return result
}

func sameStringSet(left []string, right []string) bool {
	normalizedLeft := normalizeAppKeys(left)
	normalizedRight := normalizeAppKeys(right)
	if len(normalizedLeft) != len(normalizedRight) {
		return false
	}
	seen := make(map[string]struct{}, len(normalizedLeft))
	for _, item := range normalizedLeft {
		seen[item] = struct{}{}
	}
	for _, item := range normalizedRight {
		if _, ok := seen[item]; !ok {
			return false
		}
	}
	return true
}

func (s *roleService) Delete(id uuid.UUID) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.CollaborationWorkspaceID != nil {
		return ErrCollaborationWorkspaceRoleManaged
	}
	if role.Code == "admin" || role.Code == "collaboration_admin" || role.Code == "collaboration_member" {
		return ErrSystemRoleCannotDelete
	}
	var affectedUserIDs []uuid.UUID
	if err := database.DB.Model(&user.UserRole{}).
		Where("role_id = ? AND collaboration_workspace_id IS NULL", id).
		Distinct("user_id").
		Pluck("user_id", &affectedUserIDs).Error; err != nil {
		return err
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
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
		if err := tx.Where("role_id = ?", id).Delete(&user.RoleAppScope{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&models.PersonalWorkspaceRoleAccessSnapshot{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user.Role{}, id).Error
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPersonalWorkspaceUsers(affectedUserIDs)
	}
	return nil
}

func (s *roleService) GetRolePackages(roleID uuid.UUID, appKey string) ([]uuid.UUID, []user.FeaturePackage, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, ErrRoleNotFound
		}
		return nil, nil, err
	}
	if role.CollaborationWorkspaceID != nil {
		return nil, nil, ErrCollaborationWorkspaceRoleManaged
	}
	if err := s.ensureRoleEffectiveInApp(role, appKey); err != nil {
		return nil, nil, err
	}
	packageIDs, err := appscope.PackageIDsByRole(s.db, roleID, appKey)
	if err != nil {
		return nil, nil, err
	}
	packages, err := s.featurePkgRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, nil, err
	}
	return packageIDs, packages, nil
}

func (s *roleService) SetRolePackages(roleID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.CollaborationWorkspaceID != nil {
		return ErrCollaborationWorkspaceRoleManaged
	}
	if err := s.ensureRoleEffectiveInApp(role, appKey); err != nil {
		return err
	}
	normalizedAppKey := appscope.Normalize(appKey)
	if len(packageIDs) > 0 {
		packages, err := s.featurePkgRepo.GetByIDs(packageIDs)
		if err != nil {
			return err
		}
		if len(packages) != len(packageIDs) {
			return errors.New("无效的功能包")
		}
		for _, pkg := range packages {
			if appscope.Normalize(pkg.AppKey) != normalizedAppKey {
				return errors.New("包含不属于当前应用的功能包")
			}
			if !supportsPersonalWorkspacePackageContext(pkg.ContextType) {
				return errors.New("包含不支持个人空间权限的功能包")
			}
		}
	}
	if err := appscope.ReplaceRolePackagesInApp(s.db, roleID, normalizedAppKey, packageIDs, grantedBy); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPersonalWorkspaceRole(roleID)
	}
	return nil
}

func supportsPersonalWorkspacePackageContext(contextType string) bool {
	switch strings.TrimSpace(contextType) {
	case "personal", "common", "personal,collaboration", "collaboration,personal":
		return true
	default:
		return false
	}
}

func (s *roleService) GetRoleMenuIDs(roleID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	boundary, err := s.GetRoleMenuBoundary(roleID, appKey)
	if err != nil {
		return nil, err
	}
	return boundary.EffectiveMenuIDs, nil
}

func (s *roleService) SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID, appKey string) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.CollaborationWorkspaceID != nil {
		return ErrCollaborationWorkspaceRoleManaged
	}
	if err := s.ensureRoleEffectiveInApp(role, appKey); err != nil {
		return err
	}
	boundary, err := s.GetRoleMenuBoundary(roleID, appKey)
	if err != nil {
		return err
	}
	selectedMenuIDs, err := ensureSubsetUUIDs(menuIDs, boundary.AvailableMenuIDs, "role menus exceed bound feature packages")
	if err != nil {
		return err
	}
	hiddenMenuIDs := subtractUUIDs(boundary.AvailableMenuIDs, selectedMenuIDs)
	if err := appscope.ReplaceRoleHiddenMenusInApp(s.db, roleID, appKey, hiddenMenuIDs); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPersonalWorkspaceRole(roleID)
	}
	return nil
}

func (s *roleService) GetRoleKeys(roleID uuid.UUID, appKey string) ([]user.RoleKeyPermission, error) {
	boundary, err := s.GetRoleKeyBoundary(roleID, appKey)
	if err != nil {
		return nil, err
	}
	return buildRoleKeyPermissions(roleID, boundary.EffectiveKeyIDs), nil
}

func (s *roleService) SetRoleKeys(roleID uuid.UUID, keys []user.RoleKeyPermission, appKey string) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	if role.CollaborationWorkspaceID != nil {
		return ErrCollaborationWorkspaceRoleManaged
	}
	if err := s.ensureRoleEffectiveInApp(role, appKey); err != nil {
		return err
	}
	if role.Code == "collaboration_admin" || role.Code == "collaboration_member" {
		return ErrCollaborationWorkspaceRoleKeyReadonly
	}
	keyIDs := make([]uuid.UUID, 0, len(keys))
	for _, item := range keys {
		keyIDs = append(keyIDs, item.KeyID)
	}
	permissionKeys, err := s.keyRepo.GetByIDs(keyIDs)
	if err != nil {
		return err
	}
	permissionKeyByID := make(map[uuid.UUID]struct{}, len(permissionKeys))
	for _, item := range permissionKeys {
		permissionKeyByID[item.ID] = struct{}{}
	}
	for _, item := range keys {
		if _, ok := permissionKeyByID[item.KeyID]; !ok {
			return errors.New("功能权限不存在")
		}
	}
	boundary, err := s.GetRoleKeyBoundary(roleID, appKey)
	if err != nil {
		return err
	}
	selectedKeyIDs, err := ensureSubsetUUIDs(keyIDs, boundary.AvailableKeyIDs, "role permission keys exceed bound feature packages")
	if err != nil {
		return err
	}
	disabledKeyIDs := subtractUUIDs(boundary.AvailableKeyIDs, selectedKeyIDs)
	if err := appscope.ReplaceRoleDisabledActionsInScope(s.db, roleID, boundary.AvailableKeyIDs, disabledKeyIDs); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPersonalWorkspaceRole(roleID)
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
	resourceCodes, err := s.keyRepo.ListDistinctModuleCodes()
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
	if role.CollaborationWorkspaceID != nil {
		return ErrCollaborationWorkspaceRoleManaged
	}
	resourceCodes, err := s.keyRepo.ListDistinctModuleCodes()
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
		{Code: "collaboration", Name: "当前协作空间"},
		{Code: "all", Name: "全部数据"},
	}
}

func (s *roleService) GetRoleMenuBoundary(roleID uuid.UUID, appKey string) (*RoleMenuBoundary, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	if role.CollaborationWorkspaceID != nil {
		return nil, ErrCollaborationWorkspaceRoleManaged
	}
	if err := s.ensureRoleEffectiveInApp(role, appKey); err != nil {
		return nil, err
	}
	if s.roleSnapshotService != nil {
		snapshot, snapshotErr := s.roleSnapshotService.GetSnapshot(roleID, appscope.Normalize(appKey))
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

func (s *roleService) GetRoleKeyBoundary(roleID uuid.UUID, appKey string) (*RoleKeyBoundary, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	if role.CollaborationWorkspaceID != nil {
		return nil, ErrCollaborationWorkspaceRoleManaged
	}
	if err := s.ensureRoleEffectiveInApp(role, appKey); err != nil {
		return nil, err
	}
	if s.roleSnapshotService != nil {
		snapshot, snapshotErr := s.roleSnapshotService.GetSnapshot(roleID, appscope.Normalize(appKey))
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		return &RoleKeyBoundary{
			PackageIDs:         dedupeUUIDs(snapshot.PackageIDs),
			ExpandedPackageIDs: dedupeUUIDs(snapshot.ExpandedPackageIDs),
			AvailableKeyIDs:    dedupeUUIDs(snapshot.AvailableActionIDs),
			DisabledKeyIDs:     dedupeUUIDs(snapshot.DisabledActionIDs),
			EffectiveKeyIDs:    dedupeUUIDs(snapshot.EffectiveActionIDs),
			KeySourceMap:       filterUUIDSourceMap(snapshot.ActionSourceMap, snapshot.AvailableActionIDs),
		}, nil
	}
	return &RoleKeyBoundary{PackageIDs: []uuid.UUID{}, ExpandedPackageIDs: []uuid.UUID{}, AvailableKeyIDs: []uuid.UUID{}, DisabledKeyIDs: []uuid.UUID{}, EffectiveKeyIDs: []uuid.UUID{}, KeySourceMap: map[uuid.UUID][]uuid.UUID{}}, nil
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

func buildRoleKeyPermissions(roleID uuid.UUID, keyIDs []uuid.UUID) []user.RoleKeyPermission {
	result := make([]user.RoleKeyPermission, 0, len(keyIDs))
	for _, keyID := range keyIDs {
		result = append(result, user.RoleKeyPermission{
			RoleID: roleID,
			KeyID:  keyID,
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
