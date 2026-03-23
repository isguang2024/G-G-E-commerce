package role

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

type RoleModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewRoleModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *RoleModule {
	return &RoleModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *RoleModule) Init() error {
	m.logger.Info("Initializing Role module")
	return nil
}

func (m *RoleModule) RegisterRoutes(rg *gin.RouterGroup) {
	roleRepo := user.NewRoleRepository(m.db)
	roleMenuRepo := user.NewRoleMenuRepository(m.db)
	roleActionRepo := user.NewRoleActionPermissionRepository(m.db)
	roleDataRepo := user.NewRoleDataPermissionRepository(m.db)
	actionRepo := user.NewPermissionActionRepository(m.db)
	roleService := NewRoleService(roleRepo, roleMenuRepo, roleActionRepo, roleDataRepo, actionRepo, m.logger)
	userRepo := user.NewUserRepository(m.db)
	roleHandler := NewRoleHandler(roleService, userRepo, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	roles := rg.Group("/roles")
	reg := apiregistry.NewRegistrar(roles, "role")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取角色列表", ResourceCode: "role", ActionCode: "list"}, authzService.RequireAction("system.role.manage"), roleHandler.List)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取角色详情", ResourceCode: "role", ActionCode: "get"}, authzService.RequireAction("system.role.manage"), roleHandler.Get)
		reg.GET("/:id/menus", &apiregistry.RouteMeta{Summary: "获取角色菜单权限", ResourceCode: "role", ActionCode: "assign_menu"}, authzService.RequireAction("system.role.assign_menu"), roleHandler.GetRoleMenus)
		reg.PUT("/:id/menus", &apiregistry.RouteMeta{Summary: "配置角色菜单权限", ResourceCode: "role", ActionCode: "assign_menu"}, authzService.RequireAction("system.role.assign_menu"), roleHandler.SetRoleMenus)
		reg.GET("/:id/actions", &apiregistry.RouteMeta{Summary: "获取角色功能权限", ResourceCode: "role", ActionCode: "assign_action"}, authzService.RequireAction("system.role.assign_action"), roleHandler.GetRoleActions)
		reg.PUT("/:id/actions", &apiregistry.RouteMeta{Summary: "配置角色功能权限", ResourceCode: "role", ActionCode: "assign_action"}, authzService.RequireAction("system.role.assign_action"), roleHandler.SetRoleActions)
		reg.GET("/:id/data-permissions", &apiregistry.RouteMeta{Summary: "获取角色数据权限", ResourceCode: "role", ActionCode: "assign_data"}, authzService.RequireAction("system.role.assign_data"), roleHandler.GetRoleDataPermissions)
		reg.PUT("/:id/data-permissions", &apiregistry.RouteMeta{Summary: "配置角色数据权限", ResourceCode: "role", ActionCode: "assign_data"}, authzService.RequireAction("system.role.assign_data"), roleHandler.SetRoleDataPermissions)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建角色", ResourceCode: "role", ActionCode: "create"}, authzService.RequireAction("system.role.manage"), roleHandler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新角色", ResourceCode: "role", ActionCode: "update"}, authzService.RequireAction("system.role.manage"), roleHandler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除角色", ResourceCode: "role", ActionCode: "delete"}, authzService.RequireAction("system.role.manage"), roleHandler.Delete)
	}
}

func init() {
	module.GetRegistry().Register(&roleModuleWrapper{})
}

type roleModuleWrapper struct{}

func (w *roleModuleWrapper) Init() error {
	return nil
}

func (w *roleModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
