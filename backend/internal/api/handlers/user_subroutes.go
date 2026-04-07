// user_subroutes.go: ogen handler implementations for user sub-routes.
// Phase 4: GetUserMenus, SetUserMenus, GetUserPackages, SetUserPackages,
// GetUserPermissions, GetUserPermissionDiagnosis.
package handlers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/appscope"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
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

func (h *APIHandler) GetUserMenus(ctx context.Context, params gen.GetUserMenusParams) (gen.AnyObject, error) {
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

	return marshalAnyObject(map[string]interface{}{
		"menu_ids":             uuidSliceToStrings(menuIDs),
		"available_menu_ids":   uuidSliceToStrings(availableMenuIDs),
		"hidden_menu_ids":      uuidSliceToStrings(hiddenMenuIDs),
		"expanded_package_ids": uuidSliceToStrings(snapshot.ExpandedPackageIDs),
		"derived_sources":      buildMenuSourceMaps(snapshot.AvailableMenuMap),
		"has_package_config":   snapshot.HasPackageConfig,
	}), nil
}

// ── SetUserMenus ─────────────────────────────────────────────────────────────

func (h *APIHandler) SetUserMenus(ctx context.Context, req gen.AnyObject, params gen.SetUserMenusParams) (*gen.MutationResult, error) {
	if _, err := h.userSvc.Get(params.ID); err != nil {
		return nil, err
	}

	// Extract app_key and menu_ids from the AnyObject body.
	var body struct {
		AppKey  string   `json:"app_key"`
		MenuIDs []string `json:"menu_ids"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	appKey := normalizeAppKeyStr(body.AppKey)

	snapshot, err := h.getPersonalSnapshotForUser(params.ID, appKey)
	if err != nil {
		h.logger.Error("get personal workspace snapshot for set-menus failed", zap.Error(err))
		return nil, err
	}
	if !snapshot.HasPackageConfig {
		return &gen.MutationResult{Success: false}, nil
	}

	availableMenuSet := uuidSetFromSlice(snapshot.AvailableMenuIDs)
	menuIDs := make([]uuid.UUID, 0, len(body.MenuIDs))
	for _, s := range body.MenuIDs {
		menuID, parseErr := uuid.Parse(s)
		if parseErr != nil {
			return nil, parseErr
		}
		if !availableMenuSet[menuID] {
			return &gen.MutationResult{Success: false}, nil
		}
		menuIDs = append(menuIDs, menuID)
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

func (h *APIHandler) GetUserPermissions(ctx context.Context, params gen.GetUserPermissionsParams) (gen.AnyObject, error) {
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

	return marshalAnyObject(map[string]interface{}{"menu_tree": marshalList(tree)}), nil
}

// ── GetUserPermissionDiagnosis ────────────────────────────────────────────────

func (h *APIHandler) GetUserPermissionDiagnosis(ctx context.Context, params gen.GetUserPermissionDiagnosisParams) (gen.AnyObject, error) {
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

	userInfo := map[string]interface{}{
		"id":             userEntity.ID.String(),
		"user_name":      userEntity.Username,
		"nick_name":      userEntity.Nickname,
		"status":         userEntity.Status,
		"is_super_admin": userEntity.IsSuperAdmin,
	}

	if collaborationWorkspaceID == nil {
		snapshot, snapshotErr := h.getPersonalSnapshotForUser(params.ID, appKey)
		if snapshotErr != nil {
			h.logger.Error("get personal snapshot for diagnosis failed", zap.Error(snapshotErr))
			return nil, snapshotErr
		}
		payload := map[string]interface{}{
			"user":      userInfo,
			"context":   map[string]interface{}{"type": "personal", "binding_workspace_id": "", "current_collaboration_workspace_id": "", "current_collaboration_workspace_name": ""},
			"snapshot":  buildPersonalSnapshotSummary(snapshot),
			"roles":     []interface{}{},
			"diagnosis": nil,
		}
		if permissionKey != "" {
			payload["diagnosis"] = map[string]interface{}{"permission_key": permissionKey}
		}
		return marshalAnyObject(payload), nil
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

	payload := map[string]interface{}{
		"user": userInfo,
		"context": map[string]interface{}{
			"type":                                 "collaboration",
			"binding_workspace_id":                 currentCWID,
			"current_collaboration_workspace_id":   collaborationWorkspaceID.String(),
			"current_collaboration_workspace_name": "",
		},
		"snapshot":  marshalAnyObject(cwSnapshot),
		"roles":     []interface{}{},
		"diagnosis": nil,
	}
	if permissionKey != "" {
		payload["diagnosis"] = map[string]interface{}{"permission_key": permissionKey}
	}
	return marshalAnyObject(payload), nil
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

func buildMenuSourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []map[string]interface{} {
	if len(sourceMap) == 0 {
		return []map[string]interface{}{}
	}
	items := make([]map[string]interface{}, 0, len(sourceMap))
	for menuID, packageIDs := range sourceMap {
		items = append(items, map[string]interface{}{
			"menu_id":     menuID.String(),
			"package_ids": uuidSliceToStrings(packageIDs),
		})
	}
	return items
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

func buildPersonalSnapshotSummary(s *platformaccess.Snapshot) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"direct_package_ids":   uuidSliceToStrings(s.DirectPackageIDs),
		"expanded_package_ids": uuidSliceToStrings(s.ExpandedPackageIDs),
		"action_ids":           uuidSliceToStrings(s.ActionIDs),
		"menu_ids":             uuidSliceToStrings(s.MenuIDs),
		"available_menu_ids":   uuidSliceToStrings(s.AvailableMenuIDs),
		"has_package_config":   s.HasPackageConfig,
	}
}
