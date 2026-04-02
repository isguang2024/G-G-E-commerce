import type { PropsWithChildren, ReactNode } from 'react'
import { Body1, Body1Strong, makeStyles, tokens } from '@fluentui/react-components'

const useStyles = makeStyles({
  layout: {
    display: 'grid',
    gap: '16px',
  },
  header: {
    display: 'flex',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: '12px',
    flexWrap: 'wrap',
  },
  content: {
    display: 'grid',
    gap: '14px',
  },
  muted: {
    color: tokens.colorNeutralForeground3,
  },
  actions: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
    justifyContent: 'flex-end',
  },
  meta: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
})

export function EntityPageLayout({
  title,
  description,
  meta,
  actions,
  children,
}: PropsWithChildren<{
  title: ReactNode
  description?: ReactNode
  meta?: ReactNode
  actions?: ReactNode
}>) {
  const styles = useStyles()

  return (
    <div className={styles.layout}>
      <div className={styles.header}>
        <div className={styles.content}>
          <Body1Strong>{title}</Body1Strong>
          {description ? <Body1 className={styles.muted}>{description}</Body1> : null}
          {meta ? <div className={styles.meta}>{meta}</div> : null}
        </div>
        {actions ? <div className={styles.actions}>{actions}</div> : null}
      </div>
      {children}
    </div>
  )
}
