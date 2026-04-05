package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// DB 数据库实例
var DB *gorm.DB

type RuntimeOptions struct {
	Env      string
	LogLevel string
}

// Init 初始化数据库连接
func Init(cfg *config.DBConfig, runtimeOptions ...RuntimeOptions) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
		cfg.TimeZone,
	)

	runtime := RuntimeOptions{}
	if len(runtimeOptions) > 0 {
		runtime = runtimeOptions[0]
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(resolveGormLogLevel(runtime)),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return DB, nil
}

func resolveGormLogLevel(runtime RuntimeOptions) gormlogger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(runtime.LogLevel)) {
	case "debug":
		return gormlogger.Info
	case "warn":
		return gormlogger.Warn
	case "error":
		return gormlogger.Error
	case "silent":
		return gormlogger.Silent
	}
	if strings.EqualFold(strings.TrimSpace(runtime.Env), "production") {
		return gormlogger.Warn
	}
	return gormlogger.Info
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 启用 PostgreSQL 扩展
	if err := enableExtensions(); err != nil {
		return fmt.Errorf("failed to enable extensions: %w", err)
	}

	// 执行数据库迁移
	err := DB.AutoMigrate(
		// 用户和角色相关
		&models.User{},
		&models.Role{},
		&models.App{},
		&models.AppHostBinding{},
		&models.UserRole{},
		&models.PermissionGroup{},
		&models.PermissionKey{},
		&models.FeaturePackage{},
		&models.FeaturePackageBundle{},
		&models.FeaturePackageKey{},
		&models.FeaturePackageMenu{},
		&models.TeamFeaturePackage{},
		&models.UserFeaturePackage{},
		&models.RoleFeaturePackage{},
		&models.RoleHiddenMenu{},
		&models.RoleDisabledAction{},
		&models.RoleDataPermission{},
		&models.TeamBlockedMenu{},
		&models.TeamBlockedAction{},
		&models.UserActionPermission{},
		&models.UserHiddenMenu{},
		&models.PlatformUserAccessSnapshot{},
		&models.PlatformRoleAccessSnapshot{},
		&models.TeamAccessSnapshot{},
		&models.TeamRoleAccessSnapshot{},
		&models.APIEndpointCategory{},
		&models.APIEndpoint{},
		&models.APIEndpointPermissionBinding{},
		&models.MenuSpace{},
		&models.MenuSpaceHostBinding{},
		&models.MenuManageGroup{},
		&models.MenuDefinition{},
		&models.SpaceMenuPlacement{},
		&models.Menu{},
		&models.UIPage{},
		&models.PageSpaceBinding{},
		&models.Tenant{},
		&models.TenantMember{},
		&models.APIKey{},
		&models.MediaAsset{},
		&models.MenuBackup{},
		&models.SystemSetting{},
		&models.MessageTemplate{},
		&models.Message{},
		&models.MessageDelivery{},
		&models.RiskOperationAudit{},
		&models.FeaturePackageVersion{},
		&models.PermissionBatchTemplate{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := ensureAppBootstrap(); err != nil {
		return fmt.Errorf("failed to bootstrap default app data: %w", err)
	}
	// 创建唯一索引（AutoMigrate 不会自动创建唯一索引）
	if err := createUniqueIndexes(); err != nil {
		return fmt.Errorf("failed to create unique indexes: %w", err)
	}

	return nil
}

// createUniqueIndexes 创建唯一索引
func createUniqueIndexes() error {
	// tenant_members 表的 (tenant_id, user_id) 唯一索引
	indexName := "idx_tenant_members_tenant_user_unique"
	var count int64
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", indexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + indexName + " ON tenant_members (tenant_id, user_id)").Error; err != nil {
			return err
		}
	}

	legacyIndexName := "idx_user_role"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", legacyIndexName).Scan(&count)
	if count > 0 {
		if err := DB.Exec("DROP INDEX IF EXISTS " + legacyIndexName).Error; err != nil {
			return err
		}
	}

	globalIndexName := "idx_user_roles_global_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", globalIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + globalIndexName + " ON user_roles (user_id, role_id) WHERE tenant_id IS NULL").Error; err != nil {
			return err
		}
	}

	tenantIndexName := "idx_user_roles_tenant_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", tenantIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + tenantIndexName + " ON user_roles (user_id, role_id, tenant_id) WHERE tenant_id IS NOT NULL").Error; err != nil {
			return err
		}
	}

	actionUniqueIndexName := "idx_permission_keys_permission_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", actionUniqueIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("DROP INDEX IF EXISTS idx_permission_actions_resource_action_unique").Error; err != nil {
			return err
		}
		if err := DB.Exec("DROP INDEX IF EXISTS idx_permission_actions_permission_key").Error; err != nil {
			return err
		}
		if err := DB.Exec("CREATE UNIQUE INDEX " + actionUniqueIndexName + " ON permission_keys (permission_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	userActionGlobalIndexName := "idx_user_action_permissions_global_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + userActionGlobalIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userActionGlobalIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userActionGlobalIndexName + " ON user_action_permissions (app_key, user_id, action_id) WHERE tenant_id IS NULL").Error; err != nil {
			return err
		}
	}

	userActionTenantIndexName := "idx_user_action_permissions_tenant_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + userActionTenantIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userActionTenantIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userActionTenantIndexName + " ON user_action_permissions (app_key, user_id, action_id, tenant_id) WHERE tenant_id IS NOT NULL").Error; err != nil {
			return err
		}
	}

	apiEndpointIndexName := "idx_api_endpoints_method_path_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", apiEndpointIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + apiEndpointIndexName + " ON api_endpoints (method, path)").Error; err != nil {
			return err
		}
	}

	apiEndpointCategoryCodeIndexName := "idx_api_endpoint_categories_code"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", apiEndpointCategoryCodeIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + apiEndpointCategoryCodeIndexName + " ON api_endpoint_categories (code) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	appKeyIndexName := "idx_apps_app_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", appKeyIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + appKeyIndexName + " ON apps (app_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	appHostIndexName := "idx_app_host_bindings_host"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", appHostIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + appHostIndexName + " ON app_host_bindings (host) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	appHostAppIndexName := "idx_app_host_bindings_app_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", appHostAppIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + appHostAppIndexName + " ON app_host_bindings (app_key)").Error; err != nil {
			return err
		}
	}

	menuSpaceCodeIndexName := "idx_menu_spaces_space_key"
	if err := DB.Exec("DROP INDEX IF EXISTS " + menuSpaceCodeIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuSpaceCodeIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + menuSpaceCodeIndexName + " ON menu_spaces (app_key, space_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	pageSpaceBindingIndexName := "idx_page_space_bindings_page_space_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + pageSpaceBindingIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", pageSpaceBindingIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + pageSpaceBindingIndexName + " ON page_space_bindings (app_key, page_id, space_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	menuSpaceHostIndexName := "idx_menu_space_host_bindings_host"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuSpaceHostIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + menuSpaceHostIndexName + " ON menu_space_host_bindings (host)").Error; err != nil {
			return err
		}
	}

	menuSpaceBindingSpaceIndexName := "idx_menu_space_host_bindings_space_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuSpaceBindingSpaceIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + menuSpaceBindingSpaceIndexName + " ON menu_space_host_bindings (space_key)").Error; err != nil {
			return err
		}
	}

	menuSpaceTableIndexName := "idx_menus_space_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuSpaceTableIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + menuSpaceTableIndexName + " ON menus (app_key, space_key)").Error; err != nil {
			return err
		}
	}

	menuDefinitionIndexName := "idx_menu_definitions_key_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + menuDefinitionIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuDefinitionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + menuDefinitionIndexName + " ON menu_definitions (app_key, menu_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	menuDefinitionLegacyNameIndex := "idx_menu_definitions_app_name"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuDefinitionLegacyNameIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + menuDefinitionLegacyNameIndex + " ON menu_definitions (app_key, name)").Error; err != nil {
			return err
		}
	}

	spaceMenuPlacementIndexName := "idx_space_menu_placements_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + spaceMenuPlacementIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", spaceMenuPlacementIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + spaceMenuPlacementIndexName + " ON space_menu_placements (app_key, space_key, menu_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	spaceMenuPlacementParentIndexName := "idx_space_menu_placements_parent"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", spaceMenuPlacementParentIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + spaceMenuPlacementParentIndexName + " ON space_menu_placements (app_key, space_key, parent_menu_key, sort_order)").Error; err != nil {
			return err
		}
	}

	uiPageSpaceIndexName := "idx_ui_pages_space_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageSpaceIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageSpaceIndexName + " ON ui_pages (space_key)").Error; err != nil {
			return err
		}
	}

	apiEndpointPermissionBindingUnique := "idx_api_endpoint_permission_bindings_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + apiEndpointPermissionBindingUnique).Error; err != nil {
		return err
	}
	if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS " + apiEndpointPermissionBindingUnique + " ON api_endpoint_permission_bindings (endpoint_code, permission_key) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}

	permissionActionHotIndexName := "idx_permission_keys_status_sort_created"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", permissionActionHotIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("DROP INDEX IF EXISTS idx_permission_actions_status_sort_created").Error; err != nil {
			return err
		}
		if err := DB.Exec("CREATE INDEX " + permissionActionHotIndexName + " ON permission_keys (status, sort_order, created_at DESC)").Error; err != nil {
			return err
		}
	}

	permissionGroupIndexName := "idx_permission_groups_type_code"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", permissionGroupIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + permissionGroupIndexName + " ON permission_groups (group_type, code) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	featurePackageIndexName := "idx_feature_packages_key_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + featurePackageIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + featurePackageIndexName + " ON feature_packages (app_key, package_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	featurePackageActionIndexName := "idx_feature_package_keys_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("DROP INDEX IF EXISTS idx_feature_package_actions_unique").Error; err != nil {
			return err
		}
		if err := DB.Exec("CREATE UNIQUE INDEX " + featurePackageActionIndexName + " ON feature_package_keys (package_id, action_id)").Error; err != nil {
			return err
		}
	}

	featurePackageMenuIndexName := "idx_feature_package_menus_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + featurePackageMenuIndexName + " ON feature_package_menus (package_id, menu_id)").Error; err != nil {
			return err
		}
	}

	featurePackageBundleIndexName := "idx_feature_package_bundles_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageBundleIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + featurePackageBundleIndexName + " ON feature_package_bundles (package_id, child_package_id)").Error; err != nil {
			return err
		}
	}

	teamFeaturePackageIndexName := "idx_team_feature_packages_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + teamFeaturePackageIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", teamFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + teamFeaturePackageIndexName + " ON team_feature_packages (app_key, team_id, package_id)").Error; err != nil {
			return err
		}
	}

	userFeaturePackageIndexName := "idx_user_feature_packages_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + userFeaturePackageIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userFeaturePackageIndexName + " ON user_feature_packages (app_key, user_id, package_id)").Error; err != nil {
			return err
		}
	}

	roleFeaturePackageIndexName := "idx_role_feature_packages_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + roleFeaturePackageIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", roleFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + roleFeaturePackageIndexName + " ON role_feature_packages (app_key, role_id, package_id)").Error; err != nil {
			return err
		}
	}

	roleHiddenMenuIndexName := "idx_role_hidden_menus_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + roleHiddenMenuIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", roleHiddenMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + roleHiddenMenuIndexName + " ON role_hidden_menus (app_key, role_id, menu_id)").Error; err != nil {
			return err
		}
	}

	roleDisabledActionIndexName := "idx_role_disabled_actions_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + roleDisabledActionIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", roleDisabledActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + roleDisabledActionIndexName + " ON role_disabled_actions (app_key, role_id, action_id)").Error; err != nil {
			return err
		}
	}

	teamBlockedMenuIndexName := "idx_team_blocked_menus_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + teamBlockedMenuIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", teamBlockedMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + teamBlockedMenuIndexName + " ON team_blocked_menus (app_key, team_id, menu_id)").Error; err != nil {
			return err
		}
	}

	teamBlockedActionIndexName := "idx_team_blocked_actions_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + teamBlockedActionIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", teamBlockedActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + teamBlockedActionIndexName + " ON team_blocked_actions (app_key, team_id, action_id)").Error; err != nil {
			return err
		}
	}

	userHiddenMenuIndexName := "idx_user_hidden_menus_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + userHiddenMenuIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userHiddenMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userHiddenMenuIndexName + " ON user_hidden_menus (app_key, user_id, menu_id)").Error; err != nil {
			return err
		}
	}

	menuManageGroupNameIndex := "idx_menu_manage_groups_name_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuManageGroupNameIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + menuManageGroupNameIndex + " ON menu_manage_groups (name) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	menuManageGroupSortIndex := "idx_menu_manage_groups_sort_status"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuManageGroupSortIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + menuManageGroupSortIndex + " ON menu_manage_groups (sort_order, status)").Error; err != nil {
			return err
		}
	}

	menuManageGroupIDIndex := "idx_menus_manage_group_id"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", menuManageGroupIDIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + menuManageGroupIDIndex + " ON menus (manage_group_id)").Error; err != nil {
			return err
		}
	}

	uiPageKeyIndexName := "idx_ui_pages_page_key_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + uiPageKeyIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageKeyIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + uiPageKeyIndexName + " ON ui_pages (app_key, page_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	uiPageRouteNameIndexName := "idx_ui_pages_route_name_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + uiPageRouteNameIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageRouteNameIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + uiPageRouteNameIndexName + " ON ui_pages (app_key, route_name) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	uiPageParentMenuIndexName := "idx_ui_pages_parent_menu_id"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageParentMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageParentMenuIndexName + " ON ui_pages (parent_menu_id)").Error; err != nil {
			return err
		}
	}

	uiPageModuleIndexName := "idx_ui_pages_module_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageModuleIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageModuleIndexName + " ON ui_pages (module_key)").Error; err != nil {
			return err
		}
	}

	uiPageParentPageIndexName := "idx_ui_pages_parent_page_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageParentPageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageParentPageIndexName + " ON ui_pages (parent_page_key)").Error; err != nil {
			return err
		}
	}

	uiPageDisplayGroupIndexName := "idx_ui_pages_display_group_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageDisplayGroupIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageDisplayGroupIndexName + " ON ui_pages (display_group_key)").Error; err != nil {
			return err
		}
	}

	uiPageTypeStatusIndexName := "idx_ui_pages_page_type_status"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageTypeStatusIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageTypeStatusIndexName + " ON ui_pages (page_type, status)").Error; err != nil {
			return err
		}
	}

	uiPageAccessModeIndexName := "idx_ui_pages_access_mode"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageAccessModeIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageAccessModeIndexName + " ON ui_pages (access_mode)").Error; err != nil {
			return err
		}
	}

	uiPageParentSortIndexName := "idx_ui_pages_parent_sort_order"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", uiPageParentSortIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + uiPageParentSortIndexName + " ON ui_pages (parent_page_key, sort_order)").Error; err != nil {
			return err
		}
	}

	systemSettingKeyIndexName := "idx_system_settings_key_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", systemSettingKeyIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + systemSettingKeyIndexName + " ON system_settings (key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	messageTemplateKeyIndexName := "idx_message_templates_template_key_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", messageTemplateKeyIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + messageTemplateKeyIndexName + " ON message_templates (template_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	messageDeliveryRecipientIndexName := "idx_message_deliveries_message_recipient_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", messageDeliveryRecipientIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + messageDeliveryRecipientIndexName + " ON message_deliveries (message_id, recipient_user_id) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	riskAuditObjectIndexName := "idx_risk_operation_audits_object"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", riskAuditObjectIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + riskAuditObjectIndexName + " ON risk_operation_audits (object_type, object_id, created_at DESC)").Error; err != nil {
			return err
		}
	}

	featurePackageVersionUniqueIndex := "idx_feature_package_versions_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageVersionUniqueIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + featurePackageVersionUniqueIndex + " ON feature_package_versions (package_id, version_no) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	featurePackageVersionTimeIndex := "idx_feature_package_versions_created"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageVersionTimeIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + featurePackageVersionTimeIndex + " ON feature_package_versions (package_id, created_at DESC)").Error; err != nil {
			return err
		}
	}

	if err := ensureAppScopedSnapshotPrimaryKeys(); err != nil {
		return err
	}

	return nil
}

