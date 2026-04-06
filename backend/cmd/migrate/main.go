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

	logger.Info("Using final workspace schema and canonical permission seeds")

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

	if err := syncAPIRegistry(logger, cfg); err != nil {
		logger.Warn("Failed to sync API registry", zap.Error(err))
	} else {
		logger.Info("API registry synchronized successfully")
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
	packageKeys := []string{"personal.system_admin", "personal.menu_admin"}
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
