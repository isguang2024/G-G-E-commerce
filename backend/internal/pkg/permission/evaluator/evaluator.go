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
	if workspaceID == uuid.Nil {
		return nil, errors.New("evaluator: workspace id is required")
	}

	resolved := &ResolvedPermissions{
		AccountID:   accountID,
		WorkspaceID: workspaceID,
		Keys:        make(map[string]struct{}),
	}

	// 1. Feature-package side of the intersection.
	keys, err := e.queryFeaturePackageKeys(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluator: load feature package keys: %w", err)
	}
	for _, key := range keys {
		resolved.Keys[key] = struct{}{}
	}

	// 2. Role side of the intersection.
	// TODO(phase-3-followup): wire workspace_role_bindings → role permissions
	// once roles consistently expose permission_key strings rather than
	// action ids. For now we treat roles as "all-or-nothing" so the package
	// surface returned above is the effective permission set.

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
	return &Explanation{
		Resolved:           resolved,
		FeaturePackageKeys: bySource,
		RoleKeys:           map[string][]uuid.UUID{},
	}, nil
}

func (e *gormEvaluator) queryFeaturePackageKeys(ctx context.Context, workspaceID uuid.UUID) ([]string, error) {
	const q = `
SELECT DISTINCT pk.permission_key
FROM workspace_feature_packages wfp
JOIN feature_package_keys fpk ON fpk.package_id = wfp.package_id
JOIN permission_keys pk        ON pk.id = fpk.action_id
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
JOIN feature_package_keys fpk ON fpk.package_id = wfp.package_id
JOIN permission_keys pk        ON pk.id = fpk.action_id
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
