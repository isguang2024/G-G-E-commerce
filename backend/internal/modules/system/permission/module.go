package permission

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type PermissionModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewPermissionModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *PermissionModule {
	return &PermissionModule{db: db, config: cfg, logger: logger}
}

func (m *PermissionModule) Init() error {
	m.logger.Info("Initializing Permission module")
	return nil
}

func (m *PermissionModule) RegisterRoutes(rg *gin.RouterGroup) {
	_ = rg
	return
}

func init() {
	module.GetRegistry().Register(&permissionModuleWrapper{})
}

type permissionModuleWrapper struct{}

func (w *permissionModuleWrapper) Init() error {
	return nil
}

func (w *permissionModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
