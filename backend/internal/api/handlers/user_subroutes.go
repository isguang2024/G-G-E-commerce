// user_subroutes.go: ogen handler implementations for user sub-routes.
// Phase 4: GetUserMenus, SetUserMenus, GetUserPackages, SetUserPackages,
// GetUserPermissions, GetUserPermissionDiagnosis.
package handlers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/appscope"
	"github.com/maben/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/maben/backend/internal/pkg/platformaccess"
	"github.com/maben/backend/internal/pkg/workspacerolebinding"
)

// normalizeAppKeyStr mirrors appctx.NormalizeAppKey without the gin dependency.
func normalizeAppKeyStr(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return models.DefaultAppKey
	}
	return target
}

// resolveAppKey extracts and normalises the app_key from an OptString param.
func resolveAppKey(opt gen.OptString) string {
	if opt.Set {
		return normalizeAppKeyStr(opt.Value)
	}
	return models.DefaultAppKey
}

// ── GetUserMenus ─────────────────────────────────────────────────────────────

func (h *APIHandler) GetUserMenus(ctx context.Context, params gen.GetUserMenusParams) (*gen.UserMenusResponse, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		h.logger.Error("get user for menus failed", zap.Error(err))
		return nil, err
	}
	appKey := resolveAppKey(params.AppKey)
	snapshot, err := h.getPersonalSnapshotForUser(params.ID, appKey)
	if err != nil {
		h.logger.Error("get personal workspace snapshot for menus failed", zap.Error(err))
		return nil, err
	}

	menuIDs := snapshot.MenuIDs
	if menuIDs == nil {
		menuIDs = []uuid.UUID{}
	}
	availableMenuIDs := snapshot.AvailableMenuIDs
	if availableMenuIDs == nil {
		availableMenuIDs = []uuid.UUID{}
	}
	hiddenMenuIDs := snapshot.HiddenMenuIDs
	if hiddenMenuIDs == nil {
		hiddenMenuIDs = []uuid.UUID{}
	}

	return &gen.UserMenusResponse{
		MenuIds:            menuIDs,
		AvailableMenuIds:   availableMenuIDs,
		HiddenMenuIds:      hiddenMenuIDs,
		ExpandedPackageIds: snapshot.ExpandedPackageIDs,
		DerivedSources:     buildMenuSourceItems(snapshot.AvailableMenuMap),
		HasPackageConfig:   snapshot.HasPackageConfig,
	}, nil
}

// ── SetUserMenus ─────────────────────────────────────────────────────────────

func (h *APIHandler) SetUserMenus(ctx context.Context, req *gen.UserMenusResponse, params gen.SetUserMenusParams) (*gen.MutationResult, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		return nil, err
	}

	appKey := models.DefaultAppKey
	menuIDs := []uuid.UUID{}
	if req != nil {
		menuIDs = req.MenuIds
	}

	snapshot, err := h.getPersonalSnapshotForUser(params.ID, appKey)
	if err != nil {
		h.logger.Error("get personal workspace snapshot for set-menus failed", zap.Error(err))
		return nil, err
	}
	if !snapshot.HasPackageConfig {
		return &gen.MutationResult{Success: false}, nil
	}

	availableMenuSet := uuidSetFromSlice(snapshot.AvailableMenuIDs)
	for _, menuID := range menuIDs {
		if !availableMenuSet[menuID] {
			return &gen.MutationResult{Success: false}, nil
		}
	}

	blockedMenuIDs := excludeUUIDsFromSlice(snapshot.AvailableMenuIDs, menuIDs)
	if err := appscope.ReplaceUserHiddenMenusInApp(h.db, params.ID, appKey, blockedMenuIDs); err != nil {
		h.logger.Error("set user hidden menus failed", zap.Error(err))
		return nil, err
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshPersonalWorkspaceUser(params.ID); err != nil {
			h.logger.Error("refresh personal workspace user after set-menus failed", zap.Error(err))
			return nil, err
		}
	}
	return ok(), nil
}

// ── GetUserPackages ───────────────────────────────────────────────────────────

