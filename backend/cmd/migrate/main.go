package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	apirouter "github.com/gg-ecommerce/backend/internal/api/router"
	"github.com/gg-ecommerce/backend/internal/config"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
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

	// 初始化默认角色菜单关联
	if err := initDefaultRoleMenus(logger); err != nil {
		logger.Warn("Failed to initialize role-menus", zap.Error(err))
	} else {
		logger.Info("Role-menus initialized successfully")
	}

	if err := initDefaultPermissionActionsNoScope(logger); err != nil {
		logger.Warn("Failed to initialize permission actions", zap.Error(err))
	} else {
		logger.Info("Permission actions initialized successfully")
	}

	if err := initDefaultFeaturePackages(logger); err != nil {
		logger.Warn("Failed to initialize feature packages", zap.Error(err))
	} else {
		logger.Info("Feature packages initialized successfully")
	}

	if err := initDefaultRoleFeaturePackages(logger); err != nil {
		logger.Warn("Failed to initialize default role feature packages", zap.Error(err))
	} else {
		logger.Info("Default role feature packages initialized successfully")
	}

	if err := initDefaultRoleActionPermissionsNoScope(logger); err != nil {
		logger.Warn("Failed to initialize role action permissions", zap.Error(err))
	} else {
		logger.Info("Role action permissions initialized successfully")
	}

	if err := syncAPIRegistry(logger, cfg); err != nil {
		logger.Warn("Failed to sync API registry", zap.Error(err))
	} else {
		logger.Info("API registry synchronized successfully")
	}

	fmt.Println("✅ Migration completed successfully!")
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
			Name: "20260323_permission_key_consolidation",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE permission_actions ADD COLUMN IF NOT EXISTS permission_key varchar(150)`,
					`DROP INDEX IF EXISTS idx_permission_actions_resource_action_unique`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}

				var actions []usermodel.PermissionAction
				if err := database.DB.Order("created_at ASC, id ASC").Find(&actions).Error; err != nil {
					return err
				}

				canonicalByKey := make(map[string]uuid.UUID, len(actions))
				duplicateIDs := make([]uuid.UUID, 0)
				for _, action := range actions {
					mapping := permissionkey.FromLegacy(action.ResourceCode, action.ActionCode)
					targetKey := permissionkey.Normalize(action.PermissionKey)
					if targetKey == "" {
						targetKey = mapping.Key
					}
					targetResource := strings.TrimSpace(mapping.ResourceCode)
					if targetResource == "" {
						targetResource = strings.TrimSpace(action.ResourceCode)
					}
					targetAction := strings.TrimSpace(mapping.ActionCode)
					if targetAction == "" {
						targetAction = strings.TrimSpace(action.ActionCode)
					}
					if err := database.DB.Model(&usermodel.PermissionAction{}).
						Where("id = ?", action.ID).
						Updates(map[string]interface{}{
							"permission_key": targetKey,
							"resource_code":  targetResource,
							"action_code":    targetAction,
						}).Error; err != nil {
						return err
					}
					if canonicalID, exists := canonicalByKey[targetKey]; exists {
						if err := rebindPermissionActionReferences(action.ID, canonicalID); err != nil {
							return err
						}
						duplicateIDs = append(duplicateIDs, action.ID)
						continue
					}
					canonicalByKey[targetKey] = action.ID
				}

				if len(duplicateIDs) > 0 {
					if err := database.DB.Where("id IN ?", duplicateIDs).Delete(&usermodel.PermissionAction{}).Error; err != nil {
						return err
					}
				}

				finishStatements := []string{
					`UPDATE permission_actions SET permission_key = CONCAT(resource_code, '.', action_code) WHERE COALESCE(permission_key, '') = ''`,
					`ALTER TABLE permission_actions ALTER COLUMN permission_key SET NOT NULL`,
					`DROP INDEX IF EXISTS idx_permission_actions_permission_key`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_permission_actions_permission_key ON permission_actions (permission_key) WHERE deleted_at IS NULL`,
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
				var actions []usermodel.PermissionAction
				if err := database.DB.Where("permission_key LIKE ?", "%:%").Order("created_at ASC, id ASC").Find(&actions).Error; err != nil {
					return err
				}
				if len(actions) == 0 {
					logger.Info("Named migration applied", zap.String("name", "20260324_permission_key_dot_cleanup"), zap.Int("updated", 0))
					return nil
				}

				canonicalByKey := map[string]uuid.UUID{}
				var existing []usermodel.PermissionAction
				if err := database.DB.Order("created_at ASC, id ASC").Find(&existing).Error; err != nil {
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
					mapping := permissionkey.FromLegacy(action.ResourceCode, action.ActionCode)
					targetKey := permissionkey.Normalize(mapping.Key)
					if targetKey == "" {
						continue
					}
					targetResource := strings.TrimSpace(mapping.ResourceCode)
					if targetResource == "" {
						targetResource = strings.TrimSpace(action.ResourceCode)
					}
					targetAction := strings.TrimSpace(mapping.ActionCode)
					if targetAction == "" {
						targetAction = strings.TrimSpace(action.ActionCode)
					}
					if canonicalID, exists := canonicalByKey[targetKey]; exists && canonicalID != action.ID {
						if err := rebindPermissionActionReferences(action.ID, canonicalID); err != nil {
							return err
						}
						if err := database.DB.Delete(&usermodel.PermissionAction{}, action.ID).Error; err != nil {
							return err
						}
						updatedCount++
						continue
					}
					if err := database.DB.Model(&usermodel.PermissionAction{}).
						Where("id = ?", action.ID).
						Updates(map[string]interface{}{
							"permission_key": targetKey,
							"resource_code":  targetResource,
							"action_code":    targetAction,
						}).Error; err != nil {
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
					if err := database.DB.Model(&usermodel.PermissionAction{}).
						Where("permission_key = ?", mapping.Key).
						Update("context_type", contextType).Error; err != nil {
						return err
					}
				}

				statements := []string{
					`UPDATE permission_actions SET context_type = 'platform' WHERE permission_key LIKE 'system.%'`,
					`UPDATE permission_actions SET context_type = 'platform' WHERE permission_key LIKE 'tenant.%'`,
					`UPDATE permission_actions SET context_type = 'platform' WHERE permission_key LIKE 'platform.%'`,
					`UPDATE permission_actions SET context_type = 'team' WHERE permission_key LIKE 'team.%'`,
					`UPDATE permission_actions SET permission_key = 'system.permission.manage', context_type = 'platform' WHERE permission_key IN ('system_permission.manage_action_registry', 'permission_action.manage')`,
					`UPDATE permission_actions SET permission_key = 'system.role.assign_action', context_type = 'platform' WHERE permission_key = 'system_permission.assign_role_action'`,
					`UPDATE permission_actions SET context_type = 'platform' WHERE resource_code IN ('role', 'user', 'menu', 'menu_backup', 'permission_action', 'api_endpoint', 'feature_package')`,
					`UPDATE permission_actions SET context_type = 'platform' WHERE permission_key IN ('feature_package.assign_action', 'feature_package.assign_menu', 'feature_package.assign_team')`,
					`UPDATE permission_actions SET context_type = 'team' WHERE permission_key IN ('team.configure_action_boundary', 'team.configure_menu_boundary')`,
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
			Name: "20260323_permission_system_backfill_defaults",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE permission_actions SET status = 'normal' WHERE COALESCE(status, '') = ''`,
					`UPDATE user_action_permissions SET effect = 'allow' WHERE COALESCE(effect, '') = ''`,
					`UPDATE tenant_action_permissions SET enabled = TRUE WHERE enabled IS NULL`,
					`UPDATE permission_actions SET source = 'system' WHERE COALESCE(source, '') = ''`,
					`UPDATE permission_actions SET module_code = COALESCE(NULLIF(module_code, ''), resource_code)`,
					`ALTER TABLE permission_actions ADD COLUMN IF NOT EXISTS context_type varchar(20)`,
					`UPDATE permission_actions SET context_type = 'platform' WHERE COALESCE(context_type, '') = '' AND (permission_key LIKE 'system.%' OR permission_key LIKE 'tenant.%' OR permission_key LIKE 'platform.%')`,
					`UPDATE permission_actions SET context_type = 'team' WHERE COALESCE(context_type, '') = ''`,
					`UPDATE permission_actions SET feature_kind = 'system' WHERE COALESCE(feature_kind, '') = ''`,
					`UPDATE permission_actions SET feature_kind = 'business' WHERE source = 'business' AND COALESCE(feature_kind, '') <> 'business'`,
					`UPDATE permission_actions SET source = 'system', feature_kind = 'system', module_code = 'system_permission' WHERE resource_code = 'system_permission'`,
					`UPDATE api_endpoints SET feature_kind = 'system' WHERE COALESCE(feature_kind, '') = ''`,
					`UPDATE permission_actions pa
					   SET source = 'api',
					       module_code = COALESCE(NULLIF(ae.module, ''), pa.module_code)
					  FROM api_endpoints ae
					 WHERE pa.resource_code = ae.resource_code
					   AND pa.action_code = ae.action_code
					   AND COALESCE(ae.resource_code, '') <> ''
					   AND COALESCE(ae.action_code, '') <> ''`,
				}
				if hasRoleActionEffect, err := hasColumn("role_action_permissions", "effect"); err != nil {
					return err
				} else if hasRoleActionEffect {
					statements = append(statements, `UPDATE role_action_permissions SET effect = 'allow' WHERE COALESCE(effect, '') = ''`)
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_permission_system_backfill_defaults"))
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
					`CREATE TABLE IF NOT EXISTS feature_package_actions (
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
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_package_actions_unique ON feature_package_actions (package_id, action_id)`,
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
			Name: "20260323_team_manual_action_schema",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS team_manual_action_permissions (
						tenant_id uuid NOT NULL,
						action_id uuid NOT NULL,
						enabled boolean NOT NULL DEFAULT TRUE,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_team_manual_action_permissions_unique ON team_manual_action_permissions (tenant_id, action_id)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_team_manual_action_schema"))
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
						manual_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
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
			Name: "20260323_team_manual_action_backfill",
			Run: func(logger *zap.Logger) error {
				statement := `
					INSERT INTO team_manual_action_permissions (tenant_id, action_id, enabled, created_at, updated_at)
					SELECT tap.tenant_id, tap.action_id, TRUE, NOW(), NOW()
					FROM tenant_action_permissions tap
					LEFT JOIN team_feature_packages tfp
						ON tfp.team_id = tap.tenant_id AND tfp.enabled = TRUE
					LEFT JOIN feature_package_actions fpa
						ON fpa.package_id = tfp.package_id AND fpa.action_id = tap.action_id
					WHERE tap.enabled = TRUE
					  AND fpa.action_id IS NULL
					ON CONFLICT (tenant_id, action_id) DO NOTHING
				`
				if err := database.DB.Exec(statement).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260323_team_manual_action_backfill"))
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
						updates["resource_code"] = strings.TrimSpace(mapping.ResourceCode)
					}
					if strings.TrimSpace(mapping.ActionCode) != "" {
						updates["action_code"] = strings.TrimSpace(mapping.ActionCode)
					}
					if strings.TrimSpace(mapping.ContextType) != "" {
						updates["context_type"] = strings.TrimSpace(mapping.ContextType)
					}
					if len(updates) == 0 {
						continue
					}
					if err := database.DB.Model(&usermodel.PermissionAction{}).
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
				if err := database.DB.Exec(`ALTER TABLE permission_actions DROP COLUMN IF EXISTS category`).Error; err != nil {
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
					`ALTER TABLE permission_actions DROP COLUMN IF EXISTS requires_tenant_context`,
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
					`DELETE FROM role_menus WHERE menu_id IN (SELECT id FROM menus WHERE name = 'Scope' OR path = 'scope' OR component = '/system/scope')`,
					`DELETE FROM menus WHERE name = 'Scope' OR path = 'scope' OR component = '/system/scope'`,
					`DELETE FROM role_action_permissions WHERE action_id IN (SELECT id FROM permission_actions WHERE resource_code = 'scope')`,
					`DELETE FROM user_action_permissions WHERE action_id IN (SELECT id FROM permission_actions WHERE resource_code = 'scope')`,
					`DELETE FROM tenant_action_permissions WHERE action_id IN (SELECT id FROM permission_actions WHERE resource_code = 'scope')`,
					`DELETE FROM permission_actions WHERE resource_code = 'scope'`,
					`DELETE FROM api_endpoints WHERE resource_code = 'scope'`,
					`ALTER TABLE role_data_permissions ADD COLUMN IF NOT EXISTS data_scope varchar(30)`,
					`ALTER TABLE roles DROP COLUMN IF EXISTS scope_id`,
					`ALTER TABLE permission_actions DROP COLUMN IF EXISTS scope_id`,
					`ALTER TABLE permission_actions DROP COLUMN IF EXISTS scope_type`,
					`ALTER TABLE api_endpoints DROP COLUMN IF EXISTS scope_id`,
					`DROP INDEX IF EXISTS idx_permission_actions_scope_id`,
					`DROP INDEX IF EXISTS idx_permission_actions_resource_action_unique`,
					`DROP INDEX IF EXISTS idx_permission_actions_permission_key`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_permission_actions_permission_key ON permission_actions (permission_key) WHERE deleted_at IS NULL`,
					`DROP TABLE IF EXISTS role_scopes`,
					`DROP TABLE IF EXISTS scopes`,
				}
				if hasRoleActionEffect, err := hasColumn("role_action_permissions", "effect"); err != nil {
					return err
				} else if hasRoleActionEffect {
					statements = append([]string{
						`DELETE FROM role_action_permissions WHERE COALESCE(effect, '') = 'deny'`,
						`UPDATE role_action_permissions SET effect = 'allow' WHERE COALESCE(effect, '') = ''`,
					}, statements...)
					statements = append(statements, `ALTER TABLE role_action_permissions DROP COLUMN IF EXISTS effect`)
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

func rebindPermissionActionReferences(fromID, toID uuid.UUID) error {
	if fromID == toID {
		return nil
	}
	statements := []struct {
		sql  string
		args []interface{}
	}{
		{
			sql: `UPDATE role_action_permissions target
				    SET action_id = ?
				  WHERE action_id = ?
				    AND NOT EXISTS (
				      SELECT 1 FROM role_action_permissions existing
				      WHERE existing.role_id = target.role_id AND existing.action_id = ?
				    )`,
			args: []interface{}{toID, fromID, toID},
		},
		{
			sql:  `DELETE FROM role_action_permissions WHERE action_id = ?`,
			args: []interface{}{fromID},
		},
		{
			sql: `UPDATE tenant_action_permissions target
				    SET action_id = ?
				  WHERE action_id = ?
				    AND NOT EXISTS (
				      SELECT 1 FROM tenant_action_permissions existing
				      WHERE existing.tenant_id = target.tenant_id AND existing.action_id = ?
				    )`,
			args: []interface{}{toID, fromID, toID},
		},
		{
			sql:  `DELETE FROM tenant_action_permissions WHERE action_id = ?`,
			args: []interface{}{fromID},
		},
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
	}
	for _, statement := range statements {
		if err := database.DB.Exec(statement.sql, statement.args...).Error; err != nil {
			return err
		}
	}
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

	metaSuperAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}}
	metaSuperAdminAndAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER", "R_ADMIN"}}
	metaTeamAdminOnly := usermodel.MetaJSON{
		"roles":     []interface{}{"R_ADMIN"},
		"keepAlive": true,
	}

	specs := []menuSeedSpec{
		{Name: "Dashboard", Path: "/dashboard", Component: "/index/index", Title: "menus.dashboard.title", Icon: "ri:pie-chart-line", SortOrder: 1, Meta: metaSuperAdminAndAdmin},
		{Name: "System", Path: "/system", Component: "/index/index", Title: "menus.system.title", Icon: "ri:user-3-line", SortOrder: 2, Meta: metaSuperAdminAndAdmin},
		{Name: "Result", Path: "/result", Component: "/index/index", Title: "menus.result.title", Icon: "ri:checkbox-circle-line", SortOrder: 3},
		{Name: "Exception", Path: "/exception", Component: "/index/index", Title: "menus.exception.title", Icon: "ri:error-warning-line", SortOrder: 4},
		{Name: "TeamRoot", Path: "/team", Component: "/index/index", Title: "menus.system.team", Icon: "ri:team-line", SortOrder: 5, Meta: metaSuperAdminAndAdmin},

		{Name: "Console", ParentName: "Dashboard", Path: "console", Component: "/dashboard/console", Title: "menus.dashboard.console", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": false, "fixedTab": true}},
		{Name: "UserCenter", ParentName: "Dashboard", Path: "user-center", Component: "/system/user-center", Title: "menus.system.userCenter", SortOrder: 2, Meta: usermodel.MetaJSON{"isHide": true, "keepAlive": true, "isHideTab": true}},

		{Name: "Role", ParentName: "System", Path: "role", Component: "/system/role", Title: "menus.system.role", SortOrder: 1, Meta: metaSuperAdmin},
		{Name: "User", ParentName: "System", Path: "user", Component: "/system/user", Title: "menus.system.user", SortOrder: 2, Meta: metaSuperAdminAndAdmin},
		{Name: "TeamRolesAndPermissions", ParentName: "System", Path: "team-roles-permissions", Component: "/system/team-roles-permissions", Title: "menus.system.teamRolesAndPermissions", SortOrder: 3, Meta: metaSuperAdmin},
		{Name: "Menus", ParentName: "System", Path: "menu", Component: "/system/menu", Title: "menus.system.menu", SortOrder: 4, Meta: usermodel.MetaJSON{
			"roles":     []interface{}{"R_SUPER"},
			"keepAlive": true,
		}},
		{Name: "ActionPermission", ParentName: "System", Path: "action-permission", Component: "/system/action-permission", Title: "功能权限", SortOrder: 5, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "ApiEndpoint", ParentName: "System", Path: "api-endpoint", Component: "/system/api-endpoint", Title: "API管理", SortOrder: 6, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "FeaturePackage", ParentName: "System", Path: "feature-package", Component: "/system/feature-package", Title: "功能包管理", SortOrder: 7, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},

		{Name: "TeamManagement", ParentName: "TeamRoot", Path: "team", Component: "/team/team", Title: "团队管理", SortOrder: 1, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "TeamMembers", ParentName: "TeamRoot", Path: "members", Component: "/team/team-members", Title: "menus.system.teamMembers", SortOrder: 2, Meta: metaTeamAdminOnly},

		{Name: "ResultSuccess", ParentName: "Result", Path: "success", Component: "/result/success", Title: "menus.result.success", Icon: "ri:checkbox-circle-line", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": true}},
		{Name: "ResultFail", ParentName: "Result", Path: "fail", Component: "/result/fail", Title: "menus.result.fail", Icon: "ri:close-circle-line", SortOrder: 2, Meta: usermodel.MetaJSON{"keepAlive": true}},

		{Name: "Exception403", ParentName: "Exception", Path: "403", Component: "/exception/403", Title: "menus.exception.forbidden", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}},
		{Name: "Exception404", ParentName: "Exception", Path: "404", Component: "/exception/404", Title: "menus.exception.notFound", SortOrder: 2, Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}},
		{Name: "Exception500", ParentName: "Exception", Path: "500", Component: "/exception/500", Title: "menus.exception.serverError", SortOrder: 3, Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}},
	}

	for _, spec := range specs {
		if _, err := ensureMenuSeed(spec); err != nil {
			return err
		}
	}

	logger.Info("Default menus ensured")
	return nil
}

