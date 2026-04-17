# Workspace 单主域收口任务书

## 1. 文档目标

本文作为 `tsk_01KPCPN9J96340BCPWN7B7`(COLLAB-WS-CLEANUP)的前置真相源，固定"workspace 是唯一主概念"的边界，用于约束后续 S2–S7 执行阶段不再回流 collaboration_workspace 的独立主域。

- **目标**：明确 workspace 单主域 + 两类 type 的概念模型、落地边界、鉴权分流规则、禁止双轨项清单。
- **目标**：清理当前 header / JWT / OpenAPI / store / model / SQL 多层并行存在的双轨，使"协作空间"只作为 `workspace.workspace_type = 'collaboration'` 的**类型值与中文展示**存在，不再作为独立聚合根。
- **非目标**：本文不规定 Vben 新前端(`frontend-platform/`)的改造；Vben 侧改造在本轮暂停。
- **非目标**：本文不引入新的表/新的字段；仅规定删除与收敛。

## 2. 现状盘点(双轨泄漏证据)

### 2.1 模型层并行

- `workspaces` 表已是单主域，`workspace_type ∈ {personal, collaboration}` 作为类型区分列存在于 `backend/internal/modules/system/models/workspace.go` L11-12、L24。
- 同时 `collaboration_workspaces` / `collaboration_workspace_members` 仍以独立聚合根形式存在于 `backend/internal/modules/system/models/model.go` L579-595、L597-613；这是半迁移残留。
- `workspaces.collaboration_workspace_id` 作为指向老聚合根的**回指外键**仍保留(`workspace.go` L28, L46),索引也未清理。
- `roles`、`user_roles`、`user_action_permissions` 三张表各自挂了 `collaboration_workspace_id` bridge 字段(`model.go` L101、L216、L414),权限评估 SQL 依赖这条桥。
- 额外 4 个 legacy snapshot 模型残存(`model.go` L315-329、L386-408、L484-504、L506-528)。

### 2.2 鉴权链路双轨

- `backend/internal/modules/system/auth/middleware.go` L17 声明 `collaborationWorkspaceHeader` 常量;L85-97 使用 3-tier fallback(header → JWT claim → query)同时解析 `X-Auth-Workspace-Id` 与 `X-Collaboration-Workspace-Id`。
- `backend/internal/pkg/jwt/jwt.go` L17-22 Claims 同时持有 workspace_id 与 CollaborationWorkspaceID,登录处(auth/service.go L77-83 L80)硬编码空串填充。
- Context 同时写 `auth_workspace_id` 与 `collaboration_workspace_id`,下游自行选择读哪个。

### 2.3 权限评估器 SQL 桥

- `backend/internal/pkg/permission/evaluator/evaluator.go` L205-240 `queryRoleKeys` 中存在 `ur.collaboration_workspace_id = ws.collaboration_workspace_id` 的 legacy JOIN,绕过 `workspaces.id` 这条新主键,依赖 bridge 外键桥接。
- `queryRoleKeysBySource` L252-261 同款。

### 2.4 OpenAPI 双路径

- `backend/api/openapi/domains/workspace/paths.yaml` 内 `/workspaces/*` 与 `/collaboration-workspaces/*` 两族共存,约 30 条路径重复。
- 独立模块 `backend/internal/modules/system/collaborationworkspace/` 暴露 8 方法接口(List/Get/Create/Update/Delete/ListMembers/AddMember/RemoveMember),与 `workspace/` 模块业务重叠。
- Seed(`permissionseed/openapi_seed.json`)因此同时收录两套权限 key。

### 2.5 前端双 store

- `frontend/src/store/modules/collaboration-workspace.ts` 以独立 store 存在,与 `workspace` store 并行。
- 请求层注入 `X-Collaboration-Workspace-Id` header,与 `X-Auth-Workspace-Id` 同发。
- `frontend/src/api/workspace.ts` 同时调用 `/workspaces/*` 与 `/collaboration-workspaces/*`。

## 3. 核心概念模型(最终态)

### 3.1 主域与类型

**主域唯一**:全系统只有一个聚合根 `Workspace`,对应单表 `workspaces`。

**类型两类**:

| `workspace_type` | 中文展示 | 拥有者模型 | 成员模型 | 角色来源 |
|---|---|---|---|---|
| `personal` | 个人空间 | `owner_user_id` 唯一 | 无成员表 | 固定 owner 角色 + 个人默认 feature package |
| `collaboration` | 协作空间 | `owner_user_id`(创建者) | `workspace_members`(owner/admin/member/viewer) | `workspace_role_bindings` 绑定的角色列表 |

**中文展示仅在 UI 层**:`frontend/` 下的展示层按 `workspace_type` 翻译为"个人空间/协作空间";后端 payload、DB 字段、URL、header、JWT claim 一律用英文枚举值 `personal` / `collaboration`,不产生第二套命名。

### 3.2 不再存在的概念

下列命名不应再出现在任何代码路径中(历史 migration 文件内保留原样):

