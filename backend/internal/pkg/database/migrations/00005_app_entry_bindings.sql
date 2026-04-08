-- +goose Up
-- +goose StatementBegin
-- 扩展 app_host_bindings：支持多种匹配类型 + 路径模式 + 优先级。
-- 旧数据保持 host_exact 行为不变。
ALTER TABLE app_host_bindings
    ADD COLUMN IF NOT EXISTS match_type   varchar(30)  NOT NULL DEFAULT 'host_exact',
    ADD COLUMN IF NOT EXISTS path_pattern varchar(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS priority     integer      NOT NULL DEFAULT 0;

-- 旧 host 字段允许为空，因为 path_prefix 类型可能没有 host。
ALTER TABLE app_host_bindings
    ALTER COLUMN host DROP NOT NULL;
ALTER TABLE app_host_bindings
    ALTER COLUMN host SET DEFAULT '';

-- 旧的唯一索引限制 host 全局唯一，无法支持 host_and_path / path_prefix。
DROP INDEX IF EXISTS idx_app_host_bindings_host;
CREATE UNIQUE INDEX IF NOT EXISTS uk_app_entry_binding_rule
    ON app_host_bindings (match_type, host, path_pattern)
    WHERE deleted_at IS NULL;

-- Level 2: 菜单空间入口解析绑定。
-- 不复用 menu_space_host_bindings（那是 SSO/Cookie 配置，语义不同）。
CREATE TABLE IF NOT EXISTS menu_space_entry_bindings (
    id           uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    app_key      varchar(100) NOT NULL,
    space_key    varchar(100) NOT NULL,
    match_type   varchar(30)  NOT NULL DEFAULT 'host_exact',
    host         varchar(255) NOT NULL DEFAULT '',
    path_pattern varchar(255) NOT NULL DEFAULT '',
    priority     integer      NOT NULL DEFAULT 0,
    is_primary   boolean      NOT NULL DEFAULT false,
    description  text         NOT NULL DEFAULT '',
    status       varchar(20)  NOT NULL DEFAULT 'normal',
    meta         jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

CREATE INDEX IF NOT EXISTS idx_menu_space_entry_bindings_app_key
    ON menu_space_entry_bindings (app_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_menu_space_entry_bindings_space_key
    ON menu_space_entry_bindings (space_key)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_menu_space_entry_binding_rule
    ON menu_space_entry_bindings (app_key, match_type, host, path_pattern)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS menu_space_entry_bindings;
DROP INDEX IF EXISTS uk_app_entry_binding_rule;
ALTER TABLE app_host_bindings DROP COLUMN IF EXISTS match_type;
ALTER TABLE app_host_bindings DROP COLUMN IF EXISTS path_pattern;
ALTER TABLE app_host_bindings DROP COLUMN IF EXISTS priority;
-- +goose StatementEnd
