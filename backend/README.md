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

## 关键约定（一句话各）

- **/api/v1 路由 seed 驱动自动挂载**，`router.go` 不手改，细节见 [CLAUDE.md](CLAUDE.md#3-implement-the-operation-sub-handler--service)
- **Sub-handler 按 domain 拆分**，禁止堆回 god `APIHandler`，细节见 [CLAUDE.md](CLAUDE.md#3-implement-the-operation-sub-handler--service)
- **权限键合并 / 改名协同四处同步**（spec / permissionkey.go / seeds.go / cmd/migrate），细节见 [api/openapi/README.md](api/openapi/README.md#%E6%94%B9-x-permission-key-%E7%9A%84%E5%8D%8F%E5%90%8C%E6%B8%85%E5%8D%95)

## 不在这里重复写的内容

- API 固定闭环流程 → [../docs/API_OPENAPI_FIXED_FLOW.md](../docs/API_OPENAPI_FIXED_FLOW.md)
- OpenAPI 生成链和编辑规则 → [api/openapi/README.md](api/openapi/README.md)
- Sub-handler 拆分、深度规则、扩展步骤 → [CLAUDE.md](CLAUDE.md)
