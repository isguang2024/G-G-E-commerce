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
        :columns="columns"
        :data="filteredTableData"
        :stripe="false"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="false"
      />

      <!-- 菜单弹窗 -->
      <MenuDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :editData="editData"
        :menuTree="tableData"
        :editingMenuId="editData?.id"
        :initialParentId="parentRowForAdd?.id ?? ''"
        :lockType="lockMenuType"
        @submit="handleSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { formatMenuTitle } from '@/utils/router'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import type { AppRouteRecord } from '@/types/router'
  import MenuDialog from './modules/menu-dialog.vue'
  import { asyncRoutes } from '@/router/routes/asyncRoutes'
  import {
    fetchGetMenuTreeAll,
    fetchCreateMenu,
    fetchUpdateMenu,
    fetchDeleteMenu
  } from '@/api/system-manage'
  import { ElTag, ElMessageBox, ElMessage, ElTooltip, ElButton, ElSwitch } from 'element-plus'

  defineOptions({ name: 'Menus' })

  // 状态管理
  const loading = ref(false)
  const isExpanded = ref(false)
  /** 是否在列表中显示内页（默认不显示，减少杂乱） */
  const showInnerPages = ref(false)
  const tableRef = ref()

  // 弹窗相关
  const dialogVisible = ref(false)
  const dialogType = ref<'menu' | 'button'>('menu')
  const editData = ref<AppRouteRecord | any>(null)
  /** 添加子菜单时的父级行（仅新建时使用，用于传 parent_id） */
  const parentRowForAdd = ref<AppRouteRecord | null>(null)
  const lockMenuType = ref(false)

  // 搜索相关
  const initialSearchState = {
    name: '',
    route: ''
  }

  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })

  const formItems = computed(() => [
    {
      label: '菜单名称',
      key: 'name',
      type: 'input',
      props: { clearable: true }
    },
    {
      label: '路由地址',
      key: 'route',
      type: 'input',
      props: { clearable: true }
    }
  ])

  onMounted(() => {
    getMenuList()
  })

  /**
   * 获取菜单列表：优先从后端拉取完整树（all=1），失败时回退为前端路由
   */
  const getMenuList = async (): Promise<void> => {
    loading.value = true
    dataFromBackend.value = false
    try {
      const list = await fetchGetMenuTreeAll()
      tableData.value = Array.isArray(list) ? list : []
      dataFromBackend.value = tableData.value.length > 0
    } catch {
      const list = JSON.parse(JSON.stringify(asyncRoutes)) as (AppRouteRecord & { id?: string })[]
      ensureId(list)
      tableData.value = list
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取菜单类型标签颜色
   * @param row 菜单行数据
   * @returns 标签颜色类型
   */
  const getMenuTypeTag = (
    row: AppRouteRecord
  ): 'primary' | 'success' | 'warning' | 'info' | 'danger' => {
    if (row.meta?.isAuthButton) return 'danger'
    if (row.meta?.isInnerPage) return 'warning'
    if (row.children?.length) return 'info'
    if (row.meta?.link && row.meta?.isIframe) return 'success'
    if (row.path) return 'primary'
    if (row.meta?.link) return 'warning'
    return 'info'
  }

  /**
   * 获取菜单类型文本
   * @param row 菜单行数据
   * @returns 菜单类型文本
   */
  const getMenuTypeText = (row: AppRouteRecord): string => {
    if (row.meta?.isAuthButton) return '按钮'
    if (row.meta?.isInnerPage) return '内页'
    if (row.children?.length) return '目录'
    if (row.meta?.link && row.meta?.isIframe) return '内嵌'
    if (row.path) return '菜单'
    if (row.meta?.link) return '外链'
    return '未知'
  }

  // 表格列配置
  const { columnChecks, columns } = useTableColumns(() => [
    {
      prop: 'meta.title',
      label: '菜单名称',
      minWidth: 120,
      formatter: (row: AppRouteRecord) => formatMenuTitle(row.meta?.title)
    },
    {
      prop: 'type',
      label: '菜单类型',
      formatter: (row: AppRouteRecord) => {
        return h(ElTag, { type: getMenuTypeTag(row) }, () => getMenuTypeText(row))
      }
    },
    {
      prop: 'path',
      label: '路由',
      formatter: (row: AppRouteRecord) => {
        if (row.meta?.isAuthButton) return ''
        return row.meta?.link || row.path || ''
      }
    },
    {
      prop: 'meta.authList',
      label: '权限标识',
      formatter: (row: AppRouteRecord) => {
        if (row.meta?.isAuthButton) {
          return row.meta?.authMark || ''
        }
        if (!row.meta?.authList?.length) return ''
        return `${row.meta.authList.length} 个权限标识`
      }
    },
    {
      prop: 'date',
      label: '编辑时间',
      formatter: () => '2022-3-12 12:00:00'
    },
    {
      prop: 'status',
      label: '状态',
      formatter: () => h(ElTag, { type: 'success' }, () => '启用')
    },
    {
      prop: 'operation',
      label: '操作',
      width: 180,
      align: 'right',
      formatter: (row: AppRouteRecord) => {
        const buttonStyle = { style: 'text-align: right' }

        if (row.meta?.isAuthButton) {
          return h('div', buttonStyle, [
            h(ArtButtonTable, {
              type: 'edit',
              onClick: () => handleEditAuth(row)
            }),
            h(ArtButtonTable, {
              type: 'delete',
              onClick: () => handleDeleteAuth()
            })
          ])
        }

        const isSystem = (row as any).is_system === true
        return h('div', { class: 'flex items-center gap-1', style: buttonStyle.style }, [
          h(ElTooltip, { content: '可在此菜单下新增子菜单或权限按钮', placement: 'top' }, {
            default: () =>
              h(ArtButtonTable, { type: 'add', onClick: () => handleAddUnderRow(row) })
          }),
          h(ArtButtonTable, {
            type: 'edit',
            onClick: () => handleEditMenu(row)
          }),
          isSystem
            ? h(
                ElTooltip,
                { content: '系统默认菜单不可删除', placement: 'top' },
                { default: () => h(ElButton, { type: 'danger', link: true, disabled: true, size: 'small' }, () => '删除') }
              )
            : h(ArtButtonTable, {
                type: 'delete',
                onClick: () => handleDeleteMenu(row)
              })
        ])
      }
    }
  ])

  // 数据相关
  const tableData = ref<AppRouteRecord[]>([])
  /** 当前列表是否来自后端（用于判断是否可增删改） */
  const dataFromBackend = ref(false)

  /** 表格行 key：后端有 id 用 id，否则用 path */
  const rowKey = (row: AppRouteRecord & { id?: string }) => row.id ?? row.path

  /**
   * 为前端路由树递归补充 id（用于 rowKey）
   */
  const ensureId = (items: (AppRouteRecord & { id?: string })[]): void => {
    items.forEach((item) => {
      if (item.id == null) (item as Record<string, unknown>).id = item.path
      if (item.children?.length) ensureId(item.children as (AppRouteRecord & { id?: string })[])
    })
  }

  /**
   * 重置搜索条件
   */
  const handleReset = (): void => {
    Object.assign(formFilters, { ...initialSearchState })
    Object.assign(appliedFilters, { ...initialSearchState })
    getMenuList()
  }

  /**
   * 执行搜索
   */
  const handleSearch = (): void => {
    Object.assign(appliedFilters, { ...formFilters })
    getMenuList()
  }

  /**
   * 刷新菜单列表
   */
  const handleRefresh = (): void => {
    getMenuList()
  }

  /**
   * 深度克隆对象
   * @param obj 要克隆的对象
   * @returns 克隆后的对象
   */
  const deepClone = <T,>(obj: T): T => {
    if (obj === null || typeof obj !== 'object') return obj
    if (obj instanceof Date) return new Date(obj) as T
    if (Array.isArray(obj)) return obj.map((item) => deepClone(item)) as T

    const cloned = {} as T
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        cloned[key] = deepClone(obj[key])
      }
    }
    return cloned
  }

  /**
   * 将权限列表转换为子节点
   * @param items 菜单项数组
   * @returns 转换后的菜单项数组
   */
  const convertAuthListToChildren = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items.map((item) => {
      const clonedItem = deepClone(item)

      if (clonedItem.children?.length) {
        clonedItem.children = convertAuthListToChildren(clonedItem.children)
      }

      if (item.meta?.authList?.length) {
        const authChildren: AppRouteRecord[] = item.meta.authList.map(
          (auth: { title: string; authMark: string }) => ({
            path: `${item.path}_auth_${auth.authMark}`,
            name: `${String(item.name)}_auth_${auth.authMark}`,
            meta: {
              title: auth.title,
              authMark: auth.authMark,
              isAuthButton: true,
              parentPath: item.path
            }
          })
        )

        clonedItem.children = clonedItem.children?.length
          ? [...clonedItem.children, ...authChildren]
          : authChildren
      }

      return clonedItem
    })
  }

  /**
   * 过滤内页：当不显示内页时，递归移除 meta.isInnerPage 的节点
   */
  const filterInnerPages = (items: AppRouteRecord[], showInner: boolean): AppRouteRecord[] => {
    if (showInner) return items
    return items
      .filter((item) => !item.meta?.isInnerPage)
      .map((item) => {
        const cloned = deepClone(item)
        if (cloned.children?.length) {
          cloned.children = filterInnerPages(cloned.children, showInner)
        }
        return cloned
      })
  }

  /**
   * 搜索菜单
   * @param items 菜单项数组
   * @returns 搜索结果数组
   */
  const searchMenu = (items: AppRouteRecord[]): AppRouteRecord[] => {
    const results: AppRouteRecord[] = []

    for (const item of items) {
      const searchName = appliedFilters.name?.toLowerCase().trim() || ''
      const searchRoute = appliedFilters.route?.toLowerCase().trim() || ''
      const menuTitle = formatMenuTitle(item.meta?.title || '').toLowerCase()
      const menuPath = (item.path || '').toLowerCase()
      const nameMatch = !searchName || menuTitle.includes(searchName)
      const routeMatch = !searchRoute || menuPath.includes(searchRoute)

      if (item.children?.length) {
        const matchedChildren = searchMenu(item.children)
        if (matchedChildren.length > 0) {
          const clonedItem = deepClone(item)
          clonedItem.children = matchedChildren
          results.push(clonedItem)
          continue
        }
      }

      if (nameMatch && routeMatch) {
        results.push(deepClone(item))
      }
    }

    return results
  }

  // 过滤后的表格数据（先按内页开关过滤，再搜索，再展开权限按钮）
  const filteredTableData = computed(() => {
    const withoutInner = filterInnerPages(tableData.value, showInnerPages.value)
    const searchedData = searchMenu(withoutInner)
    return convertAuthListToChildren(searchedData)
  })

  /**
   * 添加菜单（顶级）
   */
  const handleAddMenu = (): void => {
    dialogType.value = 'menu'
    editData.value = null
    parentRowForAdd.value = null
    lockMenuType.value = true
    dialogVisible.value = true
  }

  /**
   * 在当前行下新增（弹窗内可选「菜单」添加子菜单，或「按钮」添加权限）
   */
  const handleAddUnderRow = (parentRow: AppRouteRecord & { id?: string }): void => {
    dialogType.value = 'menu'
    editData.value = null
    parentRowForAdd.value = parentRow as AppRouteRecord
    lockMenuType.value = false
    dialogVisible.value = true
  }

  /**
   * 编辑菜单
   * @param row 菜单行数据
   */
  const handleEditMenu = (row: AppRouteRecord): void => {
    dialogType.value = 'menu'
    editData.value = row
    parentRowForAdd.value = null
    lockMenuType.value = true
    dialogVisible.value = true
  }

  /**
   * 编辑权限按钮
   * @param row 权限行数据
   */
  const handleEditAuth = (row: AppRouteRecord): void => {
    dialogType.value = 'button'
    editData.value = {
      title: row.meta?.title,
      authMark: row.meta?.authMark
    }
    lockMenuType.value = false
    dialogVisible.value = true
  }

  /**
   * 菜单表单数据类型
   */
  interface MenuFormData {
    name: string
    path: string
    component?: string
    icon?: string
    parentId?: string
    roles?: string[]
    sort?: number
    [key: string]: any
  }

  /**
   * 将弹窗表单数据映射为后端菜单参数
   * 所有开关（页面缓存、隐藏菜单、是否内嵌等）存入 meta，由后端 menus.meta (JSONB) 持久化
   */
  const mapFormToMenuParams = (formData: MenuFormData & { menuType?: string }) => {
    const isInner = formData.menuType === 'inner'
    const isHide = isInner ? true : !!formData.isHide
    return {
      path: formData.path || '/',
      name: formData.label || '', // label 是权限标识（name）
      component: formData.component || '',
      title: formData.name || '', // name 是菜单名称（title）
      icon: formData.icon || '',
      sort_order: formData.sort ?? 0,
      meta: {
        roles: formData.roles,
        isEnable: formData.isEnable,
        keepAlive: formData.keepAlive,
        isHide,
        isHideTab: formData.isHideTab,
        isIframe: formData.isIframe,
        showBadge: formData.showBadge,
        showTextBadge: formData.showTextBadge || undefined,
        fixedTab: formData.fixedTab,
        activePath: formData.activePath || undefined,
        link: formData.link || undefined,
        isFullPage: formData.isFullPage,
        isInnerPage: isInner
      },
      hidden: isHide
    }
  }

  /**
   * 提交表单数据（仅在后端数据模式下调用接口）
   */
  const handleSubmit = async (formData: MenuFormData): Promise<void> => {
    if (!dataFromBackend.value) {
      ElMessage.info('当前为前端路由预览，保存将刷新列表')
      getMenuList()
      return
    }
    try {
      const payload = mapFormToMenuParams(formData)
      if (editData.value?.id) {
        // 总是发送 parent_id 字段：有值表示设置上级，null 表示设置为顶级
        const updateParams: any = { ...payload }
        if (formData.parentId && String(formData.parentId).trim() !== '') {
          // 用户选择了上级菜单
          updateParams.parent_id = String(formData.parentId).trim()
        } else {
          // 用户选择了顶级菜单，发送 null
          updateParams.parent_id = null
        }
        await fetchUpdateMenu(String(editData.value.id), updateParams)
        ElMessage.success('更新成功')
      } else {
        // 新增菜单时的父级ID逻辑：
        // 1. 如果用户在表单中选择了上级菜单（parentId有值），使用选择的值
        // 2. 如果用户选择了"顶级菜单"（parentId为空字符串），设为null表示顶级
        // 3. 如果既没有选择上级菜单，也没有parentRowForAdd，使用null
        // 4. 只有在通过"新增"按钮点击时，才使用parentRowForAdd的ID
        let parentId: string | null = null
        if (formData.parentId && formData.parentId.trim() !== '') {
          // 用户明确选择了上级菜单
          parentId = formData.parentId.trim()
        } else if (!formData.parentId && parentRowForAdd.value?.id) {
          // 通过"新增"按钮点击，自动设置为当前行的子菜单
          parentId = parentRowForAdd.value.id
        }
        // 总是发送 parent_id 字段：有值或 null
        await fetchCreateMenu({ ...payload, parent_id: parentId })
        parentRowForAdd.value = null
        ElMessage.success('新增成功')
      }
      await getMenuList()
    } catch (e) {
      ElMessage.error(e?.message || '保存失败')
    }
  }

  /**
   * 删除菜单（仅在后端数据模式下调用接口）
   */
  const handleDeleteMenu = async (row: AppRouteRecord & { id?: string; is_system?: boolean }): Promise<void> => {
    if (!dataFromBackend.value || !row.id) {
      ElMessage.info('当前为前端路由预览，无法删除')
      return
    }
    if (row.is_system) {
      ElMessage.warning('系统默认菜单不可删除')
      return
    }
    try {
      await ElMessageBox.confirm('确定要删除该菜单吗？删除后无法恢复', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await fetchDeleteMenu(String(row.id))
      ElMessage.success('删除成功')
      await getMenuList()
    } catch (e) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '删除失败')
      }
    }
  }

  /**
   * 删除权限按钮
   */
  const handleDeleteAuth = async (): Promise<void> => {
    try {
      await ElMessageBox.confirm('确定要删除该权限吗？删除后无法恢复', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      ElMessage.success('删除成功')
      getMenuList()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error('删除失败')
      }
    }
  }

  /**
   * 切换展开/收起所有菜单
   */
  const toggleExpand = (): void => {
    isExpanded.value = !isExpanded.value
    nextTick(() => {
      if (tableRef.value?.elTableRef && filteredTableData.value) {
        const processRows = (rows: AppRouteRecord[]) => {
          rows.forEach((row) => {
            if (row.children?.length) {
              tableRef.value.elTableRef.toggleRowExpansion(row, isExpanded.value)
              processRows(row.children)
            }
          })
        }
        processRows(filteredTableData.value)
      }
    })
  }
</script>
