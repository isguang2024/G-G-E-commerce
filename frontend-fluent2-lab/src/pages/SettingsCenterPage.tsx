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
  layout: {
    display: 'grid',
    gridTemplateColumns: '260px minmax(0, 1fr)',
    gap: '18px',
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  rail: {
    display: 'grid',
    gap: '8px',
    alignContent: 'start',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  railItem: {
    display: 'grid',
    gap: '4px',
    padding: '10px 12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  content: {
    display: 'grid',
    gap: '16px',
    padding: '20px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  section: { display: 'grid', gap: '12px' },
  footer: { display: 'flex', gap: '10px', flexWrap: 'wrap' },
});

export function SettingsCenterPage() {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <Title3>设置中心测试页</Title3>
      <div className={styles.layout}>
        <aside className={styles.rail}>
          {['显示', '通知', '隐私', '访问控制'].map(item => (
            <div key={item} className={styles.railItem}>
              <Body1Strong>{item}</Body1Strong>
              <Caption1>设置分组入口</Caption1>
            </div>
          ))}
        </aside>
        <section className={styles.content}>
          <div className={styles.section}>
            <Body1Strong>显示设置</Body1Strong>
            <Caption1>用来验证设置页分组、说明与开关之间的关系。</Caption1>
            <Switch label="启用紧凑布局" defaultChecked />
            <Switch label="显示模块说明" defaultChecked />
            <Switch label="自动折叠低频栏位" />
          </div>
          <Divider />
          <div className={styles.section}>
            <Body1Strong>通知设置</Body1Strong>
            <Body1>把说明文本、开关和保存动作放在统一节奏里，避免设置页看起来像堆叠表单。</Body1>
            <Switch label="接收高优先级提醒" defaultChecked />
            <Switch label="接收日报摘要" />
          </div>
          <div className={styles.footer}>
            <Button appearance="primary">保存设置</Button>
            <Button appearance="secondary">恢复默认</Button>
          </div>
        </section>
      </div>
    </div>
  );
}
