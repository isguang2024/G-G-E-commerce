---
name: fluent-react-v9
description: 用 Fluent UI React v9 为 React、Vite、TypeScript 前端搭建、重构、评审后台、工作台、数据页、表单页和应用壳。用于新建 React 页面、迁移到 Fluent、替换旧组件体系、统一主题与布局、按 Fluent v9 组织导航、列表、设置页和详情页时。
---

# Fluent React v9

## 概览

把任务视为“用 Fluent UI React v9 组织一套专业产品界面”，不要退化成只替换几个按钮和输入框。

默认面向 React 后台、工作台、设置中心、数据页和管理壳层。

## 工作流

1. 先确认技术边界。
- 优先用于 React + TypeScript 项目。
- 如果当前项目是 Vue 或 Web Components 主线，先判断是否真的要新建 React 子应用。
- 若仓库里同时存在 `@fluentui/react` v8 和 `@fluentui/react-components`，新页面默认只用 v9。

2. 再确认工程骨架。
- 检查是否已经有 `FluentProvider`。
- 检查主题来源：`webLightTheme`、`webDarkTheme` 或自定义主题。
- 检查路由、数据请求、状态管理和错误边界的接入方式。
- 检查是否要求 RTL、键盘导航、密度切换和响应式。

3. 先选官方组件，再补组合层。
- 先查 v9 官方组件和 Storybook 示例。
- 优先复用官方组件，不先写一层仿制组件。
- 只在官方能力无法覆盖目标交互时，再补轻量封装或组合层。
- 若任务涉及安装、Provider、样式、SSR、定位、Slots、主题或 Web Components 互操作，先读取 [references/storybook-doc-map.md](references/storybook-doc-map.md)。

4. 按任务类型选择资料。
- 新建后台或工作台页面时，先读 [references/implementation-checklist.md](references/implementation-checklist.md) 和 [references/admin-workbench-recipes.md](references/admin-workbench-recipes.md)。
- 遇到组件选型或交互拆分时，读 [references/high-frequency-component-recipes.md](references/high-frequency-component-recipes.md)。
- 需要深挖单个高频组件时，按需读取 `datagrid-recipes`、`drawer-recipes`、`nav-recipes`、`tree-recipes`、`field-recipes` 五份专题参考。
- 需要直接套后台页面骨架时，优先读取 `code-patterns-workspace-grid-drawer`、`code-patterns-tree-detail-pane`、`code-patterns-form-sections` 三份代码模式参考。
- 遇到官方文档入口不清楚时，回到 [references/storybook-doc-map.md](references/storybook-doc-map.md)。

## 使用顺序

- 先定页面或壳层结构，再定主题和样式，再定组件组合，最后做局部定制。
- 先决定页面是列表、详情、设置、登录还是应用壳，再决定要用 `DataGrid`、`Table`、`Drawer`、`Dialog`、`Nav`、`Tree` 中哪一组组合。
- 先用官方能力达到 80% 目标，再决定是否需要业务封装。
- 做代码评审时，重点检查 Provider、主题、一致性、焦点、ARIA、组件混用和样式并行体系。
- 若需要快速起手而不是从零搭页面，优先复制 [assets/fluent-react-page-starters](assets/fluent-react-page-starters) 下的小型模板，再按专题参考收口。


## 默认技术主线

- 用 `FluentProvider` 统一主题、方向和全局字体。
- 用 design tokens 与 `makeStyles` 管样式，不再并行维护第二套视觉 token。
- 用 `React Router` 管应用壳和页面切换。
- 用稳定的数据请求层和状态管理，不让 UI 组件直接承担业务同步逻辑。
- 让组件先遵循 v9 默认焦点态、尺寸和间距，再做有限调整。
- 需要确认主线和组件入口时，读取 [references/react-v9-official-notes.md](references/react-v9-official-notes.md)。
- 需要按 Storybook 目录快速定位文档时，读取 [references/storybook-doc-map.md](references/storybook-doc-map.md)。

## 组件选型顺序

- 基础输入优先使用 `Field`、`Input`、`Textarea`、`Combobox`、`Select`、`Checkbox`、`RadioGroup`、`Switch`。
- 操作与反馈优先使用 `Button`、`Menu`、`Toolbar`、`MessageBar`、`Toast`、`Dialog`、`Drawer`、`Popover`。
- 数据展示优先使用 `Table`、`DataGrid`、`Badge`、`Tag`、`Card`、`Skeleton`、`Spinner`。
- 导航与结构优先使用 `Nav`、`Tree`、`TabList`、`Breadcrumb`、`Divider`。
- 非必要不要混入旧 v8 组件或其他重型 UI 库。

