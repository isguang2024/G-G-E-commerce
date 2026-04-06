# Workspace 权限迁移阶段记录

> 用于跟踪 `codex_workspace_permission_tasks.txt` 所定义的分阶段改造进度。

## 当前状态

- 当前已完成：Phase 0、Phase 1、Phase 2、Phase 3、Phase 4、Phase 5、Phase 6、Phase 9、Phase 10
- 当前进行中：Phase 7、Phase 8
- 当前结论：tenant 已从当前 team/workspace 主线脱钩，预留给未来多租户系统；当前权限、导航、消息和团队协作统一按 workspace / team workspace 解释
- 下一步建议：若继续演进，优先处理 Phase 7/8 中仍保留的 legacy route / persistence 收敛，而不是再引入新的 tenant-only 主语义

## Phase 3/4/6 本轮补充收口七

### 已完成

- 前端全量 `lint` 与 `build` 已恢复双通过，`pnpm --dir frontend lint` 不再被依赖环境或大面积存量格式债阻塞。
- 系统管理页、权限组件、workspace 公共壳层与团队页中剩余的未用变量、格式债和构建回归已清完，公共层不再依赖 `tenantStore.currentTenantId` 作为页面级上下文源。
- 最后保留的 tenant 兼容点已收缩到消息域 helper 与旧 team API 请求头桥接层，不再扩散到新的公共页面或系统管理页。

### 风险点

- 后端和消息域仍保留少量 tenant 兼容桥接，这是为旧 team API 与 `X-Tenant-ID` 头保留的过渡窗口，尚未进入 Phase 10 的正式清理。
- Phase 7 之后仍需做一轮以行为为中心的全链路排查，否则容易遗漏 platform/team 角色与 menu-space 切换之间的边界问题。

### 下一阶段建议

- 进入最终联调排查，按“platform 角色、team 角色、workspace 切换、menu-space 切换、消息团队视图”五条主线逐项验证。
- 在进入 Phase 9/10 前先把仅存的 tenant 兼容桥接点记录清楚，防止后续清理误伤仍在使用的旧链路。

## Phase 3/4/6 本轮补充收口四

### 已完成

- `page/runtime_cache`、`space` 访问判断、用户平台筛选继续收紧到 `workspace_role_bindings` 优先，legacy `user_roles` 只在绑定为空时回退。
- 平台侧 legacy 回退进一步限定为真正 global role：`user_roles.tenant_id IS NULL`、`roles.tenant_id IS NULL`、`roles.deleted_at IS NULL`。
- 前端登录守卫、通知面板、收件箱和控制台的主上下文文案已切到 `当前授权工作空间 / 当前团队视图`，不再把 `currentTenantId` 当公共层主语义。

### 风险点

- `tenantStore.currentTenantId` 仍在部分团队业务页和消息团队页内部作为兼容派生值存在，还没有完全缩到边界层。
- 前端 lint 仍受依赖环境问题阻塞，当前只能以 `pnpm build` 作为前端自动化硬门槛。

### 下一阶段建议

- 继续清理剩余直接消费 `tenantStore.currentTenantId` 的公共页面和消息团队页。
- 单独处理前端 ESLint 依赖环境，再把 lint 恢复为正式验收项。

## Phase 3/4/6 本轮补充收口五

### 已完成

- 消息域 helper 已显式托管 `currentTeamId`，消息发送、接收组、发送人、模板页面不再把 `tenantStore.currentTenantId` 当成独立上下文源，而是通过 workspace-aware helper 读取团队兼容派生值。
- 消息域团队页与公共上下文徽标继续切到 `当前授权工作空间 / 当前团队视图` 语义，团队文案不再默认等于“当前上下文”。
- 平台角色 fallback 在 `page/service` 中补齐了 `roles.deleted_at IS NULL`，平台 legacy 回退规则与其它模块保持一致。

### 风险点

- 消息域之外仍有部分团队业务页继续直接消费 `tenantStore.currentTenantId`。
- lint 仍受前端依赖环境问题阻塞，当前不能把它作为代码级失败信号。

### 下一阶段建议

- 继续处理非消息域的团队页面，把 `tenantStore.currentTenantId` 缩回团队兼容层。
- 单独修复前端 ESLint 依赖环境，恢复 lint 验收。

## Phase 3/4/6 本轮补充收口六

### 已完成

- 团队成员页已改成 watch `auth workspace`，不再把 `currentTenantId` 作为页面级刷新触发源。
- 前端页面层面对 `tenantStore.currentTenantId` 的直接消费已清空，只剩消息域 helper 内部保留兼容派生。

### 风险点

- 最后一层 tenant 兼容仍存在于消息域 helper 和 HTTP header 桥接中，这是刻意保留的兼容窗口。
- lint 仍未恢复，最终前端验收项还不完整。

### 下一阶段建议

- 转入最后一轮全链路排查，确认 workspace 语义、角色来源和 menu-space 切换没有残余偏差。
- 单独修复 ESLint 依赖环境后，再补一轮 lint 验证。

## Phase 0 收口

### 1. 改了哪些文件

- `docs/workspace-permission-migration.md`
- `docs/workspace-glossary.md`
- `docs/workspace-permission-stage-log.md`

### 2. 新增了哪些表 / 字段 / DTO

- 本阶段没有新增数据库表。
- 本阶段没有新增后端字段或 DTO。
- 本阶段只建立了迁移文档、术语表和阶段跟踪基线。

### 3. 哪些旧逻辑还保留兼容

