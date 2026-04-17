<template>
  <div class="page-management-page art-full-height">
    <div class="page-top-stack">
      <ArtSearchBar
        class="page-search-bar"
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
        title="受管页面"
        description="只管理非菜单直达页、逻辑分组与普通分组；菜单入口页回到菜单定义或空间布局维护。"
        :metrics="summaryStats"
      >
        <div class="page-hero-actions">
          <AppKeySelect
            v-model="selectedAppKey"
            placeholder="选择 App"
            class="page-app-select"
            @change="handleManagedAppChange"
          />
          <ElDropdown trigger="click" @command="handleCreateCommand">
            <ElButton v-action="'system.page.manage'" type="primary" v-ripple> 新增 </ElButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="page">新增受管页面</ElDropdownItem>
                <ElDropdownItem command="group">新增逻辑分组</ElDropdownItem>
                <ElDropdownItem command="display_group">新增普通分组</ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
          <ElButton
            v-action="'system.page.sync'"
            :loading="syncing"
            @click="unregisteredDialogVisible = true"
            v-ripple
          >
            扫描未注册受管页
          </ElButton>
        </div>
      </AdminWorkspaceHero>
    </div>

    <ElCard class="art-table-card" shadow="never">
      <ElAlert
        class="page-governance-alert"
        type="info"
        :closable="false"
        show-icon
        title="共存治理规则：本地配置页负责当前后台直接维护；扫描同步页优先回到扫描源修正；远端页只保留 link / remote meta 作为入口，不再补同一路由的本地组件。"
      />
      <ElAlert
        v-if="loadError"
        class="page-inline-alert"
        type="info"
        :closable="false"
        show-icon
        :title="loadError"
      />
      <ArtTableHeader
        :loading="loading"
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        @refresh="handleRefresh"
      >
        <template #left>
          <div class="page-toolbar">
            <div class="page-toolbar-tip">
              页面挂载关系、访问方式和父链路在这里集中治理，避免菜单入口页与受管页重复维护。
            </div>
            <div class="page-toolbar-actions">
              <div v-if="menuSpaces.length > 1" class="page-space-filter">
                <span class="page-space-filter__label">配置空间</span>
                <ElSelect
                  v-model="activeSpaceKey"
                  clearable
                  placeholder="全部空间"
                  style="width: 180px"
                  @change="handleSpaceScopeChange"
                >
                  <ElOption label="全部空间" value="" />
                  <ElOption
                    v-for="item in menuSpaces"
                    :key="item.menuSpaceKey"
                    :label="item.isDefault ? `${item.name}（默认）` : item.name"
                    :value="item.menuSpaceKey"
                  />
                </ElSelect>
              </div>
              <div class="page-switch">
                <span class="page-switch__label">展开分组</span>
                <ElSwitch v-model="isExpanded" @change="handleExpandSwitchChange" />
              </div>
              <div class="page-switch">
                <span class="page-switch__label">显示停用</span>
                <ElSwitch v-model="showSuspended" />
              </div>
            </div>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        :rowKey="rowKey"
        :loading="loading"
        :data="tableData"
        :columns="displayColumns"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="false"
      >
        <template #name="{ row }">
          <div
            :class="[
              'page-name-cell',
              {
                'page-name-cell--logic-group': row.pageType === 'group',
                'page-name-cell--display-group': row.pageType === 'display_group'
              }
            ]"
          >
            <div class="page-name-cell__main">
              <div class="page-name-cell__title">
                <ElTag :type="getPageTypeTag(row)" effect="plain" size="small">
                  {{ getPageTypeText(row) }}
                </ElTag>
                <ElTag :type="getPageSourceTag(row)" effect="plain" size="small">
                  {{ getPageSourceText(row) }}
                </ElTag>
                <span class="page-name-cell__text">{{ row.name }}</span>
                <span class="page-inline-relation">{{ getRelationDisplayText(row) }}</span>
              </div>
              <div class="page-name-cell__subtext">
                {{ getPageGovernanceText(row) }}
              </div>
            </div>
          </div>
        </template>

        <template #route="{ row }">
          <div class="page-route-cell">
            <code :class="['page-route-text', { 'page-muted-text': !getRouteDisplayText(row) }]">
              {{ getRouteDisplayText(row) || '-' }}
            </code>
          </div>
        </template>

        <template #component="{ row }">
          <div class="page-component-cell">
            <span class="page-muted-text">
              {{
                row.pageType === 'group' || row.pageType === 'display_group'
                  ? '不需要组件'
                  : row.component || '-'
              }}
            </span>
          </div>
        </template>

        <template #sortOrder="{ row }">
          <div class="page-sort-cell">
            <template v-if="editingSortId === row.id">
              <ElInput
                v-model="sortDraftMap[row.id]"
                size="small"
                class="page-sort-input"
                inputmode="numeric"
              />
              <div class="page-sort-actions">
                <ElButton
                  type="primary"
                  link
                  size="small"
                  :loading="savingSortIds.has(row.id)"
                  @click="saveSortOrder(row)"
                >
                  保存
                </ElButton>
                <ElButton link size="small" @click="cancelSortEdit(row)">取消</ElButton>
              </div>
            </template>
            <template v-else>
              <div class="page-sort-view">
                <span class="page-sort-value">{{ row.sortOrder ?? 0 }}</span>
                <ElButton
                  type="primary"
                  link
                  size="small"
                  class="page-sort-edit-btn"
                  @click="startSortEdit(row)"
                >
                  编辑
                </ElButton>
              </div>
            </template>
          </div>
        </template>

        <template #accessMode="{ row }">
          <div class="page-access-cell">
            <ElTag effect="plain" size="small" type="info">
              {{ getMountModeText(row) }}
            </ElTag>
            <ElTag
              v-if="row.pageType !== 'display_group'"
              :effect="'plain'"
              :type="getAccessModeTag(row.accessMode)"
            >
              {{ getAccessModeText(row.accessMode) }}
            </ElTag>
            <span v-else class="page-muted-text">-</span>
            <ElTag :type="row.status === 'normal' ? 'success' : 'danger'" effect="light">
              {{ row.status === 'normal' ? '正常' : '停用' }}
            </ElTag>
          </div>
        </template>

        <template #effectiveChain="{ row }">
          <span class="page-muted-text">{{ getEffectiveChainText(row) }}</span>
        </template>

        <template #mountTarget="{ row }">
          <span class="page-muted-text">{{ getMountTargetText(row) }}</span>
        </template>

        <template #parentChainStatus="{ row }">
          <span
            :class="[
              'page-muted-text',
              { 'page-chain-status--error': getParentChainStatusText(row).startsWith('异常') }
            ]"
          >
            {{ getParentChainStatusText(row) }}
          </span>
        </template>

        <template #updatedAt="{ row }">
          <span class="page-muted-text">{{ formatUpdatedAt(row.updatedAt) }}</span>
        </template>

        <template #operation="{ row }">
          <div class="flex items-center justify-center gap-2">
            <ArtButtonMore
              :list="getOperationList(row)"
              @click="(item) => handleOperation(item, row)"
            />
          </div>
        </template>
      </ArtTable>
    </ElCard>

    <PageDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :page-data="currentPage"
      :default-data="defaultPageData"
      :app-key="targetAppKey"
      :menu-spaces="menuSpaces"
      :initial-parent-page-key="initialParentPageKey"
      :initial-parent-menu-id="initialParentMenuId"
      :initial-page-type="initialPageType"
      @success="handleRefresh"
    />
    <PageUnregisteredDialog
      v-model="unregisteredDialogVisible"
      :app-key="targetAppKey"
      @synced="handleRefresh"
      @create-candidate="handleCreateFromCandidate"
    />
  </div>
