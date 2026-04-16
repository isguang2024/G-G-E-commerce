-- +goose Up
-- +goose StatementBegin
ALTER TABLE upload_keys
    ADD COLUMN IF NOT EXISTS upload_mode varchar(20) NOT NULL DEFAULT 'auto',
    ADD COLUMN IF NOT EXISTS is_frontend_visible boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS permission_key varchar(150) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fallback_key varchar(150) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS client_accept jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS direct_size_threshold_bytes bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS extra_schema jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE upload_key_rules
    ADD COLUMN IF NOT EXISTS mode_override varchar(20) NOT NULL DEFAULT 'inherit',
    ADD COLUMN IF NOT EXISTS visibility_override varchar(20) NOT NULL DEFAULT 'inherit',
    ADD COLUMN IF NOT EXISTS client_accept jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS extra_schema jsonb NOT NULL DEFAULT '{}'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE upload_key_rules
    DROP COLUMN IF EXISTS extra_schema,
    DROP COLUMN IF EXISTS client_accept,
    DROP COLUMN IF EXISTS visibility_override,
    DROP COLUMN IF EXISTS mode_override;

ALTER TABLE upload_keys
    DROP COLUMN IF EXISTS extra_schema,
    DROP COLUMN IF EXISTS direct_size_threshold_bytes,
    DROP COLUMN IF EXISTS client_accept,
    DROP COLUMN IF EXISTS fallback_key,
    DROP COLUMN IF EXISTS permission_key,
    DROP COLUMN IF EXISTS is_frontend_visible,
    DROP COLUMN IF EXISTS upload_mode;
-- +goose StatementEnd
