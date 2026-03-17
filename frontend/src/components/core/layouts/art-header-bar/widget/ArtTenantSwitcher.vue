<template>
  <div
    v-if="teamList.length"
    class="tenant-switcher"
    :class="{ 'max-md:!hidden': !compact, 'tenant-switcher-compact': compact }"
  >
    <div v-if="compact" class="tenant-label">当前团队</div>
    <ElSelect
      :model-value="currentTenantId"
      class="tenant-select"
      size="default"
      placeholder="选择团队"
      :loading="loading"
      @change="handleChange"
    >
      <ElOption
        v-for="team in teamList"
        :key="team.id"
        :label="buildTeamLabel(team)"
        :value="team.id"
      />
    </ElSelect>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useTenantStore } from '@/store/modules/tenant'
  import { refreshCurrentUserInfoContext, refreshUserMenus } from '@/router'

  defineOptions({ name: 'ArtTenantSwitcher' })
  withDefaults(defineProps<{ compact?: boolean }>(), {
    compact: false
  })

  const router = useRouter()
  const tenantStore = useTenantStore()
  const { currentTenantId, teamList, loading } = storeToRefs(tenantStore)

  const buildTeamLabel = (team: Api.SystemManage.TeamListItem) => {
    const suffix = team.currentRoleCode === 'team_admin' ? '管理员' : '成员'
    return `${team.name} · ${suffix}`
  }

  const handleChange = async (tenantId: string) => {
    if (!tenantId || tenantId === currentTenantId.value) return

    try {
      tenantStore.setCurrentTenantId(tenantId)
      await refreshCurrentUserInfoContext()
      await refreshUserMenus()
      await router.push('/')
      ElMessage.success('已切换团队')
    } catch (error) {
      console.error('[TenantSwitcher] 切换团队失败:', error)
      await tenantStore.loadMyTeams()
      ElMessage.error('切换团队失败')
    }
  }
</script>

<style scoped lang="scss">
  .tenant-switcher {
    min-width: 180px;
    margin-right: 6px;
  }

  .tenant-switcher-compact {
    width: 100%;
    min-width: 0;
    margin-right: 0;
  }

  .tenant-label {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--art-text-gray-600);
  }

  .tenant-select {
    width: 100%;
  }

  :deep(.el-select__wrapper) {
    min-height: 36px;
    border-radius: 10px;
  }

  .tenant-switcher-compact :deep(.el-select__wrapper) {
    min-height: 40px;
    background-color: var(--art-gray-100);
  }
</style>
