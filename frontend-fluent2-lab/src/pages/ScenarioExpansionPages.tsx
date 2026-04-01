import {
  Avatar,
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Divider,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import { foundationScenarioData, fluentScenarioData, teamsScenarioData } from '../lab/scenario-data';
import type {
  FluentScenarioConfig,
  FoundationScenarioConfig,
  TeamsScenarioConfig,
} from '../lab/scenario-templates';
import { LabBadgeRow, LabRailCard, LabSectionTitle, LabStatGrid, LabSurfaceCard } from '../lab/primitives';

type Tone = 'brand' | 'success' | 'warning' | 'danger';

const useStyles = makeStyles({
  page: { display: 'grid', gap: '18px' },
  shell: {
    display: 'grid',
    gap: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow8,
    overflow: 'hidden',
  },
  hero: {
    display: 'grid',
    gap: '12px',
    padding: '20px 22px 16px',
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.12) 0%, rgba(15,108,189,0.04) 56%, rgba(255,255,255,0.02) 100%)',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroTeams: {
    background:
      'linear-gradient(135deg, rgba(98,100,167,0.16) 0%, rgba(98,100,167,0.05) 56%, rgba(255,255,255,0.02) 100%)',
  },
  heroWarn: {
    background:
      'linear-gradient(135deg, rgba(188,96,0,0.14) 0%, rgba(188,96,0,0.05) 56%, rgba(255,255,255,0.02) 100%)',
  },
  teamsSpotlight: {
    display: 'grid',
    gap: '12px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(98,100,167,0.18) 0%, rgba(98,100,167,0.06) 55%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  headerRow: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '16px',
    flexWrap: 'wrap',
  },
  heroText: { display: 'grid', gap: '8px', maxWidth: '900px' },
  actions: { display: 'flex', gap: '10px', flexWrap: 'wrap' },
  triptych: {
    display: 'grid',
    gridTemplateColumns: '250px minmax(0, 1fr) 280px',
    minHeight: '760px',
    '@media (max-width: 1180px)': { gridTemplateColumns: '240px minmax(0, 1fr)' },
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  split: {
    display: 'grid',
    gridTemplateColumns: '1.15fr 0.85fr',
    gap: '16px',
    '@media (max-width: 980px)': { gridTemplateColumns: '1fr' },
  },
  wideSplit: {
    display: 'grid',
    gridTemplateColumns: '1.35fr 0.65fr',
    gap: '16px',
    '@media (max-width: 980px)': { gridTemplateColumns: '1fr' },
  },
  explorer: {
    display: 'grid',
    gridTemplateColumns: '280px minmax(0, 1fr)',
    gap: '16px',
    '@media (max-width: 980px)': { gridTemplateColumns: '1fr' },
  },
  doc: {
    display: 'grid',
    gridTemplateColumns: '220px minmax(0, 1fr)',
    gap: '16px',
    '@media (max-width: 980px)': { gridTemplateColumns: '1fr' },
  },
  rail: {
    display: 'grid',
    gap: '12px',
    alignContent: 'start',
    padding: '18px 16px',
    backgroundColor: tokens.colorNeutralBackground2,
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  rightRail: {
    display: 'grid',
    gap: '12px',
    alignContent: 'start',
    padding: '18px 16px',
    backgroundColor: tokens.colorNeutralBackground2,
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 1180px)': { display: 'none' },
  },
  content: { display: 'grid', gap: '16px', padding: '18px 20px 22px', minWidth: 0 },
  list: { display: 'grid', gap: '12px' },
  item: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  spotlight: {
    display: 'grid',
    gap: '10px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.12) 0%, rgba(15,108,189,0.03) 52%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  row: { display: 'flex', flexWrap: 'wrap', gap: '8px', alignItems: 'center' },
  quad: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  board: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 980px)': { gridTemplateColumns: '1fr' },
  },
  teamsBoard: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 1080px)': { gridTemplateColumns: '1fr' },
  },
  phaseStrip: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 900px)': { gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' },
    '@media (max-width: 640px)': { gridTemplateColumns: '1fr' },
  },
  tableRow: {
    display: 'grid',
    gridTemplateColumns: '1.1fr 0.7fr auto',
    gap: '12px',
    alignItems: 'center',
    padding: '12px 14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    '@media (max-width: 760px)': { gridTemplateColumns: '1fr' },
  },
  cardGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 980px)': { gridTemplateColumns: '1fr' },
  },
  specimen: {
    display: 'grid',
    gap: '10px',
    minHeight: '170px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(180deg, rgba(15,108,189,0.08) 0%, rgba(15,108,189,0.02) 48%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  agendaStrip: {
    display: 'grid',
    gridTemplateColumns: 'repeat(5, minmax(0, 1fr))',
    gap: '10px',
    '@media (max-width: 980px)': { gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' },
    '@media (max-width: 640px)': { gridTemplateColumns: '1fr' },
  },
  agendaCard: {
    display: 'grid',
    gap: '6px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    borderLeft: `4px solid ${tokens.colorBrandStroke1}`,
  },
  teamsChecklist: {
    display: 'grid',
    gap: '10px',
  },
  memberStrip: {
    display: 'flex',
    gap: '10px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  memberChip: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '8px 10px',
    borderRadius: tokens.borderRadiusCircular,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  stageGrid: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '16px',
    '@media (max-width: 980px)': { gridTemplateColumns: '1fr' },
  },
  stageFrame: {
    display: 'grid',
    gap: '12px',
    minHeight: '280px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(180deg, rgba(98,100,167,0.14) 0%, rgba(98,100,167,0.03) 52%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  fileShelf: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 760px)': { gridTemplateColumns: '1fr' },
  },
  fileCard: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    minHeight: '140px',
  },
  bulletinFeed: {
    display: 'grid',
    gap: '12px',
  },
  bulletinCard: {
    display: 'grid',
    gap: '10px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    borderTop: `3px solid ${tokens.colorBrandStroke1}`,
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  handoffBoard: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 1100px)': { gridTemplateColumns: '1fr' },
  },
  specimenRow: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '10px',
    '@media (max-width: 720px)': { gridTemplateColumns: '1fr' },
  },
  specimenBox: {
    minHeight: '68px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  thread: {
    display: 'grid',
    gridTemplateColumns: '270px minmax(0, 1fr) 270px',
    minHeight: '760px',
    '@media (max-width: 1180px)': { gridTemplateColumns: '260px minmax(0, 1fr)' },
    '@media (max-width: 900px)': { gridTemplateColumns: '1fr' },
  },
  threadList: {
    display: 'grid',
    gap: '10px',
    padding: '18px 16px',
    backgroundColor: tokens.colorNeutralBackground2,
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  post: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  author: { display: 'flex', alignItems: 'center', gap: '10px' },
});

function toneBadge(text?: string, tone?: Tone) {
  return text ? (
    <Badge appearance="tint" color={tone}>
      {text}
    </Badge>
  ) : null;
}

function Header({
  eyebrow,
  title,
  description,
  emphasis,
  badges,
  stats,
  actions,
  tone = 'brand',
}: {
  eyebrow: string;
  title: string;
  description: string;
  emphasis: string;
  badges: Array<{ label: string; color?: Tone; appearance?: 'filled' | 'tint' | 'outline' }>;
  stats: FoundationScenarioConfig['stats'] | FluentScenarioConfig['stats'] | TeamsScenarioConfig['stats'];
  actions: string[];
  tone?: 'brand' | 'teams' | 'warn';
}) {
  const styles = useStyles();
  return (
    <div className={[styles.hero, tone === 'teams' ? styles.heroTeams : '', tone === 'warn' ? styles.heroWarn : ''].join(' ')}>
      <LabBadgeRow>
        {badges.map(badge => (
          <Badge key={badge.label} appearance={badge.appearance ?? 'tint'} color={badge.color}>
            {badge.label}
          </Badge>
        ))}
      </LabBadgeRow>
      <div className={styles.headerRow}>
        <div className={styles.heroText}>
          <Caption1>{eyebrow}</Caption1>
          <Title3>{title}</Title3>
          <Body1>{description}</Body1>
          <Caption1>设计侧重点：{emphasis}</Caption1>
        </div>
        <div className={styles.actions}>
          {actions.map((action, index) => (
            <Button key={action} appearance={index === actions.length - 1 ? 'primary' : 'secondary'}>
              {action}
            </Button>
          ))}
        </div>
      </div>
      <LabStatGrid items={stats as FoundationScenarioConfig['stats']} />
    </div>
  );
}

function FoundationDeck({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看规则', '打开主流程']} />
        <div className={styles.triptych}>
          <aside className={styles.rail}>
            <LabSectionTitle overline="Focus lanes" title={config.railTitle} description={config.railDescription} />
            {config.railItems.map(item => (
              <LabRailCard key={item.title} active={item.active}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </LabRailCard>
            ))}
          </aside>
          <main className={styles.content}>
            <div className={styles.split}>
              <LabSurfaceCard>
                <LabSectionTitle title={config.primaryTitle} />
                <div className={styles.list}>
                  {config.primaryItems.map(item => (
                    <div key={item.title} className={styles.item}>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Body1>{item.description}</Body1>
                      <div className={styles.row}>
                        <Caption1>{item.meta}</Caption1>
                        {toneBadge(item.badgeText, item.badgeTone)}
                      </div>
                    </div>
                  ))}
                </div>
              </LabSurfaceCard>
              <LabSurfaceCard subtle>
                <LabSectionTitle title={config.secondaryTitle} />
                <div className={styles.list}>
                  {config.secondaryItems.map(item => (
                    <div key={item.title} className={styles.item}>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Body1>{item.description}</Body1>
                      <div className={styles.row}>
                        <Caption1>{item.meta}</Caption1>
                        {toneBadge(item.badgeText, item.badgeTone)}
                      </div>
                    </div>
                  ))}
                </div>
              </LabSurfaceCard>
            </div>
          </main>
          <aside className={styles.rightRail}>
            <LabSectionTitle overline="Context" title={config.asideTitle} description={config.asideDescription} />
            {config.asideItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
                {toneBadge(item.badgeText, item.badgeTone)}
              </div>
            ))}
          </aside>
        </div>
      </div>
    </div>
  );
}

