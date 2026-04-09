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

func (h *APIHandler) ListApps(ctx context.Context) (*gen.SystemAppListResponse, error) {
	items, err := h.appSvc.ListApps()
	if err != nil {
		h.logger.Error("list apps failed", zap.Error(err))
		return nil, err
	}
	return &gen.SystemAppListResponse{
		Records: systemAppItemsFromModels(items),
		Total:   int64(len(items)),
	}, nil
}

func (h *APIHandler) SaveApp(ctx context.Context, req *gen.SystemAppSaveRequest) (*gen.SystemAppItem, error) {
	if req == nil {
		req = &gen.SystemAppSaveRequest{}
	}
	body := appmod.SaveAppRequest{
		AppKey:          req.AppKey,
		Name:            req.Name,
		Description:     optString(req.Description),
		SpaceMode:       optString(req.SpaceMode),
		DefaultSpaceKey: optString(req.DefaultSpaceKey),
		AuthMode:        optString(req.AuthMode),
		Status:          optString(req.Status),
		IsDefault:       optBool(req.IsDefault),
		Meta:            optSystemMetaToMap(req.Meta),
	}
	record, err := h.appSvc.SaveApp(&body)
	if err != nil {
		h.logger.Error("save app failed", zap.Error(err))
		return nil, err
	}
	return systemAppItemFromModel(record)
}

func (h *APIHandler) ListAppHostBindings(ctx context.Context, params gen.ListAppHostBindingsParams) (*gen.SystemAppHostBindingListResponse, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	items, err := h.appSvc.ListHostBindings(appKey)
	if err != nil {
		h.logger.Error("list app host bindings failed", zap.Error(err))
		return nil, err
	}
	return &gen.SystemAppHostBindingListResponse{
		Records: systemAppHostBindingItemsFromModels(items),
		Total:   int64(len(items)),
	}, nil
}

