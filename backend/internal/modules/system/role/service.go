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
)

var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrRoleCodeExists         = errors.New("role code already exists")
	ErrSystemRoleCannotDelete = errors.New("system role cannot be deleted")
)

type RoleService interface {
	List(req *dto.RoleListRequest) ([]user.Role, int64, error)
	Get(id uuid.UUID) (*user.Role, error)
	Create(req *dto.RoleCreateRequest) (*user.Role, error)
	Update(id uuid.UUID, req *dto.RoleUpdateRequest) error
	Delete(id uuid.UUID) error
	GetRoleMenuIDs(roleID uuid.UUID) ([]uuid.UUID, error)
	SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error
	GetRoleActions(roleID uuid.UUID) ([]user.RoleActionPermission, error)
	SetRoleActions(roleID uuid.UUID, actions []user.RoleActionPermission) error
	GetRoleDataPermissions(roleID uuid.UUID) ([]user.RoleDataPermission, []string, []DataPermissionScopeOption, error)
	SetRoleDataPermissions(roleID uuid.UUID, permissions []user.RoleDataPermission) error
}

type DataPermissionScopeOption struct {
	Code string
	Name string
}

type roleService struct {
	roleRepo       user.RoleRepository
	roleMenuRepo   user.RoleMenuRepository
	roleActionRepo user.RoleActionPermissionRepository
	roleDataRepo   user.RoleDataPermissionRepository
	actionRepo     user.PermissionActionRepository
	userRoleRepo   user.UserRoleRepository
	scopeRepo      user.ScopeRepository
	logger         *zap.Logger
}

func NewRoleService(roleRepo user.RoleRepository, roleMenuRepo user.RoleMenuRepository, roleActionRepo user.RoleActionPermissionRepository, roleDataRepo user.RoleDataPermissionRepository, actionRepo user.PermissionActionRepository, userRoleRepo user.UserRoleRepository, scopeRepo user.ScopeRepository, logger *zap.Logger) RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		roleMenuRepo:   roleMenuRepo,
		roleActionRepo: roleActionRepo,
		roleDataRepo:   roleDataRepo,
		actionRepo:     actionRepo,
		userRoleRepo:   userRoleRepo,
		scopeRepo:      scopeRepo,
		logger:         logger,
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
	scopeCodes := mergeScopeFilters(req.Scope, req.Scopes, req.GlobalOnly)
	if len(scopeCodes) > 0 {
		scopes, err := s.resolveScopesByFilters(scopeCodes)
		if err != nil {
			return nil, 0, err
		}
		scopeIDs := make([]uuid.UUID, 0, len(scopes))
		for _, scope := range scopes {
			scopeIDs = append(scopeIDs, scope.ID)
		}
		return s.roleRepo.ListByScopeIDs(
			scopeIDs,
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
	_, err := s.roleRepo.GetByCode(req.Code)
	if err == nil {
		return nil, ErrRoleCodeExists
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	scopeIDs, err := s.resolveScopeIDs(req.ScopeIDs, true)
	if err != nil {
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
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		records := s.buildRoleScopeRecords(role.ID, scopeIDs)
		if len(records) == 0 {
			return nil
		}
		return tx.Create(records).Error
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
	scopeIDs, err := s.resolveScopeIDs(req.ScopeIDs, false)
	if err != nil {
		return err
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			updates["updated_at"] = time.Now()
			s.logger.Info("更新角色字段", zap.String("roleId", id.String()), zap.Any("updates", updates))
			if err := tx.Model(&user.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.ScopeIDs != nil {
			if err := tx.Where("role_id = ?", id).Delete(&user.RoleScope{}).Error; err != nil {
				return err
			}
			records := s.buildRoleScopeRecords(id, scopeIDs)
			if len(records) == 0 {
				return nil
			}
			if err := tx.Create(records).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.logger.Error("更新角色失败", zap.String("roleId", id.String()), zap.Error(err))
		return err
	}
	if len(updates) == 0 && req.ScopeIDs == nil {
		s.logger.Info("无需更新角色", zap.String("roleId", id.String()))
		return nil
	}
	s.logger.Info("角色更新成功", zap.String("roleId", id.String()))
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
	if role.Code == "admin" || role.Code == "team_admin" || role.Code == "team_member" {
		return ErrSystemRoleCannotDelete
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&user.RoleScope{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}

		if err := tx.Where("role_id = ?", id).Delete(&user.RoleMenu{}).Error; err != nil {
			return err
		}

		if err := tx.Where("role_id = ?", id).Delete(&user.RoleActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("role_id = ?", id).Delete(&user.RoleDataPermission{}).Error; err != nil {
			return err
		}

		return tx.Delete(&user.Role{}, id).Error
	})
}

func (s *roleService) GetRoleMenuIDs(roleID uuid.UUID) ([]uuid.UUID, error) {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return s.roleMenuRepo.GetMenuIDsByRoleID(roleID)
}

func (s *roleService) SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	return s.roleMenuRepo.SetRoleMenus(roleID, menuIDs)
}

func (s *roleService) GetRoleActions(roleID uuid.UUID) ([]user.RoleActionPermission, error) {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return s.roleActionRepo.GetByRoleID(roleID)
}

func (s *roleService) SetRoleActions(roleID uuid.UUID, actions []user.RoleActionPermission) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}
	actionIDs := make([]uuid.UUID, 0, len(actions))
	for _, item := range actions {
		actionIDs = append(actionIDs, item.ActionID)
	}
	permissionActions, err := s.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		return err
	}
	actionScopeByID := make(map[uuid.UUID]string, len(permissionActions))
	for _, action := range permissionActions {
		actionScopeByID[action.ID] = action.Scope.Code
	}
	allowedScopes := roleScopeCodeSet(role)
	for _, item := range actions {
		scopeCode, ok := actionScopeByID[item.ActionID]
		if !ok {
			return errors.New("功能权限不存在")
		}
		if _, ok := allowedScopes[scopeCode]; !ok {
			return errors.New("功能权限作用域与角色作用域不匹配")
		}
	}
	return s.roleActionRepo.SetRoleActions(roleID, actions)
}

func (s *roleService) GetRoleDataPermissions(roleID uuid.UUID) ([]user.RoleDataPermission, []string, []DataPermissionScopeOption, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil, ErrRoleNotFound
		}
		return nil, nil, nil, err
	}
	records, err := s.roleDataRepo.GetByRoleID(roleID)
	if err != nil {
		return nil, nil, nil, err
	}
	scopeCodes := roleScopeCodes(role)
	resourceCodes, err := s.listDistinctResourceCodesByScopes(scopeCodes)
	if err != nil {
		return nil, nil, nil, err
	}
	scopeOptions := buildAvailableDataScopeOptions(role.Scopes)
	return records, resourceCodes, scopeOptions, nil
}

