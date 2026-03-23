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
	userActionRepo := user.NewUserActionPermissionRepository(m.db)
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
	tenantHandler := NewTenantHandler(tenantService, tenantMemberRepo, userRepo, roleRepo, roleHiddenMenuRepo, roleDisabledActionRepo, userRoleRepo, actionRepo, blockedMenuRepo, blockedActionRepo, userActionRepo, teamPackageRepo, rolePackageRepo, featurePkgRepo, packageActionRepo, packageMenuRepo, boundaryService, refresher, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	tenants := rg.Group("/tenants")
	reg := apiregistry.NewRegistrar(tenants, "tenant")
	{
		reg.GET("/my-teams", &apiregistry.RouteMeta{Summary: "获取我的团队列表"}, tenantHandler.ListMyTeams)
		reg.GET("/my-team", &apiregistry.RouteMeta{Summary: "获取当前团队详情"}, tenantHandler.GetMyTeam)
		reg.GET("/my-team/members", &apiregistry.RouteMeta{Summary: "获取当前团队成员列表"}, tenantHandler.ListMyMembers)
		reg.POST("/my-team/members", &apiregistry.RouteMeta{Summary: "添加当前团队成员", ResourceCode: "team_member", ActionCode: "create"}, authzService.RequireAction("team.member.manage"), tenantHandler.AddMyMember)
		reg.DELETE("/my-team/members/:userId", &apiregistry.RouteMeta{Summary: "移除当前团队成员", ResourceCode: "team_member", ActionCode: "delete"}, authzService.RequireAction("team.member.manage"), tenantHandler.RemoveMyMember)
		reg.PUT("/my-team/members/:userId/role", &apiregistry.RouteMeta{Summary: "更新当前团队成员身份", ResourceCode: "team_member", ActionCode: "update_role"}, authzService.RequireAction("team.member.manage"), tenantHandler.UpdateMyMemberRole)
		reg.GET("/my-team/members/:userId/roles", &apiregistry.RouteMeta{Summary: "获取当前团队成员角色", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.GetMyTeamMemberRoles)
		reg.PUT("/my-team/members/:userId/roles", &apiregistry.RouteMeta{Summary: "配置当前团队成员角色", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.SetMyTeamMemberRoles)
		reg.GET("/my-team/members/:userId/actions", &apiregistry.RouteMeta{Summary: "获取当前团队成员功能权限", ResourceCode: "team_member", ActionCode: "assign_action"}, authzService.RequireAction("team.member.assign_action"), tenantHandler.GetMyTeamMemberActionPermissions)
		reg.PUT("/my-team/members/:userId/actions", &apiregistry.RouteMeta{Summary: "配置当前团队成员功能权限", ResourceCode: "team_member", ActionCode: "assign_action"}, authzService.RequireAction("team.member.assign_action"), tenantHandler.SetMyTeamMemberActionPermissions)
		reg.GET("/my-team/roles", &apiregistry.RouteMeta{Summary: "获取当前团队可分配角色", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.ListMyTeamRoles)
		reg.POST("/my-team/roles", &apiregistry.RouteMeta{Summary: "创建当前团队角色", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.CreateMyTeamRole)
		reg.PUT("/my-team/roles/:roleId", &apiregistry.RouteMeta{Summary: "更新当前团队角色", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.UpdateMyTeamRole)
		reg.DELETE("/my-team/roles/:roleId", &apiregistry.RouteMeta{Summary: "删除当前团队角色", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.DeleteMyTeamRole)
		reg.GET("/my-team/roles/:roleId/packages", &apiregistry.RouteMeta{Summary: "获取当前团队角色功能包", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.GetMyTeamRolePackages)
		reg.PUT("/my-team/roles/:roleId/packages", &apiregistry.RouteMeta{Summary: "配置当前团队角色功能包", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.SetMyTeamRolePackages)
		reg.GET("/my-team/roles/:roleId/menus", &apiregistry.RouteMeta{Summary: "获取当前团队角色菜单权限", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.GetMyTeamRoleMenus)
		reg.PUT("/my-team/roles/:roleId/menus", &apiregistry.RouteMeta{Summary: "配置当前团队角色菜单权限", ResourceCode: "team_member", ActionCode: "assign_role"}, authzService.RequireAction("team.member.assign_role"), tenantHandler.SetMyTeamRoleMenus)
		reg.GET("/my-team/roles/:roleId/actions", &apiregistry.RouteMeta{Summary: "获取当前团队角色功能权限", ResourceCode: "team_member", ActionCode: "assign_action"}, authzService.RequireAction("team.member.assign_action"), tenantHandler.GetMyTeamRoleActions)
		reg.PUT("/my-team/roles/:roleId/actions", &apiregistry.RouteMeta{Summary: "配置当前团队角色功能权限", ResourceCode: "team_member", ActionCode: "assign_action"}, authzService.RequireAction("team.member.assign_action"), tenantHandler.SetMyTeamRoleActions)
		reg.GET("/my-team/menus", &apiregistry.RouteMeta{Summary: "获取当前团队菜单边界", ResourceCode: "team", ActionCode: "configure_menu_boundary"}, authzService.RequireAction("team.boundary.manage"), tenantHandler.GetMyTeamMenus)
		reg.GET("/my-team/menu-origins", &apiregistry.RouteMeta{Summary: "获取当前团队菜单来源", ResourceCode: "team", ActionCode: "configure_menu_boundary"}, authzService.RequireAction("team.boundary.manage"), tenantHandler.GetMyTeamMenuOrigins)
		reg.GET("/my-team/actions", &apiregistry.RouteMeta{Summary: "获取当前团队功能权限边界", ResourceCode: "team", ActionCode: "configure_action_boundary"}, authzService.RequireAction("team.boundary.manage"), tenantHandler.GetMyTeamActions)
		reg.GET("/my-team/action-origins", &apiregistry.RouteMeta{Summary: "获取当前团队功能权限来源", ResourceCode: "team", ActionCode: "configure_action_boundary"}, authzService.RequireAction("team.boundary.manage"), tenantHandler.GetMyTeamActionOrigins)

		reg.GET("", &apiregistry.RouteMeta{Summary: "获取团队列表", ResourceCode: "tenant", ActionCode: "list"}, authzService.RequireAction("tenant.manage"), tenantHandler.List)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取团队详情", ResourceCode: "tenant", ActionCode: "get"}, authzService.RequireAction("tenant.manage"), tenantHandler.Get)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建团队", ResourceCode: "tenant", ActionCode: "create"}, authzService.RequireAction("tenant.manage"), tenantHandler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新团队", ResourceCode: "tenant", ActionCode: "update"}, authzService.RequireAction("tenant.manage"), tenantHandler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除团队", ResourceCode: "tenant", ActionCode: "delete"}, authzService.RequireAction("tenant.manage"), tenantHandler.Delete)
		reg.GET("/:id/menus", &apiregistry.RouteMeta{Summary: "获取团队菜单边界", ResourceCode: "tenant", ActionCode: "configure_menu_boundary"}, authzService.RequireAction("tenant.boundary.manage"), tenantHandler.GetTenantMenus)
		reg.GET("/:id/menu-origins", &apiregistry.RouteMeta{Summary: "获取团队菜单来源", ResourceCode: "tenant", ActionCode: "configure_menu_boundary"}, authzService.RequireAction("tenant.boundary.manage"), tenantHandler.GetTenantMenuOrigins)
		reg.PUT("/:id/menus", &apiregistry.RouteMeta{Summary: "配置团队菜单边界", ResourceCode: "tenant", ActionCode: "configure_menu_boundary"}, authzService.RequireAction("tenant.boundary.manage"), tenantHandler.SetTenantMenus)
		reg.GET("/:id/actions", &apiregistry.RouteMeta{Summary: "获取团队功能权限边界", ResourceCode: "tenant", ActionCode: "configure_action_boundary"}, authzService.RequireAction("tenant.boundary.manage"), tenantHandler.GetTenantActions)
		reg.GET("/:id/action-origins", &apiregistry.RouteMeta{Summary: "获取团队功能权限来源", ResourceCode: "tenant", ActionCode: "configure_action_boundary"}, authzService.RequireAction("tenant.boundary.manage"), tenantHandler.GetTenantActionOrigins)
		reg.PUT("/:id/actions", &apiregistry.RouteMeta{Summary: "配置团队功能权限边界", ResourceCode: "tenant", ActionCode: "configure_action_boundary"}, authzService.RequireAction("tenant.boundary.manage"), tenantHandler.SetTenantActions)
		reg.GET("/:id/members", &apiregistry.RouteMeta{Summary: "获取团队成员列表", ResourceCode: "tenant_member_admin", ActionCode: "list"}, authzService.RequireAction("tenant.member.manage"), tenantHandler.ListMembers)
		reg.POST("/:id/members", &apiregistry.RouteMeta{Summary: "添加团队成员", ResourceCode: "tenant_member_admin", ActionCode: "create"}, authzService.RequireAction("tenant.member.manage"), tenantHandler.AddMember)
		reg.DELETE("/:id/members/:userId", &apiregistry.RouteMeta{Summary: "移除团队成员", ResourceCode: "tenant_member_admin", ActionCode: "delete"}, authzService.RequireAction("tenant.member.manage"), tenantHandler.RemoveMember)
		reg.PUT("/:id/members/:userId/role", &apiregistry.RouteMeta{Summary: "更新团队成员身份", ResourceCode: "tenant_member_admin", ActionCode: "update_role"}, authzService.RequireAction("tenant.member.manage"), tenantHandler.UpdateMemberRole)
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
