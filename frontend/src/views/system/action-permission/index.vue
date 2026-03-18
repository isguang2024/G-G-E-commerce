<template>
  <div class="art-full-height">
    <ArtSearchBar
      v-show="showSearchBar"
      v-model="searchForm"
      :items="searchItems"
      @search="handleSearch"
      @reset="handleReset"
    >
      <template #category>
        <ElAutocomplete
          v-model="searchForm.category"
          :fetch-suggestions="queryCategorySuggestions"
          clearable
          placeholder="输入或选择历史分类"
        />
      </template>
      <template #moduleCode>
        <ElAutocomplete
          v-model="searchForm.moduleCode"
          :fetch-suggestions="queryModuleSuggestions"
          clearable
          placeholder="输入或选择模块归属"
        />
      </template>
    </ArtSearchBar>

    <ElRow :gutter="12" class="stats-row">
      <ElCol :xs="24" :sm="12" :md="8" :lg="6">
        <ElCard shadow="never" class="stats-card">
          <div class="stats-label">当前筛选权限数</div>
          <div class="stats-value">{{ stats.total }}</div>
          <div class="stats-help">按当前搜索条件汇总</div>
        </ElCard>
      </ElCol>
      <ElCol :xs="24" :sm="12" :md="8" :lg="6">
        <ElCard shadow="never" class="stats-card">
          <div class="stats-label">来源分布</div>
          <div class="stats-tags">
            <ElTag v-for="item in stats.sourceTags" :key="item.label" :type="item.type" effect="light">
              {{ item.label }} {{ item.count }}
            </ElTag>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :xs="24" :sm="12" :md="8" :lg="6">
        <ElCard shadow="never" class="stats-card">
          <div class="stats-label">功能归属</div>
          <div class="stats-tags">
            <ElTag v-for="item in stats.featureKindTags" :key="item.label" type="success" effect="light">
              {{ item.label }} {{ item.count }}
            </ElTag>
            <span v-if="stats.featureKindTags.length === 0" class="stats-empty">暂无归属</span>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :xs="24" :sm="12" :md="8" :lg="6">
        <ElCard shadow="never" class="stats-card">
          <div class="stats-label">分类分布</div>
          <div class="stats-tags">
            <ElTag v-for="item in stats.categoryTags" :key="item.label" type="info" effect="light">
              {{ item.label }} {{ item.count }}
            </ElTag>
            <span v-if="stats.categoryTags.length === 0" class="stats-empty">暂无分类</span>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :xs="24" :sm="12" :md="8" :lg="6">
        <ElCard shadow="never" class="stats-card">
          <div class="stats-label">模块归属分布</div>
          <div class="stats-tags">
            <ElTag v-for="item in stats.moduleTags" :key="item.label" type="primary" effect="light">
              {{ item.label }} {{ item.count }}
            </ElTag>
            <span v-if="stats.moduleTags.length === 0" class="stats-empty">暂无模块</span>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :xs="24" :sm="12" :md="8" :lg="6">
        <ElCard shadow="never" class="stats-card">
          <div class="stats-label">资源编码分布</div>
          <div class="stats-tags">
            <ElTag v-for="item in stats.resourceTags" :key="item.label" type="warning" effect="light">
              {{ item.label }} {{ item.count }}
            </ElTag>
            <span v-if="stats.resourceTags.length === 0" class="stats-empty">暂无资源</span>
          </div>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElCard
      class="art-table-card"
      shadow="never"
      :style="{ 'margin-top': '12px' }"
    >
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="handleRefresh"
      >
        <template #left>
          <ElButton v-action="'permission_action:create'" type="primary" @click="openDialog('add')" v-ripple>
            新增功能权限
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

    <ActionPermissionDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :action-data="currentAction"
      :category-options="categoryOptions"
      :module-options="moduleOptions"
      @success="handleRefresh"
    />
  </div>
</template>

