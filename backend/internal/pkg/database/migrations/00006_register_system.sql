-- +goose Up
-- +goose StatementBegin
-- 注册体系第一期：扩展 users 审计字段 + 新增 register_entries / register_policies
-- / register_policy_feature_packages / register_policy_roles 四张表。
-- 详细设计见 docs/register-system-design.md。

-- 1. users 表扩展注册审计字段（现有 register_source / invited_by 已存在，复用）
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS register_app_key      varchar(64)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS register_entry_code   varchar(64)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS register_policy_code  varchar(64)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS register_ip           varchar(64)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS register_user_agent   varchar(512) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS agreement_version     varchar(32)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS email_verified_at     timestamptz,
    ADD COLUMN IF NOT EXISTS invitation_code_id    uuid;

CREATE INDEX IF NOT EXISTS idx_users_register_entry_code
    ON users (register_entry_code)
    WHERE deleted_at IS NULL;

-- 2. register_entries：公开注册入口
CREATE TABLE IF NOT EXISTS register_entries (
    id                     uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    app_key                varchar(64)  NOT NULL,
    entry_code             varchar(64)  NOT NULL,
    name                   varchar(128) NOT NULL,
    host                   varchar(128) NOT NULL DEFAULT '',
    path_prefix            varchar(256) NOT NULL DEFAULT '',
    register_source        varchar(32)  NOT NULL DEFAULT 'self',
    policy_code            varchar(64)  NOT NULL,
    status                 varchar(16)  NOT NULL DEFAULT 'enabled',
    allow_public_register  boolean,
    require_invite         boolean,
    require_email_verify   boolean,
    require_captcha        boolean,
    auto_login             boolean,
    sort_order             integer      NOT NULL DEFAULT 0,
    remark                 text         NOT NULL DEFAULT '',
    created_at             timestamptz  NOT NULL DEFAULT now(),
    updated_at             timestamptz  NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_register_entries_entry_code
    ON register_entries (entry_code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_register_entries_match
    ON register_entries (host, path_prefix, sort_order)
    WHERE deleted_at IS NULL AND status = 'enabled';

-- 3. register_policies：注册策略
CREATE TABLE IF NOT EXISTS register_policies (
    id                            uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    app_key                       varchar(64)  NOT NULL,
    policy_code                   varchar(64)  NOT NULL,
    name                          varchar(128) NOT NULL,
    description                   text         NOT NULL DEFAULT '',
    target_app_key                varchar(64)  NOT NULL,
    target_navigation_space_key   varchar(64)  NOT NULL,
    target_home_path              varchar(256) NOT NULL DEFAULT '',
    default_workspace_type        varchar(32)  NOT NULL DEFAULT 'personal',
    status                        varchar(16)  NOT NULL DEFAULT 'enabled',
    welcome_message_template_key  varchar(128) NOT NULL DEFAULT '',
    allow_public_register         boolean      NOT NULL DEFAULT false,
    require_invite                boolean      NOT NULL DEFAULT false,
    require_email_verify          boolean      NOT NULL DEFAULT false,
    require_captcha               boolean      NOT NULL DEFAULT false,
    auto_login                    boolean      NOT NULL DEFAULT true,
    created_at                    timestamptz  NOT NULL DEFAULT now(),
    updated_at                    timestamptz  NOT NULL DEFAULT now(),
    deleted_at                    timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_register_policies_policy_code
    ON register_policies (policy_code)
    WHERE deleted_at IS NULL;

-- 4. register_policy_feature_packages：策略 → 功能包
CREATE TABLE IF NOT EXISTS register_policy_feature_packages (
    id              uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_code     varchar(64)  NOT NULL,
    package_id      uuid         NOT NULL,
    workspace_scope varchar(32)  NOT NULL DEFAULT 'personal',
    sort_order      integer      NOT NULL DEFAULT 0,
    created_at      timestamptz  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_rpfp_policy_package
    ON register_policy_feature_packages (policy_code, package_id, workspace_scope);

-- 5. register_policy_roles：策略 → 角色
CREATE TABLE IF NOT EXISTS register_policy_roles (
    id              uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_code     varchar(64)  NOT NULL,
    role_id         uuid         NOT NULL,
    workspace_scope varchar(32)  NOT NULL DEFAULT 'personal',
    sort_order      integer      NOT NULL DEFAULT 0,
    created_at      timestamptz  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_rpr_policy_role
    ON register_policy_roles (policy_code, role_id, workspace_scope);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS register_policy_roles;
DROP TABLE IF EXISTS register_policy_feature_packages;
DROP TABLE IF EXISTS register_policies;
DROP TABLE IF EXISTS register_entries;

DROP INDEX IF EXISTS idx_users_register_entry_code;
ALTER TABLE users
    DROP COLUMN IF EXISTS invitation_code_id,
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS agreement_version,
    DROP COLUMN IF EXISTS register_user_agent,
    DROP COLUMN IF EXISTS register_ip,
    DROP COLUMN IF EXISTS register_policy_code,
    DROP COLUMN IF EXISTS register_entry_code,
    DROP COLUMN IF EXISTS register_app_key;
-- +goose StatementEnd
