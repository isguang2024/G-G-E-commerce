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
