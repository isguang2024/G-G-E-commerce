import {
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Input,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

const useStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  toolbar: { display: 'flex', justifyContent: 'space-between', gap: '12px', flexWrap: 'wrap' },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '16px',
    '@media (max-width: 1000px)': { gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' },
    '@media (max-width: 640px)': { gridTemplateColumns: '1fr' },
  },
  item: {
    display: 'grid',
    gap: '10px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  preview: {
    minHeight: '120px',
    borderRadius: tokens.borderRadiusLarge,
    background: 'linear-gradient(135deg, rgba(15,108,189,0.1), rgba(232,242,252,0.8))',
  },
});

export function ResourceLibraryPage() {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <header className={styles.toolbar}>
        <div>
          <Title3>资源库测试页</Title3>
          <Caption1>验证资源卡、预览块、标签和搜索条的组合方式。</Caption1>
        </div>
        <Input placeholder="搜索模板、组件或资源" />
      </header>
      <section className={styles.grid}>
        {['Design assets', 'Pattern notes', 'Launch kit', 'Icon pack', 'Spec export', 'Review checklist'].map(name => (
          <article key={name} className={styles.item}>
            <div className={styles.preview} />
            <Body1Strong>{name}</Body1Strong>
            <Body1>用于测试资源型页面的轻量卡片、预览占位和元数据节奏。</Body1>
            <Badge appearance="tint">Library</Badge>
            <Button appearance="subtle">查看资源</Button>
          </article>
        ))}
      </section>
    </div>
  );
}
