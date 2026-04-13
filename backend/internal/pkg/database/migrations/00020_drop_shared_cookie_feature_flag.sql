-- +goose Up
-- +goose StatementBegin
UPDATE apps
SET
    meta = jsonb_set(
        meta,
        '{feature_flags}',
        (meta -> 'feature_flags') - 'shared_cookie'
    ),
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND meta ? 'feature_flags'
  AND jsonb_typeof(meta -> 'feature_flags') = 'object'
  AND (meta -> 'feature_flags') ? 'shared_cookie';

UPDATE apps
SET
    meta = meta - 'feature_flags',
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND meta ? 'feature_flags'
  AND jsonb_typeof(meta -> 'feature_flags') = 'object'
  AND meta -> 'feature_flags' = '{}'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
