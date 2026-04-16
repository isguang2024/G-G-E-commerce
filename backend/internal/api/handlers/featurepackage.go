// featurepackage.go: ogen handler implementations for /feature-packages/*.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/modules/system/featurepackage"
	permissionrefresh "github.com/maben/backend/internal/pkg/permissionrefresh"
)

func (h *APIHandler) ListFeaturePackages(ctx context.Context, params gen.ListFeaturePackagesParams) (*gen.FeaturePackageList, error) {
	req := &dto.FeaturePackageListRequest{
		Current:     optInt(params.Current, 1),
		Size:        optInt(params.Size, 20),
		Keyword:     optString(params.Keyword),
		PackageType: optString(params.PackageType),
		Status:      optString(params.Status),
	}
	list, total, err := h.featurePkgSvc.List(req)
	if err != nil {
		h.logger.Error("list feature packages failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.FeaturePackageSummary, 0, len(list))
	for i := range list {
		p := &list[i]
		s := gen.FeaturePackageSummary{
			ID:         p.ID,
			PackageKey: p.PackageKey,
			Name:       p.Name,
			Status:     p.Status,
			IsBuiltin:  gen.NewOptBool(p.IsBuiltin),
			SortOrder:  gen.NewOptInt(p.SortOrder),
		}
		if p.PackageType != "" {
			s.PackageType = gen.NewOptNilString(p.PackageType)
		}
		if p.Description != "" {
			s.Description = gen.NewOptNilString(p.Description)
		}
		if p.ContextType != "" {
			s.ContextType = gen.NewOptNilString(p.ContextType)
		}
		records = append(records, s)
	}
	return &gen.FeaturePackageList{
		Records: records,
		Total:   total,
		Current: req.Current,
		Size:    req.Size,
	}, nil
}

func (h *APIHandler) ListFeaturePackageOptions(ctx context.Context, params gen.ListFeaturePackageOptionsParams) (*gen.FeaturePackageOptions, error) {
	req := &dto.FeaturePackageListRequest{
		PackageType: optString(params.PackageType),
		AppKey:      optString(params.AppKey),
	}
	list, err := h.featurePkgSvc.ListOptions(req)
	if err != nil {
		h.logger.Error("list feature package options failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageOptions{
		Records: featurePackageRefsFromModels(list),
		Total:   len(list),
	}, nil
}

func (h *APIHandler) GetFeaturePackage(ctx context.Context, params gen.GetFeaturePackageParams) (*gen.FeaturePackageSummary, error) {
	p, err := h.featurePkgSvc.Get(params.ID)
	if err != nil {
		if errors.Is(err, featurepackage.ErrFeaturePackageNotFound) {
			return nil, err
		}
		h.logger.Error("get feature package failed", zap.Error(err))
		return nil, err
	}
	s := gen.FeaturePackageSummary{
		ID:         p.ID,
		PackageKey: p.PackageKey,
		Name:       p.Name,
		Status:     p.Status,
		IsBuiltin:  gen.NewOptBool(p.IsBuiltin),
		SortOrder:  gen.NewOptInt(p.SortOrder),
	}
	if p.PackageType != "" {
		s.PackageType = gen.NewOptNilString(p.PackageType)
	}
	if p.Description != "" {
		s.Description = gen.NewOptNilString(p.Description)
	}
	if p.ContextType != "" {
		s.ContextType = gen.NewOptNilString(p.ContextType)
	}
	return &s, nil
}

func (h *APIHandler) CreateFeaturePackage(ctx context.Context, req *gen.FeaturePackageSaveRequest) (*gen.IDResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.FeaturePackageCreateRequest{
		PackageKey:  req.PackageKey,
		Name:        req.Name,
		Description: optString(req.Description),
		PackageType: optString(req.PackageType),
		Status:      optString(req.Status),
		SortOrder:   optInt(req.SortOrder, 0),
		AppKeys:     req.AppKeys,
	}
	created, err := h.featurePkgSvc.Create(dtoReq)
	if err != nil {
		h.logger.Error("create feature package failed", zap.Error(err))
		return nil, err
	}
	return &gen.IDResult{ID: created.ID}, nil
}

func (h *APIHandler) UpdateFeaturePackage(ctx context.Context, req *gen.FeaturePackageSaveRequest, params gen.UpdateFeaturePackageParams) (*gen.FeaturePackageMutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.FeaturePackageUpdateRequest{
		PackageKey:  req.PackageKey,
		Name:        req.Name,
		Description: optString(req.Description),
		PackageType: optString(req.PackageType),
		Status:      optString(req.Status),
		SortOrder:   optInt(req.SortOrder, 0),
		AppKeys:     req.AppKeys,
	}
	stats, err := h.featurePkgSvc.Update(params.ID, dtoReq)
	if err != nil {
		h.logger.Error("update feature package failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

func (h *APIHandler) DeleteFeaturePackage(ctx context.Context, params gen.DeleteFeaturePackageParams) (*gen.FeaturePackageMutationResult, error) {
	stats, err := h.featurePkgSvc.Delete(params.ID)
	if err != nil {
		h.logger.Error("delete feature package failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

func (h *APIHandler) GetFeaturePackageChildren(ctx context.Context, params gen.GetFeaturePackageChildrenParams) (*gen.FeaturePackageAssignmentResponse, error) {
	ids, pkgs, err := h.featurePkgSvc.GetPackageChildren(params.ID, "")
	if err != nil {
		h.logger.Error("get feature package children failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageAssignmentResponse{
		PackageIds: ids,
		Packages:   featurePackageRefsFromModels(pkgs),
	}, nil
}

func (h *APIHandler) SetFeaturePackageChildren(ctx context.Context, req *gen.UUIDListRequest, params gen.SetFeaturePackageChildrenParams) (*gen.FeaturePackageMutationResult, error) {
	stats, err := h.featurePkgSvc.SetPackageChildren(params.ID, uuidIDsFromRequest(req), "")
	if err != nil {
		h.logger.Error("set feature package children failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

func (h *APIHandler) GetFeaturePackageActions(ctx context.Context, params gen.GetFeaturePackageActionsParams) (*gen.FeaturePackageActionsResponse, error) {
	ids, _, err := h.featurePkgSvc.GetPackageKeys(params.ID, "")
	if err != nil {
		h.logger.Error("get feature package actions failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageActionsResponse{
		ActionIds: ids,
		Actions:   []gen.PermissionActionRef{},
	}, nil
}

func (h *APIHandler) SetFeaturePackageActions(ctx context.Context, req *gen.UUIDListRequest, params gen.SetFeaturePackageActionsParams) (*gen.FeaturePackageMutationResult, error) {
	stats, err := h.featurePkgSvc.SetPackageKeys(params.ID, uuidIDsFromRequest(req), "")
	if err != nil {
		h.logger.Error("set feature package actions failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

func (h *APIHandler) GetFeaturePackageMenus(ctx context.Context, params gen.GetFeaturePackageMenusParams) (*gen.FeaturePackageMenusResponse, error) {
	ids, menus, err := h.featurePkgSvc.GetPackageMenus(params.ID, "")
	if err != nil {
		h.logger.Error("get feature package menus failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageMenusResponse{
		MenuIds: ids,
		Menus:   featurePackageMenuItemsFromModels(menus),
	}, nil
}

func (h *APIHandler) SetFeaturePackageMenus(ctx context.Context, req *gen.UUIDListRequest, params gen.SetFeaturePackageMenusParams) (*gen.FeaturePackageMutationResult, error) {
	stats, err := h.featurePkgSvc.SetPackageMenus(params.ID, uuidIDsFromRequest(req), "")
	if err != nil {
		h.logger.Error("set feature package menus failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

func (h *APIHandler) GetFeaturePackageCollaborationWorkspaces(ctx context.Context, params gen.GetFeaturePackageCollaborationWorkspacesParams) (*gen.FeaturePackageCollaborationWorkspaceList, error) {
	ids, err := h.featurePkgSvc.GetPackageCollaborationWorkspaces(params.ID, "")
	if err != nil {
		h.logger.Error("get feature package cws failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.FeaturePackageCollaborationWorkspaceItem, 0, len(ids))
	for _, id := range ids {
		records = append(records, gen.FeaturePackageCollaborationWorkspaceItem{ID: id})
	}
	return &gen.FeaturePackageCollaborationWorkspaceList{Records: records, Total: int64(len(records))}, nil
}

func (h *APIHandler) SetFeaturePackageCollaborationWorkspaces(ctx context.Context, req *gen.UUIDListRequest, params gen.SetFeaturePackageCollaborationWorkspacesParams) (*gen.FeaturePackageMutationResult, error) {
	var grantedBy *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		grantedBy = &uid
	}
	stats, err := h.featurePkgSvc.SetPackageCollaborationWorkspaces(params.ID, uuidIDsFromRequest(req), grantedBy, "")
	if err != nil {
		h.logger.Error("set feature package cws failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

func (h *APIHandler) GetFeaturePackageImpactPreview(ctx context.Context, params gen.GetFeaturePackageImpactPreviewParams) (*gen.FeaturePackageImpactPreview, error) {
	preview, err := h.featurePkgSvc.GetImpactPreview(params.ID)
	if err != nil {
		h.logger.Error("get feature package impact preview failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageImpactPreview{
		PackageID:                   preview.PackageID,
		RoleCount:                   preview.RoleCount,
		CollaborationWorkspaceCount: preview.CollaborationWorkspaceCount,
		UserCount:                   preview.UserCount,
		MenuCount:                   preview.MenuCount,
		ActionCount:                 preview.ActionCount,
	}, nil
}

func (h *APIHandler) ListFeaturePackageVersions(ctx context.Context, params gen.ListFeaturePackageVersionsParams) (*gen.FeaturePackageVersionList, error) {
	list, total, err := h.featurePkgSvc.ListVersions(params.ID, optInt(params.Current, 1), optInt(params.Size, 20))
	if err != nil {
		h.logger.Error("list feature package versions failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.FeaturePackageVersionItem, 0, len(list))
	for i := range list {
		item := gen.FeaturePackageVersionItem{
			ID:         list[i].ID,
			PackageID:  list[i].PackageID,
			VersionNo:  list[i].VersionNo,
			ChangeType: list[i].ChangeType,
			Snapshot:   featurePackageSnapshotFromModel(list[i].Snapshot),
			RequestID:  list[i].RequestID,
			CreatedAt:  list[i].CreatedAt,
		}
		if list[i].OperatorID != nil {
			item.OperatorID = gen.NewOptNilUUID(*list[i].OperatorID)
		}
		records = append(records, item)
	}
	return &gen.FeaturePackageVersionList{Records: records, Total: total}, nil
}

func (h *APIHandler) ListFeaturePackageRiskAudits(ctx context.Context, params gen.ListFeaturePackageRiskAuditsParams) (*gen.FeaturePackageRiskAuditList, error) {
	list, total, err := h.featurePkgSvc.ListRiskAudits(params.ID, optInt(params.Current, 1), optInt(params.Size, 20))
	if err != nil {
		h.logger.Error("list feature package risk audits failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageRiskAuditList{Records: riskAuditItemsFromModels(list), Total: total}, nil
}

func featurePackageMutationResultFromStats(stats *permissionrefresh.RefreshStats) *gen.FeaturePackageMutationResult {
	if stats == nil {
		return &gen.FeaturePackageMutationResult{
			RefreshStats: gen.RefreshStats{},
		}
	}
	return &gen.FeaturePackageMutationResult{
		RefreshStats: gen.RefreshStats{
			RequestedPackageCount:       int64(stats.RequestedPackageCount),
			ImpactedPackageCount:        int64(stats.ImpactedPackageCount),
			RoleCount:                   int64(stats.RoleCount),
			CollaborationWorkspaceCount: int64(stats.CollaborationWorkspaceCount),
			UserCount:                   int64(stats.UserCount),
			ElapsedMilliseconds:         stats.ElapsedMilliseconds,
			FinishedAt:                  stats.FinishedAt,
		},
	}
}

