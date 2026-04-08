package media

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type MediaModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewMediaModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *MediaModule {
	return &MediaModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *MediaModule) Init() error {
	m.logger.Info("Initializing Media module")
	return nil
}

func (m *MediaModule) RegisterRoutes(rg *gin.RouterGroup) {
	// Routes migrated to ogen (Phase 5) — registered in router.go via ogenBridge.
	_ = rg
}

func init() {
	module.GetRegistry().Register(&mediaModuleWrapper{})
}

type mediaModuleWrapper struct{}

func (w *mediaModuleWrapper) Init() error {
	return nil
}

func (w *mediaModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
