package system

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	cachepkg "github.com/gg-ecommerce/backend/internal/pkg/cache"
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
	var systemCache *cachepkg.Cache
	systemCache, cacheErr := cachepkg.NewCache(
		m.config.Redis.Host,
		m.config.Redis.Port,
		m.config.Redis.Password,
		m.config.Redis.DB,
	)
	if cacheErr != nil {
		m.logger.Warn("Redis cache unavailable, page-association cache disabled", zap.Error(cacheErr))
	}

	systemHandler := NewSystemHandler(m.logger, systemCache)

	system := rg.Group("/system")
	{
		system.GET("/view-pages", systemHandler.GetViewPages)
	}
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
