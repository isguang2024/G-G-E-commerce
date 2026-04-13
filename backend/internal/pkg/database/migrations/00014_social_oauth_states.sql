-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TABLE IF NOT EXISTS social_oauth_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    provider_key VARCHAR(64) NOT NULL,
    state VARCHAR(128) NOT NULL,
    login_page_key VARCHAR(128) NOT NULL DEFAULT '',
    page_scene VARCHAR(32) NOT NULL DEFAULT 'login',
    target_app_key VARCHAR(128) NOT NULL DEFAULT '',
    request_path TEXT NOT NULL DEFAULT '',
    redirect_uri TEXT NOT NULL DEFAULT '',
    nonce VARCHAR(128) NOT NULL DEFAULT '',
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_social_oauth_states_state_active
    ON social_oauth_states(tenant_id, provider_key, state)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_social_oauth_states_expires
    ON social_oauth_states(expires_at)
    WHERE deleted_at IS NULL;

COMMIT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS social_oauth_states;
-- +goose StatementEnd