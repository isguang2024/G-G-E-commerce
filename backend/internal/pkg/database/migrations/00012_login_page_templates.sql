-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS login_page_templates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    template_key varchar(64) NOT NULL,
    name varchar(128) NOT NULL,
    scene varchar(32) NOT NULL DEFAULT 'auth_family',
    app_scope varchar(32) NOT NULL DEFAULT 'shared',
    status varchar(20) NOT NULL DEFAULT 'normal',
    is_default boolean NOT NULL DEFAULT false,
    config jsonb NOT NULL DEFAULT '{}'::jsonb,
    meta jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_login_page_templates_tenant_key
    ON login_page_templates(tenant_id, template_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_login_page_templates_tenant_scene_status
    ON login_page_templates(tenant_id, scene, status)
    WHERE deleted_at IS NULL;

ALTER TABLE register_entries
    ADD COLUMN IF NOT EXISTS login_page_key varchar(64) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE register_entries
    DROP COLUMN IF EXISTS login_page_key;

DROP TABLE IF EXISTS login_page_templates;
-- +goose StatementEnd
