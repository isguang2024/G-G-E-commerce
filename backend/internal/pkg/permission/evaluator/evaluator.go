// Package evaluator is the single entry point for permission decisions in
// GGE 5.0. The final permission set for a member in a workspace is the
// intersection of:
//
//  1. Permission keys exposed by the feature packages bound to the workspace
//     (workspace_feature_packages → feature_package_keys → permission_keys)
//  2. Permission keys carried by the roles assigned to the member in that
//     workspace (workspace_role_bindings → role permissions)
//
// Phase 3 ships the interface, the workspace-feature-package side of the
// intersection, and a placeholder /permissions/explain payload. The role
// side and the casbin enforcer wrapper land in a follow-up PR alongside the
// permission_keys → api_endpoints DB consistency check; the public surface
// of this package will not change.
package evaluator

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ResolvedPermissions is the flat permission set for a member in a workspace.
// Keys are de-duplicated permission_key strings (e.g. "workspace.read").
type ResolvedPermissions struct {
	AccountID   uuid.UUID
	WorkspaceID uuid.UUID
	Keys        map[string]struct{}
}

// Has reports whether the resolved set contains the supplied key.
func (r *ResolvedPermissions) Has(key string) bool {
	if r == nil {
		return false
	}
	_, ok := r.Keys[key]
	return ok
}

// Explanation augments ResolvedPermissions with attribution metadata so the
// /permissions/explain endpoint can show *why* a key was granted (which
// feature package, which role). Phase 3 only fills FeaturePackageKeys; the
// role provenance lands together with the role-side intersection.
type Explanation struct {
	Resolved            *ResolvedPermissions
	FeaturePackageKeys  map[string][]uuid.UUID // permission_key -> source feature package ids
	RoleKeys            map[string][]uuid.UUID // permission_key -> source role ids (TODO)
	UnresolvedKeys      []string               // requested keys that were not granted
}

// Evaluator is the single entry point for permission decisions. Every
// handler that needs to gate behaviour on permission_keys MUST go through
// this interface — no direct table reads from business code.
type Evaluator interface {
	Resolve(ctx context.Context, accountID, workspaceID uuid.UUID) (*ResolvedPermissions, error)
	Can(ctx context.Context, accountID, workspaceID uuid.UUID, key string) (bool, error)
	Explain(ctx context.Context, accountID, workspaceID uuid.UUID) (*Explanation, error)
}

// New constructs the default GORM-backed evaluator.
func New(db *gorm.DB, logger *zap.Logger) Evaluator {
	return &gormEvaluator{db: db, logger: logger}
}

type gormEvaluator struct {
	db     *gorm.DB
	logger *zap.Logger
}

func (e *gormEvaluator) Resolve(ctx context.Context, accountID, workspaceID uuid.UUID) (*ResolvedPermissions, error) {
	if e.db == nil {
		return nil, errors.New("evaluator: database not initialized")
	}
	resolved := &ResolvedPermissions{
		AccountID:   accountID,
		WorkspaceID: workspaceID,
		Keys:        make(map[string]struct{}),
	}

	// super_admin shortcut: returns every permission_key in the table.
	isSuper, err := e.isSuperAdmin(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("evaluator: super_admin check: %w", err)
	}
	if isSuper {
		allKeys, err := e.queryAllPermissionKeys(ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluator: load all keys: %w", err)
		}
		for _, k := range allKeys {
			resolved.Keys[k] = struct{}{}
		}
		return resolved, nil
	}

	// Account-only path: union across all workspaces the account is a member of.
	if workspaceID == uuid.Nil {
		keys, err := e.queryAccountUnionKeys(ctx, accountID)
		if err != nil {
			return nil, fmt.Errorf("evaluator: account union: %w", err)
		}
		for _, k := range keys {
			resolved.Keys[k] = struct{}{}
		}
		return resolved, nil
	}

	// 1. Workspace (feature-package) side: the upper bound of what this
	//    workspace as a tenant subject can possibly do.
	wsKeys, err := e.queryFeaturePackageKeys(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluator: load feature package keys: %w", err)
	}

	// 2. Role side: what the *member* can do via assigned roles.
	//    Owner/admin bypass: for workspace owners/admins we don't run a
	//    role-based filter (their effective set is the workspace upper
	//    bound). For everyone else we intersect with role-derived keys.
	bypass, err := e.isOwnerOrAdmin(ctx, accountID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluator: check member type: %w", err)
	}
	if bypass {
		for _, key := range wsKeys {
			resolved.Keys[key] = struct{}{}
		}
		return resolved, nil
	}

	roleKeys, err := e.queryRoleKeys(ctx, accountID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluator: load role keys: %w", err)
	}
	roleSet := make(map[string]struct{}, len(roleKeys))
	for _, key := range roleKeys {
		roleSet[key] = struct{}{}
	}
	for _, key := range wsKeys {
		if _, ok := roleSet[key]; ok {
			resolved.Keys[key] = struct{}{}
		}
	}
	return resolved, nil
}

