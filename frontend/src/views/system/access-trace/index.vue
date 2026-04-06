<template>
  <div class="access-trace-page art-full-height">
    <ElCard shadow="never" class="art-table-card">
      <template #header>
        <div class="trace-header">
          <div class="trace-title">访问链路测试</div>
          <div class="trace-subtitle">用户 -> 角色/协作空间 -> 菜单可见 -> 页面可见性</div>
        </div>
      </template>

      <ElForm class="trace-form">
        <ElFormItem label="App">
          <ElSelect
            v-model="selectedAppKey"
            filterable
            clearable
            placeholder="选择 App"
            class="trace-field"
            @change="handleManagedAppChange"
          >
            <ElOption
              v-for="item in appOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="空间(可选)">
          <ElSelect
            v-model="query.spaceKey"
            filterable
            clearable
            placeholder="全部空间"
            class="trace-field"
          >
            <ElOption label="全部空间" value="" />
            <ElOption
              v-for="item in spaceOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="用户">
          <ElSelect
            v-model="query.userId"
            filterable
            clearable
            placeholder="请选择用户"
            class="trace-field"
          >
            <ElOption
              v-for="user in userOptions"
              :key="user.id"
              :label="formatUserLabel(user)"
              :value="user.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="协作空间(可选)">
          <ElSelect
            v-model="query.collaborationWorkspaceId"
            filterable
            clearable
            placeholder="选择协作空间"
            class="trace-field"
          >
            <ElOption
              v-for="team in teamOptions"
              :key="team.id"
              :label="team.name"
              :value="team.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="仅协作空间用户">
          <ElSwitch v-model="onlyTeamUsers" :disabled="!query.collaborationWorkspaceId" />
        </ElFormItem>
        <ElFormItem label="角色筛选">
          <ElSelect v-model="roleCodeFilter" clearable placeholder="全部角色" class="trace-field">
            <ElOption
              v-for="role in displayRoleOptions"
              :key="role.value"
              :label="role.label"
              :value="role.value"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="页面Key(可选)">
          <ElSelect
            v-model="query.pageKey"
            filterable
            clearable
            placeholder="选择页面"
            class="trace-field"
          >
            <ElOption
              v-for="page in pageOptions"
              :key="page.pageKey"
              :label="`${page.pageKey} (${page.name})`"
              :value="page.pageKey"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" :loading="loading" @click="handleQuery">测试链路</ElButton>
        </ElFormItem>
      </ElForm>

      <ElDescriptions v-if="result" :column="2" border class="trace-summary">
        <ElDescriptionsItem label="用户ID">{{ result.userId }}</ElDescriptionsItem>
        <ElDescriptionsItem label="协作空间ID">{{
          result.collaborationWorkspaceId || '-'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="空间">{{ result.spaceKey }}</ElDescriptionsItem>
        <ElDescriptionsItem label="登录态">{{
          result.authenticated ? '已认证' : '未认证'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="超级管理员">{{
          result.superAdmin ? '是' : '否'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="动作权限数">{{ result.actionKeyCount }}</ElDescriptionsItem>
      </ElDescriptions>

      <ElDivider v-if="result">角色链路</ElDivider>
      <ElTable v-if="result" :data="pagedRoles" border>
        <ElTableColumn prop="roleCode" label="角色编码" min-width="160" />
        <ElTableColumn prop="roleName" label="角色名称" min-width="180" />
        <ElTableColumn prop="status" label="状态" width="120" />
      </ElTable>
      <WorkspacePagination
        v-if="result && roleRows.length > 0"
        v-model:current-page="rolePagination.current"
        v-model:page-size="rolePagination.size"
        :total="roleRows.length"
        compact
      />

      <ElDivider v-if="result">页面链路结果</ElDivider>
      <ElTable v-if="result" :data="pagedPages" border>
        <ElTableColumn prop="pageKey" label="页面Key" min-width="180" />
        <ElTableColumn prop="pageName" label="页面名称" min-width="140" />
        <ElTableColumn prop="routePath" label="路由" min-width="160" />
        <ElTableColumn prop="accessMode" label="访问模式" width="120" />
        <ElTableColumn prop="permissionKey" label="权限键" min-width="160" />
        <ElTableColumn label="可见" width="90">
          <template #default="{ row }">
            <ElTag :type="row.visible ? 'success' : 'danger'" effect="plain">
              {{ row.visible ? '是' : '否' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="reason" label="判定原因" min-width="180" />
        <ElTableColumn prop="matchedActionKey" label="命中动作" min-width="180" />
        <ElTableColumn label="链路" min-width="220">
          <template #default="{ row }">
            <div>{{ (row.effectiveChain || []).join(' -> ') || '-' }}</div>
          </template>
        </ElTableColumn>
      </ElTable>
      <WorkspacePagination
        v-if="result && pageRows.length > 0"
        v-model:current-page="pagePagination.current"
        v-model:page-size="pagePagination.size"
        :total="pageRows.length"
        compact
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import { fetchGetCollaborationWorkspaceMembers, fetchGetTeamRoles } from '@/api/team'
  import {
    fetchGetApps,
    fetchGetMenuSpaces,
    fetchGetPageAccessTrace,
    fetchGetPageList,
    fetchGetRoleOptions,
    fetchGetTenantOptions,
    fetchGetUserList
  } from '@/api/system-manage'

  defineOptions({ name: 'SystemAccessTrace' })

  const loading = ref(false)
  const result = ref<Api.SystemManage.PageAccessTraceResult | null>(null)
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const userOptions = ref<Api.SystemManage.UserListItem[]>([])
  const pageOptions = ref<Api.SystemManage.PageItem[]>([])
  const menuSpaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const teamOptions = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
  const roleOptions = ref<
    Array<{ label: string; value: string; source: 'platform' | 'collaboration' }>
  >([])
  const selectedAppKey = ref('')
  const rolePagination = reactive({
    current: 1,
    size: 10
  })
  const pagePagination = reactive({
    current: 1,
    size: 10
  })
  const onlyTeamUsers = ref(false)
  const roleCodeFilter = ref('')
  const displayRoleOptions = computed(() =>
    query.collaborationWorkspaceId
      ? roleOptions.value.filter((item) => item.source === 'collaboration')
      : roleOptions.value.filter((item) => item.source === 'platform')
  )
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )
  const spaceOptions = computed(() =>
    menuSpaces.value.map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )
  const roleRows = computed(() => result.value?.roles || [])
  const pageRows = computed(() => result.value?.pages || [])
  const pagedRoles = computed(() => {
    const start = (rolePagination.current - 1) * rolePagination.size
    return roleRows.value.slice(start, start + rolePagination.size)
  })
  const pagedPages = computed(() => {
    const start = (pagePagination.current - 1) * pagePagination.size
    return pageRows.value.slice(start, start + pagePagination.size)
  })

  const query = reactive<Api.SystemManage.PageAccessTraceParams>({
    userId: '',
    collaborationWorkspaceId: '',
    pageKey: '',
    spaceKey: ''
  })

  async function loadAppOptions() {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }

  function formatUserLabel(user: Api.SystemManage.UserListItem) {
    const userName = `${user.userName || ''}`.trim()
    const nickName = `${user.nickName || ''}`.trim()
    if (userName && nickName && userName !== nickName) {
      return `${userName}（${nickName}）`
    }
    return userName || nickName || user.id
  }

  async function loadUserOptions() {
    if (!targetAppKey.value) {
      userOptions.value = []
      return
    }
    const useCollaborationWorkspaceMembers =
      Boolean(query.collaborationWorkspaceId) &&
      (onlyTeamUsers.value || Boolean(roleCodeFilter.value))
    if (useCollaborationWorkspaceMembers && query.collaborationWorkspaceId) {
      const teamMembers = await fetchGetCollaborationWorkspaceMembers(
        query.collaborationWorkspaceId,
        {
          role_code: roleCodeFilter.value || undefined
        }
      )
      userOptions.value = (teamMembers || []).map((item) => ({
        id: item.userId,
        userName: item.userName,
        nickName: item.nickName,
        userPhone: '',
        userEmail: item.userEmail || '',
        avatar: item.avatar || '',
        status: item.status || 'active',
        roleIDs: [],
        roleNames: [],
        roleDetails: [],
        userRoles: [],
        registerSource: '',
        invitedBy: '',
        invitedByName: '',
        createTime: '',
        updateTime: ''
      }))
    } else {
      const users = await fetchGetUserList({
        current: 1,
        size: 200,
        roleId: roleCodeFilter.value || ''
      })
      userOptions.value = users.records || []
    }

    if (query.userId && !userOptions.value.some((item) => item.id === query.userId)) {
      query.userId = ''
    }
  }

  async function loadRoleOptions() {
    if (!targetAppKey.value) {
      roleOptions.value = []
      return
    }
    if (query.collaborationWorkspaceId) {
      const teamRoles = await fetchGetTeamRoles(query.collaborationWorkspaceId)
      roleOptions.value = (teamRoles || [])
        .filter((role) => {
          const code = `${role.roleCode || ''}`.trim()
          if (!code || code === 'admin') return false
          if (code === 'collaboration_workspace_admin' || code === 'collaboration_workspace_member')
            return true
          return Boolean(role.collaborationWorkspaceId)
        })
        .map((role) => ({
          label: role.roleName || role.roleCode,
          value: role.roleCode,
          source: 'collaboration' as const
        }))
      return
    }

    const roleRes = await fetchGetRoleOptions()
    roleOptions.value = (roleRes.records || []).map((role) => ({
      label: role.roleName || role.roleCode,
      value: role.roleId,
      source: 'platform' as const
    }))
  }

  async function loadOptions() {
    if (!targetAppKey.value) {
      result.value = null
      userOptions.value = []
      pageOptions.value = []
      menuSpaces.value = []
      teamOptions.value = []
      roleOptions.value = []
      return
    }
    const [pages, teams, spaces] = await Promise.all([
      fetchGetPageList({ current: 1, size: 500, appKey: targetAppKey.value }),
      fetchGetTenantOptions({ current: 1, size: 200 }),
      fetchGetMenuSpaces(targetAppKey.value)
    ])
    pageOptions.value = pages.records || []
    teamOptions.value = teams.records || []
    menuSpaces.value = spaces.records || []
    await loadRoleOptions()
    await loadUserOptions()
  }

  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
    query.pageKey = ''
    query.spaceKey = ''
    query.userId = ''
    query.collaborationWorkspaceId = ''
    roleCodeFilter.value = ''
    onlyTeamUsers.value = false
    result.value = null
  }

  async function handleQuery() {
    if (!targetAppKey.value) {
      ElMessage.warning('请先选择 App')
      return
    }
    if (!query.userId) {
      ElMessage.warning('请先选择用户')
      return
    }
    if (query.collaborationWorkspaceId && query.collaborationWorkspaceId === query.userId) {
      ElMessage.warning('协作空间 ID 不能与用户 ID 相同，请重新选择协作空间')
      return
    }
    loading.value = true
    try {
      result.value = await fetchGetPageAccessTrace({
        ...query,
        appKey: targetAppKey.value
      })
      rolePagination.current = 1
      pagePagination.current = 1
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
    loadAppOptions().catch(() => {
      appList.value = []
    })
    loadOptions().catch(() => {
      ElMessage.error('初始化测试数据失败')
    })
  })

  watch(
    () => targetAppKey.value,
    async () => {
      selectedAppKey.value = targetAppKey.value || ''
      result.value = null
      await loadOptions()
    }
  )

  watch(
    () => query.collaborationWorkspaceId,
    async (collaborationWorkspaceId, oldTenantId) => {
      if (collaborationWorkspaceId !== oldTenantId) {
        roleCodeFilter.value = ''
      }
      if (!collaborationWorkspaceId && onlyTeamUsers.value) {
        onlyTeamUsers.value = false
      }
      await loadRoleOptions()
      await loadUserOptions()
    }
  )

  watch(
    () => [onlyTeamUsers.value, roleCodeFilter.value],
    async () => {
      await loadUserOptions()
    }
  )

  watch(
    () => rolePagination.size,
    () => {
      rolePagination.current = 1
    }
  )

  watch(
    () => pagePagination.size,
    () => {
      pagePagination.current = 1
    }
  )
</script>

<style scoped>
  .trace-header {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .trace-title {
    font-size: 16px;
    font-weight: 600;
  }
  .trace-subtitle {
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
  .trace-form {
    margin-bottom: 12px;
    display: grid;
    gap: 8px 12px;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    align-items: end;
  }
  .trace-field {
    width: 100%;
  }
  .trace-summary {
    margin-bottom: 12px;
  }

  .trace-form :deep(.el-form-item) {
    margin-bottom: 0;
  }

  @media (max-width: 1440px) {
    .trace-form {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }

  @media (max-width: 1024px) {
    .trace-form {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 768px) {
    .trace-form {
      grid-template-columns: 1fr;
    }
  }
</style>
