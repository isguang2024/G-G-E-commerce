<template>
  <div class="scope-page">
    <ElCard class="scope-card" shadow="never">
      <div class="scope-toolbar">
        <div class="scope-toolbar__title">作用域管理</div>
        <ElButton v-action="'scope:create'" type="primary" @click="showDialog('add')">
          新增作用域
        </ElButton>
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

    <ScopeEditDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :scope-data="currentScopeData"
      @success="refreshData"
    />
  </div>
</template>

<script setup lang="ts">
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchDeleteScope, fetchGetScopeList } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import ScopeEditDialog from '../role/modules/scope-edit-dialog.vue'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'ScopeManagePage' })

  type ScopeListItem = Api.SystemManage.ScopeListItem

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentScopeData = ref<ScopeListItem | undefined>(undefined)

  const {
    columns,
    data,
    loading,
    pagination,
    getData,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetScopeList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        { prop: 'scopeCode', label: '作用域编码', minWidth: 140 },
        { prop: 'scopeName', label: '作用域名称', minWidth: 140 },
        {
          prop: 'dataPermissionName',
          label: '数据权限范围',
          minWidth: 180,
          formatter: (row: ScopeListItem) => row.dataPermissionName || row.dataPermissionCode || '-'
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 220,
          showOverflowTooltip: true
        },
        { prop: 'sortOrder', label: '排序', width: 90 },
        { prop: 'createTime', label: '创建时间', width: 180 },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row: ScopeListItem) => {
            const list = [{ key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'scope:update' }]
            if (!isProtectedScope(row)) {
              list.push({
                key: 'delete',
                label: '删除',
                icon: 'ri:delete-bin-4-line',
                auth: 'scope:delete'
              })
            }
            return h('div', [
              h(ArtButtonMore, {
                list,
                onClick: (item: ButtonMoreItem) => buttonMoreClick(item, row)
              })
            ])
          }
        }
      ]
    }
  })

  function isProtectedScope(row: ScopeListItem) {
    return Boolean(row.isSystem)
  }

  const showDialog = (type: 'add' | 'edit', row?: ScopeListItem) => {
    dialogVisible.value = true
    dialogType.value = type
    currentScopeData.value = row
  }

  const buttonMoreClick = (item: ButtonMoreItem, row: ScopeListItem) => {
    switch (item.key) {
      case 'edit':
        showDialog('edit', row)
        break
      case 'delete':
        deleteScope(row)
        break
    }
  }

  const deleteScope = (row: ScopeListItem) => {
    ElMessageBox.confirm(`确定删除作用域“${row.scopeName}”吗？此操作不可恢复。`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeleteScope(row.scopeId))
      .then(() => {
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') {
          const error = e as any
          const errorData = error?.data?.data || error?.data
          if (errorData && typeof errorData === 'object' && Array.isArray(errorData.roles)) {
            const roles = errorData.roles as string[]
            const roleCount = errorData.roleCount || roles.length
            ElMessageBox.alert(
              `无法删除作用域“${row.scopeName}”，该作用域已被 ${roleCount} 个角色使用：\n${roles.join('、')}`,
              '删除失败',
              {
                confirmButtonText: '确定',
                type: 'warning',
                dangerouslyUseHTMLString: false
              }
            )
            return
          }
          ElMessage.error(error?.message || '删除失败')
        }
      })
  }

  getData()
</script>

<style scoped>
  .scope-page {
    height: 100%;
    padding: 16px;
  }

  .scope-card {
    height: 100%;
  }

  .scope-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .scope-toolbar__title {
    font-size: 18px;
    font-weight: 600;
  }
</style>
