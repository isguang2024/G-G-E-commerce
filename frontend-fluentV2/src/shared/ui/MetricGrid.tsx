import { Badge, Body1, Caption1, makeStyles, tokens } from '@fluentui/react-components'
import type { MetricCard } from '@/shared/types/admin'

const useStyles = makeStyles({
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
    gap: '14px',
  },
  card: {
    display: 'grid',
    gap: '6px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    boxShadow: tokens.shadow4,
  },
  value: {
    fontSize: tokens.fontSizeHero700,
    lineHeight: tokens.lineHeightHero700,
    fontWeight: tokens.fontWeightSemibold,
  },
  hint: {
    color: tokens.colorNeutralForeground3,
  },
})

function resolveToneColor(tone: MetricCard['tone']) {
  switch (tone) {
    case 'brand':
      return 'brand'
    case 'success':
      return 'success'
    case 'warning':
      return 'warning'
    case 'danger':
      return 'danger'
    default:
      return 'informative'
  }
}

export function MetricGrid({ metrics }: { metrics: MetricCard[] }) {
  const styles = useStyles()
  return (
    <div className={styles.grid}>
      {metrics.map((item) => (
        <div key={item.id} className={styles.card}>
          <Badge appearance="tint" color={resolveToneColor(item.tone)}>
            {item.label}
          </Badge>
          <div className={styles.value}>{item.value}</div>
          {item.hint ? <Caption1 className={styles.hint}>{item.hint}</Caption1> : <Body1 />}
        </div>
      ))}
    </div>
  )
}
