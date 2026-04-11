// Package handlers_test provides smoke tests for the ogen bridge layer.
// These tests verify that HTTP requests correctly route through ogen →
// APIHandler and return expected status codes and response shapes.
//
// Strategy: build a minimal gin engine directly (bypassing SetupRouter which
// requires a live DB) with the real ogen server and a nil-db APIHandler.
// Tests that truly require a live DB call t.Skip.
package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	apigen "github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/handlers"
	"github.com/gg-ecommerce/backend/internal/api/middleware"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/apiendpoint"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiendpointaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/permission/evaluator"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
)

// noopEvaluator implements evaluator.Evaluator. Can always returns true so the
// ogen permission middleware never blocks; Resolve/Explain return empty results.
type noopEvaluator struct{}

func (e *noopEvaluator) Resolve(_ context.Context, accountID, workspaceID uuid.UUID) (*evaluator.ResolvedPermissions, error) {
	return &evaluator.ResolvedPermissions{
		AccountID:   accountID,
		WorkspaceID: workspaceID,
		Keys:        make(map[string]struct{}),
	}, nil
}

func (e *noopEvaluator) Can(_ context.Context, _, _ uuid.UUID, _ string) (bool, error) {
	return true, nil
}

func (e *noopEvaluator) Explain(_ context.Context, accountID, workspaceID uuid.UUID) (*evaluator.Explanation, error) {
	return &evaluator.Explanation{
		Resolved: &evaluator.ResolvedPermissions{
			AccountID:   accountID,
			WorkspaceID: workspaceID,
			Keys:        make(map[string]struct{}),
		},
		FeaturePackageKeys: map[string][]uuid.UUID{},
		RoleKeys:           map[string][]uuid.UUID{},
	}, nil
}

// nilAPIEndpointSvc stubs apiendpoint.Service so NewAPIHandler compiles
// without a real DB or router.
type nilAPIEndpointSvc struct{}

func (s nilAPIEndpointSvc) List(_ *apiendpoint.ListRequest) ([]user.APIEndpoint, int64, error) {
	return nil, 0, nil
}
func (s nilAPIEndpointSvc) Overview(_ string) (*apiendpoint.EndpointOverview, error) {
	return &apiendpoint.EndpointOverview{}, nil
}
func (s nilAPIEndpointSvc) ListRuntimeStates(_ []user.APIEndpoint) map[uuid.UUID]apiendpoint.EndpointRuntimeState {
	return nil
}
func (s nilAPIEndpointSvc) ListStale(_ *apiendpoint.StaleListRequest) ([]user.APIEndpoint, int64, error) {
	return nil, 0, nil
}
func (s nilAPIEndpointSvc) ListUnregisteredRoutes(_ *apiendpoint.UnregisteredRouteListRequest) ([]apiendpoint.UnregisteredRouteItem, int64, error) {
	return nil, 0, nil
}
func (s nilAPIEndpointSvc) GetUnregisteredScanConfig() (apiendpoint.UnregisteredScanConfig, error) {
	return apiendpoint.UnregisteredScanConfig{}, nil
}
func (s nilAPIEndpointSvc) SaveUnregisteredScanConfig(cfg apiendpoint.UnregisteredScanConfig) (apiendpoint.UnregisteredScanConfig, error) {
	return cfg, nil
}
func (s nilAPIEndpointSvc) ListBindingsByEndpointCodes(_ []string) ([]user.APIEndpointPermissionBinding, error) {
	return nil, nil
}
func (s nilAPIEndpointSvc) ListBindings(_ string) ([]user.APIEndpointPermissionBinding, error) {
	return nil, nil
}
func (s nilAPIEndpointSvc) ListCategories() ([]user.APIEndpointCategory, error) {
	return nil, nil
}
func (s nilAPIEndpointSvc) Save(_ *user.APIEndpoint, _ []string, _ string) (*user.APIEndpoint, error) {
	return nil, nil
}
func (s nilAPIEndpointSvc) SaveCategory(_ *user.APIEndpointCategory) (*user.APIEndpointCategory, error) {
	return nil, nil
}
func (s nilAPIEndpointSvc) UpdateContextScope(_ uuid.UUID, _ string) (*user.APIEndpoint, error) {
	return nil, nil
}
func (s nilAPIEndpointSvc) Sync() error                                       { return nil }
func (s nilAPIEndpointSvc) CleanupStale(_ []uuid.UUID, _ string) (int, error) { return 0, nil }

