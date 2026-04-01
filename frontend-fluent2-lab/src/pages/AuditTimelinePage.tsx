import {
  Badge,
  Body1,
  Body1Strong,
  Caption1,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

const useStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  timeline: {
    display: 'grid',
    gap: '14px',
    padding: '20px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  item: {
    display: 'grid',
    gap: '6px',
    paddingLeft: '16px',
    borderLeft: `2px solid ${tokens.colorBrandStroke1}`,
  },
});

export function AuditTimelinePage() {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <Title3>审计时间线测试页</Title3>
      <section className={styles.timeline}>
        {[
          ['10:08', '创建实验场入口', '成功'],
          ['10:16', '新增 Teams 频道工作台', '完成'],
          ['10:24', '补充 Fluent 规范工作区', '复核中'],
          ['10:31', '构建验证通过', '成功'],
        ].map(([time, text, status]) => (
          <article key={time} className={styles.item}>
            <Caption1>{time}</Caption1>
            <Body1Strong>{text}</Body1Strong>
            <div>
              <Badge appearance="tint" color={status === '复核中' ? 'warning' : 'success'}>
                {status}
              </Badge>
            </div>
            <Body1>用于验证时间线、状态标签和审计型页面的层级表达。</Body1>
          </article>
        ))}
      </section>
    </div>
  );
}
