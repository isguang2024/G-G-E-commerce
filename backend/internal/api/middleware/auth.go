package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gg-ecommerce/backend/internal/api/errcode"
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

		// 提取 token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			status, resp := errcode.Response(errcode.ErrTokenBadFormat)
			c.JSON(status, resp)
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 token 并解析用户信息
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

		// 将用户信息存入 context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// APIKeyAuth API Key 认证中间件
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现 API Key 认证逻辑
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			status, resp := errcode.Response(errcode.ErrAPIKeyMissing)
			c.JSON(status, resp)
			c.Abort()
			return
		}

		// TODO: 验证 API Key 并获取 tenant_id
		c.Set("api_key", apiKey)
		c.Next()
	}
}
