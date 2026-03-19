<template>
  <ElDialog
    v-model="visible"
    :title="`用户功能权限 - ${userTitle}`"
    width="1120px"
    destroy-on-close
    class="user-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        配置个人功能权限。默认继承角色，仅在例外场景下使用单独配置。勾选后可设置为允许或拒绝，不勾选则继续继承角色。
      </div>

      <ElTabs v-model="activeTab" class="permission-tabs">
        <ElTabPane label="例外配置" name="custom">
          <div class="cascader-pane">
            <div class="cascader-toolbar">
              <div class="cascader-tags">
                <ElTag
                  v-for="item in topLevelActionTags"
                  :key="item.label"
                  effect="plain"
                  round
                >
                  {{ item.label }} {{ item.count }}
                </ElTag>
                <ElTag effect="plain" round>已配置 {{ configuredActionIds.length }}</ElTag>
                <ElTag type="info" effect="plain" round>总计 {{ actionLeafCount }}</ElTag>
              </div>

              <div class="toolbar-filters">
                <ElInput
                  v-model="actionKeyword"
                  clearable
                  placeholder="搜索功能标题或资源动作"
                  :prefix-icon="Search"
                  class="toolbar-search"
                />

                <ElSelect
                  v-model="actionScopeFilter"
                  clearable
                  placeholder="作用域"
                  class="toolbar-scope"
                >
                  <ElOption label="全部作用域" value="" />
                  <ElOption
                    v-for="item in actionScopeOptions"
                    :key="item"
                    :label="item"
                    :value="item"
                  />
                </ElSelect>
              </div>
            </div>

            <div class="cascader-card">
              <ElCascaderPanel
                ref="actionPanelRef"
                v-model="selectedActionNodeValues"
                :options="filteredActionOptions"
                :props="actionCascaderProps"
                class="permission-cascader"
              >
                <template #default="{ node, data }">
                  <div class="panel-node" :class="{ 'is-leaf': node.isLeaf }">
                    <div class="panel-node__main">
                      <span class="panel-node__label">{{ data.label }}</span>
                      <template v-if="node.isLeaf">
                        <span
                          v-if="data.permissionText"
                          class="panel-node__meta panel-node__meta--truncate"
                          :title="data.permissionText"
                        >
                          {{ data.permissionText }}
                        </span>
                        <div class="panel-node__tags">
                          <ElTag
                            v-if="data.scopeText"
                            size="small"
                            effect="plain"
                            round
                          >
                            {{ data.scopeText }}
                          </ElTag>
                          <ElTag
                            v-if="data.sourceText"
                            size="small"
                            effect="plain"
                            round
                            type="info"
                          >
                            {{ data.sourceText }}
                          </ElTag>
                        </div>
                      </template>
                    </div>

                    <template v-if="node.isLeaf">
                      <div
                        v-if="decisionMap[data.value]"
                        class="panel-node__effect"
                        @click.stop
                        @mousedown.stop
                      >
                        <ElRadioGroup
                          v-model="decisionMap[data.value]"
                          size="small"
                          class="effect-radio-group"
                        >
                          <ElRadio value="allow" border>允许</ElRadio>
                          <ElRadio value="deny" border>拒绝</ElRadio>
                        </ElRadioGroup>
                      </div>
                      <ElTag
                        v-else
                        size="small"
                        effect="plain"
                        round
                        type="info"
                        class="panel-node__state"
                      >
                        继承
                      </ElTag>
                    </template>

                    <ElTag
                      v-else
                      size="small"
                      effect="plain"
                      round
                      type="info"
                      class="panel-node__count"
                    >
                      {{ `${data.selectedLeafCount || 0}/${data.totalLeafCount || 0}` }}
                    </ElTag>
                  </div>
                </template>
              </ElCascaderPanel>
            </div>

            <div class="cascader-footer">
              <span class="footer-note">仅对当前用户生效。未勾选项继续继承角色，勾选项可单独设为允许或拒绝。</span>
              <ElButton text @click="clearActionSelection">清空选择</ElButton>
            </div>
          </div>
        </ElTabPane>

        <ElTabPane label="角色继承" name="roles">
          <div class="roles-panel">
            <div class="roles-summary">
              <ElTag effect="plain" round>角色 {{ roleTags.length }}</ElTag>
              <ElTag type="warning" effect="plain" round>例外 {{ configuredActionIds.length }}</ElTag>
            </div>

            <div class="roles-list">
              <ElEmpty v-if="roleTags.length === 0" description="当前用户未绑定角色" />
              <div v-else class="role-tag-list">
                <ElTag
                  v-for="role in roleTags"
                  :key="role"
                  effect="plain"
                  round
                >
                  {{ role }}
                </ElTag>
              </div>
            </div>
          </div>
        </ElTabPane>

        <ElTabPane label="结果预览" name="preview">
          <div class="preview-panel">
            <div class="preview-summary">
              <ElTag effect="plain" round>允许 {{ allowPreviewCount }}</ElTag>
              <ElTag type="warning" effect="plain" round>拒绝 {{ denyPreviewCount }}</ElTag>
            </div>

            <ElScrollbar max-height="520px">
              <section
                v-for="group in previewGroups"
                :key="group.key"
                class="preview-group"
              >
                <header class="preview-group__header">
                  <span>{{ group.label }}</span>
                  <ElTag effect="plain" size="small" round>{{ group.items.length }}</ElTag>
                </header>
                <div class="preview-group__body">
                  <ElTag
                    v-for="item in group.items"
                    :key="item.id"
                    :type="item.effect === 'allow' ? 'success' : 'warning'"
                    effect="plain"
                    round
                  >
                    {{ item.name }} · {{ item.effect === 'allow' ? '允许' : '拒绝' }}
                  </ElTag>
                </div>
              </section>
            </ElScrollbar>
          </div>
        </ElTabPane>
      </ElTabs>
    </div>

    <template #footer>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { CascaderProps } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import {
  fetchGetAllScopes,
  fetchGetPermissionActionList,
  fetchGetUserActions,
  fetchSetUserActions
} from '@/api/system-manage'

