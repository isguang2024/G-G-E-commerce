// page.go: ogen handler implementations for /pages/* and runtime/sync.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/system/page"
)

func (h *APIHandler) ListPages(ctx context.Context, params gen.ListPagesParams) (*gen.AnyListResponse, error) {
	req := &page.ListRequest{
		Current:  optInt(params.Current, 1),
		Size:     optInt(params.Size, 20),
		Keyword:  optString(params.Keyword),
		AppKey:   params.AppKey,
		SpaceKey: optString(params.SpaceKey),
		Status:   optString(params.Status),
	}
	list, total, err := h.pageSvc.List(req)
	if err != nil {
		h.logger.Error("list pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(list), Total: int(total)}, nil
}

func (h *APIHandler) ListPageOptions(ctx context.Context, params gen.ListPageOptionsParams) (*gen.AnyListResponse, error) {
	list, err := h.pageSvc.ListOptions(params.AppKey, optString(params.SpaceKey))
	if err != nil {
		h.logger.Error("list page options failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) ListPageMenuOptions(ctx context.Context, params gen.ListPageMenuOptionsParams) (*gen.AnyListResponse, error) {
	list, err := h.pageSvc.ListMenuOptions(params.AppKey, optString(params.SpaceKey))
	if err != nil {
		h.logger.Error("list page menu options failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) ListRuntimePages(ctx context.Context, params gen.ListRuntimePagesParams) (*gen.AnyListResponse, error) {
	var userID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		userID = &uid
	}
	cwID, _ := collaborationWorkspaceIDFromContext(ctx)
	list, err := h.pageSvc.ListRuntime(
		optString(params.AppKey),
		requestHostFromCtx(ctx),
		optString(params.SpaceKey),
		userID,
		cwID,
	)
	if err != nil {
		h.logger.Error("list runtime pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) ListPublicRuntimePages(ctx context.Context, params gen.ListPublicRuntimePagesParams) (*gen.AnyListResponse, error) {
	list, err := h.pageSvc.ListRuntimePublic(
		optString(params.AppKey),
		requestHostFromCtx(ctx),
		optString(params.SpaceKey),
		nil,
		nil,
	)
	if err != nil {
		h.logger.Error("list public runtime pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) ListUnregisteredPages(ctx context.Context, params gen.ListUnregisteredPagesParams) (*gen.AnyListResponse, error) {
	list, err := h.pageSvc.ListUnregistered(params.AppKey)
	if err != nil {
		h.logger.Error("list unregistered pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) SyncPages(ctx context.Context, params gen.SyncPagesParams) (*gen.PageSyncResult, error) {
	result, err := h.pageSvc.Sync(params.AppKey)
	if err != nil {
		h.logger.Error("sync pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.PageSyncResult{
		CreatedCount: result.CreatedCount,
		SkippedCount: result.SkippedCount,
		CreatedKeys:  result.CreatedKeys,
	}, nil
}

func (h *APIHandler) PreviewPageBreadcrumb(ctx context.Context, params gen.PreviewPageBreadcrumbParams) (*gen.AnyListResponse, error) {
	list, err := h.pageSvc.PreviewBreadcrumb(params.ID, params.AppKey)
	if err != nil {
		h.logger.Error("preview page breadcrumb failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) GetPage(ctx context.Context, params gen.GetPageParams) (gen.AnyObject, error) {
	rec, err := h.pageSvc.Get(params.ID, params.AppKey)
	if err != nil {
		h.logger.Error("get page failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(rec), nil
}

func (h *APIHandler) CreatePage(ctx context.Context, req *gen.PageSaveRequest, params gen.CreatePageParams) (gen.AnyObject, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	saveReq := pageSaveRequestFromGen(req, params.AppKey)
	rec, err := h.pageSvc.Create(saveReq)
	if err != nil {
		h.logger.Error("create page failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(rec), nil
}

func (h *APIHandler) UpdatePage(ctx context.Context, req *gen.PageSaveRequest, params gen.UpdatePageParams) (gen.AnyObject, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	saveReq := pageSaveRequestFromGen(req, params.AppKey)
	rec, err := h.pageSvc.Update(params.ID, saveReq)
	if err != nil {
		h.logger.Error("update page failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(rec), nil
}

func (h *APIHandler) DeletePage(ctx context.Context, params gen.DeletePageParams) (*gen.MutationResult, error) {
	if err := h.pageSvc.Delete(params.ID, params.AppKey); err != nil {
		h.logger.Error("delete page failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func pageSaveRequestFromGen(req *gen.PageSaveRequest, appKey string) *page.SaveRequest {
	saveReq := &page.SaveRequest{
		AppKey:          appKey,
		PageKey:         optString(req.PageKey),
		Name:            optString(req.Name),
		RouteName:       optString(req.RouteName),
		RoutePath:       optString(req.RoutePath),
		Component:       optString(req.Component),
		SpaceKeys:       req.SpaceKeys,
		PageType:        optString(req.PageType),
		VisibilityScope: optString(req.VisibilityScope),
		Source:          optString(req.Source),
		ModuleKey:       optString(req.ModuleKey),
		SortOrder:       optInt(req.SortOrder, 0),
		ParentMenuID:    optString(req.ParentMenuID),
		ParentPageKey:   optString(req.ParentPageKey),
		DisplayGroupKey: optString(req.DisplayGroupKey),
		ActiveMenuPath:  optString(req.ActiveMenuPath),
		BreadcrumbMode:  optString(req.BreadcrumbMode),
		AccessMode:      optString(req.AccessMode),
		PermissionKey:   optString(req.PermissionKey),
		Status:          optString(req.Status),
	}
	if req.InheritPermission.Set {
		value := req.InheritPermission.Value
		saveReq.InheritPermission = &value
	}
	if req.KeepAlive.Set {
		value := req.KeepAlive.Value
		saveReq.KeepAlive = &value
	}
	if req.IsFullPage.Set {
		value := req.IsFullPage.Value
		saveReq.IsFullPage = &value
	}
	if req.Meta.Set {
		meta := map[string]interface{}{}
		_ = unmarshalAnyObject(req.Meta.Value, &meta)
		saveReq.Meta = meta
	}
	return saveReq
}
