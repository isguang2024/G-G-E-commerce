-- +goose Up
-- +goose StatementBegin
ALTER TABLE register_policies
    DROP COLUMN IF EXISTS app_key,
    DROP COLUMN IF EXISTS default_workspace_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE register_policies
    ADD COLUMN IF NOT EXISTS app_key varchar(64) NOT NULL DEFAULT 'account-portal',
    ADD COLUMN IF NOT EXISTS default_workspace_type varchar(32) NOT NULL DEFAULT 'personal';
-- +goose StatementEnd
