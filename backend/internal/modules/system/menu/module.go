package menu

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

type MenuModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewMenuModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *MenuModule {
	return &MenuModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *MenuModule) Init() error {
	m.logger.Info("Initializing Menu module")
	return nil
}

func (m *MenuModule) RegisterRoutes(rg *gin.RouterGroup) {
	menuRepo := user.NewMenuRepository(m.db)
	userRepo := user.NewUserRepository(m.db)
	roleRepo := user.NewRoleRepository(m.db)
	roleMenuRepo := user.NewRoleMenuRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	rolePackageRepo := user.NewRoleFeaturePackageRepository(m.db)
	userPackageRepo := user.NewUserFeaturePackageRepository(m.db)
	packageRepo := user.NewFeaturePackageRepository(m.db)
	packageBundleRepo := user.NewFeaturePackageBundleRepository(m.db)
	roleHiddenMenuRepo := user.NewRoleHiddenMenuRepository(m.db)
	userHiddenMenuRepo := user.NewUserHiddenMenuRepository(m.db)
	tenantMemberRepo := user.NewTenantMemberRepository(m.db)
	teamPackageRepo := user.NewTeamFeaturePackageRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService)
	menuService := NewMenuService(menuRepo, refresher, m.logger)
	authzService := authorization.NewService(m.db, m.logger)
	permissionService := user.NewPermissionService(userRepo, roleRepo, userRoleRepo, roleMenuRepo, rolePackageRepo, userPackageRepo, packageRepo, packageMenuRepo, packageBundleRepo, roleHiddenMenuRepo, userHiddenMenuRepo, menuRepo, boundaryService, platformService)
	menuHandler := NewMenuHandler(menuService, permissionService, userRepo, roleRepo, roleMenuRepo, userRoleRepo, tenantMemberRepo, teamPackageRepo, packageMenuRepo, boundaryService, authzService, m.logger)

	menus := rg.Group("/menus")
	reg := apiregistry.NewRegistrar(menus, "menu")
	{
		reg.GET("/tree", &apiregistry.RouteMeta{Summary: "获取菜单树", ResourceCode: "menu", ActionCode: "list"}, menuHandler.GetTree)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建菜单", ResourceCode: "menu", ActionCode: "create"}, authzService.RequireAction("system.menu.manage"), menuHandler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新菜单", ResourceCode: "menu", ActionCode: "update"}, authzService.RequireAction("system.menu.manage"), menuHandler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除菜单", ResourceCode: "menu", ActionCode: "delete"}, authzService.RequireAction("system.menu.manage"), menuHandler.Delete)

		// 菜单备份相关路由
		backups := menus.Group("/backups")
		backupReg := apiregistry.NewRegistrar(backups, "menu_backup")
		{
			backupReg.POST("", &apiregistry.RouteMeta{Summary: "创建菜单备份", ResourceCode: "menu_backup", ActionCode: "create"}, authzService.RequireAction("system.menu.backup"), menuHandler.CreateBackup)
			backupReg.GET("", &apiregistry.RouteMeta{Summary: "获取菜单备份列表", ResourceCode: "menu_backup", ActionCode: "list"}, authzService.RequireAction("system.menu.backup"), menuHandler.ListBackups)
			backupReg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除菜单备份", ResourceCode: "menu_backup", ActionCode: "delete"}, authzService.RequireAction("system.menu.backup"), menuHandler.DeleteBackup)
			backupReg.POST("/:id/restore", &apiregistry.RouteMeta{Summary: "恢复菜单备份", ResourceCode: "menu_backup", ActionCode: "restore"}, authzService.RequireAction("system.menu.backup"), menuHandler.RestoreBackup)
		}
	}
}

func init() {
	module.GetRegistry().Register(&menuModuleWrapper{})
}

type menuModuleWrapper struct{}

func (w *menuModuleWrapper) Init() error {
	return nil
}

func (w *menuModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
