<template>
  <div class="team-page art-full-height">
    <div class="page-top-stack">
      <TeamSearch
        v-show="showSearchBar"
        v-model="searchForm"
        :showExpand="true"
        @search="handleSearch"
        @reset="handleResetSearch"
      />

      <AdminWorkspaceHero :title="'团队管理'" :description="'统一管理团队边界、管理员与授权入口。'" :metrics="heroMetrics">
        <div class="team-hero-actions">
          <ElSelect
            v-model="selectedAppKey"
            clearable
            filterable
            placeholder="选择 App"
            class="team-app-select"
            @change="handleManagedAppChange"
          >
            <ElOption
              v-for="item in appOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
          <ElButton v-action="'tenant.manage'" type="primary" @click="showDialog('add')" v-ripple>
            新增团队
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
          <div class="team-toolbar-tip">团队管理同步菜单与成员边界，建议优先从主账号确认管理员后再开通功能包。</div>
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

      <TeamDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :team-data="currentTeamData"
        @submit="handleDialogSubmit"
      />

      <TeamMembersDrawer
        v-model:visible="membersDrawerVisible"
        :team-id="currentTeamId"
        :team-name="currentTeamName"
        @refresh="refreshData"
      />

      <TeamActionPermissionDialog
        v-model="actionDialogVisible"
        :team-id="currentTeamId"
        :team-name="currentTeamName"
        :app-key="targetAppKey"
        @success="refreshData"
      />

      <TeamMenuPermissionDialog
        v-model="menuDialogVisible"
        :team-id="currentTeamId"
        :team-name="currentTeamName"
        :app-key="targetAppKey"
        @success="refreshData"
      />

      <TeamFeaturePackageDialog
        v-model="packageDialogVisible"
        :team-id="currentTeamId"
        :team-name="currentTeamName"
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
  import { fetchGetTeamList, fetchDeleteTeam, fetchCreateTeam, fetchUpdateTeam } from '@/api/team'
  import { fetchGetApps } from '@/api/system-manage'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import TeamSearch from './modules/team-search.vue'
  import TeamDialog from './modules/team-dialog.vue'
  import TeamMembersDrawer from './modules/team-members-drawer.vue'
  import TeamActionPermissionDialog from './modules/team-permission-dialog.vue'
  import TeamMenuPermissionDialog from './modules/team-menu-permission-dialog.vue'
  import TeamFeaturePackageDialog from './modules/team-feature-package-dialog.vue'
  import {
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

  defineOptions({ name: 'Team' })

  type TeamListItem = Api.SystemManage.TeamListItem
  const { hasAction } = useAuth()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()

  const dialogType = ref<DialogType>('add')
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const dialogVisible = ref(false)
  const currentTeamData = ref<Partial<TeamListItem>>({})
  const showSearchBar = ref(false)
  const membersDrawerVisible = ref(false)
  const actionDialogVisible = ref(false)
  const menuDialogVisible = ref(false)
  const packageDialogVisible = ref(false)
  const currentTeamId = ref('')
  const currentTeamName = ref('')

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
    { label: '团队总数', value: data.value.length },
    {
      label: '活跃团队',
      value: data.value.filter((item) => item.status === 'active').length
    },
    {
      label: '停用团队',
      value: data.value.filter((item) => item.status === 'inactive').length
    },
    {
      label: '管理员覆盖',
      value: data.value.reduce(
        (total, item) => total + (((item as any).adminUsers || (item as any).admin_users || []).length || 0),
        0
      )
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
            current: Number((params as any)?.current || 1),
            size: Number((params as any)?.size || 20)
          }
        }
        return fetchGetTeamList(params as Api.SystemManage.TeamSearchParams)
      },
      apiParams: {
        current: 1,
        size: 20,
        appKey: targetAppKey.value,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        { prop: 'name', label: '团队名称', minWidth: 140 },
        { prop: 'remark', label: '团队备注', width: 140 },
        {
          prop: 'adminUsers',
          label: '管理员',
          width: 200,
          formatter: (row: TeamListItem) => {
            // 兼容 adminUsers 和 admin_users 两种字段名
            const admins = (row as any).adminUsers || (row as any).admin_users || []
            if (!admins.length) return '-'
            return h('div', { class: 'flex flex-wrap gap-1' }, [
              ...admins.slice(0, 2).map((admin: any) =>
                h(ElTag, { size: 'small', type: 'info' }, () => {
                  return (
                    admin.nickname ||
                    admin.username ||
                    admin.email ||
                    admin.user_id.substring(0, 8) + '...'
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
                    hasAction('tenant.manage')
                      ? h(ElDropdownItem, { command: 'edit' }, () => [
                          h(ElIcon, {}, () => h(Edit)),
                          '编辑'
                        ])
                      : null,
                    hasAction('tenant.manage')
                      ? h(ElDropdownItem, { command: 'view' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '查看人员'
                        ])
                      : null,
                    hasAction('tenant.manage')
                      ? h(ElDropdownItem, { command: 'menu' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '菜单边界'
                        ])
                      : null,
                    hasAction('tenant.manage')
                      ? h(ElDropdownItem, { command: 'action' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '团队边界'
                        ])
                      : null,
                    hasAction('platform.package.assign')
                      ? h(ElDropdownItem, { command: 'package' }, () => [
                          h(ElIcon, {}, () => h(UserFilled)),
                          '开通功能包'
                        ])
                      : null,
                    hasAction('tenant.manage')
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
    Object.assign(searchParams, { current: 1, size: pagination.size, appKey: targetAppKey.value || '' })
    getData()
  }

  const showDialog = (type: DialogType, row?: TeamListItem) => {
    dialogType.value = type
    currentTeamData.value = row ? { ...row } : {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  const handleOperationCommand = (command: string, row: TeamListItem) => {
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
      deleteTeam(row)
    }
  }

  const showMembers = (row: TeamListItem) => {
    currentTeamId.value = row.id
    currentTeamName.value = row.name
    membersDrawerVisible.value = true
  }

  const showActionPermissions = (row: TeamListItem) => {
    currentTeamId.value = row.id
    currentTeamName.value = row.name
    actionDialogVisible.value = true
  }

  const showMenuPermissions = (row: TeamListItem) => {
    currentTeamId.value = row.id
    currentTeamName.value = row.name
    menuDialogVisible.value = true
  }

  const showFeaturePackages = (row: TeamListItem) => {
    currentTeamId.value = row.id
    currentTeamName.value = row.name
    packageDialogVisible.value = true
  }

  const deleteTeam = (row: TeamListItem) => {
    ElMessageBox.confirm(`确定要删除团队「${row.name}」吗？`, '删除团队', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeleteTeam(row.id))
      .then(() => {
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
      })
  }

  const handleDialogSubmit = async (
    payload?: Api.SystemManage.TeamCreateParams | Api.SystemManage.TeamUpdateParams
  ) => {
    if (!payload) {
      dialogVisible.value = false
      refreshData()
      return
    }
    const isAdd = dialogType.value === 'add'
    try {
      if (isAdd) {
        await fetchCreateTeam(payload as Api.SystemManage.TeamCreateParams)
        ElMessage.success('添加成功')
      } else {
        const id = (currentTeamData.value as TeamListItem).id
        if (!id) return
        await fetchUpdateTeam(id, payload as Api.SystemManage.TeamUpdateParams)
        ElMessage.success('更新成功')
      }
      dialogVisible.value = false
      currentTeamData.value = {}
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
  .team-hero-actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .team-toolbar-tip {
    color: var(--art-text-muted);
    font-size: 13px;
    line-height: 1.6;
  }

  .team-app-select {
    width: 240px;
  }
</style>