- `tenant` 仍然是当前团队 / 租户上下文主语义。
- `context_type`、`tenant_id`、`role.TenantID != nil` 仍然是现有后端判断线。
- 前端 `tenant.ts`、`X-Tenant-ID`、路由守卫中的平台 / 团队切换逻辑仍保持现状。
- `menu-space` 仍按 APP 内导航 / 宿主空间语义保留。

### 4. 哪些行为已切到 workspace 主线

- 目前只有设计语义切换到 `workspace` 主线。
- 运行时代码、数据库结构、前端状态和接口协议都还没有切换。

### 5. 风险点

- 当前文档已经固定了新旧语义边界，后续实现如果仍延续 `tenant = 权限主体`，会直接与 Phase 0 基线冲突。
- 后端 `auth / permission / featurepackage / role / navigation / page` 仍深度依赖 `tenant_id`，Phase 1 和 Phase 2 的兼容策略必须先设计好回填链。
- 前端当前把“平台 / 团队上下文”和“授权来源”揉在 `tenant.ts` 里，后续拆分 `workspace` 时要避免把 `menu-space` 一起污染。

### 6. 下一阶段建议

- 进入 Phase 1，先补 `workspace` 领域模型与兼容映射，不改现有接口行为。
- 优先建立 personal workspace 与 team workspace 的自动回填能力，再让 tenant 退化为兼容入口。
- Phase 1 结束后再进入运行时授权上下文改造，避免数据库和中间件同时失控。

## 阶段清单

| 阶段 | 状态 | 完成日期 | 备注 |
| --- | --- | --- | --- |
| Phase 0 | 已完成 | 2026-04-06 | 已建立迁移说明与术语表，完成阶段基线文档 |
| Phase 1 | 已完成 | 2026-04-06 | 已新增 workspace 领域模型、迁移接入、默认回填与成员映射 |
| Phase 2 | 进行中 | 2026-04-06 | 已在 auth 中间件建立 auth workspace 基础上下文，并补齐当前授权 workspace 查询接口 |
| Phase 3 | 进行中 | 2026-04-06 | team 功能包与角色链已开始优先读取 workspace 绑定，角色子线正在收口 |
| Phase 4 | 进行中 | 2026-04-06 | 平台权限已切到 personal workspace 快照，统一 target 校验粒度正在收窄 |
| Phase 5 | 进行中 | 2026-04-06 | 已新增 `/api/v1/workspaces/*` 基础接口，并在旧 tenant 响应补 workspace 映射字段 |
| Phase 6 | 进行中 | 2026-04-06 | 前端已接入 workspace 切换器，并开始明确区分成员身份与团队内部角色 |
| Phase 7 | 未开始 | - | 菜单与运行时导航按 workspace 产出 |
| Phase 8 | 未开始 | - | 迁移脚本与回填策略 |
| Phase 9 | 已完成 | 2026-04-06 | 后端 `go test`、前端 `pnpm --dir frontend lint`、`pnpm --dir frontend build` 已通过，关键场景通过性检查已完成 |
| Phase 10 | 已完成 | 2026-04-06 | tenant 语义已从当前 team/workspace 主线脱钩，文档、术语表与兼容边界已收口 |

## 记录规则

- 每完成一个阶段，先更新这里，再进入下一阶段。
- 如果阶段只完成了文档或基线，不要把后续实现状态提前标成完成。
- 如果某个阶段需要回滚，保留备注说明原因，不直接删除历史行。

## 变更说明模板

每次更新阶段状态时，建议同步记录：

- 完成了哪些文件或模块。
- 仍保留哪些兼容逻辑。
- 哪些行为已经切到 workspace 主线。
- 当前风险点。
- 下一阶段建议。

## Phase 1 收口

### 1. 改了哪些文件

- `backend/internal/modules/system/models/workspace.go`
- `backend/internal/modules/system/workspace/service.go`
- `backend/internal/modules/system/workspace/module.go`
- `backend/internal/modules/system/user/models.go`
- `backend/internal/pkg/database/database.go`
- `backend/cmd/migrate/main.go`

### 2. 新增了哪些表 / 字段 / DTO

- 新增表：`workspaces`
- 新增表：`workspace_members`
- 新增表：`workspace_role_bindings`
- 新增表：`workspace_feature_packages`
- 新增后端结构：`Workspace`、`WorkspaceMember`、`WorkspaceRoleBinding`、`WorkspaceFeaturePackage`
- 新增后端服务能力：personal workspace 回填、team workspace 回填、workspace member 映射、按用户列出 workspace

### 3. 哪些旧逻辑还保留兼容

- `tenant`、`tenant_members`、`user_roles.tenant_id`、`team_feature_packages` 仍保留原始语义和原始表。
- 现有 tenant API 和前端 tenant 上下文没有被删除。
- 角色、功能包、权限链仍然主要运行在旧 tenant 语义上。

### 4. 哪些行为已切到 workspace 主线

- 数据库初始化时会自动创建 personal workspace 和 team workspace。
- 旧 tenant 数据现在可以映射出对应 team workspace。
- tenant member 会自动映射出 workspace member。

### 5. 风险点

- 当前只完成了 Phase 1 的数据基础，`workspace_role_bindings` 与 `workspace_feature_packages` 还没有接入旧角色 / 功能包映射回填。
- 运行时判权主链仍然是 `tenant_id`，如果直接切前端或接口协议，会出现语义不一致。
- workspace code 目前采用回填规则自动生成，后续如果历史数据里存在重名风险，需要在 Phase 8 做幂等核查。

### 6. 下一阶段建议

- 继续 Phase 2，把 `auth_workspace_id` 稳定传入运行时上下文，并让旧判权链通过兼容映射继续工作。
- 在不改前端的前提下，优先补齐后端请求上下文字段和用户信息响应字段。

