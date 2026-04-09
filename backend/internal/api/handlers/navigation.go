// navigation.go: ogen Handler for /runtime/navigation.
// Phase 4 — navigation domain migration. Compiles the runtime manifest via
// the existing navigation.Compiler and converts the gin.H payload into the
// generated NavigationManifest schema.
package handlers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
)

func (h *APIHandler) GetNavigation(ctx context.Context, params gen.GetNavigationParams) (*gen.NavigationManifest, error) {
	appKey := strings.TrimSpace(optString(params.AppKey))
	spaceKey := strings.TrimSpace(optString(params.SpaceKey))

	var userID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		userID = &uid
	}
	var cwID *uuid.UUID
	if raw := strings.TrimSpace(stringFromCtx(ctx, CtxCollaborationWorkspaceID)); raw != "" {
		if parsed, err := uuid.Parse(raw); err == nil {
			cwID = &parsed
		}
	}

	manifest, err := h.navSvc.Compile(appKey, requestHostFromCtx(ctx), spaceKey, userID, cwID)
	if err != nil {
		h.logger.Error("compile navigation failed", zap.Error(err))
		return nil, err
	}

	out := &gen.NavigationManifest{
		VersionStamp: manifest.VersionStamp,
		MenuTree:     marshalList(manifest.MenuTree),
		EntryRoutes:  marshalList(manifest.EntryRoutes),
		ManagedPages: marshalList(manifest.ManagedPages),
	}
	if manifest.CurrentApp != nil {
		out.CurrentApp = gen.NewOptAnyObject(marshalAnyObject(manifest.CurrentApp))
	}
	if manifest.CurrentSpace != nil {
		out.CurrentSpace = gen.NewOptAnyObject(marshalAnyObject(manifest.CurrentSpace))
	}
	if manifest.Context != nil {
		out.Context = gen.NewOptAnyObject(marshalAnyObject(manifest.Context))
	}
	return out, nil
}
