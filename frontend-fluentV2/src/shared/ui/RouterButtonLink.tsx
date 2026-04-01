import type { ButtonProps } from '@fluentui/react-components'
import { Button, makeStyles } from '@fluentui/react-components'
import { Link as RouterLink } from 'react-router-dom'
import type { LinkProps as RouterLinkProps } from 'react-router-dom'

const useStyles = makeStyles({
  link: {
    display: 'inline-flex',
    textDecorationLine: 'none',
  },
})

type RouterButtonLinkProps = Pick<
  ButtonProps,
  'appearance' | 'children' | 'disabled' | 'icon' | 'iconPosition' | 'shape' | 'size'
> &
  Pick<RouterLinkProps, 'replace' | 'state' | 'to'>

export function RouterButtonLink({
  children,
  replace,
  state,
  to,
  ...buttonProps
}: RouterButtonLinkProps) {
  const styles = useStyles()

  return (
    <RouterLink className={styles.link} replace={replace} state={state} to={to}>
      <Button {...buttonProps}>{children}</Button>
    </RouterLink>
  )
}
