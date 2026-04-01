import { type CSSProperties, useEffect, useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { MessageBar, MessageBarBody, Spinner, makeStyles, tokens } from '@fluentui/react-components'
import { HeaderBar } from '@/features/shell/components/HeaderBar'
import { OpenTabsBar } from '@/features/shell/components/OpenTabsBar'
import { SideNav } from '@/features/shell/components/SideNav'
import { useAuthStore } from '@/features/session/auth.store'
import { useNavigationTreeQuery, useSpacesQuery } from '@/features/navigation/navigation.service'
import { resolveRouteByPath, resolveShellTabByPath } from '@/features/navigation/route-registry'
import { useShellStore } from '@/features/shell/store/useShellStore'
import { appConfig } from '@/shared/config/app-config'
import type { NavigationItem, SpaceKey } from '@/shared/types/navigation'

const useStyles = makeStyles({
  app: {
    minHeight: '100vh',
    display: 'grid',
    gridTemplateRows: 'auto 1fr',
    background: `linear-gradient(180deg, ${tokens.colorNeutralBackground2} 0%, ${tokens.colorNeutralBackground3} 100%)`,
  },
  body: {
    minHeight: 0,
    display: 'grid',
    gridTemplateColumns: 'var(--shell-sidebar-width) minmax(0, 1fr)',
    willChange: 'grid-template-columns',
    transitionProperty: 'grid-template-columns',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  sidebar: {
    minHeight: 0,
    padding: '16px var(--shell-sidebar-padding-x)',
    borderRight: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: tokens.colorNeutralBackground1,
    overflow: 'hidden',
    contain: 'layout paint',
    willChange: 'padding',
    transitionProperty: 'padding',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
    '@media (max-width: 960px)': {
      display: 'none',
    },
  },
  content: {
    minWidth: 0,
    display: 'grid',
    alignContent: 'start',
    gap: '16px',
    padding: '24px',
    '@media (max-width: 720px)': {
      padding: '16px',
    },
  },
  tabs: {
    minWidth: 0,
  },
  outlet: {
    minWidth: 0,
  },
  loading: {
    minHeight: '100vh',
    display: 'grid',
    placeItems: 'center',
  },
})

const NAV_CONTENT_APPEAR_DELAY_MS = 42
const NAV_COLLAPSE_SWAP_DELAY_MS = 68

function isVisibleForSpace(spaceVisibility: NavigationItem['spaceVisibility'], spaceKey: SpaceKey) {
  return spaceVisibility === 'all' || spaceVisibility.includes(spaceKey)
}

function filterNavigation(items: NavigationItem[], spaceKey: SpaceKey): NavigationItem[] {
  return items.reduce<NavigationItem[]>((result, item) => {
    if (!isVisibleForSpace(item.spaceVisibility, spaceKey)) {
      return result
    }

    const children = item.children ? filterNavigation(item.children, spaceKey) : undefined
    result.push({
      ...item,
      children,
    })
    return result
  }, [])
}

export function AppShell() {
  const styles = useStyles()
  const location = useLocation()
  const navigate = useNavigate()
  const {
    themeMode,
    navCollapsed,
    mobileNavOpen,
    currentSpaceKey,
    openTabs,
    tabsEnabled,
    toggleTheme,
    toggleNavCollapsed,
    setTabsEnabled,
    setMobileNavOpen,
    setCurrentSpaceKey,
    setActiveTopContext,
    registerTab,
    closeTab,
    clearTabs,
    toggleTabPinned,
    closeOtherTabs,
    closeTabsToLeft,
    closeTabsToRight,
    reorderTabs,
  } = useShellStore()
  const currentUser = useAuthStore((state) => state.currentUser)
  const signOut = useAuthStore((state) => state.signOut)
  const [navRenderCollapsed, setNavRenderCollapsed] = useState(navCollapsed)
  const [navContentVisible, setNavContentVisible] = useState(!navCollapsed)

  const spacesQuery = useSpacesQuery()
  const navigationQuery = useNavigationTreeQuery()

  useEffect(() => {
    if (navCollapsed) {
      setNavContentVisible(false)

      const timeoutId = window.setTimeout(() => {
        setNavRenderCollapsed(true)
      }, NAV_COLLAPSE_SWAP_DELAY_MS)

      return () => window.clearTimeout(timeoutId)
    }

    setNavRenderCollapsed(false)

    const timeoutId = window.setTimeout(() => {
      setNavContentVisible(true)
    }, NAV_CONTENT_APPEAR_DELAY_MS)

    return () => window.clearTimeout(timeoutId)
  }, [navCollapsed])

  useEffect(() => {
    if (spacesQuery.data && !spacesQuery.data.some((item) => item.key === currentSpaceKey)) {
      setCurrentSpaceKey(spacesQuery.data[0]?.key || 'default')
    }
  }, [currentSpaceKey, setCurrentSpaceKey, spacesQuery.data])

  useEffect(() => {
    const routeDefinition = resolveRouteByPath(location.pathname)
    const shellTab = resolveShellTabByPath(location.pathname)

    if (routeDefinition) {
      setActiveTopContext(routeDefinition.shellTitle)
    }

    if (tabsEnabled && shellTab) {
      registerTab(shellTab)
    }
  }, [location.pathname, registerTab, setActiveTopContext, tabsEnabled])

  if (spacesQuery.isLoading || navigationQuery.isLoading || !currentUser) {
    return (
      <div className={styles.loading}>
        <Spinner label="正在初始化壳层" />
      </div>
    )
  }

  if (spacesQuery.isError || navigationQuery.isError || !spacesQuery.data?.length) {
    return (
      <div className={styles.content}>
        <MessageBar>
          <MessageBarBody>壳层 mock 初始化失败，请检查本地数据或 Query 配置。</MessageBarBody>
        </MessageBar>
      </div>
    )
  }

  const currentSpace =
    spacesQuery.data.find((item) => item.key === currentSpaceKey) || spacesQuery.data[0]
  const filteredNavigation = filterNavigation(navigationQuery.data || [], currentSpace.key)
  const shellMotionStyle = {
    '--shell-sidebar-width': navCollapsed ? '72px' : '252px',
    '--shell-sidebar-padding-x': navCollapsed ? '8px' : '10px',
  } as CSSProperties
  const activeTab = openTabs.find((item) => item.path === location.pathname)

  function handleSelectTab(path: string) {
    if (path !== location.pathname) {
      navigate(path)
    }
  }

  function handleCloseTab(path: string) {
    const tabIndex = openTabs.findIndex((item) => item.path === path)
    if (tabIndex < 0) {
      return
    }

    const isActive = location.pathname === path
    const fallbackPath =
      openTabs[tabIndex + 1]?.path ||
      openTabs[tabIndex - 1]?.path ||
      appConfig.defaultRoute

    closeTab(path)

    if (isActive && fallbackPath !== path) {
      navigate(fallbackPath, { replace: true })
    }
  }

  function handleCloseOtherTabs(path: string) {
    const currentTab = openTabs.find((item) => item.path === location.pathname)
    const activeWillSurvive = location.pathname === path || currentTab?.pinned
    closeOtherTabs(path)

    if (!activeWillSurvive) {
      navigate(path, { replace: true })
    }
  }

  function handleCloseTabsToLeft(path: string) {
    const targetIndex = openTabs.findIndex((item) => item.path === path)
    const activeIndex = openTabs.findIndex((item) => item.path === location.pathname)
    const currentTab = openTabs[activeIndex]
    const activeWillSurvive =
      activeIndex < 0 ||
      activeIndex >= targetIndex ||
      currentTab?.pinned

    closeTabsToLeft(path)

    if (!activeWillSurvive) {
      navigate(path, { replace: true })
    }
  }

  function handleCloseTabsToRight(path: string) {
    const targetIndex = openTabs.findIndex((item) => item.path === path)
    const activeIndex = openTabs.findIndex((item) => item.path === location.pathname)
    const currentTab = openTabs[activeIndex]
    const activeWillSurvive =
      activeIndex < 0 ||
      activeIndex <= targetIndex ||
      currentTab?.pinned

    closeTabsToRight(path)

    if (!activeWillSurvive) {
      navigate(path, { replace: true })
    }
  }

  function handleSignOut() {
    clearTabs()
    signOut()
  }

  return (
    <div className={styles.app}>
      <HeaderBar
        navigationItems={filteredNavigation}
        currentUser={currentUser}
        darkMode={themeMode === 'dark'}
        navCollapsed={navCollapsed}
        tabsEnabled={tabsEnabled}
        onToggleTheme={toggleTheme}
        onToggleNav={toggleNavCollapsed}
        onSetTabsEnabled={setTabsEnabled}
        onOpenMobileNav={() => setMobileNavOpen(true)}
        onSignOut={handleSignOut}
      />

      <div className={styles.body} style={shellMotionStyle}>
        <aside className={styles.sidebar} style={shellMotionStyle}>
          <SideNav
            items={filteredNavigation}
            collapsed={navRenderCollapsed}
            contentVisible={navContentVisible}
            mobileOpen={mobileNavOpen}
            onCloseMobile={() => setMobileNavOpen(false)}
          />
        </aside>
        <main className={styles.content}>
          {tabsEnabled ? (
            <div className={styles.tabs}>
              <OpenTabsBar
                tabs={openTabs}
                activePath={location.pathname}
                onSelect={handleSelectTab}
                onClose={handleCloseTab}
                onTogglePinned={toggleTabPinned}
                onCloseOthers={handleCloseOtherTabs}
                onCloseLeft={handleCloseTabsToLeft}
                onCloseRight={handleCloseTabsToRight}
                onReorder={reorderTabs}
                activeTab={activeTab}
              />
            </div>
          ) : null}
          <div className={styles.outlet}>
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
