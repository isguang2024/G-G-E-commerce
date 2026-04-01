import { useQuery } from '@tanstack/react-query'
import { navigationMock } from '@/shared/mocks/navigation.mock'
import { pageMetaMock } from '@/shared/mocks/page-meta.mock'
import { spacesMock } from '@/shared/mocks/spaces.mock'
import { withDelay } from '@/shared/lib/delay'

async function fetchSpaces() {
  return withDelay(spacesMock)
}

async function fetchNavigationTree() {
  return withDelay(navigationMock)
}

async function fetchPageMeta(routeId: string) {
  const pageMeta = pageMetaMock[routeId]
  if (!pageMeta) {
    throw new Error(`Missing page metadata for route "${routeId}"`)
  }

  return withDelay(pageMeta, 80)
}

export function useSpacesQuery() {
  return useQuery({
    queryKey: ['shell', 'spaces'],
    queryFn: fetchSpaces,
  })
}

export function useNavigationTreeQuery() {
  return useQuery({
    queryKey: ['navigation', 'tree'],
    queryFn: fetchNavigationTree,
  })
}

export function usePageMetaQuery(routeId: string) {
  return useQuery({
    queryKey: ['pages', 'meta', routeId],
    queryFn: () => fetchPageMeta(routeId),
  })
}
