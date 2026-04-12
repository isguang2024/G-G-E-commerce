import { nextTick } from 'vue'
import { useSettingStore } from '@/store/modules/setting'
import { Router } from 'vue-router'
import NProgress from 'nprogress'
import { loadingService } from '@/utils/ui'
import { getPendingLoading, resetPendingLoading } from '@/domains/navigation/runtime/guard-state'

function scrollMainContainerToTop(): void {
  const scrollContainer = document.getElementById('app-main')
  if (scrollContainer) {
    scrollContainer.scrollTop = 0
  }
}

/** 路由全局后置守卫 */
export function setupAfterEachGuard(router: Router) {
  router.afterEach(() => {
    scrollMainContainerToTop()

    // 关闭进度条
    const settingStore = useSettingStore()
    if (settingStore.showNprogress) {
      NProgress.done()
      // 确保进度条完全移除，避免残影
      setTimeout(() => {
        NProgress.remove()
      }, 600)
    }

    // 关闭 loading 效果
    if (getPendingLoading()) {
      nextTick(() => {
        loadingService.hideLoading()
        resetPendingLoading()
      })
    }
  })
}
