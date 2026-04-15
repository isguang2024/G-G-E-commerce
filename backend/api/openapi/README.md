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

## 阅读顺序

1. 先看 `../../docs/API_OPENAPI_FIXED_FLOW.md`
2. 再回到本目录改 spec
3. 最后对照 `internal/api/handlers/` 与前端调用收口实现