interface Props {
  modelValue: boolean
  userData?: Api.SystemManage.UserListItem
}

interface CascaderOption {
  value: string
  label: string
  children?: CascaderOption[]
  leaf?: boolean
  permissionText?: string
  scopeText?: string
  sourceText?: string
  totalLeafCount?: number
  selectedLeafCount?: number
}

interface ActionOption extends CascaderOption {}

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
const saving = ref(false)
const activeTab = ref('custom')
const actionKeyword = ref('')
const actionScopeFilter = ref('')
const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
const scopeList = ref<Api.SystemManage.ScopeListItem[]>([])
const selectedActionNodeValues = ref<string[]>([])
const decisionMap = ref<Record<string, 'allow' | 'deny'>>({})
const actionPanelRef = ref()

const userTitle = computed(() => props.userData?.nickName || props.userData?.userName || '')
const configuredActionIds = computed(() => Object.keys(decisionMap.value))
const allowPreviewCount = computed(
  () => configuredActionIds.value.filter((id) => decisionMap.value[id] === 'allow').length
)
const denyPreviewCount = computed(
  () => configuredActionIds.value.filter((id) => decisionMap.value[id] === 'deny').length
)

const roleTags = computed(() => {
  const detailRoles = ((props.userData as any)?.roleDetails || []) as Array<{ code?: string; name?: string }>
  if (detailRoles.length > 0) {
    return detailRoles.map((item) => item.name || item.code || '').filter(Boolean)
  }
  const roleCodes = ((props.userData as any)?.roles || []) as string[]
  return roleCodes.filter(Boolean)
})