- 任何以 `Collaboration` 为前缀的 Go 类型:`CollaborationWorkspace`、`CollaborationWorkspaceMember`、`GetCollaborationWorkspaceByCollaborationWorkspaceID`。
- 任何 `collaboration_workspace_id` 列(除了 drop-column migration 本身)。
- 任何 `X-Collaboration-Workspace-Id` header。
- 任何 `/collaboration-workspaces/*` URL。
- 任何 `collaborationworkspace/` Go 包或 `collaboration-workspace.ts` 前端文件。
- JWT Claims 内 `CollaborationWorkspaceID` 字段(解析旧 token 时静默忽略,不再写入)。

## 4. 鉴权分流规则

workspace 对外单入口,但 evaluator 内部按 `workspace.workspace_type` 做两分支分流;super_admin 全局 bypass 先行。

### 4.1 Middleware 契约

`backend/internal/modules/system/auth/middleware.go`:

- 只解析 `X-Auth-Workspace-Id` 一个 header(不再 fallback 到旧 header/query)。
- 查询 `workspaces` 表获得 workspace 记录后,向 context 写入**三键**:
  - `auth_workspace_id`:UUID
  - `auth_workspace_type`:`personal` / `collaboration`
  - `auth_workspace_owner_user_id`:UUID
- 不再写 `collaboration_workspace_id`。

### 4.2 Evaluator 分流

`backend/internal/pkg/permission/evaluator/evaluator.go` `queryRoleKeys`:

```text
if super_admin(user):
    bypass → 全量权限
elif workspace.workspace_type == 'personal':
    if workspace.owner_user_id != user.id:
        → 403
    else:
        role_keys = {owner_role}
        feature_keys = personal_default_package.feature_keys
elif workspace.workspace_type == 'collaboration':
    member = query workspace_members by (workspace_id, user_id)
    if member is None:
        → 403
    else:
        role_keys = query workspace_role_bindings by (workspace_id, user_id)
        if member.member_type == 'owner':
            role_keys += {owner_role}
        feature_keys = merged packages of workspace
```

两条分支都不再 JOIN `ur.collaboration_workspace_id`。

### 4.3 拒绝规则矩阵

| 请求者 | personal 空间 | collaboration 空间 |
|---|---|---|
| super_admin | 全量 | 全量 |
| workspace.owner_user_id 匹配 | 全量 | owner 角色权限 |
| workspace_members 命中 | 不适用 | 按 role binding |
| 其他 | 403 | 403 |

## 5. 收口边界

### 5.1 必须收口(本任务树覆盖)

对应 S2–S7 节点:

- **S2 模型层**:删除 `CollaborationWorkspace*` 聚合根、4 个 snapshot 模型、3 张表的 bridge 字段;保留 `workspaces.workspace_type`;写 drop-column / drop-table migration。
- **S3 鉴权 + JWT**:middleware 简化为单 header 三键 context;JWT Claims 删除 collab 字段;auth/service.go 去硬编码。
- **S4 evaluator**:按 § 4.2 分流策略重写 SQL。
- **S5 OpenAPI**:删除 `/collaboration-workspaces/*` 路径、collaborationworkspace 模块、seed 对应条目;`make api` 重生成。
- **S6 重命名**:`space_key → menu_space_key`(菜单域字段不与 workspace 概念混淆)。
- **S7 旧前端**:合并 collaboration-workspace store → workspace store;去 header;去双 URL 调用。

### 5.2 暂不动(保留作为后续独立 PR)

- `modules/system/menu_spaces/` 目录名:本轮只规整字段到 `menu_space_key`,目录改名(menu_spaces → menu)放在下一轮。
- Vben 新前端 `frontend-platform/`:本轮不动,等旧前端收口完成后再同步迁移。
- 多租户 `tenant_id`:当前固定 `default`,按 `backend/CLAUDE.md` 规定所有查询仍需 filter,不在本任务范围。

## 6. 执行纪律

- **禁止新增**任何 `Collaboration` 前缀的新 Go 类型或任何 `collaboration_workspace_id` 新外键。
- **禁止新增**任何 `X-Collaboration-Workspace-Id` header 处理逻辑。
- **禁止新增**任何 `/collaboration-workspaces/*` OpenAPI 路径。
- **禁止在新前端代码中** import `collaboration-workspace.ts`。
- **中文展示只在 UI 渲染层**;store、API response、DTO 内一律使用 `workspace_type` 枚举,不持久化中文字符串。
- `backend/CLAUDE.md` "OpenAPI 是唯一真源"与 "不要手编 router.go" 的纪律在本任务中继续适用;S5.2 `make api` 后由 seed 自动挂载路由。

## 7. 与任务树映射

| Stage | 覆盖本文档的哪些节 |
|---|---|
| S2 模型层 & DB 迁移 | § 2.1、§ 5.1(模型删除)、§ 3.2(不再存在的概念 — Go 类型) |
| S3 鉴权中间件 & JWT | § 2.2、§ 4.1、§ 3.2(header / claim) |
| S4 权限评估器分流 | § 2.3、§ 4.2、§ 4.3 |
| S5 OpenAPI & 模块收口 | § 2.4、§ 3.2(URL / 模块包) |
| S6 space_key 重命名 | § 5.2(与 menu_spaces 目录改名的切分) |
| S7 旧前端收口 | § 2.5、§ 3.1(中文展示边界)、§ 3.2(前端文件) |

任务树:`tsk_01KPCPN9J96340BCPWN7B7` / `COLLAB-WS-CLEANUP`(挂于项目 `GGE`)。
