// extras.go — Phase 4 ogen quick-win handlers for already-migrated domains
// (feature package, permission, menu, page).
package handlers

import (
	"context"
	"errors"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/page"
	"github.com/gg-ecommerce/backend/internal/modules/system/permission"
)

// -------- feature package --------

func (h *APIHandler) RollbackFeaturePackage(ctx context.Context, req *gen.RollbackRequest, params gen.RollbackFeaturePackageParams) (*gen.FeaturePackageMutationResult, error) {
	if req == nil {
		return featurePackageMutationResultFromStats(nil), nil
	}
	var operatorID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		operatorID = &uid
	}
	stats, err := h.featurePkgSvc.Rollback(params.ID, req.VersionID, operatorID, "")
	if err != nil {
		h.logger.Error("rollback feature package failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

func (h *APIHandler) GetFeaturePackageRelationTree(ctx context.Context, params gen.GetFeaturePackageRelationTreeParams) (*gen.FeaturePackageRelationTree, error) {
	tree, err := h.featurePkgSvc.GetRelationTree(optString(params.WorkspaceScope), optString(params.Keyword))
	if err != nil {
		h.logger.Error("get feature package relation tree failed", zap.Error(err))
		return nil, err
	}
	out := &gen.FeaturePackageRelationTree{
		CycleDependencies: tree.CycleDependencies,
		IsolatedBaseKeys:  tree.IsolatedBaseKeys,
		Roots:             convertRelationNodesFromAny(tree.Roots),
	}
	if out.CycleDependencies == nil {
		out.CycleDependencies = [][]string{}
	}
	if out.IsolatedBaseKeys == nil {
		out.IsolatedBaseKeys = []string{}
	}
	return out, nil
}

func convertRelationNodesFromAny(src interface{}) []gen.FeaturePackageRelationNode {
	b, err := json.Marshal(src)
	if err != nil {
		return []gen.FeaturePackageRelationNode{}
	}
	var raw []map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return []gen.FeaturePackageRelationNode{}
	}
	return buildRelationNodes(raw)
}

func buildRelationNodes(raw []map[string]interface{}) []gen.FeaturePackageRelationNode {
	out := make([]gen.FeaturePackageRelationNode, 0, len(raw))
	for _, m := range raw {
		node := gen.FeaturePackageRelationNode{
			Children: []gen.FeaturePackageRelationNode{},
		}
		if v, ok := m["id"].(string); ok {
			if id, err := uuid.Parse(v); err == nil {
				node.ID = id
			}
		}
		if v, ok := m["package_key"].(string); ok {
			node.PackageKey = v
		}
		if v, ok := m["name"].(string); ok {
			node.Name = v
		}
		if v, ok := m["status"].(string); ok {
			node.Status = v
		}
		if v, ok := m["package_type"].(string); ok && v != "" {
			node.PackageType = gen.NewOptNilString(v)
		}
		if v, ok := m["workspace_scope"].(string); ok && v != "" {
			node.WorkspaceScope = gen.NewOptNilString(v)
		}
		if v, ok := m["reference_count"].(float64); ok {
			node.ReferenceCount = gen.NewOptInt(int(v))
		}
		if c, ok := m["children"].([]interface{}); ok {
			childRaw := make([]map[string]interface{}, 0, len(c))
			for _, ci := range c {
				if cm, ok := ci.(map[string]interface{}); ok {
					childRaw = append(childRaw, cm)
				}
			}
			node.Children = buildRelationNodes(childRaw)
		}
		out = append(out, node)
	}
	return out
}

// -------- permission --------

func (h *APIHandler) BatchUpdatePermissionActions(ctx context.Context, req *gen.PermissionActionBatchUpdateRequest) (*gen.PermissionActionBatchUpdateResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	body := permission.PermissionBatchUpdateRequest{
		IDs:          req.Ids,
		TemplateName: optString(req.TemplateName),
	}
	if req.Status.Set {
		value := req.Status.Value
		body.Status = &value
	}
	if req.ModuleGroupID.Set {
		value := req.ModuleGroupID.Value.String()
		body.ModuleGroupID = &value
	}
	if req.FeatureGroupID.Set {
		value := req.FeatureGroupID.Value.String()
		body.FeatureGroupID = &value
	}
	var operatorID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		operatorID = &uid
	}
	result, err := h.permSvc.BatchUpdate(&body, operatorID, "")
	if err != nil {
		h.logger.Error("batch update permission actions failed", zap.Error(err))
		return nil, err
	}
	return &gen.PermissionActionBatchUpdateResult{
		UpdatedCount: int64(result.UpdatedCount),
		SkippedIds:   result.SkippedIDs,
	}, nil
}

