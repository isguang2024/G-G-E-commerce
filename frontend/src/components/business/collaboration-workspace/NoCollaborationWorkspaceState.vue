<template>
  <ElCard shadow="never" class="empty-card">
    <ElEmpty :description="title">
      <template #description>
        <p>{{ description }}</p>
        <p v-if="tip" class="text-gray-500 text-sm mt-2">{{ tip }}</p>
      </template>
    </ElEmpty>
  </ElCard>
</template>

<script setup lang="ts">
  withDefaults(
    defineProps<{
      title?: string
      description?: string
      tip?: string
    }>(),
    {
      title: '您当前还未加入协作空间',
      description: '请先加入协作空间，或联系管理员为您创建协作空间后再进行协作空间管理。',
      tip: '加入协作空间后，您可以在这里查看成员并按权限执行操作。'
    }
  )
</script>

<style scoped>
  .empty-card {
    min-height: 320px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
</style>
