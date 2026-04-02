package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
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
	tenantMemberRepo := user.NewTenantMemberRepository(m.db)
	authService := NewAuthService(userRepo, &m.config.JWT, m.logger)
	authzService := authorization.NewService(m.db, m.logger)
	authHandler := NewAuthHandler(authService, authzService, tenantMemberRepo, m.logger)

	auth := rg.Group("/auth")
	authReg := apiregistry.NewRegistrar(auth, "auth")
	{
		authReg.POST("/login", authReg.Meta("用户登录").Build(), authHandler.Login)
		authReg.POST("/register", authReg.Meta("用户注册").Build(), authHandler.Register)
		authReg.POST("/refresh", authReg.Meta("刷新访问令牌").Build(), authHandler.RefreshToken)
	}

	authenticated := rg.Group("")
	authenticated.Use(JWTAuth(m.config.JWT.Secret))
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
