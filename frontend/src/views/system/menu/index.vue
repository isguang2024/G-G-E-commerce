<!-- 菜单管理页面 -->
<template>
  <div class="menu-page art-full-height">
    <!-- 搜索栏 -->
    <MenuSearch
      v-show="showSearchBar"
      v-model="formFilters"
      @reset="handleReset"
      @search="handleSearch"
    />

    <ElCard
      class="art-table-card"
      shadow="never"
      :style="{ marginTop: showSearchBar ? '12px' : '0' }"
    >
      <!-- 表格头部 -->
      <ArtTableHeader
        :showZebra="false"
        :loading="loading"
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        @refresh="handleRefresh"
        >
          <template #left>
            <div class="menu-filter-switches">
              <ElTooltip
                content="内页默认不显示在侧栏，仅通过按钮跳转；开启后可在列表中查看内页项"
                placement="top"
              >
                <span class="inline-flex items-center gap-2">
                  <span class="text-sm text-gray-600">显示内页</span>
                  <ElSwitch v-model="showInnerPages" />
                </span>
              </ElTooltip>
              <span class="inline-flex items-center gap-2">
                <span class="text-sm text-gray-600">显示隐藏</span>
                <ElSwitch v-model="showHiddenMenus" />
              </span>
              <span class="inline-flex items-center gap-2">
                <span class="text-sm text-gray-600">显示内嵌</span>
                <ElSwitch v-model="showIframeMenus" />
              </span>
              <span class="inline-flex items-center gap-2">
                <span class="text-sm text-gray-600">显示启用</span>
                <ElSwitch v-model="showEnabledMenus" />
              </span>
            </div>
            <ElTooltip content="创建菜单" placement="top">
              <ElButton v-action="'system.menu.manage'" type="primary" @click="handleAddMenu" v-ripple class="ml-2">
                创建菜单
            </ElButton>
          </ElTooltip>
          <ElButton @click="toggleExpand" v-ripple class="ml-2">
            {{ isExpanded ? '收起' : '展开' }}
          </ElButton>
          <ElTooltip content="备份菜单" placement="top">
            <ElButton v-action="'system.menu.backup'" @click="handleBackupMenu" v-ripple class="ml-2"> 备份 </ElButton>
          </ElTooltip>
          <ElTooltip content="管理备份" placement="top">
            <ElButton v-action="'system.menu.backup'" @click="handleManageBackups" v-ripple class="ml-2"> 管理备份 </ElButton>
          </ElTooltip>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        :rowKey="rowKey"
        :loading="loading"
        :columns="displayColumns"
        :data="tableData"
        :stripe="false"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="false"
      >
        <!-- 菜单名称列 -->
        <template #title="{ row }">
          <ArtSvgIcon v-if="row.meta?.icon" :icon="row.meta.icon" class="mr-2 text-g-500" />
          <span>{{ formatMenuTitle(row.meta?.title) }}</span>
        </template>

        <!-- 菜单类型列 -->
        <template #type="{ row }">
          <ElTag :type="getMenuTypeTag(row)">{{ getMenuTypeText(row) }}</ElTag>
        </template>

        <!-- 路由列 -->
        <template #path="{ row }">
          <span>{{ row.meta?.link || row.path || '' }}</span>
        </template>

        <!-- 组件路径列 -->
        <template #component="{ row }">
          <span class="text-gray-600">{{ row.component || '-' }}</span>
        </template>

        <!-- 高级配置列 -->
        <template #advanced="{ row }">
          <div class="advanced-configs">
            <ElTag v-if="row.meta?.keepAlive" size="small" effect="light" type="primary" class="mr-2">
              缓存
            </ElTag>
            <ElTag v-if="!row.meta?.isInnerPage && row.meta?.isHide" size="small" effect="light" type="warning" class="mr-2">
              隐藏
            </ElTag>
            <ElTag v-if="row.meta?.isIframe" size="small" effect="light" type="info" class="mr-2">
              内嵌
            </ElTag>
            <ElTag v-if="row.meta?.showBadge" size="small" effect="light" type="success" class="mr-2">
              徽章
            </ElTag>
            <ElTag v-if="row.meta?.fixedTab" size="small" effect="light" type="danger" class="mr-2">
              固定
            </ElTag>
            <ElTag v-if="row.meta?.isFullPage" size="small" effect="light" type="primary" class="mr-2">
              全屏
            </ElTag>
            <ElTag
              v-if="getMenuActionRequirement(row.meta).actions.length"
              size="small"
              effect="light"
              type="info"
              class="mr-2"
            >
              {{ getMenuActionRequirementLabel(row) }}
            </ElTag>
          </div>
        </template>

        <!-- 状态列 -->
        <template #status="{ row }">
          <ElTag :type="row.meta?.isEnable !== false ? 'success' : 'info'">
            {{ row.meta?.isEnable !== false ? '启用' : '未启用' }}
          </ElTag>
        </template>

        <!-- 操作列 -->
        <template #operation="{ row }">
          <div class="flex items-center justify-center gap-2">
            <ArtButtonMore
              :list="getOperationList(row)"
              @click="(item) => handleMenuOperation(item, row)"
            />
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

      <MenuPermissionDialog
        v-model="actionRequirementVisible"
        :menuData="actionRequirementData"
        @submit="handleActionRequirementSubmit"
      />

      <!-- 备份菜单弹窗 -->
      <ElDialog v-model="backupDialogVisible" title="备份菜单" width="500px">
        <ElForm :model="{ name: backupName, description: backupDescription }" label-width="80px">
          <ElFormItem label="备份名称" required>
            <ElInput v-model="backupName" placeholder="请输入备份名称" />
          </ElFormItem>
          <ElFormItem label="备份描述">
            <ElInput
              v-model="backupDescription"
              type="textarea"
              placeholder="请输入备份描述"
              :rows="3"
            />
          </ElFormItem>
        </ElForm>
        <template #footer>
          <span class="dialog-footer">
            <ElButton @click="backupDialogVisible = false">取消</ElButton>
            <ElButton type="primary" @click="handleCreateBackup" :loading="backupLoading">
              确认备份
            </ElButton>
          </span>
        </template>
      </ElDialog>

      <!-- 管理备份弹窗 -->
      <ElDialog v-model="backupListDialogVisible" title="管理备份" width="800px" class="backup-dialog">
        <div class="backup-list-container">
          <ElTable v-loading="backupLoading" :data="backupList" style="width: 100%" border stripe>
            <ElTableColumn prop="name" label="备份名称" width="200">
              <template #default="{ row }">
                <span class="font-medium">{{ row.name }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="description" label="备份描述">
              <template #default="{ row }">
                <span class="text-gray-600">{{ row.description || '-' }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="created_at" label="创建时间" width="200" />
            <ElTableColumn prop="created_by" label="创建人" width="150">
              <template #default="{ row }">
                <span class="text-gray-600">{{ row.created_by || '系统' }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="72" fixed="right" align="center">
              <template #default="{ row }">
                <ArtButtonMore
                  :list="getBackupOperationList()"
                  @click="(item) => handleBackupOperation(item, row)"
                />
              </template>
            </ElTableColumn>
          </ElTable>
          <div v-if="backupList.length === 0" class="empty-backup">
            <ElEmpty description="暂无备份数据" />
          </div>
        </div>
      </ElDialog>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref, reactive, watch, nextTick } from 'vue'
  import { formatMenuTitle } from '@/utils/router'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import type { AppRouteRecord } from '@/types/router'
  import MenuDialog from './modules/menu-dialog.vue'
  import MenuPermissionDialog from './modules/menu-permission-dialog.vue'
  import MenuSearch from './modules/menu-search.vue'
  import {
    fetchGetMenuTreeAll,
    fetchCreateMenu,
    fetchUpdateMenu,
    fetchDeleteMenu,
    fetchCreateMenuBackup,
    fetchGetMenuBackupList,
    fetchDeleteMenuBackup,
    fetchRestoreMenuBackup
  } from '@/api/system-manage'
  import { ElTag, ElMessageBox, ElMessage, ElTooltip, ElButton, ElSwitch } from 'element-plus'
  import { getMenuActionRequirement } from '@/utils/permission/menu'

  defineOptions({ name: 'Menus' })

  // --- 状态管理 ---
  const loading = ref(false)
  const showSearchBar = ref(true)
  const isExpanded = ref(false)
  const showInnerPages = ref(false)
  const showHiddenMenus = ref(true)
  const showIframeMenus = ref(true)
  const showEnabledMenus = ref(true)
  const tableRef = ref()
  const tableData = ref<AppRouteRecord[]>([])
  const dataFromBackend = ref(false)

  // --- 菜单备份相关状态 ---
  const backupLoading = ref(false)
  const backupDialogVisible = ref(false)
  const backupListDialogVisible = ref(false)
  const backupName = ref('')
  const backupDescription = ref('')
  const backupList = ref<any[]>([])

  // --- 搜索相关 ---
  const initialSearchState = { name: '', route: '' }
  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })
  // --- 弹窗相关 ---
  const dialogVisible = ref(false)
  const dialogType = ref<'menu' | 'inner'>('menu')
  const editData = ref<any>(null)
  const parentRowForAdd = ref<AppRouteRecord | null>(null)
  const lockMenuType = ref(false)
  const actionRequirementVisible = ref(false)
  const actionRequirementData = ref<any>(null)

  // --- 菜单列表处理 ---
  const getMenuList = async () => {
    loading.value = true
    dataFromBackend.value = false
    try {
      const list = await fetchGetMenuTreeAll()
      const rawData = Array.isArray(list) ? list : []
      tableData.value = filterAndSearch(rawData)
      dataFromBackend.value = true
    } catch (error) {
      console.error('获取菜单数据失败:', error)
      tableData.value = []
      ElMessage.error('菜单数据加载失败，请检查后端菜单配置或服务状态')
    } finally {
      loading.value = false
    }
  }

  const filterAndSearch = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items
      .filter((item) => {
        if (!showInnerPages.value && item.meta?.isInnerPage) return false
        if (!showHiddenMenus.value && !item.meta?.isInnerPage && item.meta?.isHide) return false
        if (!showIframeMenus.value && item.meta?.isIframe) return false
        if (!showEnabledMenus.value && item.meta?.isEnable !== false) return false
        return true
      })
      .map((item) => {
        const cloned = JSON.parse(JSON.stringify(item))

        if (cloned.children?.length) {
          cloned.children = filterAndSearch(cloned.children)
        }

        return cloned
      })
      .filter((item) => {
        const searchName = appliedFilters.name?.toLowerCase().trim() || ''
        const searchRoute = appliedFilters.route?.toLowerCase().trim() || ''
        const titleMatch =
          !searchName || formatMenuTitle(item.meta?.title).toLowerCase().includes(searchName)
        const routeMatch = !searchRoute || (item.path || '').toLowerCase().includes(searchRoute)
        return (titleMatch && routeMatch) || (item.children && item.children.length > 0)
      })
  }

  // --- 表格列配置 ---  
  const { columnChecks, columns: displayColumns } = useTableColumns(() => [
    { prop: 'title', label: '菜单名称', minWidth: 200, useSlot: true, slotName: 'title' },
    { prop: 'sort_order', label: '排序', width: 80, align: 'center' },
    { prop: 'type', label: '类型', width: 100, align: 'center', useSlot: true, slotName: 'type' },
    { prop: 'path', label: '路由', minWidth: 150, useSlot: true, slotName: 'path' },
    { prop: 'component', label: '组件路径', minWidth: 200, useSlot: true, slotName: 'component' },
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
  const getMenuTypeTag = (row: any) => {
    if (row.meta?.isInnerPage) return 'warning'
    if (row.children?.length) return 'info'
    return 'primary'
  }

  const getMenuTypeText = (row: any) => {
    if (row.meta?.isInnerPage) return '内页'
    if (row.children?.length) return '目录'
    return '菜单'
  }

  const getMenuActionRequirementLabel = (row: any) => {
    const requirement = getMenuActionRequirement(row.meta)
    if (!requirement.actions.length) return ''
    const visibilityText = requirement.visibilityMode === 'show' ? '显示' : '隐藏'
    return `功能权限: 不满足${visibilityText}`
  }

  const getOperationList = (row: any): ButtonMoreItem[] => {
    const list: ButtonMoreItem[] = [
      { key: 'add', label: '新增子菜单', icon: 'ri:add-fill', auth: 'system.menu.manage' },
      { key: 'edit', label: '编辑菜单', icon: 'ri:edit-2-line', auth: 'system.menu.manage' },
      { key: 'action_requirement', label: '功能权限', icon: 'ri:shield-keyhole-line', auth: 'system.menu.manage' }
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

  const getBackupOperationList = (): ButtonMoreItem[] => [
    { key: 'restore', label: '恢复备份', icon: 'ri:history-line', auth: 'system.menu.backup' },
    {
      key: 'delete',
      label: '删除备份',
      icon: 'ri:delete-bin-4-line',
      color: '#f56c6c',
      auth: 'system.menu.backup'
    }
  ]

  // --- 事件处理 ---
  const handleReset = () => {
    Object.assign(formFilters, initialSearchState)
    Object.assign(appliedFilters, initialSearchState)
    getMenuList()
  }
  const handleSearch = () => {
    Object.assign(appliedFilters, formFilters)
    getMenuList()
  }
  const handleRefresh = () => getMenuList()
  const rowKey = (row: any) => String(row.id || row.path)

  const toggleExpand = () => {
    isExpanded.value = !isExpanded.value
    nextTick(() => {
      if (tableRef.value?.elTableRef && tableData.value) {
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

  // --- CRUD 操作 ---
  const handleAddMenu = () => {
    dialogType.value = 'menu'
    editData.value = null
    parentRowForAdd.value = null
    lockMenuType.value = true
    dialogVisible.value = true
  }
  const handleAddUnderRow = (row: any) => {
    dialogType.value = 'menu'
    editData.value = null
    parentRowForAdd.value = row
    lockMenuType.value = false
    dialogVisible.value = true
  }
  const handleEditMenu = (row: any) => {
    dialogType.value = 'menu'
    editData.value = row
    parentRowForAdd.value = null
    lockMenuType.value = true
    dialogVisible.value = true
  }
  const handleEditActionRequirement = (row: any) => {
    actionRequirementData.value = row
    actionRequirementVisible.value = true
  }
  const handleMenuOperation = (item: ButtonMoreItem, row: any) => {
    if (item.key === 'add') handleAddUnderRow(row)
    else if (item.key === 'edit') handleEditMenu(row)
    else if (item.key === 'action_requirement') handleEditActionRequirement(row)
    else if (item.key === 'delete') handleDeleteMenu(row)
  }

  const handleBackupOperation = (item: ButtonMoreItem, row: any) => {
    if (item.key === 'restore') {
      handleRestoreBackup(row.id)
      return
    }
    if (item.key === 'delete') {
      handleDeleteBackup(row.id)
    }
  }

  const handleDeleteMenu = async (row: any) => {
    if (!dataFromBackend.value || !row.id) return ElMessage.info('预览模式无法删除')
    if (row.is_system) return ElMessage.warning('系统菜单不可删除')
    try {
      await ElMessageBox.confirm('确定要删除该菜单吗？', '提示', { type: 'warning' })
      await fetchDeleteMenu(String(row.id))
      ElMessage.success('删除成功')
      getMenuList()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '删除失败')
      }
    }
  }

  const handleSubmit = async (formData: any) => {
    if (!dataFromBackend.value) return getMenuList()
    try {
      const isInner = formData.menuType === 'inner'
      // 构建meta对象
      const meta: any = {
        roles: formData.roles,
        isEnable: formData.isEnable,
        keepAlive: formData.keepAlive,
        isHide: isInner ? true : !!formData.isHide,
        isHideTab: formData.isHideTab,
        isIframe: formData.isIframe,
        showBadge: formData.showBadge,
        showTextBadge: formData.showTextBadge || '',
        link: formData.link || '',
        activePath: formData.activePath || '',
        fixedTab: formData.fixedTab,
        isFullPage: formData.isFullPage,
        isInnerPage: isInner
      }
      const requiredActions = Array.from(
        new Set(
          (formData.requiredActions || [])
            .map((item: string) => `${item || ''}`.trim())
            .filter(Boolean)
        )
      )
      if (requiredActions.length === 1) {
        meta.requiredAction = requiredActions[0]
      }
      if (requiredActions.length > 0) {
        meta.actionVisibilityMode = formData.actionVisibilityMode === 'show' ? 'show' : 'hide'
      }
      if (requiredActions.length > 1) {
        meta.requiredActions = requiredActions
        meta.actionMatchMode = formData.actionMatchMode === 'all' ? 'all' : 'any'
      }
      
      // 只有当customParent有值时才添加到meta中
      if (formData.customParent && formData.customParent.trim() !== '') {
        meta.customParent = formData.customParent
      }
      
      const payload = {
        path: formData.path || '/',
        name: formData.label || '',
        component: formData.component || '',
        title: formData.name || '',
        icon: formData.icon || '',
        sort_order: Number(formData.sort ?? 0),
        meta: meta
      }
      if (editData.value?.id) {
        const parentId = formData.parentId?.trim() || null
        await fetchUpdateMenu(String(editData.value.id), { ...payload, parent_id: parentId }, { showErrorMessage: false })
      } else {
        const parentId =
          formData.parentId?.trim() ||
          (parentRowForAdd.value?.id ? String(parentRowForAdd.value.id) : null)
        await fetchCreateMenu({ ...payload, parent_id: parentId }, { showErrorMessage: false })
      }
      // 只有成功时才显示成功消息
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
      const requiredActions = Array.from(
        new Set((formData.requiredActions || []).map((item: string) => `${item || ''}`.trim()).filter(Boolean))
      )
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
      await fetchUpdateMenu(
        String(row.id),
        {
          parent_id: row.parent_id ? String(row.parent_id) : null,
          path: row.path || '',
          name: row.name || '',
          component: typeof row.component === 'string' ? row.component : '',
          title: row.meta?.title || '',
          icon: row.meta?.icon || '',
          sort_order: Number(row.sort_order ?? 0),
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
  const handleBackupMenu = () => {
    backupName.value = ''
    backupDescription.value = ''
    backupDialogVisible.value = true
  }

  const handleCreateBackup = async () => {
    if (!backupName.value.trim()) {
      return ElMessage.warning('请输入备份名称')
    }
    backupLoading.value = true
    try {
      await fetchCreateMenuBackup({
        name: backupName.value.trim(),
        description: backupDescription.value.trim()
      })
      ElMessage.success('备份成功')
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
      const list = await fetchGetMenuBackupList()
      backupList.value = list || []
      backupListDialogVisible.value = true
    } catch (e: any) {
      ElMessage.error(e?.message || '获取备份列表失败')
    } finally {
      backupLoading.value = false
    }
  }

  const handleRestoreBackup = async (id: string) => {
    try {
      await ElMessageBox.confirm('确定要恢复该备份吗？恢复后会覆盖当前菜单配置。', '提示', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      backupLoading.value = true
      await fetchRestoreMenuBackup(id)
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
      await fetchDeleteMenuBackup(id)
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

  // --- 生命周期 & 监听 ---
  onMounted(() => getMenuList())

  watch([showInnerPages, showHiddenMenus, showIframeMenus, showEnabledMenus], () => {
    getMenuList()
  })
</script>

<style lang="scss" scoped>
  .menu-filter-switches {
    display: inline-flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
  }

  .backup-dialog {
    .backup-list-container {
      padding: 10px 0;
      
      .empty-backup {
        padding: 40px 0;
        text-align: center;
      }
    }
    
    :deep(.el-table) {
      .el-table__row {
        transition: all 0.3s ease;
        
        &:hover {
          background-color: #f5f7fa !important;
        }
      }
      
      .el-table__header-wrapper th {
        background-color: #fafafa;
        font-weight: 600;
      }
    }
  }
  
  .inline-flex {
    align-items: center;
  }
  
  .advanced-configs {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
  }
  
  :deep(.el-table) {
    .el-table__row {
      transition: all 0.3s ease;
      
      &:hover {
        background-color: #f5f7fa !important;
      }
    }
    
    .el-table__header-wrapper th {
      background-color: #fafafa;
      font-weight: 600;
    }
    
    .el-table__body-wrapper {
      .el-table__row {
        height: 48px;
      }
    }
  }
</style>
