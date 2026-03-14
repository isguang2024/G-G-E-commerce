package menu

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
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
	menuRepo := user.NewMenuRepository(m.db)
	menuService := NewMenuService(menuRepo, m.logger)
	userRepo := user.NewUserRepository(m.db)
	roleMenuRepo := user.NewRoleMenuRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	tenantMemberRepo := user.NewTenantMemberRepository(m.db)
	menuHandler := NewMenuHandler(menuService, userRepo, roleMenuRepo, userRoleRepo, tenantMemberRepo, m.logger)

	menus := rg.Group("/menus")
	{
		menus.GET("/tree", menuHandler.GetTree)
		menus.POST("", menuHandler.Create)
		menus.PUT("/:id", menuHandler.Update)
		menus.DELETE("/:id", menuHandler.Delete)
	}
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
