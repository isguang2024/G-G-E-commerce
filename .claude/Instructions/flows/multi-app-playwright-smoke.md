# 多 APP 主链 Smoke Matrix

用于 V2 `5.1` 节点收口。目标不是替代完整 E2E，而是把当前最容易回归的多 APP 主链固定成一套可重复执行的浏览器检查矩阵，优先覆盖登录、APP 切换、治理页 smoke 三类场景。

## 适用端口拓扑

- 前端：`http://127.0.0.1:5174`
- 后端：`http://127.0.0.1:8080`
- 已完成本地 seed，至少包含：
  - `platform-admin`
  - `account-portal`
  - `demo-app`

## 执行方式

- 优先使用 Playwright 浏览器自动化执行。
- 当前仓库未内置 `playwright` npm 依赖，因此默认采用“稳定执行说明 + 可读断言”的方式固化矩阵。
- 若后续要落成仓库内脚本，可在补入 Playwright 依赖后，直接把下列场景翻译为 `.spec.ts`。

## Smoke Matrix

### 场景 1：后台未登录访问必须走 centralized login

- 起点：清空 localStorage / sessionStorage / cookies 后访问 `http://127.0.0.1:5174/system/page`
- 期望：
  - 页面跳转到 `/account/auth/login`
  - URL 必须包含以下 query：
    - `target_app_key=platform-admin`
    - `redirect_uri=http://127.0.0.1:5174/account/auth/callback`
    - `target_path=%2Fsystem%2Fpage`
    - `state`
    - `nonce`
    - `auth_protocol_version=callback-v1`
- 失败定位：
  - 若只剩 `redirect=/system/page`，优先检查 [beforeEach.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/guards/beforeEach.ts)
  - 若跳到错误 host，优先检查 [menu-space.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/store/modules/menu-space.ts) 与 [ArtAppSwitcher.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/core/layouts/art-header-bar/widget/ArtAppSwitcher.vue)

### 场景 2：公开 demo 页面未登录访问不得被 401 覆盖成普通登录

- 起点：清空 localStorage / sessionStorage / cookies 后访问 `http://127.0.0.1:5174/demo/lab`
- 期望：
  - 4 秒后 URL 仍是 `/demo/lab`
  - 页面标题为 `Demo 验证页 - G&G-E`
  - 浏览器控制台没有 `401` 导致的二次跳登录错误
- 失败定位：
  - 若被覆盖到 `/account/auth/login?redirect=/demo/lab`，优先检查 [art-header-bar/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/core/layouts/art-header-bar/index.vue) 与 [ArtAppSwitcher.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/core/layouts/art-header-bar/widget/ArtAppSwitcher.vue)
  - 若公开页本身 404，优先检查 [ManagedPageProcessor.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/core/ManagedPageProcessor.ts) 与运行时页面注册链

### 场景 3：认证中心登录后必须回跳后台治理页

- 起点：从场景 1 的登录页继续，使用本地 seed 管理员账号登录
- 期望：
  - 登录后进入 `/account/auth/callback`
  - 随后命中：
    - `POST /api/v1/auth/callback/exchange`
    - `GET /api/v1/auth/me`
    - `GET /api/v1/runtime/navigation`
  - 最终回到 `http://127.0.0.1:5174/system/page`
- 失败定位：
  - 若 callback 失败，优先检查 [frontend/src/views/auth/callback/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/auth/callback/index.vue) 与后端 centralized auth service
  - 若登录成功但没回到目标页，优先检查 [login/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/auth/login/index.vue) 的 `gotoAfterLogin`

### 场景 4：平台后台与 demo-app 切换链路必须可往返

- 起点：已登录后台工作台
- 操作：
  - 从 APP 切换器切到 `demo-app`
  - 再从 `demo-app` 切回 `platform-admin`
- 期望：
  - 能进入 `http://127.0.0.1:5174/demo/lab`
  - 再回到 `http://127.0.0.1:5174/dashboard/console`
  - 控制台 `error = 0`
- 失败定位：
  - 若切到错误入口或 404，优先检查 [ArtAppSwitcher.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/core/layouts/art-header-bar/widget/ArtAppSwitcher.vue)
  - 若切回后台后导航为空，优先检查 [beforeEach.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/guards/beforeEach.ts) 与 runtime navigation 刷新链

### 场景 5：治理页 smoke

- 起点：已登录后台
- 访问：
  - `http://127.0.0.1:5174/system/app`
  - `http://127.0.0.1:5174/system/page`
- 期望：
  - `App 管理` 正常展示认证中心、预检查、环境配置、Feature Flag、敏感配置治理区
  - `页面管理` 正常展示本地配置 / 扫描同步 / 远端页来源统计
  - 控制台不再出现 `menu-space-entry-bindings` 404
- 失败定位：
  - `App 管理` 异常优先看 [frontend/src/views/system/app/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/app/index.vue)
  - `页面管理` 异常优先看 [frontend/src/views/system/page/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/page/index.vue)
  - 若是运行时接口 404，先对照 [backend/internal/api/router/router_contract_test.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/api/router/router_contract_test.go)

## 当前已验证结果

- 场景 1：通过
- 场景 2：通过
- 场景 4：此前已通过，本轮未重跑
- 场景 5：此前已通过，本轮未重跑

## 观测与排障字段

- `app_key`
  - 来源：治理台 `App 管理` 当前选中应用、preflight `诊断标签`
  - 用途：先确定故障发生在哪个 APP，不要把 `platform-admin` 和 `demo-app` 混看
- `space_key`
  - 来源：`menu-space` 解析结果、运行时 navigation manifest、菜单空间入口绑定
  - 用途：区分是 APP 入口解析问题，还是 APP 内部的 Level 2 空间入口落错
- `auth_mode`
  - 来源：`App 管理` 认证中心、preflight `诊断标签`
  - 用途：判断应该查 centralized login、shared cookie，还是本地登录链
- `request_host`
  - 来源：当前应用解析响应、preflight request host
  - 用途：判断故障是否只在 `127.0.0.1 / localhost / 业务域名` 某一个 host 下发生
- `probe_status`
  - 来源：`SystemAppItem.probe_status`、preflight `诊断标签`
  - 用途：先区分“探针未配置”和“探针已配置但应用行为异常”
- `request_id`
  - 当前状态：这轮没有把它直接挂到治理台卡片；排障时优先从后端 access log / 审计日志查请求链
  - 后续建议：如果继续推进 V2 `5.2`，应把 request_id 透传到治理台详情或响应 header 对照入口

## 后续落地建议

- 若仓库决定引入正式 Playwright 依赖，优先把场景 1、2、4 先固化成自动化脚本，因为它们最容易被认证策略、壳层组件和 runtime navigation 回归打断。
- 若继续推进 V2 `5.2`，建议把关键失败点同步打到治理日志或 health probe 输出里，不要只留给浏览器控制台。
