import { computed } from 'vue'
import { useTenantStore } from '@/store/modules/tenant'
import { useWorkspaceStore } from '@/store/modules/workspace'

export function useMessageWorkspace(scope: 'platform' | 'collaboration' | 'team') {
  const collaborationWorkspaceStore = useTenantStore()
  const workspaceStore = useWorkspaceStore()

  const isTeamScope = computed(() => scope === 'collaboration' || scope === 'team')
  const skipTenantHeader = computed(() => !isTeamScope.value)
  const currentCollaborationWorkspaceId = computed(() => collaborationWorkspaceStore.currentCollaborationWorkspaceId || '')
  const currentTeamName = computed(
    () =>
      collaborationWorkspaceStore.currentTeam?.name ||
      workspaceStore.currentAuthWorkspace?.name ||
      '当前协作空间'
  )
  const currentWorkspaceName = computed(
    () => workspaceStore.currentAuthWorkspace?.name || '当前授权工作空间'
  )
  const currentWorkspaceLabel = computed(() =>
    workspaceStore.currentAuthWorkspaceType === 'collaboration' ? '协作空间' : '个人工作空间'
  )

  const ensureTeamContext = () => {
    if (!isTeamScope.value) return
    if (collaborationWorkspaceStore.currentCollaborationWorkspaceId) return
    const fallbackCollaborationWorkspaceId = collaborationWorkspaceStore.teamList[0]?.id || ''
    if (fallbackCollaborationWorkspaceId) {
      collaborationWorkspaceStore.enterTeamContext(fallbackCollaborationWorkspaceId)
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
      return target
        .replace(/<[^>]+>/g, ' ')
        .replace(/&nbsp;/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
    }
    const parser = new DOMParser()
    const doc = parser.parseFromString(target, 'text/html')
    return (doc.body.textContent || '').replace(/\s+/g, ' ').trim()
  }

  return {
    collaborationWorkspaceStore,
    workspaceStore,
    isTeamScope,
    skipTenantHeader,
    currentCollaborationWorkspaceId,
    currentTeamName,
    currentWorkspaceName,
    currentWorkspaceLabel,
    ensureTeamContext,
    formatTime,
    plainTextFromHtml
  }
}

