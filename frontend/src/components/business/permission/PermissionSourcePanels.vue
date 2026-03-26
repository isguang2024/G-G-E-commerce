<template>
  <div v-if="derivedItems.length || blockedItems.length" class="source-detail-grid">
    <div v-if="derivedItems.length" class="source-card source-card--derived">
      <div class="source-header">
        <div class="source-title">{{ derivedTitle }}</div>
        <ElButton
          v-if="selectedPackage"
          type="warning"
          text
          @click="goToFeaturePackagePage(selectedPackage)"
        >
          {{ jumpButtonText }}
        </ElButton>
      </div>

      <div v-if="sourcePackages.length" class="package-filter-row">
        <ElTag
          :type="modelValue ? 'info' : 'warning'"
          effect="plain"
          round
          class="package-filter-tag"
          @click="emit('update:modelValue', '')"
        >
          {{ packageFilterAllLabel }}
        </ElTag>
        <ElTag
          v-for="item in sourcePackages"
          :key="item.id"
          :type="modelValue === item.id ? 'warning' : 'info'"
          effect="plain"
          round
          class="package-filter-tag"
          @click="emit('update:modelValue', modelValue === item.id ? '' : item.id)"
        >
          {{ item.name }}
        </ElTag>
      </div>

      <div class="source-tags">
        <ElTag
          v-for="item in filteredDerivedItems"
          :key="item.id"
          :type="derivedTagType"
          effect="plain"
          round
          class="package-tag-link"
          :title="buildSourceText(item.id)"
          @click="goToSourcePackage(item.id)"
        >
          {{ item.label }}
        </ElTag>
      </div>

      <div v-if="sourcePackages.length && filteredDerivedItems.length === 0" class="source-empty-text">
        {{ filteredDerivedEmptyText }}
      </div>
    </div>

    <div v-if="blockedItems.length" class="source-card" :class="blockedCardClass">
      <div class="source-title">{{ blockedTitle }}</div>
      <div class="source-tags">
        <ElTag
          v-for="item in filteredBlockedItems"
          :key="item.id"
          :type="blockedTagType"
          effect="plain"
          round
          class="package-tag-link"
          :title="buildSourceText(item.id)"
          @click="goToSourcePackage(item.id)"
        >
          {{ item.label }}
        </ElTag>
      </div>

      <div v-if="sourcePackages.length && filteredBlockedItems.length === 0" class="source-empty-text">
        {{ filteredBlockedEmptyText }}
      </div>
    </div>
  </div>

  <div v-else class="source-card source-card--empty">
    <div class="source-title">{{ emptyTitle }}</div>
    <div class="source-empty-text">{{ emptyText }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

interface SourceItem {
  id: string
  label: string
}

interface Props {
  modelValue: string
  packages: Api.SystemManage.FeaturePackageItem[]
  sourceMap: Record<string, string[]>
  derivedItems: SourceItem[]
  blockedItems: SourceItem[]
  derivedTitle: string
  blockedTitle: string
  open: 'menus' | 'actions'
  derivedTagType?: 'success' | 'warning' | 'info' | 'primary' | 'danger'
  blockedTagType?: 'success' | 'warning' | 'info' | 'primary' | 'danger'
  packageFilterAllLabel?: string
  jumpButtonText?: string
  filteredDerivedEmptyText?: string
  filteredBlockedEmptyText?: string
  emptyTitle?: string
  emptyText?: string
}

const props = withDefaults(defineProps<Props>(), {
  derivedTagType: 'warning',
  blockedTagType: 'danger',
  packageFilterAllLabel: '全部功能包',
  jumpButtonText: '前往功能包',
  filteredDerivedEmptyText: '当前筛选下暂无功能包展开项',
  filteredBlockedEmptyText: '当前筛选下暂无显式裁剪项',
  emptyTitle: '当前暂无来源明细',
  emptyText: '请先检查功能包绑定、上下文范围或当前主体是否已进入功能包主链。'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const router = useRouter()

const sourcePackages = computed(() => {
  const packageIdSet = new Set(Object.values(props.sourceMap).flat())
  return props.packages.filter((item) => packageIdSet.has(item.id))
})

const selectedPackage = computed(
  () => props.packages.find((item) => item.id === props.modelValue) || null
)

const filteredDerivedItems = computed(() => {
  if (!props.modelValue) return props.derivedItems
  return props.derivedItems.filter((item) => (props.sourceMap[item.id] || []).includes(props.modelValue))
})

const filteredBlockedItems = computed(() => {
  if (!props.modelValue) return props.blockedItems
  return props.blockedItems.filter((item) => (props.sourceMap[item.id] || []).includes(props.modelValue))
})

const blockedCardClass = computed(() => {
  if (props.blockedTagType === 'primary' || props.blockedTagType === 'info') {
    return 'source-card--trimmed'
  }
  return 'source-card--blocked'
})

function buildSourceText(id: string) {
  const packageIdSet = new Set(props.sourceMap[id] || [])
  const names = props.packages.filter((item) => packageIdSet.has(item.id)).map((item) => item.name)
  return names.length ? `来源功能包：${names.join('、')}` : '来源功能包未命中'
}

function getSourcePackage(id: string) {
  const sourceIDSet = new Set(props.sourceMap[id] || [])
  return props.packages.find((item) => sourceIDSet.has(item.id)) || null
}

function goToSourcePackage(id: string) {
  const pkg = getSourcePackage(id)
  if (!pkg) return
  goToFeaturePackagePage(pkg)
}

function goToFeaturePackagePage(item: Api.SystemManage.FeaturePackageItem) {
  router.push({
    name: 'FeaturePackage',
    query: {
      packageKey: item.packageKey,
      packageType: item.packageType || 'base',
      contextType: item.contextType || 'platform',
      open: props.open
    }
  })
}
</script>

<style scoped lang="scss">
.source-detail-grid {
  display: grid;
  gap: 12px;
}

.source-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  background: #fff;
}

.source-card--derived {
  border-color: #f3d38a;
  background: #fffaf0;
}

.source-card--blocked {
  border-color: #f1c0c0;
  background: #fff6f6;
}

.source-card--trimmed {
  border-color: #bfd3ff;
  background: #f5f9ff;
}

.source-card--empty {
  border-style: dashed;
  background: #f8fafc;
}

.source-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.source-title {
  font-size: 13px;
  color: #475569;
}

.source-tags,
.package-filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.package-filter-tag,
.package-tag-link {
  cursor: pointer;
}

.source-empty-text {
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.6;
}
</style>
