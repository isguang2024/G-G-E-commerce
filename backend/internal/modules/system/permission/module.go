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
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
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
	packageActionRepo := user.NewFeaturePackageActionRepository(m.db)
	teamPackageRepo := user.NewTeamFeaturePackageRepository(m.db)
	tenantActionRepo := user.NewTenantActionPermissionRepository(m.db)
	manualActionRepo := user.NewTeamManualActionPermissionRepository(m.db)
	userActionRepo := user.NewUserActionPermissionRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	service := NewPermissionService(actionRepo, roleActionRepo, packageActionRepo, teamPackageRepo, tenantActionRepo, manualActionRepo, userActionRepo, boundaryService, refresher)
	handler := NewPermissionHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	actions := rg.Group("/permission-actions")
	reg := apiregistry.NewRegistrar(actions, "permission_action")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取功能权限列表", ResourceCode: "permission_action", ActionCode: "list"}, authzService.RequireAction("system.permission.manage"), handler.List)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取功能权限详情", ResourceCode: "permission_action", ActionCode: "get"}, authzService.RequireAction("system.permission.manage"), handler.Get)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建功能权限", ResourceCode: "permission_action", ActionCode: "create"}, authzService.RequireAction("system.permission.manage"), handler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新功能权限", ResourceCode: "permission_action", ActionCode: "update"}, authzService.RequireAction("system.permission.manage"), handler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除功能权限", ResourceCode: "permission_action", ActionCode: "delete"}, authzService.RequireAction("system.permission.manage"), handler.Delete)
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
