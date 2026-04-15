# 高风险页面深测整改矩阵

> 基于任务 `tsk_01KP5BKHXP6MYKWC9BKMR1`（高风险页面浏览器自动化深度测试）输出的结论，为后续整改任务 `tsk_01KP5TXWA8P0AWVPK8AVKJ` 提供落地清单。

## 证据与来源

| 类型 | 路径 |
|------|------|
| 报告 | `output/playwright/high-risk-deep-audit/high-risk-deep-report.json` |
| 结构探测 | `output/playwright/high-risk-deep-audit/deep-inspect.json` |
| 截图 | `output/playwright/high-risk-deep-audit/*.png`（含每页 before/after + 结构 inspect） |
| 上一轮 wrapup | task_tree task `tsk_01KP5BKHXP6MYKWC9BKMR1` |

**深测结论要点**：8/8 页面均可进入专项流程，未发现 P0/P1 阻断问题；风险集中在"可观测性不足"——配置页保存校验不显式、工具页结果区无结构化、发送页必填/确认反馈不稳定。

---

## A. 配置类页面（保存后回显 + 必填校验）

### A1. 注册入口 `/system/register-entry`

| 字段 | 值 |
|------|----|
| 现象 | 新建入口弹层可打开；保存校验尝试时未捕获明确的必填校验文本，保存后字段回显不明 |
| 重现路径 | 登录 → 系统管理 → 注册入口 → 新建入口 → 直接提交/留空关键字段 |
| 证据 | `01-register-entry-inspect.png` / `register-entry-before.png` / `register-entry-after.png`，report.issues=[] |
| 表头 | 序号 / 入口 Code / 名称 / App / 命中规则 / 公开注册 / 自动登录 / 状态 / 操作 |
| 优先级 | P2 |
| 建议修复 | 1) 补齐 `el-form` 必填项 `rules` + `data-testid=register-entry-field-error`；2) 后端 `backend/internal/api/handlers/system_register.go` 校验失败返回 `code + field` 结构化错误；3) 保存成功后表格按新 Code 高亮 & 列表 `updated_at` 回填 |
| 整改节点 | `fix-register-entry` (`nd_01KP5TXWAPQRDKM6CZC429`) |

### A2. 登录页模板 `/system/login-page-template`

| 字段 | 值 |
|------|----|
| 现象 | 新建模板弹层可打开；无明显错误，但预览入口与保存后的列表/详情一致性未验证 |
| 重现路径 | 系统管理 → 登录页模板 → 新建模板 → 填写 Key + 标题 → 预览/保存 |
| 证据 | `02-login-template-inspect.png` / `login-template-before.png` / `login-template-after.png` |
| 表头 | 序号 / 模板 Key / 名称 / 场景 / 作用域 / 状态 / 默认模板 / 配置概览 / 操作 |
| 优先级 | P2 |
| 建议修复 | 1) 新建弹层内置**实时预览**刷新（基于当前表单草稿）；2) Key/名称 rules 必填 + 唯一性前端提示；3) 保存成功后列表置顶 + `data-testid=login-template-row[key=...]` 便于自动化 |
| 整改节点 | `fix-login-template` (`nd_01KP5TXWARXQKNPE2GNS3R`) |

### A3. 应用管理 `/system/app`

| 字段 | 值 |
|------|----|
| 现象 | 新增 App 弹层可进入；认证配置编辑、空间配置联动深层交互未被脚本触发到显式反馈 |
| 重现路径 | 系统管理 → 应用管理 → 新增 App / 编辑选中 App / 空间配置 |
| 证据 | `03-app-management-inspect.png` / `app-management-before.png` / `app-management-after.png` |
| 关键按钮 | 新增 App / 编辑选中 App / 新增入口绑定 / 空间配置 / 高级空间配置 / 入口规则 / 快速预检 |
| 优先级 | P2 |
| 建议修复 | 1) 认证配置编辑抽屉错误锁定到具体字段（`el-form-item`）；2) 保存后应用卡片立即刷新 `updated_at` 与 status；3) 空间配置联动失败时错误消息含可读 reason（app_key / space_key） |
| 整改节点 | `fix-app-management` (`nd_01KP5TXWAS6A8TNBSFNN11`) |

### A4. 高级空间配置 `/system/menu-space`

