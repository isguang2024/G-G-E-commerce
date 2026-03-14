package auth

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/pkg/jwt"
)

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
