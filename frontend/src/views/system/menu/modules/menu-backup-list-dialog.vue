<template>
  <ElDialog v-model="visible" title="管理备份" width="800px" class="backup-dialog" destroy-on-close>
    <div class="backup-list-container">
      <ElTable v-loading="loading" :data="items" style="width: 100%" border stripe>
        <ElTableColumn prop="name" label="备份名称" width="200">
          <template #default="{ row }">
            <span class="font-medium">{{ row.name }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="备份描述">
          <template #default="{ row }">
            <span class="text-gray-600">{{ row.description || '-' }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="created_at" label="创建时间" width="200" />
        <ElTableColumn prop="created_by" label="创建人" width="150">
          <template #default="{ row }">
            <span class="text-gray-600">{{ row.created_by || '系统' }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="72" fixed="right" align="center">
          <template #default="{ row }">
            <ArtButtonMore :list="operationList" @click="(item) => emit('action', String(item.key), row)" />
          </template>
        </ElTableColumn>
      </ElTable>

      <div v-if="items.length === 0" class="empty-backup">
        <ElEmpty description="暂无备份数据" />
      </div>
    </div>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'

  interface MenuBackupItem {
    id: string
    name: string
    description?: string
    created_at?: string
    created_by?: string
  }

  interface Props {
    modelValue: boolean
    loading?: boolean
    items?: MenuBackupItem[]
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'action', action: string, row: MenuBackupItem): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    loading: false,
    items: () => []
  })

  const emit = defineEmits<Emits>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const operationList: ButtonMoreItem[] = [
    { key: 'restore', label: '恢复备份', icon: 'ri:history-line', auth: 'system.menu.backup' },
    {
      key: 'delete',
      label: '删除备份',
      icon: 'ri:delete-bin-4-line',
      color: '#f56c6c',
      auth: 'system.menu.backup'
    }
  ]
</script>

<style scoped lang="scss">
  .backup-list-container {
    padding: 10px 0;
  }

  .empty-backup {
    padding: 40px 0;
    text-align: center;
  }

  :deep(.el-table) {
    .el-table__row {
      transition: all 0.3s ease;

      &:hover {
        background-color: #f5f7fa !important;
      }
    }

    .el-table__header-wrapper th {
      background-color: #fafafa;
      font-weight: 600;
    }
  }
</style>
