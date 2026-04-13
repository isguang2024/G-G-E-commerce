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
            v-model="selectedAppKey"
            class="menu-app-select"
            clearable
            filterable
            placeholder="选择 App"
            @change="handleManagedAppChange"
          >
            <ElOption
              v-for="item in appOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
          <ElSelect
            v-if="selectedAppKey"
            v-model="activeSpaceKey"
            class="menu-space-select"
            filterable
            placeholder="选择菜单空间"
            @change="handleSpaceChange"
          >
            <ElOption
              v-for="item in menuSpaceOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
          <ElButton v-action="'system.menu.manage'" type="primary" @click="handleAddMenu" v-ripple>
            {{ isLayoutMode ? '创建布局菜单' : '创建菜单定义' }}
          </ElButton>
          <ElButton v-if="isLayoutMode" @click="goToDefinitionManagement" v-ripple>
            返回定义管理
          </ElButton>
        </div>
      </AdminWorkspaceHero>
    </div>

    <ElCard class="art-table-card" shadow="never">
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
            </div>
            <div class="menu-toolbar-bottom">
              <div class="menu-toolbar-switches">
                <span class="menu-switch-item">
                  <span class="menu-switch-label">显示隐藏菜单</span>
                  <ElSwitch v-model="showHiddenMenus" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">显示内嵌菜单</span>
                  <ElSwitch v-model="showIframeMenus" />
                </span>
                <span class="menu-switch-item">
                  <span class="menu-switch-label">显示启用菜单</span>
                  <ElSwitch v-model="showEnabledMenus" />
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
          <ArtSvgIcon v-if="row.meta?.icon" :icon="row.meta.icon" class="mr-2 text-g-500" />
          <span>{{ formatMenuTitle(row.meta?.title) }}</span>
        </template>

        <!-- 菜单类型列 -->
        <template #type="{ row }">
          <ElTag :type="getMenuTypeTag(row)">{{ getMenuTypeText(row) }}</ElTag>
        </template>

        <!-- 路由列 -->
        <template #path="{ row }">
          <span>{{ row.meta?.link || row.path || '' }}</span>
        </template>

        <!-- 组件路径列 -->
        <template #component="{ row }">
          <span class="text-gray-600">{{ row.component || '-' }}</span>
        </template>

        <template #linkedPage="{ row }">
          <div class="menu-linked-page-cell">
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
        </template>

        <template #space="{ row }">
          <ElTag size="small" effect="plain" type="info">
            {{ getSpaceName(row.spaceKey || row.meta?.spaceKey) }}
          </ElTag>
        </template>

        <!-- 高级配置列 -->
        <template #advanced="{ row }">
          <div class="advanced-configs">
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
            <ElTag
              v-if="!isDirectoryMenuRow(row) && row.meta?.isIframe"
              size="small"
              effect="light"
              type="info"
              class="mr-2"
            >
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
            <ElTag
              v-if="isEntryMenuRow(row) && row.meta?.fixedTab"
              size="small"
              effect="light"
              type="danger"
              class="mr-2"
            >
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
        </template>

        <!-- 状态列 -->
        <template #status="{ row }">
          <ElTag :type="row.meta?.isEnable !== false ? 'success' : 'info'">
            {{ row.meta?.isEnable !== false ? '启用' : '未启用' }}
          </ElTag>
        </template>

        <!-- 操作列 -->
        <template #operation="{ row }">
          <div class="flex items-center justify-center gap-2">
            <ArtButtonMore
              :list="getOperationList(row)"
              @click="(item) => handleMenuOperation(item, row)"
            />
          </div>
        </template>
      </ArtTable>

      <!-- 菜单弹窗 -->
      <MenuDialog
        v-model:visible="dialogVisible"
        :editData="editData"
        :menuTree="filteredMenuTree"
        :menuSpaces="menuSpaces"
        :currentSpaceKey="activeSpaceKey"
        :currentMenuPages="getLinkedPages(editData || {})"
        :editingMenuId="editData?.id !== undefined ? String(editData.id) : undefined"
        :initialParentId="String(parentRowForAdd?.id ?? '')"
        :showSpaceField="isLayoutMode"
        @submit="handleSubmit"
      />

      <MenuPermissionDialog
        v-model="actionRequirementVisible"
        :menuData="actionRequirementData"
        @submit="handleActionRequirementSubmit"
      />

      <MenuDeleteDialog
        v-model:visible="deleteDialogVisible"
        :loading="deleteLoading"
        :menuTitle="formatMenuTitle(String(deleteTargetRow?.meta?.title ?? '')) || String(deleteTargetRow?.name ?? '') || ''"
        :childCount="getMenuChildCount(deleteTargetRow)"
        :descendantCount="getMenuDescendantCount(deleteTargetRow)"
        :affectedPageCount="getAffectedPageCount(deleteTargetRow)"
        :affectedRelationCount="deletePreview?.affectedRelationCount || 0"
        :parentOptions="getDeleteParentOptions(deleteTargetRow)"
        @confirm="handleDeleteMenuConfirm"
      />

    </ElCard>
  </div>
</template>

<script setup lang="ts">
  // 视图脚本：所有 reactive state、handler、watch、lifecycle 均在 useMenuPage 中
  // 这里只做：1) 引入子组件；2) 调用 composable；3) 把返回值拉到 setup 作用域供模板访问。
  // 拆分前 1280+ 行，拆分后 ~120 行。
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import {
    ElAlert,
    ElButton,
    ElCard,
    ElOption,
    ElSelect,
    ElSwitch,
    ElTag
  } from 'element-plus'
  import MenuDeleteDialog from './modules/menu-delete-dialog.vue'
  import MenuDialog from './modules/menu-dialog.vue'
  import MenuPermissionDialog from './modules/menu-permission-dialog.vue'
  import MenuSearch from './modules/menu-search.vue'
  import { useMenuPage } from './modules/use-menu-page'

  defineOptions({ name: 'Menus' })

  const {
    loading,
    loadError,
    showSearchBar,
    isExpanded,
    showHiddenMenus,
    showIframeMenus,
    showEnabledMenus,
    tableRef,
    multiSelectEnabled,
    activeSpaceKey,
    selectedAppKey,
    menuSpaces,
    formFilters,
    dialogVisible,
    editData,
    parentRowForAdd,
    deleteDialogVisible,
    deleteLoading,
    deleteTargetRow,
    deletePreview,
    actionRequirementVisible,
    actionRequirementData,
    selectedMenuRows,
    isLayoutMode,
    menuSpaceOptions,
    appOptions,
    menuPageTitle,
    menuPageDescription,
    menuToolbarTip,
    filteredMenuTree,
    tableData,
    menuHeroMetrics,
    columnChecks,
    displayColumns,
    getLinkedPages,
    getSpaceName,
    getMenuActionRequirementLabel,
    getOperationList,
    handleReset,
    handleSearch,
    goToDefinitionManagement,
    handleSpaceChange,
    handleManagedAppChange,
    rowKey,
    handleBatchSelectionChange,
    getMenuChildCount,
    getMenuDescendantCount,
    getAffectedPageCount,
    getDeleteParentOptions,
    handleExpandSwitchChange,
    handleAddMenu,
    handleMenuOperation,
    handleDeleteMenuConfirm,
    handleSubmit,
    handleActionRequirementSubmit,
    isDirectoryMenuRow,
    isEntryMenuRow,
    getMenuTypeTag,
    getMenuTypeText,
    getAccessModeLabel,
    getAccessModeTag,
    formatMenuTitle,
    getMenuActionRequirement
  } = useMenuPage()
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

  .menu-app-select {
    width: 240px;
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
