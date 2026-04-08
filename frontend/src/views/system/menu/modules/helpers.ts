/**
 * menu 视图：纯函数辅助工具。
 *
 * 抽离自 index.vue，所有函数不依赖任何 reactive 状态，
 * 仅基于入参产出结果，便于单元测试与跨文件复用。
 */
import type { AppRouteRecord } from '@/types/router'

export const normalizeKeyword = (value?: string) => `${value || ''}`.trim().toLowerCase()

export const hashToNegativeNumber = (value: string) => {
  let hash = 0
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash * 31 + value.charCodeAt(i)) | 0
  }
  return -Math.abs(hash || 1)
}

export const isDirectoryMenuRow = (row: any) => `${row?.kind || ''}`.trim() === 'directory'
export const isEntryMenuRow = (row: any) => `${row?.kind || ''}`.trim() === 'entry'

export const getMenuTypeTag = (row: any) => {
  if (row.kind === 'external') return 'success'
  if (row.kind === 'entry') return 'primary'
  return 'info'
}

export const getMenuTypeText = (row: any) => {
  if (row.kind === 'external') return '外链'
  if (row.kind === 'entry') return '入口'
  return '目录'
}

export const getAccessModeLabel = (accessMode?: string) => {
  const mode = `${accessMode || 'permission'}`.trim()
  if (mode === 'jwt') return '登录可见'
  if (mode === 'public') return '公开可见'
  return '权限控制'
}

export const getAccessModeTag = (accessMode?: string) => {
  const mode = `${accessMode || 'permission'}`.trim()
  if (mode === 'jwt') return 'warning'
  if (mode === 'public') return 'success'
  return 'info'
}

export const cloneMenuNode = (
  item: AppRouteRecord,
  children: AppRouteRecord[]
): AppRouteRecord => ({
  ...item,
  meta: item.meta ? { ...item.meta } : item.meta,
  children
})
