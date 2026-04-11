# 多 APP 混合承载基础设计

## 1. 文档目标

本文固定 `tsk_01KNYXJ3VCJFPYJ6DGDKAY` 根节点 1、2 的分析结论，作为后续 Phase A/B/C 设计与实现的前置真相源。

- 目标：收口当前仓库里 `App`、`AppHostBinding`、`MenuSpace`、`UIPage`、`account-portal` 试点的真实职责边界。
- 目标：明确多 APP 架构的术语、入口匹配优先级、默认首页与切换规则、实现链路纪律。
- 非目标：本文不直接引入新表/新字段，不替代 Phase A/B/C 的详细实现设计。

## 2. 当前现状盘点

### 2.1 APP 与入口绑定

- `apps` 是治理对象，表达 APP 的稳定身份、默认空间、空间模式和认证模式；它不直接描述某个页面或某个菜单。对应模型见 `backend/internal/modules/system/models/app.go`。
- `app_host_bindings` 是 Level 1 入口解析规则，只负责把 `host + path` 解析成某个 APP。它不负责菜单空间权限，也不负责页面注册。对应模型见 `backend/internal/modules/system/models/app.go`。
- 当前 APP 入口匹配类型只有四种：`host_exact`、`host_suffix`、`path_prefix`、`host_and_path`。后续部署拓扑必须由这些匹配类型配合 `App.AuthMode`、`App.SpaceMode` 表达，不引入 `deployment_type` 枚举。

### 2.2 MenuSpace 与 UIPage

- `menu_spaces` 是某个 APP 内的导航视图边界，承担默认首页、空间访问模式、按 host 切换导航目标等职责；它不参与权限公式，只是承载导航组织。对应模型见 `backend/internal/modules/system/models/menu_space.go`。
- `menu_space_entry_bindings` 是 Level 2 入口解析规则，只在 APP 已确定后，把请求进一步解析到某个菜单空间；单空间 APP 直接短路到 `App.DefaultSpaceKey`。对应实现见 `backend/internal/modules/system/app/service.go`。
- `ui_pages` 是页面定义层，表达路由、组件、访问模式、页面归属与展示范围。`AppKey` 代表页面归属到哪个 APP，`SpaceKey` 与 `VisibilityScope` 决定页面暴露到哪个空间。对应模型见 `backend/internal/modules/system/models/model.go` 与 `backend/internal/modules/system/page/service.go`。
- 页面类型当前只有四种：`inner`、`standalone`、`group`、`display_group`。其中：
  - `inner` 必须挂到菜单或上级页面。
  - `standalone` 不能挂到菜单或上级页面。
  - `group`、`display_group` 是无实际组件路由的结构节点。
- 页面可见性当前只有三种有效语义：`inherit`、`app`、`spaces`。内页强制 `inherit`；无父级的独立页/分组页默认提升为 `app`，仅显式绑定时才落到 `spaces`。
- 页面访问模式当前只有四种有效值：`inherit`、`public`、`jwt`、`permission`。只有 `permission` 模式允许持有 `permission_key`。

### 2.3 account-portal 试点真实链路

- 后端入口解析链是：`AppContext` 中间件读取请求 `host/path`，调用 `ResolveAppEntry` 解析 APP，再调用 `ResolveMenuSpaceEntry` 或 `ResolveCurrentSpaceKey` 解析空间，最后把 `app_key`、`space_key`、`resolved_by` 写入请求上下文。实现见 `backend/internal/api/middleware/app_context.go`。
- `ResolveAppEntry` 当前优先级是：显式 `requestedAppKey` 命中且 APP 存在 → 启用中的 `app_host_bindings` 按具体度命中 → 默认 APP → `fallback_default`。实现见 `backend/internal/modules/system/app/service.go`。
- `account-portal` 已通过 seed 落了 `/account` 的 `path_prefix` 入口绑定，本地单域名场景默认命中该 APP；同时保留 `host_exact` 的子域名示例用于后续独立域名部署。seed 见 `backend/internal/pkg/permissionseed/register_seed.go`。
- 前端未登录路由守卫会先尝试注册公开 runtime 页面；`/account/*` 路径会被直接推断为 `account-portal`，并通过公开 runtime 路径完成动态路由注入。实现见 `frontend/src/router/guards/beforeEach.ts`。
- 公开认证页在路由转换时会保持绝对路径，不再套进主后台的一级 Layout，避免多个 `/account` 容器互相遮挡。实现见 `frontend/src/router/core/RouteTransformer.ts`。

