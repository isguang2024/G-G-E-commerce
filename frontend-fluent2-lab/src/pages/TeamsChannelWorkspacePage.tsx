import {
  Avatar,
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Divider,
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
  workspace: {
    minHeight: '760px',
    display: 'grid',
    gridTemplateColumns: '272px minmax(0, 1fr) 280px',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    boxShadow: tokens.shadow16,
    backgroundColor: tokens.colorNeutralBackground1,
    '@media (max-width: 1200px)': {
      gridTemplateColumns: '240px minmax(0, 1fr)',
    },
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  leftRail: {
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '20px 16px',
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.06) 0%, rgba(15,108,189,0.01) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  railSection: {
    display: 'grid',
    gap: '8px',
  },
  railLabel: {
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
    textTransform: 'uppercase',
    letterSpacing: '0.04em',
  },
  channelList: {
    display: 'grid',
    gap: '6px',
  },
  main: {
    display: 'grid',
    gridTemplateRows: 'auto auto auto 1fr',
    minWidth: 0,
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '12px',
    padding: '20px 24px 14px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 700px)': {
      flexDirection: 'column',
      alignItems: 'stretch',
    },
  },
  headerActions: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
  statusRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
  },
  hero: {
    display: 'grid',
    gap: '12px',
    padding: '18px 24px 12px',
  },
  heroText: {
    maxWidth: '820px',
  },
  heroBanner: {
    display: 'grid',
    gap: '12px',
    padding: '18px 20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.10) 0%, rgba(15,108,189,0.03) 52%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroRow: {
    display: 'grid',
    gridTemplateColumns: '1fr 0.8fr',
    gap: '14px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  summaryChipRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  summaryBlock: {
    display: 'grid',
    gap: '6px',
    alignContent: 'start',
  },
  board: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '18px',
    padding: '0 24px 24px',
    '@media (max-width: 1100px)': {
      gridTemplateColumns: '1fr',
    },
  },
  panel: {
    display: 'grid',
    gap: '14px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
  },
  panelHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
  },
  feed: {
    display: 'grid',
    gap: '12px',
  },
  feedItem: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  feedMeta: {
    display: 'flex',
    gap: '10px',
    alignItems: 'start',
  },
  feedActions: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  taskList: {
    display: 'grid',
    gap: '10px',
  },
  taskItem: {
    display: 'grid',
    gap: '8px',
    paddingBottom: '10px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  taskMeta: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
  },
  rightRail: {
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '20px 18px',
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.04) 0%, rgba(15,108,189,0.00) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 1200px)': {
      display: 'none',
    },
  },
  memberList: {
    display: 'grid',
    gap: '12px',
  },
  memberItem: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
  },
  memberMeta: {
    display: 'flex',
    gap: '10px',
    alignItems: 'center',
  },
});

function ChannelItem({
  title,
  subtitle,
  active = false,
}: {
  title: string;
  subtitle: string;
  active?: boolean;
}) {
  return (
    <LabRailCard active={active}>
      <Body1Strong>{title}</Body1Strong>
      <Caption1>{subtitle}</Caption1>
    </LabRailCard>
  );
}

