# backend/CLAUDE.md

Backend-specific guidance for the MaBen Go service. Read `AGENTS.md`,
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

### 3. Implement the operation (sub-handler / service)

Handlers are **domain-split**. Do not add new ops to a monolithic
`APIHandler` — find the matching sub-handler and add the method there.

#### Layout

- `internal/api/handlers/{domain}_handler.go` — `*{domain}APIHandler`
  struct + `new{Domain}APIHandler(...)` constructor (holds only the
  dependencies this domain needs)
- `internal/api/handlers/{domain}.go` — every op method with receiver
  `*{domain}APIHandler`, matching the generated `gen.Handler` signature
- `internal/api/handlers/workspace.go` — main `APIHandler` struct that
  anonymously embeds every sub-handler plus `NewAPIHandler`

Business logic stays in `internal/modules/system/{domain}`; the
sub-handler is a thin adapter.

#### Adding a new domain

1. Define `{domain}APIHandler` struct + `new{Domain}APIHandler(...)`
   in `{domain}_handler.go`
2. Move op methods to `{domain}.go` with receiver `*{domain}APIHandler`
3. Add `*{domain}APIHandler` as an anonymous embed in `APIHandler`
   inside `workspace.go`
4. Wire the constructor in `NewAPIHandler`
5. `go build ./internal/api/handlers/...` — must compile without
   ambiguous-selector errors

#### Embedding depth rule (why ambiguous selectors happen)

- `handlerBase` (wraps `gen.UnimplementedHandler`) sits at depth 2
- Every `*{domain}APIHandler` sits at depth 1
- Go picks the shallowest — sub-handler ops automatically override the
  Unimplemented stubs
- If two sub-handlers declare the same method name, Go cannot resolve
  it: **the domain boundary is wrong; split or merge domains until
  every op belongs to exactly one sub-handler**

#### Testing sub-handlers

Tests that instantiate `&APIHandler{...}` directly must set the
embedded sub-handler field, not the legacy flat layout:

```go
h := &APIHandler{
    logPolicyAPIHandler: &logPolicyAPIHandler{policyRepo: repo, ...},
}
```

Smoke tests go under `internal/api/handlers/integration_test.go`
(`//go:build integration`), reusing `integDo` / `integToken` / existing
fixtures. Run with:

```
go test -tags integration ./internal/api/handlers/...
```

#### Router — do not touch

You do **not** need to edit `internal/api/router/router.go`. The router
iterates over `permissionseed.LoadOpenAPISeed().Operations` at startup
and mounts each op to the correct Gin group (`public` → v1 without JWT;
`authenticated` / `permission` → v1 with JWT + permission middleware).
Only **non-OpenAPI** entries remain registered by hand: `/health`,
`/uploads`, OAuth callbacks, WebSocket, SSE.

Do NOT re-introduce a legacy Gin module shell
(`internal/modules/system/*/module.go`).
Do NOT pile new ops onto a god `APIHandler` — split by domain.

### 4. Ensure the permission key exists in the DB

If you introduced a new `x-permission-key`, the runtime upsert in
`internal/pkg/permissionseed.EnsureOpenAPIPermissionKeys` will materialise
it on next `cmd/migrate` run. For frequently-shipped, well-known baseline
keys you may also add them to the goose migration
`internal/pkg/database/migrations/00001_permission_seed_baseline.sql`.

**Invariant**: `permissionkey.go` mapping's `ResourceCode` must equal
the `ModuleGroup.Code` in `seeds.go`, otherwise the permission key
lands in the wrong group in the admin UI.

### 5. Verify

- `go build ./...`
- `go test ./internal/api/handlers -count=1`
- `go test ./internal/api/router -count=1` (whenever op set changes)
- `go test -tags integration ./internal/api/handlers/...` (smoke)
- Frontend: `cd ../frontend && pnpm run gen:api` then
  `pnpm exec vue-tsc --noEmit`; replace any hand-written axios calls
  with the regenerated client in `frontend/src/api/v5/`

## How to add a complete system module

For a new **management page** (not just a single endpoint), the full
14-stage closed-loop lives in
[../docs/API_OPENAPI_FIXED_FLOW.md](../docs/API_OPENAPI_FIXED_FLOW.md);
the concrete files/rows to touch are in
[../docs/guides/new-module-checklist.md](../docs/guides/new-module-checklist.md).
Do not restate the pipeline here.

## Anti-patterns

- Do NOT re-introduce a legacy Gin module shell
  (`internal/modules/system/*/module.go`) — those files have been deleted
  and must not be recreated.
- Do NOT pile new ops onto a single god `APIHandler` — every op belongs
  to exactly one `*{domain}APIHandler` sub-handler. Same-named methods
  across two sub-handlers produce ambiguous-selector compile errors,
  which is a signal the domain boundary is wrong.
- Do NOT bypass `internal/pkg/permission/evaluator` for permission
  decisions; never read `feature_package_keys` / `role_feature_packages`
  directly from a handler.
- Do NOT import `internal/pkg/authorization` — the package is empty and
  scheduled for physical deletion.
- Do NOT write tenant-unaware queries; every query must filter on
  `tenant_id` (currently always `default`).
- Do NOT hand-edit `openapi_seed.json` — it is regenerated from the spec.

