package scope

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
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

	scopes := rg.Group("/scopes")
	{
		scopes.GET("", scopeHandler.List)
		scopes.GET("/all", scopeHandler.GetAll)
		scopes.GET("/:id", scopeHandler.Get)
		scopes.POST("", scopeHandler.Create)
		scopes.PUT("/:id", scopeHandler.Update)
		scopes.DELETE("/:id", scopeHandler.Delete)
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
