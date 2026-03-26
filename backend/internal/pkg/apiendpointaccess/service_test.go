package apiendpointaccess

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequireActiveEndpointBlocksSuspendedBusinessRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &service{
		routeMap: map[string]EndpointMeta{
			routeKey(http.MethodGet, "/api/v1/orders"): {
				ID:     uuid.New(),
				Method: http.MethodGet,
				Path:   "/api/v1/orders",
				Status: "suspended",
			},
		},
	}

	router := gin.New()
	router.Use(svc.RequireActiveEndpoint())
	router.GET("/api/v1/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, resp.Code)
	}
}

func TestRequireActiveEndpointBypassesApiRegistryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &service{
		routeMap: map[string]EndpointMeta{
			routeKey(http.MethodGet, "/api/v1/api-endpoints"): {
				ID:     uuid.New(),
				Method: http.MethodGet,
				Path:   "/api/v1/api-endpoints",
				Status: "suspended",
			},
			routeKey(http.MethodPut, "/api/v1/api-endpoints/:id"): {
				ID:     uuid.New(),
				Method: http.MethodPut,
				Path:   "/api/v1/api-endpoints/:id",
				Status: "suspended",
			},
		},
	}

	router := gin.New()
	router.Use(svc.RequireActiveEndpoint())
	router.GET("/api/v1/api-endpoints", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.PUT("/api/v1/api-endpoints/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/api-endpoints", nil)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, listResp.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/api-endpoints/"+uuid.NewString(), nil)
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d", http.StatusOK, updateResp.Code)
	}
}
