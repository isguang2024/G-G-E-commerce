package featurepackage

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

type Module struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Module {
	return &Module{db: db, config: cfg, logger: logger}
}

func (m *Module) Init() error {
	m.logger.Info("Initializing feature package module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	packageRepo := user.NewFeaturePackageRepository(m.db)
	packageBundleRepo := user.NewFeaturePackageBundleRepository(m.db)
	packageActionRepo := user.NewFeaturePackageActionRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	teamPackageRepo := user.NewTeamFeaturePackageRepository(m.db)
	rolePackageRepo := user.NewRoleFeaturePackageRepository(m.db)
	actionRepo := user.NewPermissionActionRepository(m.db)
	menuRepo := user.NewMenuRepository(m.db)
	tenantRepo := user.NewTenantRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	service := NewService(packageRepo, packageBundleRepo, packageActionRepo, packageMenuRepo, teamPackageRepo, rolePackageRepo, actionRepo, menuRepo, tenantRepo, boundaryService, refresher)
	handler := NewHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	group := rg.Group("/feature-packages")
	reg := apiregistry.NewRegistrar(group, "feature_package")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取功能包列表", ResourceCode: "feature_package", ActionCode: "list"}, authzService.RequireAction("platform.package.manage"), handler.List)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取功能包详情", ResourceCode: "feature_package", ActionCode: "get"}, authzService.RequireAction("platform.package.manage"), handler.Get)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建功能包", ResourceCode: "feature_package", ActionCode: "create"}, authzService.RequireAction("platform.package.manage"), handler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新功能包", ResourceCode: "feature_package", ActionCode: "update"}, authzService.RequireAction("platform.package.manage"), handler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除功能包", ResourceCode: "feature_package", ActionCode: "delete"}, authzService.RequireAction("platform.package.manage"), handler.Delete)
		reg.GET("/:id/children", &apiregistry.RouteMeta{Summary: "获取组合包基础包", ResourceCode: "feature_package", ActionCode: "assign_bundle"}, authzService.RequireAction("platform.package.manage"), handler.GetPackageChildren)
		reg.PUT("/:id/children", &apiregistry.RouteMeta{Summary: "配置组合包基础包", ResourceCode: "feature_package", ActionCode: "assign_bundle"}, authzService.RequireAction("platform.package.manage"), handler.SetPackageChildren)
		reg.GET("/:id/actions", &apiregistry.RouteMeta{Summary: "获取功能包权限", ResourceCode: "feature_package", ActionCode: "assign_action"}, authzService.RequireAction("platform.package.manage"), handler.GetPackageActions)
		reg.PUT("/:id/actions", &apiregistry.RouteMeta{Summary: "配置功能包权限", ResourceCode: "feature_package", ActionCode: "assign_action"}, authzService.RequireAction("platform.package.manage"), handler.SetPackageActions)
		reg.GET("/:id/menus", &apiregistry.RouteMeta{Summary: "获取功能包菜单", ResourceCode: "feature_package", ActionCode: "assign_menu"}, authzService.RequireAction("platform.package.manage"), handler.GetPackageMenus)
		reg.PUT("/:id/menus", &apiregistry.RouteMeta{Summary: "配置功能包菜单", ResourceCode: "feature_package", ActionCode: "assign_menu"}, authzService.RequireAction("platform.package.manage"), handler.SetPackageMenus)
		reg.GET("/:id/teams", &apiregistry.RouteMeta{Summary: "获取功能包团队", ResourceCode: "feature_package", ActionCode: "assign_team"}, authzService.RequireAction("platform.package.assign"), handler.GetPackageTeams)
		reg.PUT("/:id/teams", &apiregistry.RouteMeta{Summary: "配置功能包团队", ResourceCode: "feature_package", ActionCode: "assign_team"}, authzService.RequireAction("platform.package.assign"), handler.SetPackageTeams)
		reg.GET("/teams/:teamId", &apiregistry.RouteMeta{Summary: "获取团队功能包", ResourceCode: "feature_package", ActionCode: "assign_team"}, authzService.RequireAction("platform.package.assign"), handler.GetTeamPackages)
		reg.PUT("/teams/:teamId", &apiregistry.RouteMeta{Summary: "配置团队功能包", ResourceCode: "feature_package", ActionCode: "assign_team"}, authzService.RequireAction("platform.package.assign"), handler.SetTeamPackages)
	}
}
