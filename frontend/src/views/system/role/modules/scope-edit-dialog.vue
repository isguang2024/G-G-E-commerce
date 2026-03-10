<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增作用域' : '编辑作用域'"
    width="30%"
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
      <ElFormItem label="描述" prop="description">
        <ElInput
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入作用域描述"
        />
      </ElFormItem>
      <ElFormItem label="排序" prop="sortOrder">
        <ElInputNumber v-model="form.sortOrder" :min="0" placeholder="请输入排序值" style="width: 100%" />
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
  import { fetchCreateScope, fetchUpdateScope } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'

  type ScopeListItem = Api.SystemManage.ScopeListItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    scopeData?: ScopeListItem
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    scopeData: undefined
  })

  const emit = defineEmits<Emits>()

  const formRef = ref<FormInstance>()

  /**
   * 弹窗显示状态双向绑定
   */
  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  /**
   * 表单验证规则
   */
  const rules = reactive<FormRules>({
    scopeCode: [
      { required: true, message: '请输入作用域编码', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    scopeName: [
      { required: true, message: '请输入作用域名称', trigger: 'blur' },
      { min: 2, max: 100, message: '长度在 2 到 100 个字符', trigger: 'blur' }
    ]
  })

  /**
   * 表单数据
   */
  const form = reactive<ScopeListItem & { sortOrder?: number }>({
    scopeId: '',
    scopeCode: '',
    scopeName: '',
    description: '',
    sortOrder: 0
  })

  /**
   * 监听弹窗打开，初始化表单数据
   */
  watch(
    () => props.modelValue,
    (newVal) => {
      if (newVal) initForm()
    }
  )

  /**
   * 监听作用域数据变化，更新表单
   */
  watch(
    () => props.scopeData,
    (newData) => {
      if (newData && props.modelValue) initForm()
    },
    { deep: true }
  )

  /**
   * 初始化表单数据
   */
  const initForm = () => {
    if (props.dialogType === 'edit' && props.scopeData) {
      Object.assign(form, props.scopeData)
    } else {
      Object.assign(form, {
        scopeId: '',
        scopeCode: '',
        scopeName: '',
        description: '',
        sortOrder: 0
      })
    }
  }

  /**
   * 关闭弹窗并重置表单
   */
  const handleClose = () => {
    visible.value = false
    formRef.value?.resetFields()
  }

  /**
   * 提交表单
   */
  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      if (props.dialogType === 'add') {
        await fetchCreateScope({
          code: form.scopeCode,
          name: form.scopeName,
          description: form.description || '',
          sort_order: form.sortOrder || 0
        })
      } else {
        const scopeId = typeof form.scopeId === 'string' ? form.scopeId : (form.scopeId as any)?.toString?.() || ''
        if (!scopeId) {
          ElMessage.error('缺少作用域ID')
          return
        }
        await fetchUpdateScope(scopeId, {
          name: form.scopeName,
          description: form.description || '',
          sort_order: form.sortOrder || 0
        })
      }
      ElMessage.success(props.dialogType === 'add' ? '新增成功' : '修改成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      if (error?.message) ElMessage.error(error.message)
    }
  }
</script>
