<template>
  <div class="p-4 audit-log-page art-full-height">
    <ElCard class="art-table-card audit-log-main" shadow="never">
      <ElForm :inline="true" :model="filter" class="audit-log-filters">
        <ElFormItem label="Action">
          <ElInput v-model="filter.action" clearable placeholder="如 system.user.create" />
        </ElFormItem>
        <ElFormItem label="Actor">
          <ElInput v-model="filter.actor_id" clearable placeholder="用户 ID / tenant-key" />
        </ElFormItem>
        <ElFormItem label="Outcome">
          <ElSelect v-model="filter.outcome" clearable placeholder="全部" style="width: 140px">
            <ElOption label="success" value="success" />
            <ElOption label="failure" value="failure" />
            <ElOption label="denied" value="denied" />
            <ElOption label="not_found" value="not_found" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Resource 类型">
          <ElInput v-model="filter.resource_type" clearable placeholder="user / role / ..." />
        </ElFormItem>
        <ElFormItem label="Resource ID">
          <ElInput v-model="filter.resource_id" clearable />
        </ElFormItem>
        <ElFormItem label="Request ID">
          <ElInput v-model="filter.request_id" clearable />
        </ElFormItem>
        <ElFormItem label="时间区间">
          <ElDatePicker
            v-model="filter.range"
            type="datetimerange"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            range-separator="-"
            start-placeholder="开始"
            end-placeholder="结束"
            :shortcuts="observabilityTimeShortcuts"
            @change="() => load(1)"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="load(1)">查询</ElButton>
          <ElButton @click="reset">重置</ElButton>
        </ElFormItem>
      </ElForm>

      <ArtTableHeader layout="refresh,fullscreen" :loading="loading" @refresh="load()">
        <template #left>
          <div class="audit-log-header">
            <div class="audit-log-title">审计日志</div>
            <div class="audit-log-tip">
              来源 audit_logs，展示写操作痕迹；数据只读，由 audit.Recorder 异步落库。
            </div>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDrawer v-model="drawer.open" title="审计详情" size="520px" :destroy-on-close="true">
      <div v-if="drawer.loading" class="audit-drawer-loading">加载中…</div>
      <template v-else-if="drawer.detail">
        <ElDescriptions :column="1" border>
          <ElDescriptionsItem label="ID">{{ drawer.detail.id }}</ElDescriptionsItem>
          <ElDescriptionsItem label="时间">{{ drawer.detail.ts }}</ElDescriptionsItem>
          <ElDescriptionsItem label="租户">{{ drawer.detail.tenant_id }}</ElDescriptionsItem>
          <ElDescriptionsItem label="Actor">
            {{ drawer.detail.actor_id }} ({{ drawer.detail.actor_type }})
          </ElDescriptionsItem>
          <ElDescriptionsItem label="Action">{{ drawer.detail.action }}</ElDescriptionsItem>
          <ElDescriptionsItem label="Outcome">
            <ElTag :type="outcomeTagType(drawer.detail.outcome)">{{ drawer.detail.outcome }}</ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.resource_type" label="Resource">
            {{ drawer.detail.resource_type }} / {{ drawer.detail.resource_id || '-' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.error_code" label="错误码">
            {{ drawer.detail.error_code }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.http_status" label="HTTP">
            {{ drawer.detail.http_status }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.request_id" label="Request ID">
            <ElButton type="primary" link @click="openTrace(drawer.detail.request_id)">
              {{ drawer.detail.request_id }}
            </ElButton>
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.app_key" label="App Key">
            {{ drawer.detail.app_key }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.workspace_id" label="Workspace">
            {{ drawer.detail.workspace_id }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.ip" label="IP">{{
            drawer.detail.ip
          }}</ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.user_agent" label="UA">
            {{ drawer.detail.user_agent }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <div v-if="drawer.detail.before" class="audit-json-block">
          <div class="audit-json-label">Before</div>
          <JsonViewer :data="drawer.detail.before" />
        </div>
        <div v-if="drawer.detail.after" class="audit-json-block">
          <div class="audit-json-label">After</div>
          <JsonViewer :data="drawer.detail.after" />
        </div>
        <div v-if="drawer.detail.metadata" class="audit-json-block">
          <div class="audit-json-label">Metadata</div>
          <JsonViewer :data="drawer.detail.metadata" />
        </div>
      </template>
    </ElDrawer>

    <TraceDrawer v-model="trace.open" :request-id="trace.requestId" />
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import {
    ElButton,
    ElCard,
    ElDatePicker,
    ElDescriptions,
    ElDescriptionsItem,
    ElDrawer,
    ElForm,
    ElFormItem,
    ElInput,
    ElMessage,
    ElOption,
    ElSelect,
    ElTag
  } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import { fetchGetAuditLog, fetchListAuditLogs } from '@/domains/governance/api/observability'
  import { observabilityTimeShortcuts } from '@/views/system/_shared/observability-shortcuts'
  import JsonViewer from '@/components/Observability/JsonViewer.vue'
  import TraceDrawer from '@/components/Observability/TraceDrawer.vue'

  defineOptions({ name: 'SystemAuditLog' })

  type FilterState = {
    action: string
    actor_id: string
    outcome: string
    resource_type: string
    resource_id: string
    request_id: string
    range: [string, string] | null
  }

  const list = ref<any[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const filter = reactive<FilterState>({
    action: '',
    actor_id: '',
    outcome: '',
    resource_type: '',
    resource_id: '',
    request_id: '',
    range: null
  })
  const drawer = reactive<{ open: boolean; loading: boolean; detail: any | null }>({
    open: false,
    loading: false,
    detail: null
  })
  const trace = reactive<{ open: boolean; requestId: string | null }>({
    open: false,
    requestId: null
  })
  const openTrace = (requestId?: string) => {
    if (!requestId) return
    trace.requestId = requestId
    trace.open = true
  }
  const pagination = computed(() => ({
    current: page.value,
    size: pageSize.value,
    total: total.value
  }))

  const outcomeTagType = (outcome?: string) => {
    switch (outcome) {
      case 'success':
        return 'success'
      case 'failure':
        return 'danger'
      case 'denied':
        return 'warning'
      default:
        return 'info'
    }
  }

  const columns = computed<ColumnOption[]>(() => [
    { prop: 'ts', label: '时间', width: 180 },
    { prop: 'action', label: 'Action', minWidth: 220, showOverflowTooltip: true },
    {
      prop: 'outcome',
      label: 'Outcome',
      width: 110,
      formatter: (row: any) =>
        h(
          ElTag,
          { type: outcomeTagType(row.outcome) as any, effect: 'plain' },
          () => row.outcome || '-'
        )
    },
    { prop: 'actor_id', label: 'Actor', width: 160, showOverflowTooltip: true },
    {
      prop: 'resource',
      label: 'Resource',
      minWidth: 200,
      formatter: (row: any) =>
        row.resource_type ? `${row.resource_type}/${row.resource_id || '-'}` : '-'
    },
    { prop: 'http_status', label: 'HTTP', width: 80 },
    { prop: 'error_code', label: '错误码', width: 160, showOverflowTooltip: true },
    {
      prop: 'request_id',
      label: 'Request ID',
      width: 220,
      showOverflowTooltip: true,
      formatter: (row: any) =>
        row.request_id
          ? h(
              ElButton,
              {
                type: 'primary',
                link: true,
                onClick: (ev: MouseEvent) => {
                  ev.stopPropagation()
                  openTrace(row.request_id)
                }
              },
              () => row.request_id
            )
          : '-'
    },
    { prop: 'ip', label: 'IP', width: 140 },
    {
      prop: 'actions',
      label: '操作',
      width: 80,
      fixed: 'right',
      formatter: (row: any) =>
        h(
          ElButton,
          { type: 'primary', link: true, onClick: () => openDetail(row.id) },
          () => '详情'
        )
    }
  ])

  const buildQuery = () => {
    const q: Record<string, unknown> = {
      current: page.value,
      size: pageSize.value
    }
    if (filter.action) q.action = filter.action
    if (filter.actor_id) q.actor_id = filter.actor_id
    if (filter.outcome) q.outcome = filter.outcome
    if (filter.resource_type) q.resource_type = filter.resource_type
    if (filter.resource_id) q.resource_id = filter.resource_id
    if (filter.request_id) q.request_id = filter.request_id
    if (filter.range && filter.range.length === 2) {
      if (filter.range[0]) q.from = filter.range[0]
      if (filter.range[1]) q.to = filter.range[1]
    }
    return q as any
  }

  const load = async (p?: number) => {
    if (p) page.value = p
    loading.value = true
    try {
      const data: any = await fetchListAuditLogs(buildQuery())
      list.value = data?.records || []
      total.value = data?.total || 0
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  const reset = () => {
    filter.action = ''
    filter.actor_id = ''
    filter.outcome = ''
    filter.resource_type = ''
    filter.resource_id = ''
    filter.request_id = ''
    filter.range = null
    load(1)
  }

  const handleSizeChange = (size: number) => {
    pageSize.value = size
    load(1)
  }

  const handleCurrentChange = (current: number) => {
    load(current)
  }

  const openDetail = async (id: number) => {
    drawer.open = true
    drawer.loading = true
    drawer.detail = null
    try {
      drawer.detail = await fetchGetAuditLog(id)
    } catch (e: any) {
      ElMessage.error(e?.message || '加载详情失败')
    } finally {
      drawer.loading = false
    }
  }

  onMounted(() => load(1))
</script>

<style scoped>
  .audit-log-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .audit-log-main {
    flex: 1;
    min-height: 0;
  }

  .audit-log-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .audit-log-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .audit-log-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .audit-log-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .audit-log-filters {
    margin-bottom: 4px;
  }

  .audit-drawer-loading {
    padding: 40px;
    text-align: center;
    color: var(--el-text-color-secondary);
  }

  .audit-json-block {
    margin-top: 16px;
  }

  .audit-json-label {
    font-weight: 600;
    margin-bottom: 6px;
    color: var(--el-text-color-primary);
  }
</style>
