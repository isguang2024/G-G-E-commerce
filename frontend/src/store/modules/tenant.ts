import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchGetMyTeams } from '@/api/team'
import { HttpError } from '@/utils/http/error'

export type AppContextMode = 'platform' | 'team'

export function hasPlatformAccessByUserInfo(
  userInfo?: Partial<Api.Auth.UserInfo> | null
): boolean {
  if (!userInfo) return false
  if (userInfo.is_super_admin) return true
  if (!Array.isArray(userInfo.actions)) return false
  return userInfo.actions.some((item) => {
    const key = `${item || ''}`.trim()
    return key.startsWith('system.') || key.startsWith('platform.') || key.startsWith('tenant.')
  })
}

export const useTenantStore = defineStore(
  'tenantStore',
  () => {
    const currentContextMode = ref<AppContextMode>('platform')
    const currentTenantId = ref('')
    const teamList = ref<Api.SystemManage.TeamListItem[]>([])
    const loading = ref(false)
    const hasPlatformAccess = ref(false)

    const currentTeam = computed(
      () => teamList.value.find((item) => item.id === currentTenantId.value) || null
    )
    const hasTeams = computed(() => teamList.value.length > 0)
    const isPlatformContext = computed(() => currentContextMode.value === 'platform')
    const shouldShowSwitcher = computed(() => teamList.value.length > 1)

    const setCurrentTenantId = (tenantId: string) => {
      currentTenantId.value = tenantId
      currentContextMode.value = tenantId ? 'team' : 'platform'
    }

    const setCurrentContextMode = (mode: AppContextMode) => {
      currentContextMode.value = mode
      if (mode !== 'team') {
        currentTenantId.value = ''
      }
    }

    const setPlatformAccess = (enabled: boolean) => {
      hasPlatformAccess.value = enabled
      if (!enabled && currentContextMode.value === 'platform' && teamList.value.length > 0) {
        currentContextMode.value = 'team'
        currentTenantId.value = teamList.value[0]?.id || ''
      }
    }

    const enterPlatformContext = () => {
      currentContextMode.value = 'platform'
      currentTenantId.value = ''
    }

    const enterTeamContext = (tenantId: string) => {
      currentContextMode.value = 'team'
      currentTenantId.value = tenantId
    }

    const setTeamList = (teams: Api.SystemManage.TeamListItem[]) => {
      teamList.value = teams
    }

    const ensureCurrentTenant = (options?: {
      preferredTenantId?: string
      preferPlatform?: boolean
    }) => {
      if (teamList.value.length === 0) {
        currentTenantId.value = ''
        currentContextMode.value = hasPlatformAccess.value ? 'platform' : 'team'
        return
      }

      if (
        options?.preferredTenantId &&
        teamList.value.some((item) => item.id === options.preferredTenantId)
      ) {
        enterTeamContext(options.preferredTenantId)
        return
      }

      if (
        currentContextMode.value === 'team' &&
        teamList.value.some((item) => item.id === currentTenantId.value)
      ) {
        return
      }

      enterTeamContext(teamList.value[0]?.id || '')
    }

    const loadMyTeams = async (options?: {
      preferredTenantId?: string
      preferPlatform?: boolean
    }) => {
      loading.value = true
      try {
        const teams = await fetchGetMyTeams()
        teamList.value = teams
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
      currentContextMode.value = 'platform'
      currentTenantId.value = ''
      teamList.value = []
      hasPlatformAccess.value = false
      loading.value = false
    }

    return {
      currentContextMode,
      currentTenantId,
      teamList,
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
      enterTeamContext,
      setTeamList,
      ensureCurrentTenant,
      loadMyTeams,
      clearTenantContext
    }
  },
  {
    persist: {
      key: 'tenant',
      storage: localStorage
    }
  }
)
