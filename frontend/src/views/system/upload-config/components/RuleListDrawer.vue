<template>
  <ElDrawer
    v-model="visible"
    :title="`上传规则管理 — ${parentLabel}`"
    size="860px"
    direction="rtl"
  >
    <div class="rule-drawer-desc">
      上传规则是上传配置的子级。同一个 UploadKey 下可按规则覆写上传方式、可见性、前端选择提示和扩展参数，用来区分图片、附件、海报等细分场景。
    </div>
    <div class="rule-drawer-toolbar">
      <ElButton type="primary" size="small" @click="loadRules">刷新</ElButton>
      <ElButton type="success" size="small" @click="openCreate">新增规则</ElButton>
    </div>
    <ArtTable :loading="loading" :data="records" :columns="columns" />

    <RuleEditorDrawer
      v-model:open="editorOpen"
      :editing-id="editingId"
      :row="editingRow"
      :parent-upload-key-id="parentUploadKeyId"
      @saved="onRuleSaved"
    />
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, h, ref, watch } from 'vue'
  import {
    ElButton,
    ElDrawer,
    ElMessage,
    ElPopconfirm,
    ElTag
  } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import {
    fetchDeleteUploadKeyRule,
    fetchListUploadKeyRules,
    type UploadKeyRuleSummary
  } from '@/domains/upload-config/api'
  import { useDictionaries } from '@/hooks/business/useDictionary'
  import RuleEditorDrawer from './RuleEditorDrawer.vue'
  import {
    formatBytes,
    formatSchemaConfigured,
    statusLabel,
    statusType,
    filenameStrategyLabel,
    uploadModeLabel,
    visibilityOverrideLabel
  } from './_shared'

  defineOptions({ name: 'RuleListDrawer' })

  const props = defineProps<{
    open: boolean
    parentUploadKeyId: string
    parentLabel: string
  }>()

  const emit = defineEmits<{
    'update:open': [value: boolean]
  }>()

  const visible = computed({
    get: () => props.open,
    set: (v) => emit('update:open', v)
  })

  const loading = ref(false)
  const records = ref<UploadKeyRuleSummary[]>([])
  const editorOpen = ref(false)
  const editingId = ref('')
  const editingRow = ref<UploadKeyRuleSummary | null>(null)

  const { dictMap } = useDictionaries([
    'upload_filename_strategy',
    'upload_mode_override',
    'upload_visibility_override',
    'storage_status'
  ])

  function pickLabel(code: string, value: string, fallback: string) {
    const list = dictMap.value[code] || []
    const hit = list.find((item) => item.value === value)
    return hit?.label || fallback
  }

  watch(
    () => [props.open, props.parentUploadKeyId] as const,
    ([open]) => {
      if (open && props.parentUploadKeyId) loadRules()
    },
    { immediate: true }
  )

  async function loadRules() {
    if (!props.parentUploadKeyId) return
    loading.value = true
    try {
      const res = await fetchListUploadKeyRules(props.parentUploadKeyId)
      records.value = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载规则列表失败')
    } finally {
      loading.value = false
    }
  }

  function openCreate() {
    editingId.value = ''
    editingRow.value = null
    editorOpen.value = true
  }

  function openEdit(row: UploadKeyRuleSummary) {
    editingId.value = row.id
    editingRow.value = row
    editorOpen.value = true
  }

  function onRuleSaved() {
    loadRules()
  }

  async function removeRule(row: UploadKeyRuleSummary) {
    try {
      await fetchDeleteUploadKeyRule(row.id)
      ElMessage.success('已删除')
      loadRules()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除规则失败')
    }
  }

  const columns = computed<ColumnOption[]>(() => [
    { prop: 'rule_key', label: '规则标识', minWidth: 130 },
    { prop: 'name', label: '名称', minWidth: 130 },
    { prop: 'sub_path', label: '子路径', minWidth: 100 },
    {
      prop: 'filename_strategy',
      label: '文件名策略',
      width: 140,
      formatter: (row: UploadKeyRuleSummary) =>
        pickLabel(
          'upload_filename_strategy',
          row.filename_strategy,
          filenameStrategyLabel[row.filename_strategy] || row.filename_strategy
        )
    },
    {
      prop: 'mode_override',
      label: '上传方式覆写',
      width: 140,
      formatter: (row: UploadKeyRuleSummary) =>
        pickLabel(
          'upload_mode_override',
          row.mode_override || 'inherit',
          uploadModeLabel[row.mode_override || 'inherit'] || row.mode_override || 'inherit'
        )
    },
    {
      prop: 'visibility_override',
      label: '可见性覆写',
      width: 140,
      formatter: (row: UploadKeyRuleSummary) =>
        pickLabel(
          'upload_visibility_override',
          row.visibility_override || 'inherit',
          visibilityOverrideLabel[row.visibility_override || 'inherit'] ||
            row.visibility_override ||
            'inherit'
        )
    },
    {
      prop: 'max_size_bytes',
      label: '文件上限',
      width: 110,
      formatter: (row: UploadKeyRuleSummary) => formatBytes(Number(row.max_size_bytes ?? 0))
    },
    {
      prop: 'allowed_mime_types',
      label: '允许类型',
      minWidth: 160,
      showOverflowTooltip: true,
      formatter: (row: UploadKeyRuleSummary) =>
        Array.isArray(row.allowed_mime_types) && row.allowed_mime_types.length
          ? row.allowed_mime_types.join(', ')
          : '不限'
    },
    {
      prop: 'extra_schema',
      label: '扩展参数',
      width: 100,
      formatter: (row: UploadKeyRuleSummary) => formatSchemaConfigured(row.extra_schema)
    },
    {
      prop: 'is_default',
      label: '默认',
      width: 70,
      formatter: (row: UploadKeyRuleSummary) =>
        row.is_default
          ? h(ElTag, { type: 'success', effect: 'plain', size: 'small' }, () => '是')
          : '-'
    },
    {
      prop: 'status',
      label: '状态',
      width: 80,
      formatter: (row: UploadKeyRuleSummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain', size: 'small' },
          () =>
            pickLabel('storage_status', row.status, statusLabel[row.status] || row.status)
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 160,
      fixed: 'right',
      formatter: (row: UploadKeyRuleSummary) =>
        h('div', { class: 'rule-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openEdit(row) },
            () => '编辑'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该规则？', onConfirm: () => removeRule(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])
</script>

<style scoped lang="scss">
  .rule-drawer-desc {
    margin-bottom: 12px;
    padding: 8px 12px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-lighter);
    border-radius: 4px;
  }

  .rule-drawer-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  :deep(.rule-row-actions) {
    display: flex;
    align-items: center;
    gap: 4px;
  }
</style>
