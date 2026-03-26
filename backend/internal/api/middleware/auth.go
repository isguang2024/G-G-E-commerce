package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/jwt"
)

// JWTAuth JWT 认证中间件
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "未授权，请先登录")
			c.JSON(status, resp)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			status, resp := errcode.Response(errcode.ErrTokenBadFormat)
			c.JSON(status, resp)
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwt.ParseToken(token, secret)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				status, resp := errcode.Response(errcode.ErrTokenExpired)
				c.JSON(status, resp)
			} else {
				status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "无效的 Token")
				c.JSON(status, resp)
			}
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// APIKeyAuth API Key 认证中间件
func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "API Key 认证未初始化")
			c.JSON(status, resp)
			c.Abort()
			return
		}

		apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))
		if apiKey == "" {
			status, resp := errcode.Response(errcode.ErrAPIKeyMissing)
			c.JSON(status, resp)
			c.Abort()
			return
		}

		var record models.APIKey
		candidates := []string{hashAPIKey(apiKey), apiKey}
		if err := db.Where("key_hash IN ?", candidates).First(&record).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "API Key 校验失败")
				c.JSON(status, resp)
				c.Abort()
				return
			}
			status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "无效的 API Key")
			c.JSON(status, resp)
			c.Abort()
			return
		}

		if record.ExpiresAt != nil && record.ExpiresAt.Before(time.Now()) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "API Key 已过期")
			c.JSON(status, resp)
			c.Abort()
			return
		}
		var tenant models.Tenant
		if err := db.Select("id", "status").Where("id = ?", record.TenantID).First(&tenant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "API Key 所属团队不存在")
				c.JSON(status, resp)
				c.Abort()
				return
			}
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "API Key 租户校验失败")
			c.JSON(status, resp)
			c.Abort()
			return
		}
		if strings.TrimSpace(tenant.Status) != "" && tenant.Status != "active" {
			status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "API Key 所属团队不可用")
			c.JSON(status, resp)
			c.Abort()
			return
		}

		now := time.Now()
		_ = db.Model(&models.APIKey{}).Where("id = ?", record.ID).Update("last_used_at", &now).Error

		c.Set("api_key", apiKey)
		c.Set("api_key_id", record.ID.String())
		c.Set("tenant_id", record.TenantID.String())
		c.Next()
	}
}

func hashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(apiKey)))
	return hex.EncodeToString(sum[:])
}
