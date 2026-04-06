import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetMyWorkspaces, fetchSwitchWorkspace } from '@/api/workspace'
import { HttpError } from '@/utils/http/error'

export type AuthWorkspaceType = 'personal' | 'collaboration'

function normalizeWorkspaceType(value?: string | null): AuthWorkspaceType {
  return `${value || ''}`.trim() === 'collaboration' ? 'collaboration' : 'personal'
}

export const useWorkspaceStore = defineStore(
  'workspaceStore',
  () => {
    const workspaceList = ref<Api.SystemManage.WorkspaceItem[]>([])
    const currentAuthWorkspaceId = ref('')
    const currentAuthWorkspaceType = ref<AuthWorkspaceType>('personal')
    const loading = ref(false)

    const personalWorkspace = computed(
      () =>
        workspaceList.value.find(
          (item) => normalizeWorkspaceType(item.workspaceType) === 'personal'
        ) || null
    )
    const collaborationWorkspaces = computed(() =>
      workspaceList.value.filter(
        (item) => normalizeWorkspaceType(item.workspaceType) === 'collaboration'
      )
    )
    const currentAuthWorkspace = computed(() => {
      const matched = workspaceList.value.find((item) => item.id === currentAuthWorkspaceId.value)
      if (matched) {
        return matched
      }
      if (currentAuthWorkspaceType.value === 'personal') {
        return personalWorkspace.value
      }
      return collaborationWorkspaces.value[0] || null
    })

    const setWorkspaceList = (items: Api.SystemManage.WorkspaceItem[]) => {
      workspaceList.value = items || []
    }

    const upsertWorkspace = (workspace?: Api.SystemManage.WorkspaceItem | null) => {
      if (!workspace?.id) return
      const nextList = [...workspaceList.value]
      const matchedIndex = nextList.findIndex((item) => item.id === workspace.id)
      if (matchedIndex >= 0) {
        nextList.splice(matchedIndex, 1, { ...nextList[matchedIndex], ...workspace })
      } else {
        nextList.push(workspace)
      }
      workspaceList.value = nextList
    }

    const setCurrentAuthWorkspace = (workspaceId?: string, workspaceType?: string | null) => {
      const normalizedWorkspaceId = `${workspaceId || ''}`.trim()
      if (normalizedWorkspaceId) {
        const matched = workspaceList.value.find((item) => item.id === normalizedWorkspaceId)
        currentAuthWorkspaceId.value = normalizedWorkspaceId
        currentAuthWorkspaceType.value = normalizeWorkspaceType(
          workspaceType || matched?.workspaceType
        )
        return
      }

      currentAuthWorkspaceId.value = ''
      currentAuthWorkspaceType.value = normalizeWorkspaceType(workspaceType)
    }

    const ensureCurrentWorkspace = (options?: {
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferredCollaborationWorkspaceId?: string
      preferredCollaborationWorkspaceIdFromRecord?: string
      preferPersonal?: boolean
    }) => {
      if (!workspaceList.value.length) {
        currentAuthWorkspaceId.value = ''
        currentAuthWorkspaceType.value = normalizeWorkspaceType(options?.preferredWorkspaceType)
        return
      }

      const preferredWorkspaceId = `${options?.preferredWorkspaceId || ''}`.trim()
      if (preferredWorkspaceId) {
        const matchedByWorkspaceId = workspaceList.value.find(
          (item) => item.id === preferredWorkspaceId
        )
        if (matchedByWorkspaceId) {
          setCurrentAuthWorkspace(matchedByWorkspaceId.id, matchedByWorkspaceId.workspaceType)
          return
        }
      }

      const preferredCollaborationWorkspaceId = `${options?.preferredCollaborationWorkspaceId || ''}`.trim()
      if (preferredCollaborationWorkspaceId) {
        const matchedByCollaborationWorkspaceId = workspaceList.value.find(
          (item) =>
            normalizeWorkspaceType(item.workspaceType) === 'collaboration' &&
            item.id === preferredCollaborationWorkspaceId
        )
        if (matchedByCollaborationWorkspaceId) {
          setCurrentAuthWorkspace(
            matchedByCollaborationWorkspaceId.id,
            matchedByCollaborationWorkspaceId.workspaceType
          )
          return
        }
      }

      const preferredCollaborationWorkspaceIdFromRecord =
        `${options?.preferredCollaborationWorkspaceIdFromRecord || ''}`.trim()
      if (preferredCollaborationWorkspaceIdFromRecord) {
        const matchedByCollaborationWorkspaceRecord = workspaceList.value.find(
          (item) =>
            normalizeWorkspaceType(item.workspaceType) === 'collaboration' &&
            `${item.collaborationWorkspaceId || ''}`.trim() ===
              preferredCollaborationWorkspaceIdFromRecord
        )
        if (matchedByCollaborationWorkspaceRecord) {
          setCurrentAuthWorkspace(
            matchedByCollaborationWorkspaceRecord.id,
            matchedByCollaborationWorkspaceRecord.workspaceType
          )
          return
        }
      }

      if (
        currentAuthWorkspaceId.value &&
        workspaceList.value.some((item) => item.id === currentAuthWorkspaceId.value)
      ) {
        return
      }

      if (options?.preferPersonal !== false && personalWorkspace.value) {
        setCurrentAuthWorkspace(personalWorkspace.value.id, personalWorkspace.value.workspaceType)
        return
      }

      const fallbackWorkspace =
        collaborationWorkspaces.value[0] || personalWorkspace.value || workspaceList.value[0]
      if (fallbackWorkspace) {
        setCurrentAuthWorkspace(fallbackWorkspace.id, fallbackWorkspace.workspaceType)
      }
    }

    const loadMyWorkspaces = async (options?: {
      preferredWorkspaceId?: string
      preferredWorkspaceType?: string
      preferredCollaborationWorkspaceId?: string
      preferredCollaborationWorkspaceIdFromRecord?: string
      preferPersonal?: boolean
    }) => {
      loading.value = true
      try {
        const res = await fetchGetMyWorkspaces()
        workspaceList.value = res.records || []
        ensureCurrentWorkspace(options)
        return workspaceList.value
      } catch (error) {
        clearWorkspaceContext()
        if (error instanceof HttpError && [400, 404, 3006].includes(error.code)) {
          return []
        }
        throw error
      } finally {
        loading.value = false
      }
    }

    const enterPersonalWorkspace = () => {
      if (personalWorkspace.value) {
        setCurrentAuthWorkspace(personalWorkspace.value.id, personalWorkspace.value.workspaceType)
        return
      }
      setCurrentAuthWorkspace('', 'personal')
    }

    const enterWorkspaceById = (workspaceId: string, workspaceType?: string) => {
      setCurrentAuthWorkspace(workspaceId, workspaceType)
    }

    const switchWorkspace = async (workspaceId: string) => {
      const normalizedWorkspaceId = `${workspaceId || ''}`.trim()
      if (!normalizedWorkspaceId) return null

      const res = await fetchSwitchWorkspace(normalizedWorkspaceId)
      if (res?.workspace) {
        upsertWorkspace(res.workspace)
      }
      setCurrentAuthWorkspace(
        res?.auth_workspace_id || normalizedWorkspaceId,
        res?.auth_workspace_type || res?.workspace?.workspaceType
      )
      return res
    }

    const clearWorkspaceContext = () => {
      workspaceList.value = []
      currentAuthWorkspaceId.value = ''
      currentAuthWorkspaceType.value = 'personal'
      loading.value = false
    }

    return {
      workspaceList,
      currentAuthWorkspaceId,
      currentAuthWorkspaceType,
      currentAuthWorkspace,
      personalWorkspace,
      collaborationWorkspaces,
      loading,
      setWorkspaceList,
      upsertWorkspace,
      setCurrentAuthWorkspace,
      ensureCurrentWorkspace,
      loadMyWorkspaces,
      enterPersonalWorkspace,
      enterWorkspaceById,
      switchWorkspace,
      clearWorkspaceContext
    }
  },
  {
    persist: {
      key: 'workspace',
      storage: localStorage
    }
  }
)
