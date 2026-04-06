<template>
  <ElDrawer
    v-model="visible"
    :title="dialogType === 'add' ? '新增功能权限' : '编辑功能权限'"
    size="640px"
    destroy-on-close
    @close="handleClose"
    direction="rtl"
    class="config-drawer"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="权限键" prop="permissionKey">
        <ElInput v-model="form.permissionKey" placeholder="例如 system.role.manage" />
      </ElFormItem>
      <ElFormItem label="模块分组" prop="moduleGroupId">
        <div class="group-select-row">
          <ElSelect
            v-model="form.moduleGroupId"
            filterable
            clearable
            style="width: 100%"
            popper-class="action-group-select-popper"
          >
            <ElOption
              v-for="item in moduleGroups"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </ElSelect>
          <ElButton text type="primary" @click="emit('open-group', 'module')">新建分组</ElButton>
        </div>
      </ElFormItem>
      <ElFormItem label="功能分组" prop="featureGroupId">
        <div class="group-select-row">
          <ElSelect
            v-model="form.featureGroupId"
            filterable
            clearable
            style="width: 100%"
            popper-class="action-group-select-popper"
          >
            <ElOption
              v-for="item in featureGroups"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </ElSelect>
          <ElButton text type="primary" @click="emit('open-group', 'feature')">新建分组</ElButton>
        </div>
      </ElFormItem>
      <ElFormItem label="上下文" prop="contextType">
        <ElSelect v-model="form.contextType" style="width: 100%">
          <ElOption label="平台" value="platform" />
          <ElOption label="协作空间" value="team" />
          <ElOption label="通用" value="common" />
        </ElSelect>
      </ElFormItem>
      <div class="context-hint">
        <span
          >平台：系统治理和平台后台能力，建议使用 `system.`、`platform.`、`collaboration_workspace.`
          前缀。</span
        >
        <span
          >协作空间：协作空间内授权能力，建议使用 `collaboration_workspace.`
          前缀或协作空间模块分组。</span
        >
        <span>通用：跨上下文业务能力，不要复用平台/协作空间专属前缀。</span>
      </div>
      <ElFormItem label="权限名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入名称" />
      </ElFormItem>
      <ElFormItem label="描述">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
      </ElFormItem>
      <ElFormItem prop="status">
        <template #label>
          <span class="label-help">
            <span>状态</span>
            <ElTooltip
              content="功能权限状态参与权限判断；停用后该权限键不再作为可用权限生效。"
              placement="top"
            >
              <ElIcon class="label-help-icon"><QuestionFilled /></ElIcon>
            </ElTooltip>
          </span>
        </template>
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
  </ElDrawer>
</template>

<script setup lang="ts">
  import { QuestionFilled } from '@element-plus/icons-vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreatePermissionAction, fetchUpdatePermissionAction } from '@/api/system-manage'
  import { ElIcon, ElMessage, ElTooltip } from 'element-plus'

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    actionData?: Api.SystemManage.PermissionActionItem
    moduleGroups?: Api.SystemManage.PermissionGroupItem[]
    featureGroups?: Api.SystemManage.PermissionGroupItem[]
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
    (e: 'open-group', value: 'module' | 'feature'): void
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
    moduleGroupId: '',
    featureGroupId: '',
    featureKind: 'business',
    contextType: 'common',
    name: '',
    description: '',
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    permissionKey: [{ required: true, message: '请输入权限键', trigger: 'blur' }],
    moduleGroupId: [{ required: true, message: '请选择模块分组', trigger: 'change' }],
    featureGroupId: [{ required: true, message: '请选择功能分组', trigger: 'change' }],
    contextType: [{ required: true, message: '请选择上下文', trigger: 'change' }],
    name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  })

  function initForm() {
    if (props.dialogType === 'edit' && props.actionData) {
      Object.assign(form, {
        id: props.actionData.id,
        permissionKey: props.actionData.permissionKey || '',
        moduleCode: props.actionData.moduleCode || '',
        moduleGroupId: props.actionData.moduleGroupId || '',
        featureGroupId: props.actionData.featureGroupId || '',
        featureKind: props.actionData.featureKind || 'system',
        contextType: props.actionData.contextType || 'common',
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
      moduleGroupId: props.moduleGroups?.[0]?.id || '',
      featureGroupId: props.featureGroups?.[0]?.id || '',
      featureKind: 'business',
      contextType: 'common',
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
        module_group_id: form.moduleGroupId,
        feature_group_id: form.featureGroupId,
        context_type: form.contextType,
        feature_kind:
          props.featureGroups?.find((item) => item.id === form.featureGroupId)?.code ||
          form.featureKind,
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

<style scoped>
  .label-help {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .label-help-icon {
    color: var(--el-text-color-placeholder);
    cursor: help;
    font-size: 14px;
  }

  .group-select-row {
    display: flex;
    gap: 8px;
    width: 100%;
  }

  .context-hint {
    margin: -4px 0 12px 110px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.5;
  }
</style>