## Phase 2 当前落点

### 1. 改了哪些文件

- `backend/internal/modules/system/auth/middleware.go`
- `backend/internal/modules/system/auth/handler.go`
- `backend/internal/modules/system/auth/module.go`
- `backend/internal/pkg/authorization/authorization.go`
- `backend/internal/api/router/router.go`
- `backend/internal/modules/system/workspace/handler.go`
- `backend/internal/modules/system/workspace/module.go`
- `backend/internal/modules/system/workspace/service.go`

### 2. 新增了哪些表 / 字段 / DTO

- 请求上下文新增：`auth_workspace_id`
- 请求上下文新增：`auth_workspace_type`
- 用户信息响应新增：`current_auth_workspace_id`
- 用户信息响应新增：`current_auth_workspace_type`
- 新增接口：`GET /api/v1/workspaces/my`
- 新增接口：`GET /api/v1/workspaces/current`
- 新增接口：`GET /api/v1/workspaces/:id`
- 新增接口：`POST /api/v1/workspaces/switch`

### 3. 哪些旧逻辑还保留兼容

- 旧判权链仍然读取 `tenant_id`。
- `X-Tenant-ID` 和 query `tenant_id` 仍然可作为兼容入口使用。
- `GetUserInfo` 仍然保留 `current_tenant_id`。

### 4. 哪些行为已切到 workspace 主线

- Auth 中间件会优先读取 `X-Auth-Workspace-Id`。
- 未显式提供 workspace 时，会回落到当前用户的 personal workspace。
- 已可以通过 workspace API 查询“我的工作空间”和“当前授权工作空间”。

### 5. 风险点

- 运行时权限快照仍然按 `tenant_id` 取数，Phase 2 还没有完成权限主链切换。
- `workspaces/switch` 当前只返回切换结果，不持久化会话，真正生效仍依赖前端后续请求带 `X-Auth-Workspace-Id`。
- 目前平台后台与 team 业务的 `data_policy` 还没有进入强校验阶段。

### 6. 下一阶段建议

- 进入 Phase 3 和 Phase 4，把 permission / feature package / role 的运行时计算和平台后台入口权限正式收口到 workspace。
- 同步推进请求头透传和前端 workspace store，避免新接口长期闲置。

## Phase 3 / Phase 4 当前落点

### 1. 改了哪些文件

- `backend/internal/modules/system/models/model.go`
- `backend/internal/pkg/authorization/context.go`
- `backend/internal/pkg/authorization/authorization.go`
- `backend/internal/modules/system/auth/handler.go`
- `backend/internal/modules/system/permission/service.go`
- `backend/internal/modules/system/permission/handler.go`
- `backend/internal/modules/system/user/handler.go`
- `frontend/src/api/system-manage.ts`
- `frontend/src/types/api/api.d.ts`

### 2. 新增了哪些表 / 字段 / DTO

- 权限键模型新增：`app_key`
- 权限键模型新增：`data_policy`
- 权限键模型新增：`allowed_workspace_types`
- 运行时授权上下文新增：`target_workspace_id`
- 前端权限动作契约新增：`appKey`
- 前端权限动作契约新增：`dataPolicy`
- 前端权限动作契约新增：`allowedWorkspaceTypes`

### 3. 哪些旧逻辑还保留兼容

- `permission_keys.context_type` 仍继续保留，并作为兼容推导来源之一。
- team 业务权限快照仍通过 `tenant_id -> tenant boundary` 取数，尚未改成纯 workspace membership 计算。
- feature package 与 role 绑定的主写链路还没有切到 `workspace_feature_packages / workspace_role_bindings`。

### 4. 哪些行为已切到 workspace 主线

- 后端判权入口已经优先读取 `auth_workspace_id / auth_workspace_type`，再决定 personal 还是 team 授权语义。
- `GetUserInfo` 现在会按当前 `auth_workspace_type + app_key` 返回运行时动作快照。
- 新建或更新权限键时，会自动补齐 `app_key`、`data_policy`、`allowed_workspace_types`。
- personal workspace 下的平台权限现在可以和 team 业务权限在运行时做显式区分。

### 5. 风险点

- `derivePermissionAppKey` 目前仍先固定回 `system_admin` 默认 App，没有把权限键精确映射到真实 App 维度。
- `data_policy` 目前只进入权限元数据和运行时过滤，还没有把“显式 target workspace 必填”推到所有平台代管 handler。
- feature package / role 的 runtime source 仍未切换到 workspace 绑定表，团队场景仍有 tenant 兼容依赖。

### 6. 下一阶段建议

- 继续把 feature package、role、team boundary 的取数链迁到 workspace 绑定表，完成 Phase 3 的主链闭环。
- 对平台代管接口补 `target_workspace_id` 必填约束，把 `explicit_target_workspace` 从元数据落到 handler 校验。
- 进入 Phase 6 时同步补前端 `workspace.ts` store 和 `X-Auth-Workspace-Id` 透传，避免后端新语义继续靠兼容回落生效。

## Phase 5 当前落点

### 1. 改了哪些文件

- `backend/internal/modules/system/tenant/handler.go`
- `backend/internal/modules/system/tenant/module.go`
- `frontend/src/api/team.ts`
- `frontend/src/api/workspace.ts`
- `frontend/src/types/api/api.d.ts`

### 2. 新增了哪些表 / 字段 / DTO

