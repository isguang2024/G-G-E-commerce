<template>
  <!-- @compat-status: transition account-portal 路径壳层，当前用于承接新登录入口路由名。 -->
  <AuthLoginPage />
</template>

<script setup lang="ts">
  import AuthLoginPage from '@/views/auth/login/index.vue'

  defineOptions({ name: 'AccountPortalLoginPage' })
</script>
