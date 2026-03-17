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
	menuService := NewMenuService(menuRepo, m.logger)
	userRepo := user.NewUserRepository(m.db)
	roleMenuRepo := user.NewRoleMenuRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	tenantMemberRepo := user.NewTenantMemberRepository(m.db)
	authzService := authorization.NewService(m.db, m.logger)
	menuHandler := NewMenuHandler(menuService, userRepo, roleMenuRepo, userRoleRepo, tenantMemberRepo, authzService, m.logger)

	menus := rg.Group("/menus")
	reg := apiregistry.NewRegistrar(menus, "menu")
	{
		reg.GET("/tree", &apiregistry.RouteMeta{Summary: "获取菜单树", ResourceCode: "menu", ActionCode: "list", ScopeCode: "global"}, menuHandler.GetTree)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建菜单", ResourceCode: "menu", ActionCode: "create", ScopeCode: "global"}, authzService.RequireAction("menu", "create"), menuHandler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新菜单", ResourceCode: "menu", ActionCode: "update", ScopeCode: "global"}, authzService.RequireAction("menu", "update"), menuHandler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除菜单", ResourceCode: "menu", ActionCode: "delete", ScopeCode: "global"}, authzService.RequireAction("menu", "delete"), menuHandler.Delete)

		// 菜单备份相关路由
		backups := menus.Group("/backups")
		backupReg := apiregistry.NewRegistrar(backups, "menu_backup")
		{
			backupReg.POST("", &apiregistry.RouteMeta{Summary: "创建菜单备份", ResourceCode: "menu_backup", ActionCode: "create", ScopeCode: "global"}, authzService.RequireAction("menu_backup", "create"), menuHandler.CreateBackup)
			backupReg.GET("", &apiregistry.RouteMeta{Summary: "获取菜单备份列表", ResourceCode: "menu_backup", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("menu_backup", "list"), menuHandler.ListBackups)
			backupReg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除菜单备份", ResourceCode: "menu_backup", ActionCode: "delete", ScopeCode: "global"}, authzService.RequireAction("menu_backup", "delete"), menuHandler.DeleteBackup)
			backupReg.POST("/:id/restore", &apiregistry.RouteMeta{Summary: "恢复菜单备份", ResourceCode: "menu_backup", ActionCode: "restore", ScopeCode: "global"}, authzService.RequireAction("menu_backup", "restore"), menuHandler.RestoreBackup)
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