func ensureAppBootstrap() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	defaultApp := models.App{
		AppKey:          models.DefaultAppKey,
		Name:            models.DefaultAppName,
		Description:     "当前内置管理员后台应用",
		DefaultSpaceKey: models.DefaultMenuSpaceKey,
		AuthMode:        "inherit_host",
		Status:          "normal",
		IsDefault:       true,
		Meta:            models.MetaJSON{},
	}

	var existing models.App
	err := DB.Where("app_key = ? AND deleted_at IS NULL", models.DefaultAppKey).First(&existing).Error
	switch {
	case err == nil:
		if updateErr := DB.Model(&existing).Updates(map[string]interface{}{
			"name":              defaultApp.Name,
			"description":       defaultApp.Description,
			"default_space_key": defaultApp.DefaultSpaceKey,
			"auth_mode":         defaultApp.AuthMode,
			"status":            "normal",
			"is_default":        true,
		}).Error; updateErr != nil {
			return updateErr
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		if createErr := DB.Create(&defaultApp).Error; createErr != nil {
			return createErr
		}
	default:
		return err
	}

	if err := backfillDefaultAppKey(); err != nil {
		return err
	}
	if err := backfillAppScopedRelations(); err != nil {
		return err
	}
	if err := backfillMenuDefinitionsAndPlacements(); err != nil {
		return err
	}

	return nil
}

