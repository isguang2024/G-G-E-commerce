package page

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
)

type Module struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Module {
	return &Module{db: db, config: cfg, logger: logger}
}

func (m *Module) Init() error {
	m.logger.Info("Initializing page module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	menuRepo := user.NewMenuRepository(m.db)
	service := NewService(m.db, menuRepo)
	handler := NewHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	group := rg.Group("/pages")
	reg := apiregistry.NewRegistrar(group, "page")
	{
		reg.GETProtected("/access-trace", reg.Meta("Get page access trace").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.GetAccessTrace)
	}
}

func (m *Module) RegisterPublicRoutes(rg *gin.RouterGroup) {}
