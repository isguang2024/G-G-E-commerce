# Fluent UI React v9 Storybook 文档索引

用于把 React Storybook 当作第一文档入口，而不是零散搜索旧文章。

## 建议阅读顺序

1. 先看 `Concepts`
2. 再看 `Theme`
3. 再看 `Components`
4. 最后按需看 `Utilities`、`Motion`、`Migration Shims`

## Concepts

### Introduction

- 先用来理解 v9 的组件组织方式和整体文档结构。

### Developer

- `Quick Start`
  用于安装、最小使用方式、Provider 接入、第一批组件落地。
- `Styling Components`
  用于理解如何给官方组件加样式，而不是重写它们。
- `Positioning Components`
  用于 `Popover`、`Tooltip`、`Menu`、`Dialog`、`Drawer` 等浮层定位。
- `Server-Side Rendering`
  用于 Next.js、SSR 和样式注入顺序。
- `Accessibility`
  用于焦点、键盘导航、ARIA 和无障碍约束。
- `Advanced Configuration`
  用于项目级接入、打包和更复杂的工程配置。
- `Advanced Styling Techniques`
  用于更细的样式控制和主题扩展。
- `Browser Support Matrix`
  用于浏览器支持边界判断。
- `Build time styles`
  用于构建期样式处理相关问题。
- `Building Custom Controls`
  用于组合官方能力构建业务组件。
- `Customizing Components with Slots`
  用于理解 slot 定制与组件组合方式。
- `React Version Support`
  用于确认 React 版本兼容性。
- `Supported Platforms`
  用于确认平台与运行环境边界。
- `Theming`
  用于 `FluentProvider`、主题和 token。
- `Unprocessed Styles`
  用于处理未加工样式或构建链特殊情况。
- `Web Components Interop`
  用于 React 与 Fluent Web Components 混用场景。

### Migration

- 用于从旧版本或旧模式迁移到 v9。

### Recipes

- 用于查常见实现套路和组合方式。

### Package Maturity Levels

- 用于判断包的稳定性与使用风险。

## Theme

- `Border Radii`
- `Colors`
- `Fonts`
- `Shadows`
- `Spacing`
- `Stroke Widths`
- `Typography`
- `Theme Designer`

用法：
- 需要调 token 时先从这里确认，而不是直接写死像素和颜色。
- 做品牌主题时优先继承现有主题，再调整 token，不要自造平行体系。

## Components

常用 React v9 组件目录包括：

- `Accordion`
- `Avatar`
- `AvatarGroup`
- `Badge`
- `Breadcrumb`
- `Button`
- `Card`
- `Carousel`
- `Checkbox`
- `ColorPicker`
- `Combobox`
- `DataGrid`
- `Dialog`
- `Divider`
- `Drawer`
- `Dropdown`
- `Field`
- `FluentProvider`
- `Image`
- `InfoLabel`
- `Input`
- `Label`
- `Link`
- `List`
- `Menu`
- `MessageBar`
- `Nav`
- `Overflow`
- `Persona`
- `Popover`
- `Portal`
- `ProgressBar`
- `RadioGroup`
- `Rating`
- `RatingDisplay`
- `SearchBox`
- `Select`
- `Skeleton`
- `Slider`
- `SpinButton`
- `Spinner`
- `SwatchPicker`
- `Switch`
- `Table`
- `TabList`
- `Tag`
- `TagPicker`
- `TeachingPopover`
- `Text`
- `Textarea`
- `Toast`
- `Toolbar`
- `Tooltip`
- `Tree`

兼容组件目录：

- `Calendar`
- `DatePicker`
- `TimePicker`

使用规则：

- 后台和工作台优先从 `Field`、`Input`、`Combobox`、`Select`、`Button`、`Menu`、`Toolbar`、`MessageBar`、`Dialog`、`Drawer`、`Table`、`DataGrid`、`Nav`、`Tree` 这些高频组件开始。
- 查组件时先看 Storybook 的 Docs，再看 Playground 或示例，不要直接凭旧记忆编码。

## Preview Components

- `Menu`
- `Icons`

只在明确接受预览能力风险时使用。

## Motion

- `APIs`
- `Choreography (preview)`
- `Components (preview)`
- `Motion Slot`
- `Tokens`

只有在动效真能帮助理解层级、状态变化和面板转场时再读。

## Utilities

- `ARIA live`
- `Focus Management`
- `Positioning`
- `Theme`

用于解决复杂焦点、弹层定位、可访问性播报和主题工具问题。

## Migration Shims

- `V0`
- `V8`

仅在兼容旧包或迁移旧项目时再读，不要把它当新项目默认入口。
