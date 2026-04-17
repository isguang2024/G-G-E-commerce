<template>
  <ElDrawer
    v-model="visible"
    :title="dialogTitle"
    size="760px"
    direction="rtl"
    class="page-display-group-drawer config-drawer"
    destroy-on-close
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <div class="dialog-intro">
        <div class="dialog-intro__main">
          <div class="dialog-intro__title">普通分组配置说明</div>
          <div class="dialog-intro__desc"
            >普通分组只用于页面管理列表归类，不参与路径、权限、菜单高亮和面包屑继承。</div
          >
        </div>
        <ElButton text type="primary" @click="showExamples = !showExamples">
          {{ showExamples ? '收起示例' : '查看示例' }}
        </ElButton>
        <div v-if="showExamples" class="dialog-intro__examples">
          <div v-for="item in examples" :key="item" class="dialog-intro__example">{{ item }}</div>
        </div>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div class="form-section__title">基础信息</div>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="分组名称" prop="name">
              <template #label>
                <PageFieldLabel
                  label="分组名称"
                  help="给人看的普通分组名称，只显示在页面管理列表和普通分组选择器里。"
                />
              </template>
              <ElInput v-model="form.name" placeholder="请输入普通分组名称" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="可见范围" prop="visibilityScope">
              <template #label>
                <PageFieldLabel
                  label="可见范围"
                  help="普通分组默认在当前 App 下全局可见；只有切到指定空间时才写入空间绑定。"
                />
              </template>
              <ElRadioGroup v-model="form.visibilityScope">
                <ElRadioButton label="app">App 全局</ElRadioButton>
                <ElRadioButton label="spaces">指定空间</ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>
          </ElCol>
          <ElCol v-if="form.visibilityScope === 'spaces'" :span="12">
            <ElFormItem label="开放空间" prop="menuSpaceKeys">
              <template #label>
                <PageFieldLabel
                  label="开放空间"
                  help="只控制这个普通分组在哪些菜单空间里可见，不改变普通分组本身的定义。"
                />
              </template>
              <ElSelect
                v-model="form.menuSpaceKeys"
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

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="排序" prop="sortOrder">
              <template #label>
                <PageFieldLabel
                  label="排序"
                  help="普通分组在页面列表里的展示顺序，数字越小越靠前。"
                />
              </template>
              <ElInputNumber v-model="form.sortOrder" :min="0" :step="1" style="width: 100%" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="状态" prop="status">
              <template #label>
                <PageFieldLabel
                  label="状态"
                  help="停用后普通分组保留数据，但不会再作为页面管理列表里的有效归类节点。"
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
  import { fetchCreatePage, fetchUpdatePage } from '@/domains/governance/api'
  import PageFieldLabel from './page-field-label.vue'

  type PageItem = Api.SystemManage.PageItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit' | 'copy'
    pageData?: Partial<PageItem>
    appKey?: string
    menuSpaces?: Api.SystemManage.MenuSpaceItem[]
    // 仅作为可见性/候选加载视角使用，不代表页面必须绑定该空间。
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
    initialParentPageKey: '',
    initialParentMenuId: '',
    initialPageType: 'display_group',
    defaultData: undefined
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const showExamples = ref(false)

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const dialogTitle = computed(() => {
    if (props.dialogType === 'copy') return '复制普通分组'
    return props.dialogType === 'edit' ? '编辑普通分组' : '新增普通分组'
  })

  const form = reactive({
    id: '',
    pageKey: '',
    name: '',
    visibilityScope: 'app',
    menuSpaceKeys: [] as string[],
    sortOrder: 0,
    status: 'normal'
  })

  const rules = reactive<FormRules>({
    name: [{ required: true, message: '请输入普通分组名称', trigger: 'blur' }]
  })

  const examples = [
    '例 1：普通分组名称=仪表盘示例，仅用于把相关页面收拢到同一个展示分组下。',
    '例 2：页面仍然可以直接挂到 /dashboard 菜单，普通分组只影响页面管理列表中的归类展示。',
    '例 3：停用普通分组后页面数据保留，只是不再作为有效归类节点显示。'
  ]
  const menuSpaceOptions = computed(() =>
    (props.menuSpaces || []).map((item) => ({
      label: item.isDefault ? `${item.name}（默认）` : item.name,
      value: item.menuSpaceKey
    }))
  )

  function initForm() {
    if (props.dialogType === 'edit' && props.pageData) {
      const menuSpaceKeys = Array.isArray((props.pageData as any).menuSpaceKeys)
        ? (props.pageData as any).menuSpaceKeys
        : Array.isArray(props.pageData.menuSpaceKeys)
          ? props.pageData.menuSpaceKeys
          : []
      Object.assign(form, {
        id: props.pageData.id || '',
        pageKey: props.pageData.pageKey || '',
        name: props.pageData.name || '',
        visibilityScope:
          `${props.pageData.visibilityScope || props.pageData.spaceScope || ''}`.trim() || 'app',
        menuSpaceKeys: menuSpaceKeys,
        sortOrder: props.pageData.sortOrder ?? 0,
        status: props.pageData.status || 'normal'
      })
      return
    }

    Object.assign(form, {
      id: '',
      pageKey: props.defaultData?.pageKey || '',
      name: props.defaultData?.name || '',
      visibilityScope:
        `${props.defaultData?.visibilityScope || props.defaultData?.spaceScope || ''}`.trim() ||
        'app',
      menuSpaceKeys: (props.defaultData as any)?.menuSpaceKeys || [],
      sortOrder: props.defaultData?.sortOrder ?? 0,
      status: props.defaultData?.status || 'normal'
    })
  }

  watch(
    () => form.visibilityScope,
    (value) => {
      if (value !== 'spaces') {
        form.menuSpaceKeys = []
      }
    }
  )

  async function prepareDialog() {
    await nextTick()
    initForm()
    await nextTick()
    formRef.value?.clearValidate()
  }

  watch(
    () => props.modelValue,
    async (value) => {
      if (!value) return
      await prepareDialog()
    }
  )

  watch(
    () => [props.dialogType, props.pageData, props.defaultData],
    () => {
      if (!props.modelValue) return
      nextTick(() => {
        initForm()
        formRef.value?.clearValidate()
      })
    },
    { deep: true }
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
      const visibilityScope = form.visibilityScope === 'spaces' ? 'spaces' : 'app'
      const payload: Api.SystemManage.PageSaveParams = {
        app_key: `${props.appKey || ''}`.trim(),
        page_key: props.dialogType === 'edit' ? form.pageKey.trim() : '',
        name: form.name.trim(),
        route_name: props.dialogType === 'edit' ? form.pageKey.trim() : '',
        route_path: '',
        component: '',
        page_type: 'display_group',
        source:
          props.dialogType === 'edit'
            ? `${props.pageData?.source || 'manual'}`
            : `${props.defaultData?.source || 'manual'}`,
        module_key: '',
        visibility_scope: visibilityScope,
        menu_space_keys: visibilityScope === 'spaces' ? form.menuSpaceKeys : [],
        sort_order: form.sortOrder,
        parent_menu_id: '',
        parent_page_key: '',
        display_group_key: '',
        active_menu_path: '',
        breadcrumb_mode: 'inherit_menu',
        access_mode: 'inherit',
        permission_key: '',
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
    margin-bottom: 4px;
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
</style>
