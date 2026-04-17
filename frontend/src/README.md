# frontend/src 结构说明

`src/` 是前端主源码目录。这里只解释目录职责，不重复前端规范或 API 流程。

## 核心目录

| 目录 | 职责 |
| --- | --- |
| `api/` | 业务 API 封装与 `api/v5/` 生成层 |
| `components/` | 复用组件，分 `core/`（全局注册）与 `business/`（按域显式 import）。清单见 [components/README.md](components/README.md) |
| `domains/` | 跨页面的领域能力编排 |
| `router/` | 路由表、守卫、动态路由运行时 |
| `store/` | Pinia 状态管理 |
| `utils/` | 工具函数与基础能力 |
| `views/` | 页面实现 |
| `assets/` / `locales/` / `types/` | 资源、文案、类型 |

## 导入约定

- 优先使用 `@/*` 等别名，不写多层相对路径。
- 页面层优先依赖 `api/`、`domains/`、`store/` 暴露的能力，不直接堆底层实现。
- `api/v5/` 仅作为生成产物输入，业务代码不要改生成文件本体。

## 延伸阅读

- 组件选用指南：[components/README.md](components/README.md)
- 页面规范：[../../docs/frontend-guideline.md](../../docs/frontend-guideline.md)
- 仓库结构：[../../docs/project-structure.md](../../docs/project-structure.md)
