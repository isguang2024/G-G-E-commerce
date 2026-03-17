package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

type UserModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewUserModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *UserModule {
	return &UserModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *UserModule) Init() error {
	m.logger.Info("Initializing User module")
	return nil
}

func (m *UserModule) RegisterRoutes(rg *gin.RouterGroup) {
	userRepo := NewUserRepository(m.db)
	roleRepo := NewRoleRepository(m.db)
	menuRepo := NewMenuRepository(m.db)
	roleMenuRepo := NewRoleMenuRepository(m.db)
	userRoleRepo := NewUserRoleRepository(m.db)
	actionRepo := NewPermissionActionRepository(m.db)
	userActionRepo := NewUserActionPermissionRepository(m.db)
	userService := NewUserService(userRepo, roleRepo, m.logger)
	permissionService := NewPermissionService(userRepo, userRoleRepo, roleMenuRepo)
	userHandler := NewUserHandler(userService, permissionService, actionRepo, userActionRepo, menuRepo, m.logger)
	authzService := authorization.NewService(m.db, m.logger)

	users := rg.Group("/users")
	reg := apiregistry.NewRegistrar(users, "user")
	{
	reg.GET("", &apiregistry.RouteMeta{Summary: "获取用户列表", ResourceCode: "user", ActionCode: "list", ScopeCode: "global"}, authzService.RequireAction("user", "list", "global"), userHandler.List)
	reg.GET("/:id", &apiregistry.RouteMeta{Summary: "获取用户详情", ResourceCode: "user", ActionCode: "get", ScopeCode: "global"}, authzService.RequireAction("user", "get", "global"), userHandler.Get)
	reg.GET("/:id/permissions", &apiregistry.RouteMeta{Summary: "获取用户菜单权限", ResourceCode: "user", ActionCode: "get", ScopeCode: "global"}, authzService.RequireAction("user", "get", "global"), userHandler.GetPermissions)
	reg.GET("/:id/actions", &apiregistry.RouteMeta{Summary: "获取用户功能权限", ResourceCode: "user", ActionCode: "assign_action", ScopeCode: "global"}, authzService.RequireAction("user", "assign_action", "global"), userHandler.GetActions)
	reg.PUT("/:id/actions", &apiregistry.RouteMeta{Summary: "配置用户功能权限", ResourceCode: "user", ActionCode: "assign_action", ScopeCode: "global"}, authzService.RequireAction("user", "assign_action", "global"), userHandler.SetActions)
	reg.POST("", &apiregistry.RouteMeta{Summary: "创建用户", ResourceCode: "user", ActionCode: "create", ScopeCode: "global"}, authzService.RequireAction("user", "create", "global"), userHandler.Create)
	reg.PUT("/:id", &apiregistry.RouteMeta{Summary: "更新用户", ResourceCode: "user", ActionCode: "update", ScopeCode: "global"}, authzService.RequireAction("user", "update", "global"), userHandler.Update)
	reg.DELETE("/:id", &apiregistry.RouteMeta{Summary: "删除用户", ResourceCode: "user", ActionCode: "delete", ScopeCode: "global"}, authzService.RequireAction("user", "delete", "global"), userHandler.Delete)
	reg.POST("/:id/roles", &apiregistry.RouteMeta{Summary: "分配用户角色", ResourceCode: "user", ActionCode: "assign_role", ScopeCode: "global"}, authzService.RequireAction("user", "assign_role", "global"), userHandler.AssignRoles)
	}
}

func init() {
	module.GetRegistry().Register(&userModuleWrapper{})
}

type userModuleWrapper struct{}

func (w *userModuleWrapper) Init() error {
	return nil
}

func (w *userModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