| 字段 | 值 |
|------|----|
| 现象 | 新增空间/新增 Host 绑定/编辑/空间布局入口均可打开；Host 冲突保存时无结构化反馈 |
| 重现路径 | 系统管理 → 高级空间配置 → 新增空间/新增 Host 绑定 → 提交 |
| 证据 | `04-menu-space-inspect.png` / `menu-space-before.png` / `menu-space-after.png` |
| 关键按钮 | 保存模式 / 新增空间 / 新增 Host 绑定 / 空间布局 / 受管页面 / 绑定 Host |
| 优先级 | P2 |
| 建议修复 | 1) 空间新建后列表立即出现空间卡片 + `data-testid=menu-space-card[key=...]`；2) Host 冲突后端返回结构化 code（`host_conflict` + 冲突 host 列表），前端渲染为提示气泡；3) `绑定 Host` 动作成功后卡片 Host 数同步自增 |
| 整改节点 | `fix-menu-space` (`nd_01KP5TXWAT093YNJ4ZDT33`) |

---

## B. 工具类页面（结构化结果与副作用提示）

### B1. 访问链路测试 `/system/access-trace`

| 字段 | 值 |
|------|----|
| 现象 | 可输入并执行"测试链路"，页面无报错，但结果区读到的文本缺少自动化友好的结构化断言点 |
| 重现路径 | 系统管理 → 访问链路测试 → 选择 App/菜单空间/角色/用户/页面Key → 测试链路 |
| 证据 | `05-access-trace-inspect.png` / `access-trace-before.png` / `access-trace-after.png` |
| 优先级 | P2 |
| 建议修复 | 1) 结果区改为 `data-testid=trace-node` 列表节点，每项含 `data-role / data-status(ok\|fail) / data-latency / data-reason`；2) 失败节点返回结构化 reason（permission_denied / menu_not_bound / host_missed 等）；3) 顶部汇总 `data-testid=trace-summary` 展示 allow/deny 总数 |
| 整改节点 | `fix-access-trace` (`nd_01KP5TXWAYDJ5BT0QTVKKK`) |

### B2. API 管理 `/system/api-endpoint`

| 字段 | 值 |
|------|----|
| 现象 | `/auth` 查询主链路正常；同步/清理等副作用按钮未被脚本触发；数据层曾显示"失效"标签 |
| 重现路径 | 系统管理 → API 管理 → 全局同步 API / 全局清理失效 API / 查询 keyword=/auth |
| 证据 | `06-api-management-inspect.png` / `api-management-before.png` / `api-management-after.png`；报告中捕获 `/auth` 结果行 |
| 关键按钮 | 全局同步 API / 全局未注册 API / 全局清理失效 API |
| 优先级 | P2（底层数据历史遗留已通过 `cleanup-stale-api` 整改落地） |
| 建议修复 | 1) 同步/清理按钮增加 `ElMessageBox` 二次确认（列出将被 removed 的 codes）；2) 成功后弹层展示 `{added, updated, removed}` 数量摘要 `data-testid=api-sync-result`；3) 后端返回结构化 diff 便于前端渲染与 E2E 断言 |
| 已完成 | 清理 5 条 2026-04-08 孤儿（POST /api-endpoints + GET/POST/PUT/DELETE /menus/groups），bindings 同步硬删，见 `cleanup-stale-api` |
| 整改节点 | `fix-api-management` (`nd_01KP5TXWAZ63JSN6KMB3HX`) |

---

## C. 发送类页面（必填校验 + 发送确认链路）

### C1. 系统消息发送 `/system/message`

| 字段 | 值 |
|------|----|
| 现象 | 可执行发送前操作并填入部分输入，页面无报错，但未直接捕获明确的必填校验文本 |
| 重现路径 | 系统管理 → 消息发送 → 选择类型/优先级/发送对象 → 留空标题或正文 → 发送消息 |
| 证据 | `07-system-message-inspect.png` / `system-message-before.png` / `system-message-after.png` |
| 关键字段 | 消息模板 / 发送人 / 消息类型(notice/message/todo) / 优先级(low/normal/high/urgent) / 发送对象 / 标题 / 摘要 / 正文(富文本) / 业务分类 / 失效时间 |
| 优先级 | P1（信箱/通知链路与用户可见度强相关） |
| 建议修复 | 1) 表单 rules 必填项：标题 / 摘要 / 正文 / 发送人 / 发送对象，错误文本 `data-testid=send-field-error[name=...]`；2) 发送按钮前插入预览弹层 `data-testid=send-preview`（含收件人摘要、标题、正文纯文本片段）；3) 发送结果 toast 改为语义节点 `data-testid=send-status[state=success\|failed]` 含 code |
| 整改节点 | `fix-system-message-required` + `fix-system-message-confirm` (`nd_01KP5TXWB3D9GWGQCS98F3` / `nd_01KP5TXWB5YERE2WB6VX2K`) |

### C2. 协作空间消息发送 `/collaboration-workspace/message`

