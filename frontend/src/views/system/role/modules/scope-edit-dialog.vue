<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增作用域' : '编辑作用域'"
    width="520px"
    align-center
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="120px">
      <ElFormItem label="作用域编码" prop="scopeCode">
        <ElInput v-model="form.scopeCode" placeholder="请输入作用域编码" :disabled="dialogType === 'edit'" />
      </ElFormItem>
      <ElFormItem label="作用域名称" prop="scopeName">
        <ElInput v-model="form.scopeName" placeholder="请输入作用域名称" />
      </ElFormItem>
      <ElFormItem label="上下文类型" prop="contextKind">
        <ElSelect v-model="form.contextKind" style="width: 100%">
          <ElOption label="全局" value="global" />
          <ElOption label="团队" value="tenant" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="数据权限编码" prop="dataPermissionCode">
        <ElInput v-model="form.dataPermissionCode" placeholder="例如：team、department，不填则不参与数据权限" />
      </ElFormItem>
      <ElFormItem label="数据权限名称" prop="dataPermissionName">
        <ElInput v-model="form.dataPermissionName" placeholder="例如：当前团队、当前部门" />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入作用域描述" />
      </ElFormItem>
      <ElFormItem label="排序" prop="sortOrder">
        <ElInputNumber v-model="form.sortOrder" :min="0" style="width: 100%" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { fetchCreateScope, fetchUpdateScope } from '@/api/system-manage'

  type ScopeListItem = Api.SystemManage.ScopeListItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    scopeData?: ScopeListItem
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    scopeData: undefined
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const formRef = ref<FormInstance>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const rules = reactive<FormRules>({
    scopeCode: [
      { required: true, message: '请输入作用域编码', trigger: 'blur' },
      { min: 2, max: 50, message: '长度应为 2 到 50 个字符', trigger: 'blur' }
    ],
    scopeName: [
      { required: true, message: '请输入作用域名称', trigger: 'blur' },
      { min: 2, max: 100, message: '长度应为 2 到 100 个字符', trigger: 'blur' }
    ],
    contextKind: [{ required: true, message: '请选择上下文类型', trigger: 'change' }]
  })

  const form = reactive({
    scopeId: '',
    scopeCode: '',
    scopeName: '',
    description: '',
    contextKind: 'global' as 'global' | 'tenant',
    dataPermissionCode: '',
    dataPermissionName: '',
    sortOrder: 0
  })

  function initForm() {
    if (props.dialogType === 'edit' && props.scopeData) {
      Object.assign(form, {
        scopeId: props.scopeData.scopeId,
        scopeCode: props.scopeData.scopeCode,
        scopeName: props.scopeData.scopeName,
        description: props.scopeData.description || '',
        contextKind: (props.scopeData.contextKind || 'global') as 'global' | 'tenant',
        dataPermissionCode: props.scopeData.dataPermissionCode || '',
        dataPermissionName: props.scopeData.dataPermissionName || '',
        sortOrder: props.scopeData.sortOrder || 0
      })
      return
    }

    Object.assign(form, {
      scopeId: '',
      scopeCode: '',
      scopeName: '',
      description: '',
      contextKind: 'global',
      dataPermissionCode: '',
      dataPermissionName: '',
      sortOrder: 0
    })
  }

  watch(
    () => props.modelValue,
    (value) => {
      if (value) initForm()
    }
  )

  watch(
    () => props.scopeData,
    () => {
      if (props.modelValue) initForm()
    },
    { deep: true }
  )

  function handleClose() {
    visible.value = false
    formRef.value?.resetFields()
    initForm()
  }

  async function handleSubmit() {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      const payload = {
        code: form.scopeCode,
        name: form.scopeName,
        description: form.description || '',
        context_kind: form.contextKind,
        data_permission_code: form.dataPermissionCode || '',
        data_permission_name: form.dataPermissionName || '',
        sort_order: form.sortOrder || 0
      }

      if (props.dialogType === 'add') {
        await fetchCreateScope(payload)
      } else {
        if (!form.scopeId) {
          ElMessage.error('缺少作用域ID')
          return
        }
        await fetchUpdateScope(form.scopeId, payload)
      }

      ElMessage.success(props.dialogType === 'add' ? '新增成功' : '修改成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      if (error?.message) ElMessage.error(error.message)
    }
  }
</script>
