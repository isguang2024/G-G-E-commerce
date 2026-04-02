# Fluent UI React v9 高频组件后台配方

用于后台、工作台和管理页的高频组合，不追求展示所有 props，只强调选型和结构。

## FluentProvider

- 每个 React 应用只保留一个明确的主题入口。
- 在应用根部包裹 `FluentProvider`，统一方向、主题和全局字体。
- 不要在业务页局部再套另一层语义不清的 Provider。

## Nav + Tree

- `Nav` 更适合一级或二级模块导航。
- `Tree` 更适合层级数据、资源结构和节点管理。
- 后台左侧模块切换优先 `Nav`，对象树或菜单树优先 `Tree`。
- 不要让两者同时承担同一层导航职责。

## Toolbar + Menu + Button

- 工具栏放高频动作、视图切换、筛选入口和刷新。
- 低频动作收纳进 `Menu`。
- 主按钮只保留一个最强动作，避免多个同权主按钮并列。

## DataGrid vs Table

- 需要排序、列定义统一、选择和密集列表时优先 `DataGrid`。
- 结构更自由、单元格排布更特别时用 `Table`。
- 后台列表页优先先判断是否真的需要网格能力，不要默认把所有列表都做成复杂表格。

## Drawer vs Dialog

- 详情查看、属性编辑、不中断主列表时优先 `Drawer`。
- 轻量确认、不可逆动作、短表单时优先 `Dialog`。
- 不要把大表单塞进狭小 `Dialog`，也不要把简单确认过度升级成 `Drawer`。

## Field + Input + Combobox + Select

- 需要标签、说明、错误信息和状态时，用 `Field` 包输入组件。
- 有搜索、异步候选或自由输入倾向时优先 `Combobox`。
- 纯固定项选择时优先 `Select`。
- 文本筛选、主搜索和短输入优先 `Input`。

## MessageBar + Toast

- 页面级或区块级问题优先 `MessageBar`。
- 瞬时成功反馈或非阻断提示可用 `Toast`。
- 不要把关键错误只放进瞬时 `Toast`。

## Tag + Badge

- `Tag` 适合过滤、轻量状态和可移除标签。
- `Badge` 适合计数、提醒和紧凑状态点。
- 不要用纯颜色块代替可读状态文本。

## SearchBox 与筛选区

- 若页面筛选逻辑复杂，不要只用一个搜索框承载全部条件。
- 高后台价值页面优先“搜索 + 若干高频筛选 + 更多筛选”结构。

## 页面组合建议

- 列表治理页：`Toolbar + DataGrid + Drawer`
- 菜单或层级页：`Tree + Toolbar + Detail Pane/Drawer`
- 设置页：`Section + Field + Switch/Select + MessageBar`
- 审批或风险页：`MessageBar + Summary + DataGrid + Dialog`

## 进一步阅读

- `DataGrid` 深入：读取 [datagrid-recipes.md](datagrid-recipes.md)
- `Drawer` 深入：读取 [drawer-recipes.md](drawer-recipes.md)
- `Nav` 深入：读取 [nav-recipes.md](nav-recipes.md)
- `Tree` 深入：读取 [tree-recipes.md](tree-recipes.md)
- `Field` 深入：读取 [field-recipes.md](field-recipes.md)
