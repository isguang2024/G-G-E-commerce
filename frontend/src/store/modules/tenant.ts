import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchGetMyTeams } from '@/api/team'
import { HttpError } from '@/utils/http/error'

export const useTenantStore = defineStore(
  'tenantStore',
  () => {
    const currentTenantId = ref('')
    const teamList = ref<Api.SystemManage.TeamListItem[]>([])
    const loading = ref(false)

    const currentTeam = computed(
      () => teamList.value.find((item) => item.id === currentTenantId.value) || null
    )
    const hasTeams = computed(() => teamList.value.length > 0)

    const setCurrentTenantId = (tenantId: string) => {
      currentTenantId.value = tenantId
    }

    const setTeamList = (teams: Api.SystemManage.TeamListItem[]) => {
      teamList.value = teams
    }

    const ensureCurrentTenant = () => {
      if (teamList.value.length === 0) {
        currentTenantId.value = ''
        return
      }

      const matched = teamList.value.some((item) => item.id === currentTenantId.value)
      if (!matched) {
        currentTenantId.value = teamList.value[0]?.id || ''
      }
    }

    const loadMyTeams = async () => {
      loading.value = true
      try {
        const teams = await fetchGetMyTeams()
        teamList.value = teams
        ensureCurrentTenant()
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
      currentTenantId.value = ''
      teamList.value = []
      loading.value = false
    }

    return {
      currentTenantId,
      teamList,
      loading,
      currentTeam,
      hasTeams,
      setCurrentTenantId,
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
