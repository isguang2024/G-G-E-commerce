<template>
  <ElDrawer
    v-model="drawerVisible"
    :title="`协作空间人员：${props.collaborationWorkspaceName || '-'}`"
    size="800px"
    destroy-on-close
    @open="onOpen"
    class="config-drawer"
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
          <ElSelect v-model="searchForm.role" placeholder="协作空间身份" clearable filterable>
            <ElOption
              v-for="role in collaborationWorkspaceRoles"
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
      <ElTable v-loading="loading" :data="pagedMembers" stripe>
        <ElTableColumn prop="userName" label="用户名" min-width="100" />
        <ElTableColumn prop="nickName" label="昵称" width="100" />
        <ElTableColumn prop="userEmail" label="邮箱" min-width="140" show-overflow-tooltip />
        <ElTableColumn label="协作空间身份" min-width="200">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-1">
              <ElTag
                v-for="(role, index) in row.roles || [row.role]"
                :key="index"
                :type="role === '协作空间管理员' ? 'success' : 'info'"
                size="small"
              >
                {{ role }}
              </ElTag>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="joinedAt" label="加入时间" width="160" />
      </ElTable>
      <WorkspacePagination
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.size"
        :total="filteredMembers.length"
        compact
      />
    </ElCard>
  </ElDrawer>
</template>

<script setup lang="ts">
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import {
    fetchGetCollaborationWorkspaceMembers,
    fetchGetMyCollaborationWorkspaceRoles
  } from '@/api/collaboration-workspace'
  import { ElMessage } from 'element-plus'
  import { ref, computed, reactive } from 'vue'

  interface Props {
    visible: boolean
    collaborationWorkspaceId: string
    collaborationWorkspaceName?: string
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

  const members = ref<Api.SystemManage.CollaborationWorkspaceMemberItem[]>([])
  const loading = ref(false)
  const collaborationWorkspaceRoles = ref<string[]>([])
  const pagination = reactive({
    current: 1,
    size: 10
  })

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

  // 获取协作空间角色
  async function loadCollaborationWorkspaceRoles() {
    try {
      const res = await fetchGetMyCollaborationWorkspaceRoles()
      collaborationWorkspaceRoles.value = (res || [])
        .map((r: any) => r.roleCode || r.roleName)
        .filter(Boolean)
    } catch (e) {
      console.error('获取协作空间角色失败:', e)
    }
  }

  // 过滤后的成员列表（用于显示）
  const filteredMembers = computed(() => members.value)

  const pagedMembers = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return filteredMembers.value.slice(start, start + pagination.size)
  })

  async function loadMembers() {
    if (!props.collaborationWorkspaceId) return
    loading.value = true
    try {
      const res = await fetchGetCollaborationWorkspaceMembers(
        props.collaborationWorkspaceId,
        searchParams.value
      )
      members.value = res ?? []
      pagination.current = 1
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员列表失败')
      members.value = []
      pagination.current = 1
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
    loadCollaborationWorkspaceRoles()
    loadMembers()
  }
</script>
