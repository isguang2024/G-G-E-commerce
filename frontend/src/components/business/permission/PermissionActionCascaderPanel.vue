<template>
  <div class="permission-cascader-panel">
    <div class="panel-header">
      <div class="header-side">
        <div v-if="topLevelTags.length" class="tag-summary">
          <ElTag
            v-for="item in topLevelTags"
            :key="item.label"
            effect="plain"
            round
            class="summary-tag"
          >
            {{ item.count > 0 ? `${item.label} ${item.count}` : item.label }}
          </ElTag>
        </div>
        <ElTag effect="plain" round class="summary-tag">已选 {{ selectedLeafIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round class="summary-tag">总计 {{ leafCount }}</ElTag>

        <ElInput
          v-model="searchKeyword"
          clearable
          placeholder="搜索标题或资源动作"
          :prefix-icon="Search"
          class="toolbar-search"
        />
      </div>
    </div>

    <div class="panel-body">
      <ElCascaderPanel
        ref="panelRef"
        v-model="selectedNodeValues"
        :options="filteredOptions"
        :props="cascaderProps"
        class="permission-panel"
      >
        <template #default="{ node, data }">
          <div class="node-content" :class="{ 'is-leaf': node.isLeaf }">
            <div class="node-main">
              <span class="node-label">{{ data.label }}</span>
            </div>

            <div v-if="node.isLeaf" class="node-leaf-meta">
              <span v-if="data.permissionText" class="meta-text" :title="data.permissionText">
                {{ data.permissionText }}
              </span>
              <ElTag
                v-if="data.sourceText"
                size="small"
                effect="plain"
                round
                type="info"
                class="meta-tag"
              >
                {{ data.sourceText }}
              </ElTag>
            </div>

            <ElTag
              v-else
              size="small"
              effect="plain"
              round
              type="info"
              class="count-tag"
            >
              {{ `${data.selectedLeafCount || 0}/${data.totalLeafCount || 0}` }}
            </ElTag>
          </div>
        </template>
      </ElCascaderPanel>
    </div>

    <div class="panel-footer">
      <div class="footer-note">{{ footerText }}</div>
      <div class="footer-actions">
        <ElButton text @click="clearSelection">清空选择</ElButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import type { CascaderOption, CascaderProps } from 'element-plus'

interface PermissionLeafItem extends Partial<Api.SystemManage.PermissionActionItem> {
  id: string
  name: string
}

interface CascaderPanelExpose {
  menus?: any[][]
  getFlattedNodes?: (leafOnly?: boolean) => any[]
}

interface PermissionOption extends CascaderOption {
  permissionText?: string
  sourceText?: string
  totalLeafCount?: number
  selectedLeafCount?: number
}

interface Props {
  actions: PermissionLeafItem[]
  selectedIds: string[]
  footerText?: string
}

const props = withDefaults(defineProps<Props>(), {
  footerText: '支持多选，父级和子级都可以直接点击选择。'
})

const emit = defineEmits<{
  (e: 'update:selectedIds', value: string[]): void
}>()

const panelRef = ref<CascaderPanelExpose>()
const selectedNodeValues = ref<string[]>([])
const syncingFromProps = ref(false)
const searchKeyword = ref('')

const baseOptions = computed<PermissionOption[]>(() => {
  const featureMap = new Map<string, PermissionOption>()

  props.actions.forEach((action) => {
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
    const children = (feature.children || []) as PermissionOption[]
    const normalizedModuleValue = `module:${featureKey}:${moduleKey}`
    let module = children.find((item) => item.value === normalizedModuleValue)

    if (!module) {
      module = {
        value: normalizedModuleValue,
        label: action.moduleCode || action.resourceCode || '未分类模块',
        children: []
      }
      children.push(module)
      feature.children = children
    }

    const leafChildren = (module.children || []) as PermissionOption[]
    leafChildren.push({
      value: action.id,
      label: action.name,
      leaf: true,
      permissionText: action.permissionKey || `${action.resourceCode}:${action.actionCode}`,
      sourceText: formatSource(action.source)
    })
    module.children = leafChildren
  })

  return Array.from(featureMap.values())
})

const leafCount = computed(() => props.actions.length)

const branchLeafMap = computed(() => {
  const map = new Map<string, string[]>()
  baseOptions.value.forEach((feature) => {
    const featureLeaves: string[] = []
    ;((feature.children || []) as PermissionOption[]).forEach((module) => {
      const moduleLeaves = ((module.children || []) as PermissionOption[]).map((leaf) => `${leaf.value}`)
      map.set(`${module.value}`, moduleLeaves)
      featureLeaves.push(...moduleLeaves)
    })
    map.set(`${feature.value}`, featureLeaves)
  })
  return map
})

const selectedLeafIds = computed(() => {
  const leafSet = new Set<string>()
  selectedNodeValues.value.forEach((value) => {
    const key = `${value || ''}`
    const mapped = branchLeafMap.value.get(key)
    if (mapped?.length) {
      mapped.forEach((id) => leafSet.add(id))
      return
    }
    leafSet.add(key)
  })
  return Array.from(leafSet)
})

const selectedLeafIdSet = computed(() => new Set(selectedLeafIds.value))

const options = computed<PermissionOption[]>(() => {
  const attachCounts = (nodes: PermissionOption[]): PermissionOption[] =>
    nodes.map((node) => {
      const children = Array.isArray(node.children)
        ? attachCounts(node.children as PermissionOption[])
        : []

      if (node.leaf) {
        return {
          ...node,
          children,
          totalLeafCount: 1,
          selectedLeafCount: selectedLeafIdSet.value.has(`${node.value}`) ? 1 : 0
        }
      }

      const totalLeafCount = children.reduce((sum, child) => sum + (child.totalLeafCount || 0), 0)
      const selectedLeafCount = children.reduce((sum, child) => sum + (child.selectedLeafCount || 0), 0)

      return {
        ...node,
        children,
        totalLeafCount,
        selectedLeafCount
      }
    })

  return attachCounts(baseOptions.value)
})

const cascaderProps: CascaderProps = {
  multiple: true,
  emitPath: false,
  checkStrictly: false,
  expandTrigger: 'click',
  checkOnClickNode: false,
  checkOnClickLeaf: false,
  showPrefix: true
}

const topLevelTags = computed(() => {
  const counter = new Map<string, number>()
  selectedLeafIds.value.forEach((value) => {
    const featureKey = resolveFeatureKey(value)
    counter.set(featureKey, (counter.get(featureKey) || 0) + 1)
  })
  return Array.from(counter.entries()).map(([value, count]) => ({
    label: formatFeature(value),
    count
  }))
})

const filteredOptions = computed<PermissionOption[]>(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()

  const filterNodes = (nodes: PermissionOption[]): PermissionOption[] => {
    return nodes
      .map((node) => {
        const children = Array.isArray(node.children) ? filterNodes(node.children as PermissionOption[]) : []
        const selfMatched = matchesNode(node, keyword)
        if (selfMatched || children.length > 0) {
          return { ...node, children }
        }
        return null
      })
      .filter(Boolean) as PermissionOption[]
  }

  return filterNodes(options.value)
})

watch(
  () => props.selectedIds,
  async (value) => {
    const next = Array.isArray(value) ? [...value] : []
    if (isSameArray(next, selectedNodeValues.value)) return
    syncingFromProps.value = true
    selectedNodeValues.value = next
    await nextTick()
    syncingFromProps.value = false
    ensureExpandedMenus()
  },
  { immediate: true, deep: true }
)

watch(
  () => filteredOptions.value,
  async () => {
    await nextTick()
    ensureExpandedMenus()
  },
  { deep: true }
)

watch(
  selectedNodeValues,
  (value) => {
    if (!syncingFromProps.value) {
      emit('update:selectedIds', [...value])
    }
  },
  { deep: true }
)

function clearSelection() {
  selectedNodeValues.value = []
}

function ensureExpandedMenus() {
  const panel = panelRef.value
  const rootMenus = panel?.menus?.[0]
  if (!panel || !rootMenus?.length) return

  const firstValue = selectedNodeValues.value[0]
  let featureNode = rootMenus[0]
  let moduleNode = featureNode?.children?.[0]

  if (firstValue) {
    const matchedNode = panel.getFlattedNodes?.(false)?.find((node) => `${node?.value}` === `${firstValue}`)
    const pathNodes = matchedNode?.pathNodes || []
    if (pathNodes[0]) featureNode = pathNodes[0]
    if (pathNodes[1]) moduleNode = pathNodes[1]
  }

  const nextMenus = [rootMenus]
  if (featureNode?.children?.length) nextMenus.push(featureNode.children)
  if (moduleNode?.children?.length) nextMenus.push(moduleNode.children)
  panel.menus = nextMenus
}

function matchesNode(node: PermissionOption, keyword: string) {
  if (!node.leaf) return !keyword

  const text = [node.label, node.permissionText, node.sourceText]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()

  if (keyword && !text.includes(keyword)) return false
  return true
}

function isSameArray(a: string[], b: string[]) {
  if (a.length !== b.length) return false
  return a.every((item, index) => item === b[index])
}

function resolveFeatureKey(value: string) {
  if (value.startsWith('feature:')) return value.slice('feature:'.length) || 'business'
  if (value.startsWith('module:')) return value.split(':')[1] || 'business'
  const action = props.actions.find((item) => item.id === value)
  return `${action?.featureKind || 'business'}`
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
</script>

<style scoped lang="scss">
.permission-cascader-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-header,
.panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.header-side {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  flex-wrap: nowrap;
}

.tag-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: nowrap;
  flex: 0 1 auto;
}

.summary-tag {
  height: 32px;
  line-height: 30px;
  flex: 0 0 auto;
}

.summary-tag :deep(.el-tag__content) {
  display: inline-flex;
  align-items: center;
  line-height: 30px;
  font-size: 12px;
  font-weight: 400;
}

.toolbar-search {
  width: 210px;
  flex: 0 0 210px;
  margin-left: 8px;
}

.panel-body {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  overflow: hidden;
  background: #fff;
}

.permission-panel {
  width: 100%;
  height: 360px;
}

.permission-panel :deep(.el-cascader-panel) {
  height: 100%;
  display: flex;
  align-items: stretch;
}

.permission-panel :deep(.el-cascader-menu) {
  height: 100%;
  min-width: 260px;
  width: auto;
  flex: 0 1 auto;
}

.permission-panel :deep(.el-cascader-menu:nth-child(2)) {
  min-width: 320px;
  width: auto;
  flex: 0 1 auto;
}

.permission-panel :deep(.el-cascader-menu:last-child) {
  width: auto;
  flex: 1 1 auto;
}

.permission-panel :deep(.el-scrollbar),
.permission-panel :deep(.el-scrollbar__wrap),
.permission-panel :deep(.el-scrollbar__view) {
  height: 100%;
}

.permission-panel :deep(.el-cascader-node) {
  min-height: 36px;
}

.node-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
}

