import * as React from 'react';
import { Badge, Body1Strong, Caption1, makeStyles, mergeClasses, tokens } from '@fluentui/react-components';

const useStyles = makeStyles({
  sectionTitle: {
    display: 'grid',
    gap: '4px',
  },
  overline: {
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
    textTransform: 'uppercase',
    letterSpacing: '0.04em',
  },
  railCard: {
    display: 'grid',
    gap: '8px',
    padding: '14px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  surfaceCard: {
    display: 'grid',
    gap: '12px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  subtleSurface: {
    backgroundColor: tokens.colorNeutralBackground2,
  },
  activeSurface: {
    backgroundColor: tokens.colorBrandBackground2,
    border: `1px solid ${tokens.colorBrandStroke1}`,
  },
  statGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 720px)': {
      gridTemplateColumns: '1fr',
    },
  },
  statCard: {
    display: 'grid',
    gap: '6px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  statValue: {
    fontSize: tokens.fontSizeHero700,
    lineHeight: tokens.lineHeightHero700,
    fontWeight: tokens.fontWeightSemibold,
  },
  badgeRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
  },
});

export function LabSectionTitle({
  overline,
  title,
  description,
}: {
  overline?: string;
  title: string;
  description?: string;
}) {
  const styles = useStyles();

  return (
    <div className={styles.sectionTitle}>
      {overline ? <span className={styles.overline}>{overline}</span> : null}
      <Body1Strong>{title}</Body1Strong>
      {description ? <Caption1>{description}</Caption1> : null}
    </div>
  );
}

export function LabRailCard({
  children,
  active = false,
}: {
  children: React.ReactNode;
  active?: boolean;
}) {
  const styles = useStyles();
  return <div className={mergeClasses(styles.railCard, active && styles.activeSurface)}>{children}</div>;
}

export function LabSurfaceCard({
  children,
  subtle = false,
  active = false,
}: {
  children: React.ReactNode;
  subtle?: boolean;
  active?: boolean;
}) {
  const styles = useStyles();
  return (
    <div className={mergeClasses(styles.surfaceCard, subtle && styles.subtleSurface, active && styles.activeSurface)}>
      {children}
    </div>
  );
}

export function LabStatGrid({
  items,
}: {
  items: Array<{ label: string; value: string; tone?: 'brand' | 'success' | 'warning' }>;
}) {
  const styles = useStyles();

  return (
    <div className={styles.statGrid}>
      {items.map(item => (
        <div key={item.label} className={styles.statCard}>
          <Caption1>{item.label}</Caption1>
          <span className={styles.statValue}>{item.value}</span>
          {item.tone ? (
            <Badge appearance="tint" color={item.tone}>
              {item.tone === 'brand' ? '进行中' : item.tone === 'success' ? '稳定' : '需关注'}
            </Badge>
          ) : null}
        </div>
      ))}
    </div>
  );
}

export function LabBadgeRow({ children }: { children: React.ReactNode }) {
  const styles = useStyles();
  return <div className={styles.badgeRow}>{children}</div>;
}
