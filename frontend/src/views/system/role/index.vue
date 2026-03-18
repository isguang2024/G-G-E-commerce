<template>
  <div class="art-full-height">
    <router-view v-if="hasNestedRoute" />
    <template v-else>
      <RoleSearch
        v-show="showSearchBar"
        v-model="searchForm"
        @search="handleSearch"
        @reset="resetSearchParams"
      />

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
              <ElButton v-action="'role:create'" @click="showDialog('add')" v-ripple>
                新增角色
              </ElButton>
            </ElSpace>
          </template>
        </ArtTableHeader>

        <ArtTable
          :loading="loading"
          :data="data"
          :columns="columns"
          :pagination="pagination"
          @pagination:size-change="handleSizeChange"
          @pagination:current-change="handleCurrentChange"
        />
      </ElCard>

      <RoleEditDialog
        v-model="dialogVisible"
        :dialog-type="dialogType"
        :role-data="currentRoleData"
        @success="refreshData"
      />

      <RolePermissionDialog
        v-model="permissionDialog"
        :role-data="currentRoleData"
        @success="onPermissionSuccess"
      />

      <RoleActionPermissionDialog
        v-model="actionPermissionDialog"
        :role-data="currentRoleData"
        @success="refreshData"
      />

      <RoleDataPermissionDialog
        v-model="dataPermissionDialog"
        :role-data="currentRoleData"
        @success="refreshData"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
  import { useRoute } from 'vue-router'
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useAuth } from '@/hooks/core/useAuth'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchDeleteRole, fetchGetRoleList } from '@/api/system-manage'
  import { getScopeTagType } from '@/utils/permission/scope'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import RoleSearch from './modules/role-search.vue'
  import RoleEditDialog from './modules/role-edit-dialog.vue'
  import RolePermissionDialog from './modules/role-permission-dialog.vue'
  import RoleActionPermissionDialog from './modules/role-action-permission-dialog.vue'
  import RoleDataPermissionDialog from './modules/role-data-permission-dialog.vue'
  import { ElTag, ElMessage, ElMessageBox } from 'element-plus'
  import { refreshUserMenus } from '@/router'

  defineOptions({ name: 'Role' })

  const route = useRoute()
  const hasNestedRoute = computed(() => route.matched.length > 2)

  type RoleListItem = Api.SystemManage.RoleListItem
  const { hasAction } = useAuth()

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
  const actionPermissionDialog = ref(false)
  const dataPermissionDialog = ref(false)
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
    core: {
      apiFn: fetchGetRoleList,
      apiParams: {
        current: 1,
        size: 20
      },
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
          minWidth: 220,
          formatter: (row: RoleListItem) => {
            const scopes = Array.isArray((row as any).scopes) ? (row as any).scopes : []
            if (scopes.length === 0) {
              const fallback = (row as any).scopeName || (row as any).scopeCode || '未配置'
              return h(ElTag, { type: 'info', effect: 'plain' }, () => fallback)
            }
            return h(
              'div',
              { class: 'flex flex-wrap gap-1' },
              scopes.map((scope: any) =>
                h(
                  ElTag,
                  {
                    key: scope.scopeId || scope.scopeCode,
                    type: getScopeTagType(scope.scopeCode, scope.contextKind),
                    effect: 'plain'
                  },
                  () => scope.scopeName || scope.scopeCode || '未命名'
                )
              )
            )
          }
        },
        {
          prop: 'status',
          label: '角色状态',
          width: 100,
          formatter: (row: RoleListItem) => {
            const statusConfig =
              row.status === 'normal'
                ? { type: 'success', text: '正常' }
                : { type: 'warning', text: '停用' }
            return h(ElTag, { type: statusConfig.type as 'success' | 'warning' }, () => statusConfig.text)
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
              {
                key: 'permission',
                label: '菜单权限',
                icon: 'ri:user-3-line',
                auth: 'role:assign_menu'
              },
              {
                key: 'actionPermission',
                label: '功能权限',
                icon: 'ri:shield-keyhole-line',
                auth: 'role:assign_action'
              },
              {
                key: 'dataPermission',
                label: '数据权限',
                icon: 'ri:database-2-line',
                auth: 'role:assign_data'
              },
              { key: 'edit', label: '编辑角色', icon: 'ri:edit-2-line', auth: 'role:update' }
            ]
            if (!isDefaultRole && hasAction('role:delete')) {
              list.push({
                key: 'delete',
                label: '删除角色',
                icon: 'ri:delete-bin-4-line',
                auth: 'role:delete'
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
      dataTransformer: (rows: RoleListItem[]) => {
        const scopes = (searchParams as any).scopes as string[] | undefined
        if (!Array.isArray(scopes) || scopes.length === 0) return rows
        const set = new Set(scopes)
        return rows.filter((row) => {
          const rowScopes = Array.isArray((row as any).scopes) ? (row as any).scopes : []
          if (rowScopes.some((item: any) => item.scopeCode && set.has(item.scopeCode))) {
            return true
          }
          return Boolean((row as any).scopeCode && set.has((row as any).scopeCode))
        })
      }
    }
  })

  const dialogType = ref<'add' | 'edit'>('add')

  const showDialog = (type: 'add' | 'edit', row?: RoleListItem) => {
    dialogVisible.value = true
    dialogType.value = type
    currentRoleData.value = row
  }

  const handleSearch = (params: Record<string, any>) => {
    const { daterange, ...filtersParams } = params
    const [startTime, endTime] = Array.isArray(daterange) ? daterange : [null, null]
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
      case 'actionPermission':
        showActionPermissionDialog(row)
        break
      case 'dataPermission':
        showDataPermissionDialog(row)
        break
      case 'delete':
        deleteRole(row)
        break
    }
  }

  const onPermissionSuccess = async () => {
    refreshData()
    await refreshUserMenus()
  }

  const showPermissionDialog = (row?: RoleListItem) => {
    permissionDialog.value = true
    currentRoleData.value = row
  }

  const showActionPermissionDialog = (row?: RoleListItem) => {
    actionPermissionDialog.value = true
    currentRoleData.value = row
  }

  const showDataPermissionDialog = (row?: RoleListItem) => {
    dataPermissionDialog.value = true
    currentRoleData.value = row
  }

  const deleteRole = (row: RoleListItem) => {
    ElMessageBox.confirm(`确定删除角色“${row.roleName}”吗？此操作不可恢复！`, '删除确认', {
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
