package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/handler"
	"github.com/gg-ecommerce/backend/internal/api/middleware"
	"github.com/gg-ecommerce/backend/internal/config"
	cachepkg "github.com/gg-ecommerce/backend/internal/pkg/cache"
	"github.com/gg-ecommerce/backend/internal/repository"
	"github.com/gg-ecommerce/backend/internal/service"
)

// SetupRouter 初始化路由
func SetupRouter(cfg *config.Config, logger *zap.Logger, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关（无需 JWT）
		auth := v1.Group("/auth")
		{
			// 初始化 Repository 和 Service
			userRepo := repository.NewUserRepository(db)
			authService := service.NewAuthService(userRepo, &cfg.JWT, logger)
			authHandler := handler.NewAuthHandler(authService, logger)
			
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 需要认证的路由
		authenticated := v1.Group("")
		authenticated.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// 媒体相关
			media := authenticated.Group("/media")
			{
				mediaHandler := handler.NewMediaHandler()
				media.POST("/upload", mediaHandler.Upload)
				media.GET("", mediaHandler.List)
				media.DELETE("/:id", mediaHandler.Delete)
			}

			// 用户信息
			userRepo := repository.NewUserRepository(db)
			authService := service.NewAuthService(userRepo, &cfg.JWT, logger)
			authHandler := handler.NewAuthHandler(authService, logger)
			authenticated.GET("/user/info", authHandler.GetUserInfo)

			// 用户管理（后台）
			roleRepo := repository.NewRoleRepository(db)
			roleMenuRepo := repository.NewRoleMenuRepository(db)
			menuRepo := repository.NewMenuRepository(db)
			permissionService := service.NewPermissionService(userRepo, roleMenuRepo, db)
			userService := service.NewUserService(userRepo, roleRepo, logger)
			userHandler := handler.NewUserHandler(userService, permissionService, menuRepo, logger)
			users := authenticated.Group("/users")
			{
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.Get)
				users.GET("/:id/permissions", userHandler.GetPermissions)
				users.POST("", userHandler.Create)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
				users.PUT("/:id/roles", userHandler.AssignRoles)
			}

			// 角色、角色菜单、用户角色
			userRoleRepo := repository.NewUserRoleRepository(db)
			scopeRepo := repository.NewScopeRepository(db)
			roleService := service.NewRoleService(roleRepo, roleMenuRepo, userRoleRepo, scopeRepo, logger)
			// Redis 缓存仅用于系统「页面关联/视图枚举」；不可用时为 nil，接口回退直查
			var systemCache *cachepkg.Cache
			systemCache, cacheErr := cachepkg.NewCache(
				cfg.Redis.Host,
				cfg.Redis.Port,
				cfg.Redis.Password,
				cfg.Redis.DB,
			)
			if cacheErr != nil {
				logger.Warn("Redis cache unavailable, page-association cache disabled", zap.Error(cacheErr))
			}

			// 团队管理（后台）
			tenantRepo := repository.NewTenantRepository(db)
			tenantMemberRepo := repository.NewTenantMemberRepository(db)
			tenantService := service.NewTenantService(tenantRepo, tenantMemberRepo, userRepo, roleRepo, userRoleRepo, logger)
			tenantHandler := handler.NewTenantHandler(tenantService, tenantMemberRepo, userRepo, roleRepo, userRoleRepo, logger)
			tenants := authenticated.Group("/tenants")
			{
				// 「我的团队」接口（team_admin 管理本团队成员，须在 /:id 前注册）
				tenants.GET("/my-team", tenantHandler.GetMyTeam)
				tenants.GET("/my-team/members", tenantHandler.ListMyMembers)
				tenants.POST("/my-team/members", tenantHandler.AddMyMember)
				tenants.DELETE("/my-team/members/:userId", tenantHandler.RemoveMyMember)
				tenants.PUT("/my-team/members/:userId/role", tenantHandler.UpdateMyMemberRole)
				tenants.GET("/my-team/members/:userId/roles", tenantHandler.GetMyTeamMemberRoles)
				tenants.PUT("/my-team/members/:userId/roles", tenantHandler.SetMyTeamMemberRoles)
				// 「我的团队」角色列表（仅全局 scope=team 角色，不再支持团队自建角色）
				tenants.GET("/my-team/roles", tenantHandler.ListMyTeamRoles)
				// 后台团队 CRUD
				tenants.GET("", tenantHandler.List)
				tenants.GET("/:id", tenantHandler.Get)
				tenants.POST("", tenantHandler.Create)
				tenants.PUT("/:id", tenantHandler.Update)
				tenants.DELETE("/:id", tenantHandler.Delete)
				tenants.GET("/:id/members", tenantHandler.ListMembers)
				tenants.POST("/:id/members", tenantHandler.AddMember)
				tenants.DELETE("/:id/members/:userId", tenantHandler.RemoveMember)
				tenants.PUT("/:id/members/:userId/role", tenantHandler.UpdateMemberRole)
			}

			// 作用域管理（后台）
			scopeService := service.NewScopeService(scopeRepo, roleRepo, logger)
			scopeHandler := handler.NewScopeHandler(scopeService, logger)
			scopes := authenticated.Group("/scopes")
			{
				scopes.GET("", scopeHandler.List)
				scopes.GET("/all", scopeHandler.GetAll)
				scopes.GET("/:id", scopeHandler.Get)
				scopes.POST("", scopeHandler.Create)
				scopes.PUT("/:id", scopeHandler.Update)
				scopes.DELETE("/:id", scopeHandler.Delete)
			}

			// 角色管理（后台）+ 角色-菜单权限（复用上方 roleService）
			roleHandler := handler.NewRoleHandler(roleService, userRepo, logger)
			roles := authenticated.Group("/roles")
			{
				roles.GET("", roleHandler.List)
				roles.GET("/:id", roleHandler.Get)
				roles.GET("/:id/menus", roleHandler.GetRoleMenus)
				roles.PUT("/:id/menus", roleHandler.SetRoleMenus)
				roles.POST("", roleHandler.Create)
				roles.PUT("/:id", roleHandler.Update)
				roles.DELETE("/:id", roleHandler.Delete)
			}

			// 菜单（树形 + CRUD）
			menuRepo = repository.NewMenuRepository(db)
			menuService := service.NewMenuService(menuRepo, logger)
			menuHandler := handler.NewMenuHandler(menuService, userRepo, roleMenuRepo, userRoleRepo, tenantMemberRepo, logger)
			menus := authenticated.Group("/menus")
			{
				menus.GET("/tree", menuHandler.GetTree)
				menus.POST("", menuHandler.Create)
				menus.PUT("/:id", menuHandler.Update)
				menus.DELETE("/:id", menuHandler.Delete)
				menus.PUT("/sort", menuHandler.UpdateSort)
				menus.PUT("/sort-by-parent", menuHandler.UpdateSortByParentID)
			}

			// 系统管理（隐藏页面文件保存、视图枚举等；仅视图枚举使用 Redis 缓存）
			systemHandler := handler.NewSystemHandler(logger, systemCache)
			system := authenticated.Group("/system")
			{
				system.POST("/save-page-association-file", systemHandler.SavePageAssociationFile)
				system.GET("/get-hidden-index-file", systemHandler.GetHiddenIndexFile)
				system.GET("/view-pages", systemHandler.GetViewPages)
			}
		}

		// 对外开放 API（API Key 认证）
		open := r.Group("/open/v1")
		open.Use(middleware.APIKeyAuth())
		{
			// TODO: 可以在这里添加其他对外开放的 API
		}
	}

	return r
}
