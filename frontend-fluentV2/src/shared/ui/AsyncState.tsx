import type { ReactNode } from 'react'
import { Body1, MessageBar, MessageBarBody, Spinner, makeStyles, tokens } from '@fluentui/react-components'

const useStyles = makeStyles({
  block: {
    minHeight: '140px',
    display: 'grid',
    placeItems: 'center',
    padding: '20px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  inner: {
    display: 'grid',
    gap: '10px',
    justifyItems: 'center',
    textAlign: 'center',
  },
})

export function LoadingState({ label = '正在加载' }: { label?: string }) {
  const styles = useStyles()
  return (
    <div className={styles.block}>
      <Spinner label={label} />
    </div>
  )
}

export function EmptyState({
  title,
  description,
  actions,
}: {
  title: string
  description?: string
  actions?: ReactNode
}) {
  const styles = useStyles()
  return (
    <div className={styles.block}>
      <div className={styles.inner}>
        <Body1>{title}</Body1>
        {description ? <Body1>{description}</Body1> : null}
        {actions}
      </div>
    </div>
  )
}

export function ErrorState({
  title = '数据加载失败',
  description,
  actions,
}: {
  title?: string
  description?: string
  actions?: ReactNode
}) {
  return (
    <MessageBar intent="error">
      <MessageBarBody>
        {title}
        {description ? `：${description}` : ''}
      </MessageBarBody>
      {actions}
    </MessageBar>
  )
}
