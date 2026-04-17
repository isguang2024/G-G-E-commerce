<template>
  <div class="dictionary-page art-full-height">
    <div ref="layoutRef" class="dictionary-layout" :class="{ 'is-stacked': isStacked }">
      <ElSplitter class="dictionary-splitter-layout" :layout="isStacked ? 'vertical' : 'horizontal'" lazy>
        <!-- ══════════════ 左侧：字典管理 ══════════════ -->
        <ElSplitterPanel
          v-model:size="typePanelSize"
          class="split-panel"
          collapsible
          :min="typePanelMin"
        >
          <div class="panel">
            <div class="panel__header">
              <span class="panel__title">字典管理</span>
            </div>

            <!-- 搜索 + 查询 + 新增字典 -->
            <div class="panel__bar">
              <ElInput
                v-model="typeKeyword"
                clearable
                placeholder="请输入字典名称或字典编码"
                class="panel__bar-input"
                @keyup.enter="loadTypeList(true)"
                @clear="loadTypeList(true)"
              >
                <template #prefix><ElIcon><Search /></ElIcon></template>
              </ElInput>
              <ElButton type="primary" @click="loadTypeList(true)">查询</ElButton>
              <ElButton type="primary" :icon="Plus" @click="openTypeAdd">新增字典</ElButton>
            </div>

            <div class="panel__meta">共 {{ typeTotal }} 条记录</div>

            <div class="panel__table-wrap">
              <ArtTable
                :data="typeList"
                :loading="typeLoading"
                :show-table-header="false"
                :pagination="typePagination"
                :row-class-name="typeRowClassName"
                highlight-current-row
                :current-row-key="selectedTypeId"
                empty-text="暂无字典类型"
                @pagination:size-change="handleTypeSizeChange"
                @pagination:current-change="handleTypeCurrentChange"
                @row-click="handleTypeRowClick"
              >
                <ElTableColumn label="序号" width="56" align="center">
                  <template #default="{ $index }">
                    {{ (typePage.current - 1) * typePage.size + $index + 1 }}
                  </template>
                </ElTableColumn>
                <ElTableColumn prop="name" label="字典名称" min-width="130" show-overflow-tooltip>
                  <template #default="{ row }">
                    <div class="name-cell">
                      <span>{{ row.name }}</span>
                      <ElTag v-if="row.is_builtin" size="small" type="info" effect="plain" class="name-cell__badge">内置</ElTag>
                    </div>
                  </template>
                </ElTableColumn>
                <ElTableColumn prop="code" label="字典编码" min-width="150" show-overflow-tooltip>
                  <template #default="{ row }"><span class="mono">{{ row.code }}</span></template>
                </ElTableColumn>
                <ElTableColumn label="状态" width="76" align="center">
                  <template #default="{ row }">
                    <ElTag size="small" :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</ElTag>
                  </template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="100" align="center" fixed="right">
                  <template #default="{ row }">
                    <ElButton text size="small" type="primary" @click.stop="openTypeEdit(row)">编辑</ElButton>
                    <ElPopconfirm
                      v-if="!row.is_builtin"
                      :title="`确认删除字典「${row.name}」？所有字典项也会被删除。`"
                      @confirm="handleDeleteType(row)"
                    >
                      <template #reference>
                        <ElButton text size="small" type="danger" @click.stop>停用</ElButton>
                      </template>
                    </ElPopconfirm>
                    <ElTooltip v-else content="内置字典不允许删除" placement="top">
                      <ElButton text size="small" type="danger" disabled @click.stop>停用</ElButton>
                    </ElTooltip>
                  </template>
                </ElTableColumn>
              </ArtTable>
            </div>
          </div>
        </ElSplitterPanel>

        <!-- ══════════════ 右侧：数据项管理 ══════════════ -->
        <ElSplitterPanel
          v-model:size="itemPanelSize"
          class="split-panel"
          collapsible
          :min="itemPanelMin"
        >
          <div class="panel">
            <div class="panel__header">
              <span class="panel__title">数据项管理</span>
            </div>

            <!-- 过滤区：两列带标签的输入 -->
            <div class="panel__filter">
              <div class="panel__filter-col">
                <span class="panel__filter-label">数据标签</span>
                <ElInput
                  v-model="itemFilter.label"
                  clearable
                  placeholder="请输入标签"
                  :disabled="!activeDetail"
                  @keyup.enter="applyItemFilter"
                />
              </div>
              <div class="panel__filter-col">
                <span class="panel__filter-label">数据编码</span>
                <ElInput
                  v-model="itemFilter.value"
                  clearable
                  placeholder="请输入编码"
                  :disabled="!activeDetail"
                  @keyup.enter="applyItemFilter"
                />
              </div>
            </div>

            <!-- 操作行：重置 + 查询 靠左，新增数据项 靠右 -->
            <div class="panel__bar">
              <ElButton :icon="RefreshRight" :disabled="!activeDetail" @click="resetItemFilter()">重置</ElButton>
              <ElButton type="primary" :disabled="!activeDetail" @click="applyItemFilter">查询</ElButton>
              <div class="panel__bar-spacer" />
              <ElButton type="primary" :icon="Plus" :disabled="!activeDetail" @click="openItemAdd">新增数据项</ElButton>
            </div>

            <div class="panel__meta">
              共 {{ rawItems.length }} 条数据项
              <template v-if="filteredItems.length !== rawItems.length">，匹配 {{ filteredItems.length }} 条</template>
            </div>

            <div class="panel__table-wrap">
              <ArtTable
                :data="pagedItems"
                :loading="detailLoading"
                :show-table-header="false"
                :pagination="itemPagination"
                :empty-text="activeDetail ? '暂无数据' : '请从左侧选择一个字典类型'"
                @pagination:size-change="handleItemSizeChange"
                @pagination:current-change="handleItemCurrentChange"
              >
                <ElTableColumn label="#" width="56" align="center">
                  <template #default="{ $index }">
                    {{ (itemPage.current - 1) * itemPage.size + $index + 1 }}
                  </template>
                </ElTableColumn>
                <ElTableColumn prop="label" label="标签" min-width="130">
                  <template #default="{ row }">
                    <div class="item-label-cell">
                      <span>{{ row.label }}</span>
                      <ElTag v-if="row.is_default" size="small" type="success" effect="light" class="item-label-cell__badge">默认</ElTag>
                      <ElTag v-if="row.is_builtin" size="small" type="info" effect="plain" class="item-label-cell__badge">内置</ElTag>
                    </div>
                  </template>
                </ElTableColumn>
                <ElTableColumn prop="value" label="数值" min-width="130">
                  <template #default="{ row }"><span class="mono">{{ row.value }}</span></template>
                </ElTableColumn>
                <ElTableColumn prop="description" label="备注" min-width="160" show-overflow-tooltip>
                  <template #default="{ row }">{{ row.description || '-' }}</template>
                </ElTableColumn>
                <ElTableColumn prop="sort_order" label="排序" width="68" align="center">
                  <template #default="{ row }">{{ row.sort_order ?? 0 }}</template>
                </ElTableColumn>
                <ElTableColumn label="状态" width="80" align="center">
                  <template #default="{ row }">
                    <ElTag size="small" :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</ElTag>
                  </template>
                </ElTableColumn>
                <ElTableColumn label="更新时间" width="155">
                  <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="160" fixed="right" align="center">
                  <template #default="{ row }">
                    <ElButton text size="small" type="primary" @click="openItemEdit(row)">编辑</ElButton>
                    <ElButton text size="small" :type="row.status === 'normal' ? 'warning' : 'success'" @click="toggleItemStatus(row)">
                      {{ row.status === 'normal' ? '停用' : '启用' }}
                    </ElButton>
                    <ElPopconfirm
                      v-if="!row.is_builtin && row.status === 'suspended'"
                      title="删除后不可恢复，确定删除？"
                      @confirm="removeItem(row)"
                    >
                      <template #reference>
                        <ElButton text size="small" type="danger">删除</ElButton>
                      </template>
                    </ElPopconfirm>
                    <ElTooltip v-else-if="!row.is_builtin" content="请先停用该字典项，再执行删除" placement="top">
                      <ElButton text size="small" type="danger" disabled>删除</ElButton>
                    </ElTooltip>
                  </template>
                </ElTableColumn>
              </ArtTable>
            </div>
          </div>
        </ElSplitterPanel>
      </ElSplitter>
    </div>

    <!-- 字典类型编辑抽屉 -->
    <DictTypeDialog
      v-model="typeDialogVisible"
      :dialog-type="typeDialogType"
      :type-data="currentTypeData"
      @success="handleTypeDialogSuccess"
    />

    <!-- 字典项编辑抽屉 -->
    <DictItemDialog
      v-model="itemDialogVisible"
      :dialog-type="itemDialogType"
      :item-data="currentItemData"
      @success="handleItemDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useResizeObserver } from '@vueuse/core'
  import { Plus, RefreshRight, Search } from '@element-plus/icons-vue'
  import {
    fetchCreateDictItem,
    fetchDeleteDictItem,
    fetchDeleteDictType,
    fetchDictTypeList,
    fetchGetDictType,
    fetchUpdateDictItem,
    type DictItemSummary,
    type DictTypeDetail,
    type DictTypeSummary
  } from '@/api/system-manage/dictionary'
  import { invalidateDict } from '@/hooks/business/useDictionary'
  import DictTypeDialog from './modules/dict-type-dialog.vue'
  import DictItemDialog from './modules/dict-item-dialog.vue'

  defineOptions({ name: 'SystemDictionary' })

  const MIN_TYPE_PANEL_WIDTH = 320
  const MIN_ITEM_PANEL_WIDTH = 480
  const MIN_STACKED_TYPE_HEIGHT = 260
  const MIN_STACKED_ITEM_HEIGHT = 320

  const layoutRef = ref<HTMLElement>()
  const layoutWidth = ref(0)
  const typePanelSize = ref<string | number>('42%')
  const itemPanelSize = ref<string | number>('58%')

  const isStacked = computed(() => layoutWidth.value > 0 && layoutWidth.value < 1024)
  const typePanelMin = computed(() => (isStacked.value ? MIN_STACKED_TYPE_HEIGHT : MIN_TYPE_PANEL_WIDTH))
  const itemPanelMin = computed(() => (isStacked.value ? MIN_STACKED_ITEM_HEIGHT : MIN_ITEM_PANEL_WIDTH))

  useResizeObserver(layoutRef, (entries) => {
    const entry = entries[0]
    if (!entry) return
    layoutWidth.value = entry.contentRect.width
  })

  // ── 字典类型列表 ─────────────────────────────────────────────────────────

  const typeLoading = ref(false)
  const typeList = ref<DictTypeSummary[]>([])
  const typeTotal = ref(0)
  const typeKeyword = ref('')
  const typePage = reactive({ current: 1, size: 20 })
  const selectedTypeId = ref('')

  const typePagination = computed(() => ({
    current: typePage.current,
    size: typePage.size,
    total: typeTotal.value
  }))

  async function loadTypeList(resetPage = false) {
    if (resetPage) typePage.current = 1
    typeLoading.value = true
    try {
      const res = await fetchDictTypeList({
        current: typePage.current,
        size: typePage.size,
        keyword: typeKeyword.value.trim() || undefined
      })
      if (res.total > 0 && res.records.length === 0 && typePage.current > 1) {
        typePage.current -= 1
        await loadTypeList()
        return
      }
      typeList.value = res.records
      typeTotal.value = res.total

      if (
        !selectedTypeId.value ||
        !res.records.some((item) => item.id === selectedTypeId.value)
      ) {
        if (res.records[0]) {
          selectTypeRow(res.records[0])
        } else {
          selectedTypeId.value = ''
          activeDetail.value = null
        }
      }
    } catch {
      ElMessage.error('加载字典类型失败')
    } finally {
      typeLoading.value = false
    }
  }

  function handleTypeSizeChange(val: number) {
    typePage.size = val
    typePage.current = 1
    loadTypeList()
  }

  function handleTypeCurrentChange(val: number) {
    typePage.current = val
    loadTypeList()
  }

  function typeRowClassName({ row }: { row: DictTypeSummary }) {
    return row.id === selectedTypeId.value ? 'is-active-row' : ''
  }

  function handleTypeRowClick(row: DictTypeSummary) {
    if (row.id === selectedTypeId.value) return
    selectTypeRow(row)
  }

  function selectTypeRow(row: DictTypeSummary) {
    selectedTypeId.value = row.id
    resetItemFilter(false)
    itemPage.current = 1
    loadActiveDetail(row.id)
  }

  // ── 字典类型编辑 ─────────────────────────────────────────────────────────

  const typeDialogVisible = ref(false)
  const typeDialogType = ref<'add' | 'edit'>('add')
  const currentTypeData = ref<DictTypeSummary | undefined>()

  function openTypeAdd() {
    typeDialogType.value = 'add'
    currentTypeData.value = undefined
    typeDialogVisible.value = true
  }

  function openTypeEdit(row: DictTypeSummary) {
    typeDialogType.value = 'edit'
    currentTypeData.value = { ...row }
    typeDialogVisible.value = true
  }

  async function handleTypeDialogSuccess() {
    await loadTypeList()
    if (selectedTypeId.value) await loadActiveDetail(selectedTypeId.value)
  }

  async function handleDeleteType(row: DictTypeSummary) {
    try {
      await fetchDeleteDictType(row.id)
      ElMessage.success('删除成功')
      if (selectedTypeId.value === row.id) {
        selectedTypeId.value = ''
        activeDetail.value = null
      }
      await loadTypeList()
    } catch (err) {
      if (err instanceof Error) ElMessage.error(err.message)
    }
  }

  // ── 字典项数据 ───────────────────────────────────────────────────────────

  const detailLoading = ref(false)
  const activeDetail = ref<DictTypeDetail | null>(null)

  const itemFilter = reactive({ label: '', value: '' })
  const appliedFilter = reactive({ label: '', value: '' })
  const itemPage = reactive({ current: 1, size: 20 })

  const itemPagination = computed(() => ({
    current: itemPage.current,
    size: itemPage.size,
    total: filteredItems.value.length
  }))

  async function loadActiveDetail(id: string) {
    detailLoading.value = true
    try {
      activeDetail.value = await fetchGetDictType(id)
    } catch {
      ElMessage.error('加载字典详情失败')
      activeDetail.value = null
    } finally {
      detailLoading.value = false
    }
  }

  async function refreshActiveDetail() {
    if (!selectedTypeId.value) return
    await loadActiveDetail(selectedTypeId.value)
  }

  const rawItems = computed<DictItemSummary[]>(() => {
    const list = activeDetail.value?.items ?? []
    return [...list].sort((a, b) => {
      const sortA = a.sort_order ?? 0
      const sortB = b.sort_order ?? 0
      if (sortA !== sortB) return sortA - sortB
      return a.label.localeCompare(b.label, 'zh-CN')
    })
  })

  const filteredItems = computed<DictItemSummary[]>(() => {
    const labelQ = appliedFilter.label.trim().toLowerCase()
    const valueQ = appliedFilter.value.trim().toLowerCase()
    return rawItems.value.filter((item) => {
      const matchLabel = !labelQ || item.label.toLowerCase().includes(labelQ)
      const matchValue = !valueQ || item.value.toLowerCase().includes(valueQ)
      return matchLabel && matchValue
    })
  })

  const pagedItems = computed(() => {
    const start = (itemPage.current - 1) * itemPage.size
    return filteredItems.value.slice(start, start + itemPage.size)
  })

  const itemPageCount = computed(() => Math.max(1, Math.ceil(filteredItems.value.length / itemPage.size)))

  watch(
    () => [filteredItems.value.length, itemPage.size],
    () => {
      if (itemPage.current > itemPageCount.value) itemPage.current = itemPageCount.value
    }
  )

  function applyItemFilter() {
    appliedFilter.label = itemFilter.label
    appliedFilter.value = itemFilter.value
    itemPage.current = 1
  }

  function resetItemFilter(apply = true) {
    itemFilter.label = ''
    itemFilter.value = ''
    appliedFilter.label = ''
    appliedFilter.value = ''
    if (apply) itemPage.current = 1
  }

  function handleItemSizeChange(val: number) {
    itemPage.size = val
    itemPage.current = 1
  }

  function handleItemCurrentChange(val: number) {
    itemPage.current = val
  }

  // ── 字典项 CRUD ─────────────────────────────────────────────────────────

  const itemDialogVisible = ref(false)
  const itemDialogType = ref<'add' | 'edit'>('add')
  const currentItemData = ref<DictItemSummary | undefined>()

  function openItemAdd() {
    if (!activeDetail.value) {
      ElMessage.warning('请先选择一个字典类型')
      return
    }
    itemDialogType.value = 'add'
    currentItemData.value = undefined
    itemDialogVisible.value = true
  }

  function openItemEdit(row: DictItemSummary) {
    itemDialogType.value = 'edit'
    currentItemData.value = { ...row }
    itemDialogVisible.value = true
  }

  async function handleItemDialogSuccess(item: DictItemSummary) {
    if (!activeDetail.value) return
    try {
      if (itemDialogType.value === 'add') {
        await fetchCreateDictItem(activeDetail.value.id, {
          label: item.label,
          value: item.value,
          description: item.description,
          is_default: item.is_default,
          status: (item.status as 'normal' | 'suspended') || 'normal',
          sort_order: Number(item.sort_order ?? rawItems.value.length)
        })
        ElMessage.success('新增成功')
      } else {
        await fetchUpdateDictItem(activeDetail.value.id, item.id, {
          label: item.label,
          value: item.value,
          description: item.description,
          is_default: item.is_default,
          status: (item.status as 'normal' | 'suspended') || 'normal',
          sort_order: Number(item.sort_order ?? 0)
        })
        ElMessage.success('保存成功')
      }
      invalidateDict(activeDetail.value.code)
      await refreshActiveDetail()
      await loadTypeList()
    } catch (err) {
      if (err instanceof Error) ElMessage.error(err.message)
    }
  }

  async function toggleItemStatus(item: DictItemSummary) {
    if (!activeDetail.value) return
    const next = item.status === 'normal' ? 'suspended' : 'normal'
    const actionText = next === 'suspended' ? '停用' : '启用'
    try {
      await ElMessageBox.confirm(
        next === 'suspended'
          ? `确定停用字典项「${item.label}」？停用后才能继续删除。`
          : `确定启用字典项「${item.label}」？`,
        `${actionText}确认`,
        {
          type: next === 'suspended' ? 'warning' : 'info',
          confirmButtonText: `确定${actionText}`,
          cancelButtonText: '取消'
        }
      )
      await fetchUpdateDictItem(activeDetail.value.id, item.id, {
        label: item.label,
        value: item.value,
        description: item.description,
        is_default: item.is_default,
        status: next,
        sort_order: Number(item.sort_order ?? 0)
      })
      invalidateDict(activeDetail.value.code)
      ElMessage.success(`${actionText}成功`)
      await refreshActiveDetail()
    } catch (err) {
      if (err !== 'cancel' && err instanceof Error) ElMessage.error(err.message)
    }
  }

  async function removeItem(item: DictItemSummary) {
    if (!activeDetail.value) return
    try {
      await fetchDeleteDictItem(activeDetail.value.id, item.id)
      invalidateDict(activeDetail.value.code)
      ElMessage.success('删除成功')
      await refreshActiveDetail()
      await loadTypeList()
    } catch (err) {
      if (err instanceof Error) ElMessage.error(err.message)
    }
  }

  // ── 工具函数 ─────────────────────────────────────────────────────────────

  function statusLabel(status?: string) {
    return status === 'suspended' ? '停用' : '正常'
  }

  function statusTagType(status?: string): 'success' | 'info' {
    return status === 'suspended' ? 'info' : 'success'
  }

  function formatDateTime(value?: string) {
    if (!value) return '-'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date)
  }

  onMounted(() => {
    loadTypeList()
  })
