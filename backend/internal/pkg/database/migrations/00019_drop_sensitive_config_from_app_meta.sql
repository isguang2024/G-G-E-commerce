-- +goose Up
-- +goose StatementBegin
UPDATE apps
SET
    meta = meta - 'sensitive_config',
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND meta ? 'sensitive_config';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