### 2.4 account-portal 与主壳当前耦合点

- 登录页仍直接依赖 `userStore`、`collaborationWorkspaceStore`、`menuSpaceStore`，并在登录成功后主动刷新工作空间、菜单空间与路由状态；说明当前 `account-portal` 仍共享主壳的用户上下文与空间解析能力。见 `frontend/src/views/auth/login/index.vue`。
- 登录页还直接调用 `resetRouterState`，说明公开认证流会修改主壳的动态路由注册状态，而不是完全独立运行。见 `frontend/src/views/auth/login/index.vue`。
- 注册页依赖 `fetchRegisterContext(window.location.host, window.location.pathname)` 从当前 URL 推断入口策略，这意味着 `account-portal` 已经按“入口驱动页面上下文”工作，但它仍沿用同一个前端工程与同一套用户 store。见 `frontend/src/views/auth/register/index.vue`。

### 2.5 APP 级隔离当前状态

- `menuSpaceStore` 已经按 `appKey` 维护 `menuSpaceConfigMap`、`overrideSpaceKeyMap`、`loadedAppKeys`，菜单空间配置具备 APP 维度缓存与切换能力。见 `frontend/src/store/modules/menu-space.ts`。
- `useManagedAppScope` 已经把“当前管理 APP”写进 `managed-app:*` 键，具备局部 APP 作用域记忆。见 `frontend/src/hooks/business/managed-app-scope.ts` 与 `frontend/src/hooks/business/useManagedAppScope.ts`。
- 但全局持久化尚未完全 APP 隔离：`app-context`、`menu-space`、`iframeRoutes`、`user`、`setting`、`worktab` 仍使用固定 localStorage/sessionStorage key，因此当前只能说“局部隔离已开始，全局壳层仍共享状态”。

## 3. 核心术语与配置组合模型

### 3.1 术语

| 术语 | 角色 |
|---|---|
| App | 治理对象，定义一个独立应用的标识、默认空间、认证模式、空间模式与品牌/能力扩展位 |
| AppHostBinding | APP 入口绑定，负责把请求解析到 APP |
| MenuSpace | APP 内的导航视图边界，负责默认首页与空间访问策略 |
| MenuSpaceEntryBinding | 菜单空间入口绑定，负责把请求解析到具体空间 |
| UIPage | 页面定义，负责路由、组件、访问模式、空间暴露规则 |
| account-portal | 当前公共认证中心试点 APP，承载登录/注册/找回密码等公开认证页 |

### 3.2 部署拓扑的配置组合心智模型

不引入 `deployment_type`。对外沟通时仅使用“配置组合”表达拓扑：

| 组合 | 沟通语义 | 说明 |
|---|---|---|
| `path_prefix + inherit_host + single/multi` | 嵌入式 / 同域挂载 | 典型试点就是 `account-portal` 的 `/account/*` |
| `host_exact + shared_cookie + multi` | 独立域名 + 共享会话 | 适合 APP 子域名独立部署、同登录态切换 |
| `host_exact + centralized_login + multi` | 独立域名 + 认证中心 | 适合后续 Phase C 的中央认证流 |

说明：

- `MatchType` 决定“请求如何命中 APP”。
- `AuthMode` 决定“命中 APP 之后如何处理登录态”。
- `SpaceMode` 决定“APP 内部是否需要二级菜单空间解析”。

## 4. 入口匹配、首页与切换规则

### 4.1 APP 入口匹配优先级

APP 入口解析固定为：

