# 空间权限术语表

> 当前生效语义：`workspace` 是唯一权限主体；`tenant` 预留给未来多租户系统。

## 核心术语

| 术语 | 当前定义 | 说明 |
| --- | --- | --- |
| `workspace` | 当前业务的统一权限主体 | 权限、角色、功能包、菜单、动作都绑定在它上面 |
| `personal workspace` | 个人空间 | `workspace` 的一种类型，表示个人空间权限上下文 |
| `collaboration workspace` | 协作空间 | `workspace` 的一种类型，表示协作空间权限上下文 |
| `workspace_type` | 空间类型 | 当前有效值只有 `personal | collaboration` |
| `auth_workspace_id` | 当前授权空间 | 当前请求的权限来源 |
| `auth_workspace_type` | 当前授权空间类型 | 当前请求必须显式携带的空间类型 |
| `target_workspace_id` | 当前操作目标空间 | 跨空间操作时显式指定 |
| `current_collaboration_workspace_id` | 当前协作空间视图 | 协作空间页面和协作域视图使用，不等价于权限主体 |
| `app` | 业务域 | 例如平台后台或某个子产品域，不等于空间类型 |
| `menu-space` | 导航宿主空间 | 只表示 APP 内导航宿主，不表示权限空间 |
| `permission_key` | 最小授权单位 | 页面、菜单、动作最终都通过它判权 |
| `feature_package` | 功能包 | 某个 `workspace` 的能力上限 |
| `role` | 角色 | 某个 `workspace` 内的权限裁剪器 |
| `member identity` | 成员身份 | 表示空间成员关系边界，不等价于权限角色 |
| `tenant` | 未来多租户保留名词 | 当前阶段不得用来表示协作空间 |

## 关键规则

1. 当前系统只有一套 `workspace` 权限模型。
2. `personal` 和 `collaboration` 是空间类型，不是两套权限系统。
3. `platform` 只是业务域，不是权限上下文类型。
4. `menu-space` 只是导航宿主，不是权限空间。
5. 当前权限判断固定基于：
   `auth_workspace_id + auth_workspace_type`

## 禁止混用

- 不要把 `personal workspace` 写成“平台上下文”
- 不要把 `collaboration workspace` 写成“团队上下文”
- 不要把“平台权限 / 团队权限”当成权限模型一级分类
- 不要把 `tenant` 当成当前协作空间
- 不要新增 `tenant_*`、`team_*` 当前业务字段来表达空间语义
- 不要再使用 `workspace_type = team`
