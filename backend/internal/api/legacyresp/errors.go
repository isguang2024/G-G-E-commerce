package legacyresp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gg-ecommerce/backend/internal/api/dto"
)

const (
	CodeUnauthorized   = 2001
	CodeTokenExpired   = 2002
	CodeForbidden      = 2003
	CodeAPIKeyMissing  = 2004
	CodeTokenBadFormat = 2005
	CodeInternal       = 5001
)

var defaultMessages = map[int]string{
	CodeUnauthorized:   "未登录或 token 无效",
	CodeTokenExpired:   "token 已过期",
	CodeForbidden:      "无权限",
	CodeAPIKeyMissing:  "缺少 API Key",
	CodeTokenBadFormat: "Token 格式错误",
	CodeInternal:       "服务器内部错误，请稍后重试",
}

func Write(c *gin.Context, status int, code int, message string) {
	if message == "" {
		message = defaultMessages[code]
	}
	c.JSON(status, dto.ErrorResponse(code, message))
	c.Abort()
}

func Unauthorized(c *gin.Context, message string) {
	Write(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

func TokenExpired(c *gin.Context, message string) {
	Write(c, http.StatusUnauthorized, CodeTokenExpired, message)
}

func Forbidden(c *gin.Context, message string) {
	Write(c, http.StatusForbidden, CodeForbidden, message)
}

func APIKeyMissing(c *gin.Context, message string) {
	Write(c, http.StatusUnauthorized, CodeAPIKeyMissing, message)
}

func TokenBadFormat(c *gin.Context, message string) {
	Write(c, http.StatusUnauthorized, CodeTokenBadFormat, message)
}

func Internal(c *gin.Context, message string) {
	Write(c, http.StatusInternalServerError, CodeInternal, message)
}
