export type WorkbenchMode = 'menu' | 'collaboration' | 'role' | 'user'
export type DecisionValue = '' | 'allow' | 'deny'

export interface WorkbenchActionItem extends Partial<Api.SystemManage.PermissionActionItem> {
  id: string
  name: string
  permissionKey?: string
  code?: string
}

export interface WorkbenchModuleGroup {
  key: string
  label: string
  actions: WorkbenchActionItem[]
  actionCount: number
  selectedCount: number
}

export interface WorkbenchFeatureGroup {
  key: string
  label: string
  modules: WorkbenchModuleGroup[]
  moduleCount: number
  actionCount: number
  selectedCount: number
}

export function uniqueByValue<T extends { value: string }>(items: T[]): T[] {
  const seen = new Set<string>()
  return items.filter((item) => {
    if (seen.has(item.value)) return false
    seen.add(item.value)
    return true
  })
}

export function formatFeature(feature: string): string {
  const map: Record<string, string> = {
    system: '系统功能',
    business: '业务功能'
  }
  return map[feature] || feature
}

export function formatModule(action: WorkbenchActionItem): string {
  return action.moduleGroup?.name || action.moduleCode || action.resourceCode || '未分类模块'
}

export function moduleKeywords(item: WorkbenchActionItem): string {
  return [
    item.moduleGroup?.name,
    item.moduleCode,
    item.resourceCode,
    item.featureGroup?.name,
    item.featureKind
  ]
    .filter(Boolean)
    .join(' ')
}

export function buildFeatureOptions(actions: WorkbenchActionItem[]) {
  return uniqueByValue(
    actions
      .map((item) => ({
        value: `${item.featureGroupId || item.featureKind || ''}`,
        label: item.featureGroup?.name || formatFeature(`${item.featureKind || ''}`)
      }))
      .filter((item) => item.value)
  )
}

export function buildDecisionOptions(mode: WorkbenchMode) {
  if (mode === 'user') {
    return [
      { value: '', label: '继承角色' },
      { value: 'allow', label: '单独允许' },
      { value: 'deny', label: '单独拒绝' }
    ]
  }
  return [
    { value: '', label: '未配置' },
    { value: 'allow', label: '允许' },
    { value: 'deny', label: '拒绝' }
  ]
}

export function buildStateOptions(mode: WorkbenchMode) {
  if (mode === 'menu') {
    return [
      { value: 'all', label: '选择状态：全部' },
      { value: 'selected', label: '选择状态：已选' },
      { value: 'unselected', label: '选择状态：未选' }
    ]
  }
  if (mode === 'collaboration') {
    return [
      { value: 'all', label: '开通状态：全部' },
      { value: 'selected', label: '开通状态：已开通' },
      { value: 'unselected', label: '开通状态：未开通' }
    ]
  }
  if (mode === 'user') {
    return [
      { value: 'all', label: '覆盖状态：全部' },
      { value: 'inherit', label: '覆盖状态：继承角色' },
      { value: 'allow', label: '覆盖状态：单独允许' },
      { value: 'deny', label: '覆盖状态：单独拒绝' }
    ]
  }
  return [
    { value: 'all', label: '配置状态：全部' },
    { value: 'unset', label: '配置状态：未配置' },
    { value: 'allow', label: '配置状态：允许' },
    { value: 'deny', label: '配置状态：拒绝' }
  ]
}

export function buildBatchCommands(mode: WorkbenchMode) {
  if (mode === 'menu' || mode === 'collaboration') {
    return [
      {
        command: 'select-visible',
        label: mode === 'collaboration' ? '批量开通当前结果' : '批量选中当前结果'
      },
      {
        command: 'clear-visible',
        label: mode === 'collaboration' ? '批量关闭当前结果' : '批量取消当前结果'
      },
      { command: 'clear-all', label: '清空全部配置' }
    ]
  }
  if (mode === 'user') {
    return [
      { command: 'inherit-visible', label: '批量继承角色' },
      { command: 'allow-visible', label: '批量单独允许' },
      { command: 'deny-visible', label: '批量单独拒绝' },
      { command: 'clear-all', label: '清空全部例外' }
    ]
  }
  return [
    { command: 'allow-visible', label: '批量允许当前结果' },
    { command: 'deny-visible', label: '批量拒绝当前结果' },
    { command: 'unset-visible', label: '批量取消当前配置' },
    { command: 'clear-all', label: '清空全部配置' }
  ]
}

export function getSelectedLabel(mode: WorkbenchMode): string {
  if (mode === 'user') return '例外'
  if (mode === 'role') return '已配置'
  if (mode === 'collaboration') return '已开通'
  return '已选'
}

export function getDecisionLabel(mode: WorkbenchMode, decision: DecisionValue): string {
  if (mode === 'user') {
    if (decision === 'allow') return '单独允许'
    if (decision === 'deny') return '单独拒绝'
    return '继承角色'
  }
  if (decision === 'allow') return '允许'
  if (decision === 'deny') return '拒绝'
  return '未配置'
}

export function getDecisionTagType(decision: DecisionValue): 'success' | 'danger' | 'info' {
  if (decision === 'allow') return 'success'
  if (decision === 'deny') return 'danger'
  return 'info'
}

export function normalizeDecision(value: string): DecisionValue {
  return value === 'allow' || value === 'deny' ? value : ''
}

export function groupActions(actions: WorkbenchActionItem[], isActive: (id: string) => boolean): WorkbenchFeatureGroup[] {
  const featureMap = new Map<string, WorkbenchFeatureGroup>()

  actions.forEach((item) => {
    const featureKey = `${item.featureGroupId || item.featureKind || 'business'}`
    const moduleKey = `${item.moduleGroupId || item.moduleCode || item.resourceCode || 'default'}`

    if (!featureMap.has(featureKey)) {
      featureMap.set(featureKey, {
        key: featureKey,
        label: item.featureGroup?.name || formatFeature(featureKey),
        modules: [],
        moduleCount: 0,
        actionCount: 0,
        selectedCount: 0
      })
    }

    const feature = featureMap.get(featureKey)!
    let module = feature.modules.find((entry) => entry.key === moduleKey)
    if (!module) {
      module = {
        key: moduleKey,
        label: formatModule(item),
        actions: [],
        actionCount: 0,
        selectedCount: 0
      }
      feature.modules.push(module)
    }

    module.actions.push(item)
    module.actionCount += 1
    feature.actionCount += 1

    if (isActive(item.id)) {
      module.selectedCount += 1
      feature.selectedCount += 1
    }
  })

  return Array.from(featureMap.values()).map((feature) => ({
    ...feature,
    moduleCount: feature.modules.length
  }))
}
