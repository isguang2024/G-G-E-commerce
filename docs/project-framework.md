# PROJECT_FRAMEWORK.md

## 项目主框架

### 项目定位

- 项目定位：**通用 Admin 管理后台脚手架**。
- 主干职责：提供认证、权限、路由/页面注册、OpenAPI 契约治理、运行时上下文、多空间治理等通用底座。
- 业务落位原则：垂直业务能力以模块化方式接入，不反向污染底座抽象。

仓库由两条主线组成，**没有第二条并行主线**：

1. `backend/` —— Go 1.25 + Gin + GORM + Postgres + Redis + Elasticsearch + ogen + goose + OpenTelemetry
   - 唯一有效的后端工程
   - OpenAPI 生成链：`backend/api/openapi/` → `backend/api/gen/`
   - 运行与维护入口以 `backend/cmd/` 下的 `server`、`migrate`、`gen-permissions`、`init-admin`、`repair-workspaces` 等命令为准

2. `frontend/` —— Vue 3 + TypeScript + Vite + Element Plus + Pinia + openapi-fetch
   - 唯一有效的管理端前端工程
   - **已接真实接口**，不走 mock
   - OpenAPI 前端生成产物位于 `frontend/src/api/v5/`（`client.ts`、`types.ts`、`schema.d.ts`、`error-codes.ts`）
   - 业务 API 封装写在 `frontend/src/api/*.ts` 与 `frontend/src/api/system-manage/*.ts`，不直接改生成文件

## 核心语义

- **Tenant 租户**：数据隔离的最外层边界。当前仅内置 `default`，前端无感知。
- **Account 账号**：tenant 内的全局认证主体，不跨 tenant。
- **Workspace 空间（团队）**：tenant 内的权限与业务归属主体，分 `personal` / `collaboration` 两类。
- **Member 成员**：账号在 workspace 内的身份记录。
- **最终权限公式**：`空间已开通功能包权限键 ∩ 成员角色权限键`。tenant 不参与该公式，仅做外层数据隔离。
- **菜单 / 页面 / 权限键** 三段分离：菜单管导航，页面管路由，权限键管访问。`menu_space` 只是某 app 下的导航视图，不参与权限计算。

## 实施约束

- 所有后端改动默认在 `backend/` 内完成，不另起后端工程。
- 所有前端改动默认在 `frontend/` 内完成，不另起前端工程。
- 开发环境数据库允许清库重建；结构变更通过 goose 迁移落地，默认数据通过 seed / ensure 或 `cmd/gen-permissions` 维护。
- API 一律 OpenAPI-first，spec 在 `backend/api/openapi/`，生成物在 `backend/api/gen/` 与 `frontend/src/api/v5/`，业务封装不在生成目录中维护。
- 权限判断只走 `backend/internal/pkg/permission/evaluator`，不允许散写。
- 所有业务表必须带 `tenant_id`，仓储层强制过滤；唯一性约束写成 `(tenant_id, business_key)`。
- 缓存 key、日志、trace、审计事件必须携带 `tenant_id`。
- API 改动必须同步更新 OpenAPI 与前端生成 client。

## 新增能力固定闭环

- 项目统一按这条顺序推进新增能力或新增接口：
  `model/domain → migration → seed/ensure → OpenAPI spec → bundle → lint → ogen → gen-permissions → restart backend → router/bridge check → sub-handler/service → frontend gen:api → frontend API 封装 → UI → build/test/browser verify`
- 这条顺序的含义是：
  - 先定领域模型、表结构和默认数据策略，再谈 API
  - OpenAPI 是后端与前端共享的唯一契约真相源
  - `restart backend` 是**硬性检查点**：`openapi_seed.json` 在 `router` 初始化时只读一次，路由↔权限键映射是进程级缓存。合并 / 重命名 / 删除权限键后若不重启，旧映射仍生效，会出现"DB 已对齐但接口 403"的假权限故障
  - `router/bridge check` 是**自动步骤 + 覆盖率核对**：跑完 `make api`、并重启后端后，`backend/internal/api/router/router.go` 的 `mountOpenAPIBridgeRoutes` 已经把每条 operation 按 `x-access-mode` 挂到 Gin 的认证 / 权限 / endpoint-status 分组；除 `/health`、`/uploads`、OAuth 回调等**非 OpenAPI** 入口外，不要人工往 `router.go` 加行。核对靠 `go test ./internal/api/router -count=1`
  - `sub-handler/service` 指按 domain 拆分的 sub-handler（`internal/api/handlers/{domain}.go` + `{domain}_handler.go`），不再写单一 god `APIHandler`
  - 前端不是只拿到 `schema.d.ts` 就结束，仍需补业务 API 封装与 UI 联调
  - 未完成生成、权限种子、**后端重启**、联编和浏览器验证前，不视为闭环完成

## 当前非目标

- 不开放跨 tenant 能力、不暴露 tenant 管理界面、不做 schema 分片。
- 不引入消息队列、GraphQL、插件化、跨服务 RPC。
- 新增设计不得以"菜单反推权限 / mock 接口 / 手写权限规则"模式扩散。
- 不维护第二套后台前端、不重启第二个后端工程。
- 不将主干演化为单一行业（如电商）专属平台。
