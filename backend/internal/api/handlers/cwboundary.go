// cwboundary.go — ogen handler implementations for CW boundary / current
// complex ops (Phase 4 CW boundary migration).
package handlers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/appscope"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// cwIDFromCtx reads the collaboration_workspace_id injected by the router
// middleware into the request context.
func cwIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	raw := stringFromCtx(ctx, CtxCollaborationWorkspaceID)
	if raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// resolveCWMember retrieves the CollaborationWorkspaceMember for the current
// user+workspace. It relies on the collaboration_workspace_id context value
// seeded by the router middleware.
func (h *APIHandler) resolveCWMember(ctx context.Context) (*user.CollaborationWorkspaceMember, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, errNoUser
	}
	cwID, ok := cwIDFromCtx(ctx)
	if !ok {
		return nil, errNoCW
	}
	member, err := h.cwMemberRepo.GetByUserAndCollaborationWorkspace(userID, cwID)
	if err != nil {
		return nil, err
	}
	return member, nil
}

// resolveCWRole resolves the member + a CW-scoped role by roleId.
func (h *APIHandler) resolveCWRole(ctx context.Context, roleID uuid.UUID) (*user.CollaborationWorkspaceMember, *user.Role, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return nil, nil, err
	}
	role, err := h.roleRepo.GetByID(roleID)
	if err != nil {
		return member, nil, err
	}
	if !cwIsAssignableRole(*role, member.CollaborationWorkspaceID) {
		return member, nil, errRoleForbidden
	}
	return member, role, nil
}

// resolveCWRoleEditable is like resolveCWRole but also validates that the role
// is workspace-specific (not global/system).
func (h *APIHandler) resolveCWRoleEditable(ctx context.Context, roleID uuid.UUID) (*user.CollaborationWorkspaceMember, *user.Role, error) {
	member, role, err := h.resolveCWRole(ctx, roleID)
	if err != nil {
		return member, role, err
	}
	if role.CollaborationWorkspaceID == nil {
		return member, role, errRoleForbidden
	}
	return member, role, nil
}

// cwIsAssignableRole mirrors the legacy isAssignableRoleForCollaborationWorkspace.
func cwIsAssignableRole(role user.Role, collaborationWorkspaceID uuid.UUID) bool {
	if role.CollaborationWorkspaceID == nil {
		return true
	}
	return *role.CollaborationWorkspaceID == collaborationWorkspaceID
}

// cwGetWorkspaceAwareTeamRoleIDs retrieves effective team role IDs for a user.
func (h *APIHandler) cwGetWorkspaceAwareTeamRoleIDs(userID, cwID uuid.UUID) ([]uuid.UUID, error) {
	roleIDs, err := workspacerolebinding.ListCollaborationWorkspaceRoleIDsByUser(h.db, cwID, userID, false)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) > 0 {
		return roleIDs, nil
	}
	return h.userRoleRepo.GetRoleIDsByUserAndCollaborationWorkspace(userID, &cwID, h.cwMemberRepo)
}

// cwSyncWorkspaceRoleBindings replaces CW role bindings for a user.
func (h *APIHandler) cwSyncWorkspaceRoleBindings(cwID, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		return workspacerolebinding.ReplaceCollaborationWorkspaceRoleBindings(tx, cwID, userID, roleIDs)
	})
}

// uuidsToStrings converts []uuid.UUID to []string.
func uuidsToStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

