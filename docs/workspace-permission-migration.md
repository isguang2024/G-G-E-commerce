# Workspace 权限迁移说明

> 最终状态文档，覆盖 `codex_workspace_permission_tasks.txt` 的 Phase 0-10 收口结果。

## 当前结论

- 当前业务权限主线已经统一到 `workspace`。
- `workspace_type` 当前有效值只有：
  - `personal`
  - `collaboration`
- 中文统一为：
  - `个人空间`
  - `协作空间`
- `tenant` 已从当前协作、权限、导航、消息主线退出，保留给未来真正的多租户系统。

## 已完成范围

### 1. 数据库与迁移

- 协作空间相关核心表已物理切换到 collaboration 命名。
- 迁移链已在 `AutoMigrate` 前增加预重命名步骤，避免旧表和新表并存。
- 当前活跃列与索引已按协作空间语义重命名。
- 历史值已完成回填与规范化，包括：
  - `workspace_type: team -> collaboration`
  - 权限、消息、模板、接收组等领域中的 `team/tenant` 活跃枚举值回填到 `collaboration`

### 2. 后端运行时

- 当前授权主线固定为：
  - `X-Auth-Workspace-Id`
  - `X-Collaboration-Workspace-Id`
- 当前业务 handler、权限判断、消息、导航、页面运行时都以：
  - `auth_workspace_id`
  - `target_workspace_id`
  - `current_collaboration_workspace_id`
  为主语义。
- 平台能力通过 `personal workspace` 生效。
- 协作空间角色、成员、功能包、菜单、动作权限通过 `collaboration workspace` 生效。
- 平台代管协作空间时，path/body 中的事实目标会先解析为目标 `workspace`，再做精确校验。

### 3. 权限与角色

- 当前权限公式固定为：
  `资源/API/页面 -> permission_key -> feature_package -> role -> workspace`
- 平台角色通过 `personal workspace` 绑定并生效。
- 协作空间角色通过 `collaboration workspace` 绑定并生效。
- 成员身份与权限角色已拆开：
  - 成员身份只表示成员关系边界
  - 角色才表示权限裁剪

### 4. 前端主线

- 当前前端唯一授权上下文源是 `workspaceStore`。
- 前端公共层、消息域、系统管理页、协作空间页和工作台页已统一使用：
  - `工作空间`
  - `协作空间`
  - `个人空间`
- 活跃请求头、DTO、页面状态和消息域输出已经切到 collaboration 命名。
- 旧 `tenant` 语义不再作为前端公共层主上下文。

## 最终主契约

### 请求头

- `X-Auth-Workspace-Id`
- `X-Collaboration-Workspace-Id`

### 核心字段

- `workspace_type = personal | collaboration`
- `current_collaboration_workspace_id`
- `target_collaboration_workspace_id`
- `target_collaboration_workspace_ids`
- `owner_collaboration_workspace_id`
- `collaboration_workspace_id`
- `collaboration_workspace_member_id`

### 主接口

- `/api/v1/workspaces/*`
- `/api/v1/collaboration-workspaces/*`
- `/api/v1/collaboration-workspaces/current/*`

## 兼容边界

以下对象仍可能存在，但不再表示当前业务主语义：

- 旧 persistence 中的历史 `tenant` 表/字段名，只用于迁移和历史映射
- 少量 legacy module / package path 中仍叫 `tenant` 的代码目录
- 少量前端兼容桥接代码

当前允许保留的兼容职责仅包括：

1. 历史迁移时识别旧 tenant 数据。
2. 对旧协作接口做最小输入桥接。
3. 在迁移链和种子链中完成旧值到 collaboration 命名的回填。

以下对象不再允许作为主输出继续扩散：

- `X-Tenant-ID`
- `X-Team-Workspace-Id`
- `current_tenant_id`
- `source_tenant_id`
- `/api/v1/tenants/*`
- `workspace_type = team`

## Phase 9 结果

- 后端编译检查通过。
- 前端 `lint` 通过。
- 前端 `build` 通过。
- 关键行为通过性检查通过：
  - `personal workspace` 平台角色 / 平台功能包
  - `collaboration workspace` 协作空间角色 / 协作空间功能包
  - workspace 切换
  - menu-space 切换
  - 消息域协作空间视图

## Phase 10 结果

- 术语表、迁移说明、阶段记录与变更日志已收口到最终语义。
- `tenant` 已明确预留给未来多租户系统。
- 当前协作、权限、导航、消息和协作管理全部统一使用 `workspace / collaboration workspace / personal workspace` 解释。

## 后续规则

1. 未来真正多租户设计必须新开 `tenant` 领域方案，不复用当前协作空间语义。
2. 当前业务新增字段、header、DTO、store、route 不得再使用 `tenant` 表达协作空间。
3. 若继续做兼容层清理，应优先清理目录命名、历史模块路径和只读桥接，不要重新引入双轨主输出。
