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

## 前端模板与组件库

- **基座模板**：[Art Design Pro](https://www.artd.pro/)（Vue 3 + TypeScript + Vite + Element Plus + Tailwind CSS 的企业级中后台模板）。
  官方文档：<https://www.artd.pro/docs/zh/guide/introduce.html>。
- **UI 基础库**：Element Plus（弹层、表单、表格、抽屉等原子组件）+ Tailwind（工具类排版/间距）。
- **全局注册的基座组件**：`frontend/src/components/core/` 下所有 `Art` 前缀组件（`ArtTable`、`ArtTableHeader`、`ArtStatsCard`、图表家族等）通过 `utils/registerGlobalComponent.ts` 自动注册，页面直接写标签即可，**无需 import**。API 形状以基座官方文档为准。
- **本仓沉淀的业务组件**：`frontend/src/components/business/`，需要 `import`。代表性组件：
  - `FieldLabel`：表单标签 + 问号 tooltip，替代冗长 `form-tip`
  - `DictSelect`：字典下拉（filterable + 缓存 + 默认项），所有枚举字段统一走它，**禁止硬编码 `<ElOption>` 列表**
  - `PermissionActionWorkbench` / `PermissionSourcePanels` 等权限工作台组件
  - `JsonViewer` / `TraceDrawer` 等可观测性组件
- **配套 Hooks**：基座 `useTable`；本仓 `useDictionary` / `useDictionaries` / `useUpload`。
- **完整清单与约束**：[frontend/src/components/README.md](frontend/src/components/README.md)（含新增组件的边界与命名原则）。
