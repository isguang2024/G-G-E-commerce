import {
  Button,
  Dialog,
  DialogBody,
  DialogContent,
  DialogSurface,
  DialogTitle,
  Menu,
  MenuItem,
  MenuList,
  MenuPopover,
  MenuTrigger,
  Tooltip,
  makeStyles,
  tokens,
} from '@fluentui/react-components'
import { ChevronDown16Regular, ChevronRight16Regular } from '@fluentui/react-icons'
import { useEffect, useState } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'
import { AppIcon } from '@/shared/ui/AppIcon'
import type { NavigationItem } from '@/shared/types/navigation'

const useStyles = makeStyles({
  rail: {
    display: 'grid',
    gap: '8px',
    alignContent: 'start',
    height: '100%',
  },
  list: {
    display: 'grid',
    gap: '6px',
  },
  section: {
    display: 'grid',
    gap: '6px',
  },
  itemLink: {
    display: 'grid',
    gap: '6px',
    minWidth: 0,
  },
  collapsedSection: {
    display: 'flex',
    justifyContent: 'center',
  },
  collapsedLink: {
    display: 'flex',
    justifyContent: 'center',
    width: '100%',
  },
  collapsedTrigger: {
    width: '40px',
    minWidth: '40px',
    height: '40px',
    paddingLeft: '0',
    paddingRight: '0',
    justifyContent: 'center',
    gap: '0',
    borderRadius: tokens.borderRadiusLarge,
    color: tokens.colorNeutralForeground2,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorTransparentStroke}`,
    transitionProperty: 'background-color, border-color, color',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1Hover,
      border: `1px solid ${tokens.colorNeutralStroke1Hover}`,
      color: tokens.colorNeutralForeground1,
    },
  },
  collapsedTriggerActive: {
    color: tokens.colorBrandForeground1,
    backgroundColor: tokens.colorBrandBackground2,
    border: `1px solid ${tokens.colorBrandStroke1}`,
    ':hover': {
      backgroundColor: tokens.colorBrandBackground2Hover,
      border: `1px solid ${tokens.colorBrandStroke1}`,
      color: tokens.colorBrandForeground1,
    },
  },
  itemRow: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 1fr) auto',
    alignItems: 'center',
    gap: '6px',
  },
  itemButton: {
    width: '100%',
    justifyContent: 'flex-start',
    gap: '10px',
    minHeight: '40px',
    paddingLeft: '12px',
    paddingRight: '12px',
    borderRadius: tokens.borderRadiusLarge,
    color: tokens.colorNeutralForeground2,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorTransparentStroke}`,
    transitionProperty: 'background-color, border-color, color',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1Hover,
      border: `1px solid ${tokens.colorNeutralStroke1Hover}`,
      color: tokens.colorNeutralForeground1,
    },
  },
  itemButtonActive: {
    color: tokens.colorBrandForeground1,
    backgroundColor: tokens.colorBrandBackground2,
    border: `1px solid ${tokens.colorBrandStroke1}`,
    ':hover': {
      backgroundColor: tokens.colorBrandBackground2Hover,
      border: `1px solid ${tokens.colorBrandStroke1}`,
      color: tokens.colorBrandForeground1,
    },
  },
  itemCollapsed: {
    justifyContent: 'center',
    paddingLeft: '8px',
    paddingRight: '8px',
  },
  itemToggle: {
    minWidth: '32px',
    width: '32px',
    height: '32px',
    paddingLeft: '0',
    paddingRight: '0',
    justifyContent: 'center',
    transitionProperty: 'opacity, transform',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
  },
  itemToggleHidden: {
    opacity: 0,
    transform: 'translateX(-6px)',
    pointerEvents: 'none',
  },
  childList: {
    display: 'grid',
    gap: '4px',
    marginLeft: '16px',
    paddingLeft: '12px',
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  childRegion: {
    display: 'grid',
    gridTemplateRows: '1fr',
    opacity: 1,
    transform: 'translateY(0)',
    transitionProperty: 'grid-template-rows, opacity, transform',
    transitionDuration: tokens.durationNormal,
    transitionTimingFunction: tokens.curveEasyEase,
  },
  childRegionCollapsed: {
    gridTemplateRows: '0fr',
    opacity: 0,
    transform: 'translateY(-6px)',
    pointerEvents: 'none',
  },
  childRegionHidden: {
    opacity: 0,
    transform: 'translateY(-6px)',
    pointerEvents: 'none',
  },
  childRegionInner: {
    overflow: 'hidden',
  },
  childLink: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    minHeight: '34px',
    paddingLeft: '10px',
    paddingRight: '10px',
    borderRadius: tokens.borderRadiusMedium,
    color: tokens.colorNeutralForeground2,
    transitionProperty: 'background-color, color',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1Hover,
      color: tokens.colorNeutralForeground1,
    },
  },
  childLinkActive: {
    backgroundColor: tokens.colorBrandBackground2,
    color: tokens.colorBrandForeground1,
    ':hover': {
      backgroundColor: tokens.colorBrandBackground2Hover,
      color: tokens.colorBrandForeground1,
    },
  },
  label: {
    display: 'inline-block',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    transitionProperty: 'opacity, transform',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
  },
  labelHidden: {
    opacity: 0,
    transform: 'translateX(-6px)',
  },
  dialogSurface: {
    width: 'min(340px, calc(100vw - 24px))',
  },
  flyoutPopover: {
    minWidth: '220px',
  },
  mobileContent: {
    paddingTop: '8px',
  },
})