1. 显式 `requestedAppKey`，且该 APP 实际存在。
2. `app_host_bindings` 命中，命中规则按 `PatternSpecificity + Priority*10` 排序。
3. 默认 APP。
4. `fallback_default`。

约束：

- `path_prefix` 与 `host_and_path` 必须同时看 path。
- `host_exact` 与 `host_suffix` 主要看 host。
- 入口匹配只决定 APP，不越权决定页面或权限。

### 4.2 空间解析优先级

空间解析固定为：

1. 如果 APP 是多空间，先尝试 `menu_space_entry_bindings`。
2. 仅当命中空间且当前用户有权访问时，才接受该入口空间。
3. 否则回退到 `ResolveCurrentSpaceKey` 的常规解析。
4. 再失败则落回 APP 默认空间。

### 4.3 redirect、默认首页与 APP 切换

- 登录页会反复解码 `redirect` 参数，但会拒绝把用户再次导回 `/auth/login` 或 `/account/auth/login`，防止登录页自循环。
- 默认首页优先级为：当前空间显式配置首页 → 默认空间首页 → `/workspace/inbox` → `/dashboard/console` → 第一个可用路径。
- 跨空间/跨 host 导航时，`menuSpaceStore.resolveSpaceNavigationTarget()` 会根据 host binding 决定使用前端 router 还是整页 `location` 跳转。
- 公开认证页保持绝对路径，不应被 APP 内普通 Layout 包裹；这是后续多 APP 并存时必须保留的规则。

## 5. 当前必须遵守的架构纪律

### 5.1 模型与数据链路

- 模型变更必须整链完成：`Model -> Migration -> Seed/Ensure -> build/test`。
- 默认数据走 seed / ensure，不把长期默认状态反复写进迁移链。
- 新模块、新表、新接口必须显式回答：是否带 `tenant_id`、是否在仓储层强制过滤。

### 5.2 API 链路

- API 一律 OpenAPI-first：先改 `backend/api/openapi/`，再 bundle、ogen、前端 `pnpm run gen:api`，最后修 handler/service/frontend 调用。
- 禁止手改 `backend/api/gen/` 与 `frontend/src/api/v5/` 生成产物。
- 权限判断只走 `backend/internal/pkg/permission/evaluator`。

### 5.3 前后端联动链路

- 后端契约一旦变更，同一任务内必须完成后端实现、前端生成物刷新、前端调用适配与至少一次联编校验。
- 前端 runtime 改动必须串行检查 `store -> router -> component`，完成标准至少包含 `pnpm exec vue-tsc --noEmit`；若涉及构建行为，继续补 `pnpm build`。
- 配置/绑定变更必须同步检查 seed、中间件、路由守卫、runtime 解析逻辑，禁止只改后台表单或只改数据库。

### 5.4 后续实现节点 instruction 规范

后续每个实现类节点的 `instruction` 至少要包含：

- 变更范围：明确到模块、文件或生成链。
- 强制链路：引用本节中适用的模型/API/runtime 链路。
- 完成标准：至少写清需要跑的 build/test/gen 命令。
- 禁止拆分声明：不能把同一条链路拆成“先改模型，后补 seed/前端”的半成品任务。

推荐模板：

```md
变更范围：model + migration + seed + openapi + handler + frontend runtime

强制链路：
1. 先改 OpenAPI / model
2. 再刷新生成物
3. 再改 handler / service / frontend
4. 最后跑 go test 与 vue-tsc

完成标准：
- `go test ./internal/api/handlers -count=1`
- `pnpm exec vue-tsc --noEmit`

禁止拆分：
- 不允许只落一半链路后进入下游开发
```

## 6. 本轮明确不做

本轮基础设计明确排除以下事项：

- 不引入微前端框架，如 `qiankun`、`wujie`。
- 不在本阶段引入 WebSocket/实时通信新架构。
- 不把多 APP 方案扩展成多前端工程或第二套后台。
- 不在本阶段展开 i18n 多语言、插件化、跨服务 RPC、GraphQL 等额外议题。
- 不引入 `deployment_type` 这类把配置组合重新固化成枚举的快捷字段。