const actionOptions = computed<ActionOption[]>(() => {
  const featureMap = new Map<string, ActionOption>()

  permissionActions.value.forEach((action) => {
    const featureKey = `${action.featureKind || 'business'}`
    const moduleKey = `${action.moduleCode || action.resourceCode || 'default'}`

    if (!featureMap.has(featureKey)) {
      featureMap.set(featureKey, {
        value: `feature:${featureKey}`,
        label: formatFeature(featureKey),
        children: []
      })
    }

    const feature = featureMap.get(featureKey)!
    const modules = (feature.children || []) as ActionOption[]
    const moduleValue = `module:${featureKey}:${moduleKey}`
    let module = modules.find((item) => item.value === moduleValue)

    if (!module) {
      module = {
        value: moduleValue,
        label: action.moduleCode || action.resourceCode || '未分类模块',
        children: []
      }
      modules.push(module)
      feature.children = modules
    }

    const leaves = (module.children || []) as ActionOption[]
    leaves.push({
      value: action.id,
      label: action.name,
      leaf: true,
      permissionText: action.permissionKey || `${action.resourceCode}:${action.actionCode}`,
      scopeText: action.scopeName || action.scopeCode || '',
      sourceText: formatSource(action.source)
    })
    module.children = leaves
  })

  return Array.from(featureMap.values()).map((feature) => ({
    ...feature,
    totalLeafCount: countOptionLeaves(feature),
    selectedLeafCount: countConfiguredLeaves(feature, new Set(configuredActionIds.value)),
    children: ((feature.children || []) as ActionOption[]).map((module) => ({
      ...module,
      totalLeafCount: countOptionLeaves(module),
      selectedLeafCount: countConfiguredLeaves(module, new Set(configuredActionIds.value))
    }))
  }))
})

const actionBranchMap = computed(() => {
  const map = new Map<string, string[]>()
  actionOptions.value.forEach((feature) => {
    const featureLeaves: string[] = []
    ;((feature.children || []) as ActionOption[]).forEach((module) => {
      const moduleLeaves = ((module.children || []) as ActionOption[]).map((leaf) => `${leaf.value}`)
      map.set(`${module.value}`, moduleLeaves)
      featureLeaves.push(...moduleLeaves)
    })
    map.set(`${feature.value}`, featureLeaves)
  })
  return map
})

const expandedSelectedActionIds = computed(() =>
  expandSelectedValues(selectedActionNodeValues.value, actionBranchMap.value)
)

const actionLeafCount = computed(() => permissionActions.value.length)

const actionScopeOptions = computed(() =>
  scopeList.value.map((item) => item.scopeName || item.scopeCode).filter(Boolean)
)

const topLevelActionTags = computed(() => {
  const counter = new Map<string, number>()
  configuredActionIds.value.forEach((actionId) => {
    const action = permissionActions.value.find((item) => item.id === actionId)
    const featureKey = `${action?.featureKind || 'business'}`
    counter.set(featureKey, (counter.get(featureKey) || 0) + 1)
  })
  return Array.from(counter.entries()).map(([key, count]) => ({
    label: formatFeature(key),
    count
  }))
})

const previewGroups = computed(() => {
  const grouped = new Map<
    string,
    { key: string; label: string; items: Array<{ id: string; name: string; effect: 'allow' | 'deny' }> }
  >()

  configuredActionIds.value.forEach((actionId) => {
    const action = permissionActions.value.find((item) => item.id === actionId)
    if (!action) return
    const key = `${action.moduleCode || action.resourceCode || 'default'}`
    if (!grouped.has(key)) {
      grouped.set(key, {
        key,
        label: action.moduleCode || action.resourceCode || '未分类模块',
        items: []
      })
    }
    grouped.get(key)!.items.push({
      id: actionId,
      name: action.name,
      effect: decisionMap.value[actionId]
    })
  })

  return Array.from(grouped.values())
})

const actionCascaderProps: CascaderProps = {
  multiple: true,
  emitPath: false,
  checkStrictly: false,
  expandTrigger: 'click',
  checkOnClickNode: false,
  checkOnClickLeaf: false,
  showPrefix: true
}

