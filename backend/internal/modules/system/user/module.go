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
	userRoleRepo := NewUserRoleRepository(m.db)
	userPackageRepo := NewUserFeaturePackageRepository(m.db)
	packageRepo := NewFeaturePackageRepository(m.db)
	userHiddenMenuRepo := NewUserHiddenMenuRepository(m.db)
	keyRepo := NewPermissionKeyRepository(m.db)
	tenantMemberRepo := NewTenantMemberRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	userService := NewUserService(m.db, userRepo, roleRepo, refresher, m.logger)
	authzService := authorization.NewService(m.db, m.logger)
	userHandler := NewUserHandler(m.db, userService, packageRepo, keyRepo, platformService, boundaryService, roleRepo, authzService, userRoleRepo, tenantMemberRepo, userPackageRepo, userHiddenMenuRepo, menuRepo, refresher, m.logger)

	users := rg.Group("/users")
	reg := apiregistry.NewRegistrar(users, "user")
	{
		reg.GETProtected("", reg.Meta("获取用户列表").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.List)
		reg.GETProtected("/:id", reg.Meta("获取用户详情").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.Get)
		reg.GETProtected("/:id/collaboration-workspaces", reg.Meta("获取用户所在协作空间").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetTeams)
		reg.GETAction("/:id/packages", "获取用户功能包", "platform.package.assign", authzService.RequireAction, userHandler.GetPackages)
		reg.PUTAction("/:id/packages", "配置用户功能包", "platform.package.assign", authzService.RequireAction, userHandler.SetPackages)
		reg.GETProtected("/:id/menus", reg.Meta("获取用户菜单裁剪").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetMenus)
		reg.PUTProtected("/:id/menus", reg.Meta("配置用户菜单裁剪").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.SetMenus)
		reg.GETProtected("/:id/permissions", reg.Meta("获取用户菜单权限").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetPermissions)
		reg.GETProtected("/:id/permission-diagnosis", reg.Meta("获取用户权限诊断").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetPermissionDiagnosis)
		reg.POSTProtected("/:id/permission-refresh", reg.Meta("刷新用户权限快照").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.RefreshPermissionSnapshot)
		reg.POSTProtected("", reg.Meta("创建用户").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.Create)
		reg.PUTProtected("/:id", reg.Meta("更新用户").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.Update)
		reg.DELETEProtected("/:id", reg.Meta("删除用户").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.Delete)
		reg.POSTProtected("/:id/roles", reg.Meta("分配用户角色").BindGroup("user").BindPermissionKey("system.user.assign_role").Build(), "system.user.assign_role", authzService.RequireAction, userHandler.AssignRoles)
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
