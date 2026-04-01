import {
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Card,
  CardHeader,
  Divider,
  Subtitle2,
  Switch,
  Tab,
  TabList,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import { LabBadgeRow, LabStatGrid } from '../lab/primitives';

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '18px',
  },
  intro: {
    display: 'grid',
    gap: '12px',
    maxWidth: '880px',
    padding: '18px 20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.10) 0%, rgba(15,108,189,0.03) 56%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  introRow: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '14px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  introText: {
    display: 'grid',
    gap: '6px',
  },
  layout: {
    display: 'grid',
    gridTemplateColumns: '220px minmax(0, 1fr)',
    gap: '18px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  nav: {
    display: 'grid',
    gap: '12px',
    alignContent: 'start',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  navItem: {
    display: 'grid',
    gap: '4px',
    padding: '10px 12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  content: {
    display: 'grid',
    gap: '18px',
  },
  heroCard: {
    display: 'grid',
    gap: '14px',
    padding: '22px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground3,
    backgroundImage: 'linear-gradient(180deg, rgba(15,108,189,0.16) 0%, rgba(15,108,189,0.04) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  docFrame: {
    display: 'grid',
    gridTemplateColumns: '220px minmax(0, 1fr)',
    gap: '0',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    overflow: 'hidden',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  docSidebar: {
    display: 'grid',
    gap: '12px',
    padding: '18px 16px',
    backgroundColor: tokens.colorNeutralBackground2,
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  docSidebarItem: {
    display: 'grid',
    gap: '4px',
    padding: '10px 12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  docBody: {
    display: 'grid',
    gap: '18px',
    padding: '22px',
  },
  docTop: {
    display: 'grid',
    gap: '10px',
  },
  docLinks: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '12px',
  },
  previewPanel: {
    display: 'grid',
    gap: '12px',
    padding: '20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.08) 0%, rgba(15,108,189,0.02) 42%, rgba(255,255,255,0.06) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  previewChrome: {
    minHeight: '240px',
    position: 'relative',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    overflow: 'hidden',
  },
  previewHeader: {
    height: '44px',
    backgroundColor: tokens.colorNeutralBackground2,
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  previewMain: {
    position: 'absolute',
    inset: '62px 18px 18px 18px',
    display: 'grid',
    gap: '12px',
  },
  previewRow: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '12px',
    '@media (max-width: 720px)': {
      gridTemplateColumns: '1fr',
    },
  },
  previewCard: {
    minHeight: '92px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  copyPill: {
    width: 'fit-content',
    padding: '8px 12px',
    borderRadius: tokens.borderRadiusCircular,
    backgroundColor: tokens.colorNeutralForeground1,
    color: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow4,
  },
  variantSection: {
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
  variantGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  variantTile: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  cardGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '16px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  specimen: {
    display: 'grid',
    gap: '14px',
    padding: '18px',
  },
  buttonRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
  badgeRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  settingsBlock: {
    display: 'grid',
    gap: '12px',
  },
  statStrip: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 640px)': {
      gridTemplateColumns: '1fr',
    },
  },
  statCard: {
    display: 'grid',
    gap: '6px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroShell: {
    display: 'grid',
    gap: '18px',
  },
  exampleRow: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
});

