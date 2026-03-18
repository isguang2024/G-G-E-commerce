<template>
  <ElDialog
    :model-value="visible"
    @update:model-value="handleClose"
    title="功能门槛"
    width="960px"
    destroy-on-close
    append-to-body
    class="menu-action-requirement-dialog"
  >
    <div class="menu-action-requirement-dialog__body" v-loading="loading">
      <section class="menu-action-requirement-dialog__panel">
        <div class="menu-action-requirement-dialog__header">
          <div>
            <div class="menu-action-requirement-dialog__title">
              {{ menuData?.meta?.title || menuData?.name || '当前菜单' }}
            </div>
            <div class="menu-action-requirement-dialog__subtitle">
              为菜单绑定基础功能权限，控制入口显示与访问策略
            </div>
          </div>
          <div class="menu-action-requirement-dialog__summary-tags">
            <ElTag effect="plain">可见 {{ visibleActionCount }} / {{ permissionActions.length }}</ElTag>
            <ElTag effect="plain" type="primary">已选 {{ form.requiredActions.length }}</ElTag>
          </div>
        </div>

        <div class="menu-action-requirement-dialog__toolbar">
          <ElInput
            v-model="filters.keyword"
            clearable
            placeholder="搜索权限名称/权限键/模块归属/分类"
            class="menu-action-requirement-dialog__input"
          />
          <ElSelect
            v-model="filters.source"
            clearable
            placeholder="来源"
            class="menu-action-requirement-dialog__select"
          >
            <ElOption label="全部来源" value="" />
            <ElOption label="接口自动" value="api" />
            <ElOption label="系统内置" value="system" />
            <ElOption label="业务定义" value="business" />
          </ElSelect>
          <ElSelect
            v-model="filters.featureKind"
            clearable
            placeholder="功能归属"
            class="menu-action-requirement-dialog__select"
          >
            <ElOption label="全部归属" value="" />
            <ElOption label="系统功能" value="system" />
            <ElOption label="业务功能" value="business" />
          </ElSelect>
          <ElSelect
            v-model="filters.selectionState"
            placeholder="选择状态"
            class="menu-action-requirement-dialog__select"
          >
            <ElOption label="全部" value="" />
            <ElOption label="仅已选择" value="selected" />
            <ElOption label="仅未选择" value="unselected" />
          </ElSelect>
        </div>

        <div class="menu-action-requirement-dialog__options">
          <div class="menu-action-requirement-dialog__option-block">
            <span class="menu-action-requirement-dialog__option-label">匹配方式</span>
            <ElRadioGroup v-model="form.actionMatchMode" size="small">
              <ElRadioButton value="any" label="any">任意满足</ElRadioButton>
              <ElRadioButton value="all" label="all">全部满足</ElRadioButton>
            </ElRadioGroup>
          </div>

          <div class="menu-action-requirement-dialog__option-block">
            <span class="menu-action-requirement-dialog__option-label">权限不满足时</span>
            <ElRadioGroup v-model="form.actionVisibilityMode" size="small">
              <ElRadioButton value="hide" label="hide">不显示</ElRadioButton>
              <ElRadioButton value="show" label="show">显示</ElRadioButton>
            </ElRadioGroup>
          </div>

          <div class="menu-action-requirement-dialog__option-actions">
            <ElButton size="small" text @click="expandAll">全部展开</ElButton>
            <ElButton size="small" text @click="collapseAll">全部收起</ElButton>
            <ElButton size="small" text @click="clearSelection">清空已选</ElButton>
          </div>
        </div>
      </section>

      <ElEmpty v-if="!loading && treeData.length === 0" description="暂无匹配的功能权限" />

      <div v-else class="menu-action-requirement-dialog__tree-wrapper">
        <ElTree
          ref="treeRef"
          :data="treeData"
          node-key="key"
          show-checkbox
          :props="treeProps"
          :default-expanded-keys="expandedKeys"
          :expand-on-click-node="true"
          class="menu-action-requirement-dialog__tree"
          @check="handleTreeCheck"
        >
          <template #default="{ data }">
            <div v-if="data.nodeType === 'feature'" class="selector-node selector-node--feature">
              <div class="selector-node__title">{{ data.label }}</div>
              <div class="selector-node__meta">{{ data.meta }}</div>
            </div>
            <div v-else-if="data.nodeType === 'module'" class="selector-node selector-node--module">
              <div class="selector-node__title">{{ data.label }}</div>
              <div class="selector-node__meta">{{ data.meta }}</div>
            </div>
            <div v-else class="selector-node selector-node--action">
              <div class="selector-node__content">
                <div class="selector-node__title">{{ data.label }}</div>
                <div class="selector-node__meta">{{ data.meta }}</div>
              </div>
              <div class="selector-node__tags">
                <ElTag size="small" effect="plain">{{ data.scopeText }}</ElTag>
                <ElTag size="small" effect="plain">{{ data.sourceText }}</ElTag>
              </div>
            </div>
          </template>
        </ElTree>
      </div>
    </div>

    <template #footer>
      <ElButton @click="handleClose(false)">取消</ElButton>
      <ElButton type="primary" @click="handleConfirm">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    buildAllExpandedKeys,
    buildDefaultExpandedKeys,
    buildPermissionGroups,
    type FeatureGroup
  } from '@/components/business/permission/permission-tree'
  import { fetchGetPermissionActionList } from '@/api/system-manage'
  import { buildScopedActionKey, resolveActionKey } from '@/utils/permission/action'
  import { formatScopeLabel } from '@/utils/permission/scope'

  interface Props {
    modelValue: boolean
    menuData?: any | null
  }

  interface FormState {
    requiredActions: string[]
    actionMatchMode: 'any' | 'all'
    actionVisibilityMode: 'hide' | 'show'
  }

  interface PermissionSelectorTreeNode {
    key: string
    label: string
    nodeType: 'feature' | 'module' | 'action'
    meta: string
    children?: PermissionSelectorTreeNode[]
    actionIds: string[]
    actionValue?: string
    scopeText?: string
    sourceText?: string
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'submit', value: FormState): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const treeRef = ref()
  const expandedKeys = ref<string[]>([])
  const treeProps = {
    children: 'children',
    label: 'label'
  }
  const sourceTextMap: Record<string, string> = {
    api: '接口自动',
    system: '系统内置',
    business: '业务定义'
  }

  const filters = reactive({
    keyword: '',
    source: '',
    featureKind: '',
    selectionState: ''
  })

  const form = reactive<FormState>({
    requiredActions: [],
    actionMatchMode: 'any',
    actionVisibilityMode: 'hide'
  })

  const visibleActionKeySet = computed(() => {
    return new Set(
      filteredActions.value.map((item) =>
        buildScopedActionKey(`${item.resourceCode}:${item.actionCode}`, item.scopeCode || item.scope)
      )
    )
  })

  const filteredActions = computed(() => {
    const keyword = filters.keyword.trim().toLowerCase()
    return permissionActions.value.filter((item) => {
      if (filters.source && item.source !== filters.source) {
        return false
      }
      if (filters.featureKind && item.featureKind !== filters.featureKind) {
        return false
      }
      const actionKey = buildScopedActionKey(
        `${item.resourceCode}:${item.actionCode}`,
        item.scopeCode || item.scope
      )
      if (filters.selectionState === 'selected' && !form.requiredActions.includes(actionKey)) {
        return false
      }
      if (filters.selectionState === 'unselected' && form.requiredActions.includes(actionKey)) {
        return false
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

  const visibleActionCount = computed(() => filteredActions.value.length)
  const groups = computed<FeatureGroup[]>(() => buildPermissionGroups(filteredActions.value))

  const treeData = computed<PermissionSelectorTreeNode[]>(() =>
    groups.value.map((featureGroup) => ({
      key: featureGroup.key,
      label: featureGroup.label,
      nodeType: 'feature',
      meta: `${featureGroup.modules.length} 个模块，${featureGroup.count} 条权限`,
      actionIds: featureGroup.actionIds.map((id) => id),
      children: featureGroup.modules.map((moduleGroup) => ({
        key: moduleGroup.key,
        label: moduleGroup.label,
        nodeType: 'module',
        meta: `${moduleGroup.count} 条权限${moduleGroup.category ? `，分类 ${moduleGroup.category}` : ''}`,
        actionIds: moduleGroup.actions.map((action) =>
          buildScopedActionKey(`${action.resourceCode}:${action.actionCode}`, action.scopeCode || action.scope)
        ),
        children: moduleGroup.actions.map((action) => {
          const actionValue = buildScopedActionKey(
            `${action.resourceCode}:${action.actionCode}`,
            action.scopeCode || action.scope
          )
          return {
            key: actionValue,
            label: action.name,
            nodeType: 'action',
            meta: [
              action.permissionKey || `${action.resourceCode}:${action.actionCode}`,
              action.description
            ]
              .filter(Boolean)
              .join('  '),
            actionIds: [actionValue],
            actionValue,
            scopeText: formatScopeLabel(action.scopeCode, action.scopeName),
            sourceText: sourceTextMap[action.source || 'business'] || '业务定义'
          }
        })
      }))
    }))
  )

  const syncExpandedKeys = (): void => {
    const nextKeys = buildDefaultExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    nextTick(() => {
      treeRef.value?.setExpandedKeys(nextKeys)
      treeRef.value?.setCheckedKeys(form.requiredActions)
    })
  }

  const normalizeRequiredActions = (
    actions: string[],
    availableActions: Api.SystemManage.PermissionActionItem[]
  ): string[] => {
    const scopedKeyMap = new Map<string, string>()
    const unscopedKeyMap = new Map<string, string[]>()

    availableActions.forEach((item) => {
      const scopedKey = buildScopedActionKey(
        `${item.resourceCode}:${item.actionCode}`,
        item.scopeCode || item.scope
      )
      const unscopedKey = resolveActionKey(scopedKey).key
      scopedKeyMap.set(scopedKey, scopedKey)
      const current = unscopedKeyMap.get(unscopedKey) || []
      current.push(scopedKey)
      unscopedKeyMap.set(unscopedKey, current)
    })

    return Array.from(
      new Set(
        actions
          .map((item) => `${item || ''}`.trim())
          .filter(Boolean)
          .map((item) => {
            if (scopedKeyMap.has(item)) {
              return item
            }

            const raw = resolveActionKey(item)
            const candidates = unscopedKeyMap.get(raw.key) || []
            if (!candidates.length) {
              return item
            }

            if (raw.scope) {
              const normalizedScopedKey = buildScopedActionKey(raw.key, raw.scope)
              const exactCandidate = candidates.find((candidate) => candidate === normalizedScopedKey)
              if (exactCandidate) {
                return exactCandidate
              }
            }

            const globalCandidate = candidates.find((candidate) => candidate.endsWith('@global'))
            if (globalCandidate) {
              return globalCandidate
            }

            return candidates[0]
          })
      )
    )
  }

  const loadPermissionActions = async (): Promise<void> => {
    if (loading.value) return
    loading.value = true
    try {
      const res = await fetchGetPermissionActionList({ current: 1, size: 1000, status: 'normal' })
      permissionActions.value = res?.records || []
      form.requiredActions = normalizeRequiredActions(form.requiredActions, permissionActions.value)
      syncExpandedKeys()
    } finally {
      loading.value = false
    }
  }

  const initFromMenuData = (): void => {
    const meta = props.menuData?.meta || {}
    const initialActions = Array.from(
      new Set(
        [meta.requiredAction, ...(Array.isArray(meta.requiredActions) ? meta.requiredActions : [])]
          .map((item: string) => `${item || ''}`.trim())
          .filter(Boolean)
      )
    )
    form.requiredActions = normalizeRequiredActions(initialActions, permissionActions.value)
    form.actionMatchMode = meta.actionMatchMode === 'all' ? 'all' : 'any'
    form.actionVisibilityMode = meta.actionVisibilityMode === 'show' ? 'show' : 'hide'
    filters.keyword = ''
    filters.source = ''
    filters.featureKind = ''
    filters.selectionState = ''
  }

  const handleTreeCheck = (): void => {
    const visibleCheckedKeys = (treeRef.value?.getCheckedKeys(true) || []) as string[]
    const preservedHiddenKeys = form.requiredActions.filter((item) => !visibleActionKeySet.value.has(item))
    form.requiredActions = Array.from(new Set([...preservedHiddenKeys, ...visibleCheckedKeys]))
  }

  const expandAll = (): void => {
    const nextKeys = buildAllExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    treeRef.value?.setExpandedKeys(nextKeys)
  }

  const collapseAll = (): void => {
    expandedKeys.value = []
    treeRef.value?.setExpandedKeys([])
  }

  const clearSelection = (): void => {
    form.requiredActions = []
    treeRef.value?.setCheckedKeys([])
  }

  const handleClose = (value = false): void => {
    visible.value = value
  }

  const handleConfirm = (): void => {
    emit('submit', {
      requiredActions: [...form.requiredActions],
      actionMatchMode: form.actionMatchMode,
      actionVisibilityMode: form.actionVisibilityMode
    })
    visible.value = false
  }

  watch(
    () => visible.value,
    (opened) => {
      if (!opened) return
      initFromMenuData()
      loadPermissionActions()
    }
  )

  watch(
    () => [filters.keyword, filters.source, filters.featureKind, filters.selectionState],
    () => {
      if (!visible.value) return
      syncExpandedKeys()
    }
  )
</script>

<style scoped>
  .menu-action-requirement-dialog__body {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .menu-action-requirement-dialog__panel {
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

  .menu-action-requirement-dialog__header,
  .menu-action-requirement-dialog__summary-tags,
  .menu-action-requirement-dialog__option-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .menu-action-requirement-dialog__header,
  .menu-action-requirement-dialog__options {
    justify-content: space-between;
  }

  .menu-action-requirement-dialog__title {
    font-size: 14px;
    font-weight: 600;
    color: #243144;
  }

  .menu-action-requirement-dialog__subtitle {
    margin-top: 4px;
    font-size: 12px;
    color: #7b889c;
  }

  .menu-action-requirement-dialog__toolbar {
    display: grid;
    grid-template-columns: minmax(320px, 1fr) 150px 150px 150px;
    gap: 12px;
    align-items: center;
  }

  .menu-action-requirement-dialog__input,
  .menu-action-requirement-dialog__select {
    width: 100%;
  }

  .menu-action-requirement-dialog__options {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .menu-action-requirement-dialog__option-block {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .menu-action-requirement-dialog__option-label {
    font-size: 12px;
    color: #69778a;
  }

  .menu-action-requirement-dialog__tree-wrapper {
    border: 1px solid #e5ebf3;
    border-radius: 16px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(249, 251, 254, 0.96) 100%);
    padding: 10px;
  }

  .menu-action-requirement-dialog__tree {
    max-height: 540px;
    overflow: auto;
  }

  .selector-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    min-height: 34px;
    padding: 6px 8px;
    border-radius: 12px;
  }

  .selector-node--feature {
    background: rgba(246, 249, 253, 0.92);
    border: 1px solid #e5ebf3;
  }

  .selector-node--module {
    background: rgba(255, 255, 255, 0.9);
    border: 1px solid #edf2f7;
  }

  .selector-node__content {
    min-width: 0;
    flex: 1;
  }

  .selector-node__title {
    font-size: 12px;
    font-weight: 600;
    color: #243144;
  }

  .selector-node__meta {
    margin-top: 2px;
    font-size: 11px;
    color: #7b889c;
    line-height: 1.45;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .selector-node__tags {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }

  :deep(.menu-action-requirement-dialog__tree .el-tree-node__content) {
    height: auto;
    min-height: 36px;
    margin: 2px 0;
    padding: 0;
    border-radius: 12px;
  }

  @media (max-width: 900px) {
    .menu-action-requirement-dialog__toolbar {
      grid-template-columns: 1fr;
    }
  }
</style>
