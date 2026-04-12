<template>
  <template v-for="(item, index) in filteredMenuItems" :key="getUniqueKey(item, index)">
    <ElSubMenu v-if="hasChildren(item)" :index="item.path || item.meta.title" :level="level">
      <template #title>
        <div class="menu-icon flex-cc">
          <ArtSvgIcon
            :icon="item.meta.icon"
            :color="theme?.iconColor"
            :style="{ color: theme.iconColor }"
          />
        </div>
        <span class="menu-name">
          {{ formatMenuTitle(item.meta.title) }}
        </span>
        <div v-if="item.meta.showBadge" class="art-badge" style="right: 10px" />
      </template>

      <SidebarSubmenu
        :list="item.children"
        :is-mobile="isMobile"
        :level="level + 1"
        :theme="theme"
        @close="closeMenu"
      />
    </ElSubMenu>

    <ElMenuItem
      v-else
      :index="isExternalLink(item) ? undefined : item.path || item.meta.title"
      :level-item="level + 1"
      @click="goPage(item)"
    >
      <div class="menu-icon flex-cc">
        <ArtSvgIcon
          :icon="item.meta.icon"
          :color="theme?.iconColor"
          :style="{ color: theme.iconColor }"
        />
      </div>
      <div
        v-show="item.meta.showBadge && level === 0 && !menuOpen"
        class="art-badge"
        style="right: 5px"
      />

      <template #title>
        <span class="menu-name">
          {{ formatMenuTitle(item.meta.title) }}
        </span>
        <div v-if="item.meta.showBadge" class="art-badge" />
        <div v-if="item.meta.showTextBadge && (level > 0 || menuOpen)" class="art-text-badge">
          {{ item.meta.showTextBadge }}
        </div>
      </template>
    </ElMenuItem>
  </template>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import type { AppRouteRecord } from '@/types/router'
  import { formatMenuTitle } from '@/utils/router'
  import { handleMenuJump } from '@/domains/navigation/utils'
  import { useSettingStore } from '@/store/modules/setting'

  interface MenuTheme {
    iconColor?: string
  }

  interface Props {
    /** 菜单标题 */
    title?: string
    /** 菜单列表 */
    list?: AppRouteRecord[]
    /** 主题配置 */
    theme?: MenuTheme
    /** 是否为移动端模式 */
    isMobile?: boolean
    /** 菜单层级 */
    level?: number
  }

  interface Emits {
    /** 关闭菜单事件 */
    (e: 'close'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    title: '',
    list: () => [],
    theme: () => ({}),
    isMobile: false,
    level: 0
  })

  const emit = defineEmits<Emits>()

  const settingStore = useSettingStore()

  const { menuOpen } = storeToRefs(settingStore)

  /**
   * 过滤后的菜单项列表
   * 只显示未隐藏的菜单项
   */
  const filteredMenuItems = computed(() => filterRoutes(props.list))

  /**
   * 跳转到指定页面
   * @param item 菜单项数据
   */
  const goPage = (item: AppRouteRecord): void => {
    closeMenu()
    handleMenuJump(item)
  }

  /**
   * 关闭菜单
   * 触发父组件的关闭事件
   */
  const closeMenu = (): void => {
    emit('close')
  }

  /**
   * 递归过滤菜单路由，移除隐藏的菜单项（如 isHide 的内页）
   * 当某菜单下仅有内页等隐藏子项时，仍保留该菜单在侧栏，以叶子节点展示（可点击进入该菜单页）
   */
  const filterRoutes = (items: AppRouteRecord[]): AppRouteRecord[] => {
    return items
      .filter((item) => {
        if (item.meta?.isHide) return false
        if (item.children?.length) return true
        return true
      })
      .map((item) => ({
        ...item,
        children: item.children?.length ? filterRoutes(item.children) : undefined
      }))
  }

  /**
   * 判断菜单项是否包含可见的子菜单
   * @param item 菜单项数据
   * @returns 是否包含可见的子菜单
   */
  const hasChildren = (item: AppRouteRecord): boolean => {
    if (!item.children || item.children.length === 0) {
      return false
    }
    // 递归检查是否有可见的子菜单
    const filteredChildren = filterRoutes(item.children)
    return filteredChildren.length > 0
  }

  /**
   * 判断是否为外部链接
   * @param item 菜单项数据
   * @returns 是否为外部链接
   */
  const isExternalLink = (item: AppRouteRecord): boolean => {
    return !!(item.meta.link && !item.meta.isIframe)
  }

  /**
   * 生成唯一的 key
   * 使用 path、title 和 index 组合确保唯一性
   * @param item 菜单项数据
   * @param index 索引
   * @returns 唯一的 key
   */
  const getUniqueKey = (item: AppRouteRecord, index: number): string => {
    return `${item.path || item.meta.title || 'menu'}-${props.level}-${index}`
  }
</script>
