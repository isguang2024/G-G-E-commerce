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

type QueryToggleOptions = {
  enabled?: boolean
}

export function useMenuSpacesQuery(options?: QueryToggleOptions) {
  return useQuery({
    queryKey: queryKeys.navigation.menuSpaces,
    queryFn: fetchMenuSpaces,
    placeholderData: (previousData) => previousData,
    enabled: options?.enabled ?? true,
  })
}

export function useRuntimeNavigationManifestQuery(spaceKey?: string, options?: QueryToggleOptions) {
  const fallbackSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const resolvedSpaceKey = spaceKey || fallbackSpaceKey

  return useQuery({
    queryKey: queryKeys.navigation.runtime(resolvedSpaceKey),
    queryFn: () => fetchRuntimeNavigationManifest(resolvedSpaceKey),
    enabled: Boolean(resolvedSpaceKey) && (options?.enabled ?? true),
    placeholderData: (previousData) => previousData,
  })
}

export function useNavigationItems(options?: QueryToggleOptions) {
  const currentSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const manifestQuery = useRuntimeNavigationManifestQuery(currentSpaceKey, options)

  const items = useMemo(
    () => buildNavigationItems(manifestQuery.data?.menuTree || []),
    [manifestQuery.data?.menuTree],
  )

  return {
    ...manifestQuery,
    items,
  }
}

export function useRouteContext(routeId?: string, options?: QueryToggleOptions) {
  const location = useLocation()
  const currentSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const manifestQuery = useRuntimeNavigationManifestQuery(currentSpaceKey, options)

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
