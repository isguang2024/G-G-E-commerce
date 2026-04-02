import type { JSX } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Navigate, matchPath, useLocation } from 'react-router-dom'
import { MessageBar, MessageBarBody, Spinner } from '@fluentui/react-components'
import { fetchRuntimeNavigationManifest } from '@/shared/api/modules/navigation.api'
import { queryKeys } from '@/shared/api/query-keys'
import type {
  NavigationGroupKey,
  NavigationItem,
  RouteContext,
  RouteDefinition,
  RuntimeNavItem,
  RuntimeNavigationManifest,
} from '@/shared/types/navigation'
import { MigrationPlaceholderPage } from '@/pages/placeholder/MigrationPlaceholderPage'
import { NotFoundPage } from '@/pages/not-found/NotFoundPage'
import { SystemHomePage } from '@/pages/system/SystemHomePage'
import { SystemMenuPage } from '@/pages/system/SystemMenuPage'
import { WelcomePage } from '@/pages/welcome/WelcomePage'
import { WorkspaceHomePage } from '@/pages/workspace/WorkspaceHomePage'
import { useShellStore } from '@/features/shell/store/useShellStore'
import { appConfig } from '@/shared/config/app-config'

type LocalRouteEntry = RouteDefinition & {
  element: JSX.Element
}

const localRoutes: LocalRouteEntry[] = [
  {
    id: 'welcome',
    path: '/welcome',
    group: 'welcome',
    status: 'implemented',
    shellTitle: '首页',
    subtitle: '真实登录、导航和空间切换已经接入，后续迁移从这里开始承接。',
    element: <WelcomePage routeId="welcome" />,
  },
  {
    id: 'workspace-home',
    path: '/workspace',
    group: 'workspace',
    status: 'implemented',
    shellTitle: '工作台',
    subtitle: '当前保留为基础工作台入口，继续承接真实运行时上下文。',
    element: <WorkspaceHomePage routeId="workspace-home" />,
  },
  {
    id: 'system-home',
    path: '/system',
    group: 'system',
    status: 'implemented',
    shellTitle: '系统管理',
    subtitle: '系统管理首页已经纳入真实认证与导航链路。',
    element: <SystemHomePage routeId="system-home" />,
  },
  {
    id: 'system-menu',
    path: '/system/menu',
    group: 'system',
    status: 'implemented',
    shellTitle: '菜单管理',
    subtitle: '当前版本提供真实菜单树浏览、详情只读和空间联动。',
    element: <SystemMenuPage routeId="system-menu" />,
  },
]

const groupLabelMap: Record<NavigationGroupKey, string> = {
  welcome: '首页',
  workspace: '工作台',
  team: '团队协作',
  message: '消息中心',
  system: '系统管理',
}

function inferGroup(pathname: string): NavigationGroupKey {
  if (pathname.startsWith('/workspace')) return 'workspace'
  if (pathname.startsWith('/team')) return 'team'
  if (pathname.startsWith('/message')) return 'message'
  if (pathname.startsWith('/system')) return 'system'
  return 'welcome'
}

function flattenRuntimeItems(items: RuntimeNavItem[], trail: RuntimeNavItem[] = []): Array<{
  item: RuntimeNavItem
  trail: RuntimeNavItem[]
}> {
  return items.flatMap((item) => {
    const nextTrail = [...trail, item]
    const current = [{ item, trail: nextTrail }]
    return item.children?.length ? [...current, ...flattenRuntimeItems(item.children, nextTrail)] : current
  })
}

export function getLocalRouteDefinitionByPath(pathname: string) {
  return localRoutes.find((item) => matchPath({ path: item.path, end: true }, pathname))
}

export function getLocalRouteDefinitionById(routeId: string) {
  return localRoutes.find((item) => item.id === routeId)
}

export function isImplementedPath(pathname: string) {
  return Boolean(getLocalRouteDefinitionByPath(pathname))
}

export function buildNavigationItems(menuTree: RuntimeNavItem[]): NavigationItem[] {
  return menuTree
    .filter((item) => !item.hidden)
    .map((item) => ({
      id: item.id,
      routeId: item.routeId,
      path: item.path,
      label: item.label,
      icon: item.icon,
      group: item.group,
      status: isImplementedPath(item.path) ? 'implemented' : 'placeholder',
      spaceKey: item.spaceKey,
      children: item.children?.length ? buildNavigationItems(item.children) : undefined,
    }))
}

