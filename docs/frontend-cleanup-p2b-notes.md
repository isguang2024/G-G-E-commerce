# 前端收口 P2-B：兼容层清单

更新时间：2026-04-12

## 目标

- 盘点前端仍在承担“旧接口形状 / 旧路径 / 旧运行时出口”兼容职责的代码。
- 为 P2B-2 的 `@compat-status` 标注和 P2B-3 的第一批删除提供依据。
- 区分“必须暂留”和“已可收口”的兼容层，避免误删仍有调用面的桥接代码。

## 分类总览

| 分类 | 当前职责 | 代表文件 | 当前判断 |
| --- | --- | --- | --- |
| Legacy HTTP 封装 | 承载 axios 链路、请求重试、旧错误封装 | `frontend/src/utils/http/index.ts` | 过渡期兼容 |
| v5 -> 旧 `Api.*` 类型桥接 | 把 v5 schema 扁平化回旧前端类型 | `frontend/src/api/system-manage/_shared.ts`、`frontend/src/api/auth.ts`、`frontend/src/api/workspace.ts` | 过渡期兼容 |
| 旧路径 / 旧入口兼容 | 老路由重定向到 account-portal 新入口 | `frontend/src/router/routes/staticRoutes.ts` | 过渡期兼容 |
| 运行时 façade 兼容导出 | 维持旧 guard/router 出口签名 | `frontend/src/router/guards/beforeEach.ts`、`frontend/src/router/index.ts` | 计划删除 |
| worktab / app-scope 历史迁移残留 | 兼容旧的 scope 与最后用户态切换 | `frontend/src/store/modules/user.ts`、`frontend/src/store/modules/worktab.ts`、`frontend/src/utils/app-scope.ts` | 继续保留 |

## 1. Legacy HTTP 封装

### 1.1 `frontend/src/utils/http/index.ts`

- `27` 仍引用 `retryAxiosRequestWithRefresh`，说明旧 axios 链路仍承担 refresh 重试。
- `79` 通过 `axios.create(...)` 自建实例，未完全切到 v5 OpenAPI client。
- `153`、`173` 仍在 response 拦截器中做 401 重试。
- `189` 定义 `createHttpError(...)`，继续向旧调用方暴露 axios 风格错误对象。
- `205`、`226`、`267`、`290-302` 保留 `retryRequest/request/doRequest/get/post/put/delete` 整套旧入口。

当前用途判断：

- 该文件不再是认证错误模型的真相源，但仓库中仍有大量 axios 风格 API 封装与调用习惯。
- 已经接入 `auth-session.ts` 后，401/refresh 的“双链分歧”部分已显著收口；但请求入口本身还没整体迁到 `v5Client`。

建议状态：

- `transition`

## 2. v5 -> 旧类型桥接

### 2.1 `frontend/src/api/system-manage/_shared.ts`

- `38` 定义 `V5PageLike`，`127` 定义 `V5PermissionActionLike`，这是典型的 v5 schema 适配输入层。
- `195-196` 明确写着“保持 `Api.SystemManage.UserListItem` 等消费方契约不变”。
- `228` `normalizePermissionKey(...)`、`240` `derivePermissionSegments(...)` 继续把 v5 权限 key 映射回旧权限分段语义。
- `488` `normalizePageItem(...)` 返回 `Api.SystemManage.PageItem`。
- 直接调用面覆盖：
  - `frontend/src/api/auth.ts`
  - `frontend/src/api/collaboration-workspace.ts`
  - `frontend/src/api/message.ts`
  - `frontend/src/api/system-manage/api-endpoint.ts`
  - `frontend/src/api/system-manage/app.ts`
  - `frontend/src/api/system-manage/menu.ts`
  - `frontend/src/api/system-manage/page.ts`
  - `frontend/src/api/system-manage/permission.ts`
  - `frontend/src/api/system-manage/register.ts`
  - `frontend/src/api/system-manage/role.ts`
  - `frontend/src/api/system-manage/user.ts`

当前用途判断：

- 这是现阶段最大的兼容桥。删除它会直接波及系统管理、协作空间、消息等多个模块。
- P2-B 的第一批清理不应直接删文件本体，而应优先减少新入口继续依赖旧 `Api.SystemManage.*`。

