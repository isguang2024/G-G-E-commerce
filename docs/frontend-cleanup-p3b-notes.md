# 前端收口 P3-B：领域化目录结构方案

更新时间：2026-04-12

## 目标

- 把前端运行时主链从“按技术层散落”收成“按业务领域归并”。
- 为 `P3B-2 ~ P3B-5` 提供可直接执行的目录目标、文件归属和迁移顺序。
- 保持页面层与通用组件层稳定，避免结构整理影响当前业务回归网。

## 不动的层

以下目录在 P3-B 中保持原位，只允许内部 import 调整，不做整体搬迁：

- `frontend/src/components/`
- `frontend/src/views/`
- `frontend/src/locales/`
- `frontend/src/api/v5/`
- `frontend/src/utils/http/`
- `frontend/src/config/`
- `frontend/src/plugins/`

原因：

- `components/views/locales` 已天然按展示层组织，移动收益低、回归成本高。
- `api/v5` 与 `utils/http` 是跨领域基础设施，不属于某个单独业务域。
- `config/plugins` 是应用壳层，不适合卷入领域拆分。

## 当前 before：真实聚合点

| 当前聚合点 | 代表文件 | 当前问题 | 目标领域 |
| --- | --- | --- | --- |
| 认证流程 | `api/auth.ts`、`composables/auth-flow/*`、`router/runtime/session.ts`、`store/modules/user.ts` | 登录、session 恢复、登出逻辑分散在 api/store/runtime/composable 四层 | `domains/auth` |
| App / 空间上下文 | `store/modules/app-context.ts`、`store/modules/menu-space.ts`、`router/runtime/app-context.ts`、`utils/app-scope.ts`、`hooks/business/useManagedAppScope.ts`、`api/system-manage/app.ts` | `appKey`、`spaceKey`、host binding、切 app 流程跨 store/runtime/api/hook | `domains/app-runtime` |
| 运行时导航 | `router/core/*`、`router/runtime/navigation.ts`、`store/modules/menu.ts`、`store/modules/worktab.ts`、`utils/navigation/*` | manifest 解析、动态路由注册、worktab、space landing 混在多个目录 | `domains/navigation` |
| 系统治理 API | `api/system-manage/*`、`utils/permission/*`、`types/api/api.d.ts` 中的 `Api.SystemManage.*` | “应用/菜单空间/菜单/页面/权限/角色/用户”统一堆在 `system-manage`，边界过粗 | `domains/governance` |
| 协作空间上下文 | `store/modules/collaboration-workspace.ts`、`store/modules/workspace.ts`、`api/collaboration-workspace.ts`、`api/workspace.ts` | 既服务认证恢复，又服务独立业务页，当前不适合立即硬搬 | 暂留原位，先作为 `auth/app-runtime` 依赖 |

## 目标 after：目录骨架

```text
frontend/src/
  domains/
    auth/
      api/
      composables/
      runtime/
      store/
      types/
      utils/
      index.ts
    app-runtime/
      api/
      hooks/
      runtime/
      store/
      types/
      utils/
      index.ts
    navigation/
      runtime/
      router-core/
      store/
      types/
      utils/
      index.ts
    governance/
      api/
        app/
        menu/
        page/
        permission/
        role/
        user/
        register/
        api-endpoint/
      types/
      utils/
      index.ts
  api/
    v5/
  components/
  views/
  locales/
  router/
    index.ts
    guards/
    routes/
  store/
    index.ts
    modules/        # 过渡期 re-export，最终清空或删除
  utils/
    http/
```

## 领域职责边界

### 1. `domains/auth`

只负责：

- 登录、注册、回调、登出
- token / refresh token / session restore
- 当前登录用户身份写入
- 认证相关 URL / centralized login 工具

不负责：

- 动态路由注册
- app / menu-space 切换
- 系统治理页的 CRUD API

建议承接文件：

| 当前文件 | 目标位置 |
| --- | --- |
| `frontend/src/api/auth.ts` | `frontend/src/domains/auth/api/auth.ts` |
| `frontend/src/composables/auth-flow/shared.ts` | `frontend/src/domains/auth/composables/shared.ts` |
| `frontend/src/composables/auth-flow/useLoginFlow.ts` | `frontend/src/domains/auth/composables/useLoginFlow.ts` |
| `frontend/src/composables/auth-flow/useRegisterFlow.ts` | `frontend/src/domains/auth/composables/useRegisterFlow.ts` |
| `frontend/src/composables/auth-flow/useCallbackFlow.ts` | `frontend/src/domains/auth/composables/useCallbackFlow.ts` |
| `frontend/src/composables/auth-flow/useLogoutFlow.ts` | `frontend/src/domains/auth/composables/useLogoutFlow.ts` |
| `frontend/src/router/runtime/session.ts` | `frontend/src/domains/auth/runtime/session.ts` |
| `frontend/src/utils/auth/centralized-login.ts` | `frontend/src/domains/auth/utils/centralized-login.ts` |
| `frontend/src/store/modules/user.ts` 中 session/login/logout 相关逻辑 | `frontend/src/domains/auth/store/session.ts` |

