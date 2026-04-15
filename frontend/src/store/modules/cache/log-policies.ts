import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { components } from '@/api/v5/schema'
import type { LogPolicyListQuery } from '@/domains/governance/api/observability'

type LogPolicyItem = components['schemas']['LogPolicyItem']
type LogPolicyList = components['schemas']['LogPolicyList']

type CacheEntry = {
  data: LogPolicyList
  cachedAt: number
}

const CACHE_TTL_MS = 30 * 1000

function toCacheKey(query: LogPolicyListQuery) {
  return JSON.stringify({
    current: query.current ?? 1,
    size: query.size ?? 20,
    pipeline: query.pipeline ?? '',
    enabled: typeof query.enabled === 'boolean' ? query.enabled : null
  })
}

function cloneListData(data: LogPolicyList): LogPolicyList {
  return {
    records: Array.isArray(data.records) ? data.records.map((item) => ({ ...item })) : [],
    total: Number(data.total || 0),
    current: Number(data.current || 1),
    size: Number(data.size || 20)
  }
}

export const useLogPoliciesCacheStore = defineStore('logPoliciesCacheStore', () => {
  const cache = ref<Record<string, CacheEntry>>({})

  const get = (query: LogPolicyListQuery): LogPolicyList | null => {
    const key = toCacheKey(query)
    const entry = cache.value[key]
    if (!entry) return null
    if (Date.now() - entry.cachedAt > CACHE_TTL_MS) {
      delete cache.value[key]
      return null
    }
    return cloneListData(entry.data)
  }

  const set = (query: LogPolicyListQuery, data: LogPolicyList) => {
    cache.value[toCacheKey(query)] = {
      data: cloneListData(data),
      cachedAt: Date.now()
    }
  }

  const clear = () => {
    cache.value = {}
  }

  const upsert = (item: LogPolicyItem) => {
    Object.keys(cache.value).forEach((key) => {
      const entry = cache.value[key]
      const nextRecords = entry.data.records.map((record) => (record.id === item.id ? { ...item } : record))
      entry.data = {
        ...entry.data,
        records: nextRecords
      }
    })
  }

  const remove = (id: string) => {
    Object.keys(cache.value).forEach((key) => {
      const entry = cache.value[key]
      const nextRecords = entry.data.records.filter((item) => item.id !== id)
      if (nextRecords.length === entry.data.records.length) return
      entry.data = {
        ...entry.data,
        records: nextRecords,
        total: Math.max(0, Number(entry.data.total || 0) - 1)
      }
    })
  }

  return {
    get,
    set,
    clear,
    upsert,
    remove
  }
})

