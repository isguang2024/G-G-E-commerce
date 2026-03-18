<template>
  <ElDialog
    v-model="visible"
    :title="`团队功能权限 - ${teamName}`"
    width="920px"
    destroy-on-close
  >
    <div v-loading="loading" class="team-action-dialog">
      <ElAlert
        type="info"
        :closable="false"
        class="dialog-alert"
        title="这里配置的是团队可开通的功能边界。团队未开通的功能，团队成员和团队管理员都不能使用。"
      />

      <section class="control-panel">
        <div class="toolbar">
          <ElInput
            v-model="filters.keyword"
            clearable
            placeholder="搜索权限名称/权限键/模块归属/分类"
            class="toolbar-input"
          />
          <ElSelect v-model="filters.featureKind" clearable placeholder="功能归属" class="toolbar-select">
            <ElOption label="全部归属" value="" />
            <ElOption label="系统功能" value="system" />
            <ElOption label="业务功能" value="business" />
          </ElSelect>
          <ElSelect v-model="filters.enabledState" clearable placeholder="开通状态" class="toolbar-select">
            <ElOption label="全部状态" value="" />
            <ElOption label="仅已开通" value="enabled" />
            <ElOption label="仅未开通" value="disabled" />
          </ElSelect>
        </div>

        <div class="option-row">
          <div class="summary">
            <ElTag effect="plain" class="summary-tag">可见 {{ filteredActionCount }} / {{ actions.length }}</ElTag>
            <ElTag effect="plain" class="summary-tag summary-tag--allow">已开通 {{ enabledCount }}</ElTag>
            <ElTag effect="plain" class="summary-tag summary-tag--deny">未开通 {{ disabledCount }}</ElTag>
          </div>

          <div class="option-switches">
            <label class="option-item">
              <span>仅已开通</span>
              <ElSwitch v-model="filters.onlyEnabled" />
            </label>
            <label class="option-item">
              <span>显示 ID/说明</span>
              <ElSwitch v-model="filters.showRemark" />
            </label>
            <label class="option-item">
              <span>紧凑模式</span>
              <ElSwitch v-model="filters.compact" />
            </label>
          </div>
        </div>

        <div class="batch-bar" v-if="filteredActionIds.length > 0">
          <span class="batch-label">批量处理当前筛选结果</span>
          <ElButton size="small" text @click="expandAll">全部展开</ElButton>
          <ElButton size="small" text @click="collapseAll">全部收起</ElButton>
          <ElButton size="small" text @click="applyEnabled(filteredActionIds, true)">全部开通</ElButton>
          <ElButton size="small" text @click="applyEnabled(filteredActionIds, false)">全部关闭</ElButton>
        </div>
      </section>

      <ActionPermissionTreePanel
        ref="treePanelRef"
        :loading="loading"
        :tree-data="treeData"
        empty-description="暂无匹配的功能权限"
      >
        <template #node="{ data }">
            <div v-if="data.nodeType === 'feature'" class="tree-node tree-node--feature" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions">
                <ElButton size="small" text @click.stop="applyEnabled(data.actionIds, true)">本组开通</ElButton>
                <ElButton size="small" text @click.stop="applyEnabled(data.actionIds, false)">本组关闭</ElButton>
              </div>
            </div>

            <div v-else-if="data.nodeType === 'module'" class="tree-node tree-node--module" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions">
                <ElButton size="small" text @click.stop="applyEnabled(data.actionIds, true)">模块开通</ElButton>
                <ElButton size="small" text @click.stop="applyEnabled(data.actionIds, false)">模块关闭</ElButton>
              </div>
            </div>

            <div v-else class="tree-node tree-node--action" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div v-if="data.meta" class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions node-actions--leaf">
                <ElTag size="small" effect="plain" class="muted-tag">{{ data.scopeText }}</ElTag>
                <ElTag
                  size="small"
                  effect="plain"
                  class="muted-tag"
                  :class="data.enabled ? 'muted-tag--allow' : 'muted-tag--deny'"
                >
                  {{ data.enabled ? '已开通' : '未开通' }}
                </ElTag>
                <ElSwitch v-model="selectedMap[data.actionId]" />
              </div>
            </div>
        </template>
      </ActionPermissionTreePanel>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    buildAllExpandedKeys,
    buildDefaultExpandedKeys,
    buildPermissionGroups,
    buildPermissionTree,
    type FeatureGroup
  } from '@/components/business/permission/permission-tree'
  import ActionPermissionTreePanel from '@/components/business/permission/action-permission-tree-panel.vue'
  import { fetchGetPermissionActionList } from '@/api/system-manage'
  import { fetchGetTeamActions, fetchSetTeamActions } from '@/api/team'
  import { formatScopeLabel } from '@/utils/permission/scope'
  import { ElMessage } from 'element-plus'

  interface Props {
    modelValue: boolean
    teamId: string
    teamName: string
  }

  type PermissionActionItem = Api.SystemManage.PermissionActionItem
  interface PermissionTreeNode {
    key: string
    label: string
    nodeType: 'feature' | 'module' | 'action'
    meta: string
    children?: PermissionTreeNode[]
    actionIds: string[]
    actionId?: string
    scopeText?: string
    enabled?: boolean
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const submitting = ref(false)
  const actions = ref<PermissionActionItem[]>([])
  const treePanelRef = ref<InstanceType<typeof ActionPermissionTreePanel>>()
  const expandedKeys = ref<string[]>([])
  const selectedMap = reactive<Record<string, boolean>>({})
  const filters = reactive({
    keyword: '',
    featureKind: '',
    enabledState: '',
    onlyEnabled: false,
    showRemark: false,
    compact: false
  })

  const filteredActions = computed(() => {
    const keyword = filters.keyword.trim().toLowerCase()
    return actions.value.filter((item) => {
      if (filters.featureKind && item.featureKind !== filters.featureKind) {
        return false
      }

      const enabled = Boolean(selectedMap[item.id])
      if (filters.onlyEnabled && !enabled) {
        return false
      }

      switch (filters.enabledState) {
        case 'enabled':
          if (!enabled) return false
          break
        case 'disabled':
          if (enabled) return false
          break
      }

      if (!keyword) return true

      const haystack = [
        item.name,
        item.permissionKey,
        item.moduleCode,
        item.category,
        item.resourceCode,
        item.actionCode,
        item.description
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()

      return haystack.includes(keyword)
    })
  })

  const filteredActionIds = computed(() => filteredActions.value.map((item) => item.id))
  const filteredActionCount = computed(() => filteredActions.value.length)
  const enabledCount = computed(() => Object.values(selectedMap).filter(Boolean).length)
  const disabledCount = computed(() => actions.value.length - enabledCount.value)

  const filteredGroups = computed<FeatureGroup[]>(() => buildPermissionGroups(filteredActions.value))

  const treeData = computed<PermissionTreeNode[]>(() =>
    buildPermissionTree(filteredGroups.value, (action) => ({
      meta: buildActionMeta(action),
      scopeText: formatScopeLabel(action.scopeCode, action.scopeName),
      enabled: Boolean(selectedMap[action.id])
    })) as PermissionTreeNode[]
  )

  function buildActionMeta(action: PermissionActionItem) {
    if (!filters.showRemark) {
      return ''
    }
    const parts = [action.permissionKey || `${action.resourceCode}:${action.actionCode}`]
    if (action.description) {
      parts.push(action.description)
    }
    return parts.join('  ')
  }

  function applyEnabled(actionIds: string[], enabled: boolean) {
    actionIds.forEach((id) => {
      selectedMap[id] = enabled
    })
  }

  function syncExpandedKeys() {
    const nextKeys = buildDefaultExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    nextTick(() => {
      treePanelRef.value?.setExpandedKeys(nextKeys)
    })
  }

  function expandAll() {
    const nextKeys = buildAllExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    treePanelRef.value?.expandAll()
  }

  function collapseAll() {
    expandedKeys.value = []
    treePanelRef.value?.collapseAll()
  }

  async function loadData() {
    if (!props.teamId) return
    loading.value = true
    try {
      const [actionList, teamActions] = await Promise.all([
        fetchGetPermissionActionList({ current: 1, size: 1000, scopeCode: 'team' }),
        fetchGetTeamActions(props.teamId)
      ])
      actions.value = actionList.records || []
      Object.keys(selectedMap).forEach((key) => delete selectedMap[key])
      actions.value.forEach((action) => {
        selectedMap[action.id] = teamActions.actionIds.includes(action.id)
      })

      Object.assign(filters, {
        keyword: '',
        featureKind: '',
        enabledState: '',
        onlyEnabled: false,
        showRemark: false,
        compact: false
      })
      syncExpandedKeys()
    } catch (e: any) {
      ElMessage.error(e?.message || '获取团队功能权限失败')
      visible.value = false
    } finally {
      loading.value = false
    }
  }

  watch(
    () => [visible.value, props.teamId],
    ([opened]) => {
      if (opened) loadData()
    }
  )

  watch(filteredGroups, () => {
    syncExpandedKeys()
  })

  async function handleSubmit() {
    submitting.value = true
    try {
      const actionIds = Object.keys(selectedMap).filter((id) => selectedMap[id])
      await fetchSetTeamActions(props.teamId, actionIds)
      ElMessage.success('保存成功')
      emit('success')
      visible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped>
  .team-action-dialog {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .dialog-alert {
    margin-bottom: 0;
  }

  .control-panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 14px;
    border: 1px solid #e7edf5;
    border-radius: 16px;
    background:
      linear-gradient(180deg, rgba(248, 251, 255, 0.92) 0%, rgba(255, 255, 255, 0.98) 100%);
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.9),
      0 8px 24px rgba(15, 23, 42, 0.04);
  }

  .toolbar {
    display: grid;
    grid-template-columns: minmax(320px, 1fr) 150px 150px;
    gap: 12px;
    align-items: center;
  }

  .toolbar-input,
  .toolbar-select {
    width: 100%;
  }

  .option-row,
  .summary,
  .batch-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
  }

  .option-row {
    justify-content: space-between;
    gap: 12px;
  }

  .summary,
  .batch-bar {
    gap: 8px;
  }

  .summary-tag {
    height: 28px;
    padding: 0 10px;
    font-size: 12px;
    color: #5f6b7a;
    border-color: #d7dde6;
    background: #f8fafc;
    border-radius: 999px;
  }

  .summary-tag--allow {
    color: #316b56;
    border-color: #cde5d9;
    background: #f2fbf6;
  }

  .summary-tag--deny {
    color: #925f64;
    border-color: #efd6da;
    background: #fff6f7;
  }

  .option-switches {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 10px;
  }

  .option-item {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 34px;
    padding: 0 12px;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.88);
    color: #526075;
    font-size: 12px;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
  }

  .batch-label {
    font-size: 12px;
    color: #69778a;
  }

  .tree-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    min-height: 34px;
    padding: 6px 8px;
    border-radius: 12px;
  }

  .tree-node--feature {
    background: rgba(246, 249, 253, 0.92);
    border: 1px solid #e5ebf3;
  }

  .tree-node--module {
    background: rgba(255, 255, 255, 0.9);
    border: 1px solid #edf2f7;
  }

  .node-main {
    min-width: 0;
    flex: 1;
  }

  .node-title {
    font-size: 12px;
    font-weight: 600;
    color: #243144;
    letter-spacing: 0.01em;
  }

  .node-subtitle {
    font-size: 11px;
    line-height: 1.45;
    color: #7b889c;
    margin-top: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .node-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    justify-content: flex-end;
  }

  .muted-tag {
    font-size: 11px;
    color: #6b7280;
    border-color: #dde3ea;
    background: #f8fafc;
    border-radius: 999px;
  }

  .muted-tag--allow {
    color: #1f6a4d;
    border-color: #bfe6ce;
    background: #eefaf3;
  }

  .muted-tag--deny {
    color: #9b4956;
    border-color: #efc8cf;
    background: #fff3f4;
  }

  .is-compact {
    padding-top: 4px;
    padding-bottom: 4px;
  }

  .is-compact .node-subtitle {
    display: none;
  }

  @media (max-width: 900px) {
    .toolbar {
      grid-template-columns: 1fr;
    }

    .option-row {
      flex-direction: column;
      align-items: flex-start;
    }

    .option-switches {
      justify-content: flex-start;
    }

    .tree-node {
      flex-direction: column;
      align-items: flex-start;
    }

    .node-actions {
      justify-content: flex-start;
    }
  }
</style>
