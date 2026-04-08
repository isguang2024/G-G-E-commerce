/**
 * api-endpoint 视图主 composable。
 *
 * 将原 index.vue (2538 行) 的 1300 行 script 整体抽离，
 * 视图层只保留 template + 极少量调用胶水。
 *
 * 设计原则：
 * - 一次性拉出全部 reactive state、handler 与 useTable，避免拆得过细造成跨文件依赖循环；
 * - 纯函数 formatter / tag-type 已抽到 ./helpers.ts；
 * - 返回值即视图模板使用的所有变量与方法。
 */
import { computed, h, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox, ElOption, ElSelect, ElTag } from 'element-plus'
import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
import { useTable } from '@/hooks/core/useTable'
import { useAuth } from '@/hooks/core/useAuth'
import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
import {
  fetchAddPermissionActionEndpoint,
  fetchCleanupStaleApiEndpoints,
  fetchCreateApiEndpoint,
  fetchCreateApiEndpointCategory,
  fetchDeletePermissionActionEndpoint,
  fetchGetApiEndpointCategories,
  fetchGetApiEndpointList,
  fetchGetApiEndpointOverview,
  fetchGetPermissionActionOptions,
  fetchGetStaleApiEndpointList,
  fetchGetUnregisteredApiScanConfig,
  fetchGetUnregisteredApiRouteList,
  fetchSaveUnregisteredApiScanConfig,
  fetchSyncApiEndpoints,
  fetchUpdateApiEndpoint,
  fetchUpdateApiEndpointCategory,
  fetchUpdateApiEndpointContextScope
} from '@/api/system-manage'
import {
  formatPermissionContext,
  formatPermissionPattern,
  formatSource,
  methodTagType,
  permissionPatternTagType,
  sourceTagType
} from './helpers'

type APIEndpointItem = Api.SystemManage.APIEndpointItem
type APIEndpointCategoryItem = Api.SystemManage.APIEndpointCategoryItem
type APIUnregisteredRouteItem = Api.SystemManage.APIUnregisteredRouteItem

export type CategoryTreeNode = {
  id: string
  label: string
  count: number
  type: 'all' | 'uncategorized' | 'category'
  status?: string
  category?: APIEndpointCategoryItem
  children?: CategoryTreeNode[]
}

type PersistedTableState = {
  selectedSource: string
  selectedCategoryTreeKey: string
  tableQuery: {
    method: string
    path: string
    keyword: string
    permissionKey: string
    permissionPattern: string
    contextScope: string
    featureKind: string
    status: string
    hasPermissionKey: string
  }
}

