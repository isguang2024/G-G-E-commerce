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

// ─── GetCurrentCollaborationBoundaryRolePackages ─────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationBoundaryRolePackages(ctx context.Context, params gen.GetCurrentCollaborationBoundaryRolePackagesParams) (*gen.CollaborationBoundaryRolePackagesResponse, error) {
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
	return &gen.CollaborationBoundaryRolePackagesResponse{
		PackageIds: snapshot.PackageIDs,
		Packages:   featurePackageRefsFromModels(pkgs),
		Inherited:  snapshot.Inherited,
	}, nil
}

// ─── SetCurrentCollaborationBoundaryRolePackages ─────────────────────

func (h *cwAPIHandler) SetCurrentCollaborationBoundaryRolePackages(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationBoundaryRolePackagesParams) (*gen.MutationResult, error) {
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

// ─── GetCurrentCollaborationBoundaryRoleMenus ────────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationBoundaryRoleMenus(ctx context.Context, params gen.GetCurrentCollaborationBoundaryRoleMenusParams) (*gen.RoleMenusResponse, error) {
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

// ─── SetCurrentCollaborationBoundaryRoleMenus ────────────────────────

func (h *cwAPIHandler) SetCurrentCollaborationBoundaryRoleMenus(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationBoundaryRoleMenusParams) (*gen.MutationResult, error) {
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

// ─── GetCurrentCollaborationBoundaryRoleActions ──────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationBoundaryRoleActions(ctx context.Context, params gen.GetCurrentCollaborationBoundaryRoleActionsParams) (*gen.RoleActionsResponse, error) {
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

// ─── SetCurrentCollaborationBoundaryRoleActions ──────────────────────

func (h *cwAPIHandler) SetCurrentCollaborationBoundaryRoleActions(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationBoundaryRoleActionsParams) (*gen.MutationResult, error) {
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

// ─── GetCurrentCollaborationBoundaryPackages ─────────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationBoundaryPackages(ctx context.Context) (*gen.FeaturePackageAssignmentResponse, error) {
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

// ─── GetCurrentCollaborationMenus ────────────────────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationMenus(ctx context.Context) (*gen.CollaborationMenusResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationMenusResponse{MenuIds: []uuid.UUID{}}, nil
	}
	snapshot, err := h.boundarySvc.GetMenuSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw menus failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationMenusResponse{MenuIds: snapshot.EffectiveIDs}, nil
}

// ─── GetCurrentCollaborationMenuOrigins ──────────────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationMenuOrigins(ctx context.Context) (*gen.CollaborationMenuOriginsResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationMenuOriginsResponse{
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
	return &gen.CollaborationMenuOriginsResponse{
		DerivedMenuIds: snapshot.DerivedIDs,
		DerivedSources: menuSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedMenuIds: snapshot.BlockedIDs,
	}, nil
}

// ─── GetCurrentCollaborationActions ──────────────────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationActions(ctx context.Context) (*gen.CollaborationActionsResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationActionsResponse{
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
	return &gen.CollaborationActionsResponse{
		ActionIds: snapshot.EffectiveIDs,
		Actions:   permissionActionRefsFromModels(actions),
	}, nil
}

// ─── GetCurrentCollaborationActionOrigins ────────────────────────────

func (h *cwAPIHandler) GetCurrentCollaborationActionOrigins(ctx context.Context) (*gen.CollaborationActionOriginsResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationActionOriginsResponse{
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
	return &gen.CollaborationActionOriginsResponse{
		DerivedActionIds: snapshot.DerivedIDs,
		DerivedSources:   actionSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedActionIds: snapshot.BlockedIDs,
	}, nil
}

// ─── GetCollaborationMenus ──────────────────────────────────────────

func (h *cwAPIHandler) GetCollaborationMenus(ctx context.Context, params gen.GetCollaborationMenusParams) (*gen.CollaborationMenusResponse, error) {
	snapshot, err := h.boundarySvc.GetMenuSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw menus failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationMenusResponse{MenuIds: snapshot.EffectiveIDs}, nil
}

// ─── GetCollaborationMenuOrigins ─────────────────────────────────────

func (h *cwAPIHandler) GetCollaborationMenuOrigins(ctx context.Context, params gen.GetCollaborationMenuOriginsParams) (*gen.CollaborationMenuOriginsResponse, error) {
	snapshot, err := h.boundarySvc.GetMenuSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw menu origins failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationMenuOriginsResponse{
		DerivedMenuIds: snapshot.DerivedIDs,
		DerivedSources: menuSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedMenuIds: snapshot.BlockedIDs,
	}, nil
}

// ─── SetCollaborationMenus ──────────────────────────────────────────

func (h *cwAPIHandler) SetCollaborationMenus(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCollaborationMenusParams) (*gen.MutationResult, error) {
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

// ─── GetCollaborationActions ─────────────────────────────────────────

func (h *cwAPIHandler) GetCollaborationActions(ctx context.Context, params gen.GetCollaborationActionsParams) (*gen.CollaborationActionsResponse, error) {
	snapshot, err := h.boundarySvc.GetSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw actions failed", zap.Error(err))
		return nil, err
	}
	actions, err := h.keyRepo.GetByIDs(snapshot.EffectiveIDs)
	if err != nil {
		return nil, err
	}
	return &gen.CollaborationActionsResponse{
		ActionIds: snapshot.EffectiveIDs,
		Actions:   permissionActionRefsFromModels(actions),
	}, nil
}

// ─── GetCollaborationActionOrigins ───────────────────────────────────

func (h *cwAPIHandler) GetCollaborationActionOrigins(ctx context.Context, params gen.GetCollaborationActionOriginsParams) (*gen.CollaborationActionOriginsResponse, error) {
	snapshot, err := h.boundarySvc.GetSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw action origins failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationActionOriginsResponse{
		DerivedActionIds: snapshot.DerivedIDs,
		DerivedSources:   actionSourceEntriesFromMap(snapshot.DerivedMap),
		BlockedActionIds: snapshot.BlockedIDs,
	}, nil
}

// ─── SetCollaborationActions ─────────────────────────────────────────

func (h *cwAPIHandler) SetCollaborationActions(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCollaborationActionsParams) (*gen.MutationResult, error) {
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


