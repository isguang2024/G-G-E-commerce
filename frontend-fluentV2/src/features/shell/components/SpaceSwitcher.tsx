import { Badge, Body1, Button, Caption1, Menu, MenuItem, MenuList, MenuPopover, MenuTrigger, makeStyles, tokens } from '@fluentui/react-components'
import type { NavigationSpace } from '@/shared/types/navigation'

const useStyles = makeStyles({
  button: {
    minWidth: '220px',
    justifyContent: 'space-between',
  },
  popover: {
    minWidth: '280px',
  },
  item: {
    display: 'grid',
    gap: '2px',
  },
  description: {
    color: tokens.colorNeutralForeground3,
  },
})

export function SpaceSwitcher({
  currentSpace,
  spaces,
  onSelect,
}: {
  currentSpace: NavigationSpace
  spaces: NavigationSpace[]
  onSelect: (key: NavigationSpace['key']) => void
}) {
  const styles = useStyles()

  return (
    <Menu>
      <MenuTrigger disableButtonEnhancement>
        <Button className={styles.button} appearance="secondary" icon={<Badge appearance="tint">空间</Badge>}>
          {currentSpace.label}
        </Button>
      </MenuTrigger>
      <MenuPopover className={styles.popover}>
        <MenuList>
          {spaces.map((space) => (
            <MenuItem key={space.key} onClick={() => onSelect(space.key)}>
              <div className={styles.item}>
                <Body1>{space.label}</Body1>
                <Caption1 className={styles.description}>{space.description}</Caption1>
              </div>
            </MenuItem>
          ))}
        </MenuList>
      </MenuPopover>
    </Menu>
  )
}
