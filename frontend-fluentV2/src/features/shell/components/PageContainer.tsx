import type { PropsWithChildren, ReactNode } from 'react'
import { Badge, Body1, Spinner, makeStyles, tokens } from '@fluentui/react-components'
import { useRouteContext } from '@/features/navigation/navigation.service'
import { BreadcrumbsBar } from '@/features/shell/components/BreadcrumbsBar'

const useStyles = makeStyles({
  root: {
    display: 'grid',
    gap: '18px',
  },
  header: {
    display: 'grid',
    gap: '12px',
    padding: '24px 28px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow8,
  },
  titleRow: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '16px',
    flexWrap: 'wrap',
  },
  titleBlock: {
    display: 'grid',
    gap: '6px',
    minWidth: 0,
  },
  title: {
    margin: 0,
    color: tokens.colorNeutralForeground1,
    fontSize: tokens.fontSizeHero700,
    lineHeight: tokens.lineHeightHero700,
    fontWeight: tokens.fontWeightSemibold,
    letterSpacing: '-0.02em',
  },
  subtitle: {
    color: tokens.colorNeutralForeground3,
    maxWidth: '780px',
  },
  actions: {
    display: 'flex',
    gap: '10px',
    flexWrap: 'wrap',
  },
  metaRow: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  content: {
    display: 'grid',
    gap: '18px',
  },
  loading: {
    minHeight: '160px',
    display: 'grid',
    placeItems: 'center',
  },
})

export function PageContainer({
  routeId,
  actions,
  children,
}: PropsWithChildren<{ routeId?: string; actions?: ReactNode }>) {
  const styles = useStyles()
  const { context, isLoading } = useRouteContext(routeId)

  if (isLoading) {
    return (
      <div className={styles.loading}>
        <Spinner label="正在读取页面信息" />
      </div>
    )
  }

  if (!context) {
    return null
  }

  return (
    <div className={styles.root}>
      <header className={styles.header}>
        <BreadcrumbsBar routeId={routeId} />
        <div className={styles.titleRow}>
          <div className={styles.titleBlock}>
            <h1 className={styles.title}>{context.title}</h1>
            <Body1 className={styles.subtitle}>{context.subtitle}</Body1>
          </div>
          {actions ? <div className={styles.actions}>{actions}</div> : null}
        </div>
        <div className={styles.metaRow}>
          <Badge appearance="filled" color="brand">
            {context.groupLabel}
          </Badge>
          {context.spaceKey ? <Badge appearance="tint">空间：{context.spaceKey}</Badge> : null}
          <Badge appearance="outline">{context.status === 'implemented' ? '已实现' : '占位承接'}</Badge>
        </div>
      </header>
      <div className={styles.content}>{children}</div>
    </div>
  )
}
