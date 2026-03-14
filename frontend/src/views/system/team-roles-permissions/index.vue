<!-- 团队角色及权限：仅管理系统全局角色，所有团队获取角色时都会携带这些全局角色，团队不可修改其权限 -->
<template>
  <div class="art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="font-medium">系统全局角色</div>
        <div class="mt-2 text-sm text-gray-500">
          此处管理的角色为系统全局角色，所有团队在「团队角色」页获取角色时都会携带这些角色；团队不能修改全局角色的权限。默认角色「团队管理员」「团队成员」权限固定，不可删除。
        </div>
      </template>

      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="refreshData"
      />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <RolePermissionDialog
      v-model="permissionDialog"
      :role-data="currentRoleData"
      @success="onPermissionSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetRoleList } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import RolePermissionDialog from '../role/modules/role-permission-dialog.vue'
  import { ElTag } from 'element-plus'
  import { refreshUserMenus } from '@/router'

  defineOptions({ name: 'TeamRolesAndPermissions' })

  type RoleListItem = Api.SystemManage.RoleListItem

  const showSearchBar = ref(false)
  const permissionDialog = ref(false)
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
        size: 20,
        globalOnly: true
      },
      excludeParams: [],
      columnsFactory: () => [
        { prop: 'roleName', label: '角色名称', minWidth: 120 },
        { prop: 'roleCode', label: '角色编码', minWidth: 120 },
        { prop: 'description', label: '描述', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'scope',
          label: '作用域',
          width: 100,
          formatter: (row: RoleListItem) => {
            const scopeConfig =
              row.scope === 'global'
                ? { type: 'primary', text: '全局' }
                : { type: 'success', text: '团队' }
            return h(
              ElTag,
              { type: scopeConfig.type as 'primary' | 'success', size: 'small' },
              () => scopeConfig.text
            )
          }
        },
        { prop: 'createTime', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 100,
          fixed: 'right',
          formatter: (row: RoleListItem) => {
            const list = [{ key: 'permission', label: '菜单权限', icon: 'ri:user-3-line' }]
            return h('div', [
              h(ArtButtonMore, {
                list,
                onClick: (item: ButtonMoreItem) => buttonMoreClick(item, row)
              })
            ])
          }
        }
      ]
    }
  })

  const buttonMoreClick = (item: ButtonMoreItem, row: RoleListItem) => {
    if (item.key === 'permission') {
      permissionDialog.value = true
      currentRoleData.value = row
    }
  }

  const onPermissionSuccess = async () => {
    refreshData()
    await refreshUserMenus()
  }
</script>