export function useApiEndpointPage() {
  const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
  const { hasAction } = useAuth()
  const { targetAppKey } = useManagedAppScope()
  const managedAppMissingText = '缺少 app 上下文，请先从应用管理选择 App'
  const API_ENDPOINT_TABLE_STATE_KEY = 'system:api-endpoint:table-state'

  const staleTableRef = ref<any>(null)
  const syncing = ref(false)
  const cleaningStale = ref(false)
  const loadError = ref('')
  const showSearchBar = ref(false)
  const saving = ref(false)
  const categorySaving = ref(false)
  const categorySwitchingId = ref('')
  const selectedSource = ref('')
  const selectedCategoryTreeKey = ref('all')
  const formVisible = ref(false)
  const categoryDrawerVisible = ref(false)
  const permissionBindVisible = ref(false)
  const permissionDialogMode = ref<'add' | 'remove'>('add')
  const unregisteredVisible = ref(false)
  const scanConfigVisible = ref(false)
  const scanConfigSaving = ref(false)
  const staleDialogVisible = ref(false)
  const unregisteredLoading = ref(false)
  const shouldRefreshUnregistered = ref(false)
  const editingId = ref('')
  const pendingLocateRoute = ref<{ method: string; path: string; source: string } | null>(null)
  const categories = ref<APIEndpointCategoryItem[]>([])
  const permissionActionOptions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const permissionActionLoading = ref(false)
  const permissionBinding = reactive({
    endpointCode: '',
    endpointSpec: '',
    endpointPermissionKeys: [] as string[],
    permissionActionId: ''
  })
  const unregisteredRoutes = ref<APIUnregisteredRouteItem[]>([])
  const scanConfig = reactive<Api.SystemManage.APIUnregisteredScanConfig>({
    enabled: false,
    frequencyMinutes: 60,
    defaultCategoryId: '',
    defaultPermissionKey: '',
    markAsNoPermission: false
  })
  const staleCandidates = ref<APIEndpointItem[]>([])
  const selectedStaleIds = ref<string[]>([])
  const totalCount = ref(0)
  const noPermissionCount = ref(0)
  const sharedPermissionCount = ref(0)
  const crossContextSharedCount = ref(0)
  const staleCount = ref(0)
  const unregisteredCount = ref(0)
  const uncategorizedCount = ref(0)
  const categoryCountMap = ref<Record<string, number>>({})
  const unregisteredPagination = reactive({ current: 1, size: 20, total: 0 })
  const stalePagination = reactive({ current: 1, size: 20, total: 0 })

  const formState = reactive({
    appScope: 'app',
    method: 'GET',
    path: '',
    summary: '',
    featureKind: 'system',
    categoryId: '',
    contextScope: 'optional',
    source: 'manual',
    status: 'normal',
    permissionKeys: [] as string[]
  })

  const categoryForm = reactive({
    id: '',
    code: '',
    name: '',
    nameEn: '',
    sortOrder: 0,
    status: 'normal'
  })

  const unregisteredQuery = reactive({
    method: '',
    path: '',
    keyword: '',
    onlyNoMeta: false
  })

  const tableQuery = reactive({
    method: '',
    path: '',
    keyword: '',
    permissionKey: '',
    permissionPattern: '',
    categoryId: '',
    contextScope: '',
    featureKind: '',
    status: '',
    hasPermissionKey: '',
    hasCategory: ''
  })

  const searchForm = reactive({
    source: '',
    method: '',
    path: '',
    keyword: '',
    permissionKey: '',
    permissionPattern: '',
    contextScope: '',
    featureKind: '',
    status: '',
    hasPermissionKey: ''
  })

  function syncSearchFormFromQuery() {
    searchForm.source = selectedSource.value
    searchForm.method = tableQuery.method
    searchForm.path = tableQuery.path
    searchForm.keyword = tableQuery.keyword
    searchForm.permissionKey = tableQuery.permissionKey
    searchForm.permissionPattern = tableQuery.permissionPattern
    searchForm.contextScope = tableQuery.contextScope
    searchForm.featureKind = tableQuery.featureKind
    searchForm.status = tableQuery.status
    searchForm.hasPermissionKey = tableQuery.hasPermissionKey
  }

  function syncQueryFromSearchForm() {
    selectedSource.value = searchForm.source || ''
    tableQuery.method = searchForm.method || ''
    tableQuery.path = searchForm.path || ''
    tableQuery.keyword = searchForm.keyword || ''
    tableQuery.permissionKey = searchForm.permissionKey || ''
    tableQuery.permissionPattern = searchForm.permissionPattern || ''
    tableQuery.contextScope = searchForm.contextScope || ''
    tableQuery.featureKind = searchForm.featureKind || ''
    tableQuery.status = searchForm.status || ''
    tableQuery.hasPermissionKey = searchForm.hasPermissionKey || ''
  }

  const sortedCategories = computed(() =>
    [...categories.value].sort(
      (a, b) =>
        (a.sortOrder ?? 0) - (b.sortOrder ?? 0) ||
        `${a.name || ''}`.localeCompare(`${b.name || ''}`, 'zh-CN')
    )
  )

  const categoryTreeData = computed<CategoryTreeNode[]>(() => [
    {
      id: 'all',
      label: '全部 API',
      count: totalCount.value,
      type: 'all',
      children: [
        {
          id: 'uncategorized',
          label: '未分类',
          count: uncategorizedCount.value,
          type: 'uncategorized'
        },
        ...sortedCategories.value.map((item) => ({
          id: `category:${item.id}`,
          label: item.name || item.code || '未命名分类',
          count: categoryCountMap.value[item.id] || 0,
          type: 'category' as const,
          status: item.status,
          category: item
        }))
      ]
    }
  ])

  const summaryMetrics = computed(() => [
    { label: '管理 App', value: targetAppKey.value || '-' },
    { label: '注册总量', value: totalCount.value || 0 },
    { label: '无权限键', value: noPermissionCount.value || 0 },
    { label: '共享接口', value: sharedPermissionCount.value || 0 },
    { label: '跨空间共享', value: crossContextSharedCount.value || 0 },
    { label: '未分类', value: uncategorizedCount.value || 0 },
    { label: '失效', value: staleCount.value || 0 },
    { label: '未注册', value: unregisteredCount.value || 0 }
  ])

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetApiEndpointList,
      apiParams: {
        current: 1,
        size: 20,
        source: '',
        appKey: targetAppKey.value
      },
      columnsFactory: () => [
        {
          prop: 'method',
          label: 'Method',
          width: 92,
          fixed: 'left',
          formatter: (row: APIEndpointItem) =>
            h(
              ElTag,
              { type: methodTagType(row.method), effect: 'dark' },
              () => row.method
            )
        },
        {
          prop: 'path',
          label: '路径',
          minWidth: 300,
          showOverflowTooltip: true,
          formatter: (row: APIEndpointItem) =>
            h('div', { class: 'path-cell' }, [h('div', { class: 'path-main' }, row.path)])
        },
        {
          prop: 'appScope',
          label: '范围',
          width: 90,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: row.appScope === 'app' ? 'warning' : 'info', effect: 'plain' }, () =>
              row.appScope === 'app' ? 'App' : '共享'
            )
        },
        {
          prop: 'appKey',
          label: 'App',
          width: 140,
          formatter: (row: APIEndpointItem) => row.appKey || '-'
        },
        {
          prop: 'summary',
          label: '介绍',
          minWidth: 220,
          showOverflowTooltip: true,
          formatter: (row: APIEndpointItem) => row.summary || '-'
        },
        {
          prop: 'category',
          label: '分类',
          minWidth: 180,
          formatter: (row: APIEndpointItem) => row.category?.name || '-'
        },
        {
          prop: 'permissionKey',
          label: '权限键',
          minWidth: 240,
          formatter: (row: APIEndpointItem) =>
            (row.permissionKeys || []).join(', ') || row.permissionKey || '-'
        },
        {
          prop: 'permissionBindingMode',
          label: '权限结构',
          minWidth: 220,
          formatter: (row: APIEndpointItem) =>
            h('div', { class: 'permission-structure-cell' }, [
              h(
                ElTag,
                {
                  type: permissionPatternTagType(row.permissionBindingMode),
                  effect: 'plain'
                },
                () => formatPermissionPattern(row.permissionBindingMode)
              ),
              row.permissionContexts?.length
                ? h(
                    'div',
                    { class: 'permission-structure-contexts' },
                    row.permissionContexts.map((item) => formatPermissionContext(item)).join(' / ')
                  )
                : null,
              row.permissionNote
                ? h('div', { class: 'permission-structure-note' }, row.permissionNote)
                : null
            ])
        },
        {
          prop: 'contextScope',
          label: '协作空间要求',
          width: 140,
          formatter: (row: APIEndpointItem) =>
            h(
              ElSelect,
              {
                modelValue: row.contextScope || 'optional',
                size: 'small',
                onChange: (value: string) => handleContextScopeChange(row, value)
              },
              () => [
                h(ElOption, { label: '可选', value: 'optional' }),
                h(ElOption, { label: '必需', value: 'required' }),
                h(ElOption, { label: '禁止', value: 'forbidden' })
              ]
            )
        },
        {
          prop: 'source',
          label: '来源',
          width: 100,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: sourceTagType(row.source), effect: 'plain' }, () =>
              formatSource(row.source)
            )
        },
        {
          prop: 'featureKind',
          label: '功能归属',
          width: 100,
          formatter: (row: APIEndpointItem) =>
            h(
              ElTag,
              { type: row.featureKind === 'business' ? 'success' : 'info', effect: 'plain' },
              () => (row.featureKind === 'business' ? '业务' : '系统')
            )
        },
        {
          prop: 'status',
          label: '状态',
          width: 180,
          formatter: (row: APIEndpointItem) => {
            if (row.stale) {
              return h('div', { class: 'status-cell' }, [
                h(ElTag, { type: 'warning' }, () => '失效'),
                h(
                  'div',
                  { class: 'status-note' },
                  row.staleReason || '源码中已不存在该自动同步 API'
                )
              ])
            }
            return h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
          }
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operate',
          label: '操作',
          width: 70,
          fixed: 'right',
          formatter: (row: APIEndpointItem) => {
            const list: ButtonMoreItem[] = [
              { key: 'edit', label: '编辑', icon: 'ri:edit-2-line' },
              {
                key: 'add',
                label: '加入权限键',
                icon: 'ri:links-line',
                auth: 'system.permission.manage'
              },
              {
                key: 'remove',
                label: '移除权限键',
                icon: 'ri:link-unlink',
                auth: 'system.permission.manage'
              }
            ]
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleOperateCommand(row, item.key as string)
            })
          }
        }
      ]
    }
  })

  function handleOperateCommand(row: APIEndpointItem, command: string) {
    if (command === 'edit') {
      openEditDialog(row)
      return
    }
    if (command === 'add') {
      openPermissionBindDialog(row, 'add')
      return
    }
    if (command === 'remove') {
      openPermissionBindDialog(row, 'remove')
    }
  }

  const currentPermissionActionOptions = computed(() => {
    if (permissionDialogMode.value !== 'remove') {
      return permissionActionOptions.value
    }
    const boundKeys = new Set(
      (permissionBinding.endpointPermissionKeys || []).map((item) => `${item || ''}`.trim())
    )
    return permissionActionOptions.value.filter((item) =>
      boundKeys.has(`${item.permissionKey || ''}`.trim())
    )
  })

  async function loadPermissionActionOptions() {
    permissionActionLoading.value = true
    try {
      const res = await fetchGetPermissionActionOptions()
      permissionActionOptions.value = res.records || []
    } catch (error: any) {
      ElMessage.error(error?.message || '获取权限键失败')
    } finally {
      permissionActionLoading.value = false
    }
  }

  async function openPermissionBindDialog(row: APIEndpointItem, mode: 'add' | 'remove') {
    if (!hasAction('system.permission.manage')) {
      ElMessage.warning('无权限操作')
      return
    }
    if (!row.code) {
      ElMessage.warning('当前 API 缺少固定编码，请先重建 API 注册表')
      return
    }
    permissionDialogMode.value = mode
    permissionBinding.endpointCode = row.code
    permissionBinding.endpointSpec = `${row.method || ''} ${row.path || ''}`.trim()
    permissionBinding.endpointPermissionKeys = [
      ...(row.permissionKeys || (row.permissionKey ? [row.permissionKey] : []))
    ]
    permissionBinding.permissionActionId = ''
    permissionBindVisible.value = true
    await loadPermissionActionOptions()
    if (mode === 'remove' && currentPermissionActionOptions.value.length === 0) {
      ElMessage.info('当前接口没有可移除的权限键')
      permissionBindVisible.value = false
    }
  }

  async function submitPermissionBind() {
    if (!permissionBinding.endpointCode || !permissionBinding.permissionActionId) {
      ElMessage.warning('请选择权限键')
      return
    }
    try {
      if (permissionDialogMode.value === 'remove') {
        await fetchDeletePermissionActionEndpoint(
          permissionBinding.permissionActionId,
          permissionBinding.endpointCode
        )
      } else {
        await fetchAddPermissionActionEndpoint(
          permissionBinding.permissionActionId,
          permissionBinding.endpointCode
        )
      }
      ElMessage.success(permissionDialogMode.value === 'remove' ? '已移除权限键' : '已加入权限键')
      permissionBindVisible.value = false
      await refreshData()
    } catch (error: any) {
      ElMessage.error(
        error?.message ||
          (permissionDialogMode.value === 'remove' ? '移除权限键失败' : '加入权限键失败')
      )
    }
  }

  function saveTableState() {
    const payload: PersistedTableState = {
      selectedSource: selectedSource.value,
      selectedCategoryTreeKey: selectedCategoryTreeKey.value,
      tableQuery: {
        method: tableQuery.method,
        path: tableQuery.path,
        keyword: tableQuery.keyword,
        permissionKey: tableQuery.permissionKey,
        permissionPattern: tableQuery.permissionPattern,
        contextScope: tableQuery.contextScope,
        featureKind: tableQuery.featureKind,
        status: tableQuery.status,
        hasPermissionKey: tableQuery.hasPermissionKey
      }
    }
    localStorage.setItem(API_ENDPOINT_TABLE_STATE_KEY, JSON.stringify(payload))
  }

  function restoreTableState() {
    const raw = localStorage.getItem(API_ENDPOINT_TABLE_STATE_KEY)
    if (!raw) return
    try {
      const payload = JSON.parse(raw) as Partial<PersistedTableState>
      selectedSource.value = payload.selectedSource || ''
      selectedCategoryTreeKey.value = payload.selectedCategoryTreeKey || 'all'
      Object.assign(tableQuery, {
        method: payload.tableQuery?.method || '',
        path: payload.tableQuery?.path || '',
        keyword: payload.tableQuery?.keyword || '',
        permissionKey: payload.tableQuery?.permissionKey || '',
        permissionPattern: payload.tableQuery?.permissionPattern || '',
        contextScope: payload.tableQuery?.contextScope || '',
        featureKind: payload.tableQuery?.featureKind || '',
        status: payload.tableQuery?.status || '',
        hasPermissionKey: payload.tableQuery?.hasPermissionKey || ''
      })
      syncSearchFormFromQuery()
    } catch {
      localStorage.removeItem(API_ENDPOINT_TABLE_STATE_KEY)
    }
  }

  function ensureManagedAppReady(showMessage = false) {
    if (targetAppKey.value) {
      loadError.value = ''
      return true
    }
    loadError.value = managedAppMissingText
    data.value = []
    staleCandidates.value = []
    categories.value = []
    totalCount.value = 0
    noPermissionCount.value = 0
    sharedPermissionCount.value = 0
    crossContextSharedCount.value = 0
    uncategorizedCount.value = 0
    staleCount.value = 0
    unregisteredCount.value = 0
    categoryCountMap.value = {}
    if (showMessage) {
      ElMessage.warning(managedAppMissingText)
    }
    return false
  }

  function resetScopedState(message = managedAppMissingText) {
    loadError.value = message
    data.value = []
    totalCount.value = 0
    noPermissionCount.value = 0
    sharedPermissionCount.value = 0
    crossContextSharedCount.value = 0
    uncategorizedCount.value = 0
    staleCount.value = 0
    categoryCountMap.value = {}
  }

  async function loadCategories() {
    const res = await fetchGetApiEndpointCategories()
    categories.value = [...(res.records || [])]
  }

  async function handleSync() {
    syncing.value = true
    try {
      await fetchSyncApiEndpoints()
      ElMessage.success('同步成功')
      await loadUnregisteredCount()
      if (targetAppKey.value) {
        await Promise.all([refreshData(), loadCategorySummary()])
      } else {
        resetScopedState()
      }
    } catch (error: any) {
      ElMessage.error(error?.message || '同步失败')
    } finally {
      syncing.value = false
    }
  }

  async function handleCleanupStale() {
    stalePagination.current = 1
    selectedStaleIds.value = []
    try {
      await loadStaleCandidates()
    } catch (error: any) {
      ElMessage.error(error?.message || '获取失效 API 列表失败')
      return
    }
    if (!stalePagination.total) {
      ElMessage.info('当前没有可清理的失效 API')
      return
    }
    staleDialogVisible.value = true
  }

  function closeStaleDialog() {
    staleDialogVisible.value = false
    selectedStaleIds.value = []
    staleCandidates.value = []
    staleTableRef.value?.clearSelection?.()
  }

  function handleStaleSelectionChange(rows: APIEndpointItem[]) {
    const currentPageIds = new Set(staleCandidates.value.map((item) => item.id).filter(Boolean))
    const selectedSet = new Set(selectedStaleIds.value)
    currentPageIds.forEach((id) => selectedSet.delete(id))
    rows.forEach((item) => {
      if (item.id) {
        selectedSet.add(item.id)
      }
    })
    selectedStaleIds.value = Array.from(selectedSet)
  }

  async function handleStaleCurrentChange(page: number) {
    stalePagination.current = page
    await loadStaleCandidates()
  }

  async function handleStaleSizeChange(size: number) {
    stalePagination.size = size
    stalePagination.current = 1
    await loadStaleCandidates()
  }

  async function submitCleanupStale() {
    if (selectedStaleIds.value.length === 0) {
      ElMessage.warning('请先勾选要删除的失效 API')
      return
    }
    cleaningStale.value = true
    try {
      const res = await fetchCleanupStaleApiEndpoints(selectedStaleIds.value)
      closeStaleDialog()
      await loadUnregisteredCount()
      if (targetAppKey.value) {
        await Promise.all([refreshData(), loadCategorySummary()])
      } else {
        resetScopedState()
      }
      if (shouldRefreshUnregistered.value) {
        await loadUnregisteredRoutes()
      }
      ElMessage.success(`已清理 ${res.deletedCount || 0} 个失效 API`)
    } catch (error: any) {
      ElMessage.error(error?.message || '清理失效 API 失败')
    } finally {
      cleaningStale.value = false
    }
  }

  async function handleContextScopeChange(row: APIEndpointItem, value: string) {
    try {
      await fetchUpdateApiEndpointContextScope(row.id, value)
      row.contextScope = value
      ElMessage.success('协作空间要求已更新')
    } catch (error: any) {
      ElMessage.error(error?.message || '更新失败')
    }
  }

  function resetForm() {
    editingId.value = ''
    pendingLocateRoute.value = null
    formState.appScope = 'app'
    formState.method = 'GET'
    formState.path = ''
    formState.summary = ''
    formState.featureKind = 'system'
    formState.categoryId = ''
    formState.contextScope = 'optional'
    formState.source = 'manual'
    formState.status = 'normal'
    formState.permissionKeys = []
  }

  function resetCategoryForm() {
    categoryForm.id = ''
    categoryForm.code = ''
    categoryForm.name = ''
    categoryForm.nameEn = ''
    categoryForm.sortOrder = 0
    categoryForm.status = 'normal'
  }

  function openCreateDialog() {
    resetForm()
    formVisible.value = true
  }

  function resolveCategoryIdByCode(code?: string) {
    const target = `${code || ''}`.trim().toLowerCase()
    if (!target) return ''
    return (
      categories.value.find((item) => `${item.code || ''}`.trim().toLowerCase() === target)?.id ||
      ''
    )
  }

  function openEditDialog(row: APIEndpointItem) {
    editingId.value = row.id
    pendingLocateRoute.value = null
    formState.appScope = row.appScope === 'shared' ? 'shared' : 'app'
    formState.method = (row.method || 'GET').toUpperCase()
    formState.path = row.path || ''
    formState.summary = row.summary || ''
    formState.featureKind = row.featureKind || 'system'
    formState.categoryId = row.categoryId || ''
    formState.contextScope = row.contextScope || 'optional'
    formState.source = row.source || 'manual'
    formState.status = row.status || 'normal'
    formState.permissionKeys = [
      ...(row.permissionKeys || (row.permissionKey ? [row.permissionKey] : []))
    ]
    formVisible.value = true
  }

  function startCreateCategory() {
    resetCategoryForm()
  }

  function openCategoryDrawer(category?: APIEndpointCategoryItem) {
    if (category) {
      categoryForm.id = category.id || ''
      categoryForm.code = category.code || ''
      categoryForm.name = category.name || ''
      categoryForm.nameEn = category.nameEn || ''
      categoryForm.sortOrder = category.sortOrder ?? 0
      categoryForm.status = category.status || 'normal'
    } else {
      resetCategoryForm()
    }
    categoryDrawerVisible.value = true
  }

  async function handleCategoryTreeSelect(node: CategoryTreeNode) {
    selectedCategoryTreeKey.value = node.id
    if (node.type === 'uncategorized') {
      tableQuery.categoryId = ''
      tableQuery.hasCategory = 'false'
    } else if (node.type === 'category') {
      tableQuery.categoryId = node.category?.id || ''
      tableQuery.hasCategory = 'true'
    } else {
      tableQuery.categoryId = ''
      tableQuery.hasCategory = ''
    }
    await applyTableFilters()
  }

  function syncCategoryFilterFromTree() {
    if (selectedCategoryTreeKey.value === 'uncategorized') {
      tableQuery.categoryId = ''
      tableQuery.hasCategory = 'false'
      return
    }
    if (selectedCategoryTreeKey.value.startsWith('category:')) {
      tableQuery.categoryId = selectedCategoryTreeKey.value.replace(/^category:/, '')
      tableQuery.hasCategory = 'true'
      return
    }
    tableQuery.categoryId = ''
    tableQuery.hasCategory = ''
  }

  async function loadUnregisteredRoutes() {
    unregisteredLoading.value = true
    try {
      const res = await fetchGetUnregisteredApiRouteList({
        current: unregisteredPagination.current,
        size: unregisteredPagination.size,
        method: unregisteredQuery.method || undefined,
        path: unregisteredQuery.path || undefined,
        keyword: unregisteredQuery.keyword || undefined,
        only_no_meta: unregisteredQuery.onlyNoMeta || undefined
      })
      unregisteredRoutes.value = res.records || []
      unregisteredPagination.total = res.total || 0
      unregisteredCount.value = res.total || 0
    } catch (error: any) {
      ElMessage.error(error?.message || '获取未注册 API 失败')
    } finally {
      unregisteredLoading.value = false
    }
  }

  async function loadUnregisteredCount() {
    try {
      const res = await fetchGetUnregisteredApiRouteList({ current: 1, size: 1 })
      unregisteredCount.value = res.total || 0
    } catch (error: any) {
      ElMessage.error(error?.message || '获取未注册路由统计失败')
    }
  }

  async function loadStaleCandidates() {
    const res = await fetchGetStaleApiEndpointList({
      current: stalePagination.current,
      size: stalePagination.size
    })
    staleCandidates.value = res.records || []
    stalePagination.total = res.total || 0
    await nextTick()
    syncStaleSelection()
  }

  function syncStaleSelection() {
    const table = staleTableRef.value
    if (!table) return
    const selectedSet = new Set(selectedStaleIds.value)
    table.clearSelection?.()
    staleCandidates.value.forEach((item) => {
      if (selectedSet.has(item.id)) {
        table.toggleRowSelection?.(item, true)
      }
    })
  }

  async function openUnregisteredDialog() {
    unregisteredVisible.value = true
    shouldRefreshUnregistered.value = true
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  async function openScanConfigDialog() {
    scanConfigVisible.value = true
    try {
      const config = await fetchGetUnregisteredApiScanConfig()
      scanConfig.enabled = Boolean(config.enabled)
      scanConfig.frequencyMinutes = Number(config.frequencyMinutes || 60)
      scanConfig.defaultCategoryId = config.defaultCategoryId || ''
      scanConfig.defaultPermissionKey = config.defaultPermissionKey || ''
      scanConfig.markAsNoPermission = Boolean(config.markAsNoPermission)
    } catch (error: any) {
      ElMessage.error(error?.message || '获取扫描配置失败')
    }
  }

  async function saveScanConfig() {
    scanConfigSaving.value = true
    try {
      const saved = await fetchSaveUnregisteredApiScanConfig({
        enabled: scanConfig.enabled,
        frequencyMinutes: scanConfig.frequencyMinutes,
        defaultCategoryId: (scanConfig.defaultCategoryId || '').trim(),
        defaultPermissionKey: (scanConfig.defaultPermissionKey || '').trim(),
        markAsNoPermission: scanConfig.markAsNoPermission
      })
      scanConfig.enabled = Boolean(saved.enabled)
      scanConfig.frequencyMinutes = Number(saved.frequencyMinutes || 60)
      scanConfig.defaultCategoryId = saved.defaultCategoryId || ''
      scanConfig.defaultPermissionKey = saved.defaultPermissionKey || ''
      scanConfig.markAsNoPermission = Boolean(saved.markAsNoPermission)
      scanConfigVisible.value = false
      ElMessage.success('扫描配置已保存')
    } catch (error: any) {
      ElMessage.error(error?.message || '保存扫描配置失败')
    } finally {
      scanConfigSaving.value = false
    }
  }

  async function handleUnregisteredSearch() {
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  async function handleUnregisteredCurrentChange(page: number) {
    unregisteredPagination.current = page
    await loadUnregisteredRoutes()
  }

  async function handleUnregisteredSizeChange(size: number) {
    unregisteredPagination.size = size
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  async function resetUnregisteredQuery() {
    unregisteredQuery.method = ''
    unregisteredQuery.path = ''
    unregisteredQuery.keyword = ''
    unregisteredQuery.onlyNoMeta = false
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  function handleUseUnregisteredRoute(route: APIUnregisteredRouteItem) {
    resetForm()
    formState.appScope = 'app'
    formState.method = (route.method || 'GET').toUpperCase()
    formState.path = route.path || ''
    formState.summary = route.meta?.summary || ''
    formState.featureKind = route.meta?.feature_kind || 'system'
    formState.categoryId = resolveCategoryIdByCode(route.meta?.category_code)
    formState.contextScope = route.meta?.context_scope || 'optional'
    formState.source = route.meta?.source || (route.hasMeta ? 'sync' : 'manual')
    formState.permissionKeys = [...(route.meta?.permission_keys || [])]
    pendingLocateRoute.value = {
      method: formState.method,
      path: formState.path,
      source: formState.source
    }
    shouldRefreshUnregistered.value = true
    unregisteredVisible.value = false
    formVisible.value = true
    ElMessage.success('已带入新增 API 表单')
  }

  async function submitCategory() {
    categorySaving.value = true
    try {
      const payload = {
        code: categoryForm.code,
        name: categoryForm.name,
        name_en: categoryForm.nameEn,
        sort_order: categoryForm.sortOrder,
        status: categoryForm.status || 'normal'
      }
      let savedCategory: APIEndpointCategoryItem
      if (categoryForm.id) {
        savedCategory = await fetchUpdateApiEndpointCategory(categoryForm.id, payload)
      } else {
        savedCategory = await fetchCreateApiEndpointCategory(payload)
        if (formVisible.value) {
          formState.categoryId = savedCategory.id
        }
      }
      await Promise.all([loadCategories(), refreshData(), loadCategorySummary()])
      if (selectedCategoryTreeKey.value.startsWith('category:')) {
        syncCategoryFilterFromTree()
      }
      ElMessage.success('分类保存成功')
    } catch (error: any) {
      ElMessage.error(error?.message || '分类保存失败')
    } finally {
      categorySaving.value = false
    }
  }

  async function toggleCategoryStatus(category: APIEndpointCategoryItem) {
    const nextStatus = category.status === 'suspended' ? 'normal' : 'suspended'
    const actionText = nextStatus === 'normal' ? '启用' : '停用'
    try {
      await ElMessageBox.confirm(
        `${actionText}后不会删除已有 API 归属，但后续分配会按新状态执行。`,
        `确认${actionText}分类`,
        { type: 'warning' }
      )
    } catch {
      return
    }

    categorySwitchingId.value = category.id
    try {
      const payload = {
        code: category.code,
        name: category.name,
        name_en: category.nameEn,
        sort_order: category.sortOrder ?? 0,
        status: nextStatus
      }
      await fetchUpdateApiEndpointCategory(category.id, payload)
      await Promise.all([loadCategories(), refreshData(), loadCategorySummary()])
      ElMessage.success(`分类已${actionText}`)
    } catch (error: any) {
      ElMessage.error(error?.message || `分类${actionText}失败`)
    } finally {
      categorySwitchingId.value = ''
    }
  }

  async function submitForm() {
    if (!ensureManagedAppReady(true)) return
    const isEditing = !!editingId.value
    const payload = {
      app_scope: formState.appScope,
      app_key: formState.appScope === 'app' ? targetAppKey.value : '',
      method: formState.method,
      path: formState.path,
      summary: formState.summary,
      feature_kind: formState.featureKind,
      category_id: formState.categoryId || undefined,
      context_scope: formState.contextScope,
      source: formState.source,
      status: formState.status,
      permission_keys: formState.permissionKeys
    }
    saving.value = true
    try {
      if (editingId.value) {
        await fetchUpdateApiEndpoint(editingId.value, payload)
      } else {
        await fetchCreateApiEndpoint(payload)
        pendingLocateRoute.value = {
          method: payload.method,
          path: payload.path,
          source: payload.source
        }
      }
      ElMessage.success('保存成功')
      formVisible.value = false
      await refreshData()
      await loadCategorySummary()
      await loadUnregisteredCount()
      if (shouldRefreshUnregistered.value) {
        await loadUnregisteredRoutes()
      }
      if (!isEditing && pendingLocateRoute.value) {
        selectedSource.value = pendingLocateRoute.value.source || ''
        tableQuery.method = pendingLocateRoute.value.method || ''
        tableQuery.path = pendingLocateRoute.value.path || ''
        syncSearchFormFromQuery()
        await applyTableFilters()
        pendingLocateRoute.value = null
      }
    } catch (error: any) {
      ElMessage.error(error?.message || '保存失败')
    } finally {
      if (isEditing) pendingLocateRoute.value = null
      saving.value = false
    }
  }

  async function loadCategorySummary() {
    if (!ensureManagedAppReady()) return
    const res = await fetchGetApiEndpointOverview(targetAppKey.value)
    totalCount.value = res.totalCount || 0
    noPermissionCount.value = res.noPermissionCount || 0
    sharedPermissionCount.value = res.sharedPermissionCount || 0
    crossContextSharedCount.value = res.crossContextSharedCount || 0
    uncategorizedCount.value = res.uncategorizedCount || 0
    staleCount.value = res.staleCount || 0
    categoryCountMap.value = Object.fromEntries(
      (res.categoryCounts || []).map((item: any) => [item.categoryId, item.count || 0])
    )
  }

  async function applyTableFilters() {
    if (!ensureManagedAppReady()) return
    Object.assign(searchParams, {
      appKey: targetAppKey.value,
      source: selectedSource.value || undefined,
      method: tableQuery.method || undefined,
      path: tableQuery.path || undefined,
      keyword: tableQuery.keyword || undefined,
      permissionKey: tableQuery.permissionKey || undefined,
      permissionPattern: tableQuery.permissionPattern || undefined,
      categoryId: tableQuery.categoryId || undefined,
      contextScope: tableQuery.contextScope || undefined,
      featureKind: tableQuery.featureKind || undefined,
      status: tableQuery.status || undefined,
      hasPermissionKey:
        tableQuery.hasPermissionKey === '' ? undefined : tableQuery.hasPermissionKey === 'true',
      hasCategory: tableQuery.hasCategory === '' ? undefined : tableQuery.hasCategory === 'true',
      current: 1
    })
    saveTableState()
    await getData()
  }

  async function handleTableSearch() {
    syncQueryFromSearchForm()
    await applyTableFilters()
  }

  async function resetTableQuery() {
    searchForm.source = ''
    searchForm.method = ''
    searchForm.path = ''
    searchForm.keyword = ''
    searchForm.permissionKey = ''
    searchForm.permissionPattern = ''
    searchForm.contextScope = ''
    searchForm.featureKind = ''
    searchForm.status = ''
    searchForm.hasPermissionKey = ''
    syncQueryFromSearchForm()
    selectedCategoryTreeKey.value = 'all'
    tableQuery.categoryId = ''
    tableQuery.hasCategory = ''
    await applyTableFilters()
  }

  onMounted(async () => {
    restoreTableState()
    await Promise.all([loadCategories(), loadUnregisteredCount()])
    if (!targetAppKey.value) {
      resetScopedState()
      return
    }
    await loadCategorySummary()
    syncCategoryFilterFromTree()
    await applyTableFilters()
  })

  watch(
    () => targetAppKey.value,
    async () => {
      if (!targetAppKey.value) {
        resetScopedState()
        if (unregisteredVisible.value) {
          await loadUnregisteredRoutes()
        }
        if (staleDialogVisible.value) {
          await loadStaleCandidates()
        }
        return
      }
      loadError.value = ''
      await Promise.all([loadCategorySummary(), loadUnregisteredCount()])
      if (unregisteredVisible.value) {
        await loadUnregisteredRoutes()
      }
      if (staleDialogVisible.value) {
        await loadStaleCandidates()
      }
      await applyTableFilters()
    }
  )

  return {
    // 常量 / 工具
    methodOptions,
    targetAppKey,
    hasAction,
    methodTagType,
    sourceTagType,
    formatSource,
    formatPermissionPattern,
    permissionPatternTagType,
    formatPermissionContext,
    // refs
    staleTableRef,
    syncing,
    cleaningStale,
    loadError,
    showSearchBar,
    saving,
    categorySaving,
    categorySwitchingId,
    selectedSource,
    selectedCategoryTreeKey,
    formVisible,
    categoryDrawerVisible,
    permissionBindVisible,
    permissionDialogMode,
    unregisteredVisible,
    scanConfigVisible,
    scanConfigSaving,
    staleDialogVisible,
    unregisteredLoading,
    editingId,
    pendingLocateRoute,
    categories,
    permissionActionOptions,
    permissionActionLoading,
    permissionBinding,
    unregisteredRoutes,
    scanConfig,
    staleCandidates,
    selectedStaleIds,
    totalCount,
    noPermissionCount,
    sharedPermissionCount,
    crossContextSharedCount,
    staleCount,
    unregisteredCount,
    uncategorizedCount,
    categoryCountMap,
    unregisteredPagination,
    stalePagination,
    formState,
    categoryForm,
    unregisteredQuery,
    tableQuery,
    searchForm,
    // computed
    sortedCategories,
    categoryTreeData,
    summaryMetrics,
    currentPermissionActionOptions,
    // useTable
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    // handlers
    handleOperateCommand,
    loadPermissionActionOptions,
    openPermissionBindDialog,
    submitPermissionBind,
    saveTableState,
    restoreTableState,
    ensureManagedAppReady,
    resetScopedState,
    loadCategories,
    handleSync,
    handleCleanupStale,
    closeStaleDialog,
    handleStaleSelectionChange,
    handleStaleCurrentChange,
    handleStaleSizeChange,
    submitCleanupStale,
    handleContextScopeChange,
    resetForm,
    resetCategoryForm,
    openCreateDialog,
    resolveCategoryIdByCode,
    openEditDialog,
    startCreateCategory,
    openCategoryDrawer,
    handleCategoryTreeSelect,
    syncCategoryFilterFromTree,
    loadUnregisteredRoutes,
    loadUnregisteredCount,
    loadStaleCandidates,
    syncStaleSelection,
    openUnregisteredDialog,
    openScanConfigDialog,
    saveScanConfig,
    handleUnregisteredSearch,
    handleUnregisteredCurrentChange,
    handleUnregisteredSizeChange,
    resetUnregisteredQuery,
    handleUseUnregisteredRoute,
    submitCategory,
    toggleCategoryStatus,
    submitForm,
    loadCategorySummary,
    applyTableFilters,
    handleTableSearch,
    resetTableQuery
  }
}
