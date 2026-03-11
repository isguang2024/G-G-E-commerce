<!-- 图标组件 -->
<template>
  <Icon v-if="icon" :icon="icon" v-bind="bindAttrs" class="art-svg-icon inline" />
</template>

<script setup lang="ts">
  import { computed, useAttrs } from 'vue'
  import { Icon } from '@iconify/vue'

  defineOptions({ name: 'ArtSvgIcon', inheritAttrs: false })

  interface Props {
    /** Iconify icon name */
    icon?: string
  }

  defineProps<Props>()

  const attrs = useAttrs()

  const bindAttrs = computed(() => {
    const filteredAttrs: any = {}
    for (const [key, value] of Object.entries(attrs)) {
      if (key !== 'class' && key !== 'style') {
        filteredAttrs[key] = value
      }
    }
    return {
      class: (attrs.class as string) || '',
      style: (attrs.style as string) || '',
      ...filteredAttrs
    }
  })
</script>
