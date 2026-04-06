package authorization

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
)

const targetWorkspaceHeader = "X-Target-Workspace-Id"

type AuthorizationContext struct {
	UserID                          uuid.UUID
	AuthWorkspaceID                 *uuid.UUID
	AuthWorkspaceType               string
	TargetWorkspaceID               *uuid.UUID
	CurrentCollaborationWorkspaceID *uuid.UUID
	CollaborationWorkspaceID        *uuid.UUID
	AppKey                          string
}

func ResolveContext(c *gin.Context) (*AuthorizationContext, error) {
	if c == nil {
		return nil, ErrUnauthorized
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		return nil, err
	}

	authWorkspaceID, err := parseOptionalUUID(c.GetString("auth_workspace_id"))
	if err != nil {
		return nil, err
	}
	collaborationWorkspaceID, err := resolveContextCollaborationWorkspaceID(c)
	if err != nil {
		return nil, err
	}
	currentCollaborationWorkspaceID, err := parseOptionalUUID(c.GetString("collaboration_workspace_id"))
	if err != nil {
		return nil, err
	}
	targetWorkspaceID, err := resolveTargetWorkspaceID(c)
	if err != nil {
		return nil, err
	}

	return &AuthorizationContext{
		UserID:                          userID,
		AuthWorkspaceID:                 authWorkspaceID,
		AuthWorkspaceType:               normalizeWorkspaceType(c.GetString("auth_workspace_type"), collaborationWorkspaceID),
		TargetWorkspaceID:               targetWorkspaceID,
		CurrentCollaborationWorkspaceID: currentCollaborationWorkspaceID,
		CollaborationWorkspaceID:        collaborationWorkspaceID,
		AppKey:                          appctx.NormalizeAppKey(appctx.CurrentAppKey(c)),
	}, nil
}

func resolveTargetWorkspaceID(c *gin.Context) (*uuid.UUID, error) {
	candidates := []string{
		strings.TrimSpace(c.Query("target_workspace_id")),
		strings.TrimSpace(c.GetHeader(targetWorkspaceHeader)),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		return parseOptionalUUID(candidate)
	}
	bodyValue, err := resolveTargetWorkspaceIDFromBody(c)
	if err != nil {
		return nil, err
	}
	if bodyValue != "" {
		return parseOptionalUUID(bodyValue)
	}
	return nil, nil
}

func resolveTargetWorkspaceIDFromBody(c *gin.Context) (string, error) {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return "", nil
	}
	contentType := strings.ToLower(strings.TrimSpace(c.ContentType()))
	if !strings.Contains(contentType, "application/json") {
		return "", nil
	}
	switch strings.ToUpper(strings.TrimSpace(c.Request.Method)) {
	case "POST", "PUT", "PATCH":
	default:
		return "", nil
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return "", nil
	}

	payload := make(map[string]interface{})
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return "", nil
	}
	if value, ok := payload["target_workspace_id"]; ok {
		switch typed := value.(type) {
		case string:
			return strings.TrimSpace(typed), nil
		}
	}
	return "", nil
}

func parseOptionalUUID(value string) (*uuid.UUID, error) {
	target := strings.TrimSpace(value)
	if target == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(target)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func resolveContextCollaborationWorkspaceID(c *gin.Context) (*uuid.UUID, error) {
	if c == nil {
		return nil, nil
	}
	if parsed, err := parseOptionalUUID(c.GetString("collaboration_workspace_id")); err != nil {
		return nil, err
	} else if parsed != nil {
		return parsed, nil
	}
	if parsed, err := parseOptionalUUID(c.Query("collaboration_workspace_id")); err != nil {
		return nil, err
	} else if parsed != nil {
		return parsed, nil
	}
	if parsed, err := parseOptionalUUID(c.GetHeader("X-Collaboration-Workspace-Id")); err != nil {
		return nil, err
	} else if parsed != nil {
		return parsed, nil
	}
	return nil, nil
}

func normalizeWorkspaceType(value string, collaborationWorkspaceID *uuid.UUID) string {
	switch strings.TrimSpace(value) {
	case models.WorkspaceTypePersonal, models.WorkspaceTypeCollaboration:
		return strings.TrimSpace(value)
	default:
		if collaborationWorkspaceID != nil && *collaborationWorkspaceID != uuid.Nil {
			return models.WorkspaceTypeCollaboration
		}
		return models.WorkspaceTypePersonal
	}
}

func ensureWorkspaceContext(authCtx *AuthorizationContext) error {
	if authCtx == nil {
		return ErrUnauthorized
	}
	if authCtx.UserID == uuid.Nil {
		return ErrUnauthorized
	}
	if authCtx.AuthWorkspaceType == "" {
		return errors.New("workspace context missing")
	}
	return nil
}
