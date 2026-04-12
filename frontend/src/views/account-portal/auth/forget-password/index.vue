<template>
  <!-- @compat-status: transition account-portal 路径壳层，当前用于承接新找回密码入口路由名。 -->
  <AuthForgetPasswordPage />
</template>

<script setup lang="ts">
  import AuthForgetPasswordPage from '@/views/auth/forget-password/index.vue'

  defineOptions({ name: 'AccountPortalForgetPasswordPage' })
</script>
