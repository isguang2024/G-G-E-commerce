/**
 * site-config 前端状态管理
 *
 * 职责：
 *   1. 管理端：configs / sets 列表的加载、增删改，缓存在 store；写入后失效 resolved 缓存。
 *   2. 运行时：按 (appKey, keys/setCodes) 指纹缓存 resolveSiteConfigs 的结果，
 *      后续组件读取 key 时直接从内存拿；resolved 内部按 config_key 建立索引，
 *      UI 组件可通过 `getValue(key)` / `getImage(key)` 等 helper 直接读取。
 *   3. 启动时 `loadInitial(appKey, keys)` 加载站点级基础配置（名称、Logo、favicon），
 *      任一标签页写入后通过 BroadcastChannel 通知其它 tab 失效并自动重新 resolve 默认 bucket。
 *
 * 持久化：
 *   - resolved 结果不落 localStorage，仅存内存 + sessionStorage（避免跨登录污染）。
 *   - 登录后如果 user 变化，调用 `resetRuntime()` 清空。
 */

import { defineStore } from 'pinia'
import { computed, ref, shallowRef } from 'vue'

import {
  fetchDeleteSiteConfig,
  fetchDeleteSiteConfigSet,
  fetchListSiteConfigSets,
  fetchListSiteConfigs,
  fetchResolveSiteConfigs,
  fetchUpdateSiteConfig,
  fetchUpdateSiteConfigSet,
  fetchUpdateSiteConfigSetItems,
  fetchUpsertSiteConfig,
  fetchUpsertSiteConfigSet
} from '@/domains/site-config/api'
import type {
  SiteConfigResolveResponse,
  SiteConfigResolvedItem,
  SiteConfigSaveRequest,
  SiteConfigSetItemsRequest,
  SiteConfigSetSaveRequest,
  SiteConfigSetSummary,
  SiteConfigSummary
} from '@/domains/site-config/types'
import {
  readResolvedBool,
  readResolvedImage,
  readResolvedNumber,
  readResolvedString
} from '@/domains/site-config/types'
import { logger } from '@/utils/logger'

interface ResolvedBucket {
  appKey: string
  keys: string[]
  setCodes: string[]
  version: string
  items: Record<string, SiteConfigResolvedItem>
}

const SITE_CONFIG_BROADCAST_CHANNEL = 'gg:site-config:sync'

type SiteConfigSyncMessage = {
  type: 'invalidate'
  senderId: string
  scope?: 'all' | 'default'
}

function createClientId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

function normalizeList(values?: string[]): string[] {
  if (!values) return []
  const seen = new Set<string>()
  const out: string[] = []
  for (const raw of values) {
    const v = raw?.trim()
    if (!v || seen.has(v)) continue
    seen.add(v)
    out.push(v)
  }
  out.sort()
  return out
}

function bucketKey(appKey: string, keys: string[], setCodes: string[]): string {
  return `${appKey || '__global__'}::k=${keys.join(',')}::s=${setCodes.join(',')}`
}

