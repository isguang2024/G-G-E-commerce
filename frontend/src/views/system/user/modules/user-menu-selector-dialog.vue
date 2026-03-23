<template>
  <ElDialog
    v-model="visible"
    :title="`用户菜单裁剪 - ${userTitle}`"
    width="960px"
    destroy-on-close
    class="user-menu-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        这里配置的是平台用户菜单减法裁剪。请先绑定平台功能包；此处只会在功能包展开菜单范围内隐藏个别入口，不会额外新增菜单能力。
      </div>

      <div
        class="compat-banner"
        :class="hasPackageConfig ? 'compat-banner--success' : 'compat-banner--warning'"
      >
        <span>
          {{
            hasPackageConfig
              ? '当前用户已进入功能包约束模式：这里只能对已绑定功能包展开范围内的菜单做个人隐藏。'
              : '当前用户尚未绑定功能包，无法配置菜单裁剪，请先绑定平台功能包。'
          }}
        </span>
        <ElButton v-if="!hasPackageConfig" type="warning" text @click="emit('open-packages')">
          前往绑定功能包
        </ElButton>
      </div>

      <PermissionSummaryTags :items="summaryItems" />

      <div v-if="featurePackages.length" class="package-card">
        <div class="package-title">当前生效功能包</div>
        <div class="package-tags">
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

      <PermissionSourcePanels
        v-model="selectedDerivedPackageId"
        :packages="featurePackages"
        :source-map="derivedSourceMap"
        :derived-items="derivedMenus"
        :blocked-items="blockedMenus"
        derived-title="功能包展开菜单"
        blocked-title="当前用户已隐藏菜单"
        open="menus"
        blocked-tag-type="primary"
        filtered-blocked-empty-text="当前筛选下暂无用户显式隐藏菜单"
        empty-title="当前暂无用户菜单来源"
        empty-text="请先检查用户直绑功能包、平台角色功能包或平台用户快照是否已经生成。"
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
      <ElButton :disabled="!hasPackageConfig" @click="checkAll">全部保留</ElButton>
      <ElButton :disabled="!hasPackageConfig" @click="clearAll">全部隐藏</ElButton>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton
        type="primary"
        :disabled="!hasPackageConfig"
        :loading="saving"
        @click="handleSave"
      >
        保存
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import type { AppRouteRecord } from '@/types/router'
import { formatMenuTitle } from '@/utils/router'
import { fetchGetMenuTreeAll, fetchGetUserMenus, fetchGetUserPackages, fetchSetUserMenus } from '@/api/system-manage'
import PermissionSourcePanels from '@/components/business/permission/PermissionSourcePanels.vue'
import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'

interface Props {
  modelValue: boolean
  userData?: Api.SystemManage.UserListItem
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
  (e: 'open-packages'): void
}>()
const router = useRouter()

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
const availableMenuIds = ref<string[]>([])
const derivedMenuIds = ref<string[]>([])
const hiddenMenuIds = ref<string[]>([])
const derivedSourceMap = ref<Record<string, string[]>>({})
const menuSourceList = ref<Array<{ id: string; label: string }>>([])
const selectedDerivedPackageId = ref('')
const hasPackageConfig = ref(false)

const userTitle = computed(() => props.userData?.nickName || props.userData?.userName || '')
const summaryItems = computed(() => [
  { label: '用户', value: userTitle.value || '-' },
  { label: '当前显示', value: selectedIds.value.length, type: 'success' as const },
  { label: '候选', value: availableMenuIds.value.length, type: 'info' as const },
  { label: '功能包', value: featurePackages.value.length },
  { label: '功能包展开', value: derivedMenuIds.value.length, type: 'warning' as const },
  { label: '已隐藏', value: hiddenMenuIds.value.length, type: 'primary' as const }
])
const defaultProps = {
  children: 'children',
  label: (data: any) =>
    String(formatMenuTitle(data?.meta?.title) || data?.name || data?.path || data?.id || '')
}

const derivedMenus = computed(() => {
  const idSet = new Set(derivedMenuIds.value)
  return menuSourceList.value.filter((item) => idSet.has(item.id))
})
const blockedMenus = computed(() => {
  const idSet = new Set(hiddenMenuIds.value)
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
  const userId = props.userData?.id
  if (!userId) return
  loading.value = true
  try {
    const [allMenus, menuRes, packageRes] = await Promise.all([
      fetchGetMenuTreeAll(),
      fetchGetUserMenus(userId),
      fetchGetUserPackages(userId)
    ])

    featurePackages.value = packageRes?.packages || []
    availableMenuIds.value = menuRes?.availableMenuIds || []
    derivedMenuIds.value = menuRes?.availableMenuIds || []
    hiddenMenuIds.value = menuRes?.hiddenMenuIds || []
    selectedIds.value = normalizeSelectedMenuIDs(menuRes?.menuIds || [], derivedMenuIds.value)
    derivedSourceMap.value = Object.fromEntries(
      (menuRes?.derivedSources || []).map((item: { menuId: string; packageIds: string[] }) => [
        item.menuId,
        item.packageIds
      ])
    )
    hasPackageConfig.value = Boolean(menuRes?.hasPackageConfig)
    selectedDerivedPackageId.value = ''

    const allMenuList = Array.isArray(allMenus) ? allMenus : []
    menuSourceList.value = buildMenuSourceList(allMenuList, derivedMenuIds.value)
    menuTree.value = filterMenuTreeByAllowedIDs(allMenuList, new Set(derivedMenuIds.value))
    await nextTick()
    treeRef.value?.setCheckedKeys(selectedIds.value)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载用户菜单裁剪失败')
  } finally {
    loading.value = false
  }
}

function handleCheck(_: any, checkedState: any) {
  const checkedKeys = (checkedState?.checkedKeys || []).map((key: string | number) => String(key))
  selectedIds.value = normalizeSelectedMenuIDs(checkedKeys, derivedMenuIds.value)
}

function handleCancel() {
  visible.value = false
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
  const userId = props.userData?.id
  if (!userId || !hasPackageConfig.value) return
  saving.value = true
  try {
    await fetchSetUserMenus(userId, normalizeSelectedMenuIDs(selectedIds.value, derivedMenuIds.value))
    ElMessage.success('用户菜单裁剪已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存用户菜单裁剪失败')
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
      contextType: item.contextType || 'platform',
      open: 'menus'
    }
  })
}
</script>

<style scoped lang="scss">
.dialog-shell {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dialog-note,
.package-title {
  color: #4b5563;
  line-height: 1.6;
}

.compat-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  font-size: 13px;
}

.compat-banner--success {
  background: #ecfdf5;
  color: #047857;
}

.compat-banner--warning {
  background: #fff7ed;
  color: #c2410c;
}

.package-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.package-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
}

.tree-shell {
  max-height: 70vh;
  overflow: auto;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
}
</style>
