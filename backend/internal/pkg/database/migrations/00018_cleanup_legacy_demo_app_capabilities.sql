-- +goose Up
-- +goose StatementBegin
UPDATE apps
SET
    capabilities = capabilities - 'managed_pages' - 'runtime_navigation' - 'app_switchable',
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND (
      capabilities ? 'managed_pages'
      OR capabilities ? 'runtime_navigation'
      OR capabilities ? 'app_switchable'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