func (h *APIHandler) CreatePermissionActionGroup(ctx context.Context, req *gen.PermissionActionGroupSaveRequest) (*gen.IDResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	body := dto.PermissionGroupSaveRequest{
		Code:        req.Code,
		Name:        req.Name,
		NameEn:      optString(req.NameEn),
		Description: optString(req.Description),
		GroupType:   req.GroupType,
		Status:      optString(req.Status),
		SortOrder:   optInt(req.SortOrder, 0),
	}
	item, err := h.permSvc.CreateGroup(&body)
	if err != nil {
		h.logger.Error("create permission group failed", zap.Error(err))
		return nil, err
	}
	return &gen.IDResult{ID: item.ID}, nil
}

func (h *APIHandler) UpdatePermissionActionGroup(ctx context.Context, req *gen.PermissionActionGroupSaveRequest, params gen.UpdatePermissionActionGroupParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	body := dto.PermissionGroupSaveRequest{
		Code:        req.Code,
		Name:        req.Name,
		NameEn:      optString(req.NameEn),
		Description: optString(req.Description),
		GroupType:   req.GroupType,
		Status:      optString(req.Status),
		SortOrder:   optInt(req.SortOrder, 0),
	}
	if err := h.permSvc.UpdateGroup(params.ID, &body); err != nil {
		h.logger.Error("update permission group failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeletePermissionActionGroup(ctx context.Context, params gen.DeletePermissionActionGroupParams) (*gen.MutationResult, error) {
	if err := h.permSvc.DeleteGroup(params.ID); err != nil {
		h.logger.Error("delete permission group failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) SavePermissionActionBatchTemplate(ctx context.Context, req *gen.PermissionActionBatchTemplateSaveRequest) (*gen.PermissionActionBatchTemplateItem, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	body := permission.PermissionBatchTemplateSaveRequest{
		Name:        req.Name,
		Description: optString(req.Description),
	}
	if req.Payload.Set {
		body.Payload = permissionBatchTemplatePayloadToMap(req.Payload.Value)
	}
	var operatorID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		operatorID = &uid
	}
	item, err := h.permSvc.SaveBatchTemplate(&body, operatorID)
	if err != nil {
		h.logger.Error("save permission batch template failed", zap.Error(err))
		return nil, err
	}
	out := permissionBatchTemplateItemFromModel(*item)
	return &out, nil
}

// -------- menu --------

func (h *APIHandler) GetMenuDeletePreview(ctx context.Context, params gen.GetMenuDeletePreviewParams) (*gen.MenuDeletePreviewResponse, error) {
	preview, err := h.menuSvc.DeletePreview(params.ID, "", nil)
	if err != nil {
		h.logger.Error("get menu delete preview failed", zap.Error(err))
		return nil, err
	}
	return &gen.MenuDeletePreviewResponse{
		Mode:                  preview.Mode,
		MenuCount:             preview.MenuCount,
		ChildCount:            preview.ChildCount,
		AffectedPageCount:     preview.AffectedPageCount,
		AffectedRelationCount: preview.AffectedRelationCount,
	}, nil
}

// -------- page --------

func (h *APIHandler) GetPageAccessTrace(ctx context.Context, params gen.GetPageAccessTraceParams) (*gen.PageAccessTraceResponse, error) {
	req := &page.AccessTraceRequest{
		AppKey: strings.TrimSpace(params.AppKey),
		UserID: params.UserID.String(),
	}
	if params.CollaborationWorkspaceID.Set {
		req.CollaborationWorkspaceID = params.CollaborationWorkspaceID.Value.String()
	}
	if params.SpaceKey.Set {
		req.SpaceKey = params.SpaceKey.Value
	}
	if params.PageKey.Set {
		req.PageKey = params.PageKey.Value
	}
	if params.PageKeys.Set {
		req.PageKeys = params.PageKeys.Value
	}
	if params.RoutePath.Set {
		req.RoutePath = params.RoutePath.Value
	}
	result, err := h.pageSvc.GetAccessTrace(req.AppKey, req)
	if err != nil {
		h.logger.Error("get page access trace failed", zap.Error(err))
		return nil, err
	}
	out, err := mapJSON[gen.PageAccessTraceResponse](result)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
