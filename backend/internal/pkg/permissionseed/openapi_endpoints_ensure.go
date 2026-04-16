// openapi_endpoints_ensure.go — idempotent startup upsert of api_endpoints
// and api_endpoint_permission_bindings derived from openapi_seed.json.
//
// Design goals:
//   - Single source of truth: all data comes from openapi.yaml via the seed.
//   - Deterministic IDs: StableID("openapi-api-endpoint", METHOD+" "+path)
//     produces consistent UUIDv5 across environments and DB resets.
//   - Non-destructive: existing rows are only patched when the operator has
//     not customised them (summary/description/category_id filled in from
//     seed only when the column is currently empty).
//   - Stale marking: endpoints / bindings that the seed no longer declares
//     are marked stale (endpoints: status='stale'; bindings: soft-deleted).
//     Readded operations are revived back to status='normal' — this keeps
//     the /api-endpoints/sync UI meaningful when operators rename or delete
//     operations in openapi.yaml.
//   - Category resolution order:
//       1. x-api-category extension in openapi.yaml
//       2. First tag on the operation
//       3. First segment of x-permission-key (e.g. "user" from "user.list")
//       4. First path segment after /api/vN/ prefix
//       5. Hard fallback: "uncategorized"
//
// Call order in cmd/migrate (after EnsureOpenAPIPermissionKeys):
//
//	EnsureOpenAPIEndpoints(db)
//	EnsureOpenAPIPermissionBindings(db)
package permissionseed

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	systemmodels "github.com/maben/backend/internal/modules/system/models"
)

// endpointStatusNormal / endpointStatusStale are the two statuses the ensure
// pipeline writes. Operators may also set custom statuses through the admin
// UI (e.g. "disabled"); ensure never overrides those — it only toggles
// between normal ↔ stale for seed-managed rows.
const (
	endpointStatusNormal = "normal"
	endpointStatusStale  = "stale"
)

// EnsureOpenAPIEndpoints upserts one api_endpoints row per operation in the
// embedded seed, and marks any pre-existing row no longer referenced by the
// seed as status='stale'. Returns the number of newly inserted rows.
//
// The whole upsert + stale sweep runs in a single DB transaction so a mid-
// boot crash cannot leave half the table stale and the other half normal.
func EnsureOpenAPIEndpoints(db *gorm.DB) (int, error) {
	if db == nil {
		return 0, errors.New("permissionseed: nil db")
	}
	seed, err := LoadOpenAPISeed()
	if err != nil {
		return 0, err
	}

	// Pre-load category map once (read-only, small). Done outside the tx so
	// the write path stays short.
	categoryByCode, err := loadCategoryMap(db)
	if err != nil {
		return 0, err
	}

	created := 0
	now := time.Now()
	seedCodes := make(map[string]struct{}, len(seed.Operations))

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, op := range seed.Operations {
			method := strings.ToUpper(strings.TrimSpace(op.Method))
			path := strings.TrimSpace(op.Path)
			if method == "" || path == "" {
				continue
			}

			stableID := StableID("openapi-api-endpoint", method+" "+path)
			stableCode := stableID.String()
			seedCodes[stableCode] = struct{}{}
			categoryID := resolveOperationCategory(op, categoryByCode)

			summary := strings.TrimSpace(op.Summary)
			description := strings.TrimSpace(op.Description)
			if description == "" {
				description = summary
			}
			// Derive a human-readable name: prefer summary, fall back to operationId.
			name := summary
			if name == "" {
				name = op.OperationID
			}
			if name == "" {
				name = method + " " + path
			}

			// Try to find existing row by stable code first, then by method+path.
			var existing systemmodels.APIEndpoint
			err := tx.Where("code = ?", stableCode).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = tx.Where("method = ? AND path = ? AND deleted_at IS NULL",
					method, path).First(&existing).Error
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Fallback for historical data: a soft-deleted row may still occupy
				// the deterministic UUID primary key and cause duplicate pkey errors.
				err = tx.Unscoped().
					Where("id = ? OR code = ? OR (method = ? AND path = ?)",
						stableID, stableCode, method, path).
					First(&existing).Error
			}

			if err == nil {
				// Row exists — back-fill empty fields only (preserve operator edits).
				updates := map[string]interface{}{"updated_at": now}
				if existing.DeletedAt.Valid {
					// Revive historical soft-deleted rows to keep startup idempotent.
					updates["deleted_at"] = nil
					updates["status"] = endpointStatusNormal
				} else if existing.Status == endpointStatusStale {
					// Operation disappeared in a prior seed but is back — clear stale.
					updates["status"] = endpointStatusNormal
				}
				if existing.ID == stableID {
					// Deterministic ID row should stay aligned with its source operation.
					updates["method"] = method
					updates["path"] = path
				}
				if strings.TrimSpace(existing.Code) == "" {
					updates["code"] = stableCode
				}
				if strings.TrimSpace(existing.Summary) == "" && name != "" {
					updates["summary"] = name
				}
				if existing.CategoryID == nil && categoryID != nil {
					updates["category_id"] = categoryID
				}
				if len(updates) > 1 { // more than just updated_at
					if err := tx.Unscoped().Model(&existing).Updates(updates).Error; err != nil {
						return err
					}
				}
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			// Insert new row.
			row := systemmodels.APIEndpoint{
				ID:         stableID,
				Code:       stableCode,
				Method:     method,
				Path:       path,
				Summary:    name,
				CategoryID: categoryID,
				Status:     endpointStatusNormal,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			created++
		}

		return markStaleEndpoints(tx, seedCodes, now)
	}); err != nil {
		return created, err
	}
	return created, nil
}

