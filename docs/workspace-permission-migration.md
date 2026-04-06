# Workspace 权限迁移说明

> 适用阶段：Phase 0
>
> 当前状态：已完成文档基线，后续进入分阶段实现

## 阶段记录

| 阶段 | 状态 | 说明 |
| --- | --- | --- |
| Phase 0 | 已完成 | 已建立迁移说明、术语表和阶段记录文档，作为后续改造基线 |
| Phase 1 | 已完成 | 已新增 Workspace 领域模型、基础表结构与默认回填链，兼容旧 tenant 映射 |
| Phase 2 | 进行中 | 已在 auth 中间件建立 auth workspace 基础上下文，并提供当前授权 workspace 查询接口 |
| Phase 3 | 进行中 | team boundary 与 feature package 已开始优先读取 workspace 绑定表，team 角色读取已支持 workspace_role_bindings 优先 |
| Phase 4 | 进行中 | `explicit_target_workspace` 已进入统一授权层，平台代管请求开始强制 `target_workspace_id` |
| Phase 5 | 进行中 | 已新增 `/api/v1/workspaces/*` 基础接口，并给旧 tenant 响应补 workspace 映射字段 |
| Phase 6 | 进行中 | 前端已接入统一 workspace 切换入口，头部切换器开始显式展示 personal/team workspace，且 `lint + build` 已恢复为正式验收门槛 |
| Phase 7 | 待开始 | 菜单、运行时导航、APP 入口按 workspace 产出 |
| Phase 8 | 待开始 | 补种子、迁移脚本、回填策略 |
| Phase 9 | 已完成 | 后端测试、前端 `lint + build` 与关键场景通过性检查已完成 |
| Phase 10 | 已完成 | 文档、术语表、兼容边界与 tenant 语义回收已完成 |

## 1. 背景

当前仓库存在多套相互交叠的语义：

- `tenant` 被当成团队或租户上下文。
- `platform/team/common` 或类似分类仍是旧上下文判断线。
- `menu-space` 已经承担 APP 内菜单/宿主空间职责。
- `app` 是业务域。
- `role.TenantID != nil`、`context_type` 等逻辑仍在承担权限分流。

这会导致两个问题：

1. 权限主体不清晰，账号、tenant、menu-space、app 混在一起。
2. 平台后台和团队业务使用不同语义时，运行时很容易走偏。

## 2. 迁移目标

迁移后的主语义应统一为：

- `workspace` 是唯一业务权限主体。
- `workspace.type = personal | team`。
- `tenant` 不再承担当前团队主语义，预留给未来真正的多租户系统。
- `menu-space` 继续只表示 APP 内导航空间。
- `auth_workspace_id` 表示当前授权来源空间。
- `target_workspace_id` 表示当前操作目标空间。
- `permission.data_policy` 决定是否按空间数据域约束。

## 3. 当前旧语义

这里把现状写死，作为后续迁移的参照：

- `tenant` = 团队 / 租户上下文。
- `platform/team` 或 `platform/team/common` = 旧上下文分类。
- `menu-space` = APP 内菜单 / 宿主空间。
- `app` = 业务域。
- `role.TenantID != nil` 与 `context_type` = 旧判断线。

## 4. 新语义

- `workspace` 是唯一业务权限主体。
- `workspace.type` 只有两类：`personal`、`team`。
- `tenant` 不直接消失，但只保留在 legacy route、legacy persistence 与历史映射中。
- `menu-space` 不改造成权限空间。
- 平台后台权限也由 `personal workspace` 承载。
- 不再引入账号级全局业务角色。

## 5. 权限链

后续代码收口目标为：

`API / 页面 / APP入口 → 权限键 → 功能包 → 角色 → Workspace`

这里的职责边界如下：

- 资源层：API、页面、按钮、APP 入口。
- 授权单位：`permission_key`。
- 开通上限：`feature_package`。
- 分配裁剪：`role`。
- 最终承载：`workspace`。

## 6. 兼容原则

Phase 0 只做设计收口，不改业务行为。后续阶段遵循：

- 不直接删除 `tenant`。
- 不把 `menu-space` 改成权限空间。
- 不增加账号级全局业务角色。
- 不通过 `if admin return true` 继续扩散 bypass。
- 不强行给所有表增加 `workspace_id`。

## 7. 当前阶段完成项

