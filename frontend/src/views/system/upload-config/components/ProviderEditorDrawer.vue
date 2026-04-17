<template>
  <ElDrawer
    v-model="visible"
    :title="editingId ? '编辑存储服务' : '新增存储服务'"
    size="560px"
    direction="rtl"
    destroy-on-close
    @closed="onDrawerClosed"
  >
    <ElForm label-position="top" class="editor-form">
      <ElFormItem>
        <template #label>
          <FieldLabel
            label="服务标识"
            help="驱动实例的唯一短标识，落库后不可修改。建议用英文小写+连字符，例如 local-default、oss-prod。"
            required
          />
        </template>
        <ElInput
          v-model="form.provider_key"
          :disabled="!!editingId"
          placeholder="如 local-default、oss-prod"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel label="名称" help="面向后台运营同学的显示名。" required />
        </template>
        <ElInput v-model="form.name" placeholder="便于识别的显示名称" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="驱动类型"
            help="决定底层用哪种对象存储实现。local 适合本地/内网；aliyun_oss 支持 CDN、前端直传、STS。"
            required
          />
        </template>
        <DictSelect
          v-model="form.driver"
          code="storage_driver"
          :clearable="false"
          :disabled="!!editingId"
          @change="onDriverChange"
        />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="基础访问地址"
            help="文件公网访问根地址，通常是 CDN 域名。留空则使用存储桶自身配置。"
          />
        </template>
        <ElInput v-model="form.base_url" placeholder="如 https://cdn.example.com" />
      </ElFormItem>

      <template v-if="form.driver === 'aliyun_oss'">
        <ElFormItem>
          <template #label>
            <FieldLabel
              label="接入点地址"
              help="阿里云 OSS 接入域名，通常是 oss-<region>.aliyuncs.com。"
            />
          </template>
          <ElInput v-model="form.endpoint" placeholder="如 oss-cn-hangzhou.aliyuncs.com" />
        </ElFormItem>
        <ElFormItem>
          <template #label>
            <FieldLabel label="地域" help="OSS Bucket 所在地域，用于签名与路由。" />
          </template>
          <ElInput v-model="form.region" placeholder="如 cn-hangzhou" />
        </ElFormItem>
        <ElFormItem>
          <template #label>
            <FieldLabel
              label="访问密钥（AK）"
              help="AccessKeyId；编辑时留空表示保留原值。生产环境建议使用 STS 临时密钥而非长期 AK。"
            />
          </template>
          <ElInput
            v-model="form.access_key"
            placeholder="留空表示保留原值"
            autocomplete="off"
          />
        </ElFormItem>
        <ElFormItem>
          <template #label>
            <FieldLabel
              label="安全密钥（SK）"
              help="AccessKeySecret；编辑时留空表示保留原值。列表中只会展示掩码。"
            />
          </template>
          <ElInput
            v-model="form.secret_key"
            type="password"
            show-password
            placeholder="留空表示保留原值"
            autocomplete="new-password"
          />
        </ElFormItem>
      </template>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="设为默认"
            help="开启后，没有显式指定存储服务的业务场景会优先使用本服务。"
          />
        </template>
        <ElSwitch v-model="form.is_default" />
      </ElFormItem>

      <ElFormItem>
        <template #label>
          <FieldLabel
            label="状态"
            help="停用后运行时不再选中此服务；异常通常由健康检查自动标记。"
            required
          />
        </template>
        <DictSelect v-model="form.status" code="storage_status" :clearable="false" />
      </ElFormItem>

      <div class="driver-guide-card">
        <div class="driver-guide-head">
          <div class="driver-guide-title">
            当前驱动：{{ driverLabel[form.driver] || form.driver }}
          </div>
          <ElButton link type="primary" @click="restoreDriverDefaults">恢复推荐默认值</ElButton>
        </div>
        <ul class="driver-guide-list">
          <li v-for="item in driverGuide" :key="item">{{ item }}</li>
        </ul>
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
            用于承载当前表单未结构化覆盖的驱动扩展键；键名不能与上方已结构化字段重复。
          </div>
          <ElFormItem>
            <template #label>
              <FieldLabel
                label="附加参数 JSON"
                help="以 JSON 对象形式补充驱动新增参数或内部定制能力。例如 {&quot;custom_endpoint_policy&quot;: &quot;internal-only&quot;}。"
              />
            </template>
            <ElInput
              v-model="customExtraText"
              type="textarea"
              :autosize="{ minRows: 4, maxRows: 10 }"
              placeholder='{&#10;  "custom_endpoint_policy": "internal-only"&#10;}'
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
    ElSwitch
  } from 'element-plus'
  import FieldLabel from '@/components/business/common/FieldLabel.vue'
  import DictSelect from '@/components/business/dictionary/DictSelect.vue'
  import {
    fetchCreateStorageProvider,
    fetchUpdateStorageProvider,
    type StorageProviderSaveRequest,
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

  defineOptions({ name: 'ProviderEditorDrawer' })

  const props = defineProps<{
    open: boolean
    editingId: string
    row: StorageProviderSummary | null
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
    provider_key: string
    name: string
    driver: StorageDriver
    endpoint: string
    region: string
    base_url: string
    access_key: string
    secret_key: string
    extra: DriverExtraValueMap
    is_default: boolean
    status: Exclude<StorageProviderSaveRequest['status'], undefined>
  }>({
    provider_key: '',
    name: '',
    driver: 'local',
    endpoint: '',
    region: '',
    base_url: '',
    access_key: '',
    secret_key: '',
    extra: {},
    is_default: false,
    status: 'ready'
  })

  const driverGuide = computed(() => getDriverGuide(form.driver, 'provider'))
  const driverSections = computed(() => getDriverExtraSections(form.driver, 'provider'))

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
    form.provider_key = ''
    form.name = ''
    form.driver = 'local'
    form.endpoint = ''
    form.region = ''
    form.base_url = ''
    form.access_key = ''
    form.secret_key = ''
    form.is_default = false
    form.status = 'ready'
    applyDriverExtra('local')
  }

  function hydrateFromRow(row: StorageProviderSummary) {
    form.provider_key = row.provider_key
    form.name = row.name
    form.driver = row.driver
    form.endpoint = row.endpoint || ''
    form.region = row.region || ''
    form.base_url = row.base_url || ''
    form.access_key = ''
    form.secret_key = ''
    form.is_default = !!row.is_default
    form.status = row.status === 'error' ? 'ready' : row.status
    applyDriverExtra(row.driver, row.extra)
  }

  function applyDriverExtra(driver: StorageDriver | '' | undefined, value?: unknown) {
    const draft = buildDriverExtraDraft(driver, 'provider', value)
    form.extra = draft.values
    objectExtraText.value = draft.objectText
    customExtraText.value = draft.customText
    activePanels.value = draft.activePanels
  }

  function onDriverChange(driver: string | string[]) {
    const next = (Array.isArray(driver) ? driver[0] : driver) as StorageDriver
    applyDriverExtra(next)
  }

  function restoreDriverDefaults() {
    applyDriverExtra(form.driver)
  }

  function onDrawerClosed() {
    submitting.value = false
  }

  function buildBody(): StorageProviderSaveRequest | null {
    const extra = buildDriverExtraBody(
      '存储服务',
      form.driver,
      'provider',
      form.extra,
      objectExtraText.value,
      customExtraText.value
    )
    if (extra === null) return null
    const body: StorageProviderSaveRequest = {
      provider_key: form.provider_key.trim(),
      name: form.name.trim(),
      driver: form.driver,
      is_default: form.is_default,
      status: form.status
    }
    if (form.base_url.trim()) body.base_url = form.base_url.trim()
    if (form.driver === 'aliyun_oss') {
      if (form.endpoint.trim()) body.endpoint = form.endpoint.trim()
      if (form.region.trim()) body.region = form.region.trim()
      if (form.access_key.trim()) body.access_key = form.access_key
      if (form.secret_key.trim()) body.secret_key = form.secret_key
    }
    if (extra) body.extra = extra
    return body
  }

  async function onSubmit() {
    const body = buildBody()
    if (!body) return
    if (!body.provider_key || !body.name) {
      ElMessage.warning('服务标识和名称必填')
      return
    }
    submitting.value = true
    try {
      if (editingId.value) {
        await fetchUpdateStorageProvider(editingId.value, body)
        ElMessage.success('存储服务已更新')
      } else {
        await fetchCreateStorageProvider(body)
        ElMessage.success('存储服务已创建')
      }
      emit('saved')
      visible.value = false
    } catch (err: any) {
      ElMessage.error(err?.message || '保存存储服务失败')
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
