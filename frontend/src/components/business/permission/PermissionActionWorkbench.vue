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
                          v-else-if="mode === 'team'"
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
                      <template v-if="mode === 'menu' || mode === 'team'">
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
  import { computed, ref, watch } from 'vue'
  import { ArrowDown, Search } from '@element-plus/icons-vue'

  type WorkbenchMode = 'menu' | 'team' | 'role' | 'user'
  type DecisionValue = '' | 'allow' | 'deny'

  interface WorkbenchActionItem extends Partial<Api.SystemManage.PermissionActionItem> {
    id: string
    name: string
    permissionKey?: string
    code?: string
  }

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

  const searchKeyword = ref('')
  const featureFilter = ref('')
  const stateFilter = ref('all')
  const showMeta = ref(false)
  const compactMode = ref(false)
  const activeFeatures = ref<string[]>([])

  const featureOptions = computed(() => {
    return uniqueByValue(
      props.actions
        .map((item) => ({
          value: `${item.featureGroupId || item.featureKind || ''}`,
          label: item.featureGroup?.name || formatFeature(`${item.featureKind || ''}`)
        }))
        .filter((item) => item.value)
    )
  })

  const moduleKeywords = (item: WorkbenchActionItem) =>
    [
      item.moduleGroup?.name,
      item.moduleCode,
      item.resourceCode,
      item.featureGroup?.name,
      item.featureKind
    ]
      .filter(Boolean)
      .join(' ')

  const filteredActions = computed(() => {
    const keyword = searchKeyword.value.trim().toLowerCase()
    return props.actions.filter((item) => {
      if (
        featureFilter.value &&
        `${item.featureGroupId || item.featureKind || ''}` !== featureFilter.value
      ) {
        return false
      }

      if (keyword) {
        const text = [
          item.name,
          item.permissionKey,
          item.code,
          item.description,
          moduleKeywords(item)
        ]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
        if (!text.includes(keyword)) {
          return false
        }
      }

      if (!matchState(item.id)) {
        return false
      }

      return true
    })
  })

  const groupedActions = computed(() => {
    const featureMap = new Map<
      string,
      {
        key: string
        label: string
        modules: Array<{
          key: string
          label: string
          actions: WorkbenchActionItem[]
          actionCount: number
          selectedCount: number
        }>
        moduleCount: number
        actionCount: number
        selectedCount: number
      }
    >()

    filteredActions.value.forEach((item) => {
      const featureKey = `${item.featureGroupId || item.featureKind || 'business'}`
      const moduleKey = `${item.moduleGroupId || item.moduleCode || item.resourceCode || 'default'}`

      if (!featureMap.has(featureKey)) {
        featureMap.set(featureKey, {
          key: featureKey,
          label: item.featureGroup?.name || formatFeature(featureKey),
          modules: [],
          moduleCount: 0,
          actionCount: 0,
          selectedCount: 0
        })
      }

      const feature = featureMap.get(featureKey)!
      let module = feature.modules.find((entry) => entry.key === moduleKey)
      if (!module) {
        module = {
          key: moduleKey,
          label: formatModule(item),
          actions: [],
          actionCount: 0,
          selectedCount: 0
        }
        feature.modules.push(module)
      }

      module.actions.push(item)
      module.actionCount += 1
      feature.actionCount += 1

      if (isActive(item.id)) {
        module.selectedCount += 1
        feature.selectedCount += 1
      }
    })

    return Array.from(featureMap.values()).map((feature) => ({
      ...feature,
      moduleCount: feature.modules.length
    }))
  })

  const totalCount = computed(() => props.actions.length)
  const visibleCount = computed(() => filteredActions.value.length)
  const selectedCount = computed(() => props.actions.filter((item) => isActive(item.id)).length)

  const selectedLabel = computed(() => {
    if (props.mode === 'user') return '例外'
    if (props.mode === 'role') return '已配置'
    if (props.mode === 'team') return '已开通'
    return '已选'
  })

  const positiveText = computed(() => (props.mode === 'team' ? '已开通' : '已选'))
  const neutralText = computed(() => (props.mode === 'team' ? '未开通' : '未选'))

  const decisionOptions = computed(() => {
    if (props.mode === 'user') {
      return [
        { value: '', label: '继承角色' },
        { value: 'allow', label: '单独允许' },
        { value: 'deny', label: '单独拒绝' }
      ]
    }

    return [
      { value: '', label: '未配置' },
      { value: 'allow', label: '允许' },
      { value: 'deny', label: '拒绝' }
    ]
  })

  const stateOptions = computed(() => {
    if (props.mode === 'menu') {
      return [
        { value: 'all', label: '选择状态：全部' },
        { value: 'selected', label: '选择状态：已选' },
        { value: 'unselected', label: '选择状态：未选' }
      ]
    }

    if (props.mode === 'team') {
      return [
        { value: 'all', label: '开通状态：全部' },
        { value: 'selected', label: '开通状态：已开通' },
        { value: 'unselected', label: '开通状态：未开通' }
      ]
    }

    if (props.mode === 'user') {
      return [
        { value: 'all', label: '覆盖状态：全部' },
        { value: 'inherit', label: '覆盖状态：继承角色' },
        { value: 'allow', label: '覆盖状态：单独允许' },
        { value: 'deny', label: '覆盖状态：单独拒绝' }
      ]
    }

    return [
      { value: 'all', label: '配置状态：全部' },
      { value: 'unset', label: '配置状态：未配置' },
      { value: 'allow', label: '配置状态：允许' },
      { value: 'deny', label: '配置状态：拒绝' }
    ]
  })

  const stateFilterPlaceholder = computed(() => stateOptions.value[0]?.label || '状态')

  const batchCommands = computed(() => {
    if (props.mode === 'menu' || props.mode === 'team') {
      return [
        {
          command: 'select-visible',
          label: props.mode === 'team' ? '批量开通当前结果' : '批量选中当前结果'
        },
        {
          command: 'clear-visible',
          label: props.mode === 'team' ? '批量关闭当前结果' : '批量取消当前结果'
        },
        { command: 'clear-all', label: '清空全部配置' }
      ]
    }

    if (props.mode === 'user') {
      return [
        { command: 'inherit-visible', label: '批量继承角色' },
        { command: 'allow-visible', label: '批量单独允许' },
        { command: 'deny-visible', label: '批量单独拒绝' },
        { command: 'clear-all', label: '清空全部例外' }
      ]
    }

    return [
      { command: 'allow-visible', label: '批量允许当前结果' },
      { command: 'deny-visible', label: '批量拒绝当前结果' },
      { command: 'unset-visible', label: '批量取消当前配置' },
      { command: 'clear-all', label: '清空全部配置' }
    ]
  })

  watch(
    groupedActions,
    (value) => {
      const keys = value.map((item) => item.key)
      if (!activeFeatures.value.length) {
        activeFeatures.value = keys
        return
      }
      activeFeatures.value = activeFeatures.value.filter((item) => keys.includes(item))
      if (!activeFeatures.value.length) {
        activeFeatures.value = keys
      }
    },
    { immediate: true }
  )

  function uniqueByValue<T extends { value: string }>(items: T[]) {
    const seen = new Set<string>()
    return items.filter((item) => {
      if (seen.has(item.value)) return false
      seen.add(item.value)
      return true
    })
  }

  function formatFeature(feature: string) {
    const map: Record<string, string> = {
      system: '系统功能',
      business: '业务功能'
    }
    return map[feature] || feature
  }

  function formatModule(action: WorkbenchActionItem) {
    return action.moduleGroup?.name || action.moduleCode || action.resourceCode || '未分类模块'
  }

  function matchState(actionId: string) {
    if (props.mode === 'menu' || props.mode === 'team') {
      if (stateFilter.value === 'selected') return isSelected(actionId)
      if (stateFilter.value === 'unselected') return !isSelected(actionId)
      return true
    }

    const decision = getDecision(actionId)
    if (props.mode === 'user') {
      if (stateFilter.value === 'inherit') return !decision
      return stateFilter.value === 'all' ? true : decision === stateFilter.value
    }

    if (stateFilter.value === 'unset') return !decision
    return stateFilter.value === 'all' ? true : decision === stateFilter.value
  }

  function isSelected(actionId: string) {
    return props.selectedIds.includes(actionId)
  }

  function isActive(actionId: string) {
    if (props.mode === 'menu' || props.mode === 'team') {
      return isSelected(actionId)
    }
    return Boolean(getDecision(actionId))
  }

  function toggleSelection(actionId: string, value: boolean) {
    const next = new Set(props.selectedIds)
    if (value) {
      next.add(actionId)
    } else {
      next.delete(actionId)
    }
    emit('update:selectedIds', Array.from(next))
  }

  function getDecision(actionId: string): DecisionValue {
    return props.decisionMap[actionId] || ''
  }

  function setDecision(actionId: string, value: string) {
    const next = { ...props.decisionMap }
    const normalized = value === 'allow' || value === 'deny' ? value : ''
    if (!normalized) {
      delete next[actionId]
    } else {
      next[actionId] = normalized
    }
    emit('update:decisionMap', next)
  }

  function getDecisionLabel(actionId: string) {
    const decision = getDecision(actionId)
    if (props.mode === 'user') {
      if (decision === 'allow') return '单独允许'
      if (decision === 'deny') return '单独拒绝'
      return '继承角色'
    }
    if (decision === 'allow') return '允许'
    if (decision === 'deny') return '拒绝'
    return '未配置'
  }

  function getDecisionTagType(actionId: string) {
    const decision = getDecision(actionId)
    if (decision === 'allow') return 'success'
    if (decision === 'deny') return 'danger'
    return 'info'
  }

  function expandAll() {
    activeFeatures.value = groupedActions.value.map((item) => item.key)
  }

  function collapseAll() {
    activeFeatures.value = []
  }

  function handleBatchCommand(command: string) {
    const visibleIds = filteredActions.value.map((item) => item.id)
    if (!visibleIds.length) return

    if (props.mode === 'menu' || props.mode === 'team') {
      const next = new Set(props.selectedIds)
      if (command === 'select-visible') {
        visibleIds.forEach((id) => next.add(id))
      } else if (command === 'clear-visible') {
        visibleIds.forEach((id) => next.delete(id))
      } else if (command === 'clear-all') {
        next.clear()
      }
      emit('update:selectedIds', Array.from(next))
      return
    }

    const next = { ...props.decisionMap }
    if (command === 'clear-all') {
      emit('update:decisionMap', {})
      return
    }

    let decision: DecisionValue = ''
    if (command === 'allow-visible') decision = 'allow'
    if (command === 'deny-visible') decision = 'deny'
    if (command === 'unset-visible' || command === 'inherit-visible') decision = ''

    visibleIds.forEach((id) => {
      if (!decision) {
        delete next[id]
      } else {
        next[id] = decision
      }
    })
    emit('update:decisionMap', next)
  }
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
