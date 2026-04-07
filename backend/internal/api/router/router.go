package router

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	apigen "github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/handlers"
	"github.com/gg-ecommerce/backend/internal/api/middleware"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/apiendpoint"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
	"github.com/gg-ecommerce/backend/internal/modules/system/featurepackage"
	"github.com/gg-ecommerce/backend/internal/modules/system/media"
	"github.com/gg-ecommerce/backend/internal/modules/system/menu"
	"github.com/gg-ecommerce/backend/internal/modules/system/navigation"
	"github.com/gg-ecommerce/backend/internal/modules/system/page"
	"github.com/gg-ecommerce/backend/internal/modules/system/permission"
	"github.com/gg-ecommerce/backend/internal/modules/system/role"
	"github.com/gg-ecommerce/backend/internal/modules/system/system"
	collaborationworkspace "github.com/gg-ecommerce/backend/internal/modules/system/collaborationworkspace"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	"github.com/gg-ecommerce/backend/internal/pkg/apiendpointaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/openapidocs"
	"github.com/gg-ecommerce/backend/internal/pkg/permission/evaluator"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
)

func SetupRouter(cfg *config.Config, logger *zap.Logger, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.AppContext(db))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Phase 1 cleanup: serve the embedded OpenAPI spec + Swagger UI.
	openapidocs.Mount(r)

	// Phase 1 cleanup: validate the gen-permissions seed early so a missing
	// or malformed permission key fails the boot rather than at first request.
	if seed, err := permissionseed.LoadOpenAPISeed(); err != nil {
		logger.Fatal("openapi seed validation failed", zap.Error(err))
	} else {
		logger.Info("openapi seed loaded", zap.Int("operations", len(seed.Operations)))
	}

	// Phase 3: build the permission evaluator once and share across handlers.
	permEvaluator := evaluator.New(db, logger)

	endpointAccessService := apiendpointaccess.NewService(db, logger)

	authModule := auth.NewAuthModule(db, cfg, logger)
	userModule := user.NewUserModule(db, cfg, logger)
	menuModule := menu.NewMenuModule(db, cfg, logger)
	navigationModule := navigation.NewModule(db, cfg, logger)
	pageModule := page.NewModule(db, cfg, logger)
	permissionModule := permission.NewPermissionModule(db, cfg, logger)
	featurePackageModule := featurepackage.NewModule(db, cfg, logger)
	roleModule := role.NewRoleModule(db, cfg, logger)
	collaborationWorkspaceModule := collaborationworkspace.NewCollaborationWorkspaceModule(db, cfg, logger)
	workspaceModule := workspace.NewModule(db, cfg, logger)
	mediaModule := media.NewMediaModule(db, cfg, logger)
	systemModule := system.NewSystemModule(db, cfg, logger)
	apiEndpointModule := apiendpoint.NewModule(db, cfg, logger, r, endpointAccessService)

	modules := module.GetRegistry().GetModules()
	for _, m := range modules {
		if err := m.Init(); err != nil {
			logger.Error("Failed to initialize module", zap.Error(err))
		}
	}

	v1 := r.Group("/api/v1")
	v1.Use(endpointAccessService.RequireActiveEndpoint())
	{
		authModule.RegisterRoutes(v1)
		pageModule.RegisterPublicRoutes(v1)

		authenticated := v1.Group("")
		authenticated.Use(auth.JWTAuth(cfg.JWT.Secret, db), middleware.AppContext(db))
		{
			// OpenAPI-first 路径：由 ogen 生成的 server 直接处理。
			// 入口在 Gin 中以普通路由声明，便于复用 JWT/AppContext 中间件，
			// 之后再把请求 ctx 注入 user_id 后转交给 ogen handler。
			ogenServer, err := apigen.NewServer(handlers.NewAPIHandler(db, logger, permEvaluator))
			if err != nil {
				logger.Fatal("failed to build ogen server", zap.Error(err))
			}
			ogenBridge := func(c *gin.Context) {
				userID := c.GetString("user_id")
				req := c.Request.Clone(context.WithValue(c.Request.Context(), handlers.CtxUserID, userID))
				req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
				ogenServer.ServeHTTP(c.Writer, req)
			}
			// OpenAPI-first 路径：legacy /workspaces/{my,current,:id} 与
			// /permissions/explain 全部由 ogen handler 接管。
			authenticated.GET("/workspaces/my", ogenBridge)
			authenticated.GET("/workspaces/current", ogenBridge)
			authenticated.GET("/workspaces/:id", ogenBridge)
			authenticated.GET("/permissions/explain", ogenBridge)

			userModule.RegisterRoutes(authenticated)
			menuModule.RegisterRoutes(authenticated)
			navigationModule.RegisterRoutes(authenticated)
			pageModule.RegisterRoutes(authenticated)
			permissionModule.RegisterRoutes(authenticated)
			featurePackageModule.RegisterRoutes(authenticated)
			roleModule.RegisterRoutes(authenticated)
			collaborationWorkspaceModule.RegisterRoutes(authenticated)
			workspaceModule.RegisterRoutes(authenticated)
			mediaModule.RegisterRoutes(authenticated)
			systemModule.RegisterRoutes(authenticated)
			apiEndpointModule.RegisterRoutes(authenticated)
		}

		open := r.Group("/open/v1")
		open.Use(endpointAccessService.RequireActiveEndpoint(), middleware.APIKeyAuth(db))
		{
		}
	}

	if err := apiregistry.SyncRoutes(db, logger, r.Routes()); err != nil {
		logger.Error("Failed to sync API registry", zap.Error(err))
	}
	if err := endpointAccessService.Refresh(); err != nil {
		logger.Error("Failed to refresh API endpoint runtime cache", zap.Error(err))
	}

	return r
}
