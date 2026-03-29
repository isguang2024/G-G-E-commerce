package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	apirouter "github.com/gg-ecommerce/backend/internal/api/router"
	"github.com/gg-ecommerce/backend/internal/config"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting database migration...")

	_, err = database.Init(&cfg.DB)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	logger.Info("Database connected successfully")

	if err := preparePermissionTableRenames(logger); err != nil {
		logger.Fatal("Failed to prepare permission table renames", zap.Error(err))
	}

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}

	logger.Info("Database migration completed successfully!")

	// 初始化默认作用域
	if err := runNamedMigrations(logger); err != nil {
		logger.Fatal("Named migrations failed", zap.Error(err))
	}

	// 初始化默认角色
	if err := initDefaultRolesNoScope(logger); err != nil {
		logger.Warn("Failed to initialize default roles", zap.Error(err))
	} else {
		logger.Info("Default roles initialized successfully")
	}

	// 初始化默认管理员
	if err := initDefaultAdmin(logger); err != nil {
		logger.Warn("Failed to initialize default admin", zap.Error(err))
	} else {
		logger.Info("Default admin initialized successfully")
	}

	// 初始化默认菜单
	if err := initDefaultMenusNoScope(logger); err != nil {
		logger.Warn("Failed to initialize default menus", zap.Error(err))
	} else {
		logger.Info("Default menus initialized successfully")
	}

	if err := initDefaultPermissionGroups(logger); err != nil {
		logger.Warn("Failed to initialize default permission groups", zap.Error(err))
	} else {
		logger.Info("Default permission groups initialized successfully")
	}

	if err := initDefaultPermissionKeysNoScope(logger); err != nil {
		logger.Warn("Failed to initialize permission keys", zap.Error(err))
	} else {
		logger.Info("Permission keys initialized successfully")
	}

	if err := initDefaultFeaturePackages(logger); err != nil {
		logger.Warn("Failed to initialize feature packages", zap.Error(err))
	} else {
		logger.Info("Feature packages initialized successfully")
	}

	if err := initDefaultFeaturePackageBundles(logger); err != nil {
		logger.Warn("Failed to initialize feature package bundles", zap.Error(err))
	} else {
		logger.Info("Feature package bundles initialized successfully")
	}

	if err := initDefaultRoleFeaturePackages(logger); err != nil {
		logger.Warn("Failed to initialize default role feature packages", zap.Error(err))
	} else {
		logger.Info("Default role feature packages initialized successfully")
	}

	if err := initDefaultAPIEndpointCategories(logger); err != nil {
		logger.Warn("Failed to initialize api endpoint categories", zap.Error(err))
	} else {
		logger.Info("API endpoint categories initialized successfully")
	}

	if err := finalizeAPIEndpointSchema(logger); err != nil {
		logger.Warn("Failed to finalize api endpoint schema", zap.Error(err))
	} else {
		logger.Info("API endpoint schema finalized successfully")
	}

	if err := syncAPIRegistry(logger, cfg); err != nil {
		logger.Warn("Failed to sync API registry", zap.Error(err))
	} else {
		logger.Info("API registry synchronized successfully")
	}

	if err := mergeDeprecatedTeamPermissionKeys(logger); err != nil {
		logger.Warn("Failed to merge deprecated team permission keys after API sync", zap.Error(err))
	}
	if err := syncCanonicalPermissionKeys(logger); err != nil {
		logger.Warn("Failed to sync canonical permission keys", zap.Error(err))
	} else {
		logger.Info("Canonical permission keys synchronized successfully")
	}

	if err := backfillTenantIdentityUserRoles(logger); err != nil {
		logger.Warn("Failed to backfill tenant identity user roles", zap.Error(err))
	} else {
		logger.Info("Tenant identity user roles backfilled successfully")
	}

	if err := refreshDefaultAccessSnapshots(logger); err != nil {
		logger.Warn("Failed to refresh default access snapshots", zap.Error(err))
	} else {
		logger.Info("Default access snapshots refreshed successfully")
	}

	logger.Info("Migration completed successfully")
}

type migrationTask struct {
	Name string
	Run  func(logger *zap.Logger) error
}