func (h *APIHandler) GetUserPackages(ctx context.Context, params gen.GetUserPackagesParams) (*gen.UserPackagesResponse, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		return nil, err
	}
	appKey := resolveAppKey(params.AppKey)

	packageIDs, err := appscope.PackageIDsByUser(h.db, params.ID, appKey)
	if err != nil {
		h.logger.Error("get user packages failed", zap.Error(err))
		return nil, err
	}

	var refs []gen.FeaturePackageRef
	if len(packageIDs) > 0 {
		packages, pkgErr := h.featurePkgRepo.GetByIDs(packageIDs)
		if pkgErr != nil {
			h.logger.Error("get user package details failed", zap.Error(pkgErr))
			return nil, pkgErr
		}
		refs = make([]gen.FeaturePackageRef, 0, len(packages))
		for _, pkg := range packages {
			refs = append(refs, featurePkgToRef(pkg))
		}
	} else {
		refs = []gen.FeaturePackageRef{}
	}

	return &gen.UserPackagesResponse{
		PackageIds: packageIDs,
		Packages:   refs,
	}, nil
}

// ── SetUserPackages ───────────────────────────────────────────────────────────

func (h *APIHandler) SetUserPackages(ctx context.Context, req *gen.UUIDListRequest, params gen.SetUserPackagesParams) (*gen.MutationResult, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		return nil, err
	}
	if req == nil {
		req = &gen.UUIDListRequest{Ids: []uuid.UUID{}}
	}

	// Use the caller's user ID as grantedBy if available.
	var grantedBy *uuid.UUID
	if callerID, ok := userIDFromContext(ctx); ok {
		grantedBy = &callerID
	}

	// SetUserPackages has no app_key param; use the default app key.
	appKey := models.DefaultAppKey
	if err := appscope.ReplaceUserPackagesInApp(h.db, params.ID, appKey, req.Ids, grantedBy); err != nil {
		h.logger.Error("set user packages failed", zap.Error(err))
		return nil, err
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshPersonalWorkspaceUser(params.ID); err != nil {
			h.logger.Error("refresh personal workspace user after set-packages failed", zap.Error(err))
			return nil, err
		}
	}
	return ok(), nil
}

// ── GetUserPermissions ────────────────────────────────────────────────────────

func (h *APIHandler) GetUserPermissions(ctx context.Context, params gen.GetUserPermissionsParams) (*gen.UserPermissionsResponse, error) {
	appKey := resolveAppKey(params.AppKey)

	snapshot, err := h.getPersonalSnapshotForUser(params.ID, appKey)
	if err != nil {
		h.logger.Error("get user permissions snapshot failed", zap.Error(err))
		return nil, err
	}

	tree, err := h.menuSvc.GetTree(false, snapshot.MenuIDs, appKey, "")
	if err != nil {
		h.logger.Error("get menu tree for user permissions failed", zap.Error(err))
		return nil, err
	}

	return &gen.UserPermissionsResponse{MenuTree: buildUserPermissionTreeItems(tree)}, nil
}

// ── GetUserPermissionDiagnosis ────────────────────────────────────────────────

func (h *APIHandler) GetUserPermissionDiagnosis(ctx context.Context, params gen.GetUserPermissionDiagnosisParams) (*gen.UserPermissionDiagnosisResponse, error) {
	appKey := resolveAppKey(params.AppKey)
	permissionKey := ""
	if params.PermissionKey.Set {
		permissionKey = params.PermissionKey.Value
	}

	var collaborationWorkspaceID *uuid.UUID
	if params.CollaborationWorkspaceID.Set && params.CollaborationWorkspaceID.Value != "" {
		parsed, parseErr := uuid.Parse(params.CollaborationWorkspaceID.Value)
		if parseErr == nil {
			collaborationWorkspaceID = &parsed
		}
	}

	userEntity, err := h.userSvc.Get(params.ID)
	if err != nil {
		return nil, err
	}
	userInfo := buildUserPermissionDiagnosisUser(userEntity)

	if collaborationWorkspaceID == nil {
		snapshot, snapshotErr := h.getPersonalSnapshotForUser(params.ID, appKey)
		if snapshotErr != nil {
			h.logger.Error("get personal snapshot for diagnosis failed", zap.Error(snapshotErr))
			return nil, snapshotErr
		}
		resp := &gen.UserPermissionDiagnosisResponse{
			User:                           userInfo,
			Context:                        buildUserPermissionDiagnosisContext("personal", "", "", ""),
			Snapshot:                       buildPersonalSnapshotSummary(snapshot),
			Roles:                          []gen.UserPermissionRoleResult{},
			CollaborationWorkspacePackages: []gen.FeaturePackageRef{},
			Menus:                          []gen.UserPermissionMenuTreeItem{},
		}
		if permissionKey != "" {
			resp.Diagnosis = gen.NewOptUserPermissionDiagnosisResult(buildUserPermissionDiagnosisResult(permissionKey))
		}
		return resp, nil
	}

	// Collaboration workspace diagnosis.
	cwSnapshot, cwErr := h.boundarySvc.GetSnapshot(*collaborationWorkspaceID, normalizeAppKeyStr(appKey))
	if cwErr != nil {
		h.logger.Error("get collaboration workspace snapshot for diagnosis failed", zap.Error(cwErr))
		return nil, cwErr
	}

	currentCWID := ""
	if ws, wsErr := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(h.db, *collaborationWorkspaceID); wsErr == nil && ws != nil {
		currentCWID = ws.ID.String()
	}

	resp := &gen.UserPermissionDiagnosisResponse{
		User:                           userInfo,
		Context:                        buildUserPermissionDiagnosisContext("collaboration", currentCWID, collaborationWorkspaceID.String(), ""),
		Snapshot:                       buildCollaborationSnapshotSummary(cwSnapshot),
		Roles:                          []gen.UserPermissionRoleResult{},
		CollaborationWorkspacePackages: []gen.FeaturePackageRef{},
		Menus:                          []gen.UserPermissionMenuTreeItem{},
	}
	if permissionKey != "" {
		resp.Diagnosis = gen.NewOptUserPermissionDiagnosisResult(buildUserPermissionDiagnosisResult(permissionKey))
	}
	return resp, nil
}

