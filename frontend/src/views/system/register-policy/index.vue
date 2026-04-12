<template>
  <div class="p-4 register-policy-page">
    <div class="page-hero">
      <div>
        <h3 class="text-lg font-semibold">注册策略</h3>
        <p class="hero-desc">
          策略就是“注册模板”。它定义公开注册是否开启、是否需要邀请码/邮箱验证/验证码、注册成功后进入哪个 App/空间/首页，以及默认绑定哪些角色和功能包。
        </p>
      </div>
      <ElButton type="primary" @click="openCreate">新建策略</ElButton>
    </div>

    <div class="template-grid">
      <button
        v-for="template in templates"
        :key="template.code"
        type="button"
        class="template-card"
        @click="openTemplate(template.code)"
      >
        <div class="template-title">{{ template.name }}</div>
        <div class="template-desc">{{ template.description }}</div>
        <div class="template-tags">
          <ElTag v-for="tag in template.tags" :key="tag" effect="plain">{{ tag }}</ElTag>
        </div>
      </button>
    </div>

    <ElAlert
      class="mb-4"
      type="info"
      :closable="false"
      title="推荐顺序"
      description="先选一个策略模板，再检查目标 App / 空间 / 首页，最后让注册入口引用这个策略。默认示例：default.self -> platform-admin / self-service / /self/user-center。"
    />

    <ElTable :data="list" border stripe>
      <ElTableColumn prop="policy_code" label="策略 Code" width="170" />
      <ElTableColumn prop="name" label="名称" width="180" />
      <ElTableColumn prop="app_key" label="所属 App" width="140" />
      <ElTableColumn label="注册后去向" min-width="260">
        <template #default="{ row }">
          <div class="font-medium">{{ buildLandingSummary(row) }}</div>
          <div class="text-xs text-gray-500">
            {{ row.auto_login ? '注册后自动登录' : '注册后回到登录页' }}
          </div>
        </template>
      </ElTableColumn>
      <ElTableColumn label="要求" min-width="220">
        <template #default="{ row }">
          <div class="policy-tags">
            <ElTag v-for="tag in buildRequirementTags(row)" :key="tag" effect="plain">{{ tag }}</ElTag>
          </div>
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

    <ElDialog v-model="dialogVisible" :title="editing ? '编辑策略' : '新建策略'" width="980px">
      <div class="dialog-layout">
        <ElForm :model="form" label-width="170px" class="dialog-form">
          <ElFormItem label="策略 Code" required>
            <ElInput v-model="form.policy_code" :disabled="!!editing" placeholder="如 default.self / invite.only" />
          </ElFormItem>
          <ElFormItem label="名称" required>
            <ElInput v-model="form.name" placeholder="给运营可读的模板名称" />
          </ElFormItem>
          <ElFormItem label="所属 App Key" required>
            <ElInput v-model="form.app_key" placeholder="如 account-portal" />
            <div class="field-tip">谁承载注册页，就由谁持有策略。公开认证页建议统一放到 account-portal。</div>
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
            <div class="field-tip">注册成功后的 landing。auto_login 开启时会直接跳这里。</div>
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
            <div class="field-tip">关闭后，即使命中入口，公开页也只会提示“当前未开启公开注册”。</div>
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
              <ElSelect v-model="form.captcha_provider" style="width: 220px">
                <ElOption label="无（文本输入降级）" value="none" />
                <ElOption label="reCAPTCHA v3" value="recaptcha" />
                <ElOption label="hCaptcha" value="hcaptcha" />
                <ElOption label="Turnstile" value="turnstile" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem
              label="Site Key"
              v-if="form.captcha_provider && form.captcha_provider !== 'none'"
            >
              <ElInput v-model="form.captcha_site_key" placeholder="公开 site key（前端渲染 widget 用）" />
            </ElFormItem>
          </template>
          <ElFormItem label="注册后自动登录">
            <ElSwitch v-model="form.auto_login" />
          </ElFormItem>
          <ElFormItem label="绑定角色 Codes">
            <ElSelect v-model="form.role_codes" multiple filterable allow-create style="width: 100%">
            </ElSelect>
            <div class="field-tip">示例：personal.self_user。这里决定新用户创建后额外挂哪些角色。</div>
          </ElFormItem>
          <ElFormItem label="绑定功能包 Keys">
            <ElSelect
              v-model="form.feature_package_keys"
              multiple
              filterable
              allow-create
              style="width: 100%"
            >
            </ElSelect>
            <div class="field-tip">示例：self_service.basic。建议和角色一起维护，避免注册后没有可用能力。</div>
          </ElFormItem>
        </ElForm>

        <div class="dialog-preview">
          <div class="preview-card">
            <div class="preview-title">模板预览</div>
            <ElDescriptions :column="1" border size="small">
              <ElDescriptionsItem label="模板类型">{{ activeTemplateName }}</ElDescriptionsItem>
              <ElDescriptionsItem label="公开注册">{{ form.allow_public_register ? '开启' : '关闭' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="字段要求">{{ buildRequirementText(form) }}</ElDescriptionsItem>
              <ElDescriptionsItem label="落地目标">{{ buildLandingSummary(form) }}</ElDescriptionsItem>
              <ElDescriptionsItem label="角色 / 功能包">
                {{ buildBindingsSummary(form) }}
              </ElDescriptionsItem>
            </ElDescriptions>
          </div>

          <ElAlert
            class="mt-4"
            type="success"
            :closable="false"
            title="保存后怎么验"
            :description="`1. 让注册入口引用当前策略；2. 打开 /account/auth/register；3. 检查页面顶部是否显示入口与策略，并验证本页要求的字段是否出现。`"
          />
        </div>
      </div>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="submit">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    fetchCreateRegisterPolicy,
    fetchDeleteRegisterPolicy,
    fetchListRegisterPolicies,
    fetchUpdateRegisterPolicy
  } from '@/domains/governance/api/register'

  defineOptions({ name: 'SystemRegisterPolicy' })

  const templates = [
    {
      code: 'public-default',
      name: '默认公开注册',
      description: '适合默认自助注册。开放公开注册，成功后进入 self-service 空间。',
      tags: ['公开注册', '自动登录', '自助空间']
    },
    {
      code: 'invite-only',
      name: '邀请码注册',
      description: '适合活动或私域邀请。公开页存在，但必须携带邀请码才能通过。',
      tags: ['邀请码', '可公开访问', '可复用默认 landing']
    },
    {
      code: 'email-verify',
      name: '邮箱验证注册',
      description: '适合对账户真实性要求更高的场景，可叠加验证码。',
      tags: ['邮箱验证', '可叠加验证码', '更严格']
    }
  ] as const

  const list = ref<any[]>([])
  const dialogVisible = ref(false)
  const editing = ref<any>(null)
  const currentTemplateCode = ref('public-default')

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
    role_codes: ['personal.self_user'] as string[],
    feature_package_keys: ['self_service.basic'] as string[]
  })
  const form = reactive<any>(emptyForm())

  const activeTemplateName = computed(() => {
    return templates.find((item) => item.code === currentTemplateCode.value)?.name || '自定义模板'
  })

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
    currentTemplateCode.value = 'public-default'
    Object.assign(form, emptyForm())
    dialogVisible.value = true
  }

  const openEdit = (row: any) => {
    editing.value = row
    currentTemplateCode.value = detectTemplateCode(row)
    Object.assign(form, emptyForm(), row)
    dialogVisible.value = true
  }

  const openTemplate = (code: string) => {
    editing.value = null
    currentTemplateCode.value = code
    Object.assign(form, emptyForm(), buildTemplateForm(code))
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

  function buildTemplateForm(code: string) {
    const base = emptyForm()
    if (code === 'invite-only') {
      return {
        ...base,
        policy_code: 'invite.only',
        name: '邀请码注册模板',
        description: '公开页可访问，但必须输入邀请码后才能完成注册。',
        allow_public_register: true,
        require_invite: true,
        auto_login: false
      }
    }
    if (code === 'email-verify') {
      return {
        ...base,
        policy_code: 'email.verify',
        name: '邮箱验证注册模板',
        description: '要求填写邮箱，可按需叠加验证码能力。',
        allow_public_register: true,
        require_email_verify: true,
        require_captcha: true,
        captcha_provider: 'none'
      }
    }
    return {
      ...base,
      policy_code: 'default.self',
      name: '默认自注册策略',
      description: '公开注册默认策略：注册成功后进入 platform-admin/self-service 空间',
      allow_public_register: true
    }
  }

  function detectTemplateCode(row: any) {
    if (row?.require_invite) return 'invite-only'
    if (row?.require_email_verify) return 'email-verify'
    return 'public-default'
  }

  function buildLandingSummary(row: {
    target_app_key?: string
    target_navigation_space_key?: string
    target_home_path?: string
  }) {
    const appKey = `${row.target_app_key || '-'}`.trim()
    const spaceKey = `${row.target_navigation_space_key || '-'}`.trim()
    const homePath = `${row.target_home_path || '-'}`.trim()
    return `${appKey} / ${spaceKey} / ${homePath}`
  }

  function buildRequirementTags(row: {
    allow_public_register?: boolean
    require_invite?: boolean
    require_email_verify?: boolean
    require_captcha?: boolean
    auto_login?: boolean
  }) {
    const tags = [row.allow_public_register ? '公开注册' : '关闭公开注册']
    if (row.require_invite) tags.push('邀请码')
    if (row.require_email_verify) tags.push('邮箱验证')
    if (row.require_captcha) tags.push('人机验证')
    tags.push(row.auto_login ? '自动登录' : '注册后手动登录')
    return tags
  }

  function buildRequirementText(row: {
    allow_public_register?: boolean
    require_invite?: boolean
    require_email_verify?: boolean
    require_captcha?: boolean
  }) {
    return buildRequirementTags(row).join(' / ')
  }

  function buildBindingsSummary(row: { role_codes?: string[]; feature_package_keys?: string[] }) {
    const roles = Array.isArray(row.role_codes) && row.role_codes.length ? row.role_codes.join(', ') : '无角色'
    const packages =
      Array.isArray(row.feature_package_keys) && row.feature_package_keys.length
        ? row.feature_package_keys.join(', ')
        : '无功能包'
    return `${roles} / ${packages}`
  }

  onMounted(load)
</script>

<style scoped>
  .page-hero {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 16px;
  }

  .hero-desc {
    margin-top: 6px;
    max-width: 820px;
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .template-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
    margin-bottom: 16px;
  }

  .template-card {
    padding: 16px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 16px;
    background: linear-gradient(180deg, rgb(248 250 252 / 95%), #fff);
    text-align: left;
    cursor: pointer;
    transition:
      transform 0.2s ease,
      border-color 0.2s ease;
  }

  .template-card:hover {
    transform: translateY(-2px);
    border-color: var(--el-color-primary-light-5);
  }

  .template-title {
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .template-desc {
    margin-top: 8px;
    min-height: 44px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }

  .template-tags,
  .policy-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .dialog-layout {
    display: grid;
    grid-template-columns: minmax(0, 1.45fr) minmax(300px, 0.9fr);
    gap: 20px;
  }

  .preview-card {
    padding: 16px;
    border-radius: 16px;
    border: 1px solid var(--el-border-color-light);
    background: linear-gradient(180deg, rgb(248 250 252 / 95%), #fff);
  }

  .preview-title {
    margin-bottom: 12px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .field-tip {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }
</style>
