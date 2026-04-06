# Workspace 术语表

> 该术语表用于统一迁移期间的概念边界，避免 `tenant`、`workspace`、`menu-space`、`app` 混用。

## 术语

| 术语 | 当前定义 | 迁移备注 |
| --- | --- | --- |
| `workspace` | 业务权限主体 | 后续唯一承载权限的空间概念 |
| `personal workspace` | 个人空间 | 每个用户默认拥有，承载平台后台等个人授权能力 |
| `team workspace` | 团队空间 | 由旧 `tenant` 逐步映射而来 |
| `tenant` | 预留给未来多租户系统的命名空间 | 当前不再承担 team / workspace 主语义；现阶段只保留在 legacy 表、legacy 路由和兼容映射中 |
| `app` | 业务域 | 例如平台后台、业务后台、某个子产品域 |
| `menu-space` | APP 内导航 / 宿主空间 | 只管菜单与导航结构，不承担权限主体职责 |
| `permission_key` | 权限键 | 授权单位，资源层最终会映射到它 |
| `feature_package` | 功能包 | workspace 的能力上限 |
| `role` | 角色 | workspace 内成员裁剪器，不是账号级全局业务角色 |
| `platform role` | 平台角色 | 平台后台配置的角色目录，通过 `personal workspace` 生效 |
| `team internal role` | 团队内部角色 | 团队工作空间内的角色目录或成员绑定，通过 `team workspace` 生效 |
| `member identity` | 成员身份 | 例如 `tenant_members.role_code`、`workspace_members.member_type`，表示成员关系边界，不等于权限角色 |
| `auth_workspace_id` | 当前授权来源空间 | 请求运行时的主授权上下文 |
| `target_workspace_id` | 当前操作目标空间 | 管理别的空间时显式携带 |
| `permission.data_policy` | 数据域策略 | 决定是否按 workspace 数据域约束 |

## 概念边界

- `workspace` 和 `menu-space` 不是同一个概念。
- `app` 是业务域，不是权限主体。
- `tenant` 已从当前 team/workspace 主线脱钩，预留给未来真正的多租户系统。
- `role` 只在某个 workspace 中生效。
- `platform role` 和 `team internal role` 可以同时存在于同一用户身上，但各自只在对应 workspace 生效。
- `member identity` 不直接替代权限角色。
- `feature_package` 是 workspace 能力上限，不是账号能力上限。

## 关键规则

1. 平台后台 APP 的权限也由 `personal workspace` 承载。
2. 管理其他团队时必须显式使用 `target_workspace_id`。
3. 权限判断以 `permission_key` 为核心，不以菜单可见性代替授权。
4. 运行时应显式区分 `auth_workspace_id` 和 `target_workspace_id`。
5. 平台角色由 `personal workspace` 绑定，团队内部角色由 `team workspace` 绑定。

## 禁止混用

- 不要把 `tenant` 当成最终业务权限主体。
- 不要继续新增 `tenant` 命名的 header、上下文键、页面状态或主输出字段来表达当前团队语义。
- 不要把 `menu-space` 当成权限空间。
- 不要把账号直接当成业务权限承载体。
- 不要把平台后台权限做成账号级全局业务角色。

## 迁移读法

阅读现有代码或设计时，建议按以下顺序映射：

1. 先判断它属于 `app` 还是 `menu-space`。
2. 再判断它是 `auth_workspace_id` 语义还是 `target_workspace_id` 语义。
3. 最后再落到 `permission_key`、`feature_package`、`role`。
