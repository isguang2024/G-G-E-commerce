package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	workspacepkg "github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	"github.com/gg-ecommerce/backend/internal/pkg/jwt"
)

const authWorkspaceHeader = "X-Auth-Workspace-Id"
const collaborationWorkspaceHeader = "X-Collaboration-Workspace-Id"

func JWTAuth(secret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "未授权，请先登录")
			c.JSON(status, resp)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			status, resp := errcode.Response(errcode.ErrTokenBadFormat)
			c.JSON(status, resp)
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := jwt.ParseToken(token, secret)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				status, resp := errcode.Response(errcode.ErrTokenExpired)
				c.JSON(status, resp)
			} else {
				status, resp := errcode.ResponseWithMsg(errcode.ErrUnauthorized, "无效的 Token")
				c.JSON(status, resp)
			}
			c.Abort()
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

	if authWorkspaceID := strings.TrimSpace(c.GetHeader(authWorkspaceHeader)); authWorkspaceID != "" {
		if parsedWorkspaceID, parseErr := uuid.Parse(authWorkspaceID); parseErr == nil {
			authWorkspace, _ = workspaceService.GetByID(parsedWorkspaceID)
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
				authWorkspace, _ = workspaceService.GetCollaborationWorkspaceByCollaborationWorkspaceID(parsedCollaborationWorkspaceID)
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
		c.Set("legacy_collaboration_workspace_id", authWorkspace.CollaborationWorkspaceID.String())
		return
	}
	c.Set("collaboration_workspace_id", "")
	c.Set("legacy_collaboration_workspace_id", "")
}
