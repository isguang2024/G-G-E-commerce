import {
  Body1Strong,
  Button,
  Caption1,
  Dialog,
  DialogBody,
  DialogContent,
  DialogSurface,
  Menu,
  MenuDivider,
  MenuItem,
  MenuList,
  MenuPopover,
  MenuTrigger,
  Tooltip,
  makeStyles,
  mergeClasses,
  tokens,
} from '@fluentui/react-components'
import {
  ChevronDown16Regular,
  ChevronRight16Regular,
  Dismiss20Regular,
} from '@fluentui/react-icons'
import { type CSSProperties, useEffect, useMemo } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'
import { useShellStore } from '@/features/shell/store/useShellStore'
import { AppIcon } from '@/shared/ui/AppIcon'
import type { NavigationItem } from '@/shared/types/navigation'

type NavMode = 'desktop-expanded' | 'desktop-collapsed' | 'mobile-drawer'

const EXPANDED_NAV_MIN_WIDTH = 236
const EXPANDED_NAV_MAX_WIDTH = 360
const FLYOUT_MIN_WIDTH = 240
const FLYOUT_MAX_WIDTH = 320
const MENU_TOOLTIP_DELAY_MS = 1000

const useStyles = makeStyles({
  rail: {
    display: 'grid',
    gap: '6px',
    alignContent: 'start',
    height: '100%',
  },
  list: {
    display: 'grid',
    gap: '4px',
  },
  node: {
    display: 'grid',
    gap: '2px',
  },
  nodeRow: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 1fr) auto',
    alignItems: 'center',
    gap: '4px',
  },
  nodeRowSingle: {
    gridTemplateColumns: 'minmax(0, 1fr)',
  },
  itemLink: {
    display: 'grid',
    minWidth: 0,
  },
  itemButton: {
    width: '100%',
    justifyContent: 'flex-start',
    gap: '8px',
    minHeight: '36px',
    paddingLeft: '10px',
    paddingRight: '10px',
    borderRadius: tokens.borderRadiusMedium,
    fontSize: tokens.fontSizeBase200,
    lineHeight: tokens.lineHeightBase200,
    color: tokens.colorNeutralForeground2,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorTransparentStroke}`,
    transitionProperty: 'background-color, border-color, color, transform',
    transitionDuration: '180ms',
    transitionTimingFunction: tokens.curveEasyEase,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1Hover,
      border: `1px solid ${tokens.colorNeutralStroke1Hover}`,
      color: tokens.colorNeutralForeground1,
    },
  },
  itemButtonNested: {
    minHeight: '31px',
    borderRadius: tokens.borderRadiusMedium,
    fontSize: tokens.fontSizeBase100,
    lineHeight: tokens.lineHeightBase100,
  },
  itemButtonNestedDeep: {
    minHeight: '29px',
    paddingLeft: '8px',
    paddingRight: '8px',
    fontSize: tokens.fontSizeBase100,
    lineHeight: tokens.lineHeightBase100,
  },
  itemButtonAncestor: {
    color: tokens.colorNeutralForeground1,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
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
  itemToggle: {
    minWidth: '28px',
    width: '28px',
    height: '28px',
    paddingLeft: '0',
    paddingRight: '0',
    justifyContent: 'center',
    transitionProperty: 'opacity, transform',
    transitionDuration: '180ms',
    transitionTimingFunction: tokens.curveEasyEase,
  },
  itemToggleHidden: {
    opacity: 0,
    transform: 'translateX(-6px)',
    pointerEvents: 'none',
  },
  childRegion: {
    display: 'grid',
    gridTemplateRows: '1fr',
    opacity: 1,
    transform: 'translateY(0)',
    transitionProperty: 'grid-template-rows, opacity, transform',
    transitionDuration: '180ms',
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
  childList: {
    display: 'grid',
    gap: '2px',
    marginLeft: '6px',
    paddingLeft: '6px',
    borderLeft: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  childListDeep: {
    marginLeft: '4px',
    paddingLeft: '4px',
    borderLeft: `1px solid ${tokens.colorNeutralStroke1}`,
  },
  label: {
    display: 'inline-block',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    fontSize: tokens.fontSizeBase200,
    lineHeight: tokens.lineHeightBase200,
    transitionProperty: 'opacity, transform',
    transitionDuration: '180ms',
    transitionTimingFunction: tokens.curveEasyEase,
  },
  labelHidden: {
    opacity: 0,
    transform: 'translateX(-6px)',
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
    width: '36px',
    minWidth: '36px',
    height: '36px',
    paddingLeft: '0',
    paddingRight: '0',
    justifyContent: 'center',
    gap: '0',
    borderRadius: tokens.borderRadiusMedium,
    fontSize: tokens.fontSizeBase100,
    color: tokens.colorNeutralForeground2,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorTransparentStroke}`,
    transitionProperty: 'background-color, border-color, color',
    transitionDuration: '180ms',
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
  flyoutPopover: {
    width: 'var(--shell-flyout-width, 240px)',
    minWidth: 'var(--shell-flyout-width, 240px)',
    maxWidth: 'min(var(--shell-flyout-width, 240px), calc(100vw - 24px))',
    borderRadius: tokens.borderRadiusMedium,
    boxShadow: '0 8px 18px rgba(0, 0, 0, 0.10)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    overflow: 'hidden',
  },
  flyoutItem: {
    borderRadius: tokens.borderRadiusMedium,
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    minWidth: 0,
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    fontSize: tokens.fontSizeBase100,
    lineHeight: tokens.lineHeightBase100,
  },
  flyoutItemAncestor: {
    backgroundColor: tokens.colorNeutralBackground2,
    color: tokens.colorNeutralForeground1,
  },
  flyoutItemActive: {
    backgroundColor: tokens.colorBrandBackground2,
    color: tokens.colorBrandForeground1,
  },
  dialogSurface: {
    width: 'calc(100vw - 12px)',
    maxWidth: 'none',
    height: 'calc(100vh - 12px)',
    maxHeight: 'calc(100vh - 12px)',
    borderRadius: tokens.borderRadiusXLarge,
    padding: '0',
    overflow: 'hidden',
  },
  mobileShell: {
    height: '100%',
    display: 'grid',
    gridTemplateRows: 'auto 1fr',
  },
  mobileHeader: {
    display: 'flex',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: '12px',
    padding: '14px 16px 10px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  mobileHeaderCopy: {
    display: 'grid',
    gap: '2px',
  },
  mobileBody: {
    minHeight: 0,
    padding: '8px 10px 14px',
  },
  mobileScroll: {
    height: '100%',
    overflowY: 'auto',
    paddingRight: '4px',
  },
})

