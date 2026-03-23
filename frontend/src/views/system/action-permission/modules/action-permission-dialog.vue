<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增功能权限' : '编辑功能权限'"
    width="640px"
    destroy-on-close
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="来源">
        <ElTag :type="sourceTagType">{{ sourceLabel }}</ElTag>
      </ElFormItem>
      <ElFormItem label="权限键" prop="permissionKey">
        <ElInput v-model="form.permissionKey" placeholder="例如 system.role.manage" />
      </ElFormItem>
      <ElFormItem label="模块归属" prop="moduleCode">
        <ElAutocomplete
          v-model="form.moduleCode"
          :fetch-suggestions="queryModuleSuggestions"
          clearable
          placeholder="例如 order_center"
        />
      </ElFormItem>
      <ElFormItem label="功能归属" prop="featureKind">
        <ElSelect v-model="form.featureKind" style="width: 100%">
          <ElOption label="系统功能" value="system" />
          <ElOption label="业务功能" value="business" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="上下文" prop="contextType">
        <ElSelect v-model="form.contextType" style="width: 100%">
          <ElOption label="平台" value="platform" />
          <ElOption label="团队" value="team" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="权限名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入名称" />
      </ElFormItem>
      <ElFormItem label="描述">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSelect v-model="form.status" style="width: 100%">
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="suspended" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="排序">
        <ElInputNumber v-model="form.sortOrder" :min="0" :max="9999" style="width: 100%" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreatePermissionAction, fetchUpdatePermissionAction } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    actionData?: Api.SystemManage.PermissionActionItem
    moduleOptions?: string[]
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    actionData: undefined,
    moduleOptions: () => []
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const form = reactive({
    id: '',
    permissionKey: '',
    moduleCode: '',
    source: 'business',
    featureKind: 'business',
    contextType: 'team',
    name: '',
    description: '',
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    permissionKey: [{ required: true, message: '请输入权限键', trigger: 'blur' }],
    moduleCode: [{ required: true, message: '请输入模块归属', trigger: 'blur' }],
    featureKind: [{ required: true, message: '请选择功能归属', trigger: 'change' }],
    contextType: [{ required: true, message: '请选择上下文', trigger: 'change' }],
    name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  })

  const permissionKeyPreview = computed(() => {
    return form.permissionKey.trim()
  })

  const sourceLabel = computed(() => {
    if (form.source === 'api') return '接口自动注册'
    if (form.source === 'system') return '系统内置'
    return '业务定义'
  })

  const sourceTagType = computed(() => {
    if (form.source === 'api') return 'success'
    if (form.source === 'system') return 'info'
    return 'warning'
  })

  function queryModuleSuggestions(
    queryString: string,
    cb: (items: Array<{ value: string }>) => void
  ) {
    const keyword = queryString.trim().toLowerCase()
    const suggestions = (props.moduleOptions || [])
      .filter((item) => !keyword || item.toLowerCase().includes(keyword))
      .slice(0, 12)
      .map((value) => ({ value }))
    cb(suggestions)
  }

  function initForm() {
    if (props.dialogType === 'edit' && props.actionData) {
      Object.assign(form, {
        id: props.actionData.id,
        permissionKey: props.actionData.permissionKey || `${props.actionData.resourceCode}:${props.actionData.actionCode}`,
        moduleCode: props.actionData.moduleCode || props.actionData.resourceCode,
        source: props.actionData.source || 'business',
        featureKind: props.actionData.featureKind || 'system',
        contextType: props.actionData.contextType || 'team',
        name: props.actionData.name,
        description: props.actionData.description || '',
        status: props.actionData.status || 'normal',
        sortOrder: props.actionData.sortOrder ?? 0
      })
      return
    }
    Object.assign(form, {
      id: '',
      permissionKey: '',
      moduleCode: '',
      source: 'business',
      featureKind: 'business',
      contextType: 'team',
      name: '',
      description: '',
      status: 'normal',
      sortOrder: 0
    })
  }

  watch(
    () => [props.modelValue, props.actionData, props.dialogType],
    async ([opened]) => {
      if (!opened) return
      initForm()
    },
    { deep: true }
  )

  function handleClose() {
    visible.value = false
    formRef.value?.resetFields()
  }

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitting.value = true
    try {
      const payload = {
        permission_key: form.permissionKey.trim(),
        module_code: form.moduleCode.trim(),
        context_type: form.contextType,
        feature_kind: form.featureKind,
        name: form.name.trim(),
        description: form.description.trim(),
        status: form.status,
        sort_order: form.sortOrder ?? 0
      }
      if (props.dialogType === 'add') {
        await fetchCreatePermissionAction(payload)
      } else {
        await fetchUpdatePermissionAction(form.id, payload)
      }
      ElMessage.success(props.dialogType === 'add' ? '新增成功' : '更新成功')
      emit('success')
      handleClose()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    } finally {
      submitting.value = false
    }
  }
</script>
