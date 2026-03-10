package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/pkg/jwt"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrUserInactive       = errors.New("user account is inactive")
)

// AuthService 认证服务接口
type AuthService interface {
	Login(username, password, ip string) (*dto.LoginResponse, error)
	Register(req *dto.RegisterRequest) (*dto.LoginResponse, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	GetUserInfo(userID uuid.UUID) (*model.User, error)
}

// authService 认证服务实现
type authService struct {
	userRepo repository.UserRepository
	jwtCfg   *config.JWTConfig
	logger   *zap.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository, jwtCfg *config.JWTConfig, logger *zap.Logger) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
		logger:   logger,
	}
}

// Login 登录
func (s *authService) Login(username, passwordStr, ip string) (*dto.LoginResponse, error) {
	// 查找用户（优先使用用户名）
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Login failed: user not found", zap.String("username", username))
			return nil, ErrInvalidCredentials
		}
		// 数据库错误
		s.logger.Error("Database error when finding user", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		s.logger.Warn("Login failed: user inactive", zap.String("username", username), zap.String("status", user.Status))
		return nil, ErrUserInactive
	}

	// 验证密码
	if !password.Verify(passwordStr, user.PasswordHash) {
		s.logger.Warn("Login failed: invalid password", zap.String("username", username))
		return nil, ErrInvalidCredentials
	}

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(user.ID, ip); err != nil {
		s.logger.Warn("Failed to update last login", zap.Error(err))
	}

	// 生成 Token
	accessToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret,
		user.ID.String(),
		"", // TenantID 暂时为空，后续根据业务需求添加
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

	// 构建用户信息（不包含敏感信息）
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
		ExpiresIn:    s.jwtCfg.AccessExpire * 60, // 转换为秒
		User:         userInfo,
	}, nil
}

// Register 注册
func (s *authService) Register(req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	// 检查用户名是否已存在（用户名必填）
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		s.logger.Error("Failed to check username existence", zap.Error(err))
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// 如果提供了邮箱，检查邮箱是否已存在
	if req.Email != "" {
		exists, err := s.userRepo.ExistsByEmail(req.Email)
		if err != nil {
			s.logger.Error("Failed to check email existence", zap.Error(err))
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if exists {
			return nil, ErrUserExists
		}
	}

	// 加密密码
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &model.User{
		Email:        req.Email,
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

	// 自动登录（使用用户名）
	return s.Login(req.Username, req.Password, "")
}

// RefreshToken 刷新 Token
func (s *authService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	// 解析 Token
	claims, err := jwt.ParseToken(refreshToken, s.jwtCfg.Secret)
	if err != nil {
		s.logger.Warn("Failed to parse refresh token", zap.Error(err))
		return nil, errors.New("invalid refresh token")
	}

	// 验证用户是否存在
	user, err := s.userRepo.GetByID(uuid.MustParse(claims.UserID))
	if err != nil {
		s.logger.Warn("User not found for refresh token", zap.String("user_id", claims.UserID))
		return nil, ErrUserNotFound
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, ErrUserInactive
	}

	// 生成新的 Token
	accessToken, err := jwt.GenerateToken(
		s.jwtCfg.Secret,
		user.ID.String(),
		claims.TenantID,
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
		claims.TenantID,
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

// GetUserInfo 获取用户信息
func (s *authService) GetUserInfo(userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found", zap.String("user_id", userID.String()))
			return nil, ErrUserNotFound
		}
		s.logger.Error("Database error when getting user info", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}
	return user, nil
}
