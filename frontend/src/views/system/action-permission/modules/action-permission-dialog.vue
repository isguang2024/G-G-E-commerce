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
      <ElFormItem label="权限键">
        <ElInput :model-value="permissionKeyPreview" disabled />
      </ElFormItem>
      <ElFormItem label="分类" prop="category">
        <ElAutocomplete
          v-model="form.category"
          :fetch-suggestions="queryCategorySuggestions"
          clearable
          placeholder="输入或选择历史分类"
        />
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
      <ElFormItem label="资源编码" prop="resourceCode">
        <ElInput v-model="form.resourceCode" placeholder="例如 team_member" />
      </ElFormItem>
      <ElFormItem label="动作编码" prop="actionCode">
        <ElInput v-model="form.actionCode" placeholder="例如 assign_action" />
      </ElFormItem>
      <ElFormItem label="权限名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入名称" />
      </ElFormItem>
      <ElFormItem label="描述">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
      </ElFormItem>
      <ElFormItem label="作用域" prop="scopeId">
        <ElSelect v-model="form.scopeId" style="width: 100%" :loading="scopeLoading">
          <ElOption
            v-for="scope in scopeList"
            :key="scope.scopeId"
            :label="scope.scopeName"
            :value="scope.scopeId"
          >
            <span>{{ scope.scopeName }}</span>
            <span style="color: #8492a6; font-size: 13px; margin-left: 8px"
              >({{ scope.scopeCode }})</span
            >
          </ElOption>
        </ElSelect>
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
  import {
    fetchGetAllScopes,
    fetchCreatePermissionAction,
    fetchUpdatePermissionAction
  } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    actionData?: Api.SystemManage.PermissionActionItem
    categoryOptions?: string[]
    moduleOptions?: string[]
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    actionData: undefined,
    categoryOptions: () => [],
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
  const scopeLoading = ref(false)
  const scopeList = ref<Api.SystemManage.ScopeListItem[]>([])
  const form = reactive({
    id: '',
    resourceCode: '',
    actionCode: '',
    moduleCode: '',
    category: '',
    source: 'business',
    featureKind: 'business',
    name: '',
    description: '',
    scopeId: '',
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    category: [{ max: 100, message: '分类长度不能超过 100 个字符', trigger: 'blur' }],
    moduleCode: [{ required: true, message: '请输入模块归属', trigger: 'blur' }],
    featureKind: [{ required: true, message: '请选择功能归属', trigger: 'change' }],
    resourceCode: [{ required: true, message: '请输入资源编码', trigger: 'blur' }],
    actionCode: [{ required: true, message: '请输入动作编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
    scopeId: [{ required: true, message: '请选择作用域', trigger: 'change' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  })

  const permissionKeyPreview = computed(() => {
    const resourceCode = form.resourceCode.trim()
    const actionCode = form.actionCode.trim()
    if (!resourceCode && !actionCode) {
      return ''
    }
    return `${resourceCode || 'resource'}:${actionCode || 'action'}`
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

  function queryCategorySuggestions(
    queryString: string,
    cb: (items: Array<{ value: string }>) => void
  ) {
    const keyword = queryString.trim().toLowerCase()
    const suggestions = (props.categoryOptions || [])
      .filter((item) => !keyword || item.toLowerCase().includes(keyword))
      .slice(0, 12)
      .map((value) => ({ value }))
    cb(suggestions)
  }

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

  async function loadScopes() {
    try {
      scopeLoading.value = true
      scopeList.value = await fetchGetAllScopes()
    } finally {
      scopeLoading.value = false
    }
  }

  function initForm() {
    if (props.dialogType === 'edit' && props.actionData) {
      let scopeId = props.actionData.scopeId || ''
      if (!scopeId && props.actionData.scopeCode && scopeList.value.length > 0) {
        const matchedScope = scopeList.value.find(
          (scope) => scope.scopeCode === props.actionData?.scopeCode
        )
        scopeId = matchedScope?.scopeId || ''
      }
      Object.assign(form, {
        id: props.actionData.id,
        resourceCode: props.actionData.resourceCode,
        actionCode: props.actionData.actionCode,
        moduleCode: props.actionData.moduleCode || props.actionData.resourceCode,
        category: props.actionData.category || '',
        source: props.actionData.source || 'business',
        featureKind: props.actionData.featureKind || 'system',
        name: props.actionData.name,
        description: props.actionData.description || '',
        scopeId,
        status: props.actionData.status || 'normal',
        sortOrder: props.actionData.sortOrder ?? 0
      })
      return
    }
    Object.assign(form, {
        id: '',
        resourceCode: '',
        actionCode: '',
        moduleCode: '',
        category: '',
      source: 'business',
      featureKind: 'business',
      name: '',
      description: '',
      scopeId: scopeList.value[0]?.scopeId || '',
      status: 'normal',
      sortOrder: 0
    })
  }

  watch(
    () => [props.modelValue, props.actionData, props.dialogType],
    async ([opened]) => {
      if (!opened) return
      if (scopeList.value.length === 0) {
        await loadScopes()
      }
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
        resource_code: form.resourceCode.trim(),
        action_code: form.actionCode.trim(),
        module_code: form.moduleCode.trim(),
        category: form.category.trim(),
        feature_kind: form.featureKind,
        name: form.name.trim(),
        description: form.description.trim(),
        scope_id: form.scopeId,
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
