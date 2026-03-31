<template>
  <div class="team-members-page art-full-height">
    <template v-if="!team">
      <NoTeamState v-if="teamLoadDone" />
      <ElCard v-else shadow="never" class="team-members-empty">
        <div class="team-members-loading">
          <ElIcon class="is-loading" :size="32"><Loading /></ElIcon>
        </div>
      </ElCard>
    </template>

    <template v-else>
      <AdminWorkspaceHero
        :title="`团队成员（${team.name}）`"
        description="当前团队成员与角色变更均会影响权限快照，请按实际管理员配置执行操作。"
        :metrics="heroMetrics"
      >
        <div class="team-members-hero-actions">
          <ElButton v-if="hasAction('team.member.manage')" type="primary" :loading="addLoading" @click="handleAddMember">
            添加成员
          </ElButton>
        </div>
      </AdminWorkspaceHero>

      <ElCard shadow="never" class="art-table-card team-members-main">
        <section v-if="hasAction('team.member.manage')" class="team-members-add art-card">
          <header class="team-members-add__header">
            <h3>快速添加</h3>
            <p>可直接录入用户 ID 分配团队角色。</p>
          </header>
          <ElForm :model="addForm" label-width="80px" label-position="top" class="team-members-add__form">
            <ElFormItem label="用户 ID" class="mb-0">
              <ElInput
                v-model="addForm.user_id"
                placeholder="请输入用户 ID（UUID）"
                clearable
                @keyup.enter="handleAddMember"
              />
            </ElFormItem>
            <ElFormItem label="团队角色" class="mb-0">
              <ElSelect v-model="addForm.role_code" placeholder="请选择角色">
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
        </section>

        <section class="team-members-table art-card">
          <ArtTableHeader
            layout="refresh"
            :loading="loading"
            @refresh="loadMembers"
          >
            <template #left>
              <div class="team-members-table-summary">
                <strong>成员列表</strong>
                <span>成员角色调整会直接影响当前团队权限快照。</span>
              </div>
            </template>
            <template #right>
              <ElTag effect="plain">{{ members.length }} 人</ElTag>
            </template>
          </ArtTableHeader>
          <ElTable v-loading="loading" :data="pagedMembers" stripe>
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
                        <ElDropdownItem v-if="hasAction('team.member.manage')" command="assign">
                          <ElIcon><UserFilled /></ElIcon>
                          分配角色
                        </ElDropdownItem>
                        <ElDropdownItem
                          v-if="hasAction('team.member.manage')"
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
          <WorkspacePagination
            v-model:current-page="pagination.current"
            v-model:page-size="pagination.size"
            :total="members.length"
            compact
          />
        </section>
      </ElCard>
    </template>

    <MemberRoleDialog ref="roleDialogRef" :member="currentMember" @success="loadMembers" />
  </div>
</template>

<script setup lang="ts">
  import { Loading, MoreFilled, UserFilled, Delete } from '@element-plus/icons-vue'
  import { storeToRefs } from 'pinia'
  import { useAuth } from '@/hooks/core/useAuth'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
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

  defineOptions({ name: 'TeamMembers' })

  const roleDialogRef = ref()
  const currentMember = ref<Api.SystemManage.TeamMemberItem | null>(null)
  const tenantStore = useTenantStore()
  const { hasAction } = useAuth()
  const { currentTenantId, hasTeams } = storeToRefs(tenantStore)

  const team = ref<Api.SystemManage.TeamListItem | null>(null)
  const teamLoadDone = ref(false)
  const members = ref<Api.SystemManage.TeamMemberItem[]>([])
  const hasMemberOperationPermission = computed(
    () => hasAction('team.member.manage')
  )
  const loading = ref(false)
  const addLoading = ref(false)
  const pagination = reactive({
    current: 1,
    size: 10
  })
  const heroMetrics = computed(() => [
    { label: '成员总数', value: members.value.length },
    { label: '管理员', value: members.value.filter((item) => item.roleCode === 'team_admin' || item.role === 'team_admin').length },
    { label: '普通成员', value: members.value.filter((item) => item.roleCode !== 'team_admin' && item.role !== 'team_admin').length }
  ])

  const addForm = reactive({
    user_id: '',
    role_code: 'team_member'
  })

  const pagedMembers = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return members.value.slice(start, start + pagination.size)
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
      pagination.current = 1
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员列表失败')
      members.value = []
      pagination.current = 1
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
      await fetchAddMyTeamMember({ user_id: uid, role_code: addForm.role_code })
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
  .team-members-empty {
    min-height: 320px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .team-members-loading :deep(.el-icon) {
    animation: rotate 1s linear infinite;
  }

  .team-members-main {
    display: grid;
    gap: 14px;
  }

  .team-members-hero-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .team-members-add {
    padding: 14px;
    border-radius: 18px;
    border: 1px solid var(--art-card-border);
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.94));
    box-shadow: var(--art-shadow-sm);
  }

  .team-members-add__header {
    margin-bottom: 12px;
  }

  .team-members-add__header h3 {
    margin: 0;
    font-size: 15px;
    color: var(--art-text-strong);
    font-weight: 700;
  }

  .team-members-add__header p {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--art-text-muted);
  }

  .team-members-add__form {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0 14px;
  }

  .team-members-table {
    border: 1px solid var(--art-card-border);
    border-radius: 18px;
    padding: 14px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.94));
    box-shadow: var(--art-shadow-sm);
  }

  .team-members-table-summary {
    display: grid;
    gap: 4px;
  }

  .team-members-table-summary strong {
    font-size: 15px;
    color: var(--art-text-strong);
    font-weight: 700;
  }

  .team-members-table-summary span {
    font-size: 12px;
    line-height: 1.6;
    color: var(--art-text-muted);
  }

  @keyframes rotate {
    from {
      transform: rotate(0deg);
    }

    to {
      transform: rotate(360deg);
    }
  }

  @media (max-width: 960px) {
    .team-members-add__form {
      grid-template-columns: 1fr;
    }
  }
</style>
