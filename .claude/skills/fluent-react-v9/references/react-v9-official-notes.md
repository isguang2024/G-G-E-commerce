# Fluent UI React v9 官方要点

## 主线判断

- `@fluentui/react-components` 是当前 Fluent UI 的 React 主线。
- 官方仓库将 React Components v9 描述为面向未来的新主线，并鼓励从旧版本迁移到 v9。
- 不要把新页面建立在 `@fluentui/react` v8 的心智上。
- 预览组件和迁移垫片不是新项目默认入口，只有在明确接受风险或做兼容迁移时才读。

## 工程接入

- 入口优先使用 `FluentProvider`。
- 主题优先从 `webLightTheme`、`webDarkTheme` 或自定义主题派生。
- 样式优先使用 design tokens 与 `makeStyles`。
- 方向、密度和无障碍优先走官方组件默认机制，不要自建并行体系。
- Storybook 的 `Developer -> Quick Start` 应作为最小接入入口。

## 常见组件

- 结构与导航：`Nav`、`Tree`、`Breadcrumb`、`TabList`、`Divider`
- 表单与输入：`Field`、`Input`、`Textarea`、`Combobox`、`Select`、`Checkbox`、`RadioGroup`、`Switch`
- 数据展示：`Table`、`DataGrid`、`Badge`、`Tag`、`Card`
- 反馈与浮层：`Dialog`、`Drawer`、`Popover`、`MessageBar`、`Toast`、`Spinner`、`Skeleton`
- 操作区：`Button`、`Menu`、`Toolbar`

## 文档入口

- 仓库总入口：`microsoft/fluentui`
- React 组件示例与组件目录：React Storybook
- 当组件 API 和旧博客、旧示例不一致时，优先以 Storybook 和当前包文档为准
- 若要从 Storybook 中快速定位开发者文档、主题文档、组件文档和工具文档，读取 [storybook-doc-map.md](storybook-doc-map.md)

## 开发者 Quick Start 关注点

- 安装与最小接入：先看 `Concepts -> Developer -> Quick Start`
- 组件样式覆盖：看 `Styling Components` 与 `Advanced Styling Techniques`
- 定位与浮层：看 `Positioning Components`
- 服务端渲染：看 `Server-Side Rendering`
- 无障碍：看 `Accessibility`
- 主题与 Provider：看 `Theming`
- 构建控制与样式产物：看 `Advanced Configuration`、`Build time styles`、`Unprocessed Styles`
- 自定义能力：看 `Building Custom Controls`、`Customizing Components with Slots`
- 兼容与平台边界：看 `React Version Support`、`Supported Platforms`、`Browser Support Matrix`
- React 与 Web Components 互操作：看 `Web Components Interop`

## 实施提醒

- 先确定页面结构，再套组件。
- 优先组合官方组件，不先写自定义大而全基础库。
- 主题、焦点态、禁用态、空态和加载态要一起看，不要只看静态截图。
- 处理复杂问题时，优先按 Storybook 的“概念 -> 主题 -> 组件 -> 工具”顺序定位资料，而不是直接在社区博客里搜旧答案。
- 后台高频页的实践组合可继续读取 [high-frequency-component-recipes.md](high-frequency-component-recipes.md)。
