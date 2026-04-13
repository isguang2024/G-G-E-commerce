-- +goose Up
-- +goose StatementBegin
BEGIN;

-- 新分组：账号与登录（若不存在则创建）
INSERT INTO menu_definitions (
    app_key, menu_key, kind, path, name, component, page_key, permission_key,
    default_title, default_icon, status, meta, created_at, updated_at
)
SELECT
    'platform-admin', 'SystemAccount', 'directory', 'account', 'SystemAccount', '', '', '',
    '账号与登录', 'ri:account-pin-circle-line', 'normal', '{}'::jsonb, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM menu_definitions
    WHERE app_key = 'platform-admin' AND menu_key = 'SystemAccount' AND deleted_at IS NULL
);

-- 菜单入口：访问链路测试（若不存在则创建）
INSERT INTO menu_definitions (
    app_key, menu_key, kind, path, name, component, page_key, permission_key,
    default_title, default_icon, status, meta, created_at, updated_at
)
SELECT
    'platform-admin', 'AccessTrace', 'entry', '/system/access-trace', 'AccessTrace', '/system/access-trace', '', '',
    '访问链路测试', '', 'normal', '{}'::jsonb, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM menu_definitions
    WHERE app_key = 'platform-admin' AND menu_key = 'AccessTrace' AND deleted_at IS NULL
);

-- 一级菜单排序与标题归位
UPDATE space_menu_placements
SET parent_menu_key = 'System', sort_order = 1, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'SystemAccess' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'System', sort_order = 2, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'SystemNavigation' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'System', sort_order = 3, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'SystemAccount' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'System', sort_order = 4, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'SystemIntegration' AND deleted_at IS NULL;

-- 导航与界面
UPDATE space_menu_placements
SET parent_menu_key = 'SystemNavigation', sort_order = 1, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'AppManage' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemNavigation', sort_order = 2, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'Menus' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemNavigation', sort_order = 3, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'PageManagement' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemNavigation', sort_order = 4, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'FastEnterManage' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemNavigation', sort_order = 5, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'MenuSpaceManage' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemNavigation', sort_order = 6, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'AccessTrace' AND deleted_at IS NULL;

-- 账号与登录
UPDATE space_menu_placements
SET parent_menu_key = 'SystemAccount', sort_order = 1, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'RegisterEntry' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemAccount', sort_order = 2, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'RegisterPolicy' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemAccount', sort_order = 3, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'RegisterLog' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemAccount', sort_order = 4, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'LoginPageTemplate' AND deleted_at IS NULL;

-- 开放接口与消息
UPDATE space_menu_placements
SET parent_menu_key = 'SystemIntegration', sort_order = 1, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'ApiEndpoint' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemIntegration', sort_order = 2, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'MessageManage' AND deleted_at IS NULL;

UPDATE space_menu_placements
SET parent_menu_key = 'SystemIntegration', sort_order = 3, updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'Dictionary' AND deleted_at IS NULL;

-- 标题同步（避免旧标题残留）
UPDATE menu_definitions
SET default_title = '开放接口与消息', default_icon = 'ri:link-m', updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'SystemIntegration' AND deleted_at IS NULL;

UPDATE menu_definitions
SET default_title = '账号与登录', default_icon = 'ri:account-pin-circle-line', updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'SystemAccount' AND deleted_at IS NULL;

UPDATE menu_definitions
SET default_title = '访问链路测试', updated_at = NOW()
WHERE app_key = 'platform-admin' AND menu_key = 'AccessTrace' AND deleted_at IS NULL;

-- 为现有菜单空间补齐新分组/菜单挂载
INSERT INTO space_menu_placements (
    app_key, space_key, menu_key, parent_menu_key, sort_order, hidden,
    title_override, icon_override, meta_override, created_at, updated_at
)
SELECT
    'platform-admin', s.space_key, 'SystemAccount', 'System', 3, FALSE,
    '', '', '{}'::jsonb, NOW(), NOW()
FROM menu_spaces s
WHERE s.app_key = 'platform-admin' AND s.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM space_menu_placements p
      WHERE p.app_key = 'platform-admin'
        AND p.space_key = s.space_key
        AND p.menu_key = 'SystemAccount'
        AND p.deleted_at IS NULL
  );

INSERT INTO space_menu_placements (
    app_key, space_key, menu_key, parent_menu_key, sort_order, hidden,
    title_override, icon_override, meta_override, created_at, updated_at
)
SELECT
    'platform-admin', s.space_key, 'AccessTrace', 'SystemNavigation', 6, FALSE,
    '', '', '{}'::jsonb, NOW(), NOW()
FROM menu_spaces s
WHERE s.app_key = 'platform-admin' AND s.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM space_menu_placements p
      WHERE p.app_key = 'platform-admin'
        AND p.space_key = s.space_key
        AND p.menu_key = 'AccessTrace'
        AND p.deleted_at IS NULL
  );

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;

-- 仅回退新增菜单，排序/父子关系不做强制回滚。
UPDATE menu_definitions
SET deleted_at = NOW(), updated_at = NOW()
WHERE app_key = 'platform-admin'
  AND menu_key IN ('SystemAccount', 'AccessTrace')
  AND deleted_at IS NULL;

COMMIT;
-- +goose StatementEnd

