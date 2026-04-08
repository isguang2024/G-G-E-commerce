-- +goose Up
-- +goose StatementBegin
UPDATE ui_pages SET page_type = 'standalone' WHERE page_type = 'global';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- no-op: merge is irreversible
-- +goose StatementEnd