func (e *gormEvaluator) Can(ctx context.Context, accountID, workspaceID uuid.UUID, key string) (bool, error) {
	if key == "" {
		return false, errors.New("evaluator: permission key is required")
	}
	resolved, err := e.Resolve(ctx, accountID, workspaceID)
	if err != nil {
		return false, err
	}
	return resolved.Has(key), nil
}

func (e *gormEvaluator) Explain(ctx context.Context, accountID, workspaceID uuid.UUID) (*Explanation, error) {
	resolved, err := e.Resolve(ctx, accountID, workspaceID)
	if err != nil {
		return nil, err
	}
	bySource, err := e.queryFeaturePackageKeysBySource(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluator: explain: %w", err)
	}
	roleSources, err := e.queryRoleKeysBySource(ctx, accountID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluator: explain roles: %w", err)
	}
	return &Explanation{
		Resolved:           resolved,
		FeaturePackageKeys: bySource,
		RoleKeys:           roleSources,
	}, nil
}

// isOwnerOrAdmin returns true if the account is a workspace_members row with
// member_type owner or admin in the target workspace.
func (e *gormEvaluator) isOwnerOrAdmin(ctx context.Context, accountID, workspaceID uuid.UUID) (bool, error) {
	const q = `
SELECT member_type FROM workspace_members
WHERE workspace_id = ? AND user_id = ? AND deleted_at IS NULL
LIMIT 1
`
	var memberType string
	if err := e.db.WithContext(ctx).Raw(q, workspaceID, accountID).Scan(&memberType).Error; err != nil {
		return false, err
	}
	return memberType == "owner" || memberType == "admin", nil
}

// queryRoleKeys returns permission_key strings derived from the user's roles
// in the target workspace. Roles are joined via the legacy
// workspaces.collaboration_workspace_id back-reference (which is the only
// link we have until we introduce workspace_member_roles in a later phase).
// For personal workspaces (no collaboration_workspace_id) we look for
// account-level roles where user_roles.collaboration_workspace_id IS NULL.
// role_disabled_actions are subtracted from the final set.
func (e *gormEvaluator) queryRoleKeys(ctx context.Context, accountID, workspaceID uuid.UUID) ([]string, error) {
	// Bundle-aware: follows feature_package_bundles so that roles bound to
	// a parent bundle package also inherit keys from its child packages.
	const q = `
WITH ws AS (
  SELECT collaboration_workspace_id FROM workspaces WHERE id = ? AND deleted_at IS NULL
),
user_role_ids AS (
  SELECT ur.role_id
  FROM user_roles ur, ws
  WHERE ur.user_id = ?
    AND (
      (ws.collaboration_workspace_id IS NOT NULL AND ur.collaboration_workspace_id = ws.collaboration_workspace_id)
      OR
      (ws.collaboration_workspace_id IS NULL AND ur.collaboration_workspace_id IS NULL)
    )
),
pkg_keys AS (
  SELECT fpk.package_id, fpk.action_id FROM feature_package_keys fpk
  UNION
  SELECT fpb.package_id, fpk.action_id
  FROM feature_package_bundles fpb
  JOIN feature_package_keys fpk ON fpk.package_id = fpb.child_package_id
),
granted AS (
  SELECT DISTINCT pk.permission_key, urr.role_id, pkeys.action_id
  FROM user_role_ids urr
  JOIN role_feature_packages rfp ON rfp.role_id = urr.role_id AND rfp.enabled = true
  JOIN pkg_keys pkeys             ON pkeys.package_id = rfp.package_id
  JOIN permission_keys pk        ON pk.id = pkeys.action_id
  WHERE pk.deleted_at IS NULL
    AND pk.permission_key <> ''
)
SELECT DISTINCT g.permission_key
FROM granted g
WHERE NOT EXISTS (
  SELECT 1 FROM role_disabled_actions rda
  WHERE rda.role_id = g.role_id AND rda.action_id = g.action_id
)
`
	var out []string
	if err := e.db.WithContext(ctx).Raw(q, workspaceID, accountID).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (e *gormEvaluator) queryRoleKeysBySource(ctx context.Context, accountID, workspaceID uuid.UUID) (map[string][]uuid.UUID, error) {
	const q = `
WITH ws AS (
  SELECT collaboration_workspace_id FROM workspaces WHERE id = ? AND deleted_at IS NULL
),
user_role_ids AS (
  SELECT ur.role_id
  FROM user_roles ur, ws
  WHERE ur.user_id = ?
    AND (
      (ws.collaboration_workspace_id IS NOT NULL AND ur.collaboration_workspace_id = ws.collaboration_workspace_id)
      OR
      (ws.collaboration_workspace_id IS NULL AND ur.collaboration_workspace_id IS NULL)
    )
),
pkg_keys AS (
  SELECT fpk.package_id, fpk.action_id FROM feature_package_keys fpk
  UNION
  SELECT fpb.package_id, fpk.action_id
  FROM feature_package_bundles fpb
  JOIN feature_package_keys fpk ON fpk.package_id = fpb.child_package_id
),
granted AS (
  SELECT pk.permission_key, urr.role_id, pkeys.action_id
  FROM user_role_ids urr
  JOIN role_feature_packages rfp ON rfp.role_id = urr.role_id AND rfp.enabled = true
  JOIN pkg_keys pkeys             ON pkeys.package_id = rfp.package_id
  JOIN permission_keys pk        ON pk.id = pkeys.action_id
  WHERE pk.deleted_at IS NULL
    AND pk.permission_key <> ''
)
SELECT DISTINCT g.permission_key, g.role_id
FROM granted g
WHERE NOT EXISTS (
  SELECT 1 FROM role_disabled_actions rda
  WHERE rda.role_id = g.role_id AND rda.action_id = g.action_id
)
`
	type row struct {
		PermissionKey string    `gorm:"column:permission_key"`
		RoleID        uuid.UUID `gorm:"column:role_id"`
	}
	var rows []row
	if err := e.db.WithContext(ctx).Raw(q, workspaceID, accountID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string][]uuid.UUID, len(rows))
	for _, r := range rows {
		out[r.PermissionKey] = append(out[r.PermissionKey], r.RoleID)
	}
	return out, nil
}