</template>

<script setup lang="ts">
  // 视图脚本：所有 reactive state、handler、watch、lifecycle 均在 usePagePage 中
  // 这里只做：1) 引入子组件；2) 调用 composable；3) 把返回值拉到 setup 作用域供模板访问。
  import { ElButton, ElInput, ElOption, ElSelect, ElTag } from 'element-plus'
  import AppKeySelect from '@/components/business/app/AppKeySelect.vue'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import PageDialog from './modules/page-dialog.vue'
  import PageUnregisteredDialog from './modules/page-unregistered-dialog.vue'
  import { usePagePage } from './modules/use-page-page'

  defineOptions({ name: 'PageManagement' })

  const {
    loading,
    loadError,
    showSearchBar,
    isExpanded,
    syncing,
    showSuspended,
    sortDraftMap,
    savingSortIds,
    editingSortId,
    tableRef,
    targetAppKey,
    selectedAppKey,
    menuSpaces,
    activeSpaceKey,
    dialogVisible,
    dialogType,
    currentPage,
    defaultPageData,
    initialParentPageKey,
    initialParentMenuId,
    initialPageType,
    unregisteredDialogVisible,
    searchForm,
    searchItems,
    columnChecks,
    displayColumns,
    tableData,
    summaryStats,
    handleSearch,
    handleReset,
    handleRefresh,
    handleSpaceScopeChange,
    handleManagedAppChange,
    handleCreateCommand,
    handleOperation,
    handleCreateFromCandidate,
    handleExpandSwitchChange,
    getOperationList,
    getRouteDisplayText,
    getRelationDisplayText,
    getPageGovernanceText,
    getPageSourceTag,
    getPageSourceText,
    getMountTargetText,
    getEffectiveChainText,
    getParentChainStatusText,
    getPageTypeTag,
    getPageTypeText,
    getAccessModeText,
    getAccessModeTag,
    getMountModeText,
    formatUpdatedAt,
    startSortEdit,
    cancelSortEdit,
    saveSortOrder,
    rowKey
  } = usePagePage()

</script>

