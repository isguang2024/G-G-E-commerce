<template>
  <div
    v-if="shouldShowSwitcher"
    class="collaboration-workspace-switcher"
    :class="{ 'max-md:!hidden': !compact, 'collaboration-workspace-switcher-compact': compact }"
  >
    <div v-if="compact" class="collaboration-workspace-label">{{ currentContextLabel }}</div>
    <ElSelect
      :model-value="selectedValue"
      class="collaboration-workspace-select"
      size="default"
      placeholder="切换工作空间"
      :loading="loading"
      @change="handleChange"
    >
      <ElOptionGroup v-if="personalWorkspace" label="个人空间">
        <ElOption
          :key="personalWorkspace.id"
          :label="buildWorkspaceLabel(personalWorkspace)"
          :value="personalWorkspace.id"
        />
      </ElOptionGroup>
      <ElOptionGroup v-if="collaborationWorkspaces.length" label="协作空间">
        <ElOption
          v-for="workspace in collaborationWorkspaces"
          :key="workspace.id"
          :label="buildWorkspaceLabel(workspace)"
          :value="workspace.id"
        />
      </ElOptionGroup>
    </ElSelect>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { useMenuStore } from '@/store/modules/menu'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { useCollaborationWorkspaceStore } from '@/store/modules/collaboration-workspace'
  import { useWorkspaceStore } from '@/store/modules/workspace'
  import { refreshCurrentUserInfoContext, refreshUserMenus } from '@/router'
  import { findRegisteredRouteByPath } from '@/utils/router'

  defineOptions({ name: 'ArtCollaborationWorkspaceSwitcher' })
  withDefaults(defineProps<{ compact?: boolean }>(), {
    compact: false
  })

  const router = useRouter()
  const menuStore = useMenuStore()
  const collaborationWorkspaceStore = useCollaborationWorkspaceStore()
  const workspaceStore = useWorkspaceStore()
  const menuSpaceStore = useMenuSpaceStore()

  const { collaborationWorkspaceList } = storeToRefs(collaborationWorkspaceStore)
  const {
    workspaceList,
    personalWorkspace,
    collaborationWorkspaces,
    currentAuthWorkspace,
    currentAuthWorkspaceId,
    loading
  } = storeToRefs(workspaceStore)

  const shouldShowSwitcher = computed(() => workspaceList.value.length > 0)
  const selectedValue = computed(() => currentAuthWorkspaceId.value)
  const currentContextLabel = computed(() => {
    if (!currentAuthWorkspace.value) return '当前工作空间'
    const typeLabel =
      currentAuthWorkspace.value.workspaceType === 'collaboration' ? '协作空间' : '个人空间'
    return `${currentAuthWorkspace.value.name} · ${typeLabel}`
  })

  const buildWorkspaceLabel = (workspace: Api.SystemManage.WorkspaceItem) => {
    if (workspace.workspaceType === 'personal') {
      return `${workspace.name} · 个人空间`
    }
    const matchedCollaborationWorkspace = collaborationWorkspaceList.value.find(
      (item) => item.workspaceId === workspace.id
    )
    const roleLabel =
      matchedCollaborationWorkspace?.currentRoleCode === 'collaboration_workspace_admin'
        ? '管理员视图'
        : '成员视图'
    return `${workspace.name} · 协作空间 · ${roleLabel}`
  }

  const handleChange = async (value: string) => {
    const workspaceId = `${value || ''}`.trim()
    if (!workspaceId || workspaceId === currentAuthWorkspaceId.value) return

    try {
      await workspaceStore.switchWorkspace(workspaceId)
      await refreshCurrentUserInfoContext()
      await refreshUserMenus()

      const landingPath = menuStore.getHomePath() || '/'
      const resolvedRoute = findRegisteredRouteByPath(router, landingPath)
      const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
        landingPath,
        `${resolvedRoute?.meta?.spaceKey || ''}`.trim() || undefined
      )
      if (nextTarget.mode === 'router') {
        await router.push(nextTarget.target)
      } else {
        window.location.assign(nextTarget.target)
        return
      }

      const workspaceTypeLabel =
        currentAuthWorkspace.value?.workspaceType === 'collaboration' ? '协作空间' : '个人空间'
      ElMessage.success(`已切换到${workspaceTypeLabel}`)
    } catch (error) {
      console.error('[WorkspaceSwitcher] 切换工作空间失败:', error)
      await collaborationWorkspaceStore.loadMyCollaborationWorkspaces({
        preferredWorkspaceId: currentAuthWorkspaceId.value
      })
      ElMessage.error('切换工作空间失败')
    }
  }
</script>

<style scoped lang="scss">
  .collaboration-workspace-switcher {
    min-width: 210px;
    margin-right: 6px;
  }

  .collaboration-workspace-switcher-compact {
    width: 100%;
    min-width: 0;
    margin-right: 0;
  }

  .collaboration-workspace-label {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--art-text-gray-600);
  }

  .collaboration-workspace-select {
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

  .collaboration-workspace-switcher-compact :deep(.el-select__wrapper) {
    min-height: 40px;
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.96), transparent 58%),
      linear-gradient(135deg, rgb(248 250 252 / 0.98), rgb(241 245 249 / 0.94));
  }
</style>
