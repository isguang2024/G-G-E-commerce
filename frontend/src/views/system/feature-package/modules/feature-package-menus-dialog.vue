<template>
  <ElDrawer
    v-model="visible"
    :title="`功能包绑定菜单 - ${packageName}`"
    size="1100px"
    direction="rtl"
    destroy-on-close
    class="config-drawer"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="summary-tags">
        <ElTag effect="plain" round>功能包 {{ packageName }}</ElTag>
        <ElTag type="warning" effect="plain" round>适用空间 {{ formatWorkspaceScope(workspaceScope) }}</ElTag>
        <ElTag type="success" effect="plain" round>已保留 {{ expandedSelectedMenuIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>树节点 {{ menuNodeCount }}</ElTag>
      </div>

      <div class="cascader-toolbar">
        <AppKeySelect
          v-model="selectedAppKey"
          placeholder="选择 App"
          class="toolbar-select"
          :eager="false"
          @change="handleAppChange"
        />
        <ElSelect
          v-model="selectedMenuSpaceKey"
          filterable
          clearable
          placeholder="全部菜单空间"
          class="toolbar-select"
          @change="handleMenuSpaceChange"
        >
          <ElOption label="全部菜单空间" value="" />
          <ElOption
            v-for="space in menuSpaceOptions"
            :key="space.spaceKey"
            :label="`${space.name} (${space.spaceKey})`"
            :value="space.spaceKey"
          />
        </ElSelect>
        <ElInput
          v-model="menuKeyword"
          clearable
          placeholder="搜索菜单标题或路由"
          :prefix-icon="Search"
          class="toolbar-search"
        />
        <span class="toolbar-switch">
          <span class="toolbar-switch__label">显示隐藏</span>
          <ElSwitch v-model="showHiddenMenus" size="small" />
        </span>
        <span class="toolbar-switch">
          <span class="toolbar-switch__label">显示内嵌</span>
          <ElSwitch v-model="showIframeMenus" size="small" />
        </span>
        <span class="toolbar-switch">
          <span class="toolbar-switch__label">显示启用</span>
          <ElSwitch v-model="showEnabledMenus" size="small" />
        </span>
        <span class="toolbar-switch">
          <span class="toolbar-switch__label">显示路径</span>
          <ElSwitch v-model="showMenuPath" size="small" />
        </span>
      </div>

      <div class="cascader-card">
        <ElCascaderPanel
          ref="menuPanelRef"
          v-model="selectedMenuNodeValues"
          :options="filteredMenuOptions"
          :props="menuCascaderProps"
          class="permission-cascader"
        >
          <template #default="{ node, data }">
            <div class="panel-node" :class="{ 'is-leaf': node.isLeaf }">
              <div class="panel-node__main">
                <span class="panel-node__label">{{ data.label }}</span>
                <span v-if="showMenuPath && data.path" class="panel-node__meta">{{
                  data.path
                }}</span>
              </div>
              <ElTag
                v-if="!node.isLeaf"
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
        <span class="footer-note">
          菜单空间可为空，留空时默认显示当前 App 中的全部菜单；保存时按当前 App 与菜单选择共同生效。
        </span>
        <ElButton text @click="clearMenuSelection">清空选择</ElButton>
      </div>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue'
  import { Search } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import type { CascaderOption, CascaderProps } from 'element-plus'
  import AppKeySelect from '@/components/business/app/AppKeySelect.vue'
  import {
    fetchGetMenuSpaces,
    fetchGetMenuTreeAll,
    fetchGetFeaturePackageMenus,
    fetchSetFeaturePackageMenus
  } from '@/domains/governance/api'
  import { formatMenuTitle } from '@/utils/router'

  interface Props {
    modelValue: boolean
    packageId: string
    packageName: string
    workspaceScope?: 'all' | 'personal' | 'collaboration' | string
    appKey?: string
    appKeys?: string[]
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    packageId: '',
    packageName: '',
    workspaceScope: 'all',
    appKey: '',
    appKeys: () => []
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  interface RawMenuNode {
    id: string
    label?: string
    name?: string
    path?: string
    meta?: {
      title?: string
      isHide?: boolean
      isIframe?: boolean
      isEnable?: boolean
    }
    isHide?: boolean
    isIframe?: boolean
    isEnable?: boolean
    children?: RawMenuNode[]
  }

  interface MenuOption extends CascaderOption {
    path?: string
    isHide?: boolean
    isIframe?: boolean
    isEnable?: boolean
    totalLeafCount?: number
    selectedLeafCount?: number
  }

  const menuPanelRef = ref<any>()
  const loading = ref(false)
  const saving = ref(false)
  const menuSpaceOptions = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const selectedAppKey = ref('')
  const selectedMenuSpaceKey = ref('')
  const menuTreeData = ref<RawMenuNode[]>([])
  const selectedMenuNodeValues = ref<string[]>([])
  const menuKeyword = ref('')
  const showHiddenMenus = ref(true)
  const showIframeMenus = ref(true)
  const showEnabledMenus = ref(true)
  const showMenuPath = ref(false)

  const menuCascaderProps: CascaderProps = {
    multiple: true,
    emitPath: false,
    checkStrictly: false,
    expandTrigger: 'click',
    checkOnClickNode: false,
    checkOnClickLeaf: false,
    showPrefix: true
  }

  const selectedMenuIdSet = computed(
    () => new Set(selectedMenuNodeValues.value.map((item) => `${item}`))
  )

  const menuOptions = computed<MenuOption[]>(() =>
    normalizeMenuOptions(menuTreeData.value, selectedMenuIdSet.value)
  )

  const menuBranchMap = computed(() => {
    const map = new Map<string, string[]>()
    const visit = (nodes: MenuOption[]): string[] => {
      const collected: string[] = []
      nodes.forEach((node) => {
        const childIds = visit((node.children || []) as MenuOption[])
        const current = [`${node.value}`]
        const all = current.concat(childIds)
        map.set(`${node.value}`, all)
        collected.push(...all)
      })
      return collected
    }
    visit(menuOptions.value)
    return map
  })

  const expandedSelectedMenuIds = computed(() =>
    expandSelectedValues(selectedMenuNodeValues.value, menuBranchMap.value)
  )

  const filteredMenuOptions = computed(() => {
    const keyword = menuKeyword.value.trim().toLowerCase()
    return filterNestedOptions(menuOptions.value, (node) => {
      if (!node.leaf) return !keyword
      if (!showHiddenMenus.value && node.isHide) return false
      if (!showIframeMenus.value && node.isIframe) return false
      if (!showEnabledMenus.value && node.isEnable !== false) return false
      if (keyword && !`${node.label || ''} ${node.path || ''}`.toLowerCase().includes(keyword)) {
        return false
      }
      return true
    })
  })

  const menuNodeCount = computed(() => flattenMenuOptionIds(filteredMenuOptions.value).length)

  watch(
    () => props.modelValue,
    async (open) => {
      if (open) {
        await initializeContext()
        await loadData()
      }
    }
  )

  watch(
    () => [props.appKey, props.appKeys],
    async () => {
      if (!visible.value) {
        return
      }
      await initializeContext(true)
      await loadData()
    },
    { deep: true }
  )

  async function initializeContext(force = false) {
    if (force) {
      selectedAppKey.value = ''
      selectedMenuSpaceKey.value = ''
    }
    if (!selectedAppKey.value) {
      selectedAppKey.value = resolveInitialAppKey()
    }
    if (!selectedMenuSpaceKey.value) {
      selectedMenuSpaceKey.value = ''
    }
    if (selectedAppKey.value) {
      await loadMenuSpaces()
    }
  }

  function resolveInitialAppKey() {
    const explicit = `${props.appKey || ''}`.trim()
    if (explicit) return explicit
    const scoped = props.appKeys?.map((item) => `${item || ''}`.trim()).filter(Boolean) || []
    if (scoped.length > 0) return scoped[0]
    return ''
  }

  async function loadMenuSpaces() {
    if (!selectedAppKey.value) {
      menuSpaceOptions.value = []
      return
    }
    try {
      const result = await fetchGetMenuSpaces(selectedAppKey.value)
      menuSpaceOptions.value = result.records || []
    } catch {
      menuSpaceOptions.value = []
    }
  }

  async function loadData() {
    if (!props.packageId) {
      return
    }
    if (!selectedAppKey.value) {
      await initializeContext()
    }
    if (!selectedAppKey.value) {
      ElMessage.warning('请选择 App')
      return
    }
    loading.value = true
    try {
      const [menus, assigned] = await Promise.all([
        fetchGetMenuTreeAll(selectedMenuSpaceKey.value || undefined, selectedAppKey.value),
        fetchGetFeaturePackageMenus(props.packageId, selectedAppKey.value)
      ])
      menuTreeData.value = sanitizeMenuTree(menus)
      selectedMenuNodeValues.value = (assigned?.menu_ids || []).map(
        (item: string | number) => `${item}`
      )
      await nextTick()
      ensureExpandedMenus(menuPanelRef.value, selectedMenuNodeValues.value)
    } catch (error: any) {
      ElMessage.error(error?.message || '加载功能包绑定菜单失败')
    } finally {
      loading.value = false
    }
  }

  async function handleSave() {
    if (!props.packageId) {
      return
    }
    if (!selectedAppKey.value) {
      ElMessage.warning('请选择 App')
      return
    }
    saving.value = true
    try {
      const menuIds = expandedSelectedMenuIds.value
      const stats = await fetchSetFeaturePackageMenus(props.packageId, menuIds, selectedAppKey.value)
      ElMessage.success(formatRefreshMessage(stats))
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存功能包绑定菜单失败')
    } finally {
      saving.value = false
    }
  }

  async function handleAppChange() {
    selectedMenuSpaceKey.value = ''
    await loadMenuSpaces()
    await loadData()
  }

  async function handleMenuSpaceChange() {
    await loadData()
  }

  function formatWorkspaceScope(workspaceScope?: string) {
    if (workspaceScope === 'all' || workspaceScope === 'common') return '所有空间'
    if (workspaceScope === 'personal') return '个人空间'
    if (workspaceScope === 'collaboration') return '协作空间'
    return workspaceScope || '-'
  }

  function clearMenuSelection() {
    selectedMenuNodeValues.value = []
  }

  function formatRefreshMessage(stats?: Api.SystemManage.RefreshStats) {
    return `本次增量刷新：角色 ${stats?.roleCount || 0}、协作空间 ${stats?.collaborationWorkspaceCount || 0}、用户 ${stats?.userCount || 0}、耗时 ${stats?.elapsedMilliseconds || 0} ms`
  }

  function normalizeMenuOptions(items: RawMenuNode[], selectedSet: Set<string>): MenuOption[] {
    return items.map((item) => {
      const children = normalizeMenuOptions(item.children || [], selectedSet)
      const title = formatMenuTitle(item.meta?.title || '') || item.label || item.name || ''
      return {
        value: `${item.id}`,
        label: title,
        path: item.path || '',
        isHide: item.meta?.isHide ?? item.isHide,
        isIframe: item.meta?.isIframe ?? item.isIframe,
        isEnable: item.meta?.isEnable ?? item.isEnable,
        leaf: !item.children?.length,
        totalLeafCount: countMenuLeaves(item),
        selectedLeafCount: countSelectedMenuLeaves(item, selectedSet),
        children
      }
    })
  }

  function countMenuLeaves(node: RawMenuNode): number {
    if (!node.children?.length) return 1
    return node.children.reduce((sum, child) => sum + countMenuLeaves(child), 0)
  }

  function countSelectedMenuLeaves(node: RawMenuNode, selectedSet: Set<string>): number {
    if (!node.children?.length) return selectedSet.has(`${node.id}`) ? 1 : 0
    return node.children.reduce(
      (sum, child) => sum + countSelectedMenuLeaves(child, selectedSet),
      0
    )
  }

  function filterNestedOptions<T extends CascaderOption>(
    items: T[],
    predicate: (node: T) => boolean
  ): T[] {
    return items
      .map((item) => {
        const children = filterNestedOptions(((item.children || []) as T[]) || [], predicate)
        const passed = predicate(item)
        if (!passed && !children.length) return null
        return {
          ...item,
          children
        } as T
      })
      .filter((item): item is T => Boolean(item))
  }

  function flattenMenuOptionIds(items: MenuOption[]): string[] {
    return items.flatMap((item) => [
      `${item.value}`,
      ...flattenMenuOptionIds((item.children || []) as MenuOption[])
    ])
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
    let rootNode = rootMenus[0]
    let childNode = rootNode?.children?.[0]
    if (firstValue) {
      const matchedNode = panel
        .getFlattedNodes?.(false)
        ?.find((node: any) => `${node?.value}` === `${firstValue}`)
      const pathNodes = matchedNode?.pathNodes || []
      if (pathNodes[0]) rootNode = pathNodes[0]
      if (pathNodes[1]) childNode = pathNodes[1]
    }
    const nextMenus = [rootMenus]
    if (rootNode?.children?.length) nextMenus.push(rootNode.children)
    if (childNode?.children?.length) nextMenus.push(childNode.children)
    panel.menus = nextMenus
  }

  function sanitizeMenuTree(source: unknown): RawMenuNode[] {
    if (!Array.isArray(source)) return []
    return source
      .map((item: any) => {
        const rawId = item?.id
        if (rawId === undefined || rawId === null || rawId === '') return null
        return {
          id: `${rawId}`,
          label: item?.label,
          name: item?.name,
          path: item?.path,
          meta: item?.meta,
          isHide: item?.isHide,
          isIframe: item?.isIframe,
          isEnable: item?.isEnable,
          children: sanitizeMenuTree(item?.children)
        } as RawMenuNode
      })
      .filter((item): item is RawMenuNode => Boolean(item))
  }
</script>

<style scoped lang="scss">
  .dialog-shell {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .summary-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .cascader-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px;
  }

  .toolbar-select {
    width: 260px;
  }

  .toolbar-search {
    width: 260px;
  }

  .toolbar-switch {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #4b5563;
  }

  .toolbar-switch__label {
    white-space: nowrap;
  }

  .cascader-card {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    padding: 8px;
    min-height: 520px;
  }

  .permission-cascader {
    width: 100%;
  }

  :deep(.permission-cascader .el-cascader-menu) {
    width: 33.333%;
    min-width: 280px;
  }

  :deep(.permission-cascader .el-cascader-menu__wrap) {
    height: 500px;
  }

  .panel-node {
    width: 100%;
    display: inline-flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .panel-node__main {
    min-width: 0;
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .panel-node__label {
    color: #111827;
    font-weight: 500;
  }

  .panel-node__meta {
    color: #9ca3af;
    font-size: 12px;
  }

  .panel-node__count {
    flex-shrink: 0;
  }

  .cascader-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .footer-note {
    color: #6b7280;
    font-size: 13px;
  }
</style>
