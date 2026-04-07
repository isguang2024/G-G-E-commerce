// phase4_extras.go — Phase 4 ogen handlers for the small leftover endpoints
// that have a clean service boundary: system fast-enter, view-pages, and the
// two simplest user sub-routes (collaboration workspaces, refresh snapshot).
package handlers

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	systemmod "github.com/gg-ecommerce/backend/internal/modules/system/system"
)

// ── system fast-enter ───────────────────────────────────────────────────────

func (h *APIHandler) GetFastEnterConfig(ctx context.Context) (gen.AnyObject, error) {
	cfg, err := h.systemFacade.GetFastEnterConfig()
	if err != nil {
		h.logger.Error("get fast enter config failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(cfg), nil
}

func (h *APIHandler) UpdateFastEnterConfig(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	var body systemmod.FastEnterSaveRequestPublic
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if _, err := h.systemFacade.UpdateFastEnterConfig(body); err != nil {
		h.logger.Error("update fast enter config failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ── system view-pages ───────────────────────────────────────────────────────

func (h *APIHandler) GetSystemViewPages(ctx context.Context, params gen.GetSystemViewPagesParams) (*gen.ViewPagesResponse, error) {
	force := false
	if params.Force.Set {
		v := strings.ToLower(strings.TrimSpace(params.Force.Value))
		force = v == "1" || v == "true"
	}
	pages, refreshed, refreshedAt, err := h.systemFacade.GetViewPages(ctx, force)
	if err != nil {
		h.logger.Error("get system view pages failed", zap.Error(err))
		return nil, err
	}
	out := &gen.ViewPagesResponse{
		Pages:       marshalList(pages),
		Refreshed:   refreshed,
		RefreshedAt: refreshedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	return out, nil
}

// ── user collaboration workspaces ───────────────────────────────────────────

func (h *APIHandler) GetUserCollaborationWorkspaces(ctx context.Context, params gen.GetUserCollaborationWorkspacesParams) (gen.AnyObject, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		return nil, err
	}
	if h.cwMemberRepo == nil {
		return marshalAnyObject(map[string]interface{}{"records": []interface{}{}}), nil
	}
	items, err := h.cwMemberRepo.GetCollaborationWorkspacesByUserID(params.ID)
	if err != nil {
		h.logger.Error("get user collaboration workspaces failed", zap.Error(err))
		return nil, err
	}
	records := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		records = append(records, map[string]interface{}{
			"id":     item.ID.String(),
			"name":   item.Name,
			"status": item.Status,
		})
	}
	return marshalAnyObject(map[string]interface{}{"records": records}), nil
}

// ── refresh user permission snapshot ────────────────────────────────────────

func (h *APIHandler) RefreshUserPermissionSnapshot(ctx context.Context, params gen.RefreshUserPermissionSnapshotParams) (*gen.MutationResult, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		return nil, err
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshPersonalWorkspaceUser(params.ID); err != nil {
			h.logger.Error("refresh personal workspace snapshot failed", zap.Error(err))
			return nil, err
		}
	}
	return ok(), nil
}
