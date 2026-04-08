package system

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type SystemModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewSystemModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *SystemModule {
	return &SystemModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *SystemModule) Init() error {
	m.logger.Info("Initializing System module")
	return nil
}

func (m *SystemModule) RegisterRoutes(rg *gin.RouterGroup) {
	// Phase 4: all /system/* and /messages/* routes migrated to ogen handlers.
	_ = rg
}

func init() {
	module.GetRegistry().Register(&systemModuleWrapper{})
}

type systemModuleWrapper struct{}

func (w *systemModuleWrapper) Init() error {
	return nil
}

func (w *systemModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
