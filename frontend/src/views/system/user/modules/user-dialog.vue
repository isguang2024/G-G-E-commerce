<template>
  <ElDrawer
    v-model="dialogVisible"
    :title="dialogType === 'add' ? '添加用户' : '编辑用户'"
    size="620px"
    direction="rtl"
    class="config-drawer user-dialog-drawer"
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="80px">
      <div class="form-intro">
        <div class="form-intro__title">{{
          dialogType === 'add' ? '创建平台账号' : '更新账号信息'
        }}</div>
        <div class="form-intro__text">
          先确定账号基础信息，再配置状态和角色。角色会决定平台侧可见能力，协作空间侧生效请在权限测试里核对。
        </div>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div class="form-section__title">基础信息</div>
          <div class="form-section__desc">用户名用于登录，名称用于后台展示。</div>
        </div>
        <ElFormItem label="用户名" prop="username">
          <ElInput v-model="formData.username" placeholder="请输入用户名" />
        </ElFormItem>
        <ElFormItem label="名称" prop="nickname">
          <ElInput v-model="formData.nickname" placeholder="选填，用户展示名称" />
        </ElFormItem>
        <ElFormItem label="邮箱" prop="email">
          <ElInput v-model="formData.email" placeholder="选填" />
        </ElFormItem>
        <ElFormItem label="手机号" prop="phone">
          <ElInput v-model="formData.phone" placeholder="选填，11位手机号" />
        </ElFormItem>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div class="form-section__title">账号状态</div>
          <div class="form-section__desc">编辑时密码留空表示不修改，禁用后账号将不能继续登录。</div>
        </div>
        <ElFormItem label="密码" prop="password">
          <ElInput
            v-model="formData.password"
            type="password"
            :placeholder="dialogType === 'add' ? '请输入密码（至少6位）' : '留空表示不修改'"
            show-password
          />
        </ElFormItem>
        <ElFormItem label="状态" prop="status">
          <ElSelect v-model="formData.status" placeholder="请选择状态">
            <ElOption label="正常" value="active" />
            <ElOption label="禁用" value="inactive" />
          </ElSelect>
        </ElFormItem>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div class="form-section__title">角色与备注</div>
          <div class="form-section__desc">角色来自正式角色表，系统备注仅管理员可见。</div>
        </div>
        <ElFormItem label="角色" prop="roleIds">
          <ElSelect
            v-model="formData.roleIds"
            multiple
            placeholder="请选择角色（来自数据库）"
            :loading="roleLoading"
          >
            <ElOption
              v-for="role in roleList"
              :key="role.roleId"
              :value="role.roleId"
              :label="role.roleName"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="系统备注" prop="systemRemark">
          <ElInput
            v-model="formData.systemRemark"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="仅管理员可见"
          />
        </ElFormItem>
      </div>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">提交</ElButton>
      </div>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { fetchGetRoleListSimple } from '@/api/system-manage'
  import type { FormInstance, FormRules } from 'element-plus'

  interface Props {
    visible: boolean
    type: string
    userData?: Partial<Api.SystemManage.UserListItem>
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (
      e: 'submit',
      payload: Api.SystemManage.UserCreateParams | Api.SystemManage.UserUpdateParams
    ): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  // 角色列表（从后端获取，仅显示数据库中的角色）
  const roleList = ref<Api.SystemManage.RoleListItem[]>([])
  const roleLoading = ref(false)

  async function loadRoleList() {
    if (roleList.value.length > 0) return
    roleLoading.value = true
    try {
      const res = await fetchGetRoleListSimple()
      roleList.value = (res?.records || []) as Api.SystemManage.RoleListItem[]
    } finally {
      roleLoading.value = false
    }
  }

  // 对话框显示控制
  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const dialogType = computed(() => props.type)

  // 表单实例
  const formRef = ref<FormInstance>()

  // 表单数据（roleIds 为后端角色 ID 数组；status 与后端一致：active / inactive）
  const formData = reactive({
    username: '',
    email: '',
    nickname: '',
    password: '',
    status: 'active',
    phone: '',
    systemRemark: '',
    roleIds: [] as string[]
  })

  // 表单验证规则
  const rules = computed<FormRules>(() => ({
    username: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
      { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
    ],
    email: [{ type: 'email', message: '请输入正确邮箱格式', trigger: 'blur' }],
    password:
      props.type === 'add'
        ? [
            { required: true, message: '请输入密码', trigger: 'blur' },
            { min: 6, message: '密码至少 6 位', trigger: 'blur' }
          ]
        : [{ min: 6, message: '密码至少 6 位', trigger: 'blur' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }],
    phone: [{ pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号格式', trigger: 'blur' }],
    roleIds: [{ required: true, message: '请选择角色', trigger: 'change', type: 'array', min: 1 }]
  }))

  /**
   * 初始化表单数据
   * 编辑时：将后端返回的 userRoles（角色 code 数组）转为 roleIds（从 roleList 中按 code 匹配）
   */
  const initFormData = () => {
    const isEdit = props.type === 'edit' && props.userData
    const row = props.userData
    const roles = row?.userRoles
    const codes = Array.isArray(roles) ? roles : []

    const roleIds =
      isEdit && codes.length > 0 && roleList.value.length > 0
        ? roleList.value.filter((r) => codes.includes(r.roleCode)).map((r) => r.roleId)
        : []

    Object.assign(formData, {
      username: isEdit && row ? row.userName || '' : '',
      email: isEdit && row ? (row.userEmail ?? '') : '',
      nickname: isEdit && row ? row.nickName || '' : '',
      password: '',
      status: isEdit && row && row.status ? row.status : 'active',
      phone: isEdit && row ? row.userPhone || '' : '',
      systemRemark: isEdit && row ? row.systemRemark || '' : '',
      roleIds
    })
  }

  /**
   * 监听对话框打开：先拉取角色列表（仅数据库中的角色），再初始化表单
   */
  watch(
    () => props.visible,
    async (visible) => {
      if (visible) {
        await loadRoleList()
        initFormData()
        nextTick(() => formRef.value?.clearValidate())
      }
    }
  )

  watch(
    () => [props.type, props.userData],
    () => {
      if (props.visible && roleList.value.length > 0) initFormData()
    },
    { deep: true }
  )

  /**
   * 提交表单：将表单数据（含 roleIds）传给父组件，由父组件调用创建/更新接口
   */
  const handleSubmit = async () => {
    if (!formRef.value) return
    await formRef.value.validate((valid) => {
      if (valid) {
        emit('submit', {
          username: formData.username,
          email: formData.email,
          nickname: formData.nickname,
          password: formData.password,
          status: formData.status,
          phone: formData.phone,
          systemRemark: formData.systemRemark,
          roleIds: formData.roleIds
        })
      }
    })
  }
</script>

<style scoped lang="scss">
  .user-dialog-drawer :deep(.el-drawer__body) {
    padding-top: 8px;
  }

  .form-intro {
    padding: 14px 16px;
    margin-bottom: 16px;
    border: 1px solid rgb(226 232 240 / 0.95);
    border-radius: 16px;
    background: linear-gradient(135deg, rgb(248 250 252 / 0.98), rgb(241 245 249 / 0.95));
  }

  .form-intro__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .form-intro__text,
  .form-section__desc {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .form-section {
    padding: 14px 16px 4px;
    margin-bottom: 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 16px;
    background: rgb(255 255 255 / 0.96);
  }

  .form-section__header {
    margin-bottom: 12px;
  }

  .form-section__title {
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
  }
</style>

