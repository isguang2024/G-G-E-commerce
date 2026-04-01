import type { PropsWithChildren, ReactNode } from 'react'
import { Body1, Body1Strong, Card, makeStyles, tokens } from '@fluentui/react-components'

const useStyles = makeStyles({
  card: {
    display: 'grid',
    gap: '14px',
    padding: '20px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow4,
  },
  header: {
    display: 'grid',
    gap: '6px',
  },
  description: {
    color: tokens.colorNeutralForeground3,
  },
})

export function SectionCard({
  title,
  description,
  actions,
  children,
}: PropsWithChildren<{ title: string; description?: string; actions?: ReactNode }>) {
  const styles = useStyles()

  return (
    <Card className={styles.card}>
      <div className={styles.header}>
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12, alignItems: 'start' }}>
          <Body1Strong>{title}</Body1Strong>
          {actions}
        </div>
        {description ? <Body1 className={styles.description}>{description}</Body1> : null}
      </div>
      {children}
    </Card>
  )
}
