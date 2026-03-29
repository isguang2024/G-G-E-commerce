package apiendpoint

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
)

func TestResolveEndpointCodeForSavePreservesExistingCode(t *testing.T) {
	current := &user.APIEndpoint{
		ID:     uuid.New(),
		Code:   uuid.NewString(),
		Method: http.MethodGet,
		Path:   "/api/v1/menus/tree",
	}
	endpoint := &user.APIEndpoint{
		ID:     current.ID,
		Method: http.MethodGet,
		Path:   "/api/v1/menus/runtime",
	}

	got := resolveEndpointCodeForSave(endpoint, current)
	if got != current.Code {
		t.Fatalf("resolveEndpointCodeForSave() = %q, want %q", got, current.Code)
	}
}

func TestResolveEndpointCodeForSaveDerivesCodeOnCreate(t *testing.T) {
	endpoint := &user.APIEndpoint{
		Method: http.MethodPost,
		Path:   "/api/v1/api-endpoints/sync",
	}

	got := resolveEndpointCodeForSave(endpoint, nil)
	want := deriveStableEndpointCode(endpoint.Method, endpoint.Path)
	if got != want {
		t.Fatalf("resolveEndpointCodeForSave() = %q, want %q", got, want)
	}
}

func TestListRuntimeStates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	reg := apiregistry.NewRegistrar(router.Group("/api/v1/api-endpoints"), "api_endpoint")
	reg.GET("", reg.Meta("获取 API 注册表").Build(), func(c *gin.Context) {})
	router.GET("/api/v1/runtime-legacy", func(c *gin.Context) {})

	svc := &service{router: router}
	staleFixedCode := apiregistry.ResolveRouteCode(http.MethodPost, "/api/v1/api-endpoints/cleanup-stale", nil)

	endpoints := []user.APIEndpoint{
		{
			ID:     uuid.New(),
			Code:   apiregistry.ResolveRouteCode(http.MethodGet, "/api/v1/api-endpoints", nil),
			Method: http.MethodGet,
			Path:   "/api/v1/changed-by-admin",
			Source: "sync",
		},
		{
			ID:     uuid.New(),
			Code:   staleFixedCode,
			Method: http.MethodPost,
			Path:   "/api/v1/api-endpoints/cleanup-stale",
			Source: "sync",
		},
		{
			ID:     uuid.New(),
			Method: http.MethodGet,
			Path:   "/api/v1/runtime-legacy",
			Source: "manual",
		},
		{
			ID:     uuid.New(),
			Code:   uuid.NewString(),
			Method: http.MethodGet,
			Path:   "/api/v1/manual-missing",
			Source: "manual",
		},
	}

	states := svc.ListRuntimeStates(endpoints)
	if !states[endpoints[0].ID].RuntimeExists || states[endpoints[0].ID].Stale {
		t.Fatalf("expected fixed-code endpoint to be matched by runtime code, got %#v", states[endpoints[0].ID])
	}
	if !states[endpoints[1].ID].Stale || states[endpoints[1].ID].RuntimeExists {
		t.Fatalf("expected missing sync endpoint to be marked stale, got %#v", states[endpoints[1].ID])
	}
	if !states[endpoints[2].ID].RuntimeExists || states[endpoints[2].ID].Stale {
		t.Fatalf("expected legacy method+path endpoint to be matched, got %#v", states[endpoints[2].ID])
	}
	if states[endpoints[3].ID].RuntimeExists || states[endpoints[3].ID].Stale {
		t.Fatalf("expected manual missing endpoint to stay non-stale, got %#v", states[endpoints[3].ID])
	}
}
