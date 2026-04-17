<template>
  <ElDrawer
    v-model="visible"
    :title="`权限测试 - ${userTitle}`"
    size="1080px"
    destroy-on-close
    direction="rtl"
    class="user-permission-drawer config-drawer"
  >
    <div class="permission-shell" v-loading="loading">
      <div class="dialog-note">
        这里用于诊断用户在个人空间或协作空间中的空间权限生效情况。所有权限都绑定在工作空间上；个人空间和协作空间只表示空间类型，不表示两套不同的权限系统。
      </div>

      <div class="toolbar-row">
        <ElSelect v-model="contextType" class="toolbar-select">
          <ElOption label="个人空间" value="personal" />
          <ElOption label="协作空间" value="collaboration" />
        </ElSelect>

        <ElSelect
          v-if="contextType === 'collaboration'"
          v-model="selectedWorkspaceId"
          filterable
          clearable
          class="toolbar-select toolbar-select--collaborationWorkspace"
          placeholder="请选择协作空间"
        >
          <ElOption
            v-for="collaborationWorkspace in workspaceOptions"
            :key="collaborationWorkspace.id"
            :label="collaborationWorkspace.name"
            :value="collaborationWorkspace.id"
          />
        </ElSelect>

        <ElInput
          v-model="permissionKey"
          class="toolbar-input"
          clearable
          placeholder="输入权限键，例如 system.user.manage"
          @keyup.enter="handleTest"
        />

        <ElButton :loading="testing" type="primary" @click="handleTest">测试权限</ElButton>
        <ElButton :loading="refreshing" @click="handleRefresh">刷新快照</ElButton>
      </div>

      <PermissionSummaryTags :items="summaryItems" />

      <ElTabs v-model="activeTab" class="permission-tabs">
        <ElTabPane label="权限测试" name="permission">
          <div class="panel-grid">
            <section class="info-panel">
              <div class="panel-title">快照信息</div>
              <div class="kv-grid">
                <div class="kv-item">
                  <span class="kv-label">上下文</span>
                  <span class="kv-value">{{
                    contextType === 'personal'
                      ? '个人空间'
                      : selectedWorkspaceName || '未选择协作空间'
                  }}</span>
                </div>
                <div class="kv-item">
                  <span class="kv-label">刷新时间</span>
                  <span class="kv-value">{{ diagnosisData?.snapshot?.refreshedAt || '-' }}</span>
                </div>
                <div class="kv-item">
                  <span class="kv-label">更新时间</span>
                  <span class="kv-value">{{ diagnosisData?.snapshot?.updatedAt || '-' }}</span>
                </div>
                <div class="kv-item">
                  <span class="kv-label">用户状态</span>
                  <span class="kv-value">
                    <ElTag
                      :type="diagnosisData?.user?.status === 'active' ? 'success' : 'danger'"
                      effect="plain"
                    >
                      {{ diagnosisData?.user?.status === 'active' ? '正常' : '停用' }}
                    </ElTag>
                  </span>
                </div>
                <div class="kv-item">
                  <span class="kv-label">角色数量</span>
                  <span class="kv-value">{{
                    diagnosisData?.snapshot?.roleCount ?? diagnosisData?.roles?.length ?? 0
                  }}</span>
                </div>
                <div class="kv-item">
                  <span class="kv-label">权限数量</span>
                  <span class="kv-value">
                    {{
                      contextType === 'personal'
                        ? (diagnosisData?.snapshot?.actionCount ?? 0)
                        : (diagnosisData?.snapshot?.effectiveActionCount ?? 0)
                    }}
                  </span>
                </div>
              </div>
            </section>

            <section class="info-panel">
              <div class="panel-title">测试结果</div>
              <div v-if="diagnosisData?.diagnosis" class="result-shell">
                <div
                  class="result-banner"
                  :class="
                    diagnosisData.diagnosis.allowed
                      ? 'result-banner--success'
                      : 'result-banner--danger'
                  "
                >
                  <div class="result-title">
                    {{ diagnosisData.diagnosis.allowed ? '权限通过' : '权限未通过' }}
                  </div>
                  <div class="result-text">
                    {{ diagnosisData.diagnosis.reasonText || '暂无说明' }}
                  </div>
                  <div v-if="diagnosisData.diagnosis.bypassedBySuperAdmin" class="result-hint">
                    当前结果来自超级管理员直通，不代表快照内已实际命中该权限。
                  </div>
                  <div
                    v-else-if="
                      diagnosisData.diagnosis.denialStage || diagnosisData.diagnosis.denialReason
                    "
                    class="result-hint"
                  >
                    {{
                      [
                        diagnosisData.diagnosis.denialStage
                          ? `拒绝层级：${diagnosisData.diagnosis.denialStage}`
                          : '',
                        diagnosisData.diagnosis.denialReason
                          ? `拒绝原因：${diagnosisData.diagnosis.denialReason}`
                          : ''
                      ]
                        .filter(Boolean)
                        .join('；')
                    }}
                  </div>
                </div>

                <div class="kv-grid">
                  <div class="kv-item">
                    <span class="kv-label">权限键</span>
                    <span class="kv-value">{{ diagnosisData.diagnosis.permissionKey || '-' }}</span>
                  </div>
                  <div class="kv-item">
                    <span class="kv-label">快照命中</span>
                    <span class="kv-value">
                      <ElTag
                        :type="diagnosisData.diagnosis.matchedInSnapshot ? 'success' : 'info'"
                        effect="plain"
                      >
                        {{ diagnosisData.diagnosis.matchedInSnapshot ? '已命中' : '未命中' }}
                      </ElTag>
                    </span>
                  </div>
                  <div v-if="contextType === 'collaboration'" class="kv-item">
                    <span class="kv-label">成员状态</span>
                    <span class="kv-value">
                      <ElTag
                        :type="getMemberStatusTagType(diagnosisData.diagnosis.memberStatus)"
                        effect="plain"
                      >
                        {{ formatMemberStatus(diagnosisData.diagnosis.memberStatus) }}
                      </ElTag>
                    </span>
                  </div>
                  <div v-if="contextType === 'collaboration'" class="kv-item">
                    <span class="kv-label">边界链路</span>
                    <span class="kv-value">
                      <ElTag
                        :type="getBoundaryStateTagType(diagnosisData.diagnosis.boundaryState)"
                        effect="plain"
                      >
                        {{ formatBoundaryState(diagnosisData.diagnosis.boundaryState) }}
                      </ElTag>
                    </span>
                  </div>
                  <div v-if="contextType === 'collaboration'" class="kv-item">
                    <span class="kv-label">角色链路</span>
                    <span class="kv-value">
                      <ElTag
                        :type="diagnosisData.diagnosis.roleChainMatched ? 'success' : 'info'"
                        effect="plain"
                      >
                        {{
                          diagnosisData.diagnosis.roleChainMatched
                            ? '命中'
                            : diagnosisData.diagnosis.roleChainDisabled
                              ? '禁用'
                              : diagnosisData.diagnosis.roleChainAvailable
                                ? '可用未生效'
                                : '未命中'
                        }}
                      </ElTag>
                    </span>
                  </div>
                  <div v-if="contextType === 'collaboration'" class="kv-item">
                    <span class="kv-label">拒绝层级</span>
                    <span class="kv-value">{{ diagnosisData.diagnosis.denialStage || '-' }}</span>
                  </div>
                  <div class="kv-item">
                    <span class="kv-label">直通放行</span>
                    <span class="kv-value">
                      <ElTag
                        :type="diagnosisData.diagnosis.bypassedBySuperAdmin ? 'warning' : 'info'"
                        effect="plain"
                      >
                        {{ diagnosisData.diagnosis.bypassedBySuperAdmin ? '超级管理员' : '否' }}
                      </ElTag>
                    </span>
                  </div>
                  <div class="kv-item">
                    <span class="kv-label">功能键状态</span>
                    <span class="kv-value">
                      {{ formatPermissionStatus(diagnosisData.diagnosis.action?.status) }}
                    </span>
                  </div>
                  <div class="kv-item">
                    <span class="kv-label">自身状态</span>
                    <span class="kv-value">
                      {{ formatPermissionStatus(diagnosisData.diagnosis.action?.selfStatus) }}
                    </span>
                  </div>
                  <div class="kv-item">
                    <span class="kv-label">模块分组</span>
                    <span class="kv-value">
                      {{
                        diagnosisData.diagnosis.action?.moduleGroup?.name
                          ? `${diagnosisData.diagnosis.action?.moduleGroup?.name} / ${formatPermissionStatus(
                              diagnosisData.diagnosis.action?.moduleGroupStatus
                            )}`
                          : '-'
                      }}
                    </span>
                  </div>
                  <div class="kv-item">
                    <span class="kv-label">功能分组</span>
                    <span class="kv-value">
                      {{
                        diagnosisData.diagnosis.action?.featureGroup?.name
                          ? `${diagnosisData.diagnosis.action?.featureGroup?.name} / ${formatPermissionStatus(
                              diagnosisData.diagnosis.action?.featureGroupStatus
                            )}`
                          : '-'
                      }}
                    </span>
                  </div>
                </div>

                <div class="source-panel">
                  <div class="source-title">来源功能包</div>
                  <div class="source-tags">
                    <ElTag
                      v-for="item in diagnosisData.diagnosis.sourcePackages || []"
                      :key="item.id"
                      effect="plain"
                      type="primary"
                      round
                    >
                      {{ item.name }}
                    </ElTag>
                    <span
                      v-if="!(diagnosisData.diagnosis.sourcePackages || []).length"
                      class="empty-text"
                      >未命中来源功能包</span
                    >
                  </div>
                </div>
                <div v-if="contextType === 'collaboration'" class="source-panel">
                  <div class="source-title">当前协作空间成员记录</div>
                  <div class="kv-grid">
                    <div class="kv-item">
                      <span class="kv-label">成员命中</span>
                      <span class="kv-value">
                        <ElTag
                          :type="
                            diagnosisData.collaborationWorkspaceMember?.matched
                              ? 'success'
                              : 'warning'
                          "
                          effect="plain"
                        >
                          {{
                            diagnosisData.collaborationWorkspaceMember?.matched
                              ? '已加入协作空间'
                              : '未加入协作空间'
                          }}
                        </ElTag>
                      </span>
                    </div>
                    <div class="kv-item">
                      <span class="kv-label">成员状态</span>
                      <span class="kv-value">{{
                        formatMemberStatus(diagnosisData.collaborationWorkspaceMember?.status)
                      }}</span>
                    </div>
                    <div class="kv-item">
                      <span class="kv-label">成员身份</span>
                      <span class="kv-value">{{
                        formatRoleCode(diagnosisData.collaborationWorkspaceMember?.roleCode)
                      }}</span>
                    </div>
                  </div>
                </div>
                <div v-if="contextType === 'collaboration'" class="source-panel">
                  <div class="source-title">当前协作空间功能包</div>
                  <div class="source-tags">
                    <ElTag
                      v-for="item in diagnosisData.collaborationWorkspacePackages || []"
                      :key="item.id"
                      effect="plain"
                      type="success"
                      round
                    >
                      {{ item.name }}
                    </ElTag>
                    <span
                      v-if="!(diagnosisData.collaborationWorkspacePackages || []).length"
                      class="empty-text"
                      >当前协作空间未开通功能包</span
                    >
                  </div>
                </div>
              </div>

              <ElEmpty
                v-else
                description="输入权限键后可直接测试当前用户在所选上下文中的权限结果"
              />
            </section>
          </div>
        </ElTabPane>

        <ElTabPane label="菜单测试" name="menus">
          <section class="menu-panel">
            <div class="menu-toolbar">
              <ElInput
                v-model="menuKeyword"
                clearable
                placeholder="搜索菜单标题或路由"
                :prefix-icon="Search"
                class="menu-toolbar__search"
              />
              <span class="menu-toolbar__switch">
                <span class="menu-toolbar__label">显示隐藏</span>
                <ElSwitch v-model="showHiddenMenus" size="small" />
              </span>
              <span class="menu-toolbar__switch">
                <span class="menu-toolbar__label">显示内嵌</span>
                <ElSwitch v-model="showIframeMenus" size="small" />
              </span>
              <span class="menu-toolbar__switch">
                <span class="menu-toolbar__label">显示启用</span>
                <ElSwitch v-model="showEnabledMenus" size="small" />
              </span>
              <span class="menu-toolbar__switch">
                <span class="menu-toolbar__label">显示路径</span>
                <ElSwitch v-model="showMenuPath" size="small" />
              </span>
            </div>

            <div class="menu-card">
              <ElCascaderPanel
                ref="menuPanelRef"
                v-model="selectedMenuPath"
                :options="filteredMenuOptions"
                :props="menuCascaderProps"
                class="permission-cascader permission-cascader--readonly"
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
                      {{ `${data.visibleLeafCount || 0}/${data.totalLeafCount || 0}` }}
                    </ElTag>
                  </div>
                </template>
              </ElCascaderPanel>
            </div>

            <div class="menu-footer">
              <span class="empty-text"
                >这里只展示当前用户在所选上下文下最终可见的菜单，不参与编辑。</span
              >
              <ElButton text @click="selectedMenuPath = []">清空浏览</ElButton>
            </div>
          </section>
        </ElTabPane>

        <ElTabPane :disabled="contextType !== 'collaboration'" label="角色链路" name="roles">
          <section class="role-panel">
            <div class="panel-title">角色链路</div>
            <ElTable :data="pagedRoleRows" border max-height="320">
              <ElTableColumn
                prop="roleCode"
                label="角色编码"
                min-width="140"
                show-overflow-tooltip
              />
              <ElTableColumn
                prop="roleName"
                label="角色名称"
                min-width="140"
                show-overflow-tooltip
              />
              <ElTableColumn label="继承协作空间" width="100">
                <template #default="{ row }">
                  <ElTag :type="row.inherited ? 'warning' : 'info'" effect="plain">
                    {{ row.inherited ? '是' : '否' }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn label="命中结果" width="110">
                <template #default="{ row }">
                  <ElTag
                    :type="row.matched ? 'success' : row.disabled ? 'danger' : 'info'"
                    effect="plain"
                  >
                    {{ row.matched ? '生效' : row.disabled ? '角色禁用' : '未命中' }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn label="快照时间" min-width="160">
                <template #default="{ row }">{{ row.refreshedAt || '-' }}</template>
              </ElTableColumn>
              <ElTableColumn label="来源功能包" min-width="240" show-overflow-tooltip>
                <template #default="{ row }">
                  <div class="inline-tags">
                    <ElTag
                      v-for="item in row.sourcePackages || []"
                      :key="item.id"
                      size="small"
                      effect="plain"
                      round
                    >
                      {{ item.name }}
                    </ElTag>
                    <span v-if="!(row.sourcePackages || []).length" class="empty-text">-</span>
                  </div>
                </template>
              </ElTableColumn>
            </ElTable>
            <WorkspacePagination
              v-if="roleRows.length > 0"
              v-model:current-page="rolePagination.current"
              v-model:page-size="rolePagination.size"
              :total="roleRows.length"
              compact
            />
          </section>
        </ElTabPane>
      </ElTabs>
    </div>

    <template #footer>
      <ElButton @click="visible = false">关闭</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  // 视图脚本：所有 reactive state、computed、watch、handler 集中在 useUserPermissionTestDrawer。
  // 这里只做：1) 引入子组件；2) 调用 composable；3) 把返回值拉到 setup 作用域供模板访问。
  import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import { useUserPermissionTestDrawer } from './use-user-permission-test-drawer'

  interface Props {
    modelValue: boolean
    userData?: Api.SystemManage.UserListItem
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
  }>()

  const {
    Search,
    visible,
    loading,
    testing,
    refreshing,
    contextType,
    activeTab,
    selectedWorkspaceId,
    permissionKey,
    diagnosisData,
    workspaceOptions,
    menuKeyword,
    showHiddenMenus,
    showIframeMenus,
    showEnabledMenus,
    showMenuPath,
    selectedMenuPath,
    rolePagination,
    menuPanelRef,
    menuCascaderProps,
    userTitle,
    selectedWorkspaceName,
    filteredMenuOptions,
    summaryItems,
    roleRows,
    pagedRoleRows,
    handleTest,
    handleRefresh,
    formatPermissionStatus,
    formatMemberStatus,
    formatRoleCode,
    getMemberStatusTagType,
    formatBoundaryState,
    getBoundaryStateTagType
  } = useUserPermissionTestDrawer(props, emit)
</script>

<style scoped lang="scss">
  .permission-shell {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .dialog-note {
    color: #4b5563;
    line-height: 1.7;
  }

  .toolbar-row {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
  }

  .permission-tabs {
    margin-top: 2px;
  }

  .toolbar-select {
    width: 160px;
  }

  .toolbar-select--collaborationWorkspace {
    width: 260px;
  }

  .toolbar-input {
    flex: 1;
    min-width: 280px;
  }

  .panel-grid {
    display: grid;
    grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
    gap: 14px;
  }

  .info-panel,
  .role-panel {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 14px;
    padding: 16px 18px;
    background: #fff;
  }

  .panel-title,
  .source-title {
    font-size: 14px;
    font-weight: 600;
    color: #111827;
    margin-bottom: 12px;
  }

  .kv-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px 16px;
  }

  .kv-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-height: 48px;
  }

  .kv-label {
    font-size: 12px;
    color: #6b7280;
  }

  .kv-value {
    color: #111827;
    line-height: 1.6;
    word-break: break-all;
  }

  .result-shell {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .result-banner {
    border-radius: 12px;
    padding: 12px 14px;
    border: 1px solid transparent;
  }

  .result-banner--success {
    background: #f0fdf4;
    border-color: #bbf7d0;
  }

  .result-banner--danger {
    background: #fef2f2;
    border-color: #fecaca;
  }

  .result-title {
    font-weight: 600;
    color: #111827;
    margin-bottom: 4px;
  }

  .result-text {
    color: #4b5563;
    line-height: 1.6;
  }

  .result-hint {
    margin-top: 8px;
    color: #92400e;
    font-size: 12px;
    line-height: 1.6;
  }

  .source-panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .menu-panel {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 14px;
    padding: 16px 18px;
    background: #fff;
  }

  .menu-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 12px;
  }

  .menu-toolbar__search {
    width: 260px;
  }

  .menu-toolbar__switch {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .menu-toolbar__label {
    color: #4b5563;
    font-size: 13px;
    white-space: nowrap;
  }

  .menu-card {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    padding: 8px;
    min-height: 360px;
  }

  .menu-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 12px;
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

  :deep(.permission-cascader) {
    width: 100%;
  }

  :deep(.permission-cascader .el-cascader-menu) {
    width: 33.333%;
    min-width: 280px;
  }

  :deep(.permission-cascader .el-cascader-menu__wrap) {
    height: 320px;
  }

  .source-tags,
  .inline-tags {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    align-items: center;
  }

  .empty-text {
    color: #9ca3af;
  }

  @media (max-width: 1100px) {
    .panel-grid {
      grid-template-columns: 1fr;
    }
  }
</style>

