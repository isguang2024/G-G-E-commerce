<template>
  <div class="site-config-page art-full-height">
    <div class="site-config-stack">
      <AdminWorkspaceHero
        title="参数管理"
        description="统一管理平台参数与 APP 作用域参数：同一个 config_key 可以在不同作用域重复创建；全局也是独立作用域；不同 APP 可复用相同 key。参数集合可将一组 config_key 编组成批量拉取单元。保存后会广播失效运行时缓存。"
        :metrics="summaryMetrics"
      >
        <div class="site-config-hero-actions">
          <ElButton @click="refreshAll(true)" v-ripple>刷新</ElButton>
          <ElButton v-if="activeTab === 'configs'" type="primary" @click="openConfigCreate" v-ripple>
            新增参数项
          </ElButton>
          <ElButton v-else type="primary" @click="openSetCreate" v-ripple>新增集合</ElButton>
        </div>
      </AdminWorkspaceHero>

      <ElCard class="art-table-card site-config-main" shadow="never">
        <ElAlert
          class="site-config-notice"
          type="info"
          :closable="false"
          show-icon
          title="作用域规则"
          description="同一个 config_key 可以在不同作用域重复创建；全局本身也是独立作用域；不同 APP 可以复用同一个 key。"
        />
        <ElTabs v-model="activeTab" class="site-config-tabs">
          <!-- ═══ 参数项 ═══ -->
          <ElTabPane label="参数项" name="configs">
            <div class="site-config-toolbar">
              <div class="site-config-toolbar__group">
                <ElRadioGroup
                  v-model="configScopeMode"
                  size="default"
                  @change="onScopeModeChange"
                >
                  <ElRadioButton value="global">仅全局</ElRadioButton>
                  <ElRadioButton value="all">全部作用域</ElRadioButton>
                  <ElRadioButton value="app">指定应用</ElRadioButton>
                </ElRadioGroup>
                <AppKeySelect
                  v-if="configScopeMode === 'app'"
                  v-model="selectedAppKey"
                  class="site-config-toolbar__app-select"
                  placeholder="选择或输入作用域 APP key"
                  allow-create
                  @change="onAppKeyChange"
                />
              </div>
              <div class="site-config-toolbar__group">
                <ElInput
                  v-model="configKeyword"
                  class="site-config-toolbar__search"
                  placeholder="按 config_key / 展示名搜索"
                  clearable
                >
                  <template #prefix>
                    <ArtSvgIcon icon="ri:search-line" style="font-size: 14px" />
                  </template>
                </ElInput>
              </div>
            </div>
            <ArtTable
              :loading="store.configsLoading"
              :data="filteredConfigs"
              :columns="configColumns"
              empty-text="暂无符合条件的参数项"
            />
          </ElTabPane>

          <!-- ═══ 参数集合 ═══ -->
          <ElTabPane label="参数集合" name="sets">
            <div class="site-config-toolbar">
              <div class="site-config-toolbar__group">
                <ElInput
                  v-model="setKeyword"
                  class="site-config-toolbar__search"
                  placeholder="按集合编码 / 名称搜索"
                  clearable
                >
                  <template #prefix>
                    <ArtSvgIcon icon="ri:search-line" style="font-size: 14px" />
                  </template>
                </ElInput>
              </div>
            </div>
            <ArtTable
              :loading="store.setsLoading"
              :data="filteredSets"
              :columns="setColumns"
              empty-text="暂无参数集合"
            />
          </ElTabPane>
        </ElTabs>
      </ElCard>
    </div>

    <!-- 参数项编辑器 -->
    <ElDialog
      v-model="configEditor.open"
      :title="configEditor.editingId ? '编辑参数项' : '新增参数项'"
      width="720px"
      :close-on-click-modal="false"
      @closed="resetConfigEditor"
    >
      <ElForm label-width="110px">
        <ElFormItem label="作用域">
          <ElRadioGroup
            v-model="configEditor.scope"
            :disabled="!!configEditor.editingId"
          >
            <ElRadio value="global">全局</ElRadio>
            <ElRadio value="app">指定应用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem v-if="configEditor.scope === 'app'" label="scope_key">
          <AppKeySelect
            v-model="configEditor.form.scope_key"
            :disabled="!!configEditor.editingId"
            allow-create
            placeholder="如 admin / shop / mobile"
          />
        </ElFormItem>
        <ElFormItem label="config_key">
          <ElInput
            v-model="configEditor.form.config_key"
            :disabled="!!configEditor.editingId"
            placeholder="如 site.name、site.logo"
          />
        </ElFormItem>
        <ElFormItem label="值类型">
          <ElSelect v-model="configEditor.form.value_type" style="width: 200px">
            <ElOption label="字符串 string" value="string" />
            <ElOption label="数值 number" value="number" />
            <ElOption label="布尔 bool" value="bool" />
            <ElOption label="图片 image" value="image" />
            <ElOption label="SVG 文本 svg" value="svg" />
            <ElOption label="JSON" value="json" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem v-if="configEditor.scope === 'global' && showFallbackPolicy" label="回退策略">
          <div class="site-config-form-row">
            <ElSelect v-model="configEditor.form.fallback_policy" style="width: 200px">
              <ElOption label="可继承 inherit" value="inherit" />
              <ElOption label="严格 strict" value="strict" />
            </ElSelect>
            <div class="site-config-form-hint">
              严格参数不会从当前作用域回退到全局默认。
            </div>
          </div>
        </ElFormItem>
        <ElFormItem label="值">
          <ElInput
            v-if="configEditor.form.value_type === 'string'"
            v-model="configEditor.scalarValue"
            placeholder="字符串值"
          />
          <ElInputNumber
            v-else-if="configEditor.form.value_type === 'number'"
            v-model="configEditor.numberValue"
            :controls="true"
          />
          <ElSwitch
            v-else-if="configEditor.form.value_type === 'bool'"
            v-model="configEditor.boolValue"
          />
          <div v-else-if="configEditor.form.value_type === 'image'" class="image-editor">
            <ElInput
              v-model="configEditor.imageUrl"
              placeholder="图片 URL（支持 .svg / .png / .webp 等）"
              clearable
              class="image-editor__input"
            />
            <ElButton
              :loading="configEditor.imageUploading"
              @click="($refs.imageFileInput as HTMLInputElement).click()"
            >
              上传文件
            </ElButton>
            <input
              ref="imageFileInput"
              type="file"
              accept="image/*,image/svg+xml,.svg"
              style="display: none"
              @change="handleImageUpload"
            />
            <img
              v-if="configEditor.imageUrl"
              :src="configEditor.imageUrl"
              class="image-editor__preview"
              alt="预览"
            />
          </div>
          <div v-else-if="configEditor.form.value_type === 'svg'" class="svg-editor">
            <ElInput
              v-model="configEditor.svgText"
              type="textarea"
              :rows="8"
              placeholder="粘贴或输入 SVG 标记，如 <svg xmlns=&quot;...&quot;>...</svg>"
              class="svg-editor__textarea"
            />
            <div v-if="configEditor.svgText.trim()" class="svg-editor__preview">
              <div class="svg-editor__preview-label">预览</div>
              <img
                :src="`data:image/svg+xml;charset=utf-8,${encodeURIComponent(configEditor.svgText)}`"
                alt="SVG 预览"
                class="svg-editor__preview-img"
              />
            </div>
          </div>
          <ElInput
            v-else
            v-model="configEditor.jsonText"
            type="textarea"
            :rows="6"
            placeholder="合法 JSON"
          />
        </ElFormItem>
        <ElFormItem v-if="configEditor.scope === 'global'">
          <ElButton link type="primary" @click="showFallbackPolicy = !showFallbackPolicy">
            {{ showFallbackPolicy ? '隐藏回退策略' : '显示回退策略' }}
          </ElButton>
        </ElFormItem>
        <ElFormItem label="展示名">
          <ElInput v-model="configEditor.form.label" placeholder="用于管理页显示" />
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput v-model="configEditor.form.description" type="textarea" :rows="2" />
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber v-model="configEditor.form.sort_order" :min="0" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="configEditor.form.status" style="width: 160px">
            <ElOption label="启用" value="normal" />
            <ElOption label="停用" value="suspended" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="configEditor.open = false">取消</ElButton>
        <ElButton
          type="primary"
          :loading="configEditor.submitting"
          @click="submitConfig"
        >保存</ElButton>
      </template>
    </ElDialog>

    <!-- 参数集合编辑器 -->
    <ElDialog
      v-model="setEditor.open"
      :title="setEditor.editingId ? '编辑参数集合' : '新增参数集合'"
      width="560px"
      :close-on-click-modal="false"
      @closed="resetSetEditor"
    >
      <ElForm label-width="100px">
        <ElFormItem label="集合编码">
          <ElInput
            v-model="setEditor.form.set_code"
            :disabled="!!setEditor.editingId"
            placeholder="唯一标识，如 site.branding"
          />
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="setEditor.form.set_name" placeholder="展示名" />
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput v-model="setEditor.form.description" type="textarea" :rows="2" />
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber v-model="setEditor.form.sort_order" :min="0" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="setEditor.form.status" style="width: 160px">
            <ElOption label="启用" value="normal" />
            <ElOption label="停用" value="suspended" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="setEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="setEditor.submitting" @click="submitSet">保存</ElButton>
      </template>
    </ElDialog>

    <!-- 参数集合 items 编辑器 -->
    <ElDialog
      v-model="itemsEditor.open"
      title="编辑参数集合成员"
      width="640px"
      :close-on-click-modal="false"
      @closed="resetItemsEditor"
    >
      <div class="items-editor-tip">
        通过下拉选择已登记的 config_key，也可输入新 key 回车创建。保存后后端会覆盖该集合的成员列表。
      </div>
      <ElSelect
        v-model="itemsEditor.keys"
        class="items-editor-select"
        multiple
        filterable
        allow-create
        default-first-option
        placeholder="输入或选择 config_key"
      >
        <ElOption
          v-for="k in configKeyOptions"
          :key="k"
          :label="k"
          :value="k"
        />
      </ElSelect>
      <template #footer>
        <ElButton @click="itemsEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="itemsEditor.submitting" @click="submitItems">保存</ElButton>
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
    ElDialog,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElOption,
    ElPopconfirm,
    ElRadio,
    ElRadioButton,
    ElRadioGroup,
    ElSwitch,
    ElTabPane,
    ElTabs,
    ElTag
  } from 'element-plus'
  import AppKeySelect from '@/components/business/app/AppKeySelect.vue'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { describeMediaUploadPlan, uploadMediaWithPlan } from '@/domains/upload/api'
  import type { ColumnOption } from '@/types/component'
  import { useSiteConfigStore } from '@/store/modules/site-config'
  import {
    type SiteConfigManageScopeType,
    type SiteConfigFallbackPolicy,
    type SiteConfigSaveRequest,
    type SiteConfigSetSaveRequest,
    type SiteConfigSetSummary,
    type SiteConfigSummary,
    type SiteConfigValueType
  } from '@/domains/site-config/types'

  defineOptions({ name: 'SystemSiteConfig' })

  const store = useSiteConfigStore()
  const activeTab = ref<'configs' | 'sets'>('configs')

  // ── 作用域切换 ───────────────────────────────────────────────────────────
  type ScopeMode = 'global' | 'all' | 'app'
  const configScopeMode = ref<ScopeMode>('global')
  const selectedAppKey = ref<string>('')
  const configKeyword = ref('')
  const setKeyword = ref('')

  const configKeyOptions = computed(() => {
    const set = new Set<string>()
    for (const item of store.configs) {
      if (item.config_key) set.add(item.config_key)
    }
    return Array.from(set).sort()
  })

  function currentScopeQuery(): {
    scopeType: SiteConfigManageScopeType
    scopeKey?: string
  } {
    if (configScopeMode.value === 'all') return { scopeType: 'all' }
    if (configScopeMode.value === 'app') {
      return { scopeType: 'app', scopeKey: selectedAppKey.value.trim() }
    }
    return { scopeType: 'global' }
  }

  async function loadConfigs(force = false) {
    try {
      if (configScopeMode.value === 'app' && !selectedAppKey.value.trim()) {
        // 作用域指定 APP 但尚未选择 scopeKey，清空后台以免误导。
        return
      }
      await store.listConfigs(currentScopeQuery(), force)
    } catch (err: any) {
      ElMessage.error(err?.message || '加载参数项失败')
    }
  }

  async function loadSets(force = false) {
    try {
      await store.listSets(force)
    } catch (err: any) {
      ElMessage.error(err?.message || '加载参数集合失败')
    }
  }

  async function refreshAll(force = false) {
    await Promise.all([loadConfigs(force), loadSets(force)])
  }

  function onScopeModeChange() {
    if (configScopeMode.value !== 'app') selectedAppKey.value = ''
    loadConfigs(true)
  }

  function onAppKeyChange() {
    loadConfigs(true)
  }

  // ── 过滤 ─────────────────────────────────────────────────────────────────
  const filteredConfigs = computed<SiteConfigSummary[]>(() => {
    const kw = configKeyword.value.trim().toLowerCase()
    if (!kw) return store.configs
    return store.configs.filter((row) => {
      const key = (row.config_key || '').toLowerCase()
      const label = (row.label || '').toLowerCase()
      return key.includes(kw) || label.includes(kw)
    })
  })

  const filteredSets = computed<SiteConfigSetSummary[]>(() => {
    const kw = setKeyword.value.trim().toLowerCase()
    if (!kw) return store.sets
    return store.sets.filter((row) => {
      const code = (row.set_code || '').toLowerCase()
      const name = (row.set_name || '').toLowerCase()
      return code.includes(kw) || name.includes(kw)
    })
  })

  // ── Hero 指标 ─────────────────────────────────────────────────────────────
  const summaryMetrics = computed(() => {
    const configs = store.configs
    const globalCount = configs.filter((r) => r.scope_type === 'global').length
    const appCount = configs.filter((r) => r.scope_type === 'app').length
    return [
      { label: '参数项总数', value: configs.length },
      { label: '全局作用域', value: globalCount },
      { label: 'APP 作用域', value: appCount },
      { label: '参数集合', value: store.sets.length }
    ]
  })

  // ── Configs 编辑器 ───────────────────────────────────────────────────────

  interface ConfigForm {
    scope_key: string
    config_key: string
    value_type: SiteConfigValueType
    fallback_policy: SiteConfigFallbackPolicy
    label: string
    description: string
    sort_order: number
    status: 'normal' | 'suspended'
  }

  const configEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    scope: 'global' as 'global' | 'app',
    scalarValue: '',
    numberValue: 0,
    boolValue: false,
    imageUrl: '',
    imageUploading: false,
    jsonText: '{}',
    svgText: '',
    form: {
      scope_key: '',
      config_key: '',
      value_type: 'string' as SiteConfigValueType,
      fallback_policy: 'inherit' as SiteConfigFallbackPolicy,
      label: '',
      description: '',
      sort_order: 0,
      status: 'normal'
    } as ConfigForm
  })
  const showFallbackPolicy = ref(false)

  function resetConfigEditor() {
    configEditor.submitting = false
    configEditor.editingId = ''
    configEditor.scope = 'global'
    configEditor.scalarValue = ''
    configEditor.numberValue = 0
    configEditor.boolValue = false
    configEditor.imageUrl = ''
    configEditor.imageUploading = false
    configEditor.jsonText = '{}'
    configEditor.svgText = ''
    showFallbackPolicy.value = false
    configEditor.form = {
      scope_key: '',
      config_key: '',
      value_type: 'string',
      fallback_policy: 'inherit',
      label: '',
      description: '',
      sort_order: 0,
      status: 'normal'
    }
  }

  function openConfigCreate() {
    resetConfigEditor()
    configEditor.open = true
  }

  function openConfigEdit(row: SiteConfigSummary) {
    resetConfigEditor()
    configEditor.editingId = row.id
    configEditor.scope = row.scope_type === 'app' ? 'app' : 'global'
    configEditor.form.scope_key = row.scope_key || ''
    configEditor.form.config_key = row.config_key
    configEditor.form.value_type = row.value_type
    configEditor.form.fallback_policy = row.fallback_policy || 'inherit'
    showFallbackPolicy.value = false
    configEditor.form.label = row.label || ''
    configEditor.form.description = row.description || ''
    configEditor.form.sort_order = row.sort_order ?? 0
    configEditor.form.status =
      row.status === 'suspended' ? 'suspended' : 'normal'
    // 根据类型填充 value 编辑字段
    const raw = (row.config_value || {}) as Record<string, unknown>
    switch (row.value_type) {
      case 'string':
        configEditor.scalarValue = typeof raw.value === 'string' ? raw.value : ''
        break
      case 'number':
        configEditor.numberValue =
          typeof raw.value === 'number'
            ? raw.value
            : Number(raw.value ?? 0) || 0
        break
      case 'bool':
        configEditor.boolValue = raw.value === true || raw.value === 'true'
        break
      case 'image':
        configEditor.imageUrl = typeof raw.url === 'string' ? raw.url : ''
        break
      case 'json':
        configEditor.jsonText = JSON.stringify(raw, null, 2)
        break
      case 'svg':
        configEditor.svgText = typeof raw.value === 'string' ? raw.value : ''
        break
    }
    configEditor.open = true
  }

  // ── 图片上传 ─────────────────────────────────────────────────────────────
  async function handleImageUpload(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    // 重置 input 以支持重复选同一文件
    input.value = ''
    configEditor.imageUploading = true
    try {
      const result = await uploadMediaWithPlan(file, { key: 'default' })
      if (result.media.url) {
        configEditor.imageUrl = result.media.url
        ElMessage.success(`图片上传成功（${describeMediaUploadPlan(result.plan)}）`)
      } else {
        ElMessage.warning('上传成功但未返回 URL，请手动填写')
      }
    } catch (err: any) {
      ElMessage.error(err?.message || '图片上传失败')
    } finally {
      configEditor.imageUploading = false
    }
  }

  function buildConfigValue(): Record<string, unknown> | undefined {
    switch (configEditor.form.value_type) {
      case 'string':
        return { value: configEditor.scalarValue }
      case 'number':
        return { value: Number(configEditor.numberValue) || 0 }
      case 'bool':
        return { value: !!configEditor.boolValue }
      case 'image':
        return { url: configEditor.imageUrl }
      case 'svg':
        return { value: configEditor.svgText }
      case 'json': {
        try {
          const parsed = JSON.parse(configEditor.jsonText || '{}')
          if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
            return parsed as Record<string, unknown>
          }
          return { value: parsed }
        } catch {
          throw new Error('JSON 值格式不合法')
        }
      }
    }
  }

  async function submitConfig() {
    try {
      const key = configEditor.form.config_key.trim()
      if (!key) {
        ElMessage.warning('config_key 必填')
        return
      }
      if (configEditor.scope === 'app' && !configEditor.form.scope_key.trim()) {
        ElMessage.warning('选择"指定应用"时 scope_key 必填')
        return
      }
      const value = buildConfigValue()
      const body: SiteConfigSaveRequest = {
        scope_type: configEditor.scope,
        scope_key:
          configEditor.scope === 'app' ? configEditor.form.scope_key.trim() : undefined,
        config_key: key,
        config_value: value,
        value_type: configEditor.form.value_type,
        fallback_policy:
          configEditor.scope === 'app' ? 'inherit' : configEditor.form.fallback_policy,
        label: configEditor.form.label || undefined,
        description: configEditor.form.description || undefined,
        sort_order: configEditor.form.sort_order,
        status: configEditor.form.status
      }
      configEditor.submitting = true
      if (configEditor.editingId) {
        await store.updateConfig(configEditor.editingId, body)
        ElMessage.success('参数项已更新')
      } else {
        await store.upsertConfig(body)
        ElMessage.success('参数项已保存')
      }
      configEditor.open = false
    } catch (err: any) {
      ElMessage.error(err?.message || '保存参数项失败')
    } finally {
      configEditor.submitting = false
    }
  }

  async function removeConfig(row: SiteConfigSummary) {
    try {
      await store.deleteConfig(row.id)
      ElMessage.success('已删除')
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  // ── 富展示 ───────────────────────────────────────────────────────────────
  type TagType = 'info' | 'primary' | 'success' | 'warning' | 'danger'
  const TYPE_COLORS: Record<SiteConfigValueType, TagType> = {
    string: 'info',
    number: 'primary',
    bool: 'success',
    image: 'warning',
    json: 'danger',
    svg: 'warning'
  }

  function renderValueCell(row: SiteConfigSummary) {
    const raw = (row.config_value || {}) as Record<string, unknown>
    if (row.value_type === 'image') {
      const url = typeof raw.url === 'string' ? raw.url : ''
      if (!url) return h('span', { class: 'value-empty' }, '未设置')
      return h('div', { class: 'value-image' }, [
        h('img', { src: url, alt: row.config_key, class: 'value-image__thumb', loading: 'lazy' }),
        h('span', { class: 'value-image__url' }, url)
      ])
    }
    if (row.value_type === 'svg') {
      const svgText = typeof raw.value === 'string' ? raw.value : ''
      if (!svgText) return h('span', { class: 'value-empty' }, '未设置')
      const dataUri = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svgText)}`
      return h('div', { class: 'value-image' }, [
        h('img', { src: dataUri, alt: row.config_key, class: 'value-image__thumb' }),
        h('span', { class: 'value-image__url' }, `SVG (${svgText.length} 字节)`)
      ])
    }
    if (row.value_type === 'bool') {
      const on = raw.value === true || raw.value === 'true'
      return h(
        ElTag,
        { type: on ? 'success' : 'info', size: 'small', effect: on ? 'dark' : 'plain' },
        () => (on ? 'true' : 'false')
      )
    }
    if (row.value_type === 'string' || row.value_type === 'number') {
      if (raw.value === undefined || raw.value === null || raw.value === '') {
        return h('span', { class: 'value-empty' }, '未设置')
      }
      return h('span', { class: 'value-scalar' }, String(raw.value))
    }
    // json
    try {
      return h('code', { class: 'value-json' }, JSON.stringify(raw))
    } catch {
      return h('span', { class: 'value-empty' }, '-')
    }
  }

  const configColumns = computed<ColumnOption[]>(() => [
    {
      prop: 'config_key',
      label: 'config_key',
      minWidth: 200,
      formatter: (row: SiteConfigSummary) =>
        h('code', { class: 'site-config-key' }, row.config_key)
    },
    {
      prop: 'scope_key',
      label: '作用域',
      width: 140,
      formatter: (row: SiteConfigSummary) =>
        row.scope_type === 'app'
          ? h(ElTag, { type: 'warning', effect: 'plain', size: 'small' }, () => `app: ${row.scope_key}`)
          : h(ElTag, { type: 'info', effect: 'plain', size: 'small' }, () => '全局')
    },
    {
      prop: 'value_type',
      label: '类型',
      width: 100,
      formatter: (row: SiteConfigSummary) =>
        h(
          ElTag,
          {
            type: TYPE_COLORS[row.value_type] || 'info',
            effect: 'plain',
            size: 'small'
          },
          () => row.value_type
        )
    },
    {
      prop: 'config_value',
      label: '当前值',
      minWidth: 260,
      showOverflowTooltip: false,
      formatter: (row: SiteConfigSummary) => renderValueCell(row)
    },
    { prop: 'label', label: '展示名', minWidth: 140 },
    { prop: 'sort_order', label: '排序', width: 80 },
    {
      prop: 'status',
      label: '状态',
      width: 90,
      formatter: (row: SiteConfigSummary) =>
        h(
          ElTag,
          {
            type: row.status === 'normal' ? 'success' : 'info',
            effect: 'plain',
            size: 'small'
          },
          () => (row.status === 'normal' ? '启用' : '停用')
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 160,
      fixed: 'right',
      formatter: (row: SiteConfigSummary) =>
        h('div', { class: 'site-config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openConfigEdit(row) },
            () => '编辑'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该参数项？', onConfirm: () => removeConfig(row) },
            {
              reference: () =>
                h(ElButton, { type: 'danger', link: true }, () => '删除')
            }
          )
        ])
    }
  ])

  // ── Sets 编辑器 ──────────────────────────────────────────────────────────

  const setEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    form: {
      set_code: '',
      set_name: '',
      description: '',
      sort_order: 0,
      status: 'normal' as 'normal' | 'suspended'
    }
  })

  function resetSetEditor() {
    setEditor.submitting = false
    setEditor.editingId = ''
    setEditor.form = {
      set_code: '',
      set_name: '',
      description: '',
      sort_order: 0,
      status: 'normal'
    }
  }

  function openSetCreate() {
    resetSetEditor()
    setEditor.open = true
  }

  function openSetEdit(row: SiteConfigSetSummary) {
    resetSetEditor()
    setEditor.editingId = row.id
    setEditor.form.set_code = row.set_code
    setEditor.form.set_name = row.set_name
    setEditor.form.description = row.description || ''
    setEditor.form.sort_order = row.sort_order ?? 0
    setEditor.form.status = row.status === 'suspended' ? 'suspended' : 'normal'
    setEditor.open = true
  }

  async function submitSet() {
    try {
      const code = setEditor.form.set_code.trim()
      const name = setEditor.form.set_name.trim()
      if (!code || !name) {
        ElMessage.warning('集合编码与名称必填')
        return
      }
      const body: SiteConfigSetSaveRequest = {
        set_code: code,
        set_name: name,
        description: setEditor.form.description || undefined,
        sort_order: setEditor.form.sort_order,
        status: setEditor.form.status
      }
      setEditor.submitting = true
      if (setEditor.editingId) {
        await store.updateSet(setEditor.editingId, body)
        ElMessage.success('集合已更新')
      } else {
        await store.upsertSet(body)
        ElMessage.success('集合已保存')
      }
      setEditor.open = false
    } catch (err: any) {
      ElMessage.error(err?.message || '保存集合失败')
    } finally {
      setEditor.submitting = false
    }
  }

  async function removeSet(row: SiteConfigSetSummary) {
    try {
      await store.deleteSet(row.id)
      ElMessage.success('已删除')
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  // ── Sets items 编辑 ──────────────────────────────────────────────────────

  const itemsEditor = reactive({
    open: false,
    submitting: false,
    setId: '',
    keys: [] as string[]
  })

  function resetItemsEditor() {
    itemsEditor.submitting = false
    itemsEditor.setId = ''
    itemsEditor.keys = []
  }

  function openItemsEdit(row: SiteConfigSetSummary) {
    resetItemsEditor()
    itemsEditor.setId = row.id
    itemsEditor.keys = [...(row.config_keys || [])]
    itemsEditor.open = true
  }

  async function submitItems() {
    try {
      const seen = new Set<string>()
      const keys: string[] = []
      for (const raw of itemsEditor.keys) {
        const k = (raw || '').trim()
        if (!k || seen.has(k)) continue
        seen.add(k)
        keys.push(k)
      }
      itemsEditor.submitting = true
      await store.updateSetItems(itemsEditor.setId, { config_keys: keys })
      ElMessage.success('集合成员已更新')
      itemsEditor.open = false
    } catch (err: any) {
      ElMessage.error(err?.message || '更新集合成员失败')
    } finally {
      itemsEditor.submitting = false
    }
  }

  const setColumns = computed<ColumnOption[]>(() => [
    {
      prop: 'set_code',
      label: 'set_code',
      minWidth: 180,
      formatter: (row: SiteConfigSetSummary) =>
        h('code', { class: 'site-config-key' }, row.set_code)
    },
    { prop: 'set_name', label: '名称', minWidth: 160 },
    {
      prop: 'config_keys',
      label: '成员 key',
      minWidth: 260,
      formatter: (row: SiteConfigSetSummary) => {
        const keys = row.config_keys || []
        if (keys.length === 0) return h('span', { class: 'value-empty' }, '未配置成员')
        return h(
          'div',
          { class: 'site-config-set-keys' },
          keys
            .slice(0, 6)
            .map((k) => h(ElTag, { type: 'info', effect: 'plain', size: 'small' }, () => k))
            .concat(
              keys.length > 6
                ? [h('span', { class: 'more-tip' }, `+${keys.length - 6}`)]
                : []
            )
        )
      }
    },
    { prop: 'sort_order', label: '排序', width: 80 },
    {
      prop: 'status',
      label: '状态',
      width: 90,
      formatter: (row: SiteConfigSetSummary) =>
        h(
          ElTag,
          {
            type: row.status === 'normal' ? 'success' : 'info',
            effect: 'plain',
            size: 'small'
          },
          () => (row.status === 'normal' ? '启用' : '停用')
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 240,
      fixed: 'right',
      formatter: (row: SiteConfigSetSummary) =>
        h('div', { class: 'site-config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openSetEdit(row) },
            () => '编辑'
          ),
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openItemsEdit(row) },
            () => '成员'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该集合？', onConfirm: () => removeSet(row) },
            {
              reference: () =>
                h(ElButton, { type: 'danger', link: true }, () => '删除')
            }
          )
        ])
    }
  ])

  // ── 启动加载 ─────────────────────────────────────────────────────────────

  onMounted(async () => {
    await loadConfigs()
    await loadSets()
  })
</script>

<style scoped lang="scss">
  .site-config-page {
    display: flex;
    flex-direction: column;
    padding: 16px;
  }

  .site-config-stack {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 16px;
    min-height: 0;
  }

  .site-config-hero-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
  }

  .site-config-main {
    flex: 1;
    min-height: 0;
  }

  .site-config-notice {
    margin-bottom: 14px;
  }

  .site-config-form-row {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
  }

  .site-config-form-hint {
    font-size: 12px;
    color: var(--art-text-gray-500);
    line-height: 1.5;
  }

  // ── 工具栏 ───────────────────────────────────────────────────────────
  .site-config-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 10px;
  }

  .site-config-toolbar__group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .site-config-toolbar__app-select {
    width: 200px;
  }

  .site-config-toolbar__search {
    width: 260px;
  }

  // ── 表格值富展示 ─────────────────────────────────────────────────────
  .site-config-key {
    display: inline-block;
    padding: 2px 6px;
    border-radius: 4px;
    background: var(--art-bg-soft);
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
    color: var(--art-text-strong);
  }

  .value-empty {
    font-size: 12px;
    color: var(--art-text-soft);
    font-style: italic;
  }

  .value-image {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .value-image__thumb {
    width: 32px;
    height: 32px;
    border-radius: 4px;
    object-fit: contain;
    background: var(--art-bg-soft);
    border: 1px solid var(--art-border-color);
  }

  .value-image__url {
    font-size: 12px;
    color: var(--art-text-muted);
    max-width: 220px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .value-scalar {
    font-size: 13px;
    color: var(--art-text-strong);
  }

  .value-json {
    display: inline-block;
    padding: 2px 6px;
    border-radius: 4px;
    background: var(--art-bg-soft);
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 11px;
    color: var(--art-text-muted);
    max-width: 320px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
  }

  // ── 表格行操作 ───────────────────────────────────────────────────────
  .site-config-row-actions :deep(.el-button + .el-button),
  .site-config-row-actions :deep(.el-button + .el-popconfirm),
  .site-config-row-actions :deep(.el-popconfirm + .el-button) {
    margin-left: 8px;
  }

  .site-config-set-keys {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    .more-tip {
      font-size: 12px;
      color: var(--art-text-gray-500);
      align-self: center;
    }
  }

  .items-editor-tip {
    font-size: 12px;
    color: var(--art-text-gray-500);
    margin-bottom: 10px;
  }

  .items-editor-select {
    width: 100%;
  }

  // ── 编辑器：image 上传 ────────────────────────────────────────────────
  .image-editor {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;

    &__input {
      width: 100%;
    }

    &__preview {
      max-width: 160px;
      max-height: 80px;
      border-radius: 6px;
      border: 1px solid var(--art-border-color);
      object-fit: contain;
      background: var(--art-bg-soft);
      padding: 4px;
    }
  }

  // ── 编辑器：SVG 文本 ──────────────────────────────────────────────────
  .svg-editor {
    display: flex;
    flex-direction: column;
    gap: 10px;
    width: 100%;

    &__textarea {
      width: 100%;
      font-family: var(--el-font-family-monospace, monospace);
      font-size: 12px;
    }

    &__preview {
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    &__preview-label {
      font-size: 12px;
      color: var(--art-text-muted);
    }

    &__preview-img {
      max-width: 200px;
      max-height: 100px;
      border-radius: 6px;
      border: 1px solid var(--art-border-color);
      background:
        linear-gradient(45deg, rgba(0, 0, 0, 0.04) 25%, transparent 25%) 0 0 / 8px 8px,
        linear-gradient(-45deg, rgba(0, 0, 0, 0.04) 25%, transparent 25%) 0 0 / 8px 8px;
      object-fit: contain;
      padding: 4px;
    }
  }
</style>
