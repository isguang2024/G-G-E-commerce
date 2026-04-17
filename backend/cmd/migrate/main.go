package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	apirouter "github.com/maben/backend/internal/api/router"
	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/logpolicy"
	"github.com/maben/backend/internal/modules/observability/telemetry"
	"github.com/maben/backend/internal/modules/system/dictionary"
	systemmodels "github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/siteconfig"
	space "github.com/maben/backend/internal/modules/system/space"
	systemservice "github.com/maben/backend/internal/modules/system/system"
	"github.com/maben/backend/internal/modules/system/upload"
	usermodel "github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/maben/backend/internal/pkg/database"
	"github.com/maben/backend/internal/pkg/logger"
	"github.com/maben/backend/internal/pkg/password"
	"github.com/maben/backend/internal/pkg/permissionkey"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
	"github.com/maben/backend/internal/pkg/permissionseed"
	"github.com/maben/backend/internal/pkg/platformaccess"
	"github.com/maben/backend/internal/pkg/platformroleaccess"
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

	// goose 迁移先跑（permission_keys baseline 等），随后 GORM AutoMigrate 接管 schema。
	if err := database.RunGooseMigrations(database.DB); err != nil {
		logger.Fatal("goose migration failed", zap.Error(err))
	}

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}

	if err := runRequiredMigrationTasks(logger, "schema-finalizers", []migrationTask{
		{
			Name: "feature_packages.app_keys",
			Run: func(logger *zap.Logger) error {
				return ensureFeaturePackageAppKeysColumn()
			},
		},
		{
			Name: "audit_logs.monthly_partitions",
			Run: func(logger *zap.Logger) error {
				return ensureAuditLogMonthlyPartitions(logger)
			},
		},
		{
			Name: "log_policy.compliance_defaults",
			Run: func(logger *zap.Logger) error {
				return logpolicy.EnsureCompliancePolicies(context.Background(), logpolicy.NewRepository(database.DB))
			},
		},
	}); err != nil {
		logger.Fatal("Required migration task failed", zap.Error(err))
	}

	logger.Info("Database migration completed successfully!")

	logger.Info("Using final workspace schema and canonical permission seeds")

	dictSvc := dictionary.NewService(database.DB, logger)
	runOptionalMigrationTasks(logger, "default-seeds", []migrationTask{
		{Name: "default.roles", Run: initDefaultRolesNoScope},
		{Name: "default.admin", Run: initDefaultAdmin},
		{Name: "default.menu_spaces", Run: initDefaultMenuSpaces},
		{Name: "default.menus", Run: initDefaultMenusNoScope},
		{Name: "default.pages", Run: initDefaultPages},
		{Name: "default.permission_groups", Run: initDefaultPermissionGroups},
		{Name: "default.permission_keys", Run: initDefaultPermissionKeysNoScope},
		{Name: "default.feature_packages", Run: initDefaultFeaturePackages},
		{Name: "default.feature_package_bundles", Run: initDefaultFeaturePackageBundles},
		{
			Name: "default.access_trace_navigation",
			Run: func(logger *zap.Logger) error {
				return ensureAccessTraceNavigationSeed()
			},
		},
		{Name: "default.role_feature_packages", Run: initDefaultRoleFeaturePackages},
		{
			Name: "default.register_system",
			Run: func(logger *zap.Logger) error {
				return permissionseed.EnsureRegisterSystemSeeds(database.DB)
			},
		},
		{
			Name: "default.social_auth_providers",
			Run: func(logger *zap.Logger) error {
				return permissionseed.EnsureSocialAuthProviders(database.DB)
			},
		},
		{
			Name: "default.upload",
			Run: func(logger *zap.Logger) error {
				return upload.EnsureDefaultSeeds(context.Background(), upload.NewRepository(database.DB), logger)
			},
		},
		{
			Name: "default.site_config",
			Run: func(logger *zap.Logger) error {
				return siteconfig.SeedDefaults(context.Background(), database.DB, logger)
			},
		},
		{
			Name: "default.builtin_dictionaries",
			Run: func(logger *zap.Logger) error {
				return dictSvc.EnsureBuiltinDicts(context.Background())
			},
		},
		{Name: "default.api_endpoint_categories", Run: initDefaultAPIEndpointCategories},
		{Name: "default.message_templates", Run: seedDefaultMessageTemplates},
		{Name: "default.fast_enter", Run: ensureDefaultFastEnterConfig},
	})

	runOptionalMigrationTasks(logger, "runtime-sync", []migrationTask{
		{Name: "sync.consolidate_permission_keys", Run: consolidatePermissionKeys},
		{Name: "sync.prune_feature_package_keys", Run: pruneBuiltinFeaturePackageKeys},
		{Name: "sync.backfill_media_upload_on_manage", Run: backfillCustomMediaUploadOnManage},
		{Name: "sync.openapi_endpoints", Run: initOpenAPIEndpoints},
		{
			Name: "sync.api_registry",
			Run: func(logger *zap.Logger) error {
				return syncAPIRegistry(logger, cfg)
			},
		},
		{Name: "sync.canonical_permission_keys", Run: syncCanonicalPermissionKeys},
		{Name: "sync.default_access_snapshots", Run: refreshDefaultAccessSnapshots},
	})

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

