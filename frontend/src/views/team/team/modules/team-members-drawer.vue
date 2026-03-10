<template>
  <ElDrawer
    v-model="drawerVisible"
    :title="`团队人员：${teamName || '-'}`"
    size="520px"
    destroy-on-close
    @open="onOpen"
  >
    <!-- 成员列表 -->
    <ElCard shadow="never">
      <template #header>
        <span>成员列表（{{ members.length }}）</span>
      </template>
      <ElTable v-loading="loading" :data="members" stripe>
        <ElTableColumn prop="userName" label="用户名" min-width="100" />
        <ElTableColumn prop="nickName" label="昵称" width="100" />
        <ElTableColumn prop="userEmail" label="邮箱" min-width="140" show-overflow-tooltip />
        <ElTableColumn prop="role" label="角色" width="120">
          <template #default="{ row }">
            <span v-if="row.role === 'team_admin'" class="text-gray-500">团队管理员</span>
            <span v-else class="text-gray-600">团队成员</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="joinedAt" label="加入时间" width="160" />
      </ElTable>
    </ElCard>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { fetchGetTeamMembers } from '@/api/team'
  import { ElMessage } from 'element-plus'

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

  async function loadMembers() {
    if (!props.teamId) return
    loading.value = true
    try {
      const res = await fetchGetTeamMembers(props.teamId)
      members.value = res?.records ?? []
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员列表失败')
      members.value = []
    } finally {
      loading.value = false
    }
  }

  function onOpen() {
    loadMembers()
  }
</script>