type menuSeedSpec struct {
	Name       string
	ParentName string
	Path       string
	Component  string
	Title      string
	Icon       string
	SortOrder  int
	Meta       usermodel.MetaJSON
}

func ensureMenuSeed(spec menuSeedSpec) (*usermodel.Menu, error) {
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

func cleanupDeprecatedMenus(logger *zap.Logger) error {
	deprecatedNames := []string{"PageAssociation", "TeamManagementRedirect", "Scope"}
	var deprecatedMenus []usermodel.Menu
	if err := database.DB.Where("name IN ?", deprecatedNames).Find(&deprecatedMenus).Error; err != nil {
		return err
	}
	for _, menu := range deprecatedMenus {
		if err := database.DB.Where("menu_id = ?", menu.ID).Delete(&usermodel.RoleMenu{}).Error; err != nil {
			return err
		}
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
	if err := database.DB.Where("menu_id IN ?", collected).Delete(&usermodel.RoleMenu{}).Error; err != nil {
		return err
	}
	if err := database.DB.Where("menu_id IN ?", collected).Delete(&usermodel.FeaturePackageMenu{}).Error; err != nil {
		return err
	}
	return database.DB.Where("id IN ?", collected).Delete(&usermodel.Menu{}).Error
}

func initDefaultRoleMenus(logger *zap.Logger) error {
	var existingCount int64
	if err := database.DB.Model(&usermodel.RoleMenu{}).Count(&existingCount).Error; err != nil {
		logger.Error("Failed to count role-menus", zap.Error(err))
		return err
	}
	if existingCount > 0 {
		logger.Info("Role-menus already exist, skip default seed to preserve configured permissions", zap.Int64("count", existingCount))
		return nil
	}

	var roles []usermodel.Role
	if err := database.DB.Where("tenant_id IS NULL AND code IN ?", []string{"admin", "team_admin", "team_member"}).Find(&roles).Error; err != nil {
		return err
	}
	roleByCode := make(map[string]uuid.UUID)
	for _, r := range roles {
		roleByCode[r.Code] = r.ID
	}
	adminID, hasAdmin := roleByCode["admin"]
	teamAdminID, hasTeamAdmin := roleByCode["team_admin"]
	teamMemberID, hasTeamMember := roleByCode["team_member"]

	var menus []usermodel.Menu
	if err := database.DB.Find(&menus).Error; err != nil {
		return err
	}

	roleMenus := make([]usermodel.RoleMenu, 0)
	for _, m := range menus {
		rolesVal, _ := m.Meta["roles"].([]interface{})
		addAdmin, addTeamAdmin, addTeamMember := false, false, false
		if len(rolesVal) == 0 {
			addAdmin, addTeamAdmin, addTeamMember = true, true, true
		} else {
			for _, r := range rolesVal {
				s, _ := r.(string)
				switch s {
				case "R_SUPER":
					addAdmin = true
				case "R_ADMIN":
					addAdmin = true
					addTeamAdmin = true
				case "R_USER":
					addTeamMember = true
				}
			}
		}
		if addAdmin && hasAdmin {
			roleMenus = append(roleMenus, usermodel.RoleMenu{RoleID: adminID, MenuID: m.ID})
		}
		if addTeamAdmin && hasTeamAdmin {
			roleMenus = append(roleMenus, usermodel.RoleMenu{RoleID: teamAdminID, MenuID: m.ID})
		}
		if addTeamMember && hasTeamMember {
			roleMenus = append(roleMenus, usermodel.RoleMenu{RoleID: teamMemberID, MenuID: m.ID})
		}
	}
	seen := make(map[string]struct{})
	for _, rm := range roleMenus {
		key := rm.RoleID.String() + ":" + rm.MenuID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if err := database.DB.Create(&rm).Error; err != nil {
			return err
		}
	}
	logger.Info("Default role-menus seeded")
	return nil
}

func initDefaultPermissionActionsNoScope(logger *zap.Logger) error {
	for index, actionSeed := range defaultPermissionActionSeeds() {
		actionData := usermodel.PermissionAction{
			PermissionKey: actionSeed.PermissionKey,
			ResourceCode:  actionSeed.ResourceCode,
			ActionCode:    actionSeed.ActionCode,
			ModuleCode:    actionSeed.ModuleCode,
			ContextType:   actionSeed.ContextType,
			Source:        actionSeed.Source,
			FeatureKind:   actionSeed.FeatureKind,
			Name:          actionSeed.Name,
			Description:   actionSeed.Description,
			Status:        "normal",
			SortOrder:     index + 1,
		}
		var action usermodel.PermissionAction
		result := database.DB.Where("permission_key = ?", actionData.PermissionKey).First(&action)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := database.DB.Create(&actionData).Error; err != nil {
					return err
				}
				continue
			}
			return result.Error
		}
		updates := map[string]interface{}{
			"permission_key": actionData.PermissionKey,
			"resource_code":  actionData.ResourceCode,
			"action_code":    actionData.ActionCode,
			"name":           actionData.Name,
			"description":    actionData.Description,
			"module_code":    actionData.ModuleCode,
			"context_type":   actionData.ContextType,
			"source":         actionData.Source,
			"feature_kind":   actionData.FeatureKind,
			"status":         actionData.Status,
			"sort_order":     actionData.SortOrder,
		}
		if err := database.DB.Model(&action).Updates(updates).Error; err != nil {
			return err
		}
	}
	logger.Info("Default permission actions seeded")
	return nil
}

