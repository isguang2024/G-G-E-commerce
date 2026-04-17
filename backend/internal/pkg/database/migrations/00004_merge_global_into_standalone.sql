-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'ui_pages'
    ) THEN
        UPDATE ui_pages SET page_type = 'standalone' WHERE page_type = 'global';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- no-op: merge is irreversible
-- +goose StatementEnd
