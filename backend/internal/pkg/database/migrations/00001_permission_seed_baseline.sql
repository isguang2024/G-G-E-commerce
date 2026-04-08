-- +goose Up
-- +goose StatementBegin

-- Permission seed baseline. Hot-fix commit 1a2e731 inserted these 12 keys
-- and a workspace_feature_packages binding by hand against a live DB; this
-- migration freezes that state into the migration chain so a fresh
-- `make db-reset` no longer loses them.
--
-- The runtime ensure path (permissionseed.EnsureOpenAPIPermissionKeys, called
-- from cmd/migrate at startup) keeps the table in sync with openapi.yaml on
-- subsequent edits, so future spec changes do NOT require a new migration.
--
-- Both the migration and the ensure path are idempotent. The migration is
-- safe to re-apply against an already-seeded database.

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables WHERE table_name = 'permission_keys'
  ) THEN
    -- The legacy table is built by GORM AutoMigrate during cmd/migrate run.
    -- If goose runs before AutoMigrate (the new ordering), the table doesn't
    -- exist yet — skip and rely on EnsureOpenAPIPermissionKeys at startup.
    RAISE NOTICE 'permission_keys table not present yet, skipping baseline seed';
    RETURN;
  END IF;

  INSERT INTO permission_keys (
    id, code, permission_key, app_key, module_code, context_type, feature_kind,
    data_policy, allowed_workspace_types, name, description, status, sort_order,
    is_builtin, created_at, updated_at
  )
  SELECT gen_random_uuid(), substr(md5(k.permission_key), 1, 32), k.permission_key,
         'platform-admin', split_part(k.permission_key, '.', 1),
         'common', 'system', 'none', 'personal,collaboration',
         k.permission_key, k.permission_key, 'normal', 0, true, now(), now()
  FROM (VALUES
    ('workspace.read'),
    ('workspace.switch'),
    ('user.list'),
    ('user.create'),
    ('user.update'),
    ('user.delete'),
    ('user.read'),
    ('role.list'),
    ('role.create'),
    ('role.update'),
    ('role.delete'),
    ('permission.explain')
  ) AS k(permission_key)
  WHERE NOT EXISTS (
    SELECT 1 FROM permission_keys pk WHERE pk.permission_key = k.permission_key
  );
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- This baseline is not reversible; the keys may have downstream bindings.
SELECT 1;
-- +goose StatementEnd
