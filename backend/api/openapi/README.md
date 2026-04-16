# backend/api/openapi

这里是后端 API 契约真相源。`backend/api/gen/`、权限种子和前端 `frontend/src/api/v5/` 都从这里派生。

## 当前目录

| 路径 | 说明 |
| --- | --- |
| `openapi.root.yaml` | 顶层入口，只放基础信息、标签、安全定义和各域引用 |
| `components/` | 共享 schema、错误响应、安全方案 |
| `domains/` | 按业务域拆分的路径与局部 schema |
| `dist/openapi.yaml` | bundle 后产物，机器生成，禁止手改 |
| `embed.go` | 将 `dist/openapi.yaml` 嵌入后端二进制 |
| `redocly.yaml` | bundle / lint 规则 |
| `paths/` | 兼容历史结构保留，新增内容不要继续写到这里 |

## 编辑原则

- 改接口先改 `domains/{domain}/paths.yaml`
- 共享 schema 改 `components/`，单域 schema 改 `domains/{domain}/schemas.yaml`
- 不手改 `dist/openapi.yaml`
- 不手改 `backend/api/gen/`
- OpenAPI 变更完成后，必须继续刷新后端生成物、权限种子和前端 client

## 生成链

1. `make api-bundle` 生成 `dist/openapi.yaml`
2. `make api-gen` 刷新 `backend/api/gen/`
3. `make api-perms` 刷新权限种子与前端错误码
4. 前端执行 `pnpm run gen:api` 刷新 `frontend/src/api/v5/`

Windows 下可用 `update-openapi.bat` 做后端链路刷新，但前端 `gen:api` 仍需单独执行。

## 生成之后的完整 pipeline

`make api` 只是整条闭环的前半段。完整 15 阶段顺序（出自 `../../docs/API_OPENAPI_FIXED_FLOW.md`）：

```
① model/domain
② migration
③ seed/ensure
④ OpenAPI spec           ← 本目录
⑤ bundle
⑥ lint
⑦ ogen
⑧ gen-permissions        ← 产出 internal/pkg/permissionseed/openapi_seed.json
⑨ restart backend        ← 硬性检查点，不可跳
⑩ router/bridge check   ← 自动挂载 + go test ./internal/api/router
⑪ sub-handler/service    ← internal/api/handlers/{domain}.go
⑫ frontend gen:api
⑬ frontend API 封装
⑭ UI
⑮ build/test/browser verify
```

### 为什么 ⑨ restart backend 不可跳

- `openapi_seed.json` 在后端 `router` 初始化时**只读一次**，产出"路由 → 权限键 / access_mode"的进程级缓存
- 中间件 `openapiperm.go` 查这份缓存做拦截，不会每次请求回源
- 因此合并 / 重命名 / 删除权限键后，DB 对齐但**缓存仍是旧的**，接口一律 403
- 典型触发：跑完 `cmd/migrate` 的 `consolidatePermissionKeys`、或 `make api` 改了 `x-permission-key`
- 操作：`docker compose restart backend` 或重启 `go run ./cmd/server`

### 改 `x-permission-key` 的协同清单

只要你改了 `x-permission-key` 或 `x-access-mode`，以下四处必须同步：

1. `backend/api/openapi/domains/{domain}/paths.yaml`（本目录）
2. `backend/internal/pkg/permissionkey/permissionkey.go` 的 legacy→canonical 映射
3. `backend/internal/pkg/permissionseed/seeds.go` 的 `DefaultPermissionKeys` / feature package 绑定
4. `backend/cmd/migrate/main.go` 的 `consolidatePermissionKeys` 任务（合并 / 改名 / 删除场景）

跑完 `make api` + `cmd/migrate` 后**必须重启后端**，否则出现 403 先检查是否遗漏了重启。

## 阅读顺序

1. 先看 `../../docs/API_OPENAPI_FIXED_FLOW.md`
2. 再回到本目录改 spec
3. 最后对照 `internal/api/handlers/{domain}.go` 与前端调用收口实现（sub-handler 按 domain 拆分）
