# API / OpenAPI 固定闭环流程

本文用于固化本仓库“新增能力 / 新增 API / 修改 API 契约”的固定顺序。

如果只记一句话，按下面这条链执行：

`model/domain → migration → seed/ensure → OpenAPI spec → bundle → lint → ogen → gen-permissions → router/bridge check → sub-handler/service → frontend gen:api → frontend API 封装 → UI → build/test/browser verify`

## 1. 简版顺序

### 1.1 固定顺序

1. `model/domain`
2. `migration`
3. `seed/ensure`
4. `OpenAPI spec`
5. `bundle`
6. `lint`
7. `ogen`
8. `gen-permissions`
9. `router/bridge check`
10. `sub-handler/service`
11. `frontend gen:api`
12. `frontend API 封装`
13. `UI`
14. `build/test/browser verify`

### 1.2 一句话解释

- `model/domain`：先定领域模型、表结构、约束和仓储过滤边界。
- `migration`：只负责结构变更和一次性历史修正。
- `seed/ensure`：长期默认数据、注册表、幂等补齐逻辑放这里，不反复写进 migration。
- `OpenAPI spec`：后端与前端共享的唯一契约真相源。
- `bundle/lint/ogen`：把多文件 spec 收敛成最终契约并生成后端产物。
- `gen-permissions`：从 OpenAPI 扩展字段派生 `permission_keys / api_endpoints / bindings`。
- `router/bridge check`：**自动挂载 + 覆盖率核对**。seed 驱动的 `mountOpenAPIBridgeRoutes` 已经把每条 operation 挂到 Gin 分组；这一步只做核对（跑 `go test ./internal/api/router`），仅当引入非 OpenAPI 入口时才需要手动往 `router.go` 加行。
- `sub-handler/service`：按 domain 拆分到 `internal/api/handlers/{domain}.go` + `{domain}_handler.go`，基于最新生成物实现业务逻辑，不手改生成代码，也不回退到 god `APIHandler` 堆方法。
- `frontend gen:api`：刷新前端类型与 client 契约。
- `frontend API 封装`：在生成层之外写业务调用封装。
- `UI`：页面、弹窗、表单、列表、状态切换、错误提示联调。
- `build/test/browser verify`：联编、测试、浏览器验证全部通过后才算闭环。

## 2. 详细流程

## 2.1 model / domain

新增能力先回答这些问题：

- 是否新增表、字段、索引、唯一约束
- 是否需要 `tenant_id`
- 仓储层是否必须强制 tenant / workspace 过滤
- 默认数据是一次性修复，还是长期幂等数据

输出物通常包括：

- `backend/internal/modules/.../models`
- service / repository 领域结构

## 2.2 migration

适用场景：

- 新表
- 新字段
- 索引 / 约束调整
- 一次性历史数据修正

不适合放 migration 的内容：

- 长期默认角色
- 长期默认 app / menu space / page / endpoint 注册
- 反复需要 ensure 的幂等数据

这些应放 `seed / ensure`。

## 2.3 seed / ensure

适用场景：

- 默认权限键
- 默认 app / menu space / page / endpoint 注册
- 启动时可重复执行的幂等补齐

项目原则：

- migration 管结构
- seed/ensure 管长期默认状态

## 2.4 OpenAPI spec

只改这里：

- `backend/api/openapi/`

当前主入口：

- `backend/api/openapi/openapi.root.yaml`

新增 API 时至少补齐：

- path + method
- request params / request body
- response schema
- error schema
- `operationId`
- `x-permission-key`
- `x-access-mode`
- 必要时 app / workspace / tenant 相关扩展字段

原则：

- spec 即真相
- 不允许先写 handler 再补 spec

## 2.5 bundle / lint / ogen

当前仓库入口在：

- `backend/Makefile`
- `backend/update-openapi.bat`

常规顺序：

1. `bundle`
2. `lint`
3. `ogen`

含义：

- `bundle`：把拆分的 domains/components 汇总成最终 `dist/openapi.yaml`
- `lint`：在生成前先拦截契约错误
- `ogen`：刷新 `backend/api/gen/`

注意：

- `backend/api/gen/` 是生成产物，禁止手改

## 2.6 gen-permissions

这是本仓库比普通 OpenAPI 项目多出来的一环。

作用：

- 解析 OpenAPI 扩展字段
- 派生 `permission_keys`
- 派生 `api_endpoints`
- 派生 `permission_key_api_bindings`

所以新增受控 API 后，不能只跑 ogen，还要继续跑：

- `go run ./cmd/gen-permissions`

否则会出现：

- spec 已更新
- handler 已实现
- router 已自动桥接（seed 驱动）
- 但权限注册表和 API 注册表不同步

## 2.7 router / bridge check

本仓库的"API 网关配置"不是单独一套外部网关文件，而是：

- OpenAPI 扩展字段（`x-access-mode`、`x-permission-key`、`x-app-scope` 等）
- `router.go` 顶部注册的全局 middleware 分组（RequestID / Logger /
  Recovery / AppContext / DynamicAppSecurity / endpoint-status / JWTAuth /
  permission middleware）
- 启动期由 seed 驱动的自动挂载（`mountOpenAPIBridgeRoutes`）

**这一步默认自动，只做覆盖率核对**。跑完 `make api` 之后：

