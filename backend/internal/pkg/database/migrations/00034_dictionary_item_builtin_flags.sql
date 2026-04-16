-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.dict_items') IS NOT NULL THEN
        ALTER TABLE dict_items
            ADD COLUMN IF NOT EXISTS is_builtin boolean NOT NULL DEFAULT false;

        UPDATE dict_items
           SET is_builtin = true,
               updated_at = NOW()
          FROM dict_types dt,
               (VALUES
                ('common_status', 'normal'),
                ('common_status', 'suspended'),
                ('gender', 'male'),
                ('gender', 'female'),
                ('page_type', 'group'),
                ('page_type', 'display_group'),
                ('page_type', 'inner'),
                ('page_type', 'standalone'),
                ('access_mode', 'inherit'),
                ('access_mode', 'public'),
                ('access_mode', 'jwt'),
                ('access_mode', 'permission'),
                ('page_source', 'manual'),
                ('page_source', 'sync'),
                ('page_source', 'seed'),
                ('page_source', 'remote'),
                ('http_method', 'GET'),
                ('http_method', 'POST'),
                ('http_method', 'PUT'),
                ('http_method', 'PATCH'),
                ('http_method', 'DELETE'),
                ('message_type', 'notice'),
                ('message_type', 'message'),
                ('message_type', 'todo'),
                ('workspace_plan', 'free'),
                ('workspace_plan', 'pro'),
                ('workspace_plan', 'enterprise'),
                ('register_source', 'self'),
                ('register_source', 'invite'),
                ('register_source', 'admin'),
                ('register_source', 'email'),
                ('register_source', 'sms'),
                ('register_source', 'oauth')
               ) AS v(type_code, item_value)
         WHERE dt.id = dict_items.dict_type_id
           AND dt.tenant_id = dict_items.tenant_id
           AND dict_items.tenant_id = 'default'
           AND dt.code = v.type_code
           AND dict_items.value = v.item_value;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.dict_items') IS NOT NULL THEN
        ALTER TABLE dict_items
            DROP COLUMN IF EXISTS is_builtin;
    END IF;
END $$;
-- +goose StatementEnd
