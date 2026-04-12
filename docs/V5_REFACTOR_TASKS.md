# V5 重构任务起点

> 本文件是 GGE 5.0 重构的唯一任务入口。所有新工作从这里出发，历史文档已清空。
> 权威设计基线见根目录之外的《GGE_5.0_初始化架构文档.docx》（最新版含第 10 章 多租户预留）。

## 0. 当前状态

- 后端 `backend/`：Go 1.22 + Gin + GORM + Postgres + Redis + ES，已朝 5.0 术语收敛过一轮（workspace / featurepackage / role / permissionkey / appscope 等模块到位），但仍有 4.5 半迁移残留：`collaborationworkspace` 独立持久化、菜单与权限存在反推、API 注册手写。
- 前端 `frontend/`：Vue3 + TS + Vite + Element Plus + Pinia，**已接真实接口**，不再走 mock。
- 数据库：允许清空重建（baseline 一次性迁移）。

## 1. 核心目标

1. 把 5.0 文档第 1–9 章的模型在仓库内硬切落地（workspace 单一权限主体、菜单/页面/权限键三段分离、空间能力 ∩ 成员角色公式）。
2. 全链路预留 tenant 维度（文档第 10 章），workspace 保持团队语义。
3. API 改为 **OpenAPI-first**：spec 即真相，反推 handler、permission_keys、api_endpoints、前端 client、Swagger 文档。
4. 引入成熟 Go 模块，停止手写中间件/路由/校验/错误层。

## 2. 技术栈定锤

| 层 | 选型 |
|---|---|
| Web | Gin（保留） |
| API 契约 | **ogen**（OpenAPI 3 → Go server/client 生成） |
| 文档 UI | swaggo/http-swagger 挂 ogen spec |
| 权限引擎 | **casbin/v2** + 自研 evaluator 包装 |
| ORM | GORM（保留） |
| 迁移 | **pressly/goose**（替代 cmd/migrate） |
| 配置 | viper（保留） |
| 日志 | zap（保留） |
| 错误 | **cockroachdb/errors** |
| DI | **google/wire** |
| 缓存 | **eko/gocache/v3**（local LRU + Redis 二级） |
| 测试 | testify + **testcontainers-go** |
| 可观测 | otel SDK 显式接入 |

不做：换框架、换 ORM、引入 MQ / GraphQL / 多租户业务能力 / 插件化。

## 3. OpenAPI 工作流（核心）

### 3.1 目录
```
backend/
  api/openapi/
    openapi.yaml
    paths/{域}.yaml
    components/{schemas,responses,parameters}/
  api/gen/                # ogen 生成产物（git 跟踪）
  internal/api/handlers/  # 实现 ogen 接口
  internal/api/mapper/    # ogen DTO ↔ domain
```

### 3.2 OpenAPI 扩展字段（唯一真相源）
```yaml
post:
  operationId: createWorkspaceMember
  x-permission-key: workspace.member.create
  x-app-scope: required
  x-access-mode: permission
  x-tenant-scoped: true
```

### 3.3 自动派生
`cmd/gen-permissions` 解析 openapi.yaml → 生成 `permission_keys`、`api_endpoints`、`permission_key_api_bindings` seed。启动时校验：所有 `x-permission-key` 必须在 DB 中存在，否则拒启。

### 3.4 前端联动
ogen 同步出 TS client（或 openapi-typescript），`frontend/src/api/` 直接消费生成产物，不再手写 axios 类型。

## 4. 权限引擎

Casbin 模型：`(member, workspace, permission_key)` + 成员-角色绑定 `g(member, role, workspace)`。
上层 `internal/pkg/permission/evaluator`：
```go
type Evaluator interface {
    Resolve(ctx, accountID, workspaceID) (*ResolvedPermissions, error)
    Can(ctx, accountID, workspaceID, key) (bool, error)
    Explain(ctx, accountID, workspaceID) (*Explanation, error)
}
```
内部：`workspace_feature_packages 权限并集 ∩ casbin 决策`，二级缓存。Gin 中间件按 ogen operationId 反查 `x-permission-key` 调用。

## 5. 数据库 baseline

按 5.0 文档第 6 节 + 第 10 章落 14 张表（含 `tenants`），goose 第 1 号迁移一次性建表，清库重建。

**全部业务表必须带 `tenant_id`**：accounts、workspaces、workspace_members、roles、feature_packages、workspace_feature_packages、permission_keys、apps、menu_spaces、menu_definitions、ui_pages、api_endpoints、permission_key_api_bindings。
唯一性约束改为 `(tenant_id, business_key)`。当前阶段所有数据归属内置 `default` tenant。

后续默认数据走 `pkg/permissionseed` 等幂等 seed，不再写进迁移链。

## 6. 模块边界（用 wire 组装）

```
internal/
  modules/
    tenant/         (新增，仅 default 兜底)
    account/        (原 user 重命名，全局账号)
    workspace/      (合并 collaborationworkspace)
    member/
    role/
    featurepackage/
    permission/     (key 定义、seed)
    app/
    menuspace/
    menu/
    page/
    apiendpoint/    (由 OpenAPI 派生)
  pkg/
    permission/{evaluator,cache}
    openapi/        (ogen 中间件适配)
    tenantctx/      (RequestContext 注入)
    db/ cache/ errors/ logger/
```

## 7. 阶段拆分（按 PR 切片）

| Phase | 内容 | 输出 |
|---|---|---|
| **0** | 引入新依赖、建 `refactor/v5-baseline` 长期分支、本文件落档 | 1 PR |
| **1** | OpenAPI 骨架 + ogen 接入 + `cmd/gen-permissions` + 启动校验 + workspace 域示例 | 1–2 PR |
| **2** | goose 接入、14 张表 baseline 迁移、清库 + seed pipeline、删 collaborationworkspace 旧表 | 1 PR |
| **3** | `pkg/permission/evaluator` + casbin 适配 + `/permissions/explain` + 二级缓存 | 1–2 PR |
| **4** | 模块逐域迁移到 OpenAPI（顺序：account → workspace → member → role → featurepackage → permission → app → menuspace → menu → page → apiendpoint），每域含前端 client 替换 | 多 PR |
| **5** | 前端真实接口全量切到生成 client、路由守卫接 `/permissions/explain` | 1–2 PR |
| **6** | `/swagger` UI、otel 接入、删除所有 `inherit_permission` 与菜单反推权限残留 | 1 PR |

## 8. 多租户预留纪律

- RequestContext 必带 `tenant_id`，中间件统一注入，当前恒为 default。
- 所有 repository 查询强制按 `tenant_id` 过滤；通过 GORM scope 或基类约束，不允许裸 SELECT。
- 缓存 key、日志、trace、审计事件一律带 `tenant_id` 前缀/字段。
- OpenAPI 不引入 tenant 路径参数；未来通过子域名或 `X-Tenant-Code` Header 解析。
- 前端不感知 tenant。
- 每个新模块 / 新表 / 新接口评审必须显式回答：**是否带 tenant_id、是否在仓储层强制过滤**。

## 9. 立即执行的第一步

Phase 0 + Phase 1 的 workspace 示例域，跑通完整链路：
`openapi.yaml → ogen 生成 → handler 实现 → permission_keys 自动 seed → /swagger 可访问 → 前端用生成 client 调一个接口`。

跑通后再批量铺开 Phase 2–6。

## 10. 阶段进度

### Phase 0 — 已完成
- 后端依赖引入：`ogen-go/ogen`、`casbin/v2`、`google/wire`、`pressly/goose/v3`、`cockroachdb/errors`、`eko/gocache/v4`、`testcontainers-go`（`go mod tidy` + `go build ./...` 通过）。
- 骨架目录：`backend/api/openapi/{paths,components}`、`backend/api/gen/`、`backend/internal/api/{handlers,mapper}/`、`backend/internal/pkg/permission/evaluator/`、`backend/internal/pkg/tenantctx/`。
- `backend/api/openapi/openapi.yaml` 主入口落档（空 paths，等 Phase 1 填充）。
- `backend/Makefile` 新增 `make gen`（ogen 生成）与 `make gen-permissions`（占位）。
- `backend/cmd/gen-permissions/main.go` 占位，Phase 1 同步实现解析逻辑。
- 决策定锤：API 不并存 `/v2`，直接替换；前端走 `openapi-typescript + openapi-fetch`。

### Phase 1 — 已完成（workspace 示例域）
- `backend/api/openapi/openapi.yaml` 落 `GET /workspaces/{id}`，含 `x-permission-key: workspace.read`、`x-tenant-scoped: true`、`x-app-scope: optional`、`x-access-mode: permission`。
- `make gen` 通过，`backend/api/gen/` 出 ogen server / schemas / router 全套（已纳入 git）。补齐 ogen runtime 依赖：`go-faster/errors`、`go-faster/jx`、`ogen-go/ogen`。
- `backend/cmd/gen-permissions` 解析器实现：扫描所有 operation，校验 `x-permission-key`，输出 `internal/pkg/permissionseed/openapi_seed.json`。当前仅生成 1 条 `getWorkspace`。
- `backend/internal/api/handlers/workspace.go` 实现 `gen.Handler`（嵌入 `UnimplementedHandler`），调用既有 `workspace.Service`，把 domain model 映射到生成的 `WorkspaceSummary`。用 `ctx` 传递 `user_id`。
- `backend/internal/api/router/router.go`：在 authenticated 组里直接挂 ogen `*Server`，通过 gin bridge 注入 `user_id` 到 `r.Context()` 并 strip `/api/v1` 前缀。**直接替换**了旧的 `GET /workspaces/:id`（在 `workspace/module.go` 中删除）。其余 workspace 路由（`/my`、`/current`、`/switch`）暂保留旧 handler，等后续 PR 一起迁。
- `go build ./...` 通过。
- 后续 Phase 1 收尾项（留到 Phase 2 / 3 一起做）：启动时把 `openapi_seed.json` 与 DB 中 `permission_keys` 对账校验、`/swagger` UI 挂载、ogen 中间件接入 evaluator。

### Phase 2a — 已完成（goose + 租户基线）
- 引入 `pressly/goose/v3`，迁移文件落 `backend/internal/pkg/database/migrations/`，由 `database.RunGooseMigrations` 通过 `embed.FS` 加载。
- 第 1 号迁移 `00001_tenants_baseline.sql`：启用 `uuid-ossp` / `pgcrypto`，建 `tenants` 表（含 `is_default` 部分唯一索引），并 seed 内置 `default` 租户。
- 新增 `models.Tenant` 与 `models.TenantScoped` 嵌入结构 + `DefaultTenantCode` 常量。`Workspace` 与 `WorkspaceMember` 已嵌入 `TenantScoped`，AutoMigrate 同步建出 `tenant_id` 列。
- `cmd/migrate` 增加：goose 先于 AutoMigrate 跑；`ensureDefaultTenantBackfill` 把存量 `workspaces` / `workspace_members` 行的 `tenant_id` 回填为默认租户。
- 范围控制：本次只把 `tenant_id` 推到 v5 权限主轴的 2 张表（`workspaces`、`workspace_members`），其余表仍按旧结构跑。剩余 11 张目标表的 `tenant_id` 与 14 张 v5 baseline 重建留给 Phase 2b/Phase 4 各域迁移时同步落地，避免一次 PR 触动 60+ 文件。
- `collaborationworkspace` 模块暂未删除，沿用兼容字段；按计划在 Phase 4 各业务域迁到 OpenAPI 时随域清理。
- `go build ./...` 通过；live DB 验证 (`make db-reset && make migrate`) 留给本地执行。

### Phase 1 收尾 + Phase 3 + Phase 4 首刀 — 已完成（合并提交）

**Phase 1 收尾**
- `backend/api/openapi/embed.go`：把 `openapi.yaml` 通过 `//go:embed` 暴露给运行时（`SpecBytes`）。
- `backend/internal/pkg/openapidocs`：`Mount(*gin.Engine)` 在根路由上挂 `/openapi.yaml` 与 `/swagger`（CDN 版 Swagger UI HTML 一页流，零额外资产）。
- `backend/internal/pkg/permissionseed/openapi_loader.go`：embed `openapi_seed.json`，`LoadOpenAPISeed()` 在启动时校验所有 operation 必须带 `permission_key`，缺一即 fatal。`router.go` 启动时调用并日志输出 operation 数。

**Phase 3：权限决策入口**
- `backend/internal/pkg/permission/evaluator`：`Evaluator` 接口（`Resolve` / `Can` / `Explain`），`ResolvedPermissions` 与 `Explanation` 类型。
- `gormEvaluator` 用 raw SQL 走 `workspace_feature_packages → feature_package_keys → permission_keys` 求 workspace 的功能包权限并集；按 package_id 出处为 `Explain` 的 `feature_package_sources` 赋值。
- 角色侧的交集（`workspace_role_bindings → role permissions`）留 `TODO(phase-3-followup)`，因为角色目前用 action_id 模型，需要先与 permission_key 对齐；接口签名不会变。
- casbin 暂未引入（现有 schema 直接 SQL 即可，引入 casbin 反而要先建 policy adapter）；接口预留，未来在 `gormEvaluator` 内部替换为 enforcer 不影响调用方。

**Phase 4 首刀：workspace 域 + 新增 permission 域**
- `api/openapi/openapi.yaml` 新增：`GET /workspaces/my`、`GET /workspaces/current`、`GET /permissions/explain`。每个 operation 都带 `x-permission-key` / `x-tenant-scoped` / `x-app-scope` / `x-access-mode`。
- `make gen` 重跑，`api/gen/` 全量再生。`gen-permissions` 输出 4 条 operation 到 `openapi_seed.json`。
- `internal/api/handlers/workspace.go` 重命名 `WorkspaceHandler` → `APIHandler`，新增 3 个 method：`ListMyWorkspaces`、`GetCurrentWorkspace`、`ExplainPermissions`。`APIHandler` 持有 `evaluator.Evaluator`，`/permissions/explain` 直接走它。
- `router.go`：所有 OpenAPI 路径（`/workspaces/my`、`/workspaces/current`、`/workspaces/:id`、`/permissions/explain`）由同一个 ogen `*Server` 经 gin bridge 接管。`workspace/module.go` 中的 legacy `GET /my`、`GET /current` 路由已删除；`POST /switch` 仍保留，等下个 PR 一起迁。
- `go build ./...` 通过。





## 阶段进度

### 2026-04-12 项目清理优化收口（Phase Cleanup）

