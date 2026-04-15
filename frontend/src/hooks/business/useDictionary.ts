import { ref, computed, type Ref, type ComputedRef } from 'vue'
import { fetchDictsByCodes, type DictItemSummary } from '@/api/system-manage/dictionary'
import { logger } from '@/utils/logger'

// ─── Types ───────────────────────────────────────────────────────────────────

export interface DictOption {
  label: string
  value: string
  extra?: Record<string, unknown>
}

// ─── Module-level cache ──────────────────────────────────────────────────────
// Cache lives at module scope (SPA lifecycle). Components unmounting does NOT
// clear it — this is by design to avoid redundant requests. Use invalidateDict()
// or invalidateAllDicts() to force a refetch.

const cache = new Map<string, DictItemSummary[]>()
const pendingCodes = new Set<string>()
let batchTimer: ReturnType<typeof setTimeout> | null = null
let batchResolvers: Array<() => void> = []

function scheduleBatchFetch() {
  if (batchTimer !== null) return
  batchTimer = setTimeout(async () => {
    const codes = [...pendingCodes]
    const resolvers = [...batchResolvers]
    pendingCodes.clear()
    batchResolvers = []
    batchTimer = null

    if (codes.length === 0) {
      resolvers.forEach((r) => r())
      return
    }

    try {
      const result = await fetchDictsByCodes(codes)
      for (const code of codes) {
        cache.set(code, (result as Record<string, DictItemSummary[]>)[code] ?? [])
      }
    } catch (err) {
      logger.warn('dictionary.batch_fetch_failed', { err, codes })
      // Set empty arrays so failed codes don't block future calls.
      // Use invalidateDict(code) to clear and allow retry.
      for (const code of codes) {
        if (!cache.has(code)) {
          cache.set(code, [])
        }
      }
    }
    resolvers.forEach((r) => r())
  }, 0)
}

function requestCodes(codes: string[]): Promise<void> {
  const missing = codes.filter((c) => !cache.has(c))
  if (missing.length === 0) return Promise.resolve()

  for (const c of missing) {
    pendingCodes.add(c)
  }
  return new Promise<void>((resolve) => {
    batchResolvers.push(resolve)
    scheduleBatchFetch()
  })
}

// ─── useDictionary ───────────────────────────────────────────────────────────

export function useDictionary(code: string): {
  options: Ref<DictOption[]>
  loading: Ref<boolean>
  map: ComputedRef<Record<string, string>>
} {
  const options = ref<DictOption[]>([])
  const loading = ref(true)

  requestCodes([code]).then(() => {
    const items = cache.get(code) ?? []
    options.value = items.map((item) => ({
      label: item.label,
      value: item.value,
      extra: item.extra as Record<string, unknown> | undefined
    }))
    loading.value = false
  })

  const map = computed(() => {
    const m: Record<string, string> = {}
    for (const opt of options.value) {
      m[opt.value] = opt.label
    }
    return m
  })

  return { options, loading, map }
}

// ─── useDictionaries ─────────────────────────────────────────────────────────

export function useDictionaries(codes: string[]): {
  dictMap: Ref<Record<string, DictOption[]>>
  loading: Ref<boolean>
} {
  const dictMap = ref<Record<string, DictOption[]>>({})
  const loading = ref(true)

  requestCodes(codes).then(() => {
    const result: Record<string, DictOption[]> = {}
    for (const code of codes) {
      const items = cache.get(code) ?? []
      result[code] = items.map((item) => ({
        label: item.label,
        value: item.value,
        extra: item.extra as Record<string, unknown> | undefined
      }))
    }
    dictMap.value = result
    loading.value = false
  })

  return { dictMap, loading }
}

// ─── Cache invalidation ──────────────────────────────────────────────────────

export function invalidateDict(code: string) {
  cache.delete(code)
}

export function invalidateAllDicts() {
  cache.clear()
}
