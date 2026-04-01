import {
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

const useStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  header: { display: 'flex', justifyContent: 'space-between', gap: '12px', flexWrap: 'wrap' },
  actions: { display: 'flex', gap: '10px', flexWrap: 'wrap' },
  stats: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 900px)': { gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' },
    '@media (max-width: 640px)': { gridTemplateColumns: '1fr' },
  },
  card: {
    display: 'grid',
    gap: '8px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  layout: {
    display: 'grid',
    gridTemplateColumns: '1.3fr 0.7fr',
    gap: '18px',
    '@media (max-width: 1000px)': { gridTemplateColumns: '1fr' },
  },
  chart: {
    minHeight: '280px',
    display: 'grid',
    gap: '16px',
    alignContent: 'end',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground3,
    backgroundImage: 'linear-gradient(180deg, rgba(15,108,189,0.18), rgba(15,108,189,0.04))',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  bars: {
    display: 'grid',
    gridTemplateColumns: 'repeat(8, minmax(0, 1fr))',
    gap: '10px',
    alignItems: 'end',
    minHeight: '180px',
  },
  bar: { borderRadius: `${tokens.borderRadiusMedium} ${tokens.borderRadiusMedium} 0 0`, backgroundColor: tokens.colorBrandBackground },
  sideList: { display: 'grid', gap: '12px' },
});

export function AnalyticsOverviewPage() {
  const styles = useStyles();
  const heights = ['38%', '52%', '48%', '65%', '72%', '58%', '80%', '68%'];
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div>
          <Title3>分析总览测试页</Title3>
          <Caption1>验证 Fluent 2 数据总览、趋势图和右侧摘要区的组合节奏。</Caption1>
        </div>
        <div className={styles.actions}>
          <Button appearance="secondary">导出快照</Button>
          <Button appearance="primary">生成周报</Button>
        </div>
      </header>
      <section className={styles.stats}>
        {['访问量', '转化率', '活跃空间', '风险告警'].map((label, index) => (
          <article key={label} className={styles.card}>
            <Caption1>{label}</Caption1>
            <Body1Strong>{['182k', '12.4%', '48', '3'][index]}</Body1Strong>
            <Badge appearance="tint" color={index === 3 ? 'danger' : 'success'}>
              {index === 3 ? '需关注' : '稳定'}
            </Badge>
          </article>
        ))}
      </section>
      <section className={styles.layout}>
        <article className={styles.chart}>
          <Body1Strong>八日趋势</Body1Strong>
          <div className={styles.bars}>
            {heights.map((height, index) => (
              <div key={index} className={styles.bar} style={{ height }} />
            ))}
          </div>
        </article>
        <aside className={styles.card}>
          <Body1Strong>洞察摘要</Body1Strong>
          <div className={styles.sideList}>
            <Body1>本周高峰出现在周四下午，审批与消息相关页面使用频次明显上升。</Body1>
            <Body1>导航实验页的停留时长稳定，说明工作台骨架已经接近可用状态。</Body1>
            <Body1>下一步应把分析页中的占位图形替换为更真实的数据卡和明细区。</Body1>
          </div>
        </aside>
      </section>
    </div>
  );
}
