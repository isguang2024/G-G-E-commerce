-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    v_relkind   "char";
    v_partstrat "char";
    v_from      timestamptz;
    v_to        timestamptz;
    v_next_to   timestamptz;
BEGIN
    SELECT c.relkind, pt.partstrat
      INTO v_relkind, v_partstrat
      FROM pg_class c
      JOIN pg_namespace n ON n.oid = c.relnamespace
 LEFT JOIN pg_partitioned_table pt ON pt.partrelid = c.oid
     WHERE n.nspname = CURRENT_SCHEMA()
       AND c.relname = 'audit_logs';

    IF v_relkind IS NULL THEN
        RAISE EXCEPTION 'audit_logs table not found';
    END IF;

    IF v_partstrat IS DISTINCT FROM 'r' THEN
        ALTER TABLE audit_logs RENAME TO audit_logs_legacy_unpartitioned;

        IF EXISTS (
            SELECT 1
              FROM pg_constraint
             WHERE conrelid = 'audit_logs_legacy_unpartitioned'::regclass
               AND conname = 'audit_logs_pkey'
        ) THEN
            ALTER TABLE audit_logs_legacy_unpartitioned
                RENAME CONSTRAINT audit_logs_pkey TO audit_logs_legacy_unpartitioned_pkey;
        END IF;

        DROP INDEX IF EXISTS idx_audit_logs_tenant_ts;
        DROP INDEX IF EXISTS idx_audit_logs_actor_ts;
        DROP INDEX IF EXISTS idx_audit_logs_resource_ts;
        DROP INDEX IF EXISTS idx_audit_logs_action_ts;
        DROP INDEX IF EXISTS idx_audit_logs_request_id;
        DROP INDEX IF EXISTS idx_audit_logs_outcome_ts;

        IF to_regclass('audit_logs_id_seq') IS NULL THEN
            CREATE SEQUENCE audit_logs_id_seq;
        END IF;

        CREATE TABLE audit_logs (
            id            BIGINT       NOT NULL DEFAULT nextval('audit_logs_id_seq'::regclass),
            ts            TIMESTAMPTZ  NOT NULL DEFAULT now(),
            request_id    varchar(64)  NOT NULL DEFAULT '',
            tenant_id     varchar(64)  NOT NULL DEFAULT 'default',
            actor_id      varchar(64)  NOT NULL DEFAULT '',
            actor_type    varchar(32)  NOT NULL DEFAULT 'anonymous',
            app_key       varchar(64)  NOT NULL DEFAULT '',
            workspace_id  varchar(64)  NOT NULL DEFAULT '',
            action        varchar(128) NOT NULL,
            resource_type varchar(64)  NOT NULL DEFAULT '',
            resource_id   varchar(128) NOT NULL DEFAULT '',
            outcome       varchar(16)  NOT NULL DEFAULT 'success',
            error_code    varchar(32)  NOT NULL DEFAULT '',
            http_status   integer      NOT NULL DEFAULT 0,
            ip            varchar(64)  NOT NULL DEFAULT '',
            user_agent    text         NOT NULL DEFAULT '',
            before_json   jsonb        NOT NULL DEFAULT 'null'::jsonb,
            after_json    jsonb        NOT NULL DEFAULT 'null'::jsonb,
            metadata      jsonb        NOT NULL DEFAULT '{}'::jsonb,
            created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
            CONSTRAINT audit_logs_pkey PRIMARY KEY (id, ts)
        ) PARTITION BY RANGE (ts);

        ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;
    END IF;

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

    v_from := date_trunc('month', now());
    v_to := v_from + interval '1 month';
    v_next_to := v_to + interval '1 month';

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_logs FOR VALUES FROM (%L) TO (%L)',
        format('audit_logs_%s', to_char(v_from, 'YYYY_MM')),
        v_from,
        v_to
    );
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_logs FOR VALUES FROM (%L) TO (%L)',
        format('audit_logs_%s', to_char(v_to, 'YYYY_MM')),
        v_to,
        v_next_to
    );

    EXECUTE 'CREATE TABLE IF NOT EXISTS audit_logs_default PARTITION OF audit_logs DEFAULT';

    IF to_regclass('audit_logs_legacy_unpartitioned') IS NOT NULL THEN
        INSERT INTO audit_logs (
            id, ts, request_id, tenant_id, actor_id, actor_type, app_key, workspace_id,
            action, resource_type, resource_id, outcome, error_code, http_status,
            ip, user_agent, before_json, after_json, metadata, created_at
        )
        SELECT
            id, ts, request_id, tenant_id, actor_id, actor_type, app_key, workspace_id,
            action, resource_type, resource_id, outcome, error_code, http_status,
            ip, user_agent, before_json, after_json, metadata, created_at
          FROM audit_logs_legacy_unpartitioned
      ORDER BY ts, id;

        DROP TABLE audit_logs_legacy_unpartitioned;
    END IF;

    PERFORM setval('audit_logs_id_seq', COALESCE((SELECT MAX(id) FROM audit_logs), 0) + 1, false);
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    v_partstrat "char";
BEGIN
    SELECT pt.partstrat
      INTO v_partstrat
      FROM pg_class c
      JOIN pg_namespace n ON n.oid = c.relnamespace
      JOIN pg_partitioned_table pt ON pt.partrelid = c.oid
     WHERE n.nspname = CURRENT_SCHEMA()
       AND c.relname = 'audit_logs';

    IF v_partstrat IS DISTINCT FROM 'r' THEN
        RETURN;
    END IF;

    ALTER TABLE audit_logs RENAME TO audit_logs_partitioned_backup;

    IF EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'audit_logs_partitioned_backup'::regclass
           AND conname = 'audit_logs_pkey'
    ) THEN
        ALTER TABLE audit_logs_partitioned_backup
            RENAME CONSTRAINT audit_logs_pkey TO audit_logs_partitioned_backup_pkey;
    END IF;

    DROP INDEX IF EXISTS idx_audit_logs_tenant_ts;
    DROP INDEX IF EXISTS idx_audit_logs_actor_ts;
    DROP INDEX IF EXISTS idx_audit_logs_resource_ts;
    DROP INDEX IF EXISTS idx_audit_logs_action_ts;
    DROP INDEX IF EXISTS idx_audit_logs_request_id;
    DROP INDEX IF EXISTS idx_audit_logs_outcome_ts;

    IF to_regclass('audit_logs_id_seq') IS NULL THEN
        CREATE SEQUENCE audit_logs_id_seq;
    END IF;

    CREATE TABLE audit_logs (
        id            BIGINT       NOT NULL DEFAULT nextval('audit_logs_id_seq'::regclass),
        ts            TIMESTAMPTZ  NOT NULL DEFAULT now(),
        request_id    varchar(64)  NOT NULL DEFAULT '',
        tenant_id     varchar(64)  NOT NULL DEFAULT 'default',
        actor_id      varchar(64)  NOT NULL DEFAULT '',
        actor_type    varchar(32)  NOT NULL DEFAULT 'anonymous',
        app_key       varchar(64)  NOT NULL DEFAULT '',
        workspace_id  varchar(64)  NOT NULL DEFAULT '',
        action        varchar(128) NOT NULL,
        resource_type varchar(64)  NOT NULL DEFAULT '',
        resource_id   varchar(128) NOT NULL DEFAULT '',
        outcome       varchar(16)  NOT NULL DEFAULT 'success',
        error_code    varchar(32)  NOT NULL DEFAULT '',
        http_status   integer      NOT NULL DEFAULT 0,
        ip            varchar(64)  NOT NULL DEFAULT '',
        user_agent    text         NOT NULL DEFAULT '',
        before_json   jsonb        NOT NULL DEFAULT 'null'::jsonb,
        after_json    jsonb        NOT NULL DEFAULT 'null'::jsonb,
        metadata      jsonb        NOT NULL DEFAULT '{}'::jsonb,
        created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
        CONSTRAINT audit_logs_pkey PRIMARY KEY (id)
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

    INSERT INTO audit_logs (
        id, ts, request_id, tenant_id, actor_id, actor_type, app_key, workspace_id,
        action, resource_type, resource_id, outcome, error_code, http_status,
        ip, user_agent, before_json, after_json, metadata, created_at
    )
    SELECT
        id, ts, request_id, tenant_id, actor_id, actor_type, app_key, workspace_id,
        action, resource_type, resource_id, outcome, error_code, http_status,
        ip, user_agent, before_json, after_json, metadata, created_at
      FROM audit_logs_partitioned_backup
  ORDER BY ts, id;

    ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;
    PERFORM setval('audit_logs_id_seq', COALESCE((SELECT MAX(id) FROM audit_logs), 0) + 1, false);

    DROP TABLE audit_logs_partitioned_backup CASCADE;
END;
$$;
-- +goose StatementEnd
