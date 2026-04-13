<template>
  <div class="p-4 register-policy-page art-full-height">
    <div class="page-top-stack">
      <div class="page-card">
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

      </div>
    </div>

    <ElCard class="art-table-card register-policy-main" shadow="never">
      <ArtTableHeader layout="refresh,fullscreen" :loading="loading" @refresh="load">
        <template #left>
          <div class="table-toolbar-tip">策略列表支持前端分页浏览，模板卡片用于快速套用预设，保存后会回刷完整策略目录。</div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="pagedList"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDrawer
      v-model="dialogVisible"
      :title="editing ? '编辑策略' : '新建策略'"
      size="44%"
      direction="rtl"
      class="policy-editor-drawer"
    >
      <div class="drawer-shell">
        <div class="dialog-layout">
          <ElForm :model="form" label-width="170px" class="dialog-form">
          <div class="form-section">
            <div class="section-title">基础信息</div>
            <div class="section-tip">先定义策略标识和名称，便于后续入口引用与运营识别。</div>
            <ElFormItem label="策略 Code" required>
              <ElInput v-model="form.policy_code" :disabled="!!editing" placeholder="如 default.self / invite.only" />
            </ElFormItem>
            <ElFormItem label="名称" required>
              <ElInput v-model="form.name" placeholder="给运营可读的模板名称" />
            </ElFormItem>
            <ElFormItem label="状态">
              <ElSelect v-model="form.status">
                <ElOption label="enabled" value="enabled" />
                <ElOption label="disabled" value="disabled" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="描述">
              <ElInput v-model="form.description" type="textarea" :rows="2" />
            </ElFormItem>
          </div>

          <div class="form-section">
            <div class="section-title">注册规则</div>
            <div class="section-tip">这里决定公开访问条件与注册校验强度。</div>
            <div class="switch-grid">
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
              <ElFormItem label="注册后自动登录">
                <ElSwitch v-model="form.auto_login" />
              </ElFormItem>
            </div>
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
          </div>

          <div class="form-section">
            <div class="section-title">注册后去向</div>
            <div class="section-tip">定义注册成功后进入的 App、空间与首页。</div>
            <ElFormItem label="目标 App Key" required>
              <ElInput v-model="form.target_app_key" placeholder="如 platform-admin" />
            </ElFormItem>
            <ElFormItem label="目标导航空间 Key" required>
              <ElInput v-model="form.target_navigation_space_key" placeholder="如 self-service" />
            </ElFormItem>
            <ElFormItem label="目标 Home Path">
              <ElInput v-model="form.target_home_path" placeholder="如 /self/user-center" />
            </ElFormItem>
          </div>

          <div class="form-section">
            <div class="section-title">绑定能力</div>
            <div class="section-tip">给新用户附加默认角色和功能包，避免注册后没有可用能力。</div>
            <ElFormItem label="绑定角色 Codes">
              <ElSelect v-model="form.role_codes" multiple filterable allow-create style="width: 100%">
              </ElSelect>
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
            </ElFormItem>
          </div>
          </ElForm>
        </div>
        <div class="drawer-footer">
          <div class="drawer-footer-tip">抽屉已改为更紧凑的半屏宽度，适合集中编辑策略配置。</div>
          <div class="drawer-footer-actions">
            <ElButton @click="dialogVisible = false">取消</ElButton>
            <ElButton type="primary" @click="submit">保存</ElButton>
          </div>
        </div>
      </div>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
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
  const loading = ref(false)
  const pagination = reactive({
    current: 1,
    size: 10,
    total: 0
  })

  const emptyForm = () => ({
    policy_code: '',
    name: '',
    description: '',
    target_app_key: 'platform-admin',
    target_navigation_space_key: 'self-service',
    target_home_path: '/self/user-center',
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

  const pagedList = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return list.value.slice(start, start + pagination.size)
  })
  const columns = computed<ColumnOption[]>(() => [
    { type: 'index', label: '序号', width: 70 },
    { prop: 'policy_code', label: '策略 Code', minWidth: 170, showOverflowTooltip: true },
    { prop: 'name', label: '名称', minWidth: 180, showOverflowTooltip: true },
    {
      prop: 'landing',
      label: '注册后去向',
      minWidth: 280,
      formatter: (row) =>
        h('div', {}, [
          h('div', { class: 'font-medium' }, buildLandingSummary(row)),
          h(
            'div',
            { class: 'text-xs text-gray-500' },
            row.auto_login ? '注册后自动登录' : '注册后回到登录页'
          )
        ])
    },
    {
      prop: 'requirements',
      label: '要求',
      minWidth: 240,
      formatter: (row) =>
        h(
          'div',
          { class: 'policy-tags' },
          buildRequirementTags(row).map((tag) => h(ElTag, { effect: 'plain' }, () => tag))
        )
    },
    { prop: 'status', label: '状态', width: 110 },
    {
      prop: 'actions',
      label: '操作',
      width: 150,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'table-actions' }, [
          h(
            ElButton,
            {
              link: true,
              type: 'primary',
              onClick: () => openEdit(row)
            },
            () => '编辑'
          ),
          h(
            ElButton,
            {
              link: true,
              type: 'danger',
              onClick: () => confirmRemove(row)
            },
            () => '删除'
          )
        ])
    }
  ])

  const load = async () => {
    loading.value = true
    try {
      const data: any = await fetchListRegisterPolicies()
      list.value = data?.records || []
      pagination.total = list.value.length
      syncCurrentPage()
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
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

  const confirmRemove = async (row: any) => {
    try {
      await ElMessageBox.confirm(`确认删除策略“${row.name || row.policy_code}”吗？`, '删除确认', {
        type: 'warning'
      })
      await remove(row)
    } catch {}
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

  function syncCurrentPage() {
    const totalPages = Math.max(1, Math.ceil((pagination.total || 0) / pagination.size))
    if (pagination.current > totalPages) {
      pagination.current = totalPages
    }
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    pagination.current = 1
    syncCurrentPage()
  }

  function handleCurrentChange(current: number) {
    pagination.current = current
    syncCurrentPage()
  }

  onMounted(load)
</script>

<style scoped>
  .page-card {
    padding: 18px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 20px;
    background: #fff;
    box-shadow:
      0 12px 30px rgb(15 23 42 / 5%),
      0 2px 8px rgb(15 23 42 / 4%);
  }

  .register-policy-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-height: 0;
  }

  .register-policy-main {
    flex: 1;
    min-height: 0;
  }

  .register-policy-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .page-hero {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }

  .hero-desc {
    margin-top: 4px;
    max-width: 820px;
    color: var(--el-text-color-secondary);
    line-height: 1.6;
  }

  .template-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
    margin-bottom: 12px;
  }

  .template-card {
    padding: 12px;
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
    margin-top: 6px;
    min-height: 36px;
    line-height: 1.5;
    color: var(--el-text-color-secondary);
  }

  .template-tags,
  .policy-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .table-toolbar-tip {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .table-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  :deep(.policy-editor-drawer) {
    --el-drawer-padding-primary: 20px;
  }

  :deep(.policy-editor-drawer .el-drawer__header) {
    margin-bottom: 0;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--el-border-color-light);
  }

  :deep(.policy-editor-drawer .el-drawer__body) {
    padding: 0;
    overflow: hidden;
  }

  .drawer-shell {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 73px);
    background: var(--el-bg-color);
  }

  .dialog-layout {
    flex: 1;
    display: flex;
    min-width: 0;
    min-height: 0;
    padding: 20px;
    overflow: hidden;
  }

  .dialog-form {
    flex: 1;
    width: 100%;
    min-height: 0;
    height: 100%;
    overflow-y: auto;
    padding-right: 8px;
  }

  .form-section {
    margin-bottom: 20px;
    padding: 16px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 14px;
    background: var(--el-fill-color-blank);
  }

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .section-tip {
    margin: 4px 0 12px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }

  .switch-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 16px;
  }

  .drawer-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 16px 20px;
    border-top: 1px solid var(--el-border-color-light);
    background: var(--el-fill-color-blank);
  }

  .drawer-footer-tip {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .drawer-footer-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  @media (max-width: 992px) {
    .drawer-shell {
      height: calc(100vh - 61px);
    }

    .dialog-layout {
      padding: 16px;
    }
  }

  @media (max-width: 768px) {
    .switch-grid {
      grid-template-columns: 1fr;
    }

    .drawer-footer {
      flex-direction: column;
      align-items: stretch;
    }

    .drawer-footer-actions {
      justify-content: flex-end;
    }
  }
</style>
