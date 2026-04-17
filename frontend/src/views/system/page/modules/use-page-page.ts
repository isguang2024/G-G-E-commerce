/**
 * page 视图主 composable。
 *
 * 将原 index.vue 的 1200+ 行 script 整体抽离，
 * 视图层只保留 template + 极少量调用胶水。
 *
 * 设计原则：
 * - 一次性拉出全部 reactive state、handler、watch 与 lifecycle；
 * - 纯函数 helper 已抽到 ./helpers.ts；
 * - 返回值即视图模板使用的所有变量与方法。
 */
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormItem } from '@/components/core/forms/art-form/index.vue'
import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
import { useTableColumns } from '@/hooks/core/useTableColumns'
import { useManagedAppScope } from '@/domains/app-runtime/useManagedAppScope'
import {
  fetchDeletePage,
  fetchGetApps,
  fetchGetMenuSpaces,
  fetchGetPageList,
  fetchGetPageMenuOptions,
  fetchUpdatePage
} from '@/domains/governance/api'
import { router } from '@/router'
import { joinManagedPagePath, resolveManagedPageRoutePath } from '@/domains/navigation/utils/managed-page'
import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import {
  buildCopyPageData,
  buildMenuPathMap,
  buildPageTree,
  countTreeNodes,
  formatUpdatedAt,
  getAccessModeTag,
  getAccessModeText,
  getMountModeText,
  getMountTargetText,
  getOperationList,
  getPageGovernanceText,
  getPageSourceKind,
  getPageSourceTag,
  getPageSourceText,
  getPageTypeTag,
  getPageTypeText,
  getRelationDisplayText,
  normalizeKeyword,
  toPageSaveParams,
  type TreePageItem
} from './helpers'

type PageItem = Api.SystemManage.PageItem

