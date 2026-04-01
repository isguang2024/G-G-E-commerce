# FRONTEND_GUIDELINE.md

## React Fluent 2 前端规范

适用范围：`frontend-fluentV2/`

## 视觉方向

- Fluent 2 风格
- 微软企业感
- 克制、清晰、低噪
- 先用留白、层级和排版组织信息，不靠旧卡片皮肤堆视觉

## 壳层规范

- 顶部栏只放品牌、当前区域、空间切换、主题入口、用户区和预留全局操作
- 左侧导航负责模块切换和层级定位，不堆业务操作
- 内容区统一使用面包屑、标题、副标题和页面容器
- 未迁移页面必须进入统一占位页，禁止白屏

## 页面组织

- 标题区、操作区、内容区分层明确
- mock、样式、状态、布局不要揉在一个文件
- 页面优先消费 Query 与 typed metadata，不直接读散落常量
- 列表、设置、三栏工作区都要尽量复用统一容器

## 组件与样式

- 优先使用 Fluent UI React v9 官方组件
- 页面样式优先使用 `makeStyles + tokens`
- 不引入 Tailwind
- 不引入其他 UI 库

## 动效规范

- React Fluent 2 线的动效优先使用 Fluent 组件自带的 motion 行为，不额外手搓第二套交互节奏
- 需要自定义过渡时，优先复用 Fluent token 中的时长与缓动，例如 `tokens.durationNormal`、`tokens.curveEasyEase`
- 仅在 Fluent 组件现成能力无法覆盖时，才补最小范围的自定义 CSS transition / animation
- 当前不要直接把 `@fluentui/react-motion` 当作生产基建；该包在当前版本中仍标注为非 production-ready，后续待官方稳定后再评估升级

## 数据与状态

- Zustand 只管理壳层状态
- TanStack Query 管页面资源和 mock 查询
- Axios 只做请求 client 和错误归一化骨架
