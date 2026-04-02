import { type CSSProperties, useEffect, useMemo, useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Spinner, makeStyles, tokens } from '@fluentui/react-components'
import { useAuthStore } from '@/features/auth/auth.store'
import { useNavigationItems, useRouteContext, useRuntimeNavigationManifestQuery, useMenuSpacesQuery } from '@/features/navigation/navigation.service'
import { getLocalRouteDefinitionByPath, getNavigationGroupLabel } from '@/features/navigation/route-registry'
import { HeaderBar } from '@/features/shell/components/HeaderBar'
import { OpenTabsBar } from '@/features/shell/components/OpenTabsBar'
import { SideNav } from '@/features/shell/components/SideNav'
import { useShellStore } from '@/features/shell/store/useShellStore'
import { invalidateSpaceScopedQueries } from '@/shared/api/query-client'
import { appConfig } from '@/shared/config/app-config'
import { PageStatusBanner } from '@/shared/ui/PageStatusBanner'
import type { SessionUser } from '@/shared/types/session'

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
    transitionDuration: '180ms',
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
    transitionDuration: '180ms',
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
const EXPANDED_NAV_BASE_WIDTH = 252

export function AppShell() {
  const styles = useStyles()
  const location = useLocation()
  const navigate = useNavigate()
  const currentUser = useAuthStore((state) => state.currentUser)
  const clearAuth = useAuthStore((state) => state.clearAuth)
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
    replaceTabs,
    closeTab,
    clearTabs,
    toggleTabPinned,
    closeOtherTabs,
    closeTabsToLeft,
    closeTabsToRight,
    reorderTabs,
  } = useShellStore()
  const shellUser: SessionUser = currentUser || {
    id: 'shell-reset',
    username: 'shell-reset',
    displayName: '壳层重置',
    email: '',
    avatarUrl: '',
    phone: '',
    status: 'inactive',
    isSuperAdmin: false,
    currentTenantId: null,
    roles: [],
    actions: [],
    badges: [],
  }
  const runtimeEnabled = Boolean(currentUser)
  const spacesQuery = useMenuSpacesQuery({ enabled: runtimeEnabled })
  const navigationItemsQuery = useNavigationItems({ enabled: runtimeEnabled })
  const manifestQuery = useRuntimeNavigationManifestQuery(currentSpaceKey, { enabled: runtimeEnabled })
  const routeContextQuery = useRouteContext(undefined, { enabled: runtimeEnabled })
  const [navRenderCollapsed, setNavRenderCollapsed] = useState(navCollapsed)
  const [navContentVisible, setNavContentVisible] = useState(!navCollapsed)
  const [navExpandedWidth, setNavExpandedWidth] = useState(EXPANDED_NAV_BASE_WIDTH)

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
    const serverSpaceKey = manifestQuery.data?.currentSpace?.space.key
    if (serverSpaceKey && serverSpaceKey !== currentSpaceKey) {
      setCurrentSpaceKey(serverSpaceKey)
    }
  }, [currentSpaceKey, manifestQuery.data?.currentSpace?.space.key, setCurrentSpaceKey])

  useEffect(() => {
    if (!openTabs.length) {
      return
    }

    let hasChanged = false
    const nextTabs = openTabs.map((tab) => {
      const localRoute = getLocalRouteDefinitionByPath(tab.path)
      if (!localRoute) {
        return tab
      }

      const nextLabel = localRoute.shellTitle
      const nextGroup = localRoute.group
      const nextGroupLabel = getNavigationGroupLabel(localRoute.group)

      if (tab.label === nextLabel && tab.group === nextGroup && tab.groupLabel === nextGroupLabel) {
        return tab
      }

      hasChanged = true
      return {
        ...tab,
        label: nextLabel,
        group: nextGroup,
        groupLabel: nextGroupLabel,
      }
    })

    if (hasChanged) {
      replaceTabs(nextTabs)
    }
  }, [openTabs, replaceTabs])

  useEffect(() => {
    if (!routeContextQuery.context) {
      return
    }

    setActiveTopContext(routeContextQuery.context.title)

    if (tabsEnabled) {
      registerTab({
        routeId: routeContextQuery.context.routeId,
        path: routeContextQuery.context.path,
        label: routeContextQuery.context.title,
        group: routeContextQuery.context.group,
        groupLabel: routeContextQuery.context.groupLabel,
      })
    }
  }, [registerTab, routeContextQuery.context, setActiveTopContext, tabsEnabled])

  const currentSpace = useMemo(
    () =>
      spacesQuery.data?.find((item) => item.key === currentSpaceKey) ||
      manifestQuery.data?.currentSpace?.space ||
      null,
    [currentSpaceKey, manifestQuery.data?.currentSpace?.space, spacesQuery.data],
  )
  const activeTab = openTabs.find((item) => item.path === location.pathname)

  if (spacesQuery.isLoading || manifestQuery.isLoading || navigationItemsQuery.isLoading) {
    return (
      <div className={styles.loading}>
        <Spinner label="正在初始化壳层" />
      </div>
    )
  }

  if (spacesQuery.isError || manifestQuery.isError || navigationItemsQuery.isError) {
    return (
      <div className={styles.content}>
        <PageStatusBanner
          intent="error"
          title="运行时初始化失败"
          description="请检查登录态、菜单空间接口和导航接口。"
        />
      </div>
    )
  }

  const shellMotionStyle = {
    '--shell-sidebar-width': navCollapsed ? '72px' : `${navExpandedWidth}px`,
    '--shell-sidebar-padding-x': navCollapsed ? '8px' : '10px',
  } as CSSProperties

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
    const fallbackPath = openTabs[tabIndex + 1]?.path || openTabs[tabIndex - 1]?.path || appConfig.defaultRoute

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
    const activeWillSurvive = activeIndex < 0 || activeIndex >= targetIndex || currentTab?.pinned

    closeTabsToLeft(path)

    if (!activeWillSurvive) {
      navigate(path, { replace: true })
    }
  }

  function handleCloseTabsToRight(path: string) {
    const targetIndex = openTabs.findIndex((item) => item.path === path)
    const activeIndex = openTabs.findIndex((item) => item.path === location.pathname)
    const currentTab = openTabs[activeIndex]
    const activeWillSurvive = activeIndex < 0 || activeIndex <= targetIndex || currentTab?.pinned

    closeTabsToRight(path)

    if (!activeWillSurvive) {
      navigate(path, { replace: true })
    }
  }

  function handleSignOut() {
    clearTabs()
    clearAuth()
    navigate('/', { replace: true })
  }

  function handleSelectSpace(spaceKey: string) {
    if (spaceKey === currentSpaceKey) {
      return
    }

    const currentPathIsLocal = Boolean(getLocalRouteDefinitionByPath(location.pathname))
    const targetFallback =
      spacesQuery.data?.find((item) => item.key === spaceKey)?.defaultLandingRoute || appConfig.defaultRoute
    setCurrentSpaceKey(spaceKey)
    void invalidateSpaceScopedQueries(spaceKey)

    if (!currentPathIsLocal) {
      navigate(targetFallback, { replace: true })
    }
  }

  return (
    <div className={styles.app}>
      <HeaderBar
        navigationItems={navigationItemsQuery.items}
        currentUser={shellUser}
        darkMode={themeMode === 'dark'}
        navCollapsed={navCollapsed}
        tabsEnabled={tabsEnabled}
        currentSpace={runtimeEnabled ? currentSpace : null}
        spaces={runtimeEnabled ? spacesQuery.data || [] : []}
        onToggleTheme={toggleTheme}
        onToggleNav={toggleNavCollapsed}
        onSetTabsEnabled={setTabsEnabled}
        onOpenMobileNav={() => setMobileNavOpen(true)}
        onSelectSpace={handleSelectSpace}
        onSignOut={handleSignOut}
      />

      <div className={styles.body} style={shellMotionStyle}>
        <aside className={styles.sidebar} style={shellMotionStyle}>
          <SideNav
            items={navigationItemsQuery.items}
            collapsed={navRenderCollapsed}
            contentVisible={navContentVisible}
            mobileOpen={mobileNavOpen}
            currentSpaceKey={currentSpaceKey}
            onExpandedWidthChange={setNavExpandedWidth}
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
            {currentUser ? (
              <Outlet />
            ) : (
              <PageStatusBanner
                intent="info"
                title="页面已清空"
                description="当前已删除所有页面实现，只保留壳层与导航骨架，便于重新重构。"
              />
            )}
          </div>
        </main>
      </div>
    </div>
  )
}
