<template>
  <div
    v-if="shouldShowSwitcher"
    class="tenant-switcher"
    :class="{ 'max-md:!hidden': !compact, 'tenant-switcher-compact': compact }"
  >
    <div v-if="compact" class="tenant-label">{{ currentContextLabel }}</div>
    <ElSelect
      :model-value="selectedValue"
      class="tenant-select"
      size="default"
      placeholder="选择上下文"
      :loading="loading"
      @change="handleChange"
    >
      <ElOption
        v-if="hasPlatformAccess"
        key="__platform__"
        label="平台空间"
        value="__platform__"
      />
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
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { useTenantStore } from '@/store/modules/tenant'
  import { refreshCurrentUserInfoContext, refreshUserMenus } from '@/router'

  defineOptions({ name: 'ArtTenantSwitcher' })
  withDefaults(defineProps<{ compact?: boolean }>(), {
    compact: false
  })

  const router = useRouter()
  const tenantStore = useTenantStore()
  const { currentContextMode, currentTenantId, teamList, loading, hasPlatformAccess, shouldShowSwitcher } =
    storeToRefs(tenantStore)
  const platformValue = '__platform__'

  const selectedValue = computed(() =>
    currentContextMode.value === 'platform' ? platformValue : currentTenantId.value
  )
  const currentContextLabel = computed(() =>
    currentContextMode.value === 'platform' ? '当前空间' : '当前团队'
  )

  const buildTeamLabel = (team: Api.SystemManage.TeamListItem) => {
    const suffix = team.currentRoleCode === 'team_admin' ? '管理员' : '成员'
    return `${team.name} · ${suffix}`
  }

  const handleChange = async (value: string) => {
    if (!value) return
    if (value === platformValue && currentContextMode.value === 'platform') return
    if (value !== platformValue && currentContextMode.value === 'team' && value === currentTenantId.value) {
      return
    }

    try {
      if (value === platformValue) {
        tenantStore.enterPlatformContext()
      } else {
        tenantStore.enterTeamContext(value)
      }
      await refreshCurrentUserInfoContext()
      await refreshUserMenus()
      await router.push('/')
      ElMessage.success(value === platformValue ? '已切换到平台空间' : '已切换团队')
    } catch (error) {
      console.error('[TenantSwitcher] 切换团队失败:', error)
      await tenantStore.loadMyTeams({
        preferredTenantId: currentContextMode.value === 'team' ? currentTenantId.value : '',
        preferPlatform: currentContextMode.value === 'platform'
      })
      ElMessage.error(value === platformValue ? '切换平台空间失败' : '切换团队失败')
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