**本次改动**
- 清理前后端低风险历史残留：`frontend/src` 的 `as any` 已清零，`console.log` 已清零，删除消息模块 `_unused` 占位函数与旧欢迎信息残留。
- 删除废弃 `internal/api/errcode` 包，新增轻量 `legacyresp` 兼容 helper，完成最后 4 处 legacy gin 错误响应迁移。
- 收口源码 TODO/FIXME 标记、将 `backend/cmd/diagnose` 改为结构化日志，并规范 `backend/Makefile` 兼容别名说明。
- 补 `backend/.env.example`，为 `backend/docker-compose.yml` 的 Elasticsearch/MinIO 增加 `optional` profile 与默认凭证安全警告；删除根目录 `start-frontend-shadcn.debug.log`。

**下次方向**
- 如确认 `.claude/worktrees/` 下 9 个历史 worktree 无需保留，再执行物理删除并收口清理任务最后一个阻塞节点。
- 若继续做类型治理，可把剩余 `any` 注解与隐式组件内部类型访问继续替换为公开 schema / 组件类型。

### 2026-04-12 P3-B 首步：领域化目录方案定稿（Phase Cleanup）

**本次改动**
- 新增 `docs/frontend-cleanup-p3b-notes.md`，把前端目录归并的目标结构正式定稿为 `domains/auth`、`domains/app-runtime`、`domains/navigation`、`domains/governance` 四个领域。
- 文档明确了当前 `api/composables/router/store/utils` 的真实聚合点，给出 `before/after` 对照、文件归属建议与不动的公共层（`components/views/locales/api/v5/utils/http`）。
- 同步定下迁移顺序：`auth -> app-runtime -> navigation -> governance -> compat cleanup`，供 `P3B-2 ~ P3B-6` 直接执行。

**下次方向**
- 进入 `P3B-2`，先落 `domains/auth` 骨架，把 `auth-flow`、`session runtime`、`api/auth` 收到单一领域目录。
- 迁移过程中保留 `store/modules/*` 与 `router/runtime/*` 的兼容 re-export，等 `P3-B` 尾声统一清理旧 import。

### 2026-04-12 P3-B 第二步：auth 领域首批归并（Phase Cleanup）

**本次改动**
- 新增 `frontend/src/domains/auth/`，首批落地 `api.ts`、`store.ts`、`centralized-login.ts`、`runtime/session.ts` 和 `flows/*`，把认证主链的 API、session runtime、登录注册回调流程、用户会话 store 收到同一领域目录。
- 旧路径 `api/auth.ts`、`store/modules/user.ts`、`utils/auth/centralized-login.ts`、`router/runtime/session.ts`、`composables/auth-flow/*` 已退化为兼容 re-export，不再承载真实实现。
- 全仓消费方 import 已切到 `@/domains/auth/*`，并在 `frontend/tsconfig.json` 补了 `@/domains/auth` 路径映射；`pnpm exec vue-tsc --noEmit` 已通过。

**下次方向**
- 进入 `P3B-3`，继续落 `domains/app-runtime`，优先搬 `app-context`、`menu-space`、`managed-app-scope` 与对应 runtime 入口。
- `domains/auth/store.ts` 目前仍直接依赖 `store/modules/{app-context,menu-space,...}` 的旧路径；等 `app-runtime` 迁完后，再把这些依赖切到新域路径，逐步压缩 `store/modules/*` 的兼容层。 

### 2026-04-12 P3-B 第三步：app-runtime 领域首批归并（Phase Cleanup）

**本次改动**
- 新增 `frontend/src/domains/app-runtime/`，首批落地 `context.ts`、`menu-space.ts`、`managed-app-scope.ts`、`useManagedAppScope.ts`、`app-scope.ts`、`runtime/app-context.ts` 与 `index.ts`，把 app context、menu-space、scope 读写、runtime app 切换入口收口到同一领域。
- 旧路径 `store/modules/app-context.ts`、`store/modules/menu-space.ts`、`hooks/business/{managed-app-scope,useManagedAppScope}.ts`、`utils/app-scope.ts`、`router/runtime/app-context.ts` 已退化为兼容 re-export；全仓消费方 import 已切到 `@/domains/app-runtime/*`。
- `domains/auth/*` 里对 `app-context/menu-space` 的依赖也已同步切到新域路径；`pnpm exec vue-tsc --noEmit` 再次通过。

**下次方向**
- 进入 `P3B-4`，继续迁 `router/core/*`、`router/runtime/navigation.ts`、`store/modules/menu.ts`、`store/modules/worktab.ts`、`utils/navigation/*` 到 `domains/navigation`。
- `domains/app-runtime/runtime/app-context.ts` 目前仍直接依赖 `router/core`、`router/runtime/navigation`、`store/modules/menu`、`store/modules/worktab` 的旧路径；等 `navigation` 领域迁完后，需要回头把这些依赖切到新域实现。 

### 2026-04-12 P3-B 收口：navigation/governance 归并 + 全量验证（Phase Cleanup）

**本次改动**
- 新增 `frontend/src/domains/navigation/` 承接 `router/core/*`、`router/runtime/navigation.ts`、`store/modules/{menu,worktab}.ts`、`utils/navigation/*`；旧路径已退化为兼容 re-export，并修正了 `ComponentLoader` 在新目录下的 `views` 相对路径，恢复首页与 Demo 页真实加载。
- 新增 `frontend/src/domains/governance/` 承接 `api/system-manage/*` 与 `utils/permission/*`；`frontend/src` 内真实消费方 import 已切到 `@/domains/governance/api*`、`@/domains/governance/utils/*` 与 `@/domains/navigation/*`，旧路径仅保留兼容壳。
- `frontend/playwright.config.ts` 已补 `loadEnv('test')` 到 `process.env` 的透传，真实回归和临时 webServer 统一走 `E2E_BASE_URL`；`pnpm exec vue-tsc --noEmit`、`pnpm build`、`pnpm exec playwright test` 全部通过，当前 `10 passed`。
- 针对迁移后的旧路径做了全仓搜索，`@/api/system-manage*`、`@/utils/permission*`、`@/store/modules/{menu,worktab}`、`@/router/runtime/navigation`、`@/utils/navigation*`、`@/router/core` 的真实消费方命中已清零。

**下次方向**
- 若继续清历史结构债，可把 `madge` 当前报出的 44 条循环链单独开新节点，优先处理 `auth/http/storage/router` 与 `app-runtime -> governance/_shared -> http` 两条总链。
- repo 级 eslint 仍有历史 `any`、`vue/no-unused-vars` 与生成文件格式噪音；如继续做治理，建议单开 lint 基线收口，不与 P3 目录迁移混在同一轮。

### 2026-04-08 Phase 4 全量推进：role/navigation/featurepackage/permission/menu/page/cw 迁移到 ogen（Phase 4）

**本次改动**
- 重新生成 `backend/api/gen/*` 与 `permissionseed/openapi_seed.json`（181 ops），`APIHandler` 已注入全部领域 service。
- 新增 ogen handler：`role.go`(14) / `navigation.go`(1) / `featurepackage.go`(17) / `permission.go`(15) / `menu.go`(10) / `page.go`(12) / `collaborationworkspace.go`(12)，`go build ./...` 通过。
- `router.go` 已桥接 role + navigation 路径；对应 legacy `module.go` 采用 `_ = rg; return` 跳过 gin 注册。

**下次方向**
- 在 `router.go` 批量补齐 featurepackage / permission / menu / page / collaborationworkspace 的 `ogenBridge` 入口，并在对应 legacy `module.go` 顶部早 return，让新 handler 真正生效后再次 `go build ./...`。
- 收尾剩余 501 stub：user 子路由、system、message、CW boundary 复杂操作、menu backup restore、fp rollback、permission 批量更新等。

### 2026-04-08 Phase 4 续推：28 ops 上线 + 路由桥接（Phase 4）

**本次改动**
- 新增 `handlers/extras.go`、`system.go`、`cwcurrent.go`，合计 28 个 ogen 操作（feature-package/permission/menu/page 尾部 + system apps/menu-spaces + cw current/my 基础）。
- `router.go` 桥接全部 28 条新路径；对应 legacy module.go 删除冲突注册：featurepackage/permission/menu/page 的 RegisterRoutes 已整体早 return，system/collaborationworkspace 仅删除已迁移行，未动的路由保持 gin 服务。
- `go build ./...` 通过；剩余 501 stub 56 个（user 子路由 8、message 16、system fast-enter/view-pages 3、CW boundary 复杂操作 ~29）。

**下次方向**
- message 域：导出 `internal/modules/system/system/message_service.go` 的请求/响应类型并在 `NewAPIHandler` 注入 messageSvc，然后落 16 个 ops。
- user 子路由：把 UserHandler 内的 snapshot/diagnosis 逻辑下沉到 `user.UserService` 后再迁移 8 个 ops。
- CW boundary 写操作：给 boundarySvc 增加 role/menu/action 栅格的写入入口或新建 cwroles 服务，再迁移剩余 ~24 个 ops。

### 2026-04-08 Phase 4 CW boundary ogen 迁移 ~29 ops（Phase 4）

**本次改动**
- 新增 `handlers/cwboundary.go`（888 行，29 个 ogen 操作）：ListCurrentCWRoles、CreateCurrentCWRole/BoundaryRole、UpdateCurrentCWBoundaryRole、DeleteCurrentCWBoundaryRole、Get/SetCurrentCWBoundaryRole{Packages,Menus,Actions}、GetCurrentCWBoundaryPackages、GetCurrentCW{Menus,MenuOrigins,Actions,ActionOrigins}、Get/SetCurrentCWMemberRoles、GetCW{Menus,MenuOrigins,Actions,ActionOrigins}、SetCW{Menus,Actions}、ListCWRoles。
- `workspace.go` 为 APIHandler 新增 `db`/`roleRepo`/`userRoleRepo`/`featurePkgRepo`/`cwFeaturePkgRepo`/`keyRepo` 字段并在 `NewAPIHandler` 中注入。
- `router.go` 追加 29 条 `ogenBridge` 路径（`/current/roles`、`/current/boundary/…`、`/current/menus`、`/current/action-origins` 等及 `/:id/roles`、`/:id/menus`、`/:id/actions` 等）；`module.go` 同步删除对应 27 条 legacy gin 注册，`go build ./...` 通过。

**下次方向**
- message 域 16 ops：导出 messageSvc 后迁移。
- user 子路由 8 ops（snapshot/diagnosis）：下沉到 UserService 后迁移。
- system fast-enter/view-pages 3 ops 及剩余 stub 清零后可删除所有 legacy module RegisterRoutes 入口。

### 2026-04-08 Phase 4 全量完成：181 ops 全部实现，0 stub 残留（Phase 4）

**本次改动**
- 并行三子代理完成 user 子路由（6+2 ops）、message 域（19 ops facade）、CW boundary（29 ops）全量实现；结合之前批次，ogen 实现数达到 181/181，`UnimplementedHandler` 残留归零。
- `internal/modules/system/system/message_facade.go` 导出 messageService 全部方法；`internal/api/handlers/cwboundary.go` 实现 CW boundary/current 复杂操作；`internal/api/handlers/user_subroutes.go` 补齐 user 子路由。
- router.go 桥接全部新路径；各 legacy module.go RegisterRoutes 删除冲突注册（featurepackage/permission/menu/page 已整体早 return，cw/system/user 仅保留未迁移残余）。`go build ./...` 通过。

**下次方向**
- 集成测试：对关键路径（role CRUD、navigation、permission/menu/page、CW boundary、message dispatch）进行冒烟测试，确认 ogen bridge 与 legacy 行为一致。
- 清理完全空壳的 legacy module（RegisterRoutes 已是 `_ = rg; return` 的文件可以考虑删除或合并）。
- 推进 Phase 5：前端切换到 v5 client，下线 legacy gin handler 文件本身（handler.go per module）。

### 2026-04-08 Phase 4 收尾：legacy 死代码清理（Phase 4）

**本次改动**
- 删除 8 个已完全被 ogen 接管的 legacy handler.go（featurepackage/permission/menu/page/role/navigation/app/space，共 ~4100 行）；从 menu/page 中提取孤儿 helper 函数到独立文件保持包内引用完整。
- 清理 role/navigation/collaborationworkspace/featurepackage/permission/menu/page/user/workspace 等 9 个 module.go 的早 return 后死代码，import 全部精简，文件平均缩至 30–50 行。
- `go build ./...` 通过，残留 legacy handler.go：collaborationworkspace/user/system/media/apiendpoint（均有活跃路由或仍被引用）。

**下次方向**
- Phase 5：前端切换 v5 client 后，继续下线 collaborationworkspace/handler.go 和 user/handler.go（需先将其中剩余 gin 路由的 handler 逻辑下沉到 service 层）。
- 考虑对 system/handler.go 做同样的 facade 化，统一通过 system/facade.go 暴露。

### 2026-04-08 Phase 5 handler 下线（Phase 5）

**本次改动**
- 删除 `collaborationworkspace/handler.go`（2266 行/74 函数）：无任何外部引用，全量安全删除。
- 删除 `user/handler.go`（1800 行/60 函数）：将 `subroute_service.go` 对 gin.UserHandler 的依赖重构为新建的 `subroute_core.go`（962 行），彻底去除 gin 依赖；重构后 UserHandler 零引用，安全删除。
- 净削减 ~3100 行，`go build ./...` 通过。剩余 legacy handler.go：`apiendpoint`/`media`/`system`（均有活跃 gin 路由或仍被 module.go 调用）。

### 2026-04-09 OpenAPI 契约继续收口（Phase 5）

**本次改动**
- 继续收紧 permission / workspace 域 OpenAPI：补齐 `PermissionActionList`、`PermissionActionOptions`、`PermissionActionGroupList`、`PermissionActionRiskAuditList`、`PermissionActionBatchTemplateList`、`CollaborationWorkspaceRoleList` 的强类型响应，关闭一批 `AnyListResponse` 兜底。
- 后端 handler 与 ogen 生成类型完成对齐，修正了 `OptString`、`OptInt64`、`OptDateTime`、`NilUUID` 等 wrapper 赋值问题，`go build ./...` 通过。
- 前端 `frontend/src/api/system-manage/permission.ts` 与 `frontend/src/api/collaboration-workspace.ts` 去掉对应列表接口的 `toV5ListResponse` 依赖，`npm run gen:api` 与 `npm run build` 通过。

**下次方向**
- 继续清理 permission 域剩余的弱类型列表返回，优先把 `permission-actions/{id}/endpoints` 这类接口从 `AnyListResponse` 收口成显式 schema。
- 继续压缩前端包装层的 `any` / 断言，优先把仍依赖兜底 normalizer 的接口逐个改成直接消费生成类型。

**下次方向**
- `system/handler.go`：检查是否仍被 system/module.go 调用；若仅剩 view-pages/fast-enter（已 ogen 化），可继续 facade 化后删除。
- `apiendpoint`/`media`：这两个域尚未纳入 openapi.yaml，待 spec 扩展后再迁移。
- 推进集成测试覆盖：ogen bridge 路径的冒烟测试。

