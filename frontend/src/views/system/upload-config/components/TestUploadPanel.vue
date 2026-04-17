<template>
  <div class="test-upload-panel">
    <ElAlert
      type="info"
      :closable="false"
      show-icon
      title="测试上传面板"
      description="用于在配置中心内联验证当前 UploadKey/Rule 的前端直传链路：选择可见 UploadKey 与 Rule，挑选文件，点击开始即可走一遍 prepare → upload/直传 → complete 的完整流程。文件真实落盘在所选 UploadKey 绑定的 Bucket 上，落点由后端的 base_path + key_prefix 模板决定，无需前端再指定。"
    />

    <ElForm class="test-upload-form" label-width="120px" :inline="false">
      <ElFormItem label="UploadKey">
        <div class="row-flex">
          <ElSelect
            v-model="selectedKey"
            placeholder="请选择上传配置"
            style="width: 320px"
            :loading="keysLoading"
            filterable
            @change="onKeyChange"
          >
            <ElOption
              v-for="k in visibleKeys"
              :key="k.key"
              :label="`${k.name}（${k.key}）`"
              :value="k.key"
            />
          </ElSelect>
          <ElButton :loading="keysLoading" @click="reloadKeys">刷新可见 Key</ElButton>
        </div>
        <div v-if="currentKey" class="row-hint">
          模式：{{ currentKey.uploadMode }} · 可见性：{{ currentKey.visibility }} · 最大
          {{ formatSize(currentKey.maxSizeBytes) }} · 直传阈值
          {{ formatSize(currentKey.directSizeThresholdBytes) }}
        </div>
      </ElFormItem>

      <ElFormItem label="Rule">
        <ElSelect
          v-model="selectedRule"
          placeholder="请选择上传规则"
          :disabled="!currentKey"
          style="width: 320px"
          filterable
        >
          <ElOption
            v-for="r in currentRules"
            :key="r.ruleKey"
            :label="`${r.name}（${r.ruleKey}${r.isDefault ? ' · 默认' : ''}）`"
            :value="r.ruleKey"
          />
        </ElSelect>
        <div v-if="currentRule" class="row-hint">
          模式：{{ currentRule.uploadMode }} · 可见性：{{ currentRule.visibility }} · 最大
          {{ formatSize(currentRule.maxSizeBytes) }} · 允许类型：{{
            formatAcceptList(currentRule.allowedMimeTypes)
          }}
        </div>
      </ElFormItem>

      <ElFormItem label="文件">
        <div class="file-drop-zone" :class="{ 'is-drag': isDragging }"
             @dragover.prevent="onDragOver"
             @dragleave.prevent="onDragLeave"
             @drop.prevent="onDrop"
             @click="triggerFileInput">
          <input
            ref="fileInputRef"
            type="file"
            class="file-hidden-input"
            :accept="acceptAttr"
            @change="onFilePicked"
          />
          <div v-if="selectedFile" class="file-info">
            <div class="file-name">{{ selectedFile.name }}</div>
            <div class="file-meta">
              {{ formatSize(selectedFile.size) }} · {{ selectedFile.type || '未知 MIME' }}
            </div>
          </div>
          <div v-else class="file-placeholder">
            点击或拖拽文件到此处；当前 accept：{{ acceptAttr || '*' }}
          </div>
        </div>
        <div v-if="sizeWarning" class="row-warning">{{ sizeWarning }}</div>
      </ElFormItem>

      <ElFormItem>
        <ElButton
          type="primary"
          :loading="uploading"
          :disabled="!canSubmit"
          @click="runUpload"
        >
          开始测试上传
        </ElButton>
        <ElButton :disabled="uploading" @click="clearAll">清空结果</ElButton>
      </ElFormItem>
    </ElForm>

    <ElAlert
      v-if="errorText"
      class="test-upload-error"
      type="error"
      :closable="false"
      show-icon
      :title="`本次上传失败：${errorText}`"
    />

    <ElRow :gutter="16" class="test-upload-results">
      <ElCol :xs="24" :md="12">
        <ElCard shadow="never" class="result-card">
          <template #header>
            <div class="result-header">协商计划（Plan）</div>
          </template>
          <pre v-if="currentPlan" class="result-pre">{{ formatJson(currentPlan) }}</pre>
          <div v-else class="result-empty">暂无，发起上传后显示 prepare 协商结果。</div>
        </ElCard>
      </ElCol>
      <ElCol :xs="24" :md="12">
        <ElCard shadow="never" class="result-card">
          <template #header>
            <div class="result-header">最终结果（Result）</div>
          </template>
          <pre v-if="currentResult" class="result-pre">{{ formatJson(currentResult) }}</pre>
          <div v-else class="result-empty">暂无，上传成功后显示 storageKey / URL / mime / size / etag。</div>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElCard v-if="lastSuccess" shadow="never" class="last-success-card">
      <template #header>
        <div class="result-header">上一次成功记录</div>
      </template>
      <pre class="result-pre">{{ formatJson(lastSuccess) }}</pre>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useUpload } from '@/domains/upload/use-upload'
  import type {
    MediaUploadExecutionPlan,
    MediaUploadExecutionResult,
    MediaVisibleUploadKey,
    MediaVisibleUploadRule
  } from '@/domains/upload/api'

  const {
    uploading,
    visibleKeysLoading: keysLoading,
    visibleKeys,
    lastPlan,
    lastResult,
    error,
    fetchVisibleKeys,
    submitDetailed
  } = useUpload()

  const selectedKey = ref<string>('')
  const selectedRule = ref<string>('')
  const selectedFile = ref<File | null>(null)
  const fileInputRef = ref<HTMLInputElement | null>(null)
  const isDragging = ref(false)
  const errorText = ref('')
  const currentPlan = ref<MediaUploadExecutionPlan | null>(null)
  const currentResult = ref<MediaUploadExecutionResult | null>(null)
  const lastSuccess = ref<MediaUploadExecutionResult | null>(null)

  const currentKey = computed<MediaVisibleUploadKey | undefined>(() =>
    visibleKeys.value.find((k) => k.key === selectedKey.value)
  )
  const currentRules = computed<MediaVisibleUploadRule[]>(() => currentKey.value?.rules ?? [])
  const currentRule = computed<MediaVisibleUploadRule | undefined>(() =>
    currentRules.value.find((r) => r.ruleKey === selectedRule.value)
  )

  const acceptAttr = computed(() => {
    const sources: string[] = []
    if (currentRule.value) {
      sources.push(...(currentRule.value.allowedMimeTypes ?? []))
      sources.push(...(currentRule.value.clientAccept ?? []))
    } else if (currentKey.value) {
      sources.push(...(currentKey.value.clientAccept ?? []))
    }
    const uniq = Array.from(new Set(sources.filter(Boolean)))
    return uniq.join(',')
  })

  const sizeLimitBytes = computed(() =>
    currentRule.value?.maxSizeBytes ?? currentKey.value?.maxSizeBytes ?? 0
  )

  const sizeWarning = computed(() => {
    if (!selectedFile.value || !sizeLimitBytes.value) return ''
    return selectedFile.value.size > sizeLimitBytes.value
      ? `当前文件 ${formatSize(selectedFile.value.size)} 超出上限 ${formatSize(sizeLimitBytes.value)}`
      : ''
  })

  const canSubmit = computed(
    () => !!selectedFile.value && !!selectedKey.value && !sizeWarning.value
  )

  onMounted(() => {
    void reloadKeys()
  })

  async function reloadKeys() {
    try {
      await fetchVisibleKeys()
      if (!selectedKey.value && visibleKeys.value.length) {
        selectedKey.value = visibleKeys.value[0].key
        applyDefaultRule()
      }
    } catch (err) {
      ElMessage.error(err instanceof Error ? err.message : '加载可见上传配置失败')
    }
  }

  function onKeyChange() {
    applyDefaultRule()
  }

  function applyDefaultRule() {
    const key = currentKey.value
    if (!key) {
      selectedRule.value = ''
      return
    }
    const def = key.rules?.find((r) => r.ruleKey === key.defaultRuleKey)
    selectedRule.value = def?.ruleKey ?? key.rules?.[0]?.ruleKey ?? ''
  }

  function triggerFileInput() {
    fileInputRef.value?.click()
  }

  function onFilePicked(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0] ?? null
    selectedFile.value = file
  }

  function onDragOver() {
    isDragging.value = true
  }
  function onDragLeave() {
    isDragging.value = false
  }
  function onDrop(event: DragEvent) {
    isDragging.value = false
    const file = event.dataTransfer?.files?.[0] ?? null
    if (file) selectedFile.value = file
  }

  async function runUpload() {
    if (!selectedFile.value || !selectedKey.value) return
    errorText.value = ''
    currentPlan.value = null
    currentResult.value = null
    try {
      const result = await submitDetailed(selectedFile.value, {
        key: selectedKey.value,
        rule: selectedRule.value || undefined
      })
      currentPlan.value = lastPlan.value ?? result.plan
      currentResult.value = result
      lastSuccess.value = result
      ElMessage.success('上传成功')
    } catch (err) {
      errorText.value = err instanceof Error ? err.message : error.value || '上传失败'
      currentPlan.value = lastPlan.value
      currentResult.value = lastResult.value
    }
  }

  function clearAll() {
    selectedFile.value = null
    if (fileInputRef.value) fileInputRef.value.value = ''
    currentPlan.value = null
    currentResult.value = null
    errorText.value = ''
  }

  function formatSize(bytes?: number): string {
    if (!bytes || bytes <= 0) return '-'
    const units = ['B', 'KB', 'MB', 'GB']
    let v = bytes
    let i = 0
    while (v >= 1024 && i < units.length - 1) {
      v /= 1024
      i += 1
    }
    return `${v.toFixed(v >= 100 || i === 0 ? 0 : 1)} ${units[i]}`
  }

  function formatAcceptList(list: string[] | undefined): string {
    if (!list || !list.length) return '全部'
    return list.slice(0, 6).join(', ') + (list.length > 6 ? ` …（共 ${list.length} 项）` : '')
  }

  function formatJson(value: unknown): string {
    try {
      return JSON.stringify(value, null, 2)
    } catch {
      return String(value)
    }
  }
