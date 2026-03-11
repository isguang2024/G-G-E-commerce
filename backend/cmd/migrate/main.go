package main

import (
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 迁移步骤枚举（初始使用 / 部署时执行顺序）
//
//  1. 表结构
//     - renameTeamTablesToTenant: 旧表重命名 team_* → tenant_*（仅已有库）
//     - database.AutoMigrate: 创建/更新所有模型表（见 internal/pkg/database/database.go）
//
//  2. 角色与权限
//     - ensureRoleScopeAndTeamPerm: roles 表增加 scope；scope=team 角色权限同步到 tenant_role_permissions
//     - initDefaultRoles: 插入默认角色 admin / team_admin / team_member（不存在则创建）
//
//  3. 用户
//     - initDefaultAdmin: 创建默认管理员 admin / admin123456，并分配 admin 角色
//
//  4. 菜单
//     - initDefaultMenus: 若 menus 表为空则插入完整默认菜单树（Dashboard / System / Team / Result / Exception 及子项）
//     - ensureTeamMenusIfMissing: 若全局不存在对应 path 的团队相关菜单，则创建独立的「团队管理」顶级菜单及其子菜单（兼容老库）
//     - ensureSystemMenuFlags: 为默认 path 的菜单打 is_system=true
//
//  5. 角色-菜单
//     - initDefaultRoleMenus: 若 role_menus 表为空则为 admin/team_admin/team_member 分配菜单，并同步 scope=team 到 tenant_role_permissions
//
// 详见 cmd/migrate/MIGRATION_INDEX.md
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting database migration...")

	// 初始化数据库
	_, err = database.Init(&cfg.DB)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	logger.Info("Database connected successfully")

	// 1. 表结构
	if err := renameTeamTablesToTenant(logger); err != nil {
		logger.Warn("renameTeamTablesToTenant failed", zap.Error(err))
	}
	
	if err := renameTenantSlugToRemark(logger); err != nil {
		logger.Warn("renameTenantSlugToRemark failed", zap.Error(err))
	}
	
	// 1.1 先初始化作用域（在AutoMigrate之前，因为roles表需要scope_id外键）
	if err := initDefaultScopes(logger); err != nil {
		logger.Warn("initDefaultScopes failed", zap.Error(err))
	}
	
	// 1.2 迁移roles表的scope到scope_id（在AutoMigrate之前）
	if err := migrateRoleScopeToScopeID(logger); err != nil {
		logger.Warn("migrateRoleScopeToScopeID failed", zap.Error(err))
	}
	
	// 1.3 执行AutoMigrate（现在roles表已经有scope_id了）
	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}
	if err := ensureTenantMemberRoleCodes(logger); err != nil {
		logger.Warn("ensureTenantMemberRoleCodes failed", zap.Error(err))
	}

	// 2. 角色与权限
	if err := ensureRoleScopeAndTeamPerm(logger); err != nil {
		logger.Warn("ensureRoleScopeAndTeamPerm failed", zap.Error(err))
	}
	if err := removeRoleIsSystemColumn(logger); err != nil {
		logger.Warn("removeRoleIsSystemColumn failed", zap.Error(err))
	}
	if err := removeRoleScopeColumn(logger); err != nil {
		logger.Warn("removeRoleScopeColumn failed", zap.Error(err))
	}
	logger.Info("Database migration completed successfully!")

	if err := initDefaultRoles(logger); err != nil {
		logger.Warn("Failed to initialize default roles", zap.Error(err))
	} else {
		logger.Info("Default roles initialized successfully")
	}

	// 3. 用户
	if err := initDefaultAdmin(logger); err != nil {
		logger.Warn("Failed to initialize default admin", zap.Error(err))
	} else {
		logger.Info("Default admin initialized successfully")
	}

	// 4. 菜单
	if err := initDefaultMenus(logger); err != nil {
		logger.Warn("Failed to initialize default menus", zap.Error(err))
	} else {
		logger.Info("Default menus initialized successfully")
	}
	if err := ensureTeamMenusIfMissing(logger); err != nil {
		logger.Warn("ensureTeamMenusIfMissing failed", zap.Error(err))
	}
	if err := ensureSystemMenuFlags(logger); err != nil {
		logger.Warn("Failed to ensure system menu flags", zap.Error(err))
	}

	// 5. 角色-菜单
	if err := initDefaultRoleMenus(logger); err != nil {
		logger.Warn("Failed to initialize role-menus", zap.Error(err))
	} else {
		logger.Info("Role-menus initialized successfully")
	}

	// 6. 移除团队自建角色相关表（角色仅使用全局 scope=team，不再支持租户自建角色）
	if err := dropTeamCustomRoleTables(logger); err != nil {
		logger.Warn("dropTeamCustomRoleTables failed", zap.Error(err))
	} else {
		logger.Info("Team custom role tables dropped (if existed)")
	}

	// 7. 为 roles 表添加 enabled 字段（角色启用/禁用）
	if err := addRoleEnabledColumn(logger); err != nil {
		logger.Warn("addRoleEnabledColumn failed", zap.Error(err))
	} else {
		logger.Info("Role enabled column added (if needed)")
	}

	// 8. 移除 tenant_members 表的 role 列（已使用 role_id 关联 roles 表）
	if err := removeTenantMemberRoleColumn(logger); err != nil {
		logger.Warn("removeTenantMemberRoleColumn failed", zap.Error(err))
	} else {
		logger.Info("Tenant member role column removed (if existed)")
	}
}

