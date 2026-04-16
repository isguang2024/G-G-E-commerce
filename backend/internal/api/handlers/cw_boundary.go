// cw_boundary.go — ogen handler implementations for CW boundary role grid
// (packages, menus, actions) and workspace-level menus/actions.
package handlers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/appscope"
)

// ─── GetCurrentCollaborationWorkspaceBoundaryRolePackages ─────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryRolePackages(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceBoundaryRolePackagesParams) (*gen.CollaborationWorkspaceBoundaryRolePackagesResponse, error) {
	member, role, err := h.resolveCWRole(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	inheritAll := role.CollaborationWorkspaceID == nil
	snapshot, err := h.boundarySvc.GetRoleSnapshot(member.CollaborationWorkspaceID, role.ID, inheritAll)
	if err != nil {
		h.logger.Error("get cw role packages failed", zap.Error(err))
		return nil, err
	}
	pkgs, err := h.featurePkgRepo.GetByIDs(snapshot.PackageIDs)
	if err != nil {
		return nil, err
	}
	return &gen.CollaborationWorkspaceBoundaryRolePackagesResponse{
		PackageIds: snapshot.PackageIDs,
		Packages:   featurePackageRefsFromModels(pkgs),
		Inherited:  snapshot.Inherited,
	}, nil
}

// ─── SetCurrentCollaborationWorkspaceBoundaryRolePackages ─────────────────────