export const useSiteConfigStore = defineStore('siteConfigStore', () => {
  // ── 管理端状态 ────────────────────────────────────────────────────────────
  const configs = ref<SiteConfigSummary[]>([])
  const configsScope = ref<string | undefined>(undefined)
  const configsLoading = ref(false)

  const sets = ref<SiteConfigSetSummary[]>([])
  const setsLoading = ref(false)

  // ── 运行时 resolved 缓存 ──────────────────────────────────────────────────
  // 用 shallowRef + Map 避免深度响应式对性能的拖累。
  const buckets = shallowRef<Map<string, ResolvedBucket>>(new Map())

  // 默认 bucket（loadInitial 设置），便于 getValue() 等不传参直接用。
  const defaultBucketKey = ref<string>('')

  // ── getters ───────────────────────────────────────────────────────────────

  const currentDefaultBucket = computed<ResolvedBucket | undefined>(() => {
    if (!defaultBucketKey.value) return undefined
    return buckets.value.get(defaultBucketKey.value)
  })

  function pickBucket(key?: string): ResolvedBucket | undefined {
    if (key) return buckets.value.get(key)
    return currentDefaultBucket.value
  }

  function getItem(
    configKey: string,
    bucketKeyOverride?: string
  ): SiteConfigResolvedItem | undefined {
    const bucket = pickBucket(bucketKeyOverride)
    if (!bucket) return undefined
    return bucket.items[configKey]
  }

  function getString(configKey: string, fallback = '', bucketKeyOverride?: string): string {
    return readResolvedString(getItem(configKey, bucketKeyOverride), fallback)
  }
  function getNumber(configKey: string, fallback = 0, bucketKeyOverride?: string): number {
    return readResolvedNumber(getItem(configKey, bucketKeyOverride), fallback)
  }
  function getBool(configKey: string, fallback = false, bucketKeyOverride?: string): boolean {
    return readResolvedBool(getItem(configKey, bucketKeyOverride), fallback)
  }
  function getImage(configKey: string, fallback = '', bucketKeyOverride?: string): string {
    return readResolvedImage(getItem(configKey, bucketKeyOverride), fallback)
  }

  // ── 运行时 resolve ─────────────────────────────────────────────────────────

  interface ResolveOptions {
    appKey?: string
    keys?: string[]
    setCodes?: string[]
    force?: boolean
    setDefault?: boolean
  }

  async function resolve(opts: ResolveOptions = {}): Promise<ResolvedBucket> {
    const appKey = opts.appKey || ''
    const keys = normalizeList(opts.keys)
    const setCodes = normalizeList(opts.setCodes)
    const key = bucketKey(appKey, keys, setCodes)
    if (!opts.force) {
      const cached = buckets.value.get(key)
      if (cached) {
        if (opts.setDefault) defaultBucketKey.value = key
        return cached
      }
    }
    const resp: SiteConfigResolveResponse = await fetchResolveSiteConfigs({
      appKey,
      keys,
      setCodes
    })
    const bucket: ResolvedBucket = {
      appKey,
      keys,
      setCodes,
      version: resp.version,
      items: resp.items ?? {}
    }
    const next = new Map(buckets.value)
    next.set(key, bucket)
    buckets.value = next
    if (opts.setDefault) defaultBucketKey.value = key
    return bucket
  }

  // 启动时加载（作为默认 bucket）
  async function loadInitial(appKey: string, keys: string[]): Promise<ResolvedBucket> {
    return resolve({ appKey, keys, setDefault: true })
  }

  interface InvalidateOptions {
    /** 是否向其他标签页广播失效信号（默认 true）。 */
    broadcast?: boolean
    /** 失效后是否自动重新拉取默认 bucket（默认 true，只有存在 defaultBucketKey 时才触发）。 */
    reloadDefault?: boolean
  }

  function invalidateResolved(options: InvalidateOptions = {}): void {
    const prevKey = defaultBucketKey.value
    const prevBucket = prevKey ? buckets.value.get(prevKey) : undefined
    buckets.value = new Map()
    // 保留 defaultBucketKey，以便 UI 的 computed 继续走 currentDefaultBucket 路径。

    if (options.broadcast !== false) {
      emitSyncMessage({ type: 'invalidate', senderId: clientId, scope: 'all' })
    }

    if (options.reloadDefault !== false && prevBucket) {
      void resolve({
        appKey: prevBucket.appKey,
        keys: prevBucket.keys,
        setCodes: prevBucket.setCodes,
        force: true,
        setDefault: true
      }).catch((error) => {
        logger.warn('site-config.reload_default_failed', { err: error })
      })
    }
  }

  function resetRuntime(): void {
    // 登出 / 账号切换时的硬清空，不广播、不重载。
    buckets.value = new Map()
    defaultBucketKey.value = ''
  }

  // ── 跨标签页同步 (BroadcastChannel) ───────────────────────────────────────
  const clientId = createClientId()
  let channel: BroadcastChannel | null = null
  if (typeof window !== 'undefined' && typeof window.BroadcastChannel === 'function') {
    try {
      channel = new window.BroadcastChannel(SITE_CONFIG_BROADCAST_CHANNEL)
      channel.addEventListener('message', (event) => {
        const data = event.data as SiteConfigSyncMessage | undefined
        if (!data || data.senderId === clientId) return
        if (data.type === 'invalidate') {
          // 来自其他 tab 的失效通知：清空本地 bucket 并自动重 resolve 默认 bucket（若存在）。
          invalidateResolved({ broadcast: false, reloadDefault: true })
        }
      })
    } catch (error) {
      logger.warn('site-config.broadcast_init_failed', { err: error })
      channel = null
    }
  }

  function emitSyncMessage(message: SiteConfigSyncMessage): void {
    if (!channel) return
    try {
      channel.postMessage(message)
    } catch (error) {
      logger.warn('site-config.broadcast_post_failed', { err: error })
    }
  }

  // ── 管理端 CRUD：Configs ──────────────────────────────────────────────────

  async function listConfigs(appKey?: string, force = false) {
    if (
      !force &&
      configs.value.length > 0 &&
      configsScope.value === appKey &&
      !configsLoading.value
    ) {
      return configs.value
    }
    configsLoading.value = true
    try {
      const resp = await fetchListSiteConfigs(appKey)
      configs.value = resp.records ?? []
      configsScope.value = appKey
      return configs.value
    } finally {
      configsLoading.value = false
    }
  }

  async function upsertConfig(body: SiteConfigSaveRequest): Promise<SiteConfigSummary> {
    const saved = await fetchUpsertSiteConfig(body)
    await listConfigs(configsScope.value, true)
    invalidateResolved()
    return saved
  }

  async function updateConfig(
    id: string,
    body: SiteConfigSaveRequest
  ): Promise<SiteConfigSummary> {
    const saved = await fetchUpdateSiteConfig(id, body)
    await listConfigs(configsScope.value, true)
    invalidateResolved()
    return saved
  }

  async function deleteConfig(id: string): Promise<void> {
    await fetchDeleteSiteConfig(id)
    configs.value = configs.value.filter((item) => item.id !== id)
    invalidateResolved()
  }

  // ── 管理端 CRUD：Sets ─────────────────────────────────────────────────────

  async function listSets(force = false) {
    if (!force && sets.value.length > 0 && !setsLoading.value) {
      return sets.value
    }
    setsLoading.value = true
    try {
      const resp = await fetchListSiteConfigSets()
      sets.value = resp.records ?? []
      return sets.value
    } finally {
      setsLoading.value = false
    }
  }

  async function upsertSet(body: SiteConfigSetSaveRequest): Promise<SiteConfigSetSummary> {
    const saved = await fetchUpsertSiteConfigSet(body)
    await listSets(true)
    invalidateResolved()
    return saved
  }

  async function updateSet(
    id: string,
    body: SiteConfigSetSaveRequest
  ): Promise<SiteConfigSetSummary> {
    const saved = await fetchUpdateSiteConfigSet(id, body)
    await listSets(true)
    invalidateResolved()
    return saved
  }

  async function deleteSet(id: string): Promise<void> {
    await fetchDeleteSiteConfigSet(id)
    sets.value = sets.value.filter((item) => item.id !== id)
    invalidateResolved()
  }

  async function updateSetItems(
    id: string,
    body: SiteConfigSetItemsRequest
  ): Promise<SiteConfigSetSummary> {
    const saved = await fetchUpdateSiteConfigSetItems(id, body)
    await listSets(true)
    invalidateResolved()
    return saved
  }

  return {
    // 管理端
    configs,
    configsScope,
    configsLoading,
    sets,
    setsLoading,
    listConfigs,
    upsertConfig,
    updateConfig,
    deleteConfig,
    listSets,
    upsertSet,
    updateSet,
    deleteSet,
    updateSetItems,
    // 运行时
    buckets,
    defaultBucketKey,
    currentDefaultBucket,
    resolve,
    loadInitial,
    invalidateResolved,
    resetRuntime,
    getItem,
    getString,
    getNumber,
    getBool,
    getImage
  }
})
