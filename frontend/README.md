# Frontend

管理端前端工程，基于 **Art Design Pro** 商业模板（Vue 3 + TypeScript + Vite + Element Plus + Tailwind CSS）。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 框架 | Vue 3 + TypeScript |
| 构建 | Vite 7 |
| UI 基座 | Art Design Pro（含 Element Plus ^2.11.2） |
| 样式 | Tailwind CSS 4 + SCSS |
| 状态管理 | Pinia + pinia-plugin-persistedstate |
| 路由 | Vue Router 4 |
| HTTP | openapi-fetch（类型安全，基于 OpenAPI 生成） |
| 图表 | ECharts 6 |
| 富文本 | WangEditor 5 |

## 快速命令

```bash
pnpm install          # 安装依赖
pnpm dev              # 本地开发（自动打开浏览器）
pnpm exec vue-tsc --noEmit   # 类型检查
pnpm build            # 构建生产包
pnpm run gen:api      # 根据 OpenAPI spec 重新生成类型文件
```

## 文档导航

| 文档 | 说明 |
| --- | --- |
| [truth.md](truth.md) | 前端协作铁律与边界 |
| [truth_index.md](truth_index.md) | 前端真相文档索引 |
| [src/README.md](src/README.md) | `src/` 目录职责速览 |
| [src/components/README.md](src/components/README.md) | **组件选用指南**：三层体系 + 选型规则（开发新页面必读） |
| [src/components/docs/art-core-components.md](src/components/docs/art-core-components.md) | Art Design Pro 全部 54 个 `core/` 组件目录 |
| [src/components/docs/business-components.md](src/components/docs/business-components.md) | 本仓自封装 `business/` 组件 + 配套 Hooks |
| [Truth/frontend-guideline.md](Truth/frontend-guideline.md) | 前端开发规范（命名、目录、API 调用约定） |
| [../docs/frontend/structure.md](../docs/frontend/structure.md) | 当前前端代码结构说明 |
| [../backend/Truth/api-openapi-flow.md](../backend/Truth/api-openapi-flow.md) | OpenAPI 接口生成完整流程 |

## 基座说明

`core/` 下所有 `Art` 前缀组件及 `useTable` hook 均来自 Art Design Pro，通过 `utils/registerGlobalComponent.ts` **全局自动注册**，页面直接 `<ArtTable/>` 即可，无需 `import`。

Element Plus 通过 `unplugin-element-plus` **按需自动导入**，`<ElButton>`、`<ElForm>` 等同样无需手动 `import`。
