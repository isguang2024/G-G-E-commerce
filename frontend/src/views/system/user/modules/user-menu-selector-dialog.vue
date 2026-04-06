<template>
  <ElDrawer
    v-model="visible"
    :title="`用户菜单裁剪 - ${userTitle}`"
    size="960px"
    destroy-on-close
    class="user-menu-dialog config-drawer"
    direction="rtl"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        这里配置的是平台用户菜单减法裁剪，只影响平台上下文菜单显示，不包含协作空间内菜单。请先绑定平台功能包；此处只会在功能包展开菜单范围内隐藏个别入口，不会额外新增菜单能力。
      </div>

      <div
        class="audit-banner"
        :class="hasPackageConfig ? 'audit-banner--success' : 'audit-banner--warning'"
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
      <ElButton type="primary" :disabled="!hasPackageConfig" :loading="saving" @click="handleSave">
        保存
      </ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import type { AppRouteRecord } from '@/types/router'
  import { formatMenuTitle } from '@/utils/router'
  import {
    fetchGetMenuTreeAll,
    fetchGetUserMenus,
    fetchGetUserPackages,
    fetchSetUserMenus
  } from '@/api/system-manage'
  import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'

  interface Props {
    modelValue: boolean
    userData?: Api.SystemManage.UserListItem
    appKey?: string
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
  const candidateMenuIds = ref<string[]>([])
  const hiddenMenuIds = ref<string[]>([])
  const hasPackageConfig = ref(false)
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())

  const userTitle = computed(() => props.userData?.nickName || props.userData?.userName || '')
  const summaryItems = computed(() => [
    { label: '用户', value: userTitle.value || '-' },
    { label: '当前显示', value: selectedIds.value.length, type: 'success' as const },
    { label: '候选', value: candidateMenuIds.value.length, type: 'info' as const },
    { label: '功能包', value: featurePackages.value.length },
    { label: '已隐藏', value: hiddenMenuIds.value.length, type: 'primary' as const }
  ])
  const defaultProps = {
    children: 'children',
    label: (data: any) =>
      String(formatMenuTitle(data?.meta?.title) || data?.name || data?.path || data?.id || '')
  }

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
    if (!userId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    loading.value = true
    try {
      const [allMenus, menuRes, packageRes] = await Promise.all([
        fetchGetMenuTreeAll(undefined, currentAppKey.value),
        fetchGetUserMenus(userId, currentAppKey.value),
        fetchGetUserPackages(userId, currentAppKey.value)
      ])

      featurePackages.value = packageRes?.packages || []
      candidateMenuIds.value = menuRes?.available_menu_ids || []
      hiddenMenuIds.value = menuRes?.hidden_menu_ids || []
      selectedIds.value = normalizeSelectedMenuIDs(menuRes?.menu_ids || [], candidateMenuIds.value)
      hasPackageConfig.value = Boolean(menuRes?.has_package_config)

      const allMenuList = Array.isArray(allMenus) ? allMenus : []
      menuTree.value = filterMenuTreeByAllowedIDs(allMenuList, new Set(candidateMenuIds.value))
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
    selectedIds.value = normalizeSelectedMenuIDs(checkedKeys, candidateMenuIds.value)
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
    selectedIds.value = [...candidateMenuIds.value]
    treeRef.value?.setCheckedKeys(selectedIds.value)
  }

  function clearAll() {
    selectedIds.value = []
    treeRef.value?.setCheckedKeys([])
  }

  async function handleSave() {
    const userId = props.userData?.id
    if (!userId || !hasPackageConfig.value || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    saving.value = true
    try {
      await fetchSetUserMenus(
        userId,
        normalizeSelectedMenuIDs(selectedIds.value, candidateMenuIds.value),
        currentAppKey.value
      )
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

  function filterMenuTreeByAllowedIDs(
    source: AppRouteRecord[],
    allowed: Set<string>
  ): AppRouteRecord[] {
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
    if (!currentAppKey.value) return
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

  .audit-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 12px;
    font-size: 13px;
  }

  .audit-banner--success {
    background: #ecfdf5;
    color: #047857;
  }

  .audit-banner--warning {
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
    gap: 12px;
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
