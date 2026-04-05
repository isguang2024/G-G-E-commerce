package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
)

const tenantHeader = "X-Tenant-ID"

func AppContext(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.Next()
			return
		}

		host := spacepkg.RequestHost(c)
		requestedAppKey := appctx.RequestAppKey(c)
		appKey, _, appResolvedBy, err := apppkg.ResolveAppByHost(db, host, requestedAppKey)
		if err != nil {
			appKey = appctx.NormalizeAppKey(requestedAppKey)
			if appKey == "" {
				appKey = appctx.NormalizeAppKey("")
			}
			appResolvedBy = "fallback_default"
		}

		userID := contextUUID(c, "user_id")
		tenantID := contextUUID(c, "tenant_id")
		if tenantID == nil {
			tenantID = headerUUID(c.GetHeader(tenantHeader))
		}

		spaceKey, spaceResolvedBy, resolveErr := spacepkg.ResolveCurrentSpaceKey(
			db,
			appKey,
			host,
			spacepkg.RequestSpaceKey(c),
			userID,
			tenantID,
		)
		if resolveErr != nil {
			spaceKey = spacepkg.DefaultMenuSpaceKey
			spaceResolvedBy = "fallback_default"
		}

		c.Set("request_host", host)
		c.Set("app_key", appKey)
		c.Set("app_resolved_by", appResolvedBy)
		c.Set("space_key", spaceKey)
		c.Set("resolved_by", spaceResolvedBy)
		if tenantID != nil {
			c.Set("tenant_id", tenantID.String())
		}

		c.Next()
	}
}

func contextUUID(c *gin.Context, key string) *uuid.UUID {
	if c == nil {
		return nil
	}
	raw, ok := c.Get(key)
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	return headerUUID(value)
}

func headerUUID(value string) *uuid.UUID {
	target := strings.TrimSpace(value)
	if target == "" {
		return nil
	}
	parsed, err := uuid.Parse(target)
	if err != nil {
		return nil
	}
	return &parsed
}