- 已整理旧语义与新语义边界。
- 已固定权限链表达方式。
- 已明确 `workspace ≠ menu-space`。
- 已明确平台后台权限归属 `personal workspace`。
- 已明确后续迁移期间必须保留 `tenant` 兼容层。
- 已补齐 `workspace` 数据基线与基础查询接口。
- 已在请求上下文接入 `auth_workspace_id / auth_workspace_type`。
- 已让权限键开始显式承载 `app_key / data_policy / allowed_workspace_types`。
- 已让运行时权限快照能够区分 `personal workspace` 与 `team workspace`。
- 已在前端接入 `workspaceStore`，并让请求统一透传 `X-Auth-Workspace-Id`。
- 已把 `tenantStore` 降级为 team 兼容代理层，而不是前端唯一授权来源。
- 已让 `explicit_target_workspace` 在授权中间件阶段就校验 `target_workspace_id`，并支持从 query、header、JSON body 读取。
- 已让 team 功能包运行时优先读取 `workspace_feature_packages`，功能包分配写链开始同步写入 workspace 表。
- 已让团队成员角色读取优先使用 `workspace_role_bindings`，并在团队成员角色配置时同步写入 workspace 绑定。
- 已让平台角色读写与平台权限快照优先读取 personal workspace 角色绑定，旧 `user_roles(tenant_id IS NULL)` 退化为兼容镜像与回读来源。
- 已把团队成员身份与团队内部权限角色的文档语义拆开：`tenant_members.role_code / workspace_members.member_type` 表示成员身份，`workspace_role_bindings` 表示权限角色绑定。
- 已收窄统一授权层对 `explicit_target_workspace` 的默认兜底范围，避免把角色目录、功能包列表等非单目标接口一并误拦截。
- 已把头部团队切换器升级成统一 workspace 切换器，personal workspace 不再只能通过“进入平台管理”旁路切换。
- 已把 `feature-packages` 中带 `teamId / team_ids` 的平台代管接口接到 handler 级精确校验，旧 path/body 目标会先解析为 team workspace，再校验当前 personal workspace 是否有权代管。
- 已让 `/users/:id/packages` 语义切到“目标用户 personal workspace 的平台功能包绑定”，`workspace_feature_packages(personal workspace)` 成为主读/主写来源，旧 `user_feature_packages` 退化为镜像与回退来源。
- 已把 `space`、`page/runtime cache`、`message_service` 中剩余直接读取 `user_roles / tenant_members.role_code` 的关键运行时链改成 workspace 优先、legacy 回退。
- 已把平台快照、权限刷新链、功能包影响预览和权限消费统计中的平台用户/团队统计口径改成 workspace 优先，避免管理后台继续只统计 legacy 绑定。
- 已把 `GET /api/v1/feature-packages/:id/teams` 改成按当前 personal workspace 可代管范围过滤团队结果，避免 package 维度查询继续返回全部 team 绑定。
- 已把 `/api/v1/tenants/:id*` 这组平台治理接口开始接入基于 path `:id` 的目标 team workspace 精确校验，避免 legacy tenant 平台管理继续隐式拿当前上下文充当目标团队。
- 已把 `page` 平台访问轨迹中的平台角色读取切到 personal workspace 优先，并把 legacy 回退显式限定为 `user_roles.tenant_id IS NULL`，避免团队角色混入平台角色结果。
- 已把核心判权链 `authorization.getEffectiveActiveRoleIDs` 切到 workspace 优先，platform/team 两类运行时授权都会先读取对应 workspace 的角色绑定，再回退 legacy `user_roles`。
- 已清完前端全量 lint 存量债并恢复 `pnpm --dir frontend lint` 通过，前端 workspace 主线不再依赖旧 ESLint 环境异常或大量格式债掩盖真实问题。
- 已把系统管理页、权限组件、workspace 公共壳层和团队页中残留的 tenant-only 页面级上下文依赖收回边界层，目前只在消息域 helper 与旧 team API 请求头桥接中保留 tenant 兼容派生。
- 已把用户详情、平台快照和权限刷新链中的平台角色 legacy fallback 收紧到真正的 global role：除了 `user_roles.tenant_id IS NULL` 外，还要求 `roles.tenant_id IS NULL` 且 `roles.deleted_at IS NULL`。
- 已把活跃 header / DTO / 前端状态主线切到 workspace/team workspace 命名：`X-Auth-Workspace-Id` 为主授权头，`X-Team-Workspace-Id` 作为 team 兼容桥接头，`current_team_workspace_id / legacy_team_id / legacy_team_member_id` 等字段取代旧的 tenant 主输出语义。
- 已把消息域、workspace 切换、用户信息、团队列表和团队成员接口的活跃输出切到 `team_workspace_* / legacy_team_*` 命名；`X-Tenant-ID`、`current_tenant_id`、`source_tenant_id` 退化为兼容输入或历史映射字段。
- 已完成 Phase 9 验证：后端 `go test` 通过，前端 `pnpm --dir frontend lint` 与 `pnpm --dir frontend build` 通过。
- 已完成 Phase 10 收尾：`tenant` 已从当前 team/workspace 主线脱钩，并在术语表和阶段文档中明确预留给未来多租户系统。