.node-content.is-leaf {
  gap: 12px;
}

.node-main {
  display: flex;
  align-items: center;
  min-width: 0;
  flex: 1 1 auto;
}

.node-label {
  color: #111827;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 0 1 auto;
}

.node-leaf-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  flex: 1 1 auto;
  justify-content: flex-end;
}

.meta-text {
  color: #6b7280;
  font-size: 11px;
  line-height: 1;
  min-width: 0;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 0 1 auto;
}

.meta-tag,
.count-tag {
  height: 22px;
  line-height: 20px;
  flex: 0 0 auto;
}

.meta-tag :deep(.el-tag__content),
.count-tag :deep(.el-tag__content) {
  line-height: 20px;
  font-size: 11px;
}

.footer-note {
  color: #6b7280;
  font-size: 12px;
  line-height: 1.35;
}

.toolbar-search :deep(.el-input__wrapper) {
  min-height: 36px;
  height: 36px;
}

@media (max-width: 900px) {
  .panel-header,
  .panel-footer,
  .header-side {
    flex-direction: column;
    align-items: stretch;
  }

  .tag-summary {
    flex-wrap: wrap;
  }

  .toolbar-search {
    width: 100%;
    flex: 1 1 auto;
    margin-left: 0;
  }

  .node-content.is-leaf {
    align-items: flex-start;
    flex-direction: column;
  }

  .node-leaf-meta {
    justify-content: flex-start;
    flex-wrap: wrap;
  }
}
</style>
