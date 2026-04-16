# Backend 入口

`backend/` 是唯一有效的后端工程目录。

## 看哪份文档

- 改 API 契约：[api/openapi/README.md](api/openapi/README.md)
- 看后端协作约束：[CLAUDE.md](CLAUDE.md)
- 查常用命令：[../docs/guides/commands.md](../docs/guides/commands.md)
- 查数据库与 seed：[../docs/guides/database.md](../docs/guides/database.md)
- 看整体结构：[../docs/project-structure.md](../docs/project-structure.md)

## 目录职责

- `api/openapi/`：OpenAPI 真相源
- `api/gen/`：生成代码，只读
- `cmd/`：服务、迁移、生成、诊断入口
- `internal/`：sub-handler、service、repository 与基础设施实现

## 入口约定

- **API 路由由 OpenAPI seed 自动挂载**：`/api/v1` 下的所有 operation 在
  后端启动时由 `internal/api/router/router.go` 的 `mountOpenAPIBridgeRoutes`
  按 `permissionseed.LoadOpenAPISeed()` 派生挂载，**不需要手动往 `router.go` 加行**。
- **手动注册仅限非 OpenAPI 入口**：`/health`、`/metrics`、`/uploads`、
  OAuth 回调、WebSocket、SSE 等契约外的路径才留在 `router.go` 里显式注册。
- **Sub-handler 按 domain 拆分**：`internal/api/handlers/` 不再有 god handler。
  每个域自己持有：
  - `{domain}_handler.go` — `*{domain}APIHandler` struct + 构造函数
  - `{domain}.go` — 所有 receiver 为 `*{domain}APIHandler` 的 op 方法
  - `workspace.go` 的 `APIHandler` 只做嵌入 + `NewAPIHandler` 组装
  禁止把新 op 堆回 `APIHandler` 本体；两个 sub-handler 出现同名方法
  即意味着域边界划错，需重新拆分。

## 权限键合并 / 改名 SOP

以 `cmd/migrate` 里已沉淀的 `consolidatePermissionKeys` 为范例，改动权限键要同步：

1. `api/openapi/domains/{domain}/paths.yaml` 更新 `x-permission-key`
2. `internal/pkg/permissionkey/permissionkey.go` 增加 legacy→canonical 映射
3. `internal/pkg/permissionseed/seeds.go` 调整 `DefaultPermissionKeys` 与 feature package 绑定
4. `cmd/migrate/main.go` 补写 consolidation / prune 运行时任务
5. 执行 `make api`（bundle → lint → ogen → gen-permissions）刷新 `openapi_seed.json`
6. 执行 `cmd/migrate` 把 DB 里的遗留引用 rebind 到新键并软删旧键
7. **重启后端**，让路由↔权限键映射重新加载；不重启会出现"DB 已对齐但接口 403"的假权限故障
8. 浏览器回归：覆盖 admin / 普通角色两条链路

## 不在这里重复写的内容

- API 固定闭环流程统一看 [../docs/API_OPENAPI_FIXED_FLOW.md](../docs/API_OPENAPI_FIXED_FLOW.md)
- OpenAPI 生成链和编辑规则统一看 [api/openapi/README.md](api/openapi/README.md)
- Sub-handler 拆分细节（目录结构、深度规则、扩展步骤）看 [CLAUDE.md](CLAUDE.md) 与根目录 [../AGENTS.md](../AGENTS.md)
