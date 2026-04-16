# backend/CLAUDE.md

Backend-specific guidance for the GGE Go service. Read `AGENTS.md`,
`docs/project-framework.md`, `docs/frontend-guideline.md` first for the shared rules.
If this file conflicts with `AGENTS.md`, follow `AGENTS.md`.

## Documentation Navigation

- Execution-oriented feature/flow documents are under `.claude/Instructions/`.
- Backend API contract truth source remains `backend/api/openapi/README.md`.
- Development guides are in `docs/guides/`.

## How to add an endpoint

The OpenAPI spec is the single source of truth. Adding an endpoint is a
mechanical process. Do **not** shortcut it by editing generated code or
legacy gin modules.

### 1. Declare the operation in the spec

Edit the source files under `api/openapi/`:

- `api/openapi/openapi.root.yaml`
- `api/openapi/components/*.yaml`
- `api/openapi/domains/{tag}/paths.yaml`
- `api/openapi/domains/{tag}/schemas.yaml`

Every operation MUST carry the four `x-` extensions:

```yaml
/widgets/{id}:
  get:
    operationId: getWidget
    tags: [widget]
    x-permission-key: widget.read     # required unless x-access-mode = public/authenticated
    x-tenant-scoped: true             # always true
    x-app-scope: optional             # required | optional | none
    x-access-mode: permission         # permission | authenticated | public
    parameters: [...]
    responses: { '200': { ... } }
```

### 2. Regenerate everything

Preferred refresh commands:

- `make api` for the backend chain: bundle → lint → ogen → permission seed
- `make api-front` or `cd ../frontend && pnpm run gen:api` for the frontend
  schema
- `update-openapi.bat` for the Windows wrapper around the backend chain

Commit the regenerated `api/gen/` and
`internal/pkg/permissionseed/openapi_seed.json` files when they change.

Operational rule:

- If you changed any file under `api/openapi/`, rerun the backend OpenAPI
  chain so the generated spec stays in sync.
- If the DB already has the latest schema and default rows, a seed refresh
  is usually enough; you do not need to rerun migrations just because the
  OpenAPI spec changed.
- If you introduced new tables, columns, baseline permission keys, or any
  other schema/default-data change, rerun `cmd/migrate` as well.
- For a brand-new database, run `cmd/migrate` first, then `update-openapi.bat`.

### 3. Implement the operation method

Add a method on `APIHandler` in `internal/api/handlers/{domain}.go` that
matches the generated `gen.Handler` signature. Reach into the existing
service-layer (`internal/modules/system/{domain}`) for business logic.

Do NOT re-introduce a legacy Gin module shell
(`internal/modules/system/*/module.go`). The ogen bridge in
`internal/api/router/router.go` covers every operation declared in the spec.

### 4. Ensure the permission key exists in the DB

If you introduced a new `x-permission-key`, the runtime upsert in
`internal/pkg/permissionseed.EnsureOpenAPIPermissionKeys` will materialise
it on next `cmd/migrate` run. For frequently-shipped, well-known baseline
keys you may also add them to the goose migration
`internal/pkg/database/migrations/00001_permission_seed_baseline.sql`.

### 5. Add a smoke test

Append a test to `internal/api/handlers/integration_test.go`
(`//go:build integration`). Reuse `integDo`, `integToken`, and the existing
fixtures. Run with:

```
go test -tags integration ./internal/api/handlers/...
```

### 6. Update the frontend client

Frontend consumes the generated TS client under `frontend/src/api/v5/`.
Re-run the generator there and replace any hand-written axios calls.

## How to add a complete system module

If you're adding a new **management page** (not just a single endpoint),
the full checklist is in `docs/guides/new-module-checklist.md`. It covers:

1. Migration → Model → OpenAPI spec → permission key mapping →
   permission key seed → module group → menu seed → feature package
   binding → code gen → handler/service → frontend page → verify.

Key invariant: `permissionkey.go` mapping's `ResourceCode` must equal
the `ModuleGroup.Code` in `seeds.go`, otherwise the permission key
lands in the wrong group in the admin UI.

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
