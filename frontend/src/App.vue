<template>
  <ElConfigProvider size="default" :locale="locales[language]" :z-index="3000">
    <RouterView></RouterView>
  </ElConfigProvider>
</template>

<script setup lang="ts">
  import { LanguageEnum } from '@/enums/appEnum'
  import { useUserStore } from './store/modules/user'
  import { storeToRefs } from 'pinia'
  import zhCn from 'element-plus/es/locale/lang/zh-cn'
  import en from 'element-plus/es/locale/lang/en'
  import { systemUpgrade } from './utils/sys'
  import { toggleTransition } from './utils/ui/animation'
  import { checkStorageCompatibility } from './utils/storage'
  import { initializeTheme } from './hooks/core/useTheme'
  import { initSiteBranding } from '@/domains/site-config/branding'
  import { onBeforeMount, onMounted } from 'vue'

  const userStore = useUserStore()
  const { language } = storeToRefs(userStore)

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
    // 异步加载站点品牌配置（name/logo/favicon），失败时静默维持默认。
    void initSiteBranding()
  })
</script>
