export function formatScopeLabel(scopeCode?: string, scopeName?: string) {
  if (scopeName?.trim()) return scopeName.trim()

  switch ((scopeCode || '').trim()) {
    case 'tenant':
    case 'team':
      return '团队'
    case 'global':
      return '全局'
    default:
      return (scopeCode || '未知').trim()
  }
}

export function getScopeTagType(scopeCode?: string) {
  switch ((scopeCode || '').trim()) {
    case 'tenant':
    case 'team':
      return 'success'
    case 'global':
      return 'primary'
    default:
      return 'info'
  }
}
