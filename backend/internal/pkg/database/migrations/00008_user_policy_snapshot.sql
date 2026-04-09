-- +goose Up
-- +goose StatementBegin
-- 注册体系：在 users 表冻结注册时刻的有效策略快照，防止策略修改后历史用户数据漂移。
-- 字段包含注册时命中的布尔开关、目标 App/Space、角色 code 列表和功能包 key 列表。
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS register_policy_snapshot jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS register_policy_snapshot;
-- +goose StatementEnd