// dropTeamCustomRoleTables 删除团队自建角色相关表，角色统一为全局 scope=team
func dropTeamCustomRoleTables(logger *zap.Logger) error {
	tables := []string{"tenant_user_roles", "tenant_role_permissions", "tenant_roles"}
	for _, t := range tables {
		if err := database.DB.Exec("DROP TABLE IF EXISTS " + t + " CASCADE").Error; err != nil {
			return err
		}
		logger.Info("Dropped table (if existed)", zap.String("table", t))
	}
	return nil
}

// addRoleEnabledColumn 为 roles 表添加 enabled 字段（如果不存在）
func addRoleEnabledColumn(logger *zap.Logger) error {
	var exists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'roles' AND column_name = 'enabled'
		)
	`).Scan(&exists).Error
	if err != nil {
		return err
	}
	if exists {
		logger.Info("enabled column already exists in roles table")
		return nil
	}
	// 添加 enabled 列，默认值为 true，非空
	if err := database.DB.Exec("ALTER TABLE roles ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT true").Error; err != nil {
		return err
	}
	logger.Info("Added enabled column to roles table")
	return nil
}

// removeTenantMemberRoleColumn 移除 tenant_members 表的 role 列（已使用 role_id 关联 roles 表）
func removeTenantMemberRoleColumn(logger *zap.Logger) error {
	var exists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'tenant_members' AND column_name = 'role'
		)
	`).Scan(&exists).Error
	if err != nil {
		return err
	}
	if !exists {
		logger.Info("role column already removed from tenant_members table")
		return nil
	}
	// 删除 role 列
	if err := database.DB.Exec("ALTER TABLE tenant_members DROP COLUMN role").Error; err != nil {
		return err
	}
	logger.Info("Removed role column from tenant_members table")
	return nil
}

func renameTenantSlugToRemark(logger *zap.Logger) error {
	var exists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'tenants' AND column_name = 'slug'
		)
	`).Scan(&exists).Error
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	database.DB.Exec("ALTER TABLE tenants DROP CONSTRAINT IF EXISTS uni_tenants_slug")
	database.DB.Exec("DROP INDEX IF EXISTS idx_tenants_slug")
	database.DB.Exec("DROP INDEX IF EXISTS uni_tenants_slug")

	if err := database.DB.Exec("ALTER TABLE tenants RENAME COLUMN slug TO remark").Error; err != nil {
		return err
	}
	
	database.DB.Exec("ALTER TABLE tenants ALTER COLUMN remark DROP NOT NULL")
	database.DB.Exec("ALTER TABLE tenants ALTER COLUMN remark TYPE varchar(500)")

	logger.Info("Column 'slug' renamed to 'remark' in tenants table")
	return nil
}

// ensureTenantMemberRoleCodes 将 tenant_members.role 从旧值 owner/admin/editor 统一为角色编码 team_admin/team_member
func ensureTenantMemberRoleCodes(logger *zap.Logger) error {
	// 首先检查 role 列是否存在，如果不存在则添加
	var columnExists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_name = 'tenant_members' AND column_name = 'role'
		)
	`).Scan(&columnExists).Error
	if err != nil {
		return err
	}
	if !columnExists {
		// 添加 role 列
		if err := database.DB.Exec("ALTER TABLE tenant_members ADD COLUMN role varchar(50)").Error; err != nil {
			return err
		}
		logger.Info("Added 'role' column to tenant_members table")
	}

	// 从 role_id 迁移数据到 role（如果 role 为空且 role_id 有值）
	database.DB.Exec(`
		UPDATE tenant_members 
		SET role = CASE 
			WHEN rm.code IN ('team_admin', 'team_member') THEN rm.code
			ELSE 'team_member'
		END
		FROM roles rm
		WHERE tenant_members.role_id = rm.id AND (tenant_members.role IS NULL OR tenant_members.role = '')
	`)

	// 更新旧的角色编码
	res := database.DB.Exec("UPDATE tenant_members SET role = ? WHERE role IN (?, ?)", "team_admin", "owner", "admin")
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		logger.Info("Tenant members role migrated: owner/admin -> team_admin", zap.Int64("rows", res.RowsAffected))
	}
	res = database.DB.Exec("UPDATE tenant_members SET role = ? WHERE role = ?", "team_member", "editor")
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		logger.Info("Tenant members role migrated: editor -> team_member", zap.Int64("rows", res.RowsAffected))
	}
	return nil
}

