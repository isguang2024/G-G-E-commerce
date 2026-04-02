# Field 后台专题

基于 Fluent UI React v9 Storybook 的 `Field` 文档整理。

## Field 的职责

- 给控件补标签
- 给控件补校验信息
- 给控件补提示说明

官方定义里，`Field` 是表单控件外层语义包装，而不是输入控件本身。

## 什么时候必须先想到 Field

- 任何需要明确标签和校验的后台表单
- 设置页开关、选择器、输入框、文本域
- 需要给无标签控件补语义时

## 最佳实践

- 用 `Field` 给表单控件加标签和校验信息
- 也可用于给原本无标签的控件补标签，例如某些进度或展示控件
- 若需要横向布局，可用官方 `Horizontal Orientation` 思路

## 明确不要这样用

- 不要同时展示 `validationMessage` 和 `hint`
- 不要让一个 `Field` 包多个控件
- 不要用 `Field` 的 label 去替代 `Checkbox` 自带 label
  - `Checkbox` 仍可放在 `Field` 内拿到提示或校验信息

## 后台常见模式

### 普通表单行

- `Field` + `Input`
- `Field` + `Textarea`
- `Field` + `Select`
- `Field` + `Combobox`

### 设置页

- `Field` 提供说明和风险提示
- 输入控件承载真正操作
- 开关类项目把补充说明写清楚，不只写一个短标签

### 校验反馈

- 错误靠近字段显示
- 成功或提示信息不要和错误信息同时堆叠
- 长表单按分组展示，避免所有字段都以同一种密度堆满

## 代码评审重点

- 是否每个关键字段都有明确标签
- 是否错误信息靠近字段
- 是否把一个 `Field` 误包了多个输入
- 是否把 `hint` 和 `validationMessage` 同时塞进去
