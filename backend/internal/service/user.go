package service

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/repository"
)

// UserService 用户管理服务
type UserService interface {
	List(req *dto.UserListRequest) ([]model.User, int64, error)
	Get(id uuid.UUID) (*model.User, error)
	Create(req *dto.UserCreateRequest) (*model.User, error)
	Update(id uuid.UUID, req *dto.UserUpdateRequest) error
	Delete(id uuid.UUID) error
	AssignRoles(id uuid.UUID, roleIDs []string) error
}

type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	logger   *zap.Logger
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, logger *zap.Logger) UserService {
	return &userService{userRepo: userRepo, roleRepo: roleRepo, logger: logger}
}

func (s *userService) List(req *dto.UserListRequest) ([]model.User, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.userRepo.List(offset, req.Size, req.UserName, req.UserPhone, req.UserEmail, req.Status, req.RoleID)
}

func (s *userService) Get(id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userService) Create(req *dto.UserCreateRequest) (*model.User, error) {
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
	user := &model.User{
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
	// 更新字段：如果请求中包含字段（即使为空字符串），也要更新
	// 使用指针判断是否传递了字段，但由于 JSON 绑定，空字符串也会被传递
	// 所以这里总是更新这些字段（允许设置为空）
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
	// 只有在明确传递了 RoleIDs 时才更新角色（nil 表示不更新，空数组表示清空）
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