func backfillDefaultAppKey() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	statements := []string{
		"UPDATE menu_spaces SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE menus SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE menu_definitions SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE space_menu_placements SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE ui_pages SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE page_space_bindings SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE feature_packages SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE menu_backups SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE api_endpoints SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE platform_user_access_snapshots SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE platform_role_access_snapshots SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE team_access_snapshots SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE team_role_access_snapshots SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE api_endpoints SET app_scope = '" + models.AppScopeShared + "' WHERE COALESCE(TRIM(app_scope), '') = '' AND (path LIKE '/api/v1/auth/%' OR path = '/api/v1/pages/runtime/public' OR path LIKE '/open/v1/%' OR path = '/health')",
		"UPDATE api_endpoints SET app_scope = '" + models.AppScopeApp + "' WHERE COALESCE(TRIM(app_scope), '') = ''",
	}
	for _, statement := range statements {
		if err := DB.Exec(statement).Error; err != nil {
			return err
		}
	}

	if err := DB.Model(&models.MenuSpace{}).
		Where("space_key = ? AND deleted_at IS NULL", models.DefaultMenuSpaceKey).
		Updates(map[string]interface{}{
			"app_key":    models.DefaultAppKey,
			"is_default": true,
			"status":     "normal",
		}).Error; err != nil {
		return err
	}

	return nil
}

