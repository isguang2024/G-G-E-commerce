export type PermissionActionItem = Api.SystemManage.PermissionActionItem
export type TreeNodeType = 'feature' | 'module' | 'action'

export interface ModuleGroup {
  key: string
  label: string
  count: number
  actionIds: string[]
  actions: PermissionActionItem[]
}

export interface FeatureGroup {
  key: string
  label: string
  count: number
  actionIds: string[]
  modules: ModuleGroup[]
}

export interface PermissionTreeNodeBase {
  key: string
  label: string
  nodeType: TreeNodeType
  meta: string
  children?: PermissionTreeNodeBase[]
  actionIds: string[]
  actionId?: string
}

export function buildPermissionGroups(actions: PermissionActionItem[]): FeatureGroup[] {
  const featureMap = new Map<string, FeatureGroup>()

  actions.forEach((action) => {
    const featureKey = action.featureGroup?.id || action.featureKind || 'system'
    const featureLabel =
      action.featureGroup?.name || (action.featureKind === 'business' ? '业务功能' : '系统功能')
    let featureGroup = featureMap.get(featureKey)
    if (!featureGroup) {
      featureGroup = {
        key: featureKey,
        label: featureLabel,
        count: 0,
        actionIds: [],
        modules: []
      }
      featureMap.set(featureKey, featureGroup)
    }

    const moduleCode =
      action.moduleGroup?.name || action.moduleCode || action.resourceCode || 'common'
    const moduleKey = `${featureKey}:${moduleCode}`
    let moduleGroup = featureGroup.modules.find((item) => item.key === moduleKey)
    if (!moduleGroup) {
      moduleGroup = {
        key: moduleKey,
        label: moduleCode,
        count: 0,
        actionIds: [],
        actions: []
      }
      featureGroup.modules.push(moduleGroup)
    }

    moduleGroup.actions.push(action)
    moduleGroup.actionIds.push(action.id)
    moduleGroup.count += 1
    featureGroup.actionIds.push(action.id)
    featureGroup.count += 1
  })

  return [...featureMap.values()]
    .map((featureGroup) => ({
      ...featureGroup,
      modules: featureGroup.modules
        .map((moduleGroup) => ({
          ...moduleGroup,
          actions: [...moduleGroup.actions].sort(
            (a, b) =>
              (a.sortOrder ?? 0) - (b.sortOrder ?? 0) ||
              (a.name || '').localeCompare(b.name || '', 'zh-CN')
          )
        }))
        .sort((a, b) => a.label.localeCompare(b.label, 'zh-CN'))
    }))
    .sort((a, b) => a.key.localeCompare(b.key, 'zh-CN'))
}

export function buildPermissionTree<TLeaf extends { meta: string } & Record<string, unknown>>(
  groups: FeatureGroup[],
  mapLeaf: (action: PermissionActionItem) => TLeaf
): Array<PermissionTreeNodeBase & Record<string, unknown>> {
  return groups.map((featureGroup) => ({
    key: featureGroup.key,
    label: featureGroup.label,
    nodeType: 'feature',
    meta: `${featureGroup.modules.length} 个模块，${featureGroup.count} 条权限`,
    actionIds: featureGroup.actionIds,
    children: featureGroup.modules.map((moduleGroup) => ({
      key: moduleGroup.key,
      label: moduleGroup.label,
      nodeType: 'module',
      meta: `${moduleGroup.count} 条权限`,
      actionIds: moduleGroup.actionIds,
      children: moduleGroup.actions.map((action) => ({
        key: action.id,
        label: action.name,
        nodeType: 'action',
        actionIds: [action.id],
        actionId: action.id,
        ...mapLeaf(action)
      }))
    }))
  }))
}

export function buildDefaultExpandedKeys(treeData: Array<PermissionTreeNodeBase>): string[] {
  return treeData.flatMap((featureGroup, index) => {
    if (index > 0) return []
    return [
      featureGroup.key,
      ...(featureGroup.children || []).map((moduleGroup) => moduleGroup.key)
    ]
  })
}

export function buildAllExpandedKeys(treeData: Array<PermissionTreeNodeBase>): string[] {
  return treeData.flatMap((featureGroup) => [
    featureGroup.key,
    ...(featureGroup.children || []).map((moduleGroup) => moduleGroup.key)
  ])
}
