<template>
  <ElCard class="workspace-hero art-card-xs" shadow="never">
    <div class="workspace-hero__body">
      <div class="workspace-hero__header">
        <div class="workspace-hero__main">
          <div class="workspace-hero__eyebrow">
            <slot name="eyebrow" />
          </div>
          <div class="workspace-hero__heading">
            <h1 class="workspace-hero__title">{{ title }}</h1>
            <p v-if="description" class="workspace-hero__description">{{ description }}</p>
          </div>
        </div>
        <div v-if="metrics.length" class="workspace-hero__metrics">
          <div v-for="item in metrics" :key="item.label" class="workspace-metric">
            <span class="workspace-metric__label">{{ item.label }}</span>
            <strong class="workspace-metric__value">{{ item.value }}</strong>
          </div>
        </div>
      </div>
      <div v-if="$slots.default" class="workspace-hero__divider-wrap">
        <div class="workspace-hero__divider" />
        <div class="workspace-hero__actions">
          <slot />
        </div>
      </div>
    </div>
  </ElCard>
</template>

<script setup lang="ts">
  defineOptions({ name: 'AdminWorkspaceHero' })

  withDefaults(
    defineProps<{
      title: string
      description?: string
      metrics?: Array<{
        label: string
        value: string | number
      }>
    }>(),
    {
      metrics: () => []
    }
  )
</script>

<style scoped lang="scss">
  .workspace-hero {
    margin-bottom: 0;
  }

  .workspace-hero :deep(.el-card__body) {
    padding: 20px 22px;
  }

  .workspace-hero__body {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .workspace-hero__header {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    align-items: flex-start;
    gap: 18px;
  }

  .workspace-hero__main {
    display: flex;
    min-width: 0;
    flex: 1 1 320px;
    flex-direction: column;
    gap: 12px;
    max-width: min(100%, 720px);
  }

  .workspace-hero__eyebrow:empty {
    display: none;
  }

  .workspace-hero__eyebrow :deep(.el-tag) {
    border-radius: 9999px;
  }

  .workspace-hero__heading {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 4px;
  }

  .workspace-hero__title {
    margin: 0;
    font-size: 24px;
    font-weight: 720;
    line-height: 1.1;
    letter-spacing: -0.03em;
    color: var(--art-text-strong);
    word-break: keep-all;
    overflow-wrap: break-word;
  }

  .workspace-hero__description {
    margin: 0;
    max-width: 100%;
    font-size: 12px;
    line-height: 1.5;
    color: var(--art-text-muted);
    word-break: break-word;
  }

  .workspace-hero__divider-wrap {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .workspace-hero__divider {
    width: 100%;
    height: 1px;
    background: linear-gradient(90deg, rgb(226 232 240 / 0.95), rgb(226 232 240 / 0.55), rgb(226 232 240 / 0.95));
  }

  .workspace-hero__metrics {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 10px;
    align-self: flex-start;
    flex: 1 1 420px;
    min-width: min(100%, 420px);
  }

  .workspace-metric {
    display: inline-flex;
    flex-direction: column;
    gap: 4px;
    flex: 1 1 112px;
    max-width: 148px;
    min-width: 112px;
    min-height: 46px;
  }

  .workspace-metric__label {
    font-size: 12px;
    color: var(--art-text-soft);
  }

  .workspace-metric__value {
    font-size: 20px;
    line-height: 1;
    color: var(--art-text-strong);
  }

  .workspace-hero__actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-start;
    gap: 10px;
    padding-top: 2px;
  }

  @media (max-width: 960px) {
    .workspace-hero__header {
      flex-direction: column;
    }

    .workspace-hero__main {
      max-width: none;
    }

    .workspace-hero__metrics {
      justify-content: flex-start;
      align-self: stretch;
      min-width: 0;
    }

    .workspace-hero__actions {
      width: 100%;
      justify-content: flex-start;
    }
  }

  @media (max-width: 640px) {
    .workspace-hero :deep(.el-card__body) {
      padding: 16px;
    }

    .workspace-hero__title {
      font-size: 20px;
    }

    .workspace-hero__metrics {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      justify-content: stretch;
      width: 100%;
    }

    .workspace-metric {
      flex-basis: auto;
      min-width: 0;
      max-width: none;
    }
  }
</style>
