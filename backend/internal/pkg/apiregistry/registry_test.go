package apiregistry

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

func TestResolveRouteCode(t *testing.T) {
	explicitCode := "2c54f06b-bf0d-4f8e-b68c-f41078c1f5e0"
	got := ResolveRouteCode(http.MethodGet, "/api/v1/menus/tree", &RouteMeta{Code: explicitCode})
	if got != explicitCode {
		t.Fatalf("ResolveRouteCode() = %q, want %q", got, explicitCode)
	}

	fixedCode := ResolveRouteCode(http.MethodGet, "/api/v1/menus/tree", nil)
	want := "75d2c43b-9b1d-51a4-9f19-6424790199bc"
	if fixedCode != want {
		t.Fatalf("ResolveRouteCode() fixed = %q, want %q", fixedCode, want)
	}

	if got := ResolveRouteCode(http.MethodGet, "/api/v1/unknown-fixed-code", nil); got != "" {
		t.Fatalf("ResolveRouteCode() unknown route = %q, want empty", got)
	}
}

func TestRegistrarAssignsStableCodeWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	reg := NewRegistrar(router.Group("/api/v1/menus"), "menu")
	reg.GET("/tree", reg.Meta("获取菜单树").Build(), func(c *gin.Context) {})

	meta, ok := Lookup(http.MethodGet, "/api/v1/menus/tree")
	if !ok {
		t.Fatalf("expected route meta to be annotated")
	}

	want := "75d2c43b-9b1d-51a4-9f19-6424790199bc"
	if meta.Code != want {
		t.Fatalf("annotated meta code = %q, want %q", meta.Code, want)
	}
	if meta.CategoryCode != "menu" {
		t.Fatalf("annotated meta category = %q, want %q", meta.CategoryCode, "menu")
	}
}

func TestRegistrarActionRoutesInheritCategoryHint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	reg := NewRegistrar(router.Group("/api/v1/tenants"), "tenant")
	reg.GETAction("", "获取团队列表", "tenant.manage", nil, func(c *gin.Context) {})

	meta, ok := Lookup(http.MethodGet, "/api/v1/tenants")
	if !ok {
		t.Fatalf("expected route meta to be annotated")
	}
	if meta.CategoryCode != "tenant" {
		t.Fatalf("annotated meta category = %q, want %q", meta.CategoryCode, "tenant")
	}
}

func TestMetaWithPermissionAllowsUncategorizedByDefault(t *testing.T) {
	meta := MetaWithPermission("获取团队列表", "tenant.manage")
	if meta == nil {
		t.Fatalf("expected meta")
	}
	if meta.CategoryCode != "" {
		t.Fatalf("meta category = %q, want empty", meta.CategoryCode)
	}
}

func TestRegistrarRawMetaInheritsCategoryHint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	reg := NewRegistrar(router.Group("/api/v1/tenants"), "tenant")
	reg.GET("/my-teams", &RouteMeta{Summary: "获取我的团队列表"}, func(c *gin.Context) {})

	meta, ok := Lookup(http.MethodGet, "/api/v1/tenants/my-teams")
	if !ok {
		t.Fatalf("expected route meta to be annotated")
	}
	if meta.CategoryCode != "tenant" {
		t.Fatalf("annotated meta category = %q, want %q", meta.CategoryCode, "tenant")
	}
}

func TestFixedManagedRouteCodesAreUniqueUUIDs(t *testing.T) {
	seen := make(map[string]string, len(fixedManagedRouteCodes))
	for route, code := range fixedManagedRouteCodes {
		if _, err := uuid.Parse(code); err != nil {
			t.Fatalf("route %s has invalid uuid %q: %v", route, code, err)
		}
		if existing, ok := seen[code]; ok {
			t.Fatalf("duplicate fixed route code %q for %s and %s", code, existing, route)
		}
		seen[code] = route
	}
}

func TestShouldInsertManagedEndpoint(t *testing.T) {
	if !shouldInsertManagedEndpoint(nil) {
		t.Fatalf("expected nil endpoint to be insertable")
	}
	if shouldInsertManagedEndpoint(&models.APIEndpoint{}) {
		t.Fatalf("expected existing endpoint to skip insert")
	}
}

func TestShouldBackfillManagedEndpointCode(t *testing.T) {
	if shouldBackfillManagedEndpointCode(nil) {
		t.Fatalf("nil endpoint should not backfill code")
	}
	if !shouldBackfillManagedEndpointCode(&models.APIEndpoint{}) {
		t.Fatalf("blank code endpoint should backfill code")
	}
	if shouldBackfillManagedEndpointCode(&models.APIEndpoint{Code: "stable-guid"}) {
		t.Fatalf("existing fixed code should not backfill")
	}
}

func TestIsFixedManagedRouteCode(t *testing.T) {
	code := fixedManagedRouteCodes[routeKey(http.MethodGet, "/api/v1/api-endpoints")]
	if !IsFixedManagedRouteCode(code) {
		t.Fatalf("expected %q to be recognized as fixed managed route code", code)
	}
	if IsFixedManagedRouteCode(uuid.NewString()) {
		t.Fatalf("unexpected random uuid recognized as fixed managed route code")
	}
}
