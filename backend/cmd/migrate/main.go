package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	apirouter "github.com/gg-ecommerce/backend/internal/api/router"
	"github.com/gg-ecommerce/backend/internal/config"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	space "github.com/gg-ecommerce/backend/internal/modules/system/space"
	systemservice "github.com/gg-ecommerce/backend/internal/modules/system/system"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
	workspace "github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
)

func main() {
	freshMode := flag.Bool("fresh", false, "drop current schema and rebuild all tables and seed data from latest code")
	flag.Parse()

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

	if *freshMode {
		logger.Warn("Fresh migration mode enabled, existing public schema data will be dropped")
		if err := resetPublicSchema(logger); err != nil {
			logger.Fatal("Failed to reset public schema", zap.Error(err))
		}
		logger.Info("Public schema reset completed")
	}

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}

	logger.Info("Database migration completed successfully!")

	logger.Info("Historical rename/backfill migrations are disabled; using final workspace schema directly")

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

	if err := initWorkspaceBaseline(logger); err != nil {
		logger.Warn("Failed to initialize workspace baseline", zap.Error(err))
	} else {
		logger.Info("Workspace baseline initialized successfully")
	}

	// 初始化默认菜单空间
	if err := initDefaultMenuSpaces(logger); err != nil {
		logger.Warn("Failed to initialize default menu spaces", zap.Error(err))
	} else {
		logger.Info("Default menu spaces initialized successfully")
	}

	// 初始化默认菜单
	if err := initDefaultMenusNoScope(logger); err != nil {
		logger.Warn("Failed to initialize default menus", zap.Error(err))
	} else {
		logger.Info("Default menus initialized successfully")
	}

	if err := initDefaultPages(logger); err != nil {
		logger.Warn("Failed to initialize default pages", zap.Error(err))
	} else {
		logger.Info("Default pages initialized successfully")
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
	if err := cleanupDeprecatedPermissionKeys(logger); err != nil {
		logger.Warn("Failed to cleanup deprecated permission keys", zap.Error(err))
	} else {
		logger.Info("Deprecated permission keys cleaned up successfully")
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

	if err := ensureAccessTraceNavigationSeed(); err != nil {
		logger.Warn("Failed to initialize access trace navigation seed", zap.Error(err))
	} else {
		logger.Info("Access trace navigation seed initialized successfully")
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

	if err := seedDefaultMessageTemplates(logger); err != nil {
		logger.Warn("Failed to initialize default message templates", zap.Error(err))
	} else {
		logger.Info("Default message templates initialized successfully")
	}

	if err := ensureDefaultFastEnterConfig(logger); err != nil {
		logger.Warn("Failed to initialize default fast enter config", zap.Error(err))
	} else {
		logger.Info("Default fast enter config initialized successfully")
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

	if err := refreshDefaultAccessSnapshots(logger); err != nil {
		logger.Warn("Failed to refresh default access snapshots", zap.Error(err))
	} else {
		logger.Info("Default access snapshots refreshed successfully")
	}

	logger.Info("Migration completed successfully")
}

func resetPublicSchema(logger *zap.Logger) error {
	statements := []string{
		`SELECT pg_terminate_backend(pid)
		 FROM pg_stat_activity
		 WHERE datname = CURRENT_DATABASE()
		   AND pid <> pg_backend_pid()`,
		`DROP SCHEMA IF EXISTS public CASCADE`,
		`CREATE SCHEMA public`,
		`GRANT ALL ON SCHEMA public TO postgres`,
		`GRANT ALL ON SCHEMA public TO public`,
	}

	for _, statement := range statements {
		if err := database.DB.Exec(statement).Error; err != nil {
			return err
		}
	}

	logger.Info("Fresh schema reset applied")
	return nil
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
			Name: "20260330_menu_space_foundation",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS menu_spaces (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						space_key varchar(100) NOT NULL UNIQUE,
						name varchar(150) NOT NULL,
						description text NOT NULL DEFAULT '',
						default_home_path varchar(255) NOT NULL DEFAULT '',
						is_default boolean NOT NULL DEFAULT false,
						status varchar(20) NOT NULL DEFAULT 'normal',
						meta jsonb NOT NULL DEFAULT '{}'::jsonb,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW(),
						deleted_at timestamptz NULL
					)`,
					`CREATE TABLE IF NOT EXISTS menu_space_host_bindings (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						space_key varchar(100) NOT NULL,
						host varchar(255) NOT NULL UNIQUE,
						description text NOT NULL DEFAULT '',
						is_default boolean NOT NULL DEFAULT false,
						status varchar(20) NOT NULL DEFAULT 'normal',
						meta jsonb NOT NULL DEFAULT '{}'::jsonb,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW(),
						deleted_at timestamptz NULL
					)`,
					`ALTER TABLE menus ADD COLUMN IF NOT EXISTS space_key varchar(100) NOT NULL DEFAULT 'default'`,
					`ALTER TABLE ui_pages ADD COLUMN IF NOT EXISTS space_key varchar(100) NOT NULL DEFAULT 'default'`,
					`ALTER TABLE menus ALTER COLUMN space_key SET DEFAULT 'default'`,
					`ALTER TABLE ui_pages ALTER COLUMN space_key SET DEFAULT 'default'`,
					`UPDATE menus SET space_key = 'default' WHERE COALESCE(TRIM(space_key), '') = ''`,
					`UPDATE ui_pages SET space_key = 'default' WHERE COALESCE(TRIM(space_key), '') = ''`,
					`INSERT INTO menu_spaces (space_key, name, description, default_home_path, is_default, status, meta, created_at, updated_at)
					 SELECT 'default', '默认菜单空间', '兼容当前单域单菜单运行模式', '/dashboard/console', TRUE, 'normal', '{}'::jsonb, NOW(), NOW()
					 WHERE NOT EXISTS (
					 	SELECT 1
					 	FROM menu_spaces
					 	WHERE space_key = 'default'
					 		AND deleted_at IS NULL
					 )`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				if err := space.EnsureDefaultMenuSpace(database.DB, systemmodels.DefaultAppKey); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260330_menu_space_foundation"))
				return nil
			},
		},
		{
			Name: "20260405_app_scope_access_snapshots",
			Run: func(logger *zap.Logger) error {
				if err := ensureAppScopedSnapshotPrimaryKeysMigration(); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260405_app_scope_access_snapshots"))
				return nil
			},
		},
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
				targetMenus := []string{"Dashboard"}
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
			Name: "20260331_remove_result_exception_menu_trees",
			Run: func(logger *zap.Logger) error {
				for _, menuName := range []string{"Result", "Exception"} {
					if err := deleteMenuTreeByName(menuName); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_remove_result_exception_menu_trees"))
				return nil
			},
		},
		{
			Name: "20260401_remove_unused_user_assign_action_permission_key",
			Run: func(logger *zap.Logger) error {
				if err := database.DB.Where("permission_key = ?", "system.user.assign_action").Delete(&usermodel.PermissionKey{}).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260401_remove_unused_user_assign_action_permission_key"))
				return nil
			},
		},
		{
			Name: "20260401_fast_enter_config_seed",
			Run: func(logger *zap.Logger) error {
				service := systemservice.NewFastEnterService(database.DB)
				config, err := service.GetConfig()
				if err != nil {
					return err
				}
				if _, err := service.SaveConfig(config); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260401_fast_enter_config_seed"))
				return nil
			},
		},
		{
			Name: "20260329_system_menu_third_level_grouping",
			Run: func(logger *zap.Logger) error {
				targetMenus := []string{
					"SystemAccess",
					"SystemNavigation",
					"SystemIntegration",
					"Role",
					"User",
					"ActionPermission",
					"FeaturePackage",
					"Menus",
					"PageManagement",
					"FastEnterManage",
					"ApiEndpoint",
					"MessageManage",
				}
				for _, menuName := range targetMenus {
					if _, err := syncDefaultMenuSeedByName(menuName); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_system_menu_third_level_grouping"))
				return nil
			},
		},
		{
			Name: "20260329_system_menu_group_directory_component_cleanup",
			Run: func(logger *zap.Logger) error {
				targetMenus := []string{"SystemAccess", "SystemNavigation", "SystemIntegration"}
				if err := database.DB.Model(&usermodel.Menu{}).
					Where("name IN ?", targetMenus).
					Update("component", "").Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_system_menu_group_directory_component_cleanup"))
				return nil
			},
		},
		{
			Name: "20260329_workspace_inbox_menu_seed",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultMenuSeedByName("WorkspaceInbox"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_workspace_inbox_menu_seed"))
				return nil
			},
		},
		{
			Name: "20260330_dashboard_menu_access_mode_align",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultMenuSeedByName("Dashboard"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260330_dashboard_menu_access_mode_align"))
				return nil
			},
		},
		{
			Name: "20260329_message_manage_menu_seed",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultMenuSeedByName("MessageManage"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_manage_menu_seed"))
				return nil
			},
		},
		{
			Name: "20260329_collaboration_workspace_message_manage_menu_seed",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultMenuSeedByName("TeamMessageManage"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_collaboration_workspace_message_manage_menu_seed"))
				return nil
			},
		},
		{
			Name: "20260329_message_template_and_record_menu_seed",
			Run: func(logger *zap.Logger) error {
				targetMenus := []string{"MessageTemplateManage", "MessageRecordManage", "TeamMessageTemplateManage", "TeamMessageRecordManage"}
				for _, menuName := range targetMenus {
					if _, err := syncDefaultMenuSeedByName(menuName); err != nil {
						// 历史菜单种子已移除时，兼容跳过，避免阻塞全量迁移
						if strings.Contains(err.Error(), "default menu seed not found") {
							logger.Warn("Skip missing legacy menu seed", zap.String("menu", menuName), zap.Error(err))
							continue
						}
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_template_and_record_menu_seed"))
				return nil
			},
		},
		{
			Name: "20260329_system_pages_registry_seed",
			Run: func(logger *zap.Logger) error {
				targetPages := []string{
					"display.system_pages",
					"workspace.user_center",
					"workspace.inbox",
					"system.message.manage",
					"system.message.recipient_group.manage",
					"system.message.sender.manage",
					"system.message.template.manage",
					"system.message.record.manage",
					"collaboration_workspace.message.manage",
					"collaboration_workspace.message.recipient_group.manage",
					"collaboration_workspace.message.sender.manage",
					"collaboration_workspace.message.template.manage",
					"collaboration_workspace.message.record.manage",
				}
				for _, pageKey := range targetPages {
					if _, err := syncDefaultPageSeedByKey(pageKey); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_system_pages_registry_seed"))
				return nil
			},
		},
		{
			Name: "20260329_user_center_page_seed",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultPageSeedByKey("workspace.user_center"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_user_center_page_seed"))
				return nil
			},
		},
		{
			Name: "20260329_message_manage_page_permission_align",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultPageSeedByKey("system.message.manage"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_manage_page_permission_align"))
				return nil
			},
		},
		{
			Name: "20260329_collaboration_workspace_message_manage_page_seed",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultPageSeedByKey("collaboration_workspace.message.manage"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_collaboration_workspace_message_manage_page_seed"))
				return nil
			},
		},
		{
			Name: "20260329_message_template_and_record_page_seed",
			Run: func(logger *zap.Logger) error {
				targetPages := []string{
					"system.message.recipient_group.manage",
					"system.message.sender.manage",
					"system.message.template.manage",
					"system.message.record.manage",
					"collaboration_workspace.message.recipient_group.manage",
					"collaboration_workspace.message.sender.manage",
					"collaboration_workspace.message.template.manage",
					"collaboration_workspace.message.record.manage",
				}
				for _, pageKey := range targetPages {
					if _, err := syncDefaultPageSeedByKey(pageKey); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_template_and_record_page_seed"))
				return nil
			},
		},
		{
			Name: "20260329_message_sender_page_seed",
			Run: func(logger *zap.Logger) error {
				targetPages := []string{
					"system.message.sender.manage",
					"collaboration_workspace.message.sender.manage",
				}
				for _, pageKey := range targetPages {
					if _, err := syncDefaultPageSeedByKey(pageKey); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_sender_page_seed"))
				return nil
			},
		},
		{
			Name: "20260329_message_recipient_group_page_seed",
			Run: func(logger *zap.Logger) error {
				targetPages := []string{
					"system.message.recipient_group.manage",
					"collaboration_workspace.message.recipient_group.manage",
				}
				for _, pageKey := range targetPages {
					if _, err := syncDefaultPageSeedByKey(pageKey); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_recipient_group_page_seed"))
				return nil
			},
		},
		{
			Name: "20260402_message_more_page_seed",
			Run: func(logger *zap.Logger) error {
				targetPages := []string{
					"system.message.more",
					"collaboration_workspace.message.more",
				}
				for _, pageKey := range targetPages {
					if _, err := syncDefaultPageSeedByKey(pageKey); err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260402_message_more_page_seed"))
				return nil
			},
		},
		{
			Name: "20260329_message_send_navigation_realign",
			Run: func(logger *zap.Logger) error {
				deprecatedMenus := []string{
					"MessageTemplateManage",
					"MessageRecordManage",
					"TeamMessageTemplateManage",
					"TeamMessageRecordManage",
				}
				for _, menuName := range deprecatedMenus {
					var menu usermodel.Menu
					err := database.DB.Where("name = ?", menuName).First(&menu).Error
					if errors.Is(err, gorm.ErrRecordNotFound) {
						continue
					}
					if err != nil {
						return err
					}
					if err := deleteMenuTree(menu.ID); err != nil {
						return err
					}
				}

				targetMenus := []string{"MessageManage", "TeamMessageManage"}
				for _, menuName := range targetMenus {
					if _, err := syncDefaultMenuSeedByName(menuName); err != nil {
						return err
					}
				}

				targetPages := []string{
					"system.message.manage",
					"system.message.template.manage",
					"system.message.record.manage",
					"collaboration_workspace.message.manage",
					"collaboration_workspace.message.template.manage",
					"collaboration_workspace.message.record.manage",
				}
				for _, pageKey := range targetPages {
					if _, err := syncDefaultPageSeedByKey(pageKey); err != nil {
						return err
					}
				}

				logger.Info("Named migration applied", zap.String("name", "20260329_message_send_navigation_realign"))
				return nil
			},
		},
		{
			Name: "20260329_message_recipient_groups_table_init",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`ALTER TABLE messages ADD COLUMN IF NOT EXISTS target_group_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
					`CREATE TABLE IF NOT EXISTS message_recipient_groups (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						scope_type varchar(20) NOT NULL DEFAULT 'platform',
						scope_id uuid NULL,
						name varchar(120) NOT NULL,
						description text NOT NULL DEFAULT '',
						match_mode varchar(20) NOT NULL DEFAULT 'manual',
						status varchar(20) NOT NULL DEFAULT 'normal',
						meta jsonb NOT NULL DEFAULT '{}'::jsonb,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW(),
						deleted_at timestamptz NULL
					)`,
					`CREATE TABLE IF NOT EXISTS message_recipient_group_targets (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						group_id uuid NOT NULL,
						target_type varchar(30) NOT NULL,
						user_id uuid NULL,
						collaboration_workspace_id uuid NULL,
						role_code varchar(80) NOT NULL DEFAULT '',
						package_key varchar(120) NOT NULL DEFAULT '',
						sort_order integer NOT NULL DEFAULT 0,
						meta jsonb NOT NULL DEFAULT '{}'::jsonb,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW(),
						deleted_at timestamptz NULL
					)`,
					`CREATE INDEX IF NOT EXISTS idx_messages_target_group_ids ON messages USING gin (target_group_ids)`,
					`CREATE INDEX IF NOT EXISTS idx_message_recipient_groups_scope_type ON message_recipient_groups (scope_type)`,
					`CREATE INDEX IF NOT EXISTS idx_message_recipient_groups_scope_id ON message_recipient_groups (scope_id)`,
					`CREATE INDEX IF NOT EXISTS idx_message_recipient_groups_deleted_at ON message_recipient_groups (deleted_at)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_message_recipient_groups_scope_name_unique ON message_recipient_groups (scope_type, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid), name) WHERE deleted_at IS NULL`,
					`CREATE INDEX IF NOT EXISTS idx_message_recipient_group_targets_group_id ON message_recipient_group_targets (group_id)`,
					`CREATE INDEX IF NOT EXISTS idx_message_recipient_group_targets_user_id ON message_recipient_group_targets (user_id)`,
					`CREATE INDEX IF NOT EXISTS idx_message_recipient_group_targets_collaboration_workspace_id ON message_recipient_group_targets (collaboration_workspace_id)`,
					`CREATE INDEX IF NOT EXISTS idx_message_recipient_group_targets_deleted_at ON message_recipient_group_targets (deleted_at)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_recipient_groups_table_init"))
				return nil
			},
		},
		{
			Name: "20260329_message_senders_table_init",
			Run: func(logger *zap.Logger) error {
				statements := []string{
					`CREATE TABLE IF NOT EXISTS message_senders (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						scope_type varchar(20) NOT NULL DEFAULT 'platform',
						scope_id uuid NULL,
						name varchar(120) NOT NULL,
						description text NOT NULL DEFAULT '',
						avatar_url varchar(500) NOT NULL DEFAULT '',
						is_default boolean NOT NULL DEFAULT false,
						status varchar(20) NOT NULL DEFAULT 'normal',
						meta jsonb NOT NULL DEFAULT '{}'::jsonb,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW(),
						deleted_at timestamptz NULL
					)`,
					`ALTER TABLE messages ADD COLUMN IF NOT EXISTS sender_id uuid NULL`,
					`CREATE INDEX IF NOT EXISTS idx_message_senders_scope_type ON message_senders (scope_type)`,
					`CREATE INDEX IF NOT EXISTS idx_message_senders_scope_id ON message_senders (scope_id)`,
					`CREATE INDEX IF NOT EXISTS idx_message_senders_deleted_at ON message_senders (deleted_at)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_message_senders_scope_name_unique ON message_senders (scope_type, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid), name) WHERE deleted_at IS NULL`,
					`CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id)`,
				}
				for _, statement := range statements {
					if err := database.DB.Exec(statement).Error; err != nil {
						return err
					}
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_senders_table_init"))
				return nil
			},
		},
		{
			Name: "20260329_message_sender_deduplicate",
			Run: func(logger *zap.Logger) error {
				deleteSQL := `
					DELETE FROM message_senders target
					USING message_senders existing
					WHERE target.id <> existing.id
					  AND target.deleted_at IS NULL
					  AND existing.deleted_at IS NULL
					  AND target.scope_type = existing.scope_type
					  AND COALESCE(target.scope_id, '00000000-0000-0000-0000-000000000000'::uuid) = COALESCE(existing.scope_id, '00000000-0000-0000-0000-000000000000'::uuid)
					  AND target.name = existing.name
					  AND target.created_at > existing.created_at
				`
				if err := database.DB.Exec(deleteSQL).Error; err != nil {
					return err
				}
				if err := database.DB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_message_senders_scope_name_unique ON message_senders (scope_type, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid), name) WHERE deleted_at IS NULL`).Error; err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_sender_deduplicate"))
				return nil
			},
		},
		{
			Name: "20260329_message_templates_seed",
			Run: func(logger *zap.Logger) error {
				if err := seedDefaultMessageTemplates(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260329_message_templates_seed"))
				return nil
			},
		},
		{
			Name: "20260331_collaboration_workspace_manage_page_seed",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultPageSeedByKey("collaboration_workspace.index"); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_collaboration_workspace_manage_page_seed"))
				return nil
			},
		},
		{
			Name: "20260331_collaboration_workspace_manage_menu_binding_fix",
			Run: func(logger *zap.Logger) error {
				if _, err := syncDefaultMenuSeedByName("CollaborationWorkspaceManage"); err != nil {
					return err
				}
				if _, err := syncDefaultPageSeedByKey("collaboration_workspace.index"); err != nil {
					return err
				}
				if err := initDefaultFeaturePackages(logger); err != nil {
					return err
				}
				if err := refreshDefaultAccessSnapshots(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_team_manage_menu_binding_fix"))
				return nil
			},
		},
		{
			Name: "20260331_team_root_duplicate_cleanup",
			Run: func(logger *zap.Logger) error {
				if err := cleanupDuplicateTeamRoots(logger); err != nil {
					return err
				}
				if err := refreshDefaultAccessSnapshots(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_team_root_duplicate_cleanup"))
				return nil
			},
		},
		{
			Name: "20260331_navigation_model_foundation",
			Run: func(logger *zap.Logger) error {
				if err := ensureNavigationModelFoundation(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_navigation_model_foundation"))
				return nil
			},
		},
		{
			Name: "20260331_menu_entry_page_cleanup",
			Run: func(logger *zap.Logger) error {
				if err := cleanupMenuBackedEntryPages(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_menu_entry_page_cleanup"))
				return nil
			},
		},
		{
			Name: "20260331_space_cloned_page_cleanup",
			Run: func(logger *zap.Logger) error {
				if err := cleanupSpaceClonedPages(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_space_cloned_page_cleanup"))
				return nil
			},
		},
		{
			Name: "20260331_global_page_binding_cleanup",
			Run: func(logger *zap.Logger) error {
				if err := cleanupGlobalPageBindings(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_global_page_binding_cleanup"))
				return nil
			},
		},
		{
			Name: "20260331_access_trace_navigation_seed",
			Run: func(logger *zap.Logger) error {
				if err := ensureAccessTraceNavigationSeed(); err != nil {
					return err
				}
				if err := refreshDefaultAccessSnapshots(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_access_trace_navigation_seed"))
				return nil
			},
		},
		{
			Name: "20260401_remove_permission_simulator_navigation",
			Run: func(logger *zap.Logger) error {
				if err := removePermissionSimulatorNavigation(logger); err != nil {
					return err
				}
				if err := refreshDefaultAccessSnapshots(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260401_remove_permission_simulator_navigation"))
				return nil
			},
		},
		{
			Name: "20260331_permission_api_registry_full_repair",
			Run: func(logger *zap.Logger) error {
				if err := initDefaultPermissionGroups(logger); err != nil {
					return err
				}
				if err := initDefaultPermissionKeysNoScope(logger); err != nil {
					return err
				}
				if err := initDefaultAPIEndpointCategories(logger); err != nil {
					return err
				}
				targetPages := []string{
					"workspace.inbox",
					"system.message.manage",
					"system.message.sender.manage",
					"system.message.recipient_group.manage",
					"system.message.template.manage",
					"system.message.record.manage",
					"collaboration_workspace.message.manage",
					"collaboration_workspace.message.sender.manage",
					"collaboration_workspace.message.recipient_group.manage",
					"collaboration_workspace.message.template.manage",
					"collaboration_workspace.message.record.manage",
				}
				for _, pageKey := range targetPages {
					if _, err := syncDefaultPageSeedByKey(pageKey); err != nil {
						return err
					}
				}
				if err := normalizeLegacyPermissionAndAPIData(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_permission_api_registry_full_repair"))
				return nil
			},
		},
		{
			Name: "20260331_restore_default_space_only",
			Run: func(logger *zap.Logger) error {
				if err := cleanupLegacyOpsSpace(logger); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260331_restore_default_space_only"))
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
						"path":       "/collaboration",
						"component":  "/index/index",
						"title":      "协作空间",
						"icon":       "ri:team-line",
						"sort_order": 5,
						"meta":       teamAccessMeta,
					}).Error; err != nil {
					return err
				}

				var teamMembers usermodel.Menu
				if err := database.DB.Where("name = ?", "CollaborationWorkspaceMembers").First(&teamMembers).Error; err == nil {
					if err := database.DB.Model(&usermodel.Menu{}).
						Where("id = ?", teamMembers.ID).
						Updates(map[string]interface{}{
							"parent_id":  teamRoot.ID,
							"path":       "members",
							"component":  "/collaboration/team-members",
							"title":      "协作空间成员",
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
							"title":      "协作空间角色与权限",
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
					`ALTER TABLE collaboration_workspace_access_snapshots DROP COLUMN IF EXISTS manual_action_ids`,
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
					`UPDATE permission_keys SET context_type = 'platform' WHERE permission_key LIKE 'collaboration_workspace.%'`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE permission_key LIKE 'platform.%'`,
					`UPDATE permission_keys SET context_type = 'collaboration' WHERE permission_key LIKE 'team.%'`,
					`UPDATE permission_keys SET permission_key = 'system.permission.manage', context_type = 'platform' WHERE permission_key IN ('system_permission.manage_action_registry', 'permission_action.manage')`,
					`UPDATE permission_keys SET permission_key = 'system.role.assign_action', context_type = 'platform' WHERE permission_key = 'system_permission.assign_role_action'`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE module_code IN ('role', 'user', 'menu', 'menu_backup', 'permission_action', 'permission_key', 'api_endpoint', 'feature_package')`,
					`UPDATE permission_keys SET context_type = 'platform' WHERE permission_key IN ('feature_package.assign_action', 'feature_package.assign_menu', 'feature_package.assign_team')`,
					`UPDATE permission_keys SET context_type = 'collaboration' WHERE permission_key IN ('team.configure_action_boundary', 'team.configure_menu_boundary')`,
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
					`INSERT INTO user_roles (user_id, role_id, collaboration_workspace_id)
					SELECT tm.user_id, r.id, tm.collaboration_workspace_id
					FROM collaboration_workspace_members tm
					JOIN roles r ON r.code = tm.role_code
					WHERE tm.role_code IN ('team_admin', 'team_member')
					  AND NOT EXISTS (
					    SELECT 1
					    FROM user_roles ur
					    WHERE ur.user_id = tm.user_id
					      AND ur.role_id = r.id
					      AND ur.collaboration_workspace_id = tm.collaboration_workspace_id
					  )`,
					`UPDATE user_roles ur
					SET collaboration_workspace_id = tm.collaboration_workspace_id
					FROM (
					  SELECT user_id, MAX(collaboration_workspace_id::text)::uuid AS collaboration_workspace_id
					  FROM collaboration_workspace_members
					  GROUP BY user_id
					  HAVING COUNT(DISTINCT collaboration_workspace_id) = 1
					) tm,
					roles r
					WHERE ur.user_id = tm.user_id
					  AND r.id = ur.role_id
					  AND ur.collaboration_workspace_id IS NULL
					  AND r.code IN ('team_admin', 'team_member')
					  AND NOT EXISTS (
					    SELECT 1
					    FROM user_roles existing
					    WHERE existing.user_id = ur.user_id
					      AND existing.role_id = ur.role_id
					      AND existing.collaboration_workspace_id = tm.collaboration_workspace_id
					  )`,
					`DELETE FROM user_roles ur
					USING roles r
					WHERE ur.role_id = r.id
					  AND ur.collaboration_workspace_id IS NULL
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
					`UPDATE permission_keys SET context_type = 'platform' WHERE COALESCE(context_type, '') = '' AND (permission_key LIKE 'system.%' OR permission_key LIKE 'collaboration_workspace.%' OR permission_key LIKE 'platform.%')`,
					`UPDATE permission_keys SET context_type = 'collaboration' WHERE COALESCE(context_type, '') = ''`,
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
						context_type varchar(20) NOT NULL DEFAULT 'collaboration',
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
					`CREATE TABLE IF NOT EXISTS collaboration_workspace_feature_packages (
						id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
						collaboration_workspace_id uuid NOT NULL,
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
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_collaboration_workspace_feature_packages_unique ON collaboration_workspace_feature_packages (collaboration_workspace_id, package_id)`,
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
					`CREATE TABLE IF NOT EXISTS collaboration_workspace_blocked_menus (
						collaboration_workspace_id uuid NOT NULL,
						menu_id uuid NOT NULL,
						created_at timestamptz NOT NULL DEFAULT NOW(),
						updated_at timestamptz NOT NULL DEFAULT NOW()
					)`,
					`CREATE TABLE IF NOT EXISTS collaboration_workspace_blocked_actions (
						collaboration_workspace_id uuid NOT NULL,
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
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_collaboration_workspace_blocked_menus_unique ON collaboration_workspace_blocked_menus (collaboration_workspace_id, menu_id)`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_collaboration_workspace_blocked_actions_unique ON collaboration_workspace_blocked_actions (collaboration_workspace_id, action_id)`,
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
					`CREATE TABLE IF NOT EXISTS collaboration_workspace_access_snapshots (
						collaboration_workspace_id uuid PRIMARY KEY,
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
					`CREATE TABLE IF NOT EXISTS collaboration_workspace_role_access_snapshots (
						collaboration_workspace_id uuid NOT NULL,
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
						PRIMARY KEY (collaboration_workspace_id, role_id)
					)`,
					`ALTER TABLE collaboration_workspace_role_access_snapshots ADD COLUMN IF NOT EXISTS available_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
					`ALTER TABLE collaboration_workspace_role_access_snapshots ADD COLUMN IF NOT EXISTS disabled_action_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
					`ALTER TABLE collaboration_workspace_role_access_snapshots ADD COLUMN IF NOT EXISTS available_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
					`ALTER TABLE collaboration_workspace_role_access_snapshots ADD COLUMN IF NOT EXISTS hidden_menu_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
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
					`ALTER TABLE collaboration_workspace_role_access_snapshots DROP COLUMN IF EXISTS has_menu_boundary`,
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
					`ALTER TABLE roles ADD COLUMN IF NOT EXISTS collaboration_workspace_id uuid`,
					`DROP INDEX IF EXISTS roles_code_key`,
					`DROP INDEX IF EXISTS idx_roles_code`,
					`DROP INDEX IF EXISTS idx_roles_code_unique`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_global_code_unique ON roles (code) WHERE collaboration_workspace_id IS NULL AND deleted_at IS NULL`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_tenant_code_unique ON roles (collaboration_workspace_id, code) WHERE collaboration_workspace_id IS NOT NULL AND deleted_at IS NULL`,
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
					  SET collaboration_workspace_id = COALESCE(collaboration_workspace_id, scope_target_id)
					  WHERE collaboration_workspace_id IS NULL AND scope_target_id IS NOT NULL`)
				}
				if hasScopeTargetID, err := hasColumn("user_action_permissions", "scope_target_id"); err != nil {
					return err
				} else if hasScopeTargetID {
					statements = append(statements, `UPDATE user_action_permissions
					  SET collaboration_workspace_id = COALESCE(collaboration_workspace_id, scope_target_id)
					  WHERE collaboration_workspace_id IS NULL AND scope_target_id IS NOT NULL`)
				}
				statements = append(statements,
					`ALTER TABLE user_action_permissions DROP CONSTRAINT IF EXISTS user_action_permissions_pkey`,
					`ALTER TABLE user_action_permissions ALTER COLUMN collaboration_workspace_id DROP NOT NULL`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_global_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_tenant_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_global_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_tenant_unique`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_id`,
					`DROP INDEX IF EXISTS idx_user_action_permissions_scope_target_id`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_action_permissions_global_unique ON user_action_permissions (user_id, action_id) WHERE collaboration_workspace_id IS NULL`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_action_permissions_tenant_unique ON user_action_permissions (user_id, action_id, collaboration_workspace_id) WHERE collaboration_workspace_id IS NOT NULL`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_global_unique`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_target_unique`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_id`,
					`DROP INDEX IF EXISTS idx_user_roles_scope_target_id`,
					`DROP INDEX IF EXISTS idx_user_roles_deleted_at`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_global_unique ON user_roles (user_id, role_id) WHERE collaboration_workspace_id IS NULL`,
					`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_tenant_unique ON user_roles (user_id, role_id, collaboration_workspace_id) WHERE collaboration_workspace_id IS NOT NULL`,
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
			Name: "20260401_user_email_partial_unique_index",
			Run: func(logger *zap.Logger) error {
				if err := ensureUserEmailPartialUniqueIndex(); err != nil {
					return err
				}
				logger.Info("Named migration applied", zap.String("name", "20260401_user_email_partial_unique_index"))
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

func ensureAccessTraceNavigationSeed() error {
	meta := usermodel.MetaJSON{
		"roles":      []interface{}{"R_SUPER", "R_ADMIN"},
		"keepAlive":  true,
		"accessMode": "permission",
	}
	if err := normalizeAccessTraceNavigationSeed(systemmodels.DefaultMenuSpaceKey, "/system/access-trace"); err != nil {
		return err
	}
	definition, err := syncMenuSeed(permissionseed.MenuSeed{
		SpaceKey:   systemmodels.DefaultMenuSpaceKey,
		Name:       "AccessTrace",
		ParentName: "SystemNavigation",
		Path:       "/system/access-trace",
		Component:  "/system/access-trace",
		Title:      "访问链路测试",
		SortOrder:  5,
		Meta:       meta,
	})
	if err != nil {
		return err
	}

	var page systemmodels.UIPage
	if err := database.DB.Where("page_key = ?", "system.access_trace.manage").First(&page).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		page = systemmodels.UIPage{
			PageKey:           "system.access_trace.manage",
			Name:              "访问链路测试",
			RouteName:         "SystemAccessTrace",
			RoutePath:         "/system/access-trace",
			Component:         "/system/access-trace",
			SpaceKey:          "",
			PageType:          "inner",
			VisibilityScope:   "inherit",
			Source:            "manual",
			ModuleKey:         "page",
			SortOrder:         28,
			ParentMenuID:      &definition.ID,
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/access-trace",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "system.page.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		}
		if err := database.DB.Create(&page).Error; err != nil {
			return err
		}
	} else {
		page.Name = "访问链路测试"
		page.RouteName = "SystemAccessTrace"
		page.RoutePath = "/system/access-trace"
		page.Component = "/system/access-trace"
		page.SpaceKey = ""
		page.PageType = "inner"
		page.VisibilityScope = "inherit"
		page.Source = "manual"
		page.ModuleKey = "page"
		page.SortOrder = 28
		page.ParentMenuID = &definition.ID
		page.DisplayGroupKey = "display.system_pages"
		page.ActiveMenuPath = "/system/access-trace"
		page.BreadcrumbMode = "inherit_menu"
		page.AccessMode = "permission"
		page.PermissionKey = "system.page.manage"
		page.InheritPermission = false
		page.KeepAlive = true
		page.Status = "normal"
		if page.Meta == nil {
			page.Meta = usermodel.MetaJSON{}
		}
		if err := database.DB.Save(&page).Error; err != nil {
			return err
		}
	}
	if err := ensureAccessTracePackageMenuBinding(definition.ID); err != nil {
		return err
	}
	return nil
}

func removePermissionSimulatorNavigation(logger *zap.Logger) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var menus []usermodel.Menu
		if err := tx.Where("space_key = ? AND (name = ? OR title = ? OR path = ?)", systemmodels.DefaultMenuSpaceKey, "PermissionSimulator", "权限模拟", "/system/permission-simulator").
			Find(&menus).Error; err != nil {
			return err
		}
		menuIDs := make([]uuid.UUID, 0, len(menus))
		for _, menu := range menus {
			menuIDs = append(menuIDs, menu.ID)
		}
		if len(menuIDs) > 0 {
			tables := []string{
				"feature_package_menus",
				"role_hidden_menus",
				"collaboration_workspace_blocked_menus",
				"user_hidden_menus",
			}
			for _, table := range tables {
				if err := tx.Exec("DELETE FROM "+table+" WHERE menu_id IN ?", menuIDs).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("id IN ?", menuIDs).Delete(&usermodel.Menu{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("page_key = ? OR route_path = ? OR name IN ?", "system.permission.simulator", "/system/permission-simulator", []string{"PermissionSimulator", "权限模拟"}).
			Delete(&systemmodels.UIPage{}).Error; err != nil {
			return err
		}
		if logger != nil {
			logger.Info("Permission simulator navigation removed", zap.Int("menu_count", len(menuIDs)))
		}
		return nil
	})
}

func ensureUserEmailPartialUniqueIndex() error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			`DROP INDEX IF EXISTS idx_users_email`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL AND btrim(email) <> ''`,
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func normalizeAccessTraceNavigationSeed(spaceKey, path string) error {
	const accessTraceName = "\u8bbf\u95ee\u94fe\u8def\u6d4b\u8bd5"
	var primaryMenuID *uuid.UUID
	var definitions []systemmodels.MenuDefinition
	if err := database.DB.Where("app_key = ? AND path = ?", systemmodels.DefaultAppKey, path).Order("created_at asc").Find(&definitions).Error; err != nil {
		return err
	}
	for _, definition := range definitions {
		definition.Name = "AccessTrace"
		definition.DefaultTitle = accessTraceName
		definition.Component = path
		definition.Meta = usermodel.MetaJSON{
			"roles":      []interface{}{"R_SUPER", "R_ADMIN"},
			"keepAlive":  true,
			"accessMode": "permission",
		}
		if primaryMenuID == nil {
			primaryMenuID = &definition.ID
		}
		if err := database.DB.Model(&systemmodels.MenuDefinition{}).Where("id = ?", definition.ID).Updates(map[string]any{
			"name":          definition.Name,
			"default_title": definition.DefaultTitle,
			"component":     definition.Component,
			"meta":          definition.Meta,
		}).Error; err != nil {
			return err
		}
	}
	if err := database.DB.Model(&systemmodels.SpaceMenuPlacement{}).
		Where("app_key = ? AND space_key = ? AND menu_key = ?", systemmodels.DefaultAppKey, spaceKey, "AccessTrace").
		Updates(map[string]any{"sort_order": 5, "hidden": false}).Error; err != nil {
		return err
	}

	var pages []systemmodels.UIPage
	if err := database.DB.Where("route_path = ?", path).Order("created_at asc").Find(&pages).Error; err != nil {
		return err
	}
	for _, page := range pages {
		page.Name = accessTraceName
		page.RouteName = "SystemAccessTrace"
		page.RoutePath = path
		page.Component = path
		page.SpaceKey = ""
		page.PageType = "inner"
		page.VisibilityScope = "inherit"
		page.Source = "manual"
		page.ModuleKey = "page"
		page.DisplayGroupKey = "display.system_pages"
		page.ActiveMenuPath = path
		page.BreadcrumbMode = "inherit_menu"
		page.AccessMode = "permission"
		page.PermissionKey = "system.page.manage"
		page.InheritPermission = false
		page.KeepAlive = true
		page.Status = "normal"
		if page.Meta == nil {
			page.Meta = usermodel.MetaJSON{}
		}
		if primaryMenuID != nil {
			page.ParentMenuID = primaryMenuID
		}
		if err := database.DB.Model(&systemmodels.UIPage{}).Where("id = ?", page.ID).Updates(map[string]interface{}{
			"name":               accessTraceName,
			"route_name":         page.RouteName,
			"route_path":         page.RoutePath,
			"component":          page.Component,
			"space_key":          "",
			"visibility_scope":   page.VisibilityScope,
			"page_type":          page.PageType,
			"source":             page.Source,
			"module_key":         page.ModuleKey,
			"display_group_key":  page.DisplayGroupKey,
			"active_menu_path":   page.ActiveMenuPath,
			"breadcrumb_mode":    page.BreadcrumbMode,
			"access_mode":        page.AccessMode,
			"permission_key":     page.PermissionKey,
			"inherit_permission": page.InheritPermission,
			"keep_alive":         page.KeepAlive,
			"status":             page.Status,
			"parent_menu_id":     page.ParentMenuID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureAccessTracePackageMenuBinding(menuID uuid.UUID) error {
	packageKeys := []string{"platform.system_admin", "platform.menu_admin"}
	var packages []systemmodels.FeaturePackage
	if err := database.DB.Where("app_key = ? AND package_key IN ?", systemmodels.DefaultAppKey, packageKeys).Find(&packages).Error; err != nil {
		return err
	}
	for _, item := range packages {
		binding := systemmodels.FeaturePackageMenu{
			PackageID: item.ID,
			MenuID:    menuID,
		}
		if err := database.DB.Where("package_id = ? AND menu_id = ?", item.ID, menuID).FirstOrCreate(&binding).Error; err != nil {
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

func prepareCollaborationWorkspaceRenames(logger *zap.Logger) error {
	tableRenames := []struct {
		oldName string
		newName string
	}{
		{oldName: "tenants", newName: "collaboration_workspaces"},
		{oldName: "tenant_members", newName: "collaboration_workspace_members"},
		{oldName: "team_feature_packages", newName: "collaboration_workspace_feature_packages"},
		{oldName: "team_blocked_menus", newName: "collaboration_workspace_blocked_menus"},
		{oldName: "team_blocked_actions", newName: "collaboration_workspace_blocked_actions"},
		{oldName: "team_access_snapshots", newName: "collaboration_workspace_access_snapshots"},
		{oldName: "team_role_access_snapshots", newName: "collaboration_workspace_role_access_snapshots"},
	}
	for _, item := range tableRenames {
		if err := renameTableIfNeeded(item.oldName, item.newName, logger); err != nil {
			return err
		}
	}

	columnRenames := []struct {
		tableName string
		oldName   string
		newName   string
	}{
		{tableName: "workspaces", oldName: "source_tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "workspace_members", oldName: "source_tenant_member_id", newName: "collaboration_workspace_member_id"},
		{tableName: "roles", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "user_roles", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "user_action_permissions", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "user_hidden_menus", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "api_keys", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "media_assets", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "message_recipient_group_targets", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "messages", oldName: "target_tenant_id", newName: "target_collaboration_workspace_id"},
		{tableName: "message_templates", oldName: "owner_tenant_id", newName: "owner_collaboration_workspace_id"},
		{tableName: "collaboration_workspace_members", oldName: "tenant_id", newName: "collaboration_workspace_id"},
		{tableName: "collaboration_workspace_feature_packages", oldName: "team_id", newName: "collaboration_workspace_id"},
		{tableName: "collaboration_workspace_blocked_menus", oldName: "team_id", newName: "collaboration_workspace_id"},
		{tableName: "collaboration_workspace_blocked_actions", oldName: "team_id", newName: "collaboration_workspace_id"},
		{tableName: "collaboration_workspace_access_snapshots", oldName: "team_id", newName: "collaboration_workspace_id"},
		{tableName: "collaboration_workspace_role_access_snapshots", oldName: "team_id", newName: "collaboration_workspace_id"},
	}
	for _, item := range columnRenames {
		if err := renameColumnIfNeeded(item.tableName, item.oldName, item.newName, logger); err != nil {
			return err
		}
	}

	indexRenames := []struct {
		oldName string
		newName string
	}{
		{oldName: "idx_collaboration_workspace_members_tenant_user_unique", newName: "idx_collaboration_workspace_members_user_unique"},
		{oldName: "idx_user_roles_tenant_unique", newName: "idx_user_roles_collaboration_workspace_unique"},
		{oldName: "idx_workspaces_team_tenant_unique", newName: "idx_workspaces_collaboration_workspace_unique"},
		{oldName: "idx_user_action_permissions_tenant_unique", newName: "idx_user_action_permissions_collaboration_workspace_unique"},
	}
	for _, item := range indexRenames {
		if err := renameIndexIfNeeded(item.oldName, item.newName, logger); err != nil {
			return err
		}
	}

	valueUpdateStatements := []string{
		`UPDATE workspaces SET workspace_type = 'collaboration' WHERE workspace_type = 'team'`,
		`UPDATE permission_keys SET context_type = 'collaboration' WHERE context_type = 'team'`,
		`UPDATE permission_keys SET allowed_workspace_types = regexp_replace(COALESCE(allowed_workspace_types, ''), '(^|,)team(,|$)', '\1collaboration\2', 'g') WHERE COALESCE(allowed_workspace_types, '') LIKE '%team%'`,
		`UPDATE feature_packages SET context_type = 'collaboration' WHERE context_type = 'team'`,
		`UPDATE messages SET scope_type = 'collaboration' WHERE scope_type = 'team'`,
		`UPDATE message_templates SET owner_scope = 'collaboration' WHERE owner_scope IN ('team', 'tenant')`,
		`UPDATE message_senders SET scope_type = 'collaboration' WHERE scope_type = 'team'`,
		`UPDATE message_recipient_groups SET scope_type = 'collaboration' WHERE scope_type = 'team'`,
		`UPDATE message_templates SET audience_type = 'collaboration_workspace_admins' WHERE audience_type = 'tenant_admins'`,
		`UPDATE message_templates SET audience_type = 'collaboration_workspace_users' WHERE audience_type = 'tenant_users'`,
		`UPDATE messages SET audience_type = 'collaboration_workspace_admins' WHERE audience_type = 'tenant_admins'`,
		`UPDATE messages SET audience_type = 'collaboration_workspace_users' WHERE audience_type = 'tenant_users'`,
		`UPDATE message_recipient_group_targets SET target_type = 'collaboration_workspace_admins' WHERE target_type = 'tenant_admins'`,
		`UPDATE message_recipient_group_targets SET target_type = 'collaboration_workspace_users' WHERE target_type = 'tenant_users'`,
		`UPDATE message_templates SET template_key = 'platform.notice.collaboration_workspace_admins' WHERE template_key = 'platform.notice.tenant_admins'`,
		`UPDATE message_templates SET template_key = 'collaboration_workspace.notice.collaboration_workspace_users' WHERE template_key = 'collaboration_workspace.notice.team_members'`,
	}
	for _, statement := range valueUpdateStatements {
		if err := database.DB.Exec(statement).Error; err != nil {
			return err
		}
	}

	return nil
}

func renameTableIfNeeded(oldName, newName string, logger *zap.Logger) error {
	oldExists, err := hasTable(oldName)
	if err != nil {
		return err
	}
	newExists, err := hasTable(newName)
	if err != nil {
		return err
	}
	if !oldExists || newExists {
		return nil
	}
	if err := database.DB.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", oldName, newName)).Error; err != nil {
		return err
	}
	logger.Info("Renamed collaboration workspace table", zap.String("from", oldName), zap.String("to", newName))
	return nil
}

func renameColumnIfNeeded(tableName, oldName, newName string, logger *zap.Logger) error {
	tableExists, err := hasTable(tableName)
	if err != nil {
		return err
	}
	if !tableExists {
		return nil
	}
	oldExists, err := hasColumn(tableName, oldName)
	if err != nil {
		return err
	}
	newExists, err := hasColumn(tableName, newName)
	if err != nil {
		return err
	}
	if !oldExists || newExists {
		return nil
	}
	if err := database.DB.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, oldName, newName)).Error; err != nil {
		return err
	}
	logger.Info("Renamed collaboration workspace column",
		zap.String("table", tableName),
		zap.String("from", oldName),
		zap.String("to", newName),
	)
	return nil
}

func renameIndexIfNeeded(oldName, newName string, logger *zap.Logger) error {
	oldExists, err := hasIndex(oldName)
	if err != nil {
		return err
	}
	newExists, err := hasIndex(newName)
	if err != nil {
		return err
	}
	if !oldExists || newExists {
		return nil
	}
	if err := database.DB.Exec(fmt.Sprintf("ALTER INDEX %s RENAME TO %s", oldName, newName)).Error; err != nil {
		return err
	}
	logger.Info("Renamed collaboration workspace index", zap.String("from", oldName), zap.String("to", newName))
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

func hasIndex(indexName string) (bool, error) {
	var count int64
	err := database.DB.Raw(`
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE schemaname = CURRENT_SCHEMA()
		  AND indexname = ?
	`, indexName).Scan(&count).Error
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
				          (existing.collaboration_workspace_id IS NULL AND target.collaboration_workspace_id IS NULL) OR
				          existing.collaboration_workspace_id = target.collaboration_workspace_id
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
			sql: `UPDATE collaboration_workspace_blocked_actions target
				    SET action_id = ?
				  WHERE action_id = ?
				    AND NOT EXISTS (
				      SELECT 1 FROM collaboration_workspace_blocked_actions existing
				      WHERE existing.collaboration_workspace_id = target.collaboration_workspace_id
				        AND existing.action_id = ?
				    )`,
			args: []interface{}{toID, fromID, toID},
		},
		{
			sql:  `DELETE FROM collaboration_workspace_blocked_actions WHERE action_id = ?`,
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
		{From: "collaboration_workspace.member.manage", To: "collaboration_workspace.manage"},
		{From: "collaboration_workspace.boundary.manage", To: "collaboration_workspace.manage"},
		{From: "collaboration_workspace.configure_menu_boundary", To: "collaboration_workspace.manage"},
		{From: "user.assign_menu", To: "system.user.manage"},
		{From: "team.member.assign_role", To: "collaboration_workspace.member.manage"},
		{From: "team.member.assign_action", To: "collaboration_workspace.member.manage"},
		{From: "team.configure_menu_boundary", To: "collaboration_workspace.boundary.manage"},
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

func normalizeLegacyPermissionAndAPIData(logger *zap.Logger) error {
	if err := normalizeLegacyPagePermissionKeys(logger); err != nil {
		return err
	}
	if err := deduplicateAPIEndpointPermissionBindings(logger); err != nil {
		return err
	}
	return nil
}

func normalizeLegacyPagePermissionKeys(logger *zap.Logger) error {
	pageSeeds := append(permissionseed.DefaultPages(), permissionseed.LegacyMenuBackedPages()...)
	canonicalByKey := make(map[string]string, len(pageSeeds))
	for _, spec := range pageSeeds {
		pageKey := strings.TrimSpace(spec.PageKey)
		permissionKey := strings.TrimSpace(spec.PermissionKey)
		if pageKey == "" || permissionKey == "" || pageKey == permissionKey {
			continue
		}
		canonicalByKey[pageKey] = permissionKey
	}
	if len(canonicalByKey) == 0 {
		return nil
	}

	normalizedCount := 0
	deletedCount := 0
	for fromKey, toKey := range canonicalByKey {
		var fromAction usermodel.PermissionKey
		if err := database.DB.Where("permission_key = ? AND deleted_at IS NULL", fromKey).First(&fromAction).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		var toAction usermodel.PermissionKey
		if err := database.DB.Where("permission_key = ? AND deleted_at IS NULL", toKey).First(&toAction).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		if err := database.DB.Model(&systemmodels.UIPage{}).
			Where("permission_key = ? AND deleted_at IS NULL", fromKey).
			Update("permission_key", toKey).Error; err != nil {
			return err
		}
		if err := rebindAPIBindingPermissionKey(fromKey, toKey); err != nil {
			return err
		}
		if err := rebindPermissionKeyReferences(fromAction.ID, toAction.ID); err != nil {
			return err
		}
		normalizedCount++

		remaining, err := countPermissionKeyReferences(fromAction.ID, fromKey)
		if err != nil {
			return err
		}
		if remaining == 0 {
			if err := database.DB.Delete(&usermodel.PermissionKey{}, fromAction.ID).Error; err != nil {
				return err
			}
			deletedCount++
		}
	}

	logger.Info("Legacy page-backed permission keys normalized",
		zap.Int("normalized", normalizedCount),
		zap.Int("deleted", deletedCount),
	)
	return nil
}

func rebindAPIBindingPermissionKey(fromKey, toKey string) error {
	if strings.TrimSpace(fromKey) == "" || strings.TrimSpace(toKey) == "" || strings.TrimSpace(fromKey) == strings.TrimSpace(toKey) {
		return nil
	}
	updateSQL := `
		UPDATE api_endpoint_permission_bindings target
		   SET permission_key = ?
		 WHERE permission_key = ?
		   AND deleted_at IS NULL
		   AND NOT EXISTS (
		     SELECT 1
		       FROM api_endpoint_permission_bindings existing
		      WHERE existing.endpoint_code = target.endpoint_code
		        AND existing.permission_key = ?
		        AND existing.deleted_at IS NULL
		   )`
	if err := database.DB.Exec(updateSQL, toKey, fromKey, toKey).Error; err != nil {
		return err
	}
	return database.DB.
		Where("permission_key = ? AND deleted_at IS NULL", fromKey).
		Delete(&usermodel.APIEndpointPermissionBinding{}).Error
}

func countPermissionKeyReferences(actionID uuid.UUID, permissionKey string) (int64, error) {
	type counter struct {
		Count int64
	}
	var result counter
	query := `
		SELECT COALESCE((
			SELECT COUNT(*) FROM feature_package_keys WHERE action_id = ?
		), 0) +
		COALESCE((
			SELECT COUNT(*) FROM user_action_permissions WHERE action_id = ?
		), 0) +
		COALESCE((
			SELECT COUNT(*) FROM role_disabled_actions WHERE action_id = ?
		), 0) +
		COALESCE((
			SELECT COUNT(*) FROM collaboration_workspace_blocked_actions WHERE action_id = ?
		), 0) +
		COALESCE((
			SELECT COUNT(*) FROM ui_pages WHERE permission_key = ? AND deleted_at IS NULL
		), 0) +
		COALESCE((
			SELECT COUNT(*) FROM api_endpoint_permission_bindings WHERE permission_key = ? AND deleted_at IS NULL
		), 0) AS count`
	if err := database.DB.Raw(query, actionID, actionID, actionID, actionID, permissionKey, permissionKey).Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Count, nil
}

func deduplicateAPIEndpointPermissionBindings(logger *zap.Logger) error {
	deleteSQL := `
		DELETE FROM api_endpoint_permission_bindings target
		 WHERE deleted_at IS NULL
		   AND EXISTS (
		     SELECT 1
		       FROM api_endpoint_permission_bindings existing
		      WHERE existing.endpoint_code = target.endpoint_code
		        AND existing.permission_key = target.permission_key
		        AND existing.deleted_at IS NULL
		        AND (
		          existing.created_at < target.created_at OR
		          (existing.created_at = target.created_at AND existing.id::text < target.id::text)
		        )
		   )`
	result := database.DB.Exec(deleteSQL)
	if result.Error != nil {
		return result.Error
	}

	if err := database.DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_api_endpoint_permission_bindings_endpoint_permission_unique
		    ON api_endpoint_permission_bindings (endpoint_code, permission_key)
		 WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	logger.Info("API endpoint permission bindings deduplicated", zap.Int64("deleted", result.RowsAffected))
	return nil
}

func backfillTenantIdentityUserRoles(logger *zap.Logger) error {
	type tenantMemberRow struct {
		CollaborationWorkspaceID uuid.UUID
		UserID                   uuid.UUID
		RoleCode                 string
		Status                   string
	}

	var members []tenantMemberRow
	if err := database.DB.Model(&usermodel.TenantMember{}).
		Select("collaboration_workspace_id, user_id, role_code, status").
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
	if err := database.DB.Where("collaboration_workspace_id IS NULL AND code IN ?", roleCodes).Find(&roles).Error; err != nil {
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
			if err := tx.Where("user_id = ? AND collaboration_workspace_id = ? AND role_id IN ?", member.UserID, member.CollaborationWorkspaceID, identityRoleIDs).
				Delete(&usermodel.UserRole{}).Error; err != nil {
				return err
			}
			record := usermodel.UserRole{
				UserID:                   member.UserID,
				RoleID:                   roleID,
				CollaborationWorkspaceID: &member.CollaborationWorkspaceID,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			reboundCount++
			if _, exists := seenTeams[member.CollaborationWorkspaceID]; !exists {
				seenTeams[member.CollaborationWorkspaceID] = struct{}{}
				touchedTeams = append(touchedTeams, member.CollaborationWorkspaceID)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	boundaryService := collaborationworkspaceboundary.NewService(database.DB)
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
		{"team_admin", "协作空间管理员", "协作空间管理员，可以管理协作空间成员和协作空间内容", 2},
		{"team_member", "协作空间成员", "协作空间成员，可以查看和编辑协作空间内容", 3},
	}

	for _, roleData := range roles {
		var role usermodel.Role
		result := database.DB.Where("code = ? AND collaboration_workspace_id IS NULL", roleData.Code).First(&role)
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
	if err := database.DB.Where("code = ? AND collaboration_workspace_id IS NULL", "admin").First(&adminRole).Error; err != nil {
		logger.Error("Failed to find admin role", zap.Error(err))
		return err
	}

	var userRole usermodel.UserRole
	result := database.DB.Where("user_id = ? AND role_id = ? AND collaboration_workspace_id IS NULL", userID, adminRole.ID).First(&userRole)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			userRole = usermodel.UserRole{
				UserID:                   userID,
				RoleID:                   adminRole.ID,
				CollaborationWorkspaceID: nil,
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

func initDefaultMenuSpaces(logger *zap.Logger) error {
	defaultAppShouldUseMultiSpace := len(permissionseed.DefaultMenuSpaces()) > 1
	for _, spec := range permissionseed.DefaultMenuSpaces() {
		spaceModel := &systemmodels.MenuSpace{
			AppKey:          systemmodels.DefaultAppKey,
			SpaceKey:        strings.TrimSpace(spec.SpaceKey),
			Name:            strings.TrimSpace(spec.Name),
			Description:     strings.TrimSpace(spec.Description),
			DefaultHomePath: strings.TrimSpace(spec.DefaultHomePath),
			IsDefault:       spec.IsDefault || strings.TrimSpace(spec.SpaceKey) == systemmodels.DefaultMenuSpaceKey,
			Status:          normalizeMenuSpaceStatus(spec.Status),
			Meta:            spec.Meta,
		}
		if spaceModel.SpaceKey == "" {
			spaceModel.SpaceKey = systemmodels.DefaultMenuSpaceKey
		}
		if spaceModel.Name == "" {
			spaceModel.Name = "默认菜单空间"
		}
		if spaceModel.DefaultHomePath == "" && spaceModel.SpaceKey == systemmodels.DefaultMenuSpaceKey {
			spaceModel.DefaultHomePath = "/dashboard/console"
		}
		if spaceModel.Meta == nil {
			spaceModel.Meta = usermodel.MetaJSON{}
		}
		var existing systemmodels.MenuSpace
		err := database.DB.
			Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", spaceModel.AppKey, spaceModel.SpaceKey).
			First(&existing).Error
		switch {
		case err == nil:
			if err := database.DB.Model(&existing).Updates(map[string]any{
				"name":              spaceModel.Name,
				"description":       spaceModel.Description,
				"default_home_path": spaceModel.DefaultHomePath,
				"is_default":        spaceModel.IsDefault,
				"status":            spaceModel.Status,
				"meta":              spaceModel.Meta,
			}).Error; err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := database.DB.Create(spaceModel).Error; err != nil {
				return err
			}
		default:
			return err
		}
	}
	if err := space.EnsureDefaultMenuSpace(database.DB, systemmodels.DefaultAppKey); err != nil {
		return err
	}
	if defaultAppShouldUseMultiSpace {
		if err := database.DB.Model(&systemmodels.App{}).
			Where("app_key = ? AND deleted_at IS NULL", systemmodels.DefaultAppKey).
			Update("space_mode", "multi").Error; err != nil {
			return err
		}
	}
	logger.Info("Default menu spaces ensured")
	return nil
}

func initDefaultPages(logger *zap.Logger) error {
	for _, spec := range permissionseed.DefaultPages() {
		if _, err := syncDefaultPageSeedByKey(spec.PageKey); err != nil {
			return err
		}
	}
	logger.Info("Default pages ensured")
	return nil
}

func ensureDefaultMenuSeedByName(name string) (*systemmodels.MenuDefinition, error) {
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

func syncDefaultMenuSeedByName(name string) (*systemmodels.MenuDefinition, error) {
	targetName := strings.TrimSpace(name)
	if targetName == "" {
		return nil, fmt.Errorf("menu seed name is required")
	}

	for _, spec := range permissionseed.DefaultMenus() {
		if spec.Name != targetName {
			continue
		}
		if spec.ParentName != "" {
			if _, err := syncDefaultMenuSeedByName(spec.ParentName); err != nil {
				return nil, err
			}
		}
		return syncMenuSeed(spec)
	}

	return nil, fmt.Errorf("default menu seed not found: %s", targetName)
}

func ensureMenuSeed(spec permissionseed.MenuSeed) (*systemmodels.MenuDefinition, error) {
	return syncMenuSeed(spec)
}

func syncMenuSeed(spec permissionseed.MenuSeed) (*systemmodels.MenuDefinition, error) {
	appKey := systemmodels.DefaultAppKey
	menuKey := strings.TrimSpace(spec.Name)
	if menuKey == "" {
		return nil, fmt.Errorf("menu seed name is required")
	}

	meta := spec.Meta
	if meta == nil {
		meta = usermodel.MetaJSON{}
	}

	definition := &systemmodels.MenuDefinition{}
	err := database.DB.Where("app_key = ? AND menu_key = ? AND deleted_at IS NULL", appKey, menuKey).First(definition).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		definition = &systemmodels.MenuDefinition{
			AppKey:        appKey,
			MenuKey:       menuKey,
			Kind:          normalizeMenuSeedKind(spec),
			Path:          strings.TrimSpace(spec.Path),
			Name:          menuKey,
			Component:     strings.TrimSpace(spec.Component),
			PageKey:       "",
			PermissionKey: "",
			DefaultTitle:  strings.TrimSpace(spec.Title),
			DefaultIcon:   strings.TrimSpace(spec.Icon),
			Status:        "normal",
			Meta:          meta,
		}
		if err := database.DB.Create(definition).Error; err != nil {
			return nil, err
		}
	} else {
		updates := map[string]any{
			"kind":          normalizeMenuSeedKind(spec),
			"path":          strings.TrimSpace(spec.Path),
			"name":          menuKey,
			"component":     strings.TrimSpace(spec.Component),
			"default_title": strings.TrimSpace(spec.Title),
			"default_icon":  strings.TrimSpace(spec.Icon),
			"status":        "normal",
			"meta":          meta,
			"updated_at":    time.Now(),
		}
		if err := database.DB.Model(definition).Updates(updates).Error; err != nil {
			return nil, err
		}
		definition.Kind = normalizeMenuSeedKind(spec)
		definition.Path = strings.TrimSpace(spec.Path)
		definition.Name = menuKey
		definition.Component = strings.TrimSpace(spec.Component)
		definition.DefaultTitle = strings.TrimSpace(spec.Title)
		definition.DefaultIcon = strings.TrimSpace(spec.Icon)
		definition.Status = "normal"
		definition.Meta = meta
	}

	spaceKey := normalizeMenuSeedSpaceKey(spec.SpaceKey)
	parentMenuKey := strings.TrimSpace(spec.ParentName)
	placement := &systemmodels.SpaceMenuPlacement{}
	err = database.DB.Where("app_key = ? AND space_key = ? AND menu_key = ? AND deleted_at IS NULL", appKey, spaceKey, menuKey).First(placement).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		placement = &systemmodels.SpaceMenuPlacement{
			AppKey:        appKey,
			SpaceKey:      spaceKey,
			MenuKey:       menuKey,
			ParentMenuKey: parentMenuKey,
			SortOrder:     spec.SortOrder,
			Hidden:        false,
			MetaOverride:  usermodel.MetaJSON{},
		}
		if err := database.DB.Create(placement).Error; err != nil {
			return nil, err
		}
	} else {
		updates := map[string]any{
			"parent_menu_key": parentMenuKey,
			"sort_order":      spec.SortOrder,
			"hidden":          false,
			"updated_at":      time.Now(),
		}
		if err := database.DB.Model(placement).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return definition, nil
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

	var definition systemmodels.MenuDefinition
	if err := database.DB.Where("app_key = ? AND name = ?", systemmodels.DefaultAppKey, targetName).First(&definition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	meta := definition.Meta
	if meta == nil {
		meta = usermodel.MetaJSON{}
	}
	if current := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", meta["accessMode"]))); current == targetMode {
		return nil
	}
	meta["accessMode"] = targetMode
	return database.DB.Model(&definition).Update("meta", meta).Error
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
	if err := database.DB.Where("app_key = ? AND name IN ?", systemmodels.DefaultAppKey, deprecatedNames).Delete(&systemmodels.MenuDefinition{}).Error; err != nil {
		return err
	}
	if err := database.DB.Where("app_key = ? AND menu_key IN ?", systemmodels.DefaultAppKey, deprecatedNames).Delete(&systemmodels.SpaceMenuPlacement{}).Error; err != nil {
		return err
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

func ensureNavigationModelFoundation(logger *zap.Logger) error {
	statements := []string{
		`ALTER TABLE menus ADD COLUMN IF NOT EXISTS kind varchar(20) NOT NULL DEFAULT 'directory'`,
		`ALTER TABLE menus ALTER COLUMN kind SET DEFAULT 'directory'`,
		`CREATE INDEX IF NOT EXISTS idx_menus_kind ON menus (kind)`,
		`CREATE TABLE IF NOT EXISTS page_space_bindings (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			page_id uuid NOT NULL,
			space_key varchar(100) NOT NULL,
			created_at timestamptz NOT NULL DEFAULT NOW(),
			updated_at timestamptz NOT NULL DEFAULT NOW(),
			deleted_at timestamptz NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_page_space_bindings_page_space_unique
			ON page_space_bindings (page_id, space_key)
			WHERE deleted_at IS NULL`,
		`UPDATE menus
			SET kind = CASE
				WHEN COALESCE(NULLIF(TRIM(COALESCE(meta->>'link', '')), ''), '') <> '' THEN 'external'
				WHEN COALESCE(NULLIF(TRIM(component), ''), '') <> '' AND component <> '/index/index' THEN 'entry'
				ELSE 'directory'
			END`,
		`INSERT INTO page_space_bindings (page_id, space_key, created_at, updated_at)
			SELECT id, LOWER(TRIM(space_key)), NOW(), NOW()
			FROM ui_pages
			WHERE COALESCE(NULLIF(TRIM(space_key), ''), '') <> ''
			  AND LOWER(TRIM(space_key)) <> 'default'
			  AND parent_menu_id IS NULL
			  AND COALESCE(NULLIF(TRIM(parent_page_key), ''), '') = ''
			ON CONFLICT DO NOTHING`,
	}
	for _, statement := range statements {
		if err := database.DB.Exec(statement).Error; err != nil {
			return err
		}
	}
	if logger != nil {
		logger.Info("Navigation model foundation ensured")
	}
	return nil
}

func cleanupMenuBackedEntryPages(logger *zap.Logger) error {
	var menus []usermodel.Menu
	if err := database.DB.Order("sort_order ASC, created_at ASC").Find(&menus).Error; err != nil {
		return err
	}
	menuMap := make(map[uuid.UUID]usermodel.Menu, len(menus))
	for _, item := range menus {
		menuMap[item.ID] = item
	}
	fullPathMap := make(map[uuid.UUID]string, len(menus))
	var resolveMenuPath func(menu usermodel.Menu) string
	resolveMenuPath = func(menu usermodel.Menu) string {
		if cached := strings.TrimSpace(fullPathMap[menu.ID]); cached != "" {
			return cached
		}
		parentPath := ""
		if menu.ParentID != nil {
			if parent, ok := menuMap[*menu.ParentID]; ok {
				parentPath = resolveMenuPath(parent)
			}
		}
		fullPath := buildMenuSeedFullPath(menu.Path, parentPath)
		fullPathMap[menu.ID] = fullPath
		return fullPath
	}
	for _, item := range menus {
		resolveMenuPath(item)
	}

	var pages []systemmodels.UIPage
	if err := database.DB.Order("sort_order ASC, created_at ASC").Find(&pages).Error; err != nil {
		return err
	}
	pageMap := make(map[string]systemmodels.UIPage, len(pages))
	for _, item := range pages {
		pageMap[strings.TrimSpace(item.PageKey)] = item
	}
	deleteIDs := make([]uuid.UUID, 0)
	for _, page := range pages {
		if page.ParentMenuID == nil {
			continue
		}
		if page.PageType == systemmodels.PageTypeGroup || page.PageType == systemmodels.PageTypeDisplayGroup {
			continue
		}
		menu, ok := menuMap[*page.ParentMenuID]
		if !ok || deriveMenuKind(menu) != systemmodels.MenuKindEntry {
			continue
		}
		if normalizeSeedRoutePath(page.RoutePath) != normalizeSeedRoutePath(fullPathMap[*page.ParentMenuID]) {
			continue
		}
		if strings.TrimSpace(page.Component) == "" || strings.TrimSpace(page.Component) != strings.TrimSpace(menu.Component) {
			continue
		}
		if hasPageDescendants(page.PageKey, pageMap) {
			continue
		}
		deleteIDs = append(deleteIDs, page.ID)
	}
	if len(deleteIDs) == 0 {
		return nil
	}
	if err := database.DB.Where("id IN ?", deleteIDs).Delete(&systemmodels.UIPage{}).Error; err != nil {
		return err
	}
	if logger != nil {
		logger.Info("Menu-backed entry pages cleaned", zap.Int("deleted_count", len(deleteIDs)))
	}
	return nil
}

func cleanupSpaceClonedPages(logger *zap.Logger) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var clonedPages []systemmodels.UIPage
		if err := tx.
			Where("deleted_at IS NULL").
			Where("page_key LIKE ?", "%.ops").
			Order("created_at ASC").
			Find(&clonedPages).Error; err != nil {
			return err
		}
		if len(clonedPages) == 0 {
			return nil
		}

		var allPages []systemmodels.UIPage
		if err := tx.Where("deleted_at IS NULL").Find(&allPages).Error; err != nil {
			return err
		}
		pageMap := make(map[string]systemmodels.UIPage, len(allPages))
		for _, item := range allPages {
			pageMap[strings.TrimSpace(item.PageKey)] = item
		}

		deletedCount := 0
		skippedKeys := make([]string, 0)
		for _, page := range clonedPages {
			clonedKey := strings.TrimSpace(page.PageKey)
			canonicalKey := strings.TrimSpace(strings.TrimSuffix(clonedKey, ".ops"))
			if canonicalKey == "" || canonicalKey == clonedKey {
				skippedKeys = append(skippedKeys, clonedKey)
				continue
			}
			canonical, ok := pageMap[canonicalKey]
			if !ok {
				skippedKeys = append(skippedKeys, clonedKey)
				continue
			}

			if err := tx.Model(&systemmodels.UIPage{}).
				Where("parent_page_key = ?", clonedKey).
				Update("parent_page_key", canonicalKey).Error; err != nil {
				return err
			}
			if err := tx.Model(&systemmodels.UIPage{}).
				Where("display_group_key = ?", clonedKey).
				Update("display_group_key", canonicalKey).Error; err != nil {
				return err
			}

			// 历史空间克隆中的独立页收敛后，改用 page_space_bindings 表达“仅暴露到少量空间”。
			if page.ParentMenuID == nil &&
				strings.TrimSpace(page.ParentPageKey) == "" &&
				normalizeMenuSeedSpaceKey(page.SpaceKey) != systemmodels.DefaultMenuSpaceKey {
				if err := tx.Exec(
					`INSERT INTO page_space_bindings (page_id, space_key, created_at, updated_at)
					 VALUES (?, ?, NOW(), NOW())
					 ON CONFLICT DO NOTHING`,
					canonical.ID,
					normalizeMenuSeedSpaceKey(page.SpaceKey),
				).Error; err != nil {
					return err
				}
			}

			if err := tx.Where("page_id = ?", page.ID).Delete(&systemmodels.PageSpaceBinding{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&systemmodels.UIPage{}, "id = ?", page.ID).Error; err != nil {
				return err
			}
			deletedCount++
		}

		if logger != nil {
			fields := []zap.Field{zap.Int("deleted_count", deletedCount)}
			if len(skippedKeys) > 0 {
				fields = append(fields, zap.Strings("skipped_page_keys", skippedKeys))
			}
			logger.Info("Space-cloned pages cleaned up", fields...)
		}
		return nil
	})
}

func cleanupLegacyOpsSpace(logger *zap.Logger) error {
	const legacySpaceKey = "ops"

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var menuIDs []uuid.UUID
		if err := tx.Model(&usermodel.Menu{}).
			Where("deleted_at IS NULL AND COALESCE(NULLIF(space_key, ''), ?) = ?", systemmodels.DefaultMenuSpaceKey, legacySpaceKey).
			Pluck("id", &menuIDs).Error; err != nil {
			return err
		}
		if len(menuIDs) > 0 {
			if err := tx.Where("menu_id IN ?", menuIDs).Delete(&usermodel.FeaturePackageMenu{}).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM role_hidden_menus WHERE menu_id IN ?", menuIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM collaboration_workspace_blocked_menus WHERE menu_id IN ?", menuIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM user_hidden_menus WHERE menu_id IN ?", menuIDs).Error; err != nil {
				return err
			}
			if err := tx.Where("parent_menu_id IN ?", menuIDs).Delete(&systemmodels.UIPage{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", menuIDs).Delete(&usermodel.Menu{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("space_key = ?", legacySpaceKey).Delete(&systemmodels.PageSpaceBinding{}).Error; err != nil {
			return err
		}
		if err := tx.Where("COALESCE(NULLIF(space_key, ''), ?) = ?", systemmodels.DefaultMenuSpaceKey, legacySpaceKey).
			Delete(&systemmodels.UIPage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("space_key = ?", legacySpaceKey).Delete(&usermodel.MenuBackup{}).Error; err != nil {
			return err
		}
		if err := tx.Where("space_key = ?", legacySpaceKey).Delete(&systemmodels.MenuSpaceHostBinding{}).Error; err != nil {
			return err
		}
		if err := tx.Where("space_key = ?", legacySpaceKey).Delete(&systemmodels.MenuSpace{}).Error; err != nil {
			return err
		}

		if logger != nil {
			logger.Info("Legacy ops menu space cleaned", zap.Int("menu_count", len(menuIDs)))
		}
		return nil
	})
}

func cleanupGlobalPageBindings(logger *zap.Logger) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"parent_menu_id":   nil,
			"parent_page_key":  "",
			"active_menu_path": "",
		}
		result := tx.Model(&systemmodels.UIPage{}).
			Where("page_type = ?", systemmodels.PageTypeGlobal).
			Where("parent_menu_id IS NOT NULL OR COALESCE(parent_page_key, '') <> '' OR COALESCE(active_menu_path, '') <> ''").
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if logger != nil {
			logger.Info("Global page binding cleanup completed", zap.Int64("updated_count", result.RowsAffected))
		}
		return nil
	})
}

func hasPageDescendants(pageKey string, pageMap map[string]systemmodels.UIPage) bool {
	target := strings.TrimSpace(pageKey)
	if target == "" {
		return false
	}
	for _, item := range pageMap {
		if strings.TrimSpace(item.ParentPageKey) == target || strings.TrimSpace(item.DisplayGroupKey) == target {
			return true
		}
	}
	return false
}

func deriveMenuKind(item usermodel.Menu) string {
	link := ""
	if item.Meta != nil {
		if value, ok := item.Meta["link"].(string); ok {
			link = strings.TrimSpace(value)
		}
	}
	switch {
	case link != "":
		return systemmodels.MenuKindExternal
	case strings.TrimSpace(item.Component) != "" && strings.TrimSpace(item.Component) != "/index/index":
		return systemmodels.MenuKindEntry
	default:
		return systemmodels.MenuKindDirectory
	}
}

func buildMenuSeedFullPath(path, parentPath string) string {
	target := strings.TrimSpace(path)
	if target == "" {
		return normalizeSeedRoutePath(parentPath)
	}
	if strings.HasPrefix(target, "/") {
		return normalizeSeedRoutePath(target)
	}
	parent := normalizeSeedRoutePath(parentPath)
	if parent == "" || parent == "/" {
		return normalizeSeedRoutePath("/" + target)
	}
	return normalizeSeedRoutePath(strings.TrimRight(parent, "/") + "/" + strings.TrimLeft(target, "/"))
}

func normalizeSeedRoutePath(path string) string {
	target := strings.TrimSpace(path)
	if target == "" {
		return ""
	}
	normalized := "/" + strings.TrimLeft(target, "/")
	normalized = strings.ReplaceAll(normalized, "//", "/")
	if normalized != "/" {
		normalized = strings.TrimRight(normalized, "/")
	}
	return normalized
}

func syncDefaultPageSeedByKey(pageKey string) (*systemmodels.UIPage, error) {
	targetKey := strings.TrimSpace(pageKey)
	if targetKey == "" {
		return nil, fmt.Errorf("page seed key is required")
	}

	pageSeeds := append(permissionseed.DefaultPages(), permissionseed.LegacyMenuBackedPages()...)
	for _, spec := range pageSeeds {
		if strings.TrimSpace(spec.PageKey) != targetKey {
			continue
		}
		if spec.ParentMenuName != "" {
			if _, err := syncDefaultMenuSeedByName(spec.ParentMenuName); err != nil {
				return nil, err
			}
		}
		if spec.DisplayGroupKey != "" && spec.DisplayGroupKey != targetKey {
			if _, err := syncDefaultPageSeedByKey(spec.DisplayGroupKey); err != nil {
				return nil, err
			}
		}
		if spec.ParentPageKey != "" && spec.ParentPageKey != targetKey {
			if _, err := syncDefaultPageSeedByKey(spec.ParentPageKey); err != nil {
				return nil, err
			}
		}
		return syncUIPageSeed(spec)
	}

	return nil, fmt.Errorf("default page seed not found: %s", targetKey)
}

func syncUIPageSeed(spec permissionseed.PageSeed) (*systemmodels.UIPage, error) {
	pageType := strings.TrimSpace(spec.PageType)
	if pageType == "" {
		pageType = "inner"
	}
	source := strings.TrimSpace(spec.Source)
	if source == "" {
		source = "manual"
	}
	breadcrumbMode := strings.TrimSpace(spec.BreadcrumbMode)
	if breadcrumbMode == "" {
		breadcrumbMode = "inherit_menu"
	}
	accessMode := strings.TrimSpace(spec.AccessMode)
	if accessMode == "" {
		accessMode = "inherit"
	}
	status := strings.TrimSpace(spec.Status)
	if status == "" {
		status = "normal"
	}

	var parentMenuID *uuid.UUID
	if pageType != systemmodels.PageTypeGlobal {
		if parentMenuName := strings.TrimSpace(spec.ParentMenuName); parentMenuName != "" {
			var parentMenu systemmodels.MenuDefinition
			if err := database.DB.Where("app_key = ? AND name = ?", systemmodels.DefaultAppKey, parentMenuName).First(&parentMenu).Error; err != nil {
				return nil, err
			}
			parentMenuID = &parentMenu.ID
		}
	}

	meta := spec.Meta
	if meta == nil {
		meta = usermodel.MetaJSON{}
	}
	delete(meta, "spaceKeys")
	delete(meta, "spaceScope")

	visibilityScope := normalizePageSeedVisibilityScope(pageType, spec.VisibilityScope, parentMenuID, spec.ParentPageKey)
	spaceKeys := normalizePageSeedBindingKeys(spec.SpaceKey, spec.SpaceKeys, pageType, visibilityScope, parentMenuID, spec.ParentPageKey)
	parentPageKey := strings.TrimSpace(spec.ParentPageKey)
	activeMenuPath := strings.TrimSpace(spec.ActiveMenuPath)
	if pageType == systemmodels.PageTypeGlobal {
		parentPageKey = ""
		activeMenuPath = ""
	}
	switch visibilityScope {
	case "spaces":
		meta["spaceKeys"] = spaceKeys
		meta["spaceScope"] = "spaces"
	case "inherit":
		meta["spaceScope"] = "inherit"
	default:
		meta["spaceScope"] = "app"
	}

	item := &systemmodels.UIPage{
		AppKey:            systemmodels.DefaultAppKey,
		PageKey:           strings.TrimSpace(spec.PageKey),
		Name:              strings.TrimSpace(spec.Name),
		RouteName:         strings.TrimSpace(spec.RouteName),
		RoutePath:         strings.TrimSpace(spec.RoutePath),
		Component:         strings.TrimSpace(spec.Component),
		SpaceKey:          "",
		PageType:          pageType,
		VisibilityScope:   visibilityScope,
		Source:            source,
		ModuleKey:         strings.TrimSpace(spec.ModuleKey),
		SortOrder:         spec.SortOrder,
		ParentMenuID:      parentMenuID,
		ParentPageKey:     parentPageKey,
		DisplayGroupKey:   strings.TrimSpace(spec.DisplayGroupKey),
		ActiveMenuPath:    activeMenuPath,
		BreadcrumbMode:    breadcrumbMode,
		AccessMode:        accessMode,
		PermissionKey:     strings.TrimSpace(spec.PermissionKey),
		InheritPermission: spec.InheritPermission,
		KeepAlive:         spec.KeepAlive,
		IsFullPage:        spec.IsFullPage,
		Status:            status,
		Meta:              meta,
	}

	var existing systemmodels.UIPage
	if err := database.DB.Where("app_key = ? AND page_key = ?", systemmodels.DefaultAppKey, item.PageKey).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := database.DB.Create(item).Error; err != nil {
				return nil, err
			}
			if err := syncSeedPageSpaceBindings(item.ID, visibilityScope, spaceKeys); err != nil {
				return nil, err
			}
			return item, nil
		}
		return nil, err
	}

	item.ID = existing.ID
	updates := map[string]interface{}{
		"app_key":            item.AppKey,
		"name":               item.Name,
		"route_name":         item.RouteName,
		"route_path":         item.RoutePath,
		"component":          item.Component,
		"space_key":          item.SpaceKey,
		"page_type":          item.PageType,
		"visibility_scope":   item.VisibilityScope,
		"source":             item.Source,
		"module_key":         item.ModuleKey,
		"sort_order":         item.SortOrder,
		"parent_menu_id":     item.ParentMenuID,
		"parent_page_key":    item.ParentPageKey,
		"display_group_key":  item.DisplayGroupKey,
		"active_menu_path":   item.ActiveMenuPath,
		"breadcrumb_mode":    item.BreadcrumbMode,
		"access_mode":        item.AccessMode,
		"permission_key":     item.PermissionKey,
		"inherit_permission": item.InheritPermission,
		"keep_alive":         item.KeepAlive,
		"is_full_page":       item.IsFullPage,
		"status":             item.Status,
		"meta":               item.Meta,
		"updated_at":         time.Now(),
	}
	if err := database.DB.Model(&existing).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := syncSeedPageSpaceBindings(item.ID, visibilityScope, spaceKeys); err != nil {
		return nil, err
	}
	return item, nil
}

func normalizePageSeedVisibilityScope(pageType, value string, parentMenuID *uuid.UUID, parentPageKey string) string {
	switch strings.TrimSpace(pageType) {
	case systemmodels.PageTypeGlobal:
		if strings.TrimSpace(value) == "spaces" {
			return "spaces"
		}
		return "app"
	case systemmodels.PageTypeInner:
		return "inherit"
	case systemmodels.PageTypeStandalone, systemmodels.PageTypeGroup, systemmodels.PageTypeDisplayGroup:
		if parentMenuID != nil || strings.TrimSpace(parentPageKey) != "" {
			return "inherit"
		}
		if strings.TrimSpace(value) == "spaces" {
			return "spaces"
		}
		return "app"
	default:
		return "app"
	}
}

func normalizePageSeedBindingKeys(spaceKey string, spaceKeys []string, pageType, visibilityScope string, parentMenuID *uuid.UUID, parentPageKey string) []string {
	if normalizePageSeedVisibilityScope(pageType, visibilityScope, parentMenuID, parentPageKey) != "spaces" {
		return []string{}
	}
	candidates := make([]string, 0, len(spaceKeys)+1)
	for _, item := range spaceKeys {
		target := strings.TrimSpace(item)
		if target != "" {
			candidates = append(candidates, target)
		}
	}
	if len(candidates) == 0 {
		if target := strings.TrimSpace(spaceKey); target != "" {
			candidates = append(candidates, target)
		}
	}
	if len(candidates) == 0 {
		candidates = append(candidates, systemmodels.DefaultMenuSpaceKey)
	}
	seen := make(map[string]struct{}, len(candidates))
	result := make([]string, 0, len(candidates))
	for _, item := range candidates {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func syncSeedPageSpaceBindings(pageID uuid.UUID, visibilityScope string, spaceKeys []string) error {
	if err := database.DB.Where("app_key = ? AND page_id = ?", systemmodels.DefaultAppKey, pageID).Delete(&systemmodels.PageSpaceBinding{}).Error; err != nil {
		return err
	}
	if strings.TrimSpace(visibilityScope) != "spaces" {
		return nil
	}
	for _, key := range spaceKeys {
		binding := systemmodels.PageSpaceBinding{
			AppKey:   systemmodels.DefaultAppKey,
			PageID:   pageID,
			SpaceKey: strings.TrimSpace(key),
		}
		if err := database.DB.Create(&binding).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedDefaultMessageTemplates(logger *zap.Logger) error {
	items := []systemmodels.MessageTemplate{
		{
			TemplateKey:     "platform.notice.all_users",
			Name:            "平台全员公告",
			Description:     "平台向全部用户发送公告通知",
			MessageType:     "notice",
			OwnerScope:      "platform",
			AudienceType:    "all_users",
			TitleTemplate:   "{{title}}",
			SummaryTemplate: "{{summary}}",
			ContentTemplate: "{{content}}",
			ActionType:      "none",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"builtin": true},
		},
		{
			TemplateKey:     "platform.notice.collaboration_workspace_admins",
			Name:            "平台协作空间管理员提醒",
			Description:     "平台向协作空间管理员发送治理提醒",
			MessageType:     "message",
			OwnerScope:      "platform",
			AudienceType:    "collaboration_workspace_admins",
			TitleTemplate:   "{{title}}",
			SummaryTemplate: "{{summary}}",
			ContentTemplate: "{{content}}",
			ActionType:      "route",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"builtin": true},
		},
		{
			TemplateKey:     "collaboration_workspace.notice.collaboration_workspace_users",
			Name:            "协作空间公告模板",
			Description:     "协作空间管理员向指定协作空间发送公告或待办",
			MessageType:     "todo",
			OwnerScope:      "collaboration",
			AudienceType:    "collaboration_workspace_users",
			TitleTemplate:   "{{title}}",
			SummaryTemplate: "{{summary}}",
			ContentTemplate: "{{content}}",
			ActionType:      "route",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"builtin": true},
		},
	}

	for _, item := range items {
		var existing systemmodels.MessageTemplate
		err := database.DB.Where("template_key = ?", item.TemplateKey).First(&existing).Error
		if err == nil {
			if updateErr := database.DB.Model(&existing).Updates(map[string]interface{}{
				"name":                   item.Name,
				"description":            item.Description,
				"message_type":           item.MessageType,
				"owner_scope":            item.OwnerScope,
				"audience_type":          item.AudienceType,
				"title_template":         item.TitleTemplate,
				"summary_template":       item.SummaryTemplate,
				"content_template":       item.ContentTemplate,
				"action_type":            item.ActionType,
				"action_target_template": item.ActionTargetTemplate,
				"status":                 item.Status,
				"meta":                   item.Meta,
				"updated_at":             time.Now(),
			}).Error; updateErr != nil {
				return updateErr
			}
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if createErr := database.DB.Create(&item).Error; createErr != nil {
			return createErr
		}
	}

	logger.Info("Default message templates synchronized")
	return nil
}

func ensureDefaultFastEnterConfig(logger *zap.Logger) error {
	service := systemservice.NewFastEnterService(database.DB)
	config, err := service.GetConfig()
	if err != nil {
		return err
	}
	if _, err := service.SaveConfig(config); err != nil {
		return err
	}
	if logger != nil {
		logger.Info("Default fast enter config ensured")
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

func deleteMenuTreeByName(name string) error {
	target := strings.TrimSpace(name)
	if target == "" {
		return nil
	}
	var menu usermodel.Menu
	if err := database.DB.Where("name = ?", target).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return deleteMenuTree(menu.ID)
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

func cleanupDeprecatedPermissionKeys(logger *zap.Logger) error {
	deprecatedKeys := []string{
		"system.user.assign_action",
	}
	if err := database.DB.Where("permission_key IN ?", deprecatedKeys).Delete(&usermodel.PermissionKey{}).Error; err != nil {
		return err
	}
	logger.Info("Deprecated permission keys cleaned", zap.Int("count", len(deprecatedKeys)))
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
	boundaryService := collaborationworkspaceboundary.NewService(database.DB)
	platformService := platformaccess.NewService(database.DB)
	roleSnapshotService := platformroleaccess.NewService(database.DB)
	refresher := permissionrefresh.NewService(database.DB, boundaryService, platformService, roleSnapshotService)

	defaultRoleCodes := permissionseed.DefaultRoleCodes()
	var roles []usermodel.Role
	if err := database.DB.Where("collaboration_workspace_id IS NULL AND code IN ?", defaultRoleCodes).Find(&roles).Error; err != nil {
		return err
	}
	roleIDs := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	if err := refresher.RefreshPersonalWorkspaceRoles(roleIDs); err != nil {
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

func ensureAppScopedSnapshotPrimaryKeysMigration() error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	type snapshotSpec struct {
		table string
		cols  []string
	}

	specs := []snapshotSpec{
		{table: "platform_user_access_snapshots", cols: []string{"app_key", "user_id"}},
		{table: "platform_role_access_snapshots", cols: []string{"app_key", "role_id"}},
		{table: "collaboration_workspace_access_snapshots", cols: []string{"app_key", "collaboration_workspace_id"}},
		{table: "collaboration_workspace_role_access_snapshots", cols: []string{"app_key", "collaboration_workspace_id", "role_id"}},
	}

	for _, spec := range specs {
		if err := database.DB.Exec(
			"ALTER TABLE " + spec.table + " ADD COLUMN IF NOT EXISTS app_key varchar(100) NOT NULL DEFAULT '" + systemmodels.DefaultAppKey + "'",
		).Error; err != nil {
			return err
		}
		if err := database.DB.Exec(
			"UPDATE " + spec.table + " SET app_key = '" + systemmodels.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		).Error; err != nil {
			return err
		}
		if err := database.DB.Exec("ALTER TABLE " + spec.table + " DROP CONSTRAINT IF EXISTS " + spec.table + "_pkey").Error; err != nil {
			return err
		}
		if err := database.DB.Exec(
			"ALTER TABLE " + spec.table + " ADD PRIMARY KEY (" + strings.Join(spec.cols, ", ") + ")",
		).Error; err != nil {
			return err
		}
	}

	return nil
}

func initWorkspaceBaseline(logger *zap.Logger) error {
	service := workspace.NewService(database.DB, logger)
	if err := service.EnsureWorkspaceBackfill(); err != nil {
		return err
	}
	logger.Info("Workspace baseline backfilled")
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
	if err := database.DB.Exec(`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS app_scope varchar(20) NOT NULL DEFAULT 'shared'`).Error; err != nil {
		return err
	}
	if err := database.DB.Exec(`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS app_key varchar(100) NOT NULL DEFAULT '` + systemmodels.DefaultAppKey + `'`).Error; err != nil {
		return err
	}
	if err := database.DB.Exec(`UPDATE api_endpoints SET app_key = '` + systemmodels.DefaultAppKey + `' WHERE COALESCE(TRIM(app_key), '') = ''`).Error; err != nil {
		return err
	}
	if err := database.DB.Exec(`UPDATE api_endpoints SET app_scope = '` + systemmodels.AppScopeShared + `' WHERE COALESCE(TRIM(app_scope), '') = '' AND (path LIKE '/api/v1/auth/%' OR path = '/api/v1/pages/runtime/public' OR path LIKE '/open/v1/%' OR path = '/health')`).Error; err != nil {
		return err
	}
	if err := database.DB.Exec(`UPDATE api_endpoints SET app_scope = '` + systemmodels.AppScopeApp + `' WHERE COALESCE(TRIM(app_scope), '') = ''`).Error; err != nil {
		return err
	}
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

func normalizeMenuSpaceStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "disabled":
		return "disabled"
	default:
		return "normal"
	}
}

func normalizeMenuSeedSpaceKey(value string) string {
	target := strings.TrimSpace(strings.ToLower(value))
	if target == "" {
		return systemmodels.DefaultMenuSpaceKey
	}
	return target
}

func normalizeMenuSeedKind(spec permissionseed.MenuSeed) string {
	target := strings.TrimSpace(strings.ToLower(spec.Kind))
	switch target {
	case systemmodels.MenuKindDirectory, systemmodels.MenuKindEntry, systemmodels.MenuKindExternal:
		return target
	}
	return deriveMenuSeedKindFromSpec(spec)
}

func deriveMenuSeedKindFromSpec(spec permissionseed.MenuSeed) string {
	link := ""
	if spec.Meta != nil {
		if value, ok := spec.Meta["link"].(string); ok {
			link = strings.TrimSpace(value)
		}
	}
	switch {
	case link != "":
		return systemmodels.MenuKindExternal
	case strings.TrimSpace(spec.Component) != "" && strings.TrimSpace(spec.Component) != "/index/index":
		return systemmodels.MenuKindEntry
	default:
		return systemmodels.MenuKindDirectory
	}
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

func cleanupDuplicateTeamRoots(logger *zap.Logger) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var roots []usermodel.Menu
		if err := tx.
			Where("deleted_at IS NULL").
			Where("name = ? AND path = ? AND component = ?", "TeamRoot", "/team", "/index/index").
			Order("created_at ASC").
			Find(&roots).Error; err != nil {
			return err
		}
		if len(roots) <= 1 {
			return nil
		}

		canonical := roots[0]
		childCache := make(map[uuid.UUID][]usermodel.Menu, len(roots))
		for _, root := range roots {
			var children []usermodel.Menu
			if err := tx.Where("deleted_at IS NULL AND parent_id = ?", root.ID).Order("created_at ASC").Find(&children).Error; err != nil {
				return err
			}
			childCache[root.ID] = children
			for _, child := range children {
				if child.Name == "TeamManage" {
					canonical = root
				}
			}
		}

		canonicalChildren := make(map[string]usermodel.Menu)
		for _, child := range childCache[canonical.ID] {
			canonicalChildren[child.Name] = child
		}

		for _, root := range roots {
			if root.ID == canonical.ID {
				continue
			}

			for _, child := range childCache[root.ID] {
				if target, ok := canonicalChildren[child.Name]; ok {
					if err := transferMenuReferences(tx, child.ID, target.ID); err != nil {
						return err
					}
					if err := tx.Model(&usermodel.Menu{}).Where("parent_id = ?", child.ID).Update("parent_id", target.ID).Error; err != nil {
						return err
					}
					if err := tx.Delete(&usermodel.Menu{}, "id = ?", child.ID).Error; err != nil {
						return err
					}
					continue
				}

				if err := tx.Model(&usermodel.Menu{}).Where("id = ?", child.ID).Update("parent_id", canonical.ID).Error; err != nil {
					return err
				}
			}

			if err := transferMenuReferences(tx, root.ID, canonical.ID); err != nil {
				return err
			}
			if err := tx.Delete(&usermodel.Menu{}, "id = ?", root.ID).Error; err != nil {
				return err
			}
		}

		logger.Info("Duplicate TeamRoot menus cleaned up", zap.Int("roots", len(roots)), zap.String("canonical_id", canonical.ID.String()))
		return nil
	})
}

func transferMenuReferences(tx *gorm.DB, sourceMenuID, targetMenuID uuid.UUID) error {
	if sourceMenuID == targetMenuID {
		return nil
	}

	if err := tx.Model(&systemmodels.UIPage{}).
		Where("parent_menu_id = ?", sourceMenuID).
		Update("parent_menu_id", targetMenuID).Error; err != nil {
		return err
	}

	statements := []string{
		`INSERT INTO feature_package_menus (package_id, menu_id)
		 SELECT package_id, ? FROM feature_package_menus WHERE menu_id = ?
		 ON CONFLICT DO NOTHING`,
		`DELETE FROM feature_package_menus WHERE menu_id = ?`,
		`INSERT INTO role_hidden_menus (role_id, menu_id, created_at)
		 SELECT role_id, ?, created_at FROM role_hidden_menus WHERE menu_id = ?
		 ON CONFLICT DO NOTHING`,
		`DELETE FROM role_hidden_menus WHERE menu_id = ?`,
		`INSERT INTO collaboration_workspace_blocked_menus (collaboration_workspace_id, menu_id, created_at, updated_at)
		 SELECT collaboration_workspace_id, ?, created_at, updated_at FROM collaboration_workspace_blocked_menus WHERE menu_id = ?
		 ON CONFLICT DO NOTHING`,
		`DELETE FROM collaboration_workspace_blocked_menus WHERE menu_id = ?`,
		`INSERT INTO user_hidden_menus (user_id, menu_id, created_at, updated_at)
		 SELECT user_id, ?, created_at, updated_at FROM user_hidden_menus WHERE menu_id = ?
		 ON CONFLICT DO NOTHING`,
		`DELETE FROM user_hidden_menus WHERE menu_id = ?`,
	}

	for index, statement := range statements {
		switch index {
		case 0, 2, 4, 6:
			if err := tx.Exec(statement, targetMenuID, sourceMenuID).Error; err != nil {
				return err
			}
		default:
			if err := tx.Exec(statement, sourceMenuID).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
