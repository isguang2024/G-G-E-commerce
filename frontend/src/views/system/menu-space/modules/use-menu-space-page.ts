/**
 * menu-space 视图 mega-composable。
 *
 * 抽离自 index.vue：所有 reactive state、computed、watch、lifecycle 和 handler
 * 都集中在此处，index.vue 只负责解构并转发到模板。
 */
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { HttpError } from '@/utils/http/error'
import {
  fetchGetApps,
  fetchGetCurrentMenuSpace,
  fetchGetMenuSpaceMode,
  fetchInitializeMenuSpaceFromDefault,
  fetchGetMenuSpaceHostBindings,
  fetchGetMenuSpaces,
  fetchGetRuntimeNavigation,
  fetchSaveMenuSpace,
  fetchSaveMenuSpaceHostBinding,
  fetchUpdateMenuSpaceMode
} from '@/domains/governance/api'
import { useManagedAppScope } from '@/domains/app-runtime/useManagedAppScope'
import { buildMenuSpaceQuery, normalizeMenuSpaceKey } from '@/domains/navigation/utils/menu-space'
import {
  collectMenuPaths,
  getAccessModeLabel,
  getAccessModeSummary,
  getHostAuthModeLabel,
  isSpaceInitialized,
  normalizeInternalPath,
  normalizeRoleCodeListText,
  warnDev
} from './helpers'

type HostBindingMetaForm = NonNullable<Api.SystemManage.MenuSpaceHostBindingSaveParams['meta']>
type HostBindingSaveForm = Omit<Api.SystemManage.MenuSpaceHostBindingSaveParams, 'meta'> & {
  meta: HostBindingMetaForm
}

