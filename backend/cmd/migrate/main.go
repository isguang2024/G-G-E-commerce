package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
)

func main() {
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

	if err := renameTeamTablesToTenant(logger); err != nil {
		logger.Warn("renameTeamTablesToTenant failed", zap.Error(err))
	}

	if err := renameTenantSlugToRemark(logger); err != nil {
		logger.Warn("renameTenantSlugToRemark failed", zap.Error(err))
	}

	if err := initDefaultScopes(logger); err != nil {
		logger.Warn("initDefaultScopes failed", zap.Error(err))
	}

	if err := migrateRoleScopeToScopeID(logger); err != nil {
		logger.Warn("migrateRoleScopeToScopeID failed", zap.Error(err))
	}

	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}
	if err := ensureTenantMemberRoleCodes(logger); err != nil {
		logger.Warn("ensureTenantMemberRoleCodes failed", zap.Error(err))
	}

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

	if err := initDefaultAdmin(logger); err != nil {
		logger.Warn("Failed to initialize default admin", zap.Error(err))
	} else {
		logger.Info("Default admin initialized successfully")
	}

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

	if err := initDefaultRoleMenus(logger); err != nil {
		logger.Warn("Failed to initialize role-menus", zap.Error(err))
	} else {
		logger.Info("Role-menus initialized successfully")
	}

	if err := dropTeamCustomRoleTables(logger); err != nil {
		logger.Warn("dropTeamCustomRoleTables failed", zap.Error(err))
	} else {
		logger.Info("Team custom role tables dropped (if existed)")
	}

	if err := addRoleEnabledColumn(logger); err != nil {
		logger.Warn("addRoleEnabledColumn failed", zap.Error(err))
	} else {
		logger.Info("Role enabled column added (if needed)")
	}

	if err := removeTenantMemberRoleColumn(logger); err != nil {
		logger.Warn("removeTenantMemberRoleColumn failed", zap.Error(err))
	} else {
		logger.Info("Tenant member role column removed (if existed)")
	}

	fmt.Println("✅ Migration completed successfully!")
}

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
	if err := database.DB.Exec("ALTER TABLE roles ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT true").Error; err != nil {
		return err
	}
	logger.Info("Added enabled column to roles table")
	return nil
}

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

func ensureTenantMemberRoleCodes(logger *zap.Logger) error {
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
		if err := database.DB.Exec("ALTER TABLE tenant_members ADD COLUMN role varchar(50)").Error; err != nil {
			return err
		}
		logger.Info("Added 'role' column to tenant_members table")
	}

	database.DB.Exec(`
		UPDATE tenant_members 
		SET role = CASE 
			WHEN rm.code IN ('team_admin', 'team_member') THEN rm.code
			ELSE 'team_member'
		END
		FROM roles rm
		WHERE tenant_members.role_id = rm.id AND (tenant_members.role IS NULL OR tenant_members.role = '')
	`)

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

