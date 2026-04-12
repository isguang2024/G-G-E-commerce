# API / OpenAPI 固定闭环流程

本文用于固化本仓库“新增能力 / 新增 API / 修改 API 契约”的固定顺序。

如果只记一句话，按下面这条链执行：

`model/domain → migration → seed/ensure → OpenAPI spec → bundle → lint → ogen → gen-permissions → gin bridge/router → handler/service → frontend gen:api → frontend API 封装 → UI → build/test/browser verify`

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
9. `gin bridge/router`
10. `handler/service`
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
- `gin bridge/router`：把 OpenAPI 生成的 server 接入认证、权限、endpoint-status 等中间件分组。
- `handler/service`：基于最新生成物实现业务逻辑，不手改生成代码。
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
- router 已挂桥
- 但权限注册表和 API 注册表不同步

## 2.7 gin bridge / router

本仓库的“API 网关配置”不是单独一套外部网关文件，而是：

- OpenAPI 扩展字段
- `router.go` 里的 `ogenBridge`
- middleware 分组

新增 API 后必须检查：

- 是否挂到了正确的 gin 分组
- 是否走 authenticated / public / api-key 路径
- 是否经过 permission evaluator
- 是否经过 endpoint status / recovery / auth middleware

也就是说，新增 API 不是只有 spec 和 handler，还必须确认：

- `router.go` 真正把它接进来了

## 2.8 handler / service

生成后再写：

- handler
- service
- repository
- mapper / helper

原则：

- 基于最新 `backend/api/gen/` 实现
- 不允许继续依赖旧字段名、旧响应结构、旧签名

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

- 只改 handler，没改 OpenAPI
- 跑了 ogen，没跑前端 `gen:api`
- 改了受控 API，但没跑 `gen-permissions`
- handler 写完了，`router.go` 没挂 `ogenBridge`
- 默认数据写进 migration，而不是 seed / ensure
- UI 直接写死类型，没有吃生成 schema
- 只过编译，没有做浏览器验证

## 5. 固定口令版

团队内部如果只保留一条短句，建议固定成：

`先定模型和数据，再定 OpenAPI；先刷新生成物和权限种子，再接 gin bridge 和 handler；最后刷新前端类型、补 API 封装、做 UI 和联调验证。`
