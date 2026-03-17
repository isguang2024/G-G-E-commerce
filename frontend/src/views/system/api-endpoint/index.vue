<template>
  <div class="art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton v-action="'api_endpoint:sync'" type="primary" :loading="syncing" @click="handleSync" v-ripple>
            同步 API
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
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetApiEndpointList, fetchSyncApiEndpoints } from '@/api/system-manage'
  import { ElMessage, ElTag } from 'element-plus'

  defineOptions({ name: 'ApiEndpoint' })

  type APIEndpointItem = Api.SystemManage.APIEndpointItem

  const syncing = ref(false)

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
      apiFn: fetchGetApiEndpointList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        { prop: 'method', label: '方法', width: 90 },
        { prop: 'path', label: '路径', minWidth: 260, showOverflowTooltip: true },
        { prop: 'module', label: '模块', width: 120 },
        { prop: 'summary', label: '说明', minWidth: 180, showOverflowTooltip: true },
        { prop: 'resourceCode', label: '资源编码', minWidth: 140 },
        { prop: 'actionCode', label: '动作编码', minWidth: 160 },
        {
          prop: 'scopeName',
          label: '作用域',
          width: 90,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: row.scopeCode === 'team' ? 'success' : 'primary' }, () =>
              row.scopeName || (row.scopeCode === 'team' ? '团队' : '平台')
            )
        },
        {
          prop: 'requiresTenantContext',
          label: '依赖团队',
          width: 100,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: row.requiresTenantContext ? 'warning' : 'info' }, () =>
              row.requiresTenantContext ? '是' : '否'
            )
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 }
      ]
    }
  })

  async function handleSync() {
    syncing.value = true
    try {
      await fetchSyncApiEndpoints()
      ElMessage.success('同步成功')
      refreshData()
    } catch (error: any) {
      ElMessage.error(error?.message || '同步失败')
    } finally {
      syncing.value = false
    }
  }
</script>
