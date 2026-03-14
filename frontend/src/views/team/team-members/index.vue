<template>
  <div class="team-members-page art-full-height">
    <template v-if="!team">
      <ElCard v-if="teamLoadDone" shadow="never" class="empty-card">
        <ElEmpty description="您暂无管理的团队">
          <template #description>
            <p>您当前没有作为管理员或负责人的团队，无法在此管理成员。</p>
            <p class="text-gray-500 text-sm mt-2">请联系系统管理员将您加入团队并设置为管理员。</p>
          </template>
        </ElEmpty>
      </ElCard>
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
            <span>
              <span class="font-medium">当前团队：{{ team.name }}</span>
            </span>
          </div>
        </template>

        <div class="flex flex-col gap-4">
          <!-- 添加成员 -->
          <ElCard shadow="never" class="add-member-card">
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
              <ElFormItem label="团队内角色" class="mb-0">
                <ElSelect v-model="addForm.role" placeholder="选择角色" style="width: 140px">
                  <ElOption label="团队管理员" value="team_admin" />
                  <ElOption label="团队成员" value="team_member" />
                </ElSelect>
              </ElFormItem>
              <ElFormItem class="mb-0">
                <ElButton type="primary" :loading="addLoading" @click="handleAddMember"
                  >添加</ElButton
                >
              </ElFormItem>
            </ElForm>
          </ElCard>

          <!-- 成员列表 -->
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
                      v-for="(role, index) in row.roles || [row.role]"
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
                  <ElDropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, row)">
                    <ElButton :icon="MoreFilled" circle size="small" />
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem command="assign">
                          <ElIcon><UserFilled /></ElIcon>
                          分配角色
                        </ElDropdownItem>
                        <ElDropdownItem command="delete" :disabled="isAdmin(row)" divided>
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
  </div>
</template>

<script setup lang="ts">
  import { Loading, MoreFilled, UserFilled, Delete } from '@element-plus/icons-vue'
  import {
    fetchGetMyTeam,
    fetchGetMyTeamMembers,
    fetchAddMyTeamMember,
    fetchRemoveMyTeamMember,
    fetchUpdateMyTeamMemberRole
  } from '@/api/team'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import MemberRoleDialog from './modules/member-role-dialog.vue'

  defineOptions({ name: 'TeamMembers' })

  const roleDialogRef = ref()
  const currentMember = ref<Api.SystemManage.TeamMemberItem | null>(null)

  const team = ref<Api.SystemManage.TeamListItem | null>(null)
  const teamLoadDone = ref(false)
  const members = ref<Api.SystemManage.TeamMemberItem[]>([])
  const loading = ref(false)
  const addLoading = ref(false)

  const addForm = reactive({
    user_id: '',
    role: 'team_member'
  })

  // 判断是否为管理员（团队管理员角色编码为 team_admin）
  function isAdmin(row: Api.SystemManage.TeamMemberItem): boolean {
    const roleCodes = (row as any).roleCodes || []
    return roleCodes.includes('team_admin')
  }

  async function loadMyTeam() {
    teamLoadDone.value = false
    team.value = null
    try {
      const res = await fetchGetMyTeam()
      team.value = res as Api.SystemManage.TeamListItem
      await loadMembers()
    } catch (e: any) {
      if (e?.response?.status === 404 || e?.code === 404) {
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
      const res = await fetchGetMyTeamMembers()
      members.value = res?.records ?? []
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

  function handleCommand(command: string, member: Api.SystemManage.TeamMemberItem) {
    if (command === 'assign') {
      handleAssignRoles(member)
    } else if (command === 'delete') {
      removeMember(member)
    }
  }

  function removeMember(row: Api.SystemManage.TeamMemberItem) {
    if (isAdmin(row)) {
      ElMessage.warning('团队管理员不能被移除')
      return
    }
    ElMessageBox.confirm(`确定将「${row.userName}」移出团队吗？`, '移除成员', {
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
    const uid = addForm.user_id?.trim()
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
</script>

<style scoped>
  .team-members-page .empty-card {
    min-height: 320px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .add-member-card :deep(.el-card__body) {
    padding-bottom: 8px;
  }
</style>
