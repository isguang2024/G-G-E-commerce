-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS social_auth_providers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    provider_key varchar(64) NOT NULL,
    provider_name varchar(128) NOT NULL,
    auth_url text NOT NULL DEFAULT '',
    token_url text NOT NULL DEFAULT '',
    user_info_url text NOT NULL DEFAULT '',
    scope text NOT NULL DEFAULT '',
    client_id text NOT NULL DEFAULT '',
    client_secret text NOT NULL DEFAULT '',
    redirect_uri text NOT NULL DEFAULT '',
    enabled boolean NOT NULL DEFAULT false,
    config jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_social_auth_providers_tenant_key
    ON social_auth_providers(tenant_id, provider_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_social_auth_providers_tenant_enabled
    ON social_auth_providers(tenant_id, enabled)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS user_social_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    user_id uuid NOT NULL,
    provider_key varchar(64) NOT NULL,
    provider_uid varchar(255) NOT NULL,
    provider_username varchar(255) NOT NULL DEFAULT '',
    provider_email varchar(255) NOT NULL DEFAULT '',
    avatar_url text NOT NULL DEFAULT '',
    profile jsonb NOT NULL DEFAULT '{}'::jsonb,
    linked_at timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_user_social_accounts_tenant_provider_uid
    ON user_social_accounts(tenant_id, provider_key, provider_uid)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_user_social_accounts_tenant_user_provider
    ON user_social_accounts(tenant_id, user_id, provider_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_social_accounts_tenant_user
    ON user_social_accounts(tenant_id, user_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_social_accounts;
DROP TABLE IF EXISTS social_auth_providers;
-- +goose StatementEnd
