package scope

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

var (
	ErrScopeNotFound   = errors.New("scope not found")
	ErrScopeCodeExists = errors.New("scope code already exists")
	ErrScopeInUse      = errors.New("scope is in use and cannot be deleted")
)

type ErrScopeInUseWithRoles struct {
	Roles []user.Role
}

func (e *ErrScopeInUseWithRoles) Error() string {
	return "scope is in use and cannot be deleted"
}

type ScopeService interface {
	List(req *dto.ScopeListRequest) ([]user.Scope, int64, error)
	Get(id uuid.UUID) (*user.Scope, error)
	Create(req *dto.ScopeCreateRequest) (*user.Scope, error)
	Update(id uuid.UUID, req *dto.ScopeUpdateRequest) error
	Delete(id uuid.UUID) error
	GetAll() ([]user.Scope, error)
}

type scopeService struct {
	scopeRepo user.ScopeRepository
	roleRepo  user.RoleRepository
	logger    *zap.Logger
}

func NewScopeService(scopeRepo user.ScopeRepository, roleRepo user.RoleRepository, logger *zap.Logger) ScopeService {
	return &scopeService{
		scopeRepo: scopeRepo,
		roleRepo:  roleRepo,
		logger:    logger,
	}
}

func (s *scopeService) List(req *dto.ScopeListRequest) ([]user.Scope, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.scopeRepo.List(offset, req.Size, req.Code, req.Name)
}

func (s *scopeService) Get(id uuid.UUID) (*user.Scope, error) {
	return s.scopeRepo.GetByID(id)
}

func (s *scopeService) Create(req *dto.ScopeCreateRequest) (*user.Scope, error) {
	_, err := s.scopeRepo.GetByCode(req.Code)
	if err == nil {
		return nil, ErrScopeCodeExists
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	scope := &user.Scope{
		Code:               req.Code,
		Name:               req.Name,
		Description:        req.Description,
		IsSystem:           false,
		ContextKind:        normalizeScopeContextKind(req.ContextKind),
		DataPermissionCode: normalizeDataPermissionCode(req.DataPermissionCode),
		DataPermissionName: normalizeDataPermissionName(req.DataPermissionName, req.Name),
		SortOrder:          req.SortOrder,
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
	if req.ContextKind != "" {
		scope.ContextKind = normalizeScopeContextKind(req.ContextKind)
	}
	scope.DataPermissionCode = normalizeDataPermissionCode(req.DataPermissionCode)
	scope.DataPermissionName = normalizeDataPermissionName(req.DataPermissionName, scope.Name)
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
	if scope.IsSystem {
		return ErrScopeInUse
	}
	roles, err := s.roleRepo.GetByScopeID(id)
	if err != nil {
		return err
	}
	if len(roles) > 0 {
		return &ErrScopeInUseWithRoles{Roles: roles}
	}
	return s.scopeRepo.Delete(id)
}

func (s *scopeService) GetAll() ([]user.Scope, error) {
	return s.scopeRepo.GetAll()
}

func normalizeScopeContextKind(raw string) string {
	if raw == "tenant" {
		return "tenant"
	}
	return "global"
}

func normalizeDataPermissionCode(raw string) string {
	return strings.TrimSpace(raw)
}

func normalizeDataPermissionName(raw string, fallback string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}