func initDefaultRoleActionPermissionsNoScope(logger *zap.Logger) error {
	var roles []usermodel.Role
	if err := database.DB.Where("tenant_id IS NULL AND code IN ?", []string{"admin", "team_admin"}).Find(&roles).Error; err != nil {
		return err
	}
	roleByCode := make(map[string]usermodel.Role, len(roles))
	for _, role := range roles {
		roleByCode[role.Code] = role
	}

	var actions []usermodel.PermissionAction
	if err := database.DB.Find(&actions).Error; err != nil {
		return err
	}

	defaultAssignments := map[string]map[string]struct{}{
		"admin":      {},
		"team_admin": {},
	}
	for _, action := range actions {
		key := strings.TrimSpace(action.PermissionKey)
		if key == "" {
			key = action.ResourceCode + ":" + action.ActionCode
		}
		defaultAssignments["admin"][key] = struct{}{}
		switch key {
		case "tenant.manage", "tenant.boundary.manage", "tenant.member.manage", "team.member.manage", "team.member.assign_role", "team.member.assign_action", "team.boundary.manage":
			defaultAssignments["team_admin"][key] = struct{}{}
		}
	}

	for roleCode, assignments := range defaultAssignments {
		role, ok := roleByCode[roleCode]
		if !ok {
			continue
		}
		for _, action := range actions {
			key := strings.TrimSpace(action.PermissionKey)
			if key == "" {
				key = action.ResourceCode + ":" + action.ActionCode
			}
			if _, allowed := assignments[key]; !allowed {
				continue
			}
			record := usermodel.RoleActionPermission{
				RoleID:   role.ID,
				ActionID: action.ID,
			}
			if err := database.DB.Where("role_id = ? AND action_id = ?", role.ID, action.ID).FirstOrCreate(&record).Error; err != nil {
				return err
			}
		}
	}

	logger.Info("Default role action permissions ensured")
	return nil
}