export function useMenuSpacePage() {
  const router = useRouter()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const managedAppMissingText = '请选择当前要管理的 App'

  const loading = ref(false)
  const loadError = ref('')
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const savingSpace = ref(false)
  const savingHost = ref(false)
  const savingSpaceMode = ref(false)
  const initializingSpaceKey = ref('')
  const spaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const hostBindings = ref<Api.SystemManage.MenuSpaceHostBindingItem[]>([])
  const currentSpaceKey = ref('')
  const spaceMode = ref<'single' | 'multi'>('single')
  const currentResolvedBy = ref('')
  const currentRequestHost = ref('')
  const currentAccessGranted = ref(true)
  const loadingLandingPaths = ref(false)
  const landingPathOptions = ref<string[]>([])
  const spaceDrawerVisible = ref(false)
  const hostDrawerVisible = ref(false)
  const editingSpaceKey = ref('')
  const editingHost = ref('')

  const spaceFormRef = ref()
  const hostFormRef = ref()

  // fieldErrors: 后端 Error.details.<field> 回显容器；规范见 docs/guides/frontend-observability-spec.md §2.4
  const spaceFieldErrors = reactive<Record<string, string>>({})
  const hostFieldErrors = reactive<Record<string, string>>({})
  const spaceFormRules: FormRules = {
    name: [{ required: true, message: '请输入空间名称', trigger: 'blur' }],
    menu_space_key: [
      { required: true, message: '请输入空间标识', trigger: 'blur' },
      { pattern: /^[a-z0-9][a-z0-9-_]*$/, message: '仅允许小写字母数字和 - _', trigger: 'blur' }
    ]
  }
  const hostFormRules: FormRules = {
    host: [{ required: true, message: '请输入 Host', trigger: 'blur' }],
    menu_space_key: [{ required: true, message: '请选择导航空间', trigger: 'change' }]
  }
  function clearFieldErrors(target: Record<string, string>) {
    for (const k of Object.keys(target)) delete target[k]
  }
  function applyBackendFieldErrors(target: Record<string, string>, e: unknown): boolean {
    if (!(e instanceof HttpError)) return false
    const data = (e.data || {}) as { details?: Record<string, string> }
    const details = data.details
    if (!details || typeof details !== 'object') return false
    let applied = false
    for (const [field, reason] of Object.entries(details)) {
      if (typeof reason === 'string') {
        target[field] = reason
        applied = true
      }
    }
    return applied
  }

  const spaceForm = reactive<Api.SystemManage.MenuSpaceSaveParams>({
    app_key: '',
    menu_space_key: '',
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
    app_key: '',
    host: '',
    menu_space_key: '',
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

  const currentSpace = computed(() =>
    spaces.value.find((item) => item.menuSpaceKey === currentSpaceKey.value)
  )

  const currentSpaceLabel = computed(
    () => currentSpace.value?.name || currentSpace.value?.menuSpaceKey || '未选择空间'
  )
  const spaceModeLabel = computed(() => {
    if (spaceMode.value === 'single') {
      return '单空间'
    }
    return '多空间'
  })
  const spaceModeTagType = computed(() =>
    spaceModeLabel.value === '单空间' ? 'success' : 'warning'
  )
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
    { label: '当前 App', value: targetAppKey.value },
    { label: '导航空间', value: spaces.value.length || 0 },
    { label: 'Host 绑定', value: hostBindings.value.length || 0 },
    {
      label: '已初始化',
      value: spaces.value.filter((item) => isSpaceInitialized(item)).length || 0
    },
    { label: '当前解析', value: currentSpace.value?.menuSpaceKey || '未选择' }
  ])

  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )

  const spaceOptions = computed(() =>
    spaces.value.map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.menuSpaceKey
    }))
  )

  const spaceDrawerTitle = computed(() =>
    editingSpaceKey.value ? '编辑导航空间' : '新增导航空间'
  )
  const hostDrawerTitle = computed(() => (editingHost.value ? '编辑 Host 绑定' : '新增 Host 绑定'))
  const landingPathHint = computed(() => {
    if (loadingLandingPaths.value) {
      return '正在加载当前空间下可用的入口路径。'
    }
    const value = `${spaceForm.default_home_path || ''}`.trim()
    if (!value) {
      return '未填写时，系统不会自动补默认首页，需要在后续显式配置。'
    }
    if (!value.startsWith('/')) {
      return '默认首页必须是以 / 开头的站内路径，例如 /dashboard/console。'
    }
    if (!landingPathOptions.value.length) {
      return '当前空间下还没有可选入口路径，可以先留空，等导航树与受管页面配置完成后再回填。'
    }
    if (!landingPathOptions.value.includes(value)) {
      return '当前填写的路径不在这个导航空间的已注册入口里，保存前建议先确认导航树或独立页暴露是否已经归属到该空间。'
    }
    return '该路径已命中当前导航空间的可选页面，登录后和进入根路径时会优先跳到这里。'
  })

  async function loadData() {
    loading.value = true
    loadError.value = ''
    if (!targetAppKey.value) {
      spaces.value = []
      hostBindings.value = []
      currentSpaceKey.value = ''
      loadError.value = managedAppMissingText
      loading.value = false
      return
    }
    try {
      const [spaceRes, hostRes, currentRes, modeRes] = await Promise.all([
        fetchGetMenuSpaces(targetAppKey.value),
        fetchGetMenuSpaceHostBindings(targetAppKey.value),
        fetchGetCurrentMenuSpace(undefined, targetAppKey.value).catch(() => undefined),
        fetchGetMenuSpaceMode(targetAppKey.value).catch(() => ({ mode: 'single' }))
      ])
      spaces.value = spaceRes.records || []
      hostBindings.value = hostRes.records || []
      currentSpaceKey.value = currentRes?.space?.menuSpaceKey || ''
      spaceMode.value = `${modeRes?.mode || 'single'}`.trim() === 'multi' ? 'multi' : 'single'
      currentResolvedBy.value = `${currentRes?.resolvedBy || ''}`.trim()
      currentRequestHost.value = `${currentRes?.requestHost || ''}`.trim()
      currentAccessGranted.value = Boolean(currentRes?.accessGranted ?? true)
    } catch (error: any) {
      spaces.value = []
      hostBindings.value = []
      currentSpaceKey.value = ''
      spaceMode.value = 'single'
      currentResolvedBy.value = ''
      currentRequestHost.value = ''
      currentAccessGranted.value = true
      loadError.value = error?.message || '导航空间数据暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  async function saveSpaceMode() {
    if (!targetAppKey.value) {
      loadError.value = managedAppMissingText
      ElMessage.warning(managedAppMissingText)
      return
    }
    savingSpaceMode.value = true
    try {
      const res = await fetchUpdateMenuSpaceMode(targetAppKey.value, spaceMode.value)
      spaceMode.value = `${res?.mode || 'single'}`.trim() === 'multi' ? 'multi' : 'single'
      ElMessage.success(`导航空间模式已更新为${spaceModeLabel.value}`)
      await loadData()
    } catch (error: any) {
      ElMessage.error(error?.message || '导航空间模式保存失败')
    } finally {
      savingSpaceMode.value = false
    }
  }

  async function loadLandingPathCandidates(spaceKey: string) {
    const normalizedSpaceKey = normalizeMenuSpaceKey(spaceKey)
    if (!normalizedSpaceKey) {
      landingPathOptions.value = []
      return
    }
    loadingLandingPaths.value = true
    try {
      const manifest = await fetchGetRuntimeNavigation(normalizedSpaceKey, targetAppKey.value)
      const pagePaths = (manifest.managedPages || [])
        .filter(
          (item) =>
            item.status === 'normal' &&
            item.pageType !== 'group' &&
            item.pageType !== 'display_group' &&
            normalizeInternalPath(`${item.routePath || ''}`.trim())
        )
        .map((item) => normalizeInternalPath(`${item.routePath || ''}`.trim()))
      landingPathOptions.value = Array.from(
        new Set([...collectMenuPaths(manifest.menuTree || []), ...pagePaths])
      ).sort((a, b) => a.localeCompare(b, 'zh-CN'))
    } catch (error) {
      warnDev('load_landing_paths_failed', { err: error })
      landingPathOptions.value = []
    } finally {
      loadingLandingPaths.value = false
    }
  }

  function resetSpaceForm() {
    editingSpaceKey.value = ''
    spaceForm.menu_space_key = ''
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
    hostForm.menu_space_key = currentSpaceKey.value || ''
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
      editingSpaceKey.value = item.menuSpaceKey
      spaceForm.menu_space_key = item.menuSpaceKey
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
    loadLandingPathCandidates(spaceForm.menu_space_key || '')
  }

  function openHostDrawer(
    item?: Api.SystemManage.MenuSpaceHostBindingItem,
    preferredSpaceKey?: string
  ) {
    resetHostForm()
    if (item) {
      editingHost.value = item.host
      hostForm.host = item.host
      hostForm.menu_space_key = item.menuSpaceKey
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
      hostForm.menu_space_key = normalizeMenuSpaceKey(preferredSpaceKey)
    }
    hostDrawerVisible.value = true
  }

  async function saveSpace() {
    clearFieldErrors(spaceFieldErrors)
    const valid = await spaceFormRef.value?.validate().catch(() => false)
    if (!valid) return
    if (!spaceForm.name.trim()) {
      spaceFieldErrors.name = '请输入空间名称'
      return
    }
    if (!`${spaceForm.menu_space_key || ''}`.trim()) {
      spaceFieldErrors.menu_space_key = '请输入空间标识'
      return
    }
    const normalizedHomePath = normalizeInternalPath(spaceForm.default_home_path || '')
    if (spaceForm.default_home_path?.trim() && !normalizedHomePath) {
      spaceFieldErrors.default_home_path = '默认首页必须是以 / 开头的站内路径'
      return
    }
    if (
      normalizedHomePath &&
      landingPathOptions.value.length > 0 &&
      !landingPathOptions.value.includes(normalizedHomePath)
    ) {
      ElMessage.warning('默认首页未命中当前导航空间的已注册页面，请先确认导航或页面归属')
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
        app_key: targetAppKey.value,
        menu_space_key: normalizeMenuSpaceKey(spaceForm.menu_space_key),
        name: spaceForm.name.trim(),
        description: spaceForm.description?.trim() || '',
        default_home_path: normalizedHomePath,
        access_mode: spaceForm.access_mode || 'all',
        allowed_role_codes: spaceForm.access_mode === 'role_codes' ? allowedRoleCodes : []
      })
      ElMessage.success('导航空间已保存')
      spaceDrawerVisible.value = false
      await loadData()
    } catch (error: any) {
      if (applyBackendFieldErrors(spaceFieldErrors, error)) return
      ElMessage.error(error?.message || '导航空间保存失败')
    } finally {
      savingSpace.value = false
    }
  }

  async function saveHostBinding() {
    clearFieldErrors(hostFieldErrors)
    const valid = await hostFormRef.value?.validate().catch(() => false)
    if (!valid) return
    if (!hostForm.host.trim()) {
      hostFieldErrors.host = '请输入 Host'
      return
    }
    if (!`${hostForm.menu_space_key || ''}`.trim()) {
      hostFieldErrors.menu_space_key = '请选择导航空间'
      return
    }
    const normalizedHost = `${hostForm.host || ''}`.trim().toLowerCase()
    const duplicatedBinding = hostBindings.value.find(
      (item) =>
        `${item.host || ''}`.trim().toLowerCase() === normalizedHost &&
        normalizeMenuSpaceKey(item.menuSpaceKey) !== normalizeMenuSpaceKey(hostForm.menu_space_key)
    )
    if (duplicatedBinding) {
      hostFieldErrors.host = `该 Host 已绑定到导航空间 ${duplicatedBinding.spaceName || duplicatedBinding.menuSpaceKey}`
      return
    }
    savingHost.value = true
    try {
      await fetchSaveMenuSpaceHostBinding({
        ...hostForm,
        app_key: targetAppKey.value,
        host: hostForm.host.trim(),
        menu_space_key: normalizeMenuSpaceKey(hostForm.menu_space_key),
        description: hostForm.description?.trim() || '',
        meta: {
          ...hostForm.meta,
          scheme: `${hostForm.meta?.scheme || 'https'}`.trim() || 'https',
          route_prefix: `${hostForm.meta?.route_prefix || ''}`.trim(),
          auth_mode: `${hostForm.meta?.auth_mode || 'inherit_host'}`.trim() || 'inherit_host',
          login_host: `${hostForm.meta?.login_host || ''}`.trim(),
          callback_host: `${hostForm.meta?.callback_host || ''}`.trim(),
          cookie_scope_mode:
            `${hostForm.meta?.cookie_scope_mode || 'inherit'}`.trim() || 'inherit',
          cookie_domain: `${hostForm.meta?.cookie_domain || ''}`.trim()
        }
      })
      ElMessage.success('Host 绑定已保存')
      hostDrawerVisible.value = false
      await loadData()
    } catch (error: any) {
      if (applyBackendFieldErrors(hostFieldErrors, error)) return
      ElMessage.error(error?.message || 'Host 绑定保存失败')
    } finally {
      savingHost.value = false
    }
  }

  async function loadAppOptions() {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }

  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
    currentSpaceKey.value = ''
  }

  async function initializeSpace(item: Api.SystemManage.MenuSpaceItem) {
    if (!item?.menuSpaceKey || item.isDefault) {
      return
    }
    if (isSpaceInitialized(item)) {
      ElMessage.info('当前导航空间已经初始化，可直接进入导航管理或受管页面继续调整')
      return
    }
    initializingSpaceKey.value = item.menuSpaceKey
    try {
      const result = await fetchInitializeMenuSpaceFromDefault(targetAppKey.value, item.menuSpaceKey)
      ElMessage.success(
        `已完成初始化：复制 ${result.createdMenuCount} 个导航、同步 ${result.createdPackageMenuLinkCount} 条功能包导航关联，独立页暴露 ${result.createdPageCount || 0} 项`
      )
      await loadData()
      goToMenuManagement(item.menuSpaceKey)
    } catch (error: any) {
      ElMessage.error(error?.message || '复制默认空间导航失败')
    } finally {
      initializingSpaceKey.value = ''
    }
  }

  async function reinitializeSpace(item: Api.SystemManage.MenuSpaceItem) {
    if (!item?.menuSpaceKey || item.isDefault || !isSpaceInitialized(item)) {
      return
    }
    try {
      await ElMessageBox.confirm(
        `重新初始化会清空空间“${item.name}”当前已有的导航树、独立页暴露和功能包导航关联，然后重新复制默认空间内容。共享受管页面定义不会被复制，只会重新计算空间暴露。`,
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
    initializingSpaceKey.value = item.menuSpaceKey
    try {
      const result = await fetchInitializeMenuSpaceFromDefault(
        targetAppKey.value,
        item.menuSpaceKey,
        true
      )
      ElMessage.success(
        `已重新初始化：清空 ${result.clearedMenuCount || 0} 个导航、${result.clearedPageCount || 0} 项独立页暴露、${result.clearedPackageMenuLinkCount || 0} 条功能包导航关联，并重新复制 ${result.createdMenuCount} 个导航`
      )
      await loadData()
      goToMenuManagement(item.menuSpaceKey)
    } catch (error: any) {
      ElMessage.error(error?.message || '重新初始化失败')
    } finally {
      initializingSpaceKey.value = ''
    }
  }

  function goToMenuManagement(spaceKey: string) {
    router.push({
      path: '/system/menu',
      query: {
        ...buildMenuSpaceQuery(normalizeMenuSpaceKey(spaceKey)),
        layout: '1'
      }
    })
  }

  function goToPageManagement(spaceKey: string) {
    router.push({
      path: '/system/page',
      query: {
        ...buildMenuSpaceQuery(normalizeMenuSpaceKey(spaceKey))
      }
    })
  }

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
    loadAppOptions().catch(() => {
      appList.value = []
    })
    if (!targetAppKey.value) {
      loadError.value = managedAppMissingText
      return
    }
    loadData()
  })

  watch(
    () => targetAppKey.value,
    (value) => {
      selectedAppKey.value = value || ''
      if (!targetAppKey.value) {
        spaces.value = []
        hostBindings.value = []
        currentSpaceKey.value = ''
        loadError.value = managedAppMissingText
        return
      }
      loadData()
    }
  )

  watch(
    () => spaceForm.menu_space_key || '',
    (value, previousValue) => {
      if (!spaceDrawerVisible.value) return
      if (!`${value || ''}`.trim() || value === previousValue) return
      if (
        spaceForm.default_home_path &&
        !landingPathOptions.value.includes(spaceForm.default_home_path)
      ) {
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

  return {
    // state
    loading,
    loadError,
    selectedAppKey,
    savingSpace,
    savingHost,
    savingSpaceMode,
    initializingSpaceKey,
    spaces,
    hostBindings,
    currentSpaceKey,
    spaceMode,
    currentRequestHost,
    landingPathOptions,
    spaceDrawerVisible,
    hostDrawerVisible,
    spaceFormRef,
    hostFormRef,
    spaceForm,
    spaceFormRules,
    spaceFieldErrors,
    hostFormRules,
    hostFieldErrors,
    allowedRoleCodesText,
    hostForm,
    // computed
    currentSpace,
    currentSpaceLabel,
    spaceModeLabel,
    spaceModeTagType,
    resolveByLabel,
    summaryMetrics,
    appOptions,
    spaceOptions,
    spaceDrawerTitle,
    hostDrawerTitle,
    landingPathHint,
    // actions / handlers
    loadData,
    saveSpaceMode,
    openSpaceDrawer,
    openHostDrawer,
    saveSpace,
    saveHostBinding,
    handleManagedAppChange,
    initializeSpace,
    reinitializeSpace,
    goToMenuManagement,
    goToPageManagement,
    // pure helpers re-exported for template
    isSpaceInitialized,
    getAccessModeLabel,
    getAccessModeSummary,
    getHostAuthModeLabel
  }
}