// markStaleEndpoints flips any currently-live, non-stale endpoint row whose
// code is not in `seedCodes` to status='stale'. The WHERE filter (`status <>
// 'stale'`) guarantees repeat calls are idempotent — only rows whose state
// actually changes receive an UPDATE.
func markStaleEndpoints(tx *gorm.DB, seedCodes map[string]struct{}, now time.Time) error {
	var rows []systemmodels.APIEndpoint
	if err := tx.
		Select("id", "code").
		Where("deleted_at IS NULL AND status <> ?", endpointStatusStale).
		Find(&rows).Error; err != nil {
		return err
	}
	var staleIDs []uuid.UUID
	for _, r := range rows {
		if _, ok := seedCodes[r.Code]; ok {
			continue
		}
		staleIDs = append(staleIDs, r.ID)
	}
	if len(staleIDs) == 0 {
		return nil
	}
	return tx.Model(&systemmodels.APIEndpoint{}).
		Where("id IN ?", staleIDs).
		Updates(map[string]interface{}{
			"status":     endpointStatusStale,
			"updated_at": now,
		}).Error
}

// EnsureOpenAPIPermissionBindings upserts api_endpoint_permission_bindings
// linking each endpoint to its permission key, and soft-deletes any binding
// the seed no longer declares. Safe to call repeatedly. Returns the number
// of newly inserted bindings.
//
// Startup integrity guarantee: every seed-referenced permission_key MUST
// already exist in permission_keys before bindings are written. Missing keys
// return ErrUnknownPermissionKeys so cmd/migrate fails loudly, preventing
// a silent permission-denied at runtime.
func EnsureOpenAPIPermissionBindings(db *gorm.DB) (int, error) {
	if db == nil {
		return 0, errors.New("permissionseed: nil db")
	}
	seed, err := LoadOpenAPISeed()
	if err != nil {
		return 0, err
	}

	if missing, err := findUnknownPermissionKeys(db, seed); err != nil {
		return 0, err
	} else if len(missing) > 0 {
		return 0, &ErrUnknownPermissionKeys{Keys: missing}
	}

	created := 0
	now := time.Now()
	seedBindings := make(map[string]struct{}, len(seed.Operations))

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, op := range seed.Operations {
			permKey := strings.TrimSpace(op.PermissionKey)
			if permKey == "" {
				continue
			}
			method := strings.ToUpper(strings.TrimSpace(op.Method))
			path := strings.TrimSpace(op.Path)
			if method == "" || path == "" {
				continue
			}

			endpointCode := StableID("openapi-api-endpoint", method+" "+path).String()
			seedBindings[bindingKey(endpointCode, permKey)] = struct{}{}

			// Check if binding already exists.
			var binding systemmodels.APIEndpointPermissionBinding
			err := tx.Where("endpoint_code = ? AND permission_key = ?", endpointCode, permKey).
				First(&binding).Error
			if err == nil {
				continue // already registered
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			stableBindingID := StableID("openapi-endpoint-binding", endpointCode+":"+permKey)
			// Fallback for historical soft-deleted rows with the same deterministic ID.
			if err := tx.Unscoped().Where("id = ? OR (endpoint_code = ? AND permission_key = ?)",
				stableBindingID, endpointCode, permKey).First(&binding).Error; err == nil {
				updates := map[string]interface{}{
					"updated_at": now,
				}
				if binding.DeletedAt.Valid {
					updates["deleted_at"] = nil
				}
				if strings.TrimSpace(binding.MatchMode) == "" {
					updates["match_mode"] = "ANY"
				}
				if len(updates) > 1 {
					if err := tx.Unscoped().Model(&binding).Updates(updates).Error; err != nil {
						return err
					}
				}
				continue
			}

			// Ensure the endpoint row exists (created by EnsureOpenAPIEndpoints).
			var count int64
			if err := tx.Model(&systemmodels.APIEndpoint{}).
				Where("code = ?", endpointCode).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				continue // endpoint was not inserted (shouldn't happen in normal flow)
			}

			row := systemmodels.APIEndpointPermissionBinding{
				ID:            stableBindingID,
				EndpointCode:  endpointCode,
				PermissionKey: permKey,
				MatchMode:     "ANY",
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			created++
		}

		return pruneOrphanedBindings(tx, seedBindings)
	}); err != nil {
		return created, err
	}
	return created, nil
}

