<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增功能权限' : '编辑功能权限'"
    width="640px"
    destroy-on-close
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
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
      <ElFormItem label="依赖团队">
        <ElSwitch v-model="form.requiresTenantContext" />
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
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    actionData: undefined
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
    name: '',
    description: '',
    scopeId: '',
    requiresTenantContext: false,
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    resourceCode: [{ required: true, message: '请输入资源编码', trigger: 'blur' }],
    actionCode: [{ required: true, message: '请输入动作编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
    scopeId: [{ required: true, message: '请选择作用域', trigger: 'change' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  })

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
        name: props.actionData.name,
        description: props.actionData.description || '',
        scopeId,
        requiresTenantContext: props.actionData.requiresTenantContext,
        status: props.actionData.status || 'normal',
        sortOrder: props.actionData.sortOrder ?? 0
      })
      return
    }
    Object.assign(form, {
      id: '',
      resourceCode: '',
      actionCode: '',
      name: '',
      description: '',
      scopeId: scopeList.value[0]?.scopeId || '',
      requiresTenantContext: false,
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
        name: form.name.trim(),
        description: form.description.trim(),
        scope_id: form.scopeId,
        requires_tenant_context: form.requiresTenantContext,
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
