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
          <Subtitle2>应用壳异常</Subtitle2>
          <Title2>当前壳层暂时无法渲染</Title2>
          <Body1>
            当前已进入全局错误边界。请刷新页面重试；如果问题持续存在，再检查最近的壳层、路由或 mock 变更。
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
