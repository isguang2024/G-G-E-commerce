-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS audit_logs (
    id               BIGSERIAL PRIMARY KEY,
    ts               TIMESTAMPTZ  NOT NULL DEFAULT now(),
    request_id       varchar(64)  NOT NULL DEFAULT '',
    tenant_id        varchar(64)  NOT NULL DEFAULT 'default',
    actor_id         varchar(64)  NOT NULL DEFAULT '',
    actor_type       varchar(32)  NOT NULL DEFAULT 'anonymous',
    app_key          varchar(64)  NOT NULL DEFAULT '',
    workspace_id     varchar(64)  NOT NULL DEFAULT '',
    action           varchar(128) NOT NULL,
    resource_type    varchar(64)  NOT NULL DEFAULT '',
    resource_id      varchar(128) NOT NULL DEFAULT '',
    outcome          varchar(16)  NOT NULL DEFAULT 'success',
    error_code       varchar(32)  NOT NULL DEFAULT '',
    http_status      integer      NOT NULL DEFAULT 0,
    ip               varchar(64)  NOT NULL DEFAULT '',
    user_agent       text         NOT NULL DEFAULT '',
    before_json      jsonb        NOT NULL DEFAULT 'null'::jsonb,
    after_json       jsonb        NOT NULL DEFAULT 'null'::jsonb,
    metadata         jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_ts
    ON audit_logs (tenant_id, ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_ts
    ON audit_logs (tenant_id, actor_id, ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_ts
    ON audit_logs (tenant_id, resource_type, resource_id, ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_ts
    ON audit_logs (tenant_id, action, ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id
    ON audit_logs (request_id) WHERE request_id <> '';
CREATE INDEX IF NOT EXISTS idx_audit_logs_outcome_ts
    ON audit_logs (tenant_id, outcome, ts DESC) WHERE outcome <> 'success';

CREATE TABLE IF NOT EXISTS telemetry_logs (
    id           BIGSERIAL PRIMARY KEY,
    ts           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    request_id   varchar(64)  NOT NULL DEFAULT '',
    session_id   varchar(64)  NOT NULL DEFAULT '',
    tenant_id    varchar(64)  NOT NULL DEFAULT 'default',
    actor_id     varchar(64)  NOT NULL DEFAULT '',
    app_key      varchar(64)  NOT NULL DEFAULT '',
    level        varchar(16)  NOT NULL DEFAULT 'info',
    event        varchar(128) NOT NULL,
    message      text         NOT NULL DEFAULT '',
    url          text         NOT NULL DEFAULT '',
    user_agent   text         NOT NULL DEFAULT '',
    ip           varchar(64)  NOT NULL DEFAULT '',
    release      varchar(64)  NOT NULL DEFAULT '',
    payload      jsonb        NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_logs_tenant_ts
    ON telemetry_logs (tenant_id, ts DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_logs_level_ts
    ON telemetry_logs (tenant_id, level, ts DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_logs_session
    ON telemetry_logs (session_id, ts DESC) WHERE session_id <> '';
CREATE INDEX IF NOT EXISTS idx_telemetry_logs_request
    ON telemetry_logs (request_id) WHERE request_id <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS telemetry_logs;
DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd
