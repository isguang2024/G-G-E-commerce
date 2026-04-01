import {
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Link,
  Subtitle2,
  Title1,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

const figmaUrl =
  'https://www.figma.com/design/GFS6tbMoqoyB5MoNO555k9/Microsoft-Teams-UI-Kit--Community-?node-id=6150-147&p=f&t=0jwx28q76xvg37f0-0';

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '20px',
  },
  masthead: {
    display: 'grid',
    gap: '12px',
    paddingTop: '8px',
  },
  mastheadMeta: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
  },
  layout: {
    display: 'grid',
    gridTemplateColumns: '240px minmax(0, 1fr) 220px',
    gap: '20px',
    alignItems: 'start',
    '@media (max-width: 1200px)': {
      gridTemplateColumns: '220px minmax(0, 1fr)',
    },
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  rail: {
    display: 'grid',
    gap: '16px',
    paddingTop: '8px',
  },
  railCard: {
    display: 'grid',
    gap: '10px',
    paddingTop: '16px',
    paddingRight: '16px',
    paddingBottom: '16px',
    paddingLeft: '16px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    borderTopWidth: '1px',
    borderRightWidth: '1px',
    borderBottomWidth: '1px',
    borderLeftWidth: '1px',
    borderTopStyle: 'solid',
    borderRightStyle: 'solid',
    borderBottomStyle: 'solid',
    borderLeftStyle: 'solid',
    borderTopColor: tokens.colorNeutralStroke2,
    borderRightColor: tokens.colorNeutralStroke2,
    borderBottomColor: tokens.colorNeutralStroke2,
    borderLeftColor: tokens.colorNeutralStroke2,
  },
  railLabel: {
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
    lineHeight: tokens.lineHeightBase200,
    textTransform: 'uppercase',
    letterSpacing: '0.04em',
  },
  navList: {
    display: 'grid',
    gap: '4px',
  },
  navItem: {
    display: 'grid',
    gap: '4px',
    paddingTop: '8px',
    paddingRight: '10px',
    paddingBottom: '8px',
    paddingLeft: '10px',
    borderRadius: tokens.borderRadiusMedium,
    color: tokens.colorNeutralForeground2,
  },
  navItemActive: {
    backgroundColor: tokens.colorBrandBackground2,
    color: tokens.colorBrandForeground1,
    fontWeight: tokens.fontWeightSemibold,
  },
  nestedList: {
    display: 'grid',
    gap: '4px',
    paddingLeft: '10px',
  },
  content: {
    display: 'grid',
    gap: '24px',
    minWidth: 0,
  },
  articleCard: {
    display: 'grid',
    gap: '16px',
    paddingTop: '24px',
    paddingRight: '24px',
    paddingBottom: '24px',
    paddingLeft: '24px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow4,
  },
  articleLead: {
    maxWidth: '720px',
    color: tokens.colorNeutralForeground2,
  },
  sectionGrid: {
    display: 'grid',
    gap: '16px',
  },
  previewGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '16px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  previewCard: {
    display: 'grid',
    gap: '14px',
    paddingTop: '18px',
    paddingRight: '18px',
    paddingBottom: '18px',
    paddingLeft: '18px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    borderTopWidth: '1px',
    borderRightWidth: '1px',
    borderBottomWidth: '1px',
    borderLeftWidth: '1px',
    borderTopStyle: 'solid',
    borderRightStyle: 'solid',
    borderBottomStyle: 'solid',
    borderLeftStyle: 'solid',
    borderTopColor: tokens.colorNeutralStroke2,
    borderRightColor: tokens.colorNeutralStroke2,
    borderBottomColor: tokens.colorNeutralStroke2,
    borderLeftColor: tokens.colorNeutralStroke2,
  },
  previewHeader: {
    display: 'grid',
    gap: '4px',
  },
  phoneFrame: {
    width: '168px',
    height: '312px',
    borderRadius: '18px',
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow8,
    paddingTop: '14px',
    paddingRight: '12px',
    paddingBottom: '14px',
    paddingLeft: '12px',
    display: 'grid',
    gridTemplateRows: 'auto 1fr auto',
    gap: '12px',
  },
  phoneTopBar: {
    display: 'grid',
    gap: '6px',
  },
  phoneAvatarRow: {
    display: 'flex',
    gap: '8px',
    alignItems: 'center',
  },
  avatar: {
    width: '14px',
    height: '14px',
    borderRadius: '999px',
    backgroundColor: tokens.colorNeutralBackground4,
  },
  line: {
    height: '6px',
    borderRadius: tokens.borderRadiusCircular,
    backgroundColor: tokens.colorNeutralBackground4,
  },
  lineBrand: {
    backgroundColor: tokens.colorBrandBackground,
  },
  lineShort: {
    width: '40%',
  },
  lineMedium: {
    width: '58%',
  },
  phoneBody: {
    display: 'grid',
    alignContent: 'start',
    gap: '10px',
  },
  phoneNav: {
    display: 'grid',
    gridTemplateColumns: 'repeat(5, 1fr)',
    gap: '8px',
    alignItems: 'end',
  },
  navDot: {
    height: '8px',
    borderRadius: tokens.borderRadiusCircular,
    backgroundColor: tokens.colorNeutralBackground4,
  },
  desktopFrame: {
    width: 'min(100%, 440px)',
    minHeight: '316px',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    boxShadow: tokens.shadow8,
    backgroundColor: tokens.colorNeutralBackground1,
    display: 'grid',
    gridTemplateRows: '40px 1fr',
  },
  desktopTopBar: {
    backgroundColor: tokens.colorBrandBackground2,
    display: 'grid',
    gridTemplateColumns: '180px 1fr 28px',
    alignItems: 'center',
    gap: '12px',
    paddingTop: '0',
    paddingRight: '16px',
    paddingBottom: '0',
    paddingLeft: '16px',
  },
  desktopBody: {
    display: 'grid',
    gridTemplateColumns: '52px 1fr',
    minHeight: 0,
  },
  desktopSideRail: {
    backgroundColor: tokens.colorNeutralBackground3,
    display: 'grid',
    alignContent: 'start',
    gap: '10px',
    paddingTop: '14px',
    paddingRight: '10px',
    paddingBottom: '14px',
    paddingLeft: '10px',
  },
  sideBlock: {
    height: '18px',
    borderRadius: tokens.borderRadiusMedium,
    backgroundColor: tokens.colorNeutralBackground5,
  },
  desktopContent: {
    display: 'grid',
    gap: '14px',
    paddingTop: '16px',
    paddingRight: '18px',
    paddingBottom: '18px',
    paddingLeft: '18px',
  },
  heroBox: {
    minHeight: '92px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground3,
  },
  listBlock: {
    display: 'grid',
    gap: '10px',
  },
  ruleGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  ruleCard: {
    display: 'grid',
    gap: '10px',
    paddingTop: '16px',
    paddingRight: '16px',
    paddingBottom: '16px',
    paddingLeft: '16px',
    borderRadius: tokens.borderRadiusLarge,
    borderTopWidth: '1px',
    borderRightWidth: '1px',
    borderBottomWidth: '1px',
    borderLeftWidth: '1px',
    borderTopStyle: 'solid',
    borderRightStyle: 'solid',
    borderBottomStyle: 'solid',
    borderLeftStyle: 'solid',
    borderTopColor: tokens.colorNeutralStroke2,
    borderRightColor: tokens.colorNeutralStroke2,
    borderBottomColor: tokens.colorNeutralStroke2,
    borderLeftColor: tokens.colorNeutralStroke2,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  aside: {
    display: 'grid',
    gap: '16px',
    '@media (max-width: 1200px)': {
      gridColumnStart: '1',
      gridColumnEnd: '3',
    },
    '@media (max-width: 900px)': {
      gridColumnStart: 'auto',
      gridColumnEnd: 'auto',
    },
  },
  bulletList: {
    display: 'grid',
    gap: '10px',
    padding: 0,
    margin: 0,
    listStyleType: 'none',
  },
  bulletItem: {
    color: tokens.colorNeutralForeground2,
  },
  actionRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
});

function SidebarItem({
  label,
  active = false,
}: {
  label: string;
  active?: boolean;
}) {
  const styles = useStyles();
  return (
    <div className={active ? `${styles.navItem} ${styles.navItemActive}` : styles.navItem}>
      {label}
    </div>
  );
}

export function TeamsResponsivenessPage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <header className={styles.masthead}>
        <div className={styles.mastheadMeta}>
          <Badge appearance="filled" color="brand">
            Figma 参考页
          </Badge>
          <Badge appearance="tint">Microsoft Teams UI Kit</Badge>
          <Caption1>节点参考：`6150:147`</Caption1>
        </div>
        <Title1>Responsiveness</Title1>
        <Body1 className={styles.articleLead}>
          这张测试页不是复刻 Figma 编辑器，而是把 Teams UI Kit 里的三栏信息架构、响应式说明方式和设备预览语言抽出来，
          用 Fluent UI React v9 重组为一个可继续扩展的测试页面。
        </Body1>
        <div className={styles.actionRow}>
          <Button appearance="primary" as="a" href={figmaUrl} target="_blank">
            打开 Figma 源稿
          </Button>
          <Button appearance="secondary">切换到移动端视角</Button>
        </div>
      </header>

      <div className={styles.layout}>
        <aside className={styles.rail}>
          <section className={styles.railCard}>
            <span className={styles.railLabel}>General Information</span>
            <div className={styles.navList}>
              <SidebarItem label="Get started here" />
              <SidebarItem label="Introduction" />
              <SidebarItem label="Change log" />
            </div>
          </section>

          <section className={styles.railCard}>
            <span className={styles.railLabel}>Design System</span>
            <div className={styles.navList}>
              <SidebarItem label="Layout & Scaling" />
              <div className={styles.nestedList}>
                <SidebarItem label="Best practices" />
                <SidebarItem label="Responsiveness" active />
                <SidebarItem label="Guidelines" />
              </div>
              <SidebarItem label="Overview" />
            </div>
          </section>
        </aside>

        <main className={styles.content}>
          <section className={styles.articleCard}>
            <div className={styles.sectionGrid}>
              <Title3>响应式原则</Title3>
              <Body1 className={styles.articleLead}>
                响应式设计的重点不是单纯压缩布局，而是保证文本、操作和结构在不同窗口宽度下依然清晰。当前测试页沿用
                Teams UI Kit 的表达方式，把“移动端最低宽度”和“桌面端最低宽度”拆成两个独立预览区。
              </Body1>
            </div>

            <div className={styles.previewGrid}>
              <section className={styles.previewCard}>
                <div className={styles.previewHeader}>
                  <Subtitle2>Mobile</Subtitle2>
                  <Caption1>最小阅读宽度建议以 320 px 为起点，优先保留标题、主操作和底部导航的可见性。</Caption1>
                </div>
                <div className={styles.phoneFrame}>
                  <div className={styles.phoneTopBar}>
                    <div className={styles.phoneAvatarRow}>
                      <div className={styles.avatar} />
                      <div className={`${styles.line} ${styles.lineMedium}`} />
                    </div>
                    <div className={`${styles.line} ${styles.lineBrand} ${styles.lineShort}`} />
                  </div>
                  <div className={styles.phoneBody}>
                    <div className={`${styles.line} ${styles.lineMedium}`} />
                    <div className={`${styles.line} ${styles.lineShort}`} />
                    <div className={`${styles.line} ${styles.lineMedium}`} />
                    <div className={`${styles.line} ${styles.lineShort}`} />
                  </div>
                  <div className={styles.phoneNav}>
                    <div className={styles.navDot} />
                    <div className={styles.navDot} />
                    <div className={styles.navDot} />
                    <div className={styles.navDot} />
                    <div className={styles.navDot} />
                  </div>
                </div>
              </section>

              <section className={styles.previewCard}>
                <div className={styles.previewHeader}>
                  <Subtitle2>Desktop</Subtitle2>
                  <Caption1>桌面端保留左侧功能切换和主内容并排结构，但低频区块可压成更稳定的侧边栏。</Caption1>
                </div>
                <div className={styles.desktopFrame}>
                  <div className={styles.desktopTopBar}>
                    <div className={`${styles.line} ${styles.lineMedium}`} />
                    <div className={`${styles.line} ${styles.lineShort}`} />
                    <div className={styles.avatar} />
                  </div>
                  <div className={styles.desktopBody}>
                    <div className={styles.desktopSideRail}>
                      <div className={styles.sideBlock} />
                      <div className={styles.sideBlock} />
                      <div className={styles.sideBlock} />
                      <div className={styles.sideBlock} />
                      <div className={styles.sideBlock} />
                    </div>
                    <div className={styles.desktopContent}>
                      <div className={styles.heroBox} />
                      <div className={styles.listBlock}>
                        <div className={`${styles.line} ${styles.lineMedium}`} />
                        <div className={`${styles.line} ${styles.lineShort}`} />
                        <div className={`${styles.line} ${styles.lineMedium}`} />
                        <div className={`${styles.line} ${styles.lineShort}`} />
                        <div className={`${styles.line} ${styles.lineMedium}`} />
                      </div>
                    </div>
                  </div>
                </div>
              </section>
            </div>

            <div className={styles.ruleGrid}>
              <article className={styles.ruleCard}>
                <Badge appearance="tint" color="success">
                  建议保留
                </Badge>
                <Body1Strong>主信息优先级稳定</Body1Strong>
                <Body1>
                  标题、摘要和主 CTA 在窄屏下优先保留，不让用户依赖横向滚动或缩放才能完成主要任务。
                </Body1>
              </article>
              <article className={styles.ruleCard}>
                <Badge appearance="tint" color="danger">
                  避免出现
                </Badge>
                <Body1Strong>文本与操作互相挤压</Body1Strong>
                <Body1>
                  如果缩放后文本被截断、按钮覆盖内容，说明布局只是在缩小尺寸，而不是在真正响应变化。
                </Body1>
              </article>
            </div>
          </section>
        </main>

        <aside className={styles.aside}>
          <section className={styles.railCard}>
            <span className={styles.railLabel}>Source</span>
            <Body1Strong>Microsoft Teams UI Kit</Body1Strong>
            <Caption1>Figma Community 官方设计文件</Caption1>
            <Link href={figmaUrl} target="_blank">
              查看设计文件
            </Link>
          </section>

          <section className={styles.railCard}>
            <span className={styles.railLabel}>提炼重点</span>
            <ul className={styles.bulletList}>
              <li className={styles.bulletItem}>用左侧目录表达章节层级，不把说明文本堆成一整页。</li>
              <li className={styles.bulletItem}>主内容区只承载当前主题，移动端与桌面端分开展示。</li>
              <li className={styles.bulletItem}>次级侧栏负责来源、规则摘要和跳转，不抢主标题层级。</li>
            </ul>
          </section>

          <section className={styles.railCard}>
            <span className={styles.railLabel}>下一步</span>
            <ul className={styles.bulletList}>
              <li className={styles.bulletItem}>把这张参考页扩成完整应用壳测试页。</li>
              <li className={styles.bulletItem}>从同一套 Figma 文件中继续抽频道页或消息页的具体 Frame。</li>
              <li className={styles.bulletItem}>把设备预览替换成更接近真实业务内容的 Fluent 组件组合。</li>
            </ul>
          </section>
        </aside>
      </div>
    </div>
  );
}
