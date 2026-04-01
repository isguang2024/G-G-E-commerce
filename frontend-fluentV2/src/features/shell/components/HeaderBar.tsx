import {
  Avatar,
  Body1,
  Button,
  Caption1,
  Menu,
  MenuItem,
  MenuList,
  MenuPopover,
  MenuTrigger,
  Switch,
  Tooltip,
  makeStyles,
  mergeClasses,
  tokens,
} from '@fluentui/react-components'
import {
  AppsSettings20Regular,
  ArrowExit20Regular,
  DarkTheme20Regular,
  PanelLeftContract20Regular,
  PanelLeftExpand20Regular,
  Search20Regular,
  WeatherSunny20Regular,
} from '@fluentui/react-icons'
import { type FocusEvent, useEffect, useRef, useState } from 'react'
import { MenuSearchDialog } from '@/features/shell/components/MenuSearchDialog'
import { AppLogo } from '@/shared/ui/AppLogo'
import type { NavigationItem } from '@/shared/types/navigation'
import type { SessionUser } from '@/shared/types/session'

const useStyles = makeStyles({
  root: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: '16px',
    padding: '8px 16px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: tokens.colorNeutralBackground1,
    position: 'sticky',
    top: 0,
    zIndex: 3,
    '@media (max-width: 960px)': {
      flexWrap: 'wrap',
    },
  },
  left: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    minWidth: 0,
    flex: 1,
  },
  brandDock: {
    position: 'relative',
    display: 'flex',
    alignItems: 'center',
    minWidth: '228px',
    height: '40px',
    paddingLeft: '8px',
    overflow: 'hidden',
  },
  brandPanel: {
    display: 'flex',
    alignItems: 'center',
    minWidth: 0,
    transitionProperty: 'opacity, transform',
    transitionDuration: tokens.durationNormal,
    transitionTimingFunction: tokens.curveEasyEase,
  },
  brandPanelHidden: {
    opacity: 0,
    transform: 'translateX(14px)',
  },
  actionTray: {
    position: 'absolute',
    inset: '0',
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    opacity: 0,
    transform: 'translateX(-10px)',
    pointerEvents: 'none',
    transitionProperty: 'opacity, transform',
    transitionDuration: tokens.durationNormal,
    transitionTimingFunction: tokens.curveEasyEase,
  },
  actionTrayVisible: {
    opacity: 1,
    transform: 'translateX(0)',
    pointerEvents: 'auto',
  },
  right: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    flexWrap: 'wrap',
    justifyContent: 'flex-end',
  },
  shellButton: {
    minWidth: '32px',
    width: '32px',
    height: '32px',
    paddingLeft: '0',
    paddingRight: '0',
    borderRadius: tokens.borderRadiusCircular,
    color: tokens.colorNeutralForeground2,
    backgroundColor: tokens.colorTransparentBackground,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1Hover,
    },
  },
  mobileOnly: {
    display: 'none',
    '@media (max-width: 960px)': {
      display: 'inline-flex',
    },
  },
  mobileActions: {
    display: 'none',
    '@media (max-width: 960px)': {
      display: 'flex',
      alignItems: 'center',
      gap: '4px',
    },
  },
  desktopOnly: {
    '@media (max-width: 960px)': {
      display: 'none',
    },
  },
  mobileBrand: {
    display: 'none',
    '@media (max-width: 960px)': {
      display: 'flex',
      alignItems: 'center',
      minWidth: 0,
    },
  },
  userButton: {
    minWidth: 'auto',
    paddingLeft: '8px',
    paddingRight: '10px',
    borderRadius: tokens.borderRadiusCircular,
    color: tokens.colorNeutralForeground2,
  },
  userInfo: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    minWidth: 0,
  },
  userText: {
    minWidth: 0,
  },
  userMeta: {
    color: tokens.colorNeutralForeground3,
    '@media (max-width: 1200px)': {
      display: 'none',
    },
  },
  settingsPanel: {
    padding: '8px 12px',
    minWidth: '180px',
  },
})

