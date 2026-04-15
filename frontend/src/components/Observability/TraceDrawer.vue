<template>
  <ElDrawer v-model="visible" :title="title" size="640px" :destroy-on-close="true" append-to-body>
    <div v-if="loading" class="trace-drawer-loading">加载中…</div>
    <template v-else-if="bundle">
      <div class="trace-drawer-meta">
        <div class="trace-drawer-title">Request ID</div>
        <div class="trace-drawer-value">{{ bundle.request_id }}</div>
        <div class="trace-drawer-stat">
          audit {{ bundle.audit_logs.length }} · telemetry {{ bundle.telemetry_logs.length }}
        </div>
      </div>

      <div v-if="merged.length === 0" class="trace-drawer-empty">
        这条 request_id 没有对应的审计或遥测记录。
      </div>
      <ElTimeline v-else class="trace-drawer-timeline">
        <ElTimelineItem
          v-for="(item, idx) in merged"
          :key="`${item.source}-${item.id}-${idx}`"
          :timestamp="item.ts"
          placement="top"
          :type="timelineColor(item)"
          :hollow="item.source === 'telemetry'"
        >
          <div class="trace-item">
            <div class="trace-item-head">
              <ElTag
                size="small"
                :type="item.source === 'audit' ? 'primary' : 'info'"
                effect="plain"
              >
                {{ item.source }}
              </ElTag>
              <span class="trace-item-label">{{ item.label }}</span>
              <ElTag v-if="item.badge" :type="item.badgeType" size="small" effect="plain">
                {{ item.badge }}
              </ElTag>
            </div>
            <div v-if="item.sub" class="trace-item-sub">{{ item.sub }}</div>
            <div v-if="item.extra" class="trace-item-extra">{{ item.extra }}</div>
          </div>
        </ElTimelineItem>
      </ElTimeline>
    </template>
    <div v-else class="trace-drawer-empty">未加载到数据。</div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElDrawer, ElMessage, ElTag, ElTimeline, ElTimelineItem } from 'element-plus'
  import { fetchObservabilityTrace } from '@/domains/governance/api/observability'

  defineOptions({ name: 'TraceDrawer' })

  type AuditRow = {
    id: number
    ts: string
    action?: string
    outcome?: string
    actor_id?: string
    resource_type?: string
    resource_id?: string
    http_status?: number
    error_code?: string
  }
  type TelemetryRow = {
    id: number
    ts: string
    level?: string
    event?: string
    message?: string
    actor_id?: string
    url?: string
  }
  type TraceBundle = {
    request_id: string
    audit_logs: AuditRow[]
    telemetry_logs: TelemetryRow[]
  }
  type TimelineItem = {
    source: 'audit' | 'telemetry'
    id: number
    ts: string
    label: string
    sub?: string
    extra?: string
    badge?: string
    badgeType?: 'success' | 'danger' | 'warning' | 'info' | 'primary'
  }

  const props = defineProps<{ modelValue: boolean; requestId: string | null }>()
  const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (v) => emit('update:modelValue', v)
  })

  const loading = ref(false)
  const bundle = ref<TraceBundle | null>(null)

  const title = computed(() => `请求链路 ${props.requestId ?? ''}`.trim())

  watch(
    () => [props.modelValue, props.requestId] as const,
    async ([open, reqID]) => {
      if (!open || !reqID) return
      loading.value = true
      bundle.value = null
      try {
        bundle.value = (await fetchObservabilityTrace(reqID)) as TraceBundle
      } catch (e: any) {
        ElMessage.error(e?.message || '加载链路失败')
      } finally {
        loading.value = false
      }
    },
    { immediate: true }
  )

  const merged = computed<TimelineItem[]>(() => {
    if (!bundle.value) return []
    const items: TimelineItem[] = []
    for (const a of bundle.value.audit_logs) {
      const status = a.http_status ? `HTTP ${a.http_status}` : ''
      const resource = a.resource_type
        ? `${a.resource_type}${a.resource_id ? '/' + a.resource_id : ''}`
        : ''
      items.push({
        source: 'audit',
        id: a.id,
        ts: a.ts,
        label: a.action || '-',
        sub: [a.actor_id, resource].filter(Boolean).join(' · '),
        extra: [status, a.error_code].filter(Boolean).join(' · ') || undefined,
        badge: a.outcome,
        badgeType: outcomeBadgeType(a.outcome)
      })
    }
    for (const t of bundle.value.telemetry_logs) {
      items.push({
        source: 'telemetry',
        id: t.id,
        ts: t.ts,
        label: t.event || '-',
        sub: t.message,
        extra: t.url,
        badge: t.level,
        badgeType: levelBadgeType(t.level)
      })
    }
    // 后端已按 ts asc 返回，合并后再稳定排一次（保证两源交错正确）
    items.sort((a, b) => (a.ts > b.ts ? 1 : a.ts < b.ts ? -1 : 0))
    return items
  })

  function outcomeBadgeType(o?: string): TimelineItem['badgeType'] {
    switch (o) {
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
  function levelBadgeType(l?: string): TimelineItem['badgeType'] {
    switch (l) {
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
  function timelineColor(
    item: TimelineItem
  ): '' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
    if (item.source === 'audit') {
      if (item.badge === 'failure') return 'danger'
      if (item.badge === 'denied') return 'warning'
      if (item.badge === 'success') return 'success'
      return 'primary'
    }
    if (item.badge === 'error' || item.badge === 'fatal') return 'danger'
    if (item.badge === 'warn') return 'warning'
    return 'info'
  }
</script>

<style scoped>
  .trace-drawer-loading,
  .trace-drawer-empty {
    padding: 32px 16px;
    color: var(--el-text-color-secondary);
    text-align: center;
  }

  .trace-drawer-meta {
    padding: 8px 4px 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    margin-bottom: 12px;
  }

  .trace-drawer-title {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-bottom: 2px;
  }

  .trace-drawer-value {
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 13px;
    color: var(--el-text-color-primary);
    word-break: break-all;
  }

  .trace-drawer-stat {
    margin-top: 6px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .trace-drawer-timeline {
    margin-top: 8px;
  }

  .trace-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .trace-item-head {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .trace-item-label {
    font-weight: 600;
    color: var(--el-text-color-primary);
    word-break: break-all;
  }

  .trace-item-sub {
    color: var(--el-text-color-regular);
    font-size: 13px;
    word-break: break-all;
  }

  .trace-item-extra {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    word-break: break-all;
  }
</style>
