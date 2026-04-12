# PROJECT_STRUCTURE

## 顶层结构

```text
G-G-E-commerce/
├─ backend/          # Go 后端工程
├─ frontend/         # Vue 3 + TypeScript 前端工程
├─ docs/             # 文档中枢
├─ .claude/          # Claude Code 配置与指令
├─ AGENTS.md         # 协作约束（真相源）
├─ PROJECT_FRAMEWORK.md  # 项目架构与边界（真相源）
├─ FRONTEND_GUIDELINE.md # 前端实现规范（真相源）
├─ README.md         # 仓库入口
├─ start-backend.bat # Windows 后端启动脚本
└─ start-frontend.bat# Windows 前端启动脚本
```

## 核心子结构

### 前端 `frontend/src/`

| 目录 | 职责 |
| --- | --- |
| `api/` | API 层：`v5/` 生成产物（只读）+ 业务封装 |
| `assets/` | 静态资源（图片、样式、图标） |
| `components/` | 可复用组件：`core/` 通用 + `business/` 业务 |
| `composables/` | Vue 组合式函数 |
| `config/` | 前端配置 |
| `directives/` | 自定义指令 |
| `domains/` | 领域模块（auth、app-runtime、governance、navigation） |
| `enums/` | 枚举常量 |
| `hooks/` | 组合式逻辑：`core/` + `business/` |
| `locales/` | 国际化资源 |
| `plugins/` | 第三方插件集成 |
| `router/` | 路由配置、守卫、动态路由 |
| `store/` | Pinia 状态管理 |
| `types/` | TypeScript 类型定义 |
| `utils/` | 工具函数 |
| `views/` | 页面组件（按业务域组织） |

### 后端 `backend/`

| 目录 | 职责 |
| --- | --- |
| `api/openapi/` | OpenAPI 契约源文件（真相源） |
| `api/gen/` | ogen 生成代码（只读） |
| `cmd/` | 命令入口（server、migrate、gen-permissions、init-admin 等） |
| `internal/api/handlers/` | HTTP handler 实现 |
| `internal/api/middleware/` | 中间件（认证、权限、日志） |
| `internal/api/router/` | ogen → gin 路由桥接 |
| `internal/api/dto/` | 数据传输对象 |
| `internal/api/mapper/` | DTO ↔ Domain 映射 |
| `internal/api/apperr/` | 应用错误定义 |
| `internal/modules/system/` | 业务领域模块 |
| `internal/modules/system/models/` | GORM 数据模型 |
| `internal/pkg/` | 共享基础设施包 |
| `internal/config/` | 配置加载 |

### 文档 `docs/`

| 目录/文件 | 职责 |
| --- | --- |
| `INDEX.md` | 文档总索引 |
| `README.md` | 文档中枢入口 |
| `GUIDELINES.md` | 文档写作规范 |
| `API_OPENAPI_FIXED_FLOW.md` | API 契约闭环流程 |
| `project-structure.md` | 代码结构与模块分工 |
| `guides/` | 专题开发手册 |
| `reports/` | 历史审计记录 |

## 指标快照

- TypeScript/Vue 文件数（`.ts/.tsx/.vue`）：~2936
- Go 文件数：~1450
- Markdown 文档数：~154

## 真相源文档

- `AGENTS.md` — 协作约束
- `PROJECT_FRAMEWORK.md` — 项目架构与边界
- `FRONTEND_GUIDELINE.md` — 前端实现规范
- `backend/CLAUDE.md` — 后端开发指引
- `docs/project-structure.md` — 代码结构与模块分工
