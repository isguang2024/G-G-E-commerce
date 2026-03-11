<template>
  <div class="team-page art-full-height">
    <TeamSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增团队</ElButton>
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

      <TeamDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :team-data="currentTeamData"
        @submit="handleDialogSubmit"
      />

      <TeamMembersDrawer
        v-model:visible="membersDrawerVisible"
        :team-id="currentTeamId"
        :team-name="currentTeamName"
        @refresh="refreshData"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetTeamList,
    fetchDeleteTeam,
    fetchCreateTeam,
    fetchUpdateTeam
  } from '@/api/team'
  import TeamSearch from './modules/team-search.vue'
  import TeamDialog from './modules/team-dialog.vue'
  import TeamMembersDrawer from './modules/team-members-drawer.vue'
  import { ElTag, ElMessageBox, ElMessage, ElDropdown, ElDropdownMenu, ElDropdownItem, ElIcon } from 'element-plus'
  import { MoreFilled, Edit, Delete, UserFilled } from '@element-plus/icons-vue'
  import { DialogType } from '@/types'

  defineOptions({ name: 'Team' })

  type TeamListItem = Api.SystemManage.TeamListItem

  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentTeamData = ref<Partial<TeamListItem>>({})
  const membersDrawerVisible = ref(false)
  const currentTeamId = ref('')
  const currentTeamName = ref('')

  const searchForm = ref({
    name: undefined as string | undefined,
    status: undefined as string | undefined
  })

  const getStatusConfig = (status: string) => {
    const map: Record<string, { type: 'success' | 'info' | 'warning' | 'danger'; text: string }> = {
      active: { type: 'success', text: '正常' },
      inactive: { type: 'danger', text: '停用' }
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
    core: {
      apiFn: fetchGetTeamList,
      apiParams: {
        current: 1,
        size: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        { prop: 'name', label: '团队名称', minWidth: 140 },
        { prop: 'remark', label: '团队备注', width: 140 },
        {
          prop: 'adminUsers',
          label: '管理员',
          width: 200,
          formatter: (row: TeamListItem) => {
            const admins = (row as any).adminUsers || []
            if (!admins.length) return '-'
            return h('div', { class: 'flex flex-wrap gap-1' }, [
              ...admins.slice(0, 2).map((admin: any) =>
                h(ElTag, { size: 'small', type: 'info' }, () => {
                  return admin.nickname || admin.username || admin.email || admin.user_id.substring(0, 8) + '...'
                })
              ),
              admins.length > 2 ? h(ElTag, { size: 'small' }, () => `+${admins.length - 2}`) : null
            ])
          }
        },
        { prop: 'plan', label: '套餐', width: 100 },
        { prop: 'maxMembers', label: '最大成员数', width: 100 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row) => {
            const cfg = getStatusConfig(row.status)
            return h(ElTag, { type: cfg.type }, () => cfg.text)
          }
        },
        { prop: 'createTime', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 60,
          fixed: 'right',
          formatter: (row) => {
            const dropdown = h(ElDropdown, {
              trigger: 'click',
              onCommand: (cmd: string) => handleOperationCommand(cmd, row)
            }, {
              default: () => h(ElButton, { icon: MoreFilled, circle: true, size: 'small' }),
              dropdown: () => h(ElDropdownMenu, {}, [
                h(ElDropdownItem, { command: 'edit' }, () => [h(ElIcon, {}, () => h(Edit)), '编辑']),
                h(ElDropdownItem, { command: 'view' }, () => [h(ElIcon, {}, () => h(UserFilled)), '查看人员']),
                h(ElDropdownItem, { command: 'delete' }, () => [h(ElIcon, {}, () => h(Delete)), '删除'])
              ])
            })
            return dropdown
          }
        }
      ]
    },
    transform: {
      dataTransformer: (records) => (Array.isArray(records) ? records : [])
    }
  })

  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: TeamListItem) => {
    dialogType.value = type
    currentTeamData.value = row ? { ...row } : {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  const handleOperationCommand = (command: string, row: TeamListItem) => {
    if (command === 'edit') {
      showDialog('edit', row)
    } else if (command === 'view') {
      showMembers(row)
    } else if (command === 'delete') {
      deleteTeam(row)
    }
  }

  const showMembers = (row: TeamListItem) => {
    currentTeamId.value = row.id
    currentTeamName.value = row.name
    membersDrawerVisible.value = true
  }

  const deleteTeam = (row: TeamListItem) => {
    ElMessageBox.confirm(`确定要删除团队「${row.name}」吗？`, '删除团队', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeleteTeam(row.id))
      .then(() => {
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
      })
  }

  const handleDialogSubmit = async (
    payload?: Api.SystemManage.TeamCreateParams | Api.SystemManage.TeamUpdateParams
  ) => {
    if (!payload) {
      dialogVisible.value = false
      refreshData()
      return
    }
    const isAdd = dialogType.value === 'add'
    try {
      if (isAdd) {
        await fetchCreateTeam(payload as Api.SystemManage.TeamCreateParams)
        ElMessage.success('添加成功')
      } else {
        const id = (currentTeamData.value as TeamListItem).id
        if (!id) return
        await fetchUpdateTeam(id, payload as Api.SystemManage.TeamUpdateParams)
        ElMessage.success('更新成功')
      }
      dialogVisible.value = false
      currentTeamData.value = {}
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message || (isAdd ? '添加失败' : '更新失败'))
    }
  }
</script>
