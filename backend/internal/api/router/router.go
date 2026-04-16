package router

import (
	"context"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	apigen "github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/api/apperr"
	"github.com/maben/backend/internal/api/handlers"
	"github.com/maben/backend/internal/api/middleware"
	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/telemetry"
	"github.com/maben/backend/internal/modules/system/apiendpoint"
	"github.com/maben/backend/internal/modules/system/auth"
	"github.com/maben/backend/internal/modules/system/register"
	"github.com/maben/backend/internal/modules/system/social"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/apiendpointaccess"
	pkgLogger "github.com/maben/backend/internal/pkg/logger"
	"github.com/maben/backend/internal/pkg/openapidocs"
	"github.com/maben/backend/internal/pkg/permission/evaluator"
	"github.com/maben/backend/internal/pkg/permissionseed"
)

// SetupRouter 构建主 HTTP 路由。
//
// 参数顺序：cfg → logger → db → auditRecorder → telemetryIngester。
// 两个 observability 组件都不允许为 nil —— 关闭时请分别传 audit.Noop{} /
// telemetry.Noop{}，避免下游 handler 判空遗漏。
//
// 中间件挂载顺序的约定（从上到下严格）：
//  1. RequestID：产生/回显 X-Request-Id，并把它写进 gin.Context 与 request.Context；
//     必须是 #1，因为后续 Logger / Recovery / 审计都依赖这个字段做 join key。
//  2. Logger：access log，读 request_id + app/space/auth 标签，链路级打点。
//  3. Recovery：兜底 panic，出错也能带上 request_id 写进日志便于回溯。
//  4. AppContext → DynamicAppSecurity：解析 app_key / space_key / auth_mode。
func SetupRouter(cfg *config.Config, logger *zap.Logger, db *gorm.DB, auditRecorder audit.Recorder, telemetryIngester telemetry.Ingester) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	if auditRecorder == nil {
		auditRecorder = audit.Noop{}
	}
	if telemetryIngester == nil {
		telemetryIngester = telemetry.Noop{}
	}

	r := gin.New()

	// 约定：RequestID 必须挂在第一条，后续中间件的日志字段 / ctx 都依赖它。
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.AppContext(db))
	r.Use(middleware.DynamicAppSecurity(db, logger, cfg.Env))
	r.Static("/uploads", "./data/uploads")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	registerAppHealthRoutes(r, db, logger)

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
	permMW := middleware.OpenAPIPermission(permEvaluator, permLookup, logger, auditRecorder)
	ogenServer, err := apigen.NewServer(
		handlers.NewAPIHandler(db, cfg, logger, permEvaluator, apiEndpointSvc, auditRecorder, telemetryIngester),
		handlers.SecurityHandler{},
		apigen.WithMiddleware(permMW),
		apigen.WithErrorHandler(apperr.ErrorHandler(logger)),
	)
	if err != nil {
		logger.Fatal("failed to build ogen server", zap.Error(err))
	}
	// ogenServeWith 把 gin.Context 上已经落位的字段透出到 request.Context，
	// 同时往 pkgLogger 的类型化 ctx 里注入 request_id / actor / tenant / app /
	// workspace / client 信息 —— 下游任何 handler 拿到 ctx 都可以：
	//   - logger.With(ctx).Info(...) 得到带全链路字段的结构化日志；
	//   - auditRecorder.Record(ctx, ...) 自动补齐 audit 行上的身份/租户列。
	ogenServeWith := func(c *gin.Context, withUser bool) {
		ctx := c.Request.Context()
		if withUser {
			userID := c.GetString("user_id")
			ctx = context.WithValue(ctx, handlers.CtxUserID, userID)
			ctx = context.WithValue(ctx, handlers.CtxAuthWorkspaceID, c.GetString("auth_workspace_id"))
			ctx = context.WithValue(ctx, handlers.CtxAuthWorkspaceType, c.GetString("auth_workspace_type"))
			ctx = context.WithValue(ctx, handlers.CtxCollaborationWorkspaceID, c.GetString("collaboration_workspace_id"))
			if authTime, exists := c.Get("auth_time"); exists {
				ctx = context.WithValue(ctx, handlers.CtxAuthTime, authTime)
			}
			ctx = pkgLogger.WithActor(ctx, userID, "user")
			ctx = pkgLogger.WithWorkspace(ctx, c.GetString("collaboration_workspace_id"))
		}
		ctx = context.WithValue(ctx, handlers.CtxClientIP, c.ClientIP())
		requestHost := strings.TrimSpace(c.GetString("request_host"))
		if requestHost == "" && c.Request != nil {
			requestHost = c.Request.Host
		}
		ctx = context.WithValue(ctx, handlers.CtxRequestHost, requestHost)
		ctx = context.WithValue(ctx, handlers.CtxRequestPath, c.Request.URL.Path)
		ctx = context.WithValue(ctx, handlers.CtxUserAgent, c.Request.UserAgent())

		// pkgLogger 类型化 ctx：租户目前固定 default；app_key 来自 AppContext；
		// IP / UA 放到独立键，审计 row 直接读即可。
		if tenant := strings.TrimSpace(c.GetString("tenant_id")); tenant != "" {
			ctx = pkgLogger.WithTenant(ctx, tenant)
		} else {
			ctx = pkgLogger.WithTenant(ctx, "default")
		}
		ctx = pkgLogger.WithApp(ctx, c.GetString("app_key"))
		ctx = pkgLogger.WithClient(ctx, c.ClientIP(), c.Request.UserAgent())

		req := c.Request.Clone(ctx)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		ogenServer.ServeHTTP(c.Writer, req)
	}
	publicBridge := func(c *gin.Context) { ogenServeWith(c, false) }
	ogenBridge := func(c *gin.Context) { ogenServeWith(c, true) }
	socialSvc := social.NewService(
		db,
		auth.NewAuthService(user.NewUserRepository(db), &cfg.JWT, logger),
		user.NewUserRepository(db),
		register.NewResolver(register.NewRepository(db)),
		cfg.JWT.Secret,
		logger,
	)
	socialHandler := social.NewHTTPHandler(socialSvc, logger)
	r.GET("/api/v1/auth/oauth/:provider/authorize", socialHandler.Authorize)
	r.GET("/api/v1/auth/oauth/:provider/callback", socialHandler.Callback)

	v1 := r.Group("/api/v1")
	v1.Use(endpointAccessService.RequireActiveEndpoint())
	{
		authenticated := v1.Group("")
		authenticated.Use(auth.JWTAuth(cfg.JWT.Secret, db), middleware.AppContext(db))

		// /api/v1 下所有业务路由由 OpenAPI seed 驱动批量挂载。
		//
		// 新增/删除 operation 只需改 api/openapi/*.yaml 然后 `make api`，
		// seed 会被重新生成并嵌入二进制；mountOpenAPIBridgeRoutes 自动把每条
		// operation 桥接到 ogen Server —— router.go 不再需要为每条 API 手工
		// 新增行。这避免了历史上漏桥接导致前端 404 的那类问题。
		//
		// 仍然需要手工注册的（已在 router 上方独立挂载）：
		//   - /api/v1/auth/oauth/:provider/{authorize,callback}：socialHandler 自
		//     实现，不是 OpenAPI 操作，不会出现在 seed 里。
		//   - /health、/health/*、/uploads：不在 /api/v1 路径下，启动前就要可用。
		mountOpenAPIBridgeRoutes(v1, authenticated, seed.Operations, publicBridge, ogenBridge, logger)

		open := r.Group("/open/v1")
		open.Use(endpointAccessService.RequireActiveEndpoint(), middleware.APIKeyAuth(db))
		{
		}
	}

	if _, err := apiendpoint.SyncRoutes(db, logger, r.Routes()); err != nil {
		logger.Error("Failed to sync API registry", zap.Error(err))
	}
	if err := endpointAccessService.Refresh(); err != nil {
		logger.Error("Failed to refresh API endpoint runtime cache", zap.Error(err))
	}

	return r
}