### 2026-04-08 Phase 7 service 文件拆分（Phase 7）

**本次改动**
- `page/service.go`（2072→1407 行）拆出 `sync_service.go`、`runtime_service.go`、`breadcrumb_service.go`；`menu/service.go`（1821→1162 行）拆出 `backup_service.go`、`tree_service.go`；`featurepackage/service.go`（1669→859 行）拆出 `assign_service.go`、`audit_service.go`、`version_service.go`。
- 子代理生成的拆分文件未删除 service.go 原函数，导致重复声明；手动清除全部重复方法和孤立 import 后 `go build ./...` 再次绿灯。
- 构建绿，13/13 smoke test 通过，commit `c2fb57f`。

**下次方向**
- Phase 9 集成测试：对 role CRUD、navigation、cw boundary 等关键 ogen bridge 路径补充 `testcontainers-go` 驱动的真实 DB 集成测试。
- page/service.go 仍 1407 行，可视需求继续细拆（CRUD 与 sync 逻辑）；menu/service.go 1162 行同理。

### 2026-04-08 Phase 9 集成测试 + JSONB bug 修复（Phase 9）

**本次改动**
- 新增 `backend/internal/api/handlers/integration_test.go`（`//go:build integration`，10 个测试）：基于 live postgres，覆盖 login 正常/错误凭证/未知用户、auth/me 有/无 token、roles 有/无 token、navigation、feature-packages、collaboration-workspaces；全部 10/10 通过。
- 顺带修复 JSONB `?` 算符与 GORM 占位符冲突 bug：`app_keys ? ?` 在 GORM 中被错误替换为 `app_keys 'param'? ?`，导致 feature-packages list 接口返回 500；改为 `jsonb_exists(app_keys, ?)` 修复 user/repository.go、featurepackage/service.go、assign_service.go 共 4 处。
- 构建绿，13 smoke + 10 integration 全通过，commit `d7cf5e4`。

**下次方向**
- Phase 10 前端切换：将 frontend 的 API 调用从手写 axios 切换到 ogen 生成的 client（`openapi-typescript` + `openapi-fetch`）。
- 扩展集成测试：补充 role CRUD（POST/PUT/DELETE）、cw boundary 写操作、permission explain 的端到端验证。

### 2026-04-08 权限链路检查与修复（Permission chain）

**本次改动**
- 修复 evaluator 不遍历 bundle 层级的 bug：`queryFeaturePackageKeys`、`queryFeaturePackageKeysBySource`、`queryRoleKeys` 均只 JOIN `feature_package_keys` 直接键，忽略 `feature_package_bundles` 子包；改为 UNION CTE 后 bundle 包（admin_bundle）现可正确展开到 14→26 个 resolved keys。
- DB 补全：新增 12 条缺失 permission_keys（`workspace.read`、`user.list/create/update/delete/read`、`workspace.switch` 等）、为两个 workspace 创建 `workspace_feature_packages` 绑定（此前为 0 行，所有非 super_admin 用户的 resolved keys 均为空集）、将新 keys 绑定到对应 feature package。
- 构建绿，13 smoke + 10 integration 全通过，commit `1a2e731`。

**下次方向**
- 将 DB seed 固化进 goose migration（或 permissionseed/ensure.go），避免重建库后权限数据丢失。
- 针对非 super_admin 普通成员补充集成测试：验证 feature-package → workspace → role → permission key 交集链路正确拒绝/放行。

### 2026-04-08 API 管理裁剪：删除新增 API / 扫描配置（Phase 10）

**本次改动**
- 从 [backend/api/openapi/openapi.yaml](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/api/openapi/openapi.yaml) 删除 `POST /api-endpoints`、`GET/PUT /api-endpoints/unregistered/scan-config`，`openapi_seed.json` 重新生成后变为 195 个 operations。
- 同步清理手写后端入口：[backend/internal/api/handlers/apiendpoint.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/api/handlers/apiendpoint.go) 删除 `CreateApiEndpoint` / `GetApiEndpointScanConfig` / `SaveApiEndpointScanConfig`，[backend/internal/api/router/router.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/api/router/router.go) 删除对应路由。
- 重新运行 `go run github.com/ogen-go/ogen/cmd/ogen@latest --target api/gen --package gen --clean api/openapi/openapi.yaml` 与 `go run ./cmd/gen-permissions`，保证生成代码与权限种子同步收口；前端 `pnpm build` 通过。

**下次方向**
- 如果后续确认不再手工录入 API 注册项，可以继续收窄 API 管理页，仅保留同步、未注册、失效清理和分类维护。
- 若仍需要新增入口，建议单独做“兜底补录”而不是回到主流程，避免与 OpenAPI-first 的自动注册链路混淆。

### 2026-04-08 菜单域清理：删除备份功能 + 运行时 bug 修复（Menu 维护）

**本次改动**
- 删除菜单备份/恢复功能全链路：后端 backup_service.go、4 路由、4 handler、MenuBackup 模型/AutoMigrate/permissionseed/repository；前端 2 个 dialog 组件、4 个 API 函数、类型定义、及 menu/menu-space 两个 composable 和 view 中的全部引用。
- 修复运行时 bug：`fetchGetMenuTreeAll` 的 `res.map()` 改为 `res?.records?.map()`（后端返回包装对象而非裸数组）；`normalizeRuntimeMenuTree` 的 `meta.title` 加 `|| item?.title || item?.name` 回退，菜单标题正确显示。
- 修复代码审查问题：CreateMenu/UpdateMenu 补 Component 字段、N+1 查询改单 SQL、refreshAllMenuSnapshots 失败降级为 warn log、ref<any> 收窄类型、catch 块补 console.error。
- 修复 launch.json：backend 用 bash 绝对路径 cd 进子目录启动；frontend 用 pnpm -C；新建 .env 固定 VITE_PORT=5173。

**下次方向**
- Menu.Meta JSONB 普遍不含 title key，应在 Create/Update 时同步写入 meta.title，减少 normalizer 的 fallback 链依赖。
- 删除备份测试后 menu/service_test.go 覆盖偏少，可补 GetTree 和 countAffectedMenuRelations 单测。

### 2026-04-09 注册体系审计修复（S7 Bug Fix）

**本次改动**
- **P1-1 事务化**：`auth/service.go` 新增 `CreateUserTx` / `BuildLoginResponse`；`register/service.go` 将用户创建、审计字段回写、角色绑定（含 workspace binding snapshot）、功能包绑定全部合入单一 `db.Transaction`，任一步失败整体回滚，彻底消除半成品账号问题；移除多余的 `userRoleRepo` 依赖。
- **P1-2 路由**：`router.go` 补全 `/auth/register-context`（公开）+ 9 条治理 CRUD 路由，注册体系端点由 404 变为可达。
- **P1-3 策略子表**：`system_register.go` 新增 `upsertPolicySubTables` + `policySubTables`；创建/更新策略时在 tx 内删旧写新 `register_policy_roles` / `register_policy_feature_packages`；列表接口也返回子表内容。OpenAPI schema 同步新增 `role_codes` / `feature_package_keys` 字段。
- **P1-4 前端 auto_login**：`register/index.vue` 注册成功后若有 `access_token` 则 `setToken + setLoginStatus`，直接 push landing；`pending=true` 时跳登录页携带 `registered=1`。
- **P2-5 契约**：`LoginResponse` token 字段改 `nullable: true` 并移出 `required`，新增 `pending` 字段；`auth.go Register` 按 auto_login 分支赋值；ogen + 前端 TS 类型重新生成。go build + vue-tsc 均通过。

**下次方向**
- 策略版本快照：当前注册时使用实时 policy，历史用户的授权状态随策略修改漂移；可在 user 表加 `register_policy_snapshot jsonb` 冻结注册时刻的有效策略。
- 删除策略时未校验是否有入口仍在引用（`register_entries.policy_code`），可加守卫或级联禁用提示。
- `RegisterContext` 未包含 `field_schema` / `brand_info` / `captcha.site_key`，开启 RequireCaptcha / RequireInvite 后前端没有能力采集必填字段，需要一轮动态表单支撑。

### 2026-04-09 注册体系守卫 + 动态表单支撑（S7 续）

**本次改动**
- **删策略守卫**：`DeleteRegisterPolicy` 删除前检查 `register_entries.policy_code` 引用计数，有引用则返回 400 并提示数量，防止悬挂入口。
- **captcha 配置列**：migration `00007_register_captcha_config.sql` 给 `register_policies` 追加 `captcha_provider`（none/recaptcha/hcaptcha/turnstile）和 `captcha_site_key`，全链路贯通至 `EffectiveRegisterContext`、`RegisterContext` OpenAPI schema、策略 CRUD handler；重新生成 ogen + 前端 TS 类型。
- **前端动态表单**：`register/index.vue` 按 `ctx.require_email_verify / require_invite / require_captcha` 条件渲染邮箱、邀请码、验证码输入框，对应字段加入校验规则并随 `fetchRegister` 请求体携带；captcha_provider 非 none 时降级提示。
- **策略管理页**：`register-policy/index.vue` 补 captcha_provider 下拉、captcha_site_key 输入、role_codes / feature_package_keys 多选标签；go build + vue-tsc + migrate 全通过。

**下次方向**
- 接入真实 captcha widget（Turnstile/hCaptcha）：前端 `require_captcha=true && captcha_provider≠none` 时动态加载 widget SDK，回调写入 `captchaToken`，并在后端 `service.Register` 中接入对应 provider 的服务端验证 API。
- 策略版本快照：`users.register_policy_snapshot jsonb` 冻结注册时刻有效策略，防止策略变更后历史用户权限漂移。

### 2026-04-09 前端 OpenAPI 全量收口（Phase 10）

**本次改动**
- 补齐 OpenAPI 契约缺口：`backend/api/openapi/domains/permission/paths.yaml` 新增 `DELETE /permission-actions/groups/{id}`，同步补齐 `backend/internal/modules/system/permission/service.go`、`backend/internal/modules/system/user/repository.go`、`backend/internal/api/handlers/extras.go` 删除链路；重新生成 `backend/api/openapi/dist/openapi.yaml`、`backend/api/gen` 与前端 `src/api/v5/schema.d.ts`。
- 前端 API 层统一按生成 schema 收口：新增 `frontend/src/api/v5/types.ts` 与 `_shared.ts` 的 `AnyObject / AnyListResponse` 归一化辅助，批量改造 `message.ts`、`collaboration-workspace.ts`、`system-manage/{app,api-endpoint,menu,page,permission,role,user}.ts`，移除 `v5Client as any`/`as any` 逃逸，并把 `UUIDListRequest`、`required query`、`snake_case` 响应字段全部对齐到 OpenAPI 真相源。
- 验证通过：`go build ./...`、`npm run gen:api`、`npm run build` 全部绿灯，前端从“185 个接口里 1 个越契约 + 大量 any 逃逸”收口到生成 client 可编译状态。

**下次方向**
- 当前仍有一批 endpoint 在 spec 中使用 `AnyObject` / `AnyListResponse` / `MutationResult` 兜底，建议继续把 `permission`、`message`、`collaboration-workspace` 等域的响应 schema 精细化，减少前端 normalizer 对弱类型对象的依赖。
- 可继续补自动化校验：增加前端 API 包装层的 contract smoke test，或在 CI 增加“重新生成 schema 后无 diff + `vue-tsc` + `go build`”守门，防止后续再出现契约漂移。

### 2026-04-09 协作空间/权限响应契约强类型化（Phase 11）

**本次改动**
- `backend/api/openapi/domains/workspace/{paths,schemas}.yaml` 与 `domains/permission/{paths,schemas}.yaml` 新增协作空间详情/成员、边界菜单来源/权限来源、角色功能包继承态、权限消费者、影响预览等具体 schema，替换掉对应接口上的 `AnyObject` / `AnyListResponse`。
- 后端 handler 全量切到 ogen 生成的强类型响应：`collaborationworkspace.go`、`cwcurrent.go`、`cw_member.go`、`cw_boundary.go`、`permission.go` 改为返回 `CollaborationWorkspace*`、`PermissionAction*` 等结构体；新增 `openapi_response_helpers.go` 统一做模型到 OpenAPI 类型的转换。
- 前端 `collaboration-workspace.ts` 与 `system-manage/permission.ts` 同步改成直接消费生成类型，移除当前协作空间来源、角色功能包、权限影响预览等接口上的对象猜测逻辑；顺手修复 `current/boundary/packages`、`current/*-origins` 这类此前因契约过弱而会被前端解析成空数据的问题。
- 验证通过：`go build ./...`、`npm run gen:api`、`npm run build` 全部通过，`permissionseed` 与前端 `error-codes.ts` 已重新生成。

**下次方向**
- `permission-actions/options/groups/risk-audits/templates`、`collaboration-workspaces/current/roles`、`/{id}/roles` 等接口仍然在走弱类型 list 返回，下一轮可以继续细化为明确的 list item schema。
- 协作空间成员/协作空间详情当前仍只暴露后端真实字段，若前端需要 `user_name`、`nick_name`、`current_role_code`、`member_status` 等展示字段，应先在 service/repository 侧补真实聚合，再同步扩展 OpenAPI。

### 2026-04-09 页面/系统/消息/API 注册表强类型化继续推进（Phase 12）

**本次改动**
- 继续清理剩余 `AnyListResponse` 契约：`api_endpoint`、`featurepackage`、`page`、`system`、`message` 五个域新增显式列表 schema，覆盖 API 注册表、功能包版本/风险审计/协作空间绑定、页面列表与运行时/面包屑预览、系统应用与菜单空间绑定、消息收件箱与模板/发送人/接收组/投递记录等列表接口。
- 后端 handler 全部切换为 ogen 强类型返回，`apiendpoint.go`、`featurepackage.go`、`page.go`、`system.go`、`message.go` 不再依赖 `AnyListResponse` 兜底；`api_endpoint` 相关接口同时改为显式分页结构体，收口 API 注册表列表。
- 前端重新生成 `src/api/v5/schema.d.ts` 并通过 `npm run build`，后端 `go build ./...` 通过，说明这轮新 schema 没有破坏现有 client 和 handler 编译。

**下次方向**
- 继续向下收紧仍残留的弱类型单项响应，优先看 `AnyObject` 的 read-only 详情接口和 `permission/actions/{id}/endpoints` 这类还没有完全贴合 item schema 的路径。
- 前端包装层里仍有少量针对老响应形态的 normalizer，后续可以按域继续压缩，减少手工猜字段和 `item: any` 断言。

### 2026-04-09 OpenAPI 收口第二批：page/system/message 生成同步（Phase 12 续）

