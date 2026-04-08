<template>
  <ElDrawer
    v-model="visible"
    :title="`角色权限配置 - ${roleTitle}`"
    size="1120px"
    destroy-on-close
    class="role-permission-dialog config-drawer"
    direction="rtl"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        请先为角色绑定功能包，再在角色已绑定功能包范围内裁剪菜单权限、功能权限和数据权限。菜单权限控制入口可见，功能权限勾选后即允许，不勾选默认不允许，数据权限控制资源范围。
      </div>
      <PermissionSummaryTags :items="summaryItems" />
      <div v-if="featurePackages.length" class="package-tags">
        <ElTag
          v-for="item in featurePackages"
          :key="item.id"
          type="success"
          effect="plain"
          round
          class="package-tag-link"
          @click="goToFeaturePackagePage(item)"
        >
          {{ item.name }}
        </ElTag>
      </div>

      <ElTabs v-model="activeTab" class="permission-tabs">
        <ElTabPane label="菜单权限" name="menus">
          <div class="cascader-pane">
            <div class="cascader-toolbar">
              <div class="cascader-tags">
                <ElTag type="success" effect="plain" round
                  >已保留 {{ expandedSelectedMenuIds.length }}</ElTag
                >
                <ElTag type="info" effect="plain" round>树节点 {{ menuNodeCount }}</ElTag>
              </div>

              <div class="toolbar-filters">
                <ElInput
                  v-model="menuKeyword"
                  clearable
                  placeholder="搜索菜单标题或路由"
                  :prefix-icon="Search"
                  class="toolbar-search"
                />
                <span class="toolbar-switch">
                  <span class="toolbar-switch__label">显示隐藏</span>
                  <ElSwitch v-model="showHiddenMenus" size="small" />
                </span>
                <span class="toolbar-switch">
                  <span class="toolbar-switch__label">显示内嵌</span>
                  <ElSwitch v-model="showIframeMenus" size="small" />
                </span>
                <span class="toolbar-switch">
                  <span class="toolbar-switch__label">显示启用</span>
                  <ElSwitch v-model="showEnabledMenus" size="small" />
                </span>
                <span class="toolbar-switch">
                  <span class="toolbar-switch__label">显示路径</span>
                  <ElSwitch v-model="showMenuPath" size="small" />
                </span>
              </div>
            </div>

            <div class="cascader-card">
              <ElCascaderPanel
                ref="menuPanelRef"
                v-model="selectedMenuNodeValues"
                :options="filteredMenuOptions"
                :props="menuCascaderProps"
                class="permission-cascader"
              >
                <template #default="{ node, data }">
                  <div class="panel-node" :class="{ 'is-leaf': node.isLeaf }">
                    <div class="panel-node__main">
                      <span class="panel-node__label">{{ data.label }}</span>
                      <span v-if="showMenuPath && data.path" class="panel-node__meta">{{
                        data.path
                      }}</span>
                    </div>

                    <ElTag
                      v-if="!node.isLeaf"
                      size="small"
                      effect="plain"
                      round
                      type="info"
                      class="panel-node__count"
                    >
                      {{ `${data.selectedLeafCount || 0}/${data.totalLeafCount || 0}` }}
                    </ElTag>
                  </div>
                </template>
              </ElCascaderPanel>
            </div>

            <div class="cascader-footer">
              <span class="footer-note">角色菜单权限决定可见入口和导航结构。</span>
              <ElButton text @click="clearMenuSelection">清空选择</ElButton>
            </div>
          </div>
        </ElTabPane>

        <ElTabPane label="功能权限" name="actions">
          <div class="cascader-pane">
            <PermissionSourcePanels
              v-model="selectedActionSourcePackageId"
              :packages="featurePackages"
              :source-map="actionSourceMapText"
              :derived-items="derivedActionItems"
              :blocked-items="disabledActionItems"
              derived-title="功能包展开能力"
              blocked-title="当前角色已关闭能力"
              open="actions"
              filtered-blocked-empty-text="当前筛选下暂无角色显式关闭能力"
              empty-title="当前暂无角色能力来源"
              empty-text="请先为角色绑定功能包，或检查个人空间角色快照是否已经刷新。"
            />

            <div class="cascader-toolbar">
              <div class="cascader-tags">
                <ElTag v-for="item in topLevelActionTags" :key="item.label" effect="plain" round>
                  {{ item.label }} {{ item.count }}
                </ElTag>
                <ElTag effect="plain" round>功能包展开 {{ derivedActions.length }}</ElTag>
                <ElTag type="success" effect="plain" round
                  >已保留 {{ expandedSelectedActionIds.length }}</ElTag
                >
                <ElTag type="danger" effect="plain" round>已关闭 {{ disabledActionCount }}</ElTag>
                <ElTag type="info" effect="plain" round>总计 {{ actionLeafCount }}</ElTag>
              </div>

              <div class="toolbar-filters">
                <ElInput
                  v-model="actionKeyword"
                  clearable
                  placeholder="搜索功能标题或资源动作"
                  :prefix-icon="Search"
                  class="toolbar-search"
                />
              </div>
            </div>

            <div class="cascader-card">
              <ElCascaderPanel
                ref="actionPanelRef"
                v-model="selectedActionNodeValues"
                :options="filteredActionOptions"
                :props="actionCascaderProps"
                class="permission-cascader"
              >
                <template #default="{ node, data }">
                  <div class="panel-node" :class="{ 'is-leaf': node.isLeaf }">
                    <div class="panel-node__main">
                      <span class="panel-node__label">{{ data.label }}</span>
                      <template v-if="node.isLeaf">
                        <span
                          v-if="data.permissionText"
                          class="panel-node__meta panel-node__meta--truncate"
                          :title="data.permissionText"
                        >
                          {{ data.permissionText }}
                        </span>
                      </template>
                    </div>

                    <ElTag
                      v-if="!node.isLeaf"
                      size="small"
                      effect="plain"
                      round
                      type="info"
                      class="panel-node__count"
                    >
                      {{ data.children?.length || 0 }}
                    </ElTag>
                  </div>
                </template>
              </ElCascaderPanel>
            </div>

            <div class="cascader-footer">
              <span class="footer-note"
                >角色功能权限按勾选保存，勾选即允许，不勾选默认不允许。</span
              >
              <ElButton text @click="clearActionSelection">清空选择</ElButton>
            </div>
          </div>
        </ElTabPane>

        <ElTabPane label="数据权限" name="data">
          <div class="data-panel">
            <div class="data-summary">
              <ElTag effect="plain" round>资源 {{ dataRows.length }}</ElTag>
              <ElTag type="warning" effect="plain" round>已配置 {{ configuredDataCount }}</ElTag>
            </div>

            <ElTable :data="pagedDataRows" border class="data-table">
              <ElTableColumn prop="resourceName" label="资源" min-width="220" />
              <ElTableColumn label="数据范围" min-width="260">
                <template #default="{ row }">
                  <ElSelect
                    v-model="row.selectedDataScope"
                    placeholder="选择数据范围"
                    clearable
                    style="width: 100%"
                  >
                    <ElOption
                      v-for="scope in dataScopeOptions"
                      :key="scope.scopeCode"
                      :label="scope.scopeName"
                      :value="scope.scopeCode"
                    />
                  </ElSelect>
                </template>
              </ElTableColumn>
            </ElTable>
            <WorkspacePagination
              v-if="dataRows.length > 0"
              v-model:current-page="dataPagination.current"
              v-model:page-size="dataPagination.size"
              :total="dataRows.length"
              compact
            />
          </div>
        </ElTabPane>
      </ElTabs>
    </div>

    <template #footer>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue'
  import { Search } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import type { CascaderOption, CascaderProps } from 'element-plus'
  import { useRouter } from 'vue-router'
  import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import {
    fetchGetMenuTreeAll,
    fetchGetRoleActions,
    fetchGetRoleDataPermissions,
    fetchGetRolePackages,
    fetchGetRoleMenus,
    fetchSetRoleActions,
    fetchSetRoleDataPermissions,
    fetchSetRoleMenus
  } from '@/api/system-manage'

  interface RoleMenuNode {
    id: string
    label: string
    path?: string
    isHide?: boolean
    isIframe?: boolean
    isEnable?: boolean
    children?: RoleMenuNode[]
  }

  interface MenuOption extends CascaderOption {
    path?: string
    isHide?: boolean
    isIframe?: boolean
    isEnable?: boolean
    totalLeafCount?: number
    availableLeafCount?: number
    selectedLeafCount?: number
  }

  interface ActionOption extends CascaderOption {
    permissionText?: string
  }

  interface DataRow {
    resourceCode: string
    resourceName: string
    selectedDataScope: string
  }

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
    appKey?: string
  }

  const props = defineProps<Props>()
  const router = useRouter()

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const saving = ref(false)
  const activeTab = ref<'menus' | 'actions' | 'data'>('menus')

  const menuPanelRef = ref<any>()
  const actionPanelRef = ref<any>()

  const menuKeyword = ref('')
  const showHiddenMenus = ref(true)
  const showIframeMenus = ref(true)
  const showEnabledMenus = ref(true)
  const showMenuPath = ref(false)
  const actionKeyword = ref('')

  const menuTreeData = ref<RoleMenuNode[]>([])
  const selectedMenuNodeValues = ref<string[]>([])
  const roleMenuBoundary = ref<Api.SystemManage.RoleMenuBoundaryResponse | null>(null)
  const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])

  const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const selectedActionNodeValues = ref<string[]>([])
  const roleActionBoundary = ref<Api.SystemManage.RoleActionBoundaryResponse | null>(null)
  const selectedActionSourcePackageId = ref('')

  const dataRows = ref<DataRow[]>([])
  const dataPagination = ref({
    current: 1,
    size: 10
  })
  const dataScopeOptions = ref<Api.SystemManage.RoleDataPermissionScopeOption[]>([])
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())

  const roleTitle = computed(() => props.roleData?.roleName || '')
  const summaryItems = computed(() => [
    { label: '角色', value: roleTitle.value || '-' },
    { label: '功能包', value: featurePackages.value.length, type: 'success' as const },
    { label: '能力已关闭', value: disabledActionCount.value, type: 'danger' as const }
  ])

  const menuOptions = computed<MenuOption[]>(() => normalizeMenuOptions(menuTreeData.value))
  const selectedMenuIdSet = computed(() => new Set(expandedSelectedMenuIds.value))
  const availableMenuIdSet = computed(
    () => new Set((roleMenuBoundary.value?.available_menu_ids || []).map((item) => `${item}`))
  )

  const menuBranchMap = computed(() => {
    const map = new Map<string, string[]>()

    const visit = (nodes: MenuOption[]): string[] => {
      const collected: string[] = []
      nodes.forEach((node) => {
        const childIds = visit((node.children || []) as MenuOption[])
        const current = [`${node.value}`]
        const all = current.concat(childIds)
        map.set(`${node.value}`, all)
        collected.push(...all)
      })
      return collected
    }

    visit(menuOptions.value)
    return map
  })

  const expandedSelectedMenuIds = computed(() =>
    expandSelectedValues(selectedMenuNodeValues.value, menuBranchMap.value)
  )

  const availableActionIdSet = computed(
    () => new Set((roleActionBoundary.value?.available_action_ids || []).map((item) => `${item}`))
  )
  const disabledActionCount = computed(
    () => roleActionBoundary.value?.disabled_action_ids?.length || 0
  )
  const disabledActionIdSet = computed(
    () => new Set((roleActionBoundary.value?.disabled_action_ids || []).map((item) => `${item}`))
  )
  const actionDerivedSourceMap = computed(() =>
    Object.fromEntries(
      (roleActionBoundary.value?.derived_sources || []).map((item) => [
        item.action_id,
        item.package_ids
      ])
    )
  )
  const actionSourceMapText = computed(() =>
    Object.fromEntries(
      Object.entries(actionDerivedSourceMap.value).map(([key, value]) => [
        key,
        value.map((item) => `${item}`)
      ])
    )
  )

  const filteredPermissionActions = computed(() => {
    const allActions = permissionActions.value || []
    if (!allActions.length) return []
    // 兜底：部分历史角色或快照延迟场景下 available_action_ids 可能为空，
    // 此时不应把功能权限面板清空，回退到服务端返回的 actions 列表。
    if (availableActionIdSet.value.size === 0) return allActions
    return allActions.filter((item) => availableActionIdSet.value.has(item.id))
  })
  const derivedActions = computed(() => filteredPermissionActions.value)
  const disabledActions = computed(() =>
    derivedActions.value.filter((item) => disabledActionIdSet.value.has(item.id))
  )
  const derivedActionItems = computed(() =>
    derivedActions.value.map((item) => ({ id: item.id, label: item.name }))
  )
  const disabledActionItems = computed(() =>
    disabledActions.value.map((item) => ({ id: item.id, label: item.name }))
  )

  const actionOptions = computed<ActionOption[]>(() => {
    const featureMap = new Map<string, ActionOption>()

    filteredPermissionActions.value.forEach((action) => {
      const featureKey = `${action.featureGroupId || action.featureKind || 'business'}`
      const moduleKey = `${action.moduleGroupId || action.moduleCode || action.resourceCode || 'default'}`
      const featureLabel =
        action.featureGroup?.name || formatFeature(`${action.featureKind || 'business'}`)
      const moduleLabel =
        action.moduleGroup?.name || action.moduleCode || action.resourceCode || '未分类模块'

      if (!featureMap.has(featureKey)) {
        featureMap.set(featureKey, {
          value: `feature:${featureKey}`,
          label: featureLabel,
          children: []
        })
      }

      const feature = featureMap.get(featureKey)!
      const modules = (feature.children || []) as ActionOption[]
      const moduleValue = `module:${featureKey}:${moduleKey}`
      let module = modules.find((item) => item.value === moduleValue)

      if (!module) {
        module = {
          value: moduleValue,
          label: moduleLabel,
          children: []
        }
        modules.push(module)
        feature.children = modules
      }

      const leaves = (module.children || []) as ActionOption[]
      leaves.push({
        value: action.id,
        label: action.name,
        leaf: true,
        permissionText: action.permissionKey || `${action.resourceCode}:${action.actionCode}`
      })
      module.children = leaves
    })

    return Array.from(featureMap.values())
  })

  const actionBranchMap = computed(() => {
    const map = new Map<string, string[]>()
    actionOptions.value.forEach((feature) => {
      const featureLeaves: string[] = []
      ;((feature.children || []) as ActionOption[]).forEach((module) => {
        const moduleLeaves = ((module.children || []) as ActionOption[]).map(
          (leaf) => `${leaf.value}`
        )
        map.set(`${module.value}`, moduleLeaves)
        featureLeaves.push(...moduleLeaves)
      })
      map.set(`${feature.value}`, featureLeaves)
    })
    return map
  })

  const expandedSelectedActionIds = computed(() =>
    expandSelectedValues(selectedActionNodeValues.value, actionBranchMap.value)
  )

  const actionLeafCount = computed(() => filteredPermissionActions.value.length)

  const topLevelActionTags = computed(() => {
    const counter = new Map<string, number>()
    expandedSelectedActionIds.value.forEach((actionId) => {
      const action = filteredPermissionActions.value.find((item) => item.id === actionId)
      const featureLabel = `${action?.featureGroup?.name || action?.featureKind || 'business'}`
      counter.set(featureLabel, (counter.get(featureLabel) || 0) + 1)
    })
    return Array.from(counter.entries()).map(([key, count]) => ({
      label: formatFeature(key),
      count
    }))
  })

  const configuredDataCount = computed(
    () => dataRows.value.filter((item) => item.selectedDataScope).length
  )
  const pagedDataRows = computed(() => {
    const start = (dataPagination.value.current - 1) * dataPagination.value.size
    return dataRows.value.slice(start, start + dataPagination.value.size)
  })

  const menuCascaderProps: CascaderProps = {
    multiple: true,
    emitPath: false,
    checkStrictly: false,
    expandTrigger: 'click',
    checkOnClickNode: false,
    checkOnClickLeaf: false,
    showPrefix: true
  }

  const actionCascaderProps: CascaderProps = {
    multiple: true,
    emitPath: false,
    checkStrictly: false,
    expandTrigger: 'click',
    checkOnClickNode: false,
    checkOnClickLeaf: false,
    showPrefix: true
  }

  const filteredMenuOptions = computed(() => {
    const keyword = menuKeyword.value.trim().toLowerCase()
    return filterNestedOptions(menuOptions.value, (node) => {
      if (!node.leaf) {
        return (node.availableLeafCount || 0) > 0 && !keyword
      }
      if (!availableMenuIdSet.value.has(`${node.value || ''}`)) return false
      if (!showHiddenMenus.value && node.isHide) return false
      if (!showIframeMenus.value && node.isIframe) return false
      if (!showEnabledMenus.value && node.isEnable !== false) return false
      if (keyword && !`${node.label || ''} ${node.path || ''}`.toLowerCase().includes(keyword))
        return false
      return true
    })
  })
  const menuNodeCount = computed(() => flattenMenuOptionIds(filteredMenuOptions.value).length)

  const filteredActionOptions = computed(() => {
    const keyword = actionKeyword.value.trim().toLowerCase()

    return filterNestedOptions(actionOptions.value, (node) => {
      if (!node.leaf) return !keyword
      const text = [node.label, node.permissionText].filter(Boolean).join(' ').toLowerCase()
      if (keyword && !text.includes(keyword)) return false
      return true
    })
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (open) loadData()
    }
  )

  watch(
    [menuKeyword, showHiddenMenus, showIframeMenus, showEnabledMenus, menuTreeData],
    async () => {
      await nextTick()
      ensureExpandedMenus(menuPanelRef.value, selectedMenuNodeValues.value)
    }
  )

  watch(
    () => filteredActionOptions.value,
    async () => {
      await nextTick()
      ensureExpandedMenus(actionPanelRef.value, selectedActionNodeValues.value)
    },
    { deep: true }
  )

  async function loadData() {
    if (!props.roleData?.roleId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    loading.value = true
    activeTab.value = 'menus'
    menuKeyword.value = ''
    actionKeyword.value = ''
    selectedActionSourcePackageId.value = ''

    try {
      const [menuTree, roleMenus, rolePackages, roleActions, dataPermissionRes] = await Promise.all(
        [
          fetchGetMenuTreeAll(undefined, currentAppKey.value),
          fetchGetRoleMenus(props.roleData.roleId, currentAppKey.value),
          fetchGetRolePackages(props.roleData.roleId, currentAppKey.value),
          fetchGetRoleActions(props.roleData.roleId, currentAppKey.value),
          fetchGetRoleDataPermissions(props.roleData.roleId)
        ]
      )

      menuTreeData.value = Array.isArray(menuTree) ? normalizeMenus(menuTree) : []
      roleMenuBoundary.value = roleMenus || null
      featurePackages.value = rolePackages?.packages || []
      selectedMenuNodeValues.value = (roleMenus?.menu_ids || []).map((item: any) => `${item}`)

      permissionActions.value = roleActions?.actions || []
      roleActionBoundary.value = roleActions || null
      const availableActionIDSet = availableActionIdSet.value
      selectedActionNodeValues.value = (roleActions?.action_ids || []).filter((item: any) => {
        if (!item) return false
        if (availableActionIDSet.size === 0) return true
        return availableActionIDSet.has(item)
      })

      dataScopeOptions.value = (dataPermissionRes?.available_data_scopes || []).map((item: any) => ({
        scopeCode: item.data_scope,
        scopeName: item.label
      }))
      const selectedScopeMap = new Map<string, string>()
      ;(dataPermissionRes?.permissions || []).forEach((item: any) => {
        selectedScopeMap.set(item.resource_code, item.data_scope)
      })
      dataRows.value = (dataPermissionRes?.resources || []).map((item: any) => ({
        resourceCode: item.resource_code,
        resourceName: item.resource_name,
        selectedDataScope: selectedScopeMap.get(item.resource_code) || ''
      }))
      dataPagination.value.current = 1

      await nextTick()
      ensureExpandedMenus(menuPanelRef.value, selectedMenuNodeValues.value)
      ensureExpandedMenus(actionPanelRef.value, selectedActionNodeValues.value)
    } catch (error: any) {
      ElMessage.error(error?.message || '加载角色权限失败')
    } finally {
      loading.value = false
    }
  }

  function normalizeMenus(items: any[]): RoleMenuNode[] {
    return items.map((item) => ({
      id: `${item.id}`,
      label: item.meta?.title || item.name || item.path || '未命名菜单',
      path: item.path || '',
      isHide: Boolean(item.meta?.isHide),
      isIframe: Boolean(item.meta?.isIframe),
      isEnable: item.meta?.isEnable !== false,
      children: Array.isArray(item.children) ? normalizeMenus(item.children) : []
    }))
  }

  function normalizeMenuOptions(items: RoleMenuNode[]): MenuOption[] {
    return items.map((item) => {
      const children = normalizeMenuOptions(item.children || [])
      const isLeaf = !item.children?.length
      const isAvailable = availableMenuIdSet.value.has(item.id)
      const availableLeafCount = isLeaf
        ? isAvailable
          ? 1
          : 0
        : children.reduce((sum, child) => sum + (child.availableLeafCount || 0), 0)

      return {
        value: item.id,
        label: item.label,
        path: item.path || '',
        isHide: item.isHide,
        isIframe: item.isIframe,
        isEnable: item.isEnable,
        leaf: isLeaf,
        disabled: availableLeafCount === 0,
        totalLeafCount: countMenuLeaves(item),
        availableLeafCount,
        selectedLeafCount: countSelectedMenuLeaves(item, selectedMenuIdSet.value),
        children
      }
    })
  }

  function flattenMenuOptionIds(items: MenuOption[]): string[] {
    return items.flatMap((item) => [
      `${item.value}`,
      ...flattenMenuOptionIds((item.children || []) as MenuOption[])
    ])
  }

  function countMenuLeaves(node: RoleMenuNode): number {
    if (!node.children?.length) return 1
    return node.children.reduce((sum, child) => sum + countMenuLeaves(child), 0)
  }

  function countSelectedMenuLeaves(node: RoleMenuNode, selectedSet: Set<string>): number {
    if (!node.children?.length) return selectedSet.has(node.id) ? 1 : 0
    return node.children.reduce(
      (sum, child) => sum + countSelectedMenuLeaves(child, selectedSet),
      0
    )
  }

  function filterNestedOptions<T extends CascaderOption>(
    nodes: T[],
    matcher: (node: T) => boolean
  ): T[] {
    return nodes
      .map((node) => {
        const children = Array.isArray(node.children)
          ? filterNestedOptions(node.children as T[], matcher)
          : []
        const selfMatched = matcher(node)
        if (selfMatched || children.length > 0) {
          return {
            ...node,
            children
          }
        }
        return null
      })
      .filter(Boolean) as T[]
  }

  function expandSelectedValues(values: string[], branchMap: Map<string, string[]>) {
    const set = new Set<string>()
    values.forEach((value) => {
      const key = `${value || ''}`
      const mapped = branchMap.get(key)
      if (mapped?.length) {
        mapped.forEach((item) => set.add(item))
        return
      }
      set.add(key)
    })
    return Array.from(set)
  }

  function ensureExpandedMenus(panel: any, selectedValues: string[]) {
    const rootMenus = panel?.menus?.[0]
    if (!panel || !rootMenus?.length) return

    const firstValue = selectedValues[0]
    let featureNode = rootMenus[0]
    let moduleNode = featureNode?.children?.[0]

    if (firstValue) {
      const matchedNode = panel
        .getFlattedNodes?.(false)
        ?.find((node: any) => `${node?.value}` === `${firstValue}`)
      const pathNodes = matchedNode?.pathNodes || []
      if (pathNodes[0]) featureNode = pathNodes[0]
      if (pathNodes[1]) moduleNode = pathNodes[1]
    }

    const nextMenus = [rootMenus]
    if (featureNode?.children?.length) nextMenus.push(featureNode.children)
    if (moduleNode?.children?.length) nextMenus.push(moduleNode.children)
    panel.menus = nextMenus
  }

  function clearMenuSelection() {
    selectedMenuNodeValues.value = []
  }

  function clearActionSelection() {
    selectedActionNodeValues.value = []
  }

  function formatFeature(value: string) {
    if (value === 'system') return '系统功能'
    if (value === 'business') return '业务功能'
    return value
  }

  function handleCancel() {
    visible.value = false
  }

  watch(
    () => dataPagination.value.size,
    () => {
      dataPagination.value.current = 1
    }
  )

  function goToFeaturePackagePage(
    item: Api.SystemManage.FeaturePackageItem,
    open?: 'menus' | 'actions'
  ) {
    if (!currentAppKey.value) return
    router.push({
      name: 'FeaturePackage',
      query: {
        packageKey: item.packageKey,
        workspaceScope: item.workspaceScope || 'all',
        ...(open ? { open } : {})
      }
    })
  }

  async function handleSave() {
    if (!props.roleData?.roleId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    saving.value = true
    try {
      const dataPayload = dataRows.value
        .filter((item) => item.selectedDataScope)
        .map((item) => ({
          resource_code: item.resourceCode,
          data_scope: item.selectedDataScope
        }))

      await Promise.all([
        fetchSetRoleMenus(
          props.roleData.roleId,
          expandedSelectedMenuIds.value,
          currentAppKey.value
        ),
        fetchSetRoleActions(
          props.roleData.roleId,
          expandedSelectedActionIds.value,
          currentAppKey.value
        ),
        fetchSetRoleDataPermissions(props.roleData.roleId, dataPayload)
      ])

      ElMessage.success('角色权限已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存角色权限失败')
    } finally {
      saving.value = false
    }
  }
</script>

<style scoped lang="scss">
  .dialog-shell {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .dialog-note {
    color: #6b7280;
    line-height: 1.6;
  }

  .package-tags {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .permission-tabs :deep(.el-tabs__content) {
    padding-top: 12px;
  }

  .cascader-pane,
  .data-panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .cascader-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .cascader-tags,
  .toolbar-filters,
  .data-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .toolbar-filters {
    flex: 0 1 auto;
  }

  .toolbar-switch {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: #6b7280;
    font-size: 13px;
    line-height: 1;
    white-space: nowrap;
  }

  .toolbar-switch__label {
    color: inherit;
  }

  .toolbar-search {
    width: 220px;
  }

  .toolbar-scope {
    width: 140px;
  }

  .cascader-card {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 16px;
    overflow: hidden;
    background: #fff;
  }

  .permission-cascader {
    width: 100%;
    height: 420px;
  }

  .permission-cascader :deep(.el-cascader-panel) {
    height: 100%;
    display: flex;
    align-items: stretch;
  }

  .permission-cascader :deep(.el-cascader-menu) {
    width: auto;
    flex: 0 1 auto;
    min-width: 320px;
  }

  .permission-cascader :deep(.el-cascader-menu:nth-child(2)) {
    width: auto;
    flex: 0 1 auto;
    min-width: 240px;
  }

  .permission-cascader :deep(.el-cascader-menu:last-child) {
    flex: 1 1 auto;
    width: auto;
  }

  .permission-cascader :deep(.el-cascader-menu__wrap),
  .permission-cascader :deep(.el-scrollbar),
  .permission-cascader :deep(.el-scrollbar__wrap),
  .permission-cascader :deep(.el-scrollbar__view) {
    height: 100%;
  }

  .panel-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    min-width: 0;
  }

  .panel-node__main {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1 1 auto;
  }

  .panel-node__label {
    color: #111827;
    font-size: 14px;
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .panel-node__meta {
    color: #6b7280;
    font-size: 12px;
    line-height: 1.4;
    flex: 0 1 auto;
  }

  .panel-node__meta--truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .panel-node__tags {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 0 0 auto;
  }

  .panel-node__count {
    flex: 0 0 auto;
  }

  .cascader-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .footer-note {
    color: #6b7280;
    font-size: 13px;
    line-height: 1.5;
  }

  .data-table {
    border-radius: 16px;
    overflow: hidden;
  }

  @media (max-width: 960px) {
    .cascader-toolbar {
      align-items: stretch;
    }

    .toolbar-search,
    .toolbar-scope {
      width: 100%;
    }

    .permission-cascader :deep(.el-cascader-panel) {
      flex-direction: column;
    }

    .permission-cascader :deep(.el-cascader-menu),
    .permission-cascader :deep(.el-cascader-menu:nth-child(2)),
    .permission-cascader :deep(.el-cascader-menu:last-child) {
      width: 100%;
      flex: 1 1 auto;
    }

    .panel-node {
      align-items: flex-start;
    }

    .panel-node__main {
      flex-wrap: wrap;
    }
  }
</style>