func runNamedMigrations(logger *zap.Logger) error {
	if err := ensureMigrationHistoryTable(); err != nil {
		return err
	}

	tasks := []migrationTask{
		{
			Name: "20260325_page_management_menu_seed",
			Run: func(logger *zap.Logger) error {
				if _, err := ensureDefaultMenuSeedByName("PageManagement"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260325_page_management_menu_seed"))
				return nil
			},
		},
		{
			Name: "20260325_page_management_menu_activate",
			Run: func(logger *zap.Logger) error {
				meta := usermodel.MetaJSON{
					"roles":     []interface{}{"R_SUPER"},
					"keepAlive": true,
				}
				if err := database.DB.Model(&usermodel.Menu{}).
					Where("name = ?", "PageManagement").
					Updates(map[string]interface{}{
						"path":       "page",
						"component":  "/system/page",
						"title":      "页面管理",
						"sort_order": 8,
						"meta":       meta,
					}).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260325_page_management_menu_activate"))
				return nil
			},
		},
		{
			Name: "20260325_menu_access_mode_jwt_defaults",
			Run: func(logger *zap.Logger) error {
				targetMenus := []string{"Dashboard", "Result", "Exception", "UserCenter"}
				for _, menuName := range targetMenus {
					if err := ensureMenuAccessMode(menuName, "jwt"); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260325_menu_access_mode_jwt_defaults"))
				return nil
			},
		},
		{
			Name: "20260326_menu_manage_groups_backfill",
			Run: func(logger *zap.Logger) error {
				if err := backfillMenuManageGroups(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260326_menu_manage_groups_backfill"))
				return nil
			},
		},
		{
			Name: "20260325_permission_endpoint_binding_ops",
			Run: func(logger *zap.Logger) error {
				permissionKey := "system.permission.manage"
				categoryID := permissionseed.StableID("api-endpoint-category", "permission_key")
				categoryIDPtr := &categoryID
				endpoints := []struct {
					Method  string
					Path    string
					Summary string
				}{
					{Method: "POST", Path: "/api/v1/permission-actions/:id/endpoints", Summary: "新增功能权限关联接口"},
					{Method: "DELETE", Path: "/api/v1/permission-actions/:id/endpoints/:endpointCode", Summary: "删除功能权限关联接口"},
				}

				return database.DB.Transaction(func(tx *gorm.DB) error {
					for _, item := range endpoints {
						method := strings.ToUpper(strings.TrimSpace(item.Method))
						path := strings.TrimSpace(item.Path)
						code := apiregistry.ResolveRouteCode(method, path, nil)
						if code == "" {
							code = permissionseed.StableID("api-endpoint-code", method+" "+path).String()
						}

						endpoint := &usermodel.APIEndpoint{
							Code:         code,
							Method:       method,
							Path:         path,
							FeatureKind:  "system",
							Summary:      item.Summary,
							CategoryID:   categoryIDPtr,
							ContextScope: "optional",
							Source:       "sync",
							Status:       "normal",
						}

						var existing usermodel.APIEndpoint
						err := tx.Where("method = ? AND path = ?", method, path).First(&existing).Error
						if err != nil {
							if errors.Is(err, gorm.ErrRecordNotFound) {
								if createErr := tx.Create(endpoint).Error; createErr != nil {
									return createErr
								}
								existing = *endpoint
							} else {
								return err
							}
						} else {
							if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
								"summary":       item.Summary,
								"feature_kind":  "system",
								"category_id":   categoryID,
								"context_scope": "optional",
								"source":        "sync",
								"status":        "normal",
							}).Error; updateErr != nil {
								return updateErr
							}
						}

						var count int64
						if countErr := tx.Model(&usermodel.APIEndpointPermissionBinding{}).
							Where("endpoint_code = ? AND permission_key = ?", existing.Code, permissionKey).
							Count(&count).Error; countErr != nil {
							return countErr
						}
						if count == 0 {
							if createBindErr := tx.Create(&usermodel.APIEndpointPermissionBinding{
								EndpointCode:  existing.Code,
								PermissionKey: permissionKey,
								MatchMode:     "ANY",
								SortOrder:     0,
							}).Error; createBindErr != nil {
								return createBindErr
							}
						}
					}
					return nil
				})
			},
		},
		{
			Name: "20260324_permission_key_code_alignment_v2",
			Run: func(logger *zap.Logger) error {
				oldCategoryID := permissionseed.StableID("api-endpoint-category", "permission_action")
				newCategoryID := permissionseed.StableID("api-endpoint-category", "permission_key")
				oldModuleGroupID := permissionseed.StableID("permission-group", "module:permission_action")
				newModuleGroupID := permissionseed.StableID("permission-group", "module:permission_key")
				systemFeatureGroupID := permissionseed.StableID("permission-group", "feature:system")

				return database.DB.Transaction(func(tx *gorm.DB) error {
					if err := tx.Model(&usermodel.PermissionKey{}).
						Where("module_code = ?", "permission_action").
						Updates(map[string]interface{}{
							"module_code":      "permission_key",
							"module_group_id":  newModuleGroupID,
							"feature_group_id": systemFeatureGroupID,
						}).Error; err != nil {
						return err
					}
					if err := tx.Model(&usermodel.PermissionKey{}).
						Where("module_group_id = ?", oldModuleGroupID).
						Update("module_group_id", newModuleGroupID).Error; err != nil {
						return err
					}
					if err := tx.Model(&usermodel.APIEndpoint{}).
						Where("category_id = ?", oldCategoryID).
						Update("category_id", newCategoryID).Error; err != nil {
						return err
					}
					if err := tx.Unscoped().
						Where("group_type = ? AND code = ?", "module", "permission_action").
						Delete(&usermodel.PermissionGroup{}).Error; err != nil {
						return err
					}
					if err := tx.Unscoped().
						Where("code = ?", "permission_action").
						Delete(&usermodel.APIEndpointCategory{}).Error; err != nil {
						return err
					}
					logger.Info("Named migration applied", zap.String("name", "20260324_permission_key_code_alignment_v2"))
					return nil
				})
			},
		},
		{
			Name: "20260324_team_menu_single_context_cleanup",
			Run: func(logger *zap.Logger) error {
				var teamRoot usermodel.Menu
				if err := database.DB.Where("name = ?", "TeamRoot").First(&teamRoot).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						logger.Info("Named migration applied", zap.String("name", "20260324_team_menu_single_context_cleanup"), zap.String("status", "skipped"))
						return nil
					}
					return err
				}

				teamAccessMeta := usermodel.MetaJSON{
					"keepAlive": true,
				}

				if err := database.DB.Model(&usermodel.Menu{}).
					Where("id = ?", teamRoot.ID).
					Updates(map[string]interface{}{
						"path":       "/team",
						"component":  "/index/index",
						"title":      "menus.system.team",
						"icon":       "ri:team-line",
						"sort_order": 5,
						"meta":       teamAccessMeta,
					}).Error; err != nil {
					return err
				}

				var teamMembers usermodel.Menu
				if err := database.DB.Where("name = ?", "TeamMembers").First(&teamMembers).Error; err == nil {
					if err := database.DB.Model(&usermodel.Menu{}).
						Where("id = ?", teamMembers.ID).
						Updates(map[string]interface{}{
							"parent_id":  teamRoot.ID,
							"path":       "members",
							"component":  "/team/team-members",
							"title":      "menus.system.teamMembers",
							"sort_order": 2,
							"meta":       teamAccessMeta,
						}).Error; err != nil {
						return err
					}
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				var teamRoles usermodel.Menu
				if err := database.DB.Where("name = ?", "TeamRolesAndPermissions").First(&teamRoles).Error; err == nil {
					if err := database.DB.Model(&usermodel.Menu{}).
						Where("id = ?", teamRoles.ID).
						Updates(map[string]interface{}{
							"parent_id":  teamRoot.ID,
							"path":       "roles",
							"component":  "/system/team-roles-permissions",
							"title":      "menus.system.teamRolesAndPermissions",
							"sort_order": 3,
							"meta":       teamAccessMeta,
						}).Error; err != nil {
						return err
					}
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				var teamManagement usermodel.Menu
				if err := database.DB.Where("name = ?", "TeamManagement").First(&teamManagement).Error; err == nil {
					if err := deleteMenuTree(teamManagement.ID); err != nil {
						return err
					}
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				logger.Info("Named migration applied", zap.String("name", "20260324_team_menu_single_context_cleanup"))
				return nil
			},
		},
		{
			Name: "20260324_drop_legacy_permission_tables",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE team_access_snapshots DROP COLUMN IF EXISTS manual_action_ids`,
					`DROP INDEX IF EXISTS idx_team_manual_action_permissions_unique`,
					`DROP TABLE IF EXISTS team_manual_action_permissions`,
					`DROP TABLE IF EXISTS tenant_action_permissions`,
					`DROP TABLE IF EXISTS role_menus`,
					`DROP TABLE IF EXISTS role_action_permissions`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_drop_legacy_permission_tables"))
				return nil
			},
		},
		{
			Name: "20260323_permission_key_consolidation",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE permission_keys ADD COLUMN IF NOT EXISTS permission_key varchar(150)`,
					`DROP INDEX IF EXISTS idx_permission_actions_resource_action_unique`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}

				hasResourceCode, err := hasColumn("permission_keys", "resource_code")
				if err != nil {
					return err
				}
				hasActionCode, err := hasColumn("permission_keys", "action_code")
				if err != nil {
					return err
				}
				type legacyPermissionKey struct {
					ID            uuid.UUID
					PermissionKey string
					ResourceCode  string
					ActionCode    string
				}
				var actions []legacyPermissionKey
				selects := []string{"id", "permission_key"}
				if hasResourceCode {
					selects = append(selects, "resource_code")
				}
				if hasActionCode {
					selects = append(selects, "action_code")
				}
				if err := database.DB.Table("permission_keys").Select(strings.Join(selects, ", ")).Order("created_at ASC, id ASC").Scan(&actions).Error; err != nil {
					return err
				}

				canonicalByKey := make(map[string]uuid.UUID, len(actions))
				duplicateIDs := make([]uuid.UUID, 0)
				for _, action := range actions {
					targetKey := permissionkey.Normalize(action.PermissionKey)
					if targetKey == "" && hasResourceCode && hasActionCode {
						mapping := permissionkey.FromLegacy(action.ResourceCode, action.ActionCode)
						targetKey = mapping.Key
					}
					if err := database.DB.Model(&usermodel.PermissionKey{}).
						Where("id = ?", action.ID).
						Update("permission_key", targetKey).Error; err != nil {
						return err
					}
					if canonicalID, exists := canonicalByKey[targetKey]; exists {
						if err := rebindPermissionKeyReferences(action.ID, canonicalID); err != nil {
							return err
						}
						duplicateIDs = append(duplicateIDs, action.ID)
						continue
					}
					canonicalByKey[targetKey] = action.ID
				}

				if len(duplicateIDs) > 0 {
					if err := database.DB.Where("id IN ?", duplicateIDs).Delete(&usermodel.PermissionKey{}).Error; err != nil {
						return err
					}
				}

				finishStatements := []string{
					`ALTER TABLE permission_keys ALTER COLUMN permission_key SET NOT NULL`,
					`DROP INDEX IF EXISTS idx_permission_actions_permission_key`,
					`DROP INDEX IF EXISTS idx_permission_keys_permission_key`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_permission_keys_permission_key ON permission_keys (permission_key) WHERE deleted_at IS NULL`,
				}
				for _, statement := range finishStatements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}

				logger.Info("Named migration applied", zap.String("name", "20260323_permission_key_consolidation"))
				return nil
			},
		},
		{
			Name: "20260324_permission_key_dot_cleanup",
			Run: func(logger *zap.Logger) error {
				type legacyPermissionKey struct {
					ID            uuid.UUID
					PermissionKey string
				}
				var actions []legacyPermissionKey
				if err := database.DB.Table("permission_keys").Select("id, permission_key").Where("permission_key LIKE ?", "%:%").Order("created_at ASC, id ASC").Scan(&actions).Error; err != nil {
					return err
				}
				if len(actions) == 0 {
					logger.Info("Named migration applied", zap.String("name", "20260324_permission_key_dot_cleanup"), zap.Int("updated", 0))
					return nil
				}

				canonicalByKey := map[string]uuid.UUID{}
				var existing []legacyPermissionKey
				if err := database.DB.Table("permission_keys").Select("id, permission_key").Order("created_at ASC, id ASC").Scan(&existing).Error; err != nil {
					return err
				}
				for _, action := range existing {
					key := permissionkey.Normalize(action.PermissionKey)
					if key == "" {
						continue
					}
					if _, exists := canonicalByKey[key]; !exists {
						canonicalByKey[key] = action.ID
					}
				}

				updatedCount := 0
				for _, action := range actions {
					targetKey := permissionkey.Normalize(action.PermissionKey)
					if targetKey == "" {
						continue
					}
					if canonicalID, exists := canonicalByKey[targetKey]; exists && canonicalID != action.ID {
						if err := rebindPermissionKeyReferences(action.ID, canonicalID); err != nil {
							return err
						}
						if err := database.DB.Delete(&usermodel.PermissionKey{}, action.ID).Error; err != nil {
							return err
						}
						updatedCount++
						continue
					}
					if err := database.DB.Model(&usermodel.PermissionKey{}).
						Where("id = ?", action.ID).
						Update("permission_key", targetKey).Error; err != nil {
						return err
					}
					canonicalByKey[targetKey] = action.ID
					updatedCount++
				}

				logger.Info("Named migration applied", zap.String("name", "20260324_permission_key_dot_cleanup"), zap.Int("updated", updatedCount))
				return nil
			},
		},
		{
			Name: "20260324_permission_context_backfill",
			Run: func(logger *zap.Logger) error {
				for _, mapping := range permissionkey.ListMappings() {
					contextType := strings.TrimSpace(mapping.ContextType)
					if contextType == "" {
						contextType = permissionkey.FromKey(mapping.Key).ContextType
					}
					if contextType == "" {
						continue
					}
					if err := database.DB.Model(&usermodel.PermissionKey{}).
						Where("permission_key = ?", mapping.Key).
						Update("context_type", contextType).Error; err != nil {
						return err
					}
				}

				statements := []string{
					`UPDATE permission_keys SET context_type = 'platform' WHERE permission_key LIKE 'system.%'`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE permission_key LIKE 'tenant.%'`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE permission_key LIKE 'platform.%'`,
					`UPDATE permission_keys SET context_type = 'team' WHERE permission_key LIKE 'team.%'`,
					`UPDATE permission_keys SET permission_key = 'system.permission.manage', context_type = 'platform' WHERE permission_key IN ('system_permission.manage_action_registry', 'permission_action.manage')`,
					`UPDATE permission_keys SET permission_key = 'system.role.assign_action', context_type = 'platform' WHERE permission_key = 'system_permission.assign_role_action'`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE module_code IN ('role', 'user', 'menu', 'menu_backup', 'permission_action', 'permission_key', 'api_endpoint', 'feature_package')`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE permission_key IN ('feature_package.assign_action', 'feature_package.assign_menu', 'feature_package.assign_team')`,
					`UPDATE permission_keys SET context_type = 'team' WHERE permission_key IN ('team.configure_action_boundary', 'team.configure_menu_boundary')`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}

				logger.Info("Named migration applied", zap.String("name", "20260324_permission_context_backfill"))
				return nil
			},
		},
		{
			Name: "20260325_legacy_user_roles_backfill",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`INSERT INTO user_roles (user_id, role_id, tenant_id)
					SELECT tm.user_id, r.id, tm.tenant_id
					FROM tenant_members tm
					JOIN roles r ON r.code = tm.role_code
					WHERE tm.role_code IN ('team_admin', 'team_member')
					  AND NOT EXISTS (
					    SELECT 1
					    FROM user_roles ur
					    WHERE ur.user_id = tm.user_id
					      AND ur.role_id = r.id
					      AND ur.tenant_id = tm.tenant_id
					  )`,
					`UPDATE user_roles ur
					SET tenant_id = tm.tenant_id
					FROM (
					  SELECT user_id, MAX(tenant_id::text)::uuid AS tenant_id
					  FROM tenant_members
					  GROUP BY user_id
					  HAVING COUNT(DISTINCT tenant_id) = 1
					) tm,
					roles r
					WHERE ur.user_id = tm.user_id
					  AND r.id = ur.role_id
					  AND ur.tenant_id IS NULL
					  AND r.code IN ('team_admin', 'team_member')
					  AND NOT EXISTS (
					    SELECT 1
					    FROM user_roles existing
					    WHERE existing.user_id = ur.user_id
					      AND existing.role_id = ur.role_id
					      AND existing.tenant_id = tm.tenant_id
					  )`,
					`DELETE FROM user_roles ur
					USING roles r
					WHERE ur.role_id = r.id
					  AND ur.tenant_id IS NULL
					  AND r.code IN ('team_admin', 'team_member')`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260325_legacy_user_roles_backfill"))
				return nil
			},
		},
		{
			Name: "20260323_permission_system_backfill_defaults",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE permission_keys SET status = 'normal' WHERE COALESCE(status, '') = ''`,
					`UPDATE user_action_permissions SET effect = 'allow' WHERE COALESCE(effect, '') = ''`,
					`ALTER TABLE permission_keys ADD COLUMN IF NOT EXISTS context_type varchar(20)`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE COALESCE(context_type, '') = '' AND (permission_key LIKE 'system.%' OR permission_key LIKE 'tenant.%' OR permission_key LIKE 'platform.%')`,
					`UPDATE permission_keys SET context_type = 'team' WHERE COALESCE(context_type, '') = ''`,
					`UPDATE permission_keys SET feature_kind = 'system' WHERE COALESCE(feature_kind, '') = ''`,
					`UPDATE api_endpoints SET feature_kind = 'system' WHERE COALESCE(feature_kind, '') = ''`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				var actions []usermodel.PermissionKey
				if err := database.DB.Find(&actions).Error; err != nil {
					return err
				}
				for _, action := range actions {
					if strings.TrimSpace(action.ModuleCode) != "" {
						continue
					}
					moduleCode := strings.TrimSpace(permissionkey.FromKey(action.PermissionKey).ResourceCode)
					if moduleCode == "" {
						continue
					}
					if err := database.DB.Model(&usermodel.PermissionKey{}).
						Where("id = ?", action.ID).
						Update("module_code", moduleCode).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_permission_system_backfill_defaults"))
				return nil
			},
		},
		{
			Name: "20260324_permission_action_drop_source_column",
			Run: func(logger *zap.Logger) error {
				hasSource, err := hasColumn("permission_keys", "source")
				if err != nil {
					return err
				}
				if !hasSource {
					logger.Info("Named migration skipped", zap.String("name", "20260324_permission_action_drop_source_column"))
					return nil
				}
				if err := database.DB.Exec(`ALTER TABLE permission_keys DROP COLUMN IF EXISTS source`).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_permission_action_drop_source_column"))
				return nil
			},
		},
		{
			Name: "20260324_permission_action_drop_source_column_v2",
			Run: func(logger *zap.Logger) error {
				hasSource, err := hasColumn("permission_keys", "source")
				if err != nil {
					return err
				}
				if !hasSource {
					logger.Info("Named migration skipped", zap.String("name", "20260324_permission_action_drop_source_column_v2"))
					return nil
				}
				if err := database.DB.Exec(`ALTER TABLE permission_keys DROP COLUMN IF EXISTS source`).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_permission_action_drop_source_column_v2"))
				return nil
			},
		},
		{
			Name: "20260324_permission_action_drop_legacy_code_columns",
			Run: func(logger *zap.Logger) error {
				if err := database.DB.Exec(`ALTER TABLE permission_keys DROP COLUMN IF EXISTS resource_code`).Error; err != nil {
					return err
				}
				if err := database.DB.Exec(`ALTER TABLE permission_keys DROP COLUMN IF EXISTS action_code`).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_permission_action_drop_legacy_code_columns"))
				return nil
			},
		},
		{
			Name: "20260324_permission_action_drop_legacy_code_columns_v2",
			Run: func(logger *zap.Logger) error {
				if err := database.DB.Exec(`ALTER TABLE permission_keys DROP COLUMN IF EXISTS resource_code`).Error; err != nil {
					return err
				}
				if err := database.DB.Exec(`ALTER TABLE permission_keys DROP COLUMN IF EXISTS action_code`).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_permission_action_drop_legacy_code_columns_v2"))
				return nil
			},
		},
		{
			Name: "20260323_feature_package_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS feature_packages (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						package_key varchar(100) NOT NULL,
						name varchar(150) NOT NULL,
						description varchar(255),
						context_type varchar(20) NOT NULL DEFAULT 'team',
						status varchar(20) NOT NULL DEFAULT 'normal',
						sort_order integer NOT NULL DEFAULT 0,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW(),
						deleted_at timestamptz
					)`,
					`CREATE TABLE IF NOT EXISTS feature_package_keys (
						package_id uuid NOT NULL,
						action_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS feature_package_menus (
						package_id uuid NOT NULL,
						menu_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS team_feature_packages (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						team_id uuid NOT NULL,
						package_id uuid NOT NULL,
						enabled boolean NOT NULL DEFAULT TRUE,
						granted_by uuid,
						granted_at timestamptz,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_packages_key_unique ON feature_packages (package_key) WHERE deleted_at IS NULL`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_package_keys_unique ON feature_package_keys (package_id, action_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_package_menus_unique ON feature_package_menus (package_id, menu_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_team_feature_packages_unique ON team_feature_packages (team_id, package_id)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_feature_package_schema"))
				return nil
			},
		},
		{
			Name: "20260323_role_feature_package_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS role_feature_packages (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						role_id uuid NOT NULL,
						package_id uuid NOT NULL,
						enabled boolean NOT NULL DEFAULT TRUE,
						granted_by uuid,
						granted_at timestamptz,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_role_feature_packages_unique ON role_feature_packages (role_id, package_id)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_role_feature_package_schema"))
				return nil
			},
		},
		{
			Name: "20260323_feature_package_v2_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE feature_packages ADD COLUMN IF NOT EXISTS package_type varchar(20) NOT NULL DEFAULT 'base'`,
					`ALTER TABLE feature_packages ADD COLUMN IF NOT EXISTS is_builtin boolean NOT NULL DEFAULT FALSE`,
					`UPDATE feature_packages SET package_type = 'base' WHERE COALESCE(package_type, '') = ''`,
					`CREATE TABLE IF NOT EXISTS feature_package_bundles (
						package_id uuid NOT NULL,
						child_package_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS user_feature_packages (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						user_id uuid NOT NULL,
						package_id uuid NOT NULL,
						enabled boolean NOT NULL DEFAULT TRUE,
						granted_by uuid,
						granted_at timestamptz,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS role_hidden_menus (
						role_id uuid NOT NULL,
						menu_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS role_disabled_actions (
						role_id uuid NOT NULL,
						action_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS team_blocked_menus (
						team_id uuid NOT NULL,
						menu_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS team_blocked_actions (
						team_id uuid NOT NULL,
						action_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS user_hidden_menus (
						user_id uuid NOT NULL,
						menu_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_package_bundles_unique ON feature_package_bundles (package_id, child_package_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_feature_packages_unique ON user_feature_packages (user_id, package_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_role_hidden_menus_unique ON role_hidden_menus (role_id, menu_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_role_disabled_actions_unique ON role_disabled_actions (role_id, action_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_team_blocked_menus_unique ON team_blocked_menus (team_id, menu_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_team_blocked_actions_unique ON team_blocked_actions (team_id, action_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_hidden_menus_unique ON user_hidden_menus (user_id, menu_id)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_feature_package_v2_schema"))
				return nil
			},
		},
		{
			Name: "20260324_access_snapshot_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS platform_user_access_snapshots (
						user_id uuid PRIMARY KEY,
						role_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						role_package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						user_package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						direct_package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						expanded_package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						action_source_map jsonb NOT NULL DEFAULT '{}'::jsonb,
						available_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						available_menu_map jsonb NOT NULL DEFAULT '{}'::jsonb,
						menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						menu_source_map jsonb NOT NULL DEFAULT '{}'::jsonb,
						hidden_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						disabled_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						has_package_config boolean NOT NULL DEFAULT FALSE,
						refreshed_at timestamptz NOT NULL DEFAULT NOW(),
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS team_access_snapshots (
						team_id uuid PRIMARY KEY,
						package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						expanded_package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						derived_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						derived_action_map jsonb NOT NULL DEFAULT '{}'::jsonb,
						blocked_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						effective_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						derived_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						derived_menu_map jsonb NOT NULL DEFAULT '{}'::jsonb,
						blocked_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						effective_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						refreshed_at timestamptz NOT NULL DEFAULT NOW(),
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_access_snapshot_schema"))
				return nil
			},
		},
		{
			Name: "20260324_team_role_access_snapshot_boundary_fields",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS team_role_access_snapshots (
						team_id uuid NOT NULL,
						role_id uuid NOT NULL,
						package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						expanded_package_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						available_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						disabled_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						action_source_map jsonb NOT NULL DEFAULT '{}'::jsonb,
						available_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						hidden_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
						menu_source_map jsonb NOT NULL DEFAULT '{}'::jsonb,
						inherited boolean NOT NULL DEFAULT FALSE,
						refreshed_at timestamptz NOT NULL DEFAULT NOW(),
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW(),
						PRIMARY KEY (team_id, role_id)
					)`,
					`ALTER TABLE team_role_access_snapshots ADD COLUMN IF NOT EXISTS available_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
					`ALTER TABLE team_role_access_snapshots ADD COLUMN IF NOT EXISTS disabled_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
					`ALTER TABLE team_role_access_snapshots ADD COLUMN IF NOT EXISTS available_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
					`ALTER TABLE team_role_access_snapshots ADD COLUMN IF NOT EXISTS hidden_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_team_role_access_snapshot_boundary_fields"))
				return nil
			},
		},
		{
			Name: "20260324_role_custom_params_jsonb",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE roles ADD COLUMN IF NOT EXISTS custom_params jsonb NOT NULL DEFAULT '{}'::jsonb`,
					`UPDATE roles SET custom_params = '{}'::jsonb WHERE custom_params IS NULL`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_role_custom_params_jsonb"))
				return nil
			},
		},
		{
			Name: "20260324_drop_legacy_snapshot_columns",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE platform_role_access_snapshots DROP COLUMN IF EXISTS has_package_config`,
					`ALTER TABLE team_role_access_snapshots DROP COLUMN IF EXISTS has_menu_boundary`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260324_drop_legacy_snapshot_columns"))
				return nil
			},
		},
		{
			Name: "20260323_permission_metadata_refresh",
			Run: func(logger *zap.Logger) error {
				for _, mapping := range permissionkey.ListMappings() {
					if strings.TrimSpace(mapping.Key) == "" {
						continue
					}
					updates := map[string]interface{}{}
					if strings.TrimSpace(mapping.Name) != "" {
						updates["name"] = strings.TrimSpace(mapping.Name)
					}
					if strings.TrimSpace(mapping.Description) != "" {
						updates["description"] = strings.TrimSpace(mapping.Description)
					}
					if strings.TrimSpace(mapping.ResourceCode) != "" {
						updates["module_code"] = strings.TrimSpace(mapping.ResourceCode)
					}
					if strings.TrimSpace(mapping.ContextType) != "" {
						updates["context_type"] = strings.TrimSpace(mapping.ContextType)
					}
					if len(updates) == 0 {
						continue
					}
					if err := database.DB.Model(&usermodel.PermissionKey{}).
						Where("permission_key = ?", mapping.Key).
						Updates(updates).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_permission_metadata_refresh"))
				return nil
			},
		},
		{
			Name: "20260323_drop_permission_category",
			Run: func(logger *zap.Logger) error {
				if err := database.DB.Exec(`ALTER TABLE permission_keys DROP COLUMN IF EXISTS category`).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_drop_permission_category"))
				return nil
			},
		},
		{
			Name: "20260323_role_tenant_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE roles ADD COLUMN IF NOT EXISTS tenant_id uuid`,
					`DROP INDEX IF EXISTS roles_code_key`,
					`DROP INDEX IF EXISTS idx_roles_code`,
					`DROP INDEX IF EXISTS idx_roles_code_unique`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_global_code_unique ON roles (code) WHERE tenant_id IS NULL AND deleted_at IS NULL`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_tenant_code_unique ON roles (tenant_id, code) WHERE tenant_id IS NOT NULL AND deleted_at IS NULL`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_role_tenant_schema"))
				return nil
			},
		},
		{
			Name: "20260323_restore_permission_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{}
				if hasScopeTargetID, err := hasColumn("user_roles", "scope_target_id"); err != nil {
					return err
				} else if hasScopeTargetID {
					statements = append(statements, `UPDATE user_roles
					  SET tenant_id = COALESCE(tenant_id, scope_target_id)
					  WHERE tenant_id IS NULL AND scope_target_id IS NOT NULL`)
				}
				if hasScopeTargetID, err := hasColumn("user_action_permissions", "scope_target_id"); err != nil {
					return err
				} else if hasScopeTargetID {
					statements = append(statements, `UPDATE user_action_permissions
					  SET tenant_id = COALESCE(tenant_id, scope_target_id)
					  WHERE tenant_id IS NULL AND scope_target_id IS NOT NULL`)
				}
				statements = append(statements,
					`ALTER TABLE user_action_permissions DROP CONSTRAINT IF EXISTS user_action_permissions_pkey`,
					`ALTER TABLE user_action_permissions ALTER COLUMN tenant_id DROP NOT NULL`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_global_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_tenant_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_global_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_tenant_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_id`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_target_id`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_action_permissions_global_unique ON user_action_permissions (user_id, action_id) WHERE tenant_id IS NULL`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_action_permissions_tenant_unique ON user_action_permissions (user_id, action_id, tenant_id) WHERE tenant_id IS NOT NULL`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_global_unique`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_target_unique`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_id`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_target_id`,
					`DROP INDEX IF EXISTS idx_user_roles_deleted_at`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_global_unique ON user_roles (user_id, role_id) WHERE tenant_id IS NULL`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_tenant_unique ON user_roles (user_id, role_id, tenant_id) WHERE tenant_id IS NOT NULL`,
					`ALTER TABLE user_roles DROP COLUMN IF EXISTS scope_id`,
					`ALTER TABLE user_roles DROP COLUMN IF EXISTS scope_target_id`,
					`ALTER TABLE user_roles DROP COLUMN IF EXISTS created_at`,
					`ALTER TABLE user_roles DROP COLUMN IF EXISTS deleted_at`,
					`ALTER TABLE user_action_permissions DROP COLUMN IF EXISTS scope_id`,
					`ALTER TABLE user_action_permissions DROP COLUMN IF EXISTS scope_target_id`,
					`ALTER TABLE roles DROP COLUMN IF EXISTS enabled`,
					`DROP TABLE IF EXISTS permissions`,
					`DROP TABLE IF EXISTS role_permissions`,
					`DROP TABLE IF EXISTS role_scope_bindings`,
					`UPDATE menus
					   SET meta = COALESCE(meta, '{}'::jsonb) - 'authList' - 'authMark' - 'isAuthButton' - 'requiresTenantContext'
					 WHERE meta ? 'authList' OR meta ? 'authMark' OR meta ? 'isAuthButton' OR meta ? 'requiresTenantContext'`,
					`ALTER TABLE permission_keys DROP COLUMN IF EXISTS requires_tenant_context`,
					`ALTER TABLE api_endpoints DROP COLUMN IF EXISTS requires_tenant_context`,
				)
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_restore_permission_schema"))
				return nil
			},
		},
		{
			Name: "20260323_drop_unused_core_columns",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at`,
					`ALTER TABLE menus DROP COLUMN IF EXISTS redirect`,
					`ALTER TABLE menus DROP COLUMN IF EXISTS visible`,
					`ALTER TABLE api_keys DROP COLUMN IF EXISTS key_prefix`,
					`ALTER TABLE api_keys DROP COLUMN IF EXISTS permissions`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_drop_unused_core_columns"))
				return nil
			},
		},
		{
			Name: "20260323_drop_scope_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`DELETE FROM menus WHERE name = 'Scope' OR path = 'scope' OR component = '/system/scope'`,
					`DELETE FROM user_action_permissions WHERE action_id IN (SELECT id FROM permission_keys WHERE permission_key LIKE 'scope.%')`,
					`DELETE FROM permission_keys WHERE permission_key LIKE 'scope.%'`,
					`ALTER TABLE role_data_permissions ADD COLUMN IF NOT EXISTS data_scope varchar(30)`,
					`ALTER TABLE roles DROP COLUMN IF EXISTS scope_id`,
					`ALTER TABLE permission_keys DROP COLUMN IF EXISTS scope_id`,
					`ALTER TABLE permission_keys DROP COLUMN IF EXISTS scope_type`,
					`ALTER TABLE api_endpoints DROP COLUMN IF EXISTS scope_id`,
					`DROP INDEX IF EXISTS idx_permission_actions_scope_id`,
					`DROP INDEX IF EXISTS idx_permission_actions_resource_action_unique`,
					`DROP INDEX IF EXISTS idx_permission_actions_permission_key`,
					`DROP INDEX IF EXISTS idx_permission_keys_permission_key`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_permission_keys_permission_key ON permission_keys (permission_key) WHERE deleted_at IS NULL`,
					`DROP TABLE IF EXISTS role_scopes`,
					`DROP TABLE IF EXISTS scopes`,
				}
				roleDataScopeUpdate, err := buildRoleDataScopeUpdateStatement()
				if err != nil {
					return err
				}
				statements = append(statements,
					roleDataScopeUpdate,
					`UPDATE role_data_permissions SET data_scope = 'self' WHERE COALESCE(data_scope, '') = ''`,
					`ALTER TABLE role_data_permissions ALTER COLUMN data_scope SET DEFAULT 'self'`,
					`ALTER TABLE role_data_permissions ALTER COLUMN data_scope SET NOT NULL`,
					`ALTER TABLE role_data_permissions DROP COLUMN IF EXISTS scope_code`,
					`ALTER TABLE role_data_permissions DROP COLUMN IF EXISTS data_permission_code`,
					`ALTER TABLE role_data_permissions DROP COLUMN IF EXISTS data_permission_name`,
				)
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_drop_scope_schema"))
				return nil
			},
		},
	}

	for _, task := range tasks {
		done, err := hasMigrationRun(task.Name)
		if err != nil {
			return err
		}
		if done {
			logger.Info("Named migration already applied", zap.String("name", task.Name))
			continue
		}
		if err := task.Run(logger); err != nil {
			return err
		}
		if err := markMigrationRun(task.Name); err != nil {
			return err
		}
	}

	return nil
}

