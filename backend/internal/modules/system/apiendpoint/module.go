package apiendpoint

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

type Module struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
	router *gin.Engine
}

func NewModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger, router *gin.Engine) *Module {
	return &Module{db: db, config: cfg, logger: logger, router: router}
}

func (m *Module) Init() error {
	m.logger.Info("Initializing API endpoint module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	repo := user.NewAPIEndpointRepository(m.db)
	service := NewService(m.db, repo, m.router, m.logger)
	handler := NewHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	endpoints := rg.Group("/api-endpoints")
	reg := apiregistry.NewRegistrar(endpoints, "api_endpoint")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取 API 注册表", ResourceCode: "api_endpoint", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("api_endpoint", "list"), handler.List)
		reg.POST("/sync", &apiregistry.RouteMeta{Summary: "同步 API 注册表", ResourceCode: "api_endpoint", ActionCode: "sync", ScopeCode: "global"}, authzService.RequireAction("api_endpoint", "sync"), handler.Sync)
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
