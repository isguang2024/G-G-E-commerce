package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
		&models.RoleAppScope{},
		&models.App{},
		&models.AppHostBinding{},
		&models.AuthCallbackCode{},
		&models.UserRole{},
		&models.PermissionGroup{},
		&models.PermissionKey{},
		&models.FeaturePackage{},
		&models.FeaturePackageBundle{},
		&models.FeaturePackageKey{},
		&models.FeaturePackageMenu{},
		&models.CollaborationWorkspaceFeaturePackage{},
		&models.UserFeaturePackage{},
		&models.RoleFeaturePackage{},
		&models.RoleHiddenMenu{},
		&models.RoleDisabledAction{},
		&models.RoleDataPermission{},
		&models.CollaborationWorkspaceBlockedMenu{},
		&models.CollaborationWorkspaceBlockedAction{},
		&models.UserActionPermission{},
		&models.UserHiddenMenu{},
		&models.PersonalWorkspaceAccessSnapshot{},
		&models.PersonalWorkspaceRoleAccessSnapshot{},
		&models.CollaborationWorkspaceAccessSnapshot{},
		&models.CollaborationWorkspaceRoleAccessSnapshot{},
		&models.APIEndpointCategory{},
		&models.APIEndpoint{},
		&models.APIEndpointPermissionBinding{},
		&models.MenuSpace{},
		&models.MenuSpaceHostBinding{},
		&models.MenuSpaceEntryBinding{},
		&models.MenuDefinition{},
		&models.SpaceMenuPlacement{},
		&models.Menu{},
		&models.UIPage{},
		&models.PageSpaceBinding{},
		&models.CollaborationWorkspace{},
		&models.CollaborationWorkspaceMember{},
		&models.Workspace{},
		&models.WorkspaceMember{},
		&models.WorkspaceRoleBinding{},
		&models.WorkspaceFeaturePackage{},
		&models.APIKey{},
		&models.MediaAsset{},
		&models.SystemSetting{},
		&models.MessageTemplate{},
		&models.MessageSender{},
		&models.MessageRecipientGroup{},
		&models.MessageRecipientGroupTarget{},
		&models.Message{},
		&models.MessageDelivery{},
		&models.RiskOperationAudit{},
		&models.FeaturePackageVersion{},
		&models.PermissionBatchTemplate{},
		// 注册体系（v5.x register slice）
		&models.RegisterEntry{},
		&models.SocialAuthProvider{},
		&models.UserSocialAccount{},
		&models.SocialOAuthState{},
		// 数据字典
		&models.DictType{},
		&models.DictItem{},
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
	// collaboration_workspace_members 表的 (collaboration_workspace_id, user_id) 唯一索引
	indexName := "idx_collaboration_workspace_members_user_unique"
	var count int64
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", indexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + indexName + " ON collaboration_workspace_members (collaboration_workspace_id, user_id)").Error; err != nil {
			return err
		}
	}

	globalIndexName := "idx_user_roles_global_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", globalIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + globalIndexName + " ON user_roles (user_id, role_id) WHERE collaboration_workspace_id IS NULL").Error; err != nil {
			return err
		}
	}

	collaborationWorkspaceIndexName := "idx_user_roles_collaboration_workspace_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", collaborationWorkspaceIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + collaborationWorkspaceIndexName + " ON user_roles (user_id, role_id, collaboration_workspace_id) WHERE collaboration_workspace_id IS NOT NULL").Error; err != nil {
			return err
		}
	}

	roleAppScopeIndexName := "idx_role_app_scopes_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", roleAppScopeIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + roleAppScopeIndexName + " ON role_app_scopes (role_id, app_key)").Error; err != nil {
			return err
		}
	}

	workspaceCodeIndexName := "idx_workspaces_code_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", workspaceCodeIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + workspaceCodeIndexName + " ON workspaces (code) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	personalWorkspaceOwnerIndexName := "idx_workspaces_personal_owner_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", personalWorkspaceOwnerIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + personalWorkspaceOwnerIndexName + " ON workspaces (owner_user_id) WHERE workspace_type = 'personal' AND owner_user_id IS NOT NULL AND deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	collaborationWorkspaceIndexNameOnWorkspace := "idx_workspaces_collaboration_workspace_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", collaborationWorkspaceIndexNameOnWorkspace).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + collaborationWorkspaceIndexNameOnWorkspace + " ON workspaces (collaboration_workspace_id) WHERE workspace_type = 'collaboration' AND collaboration_workspace_id IS NOT NULL AND deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	workspaceMemberIndexName := "idx_workspace_members_workspace_user_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", workspaceMemberIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + workspaceMemberIndexName + " ON workspace_members (workspace_id, user_id) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	workspaceRoleBindingIndexName := "idx_workspace_role_bindings_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", workspaceRoleBindingIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + workspaceRoleBindingIndexName + " ON workspace_role_bindings (workspace_id, user_id, role_id) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	workspaceFeaturePackageIndexName := "idx_workspace_feature_packages_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", workspaceFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + workspaceFeaturePackageIndexName + " ON workspace_feature_packages (workspace_id, package_id) WHERE deleted_at IS NULL").Error; err != nil {
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
		if err := DB.Exec("CREATE UNIQUE INDEX " + userActionGlobalIndexName + " ON user_action_permissions (app_key, user_id, action_id) WHERE collaboration_workspace_id IS NULL").Error; err != nil {
			return err
		}
	}

	userActionCollaborationWorkspaceIndexName := "idx_user_action_permissions_collaboration_workspace_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + userActionCollaborationWorkspaceIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userActionCollaborationWorkspaceIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userActionCollaborationWorkspaceIndexName + " ON user_action_permissions (app_key, user_id, action_id, collaboration_workspace_id) WHERE collaboration_workspace_id IS NOT NULL").Error; err != nil {
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

	collaborationWorkspaceFeaturePackageIndexName := "idx_collaboration_workspace_feature_packages_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + collaborationWorkspaceFeaturePackageIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", collaborationWorkspaceFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + collaborationWorkspaceFeaturePackageIndexName + " ON collaboration_workspace_feature_packages (app_key, collaboration_workspace_id, package_id)").Error; err != nil {
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

	collaborationWorkspaceBlockedMenuIndexName := "idx_collaboration_workspace_blocked_menus_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + collaborationWorkspaceBlockedMenuIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", collaborationWorkspaceBlockedMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + collaborationWorkspaceBlockedMenuIndexName + " ON collaboration_workspace_blocked_menus (app_key, collaboration_workspace_id, menu_id)").Error; err != nil {
			return err
		}
	}

	collaborationWorkspaceBlockedActionIndexName := "idx_collaboration_workspace_blocked_actions_unique"
	if err := DB.Exec("DROP INDEX IF EXISTS " + collaborationWorkspaceBlockedActionIndexName).Error; err != nil {
		return err
	}
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", collaborationWorkspaceBlockedActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + collaborationWorkspaceBlockedActionIndexName + " ON collaboration_workspace_blocked_actions (app_key, collaboration_workspace_id, action_id)").Error; err != nil {
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

	dictTypeCodeIndex := "idx_dict_types_tenant_code_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", dictTypeCodeIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + dictTypeCodeIndex + " ON dict_types (tenant_id, code) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	dictItemValueIndex := "idx_dict_items_tenant_type_value_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", dictItemValueIndex).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + dictItemValueIndex + " ON dict_items (tenant_id, dict_type_id, value) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	return nil
}

