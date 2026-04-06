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
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
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
	packageActionRepo := user.NewFeaturePackageKeyRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	packageBundleRepo := user.NewFeaturePackageBundleRepository(m.db)
	roleHiddenMenuRepo := user.NewRoleHiddenMenuRepository(m.db)
	roleDisabledActionRepo := user.NewRoleDisabledActionRepository(m.db)
	roleDataRepo := user.NewRoleDataPermissionRepository(m.db)
	actionRepo := user.NewPermissionKeyRepository(m.db)
	boundaryService := collaborationworkspaceboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	roleService := NewRoleService(
		m.db,
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
	roleHandler := NewRoleHandler(roleService, userRepo, actionRepo, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	roles := rg.Group("/roles")
	reg := apiregistry.NewRegistrar(roles, "role")
	{
		reg.GETProtected("", reg.Meta("获取角色列表").BindGroup("role").BindPermissionKey("system.role.manage").Build(), "system.role.manage", authzService.RequireAction, roleHandler.List)
		reg.GETProtected("/options", reg.Meta("获取角色候选").BindGroup("role").BindPermissionKey("system.role.manage").Build(), "system.role.manage", authzService.RequireAction, roleHandler.ListOptions)
		reg.GETProtected("/:id", reg.Meta("获取角色详情").BindGroup("role").BindPermissionKey("system.role.manage").Build(), "system.role.manage", authzService.RequireAction, roleHandler.Get)
		reg.GETAction("/:id/packages", "获取角色功能包", "feature_package.assign_collaboration_workspace", authzService.RequireAction, roleHandler.GetRolePackages)
		reg.PUTAction("/:id/packages", "配置角色功能包", "feature_package.assign_collaboration_workspace", authzService.RequireAction, roleHandler.SetRolePackages)
		reg.GETProtected("/:id/menus", reg.Meta("获取角色菜单权限").BindGroup("role").BindPermissionKey("system.role.assign_menu").Build(), "system.role.assign_menu", authzService.RequireAction, roleHandler.GetRoleMenus)
		reg.PUTProtected("/:id/menus", reg.Meta("配置角色菜单权限").BindGroup("role").BindPermissionKey("system.role.assign_menu").Build(), "system.role.assign_menu", authzService.RequireAction, roleHandler.SetRoleMenus)
		reg.GETProtected("/:id/actions", reg.Meta("获取角色功能权限").BindGroup("role").BindPermissionKey("system.role.assign_action").Build(), "system.role.assign_action", authzService.RequireAction, roleHandler.GetRoleKeys)
		reg.PUTProtected("/:id/actions", reg.Meta("配置角色功能权限").BindGroup("role").BindPermissionKey("system.role.assign_action").Build(), "system.role.assign_action", authzService.RequireAction, roleHandler.SetRoleKeys)
		reg.GETProtected("/:id/data-permissions", reg.Meta("获取角色数据权限").BindGroup("role").BindPermissionKey("system.role.assign_data").Build(), "system.role.assign_data", authzService.RequireAction, roleHandler.GetRoleDataPermissions)
		reg.PUTProtected("/:id/data-permissions", reg.Meta("配置角色数据权限").BindGroup("role").BindPermissionKey("system.role.assign_data").Build(), "system.role.assign_data", authzService.RequireAction, roleHandler.SetRoleDataPermissions)
		reg.POSTProtected("", reg.Meta("创建角色").BindGroup("role").BindPermissionKey("system.role.manage").Build(), "system.role.manage", authzService.RequireAction, roleHandler.Create)
		reg.PUTProtected("/:id", reg.Meta("更新角色").BindGroup("role").BindPermissionKey("system.role.manage").Build(), "system.role.manage", authzService.RequireAction, roleHandler.Update)
		reg.DELETEProtected("/:id", reg.Meta("删除角色").BindGroup("role").BindPermissionKey("system.role.manage").Build(), "system.role.manage", authzService.RequireAction, roleHandler.Delete)
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
