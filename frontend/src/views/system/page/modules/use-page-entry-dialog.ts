/**
 * page-entry-dialog 视图脚本：所有 reactive state、computed、watch、handler 集中在此。
 *
 * 抽离自 page-entry-dialog.vue，.vue 文件保留 defineProps/defineEmits/defineOptions/defineExpose
 * 等编译宏与 template/style 块，调用本 composable 拉取所有模板绑定。
 */
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  fetchCreatePage,
  fetchGetPageMenuOptions,
  fetchGetPageOptions,
  fetchUpdatePage
} from '@/domains/governance/api'
import {
  isSingleSegmentManagedPagePath,
  joinManagedPagePath,
  normalizeManagedPagePath,
  resolveManagedPageRoutePath
} from '@/domains/navigation/utils/managed-page'
import { normalizeMenuId, toTreeSelectNode } from './helpers'

type PageItem = Api.SystemManage.PageItem
type PageMenuOptionItem = Api.SystemManage.PageMenuOptionItem

export interface UsePageEntryDialogProps {
  modelValue: boolean
  dialogType: 'add' | 'edit' | 'copy'
  pageData?: Partial<PageItem>
  appKey?: string
  menuSpaces?: Api.SystemManage.MenuSpaceItem[]
  initialParentPageKey?: string
  initialParentMenuId?: string
  initialPageType?: PageItem['pageType']
  defaultData?: Partial<PageItem>
}

export interface UsePageEntryDialogEmit {
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}

