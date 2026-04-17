<template>
  <ElDrawer
    v-model="visible"
    :title="editingId ? '编辑存储桶' : '新增存储桶'"
    size="560px"
    direction="rtl"
    destroy-on-close
    @closed="onDrawerClosed"
  >
    <ElForm label-position="top" class="editor-form">
      <ElFormItem>
        <template #label>
          <FieldLabel
            label="所属存储服务"
            help="决定这个桶落到哪个底层驱动实例。创建后不能切换服务。"
            required
          />
        </template>
        <ElSelect
          v-model="form.provider_id"
          :disabled="!!editingId"
          filterable
          style="width: 100%"
          @change="onProviderChange"
        >
          <ElOption
            v-for="p in providers"
            :key="p.id"
            :label="`${p.name}（${p.provider_key}）`"
            :value="p.id"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="存储桶标识"
            help="桶在业务代码里的唯一短名。用英文小写+连字符。落库后不可修改。"
            required
          />
        </template>
        <ElInput
          v-model="form.bucket_key"
          :disabled="!!editingId"
          placeholder="如 default-bucket"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel label="名称" help="用于后台显示的友好名称。" required />
        </template>
        <ElInput v-model="form.name" placeholder="便于识别的显示名称" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="存储桶名称"
            help="对应对象存储服务中实际的 Bucket 名（云 OSS 必填）；本地存储也要填一个逻辑名。"
            required
          />
        </template>
        <ElInput
          v-model="form.bucket_name"
          placeholder="对象存储中实际的 Bucket 名称"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="基础路径"
            help="所有落到本桶的对象都会加这个前缀目录。例如 public-media/。路径模板会拼在它后面。"
          />
        </template>
        <ElInput v-model="form.base_path" placeholder="可选，文件存储的前缀目录" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="公网访问地址"
            help="访问已上传对象的公网根 URL。留空则继承存储服务的基础访问地址。"
          />
        </template>
        <ElInput
          v-model="form.public_base_url"
          placeholder="访问已上传文件用的公网根地址"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="公开访问"
            help="开启后,桶中对象可被匿名访问（适合静态资源 + CDN）。关闭则必须通过签名 URL 或业务接口。"
          />
        </template>
        <ElSwitch v-model="form.is_public" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="状态"
            help="停用后运行时不再选中本桶；异常通常由健康检查自动标记。"
            required
          />
        </template>
        <DictSelect v-model="form.status" code="storage_status" :clearable="false" />
      </ElFormItem>

      <div class="driver-guide-card">
        <div class="driver-guide-head">
          <div class="driver-guide-title">
            当前驱动：{{ driverLabel[driver] || driver || '未选择存储服务' }}
          </div>
          <ElButton
            link
            type="primary"
            :disabled="!driver"
            @click="restoreDriverDefaults"
            >恢复推荐默认值</ElButton
          >
        </div>
        <ul v-if="driverGuide.length" class="driver-guide-list">
          <li v-for="item in driverGuide" :key="item">{{ item }}</li>
        </ul>
        <div v-else class="driver-section-desc"
          >请先选择存储服务，再配置当前桶的驱动扩展参数。</div
        >
      </div>

      <ElCollapse v-model="activePanels" class="driver-collapse">
        <ElCollapseItem
          v-for="section in driverSections"
          :key="section.key"
          :name="section.key"
          :title="section.title"
        >
          <div v-if="section.description" class="driver-section-desc">
            {{ section.description }}
          </div>
          <ElFormItem v-for="field in section.fields" :key="field.key">
            <template #label>
              <FieldLabel :label="field.label" :help="formatDriverExtraFieldTip(field)" />
            </template>
            <ElSwitch
              v-if="field.type === 'boolean'"
              :model-value="readExtraBooleanValue(form.extra, field.key)"
              @update:model-value="setExtraValue(form.extra, field.key, $event)"
            />
            <ElInputNumber
              v-else-if="field.type === 'number'"
              :model-value="readExtraNumberValue(form.extra, field.key)"
              :min="field.min ?? 0"
              :step="field.step ?? 1"
              controls-position="right"
              style="width: 240px"
              @update:model-value="setExtraValue(form.extra, field.key, $event)"
            />
            <ElInput
              v-else-if="field.type === 'object'"
              :model-value="objectExtraText[field.key] || ''"
              type="textarea"
              :autosize="{ minRows: field.rows ?? 4, maxRows: Math.max(field.rows ?? 4, 8) }"
              :placeholder="field.placeholder"
              @update:model-value="setExtraValue(objectExtraText, field.key, $event)"
            />
            <ElInput
              v-else
              :model-value="readExtraStringValue(form.extra, field.key)"
              :type="field.multiline ? 'textarea' : 'text'"
              :autosize="
                field.multiline
                  ? { minRows: field.rows ?? 3, maxRows: Math.max(field.rows ?? 3, 8) }
                  : undefined
              "
              :placeholder="field.placeholder"
              @update:model-value="setExtraValue(form.extra, field.key, $event)"
            />
          </ElFormItem>
        </ElCollapseItem>
        <ElCollapseItem name="custom" title="自定义扩展参数">
          <div class="driver-section-desc">
            这里保留给驱动新增键或内部约定参数，避免每次扩展都要等页面改版。
          </div>
          <ElFormItem>
            <template #label>
              <FieldLabel
                label="附加参数 JSON"
                help="用 JSON 对象补充驱动新增参数或内部约定字段；键名不能与上方已结构化字段重复。"
              />
            </template>
            <ElInput
              v-model="customExtraText"
              type="textarea"
              :autosize="{ minRows: 4, maxRows: 10 }"
              placeholder='{&#10;  "origin_access_identity": "private"&#10;}'
            />
          </ElFormItem>
        </ElCollapseItem>
      </ElCollapse>
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
    ElCollapse,
    ElCollapseItem,
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
  import {
    fetchCreateStorageBucket,
    fetchUpdateStorageBucket,
    type StorageBucketSaveRequest,
    type StorageBucketSummary,
    type StorageProviderSummary
  } from '@/domains/upload-config/api'
  import {
    getDriverExtraSections,
    getDriverGuide,
    type StorageDriver
  } from '@/domains/upload-config/driver-extra-registry'
  import {
    buildDriverExtraBody,
    buildDriverExtraDraft,
    driverLabel,
    formatDriverExtraFieldTip,
    readExtraBooleanValue,
    readExtraNumberValue,
    readExtraStringValue,
    setExtraValue,
    type DriverExtraValueMap
  } from './_shared'

  defineOptions({ name: 'BucketEditorDrawer' })

  const props = defineProps<{
    open: boolean
    editingId: string
    row: StorageBucketSummary | null
    providers: StorageProviderSummary[]
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
  const objectExtraText = ref<Record<string, string>>({})
  const customExtraText = ref('')
  const activePanels = ref<string[]>([])

  const form = reactive<{
    provider_id: string
    bucket_key: string
    name: string
    bucket_name: string
    base_path: string
    public_base_url: string
    extra: DriverExtraValueMap
    is_public: boolean
    status: Exclude<StorageBucketSaveRequest['status'], undefined>
  }>({
    provider_id: '',
    bucket_key: '',
    name: '',
    bucket_name: '',
    base_path: '',
    public_base_url: '',
    extra: {},
    is_public: false,
    status: 'ready'
  })

  const driver = computed<StorageDriver | ''>(
    () => props.providers.find((p) => p.id === form.provider_id)?.driver || ''
  )
  const driverGuide = computed(() => getDriverGuide(driver.value, 'bucket'))
  const driverSections = computed(() => getDriverExtraSections(driver.value, 'bucket'))

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
    form.provider_id = props.providers[0]?.id || ''
    form.bucket_key = ''
    form.name = ''
    form.bucket_name = ''
    form.base_path = ''
    form.public_base_url = ''
    form.is_public = false
    form.status = 'ready'
    applyDriverExtra(driver.value)
  }

  function hydrateFromRow(row: StorageBucketSummary) {
    form.provider_id = row.provider_id
    form.bucket_key = row.bucket_key
    form.name = row.name
    form.bucket_name = row.bucket_name
    form.base_path = row.base_path || ''
    form.public_base_url = row.public_base_url || ''
    form.is_public = !!row.is_public
    form.status = row.status
    applyDriverExtra(driver.value, row.extra)
  }

  function applyDriverExtra(d: StorageDriver | '' | undefined, value?: unknown) {
    const draft = buildDriverExtraDraft(d, 'bucket', value)
    form.extra = draft.values
    objectExtraText.value = draft.objectText
    customExtraText.value = draft.customText
    activePanels.value = draft.activePanels
  }

  function onProviderChange() {
    applyDriverExtra(driver.value)
  }

  function restoreDriverDefaults() {
    applyDriverExtra(driver.value)
  }

  function onDrawerClosed() {
    submitting.value = false
  }

  function buildBody(): StorageBucketSaveRequest | null {
    const extra = buildDriverExtraBody(
      '存储桶',
      driver.value,
      'bucket',
      form.extra,
      objectExtraText.value,
      customExtraText.value
    )
    if (extra === null) return null
    const body: StorageBucketSaveRequest = {
      provider_id: form.provider_id,
      bucket_key: form.bucket_key.trim(),
      name: form.name.trim(),
      bucket_name: form.bucket_name.trim(),
      is_public: form.is_public,
      status: form.status
    }
    if (form.base_path.trim()) body.base_path = form.base_path.trim()
    if (form.public_base_url.trim()) body.public_base_url = form.public_base_url.trim()
    if (extra) body.extra = extra
    return body
  }

  async function onSubmit() {
    const body = buildBody()
    if (!body) return
    if (!body.provider_id || !body.bucket_key || !body.name || !body.bucket_name) {
      ElMessage.warning('所属存储服务、存储桶标识、名称、存储桶名称均为必填')
      return
    }
    submitting.value = true
    try {
      if (editingId.value) {
        await fetchUpdateStorageBucket(editingId.value, body)
        ElMessage.success('存储桶已更新')
      } else {
        await fetchCreateStorageBucket(body)
        ElMessage.success('存储桶已创建')
      }
      emit('saved')
      visible.value = false
    } catch (err: any) {
      ElMessage.error(err?.message || '保存存储桶失败')
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped lang="scss">
  .editor-form {
    padding-right: 4px;
  }

  .driver-guide-card {
    margin: 6px 0 14px;
    padding: 12px 14px;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
  }

  .driver-guide-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
  }

  .driver-guide-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .driver-guide-list {
    margin: 0;
    padding-left: 18px;
    color: var(--el-text-color-regular);
    line-height: 1.8;
  }

  .driver-collapse {
    margin-top: 6px;
  }

  .driver-section-desc {
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
