# Frontend 入口

`frontend/` 是唯一有效的管理端前端工程目录。

## 先看什么

- 前端实现规范：[../docs/frontend-guideline.md](../docs/frontend-guideline.md)
- `src/` 目录说明：[src/README.md](src/README.md)
- 项目结构：[../docs/project-structure.md](../docs/project-structure.md)
- API 流程：[../docs/API_OPENAPI_FIXED_FLOW.md](../docs/API_OPENAPI_FIXED_FLOW.md)

## 常用命令

- 安装依赖：`pnpm install`
- 本地开发：`pnpm dev`
- 类型检查：`pnpm exec vue-tsc --noEmit`
- 构建：`pnpm build`
- 刷新接口生成物：`pnpm run gen:api`
