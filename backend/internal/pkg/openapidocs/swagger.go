// Package openapidocs serves the embedded OpenAPI spec and a minimal
// Swagger UI page. The UI is intentionally a one-file CDN-backed HTML
// shim — adding a vendored asset bundle is overkill while the spec is
// still small. Mounted in router.SetupRouter at /swagger and /openapi.yaml.
package openapidocs

import (
	"net/http"

	"github.com/gin-gonic/gin"

	openapispec "github.com/gg-ecommerce/backend/api/openapi"
)

const swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>GG E-commerce API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout",
      });
    };
  </script>
</body>
</html>`

// Mount registers /openapi.yaml and /swagger on the supplied engine.
func Mount(r *gin.Engine) {
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapispec.SpecBytes)
	})
	r.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})
}
