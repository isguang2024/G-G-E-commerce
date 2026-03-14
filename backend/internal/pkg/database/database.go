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
		&models.Scope{},
		&models.Role{},
		&models.UserRole{},
		&models.RoleMenu{},
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