</script>

<style lang="scss" scoped>
  .test-upload-panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .test-upload-form {
    margin-top: 8px;
  }
  .row-flex {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .row-hint {
    margin-top: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
  .row-warning {
    margin-top: 4px;
    font-size: 12px;
    color: var(--el-color-warning);
  }
  .file-drop-zone {
    position: relative;
    width: 100%;
    max-width: 520px;
    min-height: 96px;
    border: 1px dashed var(--el-border-color);
    border-radius: 6px;
    padding: 16px;
    cursor: pointer;
    transition: border-color 0.2s;
    background-color: var(--el-fill-color-lighter);
    &:hover,
    &.is-drag {
      border-color: var(--el-color-primary);
    }
  }
  .file-hidden-input {
    display: none;
  }
  .file-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .file-name {
    font-weight: 600;
    word-break: break-all;
  }
  .file-meta {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
  .file-placeholder {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    text-align: center;
    line-height: 1.6;
  }
  .test-upload-error {
    margin-top: 4px;
  }
  .test-upload-results {
    margin-top: 4px;
  }
  .result-card {
    height: 100%;
  }
  .result-header {
    font-weight: 600;
  }
  .result-pre {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
    font-size: 12px;
    line-height: 1.5;
    max-height: 320px;
    overflow: auto;
  }
  .result-empty {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
  .last-success-card {
    margin-top: 4px;
  }
</style>