// pruneOrphanedBindings soft-deletes any currently-live binding whose
// (endpoint_code, permission_key) pair is not present in `seedBindings`.
// GORM's soft-delete semantics (DeletedAt column) mean this is reversible —
// the binding row is kept for audit history, just hidden from the evaluator.
func pruneOrphanedBindings(tx *gorm.DB, seedBindings map[string]struct{}) error {
	var rows []systemmodels.APIEndpointPermissionBinding
	if err := tx.
		Select("id", "endpoint_code", "permission_key").
		Find(&rows).Error; err != nil {
		return err
	}
	var orphanIDs []uuid.UUID
	for _, r := range rows {
		if _, ok := seedBindings[bindingKey(r.EndpointCode, r.PermissionKey)]; ok {
			continue
		}
		orphanIDs = append(orphanIDs, r.ID)
	}
	if len(orphanIDs) == 0 {
		return nil
	}
	return tx.
		Where("id IN ?", orphanIDs).
		Delete(&systemmodels.APIEndpointPermissionBinding{}).Error
}

// bindingKey is the in-memory tuple key used by pruneOrphanedBindings. The
// separator '|' never appears in a valid UUID or permission_key, so there's
// no collision risk between distinct pairs.
func bindingKey(endpointCode, permKey string) string {
	return endpointCode + "|" + permKey
}

// resolveOperationCategory maps an operation to an api_endpoint_categories.id.
// Resolution order: x-api-category → first tag → permission_key prefix →
// path segment → "uncategorized".
func resolveOperationCategory(op OpenAPIOperation, categoryByCode map[string]uuid.UUID) *uuid.UUID {
	candidates := []string{
		strings.TrimSpace(op.APICategory),
	}
	// permission_key prefix (e.g. "user" from "user.list")
	if pk := strings.TrimSpace(op.PermissionKey); pk != "" {
		if idx := strings.Index(pk, "."); idx > 0 {
			candidates = append(candidates, pk[:idx])
		}
	}
	// path segment after /api/vN/
	candidates = append(candidates, pathCategoryHint(op.Path))
	candidates = append(candidates, "uncategorized")

	for _, c := range candidates {
		c = strings.TrimSpace(strings.ToLower(c))
		if c == "" {
			continue
		}
		if id, ok := categoryByCode[c]; ok {
			cp := id
			return &cp
		}
	}
	return nil
}

// pathSegmentCategoryAlias maps a *raw* first-meaningful path segment to its
// api_endpoint_categories.code. This replaces the previous blind TrimSuffix("s")
// hack, which turned "dictionaries" into "dictionarie" and "observability" into
// "observabilitie" (both misses in categoryByCode, so those operations fell
// through to "uncategorized"). Entries here must match a row that exists — or
// is about to exist — in api_endpoint_categories.
//
// When adding a new `/api/v1/{segment}` domain, add both:
//   - the alias entry below (if the segment is plural / dash-joined / differs
//     from the category code), and
//   - the api_endpoint_categories seed row in cmd/migrate seeds.
var pathSegmentCategoryAlias = map[string]string{
	// plural → singular
	"users":            "user",
	"roles":            "role",
	"menus":            "menu",
	"pages":            "page",
	"workspaces":       "workspace",
	"messages":         "message",
	"dictionaries":     "dictionary",
	"permissions":      "permission_key",
	// dash-joined → snake_case
	"feature-packages":         "feature_package",
	"collaboration-workspaces": "collaboration_workspace",
	"api-endpoints":            "api_endpoint",
	"site-configs":             "site_config",
	// domain-level aliases (co-located paths share one bucket)
	"observability": "observability",
	"telemetry":     "observability",
	// identity / pass-through (no trailing 's' to strip anyway)
	"media":      "media",
	"auth":       "auth",
	"system":     "system",
	"navigation": "navigation",
}

