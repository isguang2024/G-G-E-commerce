package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
)

type AuthHandler struct {
	authService  AuthService
	authzService interface {
		GetUserActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error)
		GetUserScopedActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error)
	}
	logger *zap.Logger
}

func NewAuthHandler(authService AuthService, authzService interface {
	GetUserActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error)
	GetUserScopedActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error)
}, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		authzService: authzService,
		logger:       logger,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	clientIP := c.ClientIP()

	resp, err := h.authService.Login(req.Username, req.Password, clientIP)
	if err != nil {
		h.logger.Error("Login failed", zap.String("username", req.Username), zap.Error(err))
		var status int
		var respBody *dto.Response
		if err == ErrUserInactive {
			status, respBody = errcode.ResponseWithMsg(errcode.ErrForbidden, "User account is inactive")
		} else if err == ErrInvalidCredentials {
			status, respBody = errcode.Response(errcode.ErrUnauthorized)
		} else {
			h.logger.Error("Internal server error during login", zap.Error(err))
			status, respBody = errcode.ResponseWithMsg(errcode.ErrInternal, "服务器内部错误，请稍后重试")
		}
		c.JSON(status, respBody)
		return
	}

	h.logger.Info("User logged in successfully", zap.String("username", req.Username))
	c.JSON(http.StatusOK, dto.SuccessResponse(resp))
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid register request", zap.Error(err))
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		h.logger.Warn("Register failed", zap.String("email", req.Email), zap.Error(err))
		if err == ErrUserExists {
			status, resp := errcode.Response(errcode.ErrUsernameExists)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
		c.JSON(status, resp)
		return
	}

	h.logger.Info("User registered successfully", zap.String("email", req.Email))
	c.JSON(http.StatusOK, dto.SuccessResponse(resp))
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Refresh token failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrTokenExpired, "Invalid or expired refresh token")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(resp))
}

func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("User ID not found in context")
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}

	userIDStrValue, ok := userIDStr.(string)
	if !ok {
		h.logger.Error("Invalid user ID type", zap.Any("user_id", userIDStr))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "Invalid user ID")
		c.JSON(status, resp)
		return
	}

	userID, err := uuid.Parse(userIDStrValue)
	if err != nil {
		h.logger.Error("Failed to parse user ID", zap.String("user_id", userIDStrValue), zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "Invalid user ID")
		c.JSON(status, resp)
		return
	}

	user, err := h.authService.GetUserInfo(userID)
	if err != nil {
		h.logger.Error("Failed to get user info", zap.String("user_id", userID.String()), zap.Error(err))
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "服务器内部错误，请稍后重试")
		c.JSON(status, resp)
		return
	}

	roles := make([]map[string]interface{}, 0)
	if user.Roles != nil && len(user.Roles) > 0 {
		for _, role := range user.Roles {
			roles = append(roles, map[string]interface{}{
				"id":          role.ID.String(),
				"code":        role.Code,
				"name":        role.Name,
				"description": role.Description,
			})
		}
	}

	var tenantID *uuid.UUID
	tenantIDStr := strings.TrimSpace(c.Query("tenant_id"))
	if tenantIDStr == "" {
		tenantIDStr = strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
	}
	if tenantIDStr != "" {
		if parsedTenantID, parseErr := uuid.Parse(tenantIDStr); parseErr == nil {
			tenantID = &parsedTenantID
		}
	}

	actionKeys := make([]string, 0)
	scopedActionKeys := make([]string, 0)
	if h.authzService != nil {
		if keys, keyErr := h.authzService.GetUserActionKeys(userID, tenantID); keyErr != nil {
			h.logger.Warn("Failed to resolve user actions", zap.Error(keyErr), zap.String("user_id", userID.String()))
		} else {
			actionKeys = keys
		}
		if scopedKeys, scopedKeyErr := h.authzService.GetUserScopedActionKeys(userID, tenantID); scopedKeyErr != nil {
			h.logger.Warn("Failed to resolve user scoped actions", zap.Error(scopedKeyErr), zap.String("user_id", userID.String()))
		} else {
			scopedActionKeys = scopedKeys
		}
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
		"actions":        actionKeys,
		"scoped_actions": scopedActionKeys,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
	}
	if tenantID != nil {
		userInfo["current_tenant_id"] = tenantID.String()
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(userInfo))
}