// mountOpenAPIBridgeRoutes registers every operation in `ops` onto the
// appropriate Gin router group, dispatching based on access_mode:
//
//   - "public"                        → v1 group (no JWT)
//   - "authenticated" / "permission"  → authenticated group (JWT + perm MW)
//   - anything else                   → authenticated, with a warning log
//
// Ops are registered in (path, method) ascending order. Gin's radix tree
// requires static path segments to be registered before parametric ones
// at the same level (e.g. `/audit-logs/stats` before `/audit-logs/{id}`).
// Since '{' (ASCII 123) sorts after all letters and digits, alphabetical
// ordering naturally satisfies this constraint.
//
// Extracted so both SetupRouter and the router_contract_test can share the
// exact same registration logic.
func mountOpenAPIBridgeRoutes(
	v1 *gin.RouterGroup,
	authenticated *gin.RouterGroup,
	ops []permissionseed.OpenAPIOperation,
	publicBridge gin.HandlerFunc,
	ogenBridge gin.HandlerFunc,
	logger *zap.Logger,
) {
	sorted := append([]permissionseed.OpenAPIOperation(nil), ops...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Path != sorted[j].Path {
			return sorted[i].Path < sorted[j].Path
		}
		return sorted[i].Method < sorted[j].Method
	})
	for _, op := range sorted {
		method := strings.ToUpper(strings.TrimSpace(op.Method))
		ginPath := openapiPathToGin(op.Path)
		switch op.AccessMode {
		case "public":
			v1.Handle(method, ginPath, publicBridge)
		case "authenticated", "permission":
			authenticated.Handle(method, ginPath, ogenBridge)
		default:
			if logger != nil {
				logger.Warn("openapi seed: unknown access_mode, defaulting to permission",
					zap.String("access_mode", op.AccessMode),
					zap.String("operation_id", op.OperationID),
					zap.String("method", method),
					zap.String("path", ginPath))
			}
			authenticated.Handle(method, ginPath, ogenBridge)
		}
	}
}

// openapiPathToGin translates an OpenAPI-style path ("/users/{id}/packages/{packageId}")
// into Gin's colon-prefixed placeholder form ("/users/:id/packages/:packageId").
// Used by the seed-driven route registration loop. Non-brace characters are
// copied verbatim; an unmatched '{' is also copied verbatim rather than
// silently dropped, so any upstream spec bug is visible in the final path.
func openapiPathToGin(p string) string {
	if !strings.ContainsRune(p, '{') {
		return p
	}
	var b strings.Builder
	b.Grow(len(p))
	for i := 0; i < len(p); i++ {
		if p[i] != '{' {
			b.WriteByte(p[i])
			continue
		}
		end := strings.IndexByte(p[i:], '}')
		if end <= 0 {
			b.WriteByte(p[i])
			continue
		}
		b.WriteByte(':')
		b.WriteString(p[i+1 : i+end])
		i += end
	}
	return b.String()
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
