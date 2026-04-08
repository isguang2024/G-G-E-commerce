/**
 * api-endpoint 视图：纯函数 formatter / tag-type 映射。
 * 抽离自 index.vue，避免视图脚本继续膨胀。
 */

export function methodTagType(method?: string) {
  switch (`${method || ''}`.toUpperCase()) {
    case 'POST':
      return 'success'
    case 'PUT':
      return 'warning'
    case 'DELETE':
      return 'danger'
    default:
      return 'info'
  }
}

export function sourceTagType(source?: string) {
  switch (source) {
    case 'manual':
      return 'warning'
    case 'seed':
      return 'success'
    default:
      return 'info'
  }
}

export function formatSource(source?: string) {
  switch (source) {
    case 'manual':
      return '手工维护'
    case 'seed':
      return '初始种子'
    default:
      return '自动同步'
  }
}

export function formatPermissionPattern(value?: string) {
  switch (`${value || ''}`.trim()) {
    case 'public':
      return '公开接口'
    case 'global_jwt':
      return '登录态全局'
    case 'self_jwt':
      return '登录态自服务'
    case 'api_key':
      return '开放 API Key'
    case 'single':
      return '单权限'
    case 'shared':
      return '多权限共享'
    case 'cross_context_shared':
      return '跨空间共享'
    default:
      return '无权限键'
  }
}

export function permissionPatternTagType(value?: string) {
  switch (`${value || ''}`.trim()) {
    case 'public':
      return 'success'
    case 'global_jwt':
      return 'info'
    case 'self_jwt':
      return 'warning'
    case 'api_key':
      return 'success'
    case 'single':
      return 'success'
    case 'shared':
      return 'warning'
    case 'cross_context_shared':
      return 'danger'
    default:
      return 'info'
  }
}

export function formatPermissionContext(value?: string) {
  switch (`${value || ''}`.trim()) {
    case 'personal':
      return '个人空间'
    case 'collaboration':
      return '协作空间'
    case 'common':
      return '通用'
    default:
      return value || '-'
  }
}

export function formatApiEndpointDisplayPath(path?: string, authMode?: string) {
  const normalizedPath = `${path || ''}`.trim()
  if (!normalizedPath) {
    return '-'
  }
  if (normalizedPath.startsWith('/api/v1/') || normalizedPath.startsWith('/open/v1/')) {
    return normalizedPath
  }
  const basePrefix = `${authMode || ''}`.trim() === 'public' ? '/open/v1' : '/api/v1'
  const suffix = normalizedPath.startsWith('/') ? normalizedPath : `/${normalizedPath}`
  return `${basePrefix}${suffix}`
}
