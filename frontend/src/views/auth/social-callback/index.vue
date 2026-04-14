<template>
  <div class="social-callback-page">
    <div class="social-callback-card">
      <h3>正在处理社交登录</h3>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { fetchSocialTokenExchange } from '@/domains/auth/api'
  import { finalizeAuthenticatedSession, gotoAfterAuth } from '@/domains/auth/flows/shared'
  import { RoutesAlias } from '@/router/routesAlias'

  defineOptions({ name: 'SocialCallbackPage' })

  const router = useRouter()
  const route = useRoute()
  const message = ref('请稍候...')

  async function run() {
    const socialToken = `${route.query.social_token || ''}`.trim()
    if (!socialToken) {
      throw new Error('缺少 social_token')
    }
    const result = await fetchSocialTokenExchange(socialToken)
    if (result.intent === 'login' && result.access_token) {
      message.value = '已识别绑定账号，正在为你登录...'
      await finalizeAuthenticatedSession(
        {
          access_token: result.access_token,
          refresh_token: result.refresh_token
        },
        { refreshUserContext: false }
      )
      const fallbackPath = `${route.query.target_path || '/'}`.trim()
      const navigationSpaceKey = `${route.query.navigation_space_key || ''}`.trim()
      await gotoAfterAuth({ home_path: fallbackPath, navigation_space_key: navigationSpaceKey }, router)
      return
    }

    message.value = result.intent === 'conflict' ? '检测到同邮箱已有账号，正在跳转到注册/关联页...' : '正在跳转到注册页完成绑定...'
    const query: Record<string, string> = {
      social_token: socialToken
    }
    const loginPageKey = `${route.query.login_page_key || ''}`.trim()
    if (loginPageKey) query.login_page_key = loginPageKey
    const targetPath = `${route.query.target_path || ''}`.trim()
    if (targetPath) query.redirect = targetPath
    const sourceAppKey = `${route.query.source_app_key || ''}`.trim()
    if (sourceAppKey) query.source_app_key = sourceAppKey
    const sourceSpaceKey = `${route.query.source_navigation_space_key || ''}`.trim()
    if (sourceSpaceKey) query.source_navigation_space_key = sourceSpaceKey
    const sourceHomePath = `${route.query.source_home_path || ''}`.trim()
    if (sourceHomePath) query.source_home_path = sourceHomePath
    await router.replace({ path: '/account/auth/register', query })
  }

  onMounted(async () => {
    try {
      await run()
    } catch (error) {
      const text = error instanceof Error ? error.message : '社交登录失败，请重试'
      message.value = text
      ElMessage.error(text)
      setTimeout(() => {
        void router.replace(RoutesAlias.Login)
      }, 1200)
    }
  })
</script>

<style scoped>
  .social-callback-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, #f4f7fb 0%, #e7eef9 100%);
  }

  .social-callback-card {
    width: min(420px, calc(100vw - 32px));
    padding: 32px 28px;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.94);
    box-shadow: 0 18px 48px rgba(34, 56, 101, 0.12);
    text-align: center;
  }

  h3 {
    margin: 0 0 12px;
    font-size: 22px;
    color: #223865;
  }

  p {
    margin: 0;
    color: #5b6780;
    line-height: 1.7;
  }
</style>
