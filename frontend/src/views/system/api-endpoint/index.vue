<template>
  <div class="art-full-height">
    <div class="box-border flex gap-4 h-full max-md:block max-md:gap-0 max-md:h-auto">
      <div class="flex-shrink-0 w-58 h-full max-md:w-full max-md:h-auto max-md:mb-5">
        <ElCard class="tree-card art-card-xs flex flex-col h-full mt-0" shadow="never">
          <template #header>
            <b>分类树</b>
          </template>
          <ElScrollbar>
            <ElTree
              class="category-tree"
              :data="categoryTreeData"
              node-key="id"
              default-expand-all
              highlight-current
              :expand-on-click-node="false"
              :current-node-key="selectedCategoryTreeKey"
              @node-click="handleCategoryTreeSelect"
            >
              <template #default="{ data: node }">
                <div class="category-tree-node">
                  <div class="category-tree-node-main">
                    <span
                      v-if="node.type === 'category' && node.status === 'suspended'"
                      class="category-tree-node-status"
                    >
                      <ElTag size="small" type="info" effect="plain"> 停用 </ElTag>
                    </span>
                    <span class="category-tree-node-label">{{ node.label }}</span>
                  </div>
                  <div class="category-tree-node-side">
                    <span class="category-tree-node-count">{{ node.count }}</span>
                    <ElButton
                      v-if="node.type === 'category'"
                      class="category-tree-node-edit"
                      text
                      type="primary"
                      @click.stop="openCategoryDrawer(node.category)"
                    >
                      编辑
                    </ElButton>
                  </div>
                </div>
              </template>
            </ElTree>
          </ElScrollbar>
        </ElCard>
      </div>

      <div class="flex flex-col flex-grow min-w-0">
        <div class="page-top-stack">
          <ApiEndpointSearch
            v-show="showSearchBar"
            v-model="searchForm"
            @search="handleTableSearch"
            @reset="resetTableQuery"
          />

          <ElCard
            class="flex flex-col flex-1 min-h-0 art-table-card api-table-card"
            shadow="never"
          >
            <AdminWorkspaceHero
              title="API 管理"
              description="维护接口注册、分类、权限键与运行时状态，未注册和失效接口也在同一页诊断。"
              :metrics="summaryMetrics"
            >
              <div class="api-hero-actions">
                <div class="api-hero-actions__group">
                  <ElButton
                    v-action="'system.api_registry.sync'"
                    type="primary"
                    @click="openCreateDialog"
                    v-ripple
                  >
                    新增 API
                  </ElButton>
                  <ElButton
                    v-action="'system.api_registry.sync'"
                    plain
                    :loading="syncing"
                    @click="handleSync"
                    v-ripple
                  >
                    全局同步 API
                  </ElButton>
                </div>

                <div class="api-hero-actions__group api-hero-actions__group--secondary">
                  <ElButton
                    v-action="'system.api_registry.sync'"
                    plain
                    @click="openScanConfigDialog"
                    v-ripple
                  >
                    扫描配置
                  </ElButton>
                  <ElButton
                    v-action="'system.api_registry.view'"
                    plain
                    @click="openUnregisteredDialog"
                    v-ripple
                  >
                    全局未注册 API
                    <span v-if="unregisteredCount > 0" class="toolbar-count">
                      ({{ unregisteredCount }})
                    </span>
                  </ElButton>
                  <ElButton
                    v-action="'system.api_registry.sync'"
                    plain
                    type="danger"
                    :loading="cleaningStale"
                    @click="handleCleanupStale"
                    v-ripple
                  >
                    全局清理失效 API
                    <span v-if="staleCount > 0" class="toolbar-count">({{ staleCount }})</span>
                  </ElButton>
                </div>
              </div>
            </AdminWorkspaceHero>

            <ElAlert
              v-if="loadError"
              class="api-inline-alert"
              type="info"
              :closable="false"
              show-icon
              :title="loadError"
            />

            <ArtTableHeader
              v-model:columns="columnChecks"
              v-model:showSearchBar="showSearchBar"
              :loading="loading"
              @refresh="refreshData"
            />

            <div class="api-table-main">
              <ArtTable
                :loading="loading"
                :data="data"
                :columns="columns"
                :pagination="pagination"
                size="small"
                @pagination:size-change="handleSizeChange"
                @pagination:current-change="handleCurrentChange"
              />
            </div>
          </ElCard>
        </div>
      </div>
    </div>

    <ElDrawer
      v-model="formVisible"
      :title="editingId ? '编辑 API' : '新增 API'"
      size="760px"
      direction="rtl"
      destroy-on-close
      class="config-drawer"
    >
      <ElForm :model="formState" label-width="110px" class="api-form">
        <div class="form-intro">
          <div class="form-intro__title">{{ editingId ? '调整接口元数据' : '新增接口注册项' }}</div>
          <div class="form-intro__text">
            先明确接口身份，再配置分类、团队上下文和权限键。这里保存的是正式注册信息，不是临时调试项。
          </div>
        </div>

        <div class="form-section">
          <div class="form-section__header">
            <div class="form-section__title">接口身份</div>
            <div class="form-section__desc">Method、路径、来源和功能归属决定这条接口如何进入正式注册表。</div>
          </div>
          <ElRow :gutter="12">
            <ElCol :span="8">
              <ElFormItem label="Method" prop="method">
                <ElSelect
                  v-model="formState.method"
                  placeholder="请选择"
                  popper-class="api-endpoint-select-popper"
                >
                  <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                </ElSelect>
              </ElFormItem>
            </ElCol>
            <ElCol :span="8">
              <ElFormItem label="功能归属" prop="featureKind">
                <ElSelect
                  v-model="formState.featureKind"
                  placeholder="请选择"
                  popper-class="api-endpoint-select-popper"
                >
                  <ElOption label="系统" value="system" />
                  <ElOption label="业务" value="business" />
                </ElSelect>
              </ElFormItem>
            </ElCol>
            <ElCol :span="8">
              <ElFormItem label="来源" prop="source">
                <ElSelect
                  v-model="formState.source"
                  placeholder="请选择"
                  popper-class="api-endpoint-select-popper"
                >
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
          <ElFormItem label="说明" prop="summary">
            <ElInput v-model="formState.summary" placeholder="说明这条接口的用途和影响范围" />
          </ElFormItem>
        </div>

        <div class="form-section">
          <div class="form-section__header">
            <div class="form-section__title">归属与运行时</div>
            <div class="form-section__desc">分类影响管理归档，团队上下文和状态影响运行时访问与诊断结果。</div>
          </div>
          <ElRow :gutter="12">
            <ElCol :span="12">
              <ElFormItem label="应用范围" prop="appScope">
                <ElSelect
                  v-model="formState.appScope"
                  placeholder="请选择"
                  popper-class="api-endpoint-select-popper"
                >
                  <ElOption label="共享接口" value="shared" />
                  <ElOption label="当前 App" value="app" />
                </ElSelect>
              </ElFormItem>
            </ElCol>
            <ElCol :span="12">
              <ElFormItem label="App 归属">
                <ElInput
                  :model-value="targetAppKey"
                  :disabled="formState.appScope !== 'app'"
                  placeholder="共享接口无需指定 App"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>
          <ElFormItem label="分类" prop="categoryId">
            <div class="category-input-wrap">
              <ElSelect
                v-model="formState.categoryId"
                clearable
                filterable
                placeholder="请选择分类"
                popper-class="api-endpoint-select-popper"
              >
                <ElOption
                  v-for="item in sortedCategories"
                  :key="item.id"
                  :label="`${item.name} / ${item.nameEn}${item.status === 'suspended' ? '（已停用）' : ''}`"
                  :value="item.id"
                  :disabled="item.status === 'suspended' && formState.categoryId !== item.id"
                />
              </ElSelect>
              <ElButton text type="primary" @click="openCategoryDrawer()">分类管理</ElButton>
            </div>
          </ElFormItem>

          <ElRow :gutter="12">
            <ElCol :span="12">
              <ElFormItem label="团队上下文" prop="contextScope">
                <ElSelect
                  v-model="formState.contextScope"
                  placeholder="请选择"
                  popper-class="api-endpoint-select-popper"
                >
                  <ElOption label="可选" value="optional" />
                  <ElOption label="必需" value="required" />
                  <ElOption label="禁止" value="forbidden" />
                </ElSelect>
              </ElFormItem>
            </ElCol>
            <ElCol :span="12">
              <ElFormItem prop="status">
                <template #label>
                  <span class="label-help">
                    <span>状态</span>
                    <ElTooltip content="停用后该 API 将被运行时拒绝访问。" placement="top">
                      <ElIcon class="label-help-icon"><QuestionFilled /></ElIcon>
                    </ElTooltip>
                  </span>
                </template>
                <ElSelect
                  v-model="formState.status"
                  placeholder="请选择"
                  popper-class="api-endpoint-select-popper"
                >
                  <ElOption label="正常" value="normal" />
                  <ElOption label="停用" value="suspended" />
                </ElSelect>
              </ElFormItem>
            </ElCol>
          </ElRow>
        </div>

        <div class="form-section">
          <div class="form-section__header">
            <div class="form-section__title">权限绑定</div>
            <div class="form-section__desc">权限键决定这条接口会被哪条能力链消费，没有权限键时会更接近基础接口。</div>
          </div>
          <ElFormItem label="权限键">
            <ElSelect
              v-model="formState.permissionKeys"
              multiple
              filterable
              allow-create
              default-first-option
              popper-class="api-endpoint-select-popper"
            >
              <ElOption
                v-for="item in formState.permissionKeys"
                :key="item"
                :label="item"
                :value="item"
              />
            </ElSelect>
          </ElFormItem>
        </div>
      </ElForm>
      <template #footer>
        <ElButton @click="formVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="saving" @click="submitForm">保存</ElButton>
      </template>
    </ElDrawer>

    <ElDrawer
      v-model="categoryDrawerVisible"
      :title="categoryForm.id ? '编辑分类' : '分类管理'"
      size="860px"
      destroy-on-close
      class="config-drawer"
    >
      <div class="category-drawer">
        <div class="category-drawer-toolbar">
          <div>
            <div class="module-title">分类配置</div>
            <div class="module-help">分类停用后会保留历史归属，但不建议继续分配。</div>
          </div>
          <ElButton type="primary" @click="startCreateCategory">新建分类</ElButton>
        </div>

        <div class="category-drawer-content">
          <div class="category-drawer-list">
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
                  <ElButton text type="primary" @click="openCategoryDrawer(item)">编辑</ElButton>
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

          <div class="category-form-panel">
            <div class="category-form-header">
              <div class="module-title">{{ categoryForm.id ? '编辑分类' : '新建分类' }}</div>
              <ElButton text @click="resetCategoryForm">清空</ElButton>
            </div>

            <ElForm :model="categoryForm" label-width="88px">
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
              <ElFormItem>
                <template #label>
                  <span class="label-help">
                    <span>状态</span>
                    <ElTooltip content="仅影响分类管理，不影响接口鉴权判断。" placement="top">
                      <ElIcon class="label-help-icon"><QuestionFilled /></ElIcon>
                    </ElTooltip>
                  </span>
                </template>
                <ElSelect v-model="categoryForm.status" placeholder="请选择">
                  <ElOption label="正常" value="normal" />
                  <ElOption label="停用" value="suspended" />
                </ElSelect>
              </ElFormItem>
            </ElForm>

            <div class="category-form-actions">
              <ElButton @click="resetCategoryForm">重置</ElButton>
              <ElButton type="primary" :loading="categorySaving" @click="submitCategory">
                保存
              </ElButton>
            </div>
          </div>
        </div>
      </div>
    </ElDrawer>

    <ElDialog v-model="staleDialogVisible" title="清理失效 API" width="980px" destroy-on-close>
      <div class="stale-dialog-tip">
        仅会删除“来源为自动同步、且源码中已不存在”的失效 API，请勾选后执行删除。
      </div>

      <ElTable
        ref="staleTableRef"
        :data="staleCandidates"
        border
        height="420px"
        row-key="id"
        @selection-change="handleStaleSelectionChange"
      >
        <ElTableColumn type="selection" width="52" reserve-selection />
        <ElTableColumn label="Method" width="92">
          <template #default="{ row }">
            <ElTag :type="methodTagType(row.method)" effect="dark">{{ row.method }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="path" label="路径" min-width="300" show-overflow-tooltip />
        <ElTableColumn label="分类" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.category?.name || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="110">
          <template #default="{ row }">
            <ElTag type="warning" effect="plain">{{
              row.status === 'suspended' ? '停用(失效)' : '失效'
            }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="失效原因" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.staleReason || '源码中已不存在该自动同步 API' }}
          </template>
        </ElTableColumn>
      </ElTable>

      <template #footer>
        <div class="stale-dialog-footer">
          <div class="stale-dialog-footer-meta">
            <div class="stale-dialog-footer-text">
              已选 {{ selectedStaleIds.length }} 项，共 {{ stalePagination.total }} 条失效 API
            </div>
            <ElPagination
              background
              layout="total, sizes, prev, pager, next"
              :current-page="stalePagination.current"
              :page-size="stalePagination.size"
              :page-sizes="[20, 50, 100]"
              :total="stalePagination.total"
              @current-change="handleStaleCurrentChange"
              @size-change="handleStaleSizeChange"
            />
          </div>
          <div class="stale-dialog-footer-actions">
            <ElButton @click="closeStaleDialog">取消</ElButton>
            <ElButton type="danger" :loading="cleaningStale" @click="submitCleanupStale">
              删除选中
            </ElButton>
          </div>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="unregisteredVisible" title="未注册 API" width="980px" destroy-on-close>
      <div class="unregistered-toolbar">
        <ElSelect
          class="unregistered-method-select"
          v-model="unregisteredQuery.method"
          clearable
          placeholder="Method"
        >
          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
        </ElSelect>
        <ElInput v-model="unregisteredQuery.path" placeholder="按路径筛选" clearable />
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
        <div class="unregistered-footer-text"
          >共 {{ unregisteredPagination.total }} 条未注册路由</div
        >
        <ElPagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="unregisteredPagination.current"
          :page-size="unregisteredPagination.size"
          :page-sizes="[20, 50, 100]"
          :total="unregisteredPagination.total"
          @current-change="handleUnregisteredCurrentChange"
          @size-change="handleUnregisteredSizeChange"
        />
      </div>
    </ElDialog>

    <ElDialog v-model="scanConfigVisible" title="未注册 API 扫描配置（实验功能）" width="560px" destroy-on-close>
      <ElForm label-width="140px">
        <ElAlert
          type="warning"
          :closable="false"
          show-icon
          title="当前仅保存配置，不代表已启用后台自动调度。"
          style="margin-bottom: 12px"
        />
        <ElFormItem label="启用自动扫描">
          <ElSwitch v-model="scanConfig.enabled" />
        </ElFormItem>
        <ElAlert
          v-if="scanConfig.enabled"
          type="info"
          :closable="false"
          show-icon
          title="开关开启后，仍需后台调度器接入才会真正自动执行。"
          style="margin-bottom: 12px"
        />
        <ElFormItem label="扫描频率（分钟）">
          <ElInputNumber v-model="scanConfig.frequencyMinutes" :min="5" :max="1440" style="width: 100%" />
        </ElFormItem>
        <ElFormItem label="默认分类ID">
          <ElInput v-model="scanConfig.defaultCategoryId" placeholder="可选，自动归类" />
        </ElFormItem>
        <ElFormItem label="默认权限键">
          <ElInput v-model="scanConfig.defaultPermissionKey" placeholder="可选，自动绑定权限键" />
        </ElFormItem>
        <ElFormItem label="标记无权限要求">
          <ElSwitch v-model="scanConfig.markAsNoPermission" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="scanConfigVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="scanConfigSaving" @click="saveScanConfig">保存</ElButton>
      </template>
    </ElDialog>

    <ElDrawer
      v-model="permissionBindVisible"
      :title="permissionDialogMode === 'remove' ? '移除权限键' : '加入权限键'"
      size="560px"
      direction="rtl"
      destroy-on-close
      class="config-drawer"
    >
      <ElForm label-width="92px">
        <ElFormItem label="接口">
          <ElInput :model-value="permissionBinding.endpointSpec" readonly />
        </ElFormItem>
        <ElFormItem label="权限键">
          <ElSelect
            v-model="permissionBinding.permissionActionId"
            filterable
            clearable
            placeholder="请选择权限键"
            :loading="permissionActionLoading"
            popper-class="api-endpoint-select-popper"
          >
            <ElOption
              v-for="item in currentPermissionActionOptions"
              :key="item.id"
              :label="`${item.name || item.permissionKey || '-'}（${item.permissionKey || '-'}）`"
              :value="item.id"
            />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="permissionBindVisible = false">取消</ElButton>
        <ElButton
          :type="permissionDialogMode === 'remove' ? 'danger' : 'primary'"
          @click="submitPermissionBind"
        >
          {{ permissionDialogMode === 'remove' ? '确认移除' : '确认加入' }}
        </ElButton>
      </template>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, nextTick, onMounted, reactive, ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import {
    fetchAddPermissionActionEndpoint,
    fetchCleanupStaleApiEndpoints,
    fetchCreateApiEndpoint,
    fetchCreateApiEndpointCategory,
    fetchDeletePermissionActionEndpoint,
    fetchGetApiEndpointCategories,
    fetchGetApiEndpointList,
    fetchGetApiEndpointOverview,
    fetchGetPermissionActionOptions,
    fetchGetStaleApiEndpointList,
    fetchGetUnregisteredApiScanConfig,
    fetchGetUnregisteredApiRouteList,
    fetchSaveUnregisteredApiScanConfig,
    fetchSyncApiEndpoints,
    fetchUpdateApiEndpoint,
    fetchUpdateApiEndpointCategory,
    fetchUpdateApiEndpointContextScope
  } from '@/api/system-manage'
  import {
    ElButton,
    ElCheckbox,
    ElIcon,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElSelect,
    ElSwitch,
    ElTooltip,
    ElTag
  } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'
  import ApiEndpointSearch from './modules/api-endpoint-search.vue'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'

  defineOptions({ name: 'ApiEndpoint' })

  type APIEndpointItem = Api.SystemManage.APIEndpointItem
  type APIEndpointCategoryItem = Api.SystemManage.APIEndpointCategoryItem
  type APIUnregisteredRouteItem = Api.SystemManage.APIUnregisteredRouteItem
  type CategoryTreeNode = {
    id: string
    label: string
    count: number
    type: 'all' | 'uncategorized' | 'category'
    status?: string
    category?: APIEndpointCategoryItem
    children?: CategoryTreeNode[]
  }
  type PersistedTableState = {
    selectedSource: string
    selectedCategoryTreeKey: string
    tableQuery: {
      method: string
      path: string
      keyword: string
      permissionKey: string
      permissionPattern: string
      contextScope: string
      featureKind: string
      status: string
      hasPermissionKey: string
    }
  }

  const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
  const { hasAction } = useAuth()
  const route = useRoute()
  const { targetAppKey } = useManagedAppScope()
  const managedAppMissingText = '缺少 app 上下文，请先从应用管理选择 App'
  const API_ENDPOINT_TABLE_STATE_KEY = 'system:api-endpoint:table-state'
  const staleTableRef = ref<any>(null)
  const syncing = ref(false)
  const cleaningStale = ref(false)
  const loadError = ref('')
  const showSearchBar = ref(false)
  const saving = ref(false)
  const categorySaving = ref(false)
  const categorySwitchingId = ref('')
  const selectedSource = ref('')
  const selectedCategoryTreeKey = ref('all')
  const formVisible = ref(false)
  const categoryDrawerVisible = ref(false)
  const permissionBindVisible = ref(false)
  const permissionDialogMode = ref<'add' | 'remove'>('add')
  const unregisteredVisible = ref(false)
  const scanConfigVisible = ref(false)
  const scanConfigSaving = ref(false)
  const staleDialogVisible = ref(false)
  const unregisteredLoading = ref(false)
  const shouldRefreshUnregistered = ref(false)
  const editingId = ref('')
  const pendingLocateRoute = ref<{ method: string; path: string; source: string } | null>(null)
  const categories = ref<APIEndpointCategoryItem[]>([])
  const permissionActionOptions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const permissionActionLoading = ref(false)
  const permissionBinding = reactive({
    endpointCode: '',
    endpointSpec: '',
    endpointPermissionKeys: [] as string[],
    permissionActionId: ''
  })
  const unregisteredRoutes = ref<APIUnregisteredRouteItem[]>([])
  const scanConfig = reactive<Api.SystemManage.APIUnregisteredScanConfig>({
    enabled: false,
    frequencyMinutes: 60,
    defaultCategoryId: '',
    defaultPermissionKey: '',
    markAsNoPermission: false
  })
  const staleCandidates = ref<APIEndpointItem[]>([])
  const selectedStaleIds = ref<string[]>([])
  const totalCount = ref(0)
  const noPermissionCount = ref(0)
  const sharedPermissionCount = ref(0)
  const crossContextSharedCount = ref(0)
  const staleCount = ref(0)
  const unregisteredCount = ref(0)
  const uncategorizedCount = ref(0)
  const categoryCountMap = ref<Record<string, number>>({})
  const unregisteredPagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })
  const stalePagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const formState = reactive({
    appScope: 'app',
    method: 'GET',
    path: '',
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
    keyword: '',
    onlyNoMeta: false
  })

  const tableQuery = reactive({
    method: '',
    path: '',
    keyword: '',
    permissionKey: '',
    permissionPattern: '',
    categoryId: '',
    contextScope: '',
    featureKind: '',
    status: '',
    hasPermissionKey: '',
    hasCategory: ''
  })

  const searchForm = reactive({
    source: '',
    method: '',
    path: '',
    keyword: '',
    permissionKey: '',
    permissionPattern: '',
    contextScope: '',
    featureKind: '',
    status: '',
    hasPermissionKey: ''
  })

  function syncSearchFormFromQuery() {
    searchForm.source = selectedSource.value
    searchForm.method = tableQuery.method
    searchForm.path = tableQuery.path
    searchForm.keyword = tableQuery.keyword
    searchForm.permissionKey = tableQuery.permissionKey
    searchForm.permissionPattern = tableQuery.permissionPattern
    searchForm.contextScope = tableQuery.contextScope
    searchForm.featureKind = tableQuery.featureKind
    searchForm.status = tableQuery.status
    searchForm.hasPermissionKey = tableQuery.hasPermissionKey
  }

  function syncQueryFromSearchForm() {
    selectedSource.value = searchForm.source || ''
    tableQuery.method = searchForm.method || ''
    tableQuery.path = searchForm.path || ''
    tableQuery.keyword = searchForm.keyword || ''
    tableQuery.permissionKey = searchForm.permissionKey || ''
    tableQuery.permissionPattern = searchForm.permissionPattern || ''
    tableQuery.contextScope = searchForm.contextScope || ''
    tableQuery.featureKind = searchForm.featureKind || ''
    tableQuery.status = searchForm.status || ''
    tableQuery.hasPermissionKey = searchForm.hasPermissionKey || ''
  }

  const sortedCategories = computed(() =>
    [...categories.value].sort(
      (a, b) =>
        (a.sortOrder ?? 0) - (b.sortOrder ?? 0) ||
        `${a.name || ''}`.localeCompare(`${b.name || ''}`, 'zh-CN')
    )
  )

  const categoryTreeData = computed<CategoryTreeNode[]>(() => [
    {
      id: 'all',
      label: '全部 API',
      count: totalCount.value,
      type: 'all',
      children: [
        {
          id: 'uncategorized',
          label: '未分类',
          count: uncategorizedCount.value,
          type: 'uncategorized'
        },
        ...sortedCategories.value.map((item) => ({
          id: `category:${item.id}`,
          label: item.name || item.code || '未命名分类',
          count: categoryCountMap.value[item.id] || 0,
          type: 'category' as const,
          status: item.status,
          category: item
        }))
      ]
    }
  ])

  const summaryMetrics = computed(() => [
    { label: '管理 App', value: targetAppKey.value || '-' },
    { label: '注册总量', value: totalCount.value || 0 },
    { label: '无权限键', value: noPermissionCount.value || 0 },
    { label: '共享接口', value: sharedPermissionCount.value || 0 },
    { label: '跨上下文共享', value: crossContextSharedCount.value || 0 },
    { label: '未分类', value: uncategorizedCount.value || 0 },
    { label: '失效', value: staleCount.value || 0 },
    { label: '未注册', value: unregisteredCount.value || 0 }
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
      apiFn: fetchGetApiEndpointList,
      apiParams: {
        current: 1,
        size: 20,
        source: '',
        appKey: targetAppKey.value
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
            h('div', { class: 'path-cell' }, [h('div', { class: 'path-main' }, row.path)])
        },
        {
          prop: 'appScope',
          label: '范围',
          width: 90,
          formatter: (row: APIEndpointItem) =>
            h(
              ElTag,
              { type: row.appScope === 'app' ? 'warning' : 'info', effect: 'plain' },
              () => (row.appScope === 'app' ? 'App' : '共享')
            )
        },
        {
          prop: 'appKey',
          label: 'App',
          width: 140,
          formatter: (row: APIEndpointItem) => row.appKey || '-'
        },
        {
          prop: 'summary',
          label: '介绍',
          minWidth: 220,
          showOverflowTooltip: true,
          formatter: (row: APIEndpointItem) => row.summary || '-'
        },
        {
          prop: 'category',
          label: '分类',
          minWidth: 180,
          formatter: (row: APIEndpointItem) => row.category?.name || '-'
        },
        {
          prop: 'permissionKey',
          label: '权限键',
          minWidth: 240,
          formatter: (row: APIEndpointItem) =>
            (row.permissionKeys || []).join(', ') || row.permissionKey || '-'
        },
        {
          prop: 'permissionBindingMode',
          label: '权限结构',
          minWidth: 220,
          formatter: (row: APIEndpointItem) =>
            h('div', { class: 'permission-structure-cell' }, [
              h(
                ElTag,
                {
                  type: permissionPatternTagType(row.permissionBindingMode),
                  effect: 'plain'
                },
                () => formatPermissionPattern(row.permissionBindingMode)
              ),
              row.permissionContexts?.length
                ? h(
                    'div',
                    { class: 'permission-structure-contexts' },
                    row.permissionContexts.map((item) => formatPermissionContext(item)).join(' / ')
                  )
                : null,
              row.permissionNote
                ? h('div', { class: 'permission-structure-note' }, row.permissionNote)
                : null
            ])
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
          width: 180,
          formatter: (row: APIEndpointItem) => {
            if (row.stale) {
              return h('div', { class: 'status-cell' }, [
                h(ElTag, { type: 'warning' }, () => '失效'),
                h(
                  'div',
                  { class: 'status-note' },
                  row.staleReason || '源码中已不存在该自动同步 API'
                )
              ])
            }
            return h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
          }
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operate',
          label: '操作',
          width: 70,
          fixed: 'right',
          formatter: (row: APIEndpointItem) => {
            const list: ButtonMoreItem[] = [
              {
                key: 'edit',
                label: '编辑',
                icon: 'ri:edit-2-line'
              },
              {
                key: 'add',
                label: '加入权限键',
                icon: 'ri:links-line',
                auth: 'system.permission.manage'
              },
              {
                key: 'remove',
                label: '移除权限键',
                icon: 'ri:link-unlink',
                auth: 'system.permission.manage'
              }
            ]
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleOperateCommand(row, item.key as string)
            })
          }
        }
      ]
    }
  })

  function handleOperateCommand(row: APIEndpointItem, command: string) {
    if (command === 'edit') {
      openEditDialog(row)
      return
    }
    if (command === 'add') {
      openPermissionBindDialog(row, 'add')
      return
    }
    if (command === 'remove') {
      openPermissionBindDialog(row, 'remove')
    }
  }

  const currentPermissionActionOptions = computed(() => {
    if (permissionDialogMode.value !== 'remove') {
      return permissionActionOptions.value
    }
    const boundKeys = new Set(
      (permissionBinding.endpointPermissionKeys || []).map((item) => `${item || ''}`.trim())
    )
    return permissionActionOptions.value.filter((item) =>
      boundKeys.has(`${item.permissionKey || ''}`.trim())
    )
  })

  async function loadPermissionActionOptions() {
    permissionActionLoading.value = true
    try {
      const res = await fetchGetPermissionActionOptions()
      permissionActionOptions.value = res.records || []
    } catch (error: any) {
      ElMessage.error(error?.message || '获取权限键失败')
    } finally {
      permissionActionLoading.value = false
    }
  }

  async function openPermissionBindDialog(row: APIEndpointItem, mode: 'add' | 'remove') {
    if (!hasAction('system.permission.manage')) {
      ElMessage.warning('无权限操作')
      return
    }
    if (!row.code) {
      ElMessage.warning('当前 API 缺少固定编码，请先重建 API 注册表')
      return
    }
    permissionDialogMode.value = mode
    permissionBinding.endpointCode = row.code
    permissionBinding.endpointSpec = `${row.method || ''} ${row.path || ''}`.trim()
    permissionBinding.endpointPermissionKeys = [
      ...(row.permissionKeys || (row.permissionKey ? [row.permissionKey] : []))
    ]
    permissionBinding.permissionActionId = ''
    permissionBindVisible.value = true
    await loadPermissionActionOptions()
    if (mode === 'remove' && currentPermissionActionOptions.value.length === 0) {
      ElMessage.info('当前接口没有可移除的权限键')
      permissionBindVisible.value = false
    }
  }

  async function submitPermissionBind() {
    if (!permissionBinding.endpointCode || !permissionBinding.permissionActionId) {
      ElMessage.warning('请选择权限键')
      return
    }
    try {
      if (permissionDialogMode.value === 'remove') {
        await fetchDeletePermissionActionEndpoint(
          permissionBinding.permissionActionId,
          permissionBinding.endpointCode
        )
      } else {
        await fetchAddPermissionActionEndpoint(
          permissionBinding.permissionActionId,
          permissionBinding.endpointCode
        )
      }
      ElMessage.success(permissionDialogMode.value === 'remove' ? '已移除权限键' : '已加入权限键')
      permissionBindVisible.value = false
      await refreshData()
    } catch (error: any) {
      ElMessage.error(
        error?.message ||
          (permissionDialogMode.value === 'remove' ? '移除权限键失败' : '加入权限键失败')
      )
    }
  }

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

  function formatPermissionPattern(value?: string) {
    switch (`${value || ''}`.trim()) {
      case 'public':
        return '公开接口'
      case 'global_jwt':
        return '登录态全局'
      case 'self_jwt':
        return '登录态自服务'
      case 'api_key':
        return '开放 API Key'
      case 'single':
        return '单权限'
      case 'shared':
        return '多权限共享'
      case 'cross_context_shared':
        return '跨上下文共享'
      default:
        return '无权限键'
    }
  }

  function permissionPatternTagType(value?: string) {
    switch (`${value || ''}`.trim()) {
      case 'public':
        return 'success'
      case 'global_jwt':
        return 'info'
      case 'self_jwt':
        return 'warning'
      case 'api_key':
        return 'success'
      case 'single':
        return 'success'
      case 'shared':
        return 'warning'
      case 'cross_context_shared':
        return 'danger'
      default:
        return 'info'
    }
  }

  function formatPermissionContext(value?: string) {
    switch (`${value || ''}`.trim()) {
      case 'platform':
        return '平台'
      case 'team':
        return '团队'
      case 'common':
        return '通用'
      default:
        return value || '-'
    }
  }

  function saveTableState() {
    const payload: PersistedTableState = {
      selectedSource: selectedSource.value,
      selectedCategoryTreeKey: selectedCategoryTreeKey.value,
      tableQuery: {
        method: tableQuery.method,
        path: tableQuery.path,
        keyword: tableQuery.keyword,
        permissionKey: tableQuery.permissionKey,
        permissionPattern: tableQuery.permissionPattern,
        contextScope: tableQuery.contextScope,
        featureKind: tableQuery.featureKind,
        status: tableQuery.status,
        hasPermissionKey: tableQuery.hasPermissionKey
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
      selectedSource.value = payload.selectedSource || ''
      selectedCategoryTreeKey.value = payload.selectedCategoryTreeKey || 'all'
      Object.assign(tableQuery, {
        method: payload.tableQuery?.method || '',
        path: payload.tableQuery?.path || '',
        keyword: payload.tableQuery?.keyword || '',
        permissionKey: payload.tableQuery?.permissionKey || '',
        permissionPattern: payload.tableQuery?.permissionPattern || '',
        contextScope: payload.tableQuery?.contextScope || '',
        featureKind: payload.tableQuery?.featureKind || '',
        status: payload.tableQuery?.status || '',
        hasPermissionKey: payload.tableQuery?.hasPermissionKey || ''
      })
      syncSearchFormFromQuery()
    } catch {
      localStorage.removeItem(API_ENDPOINT_TABLE_STATE_KEY)
    }
  }

  function ensureManagedAppReady(showMessage = false) {
    if (targetAppKey.value) {
      loadError.value = ''
      return true
    }
    loadError.value = managedAppMissingText
    data.value = []
    staleCandidates.value = []
    categories.value = []
    totalCount.value = 0
    noPermissionCount.value = 0
    sharedPermissionCount.value = 0
    crossContextSharedCount.value = 0
    uncategorizedCount.value = 0
    staleCount.value = 0
    unregisteredCount.value = 0
    categoryCountMap.value = {}
    if (showMessage) {
      ElMessage.warning(managedAppMissingText)
    }
    return false
  }

  function resetScopedState(message = managedAppMissingText) {
    loadError.value = message
    data.value = []
    totalCount.value = 0
    noPermissionCount.value = 0
    sharedPermissionCount.value = 0
    crossContextSharedCount.value = 0
    uncategorizedCount.value = 0
    staleCount.value = 0
    categoryCountMap.value = {}
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
      await loadUnregisteredCount()
      if (targetAppKey.value) {
        await Promise.all([refreshData(), loadCategorySummary()])
      } else {
        resetScopedState()
      }
    } catch (error: any) {
      ElMessage.error(error?.message || '同步失败')
    } finally {
      syncing.value = false
    }
  }

  async function handleCleanupStale() {
    stalePagination.current = 1
    selectedStaleIds.value = []
    try {
      await loadStaleCandidates()
    } catch (error: any) {
      ElMessage.error(error?.message || '获取失效 API 列表失败')
      return
    }
    if (!stalePagination.total) {
      ElMessage.info('当前没有可清理的失效 API')
      return
    }
    staleDialogVisible.value = true
  }

  function closeStaleDialog() {
    staleDialogVisible.value = false
    selectedStaleIds.value = []
    staleCandidates.value = []
    staleTableRef.value?.clearSelection?.()
  }

  function handleStaleSelectionChange(rows: APIEndpointItem[]) {
    const currentPageIds = new Set(staleCandidates.value.map((item) => item.id).filter(Boolean))
    const selectedSet = new Set(selectedStaleIds.value)
    currentPageIds.forEach((id) => selectedSet.delete(id))
    rows.forEach((item) => {
      if (item.id) {
        selectedSet.add(item.id)
      }
    })
    selectedStaleIds.value = Array.from(selectedSet)
  }

  async function handleStaleCurrentChange(page: number) {
    stalePagination.current = page
    await loadStaleCandidates()
  }

  async function handleStaleSizeChange(size: number) {
    stalePagination.size = size
    stalePagination.current = 1
    await loadStaleCandidates()
  }

  async function submitCleanupStale() {
    if (selectedStaleIds.value.length === 0) {
      ElMessage.warning('请先勾选要删除的失效 API')
      return
    }
    cleaningStale.value = true
    try {
      const res = await fetchCleanupStaleApiEndpoints(selectedStaleIds.value)
      closeStaleDialog()
      await loadUnregisteredCount()
      if (targetAppKey.value) {
        await Promise.all([refreshData(), loadCategorySummary()])
      } else {
        resetScopedState()
      }
      if (shouldRefreshUnregistered.value) {
        await loadUnregisteredRoutes()
      }
      ElMessage.success(`已清理 ${res.deletedCount || 0} 个失效 API`)
    } catch (error: any) {
      ElMessage.error(error?.message || '清理失效 API 失败')
    } finally {
      cleaningStale.value = false
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
    formState.appScope = 'app'
    formState.method = 'GET'
    formState.path = ''
    formState.summary = ''
    formState.featureKind = 'system'
    formState.categoryId = ''
    formState.contextScope = 'optional'
    formState.source = 'manual'
    formState.status = 'normal'
    formState.permissionKeys = []
  }

  function resetCategoryForm() {
    categoryForm.id = ''
    categoryForm.code = ''
    categoryForm.name = ''
    categoryForm.nameEn = ''
    categoryForm.sortOrder = 0
    categoryForm.status = 'normal'
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
    return (
      categories.value.find((item) => `${item.code || ''}`.trim().toLowerCase() === target)?.id ||
      ''
    )
  }

  function openEditDialog(row: APIEndpointItem) {
    editingId.value = row.id
    pendingLocateRoute.value = null
    formState.appScope = row.appScope === 'shared' ? 'shared' : 'app'
    formState.method = (row.method || 'GET').toUpperCase()
    formState.path = row.path || ''
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

  function startCreateCategory() {
    resetCategoryForm()
  }

  function openCategoryDrawer(category?: APIEndpointCategoryItem) {
    if (category) {
      categoryForm.id = category.id || ''
      categoryForm.code = category.code || ''
      categoryForm.name = category.name || ''
      categoryForm.nameEn = category.nameEn || ''
      categoryForm.sortOrder = category.sortOrder ?? 0
      categoryForm.status = category.status || 'normal'
    } else {
      resetCategoryForm()
    }
    categoryDrawerVisible.value = true
  }

  async function handleCategoryTreeSelect(node: CategoryTreeNode) {
    selectedCategoryTreeKey.value = node.id
    if (node.type === 'uncategorized') {
      tableQuery.categoryId = ''
      tableQuery.hasCategory = 'false'
    } else if (node.type === 'category') {
      tableQuery.categoryId = node.category?.id || ''
      tableQuery.hasCategory = 'true'
    } else {
      tableQuery.categoryId = ''
      tableQuery.hasCategory = ''
    }
    await applyTableFilters()
  }

  function syncCategoryFilterFromTree() {
    if (selectedCategoryTreeKey.value === 'uncategorized') {
      tableQuery.categoryId = ''
      tableQuery.hasCategory = 'false'
      return
    }
    if (selectedCategoryTreeKey.value.startsWith('category:')) {
      tableQuery.categoryId = selectedCategoryTreeKey.value.replace(/^category:/, '')
      tableQuery.hasCategory = 'true'
      return
    }
    tableQuery.categoryId = ''
    tableQuery.hasCategory = ''
  }

  async function loadUnregisteredRoutes() {
    unregisteredLoading.value = true
    try {
      const res = await fetchGetUnregisteredApiRouteList({
        current: unregisteredPagination.current,
        size: unregisteredPagination.size,
        method: unregisteredQuery.method || undefined,
        path: unregisteredQuery.path || undefined,
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

  async function loadStaleCandidates() {
    const res = await fetchGetStaleApiEndpointList({
      current: stalePagination.current,
      size: stalePagination.size
    })
    staleCandidates.value = res.records || []
    stalePagination.total = res.total || 0
    await nextTick()
    syncStaleSelection()
  }

  function syncStaleSelection() {
    const table = staleTableRef.value
    if (!table) {
      return
    }
    const selectedSet = new Set(selectedStaleIds.value)
    table.clearSelection?.()
    staleCandidates.value.forEach((item) => {
      if (selectedSet.has(item.id)) {
        table.toggleRowSelection?.(item, true)
      }
    })
  }

  async function openUnregisteredDialog() {
    unregisteredVisible.value = true
    shouldRefreshUnregistered.value = true
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  async function openScanConfigDialog() {
    scanConfigVisible.value = true
    try {
      const config = await fetchGetUnregisteredApiScanConfig()
      scanConfig.enabled = Boolean(config.enabled)
      scanConfig.frequencyMinutes = Number(config.frequencyMinutes || 60)
      scanConfig.defaultCategoryId = config.defaultCategoryId || ''
      scanConfig.defaultPermissionKey = config.defaultPermissionKey || ''
      scanConfig.markAsNoPermission = Boolean(config.markAsNoPermission)
    } catch (error: any) {
      ElMessage.error(error?.message || '获取扫描配置失败')
    }
  }

  async function saveScanConfig() {
    scanConfigSaving.value = true
    try {
      const saved = await fetchSaveUnregisteredApiScanConfig({
        enabled: scanConfig.enabled,
        frequencyMinutes: scanConfig.frequencyMinutes,
        defaultCategoryId: (scanConfig.defaultCategoryId || '').trim(),
        defaultPermissionKey: (scanConfig.defaultPermissionKey || '').trim(),
        markAsNoPermission: scanConfig.markAsNoPermission
      })
      scanConfig.enabled = Boolean(saved.enabled)
      scanConfig.frequencyMinutes = Number(saved.frequencyMinutes || 60)
      scanConfig.defaultCategoryId = saved.defaultCategoryId || ''
      scanConfig.defaultPermissionKey = saved.defaultPermissionKey || ''
      scanConfig.markAsNoPermission = Boolean(saved.markAsNoPermission)
      scanConfigVisible.value = false
      ElMessage.success('扫描配置已保存')
    } catch (error: any) {
      ElMessage.error(error?.message || '保存扫描配置失败')
    } finally {
      scanConfigSaving.value = false
    }
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
    unregisteredQuery.keyword = ''
    unregisteredQuery.onlyNoMeta = false
    unregisteredPagination.current = 1
    await loadUnregisteredRoutes()
  }

  function handleUseUnregisteredRoute(route: APIUnregisteredRouteItem) {
    resetForm()
    formState.appScope = 'app'
    formState.method = (route.method || 'GET').toUpperCase()
    formState.path = route.path || ''
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
      let savedCategory: APIEndpointCategoryItem
      if (categoryForm.id) {
        savedCategory = await fetchUpdateApiEndpointCategory(categoryForm.id, payload)
      } else {
        savedCategory = await fetchCreateApiEndpointCategory(payload)
        if (formVisible.value) {
          formState.categoryId = savedCategory.id
        }
      }
      await Promise.all([loadCategories(), refreshData(), loadCategorySummary()])
      if (selectedCategoryTreeKey.value.startsWith('category:')) {
        syncCategoryFilterFromTree()
      }
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
      await Promise.all([loadCategories(), refreshData(), loadCategorySummary()])
      ElMessage.success(`分类已${actionText}`)
    } catch (error: any) {
      ElMessage.error(error?.message || `分类${actionText}失败`)
    } finally {
      categorySwitchingId.value = ''
    }
  }

  async function submitForm() {
    if (!ensureManagedAppReady(true)) {
      return
    }
    const isEditing = !!editingId.value
    const payload = {
      app_scope: formState.appScope,
      app_key: formState.appScope === 'app' ? targetAppKey.value : '',
      method: formState.method,
      path: formState.path,
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
      await loadCategorySummary()
      await loadUnregisteredCount()
      if (shouldRefreshUnregistered.value) {
        await loadUnregisteredRoutes()
      }
      if (!isEditing && pendingLocateRoute.value) {
        selectedSource.value = pendingLocateRoute.value.source || ''
        tableQuery.method = pendingLocateRoute.value.method || ''
        tableQuery.path = pendingLocateRoute.value.path || ''
        syncSearchFormFromQuery()
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

  async function loadCategorySummary() {
    if (!ensureManagedAppReady()) {
      return
    }
    const res = await fetchGetApiEndpointOverview(targetAppKey.value)
    totalCount.value = res.totalCount || 0
    noPermissionCount.value = res.noPermissionCount || 0
    sharedPermissionCount.value = res.sharedPermissionCount || 0
    crossContextSharedCount.value = res.crossContextSharedCount || 0
    uncategorizedCount.value = res.uncategorizedCount || 0
    staleCount.value = res.staleCount || 0
    categoryCountMap.value = Object.fromEntries(
      (res.categoryCounts || []).map((item) => [item.categoryId, item.count || 0])
    )
  }

  async function applyTableFilters() {
    if (!ensureManagedAppReady()) {
      return
    }
    Object.assign(searchParams, {
      appKey: targetAppKey.value,
      source: selectedSource.value || undefined,
      method: tableQuery.method || undefined,
      path: tableQuery.path || undefined,
      keyword: tableQuery.keyword || undefined,
      permissionKey: tableQuery.permissionKey || undefined,
      permissionPattern: tableQuery.permissionPattern || undefined,
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
    syncQueryFromSearchForm()
    await applyTableFilters()
  }

  async function resetTableQuery() {
    searchForm.source = ''
    searchForm.method = ''
    searchForm.path = ''
    searchForm.keyword = ''
    searchForm.permissionKey = ''
    searchForm.permissionPattern = ''
    searchForm.contextScope = ''
    searchForm.featureKind = ''
    searchForm.status = ''
    searchForm.hasPermissionKey = ''
    syncQueryFromSearchForm()
    selectedCategoryTreeKey.value = 'all'
    tableQuery.categoryId = ''
    tableQuery.hasCategory = ''
    await applyTableFilters()
  }

  onMounted(async () => {
    restoreTableState()
    await Promise.all([loadCategories(), loadUnregisteredCount()])
    if (!targetAppKey.value) {
      resetScopedState()
      return
    }
    await loadCategorySummary()
    syncCategoryFilterFromTree()
    await applyTableFilters()
  })

  watch(
    () => targetAppKey.value,
    async () => {
      if (!targetAppKey.value) {
        resetScopedState()
        if (unregisteredVisible.value) {
          await loadUnregisteredRoutes()
        }
        if (staleDialogVisible.value) {
          await loadStaleCandidates()
        }
        return
      }
      loadError.value = ''
      await Promise.all([loadCategorySummary(), loadUnregisteredCount()])
      if (unregisteredVisible.value) {
        await loadUnregisteredRoutes()
      }
      if (staleDialogVisible.value) {
        await loadStaleCandidates()
      }
      await applyTableFilters()
    }
  )
</script>

<style scoped>
  .api-inline-alert {
    margin-bottom: 12px;
  }

  .module-card {
    margin-bottom: 12px;
  }

  .tree-card,
  .api-table-card {
    height: 100%;
    min-height: 0;
    max-height: none;
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

  .module-tag {
    cursor: pointer;
  }

  .tree-card :deep(.el-card__body) {
    flex: 1;
    min-height: 0;
    padding: 12px 2px 12px 12px;
  }

  .tree-card :deep(.el-scrollbar) {
    display: flex;
    flex: 1;
    min-height: 0;
    flex-direction: column;
  }

  .tree-card :deep(.el-scrollbar__wrap) {
    min-height: 0;
  }

  .api-table-card {
    overflow: hidden;
  }

  .api-table-card :deep(.el-card__body) {
    display: flex;
    height: 100%;
    min-height: 0;
    flex-direction: column;
    padding: 14px 16px 12px;
    overflow: hidden;
  }

  .category-tree {
    width: 100%;
  }

  .category-tree :deep(.el-tree-node__content) {
    height: auto;
    min-height: 32px;
    padding-right: 4px;
    border-radius: 8px;
  }

  .category-tree-node {
    display: flex;
    position: relative;
    width: 100%;
    align-items: stretch;
    justify-content: space-between;
    gap: 6px;
    padding-right: 12px;
    font-size: 13px;
  }

  .category-tree-node-main {
    display: flex;
    min-width: 0;
    align-items: center;
    justify-content: flex-start;
    gap: 4px;
    flex: 1;
  }

  .category-tree-node-status {
    display: inline-flex;
    justify-content: flex-start;
    flex-shrink: 0;
  }

  .category-tree-node-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .category-tree-node-side {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    min-width: 60px;
    flex-shrink: 0;
    padding-right: 2px;
  }

  .category-tree-node-count {
    min-width: 18px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-align: right;
  }

  .category-tree-node-edit {
    display: none;
    position: static;
    min-height: 0;
    padding: 0;
    font-size: 12px;
    line-height: 1;
    color: var(--el-color-primary);
    background-color: transparent;
    border: 0;
    border-radius: 0;
    --el-button-hover-bg-color: transparent;
    --el-button-active-bg-color: transparent;
    --el-button-hover-border-color: transparent;
    --el-button-active-border-color: transparent;
  }

  .category-tree-node:hover .category-tree-node-edit {
    display: inline-flex;
  }

  .category-tree-node-edit:hover {
    color: var(--el-color-primary-light-3);
    background-color: transparent;
    border: 0;
  }

  :deep(.category-tree-node-edit.el-button),
  :deep(.category-tree-node-edit.el-button:hover),
  :deep(.category-tree-node-edit.el-button:focus),
  :deep(.category-tree-node-edit.el-button:active),
  :deep(.category-tree-node-edit.el-button.is-active) {
    background-color: transparent !important;
    border-color: transparent !important;
    box-shadow: none !important;
  }

  .api-table-main {
    flex: 1;
    min-height: 0;
  }

  .api-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .api-hero-actions__group {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .api-hero-actions__group--secondary {
    padding-left: 12px;
    margin-left: 2px;
    border-left: 1px solid rgb(203 213 225 / 0.9);
  }

  .api-table-card :deep(.table-header-left) {
    gap: 12px;
    row-gap: 8px;
  }

  .api-table-card :deep(#art-table-header) {
    align-items: flex-start;
    margin-bottom: 8px;
  }

  .api-table-card :deep(#art-table-header > .flex-wrap) {
    align-items: flex-start;
  }

  .api-table-card :deep(#art-table-header > .flex-c) {
    padding-top: 0;
  }

  .api-table-card :deep(.el-button) {
    --el-component-size: 34px;
  }

  .api-table-card :deep(.el-button + .el-button) {
    margin-left: 8px;
  }

  .api-table-card :deep(.el-input__wrapper),
  .api-table-card :deep(.el-select__wrapper) {
    min-height: 34px;
  }

  .api-table-card :deep(.el-tag) {
    font-size: 12px;
  }

  .api-table-card :deep(.el-table) {
    font-size: 13px;
  }

  .api-table-card :deep(.el-table th),
  .api-table-card :deep(.el-table td) {
    padding: 8px 0;
  }

  .api-table-card :deep(.el-pagination) {
    padding-top: 8px;
    margin-top: 8px;
  }

  .category-list {
    display: grid;
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

  .category-drawer {
    display: flex;
    height: 100%;
    flex-direction: column;
    gap: 16px;
  }

  .category-drawer-toolbar {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .category-drawer-content {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 320px;
    gap: 16px;
    min-height: 0;
  }

  .category-drawer-list {
    min-height: 0;
  }

  .category-form-panel {
    padding: 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: linear-gradient(180deg, #fff, var(--el-fill-color-extra-light));
  }

  .category-form-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 16px;
  }

  .category-form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .category-input-wrap {
    display: flex;
    width: 100%;
    gap: 8px;
  }

  .category-input-wrap :deep(.el-select) {
    flex: 1;
  }

  .api-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .form-intro {
    padding: 14px 16px;
    border: 1px solid rgb(226 232 240 / 0.95);
    border-radius: 16px;
    background: linear-gradient(135deg, rgb(248 250 252 / 0.98), rgb(241 245 249 / 0.95));
  }

  .form-intro__title,
  .form-section__title {
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
  }

  .form-intro__text,
  .form-section__desc {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .form-section {
    padding: 14px 16px 4px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 16px;
    background: rgb(255 255 255 / 0.96);
  }

  .form-section__header {
    margin-bottom: 12px;
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

  .status-cell {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .status-note {
    font-size: 12px;
    line-height: 1.4;
    color: var(--el-text-color-secondary);
  }

  .permission-structure-cell {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .permission-structure-contexts,
  .permission-structure-note {
    font-size: 12px;
    line-height: 1.4;
    color: var(--el-text-color-secondary);
  }

  .operate-cell {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .unregistered-toolbar {
    display: grid;
    grid-template-columns: minmax(100px, 140px) repeat(3, minmax(0, 1fr)) auto auto;
    gap: 12px;
    align-items: center;
    margin-bottom: 16px;
  }

  .unregistered-method-select {
    width: 100%;
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

  .stale-dialog-tip {
    margin-bottom: 12px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .stale-dialog-footer {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 12px;
  }

  .stale-dialog-footer-meta {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 0;
  }

  .stale-dialog-footer-text {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .stale-dialog-footer-actions {
    display: flex;
    gap: 8px;
  }

  .toolbar-count {
    margin-left: 4px;
  }

  .label-help {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .label-help-icon {
    font-size: 14px;
    color: var(--el-text-color-secondary);
    cursor: help;
  }

  @media (max-width: 1280px) {
    .tree-card,
    .api-table-card {
      position: static;
      height: auto;
      min-height: 0;
      max-height: none;
    }

    .tree-card :deep(.el-card__body),
    .api-table-card :deep(.el-card__body) {
      height: auto;
      min-height: 320px;
    }

    .category-drawer-content {
      grid-template-columns: 1fr;
    }

    .unregistered-toolbar {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .unregistered-footer {
      flex-direction: column;
      align-items: flex-start;
    }

    .stale-dialog-footer {
      flex-direction: column;
      align-items: stretch;
    }

    .api-hero-actions__group--secondary {
      padding-left: 0;
      margin-left: 0;
      border-left: 0;
    }
  }

  @media (max-width: 768px) {
    .category-drawer-toolbar,
    .category-form-header {
      flex-direction: column;
      align-items: stretch;
    }
  }
</style>

