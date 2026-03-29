package media

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
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
	mediaHandler := NewMediaHandler()

	media := rg.Group("/media")
	reg := apiregistry.NewRegistrar(media, "media")
	{
		reg.POST("/upload", reg.Meta("上传媒体资源").Build(), mediaHandler.Upload)
		reg.GET("", reg.Meta("获取媒体资源列表").Build(), mediaHandler.List)
		reg.DELETE("/:id", reg.Meta("删除媒体资源").Build(), mediaHandler.Delete)
	}
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