// renameTeamTablesToTenant 将旧表重命名为 tenant_*（与 tenants / tenant_members 命名统一）
// 含 user_tenant_roles → tenant_user_roles，与 tenant_roles / tenant_members 的「tenant_ 前缀 + 实体」一致
func renameTeamTablesToTenant(logger *zap.Logger) error {
	renames := []struct{ old, new string }{
		{"team_roles", "tenant_roles"},
		{"team_role_permissions", "tenant_role_permissions"},
		{"user_team_roles", "tenant_user_roles"},
		{"user_tenant_roles", "tenant_user_roles"}, // 兼容此前已迁成 user_tenant_roles 的库
	}
	for _, r := range renames {
		err := database.DB.Exec("ALTER TABLE " + r.old + " RENAME TO " + r.new).Error
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				logger.Debug("Table already renamed or missing", zap.String("table", r.old))
				continue
			}
			return err
		}
		logger.Info("Table renamed", zap.String("from", r.old), zap.String("to", r.new))
	}
	return nil
}

// initDefaultRoles 初始化默认角色（全局 roles 表，scope: admin=global, team_admin/team_member=team）
func initDefaultRoles(logger *zap.Logger) error {
	// 获取作用域
	var globalScope, teamScope model.Scope
	if err := database.DB.Where("code = ?", "global").First(&globalScope).Error; err != nil {
		logger.Error("Failed to find global scope", zap.Error(err))
		return err
	}
	if err := database.DB.Where("code = ?", "team").First(&teamScope).Error; err != nil {
		logger.Error("Failed to find team scope", zap.Error(err))
		return err
	}

	roles := []struct {
		Code        string
		Name        string
		Description string
		ScopeID     uuid.UUID
		SortOrder   int
	}{
		{"admin", "管理员", "系统管理员，拥有所有权限", globalScope.ID, 1},
		{"team_admin", "团队管理员", "团队管理员，可以管理团队成员和团队内容", teamScope.ID, 2},
		{"team_member", "团队成员", "团队成员，可以查看和编辑团队内容", teamScope.ID, 3},
	}

	for _, roleData := range roles {
		var role model.Role
		result := database.DB.Where("code = ?", roleData.Code).First(&role)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				role = model.Role{
					Code:        roleData.Code,
					Name:        roleData.Name,
					Description: roleData.Description,
					ScopeID:     roleData.ScopeID,
					SortOrder:   roleData.SortOrder,
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
			// 已有角色：确保 scope_id 正确
			if role.ScopeID != roleData.ScopeID {
				_ = database.DB.Model(&role).Update("scope_id", roleData.ScopeID)
			}
			logger.Info("Role already exists", zap.String("code", roleData.Code))
		}
	}
	return nil
}

