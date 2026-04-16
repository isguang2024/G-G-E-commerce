<!-- 系统logo -->
<template>
  <div class="flex-cc">
    <img :style="logoStyle" :src="logoSrc" alt="logo" class="w-full h-full" />
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import defaultLogo from '@imgs/common/logo.webp'
  import { useSiteBranding } from '@/domains/site-config/branding'

  defineOptions({ name: 'ArtLogo' })

  interface Props {
    /** logo 大小 */
    size?: number | string
  }

  const props = withDefaults(defineProps<Props>(), {
    size: 36
  })

  const { logo } = useSiteBranding()
  const logoSrc = computed(() => logo.value || defaultLogo)
  const logoStyle = computed(() => ({ width: `${props.size}px` }))
</script>
