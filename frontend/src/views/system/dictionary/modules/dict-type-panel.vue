<template>
  <ElCard class="dict-type-panel" shadow="never">
    <template #header>
      <div class="panel-header">
        <span class="panel-title">字典类型</span>
        <ElButton type="primary" size="small" @click="showDialog('add')">
          新增
        </ElButton>
      </div>
    </template>

    <!-- Search -->
    <ElInput
      v-model="keyword"
      placeholder="搜索编码或名称..."
      clearable
      :prefix-icon="Search"
      @input="handleSearch"
      style="margin-bottom: 12px"
    />

    <!-- Type List -->
    <div v-loading="loading" class="type-list">
      <div
        v-for="item in typeList"
        :key="item.id"
        class="type-item"
        :class="{ 'is-active': selectedId === item.id }"
        @click="$emit('select', item)"
      >
        <div class="type-item-main">
          <div class="type-item-name">
            {{ item.name }}
            <ElTag v-if="item.is_builtin" size="small" type="info" style="margin-left: 6px">
              内置
            </ElTag>
          </div>
          <div class="type-item-code">{{ item.code }}</div>
        </div>
        <div class="type-item-actions">
          <span class="type-item-count">{{ item.item_count }} 项</span>
          <ElDropdown
            v-if="!item.is_builtin"
            trigger="click"
            @command="(cmd: string) => handleCommand(cmd, item)"
            @click.stop
          >
            <ElButton :icon="MoreFilled" text size="small" @click.stop />
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="edit">编辑</ElDropdownItem>
                <ElDropdownItem command="delete" divided style="color: var(--el-color-danger)">
                  删除
                </ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
          <ElDropdown
            v-else
            trigger="click"
            @command="(cmd: string) => handleCommand(cmd, item)"
            @click.stop
          >
            <ElButton :icon="MoreFilled" text size="small" @click.stop />
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="edit">编辑</ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
        </div>
      </div>

      <ElEmpty v-if="!loading && typeList.length === 0" description="暂无数据" />
    </div>

    <!-- Pagination -->
    <div v-if="total > pageSize" class="type-pagination">
      <ElPagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="prev, pager, next"
        small
        @current-change="loadData"
      />
    </div>

    <!-- Dialog -->
    <DictTypeDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :type-data="currentTypeData"
      @success="handleDialogSuccess"
    />
  </ElCard>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { Search, MoreFilled } from '@element-plus/icons-vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    fetchDictTypeList,
    fetchDeleteDictType,
    type DictTypeSummary
  } from '@/api/system-manage/dictionary'
  import DictTypeDialog from './dict-type-dialog.vue'

  interface Props {
    selectedId: string
  }

  interface Emits {
    (e: 'select', type: DictTypeSummary): void
    (e: 'refresh'): void
  }

  defineProps<Props>()
  const emit = defineEmits<Emits>()

  const loading = ref(false)
  const typeList = ref<DictTypeSummary[]>([])
  const keyword = ref('')
  const currentPage = ref(1)
  const pageSize = 20
  const total = ref(0)

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentTypeData = ref<DictTypeSummary | undefined>()

  let searchTimer: ReturnType<typeof setTimeout> | null = null

  async function loadData() {
    loading.value = true
    try {
      const res = await fetchDictTypeList({
        current: currentPage.value,
        size: pageSize,
        keyword: keyword.value || undefined
      })
      typeList.value = res.records
      total.value = res.total
    } catch {
      ElMessage.error('加载字典类型失败')
    } finally {
      loading.value = false
    }
  }

  function handleSearch() {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      currentPage.value = 1
      loadData()
    }, 300)
  }

  function showDialog(type: 'add' | 'edit', data?: DictTypeSummary) {
    dialogType.value = type
    currentTypeData.value = data
    dialogVisible.value = true
  }

  function handleCommand(cmd: string, item: DictTypeSummary) {
    if (cmd === 'edit') {
      showDialog('edit', item)
    } else if (cmd === 'delete') {
      ElMessageBox.confirm(`确定删除字典类型「${item.name}」吗？该操作将同时删除所有字典项。`, '删除确认', {
        type: 'warning',
        confirmButtonText: '确定删除',
        cancelButtonText: '取消'
      })
        .then(async () => {
          await fetchDeleteDictType(item.id)
          ElMessage.success('删除成功')
          emit('refresh')
          loadData()
        })
        .catch(() => {})
    }
  }

  function handleDialogSuccess() {
    loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>

<style scoped lang="scss">
  .dict-type-panel {
    height: 100%;
    display: flex;
    flex-direction: column;

    :deep(.el-card__body) {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .panel-title {
    font-weight: 600;
    font-size: 15px;
  }

  .type-list {
    flex: 1;
    overflow-y: auto;
  }

  .type-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: background-color 0.2s;

    &:hover {
      background-color: var(--el-fill-color-light);
    }

    &.is-active {
      background-color: var(--el-color-primary-light-9);
    }
  }

  .type-item-main {
    flex: 1;
    min-width: 0;
  }

  .type-item-name {
    font-size: 14px;
    font-weight: 500;
    display: flex;
    align-items: center;
  }

  .type-item-code {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 2px;
  }

  .type-item-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }

  .type-item-count {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
  }

  .type-pagination {
    padding-top: 12px;
    display: flex;
    justify-content: center;
  }
</style>