建议状态：

- `transition`

### 2.2 `frontend/src/api/auth.ts`

- 文件头注释明确写着：登录响应继续映射为既有 `Api.Auth.LoginResponse`，避免一次性改整个登录流程。
- `11-32` 的 `fetchLogin(...)` 最终 `return data as Api.Auth.LoginResponse`。
- `88` 左右的 `fetchRefreshToken(...)`、`108` 左右的 `fetchExchangeAuthCallback(...)` 仍复用旧登录响应类型。
- `144-146` 注释与签名明确说明 `fetchGetUserInfo()` 会把 `/auth/me` 扁平化成 `Api.Auth.UserInfo`。

当前用途判断：

- 认证主链页面、store、guard 刚完成一轮 runtime 收口，但消费层仍建立在旧 `Api.Auth.*` 契约上。
- 当前不宜删除，适合在后续 session/auth-flow 稳定后整体迁移到 v5 原生类型。

建议状态：

- `transition`

### 2.3 `frontend/src/api/workspace.ts`

- `3` `normalizeWorkspace(item: any): Api.SystemManage.WorkspaceItem`
- `37`、`44`、`51`、`66` 都在持续返回旧 `WorkspaceItem` 形状。

当前用途判断：

- 这是小型桥接文件，调用面相对集中，未来适合作为较早迁移对象。

建议状态：

- `transition`

### 2.4 旧 `Api.*` 类型仍广泛存在

按 `Api.` 文本命中数量看，兼容契约仍广泛渗透：

- `frontend/src/api/system-manage/_shared.ts`: 37
- `frontend/src/api/message.ts`: 32
- `frontend/src/api/system-manage/permission.ts`: 23
- `frontend/src/api/auth.ts`: 7
- `frontend/src/api/collaboration-workspace.ts`: 14
- `frontend/src/api/system-manage/api-endpoint.ts`: 7
- `frontend/src/api/system-manage/app.ts`: 7
- `frontend/src/api/system-manage/menu.ts`: 4
- `frontend/src/api/system-manage/page.ts`: 5
- `frontend/src/api/system-manage/role.ts`: 4
- `frontend/src/api/system-manage/user.ts`: 5
- `frontend/src/api/workspace.ts`: 1

结论：

- 当前兼容层不是单点文件，而是“旧 `Api.*` 契约网络”；P2-B 应先标记和局部收口，再做批量替换。

## 3. 旧路径 / 旧入口兼容

### 3.1 `frontend/src/router/routes/staticRoutes.ts`

- `21` 注释已明确说明“旧路径兼容：重定向到 account-portal 下的新入口”。
- `23` `/auth/login -> /account/auth/login`
- `28` `/auth/register -> /account/auth/register`
- `33` `/auth/forget-password -> /account/auth/forget-password`
- `38` `/auth/callback -> /account/auth/callback`

当前用途判断：

- 这类重定向仍有外部链接、浏览器历史、中心化登录回跳路径兼容价值。
- 在未确认所有外链与登录平台都已切新路径前，不建议删除。

建议状态：

- `transition`

### 3.2 `frontend/src/views/account-portal/auth/*/index.vue`

- `login/index.vue`、`register/index.vue`、`forget-password/index.vue` 都只是薄壳，直接转发到 `@/views/auth/*/index.vue`。
- 这组文件本质上是“为新路径保留独立路由名”的壳层，而不是独立页面实现。

当前用途判断：

- 如果后续决定统一只保留 `views/auth/*` 直接挂载，这组壳层有机会删除。
- 但当前静态路由名和 account-portal 路径已经对外暴露，仍需谨慎处理。

建议状态：

- `transition`

## 4. 运行时 façade 兼容导出

### 4.1 `frontend/src/router/guards/beforeEach.ts`

- `114` 仍导出 `refreshUserAccessAndMenus()`
- `119` 仍导出 `refreshUserMenus()`
- `123` 仍导出 `refreshCurrentUserInfoContext(...)`
- 这些函数内部已退化为调用 runtime 层，但文件位置仍停留在旧 guard 入口。

### 4.2 `frontend/src/router/index.ts`

