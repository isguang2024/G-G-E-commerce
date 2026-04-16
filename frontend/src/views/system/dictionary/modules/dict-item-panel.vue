<template>
  <ElCard class="dict-item-panel" shadow="never">
    <template #header>
      <div class="panel-header">
        <div class="panel-header-left">
          <span class="panel-title">字典项</span>
          <ElTag size="small" type="info">{{ dictType.name }} ({{ dictType.code }})</ElTag>
        </div>
        <div class="panel-header-right">
          <ElButton size="small" @click="showItemDialog('add')">新增</ElButton>
        </div>
      </div>
    </template>

    <div class="dict-item-table-wrap">
      <ElTable
        v-loading="loading"
        :data="itemList"
        border
        max-height="100%"
        style="width: 100%"
        row-key="id"
      >
        <ElTableColumn prop="sort_order" label="排序" width="70" align="center" />
        <ElTableColumn prop="label" label="标签" min-width="120" />
        <ElTableColumn prop="value" label="值" min-width="120" />
        <ElTableColumn prop="description" label="备注" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="is_default" label="默认" width="70" align="center">
          <template #default="{ row }">
            <ElTag v-if="row.is_default" size="small" type="success">是</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="is_builtin" label="内置" width="70" align="center">
          <template #default="{ row }">
            <ElTag v-if="row.is_builtin" size="small" type="info" effect="plain">是</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="row.status === 'normal' ? 'success' : 'info'" size="small">
              {{ row.status === 'normal' ? '正常' : '停用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="220" align="center" fixed="right">
          <template #default="{ row, $index }">
            <ElButton text size="small" @click="showItemDialog('edit', row, $index)">编辑</ElButton>
            <ElButton
              text
              size="small"
              :type="row.status === 'normal' ? 'warning' : 'success'"
              @click="handleToggleStatus(row, $index)"
            >
              {{ row.status === 'normal' ? '停用' : '启用' }}
            </ElButton>
            <ElButton
              v-if="!row.is_builtin"
              text
              size="small"
              type="danger"
              @click="handleDeleteItem(row, $index)"
            >
              删除
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </div>

    <!-- Item Dialog -->
    <DictItemDialog
      v-model="itemDialogVisible"
      :dialog-type="itemDialogType"
      :item-data="currentItemData"
      @success="handleItemDialogSuccess"
    />
  </ElCard>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    fetchDictItems,
    fetchCreateDictItem,
    fetchUpdateDictItem,
    fetchDeleteDictItem,
    type DictTypeSummary,
    type DictItemSummary
  } from '@/api/system-manage/dictionary'
  import { invalidateDict } from '@/hooks/business/useDictionary'
  import DictItemDialog from './dict-item-dialog.vue'

  interface Props {
    dictType: DictTypeSummary
  }

  interface Emits {
    (e: 'type-updated'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const loading = ref(false)
  const itemList = ref<DictItemSummary[]>([])

  const itemDialogVisible = ref(false)
  const itemDialogType = ref<'add' | 'edit'>('add')
  const currentItemData = ref<DictItemSummary | undefined>()
  const editingIndex = ref(-1)

  async function loadItems() {
    loading.value = true
    try {
      const items = await fetchDictItems(props.dictType.id)
      itemList.value = items as DictItemSummary[]
    } catch {
      ElMessage.error('加载字典项失败')
    } finally {
      loading.value = false
    }
  }

  function showItemDialog(type: 'add' | 'edit', item?: DictItemSummary, index?: number) {
    itemDialogType.value = type
    currentItemData.value = item
    editingIndex.value = index ?? -1
    itemDialogVisible.value = true
  }

  async function handleItemDialogSuccess(item: DictItemSummary) {
    try {
      if (itemDialogType.value === 'add') {
        const created = await fetchCreateDictItem(props.dictType.id, {
          label: item.label,
          value: item.value,
          description: item.description,
          is_default: item.is_default,
          status: (item.status as 'normal' | 'suspended') || 'normal',
          sort_order: Number(item.sort_order ?? itemList.value.length)
        })
        itemList.value.push(created)
        ElMessage.success('新增成功')
      } else if (editingIndex.value >= 0 && itemDataId(itemList.value[editingIndex.value])) {
        const current = itemList.value[editingIndex.value]
        const updated = await fetchUpdateDictItem(props.dictType.id, current.id, {
          label: item.label,
          value: item.value,
          description: item.description,
          is_default: item.is_default,
          status: (item.status as 'normal' | 'suspended') || 'normal',
          sort_order: Number(item.sort_order ?? editingIndex.value)
        })
        itemList.value.splice(editingIndex.value, 1, updated)
        ElMessage.success('保存成功')
      }
      invalidateDict(props.dictType.code)
      emit('type-updated')
    } catch (error) {
      if (error instanceof Error) {
        ElMessage.error(error.message)
      }
    }
  }

  async function handleToggleStatus(row: DictItemSummary, index: number) {
    const nextStatus = row.status === 'normal' ? 'suspended' : 'normal'
    const actionText = nextStatus === 'suspended' ? '停用' : '启用'
    try {
      await ElMessageBox.confirm(
        nextStatus === 'suspended'
          ? `确定停用字典项“${row.label}”吗？停用后才能继续删除。`
          : `确定启用字典项“${row.label}”吗？`,
        `${actionText}确认`,
        {
          type: nextStatus === 'suspended' ? 'warning' : 'info',
          confirmButtonText: `确定${actionText}`,
          cancelButtonText: '取消'
        }
      )
      const updated = await fetchUpdateDictItem(props.dictType.id, row.id, {
        label: row.label,
        value: row.value,
        description: row.description,
        is_default: row.is_default,
        status: nextStatus,
        sort_order: Number(row.sort_order ?? index)
      })
      itemList.value.splice(index, 1, updated)
      invalidateDict(props.dictType.code)
      emit('type-updated')
      ElMessage.success(`${actionText}成功`)
    } catch (error) {
      if (error !== 'cancel' && error instanceof Error) {
        ElMessage.error(error.message)
      }
    }
  }

  async function handleDeleteItem(row: DictItemSummary, index: number) {
    if (row.is_builtin) {
      ElMessage.warning('内置字典项不允许删除')
      return
    }
    if (row.status !== 'suspended') {
      ElMessage.warning('请先停用该字典项，再执行删除')
      return
    }
    try {
      await ElMessageBox.confirm(
        `确定删除字典项“${row.label}”吗？删除后不可恢复。`,
        '删除确认',
        {
          type: 'warning',
          confirmButtonText: '确定删除',
          cancelButtonText: '取消'
        }
      )
      await fetchDeleteDictItem(props.dictType.id, row.id)
      itemList.value.splice(index, 1)
      invalidateDict(props.dictType.code)
      emit('type-updated')
      ElMessage.success('删除成功')
    } catch (error) {
      if (error !== 'cancel' && error instanceof Error) {
        ElMessage.error(error.message)
      }
    }
  }

  function itemDataId(item?: DictItemSummary) {
    return `${item?.id || ''}`.trim()
  }

  onMounted(() => {
    loadItems()
  })
</script>

<style scoped lang="scss">
  .dict-item-panel {
    max-height: 100%;
    min-height: 0;
    display: flex;
    flex-direction: column;

    :deep(.el-card__body) {
      flex: 1 1 auto;
      min-height: 0;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
  }

  .dict-item-table-wrap {
    flex: 0 1 auto;
    min-height: 0;
    overflow: hidden;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .panel-header-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .panel-header-right {
    display: flex;
    gap: 8px;
  }

  .panel-title {
    font-weight: 600;
    font-size: 15px;
  }
</style>