## 7. 后续 Phase 的直接输入

基于当前仓库，后续 Phase A/B/C 设计与实现都应直接继承以下结论：

- APP 是治理对象，入口绑定与页面定义必须继续分层，不可把入口规则塞回页面模型。
- `account-portal` 已经具备“公共认证中心”的雏形，但仍共享主壳 store 与动态路由状态；Phase C 的关键不是再造一套登录页，而是把它从主壳共享状态里逐步解耦。
- APP 级持久化隔离目前只完成了局部能力，Phase A runtime 改造应优先补齐全局状态命名空间与路由缓存清理。
- 设计沟通必须坚持“配置组合表达拓扑”，避免回到枚举驱动的僵化建模。

## 8. Phase A 设计提案

本节开始是设计提案，不是当前已实现行为。

### 8.1 App 主模型扩展字段

保持 `App.SpaceMode`、`App.AuthMode` 顶层字段不变，只补“入口 URL + 能力描述”：

| 字段 | 类型 | 作用 |
|---|---|---|
| `frontend_entry_url` | `varchar` | 面向用户的前端入口地址；path_prefix 场景允许存相对路径，如 `/account` |
| `backend_entry_url` | `varchar` | 当前 APP 对应的 API/Gateway 入口地址；同域场景可为空表示继承当前 host |
| `health_check_url` | `varchar` | 后台 dry-run、探活与治理页展示使用 |
| `capabilities` | `jsonb` | 运行能力声明，驱动 runtime、后台治理与前端切换 |

建议的 `capabilities` 结构：

```json
{
  "routing": {
    "entry_mode": "path_prefix",
    "route_prefix": "/account",
    "supports_public_runtime": true
  },
  "runtime": {
    "kind": "local",
    "supports_dynamic_routes": true,
    "supports_worktab": false
  },
  "navigation": {
    "supports_multi_space": false,
    "default_landing_mode": "menu_space"
  },
  "integration": {
    "supports_app_switch": true,
    "supports_broadcast_channel": false
  }
}
```

设计原则：

- 不把 `deployment_type` 再包装回顶层枚举。
- `frontend_entry_url`、`backend_entry_url`、`health_check_url` 是“运维入口”字段，不替代 `app_host_bindings` 的解析规则。
- `capabilities` 只放会影响 runtime/治理决策的声明，不复制 `space_mode`、`auth_mode` 这类已有顶层字段。

### 8.2 capabilities / auth / branding / version 的分层

为了避免 `capabilities` 继续膨胀，建议分四层：

| 层 | 存放位置 | 建议内容 |
|---|---|---|
| capabilities | `apps.capabilities` | 路由能力、动态路由、工作台、跨 APP 通信、空间能力等运行时开关 |
| auth | 顶层 `auth_mode` + `meta.auth` | 登录页路径、回跳白名单策略、是否允许公共入口 |
| branding | `meta.branding` | APP 名称、副标题、logo、主题 token、认证页布局风格 |
| version | `meta.version` | `frontend_manifest_version`、`backend_contract_version`、`min_supported_platform_version` |

建议结构：

```json
{
  "auth": {
    "login_path": "/account/auth/login",
    "logout_path": "/account/auth/login",
    "allow_public_entry": true
  },
  "branding": {
    "display_name": "账户中心",
    "theme_key": "account-portal",
    "auth_layout": "split"
  },
  "version": {
    "frontend_manifest_version": "v1",
    "backend_contract_version": "v1",
    "min_supported_platform_version": "5.0"
  }
}
```

### 8.3 页面归属、代码归属、部署归属分离规则

后续必须显式分离三类归属：

| 维度 | 真相源 | 说明 |
|---|---|---|
| 页面归属 | `ui_pages.app_key` | 页面属于哪个 APP，由哪个 APP 的 runtime 注册 |
| 代码归属 | `ui_pages.component` + 前端组件目录 | 页面组件由哪个前端包/目录提供 |
| 部署归属 | `app_host_bindings` + `frontend_entry_url/backend_entry_url` | 用户通过什么 host/path 命中 APP |

