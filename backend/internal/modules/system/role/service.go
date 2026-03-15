package role

import (
	"errors"
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
	ErrRoleCodeExists        = errors.New("role code already exists")
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
}

type roleService struct {
	roleRepo     user.RoleRepository
	roleMenuRepo user.RoleMenuRepository
	userRoleRepo user.UserRoleRepository
	scopeRepo    user.ScopeRepository
	logger       *zap.Logger
}

func NewRoleService(roleRepo user.RoleRepository, roleMenuRepo user.RoleMenuRepository, userRoleRepo user.UserRoleRepository, scopeRepo user.ScopeRepository, logger *zap.Logger) RoleService {
	return &roleService{
		roleRepo:     roleRepo,
		roleMenuRepo: roleMenuRepo,
		userRoleRepo: userRoleRepo,
		scopeRepo:    scopeRepo,
		logger:       logger,
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
	scope := req.Scope
	if req.GlobalOnly && scope == "" {
		scope = "team"
	}
	if scope != "" {
		return s.roleRepo.ListByScope(
			scope,
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
	scopeID, err := uuid.Parse(req.ScopeID)
	if err != nil {
		return nil, errors.New("invalid scope_id")
	}
	_, err = s.scopeRepo.GetByID(scopeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("scope not found")
		}
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
		ScopeID:     scopeID,
		SortOrder:   req.SortOrder,
		Priority:    req.Priority,
		Status:      status,
	}
	if err := s.roleRepo.Create(role); err != nil {
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
	if req.Priority > 0 {
		updates["priority"] = req.Priority
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.ScopeID != "" {
		scopeID, err := uuid.Parse(req.ScopeID)
		if err != nil {
			return errors.New("invalid scope_id")
		}
		_, err = s.scopeRepo.GetByID(scopeID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("scope not found")
			}
			return err
		}
		updates["scope_id"] = scopeID
		s.logger.Info("准备更新scope_id", zap.String("roleId", id.String()), zap.String("oldScopeId", role.ScopeID.String()), zap.String("newScopeId", scopeID.String()))
	}
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		s.logger.Info("更新角色字段", zap.String("roleId", id.String()), zap.Any("updates", updates))
		if err := s.roleRepo.UpdateWithMap(id, updates); err != nil {
			s.logger.Error("更新角色失败", zap.String("roleId", id.String()), zap.Error(err))
			return err
		}
		s.logger.Info("角色更新成功", zap.String("roleId", id.String()))
		return nil
	}
	s.logger.Info("无需更新角色", zap.String("roleId", id.String()))
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
		if err := tx.Where("role_id = ?", id).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}

		if err := tx.Where("role_id = ?", id).Delete(&user.RoleMenu{}).Error; err != nil {
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
