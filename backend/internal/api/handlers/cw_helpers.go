// cw_helpers.go — shared helpers for CW boundary handlers.
package handlers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/workspacerolebinding"
)

// ─── context helpers ──────────────────────────────────────────────────────────

// cwIDFromCtx reads the collaboration_id injected by the router
// middleware into the request context.
func cwIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	raw := stringFromCtx(ctx, CtxCollaborationID)
	if raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// ─── member/role resolution ───────────────────────────────────────────────────

// resolveCWMember retrieves the CollaborationWorkspaceMember for the current
// user+workspace.
func (h *cwAPIHandler) resolveCWMember(ctx context.Context) (*user.CollaborationWorkspaceMember, error) {
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
func (h *cwAPIHandler) resolveCWRole(ctx context.Context, roleID uuid.UUID) (*user.CollaborationWorkspaceMember, *user.Role, error) {
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
func (h *cwAPIHandler) resolveCWRoleEditable(ctx context.Context, roleID uuid.UUID) (*user.CollaborationWorkspaceMember, *user.Role, error) {
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
// V5 真相源：仅查 workspace_role_bindings，旧 user_roles 回退已废弃。
func (h *cwAPIHandler) cwGetWorkspaceAwareTeamRoleIDs(userID, cwID uuid.UUID) ([]uuid.UUID, error) {
	return workspacerolebinding.ListCollaborationWorkspaceRoleIDsByUser(h.db, cwID, userID, false)
}

// cwSyncWorkspaceRoleBindings replaces CW role bindings for a user.
func (h *cwAPIHandler) cwSyncWorkspaceRoleBindings(cwID, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		return workspacerolebinding.ReplaceCollaborationWorkspaceRoleBindings(tx, cwID, userID, roleIDs)
	})
}

// ─── slice helpers ────────────────────────────────────────────────────────────

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

// ─── model serialisers ────────────────────────────────────────────────────────

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
		"collaboration_id": cwID,
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

// ─── sentinel errors ──────────────────────────────────────────────────────────

var (
	errNoUser        = errText("未认证")
	errNoCW          = errText("no current collaboration workspace")
	errRoleForbidden = errText("仅支持操作当前协作空间自定义角色")
)

type errText string

func (e errText) Error() string { return string(e) }

// ─── status/string helpers ────────────────────────────────────────────────────

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