func (h *APIHandler) SaveAppHostBinding(ctx context.Context, req *gen.SystemAppHostBindingSaveRequest) (*gen.SystemAppHostBindingItem, error) {
	if req == nil {
		req = &gen.SystemAppHostBindingSaveRequest{}
	}
	body := appmod.SaveHostBindingRequest{
		ID:              optString(req.ID),
		AppKey:          req.AppKey,
		MatchType:       optString(req.MatchType),
		Host:            req.Host,
		PathPattern:     optString(req.PathPattern),
		Priority:        optInt(req.Priority, 0),
		Description:     optString(req.Description),
		IsPrimary:       optBool(req.IsPrimary),
		DefaultSpaceKey: optString(req.DefaultSpaceKey),
		Status:          optString(req.Status),
		Meta:            optSystemMetaToMap(req.Meta),
	}
	record, err := h.appSvc.SaveHostBinding(body.AppKey, &body)
	if err != nil {
		h.logger.Error("save app host binding failed", zap.Error(err))
		return nil, err
	}
	return systemAppHostBindingItemFromModel(record)
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

func (h *APIHandler) ListMenuSpaceEntryBindings(ctx context.Context, params gen.ListMenuSpaceEntryBindingsParams) (*gen.SystemMenuSpaceEntryBindingListResponse, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	items, err := h.appSvc.ListMenuSpaceEntryBindings(appKey)
	if err != nil {
		h.logger.Error("list menu space entry bindings failed", zap.Error(err))
		return nil, err
	}
	return &gen.SystemMenuSpaceEntryBindingListResponse{
		Records: systemMenuSpaceEntryBindingItemsFromModels(items),
		Total:   int64(len(items)),
	}, nil
}

func (h *APIHandler) SaveMenuSpaceEntryBinding(ctx context.Context, req *gen.SystemMenuSpaceEntryBindingSaveRequest) (*gen.SystemMenuSpaceEntryBindingItem, error) {
	if req == nil {
		req = &gen.SystemMenuSpaceEntryBindingSaveRequest{}
	}
	body := appmod.SaveMenuSpaceEntryBindingRequest{
		ID:          optString(req.ID),
		AppKey:      req.AppKey,
		SpaceKey:    req.SpaceKey,
		MatchType:   optString(req.MatchType),
		Host:        req.Host,
		PathPattern: optString(req.PathPattern),
		Priority:    optInt(req.Priority, 0),
		Description: optString(req.Description),
		IsPrimary:   optBool(req.IsPrimary),
		Status:      optString(req.Status),
		Meta:        optSystemMetaToMap(req.Meta),
	}
	record, err := h.appSvc.SaveMenuSpaceEntryBinding(body.AppKey, &body)
	if err != nil {
		h.logger.Error("save menu space entry binding failed", zap.Error(err))
		return nil, err
	}
	return systemMenuSpaceEntryBindingItemFromModel(record)
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

func (h *APIHandler) GetCurrentApp(ctx context.Context, params gen.GetCurrentAppParams) (*gen.SystemCurrentAppResponse, error) {
	resp, err := h.appSvc.GetCurrent(requestHostFromCtx(ctx), optString(params.AppKey))
	if err != nil {
		h.logger.Error("get current app failed", zap.Error(err))
		return nil, err
	}
	return systemCurrentAppResponseFromModel(resp)
}

// -------- spaces --------

func (h *APIHandler) ListMenuSpaces(ctx context.Context, params gen.ListMenuSpacesParams) (*gen.SystemMenuSpaceListResponse, error) {
	appKey := ""
	if v, ok := params.AppKey.Get(); ok {
		appKey = v
	}
	items, err := h.spaceSvc.ListSpaces(appKey)
	if err != nil {
		h.logger.Error("list menu spaces failed", zap.Error(err))
		return nil, err
	}
	return &gen.SystemMenuSpaceListResponse{
		Records: systemMenuSpaceItemsFromModels(items),
		Total:   int64(len(items)),
	}, nil
}

func (h *APIHandler) SaveMenuSpace(ctx context.Context, req *gen.SystemMenuSpaceSaveRequest) (*gen.SystemMenuSpaceItem, error) {
	if req == nil {
		req = &gen.SystemMenuSpaceSaveRequest{}
	}
	body := spacemod.SaveSpaceRequest{
		AppKey:           req.AppKey,
		SpaceKey:         req.SpaceKey,
		Name:             req.Name,
		Description:      optString(req.Description),
		DefaultHomePath:  optString(req.DefaultHomePath),
		IsDefault:        optBool(req.IsDefault),
		Status:           optString(req.Status),
		AccessMode:       optString(req.AccessMode),
		AllowedRoleCodes: req.AllowedRoleCodes,
		Meta:             optSystemMetaToMap(req.Meta),
	}
	record, err := h.spaceSvc.SaveSpace(body.AppKey, &body)
	if err != nil {
		h.logger.Error("save menu space failed", zap.Error(err))
		return nil, err
	}
	return systemMenuSpaceItemFromModel(record)
}

func (h *APIHandler) GetCurrentMenuSpace(ctx context.Context, params gen.GetCurrentMenuSpaceParams) (*gen.SystemCurrentMenuSpaceResponse, error) {
	var userID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		userID = &uid
	}
	cwID, _ := collaborationWorkspaceIDFromContext(ctx)
	resp, err := h.spaceSvc.GetCurrent(
		optString(params.AppKey),
		requestHostFromCtx(ctx),
		optString(params.SpaceKey),
		userID,
		cwID,
	)
	if err != nil {
		h.logger.Error("get current menu space failed", zap.Error(err))
		return nil, err
	}
	return systemCurrentMenuSpaceResponseFromModel(resp)
}

func (h *APIHandler) GetMenuSpaceMode(ctx context.Context, params gen.GetMenuSpaceModeParams) (*gen.SystemMenuSpaceModeResponse, error) {
	mode, err := h.spaceSvc.GetMode(optString(params.AppKey))
	if err != nil {
		h.logger.Error("get menu space mode failed", zap.Error(err))
		return nil, err
	}
	return &gen.SystemMenuSpaceModeResponse{Mode: mode}, nil
}

func (h *APIHandler) SaveMenuSpaceMode(ctx context.Context, req *gen.SystemMenuSpaceModeSaveRequest) (*gen.SystemMenuSpaceModeResponse, error) {
	if req == nil {
		req = &gen.SystemMenuSpaceModeSaveRequest{}
	}
	mode, err := h.spaceSvc.SaveMode(req.AppKey, req.Mode)
	if err != nil {
		h.logger.Error("save menu space mode failed", zap.Error(err))
		return nil, err
	}
	return &gen.SystemMenuSpaceModeResponse{Mode: mode}, nil
}

func (h *APIHandler) InitializeMenuSpaceFromDefault(ctx context.Context, params gen.InitializeMenuSpaceFromDefaultParams) (*gen.SystemMenuSpaceInitializeResult, error) {
	var actor *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		actor = &uid
	}
	result, err := h.spaceSvc.InitializeFromDefault("", params.SpaceKey, false, actor)
	if err != nil {
		h.logger.Error("initialize menu space failed", zap.Error(err))
		return nil, err
	}
	return systemMenuSpaceInitializeResultFromModel(result)
}

func (h *APIHandler) ListMenuSpaceHostBindings(ctx context.Context) (*gen.SystemMenuSpaceHostBindingListResponse, error) {
	items, err := h.spaceSvc.ListHostBindings("")
	if err != nil {
		h.logger.Error("list menu space host bindings failed", zap.Error(err))
		return nil, err
	}
	return &gen.SystemMenuSpaceHostBindingListResponse{
		Records: systemMenuSpaceHostBindingItemsFromModels(items),
		Total:   int64(len(items)),
	}, nil
}