<style lang="scss" scoped>
  .page-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .page-app-select {
    width: 240px;
  }

  .page-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 12px 16px;
    width: 100%;
  }

  .page-inline-alert {
    margin-bottom: 12px;
  }

  .page-governance-alert {
    margin-bottom: 12px;
  }

  .page-toolbar-tip {
    color: var(--art-text-muted);
    font-size: 12px;
    line-height: 1.6;
  }

  .page-toolbar-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: flex-start;
  }

  .page-space-filter {
    align-items: center;
    display: inline-flex;
    gap: 8px;
  }

  .page-space-filter__label {
    color: var(--art-text-muted);
    font-size: 12px;
    white-space: nowrap;
  }

  .page-switch {
    align-items: center;
    display: inline-flex;
    gap: 6px;
    margin-left: 4px;
  }

  .page-switch__label {
    color: var(--art-text-base);
    font-size: 12px;
    line-height: 1;
    white-space: nowrap;
  }

  :deep(.page-search-bar .el-form-item__label) {
    white-space: nowrap;
  }

  .page-name-cell {
    display: flex;
    align-items: flex-start;
    flex: 1;
    min-width: 0;
  }

  .page-name-cell--logic-group {
    color: var(--art-text-strong);
    font-weight: 600;
  }

  .page-name-cell--display-group {
    color: color-mix(in srgb, var(--el-color-success-dark-2) 72%, black);
    font-weight: 600;
  }

  .page-name-cell__main {
    display: flex;
    flex: 1;
    align-items: flex-start;
    flex-direction: column;
    min-width: 0;
  }

  .page-name-cell__title {
    align-items: center;
    display: flex;
    flex-wrap: nowrap;
    gap: 8px;
    min-width: 0;
  }

  .page-name-cell__text {
    color: var(--art-text-strong);
    font-size: 14px;
    font-weight: 600;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-inline-relation {
    color: var(--art-text-muted);
    font-size: 12px;
    margin-left: 8px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-name-cell__subtext {
    color: var(--art-text-muted);
    font-size: 12px;
    line-height: 1.5;
    margin-top: 6px;
  }

  .page-muted-text {
    color: var(--art-text-muted);
    font-size: 12px;
    line-height: 1.4;
  }

  .page-route-cell,
  .page-component-cell {
    display: flex;
    align-items: center;
    min-width: 0;
  }

  .page-route-text {
    color: var(--art-text-strong);
    display: inline-block;
    font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
    font-size: 12px;
    line-height: 1.5;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-access-cell {
    display: inline-flex;
    align-items: center;
    flex-wrap: nowrap;
    gap: 6px;
    justify-content: center;
    white-space: nowrap;
  }

  .page-chain-status--error {
    color: var(--el-color-danger);
  }

  .page-sort-cell {
    align-items: center;
    display: inline-flex;
    gap: 4px;
    justify-content: center;
    white-space: nowrap;
    width: 100%;
  }

  .page-sort-input {
    width: 84px;
  }

  .page-sort-value {
    color: var(--art-text-strong);
    font-variant-numeric: tabular-nums;
    min-width: 24px;
    text-align: center;
  }

  .page-sort-view {
    align-items: center;
    display: flex;
    justify-content: center;
    position: relative;
    width: 100%;
  }

  :deep(.el-table .el-table__body .el-table__cell .page-sort-cell) {
    margin: 0 auto;
  }

  :deep(.page-sort-input .el-input__wrapper) {
    padding-left: 8px;
    padding-right: 8px;
  }

  .page-sort-actions {
    align-items: center;
    display: inline-flex;
    gap: 2px;
  }

  :deep(.page-sort-actions .el-button--small.is-link) {
    margin-left: 0;
    padding-left: 2px;
    padding-right: 2px;
  }

  .page-sort-edit-btn {
    position: absolute;
    right: 0;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.15s ease;
  }

  :deep(.el-table__body tr:hover .page-sort-edit-btn) {
    opacity: 1;
    pointer-events: auto;
  }

  :deep(.el-table .el-table__body .el-table__cell:nth-child(1) .cell) {
    display: flex;
    align-items: center;
  }

  :deep(.el-table .el-table__body .el-table__cell:nth-child(1) .el-table__expand-icon) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    margin-right: 8px;
    color: var(--el-text-color-regular);
    font-size: 18px;
    transform: scale(1.15);
    transform-origin: center;
  }

  :deep(.el-table .el-table__body tr:has(.page-name-cell--logic-group)) {
    background: color-mix(in srgb, var(--theme-color) 6%, white);
  }

  :deep(.el-table .el-table__body tr:has(.page-name-cell--display-group)) {
    background: color-mix(in srgb, var(--el-color-success-light-9) 45%, white);
  }

  @media (max-width: 960px) {
    .page-toolbar {
      align-items: flex-start;
    }

    .page-toolbar-actions {
      justify-content: flex-start;
      margin-left: 0;
    }
  }
</style>