function isCurrentPath(pathname: string, targetPath: string) {
  if (!targetPath) {
    return false
  }

  return pathname === targetPath || pathname.startsWith(`${targetPath}/`)
}

function hasChildren(item: NavigationItem) {
  return Boolean(item.children?.length)
}

function renderNavIcon(icon?: NavigationItem['icon']) {
  return icon ? <AppIcon icon={icon} /> : undefined
}

function canNavigateItem(item: NavigationItem) {
  if (!item.path) {
    return false
  }

  return item.status === 'implemented'
}

function collectExpandableItemIds(items: NavigationItem[]): string[] {
  return items.flatMap((item) => {
    if (!hasChildren(item)) {
      return []
    }

    return [item.id, ...collectExpandableItemIds(item.children || [])]
  })
}

function findActiveTrail(
  items: NavigationItem[],
  pathname: string,
  trail: NavigationItem[] = [],
): NavigationItem[] {
  for (const item of items) {
    const nextTrail = [...trail, item]

    if (hasChildren(item)) {
      const matchedChildTrail = findActiveTrail(item.children || [], pathname, nextTrail)
      if (matchedChildTrail.length) {
        return matchedChildTrail
      }
    }

    if (isCurrentPath(pathname, item.path)) {
      return nextTrail
    }
  }

  return []
}

function resolveBranchPadding(mode: NavMode, depth: number) {
  if (mode === 'mobile-drawer') {
    return 10 + depth * 12
  }

  if (depth === 0) {
    return 10
  }

  if (depth === 1) {
    return 9
  }

  return 8 + (depth - 1) * 8
}

function estimateTextWidth(label: string) {
  const units = Array.from(label).reduce((total, char) => {
    if (/\s/u.test(char)) {
      return total + 0.5
    }

    return total + (/[^\u0000-\u00ff]/u.test(char) ? 1.4 : 0.9)
  }, 0)

  return units * 5.8
}

function clampWidth(width: number, minWidth: number, maxWidth: number) {
  return Math.min(maxWidth, Math.max(minWidth, Math.ceil(width)))
}

function collectVisibleWidthCandidates(
  items: NavigationItem[],
  expandedIds: Set<string>,
  depth = 0,
): number[] {
  return items.flatMap((item) => {
    const hasNestedChildren = hasChildren(item)
  const baseWidth =
      88 +
      depth * 12 +
      estimateTextWidth(item.label) +
      (hasNestedChildren ? 32 : 0)

    if (!hasNestedChildren || !expandedIds.has(item.id)) {
      return [baseWidth]
    }

    return [baseWidth, ...collectVisibleWidthCandidates(item.children || [], expandedIds, depth + 1)]
  })
}

function resolveExpandedSidebarWidth(items: NavigationItem[], expandedIds: Set<string>) {
  const candidates = collectVisibleWidthCandidates(items, expandedIds)
  return clampWidth(
    candidates.length ? Math.max(...candidates) + 8 : EXPANDED_NAV_MIN_WIDTH,
    EXPANDED_NAV_MIN_WIDTH,
    EXPANDED_NAV_MAX_WIDTH,
  )
}

function resolveFlyoutWidth(items: NavigationItem[]) {
  const widestItem = items.reduce((maxWidth, item) => {
    const candidate = 74 + estimateTextWidth(item.label) + (hasChildren(item) ? 26 : 0)
    return Math.max(maxWidth, candidate)
  }, FLYOUT_MIN_WIDTH)

  return clampWidth(widestItem + 8, FLYOUT_MIN_WIDTH, FLYOUT_MAX_WIDTH)
}

function CollapsedFlyoutItem({
  item,
  pathname,
  activeTrailIds,
  onNavigate,
}: {
  item: NavigationItem
  pathname: string
  activeTrailIds: Set<string>
  onNavigate: (path: string) => void
}) {
  const styles = useStyles()
  const exactActive = pathname === item.path
  const ancestorActive = activeTrailIds.has(item.id) && !exactActive
  const flyoutItemClassName = mergeClasses(
    styles.flyoutItem,
    ancestorActive && styles.flyoutItemAncestor,
    exactActive && styles.flyoutItemActive,
  )
  const childFlyoutWidth = hasChildren(item) ? resolveFlyoutWidth(item.children || []) : FLYOUT_MIN_WIDTH
  const canNavigateSelf = canNavigateItem(item)

  if (hasChildren(item)) {
    return (
      <Tooltip relationship="label" content={item.label} showDelay={MENU_TOOLTIP_DELAY_MS}>
        <Menu openOnHover positioning={{ position: 'after', align: 'top', offset: 8 }}>
          <MenuTrigger disableButtonEnhancement>
            <MenuItem className={flyoutItemClassName} icon={renderNavIcon(item.icon)}>
              {item.label}
            </MenuItem>
          </MenuTrigger>
          <MenuPopover
            className={styles.flyoutPopover}
            style={{ '--shell-flyout-width': `${childFlyoutWidth}px` } as CSSProperties}
          >
            <MenuList>
              {canNavigateSelf ? (
                <>
                  <MenuItem
                    className={flyoutItemClassName}
                    icon={renderNavIcon(item.icon)}
                    onClick={() => onNavigate(item.path)}
                  >
                    {item.label}
                  </MenuItem>
                  <MenuDivider />
                </>
              ) : null}
              {(item.children || []).map((child) => (
                <CollapsedFlyoutItem
                  key={child.id}
                  item={child}
                  pathname={pathname}
                  activeTrailIds={activeTrailIds}
                  onNavigate={onNavigate}
                />
              ))}
            </MenuList>
          </MenuPopover>
        </Menu>
      </Tooltip>
    )
  }

  return (
    <Tooltip relationship="label" content={item.label} showDelay={MENU_TOOLTIP_DELAY_MS}>
      <MenuItem
        className={flyoutItemClassName}
        icon={renderNavIcon(item.icon)}
        onClick={() => onNavigate(item.path)}
      >
        {item.label}
      </MenuItem>
    </Tooltip>
  )
}

