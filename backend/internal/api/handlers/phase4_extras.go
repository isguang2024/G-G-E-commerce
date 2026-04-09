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

func (h *APIHandler) GetFastEnterConfig(ctx context.Context) (*gen.SystemFastEnterConfig, error) {
	cfg, err := h.systemFacade.GetFastEnterConfig()
	if err != nil {
		h.logger.Error("get fast enter config failed", zap.Error(err))
		return nil, err
	}
	out := fastEnterConfigFromModel(cfg)
	return &out, nil
}

func (h *APIHandler) UpdateFastEnterConfig(ctx context.Context, req *gen.SystemFastEnterConfig) (*gen.SystemFastEnterConfig, error) {
	if req == nil {
		req = &gen.SystemFastEnterConfig{}
	}
	body := systemmod.FastEnterSaveRequestPublic{
		Applications: fastEnterApplicationsFromGen(req.Applications),
		QuickLinks:   fastEnterQuickLinksFromGen(req.QuickLinks),
		MinWidth:     optInt(req.MinWidth, 0),
	}
	cfg, err := h.systemFacade.UpdateFastEnterConfig(body)
	if err != nil {
		h.logger.Error("update fast enter config failed", zap.Error(err))
		return nil, err
	}
	out := fastEnterConfigFromModel(cfg)
	return &out, nil
}

func fastEnterConfigFromModel(cfg systemmod.FastEnterConfig) gen.SystemFastEnterConfig {
	return gen.SystemFastEnterConfig{
		Applications: fastEnterApplicationsToGen(cfg.Applications),
		QuickLinks:   fastEnterQuickLinksToGen(cfg.QuickLinks),
		MinWidth:     optIntValue(cfg.MinWidth),
	}
}

func fastEnterApplicationsFromGen(items []gen.SystemFastEnterApplication) []systemmod.FastEnterApplication {
	out := make([]systemmod.FastEnterApplication, 0, len(items))
	for _, item := range items {
		out = append(out, systemmod.FastEnterApplication{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Icon:        item.Icon,
			IconColor:   item.IconColor,
			Enabled:     item.Enabled,
			Order:       item.Order,
			RouteName:   optString(item.RouteName),
			Link:        optString(item.Link),
		})
	}
	return out
}

func fastEnterApplicationsToGen(items []systemmod.FastEnterApplication) []gen.SystemFastEnterApplication {
	out := make([]gen.SystemFastEnterApplication, 0, len(items))
	for _, item := range items {
		out = append(out, gen.SystemFastEnterApplication{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Icon:        item.Icon,
			IconColor:   item.IconColor,
			Enabled:     item.Enabled,
			Order:       item.Order,
			RouteName:   optStringValue(item.RouteName),
			Link:        optStringValue(item.Link),
		})
	}
	return out
}

func fastEnterQuickLinksFromGen(items []gen.SystemFastEnterQuickLink) []systemmod.FastEnterQuickLink {
	out := make([]systemmod.FastEnterQuickLink, 0, len(items))
	for _, item := range items {
		out = append(out, systemmod.FastEnterQuickLink{
			ID:        item.ID,
			Name:      item.Name,
			Enabled:   item.Enabled,
			Order:     item.Order,
			RouteName: optString(item.RouteName),
			Link:      optString(item.Link),
		})
	}
	return out
}

func fastEnterQuickLinksToGen(items []systemmod.FastEnterQuickLink) []gen.SystemFastEnterQuickLink {
	out := make([]gen.SystemFastEnterQuickLink, 0, len(items))
	for _, item := range items {
		out = append(out, gen.SystemFastEnterQuickLink{
			ID:        item.ID,
			Name:      item.Name,
			Enabled:   item.Enabled,
			Order:     item.Order,
			RouteName: optStringValue(item.RouteName),
			Link:      optStringValue(item.Link),
		})
	}
	return out
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
	records, err := mapJSON[[]gen.ViewPageItem](pages)
	if err != nil {
		return nil, err
	}
	out := &gen.ViewPagesResponse{
		Pages:       records,
		Refreshed:   refreshed,
		RefreshedAt: refreshedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	return out, nil
}

// ── user collaboration workspaces ───────────────────────────────────────────

func (h *APIHandler) GetUserCollaborationWorkspaces(ctx context.Context, params gen.GetUserCollaborationWorkspacesParams) (*gen.UserCollaborationWorkspacesResponse, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		return nil, err
	}
	if h.cwMemberRepo == nil {
		return &gen.UserCollaborationWorkspacesResponse{Records: []gen.UserCollaborationWorkspacesResponseRecordsItem{}}, nil
	}
	items, err := h.cwMemberRepo.GetCollaborationWorkspacesByUserID(params.ID)
	if err != nil {
		h.logger.Error("get user collaboration workspaces failed", zap.Error(err))
		return nil, err
	}
	out := &gen.UserCollaborationWorkspacesResponse{Records: make([]gen.UserCollaborationWorkspacesResponseRecordsItem, 0, len(items))}
	for _, item := range items {
		out.Records = append(out.Records, gen.UserCollaborationWorkspacesResponseRecordsItem{
			ID:     item.ID,
			Name:   item.Name,
			Status: item.Status,
		})
	}
	return out, nil
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
