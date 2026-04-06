package workspace

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
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
	m.logger.Info("Initializing Workspace module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	service := NewService(m.db, m.logger)
	handler := NewHandler(m.logger, service)

	workspaces := rg.Group("/workspaces")
	reg := apiregistry.NewRegistrar(workspaces, "workspace")
	{
		reg.GET("/my", reg.Meta("获取我的工作空间列表").BindContextScope("optional").Build(), handler.ListMine)
		reg.GET("/current", reg.Meta("获取当前授权工作空间").BindContextScope("optional").Build(), handler.GetCurrent)
		reg.GET("/:id", reg.Meta("获取工作空间详情").BindContextScope("optional").Build(), handler.Get)
		reg.POST("/switch", reg.Meta("切换当前授权工作空间").BindContextScope("optional").Build(), handler.Switch)
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
