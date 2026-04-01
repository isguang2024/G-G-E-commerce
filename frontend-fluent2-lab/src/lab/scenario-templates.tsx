import {
  Avatar,
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Card,
  CardHeader,
  Divider,
  Subtitle2,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import { LabBadgeRow, LabRailCard, LabSectionTitle, LabStatGrid, LabSurfaceCard } from './primitives';

type ScenarioTone = 'brand' | 'success' | 'warning' | 'danger';

type ScenarioBadge = {
  label: string;
  color?: ScenarioTone;
  appearance?: 'filled' | 'tint' | 'outline';
};

type RailItem = {
  title: string;
  subtitle: string;
  active?: boolean;
};

type ScenarioItem = {
  title: string;
  description: string;
  meta: string;
  badgeText?: string;
  badgeTone?: ScenarioTone;
};

type AsideItem = {
  title: string;
  subtitle: string;
  badgeText?: string;
  badgeTone?: ScenarioTone;
};

type VariantItem = {
  title: string;
  subtitle: string;
  buttonLabel: string;
  primary?: boolean;
};

type SpecimenItem = {
  title: string;
  description: string;
  lines: string[];
};

export type FoundationScenarioConfig = {
  eyebrow: string;
  title: string;
  description: string;
  emphasis: string;
  badges: ScenarioBadge[];
  railTitle: string;
  railDescription: string;
  railItems: RailItem[];
  heroTitle: string;
  heroBody: string;
  heroSupport: string;
  stats: Array<{ label: string; value: string; tone?: 'brand' | 'success' | 'warning' }>;
  primaryTitle: string;
  primaryAction: string;
  primaryItems: ScenarioItem[];
  secondaryTitle: string;
  secondaryAction: string;
  secondaryItems: ScenarioItem[];
  asideTitle: string;
  asideDescription: string;
  asideItems: AsideItem[];
  notes: string[];
};

export type FluentScenarioConfig = {
  eyebrow: string;
  title: string;
  description: string;
  emphasis: string;
  badges: ScenarioBadge[];
  stats: Array<{ label: string; value: string; tone?: 'brand' | 'success' | 'warning' }>;
  navTitle: string;
  navItems: RailItem[];
  docTitle: string;
  docBody: string;
  docLinks: string[];
  docNotes: string[];
  variantTitle: string;
  variantAction: string;
  variants: VariantItem[];
  specimens: SpecimenItem[];
};

export type TeamsScenarioConfig = {
  eyebrow: string;
  title: string;
  description: string;
  emphasis: string;
  badges: ScenarioBadge[];
  channelTitle: string;
  channelDescription: string;
  channelItems: RailItem[];
  heroTitle: string;
  heroBody: string;
  heroSupport: string;
  stats: Array<{ label: string; value: string; tone?: 'brand' | 'success' | 'warning' }>;
  feedTitle: string;
  feedAction: string;
  feedItems: ScenarioItem[];
  taskTitle: string;
  taskAction: string;
  taskItems: ScenarioItem[];
  membersTitle: string;
  membersDescription: string;
  memberItems: AsideItem[];
  notes: string[];
};

const useFoundationStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  workspace: {
    minHeight: '760px',
    display: 'grid',
    gridTemplateColumns: '272px minmax(0, 1fr) 292px',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    boxShadow: tokens.shadow16,
    backgroundColor: tokens.colorNeutralBackground1,
    '@media (max-width: 1200px)': { gridTemplateColumns: '252px minmax(0, 1fr)' },
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  leftRail: {
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '20px 16px',
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.08) 0%, rgba(15,108,189,0.01) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  railSection: { display: 'grid', gap: '8px' },
  railLabel: {
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
    textTransform: 'uppercase',
    letterSpacing: '0.04em',
  },
  railList: { display: 'grid', gap: '6px' },
  main: { display: 'grid', gridTemplateRows: 'auto auto 1fr', minWidth: 0 },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '12px',
    padding: '20px 24px 14px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 700px)': { flexDirection: 'column', alignItems: 'stretch' },
  },
  headerActions: { display: 'flex', flexWrap: 'wrap', gap: '10px' },
  hero: { display: 'grid', gap: '12px', padding: '18px 24px 12px' },
  heroBanner: {
    display: 'grid',
    gap: '12px',
    padding: '18px 20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.12) 0%, rgba(15,108,189,0.04) 52%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroRow: {
    display: 'grid',
    gridTemplateColumns: '1.1fr 0.9fr',
    gap: '14px',
    '@media (max-width: 960px)': { gridTemplateColumns: '1fr' },
  },
  summaryBlock: { display: 'grid', gap: '6px' },
  board: {
    display: 'grid',
    gridTemplateColumns: '1.15fr 0.85fr',
    gap: '18px',
    padding: '0 24px 24px',
    '@media (max-width: 1100px)': { gridTemplateColumns: '1fr' },
  },
  panel: { display: 'grid', gap: '14px' },
  panelHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
  },
  list: { display: 'grid', gap: '12px' },
  listItem: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  metaRow: { display: 'flex', flexWrap: 'wrap', gap: '8px', alignItems: 'center' },
  rightRail: {
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '20px 18px',
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.05) 0%, rgba(15,108,189,0.00) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 1200px)': { display: 'none' },
  },
  asideList: { display: 'grid', gap: '12px' },
  asideItem: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '10px',
    alignItems: 'start',
  },
});

const useFluentStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  intro: {
    display: 'grid',
    gap: '12px',
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
    '@media (max-width: 960px)': { gridTemplateColumns: '1fr' },
  },
  introText: { display: 'grid', gap: '6px' },
  layout: {
    display: 'grid',
    gridTemplateColumns: '220px minmax(0, 1fr)',
    gap: '18px',
    '@media (max-width: 960px)': { gridTemplateColumns: '1fr' },
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
  content: { display: 'grid', gap: '18px' },
  docFrame: {
    display: 'grid',
    gridTemplateColumns: '228px minmax(0, 1fr)',
    gap: '0',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    overflow: 'hidden',
    '@media (max-width: 960px)': { gridTemplateColumns: '1fr' },
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
  docBody: { display: 'grid', gap: '18px', padding: '22px' },
  docTop: { display: 'grid', gap: '10px' },
  docLinks: { display: 'flex', flexWrap: 'wrap', gap: '12px' },
  previewPanel: {
    display: 'grid',
    gap: '12px',
    padding: '20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.08) 0%, rgba(15,108,189,0.02) 42%, rgba(255,255,255,0.06) 100%)',
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
  previewChrome: {
    minHeight: '260px',
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
    gridTemplateColumns: '1.1fr 0.9fr',
    gap: '12px',
    '@media (max-width: 720px)': { gridTemplateColumns: '1fr' },
  },
  previewCard: {
    minHeight: '92px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
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
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  variantTile: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  specimenGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '16px',
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  specimen: { display: 'grid', gap: '14px', padding: '18px' },
  specimenList: { display: 'grid', gap: '8px' },
  statStrip: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 640px)': { gridTemplateColumns: '1fr' },
  },
  statCard: {
    display: 'grid',
    gap: '6px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
});

const useTeamsStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  workspace: {
    minHeight: '760px',
    display: 'grid',
    gridTemplateColumns: '272px minmax(0, 1fr) 280px',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    boxShadow: tokens.shadow16,
    backgroundColor: tokens.colorNeutralBackground1,
    '@media (max-width: 1200px)': { gridTemplateColumns: '240px minmax(0, 1fr)' },
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  leftRail: {
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '20px 16px',
    background:
      'linear-gradient(180deg, rgba(98,100,167,0.11) 0%, rgba(98,100,167,0.02) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  railSection: { display: 'grid', gap: '8px' },
  railLabel: {
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
    textTransform: 'uppercase',
    letterSpacing: '0.04em',
  },
  channelList: { display: 'grid', gap: '6px' },
  main: { display: 'grid', gridTemplateRows: 'auto auto auto 1fr', minWidth: 0 },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '12px',
    padding: '20px 24px 14px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 700px)': { flexDirection: 'column', alignItems: 'stretch' },
  },
  headerActions: { display: 'flex', flexWrap: 'wrap', gap: '10px' },
  hero: { display: 'grid', gap: '12px', padding: '18px 24px 12px' },
  heroBanner: {
    display: 'grid',
    gap: '12px',
    padding: '18px 20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(98,100,167,0.16) 0%, rgba(98,100,167,0.05) 52%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroRow: {
    display: 'grid',
    gridTemplateColumns: '1fr 0.82fr',
    gap: '14px',
    '@media (max-width: 960px)': { gridTemplateColumns: '1fr' },
  },
  summaryBlock: { display: 'grid', gap: '6px' },
  board: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '18px',
    padding: '0 24px 24px',
    '@media (max-width: 1100px)': { gridTemplateColumns: '1fr' },
  },
  panel: { display: 'grid', gap: '14px' },
  panelHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
  },
  list: { display: 'grid', gap: '12px' },
  feedItem: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  feedMeta: { display: 'flex', gap: '10px', alignItems: 'start' },
  metaRow: { display: 'flex', flexWrap: 'wrap', gap: '8px', alignItems: 'center' },
  rightRail: {
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '20px 18px',
    background:
      'linear-gradient(180deg, rgba(98,100,167,0.08) 0%, rgba(98,100,167,0.01) 100%), var(--fluent-neutral-background-3, #f4f4f4)',
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 1200px)': { display: 'none' },
  },
  memberList: { display: 'grid', gap: '12px' },
  memberItem: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
  },
  memberMeta: { display: 'flex', gap: '10px', alignItems: 'center' },
});