func renameTeamTablesToTenant(logger *zap.Logger) error {
	renames := []struct{ old, new string }{
		{"team_roles", "tenant_roles"},
		{"team_role_permissions", "tenant_role_permissions"},
		{"user_team_roles", "tenant_user_roles"},
		{"user_tenant_roles", "tenant_user_roles"},
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

func initDefaultRoles(logger *zap.Logger) error {
	var globalScope, teamScope usermodel.Scope
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
		var role usermodel.Role
		result := database.DB.Where("code = ?", roleData.Code).First(&role)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				role = usermodel.Role{
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
			if role.ScopeID != roleData.ScopeID {
				_ = database.DB.Model(&role).Update("scope_id", roleData.ScopeID)
			}
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

		var roleCount int64
		database.DB.Model(&usermodel.UserRole{}).Where("user_id = ?", adminUser.ID).Count(&roleCount)
		if roleCount == 0 {
			if err := assignAdminRole(adminUser.ID, logger); err != nil {
				return err
			}
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
	if err := database.DB.Where("code = ? AND scope IS NULL", "admin").First(&adminRole).Error; err != nil {
		logger.Error("Failed to find admin role", zap.Error(err))
		return err
	}

	var userRole usermodel.UserRole
	result := database.DB.Where("user_id = ? AND role_id = ?", userID, adminRole.ID).First(&userRole)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			userRole = usermodel.UserRole{
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

func ensureRoleScopeAndTeamPerm(logger *zap.Logger) error {
	if err := database.DB.Exec("ALTER TABLE roles ADD COLUMN scope VARCHAR(20) NOT NULL DEFAULT 'global'").Error; err != nil && !strings.Contains(err.Error(), "already exists") {
		logger.Warn("Add scope column skipped or failed", zap.Error(err))
	}
	_ = database.DB.Exec("UPDATE roles SET scope = 'global' WHERE scope IS NULL OR scope = ''").Error
	_ = database.DB.Exec("UPDATE roles SET scope = 'team' WHERE code IN ('team_admin', 'team_member')").Error
	logger.Info("Role scope ensured")
	return nil
}

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
		var scope usermodel.Scope
		result := database.DB.Where("code = ?", scopeData.Code).First(&scope)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				scope = usermodel.Scope{
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

func migrateRoleScopeToScopeID(logger *zap.Logger) error {
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

	if scopeIDExists {
		logger.Info("scope_id column already exists, skip migration")
		return nil
	}

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

	if err := database.DB.Exec("ALTER TABLE roles ADD COLUMN scope_id UUID").Error; err != nil {
		logger.Warn("Failed to add scope_id column", zap.Error(err))
		return err
	}

	var globalScope, teamScope usermodel.Scope
	if err := database.DB.Where("code = ?", "global").First(&globalScope).Error; err != nil {
		logger.Error("Failed to find global scope", zap.Error(err))
		return err
	}
	if err := database.DB.Where("code = ?", "team").First(&teamScope).Error; err != nil {
		logger.Error("Failed to find team scope", zap.Error(err))
		return err
	}

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

	if err := database.DB.Exec("ALTER TABLE roles ALTER COLUMN scope_id SET NOT NULL").Error; err != nil {
		logger.Warn("Failed to set scope_id NOT NULL", zap.Error(err))
	}

	if err := database.DB.Exec("ALTER TABLE roles ADD CONSTRAINT fk_roles_scope FOREIGN KEY (scope_id) REFERENCES scopes(id)").Error; err != nil {
		logger.Warn("Failed to add foreign key constraint", zap.Error(err))
	}

	if err := database.DB.Exec("ALTER TABLE roles DROP COLUMN IF EXISTS scope").Error; err != nil {
		logger.Warn("Failed to drop scope column", zap.Error(err))
		return err
	}

	logger.Info("Role scope migration completed")
	return nil
}

func removeRoleScopeColumn(logger *zap.Logger) error {
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
		return nil
	}
	if !exists {
		logger.Info("scope column does not exist, skip removal")
		return nil
	}
	if err := database.DB.Exec("ALTER TABLE roles DROP COLUMN IF EXISTS scope").Error; err != nil {
		logger.Warn("Failed to drop scope column", zap.Error(err))
		return err
	}
	logger.Info("scope column removed from roles table")
	return nil
}

func removeRoleIsSystemColumn(logger *zap.Logger) error {
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
		return nil
	}
	if !exists {
		logger.Info("is_system column does not exist, skip removal")
		return nil
	}
	if err := database.DB.Exec("ALTER TABLE roles DROP COLUMN IF EXISTS is_system").Error; err != nil {
		logger.Warn("Failed to drop is_system column", zap.Error(err))
		return err
	}
	logger.Info("is_system column removed from roles table")
	return nil
}

func initDefaultMenus(logger *zap.Logger) error {
	var count int64
	if err := database.DB.Model(&usermodel.Menu{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		logger.Info("Menus already exist, skip seed")
		return nil
	}

	metaSuperAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}}
	metaSuperAdminAndAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER", "R_ADMIN"}}

	dashboard := usermodel.Menu{
		Path: "/dashboard", Name: "Dashboard", Component: "/index/index",
		Title: "menus.dashboard.title", Icon: "ri:pie-chart-line",
		SortOrder: 1, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}
	if err := database.DB.Create(&dashboard).Error; err != nil {
		return err
	}
	system := usermodel.Menu{
		Path: "/system", Name: "System", Component: "/index/index",
		Title: "menus.system.title", Icon: "ri:user-3-line",
		SortOrder: 2, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}
	if err := database.DB.Create(&system).Error; err != nil {
		return err
	}
	result := usermodel.Menu{
		Path: "/result", Name: "Result", Component: "/index/index",
		Title: "menus.result.title", Icon: "ri:checkbox-circle-line",
		SortOrder: 3, Meta: nil, IsSystem: true,
	}
	if err := database.DB.Create(&result).Error; err != nil {
		return err
	}
	exception := usermodel.Menu{
		Path: "/exception", Name: "Exception", Component: "/index/index",
		Title: "menus.exception.title", Icon: "ri:error-warning-line",
		SortOrder: 4, Meta: nil, IsSystem: true,
	}
	if err := database.DB.Create(&exception).Error; err != nil {
		return err
	}

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &dashboard.ID, Path: "console", Name: "Console", Component: "/dashboard/console",
		Title: "menus.dashboard.console", SortOrder: 1,
		Meta: usermodel.MetaJSON{"keepAlive": false, "fixedTab": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &dashboard.ID, Path: "user-center", Name: "UserCenter", Component: "/system/user-center",
		Title: "menus.system.userCenter", SortOrder: 2,
		Meta: usermodel.MetaJSON{"isHide": true, "keepAlive": true, "isHideTab": true}, IsSystem: true,
	}).Error

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "role", Name: "Role", Component: "/system/role",
		Title: "menus.system.role", SortOrder: 1, Meta: metaSuperAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "user", Name: "User", Component: "/system/user",
		Title: "menus.system.user", SortOrder: 2, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "pages-association", Name: "PageAssociation", Component: "/system/pages-association",
		Title: "menus.system.pagesAssociation", SortOrder: 3, Meta: metaSuperAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "menu", Name: "Menus", Component: "/system/menu",
		Title: "menus.system.menu", SortOrder: 4, IsSystem: true,
		Meta: usermodel.MetaJSON{
			"roles": []interface{}{"R_SUPER"},
			"keepAlive": true,
			"authList": []interface{}{
				map[string]interface{}{"title": "新增", "authMark": "add"},
				map[string]interface{}{"title": "编辑", "authMark": "edit"},
				map[string]interface{}{"title": "删除", "authMark": "delete"},
			},
		},
	}).Error

	team := usermodel.Menu{
		Path: "/team", Name: "TeamRoot", Component: "/index/index",
		Title: "menus.system.team", Icon: "ri:team-line",
		SortOrder: 5, Meta: metaSuperAdminAndAdmin, IsSystem: true,
	}
	if err := database.DB.Create(&team).Error; err != nil {
		return err
	}

	metaTeamAdminOnly := usermodel.MetaJSON{"roles": []interface{}{"R_ADMIN"}, "keepAlive": true}
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &team.ID, Path: "team-roles-permissions", Name: "TeamRolesAndPermissions", Component: "/system/team-roles-permissions",
		Title: "menus.system.teamRolesAndPermissions", SortOrder: 1, Meta: metaSuperAdmin, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &team.ID, Path: "management", Name: "TeamManagementRedirect", Component: "/team/team-members",
		Title: "menus.system.teamMembers", SortOrder: 2, Meta: usermodel.MetaJSON{"keepAlive": true, "roles": []interface{}{"R_ADMIN"}}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &team.ID, Path: "members", Name: "TeamMembers", Component: "/team/team-members",
		Title: "menus.system.teamMembers", SortOrder: 3, Meta: metaTeamAdminOnly, IsSystem: true,
	}).Error

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &result.ID, Path: "success", Name: "ResultSuccess", Component: "/result/success",
		Title: "menus.result.success", Icon: "ri:checkbox-circle-line", SortOrder: 1,
		Meta: usermodel.MetaJSON{"keepAlive": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &result.ID, Path: "fail", Name: "ResultFail", Component: "/result/fail",
		Title: "menus.result.fail", Icon: "ri:close-circle-line", SortOrder: 2,
		Meta: usermodel.MetaJSON{"keepAlive": true}, IsSystem: true,
	}).Error

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &exception.ID, Path: "403", Name: "Exception403", Component: "/exception/403",
		Title: "menus.exception.forbidden", SortOrder: 1,
		Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &exception.ID, Path: "404", Name: "Exception404", Component: "/exception/404",
		Title: "menus.exception.notFound", SortOrder: 2,
		Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}, IsSystem: true,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &exception.ID, Path: "500", Name: "Exception500", Component: "/exception/500",
		Title: "menus.exception.serverError", SortOrder: 3,
		Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}, IsSystem: true,
	}).Error

	logger.Info("Default menus seeded")
	return nil
}

