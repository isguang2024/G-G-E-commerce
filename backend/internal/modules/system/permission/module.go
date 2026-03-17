package permission

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

type PermissionModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewPermissionModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *PermissionModule {
	return &PermissionModule{db: db, config: cfg, logger: logger}
}

func (m *PermissionModule) Init() error {
	m.logger.Info("Initializing Permission module")
	return nil
}

func (m *PermissionModule) RegisterRoutes(rg *gin.RouterGroup) {
	actionRepo := user.NewPermissionActionRepository(m.db)
	roleActionRepo := user.NewRoleActionPermissionRepository(m.db)
	tenantActionRepo := user.NewTenantActionPermissionRepository(m.db)
	userActionRepo := user.NewUserActionPermissionRepository(m.db)
	scopeRepo := user.NewScopeRepository(m.db)
	service := NewPermissionService(actionRepo, roleActionRepo, tenantActionRepo, userActionRepo, scopeRepo)
	handler := NewPermissionHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	actions := rg.Group("/permission-actions")
	reg := apiregistry.NewRegistrar(actions, "permission_action")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取功能权限列表", ResourceCode: "permission_action", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("permission_action", "list"), handler.List)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取功能权限详情", ResourceCode: "permission_action", ActionCode: "get", ScopeCode: "global"}, authzService.RequireAction("permission_action", "get"), handler.Get)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建功能权限", ResourceCode: "permission_action", ActionCode: "create", ScopeCode: "global"}, authzService.RequireAction("permission_action", "create"), handler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新功能权限", ResourceCode: "permission_action", ActionCode: "update", ScopeCode: "global"}, authzService.RequireAction("permission_action", "update"), handler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除功能权限", ResourceCode: "permission_action", ActionCode: "delete", ScopeCode: "global"}, authzService.RequireAction("permission_action", "delete"), handler.Delete)
	}
}

func init() {
	module.GetRegistry().Register(&permissionModuleWrapper{})
}

type permissionModuleWrapper struct{}

func (w *permissionModuleWrapper) Init() error {
	return nil
}

func (w *permissionModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