function NavNode({
  item,
  depth,
  mode,
  pathname,
  contentVisible,
  expandedIds,
  activeTrailIds,
  onToggle,
  onNavigate,
}: {
  item: NavigationItem
  depth: number
  mode: NavMode
  pathname: string
  contentVisible: boolean
  expandedIds: Set<string>
  activeTrailIds: Set<string>
  onToggle: (itemId: string) => void
  onNavigate: (path: string) => void
}) {
  const styles = useStyles()
  const itemHasChildren = hasChildren(item)
  const exactActive = pathname === item.path
  const ancestorActive = activeTrailIds.has(item.id) && !exactActive
  const expanded = expandedIds.has(item.id)
  const navigateEnabled = canNavigateItem(item)
  const paddingLeft = resolveBranchPadding(mode, depth)

  if (mode === 'desktop-collapsed') {
    const collapsedTriggerClassName = mergeClasses(
      styles.collapsedTrigger,
      activeTrailIds.has(item.id) && styles.collapsedTriggerActive,
    )

    if (itemHasChildren) {
      return (
        <div className={styles.collapsedSection}>
          <Menu openOnHover positioning={{ position: 'after', align: 'top', offset: 14 }}>
            <MenuTrigger disableButtonEnhancement>
              <Tooltip relationship="label" content={item.label} showDelay={MENU_TOOLTIP_DELAY_MS}>
                <Button
                  aria-label={item.label}
                  className={collapsedTriggerClassName}
                  appearance="subtle"
                  icon={renderNavIcon(item.icon)}
                />
              </Tooltip>
            </MenuTrigger>
            <MenuPopover
              className={styles.flyoutPopover}
              style={{ '--shell-flyout-width': `${resolveFlyoutWidth(item.children || [])}px` } as CSSProperties}
            >
              <MenuList>
                {(item.children || []).map((child) => (
                  <CollapsedFlyoutItem
                    key={child.id}
                    item={child}
                    pathname={pathname}
                    activeTrailIds={activeTrailIds}
                    onNavigate={onNavigate}
                  />
                ))}
              </MenuList>
            </MenuPopover>
          </Menu>
        </div>
      )
    }

    if (!navigateEnabled) {
      return (
        <div className={styles.collapsedSection}>
          <Tooltip relationship="label" content={item.label} showDelay={MENU_TOOLTIP_DELAY_MS}>
            <Button
              aria-label={item.label}
              className={collapsedTriggerClassName}
              appearance="subtle"
              icon={renderNavIcon(item.icon)}
              disabled
            />
          </Tooltip>
        </div>
      )
    }

    return (
      <div className={styles.collapsedSection}>
        <Tooltip relationship="label" content={item.label} showDelay={MENU_TOOLTIP_DELAY_MS}>
          <NavLink className={styles.collapsedLink} to={item.path}>
            <Button
              aria-label={item.label}
              className={collapsedTriggerClassName}
              appearance="subtle"
              icon={renderNavIcon(item.icon)}
            />
          </NavLink>
        </Tooltip>
      </div>
    )
  }

  const buttonClassName = mergeClasses(
    styles.itemButton,
    depth > 0 && styles.itemButtonNested,
    depth > 1 && styles.itemButtonNestedDeep,
    ancestorActive && styles.itemButtonAncestor,
    exactActive && styles.itemButtonActive,
  )
  const labelClassName = mergeClasses(styles.label, !contentVisible && styles.labelHidden)
  const toggleClassName = mergeClasses(styles.itemToggle, !contentVisible && styles.itemToggleHidden)
  const childRegionClassName = mergeClasses(
    styles.childRegion,
    !expanded && styles.childRegionCollapsed,
    expanded && !contentVisible && styles.childRegionHidden,
  )
  const rowClassName = mergeClasses(styles.nodeRow, !itemHasChildren && styles.nodeRowSingle)

  const button = (
    <Tooltip relationship="label" content={item.label} showDelay={MENU_TOOLTIP_DELAY_MS}>
      <Button
        className={buttonClassName}
        appearance="subtle"
        icon={renderNavIcon(item.icon)}
        style={{ paddingLeft, paddingRight: mode === 'mobile-drawer' ? 14 : 12 }}
        onClick={!navigateEnabled && itemHasChildren ? () => onToggle(item.id) : undefined}
      >
        <span className={labelClassName}>{item.label}</span>
      </Button>
    </Tooltip>
  )

  return (
    <div className={styles.node}>
      <div className={rowClassName}>
        {navigateEnabled ? (
          <NavLink className={styles.itemLink} to={item.path}>
            {button}
          </NavLink>
        ) : (
          button
        )}

        {itemHasChildren ? (
          <Button
            className={toggleClassName}
            appearance="subtle"
            aria-label={expanded ? `收起 ${item.label}` : `展开 ${item.label}`}
            icon={expanded ? <ChevronDown16Regular /> : <ChevronRight16Regular />}
            onClick={() => onToggle(item.id)}
          />
        ) : null}
      </div>

      {itemHasChildren ? (
        <div className={childRegionClassName}>
          <div className={styles.childRegionInner}>
            <div className={mergeClasses(styles.childList, depth > 0 && styles.childListDeep)}>
              {(item.children || []).map((child) => (
                <NavNode
                  key={child.id}
                  item={child}
                  depth={depth + 1}
                  mode={mode}
                  pathname={pathname}
                  contentVisible={contentVisible}
                  expandedIds={expandedIds}
                  activeTrailIds={activeTrailIds}
                  onToggle={onToggle}
                  onNavigate={onNavigate}
                />
              ))}
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
  currentSpaceKey,
  onExpandedWidthChange,
  onCloseMobile,
}: {
  items: NavigationItem[]
  collapsed: boolean
  contentVisible: boolean
  mobileOpen: boolean
  currentSpaceKey: string
  onExpandedWidthChange?: (width: number) => void
  onCloseMobile: () => void
}) {
  const styles = useStyles()
  const location = useLocation()
  const navigate = useNavigate()
  const navExpandedBySpace = useShellStore((state) => state.navExpandedBySpace)
  const navCollapsedBySpace = useShellStore((state) => state.navCollapsedBySpace)
  const toggleNavItemExpanded = useShellStore((state) => state.toggleNavItemExpanded)
  const toggleNavItemCollapsed = useShellStore((state) => state.toggleNavItemCollapsed)
  const pruneNavExpandedForSpace = useShellStore((state) => state.pruneNavExpandedForSpace)
  const pruneNavCollapsedForSpace = useShellStore((state) => state.pruneNavCollapsedForSpace)

  const activeTrail = useMemo(() => findActiveTrail(items, location.pathname), [items, location.pathname])
  const activeTrailIds = useMemo(() => new Set(activeTrail.map((item) => item.id)), [activeTrail])
  const activeExpandedIds = useMemo(
    () => activeTrail.filter((item) => hasChildren(item)).map((item) => item.id),
    [activeTrail],
  )
  const storedExpandedIds = navExpandedBySpace[currentSpaceKey] || []
  const storedCollapsedIds = navCollapsedBySpace[currentSpaceKey] || []
  const expandedIds = useMemo(
    () =>
      new Set(
        [...storedExpandedIds, ...activeExpandedIds].filter((itemId) => !storedCollapsedIds.includes(itemId)),
      ),
    [activeExpandedIds, storedCollapsedIds, storedExpandedIds],
  )
  const expandedSidebarWidth = useMemo(
    () => resolveExpandedSidebarWidth(items, expandedIds),
    [expandedIds, items],
  )

  useEffect(() => {
    const validItemIds = collectExpandableItemIds(items)
    pruneNavExpandedForSpace(currentSpaceKey, validItemIds)
    pruneNavCollapsedForSpace(currentSpaceKey, validItemIds)
  }, [currentSpaceKey, items, pruneNavCollapsedForSpace, pruneNavExpandedForSpace])

  useEffect(() => {
    if (mobileOpen) {
      onCloseMobile()
    }
  }, [location.pathname, mobileOpen, onCloseMobile])

  useEffect(() => {
    onExpandedWidthChange?.(collapsed ? EXPANDED_NAV_MIN_WIDTH : expandedSidebarWidth)
  }, [collapsed, expandedSidebarWidth, onExpandedWidthChange])

  function handleToggle(itemId: string) {
    if (expandedIds.has(itemId)) {
      toggleNavItemExpanded(currentSpaceKey, itemId)
      toggleNavItemCollapsed(currentSpaceKey, itemId)
      return
    }

    toggleNavItemCollapsed(currentSpaceKey, itemId)
    toggleNavItemExpanded(currentSpaceKey, itemId)
  }

  function handleNavigate(path: string) {
    if (!path) {
      return
    }

    navigate(path)
    onCloseMobile()
  }

  return (
    <>
      <div className={styles.rail}>
        <div className={styles.list}>
          {items.map((item) => (
            <NavNode
              key={item.id}
              item={item}
              depth={0}
              mode={collapsed ? 'desktop-collapsed' : 'desktop-expanded'}
              pathname={location.pathname}
              contentVisible={contentVisible}
              expandedIds={expandedIds}
              activeTrailIds={activeTrailIds}
              onToggle={handleToggle}
              onNavigate={handleNavigate}
            />
          ))}
        </div>
      </div>

        <Dialog open={mobileOpen} onOpenChange={(_, data) => (!data.open ? onCloseMobile() : undefined)}>
        <DialogSurface className={styles.dialogSurface}>
          <DialogBody className={styles.mobileShell}>
            <div className={styles.mobileHeader}>
              <div className={styles.mobileHeaderCopy}>
                <Body1Strong>导航</Body1Strong>
                <Caption1>支持多级菜单展开，点击叶子节点后会自动关闭抽屉。</Caption1>
              </div>
              <Button
                appearance="subtle"
                aria-label="关闭导航"
                icon={<Dismiss20Regular />}
                onClick={onCloseMobile}
              />
            </div>
            <DialogContent className={styles.mobileBody}>
              <div className={styles.mobileScroll}>
                <div className={styles.list}>
                  {items.map((item) => (
                    <NavNode
                      key={`mobile-${item.id}`}
                      item={item}
                      depth={0}
                      mode="mobile-drawer"
                      pathname={location.pathname}
                      contentVisible={true}
                      expandedIds={expandedIds}
                      activeTrailIds={activeTrailIds}
                      onToggle={handleToggle}
                      onNavigate={handleNavigate}
                    />
                  ))}
                </div>
              </div>
            </DialogContent>
          </DialogBody>
        </DialogSurface>
      </Dialog>
    </>
  )
}
