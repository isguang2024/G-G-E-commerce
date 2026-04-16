package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/jwt"
	"github.com/maben/backend/internal/pkg/password"
)

var (
	ErrInvalidCredentials = errors.New("邮箱或密码错误")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserExists         = errors.New("用户名已存在")
	ErrEmailExists        = errors.New("邮箱已存在")
	ErrUserInactive       = errors.New("账号已被禁用")
)

type AuthService interface {
	Login(username, password, ip string) (*dto.LoginResponse, error)
	Register(req *dto.RegisterRequest) (*dto.LoginResponse, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	GetUserInfo(userID uuid.UUID) (*user.User, error)
	// CreateUserTx 在调用方提供的事务内创建用户（不生成 token）。
	CreateUserTx(tx *gorm.DB, req *dto.RegisterRequest) (*user.User, error)
	// BuildLoginResponse 为已加载的用户对象生成 token 响应。
	BuildLoginResponse(u *user.User) (*dto.LoginResponse, error)
}

type authService struct {
	userRepo user.UserRepository
	jwtCfg   *config.JWTConfig
	logger   *zap.Logger
}

func NewAuthService(userRepo user.UserRepository, jwtCfg *config.JWTConfig, logger *zap.Logger) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
		logger:   logger,
	}
}

func (s *authService) Login(username, passwordStr, ip string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Login failed: user not found", zap.String("username", username))
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("Database error when finding user", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user.Status != "active" {
		s.logger.Warn("Login failed: user inactive", zap.String("username", username), zap.String("status", user.Status))
		return nil, ErrUserInactive
	}

	if !password.Verify(passwordStr, user.PasswordHash) {
		s.logger.Warn("Login failed: invalid password", zap.String("username", username))
		return nil, ErrInvalidCredentials
	}

	if err := s.userRepo.UpdateLastLogin(user.ID, ip); err != nil {
		s.logger.Warn("Failed to update last login", zap.Error(err))
	}

	accessToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret,
		user.ID.String(),
		"",
		user.Email,
		s.jwtCfg.AccessExpire,
	)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret,
		user.ID.String(),
		"",
		user.Email,
		s.jwtCfg.RefreshExpire,
	)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	roles := make([]map[string]interface{}, 0)
	for _, role := range user.Roles {
		roles = append(roles, map[string]interface{}{
			"id":          role.ID.String(),
			"code":        role.Code,
			"name":        role.Name,
			"description": role.Description,
		})
	}

	userInfo := map[string]interface{}{
		"id":             user.ID.String(),
		"email":          user.Email,
		"username":       user.Username,
		"nickname":       user.Nickname,
		"avatar_url":     user.AvatarURL,
		"phone":          user.Phone,
		"status":         user.Status,
		"is_super_admin": user.IsSuperAdmin,
		"roles":          roles,
		"created_at":     user.CreatedAt,
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtCfg.AccessExpire * 60,
		User:         userInfo,
	}, nil
}

func (s *authService) Register(req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		s.logger.Error("Failed to check username existence", zap.Error(err))
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	email := strings.TrimSpace(req.Email)
	if email != "" {
		exists, err := s.userRepo.ExistsByEmail(email)
		if err != nil {
			s.logger.Error("Failed to check email existence", zap.Error(err))
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if exists {
			return nil, ErrEmailExists
		}
	}

	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &user.User{
		Email:        email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Nickname:     req.Nickname,
		Status:       "active",
		IsSuperAdmin: false,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User registered successfully", zap.String("email", user.Email), zap.String("user_id", user.ID.String()))

	return s.Login(req.Username, req.Password, "")
}

func (s *authService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	claims, err := jwt.ParseToken(refreshToken, s.jwtCfg.Secret)
	if err != nil {
		s.logger.Warn("Failed to parse refresh token", zap.Error(err))
		return nil, errors.New("无效或已过期的 refresh token")
	}

	user, err := s.userRepo.GetByID(uuid.MustParse(claims.UserID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user.Status != "active" {
		return nil, ErrUserInactive
	}

	accessToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret,
		user.ID.String(),
		"",
		user.Email,
		s.jwtCfg.AccessExpire,
	)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	newRefreshToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret,
		user.ID.String(),
		"",
		user.Email,
		s.jwtCfg.RefreshExpire,
	)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.jwtCfg.AccessExpire * 60,
	}, nil
}

func (s *authService) GetUserInfo(userID uuid.UUID) (*user.User, error) {
	item, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return item, nil
}

// CreateUserTx 在调用方提供的事务内创建用户，不生成 token。
// 重复检查在事务内进行，确保原子性。
func (s *authService) CreateUserTx(tx *gorm.DB, req *dto.RegisterRequest) (*user.User, error) {
	var count int64
	if err := tx.Model(&user.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		if err := tx.Model(&user.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("check email: %w", err)
		}
		if count > 0 {
			return nil, ErrEmailExists
		}
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u := &user.User{
		Email:        email,
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     req.Nickname,
		Status:       "active",
		IsSuperAdmin: false,
	}
	if err := tx.Create(u).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// BuildLoginResponse 为已加载的用户对象生成访问令牌响应，不查询 DB。
func (s *authService) BuildLoginResponse(u *user.User) (*dto.LoginResponse, error) {
	accessToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret, u.ID.String(), "", u.Email, s.jwtCfg.AccessExpire,
	)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret, u.ID.String(), "", u.Email, s.jwtCfg.RefreshExpire,
	)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	roles := make([]map[string]interface{}, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, map[string]interface{}{
			"id": r.ID.String(), "code": r.Code, "name": r.Name, "description": r.Description,
		})
	}
	userInfo := map[string]interface{}{
		"id": u.ID.String(), "email": u.Email, "username": u.Username,
		"nickname": u.Nickname, "avatar_url": u.AvatarURL, "phone": u.Phone,
		"status": u.Status, "is_super_admin": u.IsSuperAdmin,
		"roles": roles, "created_at": u.CreatedAt,
	}
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtCfg.AccessExpire * 60,
		User:         userInfo,
	}, nil
}

