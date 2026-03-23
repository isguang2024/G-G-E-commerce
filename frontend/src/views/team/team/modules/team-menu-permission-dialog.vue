<template>
  <ElDialog
    v-model="visible"
    :title="`团队菜单边界 - ${teamName}`"
    width="960px"
    destroy-on-close
    class="team-menu-boundary-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        这里配置的是团队菜单边界减法。正式开通入口请通过功能包完成；此处只会从功能包展开菜单中屏蔽个别入口，不会额外新增菜单能力。
      </div>

      <PermissionSummaryTags :items="summaryItems" />

      <div v-if="featurePackages.length" class="package-card">
        <div class="package-title">已开通功能包</div>
        <div class="package-help">功能包决定团队可见菜单候选范围；当前弹窗仅负责在候选范围内做屏蔽。</div>
        <div class="package-tags">
          <ElTag
            v-for="item in featurePackages"
            :key="item.id"
            type="success"
            effect="plain"
            round
          >
            {{ item.name }} · {{ item.menuCount ?? 0 }}
          </ElTag>
        </div>
      </div>

      <PermissionSourcePanels
        v-model="selectedDerivedPackageId"
        :packages="featurePackages"
        :source-map="derivedSourceMap"
        :derived-items="derivedMenus"
        :blocked-items="blockedMenus"
        derived-title="功能包展开菜单"
        blocked-title="团队边界已屏蔽菜单"
        open="menus"
        blocked-tag-type="primary"
        filtered-blocked-empty-text="当前筛选下暂无团队边界屏蔽菜单"
        empty-title="当前暂无团队菜单来源"
        empty-text="请先检查团队功能包、菜单快照或团队边界是否已经生成。"
      />

      <div class="tree-shell">
        <ElTree
          ref="treeRef"
          :data="menuTree"
          node-key="id"
          show-checkbox
          :default-expand-all="expandAll"
          :props="defaultProps"
          @check="handleCheck"
        />
      </div>
    </div>

    <template #footer>
      <ElButton @click="toggleExpand">{{ expandAll ? '全部收起' : '全部展开' }}</ElButton>
      <ElButton @click="checkAll">全部保留</ElButton>
      <ElButton @click="clearAll">全部屏蔽</ElButton>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { formatMenuTitle } from '@/utils/router'
import { fetchGetMenuTreeAll, fetchGetTeamFeaturePackages } from '@/api/system-manage'
import { fetchGetTeamMenus, fetchGetTeamMenuOrigins, fetchSetTeamMenus } from '@/api/team'
import type { AppRouteRecord } from '@/types/router'
import PermissionSourcePanels from '@/components/business/permission/PermissionSourcePanels.vue'
import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'

interface Props {
  modelValue: boolean
  teamId: string
  teamName: string
}

const props = defineProps<Props>()
const router = useRouter()
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
const expandAll = ref(true)
const treeRef = ref()
const menuTree = ref<AppRouteRecord[]>([])
const selectedIds = ref<string[]>([])
const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
const derivedMenuIds = ref<string[]>([])
const blockedMenuIds = ref<string[]>([])
const derivedSourceMap = ref<Record<string, string[]>>({})
const menuSourceList = ref<Array<{ id: string; label: string }>>([])
const selectedDerivedPackageId = ref('')

const defaultProps = {
  children: 'children',
  label: (data: any) =>
    String(formatMenuTitle(data?.meta?.title) || data?.name || data?.path || data?.id || '')
}

const blockedMenuCount = computed(() => blockedMenuIds.value.length)
const summaryItems = computed(() => [
  { label: '团队', value: props.teamName || '-' },
  { label: '边界结果', value: selectedIds.value.length, type: 'success' as const },
  { label: '候选', value: derivedMenuIds.value.length, type: 'info' as const },
  { label: '功能包', value: featurePackages.value.length },
  { label: '功能包展开', value: derivedMenuIds.value.length, type: 'warning' as const },
      { label: '边界已屏蔽', value: blockedMenuCount.value, type: 'primary' as const }
])
const derivedMenus = computed(() => {
  const idSet = new Set(derivedMenuIds.value)
  return menuSourceList.value.filter((item) => idSet.has(item.id))
})
const blockedMenus = computed(() => {
  const idSet = new Set(blockedMenuIds.value)
  return menuSourceList.value.filter((item) => idSet.has(item.id))
})
watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      loadData()
    }
  }
)

