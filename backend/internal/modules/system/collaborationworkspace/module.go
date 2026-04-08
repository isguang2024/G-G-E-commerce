package collaborationworkspace

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type CollaborationWorkspaceModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewCollaborationWorkspaceModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *CollaborationWorkspaceModule {
	return &CollaborationWorkspaceModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *CollaborationWorkspaceModule) Init() error {
	m.logger.Info("Initializing CollaborationWorkspace module")
	return nil
}

func (m *CollaborationWorkspaceModule) RegisterRoutes(rg *gin.RouterGroup) {
	// Phase 4: all /collaboration-workspaces/* routes migrated to ogen handlers.
	// router.go owns all registrations via ogenBridge.
	_ = rg
}

func init() {
	module.GetRegistry().Register(&collaborationWorkspaceModuleWrapper{})
}

type collaborationWorkspaceModuleWrapper struct{}

func (w *collaborationWorkspaceModuleWrapper) Init() error {
	return nil
}

func (w *collaborationWorkspaceModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