export function HeaderBar({
  navigationItems,
  currentUser,
  darkMode,
  navCollapsed,
  tabsEnabled,
  onToggleTheme,
  onToggleNav,
  onSetTabsEnabled,
  onOpenMobileNav,
  onSignOut,
}: {
  navigationItems: NavigationItem[]
  currentUser: SessionUser
  darkMode: boolean
  navCollapsed: boolean
  tabsEnabled: boolean
  onToggleTheme: () => void
  onToggleNav: () => void
  onSetTabsEnabled: (enabled: boolean) => void
  onOpenMobileNav: () => void
  onSignOut: () => void
}) {
  const styles = useStyles()
  const dockRef = useRef<HTMLDivElement>(null)
  const [dockActive, setDockActive] = useState(false)
  const [searchOpen, setSearchOpen] = useState(false)

  useEffect(() => {
    function handleGlobalKeydown(event: KeyboardEvent) {
      const isMac = navigator.platform.toUpperCase().includes('MAC')
      const modifierPressed = isMac ? event.metaKey : event.ctrlKey
      if (!modifierPressed || event.key.toLowerCase() !== 'k') {
        return
      }

      event.preventDefault()
      setSearchOpen(true)
    }

    window.addEventListener('keydown', handleGlobalKeydown)
    return () => window.removeEventListener('keydown', handleGlobalKeydown)
  }, [])

  function handleDockBlur(event: FocusEvent<HTMLDivElement>) {
    if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
      setDockActive(false)
    }
  }

  function renderTabsSettingsButton() {
    return (
      <Menu>
        <MenuTrigger disableButtonEnhancement>
          <Tooltip content="界面设置" relationship="label">
            <Button
              aria-label="界面设置"
              className={styles.shellButton}
              appearance="subtle"
              icon={<AppsSettings20Regular />}
            />
          </Tooltip>
        </MenuTrigger>
        <MenuPopover>
          <div className={styles.settingsPanel}>
            <Switch
              checked={tabsEnabled}
              label="显示标签栏"
              onChange={(_, data) => onSetTabsEnabled(data.checked)}
            />
          </div>
        </MenuPopover>
      </Menu>
    )
  }

  return (
    <header className={styles.root}>
      <div className={styles.left}>
        <Button
          className={mergeClasses(styles.shellButton, styles.mobileOnly)}
          appearance="subtle"
          aria-label="打开菜单"
          icon={<PanelLeftExpand20Regular />}
          onClick={onOpenMobileNav}
        />
        <div className={styles.mobileActions}>
          {renderTabsSettingsButton()}
          <Tooltip content={darkMode ? '切换到亮色模式' : '切换到暗色模式'} relationship="label">
            <Button
              appearance="subtle"
              aria-label={darkMode ? '切换到亮色模式' : '切换到暗色模式'}
              className={styles.shellButton}
              icon={darkMode ? <WeatherSunny20Regular /> : <DarkTheme20Regular />}
              onClick={onToggleTheme}
            />
          </Tooltip>
          <Tooltip content="搜索菜单" relationship="label">
            <Button
              aria-label="搜索菜单"
              className={styles.shellButton}
              appearance="subtle"
              icon={<Search20Regular />}
              onClick={() => setSearchOpen(true)}
            />
          </Tooltip>
        </div>
        <div
          ref={dockRef}
          className={mergeClasses(styles.brandDock, styles.desktopOnly)}
          onMouseEnter={() => setDockActive(true)}
          onMouseLeave={() => setDockActive(false)}
          onFocusCapture={() => setDockActive(true)}
          onBlurCapture={handleDockBlur}
        >
          <div className={mergeClasses(styles.brandPanel, dockActive && styles.brandPanelHidden)}>
            <AppLogo />
          </div>
          <div className={mergeClasses(styles.actionTray, dockActive && styles.actionTrayVisible)}>
            <Tooltip content={navCollapsed ? '展开侧栏' : '收起侧栏'} relationship="label">
              <Button
                className={styles.shellButton}
                appearance="subtle"
                aria-label={navCollapsed ? '展开侧栏' : '收起侧栏'}
                icon={navCollapsed ? <PanelLeftExpand20Regular /> : <PanelLeftContract20Regular />}
                onClick={onToggleNav}
              />
            </Tooltip>
            {renderTabsSettingsButton()}
            <Tooltip content={darkMode ? '切换到亮色模式' : '切换到暗色模式'} relationship="label">
              <Button
                appearance="subtle"
                aria-label={darkMode ? '切换到亮色模式' : '切换到暗色模式'}
                className={styles.shellButton}
                icon={darkMode ? <WeatherSunny20Regular /> : <DarkTheme20Regular />}
                onClick={onToggleTheme}
              />
            </Tooltip>
            <Tooltip content="搜索菜单" relationship="label">
              <Button
                aria-label="搜索菜单"
                className={styles.shellButton}
                appearance="subtle"
                icon={<Search20Regular />}
                onClick={() => setSearchOpen(true)}
              />
            </Tooltip>
          </div>
        </div>
      </div>

      <div className={styles.right}>
        <Menu>
          <MenuTrigger disableButtonEnhancement>
            <Button className={styles.userButton} appearance="subtle">
              <div className={styles.userInfo}>
                <Avatar name={currentUser.displayName} size={28} />
                <div className={styles.userText}>
                  <Body1>{currentUser.displayName}</Body1>
                  <Caption1 className={styles.userMeta}>{currentUser.title}</Caption1>
                </div>
              </div>
            </Button>
          </MenuTrigger>
          <MenuPopover>
            <MenuList>
              <MenuItem>{currentUser.email}</MenuItem>
              <MenuItem icon={<ArrowExit20Regular />} onClick={onSignOut}>
                退出登录
              </MenuItem>
            </MenuList>
          </MenuPopover>
        </Menu>
      </div>
      <MenuSearchDialog items={navigationItems} open={searchOpen} onOpenChange={setSearchOpen} />
    </header>
  )
}
