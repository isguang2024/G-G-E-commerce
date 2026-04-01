import { Body1Strong, makeStyles, tokens } from '@fluentui/react-components'

const useStyles = makeStyles({
  root: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    minWidth: 0,
  },
  mark: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '32px',
    width: '32px',
    flexShrink: 0,
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorBrandBackground,
    color: tokens.colorNeutralForegroundOnBrand,
    boxShadow: tokens.shadow4,
  },
  text: {
    color: tokens.colorNeutralForeground1,
    whiteSpace: 'nowrap',
    minWidth: 0,
  },
})

export function AppLogo() {
  const styles = useStyles()

  return (
    <div className={styles.root}>
      <Body1Strong as="span" className={styles.mark}>
        G
      </Body1Strong>
      <Body1Strong className={styles.text}>G&amp;G ERP</Body1Strong>
    </div>
  )
}
