<template>
  <div class="workspace-pagination" :class="{ 'is-compact': compact }">
    <ElPagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="pageSizes"
      :layout="layout"
      :small="small"
      background
    />
  </div>
</template>

<script setup lang="ts">
  defineOptions({ name: 'WorkspacePagination' })

  withDefaults(
    defineProps<{
      total: number
      pageSizes?: number[]
      layout?: string
      small?: boolean
      compact?: boolean
    }>(),
    {
      pageSizes: () => [10, 20, 50, 100],
      layout: 'total, sizes, prev, pager, next',
      small: false,
      compact: false
    }
  )

  const currentPage = defineModel<number>('currentPage', {
    required: true
  })

  const pageSize = defineModel<number>('pageSize', {
    required: true
  })
</script>

<style scoped lang="scss">
  .workspace-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
  }

  .workspace-pagination.is-compact {
    margin-top: 10px;
  }

  .workspace-pagination :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
  }

  .workspace-pagination :deep(.btn-prev),
  .workspace-pagination :deep(.btn-next),
  .workspace-pagination :deep(.el-pager li) {
    border-radius: 10px;
  }
</style>
