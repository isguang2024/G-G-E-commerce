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
  import { RoutesAlias } from '@/router/routesAlias'
  import { useCallbackFlow } from '@/domains/auth/flows/useCallbackFlow'

  defineOptions({ name: 'AuthCallbackPage' })

  const router = useRouter()
  const { message, run } = useCallbackFlow()

  onMounted(async () => {
    try {
      await run()
    } catch (error) {
      console.error('[AuthCallback] 回调处理失败:', error)
      message.value = error instanceof Error ? error.message : '登录回调失败，请重试'
      ElMessage.error(message.value)
      setTimeout(() => {
        void router.replace(RoutesAlias.Login)
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