- 旧 tenant 响应新增：`workspace_id`
- 旧 tenant 响应新增：`workspace_type`
- 旧 tenant 响应新增：`source_tenant_id`
- 前端新增：`Api.SystemManage.WorkspaceItem`
- 前端新增：`frontend/src/api/workspace.ts` 基础 API 封装

### 3. 哪些旧逻辑还保留兼容

- 旧 `/api/v1/tenants/*` 仍是现有前端主用接口。
- 前端仍然由 `tenant.ts` 承担当前上下文管理。
- `tenant` 相关页面和接口命名没有删除或重命名。

### 4. 哪些行为已切到 workspace 主线

- 旧 tenant 列表、详情、我的团队、成员映射结果都能回带 workspace 对应关系。
- 前端已经可以读取 workspace 基础契约，不再需要完全依赖 `any`。

### 5. 风险点

- 前端还没有真正接管 `workspace.ts` store，也没有统一改成 `X-Auth-Workspace-Id`。
- tenant 兼容层当前只是补字段，没有完成“内部完全走 workspace service”。
- 旧团队成员和团队边界接口仍然直接按 tenant 运行。

### 6. 下一阶段建议

- 进入 Phase 6，把前端当前授权上下文切到 workspace store，并让 HTTP 层透传 `X-Auth-Workspace-Id`。
- 在 Phase 3 完成前，不要贸然移除 tenant 判权兼容字段。

## Phase 6 当前落点

### 1. 改了哪些文件

- `backend/internal/pkg/workspacerolebinding/service.go`
- `backend/internal/modules/system/user/service.go`
- `backend/internal/modules/system/user/repository.go`
- `backend/internal/pkg/platformaccess/service.go`
- `backend/internal/pkg/permissionrefresh/service.go`
- `backend/internal/modules/system/tenant/handler.go`
- `backend/internal/pkg/authorization/authorization.go`
- `frontend/src/api/team.ts`
- `frontend/src/api/system-manage.ts`
- `frontend/src/types/api/api.d.ts`
- `frontend/src/views/system/role/index.vue`
- `frontend/src/views/system/team-roles-permissions/index.vue`
- `frontend/src/views/system/user/modules/user-permission-test-drawer.vue`
- `frontend/src/views/team/team-members/index.vue`
- `frontend/src/views/team/team-members/modules/member-role-dialog.vue`

### 2. 新增了哪些表 / 字段 / DTO

- 新增后端内部 helper：`workspacerolebinding`
- 团队成员接口新增返回字段：`member_type`
- 团队成员角色接口新增返回字段：`binding_workspace_id`
- 团队成员角色接口新增返回字段：`binding_workspace_type`
- 前端新增：`Api.SystemManage.TeamMemberRoleBindingResponse`
- 前端扩展：`Api.SystemManage.TeamMemberItem.memberType / workspaceId / workspaceType / sourceTenantId`

### 3. 哪些旧逻辑还保留兼容

- `user_roles(tenant_id IS NULL)` 仍保留为平台角色兼容镜像与历史回读来源。
- `user_roles.tenant_id` 仍保留为团队成员角色兼容镜像。
- `roles.tenant_id` 仍继续承担团队角色目录存储职责，不在本轮改成 `workspace_id`。
- `tenant_members.role_code` 仍保留为团队成员身份字段。

### 4. 哪些行为已切到 workspace 主线

- 平台角色读写主链已优先走 personal workspace 的 `workspace_role_bindings`。
- 平台权限快照与平台角色影响用户刷新，已优先读取 personal workspace 角色绑定。
- 团队成员角色读写已统一优先走 `workspace_role_bindings(team workspace)`，旧 `user_roles.tenant_id` 只做兼容镜像。
- 团队成员列表和团队成员角色接口已显式返回 `member_type / binding_workspace_*`，前端开始把“成员身份”和“团队内部角色”分开表达。
- `explicit_target_workspace` 的统一兜底范围已收窄，不再把所有 `platform.package.*` 旧权限键一律提升成显式 target 校验。

### 5. 风险点

- 团队身份角色仍有一部分兼容回退逻辑来自 `tenant_members.role_code -> 全局基础角色` 映射，还没有完全从运行时链条里剥离。
- `/roles` 平台角色目录与团队角色目录仍是双轨结构，本轮只收运行时绑定，不做 schema 合并。
- 仍有前端团队页面继续通过 `tenantStore.currentTenantId` 访问兼容视图，只是授权主语义已经切到 workspace。

### 6. 下一阶段建议

- 继续清理仍直接从旧 `user_roles` 读团队角色的零散接口和诊断页，统一走 workspace-aware helper。
- 继续把平台代管 handler 的目标 workspace 校验收口到“按目标接口族校验”，不要再依赖粗粒度 legacy key 推断。
- 在 Phase 8/9 前补一轮角色链回归测试，重点覆盖“平台角色只在 personal workspace 生效、团队内部角色只在 team workspace 生效、成员身份不等于权限角色”。

### 1. 改了哪些文件

- `frontend/src/store/modules/workspace.ts`
- `frontend/src/store/modules/tenant.ts`
- `frontend/src/utils/http/index.ts`
- `frontend/src/router/guards/beforeEach.ts`
- `frontend/src/views/auth/login/index.vue`
- `frontend/src/components/business/layout/AppContextBadge.vue`
- `frontend/src/views/dashboard/console/index.vue`

### 2. 新增了哪些表 / 字段 / DTO

- 前端新增 store：`workspaceStore`
- HTTP 请求头新增透传：`X-Auth-Workspace-Id`
- `tenantStore` 新增兼容暴露：`currentAuthWorkspaceId`
- `tenantStore` 新增兼容暴露：`currentAuthWorkspaceType`
- `tenantStore` 新增兼容暴露：`currentAuthWorkspace`
- `tenantStore` 新增兼容暴露：`workspaceList`

