<template>
  <ElDrawer
    :title="dialogTitle"
    :model-value="visible"
    @update:model-value="handleCancel"
    size="960px"
    direction="rtl"
    class="menu-dialog config-drawer"
    @closed="handleClosed"
    :before-close="handleCancel"
  >
    <div class="menu-dialog-intro">
      <div class="menu-dialog-intro__title">{{ isEdit ? '编辑菜单节点' : '创建菜单节点' }}</div>
      <div class="menu-dialog-intro__text">{{ modeIntroText }}</div>
      <div class="menu-dialog-intro__tip">{{ modeIntroTip }}</div>
    </div>

    <div v-if="currentMenuPageSummary" class="menu-dialog-link-summary">
      <div class="menu-dialog-link-summary__title">当前受管页面关系</div>
      <div class="menu-dialog-link-summary__text">{{ currentMenuPageSummary }}</div>
    </div>

    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      class="menu-dialog-form"
    >
      <div class="menu-kind-panel">
        <div class="menu-kind-panel__title">菜单类型</div>
        <ElRadioGroup v-model="form.kind" class="menu-kind-group">
          <ElRadioButton label="directory">目录</ElRadioButton>
          <ElRadioButton label="entry">入口</ElRadioButton>
          <ElRadioButton label="external">外链</ElRadioButton>
        </ElRadioGroup>
        <div class="menu-kind-panel__desc">{{ kindDescription }}</div>
      </div>

      <ElRow :gutter="16">
        <ElCol :span="showSpaceField ? 12 : 24">
          <ElFormItem label="上级菜单" prop="parentId">
            <ElSelect
              v-model="form.parentId"
              clearable
              filterable
              style="width: 100%"
              placeholder="不选则为顶级菜单"
            >
              <ElOption
                v-for="item in parentMenuOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </ElSelect>
            <div class="field-hint"
              >目录与入口都按菜单树层级组织；需要隐藏详情页时请改到受管页面中心。</div
            >
          </ElFormItem>
        </ElCol>
        <ElCol v-if="showSpaceField" :span="12">
          <ElFormItem label="菜单空间" prop="spaceKey">
            <ElSelect v-model="form.spaceKey" style="width: 100%">
              <ElOption
                v-for="item in menuSpaceOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </ElSelect>
            <div class="field-hint">菜单空间只管理菜单树和默认入口，不再复制全量页面定义。</div>
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="菜单标题" prop="name">
            <ElInput v-model="form.name" placeholder="例如 协作空间管理 / 菜单空间" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="路由标识" prop="label">
            <ElInput v-model="form.label" placeholder="例如 TeamManage / MenuSpaceManage" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem :label="pathLabel" prop="path">
            <ElInput v-model="form.path" :placeholder="pathPlaceholder" />
            <div class="field-hint">{{ pathHint }}</div>
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="图标">
            <ElInput v-model="form.icon" placeholder="例如 ri:group-line" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow v-if="isEntryKind" :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="组件路径" prop="component">
            <ElInput
              v-model="form.component"
              placeholder="例如 /collaboration-workspace/workspace"
            />
            <div class="field-hint">入口菜单直接注册为页面入口，不再额外创建同路径页面记录。</div>
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="访问模式" prop="accessMode">
            <ElSelect v-model="form.accessMode" style="width: 100%">
              <ElOption label="权限控制" value="permission" />
              <ElOption label="登录可见" value="jwt" />
              <ElOption label="公开可见" value="public" />
            </ElSelect>
            <div class="field-hint"
              >前端只消费后端编译后的可见结果，这里的配置会进入统一 AccessGraph。</div
            >
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow v-if="isExternalKind" :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="外部链接" prop="link">
            <ElInput v-model="form.link" placeholder="例如 https://docs.example.com" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="访问模式" prop="accessMode">
            <ElSelect v-model="form.accessMode" style="width: 100%">
              <ElOption label="权限控制" value="permission" />
              <ElOption label="登录可见" value="jwt" />
              <ElOption label="公开可见" value="public" />
            </ElSelect>
            <div class="field-hint">外链菜单是否展示仍由后端统一裁剪；这里只保留显式配置。</div>
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="菜单排序">
            <ElInputNumber v-model="form.sort" :min="0" :step="1" style="width: 100%" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <div class="menu-dialog-advanced">
        <div class="menu-dialog-advanced__title">运行时展示</div>
        <ElRow :gutter="16">
          <ElCol :span="8">
            <div class="switch-field">
              <span>启用</span>
              <ElSwitch v-model="form.isEnable" />
            </div>
          </ElCol>
          <ElCol :span="8">
            <div class="switch-field">
              <span>隐藏菜单</span>
              <ElSwitch v-model="form.isHide" />
            </div>
          </ElCol>
          <ElCol :span="8">
            <div class="switch-field">
              <span>显示徽章</span>
              <ElSwitch v-model="form.showBadge" />
            </div>
          </ElCol>
        </ElRow>

        <ElRow v-if="isEntryKind" :gutter="16" class="menu-dialog-advanced__row">
          <ElCol :span="8">
            <div class="switch-field">
              <span>页面缓存</span>
              <ElSwitch v-model="form.keepAlive" />
            </div>
          </ElCol>
          <ElCol :span="8">
            <div class="switch-field">
              <span>隐藏标签</span>
              <ElSwitch v-model="form.isHideTab" />
            </div>
          </ElCol>
          <ElCol :span="8">
            <div class="switch-field">
              <span>固定标签</span>
              <ElSwitch v-model="form.fixedTab" />
            </div>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.kind !== 'directory'" :gutter="16" class="menu-dialog-advanced__row">
          <ElCol :span="8">
            <div class="switch-field">
              <span>内嵌外链</span>
              <ElSwitch v-model="form.isIframe" />
            </div>
          </ElCol>
        </ElRow>

        <ElRow v-if="isEntryKind" :gutter="16" class="menu-dialog-advanced__row">
          <ElCol :span="8">
            <div class="switch-field">
              <span>全屏页面</span>
              <ElSwitch v-model="form.isFullPage" />
            </div>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem label="文本徽章">
              <ElInput v-model="form.showTextBadge" placeholder="例如 New" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="isEntryKind" :gutter="16">
          <ElCol :span="12">
            <ElFormItem label="激活路径">
              <ElInput v-model="form.activePath" placeholder="例如 /system/menu" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="自定义上级">
              <ElInput v-model="form.customParent" placeholder="例如 /system/navigation" />
            </ElFormItem>
          </ElCol>
        </ElRow>
      </div>
    </ElForm>

    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="handleCancel">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">保存</ElButton>
      </span>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, watch, nextTick } from 'vue'
  import type { FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { formatMenuTitle } from '@/utils/router'
  import type { AppRouteRecord } from '@/types/router'

  type MenuKind = 'directory' | 'entry' | 'external'

  interface MenuFormData {
    id: number
    kind: MenuKind
    name: string
    path: string
    label: string
    component: string
    icon: string
    parentId: string
    isEnable: boolean
    sort: number
    keepAlive: boolean
    isHide: boolean
    isHideTab: boolean
    link: string
    isIframe: boolean
    showBadge: boolean
    showTextBadge: string
    fixedTab: boolean
    activePath: string
    customParent: string
    accessMode: 'permission' | 'jwt' | 'public'
    isFullPage: boolean
    spaceKey: string
  }

  interface Props {
    visible: boolean
    editData?: AppRouteRecord | any
    menuTree?: AppRouteRecord[]
    menuSpaces?: Api.SystemManage.MenuSpaceItem[]
    currentSpaceKey?: string
    currentMenuPages?: Api.SystemManage.PageItem[]
    editingMenuId?: string
    initialParentId?: string
    showSpaceField?: boolean
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', data: MenuFormData): void
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    menuTree: () => [],
    menuSpaces: () => [],
    currentSpaceKey: '',
    currentMenuPages: () => [],
    editingMenuId: '',
    initialParentId: '',
    showSpaceField: false
  })

  const emit = defineEmits<Emits>()
  const formRef = ref()
  const isEdit = ref(false)

  const form = reactive<MenuFormData>({
    id: 0,
    kind: 'entry',
    name: '',
    path: '',
    label: '',
    component: '',
    icon: '',
    parentId: '',
    isEnable: true,
    sort: 0,
    keepAlive: true,
    isHide: false,
    isHideTab: false,
    link: '',
    isIframe: false,
    showBadge: false,
    showTextBadge: '',
    fixedTab: false,
    activePath: '',
    customParent: '',
    accessMode: 'permission',
    isFullPage: false,
    spaceKey: ''
  })

  function collectIds(node: AppRouteRecord & { id?: string; children?: any[] }): string[] {
    const ids: string[] = []
    if (node.id) ids.push(String(node.id))
    if (node.children?.length) {
      node.children.forEach((child: any) => ids.push(...collectIds(child)))
    }
    return ids
  }

  const parentMenuOptions = computed(() => {
    const tree = props.menuTree || []
    const excludeIds = new Set<string>()
    if (props.editingMenuId && tree.length) {
      const findAndCollect = (nodes: any[]): boolean => {
        for (const node of nodes) {
          if (String(node.id) === props.editingMenuId) {
            collectIds(node).forEach((id) => excludeIds.add(id))
            return true
          }
          if (node.children?.length && findAndCollect(node.children)) {
            return true
          }
        }
        return false
      }
      findAndCollect(tree)
    }

    const options: Array<{ label: string; value: string }> = [{ label: '顶级菜单', value: '' }]
    const walk = (nodes: any[], prefix = '') => {
      nodes.forEach((node) => {
        const id = `${node?.id || ''}`.trim()
        if (id && !excludeIds.has(id)) {
          const title = formatMenuTitle(node.meta?.title || node.name || '')
          options.push({ label: `${prefix}${title}`, value: id })
        }
        if (node.children?.length) {
          walk(node.children, `${prefix}  `)
        }
      })
    }
    walk(tree)
    return options
  })

  const menuSpaceOptions = computed(() =>
    (props.menuSpaces || []).map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.spaceKey
    }))
  )
  const resolvedFallbackSpaceKey = computed(
    () => props.currentSpaceKey || props.menuSpaces?.find((item) => item.isDefault)?.spaceKey || ''
  )
  const showSpaceField = computed(() => props.showSpaceField)

  const isEntryKind = computed(() => form.kind === 'entry')
  const isExternalKind = computed(() => form.kind === 'external')
  const dialogTitle = computed(() => (isEdit.value ? '编辑菜单' : '新建菜单'))
  const modeIntroText = computed(() =>
    showSpaceField.value
      ? '当前正在维护指定空间下的布局树，菜单定义会同步更新，父级、排序和可见性只作用于当前空间。'
      : '当前按菜单定义管理 App 级资源，空间只负责摆放布局；详情页和流程页继续交给受管页面中心维护。'
  )
  const modeIntroTip = computed(() =>
    showSpaceField.value
      ? '目录只负责分组；入口菜单直接持有 path / component；外链菜单只维护 link。'
      : '目录只负责分组；定义保存后会按当前查看空间自动补齐一条布局记录，后续空间差异在高级布局页维护。'
  )

  const kindDescription = computed(() => {
    if (form.kind === 'directory') {
      return '目录只负责导航分组与层级，不再承载具体页面组件。'
    }
    if (form.kind === 'external') {
      return '外链菜单只维护菜单展示与 link，适合文档、外部系统和跳转入口。'
    }
    return '入口菜单直接就是页面入口，刷新和深链都依赖这条菜单路由，不需要重复建页面记录。'
  })

  const pathLabel = computed(() => (form.kind === 'external' ? '菜单路径键' : '路由地址'))
  const pathPlaceholder = computed(() => {
    if (form.kind === 'directory') return '例如 navigation 或 /system'
    if (form.kind === 'external') return '例如 /docs 或 /outside/manual'
    return '例如 collaboration-workspace 或 /collaboration-workspace/workspaces'
  })
  const pathHint = computed(() => {
    if (form.kind === 'directory') {
      return '目录通常使用相对路径组织层级；一级目录也可以直接使用 /system 这类绝对路径。'
    }
    if (form.kind === 'external') {
      return '外链菜单仍保留一个稳定 path，方便路由注册、选中态和默认首页引用。'
    }
    return '入口菜单直接注册为页面入口；子级一般填相对路径，顶级入口可直接填绝对路径。'
  })

  const currentMenuPageSummary = computed(() => {
    const pages = props.currentMenuPages || []
    if (!isEdit.value || pages.length === 0) {
      return ''
    }
    const names = pages.map((item) => `${item.name}（${item.pageKey}）`)
    return `当前菜单下还关联 ${pages.length} 个受管页面：${names.join('、')}。这些关系现在只作为受管页面挂载结果查看，不再在菜单表单里直接编辑。`
  })

  const rules = reactive<FormRules>({
    name: [
      { required: true, message: '请输入菜单标题', trigger: 'blur' },
      { min: 2, max: 30, message: '长度在 2 到 30 个字符', trigger: 'blur' }
    ],
    label: [{ required: true, message: '请输入路由标识', trigger: 'blur' }],
    path: [
      {
        validator: (_rule, value, callback) => {
          const target = `${value || ''}`.trim()
          if (!target) {
            callback(new Error(form.kind === 'external' ? '请输入菜单路径键' : '请输入路由地址'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    component: [
      {
        validator: (_rule, value, callback) => {
          if (!isEntryKind.value) {
            callback()
            return
          }
          if (!`${value || ''}`.trim()) {
            callback(new Error('入口菜单必须填写组件路径'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    link: [
      {
        validator: (_rule, value, callback) => {
          if (!isExternalKind.value) {
            callback()
            return
          }
          const target = `${value || ''}`.trim()
          if (!target) {
            callback(new Error('外链菜单必须填写链接地址'))
            return
          }
          if (!/^https?:\/\//i.test(target)) {
            callback(new Error('外部链接必须以 http:// 或 https:// 开头'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  })

  function inferMenuKind(row: any): MenuKind {
    const target = `${row?.kind || ''}`.trim()
    if (target === 'directory' || target === 'entry' || target === 'external') {
      return target
    }
    if (`${row?.meta?.link || ''}`.trim()) {
      return 'external'
    }
    if (`${row?.component || ''}`.trim() && `${row?.component || ''}`.trim() !== '/index/index') {
      return 'entry'
    }
    return 'directory'
  }

  function resetForm() {
    form.id = 0
    form.kind = 'entry'
    form.name = ''
    form.path = ''
    form.label = ''
    form.component = ''
    form.icon = ''
    form.parentId = props.initialParentId || ''
    form.isEnable = true
    form.sort = 0
    form.keepAlive = true
    form.isHide = false
    form.isHideTab = false
    form.link = ''
    form.isIframe = false
    form.showBadge = false
    form.showTextBadge = ''
    form.fixedTab = false
    form.activePath = ''
    form.customParent = ''
    form.accessMode = 'permission'
    form.isFullPage = false
    form.spaceKey = resolvedFallbackSpaceKey.value
    formRef.value?.clearValidate?.()
  }

  function loadFormData() {
    if (!props.editData) {
      resetForm()
      return
    }
    const row = props.editData
    isEdit.value = true
    form.id = row.id || 0
    form.kind = inferMenuKind(row)
    form.parentId = `${row.parent_id || row.parentId || ''}`.trim()
    form.name = formatMenuTitle(row.meta?.title || '')
    form.path = row.path || ''
    form.label = row.name || ''
    form.component = row.component || ''
    form.icon = row.meta?.icon || ''
    form.sort = Number(row.sort_order ?? row.sortOrder ?? 0)
    form.isEnable = row.meta?.isEnable !== false
    form.keepAlive = row.meta?.keepAlive === true
    form.isHide = row.meta?.isHide === true
    form.isHideTab = row.meta?.isHideTab === true
    form.link = row.meta?.link || ''
    form.isIframe = row.meta?.isIframe === true
    form.showBadge = row.meta?.showBadge === true
    form.showTextBadge = row.meta?.showTextBadge || ''
    form.fixedTab = row.meta?.fixedTab === true
    form.activePath = row.meta?.activePath || ''
    form.customParent = row.meta?.customParent || ''
    form.accessMode = row.meta?.accessMode || 'permission'
    form.isFullPage = row.meta?.isFullPage === true
    form.spaceKey =
      `${row.spaceKey || row.space_key || row.meta?.spaceKey || resolvedFallbackSpaceKey.value || ''}`.trim()
  }

  function applyKindSideEffects(kind: MenuKind) {
    if (kind === 'directory') {
      form.component = ''
      form.link = ''
      form.isIframe = false
      form.activePath = ''
      form.customParent = ''
      form.keepAlive = false
      form.isHideTab = false
      form.fixedTab = false
      form.isFullPage = false
    }
    if (kind === 'entry') {
      form.link = ''
    }
    if (kind === 'external') {
      form.component = ''
      form.keepAlive = false
      form.fixedTab = false
      form.isFullPage = false
      form.activePath = ''
      form.customParent = ''
      form.isHideTab = false
    }
  }

  async function handleSubmit() {
    try {
      await formRef.value?.validate()
      emit('submit', { ...form })
      emit('update:visible', false)
    } catch {
      ElMessage.error('表单校验失败，请检查输入')
    }
  }

  function handleCancel() {
    emit('update:visible', false)
  }

  function handleClosed() {
    isEdit.value = false
    resetForm()
  }

  watch(
    () => props.visible,
    (value) => {
      if (!value) {
        return
      }
      nextTick(() => {
        if (props.editData) {
          loadFormData()
        } else {
          isEdit.value = false
          resetForm()
        }
      })
    }
  )

  watch(
    () => form.kind,
    (value) => {
      applyKindSideEffects(value)
      nextTick(() => {
        formRef.value?.validateField?.(['path', 'component', 'link']).catch(() => undefined)
      })
    }
  )
</script>

<style lang="scss" scoped>
  .menu-dialog-intro {
    padding: 14px 16px;
    margin-bottom: 16px;
    border: 1px solid rgb(226 232 240 / 0.95);
    border-radius: 16px;
    background: linear-gradient(135deg, rgb(248 250 252 / 0.98), rgb(241 245 249 / 0.95));
  }

  .menu-dialog-intro__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .menu-dialog-intro__text,
  .menu-dialog-intro__tip,
  .field-hint,
  .menu-kind-panel__desc {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .menu-dialog-intro__tip {
    color: #475569;
  }

  .menu-dialog-link-summary {
    padding: 12px 14px;
    margin-bottom: 16px;
    border: 1px solid rgb(219 234 254 / 0.95);
    border-radius: 14px;
    background: linear-gradient(180deg, rgb(239 246 255 / 0.95), rgb(248 250 252 / 0.98));
  }

  .menu-dialog-link-summary__title {
    font-size: 13px;
    font-weight: 600;
    color: #0f172a;
  }

  .menu-dialog-link-summary__text {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.7;
    color: #475569;
  }

  .menu-dialog-form {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .menu-kind-panel,
  .menu-dialog-advanced {
    padding: 14px 16px;
    margin-bottom: 16px;
    border: 1px solid rgb(226 232 240 / 0.95);
    border-radius: 16px;
    background: #fff;
  }

  .menu-kind-panel__title,
  .menu-dialog-advanced__title {
    font-size: 13px;
    font-weight: 700;
    color: #0f172a;
  }

  .menu-kind-group {
    margin-top: 12px;
  }

  .menu-dialog-advanced__row {
    margin-top: 2px;
  }

  .switch-field {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 32px;
    margin-top: 12px;
    padding: 0 2px;
    color: #334155;
    font-size: 13px;
  }

  .dialog-footer {
    display: inline-flex;
    gap: 12px;
  }

  @media (max-width: 768px) {
    .switch-field {
      margin-top: 0;
      margin-bottom: 12px;
    }
  }
</style>