// testEngine builds a minimal gin engine wired to the real ogen server.
// db is nil — DB-dependent handlers will fail at runtime; auth-layer and
// ogen-routing tests work without a DB.
func testEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()

	cfg := &config.Config{
		Env: "test",
		JWT: config.JWTConfig{Secret: "test-secret-for-smoke-tests"},
	}

	// Load the embedded permission seed — same as production boot.
	seed, err := permissionseed.LoadOpenAPISeed()
	if err != nil {
		t.Fatalf("permissionseed.LoadOpenAPISeed: %v", err)
	}
	permLookup := seed.PermissionKeyByOperationID()

	eval := &noopEvaluator{}

	// nil-db endpoint access service — RequireActiveEndpoint passes through
	// when the routeMap is empty (no endpoints loaded from DB).
	endpointSvc := apiendpointaccess.NewService(nil, logger)

	ogenHandler := handlers.NewAPIHandler(nil, cfg, logger, eval, nilAPIEndpointSvc{})

	permMW := middleware.OpenAPIPermission(eval, permLookup, logger)
	ogenServer, err := apigen.NewServer(
		ogenHandler,
		handlers.SecurityHandler{},
		apigen.WithMiddleware(permMW),
	)
	if err != nil {
		t.Fatalf("apigen.NewServer: %v", err)
	}

	// ogenServe mirrors the production bridge: strips /api/v1 prefix.
	ogenBridgePublic := func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), handlers.CtxClientIP, c.ClientIP())
		req := c.Request.Clone(ctx)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		ogenServer.ServeHTTP(c.Writer, req)
	}
	ogenBridgeAuth := func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), handlers.CtxUserID, c.GetString("user_id"))
		ctx = context.WithValue(ctx, handlers.CtxAuthWorkspaceID, c.GetString("auth_workspace_id"))
		ctx = context.WithValue(ctx, handlers.CtxAuthWorkspaceType, c.GetString("auth_workspace_type"))
		ctx = context.WithValue(ctx, handlers.CtxCollaborationWorkspaceID, c.GetString("collaboration_workspace_id"))
		ctx = context.WithValue(ctx, handlers.CtxClientIP, c.ClientIP())
		req := c.Request.Clone(ctx)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		ogenServer.ServeHTTP(c.Writer, req)
	}

	r := gin.New()
	r.Use(middleware.AppContext(nil)) // nil db → no-op
	r.Use(endpointSvc.RequireActiveEndpoint())

	v1 := r.Group("/api/v1")
	{
		// Public routes (no JWT required)
		v1.POST("/auth/login", ogenBridgePublic)
		v1.POST("/auth/register", ogenBridgePublic)
		v1.POST("/auth/refresh", ogenBridgePublic)
		v1.POST("/auth/callback/exchange", ogenBridgePublic)

		// Authenticated routes guarded by JWT
		authed := v1.Group("")
		authed.Use(auth.JWTAuth(cfg.JWT.Secret, nil))
		{
			authed.GET("/runtime/navigation", ogenBridgeAuth)
			authed.GET("/roles", ogenBridgeAuth)
			authed.GET("/menus/tree", ogenBridgeAuth)
			authed.GET("/feature-packages", ogenBridgeAuth)
			authed.GET("/collaboration-workspaces", ogenBridgeAuth)
			authed.GET("/auth/me", ogenBridgeAuth)
		}
	}

	return r
}

// apiResponseShape is the minimal JSON envelope shared by error responses.
type apiResponseShape struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func doRequest(t *testing.T, r *gin.Engine, method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --------------------------------------------------------------------------
// TestSmoke* — all tests in this group must pass in CI without a live DB.
// --------------------------------------------------------------------------

// TestSmokeLoginEmptyBodyReturns400 verifies the ogen handler returns 400
// when username/password fail validation.
func TestSmokeLoginEmptyBodyReturns400(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodPost, "/api/v1/auth/login",
		[]byte(`{"username":"","password":""}`), nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d — body: %s", w.Code, w.Body.String())
	}
	// ogen validation errors use error_message field
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v — body: %s", err, w.Body.String())
	}
	if len(resp) == 0 {
		t.Errorf("want non-empty error body, got empty map")
	}
}