export function buildRouteContext(
  pathname: string,
  manifest?: RuntimeNavigationManifest,
  fallbackLocalRoute?: RouteDefinition,
): RouteContext | null {
  const localRoute = getLocalRouteDefinitionByPath(pathname) || fallbackLocalRoute
  const runtimeEntries = flattenRuntimeItems(manifest?.menuTree || [])
  const matchedRuntimeItem = runtimeEntries.find((entry) =>
    matchPath({ path: entry.item.path, end: true }, pathname),
  )
  const matchedManagedPage = manifest?.managedPages.find((item) =>
    matchPath({ path: item.routePath, end: true }, pathname),
  )

  if (localRoute) {
    const runtimeTitle = matchedRuntimeItem?.item.title || matchedManagedPage?.name
    const trail = matchedRuntimeItem?.trail || []
    return {
      routeId: localRoute.id,
      path: pathname,
      title: runtimeTitle || localRoute.shellTitle,
      subtitle: localRoute.subtitle,
      group: localRoute.group,
      groupLabel: groupLabelMap[localRoute.group],
      status: 'implemented',
      breadcrumbs:
        trail.length > 0
          ? trail.map((item, index) => ({
              label: item.title,
              path: index === trail.length - 1 ? undefined : item.path,
            }))
          : [{ label: localRoute.shellTitle }],
      spaceKey:
        manifest?.currentSpace?.space.key ||
        matchedRuntimeItem?.item.spaceKey ||
        matchedManagedPage?.spaceKey,
      source: matchedRuntimeItem ? 'runtime-menu' : matchedManagedPage ? 'runtime-page' : 'local',
      pageKey: matchedManagedPage?.pageKey,
      permissionKey: matchedManagedPage?.permissionKey || matchedRuntimeItem?.item.permissionKey,
      accessMode: matchedManagedPage?.accessMode || matchedRuntimeItem?.item.accessMode,
      manageGroupName: matchedRuntimeItem?.item.manageGroupName,
    }
  }

  if (matchedRuntimeItem) {
    return {
      routeId: matchedRuntimeItem.item.routeId,
      path: pathname,
      title: matchedRuntimeItem.item.title,
      subtitle: '该页面已进入真实运行时导航，但本地 React 页面尚未迁移，当前由统一占位页承接。',
      group: matchedRuntimeItem.item.group,
      groupLabel: groupLabelMap[matchedRuntimeItem.item.group],
      status: 'placeholder',
      breadcrumbs: matchedRuntimeItem.trail.map((item, index) => ({
        label: item.title,
        path: index === matchedRuntimeItem.trail.length - 1 ? undefined : item.path,
      })),
      spaceKey: matchedRuntimeItem.item.spaceKey,
      source: 'runtime-menu',
      permissionKey: matchedRuntimeItem.item.permissionKey,
      accessMode: matchedRuntimeItem.item.accessMode,
      manageGroupName: matchedRuntimeItem.item.manageGroupName,
    }
  }

  if (matchedManagedPage) {
    const group = inferGroup(pathname)
    return {
      routeId: matchedManagedPage.pageKey || pathname,
      path: pathname,
      title: matchedManagedPage.name || matchedManagedPage.pageKey,
      subtitle: '该受管页面已进入后端运行时注册表，但本地实现尚未迁移，当前使用统一占位页展示上下文。',
      group,
      groupLabel: groupLabelMap[group],
      status: 'placeholder',
      breadcrumbs: [
        {
          label: groupLabelMap[group],
        },
        {
          label: matchedManagedPage.name || matchedManagedPage.pageKey,
        },
      ],
      spaceKey: manifest?.currentSpace?.space.key || matchedManagedPage.spaceKey || matchedManagedPage.spaceKeys[0],
      source: 'runtime-page',
      pageKey: matchedManagedPage.pageKey,
      permissionKey: matchedManagedPage.permissionKey,
      accessMode: matchedManagedPage.accessMode,
    }
  }

  return null
}

export function RuntimePageOutlet() {
  const location = useLocation()
  const currentSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const manifestQuery = useQuery({
    queryKey: queryKeys.navigation.manifest(currentSpaceKey),
    queryFn: () => fetchRuntimeNavigationManifest(currentSpaceKey),
    enabled: Boolean(currentSpaceKey),
  })
  const pathname = location.pathname

  if (pathname === '/') {
    const targetPath =
      manifestQuery.data?.currentSpace?.space.defaultLandingRoute || appConfig.defaultRoute
    return <Navigate replace to={targetPath} />
  }

  if (manifestQuery.isLoading) {
    return <Spinner label="正在加载运行时导航" />
  }

  if (manifestQuery.isError) {
    return (
      <MessageBar>
        <MessageBarBody>运行时导航加载失败，请检查后端服务或登录状态。</MessageBarBody>
      </MessageBar>
    )
  }

  const localRoute = getLocalRouteDefinitionByPath(pathname)
  if (localRoute) {
    return localRoute.element
  }

  const runtimeContext = buildRouteContext(pathname, manifestQuery.data)
  if (runtimeContext) {
    return <MigrationPlaceholderPage routeContext={runtimeContext} />
  }

  return <NotFoundPage />
}