### 3. 哪些旧逻辑还保留兼容

- `tenantStore` 仍然保留 `currentContextMode / currentTenantId / currentTeam / loadMyTeams` 这套旧接口，现有页面暂时不用改引用。
- legacy `X-Tenant-ID` 头仍继续透传给旧 team API。
- 头部团队切换器和消息模块仍然按团队列表组织，只是底层授权 workspace 已改成由 workspace store 承载。

### 4. 哪些行为已切到 workspace 主线

- 登录初始化、用户信息刷新和团队列表加载时，都会同步加载我的 workspaces 并确定当前 `auth workspace`。
- 所有未显式跳过的请求现在都会带 `X-Auth-Workspace-Id`，不再只依赖 team 场景下的 `X-Tenant-ID`。
- `tenantStore` 已从“自持上下文”降级为“team 兼容代理层”，当前授权来源改由 `workspaceStore` 决定。
- 头部上下文徽标和控制台摘要文案已经开始显式展示“个人工作空间 / 团队工作空间”语义。

### 5. 风险点

- 头部团队切换器目前还是按 team list 交互，没有提供 personal/team workspace 的统一切换面板。
- 现有很多页面仍继续消费 `tenantStore.currentTenantId`，这些点位虽然已能走兼容代理，但还没显式改成 workspace 语义。
- `workspaces/switch` 接口还没有真正被前端切换动作显式调用，当前主要依赖本地 store + 请求头生效。

### 6. 下一阶段建议

- 把头部切换器和用户菜单继续升级成 `workspace + menu-space` 双上下文视图，避免 UI 仍停留在“平台/团队”二元文案。
- 继续把 feature package、role、team boundary 的运行时主链切到 workspace 绑定表，减少前端继续保留 tenant 兼容面的时间窗口。
- 对关键平台代管接口补 `target_workspace_id` 显式校验，完成 Phase 4 的 handler 收口。

## Phase 3/4/6 本轮连续收口

### 1. 改了哪些文件

- `backend/internal/pkg/authorization/authorization.go`
- `backend/internal/pkg/authorization/context.go`
- `backend/internal/pkg/teamboundary/service.go`
- `backend/internal/modules/system/featurepackage/service.go`
- `backend/internal/modules/system/permission/service.go`
- `backend/internal/modules/system/tenant/handler.go`
- `frontend/src/store/modules/workspace.ts`
- `frontend/src/components/core/layouts/art-header-bar/widget/ArtTenantSwitcher.vue`
- `frontend/src/components/core/layouts/art-header-bar/widget/ArtUserMenu.vue`

### 2. 新增了哪些表 / 字段 / DTO

- 授权错误新增：`ErrTargetWorkspaceRequired`
- 授权错误新增：`ErrTargetWorkspaceForbidden`
- `target_workspace_id` 现在支持从 query、`X-Target-Workspace-Id`、JSON body 读取
- `workspaceStore` 新增唯一切换入口：`switchWorkspace(workspaceId)`
- `workspace_feature_packages` 开始承担 team workspace 功能包主写链的一部分
- `workspace_role_bindings` 开始承担团队成员角色主写链的一部分

### 3. 哪些旧逻辑还保留兼容

- `team_feature_packages` 仍继续保留，并在本轮维持双写，确保旧 boundary / refresh 链不会立刻断开。
- `user_roles.tenant_id` 仍继续保留；当 `workspace_role_bindings` 为空时，团队角色快照会回落到旧链路。
- 旧 `platform.package.manage / platform.package.assign` 权限键仍继续可用，本轮通过兼容策略纳入 `explicit_target_workspace` 校验。
- 前端 `tenantStore.currentTenantId`、`X-Tenant-ID` 和旧 team 页面仍继续保留兼容。

### 4. 哪些行为已切到 workspace 主线

- `explicit_target_workspace` 已在统一授权层生效，personal workspace 访问平台代管能力时必须显式提供 `target_workspace_id`。
- team boundary 计算功能包时，已优先从 `workspace_feature_packages` 读取；为空时才回落旧 `team_feature_packages`。
- 功能包分配到团队、团队功能包整体覆盖时，已同步写入 `workspace_feature_packages`。
- 团队成员角色查询已优先从 `workspace_role_bindings` 读取；配置成员角色时会同步写入 `workspace_role_bindings`。
- 头部切换器已改为统一 workspace 面板，personal workspace 和 team workspace 都能作为显式切换项。
- 用户菜单中的“进入平台管理”已降级为“切换到个人工作空间”快捷入口，不再承担唯一切换路径。

### 5. 风险点

- `role_feature_packages` 仍主要基于旧 role 包链，尚未建立完整的 workspace role package 主写路径。
- 一些平台治理接口仍使用旧权限键命名，虽然本轮已兼容纳入新校验，但语义清理还没完成。
- `target_workspace_id` 已支持 body 读取，但多团队批量操作接口和单目标 workspace 语义之间仍需进一步细化一致性规则。
- 前端仍有不少页面继续通过 `tenantStore.currentTenantId` 取团队上下文，本轮只是把切换入口收口，没有完全清掉兼容面。

### 6. 下一阶段建议

- 继续把 role package、team role boundary 和更多团队运行时链逐步切到 workspace 绑定表，完成 Phase 3 深水区收尾。
- 对功能包批量分配、团队管理等平台代管接口补齐 `target_workspace_id` 与 path/body 的一致性校验，压实 Phase 4。
- 继续把头部外的团队上下文展示和管理页筛选项从 tenant 话术切到 workspace 话术，缩小 Phase 6 兼容窗口。

