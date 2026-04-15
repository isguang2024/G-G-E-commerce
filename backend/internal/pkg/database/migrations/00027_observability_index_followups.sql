-- +goose Up
-- +goose StatementBegin
-- 观察性查询页 telemetry-log 的 event / actor_id 过滤当前会回退到 idx_telemetry_logs_tenant_ts
-- 后再 Filter；高基数 event 列表场景下扫描成本太高。补两条复合索引让查询走 index cond。
CREATE INDEX IF NOT EXISTS idx_telemetry_logs_event_ts
    ON telemetry_logs (tenant_id, event, ts DESC);

CREATE INDEX IF NOT EXISTS idx_telemetry_logs_actor_ts
    ON telemetry_logs (tenant_id, actor_id, ts DESC) WHERE actor_id <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_telemetry_logs_actor_ts;
DROP INDEX IF EXISTS idx_telemetry_logs_event_ts;
-- +goose StatementEnd
