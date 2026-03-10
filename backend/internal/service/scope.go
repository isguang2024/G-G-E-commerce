package service

import (
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/repository"
)

var ErrScopeNotFound = errors.New("scope not found")
var ErrScopeCodeExists = errors.New("scope code already exists")
var ErrScopeInUse = errors.New("scope is in use and cannot be deleted")

// ErrScopeInUseWithRoles 作用域正在使用中，包含关联的角色列表
type ErrScopeInUseWithRoles struct {
	Roles []model.Role
}

func (e *ErrScopeInUseWithRoles) Error() string {
	return "scope is in use and cannot be deleted"
}

// ScopeService 作用域管理服务
type ScopeService interface {
	List(req *dto.ScopeListRequest) ([]model.Scope, int64, error)
	Get(id uuid.UUID) (*model.Scope, error)
	Create(req *dto.ScopeCreateRequest) (*model.Scope, error)
	Update(id uuid.UUID, req *dto.ScopeUpdateRequest) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.Scope, error)
}

type scopeService struct {
	scopeRepo repository.ScopeRepository
	roleRepo  repository.RoleRepository
	logger    *zap.Logger
}

func NewScopeService(scopeRepo repository.ScopeRepository, roleRepo repository.RoleRepository, logger *zap.Logger) ScopeService {
	return &scopeService{scopeRepo: scopeRepo, roleRepo: roleRepo, logger: logger}
}

func (s *scopeService) List(req *dto.ScopeListRequest) ([]model.Scope, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.scopeRepo.List(offset, req.Size, req.Code, req.Name)
}

func (s *scopeService) Get(id uuid.UUID) (*model.Scope, error) {
	return s.scopeRepo.GetByID(id)
}

func (s *scopeService) Create(req *dto.ScopeCreateRequest) (*model.Scope, error) {
	_, err := s.scopeRepo.GetByCode(req.Code)
	if err == nil {
		return nil, ErrScopeCodeExists
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	scope := &model.Scope{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}
	if err := s.scopeRepo.Create(scope); err != nil {
		return nil, err
	}
	return s.scopeRepo.GetByID(scope.ID)
}

func (s *scopeService) Update(id uuid.UUID, req *dto.ScopeUpdateRequest) error {
	scope, err := s.scopeRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrScopeNotFound
		}
		return err
	}
	if req.Name != "" {
		scope.Name = req.Name
	}
	if req.Description != "" {
		scope.Description = req.Description
	}
	if req.SortOrder > 0 {
		scope.SortOrder = req.SortOrder
	}
	return s.scopeRepo.Update(scope)
}

func (s *scopeService) Delete(id uuid.UUID) error {
	scope, err := s.scopeRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrScopeNotFound
		}
		return err
	}
	// 检查是否为默认作用域（通过code判断）
	if scope.Code == "global" || scope.Code == "team" {
		return ErrScopeInUse
	}
	// 检查是否有角色在使用此作用域
	roles, err := s.roleRepo.GetByScopeID(id)
	if err != nil {
		return err
	}
	if len(roles) > 0 {
		return &ErrScopeInUseWithRoles{Roles: roles}
	}
	return s.scopeRepo.Delete(id)
}

func (s *scopeService) GetAll() ([]model.Scope, error) {
	return s.scopeRepo.GetAll()
}