约束：

- `ui_pages` 不承载部署语义，不保存“这个页面发布到哪个域名”。
- `app_host_bindings` 不承载页面语义，不保存具体页面组件。
- `menu_space_entry_bindings` 只决定“哪个空间接住当前请求”，不复制页面定义。
- 仅 `PageSpaceBinding` 继续用于“少量无父级独立页”的空间暴露控制，不把它升级成第二套页面路由系统。

### 8.4 菜单空间继承、默认首页与入口绑定规则

建议固定以下规则：

1. APP 命中后，先确定 `App.DefaultSpaceKey`。
2. 如果 `App.SpaceMode=single`，直接使用默认空间。
3. 如果 `App.SpaceMode=multi`，且命中了 `MenuSpaceEntryBinding` 且用户有权访问，则覆盖当前空间。
4. `MenuSpace.DefaultHomePath` 仍是该空间首页唯一真相源。
5. `AppHostBinding.DefaultSpaceKey` 仅作为入口兜底，不覆盖已经解析成功的空间首页规则。

因此：

- “入口绑定”解决的是首个落点空间。
- “默认首页”解决的是进入空间后的首个页面。
- “页面绑定”解决的是该页面能否暴露到该空间。

三者不能再混写到同一个字段里。

### 8.5 前端多 APP runtime 切换

#### 8.5.1 app context / store / worktab 按 APP 隔离

建议把前端状态分成两层：

- 全局共享层：`user`、认证 token、语言、主题、当前 workspace。
- APP 隔离层：`app-context`、`menu-space`、`worktab`、`iframeRoutes`、搜索历史、运行时缓存。

建议补齐以下命名空间：

| 模块 | 当前状态 | Phase A 建议 |
|---|---|---|
| `app-context` | 固定 key `appContextStore` | 改为 `app-context:{appKey}` 或全局主索引 + app 子键 |
| `menu-space` | 固定 key `menu-space` | 改为 `menu-space:{appKey}` |
| `worktab` | 固定 key `worktab` | 改为 `worktab:{appKey}` |
| `iframeRoutes` | 固定 key `iframeRoutes` | 改为 `iframeRoutes:{appKey}` |
| `managed-app:*` | 已局部隔离 | 保留并继续复用 |

`appContextStore` 建议增加：

- `currentRuntimeAppKey`
- `lastRuntimeAppKey`
- `switchEpoch`
- `runtimeSessionKey`

其中 `switchEpoch` 用来驱动路由缓存、工作台和 iframe 缓存的统一失效。

#### 8.5.2 动态路由、布局与缓存按 APP 切换

建议 `RouteRegistry` 增加“当前已注册 APP”概念：

- 当 `runtimeAppKey` 变化时，先 `unregister`，再按新 APP 的 manifest 注册。
- `keepAliveExclude` 与 `IframeRouteManager` 都按 `appKey` 分区。
- `validateWorktabs` 只校验当前 APP 的标签，不在 APP 切换时拿旧 APP 路由做有效性判断。

布局建议：

- `account-portal` 这类公开认证 APP 继续走“绝对路径 + 轻壳”。
- `platform-admin` 与后续业务 APP 走“管理壳 + 动态菜单 + worktab”。
- 是否展示 `worktab`、侧边菜单、空间徽标，由 `capabilities.runtime` 驱动，而不是写死在全局配置里。

#### 8.5.3 统一应用目录与 APP 切换入口

建议把“应用目录”做成独立入口，而不是散落在各处快捷入口中：

- Header 保留一个统一 `AppSwitcher / Application Directory` 入口。
- `fast-enter` 继续作为快捷方式，但不再承担唯一 APP 切换职责。
- 应用目录只展示当前用户有权访问且状态正常的 APP，排序优先级建议为：当前 APP、默认 APP、最近访问 APP、其他 APP。

#### 8.5.4 壳层保留与重置策略

建议切 APP 时：

