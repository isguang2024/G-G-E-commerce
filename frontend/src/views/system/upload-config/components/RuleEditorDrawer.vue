<template>
  <ElDrawer
    v-model="visible"
    :title="editingId ? '编辑上传规则' : '新增上传规则'"
    size="600px"
    direction="rtl"
    destroy-on-close
    append-to-body
    @closed="onDrawerClosed"
  >
    <ElForm label-position="top" class="editor-form">
      <ElFormItem>
        <template #label>
          <FieldLabel
            label="规则标识"
            help="本规则在 UploadKey 下的唯一短名，例如 image、poster、attachment。落库后不可修改。"
            required
          />
        </template>
        <ElInput
          v-model="form.rule_key"
          :disabled="!!editingId"
          placeholder="如 image、file、poster"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel label="名称" help="规则的中文显示名。" required />
        </template>
        <ElInput v-model="form.name" placeholder="如 图片上传、附件上传" />
      </ElFormItem>

      <div class="config-guide-card">
        <div class="config-guide-title">规则的配置顺序</div>
        <ol class="config-guide-list">
          <li>先定义子路径、文件名策略和大小限制，明确本规则与 UploadKey 的差异点。</li>
          <li>再决定是否覆写上传方式、可见性和前端选择提示。</li>
          <li>最后再补扩展参数和默认规则标记，用于更细的业务场景区分。</li>
        </ol>
      </div>

      <div class="config-section-title">基础规则</div>
      <div class="config-section-tip">这些字段决定落盘位置、文件名和基础文件限制。</div>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="子路径"
            help="在 UploadKey 路径模板之后追加的子目录,例如图片规则填 img,附件规则填 file。最终键=桶 base_path + UploadKey 模板 + 本子路径 + 文件名。"
          />
        </template>
        <ElInput v-model="form.sub_path" placeholder="可选，追加到上传配置路径之后" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="文件名策略"
            help="UUID 随机最安全,避免冲突；original 保留原文件名,会做安全清洗；timestamp 按时间排序；hashed 相同内容落同一对象键。"
            required
          />
        </template>
        <DictSelect
          v-model="form.filename_strategy"
          code="upload_filename_strategy"
          :clearable="false"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="单文件上限（字节）"
            help="0 表示沿用上传配置的上限。1MB ≈ 1048576。"
          />
        </template>
        <ElInputNumber
          v-model="form.max_size_bytes"
          :min="0"
          controls-position="right"
          style="width: 260px"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="允许文件类型"
            help="作用于 prepare/upload 两端的 MIME 校验。支持 image/* 通配。可直接输入未在字典里的值。"
          />
        </template>
        <DictSelect
          v-model="form.allowed_mime_types"
          code="upload_mime_preset"
          multiple
          allow-create
          placeholder="选择或直接输入；留空表示不限"
        />
      </ElFormItem>

      <div class="config-section-title">覆写策略</div>
      <div class="config-section-tip">只有与 UploadKey 默认行为不同的地方才需要在这里覆写。</div>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="前端文件选择提示"
            help="仅作用于浏览器文件选择器；留空则继承 UploadKey 的 client_accept。规则层优先级高于上传配置层。"
          />
        </template>
        <DictSelect
          v-model="form.client_accept"
          code="upload_mime_preset"
          multiple
          allow-create
          placeholder="可选，细分规则的前端选择提示"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="上传方式覆写"
            help="inherit 保持与 UploadKey 一致；direct/relay 强制覆写本规则的上传通道。"
            required
          />
        </template>
        <DictSelect
          v-model="form.mode_override"
          code="upload_mode_override"
          :clearable="false"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="可见性覆写"
            help="inherit 保持与 UploadKey 一致；public/private 强制覆写。"
            required
          />
        </template>
        <DictSelect
          v-model="form.visibility_override"
          code="upload_visibility_override"
          :clearable="false"
        />
      </ElFormItem>

      <div class="config-section-title">扩展与默认</div>
      <div class="config-section-tip"
        >通过扩展参数定义本规则特有的附加字段；默认规则标记用于 prepare 时未指定 rule 的兜底。</div
      >

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="扩展参数定义"
            help="字段化地描述规则特有的附加参数（如图片变体、水印文字等）。"
          />
        </template>
        <ExtraSchemaEditor
          ref="schemaEditorRef"
          :model-value="form.extra_schema"
          title="Rule 自定义参数"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="设为默认规则"
            help="prepare 调用未传 rule_key 时自动用此规则；每个 UploadKey 最多一条默认规则。"
          />
        </template>
        <ElSwitch v-model="form.is_default" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="状态"
            help="停用后本规则不再被选中；异常通常由系统自动标记。"
            required
          />
        </template>
        <DictSelect v-model="form.status" code="storage_status" :clearable="false" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="visible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="onSubmit">保存</ElButton>
      </div>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue'
  import {
    ElButton,
    ElDrawer,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElSwitch
  } from 'element-plus'
  import FieldLabel from '@/components/business/common/FieldLabel.vue'
  import DictSelect from '@/components/business/dictionary/DictSelect.vue'
  import ExtraSchemaEditor from '@/domains/upload-config/components/ExtraSchemaEditor.vue'
  import {
    fetchCreateUploadKeyRule,
    fetchUpdateUploadKeyRule,
    type UploadKeyRuleSaveRequest,
    type UploadKeyRuleSummary
  } from '@/domains/upload-config/api'
  import { normalizeObjectValue } from './_shared'

  defineOptions({ name: 'RuleEditorDrawer' })

  const props = defineProps<{
    open: boolean
    editingId: string
    row: UploadKeyRuleSummary | null
    parentUploadKeyId: string
  }>()

  const emit = defineEmits<{
    'update:open': [value: boolean]
    saved: []
  }>()

  const visible = computed({
    get: () => props.open,
    set: (v) => emit('update:open', v)
  })

  const submitting = ref(false)
  const schemaEditorRef = ref<InstanceType<typeof ExtraSchemaEditor> | null>(null)

  const form = reactive<{
    rule_key: string
    name: string
    sub_path: string
    filename_strategy: Exclude<UploadKeyRuleSaveRequest['filename_strategy'], undefined>
    max_size_bytes: number
    allowed_mime_types: string[]
    process_pipeline: string[]
    mode_override: Exclude<UploadKeyRuleSaveRequest['mode_override'], undefined>
    visibility_override: Exclude<UploadKeyRuleSaveRequest['visibility_override'], undefined>
    client_accept: string[]
    extra_schema: UploadKeyRuleSaveRequest['extra_schema'] | undefined
    is_default: boolean
    status: Exclude<UploadKeyRuleSaveRequest['status'], undefined>
  }>({
    rule_key: '',
    name: '',
    sub_path: '',
    filename_strategy: 'uuid',
    max_size_bytes: 0,
    allowed_mime_types: [],
    process_pipeline: [],
    mode_override: 'inherit',
    visibility_override: 'inherit',
    client_accept: [],
    extra_schema: undefined,
    is_default: false,
    status: 'ready'
  })

  const editingId = computed(() => props.editingId)

  watch(
    () => props.open,
    (open) => {
      if (!open) return
      if (props.editingId && props.row) {
        hydrateFromRow(props.row)
      } else {
        resetForm()
      }
    },
    { immediate: true }
  )

  function resetForm() {
    form.rule_key = ''
    form.name = ''
    form.sub_path = ''
    form.filename_strategy = 'uuid'
    form.max_size_bytes = 0
    form.allowed_mime_types = []
    form.process_pipeline = []
    form.mode_override = 'inherit'
    form.visibility_override = 'inherit'
    form.client_accept = []
    form.extra_schema = undefined
    form.is_default = false
    form.status = 'ready'
  }

  function hydrateFromRow(row: UploadKeyRuleSummary) {
    form.rule_key = row.rule_key
    form.name = row.name
    form.sub_path = row.sub_path || ''
    form.filename_strategy = row.filename_strategy
    form.max_size_bytes = Number(row.max_size_bytes ?? 0)
    form.allowed_mime_types = Array.isArray(row.allowed_mime_types)
      ? [...row.allowed_mime_types]
      : []
    form.process_pipeline = Array.isArray(row.process_pipeline)
      ? [...row.process_pipeline]
      : []
    form.mode_override = row.mode_override || 'inherit'
    form.visibility_override = row.visibility_override || 'inherit'
    form.client_accept = Array.isArray(row.client_accept) ? [...row.client_accept] : []
    form.extra_schema = normalizeObjectValue(row.extra_schema)
    form.is_default = !!row.is_default
    form.status = row.status
  }

  function onDrawerClosed() {
    submitting.value = false
  }

  function buildBody(): UploadKeyRuleSaveRequest | null {
    const schemaResult = schemaEditorRef.value?.buildSchema()
    if (schemaResult?.error) {
      ElMessage.warning(`Rule ${schemaResult.error}`)
      return null
    }
    const body: UploadKeyRuleSaveRequest = {
      rule_key: form.rule_key.trim(),
      name: form.name.trim(),
      filename_strategy: form.filename_strategy,
      mode_override: form.mode_override,
      visibility_override: form.visibility_override,
      is_default: form.is_default,
      status: form.status,
      allowed_mime_types: form.allowed_mime_types,
      client_accept: form.client_accept
    }
    if (form.sub_path.trim()) body.sub_path = form.sub_path.trim()
    if (Number.isFinite(form.max_size_bytes) && form.max_size_bytes > 0) {
      body.max_size_bytes = Number(form.max_size_bytes)
    }
    if (form.process_pipeline.length > 0) body.process_pipeline = form.process_pipeline
    if (schemaResult?.value) body.extra_schema = schemaResult.value
    return body
  }

  async function onSubmit() {
    const body = buildBody()
    if (!body) return
    if (!body.rule_key || !body.name) {
      ElMessage.warning('规则标识和名称必填')
      return
    }
    submitting.value = true
    try {
      if (editingId.value) {
        await fetchUpdateUploadKeyRule(editingId.value, body)
        ElMessage.success('上传规则已更新')
      } else {
        await fetchCreateUploadKeyRule(props.parentUploadKeyId, body)
        ElMessage.success('上传规则已创建')
      }
      emit('saved')
      visible.value = false
    } catch (err: any) {
      ElMessage.error(err?.message || '保存规则失败')
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped lang="scss">
  .editor-form {
    padding-right: 4px;
  }

  .config-guide-card {
    margin: 6px 0 14px;
    padding: 12px 14px;
    background: linear-gradient(135deg, var(--el-fill-color-light) 0%, var(--el-fill-color) 100%);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
  }

  .config-guide-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .config-guide-list {
    margin: 8px 0 0;
    padding-left: 18px;
    color: var(--el-text-color-regular);
    line-height: 1.8;
  }

  .config-section-title {
    margin: 12px 0 4px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .config-section-tip {
    margin-bottom: 10px;
    font-size: 12px;
    line-height: 1.7;
    color: var(--el-text-color-secondary);
  }

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
</style>
