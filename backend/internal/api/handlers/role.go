// role.go: ogen Handler implementations for the /roles/* OpenAPI surface.
// Phase 4 — role domain migration. Mirrors the legacy gin handler in
// internal/modules/system/role/handler.go but speaks the typed gen schemas.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/modules/system/role"
	"github.com/maben/backend/internal/modules/system/user"
)

const roleTimeLayout = "2006-01-02 15:04:05"

func (h *APIHandler) ListRoles(ctx context.Context, params gen.ListRolesParams) (*gen.RoleList, error) {
	req := &dto.RoleListRequest{
		Current:  optInt(params.Current, 1),
		Size:     optInt(params.Size, 20),
		RoleName: optString(params.RoleName),
		RoleCode: optString(params.RoleCode),
		AppKey:   optString(params.AppKey),
	}
	list, total, err := h.roleSvc.List(req)
	if err != nil {
		h.logger.Error("list roles failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.RoleSummary, 0, len(list))
	for i := range list {
		records = append(records, roleSummaryFromModel(&list[i]))
	}
	return &gen.RoleList{
		Records: records,
		Total:   int(total),
		Current: req.Current,
		Size:    req.Size,
	}, nil
}

func (h *APIHandler) ListRoleOptions(ctx context.Context, params gen.ListRoleOptionsParams) (*gen.RoleOptions, error) {
	list, err := h.roleSvc.ListOptions()
	if err != nil {
		h.logger.Error("list role options failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.RoleSummary, 0, len(list))
	for i := range list {
		records = append(records, roleSummaryFromModel(&list[i]))
	}
	return &gen.RoleOptions{Records: records, Total: len(records)}, nil
}

func (h *APIHandler) GetRole(ctx context.Context, params gen.GetRoleParams) (gen.GetRoleRes, error) {
	r, err := h.roleSvc.Get(params.ID)
	if err != nil {
		if errors.Is(err, role.ErrRoleNotFound) {
			return &gen.Error{Code: 404, Message: "角色不存在"}, nil
		}
		h.logger.Error("get role failed", zap.Error(err))
		return nil, err
	}
	s := roleSummaryFromModel(r)
	return &s, nil
}

func (h *APIHandler) CreateRole(ctx context.Context, req *gen.RoleCreateRequest) (*gen.RoleCreateResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.RoleCreateRequest{
		Code:        req.Code,
		Name:        req.Name,
		Description: optString(req.Description),
		AppKeys:     req.AppKeys,
		SortOrder:   optInt(req.SortOrder, 0),
		Status:      optString(req.Status),
	}
	created, err := h.roleSvc.Create(dtoReq)
	if err != nil {
		h.logger.Error("create role failed", zap.Error(err))
		return nil, err
	}
	return &gen.RoleCreateResult{RoleId: created.ID}, nil
}

func (h *APIHandler) UpdateRole(ctx context.Context, req *gen.RoleUpdateRequest, params gen.UpdateRoleParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.RoleUpdateRequest{
		Code:        optString(req.Code),
		Name:        optString(req.Name),
		Description: optString(req.Description),
		AppKeys:     req.AppKeys,
		SortOrder:   optInt(req.SortOrder, 0),
		Status:      optString(req.Status),
	}
	if err := h.roleSvc.Update(params.ID, dtoReq); err != nil {
		h.logger.Error("update role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeleteRole(ctx context.Context, params gen.DeleteRoleParams) (*gen.MutationResult, error) {
	if err := h.roleSvc.Delete(params.ID); err != nil {
		h.logger.Error("delete role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetRolePackages(ctx context.Context, params gen.GetRolePackagesParams) (*gen.RolePackagesResponse, error) {
	ids, pkgs, err := h.roleSvc.GetRolePackages(params.ID, params.AppKey)
	if err != nil {
		h.logger.Error("get role packages failed", zap.Error(err))
		return nil, err
	}
	return &gen.RolePackagesResponse{
		PackageIds: ids,
		Packages:   featurePackageRefsFromModels(pkgs),
	}, nil
}

func (h *APIHandler) SetRolePackages(ctx context.Context, req *gen.UUIDListRequest, params gen.SetRolePackagesParams) (*gen.MutationResult, error) {
	var grantedBy *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		grantedBy = &uid
	}
	ids := uuidIDsFromRequest(req)
	if err := h.roleSvc.SetRolePackages(params.ID, ids, grantedBy, params.AppKey); err != nil {
		h.logger.Error("set role packages failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetRoleMenus(ctx context.Context, params gen.GetRoleMenusParams) (*gen.RoleMenusResponse, error) {
	boundary, err := h.roleSvc.GetRoleMenuBoundary(params.ID, params.AppKey)
	if err != nil {
		h.logger.Error("get role menus failed", zap.Error(err))
		return nil, err
	}
	return &gen.RoleMenusResponse{
		MenuIds:            boundary.EffectiveMenuIDs,
		AvailableMenuIds:   boundary.AvailableMenuIDs,
		HiddenMenuIds:      boundary.HiddenMenuIDs,
		ExpandedPackageIds: boundary.ExpandedPackageIDs,
		DerivedSources:     menuSourceEntriesFromMap(boundary.MenuSourceMap),
	}, nil
}

func (h *APIHandler) SetRoleMenus(ctx context.Context, req *gen.UUIDListRequest, params gen.SetRoleMenusParams) (*gen.MutationResult, error) {
	ids := uuidIDsFromRequest(req)
	if err := h.roleSvc.SetRoleMenus(params.ID, ids, params.AppKey); err != nil {
		h.logger.Error("set role menus failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetRoleActions(ctx context.Context, params gen.GetRoleActionsParams) (*gen.RoleActionsResponse, error) {
	boundary, err := h.roleSvc.GetRoleKeyBoundary(params.ID, params.AppKey)
	if err != nil {
		h.logger.Error("get role actions failed", zap.Error(err))
		return nil, err
	}
	return &gen.RoleActionsResponse{
		ActionIds:          boundary.EffectiveKeyIDs,
		AvailableActionIds: boundary.AvailableKeyIDs,
		DisabledActionIds:  boundary.DisabledKeyIDs,
		ExpandedPackageIds: boundary.ExpandedPackageIDs,
		Actions:            []gen.PermissionActionRef{},
		DerivedSources:     actionSourceEntriesFromMap(boundary.KeySourceMap),
	}, nil
}

func (h *APIHandler) SetRoleActions(ctx context.Context, req *gen.UUIDListRequest, params gen.SetRoleActionsParams) (*gen.MutationResult, error) {
	ids := uuidIDsFromRequest(req)
	keys := make([]user.RoleKeyPermission, 0, len(ids))
	for _, kid := range ids {
		keys = append(keys, user.RoleKeyPermission{RoleID: params.ID, KeyID: kid})
	}
	if err := h.roleSvc.SetRoleKeys(params.ID, keys, params.AppKey); err != nil {
		h.logger.Error("set role actions failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetRoleDataPermissions(ctx context.Context, params gen.GetRoleDataPermissionsParams) (*gen.RoleDataPermissionsResponse, error) {
	perms, _, scopeOpts, err := h.roleSvc.GetRoleDataPermissions(params.ID)
	if err != nil {
		h.logger.Error("get role data permissions failed", zap.Error(err))
		return nil, err
	}
	items := make([]gen.DataPermissionItem, 0, len(perms))
	for _, p := range perms {
		items = append(items, gen.DataPermissionItem{
			ResourceCode: p.ResourceCode,
			DataScope:    p.DataScope,
		})
	}
	scopes := make([]gen.DataScopeOption, 0, len(scopeOpts))
	for _, s := range scopeOpts {
		scopes = append(scopes, gen.DataScopeOption{
			DataScope: s.Code,
			Label:     s.Name,
		})
	}
	return &gen.RoleDataPermissionsResponse{
		Permissions: items,
		Resources:   []gen.DataResourceOption{},
		DataScopes:  scopes,
	}, nil
}

func (h *APIHandler) SetRoleDataPermissions(ctx context.Context, req *gen.RoleDataPermissionsRequest, params gen.SetRoleDataPermissionsParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	perms := make([]user.RoleDataPermission, 0, len(req.Permissions))
	for _, p := range req.Permissions {
		perms = append(perms, user.RoleDataPermission{
			RoleID:       params.ID,
			ResourceCode: p.ResourceCode,
			DataScope:    p.DataScope,
		})
	}
	if err := h.roleSvc.SetRoleDataPermissions(params.ID, perms); err != nil {
		h.logger.Error("set role data permissions failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ── helpers ────────────────────────────────────────────────────────────────

func roleSummaryFromModel(r *user.Role) gen.RoleSummary {
	out := gen.RoleSummary{
		RoleId:     r.ID,
		RoleName:   r.Name,
		RoleCode:   r.Code,
		Status:     r.Status,
		SortOrder:  gen.NewOptInt(r.SortOrder),
		IsGlobal:   gen.NewOptBool(len(r.AppKeys) == 0),
		CreateTime: r.CreatedAt.Format(roleTimeLayout),
		CanEditPermission: gen.NewOptBool(true),
	}
	if r.Description != "" {
		out.Description = gen.NewOptNilString(r.Description)
	}
	if len(r.AppKeys) > 0 {
		var v gen.OptNilStringArray
		v.SetTo(r.AppKeys)
		out.AppKeys = v
	}
	return out
}

func featurePackageRefsFromModels(pkgs []user.FeaturePackage) []gen.FeaturePackageRef {
	out := make([]gen.FeaturePackageRef, 0, len(pkgs))
	for i := range pkgs {
		p := &pkgs[i]
		ref := gen.FeaturePackageRef{
			ID:         p.ID,
			PackageKey: p.PackageKey,
			Name:       p.Name,
			Status:     p.Status,
			IsBuiltin:  gen.NewOptBool(p.IsBuiltin),
			SortOrder:  gen.NewOptInt(p.SortOrder),
		}
		if p.PackageType != "" {
			ref.PackageType = gen.NewOptNilString(p.PackageType)
		}
		if p.Description != "" {
			ref.Description = gen.NewOptNilString(p.Description)
		}
		if p.ContextType != "" {
			ref.ContextType = gen.NewOptNilString(p.ContextType)
		}
		out = append(out, ref)
	}
	return out
}

func menuSourceEntriesFromMap(m map[uuid.UUID][]uuid.UUID) []gen.MenuSourceEntry {
	out := make([]gen.MenuSourceEntry, 0, len(m))
	for menuID, pkgs := range m {
		out = append(out, gen.MenuSourceEntry{MenuID: menuID, PackageIds: pkgs})
	}
	return out
}

func actionSourceEntriesFromMap(m map[uuid.UUID][]uuid.UUID) []gen.ActionSourceEntry {
	out := make([]gen.ActionSourceEntry, 0, len(m))
	for actionID, pkgs := range m {
		out = append(out, gen.ActionSourceEntry{ActionID: actionID, PackageIds: pkgs})
	}
	return out
}

func uuidIDsFromRequest(req *gen.UUIDListRequest) []uuid.UUID {
	if req == nil {
		return nil
	}
	return req.Ids
}


