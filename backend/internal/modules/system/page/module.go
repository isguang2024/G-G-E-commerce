package page

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
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
	service := NewService(m.db)
	handler := NewHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	group := rg.Group("/pages")
	reg := apiregistry.NewRegistrar(group, "page")
	{
		reg.GET("/runtime", reg.Meta("Get runtime pages").BindGroup("page").BindSource("sync").Build(), handler.ListRuntime)
		reg.GETProtected("/menu-options", reg.Meta("Get page menu options").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.ListMenuOptions)
		reg.GETProtected("/options", reg.Meta("Get page options").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.ListPageOptions)
		reg.GETProtected("/access-trace", reg.Meta("Get page access trace").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.GetAccessTrace)
		reg.GETProtected("/unregistered", reg.Meta("Get unregistered pages").BindGroup("page").BindPermissionKey("system.page.sync").Build(), "system.page.sync", authzService.RequireAction, handler.ListUnregistered)
		reg.POSTProtected("/sync", reg.Meta("Sync pages").BindGroup("page").BindSource("sync").BindPermissionKey("system.page.sync").Build(), "system.page.sync", authzService.RequireAction, handler.Sync)
		reg.GETProtected("", reg.Meta("List pages").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.List)
		reg.GETProtected("/:id/breadcrumb-preview", reg.Meta("Preview page breadcrumb").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.PreviewBreadcrumb)
		reg.GETProtected("/:id", reg.Meta("Get page detail").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Get)
		reg.POSTProtected("", reg.Meta("Create page").BindGroup("page").BindSource("manual").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Create)
		reg.PUTProtected("/:id", reg.Meta("Update page").BindGroup("page").BindSource("manual").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Update)
		reg.DELETEProtected("/:id", reg.Meta("Delete page").BindGroup("page").BindSource("manual").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Delete)
	}
}

func (m *Module) RegisterPublicRoutes(rg *gin.RouterGroup) {
	service := NewService(m.db)
	handler := NewHandler(service, m.logger)

	group := rg.Group("/pages")
	reg := apiregistry.NewRegistrar(group, "page")
	{
		reg.GET("/runtime/public", reg.Meta("Get public runtime pages").BindGroup("page").BindSource("sync").Build(), handler.ListRuntimePublic)
	}
}
