<!-- 注册页面 -->
<template>
  <div class="flex w-full h-screen">
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h3 class="title">{{ $t('register.title') }}</h3>
          <p class="sub-title">{{ $t('register.subTitle') }}</p>
          <ElAlert
            v-if="contextError"
            class="mt-4"
            type="warning"
            :closable="false"
            title="注册上下文读取失败"
            :description="contextError"
          />
          <div v-else-if="ctx" class="mt-4 register-context-panel">
            <div class="context-header">
              <div>
                <div class="context-title">当前入口配置</div>
                <div class="context-desc">
                  命中入口 {{ ctx.entry_name || ctx.entry_code }}，以下字段和去向都来自当前注册策略。
                </div>
              </div>
              <ElTag :type="ctx.allow_public_register ? 'success' : 'warning'">
                {{ ctx.allow_public_register ? '公开注册已开启' : '公开注册未开启' }}
              </ElTag>
            </div>
            <div class="context-grid">
              <div class="context-item">
                <span class="context-label">入口 Code</span>
                <span class="context-value">{{ ctx.entry_code }}</span>
              </div>
              <div class="context-item">
                <span class="context-label">策略 Code</span>
                <span class="context-value">{{ ctx.policy_code }}</span>
              </div>
              <div class="context-item">
                <span class="context-label">注册来源</span>
                <span class="context-value">{{ registerSourceLabel }}</span>
              </div>
              <div class="context-item">
                <span class="context-label">注册后去向</span>
                <span class="context-value">{{ landingSummary }}</span>
              </div>
            </div>
            <div class="context-section">
              <span class="context-label">本页会要求你填写</span>
              <div class="context-tags">
                <ElTag v-for="item in requiredFieldTags" :key="item" effect="plain">{{ item }}</ElTag>
              </div>
            </div>
            <ElAlert
              v-if="!ctx.allow_public_register"
              class="mt-3"
              type="warning"
              :closable="false"
              title="当前入口暂未开放公开注册"
              description="请让管理员在“注册策略”或当前“注册入口”中开启公开注册，然后再回到此页面验证。"
            />
            <ElAlert
              v-else
              class="mt-3"
              type="info"
              :closable="false"
              :title="verificationTitle"
              :description="verificationDescription"
            />
          </div>
          <ElForm
            class="mt-7.5"
            ref="formRef"
            :model="formData"
            :rules="rules"
            label-position="top"
            :key="formKey"
          >
            <ElFormItem prop="username">
              <ElInput
                class="custom-height"
                v-model.trim="formData.username"
                :placeholder="$t('register.placeholder.username')"
              />
            </ElFormItem>

            <ElFormItem prop="password">
              <ElInput
                class="custom-height"
                v-model.trim="formData.password"
                :placeholder="$t('register.placeholder.password')"
                type="password"
                autocomplete="off"
                show-password
              />
            </ElFormItem>

            <ElFormItem prop="confirmPassword">
              <ElInput
                class="custom-height"
                v-model.trim="formData.confirmPassword"
                :placeholder="$t('register.placeholder.confirmPassword')"
                type="password"
                autocomplete="off"
                show-password
              />
            </ElFormItem>

            <!-- 邮箱（策略要求邮箱验证时显示） -->
            <ElFormItem v-if="ctx?.require_email_verify" prop="email">
              <ElInput
                class="custom-height"
                v-model.trim="formData.email"
                placeholder="邮箱地址"
                type="email"
                autocomplete="email"
              />
            </ElFormItem>

            <!-- 邀请码（策略要求邀请码时显示） -->
            <ElFormItem v-if="ctx?.require_invite" prop="invitationCode">
              <ElInput
                class="custom-height"
                v-model.trim="formData.invitationCode"
                placeholder="邀请码"
                autocomplete="off"
              />
            </ElFormItem>

            <!-- 人机验证（无第三方 widget 时降级为文本输入） -->
            <ElFormItem v-if="ctx?.require_captcha" prop="captchaToken">
              <template v-if="!ctx?.captcha_provider || ctx.captcha_provider === 'none'">
                <ElInput
                  class="custom-height"
                  v-model.trim="formData.captchaToken"
                  placeholder="验证码（请联系管理员获取）"
                  autocomplete="off"
                />
              </template>
              <template v-else>
                <!-- 占位：集成真实 captcha widget 时替换此处 -->
                <div class="text-sm text-gray-500">
                  人机验证（{{ ctx.captcha_provider }}）暂不可用，请联系管理员
                </div>
              </template>
            </ElFormItem>

            <ElFormItem prop="agreement">
              <ElCheckbox v-model="formData.agreement">
                {{ $t('register.agreeText') }}
                <RouterLink
                  style="color: var(--theme-color); text-decoration: none"
                  to="/privacy-policy"
                  >{{ $t('register.privacyPolicy') }}</RouterLink
                >
              </ElCheckbox>
            </ElFormItem>

            <div style="margin-top: 15px">
              <ElButton
                class="w-full custom-height"
                type="primary"
                @click="register"
                :loading="loading"
                :disabled="isPublicRegisterDisabled"
                v-ripple
              >
                {{ $t('register.submitBtnText') }}
              </ElButton>
            </div>

            <div class="mt-5 text-sm text-g-600">
              <span>{{ $t('register.hasAccount') }}</span>
              <RouterLink class="text-theme" :to="RoutesAlias.Login">{{
                $t('register.toLogin')
              }}</RouterLink>
            </div>
          </ElForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { fetchRegister, fetchRegisterContext } from '@/api/auth'
  import { RoutesAlias } from '@/router/routesAlias'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'Register' })

  const userStore = useUserStore()

  const ctx = ref<Awaited<ReturnType<typeof fetchRegisterContext>> | null>(null)
  const contextError = ref('')

  onMounted(async () => {
    try {
      ctx.value = await fetchRegisterContext(window.location.host, window.location.pathname)
    } catch (e) {
      contextError.value = '未读取到当前 URL 对应的注册入口，请先检查注册入口、策略和 account-portal 页面是否已完成配置。'
      console.warn('fetch register context failed', e)
    }
  })

  interface RegisterForm {
    username: string
    password: string
    confirmPassword: string
    email: string
    invitationCode: string
    captchaToken: string
    agreement: boolean
  }

  const USERNAME_MIN_LENGTH = 3
  const USERNAME_MAX_LENGTH = 20
  const PASSWORD_MIN_LENGTH = 6
  const REDIRECT_DELAY = 1000

  const { t, locale } = useI18n()
  const router = useRouter()
  const formRef = ref<FormInstance>()

  const loading = ref(false)
  const formKey = ref(0)
  const isPublicRegisterDisabled = computed(() => ctx.value?.allow_public_register === false)
  const registerSourceLabel = computed(() => {
    const source = `${ctx.value?.register_source || 'self'}`.trim()
    if (source === 'invite') return '邀请码注册'
    if (source === 'self') return '默认公开注册'
    return source
  })
  const landingSummary = computed(() => {
    if (!ctx.value) return '待加载'
    const targetApp = ctx.value.target_app_key || '-'
    const targetSpace = ctx.value.target_navigation_space_key || '-'
    const targetPath = ctx.value.target_home_path || '-'
    return `${targetApp} / ${targetSpace} / ${targetPath}`
  })
  const requiredFieldTags = computed(() => {
    const tags = ['用户名', '密码', '确认密码']
    if (ctx.value?.require_email_verify) tags.push('邮箱')
    if (ctx.value?.require_invite) tags.push('邀请码')
    if (ctx.value?.require_captcha) {
      tags.push(ctx.value.captcha_provider && ctx.value.captcha_provider !== 'none' ? '人机验证' : '验证码')
    }
    return tags
  })
  const verificationTitle = computed(() => {
    return ctx.value?.auto_login ? '提交后将自动登录并跳到目标首页' : '提交后需要回到登录页继续登录'
  })
  const verificationDescription = computed(() => {
    if (!ctx.value) return ''
    const checks = [
      `访问路径：${window.location.pathname}`,
      `命中入口：${ctx.value.entry_code}`,
      `目标首页：${ctx.value.target_home_path}`
    ]
    if (ctx.value.require_email_verify) checks.push('需准备可接收邮件的邮箱')
    if (ctx.value.require_invite) checks.push('需准备有效邀请码')
    if (ctx.value.require_captcha) checks.push('需准备验证码或第三方人机验证')
    return checks.join('；')
  })

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const formData = reactive<RegisterForm>({
    username: '',
    password: '',
    confirmPassword: '',
    email: '',
    invitationCode: '',
    captchaToken: '',
    agreement: false
  })

  /**
   * 验证密码
   * 当密码输入后，如果确认密码已填写，则触发确认密码的验证
   */
  const validatePassword = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (!value) {
      callback(new Error(t('register.placeholder.password')))
      return
    }

    if (formData.confirmPassword) {
      formRef.value?.validateField('confirmPassword')
    }

    callback()
  }

  /**
   * 验证确认密码
   * 检查确认密码是否与密码一致
   */
  const validateConfirmPassword = (
    _rule: any,
    value: string,
    callback: (error?: Error) => void
  ) => {
    if (!value) {
      callback(new Error(t('register.rule.confirmPasswordRequired')))
      return
    }

    if (value !== formData.password) {
      callback(new Error(t('register.rule.passwordMismatch')))
      return
    }

    callback()
  }

  /**
   * 验证用户协议
   * 确保用户已勾选同意协议
   */
  const validateAgreement = (_rule: any, value: boolean, callback: (error?: Error) => void) => {
    if (!value) {
      callback(new Error(t('register.rule.agreementRequired')))
      return
    }
    callback()
  }

  const rules = computed<FormRules<RegisterForm>>(() => ({
    username: [
      { required: true, message: t('register.placeholder.username'), trigger: 'blur' },
      {
        min: USERNAME_MIN_LENGTH,
        max: USERNAME_MAX_LENGTH,
        message: t('register.rule.usernameLength'),
        trigger: 'blur'
      }
    ],
    password: [
      { required: true, validator: validatePassword, trigger: 'blur' },
      { min: PASSWORD_MIN_LENGTH, message: t('register.rule.passwordLength'), trigger: 'blur' }
    ],
    confirmPassword: [{ required: true, validator: validateConfirmPassword, trigger: 'blur' }],
    ...(ctx.value?.require_email_verify
      ? { email: [{ required: true, type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }] }
      : {}),
    ...(ctx.value?.require_invite
      ? { invitationCode: [{ required: true, message: '请输入邀请码', trigger: 'blur' }] }
      : {}),
    ...(ctx.value?.require_captcha &&
    (!ctx.value.captcha_provider || ctx.value.captcha_provider === 'none')
      ? { captchaToken: [{ required: true, message: '请输入验证码', trigger: 'blur' }] }
      : {}),
    agreement: [{ validator: validateAgreement, trigger: 'change' }]
  }))

  /**
   * 注册用户
   * 验证表单后提交注册请求
   */
  const register = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      loading.value = true

      if (isPublicRegisterDisabled.value) {
        ElMessage.error('当前未开启公开注册')
        loading.value = false
        return
      }
      const res = await fetchRegister({
        username: formData.username,
        password: formData.password,
        confirm_password: formData.confirmPassword,
        ...(formData.email ? { email: formData.email } : {}),
        ...(formData.invitationCode ? { invitation_code: formData.invitationCode } : {}),
        ...(formData.captchaToken ? { captcha_token: formData.captchaToken } : {}),
        agreement_version: 'v1'
      })
      loading.value = false
      ElMessage.success('注册成功')

      // auto_login: 持久化 token，直接跳转 landing
      if (res.access_token) {
        userStore.setToken(res.access_token, res.refresh_token ?? undefined)
        userStore.setLoginStatus(true)
        const homePath = res.landing?.home_path ?? '/dashboard/console'
        setTimeout(() => router.push(homePath), REDIRECT_DELAY)
      } else if (res.pending) {
        // auto_login=false：注册成功但未自动登录，引导去登录页
        setTimeout(
          () => router.push({ path: RoutesAlias.Login, query: { registered: '1' } }),
          REDIRECT_DELAY
        )
      } else if (res.landing?.home_path) {
        setTimeout(() => router.push(res.landing!.home_path!), REDIRECT_DELAY)
      } else {
        toLogin()
      }
    } catch (error) {
      console.error('表单验证失败:', error)
      loading.value = false
    }
  }

  /**
   * 跳转到登录页面
   */
  const toLogin = () => {
    setTimeout(() => {
      router.push({ path: RoutesAlias.Login })
    }, REDIRECT_DELAY)
  }
</script>

<style scoped>
  @import '../login/style.css';

  .register-context-panel {
    border: 1px solid var(--el-border-color-light);
    border-radius: 16px;
    padding: 16px;
    background: linear-gradient(180deg, rgb(248 250 252 / 96%), #fff);
  }

  .context-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .context-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .context-desc {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }

  .context-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
    margin-top: 14px;
  }

  .context-item {
    padding: 12px;
    border-radius: 12px;
    background: rgb(255 255 255 / 92%);
    border: 1px solid var(--el-border-color-lighter);
  }

  .context-label {
    display: block;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .context-value {
    display: block;
    margin-top: 4px;
    font-size: 13px;
    line-height: 1.5;
    color: var(--el-text-color-primary);
    word-break: break-all;
  }

  .context-section {
    margin-top: 14px;
  }

  .context-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 8px;
  }
</style>
