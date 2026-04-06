# Workspace 术语表

> 当前生效语义：`workspace` 是当前协作与权限主线；`tenant` 预留给未来真正的多租户系统。

## 核心术语

| 术语 | 当前定义 | 说明 |
| --- | --- | --- |
| `workspace` | 当前业务空间总概念 | 当前权限、导航、消息、协作管理都围绕它展开 |
| `personal workspace` | 个人空间 | 每个用户默认拥有，承载平台后台、个人侧授权和平台能力 |
| `collaboration workspace` | 协作空间 | 当前协作业务空间，承载成员、角色、功能包、消息和协作管理 |
| `tenant` | 未来多租户系统保留名词 | 不再承担当前协作空间主语义；现阶段只允许出现在历史迁移、旧表映射或归档文档里 |
| `app` | 业务域 | 如平台后台、某个业务后台或子产品域 |
| `menu-space` | APP 内导航 / 宿主空间 | 只承载导航层级、落地页和菜单宿主，不承载权限主体语义 |
| `permission_key` | 权限键 | 当前最小授权单位 |
| `feature_package` | 功能包 | `workspace` 的能力上限 |
| `role` | 角色 | `workspace` 内成员的权限裁剪器 |
| `platform role` | 平台角色 | 通过 `personal workspace` 生效 |
| `collaboration role` | 协作空间角色 | 通过 `collaboration workspace` 生效 |
| `member identity` | 成员身份 | 如协作空间管理员、协作空间成员；表示成员关系边界，不等价于权限角色 |
| `auth_workspace_id` | 当前授权空间 | 运行时判权来源 |
| `target_workspace_id` | 当前操作目标空间 | 代管其他空间时显式携带 |
| `current_collaboration_workspace_id` | 当前协作空间视图 | 协作空间页面、消息域和旧协作接口桥接使用 |

## 命名规则

1. 当前总概念统一使用 `workspace`。
2. 当前子类统一使用 `personal workspace` 与 `collaboration workspace`。
3. 未来真正多租户系统统一使用 `tenant`，不得与当前协作空间混用。
4. 当前活跃 header、DTO、返回字段、前端 store 和页面状态不得继续新增 `tenant_*` 命名来表达协作空间语义。
5. `menu-space` 不是权限空间，也不是协作空间。

## 权限规则

1. 平台后台权限由 `personal workspace` 承载。
2. 协作空间角色与协作空间功能包由 `collaboration workspace` 承载。
3. 运行时权限判断固定为：
   `资源/API/页面 -> permission_key -> feature_package -> role -> workspace`
4. `member identity` 只表示成员关系，不直接替代权限角色。
5. 管理其他空间时必须区分 `auth_workspace_id` 与 `target_workspace_id`。

## 禁止混用

- 不要把 `tenant` 当成当前协作空间或当前权限主体。
- 不要把 `menu-space` 当成权限空间或协作空间。
- 不要把账号直接当成业务权限承载体。
- 不要继续新增 `X-Tenant-ID`、`current_tenant_id`、`source_tenant_id` 这类当前业务主输出。
- 不要再使用 `workspace_type = team`；当前有效值只有 `personal | collaboration`。