func ensureAppBootstrap() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	defaultApp := models.App{
		AppKey:           models.DefaultAppKey,
		Name:             models.DefaultAppName,
		Description:      "当前内置管理员后台应用",
		SpaceMode:        "single",
		DefaultSpaceKey:  models.DefaultMenuSpaceKey,
		AuthMode:         "inherit_host",
		FrontendEntryURL: "/",
		BackendEntryURL:  "",
		HealthCheckURL:   "/health",
		Status:           "normal",
		IsDefault:        true,
		Capabilities:     models.DefaultPlatformAdminCapabilities(),
		Meta:             models.MetaJSON{},
	}

	var existing models.App
	err := DB.Where("app_key = ? AND deleted_at IS NULL", models.DefaultAppKey).First(&existing).Error
	switch {
	case err == nil:
		if updateErr := DB.Model(&existing).Updates(map[string]interface{}{
			"name":               defaultApp.Name,
			"description":        defaultApp.Description,
			"space_mode":         defaultApp.SpaceMode,
			"default_space_key":  defaultApp.DefaultSpaceKey,
			"auth_mode":          defaultApp.AuthMode,
			"frontend_entry_url": defaultApp.FrontendEntryURL,
			"backend_entry_url":  defaultApp.BackendEntryURL,
			"health_check_url":   defaultApp.HealthCheckURL,
			"capabilities":       defaultApp.Capabilities,
			"status":             "normal",
			"is_default":         true,
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

	if err := ensureLocalEntryBindings(); err != nil {
		return fmt.Errorf("failed to seed local entry bindings: %w", err)
	}

	return nil
}

// ensureLocalEntryBindings 为本地开发环境写入默认 APP 入口解析绑定。
// 让 localhost / 127.0.0.1 等本机访问直接命中默认管理后台 APP。
func ensureLocalEntryBindings() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	seeds := []models.AppHostBinding{
		{
			AppKey:          models.DefaultAppKey,
			MatchType:       models.EntryMatchHostSuffix,
			Host:            "localhost",
			Description:     "本地开发：localhost 及子域默认进入后台管理",
			DefaultSpaceKey: models.DefaultMenuSpaceKey,
			Status:          "normal",
			IsPrimary:       true,
			Priority:        100,
			Meta:            models.MetaJSON{},
		},
		{
			AppKey:          models.DefaultAppKey,
			MatchType:       models.EntryMatchHostExact,
			Host:            "127.0.0.1",
			Description:     "本地开发：127.0.0.1 默认进入后台管理",
			DefaultSpaceKey: models.DefaultMenuSpaceKey,
			Status:          "normal",
			Priority:        100,
			Meta:            models.MetaJSON{},
		},
	}
	for _, seed := range seeds {
		var existing models.AppHostBinding
		err := DB.Where("match_type = ? AND host = ? AND path_pattern = ?", seed.MatchType, seed.Host, seed.PathPattern).First(&existing).Error
		switch {
		case err == nil:
			continue
		case errors.Is(err, gorm.ErrRecordNotFound):
			if createErr := DB.Create(&seed).Error; createErr != nil {
				return createErr
			}
		default:
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

func ensureAPIEndpointAppColumns(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	statements := []string{
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS app_scope varchar(20) NOT NULL DEFAULT 'shared'`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS app_key varchar(100) NOT NULL DEFAULT '` + models.DefaultAppKey + `'`,
		`UPDATE api_endpoints SET app_key = '` + models.DefaultAppKey + `' WHERE COALESCE(TRIM(app_key), '') = ''`,
		`UPDATE api_endpoints SET app_scope = '` + models.AppScopeShared + `' WHERE COALESCE(TRIM(app_scope), '') = '' AND (path LIKE '/api/v1/auth/%' OR path = '/api/v1/pages/runtime/public' OR path LIKE '/open/v1/%' OR path = '/health')`,
		`UPDATE api_endpoints SET app_scope = '` + models.AppScopeApp + `' WHERE COALESCE(TRIM(app_scope), '') = ''`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