func preparePermissionTableRenames(logger *zap.Logger) error {
	renamePairs := []struct {
		oldName string
		newName string
	}{
		{oldName: "permission_actions", newName: "permission_keys"},
		{oldName: "feature_package_actions", newName: "feature_package_keys"},
	}
	for _, item := range renamePairs {
		oldExists, err := hasTable(item.oldName)
		if err != nil {
			return err
		}
		newExists, err := hasTable(item.newName)
		if err != nil {
			return err
		}
		if !oldExists || newExists {
			continue
		}
		if err := database.DB.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", item.oldName, item.newName)).Error; err != nil {
			return err
		}
		logger.Info("Renamed legacy permission table",
			zap.String("from", item.oldName),
			zap.String("to", item.newName),
		)
	}
	return nil
}

func ensureMigrationHistoryTable() error {
	return database.DB.Exec(`
		CREATE TABLE IF NOT EXISTS app_migrations (
			id bigserial PRIMARY KEY,
			name varchar(200) NOT NULL UNIQUE,
			executed_at timestamptz NOT NULL DEFAULT NOW()
		)
	`).Error
}

func hasMigrationRun(name string) (bool, error) {
	var count int64
	if err := database.DB.Raw("SELECT COUNT(*) FROM app_migrations WHERE name = ?", name).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func markMigrationRun(name string) error {
	return database.DB.Exec("INSERT INTO app_migrations (name) VALUES (?) ON CONFLICT (name) DO NOTHING", name).Error
}

func hasTable(tableName string) (bool, error) {
	var count int64
	err := database.DB.Raw(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = CURRENT_SCHEMA()
		  AND table_name = ?
		  AND table_type = 'BASE TABLE'
	`, tableName).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func hasColumn(tableName, columnName string) (bool, error) {
	var count int64
	err := database.DB.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA()
		  AND table_name = ?
		  AND column_name = ?
	`, tableName, columnName).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func buildRoleDataScopeUpdateStatement() (string, error) {
	candidates := make([]string, 0, 3)
	if hasDataScope, err := hasColumn("role_data_permissions", "data_scope"); err != nil {
		return "", err
	} else if hasDataScope {
		candidates = append(candidates, "NULLIF(data_scope, '')")
	}
	if hasScopeCode, err := hasColumn("role_data_permissions", "scope_code"); err != nil {
		return "", err
	} else if hasScopeCode {
		candidates = append(candidates, "NULLIF(scope_code, '')")
	}
	if hasDataPermissionCode, err := hasColumn("role_data_permissions", "data_permission_code"); err != nil {
		return "", err
	} else if hasDataPermissionCode {
		candidates = append(candidates, "NULLIF(data_permission_code, '')")
	}
	if len(candidates) == 0 {
		return `UPDATE role_data_permissions SET data_scope = 'self'`, nil
	}
	return fmt.Sprintf(
		`UPDATE role_data_permissions SET data_scope = COALESCE(%s, 'self')`,
		strings.Join(candidates, ", "),
	), nil
}

func rebindPermissionKeyReferences(fromID, toID uuid.UUID) error {
	if fromID == toID {
		return nil
	}
	statements := []struct {
		sql  string
		args []interface{}
	}{
		{
			sql: `UPDATE user_action_permissions target
				    SET action_id = ?
				  WHERE action_id = ?
				    AND NOT EXISTS (
				      SELECT 1 FROM user_action_permissions existing
				      WHERE existing.user_id = target.user_id
				        AND existing.action_id = ?
				        AND (
				          (existing.tenant_id IS NULL AND target.tenant_id IS NULL) OR
				          existing.tenant_id = target.tenant_id
				        )
				    )`,
			args: []interface{}{toID, fromID, toID},
		},
		{
			sql:  `DELETE FROM user_action_permissions WHERE action_id = ?`,
			args: []interface{}{fromID},
		},
		{
			sql: `UPDATE feature_package_keys target
				    SET action_id = ?
				  WHERE action_id = ?
				    AND NOT EXISTS (
				      SELECT 1 FROM feature_package_keys existing
				      WHERE existing.package_id = target.package_id
				        AND existing.action_id = ?
				    )`,
			args: []interface{}{toID, fromID, toID},
		},
		{
			sql:  `DELETE FROM feature_package_keys WHERE action_id = ?`,
			args: []interface{}{fromID},
		},
		{
			sql: `UPDATE role_disabled_actions target
				    SET action_id = ?
				  WHERE action_id = ?
				    AND NOT EXISTS (
				      SELECT 1 FROM role_disabled_actions existing
				      WHERE existing.role_id = target.role_id
				        AND existing.action_id = ?
				    )`,
			args: []interface{}{toID, fromID, toID},
		},
		{
			sql:  `DELETE FROM role_disabled_actions WHERE action_id = ?`,
			args: []interface{}{fromID},
		},
		{
			sql: `UPDATE team_blocked_actions target
				    SET action_id = ?
				  WHERE action_id = ?
				    AND NOT EXISTS (
				      SELECT 1 FROM team_blocked_actions existing
				      WHERE existing.team_id = target.team_id
				        AND existing.action_id = ?
				    )`,
			args: []interface{}{toID, fromID, toID},
		},
		{
			sql:  `DELETE FROM team_blocked_actions WHERE action_id = ?`,
			args: []interface{}{fromID},
		},
	}
	for _, statement := range statements {
		if err := database.DB.Exec(statement.sql, statement.args...).Error; err != nil {
			return err
		}
	}
	return nil
}

func mergeDeprecatedTeamPermissionKeys(logger *zap.Logger) error {
	merges := []struct {
		From string
		To   string
	}{
		{From: "tenant.member.manage", To: "tenant.manage"},
		{From: "tenant.boundary.manage", To: "tenant.manage"},
		{From: "tenant.configure_menu_boundary", To: "tenant.manage"},
		{From: "user.assign_menu", To: "system.user.manage"},
		{From: "team.member.assign_role", To: "team.member.manage"},
		{From: "team.member.assign_action", To: "team.member.manage"},
		{From: "team.configure_menu_boundary", To: "team.boundary.manage"},
		{From: "feature_package.assign_menu", To: "platform.package.manage"},
	}

	mergedCount := 0
	for _, item := range merges {
		var fromAction usermodel.PermissionKey
		if err := database.DB.Where("permission_key = ?", item.From).First(&fromAction).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		var toAction usermodel.PermissionKey
		if err := database.DB.Where("permission_key = ?", item.To).First(&toAction).Error; err != nil {
			return err
		}

		if err := rebindPermissionKeyReferences(fromAction.ID, toAction.ID); err != nil {
			return err
		}
		if err := database.DB.Delete(&usermodel.PermissionKey{}, fromAction.ID).Error; err != nil {
			return err
		}
		mergedCount++
	}

	logger.Info("Deprecated team permission key merge applied", zap.Int("merged", mergedCount))
	return nil
}

func syncCanonicalPermissionKeys(logger *zap.Logger) error {
	seen := make(map[string]struct{})
	updatedCount := 0
	for _, mapping := range permissionkey.ListMappings() {
		if mapping.Key == "" {
			continue
		}
		if _, exists := seen[mapping.Key]; exists {
			continue
		}
		seen[mapping.Key] = struct{}{}

		var item usermodel.PermissionKey
		if err := database.DB.Where("permission_key = ? AND deleted_at IS NULL", mapping.Key).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		updates := map[string]interface{}{}
		if strings.TrimSpace(item.Code) == "" {
			updates["code"] = permissionseed.StableID("permission-action-code", mapping.Key).String()
		}
		if strings.TrimSpace(item.ModuleCode) == "" {
			if moduleCode := strings.TrimSpace(mapping.ResourceCode); moduleCode != "" {
				updates["module_code"] = moduleCode
			}
		}
		if len(updates) == 0 {
			continue
		}
		updates["updated_at"] = time.Now()
		result := database.DB.Model(&item).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		updatedCount += int(result.RowsAffected)
	}
	logger.Info("Canonical permission key sync applied", zap.Int("updated", updatedCount))
	return nil
}

func backfillTenantIdentityUserRoles(logger *zap.Logger) error {
	type tenantMemberRow struct {
		TenantID uuid.UUID
		UserID   uuid.UUID
		RoleCode string
		Status   string
	}

	var members []tenantMemberRow
	if err := database.DB.Model(&usermodel.TenantMember{}).
		Select("tenant_id, user_id, role_code, status").
		Where("status = ?", "active").
		Find(&members).Error; err != nil {
		return err
	}
	if len(members) == 0 {
		logger.Info("Tenant identity user role backfill applied", zap.Int("members", 0), zap.Int("rebound", 0), zap.Int("teams", 0))
		return nil
	}

	roleCodes := make([]string, 0)
	seenCodes := make(map[string]struct{})
	for _, member := range members {
		code := strings.TrimSpace(member.RoleCode)
		if code == "" {
			continue
		}
		if _, exists := seenCodes[code]; exists {
			continue
		}
		seenCodes[code] = struct{}{}
		roleCodes = append(roleCodes, code)
	}
	if len(roleCodes) == 0 {
		logger.Info("Tenant identity user role backfill applied", zap.Int("members", len(members)), zap.Int("rebound", 0), zap.Int("teams", 0))
		return nil
	}

	var roles []usermodel.Role
	if err := database.DB.Where("tenant_id IS NULL AND code IN ?", roleCodes).Find(&roles).Error; err != nil {
		return err
	}
	roleIDByCode := make(map[string]uuid.UUID, len(roles))
	identityRoleIDs := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		code := strings.TrimSpace(role.Code)
		if code == "" {
			continue
		}
		roleIDByCode[code] = role.ID
		identityRoleIDs = append(identityRoleIDs, role.ID)
	}
	if len(identityRoleIDs) == 0 {
		logger.Info("Tenant identity user role backfill applied", zap.Int("members", len(members)), zap.Int("rebound", 0), zap.Int("teams", 0))
		return nil
	}

	reboundCount := 0
	touchedTeams := make([]uuid.UUID, 0)
	seenTeams := make(map[uuid.UUID]struct{})
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, member := range members {
			roleID, ok := roleIDByCode[strings.TrimSpace(member.RoleCode)]
			if !ok {
				continue
			}
			if err := tx.Where("user_id = ? AND tenant_id = ? AND role_id IN ?", member.UserID, member.TenantID, identityRoleIDs).
				Delete(&usermodel.UserRole{}).Error; err != nil {
				return err
			}
			record := usermodel.UserRole{
				UserID:   member.UserID,
				RoleID:   roleID,
				TenantID: &member.TenantID,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			reboundCount++
			if _, exists := seenTeams[member.TenantID]; !exists {
				seenTeams[member.TenantID] = struct{}{}
				touchedTeams = append(touchedTeams, member.TenantID)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	boundaryService := teamboundary.NewService(database.DB)
	for _, teamID := range touchedTeams {
		if _, err := boundaryService.RefreshSnapshot(teamID); err != nil {
			return err
		}
	}

	logger.Info("Tenant identity user role backfill applied",
		zap.Int("members", len(members)),
		zap.Int("rebound", reboundCount),
		zap.Int("teams", len(touchedTeams)),
	)
	return nil
}

func initDefaultRolesNoScope(logger *zap.Logger) error {
	roles := []struct {
		Code        string
		Name        string
		Description string
		SortOrder   int
	}{
		{"admin", "管理员", "系统管理员，拥有所有权限", 1},
		{"team_admin", "团队管理员", "团队管理员，可以管理团队成员和团队内容", 2},
		{"team_member", "团队成员", "团队成员，可以查看和编辑团队内容", 3},
	}

	for _, roleData := range roles {
		var role usermodel.Role
		result := database.DB.Where("code = ? AND tenant_id IS NULL", roleData.Code).First(&role)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				role = usermodel.Role{
					Code:        roleData.Code,
					Name:        roleData.Name,
					Description: roleData.Description,
					SortOrder:   roleData.SortOrder,
					IsSystem:    true,
				}
				if err := database.DB.Create(&role).Error; err != nil {
					logger.Error("Failed to create role", zap.String("code", roleData.Code), zap.Error(err))
					return err
				}
				logger.Info("Role created", zap.String("code", roleData.Code), zap.String("name", roleData.Name))
			} else {
				return result.Error
			}
		} else {
			_ = database.DB.Model(&role).Updates(map[string]interface{}{
				"name":        roleData.Name,
				"description": roleData.Description,
				"sort_order":  roleData.SortOrder,
				"is_system":   true,
			}).Error
			logger.Info("Role already exists", zap.String("code", roleData.Code))
		}
	}
	return nil
}

func initDefaultAdmin(logger *zap.Logger) error {
	userRepo := usermodel.NewUserRepository(database.DB)

	defaultEmail := "admin@gg.com"
	defaultUsername := "admin"
	defaultPassword := "admin123456"
	defaultNickname := "系统管理员"

	exists, err := userRepo.ExistsByUsername(defaultUsername)
	if err != nil {
		return err
	}

	if exists {
		adminUser, err := userRepo.GetByUsername(defaultUsername)
		if err != nil {
			return err
		}
		updates := map[string]interface{}{
			"is_super_admin": true,
			"status":         "active",
		}
		if adminUser.Nickname == "" {
			updates["nickname"] = defaultNickname
		}
		if err := database.DB.Model(&usermodel.User{}).Where("id = ?", adminUser.ID).Updates(updates).Error; err != nil {
			return err
		}

		if err := assignAdminRole(adminUser.ID, logger); err != nil {
			return err
		}

		logger.Info("Default admin already exists", zap.String("username", defaultUsername))
		return nil
	}

	passwordHash, err := password.Hash(defaultPassword)
	if err != nil {
		return err
	}

	admin := &usermodel.User{
		Email:        defaultEmail,
		Username:     defaultUsername,
		PasswordHash: passwordHash,
		Nickname:     defaultNickname,
		Status:       "active",
		IsSuperAdmin: true,
	}

	if err := userRepo.Create(admin); err != nil {
		return err
	}

	logger.Info("Default admin created",
		zap.String("email", admin.Email),
		zap.String("username", admin.Username),
		zap.String("user_id", admin.ID.String()),
	)

	if err := assignAdminRole(admin.ID, logger); err != nil {
		return err
	}

	return nil
}

func assignAdminRole(userID uuid.UUID, logger *zap.Logger) error {
	var adminRole usermodel.Role
	if err := database.DB.Where("code = ? AND tenant_id IS NULL", "admin").First(&adminRole).Error; err != nil {
		logger.Error("Failed to find admin role", zap.Error(err))
		return err
	}

	var userRole usermodel.UserRole
	result := database.DB.Where("user_id = ? AND role_id = ? AND tenant_id IS NULL", userID, adminRole.ID).First(&userRole)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			userRole = usermodel.UserRole{
				UserID:   userID,
				RoleID:   adminRole.ID,
				TenantID: nil,
			}
			if err := database.DB.Create(&userRole).Error; err != nil {
				logger.Error("Failed to assign admin role", zap.Error(err))
				return err
			}
			logger.Info("Admin role assigned", zap.String("user_id", userID.String()))
		} else {
			return result.Error
		}
	} else {
		logger.Info("Admin role already assigned", zap.String("user_id", userID.String()))
	}

	return nil
}

