<template>
  <ElDialog
    v-model="visible"
    :title="dialogType === 'add' ? '新增角色' : '编辑角色'"
    width="30%"
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
      <ElFormItem label="作用域" prop="scopeId">
        <ElSelect
          v-model="form.scopeId"
          placeholder="请选择作用域"
          style="width: 100%"
          :loading="scopeLoading"
          clearable
        >
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
  import { fetchCreateRole, fetchUpdateRole, fetchGetAllScopes } from '@/api/system-manage'
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
    roleName: [
      { required: true, message: '请输入角色名称', trigger: 'blur' },
      { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
    ],
    roleCode: [
      { required: true, message: '请输入角色编码', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    description: [{ required: true, message: '请输入角色描述', trigger: 'blur' }],
    scopeId: [{ required: true, message: '请选择作用域', trigger: 'change' }]
  })

  /**
   * 作用域列表
   */
  const scopeList = ref<Api.SystemManage.ScopeListItem[]>([])
  const scopeLoading = ref(false)

  /**
   * 表单数据
   */
  const form = reactive<
    RoleListItem & { scopeId?: string; priority?: number; status?: string; sortOrder?: number }
  >({
    roleId: '',
    roleName: '',
    roleCode: '',
    description: '',
    createTime: '',
    scopeId: '',
    sortOrder: 0,
    priority: 0,
    status: 'normal'
  })

  /**
   * 加载作用域列表
   */
  const loadScopes = async () => {
    try {
      scopeLoading.value = true
      const res = await fetchGetAllScopes()
      // 后端返回格式: { code: 0, data: { records: [...] } }
      // HTTP工具返回 res.data.data，即 { records: [...] }
      // 所以这里使用 res.records
      const data = res as any
      scopeList.value = data?.records || (Array.isArray(res) ? res : [])
      console.log('作用域列表加载:', scopeList.value, '原始响应:', res)
      // 加载完成后初始化表单（确保作用域列表已加载）
      if (props.modelValue && props.roleData) {
        initForm()
      }
    } catch (error: any) {
      console.error('加载作用域列表失败:', error)
      ElMessage.error(error?.message || '加载作用域列表失败')
    } finally {
      scopeLoading.value = false
    }
  }

  /**
   * 监听弹窗打开，初始化表单数据
   */
  watch(
    () => props.modelValue,
    (newVal) => {
      if (newVal) {
        // 先加载作用域列表，然后在loadScopes中调用initForm
        loadScopes()
      }
    }
  )

  /**
   * 监听角色数据变化，更新表单
   */
  watch(
    () => props.roleData,
    (newData) => {
      if (newData && props.modelValue && scopeList.value.length > 0) {
        // 只有在作用域列表已加载时才初始化表单
        initForm()
      }
    },
    { deep: true }
  )

  /**
   * 初始化表单数据
   * 根据弹窗类型填充表单或重置表单
   */
  const initForm = () => {
    if (props.dialogType === 'edit' && props.roleData) {
      // 编辑模式：使用角色数据填充表单
      const roleData = props.roleData
      if (!roleData) return

      let scopeId = roleData.scopeId || ''
      // 如果没有scopeId，尝试从scopeCode查找
      if (!scopeId && roleData.scopeCode && scopeList.value.length > 0) {
        const found = scopeList.value.find((s) => s.scopeCode === roleData.scopeCode)
        scopeId = found ? found.scopeId : ''
      }
      console.log(
        '初始化表单 - scopeId:',
        scopeId,
        'roleData.scopeId:',
        roleData.scopeId,
        'roleData.scopeCode:',
        roleData.scopeCode
      )
      Object.assign(form, {
        roleId: roleData.roleId,
        roleName: roleData.roleName,
        roleCode: roleData.roleCode,
        description: roleData.description || '',
        createTime: roleData.createTime || '',
        scopeId: scopeId,
        sortOrder: roleData.sortOrder ?? 0,
        priority: roleData.priority || 0,
        status: roleData.status || 'normal'
      })
    } else {
      // 新增模式：重置表单，默认选择第一个作用域
      Object.assign(form, {
        roleId: '',
        roleName: '',
        roleCode: '',
        description: '',
        createTime: '',
        scopeId: scopeList.value.length > 0 ? scopeList.value[0].scopeId : '',
        sortOrder: 0,
        priority: 0,
        status: 'normal'
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
   * 验证通过后调用接口保存数据
   */
  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      if (props.dialogType === 'add') {
        if (!form.scopeId) {
          ElMessage.error('请选择作用域')
          return
        }
        await fetchCreateRole({
          code: form.roleCode,
          name: form.roleName,
          description: form.description || '',
          scope_id: form.scopeId,
          sort_order: form.sortOrder ?? 0,
          priority: form.priority || 0,
          status: form.status || 'normal'
        })
      } else {
        const roleId =
          typeof form.roleId === 'string' ? form.roleId : (form.roleId as any)?.toString?.() || ''
        if (!roleId) {
          ElMessage.error('缺少角色ID')
          return
        }
        // 编辑时，必须传递 scope_id（即使没有修改也要传递当前值）
        const scopeId = form.scopeId || props.roleData?.scopeId || ''
        if (!scopeId) {
          ElMessage.error('缺少作用域ID')
          return
        }
        const updateData: Api.SystemManage.RoleUpdateParams = {
          code: form.roleCode,
          name: form.roleName,
          description: form.description || '',
          scope_id: scopeId,
          sort_order: form.sortOrder ?? 0,
          priority: form.priority || 0,
          status: form.status || 'normal'
        }
        await fetchUpdateRole(roleId, updateData)
      }
      ElMessage.success(props.dialogType === 'add' ? '新增成功' : '修改成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      if (error?.message) ElMessage.error(error.message)
    }
  }
</script>
