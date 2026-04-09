import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetMyCollaborationWorkspaces } from '@/api/collaboration-workspace'
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
      key === 'collaboration_workspace.manage'
    )
  })
}

export const useCollaborationWorkspaceStore = defineStore(
  'collaborationWorkspaceStore',
  () => {
    const workspaceStore = useWorkspaceStore()
    const collaborationWorkspaceList = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
    const loading = ref(false)
    const hasPersonalWorkspaceAccess = ref(false)

    const currentContextMode = computed<WorkspaceContextMode>(() =>
      workspaceStore.currentAuthWorkspaceType === 'collaboration' ? 'collaboration' : 'personal'
    )

    // V5：协作空间 ID 单一语义 = collaborationWorkspaceId（spec 字段映射）。
    // 旧 workspaceId / record.id 兜底已废弃。
    const currentAuthWorkspaceCollaborationWorkspaceId = computed(() => {
      const currentWorkspace = workspaceStore.currentAuthWorkspace
      if (!currentWorkspace) return ''
      return `${currentWorkspace.collaborationWorkspaceId || ''}`.trim()
    })

    const currentCollaborationWorkspaceId = computed(() => {
      if (workspaceStore.currentAuthWorkspaceType !== 'collaboration') return ''
      return currentAuthWorkspaceCollaborationWorkspaceId.value
    })

    const currentCollaborationWorkspace = computed(() => {
      const currentId = currentCollaborationWorkspaceId.value
      return (
        collaborationWorkspaceList.value.find(
          (item) => item.collaborationWorkspaceId === currentId
        ) || null
      )
    })

    const hasCollaborationWorkspaces = computed(() => collaborationWorkspaceList.value.length > 0)
    const isPersonalWorkspaceContext = computed(() => currentContextMode.value === 'personal')
    const shouldShowSwitcher = computed(() => collaborationWorkspaceList.value.length > 1)

    const setCurrentCollaborationWorkspaceId = (collaborationWorkspaceId: string) => {
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
      if (currentCollaborationWorkspaceId.value) return
      const firstId =
        collaborationWorkspaceList.value[0]?.collaborationWorkspaceId ||
        collaborationWorkspaceList.value[0]?.workspaceId ||
        ''
      if (firstId) enterCollaborationContext(firstId)
    }

    const setPersonalWorkspaceAccess = (enabled: boolean) => {
      hasPersonalWorkspaceAccess.value = enabled
      if (
        !enabled &&
        currentContextMode.value === 'personal' &&
        collaborationWorkspaceList.value.length > 0
      ) {
        const firstId =
          collaborationWorkspaceList.value[0]?.collaborationWorkspaceId ||
          collaborationWorkspaceList.value[0]?.workspaceId ||
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
      const matched = collaborationWorkspaceList.value.find(
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

    const setCollaborationWorkspaceList = (
      items: Api.SystemManage.CollaborationWorkspaceListItem[]
    ) => {
      collaborationWorkspaceList.value = items || []
    }

    const ensureCurrentCollaborationWorkspace = (options?: {
      preferredCollaborationWorkspaceId?: string
      preferredLegacyCollaborationWorkspaceId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferPersonalWorkspace?: boolean
    }) => {
      if (collaborationWorkspaceList.value.length === 0) {
        workspaceStore.ensureCurrentWorkspace({
          preferredWorkspaceId:
            options?.preferredCollaborationWorkspaceId ||
            options?.preferredLegacyCollaborationWorkspaceId ||
            options?.preferredWorkspaceId,
          preferredWorkspaceType: options?.preferredWorkspaceType,
          preferredCollaborationWorkspaceId:
            options?.preferredCollaborationWorkspaceId ||
            options?.preferredLegacyCollaborationWorkspaceId,
          preferPersonal: options?.preferPersonalWorkspace ?? hasPersonalWorkspaceAccess.value
        })
        return
      }

      workspaceStore.ensureCurrentWorkspace({
        preferredWorkspaceId:
          options?.preferredCollaborationWorkspaceId ||
          options?.preferredLegacyCollaborationWorkspaceId ||
          options?.preferredWorkspaceId,
        preferredWorkspaceType: options?.preferredWorkspaceType,
        preferredCollaborationWorkspaceId:
          options?.preferredCollaborationWorkspaceId ||
          options?.preferredLegacyCollaborationWorkspaceId,
        preferPersonal: options?.preferPersonalWorkspace ?? hasPersonalWorkspaceAccess.value
      })
    }

    const loadMyCollaborationWorkspaces = async (options?: {
      preferredCollaborationWorkspaceId?: string
      preferredLegacyCollaborationWorkspaceId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferPersonalWorkspace?: boolean
    }) => {
      loading.value = true
      try {
        const [collaborationWorkspaces] = await Promise.all([
          fetchGetMyCollaborationWorkspaces(),
          workspaceStore.loadMyWorkspaces({
            preferredWorkspaceId:
              options?.preferredCollaborationWorkspaceId ||
              options?.preferredLegacyCollaborationWorkspaceId ||
              options?.preferredWorkspaceId,
            preferredWorkspaceType: options?.preferredWorkspaceType,
            preferredCollaborationWorkspaceId:
              options?.preferredCollaborationWorkspaceId ||
              options?.preferredLegacyCollaborationWorkspaceId,
            preferPersonal: options?.preferPersonalWorkspace ?? hasPersonalWorkspaceAccess.value
          })
        ])
        collaborationWorkspaceList.value = collaborationWorkspaces
        ensureCurrentCollaborationWorkspace(options)
        return collaborationWorkspaces
      } catch (error) {
        clearCollaborationWorkspaceContext()
        if (error instanceof HttpError && [400, 404, 3006].includes(error.code)) {
          return []
        }
        throw error
      } finally {
        loading.value = false
      }
    }

    const clearCollaborationWorkspaceContext = () => {
      collaborationWorkspaceList.value = []
      hasPersonalWorkspaceAccess.value = false
      loading.value = false
      workspaceStore.clearWorkspaceContext()
    }

    return {
      currentContextMode,
      currentCollaborationWorkspaceId,
      collaborationWorkspaceList,
      loading,
      hasPersonalWorkspaceAccess,
      currentCollaborationWorkspace,
      hasCollaborationWorkspaces,
      isPersonalWorkspaceContext,
      shouldShowSwitcher,
      setCurrentCollaborationWorkspaceId,
      setCurrentContextMode,
      setPersonalWorkspaceAccess,
      enterPersonalWorkspaceContext,
      enterCollaborationContext,
      setCollaborationWorkspaceList,
      ensureCurrentCollaborationWorkspace,
      loadMyCollaborationWorkspaces,
      clearCollaborationWorkspaceContext,
      currentAuthWorkspaceId: computed(() => workspaceStore.currentAuthWorkspaceId),
      currentAuthWorkspaceType: computed(() => workspaceStore.currentAuthWorkspaceType),
      currentAuthWorkspace: computed(() => workspaceStore.currentAuthWorkspace),
      currentAuthWorkspaceCollaborationWorkspaceId,
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