type permissionActionSeed struct {
	PermissionKey string
	ResourceCode  string
	ActionCode    string
	ModuleCode    string
	ContextType   string
	Source        string
	FeatureKind   string
	Name          string
	Description   string
}

type featurePackageSeed struct {
	PackageKey     string
	PackageType    string
	Name           string
	Description    string
	ContextType    string
	IsBuiltin      bool
	Status         string
	SortOrder      int
	MenuNames      []string
	PermissionKeys []string
}

func defaultPermissionActionSeeds() []permissionActionSeed {
	return []permissionActionSeed{
		newPermissionActionSeed("role", "list", "查看角色列表", "允许查看角色列表"),
		newPermissionActionSeed("role", "get", "查看角色详情", "允许查看角色详情"),
		newPermissionActionSeed("role", "create", "创建角色", "允许创建角色"),
		newPermissionActionSeed("role", "update", "更新角色", "允许更新角色"),
		newPermissionActionSeed("role", "delete", "删除角色", "允许删除角色"),
		newPermissionActionSeed("role", "assign_menu", "配置角色菜单权限", "允许为角色配置菜单权限"),
		newPermissionActionSeed("role", "assign_action", "配置角色功能权限", "允许为角色配置功能权限"),
		newPermissionActionSeed("role", "assign_data", "配置角色数据权限", "允许为角色配置数据权限"),
		newPermissionActionSeed("permission_action", "list", "查看功能权限列表", "允许查看功能权限列表"),
		newPermissionActionSeed("permission_action", "get", "查看功能权限详情", "允许查看功能权限详情"),
		newPermissionActionSeed("permission_action", "create", "创建功能权限", "允许创建功能权限"),
		newPermissionActionSeed("permission_action", "update", "更新功能权限", "允许更新功能权限"),
		newPermissionActionSeed("permission_action", "delete", "删除功能权限", "允许删除功能权限"),
		newPermissionActionSeed("user", "list", "查看用户列表", "允许查看用户列表"),
		newPermissionActionSeed("user", "get", "查看用户详情", "允许查看用户详情"),
		newPermissionActionSeed("user", "create", "创建用户", "允许创建用户"),
		newPermissionActionSeed("user", "update", "更新用户", "允许更新用户"),
		newPermissionActionSeed("user", "delete", "删除用户", "允许删除用户"),
		newPermissionActionSeed("user", "assign_role", "分配用户角色", "允许为用户分配角色"),
		newPermissionActionSeed("user", "assign_action", "配置用户功能权限", "允许为用户配置平台级功能权限"),
		newPermissionActionSeed("menu", "list", "查看菜单管理树", "允许查看全部菜单管理树"),
		newPermissionActionSeed("menu", "create", "创建菜单", "允许创建菜单"),
		newPermissionActionSeed("menu", "update", "更新菜单", "允许更新菜单"),
		newPermissionActionSeed("menu", "delete", "删除菜单", "允许删除菜单"),
		newPermissionActionSeed("menu_backup", "create", "创建菜单备份", "允许创建菜单备份"),
		newPermissionActionSeed("menu_backup", "list", "查看菜单备份列表", "允许查看菜单备份列表"),
		newPermissionActionSeed("menu_backup", "delete", "删除菜单备份", "允许删除菜单备份"),
		newPermissionActionSeed("menu_backup", "restore", "恢复菜单备份", "允许恢复菜单备份"),
		newPermissionActionSeed("system", "view_page_catalog", "查看页面文件映射", "允许查看页面文件映射"),
		newPermissionActionSeed("tenant", "list", "查看团队列表", "允许查看团队列表"),
		newPermissionActionSeed("tenant", "get", "查看团队详情", "允许查看团队详情"),
		newPermissionActionSeed("tenant", "create", "创建团队", "允许创建团队"),
		newPermissionActionSeed("tenant", "update", "更新团队", "允许更新团队"),
		newPermissionActionSeed("tenant", "delete", "删除团队", "允许删除团队"),
		newPermissionActionSeed("tenant", "configure_action_boundary", "配置团队功能权限边界", "允许配置团队功能权限边界"),
		newPermissionActionSeed("tenant_member_admin", "list", "查看团队成员列表", "允许在系统管理中查看团队成员列表"),
		newPermissionActionSeed("tenant_member_admin", "create", "添加团队成员", "允许在系统管理中添加团队成员"),
		newPermissionActionSeed("tenant_member_admin", "delete", "移除团队成员", "允许在系统管理中移除团队成员"),
		newPermissionActionSeed("tenant_member_admin", "update_role", "更新团队成员身份", "允许在系统管理中更新团队成员身份"),
		newPermissionActionSeed("team_member", "create", "添加当前团队成员", "允许在当前团队中添加成员"),
		newPermissionActionSeed("team_member", "delete", "移除当前团队成员", "允许在当前团队中移除成员"),
		newPermissionActionSeed("team_member", "update_role", "更新当前团队成员身份", "允许在当前团队中更新成员身份"),
		newPermissionActionSeed("team_member", "assign_role", "配置当前团队成员角色", "允许在当前团队中配置成员角色"),
		newPermissionActionSeed("team_member", "assign_action", "配置当前团队成员功能权限", "允许在当前团队中配置成员功能权限"),
		newPermissionActionSeed("team", "configure_action_boundary", "查看和配置当前团队功能权限边界", "允许查看和配置当前团队功能权限边界"),
		newPermissionActionSeed("api_endpoint", "list", "查看 API 注册表", "允许查看 API 注册表"),
		newPermissionActionSeed("api_endpoint", "sync", "同步 API 注册表", "允许同步 API 注册表"),
		newPermissionActionSeed("feature_package", "list", "查看功能包列表", "允许查看功能包列表"),
		newPermissionActionSeed("feature_package", "get", "查看功能包详情", "允许查看功能包详情"),
		newPermissionActionSeed("feature_package", "create", "创建功能包", "允许创建功能包"),
		newPermissionActionSeed("feature_package", "update", "更新功能包", "允许更新功能包"),
		newPermissionActionSeed("feature_package", "delete", "删除功能包", "允许删除功能包"),
		newPermissionActionSeed("feature_package", "assign_action", "配置功能包权限", "允许配置功能包包含的功能权限"),
		newPermissionActionSeed("feature_package", "assign_team", "配置团队功能包", "允许给团队开通功能包"),
		newPermissionActionSeed("system_permission", "manage_action_registry", "管理功能权限注册表", "允许维护功能权限注册信息"),
		newPermissionActionSeed("system_permission", "assign_role_action", "配置角色功能权限", "允许为角色分配功能权限"),
	}
}