func backfillAppScopedRelations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	statements := []string{
		"UPDATE role_feature_packages SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE team_feature_packages SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE user_feature_packages SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE role_hidden_menus SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE team_blocked_menus SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE user_hidden_menus SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE role_disabled_actions SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE team_blocked_actions SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		"UPDATE user_action_permissions SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
	}
	for _, statement := range statements {
		if err := DB.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func backfillMenuDefinitionsAndPlacements() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	type legacyMenu struct {
		ID            string
		AppKey        string
		SpaceKey      string
		ParentID      *string
		ManageGroupID *string
		Kind          string
		Path          string
		Name          string
		Component     string
		Title         string
		Icon          string
		SortOrder     int
		Hidden        bool
		Meta          models.MetaJSON
	}

	menus := make([]legacyMenu, 0)
	if err := DB.Raw(`
		SELECT
			id::text AS id,
			app_key,
			COALESCE(NULLIF(space_key, ''), ?) AS space_key,
			CASE WHEN parent_id IS NULL THEN NULL ELSE parent_id::text END AS parent_id,
			CASE WHEN manage_group_id IS NULL THEN NULL ELSE manage_group_id::text END AS manage_group_id,
			kind,
			path,
			name,
			component,
			title,
			icon,
			sort_order,
			hidden,
			meta
		FROM menus
		WHERE deleted_at IS NULL
		ORDER BY app_key ASC, sort_order ASC, created_at ASC
	`, models.DefaultMenuSpaceKey).Scan(&menus).Error; err != nil {
		return err
	}
	if len(menus) == 0 {
		return nil
	}

	type menuKeyEntry struct {
		ID      string
		AppKey  string
		MenuKey string
	}

	keyEntries := make([]menuKeyEntry, 0, len(menus))
	keyByLegacyID := make(map[string]string, len(menus))
	usedKeys := make(map[string]int)
	for _, item := range menus {
		appKey := strings.TrimSpace(item.AppKey)
		if appKey == "" {
			appKey = models.DefaultAppKey
		}
		baseKey := strings.TrimSpace(item.Name)
		if baseKey == "" {
			baseKey = "legacy-" + strings.ToLower(strings.ReplaceAll(item.ID, "-", ""))
		}
		baseKey = strings.ToLower(strings.TrimSpace(baseKey))
		if baseKey == "" {
			baseKey = "legacy-" + strings.ToLower(strings.ReplaceAll(item.ID, "-", ""))
		}
		composite := appKey + "::" + baseKey
		if seen, ok := usedKeys[composite]; ok {
			seen++
			usedKeys[composite] = seen
			baseKey = fmt.Sprintf("%s-%d", baseKey, seen)
		} else {
			usedKeys[composite] = 1
		}
		keyByLegacyID[item.ID] = baseKey
		keyEntries = append(keyEntries, menuKeyEntry{
			ID:      item.ID,
			AppKey:  appKey,
			MenuKey: baseKey,
		})
	}

	for _, item := range menus {
		appKey := strings.TrimSpace(item.AppKey)
		if appKey == "" {
			appKey = models.DefaultAppKey
		}
		menuKey := keyByLegacyID[item.ID]
		if menuKey == "" {
			continue
		}
		definition := models.MenuDefinition{
			ID:            uuid.MustParse(item.ID),
			AppKey:        appKey,
			MenuKey:       menuKey,
			Kind:          item.Kind,
			Path:          item.Path,
			Name:          item.Name,
			Component:     item.Component,
			DefaultTitle:  item.Title,
			DefaultIcon:   item.Icon,
			Status:        "normal",
			Meta:          item.Meta,
		}
		if err := DB.
			Where("app_key = ? AND menu_key = ? AND deleted_at IS NULL", definition.AppKey, definition.MenuKey).
			Assign(map[string]interface{}{
				"path":          definition.Path,
				"name":          definition.Name,
				"kind":          definition.Kind,
				"component":     definition.Component,
				"default_title": definition.DefaultTitle,
				"default_icon":  definition.DefaultIcon,
				"status":        definition.Status,
				"meta":          definition.Meta,
			}).
			FirstOrCreate(&definition).Error; err != nil {
			return err
		}
	}

	type parentRow struct {
		ID           string
		ParentMenuID *string
	}
	parentRows := make([]parentRow, 0, len(menus))
	for _, item := range menus {
		parentRows = append(parentRows, parentRow{
			ID:           item.ID,
			ParentMenuID: item.ParentID,
		})
	}

	for _, item := range menus {
		appKey := strings.TrimSpace(item.AppKey)
		if appKey == "" {
			appKey = models.DefaultAppKey
		}
		spaceKey := strings.TrimSpace(item.SpaceKey)
		if spaceKey == "" {
			spaceKey = models.DefaultMenuSpaceKey
		}
		parentMenuKey := ""
		if item.ParentID != nil {
			parentMenuKey = keyByLegacyID[strings.TrimSpace(*item.ParentID)]
		}
		placement := models.SpaceMenuPlacement{
			AppKey:        appKey,
			SpaceKey:      spaceKey,
			MenuKey:       keyByLegacyID[item.ID],
			ParentMenuKey: parentMenuKey,
			SortOrder:     item.SortOrder,
			Hidden:        item.Hidden,
			MetaOverride:  models.MetaJSON{},
		}
		if item.ManageGroupID != nil {
			parsed, err := uuid.Parse(strings.TrimSpace(*item.ManageGroupID))
			if err == nil {
				placement.ManageGroupID = &parsed
			}
		}
		if err := DB.
			Where("app_key = ? AND space_key = ? AND menu_key = ? AND deleted_at IS NULL", placement.AppKey, placement.SpaceKey, placement.MenuKey).
			Assign(map[string]interface{}{
				"parent_menu_key": placement.ParentMenuKey,
				"manage_group_id": placement.ManageGroupID,
				"sort_order":      placement.SortOrder,
				"hidden":          placement.Hidden,
				"title_override":  placement.TitleOverride,
				"icon_override":   placement.IconOverride,
				"meta_override":   placement.MetaOverride,
			}).
			FirstOrCreate(&placement).Error; err != nil {
			return err
		}
	}

	return nil
}

