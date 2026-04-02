import React from 'react'
import { Body1, Button, FluentProvider, Subtitle2, Title2, makeStyles, tokens } from '@fluentui/react-components'
import { appLightTheme } from '@/shared/config/theme'

type ErrorBoundaryState = {
  hasError: boolean
}

const useStyles = makeStyles({
  root: {
    minHeight: '100vh',
    display: 'grid',
    placeItems: 'center',
    padding: '24px',
    backgroundColor: tokens.colorNeutralBackground2,
  },
  card: {
    width: 'min(560px, 100%)',
    display: 'grid',
    gap: '16px',
    padding: '32px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    boxShadow: tokens.shadow16,
  },
  actions: {
    display: 'flex',
    gap: '12px',
    flexWrap: 'wrap',
  },
})

function ErrorFallback() {
  const styles = useStyles()

  return (
    <FluentProvider theme={appLightTheme}>
      <div className={styles.root}>
        <div className={styles.card}>
          <Subtitle2>系统提示</Subtitle2>
          <Title2>当前页面暂时无法显示</Title2>
          <Body1>
            页面当前出现异常，请尝试刷新界面后重试；如问题仍然存在，请联系管理员处理。
          </Body1>
          <div className={styles.actions}>
            <Button appearance="primary" onClick={() => window.location.reload()}>
              刷新页面
            </Button>
          </div>
        </div>
      </div>
    </FluentProvider>
  )
}

export class AppErrorBoundary extends React.Component<React.PropsWithChildren, ErrorBoundaryState> {
  public state: ErrorBoundaryState = {
    hasError: false,
  }

  public static getDerivedStateFromError(): ErrorBoundaryState {
    return {
      hasError: true,
    }
  }

  public componentDidCatch(error: Error) {
    console.error('[AppErrorBoundary]', error)
  }

  public render() {
    if (this.state.hasError) {
      return <ErrorFallback />
    }

    return this.props.children
  }
}
