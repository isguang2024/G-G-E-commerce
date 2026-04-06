import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetMyTeams } from '@/api/team'
import { HttpError } from '@/utils/http/error'
import { useWorkspaceStore } from './workspace'

export type AppContextMode = 'platform' | 'collaboration'

export function hasPlatformAccessByUserInfo(userInfo?: Partial<Api.Auth.UserInfo> | null): boolean {
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

// 兼容层：保留文件名和导出名，避免一次性改爆所有引用；
// 当前主语义已经切到 collaboration workspace。
export const useTenantStore = defineStore(
  'collaborationWorkspaceStore',
  () => {
    const workspaceStore = useWorkspaceStore()
    const collaborationWorkspaceList = ref<Api.SystemManage.TeamListItem[]>([])
    const loading = ref(false)
    const hasPlatformAccess = ref(false)
    const legacyCollaborationWorkspaceId = ref('')

    const currentContextMode = computed<AppContextMode>(() =>
      workspaceStore.currentAuthWorkspaceType === 'collaboration' ? 'collaboration' : 'platform'
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

    const currentTeam = computed(() => {
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

    const hasTeams = computed(() => collaborationWorkspaceList.value.length > 0)
    const isPlatformContext = computed(() => currentContextMode.value === 'platform')
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

    const setCurrentTenantId = (collaborationWorkspaceId: string) => {
      if (!collaborationWorkspaceId) {
        enterPlatformContext()
        return
      }
      enterCollaborationContext(collaborationWorkspaceId)
    }

    const setCurrentContextMode = (mode: AppContextMode) => {
      if (mode !== 'collaboration') {
        enterPlatformContext()
        return
      }
      if (currentCollaborationWorkspaceId.value) return
      const firstId =
        collaborationWorkspaceList.value[0]?.collaborationWorkspaceId ||
        collaborationWorkspaceList.value[0]?.workspaceId ||
        ''
      if (firstId) enterCollaborationContext(firstId)
    }

    const setPlatformAccess = (enabled: boolean) => {
      hasPlatformAccess.value = enabled
      if (!enabled && currentContextMode.value === 'platform' && collaborationWorkspaceList.value.length > 0) {
        const firstId =
          collaborationWorkspaceList.value[0]?.collaborationWorkspaceId ||
          collaborationWorkspaceList.value[0]?.workspaceId ||
          ''
        if (firstId) {
          enterCollaborationContext(firstId)
        }
      }
    }

    const enterPlatformContext = () => {
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

    const setTeamList = (items: Api.SystemManage.TeamListItem[]) => {
      collaborationWorkspaceList.value = items || []
      syncLegacyCollaborationWorkspaceId()
    }

    const ensureCurrentTenant = (options?: {
      preferredCollaborationWorkspaceId?: string
      preferredLegacyCollaborationWorkspaceId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
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
          preferPersonal: options?.preferPlatform ?? hasPlatformAccess.value
        })
        return
      }

      workspaceStore.ensureCurrentWorkspace({
        preferredWorkspaceId: options?.preferredWorkspaceId,
        preferredWorkspaceType: options?.preferredWorkspaceType,
        preferredCollaborationWorkspaceId: options?.preferredCollaborationWorkspaceId,
        preferredCollaborationWorkspaceIdFromRecord:
          options?.preferredLegacyCollaborationWorkspaceId,
        preferPersonal: options?.preferPlatform ?? hasPlatformAccess.value
      })
      syncLegacyCollaborationWorkspaceId(options?.preferredLegacyCollaborationWorkspaceId || '')
    }

    const loadMyTeams = async (options?: {
      preferredCollaborationWorkspaceId?: string
      preferredLegacyCollaborationWorkspaceId?: string
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferPlatform?: boolean
    }) => {
      loading.value = true
      try {
        const [teams] = await Promise.all([
          fetchGetMyTeams(),
          workspaceStore.loadMyWorkspaces({
            preferredWorkspaceId: options?.preferredWorkspaceId,
            preferredWorkspaceType: options?.preferredWorkspaceType,
            preferredCollaborationWorkspaceId: options?.preferredCollaborationWorkspaceId,
            preferredCollaborationWorkspaceIdFromRecord:
              options?.preferredLegacyCollaborationWorkspaceId,
            preferPersonal: options?.preferPlatform ?? hasPlatformAccess.value
          })
        ])
        collaborationWorkspaceList.value = teams
        ensureCurrentTenant(options)
        return teams
      } catch (error) {
        clearTenantContext()
        if (error instanceof HttpError && [400, 404, 3006].includes(error.code)) {
          return []
        }
        throw error
      } finally {
        loading.value = false
      }
    }

    const clearTenantContext = () => {
      legacyCollaborationWorkspaceId.value = ''
      collaborationWorkspaceList.value = []
      hasPlatformAccess.value = false
      loading.value = false
      workspaceStore.clearWorkspaceContext()
    }

    return {
      currentContextMode,
      currentCollaborationWorkspaceId,
      teamList: collaborationWorkspaceList,
      collaborationWorkspaceList,
      loading,
      hasPlatformAccess,
      currentTeam,
      hasTeams,
      isPlatformContext,
      shouldShowSwitcher,
      setCurrentTenantId,
      setCurrentContextMode,
      setPlatformAccess,
      enterPlatformContext,
      enterTeamContext: enterCollaborationContext,
      enterCollaborationContext,
      setTeamList,
      ensureCurrentTenant,
      loadMyTeams,
      clearTenantContext,
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

export const useCollaborationWorkspaceStore = useTenantStore
