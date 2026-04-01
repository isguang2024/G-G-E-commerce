# 架构说明

## 分层

- `app/` 只负责应用入口、Provider 组合、全局样式、错误边界和路由挂载。
- `shared/` 负责类型、mock、请求骨架、配置和轻量 UI 原子。
- `features/navigation/` 负责路由定义、页面 metadata 查询和导航树查询。
- `features/session/` 负责假会话数据查询。
- `features/shell/` 负责顶部栏、侧边导航、菜单空间切换器、页面容器和壳层状态。
- `pages/` 只负责页面内容，不直接持有全局导航或会话状态。

## Provider

- `QueryClientProvider`
- `HashRouter`
- `FluentProvider`
- `AppErrorBoundary`

## 动效策略

- 优先使用 Fluent UI React v9 组件内建的展开、弹出、浮层与对话框动效
- 自定义壳层动效统一复用 Fluent token 节奏，不直接硬编码另一套时长和 easing
- 当前工程允许使用 `tokens.duration*`、`tokens.curve*` 组织过渡，但不把 `@fluentui/react-motion` 作为正式生产依赖入口
- 对壳层悬停、显隐、轻量位移等微交互，默认采用低噪、短时长、弱位移的 Fluent 风格，不做夸张缩放和弹跳

## 状态策略

Zustand 仅承载：

- `themeMode`
- `navCollapsed`
- `mobileNavOpen`
- `currentSpaceKey`
- `currentUser`
- `activeTopContext`

服务端资源模型先通过 Query + mock promise 获取，不进入 Zustand。

## 数据边界

- 当前用户：`['session', 'current-user']`
- 菜单空间：`['shell', 'spaces']`
- 导航树：`['navigation', 'tree']`
- 页面 metadata：`['pages', 'meta', routeId]`

未来接真实接口时，只替换查询函数或 adapter，不重写页面壳。
