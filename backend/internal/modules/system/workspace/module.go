package workspace

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
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
	// 全部 workspace 路由已迁移到 OpenAPI-first（router.go 中由 ogen handler 接管）。
	// 模块保留以参与 module.Registry 生命周期，未来本域新增的非 OpenAPI 端点可挂这里。
	_ = rg
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
