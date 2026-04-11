-- +goose Up
-- +goose StatementBegin
ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS frontend_entry_url varchar(500) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS backend_entry_url varchar(500) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS health_check_url varchar(500) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS capabilities jsonb NOT NULL DEFAULT '{}'::jsonb;

UPDATE apps
SET
    frontend_entry_url = CASE
        WHEN app_key = 'account-portal' THEN '/account'
        WHEN app_key = 'platform-admin' THEN '/'
        ELSE frontend_entry_url
    END,
    health_check_url = CASE
        WHEN COALESCE(TRIM(health_check_url), '') = '' THEN '/health'
        ELSE health_check_url
    END,
    capabilities = CASE
        WHEN app_key = 'account-portal' THEN
            '{
              "routing": {"entry_mode": "path_prefix", "route_prefix": "/account", "supports_public_runtime": true},
              "runtime": {"kind": "local", "supports_dynamic_routes": true, "supports_worktab": false},
              "navigation": {"supports_multi_space": false, "default_landing_mode": "menu_space"},
              "integration": {"supports_app_switch": true, "supports_broadcast_channel": false}
            }'::jsonb
        WHEN app_key = 'platform-admin' THEN
            '{
              "routing": {"entry_mode": "inherit_host", "route_prefix": "/", "supports_public_runtime": false},
              "runtime": {"kind": "local", "supports_dynamic_routes": true, "supports_worktab": true},
              "navigation": {"supports_multi_space": true, "default_landing_mode": "menu_space", "supports_space_badges": true},
              "integration": {"supports_app_switch": true, "supports_broadcast_channel": true}
            }'::jsonb
        WHEN capabilities IS NULL OR capabilities = 'null'::jsonb THEN '{}'::jsonb
        ELSE capabilities
    END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE apps
    DROP COLUMN IF EXISTS frontend_entry_url,
    DROP COLUMN IF EXISTS backend_entry_url,
    DROP COLUMN IF EXISTS health_check_url,
    DROP COLUMN IF EXISTS capabilities;
-- +goose StatementEnd
