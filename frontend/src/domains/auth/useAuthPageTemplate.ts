import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchLoginPageContext } from '@/domains/auth/api'
import { useAppContextStore } from '@/domains/app-runtime/context'

type AuthPageScene = 'login' | 'register' | 'forget_password'

type LoginPageContext = Awaited<ReturnType<typeof fetchLoginPageContext>>

const LOGIN_PAGE_THEME_CLASS_MAP: Record<string, string> = {
  default: 'auth-theme-default',
  aurora: 'auth-theme-aurora'
}

/** theme 配置块：当前仅消费品牌色与圆角 */
export interface AuthTemplateTheme {
  primaryColor?: string
  borderRadius?: string
  [key: string]: unknown
}

/** features 配置块：功能开关 */
export interface AuthTemplateFeatures {
  socialLogin?: boolean
  socialProviders?: string[]
  socialItems?: AuthTemplateSocialItem[]
  socialCustomHtml?: string
  rememberMe?: boolean
  forgetPassword?: boolean
  register?: boolean
  [key: string]: unknown
}

export interface AuthTemplateSocialItem {
  key?: string
  name?: string
  icon?: string
  url?: string
  [key: string]: unknown
}

export interface AuthTemplateSocialCapability {
  allow?: boolean
  reason?: string
  providers?: string[]
  [key: string]: unknown
}

/** social 配置块：独立社交入口配置 */
export interface AuthTemplateSocial {
  items?: AuthTemplateSocialItem[]
  customHtml?: string
  capability?: AuthTemplateSocialCapability
  [key: string]: unknown
}

/** pages.<scene>.texts：仅页面级文案，不再支持全局 texts */
export interface AuthTemplateTexts {
  title?: string
  subTitle?: string
  buttonText?: string
  secondaryButtonText?: string
  [key: string]: unknown
}

function normalizePlainObject(val: unknown): Record<string, unknown> {
  if (val && typeof val === 'object' && !Array.isArray(val)) {
    return val as Record<string, unknown>
  }
  return {}
}

function mergePlainObject(
  base: Record<string, unknown>,
  override: Record<string, unknown>
): Record<string, unknown> {
  const merged: Record<string, unknown> = { ...base }
  for (const [k, v] of Object.entries(override)) {
    const baseVal = merged[k]
    if (
      baseVal &&
      typeof baseVal === 'object' &&
      !Array.isArray(baseVal) &&
      v &&
      typeof v === 'object' &&
      !Array.isArray(v)
    ) {
      merged[k] = mergePlainObject(
        normalizePlainObject(baseVal),
        normalizePlainObject(v)
      )
      continue
    }
    merged[k] = v
  }
  return merged
}

function authPreviewDraftStorageKey(id: string): string {
  return `auth-template-preview:${id}`
}

function readAuthPreviewDraft(id: string): Record<string, unknown> {
  if (!id || typeof window === 'undefined') return {}
  try {
    const raw = window.localStorage.getItem(authPreviewDraftStorageKey(id))
    if (!raw) return {}
    const parsed = JSON.parse(raw)
    return normalizePlainObject(parsed)
  } catch {
    return {}
  }
}

