<template>
  <div class="art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="font-medium">系统团队上下文角色</div>
        <div class="mt-2 text-sm text-gray-500">
          这里展示所有需要团队上下文的角色。团队成员页引用的就是这一组角色，权限仍然由系统统一管理。
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
  import { formatScopeLabel, getScopeTagType } from '@/utils/permission/scope'

  defineOptions({ name: 'TeamRolesAndPermissions' })

  type RoleListItem = Api.SystemManage.RoleListItem

  const showSearchBar = ref(false)
  const permissionDialog = ref(false)
  const currentRoleData = ref<RoleListItem | undefined>(undefined)

  const { columns, columnChecks, data, loading, pagination, handleSizeChange, handleCurrentChange, refreshData } =
    useTable({
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
            minWidth: 180,
            formatter: (row: RoleListItem) => {
              const scopes = Array.isArray((row as any).scopes) ? (row as any).scopes : []
              return h(
                'div',
                { class: 'flex flex-wrap gap-1' },
                scopes.map((scope: any) =>
                  h(
                    ElTag,
                    {
                      key: scope.scopeId || scope.scopeCode,
                      type: getScopeTagType(scope.scopeCode, scope.contextKind),
                      size: 'small'
                    },
                    () => formatScopeLabel(scope.scopeCode, scope.scopeName, scope.contextKind)
                  )
                )
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
