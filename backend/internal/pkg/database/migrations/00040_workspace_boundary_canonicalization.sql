-- +goose Up
-- workspace/collaboration 单主域收口第二阶段：
-- 1. 协作边界与快照切到 workspace_* 真相
-- 2. 回填 workspace_members / workspace_role_bindings / workspace_feature_packages
-- 3. 旧 collaboration_workspace_* 表保留为迁移来源，不再作为运行时主真相

CREATE TABLE IF NOT EXISTS workspace_blocked_menus (
  app_key VARCHAR(100) NOT NULL DEFAULT 'platform-admin',
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  menu_id UUID NOT NULL REFERENCES menu_definitions(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (app_key, workspace_id, menu_id)
);

CREATE TABLE IF NOT EXISTS workspace_blocked_actions (
  app_key VARCHAR(100) NOT NULL DEFAULT 'platform-admin',
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  action_id UUID NOT NULL REFERENCES permission_keys(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (app_key, workspace_id, action_id)
);

CREATE TABLE IF NOT EXISTS workspace_access_snapshots (
  app_key VARCHAR(100) NOT NULL DEFAULT 'platform-admin',
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  package_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  expanded_package_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  derived_action_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  derived_action_map JSONB NOT NULL DEFAULT '{}'::jsonb,
  blocked_action_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  effective_action_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  derived_menu_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  derived_menu_map JSONB NOT NULL DEFAULT '{}'::jsonb,
  blocked_menu_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  effective_menu_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (app_key, workspace_id)
);

CREATE TABLE IF NOT EXISTS workspace_role_access_snapshots (
  app_key VARCHAR(100) NOT NULL DEFAULT 'platform-admin',
  workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  package_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  expanded_package_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  available_action_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  disabled_action_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  action_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  action_source_map JSONB NOT NULL DEFAULT '{}'::jsonb,
  available_menu_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  hidden_menu_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  menu_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  menu_source_map JSONB NOT NULL DEFAULT '{}'::jsonb,
  inherited BOOLEAN NOT NULL DEFAULT FALSE,
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (app_key, workspace_id, role_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_workspace_blocked_menus_unique
  ON workspace_blocked_menus (app_key, workspace_id, menu_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_workspace_blocked_actions_unique
  ON workspace_blocked_actions (app_key, workspace_id, action_id);

INSERT INTO workspace_members (
  workspace_id,
  user_id,
  member_type,
  status,
  collaboration_workspace_member_id,
  created_at,
  updated_at
)
SELECT
  w.id,
  cwm.user_id,
  CASE
    WHEN LOWER(COALESCE(cwm.role_code, '')) IN ('collaboration_workspace_admin', 'admin', 'owner') THEN 'admin'
    ELSE 'member'
  END,
  CASE
    WHEN LOWER(COALESCE(cwm.status, 'active')) IN ('disabled', 'inactive', 'removed') THEN 'inactive'
    ELSE 'active'
  END,
  cwm.id,
  COALESCE(cwm.created_at, CURRENT_TIMESTAMP),
  COALESCE(cwm.updated_at, CURRENT_TIMESTAMP)
FROM collaboration_workspace_members cwm
JOIN workspaces w
  ON w.workspace_type = 'collaboration'
 AND w.collaboration_workspace_id = cwm.collaboration_workspace_id
 AND w.deleted_at IS NULL
WHERE cwm.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM workspace_members wm
    WHERE wm.workspace_id = w.id
      AND wm.user_id = cwm.user_id
      AND wm.deleted_at IS NULL
  );

INSERT INTO workspace_feature_packages (
  workspace_id,
  package_id,
  enabled,
  created_at,
  updated_at,
  deleted_at
)
SELECT
  w.id,
  cwfp.package_id,
  cwfp.enabled,
  COALESCE(cwfp.created_at, CURRENT_TIMESTAMP),
  COALESCE(cwfp.updated_at, CURRENT_TIMESTAMP),
  NULL::TIMESTAMPTZ
FROM collaboration_workspace_feature_packages cwfp
JOIN workspaces w
  ON w.workspace_type = 'collaboration'
 AND w.collaboration_workspace_id = cwfp.collaboration_workspace_id
 AND w.deleted_at IS NULL
WHERE NOT EXISTS (
  SELECT 1
  FROM workspace_feature_packages wfp
  WHERE wfp.workspace_id = w.id
    AND wfp.package_id = cwfp.package_id
    AND wfp.deleted_at IS NULL
);

INSERT INTO workspace_feature_packages (
  workspace_id,
  package_id,
  enabled,
  created_at,
  updated_at,
  deleted_at
)
SELECT
  w.id,
  ufp.package_id,
  ufp.enabled,
  COALESCE(ufp.created_at, CURRENT_TIMESTAMP),
  COALESCE(ufp.updated_at, CURRENT_TIMESTAMP),
  NULL::TIMESTAMPTZ
FROM user_feature_packages ufp
JOIN workspaces w
  ON w.workspace_type = 'personal'
 AND w.owner_user_id = ufp.user_id
 AND w.deleted_at IS NULL
WHERE NOT EXISTS (
  SELECT 1
  FROM workspace_feature_packages wfp
  WHERE wfp.workspace_id = w.id
    AND wfp.package_id = ufp.package_id
    AND wfp.deleted_at IS NULL
);

INSERT INTO workspace_role_bindings (
  workspace_id,
  user_id,
  role_id,
  enabled,
  created_at,
  updated_at,
  deleted_at
)
SELECT DISTINCT
  CASE
    WHEN ur.collaboration_workspace_id IS NOT NULL THEN cw.id
    ELSE pw.id
  END,
  ur.user_id,
  ur.role_id,
  TRUE,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP,
  NULL::TIMESTAMPTZ
FROM user_roles ur
LEFT JOIN workspaces cw
  ON cw.workspace_type = 'collaboration'
 AND cw.collaboration_workspace_id = ur.collaboration_workspace_id
 AND cw.deleted_at IS NULL
LEFT JOIN workspaces pw
  ON pw.workspace_type = 'personal'
 AND pw.owner_user_id = ur.user_id
 AND pw.deleted_at IS NULL
WHERE (
    (ur.collaboration_workspace_id IS NOT NULL AND cw.id IS NOT NULL)
    OR
    (ur.collaboration_workspace_id IS NULL AND pw.id IS NOT NULL)
  )
  AND NOT EXISTS (
    SELECT 1
    FROM workspace_role_bindings wrb
    WHERE wrb.workspace_id = CASE WHEN ur.collaboration_workspace_id IS NOT NULL THEN cw.id ELSE pw.id END
      AND wrb.user_id = ur.user_id
      AND wrb.role_id = ur.role_id
      AND wrb.deleted_at IS NULL
  );

INSERT INTO workspace_blocked_menus (
  app_key,
  workspace_id,
  menu_id,
  created_at,
  updated_at
)
SELECT
  cbm.app_key,
  w.id,
  cbm.menu_id,
  COALESCE(cbm.created_at, CURRENT_TIMESTAMP),
  COALESCE(cbm.updated_at, CURRENT_TIMESTAMP)
FROM collaboration_workspace_blocked_menus cbm
JOIN workspaces w
  ON w.workspace_type = 'collaboration'
 AND w.collaboration_workspace_id = cbm.collaboration_workspace_id
 AND w.deleted_at IS NULL
ON CONFLICT (app_key, workspace_id, menu_id) DO NOTHING;

INSERT INTO workspace_blocked_actions (
  app_key,
  workspace_id,
  action_id,
  created_at,
  updated_at
)
SELECT
  cba.app_key,
  w.id,
  cba.action_id,
  COALESCE(cba.created_at, CURRENT_TIMESTAMP),
  COALESCE(cba.updated_at, CURRENT_TIMESTAMP)
FROM collaboration_workspace_blocked_actions cba
JOIN workspaces w
  ON w.workspace_type = 'collaboration'
 AND w.collaboration_workspace_id = cba.collaboration_workspace_id
 AND w.deleted_at IS NULL
ON CONFLICT (app_key, workspace_id, action_id) DO NOTHING;

INSERT INTO workspace_access_snapshots (
  app_key,
  workspace_id,
  package_ids,
  expanded_package_ids,
  derived_action_ids,
  derived_action_map,
  blocked_action_ids,
  effective_action_ids,
  derived_menu_ids,
  derived_menu_map,
  blocked_menu_ids,
  effective_menu_ids,
  refreshed_at,
  created_at,
  updated_at
)
SELECT
  s.app_key,
  w.id,
  s.package_ids,
  s.expanded_package_ids,
  s.derived_action_ids,
  s.derived_action_map,
  s.blocked_action_ids,
  s.effective_action_ids,
  s.derived_menu_ids,
  s.derived_menu_map,
  s.blocked_menu_ids,
  s.effective_menu_ids,
  COALESCE(s.refreshed_at, CURRENT_TIMESTAMP),
  COALESCE(s.created_at, CURRENT_TIMESTAMP),
  COALESCE(s.updated_at, CURRENT_TIMESTAMP)
FROM collaboration_workspace_access_snapshots s
JOIN workspaces w
  ON w.workspace_type = 'collaboration'
 AND w.collaboration_workspace_id = s.collaboration_workspace_id
 AND w.deleted_at IS NULL
ON CONFLICT (app_key, workspace_id) DO UPDATE
SET
  package_ids = EXCLUDED.package_ids,
  expanded_package_ids = EXCLUDED.expanded_package_ids,
  derived_action_ids = EXCLUDED.derived_action_ids,
  derived_action_map = EXCLUDED.derived_action_map,
  blocked_action_ids = EXCLUDED.blocked_action_ids,
  effective_action_ids = EXCLUDED.effective_action_ids,
  derived_menu_ids = EXCLUDED.derived_menu_ids,
  derived_menu_map = EXCLUDED.derived_menu_map,
  blocked_menu_ids = EXCLUDED.blocked_menu_ids,
  effective_menu_ids = EXCLUDED.effective_menu_ids,
  refreshed_at = EXCLUDED.refreshed_at,
  updated_at = EXCLUDED.updated_at;

INSERT INTO workspace_role_access_snapshots (
  app_key,
  workspace_id,
  role_id,
  package_ids,
  expanded_package_ids,
  available_action_ids,
  disabled_action_ids,
  action_ids,
  action_source_map,
  available_menu_ids,
  hidden_menu_ids,
  menu_ids,
  menu_source_map,
  inherited,
  refreshed_at,
  created_at,
  updated_at
)
SELECT
  s.app_key,
  w.id,
  s.role_id,
  s.package_ids,
  s.expanded_package_ids,
  s.available_action_ids,
  s.disabled_action_ids,
  s.action_ids,
  s.action_source_map,
  s.available_menu_ids,
  s.hidden_menu_ids,
  s.menu_ids,
  s.menu_source_map,
  s.inherited,
  COALESCE(s.refreshed_at, CURRENT_TIMESTAMP),
  COALESCE(s.created_at, CURRENT_TIMESTAMP),
  COALESCE(s.updated_at, CURRENT_TIMESTAMP)
FROM collaboration_workspace_role_access_snapshots s
JOIN workspaces w
  ON w.workspace_type = 'collaboration'
 AND w.collaboration_workspace_id = s.collaboration_workspace_id
 AND w.deleted_at IS NULL
ON CONFLICT (app_key, workspace_id, role_id) DO UPDATE
SET
  package_ids = EXCLUDED.package_ids,
  expanded_package_ids = EXCLUDED.expanded_package_ids,
  available_action_ids = EXCLUDED.available_action_ids,
  disabled_action_ids = EXCLUDED.disabled_action_ids,
  action_ids = EXCLUDED.action_ids,
  action_source_map = EXCLUDED.action_source_map,
  available_menu_ids = EXCLUDED.available_menu_ids,
  hidden_menu_ids = EXCLUDED.hidden_menu_ids,
  menu_ids = EXCLUDED.menu_ids,
  menu_source_map = EXCLUDED.menu_source_map,
  inherited = EXCLUDED.inherited,
  refreshed_at = EXCLUDED.refreshed_at,
  updated_at = EXCLUDED.updated_at;

-- +goose Down
DROP TABLE IF EXISTS workspace_role_access_snapshots;
DROP TABLE IF EXISTS workspace_access_snapshots;
DROP TABLE IF EXISTS workspace_blocked_actions;
DROP TABLE IF EXISTS workspace_blocked_menus;