func initDefaultMenusNoScope(logger *zap.Logger) error {
	if err := cleanupDeprecatedMenus(logger); err != nil {
		return err
	}

	for _, spec := range permissionseed.DefaultMenus() {
		if _, err := ensureMenuSeed(spec); err != nil {
			return err
		}
	}

	logger.Info("Default menus ensured")
	return nil
}

func ensureDefaultMenuSeedByName(name string) (*usermodel.Menu, error) {
	targetName := strings.TrimSpace(name)
	if targetName == "" {
		return nil, fmt.Errorf("menu seed name is required")
	}

	for _, spec := range permissionseed.DefaultMenus() {
		if spec.Name != targetName {
			continue
		}
		if spec.ParentName != "" {
			if _, err := ensureDefaultMenuSeedByName(spec.ParentName); err != nil {
				return nil, err
			}
		}
		return ensureMenuSeed(spec)
	}

	return nil, fmt.Errorf("default menu seed not found: %s", targetName)
}

func ensureMenuSeed(spec permissionseed.MenuSeed) (*usermodel.Menu, error) {
	var existing usermodel.Menu
	if err := database.DB.Where("name = ?", spec.Name).First(&existing).Error; err == nil {
		return &existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var parentID *uuid.UUID
	if spec.ParentName != "" {
		var parent usermodel.Menu
		if err := database.DB.Where("name = ?", spec.ParentName).First(&parent).Error; err != nil {
			return nil, err
		}
		parentID = &parent.ID
	}

	menu := &usermodel.Menu{
		ParentID:  parentID,
		Path:      spec.Path,
		Name:      spec.Name,
		Component: spec.Component,
		Title:     spec.Title,
		Icon:      spec.Icon,
		SortOrder: spec.SortOrder,
		Meta:      spec.Meta,
	}
	if err := database.DB.Create(menu).Error; err != nil {
		return nil, err
	}
	return menu, nil
}

func ensureMenuAccessMode(menuName, accessMode string) error {
	targetName := strings.TrimSpace(menuName)
	targetMode := strings.TrimSpace(strings.ToLower(accessMode))
	if targetName == "" {
		return fmt.Errorf("menu name is required")
	}
	if targetMode != "permission" && targetMode != "jwt" && targetMode != "public" {
		return fmt.Errorf("invalid access mode: %s", accessMode)
	}

	var menu usermodel.Menu
	if err := database.DB.Where("name = ?", targetName).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	meta := menu.Meta
	if meta == nil {
		meta = usermodel.MetaJSON{}
	}
	if current := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", meta["accessMode"]))); current == targetMode {
		return nil
	}
	meta["accessMode"] = targetMode
	return database.DB.Model(&menu).Update("meta", meta).Error
}

