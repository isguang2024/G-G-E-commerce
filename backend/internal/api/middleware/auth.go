package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/legacyresp"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/jwt"
)

// JWTAuth JWT 认证中间件
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			legacyresp.Unauthorized(c, "未授权，请先登录")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			legacyresp.TokenBadFormat(c, "")
			return
		}

		token := parts[1]
		claims, err := jwt.ParseToken(token, secret)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				legacyresp.TokenExpired(c, "")
			} else {
				legacyresp.Unauthorized(c, "无效的 Token")
			}
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("collaboration_workspace_id", claims.CollaborationWorkspaceID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// APIKeyAuth API Key 认证中间件
func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			legacyresp.Internal(c, "API Key 认证未初始化")
			return
		}

		apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))
		if apiKey == "" {
			legacyresp.APIKeyMissing(c, "")
			return
		}

		var record models.APIKey
		candidates := []string{hashAPIKey(apiKey), apiKey}
		if err := db.Where("key_hash IN ?", candidates).First(&record).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				legacyresp.Internal(c, "API Key 校验失败")
				return
			}
			legacyresp.Unauthorized(c, "无效的 API Key")
			return
		}

		if record.ExpiresAt != nil && record.ExpiresAt.Before(time.Now()) {
			legacyresp.Unauthorized(c, "API Key 已过期")
			return
		}
		var collaborationWorkspace models.CollaborationWorkspace
		if err := db.Select("id", "status").Where("id = ?", record.CollaborationWorkspaceID).First(&collaborationWorkspace).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				legacyresp.Unauthorized(c, "API Key 所属协作空间不存在")
				return
			}
			legacyresp.Internal(c, "API Key 协作空间校验失败")
			return
		}
		if strings.TrimSpace(collaborationWorkspace.Status) != "" && collaborationWorkspace.Status != "active" {
			legacyresp.Unauthorized(c, "API Key 所属协作空间不可用")
			return
		}

		now := time.Now()
		_ = db.Model(&models.APIKey{}).Where("id = ?", record.ID).Update("last_used_at", &now).Error

		c.Set("api_key", apiKey)
		c.Set("api_key_id", record.ID.String())
		c.Set("collaboration_workspace_id", record.CollaborationWorkspaceID.String())
		c.Next()
	}
}

func hashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(apiKey)))
	return hex.EncodeToString(sum[:])
}
