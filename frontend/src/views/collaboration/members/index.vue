<template>
  <div class="collaboration-members-page art-full-height">
    <template v-if="!currentCollaboration">
      <NoCollaborationState v-if="collaborationLoadDone" />
      <ElCard v-else shadow="never" class="collaboration-members-empty">
        <div class="collaboration-members-loading">
          <ElIcon class="is-loading" :size="32"><Loading /></ElIcon>
        </div>
      </ElCard>
    </template>

    <template v-else>
      <AdminWorkspaceHero
        :title="heroTitle"
        description="这里区分成员身份与协作空间内部角色：成员身份决定协作空间关系边界，协作空间内部角色决定当前协作空间内的权限。"
        :metrics="heroMetrics"
      >
        <div class="collaboration-members-hero-actions">
          <ElButton
            v-if="hasAction('collaboration.member.manage')"
            type="primary"
            :loading="addLoading"
            @click="handleAddMember"
          >
            添加成员
          </ElButton>
        </div>
      </AdminWorkspaceHero>

      <ElCard shadow="never" class="art-table-card collaboration-members-main">
        <section
          v-if="hasAction('collaboration.member.manage')"
          class="collaboration-members-add art-card"
        >
          <header class="collaboration-members-add__header">
            <h3>快速添加</h3>
            <p>可直接录入用户 ID 设置成员身份。</p>
          </header>
          <ElForm
            :model="addForm"
            label-width="80px"
            label-position="top"
            class="collaboration-members-add__form"
          >
            <ElFormItem label="用户 ID" class="mb-0">
              <ElInput
                v-model="addForm.user_id"
                placeholder="请输入用户 ID（UUID）"
                clearable
                @keyup.enter="handleAddMember"
              />
            </ElFormItem>
            <ElFormItem label="成员身份" class="mb-0">
              <ElSelect v-model="addForm.role_code" placeholder="请选择角色">
                <ElOption label="协作空间管理员" value="collaboration_admin" />
                <ElOption label="协作空间成员" value="collaboration_member" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem class="mb-0">
              <ElButton type="primary" :loading="addLoading" @click="handleAddMember">
                添加
              </ElButton>
            </ElFormItem>
          </ElForm>
        </section>

        <section class="collaboration-members-table art-card">
          <ArtTableHeader layout="refresh" :loading="loading" @refresh="loadMembers">
            <template #left>
              <div class="collaboration-members-table-summary">
                <strong>成员列表</strong>
                <span
                  >成员身份变化会影响协作空间边界；协作空间内部角色变化会影响当前协作空间权限快照。</span
                >
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
            <ElTableColumn label="协作空间身份" min-width="200">
              <template #default="{ row }">
                <div class="flex flex-wrap gap-1">
                  <ElTag
                    v-for="(role, index) in row.roles || [
                      formatRoleLabel(row.roleCode || row.role)
                    ]"
                    :key="index"
                    :type="role === '协作空间管理员' ? 'success' : 'info'"
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
                      <ElDropdownItem
                        v-if="hasAction('collaboration.member.manage')"
                        command="assign"
                      >
                        <ElIcon><UserFilled /></ElIcon>
                        配置协作空间内部角色
                      </ElDropdownItem>
                      <ElDropdownItem
                        v-if="hasAction('collaboration.member.manage')"
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
    fetchGetMyCollaboration,
    fetchGetMyCollaborationMembers,
    fetchAddMyCollaborationMember,
    fetchRemoveMyCollaborationMember
  } from '@/api/collaboration'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useCollaborationStore } from '@/store/modules/collaboration'
  import { useWorkspaceStore } from '@/store/modules/workspace'
  import NoCollaborationState from '@/components/business/collaboration/NoCollaborationState.vue'
  import MemberRoleDialog from './modules/collaboration-member-role-dialog.vue'

  defineOptions({ name: 'CollaborationMembers' })

  const roleDialogRef = ref()
  type WorkspaceItem = Api.SystemManage.CollaborationWorkspaceListItem
  type MemberItem = Api.SystemManage.CollaborationWorkspaceMemberItem

  const currentMember = ref<MemberItem | null>(null)
  const collaborationStore = useCollaborationStore()
  const workspaceStore = useWorkspaceStore()
  const { hasAction } = useAuth()
  const { hasCollaborations } = storeToRefs(collaborationStore)
  const { currentAuthWorkspace, currentAuthWorkspaceId, currentAuthWorkspaceType } =
    storeToRefs(workspaceStore)

  const currentCollaboration = ref<WorkspaceItem | null>(null)
  const collaborationLoadDone = ref(false)
  const members = ref<MemberItem[]>([])
  const hasMemberOperationPermission = computed(() =>
    hasAction('collaboration.member.manage')
  )
  const loading = ref(false)
  const addLoading = ref(false)
  const pagination = reactive({
    current: 1,
    size: 10
  })
  const heroMetrics = computed(() => [
    { label: '成员总数', value: members.value.length },
    {
      label: '管理员',
      value: members.value.filter(
        (item) =>
          item.roleCode === 'collaboration_admin' ||
          item.role === 'collaboration_admin'
      ).length
    },
    {
      label: '普通成员',
      value: members.value.filter(
        (item) =>
          item.roleCode !== 'collaboration_admin' &&
          item.role !== 'collaboration_admin'
      ).length
    }
  ])
  const heroTitle = computed(() => {
    const workspaceName =
      currentAuthWorkspace.value?.name || currentCollaboration.value?.name || '当前协作空间'
    const collaborationName = currentCollaboration.value?.name || workspaceName
    return `协作空间成员（${collaborationName}） · ${workspaceName}`
  })

  const addForm = reactive({
    user_id: '',
    role_code: 'collaboration_member'
  })

  const pagedMembers = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return members.value.slice(start, start + pagination.size)
  })

  function isAdmin(row: MemberItem): boolean {
    return (
      row.roleCode === 'collaboration_admin' ||
      row.role === 'collaboration_admin'
    )
  }

  function formatRoleLabel(roleCode?: string) {
    return roleCode === 'collaboration_admin' ? '协作空间管理员' : '协作空间成员'
  }

  async function loadCurrentCollaboration() {
    collaborationLoadDone.value = false
    currentCollaboration.value = null
    members.value = []

    if (!hasCollaborations.value) {
      collaborationLoadDone.value = true
      return
    }

    try {
      currentCollaboration.value = await fetchGetMyCollaboration()
      await loadMembers()
    } catch (e: any) {
      if ([400, 404, 3006].includes(e?.response?.status) || [400, 404, 3006].includes(e?.code)) {
        currentCollaboration.value = null
      } else {
        ElMessage.error(e?.message || '获取协作空间信息失败')
      }
    } finally {
      collaborationLoadDone.value = true
    }
  }

  async function loadMembers() {
    if (!currentCollaboration.value) return
    loading.value = true
    try {
      members.value = await fetchGetMyCollaborationMembers()
      pagination.current = 1
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员列表失败')
      members.value = []
      pagination.current = 1
    } finally {
      loading.value = false
    }
  }

  function handleAssignRoles(member: MemberItem) {
    currentMember.value = member
    nextTick(() => {
      roleDialogRef.value?.open()
    })
  }

  function handleCommand(
    command: string,
    member: MemberItem
  ) {
    if (command === 'assign') {
      handleAssignRoles(member)
    } else if (command === 'delete') {
      removeMember(member)
    }
  }

  function removeMember(row: MemberItem) {
    if (isAdmin(row)) {
      ElMessage.warning('协作空间管理员不能被移除')
      return
    }

    ElMessageBox.confirm(`确定将“${row.userName}”移出协作空间吗？`, '移除成员', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchRemoveMyCollaborationMember(row.userId))
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
      await fetchAddMyCollaborationMember({ user_id: uid, role_code: addForm.role_code })
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
    loadCurrentCollaboration()
  })

  watch(
    [currentAuthWorkspaceId, currentAuthWorkspaceType],
    ([workspaceId, workspaceType], [oldWorkspaceId, oldWorkspaceType]) => {
      if (workspaceId === oldWorkspaceId && workspaceType === oldWorkspaceType) return
      loadCurrentCollaboration()
    }
  )
