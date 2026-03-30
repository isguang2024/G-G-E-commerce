<template>
  <div class="menu-space-page art-full-height" v-loading="loading">
    <AdminWorkspaceHero
      title="菜单空间"
      description="先定义菜单空间，再按需绑定 Host。默认空间继续兼容当前单域单菜单模式，新增空间后也可以先在单域里逐步接入。"
      :metrics="summaryMetrics"
    >
      <div class="menu-space-hero-actions">
        <ElSelect v-model="spaceMode" style="width: 140px">
          <ElOption label="单空间模式" value="single" />
          <ElOption label="多空间模式" value="multi" />
        </ElSelect>
        <ElButton :loading="savingSpaceMode" @click="saveSpaceMode" v-ripple>保存模式</ElButton>
        <ElButton type="primary" @click="openSpaceDrawer()" v-ripple>新增空间</ElButton>
        <ElButton @click="openHostDrawer()" v-ripple>新增 Host 绑定</ElButton>
        <ElButton @click="loadData" v-ripple>刷新</ElButton>
      </div>
    </AdminWorkspaceHero>

    <ElAlert
      v-if="loadError"
      class="menu-space-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="menu-space-board">
      <ElCard class="menu-space-panel" shadow="never">
        <template #header>
          <div class="menu-space-panel__header">
            <div>
              <div class="menu-space-panel__title">菜单空间</div>
              <div class="menu-space-panel__desc">默认空间始终兜底存在；新增空间后，可以逐步把菜单和页面迁过去。</div>
            </div>
          </div>
        </template>

        <div class="menu-space-list">
          <button
            v-for="item in spaces"
            :key="item.spaceKey"
            type="button"
            class="menu-space-item"
            :class="{ 'is-current': currentSpaceKey === item.spaceKey }"
            @click="currentSpaceKey = item.spaceKey"
          >
            <div class="menu-space-item__main">
              <div class="menu-space-item__title-row">
                <span class="menu-space-item__title">{{ item.name }}</span>
                <ElTag v-if="item.isDefault" size="small" type="success" effect="plain">默认</ElTag>
                <ElTag
                  size="small"
                  :type="item.status === 'normal' ? 'info' : 'danger'"
                  effect="plain"
                >
                  {{ item.status === 'normal' ? '启用' : '停用' }}
                </ElTag>
              </div>
              <div class="menu-space-item__meta">
                <span>标识 {{ item.spaceKey }}</span>
                <span>首页 {{ item.defaultHomePath || '-' }}</span>
                <span>菜单 {{ item.menuCount || 0 }}</span>
                <span>页面 {{ item.pageCount || 0 }}</span>
                <span>准入 {{ getAccessModeLabel(item.accessMode) }}</span>
                <span>Host {{ item.hostCount || 0 }}</span>
              </div>
              <p class="menu-space-item__desc">
                {{ item.description || '当前空间未填写描述，建议补充业务边界或使用说明。' }}
              </p>
              <div class="menu-space-item__hosts">
                <span class="menu-space-chip" :class="{ 'is-soft': !isSpaceInitialized(item) }">
                  {{ isSpaceInitialized(item) ? '已初始化' : '待初始化' }}
                </span>
                <span
                  v-for="host in item.hosts?.slice(0, 3)"
                  :key="host"
                  class="menu-space-chip"
                >
                  {{ host }}
                </span>
                <span v-if="(item.hosts?.length || 0) > 3" class="menu-space-chip is-soft">
                  +{{ (item.hosts?.length || 0) - 3 }}
                </span>
              </div>
            </div>
              <div class="menu-space-item__actions">
                <ElButton text type="primary" @click.stop="openSpaceDrawer(item)">编辑</ElButton>
                <ElButton text @click.stop="goToMenuManagement(item.spaceKey)">菜单管理</ElButton>
                <ElButton text @click.stop="goToPageManagement(item.spaceKey)">页面管理</ElButton>
                <ElButton
                  v-if="!item.isDefault && !isSpaceInitialized(item)"
                  text
                  :loading="initializingSpaceKey === item.spaceKey"
                  @click.stop="initializeSpace(item)"
                >
                  复制默认菜单
                </ElButton>
                <ElButton
                  v-else-if="!item.isDefault"
                  text
                  disabled
                >
                  已初始化
                </ElButton>
                <ElButton text @click.stop="openHostDrawer(undefined, item.spaceKey)">绑定 Host</ElButton>
              </div>
          </button>
        </div>
      </ElCard>

      <ElCard class="menu-space-panel" shadow="never">
        <template #header>
          <div class="menu-space-panel__header">
            <div>
              <div class="menu-space-panel__title">Host 绑定</div>
              <div class="menu-space-panel__desc">可选配置。未命中任何 Host 时，系统会自动退回默认菜单空间。</div>
            </div>
            <div class="menu-space-panel__status">
              <ElTag effect="plain" type="info">当前解析 {{ currentSpaceLabel }}</ElTag>
              <ElTag effect="plain" :type="spaceModeTagType">模式 {{ spaceModeLabel }}</ElTag>
              <ElTag v-if="resolveByLabel" effect="plain" type="warning">来源 {{ resolveByLabel }}</ElTag>
            </div>
          </div>
        </template>

        <div v-if="currentSpace" class="menu-space-overview">
          <div class="menu-space-overview__grid">
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">当前空间</span>
              <strong>{{ currentSpace.name }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">默认首页</span>
              <strong>{{ currentSpace.defaultHomePath || '未设置' }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">菜单 / 页面</span>
              <strong>{{ currentSpace.menuCount || 0 }} / {{ currentSpace.pageCount || 0 }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">初始化状态</span>
              <strong>{{ isSpaceInitialized(currentSpace) ? '已初始化' : '待初始化' }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">空间准入</span>
              <strong>{{ getAccessModeSummary(currentSpace) }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">解析来源</span>
              <strong>{{ resolveByLabel || '未返回' }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">请求 Host</span>
              <strong>{{ currentRequestHost || '未命中' }}</strong>
            </div>
          </div>
          <div class="menu-space-overview__actions">
            <ElButton text @click="goToMenuManagement(currentSpace.spaceKey)">进入菜单管理</ElButton>
            <ElButton text @click="goToPageManagement(currentSpace.spaceKey)">进入页面管理</ElButton>
            <ElButton
              v-if="!currentSpace.isDefault && isSpaceInitialized(currentSpace)"
              text
              type="danger"
              :loading="initializingSpaceKey === currentSpace.spaceKey"
              @click="reinitializeSpace(currentSpace)"
            >
              重新初始化
            </ElButton>
          </div>
        </div>

        <div class="menu-space-binding-list">
          <div v-if="!hostBindings.length" class="menu-space-empty">
            还没有 Host 绑定。当前依然按默认菜单空间运行，不影响现有架构。
          </div>
          <button
            v-for="item in hostBindings"
            :key="item.host"
            type="button"
            class="menu-space-binding"
            @click="openHostDrawer(item)"
          >
            <div class="menu-space-binding__main">
              <div class="menu-space-binding__title-row">
                <span class="menu-space-binding__host">{{ item.host }}</span>
                <ElTag
                  size="small"
                  :type="item.status === 'normal' ? 'success' : 'danger'"
                  effect="plain"
                >
                  {{ item.status === 'normal' ? '启用' : '停用' }}
                </ElTag>
              </div>
              <div class="menu-space-binding__meta">
                <span>菜单空间 {{ item.spaceName || item.spaceKey }}</span>
                <span v-if="item.isDefault">主绑定</span>
                <span>{{ getHostAuthModeLabel(item.authMode) }}</span>
                <span v-if="item.routePrefix">前缀 {{ item.routePrefix }}</span>
                <span v-if="item.description">{{ item.description }}</span>
              </div>
            </div>
            <ArtSvgIcon icon="ri:arrow-right-s-line" />
          </button>
        </div>
      </ElCard>
    </section>

    <ElDrawer v-model="spaceDrawerVisible" :title="spaceDrawerTitle" size="520px" destroy-on-close>
      <ElForm ref="spaceFormRef" :model="spaceForm" label-position="top">
        <ElFormItem label="空间名称">
          <ElInput v-model="spaceForm.name" placeholder="例如 默认菜单空间 / 平台运营空间" />
        </ElFormItem>
        <ElFormItem label="空间标识">
          <ElInput
            v-model="spaceForm.space_key"
            :disabled="spaceForm.is_default"
            placeholder="例如 default / platform / crm"
          />
        </ElFormItem>
        <ElFormItem label="默认首页">
          <ElSelect
            v-model="spaceForm.default_home_path"
            filterable
            allow-create
            clearable
            default-first-option
            style="width: 100%"
            placeholder="请选择或输入空间默认落地页"
          >
            <ElOption
              v-for="item in landingPathOptions"
              :key="item"
              :label="item"
              :value="item"
            />
          </ElSelect>
          <div class="field-hint">
            {{ landingPathHint }}
          </div>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput v-model="spaceForm.description" type="textarea" :rows="3" placeholder="说明这个菜单空间承载什么菜单与页面" />
        </ElFormItem>
        <ElFormItem label="空间准入">
          <ElSelect v-model="spaceForm.access_mode" style="width: 100%">
            <ElOption label="全部可进" value="all" />
            <ElOption label="仅平台管理员" value="platform_admin" />
            <ElOption label="仅团队管理员" value="team_admin" />
            <ElOption label="指定角色码" value="role_codes" />
          </ElSelect>
          <div class="field-hint">先决定谁有资格进入这个菜单空间，进入后再按菜单与页面权限继续裁剪。</div>
        </ElFormItem>
        <ElFormItem v-if="spaceForm.access_mode === 'role_codes'" label="允许角色码">
          <ElInput
            v-model="allowedRoleCodesText"
            type="textarea"
            :rows="3"
            placeholder="多个角色码用英文逗号分隔，例如 admin, team_admin, ops_manager"
          />
        </ElFormItem>
        <div class="menu-space-drawer-grid">
          <ElFormItem label="默认空间">
            <ElSwitch v-model="spaceForm.is_default" />
          </ElFormItem>
          <ElFormItem label="状态">
            <ElSelect v-model="spaceForm.status">
              <ElOption label="启用" value="normal" />
              <ElOption label="停用" value="disabled" />
            </ElSelect>
          </ElFormItem>
        </div>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="spaceDrawerVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingSpace" @click="saveSpace">保存</ElButton>
        </div>
      </template>
    </ElDrawer>

    <ElDrawer v-model="hostDrawerVisible" :title="hostDrawerTitle" size="520px" destroy-on-close>
      <ElForm ref="hostFormRef" :model="hostForm" label-position="top">
        <ElFormItem label="Host / 子域名">
          <ElInput v-model="hostForm.host" placeholder="例如 admin.example.com 或 team.example.com" />
        </ElFormItem>
        <ElFormItem label="菜单空间">
          <ElSelect v-model="hostForm.space_key" filterable style="width: 100%">
            <ElOption v-for="item in spaceOptions" :key="item.value" :label="item.label" :value="item.value" />
          </ElSelect>
        </ElFormItem>
        <div class="menu-space-drawer-grid">
          <ElFormItem label="访问协议">
            <ElSelect v-model="hostForm.meta.scheme">
              <ElOption label="HTTPS" value="https" />
              <ElOption label="HTTP" value="http" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="路由前缀">
            <ElInput v-model="hostForm.meta.route_prefix" placeholder="例如 /admin，可留空" />
          </ElFormItem>
        </div>
        <ElFormItem label="认证策略">
          <ElSelect v-model="hostForm.meta.auth_mode" style="width: 100%">
            <ElOption label="沿用当前 Host 登录" value="inherit_host" />
            <ElOption label="统一登录 Host" value="centralized_login" />
            <ElOption label="共享 Cookie 域" value="shared_cookie" />
          </ElSelect>
          <div class="field-hint">默认沿用当前 Host。后续真正启用多 Host 时，再按这里的策略决定登录入口和回跳方式。</div>
        </ElFormItem>
        <div v-if="hostForm.meta.auth_mode === 'centralized_login'" class="menu-space-drawer-grid">
          <ElFormItem label="统一登录 Host">
            <ElInput v-model="hostForm.meta.login_host" placeholder="例如 auth.example.com" />
          </ElFormItem>
          <ElFormItem label="登录回调 Host">
            <ElInput v-model="hostForm.meta.callback_host" placeholder="例如 admin.example.com，可留空默认当前 Host" />
          </ElFormItem>
        </div>
        <div v-if="hostForm.meta.auth_mode === 'shared_cookie'" class="menu-space-drawer-grid">
          <ElFormItem label="Cookie 作用域">
            <ElSelect v-model="hostForm.meta.cookie_scope_mode">
              <ElOption label="沿用默认" value="inherit" />
              <ElOption label="仅当前 Host" value="host_only" />
              <ElOption label="父域共享" value="parent_domain" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="Cookie 域">
            <ElInput v-model="hostForm.meta.cookie_domain" placeholder="例如 .example.com，可留空" />
          </ElFormItem>
        </div>
        <ElFormItem label="说明">
          <ElInput v-model="hostForm.description" type="textarea" :rows="3" placeholder="例如 平台治理入口 / 团队工作区入口" />
        </ElFormItem>
        <ElFormItem label="主绑定">
          <ElSwitch v-model="hostForm.is_default" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="hostForm.status">
            <ElOption label="启用" value="normal" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="hostDrawerVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingHost" @click="saveHostBinding">保存</ElButton>
        </div>
      </template>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import type { AppRouteRecord } from '@/types/router'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import {
    fetchGetCurrentMenuSpace,
    fetchGetMenuSpaceMode,
    fetchInitializeMenuSpaceFromDefault,
    fetchGetMenuList,
    fetchGetMenuSpaceHostBindings,
    fetchGetMenuSpaces,
    fetchGetPageList,
    fetchSaveMenuSpace,
    fetchSaveMenuSpaceHostBinding,
    fetchUpdateMenuSpaceMode
  } from '@/api/system-manage'
  import { normalizeMenuSpaceKey } from '@/utils/navigation/menu-space'

  defineOptions({ name: 'MenuSpaceManage' })
  const router = useRouter()

  const loading = ref(false)
  const loadError = ref('')
  const savingSpace = ref(false)
  const savingHost = ref(false)
  const savingSpaceMode = ref(false)
  const initializingSpaceKey = ref('')
  const spaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const hostBindings = ref<Api.SystemManage.MenuSpaceHostBindingItem[]>([])
  const currentSpaceKey = ref('default')
  const spaceMode = ref<'single' | 'multi'>('single')
  const currentResolvedBy = ref('')
  const currentRequestHost = ref('')
  const currentAccessGranted = ref(true)
  const loadingLandingPaths = ref(false)
  const landingPathOptions = ref<string[]>([])
  const warnDev = (...args: any[]) => {
    if (import.meta.env.DEV) {
      console.warn(...args)
    }
  }

  const spaceDrawerVisible = ref(false)
  const hostDrawerVisible = ref(false)
  const editingSpaceKey = ref('')
  const editingHost = ref('')

  const spaceFormRef = ref()
  const hostFormRef = ref()
  type HostBindingMetaForm = NonNullable<Api.SystemManage.MenuSpaceHostBindingSaveParams['meta']>
  type HostBindingSaveForm = Omit<Api.SystemManage.MenuSpaceHostBindingSaveParams, 'meta'> & {
    meta: HostBindingMetaForm
  }

  const spaceForm = reactive<Api.SystemManage.MenuSpaceSaveParams>({
    space_key: 'default',
    name: '',
    description: '',
    default_home_path: '/dashboard/console',
    is_default: false,
    status: 'normal',
    access_mode: 'all',
    allowed_role_codes: [],
    meta: {}
  })
  const allowedRoleCodesText = ref('')

  const hostForm = reactive<HostBindingSaveForm>({
    host: '',
    space_key: 'default',
    description: '',
    is_default: false,
    status: 'normal',
    meta: {
      scheme: 'https',
      route_prefix: '',
      auth_mode: 'inherit_host',
      login_host: '',
      callback_host: '',
      cookie_scope_mode: 'inherit',
      cookie_domain: ''
    }
  })

  const currentSpace = computed(
    () => spaces.value.find((item) => item.spaceKey === currentSpaceKey.value) || spaces.value[0]
  )

  const currentSpaceLabel = computed(() => currentSpace.value?.name || currentSpace.value?.spaceKey || '默认菜单空间')
  const spaceModeLabel = computed(() => {
    if (spaceMode.value === 'single') {
      return '单空间'
    }
    return '多空间'
  })
  const spaceModeTagType = computed(() => (spaceModeLabel.value === '单空间' ? 'success' : 'warning'))
  const resolveByLabel = computed(() => {
    const value = `${currentResolvedBy.value || ''}`.trim()
    switch (value) {
      case 'single_mode':
        return '单空间默认'
      case 'single_mode_explicit':
        return '单空间显式指定'
      case 'host':
        return 'Host 命中'
      case 'explicit':
        return '参数显式指定'
      case 'default':
        return '默认空间'
      case 'fallback_default':
        return currentAccessGranted.value ? '默认空间' : '无权限回退默认空间'
      default:
        return value
    }
  })

  const summaryMetrics = computed(() => [
    { label: '菜单空间', value: spaces.value.length || 0 },
    { label: 'Host 绑定', value: hostBindings.value.length || 0 },
    { label: '已初始化', value: spaces.value.filter((item) => isSpaceInitialized(item)).length || 0 },
    { label: '当前解析', value: currentSpace.value?.spaceKey || 'default' }
  ])

  const spaceOptions = computed(() =>
    spaces.value.map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )

  const spaceDrawerTitle = computed(() => (editingSpaceKey.value ? '编辑菜单空间' : '新增菜单空间'))
  const hostDrawerTitle = computed(() => (editingHost.value ? '编辑 Host 绑定' : '新增 Host 绑定'))
  const landingPathHint = computed(() => {
    if (loadingLandingPaths.value) {
      return '正在加载当前空间下可用的页面路径。'
    }
    const value = `${spaceForm.default_home_path || ''}`.trim()
    if (!value) {
      return '未填写时，系统会退回当前菜单空间下首个可进入页面。'
    }
    if (!value.startsWith('/')) {
      return '默认首页必须是以 / 开头的站内路径，例如 /dashboard/console。'
    }
    if (!landingPathOptions.value.length) {
      return '当前空间下还没有可选页面路径，可以先留空，等菜单和页面配置完成后再回填。'
    }
    if (!landingPathOptions.value.includes(value)) {
      return '当前填写的路径不在这个菜单空间的已注册页面里，保存前建议先确认菜单或页面是否已经归属到该空间。'
    }
    return '该路径已命中当前菜单空间的可选页面，登录后和进入根路径时会优先跳到这里。'
  })

  function isSpaceInitialized(item?: Api.SystemManage.MenuSpaceItem) {
    if (!item) return false
    return Number(item.menuCount || 0) > 0 || Number(item.pageCount || 0) > 0
  }

  function normalizeRoleCodeListText(value: string) {
    return Array.from(
      new Set(
        `${value || ''}`
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean)
      )
    )
  }

  function getAccessModeLabel(value?: string) {
    switch (`${value || 'all'}`.trim()) {
      case 'platform_admin':
        return '仅平台管理员'
      case 'team_admin':
        return '仅团队管理员'
      case 'role_codes':
        return '指定角色码'
      default:
        return '全部可进'
    }
  }

  function getAccessModeSummary(item?: Api.SystemManage.MenuSpaceItem) {
    if (!item) return '全部可进'
    if (`${item.accessMode || 'all'}`.trim() !== 'role_codes') {
      return getAccessModeLabel(item.accessMode)
    }
    const codes = item.allowedRoleCodes || []
    return codes.length ? `指定角色码 · ${codes.join(' / ')}` : '指定角色码'
  }

  function getHostAuthModeLabel(value?: string) {
    switch (`${value || 'inherit_host'}`.trim()) {
      case 'centralized_login':
        return '统一登录 Host'
      case 'shared_cookie':
        return '共享 Cookie 域'
      default:
        return '沿用当前 Host'
    }
  }

  function normalizeInternalPath(value: string): string {
    const target = `${value || ''}`.trim()
    if (!target || /^https?:\/\//i.test(target)) {
      return ''
    }
    const normalized = `/${target.replace(/^\/+/, '')}`.replace(/\/+/g, '/')
    return normalized === '/' ? normalized : normalized.replace(/\/$/, '')
  }

  async function loadData() {
    loading.value = true
    loadError.value = ''
    try {
      const [spaceRes, hostRes, currentRes, modeRes] = await Promise.all([
        fetchGetMenuSpaces(),
        fetchGetMenuSpaceHostBindings(),
        fetchGetCurrentMenuSpace().catch(() => undefined),
        fetchGetMenuSpaceMode().catch(() => ({ mode: 'single' }))
      ])
      spaces.value = spaceRes.records || []
      hostBindings.value = hostRes.records || []
      currentSpaceKey.value =
        currentRes?.space?.spaceKey || spaces.value.find((item) => item.isDefault)?.spaceKey || 'default'
      spaceMode.value = `${modeRes?.mode || 'single'}`.trim() === 'multi' ? 'multi' : 'single'
      currentResolvedBy.value = `${currentRes?.resolvedBy || ''}`.trim()
      currentRequestHost.value = `${currentRes?.requestHost || ''}`.trim()
      currentAccessGranted.value = Boolean(currentRes?.accessGranted ?? true)
    } catch (error: any) {
      spaces.value = []
      hostBindings.value = []
      currentSpaceKey.value = 'default'
      spaceMode.value = 'single'
      currentResolvedBy.value = ''
      currentRequestHost.value = ''
      currentAccessGranted.value = true
      loadError.value = error?.message || '菜单空间数据暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  async function saveSpaceMode() {
    savingSpaceMode.value = true
    try {
      const res = await fetchUpdateMenuSpaceMode(spaceMode.value)
      spaceMode.value = `${res?.mode || 'single'}`.trim() === 'multi' ? 'multi' : 'single'
      ElMessage.success(`菜单空间模式已更新为${spaceModeLabel.value}`)
      await loadData()
    } catch (error: any) {
      ElMessage.error(error?.message || '菜单空间模式保存失败')
    } finally {
      savingSpaceMode.value = false
    }
  }

  function collectMenuPaths(items: AppRouteRecord[] = []): string[] {
    const result: string[] = []
    const joinMenuPath = (parentPath: string, currentPath: string) => {
      const target = `${currentPath || ''}`.trim()
      if (!target) return ''
      if (/^https?:\/\//i.test(target)) return ''
      if (target.startsWith('/')) {
        return normalizeInternalPath(target)
      }
      const base = normalizeInternalPath(parentPath)
      return normalizeInternalPath(`${base}/${target}`)
    }
    const walk = (list: AppRouteRecord[], parentPath = '') => {
      ;(list || []).forEach((item) => {
        const normalizedPath = joinMenuPath(parentPath, `${item.path || ''}`)
        if (
          normalizedPath &&
          !item.children?.length &&
          item.meta?.isEnable !== false
        ) {
          result.push(normalizedPath)
        }
        if (item.children?.length) {
          walk(item.children, normalizedPath || parentPath)
        }
      })
    }
    walk(items)
    return result
  }

  async function loadLandingPathCandidates(spaceKey: string) {
    const normalizedSpaceKey = normalizeMenuSpaceKey(spaceKey) || 'default'
    loadingLandingPaths.value = true
    try {
      const [menuRes, pageRes] = await Promise.all([
        fetchGetMenuList(normalizedSpaceKey),
        fetchGetPageList({
          current: 1,
          size: 1000,
          spaceKey: normalizedSpaceKey
        })
      ])
      const pagePaths = (pageRes.records || [])
        .filter(
          (item) =>
            item.status === 'normal' &&
            item.pageType !== 'group' &&
            item.pageType !== 'display_group' &&
            normalizeInternalPath(`${item.routePath || ''}`.trim())
        )
        .map((item) => normalizeInternalPath(`${item.routePath || ''}`.trim()))
      landingPathOptions.value = Array.from(
        new Set([...collectMenuPaths(menuRes || []), ...pagePaths])
      ).sort((a, b) => a.localeCompare(b, 'zh-CN'))
    } catch (error) {
      warnDev('[MenuSpaceManage] 加载默认首页候选失败，已回退为空列表', error)
      landingPathOptions.value = []
    } finally {
      loadingLandingPaths.value = false
    }
  }

  function resetSpaceForm() {
    editingSpaceKey.value = ''
    spaceForm.space_key = 'default'
    spaceForm.name = ''
    spaceForm.description = ''
    spaceForm.default_home_path = '/dashboard/console'
    spaceForm.is_default = false
    spaceForm.status = 'normal'
    spaceForm.access_mode = 'all'
    spaceForm.allowed_role_codes = []
    allowedRoleCodesText.value = ''
    spaceForm.meta = {}
  }

  function resetHostForm() {
    editingHost.value = ''
    hostForm.host = ''
    hostForm.space_key = currentSpaceKey.value || 'default'
    hostForm.description = ''
    hostForm.is_default = false
    hostForm.status = 'normal'
    hostForm.meta = {
      scheme: 'https',
      route_prefix: '',
      auth_mode: 'inherit_host',
      login_host: '',
      callback_host: '',
      cookie_scope_mode: 'inherit',
      cookie_domain: ''
    }
  }

  function openSpaceDrawer(item?: Api.SystemManage.MenuSpaceItem) {
    resetSpaceForm()
    if (item) {
      editingSpaceKey.value = item.spaceKey
      spaceForm.space_key = item.spaceKey
      spaceForm.name = item.name
      spaceForm.description = item.description || ''
      spaceForm.default_home_path = item.defaultHomePath || '/dashboard/console'
      spaceForm.is_default = Boolean(item.isDefault)
      spaceForm.status = item.status || 'normal'
      spaceForm.access_mode = `${item.accessMode || 'all'}`.trim() || 'all'
      spaceForm.allowed_role_codes = [...(item.allowedRoleCodes || [])]
      allowedRoleCodesText.value = (item.allowedRoleCodes || []).join(', ')
      spaceForm.meta = item.meta || {}
    }
    spaceDrawerVisible.value = true
    loadLandingPathCandidates(spaceForm.space_key)
  }

  function openHostDrawer(item?: Api.SystemManage.MenuSpaceHostBindingItem, preferredSpaceKey?: string) {
    resetHostForm()
    if (item) {
      editingHost.value = item.host
      hostForm.host = item.host
      hostForm.space_key = item.spaceKey
      hostForm.description = item.description || ''
      hostForm.is_default = Boolean(item.isDefault)
      hostForm.status = item.status || 'normal'
      hostForm.meta = {
        scheme: item.scheme || 'https',
        route_prefix: item.routePrefix || '',
        auth_mode: item.authMode || 'inherit_host',
        login_host: item.loginHost || '',
        callback_host: item.callbackHost || '',
        cookie_scope_mode: item.cookieScopeMode || 'inherit',
        cookie_domain: item.cookieDomain || '',
        ...(item.meta || {})
      }
    } else if (preferredSpaceKey) {
      hostForm.space_key = normalizeMenuSpaceKey(preferredSpaceKey)
    }
    hostDrawerVisible.value = true
  }

  async function saveSpace() {
    if (!spaceForm.name.trim()) {
      ElMessage.warning('请输入空间名称')
      return
    }
    const normalizedHomePath = normalizeInternalPath(spaceForm.default_home_path || '')
    if (spaceForm.default_home_path?.trim() && !normalizedHomePath) {
      ElMessage.warning('默认首页必须是以 / 开头的站内路径')
      return
    }
    if (
      normalizedHomePath &&
      landingPathOptions.value.length > 0 &&
      !landingPathOptions.value.includes(normalizedHomePath)
    ) {
      ElMessage.warning('默认首页未命中当前菜单空间的已注册页面，请先确认菜单或页面归属')
      return
    }
    savingSpace.value = true
    try {
      const allowedRoleCodes = normalizeRoleCodeListText(allowedRoleCodesText.value)
      if (spaceForm.access_mode === 'role_codes' && allowedRoleCodes.length === 0) {
        ElMessage.warning('请至少填写一个允许进入该空间的角色码')
        savingSpace.value = false
        return
      }
      await fetchSaveMenuSpace({
        ...spaceForm,
        space_key: normalizeMenuSpaceKey(spaceForm.space_key),
        name: spaceForm.name.trim(),
        description: spaceForm.description?.trim() || '',
        default_home_path: normalizedHomePath,
        access_mode: spaceForm.access_mode || 'all',
        allowed_role_codes: spaceForm.access_mode === 'role_codes' ? allowedRoleCodes : []
      })
      ElMessage.success('菜单空间已保存')
      spaceDrawerVisible.value = false
      await loadData()
    } catch (error: any) {
      ElMessage.error(error?.message || '菜单空间保存失败')
    } finally {
      savingSpace.value = false
    }
  }

  async function saveHostBinding() {
    if (!hostForm.host.trim()) {
      ElMessage.warning('请输入 Host')
      return
    }
    const normalizedHost = `${hostForm.host || ''}`.trim().toLowerCase()
    const duplicatedBinding = hostBindings.value.find(
      (item) =>
        `${item.host || ''}`.trim().toLowerCase() === normalizedHost &&
        normalizeMenuSpaceKey(item.spaceKey) !== normalizeMenuSpaceKey(hostForm.space_key)
    )
    if (duplicatedBinding) {
      ElMessage.warning(`该 Host 已绑定到菜单空间 ${duplicatedBinding.spaceName || duplicatedBinding.spaceKey}`)
      return
    }
    savingHost.value = true
    try {
      await fetchSaveMenuSpaceHostBinding({
        ...hostForm,
        host: hostForm.host.trim(),
        space_key: normalizeMenuSpaceKey(hostForm.space_key),
        description: hostForm.description?.trim() || '',
        meta: {
          ...hostForm.meta,
          scheme: `${hostForm.meta?.scheme || 'https'}`.trim() || 'https',
          route_prefix: `${hostForm.meta?.route_prefix || ''}`.trim(),
          auth_mode: `${hostForm.meta?.auth_mode || 'inherit_host'}`.trim() || 'inherit_host',
          login_host: `${hostForm.meta?.login_host || ''}`.trim(),
          callback_host: `${hostForm.meta?.callback_host || ''}`.trim(),
          cookie_scope_mode: `${hostForm.meta?.cookie_scope_mode || 'inherit'}`.trim() || 'inherit',
          cookie_domain: `${hostForm.meta?.cookie_domain || ''}`.trim()
        }
      })
      ElMessage.success('Host 绑定已保存')
      hostDrawerVisible.value = false
      await loadData()
    } catch (error: any) {
      ElMessage.error(error?.message || 'Host 绑定保存失败')
    } finally {
      savingHost.value = false
    }
  }

  async function initializeSpace(item: Api.SystemManage.MenuSpaceItem) {
    if (!item?.spaceKey || item.isDefault) {
      return
    }
    if (isSpaceInitialized(item)) {
      ElMessage.info('当前菜单空间已经初始化，可直接进入菜单管理或页面管理继续调整')
      return
    }
    initializingSpaceKey.value = item.spaceKey
    try {
      const result = await fetchInitializeMenuSpaceFromDefault(item.spaceKey)
      ElMessage.success(
        `已完成初始化：复制 ${result.createdMenuCount} 个菜单、${result.createdPageCount} 个页面、${result.createdPackageMenuLinkCount} 条功能包菜单关联`
      )
      await loadData()
      goToMenuManagement(item.spaceKey)
    } catch (error: any) {
      ElMessage.error(error?.message || '复制默认空间菜单失败')
    } finally {
      initializingSpaceKey.value = ''
    }
  }

  async function reinitializeSpace(item: Api.SystemManage.MenuSpaceItem) {
    if (!item?.spaceKey || item.isDefault || !isSpaceInitialized(item)) {
      return
    }
    try {
      await ElMessageBox.confirm(
        `重新初始化会清空空间“${item.name}”当前已有的菜单、页面和功能包菜单关联，然后重新复制默认空间内容。这个操作适合首次搭建后重来一次，不适合保留现有裁剪结果。`,
        '确认重新初始化',
        {
          confirmButtonText: '确认覆盖',
          cancelButtonText: '取消',
          type: 'warning',
          distinguishCancelAndClose: true
        }
      )
    } catch {
      return
    }
    initializingSpaceKey.value = item.spaceKey
    try {
      const result = await fetchInitializeMenuSpaceFromDefault(item.spaceKey, true)
      ElMessage.success(
        `已重新初始化：清空 ${result.clearedMenuCount || 0} 个菜单、${result.clearedPageCount || 0} 个页面、${result.clearedPackageMenuLinkCount || 0} 条功能包菜单关联，并重新复制 ${result.createdMenuCount} 个菜单、${result.createdPageCount} 个页面`
      )
      await loadData()
      goToMenuManagement(item.spaceKey)
    } catch (error: any) {
      ElMessage.error(error?.message || '重新初始化失败')
    } finally {
      initializingSpaceKey.value = ''
    }
  }

  function goToMenuManagement(spaceKey: string) {
    router.push({
      path: '/system/menu',
      query: { spaceKey: normalizeMenuSpaceKey(spaceKey) }
    })
  }

  function goToPageManagement(spaceKey: string) {
    router.push({
      path: '/system/page',
      query: { spaceKey: normalizeMenuSpaceKey(spaceKey) }
    })
  }

  onMounted(() => {
    loadData()
  })

  watch(
    () => spaceForm.space_key,
    (value, previousValue) => {
      if (!spaceDrawerVisible.value) return
      if (!`${value || ''}`.trim() || value === previousValue) return
      if (spaceForm.default_home_path && !landingPathOptions.value.includes(spaceForm.default_home_path)) {
        spaceForm.default_home_path = ''
      }
      loadLandingPathCandidates(value)
    }
  )

  watch(
    () => spaceForm.access_mode,
    (value) => {
      if (value !== 'role_codes') {
        allowedRoleCodesText.value = ''
      }
    }
  )

  watch(
    () => hostForm.meta?.auth_mode,
    (value) => {
      if (value !== 'centralized_login') {
        hostForm.meta = {
          ...hostForm.meta,
          login_host: '',
          callback_host: ''
        }
      }
      if (value !== 'shared_cookie') {
        hostForm.meta = {
          ...hostForm.meta,
          cookie_scope_mode: 'inherit',
          cookie_domain: ''
        }
      }
    }
  )
</script>

<style scoped lang="scss">
  .menu-space-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .menu-space-inline-alert {
    margin-top: 16px;
  }

  .menu-space-board {
    display: grid;
    grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
    gap: 16px;
    margin-top: 16px;
  }

  .menu-space-panel {
    border-radius: 18px;
  }

  .menu-space-panel__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .menu-space-panel__status {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
  }

  .menu-space-panel__title {
    font-size: 18px;
    font-weight: 600;
    color: var(--art-gray-900);
  }

  .menu-space-panel__desc {
    margin-top: 6px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--art-gray-600);
  }

  .menu-space-list,
  .menu-space-binding-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .menu-space-item,
  .menu-space-binding {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    width: 100%;
    padding: 16px 18px;
    text-align: left;
    background: linear-gradient(180deg, #fbfcff 0%, #f5f8ff 100%);
    border: 1px solid rgba(55, 125, 255, 0.12);
    border-radius: 16px;
    transition: 0.2s ease;
  }

  .menu-space-item:hover,
  .menu-space-binding:hover,
  .menu-space-item.is-current {
    border-color: rgba(55, 125, 255, 0.32);
    box-shadow: 0 12px 30px rgba(55, 125, 255, 0.08);
  }

  .menu-space-item__main,
  .menu-space-binding__main {
    min-width: 0;
    flex: 1;
  }

  .menu-space-item__title-row,
  .menu-space-binding__title-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .menu-space-item__title,
  .menu-space-binding__host {
    font-size: 15px;
    font-weight: 600;
    color: var(--art-gray-900);
  }

  .menu-space-item__meta,
  .menu-space-binding__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 14px;
    margin-top: 8px;
    font-size: 12px;
    color: var(--art-gray-500);
  }

  .menu-space-item__desc {
    margin-top: 10px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--art-gray-700);
  }

  .menu-space-item__hosts {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }

  .menu-space-chip {
    padding: 4px 10px;
    font-size: 12px;
    color: #2156d8;
    background: rgba(55, 125, 255, 0.1);
    border-radius: 999px;
  }

  .menu-space-chip.is-soft {
    color: var(--art-gray-500);
    background: rgba(148, 163, 184, 0.12);
  }

  .menu-space-item__actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 4px;
  }

  .menu-space-empty {
    padding: 28px 16px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--art-gray-500);
    text-align: center;
    background: rgba(148, 163, 184, 0.08);
    border-radius: 16px;
  }

  .menu-space-overview {
    margin-bottom: 14px;
    padding: 14px 16px;
    background: rgba(55, 125, 255, 0.05);
    border: 1px solid rgba(55, 125, 255, 0.12);
    border-radius: 14px;
  }

  .menu-space-overview__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .menu-space-overview__item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .menu-space-overview__label {
    font-size: 12px;
    color: var(--art-gray-500);
  }

  .menu-space-overview__item strong {
    font-size: 14px;
    color: var(--art-gray-900);
    word-break: break-all;
  }

  .menu-space-overview__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }

  .menu-space-drawer-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
  }

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .field-hint {
    margin-top: 8px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--art-gray-500);
  }

  @media (max-width: 1080px) {
    .menu-space-board {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 640px) {
    .menu-space-drawer-grid {
      grid-template-columns: 1fr;
    }

    .menu-space-overview__grid {
      grid-template-columns: 1fr;
    }
  }
</style>
