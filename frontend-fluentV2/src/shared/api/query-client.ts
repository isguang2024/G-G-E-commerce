import { QueryClient } from '@tanstack/react-query'
import { queryKeys } from '@/shared/api/query-keys'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      gcTime: 5 * 60_000,
      retry: false,
      refetchOnWindowFocus: false,
    },
  },
})

export async function invalidateSpaceScopedQueries(spaceKey: string) {
  await Promise.all([
    queryClient.invalidateQueries({
      queryKey: queryKeys.navigation.runtime(spaceKey),
    }),
    queryClient.invalidateQueries({
      queryKey: queryKeys.page.list('', spaceKey).slice(0, 3),
    }),
    queryClient.invalidateQueries({
      queryKey: ['menu', 'tree', spaceKey],
    }),
    queryClient.invalidateQueries({
      queryKey: ['menu', 'detail'],
    }),
    queryClient.invalidateQueries({
      queryKey: ['menu', 'pages'],
    }),
    queryClient.invalidateQueries({
      queryKey: queryKeys.menu.runtimePages(spaceKey),
    }),
    queryClient.invalidateQueries({
      queryKey: queryKeys.menu.manageGroups(spaceKey),
    }),
    queryClient.invalidateQueries({
      queryKey: queryKeys.system.menuSpaces,
    }),
  ])
}
