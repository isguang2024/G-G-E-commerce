<template>
  <div class="user-page art-full-height">
    <div class="page-top-stack">
      <UserSearch
        v-show="showSearchBar"
        v-model="searchForm"
        @search="handleSearch"
        @reset="handleResetSearch"
      />

      <AdminWorkspaceHero
        title="用户管理"
        description="管理平台账号、角色归属、菜单裁剪和权限测试，先在这里确认平台身份链路。"
        :metrics="summaryMetrics"
      >
        <div class="user-hero-actions">
          <ElSelect
            v-model="selectedAppKey"
            clearable
            filterable
            placeholder="选择 App"
            class="user-app-select"
            @change="handleManagedAppChange"
          >
            <ElOption
              v-for="item in appOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
          <ElButton
            v-action="'system.user.manage'"
            type="primary"
            @click="showDialog('add')"
            v-ripple
          >
            新增用户
          </ElButton>
        </div>
      </AdminWorkspaceHero>
    </div>

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <div class="user-toolbar-tip">角色、功能包、菜单裁剪和权限测试统一从操作菜单进入。</div>
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
        :app-key="targetAppKey"
        @success="refreshData"
      />

      <UserMenuSelectorDialog
        v-model="menuDialogVisible"
        :user-data="currentUserDataForAction"
        :app-key="targetAppKey"
        @success="refreshData"
        @open-packages="openCurrentUserPackagesFromMenu"
      />

      <UserPermissionTestDrawer
        v-model="permissionTestVisible"
        :user-data="currentUserDataForAction"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, watch } from 'vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetUserList,
    fetchDeleteUser,
    fetchCreateUser,
    fetchGetApps,
    fetchUpdateUser
  } from '@/api/system-manage'
  import UserSearch from './modules/user-search.vue'
  import UserDialog from './modules/user-dialog.vue'
  import UserPackageDialog from './modules/user-package-dialog.vue'
  import UserMenuSelectorDialog from './modules/user-menu-selector-dialog.vue'
  import UserPermissionTestDrawer from './modules/user-permission-test-drawer.vue'
  import { ElTag, ElMessageBox, ElImage, ElMessage } from 'element-plus'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'User' })

  type DialogType = 'add' | 'edit'

  type UserListItem = Api.SystemManage.UserListItem
  const userStore = useUserStore()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const canViewSystemRemark = computed(() => {
    const roles = (userStore.info?.roles || []) as string[]
    return roles.includes('R_SUPER')
  })

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const showSearchBar = ref(false)
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})
  const packageDialogVisible = ref(false)
  const menuDialogVisible = ref(false)
  const permissionTestVisible = ref(false)
  const currentUserDataForAction = ref<UserListItem | undefined>(undefined)

  // 选中行
  const selectedRows = ref<UserListItem[]>([])
  const summaryMetrics = computed(() => [
    { label: '当前 App', value: targetAppKey.value },
    { label: '当前页', value: data.value.length || 0 },
    { label: '总用户', value: pagination.total || 0 },
    { label: '已选', value: selectedRows.value.length || 0 }
  ])
  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )

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
      apiFn: async (params) => {
        if (!targetAppKey.value) {
          return {
            records: [],
            total: 0,
            current: Number((params as any)?.current || 1),
            size: Number((params as any)?.size || 20)
          }
        }
        return fetchGetUserList(params as Api.SystemManage.UserSearchParams)
      },
      apiParams: {
        current: 1,
        size: 20,
        appKey: targetAppKey.value,
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
                auth: 'feature_package.assign_collaboration_workspace'
              },
              {
                key: 'menuBoundary',
                label: '菜单裁剪',
                icon: 'ri:menu-line',
                auth: 'system.user.manage'
              },
              {
                key: 'permissionTest',
                label: '权限测试',
                icon: 'ri:shield-keyhole-line',
                auth: 'system.user.manage'
              },
              {
                key: 'edit',
                label: '编辑用户',
                icon: 'ri:edit-2-line',
                auth: 'system.user.manage'
              },
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
    Object.assign(searchParams, { ...searchForm.value, appKey: targetAppKey.value || '' })
    getData()
  }

  async function loadAppOptions() {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }

  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
  }

  function handleResetSearch() {
    resetSearchParams()
    Object.assign(searchParams, {
      current: 1,
      size: pagination.size,
      appKey: targetAppKey.value || ''
    })
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
    } else if (item.key === 'permissionTest') {
      showPermissionTestDialog(row)
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

  const showPermissionTestDialog = (row: UserListItem) => {
    currentUserDataForAction.value = row
    permissionTestVisible.value = true
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

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
    loadAppOptions().catch(() => {
      appList.value = []
    })
  })

  watch(
    () => targetAppKey.value,
    (value) => {
      selectedAppKey.value = value || ''
      Object.assign(searchParams, { appKey: value || '' })
      refreshData()
    }
  )
</script>

<style scoped lang="scss">
  .user-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .user-toolbar-tip {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .user-app-select {
    width: 240px;
  }
</style>