- 保留：用户信息、token、语言、主题、当前 workspace。
- 软重置：`menu-space` 当前空间、动态路由注册、worktab、iframeRoutes、首页缓存。
- 强重置：若切到 `supports_worktab=false` 或 `supports_dynamic_routes=false` 的 APP，则直接清空 worktab 与 keepAlive 状态。

### 8.6 前端构建与存储隔离

#### 8.6.1 按 APP 的代码分割与懒加载

建议后续组件路径按 APP 形成明确边界：

- `account-portal/*` 组件优先收拢到独立目录。
- `platform-admin/*` 继续承载管理壳与系统治理页面。
- 路由组件按 `appKey` 维度设置稳定 chunk 名，避免一个 APP 变动导致另一个 APP 首屏缓存抖动。

#### 8.6.2 localStorage / sessionStorage 命名空间

建议统一命名格式：

```text
sys-v{version}:app:{appKey}:{storeId}
```

这样可以同时满足：

- 版本升级迁移。
- APP 隔离。
- 后续灰度或多入口环境的冲突规避。

### 8.7 错误边界与 APP 间通信

#### 8.7.1 APP 级错误边界

建议在 APP 壳层外再加一层按 `appKey` 划分的 Error Boundary：

- 公开认证 APP 崩溃时，只回落到认证页错误态，不污染主后台壳层。
- 管理 APP 崩溃时，保留 header 与应用切换入口，允许用户切换到其他 APP 或重新加载当前 APP。

#### 8.7.2 同域场景下的跨 APP 通信

建议优先级：

1. 同 SPA 内切换：直接通过 router + app-context 切换，不做额外通信。
2. 同域多入口页签：优先 `BroadcastChannel`，用于登录完成、退出登录、主题切换等低频事件。
3. 跨域场景：改走重定向参数或服务端会话，不依赖浏览器内通信。

本阶段不建议引入复杂事件总线或微前端总线协议。

### 8.8 公共认证中心与跨 APP 入口体验

#### 8.8.1 account-portal 的职责上限

建议把 `account-portal` 的职责明确限制为：

- 登录
- 注册
- 找回密码
- 认证完成后的回跳
- 面向匿名用户的基础品牌展示

不建议让它继续承载：

- 管理后台菜单
- 复杂工作台
- 业务域页面

#### 8.8.2 品牌、布局与回跳策略

建议：

- 品牌信息从 `apps.meta.branding` 读取，避免登录页继续只依赖全局 `AppConfig.systemInfo`。
- 认证页布局允许按 APP 配 `split / centered / minimal` 三种样式，但组件骨架仍共用。
- 回跳目标必须同时校验 `target_app_key`、`target_navigation_space_key`、`target_home_path`，并与后端白名单策略一致。

## 9. Phase A 设计输出到实现的映射

本轮设计提案对应后续实现叶子节点的直接映射如下：

| 设计点 | 后续实现入口 |
|---|---|
| `frontend_entry_url/backend_entry_url/health_check_url/capabilities` | `4/1/1`、`4/2/1`、`4/4/1` |
| 页面/部署/空间三类归属分离 | `4/2/2`、`4/3/2` |
| APP 级 store/worktab/storage 隔离 | `4/3/1`、`4/3/4` |
| 公开认证 APP 的轻壳与错误边界 | `4/3/3`、`4/4/2` |
| 应用目录与切换入口 | `4/4/2` |

## 10. Phase B 设计提案（独立域名 + shared_cookie）

本节是 Phase B 设计输入，覆盖 `host_exact/host_suffix` 独立域名 APP 的入口、安全、会话、运维与前端接入策略。

### 10.1 入口与网关分发

#### 10.1.1 `host_exact/host_suffix` 的收益与边界

- 收益：
- 与 `path_prefix` 相比，独立域名可把 APP 的缓存、证书、限流、故障隔离到子域级别。
- 适合“同一账号体系 + 多业务域 APP”并行演进，不强制所有 APP 共享前端壳。
- 边界：
- `host_exact/host_suffix` 只解决“入口命中哪个 APP”，不解决页面归属和权限归属。
- `shared_cookie` 与 `centralized_login` 是认证策略，不是入口类型；两者不能写进 `MatchType`。
- `host_suffix` 只用于同组织子域聚合，禁止用于第三方不受控域。

