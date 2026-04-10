// permission.go: ogen handler implementations for /permission-actions/*.
package handlers

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/dto"
)

func (h *APIHandler) ListPermissionActions(ctx context.Context, params gen.ListPermissionActionsParams) (*gen.PermissionActionList, error) {
	req := &dto.PermissionKeyListRequest{
		Current:       optInt(params.Current, 1),
		Size:          optInt(params.Size, 20),
		Keyword:       optString(params.Keyword),
		ModuleGroupID: optString(params.GroupID),
		Status:        optString(params.Status),
	}
	list, total, summary, err := h.permSvc.List(req)
	if err != nil {
		h.logger.Error("list permission actions failed", zap.Error(err))
		return nil, err
	}
	return &gen.PermissionActionList{
		Records:      permissionActionListItemsFromModels(list),
		Total:        int64(total),
		Current:      req.Current,
		Size:         req.Size,
		AuditSummary: permissionActionAuditSummaryFromModel(summary),
	}, nil
}

func (h *APIHandler) ListPermissionActionOptions(ctx context.Context, params gen.ListPermissionActionOptionsParams) (*gen.PermissionActionOptions, error) {
	req := &dto.PermissionKeyListRequest{
		Keyword:       optString(params.Keyword),
		ModuleGroupID: optString(params.GroupID),
	}
	list, err := h.permSvc.ListOptions(req)
	if err != nil {
		h.logger.Error("list permission action options failed", zap.Error(err))
		return nil, err
	}
	return &gen.PermissionActionOptions{Records: permissionActionOptionItemsFromModels(list), Total: int64(len(list))}, nil
}

func (h *APIHandler) GetPermissionAction(ctx context.Context, params gen.GetPermissionActionParams) (*gen.PermissionActionDetail, error) {
	p, err := h.permSvc.Get(params.ID)
	if err != nil {
		h.logger.Error("get permission action failed", zap.Error(err))
		return nil, err
	}
	detail := gen.PermissionActionDetail{
		ID:        p.ID,
		ActionKey: p.PermissionKey,
		Name:      p.Name,
		Status:    p.Status,
	}
	if p.Description != "" {
		detail.Description = gen.NewOptNilString(p.Description)
	}
	return &detail, nil
}

func (h *APIHandler) CreatePermissionAction(ctx context.Context, req *gen.PermissionActionSaveRequest) (*gen.IDResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.PermissionKeyCreateRequest{
		PermissionKey: req.ActionKey,
		Name:          req.Name,
		Description:   optString(req.Description),
		Status:        optString(req.Status),
	}
	created, err := h.permSvc.Create(dtoReq)
	if err != nil {
		h.logger.Error("create permission action failed", zap.Error(err))
		return nil, err
	}
	return &gen.IDResult{ID: created.ID}, nil
}

func (h *APIHandler) UpdatePermissionAction(ctx context.Context, req *gen.PermissionActionSaveRequest, params gen.UpdatePermissionActionParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.PermissionKeyUpdateRequest{
		PermissionKey: req.ActionKey,
		Name:          req.Name,
		Description:   optString(req.Description),
		Status:        optString(req.Status),
	}
	if err := h.permSvc.Update(params.ID, dtoReq); err != nil {
		h.logger.Error("update permission action failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeletePermissionAction(ctx context.Context, params gen.DeletePermissionActionParams) (*gen.MutationResult, error) {
	if err := h.permSvc.Delete(params.ID); err != nil {
		h.logger.Error("delete permission action failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) ListPermissionActionEndpoints(ctx context.Context, params gen.ListPermissionActionEndpointsParams) (*gen.PermissionActionEndpointList, error) {
	list, err := h.permSvc.ListEndpoints(params.ID)
	if err != nil {
		h.logger.Error("list permission action endpoints failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.PermissionActionEndpointListItem, 0, len(list))
	for _, item := range list {
		records = append(records, permissionActionEndpointListItemFromModel(item))
	}
	return &gen.PermissionActionEndpointList{Records: records, Total: int64(len(list))}, nil
}

func (h *APIHandler) AddPermissionActionEndpoint(ctx context.Context, req *gen.PermissionActionEndpointAddRequest, params gen.AddPermissionActionEndpointParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	if err := h.permSvc.AddEndpoint(params.ID, req.EndpointCode); err != nil {
		h.logger.Error("add permission action endpoint failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) RemovePermissionActionEndpoint(ctx context.Context, params gen.RemovePermissionActionEndpointParams) (*gen.MutationResult, error) {
	if err := h.permSvc.RemoveEndpoint(params.ID, params.EndpointCode); err != nil {
		h.logger.Error("remove permission action endpoint failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetPermissionActionConsumers(ctx context.Context, params gen.GetPermissionActionConsumersParams) (*gen.PermissionActionConsumersResponse, error) {
	details, err := h.permSvc.GetConsumerDetails(params.ID)
	if err != nil {
		h.logger.Error("get permission action consumers failed", zap.Error(err))
		return nil, err
	}
	return permissionActionConsumersResponseFromModel(details), nil
}

func (h *APIHandler) GetPermissionActionImpactPreview(ctx context.Context, params gen.GetPermissionActionImpactPreviewParams) (*gen.PermissionActionImpactPreview, error) {
	preview, err := h.permSvc.GetImpactPreview(params.ID)
	if err != nil {
		h.logger.Error("get permission action impact preview failed", zap.Error(err))
		return nil, err
	}
	return permissionActionImpactPreviewFromModel(preview), nil
}

func (h *APIHandler) ListPermissionActionGroups(ctx context.Context) (*gen.PermissionActionGroupList, error) {
	list, total, err := h.permSvc.ListGroups(&dto.PermissionGroupListRequest{Current: 1, Size: 500})
	if err != nil {
		h.logger.Error("list permission action groups failed", zap.Error(err))
		return nil, err
	}
	return &gen.PermissionActionGroupList{Records: permissionGroupItemsFromModels(list), Total: int64(total)}, nil
}

func (h *APIHandler) CleanupUnusedPermissionActions(ctx context.Context) (*gen.PermissionActionCleanupResult, error) {
	result, err := h.permSvc.CleanupUnused()
	if err != nil {
		h.logger.Error("cleanup unused permission actions failed", zap.Error(err))
		return nil, err
	}
	return &gen.PermissionActionCleanupResult{
		DeletedCount: int64(result.DeletedCount),
		DeletedKeys:  result.DeletedKeys,
	}, nil
}

func (h *APIHandler) ListPermissionActionRiskAudits(ctx context.Context, params gen.ListPermissionActionRiskAuditsParams) (*gen.PermissionActionRiskAuditList, error) {
	list, total, err := h.permSvc.ListRiskAudits("", optInt(params.Current, 1), optInt(params.Size, 20))
	if err != nil {
		h.logger.Error("list permission action risk audits failed", zap.Error(err))
		return nil, err
	}
	return &gen.PermissionActionRiskAuditList{Records: riskAuditItemsFromModels(list), Total: int64(total)}, nil
}

func (h *APIHandler) ListPermissionActionBatchTemplates(ctx context.Context) (*gen.PermissionActionBatchTemplateList, error) {
	list, err := h.permSvc.ListBatchTemplates()
	if err != nil {
		h.logger.Error("list permission action batch templates failed", zap.Error(err))
		return nil, err
	}
	return &gen.PermissionActionBatchTemplateList{Records: permissionBatchTemplateItemsFromModels(list), Total: int64(len(list))}, nil
}