func cleanupDeprecatedMenus(logger *zap.Logger) error {
	deprecatedNames := permissionseed.DeprecatedDefaultMenuNames()
	var deprecatedMenus []usermodel.Menu
	if err := database.DB.Where("name IN ?", deprecatedNames).Find(&deprecatedMenus).Error; err != nil {
		return err
	}
	for _, menu := range deprecatedMenus {
		if err := database.DB.Delete(&usermodel.Menu{}, menu.ID).Error; err != nil {
			return err
		}
		logger.Info("Deprecated default menu removed", zap.String("name", menu.Name))
	}
	var legacyTeamRoots []usermodel.Menu
	if err := database.DB.Where("path = ? AND component = ? AND COALESCE(name, '') <> ?", "/team", "/index/index", "TeamRoot").Find(&legacyTeamRoots).Error; err != nil {
		return err
	}
	for _, menu := range legacyTeamRoots {
		if err := deleteMenuTree(menu.ID); err != nil {
			return err
		}
		logger.Info("Legacy team menu tree removed", zap.String("name", menu.Name), zap.String("path", menu.Path))
	}
	return nil
}

func backfillMenuManageGroups(logger *zap.Logger) error {
	var menus []usermodel.Menu
	if err := database.DB.Find(&menus).Error; err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		groupNameToID := make(map[string]uuid.UUID)

		for i := range menus {
			menu := &menus[i]
			if menu.Meta == nil {
				continue
			}
			rawName, ok := menu.Meta["manageGroup"]
			if !ok {
				continue
			}
			groupName := strings.TrimSpace(fmt.Sprint(rawName))
			delete(menu.Meta, "manageGroup")
			if groupName == "" {
				if err := tx.Model(&usermodel.Menu{}).Where("id = ?", menu.ID).Updates(map[string]interface{}{
					"meta":            menu.Meta,
					"manage_group_id": nil,
				}).Error; err != nil {
					return err
				}
				continue
			}

			groupID, exists := groupNameToID[groupName]
			if !exists {
				group := usermodel.MenuManageGroup{
					Name:      groupName,
					SortOrder: 0,
					Status:    "normal",
				}
				if err := tx.Where("name = ?", groupName).FirstOrCreate(&group).Error; err != nil {
					return err
				}
				groupID = group.ID
				groupNameToID[groupName] = groupID
			}

			if err := tx.Model(&usermodel.Menu{}).Where("id = ?", menu.ID).Updates(map[string]interface{}{
				"meta":            menu.Meta,
				"manage_group_id": groupID,
			}).Error; err != nil {
				return err
			}
		}

		logger.Info("Menu manage groups backfilled", zap.Int("menus", len(menus)), zap.Int("groups", len(groupNameToID)))
		return nil
	})
}

