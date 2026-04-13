<!-- 登录页面 -->
<template>
  <div
    class="flex w-full h-screen"
    :class="themeClass"
    :style="{
      '--auth-primary-color': theme.primaryColor || undefined,
      '--auth-border-radius': theme.borderRadius || undefined
    }"
  >
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h3 class="title">{{ texts.title || $t('login.title') }}</h3>
          <p class="sub-title">{{ texts.subTitle || $t('login.subTitle') }}</p>
          <ElTag v-if="loginPageKey" size="small" effect="plain" class="mt-3">
            模板：{{ templateName || loginPageKey }}
          </ElTag>
          <ElAlert
            v-if="submitError"
            :title="submitError"
            type="error"
            :closable="false"
            show-icon
            style="margin-top: 16px"
          />
          <ElForm
            ref="formRef"
            :model="formData"
            :rules="rules"
            :key="formKey"
            @keyup.enter="handleSubmit"
            style="margin-top: 25px"
          >
            <ElFormItem prop="username">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.username') || '用户名'"
                v-model.trim="formData.username"
                type="text"
              />
            </ElFormItem>
            <ElFormItem prop="password">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.password')"
                v-model.trim="formData.password"
                type="password"
                autocomplete="off"
                show-password
              />
            </ElFormItem>

            <div v-if="features.rememberMe !== false || features.forgetPassword !== false" class="flex-cb mt-2 text-sm">
              <ElCheckbox v-if="features.rememberMe !== false" v-model="formData.rememberPassword">{{
                $t('login.rememberPwd')
              }}</ElCheckbox>
              <RouterLink v-if="features.forgetPassword !== false" class="text-theme" to="/account/auth/forget-password">{{
                $t('login.forgetPwd')
              }}</RouterLink>
            </div>

            <div style="margin-top: 30px">
              <ElButton
                class="w-full custom-height"
                type="primary"
                @click="handleSubmit"
                :loading="loading"
                :disabled="isPreview"
                v-ripple
              >
                {{ texts.btnText || $t('login.btnText') }}
              </ElButton>
            </div>

            <div v-if="features.register !== false" class="mt-5 text-sm text-gray-600">
              <span>{{ $t('login.noAccount') }}</span>
              <RouterLink class="text-theme" :to="registerLink">{{
                $t('login.register')
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
  import { type FormInstance, type FormRules } from 'element-plus'
  import { useLoginFlow } from '@/domains/auth/flows/useLoginFlow'
  import { type LoginFormState } from '@/domains/auth/flows/shared'
  import { useAuthPageTemplate } from '@/domains/auth/useAuthPageTemplate'

  defineOptions({ name: 'Login' })

  const { t, locale } = useI18n()
  const formKey = ref(0)

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const formRef = ref<FormInstance>()
  const { loading, submitError, loadRememberedCredentials, submit } = useLoginFlow()
  const { themeClass, loginPageKey, templateName, registerLink, theme, features, texts, isPreview } =
    useAuthPageTemplate('login')

  // 登录表单默认值（不再预置系统账号密码）
  const formData = reactive<LoginFormState>({
    username: '',
    password: '',
    rememberPassword: false
  })

  const rules = computed<FormRules>(() => ({
    username: [{ required: true, message: t('login.placeholder.username'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.placeholder.password'), trigger: 'blur' }]
  }))

  // 登录
  const handleSubmit = async () => {
    if (isPreview.value) {
      console.info('[Login] preview mode — form submission blocked')
      return
    }
    if (!formRef.value) return

    try {
      const valid = await formRef.value.validate()
      if (!valid) return
      await submit(formData)
    } catch (error) {
      console.error('[Login] 表单校验失败:', error)
    }
  }

  onMounted(() => {
    loadRememberedCredentials(formData)
  })
</script>

<style scoped>
  @import './style.css';
</style>

<style lang="scss" scoped>
  :deep(.el-select__wrapper) {
    height: 40px !important;
  }
</style>
