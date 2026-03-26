<template>
  <div class="summary-tags">
    <ElTag
      v-for="item in items"
      :key="item.label"
      :type="item.type"
      effect="plain"
      round
    >
      {{ item.label }} {{ item.value }}
    </ElTag>
  </div>
</template>

<script setup lang="ts">
interface SummaryItem {
  label: string
  value: string | number
  type?: 'success' | 'warning' | 'info' | 'primary' | 'danger'
}

defineProps<{
  items: SummaryItem[]
}>()
</script>

<style scoped lang="scss">
.summary-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
