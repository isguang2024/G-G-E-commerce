# G-G-E-commerce

GGE 5.0 基线已经落地。这个仓库当前以“增量开发 + 历史收口”为主，所有新工作都应先定位到正确的真相源，再进入实现层。

若文档说明出现冲突，以 `AGENTS.md` 定义的协作文档真相源为准。

## 先读什么

1. [`docs/V5_REFACTOR_TASKS.md`](docs/V5_REFACTOR_TASKS.md) - 当前任务总入口和阶段进度
2. [`backend/api/openapi/README.md`](backend/api/openapi/README.md) - 后端 API 契约与生成链路的唯一说明
3. [`frontend/`](frontend/) - 前端主工程，已接真实接口
4. [`docs/README.md`](docs/README.md) - 文档中枢总入口
5. [`docs/GUIDELINES.md`](docs/GUIDELINES.md) - 文档与协作写作规则

## 目录分工

| 路径 | 角色 |
| --- | --- |
| `backend/` | 后端实现、领域、接口生成与服务命令 |
| `frontend/` | Vue 3 + TypeScript 管理端主工程 |
| `docs/` | 当前结构说明、任务入口、专题手册与导航中枢 |
| `Instructions/` | 仅放 agent 执行型说明、协作约束和工作流提示 |

## 快速开始

1. 安装依赖
   - 前端：在 `frontend/` 执行 `pnpm install`
   - 后端：在 `backend/` 执行 `go mod download`
2. 启动后端（默认 `:8080`）
   - 在 `backend/` 执行 `go run ./cmd/server`
3. 启动前端（默认 `:5174`）
   - 在 `frontend/` 执行 `pnpm dev`
4. API 契约变更链路
   - 先更新 `backend/api/openapi/`，再执行 `bundle -> ogen -> pnpm run gen:api`

## 当前工作方式

- API 一律 OpenAPI-first，先改 `backend/api/openapi/`，再刷新生成物，最后改实现与前端调用。
- 生成产物不手改，重点关注 `backend/api/gen/` 与 `frontend/src/api/v5/`。
- 文档导航优先指向当前有效入口，不重复堆历史背景。

## 文档入口

- [`docs/INDEX.md`](docs/INDEX.md) - 文档索引总表
- [`docs/guides/README.md`](docs/guides/README.md) - 专题手册入口
- [`PROJECT_FRAMEWORK.md`](PROJECT_FRAMEWORK.md) - 仓库主框架和边界
- [`FRONTEND_GUIDELINE.md`](FRONTEND_GUIDELINE.md) - 前端实现规范

## 文档导航（阶段 5 后）

- 文档中枢：[`docs/INDEX.md`](docs/INDEX.md)
- 任务与进度：[`docs/V5_REFACTOR_TASKS.md`](docs/V5_REFACTOR_TASKS.md) 与 [`docs/reports/`](docs/reports/)
- 功能/流程说明：[`/.claude/Instructions/`](.claude/Instructions/)  
  当前按 `features/` 与 `flows/` 分组维护。
- 文档规范：[`docs/GUIDELINES.md`](docs/GUIDELINES.md)

## 任务入口

如果你要继续推进 V5 重构、接口闭环、权限和数据结构相关工作，先看：

- [`docs/V5_REFACTOR_TASKS.md`](docs/V5_REFACTOR_TASKS.md)
- [`docs/API_OPENAPI_FIXED_FLOW.md`](docs/API_OPENAPI_FIXED_FLOW.md)
- [`docs/project-structure.md`](docs/project-structure.md)
