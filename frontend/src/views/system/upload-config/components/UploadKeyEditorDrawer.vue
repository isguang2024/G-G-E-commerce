<template>
  <ElDrawer
    v-model="visible"
    :title="editingId ? '编辑上传配置' : '新增上传配置'"
    size="680px"
    direction="rtl"
    destroy-on-close
    @closed="onDrawerClosed"
  >
    <ElForm label-position="top" class="editor-form">
      <ElFormItem>
        <template #label>
          <FieldLabel
            label="所属存储桶"
            help="这个业务场景的文件会落到哪个桶。创建后不能切换。"
            required
          />
        </template>
        <ElSelect
          v-model="form.bucket_id"
          :disabled="!!editingId"
          filterable
          style="width: 100%"
        >
          <ElOption
            v-for="b in buckets"
            :key="b.id"
            :label="`${b.name}（${b.bucket_key}）`"
            :value="b.id"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="上传标识"
            help="业务代码里指定上传场景的唯一短名,例如 avatar、attachment、editor.inline。创建后不可修改。"
            required
          />
        </template>
        <ElInput
          v-model="form.key"
          :disabled="!!editingId"
          placeholder="如 avatar、attachment、public-asset"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel label="名称" help="面向后台运营同学的友好显示名。" required />
        </template>
        <ElInput v-model="form.name" placeholder="便于识别的显示名称" />
      </ElFormItem>

      <div class="config-guide-card">
        <div class="config-guide-title">上传配置的配置顺序</div>
        <ol class="config-guide-list">
          <li>先确定存储路径和默认规则，明确这个业务场景落到哪个桶、默认走哪条规则。</li>
          <li>再决定上传方式、前端是否可见、是否公开访问以及直传阈值。</li>
          <li>最后补权限键、回退标识和自定义参数，让业务侧知道哪些附加字段可配置。</li>
        </ol>
      </div>

      <div class="config-section-title">基础路由</div>
      <div class="config-section-tip">决定对象最终写入哪个目录，以及默认使用哪条子规则。</div>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="路径模板"
            help="对象键的中段前缀。支持变量 {tenant} {key} {date} {yyyy} {mm} {dd} {hh} {uuid} {ext}。最终路径 = 桶 base_path + 本模板 + 规则 sub_path + 文件名策略。"
          />
        </template>
        <ElInput
          v-model="form.path_template"
          placeholder="{tenant}/{key}/{date}/{uuid}{ext}"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="默认规则标识"
            help="不传规则时使用的 rule_key；留空则用本配置下标记为默认的规则。"
          />
        </template>
        <ElInput
          v-model="form.default_rule_key"
          placeholder="可选，留空则使用标记为默认的规则"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="单文件上限（字节）"
            help="0 表示沿用存储桶或全局上限。1MB ≈ 1048576。超限会在 prepare 阶段拒绝。"
          />
        </template>
        <ElInputNumber
          v-model="form.max_size_bytes"
          :min="0"
          controls-position="right"
          style="width: 260px"
        />
      </ElFormItem>

      <div class="config-section-title">运行时策略</div>
      <div class="config-section-tip"
        >这里决定前端能否看到该场景，以及实际走直传、后端中转还是自动选择。</div
      >

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="允许文件类型"
            help="列表中的 MIME 会在 prepare/upload 两端同时校验。支持 image/* 通配。可以直接输入未在字典里的类型。"
          />
        </template>
        <DictSelect
          v-model="form.allowed_mime_types"
          code="upload_mime_preset"
          multiple
          allow-create
          placeholder="选择或直接输入，如 image/*、video/mp4；留空表示不限"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="前端文件选择提示"
            help="仅作用于浏览器 <input accept=...> 的候选过滤，不替代后端 MIME 校验。"
          />
        </template>
        <DictSelect
          v-model="form.client_accept"
          code="upload_mime_preset"
          multiple
          allow-create
          placeholder="与上方允许类型相同或更严格，用于前端选择器提示"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="上传方式"
            help="auto 会按驱动能力自动选择；direct 要求驱动支持签名直传；relay 由后端中转,兼容性最好。"
            required
          />
        </template>
        <DictSelect v-model="form.upload_mode" code="upload_mode" :clearable="false" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="直传阈值（字节）"
            help="文件大于该阈值时强制改走后端中转，避免前端直传大文件超时。0 表示不启用阈值。"
          />
        </template>
        <ElInputNumber
          v-model="form.direct_size_threshold_bytes"
          :min="0"
          controls-position="right"
          style="width: 260px"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="可见性"
            help="public 允许匿名 URL 直连（CDN 场景）；private 需要签名 URL 或业务接口才能访问。"
            required
          />
        </template>
        <DictSelect
          v-model="form.visibility"
          code="upload_visibility"
          :clearable="false"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="前端可见"
            help="开启后,本 UploadKey 会出现在业务前端通过 /media/upload-keys 拉取的可见上传场景列表里。"
          />
        </template>
        <ElSwitch v-model="form.is_frontend_visible" />
      </ElFormItem>

      <div class="config-section-title">访问控制与扩展</div>
      <div class="config-section-tip"
        >权限键决定谁能调用这个场景；回退标识用于兜底；扩展参数用于描述业务自定义字段。</div
      >

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="权限键"
            help="运行时按此权限键过滤是否允许调用本场景；留空表示登录即可使用。如 cms.asset.upload。"
          />
        </template>
        <ElInput
          v-model="form.permission_key"
          placeholder="如 cms.asset.upload，留空表示登录即可使用"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="回退上传标识"
            help="主配置不可用时尝试回退到另一个 UploadKey；用于灰度/灾备。目标必须存在。"
          />
        </template>
        <ElInput
          v-model="form.fallback_key"
          placeholder="可选，主配置不可用时尝试回退到的 UploadKey 标识"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="扩展参数定义"
            help="字段化地描述业务可以额外传入的参数（如水印文字、业务标签等）,供配置人员和前端一目了然。"
          />
        </template>
        <ExtraSchemaEditor
          ref="schemaEditorRef"
          :model-value="form.extra_schema"
          title="UploadKey 自定义参数"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="状态"
            help="停用后本 UploadKey 不再参与运行时解析；异常通常由系统健康检查自动标记。"
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
    ElOption,
    ElSelect,
    ElSwitch
  } from 'element-plus'
  import FieldLabel from '@/components/business/common/FieldLabel.vue'
  import DictSelect from '@/components/business/dictionary/DictSelect.vue'
  import ExtraSchemaEditor from '@/domains/upload-config/components/ExtraSchemaEditor.vue'
  import {
    fetchCreateUploadKey,
    fetchUpdateUploadKey,
    type StorageBucketSummary,
    type UploadKeySaveRequest,
    type UploadKeySummary
  } from '@/domains/upload-config/api'
  import { normalizeObjectValue } from './_shared'

  defineOptions({ name: 'UploadKeyEditorDrawer' })

  const props = defineProps<{
    open: boolean
    editingId: string
    row: UploadKeySummary | null
    buckets: StorageBucketSummary[]
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
    bucket_id: string
    key: string
    name: string
    path_template: string
    default_rule_key: string
    max_size_bytes: number
    allowed_mime_types: string[]
    upload_mode: Exclude<UploadKeySaveRequest['upload_mode'], undefined>
    is_frontend_visible: boolean
    permission_key: string
    fallback_key: string
    client_accept: string[]
    direct_size_threshold_bytes: number
    extra_schema: UploadKeySaveRequest['extra_schema'] | undefined
    visibility: Exclude<UploadKeySaveRequest['visibility'], undefined>
    status: Exclude<UploadKeySaveRequest['status'], undefined>
  }>({
    bucket_id: '',
    key: '',
    name: '',
    path_template: '',
    default_rule_key: '',
    max_size_bytes: 0,
    allowed_mime_types: [],
    upload_mode: 'auto',
    is_frontend_visible: false,
    permission_key: '',
    fallback_key: '',
    client_accept: [],
    direct_size_threshold_bytes: 0,
    extra_schema: undefined,
    visibility: 'private',
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
    form.bucket_id = props.buckets[0]?.id || ''
    form.key = ''
    form.name = ''
    form.path_template = ''
    form.default_rule_key = ''
    form.max_size_bytes = 0
    form.allowed_mime_types = []
    form.upload_mode = 'auto'
    form.is_frontend_visible = false
    form.permission_key = ''
    form.fallback_key = ''
    form.client_accept = []
    form.direct_size_threshold_bytes = 0
    form.extra_schema = undefined
    form.visibility = 'private'
    form.status = 'ready'
  }

  function hydrateFromRow(row: UploadKeySummary) {
    form.bucket_id = row.bucket_id
    form.key = row.key
    form.name = row.name
    form.path_template = row.path_template || ''
    form.default_rule_key = row.default_rule_key || ''
    form.max_size_bytes = Number(row.max_size_bytes ?? 0)
    form.allowed_mime_types = Array.isArray(row.allowed_mime_types)
      ? [...row.allowed_mime_types]
      : []
    form.upload_mode = row.upload_mode || 'auto'
    form.is_frontend_visible = !!row.is_frontend_visible
    form.permission_key = row.permission_key || ''
    form.fallback_key = row.fallback_key || ''
    form.client_accept = Array.isArray(row.client_accept) ? [...row.client_accept] : []
    form.direct_size_threshold_bytes = Number(row.direct_size_threshold_bytes ?? 0)
    form.extra_schema = normalizeObjectValue(row.extra_schema)
    form.visibility = row.visibility
    form.status = row.status
  }

  function onDrawerClosed() {
    submitting.value = false
  }

  function buildBody(): UploadKeySaveRequest | null {
    const schemaResult = schemaEditorRef.value?.buildSchema()
    if (schemaResult?.error) {
      ElMessage.warning(`UploadKey ${schemaResult.error}`)
      return null
    }
    const body: UploadKeySaveRequest = {
      bucket_id: form.bucket_id,
      key: form.key.trim(),
      name: form.name.trim(),
      upload_mode: form.upload_mode,
      is_frontend_visible: form.is_frontend_visible,
      visibility: form.visibility,
      status: form.status,
      allowed_mime_types: form.allowed_mime_types,
      client_accept: form.client_accept
    }
    if (form.path_template.trim()) body.path_template = form.path_template.trim()
    if (form.default_rule_key.trim()) body.default_rule_key = form.default_rule_key.trim()
    if (form.permission_key.trim()) body.permission_key = form.permission_key.trim()
    if (form.fallback_key.trim()) body.fallback_key = form.fallback_key.trim()
    if (Number.isFinite(form.max_size_bytes) && form.max_size_bytes > 0) {
      body.max_size_bytes = Number(form.max_size_bytes)
    }
    if (
      Number.isFinite(form.direct_size_threshold_bytes) &&
      form.direct_size_threshold_bytes > 0
    ) {
      body.direct_size_threshold_bytes = Number(form.direct_size_threshold_bytes)
    }
    if (schemaResult?.value) body.extra_schema = schemaResult.value
    return body
  }

  async function onSubmit() {
    const body = buildBody()
    if (!body) return
    if (!body.bucket_id || !body.key || !body.name) {
      ElMessage.warning('所属存储桶、上传标识、名称必填')
      return
    }
    submitting.value = true
    try {
      if (editingId.value) {
        await fetchUpdateUploadKey(editingId.value, body)
        ElMessage.success('上传配置已更新')
      } else {
        await fetchCreateUploadKey(body)
        ElMessage.success('上传配置已创建')
      }
      emit('saved')
      visible.value = false
    } catch (err: any) {
      ElMessage.error(err?.message || '保存上传配置失败')
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
