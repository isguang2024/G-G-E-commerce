-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.dict_items') IS NOT NULL THEN
        ALTER TABLE dict_items
            ADD COLUMN IF NOT EXISTS description varchar(500) NOT NULL DEFAULT '';

        UPDATE dict_types
           SET description = v.description,
               updated_at = NOW()
          FROM (VALUES
            ('common_status', '通用启用/停用状态'),
            ('gender', '性别选项'),
            ('page_type', '页面类型枚举'),
            ('access_mode', '页面/接口访问模式'),
            ('page_source', '页面创建来源'),
            ('http_method', 'HTTP 请求方法'),
            ('message_type', '消息分类'),
            ('workspace_plan', '工作空间套餐等级'),
            ('register_source', '用户注册来源标识')
          ) AS v(code, description)
         WHERE dict_types.tenant_id = 'default'
           AND dict_types.code = v.code
           AND COALESCE(dict_types.description, '') <> v.description;

        UPDATE dict_items
           SET description = v.description,
               updated_at = NOW()
          FROM dict_types dt,
               (VALUES
                ('common_status', 'normal', '默认可用状态，参与正常展示与选择。'),
                ('common_status', 'suspended', '禁用状态，通常用于下线或临时关闭。'),
                ('gender', 'male', '男性用户标识。'),
                ('gender', 'female', '女性用户标识。'),
                ('page_type', 'group', '用于组织页面层级的目录型节点。'),
                ('page_type', 'display_group', '仅用于前端展示分组，不直接承载页面。'),
                ('page_type', 'inner', '挂载在菜单体系中的业务页面。'),
                ('page_type', 'standalone', '不依赖菜单层级的独立访问页面。'),
                ('access_mode', 'inherit', '继承父级或上游配置决定访问控制。'),
                ('access_mode', 'public', '无需登录即可访问。'),
                ('access_mode', 'jwt', '要求已登录并持有有效 JWT。'),
                ('access_mode', 'permission', '要求通过权限点校验后访问。'),
                ('page_source', 'manual', '由管理员手工创建或维护。'),
                ('page_source', 'sync', '由系统同步流程自动写入。'),
                ('page_source', 'seed', '来自默认 seed 或初始化脚本。'),
                ('page_source', 'remote', '来自远端系统或注册中心。'),
                ('http_method', 'GET', '读取资源，不应产生副作用。'),
                ('http_method', 'POST', '创建资源或提交动作请求。'),
                ('http_method', 'PUT', '整体更新资源。'),
                ('http_method', 'PATCH', '局部更新资源。'),
                ('http_method', 'DELETE', '删除资源。'),
                ('message_type', 'notice', '广播型通知，强调告知。'),
                ('message_type', 'message', '普通消息沟通记录。'),
                ('message_type', 'todo', '需要用户处理的动作项。'),
                ('workspace_plan', 'free', '默认基础套餐，适合轻量使用场景。'),
                ('workspace_plan', 'pro', '提供更多协作与管理能力的进阶套餐。'),
                ('workspace_plan', 'enterprise', '面向组织治理与定制化能力的企业套餐。'),
                ('register_source', 'self', '用户通过公开注册页自主完成注册。'),
                ('register_source', 'invite', '用户通过邀请码或邀请链路进入注册。'),
                ('register_source', 'admin', '由后台管理员直接创建用户。'),
                ('register_source', 'email', '通过邮箱验证流程完成注册。'),
                ('register_source', 'sms', '通过短信验证码流程完成注册。'),
                ('register_source', 'oauth', '通过第三方身份提供商完成注册。')
               ) AS v(type_code, item_value, description)
         WHERE dict_items.tenant_id = 'default'
           AND dt.id = dict_items.dict_type_id
           AND dt.tenant_id = dict_items.tenant_id
           AND dt.code = v.type_code
           AND dict_items.value = v.item_value
           AND COALESCE(dict_items.description, '') <> v.description;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.dict_items') IS NOT NULL THEN
        ALTER TABLE dict_items
            DROP COLUMN IF EXISTS description;
    END IF;
END $$;
-- +goose StatementEnd
