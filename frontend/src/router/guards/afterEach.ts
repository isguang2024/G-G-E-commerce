import { nextTick } from 'vue'
import { useSettingStore } from '@/store/modules/setting'
import { Router } from 'vue-router'
import NProgress from 'nprogress'
import { loadingService } from '@/utils/ui'
import { getPendingLoading, resetPendingLoading } from '@/domains/navigation/runtime/guard-state'
import { logger } from '@/utils/logger'

function scrollMainContainerToTop(): void {
  const scrollContainer = document.getElementById('app-main')
  if (scrollContainer) {
    scrollContainer.scrollTop = 0
  }
}

/** 路由全局后置守卫 */
export function setupAfterEachGuard(router: Router) {
  router.afterEach((to) => {
    // 同步当前路由给 logger —— 上报的每条日志都带 route 字段，
    // 便于排查"某个页面下的报错"场景。
    logger.setRoute(to.fullPath || to.path || '')

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
