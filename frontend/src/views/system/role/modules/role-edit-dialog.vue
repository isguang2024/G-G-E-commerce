<template>
  <ElDrawer
    v-model="visible"
    :title="dialogType === 'add' ? '新增角色' : '编辑角色'"
    size="36%"
    @close="handleClose"
    direction="rtl"
    class="config-drawer"
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
      <ElFormItem label="生效 App">
        <AppKeySelect
          v-model="form.appKeys"
          multiple
          clearable
          placeholder="留空表示全局通用"
          :eager="false"
        />
        <div class="form-help-text"
          >未配置 App 时角色对所有 App 通用；配置后仅在这些 App 下生效。</div
        >
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
      <ElFormItem label="自定义参数(JSON)" prop="customParamsText">
        <ElInput
          v-model="form.customParamsText"
          type="textarea"
          :rows="6"
          placeholder='请输入 JSON 对象，例如：{"dataScope":"collaboration","editable":true}'
        />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import AppKeySelect from '@/components/business/app/AppKeySelect.vue'
  import { fetchCreateRole, fetchUpdateRole } from '@/domains/governance/api'
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
    description: [{ required: true, message: '请输入角色描述', trigger: 'blur' }],
    customParamsText: [
      {
        validator: (_rule, value, callback) => {
          const raw = `${value || ''}`.trim()
          if (!raw) {
            callback()
            return
          }
          try {
            const parsed = JSON.parse(raw)
            if (parsed === null || Array.isArray(parsed) || typeof parsed !== 'object') {
              callback(new Error('自定义参数必须是 JSON 对象'))
              return
            }
            callback()
          } catch {
            callback(new Error('JSON 格式不正确'))
          }
        },
        trigger: 'blur'
      }
    ]
  })

  const form = reactive({
    roleId: '',
    roleName: '',
    roleCode: '',
    description: '',
    appKeys: [] as string[],
    createTime: '',
    sortOrder: 0,
    priority: 0,
    status: 'normal',
    customParamsText: '{}'
  })

  const formatCustomParams = (value?: Record<string, any>) => {
    const target = value && typeof value === 'object' && !Array.isArray(value) ? value : {}
    return JSON.stringify(target, null, 2)
  }

  const initForm = () => {
    if (props.dialogType === 'edit' && props.roleData) {
      const roleData = props.roleData
      Object.assign(form, {
        roleId: roleData.roleId,
        roleName: roleData.roleName,
        roleCode: roleData.roleCode,
        description: roleData.description || '',
        appKeys: Array.isArray(roleData.appKeys) ? [...roleData.appKeys] : [],
        createTime: roleData.createTime || '',
        sortOrder: roleData.sortOrder ?? 0,
        priority: roleData.priority || 0,
        status: roleData.status || 'normal',
        customParamsText: formatCustomParams(roleData.customParams)
      })
      return
    }

    Object.assign(form, {
      roleId: '',
      roleName: '',
      roleCode: '',
      description: '',
      appKeys: [],
      createTime: '',
      sortOrder: 0,
      priority: 0,
      status: 'normal',
      customParamsText: '{}'
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
      const parsedCustomParams = JSON.parse(form.customParamsText || '{}')
      const payload = {
        code: form.roleCode,
        name: form.roleName,
        description: form.description || '',
        app_keys: [...form.appKeys],
        sort_order: form.sortOrder ?? 0,
        priority: form.priority || 0,
        custom_params: parsedCustomParams,
        status: form.status || 'normal'
      }

      if (props.dialogType === 'add') {
        await fetchCreateRole(payload)
      } else {
        const roleId = `${form.roleId || ''}`.trim()
        if (!roleId) {
          ElMessage.error('缺少角色ID')
          return
        }
        await fetchUpdateRole(roleId, payload)
      }

      ElMessage.success(props.dialogType === 'add' ? '新增成功' : '修改成功')
      emit('success')
      handleClose()
    } catch (error) {
      if (error instanceof Error && error.message) ElMessage.error(error.message)
    }
  }
</script>

<style scoped>
  .form-help-text {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }
</style>