func (h *APIHandler) SetCurrentCollaborationWorkspaceBoundaryRolePackages(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationWorkspaceBoundaryRolePackagesParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	packageIDs := uuidIDsFromRequest(req)

	cwPkgIDs, err := h.cwFeaturePkgRepo.GetPackageIDsByCollaborationWorkspaceID(member.CollaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	allowedSet := make(map[uuid.UUID]struct{}, len(cwPkgIDs))
	for _, id := range cwPkgIDs {
		allowedSet[id] = struct{}{}
	}
	for _, pkgID := range packageIDs {
		if _, ok := allowedSet[pkgID]; !ok {
			return nil, errText("存在未向当前协作空间开通的功能包")
		}
	}

	userID, _ := userIDFromContext(ctx)
	if err := appscope.ReplaceRolePackagesInApp(h.db, role.ID, "", packageIDs, &userID); err != nil {
		h.logger.Error("set cw role packages failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── GetCurrentCollaborationWorkspaceBoundaryRoleMenus ────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryRoleMenus(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceBoundaryRoleMenusParams) (*gen.RoleMenusResponse, error) {
	member, role, err := h.resolveCWRole(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	inheritAll := role.CollaborationWorkspaceID == nil
	snapshot, err := h.boundarySvc.GetRoleSnapshot(member.CollaborationWorkspaceID, role.ID, inheritAll)
	if err != nil {
		h.logger.Error("get cw role menus failed", zap.Error(err))
		return nil, err
	}
	return &gen.RoleMenusResponse{
		MenuIds:            snapshot.MenuIDs,
		AvailableMenuIds:   snapshot.AvailableMenuIDs,
		HiddenMenuIds:      snapshot.HiddenMenuIDs,
		ExpandedPackageIds: snapshot.ExpandedPackageIDs,
		DerivedSources:     menuSourceEntriesFromMap(snapshot.MenuSourceMap),
	}, nil
}

// ─── SetCurrentCollaborationWorkspaceBoundaryRoleMenus ────────────────────────

func (h *APIHandler) SetCurrentCollaborationWorkspaceBoundaryRoleMenus(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationWorkspaceBoundaryRoleMenusParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	menuIDs := uuidIDsFromRequest(req)

	inheritAll := role.CollaborationWorkspaceID == nil
	snapshot, err := h.boundarySvc.GetRoleSnapshot(member.CollaborationWorkspaceID, role.ID, inheritAll)
	if err != nil {
		h.logger.Error("get cw role menu boundary failed", zap.Error(err))
		return nil, err
	}
	enabledSet := uuidSetFromSlice(snapshot.AvailableMenuIDs)
	for _, menuID := range menuIDs {
		if !enabledSet[menuID] {
			return nil, errText("存在超出当前角色已绑定功能包范围的菜单")
		}
	}
	hiddenMenuIDs := excludeUUIDs(snapshot.AvailableMenuIDs, menuIDs)
	if err := appscope.ReplaceRoleHiddenMenusInApp(h.db, role.ID, "", hiddenMenuIDs); err != nil {
		h.logger.Error("set cw role hidden menus failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── GetCurrentCollaborationWorkspaceBoundaryRoleActions ──────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryRoleActions(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceBoundaryRoleActionsParams) (*gen.RoleActionsResponse, error) {
	member, role, err := h.resolveCWRole(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	inheritAll := role.CollaborationWorkspaceID == nil
	snapshot, err := h.boundarySvc.GetRoleSnapshot(member.CollaborationWorkspaceID, role.ID, inheritAll)
	if err != nil {
		h.logger.Error("get cw role actions failed", zap.Error(err))
		return nil, err
	}
	actions, err := h.keyRepo.GetByIDs(snapshot.AvailableActionIDs)
	if err != nil {
		return nil, err
	}
	return &gen.RoleActionsResponse{
		ActionIds:          snapshot.ActionIDs,
		AvailableActionIds: snapshot.AvailableActionIDs,
		DisabledActionIds:  snapshot.DisabledActionIDs,
		ExpandedPackageIds: snapshot.ExpandedPackageIDs,
		Actions:            permissionActionRefsFromModels(actions),
		DerivedSources:     actionSourceEntriesFromMap(snapshot.ActionSourceMap),
	}, nil
}

// ─── SetCurrentCollaborationWorkspaceBoundaryRoleActions ──────────────────────

func (h *APIHandler) SetCurrentCollaborationWorkspaceBoundaryRoleActions(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationWorkspaceBoundaryRoleActionsParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	actionIDs := uuidIDsFromRequest(req)

	inheritAll := role.CollaborationWorkspaceID == nil
	snapshot, err := h.boundarySvc.GetRoleSnapshot(member.CollaborationWorkspaceID, role.ID, inheritAll)
	if err != nil {
		h.logger.Error("get cw role action boundary failed", zap.Error(err))
		return nil, err
	}
	enabledSet := uuidSetFromSlice(snapshot.AvailableActionIDs)
	for _, actionID := range actionIDs {
		if !enabledSet[actionID] {
			return nil, errText("存在超出当前角色已绑定功能包范围的功能权限")
		}
	}
	disabledIDs := excludeUUIDs(snapshot.AvailableActionIDs, actionIDs)
	if err := appscope.ReplaceRoleDisabledActionsInScope(h.db, role.ID, snapshot.AvailableActionIDs, disabledIDs); err != nil {
		h.logger.Error("set cw role disabled actions failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── GetCurrentCollaborationWorkspaceBoundaryPackages ─────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryPackages(ctx context.Context) (*gen.FeaturePackageAssignmentResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.FeaturePackageAssignmentResponse{
			PackageIds: []uuid.UUID{},
			Packages:   []gen.FeaturePackageRef{},
		}, nil
	}
	packageIDs, err := appscope.PackageIDsByCollaborationWorkspace(h.db, member.CollaborationWorkspaceID, "")
	if err != nil {
		h.logger.Error("get cw boundary packages failed", zap.Error(err))
		return nil, err
	}
	if len(packageIDs) == 0 {
		return &gen.FeaturePackageAssignmentResponse{
			PackageIds: []uuid.UUID{},
			Packages:   []gen.FeaturePackageRef{},
		}, nil
	}
	pkgs, err := h.featurePkgRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, err
	}
	filtered := make([]user.FeaturePackage, 0, len(pkgs))
	for _, p := range pkgs {
		if strings.TrimSpace(p.Status) != "" && p.Status != "normal" {
			continue
		}
		if p.ContextType != "" && p.ContextType != "collaboration" && p.ContextType != "common" {
			continue
		}
		filtered = append(filtered, p)
	}
	filteredIDs := make([]uuid.UUID, 0, len(filtered))
	for _, item := range filtered {
		filteredIDs = append(filteredIDs, item.ID)
	}
	return &gen.FeaturePackageAssignmentResponse{
		PackageIds: filteredIDs,
		Packages:   featurePackageRefsFromModels(filtered),
	}, nil
}

// ─── GetCurrentCollaborationWorkspaceMenus ────────────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceMenus(ctx context.Context) (*gen.CollaborationWorkspaceMenusResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationWorkspaceMenusResponse{MenuIds: []uuid.UUID{}}, nil
	}
	snapshot, err := h.boundarySvc.GetMenuSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw menus failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceMenusResponse{MenuIds: snapshot.EffectiveIDs}, nil
}

// ─── GetCurrentCollaborationWorkspaceMenuOrigins ──────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceMenuOrigins(ctx context.Context) (*gen.CollaborationWorkspaceMenuOriginsResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationWorkspaceMenuOriginsResponse{
			DerivedMenuIds: []uuid.UUID{},
			DerivedSources: []gen.MenuSourceEntry{},
			BlockedMenuIds: []uuid.UUID{},
		}, nil
	}
	snapshot, err := h.boundarySvc.GetMenuSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw menu origins failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceMenuOriginsResponse{
		DerivedMenuIds: snapshot.DerivedIDs,
		DerivedSources: menuSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedMenuIds: snapshot.BlockedIDs,
	}, nil
}

// ─── GetCurrentCollaborationWorkspaceActions ──────────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceActions(ctx context.Context) (*gen.CollaborationWorkspaceActionsResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationWorkspaceActionsResponse{
			ActionIds: []uuid.UUID{},
			Actions:   []gen.PermissionActionRef{},
		}, nil
	}
	snapshot, err := h.boundarySvc.GetSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw actions failed", zap.Error(err))
		return nil, err
	}
	actions, err := h.keyRepo.GetByIDs(snapshot.EffectiveIDs)
	if err != nil {
		return nil, err
	}
	return &gen.CollaborationWorkspaceActionsResponse{
		ActionIds: snapshot.EffectiveIDs,
		Actions:   permissionActionRefsFromModels(actions),
	}, nil
}