</script>

<style scoped>
  .collaboration-members-empty {
    min-height: 320px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .collaboration-members-loading :deep(.el-icon) {
    animation: rotate 1s linear infinite;
  }

  .collaboration-members-main {
    display: grid;
    gap: 14px;
  }

  .collaboration-members-hero-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .collaboration-members-add {
    padding: 14px;
    border-radius: 18px;
    border: 1px solid var(--art-card-border);
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.94));
    box-shadow: var(--art-shadow-sm);
  }

  .collaboration-members-add__header {
    margin-bottom: 12px;
  }

  .collaboration-members-add__header h3 {
    margin: 0;
    font-size: 15px;
    color: var(--art-text-strong);
    font-weight: 700;
  }

  .collaboration-members-add__header p {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--art-text-muted);
  }

  .collaboration-members-add__form {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0 14px;
  }

  .collaboration-members-table {
    border: 1px solid var(--art-card-border);
    border-radius: 18px;
    padding: 14px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.94));
    box-shadow: var(--art-shadow-sm);
  }

  .collaboration-members-table-summary {
    display: grid;
    gap: 4px;
  }

  .collaboration-members-table-summary strong {
    font-size: 15px;
    color: var(--art-text-strong);
    font-weight: 700;
  }

  .collaboration-members-table-summary span {
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
    .collaboration-members-add__form {
      grid-template-columns: 1fr;
    }
  }
</style>
