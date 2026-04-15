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
- `internal/`：handler、service、repository 与基础设施实现

## 不在这里重复写的内容

- API 固定闭环流程统一看 [../docs/API_OPENAPI_FIXED_FLOW.md](../docs/API_OPENAPI_FIXED_FLOW.md)
- OpenAPI 生成链和编辑规则统一看 [api/openapi/README.md](api/openapi/README.md)
