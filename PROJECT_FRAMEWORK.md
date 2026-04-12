# PROJECT_FRAMEWORK.md

## 项目主框架

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
  `model/domain → migration → seed/ensure → OpenAPI spec → bundle → lint → ogen → gen-permissions → gin bridge/router → handler/service → frontend gen:api → frontend API 封装 → UI → build/test/browser verify`
- 这条顺序的含义是：
  - 先定领域模型、表结构和默认数据策略，再谈 API
  - OpenAPI 是后端与前端共享的唯一契约真相源
  - `gin bridge/router` 属于仓库内 API 网关配置的一部分，负责把 OpenAPI 生成的 server 接到认证、权限、endpoint-status 等中间件分组
  - 前端不是只拿到 `schema.d.ts` 就结束，仍需补业务 API 封装与 UI 联调
  - 未完成生成、桥接、权限种子、联编和浏览器验证前，不视为闭环完成

## 当前非目标

- 不开放跨 tenant 能力、不暴露 tenant 管理界面、不做 schema 分片。
- 不引入消息队列、GraphQL、插件化、跨服务 RPC。
- 新增设计不得以"菜单反推权限 / mock 接口 / 手写权限规则"模式扩散。
- 不维护第二套后台前端、不重启第二个后端工程。
