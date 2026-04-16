package appctx

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/maben/backend/internal/modules/system/models"
)

const (
	RequestAppKeyQuery  = "app_key"
	RequestAppKeyHeader = "X-App-Key"
)

var ErrAppKeyRequired = errors.New("app_key is required")

func NormalizeExplicitAppKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func NormalizeAppKey(value string) string {
	target := NormalizeExplicitAppKey(value)
	if target == "" {
		return models.DefaultAppKey
	}
	return target
}

func ExplicitRequestAppKey(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value := strings.TrimSpace(c.Query(RequestAppKeyQuery)); value != "" {
		return NormalizeExplicitAppKey(value)
	}
	if value := strings.TrimSpace(c.GetHeader(RequestAppKeyHeader)); value != "" {
		return NormalizeExplicitAppKey(value)
	}
	return ""
}

func RequestAppKey(c *gin.Context) string {
	if value := ExplicitRequestAppKey(c); value != "" {
		return NormalizeAppKey(value)
	}
	return models.DefaultAppKey
}

func RequireRequestAppKey(c *gin.Context) (string, error) {
	value := ExplicitRequestAppKey(c)
	if value == "" {
		return "", ErrAppKeyRequired
	}
	return NormalizeAppKey(value), nil
}

func ResolveManagedAppKey(explicit string, c *gin.Context) (string, error) {
	if value := NormalizeExplicitAppKey(explicit); value != "" {
		return NormalizeAppKey(value), nil
	}
	return RequireRequestAppKey(c)
}

func CurrentAppKey(c *gin.Context) string {
	if c == nil {
		return models.DefaultAppKey
	}
	if value := NormalizeAppKey(c.GetString("app_key")); value != "" {
		return value
	}
	return RequestAppKey(c)
}