func deleteMenuTree(rootID uuid.UUID) error {
	queue := []uuid.UUID{rootID}
	collected := make([]uuid.UUID, 0, 8)
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		collected = append(collected, currentID)
		var children []usermodel.Menu
		if err := database.DB.Select("id").Where("parent_id = ?", currentID).Find(&children).Error; err != nil {
			return err
		}
		for _, child := range children {
			queue = append(queue, child.ID)
		}
	}
	if len(collected) == 0 {
		return nil
	}
	if err := database.DB.Where("menu_id IN ?", collected).Delete(&usermodel.FeaturePackageMenu{}).Error; err != nil {
		return err
	}
	return database.DB.Where("id IN ?", collected).Delete(&usermodel.Menu{}).Error
}

func initDefaultPermissionGroups(logger *zap.Logger) error {
	if err := permissionseed.EnsureDefaultPermissionGroups(database.DB); err != nil {
		return err
	}
	logger.Info("Default permission groups seeded")
	return nil
}

func initDefaultPermissionKeysNoScope(logger *zap.Logger) error {
	if err := permissionseed.EnsureDefaultPermissionKeys(database.DB); err != nil {
		return err
	}
	logger.Info("Default permission keys seeded")
	return nil
}

func initDefaultFeaturePackages(logger *zap.Logger) error {
	if err := permissionseed.EnsureDefaultFeaturePackages(database.DB); err != nil {
		return err
	}
	logger.Info("Default feature packages seeded")
	return nil
}