func ensureTeamMenusIfMissing(logger *zap.Logger) error {
	var teamRoot usermodel.Menu
	err := database.DB.Where("path = ? AND name = ?", "/team", "TeamRoot").First(&teamRoot).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	metaSuperAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}}
	metaSuperAdminAndAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER", "R_ADMIN"}}
	metaTeamAdminOnly := usermodel.MetaJSON{"roles": []interface{}{"R_ADMIN"}, "keepAlive": true}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		teamRoot = usermodel.Menu{
			Path: "/team", Name: "TeamRoot", Component: "/index/index",
			Title: "menus.system.team", Icon: "ri:team-line",
			SortOrder: 5, Meta: metaSuperAdminAndAdmin, IsSystem: true,
		}
		if err := database.DB.Create(&teamRoot).Error; err != nil {
			return err
		}
		logger.Info("Team root menu created", zap.String("path", "/team"))
	}

	teamChildMenuSpecs := []struct {
		pathsToCheck []string
		path         string
		name         string
		component    string
		title        string
		sortOrder    int
		meta         usermodel.MetaJSON
		roleCodes    []string
	}{
		{[]string{"team-roles-permissions"}, "team-roles-permissions", "TeamRolesAndPermissions", "/system/team-roles-permissions", "menus.system.teamRolesAndPermissions", 1, metaSuperAdmin, []string{"admin"}},
		{[]string{"management"}, "management", "TeamManagementRedirect", "/team/team-members", "menus.system.teamMembers", 2, usermodel.MetaJSON{"keepAlive": true, "roles": []interface{}{"R_ADMIN"}}, []string{"admin", "team_admin"}},
		{[]string{"members"}, "members", "TeamMembers", "/team/team-members", "menus.system.teamMembers", 3, metaTeamAdminOnly, []string{"admin", "team_admin"}},
	}

	for _, spec := range teamChildMenuSpecs {
		var exist int64
		if err := database.DB.Model(&usermodel.Menu{}).Where("path IN ?", spec.pathsToCheck).Count(&exist).Error; err != nil {
			return err
		}
		if exist > 0 {
			continue
		}
		m := usermodel.Menu{
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
		database.DB.Model(&usermodel.Role{}).Where("code IN ?", spec.roleCodes).Pluck("id", &roleIDs)
		for _, roleID := range roleIDs {
			_ = database.DB.Create(&usermodel.RoleMenu{RoleID: roleID, MenuID: m.ID}).Error
		}
		logger.Info("Team menu ensured", zap.String("path", spec.path))
	}
	return nil
}

