/**
 * useFastEnter - 快速入口管理
 *
 * 管理顶部栏的快速入口功能，提供应用列表和快速链接的配置和过滤。
 * 支持动态启用/禁用、自定义排序、响应式宽度控制等功能。
 *
 * ## 主要功能
 *
 * 1. 应用列表管理 - 获取启用的应用列表，自动按排序权重排序
 * 2. 快速链接管理 - 获取启用的快速链接，支持自定义排序
 * 3. 响应式配置 - 所有配置自动响应变化，无需手动更新
 * 4. 宽度控制 - 提供最小显示宽度配置，支持响应式布局
 *
 * @module useFastEnter
 * @author Art Design Pro Team
 */

import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useFastEnterStore } from '@/store/modules/fast-enter'
import { useMenuSpaceStore } from '@/store/modules/menu-space'
import type { FastEnterApplication, FastEnterQuickLink } from '@/types/config'
import { findRegisteredRouteByPath, hasRegisteredRoutePath } from '@/utils/router'

const FAST_ENTER_DEFAULT_MIN_WIDTH = 1450
const COMPACT_MESSAGE_WORKSPACE_ROUTE_NAMES = new Set([
  'MessageTemplateManage',
  'MessageRecordManage',
  'MessageSenderManage',
  'MessageRecipientGroupManage',
  'TeamMessageManage',
  'TeamMessageTemplateManage',
  'TeamMessageRecordManage',
  'TeamMessageSenderManage',
  'TeamMessageRecipientGroupManage'
])

export function useFastEnter() {
  const router = useRouter()
  const fastEnterStore = useFastEnterStore()
  const menuSpaceStore = useMenuSpaceStore()
  const { config: fastEnterConfig } = storeToRefs(fastEnterStore)

  const isExternalLink = (value?: string) => /^https?:\/\//i.test(`${value || ''}`.trim())
  const normalizeTargetKey = (item: FastEnterApplication | FastEnterQuickLink) => {
    const routeName = `${item.routeName || ''}`.trim()
    if (routeName) return `route:${routeName}`

    const link = `${item.link || ''}`.trim()
    if (!link) return ''
    return isExternalLink(link) ? `link:${link.toLowerCase()}` : `path:${link}`
  }

  const isAllowedItem = (item: FastEnterApplication | FastEnterQuickLink) => {
    if (item.enabled === false) return false

    const routeName = `${item.routeName || ''}`.trim()
    const link = `${item.link || ''}`.trim()

    if (routeName && COMPACT_MESSAGE_WORKSPACE_ROUTE_NAMES.has(routeName)) {
      return false
    }

    if (routeName && router.hasRoute(routeName)) {
      return true
    }

    if (isExternalLink(link)) {
      return true
    }

    if (link.startsWith('/')) {
      return hasRegisteredRoutePath(router, link)
    }

    return false
  }

  const dedupeItems = <T extends FastEnterApplication | FastEnterQuickLink>(items: T[]) => {
    const seen = new Set<string>()
    return items.filter((item) => {
      const targetKey = normalizeTargetKey(item)
      if (!targetKey) return false
      if (seen.has(targetKey)) return false
      seen.add(targetKey)
      return true
    })
  }

  // 获取启用的应用列表（按排序权重排序）
  const enabledApplications = computed<FastEnterApplication[]>(() => {
    if (!fastEnterConfig.value?.applications) return []

    return dedupeItems(
      [...fastEnterConfig.value.applications]
        .filter(isAllowedItem)
        .sort((a, b) => (a.order || 0) - (b.order || 0))
    )
  })

  // 获取启用的快速链接（按排序权重排序）
  const enabledQuickLinks = computed<FastEnterQuickLink[]>(() => {
    if (!fastEnterConfig.value?.quickLinks) return []

    return dedupeItems(
      [...fastEnterConfig.value.quickLinks]
        .filter(isAllowedItem)
        .sort((a, b) => (a.order || 0) - (b.order || 0))
    )
  })

  // 获取最小显示宽度
  const minWidth = computed(() => {
    return fastEnterConfig.value?.minWidth || FAST_ENTER_DEFAULT_MIN_WIDTH
  })

  const openNavigationTarget = async (routeName?: string, link?: string) => {
    const routeTarget = `${routeName || ''}`.trim()
    const linkTarget = `${link || ''}`.trim()

    if (routeTarget && router.hasRoute(routeTarget)) {
      const resolved = router.resolve({ name: routeTarget })
      const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
        resolved.fullPath || resolved.path,
        `${resolved.meta?.spaceKey || ''}`.trim() || undefined
      )
      if (nextTarget.mode === 'router') {
        await router.push(nextTarget.target)
      } else {
        window.location.assign(nextTarget.target)
      }
      return true
    }

    if (isExternalLink(linkTarget)) {
      window.open(linkTarget, '_blank')
      return true
    }

    if (linkTarget.startsWith('/')) {
      const resolvedRoute = findRegisteredRouteByPath(router, linkTarget)
      if (!resolvedRoute) {
        return false
      }
      const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
        linkTarget,
        `${resolvedRoute.meta?.spaceKey || ''}`.trim() || undefined
      )
      if (nextTarget.mode === 'router') {
        await router.push(linkTarget)
      } else {
        window.location.assign(nextTarget.target)
      }
      return true
    }

    return false
  }

  return {
    fastEnterConfig,
    enabledApplications,
    enabledQuickLinks,
    minWidth,
    openNavigationTarget
  }
}
