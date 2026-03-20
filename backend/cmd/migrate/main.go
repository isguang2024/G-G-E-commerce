package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	apirouter "github.com/gg-ecommerce/backend/internal/api/router"
	"github.com/gg-ecommerce/backend/internal/config"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
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
	if err := initDefaultScopes(logger); err != nil {
		logger.Warn("initDefaultScopes failed", zap.Error(err))
	}

	if err := runNamedMigrations(logger); err != nil {
		logger.Fatal("Named migrations failed", zap.Error(err))
	}

	// 初始化默认角色
	if err := initDefaultRoles(logger); err != nil {
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
	if err := initDefaultMenus(logger); err != nil {
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

	if err := initDefaultPermissionActions(logger); err != nil {
		logger.Warn("Failed to initialize permission actions", zap.Error(err))
	} else {
		logger.Info("Permission actions initialized successfully")
	}

	if err := initDefaultRoleActionPermissions(logger); err != nil {
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
			Name: "20260317_permission_system_backfill_defaults",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE permission_actions SET status = 'normal' WHERE COALESCE(status, '') = ''`,
					`UPDATE user_action_permissions SET effect = 'allow' WHERE COALESCE(effect, '') = ''`,
					`UPDATE tenant_action_permissions SET enabled = TRUE WHERE enabled IS NULL`,
				}
				if hasRoleActionEffect, err := hasColumn("role_action_permissions", "effect"); err != nil {
					return err
				} else if hasRoleActionEffect {
					statements = append(statements, `UPDATE role_action_permissions SET effect = 'allow' WHERE COALESCE(effect, '') = ''`)
				}
				if hasLegacyScopeType, err := hasColumn("permission_actions", "scope_type"); err != nil {
					return err
				} else if hasLegacyScopeType {
					statements = append([]string{
						`UPDATE permission_actions SET scope_type = 'global' WHERE COALESCE(scope_type, '') = ''`,
					}, statements...)
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_permission_system_backfill_defaults"))
				return nil
			},
		},
		{
			Name: "20260317_permission_actions_unify_scope_with_scopes",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE permission_actions ADD COLUMN IF NOT EXISTS scope_id uuid`,
					`CREATE INDEX IF NOT EXISTS idx_permission_actions_scope_id ON permission_actions (scope_id)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				if hasLegacyScopeType, err := hasColumn("permission_actions", "scope_type"); err != nil {
					return err
				} else if hasLegacyScopeType {
					if err := database.DB.Exec(`
						UPDATE permission_actions pa
						SET scope_id = s.id
						FROM scopes s
						WHERE pa.scope_id IS NULL
						  AND (
						    (COALESCE(pa.scope_type, 'global') = 'global' AND s.code = 'global') OR
						    (COALESCE(pa.scope_type, 'global') IN ('team', 'tenant') AND s.code = 'team')
						  )
					`).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_permission_actions_unify_scope_with_scopes"))
				return nil
			},
		},
		{
			Name: "20260317_permission_actions_finalize_scope_id",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE permission_actions pa
					 SET scope_id = s.id
					 FROM scopes s
					 WHERE pa.scope_id IS NULL
					   AND s.code = 'global'`,
					`CREATE INDEX IF NOT EXISTS idx_permission_actions_scope_id ON permission_actions (scope_id)`,
					`ALTER TABLE permission_actions ALTER COLUMN scope_id SET NOT NULL`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				if hasLegacyScopeType, err := hasColumn("permission_actions", "scope_type"); err != nil {
					return err
				} else if hasLegacyScopeType {
					if err := database.DB.Exec(`ALTER TABLE permission_actions DROP COLUMN scope_type`).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_permission_actions_finalize_scope_id"))
				return nil
			},
		},
		{
			Name: "20260317_permission_actions_expand_unique_scope",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`DROP INDEX IF EXISTS idx_permission_actions_resource_action_unique`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_permission_actions_resource_action_unique ON permission_actions (resource_code, action_code, scope_id)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_permission_actions_expand_unique_scope"))
				return nil
			},
		},
		{
			Name: "20260317_menus_remove_legacy_button_meta",
			Run: func(logger *zap.Logger) error {
				if err := database.DB.Exec(`
					UPDATE menus
					SET meta = COALESCE(meta, '{}'::jsonb) - 'authList' - 'authMark' - 'isAuthButton'
					WHERE meta ? 'authList' OR meta ? 'authMark' OR meta ? 'isAuthButton'
				`).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_menus_remove_legacy_button_meta"))
				return nil
			},
		},
		{
			Name: "20260317_permission_actions_backfill_source_category",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE permission_actions SET source = 'system' WHERE COALESCE(source, '') = ''`,
					`UPDATE permission_actions SET category = resource_code WHERE COALESCE(category, '') = ''`,
					`UPDATE permission_actions pa
					 SET source = 'api',
					     category = COALESCE(NULLIF(ae.module, ''), pa.category)
					 FROM api_endpoints ae
					 WHERE pa.resource_code = ae.resource_code
					   AND pa.action_code = ae.action_code
					   AND pa.scope_id = ae.scope_id
					   AND COALESCE(ae.resource_code, '') <> ''
					   AND COALESCE(ae.action_code, '') <> ''`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_permission_actions_backfill_source_category"))
				return nil
			},
		},
		{
			Name: "20260317_permission_actions_normalize_source_and_feature_kind",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE permission_actions SET source = 'business' WHERE source = 'manual'`,
					`UPDATE permission_actions SET feature_kind = 'system' WHERE COALESCE(feature_kind, '') = ''`,
					`UPDATE permission_actions SET feature_kind = 'business' WHERE source = 'business'`,
					`UPDATE permission_actions SET source = 'system', feature_kind = 'system', category = 'system_permission' WHERE resource_code = 'system_permission'`,
					`UPDATE api_endpoints SET feature_kind = 'system' WHERE COALESCE(feature_kind, '') = ''`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_permission_actions_normalize_source_and_feature_kind"))
				return nil
			},
		},
		{
			Name: "20260317_permission_actions_add_module_code",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE permission_actions ADD COLUMN IF NOT EXISTS module_code varchar(100)`,
					`UPDATE permission_actions pa
					   SET module_code = COALESCE(NULLIF(ae.module, ''), NULLIF(pa.category, ''), pa.resource_code)
					  FROM api_endpoints ae
					 WHERE COALESCE(pa.module_code, '') = ''
					   AND pa.resource_code = ae.resource_code
					   AND pa.action_code = ae.action_code
					   AND pa.scope_id = ae.scope_id
					   AND COALESCE(ae.resource_code, '') <> ''
					   AND COALESCE(ae.action_code, '') <> ''`,
					`UPDATE permission_actions
					    SET module_code = COALESCE(NULLIF(category, ''), resource_code)
					  WHERE COALESCE(module_code, '') = ''`,
					`ALTER TABLE permission_actions ALTER COLUMN module_code SET NOT NULL`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260317_permission_actions_add_module_code"))
				return nil
			},
		},
		{
			Name: "20260318_restore_permission_scope_schema",
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
				if hasRoleScopeBindings, err := hasTable("role_scope_bindings"); err != nil {
					return err
				} else if hasRoleScopeBindings {
					if hasRoleScopeID, columnErr := hasColumn("roles", "scope_id"); columnErr != nil {
						return columnErr
					} else if hasRoleScopeID {
						statements = append(statements, `UPDATE roles r
						  SET scope_id = rsb.scope_id
						  FROM role_scope_bindings rsb
						  WHERE r.id = rsb.role_id AND r.scope_id IS NULL`)
					}
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
				)
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260318_restore_permission_scope_schema"))
				return nil
			},
		},
		{
			Name: "20260318_drop_unused_core_columns",
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
				logger.Info("Named migration applied", zap.String("name", "20260318_drop_unused_core_columns"))
				return nil
			},
		},
		{
			Name: "20260318_roles_use_role_scopes",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS role_scopes (
						role_id uuid NOT NULL,
						scope_id uuid NOT NULL,
						PRIMARY KEY (role_id, scope_id)
					)`,
					`CREATE INDEX IF NOT EXISTS idx_role_scopes_role_id ON role_scopes (role_id)`,
					`CREATE INDEX IF NOT EXISTS idx_role_scopes_scope_id ON role_scopes (scope_id)`,
				}
				if hasRoleScopeID, err := hasColumn("roles", "scope_id"); err != nil {
					return err
				} else if hasRoleScopeID {
					statements = append(statements,
						`INSERT INTO role_scopes (role_id, scope_id)
						 SELECT id, scope_id FROM roles WHERE scope_id IS NOT NULL
						 ON CONFLICT DO NOTHING`,
						`ALTER TABLE roles DROP COLUMN IF EXISTS scope_id`,
					)
				}
				statements = append(statements, `DROP TABLE IF EXISTS role_scope_bindings`)
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260318_roles_use_role_scopes"))
				return nil
			},
		},
		{
			Name: "20260318_role_action_permissions_remove_effect",
			Run: func(logger *zap.Logger) error {
				if hasRoleActionEffect, err := hasColumn("role_action_permissions", "effect"); err != nil {
					return err
				} else if !hasRoleActionEffect {
					logger.Info("Named migration applied", zap.String("name", "20260318_role_action_permissions_remove_effect"))
					return nil
				}

				statements := []string{
					`DELETE FROM role_action_permissions WHERE COALESCE(effect, '') = 'deny'`,
					`UPDATE role_action_permissions SET effect = 'allow' WHERE COALESCE(effect, '') = ''`,
					`ALTER TABLE role_action_permissions DROP COLUMN IF EXISTS effect`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260318_role_action_permissions_remove_effect"))
				return nil
			},
		},
		{
			Name: "20260318_scope_rule_fields",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE scopes ADD COLUMN IF NOT EXISTS context_kind varchar(20)`,
					`ALTER TABLE scopes ADD COLUMN IF NOT EXISTS data_permission_code varchar(50)`,
					`ALTER TABLE scopes ADD COLUMN IF NOT EXISTS data_permission_name varchar(100)`,
					`UPDATE scopes SET context_kind = CASE WHEN code IN ('team', 'tenant') THEN 'tenant' ELSE 'global' END WHERE COALESCE(context_kind, '') = ''`,
					`UPDATE scopes SET data_permission_code = CASE WHEN code IN ('team', 'tenant') THEN 'team' ELSE '' END WHERE COALESCE(data_permission_code, '') = ''`,
					`UPDATE scopes SET data_permission_name = CASE WHEN code IN ('team', 'tenant') THEN name ELSE COALESCE(NULLIF(data_permission_name, ''), name) END WHERE COALESCE(data_permission_name, '') = ''`,
					`ALTER TABLE scopes ALTER COLUMN context_kind SET DEFAULT 'global'`,
					`ALTER TABLE scopes ALTER COLUMN data_permission_code SET DEFAULT ''`,
					`ALTER TABLE scopes ALTER COLUMN data_permission_name SET DEFAULT ''`,
					`ALTER TABLE scopes ALTER COLUMN context_kind SET NOT NULL`,
					`ALTER TABLE scopes ALTER COLUMN data_permission_code SET NOT NULL`,
					`ALTER TABLE scopes ALTER COLUMN data_permission_name SET NOT NULL`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260318_scope_rule_fields"))
				return nil
			},
		},
		{
			Name: "20260318_scope_system_flag",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE scopes ADD COLUMN IF NOT EXISTS is_system boolean`,
					`UPDATE scopes SET is_system = CASE WHEN code IN ('global', 'team') THEN TRUE ELSE FALSE END WHERE is_system IS NULL`,
					`ALTER TABLE scopes ALTER COLUMN is_system SET DEFAULT FALSE`,
					`ALTER TABLE scopes ALTER COLUMN is_system SET NOT NULL`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260318_scope_system_flag"))
				return nil
			},
		},
		{
			Name: "20260318_scopes_drop_context_kind",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE scopes SET code = 'team', name = '团队', is_system = TRUE WHERE code = 'tenant' AND NOT EXISTS (SELECT 1 FROM scopes existing WHERE existing.code = 'team')`,
					`WITH team_scope AS (
						SELECT id FROM scopes WHERE code = 'team' LIMIT 1
					), tenant_scope AS (
						SELECT id FROM scopes WHERE code = 'tenant' LIMIT 1
					)
					INSERT INTO role_scopes (role_id, scope_id)
					SELECT rs.role_id, ts.id
					FROM role_scopes rs, tenant_scope tns, team_scope ts
					WHERE rs.scope_id = tns.id
					ON CONFLICT DO NOTHING`,
					`WITH team_scope AS (
						SELECT id FROM scopes WHERE code = 'team' LIMIT 1
					), tenant_scope AS (
						SELECT id FROM scopes WHERE code = 'tenant' LIMIT 1
					)
					UPDATE permission_actions
					SET scope_id = ts.id
					FROM tenant_scope tns, team_scope ts
					WHERE permission_actions.scope_id = tns.id`,
					`WITH team_scope AS (
						SELECT id FROM scopes WHERE code = 'team' LIMIT 1
					), tenant_scope AS (
						SELECT id FROM scopes WHERE code = 'tenant' LIMIT 1
					)
					UPDATE api_endpoints
					SET scope_id = ts.id
					FROM tenant_scope tns, team_scope ts
					WHERE api_endpoints.scope_id = tns.id`,
					`DELETE FROM role_scopes WHERE scope_id IN (SELECT id FROM scopes WHERE code = 'tenant')`,
					`DELETE FROM scopes WHERE code = 'tenant'`,
					`ALTER TABLE scopes DROP COLUMN IF EXISTS context_kind`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260318_scopes_drop_context_kind"))
				return nil
			},
		},
		{
			Name: "20260319_remove_requires_tenant_context",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`UPDATE menus
					 SET meta = COALESCE(meta, '{}'::jsonb) - 'requiresTenantContext'
					 WHERE meta ? 'requiresTenantContext'`,
					`ALTER TABLE permission_actions DROP COLUMN IF EXISTS requires_tenant_context`,
					`ALTER TABLE api_endpoints DROP COLUMN IF EXISTS requires_tenant_context`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260319_remove_requires_tenant_context"))
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

func hasTable(tableName string) (bool, error) {
	var count int64
	err := database.DB.Raw(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = CURRENT_SCHEMA()
		  AND table_name = ?
	`, tableName).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func initDefaultRoles(logger *zap.Logger) error {
	var globalScope usermodel.Scope
	if err := database.DB.Where("code = ?", "global").First(&globalScope).Error; err != nil {
		logger.Error("Failed to find global scope", zap.Error(err))
		return err
	}
	var teamScope usermodel.Scope
	if err := database.DB.Where("code = ?", "team").First(&teamScope).Error; err != nil {
		logger.Error("Failed to find team scope", zap.Error(err))
		return err
	}

	roles := []struct {
		Code        string
		Name        string
		Description string
		ScopeCodes  []string
		SortOrder   int
	}{
		{"admin", "管理员", "系统管理员，拥有所有权限", []string{"global"}, 1},
		{"team_admin", "团队管理员", "团队管理员，可以管理团队成员和团队内容", []string{"team"}, 2},
		{"team_member", "团队成员", "团队成员，可以查看和编辑团队内容", []string{"team"}, 3},
	}

	_ = database.DB.Model(&usermodel.Scope{}).Where("code = ?", "global").Updates(map[string]interface{}{
		"data_permission_code": "",
		"data_permission_name": "全局",
	}).Error
	_ = database.DB.Model(&usermodel.Scope{}).Where("code IN ?", []string{"team", "tenant"}).Updates(map[string]interface{}{
		"data_permission_code": "team",
		"data_permission_name": "当前团队",
	}).Error

	for _, roleData := range roles {
		var role usermodel.Role
		result := database.DB.Where("code = ?", roleData.Code).First(&role)
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
				if err := ensureRoleScopeBindings(role.ID, roleData.ScopeCodes); err != nil {
					logger.Error("Failed to create role-scope binding", zap.String("code", roleData.Code), zap.Error(err))
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
			_ = ensureRoleScopeBindings(role.ID, roleData.ScopeCodes)
			logger.Info("Role already exists", zap.String("code", roleData.Code))
		}
	}
	return nil
}