#### 10.1.2 统一网关 vs 独立入口

建议保留两种拓扑，并通过配置组合表达：

- 统一网关：`edge gateway -> app router -> backend services`，适合统一证书、统一审计、统一限流。
- 独立入口：APP 自带网关或 ingress，统一平台只保留注册中心与策略下发能力。

路由分发规则：

1. 先按 `host` 命中 `AppHostBinding`。
2. 再按 `X-App-Key` 与 host 命中结果做一致性校验（防伪造）。
3. 最后把请求下发到 APP 对应上游（service discovery 或静态 upstream）。

#### 10.1.3 按 APP 的限流、熔断与版本路由

- 所有网关策略至少带 `app_key` 维度：`rate_limit`, `circuit_breaker`, `retry_budget`, `timeout_budget`。
- 版本路由采用“同 APP 内按 API 版本分流”，禁止跨 APP 复用版本键导致串流。
- 建议 header：
- `X-App-Key`: 当前 APP
- `X-App-Version`: 前端构建版本（用于灰度与回滚定位）
- `X-Request-ID`: 贯穿 trace 与日志

### 10.2 跨域安全与会话策略

#### 10.2.1 Cookie / SameSite / CORS / CSRF / CSP 基线

- shared_cookie 场景：Cookie Domain 设为主域（如 `.example.com`），`Secure + HttpOnly + SameSite=None`。
- CORS：仅允许注册白名单来源，禁止 `*`；凭证模式必须显式 `Access-Control-Allow-Credentials: true`。
- CSRF：对状态变更接口强制双重校验（CSRF token + Origin/Referer）。
- CSP：按 APP 下发动态策略，至少限制 `script-src`, `frame-ancestors`, `connect-src` 到白名单域。

#### 10.2.2 redirect 白名单与开放跳转防护

后端协议层硬约束：

1. `redirect_uri` 必须命中 APP 维度白名单。
2. 目标地址必须校验 `scheme/host/path_prefix`，拒绝协议相对 URL 与双重编码绕过。
3. `target_app_key`、`target_navigation_space_key`、`target_home_path` 必须做组合校验。

前端兜底层：

- 仅消费后端签名通过的回跳参数。
- 若参数无效，回退到 APP 默认首页，不按前端自行拼接未知 URL。

#### 10.2.3 统一 token / refresh / logout 协议

- token 结构统一：`iss`（认证中心）、`aud`（目标 APP 或平台网关）、`tenant_id`、`session_id`、`scope`。
- refresh token 由认证中心统一发放与轮换，APP 仅透传刷新请求，不私自扩展 refresh 协议。
- logout 分两层：
- 局部登出：当前 APP session 失效。
- 全局登出：认证中心触发全部 APP session 失效，并广播登出事件。

#### 10.2.4 匿名页、受保护页与 redirect 安全

- 匿名页（登录/注册/找回）仅允许进入白名单路径，不接受任意外部 URL 回跳。
- 受保护页未登录时统一跳认证中心，并携带一次性 `state/nonce`。
- redirect 仅允许：
- 同 APP 内部路径
- 白名单内跨 APP 路径
- 其余场景一律拒绝并降级到默认首页。

### 10.3 数据库、部署与运维

#### 10.3.1 共享库 vs 独立库边界

- 共享库（平台级）：`users`、`tenants`、`roles/permissions`、`app registry`、审计日志。
- 业务域库（APP 级）：高写入量业务表、APP 专属配置、APP 专属事件流。
- 原则：权限与身份可共享，业务热点与生命周期冲突数据优先 APP 独立。

#### 10.3.2 跨部署数据同步与 migration 一致性

- 迁移版本必须全局单调，采用“平台核心迁移 + APP 扩展迁移”双通道版本治理。
- APP 独立部署时仍要上报当前 migration version 到平台注册中心，用于一致性巡检。
- 禁止“仅在某环境手工补 SQL 不入迁移链”。

