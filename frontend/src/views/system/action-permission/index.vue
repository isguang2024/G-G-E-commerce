<template>
  <div class="permission-page art-full-height">
    <div class="page-top-stack">
      <ActionPermissionSearch
        v-show="showSearchBar"
        v-model="searchForm"
        :module-group-options="moduleGroupOptions"
        :feature-group-options="featureGroupOptions"
        @search="handleSearch"
        @reset="handleReset"
      />

      <AdminWorkspaceHero
        title="功能权限"
        description="统一检查权限键被 API、页面和功能包消费的情况，并筛出跨空间镜像与疑似重复键。"
        :metrics="summaryMetrics"
      >
        <div class="permission-hero-actions">
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
          <ElButton
            v-action="'system.permission.manage'"
            plain
            type="danger"
            :loading="cleaningUnused"
            @click="handleCleanupUnused"
            v-ripple
          >
            清理未消费自定义权限
          </ElButton>
        </div>
      </AdminWorkspaceHero>
    </div>

    <ElCard class="art-table-card permission-table-card" shadow="never">
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="handleRefresh"
      >
        <template #left>
          <div class="permission-toolbar-tip">
            优先处理“未被消费”和“疑似重复”，跨空间镜像键保留独立授权。
          </div>
        </template>
      </ArtTableHeader>

      <div class="permission-table-wrap">
        <ArtTable
          :loading="loading"
          :data="data"
          :columns="columns"
          :pagination="pagination"
          @pagination:size-change="handleSizeChange"
          @pagination:current-change="handleCurrentChange"
        />
      </div>
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

    <ElDialog v-model="consumerDialogVisible" title="权限消费明细" width="980px" destroy-on-close>
      <div v-loading="consumerLoading" class="consumer-dialog">
        <ElDescriptions :column="1" border size="small">
          <ElDescriptionsItem label="权限键">
            {{ consumerDetail.permissionKey || '-' }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <div class="consumer-section">
          <h4>API 消费（{{ consumerDetail.apis.length }}）</h4>
          <ElTable :data="consumerDetail.apis" size="small" border max-height="180">
            <ElTableColumn prop="method" label="Method" width="90" />
            <ElTableColumn prop="path" label="路径" min-width="280" />
            <ElTableColumn prop="summary" label="说明" min-width="220" show-overflow-tooltip />
          </ElTable>
        </div>

        <div class="consumer-section">
          <h4>页面消费（{{ consumerDetail.pages.length }}）</h4>
          <ElTable :data="consumerDetail.pages" size="small" border max-height="180">
            <ElTableColumn prop="pageKey" label="页面Key" min-width="180" />
            <ElTableColumn prop="name" label="页面名称" min-width="140" />
            <ElTableColumn prop="routePath" label="路由" min-width="220" />
            <ElTableColumn prop="accessMode" label="访问模式" width="100" />
          </ElTable>
        </div>

        <div class="consumer-section">
          <h4>功能包消费（{{ consumerDetail.featurePackages.length }}）</h4>
          <ElTable :data="consumerDetail.featurePackages" size="small" border max-height="180">
            <ElTableColumn prop="packageKey" label="包编码" min-width="180" />
            <ElTableColumn prop="name" label="包名称" min-width="140" />
            <ElTableColumn prop="packageType" label="类型" width="100" />
            <ElTableColumn prop="contextType" label="空间范围" width="100" />
          </ElTable>
        </div>

        <div class="consumer-section">
          <h4>角色引用（{{ consumerDetail.roles.length }}）</h4>
          <ElTable :data="consumerDetail.roles" size="small" border max-height="180">
            <ElTableColumn prop="code" label="角色编码" min-width="160" />
            <ElTableColumn prop="name" label="角色名称" min-width="150" />
            <ElTableColumn prop="contextType" label="空间范围" width="100" />
          </ElTable>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, reactive, ref } from 'vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetPermissionActionImpactPreview,
    fetchCleanupUnusedPermissionActions,
    fetchUpdatePermissionAction,
    fetchDeletePermissionAction,
    fetchGetPermissionActionConsumers,
    fetchGetPermissionActionList,
    fetchGetPermissionGroupList
  } from '@/api/system-manage'
  import ActionPermissionDialog from './modules/action-permission-dialog.vue'
  import ActionPermissionEndpointsDialog from './modules/action-permission-endpoints-dialog.vue'
  import ActionPermissionSearch from './modules/action-permission-search.vue'
  import PermissionGroupDialog from './modules/permission-group-dialog.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'ActionPermission' })

  type PermissionActionItem = Api.SystemManage.PermissionActionItem
  type PermissionGroupItem = Api.SystemManage.PermissionGroupItem
  type PermissionActionAuditSummary = Api.SystemManage.PermissionActionAuditSummary
  type PermissionActionListResponse = Api.SystemManage.PermissionActionList

  const dialogVisible = ref(false)
  const endpointDialogVisible = ref(false)
  const groupDialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const groupDialogType = ref<'module' | 'feature'>('module')
  const showSearchBar = ref(false)
  const currentAction = ref<PermissionActionItem>()
  const currentGroup = ref<PermissionGroupItem>()
  const moduleGroups = ref<PermissionGroupItem[]>([])
  const featureGroups = ref<PermissionGroupItem[]>([])
  const auditSummary = ref<PermissionActionAuditSummary>(createEmptyAuditSummary())
  const cleaningUnused = ref(false)
  const consumerDialogVisible = ref(false)
  const consumerLoading = ref(false)
  const consumerDetail = ref<Api.SystemManage.PermissionActionConsumerDetails>({
    permissionKey: '',
    apis: [],
    pages: [],
    featurePackages: [],
    roles: []
  })

  const searchForm = reactive({
    keyword: '',
    moduleGroupId: '',
    featureGroupId: '',
    status: '',
    isBuiltin: '',
    usagePattern: '',
    duplicatePattern: ''
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

  function isGroupSuspended(row: PermissionActionItem) {
    return row.moduleGroup?.status === 'suspended' || row.featureGroup?.status === 'suspended'
  }

  function createEmptyAuditSummary(): PermissionActionAuditSummary {
    return {
      totalCount: 0,
      unusedCount: 0,
      apiOnlyCount: 0,
      pageOnlyCount: 0,
      packageOnlyCount: 0,
      multiConsumerCount: 0,
      crossContextMirrorCount: 0,
      suspectedDuplicateCount: 0
    }
  }

  const summaryMetrics = computed(() => [
    { label: '当前页', value: data.value.length || 0 },
    { label: '当前筛选', value: auditSummary.value.totalCount || pagination.total || 0 },
    { label: '未被消费', value: auditSummary.value.unusedCount || 0 },
    { label: '多方复用', value: auditSummary.value.multiConsumerCount || 0 },
    { label: '跨空间镜像', value: auditSummary.value.crossContextMirrorCount || 0 },
    { label: '疑似重复', value: auditSummary.value.suspectedDuplicateCount || 0 }
  ])

  function renderConsumerCountTag(
    label: string,
    count: number | undefined,
    type: 'success' | 'primary' | 'warning'
  ) {
    const value = Number(count || 0)
    return h(
      ElTag,
      { type: value > 0 ? type : 'info', effect: 'plain', size: 'small' },
      () => `${label} ${value}`
    )
  }

  function renderUsageCell(row: PermissionActionItem) {
    return h('div', { class: 'permission-audit-cell' }, [
      h('div', { class: 'permission-audit-cell__tags' }, [
        renderConsumerCountTag('API', row.apiCount, 'success'),
        renderConsumerCountTag('页面', row.pageCount, 'primary'),
        renderConsumerCountTag('功能包', row.packageCount, 'warning')
      ]),
      h('div', { class: 'permission-audit-cell__note' }, row.usageNote || '-'),
      h(
        ElButton,
        {
          text: true,
          type: 'primary',
          size: 'small',
          onClick: () => openConsumerDialog(row)
        },
        () => '查看消费明细'
      )
    ])
  }

  function renderDuplicateCell(row: PermissionActionItem) {
    if (!row.duplicatePattern || row.duplicatePattern === 'none') {
      return h('div', { class: 'permission-audit-cell' }, [
        h('div', { class: 'permission-audit-cell__note' }, '-')
      ])
    }
    const tagType = row.duplicatePattern === 'cross_context_mirror' ? 'warning' : 'danger'
    const tagLabel = row.duplicatePattern === 'cross_context_mirror' ? '跨空间镜像' : '疑似重复'
    const relatedKeys = row.duplicateKeys || []
    return h('div', { class: 'permission-audit-cell' }, [
      h('div', { class: 'permission-audit-cell__tags' }, [
        h(ElTag, { type: tagType, effect: 'plain', size: 'small' }, () => tagLabel),
        ...relatedKeys
          .slice(0, 2)
          .map((key) => h(ElTag, { type: 'info', effect: 'plain', size: 'small' }, () => key))
      ]),
      h(
        'div',
        { class: 'permission-audit-cell__note' },
        row.duplicateNote || relatedKeys.join(' / ') || '-'
      )
    ])
  }

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
      immediate: false,
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
          formatter: (row: PermissionActionItem) => row.moduleGroup?.name || '-'
        },
        {
          prop: 'featureGroup',
          label: '功能分组',
          minWidth: 140,
          formatter: (row: PermissionActionItem) => row.featureGroup?.name || '-'
        },
        {
          prop: 'usagePattern',
          label: '消费结构',
          minWidth: 250,
          formatter: (row: PermissionActionItem) => renderUsageCell(row)
        },
        {
          prop: 'duplicatePattern',
          label: '重复检查',
          minWidth: 260,
          formatter: (row: PermissionActionItem) => renderDuplicateCell(row)
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
          minWidth: 180,
          showOverflowTooltip: true,
          formatter: (row: PermissionActionItem) => row.description || '-'
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () => {
              if (row.status === 'normal') return '正常'
              return isGroupSuspended(row) ? '停用(分组)' : '停用'
            })
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
            if (!isGroupSuspended(row)) {
              list.push({
                key: row.status === 'normal' ? 'disable' : 'enable',
                label: row.status === 'normal' ? '停用' : '启用',
                icon: row.status === 'normal' ? 'ri:forbid-2-line' : 'ri:check-line',
                auth: 'system.permission.manage'
              })
            }
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleAction(item.key as string, row)
            })
          }
        }
      ]
    },
    transform: {
      responseAdapter: (response: PermissionActionListResponse) => ({
        records: response?.records || [],
        total: response?.total || 0,
        current: response?.current,
        size: response?.size,
        auditSummary: response?.auditSummary || createEmptyAuditSummary()
      })
    },
    hooks: {
      onSuccess: (_rows, response: any) => {
        auditSummary.value = response?.auditSummary || createEmptyAuditSummary()
      }
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
      status: searchForm.status || undefined,
      isBuiltin: searchForm.isBuiltin === '' ? undefined : searchForm.isBuiltin === 'true',
      usagePattern: searchForm.usagePattern || undefined,
      duplicatePattern: searchForm.duplicatePattern || undefined,
      current: 1
    })
    await getData()
  }

  async function handleReset() {
    Object.assign(searchForm, {
      keyword: '',
      moduleGroupId: '',
      featureGroupId: '',
      status: '',
      isBuiltin: '',
      usagePattern: '',
      duplicatePattern: ''
    })
    await handleSearch()
  }

  async function handleRefresh() {
    await Promise.all([refreshData(), loadGroups()])
  }

  async function handleCleanupUnused() {
    ElMessageBox.confirm(
      '将删除所有“未被 API、页面、功能包消费”且“非内置”的功能权限，是否继续？',
      '清理确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
      .then(async () => {
        cleaningUnused.value = true
        const result = await fetchCleanupUnusedPermissionActions()
        if (result.deletedCount > 0) {
          ElMessage.success(`已清理 ${result.deletedCount} 个未消费自定义权限`)
        } else {
          ElMessage.info('没有可清理的未消费自定义权限')
        }
        await handleRefresh()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '清理失败')
      })
      .finally(() => {
        cleaningUnused.value = false
      })
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

  async function openConsumerDialog(row: PermissionActionItem) {
    if (!row.id) return
    consumerDialogVisible.value = true
    consumerLoading.value = true
    try {
      consumerDetail.value = await fetchGetPermissionActionConsumers(row.id)
    } catch (error: any) {
      ElMessage.error(error?.message || '获取消费明细失败')
    } finally {
      consumerLoading.value = false
    }
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
      fetchGetPermissionActionImpactPreview(row.id)
        .then((impact) =>
          ElMessageBox.confirm(
            `${actionText}影响：API ${impact.apiCount}、页面 ${impact.pageCount}、功能包 ${impact.packageCount}、角色 ${impact.roleCount}、协作空间 ${impact.collaborationWorkspaceCount}、用户 ${impact.userCount}。确定继续？`,
            `${actionText}确认`,
            {
              confirmButtonText: '确定',
              cancelButtonText: '取消',
              type: 'warning'
            }
          )
        )
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
    fetchGetPermissionActionImpactPreview(row.id)
      .then((impact) =>
        ElMessageBox.confirm(
          `删除影响：API ${impact.apiCount}、页面 ${impact.pageCount}、功能包 ${impact.packageCount}、角色 ${impact.roleCount}、协作空间 ${impact.collaborationWorkspaceCount}、用户 ${impact.userCount}。确定删除「${row.name}」？`,
          '删除确认',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
      )
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
  .permission-page {
    display: flex;
    min-height: 0;
    flex-direction: column;
  }

  .permission-table-card {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
  }

  .permission-table-card :deep(.el-card__body) {
    display: flex;
    min-height: 0;
    flex: 1;
    flex-direction: column;
    overflow: hidden;
  }

  .permission-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .permission-toolbar-tip {
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }

  .permission-audit-cell {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .permission-audit-cell__tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .permission-audit-cell__note {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.5;
  }

  .permission-table-wrap {
    min-height: 0;
    flex: 1;
    overflow: auto;
  }

  .consumer-dialog {
    display: grid;
    gap: 12px;
  }

  .consumer-section {
    display: grid;
    gap: 6px;
  }

  .consumer-section h4 {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
  }

  .art-table-card :deep(.table-header-left) {
    gap: 12px;
    row-gap: 8px;
  }
</style>
