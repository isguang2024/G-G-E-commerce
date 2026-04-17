package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	apppkg "github.com/maben/backend/internal/modules/system/app"
	spacepkg "github.com/maben/backend/internal/modules/system/space"
	appctx "github.com/maben/backend/internal/pkg/appctx"
)

const (
	collaborationWorkspaceHeader = "X-Collaboration-Workspace-Id"
)

func AppContext(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.Next()
			return
		}

		host := spacepkg.RequestHost(c)
		path := ""
		if c.Request != nil && c.Request.URL != nil {
			path = c.Request.URL.Path
		}
		requestedAppKey := appctx.RequestAppKey(c)
		appKey, appBinding, appResolvedBy, err := apppkg.ResolveAppEntry(db, host, path, requestedAppKey)
		if err != nil {
			appKey = appctx.NormalizeAppKey(requestedAppKey)
			if appKey == "" {
				appKey = appctx.NormalizeAppKey("")
			}
			appResolvedBy = "fallback_default"
		}

		userID := contextUUID(c, "user_id")
		collaborationWorkspaceID := contextUUID(c, "collaboration_workspace_id")
		if collaborationWorkspaceID == nil {
			collaborationWorkspaceID = headerUUID(c.GetHeader(collaborationWorkspaceHeader))
		}

		// 优先尝试 Level 2 入口绑定（path 感知）
		spaceKey, spaceResolvedBy := "", ""
		if entryMenuSpaceKey, entryResolvedBy, entryErr := apppkg.ResolveMenuSpaceEntry(db, appKey, host, path); entryErr == nil && strings.TrimSpace(entryMenuSpaceKey) != "" && entryResolvedBy == "entry_binding" {
			// P1: 校验用户是否有权访问该空间，无权则继续走常规解析（回退默认空间）。
			allowed, _ := spacepkg.CanAccessSpace(db, userID, collaborationWorkspaceID, appKey, entryMenuSpaceKey)
			if allowed {
				spaceKey = entryMenuSpaceKey
				spaceResolvedBy = entryResolvedBy
			}
		}
		if spaceKey == "" {
			var resolveErr error
			spaceKey, spaceResolvedBy, resolveErr = spacepkg.ResolveCurrentMenuSpaceKey(
				db,
				appKey,
				host,
				spacepkg.RequestMenuSpaceKey(c),
				userID,
				collaborationWorkspaceID,
			)
			if resolveErr != nil {
				spaceKey = spacepkg.DefaultMenuSpaceKey
				spaceResolvedBy = "fallback_default"
			}
		}

		c.Set("request_host", host)
		c.Set("app_key", appKey)
		c.Set("app_resolved_by", appResolvedBy)
		if appBinding != nil {
			c.Set("app_entry_binding", appBinding)
		}
		c.Set("menu_space_key", spaceKey)
		c.Set("resolved_by", spaceResolvedBy)
		if collaborationWorkspaceID != nil {
			c.Set("collaboration_workspace_id", collaborationWorkspaceID.String())
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
