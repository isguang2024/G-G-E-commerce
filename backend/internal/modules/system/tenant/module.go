package tenant

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
	actionRepo := user.NewPermissionActionRepository(m.db)
	blockedMenuRepo := user.NewTeamBlockedMenuRepository(m.db)
	blockedActionRepo := user.NewTeamBlockedActionRepository(m.db)
	teamPackageRepo := user.NewTeamFeaturePackageRepository(m.db)
	rolePackageRepo := user.NewRoleFeaturePackageRepository(m.db)
	featurePkgRepo := user.NewFeaturePackageRepository(m.db)
	packageActionRepo := user.NewFeaturePackageActionRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)

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
		m.logger,
	)
	authzService := authorization.NewService(m.db, m.logger)

	tenants := rg.Group("/tenants")
	reg := apiregistry.NewRegistrar(tenants, "tenant")
	{
		reg.GET("/my-teams", &apiregistry.RouteMeta{Summary: "获取我的团队列表"}, tenantHandler.ListMyTeams)
		reg.GET("/my-team", &apiregistry.RouteMeta{Summary: "获取当前团队详情"}, tenantHandler.GetMyTeam)
		reg.GET("/my-team/members", &apiregistry.RouteMeta{Summary: "获取当前团队成员列表"}, tenantHandler.ListMyMembers)
		reg.POSTAction("/my-team/members", "添加当前团队成员", "team.member.manage", authzService.RequireAction, tenantHandler.AddMyMember)
		reg.DELETEAction("/my-team/members/:userId", "移除当前团队成员", "team.member.manage", authzService.RequireAction, tenantHandler.RemoveMyMember)
		reg.PUTAction("/my-team/members/:userId/role", "更新当前团队成员身份", "team.member.manage", authzService.RequireAction, tenantHandler.UpdateMyMemberRole)
		reg.GETAction("/my-team/members/:userId/roles", "获取当前团队成员角色", "team.member.manage", authzService.RequireAction, tenantHandler.GetMyTeamMemberRoles)
		reg.PUTAction("/my-team/members/:userId/roles", "配置当前团队成员角色", "team.member.manage", authzService.RequireAction, tenantHandler.SetMyTeamMemberRoles)
		reg.GETAction("/my-team/roles", "获取当前团队可分配角色", "team.member.manage", authzService.RequireAction, tenantHandler.ListMyTeamRoles)
		reg.POSTAction("/my-team/roles", "创建当前团队角色", "team.member.manage", authzService.RequireAction, tenantHandler.CreateMyTeamRole)
		reg.GETAction("/my-team/boundary/roles", "获取当前团队边界可见角色", "team.boundary.manage", authzService.RequireAction, tenantHandler.ListMyTeamRoles)
		reg.PUTAction("/my-team/boundary/roles/:roleId", "更新当前团队角色(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.UpdateMyTeamRole)
		reg.DELETEAction("/my-team/boundary/roles/:roleId", "删除当前团队角色(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.DeleteMyTeamRole)
		reg.GETAction("/my-team/boundary/roles/:roleId/packages", "获取当前团队角色功能包(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamRolePackages)
		reg.PUTAction("/my-team/boundary/roles/:roleId/packages", "配置当前团队角色功能包(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.SetMyTeamRolePackages)
		reg.GETAction("/my-team/boundary/roles/:roleId/menus", "获取当前团队角色菜单权限(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamRoleMenus)
		reg.PUTAction("/my-team/boundary/roles/:roleId/menus", "配置当前团队角色菜单权限(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.SetMyTeamRoleMenus)
		reg.GETAction("/my-team/boundary/roles/:roleId/actions", "获取当前团队角色功能权限(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamRoleActions)
		reg.PUTAction("/my-team/boundary/roles/:roleId/actions", "配置当前团队角色功能权限(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.SetMyTeamRoleActions)
		reg.GETAction("/my-team/boundary/packages", "获取当前团队已开通功能包(边界管理)", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamBoundaryPackages)
		reg.GETAction("/my-team/menus", "获取当前团队菜单边界", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamMenus)
		reg.GETAction("/my-team/menu-origins", "获取当前团队菜单来源", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamMenuOrigins)
		reg.GETAction("/my-team/actions", "获取当前团队功能权限边界", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamActions)
		reg.GETAction("/my-team/action-origins", "获取当前团队功能权限来源", "team.boundary.manage", authzService.RequireAction, tenantHandler.GetMyTeamActionOrigins)

		reg.GETAction("", "获取团队列表", "tenant.manage", authzService.RequireAction, tenantHandler.List)
		reg.GETAction("/:id", "获取团队详情", "tenant.manage", authzService.RequireAction, tenantHandler.Get)
		reg.POSTAction("", "创建团队", "tenant.manage", authzService.RequireAction, tenantHandler.Create)
		reg.PUTAction("/:id", "更新团队", "tenant.manage", authzService.RequireAction, tenantHandler.Update)
		reg.DELETEAction("/:id", "删除团队", "tenant.manage", authzService.RequireAction, tenantHandler.Delete)
		reg.GETAction("/:id/menus", "获取团队菜单边界", "tenant.manage", authzService.RequireAction, tenantHandler.GetTenantMenus)
		reg.GETAction("/:id/menu-origins", "获取团队菜单来源", "tenant.manage", authzService.RequireAction, tenantHandler.GetTenantMenuOrigins)
		reg.PUTAction("/:id/menus", "配置团队菜单边界", "tenant.manage", authzService.RequireAction, tenantHandler.SetTenantMenus)
		reg.GETAction("/:id/actions", "获取团队功能权限边界", "tenant.manage", authzService.RequireAction, tenantHandler.GetTenantActions)
		reg.GETAction("/:id/action-origins", "获取团队功能权限来源", "tenant.manage", authzService.RequireAction, tenantHandler.GetTenantActionOrigins)
		reg.PUTAction("/:id/actions", "配置团队功能权限边界", "tenant.manage", authzService.RequireAction, tenantHandler.SetTenantActions)
		reg.GETAction("/:id/members", "获取团队成员列表", "tenant.manage", authzService.RequireAction, tenantHandler.ListMembers)
		reg.POSTAction("/:id/members", "添加团队成员", "tenant.manage", authzService.RequireAction, tenantHandler.AddMember)
		reg.DELETEAction("/:id/members/:userId", "移除团队成员", "tenant.manage", authzService.RequireAction, tenantHandler.RemoveMember)
		reg.PUTAction("/:id/members/:userId/role", "更新团队成员身份", "tenant.manage", authzService.RequireAction, tenantHandler.UpdateMemberRole)
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
