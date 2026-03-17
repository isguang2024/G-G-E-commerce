<template>
  <div class="team-members-page art-full-height">
    <template v-if="!team">
      <NoTeamState v-if="teamLoadDone" />
      <ElCard v-else shadow="never" class="empty-card">
        <div class="flex justify-center py-8">
          <ElIcon class="is-loading" :size="32"><Loading /></ElIcon>
        </div>
      </ElCard>
    </template>

    <template v-else>
      <ElCard shadow="never" class="art-table-card">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-medium">当前团队：{{ team.name }}</span>
          </div>
        </template>

        <div class="flex flex-col gap-4">
          <ElCard v-if="hasAction('team_member:create')" shadow="never" class="add-member-card">
            <template #header>
              <span>添加成员</span>
            </template>
            <ElForm
              :model="addForm"
              label-width="80px"
              label-position="top"
              class="flex flex-wrap items-end gap-4"
            >
              <ElFormItem label="用户 ID" class="mb-0">
                <ElInput
                  v-model="addForm.user_id"
                  placeholder="请输入用户 ID（UUID）"
                  clearable
                  style="width: 280px"
                  @keyup.enter="handleAddMember"
                />
              </ElFormItem>
              <ElFormItem label="团队角色" class="mb-0">
                <ElSelect v-model="addForm.role" placeholder="请选择角色" style="width: 140px">
                  <ElOption label="团队管理员" value="team_admin" />
                  <ElOption label="团队成员" value="team_member" />
                </ElSelect>
              </ElFormItem>
              <ElFormItem class="mb-0">
                <ElButton type="primary" :loading="addLoading" @click="handleAddMember">
                  添加
                </ElButton>
              </ElFormItem>
            </ElForm>
          </ElCard>

          <ElCard shadow="never">
            <template #header>
              <span>成员列表（{{ members.length }}）</span>
            </template>
            <ElTable v-loading="loading" :data="members" stripe>
              <ElTableColumn prop="userName" label="用户名" min-width="100" />
              <ElTableColumn prop="nickName" label="昵称" width="100" />
              <ElTableColumn prop="userEmail" label="邮箱" min-width="140" show-overflow-tooltip />
              <ElTableColumn label="团队身份" min-width="200">
                <template #default="{ row }">
                  <div class="flex flex-wrap gap-1">
                    <ElTag
                      v-for="(role, index) in row.roles || [formatRoleLabel(row.roleCode || row.role)]"
                      :key="index"
                      :type="role === '团队管理员' ? 'success' : 'info'"
                      size="small"
                    >
                      {{ role }}
                    </ElTag>
                  </div>
                </template>
              </ElTableColumn>
              <ElTableColumn prop="joinedAt" label="加入时间" width="170" />
              <ElTableColumn label="操作" width="60" fixed="right">
                <template #default="{ row }">
                  <ElDropdown
                    v-if="hasMemberOperationPermission"
                    trigger="click"
                    @command="(cmd: string) => handleCommand(cmd, row)"
                  >
                    <ElButton :icon="MoreFilled" circle size="small" />
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem v-if="hasAction('team_member:assign_role')" command="assign">
                          <ElIcon><UserFilled /></ElIcon>
                          分配角色
                        </ElDropdownItem>
                        <ElDropdownItem v-if="hasAction('team_member:assign_action')" command="action">
                          <ElIcon><UserFilled /></ElIcon>
                          功能权限
                        </ElDropdownItem>
                        <ElDropdownItem
                          v-if="hasAction('team_member:delete')"
                          command="delete"
                          :disabled="isAdmin(row)"
                          divided
                        >
                          <ElIcon><Delete /></ElIcon>
                          删除
                        </ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </template>
              </ElTableColumn>
            </ElTable>
          </ElCard>
        </div>
      </ElCard>
    </template>

    <MemberRoleDialog ref="roleDialogRef" :member="currentMember" @success="loadMembers" />
    <MemberActionDialog ref="actionDialogRef" :member="currentMember" @success="loadMembers" />
  </div>
</template>

