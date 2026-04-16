<template>
  <div class="access-trace-page">
    <!-- 头部说明 -->
    <ElCard shadow="never" class="trace-intro-card">
      <div class="intro-wrap">
        <div class="intro-main">
          <div class="intro-title">
            <span class="intro-dot"></span>
            访问链路测试
          </div>
          <div class="intro-desc">
            按
            <b>用户 → 角色/协作空间 → 菜单可见 → 页面可见性</b>
            的链路模拟运行时权限评估，用于排查菜单/页面不可见、权限命中异常等问题。
          </div>
        </div>
        <div class="intro-tips">
          <ElTag size="small" type="info" effect="plain">只读测试</ElTag>
          <ElTag size="small" type="info" effect="plain">不触发登录</ElTag>
        </div>
      </div>
    </ElCard>

    <!-- 筛选条件 -->
    <ElCard shadow="never" class="trace-filter-card">
      <template #header>
        <div class="card-header">
          <span class="card-header-title">筛选条件</span>
          <div class="card-header-actions">
            <ElButton plain :disabled="loading" @click="handleReset">重置</ElButton>
            <ElButton type="primary" :loading="loading" @click="handleQuery"> 测试链路 </ElButton>
          </div>
        </div>
      </template>

      <ElForm class="trace-form" label-position="top">
        <ElFormItem label="App">
          <AppKeySelect
            v-model="selectedAppKey"
            placeholder="选择 App"
            class="trace-field"
            @change="handleManagedAppChange"
          />
        </ElFormItem>

        <ElFormItem label="菜单空间">
          <ElSelect
            v-model="query.spaceKey"
            filterable
            clearable
            placeholder="默认空间"
            class="trace-field"
          >
            <ElOption label="默认空间" value="" />
            <ElOption
              v-for="item in spaceOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="协作空间">
          <ElSelect
            v-model="query.collaborationWorkspaceId"
            filterable
            clearable
            placeholder="个人空间(默认)"
            class="trace-field"
          >
            <ElOption
              v-for="cw in collaborationWorkspaceOptions"
              :key="cw.id"
              :label="cw.name"
              :value="cw.id"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="角色筛选">
          <ElSelect v-model="roleCodeFilter" clearable placeholder="全部角色" class="trace-field">
            <ElOption
              v-for="role in displayRoleOptions"
              :key="role.value"
              :label="role.label"
              :value="role.value"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="用户">
          <ElSelect
            v-model="query.userId"
            filterable
            clearable
            placeholder="请选择用户"
            class="trace-field"
          >
            <ElOption
              v-for="user in userOptions"
              :key="user.id"
              :label="formatUserLabel(user)"
              :value="user.id"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="页面Key">
          <ElSelect
            v-model="query.pageKey"
            filterable
            clearable
            placeholder="不指定则返回全部可见页面"
            class="trace-field"
          >
            <ElOption
              v-for="page in pageOptions"
              :key="page.pageKey"
              :label="`${page.pageKey} (${page.name})`"
              :value="page.pageKey"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="仅协作空间成员" class="trace-switch-item">
          <ElSwitch
            v-model="onlyCollaborationWorkspaceUsers"
            :disabled="!query.collaborationWorkspaceId"
          />
          <span class="trace-switch-hint">
            {{
              query.collaborationWorkspaceId
                ? '仅从当前协作空间成员中选择用户'
                : '先选择协作空间后生效'
            }}
          </span>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 空状态 -->
    <ElCard v-if="!result" shadow="never" class="trace-empty-card">
      <ElEmpty description="选择 App 与用户后，点击「测试链路」查看结果" />
    </ElCard>

    <!-- 结果区 -->
    <template v-if="result">
      <!-- 观测点：结果总览元数据 -->
      <div
        data-testid="trace-summary"
        :data-authenticated="result.authenticated ? '1' : '0'"
        :data-super-admin="result.superAdmin ? '1' : '0'"
        :data-action-key-count="result.actionKeyCount ?? 0"
        :data-visible-menu-count="visibleMenuCount"
        :data-visible-page-count="visiblePageCount"
        :data-total-page-count="pageRows.length"
        :data-space-key="result.spaceKey || ''"
        :data-collaboration-workspace-id="result.collaborationWorkspaceId || ''"
        :data-user-id="result.userId || ''"
        hidden
      ></div>

      <!-- 概览统计卡 -->
      <div class="trace-stats">
        <div class="stat-card">
          <div class="stat-label">登录态</div>
          <div class="stat-value">
            <ElTag
              :type="result.authenticated ? 'success' : 'danger'"
              size="large"
              effect="light"
              round
            >
              {{ result.authenticated ? '已认证' : '未认证' }}
            </ElTag>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-label">超级管理员</div>
          <div class="stat-value">
            <ElTag :type="result.superAdmin ? 'warning' : 'info'" size="large" effect="light" round>
              {{ result.superAdmin ? '是' : '否' }}
            </ElTag>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-label">动作权限数</div>
          <div class="stat-value stat-number">
            {{ result.superAdmin ? '不限' : result.actionKeyCount }}
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-label">可见菜单数</div>
          <div class="stat-value stat-number">{{ result.visibleMenuIds?.length || 0 }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">可见页面</div>
          <div class="stat-value stat-number">
            {{ visiblePageCount }}<span class="stat-sub">/ {{ pageRows.length }}</span>
          </div>
        </div>
        <div class="stat-card stat-card-wide">
          <div class="stat-label">菜单空间 · 协作空间</div>
          <div class="stat-value stat-meta">
            <ElTag size="small" effect="plain">{{ result.spaceKey || '-' }}</ElTag>
            <ElTag size="small" effect="plain" type="info">
              {{ result.collaborationWorkspaceId ? '协作空间' : '个人空间' }}
            </ElTag>
          </div>
        </div>
      </div>

      <!-- 身份信息小条 -->
      <div class="trace-identity">
        <div class="identity-item">
          <span class="identity-key">用户 ID</span>
          <span class="identity-value mono">{{ result.userId || '-' }}</span>
        </div>
        <div class="identity-item">
          <span class="identity-key">协作空间 ID</span>
          <span class="identity-value mono">{{ result.collaborationWorkspaceId || '-' }}</span>
        </div>
      </div>

      <!-- 角色链路 -->
      <ElCard shadow="never" class="trace-result-card">
        <template #header>
          <div class="card-header">
            <span class="card-header-title">
              角色链路
              <ElTag size="small" class="count-badge" round>{{ roleRows.length }}</ElTag>
            </span>
            <span class="card-header-hint">
              {{
                result.collaborationWorkspaceId ? '来自协作空间有效启用角色' : '来自账号级个人角色'
              }}
            </span>
          </div>
        </template>

        <ElTable
          v-if="roleRows.length > 0"
          :data="pagedRoles"
          stripe
          :border="false"
          class="trace-table"
        >
          <ElTableColumn prop="roleCode" label="角色编码" min-width="180">
            <template #default="{ row }">
              <span
                class="mono"
                data-testid="trace-node"
                data-node-type="role"
                :data-role-code="row.roleCode || ''"
                :data-role-id="row.roleId || ''"
                :data-status="row.status || ''"
                >{{ row.roleCode || '-' }}</span
              >
            </template>
          </ElTableColumn>
          <ElTableColumn prop="roleName" label="角色名称" min-width="180" />
          <ElTableColumn prop="status" label="状态" width="120">
            <template #default="{ row }">
              <ElTag
                :type="row.status === 'normal' ? 'success' : 'info'"
                effect="plain"
                size="small"
                :data-testid="'trace-node-status'"
                :data-status="row.status || ''"
              >
                {{ row.status || '-' }}
              </ElTag>
            </template>
          </ElTableColumn>
        </ElTable>
        <ElEmpty v-else description="当前用户在该上下文下未匹配任何角色" :image-size="80" />
        <WorkspacePagination
          v-if="roleRows.length > 0"
          v-model:current-page="rolePagination.current"
          v-model:page-size="rolePagination.size"
          :total="roleRows.length"
          compact
        />
      </ElCard>

      <!-- 菜单链路 -->
      <ElCard shadow="never" class="trace-result-card">
        <template #header>
          <div class="card-header">
            <span class="card-header-title">
              菜单链路
              <ElTag size="small" class="count-badge" round>
                {{ visibleMenuCount }} / {{ menuRows.length }}
              </ElTag>
            </span>
            <div class="card-header-actions">
              <ElInput
                v-model="menuKeyword"
                placeholder="搜索菜单名 / 路径"
                clearable
                size="default"
                class="page-search"
              >
                <template #prefix>
                  <ElIcon><Search /></ElIcon>
                </template>
              </ElInput>
              <ElRadioGroup v-model="menuVisibleFilter" size="default">
                <ElRadioButton value="all">全部</ElRadioButton>
                <ElRadioButton value="visible">可见</ElRadioButton>
                <ElRadioButton value="hidden">不可见</ElRadioButton>
              </ElRadioGroup>
              <ElButton size="default" plain @click="handleToggleMenuExpand">
                {{ menuExpandAll ? '全部折叠' : '全部展开' }}
              </ElButton>
            </div>
          </div>
        </template>

        <ElTree
          v-if="menuTree.length > 0"
          ref="menuTreeRef"
          :data="menuTree"
          node-key="id"
          :default-expand-all="menuExpandAll"
          :expand-on-click-node="false"
          :filter-node-method="menuFilterMethod"
          class="menu-tree"
        >
          <template #default="{ node, data }">
            <div
              class="menu-node"
              :class="{ 'menu-node-hidden': !data.visible }"
              data-testid="trace-node"
              data-node-type="menu"
              :data-menu-id="data.id"
              :data-parent-id="data.parentId || ''"
              :data-visible="data.visible ? '1' : '0'"
              :data-status="data.visible ? 'visible' : 'denied'"
              :data-kind="data.kind || ''"
              :data-full-path="data.fullPath || ''"
              :data-hidden="data.hidden ? '1' : '0'"
            >
              <ElTag
                size="small"
                :type="data.visible ? 'success' : 'danger'"
                effect="light"
                round
                class="menu-node-tag"
              >
                {{ data.visible ? '可见' : '拒绝' }}
              </ElTag>
              <span class="menu-node-name">{{ resolveMenuLabel(data) }}</span>
              <ElTag v-if="data.kind" size="small" effect="plain" class="menu-node-kind">
                {{ data.kind }}
              </ElTag>
              <span v-if="data.fullPath" class="menu-node-path mono">{{ data.fullPath }}</span>
              <ElTag v-if="data.hidden" size="small" effect="plain" type="info"> hidden </ElTag>
            </div>
          </template>
        </ElTree>
        <ElEmpty v-else description="当前空间下没有菜单" :image-size="80" />
      </ElCard>

      <!-- 页面链路 -->
      <ElCard shadow="never" class="trace-result-card">
        <template #header>
          <div class="card-header">
            <span class="card-header-title">
              页面链路结果
              <ElTag size="small" class="count-badge" round>{{ filteredPageRows.length }}</ElTag>
            </span>
            <div class="card-header-actions">
              <ElInput
                v-model="pageKeyword"
                placeholder="搜索 PageKey / 名称 / 路由"
                clearable
                size="default"
                class="page-search"
              >
                <template #prefix>
                  <ElIcon><Search /></ElIcon>
                </template>
              </ElInput>
              <ElRadioGroup v-model="pageVisibleFilter" size="default">
                <ElRadioButton value="all">全部</ElRadioButton>
                <ElRadioButton value="visible">可见</ElRadioButton>
                <ElRadioButton value="hidden">不可见</ElRadioButton>
              </ElRadioGroup>
            </div>
          </div>
        </template>

        <ElTable
          v-if="filteredPageRows.length > 0"
          :data="pagedPages"
          stripe
          :border="false"
          class="trace-table"
          row-key="pageKey"
        >
          <ElTableColumn label="可见" width="84" align="center">
            <template #default="{ row }">
              <ElTag :type="row.visible ? 'success' : 'danger'" effect="light" size="small" round>
                {{ row.visible ? '可见' : '拒绝' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="页面" min-width="260">
            <template #default="{ row }">
              <div
                class="page-cell"
                data-testid="trace-node"
                data-node-type="page"
                :data-page-key="row.pageKey || ''"
                :data-page-name="row.pageName || ''"
                :data-visible="row.visible ? '1' : '0'"
                :data-status="row.visible ? 'visible' : 'denied'"
                :data-reason="row.reason || ''"
                :data-access-mode="row.accessMode || ''"
                :data-permission-key="row.permissionKey || ''"
                :data-matched-action-key="row.matchedActionKey || ''"
                :data-route-path="row.routePath || ''"
              >
                <div class="page-cell-name">{{ row.pageName || '-' }}</div>
                <div class="page-cell-key mono">{{ row.pageKey }}</div>
              </div>
            </template>
          </ElTableColumn>
          <ElTableColumn label="路由" min-width="180">
            <template #default="{ row }">
              <span class="mono">{{ row.routePath || '-' }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="访问模式" width="120" align="center">
            <template #default="{ row }">
              <ElTag size="small" effect="plain" :type="accessModeTagType(row.accessMode)">
                {{ row.accessMode || '-' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="权限键" min-width="180">
            <template #default="{ row }">
              <span class="mono text-muted">{{ row.permissionKey || '—' }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="判定" min-width="160">
            <template #default="{ row }">
              <ElTag
                size="small"
                effect="plain"
                :type="row.visible ? 'success' : 'danger'"
                :data-testid="row.visible ? 'trace-node-reason' : 'trace-node-error'"
                :data-reason-code="row.reason || ''"
                :data-visible="row.visible ? '1' : '0'"
              >
                {{ formatReason(row.reason) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="命中动作" min-width="180">
            <template #default="{ row }">
              <span v-if="row.matchedActionKey" class="mono">{{ row.matchedActionKey }}</span>
              <span v-else-if="result?.superAdmin" class="text-muted">超管旁路</span>
              <span v-else class="text-muted">—</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="链路" min-width="240">
            <template #default="{ row }">
              <div class="chain-cell">
                <template v-for="(item, idx) in row.effectiveChain || []" :key="idx">
                  <span class="chain-seg">{{ item }}</span>
                  <ElIcon v-if="idx < (row.effectiveChain || []).length - 1" class="chain-arrow">
                    <ArrowRight />
                  </ElIcon>
                </template>
                <span v-if="!(row.effectiveChain || []).length" class="text-muted">-</span>
              </div>
            </template>
          </ElTableColumn>
        </ElTable>
        <ElEmpty v-else description="没有符合筛选条件的页面" :image-size="80" />
        <WorkspacePagination
          v-if="filteredPageRows.length > 0"
          v-model:current-page="pagePagination.current"
          v-model:page-size="pagePagination.size"
          :total="filteredPageRows.length"
          compact
        />
      </ElCard>
    </template>
  </div>
</template>

<script setup lang="ts">
  import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, type ElTree as ElTreeType } from 'element-plus'
  import { ArrowRight, Search } from '@element-plus/icons-vue'
  import AppKeySelect from '@/components/business/app/AppKeySelect.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import { useManagedAppScope } from '@/domains/app-runtime/useManagedAppScope'
  import {
    fetchGetCollaborationWorkspaceMembers,
    fetchGetCollaborationWorkspaceRoles
  } from '@/api/collaboration-workspace'
  import {
    fetchGetMenuSpaces,
    fetchGetPageAccessTrace,
    fetchGetPageList,
    fetchGetRoleOptions,
    fetchGetCollaborationWorkspaceOptions,
    fetchGetUserList
  } from '@/domains/governance/api'

  defineOptions({ name: 'SystemAccessTrace' })

  const { t, te } = useI18n()
  function resolveMenuLabel(data: { title?: string; name?: string }) {
    const title = (data.title || '').trim()
    if (title) return title
    const name = (data.name || '').trim()
    if (!name) return '-'
    // name 可能是 i18n key（如 menus.dashboard.title）
    if (te(name)) {
      const translated = t(name)
      if (translated && translated !== name) return translated
    }
    return name
  }

  const loading = ref(false)
  const result = ref<Api.SystemManage.PageAccessTraceResult | null>(null)
  const userOptions = ref<Api.SystemManage.UserListItem[]>([])
  const pageOptions = ref<Api.SystemManage.PageItem[]>([])
  const menuSpaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const collaborationWorkspaceOptions = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
  const roleOptions = ref<
    Array<{ label: string; value: string; source: 'personal' | 'collaboration' }>
  >([])
  const selectedAppKey = ref('')
  const rolePagination = reactive({
    current: 1,
    size: 10
  })
  const pagePagination = reactive({
    current: 1,
    size: 10
  })
  const onlyCollaborationWorkspaceUsers = ref(false)
  const roleCodeFilter = ref('')
  const displayRoleOptions = computed(() =>
    query.collaborationWorkspaceId
      ? roleOptions.value.filter((item) => item.source === 'collaboration')
      : roleOptions.value.filter((item) => item.source === 'personal')
  )
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const spaceOptions = computed(() =>
    menuSpaces.value.map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )
  const roleRows = computed(() => result.value?.roles || [])
  const pageRows = computed(() => result.value?.pages || [])
  const menuRows = computed(() => result.value?.menus || [])
  const visiblePageCount = computed(() => pageRows.value.filter((p) => p.visible).length)
  const visibleMenuCount = computed(() => menuRows.value.filter((m) => m.visible).length)

  // 菜单树
  type MenuNode = Api.SystemManage.PageAccessTraceMenuItem & { children: MenuNode[] }
  type MenuTreeNodeState = { expanded?: boolean }
  type MenuTreeStore = { nodesMap?: Record<string, MenuTreeNodeState> }
  type MenuTreeExposed = InstanceType<typeof ElTreeType> & {
    store?: MenuTreeStore
    filter?: (keyword: string) => void
  }
  const menuKeyword = ref('')
  const menuVisibleFilter = ref<'all' | 'visible' | 'hidden'>('all')
  const menuExpandAll = ref(true)
  const menuTreeRef = ref<InstanceType<typeof ElTreeType> | null>(null)
  const menuTree = computed<MenuNode[]>(() => {
    const nodes = menuRows.value.map<MenuNode>((item) => ({ ...item, children: [] }))
    const map = new Map<string, MenuNode>()
    nodes.forEach((n) => map.set(n.id, n))
    const roots: MenuNode[] = []
    nodes.forEach((n) => {
      if (n.parentId && map.has(n.parentId)) {
        map.get(n.parentId)!.children.push(n)
      } else {
        roots.push(n)
      }
    })
    const sortRec = (list: MenuNode[]) => {
      list.sort((a, b) => a.sortOrder - b.sortOrder || a.name.localeCompare(b.name))
      list.forEach((item) => sortRec(item.children))
    }
    sortRec(roots)
    return roots
  })
  function menuFilterMethod(_value: string, rawData: Record<string, any>) {
    const data = rawData as MenuNode
    const kw = menuKeyword.value.trim().toLowerCase()
    if (menuVisibleFilter.value === 'visible' && !data.visible) return false
    if (menuVisibleFilter.value === 'hidden' && data.visible) return false
    if (!kw) return true
    return (
      (data.name || '').toLowerCase().includes(kw) ||
      (data.title || '').toLowerCase().includes(kw) ||
      resolveMenuLabel(data).toLowerCase().includes(kw) ||
      (data.path || '').toLowerCase().includes(kw) ||
      (data.fullPath || '').toLowerCase().includes(kw)
    )
  }
  function handleToggleMenuExpand() {
    menuExpandAll.value = !menuExpandAll.value
    nextTick(() => {
      const tree = menuTreeRef.value as MenuTreeExposed | null
      if (!tree) return
      const store = tree.store
      if (!store) return
      Object.values(store.nodesMap || {}).forEach((node) => {
        node.expanded = menuExpandAll.value
      })
    })
  }
  watch([menuKeyword, menuVisibleFilter], () => {
    nextTick(() => {
      const tree = menuTreeRef.value as MenuTreeExposed | null
      tree?.filter?.(menuKeyword.value)
    })
  })
  const pageKeyword = ref('')
  const pageVisibleFilter = ref<'all' | 'visible' | 'hidden'>('all')
  const filteredPageRows = computed(() => {
    const kw = pageKeyword.value.trim().toLowerCase()
    return pageRows.value.filter((row) => {
      if (pageVisibleFilter.value === 'visible' && !row.visible) return false
      if (pageVisibleFilter.value === 'hidden' && row.visible) return false
      if (!kw) return true
      return (
        (row.pageKey || '').toLowerCase().includes(kw) ||
        (row.pageName || '').toLowerCase().includes(kw) ||
        (row.routePath || '').toLowerCase().includes(kw)
      )
    })
  })
  const pagedRoles = computed(() => {
    const start = (rolePagination.current - 1) * rolePagination.size
    return roleRows.value.slice(start, start + rolePagination.size)
  })
  const pagedPages = computed(() => {
    const start = (pagePagination.current - 1) * pagePagination.size
    return filteredPageRows.value.slice(start, start + pagePagination.size)
  })

  const REASON_TEXT: Record<string, string> = {
    visible_in_runtime: '运行时可见',
    denied_or_not_in_runtime: '不在运行时可见集'
  }
  function formatReason(reason?: string) {
    if (!reason) return '-'
    return REASON_TEXT[reason] || reason
  }
  function accessModeTagType(mode?: string) {
    switch ((mode || '').toLowerCase()) {
      case 'public':
        return 'success'
      case 'jwt':
      case 'authenticated':
        return 'info'
      case 'permission':
        return 'warning'
      default:
        return 'info'
    }
  }
  function handleReset() {
    query.userId = ''
    query.collaborationWorkspaceId = ''
    query.pageKey = ''
    query.spaceKey = ''
    roleCodeFilter.value = ''
    onlyCollaborationWorkspaceUsers.value = false
    pageKeyword.value = ''
    pageVisibleFilter.value = 'all'
    result.value = null
  }

  const query = reactive<Api.SystemManage.PageAccessTraceParams>({
    userId: '',
    collaborationWorkspaceId: '',
    pageKey: '',
    spaceKey: ''
  })

  function formatUserLabel(user: Api.SystemManage.UserListItem) {
    const userName = `${user.userName || ''}`.trim()
    const nickName = `${user.nickName || ''}`.trim()
    if (userName && nickName && userName !== nickName) {
      return `${userName}（${nickName}）`
    }
    return userName || nickName || user.id
  }

  async function loadUserOptions() {
    if (!targetAppKey.value) {
      userOptions.value = []
      return
    }
    const useCollaborationWorkspaceMembers =
      Boolean(query.collaborationWorkspaceId) &&
      (onlyCollaborationWorkspaceUsers.value || Boolean(roleCodeFilter.value))
    if (useCollaborationWorkspaceMembers && query.collaborationWorkspaceId) {
      const collaborationWorkspaceMembers = await fetchGetCollaborationWorkspaceMembers(
        query.collaborationWorkspaceId,
        {
          role_code: roleCodeFilter.value || undefined
        }
      )
      userOptions.value = (collaborationWorkspaceMembers || []).map((item: any) => ({
        id: item.userId,
        userName: item.userName,
        nickName: item.nickName,
        userPhone: '',
        userEmail: item.userEmail || '',
        avatar: item.avatar || '',
        status: item.status || 'active',
        roleIDs: [],
        roleNames: [],
        roleDetails: [],
        userRoles: [],
        registerSource: '',
        invitedBy: '',
        invitedByName: '',
        createTime: '',
        updateTime: ''
      }))
    } else {
      // 个人空间场景下 roleCodeFilter 存的是 roleId（见 loadRoleOptions），
      // 直接作为 roleId 过滤；协作空间场景已由上面的分支处理。
      const users = await fetchGetUserList({
        current: 1,
        size: 200,
        roleId: roleCodeFilter.value || ''
      })
      userOptions.value = users.records || []
    }

    if (query.userId && !userOptions.value.some((item) => item.id === query.userId)) {
      query.userId = ''
    }
  }

  async function loadRoleOptions() {
    if (!targetAppKey.value) {
      roleOptions.value = []
      return
    }
    if (query.collaborationWorkspaceId) {
      const collaborationWorkspaceRoles = await fetchGetCollaborationWorkspaceRoles(
        query.collaborationWorkspaceId
      )
      roleOptions.value = (collaborationWorkspaceRoles || [])
        .filter((role: any) => {
          const code = `${role.roleCode || ''}`.trim()
          if (!code || code === 'admin') return false
          if (code === 'collaboration_workspace_admin' || code === 'collaboration_workspace_member')
            return true
          return Boolean(role.collaborationWorkspaceId)
        })
        .map((role: any) => ({
          label: role.roleName || role.roleCode,
          value: role.roleCode,
          source: 'collaboration' as const
        }))
      return
    }

    const roleRes = await fetchGetRoleOptions()
    roleOptions.value = (roleRes.records || []).map((role: any) => ({
      label: role.roleName || role.roleCode,
      value: role.roleId,
      source: 'personal' as const
    }))
  }

  async function loadOptions() {
    if (!targetAppKey.value) {
      result.value = null
      userOptions.value = []
      pageOptions.value = []
      menuSpaces.value = []
      collaborationWorkspaceOptions.value = []
      roleOptions.value = []
      return
    }
    const [pages, collaborationWorkspacePage, spaces] = await Promise.all([
      fetchGetPageList({ current: 1, size: 500, appKey: targetAppKey.value }),
      fetchGetCollaborationWorkspaceOptions({ current: 1, size: 200 }),
      fetchGetMenuSpaces(targetAppKey.value)
    ])
    pageOptions.value = pages.records || []
    collaborationWorkspaceOptions.value = collaborationWorkspacePage.records || []
    menuSpaces.value = spaces.records || []
    await loadRoleOptions()
    await loadUserOptions()
  }

  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
    query.pageKey = ''
    query.spaceKey = ''
    query.userId = ''
    query.collaborationWorkspaceId = ''
    roleCodeFilter.value = ''
    onlyCollaborationWorkspaceUsers.value = false
    result.value = null
  }

  async function handleQuery() {
    if (!targetAppKey.value) {
      ElMessage.warning('请先选择 App')
      return
    }
    if (!query.userId) {
      ElMessage.warning('请先选择用户')
      return
    }
    loading.value = true
    try {
      result.value = await fetchGetPageAccessTrace({
        ...query,
        appKey: targetAppKey.value
      })
      rolePagination.current = 1
      pagePagination.current = 1
    } catch (err: any) {
      ElMessage.error(err?.message || '访问链路测试失败')
      result.value = null
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
    loadOptions().catch(() => {
      ElMessage.error('初始化测试数据失败')
    })
  })

  watch(
    () => targetAppKey.value,
    async () => {
      selectedAppKey.value = targetAppKey.value || ''
      result.value = null
      await loadOptions()
    }
  )

  watch(
    () => query.collaborationWorkspaceId,
    async (collaborationWorkspaceId, previousCollaborationWorkspaceId) => {
      if (collaborationWorkspaceId !== previousCollaborationWorkspaceId) {
        roleCodeFilter.value = ''
      }
      if (!collaborationWorkspaceId && onlyCollaborationWorkspaceUsers.value) {
        onlyCollaborationWorkspaceUsers.value = false
      }
      await loadRoleOptions()
      await loadUserOptions()
    }
  )

  watch(
    () => [onlyCollaborationWorkspaceUsers.value, roleCodeFilter.value],
    async () => {
      await loadUserOptions()
    }
  )

  watch(
    () => rolePagination.size,
    () => {
      rolePagination.current = 1
    }
  )

  watch(
    () => pagePagination.size,
    () => {
      pagePagination.current = 1
    }
  )

  watch([pageKeyword, pageVisibleFilter], () => {
    pagePagination.current = 1
  })
</script>

<style scoped>
  .access-trace-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 4px;
    min-height: 100%;
    overflow: visible;
  }

  /* 顶部说明卡 */
  .trace-intro-card :deep(.el-card__body) {
    padding: 18px 20px;
  }
  .intro-wrap {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
  }
  .intro-main {
    flex: 1;
    min-width: 280px;
  }
  .intro-title {
    font-size: 16px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--el-text-color-primary);
  }
  .intro-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--el-color-primary);
    box-shadow: 0 0 0 4px var(--el-color-primary-light-8);
  }
  .intro-desc {
    margin-top: 6px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--el-text-color-secondary);
  }
  .intro-desc b {
    color: var(--el-text-color-primary);
    font-weight: 500;
  }
  .intro-tips {
    display: flex;
    gap: 6px;
    flex-shrink: 0;
  }

  /* 卡片通用 header */
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }
  .card-header-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .count-badge {
    font-variant-numeric: tabular-nums;
  }
  .card-header-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
  .card-header-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  /* 筛选卡 */
  .trace-filter-card :deep(.el-card__body) {
    padding: 18px 20px 8px;
  }
  .trace-form {
    display: grid;
    gap: 4px 16px;
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
  .trace-field {
    width: 100%;
  }
  .trace-form :deep(.el-form-item) {
    margin-bottom: 14px;
  }
  .trace-form :deep(.el-form-item__label) {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    padding-bottom: 4px !important;
    line-height: 1.4;
  }
  .trace-switch-item :deep(.el-form-item__content) {
    gap: 10px;
    align-items: center;
  }
  .trace-switch-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  /* 空态 */
  .trace-empty-card :deep(.el-card__body) {
    padding: 40px 20px;
  }

  /* 统计卡片网格 */
  .trace-stats {
    display: grid;
    grid-template-columns: repeat(6, minmax(0, 1fr));
    gap: 12px;
  }
  .stat-card {
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 82px;
    transition: border-color 0.2s;
  }
  .stat-card:hover {
    border-color: var(--el-color-primary-light-5);
  }
  .stat-card-wide {
    grid-column: span 2;
  }
  .stat-label {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
  .stat-value {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }
  .stat-number {
    font-size: 22px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
  }
  .stat-sub {
    margin-left: 4px;
    font-size: 13px;
    font-weight: 400;
    color: var(--el-text-color-secondary);
  }
  .stat-meta {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  /* 身份条 */
  .trace-identity {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 20px;
    padding: 10px 14px;
    background: var(--el-fill-color-light);
    border-radius: 6px;
    font-size: 12px;
  }
  .identity-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .identity-key {
    color: var(--el-text-color-secondary);
  }
  .identity-value {
    color: var(--el-text-color-primary);
  }

  /* 结果卡 */
  .trace-result-card :deep(.el-card__body) {
    padding: 8px 12px 12px;
  }
  .trace-table {
    width: 100%;
    font-size: 12px;
  }
  .trace-table :deep(.el-table__header) th .cell {
    font-size: 12px;
    padding: 0 8px;
    line-height: 1.4;
  }
  .trace-table :deep(.el-table__row) {
    font-size: 12px;
  }
  .trace-table :deep(.el-table__row) td .cell {
    padding: 0 8px;
    line-height: 1.45;
  }
  .trace-table :deep(.el-table__row) td {
    padding: 6px 0;
  }
  .trace-table :deep(.el-tag) {
    font-size: 11px;
    height: 20px;
    padding: 0 6px;
    line-height: 18px;
  }

  /* 菜单树 */
  .menu-tree {
    padding: 6px 4px 10px;
    background: transparent;
    font-size: 13px;
  }
  .menu-tree :deep(.el-tree-node__content) {
    height: 30px;
    border-radius: 4px;
  }
  .menu-tree :deep(.el-tree-node__content:hover) {
    background: var(--el-fill-color-light);
  }
  .menu-node {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    min-width: 0;
  }
  .menu-node-hidden .menu-node-name {
    color: var(--el-text-color-placeholder);
    text-decoration: line-through;
  }
  .menu-node-tag {
    flex-shrink: 0;
  }
  .menu-node-name {
    font-weight: 500;
    color: var(--el-text-color-primary);
  }
  .menu-node-kind {
    font-size: 10px;
  }
  .menu-node-path {
    color: var(--el-text-color-secondary);
    font-size: 11px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .page-search {
    width: 260px;
  }
  .page-cell {
    display: flex;
    flex-direction: column;
    gap: 2px;
    line-height: 1.35;
  }
  .page-cell-name {
    font-size: 13px;
    color: var(--el-text-color-primary);
    font-weight: 500;
  }
  .page-cell-key {
    font-size: 11px;
    color: var(--el-text-color-secondary);
  }
  .chain-cell {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
    font-size: 12px;
  }
  .chain-seg {
    padding: 2px 8px;
    background: var(--el-fill-color);
    border-radius: 4px;
    color: var(--el-text-color-regular);
    font-family: var(--el-font-family-mono, ui-monospace, Menlo, monospace);
    font-size: 11px;
  }
  .chain-arrow {
    color: var(--el-text-color-placeholder);
    font-size: 12px;
  }

  .mono {
    font-family: var(--el-font-family-mono, ui-monospace, Menlo, Consolas, monospace);
    font-size: 12px;
  }
  .text-muted {
    color: var(--el-text-color-placeholder);
  }

  /* 响应式 */
  @media (max-width: 1440px) {
    .trace-form {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
    .trace-stats {
      grid-template-columns: repeat(4, minmax(0, 1fr));
    }
    .stat-card-wide {
      grid-column: span 2;
    }
  }
  @media (max-width: 1024px) {
    .trace-form {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .trace-stats {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .stat-card-wide {
      grid-column: span 2;
    }
    .page-search {
      width: 180px;
    }
  }
  @media (max-width: 768px) {
    .trace-form {
      grid-template-columns: 1fr;
    }
    .trace-stats {
      grid-template-columns: 1fr;
    }
    .stat-card-wide {
      grid-column: span 1;
    }
  }
</style>