- `7-16` 继续把上述函数从 router 根入口 re-export 出去。

当前用途判断：

- 这是典型的“过渡性 façade”，价值在于不打断旧 import path。
- 一旦消费方完成迁移，应优先删除，因为它会模糊 `router/runtime/*` 已经建立的单一主链。

建议状态：

- `delete`

## 5. worktab / app-scope 历史迁移残留

### 5.1 `frontend/src/store/modules/user.ts`

- `198`、`221`、`303`、`320` 使用 `StorageConfig.LAST_USER_ID_KEY`
- `302` 定义 `checkAndClearWorktabs()`
- `204`、`316` 在切用户场景下 `useWorktabStore().clearAll()`

### 5.2 `frontend/src/utils/storage/storage-config.ts`

- `45` 定义 `LAST_USER_ID_KEY = 'sys-last-user-id'`

### 5.3 `frontend/src/store/modules/worktab.ts`

- `48` 依赖 `APP_SCOPE_GLOBAL`、`normalizeAppScopeKey`
- `103` `switchScope(...)`
- `117`、`125` 仍保留 scope 兼容切换

### 5.4 `frontend/src/utils/app-scope.ts`

- `15` `readAppContextStoreFallback()`
- `43` 在拿不到最新上下文时继续走 fallback 读法

当前用途判断：

- 这批代码不是纯粹的“旧 API 壳层”，而是当前 worktab/app-scope 迁移尚未完全收口时的保护逻辑。
- 现在删除容易引入跨账号、跨 app 下 tab 污染或 scope 丢失。

建议状态：

- `keep`

## 6. 第一批可直接动手的对象

优先级从高到低：

1. 运行时 façade 兼容导出
   - 目标：把消费方从 `router/index.ts` / `beforeEach.ts` 迁到 `router/runtime/*`
   - 风险：低
2. 小型桥接文件
   - 目标：优先审视 `frontend/src/api/workspace.ts` 这类调用面较窄的旧类型适配器
   - 风险：中
3. 薄壳路径页面
   - 目标：确认 account-portal 路由名和外部回跳是否必须，再决定是否可合并
   - 风险：中

不建议在 P2-B 第一批直接删除的对象：

- `frontend/src/api/system-manage/_shared.ts`
- `frontend/src/utils/http/index.ts`
- `frontend/src/api/auth.ts`
- `frontend/src/store/modules/worktab.ts`
- `frontend/src/utils/app-scope.ts`

## 7. P2B-2 直接承接

下一步在代码中补 `// @compat-status: keep|delete|transition`，建议优先标记：

- `frontend/src/utils/http/index.ts` -> `transition`
- `frontend/src/api/system-manage/_shared.ts` -> `transition`
- `frontend/src/api/auth.ts` -> `transition`
- `frontend/src/api/workspace.ts` -> `transition`
- `frontend/src/router/routes/staticRoutes.ts` -> `transition`
- `frontend/src/router/guards/beforeEach.ts` -> `delete`（仅 façade 导出段）
- `frontend/src/router/index.ts` -> `delete`（仅兼容 re-export 段）
- `frontend/src/store/modules/user.ts` -> `keep`（仅 worktab 迁移保护段）
- `frontend/src/store/modules/worktab.ts` -> `keep`
- `frontend/src/utils/app-scope.ts` -> `keep`

## 8. 已完成的第一批清理

已落地：

- 已在以下文件补 `@compat-status` 注释：
  - `frontend/src/utils/http/index.ts`
  - `frontend/src/api/system-manage/_shared.ts`
  - `frontend/src/api/auth.ts`
  - `frontend/src/api/workspace.ts`
  - `frontend/src/router/routes/staticRoutes.ts`
  - `frontend/src/store/modules/user.ts`
  - `frontend/src/store/modules/worktab.ts`
  - `frontend/src/utils/app-scope.ts`
  - `frontend/src/views/account-portal/auth/login/index.vue`
  - `frontend/src/views/account-portal/auth/register/index.vue`
  - `frontend/src/views/account-portal/auth/forget-password/index.vue`

