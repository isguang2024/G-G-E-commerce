<template>
  <div v-if="!hasError" class="art-app-error-boundary__content">
    <slot />
  </div>
  <div v-else class="art-app-error-boundary">
    <ElResult
      icon="error"
      :title="isPublicApp ? '页面加载失败' : '应用运行异常'"
      :sub-title="errorHint"
    >
      <template #extra>
        <div class="art-app-error-boundary__actions">
          <ElButton type="primary" @click="handleRetry">重试</ElButton>
          <ElButton @click="handleGoHome">返回首页</ElButton>
        </div>
      </template>
    </ElResult>
  </div>
</template>

<script setup lang="ts">
  import { computed, onErrorCaptured, ref } from 'vue'
  import { useRouter } from 'vue-router'

  const props = withDefaults(
    defineProps<{
      appKey?: string
      isPublicApp?: boolean
    }>(),
    {
      appKey: '',
      isPublicApp: false
    }
  )

  const router = useRouter()
  const hasError = ref(false)
  const errorHint = ref('请稍后重试，或返回首页恢复。')

  onErrorCaptured((error, instance, info) => {
    hasError.value = true
    errorHint.value = `${info || '运行时错误'}（app=${props.appKey || 'unknown'}）`
    const componentName =
      (instance as { $options?: { name?: string } } | null)?.$options?.name || 'anonymous'
    // telemetry 挂点：保留 APP 维度上下文，方便后续接入 Sentry/日志平台
    console.error('[AppErrorBoundary]', {
      appKey: props.appKey || 'unknown',
      info,
      component: componentName,
      error
    })
    return false
  })

  const handleRetry = () => {
    hasError.value = false
    window.location.reload()
  }

  const handleGoHome = () => {
    hasError.value = false
    const fallback = props.isPublicApp ? '/account/auth/login' : '/dashboard/console'
    void router.replace(fallback)
  }

  defineExpose({
    reset: () => {
      hasError.value = false
      errorHint.value = '请稍后重试，或返回首页恢复。'
    }
  })
</script>

<style scoped lang="scss">
  .art-app-error-boundary {
    width: 100%;
    padding: 24px;
    border-radius: 14px;
    background: linear-gradient(180deg, rgb(248 250 252 / 0.9), rgb(241 245 249 / 0.9));
  }

  .art-app-error-boundary__content {
    width: 100%;
  }

  .art-app-error-boundary__actions {
    display: flex;
    gap: 12px;
    justify-content: center;
  }
</style>
