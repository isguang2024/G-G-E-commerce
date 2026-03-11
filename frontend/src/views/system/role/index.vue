<!-- 角色管理页面：有子路由（如内页）时渲染 router-view，否则显示本页内容 -->
<template>
  <div class="art-full-height">
    <!-- 当前为子路由（如 /system/role/role1）时，只渲染子路由组件 -->
    <router-view v-if="hasNestedRoute" />
    <template v-else>
      <RoleSearch
        v-show="showSearchBar"
        v-model="searchForm"
        @search="handleSearch"
        @reset="resetSearchParams"
      ></RoleSearch>

      <ElCard
      class="art-table-card"
      shadow="never"
      :style="{ 'margin-top': showSearchBar ? '12px' : '0' }"
    >
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增角色</ElButton>
            <ElButton @click="showScopeDialog" v-ripple>作用域管理</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>
    </ElCard>

    <!-- 角色编辑弹窗 -->
    <RoleEditDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :role-data="currentRoleData"
      @success="refreshData"
    />

    <!-- 菜单权限弹窗 -->
    <RolePermissionDialog
      v-model="permissionDialog"
      :role-data="currentRoleData"
      @success="onPermissionSuccess"
    />

      <!-- 作用域管理弹窗 -->
      <ScopeManageDialog v-model="scopeDialogVisible" />
    </template>
  </div>
</template>

<script setup lang="ts">
  import { useRoute } from 'vue-router'
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetRoleList } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import RoleSearch from './modules/role-search.vue'
  import RoleEditDialog from './modules/role-edit-dialog.vue'
  import RolePermissionDialog from './modules/role-permission-dialog.vue'
  import ScopeManageDialog from './modules/scope-manage-dialog.vue'
  import { fetchDeleteRole } from '@/api/system-manage'
  import { ElTag, ElMessage, ElMessageBox } from 'element-plus'
  import { refreshUserMenus } from '@/router'

  defineOptions({ name: 'Role' })

  const route = useRoute()
  /** 当前匹配到的是本页下的子路由（如内页），应渲染 router-view 而非本页表格 */
  const hasNestedRoute = computed(() => route.matched.length > 2)

  type RoleListItem = Api.SystemManage.RoleListItem

  // 搜索表单
  const searchForm = ref({
    roleName: undefined,
    roleCode: undefined,
    description: undefined,
    enabled: undefined,
    daterange: undefined,
    scopes: undefined as string[] | undefined
  })

  const showSearchBar = ref(false)

  const dialogVisible = ref(false)
  const permissionDialog = ref(false)
  const scopeDialogVisible = ref(false)
  const currentRoleData = ref<RoleListItem | undefined>(undefined)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    // 核心配置
    core: {
      apiFn: fetchGetRoleList,
      apiParams: {
        current: 1,
        size: 20
      },
      // 排除 apiParams 中的属性
      excludeParams: ['daterange'],
      columnsFactory: () => [
        {
          prop: 'roleId',
          label: '角色ID',
          minWidth: 280,
          showOverflowTooltip: true
        },
        {
          prop: 'roleName',
          label: '角色名称',
          minWidth: 120
        },
        {
          prop: 'roleCode',
          label: '角色编码',
          minWidth: 120
        },
        {
          prop: 'description',
          label: '角色描述',
          minWidth: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'scope',
          label: '作用域',
          width: 100,
          formatter: (row) => {
            const scopeName = row.scopeName || (row.scope === 'global' ? '全局' : row.scope === 'team' ? '团队' : '')
            const scopeCode = row.scopeCode || row.scope || ''
            const scopeConfig = scopeCode === 'global'
              ? { type: 'primary', text: scopeName || '全局' }
              : scopeCode === 'team'
              ? { type: 'success', text: scopeName || '团队' }
              : { type: 'info', text: scopeName || '未知' }
            return h(
              ElTag,
              { type: scopeConfig.type as 'primary' | 'success' | 'info' },
              () => scopeConfig.text
            )
          }
        },
        {
          prop: 'status',
          label: '角色状态',
          width: 100,
          formatter: (row: RoleListItem) => {
            const statusConfig = row.status === 'normal'
              ? { type: 'success', text: '正常' }
              : { type: 'warning', text: '停用' }
            return h(
              ElTag,
              { type: statusConfig.type as 'success' | 'warning' },
              () => statusConfig.text
            )
          }
        },
        {
          prop: 'priority',
          label: '优先级',
          width: 80,
          formatter: (row: RoleListItem) => row.priority || 0
        },
        {
          prop: 'createTime',
          label: '创建日期',
          width: 180,
          sortable: true
        },
        {
          prop: 'operation',
          label: '操作',
          width: 80,
          fixed: 'right',
          formatter: (row) => {
            const isDefaultRole = ['admin', 'team_admin', 'team_member'].includes(row.roleCode)
            const list = [
              { key: 'permission', label: '菜单权限', icon: 'ri:user-3-line' },
              { key: 'edit', label: '编辑角色', icon: 'ri:edit-2-line' }
            ]
            if (!isDefaultRole) {
              list.push({
                key: 'delete',
                label: '删除角色',
                icon: 'ri:delete-bin-4-line'
              })
            }
            return h('div', [
              h(ArtButtonMore, {
                list,
                onClick: (item: ButtonMoreItem) => buttonMoreClick(item, row)
              })
            ])
          }
        }
      ]
    },
    transform: {
      // 前端按作用域多选过滤（按 scopeCode 匹配），不额外增加后端参数
      dataTransformer: (rows: RoleListItem[]) => {
        const scopes = (searchParams as any).scopes as string[] | undefined
        if (!Array.isArray(scopes) || scopes.length === 0) return rows
        const set = new Set(scopes)
        return rows.filter((row) => row.scopeCode && set.has(row.scopeCode))
      }
    }
  })

  const dialogType = ref<'add' | 'edit'>('add')

  const showDialog = (type: 'add' | 'edit', row?: RoleListItem) => {
    dialogVisible.value = true
    dialogType.value = type
    currentRoleData.value = row
  }

  /**
   * 搜索处理
   * @param params 搜索参数
   */
  const handleSearch = (params: Record<string, any>) => {
    // 处理日期区间参数，把 daterange 转换为 startTime 和 endTime
    const { daterange, ...filtersParams } = params
    const [startTime, endTime] = Array.isArray(daterange) ? daterange : [null, null]

    // 搜索参数赋值
    Object.assign(searchParams, { ...filtersParams, startTime, endTime })
    getData()
  }

  const buttonMoreClick = (item: ButtonMoreItem, row: RoleListItem) => {
    switch (item.key) {
      case 'permission':
        showPermissionDialog(row)
        break
      case 'edit':
        showDialog('edit', row)
        break
      case 'delete':
        deleteRole(row)
        break
    }
  }

  /** 菜单权限保存成功：刷新表格并刷新当前用户菜单/路由，使新勾选的菜单立即在侧栏生效 */
  const onPermissionSuccess = async () => {
    refreshData()
    await refreshUserMenus()
  }

  const showPermissionDialog = (row?: RoleListItem) => {
    permissionDialog.value = true
    currentRoleData.value = row
  }

  const showScopeDialog = () => {
    scopeDialogVisible.value = true
  }

  const deleteRole = (row: RoleListItem) => {
    ElMessageBox.confirm(`确定删除角色"${row.roleName}"吗？此操作不可恢复！`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeleteRole(row.roleId))
      .then(() => {
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error((e as any)?.message || '删除失败')
      })
  }
</script>
