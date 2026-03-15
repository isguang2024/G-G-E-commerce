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
	userRepo := NewUserRepository(m.db)
	roleRepo := NewRoleRepository(m.db)
	menuRepo := NewMenuRepository(m.db)
	roleMenuRepo := NewRoleMenuRepository(m.db)
	userRoleRepo := NewUserRoleRepository(m.db)
	userService := NewUserService(userRepo, roleRepo, m.logger)
	permissionService := NewPermissionService(userRepo, userRoleRepo, roleMenuRepo)
	userHandler := NewUserHandler(userService, permissionService, menuRepo, m.logger)

	users := rg.Group("/users")
	{
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.Get)
		users.GET("/:id/permissions", userHandler.GetPermissions)
		users.POST("", userHandler.Create)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
		users.POST("/:id/roles", userHandler.AssignRoles)
	}
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
