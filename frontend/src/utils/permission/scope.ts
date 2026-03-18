export function formatScopeLabel(scopeCode?: string, scopeName?: string, contextKind?: string) {
  if (scopeName?.trim()) return scopeName.trim()

  switch ((contextKind || scopeCode || '').trim()) {
    case 'tenant':
      return '团队'
    case 'global':
      return '全局'
    default:
      return (scopeCode || '未知').trim()
  }
}

export function getScopeTagType(scopeCode?: string, contextKind?: string) {
  switch ((contextKind || scopeCode || '').trim()) {
    case 'tenant':
      return 'success'
    case 'global':
      return 'primary'
    default:
      return 'info'
  }
}
