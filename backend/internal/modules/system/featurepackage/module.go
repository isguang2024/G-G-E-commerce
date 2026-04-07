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
		reg.GETProtected("/relationship-tree", reg.Meta("获取功能包关系树").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.GetRelationTree)
		reg.POSTProtected("/:id/rollback", reg.Meta("回滚功能包版本").BindPermissionKey("feature_package.manage").Build(), "feature_package.manage", authzService.RequireAction, handler.Rollback)
	}
}
