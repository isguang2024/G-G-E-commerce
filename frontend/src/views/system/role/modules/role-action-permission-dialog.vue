<template>
  <ElDialog
    v-model="visible"
    :title="`功能权限 - ${roleData?.roleName || ''}`"
    width="920px"
    destroy-on-close
  >
    <div v-loading="loading" class="role-action-dialog">
      <section class="control-panel">
        <div class="toolbar">
          <ElInput
            v-model="filters.keyword"
            clearable
            placeholder="搜索权限名称/权限键/模块归属/分类"
            class="toolbar-input"
          />
          <ElSelect v-model="filters.source" clearable placeholder="来源" class="toolbar-select">
            <ElOption label="全部来源" value="" />
            <ElOption label="接口自动" value="api" />
            <ElOption label="系统内置" value="system" />
            <ElOption label="业务定义" value="business" />
          </ElSelect>
          <ElSelect v-model="filters.featureKind" clearable placeholder="功能归属" class="toolbar-select">
            <ElOption label="全部归属" value="" />
            <ElOption label="系统功能" value="system" />
            <ElOption label="业务功能" value="business" />
          </ElSelect>
        </div>

        <div class="option-row">
          <div class="summary">
            <ElTag effect="plain" class="summary-tag">可见 {{ filteredActionCount }} / {{ actions.length }}</ElTag>
            <ElTag effect="plain" class="summary-tag summary-tag--allow">允许 {{ allowedCount }}</ElTag>
            <ElTag effect="plain" class="summary-tag summary-tag--deny">拒绝 {{ deniedCount }}</ElTag>
          </div>

          <div class="option-switches">
            <label class="option-item option-item--select">
              <span>配置状态</span>
              <ElSelect v-model="filters.effectState" size="small" class="filter-select" placeholder="全部">
                <ElOption label="全部" value="" />
                <ElOption label="未配置" value="empty" />
                <ElOption label="已配置" value="configured" />
                <ElOption label="允许" value="allow" />
                <ElOption label="拒绝" value="deny" />
              </ElSelect>
            </label>
            <label class="option-item">
              <span>查看备注</span>
              <ElSwitch v-model="filters.showRemark" />
            </label>
            <label class="option-item">
              <span>紧凑模式</span>
              <ElSwitch v-model="filters.compact" />
            </label>
          </div>
        </div>

        <div class="batch-bar" v-if="filteredActionIds.length > 0">
          <span class="batch-label">批量操作当前筛选结果</span>
          <ElButton size="small" text @click="expandAll">全部展开</ElButton>
          <ElButton size="small" text @click="collapseAll">全部收起</ElButton>
          <ElButton size="small" text @click="applyEffects(filteredActionIds, 'allow')">全部允许</ElButton>
          <ElButton size="small" text @click="applyEffects(filteredActionIds, 'deny')">全部拒绝</ElButton>
          <ElButton size="small" text @click="applyEffects(filteredActionIds, '')">全部清空</ElButton>
        </div>
      </section>

      <ElEmpty v-if="!loading && filteredGroups.length === 0" description="暂无匹配的功能权限" />

      <div v-else class="tree-wrapper">
        <ElTree
          ref="treeRef"
          :data="treeData"
          node-key="key"
          :props="treeProps"
          :default-expanded-keys="expandedKeys"
          :expand-on-click-node="true"
          :highlight-current="false"
          class="permission-tree"
        >
          <template #default="{ data }">
            <div v-if="data.nodeType === 'feature'" class="tree-node tree-node--feature" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions">
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'allow')">本组允许</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'deny')">本组拒绝</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, '')">本组清空</ElButton>
              </div>
            </div>

            <div v-else-if="data.nodeType === 'module'" class="tree-node tree-node--module" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions">
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'allow')">模块允许</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'deny')">模块拒绝</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, '')">模块清空</ElButton>
              </div>
            </div>

            <div v-else class="tree-node tree-node--action" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div v-if="data.meta" class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions node-actions--leaf">
                <ElTag size="small" effect="plain" class="muted-tag">{{ data.scopeText }}</ElTag>
                <ElTag size="small" effect="plain" class="muted-tag">{{ data.sourceText }}</ElTag>
                <ElSelect
                  v-model="effectMap[data.actionId]"
                  class="effect-select"
                  :class="{
                    'effect-select--allow': effectMap[data.actionId] === 'allow',
                    'effect-select--deny': effectMap[data.actionId] === 'deny',
                    'effect-select--empty': !effectMap[data.actionId]
                  }"
                  size="small"
                  placeholder="未配置"
                >
                  <ElOption label="未配置" value="" />
                  <ElOption label="允许" value="allow" />
                  <ElOption label="拒绝" value="deny" />
                </ElSelect>
              </div>
            </div>
          </template>
        </ElTree>
      </div>
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
  import {
    fetchGetPermissionActionList,
    fetchGetRoleActions,
    fetchSetRoleActions
  } from '@/api/system-manage'
  import { formatScopeLabel } from '@/utils/permission/scope'
  import { ElMessage } from 'element-plus'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
  }

  type EffectValue = 'allow' | 'deny' | ''
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
    sourceText?: string
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
  const treeRef = ref()
  const expandedKeys = ref<string[]>([])
  const effectMap = reactive<Record<string, EffectValue>>({})
  const filters = reactive({
    keyword: '',
    source: '',
    featureKind: '',
    effectState: '',
    showRemark: true,
    compact: false
  })

  const sourceTextMap: Record<string, string> = {
    api: '接口自动',
    system: '系统内置',
    business: '业务定义'
  }
  const treeProps = {
    children: 'children',
    label: 'label'
  }

  const filteredActions = computed(() => {
    const keyword = filters.keyword.trim().toLowerCase()
    return actions.value.filter((item) => {
      if (filters.source && item.source !== filters.source) {
        return false
      }
      if (filters.featureKind && item.featureKind !== filters.featureKind) {
        return false
      }
      const effect = effectMap[item.id] || ''
      switch (filters.effectState) {
        case 'configured':
          if (!effect) return false
          break
        case 'empty':
          if (effect) return false
          break
        case 'allow':
          if (effect !== 'allow') return false
          break
        case 'deny':
          if (effect !== 'deny') return false
          break
      }
      if (!keyword) {
        return true
      }
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
  const allowedCount = computed(() => Object.values(effectMap).filter((item) => item === 'allow').length)
  const deniedCount = computed(() => Object.values(effectMap).filter((item) => item === 'deny').length)

  const filteredGroups = computed<FeatureGroup[]>(() => buildPermissionGroups(filteredActions.value))

  const treeData = computed<PermissionTreeNode[]>(() =>
    buildPermissionTree(filteredGroups.value, (action) => ({
      meta: buildActionMeta(action),
      scopeText: formatScopeLabel(action.scopeCode, action.scopeName),
      sourceText: sourceTextMap[action.source || 'business'] || '业务定义'
    })) as PermissionTreeNode[]
  )

  function syncExpandedKeys() {
    const nextKeys = buildDefaultExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    nextTick(() => {
      if (!treeRef.value) return
      treeRef.value.setExpandedKeys(nextKeys)
    })
  }

  function buildActionMeta(action: PermissionActionItem) {
    const parts = [action.permissionKey || `${action.resourceCode}:${action.actionCode}`]
    if (filters.showRemark && action.description) {
      parts.push(action.description)
    }
    return parts.filter(Boolean).join('  ')
  }

  function expandAll() {
    const nextKeys = buildAllExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    treeRef.value?.setExpandedKeys(nextKeys)
  }

  function collapseAll() {
    expandedKeys.value = []
    treeRef.value?.setExpandedKeys([])
  }

  async function loadData() {
    if (!props.roleData?.roleId) return
    loading.value = true
    try {
      const [actionList, roleActionRes] = await Promise.all([
        fetchGetPermissionActionList({
          current: 1,
          size: 1000,
          scopeCode: props.roleData.scopeCode || props.roleData.scope
        }),
        fetchGetRoleActions(props.roleData.roleId)
      ])
      actions.value = actionList.records || []
      Object.keys(effectMap).forEach((key) => delete effectMap[key])
      for (const action of actions.value) {
        effectMap[action.id] = ''
      }
      for (const item of roleActionRes?.actions || []) {
        effectMap[item.action_id] = item.effect
      }
      Object.assign(filters, {
        keyword: '',
        source: '',
        featureKind: '',
        effectState: '',
        showRemark: true,
        compact: false
      })
      syncExpandedKeys()
    } catch (e: any) {
      ElMessage.error(e?.message || '获取角色功能权限失败')
      visible.value = false
    } finally {
      loading.value = false
    }
  }

  function applyEffects(actionIds: string[], effect: EffectValue) {
    actionIds.forEach((id) => {
      effectMap[id] = effect
    })
  }

  watch(
    () => visible.value,
    (opened) => {
      if (opened) loadData()
    }
  )

  watch(filteredGroups, () => {
    syncExpandedKeys()
  })

  async function handleSubmit() {
    if (!props.roleData?.roleId) return
    submitting.value = true
    try {
      const payload = Object.entries(effectMap)
        .filter(([, effect]) => effect)
        .map(([actionId, effect]) => ({
          action_id: actionId,
          effect: effect as 'allow' | 'deny'
        }))
      await fetchSetRoleActions(props.roleData.roleId, payload)
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
  .role-action-dialog {
    display: flex;
    flex-direction: column;
    gap: 12px;
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

  .option-item--select {
    gap: 10px;
    padding-right: 8px;
  }

  .filter-select {
    width: 104px;
  }

  .batch-label {
    font-size: 12px;
    color: #69778a;
  }

  .tree-wrapper {
    border: 1px solid #e5ebf3;
    border-radius: 18px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(249, 251, 254, 0.96) 100%);
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.8),
      0 10px 28px rgba(15, 23, 42, 0.04);
    padding: 10px;
  }

  .permission-tree {
    max-height: 520px;
    overflow: auto;
    padding-right: 2px;
  }

  :deep(.permission-tree .el-tree-node__content) {
    height: auto;
    min-height: 34px;
    margin: 2px 0;
    padding: 0;
    border-radius: 14px;
  }

  :deep(.permission-tree .el-tree-node__expand-icon) {
    color: #8a94a6;
    font-size: 12px;
  }

  :deep(.permission-tree .el-tree-node__expand-icon.expanded) {
    color: #5d6c86;
  }

  .tree-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    padding: 8px 10px;
    border-radius: 14px;
    transition:
      background-color 0.18s ease,
      box-shadow 0.18s ease,
      transform 0.18s ease;
  }

  .tree-node:hover {
    background: rgba(244, 247, 252, 0.92);
  }

  .tree-node--feature {
    padding: 10px 12px;
    background:
      linear-gradient(135deg, rgba(248, 251, 255, 0.96) 0%, rgba(242, 246, 252, 0.9) 100%);
    border: 1px solid #e6edf6;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
  }

  .tree-node--module {
    padding: 8px 10px;
    background: rgba(255, 255, 255, 0.84);
    border: 1px solid #edf2f7;
  }

  .tree-node--action {
    padding: 6px 8px;
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

  .node-actions--leaf {
    gap: 4px;
  }

  .muted-tag {
    font-size: 11px;
    color: #6b7280;
    border-color: #dde3ea;
    background: #f8fafc;
    border-radius: 999px;
  }

  .effect-select {
    width: 78px;
  }

  :deep(.effect-select .el-input__wrapper) {
    border-radius: 999px;
    box-shadow: 0 0 0 1px #dde5ef inset;
    background: #f8fafc;
    transition:
      box-shadow 0.18s ease,
      background-color 0.18s ease;
  }

  :deep(.effect-select .el-input__inner) {
    font-size: 12px;
    color: #415064;
  }

  :deep(.effect-select--allow .el-input__wrapper) {
    background: #eefaf3;
    box-shadow: 0 0 0 1px #bfe6ce inset;
  }

  :deep(.effect-select--allow .el-input__inner) {
    color: #1f6a4d;
  }

  :deep(.effect-select--deny .el-input__wrapper) {
    background: #fff3f4;
    box-shadow: 0 0 0 1px #efc8cf inset;
  }

  :deep(.effect-select--deny .el-input__inner) {
    color: #9b4956;
  }

  :deep(.effect-select--empty .el-input__wrapper) {
    background: #f8fafc;
    box-shadow: 0 0 0 1px #dde5ef inset;
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
