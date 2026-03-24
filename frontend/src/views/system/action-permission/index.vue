<template>
  <div class="art-full-height">
    <ElCard class="group-card" shadow="never">
      <div class="group-header">
        <div>
          <div class="group-title">功能权限分组</div>
          <div class="group-help">模块分组和功能分组统一维护，功能权限只绑定分组，不再使用来源筛选。</div>
        </div>
        <div class="group-actions">
          <ElButton v-action="'system.permission.manage'" @click="openGroupDialog('module')">新建模块分组</ElButton>
          <ElButton v-action="'system.permission.manage'" type="primary" @click="openGroupDialog('feature')">新建功能分组</ElButton>
        </div>
      </div>

      <div class="group-grid">
        <div class="group-panel">
          <div class="panel-title">模块分组</div>
          <div class="panel-tags">
            <ElTag
              v-for="item in moduleGroups"
              :key="item.id"
              :type="searchForm.moduleGroupId === item.id ? 'primary' : 'info'"
              effect="light"
              class="panel-tag"
              @click="handleGroupFilter('module', item.id)"
            >
              {{ item.name }}
            </ElTag>
          </div>
        </div>
        <div class="group-panel">
          <div class="panel-title">功能分组</div>
          <div class="panel-tags">
            <ElTag
              v-for="item in featureGroups"
              :key="item.id"
              :type="searchForm.featureGroupId === item.id ? 'primary' : 'success'"
              effect="light"
              class="panel-tag"
              @click="handleGroupFilter('feature', item.id)"
            >
              {{ item.name }}
            </ElTag>
          </div>
        </div>
      </div>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        @refresh="handleRefresh"
      >
        <template #left>
          <ElButton v-action="'system.permission.manage'" type="primary" @click="openDialog('add')" v-ripple>
            新增功能权限
          </ElButton>
        </template>
      </ArtTableHeader>

      <div class="query-row">
        <ElInput v-model="searchForm.keyword" clearable placeholder="名称/描述/权限键" />
        <ElSelect v-model="searchForm.moduleGroupId" clearable filterable placeholder="模块分组">
          <ElOption v-for="item in moduleGroups" :key="item.id" :label="item.name" :value="item.id" />
        </ElSelect>
        <ElSelect v-model="searchForm.featureGroupId" clearable filterable placeholder="功能分组">
          <ElOption v-for="item in featureGroups" :key="item.id" :label="item.name" :value="item.id" />
        </ElSelect>
        <ElSelect v-model="searchForm.contextType" clearable placeholder="上下文">
          <ElOption label="平台" value="platform" />
          <ElOption label="团队" value="team" />
          <ElOption label="通用" value="common" />
        </ElSelect>
        <ElSelect v-model="searchForm.status" clearable placeholder="状态">
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="suspended" />
        </ElSelect>
        <ElSelect v-model="searchForm.isBuiltin" clearable placeholder="是否内置">
          <ElOption label="内置" value="true" />
          <ElOption label="自定义" value="false" />
        </ElSelect>
        <div class="query-actions">
          <ElButton type="primary" @click="handleSearch">查询</ElButton>
          <ElButton @click="handleReset">重置</ElButton>
        </div>
      </div>

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
    fetchDeletePermissionAction,
    fetchGetPermissionActionList,
    fetchGetPermissionGroupList
  } from '@/api/system-manage'
  import ActionPermissionDialog from './modules/action-permission-dialog.vue'
  import ActionPermissionEndpointsDialog from './modules/action-permission-endpoints-dialog.vue'
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
              { key: 'view-apis', label: '查看接口', icon: 'ri:links-line', auth: 'system.api_registry.view' },
              { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'system.permission.manage' }
            ]
            if (!row.isBuiltin) {
              list.push({
                key: 'delete',
                label: '删除',
                icon: 'ri:delete-bin-4-line',
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
    }
  })

  const groupMap = computed(() => ({
    module: moduleGroups.value,
    feature: featureGroups.value
  }))

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

  async function handleGroupFilter(type: 'module' | 'feature', id: string) {
    if (type === 'module') {
      searchForm.moduleGroupId = searchForm.moduleGroupId === id ? '' : id
    } else {
      searchForm.featureGroupId = searchForm.featureGroupId === id ? '' : id
    }
    await handleSearch()
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
  .group-card {
    margin-bottom: 12px;
  }

  .group-header {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }

  .group-title {
    font-size: 14px;
    font-weight: 600;
  }

  .group-help {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 4px;
  }

  .group-actions {
    display: flex;
    gap: 8px;
  }

  .group-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .group-panel {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    padding: 12px;
  }

  .panel-title {
    font-size: 13px;
    font-weight: 600;
    margin-bottom: 10px;
  }

  .panel-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .panel-tag {
    cursor: pointer;
  }

  .query-row {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr)) auto;
    gap: 12px;
    margin-bottom: 12px;
  }

  .query-actions {
    display: flex;
    gap: 8px;
  }
</style>
