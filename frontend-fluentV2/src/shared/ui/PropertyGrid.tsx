import { Body1, Caption1, makeStyles, tokens } from '@fluentui/react-components'

const useStyles = makeStyles({
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
    gap: '12px',
  },
  item: {
    display: 'grid',
    gap: '4px',
    padding: '12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  label: {
    color: tokens.colorNeutralForeground3,
  },
  value: {
    wordBreak: 'break-word',
  },
})

export function PropertyGrid({ items }: { items: Array<{ label: string; value: string }> }) {
  const styles = useStyles()
  return (
    <div className={styles.grid}>
      {items.map((item) => (
        <div key={item.label} className={styles.item}>
          <Caption1 className={styles.label}>{item.label}</Caption1>
          <Body1 className={styles.value}>{item.value || '-'}</Body1>
        </div>
      ))}
    </div>
  )
}
