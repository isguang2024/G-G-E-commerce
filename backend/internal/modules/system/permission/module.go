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
		reg.GETProtected("", reg.Meta("获取功能权限列表").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.List)
		reg.GETProtected("/options", reg.Meta("获取功能权限候选").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.ListOptions)
		reg.GETProtected("/groups", reg.Meta("获取功能权限分组列表").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.ListGroups)
		reg.GETProtected("/:id", reg.Meta("获取功能权限详情").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.Get)
		reg.GETProtected("/:id/endpoints", reg.Meta("获取功能权限关联接口").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.ListEndpoints)
		reg.GETProtected("/:id/consumers", reg.Meta("获取功能权限消费明细").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.GetConsumerDetails)
		reg.GETProtected("/:id/impact-preview", reg.Meta("获取功能权限影响预览").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.GetImpactPreview)
		reg.GETProtected("/risk-audits", reg.Meta("获取功能权限最近变更").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.ListRiskAudits)
		reg.POSTProtected("/cleanup-unused", reg.Meta("清理未消费功能权限").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.CleanupUnused)
		reg.POSTProtected("/batch", reg.Meta("批量治理功能权限").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.BatchUpdate)
		reg.GETProtected("/templates", reg.Meta("获取功能权限批量模板").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.ListBatchTemplates)
		reg.POSTProtected("/templates", reg.Meta("保存功能权限批量模板").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.SaveBatchTemplate)
		reg.POSTProtected("/:id/endpoints", reg.Meta("新增功能权限关联接口").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.AddEndpoint)
		reg.DELETEProtected("/:id/endpoints/:endpointCode", reg.Meta("删除功能权限关联接口").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.RemoveEndpoint)
		reg.POSTProtected("/groups", reg.Meta("创建功能权限分组").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.CreateGroup)
		reg.PUTProtected("/groups/:id", reg.Meta("更新功能权限分组").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.UpdateGroup)
		reg.POSTProtected("", reg.Meta("创建功能权限").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.Create)
		reg.PUTProtected("/:id", reg.Meta("更新功能权限").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.Update)
		reg.DELETEProtected("/:id", reg.Meta("删除功能权限").BindPermissionKey("system.permission.manage").Build(), "system.permission.manage", authzService.RequireAction, handler.Delete)
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