func ensureAppScopedSnapshotPrimaryKeys() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	type snapshotSpec struct {
		table string
		cols  []string
	}

	specs := []snapshotSpec{
		{table: "platform_user_access_snapshots", cols: []string{"app_key", "user_id"}},
		{table: "platform_role_access_snapshots", cols: []string{"app_key", "role_id"}},
		{table: "team_access_snapshots", cols: []string{"app_key", "team_id"}},
		{table: "team_role_access_snapshots", cols: []string{"app_key", "team_id", "role_id"}},
	}

	for _, spec := range specs {
		if err := DB.Exec(
			"ALTER TABLE " + spec.table + " ADD COLUMN IF NOT EXISTS app_key varchar(100) NOT NULL DEFAULT '" + models.DefaultAppKey + "'",
		).Error; err != nil {
			return err
		}
		if err := DB.Exec(
			"UPDATE " + spec.table + " SET app_key = '" + models.DefaultAppKey + "' WHERE COALESCE(TRIM(app_key), '') = ''",
		).Error; err != nil {
			return err
		}
		if err := DB.Exec("ALTER TABLE " + spec.table + " DROP CONSTRAINT IF EXISTS " + spec.table + "_pkey").Error; err != nil {
			return err
		}
		if err := DB.Exec(
			"ALTER TABLE " + spec.table + " ADD PRIMARY KEY (" + strings.Join(spec.cols, ", ") + ")",
		).Error; err != nil {
			return err
		}
	}

	return nil
}

// enableExtensions 启用 PostgreSQL 扩展
func enableExtensions() error {
	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
		"CREATE EXTENSION IF NOT EXISTS \"ltree\"",
		"CREATE EXTENSION IF NOT EXISTS \"pg_trgm\"",
	}

	for _, ext := range extensions {
		if err := DB.Exec(ext).Error; err != nil {
			return fmt.Errorf("failed to create extension: %w", err)
		}
	}

	return nil
}
