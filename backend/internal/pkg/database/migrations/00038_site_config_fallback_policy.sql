-- +goose Up
-- +goose StatementBegin
ALTER TABLE site_configs
    ADD COLUMN IF NOT EXISTS fallback_policy varchar(20) NOT NULL DEFAULT 'inherit';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE site_configs
    DROP COLUMN IF EXISTS fallback_policy;
-- +goose StatementEnd
