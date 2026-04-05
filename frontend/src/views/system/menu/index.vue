<!-- 菜单管理页面 -->
<template>
  <div class="menu-page art-full-height">
    <div class="menu-top-stack">
      <!-- 搜索栏 -->
      <MenuSearch
        v-show="showSearchBar"
        v-model="formFilters"
        @reset="handleReset"
        @search="handleSearch"
      />

      <AdminWorkspaceHero
        class="menu-hero"
        :title="menuPageTitle"
        :description="menuPageDescription"
        :metrics="menuHeroMetrics"
      >
        <div class="menu-hero-actions">
          <ElSelect
            v-if="isLayoutMode"
            v-model="activeSpaceKey"
            class="menu-space-select"
            filterable
            @change="handleSpaceChange"
          >
            <ElOption
              v-for="item in menuSpaceOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
          <ElButton
            v-action="'system.menu.manage'"
            type="primary"
            @click="handleAddMenu"
            v-ripple
          >
            {{ isLayoutMode ? '创建布局菜单' : '创建菜单定义' }}
          </ElButton>
          <ElButton v-if="isLayoutMode" @click="goToDefinitionManagement" v-ripple>
            返回定义管理
          </ElButton>
          <ElDropdown @command="handleMoreActionCommand">
            <ElButton v-ripple>更多操作</ElButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="manageGroup" :disabled="!groupingAvailable">
                  管理分组
                </ElDropdownItem>
                <template v-if="!isLayoutMode">
                  <ElDropdownItem command="backupGlobal">备份定义集</ElDropdownItem>
                  <ElDropdownItem command="backupList">管理定义备份</ElDropdownItem>
                </template>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
        </div>
      </AdminWorkspaceHero>
    </div>

    <ElCard
      class="art-table-card"
      shadow="never"
    >
      <ElAlert
        v-if="loadError"
        class="menu-inline-alert"
        type="info"
        :closable="false"
        show-icon
        :title="loadError"
      />
      <!-- 表格头部 -->
      <ArtTableHeader
        layout="search,refresh,size,fullscreen,columns"
        :showZebra="false"
        :loading="loading"
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
      >
        <template #left>
          <div class="menu-toolbar">
            <div class="menu-toolbar-top">
              <div class="menu-toolbar-tip">
                {{ menuToolbarTip }}
              </div>
              <div v-if="menuGroupApiUnavailable" class="menu-inline-note">
                菜单分组暂不可用，当前按普通菜单树显示
              </div>
            </div>
            <div class="menu-toolbar-bottom">
              <div class="menu-toolbar-switches">
                <span class="menu-switch-item">
                  <span class="menu-switch-label">显示隐藏</span>
                  <ElSwitch v-model="showHiddenMenus" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">显示内嵌</span>
                  <ElSwitch v-model="showIframeMenus" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">显示启用</span>
                  <ElSwitch v-model="showEnabledMenus" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">启用分组</span>
                  <ElSwitch v-model="groupingEnabled" :disabled="!groupingAvailable" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">分组可视</span>
                  <ElSwitch v-model="groupedMenuVisible" :disabled="!groupingAvailable" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">多选模式</span>
                  <ElSwitch v-model="multiSelectEnabled" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">展开菜单</span>
                  <ElSwitch v-model="isExpanded" @change="handleExpandSwitchChange" />
                </span>
              </div>
              <div v-if="multiSelectEnabled" class="menu-toolbar-actions menu-toolbar-batch">
                <span class="menu-batch-count">已选 {{ selectedMenuRows.length }} 项</span>
                <ElDropdown @command="handleBatchCommand">
                  <ElButton type="primary" plain :disabled="selectedMenuRows.length === 0">
                    批量操作
                  </ElButton>
                  <template #dropdown>
                    <ElDropdownMenu>
                      <ElDropdownItem command="assign">移入分组</ElDropdownItem>
                      <ElDropdownItem command="remove">移出分组</ElDropdownItem>
                    </ElDropdownMenu>
                  </template>
                </ElDropdown>
              </div>
            </div>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        class="menu-table"
        :class="{ 'menu-table-multi-disabled': !multiSelectEnabled }"
        :rowKey="rowKey"
        :loading="loading"
        :columns="displayColumns"
        :data="tableData"
        :stripe="false"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="false"
        @selection-change="handleBatchSelectionChange"
      >
        <!-- 菜单名称列 -->
        <template #title="{ row }">
          <template v-if="isManageGroupRow(row)">
            <span class="menu-group-title">{{ formatMenuTitle(row.meta?.title) }}</span>
          </template>
          <template v-else>
            <ArtSvgIcon v-if="row.meta?.icon" :icon="row.meta.icon" class="mr-2 text-g-500" />
            <span>{{ formatMenuTitle(row.meta?.title) }}</span>
          </template>
        </template>

        <!-- 菜单类型列 -->
        <template #type="{ row }">
          <ElTag :type="getMenuTypeTag(row)">{{ getMenuTypeText(row) }}</ElTag>
        </template>

        <!-- 路由列 -->
        <template #path="{ row }">
          <span>{{ isManageGroupRow(row) ? '-' : row.meta?.link || row.path || '' }}</span>
        </template>

        <!-- 组件路径列 -->
        <template #component="{ row }">
          <span class="text-gray-600">{{
            isManageGroupRow(row) ? '-' : row.component || '-'
          }}</span>
        </template>

        <template #linkedPage="{ row }">
          <div v-if="!isManageGroupRow(row)" class="menu-linked-page-cell">
            <template v-if="getLinkedPages(row).length">
              <span class="menu-linked-page-cell__primary">{{ getLinkedPages(row)[0].name }}</span>
              <span class="menu-linked-page-cell__meta">
                {{ getLinkedPages(row)[0].pageKey }}
                <template v-if="getLinkedPages(row).length > 1">
                  · 另有 {{ getLinkedPages(row).length - 1 }} 个受管页面
                </template>
              </span>
            </template>
            <span v-else class="text-gray-400">无受管页面</span>
          </div>
          <span v-else class="text-gray-400">-</span>
        </template>

        <template #space="{ row }">
          <ElTag size="small" effect="plain" type="info">
            {{ getSpaceName(row.spaceKey || row.meta?.spaceKey) }}
          </ElTag>
        </template>

        <!-- 高级配置列 -->
        <template #advanced="{ row }">
          <div v-if="!isManageGroupRow(row)" class="advanced-configs">
            <ElTag
              v-if="isEntryMenuRow(row) && row.meta?.keepAlive"
              size="small"
              effect="light"
              type="primary"
              class="mr-2"
            >
              缓存
            </ElTag>
            <ElTag v-if="row.meta?.isHide" size="small" effect="light" type="warning" class="mr-2">
              隐藏
            </ElTag>
            <ElTag v-if="!isDirectoryMenuRow(row) && row.meta?.isIframe" size="small" effect="light" type="info" class="mr-2">
              内嵌
            </ElTag>
            <ElTag
              v-if="row.meta?.showBadge"
              size="small"
              effect="light"
              type="success"
              class="mr-2"
            >
              徽章
            </ElTag>
            <ElTag v-if="isEntryMenuRow(row) && row.meta?.fixedTab" size="small" effect="light" type="danger" class="mr-2">
              固定
            </ElTag>
            <ElTag
              v-if="isEntryMenuRow(row) && row.meta?.isFullPage"
              size="small"
              effect="light"
              type="primary"
              class="mr-2"
            >
              全屏
            </ElTag>
            <ElTag
              size="small"
              effect="light"
              :type="getAccessModeTag(row.meta?.accessMode)"
              class="mr-2"
            >
              {{ getAccessModeLabel(row.meta?.accessMode) }}
            </ElTag>
            <ElTag
              v-if="
                getMenuActionRequirement(row.meta).actions.length &&
                `${row.meta?.accessMode || 'permission'}` === 'permission'
              "
              size="small"
              effect="light"
              type="info"
              class="mr-2"
            >
              {{ getMenuActionRequirementLabel(row) }}
            </ElTag>
          </div>
          <span v-else class="text-gray-400">-</span>
        </template>

        <!-- 状态列 -->
        <template #status="{ row }">
          <ElTag
            v-if="!isManageGroupRow(row)"
            :type="row.meta?.isEnable !== false ? 'success' : 'info'"
          >
            {{ row.meta?.isEnable !== false ? '启用' : '未启用' }}
          </ElTag>
          <span v-else class="text-gray-400">-</span>
        </template>

        <!-- 操作列 -->
        <template #operation="{ row }">
          <div v-if="!isManageGroupRow(row)" class="flex items-center justify-center gap-2">
            <ArtButtonMore
              :list="getOperationList(row)"
              @click="(item) => handleMenuOperation(item, row)"
            />
          </div>
          <span v-else class="text-gray-400">-</span>
        </template>
      </ArtTable>

      <!-- 菜单弹窗 -->
      <MenuDialog
        v-model:visible="dialogVisible"
        :editData="editData"
        :menuTree="filteredMenuTree"
        :manageGroups="menuGroups"
        :menuSpaces="menuSpaces"
        :currentSpaceKey="activeSpaceKey"
        :currentMenuPages="getLinkedPages(editData || {})"
        :editingMenuId="editData?.id"
        :initialParentId="String(parentRowForAdd?.id ?? '')"
        :showSpaceField="isLayoutMode"
        @submit="handleSubmit"
      />

      <MenuGroupDrawer
        v-model="manageGroupDrawerVisible"
        :items="menuGroups"
        :loading="loading"
        :saving="groupSaving"
        @save="handleSaveManageGroup"
        @delete="handleDeleteManageGroup"
      />

      <MenuPermissionDialog
        v-model="actionRequirementVisible"
        :menuData="actionRequirementData"
        @submit="handleActionRequirementSubmit"
      />

      <MenuBackupDialog
        v-model="backupDialogVisible"
        :loading="backupLoading"
        :scopeType="backupScopeType"
        :currentSpaceName="currentSpaceName"
        :dialog-title="backupDialogTitle"
        :alert-title="backupAlertTitle"
        :alert-description="backupAlertDescription"
        @submit="handleCreateBackup"
      />

      <MenuBackupListDialog
        v-model="backupListDialogVisible"
        :loading="backupLoading"
        :items="backupList"
        :title="backupListTitle"
        :alert-description="backupListAlertDescription"
        :empty-description="backupListEmptyDescription"
        @action="handleBackupListAction"
      />

      <MenuDeleteDialog
        v-model:visible="deleteDialogVisible"
        :loading="deleteLoading"
        :menuTitle="formatMenuTitle(deleteTargetRow?.meta?.title) || deleteTargetRow?.name || ''"
        :childCount="getMenuChildCount(deleteTargetRow)"
        :descendantCount="getMenuDescendantCount(deleteTargetRow)"
        :affectedPageCount="getAffectedPageCount(deleteTargetRow)"
        :affectedRelationCount="deletePreview?.affectedRelationCount || 0"
        :parentOptions="getDeleteParentOptions(deleteTargetRow)"
        @confirm="handleDeleteMenuConfirm"
      />

      <ElDialog
        v-model="batchAssignDialogVisible"
        title="批量移入分组"
        width="460px"
        destroy-on-close
      >
        <div class="menu-batch-dialog">
          <div class="menu-batch-dialog-count"
            >已选 {{ selectedMenuRows.length }} 项，将同步作用于所选菜单及其下级。</div
          >
          <ElSelect
            v-model="batchTargetGroupId"
            filterable
            clearable
            placeholder="请选择目标分组，可搜索"
            style="width: 100%"
          >
            <ElOption
              v-for="item in menuGroups"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </ElSelect>
        </div>
        <template #footer>
          <div class="menu-batch-dialog-footer">
            <ElButton @click="batchAssignDialogVisible = false">取消</ElButton>
            <ElButton type="primary" @click="handleBatchAssignSubmit">确认移入</ElButton>
          </div>
        </template>
      </ElDialog>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, reactive, nextTick, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { formatMenuTitle } from '@/utils/router'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import type { AppRouteRecord } from '@/types/router'
  import MenuDialog from './modules/menu-dialog.vue'
  import MenuGroupDrawer from './modules/menu-group-drawer.vue'
  import MenuBackupDialog from './modules/menu-backup-dialog.vue'
  import MenuBackupListDialog from './modules/menu-backup-list-dialog.vue'
  import MenuDeleteDialog from './modules/menu-delete-dialog.vue'
  import MenuPermissionDialog from './modules/menu-permission-dialog.vue'
  import MenuSearch from './modules/menu-search.vue'
  import {
    fetchGetMenuTreeAll,
    fetchCreateMenu,
    fetchUpdateMenu,
    fetchDeleteMenu,
    fetchGetMenuDeletePreview,
    fetchGetMenuManageGroups,
    fetchCreateMenuManageGroup,
    fetchUpdateMenuManageGroup,
    fetchDeleteMenuManageGroup,
    fetchGetPageOptions,
    fetchGetMenuSpaces,
    fetchCreateMenuBackup,
    fetchGetMenuBackupList,
    fetchDeleteMenuBackup,
    fetchRestoreMenuBackup
  } from '@/api/system-manage'
  import {
    ElTag,
    ElMessageBox,
    ElMessage,
    ElButton,
    ElSwitch,
    ElDropdown,
    ElDropdownMenu,
    ElDropdownItem,
    ElSelect,
    ElOption
  } from 'element-plus'
  import { getMenuActionRequirement } from '@/utils/permission/menu'

  defineOptions({ name: 'Menus' })

  type MenuBackupScopeType = 'space' | 'global'
  type MenuDeleteMode = 'single' | 'cascade' | 'promote_children'

  // --- 状态管理 ---
  const loading = ref(false)
  const loadError = ref('')
  const showSearchBar = ref(false)
  const isExpanded = ref(false)
  const showHiddenMenus = ref(true)
  const showIframeMenus = ref(true)
  const showEnabledMenus = ref(true)
  const groupingEnabled = ref(true)
  const groupedMenuVisible = ref(true)
  const tableRef = ref()
  const multiSelectEnabled = ref(false)
  const rawMenuTree = ref<AppRouteRecord[]>([])
  const rawPages = ref<Api.SystemManage.PageItem[]>([])
  const menuSpaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const activeSpaceKey = ref('')
  const route = useRoute()
  const router = useRouter()
  const { targetAppKey } = useManagedAppScope()
  const managedAppMissingText = '缺少 app 上下文，请先从应用管理选择 App'
  const isLayoutMode = computed(() => `${route.query.layout || ''}`.trim() === '1')
  const menuGroups = ref<Api.SystemManage.MenuManageGroupItem[]>([])
  const dataFromBackend = ref(false)
  const menuGroupApiUnavailable = ref(false)

  // --- 菜单备份相关状态 ---
  const backupLoading = ref(false)
  const backupDialogVisible = ref(false)
  const backupListDialogVisible = ref(false)
  const backupScopeType = ref<MenuBackupScopeType>('space')
  const backupList = ref<Api.SystemManage.MenuBackupItem[]>([])
  const warnDev = (...args: any[]) => {
    if (import.meta.env.DEV) {
      console.info(...args)
    }
  }

  // --- 搜索相关 ---
  const initialSearchState = { name: '', route: '' }
  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })
  // --- 弹窗相关 ---
  const dialogVisible = ref(false)
  const manageGroupDrawerVisible = ref(false)
  const groupSaving = ref(false)
  const editData = ref<any>(null)
  const parentRowForAdd = ref<AppRouteRecord | null>(null)
  const deleteDialogVisible = ref(false)
  const deleteLoading = ref(false)
  const deleteTargetRow = ref<any>(null)
  const deletePreview = ref<Api.SystemManage.MenuDeletePreviewItem | null>(null)
  const actionRequirementVisible = ref(false)
  const actionRequirementData = ref<any>(null)
  const selectedMenuRows = ref<any[]>([])
  const batchTargetGroupId = ref('')
  const batchAssignDialogVisible = ref(false)

  // --- 菜单列表处理 ---
  const normalizeKeyword = (value?: string) => `${value || ''}`.trim().toLowerCase()

  const matchesMenuFilters = (item: AppRouteRecord) => {
    if (!showHiddenMenus.value && item.meta?.isHide) return false
    if (!showIframeMenus.value && item.meta?.isIframe) return false
    if (!showEnabledMenus.value && item.meta?.isEnable !== false) return false
    if (!groupedMenuVisible.value && getManageGroupId(item)) return false
    return true
  }

  const matchesMenuSearch = (item: AppRouteRecord) => {
    const searchName = normalizeKeyword(appliedFilters.name)
    const searchRoute = normalizeKeyword(appliedFilters.route)
    const title = normalizeKeyword(formatMenuTitle(item.meta?.title))
    const path = normalizeKeyword(item.path)
    const titleMatch = !searchName || title.includes(searchName)
    const routeMatch = !searchRoute || path.includes(searchRoute)
    return titleMatch && routeMatch
  }

  const cloneMenuNode = (item: AppRouteRecord, children: AppRouteRecord[]): AppRouteRecord => ({
    ...item,
    meta: item.meta ? { ...item.meta } : item.meta,
    children
  })

  const isManageGroupRow = (item: any) => Boolean(item?.meta?.__manageGroupNode)

  const menuGroupMap = computed(() => new Map(menuGroups.value.map((item) => [item.id, item])))

  const menuSpaceMap = computed(
    () => new Map(menuSpaces.value.map((item) => [item.spaceKey, item]))
  )
  const menuSpaceOptions = computed(() =>
    menuSpaces.value.map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )

  const linkedPagesByMenuId = computed(() => {
    const map = new Map<string, Api.SystemManage.PageItem[]>()
    rawPages.value.forEach((item) => {
      const menuId = `${item.parentMenuId || ''}`.trim()
      if (!menuId) return
      const list = map.get(menuId) || []
      list.push(item)
      map.set(menuId, list)
    })
    map.forEach((list, key) => {
      list.sort((left, right) => {
        const sortDiff = Number(left.sortOrder || 0) - Number(right.sortOrder || 0)
        if (sortDiff !== 0) return sortDiff
        return `${left.name || ''}${left.pageKey || ''}`.localeCompare(
          `${right.name || ''}${right.pageKey || ''}`,
          'zh-Hans-CN'
        )
      })
      map.set(key, list)
    })
    return map
  })

  const getManageGroupId = (item: AppRouteRecord) =>
    `${item?.manage_group_id || item?.manage_group?.id || ''}`.trim()

  const getLinkedPages = (item: any) => linkedPagesByMenuId.value.get(String(item?.id || '')) || []

  const getSpaceName = (spaceKey?: string) => {
    const normalized = `${spaceKey || ''}`.trim()
    if (!normalized) return '未选择空间'
    return menuSpaceMap.value.get(normalized)?.name || normalized
  }

  const currentSpaceName = computed(() => getSpaceName(activeSpaceKey.value))
  const menuPageTitle = computed(() => (isLayoutMode.value ? '空间布局' : '菜单定义管理'))
  const menuPageDescription = computed(() =>
    isLayoutMode.value
      ? '维护当前 App 在不同空间下的菜单树摆放、排序与可见性；菜单定义本体仍由定义管理页统一维护。'
      : '统一管理当前 App 的目录、入口路由与外链定义；空间差异化布局与 Host 绑定统一放到高级空间配置里。'
  )
  const menuToolbarTip = computed(() =>
    isLayoutMode.value
      ? '当前按空间查看菜单布局树；同一菜单定义在多个空间复用时，共享一份授权与裁剪状态。'
      : '当前按 App 维护菜单定义；父级、排序和显示位置以当前查看空间做布局参考，不再把空间当成菜单主归属。'
  )
  const backupDialogTitle = computed(() =>
    isLayoutMode.value ? '备份当前空间布局' : '备份菜单定义'
  )
  const backupAlertTitle = computed(() =>
    isLayoutMode.value
      ? `当前将备份空间布局：${currentSpaceName.value}`
      : `当前将创建 App 级定义备份：${targetAppKey.value}`
  )
  const backupAlertDescription = computed(() =>
    isLayoutMode.value
      ? '该备份只保存当前 App 下当前空间的布局树和相关菜单分组，用于后续覆盖恢复当前空间。'
      : '该备份只保存当前 App 的菜单定义集合，不含各空间的父级、排序和显隐差异；空间级恢复请到高级空间配置页处理。'
  )
  const backupListTitle = computed(() =>
    isLayoutMode.value ? '管理空间布局备份' : '管理定义备份'
  )
  const backupListAlertDescription = computed(() =>
    isLayoutMode.value
      ? '这里只展示当前 App 下当前空间的布局备份，不包含 App 级菜单定义备份。'
      : '这里只展示当前 App 的定义备份；空间级布局备份请到空间高级配置页管理。'
  )
  const backupListEmptyDescription = computed(() =>
    isLayoutMode.value ? '当前空间暂无布局备份' : '当前 App 暂无定义备份'
  )

  const getBackupScopeLabel = (item: Api.SystemManage.MenuBackupItem) => {
    if (`${item.scope_type || ''}`.trim() === 'global') {
      return isLayoutMode.value ? '全空间备份' : '定义备份'
    }
    return getSpaceName(item.space_key || activeSpaceKey.value)
  }

  const hashToNegativeNumber = (value: string) => {
    let hash = 0
    for (let i = 0; i < value.length; i += 1) {
      hash = (hash * 31 + value.charCodeAt(i)) | 0
    }
    return -Math.abs(hash || 1)
  }

  const buildManageGroupNode = (
    group: Api.SystemManage.MenuManageGroupItem,
    parentKey: string
  ): AppRouteRecord => ({
    id: hashToNegativeNumber(`__manage_group__${parentKey}__${group.id}`),
    path: '',
    name: `manage-group-${group.id}`,
    component: '',
    sort_order: group.sortOrder ?? 0,
    manage_group_id: group.id,
    manage_group: {
      id: group.id,
      name: group.name,
      sort_order: group.sortOrder,
      status: group.status
    },
    meta: {
      title: group.name,
      __manageGroupNode: true,
      isEnable: group.status !== 'disabled'
    },
    children: []
  })

  const injectManageGroups = (
    items: AppRouteRecord[],
    parentKey = 'root',
    inheritedGroupId = ''
  ): AppRouteRecord[] => {
    const result: AppRouteRecord[] = []
    const groupNodeMap = new Map<string, AppRouteRecord>()

    items.forEach((item) => {
      const currentGroupID = getManageGroupId(item)
      const children = item.children?.length
        ? injectManageGroups(
            item.children as AppRouteRecord[],
            `${item.id || item.path || parentKey}`,
            currentGroupID || inheritedGroupId
          )
        : []
      const cloned = cloneMenuNode(item, children)
      const groupID = getManageGroupId(cloned)
      const group = groupID ? menuGroupMap.value.get(groupID) : undefined
      if (!group || groupID === inheritedGroupId) {
        result.push(cloned)
        return
      }

      let groupNode = groupNodeMap.get(group.id)
      if (!groupNode) {
        groupNode = buildManageGroupNode(group, parentKey)
        groupNodeMap.set(group.id, groupNode)
        result.push(groupNode)
      }
      const groupedChildren = (groupNode.children || []) as AppRouteRecord[]
      groupedChildren.push(cloned)
      groupNode.children = groupedChildren
    })

    result.sort((a, b) => Number(a.sort_order ?? 0) - Number(b.sort_order ?? 0))
    return result
  }

  const filterMenuTree = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items.reduce<AppRouteRecord[]>((result, item) => {
      if (!matchesMenuFilters(item)) {
        return result
      }

      const children = item.children?.length
        ? filterMenuTree(item.children as AppRouteRecord[])
        : []
      if (!matchesMenuSearch(item) && children.length === 0) {
        return result
      }

      result.push(cloneMenuNode(item, children))
      return result
    }, [])
  }

  const filteredMenuTree = computed(() => filterMenuTree(rawMenuTree.value))

  const groupingAvailable = computed(() => !menuGroupApiUnavailable.value)

  const tableData = computed(() =>
    groupingAvailable.value && groupingEnabled.value
      ? injectManageGroups(filteredMenuTree.value)
      : filteredMenuTree.value
  )

  const menuStats = computed(() => {
    const stats = {
      total: 0,
      directory: 0,
      entry: 0,
      external: 0,
      groups: menuGroups.value.length
    }
    const walk = (items: AppRouteRecord[]) => {
      items.forEach((item) => {
        stats.total += 1
        const kind = `${item.kind || ''}`.trim()
        if (kind === 'external') {
          stats.external += 1
        } else if (kind === 'entry') {
          stats.entry += 1
        } else {
          stats.directory += 1
        }
        if (item.children?.length) {
          walk(item.children as AppRouteRecord[])
        }
      })
    }
    walk(filteredMenuTree.value)
    return stats
  })

  const menuHeroMetrics = computed(() => [
    { label: '当前 App', value: targetAppKey.value },
    { label: isLayoutMode.value ? '布局空间' : '参考空间', value: currentSpaceName.value },
    { label: '总数', value: menuStats.value.total },
    { label: '目录', value: menuStats.value.directory },
    { label: '入口', value: menuStats.value.entry },
    { label: '外链', value: menuStats.value.external },
    { label: '分组', value: menuStats.value.groups }
  ])

  const getMenuList = async () => {
    loading.value = true
    loadError.value = ''
    dataFromBackend.value = false
    if (!targetAppKey.value) {
      rawMenuTree.value = []
      rawPages.value = []
      menuGroups.value = []
      activeSpaceKey.value = ''
      loadError.value = managedAppMissingText
      loading.value = false
      return
    }
    try {
      const [list, pagesResult, groupsResult] = await Promise.all([
        fetchGetMenuTreeAll(activeSpaceKey.value, targetAppKey.value),
        fetchGetPageOptions(activeSpaceKey.value, targetAppKey.value).then((res) => res.records || []),
        fetchGetMenuManageGroups()
          .then((groups) => ({ ok: true as const, groups }))
          .catch((error) => ({
            ok: false as const,
            error,
            groups: [] as Api.SystemManage.MenuManageGroupItem[]
          }))
      ])
      rawMenuTree.value = Array.isArray(list) ? list : []
      rawPages.value = Array.isArray(pagesResult) ? pagesResult : []
      menuGroups.value = groupsResult.groups || []
      dataFromBackend.value = true

      if (!groupsResult.ok) {
        warnDev('[Menus] 菜单分组接口不可用，已降级为无分组模式', groupsResult.error)
        menuGroupApiUnavailable.value = true
      } else {
        menuGroupApiUnavailable.value = false
      }
    } catch (error) {
      warnDev('[Menus] 获取菜单数据失败，已回退为空列表', error)
      rawMenuTree.value = []
      rawPages.value = []
      menuGroups.value = []
      menuGroupApiUnavailable.value = false
      loadError.value = '菜单数据暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  // --- 表格列配置 ---
  const { columnChecks, columns: displayColumns } = useTableColumns(() => [
    {
      type: 'selection',
      width: 52,
      align: 'center',
      selectable: (row: any) => !isManageGroupRow(row),
      className: 'menu-selection-column',
      labelClassName: 'menu-selection-column'
    } as any,
    { prop: 'title', label: '菜单名称', minWidth: 200, useSlot: true, slotName: 'title' },
    { prop: 'sort_order', label: '排序', width: 80, align: 'center' },
    { prop: 'type', label: '类型', width: 100, align: 'center', useSlot: true, slotName: 'type' },
    { prop: 'path', label: '路由', minWidth: 150, useSlot: true, slotName: 'path' },
    { prop: 'component', label: '组件路径', minWidth: 200, useSlot: true, slotName: 'component' },
    {
      prop: 'appKey',
      label: 'App',
      width: 150,
      align: 'center',
      formatter: (row: any) => row.appKey || targetAppKey.value
    },
    {
      prop: 'space',
      label: '菜单空间',
      width: 120,
      align: 'center',
      useSlot: true,
      slotName: 'space'
    },
    { prop: 'linkedPage', label: '受管页面', minWidth: 220, useSlot: true, slotName: 'linkedPage' },
    {
      prop: 'advanced',
      label: '高级配置',
      minWidth: 200,
      align: 'center',
      useSlot: true,
      slotName: 'advanced'
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      align: 'center',
      useSlot: true,
      slotName: 'status'
    },
    {
      prop: 'operation',
      label: '操作',
      width: 120,
      align: 'center',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  // --- 辅助方法 ---
  const getMenuTypeTag = (row: any) => {
    if (isManageGroupRow(row)) return 'warning'
    if (row.kind === 'external') return 'success'
    if (row.kind === 'entry') return 'primary'
    return 'info'
  }

  const getMenuTypeText = (row: any) => {
    if (isManageGroupRow(row)) return '分组'
    if (row.kind === 'external') return '外链'
    if (row.kind === 'entry') return '入口'
    return '目录'
  }

  const isDirectoryMenuRow = (row: any) => `${row?.kind || ''}`.trim() === 'directory'
  const isEntryMenuRow = (row: any) => `${row?.kind || ''}`.trim() === 'entry'

  const getMenuActionRequirementLabel = (row: any) => {
    const requirement = getMenuActionRequirement(row.meta)
    if (!requirement.actions.length) return ''
    const visibilityText = requirement.visibilityMode === 'show' ? '显示' : '隐藏'
    return `功能权限: 不满足${visibilityText}`
  }

  const getAccessModeLabel = (accessMode?: string) => {
    const mode = `${accessMode || 'permission'}`.trim()
    if (mode === 'jwt') return '登录可见'
    if (mode === 'public') return '公开可见'
    return '权限控制'
  }

  const getAccessModeTag = (accessMode?: string) => {
    const mode = `${accessMode || 'permission'}`.trim()
    if (mode === 'jwt') return 'warning'
    if (mode === 'public') return 'success'
    return 'info'
  }

  const getOperationList = (row: any): ButtonMoreItem[] => {
    if (isManageGroupRow(row)) return []
    const list: ButtonMoreItem[] = [
      { key: 'add', label: '新增子菜单', icon: 'ri:add-fill', auth: 'system.menu.manage' },
      { key: 'edit', label: '编辑菜单', icon: 'ri:edit-2-line', auth: 'system.menu.manage' },
      {
        key: 'action_requirement',
        label: '功能权限',
        icon: 'ri:shield-keyhole-line',
        auth: 'system.menu.manage'
      }
    ]
    if (!row.is_system) {
      list.push({
        key: 'delete',
        label: '删除菜单',
        icon: 'ri:delete-bin-4-line',
        color: '#f56c6c',
        auth: 'system.menu.manage'
      })
    }
    return list
  }

  // --- 事件处理 ---
  const handleReset = () => {
    Object.assign(formFilters, initialSearchState)
    Object.assign(appliedFilters, initialSearchState)
  }
  const handleSearch = () => {
    Object.assign(appliedFilters, formFilters)
  }

  const syncRouteSpaceKey = (spaceKey: string) => {
    router.replace({
      query: {
        ...route.query,
        spaceKey
      }
    })
  }

  const goToDefinitionManagement = () => {
    router.push({
      path: '/system/menu',
      query: {
        ...route.query,
        app_key: targetAppKey.value,
        spaceKey: activeSpaceKey.value || undefined,
        layout: undefined
      }
    })
  }

  const resolveInitialSpaceKey = () => {
    const requestedSpaceKey = `${route.query.spaceKey || ''}`.trim()
    if (requestedSpaceKey && menuSpaces.value.some((item) => item.spaceKey === requestedSpaceKey)) {
      return requestedSpaceKey
    }
    return menuSpaces.value.find((item) => item.isDefault)?.spaceKey || menuSpaces.value[0]?.spaceKey || ''
  }

  const syncMenuSpaces = async () => {
    if (!targetAppKey.value) {
      menuSpaces.value = []
      activeSpaceKey.value = ''
      return
    }
    const res = await fetchGetMenuSpaces(targetAppKey.value)
    menuSpaces.value = res.records || []
    activeSpaceKey.value = resolveInitialSpaceKey()
  }

  const handleRefresh = () => getMenuList()
  const handleSpaceChange = () => {
    syncRouteSpaceKey(activeSpaceKey.value)
    getMenuList()
  }
  const rowKey = (row: any) => String(row.id || row.path)

  const clearBatchSelection = () => {
    selectedMenuRows.value = []
    batchTargetGroupId.value = ''
    batchAssignDialogVisible.value = false
    tableRef.value?.elTableRef?.clearSelection?.()
  }

  const handleBatchSelectionChange = (rows: any[]) => {
    selectedMenuRows.value = (rows || []).filter((row) => !isManageGroupRow(row))
  }

  const collectMenuSubtree = (rows: any[]) => {
    const result: any[] = []
    const seen = new Set<string>()

    const visit = (row: any) => {
      if (!row || isManageGroupRow(row)) return
      const key = String(row.id || row.path || '')
      if (!key || seen.has(key)) return
      seen.add(key)
      result.push(row)
      ;(row.children || []).forEach((child: any) => visit(child))
    }

    rows.forEach((row) => visit(row))
    return result
  }

  const getMenuChildCount = (row: any) => ((row?.children || []).filter((item: any) => !isManageGroupRow(item))).length

  const getMenuDescendantCount = (row: any) => {
    if (!row) return 0
    return collectMenuSubtree([row]).length
  }

  const getAffectedPageCount = (row: any) => {
    if (!row?.id) return 0
    const subtree = collectMenuSubtree([row]).map((item) => String(item.id || ''))
    if (subtree.length === 0) return 0
    const seen = new Set<string>()
    let count = 0
    subtree.forEach((menuId) => {
      const pages = getLinkedPages({ id: menuId })
      pages.forEach((page) => {
        const key = String(page.pageKey || page.id || '')
        if (!key || seen.has(key)) return
        seen.add(key)
        count += 1
      })
    })
    return count
  }

  type MenuDeleteParentOption = {
    label: string
    value: string
    children?: MenuDeleteParentOption[]
  }

  const getDeleteParentOptions = (row: any): MenuDeleteParentOption[] => {
    if (!row?.id) return []
    const excluded = new Set<string>(collectMenuSubtree([row]).map((item) => String(item.id || '')))
    const walk = (items: AppRouteRecord[]) => {
      return items.reduce<MenuDeleteParentOption[]>((acc, item) => {
        if (isManageGroupRow(item)) return acc
        const key = String(item.id || '')
        if (!key || excluded.has(key)) return acc
        const children = item.children
          ? walk(item.children as AppRouteRecord[])
          : []
        acc.push({
          label: formatMenuTitle(item.meta?.title) || String(item.name || key),
          value: key,
          children: children.length > 0 ? children : undefined
        })
        return acc
      }, [])
    }
    return walk(filteredMenuTree.value)
  }

  const handleMoreActionCommand = (command: string) => {
    if (command === 'manageGroup') {
      if (!groupingAvailable.value) return
      manageGroupDrawerVisible.value = true
      return
    }
    if (command === 'backupGlobal') {
      handleBackupMenu('global')
      return
    }
    if (command === 'backupList') {
      handleManageBackups()
    }
  }

  const hasOwnManageGroup = (row: any) => Boolean(`${row?.manage_group_id || ''}`.trim())

  const buildMenuUpdatePayloadFromRow = (row: any, manageGroupID: string | null) => {
    const meta = buildMenuMetaForUpdate(row)
    return {
      parent_id: row.parent_id ? String(row.parent_id) : null,
      kind: row.kind || 'directory',
      path: row.path || '',
      name: row.name || '',
      component: typeof row.component === 'string' ? row.component : '',
      title: row.meta?.title || '',
      icon: row.meta?.icon || '',
      sort_order: Number(row.sort_order ?? 0),
      space_key: `${row.spaceKey || row.space_key || row.meta?.spaceKey || activeSpaceKey.value || ''}`.trim(),
      manage_group_id: manageGroupID,
      meta
    }
  }

  const applyBatchGroupAction = async (action: 'assign' | 'remove', assignGroupID?: string) => {
    if (selectedMenuRows.value.length === 0) {
      ElMessage.warning('请先勾选菜单')
      return
    }
    const nextGroupID = action === 'assign' ? `${assignGroupID || ''}`.trim() : ''
    if (action === 'assign' && !nextGroupID) {
      ElMessage.warning('请选择目标分组')
      return
    }
    const targetGroupID = action === 'remove' ? null : nextGroupID
    const expandedRows = collectMenuSubtree(selectedMenuRows.value)
    const actionableRows =
      action === 'remove' ? expandedRows.filter((row) => hasOwnManageGroup(row)) : expandedRows

    if (action === 'remove' && actionableRows.length === 0) {
      ElMessage.warning('所选菜单及其下级没有已绑定的分组')
      return
    }

    const text = action === 'remove' ? '移出所选菜单的分组归属' : '移入所选菜单到目标分组'
    try {
      await ElMessageBox.confirm(`确定要${text}吗？`, '批量操作确认', { type: 'warning' })
      await Promise.all(
        actionableRows.map((row) =>
          fetchUpdateMenu(String(row.id), buildMenuUpdatePayloadFromRow(row, targetGroupID), {
            showErrorMessage: false
          })
        )
      )
      if (action === 'remove' && actionableRows.length !== expandedRows.length) {
        ElMessage.success(
          `批量移出成功，已跳过 ${expandedRows.length - actionableRows.length} 项未绑定分组菜单`
        )
      } else {
        ElMessage.success('批量操作成功')
      }
      await getMenuList()
      clearBatchSelection()
      batchTargetGroupId.value = ''
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '批量操作失败')
      }
    }
  }

  const handleBatchCommand = (command: 'assign' | 'remove') => {
    if (command === 'assign') {
      if (selectedMenuRows.value.length === 0) {
        ElMessage.warning('请先勾选菜单')
        return
      }
      batchTargetGroupId.value = ''
      batchAssignDialogVisible.value = true
      return
    }
    applyBatchGroupAction(command)
  }

  const handleBatchAssignSubmit = async () => {
    await applyBatchGroupAction('assign', batchTargetGroupId.value)
  }

  const setExpandState = (expanded: boolean) => {
    isExpanded.value = expanded
    nextTick(() => {
      if (tableRef.value?.elTableRef && tableData.value.length) {
        const processRows = (rows: any[]) => {
          rows.forEach((row) => {
            if (row.children?.length) {
              tableRef.value.elTableRef.toggleRowExpansion(row, isExpanded.value)
              processRows(row.children)
            }
          })
        }
        processRows(tableData.value)
      }
    })
  }

  const handleExpandSwitchChange = (value: string | number | boolean) => {
    setExpandState(Boolean(value))
  }

  // --- CRUD 操作 ---
  const handleAddMenu = () => {
    editData.value = null
    parentRowForAdd.value = null
    dialogVisible.value = true
  }
  const handleAddUnderRow = (row: any) => {
    if (isManageGroupRow(row)) return
    editData.value = null
    parentRowForAdd.value = row
    dialogVisible.value = true
  }
  const handleEditMenu = (row: any) => {
    if (isManageGroupRow(row)) return
    editData.value = row
    parentRowForAdd.value = null
    dialogVisible.value = true
  }
  const handleEditActionRequirement = (row: any) => {
    actionRequirementData.value = row
    actionRequirementVisible.value = true
  }

  const normalizeRequiredActions = (items?: string[]) =>
    Array.from(new Set((items || []).map((item) => `${item || ''}`.trim()).filter(Boolean)))

  const applyActionRequirementToMeta = (
    meta: Record<string, any>,
    formData: {
      requiredActions?: string[]
      actionMatchMode?: 'any' | 'all'
      actionVisibilityMode?: 'hide' | 'show'
    }
  ) => {
    const requiredActions = normalizeRequiredActions(formData.requiredActions)
    delete meta.requiredAction
    delete meta.requiredActions
    delete meta.actionMatchMode
    delete meta.actionVisibilityMode

    if (requiredActions.length === 1) {
      meta.requiredAction = requiredActions[0]
    }
    if (requiredActions.length > 1) {
      meta.requiredActions = requiredActions
      meta.actionMatchMode = formData.actionMatchMode === 'all' ? 'all' : 'any'
    }
    if (requiredActions.length > 0) {
      meta.actionVisibilityMode = formData.actionVisibilityMode === 'show' ? 'show' : 'hide'
    }
  }

  const buildMenuMetaFromForm = (formData: any) => {
    const isEntry = `${formData.kind || 'entry'}` === 'entry'
    const isExternal = `${formData.kind || ''}` === 'external'
    const workingSpaceKey = `${formData.spaceKey || activeSpaceKey.value || ''}`.trim()
    const meta: Record<string, any> = {
      isEnable: formData.isEnable,
      keepAlive: isEntry ? formData.keepAlive : false,
      isHide: !!formData.isHide,
      isHideTab: isEntry ? formData.isHideTab : false,
      isIframe: `${formData.kind || ''}` !== 'directory' ? formData.isIframe : false,
      showBadge: formData.showBadge,
      showTextBadge: formData.showTextBadge || '',
      link: isExternal ? formData.link || '' : '',
      activePath: isEntry ? formData.activePath || '' : '',
      fixedTab: isEntry ? formData.fixedTab : false,
      isFullPage: isEntry ? formData.isFullPage : false,
      accessMode: formData.accessMode || 'permission',
      spaceKey: workingSpaceKey
    }
    if (isEntry && formData.customParent?.trim()) {
      meta.customParent = formData.customParent.trim()
    }
    applyActionRequirementToMeta(meta, formData)
    return meta
  }

  const buildMenuRequestPayload = (formData: any, meta: Record<string, any>) => ({
    app_key: targetAppKey.value,
    kind: formData.kind || 'entry',
    path: formData.path || '/',
    name: formData.label || '',
    component: formData.kind === 'entry' ? formData.component || '' : '',
    title: formData.name || '',
    icon: formData.icon || '',
    sort_order: Number(formData.sort ?? 0),
    space_key: `${formData.spaceKey || activeSpaceKey.value || ''}`.trim(),
    manage_group_id: formData.manageGroupId?.trim() || null,
    meta
  })

  const resolveParentId = (formData: any) =>
    formData.parentId?.trim() ||
    (parentRowForAdd.value?.id ? String(parentRowForAdd.value.id) : null)
  const handleMenuOperation = (item: ButtonMoreItem, row: any) => {
    if (item.key === 'add') handleAddUnderRow(row)
    else if (item.key === 'edit') handleEditMenu(row)
    else if (item.key === 'action_requirement') handleEditActionRequirement(row)
    else if (item.key === 'delete') handleDeleteMenu(row)
  }

  const handleBackupListAction = (action: string, row: Api.SystemManage.MenuBackupItem) => {
    if (action === 'restore') {
      handleRestoreBackup(row)
      return
    }
    if (action === 'delete') {
      handleDeleteBackup(row.id)
    }
  }

  const handleDeleteMenu = async (row: any) => {
    if (!dataFromBackend.value || !row.id) return ElMessage.info('预览模式无法删除')
    if (row.is_system) return ElMessage.warning('系统菜单不可删除')
    deleteTargetRow.value = row
    deleteDialogVisible.value = true
    deleteLoading.value = true
    try {
      deletePreview.value = await fetchGetMenuDeletePreview(String(row.id), { mode: 'cascade' })
    } catch (e: any) {
      ElMessage.error(e?.message || '获取删除预览失败')
      deleteDialogVisible.value = false
      deleteTargetRow.value = null
    } finally {
      deleteLoading.value = false
    }
  }

  const handleDeleteMenuConfirm = async (payload: {
    mode: MenuDeleteMode
    targetParentId?: string | null
  }) => {
    if (!deleteTargetRow.value?.id) return
    deleteLoading.value = true
    try {
      await fetchDeleteMenu(String(deleteTargetRow.value.id), {
        mode: payload.mode,
        targetParentId: payload.targetParentId || undefined
      })
      ElMessage.success(payload.mode === 'cascade' ? '菜单树已删除' : '菜单已删除')
      deleteDialogVisible.value = false
      deleteTargetRow.value = null
      deletePreview.value = null
      await getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    } finally {
      deleteLoading.value = false
    }
  }

  const handleSubmit = async (formData: any) => {
    if (!dataFromBackend.value) return getMenuList()
    try {
      const nextManageGroupID = formData.manageGroupId?.trim() || null
      const currentManageGroupID = editData.value?.manage_group_id
        ? String(editData.value.manage_group_id)
        : null
      const payload = buildMenuRequestPayload(formData, buildMenuMetaFromForm(formData))
      if (editData.value?.id) {
        const parentId = formData.parentId?.trim() || null
        await fetchUpdateMenu(
          String(editData.value.id),
          { ...payload, parent_id: parentId },
          { showErrorMessage: false }
        )

        if (currentManageGroupID !== nextManageGroupID) {
          const descendants = collectMenuSubtree(editData.value.children || []).filter(
            (row) => !isManageGroupRow(row)
          )
          if (descendants.length) {
            await Promise.all(
              descendants.map((row) =>
                fetchUpdateMenu(
                  String(row.id),
                  buildMenuUpdatePayloadFromRow(row, nextManageGroupID),
                  { showErrorMessage: false }
                )
              )
            )
          }
        }
      } else {
        const parentId = resolveParentId(formData)
        await fetchCreateMenu({ ...payload, parent_id: parentId }, { showErrorMessage: false })
      }
      ElMessage.success('保存成功')
      getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  }

  const buildMenuMetaForUpdate = (row: any) => {
    const meta = { ...(row?.meta || {}) }
    delete meta.title
    return meta
  }

  const handleActionRequirementSubmit = async (formData: {
    requiredActions: string[]
    actionMatchMode: 'any' | 'all'
    actionVisibilityMode: 'hide' | 'show'
  }) => {
    if (!dataFromBackend.value || !actionRequirementData.value?.id) return
    try {
      const row = actionRequirementData.value
      const meta = buildMenuMetaForUpdate(row)
      applyActionRequirementToMeta(meta, formData)
      await fetchUpdateMenu(
        String(row.id),
        {
          app_key: targetAppKey.value,
          parent_id: row.parent_id ? String(row.parent_id) : null,
          kind: row.kind || 'directory',
          path: row.path || '',
          name: row.name || '',
          component: typeof row.component === 'string' ? row.component : '',
          title: row.meta?.title || '',
          icon: row.meta?.icon || '',
          sort_order: Number(row.sort_order ?? 0),
          space_key:
            `${row.spaceKey || row.space_key || row.meta?.spaceKey || activeSpaceKey.value || ''}`.trim(),
          meta
        },
        { showErrorMessage: false }
      )
      ElMessage.success('功能权限已保存')
      actionRequirementVisible.value = false
      actionRequirementData.value = null
      getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '功能权限保存失败')
    }
  }

  // --- 菜单备份相关方法 ---
  const handleBackupMenu = (scopeType: MenuBackupScopeType) => {
    backupScopeType.value = scopeType
    backupDialogVisible.value = true
  }

  const handleSaveManageGroup = async (
    payload: Api.SystemManage.MenuManageGroupSaveParams & { id?: string }
  ) => {
    groupSaving.value = true
    try {
      if (payload.id) {
        await fetchUpdateMenuManageGroup(payload.id, payload)
      } else {
        await fetchCreateMenuManageGroup(payload)
      }
      ElMessage.success('菜单分组已保存')
      await getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '菜单分组保存失败')
    } finally {
      groupSaving.value = false
    }
  }

  const handleDeleteManageGroup = async (id: string) => {
    groupSaving.value = true
    try {
      await fetchDeleteMenuManageGroup(id)
      ElMessage.success('菜单分组已删除')
      await getMenuList()
    } catch (e: any) {
      ElMessage.error(e?.message || '菜单分组删除失败')
    } finally {
      groupSaving.value = false
    }
  }

  const handleCreateBackup = async (formData: { name: string; description: string }) => {
    backupLoading.value = true
    try {
      const payload: Api.SystemManage.MenuBackupCreateParams = {
        app_key: targetAppKey.value,
        name: formData.name,
        description: formData.description,
        // scope_type 显式声明是“当前空间备份”还是“全部空间备份”，
        // 后端创建逻辑不再依赖 space_key 是否缺省来猜测语义。
        scope_type: backupScopeType.value
      }
      if (backupScopeType.value === 'space') {
        payload.space_key = activeSpaceKey.value
      }
      await fetchCreateMenuBackup(payload)
      ElMessage.success(backupScopeType.value === 'global' ? '全局备份已创建' : '空间备份已创建')
      backupDialogVisible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message || '备份失败')
    } finally {
      backupLoading.value = false
    }
  }

  const handleManageBackups = async () => {
    backupLoading.value = true
    try {
      const list = await fetchGetMenuBackupList(
        isLayoutMode.value ? activeSpaceKey.value : undefined,
        targetAppKey.value
      )
      const filteredList = (list || []).filter((item) =>
        isLayoutMode.value ? `${item.scope_type || ''}`.trim() !== 'global' : `${item.scope_type || ''}`.trim() === 'global'
      )
      backupList.value = filteredList.map((item) => ({
        ...item,
        // scope_origin 由后端显式区分正式全局备份和历史兼容备份，前端这里只负责映射清晰文案。
        space_name: getBackupScopeLabel(item)
      }))
      backupListDialogVisible.value = true
    } catch (e: any) {
      ElMessage.error(e?.message || '获取备份列表失败')
    } finally {
      backupLoading.value = false
    }
  }

  const buildBackupRestoreMessage = (item: Api.SystemManage.MenuBackupItem) => {
    if (item.scope_type === 'global') {
      return '确定要恢复该备份吗？该操作会覆盖当前 App 的菜单定义集合。'
    }
    return `确定要恢复该备份吗？恢复后会覆盖当前菜单空间“${getSpaceName(item.space_key || activeSpaceKey.value)}”的菜单配置。`
  }

  const handleRestoreBackup = async (item: Api.SystemManage.MenuBackupItem) => {
    try {
      await ElMessageBox.confirm(buildBackupRestoreMessage(item), '提示', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      backupLoading.value = true
      await fetchRestoreMenuBackup(item.id, targetAppKey.value)
      ElMessage.success('恢复成功')
      backupListDialogVisible.value = false
      getMenuList()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '恢复失败')
      }
    } finally {
      backupLoading.value = false
    }
  }

  const handleDeleteBackup = async (id: string) => {
    try {
      await ElMessageBox.confirm('确定要删除该备份吗？', '提示', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      backupLoading.value = true
      await fetchDeleteMenuBackup(id, targetAppKey.value)
      ElMessage.success('删除成功')
      handleManageBackups()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.message || '删除失败')
      }
    } finally {
      backupLoading.value = false
    }
  }

  watch(multiSelectEnabled, (enabled) => {
    if (!enabled) {
      clearBatchSelection()
    }
  })

  watch(deleteDialogVisible, (visible) => {
    if (!visible && !deleteLoading.value) {
      deleteTargetRow.value = null
      deletePreview.value = null
    }
  })

  watch(groupingEnabled, (enabled) => {
    if (!enabled) {
      setExpandState(false)
    }
    localStorage.setItem('system:menu:grouping-enabled', enabled ? '1' : '0')
  })

  watch(menuGroupApiUnavailable, (unavailable) => {
    if (unavailable) {
      setExpandState(false)
    }
  })

  watch(groupedMenuVisible, (visible) => {
    localStorage.setItem('system:menu:grouped-visible', visible ? '1' : '0')
  })

  // --- 生命周期 & 监听 ---
  onMounted(() => {
    groupingEnabled.value = localStorage.getItem('system:menu:grouping-enabled') !== '0'
    groupedMenuVisible.value = localStorage.getItem('system:menu:grouped-visible') !== '0'
    if (!targetAppKey.value) {
      loadError.value = managedAppMissingText
      return
    }
    syncMenuSpaces()
      .finally(() => {
        getMenuList()
      })
  })

  watch(
    () => [route.query.app_key, route.query.spaceKey],
    async ([, value]) => {
      if (!targetAppKey.value) {
        rawMenuTree.value = []
        rawPages.value = []
        menuSpaces.value = []
        activeSpaceKey.value = ''
        loadError.value = managedAppMissingText
        return
      }
      await syncMenuSpaces()
      const requestedSpaceKey = `${value || ''}`.trim()
      if (!requestedSpaceKey || requestedSpaceKey === activeSpaceKey.value) {
        await getMenuList()
        return
      }
      if (!menuSpaces.value.some((item) => item.spaceKey === requestedSpaceKey)) {
        await getMenuList()
        return
      }
      activeSpaceKey.value = requestedSpaceKey
      await getMenuList()
    }
  )
