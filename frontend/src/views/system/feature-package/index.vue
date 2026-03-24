<template>
  <div class="feature-package-page art-full-height">
    <ElTabs v-model="activePackageType" class="package-tabs" @tab-change="handleTabChange">
      <ElTabPane label="基础包" name="base" />
      <ElTabPane label="组合包" name="bundle" />
    </ElTabs>

    <ArtSearchBar
      v-show="showSearchBar"
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <div class="stats-row">
      <ElCard shadow="never" class="stats-card">
        <div class="stats-label">当前页功能包数</div>
        <div class="stats-value">{{ data.length }}</div>
      </ElCard>
      <ElCard shadow="never" class="stats-card">
        <div class="stats-label">平台功能包</div>
        <div class="stats-value">{{ platformPackageCount }}</div>
      </ElCard>
      <ElCard shadow="never" class="stats-card">
        <div class="stats-label">团队功能包</div>
        <div class="stats-value">{{ teamPackageCount }}</div>
      </ElCard>
      <ElCard shadow="never" class="stats-card">
        <div class="stats-label">双上下文功能包</div>
        <div class="stats-value">{{ sharedPackageCount }}</div>
      </ElCard>
      <ElCard shadow="never" class="stats-card">
        <div class="stats-label">{{ activePackageType === 'base' ? '已组合功能范围数' : '组合包数量' }}</div>
        <div class="stats-value">{{ activePackageType === 'base' ? totalActionCount : bundleCount }}</div>
      </ElCard>
      <ElCard shadow="never" class="stats-card">
        <div class="stats-label">{{ activePackageType === 'base' ? '已绑定菜单数' : '团队开通数' }}</div>
        <div class="stats-value">{{ activePackageType === 'base' ? totalMenuCount : totalTeamCount }}</div>
      </ElCard>
      <ElCard shadow="never" class="stats-card">
        <div class="stats-label">停用功能包</div>
        <div class="stats-value">{{ disabledPackageCount }}</div>
      </ElCard>
    </div>

    <ElCard class="art-table-card" shadow="never" :style="{ marginTop: showSearchBar ? '12px' : '0' }">
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="handleRefresh"
      >
        <template #left>
          <ElButton v-action="'platform.package.manage'" type="primary" @click="openDialog('add')" v-ripple>
            新增{{ activePackageType === 'base' ? '基础包' : '组合包' }}
          </ElButton>
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
      :default-package-type="activePackageType"
      @success="handleRefresh"
    />

    <FeaturePackageBundlesDialog
      v-model="bundlesDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'team'"
      @success="handleRefresh"
    />

    <FeaturePackageActionsDialog
      v-model="actionsDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'team'"
      @success="handleRefresh"
    />

    <FeaturePackageMenusDialog
      v-model="menusDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'team'"
      @success="handleRefresh"
    />

    <FeaturePackageTeamsDialog
      v-model="teamsDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'team'"
      @success="handleRefresh"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, h, reactive, ref, watch } from 'vue'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useRoute } from 'vue-router'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchDeleteFeaturePackage, fetchGetFeaturePackageList } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import FeaturePackageDialog from './modules/feature-package-dialog.vue'
  import FeaturePackageBundlesDialog from './modules/feature-package-bundles-dialog.vue'
  import FeaturePackageActionsDialog from './modules/feature-package-actions-dialog.vue'
  import FeaturePackageMenusDialog from './modules/feature-package-menus-dialog.vue'
  import FeaturePackageTeamsDialog from './modules/feature-package-teams-dialog.vue'

  defineOptions({ name: 'FeaturePackage' })

  type PackageItem = Api.SystemManage.FeaturePackageItem

  type SearchForm = {
    keyword: string
    packageKey: string
    name: string
    contextType: string
    status: string
  }
  const showSearchBar = ref(true)
  const route = useRoute()
  const activePackageType = ref<'base' | 'bundle'>('base')
  const dialogVisible = ref(false)
  const bundlesDialogVisible = ref(false)
  const actionsDialogVisible = ref(false)
  const menusDialogVisible = ref(false)
  const teamsDialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentPackage = ref<Partial<PackageItem>>({})
  const routeOpenSignature = ref('')
  const platformPackageCount = computed(() => data.value.filter((item) => supportsPlatform(item.contextType)).length)
  const teamPackageCount = computed(() => data.value.filter((item) => supportsTeam(item.contextType)).length)
  const sharedPackageCount = computed(() => data.value.filter((item) => item.contextType === 'common').length)
  const bundleCount = computed(() => data.value.filter((item) => item.packageType === 'bundle').length)
  const totalActionCount = computed(() => data.value.reduce((sum, item) => sum + (item.actionCount || 0), 0))
  const totalMenuCount = computed(() => data.value.reduce((sum, item) => sum + (item.menuCount || 0), 0))
  const totalTeamCount = computed(() => data.value.reduce((sum, item) => sum + (item.teamCount || 0), 0))
  const disabledPackageCount = computed(() => data.value.filter((item) => item.status === 'disabled').length)

  const searchForm = reactive<SearchForm>({
    keyword: '',
    packageKey: '',
    name: '',
    contextType: '',
    status: ''
  })

  const contextTypeOptions = [
    { label: '全部上下文', value: '' },
    { label: '平台功能包', value: 'platform' },
    { label: '团队功能包', value: 'team' },
    { label: '通用功能包', value: 'common' }
  ]

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '正常', value: 'normal' },
    { label: '停用', value: 'disabled' }
  ]

  const searchItems = computed<FormItem[]>(() => [
    { label: '关键词', key: 'keyword', type: 'input', props: { placeholder: '名称/编码/描述' } },
    { label: '功能包编码', key: 'packageKey', type: 'input', props: { placeholder: '请输入功能包编码' } },
    { label: '功能包名称', key: 'name', type: 'input', props: { placeholder: '请输入功能包名称' } },
    {
      label: '上下文类型',
      key: 'contextType',
      type: 'select',
      props: { options: contextTypeOptions, clearable: true }
    },
    { label: '状态', key: 'status', type: 'select', props: { options: statusOptions, clearable: true } }
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
          label: '上下文',
          width: 120,
          formatter: (row: PackageItem) =>
            h(
              ElTag,
              { type: row.contextType === 'platform' ? 'success' : row.contextType === 'team' ? 'info' : 'warning' },
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
          formatter: (row: PackageItem) => (row.packageType === 'bundle' ? '-' : row.actionCount ?? 0)
        },
        {
          prop: 'menuCount',
          label: '绑定菜单数',
          width: 96,
          formatter: (row: PackageItem) => (row.packageType === 'bundle' ? '-' : row.menuCount ?? 0)
        },
        { prop: 'teamCount', label: '团队数', width: 90, formatter: (row: PackageItem) => row.teamCount ?? 0 },
        { prop: 'sortOrder', label: '排序', width: 80, formatter: (row: PackageItem) => row.sortOrder ?? 0 },
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
                      auth: 'platform.package.manage'
                    },
                    {
                      key: 'teams',
                      label: '开通团队',
                      icon: 'ri:team-line',
                      auth: 'platform.package.assign',
                      disabled: !supportsTeam(row.contextType)
                    },
                    { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'platform.package.manage' },
                    { key: 'delete', label: '删除', icon: 'ri:delete-bin-4-line', auth: 'platform.package.manage' }
                  ]
                : [
                    {
                      key: 'actions',
                      label: '配置功能范围',
                      icon: 'ri:key-2-line',
                      auth: 'platform.package.manage'
                    },
                    {
                      key: 'menus',
                      label: '绑定菜单',
                      icon: 'ri:menu-line',
                      auth: 'platform.package.manage'
                    },
                    {
                      key: 'teams',
                      label: '开通团队',
                      icon: 'ri:team-line',
                      auth: 'platform.package.assign',
                      disabled: !supportsTeam(row.contextType)
                    },
                    { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'platform.package.manage' },
                    { key: 'delete', label: '删除', icon: 'ri:delete-bin-4-line', auth: 'platform.package.manage' }
                  ]
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleAction(item.key as string, row)
            })
          }
        }
      ]
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

  async function syncRouteFilters() {
    activePackageType.value = normalizePackageType(String(route.query.tab || route.query.packageType || '')) || 'base'
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
    const packageType = normalizePackageType(String(route.query.tab || route.query.packageType || ''))
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
    if (openMode === 'teams') {
      teamsDialogVisible.value = true
      return
    }
    if (openMode === 'edit') {
      openDialog('edit', target)
    }
  }

  function openDialog(type: 'add' | 'edit', row?: PackageItem) {
    dialogType.value = type
    currentPackage.value = row ? { ...row } : { packageType: activePackageType.value }
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
    if (command === 'teams') {
      currentPackage.value = { ...row }
      teamsDialogVisible.value = true
      return
    }
    if (command === 'edit') {
      openDialog('edit', row)
      return
    }
    if (command === 'delete') {
      ElMessageBox.confirm(`确定删除功能包「${row.name}」吗？`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() =>
          fetchDeleteFeaturePackage(row.id)
        )
        .then(() => {
          ElMessage.success('删除成功')
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

  function supportsPlatform(contextType?: string) {
    return contextType === 'platform' || contextType === 'common'
  }

  function supportsTeam(contextType?: string) {
    return contextType === 'team' || contextType === 'common'
  }

  function formatContextType(contextType?: string) {
    if (contextType === 'platform') return '平台'
    if (contextType === 'team') return '团队'
    if (contextType === 'common') return '通用'
    return contextType || '-'
  }

  function normalizePackageType(value?: string) {
    return value === 'bundle' ? 'bundle' : value === 'base' ? 'base' : ''
  }
</script>

<style scoped lang="scss">
  .package-tabs {
    margin-bottom: 12px;
  }

  .stats-row {
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    gap: 12px;
  }

  .stats-card {
    min-height: 112px;
  }

  .stats-label {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .stats-value {
    margin-top: 10px;
    font-size: 30px;
    font-weight: 700;
    line-height: 1.1;
    color: var(--el-text-color-primary);
  }

  @media (max-width: 1200px) {
    .stats-row {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 768px) {
    .stats-row {
      grid-template-columns: 1fr;
    }
  }
</style>