<script setup lang="ts">
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchDeletePermissionAction,
    fetchGetAllScopes,
    fetchGetPermissionActionList
  } from '@/api/system-manage'
  import { formatScopeLabel, getScopeTagType } from '@/utils/permission/scope'
  import ActionPermissionDialog from './modules/action-permission-dialog.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'ActionPermission' })

  type PermissionActionItem = Api.SystemManage.PermissionActionItem
  type TenantContextFilter = '' | 'true' | 'false' | undefined
  type SuggestionItem = { value: string }
  type StatsTag = {
    label: string
    count: number
    type?: 'success' | 'info' | 'warning'
  }

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentAction = ref<PermissionActionItem | undefined>()
  const showSearchBar = ref(true)
  const categoryOptions = ref<string[]>([])
  const moduleOptions = ref<string[]>([])
  const stats = reactive<{
    total: number
    sourceTags: StatsTag[]
    featureKindTags: StatsTag[]
    categoryTags: StatsTag[]
    moduleTags: StatsTag[]
    resourceTags: StatsTag[]
  }>({
    total: 0,
    sourceTags: [],
    featureKindTags: [],
    categoryTags: [],
    moduleTags: [],
    resourceTags: []
  })

  const searchForm = reactive<{
    keyword: string
    source: string
    featureKind: string
    moduleCode: string
    category: string
    scopeCode: string
    status: string
    requiresTenantContext: TenantContextFilter
  }>({
    keyword: '',
    source: '',
    featureKind: '',
    moduleCode: '',
    category: '',
    scopeCode: '',
    status: '',
    requiresTenantContext: undefined
  })

  const sourceOptions = [
    { label: '全部来源', value: '' },
    { label: '接口自动注册', value: 'api' },
    { label: '系统内置', value: 'system' },
    { label: '业务定义', value: 'business' }
  ]

  const featureKindOptions = [
    { label: '全部归属', value: '' },
    { label: '系统功能', value: 'system' },
    { label: '业务功能', value: 'business' }
  ]

  const scopeOptions = ref([{ label: '全部作用域', value: '' }])

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '正常', value: 'normal' },
    { label: '停用', value: 'suspended' }
  ]

  const tenantContextOptions = [
    { label: '全部', value: '' },
    { label: '依赖团队', value: 'true' },
    { label: '不依赖团队', value: 'false' }
  ]

  onMounted(async () => {
    try {
      const scopes = await fetchGetAllScopes()
      scopeOptions.value = [
        { label: '全部作用域', value: '' },
        ...scopes.map((scope) => ({
          label: scope.scopeName || scope.scopeCode,
          value: scope.scopeCode
        }))
      ]
    } catch {
      scopeOptions.value = [{ label: '全部作用域', value: '' }]
    }
  })

  const searchItems = computed<FormItem[]>(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: { placeholder: '名称/描述/资源编码/动作编码/分类' }
    },
    {
      label: '来源',
      key: 'source',
      type: 'select',
      props: { options: sourceOptions, clearable: true }
    },
    {
      label: '模块归属',
      key: 'moduleCode',
      type: 'input',
      props: { placeholder: '输入或选择模块归属' }
    },
    {
      label: '分类',
      key: 'category',
      type: 'input',
      props: { placeholder: '输入或选择历史分类' }
    },
    {
      label: '功能归属',
      key: 'featureKind',
      type: 'select',
      props: { options: featureKindOptions, clearable: true }
    },
    {
      label: '作用域',
      key: 'scopeCode',
      type: 'select',
      props: { options: scopeOptions.value, clearable: true }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: { options: statusOptions, clearable: true }
    },
    {
      label: '依赖团队',
      key: 'requiresTenantContext',
      type: 'select',
      props: { options: tenantContextOptions, clearable: true }
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
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetPermissionActionList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        { prop: 'name', label: '权限名称', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'permissionKey',
          label: '权限键',
          minWidth: 220,
          formatter: (row: PermissionActionItem) => row.permissionKey || `${row.resourceCode}:${row.actionCode}`
        },
        {
          prop: 'moduleCode',
          label: '模块归属',
          minWidth: 120,
          formatter: (row: PermissionActionItem) => row.moduleCode || '-'
        },
        {
          prop: 'category',
          label: '分类',
          minWidth: 120,
          formatter: (row: PermissionActionItem) => row.category || '-'
        },
        {
          prop: 'featureKind',
          label: '功能归属',
          width: 110,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.featureKind === 'business' ? 'success' : 'info' }, () =>
              row.featureKind === 'business' ? '业务功能' : '系统功能'
            )
        },
        {
          prop: 'source',
          label: '来源',
          width: 110,
          formatter: (row: PermissionActionItem) => {
            const sourceConfig =
              row.source === 'api'
                ? { type: 'success', text: '接口自动' }
                : row.source === 'system'
                  ? { type: 'info', text: '系统内置' }
                  : { type: 'warning', text: '业务定义' }
            return h(ElTag, { type: sourceConfig.type as 'success' | 'info' | 'warning' }, () => sourceConfig.text)
          }
        },
        {
          prop: 'scopeName',
          label: '作用域',
          width: 100,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: getScopeTagType(row.scopeCode) }, () =>
              formatScopeLabel(row.scopeCode, row.scopeName)
            )
        },
        {
          prop: 'requiresTenantContext',
          label: '依赖团队',
          width: 100,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.requiresTenantContext ? 'warning' : 'info' }, () =>
              row.requiresTenantContext ? '是' : '否'
            )
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 180,
          showOverflowTooltip: true,
          formatter: (row: PermissionActionItem) => row.description || '-'
        },
        { prop: 'sortOrder', label: '排序', width: 80 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PermissionActionItem) =>
            h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 70,
          fixed: 'right',
          formatter: (row: PermissionActionItem) => {
            const list: ButtonMoreItem[] = [
              {
                key: 'edit',
                label: '编辑',
                icon: 'ri:edit-2-line',
                auth: 'permission_action:update'
              }
            ]
            if (row.source !== 'api' && row.source !== 'system') {
              list.push({
                key: 'delete',
                label: '删除',
                icon: 'ri:delete-bin-4-line',
                auth: 'permission_action:delete'
              })
            }
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleAction(item.key as string, row)
            })
          }
        }
      ]
    },
  })

  function normalizeSearchParams() {
    return {
      keyword: searchForm.keyword?.trim() || undefined,
      source: searchForm.source || undefined,
      featureKind: searchForm.featureKind || undefined,
      moduleCode: searchForm.moduleCode?.trim() || undefined,
      category: searchForm.category?.trim() || undefined,
      scopeCode: searchForm.scopeCode || undefined,
      status: searchForm.status || undefined,
      requiresTenantContext:
        searchForm.requiresTenantContext === 'true'
          ? true
          : searchForm.requiresTenantContext === 'false'
            ? false
            : undefined
    }
  }

  function buildTopTags(items: PermissionActionItem[], getter: (item: PermissionActionItem) => string) {
    const counter = new Map<string, number>()
    items.forEach((item) => {
      const key = getter(item).trim()
      if (!key) return
      counter.set(key, (counter.get(key) || 0) + 1)
    })
    return [...counter.entries()]
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0], 'zh-CN'))
      .slice(0, 6)
      .map(([label, count]) => ({ label, count }))
  }

  function updateCategoryOptions(items: PermissionActionItem[]) {
    const nextOptions = new Set(categoryOptions.value)
    const nextModuleOptions = new Set(moduleOptions.value)
    items.forEach((item) => {
      const moduleCode = item.moduleCode?.trim()
      if (moduleCode) {
        nextModuleOptions.add(moduleCode)
      }
      const category = item.category?.trim()
      if (category) {
        nextOptions.add(category)
      }
    })
    categoryOptions.value = [...nextOptions].sort((a, b) => a.localeCompare(b, 'zh-CN'))
    moduleOptions.value = [...nextModuleOptions].sort((a, b) => a.localeCompare(b, 'zh-CN'))
  }

  async function loadStats() {
    const res = await fetchGetPermissionActionList({
      current: 1,
      size: 1000,
      ...normalizeSearchParams()
    })
    const records = res.records || []
    updateCategoryOptions(records)
    stats.total = res.total || records.length
    const sourceCount = {
      api: records.filter((item) => item.source === 'api').length,
      business: records.filter((item) => item.source === 'business').length,
      system: records.filter((item) => item.source === 'system').length
    }
    stats.sourceTags = [
      { label: '接口自动', count: sourceCount.api, type: 'success' },
      { label: '系统内置', count: sourceCount.system, type: 'info' },
      { label: '业务定义', count: sourceCount.business, type: 'warning' }
    ] as StatsTag[]
    stats.sourceTags = stats.sourceTags.filter((item) => item.count > 0)
    stats.featureKindTags = [
      { label: '系统功能', count: records.filter((item) => item.featureKind !== 'business').length },
      { label: '业务功能', count: records.filter((item) => item.featureKind === 'business').length }
    ].filter((item) => item.count > 0)
    stats.categoryTags = buildTopTags(records, (item) => item.category || '')
    stats.moduleTags = buildTopTags(records, (item) => item.moduleCode || '')
    stats.resourceTags = buildTopTags(records, (item) => item.resourceCode || '')
  }

  function queryCategorySuggestions(queryString: string, cb: (items: SuggestionItem[]) => void) {
    const keyword = queryString.trim().toLowerCase()
    const suggestions = categoryOptions.value
      .filter((item) => !keyword || item.toLowerCase().includes(keyword))
      .slice(0, 12)
      .map((value) => ({ value }))
    cb(suggestions)
  }

  function queryModuleSuggestions(queryString: string, cb: (items: SuggestionItem[]) => void) {
    const keyword = queryString.trim().toLowerCase()
    const suggestions = moduleOptions.value
      .filter((item) => !keyword || item.toLowerCase().includes(keyword))
      .slice(0, 12)
      .map((value) => ({ value }))
    cb(suggestions)
  }

  async function handleSearch() {
    Object.assign(searchParams, normalizeSearchParams())
    await getData()
    await loadStats()
  }

  async function handleReset() {
    Object.assign(searchForm, {
      keyword: '',
      source: '',
      featureKind: '',
      moduleCode: '',
      category: '',
      scopeCode: '',
      status: '',
      requiresTenantContext: undefined
    })
    await resetSearchParams()
    await loadStats()
  }

  async function handleRefresh() {
    await refreshData()
    await loadStats()
  }

  function openDialog(type: 'add' | 'edit', row?: PermissionActionItem) {
    dialogType.value = type
    currentAction.value = row
    dialogVisible.value = true
  }

  function handleAction(command: string, row: PermissionActionItem) {
    if (command === 'edit') {
      openDialog('edit', row)
      return
    }
    ElMessageBox.confirm(`确定删除功能权限「${row.name}」吗？`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeletePermissionAction(row.id))
      .then(() => {
        ElMessage.success('删除成功')
        handleRefresh()
      })
      .catch((e) => {
        if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
      })
  }

  onMounted(() => {
    loadStats()
  })
</script>

<style scoped>
  .stats-row {
    margin-top: 12px;
  }

  .stats-card {
    min-height: 128px;
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

  .stats-help {
    margin-top: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .stats-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }

  .stats-empty {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
</style>
