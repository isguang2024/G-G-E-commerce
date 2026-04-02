import { Body1, Caption1, makeStyles, tokens } from '@fluentui/react-components'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'

const useStyles = makeStyles({
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
    gap: '14px',
  },
  card: {
    display: 'grid',
    gap: '10px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    boxShadow: tokens.shadow4,
  },
  description: {
    color: tokens.colorNeutralForeground3,
  },
})

export interface LinkCardItem {
  id: string
  title: string
  description: string
  to: string
  actionLabel: string
}

export function LinkCardGrid({ items }: { items: LinkCardItem[] }) {
  const styles = useStyles()

  return (
    <div className={styles.grid}>
      {items.map((item) => (
        <div key={item.id} className={styles.card}>
          <Body1>{item.title}</Body1>
          <Caption1 className={styles.description}>{item.description}</Caption1>
          <RouterButtonLink appearance="secondary" to={item.to}>
            {item.actionLabel}
          </RouterButtonLink>
        </div>
      ))}
    </div>
  )
}
