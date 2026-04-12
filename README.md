# G-G-E-commerce

GGE 多租户 SaaS 管理平台。后端 Go + 前端 Vue 3，采用 OpenAPI-first 契约驱动开发。

若文档说明出现冲突，以 `AGENTS.md` 定义的协作文档真相源为准。

## 先读什么

1. [`PROJECT_FRAMEWORK.md`](PROJECT_FRAMEWORK.md) - 项目架构、核心语义与实施约束
2. [`AGENTS.md`](AGENTS.md) - 协作约束与 API 变更固定步骤
3. [`backend/api/openapi/README.md`](backend/api/openapi/README.md) - 后端 API 契约与生成链路
4. [`FRONTEND_GUIDELINE.md`](FRONTEND_GUIDELINE.md) - 前端实现规范
5. [`docs/INDEX.md`](docs/INDEX.md) - 文档索引总表

## 目录分工

| 路径 | 角色 |
| --- | --- |
| `backend/` | Go 后端：领域模块、API 生成、服务命令、数据库迁移 |
| `frontend/` | Vue 3 + TypeScript 管理端主工程 |
| `docs/` | 文档中枢：结构说明、专题手册、导航索引 |

## 快速开始

1. 启动基础服务（PostgreSQL + Redis）
   - 在 `backend/` 执行 `docker-compose up -d`
2. 安装依赖
   - 前端：在 `frontend/` 执行 `pnpm install`
   - 后端：在 `backend/` 执行 `go mod download`
3. 数据库初始化
   - 在 `backend/` 执行 `go run ./cmd/migrate`
4. 启动后端（默认 `:8080`）
   - 在 `backend/` 执行 `go run ./cmd/server`
5. 启动前端（默认 `:5174`）
   - 在 `frontend/` 执行 `pnpm dev`

## 开发工作流

- API 一律 OpenAPI-first，先改 `backend/api/openapi/`，再刷新生成物，最后改实现与前端调用。
- 生成产物不手改，重点关注 `backend/api/gen/` 与 `frontend/src/api/v5/`。
- 新增能力按固定闭环推进，详见 [`AGENTS.md`](AGENTS.md) 中的"API 变更固定步骤"。

## 文档入口

- [`docs/INDEX.md`](docs/INDEX.md) - 文档索引总表
- [`docs/guides/README.md`](docs/guides/README.md) - 专题手册入口
- [`docs/project-structure.md`](docs/project-structure.md) - 当前代码结构和模块分工
- [`docs/API_OPENAPI_FIXED_FLOW.md`](docs/API_OPENAPI_FIXED_FLOW.md) - API 契约闭环流程