<script setup lang="ts">
  import { Loading, MoreFilled, UserFilled, Delete } from '@element-plus/icons-vue'
  import { storeToRefs } from 'pinia'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    fetchGetMyTeam,
    fetchGetMyTeamMembers,
    fetchAddMyTeamMember,
    fetchRemoveMyTeamMember
  } from '@/api/team'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useTenantStore } from '@/store/modules/tenant'
  import NoTeamState from '@/components/business/team/NoTeamState.vue'
  import MemberRoleDialog from './modules/member-role-dialog.vue'
  import MemberActionDialog from './modules/member-action-dialog.vue'

  defineOptions({ name: 'TeamMembers' })

  const roleDialogRef = ref()
  const actionDialogRef = ref()
  const currentMember = ref<Api.SystemManage.TeamMemberItem | null>(null)
  const tenantStore = useTenantStore()
  const { hasAction } = useAuth()
  const { currentTenantId, hasTeams } = storeToRefs(tenantStore)

  const team = ref<Api.SystemManage.TeamListItem | null>(null)
  const teamLoadDone = ref(false)
  const members = ref<Api.SystemManage.TeamMemberItem[]>([])
  const hasMemberOperationPermission = computed(
    () =>
      hasAction('team_member:assign_role') ||
      hasAction('team_member:assign_action') ||
      hasAction('team_member:delete')
  )
  const loading = ref(false)
  const addLoading = ref(false)

  const addForm = reactive({
    user_id: '',
    role: 'team_member'
  })

  function isAdmin(row: Api.SystemManage.TeamMemberItem): boolean {
    return row.roleCode === 'team_admin' || row.role === 'team_admin'
  }

  function formatRoleLabel(roleCode?: string) {
    return roleCode === 'team_admin' ? '团队管理员' : '团队成员'
  }

  async function loadMyTeam() {
    teamLoadDone.value = false
    team.value = null
    members.value = []

    if (!hasTeams.value) {
      teamLoadDone.value = true
      return
    }

    try {
      team.value = await fetchGetMyTeam()
      await loadMembers()
    } catch (e: any) {
      if ([400, 404, 3006].includes(e?.response?.status) || [400, 404, 3006].includes(e?.code)) {
        team.value = null
      } else {
        ElMessage.error(e?.message || '获取团队信息失败')
      }
    } finally {
      teamLoadDone.value = true
    }
  }

  async function loadMembers() {
    if (!team.value) return
    loading.value = true
    try {
      members.value = await fetchGetMyTeamMembers()
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员列表失败')
      members.value = []
    } finally {
      loading.value = false
    }
  }

  function handleAssignRoles(member: Api.SystemManage.TeamMemberItem) {
    currentMember.value = member
    nextTick(() => {
      roleDialogRef.value?.open()
    })
  }

  function handleAssignActions(member: Api.SystemManage.TeamMemberItem) {
    currentMember.value = member
    nextTick(() => {
      actionDialogRef.value?.open()
    })
  }

  function handleCommand(command: string, member: Api.SystemManage.TeamMemberItem) {
    if (command === 'assign') {
      handleAssignRoles(member)
    } else if (command === 'action') {
      handleAssignActions(member)
    } else if (command === 'delete') {
      removeMember(member)
    }
  }

  function removeMember(row: Api.SystemManage.TeamMemberItem) {
    if (isAdmin(row)) {
      ElMessage.warning('团队管理员不能被移除')
      return
    }

    ElMessageBox.confirm(`确定将“${row.userName}”移出团队吗？`, '移除成员', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchRemoveMyTeamMember(row.userId))
      .then(() => {
        ElMessage.success('已移除')
        loadMembers()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '移除失败')
      })
  }

  async function handleAddMember() {
    const uid = addForm.user_id.trim()
    if (!uid) {
      ElMessage.warning('请输入用户 ID')
      return
    }

    addLoading.value = true
    try {
      await fetchAddMyTeamMember({ user_id: uid, role: addForm.role })
      ElMessage.success('添加成功')
      addForm.user_id = ''
      await loadMembers()
    } catch (e: any) {
      ElMessage.error(e?.message || '添加失败')
    } finally {
      addLoading.value = false
    }
  }

  onMounted(() => {
    loadMyTeam()
  })

  watch(currentTenantId, (tenantId, oldTenantId) => {
    if (tenantId === oldTenantId) return
    loadMyTeam()
  })
</script>

<style scoped>
  .empty-card {
    min-height: 320px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .add-member-card :deep(.el-card__body) {
    padding-bottom: 8px;
  }
</style>