// excludeUUIDs returns source minus excluded (ordered, no dup).
func excludeUUIDs(source, excluded []uuid.UUID) []uuid.UUID {
	excSet := make(map[uuid.UUID]struct{}, len(excluded))
	for _, id := range excluded {
		excSet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(source))
	seen := make(map[uuid.UUID]struct{}, len(source))
	for _, id := range source {
		if _, skip := excSet[id]; skip {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// uuidSetFromSlice builds a set from a slice.
func uuidSetFromSlice(ids []uuid.UUID) map[uuid.UUID]bool {
	out := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out
}

// parseUUIDs converts []uuid.UUID (already decoded by ogen) to a clean slice.
func parseUUIDs(ids []uuid.UUID) []uuid.UUID { return ids }

// roleToMap serialises a Role to a generic map.
func roleToMap(r user.Role) map[string]interface{} {
	cwID := ""
	if r.CollaborationWorkspaceID != nil {
		cwID = r.CollaborationWorkspaceID.String()
	}
	return map[string]interface{}{
		"id":                         r.ID.String(),
		"code":                       r.Code,
		"name":                       r.Name,
		"description":                r.Description,
		"status":                     r.Status,
		"is_system":                  r.IsSystem,
		"collaboration_workspace_id": cwID,
		"is_global":                  r.CollaborationWorkspaceID == nil,
		"create_time":                r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// pkgToMap serialises a FeaturePackage.
func pkgToMap(p user.FeaturePackage) map[string]interface{} {
	return map[string]interface{}{
		"id":           p.ID.String(),
		"package_key":  p.PackageKey,
		"name":         p.Name,
		"description":  p.Description,
		"context_type": p.ContextType,
		"status":       p.Status,
		"sort_order":   p.SortOrder,
		"created_at":   p.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// actionKeyToMap serialises a PermissionKey.
func actionKeyToMap(k user.PermissionKey) map[string]interface{} {
	return map[string]interface{}{
		"id":             k.ID.String(),
		"module_code":    k.ModuleCode,
		"context_type":   k.ContextType,
		"permission_key": k.PermissionKey,
		"feature_kind":   k.FeatureKind,
		"name":           k.Name,
		"description":    k.Description,
		"status":         k.Status,
		"sort_order":     k.SortOrder,
		"created_at":     k.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":     k.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func derSourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []map[string]interface{} {
	if len(sourceMap) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(sourceMap))
	for actionID, pkgIDs := range sourceMap {
		out = append(out, map[string]interface{}{
			"action_id":   actionID.String(),
			"package_ids": uuidsToStrings(pkgIDs),
		})
	}
	return out
}

func menuSourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []map[string]interface{} {
	if len(sourceMap) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(sourceMap))
	for menuID, pkgIDs := range sourceMap {
		out = append(out, map[string]interface{}{
			"menu_id":     menuID.String(),
			"package_ids": uuidsToStrings(pkgIDs),
		})
	}
	return out
}

// sentinel errors (internal only)
var (
	errNoUser       = errText("未认证")
	errNoCW         = errText("no current collaboration workspace")
	errRoleForbidden = errText("仅支持操作当前协作空间自定义角色")
)

type errText string

func (e errText) Error() string { return string(e) }

// ─── ListCurrentCollaborationWorkspaceRoles ───────────────────────────────────

func (h *APIHandler) ListCurrentCollaborationWorkspaceRoles(ctx context.Context) (*gen.AnyListResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.AnyListResponse{Records: []gen.AnyObject{}, Total: 0}, nil
	}
	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("list cw roles failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(allRoles), Total: len(allRoles)}, nil
}

// ─── CreateCurrentCollaborationWorkspaceRole ──────────────────────────────────

func (h *APIHandler) CreateCurrentCollaborationWorkspaceRole(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return nil, err
	}
	var body struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
		Priority    int    `json:"priority"`
		Status      string `json:"status"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(body.Code)
	if code == "" || strings.TrimSpace(body.Name) == "" {
		return nil, errText("角色编码和角色名称不能为空")
	}
	existingRoles, err := h.roleRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingRoles {
		if existing.CollaborationWorkspaceID != nil && *existing.CollaborationWorkspaceID == member.CollaborationWorkspaceID {
			return nil, errText("角色编码已存在")
		}
	}
	role := &user.Role{
		CollaborationWorkspaceID: &member.CollaborationWorkspaceID,
		Code:                     code,
		Name:                     strings.TrimSpace(body.Name),
		Description:              strings.TrimSpace(body.Description),
		SortOrder:                body.SortOrder,
		Priority:                 body.Priority,
		Status:                   cwNormalizeStatus(body.Status),
	}
	if err := h.roleRepo.Create(role); err != nil {
		h.logger.Error("create cw role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── ListCurrentCollaborationWorkspaceBoundaryRoles ───────────────────────────

func (h *APIHandler) ListCurrentCollaborationWorkspaceBoundaryRoles(ctx context.Context) (*gen.AnyListResponse, error) {
	return h.ListCurrentCollaborationWorkspaceRoles(ctx)
}

// ─── CreateCurrentCollaborationWorkspaceBoundaryRole ─────────────────────────

func (h *APIHandler) CreateCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	return h.CreateCurrentCollaborationWorkspaceRole(ctx, req)
}

// ─── UpdateCurrentCollaborationWorkspaceBoundaryRole ─────────────────────────

func (h *APIHandler) UpdateCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, req gen.AnyObject, params gen.UpdateCurrentCollaborationWorkspaceBoundaryRoleParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	var body struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
		Priority    int    `json:"priority"`
		Status      string `json:"status"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"name":        strings.TrimSpace(cwDefaultString(body.Name, role.Name)),
		"description": strings.TrimSpace(cwDefaultString(body.Description, role.Description)),
		"sort_order":  body.SortOrder,
		"priority":    body.Priority,
		"status":      cwNormalizeStatus(cwDefaultString(body.Status, role.Status)),
	}
	if code := strings.TrimSpace(body.Code); code != "" && code != role.Code {
		existingRoles, findErr := h.roleRepo.FindByCode(code)
		if findErr != nil {
			return nil, findErr
		}
		for _, existing := range existingRoles {
			if existing.ID == role.ID {
				continue
			}
			if existing.CollaborationWorkspaceID != nil && *existing.CollaborationWorkspaceID == member.CollaborationWorkspaceID {
				return nil, errText("角色编码已存在")
			}
		}
		updates["code"] = code
	}
	if err := h.roleRepo.UpdateWithMap(role.ID, updates); err != nil {
		h.logger.Error("update cw role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── DeleteCurrentCollaborationWorkspaceBoundaryRole ─────────────────────────

func (h *APIHandler) DeleteCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, params gen.DeleteCurrentCollaborationWorkspaceBoundaryRoleParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND collaboration_workspace_id = ?", role.ID, member.CollaborationWorkspaceID).
			Delete(&user.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleFeaturePackage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleHiddenMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleDisabledAction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleDataPermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user.Role{}, role.ID).Error
	}); err != nil {
		h.logger.Error("delete cw role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── GetCurrentCollaborationWorkspaceBoundaryRolePackages ─────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryRolePackages(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceBoundaryRolePackagesParams) (gen.AnyObject, error) {
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
	return marshalAnyObject(map[string]interface{}{
		"package_ids": uuidsToStrings(snapshot.PackageIDs),
		"packages":    marshalList(pkgs),
		"inherited":   snapshot.Inherited,
	}), nil
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

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryRoleMenus(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceBoundaryRoleMenusParams) (gen.AnyObject, error) {
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
	return marshalAnyObject(map[string]interface{}{
		"menu_ids":             uuidsToStrings(snapshot.MenuIDs),
		"available_menu_ids":   uuidsToStrings(snapshot.AvailableMenuIDs),
		"hidden_menu_ids":      uuidsToStrings(snapshot.HiddenMenuIDs),
		"expanded_package_ids": uuidsToStrings(snapshot.ExpandedPackageIDs),
		"derived_sources":      menuSourceMaps(snapshot.MenuSourceMap),
	}), nil
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

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryRoleActions(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceBoundaryRoleActionsParams) (gen.AnyObject, error) {
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
	return marshalAnyObject(map[string]interface{}{
		"action_ids":           uuidsToStrings(snapshot.ActionIDs),
		"available_action_ids": uuidsToStrings(snapshot.AvailableActionIDs),
		"disabled_action_ids":  uuidsToStrings(snapshot.DisabledActionIDs),
		"actions":              marshalList(actions),
		"expanded_package_ids": uuidsToStrings(snapshot.ExpandedPackageIDs),
		"derived_sources":      derSourceMaps(snapshot.ActionSourceMap),
	}), nil
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

func (h *APIHandler) GetCurrentCollaborationWorkspaceBoundaryPackages(ctx context.Context) (*gen.AnyListResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.AnyListResponse{Records: []gen.AnyObject{}, Total: 0}, nil
	}
	packageIDs, err := appscope.PackageIDsByCollaborationWorkspace(h.db, member.CollaborationWorkspaceID, "")
	if err != nil {
		h.logger.Error("get cw boundary packages failed", zap.Error(err))
		return nil, err
	}
	if len(packageIDs) == 0 {
		return &gen.AnyListResponse{Records: []gen.AnyObject{}, Total: 0}, nil
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
	return &gen.AnyListResponse{Records: marshalList(filtered), Total: len(filtered)}, nil
}

// ─── GetCurrentCollaborationWorkspaceMenus ────────────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceMenus(ctx context.Context) (gen.AnyObject, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return marshalAnyObject(map[string]interface{}{"menu_ids": []string{}}), nil
	}
	snapshot, err := h.boundarySvc.GetMenuSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw menus failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(map[string]interface{}{
		"menu_ids": uuidsToStrings(snapshot.EffectiveIDs),
	}), nil
}

// ─── GetCurrentCollaborationWorkspaceMenuOrigins ──────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceMenuOrigins(ctx context.Context) (*gen.AnyListResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.AnyListResponse{Records: []gen.AnyObject{}, Total: 0}, nil
	}
	snapshot, err := h.boundarySvc.GetMenuSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw menu origins failed", zap.Error(err))
		return nil, err
	}
	item := marshalAnyObject(map[string]interface{}{
		"derived_menu_ids": uuidsToStrings(snapshot.DerivedIDs),
		"derived_sources":  menuSourceMaps(snapshot.DerivedMap),
		"blocked_menu_ids": uuidsToStrings(snapshot.BlockedIDs),
	})
	return &gen.AnyListResponse{Records: []gen.AnyObject{item}, Total: 1}, nil
}

// ─── GetCurrentCollaborationWorkspaceActions ──────────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceActions(ctx context.Context) (gen.AnyObject, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return marshalAnyObject(map[string]interface{}{"action_ids": []string{}, "actions": []interface{}{}}), nil
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
	return marshalAnyObject(map[string]interface{}{
		"action_ids": uuidsToStrings(snapshot.EffectiveIDs),
		"actions":    marshalList(actions),
	}), nil
}

// ─── GetCurrentCollaborationWorkspaceActionOrigins ────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceActionOrigins(ctx context.Context) (*gen.AnyListResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.AnyListResponse{Records: []gen.AnyObject{}, Total: 0}, nil
	}
	snapshot, err := h.boundarySvc.GetSnapshot(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get current cw action origins failed", zap.Error(err))
		return nil, err
	}
	item := marshalAnyObject(map[string]interface{}{
		"derived_action_ids": uuidsToStrings(snapshot.DerivedIDs),
		"derived_sources":    derSourceMaps(snapshot.DerivedMap),
		"blocked_action_ids": uuidsToStrings(snapshot.BlockedIDs),
	})
	return &gen.AnyListResponse{Records: []gen.AnyObject{item}, Total: 1}, nil
}

// ─── GetCurrentCollaborationWorkspaceMemberRoles ──────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceMemberRoles(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceMemberRolesParams) (gen.AnyObject, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return marshalAnyObject(map[string]interface{}{"role_ids": []string{}}), nil
	}
	targetMember, err := h.cwMemberRepo.GetByUserAndCollaborationWorkspace(params.UserId, member.CollaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errText("成员不存在")
		}
		return nil, err
	}
	roleIDs, err := h.cwGetWorkspaceAwareTeamRoleIDs(params.UserId, member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get cw member role ids failed", zap.Error(err))
		return nil, err
	}
	roles, err := h.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	roleIDsStr := make([]string, 0, len(roles))
	roleList := make([]map[string]interface{}, 0, len(roles))
	for _, r := range roles {
		roleIDsStr = append(roleIDsStr, r.ID.String())
		roleList = append(roleList, map[string]interface{}{"id": r.ID.String(), "code": r.Code, "name": r.Name})
	}
	_ = targetMember // used for binding meta if needed
	return marshalAnyObject(map[string]interface{}{
		"role_ids": roleIDsStr,
		"roles":    roleList,
	}), nil
}

