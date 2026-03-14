package role

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type RoleModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewRoleModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *RoleModule {
	return &RoleModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *RoleModule) Init() error {
	m.logger.Info("Initializing Role module")
	return nil
}

func (m *RoleModule) RegisterRoutes(rg *gin.RouterGroup) {
	roleRepo := user.NewRoleRepository(m.db)
	roleMenuRepo := user.NewRoleMenuRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	scopeRepo := user.NewScopeRepository(m.db)
	roleService := NewRoleService(roleRepo, roleMenuRepo, userRoleRepo, scopeRepo, m.logger)
	userRepo := user.NewUserRepository(m.db)
	roleHandler := NewRoleHandler(roleService, userRepo, m.logger)

	roles := rg.Group("/roles")
	{
		roles.GET("", roleHandler.List)
		roles.GET("/:id", roleHandler.Get)
		roles.GET("/:id/menus", roleHandler.GetRoleMenus)
		roles.PUT("/:id/menus", roleHandler.SetRoleMenus)
		roles.POST("", roleHandler.Create)
		roles.PUT("/:id", roleHandler.Update)
		roles.DELETE("/:id", roleHandler.Delete)
	}
}

func init() {
	module.GetRegistry().Register(&roleModuleWrapper{})
}

type roleModuleWrapper struct{}

func (w *roleModuleWrapper) Init() error {
	return nil
}

func (w *roleModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
