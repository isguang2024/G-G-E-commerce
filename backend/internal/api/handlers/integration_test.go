//go:build integration

// Package handlers_test — integration tests for the ogen bridge layer.
//
// These tests connect to a real Postgres instance and exercise full HTTP
// request/response cycles through SetupRouter. They are gated behind the
// "integration" build tag and must be run with:
//
//	go test -tags integration ./internal/api/handlers/...
//
// Environment variables (all optional, fall back to defaults):
//
//	TEST_DB_HOST     — default "localhost"
//	TEST_DB_PORT     — default "5432"
//	TEST_DB_USER     — default "postgres"
//	TEST_DB_PASSWORD — default "postgres"
//	TEST_DB_NAME     — default "gg_ecommerce"
package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/gg-ecommerce/backend/internal/api/router"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
)

// ── test fixtures ─────────────────────────────────────────────────────────────

const (
	integTestUsername = "integration_test_user"
	integTestPassword = "Integration@Test123"
	integTestEmail    = "integration_test@gg-test.local"
	integTestNickname = "Integration Test User"
)

var (
	integDB     *gorm.DB
	integRouter http.Handler
	integUserID uuid.UUID
	integToken  string // filled by TestIntegrationLogin
)

// ── TestMain ──────────────────────────────────────────────────────────────────

func TestMain(m *testing.M) {
	db, err := openIntegDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: cannot connect to DB: %v\n", err)
		os.Exit(1)
	}
	integDB = db

	// Create a known test user; delete-on-exit.
	userID, err := ensureTestUser(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: ensureTestUser: %v\n", err)
		os.Exit(1)
	}
	integUserID = userID

	// Build the full production-equivalent router.
	cfg := integConfig()
	logger := zap.NewNop()
	integRouter = router.SetupRouter(cfg, logger, db)

	code := m.Run()

	// Cleanup: delete the test user we created.
	db.Unscoped().Where("id = ?", userID).Delete(&models.User{})

	os.Exit(code)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func openIntegDB() (*gorm.DB, error) {
	host := envOr("TEST_DB_HOST", "localhost")
	port := envOr("TEST_DB_PORT", "5432")
	user := envOr("TEST_DB_USER", "postgres")
	pass := envOr("TEST_DB_PASSWORD", "postgres")
	name := envOr("TEST_DB_NAME", "gg_ecommerce")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, user, pass, name, port,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		NowFunc: func() time.Time { return time.Now().Local() },
	})
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func integConfig() *config.Config {
	return &config.Config{
		Env: "test",
		JWT: config.JWTConfig{
			Secret:        "integration-test-secret",
			AccessExpire:  3600,
			RefreshExpire: 86400,
		},
		DB: config.DBConfig{
			Host:     envOr("TEST_DB_HOST", "localhost"),
			Port:     5432,
			User:     envOr("TEST_DB_USER", "postgres"),
			Password: envOr("TEST_DB_PASSWORD", "postgres"),
			DBName:   envOr("TEST_DB_NAME", "gg_ecommerce"),
			SSLMode:  "disable",
			TimeZone: "Asia/Shanghai",
		},
	}
}

// ensureTestUser creates (or finds) the integration test user and returns its ID.
func ensureTestUser(db *gorm.DB) (uuid.UUID, error) {
	var existing models.User
	err := db.Where("username = ?", integTestUsername).First(&existing).Error
	if err == nil {
		return existing.ID, nil // already exists
	}
	if err != gorm.ErrRecordNotFound {
		return uuid.UUID{}, fmt.Errorf("lookup test user: %w", err)
	}

	hash, err := password.Hash(integTestPassword)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("hash password: %w", err)
	}

	u := models.User{
		Username:     integTestUsername,
		Email:        integTestEmail,
		Nickname:     integTestNickname,
		PasswordHash: hash,
		Status:       "active",
		IsSuperAdmin: true, // bypass permission evaluator in integration tests
	}
	if err := db.Create(&u).Error; err != nil {
		return uuid.UUID{}, fmt.Errorf("create test user: %w", err)
	}
	return u.ID, nil
}

