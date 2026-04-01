import {
  Avatar,
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Input,
  Subtitle2,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import { LabBadgeRow, LabRailCard, LabSectionTitle, LabStatGrid, LabSurfaceCard } from '../lab/primitives';

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '18px',
  },
  shell: {
    minHeight: '760px',
    display: 'grid',
    gridTemplateColumns: '300px minmax(0, 1fr) 280px',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow16,
    '@media (max-width: 1200px)': {
      gridTemplateColumns: '300px minmax(0, 1fr)',
    },
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  threadRail: {
    display: 'grid',
    gridTemplateRows: 'auto auto 1fr',
    gap: '12px',
    padding: '18px 16px',
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.05) 0%, rgba(15,108,189,0.01) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  threadList: {
    display: 'grid',
    gap: '10px',
    alignContent: 'start',
  },
  main: {
    display: 'grid',
    gridTemplateRows: 'auto auto 1fr auto',
    minWidth: 0,
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '12px',
    padding: '20px 24px 14px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  headerActions: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
  feed: {
    display: 'grid',
    gap: '14px',
    padding: '18px 24px',
    alignContent: 'start',
    background:
      'radial-gradient(circle at top right, rgba(15,108,189,0.06), transparent 34%), linear-gradient(180deg, var(--fluent-neutral-background-2, #f8f8f8) 0%, var(--fluent-neutral-background-1, #ffffff) 100%)',
  },
  bubble: {
    display: 'grid',
    gap: '8px',
    maxWidth: '760px',
    padding: '14px 16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  bubbleBrand: {
    backgroundColor: tokens.colorBrandBackground2,
    border: `1px solid ${tokens.colorBrandStroke1}`,
  },
  bubbleMeta: {
    display: 'flex',
    gap: '10px',
    alignItems: 'start',
  },
  composer: {
    display: 'grid',
    gap: '12px',
    padding: '16px 24px 20px',
    borderTop: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  composerActions: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '10px',
    flexWrap: 'wrap',
  },
  aside: {
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '18px',
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.03) 0%, rgba(15,108,189,0.00) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 1200px)': {
      display: 'none',
    },
  },
  sideCard: {
    display: 'grid',
    gap: '10px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  threadBanner: {
    display: 'grid',
    gap: '12px',
    padding: '18px 20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.10) 0%, rgba(15,108,189,0.03) 58%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  threadGrid: {
    display: 'grid',
    gridTemplateColumns: '1fr 0.85fr',
    gap: '14px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  threadSummary: {
    display: 'grid',
    gap: '6px',
  },
});

function ThreadCard({
  title,
  preview,
  active = false,
}: {
  title: string;
  preview: string;
  active?: boolean;
}) {
  return (
    <LabRailCard active={active}>
      <Body1Strong>{title}</Body1Strong>
      <Caption1>{preview}</Caption1>
    </LabRailCard>
  );
}

export function TeamsConversationPage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <Title3>Teams 消息线程测试页</Title3>
      <div className={styles.shell}>
        <aside className={styles.threadRail}>
          <LabSectionTitle overline="Threads" title="发布协调线程" description="左栏保留线程检索与上下文切换，不承担正文阅读。" />
          <Input placeholder="搜索线程或成员" />
          <div className={styles.threadList}>
            <ThreadCard title="发布同步" preview="检查今天构建与页面进度。" active />
            <ThreadCard title="设计反馈" preview="需要确认频道页的右栏信息层级。" />
            <ThreadCard title="审批补充" preview="北区策略变更等待补充说明。" />
          </div>
        </aside>

        <main className={styles.main}>
          <header className={styles.header}>
            <div>
              <Body1Strong>发布同步</Body1Strong>
              <Caption1>把 Teams 的消息线程结构转成更适合协作后台的讨论区和操作区。</Caption1>
            </div>
            <div className={styles.headerActions}>
              <Badge appearance="tint" color="success">
                进行中
              </Badge>
              <Button appearance="subtle">打开附件</Button>
            </div>
          </header>

          <section className={styles.feed}>
            <div className={styles.threadBanner}>
              <LabBadgeRow>
                <Badge appearance="filled" color="brand">
                  Teams conversation
                </Badge>
                <Badge appearance="tint" color="success">
                  2 条未读
                </Badge>
                <Badge appearance="tint">3 个关联任务</Badge>
              </LabBadgeRow>
              <div className={styles.threadGrid}>
                <div className={styles.threadSummary}>
                  <Subtitle2>把讨论留在同一条线程里，减少来回解释。</Subtitle2>
                  <Caption1>这张页更强调信息流、回复和任务联动，而不是聊天样式本身。</Caption1>
                </div>
                <LabStatGrid
                  items={[
                    { label: '参与成员', value: '6', tone: 'success' },
                    { label: '未读回复', value: '2', tone: 'warning' },
                    { label: '联动动作', value: '5', tone: 'brand' },
                  ]}
                />
              </div>
            </div>
          </section>

          <section className={styles.feed}>
            <LabSurfaceCard>
              <article className={styles.bubble}>
              <div className={styles.bubbleMeta}>
                <Avatar name="Design Ops" color="colorful" />
                <div>
                  <Body1Strong>Design Ops</Body1Strong>
                  <Caption1>今天 10:12</Caption1>
                </div>
              </div>
              <Body1>
                频道工作台的右栏信息量已经接近上限，建议把“规则说明”和“成员在线状态”拆开，不要继续往同一张卡里堆内容。
              </Body1>
              </article>
            </LabSurfaceCard>

            <LabSurfaceCard active>
              <article className={`${styles.bubble} ${styles.bubbleBrand}`}>
              <div className={styles.bubbleMeta}>
                <Avatar name="Frontend Lab" color="brand" />
                <div>
                  <Body1Strong>Frontend Lab</Body1Strong>
                  <Caption1>今天 10:18</Caption1>
                </div>
              </div>
              <Body1>
                已完成实验场扩充，下一步准备把一个更具体的 Figma Frame 映射到消息线程页，优先验证左侧线程列表和主内容区的切换稳定性。
              </Body1>
              </article>
            </LabSurfaceCard>
          </section>

          <footer className={styles.composer}>
            <Input placeholder="输入回复内容，用于验证输入区、按钮区和页脚节奏。" />
            <div className={styles.composerActions}>
              <div className={styles.headerActions}>
                <Button appearance="subtle">添加附件</Button>
                <Button appearance="subtle">插入任务</Button>
              </div>
              <Button appearance="primary">发送更新</Button>
            </div>
          </footer>
        </main>

        <aside className={styles.aside}>
          <LabSurfaceCard subtle>
            <section className={styles.sideCard}>
              <Subtitle2>线程信息</Subtitle2>
              <Caption1>参与成员 6 人，未读回复 2 条，关联任务 3 个。</Caption1>
            </section>
          </LabSurfaceCard>
          <LabSurfaceCard>
            <section className={styles.sideCard}>
              <Subtitle2>联动建议</Subtitle2>
              <Caption1>把消息线程页与频道工作台页一起看，重点观察左栏、主区和右栏的权重是否稳定。</Caption1>
            </section>
          </LabSurfaceCard>
        </aside>
      </div>
    </div>
  );
}
