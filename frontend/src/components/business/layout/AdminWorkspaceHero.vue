<template>
  <section class="workspace-hero">
    <div class="workspace-hero__main">
      <div class="workspace-hero__eyebrow">
        <slot name="eyebrow" />
      </div>
      <div class="workspace-hero__heading">
        <h1 class="workspace-hero__title">{{ title }}</h1>
        <p v-if="description" class="workspace-hero__description">{{ description }}</p>
      </div>
      <div v-if="metrics.length" class="workspace-hero__metrics">
        <div v-for="item in metrics" :key="item.label" class="workspace-metric">
          <span class="workspace-metric__label">{{ item.label }}</span>
          <strong class="workspace-metric__value">{{ item.value }}</strong>
        </div>
      </div>
    </div>
    <div class="workspace-hero__side">
      <slot />
    </div>
  </section>
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
    display: flex;
    justify-content: space-between;
    gap: 20px;
    padding: 18px 20px;
    margin-bottom: 12px;
    border: 1px solid color-mix(in srgb, var(--el-border-color) 76%, white);
    border-radius: 22px;
    background:
      radial-gradient(circle at top right, rgb(217 249 157 / 0.26), transparent 34%),
      radial-gradient(circle at top left, rgb(191 219 254 / 0.34), transparent 40%),
      linear-gradient(135deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.96));
  }

  .workspace-hero__main {
    display: flex;
    min-width: 0;
    flex: 1;
    flex-direction: column;
    gap: 12px;
  }

  .workspace-hero__eyebrow:empty {
    display: none;
  }

  .workspace-hero__heading {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 6px;
  }

  .workspace-hero__title {
    margin: 0;
    font-size: 26px;
    font-weight: 700;
    line-height: 1.04;
    letter-spacing: -0.03em;
    color: #0f172a;
  }

  .workspace-hero__description {
    margin: 0;
    max-width: 720px;
    font-size: 13px;
    line-height: 1.6;
    color: #475569;
  }

  .workspace-hero__metrics {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .workspace-metric {
    display: inline-flex;
    min-width: 110px;
    flex-direction: column;
    gap: 4px;
    padding: 10px 12px;
    border-radius: 14px;
    background: rgb(255 255 255 / 0.78);
    box-shadow: inset 0 0 0 1px rgb(226 232 240 / 0.9);
  }

  .workspace-metric__label {
    font-size: 11px;
    letter-spacing: 0.04em;
    color: #64748b;
    text-transform: uppercase;
  }

  .workspace-metric__value {
    font-size: 20px;
    line-height: 1;
    color: #0f172a;
  }

  .workspace-hero__side {
    display: flex;
    flex-shrink: 0;
    align-items: flex-start;
    justify-content: flex-end;
    gap: 12px;
  }

  @media (max-width: 960px) {
    .workspace-hero {
      flex-direction: column;
    }

    .workspace-hero__side {
      width: 100%;
      justify-content: flex-start;
      flex-wrap: wrap;
    }
  }

  @media (max-width: 640px) {
    .workspace-hero {
      padding: 16px;
      border-radius: 18px;
    }

    .workspace-hero__title {
      font-size: 22px;
    }
  }
</style>
