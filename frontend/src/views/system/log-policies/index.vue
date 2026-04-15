<template>
  <div class="p-4 log-policies-page art-full-height">
    <ElCard class="art-table-card log-policies-main" shadow="never">
      <ElForm :inline="true" class="log-policies-filters">
        <ElFormItem label="Pipeline">
          <ElSelect v-model="filters.pipeline" clearable placeholder="全部" style="width: 160px">
            <ElOption label="audit" value="audit" />
            <ElOption label="telemetry" value="telemetry" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="启用状态">
          <ElSelect v-model="filters.enabled" style="width: 160px">
            <ElOption label="全部" value="" />
            <ElOption label="启用" value="true" />
            <ElOption label="停用" value="false" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="load(1)">查询</ElButton>
          <ElButton @click="resetFilters">重置</ElButton>
          <ElButton type="success" @click="openCreateDialog">新增策略</ElButton>
        </ElFormItem>
      </ElForm>

      <ArtTableHeader layout="refresh,fullscreen" :loading="loading" @refresh="load()">
        <template #left>
          <div class="log-policies-header">
            <div class="log-policies-title">日志策略管理</div>
            <div class="log-policies-tip">
              维护审计/遥测策略规则。命中 compliance lock 的规则仅允许查看与预览，不允许编辑或删除。
            </div>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="records"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDialog
      v-model="editor.open"
      :title="editor.mode === 'create' ? '新增日志策略' : '编辑日志策略'"
      width="680px"
      :close-on-click-modal="false"
      @closed="resetEditor"
    >
      <ElForm label-width="110px">
        <ElFormItem label="Pipeline">
          <ElSelect
            v-model="editor.form.pipeline"
            style="width: 220px"
            :disabled="editor.mode === 'edit'"
            @change="onPipelineChange"
          >
            <ElOption label="audit" value="audit" />
            <ElOption label="telemetry" value="telemetry" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Match Field">
          <ElSelect v-model="editor.form.match_field" style="width: 220px">
            <ElOption
              v-for="option in matchFieldOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Pattern">
          <ElInput v-model="editor.form.pattern" placeholder="支持后缀 * 前缀匹配，例如 observability.policy.*" />
        </ElFormItem>
        <ElFormItem label="Decision">
          <ElSelect v-model="editor.form.decision" style="width: 220px">
            <ElOption label="allow" value="allow" />
            <ElOption label="deny" value="deny" />
            <ElOption label="sample" value="sample" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Sample Rate">
          <ElInputNumber
            v-model="editor.form.sample_rate"
            :disabled="editor.form.decision !== 'sample'"
            :min="1"
            :max="100"
            controls-position="right"
          />
          <span class="form-tip">仅 decision=sample 时生效，范围 1-100</span>
        </ElFormItem>
        <ElFormItem label="Priority">
          <ElInputNumber v-model="editor.form.priority" controls-position="right" />
        </ElFormItem>
        <ElFormItem label="Enabled">
          <ElSwitch v-model="editor.form.enabled" />
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput
            v-model="editor.form.note"
            type="textarea"
            :rows="3"
            placeholder="可选，说明规则用途"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="editor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="editor.submitting" @click="submitEditor">保存</ElButton>
      </template>
    </ElDialog>

    <ElDialog
      v-model="preview.open"
      title="策略预览"
      width="620px"
      :close-on-click-modal="false"
      @closed="resetPreview"
    >
      <ElForm label-width="110px">
        <ElFormItem label="Pipeline">
          <ElInput :model-value="preview.pipeline" disabled />
        </ElFormItem>
        <ElFormItem label="Match Field">
          <ElInput :model-value="preview.match_field" disabled />
        </ElFormItem>
        <ElFormItem label="字段值">
          <ElInput v-model="preview.fieldValue" placeholder="输入待匹配值后执行预览" />
        </ElFormItem>
      </ElForm>
      <ElAlert
        v-if="preview.result"
        :type="preview.result.decision === 'deny' ? 'error' : preview.result.decision === 'sample' ? 'warning' : 'success'"
        show-icon
        :closable="false"
      >
        <template #title>
          结果：{{ preview.result.decision }}
          <span v-if="preview.result.sample_rate != null">（sample_rate={{ preview.result.sample_rate }}）</span>
        </template>
      </ElAlert>
      <ElDescriptions v-if="preview.result?.policy" :column="1" border class="preview-policy">
        <ElDescriptionsItem label="命中策略 ID">{{ preview.result.policy.id }}</ElDescriptionsItem>
        <ElDescriptionsItem label="Pattern">{{ preview.result.policy.pattern }}</ElDescriptionsItem>
        <ElDescriptionsItem label="Decision">{{ preview.result.policy.decision }}</ElDescriptionsItem>
        <ElDescriptionsItem label="Priority">{{ preview.result.policy.priority }}</ElDescriptionsItem>
      </ElDescriptions>
      <template #footer>
        <ElButton @click="preview.open = false">关闭</ElButton>
        <ElButton type="primary" :loading="preview.loading" @click="runPreview">执行预览</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import {
    ElAlert,
    ElButton,
    ElCard,
    ElDescriptions,
    ElDescriptionsItem,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElOption,
    ElPopconfirm,
    ElSelect,
    ElSwitch,
    ElTag,
    ElTooltip
  } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import type { components } from '@/api/v5/schema'
  import { useLogPolicies } from './modules/use-log-policies'

  defineOptions({ name: 'SystemLogPolicies' })

  type Pipeline = components['schemas']['LogPolicyItem']['pipeline']
  type MatchField = components['schemas']['LogPolicyItem']['match_field']
  type Decision = components['schemas']['LogPolicyItem']['decision']
  type LogPolicyItem = components['schemas']['LogPolicyItem']
  type LogPolicyPreviewResponse = components['schemas']['LogPolicyPreviewResponse']

  type EditorMode = 'create' | 'edit'

  const filters = reactive<{ pipeline: '' | Pipeline; enabled: '' | 'true' | 'false' }>({
    pipeline: '',
    enabled: ''
  })
  const loading = ref(false)
  const records = ref<LogPolicyItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)

  const editor = reactive<{
    open: boolean
    mode: EditorMode
    submitting: boolean
    editingID: string
    form: {
      pipeline: Pipeline
      match_field: MatchField
      pattern: string
      decision: Decision
      sample_rate: number
      priority: number
      enabled: boolean
      note: string
    }
  }>({
    open: false,
    mode: 'create',
    submitting: false,
    editingID: '',
    form: {
      pipeline: 'audit',
      match_field: 'action',
      pattern: '',
      decision: 'allow',
      sample_rate: 50,
      priority: 0,
      enabled: true,
      note: ''
    }
  })

  const preview = reactive<{
    open: boolean
    loading: boolean
    pipeline: Pipeline
    match_field: MatchField
    fieldValue: string
    result: LogPolicyPreviewResponse | null
  }>({
    open: false,
    loading: false,
    pipeline: 'audit',
    match_field: 'action',
    fieldValue: '',
    result: null
  })

  const pagination = computed(() => ({
    current: page.value,
    size: pageSize.value,
    total: total.value
  }))

  const matchFieldMap: Record<Pipeline, Array<{ label: string; value: MatchField }>> = {
    audit: [
      { label: 'action', value: 'action' },
      { label: 'outcome', value: 'outcome' },
      { label: 'resource_type', value: 'resource_type' }
    ],
    telemetry: [
      { label: 'level', value: 'level' },
      { label: 'event', value: 'event' },
      { label: 'route', value: 'route' }
    ]
  }

  const matchFieldOptions = computed(() => matchFieldMap[editor.form.pipeline])
  const policyActions = useLogPolicies()

  const columns = computed<ColumnOption[]>(() => [
    { prop: 'pipeline', label: 'Pipeline', width: 110 },
    { prop: 'match_field', label: 'Match Field', width: 140 },
    { prop: 'pattern', label: 'Pattern', minWidth: 220, showOverflowTooltip: true },
    {
      prop: 'decision',
      label: 'Decision',
      width: 160,
      formatter: (row: LogPolicyItem) => {
        const label =
          row.decision === 'sample' && row.sample_rate != null
            ? `sample (${row.sample_rate}%)`
            : row.decision
        const type = row.decision === 'deny' ? 'danger' : row.decision === 'sample' ? 'warning' : 'success'
        return h(ElTag, { type, effect: 'plain' }, () => label)
      }
    },
    { prop: 'priority', label: 'Priority', width: 100 },
    {
      prop: 'enabled',
      label: '状态',
      width: 100,
      formatter: (row: LogPolicyItem) =>
        h(
          ElTag,
          { type: row.enabled ? 'success' : 'info', effect: 'plain' },
          () => (row.enabled ? '启用' : '停用')
        )
    },
    {
      prop: 'compliance_locked',
      label: 'Lock',
      width: 120,
      formatter: (row: LogPolicyItem) =>
        row.compliance_locked ? h(ElTag, { type: 'warning', effect: 'plain' }, () => 'compliance') : '-'
    },
    { prop: 'updated_at', label: '更新时间', width: 190 },
    {
      prop: 'actions',
      label: '操作',
      width: 250,
      fixed: 'right',
      formatter: (row: LogPolicyItem) => {
        const editButton = row.compliance_locked
          ? h(
              ElTooltip,
              { content: 'compliance lock 策略不可编辑' },
              () => h(ElButton, { type: 'primary', link: true, disabled: true }, () => '编辑')
            )
          : h(ElButton, { type: 'primary', link: true, onClick: () => openEditDialog(row) }, () => '编辑')

        const deleteButton = row.compliance_locked
          ? h(
              ElTooltip,
              { content: 'compliance lock 策略不可删除' },
              () => h(ElButton, { type: 'danger', link: true, disabled: true }, () => '删除')
            )
          : h(
              ElPopconfirm,
              { title: '确认删除该策略？', onConfirm: () => removePolicy(row) },
              {
                reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除')
              }
            )

        return h('div', { class: 'policy-row-actions' }, [
          editButton,
          deleteButton,
          h(ElButton, { type: 'primary', link: true, onClick: () => openPreviewDialog(row) }, () => '预览')
        ])
      }
    }
  ])

  const buildQuery = () => {
    const query: Record<string, unknown> = {
      current: page.value,
      size: pageSize.value
    }
    if (filters.pipeline) query.pipeline = filters.pipeline
    if (filters.enabled === 'true') query.enabled = true
    if (filters.enabled === 'false') query.enabled = false
    return query as any
  }

  const load = async (targetPage?: number) => {
    if (targetPage) page.value = targetPage
    loading.value = true
    try {
      const data = await policyActions.list(buildQuery())
      records.value = data.records || []
      total.value = Number(data.total || 0)
    } catch (error: any) {
      ElMessage.error(error?.message || '加载日志策略失败')
    } finally {
      loading.value = false
    }
  }

  const resetFilters = () => {
    filters.pipeline = ''
    filters.enabled = ''
    load(1)
  }

  const handleSizeChange = (size: number) => {
    pageSize.value = size
    load(1)
  }

  const handleCurrentChange = (current: number) => {
    load(current)
  }

  const resetEditor = () => {
    editor.submitting = false
    editor.editingID = ''
    editor.mode = 'create'
    editor.form.pipeline = 'audit'
    editor.form.match_field = 'action'
    editor.form.pattern = ''
    editor.form.decision = 'allow'
    editor.form.sample_rate = 50
    editor.form.priority = 0
    editor.form.enabled = true
    editor.form.note = ''
  }

  const openCreateDialog = () => {
    resetEditor()
    editor.open = true
    editor.mode = 'create'
  }

  const openEditDialog = (row: LogPolicyItem) => {
    resetEditor()
    editor.open = true
    editor.mode = 'edit'
    editor.editingID = row.id
    editor.form.pipeline = row.pipeline
    editor.form.match_field = row.match_field
    editor.form.pattern = row.pattern
    editor.form.decision = row.decision
    editor.form.sample_rate = row.sample_rate ?? 50
    editor.form.priority = Number(row.priority || 0)
    editor.form.enabled = !!row.enabled
    editor.form.note = row.note || ''
  }

  const onPipelineChange = () => {
    const options = matchFieldMap[editor.form.pipeline]
    if (!options.find((item) => item.value === editor.form.match_field)) {
      editor.form.match_field = options[0].value
    }
  }

  const validateEditor = () => {
    if (!editor.form.pattern.trim()) {
      ElMessage.warning('pattern 不能为空')
      return false
    }
    if (editor.form.decision === 'sample') {
      if (!Number.isFinite(editor.form.sample_rate) || editor.form.sample_rate < 1 || editor.form.sample_rate > 100) {
        ElMessage.warning('sample_rate 必须在 1-100')
        return false
      }
    }
    return true
  }

  const submitEditor = async () => {
    if (!validateEditor()) return
    editor.submitting = true
    try {
      if (editor.mode === 'create') {
        await policyActions.create({
          pipeline: editor.form.pipeline,
          match_field: editor.form.match_field,
          pattern: editor.form.pattern.trim(),
          decision: editor.form.decision,
          sample_rate: editor.form.decision === 'sample' ? Number(editor.form.sample_rate) : undefined,
          priority: Number(editor.form.priority || 0),
          enabled: !!editor.form.enabled,
          note: editor.form.note.trim() || undefined
        })
        ElMessage.success('创建成功')
      } else {
        await policyActions.update(editor.editingID, {
          match_field: editor.form.match_field,
          pattern: editor.form.pattern.trim(),
          decision: editor.form.decision,
          sample_rate: editor.form.decision === 'sample' ? Number(editor.form.sample_rate) : null,
          priority: Number(editor.form.priority || 0),
          enabled: !!editor.form.enabled,
          note: editor.form.note.trim() || null
        })
        ElMessage.success('更新成功')
      }
      editor.open = false
      load(editor.mode === 'create' ? 1 : undefined)
    } catch (error: any) {
      ElMessage.error(error?.message || '保存失败')
    } finally {
      editor.submitting = false
    }
  }

  const removePolicy = async (row: LogPolicyItem) => {
    try {
      await policyActions.remove(row.id)
      ElMessage.success('删除成功')
      load(records.value.length === 1 && page.value > 1 ? page.value - 1 : undefined)
    } catch (error: any) {
      ElMessage.error(error?.message || '删除失败')
    }
  }

  const resetPreview = () => {
    preview.loading = false
    preview.pipeline = 'audit'
    preview.match_field = 'action'
    preview.fieldValue = ''
    preview.result = null
  }

  const openPreviewDialog = (row: LogPolicyItem) => {
    resetPreview()
    preview.open = true
    preview.pipeline = row.pipeline
    preview.match_field = row.match_field
    preview.fieldValue = row.pattern
  }

  const runPreview = async () => {
    if (!preview.fieldValue.trim()) {
      ElMessage.warning('请输入字段值')
      return
    }
    preview.loading = true
    try {
      preview.result = await policyActions.preview({
        pipeline: preview.pipeline,
        fields: {
          [preview.match_field]: preview.fieldValue.trim()
        }
      })
    } catch (error: any) {
      ElMessage.error(error?.message || '预览失败')
    } finally {
      preview.loading = false
    }
  }

  onMounted(() => load(1))
</script>

<style scoped>
  .log-policies-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .log-policies-main {
    flex: 1;
    min-height: 0;
  }

  .log-policies-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .log-policies-filters {
    margin-bottom: 4px;
  }

  .log-policies-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .log-policies-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .log-policies-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .policy-row-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .form-tip {
    margin-left: 12px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .preview-policy {
    margin-top: 12px;
  }
</style>
