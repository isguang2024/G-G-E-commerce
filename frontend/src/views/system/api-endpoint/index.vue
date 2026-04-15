<template>
  <div class="art-full-height api-endpoint-page">
    <div class="api-endpoint-page__layout">
      <aside class="api-endpoint-page__sidebar">
        <ElCard class="tree-card art-card-xs" shadow="never">
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
      </aside>

      <section class="api-endpoint-page__main">
        <div class="page-top-stack">
          <ApiEndpointSearch
            v-show="showSearchBar"
            v-model="searchForm"
            @search="handleTableSearch"
            @reset="resetTableQuery"
          />

          <ElCard class="art-table-card api-table-card" shadow="never">
            <AdminWorkspaceHero
              title="API 管理"
              description="维护接口注册、分类、权限键与运行时状态，未注册和失效接口也在同一页诊断。"
              :metrics="summaryMetrics"
            >
              <div class="api-hero-actions">
                <div class="api-hero-actions__group">
                  <ElButton
                    v-action="'system.api_registry.sync'"
                    plain
                    :loading="syncing"
                    data-testid="api-endpoint-sync-button"
                    :data-loading="syncing ? '1' : '0'"
                    @click="handleSync"
                    v-ripple
                  >
                    全局同步 API
                  </ElButton>
                </div>

                <div class="api-hero-actions__group api-hero-actions__group--secondary">
                  <ElButton
                    v-action="'system.api_registry.view'"
                    plain
                    data-testid="api-endpoint-unregistered-button"
                    :data-count="unregisteredCount"
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
                    data-testid="api-endpoint-cleanup-stale-button"
                    :data-loading="cleaningStale ? '1' : '0'"
                    :data-count="staleCount"
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

            <ArtTable
              :loading="loading"
              :data="data"
              :columns="columns"
              :pagination="pagination"
              empty-height="320px"
              @pagination:size-change="handleSizeChange"
              @pagination:current-change="handleCurrentChange"
            />
          </ElCard>
        </div>
      </section>
    </div>

    <ElDrawer
      v-model="formVisible"
      title="编辑 API"
      size="760px"
      direction="rtl"
      destroy-on-close
      class="config-drawer"
    >
      <ElForm :model="formState" label-width="110px" class="api-form">
        <div class="form-intro">
          <div class="form-intro__title">调整接口元数据</div>
          <div class="form-intro__text">
            先明确接口身份，再配置分类、协作空间要求和权限键。这里保存的是正式注册信息。
          </div>
        </div>

        <div class="form-section">
          <div class="form-section__header">
            <div class="form-section__title">接口身份</div>
            <div class="form-section__desc">Method、路径和说明决定这条接口如何进入正式注册表。</div>
          </div>
          <ElFormItem label="Method" prop="method">
            <ElSelect
              v-model="formState.method"
              placeholder="请选择"
              popper-class="api-endpoint-select-popper"
            >
              <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="路径" prop="path">
            <ElInput v-model="formState.path" placeholder="/..." />
          </ElFormItem>
          <ElFormItem label="说明" prop="summary">
            <ElInput v-model="formState.summary" placeholder="说明这条接口的用途和影响范围" />
          </ElFormItem>
        </div>

        <div class="form-section">
          <div class="form-section__header">
            <div class="form-section__title">归属与运行时</div>
            <div class="form-section__desc">分类影响管理归档，状态影响运行时访问与诊断结果。</div>
          </div>
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
        </div>

        <div class="form-section">
          <div class="form-section__header">
            <div class="form-section__title">权限绑定</div>
            <div class="form-section__desc"
              >权限键决定这条接口会被哪条能力链消费，没有权限键时会更接近基础接口。</div
            >
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

    <ElDialog
      v-model="staleDialogVisible"
      title="清理失效 API"
      width="980px"
      destroy-on-close
      data-testid="api-endpoint-stale-dialog"
    >
      <div class="stale-dialog-tip">
        仅会删除"来源为自动同步、且源码中已不存在"的失效 API，请勾选后执行删除。
      </div>

      <ElTable
        ref="staleTableRef"
        :data="staleCandidates"
        border
        height="420px"
        row-key="id"
        data-testid="api-endpoint-stale-table"
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
            <ElButton
              type="danger"
              :loading="cleaningStale"
              data-testid="api-endpoint-stale-confirm-button"
              :data-selected="selectedStaleIds.length"
              @click="submitCleanupStale"
            >
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
  // 视图脚本：所有 reactive state、handler、useTable、watch 均在 useApiEndpointPage 中
  // 这里只做：1) 调用 composable；2) 把返回值拉到 setup 作用域供模板访问。
  // 拆分前 1300+ 行，拆分后 ~80 行。
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { QuestionFilled } from '@element-plus/icons-vue'
  import ApiEndpointSearch from './modules/api-endpoint-search.vue'
  import { useApiEndpointPage } from './modules/use-api-endpoint-page'

  defineOptions({ name: 'ApiEndpoint' })

  const {
    methodOptions,
    targetAppKey,
    methodTagType,
    syncing,
    cleaningStale,
    loadError,
    showSearchBar,
    saving,
    categorySaving,
    categorySwitchingId,
    selectedCategoryTreeKey,
    formVisible,
    categoryDrawerVisible,
    permissionBindVisible,
    permissionDialogMode,
    unregisteredVisible,
    staleDialogVisible,
    unregisteredLoading,
    editingId,
    permissionBinding,
    permissionActionLoading,
    unregisteredRoutes,
    staleCandidates,
    selectedStaleIds,
    staleCount,
    unregisteredCount,
    unregisteredPagination,
    stalePagination,
    formState,
    categoryForm,
    unregisteredQuery,
    searchForm,
    sortedCategories,
    categoryTreeData,
    summaryMetrics,
    currentPermissionActionOptions,
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    submitPermissionBind,
    handleSync,
    handleCleanupStale,
    closeStaleDialog,
    handleStaleSelectionChange,
    handleStaleCurrentChange,
    handleStaleSizeChange,
    submitCleanupStale,
    openEditDialog: _openEditDialog,
    startCreateCategory,
    openCategoryDrawer,
    handleCategoryTreeSelect,
    openUnregisteredDialog,
    handleUnregisteredSearch,
    handleUnregisteredCurrentChange,
    handleUnregisteredSizeChange,
    resetUnregisteredQuery,
    submitCategory,
    toggleCategoryStatus,
    submitForm,
    handleTableSearch,
    resetTableQuery,
    resetCategoryForm,
    staleTableRef
  } = useApiEndpointPage()
