<template>
  <ElDrawer
    v-model="visible"
    :title="dialogType === 'add' ? '新增功能包' : '编辑功能包'"
    size="36%"
    @close="handleClose"
    direction="rtl"
    class="config-drawer"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="功能包编码" prop="packageKey">
        <ElInput v-model="form.packageKey" placeholder="例如 platform.system_admin" />
      </ElFormItem>
      <ElFormItem label="功能包类型" prop="packageType">
        <ElSelect v-model="form.packageType" placeholder="请选择功能包类型" style="width: 100%">
          <ElOption label="基础包" value="base" />
          <ElOption label="组合包" value="bundle" />
        </ElSelect>
      </ElFormItem>
      <div v-if="form.packageType === 'bundle'" class="form-note">
        组合包不直接配置功能范围和菜单，保存后请通过“配置基础包”维护组合集合。
      </div>
      <ElFormItem label="功能包名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入功能包名称" />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
      </ElFormItem>
      <ElFormItem label="上下文类型" prop="contextType">
        <ElSelect v-model="form.contextType" placeholder="请选择上下文类型" style="width: 100%">
          <ElOption label="平台" value="platform" />
          <ElOption label="协作空间" value="team" />
          <ElOption label="通用" value="common" />
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
  </ElDrawer>
</template>

<script setup lang="ts">
  import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
  import { fetchCreateFeaturePackage, fetchUpdateFeaturePackage } from '@/api/system-manage'

  type PackageItem = Api.SystemManage.FeaturePackageItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    packageData?: Partial<PackageItem>
    appKey?: string
    defaultPackageType?: 'base' | 'bundle'
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    packageData: undefined,
    defaultPackageType: 'base'
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const form = reactive({
    id: '',
    packageKey: '',
    packageType: 'base' as 'base' | 'bundle',
    name: '',
    description: '',
    contextType: 'team',
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    packageKey: [{ required: true, message: '请输入功能包编码', trigger: 'blur' }],
    packageType: [{ required: true, message: '请选择功能包类型', trigger: 'change' }],
    name: [{ required: true, message: '请输入功能包名称', trigger: 'blur' }],
    contextType: [{ required: true, message: '请选择上下文类型', trigger: 'change' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  })

  function initForm() {
    if (props.dialogType === 'edit' && props.packageData) {
      Object.assign(form, {
        id: props.packageData.id || '',
        packageKey: props.packageData.packageKey || '',
        packageType:
          (props.packageData.packageType as 'base' | 'bundle') || props.defaultPackageType,
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
      packageType: props.defaultPackageType,
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
    if (!currentAppKey.value) {
      ElMessage.warning('缺少 app 上下文')
      return
    }
    await formRef.value.validate(async (valid) => {
      if (!valid) return
      const payload = {
        app_key: currentAppKey.value,
        package_key: form.packageKey.trim(),
        package_type: form.packageType,
        name: form.name.trim(),
        description: form.description.trim(),
        context_type: form.contextType,
        status: form.status,
        sort_order: Number(form.sortOrder || 0)
      }
      try {
        if (props.dialogType === 'add') {
          await fetchCreateFeaturePackage(payload)
          ElMessage.success('新增成功')
        } else {
          const stats = await fetchUpdateFeaturePackage(form.id, payload)
          ElMessage.success(formatRefreshMessage(stats))
        }
        emit('success')
        handleClose()
      } catch (error: any) {
        ElMessage.error(error?.message || '保存失败')
      }
    })
  }

  function formatRefreshMessage(stats?: Api.SystemManage.RefreshStats) {
    return `本次增量刷新：角色 ${stats?.roleCount || 0}、协作空间 ${stats?.teamCount || 0}、用户 ${stats?.userCount || 0}、耗时 ${stats?.elapsedMilliseconds || 0} ms`
  }
</script>

<style scoped lang="scss">
  .form-note {
    margin: -6px 0 12px 110px;
    color: #6b7280;
    line-height: 1.6;
  }
</style>