export function TeamsChannelWorkspacePage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <Title3>Teams 协作工作台测试页</Title3>
      <div className={styles.workspace}>
        <aside className={styles.leftRail}>
          <section className={styles.railSection}>
            <LabSectionTitle overline="Space" title="North Region Launch" description="用于验证频道、会话流和成员侧栏的布局关系。" />
            <LabRailCard>
              <Body1Strong>North Region Launch</Body1Strong>
              <Caption1>Fluent 2 基座共享组件，Teams 线只改变频道节奏和协作布局语法。</Caption1>
            </LabRailCard>
          </section>

          <section className={styles.railSection}>
            <span className={styles.railLabel}>Channels</span>
            <div className={styles.channelList}>
              <ChannelItem title="总览" subtitle="项目摘要与风险" />
              <ChannelItem title="发布协调" subtitle="今日更新 12 条" active />
              <ChannelItem title="设计同步" subtitle="组件与规范" />
              <ChannelItem title="审批台" subtitle="待确认 4 项" />
            </div>
          </section>
        </aside>

        <main className={styles.main}>
          <header className={styles.header}>
            <div>
              <Body1Strong>发布协调</Body1Strong>
              <Caption1>以 Teams 风格的信息架构验证频道工作台、更新流和右侧成员栏。</Caption1>
            </div>
            <div className={styles.headerActions}>
              <Button appearance="secondary">查看日程</Button>
              <Button appearance="primary">发起更新</Button>
            </div>
          </header>

          <section className={styles.hero}>
            <div className={styles.heroBanner}>
              <LabBadgeRow>
                <Badge appearance="filled" color="brand">
                  Channel workspace
                </Badge>
                <Badge appearance="tint" color="success">
                  3 项已完成
                </Badge>
                <Badge appearance="tint" color="warning">
                  2 项待确认
                </Badge>
              </LabBadgeRow>
              <div className={styles.heroRow}>
                <div className={styles.summaryBlock}>
                  <Subtitle2 className={styles.heroText}>
                    这一页不模拟聊天气泡，而是把 Teams 的频道组织方式转成更适合后台工作台的活动流与任务联动布局。
                  </Subtitle2>
                  <Caption1>重点不是聊天，而是频道、任务和成员的联动。</Caption1>
                </div>
                <div className={styles.summaryBlock}>
                  <Caption1>协作概况</Caption1>
                  <Body1Strong>频道节奏稳定，任务流可追踪</Body1Strong>
                  <Caption1>右侧成员与规则保持低噪音，不抢主工作区内容。</Caption1>
                </div>
              </div>
            </div>
            <LabStatGrid
              items={[
                { label: '今日更新', value: '12', tone: 'brand' },
                { label: '开放任务', value: '7', tone: 'warning' },
                { label: '在线成员', value: '18', tone: 'success' },
              ]}
            />
          </section>

          <section className={styles.board}>
            <LabSurfaceCard>
              <article className={styles.panel}>
              <div className={styles.panelHeader}>
                <Body1Strong>更新流</Body1Strong>
                <Button appearance="subtle">筛选消息</Button>
              </div>
              <div className={styles.feed}>
                <div className={styles.feedItem}>
                  <div className={styles.feedMeta}>
                    <Avatar name="Liu Guang" color="brand" />
                    <div>
                      <Body1Strong>已同步今天的发布检查项</Body1Strong>
                      <Caption1>刘广 · 5 分钟前</Caption1>
                    </div>
                  </div>
                  <Body1>
                    已确认菜单治理页和消息中心测试页通过构建，建议下一步收口频道工作台的右侧成员面板和任务区。
                  </Body1>
                  <div className={styles.feedActions}>
                    <Button appearance="subtle">查看附件</Button>
                    <Button appearance="subtle">回复</Button>
                  </div>
                </div>

                <div className={styles.feedItem}>
                  <div className={styles.feedMeta}>
                    <Avatar name="Design Ops" color="colorful" />
                    <div>
                      <Body1Strong>Figma 响应式说明已补充到实验场</Body1Strong>
                      <Caption1>Design Ops · 今天 09:30</Caption1>
                    </div>
                  </div>
                  <Body1>
                    当前建议优先把协作页做成“三栏稳定布局”，避免把说明、成员和操作全部挤进主工作区。
                  </Body1>
                  <div className={styles.feedActions}>
                    <Button appearance="subtle">查看设计</Button>
                    <Button appearance="subtle">转成任务</Button>
                  </div>
                </div>
              </div>
              </article>
            </LabSurfaceCard>

            <LabSurfaceCard subtle>
              <article className={styles.panel}>
              <div className={styles.panelHeader}>
                <Body1Strong>关联任务</Body1Strong>
                <Button appearance="subtle">全部查看</Button>
              </div>
              <div className={styles.taskList}>
                <div className={styles.taskItem}>
                  <Body1Strong>完成频道页测试样式收口</Body1Strong>
                  <div className={styles.taskMeta}>
                    <Caption1>负责人：前端实验线</Caption1>
                    <Badge appearance="tint" color="warning">
                      高优先级
                    </Badge>
                  </div>
                </div>
                <div className={styles.taskItem}>
                  <Body1Strong>接入一个更具体的 Teams Frame</Body1Strong>
                  <div className={styles.taskMeta}>
                    <Caption1>负责人：设计联调</Caption1>
                    <Badge appearance="tint">中优先级</Badge>
                  </div>
                </div>
                <div className={styles.taskItem}>
                  <Body1Strong>验证移动端栏位折叠策略</Body1Strong>
                  <div className={styles.taskMeta}>
                    <Caption1>负责人：UI QA</Caption1>
                    <Badge appearance="tint">中优先级</Badge>
                  </div>
                </div>
              </div>
              </article>
            </LabSurfaceCard>
          </section>
        </main>

        <aside className={styles.rightRail}>
          <section className={styles.railSection}>
            <LabSectionTitle overline="Members" title="频道成员" description="右栏只放上下文，不重复主工作区内容。" />
            <div className={styles.memberList}>
              <div className={styles.memberItem}>
                <div className={styles.memberMeta}>
                  <Avatar name="Liu Guang" color="brand" />
                  <div>
                    <Body1Strong>刘广</Body1Strong>
                    <Caption1>项目协调</Caption1>
                  </div>
                </div>
                <Badge appearance="tint" color="success">
                  在线
                </Badge>
              </div>
              <div className={styles.memberItem}>
                <div className={styles.memberMeta}>
                  <Avatar name="Design Ops" color="colorful" />
                  <div>
                    <Body1Strong>Design Ops</Body1Strong>
                    <Caption1>设计规范</Caption1>
                  </div>
                </div>
                <Badge appearance="tint">同步中</Badge>
              </div>
            </div>
          </section>

          <Divider />

          <section className={styles.railSection}>
            <LabSectionTitle overline="Rules" title="布局约束" />
            <Caption1>右侧栏只保留成员、状态与低频信息，不让它抢主频道内容。</Caption1>
            <Caption1>频道工作台优先承载更新流与任务联动，不直接复刻聊天产品形态。</Caption1>
          </section>
        </aside>
      </div>
    </div>
  );
}
