<template>
  <ElDialog
    v-model="visible"
    :title="`用户权限例外审计 - ${userTitle}`"
    width="1120px"
    destroy-on-close
    class="user-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
        <div class="dialog-note">
          平台用户请优先绑定功能包和菜单裁剪。这里保留的是历史兼容权限例外审计视图，用于查看既有 allow/deny 例外，不再作为主配置入口。
        </div>
        <div
          class="compat-banner"
          :class="hasPackageConfig ? 'compat-banner--success' : 'compat-banner--warning'"
        >
          <span>
            {{
              hasPackageConfig
              ? '当前用户已进入功能包约束模式：这里只能对已绑定功能包展开范围内的能力做个人例外。'
              : '当前用户尚未绑定功能包，仍处于兼容回退模式：这里只展示历史兼容例外，建议先绑定功能包并使用菜单裁剪。'
            }}
          </span>
          <ElButton v-if="!hasPackageConfig" type="warning" text @click="emit('open-packages')">
            前往绑定功能包
          </ElButton>
        </div>

      <ElTabs v-model="activeTab" class="permission-tabs">
        <ElTabPane label="兼容例外审计" name="custom">
          <div class="cascader-pane">
            <div class="cascader-toolbar">
              <div class="cascader-tags">
                <ElTag type="success" effect="plain" round>功能包 {{ featurePackages.length }}</ElTag>
                <ElTag type="warning" effect="plain" round>直绑 {{ directPackageIds.length }}</ElTag>
                <ElTag type="info" effect="plain" round>展开 {{ expandedPackageIds.length }}</ElTag>
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
              <div v-if="featurePackages.length" class="package-tags">
                <ElTag
                  v-for="item in featurePackages"
                  :key="item.id"
                  type="success"
                  effect="plain"
                  round
                  class="package-tag-link"
                  @click="goToFeaturePackagePage(item)"
                >
                  {{ item.name }}
                </ElTag>
              </div>
              <div v-else class="package-fallback-note">
                当前未绑定功能包，权限例外页正在使用兼容候选范围。
              </div>

              <div class="toolbar-filters">
                <ElInput
                  v-model="actionKeyword"
                  clearable
                  placeholder="搜索功能标题或资源动作"
                  :prefix-icon="Search"
                  class="toolbar-search"
                />

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
                            v-if="data.sourceText"
                            size="small"
                            effect="plain"
                            round
                            type="info"
                          >
                            {{ data.sourceText }}
                          </ElTag>
                          <ElTag
                            v-if="data.packageText"
                            size="small"
                            effect="plain"
                            round
                            type="success"
                            :title="data.packageText"
                            class="panel-node__package-tag"
                            @click.stop="goToActionPackagePage(data.value)"
                          >
                            {{ data.packageText }}
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
                        <ElTag
                          size="small"
                          effect="plain"
                          round
                          :type="decisionMap[data.value] === 'allow' ? 'success' : 'warning'"
                        >
                          {{ decisionMap[data.value] === 'allow' ? '允许' : '拒绝' }}
                        </ElTag>
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
              <span class="footer-note">当前页仅用于审计历史兼容例外。平台用户正式主链为“功能包 + 菜单裁剪”。</span>
            </div>
          </div>
        </ElTabPane>

        <ElTabPane label="角色继承" name="roles">
          <div class="roles-panel">
            <div class="roles-summary">
              <ElTag effect="plain" round>角色 {{ roleTags.length }}</ElTag>
              <ElTag type="success" effect="plain" round>功能包 {{ featurePackages.length }}</ElTag>
              <ElTag :type="hasPackageConfig ? 'success' : 'warning'" effect="plain" round>
                {{ hasPackageConfig ? '功能包约束' : '兼容回退' }}
              </ElTag>
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
              <div v-if="featurePackages.length" class="roles-package-list">
                <ElTag
                  v-for="item in featurePackages"
                  :key="item.id"
                  type="success"
                  effect="plain"
                  round
                  class="package-tag-link"
                  @click="goToFeaturePackagePage(item)"
                >
                  {{ item.name }}
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
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { CascaderOption, CascaderProps } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { fetchGetUserPackages, fetchGetUserActionOverrides } from '@/api/system-manage'

interface Props {
  modelValue: boolean
  userData?: Api.SystemManage.UserListItem
}

