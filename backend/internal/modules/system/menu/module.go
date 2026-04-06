package menu

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

type MenuModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewMenuModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *MenuModule {
	return &MenuModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *MenuModule) Init() error {
	m.logger.Info("Initializing Menu module")
	return nil
}

func (m *MenuModule) RegisterRoutes(rg *gin.RouterGroup) {
	menuRepo := user.NewMenuRepository(m.db)
	userRepo := user.NewUserRepository(m.db)
	roleRepo := user.NewRoleRepository(m.db)
	userRoleRepo := user.NewUserRoleRepository(m.db)
	boundaryService := collaborationworkspaceboundary.NewService(m.db)
	personalWorkspaceAccessService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, personalWorkspaceAccessService, roleSnapshotService)
	menuService := NewMenuService(m.db, menuRepo, refresher, m.logger)
	authzService := authorization.NewService(m.db, m.logger)
	menuHandler := NewMenuHandler(m.db, menuService, userRepo, menuRepo, roleRepo, userRoleRepo, boundaryService, authzService, personalWorkspaceAccessService, m.logger)

	menus := rg.Group("/menus")
	reg := apiregistry.NewRegistrar(menus, "menu")
	{
		reg.GET("/tree", reg.Meta("获取菜单树").Build(), menuHandler.GetTree)
		reg.GETProtected("/:id/delete-preview", reg.Meta("获取菜单删除预览").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.DeletePreview)
		reg.POSTProtected("", reg.Meta("创建菜单").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.Create)
		reg.PUTProtected("/:id", reg.Meta("更新菜单").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.Update)
		reg.DELETEProtected("/:id", reg.Meta("删除菜单").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.Delete)
		reg.GETProtected("/groups", reg.Meta("获取菜单分组列表").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.ListGroups)
		reg.POSTProtected("/groups", reg.Meta("创建菜单分组").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.CreateGroup)
		reg.PUTProtected("/groups/:id", reg.Meta("更新菜单分组").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.UpdateGroup)
		reg.DELETEProtected("/groups/:id", reg.Meta("删除菜单分组").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuHandler.DeleteGroup)

		// 菜单备份相关路由
		backups := menus.Group("/backups")
		backupReg := apiregistry.NewRegistrar(backups, "menu_backup")
		{
			backupReg.POSTProtected("", backupReg.Meta("创建菜单备份").BindPermissionKey("system.menu.backup").Build(), "system.menu.backup", authzService.RequireAction, menuHandler.CreateBackup)
			backupReg.GETProtected("", backupReg.Meta("获取菜单备份列表").BindPermissionKey("system.menu.backup").Build(), "system.menu.backup", authzService.RequireAction, menuHandler.ListBackups)
			backupReg.DELETEProtected("/:id", backupReg.Meta("删除菜单备份").BindPermissionKey("system.menu.backup").Build(), "system.menu.backup", authzService.RequireAction, menuHandler.DeleteBackup)
			backupReg.POSTProtected("/:id/restore", backupReg.Meta("恢复菜单备份").BindPermissionKey("system.menu.backup").Build(), "system.menu.backup", authzService.RequireAction, menuHandler.RestoreBackup)
		}
	}
}

func init() {
	module.GetRegistry().Register(&menuModuleWrapper{})
}

type menuModuleWrapper struct{}

func (w *menuModuleWrapper) Init() error {
	return nil
}

func (w *menuModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
