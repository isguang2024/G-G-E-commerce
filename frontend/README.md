# Frontend 目录说明（V5）

`frontend/` 是前端主工程目录，负责管理端页面、状态、路由与 API 调用集成。

## 目录入口

- `src/`：主源码目录（详见 `src/README.md`）
- `package.json`：前端依赖与脚本入口
- `vite.config.ts`：构建配置
- `tsconfig.json`：TypeScript 与路径别名配置

## 常用命令

- 安装依赖：`pnpm install`
- 本地开发：`pnpm dev`
- 类型检查：`pnpm exec vue-tsc --noEmit`
- 构建：`pnpm build`
- 刷新 API 生成物：`pnpm run gen:api`

## 维护约定

- `src/api/v5/` 属于生成产物，不手改生成文件本体。
- API 契约变更需先更新后端 OpenAPI，再执行 `pnpm run gen:api`。
- 页面与业务逻辑开发优先遵循 `FRONTEND_GUIDELINE.md`。