export function FluentComponentGalleryPage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <header className={styles.intro}>
        <LabBadgeRow>
          <Badge appearance="filled" color="brand">
            Microsoft Fluent 2 Web
          </Badge>
          <Badge appearance="tint">组件目录</Badge>
        </LabBadgeRow>
        <div className={styles.introRow}>
          <div className={styles.introText}>
            <Title3>组件与状态展示测试页</Title3>
            <Body1>
              这张页偏向 `Microsoft Fluent 2 Web` 的组件说明与视觉收口方式，用来验证组件组合、状态表达、选项卡和轻量设置区在
              Fluent 2 下的整体质感。
            </Body1>
          </div>
          <LabStatGrid
            items={[
              { label: '状态类型', value: '6', tone: 'brand' },
              { label: '组件样例', value: '12', tone: 'success' },
              { label: '设置块', value: '4', tone: 'warning' },
            ]}
          />
        </div>
      </header>

      <div className={styles.layout}>
        <aside className={styles.nav}>
          <Body1Strong>Specimens</Body1Strong>
          <div className={styles.navItem}>
            <Body1Strong>Actions</Body1Strong>
            <Caption1>按钮、主次操作和命令区密度。</Caption1>
          </div>
          <div className={styles.navItem}>
            <Body1Strong>Status</Body1Strong>
            <Caption1>Badge、消息和状态轻重关系。</Caption1>
          </div>
          <div className={styles.navItem}>
            <Body1Strong>Configuration</Body1Strong>
            <Caption1>选项卡与设置区的组合方式。</Caption1>
          </div>
          <div className={styles.navItem}>
            <Body1Strong>Variants</Body1Strong>
            <Caption1>尺寸、形态和结构变化。</Caption1>
          </div>
        </aside>

        <main className={styles.content}>
          <section className={styles.heroShell}>
            <section className={styles.docFrame}>
              <aside className={styles.docSidebar}>
                <Body1Strong>File</Body1Strong>
                <div className={styles.docSidebarItem}>
                  <Body1Strong>Accordion</Body1Strong>
                  <Caption1>组件说明</Caption1>
                </div>
                <div className={styles.docSidebarItem}>
                  <Body1Strong>Button</Body1Strong>
                  <Caption1>动作与尺寸</Caption1>
                </div>
                <div className={styles.docSidebarItem}>
                  <Body1Strong>Calendar</Body1Strong>
                  <Caption1>复杂状态</Caption1>
                </div>
              </aside>

              <div className={styles.docBody}>
                <div className={styles.docTop}>
                  <div className={styles.docLinks}>
                    <Caption1>View documentation</Caption1>
                    <Caption1>Engineering assets</Caption1>
                  </div>
                  <Title3>Accordion</Title3>
                  <Body1>
                    An accordion groups sections of related content that can be opened and closed. This page uses a calm
                    Fluent 2 composition to mirror the reference layout.
                  </Body1>
                </div>

                <div className={styles.previewPanel}>
                  <div className={styles.copyPill}>Copy me</div>
                  <div className={styles.previewChrome}>
                    <div className={styles.previewHeader} />
                    <div className={styles.previewMain}>
                      <div className={styles.previewRow}>
                        <div className={styles.previewCard} />
                        <div className={styles.previewCard} />
                      </div>
                      <div className={styles.previewRow}>
                        <div className={styles.previewCard} />
                        <div className={styles.previewCard} />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </section>

            <div className={styles.statStrip}>
              <article className={styles.statCard}>
                <Caption1>状态类型</Caption1>
                <Body1Strong>6 种</Body1Strong>
              </article>
              <article className={styles.statCard}>
                <Caption1>组件样例</Caption1>
                <Body1Strong>12 组</Body1Strong>
              </article>
              <article className={styles.statCard}>
                <Caption1>目标</Caption1>
                <Body1Strong>统一视觉语言</Body1Strong>
              </article>
            </div>
          </section>

          <section className={styles.variantSection}>
            <div className={styles.panelHeader}>
              <Body1Strong>Variants</Body1Strong>
              <Button appearance="subtle">View more</Button>
            </div>
            <div className={styles.variantGrid}>
              <div className={styles.variantTile}>
                <Body1Strong>Chevron</Body1Strong>
                <Caption1>Medium size example</Caption1>
                <Button appearance="primary">Button</Button>
              </div>
              <div className={styles.variantTile}>
                <Body1Strong>Circular</Body1Strong>
                <Caption1>Rounded action style</Caption1>
                <Button appearance="secondary">Button</Button>
              </div>
              <div className={styles.variantTile}>
                <Body1Strong>Square</Body1Strong>
                <Caption1>Compact utility style</Caption1>
                <Button appearance="secondary">Button</Button>
              </div>
            </div>
          </section>

          <section className={styles.exampleRow}>
            <Card className={styles.specimen}>
              <CardHeader
                header={<Body1Strong>按钮与命令</Body1Strong>}
                description={<Caption1>区分主动作、次级动作和轻量命令。</Caption1>}
              />
              <div className={styles.buttonRow}>
                <Button appearance="primary">发布更新</Button>
                <Button appearance="secondary">保存草稿</Button>
                <Button appearance="subtle">查看更多</Button>
              </div>
            </Card>

            <Card className={styles.specimen}>
              <CardHeader
                header={<Body1Strong>状态表达</Body1Strong>}
                description={<Caption1>Badge 不抢主内容，但要足够清晰。</Caption1>}
              />
              <div className={styles.badgeRow}>
                <Badge appearance="filled" color="brand">
                  Brand
                </Badge>
                <Badge appearance="tint" color="success">
                  Success
                </Badge>
                <Badge appearance="tint" color="warning">
                  Warning
                </Badge>
                <Badge appearance="tint" color="danger">
                  Danger
                </Badge>
                <Badge appearance="outline">Subtle</Badge>
              </div>
            </Card>

            <Card className={styles.specimen}>
              <CardHeader
                header={<Body1Strong>分段内容</Body1Strong>}
                description={<Caption1>用 TabList 承接切换，不堆过多水平按钮。</Caption1>}
              />
              <TabList selectedValue="tokens">
                <Tab value="tokens">Tokens</Tab>
                <Tab value="patterns">Patterns</Tab>
                <Tab value="motion">Motion</Tab>
              </TabList>
              <Divider />
              <Caption1>适合把长说明拆成有限的几个主题段，而不是把所有配置塞进一个面板。</Caption1>
            </Card>

            <Card className={styles.specimen}>
              <CardHeader
                header={<Body1Strong>轻量设置区</Body1Strong>}
                description={<Caption1>Switch 和说明文本适合放在右栏或卡片底部。</Caption1>}
              />
              <div className={styles.settingsBlock}>
                <Switch label="启用高对比状态色" defaultChecked />
                <Switch label="显示说明性摘要" defaultChecked />
                <Switch label="自动折叠低频配置" />
              </div>
            </Card>
          </section>
        </main>
      </div>
    </div>
  );
}