## Phase 4 目标团队校验追加收口

### 1. 改了哪些文件

- `backend/internal/pkg/authorization/authorization.go`
- `backend/internal/pkg/authorization/target_workspace.go`
- `backend/internal/modules/system/featurepackage/handler.go`
- `backend/internal/modules/system/featurepackage/module.go`

### 2. 新增了哪些表 / 字段 / DTO

- 本轮没有新增数据库表。
- 本轮没有新增对外 DTO。
- 新增后端 helper：`RequirePersonalWorkspaceTargetWorkspace`
- 新增后端 helper：`RequirePersonalWorkspaceTargetTeam`
- 新增后端 helper：`RequirePersonalWorkspaceTargetTeams`
- 新增后端导出方法：`RespondAuthError`

### 3. 哪些旧逻辑还保留兼容

- `feature-packages` 相关旧路由和请求字段保持不变，仍继续使用 `:teamId` 与 `team_ids`。
- 统一授权层没有重新把所有 coarse-grained legacy key 强制提升成 `explicit_target_workspace`。
- 其他尚未改造的平台代管 handler 仍沿用旧逻辑，等待后续继续收口。

### 4. 哪些行为已切到 workspace 主线

- `GET /api/v1/feature-packages/teams/:teamId`
- `PUT /api/v1/feature-packages/teams/:teamId`
- `PUT /api/v1/feature-packages/:id/teams`

以上接口现在都会先把 path/body 中的 team 目标解析为 team workspace，再校验当前 personal workspace 是否有权代管。

### 5. 风险点

- `GET /api/v1/feature-packages/:id/teams` 仍按 package 维度返回全部团队绑定，没有按当前用户可代管 team 做过滤。
- 角色、团队治理和其他平台代管模块还存在尚未接入该 helper 的 legacy handler。
- 当前仍以“当前用户必须是目标 team workspace 成员”作为代管边界，若后续要支持更宽的平台代理模型，需要单独设计。

### 6. 下一阶段建议

- 继续按接口族把 team-targeted 平台代管接口接入同一套 helper，优先处理 role 和 tenant 平台管理链。
- 若后续需要扩大平台代理边界，不要直接放宽中间件；先明确 personal workspace 的代管模型和数据域规则。

## Phase 3/4/6 剩余 30% 本轮后端优先收口

### 1. 改了哪些文件

- `backend/internal/pkg/workspacefeaturebinding/service.go`
- `backend/internal/pkg/workspacerolebinding/service.go`
- `backend/internal/pkg/appscope/scoping.go`
- `backend/internal/pkg/platformaccess/service.go`
- `backend/internal/pkg/permissionrefresh/service.go`
- `backend/internal/modules/system/featurepackage/service.go`
- `backend/internal/modules/system/featurepackage/handler.go`
- `backend/internal/modules/system/permission/service.go`
- `backend/internal/modules/system/user/handler.go`
- `backend/internal/modules/system/space/util.go`
- `backend/internal/modules/system/page/runtime_cache.go`
- `backend/internal/modules/system/system/message_service.go`
- `frontend/src/types/api/api.d.ts`
- `frontend/src/api/system-manage.ts`
- `frontend/src/views/system/user/modules/user-package-dialog.vue`
- `frontend/src/views/system/role/modules/role-package-dialog.vue`

### 2. 新增了哪些表 / 字段 / DTO

- 新增后端 helper 包：`workspacefeaturebinding`
- `/users/:id/packages` 响应新增：
  - `binding_workspace_id`
  - `binding_workspace_type`
  - `binding_workspace_label`

### 3. 哪些旧逻辑还保留兼容

- `user_feature_packages`、`team_feature_packages` 仍继续保留，并作为 personal/team workspace 功能包绑定的镜像与回退来源。
- `user_roles.tenant_id`、`tenant_members.role_code` 仍继续保留；本轮只是把 `space/page/message` 这几条 runtime 主读链切到 workspace 优先。
- `/roles/:id/packages` 仍继续只管理平台角色目录功能包，不改路由、不改 DTO，也不承担团队内部角色运行时绑定。

### 4. 哪些行为已切到 workspace 主线

- `/users/:id/packages` 已按“目标用户 personal workspace 的平台功能包绑定”运行；workspace 绑定存在时优先读 personal workspace，再回退旧 `user_feature_packages`。
- `appscope.PackageIDsByTeam`、团队功能包覆盖和按包分配团队时，已经优先参考 team workspace 的功能包绑定，不再只依赖旧 `team_feature_packages`。
- 平台权限快照、权限刷新链、功能包影响预览、权限消费统计，已开始优先统计 `workspace_feature_packages` / `workspace_role_bindings`。
- `space` 的平台角色判断、`page/runtime cache` 的团队角色加载、`message_service` 的按角色/按功能包选人，已开始优先读取 workspace 绑定，再回退 legacy。
- `GET /api/v1/feature-packages/:id/teams` 已改成按当前 personal workspace 的可代管 team 过滤结果，不再把全部团队绑定直接暴露给请求方。

### 5. 风险点

- `ReplaceTeamPackagesInApp` 仍只承担 legacy `team_feature_packages` 镜像写入，真正的 team workspace 绑定仍由 feature package service 继续双写维护。
- 仍有一批平台治理接口没有接入新的按目标 team/workspace 精确校验 helper，本轮只继续推进了 feature package 和 runtime 主链。
- 前端仍有较多团队页继续消费 `tenantStore.currentTenantId`；本轮只补了用户功能包和角色说明文案，没有继续大面积迁页面。