const filteredActionOptions = computed(() => {
  const keyword = actionKeyword.value.trim().toLowerCase()
  const scope = actionScopeFilter.value.trim()

  return filterNestedOptions(actionOptions.value, (node) => {
    if (!node.leaf) return !keyword && !scope
    const text = [node.label, node.permissionText, node.scopeText, node.sourceText]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
    if (scope && node.scopeText !== scope) return false
    if (keyword && !text.includes(keyword)) return false
    return true
  })
})

watch(
  () => props.modelValue,
  (open) => {
    if (open) loadData()
  }
)

watch(
  () => filteredActionOptions.value,
  async () => {
    await nextTick()
    ensureExpandedMenus(actionPanelRef.value, selectedActionNodeValues.value)
  },
  { deep: true }
)

watch(
  () => selectedActionNodeValues.value,
  () => {
    const actionIds = expandedSelectedActionIds.value
    const nextMap: Record<string, 'allow' | 'deny'> = {}
    actionIds.forEach((actionId) => {
      nextMap[actionId] = decisionMap.value[actionId] || 'allow'
    })
    decisionMap.value = nextMap
  },
  { deep: true }
)

async function loadData() {
  if (!props.userData?.id) return
  loading.value = true
  activeTab.value = 'custom'
  actionKeyword.value = ''
  actionScopeFilter.value = ''

  try {
    const [scopeRes, actionsRes, currentRes] = await Promise.all([
      fetchGetAllScopes(),
      fetchGetPermissionActionList({ current: 1, size: 1000, status: 'normal' }),
      fetchGetUserActions(props.userData.id)
    ])

    scopeList.value = scopeRes || []
    permissionActions.value = actionsRes?.records || []
    const availableActionIDSet = new Set(permissionActions.value.map((item) => item.id))

    const nextMap: Record<string, 'allow' | 'deny'> = {}
    currentRes.forEach((item) => {
      if (
        item.actionId &&
        availableActionIDSet.has(item.actionId) &&
        (item.effect === 'allow' || item.effect === 'deny')
      ) {
        nextMap[item.actionId] = item.effect
      }
    })
    decisionMap.value = nextMap
    selectedActionNodeValues.value = Object.keys(nextMap)

    await nextTick()
    ensureExpandedMenus(actionPanelRef.value, selectedActionNodeValues.value)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载用户权限失败')
  } finally {
    loading.value = false
  }
}

function filterNestedOptions<T extends CascaderOption>(
  nodes: T[],
  matcher: (node: T) => boolean
): T[] {
  return nodes
    .map((node) => {
      const children = Array.isArray(node.children)
        ? filterNestedOptions(node.children as T[], matcher)
        : []
      const selfMatched = matcher(node)
      if (selfMatched || children.length > 0) {
        return {
          ...node,
          children
        }
      }
      return null
    })
    .filter(Boolean) as T[]
}

function expandSelectedValues(values: string[], branchMap: Map<string, string[]>) {
  const set = new Set<string>()
  values.forEach((value) => {
    const key = `${value || ''}`
    const mapped = branchMap.get(key)
    if (mapped?.length) {
      mapped.forEach((item) => set.add(item))
      return
    }
    set.add(key)
  })
  return Array.from(set)
}

function ensureExpandedMenus(panel: any, selectedValues: string[]) {
  const rootMenus = panel?.menus?.[0]
  if (!panel || !rootMenus?.length) return

  const firstValue = selectedValues[0]
  let featureNode = rootMenus[0]
  let moduleNode = featureNode?.children?.[0]

  if (firstValue) {
    const matchedNode = panel.getFlattedNodes?.(false)?.find((node: any) => `${node?.value}` === `${firstValue}`)
    const pathNodes = matchedNode?.pathNodes || []
    if (pathNodes[0]) featureNode = pathNodes[0]
    if (pathNodes[1]) moduleNode = pathNodes[1]
  }

  const nextMenus = [rootMenus]
  if (featureNode?.children?.length) nextMenus.push(featureNode.children)
  if (moduleNode?.children?.length) nextMenus.push(moduleNode.children)
  panel.menus = nextMenus
}

