/**
 * menu-space 视图：纯函数辅助工具。
 *
 * 抽离自 index.vue，所有函数不依赖任何 reactive 状态，
 * 仅基于入参产出结果，便于单元测试与跨文件复用。
 */
import type { AppRouteRecord } from '@/types/router'
import { logger } from '@/utils/logger'

export function isSpaceInitialized(item?: Api.SystemManage.MenuSpaceItem) {
  if (!item) return false
  return Number(item.menuCount || 0) > 0 || Number(item.pageCount || 0) > 0
}

export function normalizeRoleCodeListText(value: string) {
  return Array.from(
    new Set(
      `${value || ''}`
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean)
    )
  )
}

export function getAccessModeLabel(value?: string) {
  switch (`${value || 'all'}`.trim()) {
    case 'personal_workspace_admin':
      return '仅个人空间管理员'
    case 'collaboration_admin':
      return '仅协作空间管理员'
    case 'role_codes':
      return '指定空间角色码'
    default:
      return '全部可进'
  }
}

export function getAccessModeSummary(item?: Api.SystemManage.MenuSpaceItem) {
  if (!item) return '全部可进'
  if (`${item.accessMode || 'all'}`.trim() !== 'role_codes') {
    return getAccessModeLabel(item.accessMode)
  }
  const codes = item.allowedRoleCodes || []
  return codes.length ? `指定空间角色码 · ${codes.join(' / ')}` : '指定空间角色码'
}

export function getHostAuthModeLabel(value?: string) {
  switch (`${value || 'inherit_host'}`.trim()) {
    case 'centralized_login':
      return '统一登录 Host'
    case 'shared_cookie':
      return '共享 Cookie 域'
    default:
      return '沿用当前 Host'
  }
}

export function normalizeInternalPath(value: string): string {
  const target = `${value || ''}`.trim()
  if (!target || /^https?:\/\//i.test(target)) {
    return ''
  }
  const normalized = `/${target.replace(/^\/+/, '')}`.replace(/\/+/g, '/')
  return normalized === '/' ? normalized : normalized.replace(/\/$/, '')
}

export function collectMenuPaths(items: AppRouteRecord[] = []): string[] {
  const result: string[] = []
  const joinMenuPath = (parentPath: string, currentPath: string) => {
    const target = `${currentPath || ''}`.trim()
    if (!target) return ''
    if (/^https?:\/\//i.test(target)) return ''
    if (target.startsWith('/')) {
      return normalizeInternalPath(target)
    }
    const base = normalizeInternalPath(parentPath)
    return normalizeInternalPath(`${base}/${target}`)
  }
  const walk = (list: AppRouteRecord[], parentPath = '') => {
    ;(list || []).forEach((item) => {
      const normalizedPath = joinMenuPath(parentPath, `${item.path || ''}`)
      if (normalizedPath && !item.children?.length && item.meta?.isEnable !== false) {
        result.push(normalizedPath)
      }
      if (item.children?.length) {
        walk(item.children, normalizedPath || parentPath)
      }
    })
  }
  walk(items)
  return result
}

export function warnDev(event: string, context?: Record<string, unknown>) {
  if (import.meta.env.DEV) {
    logger.debug(`system.menu_space.${event}`, context)
  }
}