// pathCategoryHint picks the first meaningful path segment and translates it
// through pathSegmentCategoryAlias. Non-aliased segments are returned verbatim
// so resolveOperationCategory still gets a chance (a future category whose
// code == the raw segment will match directly). No more silent `s`-stripping.
//
// Examples:
//
//	/api/v1/users/123               → "user"
//	/workspaces/{id}                → "workspace"
//	/observability/audit-logs       → "observability"
//	/dictionaries/by-codes          → "dictionary"
//	/brand-new-thing                → "brand-new-thing" (unmapped; let
//	                                   resolveOperationCategory handle it)
func pathCategoryHint(rawPath string) string {
	trimmed := strings.Trim(strings.TrimSpace(rawPath), "/")
	if trimmed == "" {
		return ""
	}
	for _, seg := range strings.Split(trimmed, "/") {
		seg = strings.TrimSpace(seg)
		if seg == "" || seg == "api" || seg == "v1" || seg == "v2" || seg == "v3" {
			continue
		}
		if strings.HasPrefix(seg, "{") {
			continue
		}
		if mapped, ok := pathSegmentCategoryAlias[seg]; ok {
			return mapped
		}
		return seg
	}
	return ""
}

// ErrUnknownPermissionKeys is returned by EnsureOpenAPIPermissionBindings
// when the seed references one or more permission_keys that do not exist in
// the permission_keys table. Startup code (cmd/migrate) must surface this as
// a fatal error — continuing would bind requests to keys the evaluator can
// never resolve, which would manifest as silent 403s in production.
type ErrUnknownPermissionKeys struct {
	Keys []string
}

func (e *ErrUnknownPermissionKeys) Error() string {
	return "openapi seed references permission_keys missing from permission_keys table: " + strings.Join(e.Keys, ", ")
}

// findUnknownPermissionKeys collects every distinct permission_key referenced
// by the seed and returns the subset absent from the permission_keys table.
// Soft-deleted rows are treated as absent — a deleted key cannot authorise a
// request and must be re-seeded before bindings point at it.
func findUnknownPermissionKeys(db *gorm.DB, seed *OpenAPISeed) ([]string, error) {
	seedKeys := make(map[string]struct{}, len(seed.Operations))
	for _, op := range seed.Operations {
		k := strings.TrimSpace(op.PermissionKey)
		if k == "" {
			continue
		}
		seedKeys[k] = struct{}{}
	}
	if len(seedKeys) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(seedKeys))
	for k := range seedKeys {
		keys = append(keys, k)
	}

	var present []string
	if err := db.Table("permission_keys").
		Where("permission_key IN ? AND deleted_at IS NULL", keys).
		Pluck("permission_key", &present).Error; err != nil {
		return nil, err
	}
	presentSet := make(map[string]struct{}, len(present))
	for _, k := range present {
		presentSet[k] = struct{}{}
	}

	var missing []string
	for k := range seedKeys {
		if _, ok := presentSet[k]; !ok {
			missing = append(missing, k)
		}
	}
	// Deterministic order so log output is diff-stable across boots.
	sortStrings(missing)
	return missing, nil
}

// ScanOrphanedPermissionKeys returns permission_keys rows that are not
// referenced by any api_endpoint_permission_bindings row. These are typically
// legacy hand-managed keys that have no OpenAPI operation pointing to them.
// Call at the tail of cmd/migrate and log at Warn level — orphan keys do not
// block startup (operators may intentionally keep them for non-HTTP checks)
// but are almost always cleanup candidates.
func ScanOrphanedPermissionKeys(db *gorm.DB) ([]string, error) {
	if db == nil {
		return nil, errors.New("permissionseed: nil db")
	}
	var orphans []string
	// Left-anti-join via NOT EXISTS keeps the query index-friendly; we also
	// skip IsBuiltin baseline keys so the warn list doesn't drown in seed
	// rows that intentionally have no binding (e.g. menu/page permission
	// scopes that are enforced in other layers).
	if err := db.Table("permission_keys AS pk").
		Where("pk.deleted_at IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM api_endpoint_permission_bindings b WHERE b.permission_key = pk.permission_key AND b.deleted_at IS NULL)").
		Where("pk.is_builtin = ?", false).
		Order("pk.permission_key").
		Pluck("pk.permission_key", &orphans).Error; err != nil {
		return nil, err
	}
	return orphans, nil
}

// sortStrings is a local helper so findUnknownPermissionKeys does not have
// to import "sort" in every compile unit. Keeps the diff narrow.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

// loadCategoryMap fetches all non-deleted api_endpoint_categories rows and
// returns a code → ID map for fast lookup.
func loadCategoryMap(db *gorm.DB) (map[string]uuid.UUID, error) {
	var rows []struct {
		Code string
		ID   uuid.UUID
	}
	if err := db.Table("api_endpoint_categories").
		Select("code, id").
		Where("deleted_at IS NULL").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]uuid.UUID, len(rows))
	for _, r := range rows {
		m[strings.ToLower(strings.TrimSpace(r.Code))] = r.ID
	}
	return m, nil
}
