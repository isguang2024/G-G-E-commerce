package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
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
	collaborationWorkspaceMemberRepo := NewCollaborationWorkspaceMemberRepository(m.db)
	boundaryService := collaborationworkspaceboundary.NewService(m.db)
	personalWorkspaceAccessService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, personalWorkspaceAccessService, roleSnapshotService)
	userService := NewUserService(m.db, userRepo, roleRepo, refresher, m.logger)
	authzService := authorization.NewService(m.db, m.logger)
	userHandler := NewUserHandler(m.db, userService, packageRepo, keyRepo, personalWorkspaceAccessService, boundaryService, roleRepo, authzService, userRoleRepo, collaborationWorkspaceMemberRepo, userPackageRepo, userHiddenMenuRepo, menuRepo, refresher, m.logger)

	users := rg.Group("/users")
	reg := apiregistry.NewRegistrar(users, "user")
	{
		// Phase 4 slice 5: list/get/create/update/delete/assignRoles migrated
		// to ogen handlers in internal/api/handlers/user.go. Only not-yet-migrated
		// secondary routes remain wired here.
		reg.GETProtected("/:id/collaboration-workspaces", reg.Meta("获取用户所在协作空间").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetCollaborationWorkspaces)
		reg.GETAction("/:id/packages", "获取用户功能包", "feature_package.assign_collaboration_workspace", authzService.RequireAction, userHandler.GetPackages)
		reg.PUTAction("/:id/packages", "配置用户功能包", "feature_package.assign_collaboration_workspace", authzService.RequireAction, userHandler.SetPackages)
		reg.GETProtected("/:id/menus", reg.Meta("获取用户菜单裁剪").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetMenus)
		reg.PUTProtected("/:id/menus", reg.Meta("配置用户菜单裁剪").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.SetMenus)
		reg.GETProtected("/:id/permissions", reg.Meta("获取用户菜单权限").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetPermissions)
		reg.GETProtected("/:id/permission-diagnosis", reg.Meta("获取用户权限诊断").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.GetPermissionDiagnosis)
		reg.POSTProtected("/:id/permission-refresh", reg.Meta("刷新用户权限快照").BindGroup("user").BindPermissionKey("system.user.manage").Build(), "system.user.manage", authzService.RequireAction, userHandler.RefreshPermissionSnapshot)
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
