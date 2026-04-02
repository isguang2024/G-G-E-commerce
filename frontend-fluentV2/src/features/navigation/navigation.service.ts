import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useLocation } from 'react-router-dom'
import { fetchMenuSpaces, fetchRuntimeNavigationManifest } from '@/shared/api/modules/navigation.api'
import { queryKeys } from '@/shared/api/query-keys'
import { useShellStore } from '@/features/shell/store/useShellStore'
import {
  buildNavigationItems,
  buildRouteContext,
  getLocalRouteDefinitionById,
  getLocalRouteDefinitionByPath,
} from '@/features/navigation/route-registry'

export function useMenuSpacesQuery() {
  return useQuery({
    queryKey: queryKeys.navigation.spaces,
    queryFn: fetchMenuSpaces,
  })
}

export function useRuntimeNavigationManifestQuery(spaceKey?: string) {
  const fallbackSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const resolvedSpaceKey = spaceKey || fallbackSpaceKey

  return useQuery({
    queryKey: queryKeys.navigation.manifest(resolvedSpaceKey),
    queryFn: () => fetchRuntimeNavigationManifest(resolvedSpaceKey),
    enabled: Boolean(resolvedSpaceKey),
  })
}

export function useNavigationItems() {
  const currentSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const manifestQuery = useRuntimeNavigationManifestQuery(currentSpaceKey)

  const items = useMemo(
    () => buildNavigationItems(manifestQuery.data?.menuTree || []),
    [manifestQuery.data?.menuTree],
  )

  return {
    ...manifestQuery,
    items,
  }
}

export function useRouteContext(routeId?: string) {
  const location = useLocation()
  const currentSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const manifestQuery = useRuntimeNavigationManifestQuery(currentSpaceKey)

  const context = useMemo(() => {
    const pathname = location.pathname
    const fallbackRoute =
      (routeId ? getLocalRouteDefinitionById(routeId) : undefined) ||
      getLocalRouteDefinitionByPath(pathname)

    return buildRouteContext(pathname, manifestQuery.data, fallbackRoute)
  }, [location.pathname, manifestQuery.data, routeId])

  return {
    ...manifestQuery,
    context,
  }
}
