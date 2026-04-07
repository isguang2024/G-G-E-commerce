# Change Log

## 2026-04-07 删除技能锁文件

### 本次改动
- 删除根目录 [skills-lock.json](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/skills-lock.json)，该文件仅记录了 `shadcn/ui` 的技能锁缓存，而当前项目并不使用 shadcn 体系。
- 这一步是前面清理本地技能目录后的进一步瘦身，避免仓库继续保留无用的技能缓存文件。
- 本轮未修改业务代码，也未执行构建或测试。

### 下次方向
- 如果后续重新接入 shadcn 或其他技能锁，再按需生成即可；当前状态下无需恢复。
- 也可以继续检查根目录是否还有类似的临时缓存或一次性调试文件需要清掉。

## 2026-04-07 简化项目框架文档

### 本次改动
- 删除 [PROJECT_FRAMEWORK.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/PROJECT_FRAMEWORK.md) 中的“当前前端实施约束”整段，只保留项目主框架和实施顺序。
- 这次收口的目的，是把重复于 [FRONTEND_GUIDELINE.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/FRONTEND_GUIDELINE.md) 的前端约束描述移除，减少文档重叠。
- 本轮仅调整文档，不修改业务代码，也未执行构建或测试。

### 下次方向
- 如果还要继续简化，可以再检查 `PROJECT_FRAMEWORK.md` 和 `FRONTEND_GUIDELINE.md` 的职责边界，看看是否还能再合并一小部分表述。
- 也可以把 `docs/change-log.md` 再压缩成更短的摘要结构，降低历史记录的体量。

## 2026-04-07 清空失效本地技能目录

### 本次改动
- 清理空目录 [`.claude/skills`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.claude/skills) 和 [`.agents/skills`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills)，仅保留还在用的 `.claude/settings.local.json` 之类本地配置。
- 这一步是前面删除仓库内重复技能副本后的收尾，避免目录层面继续残留已失效的技能入口。
- 本轮未修改业务代码，也未执行构建或测试。

### 下次方向
- 如果后续确认不再使用任何本地技能接入，可以继续评估 `.claude/` 和 `.agents/` 下是否还有其他可收口的配置。
- 也可以保持当前状态，继续完全依赖全局技能版本。

## 2026-04-07 删除 superpowers 接入说明

### 本次改动
- 删除 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md)，不再把 superpowers 接入说明作为当前有效文档。
- 同步从 [docs/project-structure.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/project-structure.md) 的“当前有效文档”列表中移除该入口，避免后续查阅时继续把它当作当前规范。
- 本轮仅调整文档，不修改业务代码，也未执行构建或测试。

### 下次方向
- 如果后续确认仓库内不再使用任何 superpowers 相关过程文档，可以进一步检查 `.claude/`、`.agents/` 里是否还保留需要清掉的历史说明。
- 也可以继续收口 `docs/change-log.md` 的体量，把历史记录再做一次摘要化处理。

## 2026-04-07 仓库内重复技能清理

### 本次改动
- 删除仓库内重复的本地技能副本：[`.claude/skills/change-wrapup`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.claude/skills/change-wrapup)、[`.claude/skills/fluent-react-v9`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.claude/skills/fluent-react-v9)、[`.claude/skills/fluent2-frontend-style`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.claude/skills/fluent2-frontend-style)。
- 保留全局技能版本，不再让仓库内维护同名重复内容；仓库只保留 `settings.local.json` 等本地配置，不再保留这些技能快照。
- 本轮未执行构建或测试，改动范围仅限技能目录清理。

### 下次方向
- 如果后续还想进一步瘦身，可以检查 `docs/superpowers-integration.md` 是否还需要保留当前这么多流程说明。
- 也可以继续评估是否需要在仓库内单独保留任何技能快照，默认建议继续只依赖全局技能。

## 2026-04-07 docs 当前态结构重写

### 本次改动
- 重写 [docs/project-structure.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/project-structure.md)，改成只描述当前有效结构的索引文档，按权限、用户、空间、菜单、APP、功能包等主题重新分层。
- 删除 [docs/workspace-glossary.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-glossary.md)、[docs/workspace-permission-migration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-migration.md)、[docs/workspace-permission-stage-log.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-stage-log.md)，不再保留这组三份历史语义文档。
- 本轮仅调整文档结构，没有修改业务代码，也没有执行构建或测试。

### 下次方向
- 如果还要继续收口，可以把 `docs/change-log.md` 再压缩成月度摘要，或者把它从当前有效文档索引里降级为纯历史记录。
- 也可以继续检查 `docs/superpowers-integration.md` 是否还能进一步短化，避免文档入口过多。

## 2026-04-07 docs 权限文档收口

### 本次改动
- 压缩 [docs/workspace-glossary.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-glossary.md)、[docs/workspace-permission-migration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-migration.md)、[docs/workspace-permission-stage-log.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-stage-log.md) 的重复内容，三者分别只保留术语、唯一迁移基线和阶段进度。
- 去掉三份文档之间反复出现的规则解释和禁止项展开，改为互相引用，降低后续查阅成本。
- 本轮仅调整 Markdown 内容，未执行构建或测试。

### 下次方向
- 如果还要继续瘦身，可以把 `workspace-permission-stage-log.md` 再并入 `change-log.md`，让 `docs/` 里只留术语表和迁移说明。
- 也可以继续检查 `superpowers-integration.md` 是否还能再压缩掉一层流程性描述。

## 2026-04-07 项目结构文档简化

### 本次改动
- 新增 [docs/project-structure.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/project-structure.md)，把项目结构收敛为根目录、后端、前端正式页面、公共组件和目录边界五部分，只保留当前有效主线。
- 新文档移除了历史迁移叙述、旧目录背景和大段术语展开，改为按“后端入口 - 前端入口 - 正式页面 - 边界约束”组织，便于后续修改时快速定位。
- 当前未同步执行构建或测试，改动仅限 Markdown 文档。

### 下次方向
- 如果还要继续清理，可以把 `docs/workspace-glossary.md`、`docs/workspace-permission-migration.md`、`docs/workspace-permission-stage-log.md` 进一步压缩成更短的术语与迁移索引。
- 也可以在根目录补一个更短的 `README`，把“去哪里改后端、去哪里改前端页面”再做一层入口导航。

## 2026-04-07 空间权限最终态代码回写

### 本次改动
- backend 继续清理 active 运行时里的旧 `team / tenant` 业务语义：`permissionrefresh`、`featurepackage`、`authorization`、`collaborationworkspaceboundary`、`collaborationworkspace` 等主链只保留 `workspace / personal / collaboration` 的 canonical 方法和上下文字段，旧 `RefreshTeam / GetMyTeam / ListTenantRoles / tenantToMap / requireTargetTenant` 等主语义入口已不再出现在 active 主链扫描结果里。
- `backend/cmd/migrate/main.go` 已收成当前本地空项目最终态路径：迁移入口不再执行历史 rename/backfill 兼容链，并补齐“直接使用最终 workspace schema”的收口说明；协作空间相关 seed、权限刷新和边界快照逻辑继续对齐到 `personal / collaboration`。
- frontend 继续完成目录级和 public type 收口：`team.ts` 已删除，`views/team/**` 与 `system/team-roles-permissions/**` 已让位给 `collaboration-workspace/**` 与 `system/collaboration-workspace-roles-permissions/**`；`feature-package-teams-dialog.vue` 已物理改名为 `feature-package-collaboration-workspaces-dialog.vue`，active import 全部切到新路径。
- frontend 页面和类型继续去掉最后一批 `team/platform` 作为空间语义的表达：`AppContextBadge`、`message-dispatch-console`、`access-trace`、`feature-package`、`user-permission-test-drawer`、`collaboration-workspace` 页面主变量和提示文案统一改成“个人空间 / 协作空间 / 空间权限”；`api.d.ts`、`system-manage.ts` 和 `collaboration-workspace.ts` 不再暴露 `blockedByTeam / DispatchTeamOption / owner_tenant_name / tenant_name` 这类 public type 或 normalize 结果。
- 本轮对 active 代码再次做了 grep 收口，`backend` 和 `frontend/src` 中已不再命中以下高价值旧主契约或旧主语义：`/api/v1/tenants/*`、`X-Tenant-ID`、`X-Team-Workspace-Id`、`@/api/team`、页面层 `tenantStore`、`blockedByTeam`、`DispatchTeamOption`、`showTargetTeams`、`onlyTeamUsers`、`平台上下文 / 团队上下文 / 平台权限 / 团队权限`。
- 本轮已执行并通过：
  - `go test ./... -run '^$'`
  - `pnpm --dir frontend lint`
  - `pnpm --dir frontend build`

### 下次方向
- 如果还要继续抠到更高洁净度，优先处理模型层和数据库层仍保留的 `Tenant / TenantMember` 结构名，以及极少量历史注释、README 和作者标识里的 `team / tenant` 字样；这些已不影响当前运行时主语义和对外契约。
- 未来如果要真正引入多租户 `tenant` 领域，应另起独立模型、schema 和 API，不再回头复用当前协作空间实现。

## 2026-04-07 空间权限语义最终收口

### 本次改动
- 后端运行时继续硬切到 `workspace / personal / collaboration` 主语义：`role/service` 删除了协作空间角色上的旧 `ErrTenantRoleManagedByTeam / ErrTeamRoleKeyReadonly` 引用，`api/errcode` 将 `ErrCollaborationWorkspace*` 和 `ErrNoCurrentCollaborationWorkspace` 设为唯一主错误码，旧 `ErrTenant* / ErrNoTeam / ErrTeamRoleNotFound` alias 已从 active 源码中移除。
- `appscope`、`featurepackage`、`user/repository` 已把协作空间功能包主操作名统一成 `ReplaceCollaborationWorkspacePackagesInApp / ReplaceCollaborationWorkspacePackages`，不再以 `ReplaceTeamPackages*` 作为 active 主接口；`featurepackage` 的上下文归一化只接受 `personal | collaboration | common`。
- `permission/service` 修正了空间上下文判断：`normalizeContextType` 不再接受旧 `platform/team` 作为主上下文，`deriveContextType` 不再把 `collaboration_workspace.*` 错归到个人空间，`deriveModuleContextBoundary` 与 `platformroleaccess` 的上下文判断也统一到了 `personal / collaboration`。
- `collaborationworkspace/handler`、`user/handler`、`message_service`、`role/handler`、`apiendpoint/permission_audit` 等运行时代码已继续清掉旧 `my-team / team context / blocked_by_team / Refresh team` 等对外语义，协作空间诊断输出主字段已统一为 `blocked_by_collaboration_workspace`，默认上下文值也改成 `collaboration`。
- 前端 public type 和 normalize 继续收口：`system-manage.ts` 不再读取 `teamMember / blocked_by_team` fallback，权限诊断、角色功能包对话框、消息模板、菜单表单、角色编辑表单等活跃页面与组件统一改成 `personal / collaboration` 语义，示例占位文案和上下文筛选值不再使用 `team`。
- active 代码扫描结果已清空以下旧主契约与旧顶层术语：`/api/v1/tenants/*`、`X-Tenant-ID`、`X-Team-Workspace-Id`、`current_tenant_id / target_tenant_id / source_tenant_id`、`平台上下文 / 团队上下文 / 平台权限 / 团队权限`、`@/api/team`、页面层 `tenantStore`。
- 本轮已执行并通过：
  - `Set-Location backend; go test ./... -run '^$'`
  - `pnpm --dir frontend lint`
  - `pnpm --dir frontend build`

### 下次方向
- 如果继续追求代码洁净度，优先处理测试名、README、作者注释、图标名和少量局部变量中的 `team` 字样；这些已经不影响当前运行时语义和对外契约。
- 如果后续真的引入多租户 `tenant` 领域，应从零设计独立 schema、DTO 和 API，不再复用当前协作空间模型。

## 2026-04-07 空间权限语义代码回写

### 本次改动
- 后端权限诊断和角色功能包校验已改成“个人空间 / 协作空间 / 空间权限”语义：`role/service` 不再返回“平台上下文”文案，`user/handler` 的权限诊断上下文类型已从 `platform` 收口为 `personal`，并统一使用“当前个人空间下未生效此权限”“仅支持绑定个人空间功能包”“刷新个人空间权限快照失败”等表述。
- 前端空间上下文适配层已开始脱离 `platform = personal` 的旧表达：`frontend/src/store/modules/collaboration-workspace.ts` 现在以 `personal | collaboration` 作为上下文模式，新增 `hasPersonalWorkspaceAccessByUserInfo`、`hasPersonalWorkspaceAccess`、`setPersonalWorkspaceAccess`、`enterPersonalWorkspaceContext` 等主语义接口，同时保留旧别名仅作兼容。
- 前端权限诊断、用户菜单裁剪、角色功能包、用户功能包、工作台等核心页面已统一改写为“个人空间 / 协作空间 / 空间权限”文案，不再把“平台上下文 / 团队上下文”当成一级模型。
- 用户可见的角色、功能包和快捷入口说明已同步收口：个人空间角色、个人空间功能包、个人空间菜单裁剪等说法替换了旧的“平台角色 / 平台功能包 / 平台用户菜单”表述；权限键元数据中的功能包管理名称和说明也已同步改成个人空间语义。
- 本轮已执行并通过：
  - `Set-Location backend; go test ./... -run '^$'`
  - `pnpm --dir frontend lint`
  - `pnpm --dir frontend build`

### 下次方向
- 继续清 backend/frontend 活跃实现中的 `tenant/team` 内部命名，尤其是 `RefreshPlatform* / RefreshTeam* / supportsTeam / TeamList` 这类历史实现名，进一步贴近 `workspace / personal / collaboration` 最终模型。
- 继续收口迁移、模型、种子和 public type，把仍然承载旧兼容语义的字段和别名降到最薄。

## 2026-04-07 空间权限术语基线重置

### 本次改动
- 重写 `docs/workspace-permission-migration.md`，将当前目标语义正式收口为：`workspace` 是唯一权限主体，`workspace_type = personal | collaboration`，运行时权限上下文固定为 `auth_workspace_id + auth_workspace_type`。
- 重写 `docs/workspace-glossary.md`，明确 `personal workspace` 是个人空间、`collaboration workspace` 是协作空间，`platform` 只是业务域 / app，`tenant` 只保留为未来多租户保留名词。
- 重写 `docs/workspace-permission-stage-log.md`，把当前阶段改成“文档先定死术语，后续再回写前后端、迁移、种子和 API”，不再把旧的“平台上下文 / 团队上下文”叙述继续当作有效基线。
- 明确本轮不再接受以下旧主语义继续扩散：`workspace_type = team`、`X-Tenant-ID`、`/api/v1/tenants/*`、`current_tenant_id`、`平台权限 / 团队权限` 顶层术语。

### 下次方向
- 按文档基线继续回写后端运行时、迁移、模型、种子和 API，把“平台上下文 / 团队上下文”彻底替换为“个人空间 / 协作空间 / 空间权限”。
- 同步回写前端页面、路由、文案、权限诊断和组件说明，确保活跃 UI 不再把 `personal` 误写成“平台上下文”。

## 2026-04-06 协作空间全量改名落地 - 后端迁移收口

### 本次改动
- 在 [backend/cmd/migrate/main.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/migrate/main.go) 中补齐协作空间预迁移，旧库升级会先原地重命名 `tenants / tenant_members / team_*` 相关表、列和索引，再进入 `AutoMigrate`；`fresh` 模式直接跳过旧 schema 兼容步骤，确保只产出新 schema。
- 将后端协作空间主模块迁到 [backend/internal/modules/system/collaborationworkspace](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/collaborationworkspace)，并把边界服务迁到 [backend/internal/pkg/collaborationworkspaceboundary](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/collaborationworkspaceboundary)，同步更新路由挂载和 import 路径。
- 收口协作空间上下文：`X-Auth-Workspace-Id` 与 `X-Collaboration-Workspace-Id` 作为主头部，`legacy_collaboration_workspace_id` 已从中间件和响应体移除，授权上下文统一读取 `collaboration_workspace_id`。
- 同步清理了协作空间模块内对外日志和错误文案中的 `Tenant / team` 残留，避免新主线继续暴露旧命名。
- 已执行 `go test ./... -run '^$'`，结果通过。

### 下次方向
- 继续收敛 `tenantID / teamID` 这类内部参数名和辅助函数名，逐步把残留旧命名从实现细节中清掉。
- 如果后续继续做 schema 迁移，优先补一轮旧库回放验证，确认重命名表、列、索引在真实历史库上都能稳定执行。

## 2026-04-06 tenant 语义回收为未来多租户，当前主线切到协作空间

### 本次改动
- 完成当前业务主语义切换：`workspace_type` 当前有效值统一为 `personal | collaboration`，中文统一为“个人空间 / 协作空间”，`tenant` 正式退出当前协作、权限、导航、消息主线，保留给未来真正的多租户系统。
- 完成数据库与迁移链的正式收口：协作空间核心表、活跃列、索引和枚举值已切到 collaboration 命名，并在 [backend/cmd/migrate/main.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/migrate/main.go) 中通过预重命名 + value backfill 保证旧库升级和 `fresh` 初始化都直接落在新 schema 上。
- 完成后端主契约收口：当前主请求头为 `X-Auth-Workspace-Id` 与 `X-Collaboration-Workspace-Id`，主接口为 `/api/v1/workspaces/*` 与 `/api/v1/collaboration-workspaces/*`，平台能力通过 `personal workspace` 生效，协作能力通过 `collaboration workspace` 生效。
- 完成前端主语义收口：`workspaceStore` 仍是唯一授权上下文源，协作空间适配层与消息域 helper 已切到 `currentCollaborationWorkspaceId / collaborationWorkspaceList / loadMyCollaborationWorkspaces` 等命名，不再继续把 `currentTeam / teamList / loadMyTeams` 当成当前业务主语义。
- 补齐 canonical 代码入口路径：新增 [frontend/src/api/collaboration-workspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/collaboration-workspace.ts) 作为前端协作空间 API 主入口，新增 [backend/internal/modules/system/collaborationworkspace/module.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/collaborationworkspace/module.go) 作为后端协作空间模块主入口；active imports 已不再直接依赖旧 `@/api/team` 与 `internal/modules/system/tenant` 路径。
- 继续清理活跃页面和内部变量里的历史 `team / tenant` 命名：协作空间管理页及其成员、菜单、功能包、权限对话框统一改成 `collaborationWorkspaceName / collaborationWorkspaceId` 语义；消息域请求选项补齐 `skipCollaborationWorkspaceHeader`，权限诊断返回结构补齐 `collaborationWorkspaceMember / collaborationWorkspacePackages` 主字段，同时保留 `team*` 兼容别名；当前 `frontend` 的 `lint` 与 `build` 以及 `backend` 的空测编译均已恢复绿色。
- 重写 [docs/workspace-glossary.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-glossary.md)、[docs/workspace-permission-migration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-migration.md)、[docs/workspace-permission-stage-log.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-stage-log.md)，把最终状态收口成：当前主线使用 `workspace / personal workspace / collaboration workspace`，`tenant` 仅作为未来多租户保留名词。
- 本轮已执行并通过：
  - `go test ./... -run '^$'`
  - `pnpm --dir frontend lint`
  - `pnpm --dir frontend build`

### 下次方向
- 若继续推进，优先清理代码目录、内部变量和历史模块路径中残留的 `team / tenant` 命名，但这已经不影响当前运行时契约与对外语义。
- 若未来引入真正的多租户系统，应新建独立 `tenant` 领域设计和 schema，不再复用当前协作空间模型。

## 2026-04-06 workspace 权限迁移继续推进七

### 本次改动
- 清理 [frontend](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend) 下的全量前端 lint 存量债，覆盖权限组件、系统管理页、协作空间页、workspace 公共壳层、路由与 store，重点收掉了 `prettier/prettier`、`no-unused-vars`、`prefer-as-const`、`no-irregular-whitespace` 等规则问题。
- 更新 [frontend/src/views/system/api-endpoint/index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/api-endpoint/index.vue)、[frontend/src/views/system/menu-space/index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/menu/index.vue)、[frontend/src/views/system/page/index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/page/index.vue)、[frontend/src/views/system/user/modules/user-permission-test-drawer.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/user/modules/user-permission-test-drawer.vue)、[frontend/src/views/team/team/modules/team-menu-permission-dialog.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/team/team/modules/team-menu-permission-dialog.vue) 与 [frontend/vite.config.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/vite.config.ts)，移除已失效的 `route` / `router` / `handleSyncPages` / `handleRefresh` / `VITE_API_URL` 等未用变量，并修正系统页中的不规则空白和类型断言写法。
- 更新 [frontend/src/router/core/MenuProcessor.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/core/MenuProcessor.ts)，修正 `logPathError` 签名收窄后的调用点，避免 lint 通过但构建仍因参数不匹配失败。
- 本轮已执行 `pnpm --dir frontend lint` 与 `pnpm --dir frontend build`，均通过；前端正式验收项已从“仅 build”恢复为“lint + build”双门槛。

### 下次方向
- 进入最后一轮全链路排查，确认 platform/team 角色、workspace 切换、menu-space 切换和消息域协作空间兼容派生之间没有残余语义偏差。
- 梳理并标记最后保留的 tenant 兼容桥接点，只保留消息域 helper 与旧 team API 请求头桥接，不再让新页面重新依赖 tenant-only 语义。

## 2026-04-06 workspace 权限迁移继续推进六

### 本次改动
- 更新 [index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/team/team-members/index.vue)，把协作空间成员页的刷新监听从 `currentCollaborationWorkspaceId` 改成 `currentAuthWorkspaceId/currentAuthWorkspaceType`，避免该页面继续把 tenant 兼容值当作页面级上下文触发源。
- 补跑 `pnpm --dir frontend build` 与 `go test ./internal/modules/system/system ./internal/modules/system/user ./internal/modules/system/page ./internal/modules/system/space ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/pkg/authorization ./internal/api/router`，均通过。
- 本轮复查后，前端页面层面已没有直接消费 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 的点位；仅 [useMessageWorkspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/useMessageWorkspace.ts) 内部保留该兼容派生，用于旧 team API 请求头和协作空间视图桥接。

### 下次方向
- 下一步可以转入最后一轮全链路排查，确认 platform/team 角色、workspace 切换和 menu-space 切换之间没有残余语义漂移。
- 若要补齐前端最终验收，仍需单独修复 ESLint 的 `minimatch / brace-expansion` 环境问题。

## 2026-04-06 workspace 权限迁移继续推进五

### 本次改动
- 更新 [useMessageWorkspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/useMessageWorkspace.ts)，把 `currentCollaborationWorkspaceId` 明确收进消息域 workspace helper，协作空间消息页不再直接把 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 当成独立上下文源使用。
- 更新 [message-dispatch-console.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-dispatch-console.vue)、[message-recipient-group-console.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-recipient-group-console.vue)、[message-sender-console.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-sender-console.vue)、[message-template-console.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-template-console.vue)，把协作空间消息页中的上下文说明统一改成“当前授权工作空间 + 当前协作空间视图”，并把协作空间 ID 的兼容派生集中到消息域 helper。
- 更新 [AppContextBadge.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/business/layout/AppContextBadge.vue)，让协作空间模式下的兜底文案改成 `未启用协作空间视图`，避免继续把“未选择协作空间”当作主上下文提示。
- 更新 [service.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/page/service.go)，补齐平台角色 legacy fallback 的 `roles.deleted_at IS NULL` 过滤，和其它 platform fallback 规则保持一致。
- 本轮已执行 `go test ./internal/modules/system/system ./internal/modules/system/user ./internal/modules/system/page ./internal/modules/system/space ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/pkg/authorization ./internal/api/router` 与 `pnpm --dir frontend build`，均通过；`pnpm --dir frontend lint` 仍被现有 `minimatch / brace-expansion` 依赖异常阻塞。

### 下次方向
- 继续收消息域之外仍直接消费 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 的协作空间页，把 tenant 兼容派生值进一步压回边界层。
- 若要把前端验收补齐到 lint，下一步单独修复 ESLint 依赖环境，不和业务改动混在一起。

## 2026-04-06 workspace 权限迁移继续推进四

### 本次改动
- 更新 [runtime_cache.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/page/runtime_cache.go)、[util.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/space/util.go) 与 [repository.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/repository.go)，把协作空间 runtime 与平台筛选中的 legacy 角色回退进一步收紧到 `workspace` 优先、legacy 回退，并补齐 `roles.deleted_at IS NULL`、`roles.collaboration_workspace_id IS NULL` 等边界，避免成员身份或协作空间角色误混入平台判权。
- 更新 [beforeEach.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/guards/beforeEach.ts)，让前端登录后和刷新用户信息时优先比对 `auth workspace`，不再只靠 `currentCollaborationWorkspaceId` 判断上下文是否漂移。
- 更新 [index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/core/layouts/art-notification/index.vue)、[index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/workspace/inbox/index.vue)、[index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/dashboard/console/index.vue)，把通知面板、收件箱和控制台的上下文表达统一成“当前授权工作空间 / 当前协作空间视图”，不再把“当前协作空间”当作唯一主上下文。
- 本轮已执行 `go test ./internal/modules/system/system ./internal/modules/system/user ./internal/modules/system/page ./internal/modules/system/space ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/pkg/authorization ./internal/api/router` 与 `pnpm --dir frontend build`，均通过；`pnpm --dir frontend lint` 仍被现有 `minimatch / brace-expansion` 依赖异常阻塞。