**本次改动**
- 继续收口 `message`、`page`、`system` 三域的 OpenAPI 契约，补齐 `InboxSummary`、`InboxDetail`、`InboxTodoActionRequest`、`MessageDispatchOptions`、`MessageDispatchRequest`、`MessageTemplateSaveRequest`、`MessageSenderSaveRequest`、`MessageRecipientGroupSaveRequest`、`MessageDispatchRecord`、`PageAccessTraceResponse`、`PageSaveRequest`、`PageSaveResult`、`SystemFastEnterConfig`、`SystemAppSaveRequest`、`SystemCurrentAppResponse`、`SystemAppHostBindingSaveRequest`、`SystemMenuSpaceEntryBindingSaveRequest`、`SystemCurrentMenuSpaceResponse`、`SystemMenuSpaceModeResponse`、`SystemMenuSpaceModeSaveRequest`、`SystemMenuSpaceSaveRequest`、`SystemMenuSpaceHostBindingSaveRequest` 等 schema。
- 重新 bundle + `ogen` + `openapi-typescript`，同步刷新 `backend/api/gen` 与 `frontend/src/api/v5/schema.d.ts`，让前后端生成物都跟上这轮 spec 变化。
- 验证上，`redocly lint` 只剩仓库既有的历史 warning，`go run github.com/ogen-go/ogen/cmd/ogen@latest ...` 与 `pnpm gen:api` 都已完成。

### 2026-04-10 OpenAPI 收口第三批：message/system/page handler 签名对齐

**本次改动**
- 继续把生成后仍停留在 `AnyObject / AnyListResponse` 的 handler 收紧为 ogen 生成类型，重点覆盖 `message`、`page`、`system`、`phase4_extras`、`apiendpoint`、`collaboration-workspace`、`permission`、`media`、`menu`、`user`。
- `backend/internal/api/handlers` 已通过 `go test ./internal/api/handlers`，说明这一批签名收口和返回体映射已能编译。

**下次方向**
- 继续追剩余历史弱契约接口，优先收 `apiendpoint`、`system` 其他写接口、`message` 剩余路径，并把前端调用层一并跟到生成 client。

**下次方向**
- 这一批 schema 里仍有少数定义为开放字典，ogen 生成后仍会落到 `AnyObject`，后续要继续把这些响应拆成字段明确的对象，才能真正把弱契约收干净。
- 下一批建议继续推进 `collaboration-workspace`、`permission`、`user` 这些还保留较多对象透传的接口，把读接口的返回体也进一步结构化。
### 2026-04-10 OpenAPI 收口第四批：前端 wrapper 与生成类型对齐

**本次改动**
- `frontend/src/api/system-manage/_shared.ts` 的 `toV5Body` 改为保留字段级类型信息，去掉把请求体统一压扁成 `Record<string, unknown>` 的弱化行为，消除新生成 schema 落地后的大部分 body 赋值错误。
- `frontend/src/api/collaboration-workspace.ts`、`frontend/src/api/system-manage/api-endpoint.ts`、`frontend/src/api/system-manage/page.ts`、`frontend/src/api/system-manage/user.ts` 按最新 OpenAPI 生成类型补齐必填字段与 snake_case 请求体，`PUT /users/{id}/menus` 也从旧 `{ ids }` 迁到显式菜单裁剪结构。
- 验证通过：`pnpm exec vue-tsc --noEmit` 已通过；`backend/api/gen/oas_server_gen.go` 当前已查不到 `AnyObject` / `AnyListResponse` 生成签名残留。

**下次方向**
- 继续把前端 API 包装层里仍依赖 `toV5Record` / `toV5ListResponse` 的读接口按域压缩，优先 `message`、`permission`、`collaboration-workspace`，减少 normalizer 对弱对象的兜底解析。
- 在更大范围上补一轮联编守门，建议至少固定执行 `go test ./internal/api/handlers`、`pnpm exec vue-tsc --noEmit` 与生成链 diff 检查，防止后续再出现 spec 与 wrapper 漂移。
### 2026-04-10 OpenAPI 收口第五批：message 域契约补强 + API endpoint 前端收紧

**本次改动**
- `backend/api/openapi/domains/message/{schemas,paths}.yaml` 补齐 `InboxListResponse`、`MessageDispatchResult`、`MessageTemplateItem/ListResponse`、`MessageSenderItem/ListResponse`、`MessageRecipientGroupItem/ListResponse`、`MessageDispatchRecordListResponse` 等显式 schema，`dispatch` 与模板/发送人/接收组写接口不再停留在 `MutationResult / AnyObject`。
- `backend/internal/api/handlers/message.go` 按新 ogen 签名重写列表、详情、发送与保存返回体，统一改成强类型响应映射；`go test ./internal/api/handlers -count=1` 已通过。
- `frontend/src/api/message.ts` 全量移除 `as unknown as`，按新生成 schema 补齐 inbox/dispatch/template/sender/recipient-group/record 的归一化；同时 `frontend/src/api/system-manage/api-endpoint.ts` 去掉了对 overview/list/category/stale/unregistered 的 `toV5Record / toV5ListResponse` 依赖。

**下次方向**
- 继续处理仍大量依赖 `toV5Record / toV5ListResponse` 的前端域，优先 `system-manage/app.ts`、`system-manage/permission.ts`、`system-manage/menu.ts`、`system-manage/user.ts`。
- `message` 域虽然已经切到强响应，但部分 normalizer 还在兼容历史字段，后续可以继续把前端 `Api.Message` 类型和 OpenAPI 真相源进一步收拢，减少双份字段语义。

### 2026-04-10 OpenAPI 收口第六批：system 写接口回正 + permission 批量治理结果显式化

**本次改动**
- `backend/api/openapi/domains/system/{schemas,paths}.yaml` 新增 `SystemAppItem/ListResponse`、`SystemAppHostBindingItem/ListResponse`、`SystemMenuSpaceEntryBindingItem/ListResponse`、`SystemMenuSpaceItem/ListResponse`、`SystemMenuSpaceHostBindingItem/ListResponse`、`SystemMenuSpaceInitializeResult`，并把 `saveApp/saveMenuSpace/save*Binding/saveMenuSpaceMode/updateFastEnterConfig` 这些前端实际消费对象的接口从 `MutationResult` 收回显式响应。
- `backend/internal/api/handlers/system.go`、`phase4_extras.go`、`extras.go`、`permission.go` 同步对齐 ogen 新签名；`system` 域列表/当前态/初始化结果改成强类型映射，`permission-actions/cleanup-unused`、`/batch`、`/groups`、`/templates` 也补上显式结果返回。
- `frontend/src/api/system-manage/app.ts`、`frontend/src/api/system-manage/permission.ts` 按新生成 schema 去掉一批弱契约解析，补齐 snake_case 字段消费；验证通过：`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit`。

**下次方向**
- 继续处理 `system` 域剩余仍保留开放 request schema 的写接口，尽量把 `System*SaveRequest` 也从 `additionalProperties: true` 收成显式字段，减少前后端对运行时 JSON 形状的猜测。
- `permission.ts` 里功能包相关接口仍有多处 `toV5Record + refresh_stats` 兜底，下一轮优先把 feature package 域的写接口返回也结构化，继续压缩弱解析面。

### 2026-04-10 OpenAPI 收口第七批：PageSaveRequest 显式化 + 生成链验证（Phase 12 续）

**本次改动**
- `backend/api/openapi/components/common.yaml` 将 `PageSaveRequest` 统一收敛为显式字段并加 `additionalProperties: false`，`backend/api/openapi/domains/page/paths.yaml` 的 `POST /pages`、`PUT /pages/{id}` 改为引用公共组件，消除 page 域内重复 schema 与旧定义撞名；重新执行 `bundle -> ogen -> gen:api` 后，`backend/api/gen/oas_schemas_gen.go` 中 `SchemasPageSaveRequest` 已生成强类型 struct。
- `backend/internal/api/handlers/page.go` 的 `CreatePage` / `UpdatePage` 改为直接接收 `*gen.SchemasPageSaveRequest`，字段逐项映射到 `page.SaveRequest`，移除 `AnyObject` 反序列化桥接；同时 `frontend/src/api/system-manage/page.ts` 改为显式构造 pages create/update body，不再对 `app_key/appKey` 做弱兼容解析。
- 顺手补齐 `frontend/src/api/system-manage/app.ts` 的 fast-enter 更新请求体映射，把可选 `FastEnter*Item` 收敛成符合生成 schema 的必填数组项，消除这轮类型收紧后暴露的前端编译错误。
- 验证通过：`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit`、`redocly bundle`、`ogen`、`pnpm run gen:api` 全部通过。

**下次方向**
- 继续收 `frontend/src/api/system-manage/permission.ts` 和 `frontend/src/api/message.ts` 中仍依赖 `toV5Body` 的写接口，把 permission group、inbox todo、dispatch、template/sender/recipient-group 保存接口改成显式 body。
- 评估是否继续把 `PageSaveResult` 与其他 page 读接口响应从开放对象收成显式 schema，进一步压缩 `normalizePageItem` 对弱响应形态的兼容路径。

### 2026-04-10 OpenAPI 收口第八批：permission/message 写接口显式 body 清尾（Phase 12 续）

**本次改动**
- `frontend/src/api/system-manage/permission.ts` 的功能权限分组创建/更新已改为显式构造 `PermissionActionGroupSaveRequest`，不再通过 `toV5Body` 兜底透传。
- `frontend/src/api/message.ts` 的 inbox todo、dispatch、message template、message sender、message recipient group 全部写接口已改成显式 `V5RequestBody`，字段逐项贴合生成 schema；`frontend/src/api/system-manage/user.ts` 里刷新权限快照的响应包装也去掉了 `toV5Body`。
- 验证通过：`pnpm exec vue-tsc --noEmit`、`go test ./internal/api/handlers -count=1` 均通过；当前 `frontend/src/api` 下已查不到 `toV5Body(` 调用残留。

**下次方向**
- 若继续深挖，可把 `PageSaveResult`、`PermissionActionBatchUpdateRequest` 等仍保留开放对象的契约继续拆成显式 schema，进一步压缩 `_shared.ts` 中的弱类型辅助。
- 同步审视 `toV5Record` / `AnyObject` 在 read-only normalizer 中的残留点，按域逐步替换为生成类型或显式 schema。

### 2026-04-10 OpenAPI 收口第九批：PageSaveResult + PermissionActionBatchUpdateRequest 显式化（Phase 12 续）

**本次改动**
- `backend/api/openapi/components/common.yaml` 新增显式 `PageSaveResult` 与 `PermissionActionBatchUpdateRequest`，并将 `backend/api/openapi/domains/page/paths.yaml`、`backend/api/openapi/domains/permission/paths.yaml` 改为引用公共组件，移除 page/permission 域内同名开放对象定义；重新执行 `bundle -> ogen -> gen:api` 后，后端与前端生成物都已切到明确字段结构。
- `backend/internal/api/handlers/page.go` 的 `GetPage`、`CreatePage`、`UpdatePage` 已对齐最新 ogen 指针签名，返回体改为显式映射 `gen.PageSaveResult`，不再通过 `marshalAnyObject` 兜底；`backend/internal/api/handlers/extras.go` 的 `BatchUpdatePermissionActions` 也改为直接接收 `*gen.PermissionActionBatchUpdateRequest`，移除 JSON 桥接。
- 验证通过：`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit` 均通过，说明 `PageSaveResult` 与权限批量更新请求体这两条链路已经完成 `OpenAPI -> handler -> frontend 类型` 闭环。

**下次方向**
- 继续清理仍残留 JSON 桥接的权限写接口，优先 `SavePermissionActionBatchTemplate`，把 permission 域剩余写接口也完全贴到生成类型。
- 继续收缩 page/system 的弱 meta 形态，评估 `PageSaveResult.meta`、system 若干 save request 中仍保留开放对象的字段是否要继续拆成显式结构。

### 2026-04-10 OpenAPI 收口第十批：permission 批量模板与端点列表收口

**本次改动**
- `backend/api/openapi/domains/permission/schemas.yaml` 将 `PermissionActionBatchTemplateSaveRequest` 收敛为显式请求体，并新增 `PermissionActionEndpointListItem`，`PermissionActionEndpointList.records` 不再沿用开放对象。
- `backend/internal/api/handlers/extras.go` 的 `SavePermissionActionBatchTemplate` 改为直接接收 `*gen.PermissionActionBatchTemplateSaveRequest`，只在 `payload` 边界保留动态字典转换；`backend/internal/api/handlers/permission.go` 的端点列表改成显式 item 映射。
- `frontend/src/api/system-manage/permission.ts` 继续对齐生成 schema，修正批量模板写接口的必填字段；`redocly bundle`、`ogen`、`pnpm run gen:api`、`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit` 均通过。

**下次方向**
- 继续向下收 `permission` 域里还保留开放对象的项级字段，优先 `RiskAuditItem`、`PermissionActionBatchTemplateItem.payload`、`PermissionActionEndpointListItem.category` 这类边界字段。
- 再往后就是继续压 `system/page` 的剩余 `meta` / `snapshot` 结构，确保 OpenAPI、handler、前端三层持续一致。

### 2026-04-10 OpenAPI 收口第十一批：system/featurepackage/media 收尾 + 前端 normalizer 清尾

**本次改动**
- `backend/api/openapi/domains/system/schemas.yaml` 把 app / host binding / menu space 相关 `meta` 收敛成命名的 `SystemMeta` schema，`backend/internal/api/handlers/system.go` 改用 `optSystemMetaToMap` 适配 `gen.OptSystemMeta`，不再直接把 `meta` 当 `AnyObject` 处理。
- `backend/api/openapi/components/common.yaml` 将 `FeaturePackageMenusResponse.menus` 收成显式 `FeaturePackageMenuItem`，`backend/api/openapi/domains/featurepackage/schemas.yaml` 新增 `FeaturePackageSnapshot` 命名 schema；`backend/internal/api/handlers/featurepackage.go` 不再用 `marshalAnyObject/marshalList` 透传菜单和快照。
- `frontend/src/api/auth.ts` 去掉 login / refresh 的 `as unknown as` 双重断言，`frontend/src/api/system-manage/permission.ts` 清掉不再需要的 `toV5Record` 导入；当前 `frontend/src/api` 下业务层已无 `toV5Record/toV5ListResponse` 调用残留。

**验证**
- `redocly bundle`
- `go run github.com/ogen-go/ogen/cmd/ogen@latest --target api\\gen --package gen --clean api\\openapi\\dist\\openapi.yaml`
- `pnpm run gen:api`
- `go test ./internal/api/handlers -count=1`
- `pnpm exec vue-tsc --noEmit`

**下次方向**
- 当前任务树只剩最终收尾节点，若继续就是做一次最终对账：确认文档、生成物和 task-tree 的完成态一致。

