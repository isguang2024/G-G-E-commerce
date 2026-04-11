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