func integDo(method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	var req *http.Request
	if len(body) > 0 {
		req, _ = http.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequestWithContext(context.Background(), method, path, nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	integRouter.ServeHTTP(w, req)
	return w
}

func integBearerHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// ── tests ─────────────────────────────────────────────────────────────────────

// TestIntegrationLogin verifies that a known user can log in and receive a JWT.
// This is the foundation test — it sets integToken for subsequent tests.
func TestIntegrationLogin(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"username": integTestUsername,
		"password": integTestPassword,
	})
	w := integDo(http.MethodPost, "/api/v1/auth/login", body, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("login: invalid JSON: %v — body: %s", err, w.Body.String())
	}

	token, _ := resp["access_token"].(string)
	if token == "" {
		// tolerate nested: resp["data"]["access_token"]
		if data, ok := resp["data"].(map[string]interface{}); ok {
			token, _ = data["access_token"].(string)
		}
	}
	if token == "" {
		t.Fatalf("login: no access_token in response: %s", w.Body.String())
	}

	integToken = token
	t.Logf("login: got token (len=%d)", len(token))
}

// TestIntegrationLoginWrongPassword verifies that a bad password returns 401/400.
func TestIntegrationLoginWrongPassword(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"username": integTestUsername,
		"password": "wrong-password",
	})
	w := integDo(http.MethodPost, "/api/v1/auth/login", body, nil)
	if w.Code == http.StatusOK {
		t.Errorf("expected non-200 for wrong password, got %d", w.Code)
	}
}

// TestIntegrationLoginUnknownUser verifies that an unknown user returns non-200.
func TestIntegrationLoginUnknownUser(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"username": "does_not_exist_xyz_9999",
		"password": "whatever",
	})
	w := integDo(http.MethodPost, "/api/v1/auth/login", body, nil)
	if w.Code == http.StatusOK {
		t.Errorf("expected non-200 for unknown user, got %d", w.Code)
	}
}

// TestIntegrationAuthMe verifies that /auth/me works with a valid JWT.
func TestIntegrationAuthMe(t *testing.T) {
	if integToken == "" {
		t.Skip("integToken not set — TestIntegrationLogin must run first")
	}
	w := integDo(http.MethodGet, "/api/v1/auth/me", nil, integBearerHeader(integToken))
	if w.Code != http.StatusOK {
		t.Fatalf("auth/me: expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("auth/me: invalid JSON: %v", err)
	}
}

// TestIntegrationAuthMeNoToken verifies that /auth/me without JWT returns 401.
func TestIntegrationAuthMeNoToken(t *testing.T) {
	w := integDo(http.MethodGet, "/api/v1/auth/me", nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("auth/me no token: expected 401, got %d", w.Code)
	}
}

// TestIntegrationRolesList verifies that /roles returns 200 with a valid token.
// The response array may be empty if no roles exist in the test DB.
func TestIntegrationRolesList(t *testing.T) {
	if integToken == "" {
		t.Skip("integToken not set — TestIntegrationLogin must run first")
	}
	w := integDo(http.MethodGet, "/api/v1/roles", nil, integBearerHeader(integToken))
	if w.Code != http.StatusOK {
		t.Fatalf("roles: expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestIntegrationRolesListNoToken verifies that /roles requires authentication.
func TestIntegrationRolesListNoToken(t *testing.T) {
	w := integDo(http.MethodGet, "/api/v1/roles", nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("roles no token: expected 401, got %d", w.Code)
	}
}

// TestIntegrationNavigation verifies that GET /runtime/navigation returns 200.
func TestIntegrationNavigation(t *testing.T) {
	if integToken == "" {
		t.Skip("integToken not set — TestIntegrationLogin must run first")
	}
	w := integDo(http.MethodGet, "/api/v1/runtime/navigation", nil, integBearerHeader(integToken))
	if w.Code != http.StatusOK {
		t.Fatalf("navigation: expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestIntegrationFeaturePackagesList verifies that GET /feature-packages returns 200.
func TestIntegrationFeaturePackagesList(t *testing.T) {
	if integToken == "" {
		t.Skip("integToken not set — TestIntegrationLogin must run first")
	}
	w := integDo(http.MethodGet, "/api/v1/feature-packages", nil, integBearerHeader(integToken))
	if w.Code != http.StatusOK {
		t.Fatalf("feature-packages: expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestIntegrationCollaborationWorkspacesList verifies GET /collaboration-workspaces returns 200.
func TestIntegrationCollaborationWorkspacesList(t *testing.T) {
	if integToken == "" {
		t.Skip("integToken not set — TestIntegrationLogin must run first")
	}
	w := integDo(http.MethodGet, "/api/v1/collaboration-workspaces", nil, integBearerHeader(integToken))
	if w.Code != http.StatusOK {
		t.Fatalf("collaboration-workspaces: expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}
