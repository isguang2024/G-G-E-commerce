# PROJECT_FRAMEWORK.md

## 项目主框架

仓库由两条主线组成，**没有第二条并行主线**：

1. `backend/` —— Go 1.22 + Gin + GORM + Postgres + Redis + ES
   - 唯一有效的后端工程
   - 5.0 重构期会引入：ogen（OpenAPI-first）、casbin、wire、goose、cockroachdb/errors、gocache、testcontainers
   - 模块边界、术语、表结构以《GGE_5.0_初始化架构文档.docx》与 `docs/V5_REFACTOR_TASKS.md` 为准

2. `frontend/` —— Vue 3 + TypeScript + Vite + Element Plus + Pinia
   - 唯一有效的管理端前端工程
   - **已接真实接口**，不再走 mock；后端契约变更必须同步前端生成 client

## 核心语义（5.0 基线）

- **Tenant 租户**：数据隔离的最外层边界。当前仅内置 `default`，前端无感知。
- **Account 账号**：tenant 内的全局认证主体，不跨 tenant。
- **Workspace 空间（团队）**：tenant 内的权限与业务归属主体，分 `personal` / `collaboration` 两类。
- **Member 成员**：账号在 workspace 内的身份记录。
- **最终权限公式**：`空间已开通功能包权限键 ∩ 成员角色权限键`。tenant 不参与该公式，仅做外层数据隔离。
- **菜单 / 页面 / 权限键** 三段分离：菜单管导航，页面管路由，权限键管访问。`menu_space` 只是某 app 下的导航视图，不参与权限计算。

## 实施约束

- 所有后端改动默认在 `backend/` 内完成，不另起后端工程。
- 所有前端改动默认在 `frontend/` 内完成，不另起前端工程。
- 数据库可清库重建；表结构一次性 baseline 迁移落地。
- API 一律 OpenAPI-first，spec 在 `backend/api/openapi/`，handler 由 ogen 生成。
- 权限判断只走 `pkg/permission/evaluator`，不允许散写。
- 所有业务表必须带 `tenant_id`，仓储层强制过滤；唯一性约束写成 `(tenant_id, business_key)`。
- 缓存 key、日志、trace、审计事件必须携带 `tenant_id`。
- 大改动收尾时同步更新 `docs/V5_REFACTOR_TASKS.md` 中的阶段进度。

## 当前非目标

- 不开放跨 tenant 能力、不暴露 tenant 管理界面、不做 schema 分片。
- 不引入消息队列、GraphQL、插件化、跨服务 RPC。
- 不再兼容 4.5 旧术语（`collaboration_workspace` 独立持久化、菜单反推权限、`inherit_permission` 等），统一硬切。
- 不维护第二套后台前端、不重启第二个后端工程。
