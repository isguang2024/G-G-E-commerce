import type { NavigationGroupKey, NavigationItem, NavIconKey, RouteStatus } from '@/shared/types/navigation'

export interface MenuSearchEntry {
  id: string
  routeId: string
  path: string
  label: string
  icon: NavIconKey
  group: NavigationGroupKey
  groupLabel: string
  status: RouteStatus
  trail: string[]
}

const groupLabels: Record<NavigationGroupKey, string> = {
  welcome: '首页',
  workspace: '工作台',
  team: '团队协作',
  message: '消息中心',
  system: '系统管理',
}

function normalizeText(value: string) {
  return value.trim().toLowerCase()
}

function hasNonAscii(value: string) {
  return /[^\x00-\x7F]/.test(value)
}

function splitPathSegments(path: string) {
  return path
    .split('/')
    .map((segment) => segment.trim())
    .filter(Boolean)
}

function buildLeafEntries(items: NavigationItem[], parents: NavigationItem[] = []): MenuSearchEntry[] {
  return items.flatMap((item) => {
    if (item.children?.length) {
      return buildLeafEntries(item.children, [...parents, item])
    }

    return [
      {
        id: item.id,
        routeId: item.routeId,
        path: item.path,
        label: item.label,
        icon: item.icon,
        group: item.group,
        groupLabel: groupLabels[item.group],
        status: item.status,
        trail: [...parents.map((parent) => parent.label), item.label],
      },
    ]
  })
}

function getSearchScore(entry: MenuSearchEntry, query: string) {
  const label = normalizeText(entry.label)
  const path = normalizeText(entry.path)
  const groupLabel = normalizeText(entry.groupLabel)
  const trail = normalizeText(entry.trail.join(' '))
  const pathSegments = splitPathSegments(path)
  const queryHasNonAscii = hasNonAscii(query)
  const allowBroadTextMatch = queryHasNonAscii || query.length >= 2
  const allowPathMatch = query.length >= 2

  if (label === query) return 100
  if (label.startsWith(query)) return 90
  if (allowBroadTextMatch && label.includes(query)) return 84
  if (entry.trail.some((segment) => normalizeText(segment).startsWith(query))) return 78
  if (allowBroadTextMatch && trail.includes(query)) return 72
  if (allowBroadTextMatch && groupLabel.includes(query)) return 58
  if (allowPathMatch && pathSegments.some((segment) => segment.startsWith(query))) return 48
  if (allowPathMatch && path.includes(query)) return 42
  return 0
}

export function buildMenuSearchEntries(items: NavigationItem[]) {
  const leafEntries = buildLeafEntries(items)
  const deduped = new Map<string, MenuSearchEntry>()

  for (const entry of leafEntries) {
    if (!deduped.has(entry.path)) {
      deduped.set(entry.path, entry)
    }
  }

  return [...deduped.values()]
}

export function filterMenuSearchEntries(entries: MenuSearchEntry[], rawQuery: string) {
  const query = normalizeText(rawQuery)
  if (!query) return []

  return entries
    .map((entry) => ({
      entry,
      score: getSearchScore(entry, query),
    }))
    .filter((item) => item.score > 0)
    .sort((left, right) => right.score - left.score || left.entry.label.localeCompare(right.entry.label, 'zh-CN'))
    .map((item) => item.entry)
}

export const MENU_SEARCH_HISTORY_KEY = 'frontend-fluentV2.menu-search-history'
export const MENU_SEARCH_HISTORY_LIMIT = 8
