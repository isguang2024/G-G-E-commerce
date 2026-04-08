package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type UserModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewUserModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *UserModule {
	return &UserModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *UserModule) Init() error {
	m.logger.Info("Initializing User module")
	return nil
}

func (m *UserModule) RegisterRoutes(rg *gin.RouterGroup) {
	// Phase 4: all user routes (including sub-routes) have been migrated to
	// ogen handlers in internal/api/handlers/. Nothing to register here.
	_ = rg
}

func init() {
	module.GetRegistry().Register(&userModuleWrapper{})
}

type userModuleWrapper struct{}

func (w *userModuleWrapper) Init() error {
	return nil
}

func (w *userModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
