package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

func main() {
	fmt.Println("🔍 G&G E-commerce 后端诊断工具")
	fmt.Println("=====================================")

	fmt.Println("1. 检查配置文件...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 配置文件加载失败: %v\n", err)
		fmt.Println("\n💡 提示: 请确保 configs/config.yaml 文件存在")
		os.Exit(1)
	}
	fmt.Println("✅ 配置文件加载成功")

	fmt.Println("\n2. 检查日志...")
	_, err = logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	fmt.Println("\n3. 检查数据库连接...")
	db, err := database.Init(&cfg.DB)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		fmt.Printf("\n💡 提示:\n")
		fmt.Printf("   - 请确保 PostgreSQL 服务正在运行\n")
		fmt.Printf("   - 检查 configs/config.yaml 中的数据库配置\n")
		fmt.Printf("   - 数据库配置: host=%s, port=%d, user=%s, dbname=%s\n",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.DBName)
		os.Exit(1)
	}
	defer database.Close()
	fmt.Println("✅ 数据库连接成功")

	fmt.Println("\n4. 检查数据库表...")
	var count int64
	if err := db.Model(&user.User{}).Count(&count).Error; err != nil {
		fmt.Printf("❌ 用户表不存在或无法访问: %v\n", err)
		fmt.Println("\n💡 提示: 请运行数据库迁移命令:")
		fmt.Println("   go run cmd/migrate/main.go")
		os.Exit(1)
	}
	fmt.Printf("✅ 用户表存在，当前有 %d 条记录\n", count)

	fmt.Println("\n5. 检查默认角色...")
	var roleCount int64
	if err := db.Model(&user.Role{}).Count(&roleCount).Error; err != nil {
		fmt.Printf("❌ 角色表不存在或无法访问: %v\n", err)
	} else {
		fmt.Printf("✅ 角色表存在，当前有 %d 个角色\n", roleCount)

		defaultRoles := []string{"admin", "team_admin", "team_member"}
		for _, roleCode := range defaultRoles {
			var role user.Role
			if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
				fmt.Printf("   ⚠️  角色 %s 不存在\n", roleCode)
			} else {
				fmt.Printf("   ✅ %s (%s)\n", role.Name, role.Code)
			}
		}
	}

	fmt.Println("\n6. 检查默认管理员账号...")
	userRepo := user.NewUserRepository(db)
	adminUsername := "admin"
	exists, err := userRepo.ExistsByUsername(adminUsername)
	if err != nil {
		fmt.Printf("❌ 检查管理员账号失败: %v\n", err)
		os.Exit(1)
	}

	if !exists {
		fmt.Printf("⚠️  默认管理员账号不存在 (username: %s)\n", adminUsername)
		fmt.Println("\n💡 提示: 请运行以下命令创建默认管理员:")
		fmt.Println("   go run cmd/init-admin/main.go")
		fmt.Println("\n   或运行数据库迁移命令（会自动创建）:")
		fmt.Println("   go run cmd/migrate/main.go")
	} else {
		fmt.Printf("✅ 默认管理员账号存在 (username: %s)\n", adminUsername)
		user, err := userRepo.GetByUsername(adminUsername)
		if err == nil {
			fmt.Printf("   - 用户名: %s\n", user.Username)
			if user.Email != "" {
				fmt.Printf("   - 邮箱: %s\n", user.Email)
			}
			fmt.Printf("   - 昵称: %s\n", user.Nickname)
			fmt.Printf("   - 状态: %s\n", user.Status)
			fmt.Printf("   - 超级管理员: %v\n", user.IsSuperAdmin)

			if len(user.Roles) > 0 {
				fmt.Printf("   - 角色: ")
				for i, role := range user.Roles {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", role.Name)
				}
				fmt.Println()
			} else {
				fmt.Printf("   ⚠️  未分配角色\n")
			}
		}
	}

	fmt.Println("\n7. 检查 JWT 配置...")
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key-change-in-production" {
		fmt.Println("⚠️  JWT Secret 使用默认值，建议在生产环境中修改")
	} else {
		fmt.Println("✅ JWT Secret 已配置")
	}
	fmt.Printf("   - Access Token 有效期: %d 分钟\n", cfg.JWT.AccessExpire)
	fmt.Printf("   - Refresh Token 有效期: %d 分钟\n", cfg.JWT.RefreshExpire)

	fmt.Println("\n=====================================")
	fmt.Println("✅ 诊断完成！")
	fmt.Println("\n📝 下一步:")
	fmt.Println("   1. 如果数据库表不存在，运行: go run cmd/migrate/main.go")
	fmt.Println("   2. 如果管理员不存在，运行: go run cmd/init-admin/main.go")
	fmt.Println("   3. 启动服务器: go run cmd/server/main.go")
}
