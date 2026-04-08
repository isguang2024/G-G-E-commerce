package router

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiendpointaccess"
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
	seed, err := permissionseed.LoadOpenAPISeed()
	if err != nil {
		logger.Fatal("openapi seed validation failed", zap.Error(err))
	}
	logger.Info("openapi seed loaded", zap.Int("operations", len(seed.Operations)))
	permLookup := seed.PermissionKeyByOperationID()

	// Phase 3 follow-up: consistency check. Every non-public/non-authenticated
	// operation must resolve to a permission_keys row — otherwise the evaluator
	// will silently deny at runtime. Warn-only so DB-less smoke boots keep
	// working; prod boot treats count > 0 as a release blocker.
	if missing := findMissingPermissionKeys(db, seed); len(missing) > 0 {
		logger.Warn("openapi seed references permission_keys missing from DB",
			zap.Strings("keys", missing),
			zap.Int("count", len(missing)))
	}

	// Phase 3: build the permission evaluator once and share across handlers.
	permEvaluator := evaluator.New(db, logger)

	endpointAccessService := apiendpointaccess.NewService(db, logger)

	// Build apiendpoint service early so it can be shared with the ogen handler.
	apiEndpointSvc := apiendpoint.NewService(
		db,
		user.NewAPIEndpointRepository(db),
		user.NewAPIEndpointCategoryRepository(db),
		user.NewAPIEndpointPermissionBindingRepository(db),
		r,
		logger,
		cfg.Env,
		endpointAccessService,
	)

	// Build the ogen server once. It handles all OpenAPI-first operations
	// (both public and authenticated). The Gin layer routes the public ones
	// outside the JWT middleware and the rest inside.
	permMW := middleware.OpenAPIPermission(permEvaluator, permLookup, logger)
	openapiErrHandler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, middleware.ErrPermissionDenied) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 403, "message": "无权访问"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 500, "message": err.Error()})
	}
	ogenServer, err := apigen.NewServer(
		handlers.NewAPIHandler(db, cfg, logger, permEvaluator, apiEndpointSvc),
		apigen.WithMiddleware(permMW),
		apigen.WithErrorHandler(openapiErrHandler),
	)
	if err != nil {
		logger.Fatal("failed to build ogen server", zap.Error(err))
	}
	ogenServeWith := func(c *gin.Context, withUser bool) {
		ctx := c.Request.Context()
		if withUser {
			ctx = context.WithValue(ctx, handlers.CtxUserID, c.GetString("user_id"))
			ctx = context.WithValue(ctx, handlers.CtxAuthWorkspaceID, c.GetString("auth_workspace_id"))
			ctx = context.WithValue(ctx, handlers.CtxAuthWorkspaceType, c.GetString("auth_workspace_type"))
			ctx = context.WithValue(ctx, handlers.CtxCollaborationWorkspaceID, c.GetString("collaboration_workspace_id"))
		}
		ctx = context.WithValue(ctx, handlers.CtxClientIP, c.ClientIP())
		req := c.Request.Clone(ctx)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		ogenServer.ServeHTTP(c.Writer, req)
	}
	publicBridge := func(c *gin.Context) { ogenServeWith(c, false) }
	ogenBridge := func(c *gin.Context) { ogenServeWith(c, true) }

	v1 := r.Group("/api/v1")
	v1.Use(endpointAccessService.RequireActiveEndpoint())
	{
		// Public (unauthenticated) operations served by the ogen bridge.
		v1.POST("/auth/login", publicBridge)
		v1.POST("/auth/register", publicBridge)
		v1.POST("/auth/refresh", publicBridge)
		v1.GET("/pages/runtime/public", publicBridge)

		authenticated := v1.Group("")
		authenticated.Use(auth.JWTAuth(cfg.JWT.Secret, db), middleware.AppContext(db))
		{
			authenticated.GET("/auth/me", ogenBridge)
			authenticated.POST("/workspaces/switch", ogenBridge)
			authenticated.GET("/workspaces/my", ogenBridge)
			authenticated.GET("/workspaces/current", ogenBridge)
			authenticated.GET("/workspaces/:id", ogenBridge)
			authenticated.GET("/permissions/explain", ogenBridge)

			// /users/* read + write + role assignment
			authenticated.GET("/users", ogenBridge)
			authenticated.POST("/users", ogenBridge)
			authenticated.GET("/users/:id", ogenBridge)
			authenticated.PUT("/users/:id", ogenBridge)
			authenticated.DELETE("/users/:id", ogenBridge)
			authenticated.POST("/users/:id/roles", ogenBridge)

			// role + navigation
			authenticated.GET("/roles", ogenBridge)
			authenticated.POST("/roles", ogenBridge)
			authenticated.GET("/roles/options", ogenBridge)
			authenticated.GET("/roles/:id", ogenBridge)
			authenticated.PUT("/roles/:id", ogenBridge)
			authenticated.DELETE("/roles/:id", ogenBridge)
			authenticated.GET("/roles/:id/packages", ogenBridge)
			authenticated.PUT("/roles/:id/packages", ogenBridge)
			authenticated.GET("/roles/:id/menus", ogenBridge)
			authenticated.PUT("/roles/:id/menus", ogenBridge)
			authenticated.GET("/roles/:id/actions", ogenBridge)
			authenticated.PUT("/roles/:id/actions", ogenBridge)
			authenticated.GET("/roles/:id/data-permissions", ogenBridge)
			authenticated.PUT("/roles/:id/data-permissions", ogenBridge)
			authenticated.GET("/runtime/navigation", ogenBridge)

			// feature-package
			authenticated.GET("/feature-packages/relationship-tree", ogenBridge)
			authenticated.POST("/feature-packages/:id/rollback", ogenBridge)
			authenticated.GET("/feature-packages", ogenBridge)
			authenticated.GET("/feature-packages/options", ogenBridge)
			authenticated.POST("/feature-packages", ogenBridge)
			authenticated.GET("/feature-packages/:id", ogenBridge)
			authenticated.PUT("/feature-packages/:id", ogenBridge)
			authenticated.DELETE("/feature-packages/:id", ogenBridge)
			authenticated.GET("/feature-packages/:id/children", ogenBridge)
			authenticated.PUT("/feature-packages/:id/children", ogenBridge)
			authenticated.GET("/feature-packages/:id/actions", ogenBridge)
			authenticated.PUT("/feature-packages/:id/actions", ogenBridge)
			authenticated.GET("/feature-packages/:id/menus", ogenBridge)
			authenticated.PUT("/feature-packages/:id/menus", ogenBridge)
			authenticated.GET("/feature-packages/:id/collaboration-workspaces", ogenBridge)
			authenticated.PUT("/feature-packages/:id/collaboration-workspaces", ogenBridge)
			authenticated.GET("/feature-packages/:id/impact-preview", ogenBridge)
			authenticated.GET("/feature-packages/:id/versions", ogenBridge)
			authenticated.GET("/feature-packages/:id/risk-audits", ogenBridge)
			authenticated.GET("/feature-packages/collaboration-workspaces/:collaborationWorkspaceId", ogenBridge)
			authenticated.PUT("/feature-packages/collaboration-workspaces/:collaborationWorkspaceId", ogenBridge)

			// permission
			authenticated.GET("/permission-actions", ogenBridge)
			authenticated.GET("/permission-actions/options", ogenBridge)
			authenticated.GET("/permission-actions/:id", ogenBridge)
			authenticated.GET("/permission-actions/:id/endpoints", ogenBridge)
			authenticated.GET("/permission-actions/:id/consumers", ogenBridge)
			authenticated.GET("/permission-actions/:id/impact-preview", ogenBridge)
			authenticated.GET("/permission-actions/groups", ogenBridge)
			authenticated.GET("/permission-actions/risk-audits", ogenBridge)
			authenticated.GET("/permission-actions/templates", ogenBridge)
			authenticated.POST("/permission-actions", ogenBridge)
			authenticated.POST("/permission-actions/:id/endpoints", ogenBridge)
			authenticated.POST("/permission-actions/cleanup-unused", ogenBridge)
			authenticated.PUT("/permission-actions/:id", ogenBridge)
			authenticated.DELETE("/permission-actions/:id", ogenBridge)
			authenticated.DELETE("/permission-actions/:id/endpoints/:endpointCode", ogenBridge)
			authenticated.POST("/permission-actions/batch", ogenBridge)
			authenticated.POST("/permission-actions/templates", ogenBridge)
			authenticated.POST("/permission-actions/groups", ogenBridge)
			authenticated.PUT("/permission-actions/groups/:id", ogenBridge)

			// menu
			authenticated.GET("/menus/tree", ogenBridge)
			authenticated.POST("/menus", ogenBridge)
			authenticated.PUT("/menus/:id", ogenBridge)
			authenticated.DELETE("/menus/:id", ogenBridge)
			authenticated.GET("/menus/groups", ogenBridge)
			authenticated.POST("/menus/groups", ogenBridge)
			authenticated.PUT("/menus/groups/:id", ogenBridge)
			authenticated.DELETE("/menus/groups/:id", ogenBridge)
			authenticated.GET("/menus/backups", ogenBridge)
			authenticated.POST("/menus/backups", ogenBridge)
			authenticated.DELETE("/menus/backups/:id", ogenBridge)
			authenticated.GET("/menus/:id/delete-preview", ogenBridge)
			authenticated.POST("/menus/backups/:id/restore", ogenBridge)

			// page
			authenticated.GET("/pages", ogenBridge)
			authenticated.GET("/pages/options", ogenBridge)
			authenticated.GET("/pages/menu-options", ogenBridge)
			authenticated.GET("/pages/runtime", ogenBridge)
			authenticated.GET("/pages/unregistered", ogenBridge)
			authenticated.GET("/pages/:id/breadcrumb-preview", ogenBridge)
			authenticated.GET("/pages/:id", ogenBridge)
			authenticated.POST("/pages", ogenBridge)
			authenticated.POST("/pages/sync", ogenBridge)
			authenticated.PUT("/pages/:id", ogenBridge)
			authenticated.DELETE("/pages/:id", ogenBridge)
			authenticated.GET("/pages/access-trace", ogenBridge)

			// collaboration-workspace
			authenticated.GET("/collaboration-workspaces", ogenBridge)
			authenticated.GET("/collaboration-workspaces/options", ogenBridge)
			authenticated.GET("/collaboration-workspaces/:id", ogenBridge)
			authenticated.POST("/collaboration-workspaces", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/:id", ogenBridge)
			authenticated.DELETE("/collaboration-workspaces/:id", ogenBridge)
			authenticated.GET("/collaboration-workspaces/:id/members", ogenBridge)
			authenticated.POST("/collaboration-workspaces/:id/members", ogenBridge)
			authenticated.DELETE("/collaboration-workspaces/:id/members/:userId", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/:id/members/:userId/role", ogenBridge)
			authenticated.GET("/collaboration-workspaces/mine", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/members", ogenBridge)
			authenticated.POST("/collaboration-workspaces/current/members", ogenBridge)
			authenticated.DELETE("/collaboration-workspaces/current/members/:userId", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/current/members/:userId/role", ogenBridge)

			// system app + menu-space
			authenticated.GET("/system/apps", ogenBridge)
			authenticated.POST("/system/apps", ogenBridge)
			authenticated.GET("/system/apps/current", ogenBridge)
			authenticated.GET("/system/app-host-bindings", ogenBridge)
			authenticated.POST("/system/app-host-bindings", ogenBridge)
			authenticated.GET("/system/menu-spaces", ogenBridge)
			authenticated.POST("/system/menu-spaces", ogenBridge)
			authenticated.GET("/system/menu-spaces/current", ogenBridge)
			authenticated.GET("/system/menu-space-mode", ogenBridge)
			authenticated.PUT("/system/menu-space-mode", ogenBridge)
			authenticated.POST("/system/menu-spaces/:spaceKey/initialize-default", ogenBridge)
			authenticated.GET("/system/menu-space-host-bindings", ogenBridge)
			authenticated.POST("/system/menu-space-host-bindings", ogenBridge)

			// system fast-enter + view-pages
			authenticated.GET("/system/view-pages", ogenBridge)
			authenticated.GET("/system/fast-enter", ogenBridge)
			authenticated.PUT("/system/fast-enter", ogenBridge)

			// message
			authenticated.GET("/messages/inbox/summary", ogenBridge)
			authenticated.GET("/messages/inbox", ogenBridge)
			authenticated.GET("/messages/inbox/:deliveryId", ogenBridge)
			authenticated.POST("/messages/inbox/:deliveryId/read", ogenBridge)
			authenticated.POST("/messages/inbox/read-all", ogenBridge)
			authenticated.POST("/messages/inbox/:deliveryId/todo-action", ogenBridge)
			authenticated.GET("/messages/dispatch/options", ogenBridge)
			authenticated.POST("/messages/dispatch", ogenBridge)
			authenticated.GET("/messages/templates", ogenBridge)
			authenticated.POST("/messages/templates", ogenBridge)
			authenticated.PUT("/messages/templates/:templateId", ogenBridge)
			authenticated.GET("/messages/senders", ogenBridge)
			authenticated.POST("/messages/senders", ogenBridge)
			authenticated.PUT("/messages/senders/:senderId", ogenBridge)
			authenticated.GET("/messages/recipient-groups", ogenBridge)
			authenticated.POST("/messages/recipient-groups", ogenBridge)
			authenticated.PUT("/messages/recipient-groups/:groupId", ogenBridge)
			authenticated.GET("/messages/records", ogenBridge)
			authenticated.GET("/messages/records/:recordId", ogenBridge)

			// user sub-routes — collaboration workspaces + refresh
			authenticated.GET("/users/:id/collaboration-workspaces", ogenBridge)
			authenticated.POST("/users/:id/permission-refresh", ogenBridge)

			// user sub-routes — menus / packages / permissions / diagnosis
			authenticated.GET("/users/:id/menus", ogenBridge)
			authenticated.PUT("/users/:id/menus", ogenBridge)
			authenticated.GET("/users/:id/packages", ogenBridge)
			authenticated.PUT("/users/:id/packages", ogenBridge)
			authenticated.GET("/users/:id/permissions", ogenBridge)
			authenticated.GET("/users/:id/permission-diagnosis", ogenBridge)

			// collaboration-workspace boundary — current workspace complex ops
			authenticated.GET("/collaboration-workspaces/current/roles", ogenBridge)
			authenticated.POST("/collaboration-workspaces/current/roles", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/boundary/roles", ogenBridge)
			authenticated.POST("/collaboration-workspaces/current/boundary/roles", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/current/boundary/roles/:roleId", ogenBridge)
			authenticated.DELETE("/collaboration-workspaces/current/boundary/roles/:roleId", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/boundary/roles/:roleId/packages", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/current/boundary/roles/:roleId/packages", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/boundary/roles/:roleId/menus", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/current/boundary/roles/:roleId/menus", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/boundary/roles/:roleId/actions", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/current/boundary/roles/:roleId/actions", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/boundary/packages", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/menus", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/menu-origins", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/actions", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/action-origins", ogenBridge)
			authenticated.GET("/collaboration-workspaces/current/members/:userId/roles", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/current/members/:userId/roles", ogenBridge)
			// collaboration-workspace boundary — workspace-scoped ops
			authenticated.GET("/collaboration-workspaces/:id/roles", ogenBridge)
			authenticated.GET("/collaboration-workspaces/:id/menus", ogenBridge)
			authenticated.GET("/collaboration-workspaces/:id/menu-origins", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/:id/menus", ogenBridge)
			authenticated.GET("/collaboration-workspaces/:id/actions", ogenBridge)
			authenticated.GET("/collaboration-workspaces/:id/action-origins", ogenBridge)
			authenticated.PUT("/collaboration-workspaces/:id/actions", ogenBridge)

			// ── api-endpoints ─────────────────────────────────────────────
			authenticated.GET("/api-endpoints", ogenBridge)
			authenticated.POST("/api-endpoints", ogenBridge)
			authenticated.GET("/api-endpoints/overview", ogenBridge)
			authenticated.GET("/api-endpoints/stale", ogenBridge)
			authenticated.POST("/api-endpoints/sync", ogenBridge)
			authenticated.POST("/api-endpoints/cleanup-stale", ogenBridge)
			authenticated.GET("/api-endpoints/unregistered", ogenBridge)
			authenticated.GET("/api-endpoints/categories", ogenBridge)
			authenticated.PUT("/api-endpoints/categories/:id", ogenBridge)
			authenticated.PUT("/api-endpoints/:id", ogenBridge)
			authenticated.PUT("/api-endpoints/:id/context-scope", ogenBridge)

			// ── media ─────────────────────────────────────────────────────
			authenticated.POST("/media/upload", ogenBridge)
			authenticated.GET("/media", ogenBridge)
			authenticated.DELETE("/media/:id", ogenBridge)
		}

		open := r.Group("/open/v1")
		open.Use(endpointAccessService.RequireActiveEndpoint(), middleware.APIKeyAuth(db))
		{
		}
	}

	if err := apiendpoint.SyncRoutes(db, logger, r.Routes()); err != nil {
		logger.Error("Failed to sync API registry", zap.Error(err))
	}
	if err := endpointAccessService.Refresh(); err != nil {
		logger.Error("Failed to refresh API endpoint runtime cache", zap.Error(err))
	}

	return r
}

// findMissingPermissionKeys returns permission_key strings referenced by
// the OpenAPI seed but absent from the permission_keys table. Used by the
// startup consistency check — warn-only, so DB-less unit/CI smoke boots
// don't fail just because the seed row hasn't been inserted yet.
func findMissingPermissionKeys(db *gorm.DB, seed *permissionseed.OpenAPISeed) []string {
	if db == nil || seed == nil {
		return nil
	}
	wanted := make(map[string]struct{})
	for _, op := range seed.Operations {
		if op.PermissionKey == "" {
			continue
		}
		wanted[op.PermissionKey] = struct{}{}
	}
	if len(wanted) == 0 {
		return nil
	}
	keys := make([]string, 0, len(wanted))
	for k := range wanted {
		keys = append(keys, k)
	}
	var existing []string
	if err := db.Raw(
		`SELECT permission_key FROM permission_keys WHERE permission_key IN ? AND deleted_at IS NULL`,
		keys,
	).Scan(&existing).Error; err != nil {
		return nil
	}
	have := make(map[string]struct{}, len(existing))
	for _, k := range existing {
		have[k] = struct{}{}
	}
	missing := make([]string, 0)
	for k := range wanted {
		if _, ok := have[k]; !ok {
			missing = append(missing, k)
		}
	}
	return missing
}
