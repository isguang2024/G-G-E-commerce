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
        这里用于诊断用户在平台或协作空间上下文中的功能权限生效情况。平台角色通过个人工作空间生效；协作空间内权限则由成员身份、协作空间边界和协作空间内部角色共同决定。
      </div>

      <div class="toolbar-row">
        <ElSelect v-model="contextType" class="toolbar-select">
          <ElOption label="平台上下文" value="platform" />
          <ElOption label="协作空间上下文" value="collaboration" />
        </ElSelect>

        <ElSelect
          v-if="contextType === 'collaboration'"
          v-model="selectedCollaborationWorkspaceId"
          filterable
          clearable
          class="toolbar-select toolbar-select--team"
          placeholder="请选择协作空间"
        >
          <ElOption
            v-for="team in teamOptions"
            :key="team.id"
            :label="team.name"
            :value="team.id"
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
                    contextType === 'platform' ? '平台' : selectedTeamName || '未选择协作空间'
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
                      contextType === 'platform'
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
                          :type="diagnosisData.teamMember?.matched ? 'success' : 'warning'"
                          effect="plain"
                        >
                          {{
                            diagnosisData.teamMember?.matched ? '已加入协作空间' : '未加入协作空间'
                          }}
                        </ElTag>
                      </span>
                    </div>
                    <div class="kv-item">
                      <span class="kv-label">成员状态</span>
                      <span class="kv-value">{{
                        formatMemberStatus(diagnosisData.teamMember?.status)
                      }}</span>
                    </div>
                    <div class="kv-item">
                      <span class="kv-label">成员身份</span>
                      <span class="kv-value">{{
                        formatRoleCode(diagnosisData.teamMember?.roleCode)
                      }}</span>
                    </div>
                  </div>
                </div>
                <div v-if="contextType === 'collaboration'" class="source-panel">
                  <div class="source-title">当前协作空间功能包</div>
                  <div class="source-tags">
                    <ElTag
                      v-for="item in diagnosisData.teamPackages || []"
                      :key="item.id"
                      effect="plain"
                      type="success"
                      round
                    >
                      {{ item.name }}
                    </ElTag>
                    <span v-if="!(diagnosisData.teamPackages || []).length" class="empty-text"
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
  import { computed, nextTick, ref, watch } from 'vue'
  import { Search } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import type { CascaderOption, CascaderProps } from 'element-plus'
  import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import {
    fetchGetUserPermissionDiagnosis,
    fetchGetUserPermissionMenus,
    fetchGetUserCollaborationWorkspaces,
    fetchRefreshUserPermissionSnapshot
  } from '@/api/system-manage'
  import { formatMenuTitle } from '@/utils/router'

  interface Props {
    modelValue: boolean
    userData?: Api.SystemManage.UserListItem
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const testing = ref(false)
  const refreshing = ref(false)
  const contextType = ref<'platform' | 'collaboration'>('platform')
  const activeTab = ref<'permission' | 'menus' | 'roles'>('permission')
  const selectedCollaborationWorkspaceId = ref('')
  const permissionKey = ref('')
  const diagnosisData = ref<Api.SystemManage.UserPermissionDiagnosisResponse>()
  const teamOptions = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
  const permissionMenus = ref<Api.SystemManage.UserPermissionMenuNode[]>([])
  const menuKeyword = ref('')
  const showHiddenMenus = ref(true)
  const showIframeMenus = ref(true)
  const showEnabledMenus = ref(true)
  const showMenuPath = ref(false)
  const selectedMenuPath = ref<string[]>([])
  const rolePagination = ref({
    current: 1,
    size: 10
  })
  const menuPanelRef = ref<any>()

  interface MenuOption extends CascaderOption {
    path?: string
    hidden?: boolean
    isIframe?: boolean
    isEnable?: boolean
    totalLeafCount?: number
    visibleLeafCount?: number
  }

  const menuCascaderProps: CascaderProps = {
    emitPath: true,
    checkStrictly: true,
    expandTrigger: 'click',
    showPrefix: true
  }

  const userTitle = computed(
    () => props.userData?.nickName || props.userData?.userName || props.userData?.id || ''
  )

  const selectedTeamName = computed(
    () =>
      teamOptions.value.find((item) => item.id === selectedCollaborationWorkspaceId.value)?.name ||
      ''
  )

  const summaryItems = computed(() => {
    const items: Array<{
      label: string
      value: string | number
      type?: 'success' | 'warning' | 'info' | 'primary' | 'danger'
    }> = []
    const snapshot = diagnosisData.value?.snapshot
    if (!snapshot) {
      return [
        {
          label: '快照',
          value: '未加载',
          type: 'info' as const
        }
      ]
    }
    items.push(
      {
        label: '功能包',
        value:
          contextType.value === 'platform'
            ? (snapshot.expandedPackageCount ?? 0)
            : (snapshot.expandedPackageCount ?? 0)
      },
      {
        label: contextType.value === 'platform' ? '权限数' : '协作空间生效权限',
        value:
          contextType.value === 'platform'
            ? (snapshot.actionCount ?? 0)
            : (snapshot.effectiveActionCount ?? 0),
        type: 'success'
      },
      {
        label: '菜单数',
        value: countVisibleMenuLeaves(filteredMenuOptions.value),
        type: 'primary'
      }
    )
    if (contextType.value === 'platform') {
      items.push({
        label: '已禁用',
        value: snapshot.disabledActionCount ?? 0,
        type: 'warning'
      })
    } else {
      items.push({
        label: '协作空间屏蔽',
        value: snapshot.blockedActionCount ?? 0,
        type: 'warning'
      })
    }
    items.push({
      label: '刷新时间',
      value: snapshot.refreshedAt || '-',
      type: 'info'
    })
    return items
  })
  const roleRows = computed(() => diagnosisData.value?.roles || [])
  const pagedRoleRows = computed(() => {
    const start = (rolePagination.value.current - 1) * rolePagination.value.size
    return roleRows.value.slice(start, start + rolePagination.value.size)
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (!open) return
      void initialize()
    }
  )

  watch(contextType, async (value) => {
    if (!visible.value) return
    if (value === 'platform') {
      selectedCollaborationWorkspaceId.value = ''
      if (activeTab.value === 'roles') activeTab.value = 'permission'
    }
    await loadDiagnosis()
  })

  watch(selectedCollaborationWorkspaceId, async () => {
    if (!visible.value || contextType.value !== 'collaboration') return
    await loadDiagnosis()
  })

  async function initialize() {
    activeTab.value = 'permission'
    await loadTeams()
    await loadDiagnosis()
  }

  async function loadTeams() {
    const userId = props.userData?.id
    if (!userId) {
      teamOptions.value = []
      return
    }
    try {
      teamOptions.value = await fetchGetUserCollaborationWorkspaces(userId)
      if (
        contextType.value === 'collaboration' &&
        selectedCollaborationWorkspaceId.value &&
        !teamOptions.value.some((item) => item.id === selectedCollaborationWorkspaceId.value)
      ) {
        selectedCollaborationWorkspaceId.value = ''
      }
      if (
        contextType.value === 'collaboration' &&
        !selectedCollaborationWorkspaceId.value &&
        teamOptions.value.length === 1
      ) {
        selectedCollaborationWorkspaceId.value = teamOptions.value[0].id
      }
    } catch {
      teamOptions.value = []
    }
  }

  async function loadDiagnosis() {
    const userId = props.userData?.id
    if (!userId) return
    if (contextType.value === 'collaboration' && !selectedCollaborationWorkspaceId.value) {
      diagnosisData.value = undefined
      permissionMenus.value = []
      return
    }
    loading.value = true
    try {
      const collaborationWorkspaceId =
        contextType.value === 'collaboration' ? selectedCollaborationWorkspaceId.value : undefined
      const [diagnosis, menus] = await Promise.all([
        fetchGetUserPermissionDiagnosis(userId, {
          collaborationWorkspaceId,
          permissionKey: permissionKey.value || undefined
        }),
        fetchGetUserPermissionMenus(userId, collaborationWorkspaceId)
      ])
      diagnosisData.value = diagnosis
      permissionMenus.value = menus
      rolePagination.value.current = 1
      await nextTick()
      ensureExpandedMenus(menuPanelRef.value, selectedMenuPath.value)
    } catch (error: any) {
      diagnosisData.value = undefined
      permissionMenus.value = []
      ElMessage.error(error?.message || '加载权限诊断失败')
    } finally {
      loading.value = false
    }
  }

  async function handleTest() {
    if (!permissionKey.value.trim()) {
      ElMessage.warning('请输入权限键')
      return
    }
    if (contextType.value === 'collaboration' && !selectedCollaborationWorkspaceId.value) {
      ElMessage.warning('请选择协作空间')
      return
    }
    testing.value = true
    try {
      const collaborationWorkspaceId =
        contextType.value === 'collaboration' ? selectedCollaborationWorkspaceId.value : undefined
      const [diagnosis, menus] = await Promise.all([
        fetchGetUserPermissionDiagnosis(props.userData?.id || '', {
          collaborationWorkspaceId,
          permissionKey: permissionKey.value.trim()
        }),
        fetchGetUserPermissionMenus(props.userData?.id || '', collaborationWorkspaceId)
      ])
      diagnosisData.value = diagnosis
      permissionMenus.value = menus
      rolePagination.value.current = 1
    } catch (error: any) {
      ElMessage.error(error?.message || '权限测试失败')
    } finally {
      testing.value = false
    }
  }

  async function handleRefresh() {
    const userId = props.userData?.id
    if (!userId) return
    if (contextType.value === 'collaboration' && !selectedCollaborationWorkspaceId.value) {
      ElMessage.warning('请选择协作空间')
      return
    }
    refreshing.value = true
    try {
      const collaborationWorkspaceId =
        contextType.value === 'collaboration' ? selectedCollaborationWorkspaceId.value : undefined
      const [diagnosis, menus] = await Promise.all([
        fetchRefreshUserPermissionSnapshot(userId, collaborationWorkspaceId),
        fetchGetUserPermissionMenus(userId, collaborationWorkspaceId)
      ])
      diagnosisData.value = diagnosis
      permissionMenus.value = menus
      rolePagination.value.current = 1
      ElMessage.success('权限快照已刷新')
    } catch (error: any) {
      ElMessage.error(error?.message || '刷新权限快照失败')
    } finally {
      refreshing.value = false
    }
  }

  function formatPermissionStatus(status?: string) {
    if (!status) return '-'
    return status === 'normal' || status === 'active' ? '正常' : '停用'
  }

  function formatMemberStatus(status?: string) {
    switch (status) {
      case 'active':
        return '有效成员'
      case 'inactive':
        return '成员停用'
      case 'missing':
        return '未加入协作空间'
      default:
        return '-'
    }
  }

  function formatRoleCode(roleCode?: string) {
    switch (roleCode) {
      case 'collaboration_workspace_admin':
        return '协作空间管理员'
      case 'collaboration_workspace_member':
        return '协作空间成员'
      default:
        return roleCode || '-'
    }
  }

  function getMemberStatusTagType(status?: string) {
    switch (status) {
      case 'active':
        return 'success'
      case 'inactive':
        return 'danger'
      case 'missing':
        return 'warning'
      default:
        return 'info'
    }
  }

  function formatBoundaryState(state?: string) {
    switch (state) {
      case '命中':
        return '命中'
      case '拦截':
        return '拦截'
      case '未配置':
        return '未配置'
      case '未命中':
        return '未命中'
      case '超级管理员直通':
        return '超级管理员直通'
      default:
        return '-'
    }
  }

  function getBoundaryStateTagType(state?: string) {
    switch (state) {
      case '命中':
        return 'success'
      case '拦截':
        return 'danger'
      case '未配置':
        return 'warning'
      case '超级管理员直通':
        return 'warning'
      default:
        return 'info'
    }
  }

  const menuOptions = computed<MenuOption[]>(() =>
    normalizePermissionMenuOptions(permissionMenus.value)
  )

  const filteredMenuOptions = computed(() => {
    const keyword = menuKeyword.value.trim().toLowerCase()
    return filterNestedOptions(menuOptions.value, (node) => {
      if (!node.leaf) return !keyword
      if (!showHiddenMenus.value && node.hidden) return false
      if (!showIframeMenus.value && node.isIframe) return false
      if (!showEnabledMenus.value && node.isEnable !== false) return false
      if (keyword && !`${node.label || ''} ${node.path || ''}`.toLowerCase().includes(keyword))
        return false
      return true
    })
  })

  watch(filteredMenuOptions, async () => {
    await nextTick()
    ensureExpandedMenus(menuPanelRef.value, selectedMenuPath.value)
  })

  function normalizePermissionMenuOptions(
    items: Api.SystemManage.UserPermissionMenuNode[]
  ): MenuOption[] {
    return items.map((item) => {
      const children = normalizePermissionMenuOptions(item.children || [])
      return {
        value: item.id,
        label: formatMenuTitle(item.title || item.name || ''),
        path: item.path || '',
        hidden: Boolean(item.hidden),
        isIframe: Boolean(item.path && /^https?:\/\//.test(item.path)),
        isEnable: true,
        leaf: !(item.children || []).length,
        totalLeafCount: countPermissionMenuLeaves(item),
        visibleLeafCount: countVisibleMenuLeaves(children),
        children
      }
    })
  }

  function countPermissionMenuLeaves(node: Api.SystemManage.UserPermissionMenuNode): number {
    if (!(node.children || []).length) return 1
    return (node.children || []).reduce((sum, child) => sum + countPermissionMenuLeaves(child), 0)
  }

  function countVisibleMenuLeaves(items: MenuOption[]): number {
    return items.reduce((sum, item) => {
      if (!(item.children || []).length) return sum + 1
      return sum + countVisibleMenuLeaves((item.children || []) as MenuOption[])
    }, 0)
  }

  function filterNestedOptions<T extends CascaderOption>(
    items: T[],
    predicate: (node: T) => boolean
  ): T[] {
    return items
      .map((item) => {
        const children = filterNestedOptions(((item.children || []) as T[]) || [], predicate)
        const passed = predicate(item)
        if (!passed && !children.length) return null
        return {
          ...item,
          children
        } as T
      })
      .filter((item): item is T => Boolean(item))
  }

  function ensureExpandedMenus(panel: any, selectedValues: string[]) {
    const rootMenus = panel?.menus?.[0]
    if (!panel || !rootMenus?.length) return
    const firstValue = selectedValues?.[selectedValues.length - 1]
    let rootNode = rootMenus[0]
    let childNode = rootNode?.children?.[0]
    if (firstValue) {
      const matchedNode = panel
        .getFlattedNodes?.(false)
        ?.find((node: any) => `${node?.value}` === `${firstValue}`)
      const pathNodes = matchedNode?.pathNodes || []
      if (pathNodes[0]) rootNode = pathNodes[0]
      if (pathNodes[1]) childNode = pathNodes[1]
    }
    const nextMenus = [rootMenus]
    if (rootNode?.children?.length) nextMenus.push(rootNode.children)
    if (childNode?.children?.length) nextMenus.push(childNode.children)
    panel.menus = nextMenus
  }

  watch(
    () => rolePagination.value.size,
    () => {
      rolePagination.value.current = 1
    }
  )
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

  .toolbar-select--team {
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
