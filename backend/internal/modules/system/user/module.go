package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

type UserModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewUserModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *UserModule {
	return &UserModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *UserModule) Init() error {
	m.logger.Info("Initializing User module")
	return nil
}

func (m *UserModule) RegisterRoutes(rg *gin.RouterGroup) {
	userRepo := NewUserRepository(m.db)
	roleRepo := NewRoleRepository(m.db)
	menuRepo := NewMenuRepository(m.db)
	roleMenuRepo := NewRoleMenuRepository(m.db)
	userRoleRepo := NewUserRoleRepository(m.db)
	userPackageRepo := NewUserFeaturePackageRepository(m.db)
	packageRepo := NewFeaturePackageRepository(m.db)
	userHiddenMenuRepo := NewUserHiddenMenuRepository(m.db)
	actionRepo := NewPermissionActionRepository(m.db)
	userActionRepo := NewUserActionPermissionRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	userService := NewUserService(userRepo, roleRepo, m.logger)
	userHandler := NewUserHandler(userService, actionRepo, packageRepo, platformService, boundaryService, roleRepo, roleMenuRepo, userRoleRepo, userPackageRepo, userHiddenMenuRepo, userActionRepo, menuRepo, refresher, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	users := rg.Group("/users")
	reg := apiregistry.NewRegistrar(users, "user")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取用户列表", ResourceCode: "user", ActionCode: "list"}, authzService.RequireAction("system.user.manage"), userHandler.List)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取用户详情", ResourceCode: "user", ActionCode: "get"}, authzService.RequireAction("system.user.manage"), userHandler.Get)
		reg.GET("/:id/packages", &apiregistry.RouteMeta{Summary: "获取用户功能包"}, authzService.RequireAction("platform.package.assign"), userHandler.GetPackages)
		reg.PUT("/:id/packages", &apiregistry.RouteMeta{Summary: "配置用户功能包"}, authzService.RequireAction("platform.package.assign"), userHandler.SetPackages)
		reg.GET("/:id/menus", &apiregistry.RouteMeta{Summary: "获取用户菜单裁剪", ResourceCode: "user", ActionCode: "get"}, authzService.RequireAction("system.user.manage"), userHandler.GetMenus)
		reg.PUT("/:id/menus", &apiregistry.RouteMeta{Summary: "配置用户菜单裁剪", ResourceCode: "user", ActionCode: "assign_menu"}, authzService.RequireAction("system.user.manage"), userHandler.SetMenus)
		reg.GET("/:id/permissions", &apiregistry.RouteMeta{Summary: "获取用户菜单权限", ResourceCode: "user", ActionCode: "get"}, authzService.RequireAction("system.user.manage"), userHandler.GetPermissions)
		reg.GET("/:id/actions", &apiregistry.RouteMeta{Summary: "获取用户权限例外（兼容）", ResourceCode: "user", ActionCode: "assign_action"}, authzService.RequireAction("system.user.assign_action"), userHandler.GetActions)
		reg.PUT("/:id/actions", &apiregistry.RouteMeta{Summary: "配置用户权限例外（兼容）", ResourceCode: "user", ActionCode: "assign_action"}, authzService.RequireAction("system.user.assign_action"), userHandler.SetActions)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建用户", ResourceCode: "user", ActionCode: "create"}, authzService.RequireAction("system.user.manage"), userHandler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新用户", ResourceCode: "user", ActionCode: "update"}, authzService.RequireAction("system.user.manage"), userHandler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除用户", ResourceCode: "user", ActionCode: "delete"}, authzService.RequireAction("system.user.manage"), userHandler.Delete)
		reg.POST("/:id/roles", &apiregistry.RouteMeta{Summary: "分配用户角色", ResourceCode: "user", ActionCode: "assign_role"}, authzService.RequireAction("system.user.assign_role"), userHandler.AssignRoles)
	}
}

func init() {
	module.GetRegistry().Register(&userModuleWrapper{})
}

type userModuleWrapper struct{}

func (w *userModuleWrapper) Init() error {
	return nil
}

func (w *userModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
