import {
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

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '20px',
  },
  shell: {
    minHeight: '760px',
    display: 'grid',
    gridTemplateColumns: '248px minmax(0, 1fr)',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    boxShadow: tokens.shadow16,
    backgroundColor: tokens.colorNeutralBackground1,
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  sidebar: {
    display: 'grid',
    alignContent: 'start',
    gap: '18px',
    padding: '22px 18px',
    backgroundColor: tokens.colorNeutralBackground1,
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  brand: {
    display: 'grid',
    gap: '10px',
  },
  brandAccent: {
    width: '68px',
    height: '6px',
    borderRadius: tokens.borderRadiusCircular,
    backgroundColor: tokens.colorBrandBackground,
  },
  navGroup: {
    display: 'grid',
    gap: '8px',
  },
  navLabel: {
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
    textTransform: 'uppercase',
    letterSpacing: '0.04em',
  },
  navItem: {
    display: 'grid',
    gap: '4px',
    padding: '12px 12px',
    borderRadius: tokens.borderRadiusLarge,
    color: tokens.colorNeutralForeground2,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  navItemActive: {
    backgroundColor: tokens.colorBrandBackground2,
    boxShadow: tokens.shadow4,
    color: tokens.colorBrandForeground1,
  },
  content: {
    display: 'grid',
    gridTemplateRows: 'auto auto auto 1fr',
    minWidth: 0,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  topbar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '12px',
    padding: '18px 24px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 700px)': {
      flexDirection: 'column',
      alignItems: 'stretch',
    },
  },
  topbarActions: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
  hero: {
    display: 'grid',
    gap: '14px',
    padding: '24px',
  },
  heroBanner: {
    display: 'grid',
    gap: '14px',
    padding: '22px 24px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.08) 0%, rgba(15,108,189,0.03) 58%, rgba(255,255,255,0.05) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  docFrame: {
    display: 'grid',
    gap: '18px',
    padding: '22px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  docHeader: {
    display: 'grid',
    gap: '10px',
  },
  docPreview: {
    minHeight: '280px',
    borderRadius: tokens.borderRadiusLarge,
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.08) 0%, rgba(15,108,189,0.02) 45%, rgba(255,255,255,0.06) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    position: 'relative',
    overflow: 'hidden',
  },
  docPreviewBar: {
    position: 'absolute',
    inset: '36px 18px auto 18px',
    height: '22px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground3,
  },
  docPreviewPanel: {
    position: 'absolute',
    left: '18px',
    bottom: '18px',
    width: '58%',
    height: '72%',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow8,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    display: 'grid',
    gridTemplateColumns: '160px minmax(0, 1fr)',
    overflow: 'hidden',
    '@media (max-width: 720px)': {
      width: 'calc(100% - 36px)',
      gridTemplateColumns: '1fr',
    },
  },
  docPreviewRail: {
    display: 'grid',
    gap: '10px',
    padding: '16px 14px',
    backgroundColor: tokens.colorNeutralBackground2,
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  docPreviewRailItem: {
    display: 'grid',
    gap: '4px',
    padding: '8px 10px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  docPreviewBody: {
    display: 'grid',
    alignContent: 'start',
    gap: '14px',
    padding: '18px',
  },
  breadcrumb: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
  },
  copyPill: {
    width: 'fit-content',
    padding: '8px 12px',
    borderRadius: tokens.borderRadiusCircular,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow4,
    color: tokens.colorNeutralForeground1,
    fontWeight: 600,
  },
  variantsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  variantCard: {
    display: 'grid',
    gap: '12px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  variantRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
    alignItems: 'center',
  },
  heroKicker: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
  },
  summaryGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, minmax(0, 1fr))',
    gap: '14px',
    padding: '0 24px 24px',
    '@media (max-width: 1100px)': {
      gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    },
    '@media (max-width: 640px)': {
      gridTemplateColumns: '1fr',
    },
  },
  statCard: {
    display: 'grid',
    gap: '8px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  workspace: {
    display: 'grid',
    gridTemplateColumns: '1.5fr 1fr',
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
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  panelHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
  },
  chart: {
    minHeight: '260px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground3,
    backgroundImage: 'linear-gradient(180deg, rgba(15,108,189,0.18) 0%, rgba(15,108,189,0.04) 100%)',
    position: 'relative',
    overflow: 'hidden',
  },
  chartBars: {
    position: 'absolute',
    inset: '24px 20px 20px 20px',
    display: 'grid',
    gridTemplateColumns: 'repeat(7, minmax(0, 1fr))',
    alignItems: 'end',
    gap: '10px',
  },
  bar: {
    borderRadius: `${tokens.borderRadiusMedium} ${tokens.borderRadiusMedium} 0 0`,
    backgroundColor: tokens.colorBrandBackground,
    opacity: 0.9,
  },
  activityList: {
    display: 'grid',
    gap: '12px',
  },
  activityItem: {
    display: 'grid',
    gap: '4px',
    paddingBottom: '14px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroRow: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '18px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  introBlock: {
    display: 'grid',
    gap: '10px',
  },
  statusPanel: {
    display: 'grid',
    gap: '10px',
    alignContent: 'start',
  },
  ghostRow: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    },
    '@media (max-width: 640px)': {
      gridTemplateColumns: '1fr',
    },
  },
  ghostTile: {
    minHeight: '88px',
    borderRadius: tokens.borderRadiusLarge,
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.10) 0%, rgba(15,108,189,0.03) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
});

