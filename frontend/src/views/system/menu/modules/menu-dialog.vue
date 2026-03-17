<template>
  <ElDialog
    :title="dialogTitle"
    :model-value="visible"
    @update:model-value="handleCancel"
    width="1200px"
    align-center
    class="menu-dialog"
    @closed="handleClosed"
    :before-close="handleCancel"
  >
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
      <template #menuType>
        <div class="menu-type-container">
          <ElRadioGroup v-model="form.menuType" :disabled="disableMenuType" class="mb-3">
            <ElRadioButton value="menu" label="menu">菜单</ElRadioButton>
            <ElRadioButton value="inner" label="inner">内页</ElRadioButton>
          </ElRadioGroup>
          
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
            <ElButton size="small" @click="applyTemplate('inner')">
              内页模板
            </ElButton>
          </div>
        </div>
      </template>
      
      <template #advancedConfig>
        <div v-if="form.menuType === 'menu' || form.menuType === 'inner'" class="advanced-config-container w-full">
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
                  <ElSwitch v-model="form.isHide" :disabled="form.menuType === 'inner'" />
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
                <div class="flex items-center">
                  <span class="w-24">团队上下文</span>
                  <ElSwitch v-model="form.requiresTenantContext" />
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
  </ElDialog>
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

  /**
   * 创建带 tooltip 的表单标签
   * @param label 标签文本
   * @param tooltip 提示文本
   * @returns 渲染函数
   */
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
    roles: string[]
    isFullPage: boolean
    requiresTenantContext: boolean
  }

  interface Props {
    visible: boolean
    editData?: AppRouteRecord | any
    /** 完整菜单树，用于生成「上级菜单」选项 */
    menuTree?: AppRouteRecord[]
    /** 当前编辑的菜单 id，其自身及子级不可选为上级（防循环） */
    editingMenuId?: string
    /** 新增时默认上级（如从某行点「新增」则为该行 id） */
    initialParentId?: string
    type?: 'menu' | 'inner'
    lockType?: boolean
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', data: MenuFormData): void
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    menuTree: () => [],
    editingMenuId: '',
    initialParentId: '',
    type: 'menu',
    lockType: false
  })

  /** 递归收集节点及其所有子孙的 id（用于编辑时排除，防止循环） */
  function collectIds(node: AppRouteRecord & { id?: string; children?: any[] }): string[] {
    const ids: string[] = []
    if (node.id) ids.push(node.id)
    if (node.children?.length) {
      node.children.forEach((ch: any) => ids.push(...collectIds(ch)))
    }
    return ids
  }

  /** 将菜单树扁平化为「上级菜单」选项（含顶级），编辑时排除当前节点及子级 */
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

  const emit = defineEmits<Emits>()

  const formRef = ref()
  const isEdit = ref(false)
  const activeCollapse = ref(['1'])

  const form = reactive<MenuFormData & { menuType: 'menu' | 'inner' }>({
    menuType: 'menu',
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
    roles: [],
    isFullPage: false,
    requiresTenantContext: false
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

  /**
   * 表单项配置
   */
  const formItems = computed<FormItem[]>(() => {
    const baseItems: FormItem[] = [{ label: '菜单类型', key: 'menuType', span: 24 }]

    if (form.menuType === 'menu' || form.menuType === 'inner') {
      return [
        ...baseItems,
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
    }
    return baseItems
  })

  const dialogTitle = computed(() => {
    const typeMap = { menu: '菜单', inner: '内页' }
    const type = typeMap[form.menuType] ?? '菜单'
    return isEdit.value ? `编辑${type}` : `新建${type}`
  })

  /**
   * 是否禁用菜单类型切换
   */
  const disableMenuType = computed(() => {
    return false
  })

  /**
   * 重置表单数据
   */
  const resetForm = (): void => {
    formRef.value?.reset()
    form.menuType = 'menu'
  }

  /**
   * 从菜单树中查找节点的父级 ID（用于回显上级菜单）
   */
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

  /**
   * 加载表单数据（编辑模式）
   */
  const loadFormData = (): void => {
    if (!props.editData) return

    isEdit.value = true

    const row = props.editData
    form.menuType = row.meta?.isInnerPage ? 'inner' : 'menu'
    form.id = row.id || 0

    // 获取父级ID的逻辑：
    // 1. 优先使用 row.parent_id（如果存在且不为null/undefined）
    // 2. 如果没有，从菜单树中查找
    // 3. 如果都找不到，说明是顶级菜单
    let parentId = ''
    if (row.parent_id != null && row.parent_id !== undefined && row.parent_id !== '') {
      // 后端返回的 parent_id 有效
      parentId = String(row.parent_id)
    } else if (row.parentId != null && row.parentId !== undefined && row.parentId !== '') {
      // 兼容旧字段名
      parentId = String(row.parentId)
    } else {
      // 从菜单树中查找父级ID
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
    form.isHide = row.meta?.isInnerPage ? true : (row.meta?.isHide ?? false)
    form.isHideTab = row.meta?.isHideTab ?? false
    form.isEnable = row.meta?.isEnable ?? true
    form.link = row.meta?.link || ''
    form.isIframe = row.meta?.isIframe ?? false
    form.showBadge = row.meta?.showBadge ?? false
    form.showTextBadge = row.meta?.showTextBadge || ''
    form.fixedTab = row.meta?.fixedTab ?? false
    form.activePath = row.meta?.activePath || ''
    form.customParent = row.meta?.customParent || ''
    form.roles = row.meta?.roles || []
    form.isFullPage = row.meta?.isFullPage ?? false
    form.requiresTenantContext = row.meta?.requiresTenantContext ?? false
  }

  /**
   * 提交表单
   */
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

  /**
   * 取消操作
   */
  const handleCancel = (): void => {
    emit('update:visible', false)
  }

  /**
   * 对话框关闭后的回调
   */
  const handleClosed = (): void => {
    resetForm()
    isEdit.value = false
  }

  /**
   * 监听对话框显示状态
   */
  watch(
    () => props.visible,
    (newVal) => {
      if (newVal) {
        form.menuType = props.type === 'inner' ? 'inner' : props.type
        nextTick(() => {
          if (props.editData) {
            loadFormData()
          } else {
            form.parentId = props.initialParentId || ''
            if (form.menuType === 'inner') form.isHide = true
          }
        })
      }
    }
  )

  /**
   * 应用模板
   */
  const applyTemplate = (templateType: string) => {
    switch (templateType) {
      case 'outer':
        // 外链模板
        form.menuType = 'menu'
        form.isIframe = true
        form.link = 'https://www.example.com'
        form.path = '/external-link'
        form.component = ''
        break
      case 'top':
        // 顶层模板菜单
        form.menuType = 'menu'
        form.parentId = ''
        form.path = '/top-menu'
        form.component = '/index/index'
        form.isIframe = false
        form.link = ''
        break
      case 'sub':
        // 子菜单模板
        form.menuType = 'menu'
        form.path = 'sub-menu'
        form.component = '/system/sub-menu'
        form.isIframe = false
        form.link = ''
        form.isHide = false
        form.requiresTenantContext = false
        break
      case 'inner':
        // 内页模板
        form.menuType = 'inner'
        form.isHide = true
        form.path = 'inner-page'
        form.component = '/system/inner-page'
        form.isIframe = false
        form.link = ''
        form.requiresTenantContext = false
        break
    }
  }

  /**
   * 监听菜单类型变化
   */
  watch(
    () => props.type,
    (newType) => {
      if (props.visible) {
        form.menuType = newType
      }
    }
  )


</script>

<style lang="scss" scoped>
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