- `backend/internal/pkg/permissionseed/openapi_seed.json` 已经包含每条
  operation 的 `method / path / access_mode`
- 后端启动时，`backend/internal/api/router/router.go` 里的
  `mountOpenAPIBridgeRoutes` 按字母序遍历 seed，把每条 op 按 `access_mode`
  挂到对应的 Gin group（`public` → 不过 JWT；`authenticated` /
  `permission` → 过 JWT，`permission` 再过权限 middleware）
- `router.go` 不再保存每条业务路由的显式 `authenticated.GET(..., ogenBridge)`
  行；只有 `/health`、`/uploads`、OAuth 回调、WebSocket 等**非 OpenAPI** 路由仍然显式注册

因此新增 API 后要检查的不是"router.go 有没有加行"，而是：

- spec 的 `x-access-mode` 是否正确（public / authenticated / permission）
- 是否带上了 `x-permission-key`（`permission` 模式必填）
- 对账测试 `go test ./internal/api/router -count=1` 通过
  （该测试构造空 gin.Engine、跑一次 seed 驱动注册、对比 seed 和实际
  `engine.Routes()`；任何漏项、多项、错配、radix tree 冲突都会失败）

需要**手工改 `router.go`** 的只有这几种情况：

- 新增 `/health`、`/metrics` 之类的非契约探针
- 新增文件下载 / WebSocket / SSE 等不走 ogen 的端点
- 新增 OAuth / 回调等 public 直通路径

## 2.8 sub-handler / service

生成后按 domain 拆分写 sub-handler + service + repository + mapper / helper。核心纪律：

- 基于最新 `backend/api/gen/` 实现，不允许继续依赖旧字段名、旧响应结构、旧签名
- **禁止**把新 op 堆回 god `APIHandler`；新 op 必须归到某个域的 sub-handler
- 两个 sub-handler 出现同名方法 = 域边界错了，重新划分

目录约定、扩展新域的步骤、嵌入深度规则、测试构造姿势见 [`backend/CLAUDE.md`](../backend/CLAUDE.md)，不在本文重复。

## 2.9 frontend gen:api

后端契约变更后，必须继续执行：

- `pnpm run gen:api`

刷新产物主要在：

- `frontend/src/api/v5/schema.d.ts`

以及可能联动：

- `frontend/src/api/v5/client.ts`
- `frontend/src/api/v5/types.ts`
- `frontend/src/api/v5/error-codes.ts`

原则：

- `frontend/src/api/v5/` 为生成层，禁止手改

## 2.10 frontend API 封装

UI 不直接散用生成 client。

应该写在：

- `frontend/src/api/*.ts`
- `frontend/src/api/system-manage/*.ts`

这些封装负责：

- query/body 参数转换
- snake_case / camelCase 适配
- 列表响应归一化
- 业务态兜底

## 2.11 UI

最后才进入页面层：

- 列表
- 表单
- 弹窗
- 抽屉
- 提交态
- 错误提示
- 空态
- 权限可见性
- 路由入口

如果是多 app / 多 space 治理能力，还要联调：

- app-context
- menu-space
- runtime navigation
- 切换入口与 landing

## 2.12 build / test / browser verify

最少校验项：

后端：

- `go build ./...`
- `go test ./internal/api/handlers -count=1`

如涉及 router / bridge：

- `go test ./internal/api/router -count=1`

前端：

- `pnpm exec vue-tsc --noEmit`

如有 UI 变更：

- 浏览器实测成功链路
- 浏览器实测失败链路
- 至少验证一次登录态 / 未登录态 / 权限态中的相关分支

## 3. 当前仓库推荐执行命令

后端 OpenAPI 链：

1. `cd backend`
2. `make api-bundle`
3. `make api-lint`
4. `make api-gen`
5. `make api-perms`

前端类型链：

1. `cd frontend`
2. `pnpm run gen:api`

联编校验：

1. `cd backend && go test ./internal/api/handlers -count=1`
2. `cd backend && go build ./...`
3. `cd frontend && pnpm exec vue-tsc --noEmit`

如要用批处理快捷入口：

- `backend/update-openapi.bat`

但它不是完整闭环，跑完后仍应补：

- 前端 `pnpm run gen:api`
- 联编校验

## 4. 常见疏漏

最常见的漏项是：

- 只改 sub-handler，没改 OpenAPI
- 跑了 ogen，没跑前端 `gen:api`
- 改了受控 API，但没跑 `gen-permissions`，权限注册表和 API 注册表不同步
- 改了 spec 但忘记跑生成链（`make api`），旧 seed 还在 `openapi_seed.json` 里，启动时自动挂载的 Gin 路由集合还停留在旧版本
- 默认数据写进 migration，而不是 seed / ensure
- 把新 op 堆回 god `APIHandler`，绕过 sub-handler 拆分
- UI 直接写死类型，没有吃生成 schema
- 只过编译，没有做浏览器验证

## 5. 固定口令版

团队内部如果只保留一条短句，建议固定成：

`先定模型和数据，再定 OpenAPI；先刷新生成物和权限种子（router/bridge check 随 seed 在启动时自动挂载），再按 domain 写 sub-handler；最后刷新前端类型、补 API 封装、做 UI 和联调验证。`
