# backend/

Go 后端工程。

## 目录

- `api/openapi/` — OpenAPI 契约真相源（spec 源文件）
- `api/gen/` — OpenAPI 生成产物（只读）
- `cmd/` — 服务、迁移、生成、诊断等可执行入口
- `internal/` — sub-handler、service、repository 与基础设施实现
- `Truth/` — 后端开发真相
- `truth.md` / `truth_index.md` — 真相摘要与索引

## 入口

- 开发真相入口：[truth.md](./truth.md)
- AI 协作约束：[../AGENTS.md](../AGENTS.md)
- 结构说明（非真相）：[../docs/backend/structure.md](../docs/backend/structure.md)
