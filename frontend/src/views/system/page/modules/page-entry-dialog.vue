<template>
  <ElDrawer
    v-model="visible"
    :title="dialogTitle"
    size="980px"
    direction="rtl"
    class="page-entry-drawer config-drawer"
    destroy-on-close
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <div class="dialog-intro">
        <div class="dialog-intro__main">
          <div class="dialog-intro__title">{{ configHintTitle }}</div>
          <div class="dialog-intro__desc">{{ configHintDescription }}</div>
          <div v-if="isUnregisteredCandidate" class="dialog-intro__meta">
            <ElTag size="small" effect="plain" type="warning">未注册来源，组件路径固定</ElTag>
          </div>
        </div>
        <ElButton text type="primary" @click="showExamples = !showExamples">
          {{ showExamples ? '收起示例' : '查看示例' }}
        </ElButton>
        <div v-if="showExamples" class="dialog-intro__examples">
          <div v-for="item in pageExamples" :key="item" class="dialog-intro__example">{{
            item
          }}</div>
        </div>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">基础信息</div>
          </div>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="页面名称" prop="name">
              <template #label>
                <PageFieldLabel
                  label="页面名称"
                  help="给人看的名称，显示在页面管理、面包屑预览和关联选择里。"
                />
              </template>
              <ElInput v-model="form.name" placeholder="请输入页面名称" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="页面标识" prop="pageKey">
              <template #label>
                <PageFieldLabel
                  label="页面标识"
                  help="页面的稳定业务标识，用于父子页面关联、同步识别和配置引用。上线后尽量不要改。"
                />
              </template>
              <ElInput v-model="form.pageKey" placeholder="例如 store.management.detail" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="页面类型" prop="pageType">
              <template #label>
                <PageFieldLabel
                  label="页面类型"
                  help="内页会继承菜单或上级页面链路；全局页独立存在，不要求挂到菜单。"
                />
              </template>
              <ElSelect v-model="form.pageType" style="width: 100%">
                <ElOption label="内页" value="inner" />
                <ElOption label="全局页" value="global" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="排序" prop="sortOrder">
              <template #label>
                <PageFieldLabel label="排序" help="同级页面或分组的排序值，数字越小越靠前。" />
              </template>
              <ElInputNumber v-model="form.sortOrder" :min="0" :step="1" style="width: 100%" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="模块标识" prop="moduleKey">
              <template #label>
                <PageFieldLabel
                  label="模块标识"
                  help="页面所属业务模块，用于筛选、归类和后续批量管理，例如 system、dashboard、order。"
                />
              </template>
              <ElInput v-model="form.moduleKey" placeholder="例如 system / order" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="空间视角" prop="spaceKey">
              <template #label>
                <PageFieldLabel
                  label="空间视角"
                  help="用于加载当前空间下的菜单候选和父页面候选。受管页面默认全局定义，只有少数独立页才会额外绑定到特定空间。"
                />
              </template>
              <ElSelect v-model="form.spaceKey" style="width: 100%">
                <ElOption
                  v-for="item in menuSpaceOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="路由名称" prop="routeName">
              <template #label>
                <PageFieldLabel
                  label="路由名称"
                  help="Vue Router 内部路由名。可留空，留空时默认使用页面标识。"
                />
              </template>
              <ElInput
                v-model="form.routeName"
                placeholder="例如 StoreManagementDetail；留空时默认使用页面标识"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">路由与渲染</div>
          </div>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="路由路径" prop="routePath">
              <template #label>
                <PageFieldLabel
                  label="路由路径"
                  help="单段路径会按上级菜单或上级页面自动拼接；多段绝对路径会按完整路径注册。"
                />
              </template>
              <ElInput v-model="form.routePath" :placeholder="routePathPlaceholder" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="组件路径" prop="component">
              <template #label>
                <PageFieldLabel
                  label="组件路径"
                  help="实际渲染的前端页面组件路径。内嵌模式下会自动改为 /outside/Iframe。"
                />
              </template>
              <ElInput
                v-model="form.component"
                :disabled="form.isIframe || isComponentLocked"
                :placeholder="getComponentPlaceholder()"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="是否内嵌" prop="isIframe">
              <template #label>
                <PageFieldLabel
                  label="是否内嵌"
                  help="开启后页面将通过 iframe 加载外部地址，组件路径自动切为 /outside/Iframe。"
                />
              </template>
              <ElSwitch v-model="form.isIframe" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="状态" prop="status">
              <template #label>
                <PageFieldLabel
                  label="状态"
                  help="正常状态才会参与运行时注册；停用后页面保留数据，但不会被动态加载。"
                />
              </template>
              <div class="inline-flex items-center gap-2">
                <ElSwitch v-model="form.status" active-value="normal" inactive-value="suspended" />
                <ElTag :type="form.status === 'normal' ? 'success' : 'danger'" effect="plain">
                  {{ form.status === 'normal' ? '正常' : '停用' }}
                </ElTag>
              </div>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.isIframe" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="外链地址" prop="link">
              <template #label>
                <PageFieldLabel
                  label="外链地址"
                  help="内嵌模式下必填，填写要加载的 http:// 或 https:// 地址。"
                />
              </template>
              <ElInput v-model="form.link" placeholder="例如 https://example.com" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElFormItem label="最终路径" class="final-path-item">
          <template #label>
            <PageFieldLabel
              label="最终路径"
              help="系统根据路由路径、挂载方式、上级菜单和上级页面推导出的真实访问路径。"
            />
          </template>
          <div class="route-preview-box">
            <code>{{ resolvedRoutePreview || '-' }}</code>
          </div>
        </ElFormItem>
        <div class="field-hint field-hint--section">
          {{ routePreviewHint }}
        </div>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">挂载与归属</div>
          </div>
        </div>

        <ElRow v-if="form.pageType !== 'global'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="挂载方式" prop="mountMode">
              <template #label>
                <PageFieldLabel
                  label="挂载方式"
                  help="决定当前页面是独立存在，还是归属某个菜单，或归属到另一个页面/分组下面。"
                />
              </template>
              <ElRadioGroup v-model="mountMode" class="mount-mode-group">
                <ElRadioButton label="none">不挂载</ElRadioButton>
                <ElRadioButton label="menu">挂到菜单</ElRadioButton>
                <ElRadioButton label="page">挂到页面/分组</ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>
            <div v-if="mountOwnershipSummary" class="mount-summary-box is-neutral">
              <div class="mount-summary-box__title">当前归属说明</div>
              <div class="mount-summary-box__text">{{ mountOwnershipSummary }}</div>
            </div>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.pageType !== 'global' && mountMode === 'menu'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="上级菜单" prop="parentMenuId">
              <template #label>
                <PageFieldLabel
                  label="上级菜单"
                  help="页面直接归属的菜单。单段路由会自动拼到该菜单路径后，并继承菜单高亮与菜单准入。若页面再单独配置权限，则最终按菜单权限与页面权限交集放行。"
                />
              </template>
              <ElCascader
                v-model="form.parentMenuId"
                :options="menuTreeOptions"
                :props="menuCascaderProps"
                filterable
                clearable
                show-all-levels
                style="width: 100%"
                placeholder="请选择上级菜单"
              />
            </ElFormItem>
            <div v-if="mountMenuSummary" class="mount-summary-box">
              <div class="mount-summary-box__title">挂接关系预览</div>
              <div class="mount-summary-box__text">{{ mountMenuSummary }}</div>
            </div>
            <div v-if="menuSiblingSummary" class="mount-summary-box is-neutral">
              <div class="mount-summary-box__title">同菜单页面摘要</div>
              <div class="mount-summary-box__text">{{ menuSiblingSummary }}</div>
            </div>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.pageType !== 'global' && mountMode === 'page'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="上级页面" prop="parentPageKey">
              <template #label>
                <PageFieldLabel
                  label="上级页面"
                  help="页面直接归属的父页面或逻辑分组。选择后会优先继承其访问路径、菜单链和默认面包屑。"
                />
              </template>
              <ElSelect
                v-model="form.parentPageKey"
                clearable
                filterable
                style="width: 100%"
                placeholder="请选择上级页面或逻辑分组"
              >
                <ElOption
                  v-for="item in parentPageOptions"
                  :key="item.pageKey"
                  :label="`${item.name} (${item.pageKey})`"
                  :value="item.pageKey"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.pageType === 'global' || mountMode !== 'page'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="普通分组" prop="displayGroupKey">
              <template #label>
                <PageFieldLabel
                  label="普通分组"
                  help="仅用于页面管理列表归类，不影响页面的菜单挂载、路径、权限和面包屑继承。"
                />
              </template>
              <ElSelect
                v-model="form.displayGroupKey"
                clearable
                filterable
                style="width: 100%"
                placeholder="可选，选择普通分组"
              >
                <ElOption
                  v-for="item in displayGroupOptions"
                  :key="item.pageKey"
                  :label="item.name"
                  :value="item.pageKey"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">访问与行为</div>
          </div>
          <ElButton text type="primary" @click="showAdvanced = !showAdvanced">
            {{ showAdvanced ? '收起高级配置' : '展开高级配置' }}
          </ElButton>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="访问模式" prop="accessMode">
              <template #label>
                <PageFieldLabel
                  label="访问模式"
                  help="继承表示跟随上级菜单或页面；登录表示只验登录；权限表示还需校验权限键。挂到菜单时，继承即默认跟菜单权限走；若改成权限模式，则在菜单准入基础上再校验页面权限。"
                />
              </template>
              <ElSelect v-model="form.accessMode" style="width: 100%">
                <ElOption label="继承" value="inherit" />
                <ElOption label="公开" value="public" />
                <ElOption label="登录" value="jwt" />
                <ElOption label="权限" value="permission" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="权限键" prop="permissionKey">
              <template #label>
                <PageFieldLabel
                  label="权限键"
                  help="仅在访问模式为权限时生效。挂到菜单时，这里不是覆盖菜单权限，而是在菜单准入基础上追加页面权限校验。"
                />
              </template>
              <ElInput
                v-model="form.permissionKey"
                :disabled="form.accessMode !== 'permission'"
                placeholder="accessMode=permission 时必填"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="缓存页面" prop="keepAlive">
              <template #label>
                <PageFieldLabel
                  label="缓存页面"
                  help="开启后页面会进入 keep-alive 缓存，适合表单或列表类页面；内嵌页通常不缓存。"
                />
              </template>
              <ElSwitch v-model="form.keepAlive" :disabled="form.isIframe" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="全屏页面" prop="isFullPage">
              <template #label>
                <PageFieldLabel
                  label="全屏页面"
                  help="开启后页面按全屏模式展示，通常用于沉浸式页面或不依赖常规布局的场景。"
                />
              </template>
              <ElSwitch v-model="form.isFullPage" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="showAdvanced" :gutter="14" class="advanced-grid">
          <ElCol :span="12">
            <ElFormItem label="面包屑模式" prop="breadcrumbMode">
              <template #label>
                <PageFieldLabel
                  label="面包屑模式"
                  help="继承菜单表示按菜单链展示；继承页面表示把父页面链也带上；自定义用于高级覆盖。"
                />
              </template>
              <ElSelect v-model="form.breadcrumbMode" style="width: 100%">
                <ElOption label="继承菜单" value="inherit_menu" />
                <ElOption label="继承页面" value="inherit_page" />
                <ElOption label="自定义" value="custom" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="高亮菜单路径" prop="activeMenuPath">
              <template #label>
                <PageFieldLabel
                  label="高亮菜单路径"
                  help="仅在自动推导不满足时手工覆盖菜单高亮路径。大多数页面可留空。"
                />
              </template>
              <ElInput v-model="form.activeMenuPath" placeholder="可选，例如 /system/page" />
            </ElFormItem>
          </ElCol>
        </ElRow>
      </div>
    </ElForm>

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="handleClose">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
      </div>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, nextTick, reactive, ref, watch } from 'vue'
  import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
  import {
    fetchCreatePage,
    fetchGetPageMenuOptions,
    fetchGetPageOptions,
    fetchUpdatePage
  } from '@/api/system-manage'
  import {
    isSingleSegmentManagedPagePath,
    joinManagedPagePath,
    normalizeManagedPagePath,
    resolveManagedPageRoutePath
  } from '@/utils/navigation/managed-page'
  import { formatMenuTitle } from '@/utils/router'
  import PageFieldLabel from './page-field-label.vue'

  type PageItem = Api.SystemManage.PageItem
  type PageMenuOptionItem = Api.SystemManage.PageMenuOptionItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit' | 'copy'
    pageData?: Partial<PageItem>
    menuSpaces?: Api.SystemManage.MenuSpaceItem[]
    currentSpaceKey?: string
    initialParentPageKey?: string
    initialParentMenuId?: string
    initialPageType?: PageItem['pageType']
    defaultData?: Partial<PageItem>
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    pageData: undefined,
    menuSpaces: () => [],
    currentSpaceKey: 'default',
    initialParentPageKey: '',
    initialParentMenuId: '',
    initialPageType: 'inner',
    defaultData: undefined
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const menuOptions = ref<PageMenuOptionItem[]>([])
  const allPages = ref<PageItem[]>([])
  const mountMode = ref<'none' | 'menu' | 'page'>('none')
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
    return form.pageType === 'global' ? `${actionText}全局页` : `${actionText}页面`
  })

  const form = reactive({
    id: '',
    pageKey: '',
    name: '',
    routeName: '',
    routePath: '',
    component: '',
    pageType: 'inner',
    moduleKey: '',
    spaceKey: 'default',
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
    form.pageType === 'global' ? '全局页配置说明' : '页面配置说明'
  )
  const isUnregisteredCandidate = computed(
    () => props.dialogType === 'add' && Boolean(props.defaultData?.meta?.fromUnregistered)
  )
  const isComponentLocked = computed(() => isUnregisteredCandidate.value && !form.isIframe)
  const shouldAutoAdjustRoutePath = computed(
    () => props.dialogType === 'edit' || isUnregisteredCandidate.value
  )

  const configHintDescription = computed(() => {
    if (form.pageType === 'global') {
      return '全局页属于独立页面，不依赖菜单归属，适合登录后直达或公开访问的页面。'
    }
    return '普通页面可以挂到菜单、挂到上级页面，也可以作为独立内页存在；系统会自动推导最终访问路径、菜单高亮与权限继承。'
  })

  const mountOwnershipSummary = computed(() => {
    if (form.pageType === 'global') {
      return '当前页面属于独立页面，只在页面管理中维护，不占用左侧菜单入口。'
    }
    if (mountMode.value === 'menu') {
      return '当前页面会挂到菜单下，菜单负责入口可见、默认高亮和默认准入。'
    }
    if (mountMode.value === 'page') {
      return '当前页面会挂到上级页面或逻辑分组下，优先继承其路径链、菜单链和默认面包屑。'
    }
    return '当前页面作为独立内页存在，不挂菜单，也不挂到其他页面；适合个人中心、结果页、隐式工作流页这类页面。'
  })

  const routePathPlaceholder = computed(() => {
    if (form.pageType === 'global') {
      return '例如 /store-management'
    }
    if (mountMode.value === 'none') {
      return '例如 /dashboard/example-page 或 /detail/:id'
    }
    return '例如 detail 或 detail/:id；如需绝对路径可填 /system/detail'
  })

  const mountMenuSummary = computed(() => {
    if (form.pageType === 'global' || mountMode.value !== 'menu') return ''
    const menuName = menuNameMap.value.get(normalizeMenuId(form.parentMenuId)) || '所选菜单'
    const permissionText =
      form.accessMode === 'permission'
        ? '当前页面单独配置了权限，最终会按“菜单准入 + 页面权限”交集放行。'
        : '当前页面走继承模式，最终默认跟随菜单权限。'
    return `页面将挂到“${menuName}”下，菜单负责入口可见与默认准入。${permissionText}`
  })

  const menuSiblingSummary = computed(() => {
    if (form.pageType === 'global' || mountMode.value !== 'menu') return ''
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
          form.pageType === 'global' || mountMode.value !== 'menu'
            ? ''
            : normalizeMenuId(form.parentMenuId),
        parentPageKey:
          form.pageType === 'global' || mountMode.value !== 'page' ? '' : form.parentPageKey
      },
      {
        getPageByKey: (pageKey) => pageMap.value.get(pageKey),
        getMenuPathById: (menuId) => menuPathMap.value.get(menuId)
      }
    )
  )

  const routePreviewHint = computed(() => {
    if (form.pageType === 'global') {
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
    if (form.pageType === 'global') {
      return [
        '例 1：页面类型=全局页，路由路径=/workspace/overview。适合登录后可直达页面。',
        '例 2：全局页不需要选择上级菜单或上级页面，最终路径就是你填写的完整路径。',
        '例 3：如果是外部系统页，可开启“是否内嵌”，外链地址填 https://example.com。'
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

  function toTreeSelectNode(item: PageMenuOptionItem): any {
    const title = `${item.title || ''}`.trim()
    const formattedTitle = formatMenuTitle(title)
    const menuName = `${item.name || ''}`.trim()
    const labelSource = formattedTitle || menuName || `${item.path || item.id}`.trim()
    return {
      label: labelSource,
      value: item.id,
      children: Array.isArray(item.children) ? item.children.map(toTreeSelectNode) : []
    }
  }

  function normalizeMenuId(value: unknown): string {
    if (Array.isArray(value)) {
      for (let i = value.length - 1; i >= 0; i -= 1) {
        const item = `${value[i] ?? ''}`.trim()
        if (item) return item
      }
      return ''
    }
    return `${value ?? ''}`.trim()
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
      form.pageType === 'global' ||
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

  function validateRouteName(_: unknown, value: string, callback: (error?: Error) => void) {
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
    if (form.pageType === 'global') {
      callback()
      return
    }
    if (mountMode.value === 'none') {
      callback()
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
    if (props.dialogType === 'edit' && props.pageData) {
      Object.assign(form, {
        id: props.pageData.id || '',
        pageKey: props.pageData.pageKey || '',
        name: props.pageData.name || '',
        routeName: props.pageData.routeName || '',
        routePath: props.pageData.routePath || '',
        component: props.pageData.component || '',
        pageType: props.pageData.pageType === 'global' ? 'global' : 'inner',
        moduleKey: props.pageData.moduleKey || '',
        spaceKey: props.pageData.spaceKey || props.currentSpaceKey || 'default',
        sortOrder: props.pageData.sortOrder ?? 0,
        parentMenuId: props.pageData.parentMenuId || '',
        parentPageKey: props.pageData.parentPageKey || '',
        displayGroupKey: props.pageData.displayGroupKey || '',
        activeMenuPath: props.pageData.activeMenuPath || '',
        breadcrumbMode: props.pageData.breadcrumbMode || 'inherit_menu',
        accessMode: props.pageData.accessMode || 'inherit',
        permissionKey: props.pageData.permissionKey || '',
        keepAlive: props.pageData.keepAlive ?? false,
        isFullPage: props.pageData.isFullPage ?? false,
        isIframe: Boolean(props.pageData.meta?.isIframe ?? props.pageData.isIframe),
        link: `${props.pageData.meta?.link || props.pageData.link || ''}`.trim(),
        status: props.pageData.status || 'normal'
      })
      mountMode.value = form.parentPageKey ? 'page' : form.parentMenuId ? 'menu' : 'none'
      showAdvanced.value =
        Boolean(form.activeMenuPath) || form.breadcrumbMode !== defaultBreadcrumbMode()
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
        props.defaultData?.pageType === 'global' || props.initialPageType === 'global'
          ? 'global'
          : 'inner',
      moduleKey: props.defaultData?.moduleKey || '',
      spaceKey:
        props.defaultData?.spaceKey ||
        props.currentSpaceKey ||
        props.menuSpaces?.find((item) => item.isDefault)?.spaceKey ||
        'default',
      sortOrder: props.defaultData?.sortOrder ?? 0,
      parentMenuId: props.defaultData?.parentMenuId || props.initialParentMenuId || '',
      parentPageKey: props.defaultData?.parentPageKey || props.initialParentPageKey || '',
      displayGroupKey: props.defaultData?.displayGroupKey || '',
      activeMenuPath: props.defaultData?.activeMenuPath || '',
      breadcrumbMode: 'inherit_menu',
      accessMode: props.defaultData?.accessMode || 'inherit',
      permissionKey: props.defaultData?.permissionKey || '',
      keepAlive: props.defaultData?.keepAlive ?? false,
      isFullPage: props.defaultData?.isFullPage ?? false,
      isIframe: Boolean(props.defaultData?.meta?.isIframe ?? props.defaultData?.isIframe),
      link: `${props.defaultData?.meta?.link || props.defaultData?.link || ''}`.trim(),
      status: props.defaultData?.status || 'normal'
    })
    mountMode.value = form.parentPageKey ? 'page' : form.parentMenuId ? 'menu' : 'none'
    form.breadcrumbMode = defaultBreadcrumbMode()
    showAdvanced.value = Boolean(form.activeMenuPath)
  }

  async function loadOptions() {
    const [menuRes, pageRes] = await Promise.all([
      fetchGetPageMenuOptions(form.spaceKey),
      fetchGetPageOptions(form.spaceKey)
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
    () => {
      if (mountMode.value === 'menu') {
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
      maybeAdjustRoutePathForMountedEdit()
      nextTick(() =>
        formRef.value?.validateField(['parentMenuId', 'parentPageKey']).catch(() => undefined)
      )
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
      if (pageType === 'global') {
        form.parentMenuId = ''
        form.parentPageKey = ''
        form.activeMenuPath = ''
        form.breadcrumbMode = 'inherit_menu'
        if (form.accessMode === 'inherit') {
          form.accessMode = 'jwt'
        }
      } else {
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
    () => form.spaceKey,
    async (next, prev) => {
      if (!visible.value || isInitializing.value || !next || next === prev) return
      form.parentMenuId = ''
      form.parentPageKey = ''
      form.displayGroupKey = ''
      await loadOptions()
    }
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
      const payload: Api.SystemManage.PageSaveParams = {
        page_key: form.pageKey.trim(),
        name: form.name.trim(),
        route_name: form.routeName.trim() || form.pageKey.trim(),
        route_path: form.routePath.trim(),
        component: form.isIframe ? '/outside/Iframe' : form.component.trim(),
        page_type: form.pageType,
          source:
            props.dialogType === 'edit'
              ? `${props.pageData?.source || 'manual'}`
              : `${props.defaultData?.source || 'manual'}`,
          module_key: form.moduleKey.trim(),
          space_key: form.spaceKey,
          sort_order: form.sortOrder,
        parent_menu_id:
          form.pageType === 'global' || mountMode.value !== 'menu'
            ? ''
            : normalizeMenuId(form.parentMenuId),
        parent_page_key:
          form.pageType === 'global' || mountMode.value !== 'page' ? '' : form.parentPageKey || '',
        display_group_key: mountMode.value === 'page' ? '' : form.displayGroupKey || '',
        active_menu_path: showAdvanced.value ? form.activeMenuPath.trim() : '',
        breadcrumb_mode: showAdvanced.value
          ? form.breadcrumbMode
          : form.pageType === 'global'
            ? 'inherit_menu'
            : defaultBreadcrumbMode(),
        access_mode: form.accessMode,
        permission_key: form.accessMode === 'permission' ? form.permissionKey.trim() : '',
        keep_alive: form.isIframe ? false : form.keepAlive,
        is_full_page: form.isFullPage,
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
</script>

<style scoped lang="scss">
  .field-hint {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.5;
    margin: -6px 0 10px;
  }

  .field-hint--section {
    margin-top: -2px;
  }

  .dialog-intro {
    background: linear-gradient(
      180deg,
      var(--el-fill-color-light) 0%,
      color-mix(in srgb, var(--el-fill-color-light) 72%, white) 100%
    );
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 14px;
    margin-bottom: 18px;
    padding: 14px 16px;
  }

  .dialog-intro__main {
    margin-bottom: 6px;
  }

  .dialog-intro__title {
    color: var(--el-text-color-primary);
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .dialog-intro__desc {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.7;
  }

  .dialog-intro__meta {
    margin-top: 8px;
  }

  .dialog-intro__examples {
    border-top: 1px dashed var(--el-border-color-lighter);
    margin-top: 8px;
    padding-top: 10px;
  }

  .dialog-intro__example {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.7;
  }

  .form-section {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 14px;
    margin-bottom: 16px;
    padding: 16px 16px 8px;
  }

  .form-section__header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  .form-section__title {
    color: var(--el-text-color-primary);
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .mount-mode-group {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .route-preview-box {
    align-items: center;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    color: var(--el-text-color-primary);
    display: flex;
    min-height: 40px;
    padding: 0 12px;
    width: 100%;
  }

  .route-preview-box code {
    color: inherit;
    font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
    font-size: 12px;
    word-break: break-all;
  }

  .mount-summary-box {
    margin: -4px 0 10px;
    padding: 12px 14px;
    border: 1px solid rgb(219 234 254 / 0.95);
    border-radius: 12px;
    background: linear-gradient(180deg, rgb(239 246 255 / 0.95), rgb(248 250 252 / 0.98));
  }

  .mount-summary-box__title {
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 600;
  }

  .mount-summary-box__text {
    margin-top: 6px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.7;
  }

  .mount-summary-box.is-neutral {
    border-color: var(--el-border-color-lighter);
    background: linear-gradient(
      180deg,
      color-mix(in srgb, var(--el-fill-color-light) 86%, white) 0%,
      white 100%
    );
  }

  .advanced-grid {
    border-top: 1px dashed var(--el-border-color-lighter);
    margin-top: 4px;
    padding-top: 12px;
  }

  .drawer-footer {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  :deep(.el-drawer__body) {
    max-height: calc(100vh - 126px);
    overflow-y: auto;
    padding: 14px 20px 12px;
  }

  :deep(.el-drawer__footer) {
    border-top: 1px solid var(--el-border-color-lighter);
    padding: 14px 20px 18px;
  }

  :deep(.final-path-item .el-form-item__content) {
    align-items: stretch;
  }

  :deep(.mount-mode-group .el-radio-button__inner) {
    min-width: 96px;
    justify-content: flex-end;
  }
</style>
