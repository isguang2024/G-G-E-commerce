package permission

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
)

type PermissionModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewPermissionModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *PermissionModule {
	return &PermissionModule{db: db, config: cfg, logger: logger}
}

func (m *PermissionModule) Init() error {
	m.logger.Info("Initializing Permission module")
	return nil
}

func (m *PermissionModule) RegisterRoutes(rg *gin.RouterGroup) {
	actionRepo := user.NewPermissionKeyRepository(m.db)
	groupRepo := user.NewPermissionGroupRepository(m.db)
	apiEndpointRepo := user.NewAPIEndpointRepository(m.db)
	apiEndpointBindingRepo := user.NewAPIEndpointPermissionBindingRepository(m.db)
	packageActionRepo := user.NewFeaturePackageKeyRepository(m.db)
	collaborationWorkspaceFeaturePackageRepo := user.NewCollaborationWorkspaceFeaturePackageRepository(m.db)
	roleDisabledActionRepo := user.NewRoleDisabledActionRepository(m.db)
	collaborationWorkspaceBlockedActionRepo := user.NewCollaborationWorkspaceBlockedActionRepository(m.db)
	userActionRepo := user.NewUserActionPermissionRepository(m.db)
	boundaryService := collaborationworkspaceboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	service := NewPermissionService(m.db, groupRepo, actionRepo, apiEndpointRepo, apiEndpointBindingRepo, packageActionRepo, collaborationWorkspaceFeaturePackageRepo, roleDisabledActionRepo, collaborationWorkspaceBlockedActionRepo, userActionRepo, boundaryService, refresher)
	handler := NewPermissionHandler(service, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	actions := rg.Group("/permission-actions")
	reg := apiregistry.NewRegistrar(actions, "permission_key")
	{
		reg.POSTProtected("/batch", reg.Meta("批量治理功能权限").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.BatchUpdate)
		reg.POSTProtected("/templates", reg.Meta("保存功能权限批量模板").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.SaveBatchTemplate)
		reg.POSTProtected("/groups", reg.Meta("创建功能权限分组").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.CreateGroup)
		reg.PUTProtected("/groups/:id", reg.Meta("更新功能权限分组").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.UpdateGroup)
	}
}

func init() {
	module.GetRegistry().Register(&permissionModuleWrapper{})
}

type permissionModuleWrapper struct{}

func (w *permissionModuleWrapper) Init() error {
	return nil
}

func (w *permissionModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
