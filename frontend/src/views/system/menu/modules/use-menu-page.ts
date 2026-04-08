/**
 * menu 视图主 composable。
 *
 * 将原 index.vue 的 1280+ 行 script 整体抽离，
 * 视图层只保留 template + 极少量调用胶水。
 *
 * 设计原则：
 * - 一次性拉出全部 reactive state、handler、watch 与 lifecycle，避免拆得过细造成跨文件依赖循环；
 * - 纯函数 helper 已抽到 ./helpers.ts；
 * - 返回值即视图模板使用的所有变量与方法。
 */
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatMenuTitle } from '@/utils/router'
import { getMenuActionRequirement } from '@/utils/permission/menu'
import { useTableColumns } from '@/hooks/core/useTableColumns'
import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
import type { AppRouteRecord } from '@/types/router'
import {
  fetchGetMenuTreeAll,
  fetchCreateMenu,
  fetchUpdateMenu,
  fetchDeleteMenu,
  fetchGetMenuDeletePreview,
  fetchGetMenuManageGroups,
  fetchCreateMenuManageGroup,
  fetchUpdateMenuManageGroup,
  fetchDeleteMenuManageGroup,
  fetchGetPageOptions,
  fetchGetMenuSpaces,
  fetchGetApps,
  fetchCreateMenuBackup,
  fetchGetMenuBackupList,
  fetchDeleteMenuBackup,
  fetchRestoreMenuBackup
} from '@/api/system-manage'
import {
  buildManageGroupNode,
  cloneMenuNode,
  getAccessModeLabel,
  getAccessModeTag,
  getManageGroupId,
  getMenuTypeTag,
  getMenuTypeText,
  isDirectoryMenuRow,
  isEntryMenuRow,
  isManageGroupRow,
  normalizeKeyword
} from './helpers'

type MenuBackupScopeType = 'space' | 'global'
type MenuDeleteMode = 'single' | 'cascade' | 'promote_children'

type MenuDeleteParentOption = {
  label: string
  value: string
  children?: MenuDeleteParentOption[]
}

