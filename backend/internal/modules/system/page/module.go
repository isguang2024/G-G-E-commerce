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
		reg.GETProtected("/menu-options", reg.Meta("获取页面上级菜单候选").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.ListMenuOptions)
		reg.GETProtected("/options", reg.Meta("获取页面候选").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.ListPageOptions)
		reg.GETProtected("/unregistered", reg.Meta("获取未注册页面").BindGroup("page").BindPermissionKey("system.page.sync").Build(), "system.page.sync", authzService.RequireAction, handler.ListUnregistered)
		reg.POSTProtected("/sync", reg.Meta("同步页面注册表").BindGroup("page").BindSource("sync").BindPermissionKey("system.page.sync").Build(), "system.page.sync", authzService.RequireAction, handler.Sync)
		reg.GETProtected("", reg.Meta("获取页面列表").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.List)
		reg.GETProtected("/:id/breadcrumb-preview", reg.Meta("预览页面面包屑").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.PreviewBreadcrumb)
		reg.GETProtected("/:id", reg.Meta("获取页面详情").BindGroup("page").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Get)
		reg.POSTProtected("", reg.Meta("创建页面").BindGroup("page").BindSource("manual").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Create)
		reg.PUTProtected("/:id", reg.Meta("更新页面").BindGroup("page").BindSource("manual").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Update)
		reg.DELETEProtected("/:id", reg.Meta("删除页面").BindGroup("page").BindSource("manual").BindPermissionKey("system.page.manage").Build(), "system.page.manage", authzService.RequireAction, handler.Delete)
	}
}

func (m *Module) RegisterPublicRoutes(rg *gin.RouterGroup) {
	service := NewService(m.db)
	handler := NewHandler(service, m.logger)

	group := rg.Group("/pages")
	reg := apiregistry.NewRegistrar(group, "page")
	{
		reg.GET("/runtime", reg.Meta("获取运行时页面注册表").BindGroup("page").BindSource("sync").Build(), handler.ListRuntime)
		reg.GET("/runtime/public", reg.Meta("获取公开运行时页面注册表").BindGroup("page").BindSource("sync").Build(), handler.ListRuntimePublic)
	}
}
