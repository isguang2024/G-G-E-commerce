<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增功能包' : '编辑功能包'"
    width="36%"
    align-center
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="功能包编码" prop="packageKey">
        <ElInput v-model="form.packageKey" placeholder="例如 platform.system_admin" />
      </ElFormItem>
      <ElFormItem label="功能包名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入功能包名称" />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
      </ElFormItem>
      <ElFormItem label="上下文类型" prop="contextType">
        <ElSelect v-model="form.contextType" placeholder="请选择上下文类型" style="width: 100%">
          <ElOption label="平台" value="platform" />
          <ElOption label="团队" value="team" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSelect v-model="form.status" placeholder="请选择状态" style="width: 100%">
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="disabled" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="排序" prop="sortOrder">
        <ElInputNumber v-model="form.sortOrder" :min="0" :max="9999" style="width: 100%" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
  import { fetchCreateFeaturePackage, fetchUpdateFeaturePackage } from '@/api/system-manage'

  type PackageItem = Api.SystemManage.FeaturePackageItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    packageData?: Partial<PackageItem>
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    packageData: undefined
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const form = reactive({
    id: '',
    packageKey: '',
    name: '',
    description: '',
    contextType: 'team',
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    packageKey: [{ required: true, message: '请输入功能包编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入功能包名称', trigger: 'blur' }],
    contextType: [{ required: true, message: '请选择上下文类型', trigger: 'change' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  })

  function initForm() {
    if (props.dialogType === 'edit' && props.packageData) {
      Object.assign(form, {
        id: props.packageData.id || '',
        packageKey: props.packageData.packageKey || '',
        name: props.packageData.name || '',
        description: props.packageData.description || '',
        contextType: props.packageData.contextType || 'team',
        status: props.packageData.status || 'normal',
        sortOrder: props.packageData.sortOrder ?? 0
      })
      return
    }
    Object.assign(form, {
      id: '',
      packageKey: '',
      name: '',
      description: '',
      contextType: 'team',
      status: 'normal',
      sortOrder: 0
    })
  }

  watch(
    () => props.modelValue,
    (visible) => {
      if (visible) {
        initForm()
        nextTick(() => formRef.value?.clearValidate())
      }
    }
  )

  watch(
    () => [props.dialogType, props.packageData],
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
    await formRef.value.validate(async (valid) => {
      if (!valid) return
      const payload = {
        package_key: form.packageKey.trim(),
        name: form.name.trim(),
        description: form.description.trim(),
        context_type: form.contextType,
        status: form.status,
        sort_order: Number(form.sortOrder || 0)
      }
      try {
        if (props.dialogType === 'add') {
          await fetchCreateFeaturePackage(payload)
        } else {
          await fetchUpdateFeaturePackage(form.id, payload)
        }
        ElMessage.success(props.dialogType === 'add' ? '新增成功' : '修改成功')
        emit('success')
        handleClose()
      } catch (error: any) {
        ElMessage.error(error?.message || '保存失败')
      }
    })
  }
</script>
