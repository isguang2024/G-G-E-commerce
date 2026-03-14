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
          <ElTooltip
            content="内页默认不显示在侧栏，仅通过按钮跳转；开启后可在列表中查看内页项"
            placement="top"
          >
            <span class="inline-flex items-center gap-2">
              <span class="text-sm text-gray-600">显示内页</span>
              <ElSwitch v-model="showInnerPages" />
            </span>
          </ElTooltip>
          <ElTooltip content="创建菜单" placement="top">
            <ElButton type="primary" @click="handleAddMenu" v-ripple class="ml-2">
              创建菜单
            </ElButton>
          </ElTooltip>
          <ElButton @click="toggleExpand" v-ripple class="ml-2">
            {{ isExpanded ? '收起' : '展开' }}
          </ElButton>
          <ElTooltip content="备份菜单" placement="top">
            <ElButton @click="handleBackupMenu" v-ripple class="ml-2"> 备份 </ElButton>
          </ElTooltip>
          <ElTooltip content="管理备份" placement="top">
            <ElButton @click="handleManageBackups" v-ripple class="ml-2"> 管理备份 </ElButton>
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
          <span>{{ row.meta?.isAuthButton ? '' : row.meta?.link || row.path || '' }}</span>
        </template>

        <!-- 组件路径列 -->
        <template #component="{ row }">
          <span class="text-gray-600">{{ row.component || '-' }}</span>
        </template>

        <!-- 高级配置列 -->
        <template #advanced="{ row }">
          <div class="advanced-configs">
            <ElTag v-if="!row.meta?.isAuthButton && row.meta.keepAlive" size="small" effect="light" type="primary" class="mr-2">
              缓存
            </ElTag>
            <ElTag v-if="!row.meta?.isAuthButton && !row.meta?.isInnerPage && row.meta.isHide" size="small" effect="light" type="warning" class="mr-2">
              隐藏
            </ElTag>
            <ElTag v-if="!row.meta?.isAuthButton && row.meta.isIframe" size="small" effect="light" type="info" class="mr-2">
              内嵌
            </ElTag>
            <ElTag v-if="!row.meta?.isAuthButton && row.meta.showBadge" size="small" effect="light" type="success" class="mr-2">
              徽章
            </ElTag>
            <ElTag v-if="!row.meta?.isAuthButton && row.meta.fixedTab" size="small" effect="light" type="danger" class="mr-2">
              固定
            </ElTag>
            <ElTag v-if="!row.meta?.isAuthButton && row.meta.isFullPage" size="small" effect="light" type="primary" class="mr-2">
              全屏
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
            <ElTableColumn label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <div class="flex gap-2">
                  <ElButton
                    type="primary"
                    size="small"
                    @click="handleRestoreBackup(row.id)"
                  >
                    恢复
                  </ElButton>
                  <ElButton type="danger" size="small" @click="handleDeleteBackup(row.id)">
                    删除
                  </ElButton>
                </div>
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
    fetchCreateMenuBackup,
    fetchGetMenuBackupList,
    fetchDeleteMenuBackup,
    fetchRestoreMenuBackup
  } from '@/api/system-manage'
  import { ElTag, ElMessageBox, ElMessage, ElTooltip, ElButton, ElSwitch } from 'element-plus'
  import { useAuth } from '@/hooks/core/useAuth'

  defineOptions({ name: 'Menus' })

  // --- 权限管理 ---
  const { hasAuth } = useAuth()

  // --- 状态管理 ---
  const loading = ref(false)
  const isExpanded = ref(false)
  const showInnerPages = ref(false)
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
  const formItems = computed(() => [
    { label: '菜单名称', key: 'name', type: 'input', props: { clearable: true } },
    { label: '路由地址', key: 'route', type: 'input', props: { clearable: true } }
  ])

  // --- 弹窗相关 ---
  const dialogVisible = ref(false)
  const dialogType = ref<'menu' | 'inner'>('menu')
  const editData = ref<any>(null)
  const parentRowForAdd = ref<AppRouteRecord | null>(null)
  const lockMenuType = ref(false)

  // --- 菜单列表处理 ---
  const getMenuList = async () => {
    console.log('getMenuList called')
    loading.value = true
    dataFromBackend.value = false
    try {
      const list = await fetchGetMenuTreeAll()
      console.log('Menu data from backend:', list)
      const rawData = Array.isArray(list) ? list : []
      tableData.value = filterAndSearch(rawData)
      dataFromBackend.value = tableData.value.length > 0
    } catch (error) {
      console.log('Error fetching menu data:', error)
      const list = JSON.parse(JSON.stringify(asyncRoutes))
      ensureId(list)
      console.log('Menu data from asyncRoutes:', list)
      tableData.value = filterAndSearch(list)
    } finally {
      loading.value = false
      console.log('Menu data after getMenuList:', tableData.value)
    }
  }

  const ensureId = (items: any[]) => {
    items.forEach((item) => {
      if (item.id == null) item.id = item.path
      if (item.children?.length) ensureId(item.children)
    })
  }

  const filterAndSearch = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items
      .filter((item) => showInnerPages.value || !item.meta?.isInnerPage)
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
      list.push({
        key: 'delete',
        label: '删除菜单',
        icon: 'ri:delete-bin-4-line',
        color: '#f56c6c'
      })
    }
    return list
  }

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

  watch(showInnerPages, () => {
    getMenuList()
  })
</script>

<style lang="scss" scoped>
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
