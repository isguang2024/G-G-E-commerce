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
		// GET /my、GET /current、GET /:id 已迁移到 OpenAPI-first 路径，
		// 由 router.go 中挂载的 ogen handler 接管。
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
