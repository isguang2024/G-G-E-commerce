package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

// RequestIDHeader 是前后端约定的 Request ID HTTP header 名。
// 前端的 request interceptor 会透传或回显这个头部，日志 / 审计 / 遥测
// 三条管道都以它作为 join key。
const RequestIDHeader = "X-Request-Id"

// maxRequestIDLength 限制入站 Request ID 头部长度，防止日志字段被客户端撑爆。
const maxRequestIDLength = 64

// RequestID 中间件确保每个请求都有一个稳定的 request_id：
//   - 若客户端已在 X-Request-Id 头部传入合法 ASCII 字符串（≤64），原样透传；
//   - 否则服务端生成 UUID v7（带时序，便于日志排序）；失败降级 v4。
//
// 中间件同时把 request_id 写入 gin.Context（"request_id" 字符串键）、
// request.Context（logger.WithRequestID 注入）与 response header，
// 后续中间件 / handler / ogen bridge / 业务 logger 都能通过同一个 key 拿到。
// 约定：RequestID 必须是 router 上的第一个 middleware，早于 Logger / Recovery。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if rid == "" || len(rid) > maxRequestIDLength || !isPrintableASCII(rid) {
			rid = newRequestID()
		}

		c.Set("request_id", rid)
		ctx := logger.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(RequestIDHeader, rid)

		c.Next()
	}
}

// newRequestID 优先 UUID v7（带时序），失败降级 v4，保证入站链路始终有 ID。
func newRequestID() string {
	if id, err := uuid.NewV7(); err == nil {
		return id.String()
	}
	return uuid.NewString()
}

// isPrintableASCII 拒绝控制字符 / 非 ASCII，避免客户端塞脏数据到日志字段里。
func isPrintableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}