func runRequiredMigrationTasks(logger *zap.Logger, phase string, tasks []migrationTask) error {
	if len(tasks) == 0 {
		return nil
	}
	logger.Info("Running required migration tasks", zap.String("phase", phase), zap.Int("count", len(tasks)))
	for _, task := range tasks {
		if err := task.Run(logger); err != nil {
			return fmt.Errorf("%s: %w", task.Name, err)
		}
		logger.Info("Required migration task completed", zap.String("phase", phase), zap.String("task", task.Name))
	}
	return nil
}

func runOptionalMigrationTasks(logger *zap.Logger, phase string, tasks []migrationTask) {
	if len(tasks) == 0 {
		return
	}
	logger.Info("Running optional migration tasks", zap.String("phase", phase), zap.Int("count", len(tasks)))
	for _, task := range tasks {
		if err := task.Run(logger); err != nil {
			logger.Warn("Migration task failed", zap.String("phase", phase), zap.String("task", task.Name), zap.Error(err))
			continue
		}
		logger.Info("Migration task completed", zap.String("phase", phase), zap.String("task", task.Name))
	}
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

func ensureFeaturePackageAppKeysColumn() error {
	exists, err := hasColumn("feature_packages", "app_keys")
	if err != nil {
		return err
	}
	if !exists {
		if err := database.DB.Exec(`ALTER TABLE feature_packages ADD COLUMN app_keys jsonb NOT NULL DEFAULT '[]'::jsonb`).Error; err != nil {
			return err
		}
	}

	// Ensure historical rows have canonical array values.
	return database.DB.Exec(`
		UPDATE feature_packages
		SET app_keys = CASE
			WHEN COALESCE(TRIM(app_key), '') = '' THEN '[]'::jsonb
			ELSE jsonb_build_array(TRIM(app_key))
		END
		WHERE app_keys IS NULL OR jsonb_typeof(app_keys) <> 'array' OR jsonb_array_length(app_keys) = 0
	`).Error
}

func ensureAuditLogMonthlyPartitions(logger *zap.Logger) error {
	exists, err := hasTable("audit_logs")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	partitioned, err := isAuditLogsRangePartitioned()
	if err != nil {
		return err
	}
	if !partitioned {
		logger.Warn("audit_logs is not range-partitioned, skip monthly partition ensure")
		return nil
	}

	now := time.Now().UTC()
	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	next := current.AddDate(0, 1, 0)
	for _, start := range []time.Time{current, next} {
		if err := createAuditMonthlyPartition(start); err != nil {
			return err
		}
	}
	if err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS audit_logs_default PARTITION OF audit_logs DEFAULT`).Error; err != nil {
		return err
	}
	logger.Info("audit_logs monthly partitions ensured",
		zap.String("current_month", current.Format("2006-01")),
		zap.String("next_month", next.Format("2006-01")),
	)
	return nil
}

func isAuditLogsRangePartitioned() (bool, error) {
	var count int64
	err := database.DB.Raw(`
		SELECT COUNT(*)
		  FROM pg_partitioned_table pt
		  JOIN pg_class c ON c.oid = pt.partrelid
		  JOIN pg_namespace n ON n.oid = c.relnamespace
		 WHERE n.nspname = CURRENT_SCHEMA()
		   AND c.relname = 'audit_logs'
		   AND pt.partstrat = 'r'
	`).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func createAuditMonthlyPartition(monthStart time.Time) error {
	start := monthStart.UTC()
	end := start.AddDate(0, 1, 0)
	tableName := fmt.Sprintf("audit_logs_%04d_%02d", start.Year(), int(start.Month()))
	sql := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_logs FOR VALUES FROM ('%s') TO ('%s')",
		tableName,
		start.Format("2006-01-02 15:04:05Z07:00"),
		end.Format("2006-01-02 15:04:05Z07:00"),
	)
	return database.DB.Exec(sql).Error
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
	packageKeys := []string{"platform_admin.system_manage", "platform_admin.menu_manage"}
	var packages []systemmodels.FeaturePackage
	if err := database.DB.Where("package_key IN ?", packageKeys).Find(&packages).Error; err != nil {
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
		// ModuleCode 由 permission_groups JOIN 计算得出（model 字段 gorm:"-"），
		// 不是 permission_keys 的直接列，无需也无法写入。
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

// consolidatePermissionKeys 将历史遗留、已拆分过细的 permission_key 合并到新的
// 规范 key 上。合并对由 permissionKeyConsolidations() 维护，任务会幂等执行：
//
//  1. 将 feature_package_keys / user_action_permissions / role_disabled_actions /
//     collaboration_workspace_blocked_actions 的 action_id 从旧 key 重绑到新 key。
//  2. 将 api_endpoint_permission_bindings.permission_key 的值重写。
//  3. 将 ui_pages.permission_key 的值重写。
//  4. 软删除旧 permission_keys 行。
//
// 新 key 必须已由 default.permission_keys（EnsureDefaultPermissionKeys /
// EnsureOpenAPIPermissionKeys）落库；找不到新 key 时跳过并 warn 而不阻塞启动。
func consolidatePermissionKeys(logger *zap.Logger) error {
	pairs := permissionKeyConsolidations()
	mergedPairs := 0
	for _, pair := range pairs {
		fromKey := strings.TrimSpace(pair.From)
		toKey := strings.TrimSpace(pair.To)
		if fromKey == "" || toKey == "" || fromKey == toKey {
			continue
		}

		var legacy usermodel.PermissionKey
		err := database.DB.Where("permission_key = ? AND deleted_at IS NULL", fromKey).First(&legacy).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("lookup legacy permission key %s: %w", fromKey, err)
		}

		var canonical usermodel.PermissionKey
		err = database.DB.Where("permission_key = ? AND deleted_at IS NULL", toKey).First(&canonical).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("canonical permission key missing, skip consolidation",
				zap.String("from", fromKey), zap.String("to", toKey))
			continue
		}
		if err != nil {
			return fmt.Errorf("lookup canonical permission key %s: %w", toKey, err)
		}

		if err := rebindPermissionKeyReferences(legacy.ID, canonical.ID); err != nil {
			return fmt.Errorf("rebind references %s->%s: %w", fromKey, toKey, err)
		}

		if err := rebindAPIBindingPermissionKey(fromKey, toKey); err != nil {
			return fmt.Errorf("rebind api bindings %s->%s: %w", fromKey, toKey, err)
		}

		if err := database.DB.Exec(
			`UPDATE ui_pages SET permission_key = ?, updated_at = ?
			 WHERE permission_key = ? AND deleted_at IS NULL`,
			toKey, time.Now(), fromKey,
		).Error; err != nil {
			return fmt.Errorf("rewrite ui_pages permission_key %s->%s: %w", fromKey, toKey, err)
		}

		if err := database.DB.Delete(&legacy).Error; err != nil {
			return fmt.Errorf("soft delete legacy permission key %s: %w", fromKey, err)
		}

		mergedPairs++
		logger.Info("consolidated permission key",
			zap.String("from", fromKey), zap.String("to", toKey))
	}
	logger.Info("permission key consolidation applied", zap.Int("pairs", mergedPairs))
	return nil
}

// pruneBuiltinFeaturePackageKeys 将 IsBuiltin=true 功能包的 feature_package_keys
// 对齐到 DefaultFeaturePackages 里声明的 PermissionKeys 集合，删除声明之外的 action_id。
// 必要性：EnsureDefaultFeaturePackages 只追加不剪枝，seed 调整后的历史残留会污染
// 包定义；合并权限键时我们重新切分了平台管理员的包拓扑，必须剪枝才能得到干净结果。
// 仅对 seed 里显式列出 PermissionKeys 的 builtin 包生效；PermissionKeys 为空视为
// 不做限制，保留现有绑定。
func pruneBuiltinFeaturePackageKeys(logger *zap.Logger) error {
	packages := permissionseed.DefaultFeaturePackages()
	totalDeleted := int64(0)
	for _, seed := range packages {
		if !seed.IsBuiltin {
			continue
		}
		if len(seed.PermissionKeys) == 0 {
			continue
		}

		var pkg usermodel.FeaturePackage
		err := database.DB.Where("package_key = ? AND deleted_at IS NULL", seed.PackageKey).First(&pkg).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("lookup feature package %s: %w", seed.PackageKey, err)
		}

		var desiredKeys []usermodel.PermissionKey
		if err := database.DB.
			Where("permission_key IN ? AND deleted_at IS NULL", seed.PermissionKeys).
			Find(&desiredKeys).Error; err != nil {
			return fmt.Errorf("load desired permission keys for %s: %w", seed.PackageKey, err)
		}
		desiredIDs := make([]uuid.UUID, 0, len(desiredKeys))
		for _, k := range desiredKeys {
			desiredIDs = append(desiredIDs, k.ID)
		}
		if len(desiredIDs) == 0 {
			continue
		}

		result := database.DB.Exec(
			`DELETE FROM feature_package_keys WHERE package_id = ? AND action_id NOT IN ?`,
			pkg.ID, desiredIDs,
		)
		if result.Error != nil {
			return fmt.Errorf("prune feature_package_keys for %s: %w", seed.PackageKey, result.Error)
		}
		if result.RowsAffected > 0 {
			logger.Info("pruned feature_package_keys",
				zap.String("package_key", seed.PackageKey),
				zap.Int64("deleted", result.RowsAffected))
			totalDeleted += result.RowsAffected
		}
	}
	logger.Info("feature_package_keys prune applied", zap.Int64("deleted", totalDeleted))
	return nil
}

// backfillCustomMediaUploadOnManage 为历史自定义功能包补齐 system.media.upload 绑定。
// 背景：原先 system.media.manage 同时承担「上传直传链路」与「删除/治理」两种语义，
// 拆分为 system.media.upload（直传）+ system.media.manage（删除）后，builtin 包由
// EnsureDefaultFeaturePackages + pruneBuiltinFeaturePackageKeys 自动对齐，但非 builtin
// （IsBuiltin=false）包由租户手动维护，只有 manage 绑定时上传链路会 403。
// 此任务：凡是已绑定 system.media.manage 且未绑定 system.media.upload 的非 builtin 包，
// 自动追加 system.media.upload 绑定。幂等：已绑定者跳过（NOT EXISTS 保护）。
func backfillCustomMediaUploadOnManage(logger *zap.Logger) error {
	const manageKey = "system.media.manage"
	const uploadKey = "system.media.upload"

	var manage usermodel.PermissionKey
	if err := database.DB.Where("permission_key = ? AND deleted_at IS NULL", manageKey).First(&manage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("media manage key not present, skip backfill", zap.String("key", manageKey))
			return nil
		}
		return fmt.Errorf("lookup %s: %w", manageKey, err)
	}

	var upload usermodel.PermissionKey
	if err := database.DB.Where("permission_key = ? AND deleted_at IS NULL", uploadKey).First(&upload).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("media upload key not present, skip backfill", zap.String("key", uploadKey))
			return nil
		}
		return fmt.Errorf("lookup %s: %w", uploadKey, err)
	}

	result := database.DB.Exec(`
		INSERT INTO feature_package_keys (package_id, action_id, created_at)
		SELECT fpk.package_id, ?, now()
		  FROM feature_package_keys fpk
		  JOIN feature_packages fp ON fp.id = fpk.package_id AND fp.deleted_at IS NULL
		 WHERE fpk.action_id = ?
		   AND fp.is_builtin = false
		   AND NOT EXISTS (
		       SELECT 1 FROM feature_package_keys fpk2
		        WHERE fpk2.package_id = fpk.package_id
		          AND fpk2.action_id = ?
		   )
	`, upload.ID, manage.ID, upload.ID)
	if result.Error != nil {
		return fmt.Errorf("backfill media.upload for custom packages: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		logger.Info("backfilled system.media.upload on custom packages",
			zap.Int64("inserted", result.RowsAffected))
	} else {
		logger.Info("no custom packages need system.media.upload backfill")
	}
	return nil
}

// permissionKeyConsolidations 声明一次性合并规则：legacy_key -> canonical_key。
// 任何后续的权限键合并都集中在这里维护，避免散落在迁移代码各处。
func permissionKeyConsolidations() []struct{ From, To string } {
	return []struct{ From, To string }{
		{From: "observability.audit.read", To: "observability.log.read"},
		{From: "observability.telemetry.read", To: "observability.log.read"},
		{From: "system.register_log.read", To: "observability.log.read"},
		{From: "observability.policy.read", To: "observability.policy.manage"},
		{From: "observability.policy.write", To: "observability.policy.manage"},
		{From: "user.list", To: "system.user.manage"},
		{From: "user.read", To: "system.user.manage"},
		{From: "user.get", To: "system.user.manage"},
		{From: "user.create", To: "system.user.manage"},
		{From: "user.update", To: "system.user.manage"},
		{From: "user.delete", To: "system.user.manage"},
		{From: "user.assign_role", To: "system.user.manage"},
		{From: "system.user.assign_role", To: "system.user.manage"},
		{From: "system.role.assign_menu", To: "system.role.assign"},
		{From: "system.role.assign_action", To: "system.role.assign"},
		{From: "system.role.assign_data", To: "system.role.assign"},
		{From: "system.register_entry.read", To: "system.register_entry.manage"},
		{From: "system.register_entry.write", To: "system.register_entry.manage"},
	}
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

func initDefaultRolesNoScope(logger *zap.Logger) error {
	roles := []struct {
		Code        string
		Name        string
		Description string
		SortOrder   int
	}{
		{"admin", "管理员", "系统管理员，拥有所有权限", 1},
		{"collaboration_workspace_admin", "协作空间管理员", "协作空间管理员，可以管理协作空间成员和协作空间内容", 2},
		{"collaboration_workspace_member", "协作空间成员", "协作空间成员，可以查看和编辑协作空间内容", 3},
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
	permissionKey := strings.TrimSpace(spec.PermissionKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		definition = &systemmodels.MenuDefinition{
			AppKey:        appKey,
			MenuKey:       menuKey,
			Kind:          normalizeMenuSeedKind(spec),
			Path:          strings.TrimSpace(spec.Path),
			Name:          menuKey,
			Component:     strings.TrimSpace(spec.Component),
			PageKey:       "",
			PermissionKey: permissionKey,
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
			"kind":           normalizeMenuSeedKind(spec),
			"path":           strings.TrimSpace(spec.Path),
			"name":           menuKey,
			"component":      strings.TrimSpace(spec.Component),
			"permission_key": permissionKey,
			"default_title":  strings.TrimSpace(spec.Title),
			"default_icon":   strings.TrimSpace(spec.Icon),
			"status":         "normal",
			"meta":           meta,
			"updated_at":     time.Now(),
		}
		if err := database.DB.Model(definition).Updates(updates).Error; err != nil {
			return nil, err
		}
		definition.Kind = normalizeMenuSeedKind(spec)
		definition.Path = strings.TrimSpace(spec.Path)
		definition.Name = menuKey
		definition.Component = strings.TrimSpace(spec.Component)
		definition.PermissionKey = permissionKey
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
				WHEN COALESCE(NULLIF(TRIM(component), ''), '') <> '' THEN 'entry'
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
	case strings.TrimSpace(item.Component) != "":
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

	pageSeeds := permissionseed.DefaultPages()
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
	if parentMenuName := strings.TrimSpace(spec.ParentMenuName); parentMenuName != "" {
		var parentMenu systemmodels.MenuDefinition
		if err := database.DB.Where("app_key = ? AND name = ?", systemmodels.DefaultAppKey, parentMenuName).First(&parentMenu).Error; err != nil {
			return nil, err
		}
		parentMenuID = &parentMenu.ID
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
			TemplateKey:     "personal.notice.all_users",
			Name:            "个人空间全员公告",
			Description:     "个人空间向全部用户发送公告通知",
			MessageType:     "notice",
			OwnerScope:      "personal",
			AudienceType:    "all_users",
			TitleTemplate:   "{{title}}",
			SummaryTemplate: "{{summary}}",
			ContentTemplate: "{{content}}",
			ActionType:      "none",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"builtin": true},
		},
		{
			TemplateKey:     "personal.notice.collaboration_workspace_admins",
			Name:            "个人空间协作空间管理员提醒",
			Description:     "个人空间向协作空间管理员发送治理提醒",
			MessageType:     "message",
			OwnerScope:      "personal",
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
	created, err := permissionseed.EnsureOpenAPIPermissionKeys(database.DB)
	if err != nil {
		return err
	}
	logger.Info("OpenAPI-derived permission keys ensured", zap.Int("created", created))
	return nil
}

func initOpenAPIEndpoints(logger *zap.Logger) error {
	created, err := permissionseed.EnsureOpenAPIEndpoints(database.DB)
	if err != nil {
		return err
	}
	logger.Info("OpenAPI-derived api_endpoints ensured", zap.Int("created", created))

	bound, err := permissionseed.EnsureOpenAPIPermissionBindings(database.DB)
	if err != nil {
		return err
	}
	logger.Info("OpenAPI-derived permission bindings ensured", zap.Int("created", bound))

	// P2-3: 孤儿 permission_keys 扫描。permission_keys 表里没被任何 binding
	// 引用且非 builtin 的 key 通常是历史遗留或 spec 里被摘掉了的 op。只 warn
	// 不阻塞启动 —— 运维可以按这个列表做清理。
	orphans, err := permissionseed.ScanOrphanedPermissionKeys(database.DB)
	if err != nil {
		// 扫描失败不应挡住 migrate；记 warn 让运维看到即可。
		logger.Warn("orphan permission_keys scan failed", zap.Error(err))
	} else if len(orphans) > 0 {
		logger.Warn("orphan permission_keys (no api_endpoint_permission_bindings reference)",
			zap.Strings("keys", orphans),
			zap.Int("count", len(orphans)))
	} else {
		logger.Info("orphan permission_keys scan: clean")
	}
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
	// 构建一次完整 router 触发 mountOpenAPIBridgeRoutes 的挂载校验
	// （radix tree 冲突 / access_mode 非法等都会在这里 fail-fast），
	// 保证 CLI 阶段就暴露 spec 错误，而不是 HTTP 服务跑起来后才发现。
	// 显式传 Noop 跳过审计/遥测的异步 worker + channel。
	_ = apirouter.SetupRouter(cfg, logger, database.DB, audit.Noop{}, telemetry.Noop{})
	builder := permissionseed.NewDeploymentBuilder(database.DB, logger).WithCoreDefaults()
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
	case strings.TrimSpace(spec.Component) != "":
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
