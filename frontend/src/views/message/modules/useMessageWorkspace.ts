import { computed } from 'vue'
import { useCollaborationWorkspaceStore } from '@/store/modules/collaboration-workspace'
import { useWorkspaceStore } from '@/store/modules/workspace'

export function useMessageWorkspace(scope: 'platform' | 'collaboration') {
  const collaborationWorkspaceStore = useCollaborationWorkspaceStore()
  const workspaceStore = useWorkspaceStore()

  const isCollaborationScope = computed(() => scope === 'collaboration')
  const skipCollaborationWorkspaceHeader = computed(() => !isCollaborationScope.value)
  const currentCollaborationWorkspaceId = computed(
    () => collaborationWorkspaceStore.currentCollaborationWorkspaceId || ''
  )
  const currentCollaborationWorkspaceName = computed(
    () =>
      collaborationWorkspaceStore.currentCollaborationWorkspace?.name ||
      workspaceStore.currentAuthWorkspace?.name ||
      '当前协作空间'
  )
  const currentWorkspaceName = computed(
    () => workspaceStore.currentAuthWorkspace?.name || '当前授权工作空间'
  )
  const currentWorkspaceLabel = computed(() =>
    workspaceStore.currentAuthWorkspaceType === 'collaboration' ? '协作空间' : '个人工作空间'
  )

  const ensureCollaborationWorkspaceContext = () => {
    if (!isCollaborationScope.value) return
    if (collaborationWorkspaceStore.currentCollaborationWorkspaceId) return
    const fallbackCollaborationWorkspaceId =
      collaborationWorkspaceStore.collaborationWorkspaceList[0]?.id || ''
    if (fallbackCollaborationWorkspaceId) {
      collaborationWorkspaceStore.enterCollaborationContext(fallbackCollaborationWorkspaceId)
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
    isCollaborationScope,
    skipCollaborationWorkspaceHeader,
    currentCollaborationWorkspaceId,
    currentCollaborationWorkspaceName,
    currentWorkspaceName,
    currentWorkspaceLabel,
    ensureCollaborationWorkspaceContext,
    formatTime,
    plainTextFromHtml
  }
}