function isCurrentPath(pathname: string, targetPath: string) {
  return pathname === targetPath || pathname.startsWith(`${targetPath}/`)
}

function collectExpandedAncestorIds(items: NavigationItem[], pathname: string, ancestors: string[] = []) {
  return items.reduce<string[]>((result, item) => {
    if (!item.children?.length) {
      return result
    }

    const nextAncestors = [...ancestors, item.id]
    const childMatched = item.children.some((child) => {
      if (isCurrentPath(pathname, child.path)) {
        return true
      }

      return Boolean(collectExpandedAncestorIds([child], pathname, nextAncestors).length)
    })

    if (childMatched) {
      result.push(...nextAncestors)
    }

    return result
  }, [])
}

function NavEntry({
  item,
  collapsed,
  pathname,
  expanded,
  contentVisible,
  onToggleExpanded,
}: {
  item: NavigationItem
  collapsed: boolean
  pathname: string
  expanded: boolean
  contentVisible: boolean
  onToggleExpanded: () => void
}) {
  const styles = useStyles()
  const navigate = useNavigate()
  const active = isCurrentPath(pathname, item.path)
  const children = item.children ?? []
  const hasChildren = children.length > 0
  const collapsedTriggerClassName = active
    ? `${styles.collapsedTrigger} ${styles.collapsedTriggerActive}`
    : styles.collapsedTrigger
  const buttonClassName = active
    ? `${styles.itemButton} ${styles.itemButtonActive}${collapsed ? ` ${styles.itemCollapsed}` : ''}`
    : `${styles.itemButton}${collapsed ? ` ${styles.itemCollapsed}` : ''}`
  const handleParentClick = hasChildren ? onToggleExpanded : undefined

  const button = (
    <Button className={buttonClassName} appearance="subtle" icon={<AppIcon icon={item.icon} />}>
      {!collapsed ? (
        <span className={contentVisible ? styles.label : `${styles.label} ${styles.labelHidden}`}>
          {item.label}
        </span>
      ) : null}
    </Button>
  )

  if (collapsed && hasChildren) {
    return (
      <div className={styles.collapsedSection}>
        <Menu positioning={{ position: 'after', align: 'top', offset: 14 }}>
          <MenuTrigger disableButtonEnhancement>
            <Tooltip relationship="label" content={item.label}>
              <Button
                aria-label={item.label}
                className={collapsedTriggerClassName}
                appearance="subtle"
                icon={<AppIcon icon={item.icon} />}
              />
            </Tooltip>
          </MenuTrigger>
          <MenuPopover className={styles.flyoutPopover}>
            <MenuList>
              {children.map((child) => (
                <MenuItem
                  key={child.id}
                  icon={<AppIcon icon={child.icon} />}
                  onClick={() => navigate(child.path)}
                >
                  {child.label}
                </MenuItem>
              ))}
            </MenuList>
          </MenuPopover>
        </Menu>
      </div>
    )
  }

  if (collapsed) {
    return (
      <div className={styles.collapsedSection}>
        <Tooltip relationship="label" content={item.label}>
          <NavLink className={styles.collapsedLink} to={item.path}>
            <Button
              aria-label={item.label}
              className={collapsedTriggerClassName}
              appearance="subtle"
              icon={<AppIcon icon={item.icon} />}
            />
          </NavLink>
        </Tooltip>
      </div>
    )
  }

  return (
    <div className={styles.section}>
      <div className={styles.itemRow}>
        {hasChildren ? (
          <Button className={buttonClassName} appearance="subtle" icon={<AppIcon icon={item.icon} />} onClick={handleParentClick}>
            {!collapsed ? (
              <span className={contentVisible ? styles.label : `${styles.label} ${styles.labelHidden}`}>
                {item.label}
              </span>
            ) : null}
          </Button>
        ) : (
          <NavLink className={styles.itemLink} to={item.path}>
            {button}
          </NavLink>
        )}

        {!collapsed && hasChildren ? (
          <Button
            className={contentVisible ? styles.itemToggle : `${styles.itemToggle} ${styles.itemToggleHidden}`}
            appearance="subtle"
            aria-label={expanded ? `收起 ${item.label}` : `展开 ${item.label}`}
            icon={expanded ? <ChevronDown16Regular /> : <ChevronRight16Regular />}
            onClick={onToggleExpanded}
          />
        ) : null}
      </div>

      {!collapsed && hasChildren ? (
        <div
          className={
            expanded
              ? contentVisible
                ? styles.childRegion
                : `${styles.childRegion} ${styles.childRegionHidden}`
              : `${styles.childRegion} ${styles.childRegionCollapsed}`
          }
        >
          <div className={styles.childRegionInner}>
            <div className={styles.childList}>
              {children.map((child) => {
                const childActive = pathname === child.path

                return (
                  <NavLink
                    key={child.id}
                    className={childActive ? `${styles.childLink} ${styles.childLinkActive}` : styles.childLink}
                    to={child.path}
                  >
                    <AppIcon icon={child.icon} />
                    <span className={contentVisible ? styles.label : `${styles.label} ${styles.labelHidden}`}>
                      {child.label}
                    </span>
                  </NavLink>
                )
              })}
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}

export function SideNav({
  items,
  collapsed,
  contentVisible,
  mobileOpen,
  onCloseMobile,
}: {
  items: NavigationItem[]
  collapsed: boolean
  contentVisible: boolean
  mobileOpen: boolean
  onCloseMobile: () => void
}) {
  const styles = useStyles()
  const location = useLocation()
  const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({})

  useEffect(() => {
    setExpandedItems((current) => {
      const nextState = { ...current }
      for (const item of items) {
        if (item.children?.length && nextState[item.id] === undefined) {
          nextState[item.id] = false
        }
      }
      return nextState
    })
  }, [items])

  useEffect(() => {
    const activeAncestorIds = new Set(collectExpandedAncestorIds(items, location.pathname))
    if (!activeAncestorIds.size) {
      return
    }

    setExpandedItems((current) => {
      let changed = false
      const nextState = { ...current }

      for (const itemId of activeAncestorIds) {
        if (!nextState[itemId]) {
          nextState[itemId] = true
          changed = true
        }
      }

      return changed ? nextState : current
    })
  }, [items, location.pathname])

  function toggleExpanded(itemId: string) {
    setExpandedItems((current) => ({
      ...current,
      [itemId]: !current[itemId],
    }))
  }

  return (
    <>
      <div className={styles.rail}>
        <div className={styles.list}>
          {items.map((item) => (
            <NavEntry
              key={item.id}
              item={item}
              collapsed={collapsed}
              pathname={location.pathname}
              expanded={expandedItems[item.id] ?? true}
              contentVisible={contentVisible}
              onToggleExpanded={() => toggleExpanded(item.id)}
            />
          ))}
        </div>
      </div>

      <Dialog open={mobileOpen} onOpenChange={(_, data) => (!data.open ? onCloseMobile() : undefined)}>
        <DialogSurface className={styles.dialogSurface}>
          <DialogBody>
            <DialogTitle>导航</DialogTitle>
            <DialogContent className={styles.mobileContent}>
              <div className={styles.list}>
                {items.map((item) => (
                  <NavEntry
                    key={`mobile-${item.id}`}
                    item={item}
                    collapsed={false}
                    pathname={location.pathname}
                    expanded={expandedItems[item.id] ?? true}
                    contentVisible={true}
                    onToggleExpanded={() => toggleExpanded(item.id)}
                  />
                ))}
              </div>
            </DialogContent>
          </DialogBody>
        </DialogSurface>
      </Dialog>
    </>
  )
}