func (s *roleService) SetRoleDataPermissions(roleID uuid.UUID, permissions []user.RoleDataPermission) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoleNotFound
		}
		return err
	}

	scopeCodes := roleScopeCodes(role)
	resourceCodes, err := s.listDistinctResourceCodesByScopes(scopeCodes)
	if err != nil {
		return err
	}
	allowedResources := make(map[string]struct{}, len(resourceCodes))
	for _, resourceCode := range resourceCodes {
		allowedResources[resourceCode] = struct{}{}
	}
	allowedScopes := make(map[string]struct{})
	for _, option := range buildAvailableDataScopeOptions(role.Scopes) {
		allowedScopes[option.Code] = struct{}{}
	}

	normalized := make([]user.RoleDataPermission, 0, len(permissions))
	seenResources := make(map[string]struct{}, len(permissions))
	for _, item := range permissions {
		resourceCode := strings.TrimSpace(item.ResourceCode)
		scopeCode := strings.TrimSpace(item.ScopeCode)
		if resourceCode == "" || scopeCode == "" {
			return errors.New("资源编码和数据范围不能为空")
		}
		if _, ok := allowedResources[resourceCode]; !ok {
			return errors.New("存在未注册的数据权限资源")
		}
		if _, ok := allowedScopes[scopeCode]; !ok {
			return errors.New("存在无效的数据权限范围")
		}
		if _, ok := seenResources[resourceCode]; ok {
			return errors.New("同一资源只能配置一条数据权限")
		}
		seenResources[resourceCode] = struct{}{}
		normalized = append(normalized, user.RoleDataPermission{
			RoleID:       roleID,
			ResourceCode: resourceCode,
			ScopeCode:    scopeCode,
		})
	}

	return s.roleDataRepo.ReplaceRoleDataPermissions(roleID, normalized)
}

