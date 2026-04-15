package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/legacyresp"
)

// Recovery 恢复中间件：捕获 handler 里的 panic，写入 Error 级日志并返回 500。
// 日志里带上 request_id（如已被 RequestID 中间件注入），方便后续从 request_id
// 反查到审计表 / 前端 telemetry 同一条链路。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request_id", c.GetString("request_id")),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("remote_addr", c.ClientIP()),
					zap.String("stack", string(debug.Stack())),
				)
				legacyresp.Internal(c, "")
			}
		}()
		c.Next()
	}
}
