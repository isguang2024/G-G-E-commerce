package user

import (
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserService interface {
	List(req *dto.UserListRequest) ([]User, int64, error)
	Get(id uuid.UUID) (*User, error)
	GetByIDs(ids []uuid.UUID) ([]User, error)
	Create(req *dto.UserCreateRequest) (*User, error)
	Update(id uuid.UUID, req *dto.UserUpdateRequest) error
	Delete(id uuid.UUID) error
	AssignRoles(id uuid.UUID, roleIDs []string) error
}

type userService struct {
	userRepo UserRepository
	roleRepo interface {
		GetByID(id uuid.UUID) (*Role, error)
	}
	logger *zap.Logger
}

func NewUserService(userRepo UserRepository, roleRepo interface {
	GetByID(id uuid.UUID) (*Role, error)
}, logger *zap.Logger) UserService {
	return &userService{userRepo: userRepo, roleRepo: roleRepo, logger: logger}
}

func (s *userService) List(req *dto.UserListRequest) ([]User, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.userRepo.List(offset, req.Size, req.UserName, req.UserPhone, req.UserEmail, req.Status, req.RoleID, req.ID, req.RegisterSource, req.InvitedBy)
}

func (s *userService) Get(id uuid.UUID) (*User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByIDs(ids []uuid.UUID) ([]User, error) {
	return s.userRepo.GetByIDs(ids)
}

func (s *userService) Create(req *dto.UserCreateRequest) (*User, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	user := &User{
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
		Nickname:     req.Nickname,
		Phone:        req.Phone,
		SystemRemark: req.SystemRemark,
		Status:       status,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	roleUUIDs, _ := parseUUIDs(req.RoleIDs)
	if len(roleUUIDs) > 0 {
		_ = s.userRepo.ReplaceRoles(user.ID, roleUUIDs)
	}
	return s.userRepo.GetByID(user.ID)
}

func (s *userService) Update(id uuid.UUID, req *dto.UserUpdateRequest) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}
	user.Email = req.Email
	user.Nickname = req.Nickname
	user.Phone = req.Phone
	user.SystemRemark = req.SystemRemark
	if req.Status != "" {
		user.Status = req.Status
	}
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	if req.RoleIDs != nil {
		roleUUIDs, _ := parseUUIDs(req.RoleIDs)
		return s.userRepo.ReplaceRoles(id, roleUUIDs)
	}
	return nil
}

func (s *userService) Delete(id uuid.UUID) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}
	return s.userRepo.Delete(id)
}

func (s *userService) AssignRoles(id uuid.UUID, roleIDs []string) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}
	roleUUIDs, err := parseUUIDs(roleIDs)
	if err != nil {
		return err
	}
	return s.userRepo.ReplaceRoles(id, roleUUIDs)
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	var result []uuid.UUID
	for _, s := range ids {
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

type PermissionService interface {
	GetUserMenuIDs(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
}

type permissionService struct {
	userRepo     UserRepository
	userRoleRepo UserRoleRepository
	roleMenuRepo RoleMenuRepository
}

func NewPermissionService(userRepo UserRepository, userRoleRepo UserRoleRepository, roleMenuRepo RoleMenuRepository) PermissionService {
	return &permissionService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleMenuRepo: roleMenuRepo,
	}
}

func (s *permissionService) GetUserMenuIDs(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return []uuid.UUID{}, nil
	}

	roleIDs, err := s.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndTenant(userID, tenantID)
	if err != nil {
		return nil, err
	}

	return s.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
}
