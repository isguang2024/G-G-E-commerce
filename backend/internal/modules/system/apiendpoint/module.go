package apiendpoint

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/apiendpointaccess"
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
	// Routes migrated to ogen (Phase 5) — registered in router.go via ogenBridge.
	_ = rg
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
