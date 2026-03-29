import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import AppConfig from '@/config'
import { fetchGetFastEnterConfig, fetchUpdateFastEnterConfig } from '@/api/system-manage'
import type { FastEnterApplication, FastEnterConfig, FastEnterQuickLink } from '@/types/config'

type FastEnterItem = FastEnterApplication | FastEnterQuickLink

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

export function getDefaultFastEnterConfig(): FastEnterConfig {
  const base = cloneConfig(AppConfig.fastEnter || { applications: [], quickLinks: [], minWidth: 1200 })
  return {
    minWidth: Number(base.minWidth || 1200),
    applications: normalizeItems(base.applications, 'app'),
    quickLinks: normalizeItems(base.quickLinks, 'link')
  }
}

export const useFastEnterStore = defineStore('fastEnterStore', () => {
  const defaultConfig = getDefaultFastEnterConfig()

  const minWidth = ref(defaultConfig.minWidth || 1200)
  const applications = ref<FastEnterApplication[]>(defaultConfig.applications)
  const quickLinks = ref<FastEnterQuickLink[]>(defaultConfig.quickLinks)
  const loaded = ref(false)
  const loading = ref(false)

  const config = computed<FastEnterConfig>(() => ({
    minWidth: Number(minWidth.value || 1200),
    applications: cloneConfig(normalizeItems(applications.value, 'app')),
    quickLinks: cloneConfig(normalizeItems(quickLinks.value, 'link'))
  }))

  const replaceConfig = (nextConfig: FastEnterConfig) => {
    minWidth.value = Number(nextConfig.minWidth || 1200)
    applications.value = normalizeItems(nextConfig.applications, 'app')
    quickLinks.value = normalizeItems(nextConfig.quickLinks, 'link')
    loaded.value = true
  }

  const resetConfig = () => {
    replaceConfig(getDefaultFastEnterConfig())
  }

  const loadConfig = async (force = false) => {
    if (loading.value) return config.value
    if (loaded.value && !force) return config.value
    loading.value = true
    try {
      const remoteConfig = await fetchGetFastEnterConfig()
      replaceConfig(remoteConfig)
      return config.value
    } finally {
      loading.value = false
    }
  }

  const saveConfig = async (nextConfig: FastEnterConfig) => {
    const savedConfig = await fetchUpdateFastEnterConfig(nextConfig)
    replaceConfig(savedConfig)
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
