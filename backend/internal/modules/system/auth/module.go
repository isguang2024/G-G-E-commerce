package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	workspacepkg "github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
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
	userRepo := user.NewUserRepository(m.db)
	collaborationWorkspaceMemberRepo := user.NewCollaborationWorkspaceMemberRepository(m.db)
	workspaceService := workspacepkg.NewService(m.db, m.logger)
	authService := NewAuthService(userRepo, &m.config.JWT, m.logger)
	authzService := authorization.NewService(m.db, m.logger)
	authHandler := NewAuthHandler(authService, authzService, collaborationWorkspaceMemberRepo, workspaceService, m.logger)

	// /auth/login、/auth/register、/auth/refresh 已全部迁移到 OpenAPI-first
	// （router.go 中由 ogen handler 接管），不再在此挂载 public 路由组。

	authenticated := rg.Group("")
	authenticated.Use(JWTAuth(m.config.JWT.Secret, m.db))
	authenticatedReg := apiregistry.NewRegistrar(authenticated, "auth")
	{
		authenticatedReg.GET("/user/info", authenticatedReg.Meta("获取当前登录用户信息").Build(), authHandler.GetUserInfo)
	}
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