// ─── GetCurrentCollaborationWorkspaceActionOrigins ────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceActionOrigins(ctx context.Context) (*gen.CollaborationWorkspaceActionOriginsResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationWorkspaceActionOriginsResponse{
			DerivedActionIds: []uuid.UUID{},
			DerivedSources:   []gen.ActionSourceEntry{},
			BlockedActionIds: []uuid.UUID{},
		}, nil
	}
	snapshot, err := h.boundarySvc.GetSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw action origins failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceActionOriginsResponse{
		DerivedActionIds: snapshot.DerivedIDs,
		DerivedSources:   actionSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedActionIds: snapshot.BlockedIDs,
	}, nil
}

// ─── GetCollaborationWorkspaceMenus ──────────────────────────────────────────

func (h *APIHandler) GetCollaborationWorkspaceMenus(ctx context.Context, params gen.GetCollaborationWorkspaceMenusParams) (*gen.CollaborationWorkspaceMenusResponse, error) {
	snapshot, err := h.boundarySvc.GetMenuSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw menus failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceMenusResponse{MenuIds: snapshot.EffectiveIDs}, nil
}

// ─── GetCollaborationWorkspaceMenuOrigins ─────────────────────────────────────

func (h *APIHandler) GetCollaborationWorkspaceMenuOrigins(ctx context.Context, params gen.GetCollaborationWorkspaceMenuOriginsParams) (*gen.CollaborationWorkspaceMenuOriginsResponse, error) {
	snapshot, err := h.boundarySvc.GetMenuSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw menu origins failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceMenuOriginsResponse{
		DerivedMenuIds: snapshot.DerivedIDs,
		DerivedSources: menuSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedMenuIds: snapshot.BlockedIDs,
	}, nil
}

// ─── SetCollaborationWorkspaceMenus ──────────────────────────────────────────

func (h *APIHandler) SetCollaborationWorkspaceMenus(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCollaborationWorkspaceMenusParams) (*gen.MutationResult, error) {
	menuIDs := uuidIDsFromRequest(req)
	snapshot, err := h.boundarySvc.GetMenuSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw menu snapshot failed", zap.Error(err))
		return nil, err
	}
	blockedIDs := excludeUUIDs(snapshot.DerivedIDs, menuIDs)
	if err := appscope.ReplaceCollaborationWorkspaceBlockedMenusInApp(h.db, params.ID, "", blockedIDs); err != nil {
		h.logger.Error("set cw blocked menus failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── GetCollaborationWorkspaceActions ─────────────────────────────────────────

func (h *APIHandler) GetCollaborationWorkspaceActions(ctx context.Context, params gen.GetCollaborationWorkspaceActionsParams) (*gen.CollaborationWorkspaceActionsResponse, error) {
	snapshot, err := h.boundarySvc.GetSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw actions failed", zap.Error(err))
		return nil, err
	}
	actions, err := h.keyRepo.GetByIDs(snapshot.EffectiveIDs)
	if err != nil {
		return nil, err
	}
	return &gen.CollaborationWorkspaceActionsResponse{
		ActionIds: snapshot.EffectiveIDs,
		Actions:   permissionActionRefsFromModels(actions),
	}, nil
}

// ─── GetCollaborationWorkspaceActionOrigins ───────────────────────────────────

func (h *APIHandler) GetCollaborationWorkspaceActionOrigins(ctx context.Context, params gen.GetCollaborationWorkspaceActionOriginsParams) (*gen.CollaborationWorkspaceActionOriginsResponse, error) {
	snapshot, err := h.boundarySvc.GetSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw action origins failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceActionOriginsResponse{
		DerivedActionIds: snapshot.DerivedIDs,
		DerivedSources:   actionSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedActionIds: snapshot.BlockedIDs,
	}, nil
}

// ─── SetCollaborationWorkspaceActions ─────────────────────────────────────────

func (h *APIHandler) SetCollaborationWorkspaceActions(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCollaborationWorkspaceActionsParams) (*gen.MutationResult, error) {
	actionIDs := uuidIDsFromRequest(req)
	snapshot, err := h.boundarySvc.GetSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw action snapshot failed", zap.Error(err))
		return nil, err
	}
	blockedIDs := excludeUUIDs(snapshot.DerivedIDs, actionIDs)
	if err := appscope.ReplaceCollaborationWorkspaceBlockedActionsInScope(h.db, params.ID, snapshot.DerivedIDs, blockedIDs); err != nil {
		h.logger.Error("set cw blocked actions failed", zap.Error(err))
		return nil, err
	}
	if _, err := h.boundarySvc.RefreshSnapshot(params.ID); err != nil {
		h.logger.Error("refresh cw snapshot after set actions failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

