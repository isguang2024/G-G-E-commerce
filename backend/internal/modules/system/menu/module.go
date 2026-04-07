package menu

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type MenuModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewMenuModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *MenuModule {
	return &MenuModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *MenuModule) Init() error {
	m.logger.Info("Initializing Menu module")
	return nil
}

func (m *MenuModule) RegisterRoutes(rg *gin.RouterGroup) {
	_ = rg
	return
}

func init() {
	module.GetRegistry().Register(&menuModuleWrapper{})
}

type menuModuleWrapper struct{}

func (w *menuModuleWrapper) Init() error {
	return nil
}

func (w *menuModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
