import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '@/features/auth/auth.store'
import { useMenuSpacesQuery, useRuntimeNavigationManifestQuery } from '@/features/navigation/navigation.service'
import { fetchFastEnterConfig } from '@/shared/api/modules/system.api'
import { fetchInboxSummary } from '@/shared/api/modules/message.api'
import { queryKeys } from '@/shared/api/query-keys'
import type { DashboardSummary, MetricCard, UserCenterProfile } from '@/shared/types/admin'

export function useFastEnterConfigQuery() {
  return useQuery({
    queryKey: queryKeys.system.fastEnter,
    queryFn: fetchFastEnterConfig,
    placeholderData: (previousData) => previousData,
  })
}

export function useDashboardSummary() {
  const currentUser = useAuthStore((state) => state.currentUser)
  const currentSpaceKey = useAuthStore((state) => state.tenantContext.currentTenantId)
  const menuSpacesQuery = useMenuSpacesQuery()
  const manifestQuery = useRuntimeNavigationManifestQuery()
  const fastEnterQuery = useFastEnterConfigQuery()
  const inboxSummaryQuery = useQuery({
    queryKey: queryKeys.inbox.summary,
    queryFn: fetchInboxSummary,
    placeholderData: (previousData) => previousData,
  })

  const summary: DashboardSummary = {
    currentUserName: currentUser?.displayName || currentUser?.username || '未登录用户',
    currentSpaceLabel:
      menuSpacesQuery.data?.find((item) => item.key === manifestQuery.data?.currentSpace?.space.key)?.label ||
      manifestQuery.data?.currentSpace?.space.label ||
      currentSpaceKey ||
      '默认空间',
    visibleMenuCount: manifestQuery.data?.context.visibleMenuCount || 0,
    managedPageCount: manifestQuery.data?.context.managedPageCount || 0,
    unreadInboxCount: inboxSummaryQuery.data?.unreadTotal || 0,
    fastEntryCount: fastEnterQuery.data?.applications.filter((item) => item.enabled).length || 0,
    quickLinkCount: fastEnterQuery.data?.quickLinks.filter((item) => item.enabled).length || 0,
  }

  const metrics: MetricCard[] = [
    {
      id: 'visible-menu',
      label: '可见导航',
      value: `${summary.visibleMenuCount}`,
      hint: '来自运行时导航',
      tone: 'brand',
    },
    {
      id: 'managed-page',
      label: '受管页面',
      value: `${summary.managedPageCount}`,
      hint: '来自运行时页面注册表',
      tone: 'neutral',
    },
    {
      id: 'inbox',
      label: '未读消息',
      value: `${summary.unreadInboxCount}`,
      hint: '来自收件箱摘要',
      tone: summary.unreadInboxCount > 0 ? 'warning' : 'success',
    },
    {
      id: 'fast-entry',
      label: '快捷入口',
      value: `${summary.fastEntryCount}`,
      hint: `${summary.quickLinkCount} 条快速链接`,
      tone: 'neutral',
    },
  ]

  return {
    summary,
    metrics,
    fastEnterQuery,
    inboxSummaryQuery,
    manifestQuery,
    menuSpacesQuery,
    isLoading:
      fastEnterQuery.isLoading ||
      inboxSummaryQuery.isLoading ||
      manifestQuery.isLoading ||
      menuSpacesQuery.isLoading,
    isError:
      fastEnterQuery.isError ||
      inboxSummaryQuery.isError ||
      manifestQuery.isError ||
      menuSpacesQuery.isError,
    error:
      fastEnterQuery.error ||
      inboxSummaryQuery.error ||
      manifestQuery.error ||
      menuSpacesQuery.error,
  }
}

export function useUserCenterProfile(): UserCenterProfile | null {
  const currentUser = useAuthStore((state) => state.currentUser)
  if (!currentUser) {
    return null
  }

  return {
    id: currentUser.id,
    userName: currentUser.username,
    displayName: currentUser.displayName,
    email: currentUser.email,
    phone: currentUser.phone,
    avatarUrl: currentUser.avatarUrl,
    status: currentUser.status,
    badges: currentUser.badges,
    currentTenantId: currentUser.currentTenantId,
  }
}
