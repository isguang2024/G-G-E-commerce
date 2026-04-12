# 节点 1.3 重复实现审查（frontend/src + backend/internal）

## 审查范围与结论
- 范围：`frontend/src`、`backend/internal`。
- 结论：发现 4 组“重复/近重复实现”候选，其中 2 组可立即合并（低风险、行为可保持不变），2 组建议暂不处理（当前承载兼容路由或显式业务分流语义）。

## 可立即合并/删除

### 1) 后端错误码常量与默认文案双份维护（可立即合并）
- 重复点：`legacyresp` 与 `apperr` 同时维护相同错误码（`2001/2002/2003/2004/2005/5001`）与同义文案。
- 证据：
  - `backend/internal/api/legacyresp/errors.go:12-17` 定义 `CodeUnauthorized/CodeTokenExpired/CodeForbidden/CodeAPIKeyMissing/CodeTokenBadFormat/CodeInternal`。
  - `backend/internal/api/legacyresp/errors.go:20-26` 定义默认文案。
  - `backend/internal/api/apperr/codes.go:24-28,54` 定义同一组错误码。
- 处理建议（立即可做）：
  - 将 `legacyresp` 改为引用 `apperr` 的错误码常量（必要时复用统一文案映射），消除双份真相源。
  - 该调整不改变外部响应结构，只减少维护面。

### 2) 两处 JWT 头解析与鉴权失败分支近重复（可立即合并）
- 重复点：`backend/internal/api/middleware/auth.go` 与 `backend/internal/modules/system/auth/middleware.go` 中 JWT 头读取、`Bearer` 校验、过期/无效分支响应逻辑基本一致。
- 证据：
  - `backend/internal/api/middleware/auth.go:19,23,29,37,39`
  - `backend/internal/modules/system/auth/middleware.go:19,23,29,38,40`
- 处理建议（立即可做）：
  - 抽取共享 helper（例如“解析并校验 Bearer Token”），两处中间件仅保留各自差异逻辑（如 `applyAuthorizationContext`）。
  - 可先以“无行为变化重构”方式落地，并加最小回归测试。

## 暂不处理

### 1) account-portal 认证页面壳层（暂不处理）
- 现状：`account-portal` 下 3 个页面仅作为 `views/auth/*` 的壳层转发。
- 证据：
  - `frontend/src/views/account-portal/auth/login/index.vue:2-3,7`
  - `frontend/src/views/account-portal/auth/register/index.vue:2-3,7`
  - `frontend/src/views/account-portal/auth/forget-password/index.vue:2-3,7`
- 暂不处理原因：
  - 文件内已有 `@compat-status: transition ...` 注释，明确其兼容过渡职责。
  - 直接删除可能影响既有路由名/外部入口映射，需先完成路由收敛与回归验证。

### 2) 个人空间/协作空间消息页壳层（暂不处理）
- 现状：两个页面都复用同一控制台组件，但通过 `scope` 区分业务语义。
- 证据：
  - `frontend/src/views/system/message/index.vue:2,6`（`scope="personal"`）
  - `frontend/src/views/collaboration-workspace/message/index.vue:2,6`（`scope="collaboration"`）
- 暂不处理原因：
  - 路由层语义与权限挂载通常按入口拆分；即使模板相似，直接合并为单一路由存在权限与导航回归风险。
  - 建议在完成“路由命名/权限点统一建模”后再统一壳层。

## 建议落地顺序（最小风险）
1. 先做后端错误码单源收敛（`legacyresp` → `apperr` 引用）。
2. 再做 JWT 公共解析 helper 抽取（仅重构，不改业务分支）。
3. 最后评估前端壳层收敛（以路由兼容窗口和权限回归通过为前置条件）。
