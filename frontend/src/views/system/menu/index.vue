<!-- 菜单管理页面 -->
<template>
  <div class="menu-page art-full-height">
    <!-- 搜索栏 -->
    <ArtSearchBar
      v-model="formFilters"
      :items="formItems"
      :showExpand="false"
      @reset="handleReset"
      @search="handleSearch"
    />

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader
        :showZebra="false"
        :loading="loading"
        v-model:columns="columnChecks"
        @refresh="handleRefresh"
      >
        <template #left>
          <ElTooltip content="添加菜单" placement="top">
            <div v-auth="'add'" class="inline-block cursor-pointer" @click="handleAddMenu">
              <ArtButtonTable type="add" />
            </div>
          </ElTooltip>
          <ElButton @click="toggleExpand" v-ripple>
            {{ isExpanded ? '收起' : '展开' }}
          </ElButton>
          <ElTooltip content="内页默认不显示在侧栏，仅通过按钮跳转；开启后可在列表中查看内页项" placement="top">
            <span class="inline-flex items-center gap-2 ml-2">
              <span class="text-sm text-gray-600">显示内页</span>
              <ElSwitch v-model="showInnerPages" />
            </span>
          </ElTooltip>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        :rowKey="rowKey"
        :loading="loading"
        :columns="displayColumns"
        :data="draggableData"
        :stripe="false"
        :row-class-name="getRowClassName"
        :indent="0"
      >
        <!-- 菜单名称列 -->
        <template #title="{ row }">
          <div :style="{ paddingLeft: `${(row._level || 0) * 20}px` }" class="flex items-center">
            <template v-if="row.children?.length > 0">
              <div
                class="mr-1 cursor-pointer hover:text-theme"
                @click.stop="handleExpandChange(row, !isRowExpanded(String(row.id || row.path)))"
              >
                <ArtSvgIcon
                  :icon="isRowExpanded(String(row.id || row.path)) ? 'ri:arrow-down-s-line' : 'ri:arrow-right-s-line'"
                />
              </div>
            </template>
            <span v-else class="w-4 mr-1 inline-block"></span>
            <ArtSvgIcon :icon="row.meta?.icon || 'ri:menu-line'" class="mr-2 text-g-500" />
            <span>{{ formatMenuTitle(row.meta?.title) }}</span>
          </div>
        </template>

        <!-- 菜单类型列 -->
        <template #type="{ row }">
          <ElTag :type="getMenuTypeTag(row)">{{ getMenuTypeText(row) }}</ElTag>
        </template>

        <!-- 路由列 -->
        <template #path="{ row }">
          <span>{{ row.meta?.isAuthButton ? '' : (row.meta?.link || row.path || '') }}</span>
        </template>

        <!-- 状态列 -->
        <template #status>
          <ElTag type="success">启用</ElTag>
        </template>

        <!-- 操作列 -->
        <template #operation="{ row }">
          <div class="flex items-center justify-center gap-2">
            <ArtButtonMore 
              :list="getOperationList(row)" 
              @click="(item) => handleMenuOperation(item, row)" 
            />
            <!-- 拖拽手柄移动到操作项后面 -->
            <div class="drag-handle cursor-move flex-cc hover:text-theme transition-colors">
              <ArtSvgIcon icon="ri:drag-move-2-fill" class="text-g-400 text-lg" />
            </div>
          </div>
        </template>
      </ArtTable>

      <!-- 菜单弹窗 -->
      <MenuDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :editData="editData"
        :menuTree="tableData"
        :editingMenuId="editData?.id"
        :initialParentId="String(parentRowForAdd?.id ?? '')"
        :lockType="lockMenuType"
        @submit="handleSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref, reactive, computed, watch, nextTick } from 'vue'
  import { formatMenuTitle } from '@/utils/router'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import type { AppRouteRecord } from '@/types/router'
  import MenuDialog from './modules/menu-dialog.vue'
  import { asyncRoutes } from '@/router/routes/asyncRoutes'
  import {
    fetchGetMenuTreeAll,
    fetchCreateMenu,
    fetchUpdateMenu,
    fetchDeleteMenu,
    fetchUpdateMenuSortByParent
  } from '@/api/system-manage'
  import { ElTag, ElMessageBox, ElMessage, ElTooltip, ElButton, ElSwitch } from 'element-plus'
  import { useDraggable } from 'vue-draggable-plus'

  defineOptions({ name: 'Menus' })

  // --- 状态管理 ---
  const loading = ref(false)
  const isExpanded = ref(false)
  const showInnerPages = ref(false)
  const tableRef = ref()
  const tableData = ref<AppRouteRecord[]>([])
  const expandedRowKeys = ref<string[]>([])
  const dataFromBackend = ref(false)

  // --- 搜索相关 ---
  const initialSearchState = { name: '', route: '' }
  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })
  const formItems = computed(() => [
    { label: '菜单名称', key: 'name', type: 'input', props: { clearable: true } },
    { label: '路由地址', key: 'route', type: 'input', props: { clearable: true } }
  ])

  // --- 弹窗相关 ---
  const dialogVisible = ref(false)
  const dialogType = ref<'menu' | 'button'>('menu')
  const editData = ref<any>(null)
  const parentRowForAdd = ref<AppRouteRecord | null>(null)
  const lockMenuType = ref(false)

  // --- 拖拽相关 ---
  const draggableInstance = ref<any>(null)
  const draggableData = ref<any[]>([])
  const buildRowClassId = (id: string) => encodeURIComponent(id).replace(/%/g, '_')

  const getRowClassName = ({ row }: { row: any }) => {
    return `menu-row level-${row._level} parent-${row._parentId || 'root'} menu-id-${row._rowClassId}`
  }

  const initDraggable = () => {
    const el = tableRef.value?.elTableRef?.$el.querySelector('.el-table__body-wrapper tbody')
    if (!el) return

    if (draggableInstance.value) draggableInstance.value.destroy()

    draggableInstance.value = useDraggable(el, draggableData, {
      animation: 150,
      handle: '.drag-handle',
      ghostClass: 'dragging-ghost',
      onStart: () => document.body.classList.add('is-dragging'),
      onEnd: (evt) => {
        document.body.classList.remove('is-dragging')
        handleDragEnd(evt)
      },
      onMove: (evt) => {
        const draggedRow = evt.dragged
        const targetRow = evt.related
        const getParentClass = (el: HTMLElement) => Array.from(el.classList).find(c => c.startsWith('parent-'))
        const getLevelClass = (el: HTMLElement) => Array.from(el.classList).find(c => c.startsWith('level-'))
        return getParentClass(draggedRow) === getParentClass(targetRow) && getLevelClass(draggedRow) === getLevelClass(targetRow)
      }
    })
  }

  // --- 菜单列表处理 ---
  const getMenuList = async () => {
    console.log('getMenuList called')
    loading.value = true
    dataFromBackend.value = false
    try {
      const list = await fetchGetMenuTreeAll()
      console.log('Menu data from backend:', list)
      tableData.value = Array.isArray(list) ? list : []
      dataFromBackend.value = tableData.value.length > 0
    } catch (error) {
      console.log('Error fetching menu data:', error)
      const list = JSON.parse(JSON.stringify(asyncRoutes))
      ensureId(list)
      console.log('Menu data from asyncRoutes:', list)
      tableData.value = list
    } finally {
      loading.value = false
      console.log('Menu data after getMenuList:', tableData.value)
    }
  }

  const ensureId = (items: any[]) => {
    items.forEach(item => {
      if (item.id == null) item.id = item.path
      if (item.children?.length) ensureId(item.children)
    })
  }

  const findMenuNodeById = (items: any[], targetId: string): any | null => {
    for (const item of items) {
      if (String(item.id || item.path) === targetId) return item
      if (item.children?.length) {
        const child = findMenuNodeById(item.children, targetId)
        if (child) return child
      }
    }
    return null
  }

  const getRawSiblingIds = (parentId: string | null) => {
    const siblings = parentId
      ? (findMenuNodeById(tableData.value, parentId)?.children ?? [])
      : tableData.value

    return siblings.map((item: any) => String(item.id || item.path))
  }

  const moveArrayItem = <T,>(list: T[], fromIndex: number, toIndex: number) => {
    const next = [...list]
    if (fromIndex < 0 || toIndex < 0 || fromIndex >= next.length || toIndex >= next.length) return next
    const [moved] = next.splice(fromIndex, 1)
    next.splice(toIndex, 0, moved)
    return next
  }

  const buildVisibleSiblingIdsAfterDrag = (evt: any, parentId: string | null, levelClass: string, draggedMenuClass: string) => {
    const level = Number(levelClass.replace('level-', ''))
    if (Number.isNaN(level)) return []

    const currentFlat = [...draggableData.value]
    const oldIndex = Number(evt.oldIndex)
    const newIndex = Number(evt.newIndex)
    const draggedIndex = currentFlat.findIndex((item: any) => `menu-id-${item._rowClassId}` === draggedMenuClass)

    let resolvedFlat = currentFlat
    if (draggedIndex === oldIndex) {
      resolvedFlat = moveArrayItem(currentFlat, oldIndex, newIndex)
    } else if (draggedIndex === newIndex) {
      resolvedFlat = currentFlat
    } else {
      resolvedFlat = moveArrayItem(currentFlat, oldIndex, newIndex)
    }

    return resolvedFlat
      .filter((item: any) => item._parentId === parentId && item._level === level)
      .map((item: any) => String(item.id || item.path))
  }

  const mergeVisibleOrderWithRawSiblings = (rawSiblingIds: string[], visibleSiblingIds: string[]) => {
    if (rawSiblingIds.length === 0) return visibleSiblingIds
    const visibleIdSet = new Set(visibleSiblingIds)
    const visibleQueue = [...visibleSiblingIds]

    return rawSiblingIds.map((id) => {
      if (!visibleIdSet.has(id)) return id
      return visibleQueue.shift() ?? id
    })
  }

  const applySiblingOrderToTree = (parentId: string | null, nextSiblingIds: string[]) => {
    const sortMap = new Map(nextSiblingIds.map((id, index) => [id, index]))
    const sortSiblings = (siblings: any[]) => {
      siblings.sort((a: any, b: any) => {
        const aId = String(a.id || a.path)
        const bId = String(b.id || b.path)
        return (sortMap.get(aId) ?? Number.MAX_SAFE_INTEGER) - (sortMap.get(bId) ?? Number.MAX_SAFE_INTEGER)
      })
      siblings.forEach((item: any, index: number) => {
        item.sort_order = index + 1
      })
    }

    if (parentId == null) {
      sortSiblings(tableData.value)
      return
    }

    const parentNode = findMenuNodeById(tableData.value, parentId)
    if (parentNode?.children?.length) {
      sortSiblings(parentNode.children)
    }
  }

  const isRowExpanded = (id: string) => {
    console.log('Checking if row is expanded:', id, expandedRowKeys.value)
    return expandedRowKeys.value.includes(id)
  }

  const convertAuthListToChildren = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items.map(item => {
      const cloned = JSON.parse(JSON.stringify(item))
      if (cloned.children?.length) {
        cloned.children = convertAuthListToChildren(cloned.children)
      }
      if (item.meta?.authList?.length) {
        const authChildren = item.meta.authList.map((auth: any) => ({
          id: `${item.id}_auth_${auth.authMark}`,
          path: `${item.path}_auth_${auth.authMark}`,
          name: `${String(item.name)}_auth_${auth.authMark}`,
          meta: { title: auth.title, authMark: auth.authMark, isAuthButton: true, parentPath: item.path }
        }))
        cloned.children = cloned.children ? [...cloned.children, ...authChildren] : authChildren
      }
      return cloned
    })
  }

  const filterAndSearch = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items
      .filter(item => showInnerPages.value || !item.meta?.isInnerPage)
      .map(item => {
        const cloned = JSON.parse(JSON.stringify(item))
        if (cloned.children?.length) {
          cloned.children = filterAndSearch(cloned.children)
        }
        return cloned
      })
      .filter(item => {
        const searchName = appliedFilters.name?.toLowerCase().trim() || ''
        const searchRoute = appliedFilters.route?.toLowerCase().trim() || ''
        const titleMatch = !searchName || formatMenuTitle(item.meta?.title).toLowerCase().includes(searchName)
        const routeMatch = !searchRoute || (item.path || '').toLowerCase().includes(searchRoute)
        return titleMatch && routeMatch || (item.children && item.children.length > 0)
      })
  }

  /** 递归展平树结构，仅包含展开的节点 */
  const flattenTree = (nodes: AppRouteRecord[], parentId: string | null = null, level = 0): any[] => {
    const list: any[] = []
    nodes.forEach((node: any) => {
      const id = String(node.id || node.path)
      list.push({ ...node, _parentId: parentId, _level: level, _rowClassId: buildRowClassId(id) })
      
      if (node.children && node.children.length > 0 && isRowExpanded(id)) {
        list.push(...flattenTree(node.children, id, level + 1))
      }
    })
    return list
  }

  /** 获取处理后的完整菜单树（过滤+转换） */
  const getProcessedTree = (nodes: AppRouteRecord[]): AppRouteRecord[] => {
    const filtered = filterAndSearch(nodes)
    return convertAuthListToChildren(filtered)
  }

  const updateDraggableData = () => {
    const processedTree = getProcessedTree(tableData.value)
    draggableData.value = flattenTree(processedTree)
  }

  watch([tableData, expandedRowKeys, showInnerPages, appliedFilters], () => {
    updateDraggableData()
  }, { deep: true, immediate: true })

  // --- 表格列配置 ---
  const { columnChecks, columns: displayColumns } = useTableColumns(() => [
    { prop: 'title', label: '菜单名称', minWidth: 200, useSlot: true, slotName: 'title' },
    { prop: 'sort_order', label: '排序', width: 80, align: 'center' },
    { prop: 'type', label: '类型', width: 100, align: 'center', useSlot: true, slotName: 'type' },
    { prop: 'path', label: '路由', minWidth: 150, useSlot: true, slotName: 'path' },
    { prop: 'status', label: '状态', width: 100, align: 'center', useSlot: true, slotName: 'status' },
    { prop: 'operation', label: '操作', width: 120, align: 'center', useSlot: true, slotName: 'operation' }
  ])

  // --- 辅助方法 ---
  const getMenuTypeTag = (row: any) => {
    if (row.meta?.isAuthButton) return 'danger'
    if (row.meta?.isInnerPage) return 'warning'
    if (row.children?.length) return 'info'
    return 'primary'
  }

  const getMenuTypeText = (row: any) => {
    if (row.meta?.isAuthButton) return '按钮'
    if (row.meta?.isInnerPage) return '内页'
    if (row.children?.length) return '目录'
    return '菜单'
  }

  const getOperationList = (row: any): ButtonMoreItem[] => {
    const list: ButtonMoreItem[] = [
      { key: 'add', label: '新增子菜单', icon: 'ri:add-fill' },
      { key: 'edit', label: '编辑菜单', icon: 'ri:edit-2-line' }
    ]
    if (!row.is_system) {
      list.push({ key: 'delete', label: '删除菜单', icon: 'ri:delete-bin-4-line', color: '#f56c6c' })
    }
    return list
  }

  // --- 事件处理 ---
  const handleReset = () => { Object.assign(formFilters, initialSearchState); Object.assign(appliedFilters, initialSearchState); getMenuList() }
  const handleSearch = () => { Object.assign(appliedFilters, formFilters); getMenuList() }
  const handleRefresh = () => getMenuList()
  const rowKey = (row: any) => String(row.id || row.path)

  const handleExpandChange = (row: any, expanded: boolean) => {
    console.log('handleExpandChange called:', row, expanded)
    const id = String(row.id || row.path)
    const currentKeys = [...expandedRowKeys.value]
    if (expanded) {
      if (!currentKeys.includes(id)) {
        currentKeys.push(id)
      }
    } else {
      const index = currentKeys.indexOf(id)
      if (index > -1) {
        currentKeys.splice(index, 1)
      }
    }
    expandedRowKeys.value = currentKeys
    console.log('expandedRowKeys updated:', expandedRowKeys.value)
    updateDraggableData()
  }

  const toggleExpand = () => {
    isExpanded.value = !isExpanded.value
    if (isExpanded.value) {
      const allKeys: string[] = []
      const collect = (nodes: any[]) => nodes.forEach(n => { if (n.children?.length) { allKeys.push(String(n.id || n.path)); collect(n.children) } })
      collect(getProcessedTree(tableData.value))
      expandedRowKeys.value = allKeys
    } else {
      expandedRowKeys.value = []
    }
    updateDraggableData()
  }

  const handleDragEnd = async (evt: any) => {
    if (!dataFromBackend.value) return ElMessage.info('当前为预览模式，拖拽仅限本地效果')
    const { newIndex, oldIndex } = evt
    if (newIndex === oldIndex) return
    try {
      const draggedRow = evt.item as HTMLElement | undefined
      if (!draggedRow) return ElMessage.error('拖动项不存在')

      const rowClasses = Array.from(draggedRow.classList)
      const parentClass = rowClasses.find(className => className.startsWith('parent-'))
      const levelClass = rowClasses.find(className => className.startsWith('level-'))
      const draggedMenuClass = rowClasses.find(className => className.startsWith('menu-id-'))
      if (!parentClass || !levelClass || !draggedMenuClass) {
        return ElMessage.error('拖拽分组信息丢失')
      }

      const draggedItem = draggableData.value.find((item: any) => `menu-id-${item._rowClassId}` === draggedMenuClass)
      const parentId = draggedItem?._parentId ?? null
      const siblingIds = buildVisibleSiblingIdsAfterDrag(evt, parentId, levelClass, draggedMenuClass)

      if (siblingIds.length === 0) return ElMessage.error('未获取到拖拽后的排序')

      const rawSiblingIds = getRawSiblingIds(parentId)
      const finalVisibleSiblingIds = siblingIds.filter((id) => rawSiblingIds.includes(id))
      const mergedSiblingIds = mergeVisibleOrderWithRawSiblings(rawSiblingIds, finalVisibleSiblingIds)

      applySiblingOrderToTree(parentId, mergedSiblingIds)
      updateDraggableData()

      await fetchUpdateMenuSortByParent(parentId, mergedSiblingIds)
      ElMessage.success('排序更新成功')
      getMenuList()
    } catch (e: any) { ElMessage.error(e?.message || '排序失败'); getMenuList() }
  }

  // --- CRUD 操作 ---
  const handleAddMenu = () => { dialogType.value = 'menu'; editData.value = null; parentRowForAdd.value = null; lockMenuType.value = true; dialogVisible.value = true }
  const handleAddUnderRow = (row: any) => { dialogType.value = 'menu'; editData.value = null; parentRowForAdd.value = row; lockMenuType.value = false; dialogVisible.value = true }
  const handleEditMenu = (row: any) => { dialogType.value = 'menu'; editData.value = row; parentRowForAdd.value = null; lockMenuType.value = true; dialogVisible.value = true }
  const handleMenuOperation = (item: ButtonMoreItem, row: any) => {
    if (item.key === 'add') handleAddUnderRow(row)
    else if (item.key === 'edit') handleEditMenu(row)
    else if (item.key === 'delete') handleDeleteMenu(row)
  }

  const handleDeleteMenu = async (row: any) => {
    if (!dataFromBackend.value || !row.id) return ElMessage.info('预览模式无法删除')
    if (row.is_system) return ElMessage.warning('系统菜单不可删除')
    try {
      await ElMessageBox.confirm('确定要删除该菜单吗？', '提示', { type: 'warning' })
      await fetchDeleteMenu(String(row.id))
      ElMessage.success('删除成功')
      getMenuList()
    } catch {}
  }

  const handleSubmit = async (formData: any) => {
    if (!dataFromBackend.value) return getMenuList()
    try {
      const isInner = formData.menuType === 'inner'
      const payload = {
        path: formData.path || '/',
        name: formData.label || '',
        component: formData.component || '',
        title: formData.name || '',
        icon: formData.icon || '',
        sort_order: formData.sort ?? 0,
        meta: {
          roles: formData.roles, isEnable: formData.isEnable, keepAlive: formData.keepAlive,
          isHide: isInner ? true : !!formData.isHide, isHideTab: formData.isHideTab,
          isIframe: formData.isIframe, showBadge: formData.showBadge, fixedTab: formData.fixedTab,
          isFullPage: formData.isFullPage, isInnerPage: isInner
        }
      }
      if (editData.value?.id) {
        const parentId = formData.parentId?.trim() || null
        await fetchUpdateMenu(String(editData.value.id), { ...payload, parent_id: parentId })
      } else {
        const parentId = formData.parentId?.trim() || (parentRowForAdd.value?.id ? String(parentRowForAdd.value.id) : null)
        await fetchCreateMenu({ ...payload, parent_id: parentId })
      }
      ElMessage.success('保存成功')
      getMenuList()
    } catch (e: any) { ElMessage.error(e?.message || '保存失败') }
  }

  // --- 生命周期 & 监听 ---
  onMounted(() => getMenuList())
  watch(() => draggableData.value, () => nextTick(() => initDraggable()), { deep: true })
</script>

<style lang="scss" scoped>
  .dragging-ghost { opacity: 0.5; background-color: var(--el-color-primary-light-9); }
  :deep(.menu-row) { transition: background-color 0.2s; }
  :deep(.is-dragging .menu-row:not(.dragging-ghost)) { cursor: grabbing; }

  /* 彻底隐藏 Element Plus 表格默认的展开图标列和缩进占位 */
  :deep(.el-table__expand-column),
  :deep(.el-table__indent),
  :deep(.el-table__placeholder),
  :deep(.el-table__expand-icon) {
    display: none !important;
    width: 0 !important;
    padding: 0 !important;
    margin: 0 !important;
  }
</style>

