# CLEANUP-V1 P1-A / P1-B 笔记

## P1-A 当前真实现状

当前前端并不是两条错误链，而是三条并存：

| 维度       | `frontend/src/utils/http/index.ts`                             | `frontend/src/api/system-manage/_shared.ts`                                             | `frontend/src/api/v5/client.ts`                         | 是否是 bug 源                                            |
| ---------- | -------------------------------------------------------------- | --------------------------------------------------------------------------------------- | ------------------------------------------------------- | -------------------------------------------------------- |
| 响应协议   | 依赖 `{ code, data, msg/message }` 信封，`code === 0` 视为成功 | 基于 openapi-fetch 的 `{ data, error, response }`，从 `error` 和 `response.status` 归一 | 直接看 `Response.status`                                | 是。旧链按业务信封，v5 按裸 HTTP，调用方心智不一致       |
| 401 检测   | `code === 401` 或 `response.status === 401`                    | `response.status === 401`，缺少统一业务码识别                                           | `response.status === 401`                               | 是。401 / 2001 / token 过期等场景可能分叉                |
| 401 防抖   | 本地 `unauthorizedHandling` Promise 哨兵                       | 自己再维护一份 `unauthorizedHandling` Promise 哨兵                                      | 无 logout 防抖，只负责 refresh Promise                  | 是。两套 sentinel 互相不可见，可能重复 logout / 重复弹错 |
| token 刷新 | 无自动刷新，401 直接登出                                       | 无自动刷新，只在 401 时调用 `userStore.logOut()`                                        | `refreshSessionIfNeeded()`，仅 v5 fetch middleware 接入 | 是。axios 老接口和 v5 接口在同一页面并存时，行为不一致   |
| 错误展示   | 走 `utils/http/error.ts` 的 `showError()/showCodes/hideCodes`  | 只在 401 时 `showError`，其余只构造 `HttpError` 抛出                                    | 不负责展示                                              | 是。v5 底层没有和 legacy 共用同一套过滤策略              |

### 结论

1. `P1-A` 的收口目标不是“把 v5 接上 legacy”，而是把三处底层协议统一成一套：
   - 同一个未授权处理入口
   - 同一个 refresh 入口
   - 同一个 v5 错误归一规则
2. 任务树里 `P1A-1` 对 `_shared.ts` 的路径已经过期。
   - 旧描述：`frontend/src/api/v5/_shared.ts`
   - 当前真实路径：`frontend/src/api/system-manage/_shared.ts`

## P1-A 目标规范

### 统一错误拦截职责

| 模块 | 职责 |
| --- | --- |
| `frontend/src/utils/http/error.ts` | `HttpError`、展示过滤规则、`showError` / `showSuccess` |
| `frontend/src/utils/http/auth-session.ts` | 单一 unauthorized sentinel、单一 refresh promise、axios/fetch 共用的 401 旁路规则 |
| `frontend/src/utils/http/v5.ts` | v5 error 到 `HttpError` 的统一归一，不再在 `_shared.ts` 自己维护一套未授权处理 |
| `frontend/src/utils/http/index.ts` | legacy axios 请求/响应拦截，但未授权、refresh、错误码判断全部复用共享模块 |
| `frontend/src/api/v5/client.ts` | 只保留 openapi-fetch transport middleware，refresh 政策复用共享模块 |

### 统一规则

1. `HttpError.code` 优先保留业务错误码；没有业务错误码时才回退到 HTTP status。
2. 未授权判断统一为：
   - 优先 HTTP `401`
   - fallback 到业务错误码 `Unauthorized` / `TokenExpired`
3. `401` 防抖只允许存在一份 Promise sentinel。
4. refresh 只允许存在一份 Promise sentinel，axios 和 v5 共用。
5. v5 error 归一后必须进入 `HttpError`，不再在 `_shared.ts` 自己处理 logout。

## P1-B 当前真实现状

认证流程还散落在页面组件内：

| 流程         | 当前位置                                                         | 现状                                                                |
| ------------ | ---------------------------------------------------------------- | ------------------------------------------------------------------- |
| 登录         | `frontend/src/views/auth/login/index.vue`                        | 表单、记住密码、session 应用、上下文初始化、跳转全写在页面里        |
| 注册         | `frontend/src/views/auth/register/index.vue`                     | 注册上下文拉取、表单校验、自动登录跳转写在页面里                    |
| 回调换 token | `frontend/src/views/auth/callback/index.vue`                     | state 校验、exchange、session 应用、刷新用户/菜单、跳转全写在页面里 |
| 登出         | `frontend/src/store/modules/user.ts` + 多处 `userStore.logOut()` | 调用点分散，中心化登录跳转策略未统一                                |

### 结论

1. `auth-flow` 需要做成“页面薄、流程厚”：
   - 页面只保留表单、loading、alert、展示文案
   - session 应用、上下文刷新、跳转策略下沉到 composable / store
2. `logout` 不能只新建 composable，不改底层。
   - 当前大量代码直接调用 `userStore.logOut()`
   - 统一方案应该是：先把 `userStore.logOut()` 升级为统一实现，再让新代码经由 `useLogoutFlow()`

## P1-B auth-flow API

### 目录

- `frontend/src/composables/auth-flow/shared.ts`
- `frontend/src/composables/auth-flow/useLoginFlow.ts`
- `frontend/src/composables/auth-flow/useRegisterFlow.ts`
- `frontend/src/composables/auth-flow/useCallbackFlow.ts`
- `frontend/src/composables/auth-flow/useLogoutFlow.ts`

### 接口

| Flow | 输入 | 输出 | 内部职责 |
| --- | --- | --- | --- |
| `useLoginFlow()` | 页面表单状态 + 当前路由 query | `loading`、`submitError`、`submit()` | 登录 API、remember-password、session 应用、上下文初始化、登录后跳转 |
| `useRegisterFlow()` | 注册表单状态 | `ctx`、`contextError`、`loading`、`loadContext()`、`register()` | 注册上下文获取、注册提交、自动登录 / 回登录页跳转 |
| `useCallbackFlow()` | 当前 callback 路由 query | `message`、`run()` | state 校验、token exchange、session 应用、用户/菜单刷新、目标页跳转 |
| `useLogoutFlow()` | 无 | `logout()` | 包装统一后的 `userStore.logOut()` |

### 页面职责边界

1. 页面只保留表单、`loading`、`error alert`、文案和基础校验。
2. composable 负责调用 API、写 store、初始化 context、跳转。
3. `logout` 真正的统一出口在 `userStore.logOut()`，`useLogoutFlow()` 只是面向新代码的薄包装。
