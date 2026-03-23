<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增团队角色' : '编辑团队角色'"
    width="520px"
    align-center
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="角色名称" prop="roleName">
        <ElInput v-model="form.roleName" placeholder="请输入角色名称" />
      </ElFormItem>
      <ElFormItem label="角色编码" prop="roleCode">
        <ElInput v-model="form.roleCode" :disabled="dialogType === 'edit'" placeholder="请输入角色编码" />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入角色描述" />
      </ElFormItem>
      <ElFormItem label="状态">
        <ElSelect v-model="form.status" style="width: 100%">
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="suspended" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="排序">
        <ElInputNumber v-model="form.sortOrder" :min="0" :max="9999" style="width: 100%" />
      </ElFormItem>
      <ElFormItem label="优先级">
        <ElInputNumber v-model="form.priority" :min="0" :max="999" style="width: 100%" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { fetchCreateMyTeamRole, fetchUpdateMyTeamRole } from '@/api/team'

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    roleData?: Api.SystemManage.RoleListItem
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    roleData: undefined
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const formRef = ref<FormInstance>()
  const saving = ref(false)

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const form = reactive({
    roleId: '',
    roleName: '',
    roleCode: '',
    description: '',
    sortOrder: 0,
    priority: 0,
    status: 'normal'
  })

  const rules = reactive<FormRules>({
    roleName: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
    roleCode: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
    description: [{ required: true, message: '请输入角色描述', trigger: 'blur' }]
  })

  watch(
    () => [props.modelValue, props.roleData, props.dialogType] as const,
    ([open]) => {
      if (!open) return
      if (props.dialogType === 'edit' && props.roleData) {
        Object.assign(form, {
          roleId: props.roleData.roleId,
          roleName: props.roleData.roleName,
          roleCode: props.roleData.roleCode,
          description: props.roleData.description || '',
          sortOrder: props.roleData.sortOrder ?? 0,
          priority: props.roleData.priority ?? 0,
          status: props.roleData.status || 'normal'
        })
        return
      }
      resetForm()
    }
  )

  function resetForm() {
    Object.assign(form, {
      roleId: '',
      roleName: '',
      roleCode: '',
      description: '',
      sortOrder: 0,
      priority: 0,
      status: 'normal'
    })
  }

  function handleClose() {
    visible.value = false
    formRef.value?.resetFields()
    resetForm()
  }

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    saving.value = true
    try {
      const payload = {
        code: form.roleCode,
        name: form.roleName,
        description: form.description,
        sort_order: form.sortOrder,
        priority: form.priority,
        status: form.status
      }
      if (props.dialogType === 'add') {
        await fetchCreateMyTeamRole(payload)
      } else {
        await fetchUpdateMyTeamRole(form.roleId, payload)
      }
      ElMessage.success(props.dialogType === 'add' ? '团队角色已创建' : '团队角色已更新')
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error?.message || '保存团队角色失败')
    } finally {
      saving.value = false
    }
  }
</script>