// ── shared helpers ─────────────────────────────────────────────────────────────

// getPersonalSnapshotForUser returns a platformaccess snapshot, never nil.
func (h *APIHandler) getPersonalSnapshotForUser(userID uuid.UUID, appKey string) (*platformaccess.Snapshot, error) {
	if h.personalAccess == nil {
		return emptyPersonalSnapshot(), nil
	}
	snapshot, err := h.personalAccess.GetSnapshot(userID, appKey)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return emptyPersonalSnapshot(), nil
	}
	return snapshot, nil
}

func emptyPersonalSnapshot() *platformaccess.Snapshot {
	return &platformaccess.Snapshot{
		DirectPackageIDs:   []uuid.UUID{},
		ExpandedPackageIDs: []uuid.UUID{},
		ActionIDs:          []uuid.UUID{},
		ActionSourceMap:    map[uuid.UUID][]uuid.UUID{},
		AvailableMenuIDs:   []uuid.UUID{},
		AvailableMenuMap:   map[uuid.UUID][]uuid.UUID{},
		MenuIDs:            []uuid.UUID{},
		MenuSourceMap:      map[uuid.UUID][]uuid.UUID{},
		HiddenMenuIDs:      []uuid.UUID{},
		HasPackageConfig:   false,
	}
}

func uuidSliceToStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func excludeUUIDsFromSlice(source []uuid.UUID, selected []uuid.UUID) []uuid.UUID {
	sel := uuidSetFromSlice(selected)
	out := make([]uuid.UUID, 0, len(source))
	for _, id := range source {
		if !sel[id] {
			out = append(out, id)
		}
	}
	return out
}

func buildMenuSourceItems(sourceMap map[uuid.UUID][]uuid.UUID) []gen.UserMenuDerivedSourceItem {
	if len(sourceMap) == 0 {
		return []gen.UserMenuDerivedSourceItem{}
	}
	items := make([]gen.UserMenuDerivedSourceItem, 0, len(sourceMap))
	for menuID, packageIDs := range sourceMap {
		items = append(items, gen.UserMenuDerivedSourceItem{
			MenuID:     menuID,
			PackageIds: packageIDs,
		})
	}
	return items
}

func buildUserPermissionTreeItems(tree []*user.Menu) []gen.UserPermissionMenuTreeItem {
	items := make([]gen.UserPermissionMenuTreeItem, 0, len(tree))
	for _, item := range tree {
		if item == nil {
			continue
		}
		items = append(items, buildUserPermissionTreeItem(item))
	}
	return items
}

func buildUserPermissionTreeItem(item *user.Menu) gen.UserPermissionMenuTreeItem {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = strings.TrimSpace(item.Name)
	}
	return gen.UserPermissionMenuTreeItem{
		ID:        item.ID,
		Name:      item.Name,
		Title:     title,
		Path:      item.Path,
		Component: item.Component,
		Hidden:    item.Hidden,
		SortOrder: item.SortOrder,
		Children:  buildUserPermissionTreeItems(item.Children),
	}
}

