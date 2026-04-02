import { Text, makeStyles, tokens } from '@fluentui/react-components'
import { Link as RouterLink } from 'react-router-dom'
import { useRouteContext } from '@/features/navigation/navigation.service'

const useStyles = makeStyles({
  root: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
    alignItems: 'center',
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
  },
  current: {
    color: tokens.colorNeutralForeground2,
  },
  link: {
    color: tokens.colorBrandForegroundLink,
    textDecorationLine: 'none',
    ':hover': {
      color: tokens.colorBrandForegroundLinkHover,
      textDecorationLine: 'underline',
    },
  },
})

export function BreadcrumbsBar({ routeId }: { routeId?: string }) {
  const styles = useStyles()
  const { context } = useRouteContext(routeId)

  if (!context) {
    return null
  }

  return (
    <div className={styles.root}>
      {context.breadcrumbs.map((item, index) => {
        const isLast = index === context.breadcrumbs.length - 1

        return (
          <span key={`${context.routeId}-${item.label}-${index}`} style={{ display: 'contents' }}>
            {item.path && !isLast ? (
              <RouterLink className={styles.link} to={item.path}>
                {item.label}
              </RouterLink>
            ) : (
              <Text className={isLast ? styles.current : undefined}>{item.label}</Text>
            )}
            {!isLast ? <Text>/</Text> : null}
          </span>
        )
      })}
    </div>
  )
}
