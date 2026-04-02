# 分组表单代码模式

用于设置页、编辑页、配置页和详情编辑场景。

以下代码骨架强调 `Field` 的语义和表单分组，不追求完整表单库实现。

## 适用场景

- 设置中心
- 中大型编辑页
- 抽屉内配置表单
- 带风险说明的后台操作页

## 结构要点

- 顶部说明
- 多个 section 分组
- 每组用 `Field` 包裹输入
- 底部主次操作区
- 危险操作独立分区

## 代码骨架

```tsx
import * as React from 'react';
import {
  Button,
  Field,
  Input,
  MessageBar,
  Select,
  Switch,
  Textarea,
} from '@fluentui/react-components';

export function FormSectionsPage() {
  return (
    <form style={{ display: 'grid', gap: 24 }}>
      <header>
        <h1>空间设置</h1>
        <p>按分组维护基础信息、通知策略和访问控制。</p>
      </header>

      <section>
        <h2>基础信息</h2>

        <Field label="空间名称" validationMessage="请输入对业务清晰的名称">
          <Input />
        </Field>

        <Field label="空间说明">
          <Textarea />
        </Field>
      </section>

      <section>
        <h2>通知策略</h2>

        <Field label="默认通知等级" hint="用于新建任务的默认提醒级别">
          <Select>
            <option value="normal">普通</option>
            <option value="high">高</option>
          </Select>
        </Field>

        <Field label="启用邮件提醒">
          <Switch label="开启" />
        </Field>
      </section>

      <section>
        <h2>风险操作</h2>
        <MessageBar intent="warning">
          修改访问策略会影响当前空间成员的可见范围。
        </MessageBar>
        <Button>重置策略</Button>
      </section>

      <footer style={{ display: 'flex', gap: 12 }}>
        <Button appearance="primary">保存</Button>
        <Button>取消</Button>
      </footer>
    </form>
  );
}
```

## 代码评审重点

- 是否用 section 做了清晰分组
- 关键字段是否都通过 `Field` 提供标签和状态
- 是否错误地同时使用 `hint` 和 `validationMessage`
- 危险操作是否被单独隔离并写明后果
