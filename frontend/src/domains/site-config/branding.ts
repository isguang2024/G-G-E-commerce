// 站点品牌信息（name / logo / favicon）
//
// 统一封装站点配置的读取：
//   - 首次加载：`initSiteBranding()` 在应用启动时调用一次。
//   - 运行时：组件通过 `useSiteBranding()` 得到 `name/logo/favicon` 的 computed ref。
//     内部指向 pinia store 中的默认 bucket；store 更新后，ref 自动响应。
//   - DOM 同步：`favicon` 变化时会写入 `<link rel="icon">`，保证浏览器标签图标及时刷新。
//
// 约定 config_key：
//   - site.name    (string: { value: 'xxx' })
//   - site.logo    (image:  { url: 'https://...' })
//   - site.favicon (image:  { url: 'https://...' })
// 对应的默认回退来自 @/config → AppConfig.systemInfo.{name,logo,favicon}。

import { computed, watch } from 'vue'
import AppConfig from '@/config'
import { useSiteConfigStore } from '@/store/modules/site-config'

export const SITE_BRANDING_APP_KEY = 'admin'
export const SITE_BRANDING_KEYS = ['site.name', 'site.logo', 'site.favicon']

let initPromise: Promise<void> | null = null

/**
 * 启动时加载一次站点品牌配置。
 * 幂等：重复调用只会触发一次远程请求。失败时静默（维持默认值），不阻塞应用。
 */
export function initSiteBranding(): Promise<void> {
  if (initPromise) return initPromise
  const store = useSiteConfigStore()
  initPromise = store
    .loadInitial(SITE_BRANDING_APP_KEY, SITE_BRANDING_KEYS)
    .then(() => {
      applyFaviconToDom()
      registerFaviconWatcher()
    })
    .catch((err) => {
      // eslint-disable-next-line no-console
      console.warn('[site-config] initial load failed:', err)
    })
  return initPromise
}

/**
 * 组合式 API：在组件里拿到 name/logo/favicon 的响应式引用。
 */
export function useSiteBranding() {
  const store = useSiteConfigStore()
  const name = computed(() => store.getString('site.name', AppConfig.systemInfo.name))
  const logo = computed(() => store.getImage('site.logo', AppConfig.systemInfo.logo || ''))
  const favicon = computed(() =>
    store.getImage('site.favicon', AppConfig.systemInfo.favicon || '')
  )
  return { name, logo, favicon }
}

/**
 * 非响应式读取（router guard / document.title 等 setup 之外场景用）。
 */
export function getSiteName(): string {
  const store = useSiteConfigStore()
  return store.getString('site.name', AppConfig.systemInfo.name)
}
export function getSiteLogo(): string {
  const store = useSiteConfigStore()
  return store.getImage('site.logo', AppConfig.systemInfo.logo || '')
}
export function getSiteFavicon(): string {
  const store = useSiteConfigStore()
  return store.getImage('site.favicon', AppConfig.systemInfo.favicon || '')
}

// ── favicon DOM 同步 ─────────────────────────────────────────────────────

function applyFaviconToDom() {
  if (typeof document === 'undefined') return
  const url = getSiteFavicon()
  if (!url) return
  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  if (link.href !== url) {
    link.href = url
  }
}

let faviconWatcherRegistered = false
function registerFaviconWatcher() {
  if (faviconWatcherRegistered) return
  faviconWatcherRegistered = true
  const store = useSiteConfigStore()
  watch(
    () => store.getImage('site.favicon', ''),
    () => applyFaviconToDom()
  )
}
