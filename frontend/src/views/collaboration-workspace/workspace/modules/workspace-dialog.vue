<template>
  <ElDrawer
    v-model="dialogVisible"
    :title="type === 'add' ? '新增协作空间' : '编辑协作空间'"
    size="500px"
    direction="rtl"
    class="config-drawer"
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="100px">
      <ElFormItem label="协作空间名称" prop="name">
        <ElInput
          v-model="formData.name"
          placeholder="请输入协作空间名称"
          maxlength="200"
          show-word-limit
        />
      </ElFormItem>
      <ElFormItem label="协作空间备注" prop="remark">
        <ElInput v-model="formData.remark" placeholder="请输入协作空间备注" maxlength="500" />
      </ElFormItem>
      <ElFormItem label="管理员配置" prop="admin_user_ids">
        <div class="admin-tags-container">
          <ElTag
            v-for="admin in adminList"
            :key="admin.id"
            closable
            @close="removeAdmin(admin.id)"
            class="admin-tag"
          >
            {{ admin.name }} ({{ admin.id }})
          </ElTag>
        </div>
        <ElInput
          v-if="inputVisible"
          ref="inputRef"
          v-model="inputValue"
          class="input-new-tag"
          placeholder="输入用户ID后回车"
          @keyup.enter="handleInputConfirm"
          @blur="handleInputConfirm"
        />
        <ElButton v-else class="button-new-tag" @click="showInput"> + 添加管理员 </ElButton>
        <div class="text-gray-400 text-xs mt-1"
          >输入用户ID后回车确认，添加后将被设为协作空间管理员</div
        >
      </ElFormItem>
      <ElFormItem label="Logo URL" prop="logo_url">
        <ElInput v-model="formData.logo_url" placeholder="选填" clearable />
      </ElFormItem>
      <ElFormItem label="套餐" prop="plan">
        <ElSelect v-model="formData.plan" placeholder="选填" clearable style="width: 100%">
          <ElOption label="免费版" value="free" />
          <ElOption label="专业版" value="pro" />
          <ElOption label="企业版" value="enterprise" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="最大成员数" prop="max_members">
        <ElInputNumber v-model="formData.max_members" :min="1" :max="10000" style="width: 100%" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSelect v-model="formData.status" placeholder="请选择状态" style="width: 100%">
          <ElOption label="正常" value="active" />
          <ElOption label="停用" value="inactive" />
        </ElSelect>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确定</ElButton>
      </div>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { fetchGetUser } from '@/api/system-manage'

  interface AdminUser {
    id: string
    name: string
  }

  interface AdminCandidate {
    id?: string
    user_id?: string
    name?: string
    nickName?: string
    userName?: string
    userEmail?: string
    email?: string
  }

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    collaborationWorkspaceData?: Partial<Api.SystemManage.CollaborationWorkspaceListItem>
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (
      e: 'submit',
      payload?:
        | Api.SystemManage.CollaborationWorkspaceCreateParams
        | Api.SystemManage.CollaborationWorkspaceUpdateParams
    ): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const formRef = ref<FormInstance>()
  const inputRef = ref<HTMLInputElement>()
  const inputVisible = ref(false)
  const inputValue = ref('')
  const adminList = ref<AdminUser[]>([])

  const resolveAdminUsers = (
    payload?: Partial<Api.SystemManage.CollaborationWorkspaceListItem>
  ): AdminCandidate[] => {
    if (!payload) return []
    const normalized = payload as Partial<Api.SystemManage.CollaborationWorkspaceListItem> & {
      adminUsers?: AdminCandidate[]
      admin_users?: AdminCandidate[]
    }
    return normalized.adminUsers || normalized.admin_users || []
  }

  const formData = reactive({
    name: '',
    remark: '',
    admin_user_ids: [] as string[],
    logo_url: '',
    plan: 'free',
    max_members: 10,
    status: 'active'
  })

  const rules = computed<FormRules>(() => ({
    name: [{ required: true, message: '请输入协作空间名称', trigger: 'blur' }]
  }))

  // 加载管理员信息
  const loadAdminUsers = async (ids: string[]) => {
    adminList.value = []
    for (const id of ids) {
      try {
        const user = await fetchGetUser(id)
        const name = user.nickName || user.userName || user.userEmail || '未知用户'
        adminList.value.push({ id, name })
      } catch {
        adminList.value.push({ id, name: '用户不存在' })
      }
    }
  }

  const showInput = () => {
    inputVisible.value = true
    nextTick(() => {
      inputRef.value?.focus()
    })
  }

  const removeAdmin = (id: string) => {
    adminList.value = adminList.value.filter((a) => a.id !== id)
    formData.admin_user_ids = adminList.value.map((a) => a.id)
  }

  const handleInputConfirm = async () => {
    const value = inputValue.value.trim()
    if (!value) {
      inputVisible.value = false
      inputValue.value = ''
      return
    }

    // 检查重复
    if (adminList.value.some((a) => a.id === value)) {
      ElMessage.warning('该用户ID已添加')
      inputVisible.value = false
      inputValue.value = ''
      return
    }

    // 验证用户是否存在
    try {
      const user = await fetchGetUser(value)
      const name = user.nickName || user.userName || user.userEmail || '未知用户'
      adminList.value.push({ id: value, name })
      formData.admin_user_ids = adminList.value.map((a) => a.id)
      ElMessage.success(`已成功添加用户 [${name}] 为协作空间管理员`)
    } catch {
      ElMessage.error('用户不存在，请检查用户ID')
      return
    }

    inputVisible.value = false
    inputValue.value = ''
  }

  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      const collaborationWorkspaceData = props.collaborationWorkspaceData
      if (props.type === 'edit' && collaborationWorkspaceData) {
        formData.name = collaborationWorkspaceData.name ?? ''
        formData.remark = collaborationWorkspaceData.remark ?? ''
        const adminUsers = resolveAdminUsers(collaborationWorkspaceData)
        formData.admin_user_ids = adminUsers.map((admin) => admin.user_id || admin.id || '')
        formData.logo_url = collaborationWorkspaceData.logoUrl ?? ''
        formData.plan = collaborationWorkspaceData.plan ?? 'free'
        formData.max_members = collaborationWorkspaceData.maxMembers ?? 10
        formData.status = collaborationWorkspaceData.status ?? 'active'

        if (formData.admin_user_ids.length > 0) {
          await loadAdminUsers(formData.admin_user_ids)
        } else {
          adminList.value = []
        }
        return
      }

      formData.name = ''
      formData.remark = ''
      formData.admin_user_ids = []
      adminList.value = []
      formData.logo_url = ''
      formData.plan = 'free'
      formData.max_members = 10
      formData.status = 'active'
    },
    { immediate: true }
  )

  async function handleSubmit() {
    await formRef.value?.validate()
    // 过滤掉不存在的用户
    const validAdminIds = adminList.value.filter((a) => a.name !== '用户不存在').map((a) => a.id)

    if (props.type === 'add') {
      emit('submit', {
        name: formData.name,
        remark: formData.remark,
        admin_user_ids: validAdminIds,
        logo_url: formData.logo_url || undefined,
        plan: formData.plan,
        max_members: formData.max_members,
        status: formData.status
      })
    } else {
      emit('submit', {
        name: formData.name,
        remark: formData.remark,
        admin_user_ids: validAdminIds,
        logo_url: formData.logo_url || undefined,
        plan: formData.plan,
        max_members: formData.max_members,
        status: formData.status
      })
    }
  }
</script>

<style scoped>
  .admin-tags-container {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 8px;
  }
  .admin-tag {
    max-width: 300px;
  }
  .input-new-tag {
    width: 120px;
  }
  .button-new-tag {
    height: 32px;
    line-height: 30px;
    padding-top: 0;
    padding-bottom: 0;
  }
</style>
