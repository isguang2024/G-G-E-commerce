-- +goose Up
-- +goose StatementBegin
-- 站点配置项（全局 app_key='' / 应用级 app_key='xxx'）
CREATE TABLE IF NOT EXISTS site_configs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    app_key varchar(100) NOT NULL DEFAULT '',
    config_key varchar(150) NOT NULL,
    config_value jsonb NOT NULL DEFAULT '{}'::jsonb,
    value_type varchar(32) NOT NULL DEFAULT 'string',
    label varchar(200) NOT NULL DEFAULT '',
    description varchar(500) NOT NULL DEFAULT '',
    sort_order int NOT NULL DEFAULT 0,
    is_builtin boolean NOT NULL DEFAULT false,
    status varchar(20) NOT NULL DEFAULT 'normal',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

-- 全局配置唯一（app_key 为空）
CREATE UNIQUE INDEX IF NOT EXISTS uidx_site_configs_global
    ON site_configs(tenant_id, config_key)
    WHERE app_key = '' AND deleted_at IS NULL;

-- 应用级配置唯一（app_key 非空）
CREATE UNIQUE INDEX IF NOT EXISTS uidx_site_configs_app
    ON site_configs(tenant_id, app_key, config_key)
    WHERE app_key != '' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_site_configs_tenant_key
    ON site_configs(tenant_id, config_key)
    WHERE deleted_at IS NULL;

-- 配置集合（纯分组元数据）
CREATE TABLE IF NOT EXISTS site_config_sets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    set_code varchar(100) NOT NULL,
    set_name varchar(200) NOT NULL,
    description varchar(500) NOT NULL DEFAULT '',
    sort_order int NOT NULL DEFAULT 0,
    is_builtin boolean NOT NULL DEFAULT false,
    status varchar(20) NOT NULL DEFAULT 'normal',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_site_config_sets_code
    ON site_config_sets(tenant_id, set_code)
    WHERE deleted_at IS NULL;

-- 集合-Key 关联（多对多；存 config_key 字符串，非外键）
CREATE TABLE IF NOT EXISTS site_config_set_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    set_id uuid NOT NULL REFERENCES site_config_sets(id) ON DELETE CASCADE,
    config_key varchar(150) NOT NULL,
    sort_order int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_site_config_set_items
    ON site_config_set_items(tenant_id, set_id, config_key);

CREATE INDEX IF NOT EXISTS idx_site_config_set_items_key
    ON site_config_set_items(tenant_id, config_key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS site_config_set_items;
DROP TABLE IF EXISTS site_config_sets;
DROP TABLE IF EXISTS site_configs;
-- +goose StatementEnd
