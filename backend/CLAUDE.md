# backend/CLAUDE.md

Backend-specific guidance for the GGE 5.0 Go service. Read root `CLAUDE.md`
and `docs/V5_REFACTOR_TASKS.md` first for the overall architecture.

## How to add a v5 endpoint

The OpenAPI spec is the single source of truth. Adding an endpoint is a
mechanical 7-step process. Do **not** shortcut it by editing legacy gin
modules.

### 1. Declare the operation in the spec

Edit `api/openapi/openapi.yaml`. Every operation MUST carry the four
`x-` extensions:

```yaml
/widgets/{id}:
  get:
    operationId: getWidget
    tags: [widget]
    x-permission-key: widget.read     # required unless x-access-mode = public/authenticated
    x-tenant-scoped: true             # always true for v5
    x-app-scope: optional             # required | optional | none
    x-access-mode: permission         # permission | authenticated | public
    parameters: [...]
    responses: { '200': { ... } }
```

### 2. Regenerate everything

```
make api          # bundle → lint → ogen → permission seed → frontend types
```

Or step by step:
```
make api-bundle   # merges openapi.root.yaml + domains/* → dist/openapi.yaml
make api-lint     # redocly lint (must pass before gen)
make api-gen      # ogen: dist/openapi.yaml → api/gen/
make api-perms    # permission seed + frontend error-codes.ts
```

Commit the regenerated `api/gen/` and `internal/pkg/permissionseed/openapi_seed.json` files.

**Spec source layout** (edit here, not in `dist/`):
```
api/openapi/
├── openapi.root.yaml          ← entry point (info/servers/security/tags + $ref)
├── components/
│   ├── errors.yaml            ← Error schema + shared responses
│   ├── common.yaml            ← all shared schemas
│   └── security.yaml          ← securitySchemes
└── domains/{tag}/
    ├── paths.yaml             ← paths for this domain
    └── schemas.yaml           ← domain-specific schemas (move from common.yaml incrementally)
```

Operational rule:

- If you changed `api/openapi/openapi.yaml`, rerun `update-openapi.bat`
  so the embedded OpenAPI seed stays in sync.
- If the DB already has the latest schema and default rows, a seed refresh
  is usually enough; you do not need to rerun migrations just because the
  OpenAPI spec changed.
- If you introduced new tables, columns, baseline permission keys, or any
  other schema/default-data change, rerun `cmd/migrate` as well.
- For a brand-new database, run `cmd/migrate` first, then `update-openapi.bat`.

### 4. Implement the operation method

Add a method on `APIHandler` in `internal/api/handlers/{domain}.go` that
matches the generated `gen.Handler` signature. Reach into the existing
service-layer (`internal/modules/system/{domain}`) for business logic.

Do NOT re-introduce a legacy Gin module shell
(`internal/modules/system/*/module.go`). Those files have been deleted;
the ogen bridge in `internal/api/router/router.go` already covers every
operation declared in the spec.

### 5. Ensure the permission key exists in the DB

If you introduced a new `x-permission-key`, the runtime upsert in
`internal/pkg/permissionseed.EnsureOpenAPIPermissionKeys` will materialise
it on next `cmd/migrate` run. For frequently-shipped, well-known baseline
keys you may also add them to the goose migration
`internal/pkg/database/migrations/00001_permission_seed_baseline.sql`.

### 6. Add a smoke test

Append a test to `internal/api/handlers/integration_test.go`
(`//go:build integration`). Reuse `integDo`, `integToken`, and the existing
fixtures. Run with:

```
go test -tags integration ./internal/api/handlers/...
```

### 7. Update the frontend client

Frontend consumes the generated TS client under `frontend/src/api/v5/`.
Re-run the generator there and replace any hand-written axios calls.

## Anti-patterns

- Do NOT re-introduce a legacy Gin module shell
  (`internal/modules/system/*/module.go`) — those files have been deleted
  and must not be recreated.
- Do NOT bypass `internal/pkg/permission/evaluator` for permission
  decisions; never read `feature_package_keys` / `role_feature_packages`
  directly from a handler.
- Do NOT import `internal/pkg/authorization` — the package is empty and
  scheduled for physical deletion.
- Do NOT write tenant-unaware queries; every query must filter on
  `tenant_id` (currently always `default`).
- Do NOT hand-edit `openapi_seed.json` — it is regenerated from the spec.