function NavItem({
  title,
  subtitle,
  active = false,
}: {
  title: string;
  subtitle: string;
  active?: boolean;
}) {
  const styles = useStyles();
  return (
    <div className={active ? `${styles.navItem} ${styles.navItemActive}` : styles.navItem}>
      <Body1Strong>{title}</Body1Strong>
      <Caption1>{subtitle}</Caption1>
    </div>
  );
}

export function AppShellShowcasePage() {
  const styles = useStyles();
  const barHeights = ['36%', '54%', '48%', '70%', '58%', '82%', '64%'];

  return (
    <div className={styles.page}>
      <Title3>应用壳 + 仪表盘测试页</Title3>
      <div className={styles.shell}>
        <aside className={styles.sidebar}>
          <div className={styles.brand}>
            <div className={styles.brandAccent} />
            <Body1Strong>Fluent 2 Control Center</Body1Strong>
            <Caption1>验证导航、摘要、主工作区和消息层级。</Caption1>
          </div>

          <div className={styles.navGroup}>
            <span className={styles.navLabel}>Workspace</span>
            <NavItem title="总览" subtitle="状态与进度" active />
            <NavItem title="团队" subtitle="成员与空间" />
            <NavItem title="审批" subtitle="风险与治理" />
          </div>

          <Divider />

          <div className={styles.navGroup}>
            <span className={styles.navLabel}>System</span>
            <NavItem title="消息中心" subtitle="订阅与告警" />
            <NavItem title="配置" subtitle="访问与策略" />
          </div>
        </aside>

        <section className={styles.content}>
          <header className={styles.topbar}>
            <div>
              <Body1Strong>Azure DevOps Services</Body1Strong>
              <Caption1>验证 Fluent 2 应用壳、侧边导航和内容工作区的层级关系。</Caption1>
            </div>
            <div className={styles.topbarActions}>
              <Button appearance="secondary">导出周报</Button>
              <Button appearance="primary">创建任务</Button>
            </div>
          </header>

          <section className={styles.hero}>
            <div className={styles.docFrame}>
              <div className={styles.heroKicker}>
                <Badge appearance="filled" color="brand">
                  App shell
                </Badge>
                <Badge appearance="tint" color="success">
                  侧栏导航
                </Badge>
                <Badge appearance="tint">内容工作区</Badge>
              </div>
              <div className={styles.docHeader}>
                <Body1Strong>一个更接近 Fluent 2 Web 文档页的应用壳样式。</Body1Strong>
                <Body1>
                  这一版把应用标题、侧边导航、面包屑、正文区和预览块拆开，重点验证壳层的节奏，而不是把所有信息塞进一个仪表盘。
                </Body1>
              </div>
              <div className={styles.breadcrumb}>
                <span>Home</span>
                <span>›</span>
                <span>Accounts</span>
                <span>›</span>
                <span>Privacy</span>
                <span>›</span>
                <span>Notifications</span>
              </div>
              <div className={styles.docPreview}>
                <div className={styles.docPreviewBar} />
                <div className={styles.docPreviewPanel}>
                  <div className={styles.docPreviewRail}>
                    <NavItem title="Home" subtitle="概览" active />
                    <NavItem title="Accounts" subtitle="权限和账户" />
                    <NavItem title="Privacy" subtitle="隐私设置" />
                    <NavItem title="Notifications" subtitle="提醒偏好" />
                  </div>
                  <div className={styles.docPreviewBody}>
                    <Subtitle2>Azure DevOps Services</Subtitle2>
                    <Body1>
                      中间区域用于承载当前视图的主内容，侧栏用于切换系统模块，顶部条用于承接全局操作。
                    </Body1>
                    <div className={styles.ghostRow}>
                      <div className={styles.ghostTile} />
                      <div className={styles.ghostTile} />
                      <div className={styles.ghostTile} />
                      <div className={styles.ghostTile} />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section className={styles.variantsGrid}>
            <article className={styles.variantCard}>
              <div className={styles.panelHeader}>
                <Body1Strong>布局收口</Body1Strong>
                <Button appearance="subtle">Copy me</Button>
              </div>
              <Body1>将壳层控制在一个稳定的面板里，突出导航与工作区的边界。</Body1>
              <div className={styles.variantRow}>
                <Badge appearance="filled" color="brand">
                  Nav
                </Badge>
                <Badge appearance="tint">Breadcrumb</Badge>
                <Badge appearance="tint">Workspace</Badge>
              </div>
            </article>
            <article className={styles.variantCard}>
              <div className={styles.panelHeader}>
                <Body1Strong>状态层级</Body1Strong>
                <Button appearance="subtle">View detail</Button>
              </div>
              <Body1>顶部命令区只放主动作，其他操作下沉到工作区或右侧详情。</Body1>
              <div className={styles.variantRow}>
                <Badge appearance="tint" color="success">
                  Stable
                </Badge>
                <Badge appearance="tint" color="warning">
                  Attention
                </Badge>
                <Badge appearance="tint" color="danger">
                  Critical
                </Badge>
              </div>
            </article>
          </section>
        </section>
      </div>
    </div>
  );
}
