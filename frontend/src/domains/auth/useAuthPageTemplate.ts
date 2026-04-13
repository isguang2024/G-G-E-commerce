import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { fetchLoginPageContext } from '@/domains/auth/api'
import { useAppContextStore } from '@/domains/app-runtime/context'

type AuthPageScene = 'login' | 'register' | 'forget_password'

type LoginPageContext = Awaited<ReturnType<typeof fetchLoginPageContext>>

const LOGIN_PAGE_THEME_CLASS_MAP: Record<string, string> = {
  default: 'auth-theme-default',
  aurora: 'auth-theme-aurora'
}

/** theme 配置块：品牌色、Logo、背景图 */
export interface AuthTemplateTheme {
  primaryColor?: string
  logoUrl?: string
  backgroundImage?: string
  backgroundOverlay?: string
  borderRadius?: string
  [key: string]: unknown
}

/** features 配置块：功能开关 */
export interface AuthTemplateFeatures {
  socialLogin?: boolean
  socialProviders?: string[]
  socialItems?: AuthTemplateSocialItem[]
  socialCustomHtml?: string
  captcha?: boolean
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

/** texts 配置块：自定义文案 */
export interface AuthTemplateTexts {
  title?: string
  subTitle?: string
  btnText?: string
  placeholder?: Record<string, string>
  copyright?: string
  [key: string]: unknown
}

/** social 配置块：独立社交入口配置 */
export interface AuthTemplateSocial {
  items?: AuthTemplateSocialItem[]
  customHtml?: string
  capability?: AuthTemplateSocialCapability
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

export function useAuthPageTemplate(scene: AuthPageScene) {
  const route = useRoute()
  const appContextStore = useAppContextStore()
  const context = ref<LoginPageContext | null>(null)
  const loading = ref(false)
  const errorMessage = ref('')

  const loginPageKey = computed(() => {
    const queryValue = `${route.query.login_page_key || ''}`.trim()
    if (queryValue) return queryValue
    return `${context.value?.login_page_key || ''}`.trim() || 'default'
  })

  const themeClass = computed(() => {
    return LOGIN_PAGE_THEME_CLASS_MAP[loginPageKey.value] || LOGIN_PAGE_THEME_CLASS_MAP.default
  })

  const templateName = computed(() => {
    return `${(context.value as any)?.template_name || ''}`.trim()
  })

  // ── 模板配置三大块 ──────────────────────────────────────────
  const rawTemplateConfig = computed(() => {
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

  /** texts: 自定义文案（标题 / 副标题 / 按钮文案） */
  const texts = computed<AuthTemplateTexts>(() => {
    return mergePlainObject(
      normalizePlainObject(rawTemplateConfig.value.texts),
      normalizePlainObject(sceneTemplateConfig.value.texts)
    ) as AuthTemplateTexts
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
    void loadContext()
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
    texts,
    social,
    isPreview,
    registerPath,
    registerLink,
    resolvedBy,
    loginUiMode,
    loadContext
  }
}
