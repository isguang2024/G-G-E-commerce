/**
 * user-permission-test-drawer 视图脚本：所有 reactive state、computed、watch、handler 集中在此。
 *
 * 抽离自 user-permission-test-drawer.vue，.vue 文件保留 defineProps/defineEmits 等编译宏与
 * template/style 块，调用本 composable 拉取所有模板绑定。
 */
import { computed, nextTick, ref, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { CascaderProps } from 'element-plus'
import {
  fetchGetUserPermissionDiagnosis,
  fetchGetUserPermissionMenus,
  fetchGetUserCollaborationWorkspaces,
  fetchRefreshUserPermissionSnapshot
} from '@/api/system-manage'
import {
  countVisibleMenuLeaves,
  ensureExpandedMenus,
  filterNestedOptions,
  formatBoundaryState,
  formatMemberStatus,
  formatPermissionStatus,
  formatRoleCode,
  getBoundaryStateTagType,
  getMemberStatusTagType,
  normalizePermissionMenuOptions,
  type MenuOption
} from './user-permission-test-drawer.helpers'

export interface UseUserPermissionTestDrawerProps {
  modelValue: boolean
  userData?: Api.SystemManage.UserListItem
}

export interface UseUserPermissionTestDrawerEmit {
  (e: 'update:modelValue', value: boolean): void
}

export function useUserPermissionTestDrawer(
  props: UseUserPermissionTestDrawerProps,
  emit: UseUserPermissionTestDrawerEmit
) {
  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const testing = ref(false)
  const refreshing = ref(false)
  const contextType = ref<'personal' | 'collaboration'>('personal')
  const activeTab = ref<'permission' | 'menus' | 'roles'>('permission')
  const selectedCollaborationWorkspaceId = ref('')
  const permissionKey = ref('')
  const diagnosisData = ref<Api.SystemManage.UserPermissionDiagnosisResponse>()
  const collaborationWorkspaceOptions = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
  const permissionMenus = ref<Api.SystemManage.UserPermissionMenuNode[]>([])
  const menuKeyword = ref('')
  const showHiddenMenus = ref(true)
  const showIframeMenus = ref(true)
  const showEnabledMenus = ref(true)
  const showMenuPath = ref(false)
  const selectedMenuPath = ref<string[]>([])
  const rolePagination = ref({
    current: 1,
    size: 10
  })
  const menuPanelRef = ref<any>()

  const menuCascaderProps: CascaderProps = {
    emitPath: true,
    checkStrictly: true,
    expandTrigger: 'click',
    showPrefix: true
  }

  const userTitle = computed(
    () => props.userData?.nickName || props.userData?.userName || props.userData?.id || ''
  )

  const selectedCollaborationWorkspaceName = computed(
    () =>
      collaborationWorkspaceOptions.value.find(
        (item) => item.id === selectedCollaborationWorkspaceId.value
      )?.name || ''
  )

  const menuOptions = computed<MenuOption[]>(() =>
    normalizePermissionMenuOptions(permissionMenus.value)
  )

  const filteredMenuOptions = computed(() => {
    const keyword = menuKeyword.value.trim().toLowerCase()
    return filterNestedOptions(menuOptions.value, (node) => {
      if (!node.leaf) return !keyword
      if (!showHiddenMenus.value && node.hidden) return false
      if (!showIframeMenus.value && node.isIframe) return false
      if (!showEnabledMenus.value && node.isEnable !== false) return false
      if (keyword && !`${node.label || ''} ${node.path || ''}`.toLowerCase().includes(keyword))
        return false
      return true
    })
  })

  const summaryItems = computed(() => {
    const items: Array<{
      label: string
      value: string | number
      type?: 'success' | 'warning' | 'info' | 'primary' | 'danger'
    }> = []
    const snapshot = diagnosisData.value?.snapshot
    if (!snapshot) {
      return [
        {
          label: '快照',
          value: '未加载',
          type: 'info' as const
        }
      ]
    }
    items.push(
      {
        label: '功能包',
        value:
          contextType.value === 'personal'
            ? (snapshot.expandedPackageCount ?? 0)
            : (snapshot.expandedPackageCount ?? 0)
      },
      {
        label: contextType.value === 'personal' ? '个人空间权限' : '协作空间生效权限',
        value:
          contextType.value === 'personal'
            ? (snapshot.actionCount ?? 0)
            : (snapshot.effectiveActionCount ?? 0),
        type: 'success'
      },
      {
        label: '菜单数',
        value: countVisibleMenuLeaves(filteredMenuOptions.value),
        type: 'primary'
      }
    )
    if (contextType.value === 'personal') {
      items.push({
        label: '已禁用',
        value: snapshot.disabledActionCount ?? 0,
        type: 'warning'
      })
    } else {
      items.push({
        label: '协作空间屏蔽',
        value: snapshot.blockedActionCount ?? 0,
        type: 'warning'
      })
    }
    items.push({
      label: '刷新时间',
      value: snapshot.refreshedAt || '-',
      type: 'info'
    })
    return items
  })

  const roleRows = computed(() => diagnosisData.value?.roles || [])
  const pagedRoleRows = computed(() => {
    const start = (rolePagination.value.current - 1) * rolePagination.value.size
    return roleRows.value.slice(start, start + rolePagination.value.size)
  })

  async function loadCollaborationWorkspaces() {
    const userId = props.userData?.id
    if (!userId) {
      collaborationWorkspaceOptions.value = []
      return
    }
    try {
      collaborationWorkspaceOptions.value = await fetchGetUserCollaborationWorkspaces(userId)
      if (
        contextType.value === 'collaboration' &&
        selectedCollaborationWorkspaceId.value &&
        !collaborationWorkspaceOptions.value.some(
          (item) => item.id === selectedCollaborationWorkspaceId.value
        )
      ) {
        selectedCollaborationWorkspaceId.value = ''
      }
      if (
        contextType.value === 'collaboration' &&
        !selectedCollaborationWorkspaceId.value &&
        collaborationWorkspaceOptions.value.length === 1
      ) {
        selectedCollaborationWorkspaceId.value = collaborationWorkspaceOptions.value[0].id
      }
    } catch {
      collaborationWorkspaceOptions.value = []
    }
  }

  async function loadDiagnosis() {
    const userId = props.userData?.id
    if (!userId) return
    if (contextType.value === 'collaboration' && !selectedCollaborationWorkspaceId.value) {
      diagnosisData.value = undefined
      permissionMenus.value = []
      return
    }
    loading.value = true
    try {
      const collaborationWorkspaceId =
        contextType.value === 'collaboration' ? selectedCollaborationWorkspaceId.value : undefined
      const [diagnosis, menus] = await Promise.all([
        fetchGetUserPermissionDiagnosis(userId, {
          collaborationWorkspaceId,
          permissionKey: permissionKey.value || undefined
        }),
        fetchGetUserPermissionMenus(userId, collaborationWorkspaceId)
      ])
      diagnosisData.value = diagnosis
      permissionMenus.value = menus
      rolePagination.value.current = 1
      await nextTick()
      ensureExpandedMenus(menuPanelRef.value, selectedMenuPath.value)
    } catch (error: any) {
      diagnosisData.value = undefined
      permissionMenus.value = []
      ElMessage.error(error?.message || '加载权限诊断失败')
    } finally {
      loading.value = false
    }
  }

  async function initialize() {
    activeTab.value = 'permission'
    await loadCollaborationWorkspaces()
    await loadDiagnosis()
  }

  async function handleTest() {
    if (!permissionKey.value.trim()) {
      ElMessage.warning('请输入权限键')
      return
    }
    if (contextType.value === 'collaboration' && !selectedCollaborationWorkspaceId.value) {
      ElMessage.warning('请选择协作空间')
      return
    }
    testing.value = true
    try {
      const collaborationWorkspaceId =
        contextType.value === 'collaboration' ? selectedCollaborationWorkspaceId.value : undefined
      const [diagnosis, menus] = await Promise.all([
        fetchGetUserPermissionDiagnosis(props.userData?.id || '', {
          collaborationWorkspaceId,
          permissionKey: permissionKey.value.trim()
        }),
        fetchGetUserPermissionMenus(props.userData?.id || '', collaborationWorkspaceId)
      ])
      diagnosisData.value = diagnosis
      permissionMenus.value = menus
      rolePagination.value.current = 1
    } catch (error: any) {
      ElMessage.error(error?.message || '权限测试失败')
    } finally {
      testing.value = false
    }
  }

  async function handleRefresh() {
    const userId = props.userData?.id
    if (!userId) return
    if (contextType.value === 'collaboration' && !selectedCollaborationWorkspaceId.value) {
      ElMessage.warning('请选择协作空间')
      return
    }
    refreshing.value = true
    try {
      const collaborationWorkspaceId =
        contextType.value === 'collaboration' ? selectedCollaborationWorkspaceId.value : undefined
      await fetchRefreshUserPermissionSnapshot(userId, collaborationWorkspaceId)
      const [diagnosis, menus] = await Promise.all([
        fetchGetUserPermissionDiagnosis(userId, {
          collaborationWorkspaceId,
          permissionKey: permissionKey.value.trim() || undefined
        }),
        fetchGetUserPermissionMenus(userId, collaborationWorkspaceId)
      ])
      diagnosisData.value = diagnosis
      permissionMenus.value = menus
      rolePagination.value.current = 1
      ElMessage.success('权限快照已刷新')
    } catch (error: any) {
      ElMessage.error(error?.message || '刷新权限快照失败')
    } finally {
      refreshing.value = false
    }
  }

  watch(
    () => props.modelValue,
    (open) => {
      if (!open) return
      void initialize()
    }
  )

  watch(contextType, async (value) => {
    if (!visible.value) return
    if (value === 'personal') {
      selectedCollaborationWorkspaceId.value = ''
      if (activeTab.value === 'roles') activeTab.value = 'permission'
    }
    await loadDiagnosis()
  })

  watch(selectedCollaborationWorkspaceId, async () => {
    if (!visible.value || contextType.value !== 'collaboration') return
    await loadDiagnosis()
  })

  watch(filteredMenuOptions, async () => {
    await nextTick()
    ensureExpandedMenus(menuPanelRef.value, selectedMenuPath.value)
  })

  watch(
    () => rolePagination.value.size,
    () => {
      rolePagination.value.current = 1
    }
  )

  return {
    Search,
    visible,
    loading,
    testing,
    refreshing,
    contextType,
    activeTab,
    selectedCollaborationWorkspaceId,
    permissionKey,
    diagnosisData,
    collaborationWorkspaceOptions,
    menuKeyword,
    showHiddenMenus,
    showIframeMenus,
    showEnabledMenus,
    showMenuPath,
    selectedMenuPath,
    rolePagination,
    menuPanelRef,
    menuCascaderProps,
    userTitle,
    selectedCollaborationWorkspaceName,
    filteredMenuOptions,
    summaryItems,
    roleRows,
    pagedRoleRows,
    handleTest,
    handleRefresh,
    formatPermissionStatus,
    formatMemberStatus,
    formatRoleCode,
    getMemberStatusTagType,
    formatBoundaryState,
    getBoundaryStateTagType
  }
}
