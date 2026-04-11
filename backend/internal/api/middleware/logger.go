package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 日志中间件
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		requestID := c.GetHeader("X-Request-Id")

		c.Next()

		latency := time.Since(start)
		logger.Info("HTTP Request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("request_id", requestID),
			zap.String("app_key", c.GetString("app_key")),
			zap.String("space_key", c.GetString("space_key")),
			zap.String("app_resolved_by", c.GetString("app_resolved_by")),
			zap.String("tenant_id", c.GetString("tenant_id")),
			zap.String("auth_mode", c.GetString("app_auth_mode")),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}
