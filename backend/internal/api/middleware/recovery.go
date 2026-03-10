package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/errcode"
)

// Recovery 恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("remote_addr", c.ClientIP()),
				)
				logger.Error("Stack trace", zap.String("stack", string(debug.Stack())))
				status, resp := errcode.Response(errcode.ErrInternal)
				c.JSON(status, resp)
				c.Abort()
			}
		}()
		c.Next()
	}
}
