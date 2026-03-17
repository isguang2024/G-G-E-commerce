<template>
  <div class="art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <ElButton v-action="'permission_action:create'" @click="openDialog('add')" v-ripple>
            新增功能权限
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
      @success="refreshData"
    />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchDeletePermissionAction,
    fetchGetPermissionActionList
  } from '@/api/system-manage'
  import ActionPermissionDialog from './modules/action-permission-dialog.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'ActionPermission' })

  type PermissionActionItem = Api.SystemManage.PermissionActionItem

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentAction = ref<PermissionActionItem | undefined>()

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
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
        { prop: 'name', label: '权限名称', minWidth: 160 },
        { prop: 'resourceCode', label: '资源编码', minWidth: 140 },
        { prop: 'actionCode', label: '动作编码', minWidth: 160 },
        {
          prop: 'scopeName',
          label: '作用域',
          width: 90,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.scopeCode === 'team' ? 'success' : 'primary' }, () =>
              row.scopeName || (row.scopeCode === 'team' ? '团队' : '平台')
            )
        },
        {
          prop: 'requiresTenantContext',
          label: '依赖团队',
          width: 100,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.requiresTenantContext ? 'warning' : 'info' }, () =>
              row.requiresTenantContext ? '是' : '否'
            )
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
          formatter: (row: PermissionActionItem) =>
            h(ArtButtonMore, {
              list: [
                {
                  key: 'edit',
                  label: '编辑',
                  icon: 'ri:edit-2-line',
                  auth: 'permission_action:update'
                },
                {
                  key: 'delete',
                  label: '删除',
                  icon: 'ri:delete-bin-4-line',
                  auth: 'permission_action:delete'
                }
              ],
              onClick: (item: ButtonMoreItem) => handleAction(item.key as string, row)
            })
        }
      ]
    }
  })

  function openDialog(type: 'add' | 'edit', row?: PermissionActionItem) {
    dialogType.value = type
    currentAction.value = row
    dialogVisible.value = true
  }

  function handleAction(command: string, row: PermissionActionItem) {
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
      .then(() => {
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
      })
  }
</script>
