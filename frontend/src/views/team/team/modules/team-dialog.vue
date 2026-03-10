<template>
  <ElDialog
    v-model="dialogVisible"
    :title="dialogType === 'add' ? '新增团队' : '编辑团队'"
    width="500px"
    align-center
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="100px">
      <ElFormItem label="团队名称" prop="name">
        <ElInput v-model="formData.name" placeholder="请输入团队名称" maxlength="200" show-word-limit />
      </ElFormItem>
      <ElFormItem label="团队备注" prop="remark">
        <ElInput
          v-model="formData.remark"
          placeholder="请输入团队备注"
          maxlength="500"
        />
      </ElFormItem>
      <ElFormItem label="管理员配置" prop="admin_user_ids">
        <ElSelect
          v-model="formData.admin_user_ids"
          multiple
          filterable
          allow-create
          default-first-option
          placeholder="请输入用户ID并回车，可配置多个"
          style="width: 100%"
        >
        </ElSelect>
        <div class="text-gray-400 text-xs mt-1">输入用户ID后回车确认，添加后将被设为团队管理员</div>
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
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    teamData?: Partial<Api.SystemManage.TeamListItem>
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', payload?: Api.SystemManage.TeamCreateParams | Api.SystemManage.TeamUpdateParams): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const formRef = ref<FormInstance>()
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
    name: [{ required: true, message: '请输入团队名称', trigger: 'blur' }]
  }))

  watch(
    () => [props.visible, props.teamData],
    () => {
      if (!props.visible) return
      if (props.type === 'edit' && props.teamData) {
        formData.name = props.teamData.name ?? ''
        formData.remark = props.teamData.remark ?? ''
        formData.admin_user_ids = []
        formData.logo_url = props.teamData.logoUrl ?? ''
        formData.plan = props.teamData.plan ?? 'free'
        formData.max_members = props.teamData.maxMembers ?? 10
        formData.status = props.teamData.status ?? 'active'
      } else {
        formData.name = ''
        formData.remark = ''
        formData.admin_user_ids = []
        formData.logo_url = ''
        formData.plan = 'free'
        formData.max_members = 10
        formData.status = 'active'
      }
    },
    { immediate: true }
  )

  async function handleSubmit() {
    await formRef.value?.validate()
    if (props.type === 'add') {
      emit('submit', {
        name: formData.name,
        remark: formData.remark,
        admin_user_ids: formData.admin_user_ids,
        logo_url: formData.logo_url || undefined,
        plan: formData.plan,
        max_members: formData.max_members,
        status: formData.status
      })
    } else {
      emit('submit', {
        name: formData.name,
        remark: formData.remark,
        admin_user_ids: formData.admin_user_ids,
        logo_url: formData.logo_url || undefined,
        plan: formData.plan,
        max_members: formData.max_members,
        status: formData.status
      })
    }
  }
</script>
