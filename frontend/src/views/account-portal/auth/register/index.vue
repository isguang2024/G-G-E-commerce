<template>
  <!-- @compat-status: transition account-portal 路径壳层，当前用于承接新注册入口路由名。 -->
  <AuthRegisterPage />
</template>

<script setup lang="ts">
  import AuthRegisterPage from '@/views/auth/register/index.vue'

  defineOptions({ name: 'AccountPortalRegisterPage' })
</script>
