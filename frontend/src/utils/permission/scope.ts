export function formatScopeLabel(scopeCode?: string, scopeName?: string) {
  if (scopeName?.trim()) return scopeName.trim()

  switch ((scopeCode || '').trim()) {
    case 'team':
    case 'tenant':
      return '团队'
    case 'global':
      return '全局'
    default:
      return '未知'
  }
}

export function getScopeTagType(scopeCode?: string) {
  switch ((scopeCode || '').trim()) {
    case 'team':
    case 'tenant':
      return 'success'
    case 'global':
      return 'primary'
    default:
      return 'info'
  }
}
