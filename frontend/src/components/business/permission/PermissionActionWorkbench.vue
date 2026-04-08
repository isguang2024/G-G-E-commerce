<template>
  <div class="permission-workbench">
    <div class="workbench-toolbar">
      <div class="toolbar-main">
        <ElInput
          v-model="searchKeyword"
          :placeholder="searchPlaceholder"
          clearable
          class="toolbar-search"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>

        <ElSelect v-model="featureFilter" placeholder="功能归属" clearable class="toolbar-select">
          <ElOption label="全部归属" value="" />
          <ElOption
            v-for="item in featureOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>

        <ElSelect
          v-model="stateFilter"
          :placeholder="stateFilterPlaceholder"
          class="toolbar-select"
        >
          <ElOption
            v-for="item in stateOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </div>

      <div class="toolbar-extra">
        <div class="toolbar-switch">
          <span>显示 ID/说明</span>
          <ElSwitch v-model="showMeta" size="small" />
        </div>
        <div class="toolbar-switch">
          <span>紧凑模式</span>
          <ElSwitch v-model="compactMode" size="small" />
        </div>
      </div>
    </div>

    <div class="workbench-summary">
      <div class="summary-tags">
        <ElTag effect="plain" round>可见 {{ visibleCount }}</ElTag>
        <ElTag type="primary" effect="plain" round>{{ selectedLabel }} {{ selectedCount }}</ElTag>
        <ElTag type="info" effect="plain" round>总计 {{ totalCount }}</ElTag>
      </div>

      <div class="summary-actions">
        <ElButton text @click="expandAll">全部展开</ElButton>
        <ElButton text @click="collapseAll">全部收起</ElButton>
        <ElDropdown trigger="click" @command="handleBatchCommand">
          <ElButton text>
            批量操作
            <ElIcon class="el-icon--right"><ArrowDown /></ElIcon>
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem
                v-for="item in batchCommands"
                :key="item.command"
                :command="item.command"
              >
                {{ item.label }}
              </ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </div>
    </div>

    <div class="workbench-body" v-loading="loading">
      <ElEmpty v-if="groupedActions.length === 0" description="暂无可配置权限" />

      <ElScrollbar v-else max-height="560px">
        <ElCollapse v-model="activeFeatures" class="feature-collapse">
          <ElCollapseItem
            v-for="feature in groupedActions"
            :key="feature.key"
            :name="feature.key"
            class="feature-item"
          >
            <template #title>
              <div class="feature-header">
                <div>
                  <div class="feature-title">{{ feature.label }}</div>
                  <div class="feature-subtitle">
                    {{ feature.moduleCount }} 个模块，{{ feature.actionCount }} 条权限
                  </div>
                </div>
                <ElTag type="info" effect="plain" round>{{ feature.selectedCount }}</ElTag>
              </div>
            </template>

            <div class="module-list">
              <section v-for="module in feature.modules" :key="module.key" class="module-card">
                <header class="module-header">
                  <div>
                    <div class="module-title">{{ module.label }}</div>
                    <div class="module-subtitle"> {{ module.actionCount }} 条权限 </div>
                  </div>
                  <ElTag type="info" effect="plain" round>{{ module.selectedCount }}</ElTag>
                </header>

                <div class="action-list" :class="{ compact: compactMode }">
                  <div
                    v-for="action in module.actions"
                    :key="action.id"
                    class="action-row"
                    :class="{ compact: compactMode }"
                  >
                    <div class="action-main">
                      <div class="action-title-line">
                        <ElCheckbox
                          v-if="mode === 'menu'"
                          :model-value="isSelected(action.id)"
                          @change="(value) => toggleSelection(action.id, Boolean(value))"
                        />
                        <ElSwitch
                          v-else-if="mode === 'collaboration'"
                          :model-value="isSelected(action.id)"
                          size="small"
                          @change="(value) => toggleSelection(action.id, Boolean(value))"
                        />
                        <span class="action-title">{{ action.name }}</span>
                      </div>

                      <div v-if="showMeta" class="action-meta">
                        <span v-if="action.permissionKey || action.code">
                          {{ action.permissionKey || action.code }}
                        </span>
                        <span v-if="action.description">{{ action.description }}</span>
                      </div>
                    </div>

                    <div class="action-side">
                      <template v-if="mode === 'menu' || mode === 'collaboration'">
                        <ElTag
                          :type="isSelected(action.id) ? 'success' : 'info'"
                          effect="plain"
                          round
                        >
                          {{ isSelected(action.id) ? positiveText : neutralText }}
                        </ElTag>
                      </template>

                      <template v-else>
                        <ElDropdown
                          trigger="click"
                          @command="(command) => setDecision(action.id, `${command || ''}`)"
                        >
                          <ElTag
                            class="decision-tag"
                            :type="getDecisionTagType(action.id)"
                            effect="light"
                            round
                          >
                            {{ getDecisionLabel(action.id) }}
                            <ElIcon class="el-icon--right"><ArrowDown /></ElIcon>
                          </ElTag>
                          <template #dropdown>
                            <ElDropdownMenu>
                              <ElDropdownItem
                                v-for="item in decisionOptions"
                                :key="item.value"
                                :command="item.value"
                              >
                                {{ item.label }}
                              </ElDropdownItem>
                            </ElDropdownMenu>
                          </template>
                        </ElDropdown>
                      </template>
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </ElCollapseItem>
        </ElCollapse>
      </ElScrollbar>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ArrowDown, Search } from '@element-plus/icons-vue'
  import type {
    DecisionValue,
    WorkbenchActionItem,
    WorkbenchMode
  } from './permission-action-workbench.helpers'
  import { usePermissionActionWorkbench } from './use-permission-action-workbench'

  interface Props {
    mode: WorkbenchMode
    actions: WorkbenchActionItem[]
    loading?: boolean
    selectedIds?: string[]
    decisionMap?: Record<string, DecisionValue>
    searchPlaceholder?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
    selectedIds: () => [],
    decisionMap: () => ({}),
    searchPlaceholder: '搜索权限名称、权限 ID、模块'
  })

  const emit = defineEmits<{
    (e: 'update:selectedIds', value: string[]): void
    (e: 'update:decisionMap', value: Record<string, DecisionValue>): void
  }>()

  const {
    searchKeyword,
    featureFilter,
    stateFilter,
    showMeta,
    compactMode,
    activeFeatures,
    featureOptions,
    groupedActions,
    totalCount,
    visibleCount,
    selectedCount,
    selectedLabel,
    positiveText,
    neutralText,
    decisionOptions,
    stateOptions,
    stateFilterPlaceholder,
    batchCommands,
    isSelected,
    toggleSelection,
    setDecision,
    getDecisionLabel,
    getDecisionTagType,
    expandAll,
    collapseAll,
    handleBatchCommand
  } = usePermissionActionWorkbench(props, emit)