// isSuperAdmin reports whether the given user has the global super_admin flag.
func (e *gormEvaluator) isSuperAdmin(ctx context.Context, accountID uuid.UUID) (bool, error) {
	if accountID == uuid.Nil {
		return false, nil
	}
	const q = `SELECT is_super_admin FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1`
	var flag bool
	if err := e.db.WithContext(ctx).Raw(q, accountID).Scan(&flag).Error; err != nil {
		return false, err
	}
	return flag, nil
}

func (e *gormEvaluator) queryAllPermissionKeys(ctx context.Context) ([]string, error) {
	const q = `SELECT permission_key FROM permission_keys WHERE deleted_at IS NULL AND permission_key <> ''`
	var out []string
	if err := e.db.WithContext(ctx).Raw(q).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// queryAccountUnionKeys returns the union of resolved permission keys across
// every workspace the account is a member of. Used for account-only ops where
// no specific workspace is bound.
func (e *gormEvaluator) queryAccountUnionKeys(ctx context.Context, accountID uuid.UUID) ([]string, error) {
	var workspaceIDs []uuid.UUID
	if err := e.db.WithContext(ctx).
		Raw(`SELECT workspace_id FROM workspace_members WHERE user_id = ? AND deleted_at IS NULL`, accountID).
		Scan(&workspaceIDs).Error; err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, wsID := range workspaceIDs {
		r, err := e.Resolve(ctx, accountID, wsID)
		if err != nil {
			return nil, err
		}
		for k := range r.Keys {
			seen[k] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out, nil
}

func (e *gormEvaluator) queryFeaturePackageKeys(ctx context.Context, workspaceID uuid.UUID) ([]string, error) {
	// Resolves keys both from directly-bound packages and from bundle children
	// (feature_package_bundles links a parent bundle package to its child packages).
	const q = `
SELECT DISTINCT pk.permission_key
FROM workspace_feature_packages wfp
JOIN (
  SELECT fpk.package_id, fpk.action_id FROM feature_package_keys fpk
  UNION
  SELECT fpb.package_id, fpk.action_id
  FROM feature_package_bundles fpb
  JOIN feature_package_keys fpk ON fpk.package_id = fpb.child_package_id
) resolved ON resolved.package_id = wfp.package_id
JOIN permission_keys pk ON pk.id = resolved.action_id
WHERE wfp.workspace_id = ?
  AND wfp.enabled = true
  AND wfp.deleted_at IS NULL
  AND pk.deleted_at IS NULL
  AND pk.permission_key <> ''
`
	var out []string
	if err := e.db.WithContext(ctx).Raw(q, workspaceID).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (e *gormEvaluator) queryFeaturePackageKeysBySource(ctx context.Context, workspaceID uuid.UUID) (map[string][]uuid.UUID, error) {
	const q = `
SELECT pk.permission_key AS permission_key, wfp.package_id AS package_id
FROM workspace_feature_packages wfp
JOIN (
  SELECT fpk.package_id, fpk.action_id FROM feature_package_keys fpk
  UNION
  SELECT fpb.package_id, fpk.action_id
  FROM feature_package_bundles fpb
  JOIN feature_package_keys fpk ON fpk.package_id = fpb.child_package_id
) resolved ON resolved.package_id = wfp.package_id
JOIN permission_keys pk ON pk.id = resolved.action_id
WHERE wfp.workspace_id = ?
  AND wfp.enabled = true
  AND wfp.deleted_at IS NULL
  AND pk.deleted_at IS NULL
  AND pk.permission_key <> ''
`
	type row struct {
		PermissionKey string    `gorm:"column:permission_key"`
		PackageID     uuid.UUID `gorm:"column:package_id"`
	}
	var rows []row
	if err := e.db.WithContext(ctx).Raw(q, workspaceID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string][]uuid.UUID, len(rows))
	for _, r := range rows {
		out[r.PermissionKey] = append(out[r.PermissionKey], r.PackageID)
	}
	return out, nil
}