func newPermissionActionSeed(
	resourceCode, actionCode, name, description string,
) permissionActionSeed {
	mapping := permissionkey.FromLegacy(resourceCode, actionCode)
	moduleCode := strings.TrimSpace(mapping.ResourceCode)
	if moduleCode == "" {
		moduleCode = strings.TrimSpace(resourceCode)
	}
	displayName := strings.TrimSpace(mapping.Name)
	if displayName == "" {
		displayName = name
	}
	displayDescription := strings.TrimSpace(mapping.Description)
	if displayDescription == "" {
		displayDescription = description
	}
	return permissionActionSeed{
		PermissionKey: mapping.Key,
		ResourceCode:  mapping.ResourceCode,
		ActionCode:    mapping.ActionCode,
		ModuleCode:    moduleCode,
		ContextType:   strings.TrimSpace(mapping.ContextType),
		Source:        "system",
		FeatureKind:   "system",
		Name:          displayName,
		Description:   displayDescription,
	}
}

func defaultFeaturePackageSeeds() []featurePackageSeed {
	return []featurePackageSeed{
		{
			PackageKey:     "platform.system_admin",
			PackageType:    "base",
			Name:           "平台系统管理包",
			Description:    "包含平台系统管理核心能力",
			ContextType:    "platform",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      1,
			MenuNames:      []string{"System", "Role", "User", "ActionPermission", "TeamRolesAndPermissions"},
			PermissionKeys: []string{"system.user.manage", "system.role.manage", "system.permission.manage"},
		},
		{
			PackageKey:     "platform.menu_admin",
			PackageType:    "base",
			Name:           "平台菜单管理包",
			Description:    "包含菜单管理与菜单备份能力",
			ContextType:    "platform",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      2,
			MenuNames:      []string{"System", "Menus"},
			PermissionKeys: []string{"system.menu.manage", "system.menu.backup"},
		},
		{
			PackageKey:     "platform.api_admin",
			PackageType:    "base",
			Name:           "平台接口管理包",
			Description:    "包含 API 注册表查看与同步能力",
			ContextType:    "platform",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      3,
			MenuNames:      []string{"System", "ApiEndpoint", "FeaturePackage"},
			PermissionKeys: []string{"system.api_registry.view", "system.api_registry.sync", "platform.package.manage", "platform.package.assign"},
		},
		{
			PackageKey:     "team.member_admin",
			PackageType:    "base",
			Name:           "团队成员管理包",
			Description:    "包含团队成员、角色和功能权限配置能力",
			ContextType:    "team",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      10,
			MenuNames:      []string{"TeamRoot", "TeamManagement", "TeamMembers"},
			PermissionKeys: []string{"team.member.manage", "team.member.assign_role", "team.member.assign_action", "team.boundary.manage"},
		},
	}
}

