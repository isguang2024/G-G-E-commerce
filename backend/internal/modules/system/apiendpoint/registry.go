package apiendpoint

// registry.go now holds only the two helpers still required after the
// seed-driven route registration refactor:
//
//   - SyncSummary: shape of the /api-endpoints/sync response body, consumed
//     by the admin UI to render post-action feedback.
//   - deriveStableEndpointCode: deterministic UUIDv5 generator used by
//     service.Save() to backfill codes on freshly-created rows.
//
// The previous contents (Registrar / MetaBuilder / RouteMeta / RuntimeRoute /
// Annotate / Lookup / SyncRoutes / CollectRuntimeRoutes / fixedManagedRouteCodes
// / ensureManagedEndpointColumns / IsFixedManagedRouteCode) were dead code
// after /api/v1 routes moved to the OpenAPI seed pipeline and have been
// removed. See docs/API_OPENAPI_FIXED_FLOW.md for the canonical flow.

import (
	"crypto/sha1"
	"strings"

	"github.com/google/uuid"
)

// SyncSummary describes the aggregate effect of a single Sync() run so
// clients can render post-action feedback.
type SyncSummary struct {
	Processed int
	Created   int
	Updated   int
	Total     int
}

// deriveStableEndpointCode produces a deterministic UUIDv5 code for a
// method+path pair. The seed-driven ensure pipeline uses an identical
// algorithm in permissionseed.StableID so manual save + seed ensure agree
// on the same code for the same route.
func deriveStableEndpointCode(method, path string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
	return uuid.NewHash(sha1.New(), uuid.NameSpaceURL, []byte("api-endpoint:"+normalized), 5).String()
}
