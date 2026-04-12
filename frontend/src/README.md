# frontend/src 结构说明

本文档以当前 `frontend/src` 实际目录为准，说明职责边界、主要子目录与导入约定。

## 1. 总体职责

- `src/` 是前端应用主源码目录，负责：
- 应用启动与运行时装配（`main.ts`、`App.vue`）。
- 页面与路由组织（`views/`、`router/`）。
- 状态管理（`store/`）。
- 接口访问与业务封装（`api/`）。
- 领域能力沉淀（`domains/`、`composables/`、`hooks/`、`utils/`）。
- UI 资源与样式（`components/`、`assets/`、`directives/`）。

## 2. 主要子目录

- `api/`
- 接口层。包含 OpenAPI 生成产物目录 `api/v5/`，以及业务封装（如 `api/system-manage/` 与其他业务 API 文件）。
- 约束：`api/v5/` 视为生成产物，不手改生成文件本体。

- `assets/`
- 静态资源与样式（`images/`、`styles/`、`svg/`）。

- `components/`
- 可复用组件，按 `core/` 与 `business/` 分层。

- `composables/`
- Vue 组合式能力（当前包含 `auth-flow/`）。

- `config/`
- 前端运行配置与模块化配置入口。

- `directives/`
- 自定义指令能力，按 `core/` 与 `business/` 分层并统一导出。

- `domains/`
- 领域模块（当前包含 `auth`、`app-runtime`、`governance`、`navigation`）。
- 目标：承载跨页面、跨模块的业务能力与领域编排。

- `enums/`
- 枚举常量定义。

- `hooks/`
- 组合式 Hook，按 `core/` 与 `business/` 分层并统一导出。

- `locales/`
- i18n 语言资源与入口。

- `mock/`
- 历史/辅助 mock 目录（当前存在 `upgrade/`）。

- `plugins/`
- 第三方插件接入（如 `echarts`）及统一注册入口。

- `router/`
- 路由核心配置、守卫、路由表与运行时路由能力。

- `store/`
- Pinia 状态管理，按模块组织。

- `types/`
- 类型定义中心（`api/`、`common/`、`component/`、`config/`、`router/`、`store/`、`import/`）。

- `utils/`
- 通用工具能力（如 `http/`、`permission/`、`navigation/`、`storage/`、`auth/` 等）。

- `views/`
- 页面目录，按业务域拆分（如 `account-portal/`、`auth/`、`dashboard/`、`message/` 等）。

## 3. 导入约定（按 tsconfig 路径别名）

当前别名定义（来源：`frontend/tsconfig.json`）：

- `@/*` → `src/*`
- `@/domains/auth`、`@/domains/auth/*`
- `@/domains/app-runtime`、`@/domains/app-runtime/*`
- `@/domains/governance`、`@/domains/governance/*`
- `@views/*` → `src/views/*`
- `@imgs/*` → `src/assets/images/*`
- `@icons/*` → `src/assets/icons/*`
- `@utils/*` → `src/utils/*`
- `@stores/*` → `src/store/*`
- `@plugins/*` → `src/plugins/*`
- `@styles/*` → `src/assets/styles/*`

推荐导入规则：

- 跨目录导入优先使用别名路径，避免多级相对路径（如 `../../../`）。
- 领域模块优先从其入口导出导入（例如 `@/domains/auth`），减少深层内部路径耦合。
- 生成层与业务层分离：业务代码调用 `api/v5` 的 client/type 能力，但不直接修改生成产物。
- 页面层优先依赖 `api/`、`domains/`、`store/`、`composables/` 暴露的能力，避免在 `views/` 内堆叠底层实现细节。

## 4. 维护提示

- 目录新增/重命名后，请同步更新本文档与 `tsconfig` 路径别名。
- 若接口契约变更，先更新 OpenAPI 并刷新生成物，再修正调用层与页面层。
