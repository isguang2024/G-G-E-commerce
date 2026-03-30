<template>
  <div class="context-badge" :class="[`context-badge--${variant}`]">
    <span class="context-badge__eyebrow">当前作用域</span>
    <div class="context-badge__main">
      <ElTag :type="tagType" effect="plain" round size="small">
        {{ modeLabel }}
      </ElTag>
      <span class="context-badge__name">{{ scopeName }}</span>
    </div>
    <div v-if="showSpaceLabel" class="context-badge__space">
      菜单空间 · {{ spaceName }}
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useTenantStore } from '@/store/modules/tenant'
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

  const tenantStore = useTenantStore()
  const menuSpaceStore = useMenuSpaceStore()
  const { currentContextMode, currentTeam } = storeToRefs(tenantStore)
  const { currentSpace, shouldShowSpaceBadge, isDefaultSpace } = storeToRefs(menuSpaceStore)

  const modeLabel = computed(() => (currentContextMode.value === 'platform' ? '平台' : '团队'))
  const scopeName = computed(() =>
    currentContextMode.value === 'platform' ? '平台管理空间' : currentTeam.value?.name || '未选择团队'
  )
  const tagType = computed(() => (currentContextMode.value === 'platform' ? 'success' : 'warning'))
  const showSpaceLabel = computed(() => shouldShowSpaceBadge.value && !isDefaultSpace.value)
  const spaceName = computed(() => currentSpace.value?.spaceName || currentSpace.value?.spaceKey || '默认菜单空间')
</script>

<style scoped lang="scss">
  .context-badge {
    display: inline-flex;
    min-width: 0;
    flex-direction: column;
    gap: 4px;
    padding: 8px 12px;
    border: 1px solid color-mix(in srgb, var(--el-border-color) 72%, white);
    border-radius: 14px;
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.85), transparent 58%),
      linear-gradient(135deg, rgb(248 250 252 / 0.96), rgb(241 245 249 / 0.9));
    box-shadow: 0 10px 24px rgb(15 23 42 / 0.06);
  }

  .context-badge--solid {
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.22), transparent 52%),
      linear-gradient(135deg, rgb(15 23 42 / 0.92), rgb(30 41 59 / 0.92));
    border-color: rgb(148 163 184 / 0.25);
    box-shadow: 0 16px 36px rgb(15 23 42 / 0.18);
  }

  .context-badge__eyebrow {
    font-size: 11px;
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
    gap: 8px;
  }

  .context-badge__name {
    min-width: 0;
    overflow: hidden;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .context-badge--solid .context-badge__name {
    color: #f8fafc;
  }

  .context-badge__space {
    overflow: hidden;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .context-badge--solid .context-badge__space {
    color: rgb(226 232 240 / 0.86);
  }
</style>
