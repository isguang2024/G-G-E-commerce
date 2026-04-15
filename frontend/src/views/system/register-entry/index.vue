<template>
  <div class="p-4 register-entry-page art-full-height">
    <ElCard class="art-table-card register-entry-main" shadow="never">
      <ArtTableHeader layout="refresh,fullscreen" :loading="loading" @refresh="load">
        <template #left>
          <div class="register-entry-header">
            <div class="register-entry-title">注册入口</div>
            <div class="register-entry-tip">
              每个入口自带完整的注册决策（角色/功能包/开关/跳转目标）。运营无需单独管理策略，在入口内直接配置即可。
            </div>
          </div>
        </template>
        <template #right>
          <ElSelect
            v-model="filterAppKey"
            clearable
            placeholder="按 App 筛选"
            style="width: 200px; margin-right: 12px"
          >
            <ElOption v-for="key in appKeyOptions" :key="key" :label="key" :value="key" />
          </ElSelect>
          <ElDropdown trigger="click" @command="handleCreateCommand">
            <ElButton type="primary">
              新建入口 <ElIcon class="el-icon--right"><ArrowDown /></ElIcon>
            </ElButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="blank">空白入口</ElDropdownItem>
                <ElDropdownItem divided disabled>从模板创建</ElDropdownItem>
                <ElDropdownItem command="tpl:public-default">默认公开注册</ElDropdownItem>
                <ElDropdownItem command="tpl:invite-only">邀请码注册</ElDropdownItem>
                <ElDropdownItem command="tpl:email-verify">邮箱验证注册</ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
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

    <!-- ── 抽屉编辑器 ── -->
    <ElDrawer
      v-model="drawerVisible"
      :title="drawerTitle"
      size="50%"
      direction="rtl"
      destroy-on-close
      class="entry-editor-drawer"
    >
      <div class="drawer-shell">
        <div class="drawer-body">
          <ElAlert
            v-if="isSystemReserved"
            class="mb-4"
            type="warning"
            :closable="false"
            title="系统保留入口"
            description="此入口为系统保留，不可删除，不可修改 entry_code。"
          />

          <ElForm ref="formRef" :model="form" :rules="formRules" label-width="150px" class="entry-form">
            <!-- 基础信息 -->
            <div class="form-section">
              <div class="section-title">基础信息</div>
              <div class="section-tip">定义入口标识和匹配规则。当用户访问的 Host + Path 命中此规则时，使用该入口的注册配置。</div>
              <ElFormItem
                label="入口 Code"
                prop="entry_code"
                :error="fieldErrors.entry_code"
                :data-testid="'register-entry-field-error'"
                :data-field="'entry_code'"
                required
              >
                <ElInput v-model="form.entry_code" :disabled="isSystemReserved || !!editing" placeholder="如 default / invite-only" />
              </ElFormItem>
              <ElFormItem
                label="名称"
                prop="name"
                :error="fieldErrors.name"
                :data-testid="'register-entry-field-error'"
                :data-field="'name'"
                required
              >
                <ElInput v-model="form.name" placeholder="给运营可读的入口名称" />
              </ElFormItem>
              <ElFormItem
                label="App Key"
                prop="app_key"
                :error="fieldErrors.app_key"
                :data-testid="'register-entry-field-error'"
                :data-field="'app_key'"
                required
              >
                <ElInput v-model="form.app_key" placeholder="如 account-portal" />
              </ElFormItem>
              <div class="field-row">
                <ElFormItem label="Host" class="field-half">
                  <ElInput v-model="form.host" placeholder="留空匹配任意 host" />
                </ElFormItem>
                <ElFormItem label="Path 前缀" class="field-half">
                  <ElInput v-model="form.path_prefix" placeholder="/account/auth/register" />
                </ElFormItem>
              </div>
              <ElFormItem label="注册来源标识">
                <ElInput v-model="form.register_source" placeholder="self / invite / ..." />
              </ElFormItem>
              <ElFormItem label="登录页模板">
                <ElSelect v-model="form.login_page_key" filterable allow-create default-first-option style="width: 100%">
                  <ElOption
                    v-for="item in templateList"
                    :key="item.template_key"
                    :label="`${item.template_key} · ${item.name}`"
                    :value="item.template_key"
                  />
                </ElSelect>
              </ElFormItem>
            </div>

            <!-- 注册规则 -->
            <div class="form-section">
              <div class="section-title">注册规则</div>
              <div class="section-tip">控制注册开关和验证强度。关闭公开注册后，此入口页面仍可展示但无法提交注册。</div>
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
              </div>
              <ElFormItem label="自动登录">
                <ElSwitch v-model="form.auto_login" />
                <div class="field-tip">开启后注册成功直接返回 token，关闭则返回 pending 提示用户手动登录。</div>
              </ElFormItem>
            </div>

            <!-- 验证码 -->
            <div v-if="form.require_captcha" class="form-section">
              <div class="section-title">验证码配置</div>
              <div class="section-tip">选择人机验证服务商和对应的公钥。前端会在注册表单中嵌入对应的验证组件。</div>
              <ElFormItem label="验证码提供商">
                <ElSelect v-model="form.captcha_provider" style="width: 100%">
                  <ElOption value="none" label="none" />
                  <ElOption value="recaptcha" label="reCAPTCHA v3" />
                  <ElOption value="hcaptcha" label="hCaptcha" />
                  <ElOption value="turnstile" label="Turnstile" />
                </ElSelect>
              </ElFormItem>
              <ElFormItem v-if="form.captcha_provider !== 'none'" label="Site Key">
                <ElInput v-model="form.captcha_site_key" placeholder="验证码公钥" />
              </ElFormItem>
            </div>

            <!-- 注册后去向 -->
            <div class="form-section">
              <div class="section-title">注册后去向</div>
              <div class="section-tip">注册成功后前端跳转目标。优先级：Target URL > Target App Key + Home Path > 来源回源 > 前端默认。</div>
              <ElFormItem label="Target URL">
                <ElInput v-model="form.target_url" placeholder="https://app.example.com/welcome" />
                <div class="field-tip">填写后将忽略下方 App Key / Home Path 配置，直接外跳此 URL。仅允许 http(s) 或相对路径。</div>
              </ElFormItem>
              <ElFormItem label="Target App Key">
                <ElInput v-model="form.target_app_key" placeholder="注册成功后跳转到的目标 App" />
              </ElFormItem>
              <div class="field-row">
                <ElFormItem label="导航空间 Key" class="field-half">
                  <ElInput v-model="form.target_navigation_space_key" placeholder="目标 App 的菜单空间" />
                </ElFormItem>
                <ElFormItem label="Home Path" class="field-half">
                  <ElInput v-model="form.target_home_path" placeholder="如 /dashboard" />
                </ElFormItem>
              </div>
            </div>

            <!-- 注册决策 -->
            <div class="form-section">
              <div class="section-title">注册决策</div>
              <div class="section-tip">注册时自动绑定的角色和功能包。新用户创建后立即拥有这些能力，无需管理员手动分配。</div>
              <ElFormItem label="绑定角色 Codes">
                <ElSelect v-model="form.role_codes" multiple filterable allow-create default-first-option placeholder="输入角色 code 回车添加" style="width: 100%">
                </ElSelect>
              </ElFormItem>
              <ElFormItem label="绑定功能包 Keys">
                <ElSelect v-model="form.feature_package_keys" multiple filterable allow-create default-first-option placeholder="输入功能包 key 回车添加" style="width: 100%">
                </ElSelect>
              </ElFormItem>
            </div>

            <!-- 其他 -->
            <div class="form-section">
              <div class="section-title">其他</div>
              <div class="section-tip">辅助字段，用于审计溯源和运维管理。</div>
              <ElFormItem label="说明">
                <ElInput v-model="form.description" type="textarea" :rows="2" placeholder="入口用途说明" />
              </ElFormItem>
              <div class="field-row">
                <ElFormItem label="状态" class="field-half">
                  <ElSelect v-model="form.status" style="width: 100%">
                    <ElOption label="enabled" value="enabled" />
                    <ElOption label="disabled" value="disabled" />
                  </ElSelect>
                </ElFormItem>
                <ElFormItem label="排序" class="field-half">
                  <ElInputNumber v-model="form.sort_order" :min="0" />
                </ElFormItem>
              </div>
              <ElFormItem label="备注">
                <ElInput v-model="form.remark" type="textarea" :rows="2" />
              </ElFormItem>
            </div>
          </ElForm>
        </div>

        <div class="drawer-footer">
          <ElButton @click="drawerVisible = false">取消</ElButton>
          <ElButton type="primary" @click="submit">保存</ElButton>
        </div>
      </div>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref, watch } from 'vue'
  import { ArrowDown } from '@element-plus/icons-vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElButton, ElIcon, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import { HttpError } from '@/utils/http/error'
  import {
    fetchCreateRegisterEntry,
    fetchDeleteRegisterEntry,
    fetchListLoginPageTemplates,
    fetchListRegisterEntries,
    fetchUpdateRegisterEntry
  } from '@/domains/governance/api/register'

  defineOptions({ name: 'SystemRegisterEntry' })

  interface EntryForm {
    entry_code: string
    name: string
    app_key: string
    host: string
    path_prefix: string
    register_source: string
    login_page_key: string
    status: string
    sort_order: number
    allow_public_register: boolean
    require_invite: boolean
    require_email_verify: boolean
    require_captcha: boolean
    auto_login: boolean
    is_system_reserved: boolean
    target_url: string
    target_app_key: string
    target_navigation_space_key: string
    target_home_path: string
    captcha_provider: string
    captcha_site_key: string
    description: string
    role_codes: string[]
    feature_package_keys: string[]
    remark: string
  }

  const DEFAULT_FORM: EntryForm = {
    entry_code: '',
    name: '',
    app_key: 'account-portal',
    host: '',
    path_prefix: '/account/auth/register',
    register_source: 'self',
    login_page_key: 'default',
    status: 'enabled',
    sort_order: 0,
    allow_public_register: true,
    require_invite: false,
    require_email_verify: false,
    require_captcha: false,
    auto_login: true,
    is_system_reserved: false,
    target_url: '',
    target_app_key: '',
    target_navigation_space_key: '',
    target_home_path: '',
    captcha_provider: 'none',
    captcha_site_key: '',
    description: '',
    role_codes: [],
    feature_package_keys: [],
    remark: ''
  }

  const TEMPLATES: Record<string, Partial<EntryForm>> = {
    'public-default': {
      name: '默认公开注册入口',
      description: '适合默认自助注册。开放公开注册，成功后自动登录。',
      allow_public_register: true,
      auto_login: true,
      role_codes: ['personal.self_user'],
      feature_package_keys: ['self_service.basic'],
      target_app_key: 'platform-admin',
      target_navigation_space_key: 'self-service',
      target_home_path: '/self/user-center'
    },
    'invite-only': {
      name: '邀请码注册入口',
      description: '适合活动或私域邀请。必须携带邀请码才能通过注册。',
      allow_public_register: true,
      require_invite: true,
      auto_login: false,
      role_codes: ['personal.self_user'],
      feature_package_keys: ['self_service.basic']
    },
    'email-verify': {
      name: '邮箱验证注册入口',
      description: '适合对账户真实性要求更高的场景，可叠加人机验证。',
      allow_public_register: true,
      require_email_verify: true,
      require_captcha: true,
      captcha_provider: 'none',
      auto_login: true,
      role_codes: ['personal.self_user'],
      feature_package_keys: ['self_service.basic']
    }
  }

  const list = ref<any[]>([])
  const templateList = ref<any[]>([])
  const drawerVisible = ref(false)
  const editing = ref<any>(null)
  const loading = ref(false)
  const pagination = reactive({ current: 1, size: 10, total: 0 })

  const form = reactive<EntryForm>({ ...DEFAULT_FORM })
  const formRef = ref<FormInstance>()
  // fieldErrors: 后端 Error.details.<field> 回显容器；规范见 docs/guides/frontend-observability-spec.md §2.4
  const fieldErrors = reactive<Record<string, string>>({})
  const filterAppKey = ref('')

  const formRules: FormRules = {
    entry_code: [
      { required: true, message: '请输入入口 Code', trigger: 'blur' },
      { pattern: /^[a-z0-9][a-z0-9-]*$/, message: '仅允许小写字母数字和短横线', trigger: 'blur' }
    ],
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    app_key: [{ required: true, message: '请输入 App Key', trigger: 'blur' }]
  }

  function clearFieldErrors() {
    for (const k of Object.keys(fieldErrors)) delete fieldErrors[k]
  }

  function applyBackendFieldErrors(e: unknown): boolean {
    if (!(e instanceof HttpError)) return false
    const data = (e.data || {}) as { details?: Record<string, string> }
    const details = data.details
    if (!details || typeof details !== 'object') return false
    let applied = false
    for (const [field, reason] of Object.entries(details)) {
      if (typeof reason === 'string') {
        fieldErrors[field] = reason
        applied = true
      }
    }
    return applied
  }

  const isSystemReserved = computed(() => editing.value?.is_system_reserved === true)
  const drawerTitle = computed(() => {
    if (!editing.value) return '新建入口'
    return isSystemReserved.value ? '编辑入口（系统保留）' : '编辑入口'
  })
  const appKeyOptions = computed(() => {
    const keys = new Set(list.value.map((r: any) => r.app_key).filter(Boolean))
    return Array.from(keys).sort()
  })
  const filteredList = computed(() => {
    if (!filterAppKey.value) return list.value
    return list.value.filter((r: any) => r.app_key === filterAppKey.value)
  })
  const pagedList = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return filteredList.value.slice(start, start + pagination.size)
  })

  const columns = computed<ColumnOption[]>(() => [
    { type: 'index', label: '序号', width: 70 },
    {
      prop: 'entry_code',
      label: '入口 Code',
      minWidth: 160,
      showOverflowTooltip: true,
      formatter: (row) =>
        h(
          'span',
          {
            'data-testid': 'register-entry-row',
            'data-code': row.entry_code,
            'data-app-key': row.app_key
          },
          [
            row.entry_code,
            row.is_system_reserved
              ? h(ElTag, { type: 'danger', size: 'small', effect: 'plain', class: 'ml-1' }, () => '保留')
              : null
          ]
        )
    },
    { prop: 'name', label: '名称', minWidth: 180, showOverflowTooltip: true },
    { prop: 'app_key', label: 'App', width: 140 },
    {
      prop: 'match_rule',
      label: '命中规则',
      minWidth: 260,
      formatter: (row) =>
        h('div', {}, [
          h('div', { class: 'font-medium' }, buildMatchRule(row)),
          h('div', { class: 'text-xs text-gray-500' }, buildVerifyUrl(row))
        ])
    },
    {
      prop: 'allow_public_register',
      label: '公开注册',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: row.allow_public_register ? 'success' : 'warning', effect: 'plain' }, () =>
          row.allow_public_register ? '开启' : '关闭'
        )
    },
    {
      prop: 'auto_login',
      label: '自动登录',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: row.auto_login ? 'success' : 'info', effect: 'plain' }, () =>
          row.auto_login ? '是' : '否'
        )
    },
    { prop: 'status', label: '状态', width: 90 },
    {
      prop: 'actions',
      label: '操作',
      width: 150,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'table-actions' }, [
          h(ElButton, { link: true, type: 'primary', onClick: () => openEdit(row) }, () => '编辑'),
          row.is_system_reserved
            ? null
            : h(ElButton, { link: true, type: 'danger', onClick: () => confirmRemove(row) }, () => '删除')
        ])
    }
  ])

  const load = async () => {
    loading.value = true
    try {
      const data: any = await fetchListRegisterEntries()
      list.value = data?.records || []
      pagination.total = filteredList.value.length
      syncCurrentPage()
      const templates: any = await fetchListLoginPageTemplates()
      templateList.value = templates?.records || []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  function handleCreateCommand(command: string) {
    editing.value = null
    if (command === 'blank') {
      Object.assign(form, { ...DEFAULT_FORM, role_codes: [], feature_package_keys: [] })
    } else if (command.startsWith('tpl:')) {
      const tplKey = command.slice(4)
      const tpl = TEMPLATES[tplKey]
      Object.assign(form, {
        ...DEFAULT_FORM,
        ...(tpl || {}),
        role_codes: tpl?.role_codes ? [...tpl.role_codes] : [],
        feature_package_keys: tpl?.feature_package_keys ? [...tpl.feature_package_keys] : []
      })
    }
    drawerVisible.value = true
  }

  const openEdit = (row: any) => {
    editing.value = row
    Object.assign(form, {
      ...DEFAULT_FORM,
      ...row,
      role_codes: Array.isArray(row.role_codes) ? [...row.role_codes] : [],
      feature_package_keys: Array.isArray(row.feature_package_keys) ? [...row.feature_package_keys] : []
    })
    drawerVisible.value = true
  }

  const submit = async () => {
    clearFieldErrors()
    const valid = await formRef.value?.validate().catch(() => false)
    if (!valid) return
    try {
      const payload: any = { ...form }
      payload.login_page_key = `${payload.login_page_key || ''}`.trim() || 'default'
      if (editing.value) {
        await fetchUpdateRegisterEntry(editing.value.id, payload)
        ElMessage.success('更新成功')
      } else {
        await fetchCreateRegisterEntry(payload)
        ElMessage.success('创建成功')
      }
      drawerVisible.value = false
      await load()
    } catch (e: any) {
      // 后端 FieldError(details.field) 优先回显到具体 el-form-item；
      // 否则退化为通用 toast（与 utils/http/error.ts 语义一致）。
      if (applyBackendFieldErrors(e)) return
      ElMessage.error(e?.message || '保存失败')
    }
  }

  const confirmRemove = async (row: any) => {
    try {
      await ElMessageBox.confirm(`确认删除入口"${row.name || row.entry_code}"吗？`, '删除确认', {
        type: 'warning'
      })
      await fetchDeleteRegisterEntry(row.id)
      ElMessage.success('已删除')
      await load()
    } catch {}
  }

  function buildMatchRule(row: { host?: string; path_prefix?: string }) {
    const host = `${row.host || ''}`.trim() || '任意 Host'
    const pathPrefix = `${row.path_prefix || ''}`.trim() || '任意路径'
    return `${host} + ${pathPrefix}`
  }

  function buildVerifyUrl(row: { host?: string; path_prefix?: string }) {
    const host = `${row.host || ''}`.trim() || 'localhost'
    const pathPrefix = `${row.path_prefix || ''}`.trim() || '/account/auth/register'
    return `https://${host}${pathPrefix}`
  }

  function syncCurrentPage() {
    const totalPages = Math.max(1, Math.ceil((pagination.total || 0) / pagination.size))
    if (pagination.current > totalPages) pagination.current = totalPages
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

  watch(filterAppKey, () => {
    pagination.current = 1
    pagination.total = filteredList.value.length
  })

  onMounted(load)
</script>

<style scoped>
  .register-entry-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .register-entry-main {
    flex: 1;
    min-height: 0;
  }

  .register-entry-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .register-entry-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .register-entry-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .register-entry-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .table-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  /* ── Drawer ── */
  :deep(.entry-editor-drawer) {
    --el-drawer-padding-primary: 0;
  }

  :deep(.entry-editor-drawer .el-drawer__header) {
    margin-bottom: 0;
    padding: 16px 20px;
    border-bottom: 1px solid var(--el-border-color-light);
  }

  :deep(.entry-editor-drawer .el-drawer__body) {
    padding: 0;
    overflow: hidden;
  }

  .drawer-shell {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .drawer-body {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
  }

  .entry-form {
    max-width: 640px;
  }

  .drawer-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
    padding: 14px 20px;
    border-top: 1px solid var(--el-border-color-light);
    background: var(--el-fill-color-blank);
  }

  /* ── Form sections ── */
  .form-section {
    margin-bottom: 20px;
    padding: 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-fill-color-blank);
  }

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .section-tip {
    margin: 4px 0 14px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }

  .switch-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 16px;
  }

  .field-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0 16px;
  }

  .field-half {
    min-width: 0;
  }

  .field-tip {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }
</style>
