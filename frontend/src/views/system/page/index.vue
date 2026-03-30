<template>
  <div class="page-management-page art-full-height">
    <ArtSearchBar
      class="page-search-bar"
      v-show="showSearchBar"
      v-model="searchForm"
      :items="searchItems"
      :span="4"
      label-width="72px"
      :showExpand="true"
      :defaultExpanded="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard
      class="art-table-card"
      shadow="never"
      :style="{ marginTop: showSearchBar ? '12px' : '0' }"
    >
      <ElAlert
        v-if="loadError"
        class="page-inline-alert"
        type="info"
        :closable="false"
        show-icon
        :title="loadError"
      />
      <ArtTableHeader
        :loading="loading"
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        @refresh="handleRefresh"
      >
        <template #left>
          <div class="page-toolbar">
            <div class="page-toolbar-head">
              <div class="page-toolbar-title-row">
                <div class="page-toolbar-title">页面管理</div>
                <div class="page-toolbar-subtitle">独立页面、挂接页面、逻辑分组与普通分组统一查看</div>
                <div class="page-toolbar-metrics">
                  <span v-for="item in summaryStats" :key="item.label" class="page-toolbar-metric">
                    {{ item.label }} {{ item.value }}
                  </span>
                </div>
              </div>
            </div>
            <div class="page-toolbar-actions">
              <ElSelect
                v-model="activeSpaceKey"
                class="page-space-select"
                filterable
                @change="handleRefresh"
              >
                <ElOption
                  v-for="item in menuSpaceOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
              <ElDropdown trigger="click" @command="handleCreateCommand">
                <ElButton v-action="'system.page.manage'" type="primary" v-ripple>
                  新增
                </ElButton>
                <template #dropdown>
                  <ElDropdownMenu>
                    <ElDropdownItem command="page">新增页面</ElDropdownItem>
                    <ElDropdownItem command="group">新增逻辑分组</ElDropdownItem>
                    <ElDropdownItem command="display_group">新增普通分组</ElDropdownItem>
                  </ElDropdownMenu>
                </template>
              </ElDropdown>
              <ElButton v-action="'system.page.sync'" @click="unregisteredDialogVisible = true" v-ripple>
                未注册页面
              </ElButton>
              <div class="page-switch">
                <span class="page-switch__label">展开分组</span>
                <ElSwitch v-model="isExpanded" @change="handleExpandSwitchChange" />
              </div>
              <div class="page-switch">
                <span class="page-switch__label">显示停用</span>
                <ElSwitch v-model="showSuspended" />
              </div>
            </div>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        :rowKey="rowKey"
        :loading="loading"
        :data="tableData"
        :columns="displayColumns"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="false"
      >
        <template #name="{ row }">
          <div
            :class="[
              'page-name-cell',
              {
                'page-name-cell--logic-group': row.pageType === 'group',
                'page-name-cell--display-group': row.pageType === 'display_group'
              }
            ]"
          >
            <div class="page-name-cell__main">
              <div class="page-name-cell__title">
                <ElTag :type="getPageTypeTag(row)" effect="plain" size="small">
                  {{ getPageTypeText(row) }}
                </ElTag>
                <span class="page-name-cell__text">{{ row.name }}</span>
                <span class="page-inline-relation">{{ getRelationDisplayText(row) }}</span>
              </div>
            </div>
          </div>
        </template>

        <template #route="{ row }">
          <div class="page-route-cell">
            <code :class="['page-route-text', { 'page-muted-text': !getRouteDisplayText(row) }]">
              {{ getRouteDisplayText(row) || '-' }}
            </code>
          </div>
        </template>

        <template #component="{ row }">
          <div class="page-component-cell">
            <span class="page-muted-text">
              {{ row.pageType === 'group' || row.pageType === 'display_group' ? '不需要组件' : row.component || '-' }}
            </span>
          </div>
        </template>

        <template #sortOrder="{ row }">
          <div class="page-sort-cell">
            <template v-if="editingSortId === row.id">
              <ElInput
                v-model="sortDraftMap[row.id]"
                size="small"
                class="page-sort-input"
                inputmode="numeric"
              />
              <div class="page-sort-actions">
                <ElButton
                  type="primary"
                  link
                  size="small"
                  :loading="savingSortIds.has(row.id)"
                  @click="saveSortOrder(row)"
                >
                  保存
                </ElButton>
                <ElButton link size="small" @click="cancelSortEdit(row)">取消</ElButton>
              </div>
            </template>
            <template v-else>
              <div class="page-sort-view">
                <span class="page-sort-value">{{ row.sortOrder ?? 0 }}</span>
                <ElButton type="primary" link size="small" class="page-sort-edit-btn" @click="startSortEdit(row)">
                  编辑
                </ElButton>
              </div>
            </template>
          </div>
        </template>

        <template #accessMode="{ row }">
          <div class="page-access-cell">
            <ElTag effect="plain" size="small" type="info">
              {{ getMountModeText(row) }}
            </ElTag>
            <ElTag v-if="row.pageType !== 'display_group'" :effect="'plain'" :type="getAccessModeTag(row.accessMode)">
              {{ getAccessModeText(row.accessMode) }}
            </ElTag>
            <span v-else class="page-muted-text">-</span>
            <ElTag :type="row.status === 'normal' ? 'success' : 'danger'" effect="light">
              {{ row.status === 'normal' ? '正常' : '停用' }}
            </ElTag>
          </div>
        </template>

        <template #effectiveChain="{ row }">
          <span class="page-muted-text">{{ getEffectiveChainText(row) }}</span>
        </template>

        <template #space="{ row }">
          <ElTag size="small" effect="plain" type="info">
            {{ getSpaceName(row.spaceKey) }}
          </ElTag>
        </template>

        <template #mountTarget="{ row }">
          <span class="page-muted-text">{{ getMountTargetText(row) }}</span>
        </template>

        <template #parentChainStatus="{ row }">
          <span
            :class="[
              'page-muted-text',
              { 'page-chain-status--error': getParentChainStatusText(row).startsWith('异常') }
            ]"
          >
            {{ getParentChainStatusText(row) }}
          </span>
        </template>

        <template #updatedAt="{ row }">
          <span class="page-muted-text">{{ formatUpdatedAt(row.updatedAt) }}</span>
        </template>

        <template #operation="{ row }">
          <div class="flex items-center justify-center gap-2">
            <ArtButtonMore :list="getOperationList(row)" @click="(item) => handleOperation(item, row)" />
          </div>
        </template>
      </ArtTable>
    </ElCard>

    <PageDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :page-data="currentPage"
      :default-data="defaultPageData"
      :menu-spaces="menuSpaces"
      :current-space-key="activeSpaceKey"
      :initial-parent-page-key="initialParentPageKey"
      :initial-parent-menu-id="initialParentMenuId"
      :initial-page-type="initialPageType"
      @success="handleRefresh"
    />
    <PageUnregisteredDialog
      v-model="unregisteredDialogVisible"
      @synced="handleRefresh"
      @create-candidate="handleCreateFromCandidate"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, nextTick, onMounted, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
    import { ElButton, ElInput, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
    import { fetchDeletePage, fetchGetMenuSpaces, fetchGetPageList, fetchGetPageMenuOptions, fetchSyncPages, fetchUpdatePage } from '@/api/system-manage'
  import { joinManagedPagePath, resolveManagedPageRoutePath } from '@/utils/navigation/managed-page'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import PageDialog from './modules/page-dialog.vue'
  import PageUnregisteredDialog from './modules/page-unregistered-dialog.vue'

  defineOptions({ name: 'PageManagement' })

  type PageItem = Api.SystemManage.PageItem
  type TreePageItem = PageItem & { children?: TreePageItem[] }

  const loading = ref(false)
  const loadError = ref('')
  const showSearchBar = ref(false)
  const isExpanded = ref(false)
  const syncing = ref(false)
  const showSuspended = ref(true)
  const sortDraftMap = reactive<Record<string, string>>({})
  const savingSortIds = ref(new Set<string>())
  const editingSortId = ref('')
  const tableRef = ref()
  const rawPages = ref<PageItem[]>([])
  const menuPathMap = ref(new Map<string, string>())
  const menuSpaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const activeSpaceKey = ref('default')
  const route = useRoute()
  const router = useRouter()
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
  status: ''
}
  const searchForm = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })

  const searchItems = computed<FormItem[]>(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: { placeholder: '名称/标识/路由/组件/上级' }
    },
    {
      label: '页面类型',
      key: 'pageType',
      type: 'select',
      span: 4,
      props: {
        options: [
          { label: '全部', value: '' },
          { label: '逻辑分组', value: 'group' },
          { label: '普通分组', value: 'display_group' },
          { label: '内页', value: 'inner' },
          { label: '全局页', value: 'global' }
        ],
        clearable: true
      }
    },
    {
      label: '访问模式',
      key: 'accessMode',
      type: 'select',
      span: 4,
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
      label: '状态',
      key: 'status',
      type: 'select',
      span: 4,
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
    { prop: 'space', label: '菜单空间', width: 120, useSlot: true, slotName: 'space' },
    { prop: 'mountTarget', label: '挂接对象', minWidth: 180, useSlot: true, slotName: 'mountTarget' },
    { prop: 'effectiveChain', label: '生效链路', minWidth: 150, useSlot: true, slotName: 'effectiveChain' },
    { prop: 'parentChainStatus', label: '父链状态', minWidth: 130, useSlot: true, slotName: 'parentChainStatus' },
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

  const normalizeKeyword = (value?: string) => `${value || ''}`.trim().toLowerCase()

  const comparePages = (left: PageItem, right: PageItem) => {
    const sortDiff = Number(left.sortOrder || 0) - Number(right.sortOrder || 0)
    if (sortDiff !== 0) return sortDiff
    return `${left.name || ''}${left.pageKey || ''}`.localeCompare(`${right.name || ''}${right.pageKey || ''}`, 'zh-Hans-CN')
  }

  const pageTree = computed<TreePageItem[]>(() => buildPageTree(rawPages.value))
  const tableData = computed<TreePageItem[]>(() => filterPageTree(pageTree.value))
  const menuSpaceMap = computed(() => new Map(menuSpaces.value.map((item) => [item.spaceKey, item])))
  const menuSpaceOptions = computed(() =>
    menuSpaces.value.map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )
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
    const independentCount = rawPages.value.filter((item) => {
      if (item.pageType === 'display_group') return false
      return !`${item.parentMenuId || ''}`.trim() && !`${item.parentPageKey || ''}`.trim()
    }).length
    return [
      { label: '当前显示', value: visibleCount.value },
      { label: '独立页面', value: independentCount || 0 },
      { label: '停用', value: suspendedCount || 0 },
      { label: '总条目', value: rawPages.value.length }
    ]
  })

  function buildPageTree(items: PageItem[]): TreePageItem[] {
    const logicNodeMap = new Map<string, TreePageItem>()
    const displayGroupMap = new Map<string, TreePageItem>()
    const childrenMap = new Map<string, TreePageItem[]>()
    const roots: TreePageItem[] = []

    items.forEach((item) => {
      const node = { ...item, children: [] }
      if (item.pageType === 'display_group') {
        displayGroupMap.set(item.pageKey, node)
        return
      }
      logicNodeMap.set(item.pageKey, node)
    })

    Array.from(logicNodeMap.values()).forEach((item) => {
      const parentKey = `${item.parentPageKey || ''}`.trim()
      if (!parentKey || !logicNodeMap.has(parentKey)) {
        roots.push(item)
        return
      }
      const children = childrenMap.get(parentKey) || []
      children.push(item)
      childrenMap.set(parentKey, children)
    })

    const attachChildren = (node: TreePageItem) => {
      const children = (childrenMap.get(node.pageKey) || []).sort(comparePages)
      node.children = children.map((child) => attachChildren(child))
      return node
    }

    const resolvedRoots = roots.sort(comparePages).map((item) => attachChildren(item))
    const ungroupedRoots: TreePageItem[] = []
    resolvedRoots.forEach((item) => {
      const displayGroupKey = `${item.displayGroupKey || ''}`.trim()
      const displayGroup = displayGroupKey ? displayGroupMap.get(displayGroupKey) : undefined
      if (!displayGroup) {
        ungroupedRoots.push(item)
        return
      }
      const groupChildren = displayGroup.children || []
      groupChildren.push(item)
      displayGroup.children = groupChildren.sort(comparePages)
    })

    const groupedRoots = Array.from(displayGroupMap.values()).sort(comparePages)
    return [...groupedRoots, ...ungroupedRoots].sort(comparePages)
  }

  function matchesPageSearch(item: PageItem) {
    if (!showSuspended.value && item.status !== 'normal') return false
    const keyword = normalizeKeyword(appliedFilters.keyword)
    const pageType = `${appliedFilters.pageType || ''}`.trim()
    const accessMode = `${appliedFilters.accessMode || ''}`.trim()
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

  function countTreeNodes(items: TreePageItem[]): number {
    return items.reduce((total, item) => total + 1 + countTreeNodes(item.children || []), 0)
  }

  function buildMenuPathMap(items: Api.SystemManage.PageMenuOptionItem[]) {
    const nextMap = new Map<string, string>()
    const walk = (nodes: Api.SystemManage.PageMenuOptionItem[], parentPath = '') => {
      nodes.forEach((item) => {
        const fullPath = joinManagedPagePath(parentPath, item.path)
        if (item.id) {
          nextMap.set(item.id, fullPath)
        }
        if (Array.isArray(item.children) && item.children.length) {
          walk(item.children, fullPath)
        }
      })
    }
    walk(items)
    return nextMap
  }

  async function getPageList() {
    loading.value = true
    loadError.value = ''
    try {
      const [pageResponse, menuResponse] = await Promise.all([
        fetchGetPageList({ current: 1, size: 1000, spaceKey: activeSpaceKey.value }),
        fetchGetPageMenuOptions(activeSpaceKey.value)
        ])
        rawPages.value = pageResponse.records || []
        Object.keys(sortDraftMap).forEach((key) => delete sortDraftMap[key])
        editingSortId.value = ''
        rawPages.value.forEach((item) => {
          if (item.id) {
            sortDraftMap[item.id] = `${Number(item.sortOrder || 0)}`
          }
        })
        menuPathMap.value = buildMenuPathMap(menuResponse.records || [])
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
      status: searchForm.status
    })
  }

  function handleReset() {
    Object.assign(searchForm, initialSearchState)
    Object.assign(appliedFilters, initialSearchState)
  }

  function handleRefresh() {
    getPageList()
  }

  function syncRouteSpaceKey(spaceKey: string) {
    router.replace({
      query: {
        ...route.query,
        spaceKey
      }
    })
  }

  function resolveInitialSpaceKey() {
    const requestedSpaceKey = `${route.query.spaceKey || ''}`.trim()
    if (requestedSpaceKey && menuSpaces.value.some((item) => item.spaceKey === requestedSpaceKey)) {
      return requestedSpaceKey
    }
    return menuSpaces.value.find((item) => item.isDefault)?.spaceKey || 'default'
  }

  function getSpaceName(spaceKey?: string) {
    const normalized = `${spaceKey || ''}`.trim() || 'default'
    return menuSpaceMap.value.get(normalized)?.name || normalized
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
      spaceKey: type === 'edit' && row ? row.spaceKey : activeSpaceKey.value,
      ...(options?.defaultData ? { ...options.defaultData } : {})
    }
    initialParentPageKey.value = options?.parentPageKey || ''
    initialParentMenuId.value = options?.parentMenuId || ''
    initialPageType.value = options?.pageType || 'inner'
    await nextTick()
    dialogVisible.value = true
  }

  function handleAddPage() {
    openDialog('add', undefined, { pageType: 'inner' })
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
      parentPageKey: options?.parentPageKey ?? (row.pageType === 'display_group' ? '' : row.pageKey),
      parentMenuId: options?.parentMenuId ?? (row.pageType === 'display_group' ? '' : row.parentMenuId || ''),
      defaultData: options?.displayGroupKey ? { displayGroupKey: options.displayGroupKey } : undefined
    })
  }

  function buildCopyPageData(row: PageItem): Partial<PageItem> {
    const routeNameBase = `${row.routeName || row.pageKey || ''}`.trim()
    const routePathBase = `${row.routePath || ''}`.trim()
    return {
      ...row,
      id: '',
      name: `${row.name || '页面'} 副本`,
      pageKey: row.pageKey ? `${row.pageKey}.copy` : '',
      routeName: routeNameBase ? `${routeNameBase}Copy` : '',
      routePath: routePathBase,
      source: 'manual'
    }
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

  function getOperationList(row: PageItem): ButtonMoreItem[] {
    if (row.pageType === 'display_group') {
      return [
        { key: 'add-group', label: '新增组内逻辑分组', icon: 'ri:folder-add-line', auth: 'system.page.manage' },
        { key: 'add-page', label: '新增组内页面', icon: 'ri:file-add-line', auth: 'system.page.manage' },
        { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'system.page.manage' },
        { key: 'delete', label: '删除', icon: 'ri:delete-bin-4-line', auth: 'system.page.manage', color: '#f56c6c' }
      ]
    }
    const list: ButtonMoreItem[] = [
      { key: 'add-group', label: '新增子逻辑分组', icon: 'ri:folder-add-line', auth: 'system.page.manage' },
      { key: 'add-page', label: '新增子页面', icon: 'ri:file-add-line', auth: 'system.page.manage' },
      { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'system.page.manage' },
      { key: 'delete', label: '删除', icon: 'ri:delete-bin-4-line', auth: 'system.page.manage', color: '#f56c6c' }
    ]
    if (row.pageType === 'inner' || row.pageType === 'global') {
      list.splice(3, 0, { key: 'copy', label: '复制页面', icon: 'ri:file-copy-line', auth: 'system.page.manage' })
      list.splice(3, 0, { key: 'visit', label: '访问', icon: 'ri:external-link-line' })
    }
    return list
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

  function getRelationDisplayText(row: PageItem) {
    if (row.pageType === 'display_group') {
      return '仅列表归类'
    }
    if (row.pageType === 'global') {
      return '独立页面'
    }
    const parentPageName = `${row.parentPageName || ''}`.trim()
    if (parentPageName) {
      return `挂到页面 · ${parentPageName}`
    }
    const parentMenuName = `${row.parentMenuName || ''}`.trim()
    if (parentMenuName) {
      return `挂到菜单 · ${parentMenuName}`
    }
    const displayGroupName = `${row.displayGroupName || ''}`.trim()
    if (displayGroupName) {
      return `普通分组：${displayGroupName}`
    }
    return '独立内页'
  }

  function getMountTargetText(row: PageItem) {
    if (row.pageType === 'display_group') return '普通分组'
    if (row.pageType === 'group') {
      return row.displayGroupName ? `逻辑分组 · ${row.displayGroupName}` : '逻辑分组'
    }
    if (row.parentMenuName) return `挂到菜单 · ${row.parentMenuName}`
    if (row.parentPageName) return `挂到页面 · ${row.parentPageName}`
    if (row.displayGroupName) return `列表分组 · ${row.displayGroupName}`
    return row.pageType === 'global' ? '独立页面' : '独立内页'
  }

  function formatUpdatedAt(value?: string) {
    const target = `${value || ''}`.trim()
    if (!target) {
      return '-'
    }
    const date = new Date(target)
    if (Number.isNaN(date.getTime())) {
      return target
    }
    const year = date.getFullYear()
    const month = `${date.getMonth() + 1}`.padStart(2, '0')
    const day = `${date.getDate()}`.padStart(2, '0')
    const hour = `${date.getHours()}`.padStart(2, '0')
    const minute = `${date.getMinutes()}`.padStart(2, '0')
    return `${year}-${month}-${day} ${hour}:${minute}`
  }

  function resolveVisitTarget(row: PageItem): string {
    const link = `${row.link || ''}`.trim()
    if (link) return link
    const resolvedPath = getResolvedRoutePath(row)
    const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
      resolvedPath,
      row.spaceKey || activeSpaceKey.value
    )
    if (nextTarget.mode === 'location') {
      return nextTarget.target
    }
    const pathname = `${window.location.pathname || '/'}`.replace(/\/?$/, '/')
    return `${window.location.origin}${pathname}#${nextTarget.target}`
  }

  function handleVisit(row: PageItem) {
    const target = resolveVisitTarget(row)
    if (!target) {
      ElMessage.warning('当前页面未配置可访问地址')
      return
    }
    window.open(target, '_blank')
  }

  async function handleSyncPages() {
    if (syncing.value) return
    syncing.value = true
    try {
      const res = await fetchSyncPages()
      ElMessage.success(`同步完成：新增 ${res.createdCount}，跳过 ${res.skippedCount}`)
      await getPageList()
    } catch (error: any) {
      ElMessage.error(error?.message || '同步页面失败')
    } finally {
      syncing.value = false
    }
  }

  function handleCreateFromCandidate(candidate: Partial<PageItem>) {
    openDialog('add', undefined, {
      parentPageKey: candidate.parentPageKey || '',
      parentMenuId: candidate.parentMenuId || '',
      pageType: candidate.pageType || 'inner',
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
      await fetchDeletePage(row.id)
      ElMessage.success('删除成功')
      getPageList()
    } catch (error: any) {
      if (error === 'cancel' || error === 'close') return
      ElMessage.error(error?.message || '删除失败')
    }
  }

  function getPageTypeText(row: PageItem) {
    if (row.pageType === 'group') return '逻辑分组'
    if (row.pageType === 'display_group') return '普通分组'
    if (row.pageType === 'global') return '全局页'
    return '内页'
  }

  function getPageTypeTag(row: PageItem) {
    if (row.pageType === 'group') return 'info'
    if (row.pageType === 'display_group') return 'success'
    if (row.pageType === 'global') return 'primary'
    return 'warning'
  }

  function getAccessModeText(accessMode?: string) {
    const accessModeTextMap: Record<string, string> = {
      inherit: '继承',
      public: '公开',
      jwt: '登录',
      permission: '权限'
    }
    return accessModeTextMap[accessMode || 'inherit'] || accessMode || '-'
  }

  function getAccessModeTag(accessMode?: string) {
    const tagMap: Record<string, 'primary' | 'success' | 'info' | 'warning' | 'danger'> = {
      inherit: 'info',
      public: 'success',
      jwt: 'warning',
      permission: 'danger'
    }
    return tagMap[accessMode || 'inherit'] || 'info'
  }

  function getMountModeText(row: PageItem) {
    if (row.parentPageKey) return '挂到页面'
    if (row.parentMenuId) return '挂到菜单'
    return row.pageType === 'global' ? '独立页面' : '独立内页'
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
    if (row.accessMode && row.accessMode !== 'inherit') return `自身（${getAccessModeText(row.accessMode)}）`
    if (row.parentPageKey) {
      const parent = pageMap.value.get(row.parentPageKey)
      return parent ? `继承页面（${parent.name}）` : '继承页面（父级缺失）'
    }
    return '自身生效'
  }

  function getParentChainStatusText(row: PageItem) {
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

  function toPageSaveParams(row: PageItem, nextSortOrder: number): Api.SystemManage.PageSaveParams {
    return {
      page_key: row.pageKey,
      name: row.name,
      route_name: row.routeName || row.pageKey,
      route_path: row.routePath || '',
      component: row.component || '',
      page_type: row.pageType,
      source: row.source || 'manual',
      module_key: row.moduleKey || '',
      sort_order: nextSortOrder,
      parent_menu_id: row.parentMenuId || '',
      parent_page_key: row.parentPageKey || '',
      display_group_key: row.displayGroupKey || '',
      active_menu_path: row.activeMenuPath || '',
      breadcrumb_mode: row.breadcrumbMode || 'inherit_menu',
      access_mode: row.accessMode || 'inherit',
      permission_key: row.permissionKey || '',
      keep_alive: Boolean(row.keepAlive),
      is_full_page: Boolean(row.isFullPage),
      status: row.status || 'normal',
      meta: {
        ...(row.meta || {}),
        isIframe: Boolean(row.isIframe),
        isHideTab: Boolean(row.isHideTab),
        link: row.link || ''
      }
    }
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
      await fetchUpdatePage(row.id, toPageSaveParams(row, nextSortOrder))
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
    fetchGetMenuSpaces()
      .then((res) => {
        menuSpaces.value = res.records || []
        activeSpaceKey.value = resolveInitialSpaceKey()
      })
      .finally(() => {
        getPageList()
      })
  })

  watch(
    () => route.query.spaceKey,
    (value) => {
      const requestedSpaceKey = `${value || ''}`.trim()
      if (!requestedSpaceKey || requestedSpaceKey === activeSpaceKey.value) {
        return
      }
      if (!menuSpaces.value.some((item) => item.spaceKey === requestedSpaceKey)) {
        return
      }
      activeSpaceKey.value = requestedSpaceKey
      getPageList()
    }
  )

  watch(activeSpaceKey, (value, previousValue) => {
    if (!`${value || ''}`.trim() || value === previousValue) {
      return
    }
    syncRouteSpaceKey(value)
  })
</script>

<style lang="scss" scoped>
  .page-toolbar {
    display: flex;
    flex-direction: column;
    gap: 14px;
    width: 100%;
  }

  .page-inline-alert {
    margin-bottom: 14px;
  }

  .page-toolbar-head {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    justify-content: flex-start;
    width: 100%;
  }

  .page-toolbar-title-row {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .page-toolbar-title {
    color: var(--el-text-color-primary);
    font-size: 18px;
    font-weight: 700;
    line-height: 1.2;
  }

  .page-toolbar-subtitle {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.4;
  }

  .page-toolbar-metrics {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
  }

  .page-toolbar-metric {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.2;
  }

  .page-toolbar-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: flex-start;
  }

  .page-space-select {
    width: 220px;
  }

  .page-switch {
    align-items: center;
    display: inline-flex;
    gap: 6px;
    margin-left: 4px;
  }

  .page-switch__label {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1;
    white-space: nowrap;
  }

  :deep(.page-search-bar .el-form-item__label) {
    white-space: nowrap;
  }

  .page-name-cell {
    display: flex;
    align-items: flex-start;
    flex: 1;
    min-width: 0;
  }

  .page-name-cell--logic-group {
    color: var(--el-text-color-primary);
    font-weight: 600;
  }

  .page-name-cell--display-group {
    color: color-mix(in srgb, var(--el-color-success-dark-2) 72%, black);
    font-weight: 600;
  }

  .page-name-cell__main {
    display: flex;
    flex: 1;
    align-items: center;
    min-width: 0;
  }

  .page-name-cell__title {
    align-items: center;
    display: flex;
    flex-wrap: nowrap;
    gap: 8px;
    min-width: 0;
  }

  .page-name-cell__text {
    color: var(--el-text-color-primary);
    font-size: 14px;
    font-weight: 600;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-inline-relation {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    margin-left: 8px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-muted-text {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.4;
  }

  .page-route-cell,
  .page-component-cell {
    display: flex;
    align-items: center;
    min-width: 0;
  }

  .page-route-text {
    color: var(--el-text-color-primary);
    display: inline-block;
    font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
    font-size: 12px;
    line-height: 1.5;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-access-cell {
    display: inline-flex;
    align-items: center;
    flex-wrap: nowrap;
    gap: 6px;
    justify-content: center;
    white-space: nowrap;
  }

  .page-chain-status--error {
    color: var(--el-color-danger);
  }

  .page-sort-cell {
    align-items: center;
    display: inline-flex;
    gap: 4px;
    justify-content: center;
    white-space: nowrap;
    width: 100%;
  }

  .page-sort-input {
    width: 84px;
  }

  .page-sort-value {
    color: var(--el-text-color-primary);
    font-variant-numeric: tabular-nums;
    min-width: 24px;
    text-align: center;
  }

  .page-sort-view {
    align-items: center;
    display: flex;
    justify-content: center;
    position: relative;
    width: 100%;
  }

  :deep(.el-table .el-table__body .el-table__cell .page-sort-cell) {
    margin: 0 auto;
  }

  :deep(.page-sort-input .el-input__wrapper) {
    padding-left: 8px;
    padding-right: 8px;
  }

  .page-sort-actions {
    align-items: center;
    display: inline-flex;
    gap: 2px;
  }

  :deep(.page-sort-actions .el-button--small.is-link) {
    margin-left: 0;
    padding-left: 2px;
    padding-right: 2px;
  }

  .page-sort-edit-btn {
    position: absolute;
    right: 0;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.15s ease;
  }

  :deep(.el-table__body tr:hover .page-sort-edit-btn) {
    opacity: 1;
    pointer-events: auto;
  }

  :deep(.el-table .el-table__body .el-table__cell:nth-child(1) .cell) {
    display: flex;
    align-items: center;
  }

  :deep(.el-table .el-table__body .el-table__cell:nth-child(1) .el-table__expand-icon) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    margin-right: 8px;
    color: var(--el-text-color-regular);
    font-size: 18px;
    transform: scale(1.15);
    transform-origin: center;
  }

  :deep(.el-table .el-table__body tr:has(.page-name-cell--logic-group)) {
    background: color-mix(in srgb, var(--el-color-primary-light-9) 45%, white);
  }

  :deep(.el-table .el-table__body tr:has(.page-name-cell--display-group)) {
    background: color-mix(in srgb, var(--el-color-success-light-9) 65%, white);
  }

  @media (max-width: 960px) {
    .page-toolbar-head {
      flex-direction: column;
    }

    .page-toolbar-actions {
      justify-content: flex-start;
      margin-left: 0;
    }
  }
</style>
