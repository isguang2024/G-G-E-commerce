import {
  Avatar,
  Badge,
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
  card: {
    display: 'grid',
    gap: '12px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  meta: { display: 'flex', gap: '12px', alignItems: 'center' },
  actions: { display: 'flex', gap: '10px', flexWrap: 'wrap' },
});

const users = ['刘广', 'Design Ops', 'Platform QA', 'North Ops', 'Security Lead', 'Workspace PM'];

export function PeopleDirectoryPage() {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <header className={styles.toolbar}>
        <div>
          <Title3>成员目录测试页</Title3>
          <Caption1>验证卡片式成员目录、标签和轻量操作在 Fluent 2 下的密度。</Caption1>
        </div>
        <div className={styles.actions}>
          <Input placeholder="搜索成员" />
          <Button appearance="primary">邀请成员</Button>
        </div>
      </header>
      <section className={styles.grid}>
        {users.map((user, index) => (
          <article key={user} className={styles.card}>
            <div className={styles.meta}>
              <Avatar name={user} color={index % 2 === 0 ? 'brand' : 'colorful'} />
              <div>
                <Body1Strong>{user}</Body1Strong>
                <Caption1>产品协作角色</Caption1>
              </div>
            </div>
            <Badge appearance="tint" color={index % 3 === 0 ? 'success' : 'informative'}>
              {index % 3 === 0 ? '在线' : '同步中'}
            </Badge>
            <div className={styles.actions}>
              <Button appearance="subtle">查看</Button>
              <Button appearance="subtle">发消息</Button>
            </div>
          </article>
        ))}
      </section>
    </div>
  );
}
