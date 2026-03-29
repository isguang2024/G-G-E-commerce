<!-- 面包屑导航 -->
<template>
  <nav class="ml-2.5 max-lg:!hidden" aria-label="breadcrumb">
    <ul class="flex-c h-full">
      <li
        v-for="(item, index) in breadcrumbItems"
        :key="`${item.path || item.meta?.title || 'breadcrumb'}-${index}`"
        class="box-border flex-c h-7 text-sm leading-7"
      >
        <div
          :class="
            isClickable(item, index)
              ? 'c-p py-1 rounded tad-200 hover:bg-active-color hover:[&_span]:text-g-600'
              : ''
          "
          @click="handleBreadcrumbClick(item, index)"
        >
          <span
            class="block max-w-46 overflow-hidden text-ellipsis whitespace-nowrap px-1.5 text-sm text-g-600 dark:text-g-800"
            >{{ formatMenuTitle(item.meta?.title as string) }}</span
          >
        </div>
        <div
          v-if="!isLastItem(index) && item.meta?.title"
          class="mx-1 text-sm not-italic text-g-500"
          aria-hidden="true"
        >
          /
        </div>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import type { RouteLocationMatched, RouteRecordRaw } from 'vue-router'
  import type { AppRouteRecord } from '@/types/router'
  import { formatMenuTitle } from '@/utils/router'
  import { useMenuStore } from '@/store/modules/menu'

  defineOptions({ name: 'ArtBreadcrumb' })

  export interface BreadcrumbItem {
    path?: string
    meta: RouteRecordRaw['meta']
  }

  const route = useRoute()
  const router = useRouter()
  const menuStore = useMenuStore()

  // 使用computed替代watch，提高性能
  const breadcrumbItems = computed<BreadcrumbItem[]>(() => {
    const { matched } = route
    const matchedLength = matched.length

    // 处理首页情况
    if (!matchedLength || isHomeRoute(matched[0])) {
      return []
    }

    // 获取当前路由（最后一个匹配项）
    const lastIndex = matchedLength - 1
    const currentRoute = matched[lastIndex]
    const currentRouteMeta = currentRoute.meta

    // 优先使用运行时注入的完整面包屑链
    const runtimeChain = Array.isArray(currentRouteMeta?.breadcrumbChain)
      ? currentRouteMeta.breadcrumbChain
      : []
    if (runtimeChain.length) {
      const chainItems = runtimeChain.map((item: any) => ({
        path: normalizePath(String(item?.path || '')),
        meta: {
          title: item?.title || ''
        }
      }))
      return [...chainItems, createBreadcrumbItem(currentRoute)]
    }

    const resolvedMenuChain = resolveMenuChainFromMeta(currentRouteMeta)
    if (resolvedMenuChain.length) {
      return [...resolvedMenuChain, createBreadcrumbItem(currentRoute)]
    }

    // 兼容历史 customParent 逻辑
    if (
      typeof currentRouteMeta?.customParent === 'string' &&
      currentRouteMeta.customParent.trim() !== ''
    ) {
      const customParentPath = normalizePath(String(currentRouteMeta.customParent))

      const parentTitle = findMenuTitle(menuStore.menuList, customParentPath)

      // 直接创建一个自定义的面包屑项
      const customParentItem: BreadcrumbItem = {
        path: customParentPath,
        meta: {
          title: parentTitle || customParentPath.split('/').pop()
        }
      }

      const currentItem = createBreadcrumbItem(currentRoute)
      return [customParentItem, currentItem]
    }

    // 处理顶级菜单（现在所有顶级菜单都有 isFirstLevel 标志）
    if (currentRouteMeta?.isFirstLevel) {
      // 对于顶级菜单，只显示当前菜单
      return [createBreadcrumbItem(currentRoute)]
    }

    // 对于非顶级菜单，显示完整路径
    let items = matched.map(createBreadcrumbItem)

    // 过滤包裹容器：如果有多个项目且第一个是容器路由（如 /outside），则移除它
    if (items.length > 1 && isWrapperContainer(items[0])) {
      items = items.slice(1)
    }

    return items
  })

  // 辅助函数：判断是否为包裹容器路由
  const isWrapperContainer = (item: BreadcrumbItem): boolean =>
    item.path === '/outside' && !!item.meta?.isIframe

  // 辅助函数：创建面包屑项目
  const createBreadcrumbItem = (route: RouteLocationMatched): BreadcrumbItem => ({
    path: normalizePath(route.path),
    meta: route.meta
  })

  // 辅助函数：判断是否为首页
  const isHomeRoute = (route: RouteLocationMatched): boolean => route.name === '/'

  // 辅助函数：判断是否为最后一项
  const isLastItem = (index: number): boolean => {
    const itemsLength = breadcrumbItems.value.length
    return index === itemsLength - 1
  }

  // 辅助函数：判断是否可点击
  const isClickable = (item: BreadcrumbItem, index: number): boolean =>
    Boolean(item.path) && item.path !== '/outside' && !isLastItem(index)

  // 辅助函数：查找路由的第一个有效子路由
  const findFirstValidChild = (route: RouteRecordRaw) =>
    route.children?.find((child) => !child.redirect && !child.meta?.isHide)

  // 辅助函数：构建完整路径
  const buildFullPath = (childPath: string): string => `/${childPath}`.replace('//', '/')

  // 统一路径比较格式（保留根路径 "/"）
  const normalizePath = (path: string): string => {
    if (!`${path || ''}`.trim()) {
      return ''
    }
    const normalized = `/${String(path || '').replace(/^\/+/, '')}`.replace(/\/+/g, '/')
    return normalized !== '/' ? normalized.replace(/\/$/, '') : normalized
  }

  const normalizeMenuResolvedPath = (path: string, parentPath = ''): string => {
    const target = `${path || ''}`.trim()
    if (!target) return normalizePath(parentPath)
    if (/^https?:\/\//i.test(target)) return target
    if (target.startsWith('/')) return normalizePath(target)
    return normalizePath(`${normalizePath(parentPath)}/${target}`)
  }

  const buildMenuChain = (
    menus: AppRouteRecord[],
    targetPath: string,
    parentPath = ''
  ): BreadcrumbItem[] => {
    for (const menu of menus) {
      const currentPath = normalizeMenuResolvedPath(String(menu.path || ''), parentPath)
      const currentItem: BreadcrumbItem = {
        path: currentPath,
        meta: menu.meta
      }
      if (currentPath === targetPath) {
        return [currentItem]
      }
      if (menu.children?.length) {
        const childChain = buildMenuChain(menu.children, targetPath, currentPath)
        if (childChain.length) {
          return [currentItem, ...childChain]
        }
      }
    }
    return []
  }

  const findMenuTitle = (menus: AppRouteRecord[], path: string, parentPath = ''): string | undefined => {
    for (const menu of menus) {
      const currentPath = normalizeMenuResolvedPath(String(menu.path || ''), parentPath)
      if (currentPath === path) {
        return String(menu.meta?.title || '')
      }
      if (menu.children?.length) {
        const found = findMenuTitle(menu.children, path, currentPath)
        if (found) return found
      }
    }
    return undefined
  }

  const resolveMenuChainFromMeta = (meta?: RouteRecordRaw['meta']): BreadcrumbItem[] => {
    const candidatePaths = [
      normalizePath(String(meta?.activePath || '')),
      normalizePath(String(meta?.customParent || ''))
    ].filter(Boolean)

    for (const candidatePath of candidatePaths) {
      const chain = buildMenuChain(menuStore.menuList, candidatePath)
      if (chain.length) {
        return chain
      }
    }
    return []
  }

  // 处理面包屑点击事件
  async function handleBreadcrumbClick(item: BreadcrumbItem, index: number): Promise<void> {
    // 如果是最后一项或外部链接，不处理
    if (isLastItem(index) || !item.path || item.path === '/outside') {
      return
    }

    try {
      // 缓存路由表查找结果
      const routes = router.getRoutes()
      const targetRoute = routes.find((route) => route.path === item.path)

      if (!targetRoute?.children?.length) {
        await router.push(item.path)
        return
      }

      const firstValidChild = findFirstValidChild(targetRoute)
      if (firstValidChild) {
        await router.push(buildFullPath(firstValidChild.path))
      } else {
        await router.push(item.path)
      }
    } catch (error) {
      console.error('导航失败:', error)
    }
  }
</script>