### 下次方向
- 继续把剩余平台治理和消息相关页面中直接依赖 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 的点位缩到协作空间兼容层内，避免公共层再扩散 tenant-only 语义。
- 若要把验收门槛补齐到 lint，需要先单独修复前端 ESLint 依赖环境，不要把该环境问题误判为本轮代码回归。

## 2026-04-06 workspace 平台代管协作空间目标校验收口

### 本次改动
- 新增 [target_workspace.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/target_workspace.go)，补齐 `RequirePersonalWorkspaceTargetWorkspace`、`RequirePersonalWorkspaceTargetTeam`、`RequirePersonalWorkspaceTargetTeams` 三个 helper，把“personal workspace 代管 team workspace”的精确校验从零散 handler 抽成统一能力。
- 更新 [authorization.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/authorization.go)，导出 [RespondAuthError](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/authorization.go) 供 handler 复用统一错误响应，同时保持之前对 coarse-grained legacy key 的收窄策略不回退。
- 更新 [handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/featurepackage/handler.go) 与 [module.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/featurepackage/module.go)，让 `GET/PUT /feature-packages/teams/:collaborationWorkspaceId` 和 `PUT /feature-packages/:id/teams` 在旧路由不变的前提下，统一先把 path/body 中的 team 目标解析成 team workspace，再校验当前 personal workspace 是否有权代管。
- 本轮已执行 `go test ./internal/pkg/authorization ./internal/modules/system/featurepackage ./internal/modules/system/tenant ./internal/modules/system/user ./internal/pkg/teamboundary ./internal/modules/system/permission ./internal/modules/system/auth ./internal/modules/system/workspace ./internal/api/router` 与 `pnpm build`，均通过。

### 下次方向
- 继续把 role、tenant 平台治理和其他带 `collaborationWorkspaceId / collaboration_workspace_ids` 的代管接口逐族接入同一套 helper，避免继续散落目标校验逻辑。
- 若后续需要支持比“当前用户必须是目标 team 成员”更宽的平台代理模型，不要直接放宽校验；应先单独定义 personal workspace 的代理数据域规则。

## 2026-04-06 workspace 权限迁移角色子线收口

### 本次改动
- 新增 [service.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/workspacerolebinding/service.go)，把 personal/team workspace 角色绑定的解析、创建和替换收成统一 helper，避免平台角色和协作空间角色继续散写多份 SQL。
- 更新 [service.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/service.go) 与 [repository.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/repository.go)，让 `/users/:id/roles` 对内优先绑定到目标用户的 `personal workspace`；旧 `user_roles(collaboration_workspace_id IS NULL)` 退化为兼容镜像和历史回读来源。
- 更新 [service.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/platformaccess/service.go) 与 [service.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/permissionrefresh/service.go)，让平台权限快照和平台角色刷新优先读取 personal workspace 角色绑定，而不是只看全局 `user_roles`。
- 更新 [handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/tenant/handler.go)，让协作空间成员角色接口统一走 `workspace_role_bindings(team workspace)`，并补充 `binding_workspace_id`、`binding_workspace_type`、`member_type` 返回字段，明确成员身份与权限角色分离。
- 更新 [authorization.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/authorization.go)，收窄 `explicit_target_workspace` 的统一兜底范围，避免误伤角色目录、功能包列表等非单目标 legacy 接口。
- 更新 [collaboration_workspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/collaboration_workspace.ts)、[system-manage.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/system-manage.ts)、[api.d.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/types/api/api.d.ts) 以及角色/协作空间成员相关页面文案，把“平台角色经个人工作空间生效、协作空间内部角色经协作空间生效、成员身份不等于权限角色”落到前端表达层。

### 下次方向
- 继续清理剩余仍直接从旧 `user_roles` 读取协作空间角色的零散链路，把 team runtime 全量收口到 workspace-aware helper。
- 把更多平台代管接口按“目标 team/workspace 操作”逐类接入精确校验，而不是继续靠粗粒度 legacy permission key 兜底。
- 在角色链稳定后补一轮回归验证，重点覆盖 personal/team workspace 切换下的平台角色与协作空间内部角色隔离行为。

## 2026-04-06 workspace 权限迁移 Phase 0 基线

### 本次改动
- 新增 [docs/workspace-permission-migration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-migration.md)，把 `tenant / workspace / app / menu-space` 的新旧语义、权限链和兼容原则固定为后续迁移基线。
- 新增 [docs/workspace-glossary.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-glossary.md)，明确 `workspace != menu-space`、平台后台权限来自 `personal workspace`，并禁止继续扩张账号级全局业务角色。
- 新增 [docs/workspace-permission-stage-log.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/workspace-permission-stage-log.md)，记录当前已完成 Phase 0，并补齐阶段收口模板、风险点和下一阶段建议。
- 本轮没有改后端或前端行为代码，没有新增表、字段或 DTO，也没有执行构建和测试。

### 下次方向
- 进入 Phase 1，新增 `workspace` 领域模型和兼容映射表，先把 personal workspace / team workspace 的数据基础补齐。
- Phase 1 仍应以兼容 tenant 为前提，不直接切运行时权限判断。

## 2026-04-06 workspace 权限迁移 Phase 1 数据基线

### 本次改动
- 新增 [workspace.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/models/workspace.go)，补齐 `Workspace`、`WorkspaceMember`、`WorkspaceRoleBinding`、`WorkspaceFeaturePackage` 四个领域模型。
- 新增 [service.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/workspace/service.go) 和 [module.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/workspace/module.go)，提供 workspace 回填、personal/team workspace 查询和成员映射基础能力。
- 更新 [database.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/database/database.go)，把四张 workspace 表接入 `AutoMigrate`，并补齐唯一索引。
- 更新 [main.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/migrate/main.go)，在默认迁移流程中加入 workspace baseline 回填，启动后自动创建 personal workspace、team workspace 和 workspace member 映射。
- 同步更新 [models.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/models.go)，让 user 领域别名暴露 workspace 相关模型。
- 本轮已执行 `go test ./internal/pkg/database ./internal/modules/system/workspace ./cmd/migrate` 与 `go test ./internal/modules/system/auth ./internal/pkg/authorization ./internal/modules/system/workspace ./cmd/migrate`，均通过。

### 下次方向
- 继续 Phase 2，把 `auth_workspace_id`、`auth_workspace_type` 和兼容 `collaboration_workspace_id` 的运行时上下文稳定下来。
- 在不破坏旧 API 的前提下，逐步让请求从 workspace 主线进入后端判权链。

## 2026-04-06 workspace 权限迁移 Phase 2 与 Phase 5 推进

### 本次改动
- 更新 [middleware.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/auth/middleware.go)、[handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/auth/handler.go)、[module.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/auth/module.go)、[authorization.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/authorization.go)，让请求上下文优先解析 `X-Auth-Workspace-Id`，并在兼容期继续回落 `collaboration_workspace_id`。
- 新增 [handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/workspace/handler.go)，并更新 [module.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/workspace/module.go)、[router.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/api/router/router.go)，接入 `/api/v1/workspaces/my`、`/current`、`/:id`、`/switch` 四个基础接口。
- 更新 [handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/tenant/handler.go) 和 [module.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/tenant/module.go)，在旧 tenant 返回结果中补齐 `workspace_id`、`workspace_type`、`collaboration_workspace_id`，把 tenant 进一步降级为 workspace 兼容壳。
- 新增 [workspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/workspace.ts)，并更新 [collaboration_workspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/collaboration_workspace.ts) 与 [api.d.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/types/api/api.d.ts)，把前端契约补到 workspace 基础类型和响应字段。
- 本轮已执行 `go test ./internal/modules/system/workspace ./internal/modules/system/tenant ./internal/modules/system/auth ./internal/pkg/authorization ./internal/api/router` 与 `pnpm build`，均通过。

### 下次方向
- 继续 Phase 3 和 Phase 4，把 permission / feature package / role / platform-admin 入口正式迁到 workspace 主链。
- 进入 Phase 6 时优先补 `workspace.ts` store 和 HTTP 头透传，让 `workspaces/switch` 的结果真正进入运行时请求链。

## 2026-04-06 workspace 权限迁移 Phase 3 与 Phase 4 收口推进

### 本次改动
- 更新 [model.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/models/model.go)、[service.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/permission/service.go)、[handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/permission/handler.go)，让 `permission_keys` 显式承载 `app_key`、`data_policy`、`allowed_workspace_types`，并在创建、更新、返回权限动作时同步透出这些元数据。
- 新增 [context.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/context.go)，并更新 [authorization.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/authorization.go)，补齐统一 `AuthorizationContext`，把运行时判权入口改成优先基于 `auth_workspace_id / auth_workspace_type / target_workspace_id / app_key` 做决策。
- 更新 [handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/auth/handler.go)，让 `GetUserInfo` 在当前授权 workspace 和当前 App 下返回动作快照，不再只按 legacy `collaboration_workspace_id` 兜底。
- 更新 [handler.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/handler.go)、[system-manage.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/system-manage.ts)、[api.d.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/types/api/api.d.ts)，把新的权限元数据带到前端契约层，便于后续平台后台和 team 业务界面显式区分 workspace 语义。
- 本轮已执行 `go test ./internal/pkg/authorization ./internal/modules/system/auth ./internal/modules/system/permission ./internal/modules/system/user`、`go test ./internal/modules/system/workspace ./internal/modules/system/tenant ./internal/modules/system/auth ./internal/pkg/authorization ./internal/api/router` 与 `pnpm build`，均通过。

### 下次方向
- 继续 Phase 3，把 feature package / role 的运行时来源切到 `workspace_feature_packages / workspace_role_bindings`，收掉 team 业务权限对 tenant 兼容链的依赖。
- 继续 Phase 4，把 `explicit_target_workspace` 从权限元数据推进到平台代管接口的实际请求校验，避免 personal workspace 越权代管。
- 进入 Phase 6 时补前端 `workspace.ts` store 和 `X-Auth-Workspace-Id` 透传，让当前授权 workspace 从接口契约真正进入前端运行时。

## 2026-04-06 workspace 权限迁移 Phase 6 前端上下文收口

### 本次改动
- 新增 [workspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/store/modules/workspace.ts)，把“我的 workspaces、当前授权 workspace、personal/team workspace 列表”的状态从旧 tenant 语义里独立出来。
- 更新 [collaboration_workspace.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/store/modules/collaboration_workspace.ts)，让它退化为 team 兼容代理层：`currentContextMode / currentCollaborationWorkspaceId / currentTeam / loadMyTeams` 仍保留给旧页面使用，但真实授权来源改由 `workspaceStore` 决定。
- 更新 [index.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/utils/http/index.ts)，让受控请求统一透传 `X-Auth-Workspace-Id`；legacy team 请求继续保留 `X-Collaboration-Workspace-Id` 兼容头。
- 更新 [beforeEach.ts](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/guards/beforeEach.ts) 和 [index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/auth/login/index.vue)，让登录初始化与用户信息刷新阶段同步加载 workspace 上下文，避免刷新后退回旧 tenant-only 语义。
- 更新 [AppContextBadge.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/business/layout/AppContextBadge.vue) 和 [index.vue](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/dashboard/console/index.vue)，把头部和控制台摘要文案开始切到“个人工作空间 / 协作空间”语义。
- 本轮已执行 `pnpm build`，通过。

### 下次方向
- 继续 Phase 6，把头部切换器和用户菜单升级成统一的 workspace 切换面板，而不是只暴露协作空间列表。
- 继续 Phase 3 / 4，把 feature package、role、平台代管 handler 的运行时收口补完，缩短前端继续依赖 tenant 兼容链的窗口。
- 后续如果继续收尾，优先审一轮仍直接消费 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 的业务页，逐步把它们改成显式 workspace 语义。

## 2026-04-06 brainstorming 触发进一步软化

### 本次改动
- 继续收轻 [brainstorming/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/brainstorming/SKILL.md)，把触发条件从偏固定的“跨模块 / 跨前后端 / 菜单权限 App 模型 / 复杂页面重组”改成“有明显设计分歧、边界变化或需要先选方案时才用”。
- 同步更新 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md)，把仓库级触发说明改成“有多个合理方案、或者会影响信息架构 / 边界语义时再用”，避免把所有跨模块工作都硬塞进 brainstorming。
- `writing-plans` 保持当前轻量版，不再继续加硬度；当前它只在真正需要顺序编排和多步执行的场景触发。
- 本轮没有改业务代码，也没有执行构建或测试。

### 下次方向
- 如果后续还想继续减轻流程，下一步优先考虑是否再缩小 `writing-plans` 的触发范围，或者直接停用它。
- 如果你希望，我也可以把 `brainstorming` 再压成“只有明确要先设计时才用”的更极简版本。

## 2026-04-06 superpowers 根因分析技能停用

### 本次改动
- 停用了 [systematic-debugging](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/systematic-debugging) 的技能发现入口，把“先做根因分析”的独立流程从当前仓库启用清单中移除。
- 停用方式与前几项一致，仍然是把目录内 `SKILL.md` 改名为 `SKILL.disabled.md`，保留内容以便以后恢复，不影响其他技能。
- 同步更新 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md)，把 `systematic-debugging` 从启用列表和主流程里移除，改成只保留最小复现和定位的自然工作方式。
- 本轮没有改业务代码，也没有执行构建或测试。

### 下次方向
- 如果后续希望恢复系统化排障，再把 `systematic-debugging` 放回启用入口；当前仓库先按更轻量的直接定位+最小复核流程跑。
- 如果还想继续减轻流程，下一步只剩 `brainstorming` 和 `writing-plans` 是最值得再评估的入口。

## 2026-04-06 superpowers 完成前验证技能停用

### 本次改动
- 停用了 [verification-before-completion](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/verification-before-completion) 的技能发现入口，把“完成前必须先验证”的独立流程从当前仓库启用清单中移除。
- 停用方式与前几项一致，仍然是把目录内 `SKILL.md` 改名为 `SKILL.disabled.md`，保留内容以便以后恢复，不影响其他技能。
- 同步更新 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md) 和 [下次方向.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/下次方向.md)，把完成前验证从启用清单和后续建议里移除。
- 本轮没有改业务代码，也没有执行构建或测试。

### 下次方向
- 如果后续希望重新加强交付门槛，再考虑恢复 `verification-before-completion`，而不是在当前仓库里长期挂着一个不常用的硬流程。
- 若还想继续减轻流程，下一步优先考虑 `brainstorming` 或 `writing-plans` 是否也要继续缩小触发范围。

## 2026-04-06 superpowers 测试技能停用

### 本次改动
- 停用了 [test-driven-development](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/test-driven-development) 的技能发现入口，把它从当前仓库的启用清单中移除。
- 停用方式与之前几项一致，仍然是把目录内 `SKILL.md` 改名为 `SKILL.disabled.md`，保留内容以便以后恢复，不影响其他技能。
- 同步更新 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md) 和 [下次方向.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/下次方向.md)，把测试相关入口从当前启用清单和触发说明里移除。
- 本轮没有改业务代码，也没有执行构建或测试。

### 下次方向
- 如果后续前端测试体系成熟，再决定是否恢复 `test-driven-development`，不要现在保留一个长期不用的入口。
- 若还想继续减轻流程，下一步优先考虑 `brainstorming` 或 `writing-plans` 是否也要进一步缩减触发范围。

## 2026-04-06 superpowers 停用一批重流程技能

### 本次改动
- 直接停用了 [using-superpowers](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/using-superpowers)、[using-git-worktrees](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/using-git-worktrees)、[executing-plans](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/executing-plans)、[finishing-a-development-branch](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/finishing-a-development-branch)、[writing-skills](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/writing-skills) 的技能发现入口。
- 停用方式不是删除目录，而是把各目录下的 `SKILL.md` 改名为 `SKILL.disabled.md`，保留原始内容和恢复路径，同时让新会话不再发现这些技能。
- 更新 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md)，把当前仍启用的技能、已停用的技能，以及最新的仓库精简流程写清楚。
- 更新 [下次方向.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/下次方向.md)，把下一步可能继续停用 `brainstorming` / `writing-plans` 的方向记下来。
- 本轮没有改业务代码，也没有执行构建或测试。

### 下次方向
- 如果后续觉得流程还是偏重，优先考虑继续停用 `brainstorming` 或 `writing-plans`，把剩余技能进一步压缩。
- 如果未来又需要恢复某个停用技能，直接把对应目录里的 `SKILL.disabled.md` 改回 `SKILL.md` 即可。

## 2026-04-06 superpowers 子代理执行链与调试链继续收轻

### 本次改动
- 调整 [subagent-driven-development/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/subagent-driven-development/SKILL.md)，把原版“每个任务都默认实现 + 双 review + 最终分支收尾”的固定链路改成按任务风险选择 review 深度，并明确任务完成后默认停在“已验证结果”，不自动进入 branch finish。
- 更新 [implementer-prompt.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/subagent-driven-development/implementer-prompt.md) 与 [code-quality-reviewer-prompt.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/subagent-driven-development/code-quality-reviewer-prompt.md)，去掉实现子代理的默认 commit 要求，并把 code quality review 调整为“若当前任务路径包含 spec review，则在其通过后再进入”。
- 调整 [systematic-debugging/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/systematic-debugging/SKILL.md)，把“先写失败测试”收口为“先建立最强可行复现证据”：适合自动化回归时再走 failing test，不适合时至少保留可重复的最小复现和验证命令。
- 更新 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md) 与 [下次方向.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/下次方向.md)，补齐当前仓库对子代理 review 深度、commit 节奏和调试验证方式的本地化约束。
- 这轮仍然只改技能说明、模板和仓库文档，没有改业务代码，也未执行构建或测试。

### 下次方向
- 如果后续发现 `subagent-driven-development` 对小任务仍然偏重，可以继续把“单 reviewer”场景再细分成更轻的默认路径。
- 如果未来前端可自动化回归的页面越来越多，再评估是否把 `systematic-debugging` 里“失败测试优先”的适用范围向部分前端主链扩大。

## 2026-04-06 superpowers 计划链与收尾链继续收轻

### 本次改动
- 继续定制 [writing-plans/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/writing-plans/SKILL.md)，把“默认写成零上下文超细计划”的原版要求收口为“只在多步、多文件、跨模块、迁移或接口契约变化时写计划”，并把任务粒度从机械 2-5 分钟微步骤放宽到按仓库实际风险拆分。
- 调整 [executing-plans/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/executing-plans/SKILL.md)，明确它只在已有计划且需要顺序执行时触发；执行结束后默认停在“已验证的实现结果”，不再自动串到分支收尾技能。
- 调整 [finishing-a-development-branch/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/finishing-a-development-branch/SKILL.md)，把适用条件改成用户显式要求 commit / PR / merge / cleanup 时再进入，不再作为任何完成态的固定尾闸。
- 更新 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md)，把 `executing-plans` 的条件触发和 `finishing-a-development-branch` 的非默认化一并写入当前仓库精简流程。
- 本轮只修改本地 superpowers 技能说明与仓库文档，没有改业务代码，也未执行构建或测试。

### 下次方向
- 如果后续还觉得计划文档偏重，可以继续把 `writing-plans` 的任务模板再向“任务 + 验证 + 风险说明”压缩，减少不必要的示例代码块。
- 如果未来仓库的分支协作方式更固定，再决定是否为 `finishing-a-development-branch` 增加更贴近当前协作空间习惯的中文选项模板。

## 2026-04-06 superpowers 重流程入口继续收轻

### 本次改动
- 继续定制 [using-superpowers/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/using-superpowers/SKILL.md)，不再把它当作“每次开口前必须触发”的总入口，而是改成按任务收益判断是否启用 superpowers 工作流。
- 调整 [brainstorming/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/brainstorming/SKILL.md)、[test-driven-development/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/test-driven-development/SKILL.md)、[using-git-worktrees/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/using-git-worktrees/SKILL.md)、[requesting-code-review/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/requesting-code-review/SKILL.md)，把原版“默认强制”收口为“按仓库实际场景条件触发”。
- 在 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md) 明确记录当前仓库的精简流程：小改动直接实现并验证，中改动写简短执行思路，大改动再条件触发 brainstorming / plans / subagent，Bug 统一走 debugging，交付前再走 review 和 branch finish。
- 同步更新 [下次方向.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/下次方向.md)，后续重点观察收轻后的触发效果和前端测试能力是否变化。
- 本轮只修改本地 superpowers 技能说明与文档，没有改业务代码，也未执行构建或测试。

### 下次方向
- 如果后续发现 `writing-plans` 或 `executing-plans` 仍然过重，可以继续把它们的默认适用范围再收窄一层。
- 如果未来前端测试体系逐步成形，再决定是否把 `test-driven-development` 从“条件触发”重新提升到某些前端核心模块的默认流程。

## 2026-04-06 superpowers 子代理模型策略本地化

### 本次改动
- 调整 [subagent-driven-development/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/subagent-driven-development/SKILL.md) 的模型选择策略，明确当前仓库默认采用“轻模型优先、按复杂度逐级升级”的子代理策略。
- 为实现子代理、规格审查子代理、代码审查子代理分别补了默认档位：实现和规格审查默认走 `gpt-5.4-mini + low`，代码审查默认走 `gpt-5.4 + medium`，只在复杂集成、架构判断或高风险 review 时才升到更高档位。
- 更新 [dispatching-parallel-agents/SKILL.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/dispatching-parallel-agents/SKILL.md) 与三个 prompt template，避免并行代理和 reviewer 默认开高推理导致成本、时延成倍放大。
- 在 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md) 记录了当前仓库的子代理默认档位，方便后续新会话直接沿用。
- 本轮只调整本地技能说明和模板，没有改业务代码，也未执行构建或测试。

### 下次方向
- 后续如果还要继续收轻 `superpowers`，优先处理 `using-superpowers`、`brainstorming`、`using-git-worktrees` 这几个仍然偏重的硬流程入口。
- 如果未来发现 `gpt-5.4-mini + low` 对某些规格审查过轻，可以把规格审查默认档位提升到 `gpt-5.4 + medium`，但不建议直接回到全量高推理。

## 2026-04-05 shadcn 技能迁移到全局目录

### 本次改动
- 将仓库级 `shadcn` 技能从 [`.agents/skills/shadcn/`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/shadcn) 迁移到全局目录 [C:/Users/Administrator/.codex/skills/shadcn](C:/Users/Administrator/.codex/skills/shadcn)，后续其他仓库也可以直接复用同一套技能。
- 当前仓库内已删除项目级 `shadcn` 技能目录，避免项目级和全局级同时存在时出现重复发现或版本漂移。
- 迁移过程中曾因为并行执行“复制”和“删除”导致源目录先被移除，随后已改为通过 `git archive` 从仓库 `HEAD` 导出完整目录重建全局技能，确保二进制资源和文本文件都完整保留。
- 本轮未重启 Codex；全局技能在新会话或重启后生效。

### 下次方向
- 如后续仍需维护这套全局 `shadcn` 技能，建议以后直接在全局目录更新，避免再次在单仓库内保留副本。
- 若未来需要项目定制版 `shadcn` 规则，再单独在仓库内新增覆盖层，而不是回到“全局一份 + 项目一份同名副本”的结构。

## 2026-04-05 项目级接入 superpowers 技能集

### 本次改动
- 将 `obra/superpowers` 的技能快照以项目级方式接入到 [`.agents/skills/superpowers/`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers)，并一并保留上游 `LICENSE` 与 `agents/code-reviewer.md`，让当前仓库可以直接发现并使用这套流程技能。
- 更新 [AGENTS.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/AGENTS.md)，明确 `superpowers` 已接入当前仓库，且若技能流程与仓库级约束冲突，以仓库根文档为准，避免工作流反向覆盖当前前端主线规范。
- 新增 [docs/superpowers-integration.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/superpowers-integration.md) 与 [UPSTREAM.md](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/.agents/skills/superpowers/UPSTREAM.md)，记录上游来源、当前快照 commit、使用方式和更新办法。
- 同步移除本机全局 [AGENTS.md](C:/Users/Administrator/.codex/AGENTS.md) 中“最高智能、高速度、最少 token”约束，避免后续会话继续被该全局指令绑定。
- 本轮未执行业务代码构建；技能发现需要新开 Codex 会话或重启应用后生效。

### 下次方向
- 如果后续确定 `superpowers` 的默认强流程过重，可以继续在当前仓库内做二次裁剪，例如只保留 debugging、verification、review、parallel-agents 等高收益技能。
- 若要跟进上游版本，直接用新的 `skills/` 快照覆盖当前目录，并同步更新 `UPSTREAM.md` 的 commit hash。

## 2026-04-03 Vue 消息发送页空数组兜底修复

### 本次改动
- 修复 [message-dispatch-console.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-dispatch-console.vue) 在消息发送配置接口返回 `null` 数组时的渲染报错，新增 `normalizeDispatchOptions()` 统一将 `sender_options`、`template_options`、`audience_options`、`teams`、`users`、`recipient_groups`、`roles`、`feature_packages` 归一为 `[]`。
- 对模板筛选、顶部指标和发送按钮可用态增加空值兜底，避免再次出现 `filter` / `length` 读取 `null` 的运行时错误。
- 已通过 `pnpm --dir frontend build` 验证。

### 下次方向
- 检查消息模板、发送人、收件组等相邻页面是否也存在接口返回 `null` 数组但前端直接按数组读取的情况，统一在入口层归一。
- 若后端契约允许，后续可以把这些字段收紧为始终返回数组，减少前端到处做容错。

## 2026-04-03 frontend 路由整页刷新判定收口

