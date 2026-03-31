<template>
  <div class="page-content !border-0 !bg-transparent min-h-screen flex-cc">
    <div class="flex-cc max-md:!block max-md:text-center">
      <ThemeSvg :src="data.imgUrl" size="100%" class="!w-100" />
      <div class="ml-15 w-75 max-md:mx-auto max-md:mt-10 max-md:w-full max-md:text-center">
        <p class="text-xl leading-7 text-g-600 max-md:text-lg">{{ data.desc }}</p>
        <div class="mt-5 flex items-center gap-3 max-md:justify-center">
          <ElButton type="primary" size="large" @click="backHome" v-ripple>
            {{ data.btnText }}
          </ElButton>
          <ElButton size="large" @click="backPrevious" v-ripple>
            返回上一层
          </ElButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useCommon } from '@/hooks/core/useCommon'

  const router = useRouter()

  interface ExceptionData {
    /** 标题 */
    title: string
    /** 描述 */
    desc: string
    /** 按钮文本 */
    btnText: string
    /** 图片地址 */
    imgUrl: string
  }

  withDefaults(
    defineProps<{
      data: ExceptionData
    }>(),
    {}
  )

  const { homePath } = useCommon()

  const backHome = () => {
    router.push(homePath.value)
  }

  const backPrevious = () => {
    if (window.history.length > 1) {
      router.back()
      return
    }
    router.push(homePath.value)
  }
</script>
