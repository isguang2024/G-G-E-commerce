# Dashboard 首屏 `ERR_ABORTED` 根因与对策（1 页说明）

_Last updated: 2026-04-14 · owner: frontend-observability_

## 结论先行

首屏 `/api/v1/messages/inbox/summary` 与 `/api/v1/system/apps` 出现的
`ERR_ABORTED` **不是真实 bug**，而是 **路由守卫再导航触发的
Layout 组件重挂载**，导致第一次 mount 期间发起的 fetch 被浏览器取消。
不影响业务功能，但会在 DevTools 的 console / network 产生批量红色记录，
给回归判读造成噪音。本次整改不改变行为，只做两件事：

1. 把 ERR_CANCELED 的 console 输出降级为 DEV-only debug，生产回归面板不再出现红条；
2. 在会在 mount 阶段主动发请求的组件（`ArtAppSwitcher`、`art-header-bar` 的 `loadSummary`）
   里挂一个 AbortController，和 `onBeforeUnmount` 绑定，让取消成为"已知可预期的"
   事件而不是浏览器兜底 abort。

## 复现路径

1. `router/guards/beforeEach.ts::handleDynamicRoutes`（L500–L595）在用户已登录、
   进入鉴权路由时会：
   - 并行预取 userInfo + runtime navigation；
   - `ensureAuthenticatedRoutes()` 注册动态路由；
   - **最后一定 `next({ path: to.path, replace: true })`**（L569–L574）——即使目标
     和当前相同，也会触发一次"替换型"导航。
2. `replace: true` 使 Vue Router 重新挂载 Layout（`components/core/layouts/art-header-bar/*`），
   Layout 组件里的 `onMounted` 里现在 fire 的请求被丢弃，浏览器抛 `ERR_CANCELED`。
3. 观察到的三类"孤儿请求"：
   - `/system/apps` — 来自 `ArtAppSwitcher.onMounted → loadApps()`（L163–L165）；
   - `/messages/inbox/summary` — 来自 `art-header-bar/index.vue onMounted → messageStore.loadSummary()`（L221–L229）；
   - `art-*` 组件 — 实际上只是在 Network 面板里按组件名误读，没有独立的 HTTP 调用。

## 为什么过去没有挂

- `frontend/src/utils/http/error.ts::handleError`（L127–L130）已经把 `ERR_CANCELED`
  映射成 `HttpError(message, ApiStatus.error=400)`；
- `shouldShowErrorMessage(400)` 命中默认分支 `code >= 5000 === false`（L237），
  所以不会弹 ElMessage；
- 但函数里还有一行 `console.warn('Request cancelled:', ...)` —— 这是 DevTools 里
  看到的"红条"噪音源头。

## 对策（本次实现）

### A. 降级 console 输出

`src/utils/http/error.ts`

```ts
if (error.code === 'ERR_CANCELED') {
  if (import.meta.env.DEV) {
    // 仅 DEV 环境输出，生产/回归面板不再产生噪音
    console.debug('[http] request cancelled:', error.config?.url || error.message)
  }
  throw new HttpError($t('httpMsg.requestCancelled'), ApiStatus.error)
}
```

### B. 组件层 AbortController（ArtAppSwitcher、loadSummary）

在两个 onMounted 发起请求的组件里加一个 `AbortController`，`onBeforeUnmount`
里 `.abort('component-unmount')`，同时把 `signal` 透传给 `fetchGetApps` /
`fetchGetInboxSummary`。这样即使 Layout 再次挂载，第一次的 fetch 也会**主动**
被 abort，不再依赖浏览器的"DOM 销毁连带取消"。

对 API 层的最小改动是给两个函数添加可选 `signal?: AbortSignal` 参数，
`openapi-fetch` 的第二参数支持 `signal` 字段。

### C. 不做的事（deliberate 非目标）

- **不改** `beforeEach.handleDynamicRoutes` 的 `replace: true` 逻辑。它是
  运行时路由表重建后回到目标路径的标准做法，改它风险大于收益。
- **不做** SWR 级别的全局请求复用缓存，本次只消除噪音。
- **不拦截** ERR_CANCELED 的业务回调——上层若 await 了这个 Promise，
  仍然会 catch 到 HttpError，由业务自行处理（本次确认 `ArtAppSwitcher.loadApps`
  / `messageStore.loadSummary` 均已 `.catch(() => undefined)`）。

## 回归验证方式

1. 登录并访问 `/` —— DevTools Network 不再出现红色 `(canceled)` 记录；
2. 以 production build 预览首屏，console 没有 `Request cancelled` 输出；
3. 触发语言切换 / 协作工作空间切换 等会导致 Layout 重挂载的操作，
   第一次 mount 的请求产生 abort，但没有新的 ElMessage 或 console 记录；
4. `pnpm run type-check && pnpm run build` 通过。

## 参考

- `docs/guides/frontend-observability-spec.md` §2.4 观测基线
- `frontend/src/router/guards/beforeEach.ts` 动态路由注册流程
- `frontend/src/utils/http/error.ts` HTTP 错误路径
