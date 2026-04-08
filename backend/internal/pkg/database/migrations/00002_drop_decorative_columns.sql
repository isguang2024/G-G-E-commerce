-- +goose Up
-- +goose StatementBegin

-- Drop "decorative" columns that were never enforced at runtime.
-- See docs/guides/permission-audit.md for the audit that confirmed these
-- fields had zero enforcement in the middleware or evaluator.
--
-- api_endpoints:   app_scope / app_key / feature_kind / context_scope / source
-- permission_keys: app_key / module_code / context_type / feature_kind /
--                  allowed_workspace_types
--
-- Business-legitimate uses of the same identifiers on OTHER tables
-- (apps, menus, pages, feature_packages, ...) are unaffected.

ALTER TABLE api_endpoints
    DROP COLUMN IF EXISTS app_scope,
    DROP COLUMN IF EXISTS app_key,
    DROP COLUMN IF EXISTS feature_kind,
    DROP COLUMN IF EXISTS context_scope,
    DROP COLUMN IF EXISTS source;

ALTER TABLE permission_keys
    DROP COLUMN IF EXISTS app_key,
    DROP COLUMN IF EXISTS module_code,
    DROP COLUMN IF EXISTS context_type,
    DROP COLUMN IF EXISTS feature_kind,
    DROP COLUMN IF EXISTS allowed_workspace_types;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Intentionally empty: the dropped columns were decorative and unused
-- at runtime. Rolling back a database to contain them again would not
-- restore any behaviour.
SELECT 1;
-- +goose StatementEnd
