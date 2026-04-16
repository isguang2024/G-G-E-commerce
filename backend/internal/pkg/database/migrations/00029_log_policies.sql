-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS log_policies (
    id          uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   varchar(64)  NOT NULL DEFAULT 'default',
    pipeline    varchar(16)  NOT NULL,
    match_field varchar(64)  NOT NULL,
    pattern     varchar(256) NOT NULL,
    decision    varchar(16)  NOT NULL,
    sample_rate integer,
    priority    integer      NOT NULL DEFAULT 0,
    enabled     boolean      NOT NULL DEFAULT true,
    note        text         NOT NULL DEFAULT '',
    created_by  uuid,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT chk_log_policies_pipeline CHECK (pipeline IN ('audit', 'telemetry')),
    CONSTRAINT chk_log_policies_decision CHECK (decision IN ('allow', 'deny', 'sample')),
    CONSTRAINT uq_log_policies_rule UNIQUE (tenant_id, pipeline, match_field, pattern)
);

CREATE INDEX IF NOT EXISTS idx_log_policies_tenant_pipeline_enabled_priority
    ON log_policies (tenant_id, pipeline, enabled, priority DESC, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS log_policies;
-- +goose StatementEnd
