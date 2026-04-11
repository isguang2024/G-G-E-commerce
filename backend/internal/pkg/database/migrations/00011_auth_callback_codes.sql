-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS auth_callback_codes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(100) NOT NULL DEFAULT 'default',
    code varchar(120) NOT NULL,
    user_id uuid NOT NULL,
    target_app_key varchar(100) NOT NULL,
    redirect_uri text NOT NULL,
    target_path varchar(500) NOT NULL DEFAULT '',
    navigation_space_key varchar(100) NOT NULL DEFAULT '',
    state varchar(200) NOT NULL,
    nonce varchar(200) NOT NULL,
    request_host varchar(255) NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'pending',
    expires_at timestamptz NOT NULL,
    used_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_auth_callback_codes_code
    ON auth_callback_codes(code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_auth_callback_codes_tenant_app_status
    ON auth_callback_codes(tenant_id, target_app_key, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_auth_callback_codes_user_id
    ON auth_callback_codes(user_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_callback_codes;
-- +goose StatementEnd
