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
      placeholder="切换团队"
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
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { useTenantStore } from '@/store/modules/tenant'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { refreshCurrentUserInfoContext, refreshUserMenus } from '@/router'

  defineOptions({ name: 'ArtTenantSwitcher' })
  withDefaults(defineProps<{ compact?: boolean }>(), {
    compact: false
  })

  const router = useRouter()
  const tenantStore = useTenantStore()
  const menuSpaceStore = useMenuSpaceStore()
  const { currentTenantId, teamList, loading, shouldShowSwitcher, currentTeam } = storeToRefs(tenantStore)

  const selectedValue = computed(() => currentTenantId.value)
  const currentContextLabel = computed(() => currentTeam.value?.name || '当前团队')

  const buildTeamLabel = (team: Api.SystemManage.TeamListItem) => {
    const suffix = team.currentRoleCode === 'team_admin' ? '管理员' : '成员'
    return `${team.name} · ${suffix}`
  }

  const handleChange = async (value: string) => {
    if (!value) return
    if (value === currentTenantId.value) return

    try {
      tenantStore.enterTeamContext(value)
      await refreshCurrentUserInfoContext()
      await refreshUserMenus()
      const landingPath = menuSpaceStore.resolveSpaceLandingPath()
      const resolved = router.resolve(landingPath)
      const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
        resolved.href,
        `${resolved.meta?.spaceKey || ''}`.trim() || undefined
      )
      if (nextTarget.mode === 'router') {
        await router.push(landingPath)
      } else {
        window.location.assign(nextTarget.target)
        return
      }
      ElMessage.success('已切换团队')
    } catch (error) {
      console.error('[TenantSwitcher] 切换团队失败:', error)
      await tenantStore.loadMyTeams({
        preferredTenantId: currentTenantId.value || ''
      })
      ElMessage.error('切换团队失败')
    }
  }
</script>

<style scoped lang="scss">
  .tenant-switcher {
    min-width: 210px;
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
    min-height: 38px;
    border-radius: 12px;
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.92), transparent 55%),
      linear-gradient(135deg, rgb(248 250 252 / 0.98), rgb(241 245 249 / 0.94));
    box-shadow: inset 0 0 0 1px rgb(226 232 240 / 0.95);
  }

  .tenant-switcher-compact :deep(.el-select__wrapper) {
    min-height: 40px;
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.96), transparent 58%),
      linear-gradient(135deg, rgb(248 250 252 / 0.98), rgb(241 245 249 / 0.94));
  }
</style>
