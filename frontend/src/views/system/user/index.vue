<!-- 用户管理页面 -->
<!-- art-full-height 自动计算出页面剩余高度 -->
<!-- art-table-card 一个符合系统样式的 class，同时自动撑满剩余高度 -->
<!-- 更多 useTable 使用示例请移步至 功能示例 下面的高级表格示例或者查看官方文档 -->
<!-- useTable 文档：https://www.artd.pro/docs/zh/guide/hooks/use-table.html -->
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <UserSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></UserSearch>

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增用户</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- 用户弹窗 -->
      <UserDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :user-data="currentUserData"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetUserList,
    fetchDeleteUser,
    fetchCreateUser,
    fetchUpdateUser
  } from '@/api/system-manage'
  import UserSearch from './modules/user-search.vue'
  import UserDialog from './modules/user-dialog.vue'
  import { ElTag, ElMessageBox, ElImage } from 'element-plus'
  import { DialogType } from '@/types'
  import { ElMessage } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'User' })

  type UserListItem = Api.SystemManage.UserListItem
  const userStore = useUserStore()
  const canViewSystemRemark = computed(() => {
    const roles = (userStore.info?.roles || []) as string[]
    return roles.includes('R_SUPER')
  })

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})

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
                : h('div', { class: 'size-9.5 rounded-md bg-gray-200 flex items-center justify-center' }, '头'),
              h('div', { class: 'ml-2' }, [
                h('p', { class: 'user-name' }, row.nickName || row.userName || '-'),
                h('p', { class: 'email' }, `${row.userName || '-'}${row.userEmail ? ` · ${row.userEmail}` : ''}`)
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
            if (!row.userRoles || row.userRoles.length === 0) {
              return h('span', { class: 'text-gray-400' }, '暂无角色')
            }
            return h('div', { class: 'flex flex-wrap gap-1' }, 
              row.userRoles.map((roleCode: string) => 
                h(ElTag, { size: 'small', type: 'info', effect: 'plain' }, () => roleCode)
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
          width: 120,
          fixed: 'right', // 固定列
          formatter: (row) =>
            h('div', [
              h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }),
              h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deleteUser(row)
              })
            ])
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
  const handleSearch = (params: Record<string, any>) => {
    console.log(params)
    // 搜索参数赋值
    Object.assign(searchParams, params)
    getData()
  }

  /**
   * 显示用户弹窗
   */
  const showDialog = (type: DialogType, row?: UserListItem): void => {
    console.log('打开弹窗:', { type, row })
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
   * 处理弹窗提交事件
   */
  const handleDialogSubmit = async (payload?: {
    username: string
    nickname?: string
    email?: string
    password?: string
    status?: string
    phone?: string
    systemRemark?: string
    roleIds?: string[]
  }) => {
    if (!payload) {
      dialogVisible.value = false
      refreshData()
      return
    }
    const isAdd = dialogType.value === 'add'
    try {
      if (isAdd) {
        await fetchCreateUser({
          username: payload.username,
          nickname: payload.nickname,
          email: payload.email,
          password: payload.password || '123456',
          status: payload.status || 'active',
          phone: payload.phone,
          systemRemark: payload.systemRemark,
          roleIds: payload.roleIds
        })
        ElMessage.success('添加成功')
      } else {
        const id = (currentUserData.value as UserListItem).id
        if (!id) return
        await fetchUpdateUser(id, {
          email: payload.email,
          nickname: payload.nickname,
          status: payload.status,
          phone: payload.phone,
          systemRemark: payload.systemRemark,
          roleIds: payload.roleIds
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
    console.log('选中行数据:', selectedRows.value)
  }
</script>
