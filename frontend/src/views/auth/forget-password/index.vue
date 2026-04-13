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
          <h3 class="title">{{ texts.title || $t('forgetPassword.title') }}</h3>
          <p class="sub-title">{{ texts.subTitle || $t('forgetPassword.subTitle') }}</p>
          <ElTag v-if="loginPageKey" size="small" effect="plain" class="mt-3">
            模板：{{ templateName || loginPageKey }}
          </ElTag>
          <div class="mt-5">
            <span class="input-label" v-if="showInputLabel">账号</span>
            <ElInput
              class="custom-height"
              :placeholder="$t('forgetPassword.placeholder')"
              v-model.trim="username"
            />
          </div>

          <div style="margin-top: 15px">
            <ElButton
                class="w-full custom-height"
                type="primary"
                @click="register"
                :loading="loading"
                :disabled="isPreview"
                v-ripple
              >
                {{ texts.btnText || $t('forgetPassword.submitBtnText') }}
              </ElButton>
            </div>

          <div style="margin-top: 15px">
            <ElButton class="w-full custom-height" plain @click="toLogin">
              {{ $t('forgetPassword.backBtnText') }}
            </ElButton>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { RoutesAlias } from '@/router/routesAlias'
  import { useAuthPageTemplate } from '@/domains/auth/useAuthPageTemplate'

  defineOptions({ name: 'ForgetPassword' })

  const router = useRouter()
  const showInputLabel = ref(false)

  const username = ref('')
  const loading = ref(false)
  const { themeClass, loginPageKey, templateName, theme, texts, isPreview } =
    useAuthPageTemplate('forget_password')

  const register = async () => {
    if (isPreview.value) return
  }

  const toLogin = () => {
    const key = `${loginPageKey.value || ''}`.trim()
    if (!key || key === 'default') {
      router.push({ path: RoutesAlias.Login })
      return
    }
    router.push({ path: RoutesAlias.Login, query: { login_page_key: key } })
  }
</script>

<style scoped>
  @import '../login/style.css';
</style>