func initDefaultFeaturePackages(logger *zap.Logger) error {
	for _, seed := range defaultFeaturePackageSeeds() {
		item := usermodel.FeaturePackage{
			PackageKey:  seed.PackageKey,
			PackageType: seed.PackageType,
			Name:        seed.Name,
			Description: seed.Description,
			ContextType: seed.ContextType,
			IsBuiltin:   seed.IsBuiltin,
			Status:      seed.Status,
			SortOrder:   seed.SortOrder,
		}

		var existing usermodel.FeaturePackage
		result := database.DB.Where("package_key = ?", item.PackageKey).First(&existing)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := database.DB.Create(&item).Error; err != nil {
					return err
				}
				existing = item
			} else {
				return result.Error
			}
		} else {
			if err := database.DB.Model(&existing).Updates(map[string]interface{}{
				"name":         item.Name,
				"description":  item.Description,
				"package_type": item.PackageType,
				"context_type": item.ContextType,
				"is_builtin":   item.IsBuiltin,
				"status":       item.Status,
				"sort_order":   item.SortOrder,
			}).Error; err != nil {
				return err
			}
		}

		actionIDs := make([]uuid.UUID, 0, len(seed.PermissionKeys))
		for _, permissionKey := range seed.PermissionKeys {
			var action usermodel.PermissionAction
			if err := database.DB.Where("permission_key = ?", permissionKey).First(&action).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			actionIDs = append(actionIDs, action.ID)
		}

		if err := database.DB.Where("package_id = ?", existing.ID).Delete(&usermodel.FeaturePackageAction{}).Error; err != nil {
			return err
		}
		seen := make(map[uuid.UUID]struct{}, len(actionIDs))
		records := make([]usermodel.FeaturePackageAction, 0, len(actionIDs))
		for _, actionID := range actionIDs {
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			records = append(records, usermodel.FeaturePackageAction{PackageID: existing.ID, ActionID: actionID})
		}
		if len(records) > 0 {
			if err := database.DB.Create(&records).Error; err != nil {
				return err
			}
		}

		menuIDs := make([]uuid.UUID, 0, len(seed.MenuNames))
		for _, menuName := range seed.MenuNames {
			var menu usermodel.Menu
			if err := database.DB.Where("name = ?", menuName).First(&menu).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			menuIDs = append(menuIDs, menu.ID)
		}

		if err := database.DB.Where("package_id = ?", existing.ID).Delete(&usermodel.FeaturePackageMenu{}).Error; err != nil {
			return err
		}
		menuSeen := make(map[uuid.UUID]struct{}, len(menuIDs))
		menuRecords := make([]usermodel.FeaturePackageMenu, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			if _, ok := menuSeen[menuID]; ok {
				continue
			}
			menuSeen[menuID] = struct{}{}
			menuRecords = append(menuRecords, usermodel.FeaturePackageMenu{PackageID: existing.ID, MenuID: menuID})
		}
		if len(menuRecords) > 0 {
			if err := database.DB.Create(&menuRecords).Error; err != nil {
				return err
			}
		}
	}
	logger.Info("Default feature packages seeded")
	return nil
}

