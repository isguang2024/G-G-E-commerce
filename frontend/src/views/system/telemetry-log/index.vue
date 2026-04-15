<template>
  <div class="p-4 telemetry-log-page art-full-height">
    <ElCard class="art-table-card telemetry-log-main" shadow="never">
      <ElForm :inline="true" :model="filter" class="telemetry-log-filters">
        <ElFormItem label="Level">
          <ElSelect v-model="filter.level" clearable placeholder="全部" style="width: 140px">
            <ElOption label="debug" value="debug" />
            <ElOption label="info" value="info" />
            <ElOption label="warn" value="warn" />
            <ElOption label="error" value="error" />
            <ElOption label="fatal" value="fatal" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Event">
          <ElInput v-model="filter.event" clearable placeholder="domain.entity.action" />
        </ElFormItem>
        <ElFormItem label="Session">
          <ElInput v-model="filter.session_id" clearable />
        </ElFormItem>
        <ElFormItem label="Actor">
          <ElInput v-model="filter.actor_id" clearable />
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
          <div class="telemetry-log-header">
            <div class="telemetry-log-title">前端遥测日志</div>
            <div class="telemetry-log-tip">
              来源 telemetry_logs，前端 logger 通过 /telemetry/logs 异步上报；数据只读。
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

    <ElDrawer v-model="drawer.open" title="遥测详情" size="520px" :destroy-on-close="true">
      <div v-if="drawer.loading" class="telemetry-drawer-loading">加载中…</div>
      <template v-else-if="drawer.detail">
        <ElDescriptions :column="1" border>
          <ElDescriptionsItem label="ID">{{ drawer.detail.id }}</ElDescriptionsItem>
          <ElDescriptionsItem label="时间">{{ drawer.detail.ts }}</ElDescriptionsItem>
          <ElDescriptionsItem label="租户">{{ drawer.detail.tenant_id }}</ElDescriptionsItem>
          <ElDescriptionsItem label="Level">
            <ElTag :type="levelTagType(drawer.detail.level)">{{ drawer.detail.level }}</ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem label="Event">{{ drawer.detail.event }}</ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.message" label="Message">
            {{ drawer.detail.message }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.session_id" label="Session">
            {{ drawer.detail.session_id }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.actor_id" label="Actor">
            {{ drawer.detail.actor_id }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.request_id" label="Request ID">
            <ElButton type="primary" link @click="openTrace(drawer.detail.request_id)">
              {{ drawer.detail.request_id }}
            </ElButton>
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.app_key" label="App Key">
            {{ drawer.detail.app_key }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.url" label="URL">
            {{ drawer.detail.url }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.release" label="Release">
            {{ drawer.detail.release }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.ip" label="IP">
            {{ drawer.detail.ip }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="drawer.detail.user_agent" label="UA">
            {{ drawer.detail.user_agent }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <div v-if="drawer.detail.payload" class="telemetry-json-block">
          <div class="telemetry-json-label">Payload</div>
          <JsonViewer :data="drawer.detail.payload" />
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
  import {
    fetchGetTelemetryLog,
    fetchListTelemetryLogs
  } from '@/domains/governance/api/observability'
  import { observabilityTimeShortcuts } from '@/views/system/_shared/observability-shortcuts'
  import JsonViewer from '@/components/Observability/JsonViewer.vue'
  import TraceDrawer from '@/components/Observability/TraceDrawer.vue'

  defineOptions({ name: 'SystemTelemetryLog' })

  type FilterState = {
    level: string
    event: string
    session_id: string
    actor_id: string
    request_id: string
    range: [string, string] | null
  }

  const list = ref<any[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const filter = reactive<FilterState>({
    level: '',
    event: '',
    session_id: '',
    actor_id: '',
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

  const levelTagType = (level?: string) => {
    switch (level) {
      case 'error':
      case 'fatal':
        return 'danger'
      case 'warn':
        return 'warning'
      case 'info':
        return 'success'
      default:
        return 'info'
    }
  }

  const columns = computed<ColumnOption[]>(() => [
    { prop: 'ts', label: '时间', width: 180 },
    {
      prop: 'level',
      label: 'Level',
      width: 100,
      formatter: (row: any) =>
        h(ElTag, { type: levelTagType(row.level) as any, effect: 'plain' }, () => row.level || '-')
    },
    { prop: 'event', label: 'Event', minWidth: 220, showOverflowTooltip: true },
    { prop: 'message', label: 'Message', minWidth: 240, showOverflowTooltip: true },
    { prop: 'actor_id', label: 'Actor', width: 160, showOverflowTooltip: true },
    { prop: 'session_id', label: 'Session', width: 200, showOverflowTooltip: true },
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
    { prop: 'url', label: 'URL', minWidth: 220, showOverflowTooltip: true },
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
    if (filter.level) q.level = filter.level
    if (filter.event) q.event = filter.event
    if (filter.session_id) q.session_id = filter.session_id
    if (filter.actor_id) q.actor_id = filter.actor_id
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
      const data: any = await fetchListTelemetryLogs(buildQuery())
      list.value = data?.records || []
      total.value = data?.total || 0
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  const reset = () => {
    filter.level = ''
    filter.event = ''
    filter.session_id = ''
    filter.actor_id = ''
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
      drawer.detail = await fetchGetTelemetryLog(id)
    } catch (e: any) {
      ElMessage.error(e?.message || '加载详情失败')
    } finally {
      drawer.loading = false
    }
  }

  onMounted(() => load(1))
</script>

<style scoped>
  .telemetry-log-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .telemetry-log-main {
    flex: 1;
    min-height: 0;
  }

  .telemetry-log-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .telemetry-log-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .telemetry-log-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .telemetry-log-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .telemetry-log-filters {
    margin-bottom: 4px;
  }

  .telemetry-drawer-loading {
    padding: 40px;
    text-align: center;
    color: var(--el-text-color-secondary);
  }

  .telemetry-json-block {
    margin-top: 16px;
  }

  .telemetry-json-label {
    font-weight: 600;
    margin-bottom: 6px;
    color: var(--el-text-color-primary);
  }
</style>
