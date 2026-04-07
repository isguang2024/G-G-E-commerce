package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type AuthModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewAuthModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *AuthModule {
	return &AuthModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *AuthModule) Init() error {
	m.logger.Info("Initializing Auth module")
	return nil
}

func (m *AuthModule) RegisterRoutes(rg *gin.RouterGroup) {
	// Auth domain is fully OpenAPI-first: /auth/login, /auth/register,
	// /auth/refresh, /auth/me are all served by handlers/auth.go via the
	// ogen bridge in router.go. Nothing left to mount here.
}

func init() {
	module.GetRegistry().Register(&authModuleWrapper{})
}

type authModuleWrapper struct{}

func (w *authModuleWrapper) Init() error {
	return nil
}

func (w *authModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}

func SetModuleInstance(m *AuthModule) {
	moduleInstance = m
}

var moduleInstance *AuthModule
