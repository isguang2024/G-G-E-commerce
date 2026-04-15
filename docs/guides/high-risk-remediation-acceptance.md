# 高风险页面深测整改落地 - 最终验收

_Task: `tsk_01KP5TXWA8P0AWVPK8AVKJ` / 2026-04-14_

承接 `tsk_01KP5BKHXP6MYKWC9BKMR1` 的深测结论，把 8 个高风险页面（4 配置 + 2
工具 + 2 发送）的建议清单转换为**可交付代码 + 可长期复用的 Playwright
回归脚本**，并清理历史遗留噪音。本文件是验收单，把改了什么 / 证据在哪
/ 剩余什么 一次说清。

## 1. 范围与成果

| 分组 | 页面 | 代码整改 | 回归脚本 |
|------|------|----------|----------|
| A 配置 | register-entry / login-page-template / app / menu-space | field-error data-testid + `:error` 回显契约 + host-conflict 结构化提示 | `frontend/e2e/tests/high-risk-config.real.spec.ts`（4 用例） |
| B 工具 | access-trace / api-endpoint | trace-summary/trace-node 结构化节点；api-endpoint-*-button + stale-dialog/table | `frontend/e2e/tests/high-risk-tool.real.spec.ts`（2 用例） |
| C 发送 | system/message + collaboration-workspace/message | send-field-*/send-error/send-preview-dialog/send-status（含 dry-run 沙箱） | `frontend/e2e/tests/high-risk-send.real.spec.ts`（2 用例） |
| 基础设施 | dashboard 首屏 ERR_ABORTED / 请求竞态 | AbortController 收敛 + ERR_CANCELED 日志降级 | `frontend/e2e/support/high-risk.ts` + 根因文档 |

## 2. 关键文件

**整改 / 新增（未 commit，按需让用户挑选）**

- `frontend/src/utils/http/error.ts` — ERR_CANCELED 降级为 DEV-only debug
- `frontend/src/api/message.ts` — fetchGetInboxSummary 接受 signal；dispatch 支持 dry_run
- `frontend/src/store/modules/message.ts` — loadSummary 接受 signal
- `frontend/src/domains/governance/api/app.ts` — fetchGetApps 接受 signal
- `frontend/src/components/core/layouts/art-header-bar/index.vue` — summary fetch 挂 AbortController
- `frontend/src/components/core/layouts/art-header-bar/widget/ArtAppSwitcher.vue` — loadApps 挂 AbortController
- `frontend/src/views/message/modules/message-dispatch-console.vue` — 完整 el-form rules + field-error + preview dialog + send-status marker + dry-run 开关
- `frontend/src/views/system/{register-entry,login-page-template,app,menu-space,access-trace,api-endpoint}` — field-error data-testid / 结构化结果节点
- `frontend/e2e/support/high-risk.ts` — waitForDashboardReady / collectAbortEvents / filterApiAborts
- `frontend/e2e/tests/high-risk-{config,tool,send}.real.spec.ts` — 8 个 real-backend 断言用例
- `docs/guides/dashboard-request-abort-rootcause.md` — 一页根因 + 对策
- `docs/guides/high-risk-remediation-matrix.md`（已沉淀的建议清单）
- `docs/guides/frontend-observability-spec.md`（已沉淀的观测基线）
- `output/playwright/high-risk-deep-audit-v2/high-risk-deep-report-v2.json` + `diff.md` — v2 静态 diff 报告

## 3. 验证

本轮遵循"改代码只做 build/type-check，不自动起 dev server"的用户策略：

```
pnpm -C frontend exec vue-tsc --noEmit        # 通过
pnpm -C frontend exec playwright test --list  # 8 个 high-risk-*.real.spec 用例已识别
```

用户在真实后端环境（`E2E_USERNAME`/`E2E_PASSWORD` 已配置）跑下列命令即可
交叉验证 v2 diff 报告：

```
pnpm -C frontend test:e2e e2e/tests/high-risk-config.real.spec.ts
pnpm -C frontend test:e2e e2e/tests/high-risk-tool.real.spec.ts
pnpm -C frontend test:e2e e2e/tests/high-risk-send.real.spec.ts
```

## 4. 结论 / 决策 / 遗留

### 结论

- v1 报告 12 条残留项 v2 闭合 11 条，覆盖率 **92%**。
- dashboard 首屏 ERR_ABORTED 不是真实 bug，是动态路由再导航导致的 Layout
  重挂载副作用；用 AbortController + 日志降级消除回归噪音。
- 发送链路的预览弹层 + dry-run 开关允许 E2E 在**不落库**前提下反复跑。

### 决策

- 不改 `beforeEach.handleDynamicRoutes` 的 `next({replace:true})`；风险面
  大且无收益。
- 用 Pinia `loading` 标志 + AbortController 作防重入，不引入 SWR 层。
- Playwright 脚本采用 `*.real.spec.ts` 命名约定，通过 `E2E_USERNAME/PASSWORD`
  决定 skip，不影响非真实回归。

### 遗留风险 / 下一步

| ID | 描述 | 优先级 | 责任模块 |
|----|------|--------|----------|
| api-sync-result-marker | API 管理同步结果仅用 ElMessageBox.alert 文案承载；建议新增 `data-testid=api-sync-result` hidden 节点暴露 `{processed,created,updated,totalCount}` | P3 | frontend/views/system/api-endpoint |

## 5. 关联任务

- 上游：`tsk_01KP5BKHXP6MYKWC9BKMR1`（高风险页面深测）
- 本任务：`tsk_01KP5TXWA8P0AWVPK8AVKJ`（本文件对应的交付）
- 后续：仅剩 P3 backlog 一项，不阻塞继续推进其他高风险面扩展。
