<template>
  <div class="user-page art-full-height">
    <UserSearch
      v-show="showSearchBar"
      v-model="searchForm"
      @search="handleSearch"
      @reset="resetSearchParams"
    />

    <ElCard
      class="art-table-card"
      shadow="never"
      :style="{ marginTop: showSearchBar ? '12px' : '0' }"
    >
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <ElSpace wrap>
            <ElButton v-action="'system.user.manage'" @click="showDialog('add')" v-ripple>新增用户</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <UserDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :user-data="currentUserData"
        @submit="handleDialogSubmit"
      />

      <UserPackageDialog
        v-model="packageDialogVisible"
        :user-data="currentUserDataForAction"
        @success="refreshData"
      />

      <UserMenuSelectorDialog
        v-model="menuDialogVisible"
        :user-data="currentUserDataForAction"
        @success="refreshData"
        @open-packages="openCurrentUserPackagesFromMenu"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetUserList,
    fetchDeleteUser,
    fetchCreateUser,
    fetchUpdateUser
  } from '@/api/system-manage'
  import UserSearch from './modules/user-search.vue'
  import UserDialog from './modules/user-dialog.vue'
  import UserPackageDialog from './modules/user-package-dialog.vue'
  import UserMenuSelectorDialog from './modules/user-menu-selector-dialog.vue'
  import { ElTag, ElMessageBox, ElImage, ElMessage } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'User' })

  type DialogType = 'add' | 'edit'

  type UserListItem = Api.SystemManage.UserListItem
  const userStore = useUserStore()
  const canViewSystemRemark = computed(() => {
    const roles = (userStore.info?.roles || []) as string[]
    return roles.includes('R_SUPER')
  })

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const showSearchBar = ref(true)
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})
  const packageDialogVisible = ref(false)
  const menuDialogVisible = ref(false)
  const currentUserDataForAction = ref<UserListItem | undefined>(undefined)

  // 选中行
  const selectedRows = ref<UserListItem[]>([])

  // 搜索表单（与后端 status: active/inactive 一致）
  const searchForm = ref({
    userName: undefined as string | undefined,
    userPhone: undefined as string | undefined,
    userEmail: undefined as string | undefined,
    status: undefined as string | undefined,
    roleId: undefined as string | undefined
  })

  // 用户状态配置（与后端 status 一致：active / inactive）
  const getUserStatusConfig = (status: string) => {
    const map: Record<string, { type: 'success' | 'info' | 'warning' | 'danger'; text: string }> = {
      active: { type: 'success', text: '正常' },
      inactive: { type: 'danger', text: '禁用' }
    }
    return map[status] || { type: 'info', text: status || '未知' }
  }

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
      apiFn: fetchGetUserList,
      apiParams: {
        current: 1,
        size: 20,
        ...searchForm.value
      },
      // 自定义分页字段映射，未设置时将使用全局配置 tableConfig.ts 中的 paginationKey
      // paginationKey: {
      //   current: 'pageNum',
      //   size: 'pageSize'
      // },
      columnsFactory: () => [
        { type: 'selection' }, // 勾选列
        { type: 'index', width: 60, label: '序号' }, // 序号
        {
          prop: 'userInfo',
          label: '名称',
          minWidth: 260,
          formatter: (row) => {
            const avatar = row.avatar || ''
            return h('div', { class: 'user flex-c' }, [
              avatar
                ? h(ElImage, {
                    class: 'size-9.5 rounded-md',
                    src: avatar,
                    previewSrcList: [avatar],
                    previewTeleported: true
                  })
                : h(
                    'div',
                    { class: 'size-9.5 rounded-md bg-gray-200 flex items-center justify-center' },
                    '头'
                  ),
              h('div', { class: 'ml-2' }, [
                h('p', { class: 'user-name' }, row.nickName || row.userName || '-'),
                h(
                  'p',
                  { class: 'email' },
                  `${row.userName || '-'}${row.userEmail ? ` · ${row.userEmail}` : ''}`
                )
              ])
            ])
          }
        },
        {
          prop: 'status',
          label: '状态',
          formatter: (row) => {
            const statusConfig = getUserStatusConfig(row.status)
            return h(ElTag, { type: statusConfig.type }, () => statusConfig.text)
          }
        },
        {
          prop: 'userRoles',
          label: '已绑定角色',
          minWidth: 200,
          formatter: (row: UserListItem) => {
            const details = row.roleDetails || []
            if (!details || details.length === 0) {
              return h('span', { class: 'text-gray-400' }, '暂无角色')
            }
            return h(
              'div',
              { class: 'flex flex-wrap gap-1' },
              details.map((role: { code: string; name: string }) =>
                h(
                  ElTag,
                  {
                    size: 'small',
                    type: 'info',
                    effect: 'plain',
                    title: role.name || role.code
                  },
                  () => role.code
                )
              )
            )
          }
        },
        {
          prop: 'lastLoginTime',
          label: '最近登录时间',
          minWidth: 180,
          formatter: (row) => row.lastLoginTime || '-'
        },
        {
          prop: 'lastLoginIP',
          label: '最近登录IP',
          minWidth: 140,
          formatter: (row) => row.lastLoginIP || '-'
        },
        {
          prop: 'registerSource',
          label: '注册来源',
          width: 100,
          formatter: (row: UserListItem) => {
            const sourceMap: Record<
              string,
              { type: 'primary' | 'success' | 'warning' | 'info'; text: string }
            > = {
              admin: { type: 'primary', text: '管理员添加' },
              self: { type: 'success', text: '自注册' },
              invite: { type: 'warning', text: '邀请注册' }
            }
            const sourceKey = row.registerSource || ''
            const config = sourceMap[sourceKey] || { type: 'info', text: row.registerSource || '-' }
            return h(ElTag, { type: config.type, size: 'small' }, () => config.text)
          }
        },
        {
          prop: 'invitedBy',
          label: '邀请人',
          width: 120,
          formatter: (row: UserListItem) => row.invitedByName || row.invitedBy || '-'
        },
        ...(canViewSystemRemark.value
          ? [
              {
                prop: 'systemRemark',
                label: '系统备注',
                minWidth: 220,
                formatter: (row: UserListItem) => row.systemRemark || '-'
              }
            ]
          : []),
        {
          prop: 'operation',
          label: '操作',
          width: 60,
          fixed: 'right',
          formatter: (row) => {
            const list = [
              { key: 'copy', label: '复制用户ID', icon: 'ri:file-copy-line' },
              {
                key: 'packages',
                label: '功能包',
                icon: 'ri:apps-2-line',
                auth: 'platform.package.assign'
              },
              {
                key: 'menuBoundary',
                label: '菜单裁剪',
                icon: 'ri:menu-line',
                auth: 'system.user.manage'
              },
              { key: 'edit', label: '编辑用户', icon: 'ri:edit-2-line', auth: 'system.user.manage' },
              {
                key: 'delete',
                label: '删除用户',
                icon: 'ri:delete-bin-4-line',
                color: '#f56c6c',
                auth: 'system.user.manage'
              }
            ]
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleUserOperation(item, row)
            })
          }
        }
      ]
    },
    transform: {
      dataTransformer: (records) => {
        if (!Array.isArray(records)) return []
        return records
      }
    }
  })

  /**
   * 搜索处理
   * @param params 参数
   */
  const handleSearch = () => {
    // 搜索参数赋值
    Object.assign(searchParams, searchForm.value)
    getData()
  }

  /**
   * 显示用户弹窗
   */
  const showDialog = (type: DialogType, row?: UserListItem): void => {
    dialogType.value = type
    currentUserData.value = row || {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /**
   * 删除用户
   */
  const deleteUser = (row: UserListItem): void => {
    ElMessageBox.confirm(`确定要注销该用户吗？`, '注销用户', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })
      .then(() => fetchDeleteUser(row.id))
      .then(() => {
        ElMessage.success('注销成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '注销失败')
      })
  }

  /**
   * 处理用户操作
   */
  const handleUserOperation = (item: ButtonMoreItem, row: UserListItem) => {
    if (item.key === 'copy') {
      navigator.clipboard.writeText(row.id)
      ElMessage.success('用户ID已复制')
    } else if (item.key === 'packages') {
      showPackageDialog(row)
    } else if (item.key === 'menuBoundary') {
      showMenuDialog(row)
    } else if (item.key === 'edit') {
      showDialog('edit', row)
    } else if (item.key === 'delete') {
      deleteUser(row)
    }
  }

  const showPackageDialog = (row: UserListItem) => {
    currentUserDataForAction.value = row
    packageDialogVisible.value = true
  }

  const showMenuDialog = (row: UserListItem) => {
    currentUserDataForAction.value = row
    menuDialogVisible.value = true
  }

  const openCurrentUserPackagesFromMenu = () => {
    if (!currentUserDataForAction.value) return
    menuDialogVisible.value = false
    packageDialogVisible.value = true
  }

  /**
   * 处理弹窗提交事件
   */
  type UserDialogPayload = Api.SystemManage.UserCreateParams | Api.SystemManage.UserUpdateParams

  const handleDialogSubmit = async (payload?: UserDialogPayload) => {
    if (!payload) {
      dialogVisible.value = false
      refreshData()
      return
    }
    const isAdd = dialogType.value === 'add'
    try {
      if (isAdd) {
        const createPayload = payload as Api.SystemManage.UserCreateParams
        await fetchCreateUser({
          username: createPayload.username,
          nickname: createPayload.nickname,
          email: createPayload.email,
          password: createPayload.password || '123456',
          status: createPayload.status || 'active',
          phone: createPayload.phone,
          systemRemark: createPayload.systemRemark,
          roleIds: createPayload.roleIds
        })
        ElMessage.success('添加成功')
      } else {
        const updatePayload = payload as Api.SystemManage.UserUpdateParams
        const id = (currentUserData.value as UserListItem).id
        if (!id) return
        await fetchUpdateUser(id, {
          email: updatePayload.email,
          nickname: updatePayload.nickname,
          status: updatePayload.status,
          phone: updatePayload.phone,
          systemRemark: updatePayload.systemRemark,
          roleIds: updatePayload.roleIds
        })
        ElMessage.success('更新成功')
      }
      dialogVisible.value = false
      currentUserData.value = {}
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message || (isAdd ? '添加失败' : '更新失败'))
    }
  }

  /**
   * 处理表格行选择变化
   */
  const handleSelectionChange = (selection: UserListItem[]): void => {
    selectedRows.value = selection
  }
</script>
