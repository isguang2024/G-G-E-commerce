package main

import (
	"os"

	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

func fail(logger *zap.Logger, message string, fields ...zap.Field) {
	logger.Error(message, fields...)
	os.Exit(1)
}

func main() {
	bootstrapLogger, _ := zap.NewDevelopment()
	defer func() { _ = bootstrapLogger.Sync() }()
	bootstrapLogger.Info("后端诊断开始", zap.String("tool", "cmd/diagnose"))

	bootstrapLogger.Info("检查配置文件")
	cfg, err := config.Load()
	if err != nil {
		fail(bootstrapLogger, "配置文件加载失败",
			zap.Error(err),
			zap.String("hint", "请确保 configs/config.yaml 文件存在"),
		)
	}

	diagLogger, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		fail(bootstrapLogger, "日志初始化失败", zap.Error(err))
	}
	defer func() { _ = diagLogger.Sync() }()
	diagLogger.Info("配置文件加载成功")

	diagLogger.Info("检查数据库连接")
	db, err := database.Init(&cfg.DB)
	if err != nil {
		fail(diagLogger, "数据库连接失败",
			zap.Error(err),
			zap.String("hint_1", "请确保 PostgreSQL 服务正在运行"),
			zap.String("hint_2", "检查 configs/config.yaml 中的数据库配置"),
			zap.String("db_host", cfg.DB.Host),
			zap.Int("db_port", cfg.DB.Port),
			zap.String("db_user", cfg.DB.User),
			zap.String("db_name", cfg.DB.DBName),
		)
	}
	defer database.Close()
	diagLogger.Info("数据库连接成功")

	diagLogger.Info("检查数据库表")
	var count int64
	if err := db.Model(&user.User{}).Count(&count).Error; err != nil {
		fail(diagLogger, "用户表不存在或无法访问",
			zap.Error(err),
			zap.String("hint", "请运行数据库迁移命令: go run cmd/migrate/main.go"),
		)
	}
	diagLogger.Info("用户表检查完成", zap.Int64("user_count", count))

	diagLogger.Info("检查默认角色")
	var roleCount int64
	if err := db.Model(&user.Role{}).Count(&roleCount).Error; err != nil {
		diagLogger.Error("角色表不存在或无法访问", zap.Error(err))
	} else {
		diagLogger.Info("角色表检查完成", zap.Int64("role_count", roleCount))

		defaultRoles := []string{"admin", "collaboration_workspace_admin", "collaboration_workspace_member"}
		for _, roleCode := range defaultRoles {
			var role user.Role
			if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
				diagLogger.Warn("默认角色不存在", zap.String("role_code", roleCode))
			} else {
				diagLogger.Info("默认角色存在",
					zap.String("role_code", role.Code),
					zap.String("role_name", role.Name),
				)
			}
		}
	}

	diagLogger.Info("检查默认管理员账号")
	userRepo := user.NewUserRepository(db)
	adminUsername := "admin"
	exists, err := userRepo.ExistsByUsername(adminUsername)
	if err != nil {
		fail(diagLogger, "检查管理员账号失败", zap.Error(err))
	}

	if !exists {
		diagLogger.Warn("默认管理员账号不存在",
			zap.String("username", adminUsername),
			zap.String("hint_1", "请运行以下命令创建默认管理员: go run cmd/init-admin/main.go"),
			zap.String("hint_2", "或运行数据库迁移命令（会自动创建）: go run cmd/migrate/main.go"),
		)
	} else {
		diagLogger.Info("默认管理员账号存在", zap.String("username", adminUsername))
		adminUser, err := userRepo.GetByUsername(adminUsername)
		if err == nil {
			fields := []zap.Field{
				zap.String("username", adminUser.Username),
				zap.String("nickname", adminUser.Nickname),
				zap.String("status", adminUser.Status),
				zap.Bool("is_super_admin", adminUser.IsSuperAdmin),
			}
			if adminUser.Email != "" {
				fields = append(fields, zap.String("email", adminUser.Email))
			}
			roleNames := make([]string, 0, len(adminUser.Roles))
			for _, role := range adminUser.Roles {
				roleNames = append(roleNames, role.Name)
			}
			if len(roleNames) > 0 {
				fields = append(fields, zap.Strings("roles", roleNames))
			} else {
				fields = append(fields, zap.Bool("roles_empty", true))
			}
			diagLogger.Info("默认管理员详情", fields...)
		}
	}

	diagLogger.Info("检查 JWT 配置")
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key-change-in-production" {
		diagLogger.Warn("JWT Secret 使用默认值，建议在生产环境中修改")
	} else {
		diagLogger.Info("JWT Secret 已配置")
	}
	diagLogger.Info("JWT 配置摘要",
		zap.Int("access_expire_minutes", cfg.JWT.AccessExpire),
		zap.Int("refresh_expire_minutes", cfg.JWT.RefreshExpire),
	)

	diagLogger.Info("诊断完成",
		zap.String("next_step_1", "如果数据库表不存在，运行: go run cmd/migrate/main.go"),
		zap.String("next_step_2", "如果管理员不存在，运行: go run cmd/init-admin/main.go"),
		zap.String("next_step_3", "启动服务器: go run cmd/server/main.go"),
	)
}