async function loadData() {
  if (!props.teamId) return
  loading.value = true
  try {
    const [allMenus, currentRes, packageRes, originRes] = await Promise.all([
      fetchGetMenuTreeAll(),
      fetchGetTeamMenus(props.teamId),
      fetchGetTeamFeaturePackages(props.teamId),
      fetchGetTeamMenuOrigins(props.teamId)
    ])

    featurePackages.value = packageRes?.packages || []
    derivedMenuIds.value = originRes?.derivedMenuIds || []
    blockedMenuIds.value = originRes?.blockedMenuIds || []
    selectedIds.value = normalizeSelectedMenuIDs(currentRes?.menuIds || [], derivedMenuIds.value)
    derivedSourceMap.value = Object.fromEntries(
      (originRes?.derivedSources || []).map((item) => [item.menuId, item.packageIds])
    )
    selectedDerivedPackageId.value = ''

    const allMenuList = Array.isArray(allMenus) ? allMenus : []
    menuSourceList.value = buildMenuSourceList(allMenuList, derivedMenuIds.value)
    menuTree.value = filterMenuTreeByAllowedIDs(allMenuList, new Set(derivedMenuIds.value))
    await nextTick()
    treeRef.value?.setCheckedKeys(selectedIds.value)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载团队菜单边界失败')
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  visible.value = false
}

function handleCheck(_: any, checkedState: any) {
  const checkedKeys = (checkedState?.checkedKeys || []).map((key: string | number) => String(key))
  selectedIds.value = normalizeSelectedMenuIDs(checkedKeys, derivedMenuIds.value)
}

function toggleExpand() {
  const tree = treeRef.value
  if (!tree?.store?.nodesMap) return
  Object.values(tree.store.nodesMap).forEach((node: any) => {
    node.expanded = !expandAll.value
  })
  expandAll.value = !expandAll.value
}

function checkAll() {
  selectedIds.value = [...derivedMenuIds.value]
  treeRef.value?.setCheckedKeys(selectedIds.value)
}

function clearAll() {
  selectedIds.value = []
  treeRef.value?.setCheckedKeys([])
}

async function handleSave() {
  if (!props.teamId) return
  saving.value = true
  try {
    await fetchSetTeamMenus(props.teamId, normalizeSelectedMenuIDs(selectedIds.value, derivedMenuIds.value))
    ElMessage.success('团队菜单边界已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存团队菜单边界失败')
  } finally {
    saving.value = false
  }
}

function normalizeSelectedMenuIDs(selected: string[], derived: string[]) {
  const allowed = new Set(derived)
  const result: string[] = []
  const seen = new Set<string>()
  selected.forEach((id) => {
    if (!allowed.has(id) || seen.has(id)) return
    seen.add(id)
    result.push(id)
  })
  return result
}

function buildMenuSourceList(source: AppRouteRecord[], allowedIDs: string[]) {
  const indexMap: Record<string, { id: string; label: string }> = {}
  const walk = (items: AppRouteRecord[]) => {
    items.forEach((item: any) => {
      indexMap[item.id] = {
        id: item.id,
        label: formatMenuTitle(item.meta?.title) || item.label || item.name || item.path || item.id
      }
      if (Array.isArray(item.children) && item.children.length) {
        walk(item.children)
      }
    })
  }
  walk(source)
  return allowedIDs.map((id) => indexMap[id]).filter(Boolean)
}

function filterMenuTreeByAllowedIDs(source: AppRouteRecord[], allowed: Set<string>): AppRouteRecord[] {
  if (!allowed.size) return []
  return source
    .map((item: any) => {
      const children = filterMenuTreeByAllowedIDs(item.children || [], allowed)
      if (!allowed.has(item.id) && children.length === 0) return null
      return {
        ...item,
        children
      }
    })
    .filter(Boolean) as AppRouteRecord[]
}

function goToFeaturePackagePage(item: Api.SystemManage.FeaturePackageItem) {
  router.push({
    name: 'FeaturePackage',
    query: {
      packageKey: item.packageKey,
      contextType: item.contextType || 'team',
      open: 'menus'
    }
  })
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

.package-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f8fafc;
}

.package-title {
  color: #475569;
  font-size: 13px;
}

.package-help {
  color: #64748b;
  font-size: 12px;
  line-height: 1.5;
}

.package-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tree-shell {
  padding: 12px 14px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  max-height: 60vh;
  overflow: auto;
}
</style>
