import type { JSX } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Navigate, matchPath, useLocation } from 'react-router-dom'
import { Spinner } from '@fluentui/react-components'
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
import { PageStatusBanner } from '@/shared/ui/PageStatusBanner'
import { useShellStore } from '@/features/shell/store/useShellStore'
import { appConfig } from '@/shared/config/app-config'

type LocalRouteEntry = RouteDefinition & { element: JSX.Element }

const localRoutes: LocalRouteEntry[] = []

const groupLabelMap: Record<NavigationGroupKey, string> = {
  welcome: '首页',
  workspace: '工作台',
  team: '团队协作',
  message: '消息中心',
  system: '系统管理',
}

export function getNavigationGroupLabel(group: NavigationGroupKey) {
  return groupLabelMap[group]
}

const groupPathLabelMap: Record<string, string> = {
  '/workspace': getNavigationGroupLabel('workspace'),
  '/team': getNavigationGroupLabel('team'),
  '/message': getNavigationGroupLabel('message'),
  '/system': getNavigationGroupLabel('system'),
}

const runtimeRoutePathAliasMap: Record<string, string> = {
  Dashboard: '/dashboard/console',
  Console: '/dashboard/console',
  PageManagement: '/system/page',
}

const runtimeDirectoryComponentSet = new Set(['/index/index'])

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

function joinRuntimePath(parentPath: string | undefined, currentPath: string) {
  const normalizedCurrentPath = currentPath.trim().replace(/^#/, '')
  if (!normalizedCurrentPath) {
    return parentPath || '/'
  }
  if (normalizedCurrentPath.startsWith('/')) {
    return normalizedCurrentPath
  }
  if (!parentPath) {
    return `/${normalizedCurrentPath}`
  }
  return `${parentPath.replace(/\/+$/, '')}/${normalizedCurrentPath.replace(/^\/+/, '')}`
}

function normalizeComponentRoute(component: string) {
  const normalizedComponent = component.trim().replace(/^#/, '')
  if (!normalizedComponent.startsWith('/')) {
    return ''
  }

  return normalizedComponent
}

function resolveRuntimePath(item: RuntimeNavItem, parentPath?: string) {
  const componentRoute = normalizeComponentRoute(item.component)
  if (
    componentRoute &&
    !runtimeDirectoryComponentSet.has(componentRoute) &&
    isImplementedPath(componentRoute)
  ) {
    return componentRoute
  }

  const aliasedPath = runtimeRoutePathAliasMap[item.routeId]
  if (aliasedPath) {
    return aliasedPath
  }
  return joinRuntimePath(parentPath, item.path)
}

function resolveRuntimeTree(items: RuntimeNavItem[], parentPath?: string): RuntimeNavItem[] {
  return items.map((item) => {
    const resolvedPath = resolveRuntimePath(item, parentPath)
    return {
      ...item,
      path: resolvedPath,
      children: item.children?.length ? resolveRuntimeTree(item.children, resolvedPath) : undefined,
    }
  })
}

function pruneSelfDuplicateChildren(items: RuntimeNavItem[]): RuntimeNavItem[] {
  return items.map((item) => {
    const children = item.children?.length ? pruneSelfDuplicateChildren(item.children) : undefined
    const nextChildren = children?.filter(
      (child) => child.routeId !== item.routeId && child.path !== item.path && child.label !== item.label,
    )

    return {
      ...item,
      children: nextChildren?.length ? nextChildren : undefined,
    }
  })
}

function pruneSiblingDuplicates(items: RuntimeNavItem[]): RuntimeNavItem[] {
  const seen = new Set<string>()

  return items.flatMap((item) => {
    const dedupKey = `${item.path}::${item.label}`.trim()
    if (seen.has(dedupKey)) {
      return []
    }
    seen.add(dedupKey)

    return [
      {
        ...item,
        children: item.children?.length ? pruneSiblingDuplicates(item.children) : undefined,
      },
    ]
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

function getPreferredRouteTitle(pathname: string, fallbackTitle: string) {
  const localRoute = getLocalRouteDefinitionByPath(pathname)
  if (localRoute?.shellTitle) {
    return localRoute.shellTitle
  }
  if (`${fallbackTitle || ''}`.trim()) {
    return fallbackTitle
  }
  const groupLabel = groupPathLabelMap[pathname]
  if (groupLabel) {
    return groupLabel
  }
  return fallbackTitle
}

export function buildNavigationItems(menuTree: RuntimeNavItem[]): NavigationItem[] {
  const resolvedTree = pruneSiblingDuplicates(pruneSelfDuplicateChildren(resolveRuntimeTree(menuTree)))

  return resolvedTree
    .filter((item) => !item.hidden)
    .map((item) => ({
      id: item.id,
      routeId: item.routeId,
      path: item.path,
      label: getPreferredRouteTitle(item.path, item.title || item.label),
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
  const runtimeEntries = flattenRuntimeItems(resolveRuntimeTree(manifest?.menuTree || []))
  const matchedRuntimeItem = runtimeEntries.find((entry) =>
    matchPath({ path: entry.item.path, end: true }, pathname),
  )
  const matchedManagedPage = manifest?.managedPages.find((item) =>
    matchPath({ path: item.routePath, end: true }, pathname),
  )

  if (localRoute) {
    const runtimeTitle = matchedRuntimeItem?.item.title || matchedManagedPage?.name || localRoute.shellTitle
    const trail = matchedRuntimeItem?.trail || []
    return {
      routeId: localRoute.id,
      path: pathname,
      title: getPreferredRouteTitle(pathname, runtimeTitle),
      subtitle: localRoute.subtitle,
      group: localRoute.group,
      groupLabel: getNavigationGroupLabel(localRoute.group),
      status: 'implemented',
      breadcrumbs:
        trail.length > 0
          ? trail.map((item, index) => ({
              label: getPreferredRouteTitle(item.path, item.title),
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
      groupLabel: getNavigationGroupLabel(matchedRuntimeItem.item.group),
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
      groupLabel: getNavigationGroupLabel(group),
      status: 'placeholder',
      breadcrumbs: [
        {
          label: getNavigationGroupLabel(group),
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
    queryKey: queryKeys.navigation.runtime(currentSpaceKey),
    queryFn: () => fetchRuntimeNavigationManifest(currentSpaceKey),
    enabled: Boolean(currentSpaceKey),
    placeholderData: (previousData) => previousData,
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
      <PageStatusBanner
        intent="error"
        title="运行时导航加载失败"
        description="请检查后端服务、登录状态和菜单空间接口。"
      />
    )
  }

  const localRoute = getLocalRouteDefinitionByPath(pathname)
  if (localRoute) {
    return localRoute.element
  }

  const runtimeContext = buildRouteContext(pathname, manifestQuery.data)
  if (runtimeContext) {
    return (
      <PageStatusBanner
        intent="info"
        title={runtimeContext.title}
        description={runtimeContext.subtitle || '当前只保留壳层，页面实现已清空，后续将重新重构。'}
      />
    )
  }

  return (
    <PageStatusBanner
      intent="warning"
      title="页面已删除"
      description="当前项目仅保留壳层，请重新接入页面实现。"
    />
  )
}
