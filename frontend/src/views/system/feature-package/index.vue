<template>
  <div class="feature-package-page art-full-height">
    <ElTabs
      v-model="activePackageType"
      type="card"
      class="package-tabs"
      @tab-change="handleTabChange"
    >
      <ElTabPane label="基础包" name="base" />
      <ElTabPane label="组合包" name="bundle" />
    </ElTabs>

    <div class="page-top-stack">
      <ArtSearchBar
        v-show="showSearchBar"
        v-model="searchForm"
        :items="searchItems"
        label-position="top"
        :span="8"
        :gutter="16"
        :showExpand="true"
        @search="handleSearch"
        @reset="handleReset"
      />

      <AdminWorkspaceHero
        title="功能包管理"
        description="功能包目录可全局查看，基础包 / 组合包按空间范围维护；菜单、功能和协作空间绑定仍然需要进入具体 App 上下文。"
        :metrics="[
          { label: '当前 App 上下文', value: targetAppKey || '全局目录' },
          { label: '当前页功能包数', value: data.length },
          { label: '个人空间功能包', value: personalPackageCount },
          { label: '协作空间功能包', value: collaborationPackageCount },
          { label: '双空间范围功能包', value: sharedPackageCount },
          {
            label: activePackageType === 'base' ? '已组合功能范围数' : '组合包数量',
            value: activePackageType === 'base' ? totalActionCount : bundleCount
          },
          {
            label: activePackageType === 'base' ? '已绑定菜单数' : '协作空间开通数',
            value: activePackageType === 'base' ? totalMenuCount : totalCollaborationWorkspaceCount
          },
          { label: '停用功能包', value: disabledPackageCount }
        ]"
      >
        <div class="feature-package-hero-actions">
          <ElButton
            v-action="'feature_package.manage'"
            :disabled="!targetAppKey"
            @click="openRelationDialog"
            v-ripple
          >
            包关系树
          </ElButton>
          <ElButton
            v-action="'feature_package.manage'"
            type="primary"
            :disabled="!targetAppKey"
            @click="openDialog('add')"
            v-ripple
          >
            新增{{ activePackageType === 'base' ? '基础包' : '组合包' }}
          </ElButton>
        </div>
      </AdminWorkspaceHero>
    </div>

    <ElCard class="art-table-card" shadow="never">
      <ElAlert
        v-if="loadError"
        :title="loadError"
        type="error"
        show-icon
        :closable="false"
        class="feature-package-error"
      />

      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="handleRefresh"
      >
        <template #left>
          <div class="feature-package-toolbar-tip">
            列表默认展示全部功能包。包关系、菜单、功能范围和协作空间开通统一从操作菜单进入；这些动作仍需具体 App 上下文。
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <FeaturePackageDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :package-data="currentPackage"
      :app-key="resolvePackageAppKey(currentPackage)"
      :default-package-type="activePackageType"
      @success="handleRefresh"
    />

    <FeaturePackageBundlesDialog
      v-model="bundlesDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :app-key="resolvePackageAppKey(currentPackage)"
      :context-type="currentPackage.contextType || 'collaboration'"
      @success="handleRefresh"
    />

    <FeaturePackageActionsDialog
      v-model="actionsDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'collaboration'"
      @success="handleRefresh"
    />

    <FeaturePackageMenusDialog
      v-model="menusDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :app-key="resolvePackageAppKey(currentPackage)"
      :context-type="currentPackage.contextType || 'collaboration'"
      @success="handleRefresh"
    />

    <FeaturePackageCollaborationWorkspacesDialog
      v-model="collaborationWorkspacesDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'collaboration'"
      @success="handleRefresh"
    />

    <ElDialog
      v-model="relationDialogVisible"
      title="功能包包含关系树"
      width="920px"
      destroy-on-close
    >
      <div class="relation-toolbar">
        <ElInput
          v-model="relationKeyword"
          placeholder="按包名/编码过滤"
          clearable
          style="max-width: 260px"
          @keyup.enter="loadRelationTree"
        />
        <ElButton :loading="relationLoading" type="primary" @click="loadRelationTree">
          刷新关系树
        </ElButton>
      </div>
      <ElAlert
        v-if="relationTree.cycleDependencies.length"
        type="warning"
        :closable="false"
        show-icon
        class="relation-alert"
        :title="`检测到循环依赖 ${relationTree.cycleDependencies.length} 组`"
      />
      <ElAlert
        v-if="relationTree.isolatedBaseKeys.length"
        type="info"
        :closable="false"
        show-icon
        class="relation-alert"
        :title="`孤立基础包：${relationTree.isolatedBaseKeys.join('、')}`"
      />
      <ElScrollbar max-height="520px" v-loading="relationLoading">
        <ElTree
          :data="relationTree.roots"
          node-key="id"
          default-expand-all
          :expand-on-click-node="false"
        >
          <template #default="{ data: node }">
            <div class="relation-node">
              <span class="relation-node-name">{{ node.name }}</span>
              <ElTag
                size="small"
                effect="plain"
                :type="node.packageType === 'bundle' ? 'warning' : 'success'"
              >
                {{ node.packageType === 'bundle' ? '组合包' : '基础包' }}
              </ElTag>
              <ElTag
                size="small"
                effect="plain"
                :type="
                  node.contextType === 'personal'
                    ? 'warning'
                    : node.contextType === 'collaboration'
                      ? 'primary'
                      : 'info'
                "
              >
                {{
                  node.contextType === 'personal'
                    ? '个人空间'
                    : node.contextType === 'collaboration'
                      ? '协作空间'
                      : '通用'
                }}
              </ElTag>
              <ElTag size="small" effect="plain" type="info"
                >被引用 {{ node.referenceCount }}</ElTag
              >
            </div>
          </template>
        </ElTree>
      </ElScrollbar>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, reactive, ref, watch } from 'vue'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useRoute } from 'vue-router'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import { useTable } from '@/hooks/core/useTable'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import {
    fetchDeleteFeaturePackage,
    fetchGetFeaturePackageImpactPreview,
    fetchGetFeaturePackageList,
    fetchGetFeaturePackageRelationTree
  } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import FeaturePackageDialog from './modules/feature-package-dialog.vue'
  import FeaturePackageBundlesDialog from './modules/feature-package-bundles-dialog.vue'
  import FeaturePackageActionsDialog from './modules/feature-package-actions-dialog.vue'
  import FeaturePackageMenusDialog from './modules/feature-package-menus-dialog.vue'
  import FeaturePackageCollaborationWorkspacesDialog from './modules/feature-package-collaboration-workspaces-dialog.vue'

  defineOptions({ name: 'FeaturePackage' })

  type PackageItem = Api.SystemManage.FeaturePackageItem

  type SearchForm = {
    keyword: string
    packageKey: string
    name: string
    contextType: string
    status: string
  }
  const showSearchBar = ref(false)
  const route = useRoute()
  const { targetAppKey } = useManagedAppScope()
  const activePackageType = ref<'base' | 'bundle'>('base')
  const dialogVisible = ref(false)
  const bundlesDialogVisible = ref(false)
  const actionsDialogVisible = ref(false)
  const menusDialogVisible = ref(false)
  const collaborationWorkspacesDialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const relationDialogVisible = ref(false)
  const relationLoading = ref(false)
  const relationKeyword = ref('')
  const relationTree = reactive<Api.SystemManage.FeaturePackageRelationTree>({
    roots: [],
    cycleDependencies: [],
    isolatedBaseKeys: []
  })
  const currentPackage = ref<Partial<PackageItem>>({})
  const routeOpenSignature = ref('')
  const loadError = ref('')
  const personalPackageCount = computed(
    () => data.value.filter((item) => supportsPersonalWorkspaceContext(item.contextType)).length
  )
  const collaborationPackageCount = computed(
    () =>
      data.value.filter((item) => supportsCollaborationWorkspaceContext(item.contextType)).length
  )
  const sharedPackageCount = computed(
    () => data.value.filter((item) => item.contextType === 'common').length
  )
  const bundleCount = computed(
    () => data.value.filter((item) => item.packageType === 'bundle').length
  )
  const totalActionCount = computed(() =>
    data.value.reduce((sum, item) => sum + (item.actionCount || 0), 0)
  )
  const totalMenuCount = computed(() =>
    data.value.reduce((sum, item) => sum + (item.menuCount || 0), 0)
  )
  const totalCollaborationWorkspaceCount = computed(() =>
    data.value.reduce((sum, item) => sum + (item.collaborationWorkspaceCount || 0), 0)
  )
  const disabledPackageCount = computed(
    () => data.value.filter((item) => item.status === 'disabled').length
  )
  const searchForm = reactive<SearchForm>({
    keyword: '',
    packageKey: '',
    name: '',
    contextType: '',
    status: ''
  })

  const contextTypeOptions = [
    { label: '全部空间范围', value: '' },
    { label: '个人空间功能包', value: 'personal' },
    { label: '协作空间功能包', value: 'collaboration' },
    { label: '通用功能包', value: 'common' }
  ]

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '正常', value: 'normal' },
    { label: '停用', value: 'disabled' }
  ]

  const searchItems = computed<FormItem[]>(() => [
    { label: '关键词', key: 'keyword', type: 'input', props: { placeholder: '名称/编码/描述' } },
    {
      label: '包编码',
      key: 'packageKey',
      type: 'input',
      props: { placeholder: '请输入包编码' }
    },
    { label: '包名称', key: 'name', type: 'input', props: { placeholder: '请输入包名称' } },
    {
      label: '空间范围',
      key: 'contextType',
      type: 'select',
      props: { options: contextTypeOptions, clearable: true }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: { options: statusOptions, clearable: true }
    }
  ])

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetFeaturePackageList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        { prop: 'packageKey', label: '功能包编码', minWidth: 220, showOverflowTooltip: true },
        { prop: 'name', label: '功能包名称', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'appKey',
          label: 'App',
          width: 150,
          formatter: (row: PackageItem) => row.appKey || '-'
        },
        {
          prop: 'packageType',
          label: '类型',
          width: 100,
          formatter: (row: PackageItem) =>
            h(ElTag, { type: row.packageType === 'bundle' ? 'warning' : 'success' }, () =>
              row.packageType === 'bundle' ? '组合包' : '基础包'
            )
        },
        {
          prop: 'contextType',
          label: '空间范围',
          width: 120,
          formatter: (row: PackageItem) =>
            h(
              ElTag,
              {
                type:
                  row.contextType === 'personal'
                    ? 'success'
                    : row.contextType === 'collaboration'
                      ? 'info'
                      : 'warning'
              },
              () => formatContextType(row.contextType)
            )
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 220,
          showOverflowTooltip: true,
          formatter: (row: PackageItem) => row.description || '-'
        },
        {
          prop: 'actionCount',
          label: '功能范围数',
          width: 100,
          formatter: (row: PackageItem) =>
            row.packageType === 'bundle' ? '-' : (row.actionCount ?? 0)
        },
        {
          prop: 'menuCount',
          label: '绑定菜单数',
          width: 96,
          formatter: (row: PackageItem) =>
            row.packageType === 'bundle' ? '-' : (row.menuCount ?? 0)
        },
        {
          prop: 'collaborationWorkspaceCount',
          label: '协作空间数',
          width: 90,
          formatter: (row: PackageItem) => row.collaborationWorkspaceCount ?? 0
        },
        {
          prop: 'sortOrder',
          label: '排序',
          width: 80,
          formatter: (row: PackageItem) => row.sortOrder ?? 0
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PackageItem) =>
            h(ElTag, { type: row.status === 'normal' ? 'success' : 'warning' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 140,
          fixed: 'right',
          formatter: (row: PackageItem) => {
            const list: ButtonMoreItem[] =
              row.packageType === 'bundle'
                ? [
                    {
                      key: 'bundles',
                      label: '配置基础包',
                      icon: 'ri:stack-line',
                      auth: 'feature_package.manage'
                    },
                    {
                      key: 'collaborationWorkspaces',
                      label: '开通协作空间',
                      icon: 'ri:group-line',
                      auth: 'feature_package.assign_collaboration_workspace',
                      disabled: !supportsCollaborationWorkspaceContext(row.contextType)
                    },
                    {
                      key: 'edit',
                      label: '编辑',
                      icon: 'ri:edit-2-line',
                      auth: 'feature_package.manage'
                    },
                    {
                      key: 'delete',
                      label: '删除',
                      icon: 'ri:delete-bin-4-line',
                      auth: 'feature_package.manage'
                    }
                  ]
                : [
                    {
                      key: 'actions',
                      label: '配置功能范围',
                      icon: 'ri:key-2-line',
                      auth: 'feature_package.manage'
                    },
                    {
                      key: 'menus',
                      label: '绑定菜单',
                      icon: 'ri:menu-line',
                      auth: 'feature_package.manage'
                    },
                    {
                      key: 'collaborationWorkspaces',
                      label: '开通协作空间',
                      icon: 'ri:group-line',
                      auth: 'feature_package.assign_collaboration_workspace',
                      disabled: !supportsCollaborationWorkspaceContext(row.contextType)
                    },
                    {
                      key: 'edit',
                      label: '编辑',
                      icon: 'ri:edit-2-line',
                      auth: 'feature_package.manage'
                    },
                    {
                      key: 'delete',
                      label: '删除',
                      icon: 'ri:delete-bin-4-line',
                      auth: 'feature_package.manage'
                    }
                  ]
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleAction(item.key as string, row)
            })
          }
        }
      ]
    },
    hooks: {
      onSuccess: () => {
        loadError.value = ''
      },
      onError: (error) => {
        loadError.value = error?.message || '获取功能包列表失败'
        ElMessage.error(loadError.value)
      }
    }
  })

  function normalizeSearchParams() {
    return {
      keyword: searchForm.keyword.trim() || undefined,
      packageKey: searchForm.packageKey.trim() || undefined,
      name: searchForm.name.trim() || undefined,
      packageType: activePackageType.value,
      contextType: searchForm.contextType || undefined,
      status: searchForm.status || undefined
    }
  }

  async function handleSearch() {
    Object.assign(searchParams, normalizeSearchParams())
    await getData()
  }

  async function handleReset() {
    Object.assign(searchForm, {
      keyword: '',
      packageKey: '',
      name: '',
      contextType: '',
      status: ''
    })
    Object.assign(searchParams, normalizeSearchParams())
    await getData()
  }

  async function handleRefresh() {
    await refreshData()
  }

  async function openRelationDialog() {
      if (!targetAppKey.value) {
        ElMessage.warning('请先进入具体 App 上下文后再查看包关系树')
        return
      }
    relationDialogVisible.value = true
    await loadRelationTree()
  }

  async function loadRelationTree() {
    if (!targetAppKey.value) {
      relationTree.roots = []
      relationTree.cycleDependencies = []
      relationTree.isolatedBaseKeys = []
      return
    }
    relationLoading.value = true
    try {
      const result = await fetchGetFeaturePackageRelationTree({
        appKey: targetAppKey.value,
        contextType: searchForm.contextType || undefined,
        keyword: relationKeyword.value.trim() || undefined
      })
      relationTree.roots = result.roots || []
      relationTree.cycleDependencies = result.cycleDependencies || []
      relationTree.isolatedBaseKeys = result.isolatedBaseKeys || []
    } catch (error: any) {
      ElMessage.error(error?.message || '加载功能包关系树失败')
    } finally {
      relationLoading.value = false
    }
  }

  async function syncRouteFilters() {
    activePackageType.value =
      normalizePackageType(String(route.query.tab || route.query.packageType || '')) || 'base'
    searchForm.keyword = String(route.query.keyword || '')
    searchForm.packageKey = String(route.query.packageKey || '')
    searchForm.name = String(route.query.name || '')
    searchForm.contextType = String(route.query.contextType || '')
    searchForm.status = String(route.query.status || '')
    Object.assign(searchParams, normalizeSearchParams())
    await getData()
    await openRouteTargetIfNeeded()
  }

  async function openRouteTargetIfNeeded() {
    const openMode = String(route.query.open || '')
    const packageKey = String(route.query.packageKey || '')
    const contextType = String(route.query.contextType || '')
    const packageType = normalizePackageType(
      String(route.query.tab || route.query.packageType || '')
    )
    if (!openMode || !packageKey) return

    const signature = `${openMode}|${packageKey}|${contextType}|${packageType}`
    if (routeOpenSignature.value === signature) return

    const target = data.value.find(
      (item) =>
        item.packageKey === packageKey &&
        (!contextType || item.contextType === contextType) &&
        (!packageType || item.packageType === packageType)
    )
    if (!target) return

    routeOpenSignature.value = signature
    currentPackage.value = { ...target }

    const requiresAppContext = ['bundles', 'menus', 'edit'].includes(openMode)
    if (requiresAppContext && !resolvePackageAppKey(target)) {
      return
    }

    if (openMode === 'bundles') {
      bundlesDialogVisible.value = true
      return
    }
    if (openMode === 'actions') {
      actionsDialogVisible.value = true
      return
    }
    if (openMode === 'menus') {
      menusDialogVisible.value = true
      return
    }
    if (openMode === 'collaborationWorkspaces') {
      collaborationWorkspacesDialogVisible.value = true
      return
    }
    if (openMode === 'edit') {
      openDialog('edit', target)
    }
  }

  function openDialog(type: 'add' | 'edit', row?: PackageItem) {
    dialogType.value = type
    currentPackage.value = row
      ? { ...row }
      : { packageType: activePackageType.value, appKey: targetAppKey.value || '' }
    dialogVisible.value = true
  }

  async function handleTabChange(name: string | number) {
    activePackageType.value = normalizePackageType(String(name)) || 'base'
    Object.assign(searchParams, normalizeSearchParams())
    await getData()
  }

  function handleAction(command: string, row: PackageItem) {
    if (command === 'bundles') {
      currentPackage.value = { ...row }
      bundlesDialogVisible.value = true
      return
    }
    if (command === 'actions') {
      currentPackage.value = { ...row }
      actionsDialogVisible.value = true
      return
    }
    if (command === 'menus') {
      currentPackage.value = { ...row }
      menusDialogVisible.value = true
      return
    }
    if (command === 'collaborationWorkspaces') {
      currentPackage.value = { ...row }
      collaborationWorkspacesDialogVisible.value = true
      return
    }
    if (command === 'edit') {
      openDialog('edit', row)
      return
    }
    if (command === 'delete') {
      fetchGetFeaturePackageImpactPreview(row.id)
        .then((impact) =>
          ElMessageBox.confirm(
            `删除后影响：角色 ${impact.roleCount}、协作空间 ${impact.collaborationWorkspaceCount}、用户 ${impact.userCount}。确认删除功能包「${row.name}」？`,
            '删除确认',
            {
              confirmButtonText: '确认删除',
              cancelButtonText: '取消',
              type: 'warning'
            }
          )
        )
        .then(() => fetchDeleteFeaturePackage(row.id))
        .then((stats) => {
          ElMessage.success(
            `本次增量刷新：角色 ${stats?.roleCount || 0}、协作空间 ${stats?.collaborationWorkspaceCount || 0}、用户 ${stats?.userCount || 0}、耗时 ${stats?.elapsedMilliseconds || 0} ms`
          )
          handleRefresh()
        })
        .catch((e) => {
          if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
        })
    }
  }

  watch(
    () => route.query,
    () => {
      routeOpenSignature.value = ''
      syncRouteFilters()
    },
    { immediate: true }
  )

  function supportsPersonalWorkspaceContext(contextType?: string) {
    return contextType === 'personal' || contextType === 'common'
  }

  function supportsCollaborationWorkspaceContext(contextType?: string) {
    return contextType === 'collaboration' || contextType === 'common'
  }

  function formatContextType(contextType?: string) {
    if (contextType === 'personal') return '个人空间'
    if (contextType === 'collaboration') return '协作空间'
    if (contextType === 'common') return '通用'
    return contextType || '-'
  }

  function normalizePackageType(value?: string) {
    return value === 'bundle' ? 'bundle' : value === 'base' ? 'base' : ''
  }

  function resolvePackageAppKey(row?: Partial<PackageItem>) {
    return `${row?.appKey || targetAppKey.value || ''}`.trim()
  }
