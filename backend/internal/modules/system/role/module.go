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
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
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
	rolePackageRepo := user.NewRoleFeaturePackageRepository(m.db)
	featurePkgRepo := user.NewFeaturePackageRepository(m.db)
	packageActionRepo := user.NewFeaturePackageActionRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	packageBundleRepo := user.NewFeaturePackageBundleRepository(m.db)
	roleHiddenMenuRepo := user.NewRoleHiddenMenuRepository(m.db)
	roleDisabledActionRepo := user.NewRoleDisabledActionRepository(m.db)
	roleDataRepo := user.NewRoleDataPermissionRepository(m.db)
	actionRepo := user.NewPermissionActionRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	roleService := NewRoleService(
		roleRepo,
		rolePackageRepo,
		featurePkgRepo,
		packageActionRepo,
		packageMenuRepo,
		packageBundleRepo,
		roleHiddenMenuRepo,
		roleDisabledActionRepo,
		roleDataRepo,
		actionRepo,
		roleSnapshotService,
		refresher,
		m.logger,
	)
	userRepo := user.NewUserRepository(m.db)
	roleHandler := NewRoleHandler(roleService, userRepo, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	roles := rg.Group("/roles")
	reg := apiregistry.NewRegistrar(roles, "role")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取角色列表", ResourceCode: "role", ActionCode: "list"}, authzService.RequireAction("system.role.manage"), roleHandler.List)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取角色详情", ResourceCode: "role", ActionCode: "get"}, authzService.RequireAction("system.role.manage"), roleHandler.Get)
		reg.GET("/:id/packages", &apiregistry.RouteMeta{Summary: "获取角色功能包"}, authzService.RequireAction("platform.package.assign"), roleHandler.GetRolePackages)
		reg.PUT("/:id/packages", &apiregistry.RouteMeta{Summary: "配置角色功能包"}, authzService.RequireAction("platform.package.assign"), roleHandler.SetRolePackages)
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