func initDefaultFeaturePackageBundles(logger *zap.Logger) error {
	if err := permissionseed.EnsureDefaultFeaturePackageBundles(database.DB); err != nil {
		return err
	}
	logger.Info("Default feature package bundles seeded")
	return nil
}

func initDefaultRoleFeaturePackages(logger *zap.Logger) error {
	if err := permissionseed.EnsureDefaultRoleFeaturePackages(database.DB); err != nil {
		return err
	}
	logger.Info("Default role feature packages ensured")
	return nil
}

func refreshDefaultAccessSnapshots(logger *zap.Logger) error {
	boundaryService := teamboundary.NewService(database.DB)
	platformService := platformaccess.NewService(database.DB)
	roleSnapshotService := platformroleaccess.NewService(database.DB)
	refresher := permissionrefresh.NewService(database.DB, boundaryService, platformService, roleSnapshotService)

	defaultRoleCodes := permissionseed.DefaultRoleCodes()
	var roles []usermodel.Role
	if err := database.DB.Where("tenant_id IS NULL AND code IN ?", defaultRoleCodes).Find(&roles).Error; err != nil {
		return err
	}
	roleIDs := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	if err := refresher.RefreshPlatformRoles(roleIDs); err != nil {
		return err
	}

	defaultPackageKeys := permissionseed.DefaultFeaturePackageKeys()
	var packages []usermodel.FeaturePackage
	if err := database.DB.Where("package_key IN ?", defaultPackageKeys).Find(&packages).Error; err != nil {
		return err
	}
	packageIDs := make([]uuid.UUID, 0, len(packages))
	for _, item := range packages {
		packageIDs = append(packageIDs, item.ID)
	}
	if err := refresher.RefreshByPackages(packageIDs); err != nil {
		return err
	}

	logger.Info("Default access snapshots refreshed", zap.Int("roles", len(roleIDs)), zap.Int("packages", len(packageIDs)))
	return nil
}

