-- +goose Up
-- workspace/collaboration 单主域收口的第一阶段：
-- 1. 角色定义进入 roles + role_scopes
-- 2. message/template/sender/group 统一补齐 scope_type/scope_id
-- 3. 旧 collaboration_workspace_* 列先保留为兼容镜像，待运行时与前端切完再回收

ALTER TABLE roles ADD COLUMN IF NOT EXISTS scope_type VARCHAR(20) NOT NULL DEFAULT 'global';
ALTER TABLE roles ADD COLUMN IF NOT EXISTS scope_id UUID NULL;

CREATE TABLE IF NOT EXISTS role_scopes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  scope_type VARCHAR(20) NOT NULL DEFAULT 'global',
  scope_id UUID NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_role_scopes_role_unique
  ON role_scopes (role_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_role_scopes_scope_lookup
  ON role_scopes (scope_type, scope_id)
  WHERE deleted_at IS NULL;

INSERT INTO workspaces (
  workspace_type,
  name,
  code,
  owner_user_id,
  collaboration_workspace_id,
  status,
  meta,
  created_at,
  updated_at
)
SELECT
  'collaboration',
  COALESCE(NULLIF(BTRIM(cw.name), ''), 'Collaboration Workspace'),
  'collaboration-' || LOWER(REPLACE(cw.id::text, '-', '')),
  cw.owner_id,
  cw.id,
  CASE
    WHEN LOWER(COALESCE(cw.status, 'active')) IN ('disabled', 'inactive', 'suspended') THEN 'disabled'
    ELSE 'active'
  END,
  jsonb_build_object(
    'legacy_source', 'migration',
    'legacy_collaboration_workspace_id', cw.id::text
  ),
  COALESCE(cw.created_at, CURRENT_TIMESTAMP),
  COALESCE(cw.updated_at, CURRENT_TIMESTAMP)
FROM collaboration_workspaces cw
WHERE NOT EXISTS (
  SELECT 1
  FROM workspaces w
  WHERE w.workspace_type = 'collaboration'
    AND w.collaboration_workspace_id = cw.id
    AND w.deleted_at IS NULL
);

UPDATE roles
SET
  scope_type = CASE
    WHEN collaboration_workspace_id IS NOT NULL THEN 'collaboration'
    ELSE 'global'
  END,
  scope_id = CASE
    WHEN collaboration_workspace_id IS NOT NULL THEN (
      SELECT w.id
      FROM workspaces w
      WHERE w.workspace_type = 'collaboration'
        AND w.collaboration_workspace_id = roles.collaboration_workspace_id
        AND w.deleted_at IS NULL
      LIMIT 1
    )
    ELSE NULL
  END;

INSERT INTO role_scopes (role_id, scope_type, scope_id, created_at, updated_at)
SELECT
  r.id,
  r.scope_type,
  r.scope_id,
  COALESCE(r.created_at, CURRENT_TIMESTAMP),
  COALESCE(r.updated_at, CURRENT_TIMESTAMP)
FROM roles r
WHERE NOT EXISTS (
  SELECT 1
  FROM role_scopes rs
  WHERE rs.role_id = r.id
    AND rs.deleted_at IS NULL
);

ALTER TABLE message_templates ADD COLUMN IF NOT EXISTS owner_scope_id UUID NULL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS target_scope_type VARCHAR(20) NOT NULL DEFAULT 'global';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS target_scope_id UUID NULL;
ALTER TABLE message_deliveries ADD COLUMN IF NOT EXISTS recipient_scope_type VARCHAR(20) NOT NULL DEFAULT 'global';
ALTER TABLE message_deliveries ADD COLUMN IF NOT EXISTS recipient_scope_id UUID NULL;
ALTER TABLE message_recipient_group_targets ADD COLUMN IF NOT EXISTS target_scope_type VARCHAR(20) NOT NULL DEFAULT 'global';
ALTER TABLE message_recipient_group_targets ADD COLUMN IF NOT EXISTS target_scope_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_roles_scope_lookup
  ON roles (scope_type, scope_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_messages_target_scope_lookup
  ON messages (target_scope_type, target_scope_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_message_deliveries_recipient_scope_lookup
  ON message_deliveries (recipient_scope_type, recipient_scope_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_message_templates_owner_scope_lookup
  ON message_templates (owner_scope, owner_scope_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_message_recipient_group_targets_scope_lookup
  ON message_recipient_group_targets (target_scope_type, target_scope_id)
  WHERE deleted_at IS NULL;

UPDATE message_templates
SET owner_scope = 'global'
WHERE owner_scope = 'personal';

UPDATE message_senders
SET scope_type = 'global', scope_id = NULL
WHERE scope_type = 'personal' AND scope_id IS NULL;

UPDATE message_recipient_groups
SET scope_type = 'global', scope_id = NULL
WHERE scope_type = 'personal' AND scope_id IS NULL;

UPDATE messages
SET
  scope_type = 'global',
  scope_id = NULL,
  audience_scope = 'global'
WHERE scope_type = 'personal' AND scope_id IS NULL;

UPDATE message_templates
SET owner_scope_id = NULL
WHERE owner_scope = 'global';

UPDATE message_templates mt
SET owner_scope_id = w.id
FROM workspaces w
WHERE mt.owner_scope = 'collaboration'
  AND mt.owner_collaboration_workspace_id IS NOT NULL
  AND w.workspace_type = 'collaboration'
  AND w.collaboration_workspace_id = mt.owner_collaboration_workspace_id
  AND w.deleted_at IS NULL;

UPDATE messages
SET
  target_scope_type = CASE
    WHEN target_collaboration_workspace_id IS NOT NULL THEN 'collaboration'
    WHEN scope_type IN ('global', 'personal', 'collaboration') THEN scope_type
    ELSE 'global'
  END,
  target_scope_id = CASE
    WHEN scope_type = 'collaboration' THEN scope_id
    ELSE NULL
  END;

UPDATE messages m
SET target_scope_id = w.id
FROM workspaces w
WHERE m.target_collaboration_workspace_id IS NOT NULL
  AND w.workspace_type = 'collaboration'
  AND w.collaboration_workspace_id = m.target_collaboration_workspace_id
  AND w.deleted_at IS NULL;

UPDATE message_deliveries md
SET
  recipient_scope_type = CASE
    WHEN md.recipient_collaboration_workspace_id IS NOT NULL THEN 'collaboration'
    WHEN m.scope_type IN ('global', 'personal', 'collaboration') THEN m.scope_type
    ELSE 'global'
  END,
  recipient_scope_id = CASE
    WHEN m.scope_type = 'collaboration' THEN m.scope_id
    ELSE NULL
  END
FROM messages m
WHERE m.id = md.message_id;

UPDATE message_deliveries md
SET recipient_scope_id = w.id
FROM workspaces w
WHERE md.recipient_collaboration_workspace_id IS NOT NULL
  AND w.workspace_type = 'collaboration'
  AND w.collaboration_workspace_id = md.recipient_collaboration_workspace_id
  AND w.deleted_at IS NULL;

UPDATE message_recipient_group_targets t
SET
  target_scope_type = CASE
    WHEN t.collaboration_workspace_id IS NOT NULL THEN 'collaboration'
    WHEN g.scope_type IN ('global', 'personal', 'collaboration') THEN g.scope_type
    ELSE 'global'
  END,
  target_scope_id = CASE
    WHEN t.collaboration_workspace_id IS NOT NULL THEN (
      SELECT w.id
      FROM workspaces w
      WHERE w.workspace_type = 'collaboration'
        AND w.collaboration_workspace_id = t.collaboration_workspace_id
        AND w.deleted_at IS NULL
      LIMIT 1
    )
    ELSE g.scope_id
  END
FROM message_recipient_groups g
WHERE g.id = t.group_id;

-- +goose Down
DROP INDEX IF EXISTS idx_message_recipient_group_targets_scope_lookup;
DROP INDEX IF EXISTS idx_message_templates_owner_scope_lookup;
DROP INDEX IF EXISTS idx_message_deliveries_recipient_scope_lookup;
DROP INDEX IF EXISTS idx_messages_target_scope_lookup;
DROP INDEX IF EXISTS idx_roles_scope_lookup;
DROP INDEX IF EXISTS idx_role_scopes_scope_lookup;
DROP INDEX IF EXISTS idx_role_scopes_role_unique;

ALTER TABLE message_recipient_group_targets DROP COLUMN IF EXISTS target_scope_id;
ALTER TABLE message_recipient_group_targets DROP COLUMN IF EXISTS target_scope_type;
ALTER TABLE message_deliveries DROP COLUMN IF EXISTS recipient_scope_id;
ALTER TABLE message_deliveries DROP COLUMN IF EXISTS recipient_scope_type;
ALTER TABLE messages DROP COLUMN IF EXISTS target_scope_id;
ALTER TABLE messages DROP COLUMN IF EXISTS target_scope_type;
ALTER TABLE message_templates DROP COLUMN IF EXISTS owner_scope_id;

DROP TABLE IF EXISTS role_scopes;
ALTER TABLE roles DROP COLUMN IF EXISTS scope_id;
ALTER TABLE roles DROP COLUMN IF EXISTS scope_type;
