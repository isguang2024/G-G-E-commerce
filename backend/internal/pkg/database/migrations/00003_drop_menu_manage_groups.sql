-- +goose Up
-- +goose StatementBegin

-- Drop the MenuManageGroup feature entirely. 菜单空间 + 菜单树已经提供两级
-- 分层，MenuManageGroup 只是管理页的装饰性折叠标签，不参与运行时路由/
-- 权限，属于冗余概念，故物理删除。

DROP INDEX IF EXISTS idx_menus_manage_group_id;
DROP INDEX IF EXISTS idx_menu_manage_groups_name_unique;
DROP INDEX IF EXISTS idx_menu_manage_groups_sort_status;

ALTER TABLE IF EXISTS menus
    DROP COLUMN IF EXISTS manage_group_id;

ALTER TABLE IF EXISTS space_menu_placements
    DROP COLUMN IF EXISTS manage_group_id;

DROP TABLE IF EXISTS menu_manage_groups;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Intentionally empty: the MenuManageGroup feature has been removed from
-- the application code and cannot be restored by a schema rollback alone.
SELECT 1;
-- +goose StatementEnd
