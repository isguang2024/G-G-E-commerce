<template>
  <div class="art-full-height">
    <ElCard shadow="never" class="module-card">
      <div class="module-header">
        <div class="module-title">模块分类</div>
        <div class="module-help">按模块查看接口，路径相同但方法不同会保留为独立规格</div>
      </div>
      <div class="module-tags">
        <ElTag
          :type="selectedModule === '' ? 'primary' : 'info'"
          effect="light"
          class="module-tag"
          @click="handleModuleSelect('')"
        >
          全部 {{ totalCount }}
        </ElTag>
        <ElTag
          v-for="item in moduleSummary"
          :key="item.label"
          :type="selectedModule === item.label ? 'primary' : 'info'"
          effect="light"
          class="module-tag"
          @click="handleModuleSelect(item.label)"
        >
          {{ item.label }} {{ item.count }}
        </ElTag>
      </div>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton v-action="'system.api_registry.sync'" type="primary" :loading="syncing" @click="handleSync" v-ripple>
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
  type SummaryItem = { label: string; count: number }

  const syncing = ref(false)
  const selectedModule = ref('')
  const totalCount = ref(0)
  const moduleSummary = ref<SummaryItem[]>([])

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
      apiFn: fetchGetApiEndpointList,
      apiParams: {
        current: 1,
        size: 20,
        module: ''
      },
      columnsFactory: () => [
        {
          prop: 'spec',
          label: '接口规格',
          minWidth: 320,
          showOverflowTooltip: true,
          formatter: (row: APIEndpointItem) => row.spec || `${row.method} ${row.path}`
        },
        { prop: 'module', label: '模块', width: 120 },
        {
          prop: 'authMode',
          label: '鉴权模式',
          width: 110,
          formatter: (row: APIEndpointItem) => {
            const config = {
              public: { type: 'info', text: '公开' },
              jwt: { type: 'warning', text: '仅登录' },
              permission: { type: 'success', text: '功能权限' },
              api_key: { type: 'danger', text: 'API Key' }
            } as const
            const current = config[row.authMode as keyof typeof config] || { type: 'info', text: row.authMode || '-' }
            return h(ElTag, { type: current.type as 'success' | 'info' | 'warning' | 'danger' }, () => current.text)
          }
        },
        {
          prop: 'permissionKey',
          label: '功能权限键',
          minWidth: 220,
          formatter: (row: APIEndpointItem) => row.permissionKey || '-'
        },
        {
          prop: 'featureKind',
          label: '功能归属',
          width: 110,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: row.featureKind === 'business' ? 'success' : 'info' }, () =>
              row.featureKind === 'business' ? '业务功能' : '系统功能'
            )
        },
        { prop: 'summary', label: '说明', minWidth: 180, showOverflowTooltip: true },
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
      await refreshData()
      await loadModuleSummary()
    } catch (error: any) {
      ElMessage.error(error?.message || '同步失败')
    } finally {
      syncing.value = false
    }
  }

  async function loadModuleSummary() {
    const res = await fetchGetApiEndpointList({ current: 1, size: 1000 })
    const records = res.records || []
    totalCount.value = res.total || records.length
    const counter = new Map<string, number>()
    records.forEach((item) => {
      const key = (item.module || 'unknown').trim() || 'unknown'
      counter.set(key, (counter.get(key) || 0) + 1)
    })
    moduleSummary.value = [...counter.entries()]
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0], 'zh-CN'))
      .map(([label, count]) => ({ label, count }))
  }

  async function handleModuleSelect(module: string) {
    selectedModule.value = module
    Object.assign(searchParams, {
      module: module || undefined,
      current: 1
    })
    await getData()
  }

  onMounted(() => {
    loadModuleSummary()
  })
</script>

<style scoped>
  .module-card {
    margin-bottom: 12px;
  }

  .module-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }

  .module-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .module-help {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .module-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .module-tag {
    cursor: pointer;
  }
</style>
