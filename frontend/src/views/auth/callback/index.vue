<template>
  <div class="auth-callback-page">
    <div class="auth-callback-card">
      <h3>正在完成登录</h3>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { fetchExchangeAuthCallback } from '@/api/auth'
  import { useUserStore } from '@/store/modules/user'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { refreshCurrentUserInfoContext, refreshUserMenus } from '@/router/guards/beforeEach'
  import {
    consumeCentralizedAuthAttempt,
    resolveCentralizedTargetPath
  } from '@/utils/auth/centralized-login'

  defineOptions({ name: 'AuthCallbackPage' })

  const router = useRouter()
  const route = useRoute()
  const userStore = useUserStore()
  const menuSpaceStore = useMenuSpaceStore()
  const message = ref('正在校验回调参数并交换登录令牌...')

  async function handleCallback() {
    const code = `${route.query.code || ''}`.trim()
    const state = `${route.query.state || ''}`.trim()
    const targetAppKey = `${route.query.target_app_key || ''}`.trim()
    const redirectUri = `${route.query.redirect_uri || window.location.href}`.trim()
    if (!code || !state || !targetAppKey || !redirectUri) {
      throw new Error('缺少 callback 参数')
    }

    const attempt = consumeCentralizedAuthAttempt(state)
    if (!attempt?.nonce) {
      throw new Error('登录回调上下文已失效，请重新登录')
    }

    const result = await fetchExchangeAuthCallback({
      code,
      state,
      nonce: attempt.nonce,
      target_app_key: targetAppKey,
      redirect_uri: redirectUri
    })
    if (!result.access_token) {
      throw new Error('未收到访问令牌')
    }

    userStore.applySession({
      accessToken: result.access_token,
      refreshToken: result.refresh_token,
      isLogin: true
    })
    await refreshCurrentUserInfoContext()
    await refreshUserMenus()

    const landingPath = resolveCentralizedTargetPath(
      result.landing?.home_path,
      attempt.targetPath || `${route.query.target_path || ''}`.trim()
    )
    const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
      landingPath,
      result.landing?.navigation_space_key || attempt.navigationSpaceKey
    )
    if (nextTarget.mode === 'location') {
      window.location.assign(nextTarget.target)
      return
    }
    await router.replace(nextTarget.target)
  }

  onMounted(async () => {
    try {
      await handleCallback()
    } catch (error) {
      console.error('[AuthCallback] 回调处理失败:', error)
      message.value = error instanceof Error ? error.message : '登录回调失败，请重试'
      ElMessage.error(message.value)
      setTimeout(() => {
        router.replace('/account/auth/login')
      }, 1200)
    }
  })
</script>

<style scoped>
  .auth-callback-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, #f4f7fb 0%, #e7eef9 100%);
  }

  .auth-callback-card {
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
