-- +goose Up
ALTER TABLE register_entries DROP COLUMN IF EXISTS policy_code;
ALTER TABLE users DROP COLUMN IF EXISTS register_policy_code;
DROP TABLE IF EXISTS register_policy_feature_packages;
DROP TABLE IF EXISTS register_policy_roles;
DROP TABLE IF EXISTS register_policies;

-- +goose Down
CREATE TABLE IF NOT EXISTS register_policies (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    code TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    allow_public_register BOOLEAN NOT NULL DEFAULT FALSE,
    require_invite BOOLEAN NOT NULL DEFAULT FALSE,
    require_email_verify BOOLEAN NOT NULL DEFAULT FALSE,
    require_captcha BOOLEAN NOT NULL DEFAULT FALSE,
    auto_login BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS register_policy_roles (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    register_policy_id BIGINT,
    role_code TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS register_policy_feature_packages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    register_policy_id BIGINT,
    feature_package_key TEXT NOT NULL DEFAULT ''
);

ALTER TABLE register_entries ADD COLUMN IF NOT EXISTS policy_code TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS register_policy_code TEXT NOT NULL DEFAULT '';