</script>

<style scoped lang="scss">
  .feature-package-page {
    :deep(.art-search-bar .el-form-item__label) {
      white-space: nowrap;
    }
  }

  .feature-package-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .feature-package-app-select {
    width: 240px;
  }

  .package-tabs {
    margin-bottom: 0;
    padding: 0 22px;
    background: transparent;

    :deep(.el-tabs__header) {
      margin: 0;
    }

    :deep(.el-tabs__nav-wrap::after) {
      height: 1px;
      background-color: var(--art-border-soft);
    }

    :deep(.el-tabs__nav) {
      border: 0;
    }

    :deep(.el-tabs__item) {
      height: 42px;
      padding: 0 20px;
      border: 0;
      border-bottom: 2px solid transparent;
      background: transparent;
      color: var(--art-text-muted);
      font-size: 14px;
      font-weight: 600;
      transition:
        color 0.15s ease,
        border-color 0.15s ease,
        background-color 0.15s ease;
    }

    :deep(.el-tabs__item:hover) {
      color: var(--theme-color);
    }

    :deep(.el-tabs__item.is-active) {
      color: var(--theme-color);
      background: rgb(102 126 234 / 0.04);
      border-bottom-color: var(--theme-color);
    }
  }

  .feature-package-toolbar-tip {
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }

  .feature-package-error {
    margin-bottom: 12px;
  }

  .relation-toolbar {
    display: flex;
    gap: 12px;
    margin-bottom: 12px;
  }

  .relation-alert {
    margin-bottom: 8px;
  }

  .relation-node {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 32px;
  }

  .relation-node-name {
    min-width: 0;
    max-width: 280px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
