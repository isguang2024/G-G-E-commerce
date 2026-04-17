# Mabeng Admin

通用管理后台脚手架。后端 `Go + Gin + GORM + Postgres`，前端 `Vue 3 + TypeScript + Vite`，接口治理采用 OpenAPI-first。

## 真相源

- AI 协作约束：[AGENTS.md](AGENTS.md)
- 后端真相：[backend/truth.md](backend/truth.md) · [backend/Truth/](backend/Truth/)
- 前端真相：[frontend/truth.md](frontend/truth.md) · [frontend/Truth/](frontend/Truth/)
- 前端平台真相：[frontend-platform/truth.md](frontend-platform/truth.md) · [frontend-platform/Truth/](frontend-platform/Truth/)

## 项目说明（非真相）

- 文档说明入口：[docs/README.md](docs/README.md)
- 后端结构：[docs/backend/structure.md](docs/backend/structure.md)
- 前端结构：[docs/frontend/structure.md](docs/frontend/structure.md)
- 前端平台结构：[docs/frontend-platform/structure.md](docs/frontend-platform/structure.md)
- 变更日志：[docs/change-log.md](docs/change-log.md)
- 临时任务/记忆：[docs/tmp/](docs/tmp/)

## 快速开始

1. 在 `backend/` 启动基础服务：`docker-compose up -d`
2. 安装依赖：`backend/` 执行 `go mod download`，`frontend/` 执行 `pnpm install`
3. 初始化数据库：`backend/` 执行 `go run ./cmd/migrate`
4. 启动后端：`backend/` 执行 `go run ./cmd/server`
5. 启动前端：`frontend/` 执行 `pnpm dev`