</script>

<style scoped>
  .api-endpoint-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-height: 0;
  }

  .api-endpoint-page__layout {
    display: grid;
    grid-template-columns: 232px minmax(0, 1fr);
    grid-template-rows: minmax(0, 1fr);
    gap: 16px;
    align-items: stretch;
    flex: 1;
    min-height: 0;
  }

  .api-endpoint-page__sidebar,
  .api-endpoint-page__main {
    min-width: 0;
  }

  .api-endpoint-page__main {
    display: flex;
    flex: 1;
    min-height: 0;
    flex-direction: column;
  }

  .api-inline-alert {
    margin-bottom: 12px;
  }

  .module-card {
    margin-bottom: 12px;
  }

  .page-top-stack {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 16px;
    min-height: 0;
  }

  .page-top-stack > .api-table-card {
    flex: 1;
    min-height: 0;
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
    padding: 12px 2px 12px 12px;
  }

  .tree-card :deep(.el-scrollbar) {
    max-height: calc(100vh - 220px);
  }

  .api-table-card {
    min-width: 0;
    height: 100%;
  }

  .api-table-card :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    padding: 14px 16px 12px;
    gap: 12px;
  }

  .api-table-card :deep(.art-table) {
    flex: 1;
    min-height: 0;
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

  .api-table-card :deep(.pagination.custom-pagination) {
    align-items: center;
    margin-top: 6px;
    padding: 2px 0 10px;
  }

  .api-table-card :deep(.el-pagination) {
    padding-top: 0;
    padding-bottom: 0;
    margin-top: 0;
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
    .api-endpoint-page__layout {
      grid-template-columns: 1fr;
      grid-template-rows: auto;
      align-items: stretch;
    }

    .tree-card :deep(.el-card__body),
    .api-table-card :deep(.el-card__body) {
      min-height: 0;
    }

    .tree-card :deep(.el-scrollbar) {
      max-height: 320px;
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
