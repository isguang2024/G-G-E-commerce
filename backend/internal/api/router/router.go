package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/middleware"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/apiendpoint"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
	"github.com/gg-ecommerce/backend/internal/modules/system/featurepackage"
	"github.com/gg-ecommerce/backend/internal/modules/system/media"
	"github.com/gg-ecommerce/backend/internal/modules/system/menu"
	"github.com/gg-ecommerce/backend/internal/modules/system/permission"
	"github.com/gg-ecommerce/backend/internal/modules/system/role"
	"github.com/gg-ecommerce/backend/internal/modules/system/system"
	"github.com/gg-ecommerce/backend/internal/modules/system/tenant"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
)

func SetupRouter(cfg *config.Config, logger *zap.Logger, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authModule := auth.NewAuthModule(db, cfg, logger)
	userModule := user.NewUserModule(db, cfg, logger)
	menuModule := menu.NewMenuModule(db, cfg, logger)
	permissionModule := permission.NewPermissionModule(db, cfg, logger)
	featurePackageModule := featurepackage.NewModule(db, cfg, logger)
	roleModule := role.NewRoleModule(db, cfg, logger)
	tenantModule := tenant.NewTenantModule(db, cfg, logger)
	mediaModule := media.NewMediaModule(db, cfg, logger)
	systemModule := system.NewSystemModule(db, cfg, logger)
	apiEndpointModule := apiendpoint.NewModule(db, cfg, logger, r)

	modules := module.GetRegistry().GetModules()
	for _, m := range modules {
		if err := m.Init(); err != nil {
			logger.Error("Failed to initialize module", zap.Error(err))
		}
	}

	v1 := r.Group("/api/v1")
	{
		authModule.RegisterRoutes(v1)

		authenticated := v1.Group("")
		authenticated.Use(auth.JWTAuth(cfg.JWT.Secret))
		{
			userModule.RegisterRoutes(authenticated)
			menuModule.RegisterRoutes(authenticated)
			permissionModule.RegisterRoutes(authenticated)
			featurePackageModule.RegisterRoutes(authenticated)
			roleModule.RegisterRoutes(authenticated)
			tenantModule.RegisterRoutes(authenticated)
			mediaModule.RegisterRoutes(authenticated)
			systemModule.RegisterRoutes(authenticated)
			apiEndpointModule.RegisterRoutes(authenticated)
		}

		open := r.Group("/open/v1")
		open.Use(middleware.APIKeyAuth())
		{
		}
	}

	if err := apiregistry.SyncRoutes(db, logger, r.Routes()); err != nil {
		logger.Error("Failed to sync API registry", zap.Error(err))
	}

	return r
}
