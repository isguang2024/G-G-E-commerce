<template>
  <ElDialog
    v-model="visible"
    :title="`用户功能权限 - ${userData?.nickName || userData?.userName || ''}`"
    width="920px"
    destroy-on-close
  >
    <div v-loading="loading" class="user-action-dialog">
      <p class="dialog-note">
        配置个人功能权限。默认继承角色，仅在例外场景下使用单独配置。
      </p>

      <section class="control-panel">
        <div class="toolbar">
          <ElInput
            v-model="filters.keyword"
            clearable
            placeholder="搜索权限名称/权限键/模块归属/分类"
            class="toolbar-input"
          >
            <template #prefix>
              <ElIcon class="toolbar-icon"><Search /></ElIcon>
            </template>
          </ElInput>
          <ElSelect v-model="filters.featureKind" clearable placeholder="功能归属" class="toolbar-select">
            <ElOption label="全部归属" value="" />
            <ElOption label="系统功能" value="system" />
            <ElOption label="业务功能" value="business" />
          </ElSelect>
          <ElSelect v-model="filters.overrideState" clearable placeholder="覆盖状态" class="toolbar-select">
            <ElOption label="全部状态" value="" />
            <ElOption label="仅继承角色" value="inherit" />
            <ElOption label="仅单独允许" value="allow" />
            <ElOption label="仅单独拒绝" value="deny" />
            <ElOption label="仅已覆盖" value="overridden" />
          </ElSelect>
        </div>

        <div class="option-row">
          <div class="summary">
            <span class="summary-pill">全部 <em>{{ actions.length }}</em></span>
            <span class="summary-pill">继承 <em>{{ inheritCount }}</em></span>
            <span class="summary-pill summary-pill--allow">单独允许 <em>{{ overrideAllowCount }}</em></span>
            <span class="summary-pill summary-pill--deny">单独拒绝 <em>{{ overrideDenyCount }}</em></span>
          </div>

          <div class="option-switches">
            <label class="option-item">
              <span>仅看例外</span>
              <ElSwitch v-model="filters.onlyOverrides" size="small" />
            </label>
            <label class="option-item">
              <span>显示 ID/说明</span>
              <ElSwitch v-model="filters.showRemark" size="small" />
            </label>
            <label class="option-item">
              <span>紧凑模式</span>
              <ElSwitch v-model="filters.compact" size="small" />
            </label>
          </div>
        </div>

        <div class="batch-bar" v-if="filteredActionIds.length > 0">
          <span class="batch-label">批量处理当前筛选结果</span>
          <div class="batch-links">
            <button type="button" class="text-link" @click="expandAll">全部展开</button>
            <button type="button" class="text-link" @click="collapseAll">全部收起</button>
            <ElDropdown trigger="click" @command="handleBatchCommand">
              <button type="button" class="text-link">批量操作</button>
              <template #dropdown>
                <ElDropdownMenu>
                  <ElDropdownItem command="">继承</ElDropdownItem>
                  <ElDropdownItem command="allow">单独允许</ElDropdownItem>
                  <ElDropdownItem command="deny">单独拒绝</ElDropdownItem>
                </ElDropdownMenu>
              </template>
            </ElDropdown>
          </div>
        </div>
      </section>

      <ActionPermissionTreePanel
        ref="treePanelRef"
        :loading="loading"
        :tree-data="treeData"
        empty-description="暂无匹配的平台级功能权限"
      >
        <template #node="{ data }">
            <div v-if="data.nodeType === 'feature'" class="tree-node tree-node--feature" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
            </div>

            <div v-else-if="data.nodeType === 'module'" class="tree-node tree-node--module" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
            </div>

            <div v-else class="tree-node tree-node--action" :class="{ 'is-compact': filters.compact }">
              <ElDropdown trigger="click" @command="(command) => setEffect(data.actionId, command as EffectValue)">
                <button
                  type="button"
                  class="status-button"
                  :class="statusButtonClass(effectMap[data.actionId])"
                >
                  <ElIcon>
                    <component :is="statusIcon(effectMap[data.actionId])" />
                  </ElIcon>
                </button>
                <template #dropdown>
                  <ElDropdownMenu>
                    <ElDropdownItem command="">继承</ElDropdownItem>
                    <ElDropdownItem command="allow">单独允许</ElDropdownItem>
                    <ElDropdownItem command="deny">单独拒绝</ElDropdownItem>
                  </ElDropdownMenu>
                </template>
              </ElDropdown>
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div v-if="data.meta" class="node-subtitle">{{ data.meta }}</div>
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
  import {
    fetchGetPermissionActionList,
    fetchGetUserActions,
    fetchSetUserActions
  } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'
  import { Check, Close, RefreshRight, Search } from '@element-plus/icons-vue'

  interface Props {
    modelValue: boolean
    userData?: Api.SystemManage.UserListItem
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
  const effectMap = reactive<Record<string, EffectValue>>({})
  const filters = reactive({
    keyword: '',
    featureKind: '',
    overrideState: '',
    onlyOverrides: false,
    showRemark: false,
    compact: false
  })

  const filteredActions = computed(() => {
    const keyword = filters.keyword.trim().toLowerCase()
    return actions.value.filter((item) => {
      if (filters.featureKind && item.featureKind !== filters.featureKind) {
        return false
      }

      const overrideEffect = effectMap[item.id] || ''
      if (filters.onlyOverrides && !overrideEffect) {
        return false
      }

      switch (filters.overrideState) {
        case 'inherit':
          if (overrideEffect) return false
          break
        case 'allow':
          if (overrideEffect !== 'allow') return false
          break
        case 'deny':
          if (overrideEffect !== 'deny') return false
          break
        case 'overridden':
          if (!overrideEffect) return false
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
  const inheritCount = computed(() => actions.value.length - overrideAllowCount.value - overrideDenyCount.value)
  const overrideAllowCount = computed(() => Object.values(effectMap).filter((item) => item === 'allow').length)
  const overrideDenyCount = computed(() => Object.values(effectMap).filter((item) => item === 'deny').length)

  const filteredGroups = computed<FeatureGroup[]>(() => buildPermissionGroups(filteredActions.value))

  const treeData = computed<PermissionTreeNode[]>(() =>
    buildPermissionTree(filteredGroups.value, (action) => ({
      meta: buildActionMeta(action)
    })) as PermissionTreeNode[]
  )

  function buildActionMeta(action: PermissionActionItem) {
    if (!filters.showRemark) return ''
    return [action.permissionKey || `${action.resourceCode}:${action.actionCode}`, action.description]
      .filter(Boolean)
      .join('  ')
  }

  function applyEffects(actionIds: string[], effect: EffectValue) {
    actionIds.forEach((id) => {
      effectMap[id] = effect
    })
  }

  function setEffect(actionId: string, effect: EffectValue) {
    effectMap[actionId] = effect
  }

  function handleBatchCommand(command: string | number | object) {
    applyEffects(filteredActionIds.value, `${command || ''}` as EffectValue)
  }

  function statusIcon(effect: EffectValue) {
    if (effect === 'allow') return Check
    if (effect === 'deny') return Close
    return RefreshRight
  }

  function statusButtonClass(effect: EffectValue) {
    if (effect === 'allow') return 'status-button--allow'
    if (effect === 'deny') return 'status-button--deny'
    return 'status-button--inherit'
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
    if (!props.userData?.id) return
    loading.value = true
    try {
      const [actionList, userActions] = await Promise.all([
        fetchGetPermissionActionList({ current: 1, size: 1000, scopeCode: 'global' }),
        fetchGetUserActions(props.userData.id)
      ])

      actions.value = (actionList.records || []).filter((item) => !item.requiresTenantContext)

      Object.keys(effectMap).forEach((key) => delete effectMap[key])
      actions.value.forEach((item) => {
        effectMap[item.id] = ''
      })
      userActions.forEach((item) => {
        effectMap[item.actionId] = item.effect
      })

      Object.assign(filters, {
        keyword: '',
        featureKind: '',
        overrideState: '',
        onlyOverrides: false,
        showRemark: false,
        compact: false
      })
      syncExpandedKeys()
    } catch (e: any) {
      ElMessage.error(e?.message || '获取用户功能权限失败')
      visible.value = false
    } finally {
      loading.value = false
    }
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
    if (!props.userData?.id) return
    submitting.value = true
    try {
      const payload = Object.entries(effectMap)
        .filter(([, effect]) => effect)
        .map(([actionId, effect]) => ({
          action_id: actionId,
          effect: effect as 'allow' | 'deny'
        }))
      await fetchSetUserActions(props.userData.id, payload)
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
  .user-action-dialog {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .dialog-note {
    margin: 0;
    padding: 0 4px;
    color: #8a94a6;
    font-size: 12px;
    line-height: 1.6;
  }

  .control-panel {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 10px 2px 2px;
    background: transparent;
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

  .toolbar-icon {
    color: #8fb2ff;
  }

  :deep(.toolbar-input .el-input__wrapper),
  :deep(.toolbar-select .el-select__wrapper) {
    min-height: 42px;
    border-radius: 14px;
    box-shadow: 0 0 0 1px #e6edf6 inset;
    background: #fbfdff;
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
    gap: 10px;
  }

  .summary-pill {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    min-height: 28px;
    padding: 0 10px;
    font-size: 12px;
    color: #637289;
    background: #f5f8fc;
    border-radius: 999px;
  }

  .summary-pill em {
    color: #6d90ff;
    font-style: normal;
    font-weight: 600;
  }

  .summary-pill--allow {
    color: #316b56;
    background: #f2fbf6;
  }

  .summary-pill--allow em {
    color: #3cad73;
  }

  .summary-pill--deny {
    color: #925f64;
    background: #fff6f7;
  }

  .summary-pill--deny em {
    color: #e36c78;
  }

  .option-switches {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 12px;
  }

  .option-item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-height: 24px;
    padding: 0;
    color: #7a8799;
    font-size: 12px;
  }

  .batch-label {
    font-size: 12px;
    color: #69778a;
  }

  .batch-links {
    display: inline-flex;
    align-items: center;
    gap: 18px;
  }

  .text-link {
    padding: 0;
    border: 0;
    background: transparent;
    color: #6c8fff;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
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
    margin-top: 4px;
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

  .status-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    padding: 0;
    border: 0;
    border-radius: 999px;
    color: #fff;
    cursor: pointer;
  }

  .status-button--inherit {
    background: #d8dee8;
    color: #768398;
  }

  .status-button--allow {
    background: #55c78a;
  }

  .status-button--deny {
    background: #ee7b86;
  }

  .tree-node--action {
    border-bottom: 1px solid rgba(229, 235, 243, 0.72);
  }

  .tree-node--action:last-child {
    border-bottom: 0;
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
