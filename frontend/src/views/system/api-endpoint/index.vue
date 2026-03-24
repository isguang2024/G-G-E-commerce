<template>
  <div class="art-full-height">
    <ElCard class="module-card" shadow="never">
      <div class="module-header">
        <div class="module-title">模块分类</div>
        <div class="module-help">接口注册按模块展示，Method 与路径独立显示，便于核对注册配置</div>
      </div>
      <div class="module-tags">
        <ElTag
          :type="selectedModule === '' ? 'primary' : 'info'"
          effect="light"
          class="module-tag"
          @click="handleModuleSelect('')"
        >
          全部 {{ totalCount }}
        </ElTag>
        <ElTag
          v-for="item in moduleSummary"
          :key="item.label"
          :type="selectedModule === item.label ? 'primary' : 'info'"
          effect="light"
          class="module-tag"
          @click="handleModuleSelect(item.label)"
        >
          {{ item.label }} {{ item.count }}
        </ElTag>
      </div>

      <div class="status-overview">
        <div class="status-card is-sync">
          <div class="status-label">自动注册</div>
          <div class="status-value">{{ registrySummary.sync }}</div>
          <div class="status-help">带元数据并已进入 API 单元</div>
        </div>
        <div class="status-card is-manual">
          <div class="status-label">手工补录</div>
          <div class="status-value">{{ registrySummary.manual }}</div>
          <div class="status-help">后台手动创建并绑定路由</div>
        </div>
        <div class="status-card is-seed">
          <div class="status-label">初始种子</div>
          <div class="status-value">{{ registrySummary.seed }}</div>
          <div class="status-help">部署初始化直接导入</div>
        </div>
        <div class="status-card is-pending">
          <div class="status-label">未注册路由</div>
          <div class="status-value">{{ unregisteredCount }}</div>
          <div class="status-help">运行时存在但尚未进入 API 单元</div>
        </div>
      </div>

      <div class="category-panel">
        <div class="category-panel-header">
          <div class="module-title">分类管理</div>
          <div class="module-help"
            >直接在 API 管理页维护分类，停用后保留历史归属但不建议继续分配。</div
          >
        </div>
        <div v-if="sortedCategories.length" class="category-list">
          <div
            v-for="item in sortedCategories"
            :key="item.id"
            class="category-card"
            :class="{ 'is-suspended': item.status === 'suspended' }"
          >
            <div class="category-card-main">
              <div class="category-card-title">
                <span>{{ item.name }}</span>
                <ElTag
                  size="small"
                  :type="item.status === 'normal' ? 'success' : 'info'"
                  effect="plain"
                >
                  {{ item.status === 'normal' ? '正常' : '停用' }}
                </ElTag>
              </div>
              <div class="category-card-meta">{{ item.code }} / {{ item.nameEn }}</div>
            </div>
            <div class="category-card-actions">
              <ElButton text type="primary" @click="openCategoryDialog(item)">编辑</ElButton>
              <ElButton
                text
                :type="item.status === 'normal' ? 'danger' : 'success'"
                :loading="categorySwitchingId === item.id"
                @click="toggleCategoryStatus(item)"
              >
                {{ item.status === 'normal' ? '停用' : '启用' }}
              </ElButton>
            </div>
          </div>
        </div>
        <div v-else class="category-empty">暂无分类，可直接新建。</div>
      </div>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton
            v-action="'system.api_registry.sync'"
            type="primary"
            @click="openCreateDialog"
            v-ripple
          >
            新增 API
          </ElButton>
          <ElButton v-action="'system.api_registry.sync'" @click="openCategoryDialog()" v-ripple>
            新建分类
          </ElButton>
          <ElButton
            v-action="'system.api_registry.sync'"
            type="primary"
            plain
            :loading="syncing"
            @click="handleSync"
            v-ripple
          >
            同步 API
          </ElButton>
          <ElButton v-action="'system.api_registry.view'" plain @click="openUnregisteredDialog" v-ripple>
            未注册 API
            <span v-if="unregisteredCount > 0" class="toolbar-count">({{ unregisteredCount }})</span>
          </ElButton>
        </template>
      </ArtTableHeader>

      <div class="table-query-panel">
        <ElSelect v-model="tableQuery.method" clearable placeholder="Method">
          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
        </ElSelect>
        <ElInput v-model="tableQuery.path" clearable placeholder="按路径搜索" />
        <ElInput v-model="tableQuery.keyword" clearable placeholder="按摘要/处理器/模块搜索" />
        <ElInput v-model="tableQuery.permissionKey" clearable placeholder="按权限键搜索" />
        <ElSelect v-model="tableQuery.categoryId" clearable filterable placeholder="分类">
          <ElOption
            v-for="item in sortedCategories"
            :key="item.id"
            :label="`${item.name} / ${item.nameEn}`"
            :value="item.id"
          />
        </ElSelect>
        <ElSelect v-model="tableQuery.contextScope" clearable placeholder="团队上下文">
          <ElOption label="可选" value="optional" />
          <ElOption label="必需" value="required" />
          <ElOption label="禁止" value="forbidden" />
        </ElSelect>
        <ElSelect v-model="tableQuery.featureKind" clearable placeholder="功能归属">
          <ElOption label="系统" value="system" />
          <ElOption label="业务" value="business" />
        </ElSelect>
        <ElSelect v-model="tableQuery.status" clearable placeholder="状态">
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="suspended" />
        </ElSelect>
        <ElSelect v-model="tableQuery.hasPermissionKey" clearable placeholder="权限键">
          <ElOption label="有权限键" value="true" />
          <ElOption label="无权限键" value="false" />
        </ElSelect>
        <ElSelect v-model="tableQuery.hasCategory" clearable placeholder="分类归属">
          <ElOption label="已分配分类" value="true" />
          <ElOption label="未分配分类" value="false" />
        </ElSelect>
        <div class="table-query-actions">
          <ElButton type="primary" @click="handleTableSearch">查询</ElButton>
          <ElButton @click="resetTableQuery">重置</ElButton>
        </div>
      </div>

      <div class="source-filter-panel">
        <div class="source-filter-title">注册方式</div>
        <div class="source-filter-tags">
          <ElTag
            :type="selectedSource === '' ? 'primary' : 'info'"
            effect="light"
            class="module-tag"
            @click="handleSourceSelect('')"
          >
            全部 {{ totalCount }}
          </ElTag>
          <ElTag
            :type="selectedSource === 'sync' ? 'primary' : 'info'"
            effect="light"
            class="module-tag"
            @click="handleSourceSelect('sync')"
          >
            自动注册 {{ registrySummary.sync }}
          </ElTag>
          <ElTag
            :type="selectedSource === 'manual' ? 'primary' : 'info'"
            effect="light"
            class="module-tag"
            @click="handleSourceSelect('manual')"
          >
            手工补录 {{ registrySummary.manual }}
          </ElTag>
          <ElTag
            :type="selectedSource === 'seed' ? 'primary' : 'info'"
            effect="light"
            class="module-tag"
            @click="handleSourceSelect('seed')"
          >
            初始种子 {{ registrySummary.seed }}
          </ElTag>
        </div>
      </div>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDialog
      v-model="formVisible"
      :title="editingId ? '编辑 API' : '新增 API'"
      width="760px"
      destroy-on-close
    >
      <ElForm :model="formState" label-width="110px">
        <ElRow :gutter="12">
          <ElCol :span="8">
            <ElFormItem label="Method" prop="method">
              <ElSelect v-model="formState.method" placeholder="请选择">
                <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem label="功能归属" prop="featureKind">
              <ElSelect v-model="formState.featureKind" placeholder="请选择">
                <ElOption label="系统" value="system" />
                <ElOption label="业务" value="business" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem label="来源" prop="source">
              <ElSelect v-model="formState.source" placeholder="请选择">
                <ElOption label="自动同步" value="sync" />
                <ElOption label="初始种子" value="seed" />
                <ElOption label="手工维护" value="manual" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElFormItem label="路径" prop="path">
          <ElInput v-model="formState.path" placeholder="/api/v1/..." />
        </ElFormItem>

        <ElRow :gutter="12">
          <ElCol :span="12">
            <ElFormItem label="模块" prop="module">
              <ElInput v-model="formState.module" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="分类" prop="categoryId">
              <div class="category-input-wrap">
                <ElSelect
                  v-model="formState.categoryId"
                  clearable
                  filterable
                  placeholder="请选择分类"
                >
                  <ElOption
                    v-for="item in sortedCategories"
                    :key="item.id"
                    :label="`${item.name} / ${item.nameEn}${item.status === 'suspended' ? '（已停用）' : ''}`"
                    :value="item.id"
                    :disabled="item.status === 'suspended' && formState.categoryId !== item.id"
                  />
                </ElSelect>
                <ElButton text type="primary" @click="openCategoryDialog()">新建分类</ElButton>
              </div>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="12">
          <ElCol :span="12">
            <ElFormItem label="团队上下文" prop="contextScope">
              <ElSelect v-model="formState.contextScope" placeholder="请选择">
                <ElOption label="可选" value="optional" />
                <ElOption label="必需" value="required" />
                <ElOption label="禁止" value="forbidden" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="状态" prop="status">
              <ElSelect v-model="formState.status" placeholder="请选择">
                <ElOption label="正常" value="normal" />
                <ElOption label="停用" value="suspended" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElFormItem label="说明" prop="summary">
          <ElInput v-model="formState.summary" />
        </ElFormItem>

        <ElFormItem label="权限键">
          <ElSelect
            v-model="formState.permissionKeys"
            multiple
            filterable
            allow-create
            default-first-option
          >
            <ElOption
              v-for="item in formState.permissionKeys"
              :key="item"
              :label="item"
              :value="item"
            />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="formVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="saving" @click="submitForm">保存</ElButton>
      </template>
    </ElDialog>

    <ElDialog
      v-model="categoryDialogVisible"
      :title="categoryForm.id ? '编辑分类' : '新建分类'"
      width="560px"
      destroy-on-close
    >
      <ElForm :model="categoryForm" label-width="100px">
        <ElFormItem label="分类编码">
          <ElInput v-model="categoryForm.code" placeholder="例如 system_manage" />
        </ElFormItem>
        <ElFormItem label="中文名称">
          <ElInput v-model="categoryForm.name" placeholder="例如 系统管理" />
        </ElFormItem>
        <ElFormItem label="英文名称">
          <ElInput v-model="categoryForm.nameEn" placeholder="例如 System Management" />
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber
            v-model="categoryForm.sortOrder"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="categoryForm.status" placeholder="请选择">
            <ElOption label="正常" value="normal" />
            <ElOption label="停用" value="suspended" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="categoryDialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="categorySaving" @click="submitCategory">保存</ElButton>
      </template>
    </ElDialog>

    <ElDialog
      v-model="unregisteredVisible"
      title="未注册 API"
      width="980px"
      destroy-on-close
    >
      <div class="unregistered-toolbar">
        <ElSelect v-model="unregisteredQuery.method" clearable placeholder="Method" style="width: 120px">
          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
        </ElSelect>
        <ElInput v-model="unregisteredQuery.path" placeholder="按路径筛选" clearable />
        <ElInput v-model="unregisteredQuery.module" placeholder="按模块筛选" clearable />
        <ElInput v-model="unregisteredQuery.keyword" placeholder="按摘要或处理器搜索" clearable />
        <ElCheckbox v-model="unregisteredQuery.onlyNoMeta">仅看无元数据</ElCheckbox>
        <ElButton type="primary" :loading="unregisteredLoading" @click="handleUnregisteredSearch">
          查询
        </ElButton>
        <ElButton @click="resetUnregisteredQuery">重置</ElButton>
      </div>

      <ElTable :data="unregisteredRoutes" :loading="unregisteredLoading" border height="420px">
        <ElTableColumn label="Method" width="92">
          <template #default="{ row }">
            <ElTag :type="methodTagType(row.method)" effect="dark">{{ row.method }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="path" label="路径" min-width="280" show-overflow-tooltip />
        <ElTableColumn prop="module" label="模块" width="120" show-overflow-tooltip />
        <ElTableColumn label="元数据" width="120">
          <template #default="{ row }">
            <ElTag :type="row.hasMeta ? 'success' : 'info'" effect="plain">
              {{ row.hasMeta ? '已声明' : '未声明' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="说明" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.meta?.summary || row.handler || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <ElButton text type="primary" @click="handleUseUnregisteredRoute(row)">
              创建 API
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="unregistered-footer">
        <div class="unregistered-footer-text">共 {{ unregisteredPagination.total }} 条未注册路由</div>
        <ElPagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="unregisteredPagination.current"
          :page-size="unregisteredPagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="unregisteredPagination.total"
          @current-change="handleUnregisteredCurrentChange"
          @size-change="handleUnregisteredSizeChange"
        />
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchCreateApiEndpoint,
    fetchCreateApiEndpointCategory,
    fetchGetApiEndpointCategories,
    fetchGetApiEndpointList,
    fetchGetUnregisteredApiRouteList,
    fetchSyncApiEndpoints,
    fetchUpdateApiEndpoint,
    fetchUpdateApiEndpointCategory,
    fetchUpdateApiEndpointContextScope
  } from '@/api/system-manage'
  import {
    ElButton,
    ElCheckbox,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElSelect,
    ElTag
  } from 'element-plus'

  defineOptions({ name: 'ApiEndpoint' })

  type APIEndpointItem = Api.SystemManage.APIEndpointItem
  type APIEndpointCategoryItem = Api.SystemManage.APIEndpointCategoryItem
  type APIUnregisteredRouteItem = Api.SystemManage.APIUnregisteredRouteItem
  type SummaryItem = { label: string; count: number }
  type PersistedTableState = {
    selectedModule: string
    selectedSource: string
    tableQuery: {
      method: string
      path: string
      keyword: string
      permissionKey: string
      categoryId: string
      contextScope: string
      featureKind: string
      status: string
      hasPermissionKey: string
      hasCategory: string
    }
  }

  const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
  const API_ENDPOINT_TABLE_STATE_KEY = 'system:api-endpoint:table-state'
  const syncing = ref(false)
  const saving = ref(false)
  const categorySaving = ref(false)
  const categorySwitchingId = ref('')
  const selectedModule = ref('')
  const selectedSource = ref('')
  const formVisible = ref(false)
  const categoryDialogVisible = ref(false)
  const unregisteredVisible = ref(false)
  const unregisteredLoading = ref(false)
  const shouldRefreshUnregistered = ref(false)
  const editingId = ref('')
  const pendingLocateRoute = ref<{ method: string; path: string; source: string } | null>(null)
  const categories = ref<APIEndpointCategoryItem[]>([])
  const unregisteredRoutes = ref<APIUnregisteredRouteItem[]>([])
  const totalCount = ref(0)
  const unregisteredCount = ref(0)
  const moduleSummary = ref<SummaryItem[]>([])
  const registrySummary = reactive({
    sync: 0,
    manual: 0,
    seed: 0
  })
  const unregisteredPagination = reactive({
    current: 1,
    size: 10,
    total: 0
  })

  const formState = reactive({
    method: 'GET',
    path: '',
    module: '',
    summary: '',
    featureKind: 'system',
    categoryId: '',
    contextScope: 'optional',
    source: 'manual',
    status: 'normal',
    permissionKeys: [] as string[]
  })

  const categoryForm = reactive({
    id: '',
    code: '',
    name: '',
    nameEn: '',
    sortOrder: 0,
    status: 'normal'
  })

  const unregisteredQuery = reactive({
    method: '',
    path: '',
    module: '',
    keyword: '',
    onlyNoMeta: false
  })

  const tableQuery = reactive({
    method: '',
    path: '',
    keyword: '',
    permissionKey: '',
    categoryId: '',
    contextScope: '',
    featureKind: '',
    status: '',
    hasPermissionKey: '',
    hasCategory: ''
  })

  const sortedCategories = computed(() =>
    [...categories.value].sort(
      (a, b) =>
        (a.sortOrder ?? 0) - (b.sortOrder ?? 0) ||
        `${a.name || ''}`.localeCompare(`${b.name || ''}`, 'zh-CN')
    )
  )

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
      apiFn: fetchGetApiEndpointList,
      apiParams: {
        current: 1,
        size: 20,
        module: '',
        source: ''
      },
      columnsFactory: () => [
        {
          prop: 'method',
          label: 'Method',
          width: 92,
          fixed: 'left',
          formatter: (row: APIEndpointItem) =>
            h(
              ElTag,
              {
                type: methodTagType(row.method),
                effect: 'dark'
              },
              () => row.method
            )
        },
        {
          prop: 'path',
          label: '路径',
          minWidth: 300,
          showOverflowTooltip: true,
          formatter: (row: APIEndpointItem) =>
            h('div', { class: 'path-cell' }, [
              h('div', { class: 'path-main' }, row.path),
              h('div', { class: 'path-sub' }, row.summary || '-'),
              h('div', { class: 'path-tags' }, [
                h(
                  ElTag,
                  {
                    size: 'small',
                    type: sourceTagType(row.source),
                    effect: 'plain'
                  },
                  () => formatSource(row.source)
                ),
                h(
                  ElTag,
                  {
                    size: 'small',
                    type: row.featureKind === 'business' ? 'success' : 'info',
                    effect: 'plain'
                  },
                  () => (row.featureKind === 'business' ? '业务 API' : '系统 API')
                )
              ])
            ])
        },
        { prop: 'module', label: '模块', width: 120 },
        {
          prop: 'category',
          label: '分类',
          minWidth: 180,
          formatter: (row: APIEndpointItem) =>
            row.category ? `${row.category.name} / ${row.category.nameEn}` : '-'
        },
        {
          prop: 'permissionKey',
          label: '权限键',
          minWidth: 240,
          formatter: (row: APIEndpointItem) =>
            (row.permissionKeys || []).join(', ') || row.permissionKey || '-'
        },
        {
          prop: 'contextScope',
          label: '团队上下文',
          width: 140,
          formatter: (row: APIEndpointItem) =>
            h(
              ElSelect,
              {
                modelValue: row.contextScope || 'optional',
                size: 'small',
                onChange: (value: string) => handleContextScopeChange(row, value)
              },
              () => [
                h(ElOption, { label: '可选', value: 'optional' }),
                h(ElOption, { label: '必需', value: 'required' }),
                h(ElOption, { label: '禁止', value: 'forbidden' })
              ]
            )
        },
        {
          prop: 'source',
          label: '来源',
          width: 100,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: sourceTagType(row.source), effect: 'plain' }, () =>
              formatSource(row.source)
            )
        },
        {
          prop: 'featureKind',
          label: '功能归属',
          width: 100,
          formatter: (row: APIEndpointItem) =>
            h(
              ElTag,
              { type: row.featureKind === 'business' ? 'success' : 'info', effect: 'plain' },
              () => (row.featureKind === 'business' ? '业务' : '系统')
            )
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: APIEndpointItem) =>
            h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operate',
          label: '操作',
          width: 90,
          fixed: 'right',
          formatter: (row: APIEndpointItem) =>
            h(
              ElButton,
              {
                text: true,
                type: 'primary',
                onClick: () => openEditDialog(row)
              },
              () => '编辑'
            )
        }
      ]
    }
  })

  function methodTagType(method?: string) {
    switch (`${method || ''}`.toUpperCase()) {
      case 'POST':
        return 'success'
      case 'PUT':
        return 'warning'
      case 'DELETE':
        return 'danger'
      default:
        return 'info'
    }
  }

  function sourceTagType(source?: string) {
    switch (source) {
      case 'manual':
        return 'warning'
      case 'seed':
        return 'success'
      default:
        return 'info'
    }
  }

  function formatSource(source?: string) {
    switch (source) {
      case 'manual':
        return '手工维护'
      case 'seed':
        return '初始种子'
      default:
        return '自动同步'
    }
  }

  function saveTableState() {
    const payload: PersistedTableState = {
      selectedModule: selectedModule.value,
      selectedSource: selectedSource.value,
      tableQuery: {
        method: tableQuery.method,
        path: tableQuery.path,
        keyword: tableQuery.keyword,
        permissionKey: tableQuery.permissionKey,
        categoryId: tableQuery.categoryId,
        contextScope: tableQuery.contextScope,
        featureKind: tableQuery.featureKind,
        status: tableQuery.status,
        hasPermissionKey: tableQuery.hasPermissionKey,
        hasCategory: tableQuery.hasCategory
      }
    }
    localStorage.setItem(API_ENDPOINT_TABLE_STATE_KEY, JSON.stringify(payload))
  }

  function restoreTableState() {
    const raw = localStorage.getItem(API_ENDPOINT_TABLE_STATE_KEY)
    if (!raw) {
      return
    }
    try {
      const payload = JSON.parse(raw) as Partial<PersistedTableState>
      selectedModule.value = payload.selectedModule || ''
      selectedSource.value = payload.selectedSource || ''
      Object.assign(tableQuery, {
        method: payload.tableQuery?.method || '',
        path: payload.tableQuery?.path || '',
        keyword: payload.tableQuery?.keyword || '',
        permissionKey: payload.tableQuery?.permissionKey || '',
        categoryId: payload.tableQuery?.categoryId || '',
        contextScope: payload.tableQuery?.contextScope || '',
        featureKind: payload.tableQuery?.featureKind || '',
        status: payload.tableQuery?.status || '',
        hasPermissionKey: payload.tableQuery?.hasPermissionKey || '',
        hasCategory: payload.tableQuery?.hasCategory || ''
      })
    } catch {
      localStorage.removeItem(API_ENDPOINT_TABLE_STATE_KEY)
    }
  }

  async function loadCategories() {
    const res = await fetchGetApiEndpointCategories()
    categories.value = [...(res.records || [])]
  }

  async function handleSync() {
    syncing.value = true
    try {
      await fetchSyncApiEndpoints()
      ElMessage.success('同步成功')
      await refreshData()
      await loadModuleSummary()
      await loadUnregisteredCount()
    } catch (error: any) {
      ElMessage.error(error?.message || '同步失败')
    } finally {
      syncing.value = false
    }
  }

  async function handleContextScopeChange(row: APIEndpointItem, value: string) {
    try {
      await fetchUpdateApiEndpointContextScope(row.id, value)
      row.contextScope = value
      ElMessage.success('团队上下文已更新')
    } catch (error: any) {
      ElMessage.error(error?.message || '更新失败')
    }
  }

  function resetForm() {
    editingId.value = ''
    pendingLocateRoute.value = null
    formState.method = 'GET'
    formState.path = ''
    formState.module = ''
    formState.summary = ''
    formState.featureKind = 'system'
    formState.categoryId = ''
    formState.contextScope = 'optional'
    formState.source = 'manual'
    formState.status = 'normal'
    formState.permissionKeys = []
  }

  function openCreateDialog() {
    resetForm()
    formVisible.value = true
  }

  function resolveCategoryIdByCode(code?: string) {
    const target = `${code || ''}`.trim().toLowerCase()
    if (!target) {
      return ''
    }
    return categories.value.find((item) => `${item.code || ''}`.trim().toLowerCase() === target)?.id || ''
  }

  function openEditDialog(row: APIEndpointItem) {
    editingId.value = row.id
    pendingLocateRoute.value = null
    formState.method = (row.method || 'GET').toUpperCase()
    formState.path = row.path || ''
    formState.module = row.module || ''
    formState.summary = row.summary || ''
    formState.featureKind = row.featureKind || 'system'
    formState.categoryId = row.categoryId || ''
    formState.contextScope = row.contextScope || 'optional'
    formState.source = row.source || 'manual'
    formState.status = row.status || 'normal'
    formState.permissionKeys = [
      ...(row.permissionKeys || (row.permissionKey ? [row.permissionKey] : []))
    ]
    formVisible.value = true
  }

  function openCategoryDialog(category?: APIEndpointCategoryItem) {
    categoryForm.id = category?.id || ''
    categoryForm.code = category?.code || ''
    categoryForm.name = category?.name || ''
    categoryForm.nameEn = category?.nameEn || ''
    categoryForm.sortOrder = category?.sortOrder ?? 0
    categoryForm.status = category?.status || 'normal'
    categoryDialogVisible.value = true
  }

  async function loadUnregisteredRoutes() {
    unregisteredLoading.value = true
    try {
      const res = await fetchGetUnregisteredApiRouteList({
        current: unregisteredPagination.current,
        size: unregisteredPagination.size,
        method: unregisteredQuery.method || undefined,
        path: unregisteredQuery.path || undefined,
        module: unregisteredQuery.module || undefined,
        keyword: unregisteredQuery.keyword || undefined,
        only_no_meta: unregisteredQuery.onlyNoMeta || undefined
      })
      unregisteredRoutes.value = res.records || []
      unregisteredPagination.total = res.total || 0
      unregisteredCount.value = res.total || 0
    } catch (error: any) {
      ElMessage.error(error?.message || '获取未注册 API 失败')
    } finally {
      unregisteredLoading.value = false
    }
  }

  async function loadUnregisteredCount() {
    try {
      const res = await fetchGetUnregisteredApiRouteList({
        current: 1,
        size: 1
      })
      unregisteredCount.value = res.total || 0
    } catch (error: any) {
      ElMessage.error(error?.message || '获取未注册路由统计失败')
    }
  }

  async function openUnregisteredDialog() {
    unregisteredVisible.value = true
    shouldRefreshUnregistered.value = true
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  async function handleUnregisteredSearch() {
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  async function handleUnregisteredCurrentChange(page: number) {
    unregisteredPagination.current = page
    await loadUnregisteredRoutes()
  }

  async function handleUnregisteredSizeChange(size: number) {
    unregisteredPagination.size = size
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  async function resetUnregisteredQuery() {
    unregisteredQuery.method = ''
    unregisteredQuery.path = ''
    unregisteredQuery.module = ''
    unregisteredQuery.keyword = ''
    unregisteredQuery.onlyNoMeta = false
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  function handleUseUnregisteredRoute(route: APIUnregisteredRouteItem) {
    resetForm()
    formState.method = (route.method || 'GET').toUpperCase()
    formState.path = route.path || ''
    formState.module = route.meta?.module || route.module || ''
    formState.summary = route.meta?.summary || ''
    formState.featureKind = route.meta?.feature_kind || 'system'
    formState.categoryId = resolveCategoryIdByCode(route.meta?.category_code)
    formState.contextScope = route.meta?.context_scope || 'optional'
    formState.source = route.meta?.source || (route.hasMeta ? 'sync' : 'manual')
    formState.permissionKeys = [...(route.meta?.permission_keys || [])]
    pendingLocateRoute.value = {
      method: formState.method,
      path: formState.path,
      source: formState.source
    }
    shouldRefreshUnregistered.value = true
    unregisteredVisible.value = false
    formVisible.value = true
    ElMessage.success('已带入新增 API 表单')
  }

  async function submitCategory() {
    categorySaving.value = true
    try {
      const payload = {
        code: categoryForm.code,
        name: categoryForm.name,
        name_en: categoryForm.nameEn,
        sort_order: categoryForm.sortOrder,
        status: categoryForm.status || 'normal'
      }
      if (categoryForm.id) {
        await fetchUpdateApiEndpointCategory(categoryForm.id, payload)
      } else {
        await fetchCreateApiEndpointCategory(payload)
      }
      await Promise.all([loadCategories(), refreshData()])
      categoryDialogVisible.value = false
      ElMessage.success('分类保存成功')
    } catch (error: any) {
      ElMessage.error(error?.message || '分类保存失败')
    } finally {
      categorySaving.value = false
    }
  }

  async function toggleCategoryStatus(category: APIEndpointCategoryItem) {
    const nextStatus = category.status === 'suspended' ? 'normal' : 'suspended'
    const actionText = nextStatus === 'normal' ? '启用' : '停用'
    try {
      await ElMessageBox.confirm(
        `${actionText}后不会删除已有 API 归属，但后续分配会按新状态执行。`,
        `确认${actionText}分类`,
        {
          type: 'warning'
        }
      )
    } catch {
      return
    }

    categorySwitchingId.value = category.id
    try {
      const payload = {
        code: category.code,
        name: category.name,
        name_en: category.nameEn,
        sort_order: category.sortOrder ?? 0,
        status: nextStatus
      }
      await fetchUpdateApiEndpointCategory(category.id, payload)
      await Promise.all([loadCategories(), refreshData()])
      ElMessage.success(`分类已${actionText}`)
    } catch (error: any) {
      ElMessage.error(error?.message || `分类${actionText}失败`)
    } finally {
      categorySwitchingId.value = ''
    }
  }

  async function submitForm() {
    const isEditing = !!editingId.value
    const payload = {
      method: formState.method,
      path: formState.path,
      module: formState.module,
      summary: formState.summary,
      feature_kind: formState.featureKind,
      category_id: formState.categoryId || undefined,
      context_scope: formState.contextScope,
      source: formState.source,
      status: formState.status,
      permission_keys: formState.permissionKeys
    }
    saving.value = true
    try {
      if (editingId.value) {
        await fetchUpdateApiEndpoint(editingId.value, payload)
      } else {
        await fetchCreateApiEndpoint(payload)
        pendingLocateRoute.value = {
          method: payload.method,
          path: payload.path,
          source: payload.source
        }
      }
      ElMessage.success('保存成功')
      formVisible.value = false
      await refreshData()
      await loadModuleSummary()
      await loadUnregisteredCount()
      if (shouldRefreshUnregistered.value) {
        await loadUnregisteredRoutes()
      }
      if (!isEditing && pendingLocateRoute.value) {
        selectedSource.value = pendingLocateRoute.value.source || ''
        tableQuery.method = pendingLocateRoute.value.method || ''
        tableQuery.path = pendingLocateRoute.value.path || ''
        await applyTableFilters()
        pendingLocateRoute.value = null
      }
    } catch (error: any) {
      ElMessage.error(error?.message || '保存失败')
    } finally {
      if (isEditing) {
        pendingLocateRoute.value = null
      }
      saving.value = false
    }
  }

  async function loadModuleSummary() {
    const res = await fetchGetApiEndpointList({
      current: 1,
      size: 1000
    })
    const records = res.records || []
    totalCount.value = res.total || records.length
    const counter = new Map<string, number>()
    records.forEach((item) => {
      const key = (item.module || 'unknown').trim() || 'unknown'
      counter.set(key, (counter.get(key) || 0) + 1)
    })
    registrySummary.sync = 0
    registrySummary.manual = 0
    registrySummary.seed = 0
    records.forEach((item) => {
      switch (item.source) {
        case 'manual':
          registrySummary.manual += 1
          break
        case 'seed':
          registrySummary.seed += 1
          break
        default:
          registrySummary.sync += 1
          break
      }
    })
    moduleSummary.value = [...counter.entries()]
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0], 'zh-CN'))
      .map(([label, count]) => ({ label, count }))
  }

  async function handleModuleSelect(module: string) {
    selectedModule.value = module
    await applyTableFilters()
  }

  async function handleSourceSelect(source: string) {
    selectedSource.value = source
    await applyTableFilters()
  }

  async function applyTableFilters() {
    Object.assign(searchParams, {
      module: selectedModule.value || undefined,
      source: selectedSource.value || undefined,
      method: tableQuery.method || undefined,
      path: tableQuery.path || undefined,
      keyword: tableQuery.keyword || undefined,
      permissionKey: tableQuery.permissionKey || undefined,
      categoryId: tableQuery.categoryId || undefined,
      contextScope: tableQuery.contextScope || undefined,
      featureKind: tableQuery.featureKind || undefined,
      status: tableQuery.status || undefined,
      hasPermissionKey:
        tableQuery.hasPermissionKey === '' ? undefined : tableQuery.hasPermissionKey === 'true',
      hasCategory: tableQuery.hasCategory === '' ? undefined : tableQuery.hasCategory === 'true',
      current: 1
    })
    saveTableState()
    await getData()
  }

  async function handleTableSearch() {
    await applyTableFilters()
  }

  async function resetTableQuery() {
    selectedModule.value = ''
    selectedSource.value = ''
    tableQuery.method = ''
    tableQuery.path = ''
    tableQuery.keyword = ''
    tableQuery.permissionKey = ''
    tableQuery.categoryId = ''
    tableQuery.contextScope = ''
    tableQuery.featureKind = ''
    tableQuery.status = ''
    tableQuery.hasPermissionKey = ''
    tableQuery.hasCategory = ''
    await applyTableFilters()
  }

  onMounted(async () => {
    restoreTableState()
    await Promise.all([loadCategories(), loadModuleSummary(), loadUnregisteredCount()])
    if (selectedModule.value || selectedSource.value || Object.values(tableQuery).some((item) => item)) {
      await applyTableFilters()
    }
  })
</script>

<style scoped>
  .module-card {
    margin-bottom: 12px;
  }

  .module-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }

  .module-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .module-help {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .module-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .status-overview {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
    margin-top: 16px;
  }

  .status-card {
    padding: 14px 16px;
    border-radius: 14px;
    border: 1px solid var(--el-border-color-lighter);
    background: linear-gradient(135deg, var(--el-fill-color-extra-light), #fff);
  }

  .status-card.is-sync {
    border-color: rgba(64, 158, 255, 0.24);
  }

  .status-card.is-manual {
    border-color: rgba(230, 162, 60, 0.28);
  }

  .status-card.is-seed {
    border-color: rgba(103, 194, 58, 0.28);
  }

  .status-card.is-pending {
    border-color: rgba(144, 147, 153, 0.28);
  }

  .status-label {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .status-value {
    margin-top: 8px;
    font-size: 28px;
    font-weight: 700;
    line-height: 1;
    color: var(--el-text-color-primary);
  }

  .status-help {
    margin-top: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .module-tag {
    cursor: pointer;
  }

  .category-panel {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  .source-filter-panel {
    margin-bottom: 12px;
    padding: 12px 0 4px;
  }

  .table-query-panel {
    display: grid;
    grid-template-columns: 120px repeat(4, minmax(0, 1fr)) repeat(5, minmax(140px, 1fr)) auto;
    gap: 12px;
    align-items: center;
    margin-bottom: 12px;
  }

  .table-query-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .source-filter-title {
    margin-bottom: 10px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .source-filter-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .category-panel-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }

  .category-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 12px;
  }

  .category-card {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 12px;
    border: 1px solid var(--el-border-color);
    border-radius: 12px;
    background: linear-gradient(135deg, var(--el-fill-color-extra-light), #fff);
  }

  .category-card.is-suspended {
    background: linear-gradient(135deg, var(--el-fill-color-light), #fff);
    opacity: 0.8;
  }

  .category-card-main {
    min-width: 0;
    flex: 1;
  }

  .category-card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .category-card-meta {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    word-break: break-all;
  }

  .category-card-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }

  .category-empty {
    padding: 12px 0 4px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .category-input-wrap {
    display: flex;
    width: 100%;
    gap: 8px;
  }

  .category-input-wrap :deep(.el-select) {
    flex: 1;
  }

  .path-cell {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .path-main {
    font-family: Consolas, 'Courier New', monospace;
    color: var(--el-text-color-primary);
  }

  .path-sub {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .path-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .unregistered-toolbar {
    display: grid;
    grid-template-columns: 120px repeat(3, minmax(0, 1fr)) auto auto;
    gap: 12px;
    align-items: center;
    margin-bottom: 16px;
  }

  .unregistered-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 16px;
  }

  .unregistered-footer-text {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .toolbar-count {
    margin-left: 4px;
  }

  @media (max-width: 1280px) {
    .status-overview {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .table-query-panel {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .unregistered-toolbar {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .unregistered-footer {
      flex-direction: column;
      align-items: flex-start;
    }
  }

  @media (max-width: 768px) {
    .status-overview {
      grid-template-columns: 1fr;
    }

    .table-query-panel {
      grid-template-columns: 1fr;
    }
  }
</style>
