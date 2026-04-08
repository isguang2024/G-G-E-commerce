// system.go — Phase 4 ogen handlers for the system app/space domain.
// Hooks into the already-injected appSvc and spaceSvc.
package handlers

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	appmod "github.com/gg-ecommerce/backend/internal/modules/system/app"
	spacemod "github.com/gg-ecommerce/backend/internal/modules/system/space"
)

// -------- apps --------

func (h *APIHandler) ListApps(ctx context.Context) (*gen.AnyListResponse, error) {
	items, err := h.appSvc.ListApps()
	if err != nil {
		h.logger.Error("list apps failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(items), Total: len(items)}, nil
}

func (h *APIHandler) SaveApp(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	var body appmod.SaveAppRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if _, err := h.appSvc.SaveApp(&body); err != nil {
		h.logger.Error("save app failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) ListAppHostBindings(ctx context.Context, params gen.ListAppHostBindingsParams) (*gen.AnyListResponse, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	items, err := h.appSvc.ListHostBindings(appKey)
	if err != nil {
		h.logger.Error("list app host bindings failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(items), Total: len(items)}, nil
}

func (h *APIHandler) SaveAppHostBinding(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	var body appmod.SaveHostBindingRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if _, err := h.appSvc.SaveHostBinding(body.AppKey, &body); err != nil {
		h.logger.Error("save app host binding failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeleteAppHostBinding(ctx context.Context, params gen.DeleteAppHostBindingParams) (*gen.MutationResult, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	if err := h.appSvc.DeleteHostBinding(appKey, params.ID); err != nil {
		h.logger.Error("delete app host binding failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) ListMenuSpaceEntryBindings(ctx context.Context, params gen.ListMenuSpaceEntryBindingsParams) (*gen.AnyListResponse, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	items, err := h.appSvc.ListMenuSpaceEntryBindings(appKey)
	if err != nil {
		h.logger.Error("list menu space entry bindings failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(items), Total: len(items)}, nil
}

func (h *APIHandler) SaveMenuSpaceEntryBinding(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	var body appmod.SaveMenuSpaceEntryBindingRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if _, err := h.appSvc.SaveMenuSpaceEntryBinding(body.AppKey, &body); err != nil {
		h.logger.Error("save menu space entry binding failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeleteMenuSpaceEntryBinding(ctx context.Context, params gen.DeleteMenuSpaceEntryBindingParams) (*gen.MutationResult, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	if err := h.appSvc.DeleteMenuSpaceEntryBinding(appKey, params.ID); err != nil {
		h.logger.Error("delete menu space entry binding failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetCurrentApp(ctx context.Context, params gen.GetCurrentAppParams) (gen.AnyObject, error) {
	resp, err := h.appSvc.GetCurrent("", optString(params.AppKey))
	if err != nil {
		h.logger.Error("get current app failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(resp), nil
}

// -------- spaces --------

func (h *APIHandler) ListMenuSpaces(ctx context.Context, params gen.ListMenuSpacesParams) (*gen.AnyListResponse, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	items, err := h.spaceSvc.ListSpaces(appKey)
	if err != nil {
		h.logger.Error("list menu spaces failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(items), Total: len(items)}, nil
}

func (h *APIHandler) SaveMenuSpace(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	var body spacemod.SaveSpaceRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if _, err := h.spaceSvc.SaveSpace(body.AppKey, &body); err != nil {
		h.logger.Error("save menu space failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetCurrentMenuSpace(ctx context.Context, params gen.GetCurrentMenuSpaceParams) (gen.AnyObject, error) {
	resp, err := h.spaceSvc.GetCurrent(optString(params.AppKey), "", optString(params.SpaceKey), nil, nil)
	if err != nil {
		h.logger.Error("get current menu space failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(resp), nil
}

func (h *APIHandler) GetMenuSpaceMode(ctx context.Context, params gen.GetMenuSpaceModeParams) (gen.AnyObject, error) {
	mode, err := h.spaceSvc.GetMode(optString(params.AppKey))
	if err != nil {
		h.logger.Error("get menu space mode failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(map[string]interface{}{"mode": mode}), nil
}

func (h *APIHandler) SaveMenuSpaceMode(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	var body struct {
		AppKey string `json:"app_key"`
		Mode   string `json:"mode"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if _, err := h.spaceSvc.SaveMode(body.AppKey, body.Mode); err != nil {
		h.logger.Error("save menu space mode failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) InitializeMenuSpaceFromDefault(ctx context.Context, params gen.InitializeMenuSpaceFromDefaultParams) (*gen.MutationResult, error) {
	var actor *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		actor = &uid
	}
	if _, err := h.spaceSvc.InitializeFromDefault("", params.SpaceKey, false, actor); err != nil {
		h.logger.Error("initialize menu space failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) ListMenuSpaceHostBindings(ctx context.Context) (*gen.AnyListResponse, error) {
	items, err := h.spaceSvc.ListHostBindings("")
	if err != nil {
		h.logger.Error("list menu space host bindings failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(items), Total: len(items)}, nil
}

func (h *APIHandler) SaveMenuSpaceHostBinding(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	var body spacemod.SaveHostBindingRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if _, err := h.spaceSvc.SaveHostBinding(body.AppKey, &body); err != nil {
		h.logger.Error("save menu space host binding failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}
