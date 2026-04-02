import type { PropsWithChildren, ReactNode } from 'react'
import { makeStyles } from '@fluentui/react-components'

const useStyles = makeStyles({
  twoPane: {
    display: 'grid',
    gridTemplateColumns: 'minmax(320px, 0.92fr) minmax(420px, 1.08fr)',
    gap: '16px',
    alignItems: 'start',
    '@media (max-width: 1160px)': {
      gridTemplateColumns: '1fr',
    },
  },
  threePane: {
    display: 'grid',
    gridTemplateColumns: 'minmax(240px, 0.78fr) minmax(320px, 1fr) minmax(280px, 0.86fr)',
    gap: '16px',
    alignItems: 'start',
    '@media (max-width: 1260px)': {
      gridTemplateColumns: '1fr',
    },
  },
  stack: {
    display: 'grid',
    gap: '14px',
    minWidth: 0,
  },
})

export function TwoPaneWorkbench({
  primary,
  secondary,
}: {
  primary: ReactNode
  secondary: ReactNode
}) {
  const styles = useStyles()
  return <div className={styles.twoPane}>{primary}{secondary}</div>
}

export function ThreePaneWorkbench({
  left,
  center,
  right,
}: {
  left: ReactNode
  center: ReactNode
  right: ReactNode
}) {
  const styles = useStyles()
  return <div className={styles.threePane}>{left}{center}{right}</div>
}

export function WorkbenchStack({ children }: PropsWithChildren) {
  const styles = useStyles()
  return <div className={styles.stack}>{children}</div>
}
