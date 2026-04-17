<template>
  <ElConfigProvider size="default" :locale="locales[language]" :z-index="3000">
    <RouterView></RouterView>
  </ElConfigProvider>
</template>

<script setup lang="ts">
  import { storeToRefs } from 'pinia'
  import { onBeforeMount, onMounted, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { initSiteBranding } from '@/domains/site-config/branding'
  import { LanguageEnum } from '@/enums/appEnum'
  import zhCn from 'element-plus/es/locale/lang/zh-cn'
  import en from 'element-plus/es/locale/lang/en'
  import { useUserStore } from './store/modules/user'
  import { systemUpgrade } from './utils/sys'
  import { initializeTheme } from './hooks/core/useTheme'
  import { checkStorageCompatibility } from './utils/storage'
  import { toggleTransition } from './utils/ui/animation'

  const userStore = useUserStore()
  const route = useRoute()
  const { isLogin, language } = storeToRefs(userStore)

  const locales: Record<LanguageEnum, typeof zhCn> = {
    [LanguageEnum.ZH]: zhCn,
    [LanguageEnum.EN]: en
  }

  onBeforeMount(() => {
    toggleTransition(true)
    initializeTheme()
  })

  onMounted(() => {
    checkStorageCompatibility()
    toggleTransition(false)
    systemUpgrade()
    void initSiteBranding()
  })

  watch(
    () => [route.path, isLogin.value] as const,
    ([path, loggedIn]) => {
      if (loggedIn || !path.startsWith('/account/auth/')) {
        void initSiteBranding()
      }
    },
    { immediate: true }
  )
</script>
