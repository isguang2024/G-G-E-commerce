import * as React from 'react';
import {
  Body1Strong,
  Button,
  Field,
  Input,
  MessageBar,
  Select,
  Switch,
  Textarea,
  Title3,
  makeStyles,
} from '@fluentui/react-components';

const useStyles = makeStyles({
  form: {
    display: 'grid',
    gap: '24px',
    maxWidth: '720px',
  },
  section: {
    display: 'grid',
    gap: '12px',
  },
  footer: {
    display: 'flex',
    gap: '12px',
  },
});

export function FormSectionsPage() {
  const styles = useStyles();

  return (
    <form className={styles.form}>
      <header>
        <Title3>分组表单页</Title3>
        <p>使用 Field 统一标签、说明和校验语义，再按 section 管理复杂表单。</p>
      </header>

      <section className={styles.section}>
        <Body1Strong>基础信息</Body1Strong>

        <Field label="空间名称" validationMessage="请输入对业务清晰的名称">
          <Input />
        </Field>

        <Field label="空间说明">
          <Textarea />
        </Field>
      </section>

      <section className={styles.section}>
        <Body1Strong>通知策略</Body1Strong>

        <Field label="默认通知等级" hint="用于新建任务的默认提醒级别">
          <Select defaultValue="normal">
            <option value="normal">普通</option>
            <option value="high">高</option>
          </Select>
        </Field>

        <Field label="启用邮件提醒">
          <Switch label="开启" defaultChecked />
        </Field>
      </section>

      <section className={styles.section}>
        <Body1Strong>风险操作</Body1Strong>
        <MessageBar intent="warning">
          修改访问策略会影响当前空间成员的可见范围。
        </MessageBar>
        <Button>重置策略</Button>
      </section>

      <footer className={styles.footer}>
        <Button appearance="primary">保存</Button>
        <Button>取消</Button>
      </footer>
    </form>
  );
}