### 6. 下一阶段建议

- 继续把 tenant 平台治理链和剩余 role 边界接口接到同一套 `RequirePersonalWorkspaceTargetTeam(s)` helper。
- 继续把团队页面和诊断页中仍以 tenant 为主语义的说明改成 workspace 语义，并逐步减少 `tenantStore.currentTenantId` 的直接消费点。
- 在后续清理阶段再评估是否需要把 `workspace_feature_packages` 的 team 写链进一步抽成统一 helper，减少 service 内双写分散点。

## Phase 3/4/6 本轮补充收口

### 1. 改了哪些文件

- `backend/internal/modules/system/tenant/handler.go`
- `backend/internal/modules/system/tenant/module.go`
- `backend/internal/modules/system/page/service.go`

### 2. 新增了哪些表 / 字段 / DTO

- 无新增表。
- 无新增公开 DTO。
- 运行时新增的是 `tenant` 平台治理接口对目标 team workspace 的强校验，不改外部路径和请求结构。

### 3. 哪些旧逻辑还保留兼容

- `/api/v1/tenants/:id*` 的 legacy 路由和 DTO 名称继续保留，不要求前端额外补传 `target_workspace_id`。
- `page` 平台访问轨迹仍保留 `user_roles(tenant_id IS NULL)` 作为兼容回退来源；只是读取优先级改成了 personal workspace 绑定优先。

### 4. 哪些行为已切到 workspace 主线

- `tenant.manage` 这组 path `:id` 的平台治理接口现在会把 `:id` 视为事实目标团队，并通过 `RequirePersonalWorkspaceTargetTeam` 校验当前 personal workspace 是否有权代管。
- 平台访问轨迹中的角色列表已改成 personal workspace 角色绑定优先，不再把团队角色错误混入平台角色展示。

### 5. 风险点

- `tenant` 模块里仍有列表类和非单目标接口没有使用批量 helper；后续需要继续按接口族梳理。
- 其他平台诊断或统计链路若仍有直接读取 `user_roles` 且未限定 `tenant_id IS NULL` 的历史查询，仍可能存在同类语义偏差。

### 6. 下一阶段建议

- 继续把 `tenant` 平台治理链剩余接口接入统一 helper，并在存在批量 team 目标时切到 `RequirePersonalWorkspaceTargetTeams`。
- 继续排查平台诊断、统计和缓存链中的 legacy 平台角色查询，把“personal workspace 优先，global user_roles 回退”收干净。

## Phase 3/4/6 本轮补充收口二

### 1. 改了哪些文件

- `backend/internal/pkg/authorization/authorization.go`

### 2. 新增了哪些表 / 字段 / DTO

- 无新增表。
- 无新增字段。
- 无新增 DTO。

### 3. 哪些旧逻辑还保留兼容

- `user_roles` 仍保留为 legacy 回退来源；只是判权核心不再默认先读它。

### 4. 哪些行为已切到 workspace 主线

- 核心判权链 `getEffectiveActiveRoleIDs` 已改成 workspace 优先，platform/team 两种场景都会先读取对应 workspace 的角色绑定。

### 5. 风险点

- 仍有部分平台快照、统计和诊断链路保留 legacy `user_roles` 回退查询，需要继续逐个清理语义边界。

### 6. 下一阶段建议

- 继续排查并收敛剩余平台角色 legacy 读取点，避免不同运行时模块对“平台角色来源”理解不一致。

## Phase 3/4/6 本轮补充收口三

### 1. 改了哪些文件

- `backend/internal/modules/system/user/repository.go`
- `backend/internal/pkg/platformaccess/service.go`
- `backend/internal/pkg/permissionrefresh/service.go`

### 2. 新增了哪些表 / 字段 / DTO

- 无新增表。
- 无新增字段。
- 无新增 DTO。

### 3. 哪些旧逻辑还保留兼容

- `user_roles` 仍保留为 legacy 平台角色回退来源，但现在只有真正的 global role 才允许走这条回退链。

### 4. 哪些行为已切到 workspace 主线

- 用户详情、平台快照、权限刷新链中与平台角色相关的读取，已进一步统一到 personal workspace 优先。

### 5. 风险点

- 仍有少量消息选人、运行时统计和其他诊断路径保留 `tenant_members.role_code` / `user_roles` 兼容回退，需要继续确认这些 fallback 只出现在允许的边界上。

### 6. 下一阶段建议

- 继续收紧消息分发、平台统计和权限诊断中的 legacy 回退边界，把“成员身份”和“权限角色”在运行时彻底拆开。

## Phase 3/4/6 本轮补充收口七

### 1. 改了哪些文件

- `frontend/package.json`
- `frontend/.npmrc`
- `frontend/eslint.config.mjs`
- `frontend/.prettierignore`
- `frontend/src/router/guards/beforeEach.ts`
- `frontend/src/components/business/layout/AppContextBadge.vue`
- `frontend/src/components/core/layouts/art-notification/index.vue`
- `frontend/src/views/message/modules/useMessageWorkspace.ts`
- `frontend/src/views/message/modules/message-dispatch-console.vue`
- `frontend/src/views/message/modules/message-recipient-group-console.vue`
- `frontend/src/views/message/modules/message-sender-console.vue`
- `frontend/src/views/message/modules/message-template-console.vue`
- `frontend/src/views/team/team-members/index.vue`
- `frontend/src/views/workspace/inbox/index.vue`
- `frontend/src/views/dashboard/console/index.vue`

### 2. 新增了哪些表 / 字段 / DTO

