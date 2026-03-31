package navigation

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	menupkg "github.com/gg-ecommerce/backend/internal/modules/system/menu"
	pagepkg "github.com/gg-ecommerce/backend/internal/modules/system/page"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
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
	m.logger.Info("Initializing runtime navigation module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	menuRepo := user.NewMenuRepository(m.db)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	menuService := menupkg.NewMenuService(m.db, menuRepo, refresher, m.logger)
	pageService := pagepkg.NewService(m.db)
	spaceService := spacepkg.NewService(m.db, refresher, m.logger)
	handler := NewHandler(m.logger, NewService(m.db, menuService, pageService, spaceService))

	group := rg.Group("/runtime")
	reg := apiregistry.NewRegistrar(group, "navigation")
	{
		reg.GET("/navigation", reg.Meta("获取运行时导航清单").Build(), handler.GetNavigation)
	}
}

func init() {
	module.GetRegistry().Register(&moduleWrapper{})
}

type moduleWrapper struct{}

func (w *moduleWrapper) Init() error {
	return nil
}

func (w *moduleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
