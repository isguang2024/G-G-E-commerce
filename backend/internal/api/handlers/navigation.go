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

	"github.com/maben/backend/api/gen"
)

func (h *navigationAPIHandler) GetNavigation(ctx context.Context, params gen.GetNavigationParams) (*gen.NavigationManifest, error) {
	appKey := strings.TrimSpace(optString(params.AppKey))
	spaceKey := strings.TrimSpace(optString(params.MenuSpaceKey))

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
	}
	if menuTree, err := mapJSON[[]gen.MenuTreeItem](manifest.MenuTree); err == nil {
		out.MenuTree = menuTree
	} else {
		return nil, err
	}
	if entryRoutes, err := mapJSON[[]gen.MenuTreeItem](manifest.EntryRoutes); err == nil {
		out.EntryRoutes = entryRoutes
	} else {
		return nil, err
	}
	if managedPages, err := mapJSON[[]gen.PageListItem](manifest.ManagedPages); err == nil {
		out.ManagedPages = managedPages
	} else {
		return nil, err
	}
	if manifest.CurrentApp != nil {
		currentApp, err := systemCurrentAppResponseFromModel(manifest.CurrentApp)
		if err != nil {
			return nil, err
		}
		if currentApp != nil {
			out.CurrentApp = gen.NewOptSystemCurrentAppResponse(*currentApp)
		}
	}
	if manifest.CurrentSpace != nil {
		currentSpace, err := systemCurrentMenuSpaceResponseFromModel(manifest.CurrentSpace)
		if err != nil {
			return nil, err
		}
		if currentSpace != nil {
			out.CurrentMenuSpace = gen.NewOptSystemCurrentMenuSpaceResponse(*currentSpace)
		}
	}
	if manifest.Context != nil {
		contextValue, err := mapJSON[gen.NavigationContext](manifest.Context)
		if err != nil {
			return nil, err
		}
		out.Context = gen.NewOptNavigationContext(contextValue)
	}
	return out, nil
}
