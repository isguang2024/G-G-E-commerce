// cmd/gen-permissions parses api/openapi/openapi.yaml and derives the
// permission_keys / api_endpoints / permission_key_api_bindings seed.
//
// Output: internal/pkg/permissionseed/openapi_seed.json
//
// Each operation in the spec must declare:
//   x-permission-key: <key>            (required, the permission key string)
//   x-tenant-scoped:  <bool>           (optional, defaults to true)
//   x-app-scope:      required|optional|none (optional, defaults to optional)
//   x-access-mode:    permission|public (optional, defaults to permission)
//
// Phase 1: parser only emits JSON; the runtime startup-check that loads this
// JSON and validates against the DB lands together with Phase 2 baseline.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type operationSpec struct {
	OperationID   string `yaml:"operationId"`
	Summary       string `yaml:"summary"`
	PermissionKey string `yaml:"x-permission-key"`
	TenantScoped  *bool  `yaml:"x-tenant-scoped"`
	AppScope      string `yaml:"x-app-scope"`
	AccessMode    string `yaml:"x-access-mode"`
}

type pathItem map[string]operationSpec

type openAPI struct {
	Paths map[string]pathItem `yaml:"paths"`
}

type seedEntry struct {
	OperationID   string `json:"operation_id"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	PermissionKey string `json:"permission_key"`
	Summary       string `json:"summary,omitempty"`
	TenantScoped  bool   `json:"tenant_scoped"`
	AppScope      string `json:"app_scope"`
	AccessMode    string `json:"access_mode"`
}

type seedFile struct {
	Source     string      `json:"source"`
	Operations []seedEntry `json:"operations"`
}

var httpMethods = map[string]struct{}{
	"get": {}, "post": {}, "put": {}, "patch": {}, "delete": {}, "head": {}, "options": {},
}

func main() {
	specPath := filepath.Join("api", "openapi", "openapi.yaml")
	outPath := filepath.Join("internal", "pkg", "permissionseed", "openapi_seed.json")

	raw, err := os.ReadFile(specPath)
	must(err, "read spec")

	var doc openAPI
	must(yaml.Unmarshal(raw, &doc), "parse spec")

	var entries []seedEntry
	for path, item := range doc.Paths {
		for method, op := range item {
			if _, ok := httpMethods[method]; !ok {
				continue
			}
			accessModeRaw := op.AccessMode
			if accessModeRaw == "" {
				accessModeRaw = "permission"
			}
			if op.PermissionKey == "" && accessModeRaw != "public" && accessModeRaw != "authenticated" {
				fail(fmt.Sprintf("operation %s %s is missing x-permission-key", method, path))
			}
			tenantScoped := true
			if op.TenantScoped != nil {
				tenantScoped = *op.TenantScoped
			}
			appScope := op.AppScope
			if appScope == "" {
				appScope = "optional"
			}
			accessMode := accessModeRaw
			entries = append(entries, seedEntry{
				OperationID:   op.OperationID,
				Method:        method,
				Path:          path,
				PermissionKey: op.PermissionKey,
				Summary:       op.Summary,
				TenantScoped:  tenantScoped,
				AppScope:      appScope,
				AccessMode:    accessMode,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		return entries[i].Method < entries[j].Method
	})

	must(os.MkdirAll(filepath.Dir(outPath), 0o755), "mkdir output")
	out, err := os.Create(outPath)
	must(err, "create output")
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	must(enc.Encode(seedFile{Source: specPath, Operations: entries}), "encode")

	fmt.Printf("gen-permissions: wrote %d operations -> %s\n", len(entries), outPath)
}

func must(err error, what string) {
	if err != nil {
		fail(fmt.Sprintf("%s: %v", what, err))
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "gen-permissions:", msg)
	os.Exit(1)
}
