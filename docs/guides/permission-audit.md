# 权限链路排查报告

**日期**：2026-04-08
**范围**：`功能权限 (permission_keys)` ↔ `API 注册 (api_endpoints)` 运行时鉴权链路
**设计目标**：鉴权 = `workspace 的 feature_packages ∩ member 的 role 派生 keys`，单点拦截在 `backend/internal/api/middleware/openapiperm.go`；handler 不做二次校验。

---

## 结论速览

| # | 检查项 | 状态 |
|---|---|---|
| 1 | `access_mode=permission` 的 operation 全部标注 `x-permission-key` | ✅ OK |
| 2 | middleware 中 `workspace_id` 解析的三级 fallback | ✅ OK |
| 3 | handler 中是否有遗留二次鉴权 / 绕过 evaluator | ✅ OK |
| 4 | legacy gin 路由残留 | ✅ OK（仅测试 fixture）|
| 5 | `role_disabled_actions` 在 owner/admin bypass 路径 | ⚠️ 设计如此（owner/admin 全量放行）|
| 6 | JWT-only 接口的 user 解析链路 | ✅ OK |
| 7 | 数据范围（行级过滤）覆盖度 | ✅ OK（系统级资源合理豁免）|
| 8 | 鉴权拒绝的审计日志 | ✅ OK |

**真 Gap：无。** 架构实现与设计一致。

---

## 详细证据

### 1. spec permission_key 完备性 — OK
- 178 个 `x-access-mode: permission` operation 全部配置了 `x-permission-key`
- `backend/cmd/gen-permissions/main.go` 硬性拒绝缺失场景

### 2. workspace_id 解析 — OK
- `backend/internal/api/middleware/openapiperm.go`
- 优先级：`auth_workspace_id` → path/query `workspace_id|id` → `uuid.Nil`（账号级 union）

### 3. handler 二次鉴权 — OK
- `backend/internal/api/handlers/` 内 0 条对 `feature_package_keys` / `role_feature_packages` 的直接引用
- 仅 `internal/modules/system/permission/service.go`（系统管理域）与 evaluator 使用这些表

### 4. legacy gin 路由残留 — OK
- 生产代码无残留，仅 `apiendpoint/service_test.go` 测试 fixture 使用 `router.GET`

### 5. owner/admin bypass vs role_disabled_actions — 设计如此
- `internal/pkg/permission/evaluator/evaluator.go` Resolve 分支：owner/admin 直接返回 workspace 全量 keys，不走 `role_disabled_actions` 过滤
- 属于"特权等级"常见设计，不视为 Gap

### 6. user 解析 — OK
- gin auth middleware 将 `user_id` 写入 gin context
- `router.go` 通过 `context.WithValue(ctx, handlers.CtxUserID, ...)` 注入 ogen handler
- ogen middleware `userIDFromCtx` 从 `req.Context` 取出

### 7. 数据范围过滤 — OK
- 系统级资源（`users`、`roles`、`api_endpoints`）无 `workspace_id` 列，是全局资源，无需过滤
- 协作空间相关列表均正确包含作用域 WHERE

### 8. 拒绝审计日志 — OK
- middleware 在 `Can() == false` 时以 Info 级别记录 `op / key / user / workspace`

---

## 延后优化项（非 Gap）

| 优先级 | 项 | 说明 |
|---|---|---|
| Low | owner/admin 是否受 `role_disabled_actions` 约束 | 业务明确要求时再做 |
| Low-Med | spec lint 进 CI | 当前仅 `make gen` 时拦截，可加 pre-commit hook |
| Low | 拒绝日志补 IP / 时间戳 | 已有结构化日志，增量补字段 |

---

## 与本次"删除装饰字段"任务的关系

排查完成后确认：
- 运行时真正生效的只有 `permission_key` + `workspace_id` + `access_mode` 三者
- `api_endpoints` 的 `app_scope / app_key / context_scope / feature_kind / source` 与 `permission_keys` 的 `context_type / feature_kind / app_key / allowed_workspace_types / module_code` 均为装饰字段，无任何运行时 enforcement
- 已启动清理（详见同批次提交）