func ensureSystemMenuFlags(logger *zap.Logger) error {
	defaultPaths := []string{
		"/dashboard", "console", "user-center", "/system", "user", "role", "pages-association", "menu",
		"/team", "team-roles-permissions", "management", "members",
		"/result", "success", "fail", "/exception", "403", "404", "500",
	}
	res := database.DB.Model(&usermodel.Menu{}).Where("path IN ?", defaultPaths).Update("is_system", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		logger.Info("System menu flags updated", zap.Int64("rows", res.RowsAffected))
	}
	return nil
}

func initDefaultRoleMenus(logger *zap.Logger) error {
	var count int64
	if err := database.DB.Model(&usermodel.RoleMenu{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		logger.Info("Role-menus already exist, skip seed")
		return nil
	}

	var roles []usermodel.Role
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

	var menus []usermodel.Menu
	if err := database.DB.Find(&menus).Error; err != nil {
		return err
	}

	roleMenus := make([]usermodel.RoleMenu, 0)
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
			roleMenus = append(roleMenus, usermodel.RoleMenu{RoleID: adminID, MenuID: m.ID})
		}
		if addTeamAdmin && hasTeamAdmin {
			roleMenus = append(roleMenus, usermodel.RoleMenu{RoleID: teamAdminID, MenuID: m.ID})
		}
		if addTeamMember && hasTeamMember {
			roleMenus = append(roleMenus, usermodel.RoleMenu{RoleID: teamMemberID, MenuID: m.ID})
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
