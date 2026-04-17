<template>
  <div class="p-4 upload-config-page art-full-height">
    <ElCard class="art-table-card upload-config-main" shadow="never">
      <div class="upload-config-header">
        <div class="upload-config-title">上传配置中心</div>
        <div class="upload-config-tip">
          统一管理文件上传链路的四层配置：存储服务 &rarr; 存储桶 &rarr; 上传配置 &rarr; 上传规则。
          所有配置变更会自动失效缓存并广播到运行时上传链路。
        </div>
        <div class="upload-config-path-note">
          <span class="path-chip">对象最终路径</span>
          <span class="path-expr">
            存储桶 base_path <span class="path-arrow">/</span> 上传配置 path_template
            <span class="path-arrow">/</span> 规则 sub_path <span class="path-arrow">/</span>
            文件名策略
          </span>
        </div>
      </div>

      <ElTabs v-model="activeTab" class="upload-config-tabs" @tab-change="onTabChange">
        <!-- ═══ 存储服务 ═══ -->
        <ElTabPane label="存储服务" name="provider">
          <div class="tab-desc">
            存储服务是最底层的连接配置，对应一个对象存储实例（本地磁盘或云 OSS）。
          </div>
          <ElForm :inline="true" class="upload-config-filters">
            <ElFormItem>
              <ElButton type="primary" @click="loadProviders">刷新</ElButton>
              <ElButton type="success" @click="openProviderCreate">新增存储服务</ElButton>
            </ElFormItem>
          </ElForm>
          <ArtTable
            :loading="provider.loading"
            :data="provider.records"
            :columns="providerColumns"
          />
        </ElTabPane>

        <!-- ═══ 存储桶 ═══ -->
        <ElTabPane label="存储桶" name="bucket">
          <div class="tab-desc">
            存储桶隶属于某个存储服务，代表一个逻辑隔离的文件存放区域，可独立配置公网访问地址和基础路径。
          </div>
          <ElForm :inline="true" class="upload-config-filters">
            <ElFormItem label="所属存储服务">
              <ElSelect
                v-model="bucket.providerFilter"
                clearable
                filterable
                placeholder="全部"
                style="width: 240px"
                @change="loadBuckets"
              >
                <ElOption
                  v-for="p in provider.records"
                  :key="p.id"
                  :label="`${p.name}（${p.provider_key}）`"
                  :value="p.id"
                />
              </ElSelect>
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="loadBuckets">刷新</ElButton>
              <ElButton type="success" @click="openBucketCreate">新增存储桶</ElButton>
            </ElFormItem>
          </ElForm>
          <ArtTable :loading="bucket.loading" :data="bucket.records" :columns="bucketColumns" />
        </ElTabPane>

        <!-- ═══ 上传配置 ═══ -->
        <ElTabPane label="上传配置" name="upload-key">
          <div class="tab-desc">
            上传配置（UploadKey）对应一个业务上传场景，如头像、附件、编辑器图片等。
            除了大小、类型和路径模板外，还要明确上传方式、是否对前端可见、权限键和可扩展参数。
          </div>
          <ElForm :inline="true" class="upload-config-filters">
            <ElFormItem label="所属存储桶">
              <ElSelect
                v-model="uploadKey.bucketFilter"
                clearable
                filterable
                placeholder="全部"
                style="width: 240px"
                @change="loadUploadKeys"
              >
                <ElOption
                  v-for="b in bucket.records"
                  :key="b.id"
                  :label="`${b.name}（${b.bucket_key}）`"
                  :value="b.id"
                />
              </ElSelect>
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="loadUploadKeys">刷新</ElButton>
              <ElButton type="success" @click="openUploadKeyCreate">新增上传配置</ElButton>
            </ElFormItem>
          </ElForm>
          <ArtTable
            :loading="uploadKey.loading"
            :data="uploadKey.records"
            :columns="uploadKeyColumns"
          />
        </ElTabPane>

        <!-- ═══ 测试上传（页面自测） ═══ -->
        <ElTabPane label="测试上传（页面自测）" name="test-upload">
          <div class="tab-desc">
            在配置中心内联自测前端直传链路：选择可见上传配置与规则，挑选文件，按 prepare → 直传/中转
            → complete 走一遍完整流程。落点由后端 Bucket.base_path + UploadKey.path_template +
            Rule.sub_path 决定，无需也不允许在此指定。
          </div>
          <TestUploadPanel />
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <!-- 编辑抽屉：存储服务 -->
    <ProviderEditorDrawer
      v-model:open="providerEditorOpen"
      :editing-id="providerEditingId"
      :row="providerEditingRow"
      @saved="onProviderSaved"
    />

    <!-- 编辑抽屉：存储桶 -->
    <BucketEditorDrawer
      v-model:open="bucketEditorOpen"
      :editing-id="bucketEditingId"
      :row="bucketEditingRow"
      :providers="provider.records"
      @saved="onBucketSaved"
    />

    <!-- 编辑抽屉：上传配置 -->
    <UploadKeyEditorDrawer
      v-model:open="uploadKeyEditorOpen"
      :editing-id="uploadKeyEditingId"
      :row="uploadKeyEditingRow"
      :buckets="bucket.records"
      @saved="onUploadKeySaved"
    />

    <!-- 规则管理抽屉（内含规则编辑抽屉） -->
    <RuleListDrawer
      v-model:open="ruleDrawerOpen"
      :parent-upload-key-id="ruleParentId"
      :parent-label="ruleParentLabel"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import {
    ElButton,
    ElCard,
    ElForm,
    ElFormItem,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElPopconfirm,
    ElSelect,
    ElTabPane,
    ElTabs,
    ElTag
  } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import {
    fetchDeleteStorageBucket,
    fetchDeleteStorageProvider,
    fetchDeleteUploadKey,
    fetchListStorageBuckets,
    fetchListStorageProviders,
    fetchListUploadKeys,
    fetchTestStorageProvider,
    type StorageBucketSummary,
    type StorageProviderSummary,
    type UploadKeySummary
  } from '@/domains/upload-config/api'
  import { useDictionaries } from '@/hooks/business/useDictionary'
  import TestUploadPanel from './components/TestUploadPanel.vue'
  import ProviderEditorDrawer from './components/ProviderEditorDrawer.vue'
  import BucketEditorDrawer from './components/BucketEditorDrawer.vue'
  import UploadKeyEditorDrawer from './components/UploadKeyEditorDrawer.vue'
  import RuleListDrawer from './components/RuleListDrawer.vue'
  import {
    driverLabel,
    formatBytes,
    formatSchemaConfigured,
    statusLabel,
    statusType,
    uploadModeLabel,
    visibilityLabel,
    type DriverExtraValueMap
  } from './components/_shared'

  defineOptions({ name: 'SystemUploadConfig' })

  type TabKey = 'provider' | 'bucket' | 'upload-key' | 'test-upload'
  const activeTab = ref<TabKey>('provider')

  // ── 字典 map（表格列 formatter 使用） ─────────────────────────────────────

  const { dictMap } = useDictionaries([
    'storage_driver',
    'storage_status',
    'upload_mode',
    'upload_visibility'
  ])

  function pickLabel(code: string, value: string, fallback: string) {
    const list = dictMap.value[code] || []
    const hit = list.find((item) => item.value === value)
    return hit?.label || fallback
  }

  function normalizeObjectValue<T extends Record<string, unknown>>(
    value: unknown
  ): T | undefined {
    if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
    return { ...(value as T) }
  }

  function renderFeatureTags(labels: string[]) {
    if (!labels.length) return '-'
    return h(
      'div',
      { class: 'config-inline-tags' },
      labels.map((label) => h(ElTag, { type: 'info', effect: 'plain', size: 'small' }, () => label))
    )
  }

  function buildProviderFeatureLabels(row: StorageProviderSummary): string[] {
    const extra = normalizeObjectValue<DriverExtraValueMap>(row.extra)
    if (!extra) return []
    const labels: string[] = []
    if (extra.sts_role_arn) labels.push('STS')
    if (extra.use_cname === true) labels.push('CNAME')
    if (extra.use_path_style === true) labels.push('PathStyle')
    if (extra.disable_ssl === true) labels.push('HTTP')
    if (!labels.length && Object.keys(extra).length) labels.push('已配置')
    return labels
  }

  function buildBucketFeatureLabels(row: StorageBucketSummary): string[] {
    const extra = normalizeObjectValue<DriverExtraValueMap>(row.extra)
    if (!extra) return []
    const labels: string[] = []
    if (extra.success_action_status) labels.push(`状态${extra.success_action_status}`)
    if (extra.content_disposition) labels.push('下载头')
    if (extra.callback || extra.callback_var) labels.push('回调')
    if (!labels.length && Object.keys(extra).length) labels.push('已配置')
    return labels
  }

  // ── 存储服务 ─────────────────────────────────────────────────────────────

  const provider = reactive({
    loading: false,
    records: [] as StorageProviderSummary[]
  })
  const providerEditorOpen = ref(false)
  const providerEditingId = ref('')
  const providerEditingRow = ref<StorageProviderSummary | null>(null)

  async function loadProviders() {
    provider.loading = true
    try {
      const res = await fetchListStorageProviders()
      provider.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载存储服务列表失败')
    } finally {
      provider.loading = false
    }
  }

  function openProviderCreate() {
    providerEditingId.value = ''
    providerEditingRow.value = null
    providerEditorOpen.value = true
  }

  function openProviderEdit(row: StorageProviderSummary) {
    providerEditingId.value = row.id
    providerEditingRow.value = row
    providerEditorOpen.value = true
  }

  async function removeProvider(row: StorageProviderSummary) {
    try {
      await fetchDeleteStorageProvider(row.id)
      ElMessage.success('已删除')
      await loadProviders()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  async function testProvider(row: StorageProviderSummary) {
    try {
      const result = await fetchTestStorageProvider(row.id)
      const detail = `结果：${result.ok ? '正常' : '异常'}${result.message ? ` / ${result.message}` : ''}${
        typeof result.latency_ms === 'number' ? ` / 延迟 ${result.latency_ms}ms` : ''
      }`
      ElMessageBox.alert(detail, '健康检查结果', { type: result.ok ? 'success' : 'warning' })
    } catch (err: any) {
      ElMessage.error(err?.message || '健康检查失败')
    }
  }

  function onProviderSaved() {
    loadProviders()
  }

  // ── 存储桶 ──────────────────────────────────────────────────────────────

  const bucket = reactive({
    loading: false,
    providerFilter: '' as string,
    records: [] as StorageBucketSummary[]
  })
  const bucketEditorOpen = ref(false)
  const bucketEditingId = ref('')
  const bucketEditingRow = ref<StorageBucketSummary | null>(null)

  async function loadBuckets() {
    bucket.loading = true
    try {
      const res = await fetchListStorageBuckets(bucket.providerFilter || undefined)
      bucket.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载存储桶列表失败')
    } finally {
      bucket.loading = false
    }
  }

  function openBucketCreate() {
    bucketEditingId.value = ''
    bucketEditingRow.value = null
    bucketEditorOpen.value = true
  }

  function openBucketEdit(row: StorageBucketSummary) {
    bucketEditingId.value = row.id
    bucketEditingRow.value = row
    bucketEditorOpen.value = true
  }

  async function removeBucket(row: StorageBucketSummary) {
    try {
      await fetchDeleteStorageBucket(row.id)
      ElMessage.success('已删除')
      await loadBuckets()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  function onBucketSaved() {
    loadBuckets()
  }

  // ── 上传配置 ─────────────────────────────────────────────────────────────

  const uploadKey = reactive({
    loading: false,
    bucketFilter: '' as string,
    records: [] as UploadKeySummary[]
  })
  const uploadKeyEditorOpen = ref(false)
  const uploadKeyEditingId = ref('')
  const uploadKeyEditingRow = ref<UploadKeySummary | null>(null)

  async function loadUploadKeys() {
    uploadKey.loading = true
    try {
      const res = await fetchListUploadKeys(uploadKey.bucketFilter || undefined)
      uploadKey.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载上传配置列表失败')
    } finally {
      uploadKey.loading = false
    }
  }

  function openUploadKeyCreate() {
    uploadKeyEditingId.value = ''
    uploadKeyEditingRow.value = null
    uploadKeyEditorOpen.value = true
  }

  function openUploadKeyEdit(row: UploadKeySummary) {
    uploadKeyEditingId.value = row.id
    uploadKeyEditingRow.value = row
    uploadKeyEditorOpen.value = true
  }

  async function removeUploadKey(row: UploadKeySummary) {
    try {
      await fetchDeleteUploadKey(row.id)
      ElMessage.success('已删除')
      await loadUploadKeys()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  function onUploadKeySaved() {
    loadUploadKeys()
  }

  // ── 规则管理抽屉 ──────────────────────────────────────────────────────────

  const ruleDrawerOpen = ref(false)
  const ruleParentId = ref('')
  const ruleParentLabel = ref('')

  function openRuleDrawer(row: UploadKeySummary) {
    ruleParentId.value = row.id
    ruleParentLabel.value = `${row.name}（${row.key}）`
    ruleDrawerOpen.value = true
  }

  // ── 列定义 ────────────────────────────────────────────────────────────────

  const providerColumns = computed<ColumnOption[]>(() => [
    { prop: 'provider_key', label: '服务标识', minWidth: 160 },
    { prop: 'name', label: '名称', minWidth: 160 },
    {
      prop: 'driver',
      label: '驱动类型',
      width: 130,
      formatter: (row: StorageProviderSummary) =>
        pickLabel('storage_driver', row.driver, driverLabel[row.driver] || row.driver)
    },
    {
      prop: 'extra',
      label: '扩展能力',
      minWidth: 180,
      formatter: (row: StorageProviderSummary) => renderFeatureTags(buildProviderFeatureLabels(row))
    },
    { prop: 'endpoint', label: '接入点', minWidth: 200, showOverflowTooltip: true },
    { prop: 'access_key_masked', label: '访问密钥（AK 掩码）', width: 180 },
    {
      prop: 'is_default',
      label: '默认',
      width: 80,
      formatter: (row: StorageProviderSummary) =>
        row.is_default ? h(ElTag, { type: 'success', effect: 'plain' }, () => '默认') : '-'
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row: StorageProviderSummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain' },
          () =>
            pickLabel('storage_status', row.status, statusLabel[row.status] || row.status)
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 280,
      fixed: 'right',
      formatter: (row: StorageProviderSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openProviderEdit(row) },
            () => '编辑'
          ),
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => testProvider(row) },
            () => '健康检查'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该存储服务？', onConfirm: () => removeProvider(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  const bucketColumns = computed<ColumnOption[]>(() => [
    { prop: 'bucket_key', label: '存储桶标识', minWidth: 160 },
    { prop: 'name', label: '名称', minWidth: 160 },
    { prop: 'provider_key', label: '所属服务', width: 160 },
    { prop: 'bucket_name', label: '存储桶名称', minWidth: 160 },
    { prop: 'base_path', label: '基础路径', minWidth: 140 },
    {
      prop: 'extra',
      label: '扩展能力',
      minWidth: 180,
      formatter: (row: StorageBucketSummary) => renderFeatureTags(buildBucketFeatureLabels(row))
    },
    {
      prop: 'is_public',
      label: '公开',
      width: 80,
      formatter: (row: StorageBucketSummary) =>
        row.is_public
          ? h(ElTag, { type: 'success', effect: 'plain' }, () => '公开')
          : h(ElTag, { type: 'info', effect: 'plain' }, () => '私有')
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row: StorageBucketSummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain' },
          () =>
            pickLabel('storage_status', row.status, statusLabel[row.status] || row.status)
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 200,
      fixed: 'right',
      formatter: (row: StorageBucketSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openBucketEdit(row) },
            () => '编辑'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该存储桶？', onConfirm: () => removeBucket(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  const uploadKeyColumns = computed<ColumnOption[]>(() => [
    { prop: 'key', label: '上传标识', minWidth: 160 },
    { prop: 'name', label: '名称', minWidth: 160 },
    { prop: 'bucket_key', label: '所属存储桶', width: 160 },
    {
      prop: 'upload_mode',
      label: '上传方式',
      width: 120,
      formatter: (row: UploadKeySummary) =>
        pickLabel(
          'upload_mode',
          row.upload_mode || 'auto',
          uploadModeLabel[row.upload_mode || 'auto'] || row.upload_mode || 'auto'
        )
    },
    {
      prop: 'is_frontend_visible',
      label: '前端可见',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        row.is_frontend_visible
          ? h(ElTag, { type: 'success', effect: 'plain', size: 'small' }, () => '可见')
          : h(ElTag, { type: 'info', effect: 'plain', size: 'small' }, () => '隐藏')
    },
    {
      prop: 'permission_key',
      label: '权限键',
      minWidth: 160,
      showOverflowTooltip: true,
      formatter: (row: UploadKeySummary) => row.permission_key || '-'
    },
    {
      prop: 'visibility',
      label: '可见性',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        h(
          ElTag,
          { type: row.visibility === 'public' ? 'success' : 'info', effect: 'plain' },
          () =>
            pickLabel(
              'upload_visibility',
              row.visibility,
              visibilityLabel[row.visibility] || row.visibility
            )
        )
    },
    {
      prop: 'max_size_bytes',
      label: '文件上限',
      width: 140,
      formatter: (row: UploadKeySummary) => formatBytes(Number(row.max_size_bytes ?? 0))
    },
    {
      prop: 'allowed_mime_types',
      label: '允许类型',
      minWidth: 200,
      showOverflowTooltip: true,
      formatter: (row: UploadKeySummary) =>
        Array.isArray(row.allowed_mime_types) && row.allowed_mime_types.length
          ? row.allowed_mime_types.join(', ')
          : '不限'
    },
    {
      prop: 'extra_schema',
      label: '扩展参数',
      width: 100,
      formatter: (row: UploadKeySummary) => formatSchemaConfigured(row.extra_schema)
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain' },
          () =>
            pickLabel('storage_status', row.status, statusLabel[row.status] || row.status)
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 280,
      fixed: 'right',
      formatter: (row: UploadKeySummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openUploadKeyEdit(row) },
            () => '编辑'
          ),
          h(
            ElButton,
            { type: 'warning', link: true, onClick: () => openRuleDrawer(row) },
            () => '管理规则'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该上传配置？', onConfirm: () => removeUploadKey(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  function onTabChange(name: string | number) {
    if (name === 'bucket' && bucket.records.length === 0) {
      loadBuckets()
    } else if (name === 'upload-key' && uploadKey.records.length === 0) {
      loadUploadKeys()
    }
  }

  onMounted(() => {
    loadProviders()
  })
</script>

<style scoped lang="scss">
  .upload-config-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .upload-config-main {
    flex: 1;
    min-height: 0;
  }

  .upload-config-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .upload-config-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0 12px;
  }

  .upload-config-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .upload-config-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .upload-config-path-note {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    margin-top: 4px;
    background: var(--el-fill-color-lighter);
    border: 1px dashed var(--el-border-color-lighter);
    border-radius: 6px;
    font-size: 12px;
    color: var(--el-text-color-regular);
    line-height: 1.6;
    align-self: flex-start;
  }

  .path-chip {
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-weight: 600;
  }

  .path-arrow {
    color: var(--el-text-color-secondary);
  }

  .path-expr {
    font-family: var(--el-font-family-mono, ui-monospace, SFMono-Regular, monospace);
    font-size: 12px;
  }

  .upload-config-tabs {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .upload-config-tabs :deep(.el-tabs__content) {
    flex: 1;
    min-height: 0;
  }

  .upload-config-tabs :deep(.el-tab-pane) {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .upload-config-filters {
    margin-bottom: 4px;
  }

  .tab-desc {
    margin-bottom: 12px;
    padding: 8px 12px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-lighter);
    border-radius: 4px;
  }

  :deep(.config-row-actions) {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  :deep(.config-inline-tags) {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
  }
</style>
