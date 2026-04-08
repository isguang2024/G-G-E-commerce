package navigation

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
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
	m.logger.Info("Initializing runtime navigation module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// Phase 4: /runtime/navigation migrated to ogen handlers in
	// internal/api/handlers/navigation.go. Skip legacy gin registration.
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
