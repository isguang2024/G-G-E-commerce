import {
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Divider,
  Switch,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

const useStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  panel: {
    display: 'grid',
    gap: '18px',
    padding: '22px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  section: { display: 'grid', gap: '12px' },
  footer: { display: 'flex', gap: '10px', flexWrap: 'wrap' },
});

export function NotificationPreferencesPage() {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <Title3>通知偏好测试页</Title3>
      <section className={styles.panel}>
        <div className={styles.section}>
          <Body1Strong>即时通知</Body1Strong>
          <Caption1>测试 Fluent 2 在偏好页中的开关层级和说明文案节奏。</Caption1>
          <Switch label="桌面提醒" defaultChecked />
          <Switch label="移动端提醒" defaultChecked />
          <Switch label="高风险邮件通知" />
        </div>
        <Divider />
        <div className={styles.section}>
          <Body1Strong>摘要频率</Body1Strong>
          <Body1>把频率说明和偏好项拆成有限层级，不让通知设置页退化成密集表格。</Body1>
          <Switch label="日报摘要" defaultChecked />
          <Switch label="周报摘要" />
        </div>
        <div className={styles.footer}>
          <Button appearance="primary">保存偏好</Button>
          <Button appearance="secondary">恢复默认</Button>
        </div>
      </section>
    </div>
  );
}
