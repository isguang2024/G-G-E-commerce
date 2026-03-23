<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增角色' : '编辑角色'"
    width="36%"
    align-center
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="120px">
      <ElFormItem label="角色名称" prop="roleName">
        <ElInput v-model="form.roleName" placeholder="请输入角色名称" />
      </ElFormItem>
      <ElFormItem label="角色编码" prop="roleCode">
        <ElInput v-model="form.roleCode" placeholder="请输入角色编码" />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入角色描述"
        />
      </ElFormItem>
      <ElFormItem label="状态">
        <ElSelect v-model="form.status" placeholder="请选择状态" style="width: 100%">
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="suspended" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="排序">
        <ElInputNumber
          v-model="form.sortOrder"
          :min="0"
          :max="9999"
          placeholder="排序"
          style="width: 100%"
        />
      </ElFormItem>
      <ElFormItem label="优先级">
        <ElInputNumber
          v-model="form.priority"
          :min="0"
          :max="999"
          placeholder="优先级"
          style="width: 100%"
        />
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
  import { fetchCreateRole, fetchUpdateRole } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'

  type RoleListItem = Api.SystemManage.RoleListItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    roleData?: RoleListItem
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    roleData: undefined
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const rules = reactive<FormRules>({
    roleName: [
      { required: true, message: '请输入角色名称', trigger: 'blur' },
      { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
    ],
    roleCode: [
      { required: true, message: '请输入角色编码', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    description: [{ required: true, message: '请输入角色描述', trigger: 'blur' }]
  })

  const form = reactive({
    roleId: '',
    roleName: '',
    roleCode: '',
    description: '',
    createTime: '',
    sortOrder: 0,
    priority: 0,
    status: 'normal'
  })

  const initForm = () => {
    if (props.dialogType === 'edit' && props.roleData) {
      const roleData = props.roleData
      Object.assign(form, {
        roleId: roleData.roleId,
        roleName: roleData.roleName,
        roleCode: roleData.roleCode,
        description: roleData.description || '',
        createTime: roleData.createTime || '',
        sortOrder: roleData.sortOrder ?? 0,
        priority: roleData.priority || 0,
        status: roleData.status || 'normal'
      })
      return
    }

    Object.assign(form, {
      roleId: '',
      roleName: '',
      roleCode: '',
      description: '',
      createTime: '',
      sortOrder: 0,
      priority: 0,
      status: 'normal'
    })
  }

  watch(
    () => props.modelValue,
    (newVal) => {
      if (newVal) {
        initForm()
      }
    }
  )

  watch(
    () => props.roleData,
    (newData) => {
      if (newData && props.modelValue) {
        initForm()
      }
    },
    { deep: true }
  )

  const handleClose = () => {
    visible.value = false
    formRef.value?.resetFields()
    initForm()
  }

  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      const payload = {
        code: form.roleCode,
        name: form.roleName,
        description: form.description || '',
        sort_order: form.sortOrder ?? 0,
        priority: form.priority || 0,
        status: form.status || 'normal'
      }

      if (props.dialogType === 'add') {
        await fetchCreateRole(payload as any)
      } else {
        const roleId = typeof form.roleId === 'string' ? form.roleId : (form.roleId as any)?.toString?.() || ''
        if (!roleId) {
          ElMessage.error('缺少角色ID')
          return
        }
        await fetchUpdateRole(roleId, payload as any)
      }

      ElMessage.success(props.dialogType === 'add' ? '新增成功' : '修改成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      if (error?.message) ElMessage.error(error.message)
    }
  }
</script>
