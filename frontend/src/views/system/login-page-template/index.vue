<template>
  <div class="p-4 login-page-template-page">
    <div class="page-hero">
      <div>
        <h3 class="text-lg font-semibold">认证页模板管理</h3>
        <p class="hero-desc">
          统一管理登录/注册/找回密码三页模板。支持全局配置（theme/features/texts）以及按页面覆盖（pages.login/register/forget_password）。
        </p>
      </div>
      <div class="hero-actions">
        <ElButton type="primary" @click="openCreate">新建模板</ElButton>
      </div>
    </div>

    <ElAlert
      class="mb-4"
      type="info"
      :closable="false"
      title="配置说明"
      description="先配置全局 theme/features/texts，再按页面做覆盖。运行时会按 page_scene 自动合并全局与 pages.* 配置。"
    />

    <ElTable :data="list" border stripe>
      <ElTableColumn prop="template_key" label="模板 Key" width="180" />
      <ElTableColumn prop="name" label="名称" width="180" />
      <ElTableColumn prop="scene" label="场景" width="120" />
      <ElTableColumn prop="app_scope" label="作用域" width="120" />
      <ElTableColumn prop="status" label="状态" width="100" />
      <ElTableColumn label="默认模板" width="100">
        <template #default="{ row }">
          <ElTag v-if="row.is_default" type="success" effect="plain">是</ElTag>
          <span v-else>否</span>
        </template>
      </ElTableColumn>
      <ElTableColumn label="配置概览" min-width="220">
        <template #default="{ row }">
          <div class="config-badges">
            <ElTag v-if="hasConfigKey(row, 'theme')" size="small" effect="plain">theme</ElTag>
            <ElTag v-if="hasConfigKey(row, 'features')" size="small" effect="plain" type="success"
              >features</ElTag
            >
            <ElTag v-if="hasConfigKey(row, 'texts')" size="small" effect="plain" type="warning"
              >texts</ElTag
            >
            <ElTag v-if="hasConfigKey(row, 'pages')" size="small" effect="plain" type="info"
              >pages</ElTag
            >
            <span v-if="!hasAnyConfig(row)" class="text-gray-400 text-xs">未配置</span>
          </div>
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <ElButton link type="primary" @click="openEdit(row)">编辑</ElButton>
          <ElButton link type="info" @click="openPreviewOnly(row)">预览</ElButton>
          <ElPopconfirm title="确认删除该模板？" @confirm="remove(row)">
            <template #reference>
              <ElButton link type="danger">删除</ElButton>
            </template>
          </ElPopconfirm>
        </template>
      </ElTableColumn>
    </ElTable>

    <ElDialog
      v-model="dialogVisible"
      :title="editing ? '编辑模板' : '新建模板'"
      width="1160px"
      top="5vh"
    >
      <div class="dialog-layout">
        <ElForm :model="form" label-width="120px" class="dialog-form">
          <ElFormItem label="模板 Key" required>
            <ElInput
              v-model="form.template_key"
              :disabled="!!editing"
              placeholder="如 default / aurora"
            />
          </ElFormItem>
          <ElFormItem label="名称" required>
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

          <ElDivider content-position="left">全局 theme (主题配置)</ElDivider>
          <ElFormItem label="品牌色">
            <ElInput v-model="configTheme.primaryColor" placeholder="#409EFF" />
          </ElFormItem>
          <ElFormItem label="Logo URL">
            <ElInput v-model="configTheme.logoUrl" placeholder="https://..." />
          </ElFormItem>
          <ElFormItem label="背景图">
            <ElInput v-model="configTheme.backgroundImage" placeholder="url(...)" />
          </ElFormItem>
          <ElFormItem label="圆角">
            <ElInput v-model="configTheme.borderRadius" placeholder="8px" />
          </ElFormItem>

          <ElDivider content-position="left">全局 features (功能开关)</ElDivider>
          <ElFormItem label="社交登录">
            <ElSwitch v-model="configFeatures.socialLogin" />
          </ElFormItem>
          <ElFormItem label="验证码">
            <ElSwitch v-model="configFeatures.captcha" />
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

          <ElDivider content-position="left">全局 texts (自定义文案)</ElDivider>
          <ElFormItem label="标题">
            <ElInput v-model="configTexts.title" placeholder="欢迎登录" />
          </ElFormItem>
          <ElFormItem label="副标题">
            <ElInput v-model="configTexts.subTitle" placeholder="请输入账号密码" />
          </ElFormItem>
          <ElFormItem label="按钮文案">
            <ElInput v-model="configTexts.btnText" placeholder="登 录" />
          </ElFormItem>

          <ElDivider content-position="left">pages (按页面覆盖)</ElDivider>
          <ElTabs v-model="sceneEditorTab" class="scene-tabs">
            <ElTabPane v-for="scene in SCENES" :key="scene" :label="sceneLabel(scene)" :name="scene">
              <ElFormItem :label="`${sceneLabel(scene)}标题`">
                <ElInput v-model="pageOverrides[scene].title" placeholder="可选，留空则继承全局" />
              </ElFormItem>
              <ElFormItem :label="`${sceneLabel(scene)}副标题`">
                <ElInput v-model="pageOverrides[scene].subTitle" placeholder="可选，留空则继承全局" />
              </ElFormItem>
              <ElFormItem :label="`${sceneLabel(scene)}按钮文案`">
                <ElInput v-model="pageOverrides[scene].btnText" placeholder="可选，留空则继承全局" />
              </ElFormItem>
              <ElFormItem :label="`${sceneLabel(scene)}品牌色`">
                <ElInput v-model="pageOverrides[scene].primaryColor" placeholder="可选，留空则继承全局" />
              </ElFormItem>
              <ElFormItem :label="`${sceneLabel(scene)}圆角`">
                <ElInput v-model="pageOverrides[scene].borderRadius" placeholder="可选，留空则继承全局" />
              </ElFormItem>
            </ElTabPane>
          </ElTabs>

          <ElDivider content-position="left">高级</ElDivider>
          <ElFormItem label="原始 Config">
            <ElInput v-model="rawConfigText" type="textarea" :rows="4" />
            <div class="field-tip">
              直接编辑 JSON（保存时与上方表单合并，表单字段优先级高于此处）
            </div>
          </ElFormItem>
          <ElFormItem label="元数据">
            <ElInput v-model="rawMetaText" type="textarea" :rows="3" />
          </ElFormItem>
        </ElForm>

        <div class="dialog-preview">
          <div class="preview-header">
            <span class="preview-title">实时预览</span>
            <div class="preview-header-actions">
              <ElSelect v-model="previewScene" size="small" style="width: 130px">
                <ElOption label="登录页" value="login" />
                <ElOption label="注册页" value="register" />
                <ElOption label="找回密码" value="forget_password" />
              </ElSelect>
              <ElButton size="small" @click="refreshPreview">刷新</ElButton>
            </div>
          </div>
          <div class="preview-iframe-wrap">
            <iframe
              v-if="previewUrl"
              :key="previewKey"
              :src="previewUrl"
              class="preview-iframe"
              sandbox="allow-same-origin allow-scripts"
            />
            <div v-else class="preview-placeholder">
              <span>保存模板后可预览</span>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="submit">保存</ElButton>
      </template>
    </ElDialog>

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
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    fetchCreateLoginPageTemplate,
    fetchDeleteLoginPageTemplate,
    fetchListLoginPageTemplates,
    fetchUpdateLoginPageTemplate
  } from '@/domains/governance/api/register'

  defineOptions({ name: 'SystemLoginPageTemplate' })

  type SceneKey = 'login' | 'register' | 'forget_password'

  const SCENES: SceneKey[] = ['login', 'register', 'forget_password']

  const list = ref<any[]>([])
  const dialogVisible = ref(false)
  const editing = ref<any>(null)
  const previewDialogVisible = ref(false)
  const standalonePreviewTemplateKey = ref('')
  const standalonePreviewScene = ref<SceneKey>('login')
  const sceneEditorTab = ref<SceneKey>('login')
  const previewScene = ref<SceneKey>('login')
  const previewKey = ref(0)

  const form = reactive<any>({
    template_key: '',
    name: '',
    scene: 'auth_family',
    app_scope: 'shared',
    status: 'normal',
    is_default: false
  })

  const configTheme = reactive<any>({
    primaryColor: '',
    logoUrl: '',
    backgroundImage: '',
    borderRadius: ''
  })

  const configFeatures = reactive<any>({
    socialLogin: false,
    captcha: false,
    rememberMe: true,
    forgetPassword: true,
    register: true
  })

  const configTexts = reactive<any>({
    title: '',
    subTitle: '',
    btnText: ''
  })

  const pageOverrides = reactive<Record<SceneKey, any>>({
    login: { title: '', subTitle: '', btnText: '', primaryColor: '', borderRadius: '' },
    register: { title: '', subTitle: '', btnText: '', primaryColor: '', borderRadius: '' },
    forget_password: { title: '', subTitle: '', btnText: '', primaryColor: '', borderRadius: '' }
  })

  const rawConfigText = ref('{}')
  const rawMetaText = ref('{}')

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

  const previewUrl = computed(() => {
    const key = `${form.template_key || ''}`.trim()
    if (!key) return ''
    const path = scenePath(previewScene.value)
    return `${path}?login_page_key=${encodeURIComponent(key)}&preview=1`
  })

  const standalonePreviewUrl = computed(() => {
    const key = `${standalonePreviewTemplateKey.value || ''}`.trim()
    if (!key) return ''
    const path = scenePath(standalonePreviewScene.value)
    return `${path}?login_page_key=${encodeURIComponent(key)}&preview=1`
  })

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
      hasConfigKey(row, 'texts') ||
      hasConfigKey(row, 'pages')
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

  function buildPageOverrides(): Record<string, unknown> {
    const pages: Record<string, unknown> = {}
    for (const scene of SCENES) {
      const sceneData = pageOverrides[scene]
      const texts = cleanObject({
        title: sceneData.title,
        subTitle: sceneData.subTitle,
        btnText: sceneData.btnText
      })
      const theme = cleanObject({
        primaryColor: sceneData.primaryColor,
        borderRadius: sceneData.borderRadius
      })
      const block: Record<string, unknown> = {}
      if (Object.keys(texts).length > 0) block.texts = texts
      if (Object.keys(theme).length > 0) block.theme = theme
      if (Object.keys(block).length > 0) {
        pages[scene] = block
      }
    }
    return pages
  }

  function buildConfigPayload(): Record<string, unknown> {
    let base: Record<string, unknown> = {}
    try {
      base = JSON.parse(`${rawConfigText.value || '{}'}`) || {}
    } catch {
      /* ignore */
    }
    const theme = cleanObject({ ...configTheme })
    const features = cleanObject({ ...configFeatures })
    const texts = cleanObject({ ...configTexts })
    const pages = buildPageOverrides()
    if (Object.keys(theme).length > 0) base.theme = theme
    if (Object.keys(features).length > 0) base.features = features
    if (Object.keys(texts).length > 0) base.texts = texts
    if (Object.keys(pages).length > 0) base.pages = pages
    return base
  }

  function resetPageOverrides(config: any) {
    const pages = config?.pages || {}
    for (const scene of SCENES) {
      const sceneData = pages?.[scene] || (scene === 'forget_password' ? pages?.forgetPassword : {}) || {}
      const texts = sceneData?.texts || {}
      const theme = sceneData?.theme || {}
      Object.assign(pageOverrides[scene], {
        title: texts.title || '',
        subTitle: texts.subTitle || '',
        btnText: texts.btnText || '',
        primaryColor: theme.primaryColor || '',
        borderRadius: theme.borderRadius || ''
      })
    }
  }

  function resetConfigForm(config: any) {
    const theme = config?.theme || {}
    Object.assign(configTheme, {
      primaryColor: theme.primaryColor || '',
      logoUrl: theme.logoUrl || '',
      backgroundImage: theme.backgroundImage || '',
      borderRadius: theme.borderRadius || ''
    })
    const features = config?.features || {}
    Object.assign(configFeatures, {
      socialLogin: Boolean(features.socialLogin),
      captcha: Boolean(features.captcha),
      rememberMe: features.rememberMe !== false,
      forgetPassword: features.forgetPassword !== false,
      register: features.register !== false
    })
    const texts = config?.texts || {}
    Object.assign(configTexts, {
      title: texts.title || '',
      subTitle: texts.subTitle || '',
      btnText: texts.btnText || ''
    })
    resetPageOverrides(config)
  }

  const load = async () => {
    try {
      const data: any = await fetchListLoginPageTemplates()
      list.value = data?.records || []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    }
  }

  const openCreate = () => {
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
    rawConfigText.value = '{}'
    rawMetaText.value = '{}'
    sceneEditorTab.value = 'login'
    previewScene.value = 'login'
    dialogVisible.value = true
  }

  const openEdit = (row: any) => {
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
    rawConfigText.value = JSON.stringify(config, null, 2)
    rawMetaText.value = JSON.stringify(row.meta || {}, null, 2)
    sceneEditorTab.value = 'login'
    previewScene.value = 'login'
    dialogVisible.value = true
  }

  const openPreviewOnly = (row: any) => {
    const key = `${row.template_key || ''}`.trim()
    if (!key) return
    standalonePreviewTemplateKey.value = key
    standalonePreviewScene.value = 'login'
    previewDialogVisible.value = true
  }

  const refreshPreview = () => {
    previewKey.value++
  }

  const submit = async () => {
    const templateKey = `${form.template_key || ''}`.trim()
    const name = `${form.name || ''}`.trim()
    if (!templateKey || !name) {
      ElMessage.warning('请填写模板 Key 和名称')
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
      ElMessage.error(e?.message || '模板保存失败')
    }
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
</script>

<style scoped>
  .page-hero {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 16px;
  }

  .hero-desc {
    margin-top: 6px;
    max-width: 760px;
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .hero-actions {
    display: flex;
    gap: 12px;
  }

  .config-badges {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .dialog-layout {
    display: grid;
    grid-template-columns: minmax(0, 1.25fr) minmax(320px, 0.75fr);
    gap: 20px;
    max-height: 72vh;
    overflow-y: auto;
  }

  .dialog-form {
    min-width: 0;
  }

  .scene-tabs {
    margin-bottom: 8px;
  }

  .dialog-preview {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    position: sticky;
    top: 0;
    align-self: start;
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
  }

  .preview-title {
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .preview-iframe-wrap {
    border: 1px solid var(--el-border-color-light);
    border-radius: 12px;
    overflow: hidden;
    height: 520px;
    background: #f5f7fa;
  }

  .preview-iframe {
    width: 100%;
    height: 100%;
    border: none;
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
</style>
