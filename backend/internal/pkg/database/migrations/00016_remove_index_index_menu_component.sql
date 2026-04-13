-- +goose Up
-- +goose StatementBegin
BEGIN;

UPDATE menu_definitions
SET component = '', updated_at = NOW()
WHERE deleted_at IS NULL
  AND kind = 'directory'
  AND COALESCE(TRIM(component), '') <> '';

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;

UPDATE menu_definitions
SET component = '', updated_at = NOW()
WHERE deleted_at IS NULL
  AND kind = 'directory'
  AND COALESCE(TRIM(component), '') = '';

COMMIT;
-- +goose StatementEnd
