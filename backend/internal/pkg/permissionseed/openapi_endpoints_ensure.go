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

// EnsureOpenAPIEndpoints upserts one api_endpoints row per operation in the
// embedded seed. Returns the number of newly inserted rows.
func EnsureOpenAPIEndpoints(db *gorm.DB) (int, error) {
	if db == nil {
		return 0, errors.New("permissionseed: nil db")
	}
	seed, err := LoadOpenAPISeed()
	if err != nil {
		return 0, err
	}

	// Pre-load category map once: code → ID.
	categoryByCode, err := loadCategoryMap(db)
	if err != nil {
		return 0, err
	}

	created := 0
	now := time.Now()

	for _, op := range seed.Operations {
		method := strings.ToUpper(strings.TrimSpace(op.Method))
		path := strings.TrimSpace(op.Path)
		if method == "" || path == "" {
			continue
		}

		stableID := StableID("openapi-api-endpoint", method+" "+path)
		stableCode := stableID.String()
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
		err := db.Where("code = ?", stableCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = db.Where("method = ? AND path = ? AND deleted_at IS NULL",
				method, path).First(&existing).Error
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback for historical data: a soft-deleted row may still occupy
			// the deterministic UUID primary key and cause duplicate pkey errors.
			err = db.Unscoped().
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
				updates["status"] = "normal"
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
				if err := db.Model(&existing).Updates(updates).Error; err != nil {
					return created, err
				}
			}
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return created, err
		}

		// Insert new row.
		row := systemmodels.APIEndpoint{
			ID:         stableID,
			Code:       stableCode,
			Method:     method,
			Path:       path,
			Summary:    name,
			CategoryID: categoryID,
			Status:     "normal",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := db.Create(&row).Error; err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

// EnsureOpenAPIPermissionBindings upserts api_endpoint_permission_bindings
// linking each endpoint to its permission key. Safe to call repeatedly.
// Returns the number of newly inserted bindings.
func EnsureOpenAPIPermissionBindings(db *gorm.DB) (int, error) {
	if db == nil {
		return 0, errors.New("permissionseed: nil db")
	}
	seed, err := LoadOpenAPISeed()
	if err != nil {
		return 0, err
	}

	created := 0
	now := time.Now()

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

		// Check if binding already exists.
		var binding systemmodels.APIEndpointPermissionBinding
		err := db.Where("endpoint_code = ? AND permission_key = ?", endpointCode, permKey).
			First(&binding).Error
		if err == nil {
			continue // already registered
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return created, err
		}

		stableBindingID := StableID("openapi-endpoint-binding", endpointCode+":"+permKey)
		// Fallback for historical soft-deleted rows with the same deterministic ID.
		if err := db.Unscoped().Where("id = ? OR (endpoint_code = ? AND permission_key = ?)",
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
				if err := db.Unscoped().Model(&binding).Updates(updates).Error; err != nil {
					return created, err
				}
			}
			continue
		}

		// Ensure the endpoint row exists (created by EnsureOpenAPIEndpoints).
		var count int64
		if err := db.Model(&systemmodels.APIEndpoint{}).
			Where("code = ?", endpointCode).Count(&count).Error; err != nil {
			return created, err
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
		if err := db.Create(&row).Error; err != nil {
			return created, err
		}
		created++
	}
	return created, nil
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

// pathCategoryHint extracts the first meaningful path segment.
// "/api/v1/users/123" → "user"; "/workspaces/{id}" → "workspace".
func pathCategoryHint(rawPath string) string {
	trimmed := strings.Trim(strings.TrimSpace(rawPath), "/")
	if trimmed == "" {
		return ""
	}
	segments := strings.Split(trimmed, "/")
	// Skip common API version prefixes.
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" || seg == "api" || seg == "v1" || seg == "v2" || seg == "v3" {
			continue
		}
		// Strip trailing 's' for plurals and parameter segments.
		if strings.HasPrefix(seg, "{") {
			continue
		}
		return strings.TrimSuffix(seg, "s")
	}
	return ""
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


