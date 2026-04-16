<template>
  <div class="p-4 upload-config-page art-full-height">
    <ElCard class="art-table-card upload-config-main" shadow="never">
      <div class="upload-config-header">
        <div class="upload-config-title">上传配置中心</div>
        <div class="upload-config-tip">
          统一管理文件上传链路的四层配置：存储服务 &rarr; 存储桶 &rarr; 上传配置 &rarr;
          上传规则。所有配置变更会自动失效缓存并广播到运行时上传链路。
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
            上传配置（UploadKey）对应一个业务上传场景，如头像、附件、编辑器图片等，定义该场景的文件大小上限、允许类型和路径模板。
          </div>
          <ElForm :inline="true" class="upload-config-filters">
            <ElFormItem label="所属存储桶">
              <ElSelect
                v-model="uploadKey.bucketFilter"
                clearable
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
      </ElTabs>
    </ElCard>

    <!-- 存储服务编辑器 -->
    <ElDialog
      v-model="providerEditor.open"
      :title="providerEditor.editingId ? '编辑存储服务' : '新增存储服务'"
      width="640px"
      :close-on-click-modal="false"
      @closed="resetProviderEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="服务标识">
          <ElInput
            v-model="providerEditor.form.provider_key"
            :disabled="!!providerEditor.editingId"
            placeholder="如 local-default、oss-prod"
          />
          <span class="form-tip">唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="providerEditor.form.name" placeholder="便于识别的显示名称" />
        </ElFormItem>
        <ElFormItem label="驱动类型">
          <ElSelect v-model="providerEditor.form.driver" style="width: 220px">
            <ElOption label="本地存储" value="local" />
            <ElOption label="阿里云 OSS" value="aliyun_oss" />
          </ElSelect>
          <span class="form-tip">选择存储后端类型</span>
        </ElFormItem>
        <ElFormItem label="接入点地址">
          <ElInput v-model="providerEditor.form.endpoint" placeholder="如 oss-cn-hangzhou.aliyuncs.com" />
          <span class="form-tip">OSS 服务的接入域名，本地存储可留空</span>
        </ElFormItem>
        <ElFormItem label="地域">
          <ElInput v-model="providerEditor.form.region" placeholder="如 cn-hangzhou" />
        </ElFormItem>
        <ElFormItem label="基础访问地址">
          <ElInput v-model="providerEditor.form.base_url" placeholder="如 https://cdn.example.com" />
          <span class="form-tip">文件公网访问的根地址，通常为 CDN 域名</span>
        </ElFormItem>
        <ElFormItem label="访问密钥（AK）">
          <ElInput
            v-model="providerEditor.form.access_key"
            placeholder="留空表示保留原值"
            autocomplete="off"
          />
        </ElFormItem>
        <ElFormItem label="安全密钥（SK）">
          <ElInput
            v-model="providerEditor.form.secret_key"
            type="password"
            show-password
            placeholder="留空表示保留原值"
            autocomplete="new-password"
          />
        </ElFormItem>
        <ElFormItem label="设为默认">
          <ElSwitch v-model="providerEditor.form.is_default" />
          <span class="form-tip">开启后，未指定存储服务的场景将使用此服务</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="providerEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="providerEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="providerEditor.submitting" @click="submitProvider"
          >保存</ElButton
        >
      </template>
    </ElDialog>

    <!-- 存储桶编辑器 -->
    <ElDialog
      v-model="bucketEditor.open"
      :title="bucketEditor.editingId ? '编辑存储桶' : '新增存储桶'"
      width="640px"
      :close-on-click-modal="false"
      @closed="resetBucketEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="所属存储服务">
          <ElSelect
            v-model="bucketEditor.form.provider_id"
            :disabled="!!bucketEditor.editingId"
            style="width: 100%"
          >
            <ElOption
              v-for="p in provider.records"
              :key="p.id"
              :label="`${p.name}（${p.provider_key}）`"
              :value="p.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="存储桶标识">
          <ElInput
            v-model="bucketEditor.form.bucket_key"
            :disabled="!!bucketEditor.editingId"
            placeholder="如 default-bucket"
          />
          <span class="form-tip">唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="bucketEditor.form.name" placeholder="便于识别的显示名称" />
        </ElFormItem>
        <ElFormItem label="存储桶名称">
          <ElInput v-model="bucketEditor.form.bucket_name" placeholder="对象存储中实际的 Bucket 名称" />
          <span class="form-tip">对应云存储服务中的实际 Bucket</span>
        </ElFormItem>
        <ElFormItem label="基础路径">
          <ElInput v-model="bucketEditor.form.base_path" placeholder="可选，文件存储的前缀目录" />
          <span class="form-tip">所有文件都会存储在此目录下</span>
        </ElFormItem>
        <ElFormItem label="公网访问地址">
          <ElInput
            v-model="bucketEditor.form.public_base_url"
            placeholder="访问已上传文件用的公网根地址"
          />
          <span class="form-tip">留空则继承存储服务的基础访问地址</span>
        </ElFormItem>
        <ElFormItem label="公开访问">
          <ElSwitch v-model="bucketEditor.form.is_public" />
          <span class="form-tip">开启后文件可通过公网直接访问，否则需签名</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="bucketEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="bucketEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="bucketEditor.submitting" @click="submitBucket"
          >保存</ElButton
        >
      </template>
    </ElDialog>

    <!-- 上传配置编辑器 -->
    <ElDialog
      v-model="uploadKeyEditor.open"
      :title="uploadKeyEditor.editingId ? '编辑上传配置' : '新增上传配置'"
      width="680px"
      :close-on-click-modal="false"
      @closed="resetUploadKeyEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="所属存储桶">
          <ElSelect
            v-model="uploadKeyEditor.form.bucket_id"
            :disabled="!!uploadKeyEditor.editingId"
            style="width: 100%"
          >
            <ElOption
              v-for="b in bucket.records"
              :key="b.id"
              :label="`${b.name}（${b.bucket_key}）`"
              :value="b.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="上传标识">
          <ElInput
            v-model="uploadKeyEditor.form.key"
            :disabled="!!uploadKeyEditor.editingId"
            placeholder="如 avatar、attachment、public-asset"
          />
          <span class="form-tip">业务场景的唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="uploadKeyEditor.form.name" placeholder="便于识别的显示名称" />
        </ElFormItem>
        <ElFormItem label="路径模板">
          <ElInput
            v-model="uploadKeyEditor.form.path_template"
            placeholder="{tenant}/{key}/{date}/{uuid}{ext}"
          />
          <span class="form-tip">支持变量：{tenant}、{key}、{date}、{uuid}、{ext}</span>
        </ElFormItem>
        <ElFormItem label="默认规则标识">
          <ElInput v-model="uploadKeyEditor.form.default_rule_key" placeholder="可选，留空则使用标记为默认的规则" />
        </ElFormItem>
        <ElFormItem label="单文件上限">
          <ElInputNumber
            v-model="uploadKeyEditor.form.max_size_bytes"
            :min="0"
            controls-position="right"
            style="width: 240px"
          />
          <span class="form-tip">单位：字节。0 表示沿用存储桶或全局上限</span>
        </ElFormItem>
        <ElFormItem label="可见性">
          <ElSelect v-model="uploadKeyEditor.form.visibility" style="width: 220px">
            <ElOption label="公开" value="public" />
            <ElOption label="私有" value="private" />
          </ElSelect>
          <span class="form-tip">公开文件可直接通过 URL 访问</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="uploadKeyEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="允许文件类型">
          <ElInput
            :model-value="uploadKeyEditor.mimeText"
            placeholder="逗号分隔，如 image/*,video/mp4，留空表示不限"
            @update:model-value="onMimeInput"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="uploadKeyEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="uploadKeyEditor.submitting" @click="submitUploadKey"
          >保存</ElButton
        >
      </template>
    </ElDialog>

    <!-- 上传规则管理抽屉 -->
    <ElDrawer
      v-model="rule.drawerOpen"
      :title="`上传规则管理 — ${rule.parentUploadKeyLabel}`"
      size="860px"
      :close-on-click-modal="false"
    >
      <div class="tab-desc" style="margin-bottom: 8px">
        上传规则是上传配置的子级，同一个上传配置下可定义多条规则，按不同子路径、文件名策略、大小限制分别处理。
      </div>
      <div class="rule-drawer-toolbar">
        <ElButton type="primary" size="small" @click="loadRules">刷新</ElButton>
        <ElButton type="success" size="small" @click="openRuleCreate">新增规则</ElButton>
      </div>
      <ArtTable :loading="rule.loading" :data="rule.records" :columns="ruleColumns" />
    </ElDrawer>

    <!-- 上传规则编辑器 -->
    <ElDialog
      v-model="ruleEditor.open"
      :title="ruleEditor.editingId ? '编辑上传规则' : '新增上传规则'"
      width="620px"
      :close-on-click-modal="false"
      append-to-body
      @closed="resetRuleEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="规则标识">
          <ElInput
            v-model="ruleEditor.form.rule_key"
            :disabled="!!ruleEditor.editingId"
            placeholder="如 image、file、poster"
          />
          <span class="form-tip">唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="ruleEditor.form.name" placeholder="如 图片上传、附件上传" />
        </ElFormItem>
        <ElFormItem label="子路径">
          <ElInput v-model="ruleEditor.form.sub_path" placeholder="可选，追加到上传配置路径之后" />
          <span class="form-tip">文件将存储在上传配置路径 + 子路径下</span>
        </ElFormItem>
        <ElFormItem label="文件名策略">
          <ElSelect v-model="ruleEditor.form.filename_strategy" style="width: 220px">
            <ElOption label="随机生成（UUID）" value="uuid" />
            <ElOption label="保留原文件名" value="original" />
          </ElSelect>
          <span class="form-tip">UUID 可避免文件名冲突</span>
        </ElFormItem>
        <ElFormItem label="单文件上限">
          <ElInputNumber
            v-model="ruleEditor.form.max_size_bytes"
            :min="0"
            controls-position="right"
            style="width: 240px"
          />
          <span class="form-tip">单位：字节。0 表示沿用上传配置的上限</span>
        </ElFormItem>
        <ElFormItem label="允许文件类型">
          <ElInput
            :model-value="ruleEditor.ruleMimeText"
            placeholder="逗号分隔，如 image/*,video/mp4，留空表示不限"
            @update:model-value="onRuleMimeInput"
          />
        </ElFormItem>
        <ElFormItem label="设为默认规则">
          <ElSwitch v-model="ruleEditor.form.is_default" />
          <span class="form-tip">上传时未指定规则将自动使用默认规则</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="ruleEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="ruleEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="ruleEditor.submitting" @click="submitRule"
          >保存</ElButton
        >
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import {
    ElButton,
    ElCard,
    ElDialog,
    ElDrawer,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElPopconfirm,
    ElSelect,
    ElSwitch,
    ElTabPane,
    ElTabs,
    ElTag
  } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import {
    fetchCreateStorageBucket,
    fetchCreateStorageProvider,
    fetchCreateUploadKey,
    fetchCreateUploadKeyRule,
    fetchDeleteStorageBucket,
    fetchDeleteStorageProvider,
    fetchDeleteUploadKey,
    fetchDeleteUploadKeyRule,
    fetchListStorageBuckets,
    fetchListStorageProviders,
    fetchListUploadKeyRules,
    fetchListUploadKeys,
    fetchTestStorageProvider,
    fetchUpdateStorageBucket,
    fetchUpdateStorageProvider,
    fetchUpdateUploadKey,
    fetchUpdateUploadKeyRule,
    type StorageBucketSaveRequest,
    type StorageBucketSummary,
    type StorageProviderSaveRequest,
    type StorageProviderSummary,
    type UploadKeyRuleSaveRequest,
    type UploadKeyRuleSummary,
    type UploadKeySaveRequest,
    type UploadKeySummary
  } from '@/domains/upload-config/api'

  defineOptions({ name: 'SystemUploadConfig' })

  // ── 显示文本映射 ──────────────────────────────────────────────────────────

  const statusLabel: Record<string, string> = {
    ready: '启用',
    disabled: '停用',
    error: '异常'
  }
  const statusType: Record<string, 'success' | 'info' | 'danger'> = {
    ready: 'success',
    disabled: 'info',
    error: 'danger'
  }
  const visibilityLabel: Record<string, string> = {
    public: '公开',
    private: '私有'
  }
  const filenameStrategyLabel: Record<string, string> = {
    uuid: '随机（UUID）',
    original: '保留原名'
  }
  const driverLabel: Record<string, string> = {
    local: '本地存储',
    aliyun_oss: '阿里云 OSS'
  }

  type TabKey = 'provider' | 'bucket' | 'upload-key'
  const activeTab = ref<TabKey>('provider')

  // ── Provider state ────────────────────────────────────────────────────────

  const provider = reactive({
    loading: false,
    records: [] as StorageProviderSummary[]
  })
  const providerEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    form: {
      provider_key: '',
      name: '',
      driver: 'local' as StorageProviderSaveRequest['driver'],
      endpoint: '',
      region: '',
      base_url: '',
      access_key: '',
      secret_key: '',
      is_default: false,
      status: 'ready' as Exclude<StorageProviderSaveRequest['status'], undefined>
    }
  })

  function resetProviderEditor() {
    providerEditor.submitting = false
    providerEditor.editingId = ''
    providerEditor.form = {
      provider_key: '',
      name: '',
      driver: 'local',
      endpoint: '',
      region: '',
      base_url: '',
      access_key: '',
      secret_key: '',
      is_default: false,
      status: 'ready'
    }
  }

  function openProviderCreate() {
    resetProviderEditor()
    providerEditor.open = true
  }

  function openProviderEdit(row: StorageProviderSummary) {
    resetProviderEditor()
    providerEditor.editingId = row.id
    providerEditor.form.provider_key = row.provider_key
    providerEditor.form.name = row.name
    providerEditor.form.driver = row.driver
    providerEditor.form.endpoint = row.endpoint || ''
    providerEditor.form.region = row.region || ''
    providerEditor.form.base_url = row.base_url || ''
    providerEditor.form.access_key = ''
    providerEditor.form.secret_key = ''
    providerEditor.form.is_default = !!row.is_default
    providerEditor.form.status = row.status === 'error' ? 'ready' : row.status
    providerEditor.open = true
  }

  function buildProviderBody(): StorageProviderSaveRequest {
    const f = providerEditor.form
    const body: StorageProviderSaveRequest = {
      provider_key: f.provider_key.trim(),
      name: f.name.trim(),
      driver: f.driver,
      is_default: f.is_default,
      status: f.status
    }
    if (f.endpoint.trim()) body.endpoint = f.endpoint.trim()
    if (f.region.trim()) body.region = f.region.trim()
    if (f.base_url.trim()) body.base_url = f.base_url.trim()
    if (f.access_key.trim()) body.access_key = f.access_key
    if (f.secret_key.trim()) body.secret_key = f.secret_key
    return body
  }

  async function submitProvider() {
    const body = buildProviderBody()
    if (!body.provider_key || !body.name) {
      ElMessage.warning('服务标识和名称必填')
      return
    }
    providerEditor.submitting = true
    try {
      if (providerEditor.editingId) {
        await fetchUpdateStorageProvider(providerEditor.editingId, body)
        ElMessage.success('存储服务已更新')
      } else {
        await fetchCreateStorageProvider(body)
        ElMessage.success('存储服务已创建')
      }
      providerEditor.open = false
      await loadProviders()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存存储服务失败')
    } finally {
      providerEditor.submitting = false
    }
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

  // ── Bucket state ─────────────────────────────────────────────────────────

  const bucket = reactive({
    loading: false,
    providerFilter: '' as string,
    records: [] as StorageBucketSummary[]
  })
  const bucketEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    form: {
      provider_id: '',
      bucket_key: '',
      name: '',
      bucket_name: '',
      base_path: '',
      public_base_url: '',
      is_public: false,
      status: 'ready' as Exclude<StorageBucketSaveRequest['status'], undefined>
    }
  })

  function resetBucketEditor() {
    bucketEditor.submitting = false
    bucketEditor.editingId = ''
    bucketEditor.form = {
      provider_id: '',
      bucket_key: '',
      name: '',
      bucket_name: '',
      base_path: '',
      public_base_url: '',
      is_public: false,
      status: 'ready'
    }
  }

  function openBucketCreate() {
    resetBucketEditor()
    if (provider.records[0]) bucketEditor.form.provider_id = provider.records[0].id
    bucketEditor.open = true
  }

  function openBucketEdit(row: StorageBucketSummary) {
    resetBucketEditor()
    bucketEditor.editingId = row.id
    bucketEditor.form.provider_id = row.provider_id
    bucketEditor.form.bucket_key = row.bucket_key
    bucketEditor.form.name = row.name
    bucketEditor.form.bucket_name = row.bucket_name
    bucketEditor.form.base_path = row.base_path || ''
    bucketEditor.form.public_base_url = row.public_base_url || ''
    bucketEditor.form.is_public = !!row.is_public
    bucketEditor.form.status = row.status
    bucketEditor.open = true
  }

  function buildBucketBody(): StorageBucketSaveRequest {
    const f = bucketEditor.form
    const body: StorageBucketSaveRequest = {
      provider_id: f.provider_id,
      bucket_key: f.bucket_key.trim(),
      name: f.name.trim(),
      bucket_name: f.bucket_name.trim(),
      is_public: f.is_public,
      status: f.status
    }
    if (f.base_path.trim()) body.base_path = f.base_path.trim()
    if (f.public_base_url.trim()) body.public_base_url = f.public_base_url.trim()
    return body
  }

  async function submitBucket() {
    const body = buildBucketBody()
    if (!body.provider_id || !body.bucket_key || !body.name || !body.bucket_name) {
      ElMessage.warning('所属存储服务、存储桶标识、名称、存储桶名称均为必填')
      return
    }
    bucketEditor.submitting = true
    try {
      if (bucketEditor.editingId) {
        await fetchUpdateStorageBucket(bucketEditor.editingId, body)
        ElMessage.success('存储桶已更新')
      } else {
        await fetchCreateStorageBucket(body)
        ElMessage.success('存储桶已创建')
      }
      bucketEditor.open = false
      await loadBuckets()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存存储桶失败')
    } finally {
      bucketEditor.submitting = false
    }
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

  // ── UploadKey state ──────────────────────────────────────────────────────

  const uploadKey = reactive({
    loading: false,
    bucketFilter: '' as string,
    records: [] as UploadKeySummary[]
  })
  const uploadKeyEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    mimeText: '',
    form: {
      bucket_id: '',
      key: '',
      name: '',
      path_template: '',
      default_rule_key: '',
      max_size_bytes: 0,
      allowed_mime_types: [] as string[],
      visibility: 'private' as Exclude<UploadKeySaveRequest['visibility'], undefined>,
      status: 'ready' as Exclude<UploadKeySaveRequest['status'], undefined>
    }
  })

  function resetUploadKeyEditor() {
    uploadKeyEditor.submitting = false
    uploadKeyEditor.editingId = ''
    uploadKeyEditor.mimeText = ''
    uploadKeyEditor.form = {
      bucket_id: '',
      key: '',
      name: '',
      path_template: '',
      default_rule_key: '',
      max_size_bytes: 0,
      allowed_mime_types: [],
      visibility: 'private',
      status: 'ready'
    }
  }

  function openUploadKeyCreate() {
    resetUploadKeyEditor()
    if (bucket.records[0]) uploadKeyEditor.form.bucket_id = bucket.records[0].id
    uploadKeyEditor.open = true
  }

  function openUploadKeyEdit(row: UploadKeySummary) {
    resetUploadKeyEditor()
    uploadKeyEditor.editingId = row.id
    uploadKeyEditor.form.bucket_id = row.bucket_id
    uploadKeyEditor.form.key = row.key
    uploadKeyEditor.form.name = row.name
    uploadKeyEditor.form.path_template = row.path_template || ''
    uploadKeyEditor.form.default_rule_key = row.default_rule_key || ''
    uploadKeyEditor.form.max_size_bytes = Number(row.max_size_bytes ?? 0)
    uploadKeyEditor.form.allowed_mime_types = Array.isArray(row.allowed_mime_types)
      ? row.allowed_mime_types
      : []
    uploadKeyEditor.form.visibility = row.visibility
    uploadKeyEditor.form.status = row.status
    uploadKeyEditor.mimeText = uploadKeyEditor.form.allowed_mime_types.join(',')
    uploadKeyEditor.open = true
  }

  function onMimeInput(value: string) {
    uploadKeyEditor.mimeText = value
    uploadKeyEditor.form.allowed_mime_types = value
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
  }

  function buildUploadKeyBody(): UploadKeySaveRequest {
    const f = uploadKeyEditor.form
    const body: UploadKeySaveRequest = {
      bucket_id: f.bucket_id,
      key: f.key.trim(),
      name: f.name.trim(),
      visibility: f.visibility,
      status: f.status,
      allowed_mime_types: f.allowed_mime_types
    }
    if (f.path_template.trim()) body.path_template = f.path_template.trim()
    if (f.default_rule_key.trim()) body.default_rule_key = f.default_rule_key.trim()
    if (Number.isFinite(f.max_size_bytes) && f.max_size_bytes > 0) {
      body.max_size_bytes = Number(f.max_size_bytes)
    }
    return body
  }

  async function submitUploadKey() {
    const body = buildUploadKeyBody()
    if (!body.bucket_id || !body.key || !body.name) {
      ElMessage.warning('所属存储桶、上传标识、名称必填')
      return
    }
    uploadKeyEditor.submitting = true
    try {
      if (uploadKeyEditor.editingId) {
        await fetchUpdateUploadKey(uploadKeyEditor.editingId, body)
        ElMessage.success('上传配置已更新')
      } else {
        await fetchCreateUploadKey(body)
        ElMessage.success('上传配置已创建')
      }
      uploadKeyEditor.open = false
      await loadUploadKeys()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存上传配置失败')
    } finally {
      uploadKeyEditor.submitting = false
    }
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

  // ── Rule state ───────────────────────────────────────────────────────────

  const rule = reactive({
    drawerOpen: false,
    loading: false,
    parentUploadKeyId: '',
    parentUploadKeyLabel: '',
    records: [] as UploadKeyRuleSummary[]
  })
  const ruleEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    ruleMimeText: '',
    form: {
      rule_key: '',
      name: '',
      sub_path: '',
      filename_strategy: 'uuid' as Exclude<UploadKeyRuleSaveRequest['filename_strategy'], undefined>,
      max_size_bytes: 0,
      allowed_mime_types: [] as string[],
      process_pipeline: [] as string[],
      is_default: false,
      status: 'ready' as Exclude<UploadKeyRuleSaveRequest['status'], undefined>
    }
  })

  function resetRuleEditor() {
    ruleEditor.submitting = false
    ruleEditor.editingId = ''
    ruleEditor.ruleMimeText = ''
    ruleEditor.form = {
      rule_key: '',
      name: '',
      sub_path: '',
      filename_strategy: 'uuid',
      max_size_bytes: 0,
      allowed_mime_types: [],
      process_pipeline: [],
      is_default: false,
      status: 'ready'
    }
  }

  function openRuleDrawer(row: UploadKeySummary) {
    rule.parentUploadKeyId = row.id
    rule.parentUploadKeyLabel = `${row.name}（${row.key}）`
    rule.records = []
    rule.drawerOpen = true
    loadRules()
  }

  function openRuleCreate() {
    resetRuleEditor()
    ruleEditor.open = true
  }

  function openRuleEdit(row: UploadKeyRuleSummary) {
    resetRuleEditor()
    ruleEditor.editingId = row.id
    ruleEditor.form.rule_key = row.rule_key
    ruleEditor.form.name = row.name
    ruleEditor.form.sub_path = row.sub_path || ''
    ruleEditor.form.filename_strategy = row.filename_strategy
    ruleEditor.form.max_size_bytes = Number(row.max_size_bytes ?? 0)
    ruleEditor.form.allowed_mime_types = Array.isArray(row.allowed_mime_types)
      ? row.allowed_mime_types
      : []
    ruleEditor.form.process_pipeline = Array.isArray(row.process_pipeline)
      ? row.process_pipeline
      : []
    ruleEditor.form.is_default = !!row.is_default
    ruleEditor.form.status = row.status
    ruleEditor.ruleMimeText = ruleEditor.form.allowed_mime_types.join(',')
    ruleEditor.open = true
  }

  function onRuleMimeInput(value: string) {
    ruleEditor.ruleMimeText = value
    ruleEditor.form.allowed_mime_types = value
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
  }

  function buildRuleBody(): UploadKeyRuleSaveRequest {
    const f = ruleEditor.form
    const body: UploadKeyRuleSaveRequest = {
      rule_key: f.rule_key.trim(),
      name: f.name.trim(),
      filename_strategy: f.filename_strategy,
      is_default: f.is_default,
      status: f.status,
      allowed_mime_types: f.allowed_mime_types
    }
    if (f.sub_path.trim()) body.sub_path = f.sub_path.trim()
    if (Number.isFinite(f.max_size_bytes) && f.max_size_bytes > 0) {
      body.max_size_bytes = Number(f.max_size_bytes)
    }
    if (f.process_pipeline.length > 0) body.process_pipeline = f.process_pipeline
    return body
  }

  async function submitRule() {
    const body = buildRuleBody()
    if (!body.rule_key || !body.name) {
      ElMessage.warning('规则标识和名称必填')
      return
    }
    ruleEditor.submitting = true
    try {
      if (ruleEditor.editingId) {
        await fetchUpdateUploadKeyRule(ruleEditor.editingId, body)
        ElMessage.success('上传规则已更新')
      } else {
        await fetchCreateUploadKeyRule(rule.parentUploadKeyId, body)
        ElMessage.success('上传规则已创建')
      }
      ruleEditor.open = false
      await loadRules()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存规则失败')
    } finally {
      ruleEditor.submitting = false
    }
  }

  async function removeRule(row: UploadKeyRuleSummary) {
    try {
      await fetchDeleteUploadKeyRule(row.id)
      ElMessage.success('已删除')
      await loadRules()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除规则失败')
    }
  }

  async function loadRules() {
    if (!rule.parentUploadKeyId) return
    rule.loading = true
    try {
      const res = await fetchListUploadKeyRules(rule.parentUploadKeyId)
      rule.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载规则列表失败')
    } finally {
      rule.loading = false
    }
  }

  const ruleColumns = computed<ColumnOption[]>(() => [
    { prop: 'rule_key', label: '规则标识', minWidth: 130 },
    { prop: 'name', label: '名称', minWidth: 130 },
    { prop: 'sub_path', label: '子路径', minWidth: 100 },
    {
      prop: 'filename_strategy',
      label: '文件名策略',
      width: 130,
      formatter: (row: UploadKeyRuleSummary) =>
        filenameStrategyLabel[row.filename_strategy] || row.filename_strategy
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
      prop: 'is_default',
      label: '默认',
      width: 70,
      formatter: (row: UploadKeyRuleSummary) =>
        row.is_default ? h(ElTag, { type: 'success', effect: 'plain', size: 'small' }, () => '是') : '-'
    },
    {
      prop: 'status',
      label: '状态',
      width: 80,
      formatter: (row: UploadKeyRuleSummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain', size: 'small' },
          () => statusLabel[row.status] || row.status
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 160,
      fixed: 'right',
      formatter: (row: UploadKeyRuleSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(ElButton, { type: 'primary', link: true, onClick: () => openRuleEdit(row) }, () => '编辑'),
          h(
            ElPopconfirm,
            { title: '确认删除该规则？', onConfirm: () => removeRule(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  // ── 列定义 ────────────────────────────────────────────────────────────────

  const providerColumns = computed<ColumnOption[]>(() => [
    { prop: 'provider_key', label: '服务标识', minWidth: 160 },
    { prop: 'name', label: '名称', minWidth: 160 },
    {
      prop: 'driver',
      label: '驱动类型',
      width: 130,
      formatter: (row: StorageProviderSummary) => driverLabel[row.driver] || row.driver
    },
    { prop: 'endpoint', label: '接入点', minWidth: 200, showOverflowTooltip: true },
    { prop: 'access_key_masked', label: '访问密钥', width: 160 },
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
          () => statusLabel[row.status] || row.status
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 280,
      fixed: 'right',
      formatter: (row: StorageProviderSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(ElButton, { type: 'primary', link: true, onClick: () => openProviderEdit(row) }, () => '编辑'),
          h(ElButton, { type: 'primary', link: true, onClick: () => testProvider(row) }, () => '健康检查'),
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
          () => statusLabel[row.status] || row.status
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 200,
      fixed: 'right',
      formatter: (row: StorageBucketSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(ElButton, { type: 'primary', link: true, onClick: () => openBucketEdit(row) }, () => '编辑'),
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
      prop: 'visibility',
      label: '可见性',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        h(
          ElTag,
          { type: row.visibility === 'public' ? 'success' : 'info', effect: 'plain' },
          () => visibilityLabel[row.visibility] || row.visibility
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
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain' },
          () => statusLabel[row.status] || row.status
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

  function formatBytes(value: number): string {
    if (!value || value <= 0) return '不限'
    const units = ['B', 'KB', 'MB', 'GB']
    let size = value
    let unit = 0
    while (size >= 1024 && unit < units.length - 1) {
      size /= 1024
      unit += 1
    }
    return `${size.toFixed(unit === 0 ? 0 : 2)} ${units[unit]}`
  }

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

  .config-row-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .form-tip {
    margin-left: 12px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
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

  .rule-drawer-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }
</style>
