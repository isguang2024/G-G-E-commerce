-- +goose Up
-- +goose StatementBegin
-- Register-entry inline config: 将 policy 的运行时字段内联到 entry，
-- 使 entry 成为运行时唯一真相源。policy 降级为模板/预设。

-- 1. 新增注册后去向、验证码配置、系统保留标记等字段
ALTER TABLE register_entries
    ADD COLUMN IF NOT EXISTS target_url                   varchar(1024) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS target_app_key               varchar(64)   NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS target_navigation_space_key  varchar(64)   NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS target_home_path             varchar(256)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS captcha_provider             varchar(32)   NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS captcha_site_key             varchar(256)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS welcome_message_template_key varchar(128)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS description                  text          NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS is_system_reserved           boolean       NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS role_codes                   jsonb         NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS feature_package_keys         jsonb         NOT NULL DEFAULT '[]';

-- 2. 从 policy 回填去向/验证码/描述字段到 entry
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'register_entries'
          AND column_name = 'policy_code'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'register_policies'
    ) THEN
        UPDATE register_entries e
        SET
            target_app_key               = COALESCE(p.target_app_key, ''),
            target_navigation_space_key  = COALESCE(p.target_navigation_space_key, ''),
            target_home_path             = COALESCE(p.target_home_path, ''),
            captcha_provider             = COALESCE(p.captcha_provider, 'none'),
            captcha_site_key             = COALESCE(p.captcha_site_key, ''),
            welcome_message_template_key = COALESCE(p.welcome_message_template_key, ''),
            description                  = COALESCE(p.description, '')
        FROM register_policies p
        WHERE e.policy_code = p.policy_code
          AND e.deleted_at IS NULL;
    END IF;
END $$;

-- 3. 从 policy 子表回填 role_codes（展开为 code 字符串数组）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'register_entries'
          AND column_name = 'policy_code'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'register_policy_roles'
    ) THEN
        UPDATE register_entries e
        SET role_codes = sub.codes
        FROM (
            SELECT e2.id,
                   COALESCE(jsonb_agg(r.code ORDER BY pr.sort_order), '[]'::jsonb) AS codes
            FROM register_entries e2
            JOIN register_policy_roles pr ON pr.policy_code = e2.policy_code
            JOIN roles r ON r.id = pr.role_id
            WHERE e2.deleted_at IS NULL
            GROUP BY e2.id
        ) sub
        WHERE e.id = sub.id;
    END IF;
END $$;

-- 4. 从 policy 子表回填 feature_package_keys（展开为 key 字符串数组）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'register_entries'
          AND column_name = 'policy_code'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'register_policy_feature_packages'
    ) THEN
        UPDATE register_entries e
        SET feature_package_keys = sub.keys
        FROM (
            SELECT e2.id,
                   COALESCE(jsonb_agg(fp.package_key ORDER BY pf.sort_order), '[]'::jsonb) AS keys
            FROM register_entries e2
            JOIN register_policy_feature_packages pf ON pf.policy_code = e2.policy_code
            JOIN feature_packages fp ON fp.id = pf.package_id
            WHERE e2.deleted_at IS NULL
            GROUP BY e2.id
        ) sub
        WHERE e.id = sub.id;
    END IF;
END $$;

-- 5. 回填 *bool 空值：从 policy 填充到 entry（为后续 NOT NULL 做准备）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'register_entries'
          AND column_name = 'policy_code'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'register_policies'
    ) THEN
        UPDATE register_entries e
        SET
            allow_public_register = COALESCE(e.allow_public_register, p.allow_public_register, FALSE),
            require_invite        = COALESCE(e.require_invite, p.require_invite, FALSE),
            require_email_verify  = COALESCE(e.require_email_verify, p.require_email_verify, FALSE),
            require_captcha       = COALESCE(e.require_captcha, p.require_captcha, FALSE),
            auto_login            = COALESCE(e.auto_login, p.auto_login, TRUE)
        FROM register_policies p
        WHERE e.policy_code = p.policy_code
          AND e.deleted_at IS NULL;
    END IF;
END $$;

-- 兜底：无关联 policy 的 entry 也补默认值
UPDATE register_entries
SET
    allow_public_register = COALESCE(allow_public_register, FALSE),
    require_invite        = COALESCE(require_invite, FALSE),
    require_email_verify  = COALESCE(require_email_verify, FALSE),
    require_captcha       = COALESCE(require_captcha, FALSE),
    auto_login            = COALESCE(auto_login, TRUE)
WHERE deleted_at IS NULL;

-- 6. boolean 字段加 NOT NULL + DEFAULT（entry 成为运行时唯一真相后不再允许 NULL）
ALTER TABLE register_entries
    ALTER COLUMN allow_public_register SET NOT NULL,
    ALTER COLUMN allow_public_register SET DEFAULT FALSE,
    ALTER COLUMN require_invite SET NOT NULL,
    ALTER COLUMN require_invite SET DEFAULT FALSE,
    ALTER COLUMN require_email_verify SET NOT NULL,
    ALTER COLUMN require_email_verify SET DEFAULT FALSE,
    ALTER COLUMN require_captcha SET NOT NULL,
    ALTER COLUMN require_captcha SET DEFAULT FALSE,
    ALTER COLUMN auto_login SET NOT NULL,
    ALTER COLUMN auto_login SET DEFAULT TRUE;

-- 7. policy_code 改为可空（仅保留为审计溯源字段）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'register_entries'
          AND column_name = 'policy_code'
    ) THEN
        ALTER TABLE register_entries ALTER COLUMN policy_code DROP NOT NULL;
    END IF;
END $$;

-- 8. 标记默认入口为系统保留
UPDATE register_entries
SET is_system_reserved = TRUE
WHERE entry_code = 'default';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 回退 boolean 为可空
ALTER TABLE register_entries
    ALTER COLUMN allow_public_register DROP NOT NULL,
    ALTER COLUMN allow_public_register DROP DEFAULT,
    ALTER COLUMN require_invite DROP NOT NULL,
    ALTER COLUMN require_invite DROP DEFAULT,
    ALTER COLUMN require_email_verify DROP NOT NULL,
    ALTER COLUMN require_email_verify DROP DEFAULT,
    ALTER COLUMN require_captcha DROP NOT NULL,
    ALTER COLUMN require_captcha DROP DEFAULT,
    ALTER COLUMN auto_login DROP NOT NULL,
    ALTER COLUMN auto_login DROP DEFAULT;

-- 恢复 policy_code NOT NULL
ALTER TABLE register_entries ALTER COLUMN policy_code SET NOT NULL;

-- 删除新增字段
ALTER TABLE register_entries
    DROP COLUMN IF EXISTS target_url,
    DROP COLUMN IF EXISTS target_app_key,
    DROP COLUMN IF EXISTS target_navigation_space_key,
    DROP COLUMN IF EXISTS target_home_path,
    DROP COLUMN IF EXISTS captcha_provider,
    DROP COLUMN IF EXISTS captcha_site_key,
    DROP COLUMN IF EXISTS welcome_message_template_key,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS is_system_reserved,
    DROP COLUMN IF EXISTS role_codes,
    DROP COLUMN IF EXISTS feature_package_keys;
-- +goose StatementEnd