export function usePagePage() {
  const loading = ref(false)
  const loadError = ref('')
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const managedAppMissingText = '请选择当前要管理的 App'
  const showSearchBar = ref(false)
  const isExpanded = ref(false)
  const syncing = ref(false)
  const showSuspended = ref(true)
  const sortDraftMap = reactive<Record<string, string>>({})
  const savingSortIds = ref(new Set<string>())
  const editingSortId = ref('')
  const tableRef = ref()
  const rawPages = ref<PageItem[]>([])
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const menuPathMap = ref(new Map<string, string>())
  const menuSpaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const activeSpaceKey = ref('')
  const menuSpaceStore = useMenuSpaceStore()

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit' | 'copy'>('add')
  const currentPage = ref<Partial<PageItem>>({})
  const defaultPageData = ref<Partial<PageItem>>({})
  const initialParentPageKey = ref('')
  const initialParentMenuId = ref('')
  const initialPageType = ref<PageItem['pageType']>('inner')
  const unregisteredDialogVisible = ref(false)
  const initialSearchState = {
    keyword: '',
    pageType: '',
    accessMode: '',
    source: '',
    status: ''
  }
  const searchForm = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })

  const searchItems = computed<FormItem[]>(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: { placeholder: '名称/标识/路由/组件/上级' }
    },
    {
      label: '页面类型',
      key: 'pageType',
      type: 'select',
      props: {
        options: [
          { label: '全部', value: '' },
          { label: '逻辑分组', value: 'group' },
          { label: '普通分组', value: 'display_group' },
          { label: '内页', value: 'inner' },
          { label: '独立页', value: 'standalone' }
        ],
        clearable: true
      }
    },
    {
      label: '访问模式',
      key: 'accessMode',
      type: 'select',
      props: {
        options: [
          { label: '全部', value: '' },
          { label: '继承', value: 'inherit' },
          { label: '公开', value: 'public' },
          { label: '登录', value: 'jwt' },
          { label: '权限', value: 'permission' }
        ],
        clearable: true
      }
    },
    {
      label: '来源',
      key: 'source',
      type: 'select',
      props: {
        options: [
          { label: '全部', value: '' },
          { label: '本地配置', value: 'manual' },
          { label: '扫描同步', value: 'sync' },
          { label: 'Seed', value: 'seed' },
          { label: '远端页', value: 'remote' }
        ],
        clearable: true
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        options: [
          { label: '全部', value: '' },
          { label: '正常', value: 'normal' },
          { label: '停用', value: 'suspended' }
        ],
        clearable: true
      }
    }
  ])

  const { columnChecks, columns: displayColumns } = useTableColumns<PageItem>(() => [
    { prop: 'name', label: '页面', minWidth: 300, useSlot: true, slotName: 'name' },
    { prop: 'route', label: '最终路径', minWidth: 220, useSlot: true, slotName: 'route' },
    { prop: 'component', label: '组件入口', minWidth: 180, useSlot: true, slotName: 'component' },
    {
      prop: 'mountTarget',
      label: '挂接对象',
      minWidth: 180,
      useSlot: true,
      slotName: 'mountTarget'
    },
    {
      prop: 'effectiveChain',
      label: '归属链路',
      minWidth: 150,
      useSlot: true,
      slotName: 'effectiveChain'
    },
    {
      prop: 'parentChainStatus',
      label: '链路状态',
      minWidth: 130,
      useSlot: true,
      slotName: 'parentChainStatus'
    },
    {
      prop: 'sortOrder',
      label: '排序',
      width: 130,
      align: 'center',
      useSlot: true,
      slotName: 'sortOrder'
    },
    {
      prop: 'accessMode',
      label: '挂载 / 访问 / 状态',
      width: 220,
      align: 'center',
      useSlot: true,
      slotName: 'accessMode'
    },
    { prop: 'updatedAt', label: '更新', width: 160, useSlot: true, slotName: 'updatedAt' },
    {
      prop: 'operation',
      label: '操作',
      width: 70,
      align: 'center',
      fixed: 'right',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  const pageTree = computed<TreePageItem[]>(() => buildPageTree(rawPages.value))

  function matchesPageSearch(item: PageItem) {
    if (!showSuspended.value && item.status !== 'normal') return false
    const keyword = normalizeKeyword(appliedFilters.keyword)
    const pageType = `${appliedFilters.pageType || ''}`.trim()
    const accessMode = `${appliedFilters.accessMode || ''}`.trim()
    const source = `${appliedFilters.source || ''}`.trim()
    const status = `${appliedFilters.status || ''}`.trim()

    const keywordSource = [
      item.name,
      item.pageKey,
      item.routeName,
      item.routePath,
      item.component,
      item.moduleKey,
      item.parentMenuName,
      item.parentPageName,
      item.parentPageKey,
      item.displayGroupName,
      item.displayGroupKey
    ]
      .join(' ')
      .toLowerCase()

    if (keyword && !keywordSource.includes(keyword)) return false
    if (pageType && item.pageType !== pageType) return false
    if (accessMode && item.accessMode !== accessMode) return false
    if (source && getPageSourceKind(item) !== source) return false
    if (status && item.status !== status) return false
    return true
  }

  function filterPageTree(items: TreePageItem[]): TreePageItem[] {
    return items.reduce<TreePageItem[]>((result, item) => {
      const children = item.children?.length ? filterPageTree(item.children) : []
      if (!matchesPageSearch(item) && children.length === 0) {
        return result
      }
      result.push({
        ...item,
        children
      })
      return result
    }, [])
  }

  const tableData = computed<TreePageItem[]>(() => filterPageTree(pageTree.value))
  const pageMap = computed(() => {
    const map = new Map<string, PageItem>()
    rawPages.value.forEach((item) => {
      if (item.pageKey) {
        map.set(item.pageKey, item)
      }
    })
    return map
  })
  const visibleCount = computed(() => countTreeNodes(tableData.value))
  const summaryStats = computed(() => {
    const suspendedCount = rawPages.value.filter((item) => item.status !== 'normal').length
    const managedEntryCount = rawPages.value.filter(
      (item) =>
        item.pageType === 'inner' || item.pageType === 'standalone'
    ).length
    const logicGroupCount = rawPages.value.filter((item) => item.pageType === 'group').length
    const displayGroupCount = rawPages.value.filter(
      (item) => item.pageType === 'display_group'
    ).length
    const remoteCount = rawPages.value.filter((item) => getPageSourceKind(item) === 'remote').length
    const syncCount = rawPages.value.filter((item) => getPageSourceKind(item) === 'sync').length
    const localCount = rawPages.value.filter((item) => getPageSourceKind(item) === 'manual').length
    return [
      { label: '当前 App', value: targetAppKey.value },
      { label: '当前显示', value: visibleCount.value },
      { label: '受管页面', value: managedEntryCount || 0 },
      { label: '本地配置', value: localCount || 0 },
      { label: '扫描同步', value: syncCount || 0 },
      { label: '远端页', value: remoteCount || 0 },
      { label: '逻辑分组', value: logicGroupCount || 0 },
      { label: '普通分组', value: displayGroupCount || 0 },
      { label: '停用', value: suspendedCount || 0 },
      { label: '总条目', value: rawPages.value.length }
    ]
  })
  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )

  async function getPageList() {
    loading.value = true
    loadError.value = ''
    if (!targetAppKey.value) {
      rawPages.value = []
      menuPathMap.value = new Map()
      activeSpaceKey.value = ''
      loadError.value = managedAppMissingText
      loading.value = false
      return
    }
    try {
      const scopeSpaceKey = activeSpaceKey.value || undefined
      const [pageResponse, menuResponse] = await Promise.all([
        fetchGetPageList({
          current: 1,
          size: 1000,
          appKey: targetAppKey.value,
          menuSpaceKey: scopeSpaceKey
        }),
        fetchGetPageMenuOptions(scopeSpaceKey, targetAppKey.value)
      ])
      rawPages.value = pageResponse.records || []
      Object.keys(sortDraftMap).forEach((key) => delete sortDraftMap[key])
      editingSortId.value = ''
      rawPages.value.forEach((item) => {
        if (item.id) {
          sortDraftMap[item.id] = `${Number(item.sortOrder || 0)}`
        }
      })
      menuPathMap.value = buildMenuPathMap(menuResponse.records || [], joinManagedPagePath)
    } catch (error: any) {
      rawPages.value = []
      Object.keys(sortDraftMap).forEach((key) => delete sortDraftMap[key])
      editingSortId.value = ''
      menuPathMap.value = new Map()
      loadError.value = error?.message || '页面数据暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  function handleSearch() {
    Object.assign(appliedFilters, {
      keyword: searchForm.keyword.trim(),
      pageType: searchForm.pageType,
      accessMode: searchForm.accessMode,
      source: searchForm.source,
      status: searchForm.status
    })
  }

  function handleSpaceScopeChange() {
    getPageList()
  }
  async function loadAppOptions() {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }
  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
    activeSpaceKey.value = ''
  }

  function handleReset() {
    Object.assign(searchForm, initialSearchState)
    Object.assign(appliedFilters, initialSearchState)
  }

  function handleRefresh() {
    getPageList()
  }

  async function openDialog(
    type: 'add' | 'edit' | 'copy',
    row?: PageItem,
    options?: {
      parentPageKey?: string
      parentMenuId?: string
      pageType?: PageItem['pageType']
      defaultData?: Partial<PageItem>
    }
  ) {
    dialogVisible.value = false
    await nextTick()
    dialogType.value = type
    currentPage.value = type === 'edit' && row ? { ...row } : {}
    defaultPageData.value = {
      ...(options?.defaultData ? { ...options.defaultData } : {})
    }
    initialParentPageKey.value = options?.parentPageKey || ''
    initialParentMenuId.value = options?.parentMenuId || ''
    initialPageType.value = options?.pageType || 'standalone'
    await nextTick()
    dialogVisible.value = true
  }

  function handleAddPage() {
    openDialog('add', undefined, { pageType: 'standalone' })
  }

  function handleAddLogicGroup() {
    openDialog('add', undefined, { pageType: 'group' })
  }

  function handleAddDisplayGroup() {
    openDialog('add', undefined, { pageType: 'display_group' })
  }

  function handleCreateCommand(command: string) {
    if (command === 'page') {
      handleAddPage()
      return
    }
    if (command === 'group') {
      handleAddLogicGroup()
      return
    }
    if (command === 'display_group') {
      handleAddDisplayGroup()
    }
  }

  function handleAddChild(
    row: PageItem,
    pageType: PageItem['pageType'],
    options?: { displayGroupKey?: string; parentPageKey?: string; parentMenuId?: string }
  ) {
    openDialog('add', undefined, {
      pageType,
      parentPageKey:
        options?.parentPageKey ?? (row.pageType === 'display_group' ? '' : row.pageKey),
      parentMenuId:
        options?.parentMenuId ?? (row.pageType === 'display_group' ? '' : row.parentMenuId || ''),
      defaultData: options?.displayGroupKey
        ? { displayGroupKey: options.displayGroupKey }
        : undefined
    })
  }

  function handleCopy(row: PageItem) {
    openDialog('copy', undefined, {
      pageType: row.pageType,
      parentPageKey: row.parentPageKey || '',
      parentMenuId: row.parentMenuId || '',
      defaultData: buildCopyPageData(row)
    })
  }

  function handleOperation(item: ButtonMoreItem, row: PageItem) {
    if (item.key === 'add-group') {
      handleAddChild(row, 'group', {
        displayGroupKey: row.pageType === 'display_group' ? row.pageKey : '',
        parentPageKey: row.pageType === 'display_group' ? '' : row.pageKey,
        parentMenuId: row.pageType === 'display_group' ? '' : row.parentMenuId || ''
      })
      return
    }
    if (item.key === 'add-page') {
      handleAddChild(row, 'inner', {
        displayGroupKey: row.pageType === 'display_group' ? row.pageKey : '',
        parentPageKey: row.pageType === 'display_group' ? '' : row.pageKey,
        parentMenuId: row.pageType === 'display_group' ? '' : row.parentMenuId || ''
      })
      return
    }
    if (item.key === 'edit') {
      openDialog('edit', row)
      return
    }
    if (item.key === 'copy') {
      handleCopy(row)
      return
    }
    if (item.key === 'visit') {
      handleVisit(row)
      return
    }
    if (item.key === 'delete') {
      handleDelete(row)
    }
  }

  function getResolvedRoutePath(row: PageItem) {
    return resolveManagedPageRoutePath(row, {
      getPageByKey: (pageKey) => pageMap.value.get(pageKey),
      getMenuPathById: (menuId) => menuPathMap.value.get(menuId)
    })
  }

  function getRouteDisplayText(row: PageItem) {
    if (row.pageType === 'display_group') {
      return ''
    }
    const resolvedPath = getResolvedRoutePath(row)
    if (resolvedPath) {
      return resolvedPath
    }
    if (row.pageType === 'group') {
      return `${row.routePath || ''}`.trim()
    }
    return ''
  }

  function resolveVisitTarget(row: PageItem): string {
    const link = `${row.link || ''}`.trim()
    if (link) return link
    const resolvedPath = getResolvedRoutePath(row)
    const targetSpaceKey = `${(row as any).menuSpaceKeys?.[0] || menuSpaceStore.currentSpaceKey || ''}`.trim()
    const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(resolvedPath, targetSpaceKey)
    if (nextTarget.mode === 'location') {
      return nextTarget.target
    }
    return new URL(router.resolve(nextTarget.target).href, window.location.origin).toString()
  }

  function handleVisit(row: PageItem) {
    const target = resolveVisitTarget(row)
    if (!target) {
      ElMessage.warning('当前页面未配置可访问地址')
      return
    }
    window.open(target, '_blank')
  }

  function handleCreateFromCandidate(candidate: Partial<PageItem>) {
    openDialog('add', undefined, {
      parentPageKey: candidate.parentPageKey || '',
      parentMenuId: candidate.parentMenuId || '',
      pageType: candidate.pageType || 'standalone',
      defaultData: {
        ...candidate,
        meta: {
          ...(candidate.meta || {}),
          fromUnregistered: true
        }
      }
    })
  }

  async function handleDelete(row: PageItem) {
    try {
      await ElMessageBox.confirm(
        `确认删除${row.pageType === 'group' ? '逻辑分组' : row.pageType === 'display_group' ? '普通分组' : '页面'}“${row.name}”吗？`,
        '删除确认',
        { type: 'warning' }
      )
      await fetchDeletePage(row.id, targetAppKey.value)
      ElMessage.success('删除成功')
      getPageList()
    } catch (error: any) {
      if (error === 'cancel' || error === 'close') return
      ElMessage.error(error?.message || '删除失败')
    }
  }

  function getEffectiveChainText(row: PageItem) {
    if (row.pageType === 'display_group') return '仅分组展示'
    if (row.pageType === 'group' && !row.parentPageKey && !row.parentMenuId) return '分组独立生效'
    if (row.parentMenuId) {
      if (row.accessMode && row.accessMode !== 'inherit') {
        return `菜单权限 + 页面权限（${getAccessModeText(row.accessMode)}）`
      }
      return '默认继承菜单权限'
    }
    if (row.accessMode && row.accessMode !== 'inherit')
      return `自身（${getAccessModeText(row.accessMode)}）`
    if (row.parentPageKey) {
      const parent = pageMap.value.get(row.parentPageKey)
      return parent ? `继承页面（${parent.name}）` : '继承页面（父级缺失）'
    }
    return '自身生效'
  }

  function getParentChainStatusText(row: PageItem) {
    if (row.pageType === 'standalone') return '无父链（独立页）'
    if (!row.parentPageKey && !row.parentMenuId) return '无父链'
    if (row.parentMenuId && !row.parentPageKey) return '菜单链路'
    const visited = new Set<string>()
    let currentKey = `${row.parentPageKey || ''}`.trim()
    while (currentKey) {
      if (visited.has(currentKey)) return '异常（循环引用）'
      visited.add(currentKey)
      const current = pageMap.value.get(currentKey)
      if (!current) return '异常（父级缺失）'
      if (current.status !== 'normal') return '异常（父级停用）'
      currentKey = `${current.parentPageKey || ''}`.trim()
    }
    return '正常'
  }

  function startSortEdit(row: PageItem) {
    if (!row.id) return
    editingSortId.value = row.id
    sortDraftMap[row.id] = `${Number(row.sortOrder || 0)}`
  }

  function cancelSortEdit(row: PageItem) {
    if (!row.id) return
    sortDraftMap[row.id] = `${Number(row.sortOrder || 0)}`
    if (editingSortId.value === row.id) {
      editingSortId.value = ''
    }
  }

  async function saveSortOrder(row: PageItem) {
    if (!row.id) return
    const rawValue = `${sortDraftMap[row.id] ?? ''}`.trim()
    if (!/^\d+$/.test(rawValue)) {
      ElMessage.warning('排序值需为大于等于 0 的数字')
      return
    }
    const nextSortOrder = Number(rawValue)
    if (nextSortOrder === Number(row.sortOrder || 0)) return
    if (savingSortIds.value.has(row.id)) return
    savingSortIds.value.add(row.id)
    try {
      await fetchUpdatePage(row.id, toPageSaveParams(row, nextSortOrder, targetAppKey.value))
      ElMessage.success('排序已保存')
      editingSortId.value = ''
      await getPageList()
    } catch (error: any) {
      ElMessage.error(error?.message || '排序保存失败')
      sortDraftMap[row.id] = `${Number(row.sortOrder || 0)}`
    } finally {
      savingSortIds.value.delete(row.id)
    }
  }

  function toggleExpand(expanded?: boolean) {
    isExpanded.value = typeof expanded === 'boolean' ? expanded : !isExpanded.value
    nextTick(() => {
      if (!tableRef.value?.elTableRef || !tableData.value.length) return
      const processRows = (rows: TreePageItem[]) => {
        rows.forEach((row) => {
          if (row.children?.length) {
            tableRef.value.elTableRef.toggleRowExpansion(row, isExpanded.value)
            processRows(row.children)
          }
        })
      }
      processRows(tableData.value)
    })
  }

  function handleExpandSwitchChange(value: string | number | boolean) {
    toggleExpand(Boolean(value))
  }

  const rowKey = (row: PageItem) => String(row.id || row.pageKey)

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
    loadAppOptions().catch(() => {
      appList.value = []
    })
    if (!targetAppKey.value) {
      loadError.value = managedAppMissingText
      return
    }
    fetchGetMenuSpaces(targetAppKey.value)
      .then((res) => {
        menuSpaces.value = res.records || []
      })
      .finally(() => {
        getPageList()
      })
  })

  watch(
    () => targetAppKey.value,
    async () => {
      if (!targetAppKey.value) {
        menuSpaces.value = []
        rawPages.value = []
        menuPathMap.value = new Map()
        activeSpaceKey.value = ''
        loadError.value = managedAppMissingText
        return
      }
      const spacesRes = await fetchGetMenuSpaces(targetAppKey.value)
      menuSpaces.value = spacesRes.records || []
      await getPageList()
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
    syncing,
    showSuspended,
    sortDraftMap,
    savingSortIds,
    editingSortId,
    tableRef,
    targetAppKey,
    selectedAppKey,
    menuSpaces,
    activeSpaceKey,
    dialogVisible,
    dialogType,
    currentPage,
    defaultPageData,
    initialParentPageKey,
    initialParentMenuId,
    initialPageType,
    unregisteredDialogVisible,
    searchForm,
    // computed
    searchItems,
    columnChecks,
    displayColumns,
    tableData,
    summaryStats,
    appOptions,
    // handlers / helpers
    handleSearch,
    handleReset,
    handleRefresh,
    handleSpaceScopeChange,
    handleManagedAppChange,
    handleCreateCommand,
    handleOperation,
    handleCreateFromCandidate,
    handleExpandSwitchChange,
    getOperationList,
    getRouteDisplayText,
    getRelationDisplayText,
    getPageGovernanceText,
    getPageSourceTag,
    getPageSourceText,
    getMountTargetText,
    getEffectiveChainText,
    getParentChainStatusText,
    getPageTypeTag,
    getPageTypeText,
    getAccessModeText,
    getAccessModeTag,
    getMountModeText,
    formatUpdatedAt,
    startSortEdit,
    cancelSortEdit,
    saveSortOrder,
    rowKey
  }
}
