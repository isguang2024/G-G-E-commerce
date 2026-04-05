<template>
  <ElDrawer
    v-model="visible"
    :title="dialogTitle"
    size="860px"
    direction="rtl"
    class="page-group-drawer config-drawer"
    destroy-on-close
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <div class="dialog-intro">
        <div class="dialog-intro__main">
          <div class="dialog-intro__title">逻辑分组配置说明</div>
          <div class="dialog-intro__desc"
            >逻辑分组不注册运行时路由，但可以承接路径和权限继承，供下级页面或下级逻辑分组继续复用。</div
          >
        </div>
        <ElButton text type="primary" @click="showExamples = !showExamples">
          {{ showExamples ? '收起示例' : '查看示例' }}
        </ElButton>
        <div v-if="showExamples" class="dialog-intro__examples">
          <div v-for="item in groupExamples" :key="item" class="dialog-intro__example">{{
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
            <ElFormItem label="名称" prop="name">
              <template #label>
                <PageFieldLabel
                  label="名称"
                  help="给人看的逻辑分组名称，显示在页面管理树和上级逻辑分组选择中。逻辑分组标识由系统自动生成，无需手填。"
                />
              </template>
              <ElInput v-model="form.name" placeholder="请输入名称" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="排序" prop="sortOrder">
              <template #label>
                <PageFieldLabel label="排序" help="同级分组的排序值，数字越小越靠前。" />
              </template>
              <ElInputNumber v-model="form.sortOrder" :min="0" :step="1" style="width: 100%" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="空间可见" prop="spaceKey">
              <template #label>
                <PageFieldLabel
                  label="空间可见"
                  help="只影响这个逻辑分组在当前菜单空间里是否可见，不会复制页面定义。"
                />
              </template>
              <ElSelect
                v-model="form.spaceKeys"
                multiple
                collapse-tags
                collapse-tags-tooltip
                clearable
                filterable
                style="width: 100%"
              >
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

        <ElFormItem label="挂载方式" prop="mountMode">
          <template #label>
            <PageFieldLabel
              label="挂载方式"
              help="决定分组是独立存在，直接归属某个菜单，还是继续挂在另一个分组之下。"
            />
          </template>
          <ElRadioGroup v-model="mountMode" class="mount-mode-group">
            <ElRadioButton label="none">不挂载</ElRadioButton>
            <ElRadioButton label="menu">挂到菜单</ElRadioButton>
            <ElRadioButton label="page">挂到逻辑分组</ElRadioButton>
          </ElRadioGroup>
        </ElFormItem>
        <ElRow v-if="mountMode === 'menu'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="上级菜单" prop="parentMenuId">
              <template #label>
                <PageFieldLabel
                  label="上级菜单"
                  help="逻辑分组直接归属的菜单。分组下页面可继续继承这条菜单链路。"
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
          </ElCol>
        </ElRow>

        <ElRow v-else-if="mountMode === 'page'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="上级逻辑分组" prop="parentPageKey">
              <template #label>
                <PageFieldLabel
                  label="上级逻辑分组"
                  help="逻辑分组的父分组。选择后会自动沿父分组继承菜单链，无需重复选择菜单。"
                />
              </template>
              <ElSelect
                v-model="form.parentPageKey"
                filterable
                clearable
                style="width: 100%"
                placeholder="请选择上级逻辑分组"
              >
                <ElOption
                  v-for="item in parentGroupOptions"
                  :key="item.pageKey"
                  :label="`${item.name} (${item.pageKey})`"
                  :value="item.pageKey"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="mountMode !== 'page'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="普通分组" prop="displayGroupKey">
              <template #label>
                <PageFieldLabel
                  label="普通分组"
                  help="仅用于页面管理列表归类，不影响逻辑分组的路径、权限和面包屑继承。"
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
            <div class="form-section__title">继承信息</div>
          </div>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="基础路径" prop="routePath">
              <template #label>
                <PageFieldLabel
                  label="基础路径"
                  help="逻辑分组自身不注册页面，但这里的路径会成为下级页面和下级逻辑分组的可继承基础路径。"
                />
              </template>
              <ElInput v-model="form.routePath" :placeholder="routePathPlaceholder" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="最终路径">
              <template #label>
                <PageFieldLabel
                  label="最终路径"
                  help="系统根据挂载方式、上级菜单、上级分组和基础路径推导出的最终继承路径。"
                />
              </template>
              <div class="route-preview-box">
                <code>{{ resolvedRoutePreview || '-' }}</code>
              </div>
            </ElFormItem>
          </ElCol>
        </ElRow>
        <div class="field-hint field-hint--section">{{ routePreviewHint }}</div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="访问模式" prop="accessMode">
              <template #label>
                <PageFieldLabel
                  label="访问模式"
                  help="逻辑分组可作为权限继承节点使用。下级页面若选择继承，将沿分组链继续继承这里的访问模式或上层菜单权限。"
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
                  help="当访问模式为权限时必填。下级页面或下级逻辑分组选择继承后，会继续继承该权限约束。"
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
            <ElFormItem label="状态" prop="status">
              <template #label>
                <PageFieldLabel
                  label="状态"
                  help="正常状态的逻辑分组才会参与页面树展示；停用后数据保留，但不会作为有效链路使用。"
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
  import { joinManagedPagePath, resolveManagedPageRoutePath } from '@/utils/navigation/managed-page'
  import { formatMenuTitle } from '@/utils/router'
  import PageFieldLabel from './page-field-label.vue'

  type PageItem = Api.SystemManage.PageItem
  type PageMenuOptionItem = Api.SystemManage.PageMenuOptionItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit' | 'copy'
    pageData?: Partial<PageItem>
    appKey?: string
    menuSpaces?: Api.SystemManage.MenuSpaceItem[]
    // 仅作为可见性/候选加载视角使用，不代表页面必须绑定该空间。
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
    currentSpaceKey: '',
    initialParentPageKey: '',
    initialParentMenuId: '',
    initialPageType: 'group',
    defaultData: undefined
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const menuOptions = ref<PageMenuOptionItem[]>([])
  const allPages = ref<PageItem[]>([])
  const mountMode = ref<'none' | 'menu' | 'page'>('none')
  const showExamples = ref(false)

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const dialogTitle = computed(() => {
    if (props.dialogType === 'copy') {
      return '复制逻辑分组'
    }
    return props.dialogType === 'add' ? '新增逻辑分组' : '编辑逻辑分组'
  })

  const groupExamples = [
    '例 1：逻辑分组名称=仪表盘内页链路，挂载方式=挂到菜单，上级菜单=/dashboard。适合把仪表盘下需要继承路径或权限的页面归到一组。',
    '例 2：逻辑分组名称=仪表盘分析链路，挂载方式=挂到逻辑分组，上级逻辑分组=仪表盘内页链路。适合继续往下挂趋势页、报表页。',
    '例 3：如果只想在列表里归类显示，可以给逻辑分组额外挂一个普通分组；普通分组不影响路径和权限。'
  ]

  const form = reactive({
    id: '',
    pageKey: '',
    name: '',
    routePath: '',
    accessMode: 'inherit',
    permissionKey: '',
    moduleKey: '',
    // 兼容旧接口保留的视角字段：仅用于加载候选，不是页面主语义。
    spaceKey: '',
    spaceKeys: [] as string[],
    sortOrder: 0,
    parentMenuId: '',
    parentPageKey: '',
    displayGroupKey: '',
    status: 'normal'
  })

  const rules = reactive<FormRules>({
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    parentMenuId: [{ validator: validateMountTarget, trigger: 'change' }],
    parentPageKey: [{ validator: validateMountTarget, trigger: 'change' }],
    permissionKey: [{ validator: validatePermissionKey, trigger: 'blur' }]
  })

  const menuTreeOptions = computed(() => menuOptions.value.map(toTreeSelectNode))
  const menuSpaceOptions = computed(() =>
    [
      { label: '全空间可见', value: '__all__' },
      ...(props.menuSpaces || []).map((item) => ({
        label: item.isDefault ? `${item.name}（默认）` : item.name,
        value: item.spaceKey
      }))
    ]
  )
  const menuCascaderProps = {
    checkStrictly: true,
    emitPath: false
  }
  const parentGroupOptions = computed(() =>
    allPages.value.filter(
      (item) => item.id !== form.id && `${item.pageType || ''}`.trim() === 'group'
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
  const routePathPlaceholder = computed(() => {
    if (mountMode.value === 'none') {
      return '例如 /dashboard/analysis 或 analysis'
    }
    return '例如 analysis 或 report/detail'
  })
  const resolvedRoutePreview = computed(() =>
    resolveManagedPageRoutePath<
      Pick<PageItem, 'pageKey' | 'routePath' | 'parentMenuId' | 'parentPageKey'>
    >(
      {
        pageKey: form.pageKey,
        routePath: form.routePath,
        parentMenuId: mountMode.value === 'menu' ? normalizeMenuId(form.parentMenuId) : '',
        parentPageKey: mountMode.value === 'page' ? form.parentPageKey : ''
      },
      {
        getPageByKey: (pageKey) => pageMap.value.get(pageKey),
        getMenuPathById: (menuId) => menuPathMap.value.get(menuId)
      }
    )
  )
  const routePreviewHint = computed(() => {
    if (!`${form.routePath || ''}`.trim()) {
      return '不填写基础路径时，分组只作为结构节点或权限继承节点存在。'
    }
    if (mountMode.value === 'none') {
      return '不挂载时，基础路径会成为一条独立的继承前缀；下级页面填写相对路径时会继续拼接在这里。'
    }
    if (mountMode.value === 'page') {
      return '挂到分组时，当前基础路径会继续拼到上级分组路径后，下级页面再在这条路径上继续继承。'
    }
    return '挂到菜单时，当前基础路径会拼到菜单路径后，下级页面会继续继承这条完整路径。'
  })

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

  function validateMountTarget(_: unknown, __: string, callback: (error?: Error) => void) {
    if (mountMode.value === 'none') {
      callback()
      return
    }
    if (mountMode.value === 'menu' && !normalizeMenuId(form.parentMenuId)) {
      callback(new Error('请选择上级菜单'))
      return
    }
    if (mountMode.value === 'page' && !`${form.parentPageKey || ''}`.trim()) {
      callback(new Error('请选择上级分组'))
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
      const spaceKeys = Array.isArray(props.pageData.meta?.spaceKeys)
        ? props.pageData.meta?.spaceKeys
        : []
      Object.assign(form, {
        id: props.pageData.id || '',
        pageKey: props.pageData.pageKey || '',
        name: props.pageData.name || '',
        routePath: props.pageData.routePath || '',
        accessMode: props.pageData.accessMode || 'inherit',
        permissionKey: props.pageData.permissionKey || '',
        moduleKey: props.pageData.moduleKey || '',
        spaceKey: props.pageData.spaceKey || '',
        spaceKeys:
          props.pageData.pageType === 'global'
            ? ['__all__']
            : spaceKeys.length > 0
              ? spaceKeys
              : props.pageData.spaceKey
                ? [props.pageData.spaceKey]
                : [],
        sortOrder: props.pageData.sortOrder ?? 0,
        parentMenuId: props.pageData.parentMenuId || '',
        parentPageKey: props.pageData.parentPageKey || '',
        displayGroupKey: props.pageData.displayGroupKey || '',
        status: props.pageData.status || 'normal'
      })
      mountMode.value = form.parentPageKey ? 'page' : form.parentMenuId ? 'menu' : 'none'
      return
    }

    Object.assign(form, {
      id: '',
      pageKey: props.defaultData?.pageKey || '',
      name: props.defaultData?.name || '',
        routePath: props.defaultData?.routePath || '',
        accessMode: props.defaultData?.accessMode || 'inherit',
        permissionKey: props.defaultData?.permissionKey || '',
      moduleKey: props.defaultData?.moduleKey || '',
      spaceKey: props.defaultData?.spaceKey || '',
      spaceKeys:
        props.defaultData?.pageType === 'global' || props.initialPageType === 'global'
          ? ['__all__']
          : props.defaultData?.spaceKey
            ? [props.defaultData.spaceKey]
            : [],
        sortOrder: props.defaultData?.sortOrder ?? 0,
      parentMenuId: props.defaultData?.parentMenuId || props.initialParentMenuId || '',
      parentPageKey: props.defaultData?.parentPageKey || props.initialParentPageKey || '',
      displayGroupKey: props.defaultData?.displayGroupKey || '',
      status: props.defaultData?.status || 'normal'
    })
    mountMode.value = form.parentPageKey ? 'page' : form.parentMenuId ? 'menu' : 'none'
  }

  const resolvedModuleKey = computed(() => {
    const ownModuleKey = `${form.moduleKey || ''}`.trim()
    if (mountMode.value === 'page' && `${form.parentPageKey || ''}`.trim()) {
      const parent = allPages.value.find((item) => item.pageKey === form.parentPageKey)
      const parentModuleKey = `${parent?.moduleKey || ''}`.trim()
      if (parentModuleKey) {
        return parentModuleKey
      }
    }
    return ownModuleKey
  })

  async function loadOptions() {
    // 页面中心仍按“当前空间视角”返回可挂接菜单与父分组选项，避免跨空间候选串线。
    const scopeKey = (form.spaceKeys.find((item) => item !== '__all__') || form.spaceKey || '')
    const appKey = `${props.appKey || ''}`.trim()
    const [menuRes, pageRes] = await Promise.all([
      fetchGetPageMenuOptions(scopeKey || undefined, appKey),
      fetchGetPageOptions(scopeKey || undefined, appKey)
    ])
    menuOptions.value = menuRes.records || []
    allPages.value = pageRes.records || []
  }

  async function prepareDialog() {
    await nextTick()
    initForm()
    await loadOptions()
    await nextTick()
    formRef.value?.clearValidate()
  }

  watch(
    () => form.spaceKeys,
    async () => {
      if (!props.modelValue) return
      form.parentMenuId = ''
      form.parentPageKey = ''
      form.displayGroupKey = ''
      await loadOptions()
      nextTick(() =>
        formRef.value
          ?.validateField(['parentMenuId', 'parentPageKey', 'displayGroupKey'])
          .catch(() => undefined)
      )
    }
  )

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
      nextTick(() =>
        formRef.value?.validateField(['parentMenuId', 'parentPageKey']).catch(() => undefined)
      )
    }
  )

  watch(
    () => form.accessMode,
    (mode) => {
      if (mode !== 'permission') {
        form.permissionKey = ''
      }
      nextTick(() => formRef.value?.validateField(['permissionKey']).catch(() => undefined))
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
      const nextSpaceKey = `${form.spaceKeys.includes('__all__') ? '' : form.spaceKeys[0] || form.spaceKey || ''}`.trim()
      const payload: Api.SystemManage.PageSaveParams = {
        app_key: `${props.appKey || ''}`.trim(),
        page_key: props.dialogType === 'edit' ? form.pageKey.trim() : '',
        name: form.name.trim(),
        route_name: props.dialogType === 'edit' ? form.pageKey.trim() : '',
        route_path: form.routePath.trim(),
        component: '',
        page_type: 'group',
        source:
          props.dialogType === 'edit'
            ? `${props.pageData?.source || 'manual'}`
            : `${props.defaultData?.source || 'manual'}`,
        module_key: resolvedModuleKey.value,
        // 兼容后端当前写接口；真正的空间暴露归属由后端统一编译，不在这里直接复制页面定义。
        space_key: nextSpaceKey,
        space_keys: form.spaceKeys.includes('__all__')
          ? []
          : form.spaceKeys.filter((item) => item !== '__all__'),
        sort_order: form.sortOrder,
        parent_menu_id: mountMode.value === 'menu' ? normalizeMenuId(form.parentMenuId) : '',
        parent_page_key: mountMode.value === 'page' ? form.parentPageKey || '' : '',
        display_group_key: mountMode.value === 'page' ? '' : form.displayGroupKey || '',
        active_menu_path: '',
        breadcrumb_mode: mountMode.value === 'page' ? 'inherit_page' : 'inherit_menu',
        access_mode: form.accessMode,
        permission_key: form.accessMode === 'permission' ? form.permissionKey.trim() : '',
        inherit_permission: true,
        keep_alive: false,
        is_full_page: false,
        status: form.status,
        meta: {}
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
    margin: -6px 0 12px;
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

  .dialog-intro__examples {
    border-top: 1px dashed var(--el-border-color-lighter);
    margin-top: 8px;
    padding-top: 12px;
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

  :deep(.mount-mode-group .el-radio-button__inner) {
    min-width: 96px;
  }
</style>