func ensureRoleScopeBindings(roleID uuid.UUID, scopeCodes []string) error {
	scopeIDs, err := getScopeIDsByCode(scopeCodes)
	if err != nil {
		return err
	}
	for _, scopeID := range scopeIDs {
		if err := database.DB.Exec(
			`INSERT INTO role_scopes (role_id, scope_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
			roleID,
			scopeID,
		).Error; err != nil {
			return err
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
	if err := database.DB.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
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

func initDefaultScopes(logger *zap.Logger) error {
	scopes := []struct {
		Code        string
		Name        string
		Description string
		DataCode    string
		DataName    string
		IsSystem    bool
		SortOrder   int
	}{
		{"global", "全局", "跨应用全局作用域", "", "全局", true, 1},
		{"team", "团队", "仅团队功能使用的作用域", "team", "当前团队", true, 2},
	}

	for _, scopeData := range scopes {
		var scope usermodel.Scope
		result := database.DB.Where("code = ?", scopeData.Code).First(&scope)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				scope = usermodel.Scope{
					Code:               scopeData.Code,
					Name:               scopeData.Name,
					Description:        scopeData.Description,
					IsSystem:           scopeData.IsSystem,
					DataPermissionCode: scopeData.DataCode,
					DataPermissionName: scopeData.DataName,
					SortOrder:          scopeData.SortOrder,
				}
				if err := database.DB.Create(&scope).Error; err != nil {
					logger.Error("Failed to create scope", zap.String("code", scopeData.Code), zap.Error(err))
					return err
				}
				logger.Info("Scope created", zap.String("code", scopeData.Code), zap.String("name", scopeData.Name))
			} else {
				return result.Error
			}
		} else {
			updates := map[string]interface{}{
				"name":                 scopeData.Name,
				"description":          scopeData.Description,
				"is_system":            scopeData.IsSystem,
				"data_permission_code": scopeData.DataCode,
				"data_permission_name": scopeData.DataName,
				"sort_order":           scopeData.SortOrder,
			}
			if err := database.DB.Model(&scope).Updates(updates).Error; err != nil {
				return err
			}
			logger.Info("Scope already exists", zap.String("code", scopeData.Code))
		}
	}
	return nil
}

func initDefaultMenus(logger *zap.Logger) error {
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
		{Name: "Scope", ParentName: "System", Path: "scope", Component: "/system/scope", Title: "menus.system.scope", SortOrder: 2, Meta: metaSuperAdmin},
		{Name: "User", ParentName: "System", Path: "user", Component: "/system/user", Title: "menus.system.user", SortOrder: 3, Meta: metaSuperAdminAndAdmin},
		{Name: "TeamRolesAndPermissions", ParentName: "System", Path: "team-roles-permissions", Component: "/system/team-roles-permissions", Title: "menus.system.teamRolesAndPermissions", SortOrder: 4, Meta: metaSuperAdmin},
		{Name: "Menus", ParentName: "System", Path: "menu", Component: "/system/menu", Title: "menus.system.menu", SortOrder: 5, Meta: usermodel.MetaJSON{
			"roles":     []interface{}{"R_SUPER"},
			"keepAlive": true,
		}},
		{Name: "ActionPermission", ParentName: "System", Path: "action-permission", Component: "/system/action-permission", Title: "功能权限", SortOrder: 6, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "ApiEndpoint", ParentName: "System", Path: "api-endpoint", Component: "/system/api-endpoint", Title: "API管理", SortOrder: 7, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},

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
	deprecatedNames := []string{"PageAssociation", "TeamManagementRedirect"}
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
	return nil
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
	if err := database.DB.Where("code IN ?", []string{"admin", "team_admin", "team_member"}).Find(&roles).Error; err != nil {
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

func initDefaultPermissionActions(logger *zap.Logger) error {
	scopeIDs, err := getScopeIDsByCode([]string{"global", "team"})
	if err != nil {
		return err
	}
	for index, actionSeed := range defaultPermissionActionSeeds() {
		actionData := usermodel.PermissionAction{
			ResourceCode: actionSeed.ResourceCode,
			ActionCode:   actionSeed.ActionCode,
			ModuleCode:   actionSeed.ModuleCode,
			Category:     actionSeed.Category,
			Source:       actionSeed.Source,
			FeatureKind:  actionSeed.FeatureKind,
			Name:         actionSeed.Name,
			Description:  actionSeed.Description,
			ScopeID:      scopeIDs[actionSeed.ScopeCode],
			Status:       "normal",
			SortOrder:    index + 1,
		}
		var action usermodel.PermissionAction
		result := database.DB.Where("resource_code = ? AND action_code = ? AND scope_id = ?", actionData.ResourceCode, actionData.ActionCode, actionData.ScopeID).First(&action)
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
			"name":         actionData.Name,
			"description":  actionData.Description,
			"module_code":  actionData.ModuleCode,
			"category":     actionData.Category,
			"source":       actionData.Source,
			"feature_kind": actionData.FeatureKind,
			"scope_id":     actionData.ScopeID,
			"status":       actionData.Status,
			"sort_order":   actionData.SortOrder,
		}
		if err := database.DB.Model(&action).Updates(updates).Error; err != nil {
			return err
		}
	}
	logger.Info("Default permission actions seeded")
	return nil
}

func initDefaultRoleActionPermissions(logger *zap.Logger) error {
	var roles []usermodel.Role
	if err := database.DB.Preload("Scopes").Where("code IN ?", []string{"admin", "team_admin"}).Find(&roles).Error; err != nil {
		return err
	}
	roleByCode := make(map[string]usermodel.Role, len(roles))
	for _, role := range roles {
		roleByCode[role.Code] = role
	}

	var actions []usermodel.PermissionAction
	if err := database.DB.Preload("Scope").Find(&actions).Error; err != nil {
		return err
	}

	defaultAssignments := map[string]map[string]struct{}{
		"admin":      {},
		"team_admin": {},
	}
	for _, action := range actions {
		key := action.ResourceCode + ":" + action.ActionCode
		scopeCode := ""
		if action.Scope.ID != uuid.Nil {
			scopeCode = action.Scope.Code
		}
		switch scopeCode {
		case "global":
			defaultAssignments["admin"][key] = struct{}{}
		case "team":
			defaultAssignments["team_admin"][key] = struct{}{}
		}
	}

	for roleCode, assignments := range defaultAssignments {
		role, ok := roleByCode[roleCode]
		if !ok {
			continue
		}
		for _, action := range actions {
			key := action.ResourceCode + ":" + action.ActionCode
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
	ResourceCode          string
	ActionCode            string
	ModuleCode            string
	Category              string
	Source                string
	FeatureKind           string
	Name                  string
	Description           string
	ScopeCode             string
}

func defaultPermissionActionSeeds() []permissionActionSeed {
	return []permissionActionSeed{
		newPermissionActionSeed("role", "list", "查看角色列表", "允许查看角色列表", "global"),
		newPermissionActionSeed("role", "get", "查看角色详情", "允许查看角色详情", "global"),
		newPermissionActionSeed("role", "create", "创建角色", "允许创建角色", "global"),
		newPermissionActionSeed("role", "update", "更新角色", "允许更新角色", "global"),
		newPermissionActionSeed("role", "delete", "删除角色", "允许删除角色", "global"),
		newPermissionActionSeed("role", "assign_menu", "配置角色菜单权限", "允许为角色配置菜单权限", "global"),
		newPermissionActionSeed("role", "assign_action", "配置角色功能权限", "允许为角色配置功能权限", "global"),
		newPermissionActionSeed("role", "assign_data", "配置角色数据权限", "允许为角色配置数据权限", "global"),
		newPermissionActionSeed("permission_action", "list", "查看功能权限列表", "允许查看功能权限列表", "global"),
		newPermissionActionSeed("permission_action", "get", "查看功能权限详情", "允许查看功能权限详情", "global"),
		newPermissionActionSeed("permission_action", "create", "创建功能权限", "允许创建功能权限", "global"),
		newPermissionActionSeed("permission_action", "update", "更新功能权限", "允许更新功能权限", "global"),
		newPermissionActionSeed("permission_action", "delete", "删除功能权限", "允许删除功能权限", "global"),
		newPermissionActionSeed("scope", "list", "查看作用域列表", "允许查看作用域列表", "global"),
		newPermissionActionSeed("scope", "get", "查看作用域详情", "允许查看作用域详情", "global"),
		newPermissionActionSeed("scope", "create", "创建作用域", "允许创建作用域", "global"),
		newPermissionActionSeed("scope", "update", "更新作用域", "允许更新作用域", "global"),
		newPermissionActionSeed("scope", "delete", "删除作用域", "允许删除作用域", "global"),
		newPermissionActionSeed("user", "list", "查看用户列表", "允许查看用户列表", "global"),
		newPermissionActionSeed("user", "get", "查看用户详情", "允许查看用户详情", "global"),
		newPermissionActionSeed("user", "create", "创建用户", "允许创建用户", "global"),
		newPermissionActionSeed("user", "update", "更新用户", "允许更新用户", "global"),
		newPermissionActionSeed("user", "delete", "删除用户", "允许删除用户", "global"),
		newPermissionActionSeed("user", "assign_role", "分配用户角色", "允许为用户分配角色", "global"),
		newPermissionActionSeed("user", "assign_action", "配置用户功能权限", "允许为用户配置平台级功能权限", "global"),
		newPermissionActionSeed("menu", "list", "查看菜单管理树", "允许查看全部菜单管理树", "global"),
		newPermissionActionSeed("menu", "create", "创建菜单", "允许创建菜单", "global"),
		newPermissionActionSeed("menu", "update", "更新菜单", "允许更新菜单", "global"),
		newPermissionActionSeed("menu", "delete", "删除菜单", "允许删除菜单", "global"),
		newPermissionActionSeed("menu_backup", "create", "创建菜单备份", "允许创建菜单备份", "global"),
		newPermissionActionSeed("menu_backup", "list", "查看菜单备份列表", "允许查看菜单备份列表", "global"),
		newPermissionActionSeed("menu_backup", "delete", "删除菜单备份", "允许删除菜单备份", "global"),
		newPermissionActionSeed("menu_backup", "restore", "恢复菜单备份", "允许恢复菜单备份", "global"),
		newPermissionActionSeed("system", "view_page_catalog", "查看页面文件映射", "允许查看页面文件映射", "global"),
		newPermissionActionSeed("tenant", "list", "查看团队列表", "允许查看团队列表", "global"),
		newPermissionActionSeed("tenant", "get", "查看团队详情", "允许查看团队详情", "global"),
		newPermissionActionSeed("tenant", "create", "创建团队", "允许创建团队", "global"),
		newPermissionActionSeed("tenant", "update", "更新团队", "允许更新团队", "global"),
		newPermissionActionSeed("tenant", "delete", "删除团队", "允许删除团队", "global"),
		newPermissionActionSeed("tenant", "configure_action_boundary", "配置团队功能权限边界", "允许配置团队功能权限边界", "global"),
		newPermissionActionSeed("tenant_member_admin", "list", "查看团队成员列表", "允许在系统管理中查看团队成员列表", "global"),
		newPermissionActionSeed("tenant_member_admin", "create", "添加团队成员", "允许在系统管理中添加团队成员", "global"),
		newPermissionActionSeed("tenant_member_admin", "delete", "移除团队成员", "允许在系统管理中移除团队成员", "global"),
		newPermissionActionSeed("tenant_member_admin", "update_role", "更新团队成员身份", "允许在系统管理中更新团队成员身份", "global"),
		newPermissionActionSeed("team_member", "create", "添加当前团队成员", "允许在当前团队中添加成员", "team"),
		newPermissionActionSeed("team_member", "delete", "移除当前团队成员", "允许在当前团队中移除成员", "team"),
		newPermissionActionSeed("team_member", "update_role", "更新当前团队成员身份", "允许在当前团队中更新成员身份", "team"),
		newPermissionActionSeed("team_member", "assign_role", "配置当前团队成员角色", "允许在当前团队中配置成员角色", "team"),
		newPermissionActionSeed("team_member", "assign_action", "配置当前团队成员功能权限", "允许在当前团队中配置成员功能权限", "team"),
		newPermissionActionSeed("team", "configure_action_boundary", "查看和配置当前团队功能权限边界", "允许查看和配置当前团队功能权限边界", "team"),
		newPermissionActionSeed("api_endpoint", "list", "查看 API 注册表", "允许查看 API 注册表", "global"),
		newPermissionActionSeed("api_endpoint", "sync", "同步 API 注册表", "允许同步 API 注册表", "global"),
		newPermissionActionSeed("system_permission", "manage_action_registry", "管理功能权限注册表", "允许维护功能权限注册信息", "global"),
		newPermissionActionSeed("system_permission", "assign_role_action", "配置角色功能权限", "允许为角色分配功能权限", "global"),
	}
}

func newPermissionActionSeed(
	resourceCode, actionCode, name, description, scopeCode string,
) permissionActionSeed {
	return permissionActionSeed{
		ResourceCode: resourceCode,
		ActionCode:   actionCode,
		ModuleCode:   resourceCode,
		Category:     resourceCode,
		Source:       "system",
		FeatureKind:  "system",
		Name:         name,
		Description:  description,
		ScopeCode:    scopeCode,
	}
}

func getScopeIDsByCode(codes []string) (map[string]uuid.UUID, error) {
	var scopes []usermodel.Scope
	if err := database.DB.Where("code IN ?", codes).Find(&scopes).Error; err != nil {
		return nil, err
	}
	scopeIDs := make(map[string]uuid.UUID, len(scopes))
	for _, scope := range scopes {
		scopeIDs[scope.Code] = scope.ID
	}
	for _, code := range codes {
		if _, ok := scopeIDs[code]; !ok {
			return nil, fmt.Errorf("scope %s not found", code)
		}
	}
	return scopeIDs, nil
}

func syncAPIRegistry(logger *zap.Logger, cfg *config.Config) error {
	apirouter.SetupRouter(cfg, logger, database.DB)
	return nil
}