// initDefaultAdmin 初始化默认管理员
func initDefaultAdmin(logger *zap.Logger) error {
	userRepo := repository.NewUserRepository(database.DB)

	// 默认管理员信息
	defaultEmail := "admin@gg.com"
	defaultUsername := "admin"
	defaultPassword := "admin123456"
	defaultNickname := "系统管理员"

	// 检查用户名是否已存在
	exists, err := userRepo.ExistsByUsername(defaultUsername)
	if err != nil {
		return err
	}

	if exists {
		// 管理员已存在，检查是否需要分配角色
		user, err := userRepo.GetByUsername(defaultUsername)
		if err != nil {
			return err
		}
		
		// 检查是否已有角色
		var roleCount int64
		database.DB.Model(&model.UserRole{}).Where("user_id = ?", user.ID).Count(&roleCount)
		if roleCount == 0 {
			// 分配管理员角色
			if err := assignAdminRole(user.ID, logger); err != nil {
				return err
			}
		}
		
		logger.Info("Default admin already exists", zap.String("username", defaultUsername))
		return nil
	}

	// 加密密码
	passwordHash, err := password.Hash(defaultPassword)
	if err != nil {
		return err
	}

	// 创建管理员用户
	admin := &model.User{
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

	// 分配管理员角色
	if err := assignAdminRole(admin.ID, logger); err != nil {
		return err
	}

	return nil
}

// assignAdminRole 分配管理员角色
func assignAdminRole(userID uuid.UUID, logger *zap.Logger) error {
	// 查找管理员角色（全局）
	var adminRole model.Role
	if err := database.DB.Where("code = ? AND tenant_id IS NULL", "admin").First(&adminRole).Error; err != nil {
		logger.Error("Failed to find admin role", zap.Error(err))
		return err
	}

	// 检查是否已分配（全局角色 tenant_id IS NULL）
	var userRole model.UserRole
	result := database.DB.Where("user_id = ? AND role_id = ?", userID, adminRole.ID).First(&userRole)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 分配全局角色
			userRole = model.UserRole{
				UserID: userID,
				RoleID: adminRole.ID,
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

// ensureRoleScopeAndTeamPerm 确保 roles 有 scope 列（兼容旧库）；团队角色已统一为全局 scope=team，不再使用 tenant_role_permissions
func ensureRoleScopeAndTeamPerm(logger *zap.Logger) error {
	if err := database.DB.Exec("ALTER TABLE roles ADD COLUMN scope VARCHAR(20) NOT NULL DEFAULT 'global'").Error; err != nil && !strings.Contains(err.Error(), "already exists") {
		logger.Warn("Add scope column skipped or failed", zap.Error(err))
	}
	_ = database.DB.Exec("UPDATE roles SET scope = 'global' WHERE scope IS NULL OR scope = ''").Error
	_ = database.DB.Exec("UPDATE roles SET scope = 'team' WHERE code IN ('team_admin', 'team_member')").Error
	logger.Info("Role scope ensured")
	return nil
}

// initDefaultScopes 初始化默认作用域（global 和 team）
func initDefaultScopes(logger *zap.Logger) error {
	scopes := []struct {
		Code        string
		Name        string
		Description string
		SortOrder   int
	}{
		{"global", "全局", "跨应用全局作用域", 1},
		{"team", "团队", "仅团队功能使用的作用域", 2},
	}

	for _, scopeData := range scopes {
		var scope model.Scope
		result := database.DB.Where("code = ?", scopeData.Code).First(&scope)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				scope = model.Scope{
					Code:        scopeData.Code,
					Name:        scopeData.Name,
					Description: scopeData.Description,
					SortOrder:   scopeData.SortOrder,
				}
				if err := database.DB.Create(&scope).Error; err != nil {
					logger.Error("Failed to create scope", zap.String("code", scopeData.Code), zap.Error(err))
					return err
				}
				logger.Info("Scope created", zap.String("code", scopeData.Code), zap.String("name", scopeData.Name))
			} else {
				return result.Error
			}
		} else {
			logger.Info("Scope already exists", zap.String("code", scopeData.Code))
		}
	}
	return nil
}

// migrateRoleScopeToScopeID 将 roles 表的 scope 列迁移到 scope_id
func migrateRoleScopeToScopeID(logger *zap.Logger) error {
	// 检查 scope_id 列是否已存在
	var scopeIDExists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'roles' AND column_name = 'scope_id'
		)
	`).Scan(&scopeIDExists).Error
	if err != nil {
		logger.Warn("Failed to check scope_id column existence", zap.Error(err))
		return nil
	}

	// 如果 scope_id 列已存在，说明已经迁移过，跳过
	if scopeIDExists {
		logger.Info("scope_id column already exists, skip migration")
		return nil
	}

	// 检查 scope 列是否存在
	var scopeExists bool
	err = database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'roles' AND column_name = 'scope'
		)
	`).Scan(&scopeExists).Error
	if err != nil {
		logger.Warn("Failed to check scope column existence", zap.Error(err))
		return nil
	}

	if !scopeExists {
		logger.Info("scope column does not exist, skip migration")
		return nil
	}

	// 添加 scope_id 列（允许为空，稍后填充）
	if err := database.DB.Exec("ALTER TABLE roles ADD COLUMN scope_id UUID").Error; err != nil {
		logger.Warn("Failed to add scope_id column", zap.Error(err))
		return err
	}

	// 获取作用域映射
	var globalScope, teamScope model.Scope
	if err := database.DB.Where("code = ?", "global").First(&globalScope).Error; err != nil {
		logger.Error("Failed to find global scope", zap.Error(err))
		return err
	}
	if err := database.DB.Where("code = ?", "team").First(&teamScope).Error; err != nil {
		logger.Error("Failed to find team scope", zap.Error(err))
		return err
	}

	// 更新现有角色的 scope_id
	res := database.DB.Exec("UPDATE roles SET scope_id = ? WHERE scope = 'global'", globalScope.ID)
	if res.Error != nil {
		logger.Warn("Failed to update global scope_id", zap.Error(res.Error))
	} else {
		logger.Info("Updated global scope_id", zap.Int64("rows", res.RowsAffected))
	}

	res = database.DB.Exec("UPDATE roles SET scope_id = ? WHERE scope = 'team'", teamScope.ID)
	if res.Error != nil {
		logger.Warn("Failed to update team scope_id", zap.Error(res.Error))
	} else {
		logger.Info("Updated team scope_id", zap.Int64("rows", res.RowsAffected))
	}

	// 设置 scope_id 为 NOT NULL
	if err := database.DB.Exec("ALTER TABLE roles ALTER COLUMN scope_id SET NOT NULL").Error; err != nil {
		logger.Warn("Failed to set scope_id NOT NULL", zap.Error(err))
	}

	// 添加外键约束
	if err := database.DB.Exec("ALTER TABLE roles ADD CONSTRAINT fk_roles_scope FOREIGN KEY (scope_id) REFERENCES scopes(id)").Error; err != nil {
		logger.Warn("Failed to add foreign key constraint", zap.Error(err))
	}

	// 删除旧的 scope 列
	if err := database.DB.Exec("ALTER TABLE roles DROP COLUMN IF EXISTS scope").Error; err != nil {
		logger.Warn("Failed to drop scope column", zap.Error(err))
		return err
	}

	logger.Info("Role scope migration completed")
	return nil
}

