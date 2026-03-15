package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
)

func main() {
	var (
		email        = flag.String("email", "admin@gg.com", "管理员邮箱")
		username     = flag.String("username", "admin", "管理员用户名")
		passwordFlag = flag.String("password", "admin123456", "管理员密码")
		nickname     = flag.String("nickname", "系统管理员", "管理员昵称")
	)
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

	logger.Info("Starting admin initialization...")

	_, err = database.Init(&cfg.DB)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	logger.Info("Database connected successfully")

	userRepo := user.NewUserRepository(database.DB)

	exists, err := userRepo.ExistsByUsername(*username)
	if err != nil {
		logger.Fatal("Failed to check username existence", zap.Error(err))
	}

	if exists {
		logger.Warn("Admin already exists", zap.String("username", *username))
		user, err := userRepo.GetByUsername(*username)
		if err != nil {
			logger.Fatal("Failed to get existing admin", zap.Error(err))
		}
		fmt.Printf("\n管理员账号已存在:\n")
		fmt.Printf("  ID: %s\n", user.ID.String())
		fmt.Printf("  邮箱: %s\n", user.Email)
		fmt.Printf("  用户名: %s\n", user.Username)
		fmt.Printf("  昵称: %s\n", user.Nickname)
		fmt.Printf("  超级管理员: %v\n", user.IsSuperAdmin)
		fmt.Printf("  状态: %s\n\n", user.Status)
		os.Exit(0)
	}

	passwordHash, err := password.Hash(*passwordFlag)
	if err != nil {
		logger.Fatal("Failed to hash password", zap.Error(err))
	}

	admin := &user.User{
		Email:        *email,
		Username:     *username,
		PasswordHash: passwordHash,
		Nickname:     *nickname,
		Status:       "active",
		IsSuperAdmin: true,
	}

	if err := userRepo.Create(admin); err != nil {
		logger.Fatal("Failed to create admin", zap.Error(err))
	}

	if err := assignAdminRole(admin.ID, logger); err != nil {
		logger.Warn("Failed to assign admin role", zap.Error(err))
	}

	logger.Info("Admin created successfully",
		zap.String("email", admin.Email),
		zap.String("username", admin.Username),
		zap.String("user_id", admin.ID.String()),
	)

	fmt.Printf("\n✅ 默认管理员账号创建成功!\n\n")
	fmt.Printf("登录信息:\n")
	fmt.Printf("  用户名: %s\n", admin.Username)
	if admin.Email != "" {
		fmt.Printf("  邮箱: %s\n", admin.Email)
	}
	fmt.Printf("  密码: %s\n", *passwordFlag)
	fmt.Printf("  用户ID: %s\n", admin.ID.String())
	fmt.Printf("  超级管理员: 是\n")
	fmt.Printf("  角色: 管理员\n\n")
	fmt.Printf("⚠️  请妥善保管密码，首次登录后建议修改密码!\n\n")
}

func assignAdminRole(userID uuid.UUID, logger *zap.Logger) error {
	var adminRole user.Role
	if err := database.DB.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		logger.Error("Failed to find admin role", zap.Error(err))
		return err
	}

	var userRole user.UserRole
	result := database.DB.Where("user_id = ? AND role_id = ? AND tenant_id IS NULL", userID, adminRole.ID).First(&userRole)
	if result.Error == gorm.ErrRecordNotFound {
		userRole = user.UserRole{
			UserID: userID,
			RoleID: adminRole.ID,
			TenantID: nil,
		}
		if err := database.DB.Create(&userRole).Error; err != nil {
			logger.Error("Failed to assign admin role", zap.Error(err))
			return err
		}
		logger.Info("Admin role assigned", zap.String("user_id", userID.String()))
	}

	return nil
}
