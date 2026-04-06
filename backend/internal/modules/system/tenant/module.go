package tenant

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	workspacepkg "github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

type TenantModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewTenantModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *TenantModule {
	return &TenantModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *TenantModule) Init() error {
	m.logger.Info("Initializing Tenant module")
	return nil
}

func (m *TenantModule) RegisterRoutes(rg *gin.RouterGroup) {
	tenantRepo := user.NewTenantRepository(m.db)
	tenantMemberRepo := user.NewTenantMemberRepository(m.db)
	userRepo := user.NewUserRepository(m.db)
	roleRepo := user.NewRoleRepository(m.db)
	roleHiddenMenuRepo := user.NewRoleHiddenMenuRepository(m.db)
	roleDisabledActionRepo := user.NewRoleDisabledActionRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	actionRepo := user.NewPermissionKeyRepository(m.db)
	blockedMenuRepo := user.NewCollaborationWorkspaceBlockedMenuRepository(m.db)
	blockedActionRepo := user.NewCollaborationWorkspaceBlockedActionRepository(m.db)
	teamPackageRepo := user.NewCollaborationWorkspaceFeaturePackageRepository(m.db)
	rolePackageRepo := user.NewRoleFeaturePackageRepository(m.db)
	featurePkgRepo := user.NewFeaturePackageRepository(m.db)
	packageActionRepo := user.NewFeaturePackageKeyRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	workspaceService := workspacepkg.NewService(m.db, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	tenantService := NewTenantService(m.db, tenantRepo, tenantMemberRepo, userRepo, roleRepo, userRoleRepo, refresher, m.logger)
	tenantHandler := NewTenantHandler(
		tenantService,
		tenantMemberRepo,
		userRepo,
		roleRepo,
		roleHiddenMenuRepo,
		roleDisabledActionRepo,
		userRoleRepo,
		actionRepo,
		blockedMenuRepo,
		blockedActionRepo,
		teamPackageRepo,
		rolePackageRepo,
		featurePkgRepo,
		packageActionRepo,
		packageMenuRepo,
		boundaryService,
		refresher,
		workspaceService,
		authzService,
		m.logger,
	)

	collaborationWorkspaces := rg.Group("/collaboration-workspaces")
	reg := apiregistry.NewRegistrar(collaborationWorkspaces, "collaboration_workspace")
	{
		reg.GET("/mine", &apiregistry.RouteMeta{Summary: "获取我的协作空间列表"}, tenantHandler.ListMyTeams)
		reg.GET("/current", &apiregistry.RouteMeta{Summary: "获取当前协作空间详情"}, tenantHandler.GetMyTeam)
		reg.GET("/current/members", &apiregistry.RouteMeta{Summary: "获取当前协作空间成员列表"}, tenantHandler.ListMyMembers)
		reg.POSTAction("/current/members", "添加当前协作空间成员", "collaboration_workspace.member.manage", authzService.RequireAction, tenantHandler.AddMyMember)
		reg.DELETEAction("/current/members/:userId", "移除当前协作空间成员", "collaboration_workspace.member.manage", authzService.RequireAction, tenantHandler.RemoveMyMember)
		reg.PUTAction("/current/members/:userId/role", "更新当前协作空间成员身份", "collaboration_workspace.member.manage", authzService.RequireAction, tenantHandler.UpdateMyMemberRole)
		reg.GETAction("/current/members/:userId/roles", "获取当前协作空间成员角色", "collaboration_workspace.member.manage", authzService.RequireAction, tenantHandler.GetMyCollaborationWorkspaceMemberRoles)
		reg.PUTAction("/current/members/:userId/roles", "配置当前协作空间成员角色", "collaboration_workspace.member.manage", authzService.RequireAction, tenantHandler.SetMyCollaborationWorkspaceMemberRoles)
		reg.GETAction("/current/roles", "获取当前协作空间可分配角色", "collaboration_workspace.member.manage", authzService.RequireAction, tenantHandler.ListMyTeamRoles)
		reg.POSTAction("/current/roles", "创建当前协作空间角色", "collaboration_workspace.member.manage", authzService.RequireAction, tenantHandler.CreateMyTeamRole)
		reg.GETAction("/current/boundary/roles", "获取当前协作空间边界可见角色", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.ListMyTeamRoles)
		reg.POSTAction("/current/boundary/roles", "创建当前协作空间角色(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.CreateMyTeamRole)
		reg.PUTAction("/current/boundary/roles/:roleId", "更新当前协作空间角色(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.UpdateMyTeamRole)
		reg.DELETEAction("/current/boundary/roles/:roleId", "删除当前协作空间角色(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.DeleteMyTeamRole)
		reg.GETAction("/current/boundary/roles/:roleId/packages", "获取当前协作空间角色功能包(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamRolePackages)
		reg.PUTAction("/current/boundary/roles/:roleId/packages", "配置当前协作空间角色功能包(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.SetMyTeamRolePackages)
		reg.GETAction("/current/boundary/roles/:roleId/menus", "获取当前协作空间角色菜单权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamRoleMenus)
		reg.PUTAction("/current/boundary/roles/:roleId/menus", "配置当前协作空间角色菜单权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.SetMyTeamRoleMenus)
		reg.GETAction("/current/boundary/roles/:roleId/actions", "获取当前协作空间角色功能权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamRoleActions)
		reg.PUTAction("/current/boundary/roles/:roleId/actions", "配置当前协作空间角色功能权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.SetMyTeamRoleActions)
		reg.GETAction("/current/boundary/packages", "获取当前协作空间已开通功能包(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamBoundaryPackages)
		reg.GETAction("/current/menus", "获取当前协作空间菜单边界", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamMenus)
		reg.GETAction("/current/menu-origins", "获取当前协作空间菜单来源", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamMenuOrigins)
		reg.GETAction("/current/actions", "获取当前协作空间功能权限边界", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamActions)
		reg.GETAction("/current/action-origins", "获取当前协作空间功能权限来源", "collaboration_workspace.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamActionOrigins)

		reg.GETAction("", "获取协作空间列表", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.List)
		reg.GETAction("/:id/roles", "获取协作空间可分配角色", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.ListTenantRoles)
		reg.GETAction("/:id", "获取协作空间详情", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.Get)
		reg.POSTAction("", "创建协作空间", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.Create)
		reg.PUTAction("/:id", "更新协作空间", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.Update)
		reg.DELETEAction("/:id", "删除协作空间", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.Delete)
		reg.GETAction("/:id/menus", "获取协作空间菜单边界", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.GetTenantMenus)
		reg.GETAction("/:id/menu-origins", "获取协作空间菜单来源", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.GetTenantMenuOrigins)
		reg.PUTAction("/:id/menus", "配置协作空间菜单边界", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.SetTenantMenus)
		reg.GETAction("/:id/actions", "获取协作空间功能权限边界", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.GetTenantActions)
		reg.GETAction("/:id/action-origins", "获取协作空间功能权限来源", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.GetTenantActionOrigins)
		reg.PUTAction("/:id/actions", "配置协作空间功能权限边界", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.SetTenantActions)
		reg.GETAction("/:id/members", "获取协作空间成员列表", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.ListMembers)
		reg.POSTAction("/:id/members", "添加协作空间成员", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.AddMember)
		reg.DELETEAction("/:id/members/:userId", "移除协作空间成员", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.RemoveMember)
		reg.PUTAction("/:id/members/:userId/role", "更新协作空间成员身份", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.UpdateMemberRole)
		reg.GETAction("/options", "获取协作空间候选", "collaboration_workspace.manage", authzService.RequireAction, tenantHandler.ListOptions)
	}
}

func init() {
	module.GetRegistry().Register(&tenantModuleWrapper{})
}

type tenantModuleWrapper struct{}

func (w *tenantModuleWrapper) Init() error {
	return nil
}

func (w *tenantModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
