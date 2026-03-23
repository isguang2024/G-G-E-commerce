package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// DB 数据库实例
var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DBConfig) (*gorm.DB, error) {
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

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return DB, nil
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
		&models.UserRole{},
		&models.RoleMenu{},
		&models.PermissionAction{},
		&models.FeaturePackage{},
		&models.FeaturePackageBundle{},
		&models.FeaturePackageAction{},
		&models.FeaturePackageMenu{},
		&models.TeamFeaturePackage{},
		&models.UserFeaturePackage{},
		&models.RoleFeaturePackage{},
		&models.RoleHiddenMenu{},
		&models.RoleActionPermission{},
		&models.RoleDisabledAction{},
		&models.RoleDataPermission{},
		&models.TenantActionPermission{},
		&models.TeamManualActionPermission{},
		&models.TeamBlockedMenu{},
		&models.TeamBlockedAction{},
		&models.UserActionPermission{},
		&models.UserHiddenMenu{},
		&models.PlatformUserAccessSnapshot{},
		&models.PlatformRoleAccessSnapshot{},
		&models.TeamAccessSnapshot{},
		&models.TeamRoleAccessSnapshot{},
		&models.APIEndpoint{},
		&models.Menu{},
		&models.Tenant{},
		&models.TenantMember{},
		&models.APIKey{},
		&models.MediaAsset{},
		&models.MenuBackup{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 创建唯一索引（AutoMigrate 不会自动创建唯一索引）
	if err := createUniqueIndexes(); err != nil {
		return fmt.Errorf("failed to create unique indexes: %w", err)
	}

	if err := migrateLegacyUserRoles(); err != nil {
		return fmt.Errorf("failed to migrate legacy user roles: %w", err)
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

	actionUniqueIndexName := "idx_permission_actions_permission_key"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", actionUniqueIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("DROP INDEX IF EXISTS idx_permission_actions_resource_action_unique").Error; err != nil {
			return err
		}
		if err := DB.Exec("CREATE UNIQUE INDEX " + actionUniqueIndexName + " ON permission_actions (permission_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	userActionGlobalIndexName := "idx_user_action_permissions_global_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userActionGlobalIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userActionGlobalIndexName + " ON user_action_permissions (user_id, action_id) WHERE tenant_id IS NULL").Error; err != nil {
			return err
		}
	}

	userActionTenantIndexName := "idx_user_action_permissions_tenant_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userActionTenantIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userActionTenantIndexName + " ON user_action_permissions (user_id, action_id, tenant_id) WHERE tenant_id IS NOT NULL").Error; err != nil {
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

	permissionActionHotIndexName := "idx_permission_actions_status_sort_created"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", permissionActionHotIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE INDEX " + permissionActionHotIndexName + " ON permission_actions (status, sort_order, created_at DESC)").Error; err != nil {
			return err
		}
	}

	featurePackageIndexName := "idx_feature_packages_key_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + featurePackageIndexName + " ON feature_packages (package_key) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}

	featurePackageActionIndexName := "idx_feature_package_actions_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", featurePackageActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + featurePackageActionIndexName + " ON feature_package_actions (package_id, action_id)").Error; err != nil {
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

	teamManualActionIndexName := "idx_team_manual_action_permissions_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", teamManualActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + teamManualActionIndexName + " ON team_manual_action_permissions (tenant_id, action_id)").Error; err != nil {
			return err
		}
	}

	teamFeaturePackageIndexName := "idx_team_feature_packages_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", teamFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + teamFeaturePackageIndexName + " ON team_feature_packages (team_id, package_id)").Error; err != nil {
			return err
		}
	}

	userFeaturePackageIndexName := "idx_user_feature_packages_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userFeaturePackageIndexName + " ON user_feature_packages (user_id, package_id)").Error; err != nil {
			return err
		}
	}

	roleFeaturePackageIndexName := "idx_role_feature_packages_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", roleFeaturePackageIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + roleFeaturePackageIndexName + " ON role_feature_packages (role_id, package_id)").Error; err != nil {
			return err
		}
	}

	roleHiddenMenuIndexName := "idx_role_hidden_menus_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", roleHiddenMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + roleHiddenMenuIndexName + " ON role_hidden_menus (role_id, menu_id)").Error; err != nil {
			return err
		}
	}

	roleDisabledActionIndexName := "idx_role_disabled_actions_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", roleDisabledActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + roleDisabledActionIndexName + " ON role_disabled_actions (role_id, action_id)").Error; err != nil {
			return err
		}
	}

	teamBlockedMenuIndexName := "idx_team_blocked_menus_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", teamBlockedMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + teamBlockedMenuIndexName + " ON team_blocked_menus (team_id, menu_id)").Error; err != nil {
			return err
		}
	}

	teamBlockedActionIndexName := "idx_team_blocked_actions_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", teamBlockedActionIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + teamBlockedActionIndexName + " ON team_blocked_actions (team_id, action_id)").Error; err != nil {
			return err
		}
	}

	userHiddenMenuIndexName := "idx_user_hidden_menus_unique"
	DB.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", userHiddenMenuIndexName).Scan(&count)
	if count == 0 {
		if err := DB.Exec("CREATE UNIQUE INDEX " + userHiddenMenuIndexName + " ON user_hidden_menus (user_id, menu_id)").Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateLegacyUserRoles() error {
	// 先把 tenant_members 里的默认团队身份同步为租户级 user_roles
	insertDefaultTenantRoles := `
		INSERT INTO user_roles (user_id, role_id, tenant_id)
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
		  )
	`
	if err := DB.Exec(insertDefaultTenantRoles).Error; err != nil {
		return err
	}

	// 对“只有一个租户成员关系”的历史团队角色，自动迁移到对应 tenant_id
	updateSingleTenantScopedRoles := `
		UPDATE user_roles ur
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
		  )
	`
	if err := DB.Exec(updateSingleTenantScopedRoles).Error; err != nil {
		return err
	}

	// 对默认团队身份的旧全局记录做清理，避免和 tenant_members 同步后的租户角色重复
	deleteLegacyDefaultGlobalTeamRoles := `
		DELETE FROM user_roles ur
		USING roles r
		WHERE ur.role_id = r.id
		  AND ur.tenant_id IS NULL
		  AND r.code IN ('team_admin', 'team_member')
	`
	return DB.Exec(deleteLegacyDefaultGlobalTeamRoles).Error
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