export function useMenuPage() {
  // --- 状态管理 ---
  const loading = ref(false)
  const loadError = ref('')
  const showSearchBar = ref(false)
  const isExpanded = ref(false)
  const showHiddenMenus = ref(true)
  const showIframeMenus = ref(true)
  const showEnabledMenus = ref(true)
  const groupingEnabled = ref(true)
  const groupedMenuVisible = ref(true)
  const tableRef = ref()
  const multiSelectEnabled = ref(false)
  const rawMenuTree = ref<AppRouteRecord[]>([])
  const rawPages = ref<Api.SystemManage.PageItem[]>([])
  const menuSpaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const activeSpaceKey = ref('')
  const route = useRoute()
  const router = useRouter()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const managedAppMissingText = '请选择当前要管理的 App'
  const isLayoutMode = computed(() => `${route.query.layout || ''}`.trim() === '1')
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const menuGroups = ref<Api.SystemManage.MenuManageGroupItem[]>([])
  const dataFromBackend = ref(false)
  const menuGroupApiUnavailable = ref(false)

  // --- 菜单备份相关状态 ---
  const backupLoading = ref(false)
  const backupDialogVisible = ref(false)
  const backupListDialogVisible = ref(false)
  const backupScopeType = ref<MenuBackupScopeType>('space')
  const backupList = ref<Api.SystemManage.MenuBackupItem[]>([])
  const warnDev = (...args: any[]) => {
    if (import.meta.env.DEV) {
      console.info(...args)
    }
  }

  // --- 搜索相关 ---
  const initialSearchState = { name: '', route: '' }
  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })
  // --- 弹窗相关 ---
  const dialogVisible = ref(false)
  const manageGroupDrawerVisible = ref(false)
  const groupSaving = ref(false)
  const editData = ref<any>(null)
  const parentRowForAdd = ref<AppRouteRecord | null>(null)
  const deleteDialogVisible = ref(false)
  const deleteLoading = ref(false)
  const deleteTargetRow = ref<any>(null)
  const deletePreview = ref<Api.SystemManage.MenuDeletePreviewItem | null>(null)
  const actionRequirementVisible = ref(false)
  const actionRequirementData = ref<any>(null)
  const selectedMenuRows = ref<any[]>([])
  const batchTargetGroupId = ref('')
  const batchAssignDialogVisible = ref(false)

  // --- 菜单列表处理 ---
  const matchesMenuFilters = (item: AppRouteRecord) => {
    if (!showHiddenMenus.value && item.meta?.isHide) return false
    if (!showIframeMenus.value && item.meta?.isIframe) return false
    if (!showEnabledMenus.value && item.meta?.isEnable !== false) return false
    if (!groupedMenuVisible.value && getManageGroupId(item)) return false
    return true
  }

  const matchesMenuSearch = (item: AppRouteRecord) => {
    const searchName = normalizeKeyword(appliedFilters.name)
    const searchRoute = normalizeKeyword(appliedFilters.route)
    const title = normalizeKeyword(formatMenuTitle(item.meta?.title))
    const path = normalizeKeyword(item.path)
    const titleMatch = !searchName || title.includes(searchName)
    const routeMatch = !searchRoute || path.includes(searchRoute)
    return titleMatch && routeMatch
  }

  const menuGroupMap = computed(() => new Map(menuGroups.value.map((item) => [item.id, item])))

  const menuSpaceMap = computed(
    () => new Map(menuSpaces.value.map((item) => [item.spaceKey, item]))
  )
  const menuSpaceOptions = computed(() =>
    menuSpaces.value.map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )
  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )

  const linkedPagesByMenuId = computed(() => {
    const map = new Map<string, Api.SystemManage.PageItem[]>()
    rawPages.value.forEach((item) => {
      const menuId = `${item.parentMenuId || ''}`.trim()
      if (!menuId) return
      const list = map.get(menuId) || []
      list.push(item)
      map.set(menuId, list)
    })
    map.forEach((list, key) => {
      list.sort((left, right) => {
        const sortDiff = Number(left.sortOrder || 0) - Number(right.sortOrder || 0)
        if (sortDiff !== 0) return sortDiff
        return `${left.name || ''}${left.pageKey || ''}`.localeCompare(
          `${right.name || ''}${right.pageKey || ''}`,
          'zh-Hans-CN'
        )
      })
      map.set(key, list)
    })
    return map
  })

  const getLinkedPages = (item: any) =>
    linkedPagesByMenuId.value.get(String(item?.id || '')) || []

  const getSpaceName = (spaceKey?: string) => {
    const normalized = `${spaceKey || ''}`.trim()
    if (!normalized) return '未选择空间'
    return menuSpaceMap.value.get(normalized)?.name || normalized
  }

  const currentSpaceName = computed(() => getSpaceName(activeSpaceKey.value))
  const menuPageTitle = computed(() => (isLayoutMode.value ? '空间布局' : '菜单定义管理'))
  const menuPageDescription = computed(() =>
    isLayoutMode.value
      ? '维护当前 App 在不同空间下的菜单树摆放、排序与可见性；菜单定义本体仍由定义管理页统一维护。'
      : '统一管理当前 App 的目录、入口路由与外链定义；空间差异化布局与 Host 绑定统一放到高级空间配置里。'
  )
  const menuToolbarTip = computed(() =>
    isLayoutMode.value
      ? '当前按空间查看菜单布局树；同一菜单定义在多个空间复用时，共享一份授权与裁剪状态。'
      : '当前按 App 维护菜单定义；父级、排序和显示位置以当前查看空间做布局参考，不再把空间当成菜单主归属。'
  )
  const backupDialogTitle = computed(() =>
    isLayoutMode.value ? '备份当前空间布局' : '备份菜单定义'
  )
  const backupAlertTitle = computed(() =>
    isLayoutMode.value
      ? `当前将备份空间布局：${currentSpaceName.value}`
      : `当前将创建 App 级定义备份：${targetAppKey.value}`
  )
  const backupAlertDescription = computed(() =>
    isLayoutMode.value
      ? '该备份只保存当前 App 下当前空间的布局树和相关菜单分组，用于后续覆盖恢复当前空间。'
      : '该备份只保存当前 App 的菜单定义集合，不含各空间的父级、排序和显隐差异；空间级恢复请到高级空间配置页处理。'
  )
  const backupListTitle = computed(() => (isLayoutMode.value ? '管理空间布局备份' : '管理定义备份'))
  const backupListAlertDescription = computed(() =>
    isLayoutMode.value
      ? '这里只展示当前 App 下当前空间的布局备份，不包含 App 级菜单定义备份。'
      : '这里只展示当前 App 的定义备份；空间级布局备份请到空间高级配置页管理。'
  )
  const backupListEmptyDescription = computed(() =>
    isLayoutMode.value ? '当前空间暂无布局备份' : '当前 App 暂无定义备份'
  )

  const getBackupScopeLabel = (item: Api.SystemManage.MenuBackupItem) => {
    if (`${item.scope_type || ''}`.trim() === 'global') {
      return isLayoutMode.value ? '全空间备份' : '定义备份'
    }
    return getSpaceName(item.space_key || activeSpaceKey.value)
  }

  const injectManageGroups = (
    items: AppRouteRecord[],
    parentKey = 'root',
    inheritedGroupId = ''
  ): AppRouteRecord[] => {
    const result: AppRouteRecord[] = []
    const groupNodeMap = new Map<string, AppRouteRecord>()

    items.forEach((item) => {
      const currentGroupID = getManageGroupId(item)
      const children = item.children?.length
        ? injectManageGroups(
            item.children as AppRouteRecord[],
            `${item.id || item.path || parentKey}`,
            currentGroupID || inheritedGroupId
          )
        : []
      const cloned = cloneMenuNode(item, children)
      const groupID = getManageGroupId(cloned)
      const group = groupID ? menuGroupMap.value.get(groupID) : undefined
      if (!group || groupID === inheritedGroupId) {
        result.push(cloned)
        return
      }

      let groupNode = groupNodeMap.get(group.id)
      if (!groupNode) {
        groupNode = buildManageGroupNode(group, parentKey)
        groupNodeMap.set(group.id, groupNode)
        result.push(groupNode)
      }
      const groupedChildren = (groupNode.children || []) as AppRouteRecord[]
      groupedChildren.push(cloned)
      groupNode.children = groupedChildren
    })

    result.sort((a, b) => Number(a.sort_order ?? 0) - Number(b.sort_order ?? 0))
    return result
  }

  const filterMenuTree = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items.reduce<AppRouteRecord[]>((result, item) => {
      if (!matchesMenuFilters(item)) {
        return result
      }

      const children = item.children?.length
        ? filterMenuTree(item.children as AppRouteRecord[])
        : []
      if (!matchesMenuSearch(item) && children.length === 0) {
        return result
      }

      result.push(cloneMenuNode(item, children))
      return result
    }, [])
  }

  const filteredMenuTree = computed(() => filterMenuTree(rawMenuTree.value))

  const groupingAvailable = computed(() => !menuGroupApiUnavailable.value)

  const tableData = computed(() =>
    groupingAvailable.value && groupingEnabled.value
      ? injectManageGroups(filteredMenuTree.value)
      : filteredMenuTree.value
  )

  const menuStats = computed(() => {
    const stats = {
      total: 0,
      directory: 0,
      entry: 0,
      external: 0,
      groups: menuGroups.value.length
    }
    const walk = (items: AppRouteRecord[]) => {
      items.forEach((item) => {
        stats.total += 1
        const kind = `${(item as any).kind || ''}`.trim()
        if (kind === 'external') {
          stats.external += 1
        } else if (kind === 'entry') {
          stats.entry += 1
        } else {
          stats.directory += 1
        }
        if (item.children?.length) {
          walk(item.children as AppRouteRecord[])
        }
      })
    }
    walk(filteredMenuTree.value)
    return stats
  })

  const menuHeroMetrics = computed(() => [
    { label: '当前 App', value: targetAppKey.value },
    { label: isLayoutMode.value ? '布局空间' : '参考空间', value: currentSpaceName.value },
    { label: '总数', value: menuStats.value.total },
    { label: '目录', value: menuStats.value.directory },
    { label: '入口', value: menuStats.value.entry },
    { label: '外链', value: menuStats.value.external },
    { label: '分组', value: menuStats.value.groups }
  ])

  const getMenuList = async () => {
    loading.value = true
    loadError.value = ''
    dataFromBackend.value = false
    if (!targetAppKey.value) {
      rawMenuTree.value = []
      rawPages.value = []
      menuGroups.value = []
      activeSpaceKey.value = ''
      loadError.value = managedAppMissingText
      loading.value = false
      return
    }
    if (isLayoutMode.value && !activeSpaceKey.value) {
      rawMenuTree.value = []
      rawPages.value = []
      loadError.value = '请选择当前要配置的菜单空间'
      loading.value = false
      return
    }
    try {
      const [list, pagesResult, groupsResult] = await Promise.all([
        fetchGetMenuTreeAll(activeSpaceKey.value, targetAppKey.value),
        fetchGetPageOptions(activeSpaceKey.value, targetAppKey.value).then(
          (res) => res.records || []
        ),
        fetchGetMenuManageGroups()
          .then((groups) => ({ ok: true as const, groups }))
          .catch((error) => ({
            ok: false as const,
            error,
            groups: [] as Api.SystemManage.MenuManageGroupItem[]
          }))
      ])
      rawMenuTree.value = Array.isArray(list) ? list : []
      rawPages.value = Array.isArray(pagesResult) ? pagesResult : []
      menuGroups.value = groupsResult.groups || []
      dataFromBackend.value = true

      if (!groupsResult.ok) {
        warnDev('[Menus] 菜单分组接口不可用，已降级为无分组模式', groupsResult.error)
        menuGroupApiUnavailable.value = true
      } else {
        menuGroupApiUnavailable.value = false
      }
    } catch (error) {
      warnDev('[Menus] 获取菜单数据失败，已回退为空列表', error)
      rawMenuTree.value = []
      rawPages.value = []
      menuGroups.value = []
      menuGroupApiUnavailable.value = false
      loadError.value = '菜单数据暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  // --- 表格列配置 ---
  const { columnChecks, columns: displayColumns } = useTableColumns(() => [
    {
      type: 'selection',
      width: 52,
      align: 'center',
      selectable: (row: any) => !isManageGroupRow(row),
      className: 'menu-selection-column',
      labelClassName: 'menu-selection-column'
    } as any,
    { prop: 'title', label: '菜单名称', minWidth: 200, useSlot: true, slotName: 'title' },
    { prop: 'sort_order', label: '排序', width: 80, align: 'center' },
    { prop: 'type', label: '类型', width: 100, align: 'center', useSlot: true, slotName: 'type' },
    { prop: 'path', label: '路由', minWidth: 150, useSlot: true, slotName: 'path' },
    { prop: 'component', label: '组件路径', minWidth: 200, useSlot: true, slotName: 'component' },
    {
      prop: 'appKey',
      label: 'App',
      width: 150,
      align: 'center',
      formatter: (row: any) => row.appKey || targetAppKey.value
    },
    {
      prop: 'space',
      label: '菜单空间',
      width: 120,
      align: 'center',
      useSlot: true,
      slotName: 'space'
    },
    { prop: 'linkedPage', label: '受管页面', minWidth: 220, useSlot: true, slotName: 'linkedPage' },
    {
      prop: 'advanced',
      label: '高级配置',
      minWidth: 200,
      align: 'center',
      useSlot: true,
      slotName: 'advanced'
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      align: 'center',
      useSlot: true,
      slotName: 'status'
    },
    {
      prop: 'operation',
      label: '操作',
      width: 120,
      align: 'center',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  // --- 辅助方法 ---
  const getMenuActionRequirementLabel = (row: any) => {
    const requirement = getMenuActionRequirement(row.meta)
    if (!requirement.actions.length) return ''
    const visibilityText = requirement.visibilityMode === 'show' ? '显示' : '隐藏'
    return `功能权限: 不满足${visibilityText}`
  }

  const getOperationList = (row: any): ButtonMoreItem[] => {
    if (isManageGroupRow(row)) return []
    const list: ButtonMoreItem[] = [
      { key: 'add', label: '新增子菜单', icon: 'ri:add-fill', auth: 'system.menu.manage' },
      { key: 'edit', label: '编辑菜单', icon: 'ri:edit-2-line', auth: 'system.menu.manage' },
      {
        key: 'action_requirement',
        label: '功能权限',
        icon: 'ri:shield-keyhole-line',
        auth: 'system.menu.manage'
      }
    ]
    if (!row.is_system) {
      list.push({
        key: 'delete',
        label: '删除菜单',
        icon: 'ri:delete-bin-4-line',
        color: '#f56c6c',
        auth: 'system.menu.manage'
      })
    }
    return list
  }

  // --- 事件处理 ---
  const handleReset = () => {
    Object.assign(formFilters, initialSearchState)
    Object.assign(appliedFilters, initialSearchState)
  }
  const handleSearch = () => {
    Object.assign(appliedFilters, formFilters)
  }

  const syncRouteSpaceKey = (spaceKey: string) => {
    router.replace({
      query: {
        ...route.query,
        spaceKey: spaceKey || undefined
      }
    })
  }

  const goToDefinitionManagement = () => {
    router.push({
      path: '/system/menu',
      query: {
        ...route.query,
        spaceKey: activeSpaceKey.value || undefined,
        layout: undefined
      }
    })
  }

  const resolveInitialSpaceKey = () => {
    const requestedSpaceKey = `${route.query.spaceKey || ''}`.trim()
    if (requestedSpaceKey && menuSpaces.value.some((item) => item.spaceKey === requestedSpaceKey)) {
      return requestedSpaceKey
    }
    if (isLayoutMode.value) {
      return ''
    }
    return (
      menuSpaces.value.find((item) => item.isDefault)?.spaceKey ||
      menuSpaces.value[0]?.spaceKey ||
      ''
    )
  }

  const syncMenuSpaces = async () => {
    if (!targetAppKey.value) {
      menuSpaces.value = []
      activeSpaceKey.value = ''
      return
    }
    const res = await fetchGetMenuSpaces(targetAppKey.value)
    menuSpaces.value = res.records || []
    activeSpaceKey.value = resolveInitialSpaceKey()
  }

  const handleSpaceChange = () => {
    syncRouteSpaceKey(activeSpaceKey.value)
    getMenuList()
  }
  const loadAppOptions = async () => {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }
  const handleManagedAppChange = async (value?: string) => {
    await setManagedAppKey(`${value || ''}`.trim())
    activeSpaceKey.value = ''
    await router.replace({
      query: {
        ...route.query,
        spaceKey: undefined
      }
    })
  }
  const rowKey = (row: any) => String(row.id || row.path)

  const clearBatchSelection = () => {
    selectedMenuRows.value = []
    batchTargetGroupId.value = ''
    batchAssignDialogVisible.value = false
    tableRef.value?.elTableRef?.clearSelection?.()
  }

  const handleBatchSelectionChange = (rows: any[]) => {
    selectedMenuRows.value = (rows || []).filter((row) => !isManageGroupRow(row))
  }

  const collectMenuSubtree = (rows: any[]) => {
    const result: any[] = []
    const seen = new Set<string>()

    const visit = (row: any) => {
      if (!row || isManageGroupRow(row)) return
      const key = String(row.id || row.path || '')
      if (!key || seen.has(key)) return
      seen.add(key)
      result.push(row)
      ;(row.children || []).forEach((child: any) => visit(child))
    }

    rows.forEach((row) => visit(row))
    return result
  }

  const getMenuChildCount = (row: any) =>
    (row?.children || []).filter((item: any) => !isManageGroupRow(item)).length

  const getMenuDescendantCount = (row: any) => {
    if (!row) return 0
    return collectMenuSubtree([row]).length
  }

  const getAffectedPageCount = (row: any) => {
    if (!row?.id) return 0
    const subtree = collectMenuSubtree([row]).map((item) => String(item.id || ''))
    if (subtree.length === 0) return 0
    const seen = new Set<string>()
    let count = 0
    subtree.forEach((menuId) => {
      const pages = getLinkedPages({ id: menuId })
      pages.forEach((page) => {
        const key = String(page.pageKey || page.id || '')
        if (!key || seen.has(key)) return
        seen.add(key)
        count += 1
      })
    })
    return count
  }

  const getDeleteParentOptions = (row: any): MenuDeleteParentOption[] => {
    if (!row?.id) return []
    const excluded = new Set<string>(
      collectMenuSubtree([row]).map((item) => String(item.id || ''))
    )
    const walk = (items: AppRouteRecord[]) => {
      return items.reduce<MenuDeleteParentOption[]>((acc, item) => {
        if (isManageGroupRow(item)) return acc
        const key = String(item.id || '')
        if (!key || excluded.has(key)) return acc
        const children = item.children ? walk(item.children as AppRouteRecord[]) : []
        acc.push({
          label: formatMenuTitle(item.meta?.title) || String(item.name || key),
          value: key,
          children: children.length > 0 ? children : undefined
        })
        return acc
      }, [])
    }
    return walk(filteredMenuTree.value)
  }

  const handleMoreActionCommand = (command: string) => {
    if (command === 'manageGroup') {
      if (!groupingAvailable.value) return
      manageGroupDrawerVisible.value = true
      return
    }
    if (command === 'backupGlobal') {
      handleBackupMenu('global')
      return
    }
    if (command === 'backupList') {
      handleManageBackups()
    }
  }

  const hasOwnManageGroup = (row: any) => Boolean(`${row?.manage_group_id || ''}`.trim())

  const buildMenuUpdatePayloadFromRow = (row: any, manageGroupID: string | null) => {
    const meta = buildMenuMetaForUpdate(row)
    return {
      parent_id: row.parent_id ? String(row.parent_id) : null,
      kind: row.kind || 'directory',
      path: row.path || '',
      name: row.name || '',
      component: typeof row.component === 'string' ? row.component : '',
      title: row.meta?.title || '',
      icon: row.meta?.icon || '',
      sort_order: Number(row.sort_order ?? 0),
      space_key:
        `${row.spaceKey || row.space_key || row.meta?.spaceKey || activeSpaceKey.value || ''}`.trim(),
      manage_group_id: manageGroupID,
      meta
    }
  }

  const applyBatchGroupAction = async (action: 'assign' | 'remove', assignGroupID?: string) => {
    if (selectedMenuRows.value.length === 0) {
      ElMessage.warning('请先勾选菜单')
      return
    }
    const nextGroupID = action === 'assign' ? `${assignGroupID || ''}`.trim() : ''
    if (action === 'assign' && !nextGroupID) {
      ElMessage.warning('请选择目标分组')
      return
    }
    const targetGroupID = action === 'remove' ? null : nextGroupID
    const expandedRows = collectMenuSubtree(selectedMenuRows.value)
    const actionableRows =
      action === 'remove' ? expandedRows.filter((row) => hasOwnManageGroup(row)) : expandedRows

    if (action === 'remove' && actionableRows.length === 0) {
      ElMessage.warning('所选菜单及其下级没有已绑定的分组')
      return
    }

    const text = action === 'remove' ? '移出所选菜单的分组归属' : '移入所选菜单到目标分组'
    try {
      await ElMessageBox.confirm(`确定要${text}吗？`, '批量操作确认', { type: 'warning' })
      await Promise.all(
        actionableRows.map((row) =>
          fetchUpdateMenu(String(row.id), buildMenuUpdatePayloadFromRow(row, targetGroupID), {
            showErrorMessage: false
          })
        )
      )
      if (action === 'remove' && actionableRows.length !== expandedRows.length) {
        ElMessage.success(
          `批量移出成功，已跳过 ${expandedRows.length - actionableRows.length} 项未绑定分组菜单`
        )
      } else {
        ElMessage.success('批量操作成功')
      }
      await getMenuList()
      clearBatchSelection()
      batchTargetGroupId.value = ''
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '批量操作失败')
      }
    }
  }

  const handleBatchCommand = (command: 'assign' | 'remove') => {
    if (command === 'assign') {
      if (selectedMenuRows.value.length === 0) {
        ElMessage.warning('请先勾选菜单')
        return
      }
      batchTargetGroupId.value = ''
      batchAssignDialogVisible.value = true
      return
    }
    applyBatchGroupAction(command)
  }

  const handleBatchAssignSubmit = async () => {
    await applyBatchGroupAction('assign', batchTargetGroupId.value)
  }

  const setExpandState = (expanded: boolean) => {
    isExpanded.value = expanded
    nextTick(() => {
      if (tableRef.value?.elTableRef && tableData.value.length) {
        const processRows = (rows: any[]) => {
          rows.forEach((row) => {
            if (row.children?.length) {
              tableRef.value.elTableRef.toggleRowExpansion(row, isExpanded.value)
              processRows(row.children)
            }
          })
        }
        processRows(tableData.value)
      }
    })
  }

  const handleExpandSwitchChange = (value: string | number | boolean) => {
    setExpandState(Boolean(value))
  }

  // --- CRUD 操作 ---
  const handleAddMenu = () => {
    editData.value = null
    parentRowForAdd.value = null
    dialogVisible.value = true
  }
  const handleAddUnderRow = (row: any) => {
    if (isManageGroupRow(row)) return
    editData.value = null
    parentRowForAdd.value = row
    dialogVisible.value = true
  }
  const handleEditMenu = (row: any) => {
    if (isManageGroupRow(row)) return
    editData.value = row
    parentRowForAdd.value = null
    dialogVisible.value = true
  }
  const handleEditActionRequirement = (row: any) => {
    actionRequirementData.value = row
    actionRequirementVisible.value = true
  }

  const normalizeRequiredActions = (items?: string[]) =>
    Array.from(new Set((items || []).map((item) => `${item || ''}`.trim()).filter(Boolean)))

  const applyActionRequirementToMeta = (
    meta: Record<string, any>,
    formData: {
      requiredActions?: string[]
      actionMatchMode?: 'any' | 'all'
      actionVisibilityMode?: 'hide' | 'show'
    }
  ) => {
    const requiredActions = normalizeRequiredActions(formData.requiredActions)
    delete meta.requiredAction
    delete meta.requiredActions
    delete meta.actionMatchMode
    delete meta.actionVisibilityMode

    if (requiredActions.length === 1) {
      meta.requiredAction = requiredActions[0]
    }
    if (requiredActions.length > 1) {
      meta.requiredActions = requiredActions
      meta.actionMatchMode = formData.actionMatchMode === 'all' ? 'all' : 'any'
    }
    if (requiredActions.length > 0) {
      meta.actionVisibilityMode = formData.actionVisibilityMode === 'show' ? 'show' : 'hide'
    }
  }

  const buildMenuMetaFromForm = (formData: any) => {
    const isEntry = `${formData.kind || 'entry'}` === 'entry'
    const isExternal = `${formData.kind || ''}` === 'external'
    const workingSpaceKey = `${formData.spaceKey || activeSpaceKey.value || ''}`.trim()
    const meta: Record<string, any> = {
      isEnable: formData.isEnable,
      keepAlive: isEntry ? formData.keepAlive : false,
      isHide: !!formData.isHide,
      isHideTab: isEntry ? formData.isHideTab : false,
      isIframe: `${formData.kind || ''}` !== 'directory' ? formData.isIframe : false,
      showBadge: formData.showBadge,
      showTextBadge: formData.showTextBadge || '',
      link: isExternal ? formData.link || '' : '',
      activePath: isEntry ? formData.activePath || '' : '',
      fixedTab: isEntry ? formData.fixedTab : false,
      isFullPage: isEntry ? formData.isFullPage : false,
      accessMode: formData.accessMode || 'permission',
      spaceKey: workingSpaceKey
    }
    if (isEntry && formData.customParent?.trim()) {
      meta.customParent = formData.customParent.trim()
    }
    applyActionRequirementToMeta(meta, formData)
    return meta
  }

  const buildMenuRequestPayload = (formData: any, meta: Record<string, any>) => ({
    app_key: targetAppKey.value,
    kind: formData.kind || 'entry',
    path: formData.path || '/',
    name: formData.label || '',
    component: formData.kind === 'entry' ? formData.component || '' : '',
    title: formData.name || '',
    icon: formData.icon || '',
    sort_order: Number(formData.sort ?? 0),
    space_key: `${formData.spaceKey || activeSpaceKey.value || ''}`.trim(),
    manage_group_id: formData.manageGroupId?.trim() || null,
    meta
  })

  const resolveParentId = (formData: any) =>
    formData.parentId?.trim() ||
    (parentRowForAdd.value?.id ? String(parentRowForAdd.value.id) : null)
  const handleMenuOperation = (item: ButtonMoreItem, row: any) => {
    if (item.key === 'add') handleAddUnderRow(row)
    else if (item.key === 'edit') handleEditMenu(row)
    else if (item.key === 'action_requirement') handleEditActionRequirement(row)
    else if (item.key === 'delete') handleDeleteMenu(row)
  }

  const handleBackupListAction = (action: string, row: Api.SystemManage.MenuBackupItem) => {
    if (action === 'restore') {
      handleRestoreBackup(row)
      return
    }
    if (action === 'delete') {
      handleDeleteBackup(row.id)
    }
  }

  const handleDeleteMenu = async (row: any) => {
    if (!dataFromBackend.value || !row.id) return ElMessage.info('预览模式无法删除')
    if (row.is_system) return ElMessage.warning('系统菜单不可删除')
    deleteTargetRow.value = row
    deleteDialogVisible.value = true
    deleteLoading.value = true
    try {
      deletePreview.value = await fetchGetMenuDeletePreview(String(row.id), { mode: 'cascade' })
    } catch (e: any) {
      ElMessage.error(e?.message || '获取删除预览失败')
      deleteDialogVisible.value = false
      deleteTargetRow.value = null
    } finally {
      deleteLoading.value = false
    }
  }

  const handleDeleteMenuConfirm = async (payload: {
    mode: MenuDeleteMode
    targetParentId?: string | null
  }) => {
    if (!deleteTargetRow.value?.id) return
    deleteLoading.value = true
    try {
      await fetchDeleteMenu(String(deleteTargetRow.value.id), {
        mode: payload.mode,
        targetParentId: payload.targetParentId || undefined
      })
      ElMessage.success(payload.mode === 'cascade' ? '菜单树已删除' : '菜单已删除')
      deleteDialogVisible.value = false
      deleteTargetRow.value = null
      deletePreview.value = null
      await getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    } finally {
      deleteLoading.value = false
    }
  }

  const handleSubmit = async (formData: any) => {
    if (!dataFromBackend.value) return getMenuList()
    try {
      const nextManageGroupID = formData.manageGroupId?.trim() || null
      const currentManageGroupID = editData.value?.manage_group_id
        ? String(editData.value.manage_group_id)
        : null
      const payload = buildMenuRequestPayload(formData, buildMenuMetaFromForm(formData))
      if (editData.value?.id) {
        const parentId = formData.parentId?.trim() || null
        await fetchUpdateMenu(
          String(editData.value.id),
          { ...payload, parent_id: parentId },
          { showErrorMessage: false }
        )

        if (currentManageGroupID !== nextManageGroupID) {
          const descendants = collectMenuSubtree(editData.value.children || []).filter(
            (row) => !isManageGroupRow(row)
          )
          if (descendants.length) {
            await Promise.all(
              descendants.map((row) =>
                fetchUpdateMenu(
                  String(row.id),
                  buildMenuUpdatePayloadFromRow(row, nextManageGroupID),
                  { showErrorMessage: false }
                )
              )
            )
          }
        }
      } else {
        const parentId = resolveParentId(formData)
        await fetchCreateMenu({ ...payload, parent_id: parentId }, { showErrorMessage: false })
      }
      ElMessage.success('保存成功')
      getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  }

  const buildMenuMetaForUpdate = (row: any) => {
    const meta = { ...(row?.meta || {}) }
    delete meta.title
    return meta
  }

  const handleActionRequirementSubmit = async (formData: {
    requiredActions: string[]
    actionMatchMode: 'any' | 'all'
    actionVisibilityMode: 'hide' | 'show'
  }) => {
    if (!dataFromBackend.value || !actionRequirementData.value?.id) return
    try {
      const row = actionRequirementData.value
      const meta = buildMenuMetaForUpdate(row)
      applyActionRequirementToMeta(meta, formData)
      await fetchUpdateMenu(
        String(row.id),
        {
          app_key: targetAppKey.value,
          parent_id: row.parent_id ? String(row.parent_id) : null,
          kind: row.kind || 'directory',
          path: row.path || '',
          name: row.name || '',
          component: typeof row.component === 'string' ? row.component : '',
          title: row.meta?.title || '',
          icon: row.meta?.icon || '',
          sort_order: Number(row.sort_order ?? 0),
          space_key:
            `${row.spaceKey || row.space_key || row.meta?.spaceKey || activeSpaceKey.value || ''}`.trim(),
          meta
        },
        { showErrorMessage: false }
      )
      ElMessage.success('功能权限已保存')
      actionRequirementVisible.value = false
      actionRequirementData.value = null
      getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '功能权限保存失败')
    }
  }

  // --- 菜单备份相关方法 ---
  const handleBackupMenu = (scopeType: MenuBackupScopeType) => {
    backupScopeType.value = scopeType
    backupDialogVisible.value = true
  }

  const handleSaveManageGroup = async (
    payload: Api.SystemManage.MenuManageGroupSaveParams & { id?: string }
  ) => {
    groupSaving.value = true
    try {
      if (payload.id) {
        await fetchUpdateMenuManageGroup(payload.id, payload)
      } else {
        await fetchCreateMenuManageGroup(payload)
      }
      ElMessage.success('菜单分组已保存')
      await getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '菜单分组保存失败')
    } finally {
      groupSaving.value = false
    }
  }

  const handleDeleteManageGroup = async (id: string) => {
    groupSaving.value = true
    try {
      await fetchDeleteMenuManageGroup(id)
      ElMessage.success('菜单分组已删除')
      await getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '菜单分组删除失败')
    } finally {
      groupSaving.value = false
    }
  }

  const handleCreateBackup = async (formData: { name: string; description: string }) => {
    backupLoading.value = true
    try {
      const payload: Api.SystemManage.MenuBackupCreateParams = {
        app_key: targetAppKey.value,
        name: formData.name,
        description: formData.description,
        scope_type: backupScopeType.value
      }
      if (backupScopeType.value === 'space') {
        payload.space_key = activeSpaceKey.value
      }
      await fetchCreateMenuBackup(payload)
      ElMessage.success(backupScopeType.value === 'global' ? '全局备份已创建' : '空间备份已创建')
      backupDialogVisible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message || '备份失败')
    } finally {
      backupLoading.value = false
    }
  }

  const handleManageBackups = async () => {
    backupLoading.value = true
    try {
      const list = await fetchGetMenuBackupList(
        isLayoutMode.value ? activeSpaceKey.value : undefined,
        targetAppKey.value
      )
      const filteredList = (list || []).filter((item: any) =>
        isLayoutMode.value
          ? `${item.scope_type || ''}`.trim() !== 'global'
          : `${item.scope_type || ''}`.trim() === 'global'
      )
      backupList.value = filteredList.map((item: any) => ({
        ...item,
        space_name: getBackupScopeLabel(item)
      }))
      backupListDialogVisible.value = true
    } catch (e: any) {
      ElMessage.error(e?.message || '获取备份列表失败')
    } finally {
      backupLoading.value = false
    }
  }

  const buildBackupRestoreMessage = (item: Api.SystemManage.MenuBackupItem) => {
    if (item.scope_type === 'global') {
      return '确定要恢复该备份吗？该操作会覆盖当前 App 的菜单定义集合。'
    }
    return `确定要恢复该备份吗？恢复后会覆盖当前菜单空间“${getSpaceName(item.space_key || activeSpaceKey.value)}”的菜单配置。`
  }

  const handleRestoreBackup = async (item: Api.SystemManage.MenuBackupItem) => {
    try {
      await ElMessageBox.confirm(buildBackupRestoreMessage(item), '提示', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      backupLoading.value = true
      await fetchRestoreMenuBackup(item.id, targetAppKey.value)
      ElMessage.success('恢复成功')
      backupListDialogVisible.value = false
      getMenuList()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '恢复失败')
      }
    } finally {
      backupLoading.value = false
    }
  }

  const handleDeleteBackup = async (id: string) => {
    try {
      await ElMessageBox.confirm('确定要删除该备份吗？', '提示', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      backupLoading.value = true
      await fetchDeleteMenuBackup(id, targetAppKey.value)
      ElMessage.success('删除成功')
      handleManageBackups()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '删除失败')
      }
    } finally {
      backupLoading.value = false
    }
  }

  watch(multiSelectEnabled, (enabled) => {
    if (!enabled) {
      clearBatchSelection()
    }
  })

  watch(deleteDialogVisible, (visible) => {
    if (!visible && !deleteLoading.value) {
      deleteTargetRow.value = null
      deletePreview.value = null
    }
  })

  watch(groupingEnabled, (enabled) => {
    if (!enabled) {
      setExpandState(false)
    }
    localStorage.setItem('system:menu:grouping-enabled', enabled ? '1' : '0')
  })

  watch(menuGroupApiUnavailable, (unavailable) => {
    if (unavailable) {
      setExpandState(false)
    }
  })

  watch(groupedMenuVisible, (visible) => {
    localStorage.setItem('system:menu:grouped-visible', visible ? '1' : '0')
  })

  // --- 生命周期 & 监听 ---
  onMounted(() => {
    groupingEnabled.value = localStorage.getItem('system:menu:grouping-enabled') !== '0'
    groupedMenuVisible.value = localStorage.getItem('system:menu:grouped-visible') !== '0'
    selectedAppKey.value = targetAppKey.value
    loadAppOptions().catch((error) => {
      warnDev('[Menus] 加载 App 列表失败', error)
      appList.value = []
    })
    if (!targetAppKey.value) {
      loadError.value = managedAppMissingText
      return
    }
    syncMenuSpaces().finally(() => {
      getMenuList()
    })
  })

  watch(
    () => [targetAppKey.value, route.query.spaceKey],
    async ([appKey, value]) => {
      if (!targetAppKey.value) {
        rawMenuTree.value = []
        rawPages.value = []
        menuSpaces.value = []
        activeSpaceKey.value = ''
        loadError.value = managedAppMissingText
        return
      }
      if (!appKey) return
      await syncMenuSpaces()
      const requestedSpaceKey = `${value || ''}`.trim()
      if (!requestedSpaceKey || requestedSpaceKey === activeSpaceKey.value) {
        await getMenuList()
        return
      }
      if (!menuSpaces.value.some((item) => item.spaceKey === requestedSpaceKey)) {
        await getMenuList()
        return
      }
      activeSpaceKey.value = requestedSpaceKey
      await getMenuList()
    }
  )

  watch(
    () => targetAppKey.value,
    (value) => {
      selectedAppKey.value = value || ''
    },
    { immediate: true }
  )

  return {
    // state
    loading,
    loadError,
    showSearchBar,
    isExpanded,
    showHiddenMenus,
    showIframeMenus,
    showEnabledMenus,
    groupingEnabled,
    groupedMenuVisible,
    tableRef,
    multiSelectEnabled,
    activeSpaceKey,
    selectedAppKey,
    menuGroups,
    menuSpaces,
    menuGroupApiUnavailable,
    backupLoading,
    backupDialogVisible,
    backupListDialogVisible,
    backupScopeType,
    backupList,
    formFilters,
    dialogVisible,
    manageGroupDrawerVisible,
    groupSaving,
    editData,
    parentRowForAdd,
    deleteDialogVisible,
    deleteLoading,
    deleteTargetRow,
    deletePreview,
    actionRequirementVisible,
    actionRequirementData,
    selectedMenuRows,
    batchTargetGroupId,
    batchAssignDialogVisible,
    targetAppKey,
    isLayoutMode,
    // computed
    menuSpaceOptions,
    appOptions,
    currentSpaceName,
    menuPageTitle,
    menuPageDescription,
    menuToolbarTip,
    backupDialogTitle,
    backupAlertTitle,
    backupAlertDescription,
    backupListTitle,
    backupListAlertDescription,
    backupListEmptyDescription,
    filteredMenuTree,
    groupingAvailable,
    tableData,
    menuHeroMetrics,
    columnChecks,
    displayColumns,
    // methods (template-facing)
    getLinkedPages,
    getSpaceName,
    getMenuActionRequirementLabel,
    getOperationList,
    handleReset,
    handleSearch,
    goToDefinitionManagement,
    handleSpaceChange,
    handleManagedAppChange,
    rowKey,
    handleBatchSelectionChange,
    getMenuChildCount,
    getMenuDescendantCount,
    getAffectedPageCount,
    getDeleteParentOptions,
    handleMoreActionCommand,
    handleBatchCommand,
    handleBatchAssignSubmit,
    handleExpandSwitchChange,
    handleAddMenu,
    handleMenuOperation,
    handleBackupListAction,
    handleDeleteMenuConfirm,
    handleSubmit,
    handleActionRequirementSubmit,
    handleSaveManageGroup,
    handleDeleteManageGroup,
    handleCreateBackup,
    // re-exports for template convenience
    isManageGroupRow,
    isDirectoryMenuRow,
    isEntryMenuRow,
    getMenuTypeTag,
    getMenuTypeText,
    getAccessModeLabel,
    getAccessModeTag,
    formatMenuTitle,
    getMenuActionRequirement
  }
}
