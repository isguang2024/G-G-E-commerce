# backend/truth.md

> 后端开发真相摘要。详细条目见 [truth_index.md](./truth_index.md) 和 [Truth/](./Truth/)。

## 核心铁律

- **OpenAPI-first**：spec 是唯一契约真相源，改接口先改 `api/openapi/`，再 bundle → ogen → gen-permissions → 前端 gen:api。
- **生成物禁止手改**：`api/gen/`、`internal/pkg/permissionseed/openapi_seed.json`、前端 `api/v5/` 全部只能通过生成链刷新。
- **Sub-handler 按 domain 拆分**：禁止回退到单一 god `APIHandler`；新 op 归入对应 `*{domain}APIHandler`。
- **Router 不手改**：`/api/v1/*` 由 OpenAPI seed 驱动自动挂载；仅 `/health`、`/uploads`、OAuth 回调、WS 等非 OpenAPI 入口手动注册。
- **权限判断只走 evaluator**：`internal/pkg/permission/evaluator` 是唯一判权入口，禁止在 sub-handler / service 散写。
- **Migration 优先落**：涉及 migration 的改动必须先建迁移再写实现，避免并行插队导致编号冲突。
- **默认数据走 seed/ensure**：长期默认状态不反复写进 migration；临时修复型迁移达成后必须删除。
- **仓储层带 tenant 过滤**：每条查询显式过滤 `tenant_id`（当前固定 `default`）。

## 权限键改动协同（四处同步）

改 `x-permission-key` 或 `x-access-mode` 必须同步：

1. `api/openapi/domains/{domain}/paths.yaml`
2. `internal/pkg/permissionkey/permissionkey.go` 的 legacy→canonical 映射
3. `internal/pkg/permissionseed/seeds.go` 的 `DefaultPermissionKeys` / feature package 绑定
4. `cmd/migrate/main.go` 的 `consolidatePermissionKeys`（合并 / 改名 / 删除场景）

## API 变更后必跑

1. `bundle` → `dist/openapi.yaml`
2. `ogen` → `api/gen/`
3. `gen-permissions` → 权限种子 + 前端错误码
4. 前端 `pnpm run gen:api` → `frontend/src/api/v5/`
5. `go test ./internal/api/handlers -count=1`
6. `pnpm exec vue-tsc --noEmit`

未完成以上步骤前不得判定"接口改造完成"。

## 核心语义提醒

- **Workspace** 是空间主实体，分 `personal` / `collaboration`；最终权限公式 = `空间功能包权限键 ∩ 成员角色权限键`
- **菜单 / 页面 / 权限键** 三段分离：菜单管导航，页面管路由，权限键管访问

## 入口

- AI 协作约束：[../AGENTS.md](../AGENTS.md)
- 详细真相索引：[truth_index.md](./truth_index.md)
- 全部真相文档：[Truth/](./Truth/)
