package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/repository"
)

var ErrRoleNotFound = errors.New("role not found")
var ErrRoleCodeExists = errors.New("role code already exists")
var ErrSystemRoleCannotDelete = errors.New("system role cannot be deleted")

// RoleService 全局角色管理服务（仅 roles 表）
type RoleService interface {
	List(req *dto.RoleListRequest) ([]model.Role, int64, error)
	Get(id uuid.UUID) (*model.Role, error)
	Create(req *dto.RoleCreateRequest) (*model.Role, error)
	Update(id uuid.UUID, req *dto.RoleUpdateRequest) error
	Delete(id uuid.UUID) error
	GetRoleMenuIDs(roleID uuid.UUID) ([]uuid.UUID, error)
	SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error
}

type roleService struct {
	roleRepo     repository.RoleRepository
	roleMenuRepo repository.RoleMenuRepository
	userRoleRepo repository.UserRoleRepository
	scopeRepo    repository.ScopeRepository
	logger       *zap.Logger
}

func NewRoleService(roleRepo repository.RoleRepository, roleMenuRepo repository.RoleMenuRepository, userRoleRepo repository.UserRoleRepository, scopeRepo repository.ScopeRepository, logger *zap.Logger) RoleService {
	return &roleService{roleRepo: roleRepo, roleMenuRepo: roleMenuRepo, userRoleRepo: userRoleRepo, scopeRepo: scopeRepo, logger: logger}
}

func (s *roleService) List(req *dto.RoleListRequest) ([]model.Role, int64, error) {
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
	return s.roleRepo.List(
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

func (s *roleService) Get(id uuid.UUID) (*model.Role, error) {
	return s.roleRepo.GetByID(id)
}

func (s *roleService) Create(req *dto.RoleCreateRequest) (*model.Role, error) {
	_, err := s.roleRepo.GetByCode(req.Code)
	if err == nil {
		return nil, ErrRoleCodeExists
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// 验证作用域ID
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
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	role := &model.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		ScopeID:     scopeID,
		Enabled:     enabled,
		SortOrder:   req.SortOrder,
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
	// 构建更新字段映射
	updates := make(map[string]interface{})
	// 如果提供了 code 且与当前 code 不同，需要检查唯一性
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
	// 如果提供了 enabled，更新它
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.SortOrder > 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.ScopeID != "" {
		scopeID, err := uuid.Parse(req.ScopeID)
		if err != nil {
			return errors.New("invalid scope_id")
		}
		// 验证作用域是否存在
		_, err = s.scopeRepo.GetByID(scopeID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("scope not found")
			}
			return err
		}
		// 总是更新 scope_id（即使值相同，也确保数据库中的值是正确的）
		updates["scope_id"] = scopeID
		s.logger.Info("准备更新scope_id", zap.String("roleId", id.String()), zap.String("oldScopeId", role.ScopeID.String()), zap.String("newScopeId", scopeID.String()))
	}
	// 如果有更新字段，执行更新
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
	// 检查是否为默认角色（通过code判断）
	if role.Code == "admin" || role.Code == "team_admin" || role.Code == "team_member" {
		return ErrSystemRoleCannotDelete
	}
	
	// 使用事务删除角色及其所有关联关系（都是软删除）
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 软删除用户角色关联
		if err := tx.Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		
		// 2. 软删除角色菜单关联
		if err := tx.Where("role_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		
		// 3. 最后软删除角色本身
		return tx.Delete(&model.Role{}, id).Error
	})
}

func (s *roleService) GetRoleMenuIDs(roleID uuid.UUID) ([]uuid.UUID, error) {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
