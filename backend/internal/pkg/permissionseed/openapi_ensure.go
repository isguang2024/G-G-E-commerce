// openapi_ensure.go — idempotent runtime upsert of every permission_key
// referenced by the OpenAPI spec. Runs at startup (after goose migrations
// and after EnsureDefaultPermissionKeys) so any operationId added to
// openapi.yaml automatically materialises a permission_keys row without
// requiring a hand-written goose migration each time.
//
// The matching baseline goose migration (00002_permission_seed_baseline.sql)
// pre-creates the well-known keys that the previous manual hot-fix injected
// (workspace.read, user.list/create/update/delete/read, workspace.switch …)
// so that a freshly-reset DB is usable even before the Go process boots.
package permissionseed

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// EnsureOpenAPIPermissionKeys upserts every distinct permission_key found in
// the embedded openapi_seed.json. Existing rows are left untouched (we only
// fill in name/description/status when missing) so that operator edits made
// in the admin UI are not clobbered. Returns the number of newly-inserted
// rows for logging.
func EnsureOpenAPIPermissionKeys(db *gorm.DB) (int, error) {
	if db == nil {
		return 0, errors.New("permissionseed: nil db")
	}
	seed, err := LoadOpenAPISeed()
	if err != nil {
		return 0, err
	}
	seen := make(map[string]struct{}, len(seed.Operations))
	created := 0
	now := time.Now()
	for _, op := range seed.Operations {
		key := strings.TrimSpace(op.PermissionKey)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		var existing systemmodels.PermissionKey
		err := db.Where("permission_key = ?", key).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return created, err
		}
		row := systemmodels.PermissionKey{
			ID:                    uuid.New(),
			Code:                  StableID("openapi-permission-key", key).String(),
			PermissionKey:         key,
			AppKey:                systemmodels.DefaultAppKey,
			ModuleCode:            deriveModuleCode(key),
			ContextType:           "common",
			FeatureKind:           "system",
			DataPolicy:            "none",
			AllowedWorkspaceTypes: "personal,collaboration",
			Name:                  deriveDisplayName(op.Summary, key),
			Description:           op.Summary,
			Status:                "normal",
			IsBuiltin:             true,
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		if err := db.Create(&row).Error; err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func deriveModuleCode(key string) string {
	if idx := strings.Index(key, "."); idx > 0 {
		return key[:idx]
	}
	return key
}

func deriveDisplayName(summary, key string) string {
	if s := strings.TrimSpace(summary); s != "" {
		return s
	}
	return key
}
