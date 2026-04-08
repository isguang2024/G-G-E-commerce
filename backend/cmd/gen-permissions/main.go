// cmd/gen-permissions parses api/openapi/openapi.yaml and derives the
// permission_keys / api_endpoints / permission_key_api_bindings seed.
//
// Output: internal/pkg/permissionseed/openapi_seed.json
//
// Each operation in the spec must declare:
//   x-permission-key: <key>            (required, the permission key string)
//   x-tenant-scoped:  <bool>           (optional, defaults to true)
//   x-access-mode:    permission|public (optional, defaults to permission)
//
// Phase 1: parser only emits JSON; the runtime startup-check that loads this
// JSON and validates against the DB lands together with Phase 2 baseline.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type operationSpec struct {
	OperationID   string   `yaml:"operationId"`
	Summary       string   `yaml:"summary"`
	Description   string   `yaml:"description"`
	Tags          []string `yaml:"tags"`
	PermissionKey string   `yaml:"x-permission-key"`
	AccessMode    string   `yaml:"x-access-mode"`
	APICategory   string   `yaml:"x-api-category"`
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
	Description   string `json:"description,omitempty"`
	Tags          string `json:"tags,omitempty"`
	APICategory   string `json:"api_category,omitempty"`
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
	specPath := filepath.Join("api", "openapi", "dist", "openapi.yaml")
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
			accessMode := accessModeRaw
			// Resolve category: explicit x-api-category wins; fall back to first tag.
			apiCategory := strings.TrimSpace(op.APICategory)
			if apiCategory == "" && len(op.Tags) > 0 {
				apiCategory = strings.TrimSpace(op.Tags[0])
			}
			entries = append(entries, seedEntry{
				OperationID:   op.OperationID,
				Method:        method,
				Path:          path,
				PermissionKey: op.PermissionKey,
				Summary:       op.Summary,
				Description:   op.Description,
				Tags:          strings.Join(op.Tags, ","),
				APICategory:   apiCategory,
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

	// ── 派生前端错误码 TS 文件 ────────────────────────────────────────────
	if err := genErrorCodes(); err != nil {
		fmt.Fprintf(os.Stderr, "gen-permissions: error-codes: %v\n", err)
		os.Exit(1)
	}
}

// genErrorCodes 解析 internal/api/apperr/codes.go，提取 Code* 常量，
// 写出 frontend/src/api/v5/error-codes.ts。
func genErrorCodes() error {
	src := filepath.Join("internal", "api", "apperr", "codes.go")
	dst := filepath.Join("..", "frontend", "src", "api", "v5", "error-codes.ts")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, src, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", src, err)
	}

	type entry struct {
		Name    string // TS key，去掉 "Code" 前缀
		Value   int
		Comment string
	}
	var entries []entry

	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if !strings.HasPrefix(name.Name, "Code") {
					continue
				}
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.INT {
					continue
				}
				val, err := strconv.Atoi(lit.Value)
				if err != nil {
					continue
				}
				comment := ""
				if vs.Comment != nil {
					comment = strings.TrimSpace(vs.Comment.Text())
				}
				entries = append(entries, entry{
					Name:    strings.TrimPrefix(name.Name, "Code"),
					Value:   val,
					Comment: comment,
				})
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
	}

	const tpl = `// AUTO-GENERATED by cmd/gen-permissions — DO NOT EDIT.
// Source of truth: backend/internal/api/apperr/codes.go
//
// Usage:
//   import { ErrorCodes } from '@/api/v5/error-codes'
//   if (error.code === ErrorCodes.InvalidCredentials) { ... }

export const ErrorCodes = {
{{- range .}}
  /** {{.Comment}} */
  {{.Name}}: {{.Value}},
{{- end}}
} as const

export type ErrorCode = (typeof ErrorCodes)[keyof typeof ErrorCodes]
`
	t, err := template.New("ts").Parse(tpl)
	if err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()
	if err := t.Execute(out, entries); err != nil {
		return err
	}
	fmt.Printf("gen-permissions: wrote %d error codes -> %s\n", len(entries), dst)
	return nil
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