### 2026-04-10 OpenAPI 收口第十二批：审计补漏，修正 system meta 与 featurepackage snapshot 假收口

**本次改动**
- 审计发现第十一批存在“任务完成态先于源码收口”的问题：`backend/api/openapi/domains/system/schemas.yaml` 里 `SystemMenuSpaceEntryBindingItem`、`SystemMenuSpaceEntryBindingSaveRequest`、`SystemMenuSpaceItem`、`SystemMenuSpaceHostBindingSaveRequest` 仍残留 `AnyObject`。本次已统一改为 `SystemMeta`，并同步把 `backend/internal/api/handlers/system.go` 中 `SaveMenuSpaceEntryBinding`、`SaveMenuSpaceHostBinding` 切到 `optSystemMetaToMap(req.Meta)`。
- `backend/api/openapi/domains/featurepackage/schemas.yaml` 的 `FeaturePackageSnapshot` 不再是单纯命名后的开放对象，而是改成显式字段结构；`backend/internal/api/handlers/openapi_response_helpers.go` 新增基于真实持久化结构的字段级映射，将 `package_id`、`child_package_ids`、`action_ids`、`menu_ids`、`collaboration_workspace_ids` 收成 `uuid.UUID` / `[]uuid.UUID`，`snapshot_created_at` 收成 `time.Time`，彻底移除对快照的整包 JSON 透传。
- 重新执行 `redocly bundle`、`ogen`、`pnpm run gen:api` 后，`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit` 均通过；`backend/api/openapi/domains/system/schemas.yaml` 中已查不到 `AnyObject` 残留。

**下次方向**
- 当前仍有一个仓库卫生问题未处理：`docs/openapi-contract-closeout-temp.md` 是过期未跟踪临时文档，如要彻底收尾，需要在确认后删除，避免与 `docs/V5_REFACTOR_TASKS.md` 和 task-tree 的正式状态冲突。
- task-tree 当前仍显示旧的 `100%` 完成事件文本，后续若继续依赖它做审计记录，建议补一条“审计补漏完成”的事件或新建修正子任务，避免再次出现文档与源码状态漂移。

### 2026-04-10 OpenAPI 收口第十三批：生成链全量联编修复（Phase 12 收尾）

**本次改动**
- `backend/api/openapi/domains/page/paths.yaml` 修正 4 处 response 类型对调错误（`listRuntimePages`/`listPublicRuntimePages` 改为引用 `PageListResponse`，`listPageMenuOptions`/`listUnregisteredPages` 回正）；`backend/api/openapi/domains/message/schemas.yaml` 将 `MessageSenderMeta`、`MessageRecipientGroupMeta`、`MessageRecipientGroupTargetMeta` 的 `additionalProperties: false` 改为 `additionalProperties: {}`，使 ogen 正常生成 `Opt` 包装类型而非空 struct。
- `backend/internal/api/handlers/helpers.go` 定义本地 `anyObject = map[string]jx.Raw` / `optAnyObject` 类型，替换 ogen 不再生成的 `gen.AnyObject`；`page.go`、`message.go`、`extras.go` 同步对齐新签名，修复 `ListOptions` 返回 `[]models.UIPage` 需包装为 `[]page.Record` 的类型不匹配、Meta 字段从指针改 Opt 包装、`Payload.Set` 守卫缺失等问题。
- `frontend/src/api/system-manage/_shared.ts`、`frontend/src/api/collaboration-workspace.ts`、`frontend/src/api/message.ts`、`frontend/src/types/api/api.d.ts` 同步修正 schema 名称错误（`PageAccessTraceResponse`、`PermissionActionConsumersResponse`）、`null` vs `undefined` 不兼容、`BoxType` 字面量缺 `as` 断言及 `MessageDispatchRecord` 重命名等编译错误。
- 验证通过：`redocly bundle`、`ogen`、`pnpm run gen:api`、`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit` 全部通过。

**下次方向**
- 核查 `docs/openapi-contract-closeout-temp.md` 是否仍有价值，若无则删除，避免与正式任务文档状态冲突。
- 可考虑固化一次生成链 CI 守门脚本（bundle → ogen diff → vue-tsc），防止后续 spec 改动与 handler/前端类型再次漂移。

### 2026-04-10 协作文档约束回正（Phase 12 收尾）

**本次改动**
- 回正 `AGENTS.md`：将“5.0 重构进行中”“先 `make gen`”等失效表述替换为当前项目态，明确 `backend/api/openapi/ -> backend/api/gen/ -> frontend/src/api/v5/` 的生成链、`backend/internal/pkg/permission/evaluator` 的真实路径，以及新增/修改 API 后的固定执行步骤。
- 回正 `PROJECT_FRAMEWORK.md`：将后端/前端主框架从“将引入”改为“已落地”，补齐 `backend/cmd/*` 运行入口、OpenAPI 生成链与前端业务封装位置；同时把“4.5 旧术语统一硬切”修正为“历史术语允许存在于兼容层，但新增设计不得继续扩散旧模式”。
- 回正 `FRONTEND_GUIDELINE.md`：明确前端生成物位置、`v5Client` / schema 类型为唯一接口入口，normalizer 只负责归一化，不再鼓励继续扩散 `any` 解析或第二套请求体系。

**下次方向**
- 若后续继续调整协作纪律，应优先修改 `AGENTS.md` / `PROJECT_FRAMEWORK.md` / `FRONTEND_GUIDELINE.md` 这三份根文档，并同步检查 `docs/V5_REFACTOR_TASKS.md` 是否需要补阶段记录。
- 可以考虑再补一份简短的”单人 + AI 开发日常流程”约束，但应落在现有三份文档内，不再新增第四套平行规范。

### 2026-04-10 全量清尾：AnyObject/AnyListResponse 彻底删除 + 死代码扫清

**本次改动**
- `backend/api/openapi/components/common.yaml` 删除 `AnyObject`、`AnyListResponse`、旧版 `PermissionActionList`（含 AnyObject 引用）三个死亡 schema；`backend/api/openapi/openapi.root.yaml` 同步移除对应的三条 `$ref` 导出。重新执行 `redocly bundle → ogen → pnpm run gen:api` 后，生成产物中已完全查不到 `AnyObject`/`AnyListResponse` 残留。
- `PermissionActionList` 去掉命名冲突后，ogen 直接生成 `gen.PermissionActionList`（不再前缀 `Schemas`）；`backend/internal/api/handlers/permission.go` 同步对齐新名称。
- `backend/internal/api/handlers/helpers.go` 删除已无调用者的 `anyObject`/`optAnyObject`/`marshalAnyObject`/`marshalList`/`unmarshalAnyObject`/`optAnyObjectToMap` 六段死代码；`backend/internal/api/handlers/extras.go` 删除 `anyObjectToMap` 死代码。
- `backend/internal/api/handlers/featurepackage.go` 清理文件顶部过时注释（”Returns data via marshalAnyObject”已不成立）。
- `frontend/src/api/collaboration-workspace.ts` 移除不再使用的 `toV5Record` 导入；同时删除 `fetchCreateMyCollaborationWorkspaceRole`/`fetchUpdateMyCollaborationWorkspaceRole` 中历史遗留的 `priority` 字段（不在 spec schema 内）。
- `frontend/src/api/system-manage/_shared.ts` 删除 `V5AnyListResponse`/`toV5Record`/`toV5Records`/`toV5ListResponse` 四个已无业务层调用者的导出函数与类型。

**验证**
- `redocly bundle` ✓
- `ogen` ✓
- `pnpm run gen:api` ✓
- `go test ./internal/api/handlers -count=1` ✓（ok 0.080s）
- `pnpm exec vue-tsc --noEmit` ✓

**当前状态**
- `backend/api/openapi/` 与 `backend/api/gen/`：查不到 `AnyObject`/`AnyListResponse` 残留。
- `frontend/src/api/v5/schema.d.ts`：查不到 `AnyObject`/`AnyListResponse` 残留。
- `frontend/src/api/`：查不到 `toV5Record`/`toV5ListResponse`/`V5AnyListResponse` 残留。
- `backend/internal/api/handlers/`：查不到 `marshalAnyObject`/`unmarshalAnyObject`/`anyObjectToMap` 残留。
- 剩余已知开放项（受控动态边界，不属于技术债）：
  - `optSystemMetaToMap`、`permissionBatchTemplatePayloadToMap`（helpers.go）：仍在使用，对应 `SystemMeta`/`PermissionActionBatchTemplatePayload` 这两类有合理动态性的边界字段，属于受控保留。
  - `evaluator.go:51` `RoleKeys` TODO、`openapiperm.go:8` account-only path TODO、`MenuProcessor.ts:30` manifest.normalized TODO：三处已知未完成功能，文档级标注，不影响当前主链路。

### 2026-04-12 注册体系配置引导与页面归属收口（Phase 13）

**本次改动**
- 将公开认证页的真相从 `staticRoutes` 收回到 `account-portal`：`backend/internal/pkg/permissionseed/register_seed.go` 新增 `account-portal` 下 login/register/forget-password 三个 `public ui_pages` seed，前端新增 `frontend/src/views/account-portal/auth/*` 包装页，`/auth/*` 仅保留兼容跳转到 `/account/auth/*`。
- 重构 `frontend/src/views/system/register-entry/index.vue` 与 `frontend/src/views/system/register-policy/index.vue`：补了“先策略、再入口、后验证”的顶部说明区、默认模板入口、关键字段语义说明、命中 URL / landing / 字段要求预览，把“注册模板”收口为“注册策略预设”，不再新增领域对象。
- 补强 `frontend/src/views/auth/register/index.vue`：基于 `register-context` 直接展示命中的入口 Code、策略 Code、注册来源、landing 和必填项，并在未开放公开注册时禁用提交按钮、给出明确提示；`frontend/src/views/system/register-log/index.vue` 同步展示 `policy_snapshot`，便于审计当时实际生效的策略。
- 新增 `docs/register-setup-guide.md`，把页面归属、默认 seed、三类模板场景、配置顺序和验证步骤写成可直接执行的实施手册。

**下次方向**
- 让“页面管理”后台对 `account-portal` 公开页给出更显式的扫描/注册提示，减少运维还要理解 `/auth/*` 历史排除规则的心智负担。
- 若后续需要把邀请码、邮箱验证、人机验证做成真正可选的开箱能力，应继续补齐对应消息模板、验证码提供商配置和回执页，而不是只停留在字段开关层。
- 可以补一轮 Playwright 端到端验证，覆盖“命中入口 -> 展示字段 -> 注册成功跳转/回登录页”主链路，避免后续路由调整再次把公开页打回静态逻辑。

### 2026-04-12 多 APP Phase A 首批实现收口（Phase A）

**本次改动**
- 完成 `App` 运行时元数据首批落地：后端模型、`SaveApp` 服务、默认 app bootstrap、`account-portal`/`platform-admin` seed 与新迁移 `00010_app_runtime_metadata.sql` 已统一支持 `frontend_entry_url`、`backend_entry_url`、`health_check_url`、`capabilities`，并为两个默认 APP 写入首版能力默认值。
- 完成多 APP 契约扩展：`backend/api/openapi/domains/system/schemas.yaml` 新增 `SystemAppCapabilities` 并扩展 `SystemAppItem`/`SystemAppSaveRequest`；`backend/api/openapi/domains/page/schemas.yaml` 新增 `PageSpaceBindingItem`，运行时页面输出开始返回 `page_space_bindings`，治理侧与壳层都可以感知页面实际归属的 space 来源。
- 完成前端治理台与壳层首批接入：系统应用管理页补齐认证模式、入口 URL、健康检查 URL、capabilities JSON 表单；`ArtAppSwitcher` 接入顶栏，切换 APP 时会同步重置 `worktab`、动态路由、菜单树与当前 space，并按 APP 入口落点重新拉取运行时导航。
- 本轮已完成任务树三个大节点收口：`4/1`（模型/seed/migration）、`4/2`（API 契约与运行时导航）、`4/4`（UIPage/MenuSpace 绑定能力），当前下一可执行节点已切到 `4/3/1`。
- 验证通过：`redocly bundle`、`ogen`、`pnpm run gen:api`、`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit`、`go build ./...`。

**下次方向**
- 继续完成 `4/3/1` 到 `4/3/5`：让 `app-context` 真正基于 capabilities 驱动前端行为，补齐 APP 级路由替换/缓存清理、ErrorBoundary、本地存储命名空间和 Vite 分包边界。
- 继续推进 Phase A 后续节点时，应把 `account-portal` 作为公共认证中心的约束守住，只承接认证/注册/找回密码等公共能力，不把业务私有菜单和页面重新塞回公共壳。
- `redocly lint` 目前仍存在一条历史硬错误：`MenuSaveRequest.status.permission_keys` 在 bundle 后 spec 中位置异常；这不影响本轮生成、测试和构建，但后续应单独清掉，避免误判 OpenAPI 生成链异常。

### 2026-04-12 节点 4 全量收口：前端运行时隔离 + 集成验证完成（Phase A）

**本次改动**
- 一次性完成 `4/3/1` 到 `4/3/5`：`app-context` 增加 capabilities 感知与 runtime 上下文写入；`RouteRegistry` 注销/重建链增加 `ComponentLoader.clearCache()`；`IframeRouteManager` 改为按 `appKey` 分桶；`worktab` 的 `current/opened/keepAliveExclude` 改为按 app scope 切桶恢复。
- 落地 APP 级错误兜底：新增 `ArtAppErrorBoundary` 并挂到 `ArtPageContent` 的 keep-alive 与非 keep-alive 两条渲染链，错误日志带 `appKey` 作为 telemetry 上下文，并提供“重试/返回首页”恢复路径。
- 完成 APP 隔离存储命名空间：新增 `app-scope` 工具，`StorageConfig/StorageKeyManager` 支持 `sys-v{version}:app:{appKey}:{storeId}`；`menuSpaceStore/worktabStore` 切换到 APP 隔离 key，并兼容迁移旧 `menu-space/worktab` key。
- 完成分包命名规范：`vite manualChunks` 固化 `app-account-portal`、`app-platform-admin`、`app-demo`；构建产物已出现对应 chunk 名称。
- 完成 `4/5/1`、`4/5/2`：新增 `demo-app` seed（app/space/`/demo` 绑定/`/demo/lab` 页面）与前端 demo 页面，路由守卫放宽为 `menuTree` 或 `managedPages` 任一可注册即可，支撑轻量 APP 切换验证。

**验证**
- `gofmt -w backend/internal/pkg/permissionseed/register_seed.go` ✓
- `go test ./internal/api/handlers -count=1` ✓
- `pnpm exec vue-tsc --noEmit` ✓
- `pnpm run build` ✓（产物含 `app-account-portal` / `app-platform-admin` / `app-demo`）
- `go build ./...` ✓

