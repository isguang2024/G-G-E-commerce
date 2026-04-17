# backend/api/openapi

OpenAPI 契约源文件目录。

## 子目录

- `openapi.root.yaml` — 顶层入口
- `components/` — 共享 schema、错误响应、安全方案
- `domains/` — 按业务域拆分的 paths 与 schemas
- `dist/` — bundle 产物，禁止手改
- `paths/` — 历史兼容结构，不再新增

## 真相入口

编辑规则、生成链、`x-permission-key` 协同清单等开发真相见 [../../Truth/openapi-spec-rules.md](../../Truth/openapi-spec-rules.md)。