func initDefaultRoleFeaturePackages(logger *zap.Logger) error {
	roleCodes := []string{"admin", "team_admin"}
	packageKeys := []string{"platform.system_admin", "platform.menu_admin", "platform.api_admin", "team.member_admin"}

	var roles []usermodel.Role
	if err := database.DB.Where("tenant_id IS NULL AND code IN ?", roleCodes).Find(&roles).Error; err != nil {
		return err
	}
	roleByCode := make(map[string]usermodel.Role, len(roles))
	for _, role := range roles {
		roleByCode[role.Code] = role
	}

	var packages []usermodel.FeaturePackage
	if err := database.DB.Where("package_key IN ?", packageKeys).Find(&packages).Error; err != nil {
		return err
	}
	packageByKey := make(map[string]usermodel.FeaturePackage, len(packages))
	for _, item := range packages {
		packageByKey[item.PackageKey] = item
	}

	assignments := map[string][]string{
		"admin":      {"platform.system_admin", "platform.menu_admin", "platform.api_admin"},
		"team_admin": {"team.member_admin"},
	}

	for roleCode, keys := range assignments {
		role, ok := roleByCode[roleCode]
		if !ok {
			continue
		}
		for _, packageKey := range keys {
			pkg, ok := packageByKey[packageKey]
			if !ok {
				continue
			}
			record := usermodel.RoleFeaturePackage{
				RoleID:    role.ID,
				PackageID: pkg.ID,
				Enabled:   true,
			}
			if err := database.DB.Where("role_id = ? AND package_id = ?", role.ID, pkg.ID).FirstOrCreate(&record).Error; err != nil {
				return err
			}
			if err := database.DB.Model(&record).Update("enabled", true).Error; err != nil {
				return err
			}
		}
	}

	logger.Info("Default role feature packages ensured")
	return nil
}

func syncAPIRegistry(logger *zap.Logger, cfg *config.Config) error {
	apirouter.SetupRouter(cfg, logger, database.DB)
	return nil
}
