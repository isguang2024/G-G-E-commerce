<template>
  <ElDialog
    v-model="visible"
    :title="form.id ? `编辑${groupTypeLabel}` : `新建${groupTypeLabel}`"
    width="560px"
    destroy-on-close
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <ElFormItem label="分组编码" prop="code">
        <ElInput v-model="form.code" placeholder="例如 role 或 system_feature" />
      </ElFormItem>
      <ElFormItem label="分组名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入名称" />
      </ElFormItem>
      <ElFormItem label="英文名称">
        <ElInput v-model="form.nameEn" placeholder="可选" />
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
      <ElFormItem label="说明">
        <ElInput v-model="form.description" type="textarea" :rows="3" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { fetchCreatePermissionGroup, fetchUpdatePermissionGroup } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    groupType: 'module' | 'feature'
    groupData?: Api.SystemManage.PermissionGroupItem
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    groupType: 'module',
    groupData: undefined
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const groupTypeLabel = computed(() => (props.groupType === 'module' ? '模块分组' : '功能分组'))
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const form = reactive({
    id: '',
    code: '',
    name: '',
    nameEn: '',
    description: '',
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    code: [{ required: true, message: '请输入分组编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }]
  })

  function initForm() {
    Object.assign(form, {
      id: props.groupData?.id || '',
      code: props.groupData?.code || '',
      name: props.groupData?.name || '',
      nameEn: props.groupData?.nameEn || '',
      description: props.groupData?.description || '',
      status: props.groupData?.status || 'normal',
      sortOrder: props.groupData?.sortOrder ?? 0
    })
  }

  watch(
    () => [props.modelValue, props.groupData, props.groupType],
    ([opened]) => {
      if (!opened) return
      initForm()
    },
    { deep: true }
  )

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitting.value = true
    try {
      const payload = {
        code: form.code.trim(),
        name: form.name.trim(),
        name_en: form.nameEn.trim(),
        description: form.description.trim(),
        group_type: props.groupType,
        status: form.status,
        sort_order: form.sortOrder ?? 0
      }
      if (form.id) {
        await fetchUpdatePermissionGroup(form.id, payload)
      } else {
        await fetchCreatePermissionGroup(payload)
      }
      ElMessage.success('分组保存成功')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '分组保存失败')
    } finally {
      submitting.value = false
    }
  }
</script>
