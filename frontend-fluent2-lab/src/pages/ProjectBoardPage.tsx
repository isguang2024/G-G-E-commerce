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
  board: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, minmax(0, 1fr))',
    gap: '16px',
    '@media (max-width: 1200px)': { gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' },
    '@media (max-width: 700px)': { gridTemplateColumns: '1fr' },
  },
  column: {
    display: 'grid',
    gap: '12px',
    alignContent: 'start',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  card: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
});

const columns = [
  { title: '待办', items: ['补登录态视觉', '接入主题切换'] },
  { title: '进行中', items: ['细化频道页', '整理组件规范页'] },
  { title: '复核中', items: ['审批工作台文案', '消息线程右栏'] },
  { title: '已完成', items: ['实验场入口扩充', 'Teams 响应式页'] },
];

export function ProjectBoardPage() {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div>
          <Title3>项目看板测试页</Title3>
          <Caption1>验证 Fluent 2 在轻量卡片看板场景下的列结构和任务优先级表达。</Caption1>
        </div>
        <Button appearance="primary">新建卡片</Button>
      </header>
      <section className={styles.board}>
        {columns.map(column => (
          <article key={column.title} className={styles.column}>
            <Body1Strong>{column.title}</Body1Strong>
            {column.items.map(item => (
              <div key={item} className={styles.card}>
                <Body1Strong>{item}</Body1Strong>
                <Caption1>用于测试任务卡片在 Fluent 2 下的低噪表达。</Caption1>
                <Badge appearance="tint">{column.title}</Badge>
              </div>
            ))}
          </article>
        ))}
      </section>
    </div>
  );
}
