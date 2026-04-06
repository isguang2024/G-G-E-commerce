<template>
  <div class="context-badge" :class="[`context-badge--${variant}`]">
    <span class="context-badge__eyebrow">当前授权工作空间</span>
    <div class="context-badge__main">
      <ElTag :type="tagType" effect="plain" round size="small">
        {{ modeLabel }}
      </ElTag>
      <span class="context-badge__name">{{ scopeName }}</span>
    </div>
    <div class="context-badge__space">
      当前协作空间视图 · {{ collaborationWorkspaceViewName }}
    </div>
    <div v-if="showSpaceLabel" class="context-badge__space"> 菜单空间 · {{ spaceName }} </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useCollaborationWorkspaceStore } from '@/store/modules/collaboration-workspace'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'

  defineOptions({ name: 'AppContextBadge' })

  withDefaults(
    defineProps<{
      variant?: 'soft' | 'solid'
    }>(),
    {
      variant: 'soft'
    }
  )

  const collaborationWorkspaceStore = useCollaborationWorkspaceStore()
  const menuSpaceStore = useMenuSpaceStore()
  const {
    currentContextMode,
    currentCollaborationWorkspace,
    currentAuthWorkspace,
    currentAuthWorkspaceType
  } = storeToRefs(collaborationWorkspaceStore)
  const { currentSpace, shouldShowSpaceBadge, isDefaultSpace } = storeToRefs(menuSpaceStore)

  const modeLabel = computed(() =>
    currentAuthWorkspaceType.value === 'personal' ? '个人空间' : '协作空间'
  )
  const scopeName = computed(() =>
    currentContextMode.value === 'personal'
      ? currentAuthWorkspace.value?.name || '当前个人空间'
      : currentCollaborationWorkspace.value?.name ||
        currentAuthWorkspace.value?.name ||
        '未启用协作空间视图'
  )
  const tagType = computed(() => (currentContextMode.value === 'personal' ? 'success' : 'warning'))
  const collaborationWorkspaceViewName = computed(() =>
    currentAuthWorkspaceType.value === 'collaboration'
      ? currentCollaborationWorkspace.value?.name ||
        currentAuthWorkspace.value?.name ||
        '当前协作空间'
      : '未启用协作空间视图'
  )
  const showSpaceLabel = computed(() => shouldShowSpaceBadge.value && !isDefaultSpace.value)
  const spaceName = computed(
    () => currentSpace.value?.spaceName || currentSpace.value?.spaceKey || '未选择空间'
  )
</script>

<style scoped lang="scss">
  .context-badge {
    display: inline-flex;
    min-width: 0;
    flex-direction: column;
    gap: 5px;
    padding: 10px 14px;
    border: 1px solid color-mix(in srgb, var(--el-border-color) 70%, white);
    border-radius: 16px;
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.9), transparent 58%),
      linear-gradient(135deg, rgb(250 252 255 / 0.96), rgb(242 246 250 / 0.9));
    box-shadow: var(--art-shadow-sm);
  }

  .context-badge--solid {
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.22), transparent 52%),
      linear-gradient(135deg, rgb(15 23 42 / 0.92), rgb(30 41 59 / 0.92));
    border-color: rgb(148 163 184 / 0.25);
    box-shadow: var(--art-shadow-md);
  }

  .context-badge__eyebrow {
    font-size: 10px;
    line-height: 1;
    letter-spacing: 0.08em;
    color: var(--el-text-color-secondary);
    text-transform: uppercase;
  }

  .context-badge--solid .context-badge__eyebrow {
    color: rgb(203 213 225 / 0.9);
  }

  .context-badge__main {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 10px;
  }

  .context-badge__name {
    min-width: 0;
    overflow: hidden;
    font-size: 13px;
    font-weight: 650;
    color: var(--art-text-strong);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .context-badge--solid .context-badge__name {
    color: #f8fafc;
  }

  .context-badge__space {
    overflow: hidden;
    font-size: 12px;
    color: var(--art-text-muted);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .context-badge--solid .context-badge__space {
    color: rgb(226 232 240 / 0.86);
  }
</style>