## 7.1 角色语义补充

- 平台角色：平台后台配置的角色目录，运行时通过用户的 `personal workspace` 生效。
- 团队内部角色：团队工作空间内的角色目录或成员绑定，运行时通过对应 `team workspace` 生效。
- 成员身份：`tenant_members.role_code` 与 `workspace_members.member_type`，只表示成员关系边界，不等价于权限角色。

本轮实现明确遵循：

1. 账号不直接承载业务权限，平台角色主写到 `workspace_role_bindings(personal workspace)`。
2. 团队成员权限角色主写到 `workspace_role_bindings(team workspace)`，旧 `user_roles.tenant_id` 只做兼容镜像。
3. `roles.tenant_id` 当前仍允许承担团队角色目录的兼容存储职责，但不是运行时绑定主线。

## 7.2 运行时与前端语义补充

- `page/runtime_cache` 与 `space` 访问判断已经收紧为 `workspace_*` 优先、legacy 回退；成员身份字段只允许在明确的 identity 场景兜底，不再充当通用权限角色主线。
- 前端公共层开始统一展示 `当前授权工作空间 / 当前团队视图` 双语义，`tenantStore.currentTenantId` 保留为团队页面兼容派生值，不再承担公共壳层的主上下文解释。
- 消息域中的团队 ID 兼容派生已经开始统一收进 `useMessageWorkspace`，团队消息页继续走旧接口，但不再把 `tenantStore.currentTenantId` 直接扩散为页面级上下文来源。
- 目前前端页面层面对 `tenantStore.currentTenantId` 的直接消费已基本清空，仅在消息域 helper 与旧 team API header 桥接中保留兼容派生值。
- 前端 `pnpm lint` 的依赖环境已经恢复：`@typescript-eslint` 已锁到不触发 `minimatch@10` 解析问题的 `8.50.0`，`frontend/.npmrc` 也补了 `node-linker=hoisted`，当前 lint 失败点已切换为仓库内既有的存量规则债，而不是安装异常。

## 8. 验收标准

Phase 0 的文档基线满足以下条件即视为完成：

- 文档明确写出 `workspace ≠ menu-space`。
- 文档明确写出平台后台权限也来自 `personal workspace`。
- 文档明确写出不再使用账号级全局业务角色。
- 文档明确写出权限链从资源到 `workspace` 的收口方式。

## 9. 下一阶段建议

1. 后续若进入真正多租户设计，应新开 `tenant` 领域方案，不再复用当前 `team workspace` 兼容逻辑。
2. 若继续清理遗留兼容项，优先从 `X-Tenant-ID`、`current_tenant_id` 和 `/api/v1/tenants/*` 的只读桥接开始，避免影响现有 workspace 主线。
3. 保持新增接口、header、DTO 和页面状态统一使用 `workspace / team workspace / personal workspace` 命名，不再扩张 tenant-only 语义。

## 10. 当前一次性收口状态

截至 2026-04-06，本轮一次性收口与通过性检查的结论如下：

- 后端测试通过。
- 前端 `pnpm --dir frontend lint` 通过。
- 前端 `pnpm --dir frontend build` 通过。
- workspace 主线语义已经固定：
  - `personal workspace` 承载平台角色与平台功能包。
  - `team workspace` 承载团队角色与团队功能包。
  - `menu-space` 继续只承担导航宿主语义。
  - `tenant` 只保留兼容层职责。

当前明确保留的 tenant 兼容点只有：

- `frontend/src/store/modules/tenant.ts`
- `frontend/src/utils/http/index.ts`
- `frontend/src/views/message/modules/useMessageWorkspace.ts`
- 用户信息中的 `current_tenant_id` 兼容字段及其初始化读取路径

这些兼容点都不再承担授权源职责，只负责旧 team API、旧 team 视图或消息域过渡桥接。