#### 10.3.3 CDN 分发与缓存失效

- 缓存键至少包含 `app_key + asset_version + path`。
- 静态资源强制内容哈希命名；HTML/manifest 采用短缓存 + 强校验。
- 发布流程需支持“按 APP 定向失效”，避免全站清缓存。

#### 10.3.4 前端资产版本标识与产物隔离

- 构建产物目录建议：`/assets/{app_key}/{build_id}/...`。
- 前端请求头带 `X-App-Version`，后端日志与监控按版本聚合。
- 灰度策略按 APP 与版本双维度控制，不再只按全局版本切流。

#### 10.3.5 健康检查、Trace 与 Logging 标签

- 健康检查分三级：`liveness`、`readiness`、`dependency`，并携带 `app_key`。
- Trace 统一标签：`app.key`, `tenant.id`, `space.key`, `auth.mode`, `deployment.host`。
- 日志最小字段集：`timestamp`, `level`, `app_key`, `request_id`, `route`, `status`, `latency_ms`。

#### 10.3.6 独立域名时的构建策略

- 可选两种模式：
- 单仓多入口构建（推荐先行）：同前端工程、按 `app_key` 输出多入口产物。
- APP 独立构建：仅在团队/发布节奏明显分离时启用。
- 约束：无论哪种模式，都保持统一 OpenAPI 契约与统一权限模型，不拆成第二套后台协议。

### 10.4 跨域 APP 通信与前端接入

#### 10.4.1 独立域名 APP 接入选型矩阵

按“是否同域 + 是否需要深度集成”选型：

1. 同域同壳：直接 router 切换（成本最低）。
2. 子域同主域：整页跳转 + shared_cookie（默认方案）。
3. 完全跨域：重定向参数 + 服务端会话交换（安全优先）。
4. iframe 仅用于受控内部工具页，不作为核心业务默认方案。

#### 10.4.2 加载失败的降级回退

- APP 入口探测失败时，优先回退到应用目录页并展示可重试动作。
- 若目标 APP 连续失败，自动切回上一个健康 APP（`lastRuntimeAppKey`）。
- 回退日志必须包含失败阶段（DNS/TLS/HTTP/manifest/bootstrap）。

#### 10.4.3 跨域跨 APP 通信白名单

- 白名单事件示例：`AUTH_LOGIN_SUCCESS`、`AUTH_LOGOUT`、`THEME_CHANGED`、`TENANT_SWITCHED`。
- 非白名单事件禁止浏览器端直连通信，统一回退到服务端会话查询或重定向参数。
- 场景差异：
- 同域：可用 `BroadcastChannel`。
- 子域：优先整页跳转 + 服务端态；必要时用受控 `postMessage`。
- 完全跨域：仅用重定向参数与后端交换，不依赖客户端总线。

## 11. Phase B 设计输出到任务节点映射

| Phase B 设计点 | 对应节点 |
|---|---|
| 独立域名入口边界、认证策略区别 | `5/1/1` |
| 统一网关 vs 独立入口路由分发 | `5/1/2` |
| APP 维度限流/熔断/版本路由 | `5/1/3` |
| Cookie/CORS/CSRF/CSP 动态安全基线 | `5/2/1` |
| redirect 白名单与开放跳转防护 | `5/2/2` |
| token/refresh/logout 统一协议 | `5/2/3` |
| 匿名页/受保护页与回跳安全规则 | `5/2/4` |
| 共享库与独立库拆分边界 | `5/3/1` |
| 跨部署同步与 migration 一致性 | `5/3/2` |
| CDN 分发与缓存失效规则 | `5/3/3` |
| 前端版本标识与产物隔离 | `5/3/4` |
| 健康检查、Trace、日志 APP 标签 | `5/3/5` |
| 独立域名构建模式选择 | `5/3/6` |
| 独立域名前端接入选型矩阵 | `5/4/1` |
| 加载失败降级与回退机制 | `5/4/2` |
| 跨域跨 APP 通信白名单机制 | `5/4/3` |