</script>

<style lang="scss" scoped>
  .menu-overview {
    padding: 2px 0 12px;
    margin-bottom: 12px;
    border-bottom: 1px solid var(--art-card-border);
  }

  .menu-inline-alert {
    margin-bottom: 12px;
  }

  .menu-overview-main {
    min-width: 0;
  }

  .menu-overview-heading {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 12px 18px;
  }

  .menu-overview-title {
    font-size: 20px;
    font-weight: 750;
    line-height: 1.1;
    color: var(--art-text-strong);
    letter-spacing: -0.02em;
  }

  .menu-overview-metrics {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 14px;
  }

  .menu-overview-subline {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-top: 12px;
  }

  .menu-metric-item {
    font-size: 13px;
    font-weight: 600;
    color: var(--art-text-base);
    white-space: nowrap;
  }

  .menu-overview-subtitle {
    font-size: 13px;
    line-height: 1.6;
    color: var(--art-text-muted);
  }

  .menu-overview-switches {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px 16px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--art-card-border);
  }

  .menu-overview-switch-list {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 16px;
  }

  .menu-overview-tools {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin-left: auto;
  }

  .menu-tool-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    color: var(--art-text-base);
    cursor: pointer;
    background: rgb(255 255 255 / 0.9);
    border: 1px solid var(--art-card-border);
    border-radius: 12px;
    box-shadow: var(--art-shadow-sm);
    transition:
      border-color 0.15s ease,
      background-color 0.15s ease,
      transform 0.15s ease;
  }

  .menu-tool-button:hover {
    border-color: color-mix(in srgb, var(--theme-color) 20%, var(--art-card-border));
    background: color-mix(in srgb, var(--theme-color) 7%, white);
    transform: translateY(-1px);
  }

  .menu-tool-button.is-active {
    color: #ffffff;
    background: var(--el-color-primary);
  }

  .menu-tool-button.is-active:hover {
    background: color-mix(in srgb, var(--el-color-primary) 80%, white);
  }

  .menu-hero-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px;
  }

  .menu-top-stack {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .menu-hero {
    margin-top: 0;
  }

  .menu-toolbar {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 100%;
    padding: 4px 0 2px;
  }

  .menu-toolbar-top,
  .menu-toolbar-bottom {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px 14px;
  }

  .menu-toolbar-tip {
    font-size: 13px;
    line-height: 1.6;
    color: var(--art-text-muted);
  }

  .menu-toolbar-switches {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 16px;
  }

  .menu-toolbar-actions {
    display: inline-flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px 10px;
  }

  .menu-space-select {
    width: 220px;
  }

  .menu-toolbar-batch {
    padding-left: 14px;
    border-left: 1px solid var(--art-card-border);
  }

  .menu-inline-note {
    font-size: 12px;
    color: var(--art-text-muted);
    white-space: nowrap;
  }

  .menu-switch-item {
    display: inline-flex;
    align-items: center;
    gap: 12px;
    padding: 0;
  }

  .menu-switch-label {
    font-size: 13px;
    color: var(--art-text-base);
    white-space: nowrap;
  }

  .menu-batch-count {
    font-size: 13px;
    font-weight: 600;
    color: var(--art-text-strong);
  }

  .menu-batch-dialog {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .menu-batch-dialog-count {
    font-size: 13px;
    line-height: 1.6;
    color: var(--art-text-muted);
  }

  .menu-batch-dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .menu-columns-popover,
  .menu-settings-popover {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 140px;
  }

  .menu-settings-popover-text {
    font-size: 13px;
    color: var(--art-text-muted);
  }

  .advanced-configs {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
  }

  .menu-linked-page-cell {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 2px;
  }

  .menu-linked-page-cell__primary {
    overflow: hidden;
    color: var(--art-text-strong);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .menu-linked-page-cell__meta {
    font-size: 12px;
    line-height: 1.5;
    color: var(--art-text-muted);
  }

  .menu-group-title {
    color: var(--art-text-strong);
    font-weight: 700;
  }

  :deep(.el-table) {
    .el-table__row {
      transition: all 0.3s ease;

      &:hover {
        background-color: color-mix(in srgb, var(--theme-color) 4%, white) !important;
      }
    }

    .el-table__header-wrapper th {
      background-color: color-mix(in srgb, var(--default-box-color) 94%, var(--default-bg-color));
      font-weight: 600;
      color: var(--art-text-base);
    }

    .el-table__body-wrapper {
      .el-table__row {
        height: 48px;
      }
    }
  }

  :deep(.el-table .el-table__body tr:has(.menu-group-title)) {
    background: linear-gradient(
      180deg,
      color-mix(in srgb, var(--theme-color) 7%, white) 0%,
      color-mix(in srgb, var(--theme-color) 4%, white) 100%
    ) !important;
  }

  :deep(.el-table .el-table__body tr:has(.menu-group-title):hover > td.el-table__cell) {
    background-color: color-mix(in srgb, var(--theme-color) 10%, white) !important;
  }

  :deep(.menu-table-multi-disabled .menu-selection-column) {
    width: 0 !important;
    min-width: 0 !important;
    padding: 0 !important;
    border: 0 !important;
  }

  :deep(.menu-table-multi-disabled .menu-selection-column .cell) {
    display: none !important;
  }

  :deep(.el-card__body) {
    padding-top: 14px;
  }

  @media (max-width: 960px) {
    .menu-toolbar-top,
    .menu-toolbar-bottom {
      justify-content: flex-start;
      width: 100%;
    }

    .menu-toolbar-batch {
      border-left: 0;
      padding-left: 0;
    }
  }

  @media (max-width: 640px) {
    .menu-hero-actions,
    .menu-toolbar-switches {
      width: 100%;
    }

    .menu-switch-item {
      width: 100%;
      justify-content: space-between;
    }
  }
</style>