备注：

- `store/modules/user.ts` 不建议第一刀整文件搬走，应先提取 session 能力到 `domains/auth/store/session.ts`，旧文件保留装配和兼容导出。

### 2. `domains/app-runtime`

只负责：

- `appKey` / `managedAppKey` / runtime app profile
- menu-space 配置、host binding、默认空间解析
- app 切换和 runtime app 预热
- app scope 读写

不负责：

- 动态路由表构建
- 角色 / 权限 / 页面管理 API

建议承接文件：

| 当前文件 | 目标位置 |
| --- | --- |
| `frontend/src/router/runtime/app-context.ts` | `frontend/src/domains/app-runtime/runtime/app-context.ts` |
| `frontend/src/store/modules/app-context.ts` | `frontend/src/domains/app-runtime/store/app-context.ts` |
| `frontend/src/store/modules/menu-space.ts` | `frontend/src/domains/app-runtime/store/menu-space.ts` |
| `frontend/src/hooks/business/managed-app-scope.ts` | `frontend/src/domains/app-runtime/hooks/managed-app-scope.ts` |
| `frontend/src/hooks/business/useManagedAppScope.ts` | `frontend/src/domains/app-runtime/hooks/useManagedAppScope.ts` |
| `frontend/src/utils/app-scope.ts` | `frontend/src/domains/app-runtime/utils/app-scope.ts` |
| `frontend/src/api/system-manage/app.ts` 中 app/menu-space/host-binding 能力 | `frontend/src/domains/app-runtime/api/app.ts` |

备注：

- `api/system-manage/app.ts` 同时承载“应用管理页 CRUD”和“runtime menu-space 配置”两类能力，P3-B 里应拆成运行时侧与治理侧两部分，避免一个文件跨两个领域。

### 3. `domains/navigation`

只负责：

- runtime manifest 拉取与校验
- menu tree / managed page → route records 转换
- route registry / iframe route / route validator
- menu store、homePath、worktab
- space landing path / route jump / managed page URL 解析

不负责：

- 登录态恢复
- app profile 与 host binding 解析
- 系统管理领域的 CRUD 正常化

建议承接文件：

| 当前文件 | 目标位置 |
| --- | --- |
| `frontend/src/router/runtime/navigation.ts` | `frontend/src/domains/navigation/runtime/navigation.ts` |
| `frontend/src/router/core/*` | `frontend/src/domains/navigation/router-core/*` |
| `frontend/src/store/modules/menu.ts` | `frontend/src/domains/navigation/store/menu.ts` |
| `frontend/src/store/modules/worktab.ts` | `frontend/src/domains/navigation/store/worktab.ts` |
| `frontend/src/utils/navigation/index.ts` | `frontend/src/domains/navigation/utils/index.ts` |
| `frontend/src/utils/navigation/jump.ts` | `frontend/src/domains/navigation/utils/jump.ts` |
| `frontend/src/utils/navigation/managed-page.ts` | `frontend/src/domains/navigation/utils/managed-page.ts` |
| `frontend/src/utils/navigation/menu-space.ts` | `frontend/src/domains/navigation/utils/menu-space.ts` |
| `frontend/src/utils/navigation/route.ts` | `frontend/src/domains/navigation/utils/route.ts` |
| `frontend/src/utils/navigation/worktab.ts` | `frontend/src/domains/navigation/utils/worktab.ts` |
| `frontend/src/types/router/index.ts` | `frontend/src/domains/navigation/types/router.ts` |

备注：

- `router/index.ts`、`router/guards/*`、`router/routes/staticRoutes.ts` 继续留在 `src/router/`，但应降级为壳层，只负责初始化和分发，不再保存业务逻辑。

### 4. `domains/governance`

只负责：

- 系统管理领域的 API 封装与 normalizer
- 应用、菜单空间、菜单、页面、权限、角色、用户、注册管理
- `Api.SystemManage.*` 到 v5 schema 的桥接过渡

不负责：

- 登录态恢复
- 运行时导航注册
- 全局 HTTP 客户端

建议承接文件：

| 当前文件 | 目标位置 |
| --- | --- |
| `frontend/src/api/system-manage/api-endpoint.ts` | `frontend/src/domains/governance/api/api-endpoint.ts` |
| `frontend/src/api/system-manage/menu.ts` | `frontend/src/domains/governance/api/menu/index.ts` |
| `frontend/src/api/system-manage/page.ts` 中治理页 CRUD 能力 | `frontend/src/domains/governance/api/page/index.ts` |
| `frontend/src/api/system-manage/permission.ts` | `frontend/src/domains/governance/api/permission/index.ts` |
| `frontend/src/api/system-manage/register.ts` | `frontend/src/domains/governance/api/register/index.ts` |
| `frontend/src/api/system-manage/role.ts` | `frontend/src/domains/governance/api/role/index.ts` |
| `frontend/src/api/system-manage/user.ts` | `frontend/src/domains/governance/api/user/index.ts` |
| `frontend/src/api/system-manage/_shared.ts` | `frontend/src/domains/governance/utils/legacy-normalizers.ts` |
| `frontend/src/utils/permission/action.ts` | `frontend/src/domains/governance/utils/permission-action.ts` |
| `frontend/src/utils/permission/menu.ts` | `frontend/src/domains/governance/utils/permission-menu.ts` |