// ─── SetCurrentCollaborationWorkspaceMemberRoles ──────────────────────────────

func (h *APIHandler) SetCurrentCollaborationWorkspaceMemberRoles(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationWorkspaceMemberRolesParams) (*gen.MutationResult, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return nil, err
	}
	targetMember, err := h.cwMemberRepo.GetByUserAndCollaborationWorkspace(params.UserId, member.CollaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errText("成员不存在")
		}
		return nil, err
	}
	roleIDs := uuidIDsFromRequest(req)

	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(member.CollaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	allowedTeamRoleIDs := make(map[uuid.UUID]user.Role)
	protectedRoleID := uuid.Nil
	for _, role := range allRoles {
		allowedTeamRoleIDs[role.ID] = role
		if role.Code == targetMember.RoleCode {
			protectedRoleID = role.ID
		}
	}
	filteredIDs := make([]uuid.UUID, 0, len(roleIDs)+1)
	seen := make(map[uuid.UUID]struct{}, len(roleIDs)+1)
	for _, roleID := range roleIDs {
		if _, ok := allowedTeamRoleIDs[roleID]; !ok {
			continue
		}
		if _, dup := seen[roleID]; dup {
			continue
		}
		seen[roleID] = struct{}{}
		filteredIDs = append(filteredIDs, roleID)
	}
	if protectedRoleID != uuid.Nil {
		if _, exists := seen[protectedRoleID]; !exists {
			filteredIDs = append(filteredIDs, protectedRoleID)
		}
	}

	if err := h.userRoleRepo.SetUserRoles(params.UserId, filteredIDs, &member.CollaborationWorkspaceID); err != nil {
		h.logger.Error("set cw member roles failed", zap.Error(err))
		return nil, err
	}
	if err := h.cwSyncWorkspaceRoleBindings(member.CollaborationWorkspaceID, params.UserId, filteredIDs); err != nil {
		h.logger.Error("sync cw role bindings failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── GetCollaborationWorkspaceMenus ──────────────────────────────────────────

func (h *APIHandler) GetCollaborationWorkspaceMenus(ctx context.Context, params gen.GetCollaborationWorkspaceMenusParams) (gen.AnyObject, error) {
	snapshot, err := h.boundarySvc.GetMenuSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw menus failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(map[string]interface{}{
		"menu_ids": uuidsToStrings(snapshot.EffectiveIDs),
	}), nil
}

// ─── GetCollaborationWorkspaceMenuOrigins ─────────────────────────────────────

func (h *APIHandler) GetCollaborationWorkspaceMenuOrigins(ctx context.Context, params gen.GetCollaborationWorkspaceMenuOriginsParams) (*gen.AnyListResponse, error) {
	snapshot, err := h.boundarySvc.GetMenuSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw menu origins failed", zap.Error(err))
		return nil, err
	}
	item := marshalAnyObject(map[string]interface{}{
		"derived_menu_ids": uuidsToStrings(snapshot.DerivedIDs),
		"derived_sources":  menuSourceMaps(snapshot.DerivedMap),
		"blocked_menu_ids": uuidsToStrings(snapshot.BlockedIDs),
	})
	return &gen.AnyListResponse{Records: []gen.AnyObject{item}, Total: 1}, nil
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

func (h *APIHandler) GetCollaborationWorkspaceActions(ctx context.Context, params gen.GetCollaborationWorkspaceActionsParams) (gen.AnyObject, error) {
	snapshot, err := h.boundarySvc.GetSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw actions failed", zap.Error(err))
		return nil, err
	}
	actions, err := h.keyRepo.GetByIDs(snapshot.EffectiveIDs)
	if err != nil {
		return nil, err
	}
	return marshalAnyObject(map[string]interface{}{
		"action_ids": uuidsToStrings(snapshot.EffectiveIDs),
		"actions":    marshalList(actions),
	}), nil
}

// ─── GetCollaborationWorkspaceActionOrigins ───────────────────────────────────

func (h *APIHandler) GetCollaborationWorkspaceActionOrigins(ctx context.Context, params gen.GetCollaborationWorkspaceActionOriginsParams) (*gen.AnyListResponse, error) {
	snapshot, err := h.boundarySvc.GetSnapshot(params.ID)
	if err != nil {
		h.logger.Error("get cw action origins failed", zap.Error(err))
		return nil, err
	}
	item := marshalAnyObject(map[string]interface{}{
		"derived_action_ids": uuidsToStrings(snapshot.DerivedIDs),
		"derived_sources":    derSourceMaps(snapshot.DerivedMap),
		"blocked_action_ids": uuidsToStrings(snapshot.BlockedIDs),
	})
	return &gen.AnyListResponse{Records: []gen.AnyObject{item}, Total: 1}, nil
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

// ─── ListCollaborationWorkspaceRoles ──────────────────────────────────────────

func (h *APIHandler) ListCollaborationWorkspaceRoles(ctx context.Context, params gen.ListCollaborationWorkspaceRolesParams) (*gen.AnyListResponse, error) {
	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(params.ID)
	if err != nil {
		h.logger.Error("list cw roles failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(allRoles), Total: len(allRoles)}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func cwNormalizeStatus(status string) string {
	v := strings.TrimSpace(status)
	if v == "" {
		return "normal"
	}
	return v
}

func cwDefaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