**下次方向**
- 节点 `4` 已完成，任务树下一步已切到 `5/1/1`（Phase B 设计）。后续应先收口独立域名入口、跨域认证与网关分发规则，再进入 Phase B 编码。
- `demo-app` 当前是验证样本；若要长期保留，应补 feature package/role 绑定，避免变成“有入口、弱权限模型”的半成品。

### 2026-04-12 Phase B 实现收口：第六节独立域名编码全量完成（Phase B）

**本次改动**
- 完成第六节 8 个实现节点并回写任务树：后端新增动态安全中间件（按 APP 动态 CORS/CSP、shared_cookie 策略头）、APP 健康检查端点 `/health/apps`、日志 trace 标签（`app_key/space_key/auth_mode/request_id`）注入。
- 前端补齐 APP 运行时入口上下文（frontend/backend/health URL）并接入 `v5Client` 动态 base URL；路由守卫新增跨域 APP 入口兜底跳转；APP 切换器改为写入完整 runtime app context。
- 完成一轮迁移 + 重启 + 联编 + 浏览器验证：`go run cmd/migrate/main.go`、`go test ./internal/api/... -count=1`、`pnpm exec vue-tsc --noEmit`、`pnpm run build` 通过；8080/5174 服务监听正常，Playwright 验证 `/account/auth/login` 与 `/health/apps` 可访问。

**下次方向**
- 将 shared_cookie 从“策略头”推进到“真实会话 cookie 主链”（登录/刷新接口写入 HttpOnly cookie，并补 OpenAPI 契约与 E2E）。
- 在系统应用治理页补 `cors_origins/csp` 的结构化编辑与预校验，减少仅靠 JSON capabilities 维护的运维成本。
- 补 tenant context 中间件全链路注入，确保日志与 trace 的 `tenant_id` 字段从空值变为实值。

### 2026-04-12 导航收口：APP 切换与登录后回跳修复（Phase A）

**本次改动**
- 修复菜单空间 host binding 选择策略：`platform-admin/default` 同时存在 `localhost` 与 `127.0.0.1` 绑定时，前端改为优先匹配当前浏览器 host，不再盲取第一条绑定，登录后不再错误跳到 `http://localhost/`。
- 修复 `ArtAppSwitcher` 的入口回退链路：优先使用已注册路由与当前 space landing，再兜底 `frontendEntryUrl`，`account-portal` 切换后稳定落到 `/account/auth/login`，不再进入 `/account` 404。
- 修复路由守卫的运行时 app 上下文切换：当目标路径明显属于另一个 app（如 `/dashboard/*`、`/system/*`、`/account/*`）时，先切换 managed app 再刷新 runtime navigation，保证从 `account-portal` 回到后台时能重新加载 `platform-admin` 导航清单。
- 浏览器回归通过：`admin` 登录后落到 `http://127.0.0.1:5174/dashboard/console`；`platform-admin -> account-portal -> platform-admin` 全链路稳定，控制台 error 为 0。

**下次方向**
- 若后续继续做多 host/多域名部署，应把 host binding 的优先级规则沉到后端返回顺序或显式优先级字段，减少前端推断成本。
- 当前仍有 1 条非阻断 warning 未纳入本次节点处理；若后续复现，应单开节点按 warning 类型拆解，不与导航主链收口混做。

### 2026-04-12 Phase C 设计首轮收口：认证中心协议与职责边界（Phase C）

**本次改动**
- 在 `docs/multi-app-hosting-foundation.md` 补齐 Phase C 认证中心协议：明确 `account-portal` 为唯一认证中心，定义登录入口参数、callback 两段式回跳、一次性 code 交换、refresh 轮换与 logout 分层。
- 明确 branding 与回跳策略的责任分工：长期品牌配置真相源来自 `apps.meta.branding`，认证流程只消费 snapshot；后端决定“允许回跳哪里”，前端只负责展示和有限兜底校验。
- 补齐独立认证 APP 的职责上限与剥离路径：认证页不继续扩散成业务后台，也不再把后台菜单树、worktab、动态路由初始化当成登录前置条件。

**下次方向**
- 进入 `8/1/*` 实现节点时，先把 callback/logout/OpenAPI 契约补出来，再做 handler 与前端落地，避免先写代码再反推协议。
- 进入 `8/2/*` 前端适配节点时，优先收掉认证页对共享 `menu-space/worktab/route reset` 的依赖，把 `account-portal` 真正收敛成轻壳 APP。

### 2026-04-12 Phase C 实现收口：centralized login 回调链路打通（Phase C）

**本次改动**
- 完成 `8/1/1`、`8/1/2`、`8/1/3` 后端实现：OpenAPI 新增 `POST /auth/callback/exchange` 与 centralized login 参数扩展；新增 `auth_callback_codes` 模型与迁移 `00011_auth_callback_codes.sql`；`auth/login` 在 centralized 模式下改为签发一次性 callback code，`auth/callback/exchange` 负责 code 校验、token 交换与 landing 解析。
- 完成 `8/2/1`、`8/2/2` 前端适配：路由守卫未登录时会为目标 app 生成 `state/nonce` 并跳到 `account-portal`；新增 `/account/auth/callback` 页面接收 code、换取 token、初始化 `userStore` 并回跳原业务页；`userStore` 抽出 `applySession/clearSessionState` 统一登录态写入与清理。
- 补齐历史遗漏收口：`frontend/src/router/routes/staticRoutes.ts` 补上 `/account/auth/login|register|forget-password` 静态入口，避免 logout/guard 进入 account-portal 时直接 404；`backend/internal/api/router/router.go` 与 smoke test 同步补挂 `/auth/callback/exchange` public bridge，消除“spec 已生成但 Gin 未桥接”的假接通状态。
- 修正 redirect 白名单本地开发兼容性：callback 校验现在同时支持 `host:port` 精确匹配和 hostname 级匹配，本地 `127.0.0.1:5174` 回调不再因为白名单登记为 `127.0.0.1/localhost` 而误判失败。
- 浏览器端到端验证通过：未登录访问 `http://127.0.0.1:5174/system/page` 会跳到 `account-portal` 登录页；登录后进入 `/account/auth/callback`，随后成功调用 `/api/v1/auth/callback/exchange`、`/api/v1/auth/me`、`/api/v1/collaboration-workspaces/mine`、`/api/v1/runtime/navigation`，最终稳定回到 `/system/page`，控制台 `error = 0`。

**下次方向**
- 继续推进 Phase C 时，应把 `shouldUseCentralizedLogin` 从“按路径推断非 account-portal 全部走 centralized”收敛为真正读取 app `authMode/capabilities`，避免后续 demo 或本地单体场景被过度重定向。
- `frontend/src/api/auth.ts` 当前对 `POST /auth/callback/exchange` 仍用了一层 `as any` 调用，后续需要继续核查前端 `openapi-typescript` 生成层为什么没把该 path 正常暴露出来，并收掉这层临时绕过。
- 现阶段打通的是 callback token exchange 主链；全局 logout、refresh 跨 APP 一致性、shared cookie 与认证中心会话撤销仍是下一轮实现重点。

### 2026-04-12 Phase D 治理台收口：APP 注册中心分区与页面来源治理（Phase D）

**本次改动**
- `App 管理` 页面补齐注册中心治理分区：把基础标识、空间与认证、运行入口、能力声明、治理补充拆成显式表单段落，并新增三组治理卡片，直接展示入口绑定、前后端入口、健康探针、认证模式与能力声明完整度。
- 基于现有字段增加“接入预检查与本地预演”：不伪造新的 dry-run API，而是按当前 `hostBindings/frontend_entry_url/health_check_url/auth_mode` 静态推导入口命中、首跳落点和探针状态，让接入前检查从 JSON 文本变成可读治理视图。
- `受管页面` 页面补齐本地/扫描同步/远端页的来源分类：搜索条件新增来源筛选，指标区新增来源统计，列表行新增来源标签与治理提示，明确“扫描同步回源修正、远端页不重复挂本地组件”的共存约束。
- 前端验证通过：`pnpm exec vue-tsc --noEmit`、`pnpm build` 成功；Playwright 实测确认 `App 管理` 的治理卡片与 `受管页面` 的来源统计/治理告警已渲染。

**下次方向**
- 后端仍缺 `menu-space-entry-bindings` 运行时接口闭环，当前 `App 管理` 页会暴露历史 404；后续应把该接口补齐或显式下线，避免治理页长期带错误兜底运行。
- 这轮“预检查”仍是静态推导；如果后续进入真正接入托管阶段，应新增后端 dry-run/health 聚合接口，把 DNS、callback 白名单、CORS/CSP 与 issuer 校验统一下沉到服务端。
- `受管页面` 目前对“远端页”的识别依赖 `link/meta` 约定；后续如果接入远端 manifest，应把来源升级为明确契约字段，而不是继续靠前端推断。

### 2026-04-12 Phase D 治理台收口：APP 环境配置、Feature Flag 与敏感配置引用（Phase D）

**本次改动**
- 在 `App 管理` 抽屉新增三组结构化治理编辑区：`meta.env_profiles`、`meta.feature_flags`、`meta.sensitive_config`，把多环境配置、APP 级开关和敏感配置引用从自由文本说明拆成受控 JSON 编辑区。
- 保存链路改为“保留未知 meta + 合并受控治理键”：前端会在保存时只覆盖 `env_profiles/feature_flags/sensitive_config` 三个治理段，不丢历史上已经存在但当前页面未显式编辑的其他 `meta` 字段。
- 应用概览卡片新增环境/Flag/敏感治理统计，页面层能直接看到“环境 0 组 / Flag 0 项 / 敏感治理 0 组”这类覆盖度，不必先打开抽屉才能判断配置空缺。
- 浏览器实测确认编辑抽屉已出现“环境配置（meta.env_profiles） / Feature Flag（meta.feature_flags） / 敏感配置引用（meta.sensitive_config）”三项，前端联编与构建继续通过。

**下次方向**
- 当前只是把治理约定落到前端结构化编辑层，后续若要真正消费这些配置，需要在运行时 app-context、部署脚本或后端聚合接口里明确 `env_profiles/feature_flags/sensitive_config` 的真读取方。
- `sensitive_config` 当前仍是“引用治理”而非真实密钥托管；下一步若推进部署集成，应把 key vault / 环境变量 / 证书管理系统的来源类型也标准化，避免不同 APP 各自发明字段。
- `menu-space-entry-bindings` 历史 404 仍会污染 `App 管理` 页的控制台；在继续扩治理台之前，建议先补齐这个后端接口闭环或在前端显式降级提示。

### 2026-04-12 Phase D 历史遗留收口：menu-space-entry-bindings 路由桥接补齐（Phase D）

**本次改动**
- 在 [backend/internal/api/router/router.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/api/router/router.go:256) 补挂 `DELETE /system/app-host-bindings/:id`，并补齐 `GET/POST/DELETE /system/menu-space-entry-bindings` 三条 Gin → ogen 桥接路由。
- 后端 handler 与 OpenAPI 其实早已存在，这次修复的是运行时桥接缺口，不涉及 OpenAPI 或 ogen 重新生成。
- 重启后端后重新验证，`App 管理` 页控制台不再出现 `menu-space-entry-bindings` 404，历史遗留已收掉。

**下次方向**
- `menu-space-host-bindings` 这一组路由目前仍只有 `GET/POST` 桥接；若前端后续补删除能力，应同步检查 OpenAPI 与 Gin 路由是否一致，避免再次出现“spec/handler 在、路由没挂”的半接通状态。
- 这类问题本质上是路由桥接遗漏；后续如果继续扩 OpenAPI-first 链路，建议补一层“spec path 与 Gin bridge 对账”的自动检查，别再靠浏览器控制台发现。 

### 2026-04-12 V2 第一波推进：APP 接入 dry-run / preflight 后端闭环（Phase V2）

**本次改动**
- 新增 `GET /system/apps/preflight` OpenAPI 契约，并落地 `summary / checks / preview_items` 结构化返回，把治理台“接入预检查与本地预演”从前端静态推导切到后端聚合真相源。
- 后端 `app` 服务新增 `GetAppPreflight`，聚合 app、Level 1/Level 2 入口、callback host、能力声明、前后端入口与健康探针配置，输出接入检查项和预演结果。
- 前端 `App 管理` 页已接入 preflight 接口，预演卡片改为直接展示后端返回的 `入口命中 / 首跳落点 / 健康探针` 等结果；Gin bridge 同步补挂 `/api/v1/system/apps/preflight`。
- 已按 OpenAPI-first 链路完成 `bundle -> ogen -> gen-permissions -> pnpm run gen:api`，并通过 `go test ./internal/api/handlers -count=1`、`go build ./...`、`pnpm exec vue-tsc --noEmit` 与浏览器实测验证。

**下次方向**
- 继续执行 V2 `1.2`，把远端页面 / manifest / health / version 升级为治理后端真契约，收掉页面来源仍靠前端启发式推断的问题。
- 继续执行 V2 `4.1`，补 spec path / Gin bridge / 权限绑定自动对账，避免以后再出现“spec 已有但运行时仍 404”的半接通状态。
- 当前 preflight 仍是注册中心配置聚合，不做外部网络探测；真正 probe/manifest 深度校验应放到 `1.2` 与 `5.2` 继续收口。

### 2026-04-12 V2 第一波收口：OpenAPI 对账测试、陈旧引用清理与 facade 拆薄（Phase V2）

**本次改动**
- 在 `backend/internal/api/router/router_contract_test.go` 新增静态对账测试，直接用 `openapi_seed.json` 对比 Gin bridge 注册，校验 `spec path / access_mode / permission_key` 与 `router.go` 是否一致，避免再靠浏览器 404 才发现桥接遗漏。
- 按测试结果修正 `backend/internal/api/router/router.go`：补挂 `DELETE /permission-actions/groups/:id` 与 `POST /api-endpoints/categories`，删除已脱离 spec 的陈旧桥接 `POST /api-endpoints` 和 `/menus/groups/*`。
- 清理已失效旧测试桩：删除 `backend/internal/modules/system/app/handler_test.go`，并把 `service_test.go` 中两条历史英文错误断言同步到当前中文契约。
- 收掉 `fetchExchangeAuthCallback` 的 `as any` 临时绕过；同时把 `GetFastEnterConfig`、`UpdateFastEnterConfig`、`GetSystemViewPages` 三条 Phase 4 leftover handler 从 `system.Facade` 拆到显式 `FastEnterService` / `ViewPagesService` 依赖，降低继续堆积 legacy facade 的风险。
- 回归通过：`go test ./internal/modules/system/app -count=1`、`go test ./internal/api/handlers -count=1`、`go test ./internal/api/router -count=1`、`go build ./...`、`pnpm exec vue-tsc --noEmit`。