function FoundationWorkbench({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['刷新队列', '发起动作']} />
        <div className={styles.content}>
          <div className={styles.row}>
            {config.badges.slice(0, 3).map(badge => (
              <div key={badge.label} className={styles.item}>
                <Caption1>{badge.label}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title={config.primaryTitle} description="工作台类页面优先呈现队列和处理节奏。" />
              <div className={styles.list}>
                {config.primaryItems.map(item => (
                  <div key={item.title} className={styles.tableRow}>
                    <div>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Caption1>{item.description}</Caption1>
                    </div>
                    <Caption1>{item.meta}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                ))}
              </div>
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title={config.secondaryTitle} description="详情区承接影响面、规则和下一步动作。" />
              <div className={styles.list}>
                {config.notes.map(note => (
                  <div key={note} className={styles.item}>
                    <Body1>{note}</Body1>
                  </div>
                ))}
                {config.secondaryItems.map(item => (
                  <div key={item.title} className={styles.item}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                ))}
              </div>
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function FoundationExplorer({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开检索', '查看目录']} />
        <div className={styles.content}>
          <div className={styles.explorer}>
            <aside className={styles.rail}>
              <LabSectionTitle overline="Filters" title={config.railTitle} description={config.railDescription} />
              {config.railItems.map(item => (
                <LabRailCard key={item.title} active={item.active}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              <div className={styles.cardGrid}>
                {[...config.primaryItems, ...config.secondaryItems].map(item => (
                  <LabSurfaceCard key={item.title}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Body1>{item.description}</Body1>
                    <div className={styles.row}>
                      <Caption1>{item.meta}</Caption1>
                      {toneBadge(item.badgeText, item.badgeTone)}
                    </div>
                  </LabSurfaceCard>
                ))}
              </div>
              <LabSurfaceCard subtle>
                <LabSectionTitle title={config.asideTitle} description={config.asideDescription} />
                {config.asideItems.map(item => (
                  <div key={item.title} className={styles.item}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.subtitle}</Caption1>
                  </div>
                ))}
              </LabSurfaceCard>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function FluentDocs({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看规范', '复制示例']} />
        <div className={styles.content}>
          <div className={styles.doc}>
            <aside className={styles.rail}>
              <LabSectionTitle overline="Documentation" title={config.navTitle} />
              {config.navItems.map(item => (
                <LabRailCard key={item.title}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              <LabSurfaceCard>
                <LabSectionTitle title={config.docTitle} description={config.docBody} />
                {config.docNotes.map(note => (
                  <Caption1 key={note}>{note}</Caption1>
                ))}
              </LabSurfaceCard>
              <div className={styles.specimen}>
                <div className={styles.row}>
                  <Body1Strong>Specimen</Body1Strong>
                  <Badge appearance="filled" color="brand">Copy me</Badge>
                </div>
                <div className={styles.specimenRow}>
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                </div>
                <div className={styles.specimenRow}>
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function FluentPatterns({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开样例', '查看变体']} />
        <div className={styles.content}>
          <div className={styles.quad}>
            {config.variants.map(item => (
              <LabSurfaceCard key={item.title}>
                <LabSectionTitle title={item.title} description={item.subtitle} />
                <div className={styles.specimenRow}>
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                </div>
                <Button appearance={item.primary ? 'primary' : 'secondary'}>{item.buttonLabel}</Button>
              </LabSurfaceCard>
            ))}
          </div>
          <LabSurfaceCard subtle>
            <LabSectionTitle title={config.variantTitle} description="模式页强调不同使用场景，而不是继续复用规范页排版。" />
            <div className={styles.list}>
              {config.specimens.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {item.lines.map(line => (
                    <Body1 key={line}>{line}</Body1>
                  ))}
                </div>
              ))}
            </div>
          </LabSurfaceCard>
        </div>
      </div>
    </div>
  );
}

function FluentReview({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['导出评审', '生成结论']} tone="warn" />
        <div className={styles.content}>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle overline="Checklist" title={config.navTitle} description="评审类页面突出检查项和主要发现。" />
              {config.navItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
            <div className={styles.list}>
              <LabSurfaceCard subtle>
                <LabSectionTitle title={config.docTitle} description={config.docBody} />
                {config.docNotes.map(note => (
                  <div key={note} className={styles.item}>
                    <Body1>{note}</Body1>
                  </div>
                ))}
              </LabSurfaceCard>
              <LabSurfaceCard>
                <LabSectionTitle title={config.variantTitle} />
                {config.specimens.map(item => (
                  <div key={item.title} className={styles.item}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                ))}
              </LabSurfaceCard>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function TeamsChannel({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看议程', '发起更新']} tone="teams" />
        <div className={styles.triptych}>
          <aside className={styles.rail}>
            <LabSectionTitle overline="Channels" title={config.channelTitle} description={config.channelDescription} />
            {config.channelItems.map(item => (
              <LabRailCard key={item.title} active={item.active}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </LabRailCard>
            ))}
          </aside>
          <main className={styles.content}>
            <div className={styles.split}>
              <LabSurfaceCard>
                <LabSectionTitle title={config.feedTitle} />
                <div className={styles.list}>
                  {config.feedItems.map(item => (
                    <div key={item.title} className={styles.item}>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Body1>{item.description}</Body1>
                      <div className={styles.row}>
                        <Caption1>{item.meta}</Caption1>
                        {toneBadge(item.badgeText, item.badgeTone)}
                      </div>
                    </div>
                  ))}
                </div>
              </LabSurfaceCard>
              <LabSurfaceCard subtle>
                <LabSectionTitle title={config.taskTitle} />
                <div className={styles.list}>
                  {config.taskItems.map(item => (
                    <div key={item.title} className={styles.item}>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Body1>{item.description}</Body1>
                      <div className={styles.row}>
                        <Caption1>{item.meta}</Caption1>
                        {toneBadge(item.badgeText, item.badgeTone)}
                      </div>
                    </div>
                  ))}
                </div>
              </LabSurfaceCard>
            </div>
          </main>
          <aside className={styles.rightRail}>
            <LabSectionTitle overline="People" title={config.membersTitle} description={config.membersDescription} />
            {config.memberItems.map(item => (
              <div key={item.title} className={styles.item}>
                <div className={styles.author}>
                  <Avatar name={item.title} color="colorful" />
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.subtitle}</Caption1>
                  </div>
                </div>
                {toneBadge(item.badgeText, item.badgeTone)}
              </div>
            ))}
          </aside>
        </div>
      </div>
    </div>
  );
}

function TeamsThread({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['展开线程', '发送总结']} tone="teams" />
        <div className={styles.thread}>
          <aside className={styles.threadList}>
            <LabSectionTitle overline="Threads" title={config.channelTitle} description="线程页强调讨论列表和上下文，而不是频道公告。" />
            {config.feedItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.meta}</Caption1>
              </div>
            ))}
          </aside>
          <main className={styles.content}>
            {config.feedItems.map(item => (
              <div key={item.title} className={styles.post}>
                <div className={styles.author}>
                  <Avatar name={item.title} color="brand" />
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.meta}</Caption1>
                  </div>
                </div>
                <Body1>{item.description}</Body1>
                {toneBadge(item.badgeText, item.badgeTone)}
              </div>
            ))}
          </main>
          <aside className={styles.rightRail}>
            <LabSectionTitle overline="Context" title={config.taskTitle} description="右侧承接审批、文件或协同上下文。" />
            {config.taskItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.description}</Caption1>
                {toneBadge(item.badgeText, item.badgeTone)}
              </div>
            ))}
          </aside>
        </div>
      </div>
    </div>
  );
}

function TeamsBulletin({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看置顶', '安排回应']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle overline="Pinned" title={config.heroTitle} description={config.heroBody} />
              <div className={styles.list}>
                {config.feedItems.map(item => (
                  <div key={item.title} className={styles.item}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Body1>{item.description}</Body1>
                    <div className={styles.row}>
                      <Caption1>{item.meta}</Caption1>
                      {toneBadge(item.badgeText, item.badgeTone)}
                    </div>
                  </div>
                ))}
              </div>
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle overline="Operations" title={config.taskTitle} description={config.heroSupport} />
              <div className={styles.list}>
                {config.taskItems.map(item => (
                  <div key={item.title} className={styles.item}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                ))}
              </div>
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function OperationsCommandHero({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['导出简报', '进入指挥模式']} />
        <div className={styles.content}>
          <div className={styles.quad}>
            {config.stats.map(item => (
              <LabSurfaceCard key={item.label}>
                <Caption1>{item.label}</Caption1>
                <Title3>{item.value}</Title3>
                {item.tone ? toneBadge(item.tone === 'brand' ? '进行中' : item.tone === 'success' ? '稳定' : '需关注', item.tone) : null}
              </LabSurfaceCard>
            ))}
            <LabSurfaceCard subtle>
              <LabSectionTitle title="调度原则" description={config.heroBody} />
              {config.notes.map(note => (
                <Caption1 key={note}>{note}</Caption1>
              ))}
            </LabSurfaceCard>
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="行动队列" description="首屏优先展示现在该做什么，而不是背景信息。" />
              <div className={styles.list}>
                {config.primaryItems.map(item => (
                  <div key={item.title} className={styles.item}>
                    <div className={styles.row}>
                      <Body1Strong>{item.title}</Body1Strong>
                      {toneBadge(item.badgeText, item.badgeTone)}
                    </div>
                    <Body1>{item.description}</Body1>
                    <Caption1>{item.meta}</Caption1>
                  </div>
                ))}
              </div>
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="指挥链路" description="把责任链与风险提示集中成一列，形成更明显的 command center 语法。" />
              {config.asideItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
              <Divider />
              {config.secondaryItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function BillingDeskPage({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看批次', '导出名单']} />
        <div className={styles.content}>
          <div className={styles.row}>
            {['账单批次', '异常款项', '催缴节奏', '账务审计'].map(item => (
              <Badge key={item} appearance="outline">{item}</Badge>
            ))}
          </div>
          <LabSurfaceCard>
            <div className={styles.list}>
              {config.primaryItems.concat(config.secondaryItems).map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                  <Caption1>{item.meta}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </div>
          </LabSurfaceCard>
          <div className={styles.split}>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="账期与规则" description="计费运营页强调账期背景、冲正规则和提醒窗口。" />
              {config.asideItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard>
              <LabSectionTitle title="处理建议" description={config.heroSupport} />
              {config.notes.map(note => (
                <div key={note} className={styles.item}>
                  <Body1>{note}</Body1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function SearchWorkbenchPage({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['保存视图', '打开搜索']} />
        <div className={styles.content}>
          <LabSurfaceCard>
            <div className={styles.row}>
              {config.railItems.map(item => (
                <Badge key={item.title} appearance={item.active ? 'filled' : 'outline'} color={item.active ? 'brand' : undefined}>
                  {item.title}
                </Badge>
              ))}
            </div>
            <Body1>{config.heroBody}</Body1>
          </LabSurfaceCard>
          <div className={styles.explorer}>
            <aside className={styles.rail}>
              <LabSectionTitle title="最近搜索与筛选" description="搜索页左侧是过滤与范围，不是普通导航。" />
              {config.railItems.map(item => (
                <LabRailCard key={item.title} active={item.active}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              {[...config.primaryItems, ...config.secondaryItems].map(item => (
                <LabSurfaceCard key={item.title}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Body1>{item.description}</Body1>
                  <div className={styles.row}>
                    <Caption1>{item.meta}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                </LabSurfaceCard>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function DataGridSpecSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['Open patterns', 'Open specimen']} />
        <div className={styles.content}>
          <LabSurfaceCard>
            <LabSectionTitle title="Grid anatomy" description={config.docBody} />
            <div className={styles.specimenRow}>
              <div className={styles.specimenBox} />
              <div className={styles.specimenBox} />
              <div className={styles.specimenBox} />
            </div>
          </LabSurfaceCard>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="列与行模式" description="这张页专门强调 DataGrid 的结构，而不是通用组件展示。" />
              {config.specimens.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  {item.lines.map(line => (
                    <Caption1 key={line}>{line}</Caption1>
                  ))}
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="使用场景" description="把治理列表、资源列表和队列列表明显拉开。" />
              {config.variants.map(item => (
                <div key={item.title} className={styles.item}>
                  <div className={styles.row}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Button appearance={item.primary ? 'primary' : 'secondary'} size="small">{item.buttonLabel}</Button>
                  </div>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function ApprovalConversationSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['转成任务', '最终批复']} tone="teams" />
        <div className={styles.thread}>
          <aside className={styles.threadList}>
            <LabSectionTitle title="审批阶段" description="审批页左侧更像阶段与状态，不是普通线程列表。" />
            {config.channelItems.map(item => (
              <LabRailCard key={item.title} active={item.active}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </LabRailCard>
            ))}
          </aside>
          <main className={styles.content}>
            {config.feedItems.map(item => (
              <div key={item.title} className={styles.post}>
                <div className={styles.row}>
                  <Body1Strong>{item.title}</Body1Strong>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
                <Body1>{item.description}</Body1>
                <Caption1>{item.meta}</Caption1>
              </div>
            ))}
          </main>
          <aside className={styles.rightRail}>
            <LabSectionTitle title="审批上下文" description="右侧专门承接审批责任链与结果同步。" />
            {config.taskItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.description}</Caption1>
                {toneBadge(item.badgeText, item.badgeTone)}
              </div>
            ))}
            <Divider />
            {config.memberItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </aside>
        </div>
      </div>
    </div>
  );
}

function IncidentSwarmSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['导出战情', '更新状态页']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.row}>
            {config.memberItems.map(item => (
              <Badge key={item.title} appearance="outline">
                {item.title}
              </Badge>
            ))}
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="战情流" description="这张页更接近 war room，强调跨团队更新流与外部广播。" />
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <div className={styles.author}>
                    <Avatar name={item.title} color="brand" />
                    <div>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Caption1>{item.meta}</Caption1>
                    </div>
                  </div>
                  <Body1>{item.description}</Body1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="协同动作" description="右侧不再只是普通任务列表，而是战情协作动作。" />
              {config.taskItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
              <Divider />
              {config.notes.map(note => (
                <Caption1 key={note}>{note}</Caption1>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function TenantOverviewSpecial({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['进入租户列表', '查看治理建议']} />
        <div className={styles.content}>
          <div className={styles.wideSplit}>
            <LabSurfaceCard>
              <LabSectionTitle title="租户健康矩阵" description="把健康分、容量和需干预对象并排呈现，形成平台级总览首屏。" />
              <div className={styles.quad}>
                {config.stats.map(item => (
                  <div key={item.label} className={styles.item}>
                    <Caption1>{item.label}</Caption1>
                    <Title3>{item.value}</Title3>
                    {item.tone ? toneBadge(item.tone === 'success' ? '稳定' : item.tone === 'warning' ? '需干预' : '活跃', item.tone) : null}
                  </div>
                ))}
              </div>
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="租户观察清单" description={config.heroSupport} />
              {config.asideItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
          </div>
          <div className={styles.explorer}>
            <aside className={styles.rail}>
              <LabSectionTitle overline="治理范围" title={config.railTitle} description={config.railDescription} />
              {config.railItems.map(item => (
                <LabRailCard key={item.title} active={item.active}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              {[...config.primaryItems, ...config.secondaryItems].map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                  <Caption1>{item.meta}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function IncidentResponseSpecial({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['启动升级', '打开事件简报']} tone="warn" />
        <div className={styles.content}>
          <div className={styles.phaseStrip}>
            {config.railItems.map(item => (
              <div key={item.title} className={styles.item}>
                <div className={styles.row}>
                  <Body1Strong>{item.title}</Body1Strong>
                  {item.active ? <Badge appearance="filled" color="danger">当前阶段</Badge> : null}
                </div>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="处置进展" description="事件页先把升级动作、影响范围和当前决策放进主区。" />
              {config.primaryItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <div className={styles.row}>
                    <Body1Strong>{item.title}</Body1Strong>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                  <Body1>{item.description}</Body1>
                  <Caption1>{item.meta}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="影响面与协同" description={config.heroBody} />
              {config.secondaryItems.concat(
                config.asideItems.map(item => ({
                  title: item.title,
                  description: item.subtitle,
                  meta: item.badgeText ?? '',
                  badgeText: item.badgeText,
                  badgeTone: item.badgeTone,
                })),
              ).map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function PolicyStudioSpecial({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['创建策略', '发布变更']} />
        <div className={styles.triptych}>
          <aside className={styles.rail}>
            <LabSectionTitle overline="Domains" title={config.railTitle} description={config.railDescription} />
            {config.railItems.map(item => (
              <LabRailCard key={item.title} active={item.active}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </LabRailCard>
            ))}
          </aside>
          <main className={styles.content}>
            <div className={styles.spotlight}>
              <Body1Strong>{config.heroTitle}</Body1Strong>
              <Body1>{config.heroBody}</Body1>
              <Caption1>{config.heroSupport}</Caption1>
            </div>
            <div className={styles.board}>
              {config.primaryItems.map(item => (
                <LabSurfaceCard key={item.title}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Body1>{item.description}</Body1>
                  <div className={styles.row}>
                    <Caption1>{item.meta}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                </LabSurfaceCard>
              ))}
            </div>
          </main>
          <aside className={styles.rightRail}>
            <LabSectionTitle overline="Release path" title={config.secondaryTitle} description={config.asideDescription} />
            {config.secondaryItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.description}</Caption1>
                {toneBadge(item.badgeText, item.badgeTone)}
              </div>
            ))}
            <Divider />
            {config.notes.map(note => (
              <Caption1 key={note}>{note}</Caption1>
            ))}
          </aside>
        </div>
      </div>
    </div>
  );
}

function AssetInventorySpecial({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开资产台账', '导出清单']} />
        <div className={styles.content}>
          <div className={styles.explorer}>
            <aside className={styles.rail}>
              <LabSectionTitle overline="Catalog" title={config.railTitle} description={config.railDescription} />
              {config.railItems.map(item => (
                <LabRailCard key={item.title} active={item.active}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              <div className={styles.cardGrid}>
                {config.primaryItems.map(item => (
                  <LabSurfaceCard key={item.title}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Body1>{item.description}</Body1>
                    <div className={styles.row}>
                      <Caption1>{item.meta}</Caption1>
                      {toneBadge(item.badgeText, item.badgeTone)}
                    </div>
                  </LabSurfaceCard>
                ))}
              </div>
              <LabSurfaceCard subtle>
                <LabSectionTitle title="生命周期提醒" description="资产页强调目录、生命周期和责任归属，而不是普通检索结果。" />
                {config.secondaryItems.concat(
                  config.asideItems.map(item => ({
                    title: item.title,
                    description: item.subtitle,
                    meta: item.badgeText ?? '',
                    badgeText: item.badgeText,
                    badgeTone: item.badgeTone,
                  })),
                ).map(item => (
                  <div key={item.title} className={styles.tableRow}>
                    <div>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Caption1>{item.description}</Caption1>
                    </div>
                    <Caption1>{item.meta}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                ))}
              </LabSurfaceCard>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function SupportDeskSpecial({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['接管工单', '发送客户更新']} />
        <div className={styles.triptych}>
          <aside className={styles.rail}>
            <LabSectionTitle overline="Queues" title={config.railTitle} description="工单台左侧承担队列与 SLA 视图切换。" />
            {config.railItems.map(item => (
              <LabRailCard key={item.title} active={item.active}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </LabRailCard>
            ))}
          </aside>
          <main className={styles.content}>
            {config.primaryItems.map(item => (
              <div key={item.title} className={styles.post}>
                <div className={styles.row}>
                  <Body1Strong>{item.title}</Body1Strong>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
                <Body1>{item.description}</Body1>
                <Caption1>{item.meta}</Caption1>
              </div>
            ))}
          </main>
          <aside className={styles.rightRail}>
            <LabSectionTitle overline="Resolution" title={config.secondaryTitle} description={config.heroSupport} />
            {config.secondaryItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.description}</Caption1>
                {toneBadge(item.badgeText, item.badgeTone)}
              </div>
            ))}
            <Divider />
            {config.asideItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </aside>
        </div>
      </div>
    </div>
  );
}

function ReleaseControlSpecial({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['冻结发布', '查看回滚路径']} />
        <div className={styles.content}>
          <div className={styles.phaseStrip}>
            {config.railItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Caption1>{item.title}</Caption1>
                <Body1Strong>{item.subtitle}</Body1Strong>
                {item.active ? <Badge appearance="filled" color="brand">当前</Badge> : null}
              </div>
            ))}
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="发布控制面" description="把发布窗口、阻塞项和回滚门槛压成一个控制台。" />
              {config.primaryItems.concat(config.secondaryItems).map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                  <Caption1>{item.meta}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="门槛与责任人" description={config.heroBody} />
              {config.asideItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              ))}
              <Divider />
              {config.notes.map(note => (
                <Caption1 key={note}>{note}</Caption1>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function KnowledgeHubSpecial({ config }: { config: FoundationScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开知识库', '查看最新摘要']} />
        <div className={styles.content}>
          <div className={styles.spotlight}>
            <Badge appearance="filled" color="brand">Featured</Badge>
            <Title3>{config.heroTitle}</Title3>
            <Body1>{config.heroBody}</Body1>
            <Caption1>{config.heroSupport}</Caption1>
          </div>
          <div className={styles.wideSplit}>
            <div className={styles.list}>
              {config.primaryItems.map(item => (
                <LabSurfaceCard key={item.title}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Body1>{item.description}</Body1>
                  <div className={styles.row}>
                    <Caption1>{item.meta}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                </LabSurfaceCard>
              ))}
            </div>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="知识导读" description="知识页更像运营型内容中枢，而不是工单或队列表。" />
              {config.secondaryItems.concat(
                config.asideItems.map(item => ({
                  title: item.title,
                  description: item.subtitle,
                  meta: item.badgeText ?? '',
                  badgeText: item.badgeText,
                  badgeTone: item.badgeTone,
                })),
              ).map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function ShellGuidelinesSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看壳层解剖', '复制壳层示例']} />
        <div className={styles.content}>
          <div className={styles.spotlight}>
            <Body1Strong>{config.docTitle}</Body1Strong>
            <Body1>{config.docBody}</Body1>
            <div className={styles.phaseStrip}>
              {['顶部栏', '导航区', '主工作区', '次级面板'].map(item => (
                <div key={item} className={styles.item}>
                  <Caption1>{item}</Caption1>
                </div>
              ))}
            </div>
          </div>
          <div className={styles.doc}>
            <aside className={styles.rail}>
              <LabSectionTitle overline="Shell anatomy" title={config.navTitle} />
              {config.navItems.map(item => (
                <LabRailCard key={item.title}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              {config.docNotes.map(note => (
                <div key={note} className={styles.item}>
                  <Body1>{note}</Body1>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function FormPatternsSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开表单分组', '复制字段模式']} />
        <div className={styles.triptych}>
          <aside className={styles.rail}>
            <LabSectionTitle overline="Sections" title={config.navTitle} description="表单规范页更像完成度导航，而不是普通文档目录。" />
            {config.navItems.map(item => (
              <LabRailCard key={item.title}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </LabRailCard>
            ))}
          </aside>
          <main className={styles.content}>
            {config.specimens.map(item => (
              <LabSurfaceCard key={item.title}>
                <LabSectionTitle title={item.title} description={item.description} />
                {item.lines.map(line => (
                  <Caption1 key={line}>{line}</Caption1>
                ))}
              </LabSurfaceCard>
            ))}
          </main>
          <aside className={styles.rightRail}>
            <LabSectionTitle overline="Variants" title={config.variantTitle} description={config.docBody} />
            {config.variants.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
                <Button appearance={item.primary ? 'primary' : 'secondary'} size="small">
                  {item.buttonLabel}
                </Button>
              </div>
            ))}
          </aside>
        </div>
      </div>
    </div>
  );
}

function NavigationSpecSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看导航层级', '复制导航示意']} />
        <div className={styles.content}>
          <div className={styles.board}>
            {['导航深度', '切换语义', '面包屑节奏'].map(item => (
              <div key={item} className={styles.spotlight}>
                <Body1Strong>{item}</Body1Strong>
                <Caption1>{config.description}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.doc}>
            <aside className={styles.rail}>
              <LabSectionTitle title={config.navTitle} />
              {config.navItems.map(item => (
                <LabRailCard key={item.title}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              {config.specimens.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {item.lines.map(line => (
                    <Body1 key={line}>{line}</Body1>
                  ))}
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function TokenGovernanceSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看 token 流程', '导出治理规则']} tone="warn" />
        <div className={styles.content}>
          <div className={styles.wideSplit}>
            <LabSurfaceCard>
              <LabSectionTitle title="Token 生命周期" description="治理页优先展示命名、发布、弃用和主题同步。" />
              {config.navItems.concat(config.variants.map(item => ({ title: item.title, subtitle: item.subtitle }))).map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.subtitle}</Caption1>
                  </div>
                  <Caption1>治理中</Caption1>
                  <Badge appearance="tint" color="warning">需审阅</Badge>
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="评审结论" description={config.docBody} />
              {config.docNotes.map(note => (
                <div key={note} className={styles.item}>
                  <Body1>{note}</Body1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function AccessibilityReviewSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['导出问题单', '查看证据']} tone="warn" />
        <div className={styles.content}>
          <div className={styles.board}>
            {config.variants.map(item => (
              <LabSurfaceCard key={item.title}>
                <div className={styles.row}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Badge appearance={item.primary ? 'filled' : 'tint'} color={item.primary ? 'danger' : 'warning'}>
                    {item.primary ? '高优先级' : '需跟进'}
                  </Badge>
                </div>
                <Caption1>{item.subtitle}</Caption1>
              </LabSurfaceCard>
            ))}
          </div>
          <LabSurfaceCard subtle>
            <LabSectionTitle title={config.docTitle} description={config.docBody} />
            {config.specimens.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                {item.lines.map(line => (
                  <Caption1 key={line}>{line}</Caption1>
                ))}
              </div>
            ))}
          </LabSurfaceCard>
        </div>
      </div>
    </div>
  );
}

function HandoffPatternsSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看交接流', '复制模式']} tone="warn" />
        <div className={styles.content}>
          <div className={styles.phaseStrip}>
            {config.navItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="交接前" description="明确发起方需要准备什么，再决定 handoff 是否触发。" />
              {config.specimens.slice(0, 2).map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="交接后" description="把接收方的确认、责任和回流机制集中在右侧。" />
              {config.specimens.slice(2).map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                </div>
              ))}
              {config.variants.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function TemplateGallerySpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开模板库', '查看上新']} />
        <div className={styles.content}>
          <div className={styles.cardGrid}>
            {config.specimens.map(item => (
              <div key={item.title} className={styles.specimen}>
                <div className={styles.row}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Badge appearance="filled" color="brand">Template</Badge>
                </div>
                <Caption1>{item.description}</Caption1>
                <div className={styles.specimenRow}>
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                </div>
              </div>
            ))}
          </div>
          <LabSurfaceCard subtle>
            <LabSectionTitle title={config.variantTitle} description="模板画廊页应该先像模板库，再像说明文档。" />
            <div className={styles.row}>
              {config.variants.map(item => (
                <Button key={item.title} appearance={item.primary ? 'primary' : 'secondary'}>
                  {item.title}
                </Button>
              ))}
            </div>
          </LabSurfaceCard>
        </div>
      </div>
    </div>
  );
}

function MotionPrinciplesSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看时序', '打开动效样例']} />
        <div className={styles.content}>
          <div className={styles.phaseStrip}>
            {config.variants.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </div>
          <LabSurfaceCard>
            <LabSectionTitle title={config.docTitle} description={config.docBody} />
            <div className={styles.list}>
              {config.specimens.map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                  <Caption1>80-180ms</Caption1>
                  <Badge appearance="tint" color="brand">受控</Badge>
                </div>
              ))}
            </div>
          </LabSurfaceCard>
        </div>
      </div>
    </div>
  );
}

function SidepanelReferenceSpecial({ config }: { config: FluentScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看侧栏宽度', '复制侧栏布局']} />
        <div className={styles.content}>
          <div className={styles.explorer}>
            <aside className={styles.rail}>
              <LabSectionTitle overline="Panel modes" title={config.navTitle} />
              {config.navItems.map(item => (
                <LabRailCard key={item.title}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </LabRailCard>
              ))}
            </aside>
            <div className={styles.list}>
              <div className={styles.spotlight}>
                <Title3>{config.docTitle}</Title3>
                <Body1>{config.docBody}</Body1>
              </div>
              {config.specimens.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {item.lines.map(line => (
                    <Body1 key={line}>{line}</Body1>
                  ))}
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function MeetingCommandSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开议程', '同步跟进']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.agendaStrip}>
            {config.channelItems.map(item => (
              <div key={item.title} className={styles.agendaCard}>
                <div className={styles.row}>
                  <Body1Strong>{item.title}</Body1Strong>
                  {item.active ? <Badge appearance="filled" color="brand">当前</Badge> : null}
                </div>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.stageGrid}>
            <div className={styles.stageFrame}>
              <LabSectionTitle overline="Meeting stage" title={config.heroTitle} description={config.heroBody} />
              <div className={styles.memberStrip}>
                {config.memberItems.map(item => (
                  <div key={item.title} className={styles.memberChip}>
                    <Avatar name={item.title} color="colorful" size={28} />
                    <div>
                      <Body1Strong>{item.title}</Body1Strong>
                      <Caption1>{item.subtitle}</Caption1>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <LabSurfaceCard subtle>
              <LabSectionTitle overline="Live actions" title={config.taskTitle} description={config.heroSupport} />
              {config.taskItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
          </div>
          <LabSurfaceCard>
            <LabSectionTitle overline="Meeting feed" title={config.feedTitle} />
            <div className={styles.bulletinFeed}>
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.bulletinCard}>
                  <div className={styles.row}>
                    <Body1Strong>{item.title}</Body1Strong>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                  <Body1>{item.description}</Body1>
                  <Caption1>{item.meta}</Caption1>
                </div>
              ))}
            </div>
          </LabSurfaceCard>
        </div>
      </div>
    </div>
  );
}

function FrontlineBriefingSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['发送班前简报', '查看现场状态']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.teamsSpotlight}>
            <Body1Strong>{config.heroTitle}</Body1Strong>
            <Body1>{config.heroBody}</Body1>
            <Caption1>{config.heroSupport}</Caption1>
          </div>
          <div className={styles.teamsBoard}>
            {config.stats.map(item => (
              <LabSurfaceCard key={item.label}>
                <Caption1>{item.label}</Caption1>
                <Title3>{item.value}</Title3>
                {item.tone ? toneBadge(item.tone === 'success' ? '稳定' : item.tone === 'warning' ? '需关注' : '进行中', item.tone) : null}
              </LabSurfaceCard>
            ))}
          </div>
          <div className={styles.wideSplit}>
            <LabSurfaceCard>
              <LabSectionTitle title="现场简报" description="前线简报页更像班前会板，不是普通频道时间线。" />
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Body1>{item.description}</Body1>
                  <Caption1>{item.meta}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="班组动作" description={config.heroSupport} />
              <div className={styles.teamsChecklist}>
                {config.taskItems.map(item => (
                  <div key={item.title} className={styles.item}>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                ))}
              </div>
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function FileCollaborationSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开文件区', '请求审阅']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.memberStrip}>
            {config.memberItems.map(item => (
              <div key={item.title} className={styles.memberChip}>
                <Avatar name={item.title} color="colorful" size={28} />
                <div>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              </div>
            ))}
          </div>
          <div className={styles.fileShelf}>
            {config.feedItems.map(item => (
              <div key={item.title} className={styles.fileCard}>
                <div className={styles.row}>
                  <Body1Strong>{item.title}</Body1Strong>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
                <Body1>{item.description}</Body1>
                <Caption1>{item.meta}</Caption1>
                <div className={styles.specimenRow}>
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                  <div className={styles.specimenBox} />
                </div>
              </div>
            ))}
          </div>
          <div className={styles.wideSplit}>
            <LabSurfaceCard>
              <LabSectionTitle overline="Review lanes" title={config.channelTitle} description="文件协作页左侧先给出文件集合和审阅状态。" />
              {config.channelItems.map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.subtitle}</Caption1>
                  </div>
                  <Caption1>审阅流</Caption1>
                  {item.active ? <Badge appearance="filled" color="brand">当前队列</Badge> : <Badge appearance="outline">队列</Badge>}
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle overline="Review actions" title={config.taskTitle} description={config.membersDescription} />
              {config.taskItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function OnboardingHubSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['查看入职路径', '发送欢迎包']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.phaseStrip}>
            {config.channelItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="入职内容" description="把新成员路径、任务包和群组引导收成一页。" />
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Body1>{item.description}</Body1>
                  <Caption1>{item.meta}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="协作者与提醒" description={config.heroSupport} />
              {config.taskItems.concat(
                config.memberItems.map(item => ({
                  title: item.title,
                  description: item.subtitle,
                  meta: item.badgeText ?? '',
                  badgeText: item.badgeText,
                  badgeTone: item.badgeTone,
                })),
              ).map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function CommunityAnnouncementsSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['发布公告', '安排回应']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.teamsSpotlight}>
            <Badge appearance="filled" color="brand">Pinned notice</Badge>
            <Title3>{config.heroTitle}</Title3>
            <Body1>{config.heroBody}</Body1>
          </div>
          <div className={styles.wideSplit}>
            <div className={styles.bulletinFeed}>
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.bulletinCard}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Body1>{item.description}</Body1>
                  <div className={styles.row}>
                    <Caption1>{item.meta}</Caption1>
                    {toneBadge(item.badgeText, item.badgeTone)}
                  </div>
                </div>
              ))}
            </div>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="社区运营" description={config.heroSupport} />
              {config.taskItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function ShiftHandoffSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['完成交接', '查看上一班次']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.teamsSpotlight}>
            <Body1Strong>{config.heroTitle}</Body1Strong>
            <Body1>{config.heroBody}</Body1>
            <Caption1>{config.heroSupport}</Caption1>
          </div>
          <div className={styles.phaseStrip}>
            {config.channelItems.map(item => (
              <div key={item.title} className={styles.item}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.handoffBoard}>
            <LabSurfaceCard>
              <LabSectionTitle overline="Outgoing shift" title={config.feedTitle} description="上一班次的遗留、观察项和背景说明。" />
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  <Caption1>{item.meta}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle overline="Incoming shift" title={config.taskTitle} description="接班后先要执行的动作与确认。" />
              {config.taskItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard>
              <LabSectionTitle overline="Owners" title={config.membersTitle} description={config.membersDescription} />
              {config.memberItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function WebinarOperationsSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开 run of show', '同步直播状态']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.teamsSpotlight}>
            <Body1Strong>{config.heroTitle}</Body1Strong>
            <Body1>{config.heroBody}</Body1>
            <Caption1>{config.heroSupport}</Caption1>
          </div>
          <div className={styles.agendaStrip}>
            {config.channelItems.map(item => (
              <div key={item.title} className={styles.agendaCard}>
                <Body1Strong>{item.title}</Body1Strong>
                <Caption1>{item.subtitle}</Caption1>
              </div>
            ))}
          </div>
          <div className={styles.teamsBoard}>
            {config.stats.map(item => (
              <div key={item.label} className={styles.item}>
                <Caption1>{item.label}</Caption1>
                <Title3>{item.value}</Title3>
                {item.tone ? toneBadge(item.tone === 'success' ? '正常' : item.tone === 'warning' ? '需关注' : '进行中', item.tone) : null}
              </div>
            ))}
          </div>
          <div className={styles.split}>
            <LabSurfaceCard>
              <LabSectionTitle title="直播编排" description="会议运营页更像 run of show 控制台，而不是普通频道。" />
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                  <Caption1>{item.meta}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle title="制作协同" description={config.heroSupport} />
              {config.taskItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

function PartnerStandupSpecial({ config }: { config: TeamsScenarioConfig }) {
  const styles = useStyles();
  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <Header {...config} actions={['打开伙伴议程', '同步外部决议']} tone="teams" />
        <div className={styles.content}>
          <div className={styles.memberStrip}>
            {config.memberItems.map(item => (
              <div key={item.title} className={styles.memberChip}>
                <Avatar name={item.title} color="colorful" size={28} />
                <div>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.subtitle}</Caption1>
                </div>
              </div>
            ))}
          </div>
          <div className={styles.wideSplit}>
            <LabSurfaceCard>
              <LabSectionTitle overline="Standup blockers" title={config.feedTitle} description="伙伴站会首屏先看阻塞和当天共识，不先看资料。" />
              {config.feedItems.map(item => (
                <div key={item.title} className={styles.tableRow}>
                  <div>
                    <Body1Strong>{item.title}</Body1Strong>
                    <Caption1>{item.description}</Caption1>
                  </div>
                  <Caption1>{item.meta}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
            <LabSurfaceCard subtle>
              <LabSectionTitle overline="Joint actions" title={config.taskTitle} description={config.heroSupport} />
              {config.taskItems.map(item => (
                <div key={item.title} className={styles.item}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.description}</Caption1>
                  {toneBadge(item.badgeText, item.badgeTone)}
                </div>
              ))}
            </LabSurfaceCard>
          </div>
        </div>
      </div>
    </div>
  );
}

export const OperationsCommandCenterPage = () => <OperationsCommandHero config={foundationScenarioData['operations-command']} />;
export const TenantOverviewPage = () => <TenantOverviewSpecial config={foundationScenarioData['tenant-overview']} />;
export const BillingOperationsPage = () => <BillingDeskPage config={foundationScenarioData['billing-operations']} />;
export const IncidentResponsePage = () => <IncidentResponseSpecial config={foundationScenarioData['incident-response']} />;
export const PolicyStudioPage = () => <PolicyStudioSpecial config={foundationScenarioData['policy-studio']} />;
export const AssetInventoryPage = () => <AssetInventorySpecial config={foundationScenarioData['asset-inventory']} />;
export const SupportDeskPage = () => <SupportDeskSpecial config={foundationScenarioData['support-desk']} />;
export const SearchWorkspacePage = () => <SearchWorkbenchPage config={foundationScenarioData['search-workspace']} />;
export const ReleaseControlPage = () => <ReleaseControlSpecial config={foundationScenarioData['release-control']} />;
export const KnowledgeHubPage = () => <KnowledgeHubSpecial config={foundationScenarioData['knowledge-hub']} />;

export const ShellGuidelinesPage = () => <ShellGuidelinesSpecial config={fluentScenarioData['shell-guidelines']} />;
export const DataGridSpecPage = () => <DataGridSpecSpecial config={fluentScenarioData['data-grid-spec']} />;
export const FormPatternsPage = () => <FormPatternsSpecial config={fluentScenarioData['form-patterns']} />;
export const NavigationSpecPage = () => <NavigationSpecSpecial config={fluentScenarioData['navigation-spec']} />;
export const TokenGovernancePage = () => <TokenGovernanceSpecial config={fluentScenarioData['token-governance']} />;
export const AccessibilityReviewPage = () => <AccessibilityReviewSpecial config={fluentScenarioData['accessibility-review']} />;
export const HandoffPatternsPage = () => <HandoffPatternsSpecial config={fluentScenarioData['handoff-patterns']} />;
export const TemplateGalleryPage = () => <TemplateGallerySpecial config={fluentScenarioData['template-gallery']} />;
export const MotionPrinciplesPage = () => <MotionPrinciplesSpecial config={fluentScenarioData['motion-principles']} />;
export const SidepanelReferencePage = () => <SidepanelReferenceSpecial config={fluentScenarioData['sidepanel-reference']} />;

export const MeetingCommandPage = () => <MeetingCommandSpecial config={teamsScenarioData['meeting-command']} />;
export const FrontlineBriefingPage = () => <FrontlineBriefingSpecial config={teamsScenarioData['frontline-briefing']} />;
export const FileCollaborationPage = () => <FileCollaborationSpecial config={teamsScenarioData['file-collaboration']} />;
export const ApprovalConversationPage = () => <ApprovalConversationSpecial config={teamsScenarioData['approval-conversation']} />;
export const OnboardingHubPage = () => <OnboardingHubSpecial config={teamsScenarioData['onboarding-hub']} />;
export const CommunityAnnouncementsPage = () => <CommunityAnnouncementsSpecial config={teamsScenarioData['community-announcements']} />;
export const ShiftHandoffPage = () => <ShiftHandoffSpecial config={teamsScenarioData['shift-handoff']} />;
export const WebinarOperationsPage = () => <WebinarOperationsSpecial config={teamsScenarioData['webinar-operations']} />;
export const IncidentSwarmPage = () => <IncidentSwarmSpecial config={teamsScenarioData['incident-swarm']} />;
export const PartnerStandupPage = () => <PartnerStandupSpecial config={teamsScenarioData['partner-standup']} />;
