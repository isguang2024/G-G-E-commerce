-- +goose Up
-- +goose StatementBegin
UPDATE apps
SET
    meta = meta - 'env_profiles',
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND meta ? 'env_profiles';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
