<template>
  <div class="collaboration-workspace-page art-full-height">
    <div class="page-top-stack">
      <CollaborationWorkspaceSearch
        v-show="showSearchBar"
        v-model="searchForm"
        :showExpand="true"
        @search="handleSearch"
        @reset="handleResetSearch"
      />

      <AdminWorkspaceHero
        :title="'协作空间管理'"
        :description="'统一管理协作空间边界、管理员与授权入口。'"
        :metrics="heroMetrics"
      >
        <div class="collaboration-workspace-hero-actions">
          <ElSelect
            v-model="selectedAppKey"
            clearable
            filterable
            placeholder="选择 App"
            class="collaboration-workspace-app-select"
            @change="handleManagedAppChange"
          >
            <ElOption
              v-for="item in appOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
          <ElButton
            v-action="'collaboration_workspace.manage'"
            type="primary"
            @click="showDialog('add')"
            v-ripple
          >
            新增协作空间
          </ElButton>
        </div>
      </AdminWorkspaceHero>
    </div>

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <div class="collaboration-workspace-toolbar-tip"
            >协作空间管理同步菜单与成员边界，建议优先从主账号确认管理员后再开通功能包。</div
          >
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <CollaborationWorkspaceDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :collaboration-workspace-data="currentCollaborationWorkspaceData"
        @submit="handleDialogSubmit"
      />

      <CollaborationWorkspaceMembersDrawer
        v-model:visible="membersDrawerVisible"
        :collaboration-workspace-id="currentCollaborationWorkspaceId"
        :collaboration-workspace-name="currentCollaborationWorkspaceName"
        @refresh="refreshData"
      />

      <CollaborationWorkspaceActionPermissionDialog
        v-model="actionDialogVisible"
        :collaboration-workspace-id="currentCollaborationWorkspaceId"
        :collaboration-workspace-name="currentCollaborationWorkspaceName"
        :app-key="targetAppKey"
        @success="refreshData"
      />

      <CollaborationWorkspaceMenuPermissionDialog
        v-model="menuDialogVisible"
        :collaboration-workspace-id="currentCollaborationWorkspaceId"
        :collaboration-workspace-name="currentCollaborationWorkspaceName"
        :app-key="targetAppKey"
        @success="refreshData"
      />

      <CollaborationWorkspaceFeaturePackageDialog
        v-model="packageDialogVisible"
        :collaboration-workspace-id="currentCollaborationWorkspaceId"
        :collaboration-workspace-name="currentCollaborationWorkspaceName"
        :app-key="targetAppKey"
        @success="refreshData"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, watch } from 'vue'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useAuth } from '@/hooks/core/useAuth'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetCollaborationWorkspaceList,
    fetchDeleteCollaborationWorkspace,
    fetchCreateCollaborationWorkspace,
    fetchUpdateCollaborationWorkspace
  } from '@/api/collaboration-workspace'
  import { fetchGetApps } from '@/domains/governance/api'
  import { useManagedAppScope } from '@/domains/app-runtime/useManagedAppScope'
  import { useCollaborationWorkspaceStore } from '@/store/modules/collaboration-workspace'
  import { useWorkspaceStore } from '@/store/modules/workspace'
  import CollaborationWorkspaceSearch from './modules/workspace-search.vue'
  import CollaborationWorkspaceDialog from './modules/workspace-dialog.vue'
  import CollaborationWorkspaceMembersDrawer from './modules/workspace-members-drawer.vue'
  import CollaborationWorkspaceActionPermissionDialog from './modules/workspace-permission-dialog.vue'
  import CollaborationWorkspaceMenuPermissionDialog from './modules/workspace-menu-permission-dialog.vue'
  import CollaborationWorkspaceFeaturePackageDialog from './modules/workspace-feature-package-dialog.vue'
  import {
    ElButton,
    ElTag,
    ElMessageBox,
    ElMessage,
    ElDropdown,
    ElDropdownMenu,
    ElDropdownItem,
    ElIcon
  } from 'element-plus'
  import { MoreFilled, Edit, Delete, UserFilled } from '@element-plus/icons-vue'
  import { DialogType } from '@/types'

  defineOptions({ name: 'CollaborationWorkspaceManagement' })

  type CollaborationWorkspaceListItem = Api.SystemManage.CollaborationWorkspaceListItem
  type AdminCandidate = {
    id?: string
    user_id?: string
    nickname?: string
    username?: string
    email?: string
  }
  type CollaborationWorkspaceWithAdmins = CollaborationWorkspaceListItem & {
    adminUsers?: AdminCandidate[]
    admin_users?: AdminCandidate[]
  }
  const { hasAction } = useAuth()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const collaborationWorkspaceStore = useCollaborationWorkspaceStore()
  const workspaceStore = useWorkspaceStore()

  const resolveAdminUsers = (item?: CollaborationWorkspaceListItem): AdminCandidate[] => {
    const normalized = item as CollaborationWorkspaceWithAdmins | undefined
    return normalized?.adminUsers || normalized?.admin_users || []
  }

  const dialogType = ref<DialogType>('add')
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const dialogVisible = ref(false)
  const currentCollaborationWorkspaceData = ref<Partial<CollaborationWorkspaceListItem>>({})
  const showSearchBar = ref(false)
  const membersDrawerVisible = ref(false)
  const actionDialogVisible = ref(false)
  const menuDialogVisible = ref(false)
  const packageDialogVisible = ref(false)
  const currentCollaborationWorkspaceId = ref('')
  const currentCollaborationWorkspaceName = ref('')

  const searchForm = ref({
    name: undefined as string | undefined,
    status: undefined as string | undefined
  })

  const getStatusConfig = (status: string) => {
    const map: Record<string, { type: 'success' | 'info' | 'warning' | 'danger'; text: string }> = {
      active: { type: 'success', text: '正常' },
      inactive: { type: 'danger', text: '停用' }
    }
    return map[status] || { type: 'info', text: status || '未知' }
  }

  const heroMetrics = computed(() => [
    { label: '当前 App', value: targetAppKey.value },
    { label: '协作空间总数', value: data.value.length },
    {
      label: '活跃协作空间',
      value: data.value.filter((item) => item.status === 'active').length
    },
    {
      label: '停用协作空间',
      value: data.value.filter((item) => item.status === 'inactive').length
    },
    {
      label: '管理员覆盖',
      value: data.value.reduce((total, item) => total + resolveAdminUsers(item).length, 0)
    }
  ])
  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: async (params) => {
        if (!targetAppKey.value) {
          return {
            records: [],
            total: 0,
            current: Number(params?.current || 1),
            size: Number(params?.size || 20)
          }
        }
        return fetchGetCollaborationWorkspaceList(
          params as Api.SystemManage.CollaborationWorkspaceSearchParams
        )
      },
      apiParams: {
        current: 1,
        size: 20,
        appKey: targetAppKey.value,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        { prop: 'name', label: '协作空间名称', minWidth: 140 },
        { prop: 'remark', label: '协作空间备注', width: 140 },
        {
          prop: 'adminUsers',
          label: '管理员',
          width: 200,
          formatter: (row: CollaborationWorkspaceListItem) => {
            // 兼容 adminUsers 和 admin_users 两种字段名
            const admins = resolveAdminUsers(row)
            if (!admins.length) return '-'
            return h('div', { class: 'flex flex-wrap gap-1' }, [
              ...admins.slice(0, 2).map((admin) =>
                h(ElTag, { size: 'small', type: 'info' }, () => {
                  const fallbackId = `${admin.user_id || admin.id || ''}`
                  return (
                    admin.nickname ||
                    admin.username ||
                    admin.email ||
                    `${fallbackId.substring(0, 8)}...`
                  )
                })
              ),
              admins.length > 2 ? h(ElTag, { size: 'small' }, () => `+${admins.length - 2}`) : null
            ])
          }
        },
        { prop: 'plan', label: '套餐', width: 100 },
        { prop: 'maxMembers', label: '最大成员数', width: 100 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row) => {
            const cfg = getStatusConfig(row.status)
            return h(ElTag, { type: cfg.type }, () => cfg.text)
          }
        },
        { prop: 'createTime', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 60,
          fixed: 'right',
          formatter: (row) => {
            const dropdown = h(
              ElDropdown,
              {
                trigger: 'click',
                onCommand: (cmd: string) => handleOperationCommand(cmd, row)
              },
              {
                default: () => h(ElButton, { icon: MoreFilled, circle: true, size: 'small' }),
                dropdown: () =>
                  h(ElDropdownMenu, {}, () => [
                    hasAction('collaboration_workspace.manage')
                      ? h(ElDropdownItem, { command: 'edit' }, () => [
                          h(ElIcon, {}, () => h(Edit)),
                          '编辑'
                        ])
                      : null,
                    hasAction('collaboration_workspace.manage')
                      ? h(ElDropdownItem, { command: 'view' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '查看人员'
                        ])
                      : null,
                    hasAction('collaboration_workspace.manage')
                      ? h(ElDropdownItem, { command: 'menu' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '菜单边界'
                        ])
                      : null,
                    hasAction('collaboration_workspace.manage')
                      ? h(ElDropdownItem, { command: 'action' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '协作空间边界'
                        ])
                      : null,
                    hasAction('platform.package.assign')
                      ? h(ElDropdownItem, { command: 'package' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '开通功能包'
                        ])
                      : null,
                    hasAction('collaboration_workspace.manage')
                      ? h(ElDropdownItem, { command: 'delete' }, () => [
                          h(ElIcon, {}, () => h(Delete)),
                          '删除'
                        ])
                      : null
                  ])
              }
            )
            return dropdown
          }
        }
      ]
    },
    transform: {
      dataTransformer: (records) => (Array.isArray(records) ? records : [])
    }
  })

  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, { ...params, appKey: targetAppKey.value || '' })
    getData()
  }

  async function loadAppOptions() {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }

  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
  }

  const handleResetSearch = () => {
    resetSearchParams()
    Object.assign(searchParams, {
      current: 1,
      size: pagination.size,
      appKey: targetAppKey.value || ''
    })
    getData()
  }

  const refreshWorkspaceContexts = async (preferredWorkspaceId?: string) => {
    await collaborationWorkspaceStore.loadMyCollaborationWorkspaces({
      preferredWorkspaceId: preferredWorkspaceId || workspaceStore.currentAuthWorkspaceId,
      preferredWorkspaceType: workspaceStore.currentAuthWorkspaceType,
      preferPersonalWorkspace: true
    })
  }

  const showDialog = (type: DialogType, row?: CollaborationWorkspaceListItem) => {
    dialogType.value = type
    currentCollaborationWorkspaceData.value = row ? { ...row } : {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  const handleOperationCommand = (command: string, row: CollaborationWorkspaceListItem) => {
    if (command === 'edit') {
      showDialog('edit', row)
    } else if (command === 'view') {
      showMembers(row)
    } else if (command === 'menu') {
      showMenuPermissions(row)
    } else if (command === 'action') {
      showActionPermissions(row)
    } else if (command === 'package') {
      showFeaturePackages(row)
    } else if (command === 'delete') {
      deleteCollaborationWorkspace(row)
    }
  }

  const showMembers = (row: CollaborationWorkspaceListItem) => {
    currentCollaborationWorkspaceId.value = row.id
    currentCollaborationWorkspaceName.value = row.name
    membersDrawerVisible.value = true
  }

  const showActionPermissions = (row: CollaborationWorkspaceListItem) => {
    currentCollaborationWorkspaceId.value = row.id
    currentCollaborationWorkspaceName.value = row.name
    actionDialogVisible.value = true
  }

  const showMenuPermissions = (row: CollaborationWorkspaceListItem) => {
    currentCollaborationWorkspaceId.value = row.id
    currentCollaborationWorkspaceName.value = row.name
    menuDialogVisible.value = true
  }

  const showFeaturePackages = (row: CollaborationWorkspaceListItem) => {
    currentCollaborationWorkspaceId.value = row.id
    currentCollaborationWorkspaceName.value = row.name
    packageDialogVisible.value = true
  }

  const deleteCollaborationWorkspace = (row: CollaborationWorkspaceListItem) => {
    ElMessageBox.confirm(`确定要删除协作空间「${row.name}」吗？`, '删除协作空间', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeleteCollaborationWorkspace(row.id))
      .then(async () => {
        await refreshWorkspaceContexts()
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
      })
  }

  const handleDialogSubmit = async (
    payload?:
      | Api.SystemManage.CollaborationWorkspaceCreateParams
      | Api.SystemManage.CollaborationWorkspaceUpdateParams
  ) => {
    if (!payload) {
      dialogVisible.value = false
      refreshData()
      return
    }
    const isAdd = dialogType.value === 'add'
    try {
      if (isAdd) {
        const created = await fetchCreateCollaborationWorkspace(
          payload as Api.SystemManage.CollaborationWorkspaceCreateParams
        )
        await refreshWorkspaceContexts(created?.id)
        ElMessage.success('添加成功')
      } else {
        const id = (currentCollaborationWorkspaceData.value as CollaborationWorkspaceListItem).id
        if (!id) return
        await fetchUpdateCollaborationWorkspace(
          id,
          payload as Api.SystemManage.CollaborationWorkspaceUpdateParams
        )
        await refreshWorkspaceContexts(id)
        ElMessage.success('更新成功')
      }
      dialogVisible.value = false
      currentCollaborationWorkspaceData.value = {}
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message || (isAdd ? '添加失败' : '更新失败'))
    }
  }

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
    loadAppOptions().catch(() => {
      appList.value = []
    })
  })

  watch(
    () => targetAppKey.value,
    (value) => {
      selectedAppKey.value = value || ''
      Object.assign(searchParams, { appKey: value || '' })
      refreshData()
    }
  )
</script>

<style lang="scss" scoped>
  .collaboration-workspace-hero-actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .collaboration-workspace-toolbar-tip {
    color: var(--art-text-muted);
    font-size: 13px;
    line-height: 1.6;
  }

  .collaboration-workspace-app-select {
    width: 240px;
  }
</style>