function countOptionLeaves(node: CascaderOption): number {
  if (!node.children?.length) return 1
  return node.children.reduce((sum, child) => sum + countOptionLeaves(child), 0)
}

function countConfiguredLeaves(node: CascaderOption, selectedSet: Set<string>): number {
  if (!node.children?.length) return selectedSet.has(node.value) ? 1 : 0
  return node.children.reduce((sum, child) => sum + countConfiguredLeaves(child, selectedSet), 0)
}

function clearActionSelection() {
  selectedActionNodeValues.value = []
}

function formatFeature(value: string) {
  if (value === 'system') return '系统功能'
  if (value === 'business') return '业务功能'
  return value
}

function formatSource(source?: string) {
  if (source === 'api') return '接口自动'
  if (source === 'system') return '系统内置'
  if (source === 'business') return '业务定义'
  return source || ''
}

function handleCancel() {
  visible.value = false
}

async function handleSave() {
  if (!props.userData?.id) return
  saving.value = true
  try {
    const payload = Object.entries(decisionMap.value).map(([actionId, effect]) => ({
      action_id: actionId,
      effect
    }))
    await fetchSetUserActions(props.userData.id, payload)
    ElMessage.success('用户功能权限已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存用户权限失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped lang="scss">
.dialog-shell {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dialog-note {
  color: #6b7280;
  line-height: 1.6;
}

.permission-tabs :deep(.el-tabs__content) {
  padding-top: 12px;
}

.cascader-pane,
.roles-panel,
.preview-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.cascader-toolbar,
.roles-summary,
.preview-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.cascader-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.toolbar-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.toolbar-search {
  width: 220px;
}

.toolbar-scope {
  width: 140px;
}

.cascader-card,
.roles-list,
.preview-group {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  overflow: hidden;
  background: #fff;
}

.permission-cascader {
  width: 100%;
  height: 420px;
}

.permission-cascader :deep(.el-cascader-panel) {
  height: 100%;
  display: flex;
  align-items: stretch;
}

.permission-cascader :deep(.el-cascader-menu) {
  width: auto;
  flex: 0 1 auto;
  min-width: 260px;
}

.permission-cascader :deep(.el-cascader-menu:nth-child(2)) {
  width: auto;
  flex: 0 1 auto;
  min-width: 320px;
}

.permission-cascader :deep(.el-cascader-menu:last-child) {
  flex: 1 1 auto;
  width: auto;
}

.permission-cascader :deep(.el-cascader-menu__wrap),
.permission-cascader :deep(.el-scrollbar),
.permission-cascader :deep(.el-scrollbar__wrap),
.permission-cascader :deep(.el-scrollbar__view) {
  height: 100%;
}

.panel-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  min-width: 0;
}

.panel-node__main {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1 1 auto;
}

.panel-node__label {
  color: #111827;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
}

.panel-node__meta {
  color: #6b7280;
  font-size: 12px;
  line-height: 1.4;
}

.panel-node__meta--truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.panel-node__tags {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 0 0 auto;
}

.panel-node__count,
.panel-node__state {
  flex: 0 0 auto;
}

.panel-node__effect {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex: 0 0 auto;
}

.effect-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.effect-radio-group :deep(.el-radio.is-bordered) {
  height: 28px;
  margin-right: 0;
  padding: 0 10px;
}

.cascader-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.footer-note {
  color: #6b7280;
  line-height: 1.6;
}

.roles-list {
  min-height: 220px;
  padding: 20px;
}

.role-tag-list,
.preview-group__body {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.preview-group {
  margin-bottom: 12px;
}

.preview-group__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #fbfcfe;
  border-bottom: 1px solid var(--el-border-color-lighter);
  color: #111827;
  font-weight: 600;
}

.preview-group__body {
  padding: 16px;
}
</style>
