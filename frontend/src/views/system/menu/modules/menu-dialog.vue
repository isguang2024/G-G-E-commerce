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
      <div class="menu-dialog-intro__title">{{ isEdit ? '编辑导航入口' : '创建导航入口' }}</div>
      <div class="menu-dialog-intro__text">
        菜单只负责导航入口、显示方式和访问模式。页面挂载与页面内逻辑不要继续回到菜单层维护。
      </div>
      <div class="menu-dialog-intro__tip">
        需要显式建立菜单和页面的主关系时，可直接选择“挂接主页面”，保存后会同步更新页面归属。
      </div>
    </div>

    <div v-if="linkedPageSummary" class="menu-dialog-link-summary">
      <div class="menu-dialog-link-summary__title">挂接关系预览</div>
      <div class="menu-dialog-link-summary__text">{{ linkedPageSummary }}</div>
    </div>

    <div v-if="currentMenuPageSummary" class="menu-dialog-link-summary is-neutral">
      <div class="menu-dialog-link-summary__title">当前菜单页面关系</div>
      <div class="menu-dialog-link-summary__text">{{ currentMenuPageSummary }}</div>
    </div>

    <ArtForm
      ref="formRef"
      v-model="form"
      :items="formItems"
      :rules="rules"
      :span="width > 640 ? 12 : 24"
      :gutter="20"
      label-width="100px"
      :show-reset="false"
      :show-submit="false"
    >
      <template #menuTemplate>
        <div class="menu-type-container">
          <div class="template-buttons">
            <ElButton size="small" @click="applyTemplate('outer')" class="mr-2">
              外链模板
            </ElButton>
            <ElButton size="small" @click="applyTemplate('top')" class="mr-2">
              顶层模板菜单
            </ElButton>
            <ElButton size="small" @click="applyTemplate('sub')" class="mr-2">
              子菜单模板
            </ElButton>
          </div>
        </div>
      </template>

      <template #advancedConfig>
        <div class="advanced-config-container w-full">
          <div class="advanced-config-intro">统一配置运行时展示行为、标签行为和访问开关。</div>
          <ElCollapse v-model="activeCollapse" class="w-full">
            <ElCollapseItem title="高级配置" name="1" class="w-full">
              <div class="grid grid-cols-2 gap-4">
                <div class="flex items-center">
                  <span class="w-24">是否启用</span>
                  <ElSwitch v-model="form.isEnable" />
                </div>
                <div class="flex items-center">
                  <span class="w-24">页面缓存</span>
                  <ElSwitch v-model="form.keepAlive" />
                </div>
                <div class="flex items-center">
                  <span class="w-24">隐藏菜单</span>
                  <ElSwitch v-model="form.isHide" />
                </div>
                <div class="flex items-center">
                  <span class="w-24">是否内嵌</span>
                  <ElSwitch v-model="form.isIframe" />
                </div>
                <div class="flex items-center">
                  <span class="w-24">显示徽章</span>
                  <ElSwitch v-model="form.showBadge" />
                </div>
                <div class="flex items-center">
                  <span class="w-24">固定标签</span>
                  <ElSwitch v-model="form.fixedTab" />
                </div>
                <div class="flex items-center">
                  <span class="w-24">标签隐藏</span>
                  <ElSwitch v-model="form.isHideTab" />
                </div>
                <div class="flex items-center">
                  <span class="w-24">全屏页面</span>
                  <ElSwitch v-model="form.isFullPage" />
                </div>
              </div>
            </ElCollapseItem>
          </ElCollapse>
        </div>
      </template>
    </ArtForm>

    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="handleCancel">取 消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确 定</ElButton>
      </span>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { h, ref, reactive, computed, watch, nextTick } from 'vue'
  import type { FormRules } from 'element-plus'
  import { ElIcon, ElTooltip, ElMessage, ElCollapse, ElCollapseItem, ElSwitch } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'
  import { formatMenuTitle } from '@/utils/router'
  import type { AppRouteRecord } from '@/types/router'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import ArtForm from '@/components/core/forms/art-form/index.vue'
  import { useWindowSize } from '@vueuse/core'

  const { width } = useWindowSize()

  const createLabelTooltip = (label: string, tooltip: string) => {
    return () =>
      h('span', { class: 'flex items-center' }, [
        h('span', label),
        h(
          ElTooltip,
          {
            content: tooltip,
            placement: 'top'
          },
          () => h(ElIcon, { class: 'ml-0.5 cursor-help' }, () => h(QuestionFilled))
        )
      ])
  }

  interface MenuFormData {
    id: number
    name: string
    path: string
    label: string
    component: string
    icon: string
    parentId: string
    isEnable: boolean
    sort: number
    isMenu: boolean
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
    roles: string[]
    isFullPage: boolean
    manageGroupId: string
    linkedPageKey: string
  }

  interface Props {
    visible: boolean
    editData?: AppRouteRecord | any
    menuTree?: AppRouteRecord[]
    manageGroups?: Api.SystemManage.MenuManageGroupItem[]
    pageOptions?: Api.SystemManage.PageItem[]
    currentMenuPages?: Api.SystemManage.PageItem[]
    linkedPageKey?: string
    editingMenuId?: string
    initialParentId?: string
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', data: MenuFormData): void
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    menuTree: () => [],
    manageGroups: () => [],
    pageOptions: () => [],
    currentMenuPages: () => [],
    linkedPageKey: '',
    editingMenuId: '',
    initialParentId: ''
  })

  function collectIds(node: AppRouteRecord & { id?: string; children?: any[] }): string[] {
    const ids: string[] = []
    if (node.id) ids.push(node.id)
    if (node.children?.length) {
      node.children.forEach((ch: any) => ids.push(...collectIds(ch)))
    }
    return ids
  }

  const parentMenuOptions = computed(() => {
    const tree = props.menuTree || []
    const excludeIds = new Set<string>()
    if (props.editingMenuId && tree.length) {
      const findAndCollect = (nodes: any[]): boolean => {
        for (const n of nodes) {
          if (n.id === props.editingMenuId) {
            collectIds(n).forEach((id) => excludeIds.add(id))
            return true
          }
          if (n.children?.length && findAndCollect(n.children)) return true
        }
        return false
      }
      findAndCollect(tree)
    }
    const options: { label: string; value: string }[] = [{ label: '顶级菜单', value: '' }]
    const flatten = (nodes: any[], prefix = '') => {
      nodes.forEach((n) => {
        const id = n.id
        if (id && !excludeIds.has(id)) {
          const title = formatMenuTitle(n.meta?.title || n.name || '')
          options.push({ label: prefix + title, value: id })
        }
        if (n.children?.length) flatten(n.children, prefix + '　')
      })
    }
    flatten(tree)
    return options
  })

  const manageGroupOptions = computed(() => [
    { label: '不分组', value: '' },
    ...(props.manageGroups || []).map((item) => ({
      label: item.name,
      value: item.id
    }))
  ])

  const pageOptions = computed(() =>
    (props.pageOptions || [])
      .filter((item) => item.pageType !== 'group' && item.pageType !== 'display_group')
      .map((item) => ({
        label: item.parentMenuName
          ? `${item.name} · ${item.pageKey} · 当前挂在 ${item.parentMenuName}`
          : `${item.name} · ${item.pageKey}`,
        value: item.pageKey
      }))
  )

  const selectedLinkedPage = computed(() =>
    (props.pageOptions || []).find((item) => item.pageKey === form.linkedPageKey)
  )

  const linkedPageSummary = computed(() => {
    const page = selectedLinkedPage.value
    if (!page) return ''
    const currentMenu = `${page.parentMenuName || ''}`.trim()
    const relationText = currentMenu
      ? `当前该页面已挂在“${currentMenu}”，保存后会切换到本菜单。`
      : '当前该页面还没有挂接主菜单，保存后会归属到本菜单。'
    const permissionText =
      page.accessMode === 'permission'
        ? '页面自身已配置权限，挂到菜单后将按“菜单准入 + 页面权限”交集放行。'
        : '页面当前走继承模式，挂到菜单后默认跟随菜单权限。'
    return `${relationText}${permissionText}`
  })

  const currentMenuPageSummary = computed(() => {
    const pages = props.currentMenuPages || []
    if (!isEdit.value || pages.length === 0) {
      return isEdit.value ? '当前菜单下还没有挂接页面。' : ''
    }
    const primaryPageKey = `${form.linkedPageKey || props.linkedPageKey || ''}`.trim()
    const names = pages.map((item) =>
      item.pageKey === primaryPageKey ? `${item.name}（主页面）` : item.name
    )
    const primaryHint =
      pages.length > 1
        ? '一个菜单可以挂多个页面，但建议只保留一个主页面作为导航代表页。'
        : '当前只有一个挂接页面，适合作为主页面。'
    return `当前菜单下共有 ${pages.length} 个页面：${names.join('、')}。${primaryHint}`
  })

  const emit = defineEmits<Emits>()

  const formRef = ref()
  const isEdit = ref(false)
  const activeCollapse = ref(['1'])

  const form = reactive<MenuFormData>({
    id: 0,
    name: '',
    path: '',
    label: '',
    component: '',
    icon: '',
    parentId: '',
    isEnable: true,
    sort: 1,
    isMenu: true,
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
    roles: [],
    isFullPage: false,
    manageGroupId: '',
    linkedPageKey: ''
  })

  const rules = reactive<FormRules>({
    name: [
      { required: true, message: '请输入菜单名称', trigger: 'blur' },
      { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
    ],
    path: [{
      validator: (rule, value, callback) => {
        if (!value && !form.link) {
          callback(new Error('请输入路由地址'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }],
    label: [{ required: true, message: '输入权限标识', trigger: 'blur' }]
  })

  const formItems = computed<FormItem[]>(() => {
    return [
      {
        label: '',
        key: 'menuTemplate',
        type: 'custom',
        slotName: 'menuTemplate',
        span: 24
      },
      {
        label: '上级菜单',
        key: 'parentId',
        type: 'select',
        props: {
          placeholder: '不选则为顶级菜单',
          options: parentMenuOptions.value,
          clearable: true,
          style: { width: '100%' }
        }
      },
      { label: '菜单名称', key: 'name', type: 'input', props: { placeholder: '菜单名称' } },
      {
        label: createLabelTooltip(
          '路由地址',
          '一级菜单：以 / 开头的绝对路径（如 /dashboard）\n二级及以下：相对路径（如 console、user）\n外部链接有值时可留空'
        ),
        key: 'path',
        type: 'input',
        props: { placeholder: '如：/dashboard 或 console' }
      },
      { label: '权限标识', key: 'label', type: 'input', props: { placeholder: '如：User' } },
      {
        label: createLabelTooltip(
          '组件路径',
          '一级父级菜单：填写 /index/index\n具体页面：填写组件路径（如 /system/user）\n目录菜单：留空'
        ),
        key: 'component',
        type: 'input',
        props: { placeholder: '如：/system/user 或留空' }
      },
      { label: '图标', key: 'icon', type: 'input', props: { placeholder: '如：ri:user-line' } },
      {
        label: '访问模式',
        key: 'accessMode',
        type: 'select',
        props: {
          options: [
            { label: '权限控制', value: 'permission' },
            { label: '登录可见', value: 'jwt' },
            { label: '公开可见', value: 'public' }
          ],
          style: { width: '100%' }
        }
      },
      {
        label: createLabelTooltip(
          '角色权限',
          '仅用于前端权限模式：配置角色标识（如 R_SUPER、R_ADMIN）\n后端权限模式：无需配置'
        ),
        key: 'roles',
        type: 'inputtag',
        props: { placeholder: '输入角色标识后按回车，如：R_SUPER' }
      },
      {
        label: '菜单排序',
        key: 'sort',
        type: 'number',
        props: { controlsPosition: 'right', style: { width: '100%' } }
      },
      {
        label: createLabelTooltip(
          '管理分组',
          '仅用于菜单管理页分组显示，不影响运行时菜单、权限和同层排序。'
        ),
        key: 'manageGroupId',
        type: 'select',
        props: {
          placeholder: '可选，选择菜单管理分组',
          options: manageGroupOptions.value,
          clearable: true,
          style: { width: '100%' }
        }
      },
      {
        label: createLabelTooltip(
          '挂接主页面',
          '可选，为当前菜单指定一个主页面。保存后会把该页面挂到当前菜单下，便于在菜单管理和页面管理里双向查看。'
        ),
        key: 'linkedPageKey',
        type: 'select',
        props: {
          placeholder: '可选，选择挂接主页面',
          options: pageOptions.value,
          clearable: true,
          filterable: true,
          style: { width: '100%' }
        }
      },
      {
        label: '外部链接',
        key: 'link',
        type: 'input',
        props: { placeholder: '如：https://www.example.com' }
      },
      {
        label: '文本徽章',
        key: 'showTextBadge',
        type: 'input',
        props: { placeholder: '如：New、Hot' }
      },
      {
        label: createLabelTooltip(
          '激活路径',
          '用于详情页等隐藏菜单，指定高亮显示的父级菜单路径\n例如：用户详情页高亮显示"用户管理"菜单'
        ),
        key: 'activePath',
        type: 'input',
        props: { placeholder: '如：/system/user' }
      },
      {
        label: createLabelTooltip(
          '自定义上级',
          '用于顶级菜单，指定面包屑中显示的上级菜单路径\n例如：设置为/system/user，面包屑会显示"首页 > 用户管理 > 当前菜单"'
        ),
        key: 'customParent',
        type: 'input',
        props: { placeholder: '如：/system/user' }
      },
      {
        label: '',
        key: 'advancedConfig',
        type: 'custom',
        slotName: 'advancedConfig',
        span: 24
      }
    ]
  })

  const dialogTitle = computed(() => {
    return isEdit.value ? '编辑菜单' : '新建菜单'
  })

  const resetForm = (): void => {
    formRef.value?.reset()
  }

  function findParentIdInTree(nodes: any[], targetId: string, parentId: string = ''): string {
    if (!nodes || !Array.isArray(nodes) || !targetId) return ''
    for (const node of nodes) {
      if (!node) continue
      if (String(node.id) === String(targetId)) {
        return parentId || ''
      }
      if (node.children?.length) {
        const found = findParentIdInTree(node.children, targetId, String(node.id))
        if (found !== '') return found
      }
    }
    return ''
  }

  const loadFormData = (): void => {
    if (!props.editData) return

    isEdit.value = true

    const row = props.editData
    form.id = row.id || 0

    let parentId = ''
    if (row.parent_id != null && row.parent_id !== undefined && row.parent_id !== '') {
      parentId = String(row.parent_id)
    } else if (row.parentId != null && row.parentId !== undefined && row.parentId !== '') {
      parentId = String(row.parentId)
    } else {
      parentId = findParentIdInTree(props.menuTree || [], String(row.id))
    }
    form.parentId = parentId

    form.name = formatMenuTitle(row.meta?.title || '')
    form.path = row.path || ''
    form.label = row.name || ''
    form.component = row.component || ''
    form.icon = row.meta?.icon || ''
    form.sort = row.sort_order ?? 1
    form.isMenu = row.meta?.isMenu ?? true
    form.keepAlive = row.meta?.keepAlive ?? false
    form.isHide = row.meta?.isHide ?? false
    form.isHideTab = row.meta?.isHideTab ?? false
    form.isEnable = row.meta?.isEnable ?? true
    form.link = row.meta?.link || ''
    form.isIframe = row.meta?.isIframe ?? false
    form.showBadge = row.meta?.showBadge ?? false
    form.showTextBadge = row.meta?.showTextBadge || ''
    form.fixedTab = row.meta?.fixedTab ?? false
    form.activePath = row.meta?.activePath || ''
    form.customParent = row.meta?.customParent || ''
    form.accessMode = row.meta?.accessMode || 'permission'
    form.roles = row.meta?.roles || []
    form.isFullPage = row.meta?.isFullPage ?? false
    form.manageGroupId = String(row.manage_group_id || row.manageGroupId || row.manage_group?.id || '')
    form.linkedPageKey = props.linkedPageKey || ''
  }

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      emit('submit', { ...form })
      ElMessage.success(`${isEdit.value ? '编辑' : '新增'}成功`)
      handleCancel()
    } catch {
      ElMessage.error('表单校验失败，请检查输入')
    }
  }

  const handleCancel = (): void => {
    emit('update:visible', false)
  }

  const handleClosed = (): void => {
    resetForm()
    isEdit.value = false
  }

  watch(
    () => props.visible,
    (newVal) => {
      if (newVal) {
        nextTick(() => {
          if (props.editData) {
            loadFormData()
          } else {
            form.parentId = props.initialParentId || ''
            form.linkedPageKey = ''
          }
        })
      }
    }
  )

  const applyTemplate = (templateType: string) => {
    switch (templateType) {
      case 'outer':
        form.isIframe = true
        form.link = 'https://www.example.com'
        form.path = '/external-link'
        form.component = ''
        break
      case 'top':
        form.parentId = ''
        form.path = '/top-menu'
        form.component = '/index/index'
        form.isIframe = false
        form.link = ''
        break
      case 'sub':
        form.path = 'sub-menu'
        form.component = '/system/sub-menu'
        form.isIframe = false
        form.link = ''
        form.isHide = false
        break
    }
  }

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
  .advanced-config-intro {
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

  .menu-dialog-link-summary.is-neutral {
    border-color: rgb(226 232 240 / 0.95);
    background: linear-gradient(180deg, rgb(248 250 252 / 0.96), rgb(255 255 255 / 0.98));
  }

  .advanced-config-container {
    margin-top: 20px;
  }

  .menu-type-container {
    .template-buttons {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
  }
</style>
