# Workspace Phase 3-4 Implementation Plan

> **Execution note:** User already chose inline execution in this session. Continue in controlled sequence without waiting for extra confirmation.

**Goal:** 把运行时判权从单纯 `tenant_id` 分流，推进到 `auth workspace` 驱动，并补齐 `permission_key` 的 workspace 语义元数据。

**Architecture:** 保留现有平台快照和团队边界快照两套计算器，不重写快照层；在 authorization 入口新增 `AuthorizationContext`，按 `auth_workspace_type=personal|team` 选择平台或团队判权路径。权限定义层继续兼容 `context_type`，但新增 `app_key`、`data_policy`、`allowed_workspace_types` 作为 workspace 主语义元数据，并在 create/update/list 时自动推导和透出。

**Tech Stack:** Go, Gin, GORM, PostgreSQL AutoMigrate, Vue TypeScript typings

---

### Task 1: 统一 authorization workspace 上下文

**Files:**
- Create: `backend/internal/pkg/authorization/context.go`
- Modify: `backend/internal/pkg/authorization/authorization.go`
- Modify: `backend/internal/modules/system/auth/middleware.go`
- Modify: `backend/internal/modules/system/auth/handler.go`

- [ ] **Implement**

新增统一 `AuthorizationContext` 读取函数，集中解析：
- `user_id`
- `auth_workspace_id`
- `auth_workspace_type`
- 兼容 `tenant_id`
- `app_key`
- 可选 `target_workspace_id`

`RequireAction` / `RequireAnyAction` 改为先读 `AuthorizationContext`，再按 workspace 类型分流：
- `personal` 只走平台权限快照
- `team` 只走团队边界快照
- 继续兼容 `tenant_id` 作为 team workspace 的 legacy 投影

- [ ] **Verify**

Run: `go test ./internal/modules/system/auth ./internal/pkg/authorization`
Expected: `ok` 或 `[no test files]`

- [ ] **Notes**

- 这一任务不重构平台快照或团队边界快照内部算法，只改入口分流。

### Task 2: 给 permission_key 增加 workspace 元数据

**Files:**
- Modify: `backend/internal/modules/system/models/model.go`
- Modify: `backend/internal/modules/system/permission/service.go`
- Modify: `backend/internal/modules/system/permission/handler.go`
- Modify: `backend/internal/modules/system/user/handler.go`

- [ ] **Implement**

给 `PermissionKey` 新增并自动维护：
- `app_key`
- `data_policy`
- `allowed_workspace_types`

在 `permission` create/update 里根据 `permission_key`、`module_code`、`context_type` 自动推导默认值；list/detail/诊断响应把这些字段带出去。

- [ ] **Verify**

Run: `go test ./internal/modules/system/permission ./internal/modules/system/user`
Expected: `ok` 或 `[no test files]`

- [ ] **Notes**

- 本任务不要求前端页面已经使用这些字段，但契约必须先稳定下来。

### Task 3: 用 permission 元数据约束 workspace 类型

**Files:**
- Modify: `backend/internal/pkg/authorization/authorization.go`
- Modify: `backend/internal/modules/system/auth/handler.go`

- [ ] **Implement**

在判权入口增加 workspace 类型约束：
- `platform` / `tenant` / `system` 类权限默认仅 personal workspace 可用
- `team` 类权限默认仅 team workspace 可用
- `common` 类权限允许 personal/team

用户信息返回的 `actions` 快照改为优先基于当前 `auth workspace` 计算，而不是只看 `current_tenant_id`。

- [ ] **Verify**

Run: `go test ./internal/modules/system/auth ./internal/pkg/authorization ./internal/modules/system/permission`
Expected: `ok` 或 `[no test files]`

- [ ] **Notes**

- 这一步先做 workspace 类型约束，不在本轮把所有 `data_policy` 强制执行到每个接口。

### Task 4: 补前端契约字段并收口记录

**Files:**
- Modify: `frontend/src/types/api/api.d.ts`
- Modify: `docs/workspace-permission-stage-log.md`
- Modify: `docs/workspace-permission-migration.md`
- Modify: `docs/change-log.md`

- [ ] **Implement**

前端权限动作类型增加 permission workspace 元数据字段；文档记录 Phase 3-4 当前落点、验证命令和兼容保留项。

- [ ] **Verify**

Run: `pnpm build`
Expected: `built` 成功

- [ ] **Notes**

- 前端本轮只补契约，不改页面交互和 store 切换流程。
