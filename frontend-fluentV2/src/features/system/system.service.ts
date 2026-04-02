import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  cleanupStaleApiEndpoints,
  createApiCategory,
  createApiEndpoint,
  createPage,
  deletePage,
  fetchApiCategories,
  fetchApiEndpointList,
  fetchApiEndpointOverview,
  fetchFastEnterConfig,
  fetchMenuSpaceHostBindings,
  fetchMenuSpaceList,
  fetchMenuSpaceMode,
  fetchPageAccessTrace,
  fetchPageBreadcrumbPreview,
  fetchPageDetail,
  fetchPageList,
  fetchPageMenuOptions,
  fetchPageUnregisteredList,
  fetchUnregisteredApiRoutes,
  fetchUnregisteredApiScanConfig,
  initializeMenuSpaceFromDefault,
  saveMenuSpace,
  saveMenuSpaceHostBinding,
  saveUnregisteredApiScanConfig,
  syncApiEndpoints,
  syncPages,
  updateApiCategory,
  updateApiEndpoint,
  updateApiEndpointContextScope,
  updateFastEnterConfig,
  updateMenuSpaceMode,
  updatePage,
} from '@/shared/api/modules/system.api'
import { queryKeys } from '@/shared/api/query-keys'
import type {
  AccessTraceFilter,
  ApiEndpointCategory,
  ApiEndpointSavePayload,
  PageSavePayload,
} from '@/shared/types/admin'

function serializeFilters(filters?: unknown) {
  return JSON.stringify(filters || {})
}

export function usePageListQuery(filters?: Record<string, unknown>, spaceKey = 'default') {
  return useQuery({
    queryKey: queryKeys.page.list(serializeFilters(filters), spaceKey),
    queryFn: () => fetchPageList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function usePageDetailQuery(pageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.page.detail(pageId || ''),
    queryFn: () => fetchPageDetail(pageId!),
    enabled: Boolean(pageId),
    placeholderData: (previousData) => previousData,
  })
}

export function usePageMenuOptionsQuery(spaceKey?: string) {
  return useQuery({
    queryKey: queryKeys.page.menuOptions(spaceKey || ''),
    queryFn: () => fetchPageMenuOptions(spaceKey),
    placeholderData: (previousData) => previousData,
  })
}

export function usePageUnregisteredQuery() {
  return useQuery({
    queryKey: queryKeys.page.unregistered,
    queryFn: fetchPageUnregisteredList,
    placeholderData: (previousData) => previousData,
  })
}

export function usePageBreadcrumbPreviewQuery(pageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.page.breadcrumbPreview(pageId || ''),
    queryFn: () => fetchPageBreadcrumbPreview(pageId!),
    enabled: Boolean(pageId),
    placeholderData: (previousData) => previousData,
  })
}

export function usePageAccessTraceQuery(filters?: AccessTraceFilter) {
  return useQuery({
    queryKey: queryKeys.page.accessTrace(serializeFilters(filters)),
    queryFn: () => fetchPageAccessTrace(filters ?? {}),
    enabled: Boolean(filters && Object.keys(filters).length),
    placeholderData: (previousData) => previousData,
  })
}

export function useCreatePageMutation(spaceKey = 'default') {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createPage,
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.page.list('', spaceKey).slice(0, 3) }),
        client.invalidateQueries({ queryKey: queryKeys.page.unregistered }),
      ])
    },
  })
}

export function useUpdatePageMutation(pageId: string, spaceKey = 'default') {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: PageSavePayload) => updatePage(pageId, payload),
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.page.list('', spaceKey).slice(0, 3) }),
        client.invalidateQueries({ queryKey: queryKeys.page.detail(pageId) }),
        client.invalidateQueries({ queryKey: queryKeys.page.breadcrumbPreview(pageId) }),
      ])
    },
  })
}

export function useDeletePageMutation(spaceKey = 'default') {
  const client = useQueryClient()
  return useMutation({
    mutationFn: deletePage,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.page.list('', spaceKey).slice(0, 3) })
    },
  })
}

export function useSyncPagesMutation(spaceKey = 'default') {
  const client = useQueryClient()
  return useMutation({
    mutationFn: syncPages,
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.page.list('', spaceKey).slice(0, 3) }),
        client.invalidateQueries({ queryKey: queryKeys.page.unregistered }),
      ])
    },
  })
}

