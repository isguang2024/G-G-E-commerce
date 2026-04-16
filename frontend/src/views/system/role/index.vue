<template>
  <div class="art-full-height">
    <div class="page-top-stack">
      <RoleSearch
        v-show="showSearchBar"
        v-model="searchForm"
        @search="handleSearch"
        @reset="handleResetSearch"
      />

      <AdminWorkspaceHero
        title="个人空间角色管理"
        description="角色是全局角色目录；未配置 App 生效范围时全局通用，配置后仅在指定 App 下参与权限裁剪。"
        :metrics="heroMetrics"
      >
        <div class="role-hero-actions">
          <AppKeySelect
            v-model="selectedAppKey"
            placeholder="选择 App"
            class="role-app-select"
            @change="handleManagedAppChange"
          />
          <ElButton
            v-action="'system.role.manage'"
            type="primary"
            @click="showDialog('add')"
            v-ripple
          >
            新增角色
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
          <div class="role-toolbar-tip"
            >角色目录本身不依赖 App 才能查看；功能包、权限和菜单裁剪仍按具体 App 配置。</div
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
    </ElCard>

    <RoleEditDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :role-data="currentRoleData"
      @success="refreshData"
    />

    <RolePermissionDialog
      v-model="permissionDialog"
      :role-data="currentRoleData"
      :app-key="targetAppKey"
      @success="handlePermissionSuccess"
    />

    <RolePackageDialog
      v-model="packageDialog"
      :role-data="currentRoleData"
      :app-key="targetAppKey"
      @success="handlePermissionSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useAuth } from '@/hooks/core/useAuth'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchDeleteRole, fetchGetRoleList } from '@/domains/governance/api'
  import { useManagedAppScope } from '@/domains/app-runtime/useManagedAppScope'
  import { refreshUserMenus } from '@/domains/navigation/runtime/navigation'
  import AppKeySelect from '@/components/business/app/AppKeySelect.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import RoleSearch from './modules/role-search.vue'
  import RoleEditDialog from './modules/role-edit-dialog.vue'
  import RolePackageDialog from './modules/role-package-dialog.vue'
  import RolePermissionDialog from './modules/role-permission-selector-dialog.vue'

  defineOptions({ name: 'Role' })

  type RoleListItem = Api.SystemManage.RoleListItem

  const { hasAction } = useAuth()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()

  const showSearchBar = ref(false)
  const selectedAppKey = ref('')
  const dialogVisible = ref(false)
  const permissionDialog = ref(false)
  const packageDialog = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentRoleData = ref<RoleListItem | undefined>()
  const heroMetrics = computed(() => [
    { label: 'App 过滤', value: targetAppKey.value || '全部 App' },
    { label: '角色总数', value: pagination.total || data.value.length || 0 },
    { label: '当前页', value: data.value.length || 0 },
    { label: '全局通用', value: data.value.filter((item) => item.isGlobal).length || 0 }
  ])

  const searchForm = ref({
    roleName: undefined,
    roleCode: undefined,
    description: undefined,
    enabled: undefined,
    daterange: undefined
  })

  const fetchRoleListByApp: typeof fetchGetRoleList = async (params) =>
    fetchGetRoleList({
      ...(params || {}),
      ...(targetAppKey.value ? { appKey: targetAppKey.value } : {})
    } as Api.SystemManage.RoleSearchParams)

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
      apiFn: fetchRoleListByApp,
      apiParams: {
        current: 1,
        size: 20
      },
      excludeParams: ['daterange'],
      columnsFactory: () => [
        {
          prop: 'roleId',
          label: '角色 ID',
          minWidth: 260,
          showOverflowTooltip: true
        },
        {
          prop: 'roleName',
          label: '角色名称',
          minWidth: 140
        },
        {
          prop: 'roleCode',
          label: '角色编码',
          minWidth: 140
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'appKeys',
          label: '生效范围',
          minWidth: 220,
          formatter: (row: RoleListItem) => {
            if (row.isGlobal || !row.appKeys?.length) {
              return h(ElTag, { type: 'success', effect: 'plain' }, () => '全局通用')
            }
            return h(
              'div',
              { class: 'flex flex-wrap gap-1' },
              row.appKeys.map((appKey) =>
                h(
                  ElTag,
                  {
                    key: appKey,
                    size: 'small',
                    type: 'info',
                    effect: 'plain'
                  },
                  () => appKey
                )
              )
            )
          }
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: RoleListItem) => {
            const status =
              row.status === 'normal'
                ? { type: 'success', text: '正常' }
                : { type: 'warning', text: '停用' }
            return h(ElTag, { type: status.type as 'success' | 'warning' }, () => status.text)
          }
        },
        {
          prop: 'priority',
          label: '优先级',
          width: 90,
          formatter: (row: RoleListItem) => row.priority || 0
        },
        {
          prop: 'createTime',
          label: '创建时间',
          width: 180,
          sortable: true
        },
        {
          prop: 'operation',
          label: '操作',
          width: 80,
          fixed: 'right',
          formatter: (row: RoleListItem) => {
            const isDefaultRole = [
              'admin',
              'collaboration_workspace_admin',
              'collaboration_workspace_member'
            ].includes(row.roleCode)
            const list: ButtonMoreItem[] = [
              {
                key: 'packages',
                label: '功能包',
                icon: 'ri:apps-2-line',
                auth: 'feature_package.assign_collaboration_workspace'
              },
              {
                key: 'permission',
                label: '权限配置',
                icon: 'ri:shield-keyhole-line',
                auth: 'system.role.assign'
              },
              {
                key: 'edit',
                label: '编辑角色',
                icon: 'ri:edit-2-line',
                auth: 'system.role.manage'
              }
            ]
            if (!isDefaultRole && hasAction('system.role.manage')) {
              list.push({
                key: 'delete',
                label: '删除角色',
                icon: 'ri:delete-bin-4-line',
                auth: 'system.role.manage'
              })
            }

            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleActionClick(item, row)
            })
          }
        }
      ]
    }
  })

  function showDialog(type: 'add' | 'edit', row?: RoleListItem) {
    dialogType.value = type
    currentRoleData.value = row
    dialogVisible.value = true
  }

  function showPermissionDialog(row?: RoleListItem) {
    currentRoleData.value = row
    permissionDialog.value = true
  }

  function showPackageDialog(row?: RoleListItem) {
    currentRoleData.value = row
    packageDialog.value = true
  }

  function handleSearch() {
    const { daterange, ...rest } = searchForm.value
    const [startTime, endTime] = Array.isArray(daterange) ? daterange : [null, null]
    Object.assign(searchParams, { ...rest, startTime, endTime })
    getData()
  }

  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
  }

  function handleResetSearch() {
    resetSearchParams()
    Object.assign(searchParams, { current: 1, size: pagination.size })
    getData()
  }

  function handleActionClick(item: ButtonMoreItem, row: RoleListItem) {
    if (item.key === 'packages') {
      if (!targetAppKey.value) {
        ElMessage.warning('请先选择一个 App，再配置角色功能包范围')
        return
      }
      showPackageDialog(row)
      return
    }

    if (item.key === 'permission') {
      if (!targetAppKey.value) {
        ElMessage.warning('请先选择一个 App，再配置角色权限与菜单范围')
        return
      }
      showPermissionDialog(row)
      return
    }

    if (item.key === 'edit') {
      showDialog('edit', row)
      return
    }

    if (item.key === 'delete') {
      deleteRole(row)
    }
  }

  async function handlePermissionSuccess() {
    refreshData()
    await refreshUserMenus()
  }

  function deleteRole(row: RoleListItem) {
    ElMessageBox.confirm(`确定删除角色“${row.roleName}”吗？此操作不可恢复。`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeleteRole(row.roleId))
      .then(() => {
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((error) => {
        if (error !== 'cancel') {
          ElMessage.error(error instanceof Error ? error.message : '删除失败')
        }
      })
  }

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
  })

  watch(
    () => targetAppKey.value,
    (value) => {
      selectedAppKey.value = value || ''
      refreshData()
    }
  )
</script>

<style scoped lang="scss">
  .role-hero-actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .role-toolbar-tip {
    font-size: 13px;
    color: var(--art-text-muted);
  }

  .role-app-select {
    width: 240px;
  }
</style>
