import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetMyCollaborations } from '@/api/collaboration'
import { registerHttpCollaborationContext } from '@/utils/http/request-context'
import { HttpError } from '@/utils/http/error'
import { useWorkspaceStore } from './workspace'

export type WorkspaceContextMode = 'personal' | 'collaboration'
export type AppContextMode = WorkspaceContextMode

export function hasPersonalWorkspaceAccessByUserInfo(
  userInfo?: Partial<Api.Auth.UserInfo> | null
): boolean {
  if (!userInfo) return false
  if (userInfo.is_super_admin) return true
  if (!Array.isArray(userInfo.actions)) return false
  return userInfo.actions.some((item) => {
    const key = `${item || ''}`.trim()
    return (
      key.startsWith('system.') ||
      key.startsWith('feature_package.') ||
      key.startsWith('message.') ||
      key.startsWith('personal.') ||
      key === 'workspace.manage'
    )
  })
}

export const useCollaborationStore = defineStore(
  'collaborationStore',
  () => {
    const workspaceStore = useWorkspaceStore()
    const collaborationList = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
    const loading = ref(false)
    const hasPersonalWorkspaceAccess = ref(false)

    const currentContextMode = computed<WorkspaceContextMode>(() =>
      workspaceStore.currentAuthWorkspaceType === 'collaboration' ? 'collaboration' : 'personal'
    )

    // V5：协作空间 ID 单一语义 = collaborationWorkspaceId（spec 字段映射）。
    // 旧 workspaceId / record.id 兜底已废弃。
    const currentAuthWorkspaceCollaborationId = computed(() => {
      const currentWorkspace = workspaceStore.currentAuthWorkspace
      if (!currentWorkspace) return ''
      return `${currentWorkspace.collaborationWorkspaceId || ''}`.trim()
    })

    const currentCollaborationId = computed(() => {
      if (workspaceStore.currentAuthWorkspaceType !== 'collaboration') return ''
      return currentAuthWorkspaceCollaborationId.value
    })

    const currentCollaboration = computed(() => {
      const currentId = currentCollaborationId.value
      return (
        collaborationList.value.find(
          (item) => item.collaborationWorkspaceId === currentId
        ) || null
      )
    })

    const hasCollaborations = computed(() => collaborationList.value.length > 0)
    const isPersonalWorkspaceContext = computed(() => currentContextMode.value === 'personal')
    const shouldShowSwitcher = computed(() => collaborationList.value.length > 1)

    registerHttpCollaborationContext({
      getCurrentCollaborationId: () => currentCollaborationId.value,
      getCurrentContextMode: () => currentContextMode.value
    })

    const setCurrentCollaborationId = (collaborationWorkspaceId: string) => {
      if (!collaborationWorkspaceId) {
        enterPersonalWorkspaceContext()
        return
      }
      enterCollaborationContext(collaborationWorkspaceId)
    }

    const setCurrentContextMode = (mode: AppContextMode) => {
      if (mode !== 'collaboration') {
        enterPersonalWorkspaceContext()
        return
      }
      if (currentCollaborationId.value) return
      const firstId =
        collaborationList.value[0]?.collaborationWorkspaceId ||
        collaborationList.value[0]?.workspaceId ||
        ''
      if (firstId) enterCollaborationContext(firstId)
    }

    const setPersonalWorkspaceAccess = (enabled: boolean) => {
      hasPersonalWorkspaceAccess.value = enabled
      if (
        !enabled &&
        currentContextMode.value === 'personal' &&
        collaborationList.value.length > 0
      ) {
        const firstId =
          collaborationList.value[0]?.collaborationWorkspaceId ||
          collaborationList.value[0]?.workspaceId ||
          ''
        if (firstId) {
          enterCollaborationContext(firstId)
        }
      }
    }

    const enterPersonalWorkspaceContext = () => {
      workspaceStore.enterPersonalWorkspace()
    }

    const enterCollaborationContext = (collaborationWorkspaceId: string) => {
      const normalizedId = `${collaborationWorkspaceId || ''}`.trim()
      if (!normalizedId) {
        workspaceStore.setCurrentAuthWorkspace('', 'personal')
        return
      }
      const matched = collaborationList.value.find(
        (item) =>
          item.collaborationWorkspaceId === normalizedId ||
          item.workspaceId === normalizedId ||
          item.id === normalizedId
      )
      workspaceStore.enterWorkspaceById(
        matched?.workspaceId || matched?.collaborationWorkspaceId || normalizedId,
        matched?.workspaceType || 'collaboration'
      )
    }

    const setCollaborationList = (
      items: Api.SystemManage.CollaborationWorkspaceListItem[]
    ) => {
      collaborationList.value = items || []
    }

    const ensureCurrentCollaboration = (options?: {
      preferredCollaborationId?: string
      preferredLegacyCollaborationId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferPersonalWorkspace?: boolean
    }) => {
      if (collaborationList.value.length === 0) {
        workspaceStore.ensureCurrentWorkspace({
          preferredWorkspaceId:
            options?.preferredCollaborationId ||
            options?.preferredLegacyCollaborationId ||
            options?.preferredWorkspaceId,
          preferredWorkspaceType: options?.preferredWorkspaceType,
          preferredCollaborationId:
            options?.preferredCollaborationId ||
            options?.preferredLegacyCollaborationId,
          preferPersonal: options?.preferPersonalWorkspace ?? hasPersonalWorkspaceAccess.value
        })
        return
      }

      workspaceStore.ensureCurrentWorkspace({
        preferredWorkspaceId:
          options?.preferredCollaborationId ||
          options?.preferredLegacyCollaborationId ||
          options?.preferredWorkspaceId,
        preferredWorkspaceType: options?.preferredWorkspaceType,
        preferredCollaborationId:
          options?.preferredCollaborationId ||
          options?.preferredLegacyCollaborationId,
        preferPersonal: options?.preferPersonalWorkspace ?? hasPersonalWorkspaceAccess.value
      })
    }

    const loadMyCollaborations = async (options?: {
      preferredCollaborationId?: string
      preferredLegacyCollaborationId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferPersonalWorkspace?: boolean
    }) => {
      loading.value = true
      try {
        const [collaborationWorkspaces] = await Promise.all([
          fetchGetMyCollaborations(),
          workspaceStore.loadMyWorkspaces({
            preferredWorkspaceId:
              options?.preferredCollaborationId ||
              options?.preferredLegacyCollaborationId ||
              options?.preferredWorkspaceId,
            preferredWorkspaceType: options?.preferredWorkspaceType,
            preferredCollaborationId:
              options?.preferredCollaborationId ||
              options?.preferredLegacyCollaborationId,
            preferPersonal: options?.preferPersonalWorkspace ?? hasPersonalWorkspaceAccess.value
          })
        ])
        collaborationList.value = collaborationWorkspaces
        ensureCurrentCollaboration(options)
        return collaborationWorkspaces
      } catch (error) {
        clearCollaborationContext()
        if (error instanceof HttpError && [400, 404, 3006].includes(error.code)) {
          return []
        }
        throw error
      } finally {
        loading.value = false
      }
    }

    const clearCollaborationContext = () => {
      collaborationList.value = []
      hasPersonalWorkspaceAccess.value = false
      loading.value = false
      workspaceStore.clearWorkspaceContext()
    }

    return {
      currentContextMode,
      currentCollaborationId,
      collaborationList,
      loading,
      hasPersonalWorkspaceAccess,
      currentCollaboration,
      hasCollaborations,
      isPersonalWorkspaceContext,
      shouldShowSwitcher,
      setCurrentCollaborationId,
      setCurrentContextMode,
      setPersonalWorkspaceAccess,
      enterPersonalWorkspaceContext,
      enterCollaborationContext,
      setCollaborationList,
      ensureCurrentCollaboration,
      loadMyCollaborations,
      clearCollaborationContext,
      currentAuthWorkspaceId: computed(() => workspaceStore.currentAuthWorkspaceId),
      currentAuthWorkspaceType: computed(() => workspaceStore.currentAuthWorkspaceType),
      currentAuthWorkspace: computed(() => workspaceStore.currentAuthWorkspace),
      currentAuthWorkspaceCollaborationId,
      personalWorkspace: computed(() => workspaceStore.personalWorkspace),
      workspaceList: computed(() => workspaceStore.workspaceList)
    }
  },
  {
    persist: {
      storage: localStorage
    }
  }
)
