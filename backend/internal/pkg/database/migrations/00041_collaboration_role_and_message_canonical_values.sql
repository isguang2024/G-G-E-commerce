-- +goose Up
-- collaboration 语义收口第三阶段：
-- 1. 默认协作角色码从 collaboration_workspace_* 收口到 collaboration_*
-- 2. 消息 audience / 接收组 target / 投递来源元数据同步收口
-- 3. 历史消息中的 role target 值一并回填，避免 UI 和 runtime 混用新旧值

UPDATE roles
SET
  code = CASE code
    WHEN 'collaboration_workspace_admin' THEN 'collaboration_admin'
    WHEN 'collaboration_workspace_member' THEN 'collaboration_member'
    ELSE code
  END,
  updated_at = CURRENT_TIMESTAMP
WHERE code IN ('collaboration_workspace_admin', 'collaboration_workspace_member');

UPDATE collaboration_workspace_members
SET
  role_code = CASE role_code
    WHEN 'collaboration_workspace_admin' THEN 'collaboration_admin'
    WHEN 'collaboration_workspace_member' THEN 'collaboration_member'
    ELSE role_code
  END,
  updated_at = CURRENT_TIMESTAMP
WHERE role_code IN ('collaboration_workspace_admin', 'collaboration_workspace_member');

UPDATE message_templates
SET
  audience_type = CASE audience_type
    WHEN 'collaboration_workspace_admins' THEN 'collaboration_admins'
    WHEN 'collaboration_workspace_users' THEN 'collaboration_users'
    ELSE audience_type
  END,
  updated_at = CURRENT_TIMESTAMP
WHERE audience_type IN ('collaboration_workspace_admins', 'collaboration_workspace_users');

UPDATE messages
SET
  audience_type = CASE audience_type
    WHEN 'collaboration_workspace_admins' THEN 'collaboration_admins'
    WHEN 'collaboration_workspace_users' THEN 'collaboration_users'
    ELSE audience_type
  END,
  target_role_codes = COALESCE(
    (
      SELECT jsonb_agg(
        CASE value
          WHEN 'collaboration_workspace_admin' THEN 'collaboration_admin'
          WHEN 'collaboration_workspace_member' THEN 'collaboration_member'
          ELSE value
        END
      )
      FROM jsonb_array_elements_text(COALESCE(messages.target_role_codes, '[]'::jsonb)) AS value
    ),
    '[]'::jsonb
  ),
  updated_at = CURRENT_TIMESTAMP
WHERE audience_type IN ('collaboration_workspace_admins', 'collaboration_workspace_users')
   OR EXISTS (
     SELECT 1
     FROM jsonb_array_elements_text(COALESCE(messages.target_role_codes, '[]'::jsonb)) AS value
     WHERE value IN ('collaboration_workspace_admin', 'collaboration_workspace_member')
   );

UPDATE message_recipient_group_targets
SET
  target_type = CASE target_type
    WHEN 'collaboration_workspace_admins' THEN 'collaboration_admins'
    WHEN 'collaboration_workspace_users' THEN 'collaboration_users'
    ELSE target_type
  END,
  role_code = CASE role_code
    WHEN 'collaboration_workspace_admin' THEN 'collaboration_admin'
    WHEN 'collaboration_workspace_member' THEN 'collaboration_member'
    ELSE role_code
  END,
  updated_at = CURRENT_TIMESTAMP
WHERE target_type IN ('collaboration_workspace_admins', 'collaboration_workspace_users')
   OR role_code IN ('collaboration_workspace_admin', 'collaboration_workspace_member');

UPDATE message_deliveries
SET
  meta = jsonb_strip_nulls(
    jsonb_set(
      jsonb_set(
        jsonb_set(
          jsonb_set(
            COALESCE(meta, '{}'::jsonb),
            '{source_rule_type}',
            to_jsonb(
              CASE COALESCE(meta ->> 'source_rule_type', '')
                WHEN 'collaboration_workspace_admins' THEN 'collaboration_admins'
                WHEN 'collaboration_workspace_users' THEN 'collaboration_users'
                ELSE COALESCE(meta ->> 'source_rule_type', '')
              END
            ),
            true
          ),
          '{source_target_type}',
          to_jsonb(
            CASE COALESCE(meta ->> 'source_target_type', '')
              WHEN 'collaboration_workspace_admins' THEN 'collaboration_admins'
              WHEN 'collaboration_workspace_users' THEN 'collaboration_users'
              ELSE COALESCE(meta ->> 'source_target_type', '')
            END
          ),
          true
        ),
        '{source_target_value}',
        to_jsonb(
          CASE COALESCE(meta ->> 'source_target_value', '')
            WHEN 'collaboration_workspace_admin' THEN 'collaboration_admin'
            WHEN 'collaboration_workspace_member' THEN 'collaboration_member'
            ELSE COALESCE(meta ->> 'source_target_value', '')
          END
        ),
        true
      ),
      '{source_rule_label}',
      to_jsonb(
        replace(
          replace(
            COALESCE(meta ->> 'source_rule_label', ''),
            'collaboration_workspace_admin',
            'collaboration_admin'
          ),
          'collaboration_workspace_member',
          'collaboration_member'
        )
      ),
      true
    )
  ),
  updated_at = CURRENT_TIMESTAMP
WHERE COALESCE(meta ->> 'source_rule_type', '') IN ('collaboration_workspace_admins', 'collaboration_workspace_users')
   OR COALESCE(meta ->> 'source_target_type', '') IN ('collaboration_workspace_admins', 'collaboration_workspace_users')
   OR COALESCE(meta ->> 'source_target_value', '') IN ('collaboration_workspace_admin', 'collaboration_workspace_member')
   OR COALESCE(meta ->> 'source_rule_label', '') LIKE 'collaboration_workspace_%';

-- +goose Down
