package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/jwt"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrUserInactive       = errors.New("user account is inactive")
)

type AuthService interface {
	Login(username, password, ip string) (*dto.LoginResponse, error)
	Register(req *dto.RegisterRequest) (*dto.LoginResponse, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	GetUserInfo(userID uuid.UUID) (*user.User, error)
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
		return nil, errors.New("username already exists")
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
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userRepo.GetByID(uuid.MustParse(claims.UserID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
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