export function useApiEndpointListQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.apiEndpoint.list(serializeFilters(filters)),
    queryFn: () => fetchApiEndpointList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useApiEndpointOverviewQuery() {
  return useQuery({
    queryKey: queryKeys.apiEndpoint.overview,
    queryFn: fetchApiEndpointOverview,
    placeholderData: (previousData) => previousData,
  })
}

export function useApiCategoriesQuery() {
  return useQuery({
    queryKey: queryKeys.apiEndpoint.categories,
    queryFn: fetchApiCategories,
    placeholderData: (previousData) => previousData,
  })
}

export function useUnregisteredApiRoutesQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.apiEndpoint.unregistered(serializeFilters(filters)),
    queryFn: () => fetchUnregisteredApiRoutes(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useApiScanConfigQuery() {
  return useQuery({
    queryKey: queryKeys.apiEndpoint.scanConfig,
    queryFn: fetchUnregisteredApiScanConfig,
    placeholderData: (previousData) => previousData,
  })
}

export function useCreateApiEndpointMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createApiEndpoint,
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ['apiEndpoint', 'list'] }),
        client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.overview }),
      ])
    },
  })
}

export function useUpdateApiEndpointMutation(endpointId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: ApiEndpointSavePayload) => updateApiEndpoint(endpointId, payload),
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ['apiEndpoint', 'list'] }),
        client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.overview }),
      ])
    },
  })
}

export function useUpdateApiEndpointContextScopeMutation(endpointId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (contextScope: string) => updateApiEndpointContextScope(endpointId, contextScope),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['apiEndpoint', 'list'] })
    },
  })
}

export function useCreateApiCategoryMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createApiCategory,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.categories })
    },
  })
}

export function useUpdateApiCategoryMutation(categoryId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: Partial<ApiEndpointCategory>) => updateApiCategory(categoryId, payload),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.categories })
    },
  })
}

export function useSyncApiEndpointsMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: syncApiEndpoints,
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ['apiEndpoint', 'list'] }),
        client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.overview }),
        client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.unregistered('') }),
      ])
    },
  })
}

export function useCleanupStaleApiEndpointsMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: cleanupStaleApiEndpoints,
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ['apiEndpoint', 'list'] }),
        client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.overview }),
      ])
    },
  })
}

export function useSaveApiScanConfigMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: saveUnregisteredApiScanConfig,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.apiEndpoint.scanConfig })
    },
  })
}

export function useFastEnterConfigQuery() {
  return useQuery({
    queryKey: queryKeys.system.fastEnter,
    queryFn: fetchFastEnterConfig,
    placeholderData: (previousData) => previousData,
  })
}

export function useUpdateFastEnterConfigMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: updateFastEnterConfig,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.system.fastEnter })
    },
  })
}

export function useMenuSpaceModeQuery() {
  return useQuery({
    queryKey: queryKeys.system.menuSpaceMode,
    queryFn: fetchMenuSpaceMode,
    placeholderData: (previousData) => previousData,
  })
}

export function useUpdateMenuSpaceModeMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: updateMenuSpaceMode,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.system.menuSpaceMode })
    },
  })
}

export function useMenuSpaceListQuery() {
  return useQuery({
    queryKey: queryKeys.system.menuSpaces,
    queryFn: fetchMenuSpaceList,
    placeholderData: (previousData) => previousData,
  })
}

export function useSaveMenuSpaceMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: saveMenuSpace,
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.system.menuSpaces }),
        client.invalidateQueries({ queryKey: queryKeys.navigation.menuSpaces }),
      ])
    },
  })
}

export function useInitializeMenuSpaceMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ spaceKey, force }: { spaceKey: string; force?: boolean }) =>
      initializeMenuSpaceFromDefault(spaceKey, force),
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.system.menuSpaces }),
        client.invalidateQueries({ queryKey: queryKeys.navigation.menuSpaces }),
      ])
    },
  })
}

export function useMenuSpaceHostBindingsQuery() {
  return useQuery({
    queryKey: queryKeys.system.menuSpaceHostBindings,
    queryFn: fetchMenuSpaceHostBindings,
    placeholderData: (previousData) => previousData,
  })
}

export function useSaveMenuSpaceHostBindingMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: saveMenuSpaceHostBinding,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.system.menuSpaceHostBindings })
    },
  })
}