### 本次改动
- 收紧 `frontend` 菜单空间跳转判定：同 host、同协议且当前 pathname 已经处在目标 `routePrefix` 下时，不再把路由跳转降级为整页 `location.assign(...)`。
- 将菜单点击统一接入菜单空间感知导航链路，侧栏和顶部菜单现在会结合 `spaceKey` 判断是走 `router.push(...)` 还是确实需要整页跳转。
- 修复快速入口基于 `routeName` 导航时误把 `resolved.href` 当内部路径下传的问题，避免 hash 路由场景拼出异常目标地址。
- 已通过 `pnpm --dir frontend build` 验证。

### 下次方向
- 若仍偶发整页刷新，可优先检查 `window.location.assign(...)` 的剩余入口：协作空间切换、用户菜单、通知中心、富文本内部链接和登录落地页。
- 当前已在菜单空间整页导航分支补了 dev warning，后续如要继续深挖，可再给这些剩余入口统一加来源标识，便于直接从控制台判断是哪条链路触发。

## 2026-04-03 调整 frontend 历史文档归档位置

### 本次改动
- 将 `frontend/docs/legacy/` 内的旧前端文档完整保留并迁移到仓库级归档目录 `docs/archive/frontend/`，避免当前 `frontend/docs/` 继续承担历史资料入口。
- 同步更新 `AGENTS.md` 中的归档位置说明，使当前规范与实际目录结构一致。
- 删除原 `frontend/docs/legacy/` 目录；归档文档内部的绝对链接前缀也已一并改到新位置。

### 下次方向
- 若后续要在 `frontend/` 内重新建立当前生效专题文档，可直接在 `frontend/docs/` 下按主题重建，不再混放旧资料。
- 本轮未逐篇审校归档文档内容，仅完成目录迁移和链接前缀修正；如需进一步压缩历史包袱，可继续筛掉低价值草稿。

## 2026-04-03 移除 frontend-shadcn React 重写线

### 本次改动
- 删除独立的 `frontend-shadcn/` React 重写工程及其根目录启动脚本 `start-frontend-shadcn.bat`，仓库前端主线收口回 `frontend/`。
- 更新 `AGENTS.md`、`PROJECT_FRAMEWORK.md` 与 `FRONTEND_GUIDELINE.md`，移除“双前端线 / React 重写线”约束，改为只保留 `frontend/` 作为当前有效管理端主线。
- 保留历史变更记录，不回写改动旧条目，避免破坏既有演进上下文。

### 下次方向
- 若后续仍需做管理端重构，建议直接在 `frontend/` 内按页面边界逐步演进，不再新开平行工程目录。
- 下一轮可继续清理仓库中残留的历史实验线与失效脚本，但需先确认它们是否仍被外部流程引用。
- 本轮未执行 `frontend` 的构建与运行验证，仅完成目录与文档收口。

# 2026-04-03 frontend-shadcn clean-slate 主题收口与角色页紧凑基线

