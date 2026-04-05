<template>
  <ElDialog v-model="visible" :title="dialogTitle" width="800px" class="backup-dialog" destroy-on-close>
    <div class="backup-list-container">
      <ElAlert
        class="mb-4"
        type="info"
        :closable="false"
        :description="alertDescription"
      />
      <ElTable v-loading="loading" :data="pagedItems" style="width: 100%" border stripe>
        <ElTableColumn prop="name" label="备份名称" width="200">
          <template #default="{ row }">
            <span class="font-medium">{{ row.name }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="space_name" label="作用范围" width="170">
          <template #default="{ row }">
            <ElTag :type="getScopeTagType(row)" effect="light">
              {{ getScopeLabel(row) }}
            </ElTag>
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
            <ArtButtonMore
              :list="operationList"
              @click="(item) => emit('action', String(item.key), row)"
            />
          </template>
        </ElTableColumn>
      </ElTable>
      <WorkspacePagination
        v-if="items.length > 0"
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.size"
        :total="items.length"
        compact
      />

      <div v-if="items.length === 0" class="empty-backup">
        <ElEmpty :description="emptyDescription" />
      </div>
    </div>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, reactive, watch } from 'vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'

  interface MenuBackupItem {
    id: string
    name: string
    description?: string
    scope_type?: string
    space_key?: string
    space_name?: string
    created_at?: string
    created_by?: string
  }

  interface Props {
    modelValue: boolean
    loading?: boolean
    items?: MenuBackupItem[]
    title?: string
    alertDescription?: string
    emptyDescription?: string
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
  const dialogTitle = computed(() => props.title || '管理备份')
  const alertDescription = computed(
    () => props.alertDescription || '列表展示当前空间备份与全空间备份。恢复前请先核对作用范围。'
  )
  const emptyDescription = computed(() => props.emptyDescription || '暂无备份数据')

  const pagination = reactive({
    current: 1,
    size: 10
  })

  const pagedItems = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return props.items.slice(start, start + pagination.size)
  })

  const getScopeLabel = (item: MenuBackupItem) => {
    if (item.space_name) {
      return item.space_name
    }
    if (item.scope_type === 'global') {
      return '全空间备份'
    }
    return '-'
  }

  const getScopeTagType = (item: MenuBackupItem) => {
    return item.scope_type === 'global' ? 'warning' : 'success'
  }

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

  watch(
    () => [props.modelValue, props.items.length],
    () => {
      pagination.current = 1
    }
  )
</script>

<style scoped lang="scss">
  .backup-list-container {
    padding: 12px 0;
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