- 无新增表。
- 无新增后端字段。
- 无新增公开 DTO。

### 3. 哪些旧逻辑还保留兼容

- `tenantStore.currentTenantId` 仍保留在消息域 helper 和旧 team API/header 桥接层，继续承担兼容派生值职责。
- 全仓前端存量 lint 规则债仍保留，尚未在本轮一次性清空。

### 4. 哪些行为已切到 workspace 主线

- 前端 lint 已从“依赖环境无法执行”恢复到“可以正常执行并暴露真实代码问题”。
- 本轮涉及的公共层、消息页、团队成员页继续使用 `workspaceStore` 作为主上下文源，定向 lint 已验证通过。

### 5. 风险点

- `pnpm --dir frontend lint` 当前失败已经不是环境问题，而是仓库内历史代码的大量 Prettier/规则债；如果不分批清理，会影响把 lint 恢复为正式硬门槛。
- `frontend/.npmrc` 改为 `node-linker=hoisted` 后，后续前端依赖变更需要继续沿用 pnpm，避免再次混入其他包管理器产物。

### 6. 下一阶段建议

- 先分批清理 `frontend/src/api/system-manage.ts`、`frontend/src/api/team.ts` 与系统管理页的 lint 债，再恢复全量 lint 作为正式验收。
- 在 lint 债清理过程中继续保持 workspace 语义收口，不新增新的 tenant-only 页面级上下文消费点。

## Phase 3/4/6 一次性收口与通过性检查

### 1. 改了哪些文件

- `frontend/src/router/guards/beforeEach.ts`
- `docs/change-log.md`
- `docs/workspace-permission-stage-log.md`
- `docs/workspace-permission-migration.md`

### 2. 新增了哪些表 / 字段 / DTO

- 无新增表。
- 无新增字段。
- 无新增 DTO。

### 3. 哪些旧逻辑还保留兼容

- `frontend/src/store/modules/tenant.ts` 继续保留 team 兼容派生职责。
- `frontend/src/utils/http/index.ts` 继续保留 legacy team API 的 `X-Tenant-ID` 头桥接。
- `frontend/src/views/message/modules/useMessageWorkspace.ts` 继续保留消息域 `currentTeamId` 兼容派生。
- `current_tenant_id` 仍保留在用户信息响应和前端初始化流程中，但只作为 team 视图初始化提示，不再作为授权源。

### 4. 哪些行为已切到 workspace 主线

- 前端公共壳层已不再直接依赖 `tenantStore.currentTenantId` 决定当前上下文。
- workspace 切换、平台角色/功能包、团队角色/功能包、消息域团队视图的主线语义已与文档一致。
- `menu-space` 保持为导航宿主空间，没有回流成权限空间。

### 5. 风险点

- 当前工作区里仍存在 `.agents/*`、`docs/superpowers/*` 等并存脏改动；它们不影响 workspace 主线，但后续提交时需要按变更归属拆分。
- 旧 team API 与 `current_tenant_id` 兼容字段仍会长期存在一段时间，后续若继续收尾，需要把这些桥接点和真正授权上下文严格区分。

### 6. 下一阶段建议

- 进入 Phase 9/10 的测试与清理收尾，只保留已经显式标记的 tenant 兼容桥接点。
- 后续若准备拆提交或开 PR，需要把非 workspace 主线的并存改动单独归类，避免影响权限迁移变更集。

## Phase 9 收口

### 1. 已完成

- 后端测试命令已通过：
  - `go test ./internal/modules/system/auth ./internal/modules/system/featurepackage ./internal/modules/system/permission ./internal/modules/system/tenant ./internal/modules/system/user ./internal/modules/system/page ./internal/modules/system/space ./internal/modules/system/system ./internal/pkg/authorization ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/pkg/teamboundary ./internal/api/router`
- 前端测试命令已通过：
  - `pnpm --dir frontend lint`
  - `pnpm --dir frontend build`
- 关键场景已按通过性检查核对：
  - `personal workspace` 平台角色 / 平台功能包
  - `team workspace` 团队角色 / 团队功能包
  - workspace 切换
  - menu-space 切换
  - 消息域团队视图兼容桥接

### 2. 风险点

- 仍有少量 legacy persistence 与 `/api/v1/tenants/*` route 保留兼容，但它们已不再代表当前主语义。

### 3. 下一阶段建议

- 若继续清理，优先收边界，不再扩功能：仅处理 legacy route、legacy header 和 legacy DTO 的缩容。

## Phase 10 收口

### 1. 已完成

- `tenant` 已在文档和活跃契约中明确回收到“未来多租户系统预留命名空间”。
- 当前主契约已切到 `workspace / team workspace / personal workspace`：
  - `X-Auth-Workspace-Id`
  - `X-Team-Workspace-Id`
  - `current_team_workspace_id`
  - `legacy_team_id`
- `X-Tenant-ID`、`current_tenant_id`、`source_tenant_id` 等 tenant 命名字段已降级为兼容输入或历史映射字段。

### 2. 允许继续保留的 legacy 组件

- legacy route：`/api/v1/tenants/*`
- legacy persistence：旧 tenant 表、旧成员表、旧 role/package 镜像关系
- legacy bridge：
  - `frontend/src/store/modules/tenant.ts`
  - `frontend/src/utils/http/index.ts`
  - `frontend/src/views/message/modules/useMessageWorkspace.ts`

### 3. 不再允许继续扩散的 legacy contract

- `X-Tenant-ID`
- `current_tenant_id`
- 页面层 `currentTenantId`
- 新增 DTO / header / context key 中继续用 tenant 表达当前 team/workspace 语义