备注：

- `api/system-manage/page.ts` 同时含 `/runtime/navigation` 读取，迁移时要把 runtime navigation 相关调用切回 `domains/navigation`，避免治理域反向依赖导航域。
- `api/system-manage/_shared.ts` 是最后迁移对象，应先把新代码从它剥离，再拆 legacy normalizer。

## 依赖方向

建议固定为：

```text
auth -> app-runtime -> navigation
governance -> (none of the above required)
navigation -> auth, app-runtime
router shell -> auth, app-runtime, navigation
views/components -> domains/*
```

约束：

- `governance` 不依赖 `navigation/runtime`，最多只共享 `api/v5` 与纯类型。
- `auth` 不直接操作 `router/core/*`，只触发 `navigation` 的公共入口。
- `app-runtime` 不感知 `views`，只暴露 store/runtime/hook。

## 迁移顺序

### 第 0 步：搭骨架，不改行为

- 新建 `frontend/src/domains/{auth,app-runtime,navigation,governance}/`
- 每个领域先落 `index.ts`
- `src/router` 与 `src/store/modules` 先保留原路径，允许短期 re-export

### 第 1 步：先迁 `auth`

原因：

- 文件量可控
- 已在 P1/P2 把 auth-flow 与 session-runtime 收过一轮
- 对动态路由核心结构影响最小

执行顺序：

1. `composables/auth-flow/*`
2. `router/runtime/session.ts`
3. `api/auth.ts`
4. `utils/auth/*`
5. `store/modules/user.ts` 中的 session slice 提取

### 第 2 步：再迁 `app-runtime`

原因：

- `navigation` 依赖 `app-context/menu-space`
- 先收好 app/space 主链，后迁导航可避免重复改 import

执行顺序：

1. `store/modules/app-context.ts`
2. `store/modules/menu-space.ts`
3. `router/runtime/app-context.ts`
4. `hooks/business/*ManagedAppScope*`
5. `utils/app-scope.ts`
6. `api/system-manage/app.ts` 的 runtime 部分

### 第 3 步：迁 `navigation`

原因：

- 依赖 auth 与 app-runtime 的稳定入口
- 改动面最大，放在中后段更稳

执行顺序：

1. `utils/navigation/*`
2. `router/core/*`
3. `router/runtime/navigation.ts`
4. `store/modules/menu.ts`
5. `store/modules/worktab.ts`
6. `router/guards/beforeEach.ts` 改为壳层调用

### 第 4 步：最后迁 `governance`

原因：

- 量最大
- 与 legacy `Api.SystemManage.*` 契约耦合最深
- 需要在前三个领域稳定后再做清拆

执行顺序：

1. `permission/role/user/register/api-endpoint`
2. `menu/page`
3. `app.ts` 中治理页 CRUD 部分
4. `_shared.ts` 拆成 `legacy-normalizers + 小型 normalizer`
5. `api/system-manage/index.ts` 收成兼容 re-export

### 第 5 步：收口兼容层

- `store/modules/*` 改为仅 re-export 或删除
- `router/runtime/*` 改为仅 re-export 或删除
- `api/system-manage/*` 改为兼容出口或删除
- 全仓替换 import 到 `domains/*`

## 过渡期兼容策略

- `frontend/src/store/modules/*` 在 P3-B 中允许短期保留，但只做 re-export，不再新增业务逻辑。
- `frontend/src/router/runtime/*` 在 `domains/navigation` / `domains/auth` 稳定后，退化为兼容入口。
- `frontend/src/api/system-manage/index.ts` 作为最后删除的 barrel，保证视图层在迁移中不断编。
- 对 `types/api/api.d.ts` 的旧 `Api.SystemManage.*` 契约，不在 P3-B 第一轮强制删除，只限制新文件不继续扩大依赖面。

## 对应任务拆解

| 节点 | 实施范围 |
| --- | --- |
| `P3B-1` | 定方案、定边界、定迁移顺序 |
| `P3B-2` | `domains/auth` 首批落地 |
| `P3B-3` | `domains/app-runtime` 首批落地 |
| `P3B-4` | `domains/navigation` 首批落地 |
| `P3B-5` | `domains/governance` 首批落地 |
| `P3B-6` | 编译、import 清理、兼容出口裁剪 |

## 结论

P3-B 不应理解为“把所有文件搬进 `domains/`”，而是把运行时主链和系统治理主链从当前的技术层散落状态收成四个稳定领域：

- `auth`
- `app-runtime`
- `navigation`
- `governance`

其中：

- `auth` 与 `app-runtime` 是前置依赖
- `navigation` 是运行时消费层
- `governance` 是系统管理业务域

按这个顺序推进，可以把结构移动的风险压到最低，同时不给现有页面层和回归集制造大面积噪音。
