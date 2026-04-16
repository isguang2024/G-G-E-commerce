package handlers

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/maben/backend/internal/pkg/permission/evaluator"
)

func authWorkspaceIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	raw := strings.TrimSpace(stringFromCtx(ctx, CtxAuthWorkspaceID))
	if raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

func grantedPermissionKeysFromContext(ctx context.Context, eval evaluator.Evaluator, userID uuid.UUID) (map[string]struct{}, error) {
	if eval == nil || userID == uuid.Nil {
		return map[string]struct{}{}, nil
	}
	workspaceID, _ := authWorkspaceIDFromContext(ctx)
	resolved, err := eval.Resolve(ctx, userID, workspaceID)
	if err != nil {
		return nil, err
	}
	if resolved == nil || resolved.Keys == nil {
		return map[string]struct{}{}, nil
	}
	return resolved.Keys, nil
}