func featurePkgToRef(pkg user.FeaturePackage) gen.FeaturePackageRef {
	ref := gen.FeaturePackageRef{
		ID:         pkg.ID,
		PackageKey: pkg.PackageKey,
		Name:       pkg.Name,
		Status:     pkg.Status,
	}
	if pkg.PackageType != "" {
		ref.PackageType = gen.NewOptNilString(pkg.PackageType)
	}
	if pkg.Description != "" {
		ref.Description = gen.NewOptNilString(pkg.Description)
	}
	if pkg.ContextType != "" {
		ref.ContextType = gen.NewOptNilString(pkg.ContextType)
	}
	ref.IsBuiltin = gen.NewOptBool(pkg.IsBuiltin)
	ref.SortOrder = gen.NewOptInt(pkg.SortOrder)
	return ref
}

func buildPersonalSnapshotSummary(s *platformaccess.Snapshot) gen.UserPermissionSnapshotSummary {
	if s == nil {
		return gen.UserPermissionSnapshotSummary{
			RoleCount:            0,
			DirectPackageCount:   0,
			ExpandedPackageCount: 0,
			ActionCount:          0,
			DisabledActionCount:  0,
			MenuCount:            0,
			HasPackageConfig:     false,
			DerivedActionCount:   0,
			BlockedActionCount:   0,
			EffectiveActionCount: 0,
		}
	}
	effectiveCount := len(s.ActionIDs) - len(s.DisabledActionIDs)
	if effectiveCount < 0 {
		effectiveCount = 0
	}
	return gen.UserPermissionSnapshotSummary{
		RoleCount:            0,
		DirectPackageCount:   len(s.DirectPackageIDs),
		ExpandedPackageCount: len(s.ExpandedPackageIDs),
		ActionCount:          len(s.ActionIDs),
		DisabledActionCount:  len(s.DisabledActionIDs),
		MenuCount:            len(s.MenuIDs),
		HasPackageConfig:     s.HasPackageConfig,
		DerivedActionCount:   len(s.ActionIDs),
		BlockedActionCount:   0,
		EffectiveActionCount: effectiveCount,
	}
}

func buildCollaborationSnapshotSummary(s *collaborationworkspaceboundary.Snapshot) gen.UserPermissionSnapshotSummary {
	if s == nil {
		return gen.UserPermissionSnapshotSummary{
			RoleCount:            0,
			DirectPackageCount:   0,
			ExpandedPackageCount: 0,
			ActionCount:          0,
			DisabledActionCount:  0,
			MenuCount:            0,
			HasPackageConfig:     false,
			DerivedActionCount:   0,
			BlockedActionCount:   0,
			EffectiveActionCount: 0,
		}
	}
	return gen.UserPermissionSnapshotSummary{
		RoleCount:            0,
		DirectPackageCount:   len(s.PackageIDs),
		ExpandedPackageCount: len(s.ExpandedPackageIDs),
		ActionCount:          len(s.EffectiveIDs),
		DisabledActionCount:  0,
		MenuCount:            0,
		HasPackageConfig:     len(s.PackageIDs) > 0,
		DerivedActionCount:   len(s.DerivedIDs),
		BlockedActionCount:   len(s.BlockedIDs),
		EffectiveActionCount: len(s.EffectiveIDs),
	}
}

func buildUserPermissionDiagnosisUser(entity *user.User) gen.UserPermissionDiagnosisUser {
	out := gen.UserPermissionDiagnosisUser{
		ID:           entity.ID,
		UserName:     entity.Username,
		Status:       entity.Status,
		IsSuperAdmin: entity.IsSuperAdmin,
	}
	if entity.Nickname != "" {
		out.NickName = gen.NewOptString(entity.Nickname)
	}
	return out
}

func buildUserPermissionDiagnosisContext(contextType, bindingWorkspaceID, currentWorkspaceID, currentWorkspaceName string) gen.UserPermissionDiagnosisContext {
	return gen.UserPermissionDiagnosisContext{
		Type:                              contextType,
		BindingWorkspaceID:                bindingWorkspaceID,
		CurrentCollaborationWorkspaceID:   currentWorkspaceID,
		CurrentCollaborationWorkspaceName: currentWorkspaceName,
	}
}

func buildUserPermissionDiagnosisResult(permissionKey string) gen.UserPermissionDiagnosisResult {
	return gen.UserPermissionDiagnosisResult{
		PermissionKey:                   permissionKey,
		Allowed:                         false,
		Reasons:                         []string{},
		MatchedInSnapshot:               false,
		BypassedBySuperAdmin:            false,
		BlockedByCollaborationWorkspace: false,
		MemberMatched:                   false,
		BoundaryConfigured:              false,
		RoleChainMatched:                false,
		RoleChainDisabled:               false,
		RoleChainAvailable:              false,
		SourcePackages:                  []gen.FeaturePackageRef{},
		RoleResults:                     []gen.UserPermissionRoleResult{},
	}
}