// removeRoleScopeColumn 删除 roles 表的 scope 列（确保只保留 scope_id）
func removeRoleScopeColumn(logger *zap.Logger) error {
	// 检查列是否存在（PostgreSQL）
	var exists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'roles' AND column_name = 'scope'
		)
	`).Scan(&exists).Error
	if err != nil {
		logger.Warn("Failed to check scope column existence", zap.Error(err))
		return nil // 忽略错误，继续执行
	}
	if !exists {
		logger.Info("scope column does not exist, skip removal")
		return nil
	}
	// 删除列
	if err := database.DB.Exec("ALTER TABLE roles DROP COLUMN IF EXISTS scope").Error; err != nil {
		logger.Warn("Failed to drop scope column", zap.Error(err))
		return err
	}
	logger.Info("scope column removed from roles table")
	return nil
}

// removeRoleIsSystemColumn 删除 roles 表的 is_system 列
func removeRoleIsSystemColumn(logger *zap.Logger) error {
	// 检查列是否存在（PostgreSQL）
	var exists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'roles' AND column_name = 'is_system'
		)
	`).Scan(&exists).Error
	if err != nil {
		logger.Warn("Failed to check is_system column existence", zap.Error(err))
		return nil // 忽略错误，继续执行
	}
	if !exists {
		logger.Info("is_system column does not exist, skip removal")
		return nil
	}
	// 删除列
	if err := database.DB.Exec("ALTER TABLE roles DROP COLUMN IF EXISTS is_system").Error; err != nil {
		logger.Warn("Failed to drop is_system column", zap.Error(err))
		return err
	}
	logger.Info("is_system column removed from roles table")
	return nil
}

