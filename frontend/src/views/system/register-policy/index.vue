<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-lg font-semibold">注册策略</h3>
      <ElButton type="primary" @click="openCreate">新建策略</ElButton>
    </div>
    <ElTable :data="list" border stripe>
      <ElTableColumn prop="policy_code" label="策略 Code" width="160" />
      <ElTableColumn prop="name" label="名称" width="180" />
      <ElTableColumn prop="app_key" label="所属 App" width="140" />
      <ElTableColumn prop="target_app_key" label="目标 App" width="140" />
      <ElTableColumn prop="target_navigation_space_key" label="目标空间" width="140" />
      <ElTableColumn prop="target_home_path" label="目标 Home" width="180" />
      <ElTableColumn label="公开注册" width="100">
        <template #default="{ row }">
          <ElTag :type="row.allow_public_register ? 'success' : 'info'">
            {{ row.allow_public_register ? '开启' : '关闭' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="status" label="状态" width="100" />
      <ElTableColumn label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <ElButton link type="primary" @click="openEdit(row)">编辑</ElButton>
          <ElPopconfirm title="确认删除该策略？" @confirm="remove(row)">
            <template #reference>
              <ElButton link type="danger">删除</ElButton>
            </template>
          </ElPopconfirm>
        </template>
      </ElTableColumn>
    </ElTable>

    <ElDialog v-model="dialogVisible" :title="editing ? '编辑策略' : '新建策略'" width="680px">
      <ElForm :model="form" label-width="160px">
        <ElFormItem label="策略 Code" required>
          <ElInput v-model="form.policy_code" :disabled="!!editing" />
        </ElFormItem>
        <ElFormItem label="名称" required>
          <ElInput v-model="form.name" />
        </ElFormItem>
        <ElFormItem label="所属 App Key" required>
          <ElInput v-model="form.app_key" placeholder="如 account-portal" />
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="form.description" type="textarea" :rows="2" />
        </ElFormItem>
        <ElFormItem label="目标 App Key" required>
          <ElInput v-model="form.target_app_key" placeholder="如 platform-admin" />
        </ElFormItem>
        <ElFormItem label="目标空间 Key" required>
          <ElInput v-model="form.target_navigation_space_key" placeholder="如 self-service" />
        </ElFormItem>
        <ElFormItem label="目标 Home Path">
          <ElInput v-model="form.target_home_path" placeholder="如 /self/user-center" />
        </ElFormItem>
        <ElFormItem label="默认 Workspace 类型">
          <ElInput v-model="form.default_workspace_type" placeholder="personal" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="form.status">
            <ElOption label="enabled" value="enabled" />
            <ElOption label="disabled" value="disabled" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="允许公开注册">
          <ElSwitch v-model="form.allow_public_register" />
        </ElFormItem>
        <ElFormItem label="需要邀请码">
          <ElSwitch v-model="form.require_invite" />
        </ElFormItem>
        <ElFormItem label="需要邮箱验证">
          <ElSwitch v-model="form.require_email_verify" />
        </ElFormItem>
        <ElFormItem label="需要人机验证">
          <ElSwitch v-model="form.require_captcha" />
        </ElFormItem>
        <template v-if="form.require_captcha">
          <ElFormItem label="验证提供商">
            <ElSelect v-model="form.captcha_provider" style="width:200px">
              <ElOption label="无（文本输入降级）" value="none" />
              <ElOption label="reCAPTCHA v3" value="recaptcha" />
              <ElOption label="hCaptcha" value="hcaptcha" />
              <ElOption label="Turnstile" value="turnstile" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="Site Key" v-if="form.captcha_provider && form.captcha_provider !== 'none'">
            <ElInput v-model="form.captcha_site_key" placeholder="公开 site key（前端渲染 widget 用）" />
          </ElFormItem>
        </template>
        <ElFormItem label="注册后自动登录">
          <ElSwitch v-model="form.auto_login" />
        </ElFormItem>
        <ElFormItem label="绑定角色 (codes)">
          <ElSelect v-model="form.role_codes" multiple filterable allow-create style="width:100%">
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="绑定功能包 (keys)">
          <ElSelect v-model="form.feature_package_keys" multiple filterable allow-create style="width:100%">
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="submit">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    fetchCreateRegisterPolicy,
    fetchDeleteRegisterPolicy,
    fetchListRegisterPolicies,
    fetchUpdateRegisterPolicy
  } from '@/api/system-manage/register'

  defineOptions({ name: 'SystemRegisterPolicy' })

  const list = ref<any[]>([])
  const dialogVisible = ref(false)
  const editing = ref<any>(null)

  const emptyForm = () => ({
    policy_code: '',
    name: '',
    app_key: 'account-portal',
    description: '',
    target_app_key: 'platform-admin',
    target_navigation_space_key: 'self-service',
    target_home_path: '/self/user-center',
    default_workspace_type: 'personal',
    status: 'enabled',
    allow_public_register: false,
    require_invite: false,
    require_email_verify: false,
    require_captcha: false,
    auto_login: true,
    captcha_provider: 'none',
    captcha_site_key: '',
    role_codes: [] as string[],
    feature_package_keys: [] as string[]
  })
  const form = reactive<any>(emptyForm())

  const load = async () => {
    try {
      const data: any = await fetchListRegisterPolicies()
      list.value = data?.records || []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    }
  }

  const openCreate = () => {
    editing.value = null
    Object.assign(form, emptyForm())
    dialogVisible.value = true
  }

  const openEdit = (row: any) => {
    editing.value = row
    Object.assign(form, emptyForm(), row)
    dialogVisible.value = true
  }

  const submit = async () => {
    try {
      if (editing.value) {
        await fetchUpdateRegisterPolicy(editing.value.policy_code, form)
        ElMessage.success('更新成功')
      } else {
        await fetchCreateRegisterPolicy(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await load()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  }

  const remove = async (row: any) => {
    try {
      await fetchDeleteRegisterPolicy(row.policy_code)
      ElMessage.success('已删除')
      await load()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }

  onMounted(load)
</script>
