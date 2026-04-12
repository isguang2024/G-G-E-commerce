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
          <ElButton type="primary" size="small" :loading="saving" @click="handleBatchSave">
            批量保存
          </ElButton>
        </div>
      </div>
    </template>

    <ElTable
      v-loading="loading"
      :data="itemList"
      border
      style="width: 100%"
      row-key="id"
    >
      <ElTableColumn prop="sort_order" label="排序" width="70" align="center" />
      <ElTableColumn prop="label" label="标签" min-width="120" />
      <ElTableColumn prop="value" label="值" min-width="120" />
      <ElTableColumn prop="is_default" label="默认" width="70" align="center">
        <template #default="{ row }">
          <ElTag v-if="row.is_default" size="small" type="success">是</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="status" label="状态" width="80" align="center">
        <template #default="{ row }">
          <ElTag :type="row.status === 'normal' ? 'success' : 'info'" size="small">
            {{ row.status === 'normal' ? '正常' : '停用' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" width="120" align="center" fixed="right">
        <template #default="{ row, $index }">
          <ElButton text size="small" @click="showItemDialog('edit', row, $index)">编辑</ElButton>
          <ElButton text size="small" type="danger" @click="handleRemoveItem($index)">删除</ElButton>
        </template>
      </ElTableColumn>
    </ElTable>

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
  import { ElMessage } from 'element-plus'
  import {
    fetchDictItems,
    fetchSaveDictItems,
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
  const saving = ref(false)
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

  function handleItemDialogSuccess(item: DictItemSummary) {
    if (itemDialogType.value === 'add') {
      item.sort_order = itemList.value.length
      itemList.value.push(item)
    } else if (editingIndex.value >= 0) {
      itemList.value.splice(editingIndex.value, 1, item)
    }
  }

  function handleRemoveItem(index: number) {
    itemList.value.splice(index, 1)
  }

  async function handleBatchSave() {
    saving.value = true
    const snapshot = JSON.parse(JSON.stringify(itemList.value)) as DictItemSummary[]
    try {
      const body = {
        items: itemList.value.map((item, index) => ({
          label: item.label,
          value: item.value,
          is_default: item.is_default,
          status: item.status as 'normal' | 'suspended',
          sort_order: item.sort_order ?? index
        }))
      }
      const result = await fetchSaveDictItems(props.dictType.id, body)
      itemList.value = result as DictItemSummary[]
      invalidateDict(props.dictType.code)
      ElMessage.success('保存成功')
      emit('type-updated')
    } catch (error) {
      itemList.value = snapshot
      if (error instanceof Error) {
        ElMessage.error(error.message)
      }
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadItems()
  })
</script>

<style scoped lang="scss">
  .dict-item-panel {
    height: 100%;

    :deep(.el-card__body) {
      overflow: auto;
    }
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
