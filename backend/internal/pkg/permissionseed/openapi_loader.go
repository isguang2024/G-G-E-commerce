// openapi_loader.go reads the JSON produced by cmd/gen-permissions and
// exposes it to the runtime. Phase 1 cleanup: only validate that the file
// exists, has at least one operation, and every entry carries a non-empty
// permission_key. Phase 3+ wires this into the casbin evaluator and the
// startup-time DB consistency check (each permission_key must exist in
// the permission_keys table).
package permissionseed

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed openapi_seed.json
var openapiSeedRaw []byte

type OpenAPIOperation struct {
	OperationID   string `json:"operation_id"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	PermissionKey string `json:"permission_key"`
	Summary       string `json:"summary,omitempty"`
	Description   string `json:"description,omitempty"`
	Tags          string `json:"tags,omitempty"`
	APICategory   string `json:"api_category,omitempty"`
	AccessMode    string `json:"access_mode"`
}

type OpenAPISeed struct {
	Source     string             `json:"source"`
	Operations []OpenAPIOperation `json:"operations"`
}

// LoadOpenAPISeed parses the embedded openapi_seed.json. Returns an error
// if the JSON is malformed or any operation is missing a permission_key.
func LoadOpenAPISeed() (*OpenAPISeed, error) {
	var seed OpenAPISeed
	if err := json.Unmarshal(openapiSeedRaw, &seed); err != nil {
		return nil, fmt.Errorf("decode openapi_seed.json: %w", err)
	}
	for i, op := range seed.Operations {
		if op.PermissionKey == "" && op.AccessMode != "public" && op.AccessMode != "authenticated" {
			return nil, fmt.Errorf("openapi_seed.json[%d] %s %s missing permission_key", i, op.Method, op.Path)
		}
	}
	return &seed, nil
}

// PermissionKeyByOperationID flattens the seed into an operationID -> permission key map.
//
// Deprecated: this lookup is kept only as a fallback for the middleware when
// operation_id can still resolve the key. Prefer PermissionLookup whose
// main path is keyed by endpoint_code (derived from METHOD + path via
// StableID) so renaming an operation_id no longer invalidates runtime
// permission checks nor api_endpoint_permission_bindings alignment.
func (s *OpenAPISeed) PermissionKeyByOperationID() map[string]string {
	out := make(map[string]string, len(s.Operations))
	for _, op := range s.Operations {
		out[op.OperationID] = op.PermissionKey
	}
	return out
}

// PermissionLookup is the runtime view consumed by the auto permission
// middleware. The middleware resolves the permission key via two stages:
//
//  1. primary — operationID → endpoint_code (StableID("openapi-api-endpoint",
//     METHOD+" "+path)) → permission_key. The endpoint_code dimension aligns
//     with the api_endpoints.code / api_endpoint_permission_bindings.endpoint_code
//     columns, so renaming operationId while keeping METHOD+path does not
//     affect permission enforcement.
//  2. fallback — operationID → permission_key (kept for backward compatibility
//     with legacy callers). Marked deprecated and must not appear on the
//     main resolution path in production.
type PermissionLookup struct {
	// ByEndpointCode maps endpoint_code (StableID-derived) -> permission_key.
	// This is the authoritative lookup that aligns with the DB bindings table.
	ByEndpointCode map[string]string
	// OperationToEndpointCode maps ogen operation_id -> endpoint_code. Needed
	// because ogen's middleware.Request only carries operation_id; we translate
	// to endpoint_code before hitting ByEndpointCode.
	OperationToEndpointCode map[string]string
	// ByOperationIDDeprecated is the legacy operationID -> permission_key map.
	// Do NOT rely on it in new code. Present only so the middleware can emit a
	// deprecation warning if the main path misses but the fallback still hits.
	ByOperationIDDeprecated map[string]string
}

// PermissionLookup builds the two-stage lookup described on PermissionLookup.
// Called once at startup from router.SetupRouter.
func (s *OpenAPISeed) PermissionLookup() *PermissionLookup {
	byCode := make(map[string]string, len(s.Operations))
	opToCode := make(map[string]string, len(s.Operations))
	byOp := make(map[string]string, len(s.Operations))
	for _, op := range s.Operations {
		method := strings.ToUpper(strings.TrimSpace(op.Method))
		path := strings.TrimSpace(op.Path)
		endpointCode := StableID("openapi-api-endpoint", method+" "+path).String()
		if op.PermissionKey != "" {
			byCode[endpointCode] = op.PermissionKey
			byOp[op.OperationID] = op.PermissionKey
		}
		opToCode[op.OperationID] = endpointCode
	}
	return &PermissionLookup{
		ByEndpointCode:          byCode,
		OperationToEndpointCode: opToCode,
		ByOperationIDDeprecated: byOp,
	}
}

// AccessModeByMethodPath returns a map of "METHOD /path" -> access_mode.
// This is the single source of truth for endpoint auth behaviour — derived
// directly from the embedded openapi_seed.json (which mirrors x-access-mode
// in openapi.yaml). Use it instead of any hardcoded path-pattern lists.
//
// Values: "permission" | "authenticated" | "public"
func (s *OpenAPISeed) AccessModeByMethodPath() map[string]string {
	out := make(map[string]string, len(s.Operations))
	for _, op := range s.Operations {
		key := strings.ToUpper(strings.TrimSpace(op.Method)) + " " + strings.TrimSpace(op.Path)
		out[key] = op.AccessMode
	}
	return out
}