export function usePageEntryDialog(props: UsePageEntryDialogProps, emit: UsePageEntryDialogEmit) {
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const menuOptions = ref<PageMenuOptionItem[]>([])
  const allPages = ref<PageItem[]>([])
  const mountMode = ref<'none' | 'menu' | 'page'>('none')
  const mountSpaceKey = ref<string>('')
  const showAdvanced = ref(false)
  const showExamples = ref(false)
  const isInitializing = ref(false)
  const isInternalRoutePathChange = ref(false)
  const hasManualRoutePathEdit = ref(false)
  const originalRouteContext = ref({
    routePath: '',
    pageKey: '',
    parentMenuId: '',
    parentPageKey: '',
    mountMode: 'none' as 'none' | 'menu' | 'page'
  })

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const dialogTitle = computed(() => {
    if (props.dialogType === 'copy') {
      return '复制页面'
    }
    const actionText = props.dialogType === 'add' ? '新增' : '编辑'
    if (form.pageType === 'standalone') return `${actionText}全局页`
    if (form.pageType === 'standalone') return `${actionText}独立页`
    return `${actionText}页面`
  })

  const form = reactive({
    id: '',
    pageKey: '',
    name: '',
    routeName: '',
    routePath: '',
    component: '',
    pageType: 'standalone',
    visibilityScope: 'app' as 'inherit' | 'app' | 'spaces',
    moduleKey: '',
    spaceKeys: [] as string[],
    sortOrder: 0,
    parentMenuId: '',
    parentPageKey: '',
    displayGroupKey: '',
    activeMenuPath: '',
    breadcrumbMode: 'inherit_menu',
    accessMode: 'inherit',
    permissionKey: '',
    keepAlive: false,
    isFullPage: false,
    isIframe: false,
    link: '',
    status: 'normal'
  })

  const rules = reactive<FormRules>({
    pageKey: [{ required: true, message: '请输入页面标识', trigger: 'blur' }],
    name: [{ required: true, message: '请输入页面名称', trigger: 'blur' }],
    routeName: [{ validator: validateRouteName, trigger: 'blur' }],
    routePath: [{ validator: validateRoutePath, trigger: 'blur' }],
    component: [{ validator: validateComponent, trigger: 'blur' }],
    link: [{ validator: validateLink, trigger: 'blur' }],
    parentMenuId: [{ validator: validateParentBinding, trigger: 'change' }],
    parentPageKey: [{ validator: validateParentBinding, trigger: 'change' }],
    permissionKey: [{ validator: validatePermissionKey, trigger: 'blur' }]
  })

  const menuTreeOptions = computed(() => menuOptions.value.map(toTreeSelectNode))
  const menuSpaceOptions = computed(() =>
    (props.menuSpaces || []).map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )
  const showMountSection = computed(() => form.pageType === 'inner')
  const showVisibilityScopeField = computed(
    () => form.pageType === 'standalone' || form.pageType === 'standalone'
  )
  const showSpaceBindingField = computed(
    () => showVisibilityScopeField.value && form.visibilityScope === 'spaces'
  )
  const resolveSpaceBindingKeys = () => {
    if (!showSpaceBindingField.value) return []
    return form.spaceKeys.map((item) => `${item || ''}`.trim()).filter(Boolean)
  }
  const resolveSpaceScopeKey = () => {
    const values = resolveSpaceBindingKeys()
    return values[0] || ''
  }
  const menuCascaderProps = {
    checkStrictly: true,
    emitPath: false
  }

  const parentPageOptions = computed(() =>
    allPages.value.filter(
      (item) => item.id !== form.id && `${item.pageType || ''}`.trim() !== 'display_group'
    )
  )
  const displayGroupOptions = computed(() =>
    allPages.value.filter(
      (item) =>
        item.id !== form.id &&
        `${item.pageType || ''}`.trim() === 'display_group' &&
        item.status === 'normal'
    )
  )

  const pageMap = computed(() => {
    const map = new Map<string, PageItem>()
    allPages.value.forEach((item) => {
      if (item.pageKey) {
        map.set(item.pageKey, item)
      }
    })
    return map
  })

  const menuPathMap = computed(() => {
    const map = new Map<string, string>()
    const walk = (items: PageMenuOptionItem[], parentPath = '') => {
      items.forEach((item) => {
        const fullPath = joinManagedPagePath(parentPath, item.path)
        if (item.id) {
          map.set(item.id, fullPath)
        }
        if (Array.isArray(item.children) && item.children.length) {
          walk(item.children, fullPath)
        }
      })
    }
    walk(menuOptions.value)
    return map
  })

  const menuNameMap = computed(() => {
    const map = new Map<string, string>()
    const walk = (items: PageMenuOptionItem[]) => {
      items.forEach((item) => {
        if (item.id) {
          map.set(item.id, item.title || item.name || item.path || item.id)
        }
        if (Array.isArray(item.children) && item.children.length) {
          walk(item.children)
        }
      })
    }
    walk(menuOptions.value)
    return map
  })

  const configHintTitle = computed(() =>
    form.pageType === 'standalone'
      ? '全局页配置说明'
      : form.pageType === 'standalone'
        ? '独立页配置说明'
        : '页面配置说明'
  )
  const isUnregisteredCandidate = computed(
    () => props.dialogType === 'add' && Boolean(props.defaultData?.meta?.fromUnregistered)
  )
  const isComponentLocked = computed(() => isUnregisteredCandidate.value && !form.isIframe)
  const shouldAutoAdjustRoutePath = computed(
    () => props.dialogType === 'edit' || isUnregisteredCandidate.value
  )

  const configHintDescription = computed(() => {
    if (form.pageType === 'standalone') {
      return '全局页属于独立页面，不依赖菜单归属，可在当前 App 下全局可见，也可以只对指定空间开放。'
    }
    if (form.pageType === 'standalone') {
      return '独立页不挂菜单也不挂父页面，可在当前 App 全局开放，或只向选定菜单空间暴露。'
    }
    return '内页必须挂到菜单或上级页面，并自动继承其路径链、菜单高亮与默认准入。'
  })

  const mountOwnershipSummary = computed(() => {
    if (form.pageType === 'standalone') {
      return form.visibilityScope === 'spaces'
        ? '当前页面属于全局页，仅在选定菜单空间开放，不占用左侧菜单入口。'
        : '当前页面属于全局页，在当前 App 下全局可见，不占用左侧菜单入口。'
    }
    if (form.pageType === 'standalone') {
      return form.visibilityScope === 'spaces'
        ? '当前页面是独立页，仅在选定菜单空间开放。'
        : '当前页面是独立页，在当前 App 下全局可见。'
    }
    if (mountMode.value === 'menu') {
      return '当前页面会挂到菜单下，菜单负责入口可见、默认高亮和默认准入。'
    }
    if (mountMode.value === 'page') {
      return '当前页面会挂到上级页面或逻辑分组下，优先继承其路径链、菜单链和默认面包屑。'
    }
    return '当前页面会作为内页存在，必须通过菜单或父页面进入，运行时默认继承上级链路。'
  })

  const accessModeOptions = computed(() => {
    const options = [
      { label: '继承', value: 'inherit' },
      { label: '公开', value: 'public' },
      { label: '登录', value: 'jwt' },
      { label: '权限', value: 'permission' }
    ]
    if (form.pageType === 'standalone') {
      return options.filter((item) => item.value !== 'inherit')
    }
    if (form.pageType === 'standalone') {
      return options.filter((item) => item.value !== 'inherit')
    }
    return options
  })

  const routePathPlaceholder = computed(() => {
    if (form.pageType === 'standalone') {
      return '例如 /store-management'
    }
    if (form.pageType === 'standalone') {
      return '例如 /workspace/profile 或 /report/detail/:id'
    }
    if (mountMode.value === 'none') {
      return '例如 /dashboard/example-page 或 /detail/:id'
    }
    return '例如 detail 或 detail/:id；如需绝对路径可填 /system/detail'
  })

  const mountMenuSummary = computed(() => {
    if (form.pageType === 'standalone' || mountMode.value !== 'menu') return ''
    const menuName = menuNameMap.value.get(normalizeMenuId(form.parentMenuId)) || '所选菜单'
    const permissionText =
      form.accessMode === 'permission'
        ? '当前页面单独配置了权限，最终会按“菜单准入 + 页面权限”交集放行。'
        : '当前页面走继承模式，最终默认跟随菜单权限。'
    return `页面将挂到“${menuName}”下，菜单负责入口可见与默认准入。${permissionText}`
  })

  const menuSiblingSummary = computed(() => {
    if (form.pageType === 'standalone' || mountMode.value !== 'menu') return ''
    const menuId = normalizeMenuId(form.parentMenuId)
    if (!menuId) return ''
    const siblings = allPages.value.filter(
      (item) => item.id !== form.id && `${item.parentMenuId || ''}`.trim() === menuId
    )
    if (siblings.length === 0) {
      return '当前菜单下还没有其他挂接页面。'
    }
    const names = siblings.slice(0, 4).map((item) => item.name)
    const suffix = siblings.length > 4 ? ` 等 ${siblings.length} 个页面` : ''
    return `当前菜单下还有 ${siblings.length} 个页面：${names.join('、')}${suffix}。菜单入口页请直接在菜单管理维护；这里仅负责该入口下的受管页面。`
  })

  const resolvedRoutePreview = computed(() =>
    resolveManagedPageRoutePath(
      {
        ...form,
        source: props.pageData?.source || props.defaultData?.source || 'manual',
        parentMenuId:
          form.pageType === 'standalone' || mountMode.value !== 'menu'
            ? ''
            : normalizeMenuId(form.parentMenuId),
        parentPageKey:
          form.pageType === 'standalone' || mountMode.value !== 'page' ? '' : form.parentPageKey
      },
      {
        getPageByKey: (pageKey) => pageMap.value.get(pageKey),
        getMenuPathById: (menuId) => menuPathMap.value.get(menuId)
      }
    )
  )

  const routePreviewHint = computed(() => {
    if (form.pageType === 'standalone') {
      return '全局页使用独立路径，不参与上级菜单或页面链路拼接。'
    }
    if (mountMode.value === 'none') {
      return '不挂载时页面独立存在，不继承菜单或页面路径。建议直接填写完整路径；若填写单段路径，系统会自动补成 /xxx。'
    }
    if (mountMode.value === 'page') {
      return '挂到页面时，单段路径会自动拼到上级页面路径后；多段绝对路径仍按你填写的完整路径注册。'
    }
    return '挂到菜单时，单段路径会自动拼到菜单路径后；访问上默认继承菜单权限，如页面再单独配置权限，则最终按菜单权限与页面权限交集放行。'
  })

  const pageExamples = computed(() => {
    if (form.pageType === 'standalone') {
      return [
        '例 1：页面类型=全局页，可见范围=当前 App 全局，路由路径=/workspace/overview。适合登录后可直达页面。',
        '例 2：页面类型=全局页，可见范围=指定空间，勾选 ops / finance 后，仅这些空间下会暴露该页面。',
        '例 3：如果是外部系统页，可开启“是否内嵌”，外链地址填 https://example.com。'
      ]
    }
    if (form.pageType === 'standalone') {
      return [
        '例 1：页面类型=独立页，可见范围=当前 App 全局，路由路径=/workspace/profile。适合个人设置、消息中心等公共页。',
        '例 2：页面类型=独立页，可见范围=指定空间，勾选 ops / finance 后，仅这些空间下会暴露该页面。',
        '例 3：独立页不挂菜单也不挂父页面，路径就是你填写的完整路径。'
      ]
    }
    if (mountMode.value === 'none') {
      return [
        '例 1：挂载方式=不挂载，路由路径=/dashboard/example-page，最终路径就是 /dashboard/example-page。',
        '例 2：挂载方式=不挂载，路由路径=/report/overview，页面会作为独立内页存在，不参与菜单高亮继承。',
        '例 3：个人中心、结果页、引导页这类不需要出现在菜单中的页面，适合使用独立内页。'
      ]
    }
    if (mountMode.value === 'page') {
      return [
        '例 1：上级页面最终路径=/dashboard/console，路由路径=analysis，最终路径=/dashboard/console/analysis。',
        '例 2：上级页面最终路径=/dashboard/console/analysis，路由路径=trend，最终路径=/dashboard/console/analysis/trend。',
        '例 3：如果你直接填写 /dashboard/report/detail 这种多段绝对路径，则按该完整路径注册，不再继承上级路径。'
      ]
    }
    return [
      '例 1：上级菜单路径=/dashboard，路由路径=console，最终路径=/dashboard/console。',
      '例 2：上级菜单路径=/dashboard，路由路径=/console 也会按单段相对路径处理，最终仍是 /dashboard/console。',
      '例 3：如果要固定成独立地址，请直接填写 /dashboard/report/detail 这种多段绝对路径。'
    ]
  })

  function getComponentPlaceholder() {
    if (form.isIframe) {
      return '内嵌模式自动使用 /outside/Iframe'
    }
    if (isComponentLocked.value) {
      return '组件路径固定为未注册页面扫描结果'
    }
    return '例如 /system/page'
  }

  function setRoutePathSilently(value: string) {
    isInternalRoutePathChange.value = true
    form.routePath = value
    nextTick(() => {
      isInternalRoutePathChange.value = false
    })
  }

  function getResolvedPathSnapshot(routePath: string, parentMenuId: string, parentPageKey: string) {
    return resolveManagedPageRoutePath<
      Pick<PageItem, 'pageKey' | 'routePath' | 'parentMenuId' | 'parentPageKey'>
    >(
      {
        pageKey: originalRouteContext.value.pageKey,
        routePath,
        parentMenuId,
        parentPageKey
      },
      {
        getPageByKey: (pageKey) => pageMap.value.get(pageKey),
        getMenuPathById: (menuId) => menuPathMap.value.get(menuId)
      }
    )
  }

  function deriveRelativeRoutePath(routePath: string) {
    const normalizedRoute = normalizeManagedPagePath(routePath)
    if (!normalizedRoute || isSingleSegmentManagedPagePath(routePath)) {
      return ''
    }
    const previousBasePath = getResolvedPathSnapshot(
      '',
      originalRouteContext.value.parentMenuId,
      originalRouteContext.value.parentPageKey
    )
    const previousResolvedPath = getResolvedPathSnapshot(
      originalRouteContext.value.routePath,
      originalRouteContext.value.parentMenuId,
      originalRouteContext.value.parentPageKey
    )
    if (
      previousBasePath &&
      previousResolvedPath &&
      normalizedRoute.startsWith(`${previousBasePath}/`)
    ) {
      return normalizedRoute.slice(previousBasePath.length + 1)
    }
    const segments = normalizedRoute.replace(/^\/+/, '').split('/').filter(Boolean)
    if (segments.length <= 1) {
      return segments[0] || ''
    }
    return segments[segments.length - 1]
  }

  function maybeAdjustRoutePathForMountedEdit() {
    if (
      !shouldAutoAdjustRoutePath.value ||
      form.pageType === 'standalone' ||
      mountMode.value === 'none'
    ) {
      return
    }
    if (hasManualRoutePathEdit.value) {
      return
    }
    const currentRoutePath = `${form.routePath || ''}`.trim()
    const originalRoutePath = `${originalRouteContext.value.routePath || ''}`.trim()
    if (!currentRoutePath || currentRoutePath !== originalRoutePath) {
      return
    }
    if (!currentRoutePath.startsWith('/') || isSingleSegmentManagedPagePath(currentRoutePath)) {
      return
    }
    const previousParentMenuId = originalRouteContext.value.parentMenuId
    const previousParentPageKey = originalRouteContext.value.parentPageKey
    const currentParentMenuId = mountMode.value === 'menu' ? normalizeMenuId(form.parentMenuId) : ''
    const currentParentPageKey =
      mountMode.value === 'page' ? `${form.parentPageKey || ''}`.trim() : ''
    if (
      currentParentMenuId === previousParentMenuId &&
      currentParentPageKey === previousParentPageKey
    ) {
      return
    }
    const relativeRoutePath = deriveRelativeRoutePath(currentRoutePath)
    if (!relativeRoutePath) {
      return
    }
    setRoutePathSilently(relativeRoutePath)
  }

  function validateRouteName(_: unknown, _value: string, callback: (error?: Error) => void) {
    callback()
  }

  function validateRoutePath(_: unknown, value: string, callback: (error?: Error) => void) {
    if (!`${value || ''}`.trim()) {
      callback(new Error('请输入路由路径'))
      return
    }
    callback()
  }

  function validateComponent(_: unknown, value: string, callback: (error?: Error) => void) {
    if (form.isIframe) {
      callback()
      return
    }
    if (!`${value || ''}`.trim()) {
      callback(new Error('请输入组件路径'))
      return
    }
    callback()
  }

  function validateLink(_: unknown, value: string, callback: (error?: Error) => void) {
    if (!form.isIframe) {
      callback()
      return
    }
    const target = `${value || ''}`.trim()
    if (!target) {
      callback(new Error('内嵌模式下需填写外链地址'))
      return
    }
    if (!/^https?:\/\//i.test(target)) {
      callback(new Error('外链地址需以 http:// 或 https:// 开头'))
      return
    }
    callback()
  }

  function validateParentBinding(_: unknown, __: string, callback: (error?: Error) => void) {
    if (form.pageType === 'standalone' || form.pageType === 'standalone') {
      callback()
      return
    }
    if (mountMode.value === 'none') {
      callback(new Error('内页必须挂到菜单或上级页面'))
      return
    }
    if (mountMode.value === 'menu' && !normalizeMenuId(form.parentMenuId)) {
      callback(new Error('请选择上级菜单'))
      return
    }
    if (mountMode.value === 'page' && !`${form.parentPageKey || ''}`.trim()) {
      callback(new Error('请选择上级页面'))
      return
    }
    callback()
  }

  function validatePermissionKey(_: unknown, value: string, callback: (error?: Error) => void) {
    if (form.accessMode !== 'permission') {
      callback()
      return
    }
    if (!`${value || ''}`.trim()) {
      callback(new Error('权限模式下需填写权限键'))
      return
    }
    callback()
  }

  function initForm() {
    mountSpaceKey.value = ''
    if (props.dialogType === 'edit' && props.pageData) {
      const spaceKeys = Array.isArray(props.pageData.spaceKeys) ? props.pageData.spaceKeys : []
      Object.assign(form, {
        id: props.pageData.id || '',
        pageKey: props.pageData.pageKey || '',
        name: props.pageData.name || '',
        routeName: props.pageData.routeName || '',
        routePath: props.pageData.routePath || '',
        component: props.pageData.component || '',
        pageType:
          props.pageData.pageType === 'standalone'
            ? 'standalone'
            : props.pageData.pageType === 'standalone'
              ? 'standalone'
              : 'inner',
        visibilityScope:
          `${props.pageData.visibilityScope || props.pageData.spaceScope || ''}`.trim() === 'spaces'
            ? 'spaces'
            : props.pageData.pageType === 'inner'
              ? 'inherit'
              : 'app',
        moduleKey: props.pageData.moduleKey || '',
        spaceKeys: spaceKeys,
        sortOrder: props.pageData.sortOrder ?? 0,
        parentMenuId: props.pageData.parentMenuId || '',
        parentPageKey: props.pageData.parentPageKey || '',
        displayGroupKey: props.pageData.displayGroupKey || '',
        activeMenuPath: props.pageData.activeMenuPath || '',
        breadcrumbMode: props.pageData.breadcrumbMode || 'inherit_menu',
        accessMode:
          props.pageData.pageType === 'standalone'
            ? props.pageData.accessMode || 'jwt'
            : props.pageData.accessMode || 'inherit',
        permissionKey: props.pageData.permissionKey || '',
        keepAlive: props.pageData.keepAlive ?? false,
        isFullPage: props.pageData.isFullPage ?? false,
        isIframe: Boolean(props.pageData.meta?.isIframe ?? props.pageData.isIframe),
        link: `${props.pageData.meta?.link || props.pageData.link || ''}`.trim(),
        status: props.pageData.status || 'normal'
      })
      mountMode.value =
        form.pageType === 'inner'
          ? form.parentPageKey
            ? 'page'
            : form.parentMenuId
              ? 'menu'
              : 'none'
          : 'none'
      if (form.pageType === 'standalone' || form.pageType === 'standalone') {
        form.accessMode = form.accessMode === 'inherit' ? 'jwt' : form.accessMode
        showAdvanced.value = false
        mountMode.value = 'none'
      } else {
        showAdvanced.value =
          Boolean(form.activeMenuPath) || form.breadcrumbMode !== defaultBreadcrumbMode()
      }
      return
    }

    Object.assign(form, {
      id: '',
      pageKey: props.defaultData?.pageKey || '',
      name: props.defaultData?.name || '',
      routeName: props.defaultData?.routeName || '',
      routePath: props.defaultData?.routePath || '',
      component: props.defaultData?.component || '',
      pageType:
        props.defaultData?.pageType === 'standalone' || props.initialPageType === 'standalone'
          ? 'standalone'
          : props.defaultData?.pageType === 'standalone' || props.initialPageType === 'standalone'
            ? 'standalone'
            : 'inner',
      visibilityScope:
        props.defaultData?.visibilityScope === 'spaces'
          ? 'spaces'
          : props.defaultData?.pageType === 'standalone' || props.initialPageType === 'standalone'
            ? 'app'
            : props.defaultData?.pageType === 'standalone' || props.initialPageType === 'standalone'
              ? 'app'
              : 'inherit',
      moduleKey: props.defaultData?.moduleKey || '',
      spaceKeys: props.defaultData?.spaceKeys || [],
      sortOrder: props.defaultData?.sortOrder ?? 0,
      parentMenuId: props.defaultData?.parentMenuId || props.initialParentMenuId || '',
      parentPageKey: props.defaultData?.parentPageKey || props.initialParentPageKey || '',
      displayGroupKey: props.defaultData?.displayGroupKey || '',
      activeMenuPath: props.defaultData?.activeMenuPath || '',
      breadcrumbMode: 'inherit_menu',
      accessMode:
        props.defaultData?.pageType === 'standalone' || props.initialPageType === 'standalone'
          ? props.defaultData?.accessMode || 'jwt'
          : props.defaultData?.accessMode || 'inherit',
      permissionKey: props.defaultData?.permissionKey || '',
      keepAlive: props.defaultData?.keepAlive ?? false,
      isFullPage: props.defaultData?.isFullPage ?? false,
      isIframe: Boolean(props.defaultData?.meta?.isIframe ?? props.defaultData?.isIframe),
      link: `${props.defaultData?.meta?.link || props.defaultData?.link || ''}`.trim(),
      status: props.defaultData?.status || 'normal'
    })
    mountMode.value =
      form.pageType === 'inner'
        ? form.parentPageKey
          ? 'page'
          : form.parentMenuId
            ? 'menu'
            : 'none'
        : 'none'
    form.breadcrumbMode = defaultBreadcrumbMode()
    showAdvanced.value = form.pageType === 'inner' ? Boolean(form.activeMenuPath) : false
  }

  async function loadOptions() {
    let scopeKey = ''
    if (form.pageType === 'inner') {
      // 内页挂到菜单时，按所选菜单空间过滤候选菜单
      scopeKey = mountMode.value === 'menu' ? mountSpaceKey.value.trim() : ''
    } else if (form.pageType === 'standalone' && form.visibilityScope === 'spaces') {
      scopeKey = resolveSpaceScopeKey()
    }
    const appKey = `${props.appKey || ''}`.trim()
    const [menuRes, pageRes] = await Promise.all([
      fetchGetPageMenuOptions(scopeKey || undefined, appKey),
      fetchGetPageOptions(scopeKey || undefined, appKey)
    ])
    menuOptions.value = menuRes.records || []
    allPages.value = pageRes.records || []
  }

  async function prepareDialog() {
    isInitializing.value = true
    await nextTick()
    initForm()
    await loadOptions()
    originalRouteContext.value = {
      routePath: `${form.routePath || ''}`.trim(),
      pageKey: `${form.pageKey || ''}`.trim(),
      parentMenuId: normalizeMenuId(form.parentMenuId),
      parentPageKey: `${form.parentPageKey || ''}`.trim(),
      mountMode: mountMode.value
    }
    hasManualRoutePathEdit.value = false
    await nextTick()
    formRef.value?.clearValidate()
    isInitializing.value = false
  }

  function defaultBreadcrumbMode() {
    if (mountMode.value === 'page') {
      return 'inherit_page'
    }
    if (mountMode.value === 'menu') {
      return 'inherit_menu'
    }
    return 'custom'
  }

  watch(
    () => props.modelValue,
    async (value) => {
      if (!value) return
      await prepareDialog()
    }
  )

  watch(
    () => [
      props.dialogType,
      props.pageData,
      props.defaultData,
      props.initialParentPageKey,
      props.initialParentMenuId,
      props.initialPageType
    ],
    () => {
      if (!props.modelValue) return
      nextTick(() => {
        initForm()
        formRef.value?.clearValidate()
      })
    },
    { deep: true }
  )

  watch(
    () => mountMode.value,
    async () => {
      if (form.pageType !== 'inner') {
        form.parentMenuId = ''
        form.parentPageKey = ''
      } else if (mountMode.value === 'menu') {
        form.parentPageKey = ''
      } else if (mountMode.value === 'page') {
        form.parentMenuId = ''
        form.displayGroupKey = ''
      } else {
        form.parentMenuId = ''
        form.parentPageKey = ''
      }
      if (!showAdvanced.value) {
        form.breadcrumbMode = defaultBreadcrumbMode()
      }
      if (!isInitializing.value && form.pageType === 'inner') {
        await loadOptions()
      }
      maybeAdjustRoutePathForMountedEdit()
      nextTick(() =>
        formRef.value?.validateField(['parentMenuId', 'parentPageKey']).catch(() => undefined)
      )
    }
  )

  watch(
    () => mountSpaceKey.value,
    async () => {
      if (isInitializing.value) return
      if (form.pageType !== 'inner' || mountMode.value !== 'menu') return
      form.parentMenuId = ''
      await loadOptions()
    }
  )

  watch(
    () => form.parentMenuId,
    () => {
      maybeAdjustRoutePathForMountedEdit()
    }
  )

  watch(
    () => form.parentPageKey,
    () => {
      maybeAdjustRoutePathForMountedEdit()
    }
  )

  watch(
    () => form.routePath,
    () => {
      if (isInitializing.value || isInternalRoutePathChange.value) {
        return
      }
      hasManualRoutePathEdit.value = true
    }
  )

  watch(
    () => form.pageType,
    (pageType) => {
      if (pageType === 'standalone' || pageType === 'standalone') {
        mountMode.value = 'none'
        form.parentMenuId = ''
        form.parentPageKey = ''
        form.activeMenuPath = ''
        form.breadcrumbMode = 'inherit_menu'
        if (form.accessMode === 'inherit') {
          form.accessMode = 'jwt'
        }
        if (pageType === 'standalone') {
          form.visibilityScope = form.visibilityScope === 'spaces' ? 'spaces' : 'app'
        }
      } else {
        form.visibilityScope = 'inherit'
        if (!showAdvanced.value) {
          form.breadcrumbMode = defaultBreadcrumbMode()
        }
      }
      nextTick(() =>
        formRef.value
          ?.validateField(['parentMenuId', 'parentPageKey', 'permissionKey'])
          .catch(() => undefined)
      )
    }
  )

  watch(
    () => form.accessMode,
    (mode) => {
      if (mode !== 'permission') {
        form.permissionKey = ''
      }
    }
  )

  watch(
    () => form.spaceKeys,
    async () => {
      if (!visible.value || isInitializing.value) return
      form.parentMenuId = ''
      form.parentPageKey = ''
      form.displayGroupKey = ''
      await loadOptions()
    },
    { deep: true }
  )

  watch(
    () => form.isIframe,
    (value) => {
      if (value) {
        form.component = '/outside/Iframe'
        form.keepAlive = false
      } else if (form.component === '/outside/Iframe') {
        form.component = ''
        form.link = ''
      }
      nextTick(() => formRef.value?.validateField(['component', 'link']).catch(() => undefined))
    }
  )

  function handleClose() {
    visible.value = false
    submitting.value = false
    formRef.value?.resetFields()
  }

  async function handleSubmit() {
    if (!formRef.value || submitting.value) return
    try {
      const valid = await formRef.value.validate().catch(() => false)
      if (!valid) return
      submitting.value = true
      const visibilityScope =
        form.pageType === 'standalone' || form.pageType === 'standalone'
          ? form.visibilityScope
          : 'inherit'
      const payload: Api.SystemManage.PageSaveParams = {
        app_key: `${props.appKey || ''}`.trim(),
        page_key: form.pageKey.trim(),
        name: form.name.trim(),
        route_name: form.routeName.trim() || form.pageKey.trim(),
        route_path: form.routePath.trim(),
        component: form.isIframe ? '/outside/Iframe' : form.component.trim(),
        page_type: form.pageType,
        visibility_scope: visibilityScope,
        source:
          props.dialogType === 'edit'
            ? `${props.pageData?.source || 'manual'}`
            : `${props.defaultData?.source || 'manual'}`,
        module_key: form.moduleKey.trim(),
        space_keys: visibilityScope === 'spaces' ? resolveSpaceBindingKeys() : [],
        sort_order: form.sortOrder,
        parent_menu_id:
          form.pageType !== 'inner' || mountMode.value !== 'menu'
            ? ''
            : normalizeMenuId(form.parentMenuId),
        parent_page_key:
          form.pageType !== 'inner' || mountMode.value !== 'page' ? '' : form.parentPageKey || '',
        display_group_key: mountMode.value === 'page' ? '' : form.displayGroupKey || '',
        active_menu_path: showAdvanced.value ? form.activeMenuPath.trim() : '',
        breadcrumb_mode: showAdvanced.value
          ? form.breadcrumbMode
          : form.pageType === 'standalone'
            ? 'inherit_menu'
            : defaultBreadcrumbMode(),
        access_mode:
          (form.pageType === 'standalone' || form.pageType === 'standalone') &&
          form.accessMode === 'inherit'
            ? 'jwt'
            : form.accessMode,
        permission_key: form.accessMode === 'permission' ? form.permissionKey.trim() : '',
        keep_alive: form.isIframe ? false : form.keepAlive,
        is_full_page: form.isFullPage,
        remote_binding: props.pageData?.remoteBinding
          ? {
              manifest_url: props.pageData.remoteBinding.manifestUrl || '',
              remote_app_key: props.pageData.remoteBinding.remoteAppKey || '',
              remote_page_key: props.pageData.remoteBinding.remotePageKey || '',
              remote_entry_url: props.pageData.remoteBinding.remoteEntryUrl || '',
              remote_route_path: props.pageData.remoteBinding.remoteRoutePath || '',
              remote_module: props.pageData.remoteBinding.remoteModule || '',
              remote_module_name: props.pageData.remoteBinding.remoteModuleName || '',
              remote_url: props.pageData.remoteBinding.remoteUrl || '',
              runtime_version: props.pageData.remoteBinding.runtimeVersion || '',
              health_check_url: props.pageData.remoteBinding.healthCheckUrl || ''
            }
          : undefined,
        status: form.status,
        meta: {
          isIframe: form.isIframe,
          link: form.isIframe ? form.link.trim() : ''
        }
      }
      if (props.dialogType === 'edit') {
        await fetchUpdatePage(form.id, payload)
      } else {
        await fetchCreatePage(payload)
      }
      ElMessage.success(
        props.dialogType === 'edit'
          ? '修改成功'
          : props.dialogType === 'copy'
            ? '复制成功'
            : '新增成功'
      )
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error?.message || '保存失败')
    } finally {
      submitting.value = false
    }
  }

  return {
    formRef,
    submitting,
    mountMode,
    mountSpaceKey,
    showAdvanced,
    showExamples,
    visible,
    dialogTitle,
    form,
    rules,
    menuTreeOptions,
    menuSpaceOptions,
    showMountSection,
    showVisibilityScopeField,
    showSpaceBindingField,
    menuCascaderProps,
    parentPageOptions,
    displayGroupOptions,
    configHintTitle,
    isUnregisteredCandidate,
    isComponentLocked,
    configHintDescription,
    mountOwnershipSummary,
    accessModeOptions,
    routePathPlaceholder,
    mountMenuSummary,
    menuSiblingSummary,
    resolvedRoutePreview,
    routePreviewHint,
    pageExamples,
    getComponentPlaceholder,
    handleClose,
    handleSubmit
  }
}
