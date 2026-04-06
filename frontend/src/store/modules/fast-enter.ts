import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import AppConfig from '@/config'
import { fetchGetFastEnterConfig, fetchUpdateFastEnterConfig } from '@/api/system-manage'
import type { FastEnterApplication, FastEnterConfig, FastEnterQuickLink } from '@/types/config'
import { useUserStore } from './user'

type FastEnterItem = FastEnterApplication | FastEnterQuickLink
const FAST_ENTER_DEFAULT_MIN_WIDTH = 1450
const FAST_ENTER_CACHE_KEY_PREFIX = 'gg:fast-enter:config'
const FAST_ENTER_CACHE_TTL = 3 * 60 * 60 * 1000

interface FastEnterConfigCachePayload {
  expiresAt: number
  config: FastEnterConfig
}

function cloneConfig<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T
}

function createId(prefix: string) {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`
}

function normalizeItems<T extends FastEnterItem>(items: T[] | undefined, prefix: string): T[] {
  return (items || []).map((item, index) => ({
    ...item,
    id: item.id || createId(`${prefix}-${index + 1}`),
    enabled: item.enabled !== false,
    order: Number(item.order || index + 1)
  }))
}

function normalizeConfig(config: FastEnterConfig): FastEnterConfig {
  return {
    minWidth: FAST_ENTER_DEFAULT_MIN_WIDTH,
    applications: normalizeItems(config.applications, 'app'),
    quickLinks: normalizeItems(config.quickLinks, 'link')
  }
}

function buildFastEnterCacheKey(userScope: string) {
  return `${FAST_ENTER_CACHE_KEY_PREFIX}:${userScope || 'anonymous'}`
}

function readFastEnterCache(userScope: string): FastEnterConfig | null {
  if (typeof window === 'undefined') return null
  try {
    const raw = window.localStorage.getItem(buildFastEnterCacheKey(userScope))
    if (!raw) return null
    const payload = JSON.parse(raw) as FastEnterConfigCachePayload
    if (!payload?.expiresAt || payload.expiresAt <= Date.now() || !payload.config) {
      window.localStorage.removeItem(buildFastEnterCacheKey(userScope))
      return null
    }
    return normalizeConfig(payload.config)
  } catch {
    window.localStorage.removeItem(buildFastEnterCacheKey(userScope))
    return null
  }
}

function writeFastEnterCache(userScope: string, config: FastEnterConfig) {
  if (typeof window === 'undefined') return
  const payload: FastEnterConfigCachePayload = {
    expiresAt: Date.now() + FAST_ENTER_CACHE_TTL,
    config: normalizeConfig(config)
  }
  window.localStorage.setItem(buildFastEnterCacheKey(userScope), JSON.stringify(payload))
}

export function getDefaultFastEnterConfig(): FastEnterConfig {
  const base = cloneConfig(
    AppConfig.fastEnter || {
      applications: [],
      quickLinks: [],
      minWidth: FAST_ENTER_DEFAULT_MIN_WIDTH
    }
  )
  return normalizeConfig(base)
}

export const useFastEnterStore = defineStore('fastEnterStore', () => {
  const userStore = useUserStore()
  const defaultConfig = getDefaultFastEnterConfig()

  const minWidth = ref(FAST_ENTER_DEFAULT_MIN_WIDTH)
  const applications = ref<FastEnterApplication[]>(defaultConfig.applications)
  const quickLinks = ref<FastEnterQuickLink[]>(defaultConfig.quickLinks)
  const loaded = ref(false)
  const loading = ref(false)
  const loadedUserScope = ref('')

  const config = computed<FastEnterConfig>(() => ({
    minWidth: FAST_ENTER_DEFAULT_MIN_WIDTH,
    applications: cloneConfig(normalizeItems(applications.value, 'app')),
    quickLinks: cloneConfig(normalizeItems(quickLinks.value, 'link'))
  }))

  const replaceConfig = (nextConfig: FastEnterConfig) => {
    const normalized = normalizeConfig(nextConfig)
    minWidth.value = FAST_ENTER_DEFAULT_MIN_WIDTH
    applications.value = normalized.applications
    quickLinks.value = normalized.quickLinks
    loaded.value = true
    loadedUserScope.value = resolveUserScope()
  }

  const resolveUserScope = () => {
    const userInfo = userStore.getUserInfo as Partial<Api.Auth.UserInfo>
    return `${userInfo?.userId || userInfo?.id || userInfo?.userName || userInfo?.username || userInfo?.email || userStore.accessToken || 'anonymous'}`.trim()
  }

  const resetConfig = () => {
    replaceConfig(getDefaultFastEnterConfig())
  }

  const loadConfig = async (force = false) => {
    const userScope = resolveUserScope()
    if (loading.value) return config.value
    if (loaded.value && loadedUserScope.value === userScope && !force) return config.value
    if (!force) {
      const cachedConfig = readFastEnterCache(userScope)
      if (cachedConfig) {
        replaceConfig(cachedConfig)
        return config.value
      }
    }
    loading.value = true
    try {
      const remoteConfig = await fetchGetFastEnterConfig()
      replaceConfig(remoteConfig)
      writeFastEnterCache(userScope, remoteConfig)
      return config.value
    } finally {
      loading.value = false
    }
  }

  const saveConfig = async (nextConfig: FastEnterConfig) => {
    const userScope = resolveUserScope()
    const savedConfig = await fetchUpdateFastEnterConfig(nextConfig)
    replaceConfig(savedConfig)
    writeFastEnterCache(userScope, savedConfig)
    return savedConfig
  }

  return {
    minWidth,
    applications,
    quickLinks,
    loaded,
    loading,
    config,
    replaceConfig,
    resetConfig,
    loadConfig,
    saveConfig
  }
})