**下次方向**
- 优先切 V2 `2.1`，把前端 `shouldUseCentralizedLogin` 从路径前缀猜测改成读取 app `authMode/capabilities`，把认证行为真正绑定到 app 元数据。
- 继续推进 V2 `1.2` 时，建议直接复用这轮 preflight 输出，把远端页 `manifest / health / version` 聚合结果也纳入治理后端真相源，而不是再让前端猜 `link/meta`。

### 2026-04-12 V2 第二波推进：认证公开页收口、远端契约显式化与治理观测补齐（Phase V2）

**本次改动**
- 修正公开壳层未登录态的错误请求：`ArtHeaderBar` 未登录时不再拉消息摘要，`ArtAppSwitcher` 未登录时不再请求 APP 列表，收掉 `demo-app` 公开页被 401 覆盖成普通登录的历史边角；浏览器实测 `/demo/lab` 未登录可稳定停留，`/system/page` 未登录仍会进入带 `target_app_key/redirect_uri/target_path/state/nonce` 的 centralized login URL。
- 新增 [docs/multi-app-playwright-smoke.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/multi-app-playwright-smoke.md)，把 admin/account/demo 三类主链固化成 smoke matrix，并补“观测与排障字段”章节，明确 `app_key / space_key / auth_mode / request_host / probe_status / request_id` 的使用入口。
- 扩展 OpenAPI page/system schema：`SystemAppItem` / `SystemAppPreflightResponse` 新增 `manifest_url / runtime_version / probe_status`，`PageListItem` 新增 `remote_binding`；后端从现有 `meta` 归一输出这些字段，前端页面来源判定优先读取 `remoteBinding`，不再主要依赖 `link/meta` 猜测。
- `App 管理` 的预检查与治理卡片开始展示 manifest/version/probe 相关信息，preflight 预演项新增“诊断标签”，把 `app/auth/probe/request_host` 组合成可直接用于排障的治理视图。
- 已按 OpenAPI-first 链路执行 `bundle -> ogen -> gen-permissions -> pnpm run gen:api`，并通过 `go test ./internal/api/handlers -count=1`、`go build ./...`、`pnpm exec vue-tsc --noEmit`。

**下次方向**
- `SystemAppSaveRequest` 与 `PageSaveRequest` 目前仍是“读链显式、写链兼容”：后续应把 `manifest_url / runtime_version / remote_binding` 的保存契约也显式化，不再只回写到 `meta`。
- `probe_status` 当前还是配置态（`configured/missing`），不是实际探活结果；继续推进 `5.2` 时应把最近一次 probe 结果与错误摘要纳入治理台。
- `redocly lint` 当前仍被仓库既有 spec 结构错误阻断（`MenuSaveRequest.status.permission_keys`），这不是本轮新增改动引入的问题，但后续若继续强化 OpenAPI-first，应把该旧问题独立收口。

### 2026-04-12 V2 与残余专项并行收口：共享会话、显式写链、真实探活与 OpenAPI 严格链恢复（Phase V2）

**本次改动**
- 完成 V2 `2.2 / 2.3 / 3.2` 与残余专项 `1.1 / 1.2 / 2.1 / 2.2 / 3.1 / 3.2` 的并行收口：`SystemAppSaveRequest` 显式新增 `manifest_url / runtime_version`，`PageSaveRequest` 显式新增 `remote_binding`，后端保存链统一写入规范 snake_case 契约并清理旧 camelCase 别名，前端保存链同步显式透传，不再只靠 `meta` 隐式兼容。
- `app` 治理链升级为真实探活：`SystemAppItem` 与 `SystemAppPreflightResponse` 新增 `probe_status / probe_target / probe_message / probe_checked_at`，后端按 `health_check_url + primary host` 做实际 HTTP 探测并把结果下沉到 `App 管理`；前端预检查区和治理卡片改为直接消费真实 probe 结果，不再把“已配置地址”误当成“运行正常”。
- 完成跨 APP 会话与壳层联动硬化：`app-context` 开始真实消费 `env_profiles / feature_flags / capabilities`，运行时前后端入口与 health 地址支持按环境 profile 回落；`demo-app` seed 切到 `shared_cookie` 可验证模式；`v5 client` 增加 shared session 下的 401→refresh→单次重试链路；`userStore` 增加 storage 广播同步与统一 session 清理；`logout` 新增 OpenAPI 契约、Gin bridge 与显式 `Authorization` 透传，退出登录后可稳定回到登录页。
- 恢复 OpenAPI-first 严格链：修复 `MenuSaveRequest.permission_keys` 历史结构错误，补挂 `/auth/logout` root path 与 Gin bridge，完整执行 `bundle -> ogen -> gen-permissions -> pnpm run gen:api`；`redocly lint` 现已恢复为“valid with warnings”，不再被结构错误阻断，只剩仓库存量 warning。
- 浏览器联调通过：登录后访问 `App 管理`，`platform-admin` 预检查明确显示 `probe=unreachable` 与真实目标 `http://localhost/health`；`platform-admin -> demo-app -> platform-admin` 切换稳定；退出登录后 `POST /api/v1/auth/logout` 返回 `200`，受保护后台页会回到 `account-portal` 登录页，`demo-app` 公开页在登出后仍可直接访问且当前页面无新增 console error。

**下次方向**
- 这轮已经把“probe 是真实结果”补起来，但探活仍是同步轻探针；后续如需治理大规模远端应用，应改成后台异步采集 + 最近结果缓存，避免列表页阻塞在外部网络波动上。
- `redocly lint` 现在剩余的是仓库级 warning，不再是结构断路；如果要继续提升 OpenAPI 质量，建议单开规范治理任务，批量补 tag description、4xx response 和 ambiguous path。
- `Page.remote_binding` 当前已显式进写链，但 UI 仍以“保留并透传现有远端契约”为主；若后续要支持后台直接新建远端页，应再补专门的 remote binding 编辑表单。 

### 2026-04-12 OpenAPI 固定闭环文档化：基础约束与详细流程固化（Phase Cleanup）

**本次改动**
- 在 [AGENTS.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/AGENTS.md) 补充“新增能力固定顺序”短链，统一为 `model/domain → migration → seed/ensure → OpenAPI spec → bundle → lint → ogen → gen-permissions → gin bridge/router → handler/service → frontend gen:api → frontend API 封装 → UI → build/test/browser verify`，并明确 API 网关配置在本仓库内等价于 `OpenAPI 扩展字段 + gin bridge/router + middleware 分组`。
- 在 [PROJECT_FRAMEWORK.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/PROJECT_FRAMEWORK.md) 新增“新增能力固定闭环”章节，把同一条顺序提升为项目基础框架约束，补齐“先定模型和默认数据策略，再定契约和桥接”的统一口径。
- 新增 [docs/API_OPENAPI_FIXED_FLOW.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/API_OPENAPI_FIXED_FLOW.md)，把固定顺序展开成详细执行文档，覆盖 `model / migration / seed / OpenAPI / bundle / lint / ogen / gen-permissions / gin bridge / handler / frontend gen:api / frontend API 封装 / UI / build/test/browser verify` 全流程，并整理常见疏漏点，供后续沉淀成标准流程文本。

**下次方向**
- 如果后续要把这套流程进一步做成团队模板，建议再补一版“最小执行清单”，按“新增 API / 改契约 / 改权限 / 改默认数据”四类场景拆成 checklist，而不是只保留通用大链路。
- 当前这轮只做流程文档化，不涉及代码或契约变更；后续若再调整 OpenAPI 生成链，应同步回写这三份文档，保持基础约束与执行文档一致。 

### 2026-04-12 前端收口 Phase P2-A 第一段：运行时主链设计与入口归并（Phase Cleanup）

**本次改动**
- 先完成 P1 验证闭环：补做 `vue-tsc`、`build`、目标文件 `eslint` 与认证页本地路由 smoke，P1-A / P1-B 当前实现已收口，真实后端与集中登录的浏览器级断言映射到后续 P3 最小回归集。
- 新增 `docs/frontend-cleanup-p2a-notes.md` 的目标主链设计章节，明确 `session / app-context / menu-space / runtime-navigation / route guard` 五层职责、四个触发场景的目标调用顺序，以及与现状实现的 6 条 diff 清单。
- 新增 `frontend/src/router/runtime/session.ts`，并在 `userStore` 中暴露 `restoreSession()`；`beforeEach.ts` 与 `auth-flow/shared.ts` 已统一改走该入口，guard 内联的 session 恢复逻辑被删成薄包装。
- 新增 `frontend/src/router/runtime/app-context.ts`，并在 `appContextStore` 中暴露 `ensureRuntimeAppKey()` / `switchApp()`；`menu-space.ts` 不再自己兜底 `/system/current-app`，`ArtAppSwitcher.vue` 已改成只调用 `appContextStore.switchApp(targetApp)`。
- 回归通过：`pnpm exec vue-tsc --noEmit`、针对修改文件的 `eslint`、`pnpm build` 均通过；仅剩仓库历史 `any` warning，未新增错误。

**下次方向**
- 继续做 P2A-5，把 `beforeEach.ts` 进一步瘦成分发层，收掉剩余的编排职责和兼容导出，让 `navigation-runtime` 成为唯一动态路由主链入口。
- 当前 `appContextStore.switchApp()` 内部仍通过 `refreshUserMenus()` 复用旧导航层；下一步可把 manifest 拉取、route register、homePath/worktab 收口进独立 `navigation-runtime`，彻底消除 guard/组件对细节的感知。
- P1 浏览器级真实回归仍未在当前线程跑通，因为内置浏览器会话被占用且没有真实后端/集中登录状态；待 P3 E2E 基础设施就绪后，应把登录成功、register auto_login/pending、callback exchange、401/403/业务码展示补成自动化断言。

### 2026-04-12 前端收口 Phase P2-B：兼容层盘点、打标与第一批清理（Phase Cleanup）

**本次改动**
- 新增 `docs/frontend-cleanup-p2b-notes.md`，把兼容层拆成 `legacy HTTP / v5->旧类型桥接 / 旧路径兼容 / 运行时 façade / worktab-app-scope 残留` 五类，并补齐代表文件、关键行号、用途判断与建议状态。
- 在 `frontend/src/utils/http/index.ts`、`frontend/src/api/system-manage/_shared.ts`、`frontend/src/api/auth.ts`、`frontend/src/api/workspace.ts`、`frontend/src/router/routes/staticRoutes.ts`、`frontend/src/store/modules/user.ts`、`frontend/src/store/modules/worktab.ts`、`frontend/src/utils/app-scope.ts` 与 account-portal 认证壳层中补充 `@compat-status: keep|transition` 标记。
- 删除第一批已可安全移除的 façade：`frontend/src/router/guards/beforeEach.ts` 中旧的 `refreshUserAccessAndMenus / refreshUserMenus / refreshCurrentUserInfoContext` 以及 `frontend/src/router/index.ts` 中对应 re-export；调用方已改为直接依赖 `router/runtime/session` 与 `router/runtime/navigation`。
- 已同步补完迁移时间表：明确 `api/workspace.ts` 与 account-portal 薄壳适合本轮继续收口，`api/auth.ts` 与旧路径重定向适合下轮随业务迁移，`_shared.ts` 与 legacy axios 链需要专项任务拆解。

**下次方向**
- 优先处理 `frontend/src/api/workspace.ts` 这类调用面窄的桥接文件，尝试在不扩散风险的前提下再拿下一批 `transition` 项。
- `_shared.ts` 与 `utils/http/index.ts` 已经确认是大范围兼容枢纽，后续应单开“去桥接化”子任务，按菜单、页面、权限、用户、消息等领域逐块替换，避免一次性爆破。
- 旧认证路径重定向与 account-portal 壳层是否可删，取决于中心化登录回跳、历史链接和外部文档是否全部切换；这部分应在 P3 浏览器回归矩阵里一并验证后再动。 

### 2026-04-12 前端收口 Phase P3-A 第一段：Playwright E2E 基线落地（Phase Cleanup）

**本次改动**
- 在 `frontend` 新增 Playwright 测试基线：安装 `@playwright/test`，补充 `package.json` 的 `test:e2e / test:e2e:headed / test:e2e:ui` 脚本，并下载 Chromium 浏览器。
- 新增 `frontend/playwright.config.ts`，约定 `e2e/` 为测试目录，使用 `webServer` 自动拉起 `vite --mode test`，并建立 `setup -> chromium` 两阶段项目结构。
- 新增 `frontend/e2e/setup/auth.setup.ts`、`frontend/e2e/tests/smoke.spec.ts`、`frontend/.env.test.example`；其中 setup 先保证共享 `storageState` 文件存在，smoke 用例验证公开登录页能正常打开。
- `.gitignore` 已补 `playwright-report/`、`test-results/`、`e2e/.auth/`，避免本地产物污染仓库；本地已生成但不入库的 `frontend/.env.test` 用于测试模式运行。

**下次方向**
- 直接进入 `P3A-2`，把“未登录访问后台页 -> 重定向到登录页”补成真实断言，顺带把 centralized login URL 的 `target_app_key / redirect_uri / target_path / state / nonce` 一起验掉。
- 当前 `auth.setup.ts` 只做空 `storageState` 初始化，后续在需要登录态复用时再接入 `.env.test` 的测试账号与真实登录流程。
- Playwright 目前仍运行在本地前端 + 本地代理模式；若后续要稳定覆盖 callback / refresh / 401 等场景，需要补专门的测试账号和可控后端环境。 

### 2026-04-12 前端收口 Phase P3-A 第二段：未登录重定向与登录失败提示回归（Phase Cleanup）

**本次改动**
- 新增 `frontend/e2e/tests/auth-redirect.spec.ts`，覆盖“未登录访问 `/system/page` 时应回到 `/account/auth/login`”主链，并兼容当前 centralized login 参数模式，校验 `target_path=/system/page` 与 `target_app_key=platform-admin`。
- 新增 `frontend/e2e/tests/login-error.spec.ts`，通过拦截 `/api/v1/auth/login` 返回 401 业务错误，验证错误凭据登录后会停留在登录页并显示“用户名或密码错误”提示，且表单仍可重新输入。
- 已执行全量 `pnpm exec playwright test`，当前 `setup + smoke + auth-redirect + login-error` 共 4 条用例全部通过。

**下次方向**
- 继续推进 `P3A-4`，把“登录成功 -> 进入系统主页”补成断言；这一步优先评估是否使用 `.env.test` 真实账号，还是继续用接口拦截方式先稳定前端链路。
- 当前未登录重定向测试已经观察到前端会走 centralized login 参数链；后续应把 `redirect_uri / state / nonce / auth_protocol_version` 也加进显式断言，而不是只验 `target_path` 和 `target_app_key`。
- 若后续要覆盖 callback / refresh / 401 自动恢复，需要在 `auth.setup.ts` 基础上引入真实登录态生成和更完整的 API mock/测试后端。 

