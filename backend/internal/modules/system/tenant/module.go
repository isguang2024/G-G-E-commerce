package tenant

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
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
	userRoleRepo := user.NewUserRoleRepository(m.db)

	tenantService := NewTenantService(tenantRepo, tenantMemberRepo, userRepo, roleRepo, userRoleRepo, m.logger)
	tenantHandler := NewTenantHandler(tenantService, tenantMemberRepo, userRepo, roleRepo, userRoleRepo, m.logger)

	tenants := rg.Group("/tenants")
	{
		tenants.GET("/my-team", tenantHandler.GetMyTeam)
		tenants.GET("/my-team/members", tenantHandler.ListMyMembers)
		tenants.POST("/my-team/members", tenantHandler.AddMyMember)
		tenants.DELETE("/my-team/members/:userId", tenantHandler.RemoveMyMember)
		tenants.PUT("/my-team/members/:userId/role", tenantHandler.UpdateMyMemberRole)
		tenants.GET("/my-team/members/:userId/roles", tenantHandler.GetMyTeamMemberRoles)
		tenants.PUT("/my-team/members/:userId/roles", tenantHandler.SetMyTeamMemberRoles)
		tenants.GET("/my-team/roles", tenantHandler.ListMyTeamRoles)

		tenants.GET("", tenantHandler.List)
		tenants.GET("/:id", tenantHandler.Get)
		tenants.POST("", tenantHandler.Create)
		tenants.PUT("/:id", tenantHandler.Update)
		tenants.DELETE("/:id", tenantHandler.Delete)
		tenants.GET("/:id/members", tenantHandler.ListMembers)
		tenants.POST("/:id/members", tenantHandler.AddMember)
		tenants.DELETE("/:id/members/:userId", tenantHandler.RemoveMember)
		tenants.PUT("/:id/members/:userId/role", tenantHandler.UpdateMemberRole)
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
