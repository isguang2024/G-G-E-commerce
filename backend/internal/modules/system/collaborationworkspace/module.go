package collaborationworkspace

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
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
)

type CollaborationWorkspaceModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewCollaborationWorkspaceModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *CollaborationWorkspaceModule {
	return &CollaborationWorkspaceModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *CollaborationWorkspaceModule) Init() error {
	m.logger.Info("Initializing CollaborationWorkspace module")
	return nil
}

func (m *CollaborationWorkspaceModule) RegisterRoutes(rg *gin.RouterGroup) {
	collaborationWorkspaceRepo := user.NewTenantRepository(m.db)
	collaborationWorkspaceMemberRepo := user.NewTenantMemberRepository(m.db)
	userRepo := user.NewUserRepository(m.db)
	roleRepo := user.NewRoleRepository(m.db)
	roleHiddenMenuRepo := user.NewRoleHiddenMenuRepository(m.db)
	roleDisabledActionRepo := user.NewRoleDisabledActionRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	actionRepo := user.NewPermissionKeyRepository(m.db)
	blockedMenuRepo := user.NewCollaborationWorkspaceBlockedMenuRepository(m.db)
	blockedActionRepo := user.NewCollaborationWorkspaceBlockedActionRepository(m.db)
	collaborationWorkspaceFeaturePackageRepo := user.NewCollaborationWorkspaceFeaturePackageRepository(m.db)
	rolePackageRepo := user.NewRoleFeaturePackageRepository(m.db)
	featurePkgRepo := user.NewFeaturePackageRepository(m.db)
	packageActionRepo := user.NewFeaturePackageKeyRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	boundaryService := collaborationworkspaceboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	workspaceService := workspacepkg.NewService(m.db, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	collaborationWorkspaceService := NewCollaborationWorkspaceService(m.db, collaborationWorkspaceRepo, collaborationWorkspaceMemberRepo, userRepo, roleRepo, userRoleRepo, refresher, m.logger)
	collaborationWorkspaceHandler := NewCollaborationWorkspaceHandler(
		collaborationWorkspaceService,
		collaborationWorkspaceMemberRepo,
		userRepo,
		roleRepo,
		roleHiddenMenuRepo,
		roleDisabledActionRepo,
		userRoleRepo,
		actionRepo,
		blockedMenuRepo,
		blockedActionRepo,
		collaborationWorkspaceFeaturePackageRepo,
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
		reg.GET("/mine", &apiregistry.RouteMeta{Summary: "获取我的协作空间列表"}, collaborationWorkspaceHandler.ListMyTeams)
		reg.GET("/current", &apiregistry.RouteMeta{Summary: "获取当前协作空间详情"}, collaborationWorkspaceHandler.GetMyTeam)
		reg.GET("/current/members", &apiregistry.RouteMeta{Summary: "获取当前协作空间成员列表"}, collaborationWorkspaceHandler.ListMyMembers)
		reg.POSTAction("/current/members", "添加当前协作空间成员", "collaboration_workspace.member.manage", authzService.RequireAction, collaborationWorkspaceHandler.AddMyMember)
		reg.DELETEAction("/current/members/:userId", "移除当前协作空间成员", "collaboration_workspace.member.manage", authzService.RequireAction, collaborationWorkspaceHandler.RemoveMyMember)
		reg.PUTAction("/current/members/:userId/role", "更新当前协作空间成员身份", "collaboration_workspace.member.manage", authzService.RequireAction, collaborationWorkspaceHandler.UpdateMyMemberRole)
		reg.GETAction("/current/members/:userId/roles", "获取当前协作空间成员角色", "collaboration_workspace.member.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyCollaborationWorkspaceMemberRoles)
		reg.PUTAction("/current/members/:userId/roles", "配置当前协作空间成员角色", "collaboration_workspace.member.manage", authzService.RequireAction, collaborationWorkspaceHandler.SetMyCollaborationWorkspaceMemberRoles)
		reg.GETAction("/current/roles", "获取当前协作空间可分配角色", "collaboration_workspace.member.manage", authzService.RequireAction, collaborationWorkspaceHandler.ListMyTeamRoles)
		reg.POSTAction("/current/roles", "创建当前协作空间角色", "collaboration_workspace.member.manage", authzService.RequireAction, collaborationWorkspaceHandler.CreateMyTeamRole)
		reg.GETAction("/current/boundary/roles", "获取当前协作空间边界可见角色", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.ListMyTeamRoles)
		reg.POSTAction("/current/boundary/roles", "创建当前协作空间角色(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.CreateMyTeamRole)
		reg.PUTAction("/current/boundary/roles/:roleId", "更新当前协作空间角色(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.UpdateMyTeamRole)
		reg.DELETEAction("/current/boundary/roles/:roleId", "删除当前协作空间角色(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.DeleteMyTeamRole)
		reg.GETAction("/current/boundary/roles/:roleId/packages", "获取当前协作空间角色功能包(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamRolePackages)
		reg.PUTAction("/current/boundary/roles/:roleId/packages", "配置当前协作空间角色功能包(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.SetMyTeamRolePackages)
		reg.GETAction("/current/boundary/roles/:roleId/menus", "获取当前协作空间角色菜单权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamRoleMenus)
		reg.PUTAction("/current/boundary/roles/:roleId/menus", "配置当前协作空间角色菜单权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.SetMyTeamRoleMenus)
		reg.GETAction("/current/boundary/roles/:roleId/actions", "获取当前协作空间角色功能权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamRoleActions)
		reg.PUTAction("/current/boundary/roles/:roleId/actions", "配置当前协作空间角色功能权限(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.SetMyTeamRoleActions)
		reg.GETAction("/current/boundary/packages", "获取当前协作空间已开通功能包(边界管理)", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamBoundaryPackages)
		reg.GETAction("/current/menus", "获取当前协作空间菜单边界", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamMenus)
		reg.GETAction("/current/menu-origins", "获取当前协作空间菜单来源", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamMenuOrigins)
		reg.GETAction("/current/actions", "获取当前协作空间功能权限边界", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamActions)
		reg.GETAction("/current/action-origins", "获取当前协作空间功能权限来源", "collaboration_workspace.boundary.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetMyTeamActionOrigins)

		reg.GETAction("", "获取协作空间列表", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.List)
		reg.GETAction("/:id/roles", "获取协作空间可分配角色", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.ListTenantRoles)
		reg.GETAction("/:id", "获取协作空间详情", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.Get)
		reg.POSTAction("", "创建协作空间", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.Create)
		reg.PUTAction("/:id", "更新协作空间", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.Update)
		reg.DELETEAction("/:id", "删除协作空间", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.Delete)
		reg.GETAction("/:id/menus", "获取协作空间菜单边界", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetTenantMenus)
		reg.GETAction("/:id/menu-origins", "获取协作空间菜单来源", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetTenantMenuOrigins)
		reg.PUTAction("/:id/menus", "配置协作空间菜单边界", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.SetTenantMenus)
		reg.GETAction("/:id/actions", "获取协作空间功能权限边界", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetTenantActions)
		reg.GETAction("/:id/action-origins", "获取协作空间功能权限来源", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.GetTenantActionOrigins)
		reg.PUTAction("/:id/actions", "配置协作空间功能权限边界", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.SetTenantActions)
		reg.GETAction("/:id/members", "获取协作空间成员列表", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.ListMembers)
		reg.POSTAction("/:id/members", "添加协作空间成员", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.AddMember)
		reg.DELETEAction("/:id/members/:userId", "移除协作空间成员", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.RemoveMember)
		reg.PUTAction("/:id/members/:userId/role", "更新协作空间成员身份", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.UpdateMemberRole)
		reg.GETAction("/options", "获取协作空间候选", "collaboration_workspace.manage", authzService.RequireAction, collaborationWorkspaceHandler.ListOptions)
	}
}

func init() {
	module.GetRegistry().Register(&collaborationWorkspaceModuleWrapper{})
}

type collaborationWorkspaceModuleWrapper struct{}

func (w *collaborationWorkspaceModuleWrapper) Init() error {
	return nil
}

func (w *collaborationWorkspaceModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
