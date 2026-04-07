-- +goose Up
-- +goose StatementBegin

-- Phase 2a baseline: introduce the tenant dimension reserved by the
-- GGE 5.0 architecture (doc ch.10). At this stage every account, workspace
-- and downstream record belongs to the built-in "default" tenant. Future
-- migrations expand the tenant_id column onto the rest of the schema.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS tenants (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code        varchar(64)  NOT NULL UNIQUE,
    name        varchar(150) NOT NULL,
    status      varchar(20)  NOT NULL DEFAULT 'active',
    is_default  boolean      NOT NULL DEFAULT false,
    meta        jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenants_default
    ON tenants (is_default)
    WHERE is_default = true AND deleted_at IS NULL;

INSERT INTO tenants (code, name, status, is_default)
VALUES ('default', 'Default Tenant', 'active', true)
ON CONFLICT (code) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tenants;
-- +goose StatementEnd
