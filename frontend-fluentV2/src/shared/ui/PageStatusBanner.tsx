import type { ReactNode } from 'react'
import { MessageBar, MessageBarBody, MessageBarTitle, makeStyles, tokens } from '@fluentui/react-components'

const useStyles = makeStyles({
  banner: {
    borderRadius: tokens.borderRadiusLarge,
  },
  body: {
    whiteSpace: 'pre-wrap',
  },
})

export function PageStatusBanner({
  intent,
  title,
  description,
}: {
  intent: 'success' | 'error' | 'info' | 'warning'
  title: ReactNode
  description?: ReactNode
}) {
  const styles = useStyles()

  return (
    <MessageBar intent={intent} className={styles.banner}>
      <MessageBarBody className={styles.body}>
        <MessageBarTitle>{title}</MessageBarTitle>
        {description}
      </MessageBarBody>
    </MessageBar>
  )
}
