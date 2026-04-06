package featurepackage

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
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
	m.logger.Info("Initializing feature package module")
	return nil
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	packageRepo := user.NewFeaturePackageRepository(m.db)
	packageBundleRepo := user.NewFeaturePackageBundleRepository(m.db)
	packageActionRepo := user.NewFeaturePackageKeyRepository(m.db)
	packageMenuRepo := user.NewFeaturePackageMenuRepository(m.db)
	collaborationWorkspaceFeaturePackageRepo := user.NewCollaborationWorkspaceFeaturePackageRepository(m.db)
	rolePackageRepo := user.NewRoleFeaturePackageRepository(m.db)
	actionRepo := user.NewPermissionKeyRepository(m.db)
	menuRepo := user.NewMenuRepository(m.db)
	collaborationWorkspaceRepo := user.NewCollaborationWorkspaceRepository(m.db)
	boundaryService := collaborationworkspaceboundary.NewService(m.db)
	personalWorkspaceAccessService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, personalWorkspaceAccessService, roleSnapshotService)
	service := NewService(m.db, packageRepo, packageBundleRepo, packageActionRepo, packageMenuRepo, collaborationWorkspaceFeaturePackageRepo, rolePackageRepo, actionRepo, menuRepo, collaborationWorkspaceRepo, boundaryService, refresher)
	authzService := authorization.NewService(m.db, m.logger)
	handler := NewHandler(service, authzService, m.logger)

	group := rg.Group("/feature-packages")
	reg := apiregistry.NewRegistrar(group, "feature_package")
	{
		reg.GETProtected("", reg.Meta("获取功能包列表").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.List)
		reg.GETProtected("/options", reg.Meta("获取功能包候选").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.ListOptions)
		reg.GETProtected("/relationship-tree", reg.Meta("获取功能包关系树").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.GetRelationTree)
		reg.GETProtected("/:id/impact-preview", reg.Meta("获取功能包影响预览").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.GetImpactPreview)
		reg.GETProtected("/:id/versions", reg.Meta("获取功能包版本历史").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.ListVersions)
		reg.POSTProtected("/:id/rollback", reg.Meta("回滚功能包版本").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.Rollback)
		reg.GETProtected("/:id/risk-audits", reg.Meta("获取功能包最近变更").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.ListRiskAudits)
		reg.GETProtected("/:id", reg.Meta("获取功能包详情").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.Get)
		reg.POSTProtected("", reg.Meta("创建功能包").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.Create)
		reg.PUTProtected("/:id", reg.Meta("更新功能包").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.Update)
		reg.DELETEProtected("/:id", reg.Meta("删除功能包").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.Delete)
		reg.GETProtected("/:id/children", reg.Meta("获取组合包基础包").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.GetPackageChildren)
		reg.PUTProtected("/:id/children", reg.Meta("配置组合包基础包").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.SetPackageChildren)
		reg.GETProtected("/:id/actions", reg.Meta("获取功能包权限").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.GetPackageKeys)
		reg.PUTProtected("/:id/actions", reg.Meta("配置功能包权限").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.SetPackageKeys)
		reg.GETProtected("/:id/menus", reg.Meta("获取功能包菜单").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.GetPackageMenus)
		reg.PUTProtected("/:id/menus", reg.Meta("配置功能包菜单").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.SetPackageMenus)
		reg.GETProtected("/:id/collaboration-workspaces", reg.Meta("获取功能包协作空间").BindPermissionKey("feature_package.assign_collaboration_workspace").Build(), "feature_package.assign_collaboration_workspace", authzService.RequireAction, handler.GetPackageCollaborationWorkspaces)
		reg.PUTProtected("/:id/collaboration-workspaces", reg.Meta("配置功能包协作空间").BindPermissionKey("feature_package.assign_collaboration_workspace").Build(), "feature_package.assign_collaboration_workspace", authzService.RequireAction, handler.SetPackageCollaborationWorkspaces)
		reg.GETProtected("/collaboration-workspaces/:collaborationWorkspaceId", reg.Meta("获取协作空间功能包").BindPermissionKey("feature_package.assign_collaboration_workspace").Build(), "feature_package.assign_collaboration_workspace", authzService.RequireAction, handler.GetCollaborationWorkspacePackages)
		reg.PUTProtected("/collaboration-workspaces/:collaborationWorkspaceId", reg.Meta("配置协作空间功能包").BindPermissionKey("feature_package.assign_collaboration_workspace").Build(), "feature_package.assign_collaboration_workspace", authzService.RequireAction, handler.SetCollaborationWorkspacePackages)
	}
}