func buildAvailableDataScopeOptions(scopes []user.Scope) []DataPermissionScopeOption {
	seen := make(map[string]struct{})
	options := []DataPermissionScopeOption{
		{Code: "self", Name: "仅本人"},
	}
	for _, scope := range scopes {
		code := strings.TrimSpace(scope.DataPermissionCode)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		name := strings.TrimSpace(scope.DataPermissionName)
		if name == "" {
			name = strings.TrimSpace(scope.Name)
		}
		if name == "" {
			name = code
		}
		options = append(options, DataPermissionScopeOption{Code: code, Name: name})
	}
	options = append(options, DataPermissionScopeOption{Code: "all", Name: "全部数据"})
	return options
}

func mergeScopeFilters(scope string, scopes []string, globalOnly bool) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(scopes)+1)
	appendScope := func(code string) {
		code = strings.TrimSpace(code)
		if code == "" {
			return
		}
		if _, ok := seen[code]; ok {
			return
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	appendScope(scope)
	for _, item := range scopes {
		appendScope(item)
	}
	if globalOnly && len(out) == 0 {
		appendScope("tenant")
	}
	return out
}

func (s *roleService) resolveScopesByFilters(filters []string) ([]user.Scope, error) {
	allScopes, err := s.scopeRepo.GetAll()
	if err != nil {
		return nil, err
	}
	if len(filters) == 0 {
		return allScopes, nil
	}
	seen := make(map[uuid.UUID]struct{})
	result := make([]user.Scope, 0, len(filters))
	for _, filter := range filters {
		trimmed := strings.TrimSpace(filter)
		if trimmed == "" {
			continue
		}
		for _, scope := range allScopes {
			if scope.Code != trimmed && scope.ContextKind != trimmed {
				continue
			}
			if _, ok := seen[scope.ID]; ok {
				continue
			}
			seen[scope.ID] = struct{}{}
			result = append(result, scope)
		}
	}
	return result, nil
}

func (s *roleService) resolveScopeIDs(scopeIDs []string, required bool) ([]uuid.UUID, error) {
	normalized := make([]uuid.UUID, 0, len(scopeIDs)+1)
	seen := make(map[uuid.UUID]struct{})
	appendScope := func(raw string) error {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil
		}
		scopeID, err := uuid.Parse(raw)
		if err != nil {
			return errors.New("invalid scope_id")
		}
		if _, ok := seen[scopeID]; ok {
			return nil
		}
		seen[scopeID] = struct{}{}
		normalized = append(normalized, scopeID)
		return nil
	}
	for _, raw := range scopeIDs {
		if err := appendScope(raw); err != nil {
			return nil, err
		}
	}
	if len(normalized) == 0 {
		if required {
			return nil, errors.New("scope not found")
		}
		return nil, nil
	}
	scopes, err := s.scopeRepo.GetByIDs(normalized)
	if err != nil {
		return nil, err
	}
	if len(scopes) != len(normalized) {
		return nil, errors.New("scope not found")
	}
	return normalized, nil
}

func (s *roleService) buildRoleScopeRecords(roleID uuid.UUID, scopeIDs []uuid.UUID) []user.RoleScope {
	records := make([]user.RoleScope, 0, len(scopeIDs))
	for _, scopeID := range scopeIDs {
		records = append(records, user.RoleScope{RoleID: roleID, ScopeID: scopeID})
	}
	return records
}

func roleScopeCodes(role *user.Role) []string {
	codes := make([]string, 0, len(role.Scopes))
	seen := make(map[string]struct{})
	appendCode := func(code string) {
		code = strings.TrimSpace(code)
		if code == "" {
			return
		}
		if _, ok := seen[code]; ok {
			return
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	for _, scope := range role.Scopes {
		appendCode(scope.Code)
	}
	return codes
}

func roleScopeCodeSet(role *user.Role) map[string]struct{} {
	codes := roleScopeCodes(role)
	result := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		result[code] = struct{}{}
	}
	return result
}

func firstScopeCode(codes []string) string {
	if len(codes) == 0 {
		return ""
	}
	return codes[0]
}

func (s *roleService) listDistinctResourceCodesByScopes(scopeCodes []string) ([]string, error) {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, scopeCode := range scopeCodes {
		codes, err := s.actionRepo.ListDistinctResourceCodesByScope(scopeCode)
		if err != nil {
			return nil, err
		}
		for _, code := range codes {
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			result = append(result, code)
		}
	}
	return result, nil
}