### 2026-04-12 前端收口 Phase P3-A 第三段：登录成功主页态回归（Phase Cleanup）

**本次改动**
- 新增 `frontend/e2e/tests/login-success.spec.ts`，用 API mock 方式覆盖“正确凭据登录 -> 进入 `/dashboard/console` -> 菜单渲染 -> 会话持久化 -> 用户菜单展示”的完整主页态链路。
- 用例中补齐了登录后真实会触发的运行时请求 mock，包括 `/system/apps`、`/messages/inbox/summary`、`/runtime/navigation`、`/auth/me`、`/workspaces/my` 等，避免页头应用切换器和消息中心把测试拖回真实后端。
- 持久化断言已按当前 Pinia 版本化存储实现收口，不再错误假设 key 固定为 `user`，而是验证 `sys-v{version}-user` 条目中包含 token 和用户邮箱。
- 已执行 `frontend` 下 `pnpm exec playwright test e2e/tests/login-success.spec.ts` 与全量 `pnpm exec playwright test`，当前 5 条用例全部通过。

**下次方向**
- 直接进入 `P3A-5`，补“已登录状态下切换 App -> 菜单树与 app-context 同步刷新”的 E2E；若当前测试壳只有单 app，则按任务说明改成 conditional mock。
- 当前登录成功用例仍是前端 API mock 链路，后续若补真实测试账号，可在现有结构上再加一条真实登录版本，不需要推翻当前稳定用例。
- 头部组件还会拉取更多已登录态接口；后续继续补 logout、demo 公开页与管理页 CRUD 时，应沿用“按真实页面依赖把最小接口集一次 mock 完整”的策略，避免被无关 500 干扰断言。 

### 2026-04-12 前端收口 Phase P3-A 第四段：App 切换菜单与上下文回归（Phase Cleanup）

**本次改动**
- 新增 `frontend/e2e/tests/app-switch.spec.ts`，覆盖“已登录 -> 切换到另一 app -> 菜单树刷新 -> 路由跳转 -> app-context 持久化更新”的主链。
- 用例通过双 app mock 方式模拟 `platform-admin` 与 `demo-app`，并让 `/system/apps`、`/system/menu-spaces`、`/system/menu-spaces/current`、`/runtime/navigation` 按 `app_key` 返回两套运行时数据，真实驱动 `appContextStore.switchApp()` 的切换流程。
- 切换后断言已覆盖：URL 跳到 `/demo/lab`、侧边栏出现 `Demo 应用 / Demo 实验室` 新菜单、页面渲染 `Demo App 验证页`、版本化 `appContextStore` 持久化条目中包含 `demo-app`。
- 已执行 `frontend` 下 `pnpm exec playwright test e2e/tests/app-switch.spec.ts` 与全量 `pnpm exec playwright test`，当前 6 条用例全部通过。

**下次方向**
- 继续推进 `P3A-6`，补退出登录后本地状态清理与回到登录页的 E2E，优先复用现有登录成功 mock 链路，不重复搭环境。
- 目前 app 切换测试验证的是“同域 shared_cookie + 双 app runtime 菜单切换”，尚未覆盖跨 host / centralized login 切换；如果后续需要，再单列场景扩展。
- 头部壳层已确认至少依赖 app 列表、消息摘要、workspace 和 runtime navigation；后续新增 E2E 时应继续按页面真实依赖补齐 mock，避免出现伪失败。 

### 2026-04-12 前端收口 Phase P3-A 第五段：Logout 清理与多标签拦截回归（Phase Cleanup）

**本次改动**
- 新增 `frontend/e2e/tests/logout.spec.ts`，覆盖“已登录 -> 执行登出 -> 会话清理 -> 再访后台被拦回登录页”的链路，并在同一浏览器 context 下额外打开第二个 tab 验证跨标签页同步。
- 用例继续复用登录成功主链 mock，同时通过 `context.route()` 统一覆盖全部页面请求，并显式校验 `/api/v1/auth/logout` 已被调用，避免只测到前端假清理。
- 断言内容已覆盖：版本化 user 持久化条目不再包含 `mock-access-token`、当前页重新访问 `/dashboard/console` 会回到 `/account/auth/login`、第二个 tab 再访后台也同样被拦截。
- 已执行 `frontend` 下 `pnpm exec playwright test e2e/tests/logout.spec.ts` 与全量 `pnpm exec playwright test`，当前 7 条用例全部通过。

**下次方向**
- 继续推进 `P3A-7`，补 `demo-app` 公开页在未登录态下可直接访问的 E2E，并与受保护后台路由形成一正一反的访问控制对照。
- 当前 logout 用例为了稳定验证核心逻辑，直接调用了页面里的 `userStore.logOut()`，没有继续卡在 Element Plus 确认框 UI 壳层；如果后续要补完整交互，可再追加一条纯 UI 版本。
- 到这一段为止，P3-A 已经覆盖登录成功、登录失败、App 切换和 Logout；剩余重点将转向公开页与管理页 CRUD 的真实页面能力验证。 

### 2026-04-12 前端收口 Phase P3-A 第六段：Demo 公开页访问控制回归（Phase Cleanup）

**本次改动**
- 新增 `frontend/e2e/tests/demo-public.spec.ts`，覆盖未登录访问 `demo-app` 公开页 `/demo/lab` 的链路，验证 `ensurePublicRuntimeRoutes()` 能在登录校验前正确注册公开运行时页面。
- 用例只对 `/api/v1/pages/runtime/public` 做最小 mock，下发一条 `access_mode=public`、`component=demo/lab/index` 的公开页面记录，避免把后台菜单/runtime navigation 依赖带进公开页测试。
- 断言内容已覆盖：访问 `/demo/lab` 后 URL 仍停留在 `/demo/lab`、页面成功渲染 `Demo App 验证页` 与 `当前 APP：demo-app`、全程不发生登录页重定向。
- 已执行 `frontend` 下 `pnpm exec playwright test e2e/tests/demo-public.spec.ts` 与全量 `pnpm exec playwright test`，当前 8 条用例全部通过，P3-A“关键链路最小回归集”已形成闭环。

**下次方向**
- 下一步进入 `P3A-8`，开始补 `App 管理页 + 页面管理页` 的基本 CRUD 操作回归，这会从“路由/登录/上下文”转到“真实业务页面交互”。
- 当前公开页用例验证的是最小 public route 注册链；如果后续要覆盖更多 `demo-app` 公开路由，可在同一测试文件内按页面类型继续追加，不需要改现有基线。
- P3-A 现在已具备 smoke、未登录重定向、登录失败、登录成功、App 切换、Logout、公开页这 8 条稳定回归，后续新增用例建议继续保持“单文件单主链”的结构，便于定位失败面。 

### 2026-04-12 前端收口 Phase P3-A 第七段：管理页真实后端加载回归（Phase Cleanup）

**本次改动**
- 新增 `frontend/e2e/tests/management-pages.real.spec.ts`，使用真实重启后的前后端服务验证 `应用管理` 与 `受管页面` 两页可正常加载，而不是继续走 API mock。
- 用例使用管理员账号 `admin / admin123456` 登录真实后端，依次访问 `/system/app` 与 `/system/page`，显式等待并断言关键 API `system/apps`、`system/apps/current`、`pages`、`pages/menu-options` 返回 `200`，同时确认页面标题与主体内容可见、内联错误提示未出现。
- 本轮同时重启了前后端开发服务：前端运行在 `http://127.0.0.1:5174`，后端运行在 `http://127.0.0.1:8080`；首轮真实页面访问触发了一次 Vite `Outdated Optimize Dep` 预构建抖动，二次执行后已稳定消失。
- 已执行 `frontend` 下带真实环境变量的 `pnpm exec playwright test e2e/tests/management-pages.real.spec.ts` 与全量 `pnpm exec playwright test`，当前 9 条用例全部通过，P3-A 已完成 8 条最小回归 + 1 条真实管理页加载回归。

**下次方向**
- 下一步可切到 `P3-B`，开始“前端目录归并 — 按领域重组”，此时回归网已经足够支撑结构移动后的快速验证。
- `management-pages.real.spec.ts` 目前覆盖的是“页面可渲染 + 关键 API 200”，还没有继续做保存/删除类真实写操作；如果后续需要更深联调，可在这个文件上按页面拆小场景继续加。
- 真实前端开发服首次命中新依赖时仍可能出现一次性 Vite 预构建 reload；这属于开发环境噪音，不影响当前业务页最终稳定加载，但如果要进一步降低噪音，可以考虑补预热或优化 `optimizeDeps`。 

### 2026-04-12 前端收口 Phase P4：历史循环依赖收口（Phase Cleanup）

**本次改动**
- 新增 `frontend/src/domains/navigation/constants.ts`、`runtime/router-instance.ts`、`runtime/guard-state.ts`、`runtime/reset-handlers.ts`，把首页默认值、router 实例、guard 状态和运行时重置接口从 `router/index.ts` / `router/guards/beforeEach.ts` 中抽离成 navigation 域运行时边界。
- `menu.ts`、`worktab.ts`、`utils/jump.ts`、`utils/route.ts`、`utils/worktab.ts`、`auth/store.ts`、`auth/flows/shared.ts`、`app-runtime/runtime/app-context.ts` 已改走这些新边界，不再直接静态依赖 `@/router` 或 `beforeEach`；`afterEach.ts` 也不再通过 `useCommon()` 反向串回菜单 store。
- 验证通过：`frontend` 下 `pnpm exec vue-tsc --noEmit`、`pnpm build` 均通过；`pnpm dlx madge --circular --extensions ts --ts-config ./tsconfig.json src/domains src/router src/store src/api src/utils src/hooks src/directives` 的循环链数从 42 条降到 20 条。
- 当前剩余循环已从原来的 `navigation -> router/index -> beforeEach` 大簇收缩为两类历史主链：`app-runtime/context <-> runtime/app-context/menu-space` 以及 `governance/_shared -> http/auth-session -> auth/store`，定位范围已明显变小。

**下次方向**
- 如果继续清理，优先处理 `app-runtime/context` 与 `runtime/app-context` 的双向引用，尝试把运行时切换逻辑继续下沉为纯函数或 handler 注册，进一步压缩剩余主链。
- `utils/storage/index`、`store/modules/setting -> hooks/core/useCeremony` 这类历史 barrel/工具链仍保留少量循环，适合另开小节点逐条收口，不建议和主运行时链再混改。
- 这轮只做依赖方向整理，没有追加新的 E2E；若后续继续解环，建议继续沿用现有 `Playwright + vue-tsc + build + madge` 四件套做回归闸门。 

### 2026-04-12 前端收口 Phase P4-4：剩余循环链清零（Phase Cleanup）

**本次改动**
- 新增 `frontend/src/utils/http/request-context.ts`、`frontend/src/domains/app-runtime/runtime/context-handlers.ts`、`frontend/src/domains/auth/runtime/session-handlers.ts`，把 HTTP 层对 auth/workspace/app-context store 的直接读取改成注册式 request-context，把 `app-context` 与 `auth/session` 的运行时能力改成 handler 注入，拆掉剩余 `auth/api-store-session` 与 `app-runtime/context` 主链。
- `frontend/src/api/v5/client.ts`、`frontend/src/utils/http/auth-session.ts`、`frontend/src/domains/auth/api.ts`、`frontend/src/domains/auth/store.ts`、`frontend/src/domains/app-runtime/context.ts`、`frontend/src/domains/app-runtime/runtime/app-context.ts`、`frontend/src/store/modules/workspace.ts`、`frontend/src/store/modules/collaboration-workspace.ts` 已全部切到新的 request-context / runtime-handler 边界，不再在 HTTP 与运行时主链中反向拉 store。
- 新增 `frontend/src/hooks/core/ceremony-shared.ts`，并将 `setting.ts` 改为直接使用纯节日计算函数；`useCeremony.ts` 改为直连 `mittBus.ts`，`setElementThemeColor()` 改为显式接收 `isDark`，顺手剪掉 `setting -> useCeremony` 与 `setting -> utils/ui/index -> colors` 两条历史循环。
- 验证通过：`frontend` 下 `pnpm exec vue-tsc --noEmit`、`pnpm build` 均通过；`pnpm dlx madge --circular --extensions ts --ts-config ./tsconfig.json src/domains src/router src/store src/api src/utils src/hooks src/directives` 结果已从上一轮 `20 circular dependencies` 降到 `0`。

**下次方向**
- 当前循环依赖已经清零，后续更有价值的方向应回到用户可见问题，优先处理已在任务树里的“菜单切换刷新感与缓存失效”链路，而不是继续做无收益的结构性扫尾。
- `vite build` 仍保留几条动态/静态混引 warning（`locales/index.ts`、`domains/auth/api.ts`、认证页动态组件），它们不再形成循环依赖，但如果要继续收口包体边界，可以单开“chunk 与动态导入一致性”小节点处理。
- 若下一轮继续动导航与缓存链，建议直接复用本轮留下的 `request-context` / `runtime-handler` 边界，不要再把 store 反向引回 HTTP、router 或通用工具层。 

### 2026-04-12 前端收口 Phase P2A-1：现状流图补齐（Phase Cleanup）

**本次改动**
- 重写 `docs/frontend-cleanup-p2a-notes.md`，把早期计划态流图更新为当前真实实现态，覆盖“页面刷新 / 登录完成 / App 切换 / URL 带 space_key 与缺路由补救”四条主链。
- 文档中补齐了当前 `session-runtime`、`app-context-runtime`、`navigation-runtime`、`beforeEach` 的职责归属，并显式标出 4 处仍存在的职责交叉点，避免后续继续沿用已经过时的 P2A 认知。
- 额外单列了 `beforeEach.ts` 中仍应继续下沉的业务逻辑，包括 centralized login 策略预热、session 预取、space_key 切换、缺路由补救与跨域 app 跳转目标构造。

**下次方向**
- 当前这条文档节点已补齐，若继续做项目清洁，下一步建议从 `vite build` 里剩余的动态/静态混引 warning 下手，单开一轮 chunk 边界与动态导入一致性收口。
- 运行时主链虽然已基本稳定，但 `session -> menu-space`、`navigation -> app-context/menu-space`、`app-context -> navigation clear/refresh` 三条跨层编排仍可继续压缩，适合在后续“运行时纯化”节点里做小步重构。
- 菜单刷新感问题已按用户反馈收口，后续若再出现同类现象，建议直接新开独立节点，不再复用当前 cleanup 任务里的历史记录。 