### 本次改动
- 参考 [clean-slate](https://21st.dev/community/themes/clean-slate) 的中性灰度方向，重写 `frontend-shadcn/src/index.css` 的全局 token，将主背景、前景、边框、侧栏与主色统一收口到低饱和黑白灰体系，并同步补齐暗色 token。
- 收缩后台壳层密度：`app-shell` 顶栏、搜索框、快捷入口和主内容区的尺寸与留白整体下调，`page-header` 改为更轻的标题容器，减少大圆角、大面板和高噪声视觉块。
- 将 `system/role` 从多块大面板重构为“紧凑指标条 + 单主工作面 + 侧边详情”的基线布局，保留筛选、列表、详情和操作反馈，但去掉高饱和强调块和重复卡片包装。
- 已通过 `pnpm --dir frontend-shadcn build` 验证；本轮未补浏览器人工回归，当前仍有 Vite 默认的大 chunk 告警。

### 下次方向
- 继续沿这套紧凑基线改造 `system/user`，让用户页与角色页保持同一信息密度、筛选工具条和侧边详情语言。
- 若这套灰度 token 确认可用，再把列表页、详情页和表单页中的局部强调色继续收口，避免旧页面残留更强的品牌色块。
- 后续如需更贴近 `clean-slate`，优先细调字重、间距和控件半径，不直接再引入新的第三方整套 preset。

### 继续收口
- 进一步压低角色页顶部统计带的边界感，把原先的“白卡堆叠”再削薄一层，改成更轻的边框块。
- 主标题区和筛选区也继续收紧间距，优先保留信息层级，不增加新的装饰层。
- `pnpm --dir frontend-shadcn build` 仍然通过，当前只是继续做视觉密度收口，没有引入结构性风险。

### 筛选收口
- 将角色页顶部的作用域筛选和状态筛选从按钮组改为标准下拉框，横向占用更小，也更适合后续再加更多筛选项。
- 通过 shadcn 官方 `select` 组件承接筛选交互，保留现有筛选逻辑与默认值，不改数据层。
- 已再次通过 `pnpm --dir frontend-shadcn build` 验证。

### 布局再压
- 将筛选带继续压成更紧凑的三列布局，搜索框和两个下拉的横向节奏更一致，顶部占用更少。
- 右侧详情栏再收窄到 320px，进一步减弱左右并排大面板感，让主列表成为视觉重心。
- 已再次通过 `pnpm --dir frontend-shadcn build` 验证。

### 详情抽屉
- 角色详情不再以右侧常驻栏形式存在，改为点击列表行后从右侧弹出 `Sheet` 抽屉承接。
- 主页面现在只保留列表与筛选，右侧信息、概览、继承和治理内容都转入抽屉，页面本体更清爽。
- 已再次通过 `pnpm --dir frontend-shadcn build` 验证。

# 2026-04-03 frontend-shadcn 切换到 preset b5JgLt8Ce

### 本次改动
- 在 `frontend-shadcn/` 执行 `pnpm dlx shadcn@latest init --preset b5JgLt8Ce --force --reinstall`，将当前主题预设切换到 `b5JgLt8Ce`。
- `components.json` 随预设更新为 `style = radix-maia`、`iconLibrary = tabler`、`menuAccent = bold`，并同步重装覆盖现有 `ui` 基础件与 `src/index.css` 主题 token。
- 预设切换后新增 `@fontsource-variable/noto-serif`、`@tabler/icons-react` 依赖，已通过 `pnpm --dir frontend-shadcn build` 验证；当前仍有 Vite 大 chunk 告警，但不影响本轮构建通过。

### 下次方向
- 回看 `frontend-shadcn` 现有页面，确认新的 serif heading、Tabler 图标和更强的 menu accent 是否符合后台产品气质；若不合适，再局部收口 token 而不是立即回退整套 preset。
- 若后续继续新增官方组件，保持使用当前 preset 生成，避免新旧 `ui` 基础件混出两套视觉细节。

# 2026-04-03 frontend-shadcn 用户页与角色页深化

### 本次改动
- 深化 `system/user` 工作面：新增风险筛选、功能包规模列，以及“概览 / 访问 / 诊断”三段式详情区，右侧可直接承接账号风险信号、访问快照和后续权限诊断。
- 深化 `system/role` 工作面：新增状态筛选、继承规模列，以及“概览 / 继承 / 治理”三段式详情区，右侧统一承接角色继承来源、动作覆盖和治理摘要。
- 两页仍保持 `React Query -> admin.adapter` 数据边界不变，只加强页面组织和治理信息密度，不把页面重新绑回静态占位数据。
- 已通过 `pnpm --dir frontend-shadcn build` 验证；当前仍有 Vite 默认的大 chunk 告警。

### 下次方向
- 继续把用户页和角色页的右侧详情接到真实权限快照、菜单继承和动作差异接口，而不是继续停留在派生摘要。
- 若下一轮要继续深化治理页，优先推进 `system/access-trace`、`system/feature-package`，并抽用户/角色共用的详情 tabs 片段。

# 2026-04-03 frontend-shadcn 第二批治理页与 adapter 层收口

### 本次改动
- 新增 `admin.adapter` 稳定接口层，并把 `admin.service` 全部改为依赖 adapter，而不是直接引用 `admin.mock`，为后续切真实后端预留统一边界。
- 补齐 `system/page`、`system/api-endpoint`、`system/menu-space`、`system/action-permission` 四个真实工作面，统一提供筛选、列表、右侧详情和 toast 操作反馈。
- 扩充后台治理 mock 数据与 query keys，同步更新壳层路由分发和导航状态；这四页不再落回 `management-workspace-page`。
- 已通过 `pnpm --dir frontend-shadcn build` 验证；当前仍保留 Vite 默认的大 chunk 告警。

### 下次方向
- 继续把 `system/access-trace`、`system/feature-package`、消息目录类页面从通用骨架推进到真实工作面。
- 在 `admin.adapter` 上补真实后端实现或 adapter 工厂，再逐步替换当前 mock 默认实现。

# 2026-04-03 frontend-shadcn 首批治理页接入 query/mock 工作面

### 本次改动
- 为 `frontend-shadcn` 接入 `React Query`，新增 `shared/api`、`shared/mock` 与 `features/admin` 结构，把用户、角色、菜单、协作空间成员、消息调度的页面数据统一收口到 mock / adapter 链路。
- 新增 `system/user`、`system/role`、`system/menu`、`team/team-members`、`system/message` 五个真实工作面，补齐筛选、列表、详情区和操作反馈，不再继续停留在通用骨架页。
- 更新 `app-shell` 页面分发与 `frontend-shadcn/docs/architecture.md`，明确这条新线当前采用“页面工作面 + query 层 + mock module”的组织方式。
- 已通过 `pnpm --dir frontend-shadcn build` 验证；当前仍有 Vite 默认的大 chunk 告警，但不影响本轮交付。

### 下次方向
- 继续把 `system/page`、`system/api-endpoint`、`system/menu-space`、`system/action-permission` 等剩余治理页从通用骨架推进到真实工作面。
- 把当前 mock module 再抽成稳定 adapter 接口，逐步替换 `navigation.tsx` 中残留的静态占位数据。

# 2026-04-03 frontend-shadcn 新线起壳

### 本次改动
- 新建 `frontend-shadcn/`，以 React + Vite + shadcn/ui + 官方 `sidebar-07` blocks 作为新的管理端重写基座。
- 落地新的后台壳层：可折叠侧栏、顶栏、面包屑、快捷入口、主题切换，以及控制台、收件箱、用户、角色、菜单、协作空间、消息等首批页面骨架。
- 将新线主题收口到 CSS Variables，后续允许整体切换品牌色；同时补 `frontend-shadcn/docs/architecture.md`、`frontend-shadcn/README.md` 和 `start-frontend-shadcn.bat`。
- 更新根目录协作文档，使 `frontend-shadcn/` 成为当前新的管理端重写线。

### 下次方向
- 继续把用户、角色、菜单、协作空间、消息主链从骨架页推进到真实工作面。
- 补统一数据层与请求层，再逐步替换静态占位数据。

# 2026-04-03 清理 Fluent 迁移线，准备重写管理端

### 本次改动
- 清理 `frontend-fluentV2/` 目录及其启动残留文件，结束上一条 Fluent 迁移线。
- 清理根目录中与 Fluent 迁移线绑定的约束描述，将仓库状态重置为“保留 Vue 主线，新的重写线待建立”。
- 保留 `frontend/`、`backend/` 与既有仓库级文档入口，为后续新的管理端重写线留出干净起点。

### 下次方向
- 新建新的管理端前端目录，并先确定组件体系、路由模式、状态边界和 mock 数据边界。
- 在新的重写线建立完成后，再补对应目录下的专题文档与启动脚本。

# 2026-04-02 frontend-fluentV2 路由全面接线与消息/收件箱首批迁移

### 本次改动
- `route-registry` 新增工作台收件箱、系统/协作空间消息调度、模板、发送人、收件组、记录、目录以及协作空间治理页的本地路由映射，动态导航命中这些路径时可直接进入 React 版页面，不再落到占位页。
- 新增 `features/workspace/InboxPage`，结合收件箱列表、详情、已读/批量已读与待办处理主链，URL 记忆当前筛选与选中项。
- 新增 `features/message/components/MessageDispatchWorkspace`，在系统域与协作空间域接入 `/api/v1/messages/dispatch` 真实调度链路，支持模板带入、发送人/受众/优先级选择与目标勾选。
- 新增 `features/system/menu-space/MenuSpacePage`，支持菜单空间列表、创建/编辑、空间模式切换、默认菜单初始化与 Host 绑定管理，并在路由 `/system/menu-space` 接入。
- 新增 `features/system/page/PageManagementPage` 与 `features/system/api/ApiEndpointPage`，分别承接页面治理与接口治理主链，并接入 `/system/page`、`/system/api-endpoint` 本地路由。

### 下次方向
- 继续迁移 `system/access-trace`、`system/fast-enter`、`system/feature-package` 等剩余治理页到 React，并补充相应路由注册。
- 完成一轮 `frontend-fluentV2` 端到端自测（登录、导航、收件箱、消息调度、协作空间治理），将结果补充到 change-log。

# 2026-04-02 frontend-fluentV2 全量页面收口范围表与壳层状态统一

### 本次改动
- 补充第 8 版全量页面收口的范围文档 [frontend-fluentV2/docs/page-inventory.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/docs/page-inventory.md)，把 `frontend/src/views/**/*.vue`、docker PostgreSQL 里的 `menus` / `ui_pages` 与当前 React 承接页统一到一份清单里，后续页面迁移以此为准。
- 将运行时导航错误态与壳层初始化错误态从 `MessageBar` 收口为 `PageStatusBanner`，让页面级状态提示继续跟 Fluent 2 Web / Teams 的低噪风格保持一致。
- 外链页空 URL 的提示也切到统一页面状态条，减少公共页与业务页之间的视觉分裂。

### 下次方向
- 后续新增或拆分页面时，先回到 page-inventory 清单补范围，再做 React 页面或模块实现，避免只盯单一域。
- 继续把重页里的模块向 `features/<domain>` 下沉，确保第 8 版是“全量页面 + 全量模块”收口，而不是只完成几个热点域。

# 2026-04-02 仓库规范补充：数据库临时迁移与默认种子

### 本次改动
- 补充仓库级协作约定：当代码修改需要调整数据库时，可以先增加一个临时迁移并执行验证；如果迁移已经完成目标、当前数据库也已稳定，就删除这份临时迁移，避免把一次性修复脚本长期保留在迁移链里。
- 同步明确默认初始化数据应优先通过 seed / ensure 逻辑完善，迁移只承担一次性结构变更或历史数据修正，不把长期默认状态反复写进迁移步骤。

### 下次方向
- 后续凡是涉及数据库改动的任务，先判断是结构变更还是默认初始化补齐；前者可用临时迁移，后者优先补 seed。
- 如果某个临时迁移已经执行成功并且不再需要回放，就在代码库里删除它，避免后续启动重复触发历史修正。

# 2026-04-02 frontend-fluentV2 菜单重复节点接口核查与前端去重

### 本次改动
- 通过真实登录后的运行时导航接口核查到重复菜单来源：后端 `menu_tree` 里确实返回了两条同名 `PageManagement` 节点，一条路径为 `/system/page`，另一条为相对路径 `page`，两者在前端解析后最终指向同一菜单入口。
- 前端在 `buildNavigationItems` 中增加了同级保守去重，按 `path + label` 只保留一条，避免接口里的重复菜单直接渲染成两个“页面治理”。
- 菜单图标继续沿“接口没给就不渲染”的策略，不再回到默认图标兜底；本轮仅收口重复节点，不改后端菜单契约。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 若后端后续要修数据源，应从菜单表里消除同名同路径的重复记录，而不是继续依赖前端兜底。
- 继续用真实菜单数据回归其它分支，确认前端这层保守去重不会误伤合法的同名不同层级菜单。

# 2026-04-02 frontend-fluentV2 菜单图标仅在后端提供时渲染

### 本次改动
- 调整运行时菜单图标策略：如果后端菜单接口没有提供 `icon`，前端不再自行推断或补默认图标，而是直接不渲染图标，避免不同菜单因为前端兜底规则过粗而看起来一样。
- 将导航类型里的 `icon` 改为可选字段，并同步修正菜单搜索和侧栏渲染，保证“没图标就不画”在搜索结果、桌面展开树、收缩态级联浮层和移动端抽屉里都一致生效。
- 继续保留菜单的无限层级、自动扩宽、长标题截断与悬浮提示等已有行为，只收口图标表现，不回退结构层逻辑。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 回归真实菜单数据，确认后端哪些节点本来就不提供图标后，前端确实完全不渲染，而不是又被别的兜底逻辑重新补出来。
- 若后续还要增强图标表现，只能优先让后端补明确图标字段，再由前端按字段渲染，不再做模糊推断。

## 2026-04-02 frontend-fluentV2 第七版长标题宽度收窄

### 本次改动
- 进一步收紧长标题对菜单宽度的影响：侧栏宽度估算系数降低、三级及以下层级的额外宽度减小，收缩态浮层最大宽度也继续下调，避免一个很长的菜单标题把整列菜单撑得过宽。
- 菜单标题在侧栏和收缩浮层里继续保持单行截断逻辑，优先保住整列宽度和层级结构，不再为了显示完整长标题让菜单块整体变宽。
- 三级及以下节点的缩进和引导线仍然保持收浅，配合更小的字体和块高，整体更偏轻量微软风格。
- 仍然保持后端运行时导航契约和菜单空间接口不变，仅调整前端宽度估算和视觉表现。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 用更长的真实标题继续回归，重点确认现在的宽度上限是否足够保守，以及单行截断是否仍然可读。
- 如果后面还要继续压宽度，优先再收紧一级和二级节点的文本占位，而不是再放宽容器。

## 2026-04-02 frontend-fluentV2 第七版菜单父子重复修复与三级压浅

### 本次改动
- 针对菜单里出现“父节点展开后又重复显示成子节点”的问题，前端在 `buildNavigationItems` 阶段增加了保守去重，自动剔除与父节点 `routeId`、`path` 或 `label` 完全一致的自重复子节点，避免同一个菜单项被当成父子两份渲染。
- 三级及以下菜单层级的视觉再次收浅：缩进、引导线、块高和字号都继续压低，减少深层目录一层层往里堆叠的厚重感。
- 收缩态级联浮层同步减轻阴影和字号，并把宽度上限收窄，避免长标题把浮层撑得过宽。
- 这轮仍然保持导航契约、运行时 DTO 和菜单空间接口不变，只修前端侧的树清理与视觉层级表现。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 用真实菜单数据继续回归“父子重复”是否彻底消失，尤其关注那些后台树本身带有别名或自引用痕迹的节点。
- 再看一轮三级及以下的视觉层级，如果仍然显深，可以继续收缩引导线和子节点字号，但先不要把信息层级压没。

## 2026-04-02 frontend-fluentV2 第七版菜单视觉细化

### 本次改动
- 在无限层级和自动扩宽基础上继续收口菜单视觉：侧栏主项、子项、收缩态 rail 和级联浮层都下调了字号、块高和缩进强度，整体从“深、厚、重”调整为更轻的微软企业后台风格。
- 桌面展开态的侧栏宽度动画从原来的快切节奏调整为更柔和的过渡，避免深层菜单展开时侧栏显得生硬。
- 桌面展开树的子节点间距、边界留白和层级引导线都收浅，减少菜单块一层层往里压的深度感。
- 收缩态级联浮层增加了更轻的阴影和边框，子菜单项字体也同步缩小，避免浮层显得过厚。
- 移动端抽屉头部和树体间距也做了轻量收敛，保证统一视觉节奏。

### 下次方向
- 继续用深层菜单和长标题回归视觉细节，重点看展开/收起动画是否还可以再轻一点，以及层级引导线是否需要继续减弱。
- 如果后续仍觉得“菜单块比较深”，可以再做一轮仅针对一级/二级/三级的字号和留白分级，但这轮先停在轻量化收口。

## 2026-04-02 frontend-fluentV2 第七版菜单导航自动扩宽优化

### 本次改动
- 在无限层级导航基础上继续收口菜单宽度问题：桌面展开侧栏不再固定为 `252px`，而是会根据当前可见导航树的层级深度和节点标题长度自动计算推荐宽度，展开深层节点时会平滑增宽，避免子节点较多或缩进较深时可视空间不足。
- 桌面收缩态的级联浮层也改为按当前层级菜单项内容自适应宽度，不再固定使用单一 `240px-300px` 的窄浮层；子节点标题较长时，当前浮层会自动放宽。
- 这轮仍然保持后端运行时导航 DTO、菜单空间接口和搜索模型不变，只调整了 `AppShell` 与 `SideNav` 的渲染宽度策略。
- 已完成 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 用更长标题、更多深层节点和至少两个真实 `menu space` 继续回归自动扩宽策略，确认宽度上限、空间隔离和运行时导航刷新后的表现都稳定。
- 如后续需要更强的可视反馈，可继续补“当前正在扩宽到哪一级”的层级提示，但这轮先不叠加额外交互。

## 2026-04-02 frontend-fluentV2 第七版菜单导航无限层级修复

### 本次改动
- 将 `SideNav` 从固定两层结构重构为递归导航树，桌面展开态现在支持任意深度内联展开；当前路由祖先链会自动展开，并与用户手动展开状态合并。
- 桌面收缩态改为递归级联浮层，一级图标可打开首层菜单，带子节点的菜单项可继续向右打开下一级，不再只支持单层子菜单。
- 移动端导航改为近全屏抽屉树，支持任意深度展开；点击叶子节点后会自动关闭抽屉并完成路由跳转。
- `useShellStore` 新增按 `menu space` 持久化的展开状态与剪枝动作，收缩/展开切换后仍保留当前空间下的手动展开记忆。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit`、`pnpm --dir frontend-fluentV2 build`，并用浏览器实测桌面展开、桌面收缩级联和移动端抽屉三种导航态。

### 下次方向
- 用至少两个真实 `menu space` 继续回归展开状态隔离，确认切空间后展开记忆不会串空间。
- 继续清理现有控制台网络噪音和 HMR 历史日志，重点区分后端接口异常与导航壳层本身的问题。

## 2026-04-02 frontend-fluentV2 第七版协作空间联调收口与治理页回归增强

### 本次改动
- 第七版优先解决协作空间域联调收口问题。后端 `my-team` 只读接口在“当前账号暂无协作空间上下文”时，不再直接返回 404 业务错误，而是统一改成 `200 + 空结果`，覆盖协作空间边界角色列表、动作来源、菜单来源、协作空间成员列表和成员边界角色列表，避免协作空间治理页在无协作空间场景下把网络错误暴露给前端。
- 在此基础上继续补齐“默认协作空间回退”语义：`/api/v1/user/info` 和 `my-team` 在没有显式 `X-Collaboration-Workspace-Id` 时，会回退到当前用户的默认协作空间成员关系，使前端能够恢复 `current_collaboration_workspace_id` 并继续使用协作空间治理主链。
- 前端 `system/team-roles-permissions` 与 `team/team-members` 已同步改成业务空态承接：当当前账号没有协作空间归属时，页面会显示明确的引导区和协作空间入口卡，而不是停留在通用错误页；相关保存按钮也会进入清晰的禁用态。
- 消息治理页继续做第七版回归增强：`system/message-template`、`system/message-sender`、`system/message-recipient-group` 现在补齐了页面级成功/失败反馈、URL 选中恢复，以及“删除能力待后端开放”的禁用说明，不再让没有后端删除契约的能力看起来像是前端漏接。
- `system/feature-package` 继续做治理页主链收口：基础信息保存、删除、子包关系、动作关系、菜单关系、协作空间关系都补上了统一反馈，并在当前选中功能包失效后自动清理 URL，减少列表与右侧详情不同步。
- 运行时导航标题继续收口：左侧导航构建时优先使用 `RuntimeNavItem.title` 而不是后端 label key，进一步减少后端翻译键直接暴露到 Fluent 2 壳层。
- 进一步为 `/system`、`/team`、`/workspace`、`/message` 等目录根节点补上本地组标题映射，减少侧栏和面包屑根节点继续显示 `menus.*` 式后端 key。
- 本轮联调后已重新启动后端服务，并用真实接口确认 `/api/v1/collaboration-workspaces/current/boundary/roles`、`/api/v1/collaboration-workspaces/current/action-origins`、`/api/v1/collaboration-workspaces/current/menu-origins`、`/api/v1/collaboration-workspaces/current/members` 在无协作空间场景下已经全部返回 200。

### 破坏性调整
- `my-team` 只读查询接口在无协作空间场景下的 HTTP 语义已从 `404 + ErrNoTeam` 改为 `200 + 空结果`。旧前端如果明确依赖 404 来判断“暂无协作空间”，需要改为根据当前数据是否为空、或根据当前用户租户上下文判断。
- `user/info` 和 `my-team` 新增了默认协作空间成员关系回退。若旧逻辑假定只有显式 `X-Collaboration-Workspace-Id` 才会出现 `current_collaboration_workspace_id`，需要接受当前账号会自动进入默认协作空间上下文。
- 消息模板、发送人、收件组页现在会在当前选中项从列表中消失时主动清理 URL 里的选中参数；后续若继续在这些页面上做操作回滚或缓存补丁，必须保持 URL 与当前详情同步。

### 需要人工回归验证
- 当前管理员账号虽然已经可以自动回退到默认 `my-team` 上下文，但后续继续做协作空间深回归时仍建议使用专门的联调协作空间数据，避免直接污染系统默认协作空间。
- `workspace/inbox`、消息调度、模板/发送人/收件组页需要继续在真实业务数据下验证保存、刷新和回执是否完全符合预期。
- `system/feature-package` 仍需继续做真实数据下的关系配置回归，确认子包、动作、菜单、协作空间四类关系在保存后都能稳定回刷。

## 2026-04-02 frontend-fluentV2 第六版主链回归与前后端联调收口

### 本次改动
- 在第五版结构化详情与域级懒加载基础上继续推进第六版，优先把消息域、协作空间域、系统治理域的高频主链补成真正可回归的闭环，而不是只停留在结构化工作台层面。
- `workspace/inbox` 已补齐已读、批量已读、待办处理三条主链的成功/失败反馈，并在列表刷新后自动清理失效选中项；详情区补上了优先级、租户上下文、待办状态和更新时间等结构化信息。
- `system/message` 调度台现在会在发送成功后自动把新记录切成当前选中记录，并展示结构化回执；消息工作台仍继续共用系统域和协作空间域同一套 feature，不再复制两套逻辑。
- `team/team` 补上删除协作空间、成员加入、成员移除、成员角色更新等主链反馈；`team/team-members` 补上成员增删改和边界角色分配反馈，并在成员被移除后自动清理 URL 中失效的当前选中项。
- `system/team-roles-permissions` 补上边界角色创建、删除、动作授权、菜单授权、功能包授权的统一反馈；`system/user` 补上用户创建、删除、角色分配、功能包授权、菜单授权和权限快照刷新的统一反馈。
- 前后端联调上继续做非破坏式收口：前端协作空间边界角色创建接口改为 `/api/v1/collaboration-workspaces/current/boundary/roles`，后端同时新增 POST `/api/v1/collaboration-workspaces/current/boundary/roles` 别名，兼容第六版前端主链而不破坏旧路径。

### 破坏性调整
- 协作空间边界角色创建链路已从前端旧的 `/api/v1/collaboration-workspaces/current/roles` 对齐为 `/api/v1/collaboration-workspaces/current/boundary/roles`；后续若继续扩协作空间边界角色逻辑，必须沿 `boundary/roles` 这套语义继续走，不要再把成员角色和边界角色混回同一路径。
- 多个治理页新增了统一的页面级反馈条，后续如果继续补操作反馈，必须保持“操作成功/失败 -> 页面级反馈 + query invalidate”这一模式，不要回退到只靠控制台或隐式刷新。

### 需要人工回归验证
- 当前运行中的后端实例仍对 `/api/v1/collaboration-workspaces/current/action-origins`、`/api/v1/collaboration-workspaces/current/menu-origins`、`/api/v1/collaboration-workspaces/current/boundary/roles` 返回 404；仓库源码已包含这些路由，说明联调环境需要重启或重新部署到最新后端。
- `workspace/inbox` 需要继续用真实业务数据验证：未读消息点击后自动已读、批量已读、待办完成/忽略后列表与详情是否同步收口。
- `team/team`、`team/team-members`、`system/team-roles-permissions` 需要继续用真实协作空间上下文回归：协作空间删除、成员移除、边界角色授权和来源说明是否都符合预期。
- `system/user` 需要继续验证真实用户数据下的角色分配、菜单授权、功能包授权和权限快照刷新链路。

## 2026-04-02 frontend-fluentV2 第五版业务深化、结构化详情与域级懒加载

### 本次改动
- 在第四版全量页面接入基线上继续做第五版深化，优先补消息域、协作空间域和系统治理域的结构化详情，不再让右侧详情区继续依赖原始 JSON 文本框或只读拼接文本框。
- 消息域完成第一波工作台深化：`workspace/inbox`、`system/message*`、`team/message*` 继续共用一套消息 feature，调度页现在能展示结构化目标实体预览和结构化发送回执；记录详情页补齐投递汇总、状态时间线、投递明细和 payload 摘要卡，收件组页也补上了匹配模式、预计人数和目标范围表格。
- 协作空间域完成第一波结构化治理收口：`team/team-members` 成员详情改为属性区 + 角色区 + 操作区；`system/team-roles-permissions` 改成三栏治理台，并直接接入 `/api/v1/collaboration-workspaces/current/menu-origins` 与 `/api/v1/collaboration-workspaces/current/action-origins` 的来源说明面板，统一展示动作、菜单和功能包授权来源。
- 系统治理域继续去 JSON 化：`system/access-trace` 改成结构化链路摘要与记录表；`system/feature-package` 的 impact preview 改成指标卡 + 属性区；`system/user` 的权限诊断改成结构化诊断摘要、角色链与来源包列表。
- 路由层开始第五版稳态收口：认证页、公共静态页和各业务域页面全部切到按域懒加载，并在 `vite.config.ts` 中增加 `manualChunks`，构建产物已拆成 `auth / dashboard / workspace / message / system / team` 六个主业务 chunk。
- 同步更新 `frontend-fluentV2/README.md`，将当前阶段正式改写为第五版，并把当前已完成能力、剩余风险与第六版建议收口到专题说明里。

### 破坏性调整
- `frontend-fluentV2` 的本地路由定义不再默认直接静态 import 页面组件，而是通过域级 lazy route 包装；后续新增页面时必须继续沿用 `createLazyRouteElement`，不要再把大批页面直接同步打进首包。
- `message.api.ts`、`access.api.ts`、`system.api.ts`、`collaboration_workspace.api.ts` 的 adapter 返回值已进一步稳定化，页面侧如果继续直接假定原始 DTO 结构，会和当前稳定内部类型脱节。

### 需要人工回归验证
- 当前运行中的后端实例对 `/api/v1/collaboration-workspaces/current/action-origins`、`/api/v1/collaboration-workspaces/current/menu-origins`、`/api/v1/collaboration-workspaces/current/boundary/roles` 返回了 404；仓库源码里这些路由存在，需确认联调环境是否已重启到最新后端。
- 消息域需要继续做真实业务回归：收件箱待办处理、消息调度成功回执、模板/发送人/收件组编辑、记录明细查看。
- 协作空间域需要继续做真实业务回归：新增成员、修改角色、边界角色授权，以及来源说明在真实协作空间上下文中的联动。
- 域级拆包已经生效，但 `auth` 和 `vendor-fluent` 仍然偏大，下一轮需要继续细拆并检查是否还有可以下沉到更细 chunk 的页面依赖。

## 2026-04-02 frontend-fluentV2 第四版完善收尾与第五版起步

### 本次改动
- 对第四版已迁移页面做了一轮后端页面注册审计：确认 `system/access-trace` 已由独立命名迁移维护，不需要重复回灌到 `DefaultPages`；同时将 `system/more`、`team/more` 补入非菜单直达页 `UIPage` 种子，并新增 `20260402_message_more_page_seed` 命名迁移，保证新环境和存量环境都能同步到页面注册表。
- 前端运行时导航继续在第四版架构内做小步收口：导航项标题、页面标题与运行时面包屑现在优先使用本地 React 路由元数据，减少后端翻译键或旧标题直接暴露到 Fluent 2 壳层。
- 更新 `frontend-fluentV2/README.md`，明确第四版收尾结果、后端页面迁移审计结论，以及第五版起步范围，避免后续继续把“哪些页需要进后端页面注册”当成模糊问题反复判断。
- 继续推进第五版稳定化，修复了 [SideNav.tsx](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/shell/components/SideNav.tsx) 中将多个 `makeStyles` class 以模板字符串拼接后再传给 Fluent `Button` 的问题，统一改为 `mergeClasses(...)`，收掉了开发态的 `mergeClasses()` 控制台错误噪音。

### 破坏性调整
- 后端 `permissionseed.DefaultPages()` 新增了 `system.message.more`、`collaboration_workspace.message.more` 两个页面种子；若有脚本依赖旧页面列表，需要同步接受这两条新页面注册记录。
- 前端导航标题展示策略改为“本地路由优先，后端运行时标题兜底”；如果后续新增本地路由时希望保留后端标题，需显式调整本地 `shellTitle`。

### 需要人工回归验证
- 重新执行 `backend/cmd/migrate` 后，确认 `system/more`、`team/more` 已进入 `ui_pages`，并在系统页面治理页中可被查询到。
- 登录后检查左侧导航、页签、面包屑和页面头部标题，确认已实现页面优先显示 React 侧中文标题，而不是 `menus.*` 形式的旧键名。
- 对 `system/message`、`team/message` 及其 `more` 页再做一轮联动回归，确认新增页面注册不会影响消息域现有权限和激活菜单逻辑。
- 继续抽查其他壳层交互页，尤其是移动端侧栏、消息页、菜单页和页签交互，确认后续新增样式时不再回退到字符串拼接 className 的写法。

## 2026-04-02 frontend-fluentV2 第四版全量页面迁移版

### 本次改动
- 在第三版真实认证、真实运行时导航、真实菜单治理基线上继续增量演进，没有重做壳层、认证或请求分层，而是把 `route registry` 拆成按域组织的本地路由清单，开始承接 Vue 侧既有页面的 React 全量对应物。
- 新增第四版真实页面矩阵：控制台、收件箱、个人中心、页面治理、接口治理、快捷入口、菜单空间、访问轨迹、角色、用户、权限动作、功能包、协作空间边界角色，以及系统域与协作空间域的消息、模板、发送人、接收组、记录、更多入口页都已接入真实 React 页面，不再依赖占位页兜底。
- 继续沿用 `API module -> adapter -> query/mutation hook -> page` 分层，补齐 dashboard、inbox、system、access、message、team 等域服务，页面层不再直接解析后端 DTO，也不再把 response envelope 当数组直接消费。
- 新增一批 Fluent 2 工作台页范式：控制台采用摘要型 workbench，收件箱和消息页采用三栏协作布局，系统/协作空间治理页采用列表 + 详情/编辑的治理工作面，视觉方向统一为微软企业后台风格。
- 扩展公共静态路由，补齐 `/403`、`/404`、`/500`、`/result/success`、`/result/fail`、`/outside/iframe/*` 等真实 React 对应页，并让 `register`、`forgot-password` 成为正式认证页的一部分。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 通过；同时在新的 dev 端口上完成登录后全量受保护路由烟测，新增页面未再落入占位页或错误边界。

### 破坏性调整
- `frontend-fluentV2/src/features/navigation/route-registry.tsx` 不再维护单文件内联本地路由，后续新增本地已实现页面时需要继续按域写入 `core-routes.tsx`、`system-routes.tsx`、`team-routes.tsx`。
- 多个治理页现在直接依赖规范化后的 relation envelope（`ids / items / inherited / records`）；若后续改动 adapter 返回结构，必须同步调整页面而不是在页面里补 DTO 兼容。

### 需要人工回归验证
- `workspace/inbox`、系统消息页、协作空间消息页的发送、已读、待办动作是否在真实业务数据下全部符合预期。
- `system/page`、`system/api-endpoint`、`system/role`、`system/user`、`system/action-permission`、`system/feature-package`、`team/team`、`team/team-members` 的创建/编辑主链虽然已接通，但仍需结合真实业务数据做一轮逐页 CRUD 回归。
- 开发环境控制台仍存在 `mergeClasses()` 噪音，说明部分 className 组合方式还需要继续清理；虽然不阻塞构建，但应作为下一轮稳定化目标。
- 公共页中的注册、忘记密码、异常页与外链页已接入 React 路由，仍建议在完全游客态下再做一轮独立回归，确认登录态恢复和重定向不会干扰公共路由。

## 2026-04-02 frontend-fluentV2 第三版稳定化与菜单治理编辑版

### 本次改动
- 在第二版真实认证链路上补齐 refresh token 闭环，接入单飞刷新与 refresh 失败统一清会话回登录，页面层不再感知 token 刷新细节。
- 将认证启动阶段状态收口为更明确的 bootstrap 流程，并为导航、菜单树、详情和关联页面查询补上更稳定的 query key 与 placeholderData 策略，降低启动和切空间时的闪动。
- 增补统一空间级失效策略：切换菜单空间或执行菜单 create/update/delete 后，会集中刷新运行时导航、菜单树、详情、关联页面与分组查询。
- 将 `system/menu` 从第二版“真实浏览 / 只读详情”推进为第三版“受控编辑版”，支持 URL 恢复 `spaceKey / selectedMenuId / keyword`，并接入顶级/同级/子级菜单创建、核心字段编辑、类型感知表单、管理分组归属修改。
- 接入真实删除预检与删除确认：支持 `single / cascade / promote_children` 模式预览，删除成功后会刷新树、恢复右侧详情上下文并同步 URL。
- 已验证 `pnpm --dir frontend-fluentV2 install`、`pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前生产构建仍存在既有的大 chunk 警告，需要后续按页面或功能块拆分。

### 破坏性调整
- `frontend-fluentV2` 的 query key 已统一重命名分组；若后续新增 query 或 invalidate 逻辑，必须继续沿用 `auth / navigation / menu` 当前命名，不要再添加散落 key。
- `system/menu` 当前选中节点和搜索关键字已进入 URL；若后续修改菜单页路由参数策略，需要同步考虑刷新恢复与删除后回退逻辑。

### 需要人工回归验证
- access token 过期后，多请求并发触发时是否只发生一次 refresh，并且 refresh 失败后能稳定回到登录页。
- 切换菜单空间后，当前页面若为 `system/menu`，URL、树高亮、详情区、表单和删除预检是否都与新空间一致。
- `system/menu` 创建同级 / 子级节点后，树定位与右侧表单切换是否符合预期；删除当前节点后是否总能回落到合理节点或空态。

## 2026-04-02 frontend-fluentV2 第二版接入真实认证、运行时导航与菜单浏览

### 本次改动
- 将 `frontend-fluentV2` 从第一版 mock shell 升级为第二版真实运行时基础层，补齐真实登录、会话恢复、当前用户获取、退出登录、全局 `401` 清会话与 redirect 回跳闭环。
- 重构共享请求层与适配层，新增统一 Axios client、接口模块、Query key、错误模型与 adapter 映射，页面层不再直接消费后端 DTO 细节。
- 接入真实菜单空间与运行时导航：当前空间可持久化，登录后会按空间加载导航树，左侧导航、面包屑、迁移占位页与 `system/menu` 查询上下文会联动刷新。
- 路由继续保持静态壳 + route registry 模式：本地已实现页面正常进入，运行时存在但未迁移页面统一进入上下文化占位页，避免动态组件路径直驱前端实现。
- 将 `system/menu` 从占位页推进为真实浏览版，支持真实菜单树加载、空间联动、搜索、节点详情只读展示、页面绑定信息只读展示与基础空态/错误态渲染。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 通过；当前生产构建仍有既有的大 chunk 告警，但不影响本次接入结果。

### 下次方向
- 第三版优先继续沿第二版的 adapter、query 与 route registry 基础，逐步把 `system/page`、`system/interface`、`system/role` 等治理链路从占位页迁移为真实页面。
- `system/menu` 下一阶段建议补编辑、新增、删除、排序和更完整的关联治理能力，但继续保持危险操作集中在 feature/service 层，不把 DTO 兼容逻辑回灌到页面组件。

# 2026-04-02 frontend-fluentV2 去除 React Fluent 2 显式标识

### 本次改动
- 清理了 `frontend-fluentV2` 中面向用户的迁移线文案，把错误页、初始化提示、404、迁移占位页、欢迎页和系统菜单中的 `React Fluent 2` / `Fluent 2` 显式字样替换为中性表述。
- 同步更新了应用配置里的产品名与副标题，避免壳层顶部和页面空态继续暴露迁移线标签。
- 保持现有壳层逻辑不变，只调整可见文案和品牌描述。

### 下次方向
- 如果后续还要继续淡化迁移痕迹，可以再检查 README、路由元信息和 mock 数据里的剩余品牌措辞。
- 当前功能层没有改动，后续可继续围绕壳层布局、导航体验和真实业务页面迁移推进。

## 2026-04-02 React Fluent 2 页签标签栏补齐右键菜单与轻量标签组

### 本次改动
- 继续增强 `frontend-fluentV2` 的页签壳层，补上了右键上下文菜单与“关闭左侧 / 右侧标签”能力，并将页签状态持久化到本地，支持页面刷新后的恢复。
- 为页签增加了轻量“合并”模式：连续同模块标签会自动并入一个标签组，标签组支持折叠，用于在页签较多时降低横向噪音。
- 当前页签仍保留固定、刷新、关闭其他、拖动重排与横向滚动优化；右侧工具条和右键菜单形成了双入口，不需要把所有操作挤进单个页签按钮。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次页签增强无关。

### 下次方向
- 可继续补固定标签持久化策略、右键菜单中的“关闭当前组 / 展开当前组”，以及标签组的更明确视觉区分。
- 如果后续真的要做浏览器式复杂“标签合并”，建议继续往工作集 / 标签组模型推进，而不是硬套标准 Tablist。

## 2026-04-02 React Fluent 2 页签标签栏增强

### 本次改动
- 在 `frontend-fluentV2` 的页签标签栏基础上继续补齐固定标签、关闭其他标签、刷新当前标签和桌面端拖动重排能力，操作条统一收口在标签栏右侧。
- 新增页签排序与刷新状态管理：固定标签会自动保持在左侧分区，刷新当前标签会重新挂载当前内容区，关闭其他标签时会保留固定标签与当前标签。
- 对标签栏滚动体验做了优化，支持鼠标滚轮横向滚动、当前页签自动滚入可见区域，并在左右边缘增加渐隐提示，减少长标签列表的压迫感。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次页签增强无关。

### 下次方向
- 若继续往浏览器式体验靠拢，下一步建议补“右键菜单”“关闭左侧 / 右侧标签”“固定标签持久化”和“多标签恢复”。
- Fluent 官方 Tablist 更适合少量相关内容切换，并建议超出宽度时使用 overflow menu；页签分组或“合并”应视为自定义壳层模式，后续更适合做成可折叠的标签组，而不是直接复用标准 Tablist。

## 2026-04-02 React Fluent 2 壳层新增页签标签栏

### 本次改动
- 为 `frontend-fluentV2` 的应用壳新增了浏览器式标签栏，位置放在顶部栏下方、页面内容上方，形成“顶部栏 -> 标签栏 -> 页面内容”的三层结构。
- 新增壳层级页签状态：访问已注册路由会自动打开对应标签，相同路径不重复开标签，关闭当前标签时会自动跳到相邻标签。
- 标签能力接入了路由注册表与页面元数据解析，并在退出登录时清空当前会话的已打开标签，避免跨会话残留。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次页签功能无关。

### 下次方向
- 可继续补固定标签、关闭其他标签、刷新当前标签等增强交互，让页签更接近正式后台工作台。
- 若后续接入运行时菜单或 host 驱动路由，可直接复用当前壳层页签模型，只替换标签来源与恢复策略。

## 2026-04-02 React Fluent 2 补齐本地认证闭环

### 本次改动
- 在 `frontend-fluentV2` 中新增本地假认证 store，支持登录、退出登录、记住登录状态以及基于 `localStorage / sessionStorage` 的会话持久化。
- 接入了认证守卫和游客重定向：未登录访问后台壳会自动跳到 `/login`，已登录访问 `/login`、`/register`、`/forgot-password` 会直接回到默认工作区。
- 登录页已支持登录成功后按来源地址回跳，顶部用户菜单新增退出登录，认证页与后台壳形成完整本地闭环。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次认证逻辑无关。

### 下次方向
- 若继续接真实认证，下一步优先把登录提交动作替换为 adapter 请求，并补上令牌刷新、401 处理和登录态失效回退。
- 后续可继续补密码显隐、错误提示、验证码或邮箱确认等细节，同时把当前本地假会话逐步替换成真实会话模型。

## 2026-04-02 React Fluent 2 新增认证页骨架

### 本次改动
- 为 `frontend-fluentV2` 新增了不接真实认证的登录、注册、忘记密码三张独立页面，并通过壳层外路由接入 `/login`、`/register`、`/forgot-password`。
- 认证页统一采用居中单卡片布局，保持 Fluent 2 风格的简洁输入表单，去掉了第三方登录入口和不必要的营销式文案，只保留必要字段、跳转链接和本地示例反馈。
- 新增了认证页公共骨架，统一品牌、留白、底部链接和响应式行为；已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。
- 当前构建仍有既有的大 chunk 警告，但与本次认证页接入无关。

### 下次方向
- 若后续要接真实登录，可先把认证表单提交动作替换为 adapter 调用，再决定是否引入真正的未登录守卫与返回地址逻辑。
- 下一轮可继续补认证页的细节，例如密码可见切换、基础校验文案、登录成功后的 redirect 规则，以及移动端输入体验收口。

## 2026-04-02 React Fluent 2 壳层品牌配色收口

### 本次改动
- 为 `frontend-fluentV2` 新增统一品牌主题源，改用一套更克制的蓝灰企业色品牌梯度生成 Fluent light / dark theme，不再直接沿用默认品牌蓝。
- 同步更新了应用壳主题 Provider、错误边界页、搜索弹层选中态和品牌 logo 渐变，让导航高亮、链接、Badge、欢迎页渐变与 logo 颜色保持一致。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前仍存在既有的大 chunk 警告，但与本次配色调整无关。

### 下次方向
- 继续检查深色模式下的新品牌色对比度，重点看顶部栏图标、侧栏激活态和搜索结果高亮是否仍然足够清晰。
- 若后续继续细调，可把状态色、空态插画和少量硬编码图形资产进一步收口到同一品牌色板，避免局部残留旧蓝。

## 2026-04-01 Fluent 2 React 实验场新增 React 组件目录页

### 本次改动
- 基于 [Fluent 2 React 组件目录](https://fluent2.microsoft.design/components/web/react) 为 `frontend-fluent2-lab` 新增了 4 张专门的组件展示页，集中承载命令与导航、表单与选择、反馈与浮层、身份与内容四类组件。
- 新增文件 [FluentReactComponentPages.tsx](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/FluentReactComponentPages.tsx)，使用当前 `@fluentui/react-components` 已安装的稳定导出组件做真实示例，并把官方目录中的其余组件收进补充索引区。
- 同步更新 [catalog.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/lab/catalog.ts)、[App.tsx](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/App.tsx) 和 [README.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/README.md)，让实验场总页数增至 54。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过；未做浏览器逐页人工回归。当前新增组件页单文件 chunk 较大，后续如继续扩展可考虑拆分。

### 下次方向
- 继续把组件目录页和现有 Fluent 规范页打通，逐步将组件页从“目录陈列”收口成“组件 + 模式 + 场景”的联动结构。
- 若继续深化，可优先补 `Icon`、`Carousel`、`TagPicker` 等官方目录项的更完整示例，并检查新组件页在暗色主题下的观感与滚动密度。

## 2026-04-01 Fluent 2 React 实验场 Teams 场景骨架深化

### 本次改动
- 继续调整 `frontend-fluent2-lab/src/pages/ScenarioExpansionPages.tsx`，把 Teams 线页面进一步收成更明确的协作工作面，而不是只保留统一壳层。
- 重点强化了会议指挥、一线简报、文件协作、社区公告、交接班、活动运营、伙伴站会等页面的 Teams 风格骨架，引入议程条、会议舞台、成员条、文件审阅板、公告流、交接三栏、run of show 等更明确的结构。
- 这轮改动的目标是让 Teams 页面更像“频道 / 协作 / 会务 / 社区 / 值守”工作面，而不是同一模板换文案。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过；未做浏览器逐页视觉回归。

### 下次方向
- 继续对 Teams 线页面做浏览器回归，优先检查亮暗主题下的品牌紫使用强度、右侧上下文区密度和移动端收口方式。
- 若继续深化，下一步最值得做的是把会议指挥、文件协作、交接班这三类页面对齐到真实 Teams UI Kit 或 Microsoft Teams App Templates 的具体节点。

## 2026-04-01 Fluent 2 React 实验场剩余页面去模板化

### 本次改动
- 继续重构 `frontend-fluent2-lab/src/pages/ScenarioExpansionPages.tsx`，把剩余仍在复用家族骨架的页面继续拆成更明确的专用布局。
- 新增并接入租户总览、事件响应、策略工作室、资产台账、工单台、发布控制台、知识中枢、表单规范台、导航规范台、Token 治理、可访问性评审、交接模式、模板画廊、动效原则、侧栏参考、会议指挥、前线简报、文件协作、入职协作、社区公告、班次交接、直播运营、伙伴站会等专用工作面。
- 这轮调整的目标不是再扩页数，而是让 50 张实验页在结构、节奏和任务语法上真正拉开，减少“同一模板换内容”的重复感。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过；未做浏览器逐页人工视觉回归。

### 下次方向
- 继续从这批页面里挑出更核心的工作面，结合真实 Figma 节点和 Fluent 2 文档进一步收紧层级与细节。
- 后续可优先检查 `ScenarioExpansionPages` 的 chunk 体积与暗色主题观感，再决定是否拆分代码块或继续做视觉校准。

## 2026-04-01 Fluent 2 React 实验场扩展到 50 页

### 本次改动
- 为 `frontend-fluent2-lab` 新增场景数据层，并把新增页从“按来源三套整页模板”重构为“按场景骨架”组织。
- 将 30 张新增实验页重新映射到指挥台、工作台、目录页、规范页、评审页、频道页、线程页、公告页等多种布局家族，减少单纯换文案的重复页面。
- 重写实验场目录注册表并更新 `frontend-fluent2-lab/src/App.tsx`、`frontend-fluent2-lab/README.md`，让入口支持完整 50 页切换，并把总页数展示改为动态值。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过。

### 下次方向
- 从新增 30 张页面里继续挑出高价值工作台页，按真实 Figma 节点收紧布局与信息层级，而不是继续只扩数量。
- 若后续继续做视觉回归，可优先检查 `ScenarioExpansionPages` 大包的拆分策略，以及暗色主题下的新页面对比度。

## 2026-04-01 React Fluent 2 基础壳初始化

### 本次改动
- 新建 `frontend-fluentV2/` 独立 React + Fluent 2 工程，并接入 Router、Query、Axios、Zustand 与 Fluent Provider。
- 落地应用壳、顶部栏、空间切换器、侧边导航、面包屑、统一页面容器和迁移占位页。
- 重建根目录当前生效的协作约束、项目框架与前端规范，并新增 `frontend-fluentV2/docs/` 正式文档目录。
- 已验证 `pnpm --dir frontend-fluentV2 install`、`pnpm --dir frontend-fluentV2 exec tsc --noEmit`、`pnpm --dir frontend-fluentV2 build`。
- 已验证 `pnpm --dir frontend-fluentV2 dev` 可启动；由于本机 `9030` 已被占用，Vite 在 2026-04-01 实际回退到 `http://127.0.0.1:9031/`。

### 下次方向
- 优先迁移系统治理链路中的菜单、页面、接口、角色与用户页面。
- 将 mock 查询逐步替换为真实 adapter，同时保持壳层、路由和页面容器不变。
- 若继续收拢历史资料，可为 `frontend/docs/legacy/` 增加分类索引。

## 2026-04-01 React Fluent 2 侧栏壳体调整

### 本次改动
- 将品牌区从顶部栏移入左侧导航顶部，桌面端不再保留单独的“侧栏”收缩按钮。
- 侧栏顶部品牌位改为承担收起/展开交互：展开态显示品牌块，收起态显示右箭头，再次点击可展开。
- 同步收紧了侧栏宽度和头部信息布局，使顶部栏更聚焦当前区域与全局操作。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 继续细调侧栏顶部品牌块、分组标题和激活态样式，使其更接近目标参考图的节奏。
- 下一轮可继续收口菜单项密度、内容区留白以及顶部栏空间切换器的视觉权重。

## 2026-04-01 React Fluent 2 顶部栏与侧栏穿插布局

### 本次改动
- 调整桌面端壳体层级关系，让顶部栏为左侧菜单预留通道，菜单区域本体再向上抬入顶部栏范围。
- 侧栏不再只是位于内容区下方，而是形成“左侧菜单超过顶部栏基线”的视觉关系，更接近目标参考图。
- 同步根据侧栏展开/收起状态动态调整顶部栏左侧偏移，避免全局操作区与菜单区域互相挤压。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 下一轮继续细调菜单卡片顶部阴影、边界和内容区首屏对齐，让穿插关系更自然。
- 可继续收口页面标题区和菜单顶部品牌块的垂直节奏，避免出现“刚好碰线”的中间态。

## 2026-04-01 React Fluent 2 壳体回退到初版

### 本次改动
- 回退了“品牌区移入侧栏”“菜单上抬穿插顶部栏”“品牌区承担收缩交互”等实验性壳体改动。
- 恢复到最开始的稳定结构：品牌区回顶部栏、左侧菜单回归常规侧栏、桌面端独立收缩按钮恢复。
- 保留新工程既有的 Provider、路由、导航树、页面容器和 mock 数据骨架，不回退基础工程能力。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 之后如果要重新设计壳体，建议先用固定线框确认“顶部栏 / 侧栏 / 内容区”的边界，再进入视觉细调。
- 下一轮可以从初版结构上继续优化间距、层级和组件样式，但不再直接做跨区穿插实验。

## 2026-04-01 React Fluent 2 移除顶部栏空间切换

### 本次改动
- 移除了顶部栏里的手动菜单空间切换入口，不再把空间切换作为当前壳体的可见交互。
- 保留 `currentSpace` 状态和按空间过滤导航的底层能力，后续可改为按 host 自动同步，不需要推翻现有骨架。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 下一轮如接入 host 上下文，可直接把空间值写入 shell store，并移除当前 mock 默认空间初始化逻辑。
- 之后顶部栏可继续收口为“品牌 / 当前区域 / 全局操作 / 用户区”的更轻量结构。
## 2026-04-01 Fluent 2 实验场新增基础控件集与图标总览

### 本次改动
- 为 `frontend-fluent2-lab` 新增 `基础控件集` 分类，并将组件展示页、React 组件页统一迁入该分组，避免底层能力继续散落在 `Fluent 2 Web` 分类里。
- 新增 `React 图标总览` 页面，集中展示 `@fluentui/react-icons` 的基础图标 ID，支持按关键字过滤、按首字母分组浏览，并点击复制图标标识。
- 继续补齐组件页中此前只停留在“目录补充”的缺口组件，已落地 `Nav`、`Label`、`Dropdown`、`TagPicker`、`Carousel`、`FluentProvider` 的真实示例，并将 `Icon` 能力明确收口到图标总览页。
- 同步更新 `catalog`、实验场入口、README 与分组计数，已验证 `pnpm --dir frontend-fluent2-lab build` 通过。

### 下次方向
- 继续补图标页的变体使用示例，例如按 `Regular / Filled` 生成推荐导入名，或补充常用图标组合样板。
- 后续可再做一次浏览器回归，重点检查图标页的大数据量滚动、暗色主题对比度、移动端折叠表现，以及组件页新增示例的交互细节。

## 2026-04-01 React Fluent 2 顶栏接入真实菜单搜索

### 本次改动
- 参考旧 Vue 壳层里的 `ArtGlobalSearch`，在 `frontend-fluentV2` 新增了顶栏级菜单搜索弹层，而不是继续保留占位图标按钮。
- 新增导航拍平与搜索模型，按当前可见导航树递归提取叶子菜单，支持按菜单名、分组、路径过滤，并接入最近访问记录。
- 顶栏搜索按钮已可真实打开搜索弹层，并支持 `Ctrl/Cmd + K`、方向键切换、回车跳转；已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 继续补搜索结果的高亮标注和分组视觉，让命中原因更直观。
- 若后续接真实 host 上下文或动态菜单，可把当前基于 mock 的搜索数据源替换成运行时导航注册表，而不重写弹层交互。

## 2026-04-01 Fluent 2 实验场新增 React 组合模式页

### 本次改动
- 为 `frontend-fluent2-lab` 的 `基础控件集` 新增三张组合模式页：`React 组合页：导航与命令`、`React 组合页：表单与反馈`、`React 组合页：内容与协作`。
- 这三张页不再按“单组件示例”堆砌，而是把 `Nav`、`Toolbar`、`SearchBox`、`Field`、`TagPicker`、`Dialog`、`Card`、`Persona` 等基础控件收口成真实工作面，减少基础控件集继续千篇一律的风险。
- 已同步更新 `catalog`、实验场入口、README 和页面总数统计，并执行 `pnpm --dir frontend-fluent2-lab build` 验证；当前未做浏览器逐页人工回归。

### 下次方向
- 继续给组合模式页补更贴近真实开发使用的复制能力，例如图标导入名、组件组合代码片段或推荐搭配说明。
- 后续可把这三张组合页继续对齐官方 Fluent 2 React 文档结构，补一轮暗色主题和移动端滚动节奏回归。
## 2026-04-02

- 新增根目录文档 `使用与迁移部署说明.md`，整理了后端、前端、`a/` 目录的本地使用方式和迁移部署步骤。
- 明确了后端依赖的 PostgreSQL、Redis、Elasticsearch 和 MinIO 配置项，以及前端的安装和构建流程。
- 对 `a/` 下的复制副本用途做了说明，便于后续做独立迁移、打包和交付。

## 2026-04-02 React Fluent 2 标签栏新增顶部开关

### 本次改动
- 在 `frontend-fluentV2` 顶部左侧功能区新增“界面设置”入口，并提供“显示标签栏”开关，统一管理标签栏是否显示。
- 壳层在关闭标签栏时不再为后续路由访问自动开标签，仅保留已持久化标签状态，避免彻底丢失当前工作区上下文。
- 同步更新标签栏壳层规则文档，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 通过；本次未额外执行生产构建。

### 下次方向
- 继续收口标签栏体验，可补“关闭当前组”“固定标签优先恢复”“右键菜单边缘避让”等细节。
- 如后续需要更强的工作区能力，可在当前开关基础上继续扩展“默认启用策略”“按用户偏好记忆”和更完整的标签组模型。

## 2026-04-02 React Fluent 2 标签栏改为显式自由分组

### 本次改动
- 将 `frontend-fluentV2` 标签栏从“按模块自动归并”切换为“用户显式组合”，不再按上级菜单或模块名自动成组。
- 在标签右键菜单中新增“与左侧标签成组”“与右侧标签成组”“移出当前组合”“解散当前组合”，并把组合状态持久化到本地。
- 组合折叠后改为显示首个标签标题与数量；标签拖出原组合时会自动脱离原组合，保证分组结构与实际顺序一致。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补“关闭当前组合”“组合重命名”“组合内整体拖动”等更完整的工作集能力。
- 若后续需要更接近浏览器的标签体验，可继续研究组合颜色、固定组合和组合恢复策略，但建议先稳定当前显式分组心智。

## 2026-04-02 React Fluent 2 标签栏补齐即时滑动与组合拖放

### 本次改动
- 将标签拖拽重排改为拖动过程中的即时滑动，不再等到松手后才调整顺序；单标签与组合拖动都采用同一套壳层状态更新方式。
- 将“刷新当前标签”统一改名为“刷新当前页面”，与实际行为保持一致；工具栏按钮和右键菜单文案已同步收口。
- 支持把单个标签拖放到已有组合上加入该组合，组合本身也支持整组拖动重排；专题文档同步更新为“即时滑动 + 组合拖放”的新交互语义。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补“关闭当前组合”“组合重命名”“组合整体拖出为单标签”和更细的拖拽目标高亮，让组合交互更接近浏览器工作集。
- 若后续要进一步对齐浏览器标签组，可继续研究组合颜色、组合固定和组合级右键菜单，但建议优先稳定当前拖拽手感与边界行为。

## 2026-04-02 React Fluent 2 标签栏移除刷新入口并修正隔位拖拽

### 本次改动
- 移除了标签栏工具区和右键菜单中的“刷新当前页面”入口，避免标签栏继续承载与浏览器刷新心智冲突的动作。
- 将标签拖拽重排的判定从“路径锁定”改为“按悬停标签中线即时换位”，修复了拖到相邻标签后无法继续回拖、会出现一格隔离的问题。
- 专题文档同步改成当前交互语义，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补组合级右键菜单、组合整体拖出和更明确的拖拽占位高亮，让多标签工作区更接近浏览器标签组体验。
- 后续若要继续增强，可考虑加入拖拽自动横向滚动和组合级固定策略，但建议先观察当前即时换位的实际手感。

## 2026-04-02 React Fluent 2 标签栏改为拖拽合并

### 本次改动
- 去掉了标签右键菜单中的成组入口，将“合并”统一收口到拖拽手势里，不再让同一能力同时存在菜单和手势两套入口。
- 单标签拖动经过目标标签靠近来向一侧的三分之一区域时，会显示合并预览；继续拖深后才执行位置换位，从而把“合并”和“重排”分成两段手势。
- `AppShell` 与标签栏组件已切换到新的 `groupTabs(sourcePath, targetPath)` 接口，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补组合级拖拽高亮、组合整体拖出为单标签和拖拽自动横向滚动，让多标签体验更接近浏览器标签组。
- 如果后续要继续深化，可再研究组合级右键菜单，但建议只保留和组合本身强相关的动作，避免再次把合并入口做回菜单里。

## 2026-04-02 React Fluent 2 标签栏移除全部组合逻辑

### 本次改动
- 将 `frontend-fluentV2` 标签栏恢复为纯平标签轨道，删除了标签组合、折叠组合、合并预览、组级右键菜单和相关持久化状态，只保留打开、关闭、固定与拖拽换位。
- `useShellStore` 改回仅维护 `openTabs` 和 `tabsEnabled` 两类标签状态，`AppShell` 也同步移除了分组接线与页面刷新版本号逻辑，避免后续继续被旧分组模型干扰。
- `OpenTabsBar` 已重写为单层标签实现，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 通过；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续收拖拽换位的手感和横向滚动体验，但不再重新引入任何组合语义，先把纯标签工作区稳定下来。
- 如果后续还要增强标签栏，优先考虑固定标签恢复、右键菜单边缘避让和移动端展示，不再回到浏览器标签组那条路线。

## 2026-04-02 React Fluent 2 标签栏拖拽内核重构

### 本次改动
- 删除 `frontend-fluentV2` 标签栏里旧的 HTML5 DnD、旧拖拽幽灵层、旧目标判定和旧 FLIP 实现，重写为 `Pointer Events + 自定义拖拽层 + wrapper 级让位动画` 的纯平换位模型。
- 新拖拽逻辑按拖动标签实体当前位置计算换位，源标签在原位只保留占位壳；其他受影响标签在每次顺序变化前记录当前屏幕位置，并从该位置平滑过渡到新位置。
- 同步更新标签栏专题文档，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 通过；生产构建待本轮改动后一并复核。

### 下次方向
- 继续观察拖拽手感，重点回看横向自动滚动、边缘拖拽阈值与触控板场景，但不再恢复任何分组或合并逻辑。
- 若后续还要增强标签栏，优先补纯平标签模式下的细节体验，例如拖拽中的滚动辅助和右键菜单边缘避让。 

## 2026-04-02 frontend-fluentV2 第七版导航映射与协作空间域联调复核

### 本次改动
- 将运行时导航中的相对路径和旧别名统一映射到本地 React 路由，补齐 `Dashboard`、`Console`、`TeamRoot`、`TeamManage`、`CollaborationWorkspaceMembers`、`TeamRolesAndPermissions`、`TeamMessageManage` 等入口，左侧导航、标签和面包屑不再落到 `#/team`、`#/members`、`#/roles` 这类半成品路径。
- 为 `team/team` 与 `system/team-roles-permissions` 补齐选中项失效后的 URL 清理，避免删除、切换或列表刷新后遗留脏 `selectedCollaborationWorkspaceId` / `selectedRoleId`。
- 使用全新 Playwright 浏览器会话重新验证 `system/team-roles-permissions` 与 `team/team-members`；在最新后端环境下，`/api/v1/collaboration-workspaces/current/boundary/roles`、`/action-origins`、`/menu-origins` 均返回 `200`，页面首次渲染控制台 `0 error`。

### 下次方向
- 继续用真实协作空间数据做深回归，重点验证协作空间边界角色写操作、成员边界角色分配和协作空间消息主链，而不是只停留在读链路。
- 若后端运行时导航后续再新增旧风格相对路径菜单，继续把映射收口在前端导航层，不把别名兼容散落到页面组件里。

## 2026-04-02 frontend-fluentV2 第七版消息调度作用域与协作空间发送闭环

### 本次改动
- 为消息请求层补齐 `tenantMode` 和 `scope` 驱动，平台消息请求会显式移除 `X-Collaboration-Workspace-Id`，协作空间消息继续带当前租户头，`system/message` 与 `team/message` 不再误用同一组数据。
- 将消息调度 payload 映射为后端真实字段：`specified_users -> target_user_ids`、`tenant_users/tenant_admins -> target_collaboration_workspace_ids`、`recipient_group/role/feature_package -> target_group_ids`，并把调度页目标预览改成可勾选卡片，不再只靠手工输入 ID。
- 修正协作空间消息页初始草稿仍落到 `all_users` 的问题，调度草稿与发送后重置都改为以 `/api/v1/messages/dispatch/options` 的默认受众、默认发送人和默认优先级驱动。
- 后端 `messages/dispatch` 增量修正了 `normalizeTargetTenants`，`specified_users`、`recipient_group`、`role`、`feature_package` 不再被误判为“不支持的发送对象”；前端与浏览器已实测发通系统域和协作空间域消息。

### 下次方向
- 继续回归消息域的已读、全部已读、待办完成/忽略和记录详情，重点确认左列计数、中列记录和右侧详情在操作后保持同步。
- 如后续还要提升消息工作台，可继续优化调度成功回执与受众摘要，让协作空间域默认当前协作空间的回执文案和投递数表达更贴近业务语言。

## 2026-04-02 frontend-fluentV2 增补下次方向记录文档

## 2026-04-05 后台 App 维度一期接入

### 本次改动
- 后端新增 `App`、`AppHostBinding`、`AppContextMiddleware`，把导航运行时从“按菜单空间主导”调整为“按 App 解析、空间辅助”的上下文模型，默认内置 `platform-admin` 平台管理后台。
- 为 `menu_spaces`、`menus`、`ui_pages`、`page_space_bindings`、`feature_packages`、`menu_backups`、`api_endpoints` 等资源补齐 `app_key`，并让运行时导航、菜单、页面、功能包、接口注册查询全面支持 `app_key` 过滤与回填兼容。
- 前端新增“应用管理”页 `frontend/src/views/system/app/index.vue`，同时把菜单管理、页面管理、功能包管理、API 注册页全部切到当前 `App` 视角，默认锁定 `platform-admin`，菜单空间退居为 App 下的高级配置入口。
- `API 管理` 页补齐 `app_scope/app_key` 的表单保存、同步、未注册扫描、失效清理和概览统计透传；功能包、页面、菜单相关弹窗和列表统一按当前 `app_key` 加载候选资源，避免跨 App 混用。
- 继续把 `访问链路测试`、`协作空间管理`、`协作空间角色与权限`、`角色管理`、`用户管理` 相关弹窗切到当前 `app_key` 视角：页面候选、菜单树、功能包候选、协作空间边界与协作空间角色边界接口都开始显式透传 `app_key`，避免同 Host 下切换应用时回落到默认 App。
- 补齐 `frontend/src/api/system-manage.ts` 与 `frontend/src/api/collaboration_workspace.ts` 中角色、用户、协作空间、协作空间角色的菜单/功能包/边界 helper，使这些请求在 GET/PUT 场景下都能稳定把 `app_key` 带给后端 `AppContextMiddleware`。
- 已完成验证：`pnpm --dir frontend build` 通过；`go test ./...` 已在 `backend/` 模块目录执行并通过。

### 下次方向
- 第二期优先把“菜单定义”和“空间布局”进一步解耦，避免多空间场景下继续依赖单表 `menus + space_key` 承担全部语义。
- 继续把角色、协作空间、用户的功能包与菜单裁剪页显式补出 App 维度筛选，并考虑把菜单空间页完全降级为应用管理内的次级配置。
### 本次改动
- 新增 `frontend-fluentV2/docs/下次方向记录.md`，专门承接当前迁移线尚未完成的后续方向，避免未来继续推进时只依赖对话上下文。
- 在 `AGENTS.md` 中补充了维护规则：该文档存在时持续更新，全部完成后直接删除，不长期保留空文件。

### 下次方向
- 后续每轮完成时继续同步清理该文档中的已完成项，确保它只保留真实未完成事项。
## 2026-04-03 frontend-shadcn UI 选型清单

### 本次改动
- 新增 `frontend-shadcn/docs/ui-stack-checklist.md`，明确新管理端重写线的 UI 选型顺序：页面优先消费 `src/components/ui/*`，必要时仅在 UI 封装层内部补 `Radix` primitive。
- 同步更新 `frontend-shadcn/docs/architecture.md`，把这份清单挂入当前文档结构，后续可以作为页面开发和组件沉淀的统一入口。
- 本次只调整文档约束，没有修改业务实现或组件代码。

### 下次方向
- 后续新增组件时，先按清单判断是复用现有 `shadcn/ui`、补官方组件，还是在 `src/components/ui/*` 内新增 `Radix` 封装。
- 如果后面出现 2 到 3 个以上重复的高级交互场景，再考虑沉淀更细的组件分类清单和示例模板。

## 2026-04-03 frontend-shadcn 常见组件对照表

### 本次改动
- 新增 `frontend-shadcn/docs/ui-component-mapping.md`，按页面结构、数据展示、表单、反馈、浮层、导航六类场景列出推荐组件，减少页面开发时临时判断成本。
- 在 `frontend-shadcn/docs/ui-stack-checklist.md` 与 `frontend-shadcn/docs/architecture.md` 中补充对照表入口，让 UI 边界规则和具体落地映射形成一套可查文档。
- 本次仍只更新文档，不涉及页面实现和依赖调整。

### 下次方向
- 后续如果新增 `Dialog`、`Popover`、`Combobox` 等常用封装，可以继续把文档里的“推荐方案”同步到实际组件清单。
- 若页面里开始频繁出现高级交互，再补“已封装组件状态表”，区分已可直接用、待补齐、禁止页面直连三类状态。

- 暂不把这条规则下沉到全局技能；如果后续多个仓库都需要同样机制，再考虑把它抽成技能行为。

## 2026-04-02 frontend-fluentV2 下次方向记录改为持续备案

### 本次改动
- 将 `frontend-fluentV2/docs/下次方向记录.md` 的语义从“本轮下次方向”收口为“持续维护的方向备案”，明确要求后续只按条目增删，不整份覆盖。
- 同步更新 `AGENTS.md`，强调该文档允许跨轮次、跨模块持续保留未完成事项；只有全部完成时才删除。

### 下次方向
- 后续继续推进时，只修改本次涉及的事项条目；完成则删除，未完成则保留，不再把整份文档当成每轮收尾模板重写。
- 若未来多个仓库都要采用同一机制，再考虑把这条规则抽进技能，而不是当前就提升为全局行为。

## 2026-04-02 frontend-fluentV2 父菜单保留页面入口

### 本次改动
- 将侧栏点击规则收回到“有路径的父菜单仍可作为页面入口”，不再因为存在子节点就一刀切禁用跳转。
- 收缩态级联浮层同时补出父菜单自身的可点击菜单项，避免父菜单的页面入口在浮层中被子项覆盖掉。

### 下次方向
- 继续回归带子菜单的父节点，确认展开态、收缩态和移动端抽屉三种状态下，父节点页面入口与子菜单展开都符合后台菜单配置语义。
- 后续若仍出现“同名但不同语义”的菜单项，优先回查后端种子和迁移，而不是继续在前端做更激进的隐藏。

## 2026-04-02 frontend-fluentV2 第8版 Fluent 2 Web / Teams 页面收口

### 本次改动
- 新增 `PageStatusBanner`，把消息、协作空间等治理页的成功/失败反馈统一成更轻的 Fluent 2 提示条，减少各页自行拼接 MessageBar 的重复样式。
- 收口 `PageContainer`、`SectionCard`、`WorkbenchLayouts` 的视觉参数，整体压浅阴影、边框和间距，使页面更接近 Fluent 2 Web 的低噪后台风格。
- 为 `EntityPageLayout` 补齐更完整的标题/说明/元信息/操作区结构，作为后续页面模块化拆分时的统一骨架。
- 已将消息域、协作空间域中多处页面反馈接入新的统一提示组件，并完成 `tsc` 与 `build` 验证通过。
- `system/menu`、`message/workspace`、`team/team`、`system/user`、`system/team-roles-permissions`、`system/feature-package` 等高频治理页正在继续切换到统一的页面反馈组件，减少旧式 MessageBar 视觉碎片。

### 下次方向
- 继续把 `system/menu`、`system/page`、`system/user`、`team/team`、`message/*` 等大页拆回域级组件和对话框，减少 `pages/*` 的单文件体积。
- 继续按 Fluent 2 Web / Teams 视觉规范统一各页的 section、面板、列表和详情区，不再让单页自己决定深阴影和厚块布局。

## 2026-04-02 frontend-fluentV2 菜单治理页组件拆分

### 本次改动
- 将 `SystemMenuPage` 中的树节点与只读字段抽出为 `features/system/components/MenuTreeNode.tsx`，页面本体保留状态、编辑和删除流程。
- 继续把系统菜单页中的错误/提示块统一到 `PageStatusBanner`，减少页面内零散的反馈实现。
- 保持 `tsc` 与 `build` 通过，`system/menu` 页面结构开始向域组件化收口。

### 下次方向
- 继续把 `system/menu` 里的编辑表单、删除弹窗、详情摘要进一步拆成域组件。
- 再继续拆 `message/*`、`team/*`、`system/page`、`system/user` 的内部模块，减少大页体积并统一页面骨架。

## 2026-04-02 frontend-fluentV2 第8版全量页面清单与重页下沉

### 本次改动
- 新增并完善 `frontend-fluentV2/docs/page-inventory.md`，将 Vue 全量页面清单、docker 数据库中的 `menus / ui_pages` 以及当前 React 承接关系统一写成一份范围表，明确第 8 版不再只围绕 `system/menu`、`message/*`、`team/*` 推进。
- 重写 `frontend-fluentV2/README.md`，将说明口径更新到第 8 版：以 `Fluent UI React v9 + Fluent 2 Web + Teams` 为统一规范，以路由装配层 + 域组件 + 真实 API 为当前主线。
- 重写 `frontend-fluentV2/docs/architecture.md`，移除过期的 `features/session`、mock 主线和旧 query key 描述，改成当前真实认证、真实导航、Query 分层与页面范式。
- 将 `WorkspaceInboxPage` 完全切到 `features/workspace/components/InboxPanels.tsx`，收件箱页面现在只负责 URL 状态与查询编排，左中右三栏实现正式下沉到域组件。
- 将 `MessageCatalogPages.tsx` 的模板、发送人、收件组、记录、更多入口整体下沉到 `features/message/components/MessageCatalogWorkspaces.tsx`，`pages/message/MessageCatalogPages.tsx` 退回到轻量装配层。
- 将 `TeamPages.tsx` 的协作空间、成员、更多入口整体下沉到 `features/team/components/CollaborationWorkspaces.tsx`，`pages/team/TeamPages.tsx` 退回到轻量装配层。
- 将 `SystemMenuPage` 的右侧摘要、编辑、关联页面与删除确认整体下沉到 `features/system/components/SystemMenuPanels.tsx`，菜单治理页正式形成“树区 + 右侧域面板”的拆分结构。
- 已重新通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit`、`pnpm --dir frontend-fluentV2 build` 和 `go build ./...` 校验。

### 下次方向
- 继续将 `system/page`、`system/user`、`system/role`、`system/action-permission`、`system/api-endpoint`、`system/feature-package` 中残留的内联模块继续拆到 `features/system/components|dialogs|drawers`。
- 继续用真实菜单树、页面表和 Vue 文件清单回归，确认没有遗漏子页面、入口卡片和模块级交互。
## 2026-04-02

- 清空 `frontend-fluentV2/src/pages`，保留壳层与导航骨架，准备重新重构第 8 版页面体系。
- 删除 `frontend-fluentV2/docs` 专题文档，重置为干净目录，后续只保留新的重构文档。
- 清理了 `frontend-fluentV2/src/features/navigation/routes/core-routes.tsx`、`system-routes.tsx`、`team-routes.tsx` 的页面依赖，并将 `AppRouter` 收敛为壳层入口。
- 恢复最小可用登录链路和受保护壳入口，新增 `features/auth/LoginPage.tsx` 与 `features/home/ShellHomePage.tsx`，重新打通登录、默认首页和运行时菜单加载。
- `route-registry` 保留 `dashboard/console`、`system/menu`、`system/role`、`system/user` 作为本地已实现页，其余运行时页面继续由菜单承接并以占位提示显示。

## 2026-04-02 frontend-fluentV2 用户管理页治理化重构

### 本次改动
- 重写 `frontend-fluentV2/src/features/system/user/UserManagementPage.tsx`，将用户页收口为 Fluent 2 治理页：顶部摘要、筛选工具区、DataGrid 主区与响应式详情抽屉统一成一套正式结构。
- 详情抽屉改为 `概览 / 角色 / 菜单 / 功能包 / 权限诊断` 五个治理 Tabs，并接入 `features/access/access.service.ts` 里的真实用户治理 hooks，不再只停留在基础 CRUD。
- 菜单页签默认按 `default` 空间加载菜单树，角色分配、菜单保存、功能包查看和权限快照刷新都已接到现有接口；同时补齐 `UserRecord.isSuperAdmin` 与 access 归一化，统一前端类型口径。
- 已重新通过 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 继续把用户页中的角色分配与菜单编辑提炼成域级子组件，避免 `UserManagementPage` 再次膨胀成大文件。
- 后续如果要继续提升治理能力，优先补用户页的功能包编辑与更完整的菜单继承/派生可视化，而不是再扩一层独立页面。

## 2026-04-03 frontend Fluent 2 主题收敛

### 本次改动
- 调整 `frontend/src/assets/styles/core/tailwind.css` 的浅色与深色主题变量，统一页面底色、卡片底色、边框、hover/active 填充和主色强调，整体往 Fluent 2 的低噪、弱边框、单一强调色方向收敛。
- 更新 `frontend/src/assets/styles/core/el-light.scss` 和 `frontend/src/assets/styles/core/el-ui.scss` 的 Element Plus 全局覆写，让按钮、弹层、消息、下拉、日期范围和树选择的视觉基线与新主题变量保持一致。
- 微调 `frontend/src/assets/styles/core/app.scss` 的页面壳层阴影、侧栏分隔和徽标颜色，减少高亮白边和硬阴影带来的发亮感。
- 已通过 `pnpm --dir frontend build` 验证，未改动页面模板与业务逻辑，仅影响全局样式层。

### 下次方向
- 下一轮优先继续细化首页、侧栏、卡片和标签页的颜色分层，重点看实际页面的层级感和信息密度是否还偏“白”。
- 如果需要进一步贴近 Fluent 2，可以再补一轮状态色、空状态和表格斑马纹的统一规范，但不建议再引入第二套 UI 体系。

## 2026-04-05 后台 App 维度后端主链收口

### 本次改动
- 抽出 `backend/internal/pkg/appctx` 统一处理 `app_key` 归一化与请求读取，打断 `app / space / permissionrefresh / teamboundary` 之间的低层循环依赖，并让中间件、授权、作用域工具共用同一套 App 上下文入口。
- 将平台用户、平台角色、协作空间边界三条快照服务改为 App-aware，快照缓存、菜单展开、功能包过滤、隐藏菜单过滤和快照落库都按 `app_key` 分桶；同时保留变参兼容入口，避免旧调用一次性全部重写。
- 授权主链、权限刷新、页面运行时缓存和用户治理链路已经接入当前 App，上下文会沿 `Host -> App -> Space` 解析结果继续向下传，用户菜单树、权限诊断和运行时页面可见性都会限定到当前 App。
- 数据库自动迁移与 `backend/cmd/migrate` 都补了快照表 `app_key` 回填和复合主键修正，已重新通过 `go test ./...` 与 `pnpm --dir frontend build` 验证。

### 下次方向
- 继续把角色、协作空间、功能包管理页剩余仍走“默认 App 兼容调用”的保存链路改成显式 `app_key`，让前后端的保存与刷新语义彻底一致。
- 继续收口 `Menu Space` 的日常入口，只保留在 App 管理里的高级配置，并开始评估“菜单定义”和“空间布局”拆模，避免长期依赖 `menus + space_key` 混合承载两层职责。

## 2026-04-05 App 维度前后端继续收口与空间布局切换

### 本次改动
- 前端新增统一 `app-context` 运行时状态后，继续把 `system/app`、`system/role`、`system/user`、`system/team-roles-permissions`、`team/team`、`system/fast-enter` 等管理页切到 `managedApp` 主线；路由缺少 `app_key` 时优先由统一 hook 回填，不再在页内硬编码默认 `platform-admin`。
- 角色、用户、协作空间、协作空间角色、功能包相关的主要授权弹窗都清掉了 `props.appKey || 'platform-admin'` 兜底，缺少 App 上下文时会直接阻断并提示；同时 `frontend/src/api/collaboration_workspace.ts` 与 `frontend/src/types/api/api.d.ts` 也补齐了显式 `app_key` 透传所需的 helper 和类型。
- 后端继续把 App 作用域菜单读链收口到新模型：`appscope`、平台用户/角色快照、协作空间边界服务、App 统计与运行时导航版本戳已改读 `menu_definitions` / `space_menu_placements`，避免授权快照和运行时版本继续依赖旧 `menus` 主表。
- `space/service.go` 中空间菜单数量统计与“从默认空间初始化”已切到 `space_menu_placements`，空间初始化只复制布局记录，不再按空间复制菜单定义和功能包菜单关联；已重新通过 `go test ./...` 与 `pnpm --dir frontend build` 验证。

### 下次方向
- 继续把 `backend/internal/modules/system/menu`、`backend/internal/modules/system/page` 等仍直接写旧 `menus` 的服务切到“双表语义”或补齐双写，否则后续菜单编辑后的定义/布局数据仍可能出现主链和运行时不同步。
- 继续清理 `ui_pages.space_key` 的残余读写路径，并把页面候选、菜单候选和空间高级配置进一步统一到 `App -> Space -> Definition/Placement` 的稳定模型上。

## 2026-04-05 App 维度菜单定义页与空间布局页收尾

### 本次改动
- 修正 `backend/internal/modules/system/navigation/service_test.go` 的 `MenuService` 测试桩签名，并在 `backend/internal/modules/system/menu/service.go` 补上空间布局更新的 upsert 逻辑，保证菜单定义更新时，当前空间缺少 placement 记录也会直接落库，不再静默失败。
- 前端将 `system/menu` 明确收口为“菜单定义管理”主语义，新增定义页与布局页的文案区分、上下文切换入口，以及菜单备份按 App 过滤的逻辑；`menu-dialog` 默认不再暴露空间字段，只有从布局链路进入时才显示空间编辑。
- `system/menu-space` 进一步收口为“空间布局高级配置”，从这里进入菜单页时会显式带上 `app_key + spaceKey + layout=1`，页面管理页的提示语也同步改为“菜单定义/空间布局”双入口口径。
- 已重新通过 `go test ./...` 与 `pnpm --dir frontend build` 验证。

### 下次方向
- 继续把菜单页里仍偏“单页双职责”的部分拆干净，例如把空间级备份/恢复入口完全沉到 `system/menu-space`，让 `system/menu` 只保留定义管理和 App 级备份。
- 继续推进页面管理相关弹窗，把 `parent_menu_id`、空间暴露和 breadcrumb 预览全部显式解释成“菜单定义 + 空间布局”的语义，进一步减少旧 `menus/ui_pages.space_key` 兼容字段对前端表单的影响。

## 2026-04-05 App 维度空间布局备份职责回归

### 本次改动
- 将 `system/menu-space` 页补成“空间布局高级配置”的完整入口，新增当前空间布局的创建备份、查看备份、恢复备份和删除备份能力，并统一透传当前 `app_key + space_key`。
- `system/menu` 继续收口为菜单定义管理页，更多操作里移除了已不可达的空间备份分支，只保留 App 级定义备份与定义备份列表入口。
- 复用菜单备份弹窗组件时补充了可配置标题和提示文案，使定义备份与空间布局备份可以共用同一套组件而不混淆作用范围。
- 已重新通过 `pnpm --dir frontend build` 验证。

### 下次方向
- 继续把页面管理、breadcrumb 预览和父菜单候选统一解释成“菜单定义 + 空间布局”的模型，减少对旧 `menus` 和 `ui_pages.space_key` 兼容语义的依赖。
- 如果下一轮继续做彻底收尾，优先清理菜单与页面服务里仍残留的旧字段桥接逻辑，把前后端读写链完全锁到 `menu_definitions + space_menu_placements`。

## 2026-04-05 后端管理契约显式 app_key 收口

### 本次改动
- 将 `backend/internal/modules/system/space/handler.go` 与 `space/service.go` 的管理入口统一改为显式 `app_key`，`GetCurrent/List/ListHostBindings/SaveSpace/SaveHostBinding/InitializeFromDefault` 不再从上下文默认值或请求体覆盖里回退，缺失或不一致直接报错。
- 将 `backend/internal/modules/system/menu/handler.go`、`menu/service.go` 的备份链路补齐 App 边界校验，`list/detail/delete/restore` 统一依赖显式 `app_key`，并在返回中补充 `app_key` 与 `scope_origin`，防止跨 App 误操作。
- 将 `featurepackage/role/tenant/user` 等管理侧保存与列表链路里仍存在的 `ResolveManagedAppKey` 依赖替换为显式 `RequireRequestAppKey`，同时把功能包与页面 service 层的默认 App 兜底一并收紧。
- 同步修正了相关测试桩和菜单备份 scope 断言，`go test ./...` 已通过，后端契约硬化的主链已闭合。

### 下次方向
- 下一轮如果还要继续收尾，优先扫描 `apiendpoint` 和 `app` 这类仍保留默认 App 兼容的模块，判断是否也需要统一成显式 `app_key`。
- 如果要继续推进 App 体系最终收口，可以再做一轮只读审计，确认运行时路径与管理路径各自的 `app_key` 来源已经完全分离，没有遗留的静默默认值。

## 2026-04-05 前端 managedApp 强显式与页面/空间语义收口

### 本次改动
- `frontend/src/store/modules/app-context.ts`、`frontend/src/hooks/business/useManagedAppScope.ts` 与新增的 `frontend/src/hooks/business/managed-app-scope.ts` 已把 `runtimeApp` 和 `managedApp` 彻底拆开：管理态不再自动继承壳体运行时 App，只有显式路由参数或已选择的管理 App 才会进入管理链路。
- `system/api-endpoint` 页把“同步接口 / 未注册扫描 / 失效清理”明确按 shared/global 语义工作，不再向这些全局治理动作透传伪 `app_key`；应用级列表和保存仍继续按 `app_scope/app_key` 过滤。
- `system/menu`、`system/menu-space` 与菜单备份 helper 一起补齐了显式 App 边界：空间初始化、备份恢复、备份删除都显式透传 `app_key`，菜单定义页与空间布局页的职责也继续拉开。
- `system/page` 及 `page-entry/page-group/page-display-group` 三个弹窗去掉了 `currentSpaceKey: 'default'` 一类默认主归属语义，页面保存统一显式带 `app_key`，空间只保留为页面暴露范围与当前空间视角下的菜单/父页候选。
- 角色、协作空间、用户菜单授权相关弹窗继续收口为显式 App/Space 输入，避免再隐式读取全局 `currentSpaceKey` 串上下文；同时补了一条 `frontend/tests/managed-app-scope.test.ts` 作为 managedApp 解析的最小回归样例。
- 已重新通过 `pnpm --dir frontend build` 与 `go test ./...` 验证。

### 下次方向
- 如果还要继续压缩兼容层，优先评估是否彻底移除未被主页面引用的旧菜单权限弹窗，以及是否把页面管理页里的“当前空间视角”再抽成更明确的高级筛选器。
- 后续如果继续推多 App，可再补一轮围绕 `managedApp` 的交互细化，例如统一 App 选择器的空态提示、切换确认和跨页保持策略，但不建议再恢复任何默认 App 静默兜底。

## 2026-04-05 后端 app 管理契约最后收口

### 本次改动
- 将 `backend/internal/modules/system/app/service.go` 的 Host 绑定保存和列表改成显式 `app_key` 契约，空 `app_key` 直接报错，不再回落到默认 `platform-admin`；同时保留对请求体 `app_key` 的一致性校验，避免 body 偷换应用归属。
- `backend/internal/modules/system/app/handler.go` 统一改为从请求上下文显式读取 `app_key` 后再进入 Host 绑定读写链，管理侧 app 入口不再依赖默认 App。
- 重新检查了 `menu`、`page`、`space` 等管理入口，确认需要保留的显式 App 注入路径仍然稳定，空间布局与页面/菜单管理没有再引入新的默认 App 兜底。
- 已重新通过 `go test ./...` 验证。

### 下次方向
- 如果后续还要继续收口，优先再扫一遍 `apiendpoint` 和 `app` 以外的管理模块，确认是否还有类似“显式 app_key + body 兼容值”并存的入口。
- 如需进一步压缩兼容层，可以继续评估是否把当前保留的 App 选择器和空间高级配置入口再做一层更严格的显式跳转校验。

## 2026-04-05 前端最后一轮 App 收口

### 本次改动
- `frontend/src/store/modules/app-context.ts` 与 `frontend/src/hooks/business/useManagedAppScope.ts` 继续保持强显式语义，管理态不再从 runtime/default/first app 自动兜底；补充了 `frontend/tests/app-context-store.test.ts` 作为最小回归样例。
- `system/menu` 收成菜单定义管理主入口，移除了直接跳去“空间布局”的平级按钮，并清理了 `default` 作为表单/请求默认值的隐式 fallback；`system/menu-space` 则收成空间布局高级配置页，Host 绑定和空间保存都要求显式空间标识。
- `system/page` 的页面弹窗与 `page-entry/page-group/page-display-group` 三个子弹窗去掉了 `currentSpaceKey` 参与的默认回填语义，`space_key` 只保留为显式提交字段，不再作为隐式主归属来源。
- 删除了未再被主链使用的旧角色菜单弹窗 `frontend/src/views/system/role/modules/role-permission-dialog.vue`，避免后续误接回旧兼容链路；已重新通过 `pnpm --dir frontend build` 与 `pnpm exec node --import tsx --test tests/app-context-store.test.ts` 验证。

### 下次方向
- 如果还要继续压缩兼容层，建议下一轮只做只读审计，确认菜单页、空间页、页面页里是否还有残留的“当前空间视角”文案需要继续降级为高级配置说明。
- 后续若继续推进多 App，可再评估 App 选择器的空态交互和跨页保持策略，但不建议再恢复任何默认 App 静默兜底行为。

## 2026-04-05 App 收口补完与 global/shared API 语义校正

### 本次改动
- `backend/internal/modules/system/apiendpoint/handler.go` 与 `service.go` 将“失效 API 列表 / 失效清理”收回到 global/shared 维护语义：不再要求显式 `app_key`，空 `app_key` 时按全局同步源做失效诊断，应用级 API 列表与保存仍保留 `app_scope/app_key` 约束。
- `frontend/src/api/system-manage.ts`、`frontend/src/views/system/api-endpoint/index.vue` 同步改成全局治理口径，去掉失效 API 列表的伪 `appKey` 入参，并把“同步 API / 未注册 API / 清理失效 API”文案明确成全局维护动作。
- `system/menu`、`system/menu-space`、`system/page` 继续补齐显式 `app_key` 守卫；当管理页缺失 App 上下文时，页面不再偷偷回落到默认 App，而是直接阻断加载并提示先从应用管理选择 App。
- 菜单定义页和空间高级配置页剩余的 `default` 隐式空间兜底继续清理，`menu-dialog`、`page/index`、`api-endpoint/index` 的保存与加载链已统一到“显式 App + 显式空间视角”的语义。
- `frontend/src/views/system/app/index.vue` 继续去掉应用管理页里的默认空间硬编码：列表、概览和 Host 绑定展示在未配置时统一显示“未设置”，表单默认值改成显式解析当前 App / 当前空间结果，不再把 `'default'` 静默写回管理请求。
- `frontend/src/api/system-manage.ts` 与 `frontend/src/components/business/layout/AppContextBadge.vue` 的剩余显示层兜底也同步收紧：空间初始化返回不再假设 `sourceSpaceKey=default`，上下文徽章在无空间时统一显示“未选择空间”。
- `backend/internal/modules/system/app/service.go` 将 `SaveApp` / `SaveHostBinding` 的 `default_space_key` 改为显式必填，空值直接报错，不再在管理链路里静默回落到默认空间；`frontend/src/views/system/app/index.vue` 也同步增加了保存前校验。

### 下次方向
- 如果还要继续做最终审计，建议重点扫一遍 `app/service.go`、`user/service.go` 和页面 breadcrumb/active menu 推导链，确认剩余默认 App 仅存在于运行时兼容而不再泄漏到管理契约。
- 当前这轮已经把 App 管理、菜单定义、空间布局、页面管理和 API 注册的主链收口到统一语义，后续更适合做只读核查和细节润色，而不是继续扩散结构性改动。

## 2026-04-05 API 注册链 app_scope 缺列修复

### 本次改动
- `backend/internal/pkg/database/database.go` 在启动自动迁移链路中补上了 `api_endpoints.app_scope` 与 `api_endpoints.app_key` 的显式补列和回填，不再只依赖模型自动迁移命中最新字段，避免现网库在 API 注册同步阶段因缺列直接失败。
- `backend/cmd/migrate/main.go` 的 `finalizeAPIEndpointSchema` 同步加入相同的补列与回填逻辑，确保手工执行迁移命令时也能把 `api_endpoints` 升级到支持 App 作用域的结构。
- `backend/internal/pkg/apiregistry/registry.go` 将 upsert 探测里的 `First` 改成 `Find + RowsAffected`，避免同步注册时把“未找到旧记录”的正常分支刷成 `record not found` 噪音日志。

### 下次方向
- 建议继续观察一次真实启动日志，确认 API 注册链现在只剩真正的 schema/数据错误，不再混入“未命中旧记录”的探测噪音。
- 如果后续还要继续强化初始化鲁棒性，可以把类似的关键新增列补齐逻辑统一抽到 schema finalize 层，避免字段升级只分散在回填函数里。
## 2026-04-05 运行时菜单空间初始化补齐 app 上下文

- 前端 `menu-space` store 在登录后首轮 `fetchUserInfo -> refreshRuntimeConfig` 阶段，先于 `runtime/navigation` 建立 `runtimeAppKey`，导致运行时菜单空间配置同步直接报“缺少 app 上下文”，进而连带动态路由注册失败。
- 在 `frontend/src/store/modules/menu-space.ts` 增加运行时 App 自举：当 `runtimeAppKey` 为空时，先调用 `fetchGetCurrentApp()` 按当前 Host 解析运行时 App，并写回 `appContextStore.setRuntimeAppKey(...)`，再继续拉取菜单空间列表、Host 绑定和当前空间解析。
- 这次修复只作用于运行时初始化链路，不回退前面已经收口好的管理态显式 `app_key` 语义；`managedApp` 与 `runtimeApp` 仍保持分离。
- 已通过 `pnpm --dir frontend build` 验证。
## 2026-04-05 管理页缺失 app_key 自动继承当前管理 App

- 运行时数据库已执行一次最新迁移：在 `backend/` 下运行 `go run ./cmd/migrate`，把当前测试库升级到包含 `apps`、`app_host_bindings`、菜单定义/空间布局以及最新 API 注册结构的状态。
- 前端管理页继续报 `app_key is required` 的根因，不是后端缺数据，而是 `useManagedAppScope` 过度只信路由 `app_key`。当系统管理页 URL 没带 `app_key` 时，它会把 `managedApp` 清空，导致菜单、页面、功能包等页面发出 `app_key=` 的空请求。
- 在 `frontend/src/hooks/business/useManagedAppScope.ts` 调整为优先解析：
  - 路由 `app_key`
  - 已选的 `managedAppKey`
  - 当前运行时 `runtimeAppKey`
- 当管理页 URL 未携带 `app_key` 但本地已有管理 App 或运行时 App 时，自动把该 `app_key` 写回当前路由，避免继续出现空 `app_key` 请求。
- 已通过 `pnpm --dir frontend build` 验证。

## 2026-04-05 管理态 App 去 URL 与 fresh 重建核验

### 本次改动
- `frontend/src/store/modules/app-context.ts`、`frontend/src/hooks/business/useManagedAppScope.ts` 改成纯 store/persist 驱动的管理态 App 上下文：`effectiveManagedAppKey = managedAppKey || runtimeAppKey`，`ensureManagedAppKey()` 会把运行时 App 提升为管理态 App，系统管理页不再依赖 URL 里的 `app_key`。
- `frontend/src/store/modules/menu-space.ts` 改成按 App 分片持久化运行时菜单空间配置和当前空间覆盖值，不再用一份全局 `currentSpaceKey` 复用所有 App；后续即使多个子域名 / 多个 App 共用同一后台壳，也不会把空间状态串到别的 App。
- 重新执行了 `go run ./cmd/migrate --fresh`，测试库按“当前 schema + 当前 seed”重建，不再回放历史 named migrations；重建后已抽查确认：
  - `apps` 仅存在默认 `platform-admin`
  - `menu_spaces` 仅存在 `platform-admin/default`
  - `menu_definitions = 24`
  - `space_menu_placements = 24`
  - `ui_pages = 13`
  - `feature_packages = 5`
  - `message_templates = 3`
  - `访问链路测试` 菜单、`system.access_trace.manage` 页面、`ui.fast_enter` 配置都已存在
- 已通过 `pnpm --dir frontend exec node --import tsx --test tests/app-context-store.test.ts`、`pnpm --dir frontend build`、`go test ./...` 验证。

### 下次方向
- 如果下一轮继续“彻底不留历史中间层”，优先把默认菜单 seed 从旧 `menus` 改成直接写 `menu_definitions + space_menu_placements`，不再依赖 `RefreshNavigationDefinitionsFromLegacyMenus()` 作为 fresh 初始化中间桥。
- 运行时 `App` 仍建议长期坚持 `Host -> App -> Space` 解析，管理态 `App` 则继续只存在于 store / storage 或 cookie，不再恢复任何 URL `app_key` 兼容入口。
## 2026-04-05 默认菜单 seed 直写新表

### 本次改动
- `backend/cmd/migrate/main.go` 的默认菜单初始化已从“先写旧 `menus`，再 `RefreshNavigationDefinitionsFromLegacyMenus()` 回填”改成直接写 `menu_definitions + space_menu_placements`。`ensureDefaultMenuSeedByName / syncDefaultMenuSeedByName / ensureMenuSeed / syncMenuSeed` 现在全部以 `MenuDefinition` 为核心，不再把 fresh 默认菜单落到旧表。
- 初始化主链不再在 fresh 过程中调用 `RefreshNavigationDefinitionsFromLegacyMenus()`；访问链路测试菜单 `ensureAccessTraceNavigationSeed()` 也同步切到新模型，菜单定义、空间布局、页面父菜单和功能包菜单绑定都直接引用菜单定义 ID。
- `backend/internal/pkg/permissionseed/ensure.go` 的默认功能包菜单绑定已改成查询 `menu_definitions`，不再通过旧 `menus` 按名称找菜单。
- 重新执行 `go run ./cmd/migrate --fresh` 后抽查确认：
  - `menus = 0`
  - `menu_definitions = 24`
  - `space_menu_placements = 24`
  - `feature_package_menus = 32`
  - 默认 seed 已完全由新表承接
- 已通过 `go test ./...` 与 fresh 重建验证。

### 下次方向
- 旧 `menus` 表现在只剩历史审计/回滚参考用途；如果后续确认不再需要 legacy 菜单备份恢复，可继续把 `RefreshNavigationDefinitionsFromLegacyMenus()` 和相关 legacy 转换 helper 下沉为纯维护工具，进一步收缩主代码面。
- 页面 seed 目前已经通过菜单定义 ID 建立父菜单，但 `ui_pages.space_key` 仍作为历史字段存在；后续如要继续收口，可把 seed 写入时对该字段的依赖进一步降到最低。

## 2026-04-06 页面可见性模型继续收口

### 本次改动
- `backend/internal/modules/system/page/service.go` 将页面管理与运行时菜单映射进一步收口到“当前 App + 当前空间”的菜单视图：`List`、`ListOptions`、运行时访问上下文、页面 hydrate 和未注册页面推断现在都会显式按空间读取 `menu_definitions + space_menu_placements` 物化后的菜单，避免同一菜单定义在多个空间下被默认映射覆盖。
- 页面同步推断规则继续贴近新模型：未注册页面如果推断到父菜单，会生成 `inner + inherit`；否则默认生成 `standalone + app`，不再把无父菜单页面回退成旧式全局页语义。
- `frontend/src/views/system/page/modules/page-group-dialog.vue` 与 `page-display-group-dialog.vue` 已显式引入 `visibilityScope`：
  - 逻辑分组在未挂载时可选 `App 全局 / 指定空间`
  - 挂到菜单或父分组后自动继承上级空间
  - 普通分组默认 `App 全局`，只有切到指定空间时才保存 `space_keys`
- `frontend/src/views/system/page/index.vue` 的页面类型筛选与统计口径已纳入 `standalone`，`fetchSyncPages` 也改成显式传 `app_key`，不再留空请求。
- `frontend/src/api/system-manage.ts` 的页面同步与未注册页面 helper 改成必填 `appKey`，继续压缩管理态的隐式默认 App 行为。

### 验证
- `go test ./...`
- `pnpm --dir frontend build`
- 后续执行 `go run ./cmd/migrate --fresh` 时，应按最新页面可见性模型直接写入 fresh 数据，不再依赖旧 `menus` 或页面 `space_key` 主归属语义。

### 下次方向
- 如果要继续把旧兼容层压到最低，优先再审一轮 `page/runtime_cache.go` 与运行时 breadcrumb 推导，确认它们只认“页面定义 + 当前空间布局”而不再假设单一默认空间。
- 页面 seed 目前已经进入“App 归属 + visibility_scope + page_space_bindings”模型，后续可继续减少 `ui_pages.space_key` 在审计和返回结构中的存在感。

## 2026-04-06 页面种子直写 visibility_scope 与 page_space_bindings

### 本次改动
- `backend/internal/pkg/permissionseed/seeds.go` 的 `PageSeed` 新增 `VisibilityScope`、`SpaceKeys`，默认页面种子不再隐含依赖历史 `space_key` 语义；`display.system_pages` 与 `workspace.user_center` 已改成明确的 App 级页面。
- `backend/cmd/migrate/main.go` 的 `syncUIPageSeed(...)` 已直接按新模型写入：
  - 页面固定归属 `platform-admin`
  - `visibility_scope` 按 `page_type + 父菜单/父页面` 归一化
  - `ui_pages.space_key` 不再作为 seed 主字段写入
  - `page_space_bindings` 仅在 `visibility_scope = spaces` 时生成
  - `global` 页种子会自动清理父菜单、父页面与 `active_menu_path`
- `frontend/src/router/core/ManagedPageProcessor.ts` 的运行时空间解析已改成优先读取 `visibilityScope + spaceKeys`，不再让旧 `page.spaceKey/meta.spaceKey` 抢主语义；`global/app` 页面统一跟随当前运行时空间，`spaces` 页面按已绑定空间解析，`inherit` 页面继续走父页面/父菜单链。
- `frontend/src/views/system/page/index.vue` 不再在页面管理页加载后自动选中某个菜单空间作为主视角；页面列表默认按 App 管理，空间仅保留为次级筛选与访问辅助，不再隐式绑到默认空间。

### fresh 数据核对
- 已重新执行 `go run ./cmd/migrate --fresh`，并抽查确认：
  - `platform-admin/default` 正常存在
  - `ui_pages` 当前为 `display_group(app) = 1`、`global(app) = 1`、`inner(inherit) = 11`
  - `page_space_bindings = 0`
  - `workspace.user_center` 已不再挂父菜单
  - `display.system_pages` 已改为 App 级可见

### 验证
- `go test ./...`
- `pnpm --dir frontend build`
- `go run ./cmd/migrate --fresh`

### 下次方向
- 如果后续要继续压缩历史字段，可以把 `ui_pages.space_key` 在返回结构里的兼容展示也继续降权，只保留 `visibility_scope + space_keys` 作为页面主语义。
- 当前默认 seed 里还没有 `standalone + spaces` 示例页；如果后续要联调空间独立页，可以补一条最小示例 seed，专门验证 `page_space_bindings` 链路。

## 2026-04-06 App 默认空间自动生成与页面主语义继续收口

### 本次改动
- `frontend/src/views/system/app/index.vue` 的 App 新建流程继续收口：创建时不再携带空的 `default_space_key`，并在表单里明确提示“系统会自动创建当前 App 自己的默认空间 `default`”；编辑态才允许手动切换默认空间。
- `frontend/src/types/api/api.d.ts` 同步更新说明：
  - `AppSaveParams.default_space_key` 只在编辑 App 时有意义，创建由服务端自动生成默认空间。
  - 页面 `space_key` 明确降级为历史兼容字段，管理端主语义统一看 `visibility_scope + space_keys`。
- 重新执行 fresh 后核查默认数据，确认当前主链已经符合最终模型：
  - `apps` 中 `platform-admin` 为 `multi + default`
  - `menu_spaces` 中同时存在 `platform-admin/default` 与 `platform-admin/ops`
  - `ui_pages` 中 `workspace.ops_console` 为 `standalone + spaces`
  - `page_space_bindings` 中存在 `workspace.ops_console -> ops`
  - `ui_pages.space_key` 对现有 fresh 页面已经全部为空，仅保留历史兼容字段语义

### 验证
- `go test ./...`
- `go run ./cmd/migrate --fresh`
- `pnpm --dir frontend build`
- fresh 后抽查：
  - `select app_key, space_mode, default_space_key from apps`
  - `select app_key, space_key, is_default from menu_spaces`
  - `select page_key, page_type, visibility_scope, space_key from ui_pages`
  - `select p.page_key, b.space_key from page_space_bindings ...`

### 下次方向
- 如果后续还要继续压缩历史字段，可继续降低 `UIPage.SpaceKey` 在后端 DTO 和管理端类型中的可见性，只在内部兼容桥接保留。
- 当前运行时已不提供空间切换入口；后续若需要多空间站点联调，优先验证 `Host -> App -> Space` 的解析与菜单/页面加载，不再往壳层回塞空间选择逻辑。

## 2026-04-06 根文档收紧为单前端主线

### 本次改动
- 整理了根目录三份当前生效文档：
  - `AGENTS.md`
  - `PROJECT_FRAMEWORK.md`
  - `FRONTEND_GUIDELINE.md`
- 删除了 `AGENTS.md` 中混入的无关英文定位文本，并统一了排版、缩进和措辞。
- 三份文档都进一步明确：仓库当前只保留 `frontend/` 这一条有效前端主线，后续后台页面、壳层、组件、路由、状态与样式演进都以 `frontend/` 为唯一落点，不再并行发展第二套后台前端工程或 UI 体系。
- 同时补强了“单主线、不并行孵化第二套后台设计语言、不用临时新工程绕开现有主线约束”的书面说明，避免后续协作时再次分叉。

### 验证
- 本次仅调整文档与约束说明，未改动业务代码或构建配置，因此未额外执行构建与测试。

### 下次方向
- 如果后续还要继续收紧协作边界，可以把 App/菜单空间/页面模型的最新约束再补一份精简版总览文档，放在根目录或 `docs/` 下，供多人协作时快速对齐。
- 2026-04-06：补完中断前遗留的页面管理收口改动，修正 [page-entry-dialog.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/page/modules/page-entry-dialog.vue)、[page-group-dialog.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/page/modules/page-group-dialog.vue)、[page-display-group-dialog.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/page/modules/page-display-group-dialog.vue) 的 `visibilityScope + spaceKeys` 表单逻辑与提交流程，移除 [api.d.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/types/api/api.d.ts) 中 `PageSaveParams.space_key` 的管理态主语义；已重新执行 `pnpm --dir frontend build` 验证通过。

## 2026-04-06 页面运行时空间桥接继续降权

### 本次改动
- `frontend/src/router/core/ManagedPageProcessor.ts` 继续降低旧页面空间兼容字段的优先级：运行时页面空间解析现在优先只读取后端显式返回的 `visibilityScope + spaceKeys`，不再把 `page.meta.visibilityScope / page.meta.spaceKeys` 当主语义来源；运行时 `meta.spaceKey` 仍保留为桥接结果，供现有跳转与通知组件继续消费。
- `backend/internal/modules/system/page/service.go` 的访问追踪链不再把“未显式传空间”硬编码回退到固定 `default`，而是按当前 App 的默认空间解析；页面水合链在页面自身没有显式空间绑定时，也不再让页面记录主动宣称 `default` 空间。

### 验证
- `go test ./...`
- `pnpm --dir frontend build`

### 下次方向
- 如果继续收尾，可再审一轮运行时消费 `route.meta.spaceKey` 的组件，确认它们只把该字段当桥接结果，而不是反向推导页面主语义。
- `UIPage.SpaceKey` 目前已基本退出 fresh 数据与管理态主视图；后续可继续压低它在后端返回结构中的存在感，只保留内部兼容桥接。

## 2026-04-06 菜单与页面管理改为页内独立切换 App

### 本次改动
- `frontend/src/views/system/menu/index.vue` 新增页内 `App` 选择器；菜单定义管理不再依赖“先去应用管理选 App”。切到空间布局模式后，还可以继续选择当前 `App` 下的菜单空间；如果未选空间，则菜单树和页面候选保持空结果，并明确提示“请选择当前要配置的菜单空间”。
- `frontend/src/views/system/page/index.vue` 新增页内 `App` 选择器；页面管理现在先按 `App` 管理，再按空间做二级筛选与预览，不再依赖全局已选 `App`。同时补了独立页 `standalone` 的展示文案与空间筛选逻辑。
- `frontend/src/views/system/menu-space/index.vue` 新增页内 `App` 选择器；“高级空间配置”不再隐藏在应用管理后的隐式上下文里，而是可以直接在页面内切换 `App`，分别查看和维护该 `App` 下的空间、Host 绑定和空间模式。
- `backend/internal/pkg/permissionseed/seeds.go` 去掉了 `MenuSpaceManage` 默认菜单上的 `isHide` 元信息，让“高级空间配置”重新作为正常系统菜单可见。
- `backend/cmd/migrate/main.go` 修复旧命名迁移 `20260330_menu_space_foundation` 里对 `menu_spaces(space_key)` 全局唯一的过时假设；原先的 `ON CONFLICT (space_key)` 已改成 `WHERE NOT EXISTS`，兼容当前 `app_key + space_key` 模型，避免现网库在执行 `go run ./cmd/migrate` 时再次卡死。

### 验证
- `go run ./cmd/migrate`
- `go test ./...`
- `pnpm --dir frontend build`

### 下次方向
- 如果后续继续优化交互，可再把菜单管理与高级空间配置页之间的入口关系做得更清晰，例如在菜单页布局模式下补充更明确的“当前 App / 当前空间”状态提示。
- 若需要继续收紧历史兼容层，优先只读审计页面运行时仍消费 `spaceKey` 的组件，确认它们仅把该字段当桥接结果，不再作为页面主语义来源。

## 2026-04-06 管理页 App 选择改为页面内独立记忆

### 本次改动
- `frontend/src/hooks/business/useManagedAppScope.ts` 不再把管理态 `App` 选择写入全局 `appContextStore.managedAppKey`。当前实现改为按页面独立持久化：每个管理页都会使用自己的本地存储键记住最近一次选择的 `App`，页面之间不再互相联动。
- `frontend/src/hooks/business/managed-app-scope.ts` 新增页面级存储键解析逻辑，优先使用显式键，其次使用当前路由名/路径生成稳定键名，确保菜单管理、页面管理、角色管理、协作空间管理等页面各自记住自己的 `App` 选择。
- `frontend/src/views/system/app/index.vue` 去掉了从应用管理页直接写全局 `managedAppKey` 的逻辑。现在在应用管理页选择某个 `App`，只影响当前页自身的详情、Host 绑定和空间配置视图，不会把其它系统管理页一起切到相同 `App`。

### 验证
- `pnpm --dir frontend build`

### 页面规则
- 只需要 `App` 选择、不需要 `Space` 选择：角色管理、用户管理、协作空间管理、协作空间角色与权限、功能包管理、快捷应用管理。这些页面主语义是 `App` 级资源或授权边界，不是空间级布局。
- 需要 `App` 选择，并按需再加 `Space` 选择：菜单管理、页面管理、高级空间配置、访问链路测试。因为这些页面会直接受菜单布局或页面空间可见性影响，空间筛选对结果有实质作用。

## 2026-04-06 Workspace 权限迁移继续收口

### 本次改动
- `backend/internal/pkg/authorization/context.go` 开始从 query、`X-Target-Workspace-Id` 和 JSON body 解析 `target_workspace_id`，`backend/internal/pkg/authorization/authorization.go` 则把 `explicit_target_workspace` 变成真实授权约束：personal workspace 发起平台代管请求时必须显式传目标 workspace，且目标必须是当前用户可访问的 active team workspace。
- `backend/internal/pkg/teamboundary/service.go` 已改成优先从 `workspace_feature_packages` 读取 team workspace 的功能包绑定；只有新表为空时才回落旧 `collaboration_workspace_feature_packages`。
- `backend/internal/modules/system/featurepackage/service.go` 在协作空间功能包覆盖、按功能包分配协作空间、删除功能包时开始同步维护 `workspace_feature_packages`，让 feature package 写链从 tenant 表逐步转向 workspace 双写。
- `backend/internal/modules/system/tenant/handler.go` 的协作空间成员角色读取开始优先使用 `workspace_role_bindings`；配置成员角色时同步写入旧 `user_roles` 与新 `workspace_role_bindings`，避免只改读链不改写链。
- `backend/internal/modules/system/permission/service.go` 已把权限键 `app_key` 推导从“固定默认 App”改成“按权限键前缀和模块编码推导”，并继续补齐 `data_policy / allowed_workspace_types`。
- `frontend/src/store/modules/workspace.ts` 新增 `switchWorkspace(workspaceId)` 作为唯一切换入口；`frontend/src/components/core/layouts/art-header-bar/widget/ArtTenantSwitcher.vue` 已升级为统一 workspace 切换器，开始同时展示个人工作空间和协作空间；`frontend/src/components/core/layouts/art-header-bar/widget/ArtUserMenu.vue` 把“进入平台管理”降级成“切换到个人工作空间”快捷动作。

### 验证
- `go test ./internal/modules/system/auth ./internal/modules/system/workspace ./internal/api/router ./internal/pkg/authorization ./internal/pkg/teamboundary ./internal/modules/system/featurepackage ./internal/modules/system/permission ./internal/modules/system/tenant`
- `pnpm --dir frontend build`
- `pnpm --dir frontend lint`
  - 未通过，失败点在前端依赖环境的 `minimatch/brace-expansion`：`TypeError: (0 , brace_expansion_1.expand) is not a function`。本次改动本身已通过 `vue-tsc --noEmit && vite build`，但 lint 运行环境需要单独修复依赖版本。

### 下次方向
- 继续把 role package、team role boundary 和更多平台治理接口往 workspace 主链收口，缩小旧 `tenant` 表的只读兼容窗口。
- 对多协作空间批量操作接口补齐 `target_workspace_id` 与 path/body 的一致性规则，避免平台代管语义继续分叉。
- 继续把更多仍依赖 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 的管理页切到显式消费 `workspaceStore`，减少前端兼容层停留时间。

## 2026-04-06 Workspace 权限迁移剩余 30% 持续收口

### 本次改动
- `backend/internal/pkg/workspacefeaturebinding/service.go` 新增 personal/team workspace 功能包绑定 helper，开始统一承担 `workspace_feature_packages` 的 workspace 主读链能力；`backend/internal/pkg/appscope/scoping.go` 的 `/users/:id/packages` 与协作空间功能包读链已改成 workspace 优先、legacy 回退，用户平台功能包写入也开始先写 personal workspace，再镜像到旧 `user_feature_packages`。
- `backend/internal/pkg/platformaccess/service.go`、`backend/internal/pkg/permissionrefresh/service.go`、`backend/internal/modules/system/permission/service.go`、`backend/internal/modules/system/featurepackage/service.go` 已把平台快照、刷新链、统计口径和功能包影响预览改成 workspace 优先，不再只统计 `user_feature_packages / collaboration_workspace_feature_packages`。
- `backend/internal/modules/system/featurepackage/handler.go` 的 `GET /api/v1/feature-packages/:id/teams` 已按当前 personal workspace 的可代管范围过滤结果，不再直接返回全部 team 绑定。
- `backend/internal/pkg/workspacerolebinding/service.go` 新增基于 role code / role id 的 workspace 查询 helper；`backend/internal/modules/system/space/util.go`、`backend/internal/modules/system/page/runtime_cache.go`、`backend/internal/modules/system/system/message_service.go` 开始优先读取 `workspace_role_bindings`，并保留 `collaboration_workspace_members / user_roles` 作为兼容回退。
- `backend/internal/modules/system/user/handler.go` 的 `/users/:id/packages` 响应已补 `binding_workspace_id / binding_workspace_type`，前端 `frontend/src/types/api/api.d.ts` 与 `frontend/src/api/system-manage.ts` 已同步补齐类型；`frontend/src/views/system/user/modules/user-package-dialog.vue`、`frontend/src/views/system/role/modules/role-package-dialog.vue` 也改成显式说明“个人工作空间平台功能包”和“平台角色目录功能包”语义。

### 验证
- `go test ./internal/pkg/authorization ./internal/modules/system/featurepackage ./internal/modules/system/role ./internal/modules/system/user ./internal/modules/system/tenant ./internal/pkg/teamboundary ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/modules/system/permission ./internal/modules/system/auth ./internal/modules/system/workspace ./internal/api/router ./internal/pkg/appscope ./internal/pkg/workspacerolebinding ./internal/pkg/workspacefeaturebinding ./internal/modules/system/space ./internal/modules/system/page ./internal/modules/system/system`
- `pnpm --dir frontend build`

### 下次方向
- 继续把仍带 `collaborationWorkspaceId / collaboration_workspace_ids` 的平台治理接口按接口族接入同一套 `target workspace` 精确校验，优先处理 tenant 平台治理链和剩余 role 边界接口。
- 继续清理仍直接消费 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 的页面，把协作空间页面的上下文表达进一步收紧到 workspace 语义。
- 如需恢复 `pnpm lint` 作为硬门槛，先单独修复前端 `minimatch / brace-expansion` 依赖异常。

## 2026-04-06 Workspace 权限迁移继续推进

### 本次改动
- `backend/internal/modules/system/tenant/handler.go` 与 `backend/internal/modules/system/tenant/module.go` 开始把 `/api/v1/collaboration-workspaces/:id*` 这组平台治理接口接入 `RequirePersonalWorkspaceTargetTeam` 精确校验；`Get/Update/Delete`、成员管理、协作空间角色列表、协作空间菜单/功能边界及来源接口现在都会先把 path `:id` 解析为目标 team workspace，再校验当前 personal workspace 是否有权代管。
- `backend/internal/modules/system/page/service.go` 的平台访问轨迹角色读取改成 personal workspace 优先，且 legacy 回退显式限定 `user_roles.collaboration_workspace_id IS NULL`；这修掉了平台访问轨迹可能误把协作空间角色混进平台角色列表的问题。

### 验证
- `go test ./internal/modules/system/tenant ./internal/modules/system/page ./internal/pkg/authorization ./internal/api/router`

### 下次方向
- 继续把 tenant 平台治理链中剩余未接 helper 的接口补齐，并把批量目标接口收成 `RequirePersonalWorkspaceTargetTeams`。
- 继续检查其他平台读取链里是否还存在“平台上下文误混协作空间角色”的 legacy 查询。

## 2026-04-06 Workspace 权限迁移继续推进二

### 本次改动
- `backend/internal/pkg/authorization/authorization.go` 的 `getEffectiveActiveRoleIDs` 已切到 workspace 优先：platform 场景先读 `personal workspace -> workspace_role_bindings`，team 场景先读 `team workspace -> workspace_role_bindings`；只有 workspace 绑定为空时才回退旧 `user_roles`。
- 同时对 legacy 平台回退增加了 `roles.collaboration_workspace_id IS NULL` 与 `roles.deleted_at IS NULL` 约束，避免核心判权链把协作空间角色误当成平台角色。

### 验证
- `go test ./internal/pkg/authorization ./internal/modules/system/tenant ./internal/modules/system/page ./internal/api/router`

### 下次方向
- 继续排查其余平台快照、诊断、统计链中仍直接读取 `user_roles` 的位置，把平台角色读取统一收敛到 personal workspace 优先。

## 2026-04-06 Workspace 权限迁移继续推进三

### 本次改动
- `backend/internal/modules/system/user/repository.go` 已把用户角色读取和平台用户角色回填的 legacy fallback 收紧到真正的全局平台角色：当 `tenantID == nil` 时，除 `user_roles.collaboration_workspace_id IS NULL` 外，还要求 `roles.collaboration_workspace_id IS NULL` 且 `roles.deleted_at IS NULL`。
- `backend/internal/pkg/platformaccess/service.go` 与 `backend/internal/pkg/permissionrefresh/service.go` 也同步收紧了平台侧 legacy role fallback，保证平台快照、刷新链和用户详情不会把协作空间角色误当成平台角色。

### 验证
- `go test ./internal/modules/system/user ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/pkg/authorization ./internal/api/router`

### 下次方向
- 继续排查剩余直接读取 `user_roles` 或 `collaboration_workspace_members.role_code` 的运行时/诊断链，优先处理仍可能影响平台统计、消息选人和权限诊断的路径。

## 2026-04-06 Workspace 权限迁移继续推进七

### 本次改动
- `frontend/package.json` 已把 `@typescript-eslint/eslint-plugin`、`@typescript-eslint/parser` 和 `typescript-eslint` 锁到 `8.50.0`，规避 `8.56.1 -> minimatch@10` 在当前安装策略下触发的 `brace-expansion` 解析异常。
- `frontend/.npmrc` 新增 `node-linker=hoisted`，并重新安装 `frontend/node_modules`；前端依赖树已恢复到稳定可执行状态，`pnpm --dir frontend build` 可以再次正常通过。
- `frontend/eslint.config.mjs` 与 `frontend/.prettierignore` 已把 `.codex-tmp/**` 排除，避免 Codex 临时产物继续污染 lint / prettier 结果。
- 本轮相关前端文件已做定向 `prettier` 与 `eslint` 清理，`frontend/src/router/guards/beforeEach.ts`、`frontend/src/components/business/layout/AppContextBadge.vue`、`frontend/src/components/core/layouts/art-notification/index.vue`、`frontend/src/views/message/modules/*`、`frontend/src/views/team/team-members/index.vue`、`frontend/src/views/workspace/inbox/index.vue`、`frontend/src/views/dashboard/console/index.vue` 的定向 lint 已通过。

### 验证
- `pnpm --dir frontend build`
- `pnpm exec eslint eslint.config.mjs src/router/guards/beforeEach.ts src/components/business/layout/AppContextBadge.vue src/components/core/layouts/art-notification/index.vue src/views/message/modules/useMessageWorkspace.ts src/views/message/modules/message-dispatch-console.vue src/views/message/modules/message-recipient-group-console.vue src/views/message/modules/message-sender-console.vue src/views/message/modules/message-template-console.vue src/views/team/team-members/index.vue src/views/workspace/inbox/index.vue src/views/dashboard/console/index.vue`
- `pnpm --dir frontend lint`
  - 现在已经能正常执行，不再是依赖环境异常；当前失败点切换为仓库内既有的全量规则债，主要集中在 `frontend/src/api/system-manage.ts`、`frontend/src/api/collaboration_workspace.ts`、若干系统管理页和历史组件中的 Prettier/`no-unsafe-optional-chaining` 旧问题。

### 下次方向
- 继续清理全仓前端存量 lint 债务，优先处理 `frontend/src/api/system-manage.ts`、`frontend/src/api/collaboration_workspace.ts` 与系统管理相关页面，争取把 `pnpm --dir frontend lint` 重新拉回正式硬门槛。
- 继续做 workspace 迁移的最后一轮全链路排查，只保留 `tenant` 在旧 team API 与 header 桥接层的兼容职责。

## 2026-04-06 Workspace 权限迁移一次性收口与通过性检查

### 本次改动
- `frontend/src/router/guards/beforeEach.ts` 已去掉对 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 的公共层直接依赖，`preferredTenantId` 现在只使用后端兼容返回的 `current_collaboration_workspace_id` 作为初始化提示，不再把 tenant store 作为壳层主上下文来源。
- 已完成一次性通过性检查：后端权限主链继续保持 `personal workspace / team workspace` 分流，前端 `workspaceStore` 继续作为唯一授权上下文源；本轮没有新增路由、没有新增 schema、没有引入新的 workspace 语义分叉。
- 当前工作区中 `.agents/*`、`docs/superpowers/*` 等并存脏改动也纳入了本次边界核对，但它们不改变 workspace 权限主线判断；本轮只把它们标记为“非 workspace 主线但已纳入检查范围”的并存改动。

### 验证
- `go test ./internal/modules/system/auth ./internal/modules/system/featurepackage ./internal/modules/system/permission ./internal/modules/system/tenant ./internal/modules/system/user ./internal/modules/system/page ./internal/modules/system/space ./internal/modules/system/system ./internal/pkg/authorization ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/pkg/teamboundary ./internal/api/router`
- `pnpm --dir frontend lint`
- `pnpm --dir frontend build`

### 通过性检查结论
- 通过：后端测试、前端 lint、前端 build 全部通过。
- 通过：`personal workspace` 平台角色/平台功能包 与 `team workspace` 协作空间角色/协作空间功能包 的运行时主链已经固定为 `workspace_*` 优先、legacy 回退。
- 通过：workspace 切换后，公共层不再直接依赖 `collaborationWorkspaceStore.currentCollaborationWorkspaceId` 解释当前上下文；`menu-space` 继续只承担导航宿主语义。
- 通过：消息域协作空间视图仍可通过兼容桥接工作，但没有反向污染公共壳层主上下文。

### 残余兼容点
- `frontend/src/store/modules/collaboration_workspace.ts`：保留 team 兼容派生层。
- `frontend/src/utils/http/index.ts`：保留 legacy team API 的 `X-Collaboration-Workspace-Id` 请求头桥接。
- `frontend/src/views/message/modules/useMessageWorkspace.ts`：保留消息域 `currentCollaborationWorkspaceId` 兼容派生。
- `frontend/src/router/guards/beforeEach.ts`、`frontend/src/views/auth/login/index.vue`、`frontend/src/types/api/api.d.ts`：仍读取/承载 `current_collaboration_workspace_id` 兼容字段，但它只作为旧 team 视图初始化提示，不再是授权源。

## 2026-04-07 空间权限主链继续收口

### 本次改动
- `backend/internal/pkg/permissionrefresh/service.go` 已补齐 `RefreshCollaborationWorkspace* / RefreshPersonalWorkspace*` canonical 方法，并把服务内部字段与调用点收口到个人空间 / 协作空间语义；旧 `RefreshTeam / RefreshPlatform*` 仅作为兼容壳保留。
- `backend/internal/modules/system/featurepackage/service.go`、`backend/internal/modules/system/featurepackage/handler.go`、`backend/internal/modules/system/featurepackage/module.go` 已把 active service/handler surface 切到 `GetPackageCollaborationWorkspaces / SetPackageCollaborationWorkspaces / GetCollaborationWorkspacePackages / SetCollaborationWorkspacePackages`，并把刷新调用统一改到 `RefreshCollaborationWorkspace(s)`。
- `backend/internal/pkg/authorization/authorization.go` 已新增 `ErrCollaborationWorkspaceMemberNotFound / ErrCollaborationWorkspaceMemberInactive` 和协作空间角色快照 helper 的 canonical 名称；旧 `ErrTenant*` 与 `getTeam*` 仅保留薄兼容别名。
- `backend/cmd/migrate/main.go` 已继续保持“本地空项目最终态”入口：历史 rename/backfill 逻辑不再执行，迁移入口直接落到最终 schema 路径。
- `frontend/src/types/api/api.d.ts`、`frontend/src/api/system-manage.ts`、`frontend/src/api/collaboration-workspace.ts` 已去掉 `TeamList / FeaturePackageTeam* / TeamActionOriginsResponse / TeamMenuOriginsResponse` 等 public type 和 canonical API 暴露，旧 team 命名不再作为前端主契约。
- `frontend/src/views/system/feature-package/index.vue`、`frontend/src/views/system/feature-package/modules/feature-package-teams-dialog.vue`、`frontend/src/views/system/team-roles-permissions/index.vue` 已切到 collaboration canonical API，并修掉由旧命名导致的类型与构建断点。

### 验证
- `go test ./... -run '^$'`
- `pnpm --dir frontend lint`
- `pnpm --dir frontend build`

### 下次方向
- 继续删除 backend 中仅剩的兼容壳方法和旧 `team/tenant` helper 名，重点仍是 `permissionrefresh / featurepackage / authorization`。
- 视本地空项目策略决定是否直接删除 `backend/cmd/migrate/main.go` 中未再执行的历史 rename/backfill helper，实现迁移文件层面的完全最终态。
- 继续把前端目录名、组件文件名和 locale key 里的 `team` 残留压到更低，只保留历史归档或兼容壳。

## 2026-04-07 协作空间最终体验补完

### 本次改动
- `frontend/src/api/collaboration-workspace.ts` 修正了协作空间列表归一化顺序：当 `/api/v1/collaboration-workspaces/mine` 同时返回 `id` 与 `workspace_id` 时，`collaborationWorkspaceId` 现在优先取真实的协作空间 ID，不再错误回落为 workspace ID。
- `frontend/src/utils/storage/storage-key-manager.ts` 增加了一次性本地存储清理逻辑：自动删除 `sys-vundefined-* / sys-vnull-* / sys-vNaN-*` 这类无效版本 key，并清理已废弃的 `tenant / workspace / collaboration-workspace-adapter` 旧 store key。
- 同一 store 在迁移到当前版本后，会自动清除旧版本 key，避免本地开发环境长期残留 `sys-v3.0.1-*` 等历史缓存干扰当前 `workspace / collaboration` 语义。

### 验证
- `go test ./... -run '^$'`
- `pnpm --dir frontend lint`
- `pnpm --dir frontend build`
- 浏览器自动化验证：
  - 清理旧 localStorage 后重新登录 `admin / admin123456`
  - 切换到协作空间上下文
  - 访问 `#/collaboration-workspace/message`
  - 校验协作空间 store 中 `currentCollaborationWorkspaceId` 与 `/mine` 列表项中的 `collaborationWorkspaceId` 一致

## 2026-04-06 Tenant 语义回收与 Phase 9/10 收尾

### 本次改动
- `backend/internal/modules/system/auth/middleware.go`、`backend/internal/pkg/authorization/context.go`、`backend/internal/pkg/authorization/authorization.go`、`backend/internal/api/middleware/app_context.go` 已把当前协作空间语义从 `tenant` 脱钩：`X-Collaboration-Workspace-Id` 与 `collaboration_workspace_id` 现在明确表示当前协作空间，`collaboration_workspace_id / collaboration_workspace_id / X-Collaboration-Workspace-Id` 只保留为旧 team 兼容输入。
- `backend/internal/modules/system/models/workspace.go`、`backend/internal/modules/system/workspace/service.go`、`backend/internal/modules/system/workspace/handler.go`、`backend/internal/modules/system/tenant/handler.go`、`backend/internal/modules/system/user/handler.go`、`backend/internal/modules/system/system/message_service.go` 与相关前端类型/API 已把 active DTO 和消息域输出切到 `current_collaboration_workspace_id / target_collaboration_workspace_id / owner_collaboration_workspace_id / collaboration_workspace_id`；`collaboration_workspace_id / current_collaboration_workspace_id / target_collaboration_workspace_id` 退化为兼容读取或历史映射。
- `frontend/src/utils/http/index.ts`、`frontend/src/router/guards/beforeEach.ts`、`frontend/src/views/auth/login/index.vue`、`frontend/src/views/message/modules/useMessageWorkspace.ts`、`frontend/src/views/message/modules/message-*.vue` 已把前端主语义收口到 `workspace / team workspace`；`collaborationWorkspaceStore` 只保留旧 team API 与消息域兼容桥接，不再承担页面级主上下文职责。
- `docs/workspace-permission-migration.md`、`docs/workspace-permission-stage-log.md`、`docs/workspace-glossary.md` 已同步改写：`tenant` 现在明确预留给未来真正的多租户系统，不再承担当前 team/workspace 主语义；Phase 9 与 Phase 10 标记为已完成。

### 验证
- `go test ./internal/modules/system/auth ./internal/modules/system/featurepackage ./internal/modules/system/permission ./internal/modules/system/tenant ./internal/modules/system/user ./internal/modules/system/page ./internal/modules/system/space ./internal/modules/system/system ./internal/pkg/authorization ./internal/pkg/platformaccess ./internal/pkg/permissionrefresh ./internal/pkg/teamboundary ./internal/api/router`
- `pnpm --dir frontend lint`
- `pnpm --dir frontend build`

### 结论
- 当前主线已完成 tenant 语义回收：`tenant` 仅保留在 legacy route、legacy persistence、历史映射与未来多租户预留位中。
- 当前权限、导航、消息、协作空间协作主线统一采用 `workspace / personal workspace / team workspace` 语义。
- 仍允许暂留的兼容物包括 `/api/v1/collaboration-workspaces/*`、旧 tenant 表与镜像关系、`X-Collaboration-Workspace-Id` 与 `current_collaboration_workspace_id` 输入兼容；但从本次收口开始，禁止再新增任何把 `tenant` 当当前 team/workspace 主输出的字段、header、上下文键或前端状态。

## 2026-04-06 Tenant 页面级命名继续收口

### 本次改动
- `frontend/src/views/message/modules/message-dispatch-console.vue` 已把页面级目标协作空间状态从 `target_collaboration_workspace_ids` 收口为 `targetCollaborationWorkspaceIds`，发送 payload 以 `target_collaboration_workspace_ids` 为主，并继续附带 `target_collaboration_workspace_ids` 兼容字段。
- `frontend/src/views/message/modules/message-recipient-group-console.vue` 已把页面级目标协作空间状态从 `collaboration_workspace_id` 收口为 `collaborationWorkspaceId`，页面与规则汇总不再把 tenant 当成当前协作空间主语义；提交时继续同时带 `collaboration_workspace_id` 与 `collaboration_workspace_id` 兼容字段。

### 验证
- `pnpm --dir frontend exec eslint src/views/message/modules/message-dispatch-console.vue src/views/message/modules/message-recipient-group-console.vue`
- `pnpm --dir frontend build`

### 下次方向
- 继续只把 `tenant` 保留在后端 persistence、legacy route 和兼容 bridge 中，前端页面级状态不再继续扩散 tenant 命名。

## 2026-04-06 协作空间全量改名落地

### 本次改动
- `frontend/src/api/collaboration-workspace.ts` 现在是唯一 active 协作空间 API 入口，`frontend/src/api/team.ts` 已降级为纯兼容 re-export 壳，避免了 `team -> collaboration-workspace` 的循环依赖。
- `frontend/src/utils/http/index.ts`、`frontend/src/api/message.ts`、`frontend/src/api/workspace.ts` 已把请求头主线收敛到 `X-Auth-Workspace-Id` + `X-Collaboration-Workspace-Id`，并补齐 `current_collaboration_workspace_id / current_collaboration_workspace_name` 等 snake_case 兼容字段。
- `frontend/src/views/collaboration-workspace/**` 已成为协作空间相关页面的主目录，消息中心、收件箱、顶部 header / badge、消息模板和成员页都已切换到“个人工作空间 / 协作空间”命名，旧 `team` 术语仅保留在兼容字段和少数内部别名里。
- 已完成 `pnpm --dir frontend lint` 和 `pnpm --dir frontend build` 验证，当前前端可正常编译。

### 下次方向
- 如果后端后续同步完成页面键、菜单键和路由键的正式切换，可以继续收紧前端里残留的 `team` 兼容别名和旧目录壳。
- 后续若要进一步统一命名，可继续把系统管理页、功能包页里少量 `team` 兼容标签压到更窄的读写边界，只保留历史数据映射，不再参与 UI 展示。

## 2026-04-06 前端协作空间目录命名收口

### 本次改动
- `frontend/src/views/team/**` 内部变量、组件名和 `defineOptions` 已继续收口为 `collaboration` / `collaborationWorkspace` 语义，并新增 `frontend/src/views/collaboration-workspace/**` 作为协作空间页面的 canonical 目录。
- `frontend/src/api/collaboration-workspace.ts` 已成为协作空间 API 的 canonical 入口，`frontend/src/api/team.ts` 已退化为纯兼容壳；`frontend/src/store/modules/tenant.ts` 仍仅作为兼容别名，不再承担主语义。
- `frontend/src/components/core/layouts/art-header-bar/widget/ArtCollaborationWorkspaceSwitcher.vue`、`frontend/src/components/business/team/NoCollaborationWorkspaceState.vue` 已替换旧的 tenant/team 组件命名，相关 imports 与页面引用已同步修复。
- 本批修复了由于文件移动和换行差异导致的前端 lint/build 问题，最终 `pnpm --dir frontend lint` 与 `pnpm --dir frontend build` 均通过。

### 下次方向
- `team.ts`、`tenant.ts`、`fetchGetTenantOptions`、`useTenantStore` 等兼容名称可以继续保留到历史桥接完全稳定，但不建议再新增任何新的 active 依赖。
- 后续若继续清理，可以优先从 `frontend/src/views/team/**` 迁移到 `frontend/src/views/collaboration-workspace/**` 的路由与目录引用入手，只保留旧目录作为兼容壳。

## 2026-04-07 协作空间最终体验收尾与浏览器烟测

### 本次改动
- `workspace` 统一权限主体、`workspace_type = personal | collaboration`、`auth_workspace_id + auth_workspace_type` 作为运行时权限上下文这条主线，已在当前前后端代码、迁移入口、seed 和 public API 上继续收口。
- backend 重新执行了 `backend/cmd/migrate/main.go -fresh`，并再次执行 `backend/cmd/init-admin/main.go`，本地开发库已按最终态 schema 与 seed 重建。
- 通过真实浏览器完成了登录、创建协作空间、切换到协作空间、打开协作空间消息页和关键接口断言的自动化烟测，当前主线已能按 `workspace_type = personal | collaboration` 运行。

### 验证
- `go test ./... -run '^$'`
- `pnpm --dir frontend lint`
- `pnpm --dir frontend build`
- 浏览器自动化验证：
  - 登录页使用 `admin / admin123456`
  - 新建协作空间 `自动化验证协作空间`
  - 调用 `/api/v1/workspaces/switch` 成功切换到 `auth_workspace_type = collaboration`
  - 打开 `#/collaboration-workspace/message`
  - 断言 `GET /api/v1/messages/dispatch/options => 200`
  - 断言当前页面标题为 `协作空间消息发送 - G&G-E`

### 下次方向
- 若继续追求完全最终体，可继续清持久层结构名和少量历史符号名中的 `Tenant / Team` 词形，但这已经不再影响当前运行时主语义。
- 协作空间消息发送页当前会在无发信配置时给出明确告警，可后续再补默认发信配置 seed，让空项目初始化后直接具备完整发信能力。

## 2026-04-07 功能包目录与空间范围文案收口

### 本次改动
- `backend/internal/modules/system/featurepackage/service.go` 与 `backend/internal/modules/system/featurepackage/handler.go` 已将功能包列表与候选列表放宽为全局目录查询，不再强制要求 `app_key` 才能查看列表；App 仅在创建、编辑、关系树和绑定类动作中继续作为上下文约束。
- `frontend/src/views/system/feature-package/index.vue` 去掉了顶部 App 选择器，改为直接展示全量功能包目录；页面中的创建、关系树与绑定动作仍会在缺少 App 上下文时给出提示，并继续按当前 App 约束执行。
- `frontend/src/views/system/feature-package/modules/feature-package-dialog.vue`、`frontend/src/views/system/feature-package/modules/feature-package-bundles-dialog.vue`、`frontend/src/views/system/feature-package/modules/feature-package-menus-dialog.vue`、`frontend/src/views/system/feature-package/modules/feature-package-actions-dialog.vue`、`frontend/src/views/system/action-permission/index.vue`、`frontend/src/views/system/action-permission/modules/action-permission-search.vue` 统一把“上下文”相关文案收口为“空间范围 / 跨空间镜像”，减少 App 与空间语义的混淆。
- `frontend/src/api/system-manage.ts` 对功能包列表与候选集请求做了空 `appKey` 收敛，避免把未选择 App 的空值当成噪声参数发送给后端。

### 下次方向
- 功能包目录当前已经可以全局浏览，后续若继续优化，重点应放在“绑定动作在缺少 App 上下文时的引导体验”，而不是再把列表层恢复成 App 前置。

## 2026-04-07 空间权限最终体清理与协作空间同步修复

### 本次改动
- `frontend/src/store/modules/collaboration-workspace.ts` 增强了与 `workspaceStore` 的同步：当前授权空间切换后，`currentCollaborationWorkspaceId` 现在会优先从当前授权 workspace 上的协作空间标识派生，避免切换到协作空间后 store 残留旧值。
- `frontend/src/components/business/layout/AppContextBadge.vue`、`frontend/src/components/core/layouts/art-header-bar/widget/ArtCollaborationWorkspaceSwitcher.vue`、`frontend/src/components/core/layouts/art-header-bar/widget/ArtUserMenu.vue`、`frontend/src/components/core/layouts/art-notification/index.vue`、`frontend/src/views/dashboard/console/index.vue`、`frontend/src/views/workspace/inbox/index.vue`、`frontend/src/views/message/modules/useMessageWorkspace.ts`、`frontend/src/views/message/modules/message-record-console.vue`、`frontend/src/views/message/modules/message-sender-console.vue`、`backend/internal/modules/system/permission/service.go`、`backend/internal/pkg/permissionseed/seeds.go` 已把“个人工作空间 / 平台发送 / 平台用户 / 平台侧”这类旧文案继续收口为“个人空间 / 个人空间发送 / 个人空间用户 / 空间权限”。
- 已删除空置旧目录 `backend/internal/modules/system/tenant` 与 `backend/internal/pkg/teamboundary`，active 代码层不再保留这两个历史模块入口。
- `frontend/src/utils/storage/storage-config.ts` 维持了 `sys-v${CURRENT_VERSION}` 的稳定版本前缀，当前 active 代码不会再生成 `sys-vundefined-*` 的 localStorage 键。

### 验证
- `go test ./... -run '^$'`
- `pnpm --dir frontend lint`
- `pnpm --dir frontend build`

### 下次方向
- 若继续追求符号级最终体，可继续清持久层结构名和极少量历史注释中的 `Tenant / Team` 词形。
- 当前主链已统一到 `workspace / personal / collaboration`，后续优化重点应转到默认 seed 丰富度与协作空间业务体验，而不是继续调整权限主语义。

## 2026-04-07 全局用户与角色 App 生效范围收口

### 本次改动
- 用户目录已明确收口为全局用户目录：`frontend/src/views/system/user/index.vue` 不再要求先选择 App 才加载用户列表，App 只在功能包、菜单裁剪和权限测试时参与裁剪。
- 角色目录已明确收口为全局角色目录：`frontend/src/views/system/role/index.vue` 现在默认显示全部角色，并支持通过 `App` 过滤查看“全局通用角色 + 指定 App 生效角色”。
- `backend/internal/modules/system/role/service.go`、`backend/internal/modules/system/role/handler.go`、`backend/internal/modules/system/user/repository.go`、`backend/internal/modules/system/models/model.go` 新增并启用了 `role_app_scopes` 主链：
  - 角色可配置多个生效 App；
  - 未配置生效 App 时视为全局通用；
  - 在角色功能包、菜单和权限配置时会校验当前 App 是否落在角色生效范围内。
- `frontend/src/views/system/role/modules/role-edit-dialog.vue` 已增加“生效 App”多选配置；`frontend/src/views/system/user/modules/user-dialog.vue` 会在角色选项中标明“全局通用”或具体 App 范围，避免误选。
- `frontend/src/api/system-manage.ts` 与 `frontend/src/types/api/api.d.ts` 已把 `appKeys / isGlobal` 收为角色 public type 的一部分，不再让前端只能按旧的“先选 App 再看角色目录”方式工作。

### 验证
- `go test ./... -run '^$'`
- `pnpm --dir frontend lint`
- `pnpm --dir frontend build`

### 下次方向
- 若后续继续优化，可补针对 `role_app_scopes` 的 focused tests，覆盖“全局通用角色”和“指定 App 生效角色”两类场景。
- 目前用户与角色目录已经和“全局目录 + App 裁剪”的模型对齐，后续重点应放在默认 seed 和角色模板体验，而不是继续让列表层绑死 App。 

## 2026-04-07 功能权限与 API 注册空间范围语义收口

### 本次改动
- `frontend/src/api/system-manage.ts` 调整了 `deriveContextType` 的默认映射：`system.*`、`feature_package.*`、`message.*` 以及角色、用户、菜单、页面、API 注册等治理能力现在默认归到 `common`，不再误判为 `personal`。
- `frontend/src/views/system/action-permission/modules/action-permission-dialog.vue` 重写了“空间范围”提示文案：
  - 个人空间只用于明确属于个人空间的权限键；
  - 协作空间只用于成员、边界、协作消息等专属能力；
  - 平台管理 API、系统治理能力和跨空间能力统一归 `通用`。
- `frontend/src/views/system/api-endpoint/index.vue` 与 `frontend/src/views/system/api-endpoint/modules/api-endpoint-search.vue` 已把 `contextScope` 的展示文案从“协作空间上下文”统一改成“协作空间要求”，并把“跨上下文共享”改成“跨空间共享”，避免把 API 运行时要求和权限键空间范围混淆。

### 验证
- 本轮按当前协作约束未额外执行自动化验证；改动集中在前端映射与文案收口。

### 下次方向
- 若继续收口，可把 API 注册页中的 `App 上下文` 动作提示进一步改成“App 资源域”，继续降低 `app` 与 `workspace` 语义混淆。
- 后续若补验证，优先检查功能权限列表里 `system.*`、`collaboration_workspace.*` 和 `personal.*` 三类权限键是否分别显示为 `通用 / 协作空间 / 个人空间`。

## 2026-04-07 功能包模型重构

### 本次改动
- 功能包管理收口为“绑定在空间上的能力集合”，列表页不再要求先选 App，查询接口也不再按当前 App 过滤。
- 功能包主契约改为 `workspaceScope + appKeys`，`app_keys` 支持多 App 绑定，未绑定 App 时默认适用于所有 App。
- 功能包页、基础包弹窗、功能范围弹窗、协作空间开通弹窗都统一改成“适用空间 / 适用 App”的展示语义，去掉了 `contextType` 作为功能包主字段的公开暴露。

### 下次方向
- 继续核对后端 `featurepackage` 相关快照、摘要和兼容字段，确保 `contextType` 不再参与主流程判断。
- 如需继续收口体验层，再把功能包页中“适用空间 / 适用 App”的提示文案统一到最终版，减少兼容提示。

## 2026-04-07 功能包适用空间与多 App 绑定收口

### 本次改动
- 功能包模型继续回写为“绑定在空间上的能力集合”，功能包列表查询保持全局目录，不再强制 App 前置；功能包生效改为同时看 `workspaceScope` 与 `appKeys`，其中 `appKeys` 允许多 App 绑定，未绑定 App 时默认适用于所有 App。
- `backend/internal/modules/system/featurepackage/service.go`、`backend/internal/modules/system/featurepackage/handler.go`、`backend/internal/modules/system/models/model.go`、`backend/internal/api/dto/permission.go` 补齐了 `containsString`、`workspaceScope`、`appKeys` 相关收口，功能包公开返回不再继续把 `contextType` 当主字段。
- `frontend/src/api/system-manage.ts`、`frontend/src/api/collaboration-workspace.ts`、`frontend/src/views/system/feature-package/**`、`frontend/src/views/system/role/modules/role-package-dialog.vue`、`frontend/src/views/system/user/modules/user-package-dialog.vue`、`frontend/src/views/system/collaboration-workspace-roles-permissions/modules/collaboration-workspace-role-package-dialog.vue`、`frontend/src/components/business/permission/PermissionSourcePanels.vue` 等页面和 API 已把功能包相关文案与交互统一为“适用空间 / 适用 App”，并把列表与绑定动作拆开。
- 默认种子里的平台治理包名称已经收口为 `platform_admin.*`，协作空间成员包继续保留 `collaboration_workspace.member_admin`，避免再把平台治理能力误标为“个人空间功能包”。

### 下次方向
- 若继续收口，可把功能包相关的兼容字段再进一步削薄，只保留 `workspaceScope + appKeys` 作为前端主契约。
- 后续若要继续优化体验，重点应放在功能包绑定、菜单裁剪和 App 资源域提示上，而不是再把列表层恢复成 App 前置。

## 2026-04-07 功能包页面兼容文案继续收口

### 本次改动
- 功能包页面继续去掉与 App 前置相关的残余参数透传，`feature-package` 相关列表/候选查询不再携带 `appKey`，保持全局目录查询语义。
- `frontend/src/components/business/permission/PermissionSourcePanels.vue` 以及功能包相关角色/用户/协作空间弹窗里，把跳转与筛选从旧 `contextType` 收口到 `workspaceScope`，避免页面层继续把功能包当成“平台/个人空间上下文”。
- `frontend/src/views/system/feature-package/index.vue`、`frontend/src/views/system/feature-package/modules/*` 的说明与筛选文案继续收口，统一为“适用空间 / 适用 App”，并去掉“全部上下文”这类旧词形。

### 下次方向
- 若继续优化，下一步就是把功能包相关页面里还保留的少量兼容变量名再统一成 `workspaceScope` / `appKeys`，让前端主契约更干净。
- 当前功能包主流程已经可用，后续重点应转向默认 seed 的体验细化，而不是再改查询主链。

## 2026-04-07 功能包绑定菜单支持 App 选择与空菜单空间

### 本次改动
- 功能包绑定菜单抽屉新增 App 选择，并支持菜单空间留空；留空时默认查看当前 App 的全部菜单，不再强制回落到默认菜单空间。
- 后端 `menu` 树查询在 `all=true` 场景下保留 `app_key` 必填，但不再把空菜单空间自动替换成默认菜单空间，避免绑定菜单时只能看到单一菜单空间。
- 功能包菜单绑定保存时会随当前 App 一并提交，确保菜单裁剪与当前 App 资源域保持一致。

### 下次方向
- 若后续继续收口，可把功能包绑定菜单抽屉里的 App/菜单空间提示文案再统一到最终版，减少兼容说明。
- 如还出现菜单为空或缺少 App 的报错，优先检查当前选中的 App 是否确实存在菜单空间绑定。
