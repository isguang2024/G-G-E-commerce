<template>
  <ElDrawer
    v-model="drawerVisible"
    :title="`团队人员：${teamName || '-'}`"
    size="800px"
    destroy-on-close
    @open="onOpen"
  >
    <!-- 搜索和筛选 -->
    <div class="search-container mb-4">
      <ElRow :gutter="12">
        <ElCol :span="6">
          <ElInput
            v-model="searchForm.userId"
            placeholder="用户ID"
            clearable
            @keyup.enter="handleSearch"
          />
        </ElCol>
        <ElCol :span="6">
          <ElInput
            v-model="searchForm.userName"
            placeholder="用户名"
            clearable
            @keyup.enter="handleSearch"
          />
        </ElCol>
        <ElCol :span="8">
          <ElSelect
            v-model="searchForm.role"
            placeholder="团队身份"
            clearable
            filterable
          >
            <ElOption
              v-for="role in teamRoles"
              :key="role"
              :label="role"
              :value="role"
            />
          </ElSelect>
        </ElCol>
        <ElCol :span="4">
          <ElButton type="primary" @click="handleSearch">搜索</ElButton>
        </ElCol>
      </ElRow>
    </div>
    
    <!-- 成员列表 -->
    <ElCard shadow="never">
      <template #header>
        <span>成员列表（{{ filteredMembers.length }}）</span>
      </template>
      <ElTable v-loading="loading" :data="filteredMembers" stripe>
        <ElTableColumn prop="userName" label="用户名" min-width="100" />
        <ElTableColumn prop="nickName" label="昵称" width="100" />
        <ElTableColumn prop="userEmail" label="邮箱" min-width="140" show-overflow-tooltip />
        <ElTableColumn label="团队身份" min-width="200">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-1">
              <ElTag 
                v-for="(role, index) in (row.roles || [row.role])" 
                :key="index"
                :type="role === '团队管理员' ? 'success' : 'info'"
                size="small"
              >
                {{ role }}
              </ElTag>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="joinedAt" label="加入时间" width="160" />
      </ElTable>
    </ElCard>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { fetchGetTeamMembers, fetchGetMyTeamRoles } from '@/api/team'
  import { ElMessage } from 'element-plus'
  import { ref, computed, reactive } from 'vue'

  interface Props {
    visible: boolean
    teamId: string
    teamName?: string
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'refresh'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const drawerVisible = computed({
    get: () => props.visible,
    set: (v) => emit('update:visible', v)
  })

  const teamName = computed(() => props.teamName)
  const members = ref<Api.SystemManage.TeamMemberItem[]>([])
  const loading = ref(false)
  const teamRoles = ref<string[]>([])
  
  const searchForm = reactive({
    userId: '',
    userName: '',
    role: ''
  })

  // 搜索参数
  const searchParams = computed(() => ({
    user_id: searchForm.userId || undefined,
    user_name: searchForm.userName || undefined,
    role: searchForm.role || undefined
  }))

  // 判断是否为管理员（团队管理员角色编码为 team_admin）
  function isAdmin(row: Api.SystemManage.TeamMemberItem): boolean {
    const roleCodes = (row as any).roleCodes || []
    return roleCodes.includes('team_admin')
  }

  // 获取团队角色
  async function loadTeamRoles() {
    try {
      const res = await fetchGetMyTeamRoles()
      const roles = res?.records || []
      teamRoles.value = roles.map((r: any) => r.roleName)
    } catch (e) {
      console.error('获取团队角色失败:', e)
    }
  }

  // 过滤后的成员列表（用于显示）
  const filteredMembers = computed(() => members.value)

  async function loadMembers() {
    if (!props.teamId) return
    loading.value = true
    try {
      const res = await fetchGetTeamMembers(props.teamId, searchParams.value)
      members.value = res?.records ?? []
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员列表失败')
      members.value = []
    } finally {
      loading.value = false
    }
  }

  function handleSearch() {
    loadMembers()
  }

  function onOpen() {
    // 重置搜索表单
    searchForm.userId = ''
    searchForm.userName = ''
    searchForm.role = ''
    loadTeamRoles()
    loadMembers()
  }
</script>
