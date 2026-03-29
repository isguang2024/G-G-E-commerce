import { computed } from 'vue'
import { useTenantStore } from '@/store/modules/tenant'

export function useMessageWorkspace(scope: 'platform' | 'team') {
  const tenantStore = useTenantStore()

  const isTeamScope = computed(() => scope === 'team')
  const skipTenantHeader = computed(() => !isTeamScope.value)
  const currentTeamName = computed(() => tenantStore.currentTeam?.name || '当前团队')

  const ensureTeamContext = () => {
    if (!isTeamScope.value) return
    if (tenantStore.currentTenantId) return
    const fallbackTeamId = tenantStore.teamList[0]?.id || ''
    if (fallbackTeamId) {
      tenantStore.enterTeamContext(fallbackTeamId)
    }
  }

  const formatTime = (value?: string, fallback = '刚刚更新') => {
    if (!value) return fallback
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value
    return new Intl.DateTimeFormat('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date)
  }

  const plainTextFromHtml = (value?: string) => {
    const target = `${value || ''}`.trim()
    if (!target) return ''
    if (typeof window === 'undefined') {
      return target.replace(/<[^>]+>/g, ' ').replace(/&nbsp;/g, ' ').replace(/\s+/g, ' ').trim()
    }
    const parser = new DOMParser()
    const doc = parser.parseFromString(target, 'text/html')
    return (doc.body.textContent || '').replace(/\s+/g, ' ').trim()
  }

  return {
    tenantStore,
    isTeamScope,
    skipTenantHeader,
    currentTeamName,
    ensureTeamContext,
    formatTime,
    plainTextFromHtml
  }
}
