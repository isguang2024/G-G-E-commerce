<template>
  <div class="p-4 login-page-template-page art-full-height">
    <ElCard class="art-table-card login-page-template-main" shadow="never">
      <ArtTableHeader layout="refresh,fullscreen" :loading="loading" @refresh="load">
        <template #left>
          <div class="login-template-header">
            <div class="login-template-title">认证页模板管理</div>
            <div class="login-template-tip">统一管理登录/注册/找回密码三页模板。</div>
          </div>
        </template>
        <template #right>
          <ElButton type="primary" @click="openCreate">新建模板</ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="pagedList"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDrawer
      v-model="dialogVisible"
      :title="editing ? '编辑模板' : '新建模板'"
      size="100%"
      direction="rtl"
      class="template-editor-drawer"
    >
      <div class="drawer-shell">
        <div class="dialog-layout">
          <ElForm ref="formRef" :model="form" :rules="formRules" label-width="120px" class="dialog-form">
            <ElCollapse v-model="editorPanels" class="template-editor-collapse">
              <ElCollapseItem name="basic" title="基础信息">
                <div class="panel-content">
                  <ElFormItem
                    label="模板 Key"
                    prop="template_key"
                    :error="fieldErrors.template_key"
                    :data-testid="'login-template-field-error'"
                    :data-field="'template_key'"
                    required
                  >
                    <ElInput
                      v-model="form.template_key"
                      :disabled="!!editing"
                      placeholder="如 default / aurora"
                    />
                  </ElFormItem>
                  <ElFormItem
                    label="名称"
                    prop="name"
                    :error="fieldErrors.name"
                    :data-testid="'login-template-field-error'"
                    :data-field="'name'"
                    required
                  >
                    <ElInput v-model="form.name" placeholder="给运营可读的模板名称" />
                  </ElFormItem>
                  <ElFormItem label="场景">
                    <ElSelect v-model="form.scene">
                      <ElOption label="auth_family" value="auth_family" />
                    </ElSelect>
                  </ElFormItem>
                  <ElFormItem label="作用域">
                    <ElSelect v-model="form.app_scope">
                      <ElOption label="shared (共享)" value="shared" />
                      <ElOption label="app (APP 专属)" value="app" />
                    </ElSelect>
                  </ElFormItem>
                  <ElFormItem label="状态">
                    <ElSelect v-model="form.status">
                      <ElOption label="normal" value="normal" />
                      <ElOption label="disabled" value="disabled" />
                    </ElSelect>
                  </ElFormItem>
                  <ElFormItem label="默认模板">
                    <ElSwitch v-model="form.is_default" />
                    <span class="field-tip ml-2">同场景下仅一个默认模板</span>
                  </ElFormItem>
                </div>
              </ElCollapseItem>

              <ElCollapseItem name="theme" title="全局 theme">
                <div class="panel-content">
                  <div class="field-tip panel-tip">
                    右侧实时预览支持：品牌色、圆角
                  </div>
                  <ElFormItem label="品牌色">
                    <ElInput v-model="configTheme.primaryColor" placeholder="#409EFF" />
                  </ElFormItem>
                  <ElFormItem label="圆角">
                    <ElInput v-model="configTheme.borderRadius" placeholder="8px" />
                  </ElFormItem>
                </div>
              </ElCollapseItem>

              <ElCollapseItem name="features" title="全局 features">
                <div class="panel-content">
                  <div class="field-tip panel-tip">
                    右侧实时预览支持：社交登录、记住密码、忘记密码、注册入口、社交入口配置
                  </div>
                  <ElFormItem label="社交登录">
                    <ElSwitch v-model="configFeatures.socialLogin" />
                  </ElFormItem>
                  <ElFormItem label="记住密码">
                    <ElSwitch v-model="configFeatures.rememberMe" />
                  </ElFormItem>
                  <ElFormItem label="忘记密码">
                    <ElSwitch v-model="configFeatures.forgetPassword" />
                  </ElFormItem>
                  <ElFormItem label="注册入口">
                    <ElSwitch v-model="configFeatures.register" />
                  </ElFormItem>
                  <ElFormItem label="社交入口配置">
                    <div class="social-config-wrap">
                      <div class="social-config-header">
                        <ElButton size="small" @click="addSocialItem">新增入口</ElButton>
                      </div>
                      <div v-if="socialItems.length === 0" class="field-tip">未配置社交入口</div>
                      <div
                        v-for="(item, idx) in socialItems"
                        :key="`social-${idx}`"
                        class="social-config-row"
                      >
                        <ElInput v-model="item.key" placeholder="key（如 wechat）" />
                        <ElInput v-model="item.name" placeholder="显示名（如 微信）" />
                        <ElInput v-model="item.icon" placeholder="图标（emoji / URL）" />
                        <ElSelect
                          v-model="item.preset"
                          placeholder="预设"
                          @change="(val) => applySocialPreset(item, val)"
                        >
                          <ElOption label="无" value="" />
                          <ElOption label="微信" value="wechat" />
                          <ElOption label="GitHub" value="github" />
                          <ElOption label="Google" value="google" />
                        </ElSelect>
                        <ElInput v-model="item.url" placeholder="跳转 URL（/auth/oauth/wechat）" />
                        <ElButton
                          link
                          type="primary"
                          :disabled="!isValidSocialUrl(item.url)"
                          @click="previewSocialUrl(item.url)"
                        >
                          预览
                        </ElButton>
                        <ElButton link type="danger" @click="removeSocialItem(idx)">删除</ElButton>
                      </div>
                      <ElInput
                        v-model="socialCustomHtml"
                        type="textarea"
                        :rows="3"
                        placeholder="可选：自定义 HTML（仅 HTML，脚本会被过滤）"
                      />
                    </div>
                  </ElFormItem>
                </div>
              </ElCollapseItem>

              <ElCollapseItem name="pages" title="pages">
                <div class="panel-content">
                  <div class="field-tip panel-tip">
                    右侧实时预览支持：标题、副标题、按钮文案、品牌色、圆角
                  </div>
                  <ElTabs v-model="sceneEditorTab" class="scene-tabs">
                    <ElTabPane
                      v-for="scene in SCENES"
                      :key="scene"
                      :label="sceneLabel(scene)"
                      :name="scene"
                    >
                      <ElFormItem :label="`${sceneLabel(scene)}标题`">
                        <ElInput
                          v-model="pageOverrides[scene].title"
                          placeholder="可选，留空则使用页面默认文案"
                        />
                      </ElFormItem>
                      <ElFormItem :label="`${sceneLabel(scene)}副标题`">
                        <ElInput
                          v-model="pageOverrides[scene].subTitle"
                          placeholder="可选，留空则使用页面默认文案"
                        />
                      </ElFormItem>
                      <ElFormItem :label="`${sceneLabel(scene)}主按钮文案`">
                        <ElInput
                          v-model="pageOverrides[scene].buttonText"
                          placeholder="可选，留空则使用页面默认文案"
                        />
                      </ElFormItem>
                      <ElFormItem
                        v-if="scene === 'forget_password'"
                        :label="`${sceneLabel(scene)}次按钮文案`"
                      >
                        <ElInput
                          v-model="pageOverrides[scene].secondaryButtonText"
                          placeholder="可选，留空则使用页面默认文案"
                        />
                      </ElFormItem>
                      <ElFormItem :label="`${sceneLabel(scene)}品牌色`">
                        <ElInput
                          v-model="pageOverrides[scene].primaryColor"
                          placeholder="可选，留空则继承全局"
                        />
                      </ElFormItem>
                      <ElFormItem :label="`${sceneLabel(scene)}圆角`">
                        <ElInput
                          v-model="pageOverrides[scene].borderRadius"
                          placeholder="可选，留空则继承全局"
                        />
                      </ElFormItem>
                    </ElTabPane>
                  </ElTabs>
                </div>
              </ElCollapseItem>

              <ElCollapseItem name="advanced" title="高级">
                <div class="panel-content">
                  <ElFormItem label="元数据">
                    <ElInput v-model="rawMetaText" type="textarea" :rows="3" />
                  </ElFormItem>
                </div>
              </ElCollapseItem>
            </ElCollapse>
          </ElForm>

          <div class="dialog-preview">
            <div class="preview-header">
              <span class="preview-title">实时预览</span>
              <div class="preview-header-actions">
                <div class="preview-slider-wrap">
                  <span class="preview-slider-label">视口</span>
                  <ElSlider
                    v-model="previewViewportHeight"
                    class="preview-size-slider"
                    :min="720"
                    :max="1440"
                    :step="40"
                    size="small"
                  />
                </div>
                <ElSelect v-model="previewScene" size="small" style="width: 130px">
                  <ElOption label="登录页" value="login" />
                  <ElOption label="注册页" value="register" />
                  <ElOption label="找回密码" value="forget_password" />
                </ElSelect>
                <ElButton size="small" @click="refreshPreview">刷新</ElButton>
              </div>
            </div>
            <ElScrollbar class="preview-viewport-scrollbar" always>
              <div class="preview-iframe-wrap">
                <iframe
                  v-if="previewUrl"
                  :key="previewKey"
                  :src="previewUrl"
                  class="preview-iframe"
                  :style="previewFrameStyle"
                  sandbox="allow-same-origin allow-scripts"
                />
                <div v-else class="preview-placeholder">
                  <span>保存模板后可预览</span>
                </div>
              </div>
            </ElScrollbar>
          </div>
        </div>
        <div class="drawer-footer">
          <div class="drawer-footer-tip">抽屉已改为全屏，适合长配置和实时预览并排编辑。</div>
          <div class="drawer-footer-actions">
            <ElButton @click="dialogVisible = false">取消</ElButton>
            <ElButton type="primary" @click="submit">保存</ElButton>
          </div>
        </div>
      </div>
    </ElDrawer>

    <ElDialog v-model="previewDialogVisible" title="模板预览" width="520px" top="5vh">
      <div class="standalone-preview-toolbar">
        <ElSelect v-model="standalonePreviewScene" size="small" style="width: 140px">
          <ElOption label="登录页" value="login" />
          <ElOption label="注册页" value="register" />
          <ElOption label="找回密码" value="forget_password" />
        </ElSelect>
      </div>
      <div class="standalone-preview-wrap">
        <iframe
          v-if="standalonePreviewUrl"
          :src="standalonePreviewUrl"
          class="standalone-preview-iframe"
          sandbox="allow-same-origin allow-scripts"
        />
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import { HttpError } from '@/utils/http/error'
  import {
    fetchCreateLoginPageTemplate,
    fetchDeleteLoginPageTemplate,
    fetchListLoginPageTemplates,
    fetchUpdateLoginPageTemplate
  } from '@/domains/governance/api/register'

  defineOptions({ name: 'SystemLoginPageTemplate' })

  type SceneKey = 'login' | 'register' | 'forget_password'
  interface SocialConfigItem {
    key: string
    name: string
    icon: string
    url: string
    preset?: string
  }
  interface SocialPreset {
    key: string
    name: string
    icon: string
  }

  const SCENES: SceneKey[] = ['login', 'register', 'forget_password']
  const SOCIAL_PRESETS: Record<string, SocialPreset> = {
    wechat: { key: 'wechat', name: '微信', icon: '💬' },
    github: { key: 'github', name: 'GitHub', icon: '🐙' },
    google: { key: 'google', name: 'Google', icon: 'G' }
  }

  const list = ref<any[]>([])
  const loading = ref(false)
  const dialogVisible = ref(false)
  const editing = ref<any>(null)
  const previewDialogVisible = ref(false)
  const standalonePreviewTemplateKey = ref('')
  const standalonePreviewScene = ref<SceneKey>('login')
  const sceneEditorTab = ref<SceneKey>('login')
  const previewScene = ref<SceneKey>('login')
  const previewKey = ref(0)
  const previewViewportWidth = ref(1440)
  const previewViewportHeight = ref(960)
  const editorPanels = ref<string[]>(['basic'])
  const previewDraftID = ref('')
  const pagination = reactive({
    current: 1,
    size: 10,
    total: 0
  })
  let previewDraftSyncTimer: number | null = null

  const form = reactive<any>({
    template_key: '',
    name: '',
    scene: 'auth_family',
    app_scope: 'shared',
    status: 'normal',
    is_default: false
  })

  // fieldErrors: 后端 Error.details.<field> 回显容器；规范见 docs/guides/frontend-observability-spec.md §2.4
  const formRef = ref<FormInstance>()
  const fieldErrors = reactive<Record<string, string>>({})
  const formRules: FormRules = {
    template_key: [
      { required: true, message: '请输入模板 Key', trigger: 'blur' },
      { pattern: /^[a-z0-9][a-z0-9._-]*$/, message: '仅允许小写字母数字和 . _ -', trigger: 'blur' }
    ],
    name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }]
  }
  function clearFieldErrors() {
    for (const k of Object.keys(fieldErrors)) delete fieldErrors[k]
  }
  function applyBackendFieldErrors(e: unknown): boolean {
    if (!(e instanceof HttpError)) return false
    const data = (e.data || {}) as { details?: Record<string, string> }
    const details = data.details
    if (!details || typeof details !== 'object') return false
    let applied = false
    for (const [field, reason] of Object.entries(details)) {
      if (typeof reason === 'string') {
        fieldErrors[field] = reason
        applied = true
      }
    }
    return applied
  }

  const configTheme = reactive<any>({
    primaryColor: '',
    borderRadius: ''
  })

  const configFeatures = reactive<any>({
    socialLogin: false,
    rememberMe: true,
    forgetPassword: true,
    register: true
  })
  const socialItems = ref<SocialConfigItem[]>([])
  const socialCustomHtml = ref('')

  const pageOverrides = reactive<Record<SceneKey, any>>({
    login: {
      title: '',
      subTitle: '',
      buttonText: '',
      secondaryButtonText: '',
      primaryColor: '',
      borderRadius: ''
    },
    register: {
      title: '',
      subTitle: '',
      buttonText: '',
      secondaryButtonText: '',
      primaryColor: '',
      borderRadius: ''
    },
    forget_password: {
      title: '',
      subTitle: '',
      buttonText: '',
      secondaryButtonText: '',
      primaryColor: '',
      borderRadius: ''
    }
  })

  const rawMetaText = ref('{}')

  function createEmptySocialItem(): SocialConfigItem {
    return {
      key: '',
      name: '',
      icon: '',
      url: '',
      preset: ''
    }
  }

  function addSocialItem() {
    socialItems.value.push(createEmptySocialItem())
  }

  function removeSocialItem(index: number) {
    socialItems.value.splice(index, 1)
  }

  function isValidSocialUrl(url: string): boolean {
    const value = `${url || ''}`.trim()
    if (!value) return false
    return /^https?:\/\//i.test(value) || value.startsWith('/')
  }

  function previewSocialUrl(url: string) {
    const value = `${url || ''}`.trim()
    if (!isValidSocialUrl(value)) return
    const target = value.startsWith('/') ? `${window.location.origin}${value}` : value
    window.open(target, '_blank', 'noopener,noreferrer')
  }

  function applySocialPreset(item: SocialConfigItem, presetKey: string) {
    const preset = SOCIAL_PRESETS[presetKey]
    if (!preset) return
    item.key = item.key || preset.key
    item.name = item.name || preset.name
    item.icon = item.icon || preset.icon
    if (!item.url) {
      item.url = `/auth/oauth/${preset.key}`
    }
  }

  function sceneLabel(scene: SceneKey): string {
    if (scene === 'login') return '登录页'
    if (scene === 'register') return '注册页'
    return '找回密码'
  }

  function scenePath(scene: SceneKey): string {
    if (scene === 'login') return '/account/auth/login'
    if (scene === 'register') return '/account/auth/register'
    return '/account/auth/forget-password'
  }

  function createPreviewDraftId(): string {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
      return crypto.randomUUID()
    }
    return `auth-preview-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
  }

  function previewDraftStorageKey(id: string): string {
    return `auth-template-preview:${id}`
  }

  function buildPreviewDraftPayload(): Record<string, unknown> {
    return {
      template_key: `${form.template_key || ''}`.trim(),
      name: `${form.name || ''}`.trim(),
      scene: form.scene || 'auth_family',
      app_scope: form.app_scope || 'shared',
      status: form.status || 'normal',
      is_default: Boolean(form.is_default),
      config: buildConfigPayload()
    }
  }

  function syncPreviewDraftNow(): void {
    const id = `${previewDraftID.value || ''}`.trim()
    if (!dialogVisible.value || !id || typeof window === 'undefined') return
    try {
      window.localStorage.setItem(
        previewDraftStorageKey(id),
        JSON.stringify(buildPreviewDraftPayload())
      )
    } catch {
      /* ignore preview draft sync errors */
    }
  }

  function schedulePreviewDraftSync(): void {
    if (previewDraftSyncTimer !== null) {
      window.clearTimeout(previewDraftSyncTimer)
    }
    previewDraftSyncTimer = window.setTimeout(() => {
      syncPreviewDraftNow()
      previewDraftSyncTimer = null
    }, 120)
  }

  function clearPreviewDraft(): void {
    const id = `${previewDraftID.value || ''}`.trim()
    if (previewDraftSyncTimer !== null) {
      window.clearTimeout(previewDraftSyncTimer)
      previewDraftSyncTimer = null
    }
    if (!id || typeof window === 'undefined') return
    try {
      window.localStorage.removeItem(previewDraftStorageKey(id))
    } catch {
      /* ignore preview draft cleanup errors */
    }
  }

  const previewUrl = computed(() => {
    const draftId = `${previewDraftID.value || ''}`.trim()
    if (!draftId) return ''
    const key = `${form.template_key || ''}`.trim()
    const path = scenePath(previewScene.value)
    const search = new URLSearchParams({
      preview: '1',
      preview_draft_id: draftId
    })
    if (key) {
      search.set('login_page_key', key)
    }
    return `${path}?${search.toString()}`
  })

  const standalonePreviewUrl = computed(() => {
    const key = `${standalonePreviewTemplateKey.value || ''}`.trim()
    if (!key) return ''
    const path = scenePath(standalonePreviewScene.value)
    return `${path}?login_page_key=${encodeURIComponent(key)}&preview=1`
  })

  const previewFrameStyle = computed(() => ({
    width: `${previewViewportWidth.value}px`,
    height: `${previewViewportHeight.value}px`
  }))
  const pagedList = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return list.value.slice(start, start + pagination.size)
  })
  const columns = computed<ColumnOption[]>(() => [
    { type: 'index', label: '序号', width: 70 },
    {
      prop: 'template_key',
      label: '模板 Key',
      minWidth: 180,
      showOverflowTooltip: true,
      formatter: (row) =>
        h(
          'span',
          {
            'data-testid': 'login-template-row',
            'data-template-key': row.template_key,
            'data-is-default': row.is_default ? 'true' : 'false',
            'data-status': row.status
          },
          row.template_key
        )
    },
    { prop: 'name', label: '名称', minWidth: 180, showOverflowTooltip: true },
    { prop: 'scene', label: '场景', width: 120 },
    { prop: 'app_scope', label: '作用域', width: 120 },
    { prop: 'status', label: '状态', width: 100 },
    {
      prop: 'is_default',
      label: '默认模板',
      width: 110,
      formatter: (row) =>
        row.is_default
          ? h(ElTag, { type: 'success', effect: 'plain' }, () => '是')
          : h('span', '否')
    },
    {
      prop: 'config_overview',
      label: '配置概览',
      minWidth: 220,
      formatter: (row) => {
        const children = []
        if (hasConfigKey(row, 'theme')) children.push(h(ElTag, { size: 'small', effect: 'plain' }, () => 'theme'))
        if (hasConfigKey(row, 'features')) {
          children.push(h(ElTag, { size: 'small', effect: 'plain', type: 'success' }, () => 'features'))
        }
        if (hasConfigKey(row, 'pages')) {
          children.push(h(ElTag, { size: 'small', effect: 'plain', type: 'info' }, () => 'pages'))
        }
        if (hasConfigKey(row, 'social')) {
          children.push(h(ElTag, { size: 'small', effect: 'plain', type: 'danger' }, () => 'social'))
        }
        if (children.length === 0) {
          children.push(h('span', { class: 'text-gray-400 text-xs' }, '未配置'))
        }
        return h('div', { class: 'config-badges' }, children)
      }
    },
    {
      prop: 'actions',
      label: '操作',
      width: 180,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'table-actions' }, [
          h(
            ElButton,
            { link: true, type: 'primary', onClick: () => openEdit(row) },
            () => '编辑'
          ),
          h(
            ElButton,
            { link: true, type: 'info', onClick: () => openPreviewOnly(row) },
            () => '预览'
          ),
          h(
            ElButton,
            { link: true, type: 'danger', onClick: () => confirmRemove(row) },
            () => '删除'
          )
        ])
    }
  ])

  function hasConfigKey(row: any, key: string): boolean {
    const config = row?.config
    if (!config || typeof config !== 'object') return false
    const value = config[key]
    if (value === null || value === undefined) return false
    if (typeof value === 'object') return Object.keys(value).length > 0
    return true
  }

  function hasAnyConfig(row: any): boolean {
    return (
      hasConfigKey(row, 'theme') ||
      hasConfigKey(row, 'features') ||
      hasConfigKey(row, 'pages') ||
      hasConfigKey(row, 'social')
    )
  }

  function cleanObject(obj: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(obj)) {
      if (v !== '' && v !== undefined && v !== null) {
        result[k] = v
      }
    }
    return result
  }

  function buildSocialItems(): SocialConfigItem[] {
    return socialItems.value
      .map((item) => ({
        key: `${item.key || ''}`.trim(),
        name: `${item.name || ''}`.trim(),
        icon: `${item.icon || ''}`.trim(),
        url: `${item.url || ''}`.trim()
      }))
      .filter((item) => item.key && item.url)
  }

  function buildPageOverrides(): Record<string, unknown> {
    const pages: Record<string, unknown> = {}
    for (const scene of SCENES) {
      const sceneData = pageOverrides[scene]
      const theme = cleanObject({
        primaryColor: sceneData.primaryColor,
        borderRadius: sceneData.borderRadius
      })
      const texts = cleanObject({
        title: sceneData.title,
        subTitle: sceneData.subTitle,
        buttonText: sceneData.buttonText,
        secondaryButtonText: sceneData.secondaryButtonText
      })
      const block: Record<string, unknown> = {}
      if (Object.keys(theme).length > 0) block.theme = theme
      if (Object.keys(texts).length > 0) block.texts = texts
      if (Object.keys(block).length > 0) {
        pages[scene] = block
      }
    }
    return pages
  }

  function buildConfigPayload(): Record<string, unknown> {
    const theme = cleanObject({ ...configTheme })
    const features = cleanObject({ ...configFeatures })
    const pages = buildPageOverrides()
    const items = buildSocialItems()
    const customHtml = `${socialCustomHtml.value || ''}`.trim()
    const social = cleanObject({
      items: items.length > 0 ? items : undefined,
      customHtml: customHtml || undefined
    })
    const base: Record<string, unknown> = {}
    if (Object.keys(theme).length > 0) base.theme = theme
    if (Object.keys(features).length > 0) base.features = features
    if (Object.keys(pages).length > 0) base.pages = pages
    if (Object.keys(social).length > 0) base.social = social
    return base
  }

  function resetPageOverrides(config: any) {
    const pages = config?.pages || {}
    for (const scene of SCENES) {
      const sceneData = pages?.[scene] || (scene === 'forget_password' ? pages?.forgetPassword : {}) || {}
      const theme = sceneData?.theme || {}
      const texts = sceneData?.texts || {}
      Object.assign(pageOverrides[scene], {
        title: texts.title || '',
        subTitle: texts.subTitle || '',
        buttonText: texts.buttonText || '',
        secondaryButtonText: texts.secondaryButtonText || '',
        primaryColor: theme.primaryColor || '',
        borderRadius: theme.borderRadius || ''
      })
    }
  }

  function resetConfigForm(config: any) {
    const theme = config?.theme || {}
    Object.assign(configTheme, {
      primaryColor: theme.primaryColor || '',
      borderRadius: theme.borderRadius || ''
    })
    const features = config?.features || {}
    Object.assign(configFeatures, {
      socialLogin: Boolean(features.socialLogin),
      rememberMe: features.rememberMe !== false,
      forgetPassword: features.forgetPassword !== false,
      register: features.register !== false
    })
    resetPageOverrides(config)
    const social = config?.social || {}
    const parsedItems = Array.isArray(social?.items) ? social.items : features?.socialItems
    socialItems.value = (Array.isArray(parsedItems) ? parsedItems : [])
      .map((item: any) => ({
        key: `${item?.key || ''}`.trim(),
        name: `${item?.name || ''}`.trim(),
        icon: `${item?.icon || ''}`.trim(),
        url: `${item?.url || ''}`.trim(),
        preset: ''
      }))
      .filter((item: SocialConfigItem) => item.key || item.name || item.icon || item.url)
    socialCustomHtml.value = `${social?.customHtml || features?.socialCustomHtml || ''}`.trim()
  }

  const load = async () => {
    loading.value = true
    try {
      const data: any = await fetchListLoginPageTemplates()
      list.value = data?.records || []
      pagination.total = list.value.length
      syncCurrentPage()
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  const openCreate = () => {
    clearPreviewDraft()
    previewDraftID.value = createPreviewDraftId()
    editing.value = null
    Object.assign(form, {
      template_key: '',
      name: '',
      scene: 'auth_family',
      app_scope: 'shared',
      status: 'normal',
      is_default: false
    })
    resetConfigForm({})
    rawMetaText.value = '{}'
    sceneEditorTab.value = 'login'
    previewScene.value = 'login'
    dialogVisible.value = true
    syncPreviewDraftNow()
  }

  const openEdit = (row: any) => {
    clearPreviewDraft()
    previewDraftID.value = createPreviewDraftId()
    editing.value = row
    Object.assign(form, {
      template_key: row.template_key || '',
      name: row.name || '',
      scene: row.scene || 'auth_family',
      app_scope: row.app_scope || 'shared',
      status: row.status || 'normal',
      is_default: Boolean(row.is_default)
    })
    const config = row.config || {}
    resetConfigForm(config)
    rawMetaText.value = JSON.stringify(row.meta || {}, null, 2)
    sceneEditorTab.value = 'login'
    previewScene.value = 'login'
    dialogVisible.value = true
    syncPreviewDraftNow()
  }

  const openPreviewOnly = (row: any) => {
    const key = `${row.template_key || ''}`.trim()
    if (!key) return
    standalonePreviewTemplateKey.value = key
    standalonePreviewScene.value = 'login'
    previewDialogVisible.value = true
  }

  const refreshPreview = () => {
    syncPreviewDraftNow()
    previewKey.value++
  }

  const submit = async () => {
    clearFieldErrors()
    const valid = await formRef.value?.validate().catch(() => false)
    if (!valid) return
    const templateKey = `${form.template_key || ''}`.trim()
    const name = `${form.name || ''}`.trim()
    if (!templateKey) {
      fieldErrors.template_key = '请填写模板 Key'
      return
    }
    if (!name) {
      fieldErrors.name = '请填写模板名称'
      return
    }
    const invalidSocialUrls = socialItems.value
      .map((item) => `${item.url || ''}`.trim())
      .filter((url) => url && !isValidSocialUrl(url))
    if (invalidSocialUrls.length > 0) {
      ElMessage.warning('社交入口 URL 仅支持 /path 或 http(s):// 开头')
      return
    }
    const config = buildConfigPayload()
    let meta: Record<string, unknown> = {}
    try {
      meta = JSON.parse(`${rawMetaText.value || '{}'}`) || {}
    } catch {
      ElMessage.warning('元数据不是有效 JSON')
      return
    }
    const payload = {
      template_key: templateKey,
      name,
      scene: form.scene || 'auth_family',
      app_scope: form.app_scope || 'shared',
      status: form.status || 'normal',
      is_default: Boolean(form.is_default),
      config,
      meta
    }
    try {
      if (editing.value) {
        await fetchUpdateLoginPageTemplate(editing.value.template_key, payload)
      } else {
        await fetchCreateLoginPageTemplate(payload)
      }
      ElMessage.success('模板已保存')
      dialogVisible.value = false
      await load()
    } catch (e: any) {
      if (applyBackendFieldErrors(e)) return
      ElMessage.error(e?.message || '模板保存失败')
    }
  }

  const confirmRemove = async (row: any) => {
    try {
      await ElMessageBox.confirm(`确认删除模板“${row.name || row.template_key}”吗？`, '删除确认', {
        type: 'warning'
      })
      await remove(row)
    } catch {}
  }

  const remove = async (row: any) => {
    try {
      await fetchDeleteLoginPageTemplate(row.template_key)
      ElMessage.success('模板已删除')
      await load()
    } catch (e: any) {
      ElMessage.error(e?.message || '模板删除失败')
    }
  }

  onMounted(load)

  function syncCurrentPage() {
    const totalPages = Math.max(1, Math.ceil((pagination.total || 0) / pagination.size))
    if (pagination.current > totalPages) {
      pagination.current = totalPages
    }
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    pagination.current = 1
    syncCurrentPage()
  }

  function handleCurrentChange(current: number) {
    pagination.current = current
    syncCurrentPage()
  }

  const previewDraftFingerprint = computed(() =>
    JSON.stringify({
      template_key: `${form.template_key || ''}`.trim(),
      name: `${form.name || ''}`.trim(),
      scene: form.scene || 'auth_family',
      app_scope: form.app_scope || 'shared',
      status: form.status || 'normal',
      is_default: Boolean(form.is_default),
      theme: { ...configTheme },
      features: { ...configFeatures },
      pages: SCENES.map((scene) => ({ scene, ...pageOverrides[scene] })),
      socialItems: socialItems.value.map((item) => ({ ...item })),
      socialCustomHtml: socialCustomHtml.value
    })
  )

  watch(
    () => dialogVisible.value,
    (visible) => {
      if (visible) {
        syncPreviewDraftNow()
        return
      }
      clearPreviewDraft()
      previewDraftID.value = ''
    }
  )

  watch(previewDraftFingerprint, () => {
    if (!dialogVisible.value) return
    schedulePreviewDraftSync()
  })

  onBeforeUnmount(() => {
    clearPreviewDraft()
  })
</script>

<style scoped>
  :deep(.template-editor-drawer) {
    --el-drawer-padding-primary: 20px;
  }

  .login-page-template-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .login-page-template-main {
    flex: 1;
    min-height: 0;
  }

  .login-page-template-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  :deep(.template-editor-drawer .el-drawer__header) {
    margin-bottom: 0;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--el-border-color-light);
  }

  :deep(.template-editor-drawer .el-drawer__body) {
    padding: 0;
    overflow: hidden;
  }

  .login-template-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .login-template-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .login-template-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .config-badges {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .table-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .drawer-shell {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 73px);
    background: var(--el-bg-color);
  }

  .dialog-layout {
    display: grid;
    grid-template-columns: minmax(0, 1.25fr) minmax(320px, 0.75fr);
    gap: 20px;
    flex: 1;
    min-height: 0;
    align-items: stretch;
    overflow: hidden;
    padding: 20px;
  }

  .dialog-form {
    min-width: 0;
    overflow-y: auto;
    padding-right: 8px;
  }

  .template-editor-collapse {
    border: none;
    background: transparent;
  }

  :deep(.template-editor-collapse .el-collapse-item) {
    margin-bottom: 12px;
    overflow: hidden;
    border: 1px solid var(--el-border-color-light);
    border-radius: 12px;
    background: var(--el-fill-color-blank);
  }

  :deep(.template-editor-collapse .el-collapse-item__header) {
    padding: 0 16px;
    font-weight: 600;
    border-bottom: 1px solid transparent;
    background: var(--el-bg-color);
  }

  :deep(.template-editor-collapse .el-collapse-item__wrap) {
    border-bottom: none;
  }

  :deep(.template-editor-collapse .el-collapse-item__content) {
    padding-bottom: 0;
  }

  .panel-content {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
  }

  .scene-tabs {
    margin-bottom: 8px;
  }

  .dialog-preview {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    height: 100%;
    min-height: 0;
  }

  .preview-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .preview-header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .preview-title {
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .preview-slider-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 220px;
  }

  .preview-slider-label {
    flex: none;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .preview-size-slider {
    flex: 1;
  }

  .preview-iframe-wrap {
    display: inline-flex;
    align-items: flex-start;
    justify-content: flex-start;
    border: 1px solid var(--el-border-color-light);
    border-radius: 12px;
    overflow: hidden;
    flex: none;
    height: 100%;
    min-height: 0;
    background: #f5f7fa;
  }

  .preview-iframe {
    display: block;
    width: 100%;
    border: none;
  }

  .preview-viewport-scrollbar {
    flex: 1;
    min-height: 0;
  }

  :deep(.preview-viewport-scrollbar .el-scrollbar__wrap) {
    overflow: auto;
    scrollbar-gutter: stable both-edges;
  }

  :deep(.preview-viewport-scrollbar .el-scrollbar__view) {
    min-width: 100%;
    min-height: 100%;
  }

  .preview-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--el-text-color-secondary);
    font-size: 14px;
  }

  .standalone-preview-toolbar {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 10px;
  }

  .standalone-preview-wrap {
    height: 620px;
    border-radius: 12px;
    overflow: hidden;
    border: 1px solid var(--el-border-color-light);
  }

  .standalone-preview-iframe {
    width: 100%;
    height: 100%;
    border: none;
  }

  .field-tip {
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }

  .panel-tip {
    margin-bottom: 4px;
  }

  .drawer-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 16px 20px;
    border-top: 1px solid var(--el-border-color-light);
    background: rgb(255 255 255 / 96%);
    backdrop-filter: blur(10px);
  }

  .drawer-footer-tip {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .drawer-footer-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .social-config-wrap {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .social-config-header {
    display: flex;
    justify-content: flex-start;
  }

  .social-config-row {
    display: grid;
    grid-template-columns: 100px 110px minmax(110px, 1fr) 90px minmax(180px, 1fr) auto auto;
    gap: 8px;
    align-items: center;
  }

  @media (max-width: 1280px) {
    .dialog-layout {
      grid-template-columns: minmax(0, 1fr);
    }

    .dialog-preview {
      position: static;
      height: auto;
    }

    .preview-iframe-wrap {
      height: 100%;
      min-height: 0;
    }
  }

  @media (max-width: 768px) {
    .drawer-shell {
      height: calc(100vh - 61px);
    }

    .dialog-layout {
      padding: 16px;
      gap: 16px;
    }

    .social-config-row {
      grid-template-columns: minmax(0, 1fr);
    }

    .drawer-footer {
      flex-direction: column;
      align-items: stretch;
    }

    .drawer-footer-actions {
      justify-content: flex-end;
    }
  }
</style>