function renderBadge(badge: ScenarioBadge) {
  return (
    <Badge key={badge.label} appearance={badge.appearance ?? 'tint'} color={badge.color}>
      {badge.label}
    </Badge>
  );
}

function renderItemBadge(text?: string, tone?: ScenarioTone) {
  if (!text) {
    return null;
  }
  return (
    <Badge appearance="tint" color={tone}>
      {text}
    </Badge>
  );
}

export function FoundationScenarioPage({ config }: { config: FoundationScenarioConfig }) {
  const styles = useFoundationStyles();

  return (
    <div className={styles.page}>
      <Title3>{config.title}</Title3>
      <div className={styles.workspace}>
        <aside className={styles.leftRail}>
          <section className={styles.railSection}>
            <LabSectionTitle overline={config.eyebrow} title={config.railTitle} description={config.railDescription} />
            <LabBadgeRow>{config.badges.map(renderBadge)}</LabBadgeRow>
          </section>
          <section className={styles.railSection}>
            <span className={styles.railLabel}>Focus lanes</span>
            <div className={styles.railList}>
              {config.railItems.map((item) => (
                <LabRailCard key={item.title} active={item.active}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </div>
          </section>
        </aside>
        <main className={styles.main}>
          <header className={styles.header}>
            <div>
              <Body1Strong>{config.eyebrow}</Body1Strong>
              <Caption1>{config.description}</Caption1>
            </div>
            <div className={styles.headerActions}>
              <Button appearance="secondary">查看规则</Button>
              <Button appearance="primary">打开主流程</Button>
            </div>
          </header>
          <section className={styles.hero}>
            <div className={styles.heroBanner}>
              <LabBadgeRow>{config.badges.map(renderBadge)}</LabBadgeRow>
              <div className={styles.heroRow}>
                <div className={styles.summaryBlock}>
                  <Subtitle2>{config.heroTitle}</Subtitle2>
                  <Body1>{config.heroBody}</Body1>
                </div>
                <div className={styles.summaryBlock}>
                  <Caption1>设计侧重点</Caption1>
                  <Body1Strong>{config.emphasis}</Body1Strong>
                  <Caption1>{config.heroSupport}</Caption1>
                </div>
              </div>
            </div>
            <LabStatGrid items={config.stats} />
          </section>
          <section className={styles.board}>
            <LabSurfaceCard>
              <article className={styles.panel}>
                <div className={styles.panelHeader}>
                  <Body1Strong>{config.primaryTitle}</Body1Strong>
                  <Button appearance="subtle">{config.primaryAction}</Button>
                </div>
                <div className={styles.list}>
                  {config.primaryItems.map((item) => (
                    <div key={item.title} className={styles.listItem}>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Body1>{item.description}</Body1>
                      <div className={styles.metaRow}>
                        <Caption1>{item.meta}</Caption1>
                        {renderItemBadge(item.badgeText, item.badgeTone)}
                      </div>
                    </div>
                  ))}
                </div>
              </article>
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <article className={styles.panel}>
                <div className={styles.panelHeader}>
                  <Body1Strong>{config.secondaryTitle}</Body1Strong>
                  <Button appearance="subtle">{config.secondaryAction}</Button>
                </div>
                <div className={styles.list}>
                  {config.secondaryItems.map((item) => (
                    <div key={item.title} className={styles.listItem}>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Body1>{item.description}</Body1>
                      <div className={styles.metaRow}>
                        <Caption1>{item.meta}</Caption1>
                        {renderItemBadge(item.badgeText, item.badgeTone)}
                      </div>
                    </div>
                  ))}
                </div>
              </article>
            </LabSurfaceCard>
          </section>
        </main>
        <aside className={styles.rightRail}>
          <section className={styles.railSection}>
            <LabSectionTitle overline="Context" title={config.asideTitle} description={config.asideDescription} />
            <div className={styles.asideList}>
              {config.asideItems.map((item) => (
                <div key={item.title} className={styles.asideItem}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.subtitle}</Caption1>
                  </div>
                  {renderItemBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </div>
          </section>
          <Divider />
          <section className={styles.railSection}>
            <LabSectionTitle overline="Notes" title="布局约束" />
            {config.notes.map((note) => (
              <Caption1 key={note}>{note}</Caption1>
            ))}
          </section>
        </aside>
      </div>
    </div>
  );
}

export function FluentScenarioPage({ config }: { config: FluentScenarioConfig }) {
  const styles = useFluentStyles();

  return (
    <div className={styles.page}>
      <header className={styles.intro}>
        <LabBadgeRow>{config.badges.map(renderBadge)}</LabBadgeRow>
        <div className={styles.introRow}>
          <div className={styles.introText}>
            <Title3>{config.title}</Title3>
            <Body1>{config.description}</Body1>
            <Caption1>设计侧重点：{config.emphasis}</Caption1>
          </div>
          <LabStatGrid items={config.stats} />
        </div>
      </header>
      <div className={styles.layout}>
        <aside className={styles.nav}>
          <Body1Strong>{config.navTitle}</Body1Strong>
          {config.navItems.map((item) => (
            <div key={item.title} className={styles.navItem}>
              <Body1Strong>{item.title}</Body1Strong>
              <Caption1>{item.subtitle}</Caption1>
            </div>
          ))}
        </aside>
        <main className={styles.content}>
          <section className={styles.docFrame}>
            <aside className={styles.docSidebar}>
              <Body1Strong>Documentation</Body1Strong>
              {config.navItems.map((item) => (
                <div key={item.title} className={styles.docSidebarItem}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              ))}
            </aside>
            <div className={styles.docBody}>
              <div className={styles.docTop}>
                <div className={styles.docLinks}>
                  {config.docLinks.map((link) => (
                    <Caption1 key={link}>{link}</Caption1>
                  ))}
                </div>
                <Title3>{config.docTitle}</Title3>
                <Body1>{config.docBody}</Body1>
                {config.docNotes.map((note) => (
                  <Caption1 key={note}>{note}</Caption1>
                ))}
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
            {config.stats.map((item) => (
              <article key={item.label} className={styles.statCard}>
                <Caption1>{item.label}</Caption1>
                <Body1Strong>{item.value}</Body1Strong>
              </article>
            ))}
          </div>
          <section className={styles.variantSection}>
            <div className={styles.panelHeader}>
              <Body1Strong>{config.variantTitle}</Body1Strong>
              <Button appearance="subtle">{config.variantAction}</Button>
            </div>
            <div className={styles.variantGrid}>
              {config.variants.map((variant) => (
                <div key={variant.title} className={styles.variantTile}>
                  <Body1Strong>{variant.title}</Body1Strong>
                  <Caption1>{variant.subtitle}</Caption1>
                  <Button appearance={variant.primary ? 'primary' : 'secondary'}>{variant.buttonLabel}</Button>
                </div>
              ))}
            </div>
          </section>
          <section className={styles.specimenGrid}>
            {config.specimens.map((specimen) => (
              <Card key={specimen.title} className={styles.specimen}>
                <CardHeader
                  header={<Body1Strong>{specimen.title}</Body1Strong>}
                  description={<Caption1>{specimen.description}</Caption1>}
                />
                <div className={styles.specimenList}>
                  {specimen.lines.map((line) => (
                    <Caption1 key={line}>{line}</Caption1>
                  ))}
                </div>
              </Card>
            ))}
          </section>
        </main>
      </div>
    </div>
  );
}

export function TeamsScenarioPage({ config }: { config: TeamsScenarioConfig }) {
  const styles = useTeamsStyles();

  return (
    <div className={styles.page}>
      <Title3>{config.title}</Title3>
      <div className={styles.workspace}>
        <aside className={styles.leftRail}>
          <section className={styles.railSection}>
            <LabSectionTitle overline={config.eyebrow} title={config.channelTitle} description={config.channelDescription} />
            <LabBadgeRow>{config.badges.map(renderBadge)}</LabBadgeRow>
          </section>
          <section className={styles.railSection}>
            <span className={styles.railLabel}>Channels</span>
            <div className={styles.channelList}>
              {config.channelItems.map((item) => (
                <LabRailCard key={item.title} active={item.active}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </div>
          </section>
        </aside>
        <main className={styles.main}>
          <header className={styles.header}>
            <div>
              <Body1Strong>{config.eyebrow}</Body1Strong>
              <Caption1>{config.description}</Caption1>
            </div>
            <div className={styles.headerActions}>
              <Button appearance="secondary">查看议程</Button>
              <Button appearance="primary">发起更新</Button>
            </div>
          </header>
          <section className={styles.hero}>
            <div className={styles.heroBanner}>
              <LabBadgeRow>{config.badges.map(renderBadge)}</LabBadgeRow>
              <div className={styles.heroRow}>
                <div className={styles.summaryBlock}>
                  <Subtitle2>{config.heroTitle}</Subtitle2>
                  <Body1>{config.heroBody}</Body1>
                </div>
                <div className={styles.summaryBlock}>
                  <Caption1>设计侧重点</Caption1>
                  <Body1Strong>{config.emphasis}</Body1Strong>
                  <Caption1>{config.heroSupport}</Caption1>
                </div>
              </div>
            </div>
            <LabStatGrid items={config.stats} />
          </section>
          <section className={styles.board}>
            <LabSurfaceCard>
              <article className={styles.panel}>
                <div className={styles.panelHeader}>
                  <Body1Strong>{config.feedTitle}</Body1Strong>
                  <Button appearance="subtle">{config.feedAction}</Button>
                </div>
                <div className={styles.list}>
                  {config.feedItems.map((item) => (
                    <div key={item.title} className={styles.feedItem}>
                      <div className={styles.feedMeta}>
                        <Avatar name={item.title} color="brand" />
                        <div>
                          <Body1Strong>{item.title}</Body1Strong>
                          <Caption1>{item.meta}</Caption1>
                        </div>
                      </div>
                      <Body1>{item.description}</Body1>
                      <div className={styles.metaRow}>{renderItemBadge(item.badgeText, item.badgeTone)}</div>
                    </div>
                  ))}
                </div>
              </article>
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <article className={styles.panel}>
                <div className={styles.panelHeader}>
                  <Body1Strong>{config.taskTitle}</Body1Strong>
                  <Button appearance="subtle">{config.taskAction}</Button>
                </div>
                <div className={styles.list}>
                  {config.taskItems.map((item) => (
                    <div key={item.title} className={styles.feedItem}>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Body1>{item.description}</Body1>
                      <div className={styles.metaRow}>
                        <Caption1>{item.meta}</Caption1>
                        {renderItemBadge(item.badgeText, item.badgeTone)}
                      </div>
                    </div>
                  ))}
                </div>
              </article>
            </LabSurfaceCard>
          </section>
        </main>
        <aside className={styles.rightRail}>
          <section className={styles.railSection}>
            <LabSectionTitle overline="Members" title={config.membersTitle} description={config.membersDescription} />
            <div className={styles.memberList}>
              {config.memberItems.map((item) => (
                <div key={item.title} className={styles.memberItem}>
                  <div className={styles.memberMeta}>
                    <Avatar name={item.title} color="colorful" />
                    <div>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Caption1>{item.subtitle}</Caption1>
                    </div>
                  </div>
                  {renderItemBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </div>
          </section>
          <Divider />
          <section className={styles.railSection}>
            <LabSectionTitle overline="Rules" title="布局约束" />
            {config.notes.map((note) => (
              <Caption1 key={note}>{note}</Caption1>
            ))}
          </section>
        </aside>
      </div>
    </div>
  );
}