func (h *APIHandler) SaveMenuSpaceHostBinding(ctx context.Context, req *gen.SystemMenuSpaceHostBindingSaveRequest) (*gen.SystemMenuSpaceHostBindingItem, error) {
	if req == nil {
		req = &gen.SystemMenuSpaceHostBindingSaveRequest{}
	}
	body := spacemod.SaveHostBindingRequest{
		AppKey:      req.AppKey,
		Host:        req.Host,
		SpaceKey:    req.SpaceKey,
		Description: optString(req.Description),
		IsDefault:   optBool(req.IsDefault),
		Status:      optString(req.Status),
		Meta:        optSystemMetaToMap(req.Meta),
	}
	record, err := h.spaceSvc.SaveHostBinding(body.AppKey, &body)
	if err != nil {
		h.logger.Error("save menu space host binding failed", zap.Error(err))
		return nil, err
	}
	return systemMenuSpaceHostBindingItemFromModel(record)
}

func systemAppItemFromModel(item *appmod.AppRecord) (*gen.SystemAppItem, error) {
	if item == nil {
		return nil, nil
	}
	out, err := mapJSON[gen.SystemAppItem](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemAppItemsFromModels(items []appmod.AppRecord) []gen.SystemAppItem {
	out := make([]gen.SystemAppItem, 0, len(items))
	for i := range items {
		if item, err := systemAppItemFromModel(&items[i]); err == nil && item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func systemAppHostBindingItemFromModel(item *appmod.HostBindingRecord) (*gen.SystemAppHostBindingItem, error) {
	if item == nil {
		return nil, nil
	}
	out, err := mapJSON[gen.SystemAppHostBindingItem](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemAppHostBindingItemsFromModels(items []appmod.HostBindingRecord) []gen.SystemAppHostBindingItem {
	out := make([]gen.SystemAppHostBindingItem, 0, len(items))
	for i := range items {
		if item, err := systemAppHostBindingItemFromModel(&items[i]); err == nil && item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func systemMenuSpaceEntryBindingItemFromModel(item *appmod.MenuSpaceEntryBindingRecord) (*gen.SystemMenuSpaceEntryBindingItem, error) {
	if item == nil {
		return nil, nil
	}
	out, err := mapJSON[gen.SystemMenuSpaceEntryBindingItem](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemMenuSpaceEntryBindingItemsFromModels(items []appmod.MenuSpaceEntryBindingRecord) []gen.SystemMenuSpaceEntryBindingItem {
	out := make([]gen.SystemMenuSpaceEntryBindingItem, 0, len(items))
	for i := range items {
		if item, err := systemMenuSpaceEntryBindingItemFromModel(&items[i]); err == nil && item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func systemCurrentAppResponseFromModel(item *appmod.CurrentResponse) (*gen.SystemCurrentAppResponse, error) {
	if item == nil {
		return &gen.SystemCurrentAppResponse{}, nil
	}
	out, err := mapJSON[gen.SystemCurrentAppResponse](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemMenuSpaceItemFromModel(item *spacemod.SpaceRecord) (*gen.SystemMenuSpaceItem, error) {
	if item == nil {
		return nil, nil
	}
	out, err := mapJSON[gen.SystemMenuSpaceItem](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemMenuSpaceItemsFromModels(items []spacemod.SpaceRecord) []gen.SystemMenuSpaceItem {
	out := make([]gen.SystemMenuSpaceItem, 0, len(items))
	for i := range items {
		if item, err := systemMenuSpaceItemFromModel(&items[i]); err == nil && item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func systemCurrentMenuSpaceResponseFromModel(item *spacemod.CurrentResponse) (*gen.SystemCurrentMenuSpaceResponse, error) {
	if item == nil {
		return &gen.SystemCurrentMenuSpaceResponse{}, nil
	}
	out, err := mapJSON[gen.SystemCurrentMenuSpaceResponse](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemMenuSpaceInitializeResultFromModel(item *spacemod.InitializeResult) (*gen.SystemMenuSpaceInitializeResult, error) {
	if item == nil {
		return &gen.SystemMenuSpaceInitializeResult{}, nil
	}
	out, err := mapJSON[gen.SystemMenuSpaceInitializeResult](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemMenuSpaceHostBindingItemFromModel(item *spacemod.HostBindingRecord) (*gen.SystemMenuSpaceHostBindingItem, error) {
	if item == nil {
		return nil, nil
	}
	out, err := mapJSON[gen.SystemMenuSpaceHostBindingItem](item)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func systemMenuSpaceHostBindingItemsFromModels(items []spacemod.HostBindingRecord) []gen.SystemMenuSpaceHostBindingItem {
	out := make([]gen.SystemMenuSpaceHostBindingItem, 0, len(items))
	for i := range items {
		if item, err := systemMenuSpaceHostBindingItemFromModel(&items[i]); err == nil && item != nil {
			out = append(out, *item)
		}
	}
	return out
}
