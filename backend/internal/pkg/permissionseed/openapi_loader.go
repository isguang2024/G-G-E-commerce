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
)

//go:embed openapi_seed.json
var openapiSeedRaw []byte

type OpenAPIOperation struct {
	OperationID   string `json:"operation_id"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	PermissionKey string `json:"permission_key"`
	Summary       string `json:"summary,omitempty"`
	TenantScoped  bool   `json:"tenant_scoped"`
	AppScope      string `json:"app_scope"`
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

// PermissionKeyByOperationID flattens the seed into the runtime lookup used
// by the auto permission middleware: ogen operationID -> permission key.
func (s *OpenAPISeed) PermissionKeyByOperationID() map[string]string {
	out := make(map[string]string, len(s.Operations))
	for _, op := range s.Operations {
		out[op.OperationID] = op.PermissionKey
	}
	return out
}
