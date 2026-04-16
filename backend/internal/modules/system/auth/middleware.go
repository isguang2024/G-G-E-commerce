package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/api/legacyresp"
	"github.com/maben/backend/internal/modules/system/models"
	workspacepkg "github.com/maben/backend/internal/modules/system/workspace"
	"github.com/maben/backend/internal/pkg/jwt"
)

const authWorkspaceHeader = "X-Auth-Workspace-Id"
const collaborationWorkspaceHeader = "X-Collaboration-Workspace-Id"

func JWTAuth(secret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			legacyresp.Unauthorized(c, "未授权，请先登录")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			legacyresp.TokenBadFormat(c, "")
			return
		}

		token := parts[1]

		claims, err := jwt.ParseToken(token, secret)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				legacyresp.TokenExpired(c, "")
			} else {
				legacyresp.Unauthorized(c, "无效的 Token")
			}
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		applyAuthorizationContext(c, claims, db)
		c.Next()
	}
}

func applyAuthorizationContext(c *gin.Context, claims *jwt.Claims, db *gorm.DB) {
	if c == nil || claims == nil || db == nil {
		return
	}

	userID, err := uuid.Parse(strings.TrimSpace(claims.UserID))
	if err != nil {
		return
	}

	workspaceService := workspacepkg.NewService(db, nil)
	var authWorkspace *models.Workspace

	// 关键安全校验：客户端传入的 X-Auth-Workspace-Id / X-Collaboration-Workspace-Id
	// 必须先确认当前 user 是该 workspace 的成员，否则任意登录者改 header 即可越权。
	resolveIfMember := func(ws *models.Workspace) *models.Workspace {
		if ws == nil {
			return nil
		}
		if _, err := workspaceService.GetMember(ws.ID, userID); err != nil {
			return nil
		}
		return ws
	}

	if authWorkspaceID := strings.TrimSpace(c.GetHeader(authWorkspaceHeader)); authWorkspaceID != "" {
		if parsedWorkspaceID, parseErr := uuid.Parse(authWorkspaceID); parseErr == nil {
			ws, _ := workspaceService.GetByID(parsedWorkspaceID)
			authWorkspace = resolveIfMember(ws)
		}
	}

	if authWorkspace == nil {
		collaborationWorkspaceID := strings.TrimSpace(c.GetHeader(collaborationWorkspaceHeader))
		if collaborationWorkspaceID == "" {
			collaborationWorkspaceID = strings.TrimSpace(claims.CollaborationWorkspaceID)
		}
		if collaborationWorkspaceID == "" {
			collaborationWorkspaceID = strings.TrimSpace(c.Query("collaboration_workspace_id"))
		}
		if collaborationWorkspaceID != "" {
			if parsedCollaborationWorkspaceID, parseErr := uuid.Parse(collaborationWorkspaceID); parseErr == nil {
				ws, _ := workspaceService.GetCollaborationWorkspaceByCollaborationWorkspaceID(parsedCollaborationWorkspaceID)
				authWorkspace = resolveIfMember(ws)
			}
		}
	}

	if authWorkspace == nil {
		authWorkspace, _ = workspaceService.EnsurePersonalWorkspaceForUser(userID)
	}

	if authWorkspace == nil {
		return
	}

	c.Set("auth_workspace_id", authWorkspace.ID.String())
	c.Set("auth_workspace_type", authWorkspace.WorkspaceType)
	if authWorkspace.CollaborationWorkspaceID != nil && *authWorkspace.CollaborationWorkspaceID != uuid.Nil {
		c.Set("collaboration_workspace_id", authWorkspace.CollaborationWorkspaceID.String())
		return
	}
	c.Set("collaboration_workspace_id", "")
}

