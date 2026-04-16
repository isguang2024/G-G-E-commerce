-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_providers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    provider_key varchar(100) NOT NULL,
    name varchar(200) NOT NULL,
    driver varchar(32) NOT NULL,
    endpoint text NOT NULL DEFAULT '',
    region varchar(100) NOT NULL DEFAULT '',
    base_url text NOT NULL DEFAULT '',
    access_key_encrypted text NOT NULL DEFAULT '',
    secret_key_encrypted text NOT NULL DEFAULT '',
    extra jsonb NOT NULL DEFAULT '{}'::jsonb,
    is_default boolean NOT NULL DEFAULT false,
    status varchar(20) NOT NULL DEFAULT 'ready',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_storage_providers_tenant_key
    ON storage_providers(tenant_id, provider_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_storage_providers_tenant_default
    ON storage_providers(tenant_id, is_default)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS storage_buckets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    provider_id uuid NOT NULL REFERENCES storage_providers(id),
    bucket_key varchar(100) NOT NULL,
    name varchar(200) NOT NULL,
    bucket_name varchar(200) NOT NULL,
    base_path varchar(500) NOT NULL DEFAULT '',
    public_base_url text NOT NULL DEFAULT '',
    is_public boolean NOT NULL DEFAULT true,
    status varchar(20) NOT NULL DEFAULT 'ready',
    extra jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_storage_buckets_tenant_key
    ON storage_buckets(tenant_id, bucket_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_storage_buckets_tenant_provider
    ON storage_buckets(tenant_id, provider_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS upload_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    bucket_id uuid NOT NULL REFERENCES storage_buckets(id),
    key varchar(150) NOT NULL,
    name varchar(200) NOT NULL,
    path_template varchar(500) NOT NULL DEFAULT '',
    default_rule_key varchar(150) NOT NULL DEFAULT '',
    max_size_bytes bigint NOT NULL DEFAULT 0,
    allowed_mime_types jsonb NOT NULL DEFAULT '[]'::jsonb,
    visibility varchar(20) NOT NULL DEFAULT 'public',
    status varchar(20) NOT NULL DEFAULT 'ready',
    meta jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_upload_keys_tenant_key
    ON upload_keys(tenant_id, key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_upload_keys_tenant_bucket
    ON upload_keys(tenant_id, bucket_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS upload_key_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    upload_key_id uuid NOT NULL REFERENCES upload_keys(id),
    rule_key varchar(150) NOT NULL,
    name varchar(200) NOT NULL,
    sub_path varchar(255) NOT NULL DEFAULT '',
    filename_strategy varchar(50) NOT NULL DEFAULT 'uuid',
    max_size_bytes bigint NOT NULL DEFAULT 0,
    allowed_mime_types jsonb NOT NULL DEFAULT '[]'::jsonb,
    process_pipeline jsonb NOT NULL DEFAULT '[]'::jsonb,
    is_default boolean NOT NULL DEFAULT false,
    status varchar(20) NOT NULL DEFAULT 'ready',
    meta jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_upload_key_rules_tenant_key
    ON upload_key_rules(tenant_id, upload_key_id, rule_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_upload_key_rules_tenant_default
    ON upload_key_rules(tenant_id, upload_key_id, is_default)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS upload_records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id varchar(64) NOT NULL DEFAULT 'default',
    provider_id uuid NOT NULL REFERENCES storage_providers(id),
    bucket_id uuid NOT NULL REFERENCES storage_buckets(id),
    upload_key_id uuid NOT NULL REFERENCES upload_keys(id),
    rule_id uuid NULL REFERENCES upload_key_rules(id),
    uploaded_by uuid NULL,
    original_filename varchar(500) NOT NULL,
    stored_filename varchar(500) NOT NULL,
    storage_key varchar(1000) NOT NULL,
    url varchar(1000) NOT NULL,
    mime_type varchar(100) NOT NULL DEFAULT '',
    size bigint NOT NULL DEFAULT 0,
    checksum varchar(64) NOT NULL DEFAULT '',
    status varchar(20) NOT NULL DEFAULT 'active',
    meta jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE INDEX IF NOT EXISTS idx_upload_records_tenant_created
    ON upload_records(tenant_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_upload_records_tenant_key
    ON upload_records(tenant_id, upload_key_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS upload_records;
DROP TABLE IF EXISTS upload_key_rules;
DROP TABLE IF EXISTS upload_keys;
DROP TABLE IF EXISTS storage_buckets;
DROP TABLE IF EXISTS storage_providers;
-- +goose StatementEnd
