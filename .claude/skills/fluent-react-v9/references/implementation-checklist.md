# Fluent UI React v9 实施检查清单

## 工程接入

- 已安装并使用 `@fluentui/react-components`
- 已建立单一 `FluentProvider`
- 已确定主题入口与暗色/亮色策略
- 已确定路由、数据请求和状态管理边界

## 文档使用顺序

- 最小接入先看 `Developer -> Quick Start`
- 组件样式先看 `Styling Components`
- 浮层定位先看 `Positioning Components`
- 主题先看 `Theming`
- SSR 场景先看 `Server-Side Rendering`
- React 与 Web Components 混用先看 `Web Components Interop`

## 页面结构

- 已先定页面类型：登录、列表、详情、设置、应用壳
- 已确定主操作、次级操作和危险操作分层
- 已确定空态、错误态、加载态位置
- 已避免页面级元素都争抢注意力
- 已确认列表页到底该用 `DataGrid` 还是 `Table`
- 已确认导航到底该用 `Nav` 还是 `Tree`
- 已确认详情区到底该用 `Drawer` 还是 `Dialog`

## 视觉与样式

- 未混入 v8 组件
- 未并行维护多套样式体系
- 未写死大量颜色、圆角、阴影和间距
- 已优先复用 tokens 和官方样式能力

## 可访问性与国际化

- 焦点态可见
- 键盘路径连续
- 表单错误信息可读
- RTL 下布局未破坏
- 长文本、缩放和窄屏下仍可读

## 收尾前

- 组件选型有明确理由
- 高频路径已冒烟
- 预览或兼容组件的使用风险已说明
- 使用到的官方文档入口已在代码评审说明里可追溯
- 若用到了 `DataGrid`、`Drawer`、`Nav`、`Tree`、`Field`，已按对应专题参考自检
- 若页面属于高频后台模式，已参考至少一个代码模式文件
