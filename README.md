# Mabeng Admin

通用管理后台脚手架。后端 `Go + Gin + GORM + Postgres`，前端 `Vue 3 + TypeScript + Vite`，接口治理采用 OpenAPI-first。

若文档说明冲突，以以下真相源为准：

- [AGENTS.md](AGENTS.md)
- [docs/project-framework.md](docs/project-framework.md)
- [docs/frontend-guideline.md](docs/frontend-guideline.md)
- [backend/CLAUDE.md](backend/CLAUDE.md)

## 先看什么

1. [docs/INDEX.md](docs/INDEX.md)
2. [docs/project-framework.md](docs/project-framework.md)
3. [docs/project-structure.md](docs/project-structure.md)
4. [backend/api/openapi/README.md](backend/api/openapi/README.md)

## 快速开始

1. 在 `backend/` 启动基础服务：`docker-compose up -d`
2. 安装依赖：`backend/` 执行 `go mod download`，`frontend/` 执行 `pnpm install`
3. 初始化数据库：`backend/` 执行 `go run ./cmd/migrate`
4. 启动后端：`backend/` 执行 `go run ./cmd/server`
5. 启动前端：`frontend/` 执行 `pnpm dev`

## 常用入口

- 仓库导航：[docs/INDEX.md](docs/INDEX.md)
- 代码结构：[docs/project-structure.md](docs/project-structure.md)
- API 闭环流程：[docs/API_OPENAPI_FIXED_FLOW.md](docs/API_OPENAPI_FIXED_FLOW.md)
- 项目边界：[docs/project-framework.md](docs/project-framework.md)
- 前端规范：[docs/frontend-guideline.md](docs/frontend-guideline.md)
- 专题手册：[docs/guides/README.md](docs/guides/README.md)
