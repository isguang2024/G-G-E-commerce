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
	roleActionRepo := user.NewRoleActionPermissionRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	actionRepo := user.NewPermissionActionRepository(m.db)
	tenantActionRepo := user.NewTenantActionPermissionRepository(m.db)
	userActionRepo := user.NewUserActionPermissionRepository(m.db)

	tenantService := NewTenantService(m.db, tenantRepo, tenantMemberRepo, userRepo, roleRepo, userRoleRepo, m.logger)
	tenantHandler := NewTenantHandler(tenantService, tenantMemberRepo, userRepo, roleRepo, roleActionRepo, userRoleRepo, actionRepo, tenantActionRepo, userActionRepo, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	tenants := rg.Group("/tenants")
	reg := apiregistry.NewRegistrar(tenants, "tenant")
	{
		reg.GET("/my-teams", &apiregistry.RouteMeta{Summary: "获取我的团队列表"}, tenantHandler.ListMyTeams)
		reg.GET("/my-team", &apiregistry.RouteMeta{Summary: "获取当前团队详情"}, tenantHandler.GetMyTeam)
		reg.GET("/my-team/members", &apiregistry.RouteMeta{Summary: "获取当前团队成员列表"}, tenantHandler.ListMyMembers)
	reg.POST("/my-team/members", &apiregistry.RouteMeta{Summary: "添加当前团队成员", ResourceCode: "team_member", ActionCode: "create", ScopeCode: "team"}, authzService.RequireAction("team_member", "create", "team"), tenantHandler.AddMyMember)
	reg.DELETE("/my-team/members/:userId", &apiregistry.RouteMeta{Summary: "移除当前团队成员", ResourceCode: "team_member", ActionCode: "delete", ScopeCode: "team"}, authzService.RequireAction("team_member", "delete", "team"), tenantHandler.RemoveMyMember)
	reg.PUT("/my-team/members/:userId/role", &apiregistry.RouteMeta{Summary: "更新当前团队成员身份", ResourceCode: "team_member", ActionCode: "update_role", ScopeCode: "team"}, authzService.RequireAction("team_member", "update_role", "team"), tenantHandler.UpdateMyMemberRole)
	reg.GET("/my-team/members/:userId/roles", &apiregistry.RouteMeta{Summary: "获取当前团队成员角色", ResourceCode: "team_member", ActionCode: "assign_role", ScopeCode: "team"}, authzService.RequireAction("team_member", "assign_role", "team"), tenantHandler.GetMyTeamMemberRoles)
	reg.PUT("/my-team/members/:userId/roles", &apiregistry.RouteMeta{Summary: "配置当前团队成员角色", ResourceCode: "team_member", ActionCode: "assign_role", ScopeCode: "team"}, authzService.RequireAction("team_member", "assign_role", "team"), tenantHandler.SetMyTeamMemberRoles)
	reg.GET("/my-team/members/:userId/actions", &apiregistry.RouteMeta{Summary: "获取当前团队成员功能权限", ResourceCode: "team_member", ActionCode: "assign_action", ScopeCode: "team"}, authzService.RequireAction("team_member", "assign_action", "team"), tenantHandler.GetMyTeamMemberActionPermissions)
	reg.PUT("/my-team/members/:userId/actions", &apiregistry.RouteMeta{Summary: "配置当前团队成员功能权限", ResourceCode: "team_member", ActionCode: "assign_action", ScopeCode: "team"}, authzService.RequireAction("team_member", "assign_action", "team"), tenantHandler.SetMyTeamMemberActionPermissions)
	reg.GET("/my-team/roles", &apiregistry.RouteMeta{Summary: "获取当前团队可分配角色", ResourceCode: "team_member", ActionCode: "assign_role", ScopeCode: "team"}, authzService.RequireAction("team_member", "assign_role", "team"), tenantHandler.ListMyTeamRoles)
	reg.GET("/my-team/roles/:roleId/actions", &apiregistry.RouteMeta{Summary: "获取当前团队角色功能权限", ResourceCode: "team_member", ActionCode: "assign_action", ScopeCode: "team"}, authzService.RequireAction("team_member", "assign_action", "team"), tenantHandler.GetMyTeamRoleActions)
	reg.GET("/my-team/actions", &apiregistry.RouteMeta{Summary: "获取当前团队功能权限边界", ResourceCode: "team", ActionCode: "configure_action_boundary", ScopeCode: "team"}, authzService.RequireAction("team", "configure_action_boundary", "team"), tenantHandler.GetMyTeamActions)

	reg.GET("", &apiregistry.RouteMeta{Summary: "获取团队列表", ResourceCode: "tenant", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("tenant", "list", "global"), tenantHandler.List)
	reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取团队详情", ResourceCode: "tenant", ActionCode: "get", ScopeCode: "global"}, authzService.RequireAction("tenant", "get", "global"), tenantHandler.Get)
	reg.POST("", &apiregistry.RouteMeta{Summary: "创建团队", ResourceCode: "tenant", ActionCode: "create", ScopeCode: "global"}, authzService.RequireAction("tenant", "create", "global"), tenantHandler.Create)
	reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新团队", ResourceCode: "tenant", ActionCode: "update", ScopeCode: "global"}, authzService.RequireAction("tenant", "update", "global"), tenantHandler.Update)
	reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除团队", ResourceCode: "tenant", ActionCode: "delete", ScopeCode: "global"}, authzService.RequireAction("tenant", "delete", "global"), tenantHandler.Delete)
	reg.GET("/:id/actions", &apiregistry.RouteMeta{Summary: "获取团队功能权限边界", ResourceCode: "tenant", ActionCode: "configure_action_boundary", ScopeCode: "global"}, authzService.RequireAction("tenant", "configure_action_boundary", "global"), tenantHandler.GetTenantActions)
	reg.PUT("/:id/actions", &apiregistry.RouteMeta{Summary: "配置团队功能权限边界", ResourceCode: "tenant", ActionCode: "configure_action_boundary", ScopeCode: "global"}, authzService.RequireAction("tenant", "configure_action_boundary", "global"), tenantHandler.SetTenantActions)
	reg.GET("/:id/members", &apiregistry.RouteMeta{Summary: "获取团队成员列表", ResourceCode: "tenant_member_admin", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("tenant_member_admin", "list", "global"), tenantHandler.ListMembers)
	reg.POST("/:id/members", &apiregistry.RouteMeta{Summary: "添加团队成员", ResourceCode: "tenant_member_admin", ActionCode: "create", ScopeCode: "global"}, authzService.RequireAction("tenant_member_admin", "create", "global"), tenantHandler.AddMember)
	reg.DELETE("/:id/members/:userId", &apiregistry.RouteMeta{Summary: "移除团队成员", ResourceCode: "tenant_member_admin", ActionCode: "delete", ScopeCode: "global"}, authzService.RequireAction("tenant_member_admin", "delete", "global"), tenantHandler.RemoveMember)
	reg.PUT("/:id/members/:userId/role", &apiregistry.RouteMeta{Summary: "更新团队成员身份", ResourceCode: "tenant_member_admin", ActionCode: "update_role", ScopeCode: "global"}, authzService.RequireAction("tenant_member_admin", "update_role", "global"), tenantHandler.UpdateMemberRole)
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
