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
import { LabBadgeRow, LabRailCard, LabSectionTitle, LabStatGrid, LabSurfaceCard } from '../lab/primitives';

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '18px',
  },
  shell: {
    minHeight: '760px',
    display: 'grid',
    gridTemplateColumns: '240px minmax(0, 1fr)',
    borderRadius: tokens.borderRadiusXLarge,
    overflow: 'hidden',
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow16,
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
  rail: {
    display: 'grid',
    alignContent: 'start',
    gap: '12px',
    padding: '18px 16px',
    backgroundColor: tokens.colorNeutralBackground3,
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  content: {
    display: 'grid',
    gap: '18px',
    padding: '24px',
  },
  hero: {
    display: 'grid',
    gap: '10px',
    maxWidth: '840px',
  },
  previewCard: {
    display: 'grid',
    gap: '14px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  specimen: {
    display: 'grid',
    gap: '12px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  row: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
  line: {
    height: '10px',
    borderRadius: tokens.borderRadiusCircular,
    backgroundColor: tokens.colorNeutralBackground4,
  },
  specGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '16px',
    '@media (max-width: 900px)': {
      gridTemplateColumns: '1fr',
    },
  },
});

function RailItem({
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

export function FluentSpecWorkspacePage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <Title3>Fluent 规范工作区测试页</Title3>
      <div className={styles.shell}>
        <aside className={styles.rail}>
          <RailItem title="Overview" subtitle="页面概览与目标" />
          <RailItem title="Anatomy" subtitle="结构拆解" active />
          <RailItem title="Variants" subtitle="状态和外观" />
          <RailItem title="Usage" subtitle="使用建议" />
        </aside>

        <main className={styles.content}>
          <header className={styles.hero}>
            <LabBadgeRow>
              <Badge appearance="filled" color="brand">
                Microsoft Fluent 2 Web
              </Badge>
              <Badge appearance="tint">文档式工作区</Badge>
            </LabBadgeRow>
            <Subtitle2>这张页更接近组件文档里的工作区结构，而不是通用管理后台。</Subtitle2>
            <Body1>
              目标是验证“左侧章节导航 + 主内容工作区 + 预览与规则块”这种 Fluent 2 Web 常见的文档式布局，看看它在实验场里是否足够稳定。
            </Body1>
            <LabStatGrid
              items={[
                { label: '章节层级', value: '4', tone: 'brand' },
                { label: '预览区块', value: '3', tone: 'success' },
                { label: '规则提示', value: '6', tone: 'warning' },
              ]}
            />
          </header>

          <LabSurfaceCard subtle>
            <section className={styles.previewCard}>
            <div className={styles.row}>
              <Button appearance="primary">Primary action</Button>
              <Button appearance="secondary">Secondary action</Button>
              <Button appearance="subtle">Subtle action</Button>
            </div>
            <div className={styles.line} style={{ width: '78%' }} />
            <div className={styles.line} style={{ width: '56%' }} />
            <div className={styles.line} style={{ width: '64%' }} />
            </section>
          </LabSurfaceCard>

          <Divider />

          <section className={styles.specGrid}>
            <LabSurfaceCard>
              <article className={styles.specimen}>
              <Body1Strong>Anatomy</Body1Strong>
              <Caption1>适合描述组件的固定结构、最小可用骨架和重点可变区域。</Caption1>
              <div className={styles.line} style={{ width: '70%' }} />
              <div className={styles.line} style={{ width: '52%' }} />
              </article>
            </LabSurfaceCard>

            <LabSurfaceCard>
              <article className={styles.specimen}>
              <Body1Strong>Variants</Body1Strong>
              <Caption1>适合承载 Filled、Tint、Outline、Subtle 等外观和状态示例。</Caption1>
              <div className={styles.row}>
                <Badge appearance="filled">Filled</Badge>
                <Badge appearance="tint" color="success">
                  Tint
                </Badge>
                <Badge appearance="outline">Outline</Badge>
              </div>
              </article>
            </LabSurfaceCard>

            <LabSurfaceCard subtle>
              <article className={styles.specimen}>
              <Body1Strong>Best practices</Body1Strong>
              <Caption1>把正确做法和错误做法拆开，避免长文档压在一个段落里。</Caption1>
              <div className={styles.row}>
                <Badge appearance="tint" color="success">
                  Do
                </Badge>
                <Badge appearance="tint" color="danger">
                  Don&apos;t
                </Badge>
              </div>
              </article>
            </LabSurfaceCard>

            <LabSurfaceCard>
              <article className={styles.specimen}>
              <Body1Strong>Implementation note</Body1Strong>
              <Caption1>适合附上组件组合建议和实现提示，而不是直接堆过多技术字段。</Caption1>
              <Button appearance="subtle">查看实现建议</Button>
              </article>
            </LabSurfaceCard>
          </section>
        </main>
      </div>
    </div>
  );
}