</script>

<style scoped lang="scss">
  .dictionary-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
    padding: 8px;
    box-sizing: border-box;
    overflow: hidden;
  }

  .dictionary-layout {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .dictionary-splitter-layout {
    height: 100%;
    min-height: 0;
  }

  .split-panel {
    min-height: 0;
    overflow: hidden;

    :deep(.el-splitter-panel__wrapper) {
      height: 100%;
      min-height: 0;
      overflow: hidden;
    }
  }

  .dictionary-layout :deep(.el-splitter-panel) {
    min-width: 0;
    min-height: 0;
  }

  .dictionary-layout :deep(.el-splitter__horizontal > .el-splitter-bar) {
    width: 12px;
  }

  .dictionary-layout :deep(.el-splitter__vertical > .el-splitter-bar) {
    height: 12px;
  }

  .dictionary-layout :deep(.el-splitter-bar) {
    background: var(--el-fill-color-extra-light);
    transition: background-color 0.2s ease;
  }

  .dictionary-layout :deep(.el-splitter-bar:hover) {
    background: var(--el-color-primary-light-9);
  }

  .dictionary-layout :deep(.el-splitter-bar__collapse-icon) {
    opacity: 1;
    background: var(--el-bg-color-overlay);
    border: 1px solid var(--el-border-color-light);
    border-radius: 999px;
    box-shadow: 0 4px 12px rgb(15 23 42 / 8%);
  }

  // ── 面板通用布局 ─────────────────────────────────────────────────────────

  .panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    overflow: hidden;
    background: var(--el-bg-color);
    border-radius: 8px;
    box-sizing: border-box;
    padding: 16px 16px 0;
    gap: 12px;
  }

  // 面板标题
  .panel__header {
    flex-shrink: 0;
  }

  .panel__title {
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    line-height: 1.4;
  }

  // 统一工具栏行
  .panel__bar {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
    min-width: 0;
  }

  .panel__bar-input {
    flex: 1;
    min-width: 0;
  }

  .panel__bar-spacer {
    flex: 1;
  }

  // 计数行
  .panel__meta {
    flex-shrink: 0;
    font-size: 13px;
    color: var(--el-text-color-secondary);
    line-height: 1;
    margin-top: -4px;
  }

  // 右侧过滤区：两列标签 + 输入
  .panel__filter {
    display: flex;
    gap: 12px;
    flex-shrink: 0;
  }

  .panel__filter-col {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .panel__filter-label {
    font-size: 13px;
    color: var(--el-text-color-regular);
    line-height: 1;
  }

  .panel__table-wrap {
    flex: 1 1 auto;
    min-height: 0;
    overflow: hidden;
  }

  // ── 单元格样式 ───────────────────────────────────────────────────────────

  .name-cell {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .name-cell__badge {
    flex-shrink: 0;
  }

  .item-label-cell {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .item-label-cell__badge {
    flex-shrink: 0;
  }

  .mono {
    font-family: var(--el-font-family-mono, ui-monospace, SFMono-Regular, monospace);
    font-size: 12px;
    color: var(--el-text-color-regular);
  }

  :deep(.is-active-row) > td {
    background: var(--el-color-primary-light-9) !important;
  }

  :deep(.el-table__row) {
    cursor: pointer;
  }
</style>