| 字段 | 值 |
|------|----|
| 现象 | 交互路径与系统消息发送基本一致；空间范围 / 成员目标选择的真实校验未覆盖 |
| 重现路径 | 协作空间 → 消息发送 → 选空间范围 / 成员目标 → 发送 |
| 证据 | `08-collab-message-inspect.png` / `collab-message-before.png` / `collab-message-after.png` |
| 优先级 | P1 |
| 建议修复 | 1) 空间范围 + 成员目标未选时显式错误（复用 `send-field-error`）；2) 成员目标支持选中人数汇总 `data-testid=send-target-count`；3) 与系统消息发送共用校验/状态节点规范，避免脚本碎片化 |
| 整改节点 | `fix-collab-message` (`nd_01KP5TXWB6XH3VRAAA9X9N`) |

---

## D. 全局与基础设施

### D1. dashboard 首屏 request aborted

| 字段 | 值 |
|------|----|
| 现象 | 报告 `globalIssues` 记录 7 条 `ERR_ABORTED`：`/api/v1/messages/inbox/summary`、`/api/v1/system/apps`、`art-global-search/index.vue` 等 |
| 证据 | report.json 中 `globalIssues`；`/dashboard/console` 首屏发生 |
| 推测根因 | 热更新 + 路由 guard 双重触发导致中断请求；不是路径级 bug |
| 优先级 | P3（信号噪音，对比回归时干扰） |
| 建议修复 | 1) Pinia dashboard store 首屏初始化加 AbortController / dedupe；2) 路由 guard 命中同路径时短路；3) E2E 启动前 `waitForLoadState('networkidle')` |
| 整改节点 | `fix-dashboard-aborted` + `fix-request-race` (`nd_01KP5TXWBA734S6GVV1SFX` / `nd_01KP5TXWBCBWB3N0MHBNKC`) |

### D2. 沙箱 / 发送 dry-run 通道

| 字段 | 值 |
|------|----|
| 现象 | 上一轮测试为非破坏性，未真正落库/外发；二轮需要反复跑发送与配置场景 |
| 建议修复 | 1) 后端消息发送接口新增 `dry_run` 字段（仅校验，不落库/不外发），在 OpenAPI 契约中声明；2) 前端 E2E 通过 query/header 开启 dry-run；3) 测试数据用 `tenant=test` 或专用 fixture 脚本隔离，避免污染 default 租户 |
| 整改节点 | `send-dry-run` + `sandbox-and-fixtures` (`nd_01KP5TXWB8GZC8APH622K2` / `nd_01KP5TXWAMNPB3EEDB39G9`) |
| 已完成 | 后端 `MessageDispatchRequest.dry_run` 字段已落 spec + ogen + service 分支（`dispatch_status="preview"` 零副作用返回）；`cmd/fixtures-sandbox` 夹具脚本落地（`sandbox.deep-probe.notice` 模板 + `[SANDBOX] Playwright Deep Probe` sender，支持 `--cleanup` / `--purge`）——见 `sandbox-and-fixtures` 节点 |

---

## E. 优先级与整改任务映射

| 优先级 | 页面 | 节点 key | 状态 |
|-------|------|----------|------|
| P1 | 系统消息发送 | fix-system-message-required, fix-system-message-confirm | ready |
| P1 | 协作消息发送 | fix-collab-message | ready |
| P2 | 注册入口 | fix-register-entry | ready |
| P2 | 登录页模板 | fix-login-template | ready |
| P2 | 应用管理 | fix-app-management | ready |
| P2 | 高级空间配置 | fix-menu-space | ready |
| P2 | 访问链路测试 | fix-access-trace | ready |
| P2 | API 管理 | fix-api-management | ready |
| P2 | API stale 清理 | cleanup-stale-api | **done** ✅ |
| P3 | dashboard aborted | fix-dashboard-aborted, fix-request-race | ready |
| — | 可观测规范 | observability-spec | **done** ✅（`docs/guides/frontend-observability-spec.md`） |
| — | 沙箱 fixtures + dry-run | sandbox-and-fixtures | **done** ✅（`cmd/fixtures-sandbox` + `MessageDispatchRequest.dry_run`） |

---

## F. 可观测标记规范预告

下一个节点 `observability-spec` 将把下列约定落为短文档 + 代码示例，作为所有整改的统一依据：

- **前端表单错误**：`<el-form-item :data-testid="'send-field-error'" :data-field="fieldName">` + 错误 `rules` 必填；
- **结果/状态节点**：`data-testid=trace-node | send-preview | send-status | api-sync-result` 等语义节点；
- **后端错误响应**：`{ code: "host_conflict", field: "hosts[0]", message: "..." }` 的统一结构，前端按 `code` 渲染提示。

---

_最后更新：由 task `tsk_01KP5TXWA8P0AWVPK8AVKJ / collect-findings` 自动生成。_
