<template>
  <div class="art-full-height">
    <ActionPermissionSearch
      v-show="showSearchBar"
      v-model="searchForm"
      :module-group-options="moduleGroupOptions"
      :feature-group-options="featureGroupOptions"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard
      class="art-table-card"
      shadow="never"
      :style="{ marginTop: showSearchBar ? '12px' : '0' }"
    >
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="handleRefresh"
      >
        <template #left>
          <ElButton
            v-action="'system.permission.manage'"
            type="primary"
            @click="openDialog('add')"
            v-ripple
          >
            新增功能权限
          </ElButton>
          <ElButton
            v-action="'system.permission.manage'"
            @click="openGroupDialog('module')"
            v-ripple
          >
            管理模块分组
          </ElButton>
          <ElButton
            v-action="'system.permission.manage'"
            type="primary"
            plain
            @click="openGroupDialog('feature')"
            v-ripple
          >
            管理功能分组
          </ElButton>
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

    <ActionPermissionDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :action-data="currentAction"
      :module-groups="moduleGroups"
      :feature-groups="featureGroups"
      @open-group="openGroupDialog"
      @success="handlePermissionSaved"
    />

    <PermissionGroupDialog
      v-model="groupDialogVisible"
      :group-type="groupDialogType"
      :group-data="currentGroup"
      @success="handleGroupSaved"
    />

    <ActionPermissionEndpointsDialog
      v-model="endpointDialogVisible"
      :permission-id="currentAction?.id || ''"
      :permission-name="currentAction?.name || ''"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, h, reactive, ref } from 'vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchUpdatePermissionAction,
    fetchDeletePermissionAction,
    fetchGetPermissionActionList,
    fetchGetPermissionGroupList
  } from '@/api/system-manage'
  import ActionPermissionDialog from './modules/action-permission-dialog.vue'
  import ActionPermissionEndpointsDialog from './modules/action-permission-endpoints-dialog.vue'
  import ActionPermissionSearch from './modules/action-permission-search.vue'
  import PermissionGroupDialog from './modules/permission-group-dialog.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'ActionPermission' })

  type PermissionActionItem = Api.SystemManage.PermissionActionItem
  type PermissionGroupItem = Api.SystemManage.PermissionGroupItem

  const dialogVisible = ref(false)
  const endpointDialogVisible = ref(false)
  const groupDialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const groupDialogType = ref<'module' | 'feature'>('module')
  const showSearchBar = ref(true)
  const currentAction = ref<PermissionActionItem>()
  const currentGroup = ref<PermissionGroupItem>()
  const moduleGroups = ref<PermissionGroupItem[]>([])
  const featureGroups = ref<PermissionGroupItem[]>([])

  const searchForm = reactive({
    keyword: '',
    moduleGroupId: '',
    featureGroupId: '',
    contextType: '',
    status: '',
    isBuiltin: ''
  })

  const moduleGroupOptions = computed(() =>
    moduleGroups.value.map((item) => ({
      label: item.name,
      value: item.id
    }))
  )

  const featureGroupOptions = computed(() =>
    featureGroups.value.map((item) => ({
      label: item.name,
      value: item.id
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
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetPermissionActionList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        { prop: 'name', label: '权限名称', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'permissionKey',
          label: '权限键',
          minWidth: 220,
          formatter: (row: PermissionActionItem) => row.permissionKey || '-'
        },
        {
          prop: 'moduleGroup',
          label: '模块分组',
          minWidth: 140,
          formatter: (row: PermissionActionItem) => row.moduleGroup?.name || row.moduleCode || '-'
        },
        {
          prop: 'featureGroup',
          label: '功能分组',
          minWidth: 140,
          formatter: (row: PermissionActionItem) => row.featureGroup?.name || row.featureKind || '-'
        },
        {
          prop: 'contextType',
          label: '上下文',
          width: 100,
          formatter: (row: PermissionActionItem) => {
            if (row.contextType === 'platform') return h(ElTag, { type: 'warning' }, () => '平台')
            if (row.contextType === 'team') return h(ElTag, { type: 'primary' }, () => '团队')
            return h(ElTag, { type: 'info' }, () => '通用')
          }
        },
        {
          prop: 'isBuiltin',
          label: '内置',
          width: 90,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.isBuiltin ? 'success' : 'info', effect: 'plain' }, () =>
              row.isBuiltin ? '是' : '否'
            )
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 220,
          showOverflowTooltip: true,
          formatter: (row: PermissionActionItem) => row.description || '-'
        },
        { prop: 'sortOrder', label: '排序', width: 80 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 70,
          fixed: 'right',
          formatter: (row: PermissionActionItem) => {
            const list: ButtonMoreItem[] = [
              {
                key: 'view-apis',
                label: '查看接口',
                icon: 'ri:links-line',
                auth: 'system.api_registry.view'
              },
              {
                key: 'edit',
                label: '编辑',
                icon: 'ri:edit-2-line',
                auth: 'system.permission.manage'
              }
            ]
            if (!row.isBuiltin) {
              list.push({
                key: 'delete',
                label: '删除',
                icon: 'ri:delete-bin-4-line',
                auth: 'system.permission.manage'
              })
            }
            list.push({
              key: row.status === 'normal' ? 'disable' : 'enable',
              label: row.status === 'normal' ? '停用' : '启用',
              icon: row.status === 'normal' ? 'ri:forbid-2-line' : 'ri:check-line',
              auth: 'system.permission.manage'
            })
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleAction(item.key as string, row)
            })
          }
        }
      ]
    }
  })

  async function loadGroups() {
    const [moduleRes, featureRes] = await Promise.all([
      fetchGetPermissionGroupList({ current: 1, size: 200, groupType: 'module', status: 'normal' }),
      fetchGetPermissionGroupList({ current: 1, size: 200, groupType: 'feature', status: 'normal' })
    ])
    moduleGroups.value = moduleRes.records || []
    featureGroups.value = featureRes.records || []
  }

  async function handleSearch() {
    Object.assign(searchParams, {
      keyword: searchForm.keyword || undefined,
      moduleGroupId: searchForm.moduleGroupId || undefined,
      featureGroupId: searchForm.featureGroupId || undefined,
      contextType: searchForm.contextType || undefined,
      status: searchForm.status || undefined,
      isBuiltin: searchForm.isBuiltin === '' ? undefined : searchForm.isBuiltin === 'true',
      current: 1
    })
    await getData()
  }

  async function handleReset() {
    Object.assign(searchForm, {
      keyword: '',
      moduleGroupId: '',
      featureGroupId: '',
      contextType: '',
      status: '',
      isBuiltin: ''
    })
    await handleSearch()
  }

  async function handleRefresh() {
    await Promise.all([refreshData(), loadGroups()])
  }

  function openDialog(type: 'add' | 'edit', row?: PermissionActionItem) {
    dialogType.value = type
    currentAction.value = row
    dialogVisible.value = true
  }

  function openGroupDialog(type: 'module' | 'feature', row?: PermissionGroupItem) {
    groupDialogType.value = type
    currentGroup.value = row
    groupDialogVisible.value = true
  }

  async function handlePermissionSaved() {
    await handleRefresh()
  }

  async function handleGroupSaved() {
    await loadGroups()
  }

  function handleAction(command: string, row: PermissionActionItem) {
    if (command === 'view-apis') {
      currentAction.value = row
      endpointDialogVisible.value = true
      return
    }
    if (command === 'edit') {
      openDialog('edit', row)
      return
    }
    if (command === 'disable' || command === 'enable') {
      const targetStatus = command === 'disable' ? 'suspended' : 'normal'
      const actionText = command === 'disable' ? '停用' : '启用'
      ElMessageBox.confirm(`确定${actionText}功能权限「${row.name}」吗？`, `${actionText}确认`, {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() => fetchUpdatePermissionAction(row.id, { status: targetStatus }))
        .then(async () => {
          ElMessage.success(`${actionText}成功`)
          await handleRefresh()
        })
        .catch((e) => {
          if (e !== 'cancel') ElMessage.error(e?.message || `${actionText}失败`)
        })
      return
    }
    ElMessageBox.confirm(`确定删除功能权限「${row.name}」吗？`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeletePermissionAction(row.id))
      .then(async () => {
        ElMessage.success('删除成功')
        await handleRefresh()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
      })
  }

  loadGroups().then(handleSearch)
</script>

<style scoped>
  .art-table-card :deep(.table-header-left) {
    gap: 10px;
    row-gap: 8px;
  }
</style>