</script>

<style scoped lang="scss">
  .permission-workbench {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .workbench-toolbar,
  .workbench-summary,
  .workbench-body {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 16px;
    background: #fff;
  }

  .workbench-toolbar {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    padding: 16px;
    background: #fbfcfe;
  }

  .toolbar-main {
    display: grid;
    flex: 1;
    grid-template-columns: minmax(260px, 1.8fr) repeat(3, minmax(140px, 0.8fr));
    gap: 12px;
  }

  .toolbar-extra {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .toolbar-search,
  .toolbar-select {
    width: 100%;
  }

  .toolbar-switch {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--el-text-color-regular);
    font-size: 13px;
    white-space: nowrap;
  }

  .workbench-summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 16px;
  }

  .summary-tags,
  .summary-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .workbench-body {
    padding: 8px;
    background: linear-gradient(180deg, #ffffff 0%, #fbfcff 100%);
  }

  .feature-collapse {
    border: 0;
  }

  .feature-item {
    margin-bottom: 12px;
    border: 1px solid #edf1f5;
    border-radius: 14px;
    overflow: hidden;
    background: #fff;
  }

  .feature-item :deep(.el-collapse-item__header) {
    height: auto;
    min-height: 72px;
    padding: 0 20px;
    border-bottom: 0;
    background: #f8fbff;
  }

  .feature-item :deep(.el-collapse-item__wrap) {
    border-bottom: 0;
  }

  .feature-item :deep(.el-collapse-item__content) {
    padding-bottom: 0;
  }

  .feature-header,
  .module-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
  }

  .feature-title,
  .module-title,
  .action-title {
    color: #111827;
    font-weight: 600;
  }

  .feature-subtitle,
  .module-subtitle,
  .action-meta {
    color: #6b7280;
    font-size: 13px;
  }

  .module-list {
    padding: 0 14px 14px;
  }

  .module-card {
    margin-top: 12px;
    border: 1px solid #f0f3f6;
    border-radius: 12px;
    overflow: hidden;
  }

  .module-header {
    padding: 14px 16px;
    background: #fcfdff;
    border-bottom: 1px solid #f1f5f9;
  }

  .action-list {
    padding: 8px 0;
  }

  .action-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 14px 16px;
    border-bottom: 1px solid #f5f7fa;
    transition: background-color 0.2s ease;
  }

  .action-row:last-child {
    border-bottom: 0;
  }

  .action-row:hover {
    background: #fafcff;
  }

  .action-row.compact {
    padding-top: 10px;
    padding-bottom: 10px;
  }

  .action-main {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
  }

  .action-title-line {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .action-title {
    line-height: 1.4;
  }

  .action-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .action-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .action-side {
    display: flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
  }

  .decision-tag {
    cursor: pointer;
  }

  @media (max-width: 960px) {
    .workbench-toolbar,
    .workbench-summary {
      flex-direction: column;
      align-items: stretch;
    }

    .toolbar-main {
      grid-template-columns: 1fr;
    }

    .toolbar-extra {
      justify-content: flex-end;
    }

    .action-row {
      flex-direction: column;
      align-items: stretch;
    }

    .action-side {
      justify-content: flex-end;
    }
  }
</style>