interface ActionOption extends CascaderOption {
  permissionText?: string
  sourceText?: string
  packageText?: string
  totalLeafCount?: number
  selectedLeafCount?: number
}

const props = defineProps<Props>()
const router = useRouter()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
  (e: 'open-packages'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loading = ref(false)
const activeTab = ref('custom')
const actionKeyword = ref('')
const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
const availableActionIds = ref<string[]>([])
const directPackageIds = ref<string[]>([])
const expandedPackageIds = ref<string[]>([])
const derivedSourceMap = ref<Record<string, string[]>>({})
const hasPackageConfig = ref(false)
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
      sourceText: formatSource(action.source),
      packageText: buildPackageText(action.id)
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

  return filterNestedOptions(actionOptions.value, (node) => {
    if (!node.leaf) return !keyword
    const text = [node.label, node.permissionText, node.sourceText]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
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

  try {
    const [currentRes, packageRes] = await Promise.all([
      fetchGetUserActionOverrides(props.userData.id),
      fetchGetUserPackages(props.userData.id)
    ])

    hasPackageConfig.value = Boolean(currentRes?.has_package_config)
    availableActionIds.value = currentRes?.available_action_ids || []
    directPackageIds.value = packageRes?.package_ids || []
    expandedPackageIds.value = currentRes?.expanded_package_ids || []
    featurePackages.value = packageRes?.packages || []
    derivedSourceMap.value = Object.fromEntries(
      (currentRes?.derived_sources || []).map((item: { action_id: string; package_ids: string[] }) => [
        item.action_id,
        item.package_ids
      ])
    )
    const compatActions = currentRes?.actions || []
    const compatActionItems = compatActions
      .map((item) => item.action)
      .filter(Boolean) as Api.SystemManage.PermissionActionItem[]
    permissionActions.value =
      currentRes?.available_actions?.length
        ? currentRes.available_actions
        : compatActionItems
    const availableActionIDSet = new Set(
      availableActionIds.value.length
        ? availableActionIds.value
        : permissionActions.value.map((item) => item.id)
    )

    const nextMap: Record<string, 'allow' | 'deny'> = {}
    compatActions.forEach((item) => {
      if (
        item.action_id &&
        availableActionIDSet.has(item.action_id) &&
        (item.effect === 'allow' || item.effect === 'deny')
      ) {
        nextMap[item.action_id] = item.effect
      }
    })
    decisionMap.value = nextMap
    selectedActionNodeValues.value = Object.keys(nextMap)

    await nextTick()
    ensureExpandedMenus(actionPanelRef.value, selectedActionNodeValues.value)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载用户权限例外失败')
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
  if (!node.children?.length) return selectedSet.has(`${node.value || ''}`) ? 1 : 0
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

function buildPackageText(actionId: string) {
  const packageIdSet = new Set(derivedSourceMap.value[actionId] || [])
  const names = featurePackages.value.filter((item) => packageIdSet.has(item.id)).map((item) => item.name)
  if (!names.length) return ''
  if (names.length === 1) return `来自 ${names[0]}`
  return `来自 ${names[0]} 等${names.length}个包`
}

function getActionPackage(actionId: string) {
  const packageIdSet = new Set(derivedSourceMap.value[actionId] || [])
  return featurePackages.value.find((item) => packageIdSet.has(item.id)) || null
}

function goToFeaturePackagePage(item: Api.SystemManage.FeaturePackageItem) {
  router.push({
    name: 'FeaturePackage',
    query: {
      packageKey: item.packageKey,
      contextType: item.contextType || 'platform'
    }
  })
}

function goToActionPackagePage(actionId: string | number | undefined) {
  const target = getActionPackage(`${actionId || ''}`)
  if (!target) return
  goToFeaturePackagePage(target)
}

function handleCancel() {
  visible.value = false
}

async function handleSave() {
  return
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

.compat-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.6;
}

.compat-banner--success {
  color: #166534;
  border: 1px solid #bbf7d0;
  background: #f0fdf4;
}

.compat-banner--warning {
  color: #92400e;
  border: 1px solid #fde68a;
  background: #fffbeb;
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

.panel-node__package-tag,
.package-tag-link {
  cursor: pointer;
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

.package-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.package-fallback-note {
  color: #92400e;
  font-size: 13px;
  line-height: 1.6;
}

.role-tag-list,
.roles-package-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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
