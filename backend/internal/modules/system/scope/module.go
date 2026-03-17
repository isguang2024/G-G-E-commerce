package scope

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

type ScopeModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewScopeModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *ScopeModule {
	return &ScopeModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *ScopeModule) Init() error {
	m.logger.Info("Initializing Scope module")
	return nil
}

func (m *ScopeModule) RegisterRoutes(rg *gin.RouterGroup) {
	scopeRepo := user.NewScopeRepository(m.db)
	roleRepo := user.NewRoleRepository(m.db)
	scopeService := NewScopeService(scopeRepo, roleRepo, m.logger)
	scopeHandler := NewScopeHandler(scopeService, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	scopes := rg.Group("/scopes")
	reg := apiregistry.NewRegistrar(scopes, "scope")
	{
		reg.GET("", &apiregistry.RouteMeta{Summary: "获取作用域列表", ResourceCode: "scope", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("scope", "list"), scopeHandler.List)
		reg.GET("/all", &apiregistry.RouteMeta{Summary: "获取全部作用域", ResourceCode: "scope", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("scope", "list"), scopeHandler.GetAll)
		reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取作用域详情", ResourceCode: "scope", ActionCode: "get", ScopeCode: "global"}, authzService.RequireAction("scope", "get"), scopeHandler.Get)
		reg.POST("", &apiregistry.RouteMeta{Summary: "创建作用域", ResourceCode: "scope", ActionCode: "create", ScopeCode: "global"}, authzService.RequireAction("scope", "create"), scopeHandler.Create)
		reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新作用域", ResourceCode: "scope", ActionCode: "update", ScopeCode: "global"}, authzService.RequireAction("scope", "update"), scopeHandler.Update)
		reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除作用域", ResourceCode: "scope", ActionCode: "delete", ScopeCode: "global"}, authzService.RequireAction("scope", "delete"), scopeHandler.Delete)
	}
}

func init() {
	module.GetRegistry().Register(&scopeModuleWrapper{})
}

type scopeModuleWrapper struct{}

func (w *scopeModuleWrapper) Init() error {
	return nil
}

func (w *scopeModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
