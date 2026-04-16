import { fetchGetApps } from '@/domains/governance/api'

export interface AppCatalogOption {
  appKey: string
  name: string
  description: string
  label: string
  searchText: string
  isDefault: boolean
}

let cachedOptions: AppCatalogOption[] = []
let pendingRequest: Promise<AppCatalogOption[]> | null = null

function normalizeOption(item: Api.SystemManage.AppItem): AppCatalogOption {
  const appKey = `${item.appKey || ''}`.trim()
  const name = `${item.name || ''}`.trim()
  const description = `${item.description || ''}`.trim()
  const label = name ? `${name}（${appKey}）` : appKey
  const searchText = `${name} ${appKey} ${description}`.trim().toLowerCase()
  return {
    appKey,
    name,
    description,
    label,
    searchText,
    isDefault: Boolean(item.isDefault)
  }
}

export async function loadAppCatalog(force = false): Promise<AppCatalogOption[]> {
  if (cachedOptions.length > 0 && !force) {
    return cachedOptions
  }
  if (pendingRequest && !force) {
    return pendingRequest
  }

  pendingRequest = fetchGetApps()
    .then((result) => {
      const next = (result.records || [])
        .map(normalizeOption)
        .filter((item) => item.appKey)
        .sort((a, b) => {
          const aDefault = a.isDefault ? 1 : 0
          const bDefault = b.isDefault ? 1 : 0
          if (aDefault !== bDefault) return bDefault - aDefault
          return a.label.localeCompare(b.label, 'zh-CN')
        })
      cachedOptions = next
      return next
    })
    .finally(() => {
      pendingRequest = null
    })

  return pendingRequest
}

export function clearAppCatalogCache() {
  cachedOptions = []
  pendingRequest = null
}
