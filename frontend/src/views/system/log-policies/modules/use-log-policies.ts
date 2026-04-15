import type { components } from '@/api/v5/schema'
import {
  fetchCreateLogPolicy,
  fetchDeleteLogPolicy,
  fetchListLogPolicies,
  fetchPreviewLogPolicy,
  fetchUpdateLogPolicy,
  type LogPolicyCreateBody,
  type LogPolicyListQuery,
  type LogPolicyPreviewBody,
  type LogPolicyUpdateBody
} from '@/domains/governance/api/observability'
import { useLogPoliciesCacheStore } from '@/store/modules/cache/log-policies'

type LogPolicyItem = components['schemas']['LogPolicyItem']
type LogPolicyList = components['schemas']['LogPolicyList']
type LogPolicyPreviewResponse = components['schemas']['LogPolicyPreviewResponse']

export function useLogPolicies() {
  const cacheStore = useLogPoliciesCacheStore()

  const list = async (query: LogPolicyListQuery, force = false): Promise<LogPolicyList> => {
    if (!force) {
      const cached = cacheStore.get(query)
      if (cached) return cached
    }
    const data = await fetchListLogPolicies(query)
    cacheStore.set(query, data)
    return data
  }

  const create = async (body: LogPolicyCreateBody): Promise<LogPolicyItem> => {
    const item = await fetchCreateLogPolicy(body)
    cacheStore.clear()
    return item
  }

  const update = async (id: string, body: LogPolicyUpdateBody): Promise<LogPolicyItem> => {
    const item = await fetchUpdateLogPolicy(id, body)
    cacheStore.upsert(item)
    return item
  }

  const remove = async (id: string) => {
    await fetchDeleteLogPolicy(id)
    cacheStore.remove(id)
  }

  const preview = async (body: LogPolicyPreviewBody): Promise<LogPolicyPreviewResponse> => {
    return fetchPreviewLogPolicy(body)
  }

  const clearCache = () => cacheStore.clear()

  return {
    list,
    create,
    update,
    remove,
    preview,
    clearCache
  }
}

