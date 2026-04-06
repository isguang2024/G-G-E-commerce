package apiendpoint

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiendpointaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type Module struct {
	db             *gorm.DB
	config         *config.Config
	logger         *zap.Logger
	router         *gin.Engine
	endpointAccess apiendpointaccess.Service
}

func NewModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger, router *gin.Engine, endpointAccess apiendpointaccess.Service) *Module {
	return &Module{db: db, config: cfg, logger: logger, router: router, endpointAccess: endpointAccess}
}

func (m *Module) Init() error {
	m.logger.Info("Initializing API endpoint module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	repo := user.NewAPIEndpointRepository(m.db)
	categoryRepo := user.NewAPIEndpointCategoryRepository(m.db)
	bindingRepo := user.NewAPIEndpointPermissionBindingRepository(m.db)
	service := NewService(m.db, repo, categoryRepo, bindingRepo, m.router, m.logger, m.config.Env, m.endpointAccess)
	handler := NewHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	endpoints := rg.Group("/api-endpoints")
	reg := apiregistry.NewRegistrar(endpoints, "api_endpoint")
	{
		reg.GETProtected("/overview", reg.Meta("获取 API 概览").BindGroup("api_endpoint").BindPermissionKey("system.api_registry.view").Build(), "system.api_registry.view", authzService.RequireAction, handler.Overview)
		reg.GETProtected("/stale", reg.Meta("获取失效 API").BindGroup("api_endpoint").BindPermissionKey("system.api_registry.view").Build(), "system.api_registry.view", authzService.RequireAction, handler.ListStale)
		reg.GETAction("/unregistered", "获取未注册 API 路由", "system.api_registry.view", authzService.RequireAction, handler.ListUnregistered)
		reg.GETProtected("/unregistered/scan-config", reg.Meta("获取未注册 API 扫描配置").BindGroup("api_endpoint").BindPermissionKey("system.api_registry.view").Build(), "system.api_registry.view", authzService.RequireAction, handler.GetUnregisteredScanConfig)
		reg.PUTProtected("/unregistered/scan-config", reg.Meta("保存未注册 API 扫描配置").BindGroup("api_endpoint").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.SaveUnregisteredScanConfig)
		reg.GETProtected("", reg.Meta("获取 API 注册表").BindGroup("api_endpoint").BindPermissionKey("system.api_registry.view").Build(), "system.api_registry.view", authzService.RequireAction, handler.List)
		reg.GETProtected("/categories", reg.Meta("获取 API 分类").BindGroup("api_endpoint").BindPermissionKey("system.api_registry.view").Build(), "system.api_registry.view", authzService.RequireAction, handler.ListCategories)
		reg.POSTProtected("/sync", reg.Meta("同步 API 注册表").BindGroup("api_endpoint").BindSource("manual").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.Sync)
		reg.POSTProtected("/cleanup-stale", reg.Meta("清理失效 API").BindGroup("api_endpoint").BindSource("manual").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.CleanupStale)
		reg.POSTProtected("", reg.Meta("创建 API 注册项").BindGroup("api_endpoint").BindSource("manual").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.Create)
		reg.PUTProtected("/:id", reg.Meta("更新 API 注册项").BindGroup("api_endpoint").BindSource("manual").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.Update)
		reg.PUTProtected("/:id/context-scope", reg.Meta("更新 API 协作空间上下文").BindGroup("api_endpoint").BindSource("manual").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.UpdateContextScope)
		reg.POSTProtected("/categories", reg.Meta("创建 API 分类").BindGroup("api_endpoint").BindSource("manual").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.SaveCategory)
		reg.PUTProtected("/categories/:id", reg.Meta("更新 API 分类").BindGroup("api_endpoint").BindSource("manual").BindPermissionKey("system.api_registry.sync").Build(), "system.api_registry.sync", authzService.RequireAction, handler.UpdateCategory)
	}
}

func init() {
	module.GetRegistry().Register(&moduleWrapper{})
}

type moduleWrapper struct{}

func (w *moduleWrapper) Init() error {
	return nil
}

func (w *moduleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