// TestSmokeLoginNoBodyReturns400 verifies ogen returns 400 when the request
// body is completely absent.
func TestSmokeLoginNoBodyReturns400(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodPost, "/api/v1/auth/login", nil, map[string]string{
		"Content-Type": "application/json",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeRolesWithoutTokenReturns401 verifies the JWT middleware blocks
// unauthenticated access to /roles before reaching the ogen handler.
func TestSmokeRolesWithoutTokenReturns401(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/roles", nil, nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeMenusTreeWithoutTokenReturns401 verifies the JWT middleware blocks
// unauthenticated access to /menus/tree.
func TestSmokeMenusTreeWithoutTokenReturns401(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/menus/tree", nil, nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeNavigationWithoutTokenReturns401 verifies the JWT middleware blocks
// unauthenticated access to /runtime/navigation.
func TestSmokeNavigationWithoutTokenReturns401(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/runtime/navigation", nil, nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeFeaturePackagesWithoutTokenReturns401 verifies the JWT middleware
// blocks unauthenticated access to /feature-packages.
func TestSmokeFeaturePackagesWithoutTokenReturns401(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/feature-packages", nil, nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeCollaborationWorkspacesWithoutTokenReturns401 verifies that
// /collaboration-workspaces requires authentication.
func TestSmokeCollaborationWorkspacesWithoutTokenReturns401(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/collaboration-workspaces", nil, nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeProtectedRouteWithBadTokenReturns401 verifies that a malformed JWT
// on any protected route returns 401, not 500.
func TestSmokeProtectedRouteWithBadTokenReturns401(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/roles", nil, map[string]string{
		"Authorization": "Bearer this-is-not-a-valid-jwt",
	})

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeLoginResponseShape verifies the 400 response from /auth/login
// is valid JSON with at least one field (ogen uses error_message).
func TestSmokeLoginResponseShape(t *testing.T) {
	r := testEngine(t)
	body := []byte(`{"username":"","password":""}`)
	w := doRequest(t, r, http.MethodPost, "/api/v1/auth/login", body, nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response not JSON: %v — body: %s", err, w.Body.String())
	}
	// ogen returns {error_message: "..."} for validation failures
	if _, ok := resp["error_message"]; !ok {
		// fall back: any non-empty error body is acceptable
		if len(resp) == 0 {
			t.Errorf("response body is empty — want error details, got: %s", w.Body.String())
		}
	}
}

// TestSmokePermissionSeedLoads verifies the embedded openapi_seed.json parses
// cleanly and contains at least one operation.
func TestSmokePermissionSeedLoads(t *testing.T) {
	seed, err := permissionseed.LoadOpenAPISeed()
	if err != nil {
		t.Fatalf("LoadOpenAPISeed: %v", err)
	}
	if len(seed.Operations) == 0 {
		t.Fatal("seed has no operations")
	}
	t.Logf("seed loaded %d operations", len(seed.Operations))

	lookup := seed.PermissionKeyByOperationID()
	if len(lookup) == 0 {
		t.Fatal("PermissionKeyByOperationID returned empty map")
	}
}

// TestSmokeOgenServerBuilds verifies that apigen.NewServer succeeds with a
// real APIHandler (nil db) — confirming the ogen codegen is in sync with the
// handler implementation.
func TestSmokeOgenServerBuilds(t *testing.T) {
	cfg := &config.Config{
		Env: "test",
		JWT: config.JWTConfig{Secret: "test-secret-for-smoke-tests"},
	}
	logger := zap.NewNop()
	h := handlers.NewAPIHandler(nil, cfg, logger, &noopEvaluator{}, nilAPIEndpointSvc{})
	_, err := apigen.NewServer(h, handlers.SecurityHandler{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
}

// TestSmokeRegisterEmptyBodyReturns400 verifies that /auth/register also
// validates empty fields and returns 400.
func TestSmokeRegisterEmptyBodyReturns400(t *testing.T) {
	r := testEngine(t)
	body := []byte(`{"username":"","password":""}`)
	w := doRequest(t, r, http.MethodPost, "/api/v1/auth/register", body, nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestSmokeAuthMeWithoutTokenReturns401 verifies /auth/me requires a token.
func TestSmokeAuthMeWithoutTokenReturns401(t *testing.T) {
	r := testEngine(t)
	w := doRequest(t, r, http.MethodGet, "/api/v1/auth/me", nil, nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d — body: %s", w.Code, w.Body.String())
	}
}