func initDefaultAPIEndpointCategories(logger *zap.Logger) error {
	if err := permissionseed.EnsureDefaultAPIEndpointCategories(database.DB); err != nil {
		return err
	}
	logger.Info("Default api endpoint categories seeded")
	return nil
}

func finalizeAPIEndpointSchema(logger *zap.Logger) error {
	if err := database.DB.Exec(`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS category_id uuid`).Error; err != nil {
		return err
	}
	if err := database.DB.Exec(`ALTER TABLE api_endpoint_permission_bindings ADD COLUMN IF NOT EXISTS endpoint_code varchar(36)`).Error; err != nil {
		return err
	}

	hasModule, err := hasColumn("api_endpoints", "module")
	if err != nil {
		return err
	}
	if hasModule {
		if err := database.DB.Exec(`
			UPDATE api_endpoints ae
			   SET category_id = c.id
			  FROM api_endpoint_categories c
			 WHERE ae.category_id IS NULL
			   AND c.code = ae.module
		`).Error; err != nil {
			return err
		}
	}
	hasEndpointID, err := hasColumn("api_endpoint_permission_bindings", "endpoint_id")
	if err != nil {
		return err
	}
	if hasEndpointID {
		if err := database.DB.Exec(`
			UPDATE api_endpoint_permission_bindings b
			   SET endpoint_code = ae.code
			  FROM api_endpoints ae
			 WHERE (b.endpoint_code IS NULL OR b.endpoint_code = '')
			   AND ae.id = b.endpoint_id
		`).Error; err != nil {
			return err
		}
	}
	if err := database.DB.Exec(`DELETE FROM api_endpoint_permission_bindings WHERE COALESCE(endpoint_code, '') = ''`).Error; err != nil {
		return err
	}

	statements := []string{
		`ALTER TABLE api_endpoints DROP COLUMN IF EXISTS group_code`,
		`ALTER TABLE api_endpoints DROP COLUMN IF EXISTS group_name`,
		`ALTER TABLE api_endpoints DROP COLUMN IF EXISTS module`,
		`ALTER TABLE api_endpoint_permission_bindings ALTER COLUMN endpoint_code SET NOT NULL`,
		`ALTER TABLE api_endpoint_permission_bindings DROP COLUMN IF EXISTS endpoint_id`,
	}
	for _, statement := range statements {
		if err := database.DB.Exec(statement).Error; err != nil {
			return err
		}
	}
	logger.Info("API endpoint schema finalized")
	return nil
}

func syncAPIRegistry(logger *zap.Logger, cfg *config.Config) error {
	router := apirouter.SetupRouter(cfg, logger, database.DB)
	builder := permissionseed.NewDeploymentBuilder(database.DB, logger, router).
		WithCoreDefaults()
	builder.LogSummary()
	return nil
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		target := strings.TrimSpace(value)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		result = append(result, target)
	}
	return result
}
