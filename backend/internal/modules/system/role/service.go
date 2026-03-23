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
	ErrTeamRoleActionReadonly = errors.New("team role actions are managed by team boundary")
	ErrTenantRoleManagedByTeam = errors.New("tenant role is managed in team context")
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
	logger         *zap.Logger
}

func NewRoleService(roleRepo user.RoleRepository, roleMenuRepo user.RoleMenuRepository, roleActionRepo user.RoleActionPermissionRepository, roleDataRepo user.RoleDataPermissionRepository, actionRepo user.PermissionActionRepository, logger *zap.Logger) RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		roleMenuRepo:   roleMenuRepo,
		roleActionRepo: roleActionRepo,
		roleDataRepo:   roleDataRepo,
		actionRepo:     actionRepo,
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
	return s.roleRepo.UpdateWithMap(id, updates)
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
	return s.roleActionRepo.SetRoleActions(roleID, actions)
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
