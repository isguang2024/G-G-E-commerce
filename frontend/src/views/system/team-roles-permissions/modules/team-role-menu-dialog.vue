<template>
  <ElDrawer v-model="visible" :title="`团队角色菜单裁剪 - ${roleTitle}`" size="620px" @close="handleClose"
    direction="rtl"
    class="config-drawer">
    <div class="dialog-shell">
      <div class="dialog-note">
        {{
          props.roleData?.isGlobal
            ? '基础团队角色默认继承当前团队功能包的菜单范围，这里只读查看最终角色菜单结果。'
            : '请先绑定角色功能包。这里只展示当前角色功能包展开范围内可裁剪的菜单结果。'
        }}
      </div>
      <PermissionSummaryTags :items="summaryItems" />
      <div v-if="featurePackages.length" class="package-tags">
        <ElTag v-for="item in featurePackages" :key="item.id" type="success" effect="plain" round>
          {{ item.name }}
        </ElTag>
      </div>
      <PermissionSourcePanels
        v-model="selectedDerivedPackageId"
        :packages="featurePackages"
        :source-map="derivedSourceMap"
        :derived-items="derivedMenuItems"
        :blocked-items="blockedMenuItems"
        derived-title="功能包展开菜单"
        blocked-title="当前角色已屏蔽菜单"
        open="menus"
        filtered-blocked-empty-text="当前筛选下暂无角色显式屏蔽菜单"
        empty-title="当前暂无角色菜单来源"
        empty-text="请先为角色绑定功能包，或检查当前团队快照是否已经刷新。"
      />
      <ElScrollbar height="70vh">
      <ElTree
        ref="treeRef"
        :data="menuList"
        show-checkbox
        node-key="id"
        :default-expand-all="expandAll"
        :props="defaultProps"
        @check="handleCheck"
      />
      </ElScrollbar>
    </div>
    <template #footer>
      <ElButton @click="toggleExpand">{{ expandAll ? '全部收起' : '全部展开' }}</ElButton>
      <ElButton v-if="!props.roleData?.isGlobal" @click="checkAll">全部保留</ElButton>
      <ElButton v-if="!props.roleData?.isGlobal" @click="clearAll">全部屏蔽</ElButton>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton v-if="!props.roleData?.isGlobal" type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { ElButton, ElMessage } from 'element-plus'
  import PermissionSourcePanels from '@/components/business/permission/PermissionSourcePanels.vue'
  import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'
  import { fetchGetMenuTreeAll } from '@/api/system-manage'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import {
    fetchGetMyTeamBoundaryRoleMenus,
    fetchGetMyTeamBoundaryRolePackages,
    fetchSetMyTeamBoundaryRoleMenus
  } from '@/api/team'
  import { formatMenuTitle } from '@/utils/router'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void; (e: 'success'): void }>()

  const treeRef = ref()
  const menuSpaceStore = useMenuSpaceStore()
  const { currentSpaceKey } = storeToRefs(menuSpaceStore)
  const expandAll = ref(true)
  const saving = ref(false)
  const menuList = ref<any[]>([])
  const menuSourceList = ref<Array<{ id: string; label: string }>>([])
  const availableMenuIds = ref<string[]>([])
  const selectedMenuIds = ref<string[]>([])
  const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const derivedSourceMap = ref<Record<string, string[]>>({})
  const selectedDerivedPackageId = ref('')
  const inherited = ref(false)
  const roleTitle = computed(() => props.roleData?.roleName || '')
  const checkedCount = computed(() => selectedMenuIds.value.length)
  const blockedCount = computed(() => Math.max(availableMenuIds.value.length - checkedCount.value, 0))
  const summaryItems = computed(() => [
    { label: '角色', value: roleTitle.value || '-' },
    { label: '功能包', value: featurePackages.value.length, type: 'success' as const },
    { label: '继承模式', value: inherited.value ? '继承团队功能包' : '角色独立功能包', type: 'primary' as const },
    { label: '可裁剪菜单', value: availableMenuIds.value.length, type: 'warning' as const },
    { label: '已保留', value: checkedCount.value, type: 'success' as const },
    { label: '已屏蔽', value: blockedCount.value, type: 'danger' as const }
  ])
  const blockedMenus = computed(() => {
    const selectedIdSet = new Set(selectedMenuIds.value)
    return menuSourceList.value.filter((item) => !selectedIdSet.has(item.id))
  })
  const derivedMenuItems = computed(() => menuSourceList.value)
  const blockedMenuItems = computed(() => blockedMenus.value)

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const defaultProps = {
    children: 'children',
    label: (data: any) => formatMenuTitle(data.meta?.title) || data.label || data.name || ''
  }

  watch(
    () => props.modelValue,
    async (open) => {
      if (!open || !props.roleData?.roleId) return
      try {
        const [menus, assigned, packagesRes] = await Promise.all([
          fetchGetMenuTreeAll(currentSpaceKey.value),
          fetchGetMyTeamBoundaryRoleMenus(props.roleData.roleId),
          fetchGetMyTeamBoundaryRolePackages(props.roleData.roleId)
        ])
        availableMenuIds.value = assigned?.available_menu_ids || []
        featurePackages.value = packagesRes?.packages || []
        derivedSourceMap.value = Object.fromEntries(
          (assigned?.derived_sources || []).map((item) => [item.menu_id, item.package_ids])
        )
        selectedDerivedPackageId.value = ''
        inherited.value = Boolean(packagesRes?.inherited)
        menuSourceList.value = buildMenuSourceList(Array.isArray(menus) ? menus : [], availableMenuIds.value)
        menuList.value = filterMenuTreeByAllowedIds(Array.isArray(menus) ? menus : [], new Set(availableMenuIds.value))
        await nextTick()
        selectedMenuIds.value = [...(assigned?.menu_ids || [])]
        treeRef.value?.setCheckedKeys(selectedMenuIds.value)
      } catch (error: any) {
        ElMessage.error(error?.message || '加载团队角色菜单裁剪失败')
      }
    }
  )

  function toggleExpand() {
    const tree = treeRef.value
    if (!tree?.store?.nodesMap) return
    Object.values(tree.store.nodesMap).forEach((node: any) => {
      node.expanded = !expandAll.value
    })
    expandAll.value = !expandAll.value
  }

  function handleClose() {
    visible.value = false
    selectedMenuIds.value = []
    selectedDerivedPackageId.value = ''
    treeRef.value?.setCheckedKeys([])
  }

  function handleCheck(_: any, checkedState: any) {
    selectedMenuIds.value = (checkedState?.checkedKeys || []).map((key: string | number) => String(key))
  }

  function checkAll() {
    selectedMenuIds.value = [...availableMenuIds.value]
    treeRef.value?.setCheckedKeys(selectedMenuIds.value)
  }

  function clearAll() {
    selectedMenuIds.value = []
    treeRef.value?.setCheckedKeys([])
  }

  async function handleSave() {
    if (!props.roleData?.roleId) return
    saving.value = true
    try {
      await fetchSetMyTeamBoundaryRoleMenus(props.roleData.roleId, selectedMenuIds.value)
      ElMessage.success('团队角色菜单裁剪已保存')
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error?.message || '保存团队角色菜单裁剪失败')
    } finally {
      saving.value = false
    }
  }

  function filterMenuTreeByAllowedIds(source: any[], allowed: Set<string>): any[] {
    if (!allowed.size) return []
    return source
      .map((item: any) => {
        const children: any[] = filterMenuTreeByAllowedIds(item.children || [], allowed)
        if (!allowed.has(item.id) && children.length === 0) return null
        return {
          ...item,
          children
        }
      })
      .filter(Boolean) as any[]
  }

  function buildMenuSourceList(source: any[], allowedIDs: string[]) {
    const indexMap: Record<string, { id: string; label: string }> = {}
    const walk = (items: any[]) => {
      items.forEach((item) => {
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
</script>

<style scoped lang="scss">
  .dialog-shell {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .dialog-note {
    color: #6b7280;
    line-height: 1.6;
  }

  .package-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .source-tags,
  .package-filter-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
