package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 访问日志中间件：为每个 HTTP 请求输出一条结构化 JSON。
//
// 读取顺序约定：
//   - request_id：优先来自请求入口 RequestID 中间件写入的 gin.Context "request_id"
//     键（与客户端透传/服务端生成保持一致），否则退回客户端原始头部；
//   - app_key / menu_space_key / tenant_id：由 AppContext 中间件注入；
//   - auth_mode：由 DynamicAppSecurity 中间件注入。
//
// 本中间件只负责打 access log，不写审计表；审计由 audit.Recorder 单独负责。
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		requestID := c.GetString("request_id")
		if requestID == "" {
			// 兜底：RequestID 中间件未启用时，至少读一下入站头部
			requestID = c.GetHeader(RequestIDHeader)
		}
		latency := time.Since(start)
		logger.Info("HTTP Request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("request_id", requestID),
			zap.String("app_key", c.GetString("app_key")),
			zap.String("menu_space_key", c.GetString("menu_space_key")),
			zap.String("app_resolved_by", c.GetString("app_resolved_by")),
			zap.String("tenant_id", c.GetString("tenant_id")),
			zap.String("auth_mode", c.GetString("app_auth_mode")),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}
