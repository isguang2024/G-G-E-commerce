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
              <span class="text-gray-500 text-sm ml-2">（{{ team.remark }}）</span>
            </span>
          </div>
        </template>

        <div class="flex flex-col gap-4">
          <!-- 添加成员 -->
          <ElCard shadow="never" class="add-member-card">
            <template #header>
              <span>添加成员</span>
            </template>
            <ElForm :model="addForm" label-width="80px" label-position="top" class="flex flex-wrap items-end gap-4">
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
                <ElSelect v-model="addForm.role" placeholder="选择角色" style="width: 120px">
                  <ElOption label="团队管理员" value="team_admin" />
                  <ElOption label="团队成员" value="team_member" />
                </ElSelect>
              </ElFormItem>
              <ElFormItem class="mb-0">
                <ElButton type="primary" :loading="addLoading" @click="handleAddMember">添加</ElButton>
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
              <ElTableColumn prop="role" label="团队身份" width="120">
                <template #default="{ row }">
                  <span v-if="row.role === 'team_admin'" class="text-gray-500">团队管理员</span>
                  <ElSelect
                    v-else
                    v-model="row.role"
                    size="small"
                    style="width: 100px"
                    @change="(val: string) => updateMemberRole(row.userId, val)"
                  >
                    <ElOption label="团队管理员" value="team_admin" />
                    <ElOption label="团队成员" value="team_member" />
                  </ElSelect>
                </template>
              </ElTableColumn>
              <ElTableColumn label="功能角色" min-width="200">
                <template #default="{ row }">
                  <div class="flex flex-wrap gap-1 items-center">
                    <template v-if="memberRoleNames[row.userId]?.length">
                      <ElTag
                        v-for="name in memberRoleNames[row.userId]"
                        :key="name"
                        size="small"
                        type="info"
                        effect="plain"
                      >
                        {{ name }}
                      </ElTag>
                    </template>
                    <span v-else class="text-gray-400 text-xs">暂无功能角色</span>
                    <ElButton
                      type="primary"
                      link
                      size="small"
                      class="ml-1"
                      @click="handleAssignRoles(row)"
                    >
                      分配
                    </ElButton>
                  </div>
                </template>
              </ElTableColumn>
              <ElTableColumn prop="joinedAt" label="加入时间" width="170" />
              <ElTableColumn label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <ElButton
                    type="danger"
                    link
                    size="small"
                    @click="removeMember(row)"
                  >
                    删除
                  </ElButton>
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
  import { Loading } from '@element-plus/icons-vue'
  import {
    fetchGetMyTeam,
    fetchGetMyTeamMembers,
    fetchGetMyTeamMemberRoles,
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
  const memberRoleNames = ref<Record<string, string[]>>({})

  const addForm = reactive({
    user_id: '',
    role: 'team_member'
  })

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
      memberRoleNames.value = {}
      
      // 获取每个成员的角色信息（用于展示名称）
      await Promise.all(
        members.value.map(async (m) => {
          try {
            const r = await fetchGetMyTeamMemberRoles(m.userId)
            // 这里我们需要 role_names，但后端目前只返回了 role_ids
            // 为了优化，我们暂且显示 ID 的缩写或等待后端完善
            // 建议：后续后端接口直接返回 Role 对象数组
            memberRoleNames.value[m.userId] = r?.role_ids?.map(id => id.split('-')[0]) ?? []
          } catch {
            memberRoleNames.value[m.userId] = []
          }
        })
      )
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

  function removeMember(row: Api.SystemManage.TeamMemberItem) {
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

  async function updateMemberRole(userId: string, role: string) {
    try {
      await fetchUpdateMyTeamMemberRole(userId, role)
      ElMessage.success('角色已更新')
      await loadMembers()
    } catch (e: any) {
      ElMessage.error(e?.message || '更新角色失败')
      await loadMembers()
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
