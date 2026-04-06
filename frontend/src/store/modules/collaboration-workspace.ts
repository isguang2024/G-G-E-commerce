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
      key.startsWith('platform.') ||
      key.startsWith('collaboration_workspace.')
    )
  })
}

export const hasPlatformAccessByUserInfo = hasPersonalWorkspaceAccessByUserInfo

export const useCollaborationWorkspaceStore = defineStore(
  'collaborationWorkspaceStore',
  () => {
    const workspaceStore = useWorkspaceStore()
    const collaborationWorkspaceList = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
    const loading = ref(false)
    const hasPersonalWorkspaceAccess = ref(false)
    const legacyCollaborationWorkspaceId = ref('')

    const currentContextMode = computed<WorkspaceContextMode>(() =>
      workspaceStore.currentAuthWorkspaceType === 'collaboration' ? 'collaboration' : 'personal'
    )

    const currentCollaborationWorkspaceId = computed(() => {
      if (workspaceStore.currentAuthWorkspaceType !== 'collaboration') return ''
      const matched = collaborationWorkspaceList.value.find(
        (item) =>
          item.collaborationWorkspaceId === workspaceStore.currentAuthWorkspaceId ||
          item.workspaceId === workspaceStore.currentAuthWorkspaceId
      )
      return (
        matched?.collaborationWorkspaceId ||
        matched?.workspaceId ||
        legacyCollaborationWorkspaceId.value ||
        workspaceStore.currentAuthWorkspaceId
      )
    })

    const currentCollaborationWorkspace = computed(() => {
      const currentId = currentCollaborationWorkspaceId.value
      return (
        collaborationWorkspaceList.value.find(
          (item) =>
            item.collaborationWorkspaceId === currentId ||
            item.workspaceId === currentId ||
            item.id === currentId
        ) || null
      )
    })

    const hasCollaborationWorkspaces = computed(() => collaborationWorkspaceList.value.length > 0)
    const isPersonalWorkspaceContext = computed(() => currentContextMode.value === 'personal')
    const shouldShowSwitcher = computed(() => collaborationWorkspaceList.value.length > 1)

    const syncLegacyCollaborationWorkspaceId = (preferredId = '') => {
      if (workspaceStore.currentAuthWorkspaceType !== 'collaboration') {
        legacyCollaborationWorkspaceId.value = ''
        return
      }
      const matched = collaborationWorkspaceList.value.find(
        (item) =>
          item.collaborationWorkspaceId === workspaceStore.currentAuthWorkspaceId ||
          item.workspaceId === workspaceStore.currentAuthWorkspaceId
      )
      legacyCollaborationWorkspaceId.value = matched?.id || preferredId || ''
    }

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
      legacyCollaborationWorkspaceId.value = ''
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
      legacyCollaborationWorkspaceId.value = matched?.id || legacyCollaborationWorkspaceId.value
      workspaceStore.enterWorkspaceById(
        matched?.workspaceId || matched?.collaborationWorkspaceId || normalizedId,
        matched?.workspaceType || 'collaboration'
      )
    }

    const setCollaborationWorkspaceList = (
      items: Api.SystemManage.CollaborationWorkspaceListItem[]
    ) => {
      collaborationWorkspaceList.value = items || []
      syncLegacyCollaborationWorkspaceId()
    }

    const ensureCurrentCollaborationWorkspace = (options?: {
      preferredCollaborationWorkspaceId?: string
      preferredLegacyCollaborationWorkspaceId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferPersonalWorkspace?: boolean
      preferPlatform?: boolean
    }) => {
      if (collaborationWorkspaceList.value.length === 0) {
        legacyCollaborationWorkspaceId.value = ''
        workspaceStore.ensureCurrentWorkspace({
          preferredWorkspaceId: options?.preferredWorkspaceId,
          preferredWorkspaceType: options?.preferredWorkspaceType,
          preferredCollaborationWorkspaceId: options?.preferredCollaborationWorkspaceId,
          preferredCollaborationWorkspaceIdFromRecord:
            options?.preferredLegacyCollaborationWorkspaceId,
          preferPersonal:
            options?.preferPersonalWorkspace ??
            options?.preferPlatform ??
            hasPersonalWorkspaceAccess.value
        })
        return
      }

      workspaceStore.ensureCurrentWorkspace({
        preferredWorkspaceId: options?.preferredWorkspaceId,
        preferredWorkspaceType: options?.preferredWorkspaceType,
        preferredCollaborationWorkspaceId: options?.preferredCollaborationWorkspaceId,
        preferredCollaborationWorkspaceIdFromRecord:
          options?.preferredLegacyCollaborationWorkspaceId,
        preferPersonal:
          options?.preferPersonalWorkspace ??
          options?.preferPlatform ??
          hasPersonalWorkspaceAccess.value
      })
      syncLegacyCollaborationWorkspaceId(options?.preferredLegacyCollaborationWorkspaceId || '')
    }

    const loadMyCollaborationWorkspaces = async (options?: {
      preferredCollaborationWorkspaceId?: string
      preferredLegacyCollaborationWorkspaceId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferPersonalWorkspace?: boolean
      preferPlatform?: boolean
    }) => {
      loading.value = true
      try {
        const [teams] = await Promise.all([
          fetchGetMyCollaborationWorkspaces(),
          workspaceStore.loadMyWorkspaces({
            preferredWorkspaceId: options?.preferredWorkspaceId,
            preferredWorkspaceType: options?.preferredWorkspaceType,
            preferredCollaborationWorkspaceId: options?.preferredCollaborationWorkspaceId,
            preferredCollaborationWorkspaceIdFromRecord:
              options?.preferredLegacyCollaborationWorkspaceId,
            preferPersonal:
              options?.preferPersonalWorkspace ??
              options?.preferPlatform ??
              hasPersonalWorkspaceAccess.value
          })
        ])
        collaborationWorkspaceList.value = teams
        ensureCurrentCollaborationWorkspace(options)
        return teams
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
      legacyCollaborationWorkspaceId.value = ''
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
      hasPlatformAccess: hasPersonalWorkspaceAccess,
      isPlatformContext: isPersonalWorkspaceContext,
      setPlatformAccess: setPersonalWorkspaceAccess,
      enterPlatformContext: enterPersonalWorkspaceContext,
      enterCollaborationContext,
      setCollaborationWorkspaceList,
      ensureCurrentCollaborationWorkspace,
      loadMyCollaborationWorkspaces,
      clearCollaborationWorkspaceContext,
      currentAuthWorkspaceId: computed(() => workspaceStore.currentAuthWorkspaceId),
      currentAuthWorkspaceType: computed(() => workspaceStore.currentAuthWorkspaceType),
      currentAuthWorkspace: computed(() => workspaceStore.currentAuthWorkspace),
      personalWorkspace: computed(() => workspaceStore.personalWorkspace),
      workspaceList: computed(() => workspaceStore.workspaceList)
    }
  },
  {
    persist: {
      key: 'collaboration-workspace-adapter',
      storage: localStorage
    }
  }
)