## 应用壳模式

- 用“顶部栏 + 左侧导航 + 内容区 + 次级消息层”组织后台壳层。
- 顶部栏只保留品牌、环境信息、全局搜索、主操作和用户菜单。
- 左侧导航负责模块切换和层级定位，不把低频按钮堆进侧栏。
- 内容区优先用标题、摘要、工具栏、主工作区、次级面板组织层次。
- 空态、错误态、加载态统一放在内容区内部，不让页面白屏。
- 常见壳层配方见 [references/admin-workbench-recipes.md](references/admin-workbench-recipes.md)。

## 数据页模式

- 列表页默认结构是“标题区 -> 筛选区 -> 工具栏 -> 表格或数据网格 -> 分页或次级详情”。
- 筛选区优先支持高频字段，低频字段收纳到折叠区或次级面板。
- 工具栏只保留主操作、批量操作、视图切换和刷新，不堆一整排低频文本按钮。
- 表格优先保证可读性、排序、状态标签和行级操作，而不是追求密集堆列。
- 详情查看优先使用 `Drawer`、`Dialog` 或右侧属性区，而不是强跳独立页。
- 通知反馈优先用 `MessageBar`、`Toast` 和内联状态提示。
- 需要具体看高频组件拆法时，读取 [references/high-frequency-component-recipes.md](references/high-frequency-component-recipes.md)。

## 表单模式

- 用 `Field` 承载标签、说明、错误信息和校验状态。
- 复杂表单按业务分组拆成多个 section，不要做一整屏无分层长表单。
- 主按钮只保留一个最强动作，次级按钮靠近主按钮，危险操作单独隔离。
- 禁用态、只读态、加载态必须显式可见，避免用户误判。
- 对需要解释风险的操作，优先加说明文本或确认对话框。

## 实施禁忌

- 不要照搬第三方后台模板视觉语言。
- 不要把 `@fluentui/react` v8、`@fluentui/react-components` v9 和其他设计系统混成一页。
- 不要同时维护 `makeStyles`、CSS Modules、Tailwind 和内联样式四套规则。
- 不要用大量阴影、厚边框和渐变卡片破坏 Fluent 的聚焦感。
- 不要为追求“像桌面端”而牺牲 Web 上的可读性和可维护性。

## 完成前检查

- 检查主题是否由 `FluentProvider` 统一托管。
- 检查焦点态、键盘可达性、ARIA 标签和表单报错是否完整。
- 检查 RTL 下布局是否仍可用。
- 检查密度、响应式和滚动区域是否稳定。
- 检查是否仍有旧 v8 组件或其他 UI 库残留。
- 检查是否优先复用了官方组件而不是手写仿制品。
- 需要做项目级收尾检查时，读取 [references/implementation-checklist.md](references/implementation-checklist.md)。

## 参考资料

- 主线与组件入口：读取 [references/react-v9-official-notes.md](references/react-v9-official-notes.md)
- 后台/工作台页面配方：读取 [references/admin-workbench-recipes.md](references/admin-workbench-recipes.md)
- Storybook 文档树与 Quick Start 入口：读取 [references/storybook-doc-map.md](references/storybook-doc-map.md)
- 高频组件后台配方：读取 [references/high-frequency-component-recipes.md](references/high-frequency-component-recipes.md)
- 项目实施与收尾检查：读取 [references/implementation-checklist.md](references/implementation-checklist.md)
- `DataGrid` 专题：读取 [references/datagrid-recipes.md](references/datagrid-recipes.md)
- `Drawer` 专题：读取 [references/drawer-recipes.md](references/drawer-recipes.md)
- `Nav` 专题：读取 [references/nav-recipes.md](references/nav-recipes.md)
- `Tree` 专题：读取 [references/tree-recipes.md](references/tree-recipes.md)
- `Field` 专题：读取 [references/field-recipes.md](references/field-recipes.md)
- 工作区列表+抽屉代码模式：读取 [references/code-patterns-workspace-grid-drawer.md](references/code-patterns-workspace-grid-drawer.md)
- 树+详情代码模式：读取 [references/code-patterns-tree-detail-pane.md](references/code-patterns-tree-detail-pane.md)
- 分组表单代码模式：读取 [references/code-patterns-form-sections.md](references/code-patterns-form-sections.md)
- 可直接复制的小型模板资产：读取 [assets/fluent-react-page-starters](assets/fluent-react-page-starters)
