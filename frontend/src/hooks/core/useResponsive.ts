/**
 * 全局共享的视口尺寸 hook。
 *
 * 背景：仓库内多处直接调用 `@vueuse/core` 的 `useWindowSize`，每次调用都会
 * 单独绑定 `resize` 事件监听器。改为单例后整个应用只保留一份监听，避免
 * 重复触发响应式 deps 与回调。
 */
import { useWindowSize } from '@vueuse/core'
import { computed, type ComputedRef, type Ref } from 'vue'

interface SharedViewport {
  width: Ref<number>
  height: Ref<number>
  isMobile: ComputedRef<boolean>
}

const MOBILE_BREAKPOINT = 768

let shared: SharedViewport | null = null

export function useResponsive(breakpoint = MOBILE_BREAKPOINT) {
  if (!shared) {
    const { width, height } = useWindowSize()
    shared = {
      width,
      height,
      isMobile: computed(() => width.value < breakpoint)
    }
  }
  return shared
}