- 已删除第一批运行时 façade 兼容出口：
  - `frontend/src/router/guards/beforeEach.ts` 中旧的
    - `refreshUserAccessAndMenus`
    - `refreshUserMenus`
    - `refreshCurrentUserInfoContext`
  - `frontend/src/router/index.ts` 中对应的兼容 re-export

- 已完成调用方迁移：
  - `frontend/src/router/runtime/app-context.ts` -> 直接依赖 `router/runtime/navigation`
  - `frontend/src/composables/auth-flow/shared.ts` -> 直接依赖 `router/runtime/navigation`
  - `frontend/src/components/core/layouts/art-header-bar/widget/ArtCollaborationWorkspaceSwitcher.vue`
  - `frontend/src/components/core/layouts/art-header-bar/widget/ArtUserMenu.vue`
  - `frontend/src/views/system/role/index.vue`

验证结果：

- `frontend` 下 `pnpm exec vue-tsc --noEmit` 通过。
- 针对本批文件执行 `eslint` 无 error，仅保留仓库既有的 `any` warning。

## 9. 过渡期兼容代码迁移时间表

### A. 本轮收口整理内继续推进

1. `frontend/src/api/workspace.ts`
   - 状态：`transition`
   - 目标：把 `normalizeWorkspace()` 替换为基于 v5 schema 的显式前端类型，减少 `Api.SystemManage.WorkspaceItem` 依赖。
   - 触发条件：workspace 相关页面与 store 本轮内若还要继续调整，可顺手完成。
   - 原因：调用面相对窄，迁移风险低于 `_shared.ts`。

2. `frontend/src/views/account-portal/auth/*/index.vue`
   - 状态：`transition`
   - 目标：确认是否必须保留 account-portal 路由壳层；若仅为复用路径名，可考虑直接把静态路由指向 `views/auth/*`。
   - 触发条件：完成中心化登录 / 注册入口回跳验证后。
   - 原因：这组文件本身无业务逻辑，适合在验证明确后快速收口。

### B. 下一轮功能迭代时随改随迁

1. `frontend/src/api/auth.ts`
   - 状态：`transition`
   - 目标：把 `Api.Auth.LoginResponse`、`Api.Auth.UserInfo` 的旧契约逐步替换成 v5 原生类型或新的领域类型。
   - 触发条件：下次再修改登录、注册、callback、session/store 任一主链时。
   - 原因：认证主链刚完成收口，不适合在没有业务驱动的情况下继续大改消费面。

2. `frontend/src/router/routes/staticRoutes.ts`
   - 状态：`transition`
   - 目标：删除 `/auth/* -> /account/auth/*` 的旧路径重定向。
   - 触发条件：确认中心化登录平台、浏览器历史入口、外部文档链接都已切到新路径。
   - 原因：这类兼容依赖外部流量，不宜仅凭代码内无引用就删除。

### C. 需要专项任务处理

1. `frontend/src/api/system-manage/_shared.ts`
   - 状态：`transition`
   - 目标：按领域拆掉旧 `Api.SystemManage.*` 桥接，逐步迁到 v5 schema 或新的 domain DTO。
   - 触发条件：需要单独建立“系统管理 API 类型去桥接化”任务，至少拆成菜单、页面、权限、用户、消息几个子块。
   - 原因：调用面太广，直接动会波及多个页面、store 和工具函数。

2. `frontend/src/utils/http/index.ts`
   - 状态：`transition`
   - 目标：减少 axios 入口调用方，最终统一到 `v5Client` 或新的领域 API 封装。
   - 触发条件：当主要 API 封装切完 v5，且 request retry / cache / file upload 等能力有明确替代实现后。
   - 原因：现在删入口会影响仍在使用 `defHttp` 的模块，不适合做一次性替换。

### D. 长期保留，待更高层设计定稿后再议

1. `frontend/src/store/modules/user.ts` 中 `checkAndClearWorktabs`
2. `frontend/src/store/modules/worktab.ts`
3. `frontend/src/utils/app-scope.ts`

保留条件：

- 只要系统仍允许跨账号切换、跨 app 切换、scope 恢复，就继续保留这批保护逻辑。
- 只有在 worktab 存储模型和 app-context 持久化模型都重构完成后，才评估是否可删 fallback。