// initDefaultMenus 初始化默认菜单（与前端路由一致，带角色权限；移植项目时保留）
func initDefaultMenus(logger *zap.Logger) error {
	var count int64
	if err := database.DB.Model(&model.Menu{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		logger.Info("Menus already exist, skip seed")
		return nil
	}

	metaSuperAdmin := model.MetaJSON{"roles": []interface{}{"R_SUPER"}}
	metaSuperAdminAndAdmin := model.MetaJSON{"roles": []interface{}{"R_SUPER", "R_ADMIN"}}

	// 一级菜单（系统默认：自动出现在侧栏且不可删除）
	dashboard := model.Menu{
		Path: "/dashboard", Name: "Dashboard", Component: "/index/index",
		Title: "menus.dashboard.title", Icon: "ri:pie-chart-line",
		SortOrder: 1, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}
	if err := database.DB.Create(&dashboard).Error; err != nil {
		return err
	}
	system := model.Menu{
		Path: "/system", Name: "System", Component: "/index/index",
		Title: "menus.system.title", Icon: "ri:user-3-line",
		SortOrder: 2, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}
	if err := database.DB.Create(&system).Error; err != nil {
		return err
	}
	result := model.Menu{
		Path: "/result", Name: "Result", Component: "/index/index",
		Title: "menus.result.title", Icon: "ri:checkbox-circle-line",
		SortOrder: 3, Meta: nil, IsSystem: true,
	}
	if err := database.DB.Create(&result).Error; err != nil {
		return err
	}
	exception := model.Menu{
		Path: "/exception", Name: "Exception", Component: "/index/index",
		Title: "menus.exception.title", Icon: "ri:error-warning-line",
		SortOrder: 4, Meta: nil, IsSystem: true,
	}
	if err := database.DB.Create(&exception).Error; err != nil {
		return err
	}

	// Dashboard 子菜单
	_ = database.DB.Create(&model.Menu{
		ParentID: &dashboard.ID, Path: "console", Name: "Console", Component: "/dashboard/console",
		Title: "menus.dashboard.console", SortOrder: 1,
		Meta: model.MetaJSON{"keepAlive": false, "fixedTab": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &dashboard.ID, Path: "user-center", Name: "UserCenter", Component: "/system/user-center",
		Title: "menus.system.userCenter", SortOrder: 2,
		Meta: model.MetaJSON{"isHide": true, "keepAlive": true, "isHideTab": true}, IsSystem: true,
	}).Error

	// System 子菜单
	_ = database.DB.Create(&model.Menu{
		ParentID: &system.ID, Path: "role", Name: "Role", Component: "/system/role",
		Title: "menus.system.role", SortOrder: 1, Meta: metaSuperAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &system.ID, Path: "user", Name: "User", Component: "/system/user",
		Title: "menus.system.user", SortOrder: 2, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &system.ID, Path: "pages-association", Name: "PageAssociation", Component: "/system/pages-association",
		Title: "menus.system.pagesAssociation", SortOrder: 3, Meta: metaSuperAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &system.ID, Path: "menu", Name: "Menus", Component: "/system/menu",
		Title: "menus.system.menu", SortOrder: 4, IsSystem: true,
		Meta: model.MetaJSON{
			"roles": []interface{}{"R_SUPER"},
			"keepAlive": true,
			"authList": []interface{}{
				map[string]interface{}{"title": "新增", "authMark": "add"},
				map[string]interface{}{"title": "编辑", "authMark": "edit"},
				map[string]interface{}{"title": "删除", "authMark": "delete"},
			},
		},
	}).Error

	// Team 顶级菜单（独立的团队管理菜单）
	team := model.Menu{
		Path: "/team", Name: "TeamRoot", Component: "/index/index",
		Title: "menus.system.team", Icon: "ri:team-line",
		SortOrder: 5, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}
	if err := database.DB.Create(&team).Error; err != nil {
		return err
	}

	// Team 子菜单
	metaTeamAdminOnly := model.MetaJSON{"roles": []interface{}{"R_ADMIN"}, "keepAlive": true}
	_ = database.DB.Create(&model.Menu{
		ParentID: &team.ID, Path: "team-roles-permissions", Name: "TeamRolesAndPermissions", Component: "/system/team-roles-permissions",
		Title: "menus.system.teamRolesAndPermissions", SortOrder: 1, Meta: metaSuperAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &team.ID, Path: "management", Name: "TeamManagementRedirect", Component: "/team/team-members",
		Title: "menus.system.teamMembers", SortOrder: 2, Meta: model.MetaJSON{"keepAlive": true, "roles": []interface{}{"R_ADMIN"}}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &team.ID, Path: "members", Name: "TeamMembers", Component: "/team/team-members",
		Title: "menus.system.teamMembers", SortOrder: 3, Meta: metaTeamAdminOnly, IsSystem: true,
	}).Error

	// Result 子菜单
	_ = database.DB.Create(&model.Menu{
		ParentID: &result.ID, Path: "success", Name: "ResultSuccess", Component: "/result/success",
		Title: "menus.result.success", Icon: "ri:checkbox-circle-line", SortOrder: 1,
		Meta: model.MetaJSON{"keepAlive": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &result.ID, Path: "fail", Name: "ResultFail", Component: "/result/fail",
		Title: "menus.result.fail", Icon: "ri:close-circle-line", SortOrder: 2,
		Meta: model.MetaJSON{"keepAlive": true}, IsSystem: true,
	}).Error

	// Exception 子菜单
	_ = database.DB.Create(&model.Menu{
		ParentID: &exception.ID, Path: "403", Name: "Exception403", Component: "/exception/403",
		Title: "menus.exception.forbidden", SortOrder: 1,
		Meta: model.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &exception.ID, Path: "404", Name: "Exception404", Component: "/exception/404",
		Title: "menus.exception.notFound", SortOrder: 2,
		Meta: model.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&model.Menu{
		ParentID: &exception.ID, Path: "500", Name: "Exception500", Component: "/exception/500",
		Title: "menus.exception.serverError", SortOrder: 3,
		Meta: model.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}, IsSystem: true,
	}).Error

	logger.Info("Default menus seeded")
	return nil
}

// ensureTeamMenusIfMissing 若全局不存在对应 path 的团队相关菜单，则创建独立的「团队管理」顶级菜单及其子菜单（便于老库升级）。
// 判断按「全局 path」：避免用户把「团队管理」挂在 /team 或把子项挂在「团队管理」下后被重复插入。
func ensureTeamMenusIfMissing(logger *zap.Logger) error {
	// 检查团队管理顶级菜单是否存在
	var teamRoot model.Menu
	err := database.DB.Where("path = ? AND name = ?", "/team", "TeamRoot").First(&teamRoot).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	metaSuperAdmin := model.MetaJSON{"roles": []interface{}{"R_SUPER"}}
	metaSuperAdminAndAdmin := model.MetaJSON{"roles": []interface{}{"R_SUPER", "R_ADMIN"}}
	metaTeamAdminOnly := model.MetaJSON{"roles": []interface{}{"R_ADMIN"}, "keepAlive": true}

	// 如果团队管理顶级菜单不存在，创建它
	if errors.Is(err, gorm.ErrRecordNotFound) {
		teamRoot = model.Menu{
			Path: "/team", Name: "TeamRoot", Component: "/index/index",
			Title: "menus.system.team", Icon: "ri:team-line",
			SortOrder: 5, Meta: metaSuperAdminAndAdmin, IsSystem: true,
		}
		if err := database.DB.Create(&teamRoot).Error; err != nil {
			return err
		}
		logger.Info("Team root menu created", zap.String("path", "/team"))
	}

	// 团队管理子菜单配置
	teamChildMenuSpecs := []struct {
		pathsToCheck []string
		path         string
		name         string
		component    string
		title        string
		sortOrder    int
		meta         model.MetaJSON
		roleCodes    []string
	}{
		{[]string{"team-roles-permissions"}, "team-roles-permissions", "TeamRolesAndPermissions", "/system/team-roles-permissions", "menus.system.teamRolesAndPermissions", 1, metaSuperAdmin, []string{"admin"}},
		{[]string{"management"}, "management", "TeamManagementRedirect", "/team/team-members", "menus.system.teamMembers", 2, model.MetaJSON{"keepAlive": true, "roles": []interface{}{"R_ADMIN"}}, []string{"admin", "team_admin"}},
		{[]string{"members"}, "members", "TeamMembers", "/team/team-members", "menus.system.teamMembers", 3, metaTeamAdminOnly, []string{"admin", "team_admin"}},
	}

	for _, spec := range teamChildMenuSpecs {
		var exist int64
		if err := database.DB.Model(&model.Menu{}).Where("path IN ?", spec.pathsToCheck).Count(&exist).Error; err != nil {
			return err
		}
		if exist > 0 {
			continue
		}
		m := model.Menu{
			ParentID:  &teamRoot.ID,
			Path:      spec.path,
			Name:      spec.name,
			Component: spec.component,
			Title:     spec.title,
			SortOrder: spec.sortOrder,
			Meta:      spec.meta,
			IsSystem:  true,
		}
		if err := database.DB.Create(&m).Error; err != nil {
			return err
		}
		var roleIDs []uuid.UUID
		database.DB.Model(&model.Role{}).Where("code IN ?", spec.roleCodes).Pluck("id", &roleIDs)
		for _, roleID := range roleIDs {
			_ = database.DB.Create(&model.RoleMenu{RoleID: roleID, MenuID: m.ID}).Error
		}
		logger.Info("Team menu ensured", zap.String("path", spec.path))
	}
	return nil
}

// ensureSystemMenuFlags 为已有数据库中的默认菜单打上 is_system=true（与 initDefaultMenus 中的 path 一致）
func ensureSystemMenuFlags(logger *zap.Logger) error {
	defaultPaths := []string{
		"/dashboard", "console", "user-center", "/system", "user", "role", "pages-association", "menu",
		"/team", "team-roles-permissions", "management", "members",
		"/result", "success", "fail", "/exception", "403", "404", "500",
	}
	res := database.DB.Model(&model.Menu{}).Where("path IN ?", defaultPaths).Update("is_system", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		logger.Info("System menu flags updated", zap.Int64("rows", res.RowsAffected))
	}
	return nil
}

// initDefaultRoleMenus 为默认角色分配菜单（与菜单 meta.roles 一致；若已有数据则跳过）
func initDefaultRoleMenus(logger *zap.Logger) error {
	var count int64
	if err := database.DB.Model(&model.RoleMenu{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		logger.Info("Role-menus already exist, skip seed")
		return nil
	}

	var roles []model.Role
	if err := database.DB.Where("code IN ?", []string{"admin", "team_admin", "team_member"}).Find(&roles).Error; err != nil {
		return err
	}
	roleByCode := make(map[string]uuid.UUID)
	for _, r := range roles {
		roleByCode[r.Code] = r.ID
	}
	adminID, hasAdmin := roleByCode["admin"]
	teamAdminID, hasTeamAdmin := roleByCode["team_admin"]
	teamMemberID, hasTeamMember := roleByCode["team_member"]

	var menus []model.Menu
	if err := database.DB.Find(&menus).Error; err != nil {
		return err
	}

	roleMenus := make([]model.RoleMenu, 0)
	for _, m := range menus {
		rolesVal, _ := m.Meta["roles"].([]interface{})
		addAdmin, addTeamAdmin, addTeamMember := false, false, false
		if len(rolesVal) == 0 {
			addAdmin, addTeamAdmin, addTeamMember = true, true, true
		} else {
			for _, r := range rolesVal {
				s, _ := r.(string)
				switch s {
				case "R_SUPER":
					addAdmin = true
				case "R_ADMIN":
					addAdmin = true
					addTeamAdmin = true
				case "R_USER":
					addTeamMember = true
				}
			}
		}
		if addAdmin && hasAdmin {
			roleMenus = append(roleMenus, model.RoleMenu{RoleID: adminID, MenuID: m.ID})
		}
		if addTeamAdmin && hasTeamAdmin {
			roleMenus = append(roleMenus, model.RoleMenu{RoleID: teamAdminID, MenuID: m.ID})
		}
		if addTeamMember && hasTeamMember {
			roleMenus = append(roleMenus, model.RoleMenu{RoleID: teamMemberID, MenuID: m.ID})
		}
	}
	seen := make(map[string]struct{})
	for _, rm := range roleMenus {
		key := rm.RoleID.String() + ":" + rm.MenuID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if err := database.DB.Create(&rm).Error; err != nil {
			return err
		}
	}
	logger.Info("Default role-menus seeded")
	return nil
}