export function useAuthPageTemplate(scene: AuthPageScene) {
  const route = useRoute()
  const appContextStore = useAppContextStore()
  const context = ref<LoginPageContext | null>(null)
  const loading = ref(false)
  const errorMessage = ref('')
  const previewDraft = ref<Record<string, unknown>>({})

  const previewDraftID = computed(() => `${route.query.preview_draft_id || ''}`.trim())
  const handleStorage = (event: StorageEvent) => {
    if (!previewDraftID.value) return
    if (event.key !== authPreviewDraftStorageKey(previewDraftID.value)) return
    previewDraft.value = readAuthPreviewDraft(previewDraftID.value)
  }

  const loginPageKey = computed(() => {
    const queryValue = `${route.query.login_page_key || ''}`.trim()
    if (queryValue) return queryValue
    return `${context.value?.login_page_key || ''}`.trim() || 'default'
  })

  const themeClass = computed(() => {
    return LOGIN_PAGE_THEME_CLASS_MAP[loginPageKey.value] || LOGIN_PAGE_THEME_CLASS_MAP.default
  })

  const templateName = computed(() => {
    const draftName = `${previewDraft.value?.name || ''}`.trim()
    if (draftName) return draftName
    return `${(context.value as any)?.template_name || ''}`.trim()
  })

  // ── 模板配置三大块 ──────────────────────────────────────────
  const rawTemplateConfig = computed(() => {
    const draftConfig = normalizePlainObject(previewDraft.value.config)
    if (Object.keys(draftConfig).length > 0) {
      return draftConfig
    }
    return normalizePlainObject((context.value as any)?.template_config)
  })

  const sceneTemplateConfig = computed(() => {
    const pages = normalizePlainObject(rawTemplateConfig.value.pages)
    const sceneConfig =
      scene === 'forget_password'
        ? pages.forget_password ?? pages.forgetPassword
        : pages[scene]
    return normalizePlainObject(sceneConfig)
  })

  /** theme: 品牌色 / Logo / 背景图 */
  const theme = computed<AuthTemplateTheme>(() => {
    return mergePlainObject(
      normalizePlainObject(rawTemplateConfig.value.theme),
      normalizePlainObject(sceneTemplateConfig.value.theme)
    ) as AuthTemplateTheme
  })

  /** features: 功能开关（社交登录 / captcha / remember me 等） */
  const features = computed<AuthTemplateFeatures>(() => {
    return mergePlainObject(
      normalizePlainObject(rawTemplateConfig.value.features),
      normalizePlainObject(sceneTemplateConfig.value.features)
    ) as AuthTemplateFeatures
  })

  /** social: 社交登录入口（结构化项 + 可选受限 HTML） */
  const social = computed<AuthTemplateSocial>(() => {
    const merged = mergePlainObject(
      normalizePlainObject(rawTemplateConfig.value.social),
      normalizePlainObject(sceneTemplateConfig.value.social)
    ) as AuthTemplateSocial
    const featureItems = (features.value.socialItems || []) as AuthTemplateSocialItem[]
    const featureHtml = `${features.value.socialCustomHtml || ''}`.trim()
    if ((!merged.items || merged.items.length === 0) && featureItems.length > 0) {
      merged.items = featureItems
    }
    if (!`${merged.customHtml || ''}`.trim() && featureHtml) {
      merged.customHtml = featureHtml
    }
    return merged
  })

  /** texts: 仅当前页面自己的文案配置，不做全局继承 */
  const texts = computed<AuthTemplateTexts>(() => {
    return normalizePlainObject(sceneTemplateConfig.value.texts) as AuthTemplateTexts
  })

  const isPreview = computed(() => {
    return `${route.query.preview || ''}`.trim() === '1'
  })

  const registerPath = computed(() => {
    const target = `${context.value?.register_path || ''}`.trim()
    return target || '/account/auth/register'
  })

  const registerLink = computed(() => {
    const target = registerPath.value
    if (!target.startsWith('/')) return '/account/auth/register'
    if (!loginPageKey.value || loginPageKey.value === 'default') return target
    const link = new URL(target, window.location.origin)
    link.searchParams.set('login_page_key', loginPageKey.value)
    return `${link.pathname}${link.search}`
  })

  const targetAppKey = computed(() => {
    const fromQuery = `${route.query.target_app_key || ''}`.trim()
    if (fromQuery) return fromQuery
    return `${appContextStore.effectiveManagedAppKey || appContextStore.currentRuntimeAppKey || ''}`.trim()
  })

  const resolvedBy = computed(() => `${context.value?.resolved_by || ''}`.trim())
  const loginUiMode = computed(() => `${context.value?.login_ui_mode || ''}`.trim())

  async function loadContext() {
    loading.value = true
    try {
      context.value = await fetchLoginPageContext({
        host: window.location.host,
        path: window.location.pathname,
        target_app_key: targetAppKey.value || undefined,
        login_page_key: `${route.query.login_page_key || ''}`.trim() || undefined,
        page_scene: scene
      })
      errorMessage.value = ''
    } catch (error: any) {
      errorMessage.value = error?.message || '读取登录模板上下文失败'
      context.value = null
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    previewDraft.value = readAuthPreviewDraft(previewDraftID.value)
    window.addEventListener('storage', handleStorage)
    void loadContext()
  })

  watch(
    () => route.fullPath,
    () => {
      previewDraft.value = readAuthPreviewDraft(previewDraftID.value)
      void loadContext()
    }
  )

  onBeforeUnmount(() => {
    window.removeEventListener('storage', handleStorage)
  })

  return {
    context,
    loading,
    errorMessage,
    loginPageKey,
    templateName,
    themeClass,
    theme,
    features,
    social,
    texts,
    isPreview,
    registerPath,
    registerLink,
    resolvedBy,
    loginUiMode,
    loadContext
  }
}
