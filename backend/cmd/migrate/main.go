package main

import (
	"errors"
	"fmt"
	"log"

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

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}

	logger.Info("Database migration completed successfully!")

	// 初始化默认作用域
	if err := initDefaultScopes(logger); err != nil {
		logger.Warn("initDefaultScopes failed", zap.Error(err))
	}

	// 初始化默认角色
	if err := initDefaultRoles(logger); err != nil {
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

	// 初始化默认菜单
	if err := initDefaultMenus(logger); err != nil {
		logger.Warn("Failed to initialize default menus", zap.Error(err))
	} else {
		logger.Info("Default menus initialized successfully")
	}

	// 初始化默认角色菜单关联
	if err := initDefaultRoleMenus(logger); err != nil {
		logger.Warn("Failed to initialize role-menus", zap.Error(err))
	} else {
		logger.Info("Role-menus initialized successfully")
	}

	fmt.Println("✅ Migration completed successfully!")
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
	if err := database.DB.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		logger.Error("Failed to find admin role", zap.Error(err))
		return err
	}

	var userRole usermodel.UserRole
	result := database.DB.Where("user_id = ? AND role_id = ? AND tenant_id IS NULL", userID, adminRole.ID).First(&userRole)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			userRole = usermodel.UserRole{
				UserID:   userID,
				RoleID:   adminRole.ID,
				TenantID: nil,
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
		SortOrder: 1, Meta: metaSuperAdminAndAdmin,
	}
	if err := database.DB.Create(&dashboard).Error; err != nil {
		return err
	}
	system := usermodel.Menu{
		Path: "/system", Name: "System", Component: "/index/index",
		Title: "menus.system.title", Icon: "ri:user-3-line",
		SortOrder: 2, Meta: metaSuperAdminAndAdmin,
	}
	if err := database.DB.Create(&system).Error; err != nil {
		return err
	}
	result := usermodel.Menu{
		Path: "/result", Name: "Result", Component: "/index/index",
		Title: "menus.result.title", Icon: "ri:checkbox-circle-line",
		SortOrder: 3, Meta: nil,
	}
	if err := database.DB.Create(&result).Error; err != nil {
		return err
	}
	exception := usermodel.Menu{
		Path: "/exception", Name: "Exception", Component: "/index/index",
		Title: "menus.exception.title", Icon: "ri:error-warning-line",
		SortOrder: 4, Meta: nil,
	}
	if err := database.DB.Create(&exception).Error; err != nil {
		return err
	}

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &dashboard.ID, Path: "console", Name: "Console", Component: "/dashboard/console",
		Title: "menus.dashboard.console", SortOrder: 1,
		Meta: usermodel.MetaJSON{"keepAlive": false, "fixedTab": true},
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &dashboard.ID, Path: "user-center", Name: "UserCenter", Component: "/system/user-center",
		Title: "menus.system.userCenter", SortOrder: 2,
		Meta: usermodel.MetaJSON{"isHide": true, "keepAlive": true, "isHideTab": true},
	}).Error

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "role", Name: "Role", Component: "/system/role",
		Title: "menus.system.role", SortOrder: 1, Meta: metaSuperAdmin,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "user", Name: "User", Component: "/system/user",
		Title: "menus.system.user", SortOrder: 2, Meta: metaSuperAdminAndAdmin,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "pages-association", Name: "PageAssociation", Component: "/system/pages-association",
		Title: "menus.system.pagesAssociation", SortOrder: 3, Meta: metaSuperAdmin,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &system.ID, Path: "menu", Name: "Menus", Component: "/system/menu",
		Title: "menus.system.menu", SortOrder: 4,
		Meta: usermodel.MetaJSON{
			"roles":     []interface{}{"R_SUPER"},
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
		SortOrder: 5, Meta: metaSuperAdminAndAdmin,
	}
	if err := database.DB.Create(&team).Error; err != nil {
		return err
	}

	metaTeamAdminOnly := usermodel.MetaJSON{"roles": []interface{}{"R_ADMIN"}, "keepAlive": true}
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &team.ID, Path: "team-roles-permissions", Name: "TeamRolesAndPermissions", Component: "/system/team-roles-permissions",
		Title: "menus.system.teamRolesAndPermissions", SortOrder: 1, Meta: metaSuperAdmin,
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &team.ID, Path: "management", Name: "TeamManagementRedirect", Component: "/team/team-members",
		Title: "menus.system.teamMembers", SortOrder: 2, Meta: usermodel.MetaJSON{"keepAlive": true, "roles": []interface{}{"R_ADMIN"}},
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &team.ID, Path: "members", Name: "TeamMembers", Component: "/team/team-members",
		Title: "menus.system.teamMembers", SortOrder: 3, Meta: metaTeamAdminOnly,
	}).Error

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &result.ID, Path: "success", Name: "ResultSuccess", Component: "/result/success",
		Title: "menus.result.success", Icon: "ri:checkbox-circle-line", SortOrder: 1,
		Meta: usermodel.MetaJSON{"keepAlive": true},
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &result.ID, Path: "fail", Name: "ResultFail", Component: "/result/fail",
		Title: "menus.result.fail", Icon: "ri:close-circle-line", SortOrder: 2,
		Meta: usermodel.MetaJSON{"keepAlive": true},
	}).Error

	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &exception.ID, Path: "403", Name: "Exception403", Component: "/exception/403",
		Title: "menus.exception.forbidden", SortOrder: 1,
		Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true},
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &exception.ID, Path: "404", Name: "Exception404", Component: "/exception/404",
		Title: "menus.exception.notFound", SortOrder: 2,
		Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true},
	}).Error
	_ = database.DB.Create(&usermodel.Menu{
		ParentID: &exception.ID, Path: "500", Name: "Exception500", Component: "/exception/500",
		Title: "menus.exception.serverError", SortOrder: 3,
		Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true},
	}).Error

	logger.Info("Default menus seeded")
	return nil
}

func initDefaultRoleMenus(logger *zap.Logger) error {
	// 先清空所有角色菜单关联
	if err := database.DB.Exec("DELETE FROM role_menus").Error; err != nil {
		logger.Error("Failed to delete role-menus", zap.Error(err))
		return err
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
