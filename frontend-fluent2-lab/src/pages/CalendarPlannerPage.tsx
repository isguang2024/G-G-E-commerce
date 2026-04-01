import {
  Badge,
  Body1Strong,
  Button,
  Caption1,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

const useStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  toolbar: { display: 'flex', justifyContent: 'space-between', gap: '12px', flexWrap: 'wrap' },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(7, minmax(0, 1fr))',
    gap: '10px',
    '@media (max-width: 900px)': { gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' },
    '@media (max-width: 560px)': { gridTemplateColumns: '1fr' },
  },
  day: {
    display: 'grid',
    gap: '10px',
    alignContent: 'start',
    minHeight: '140px',
    padding: '12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
});

export function CalendarPlannerPage() {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <header className={styles.toolbar}>
        <div>
          <Title3>日程规划测试页</Title3>
          <Caption1>验证日历块状布局、轻量标签和多列响应式折叠方式。</Caption1>
        </div>
        <Button appearance="primary">创建日程</Button>
      </header>
      <section className={styles.grid}>
        {['一', '二', '三', '四', '五', '六', '日'].map((day, index) => (
          <article key={day} className={styles.day}>
            <Body1Strong>周{day}</Body1Strong>
            <Caption1>{10 + index} 月 0{index + 1} 日</Caption1>
            <Badge appearance="tint" color={index % 2 === 0 ? 'brand' : 'success'}>
              {index % 2 === 0 ? '设计评审' : '团队同步'}
            </Badge>
          </article>
        ))}
      </section>
    </div>
  );
}
