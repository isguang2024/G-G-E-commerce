<template>
  <div
    v-if="shouldShowSwitcher"
    class="app-switcher"
    :class="{ 'max-lg:!hidden': !compact, 'app-switcher--compact': compact }"
  >
    <div v-if="compact" class="app-switcher__label">当前应用</div>
    <ElSelect
      :model-value="selectedValue"
      class="app-switcher__select"
      placeholder="切换应用"
      :loading="loading"
      filterable
      @visible-change="handleVisibleChange"
      @change="handleChange"
    >
      <ElOption
        v-for="item in sortedApps"
        :key="item.appKey"
        :label="buildOptionLabel(item)"
        :value="item.appKey"
      />
    </ElSelect>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { fetchGetApps } from '@/api/system-manage'
  import { refreshUserMenus } from '@/router'
  import { IframeRouteManager } from '@/router/core'
  import { useAppContextStore } from '@/store/modules/app-context'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { useMenuStore } from '@/store/modules/menu'
  import { useWorktabStore } from '@/store/modules/worktab'
  import { findRegisteredRouteByPath } from '@/utils/router'

  defineOptions({ name: 'ArtAppSwitcher' })
  withDefaults(defineProps<{ compact?: boolean }>(), {
    compact: false
  })

  const APP_SWITCHER_LAST_VISITED_KEY = 'gg:app-switcher:last-visited'

  const router = useRouter()
  const appContextStore = useAppContextStore()
  const menuStore = useMenuStore()
  const menuSpaceStore = useMenuSpaceStore()
  const worktabStore = useWorktabStore()

  const loading = ref(false)
  const apps = ref<Api.SystemManage.AppItem[]>([])

  const readLastVisitedMap = () => {
    if (typeof window === 'undefined') {
      return {} as Record<string, number>
    }
    try {
      const raw = window.localStorage.getItem(APP_SWITCHER_LAST_VISITED_KEY)
      const parsed = raw ? JSON.parse(raw) : {}
      return parsed && typeof parsed === 'object' ? (parsed as Record<string, number>) : {}
    } catch {
      return {}
    }
  }

  const writeLastVisited = (appKey: string) => {
    if (typeof window === 'undefined' || !appKey) {
      return
    }
    const nextMap = {
      ...readLastVisitedMap(),
      [appKey]: Date.now()
    }
    window.localStorage.setItem(APP_SWITCHER_LAST_VISITED_KEY, JSON.stringify(nextMap))
  }

  const selectedValue = computed(
    () => `${appContextStore.effectiveManagedAppKey || appContextStore.currentRuntimeAppKey || ''}`.trim()
  )

  const sortedApps = computed(() => {
    const lastVisitedMap = readLastVisitedMap()
    return [...apps.value].sort((a, b) => {
      const aIsCurrent = a.appKey === selectedValue.value ? 1 : 0
      const bIsCurrent = b.appKey === selectedValue.value ? 1 : 0
      if (aIsCurrent !== bIsCurrent) return bIsCurrent - aIsCurrent

      const aVisited = Number(lastVisitedMap[a.appKey || ''] || 0)
      const bVisited = Number(lastVisitedMap[b.appKey || ''] || 0)
      if (aVisited !== bVisited) return bVisited - aVisited

      const aDefault = a.isDefault ? 1 : 0
      const bDefault = b.isDefault ? 1 : 0
      if (aDefault !== bDefault) return bDefault - aDefault

      return `${a.name || a.appKey || ''}`.localeCompare(`${b.name || b.appKey || ''}`, 'zh-CN')
    })
  })

  const shouldShowSwitcher = computed(() => sortedApps.value.length > 1)

  const buildOptionLabel = (item: Api.SystemManage.AppItem) => {
    const name = `${item.name || item.appKey || ''}`.trim()
    const appKey = `${item.appKey || ''}`.trim()
    const entry = `${item.frontendEntryUrl || ''}`.trim()
    const suffix = entry ? ` · ${entry}` : ''
    return `${name}${appKey ? ` · ${appKey}` : ''}${suffix}`
  }

  const normalizeInternalPath = (value?: string) => {
    const raw = `${value || ''}`.trim()
    if (!raw || /^https?:\/\//i.test(raw)) {
      return raw
    }
    return raw.startsWith('/') ? raw : `/${raw}`
  }

  const resolveAppEntryTarget = (targetApp: Api.SystemManage.AppItem) => {
    const candidates = [
      menuStore.getHomePath(),
      menuSpaceStore.resolveSpaceLandingPath(),
      targetApp.frontendEntryUrl,
      '/'
    ]
      .map((item) => normalizeInternalPath(item))
      .filter(Boolean)

    for (const candidate of candidates) {
      if (/^https?:\/\//i.test(candidate)) {
        return {
          entryTarget: candidate,
          targetPath: candidate,
          spaceKey: undefined
        }
      }

      const resolvedRoute = findRegisteredRouteByPath(router, candidate)
      if (resolvedRoute) {
        return {
          entryTarget: candidate,
          targetPath: candidate,
          spaceKey: `${resolvedRoute.meta?.spaceKey || ''}`.trim() || undefined
        }
      }
    }

    const fallbackPath = normalizeInternalPath(menuStore.getHomePath() || menuSpaceStore.resolveSpaceLandingPath() || '/')
    return {
      entryTarget: fallbackPath,
      targetPath: fallbackPath,
      spaceKey: undefined
    }
  }

  const loadApps = async (force = false) => {
    if (loading.value) return
    if (apps.value.length > 0 && !force) return
    loading.value = true
    try {
      const res = await fetchGetApps()
      apps.value = res.records || []
    } finally {
      loading.value = false
    }
  }

  const handleVisibleChange = (visible: boolean) => {
    if (visible) {
      void loadApps(true)
    }
  }

  const handleChange = async (value: string) => {
    const nextAppKey = `${value || ''}`.trim()
    if (!nextAppKey || nextAppKey === selectedValue.value) {
      return
    }

    const targetApp = sortedApps.value.find((item) => item.appKey === nextAppKey)
    if (!targetApp) {
      ElMessage.warning('未找到目标应用')
      return
    }

    try {
      loading.value = true
      writeLastVisited(nextAppKey)
      appContextStore.setRuntimeAppContext({
        appKey: nextAppKey,
        frontendEntryUrl: targetApp.frontendEntryUrl || '',
        backendEntryUrl: targetApp.backendEntryUrl || '',
        healthCheckUrl: targetApp.healthCheckUrl || '',
        capabilities: targetApp.capabilities || {}
      })
      appContextStore.setManagedAppKey(nextAppKey)
      worktabStore.clearAll()
      IframeRouteManager.getInstance().clear()
      menuStore.removeAllDynamicRoutes()
      menuStore.setMenuList([])
      menuSpaceStore.clearActiveSpaceKey()

      await menuSpaceStore.refreshRuntimeConfig(true)
      const preferredSpaceKey = `${targetApp.defaultSpaceKey || ''}`.trim()
      if (preferredSpaceKey) {
        menuSpaceStore.setActiveSpaceKey(preferredSpaceKey)
      }
      await menuSpaceStore.syncResolvedCurrentSpace(preferredSpaceKey)
      await refreshUserMenus()

      const { entryTarget, targetPath, spaceKey } = resolveAppEntryTarget(targetApp)
      if (/^https?:\/\//i.test(entryTarget)) {
        window.location.assign(entryTarget)
        return
      }
      const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
        targetPath,
        spaceKey
      )
      if (nextTarget.mode === 'router') {
        await router.replace(nextTarget.target)
      } else {
        window.location.assign(nextTarget.target)
        return
      }
      ElMessage.success(`已切换到 ${targetApp.name || targetApp.appKey}`)
    } catch (error) {
      console.error('[AppSwitcher] 切换应用失败:', error)
      ElMessage.error('切换应用失败')
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void loadApps()
  })
</script>

<style scoped lang="scss">
  .app-switcher {
    min-width: 220px;
    margin-right: 6px;
  }

  .app-switcher--compact {
    width: 100%;
    min-width: 0;
    margin-right: 0;
  }

  .app-switcher__label {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--art-text-gray-600);
  }

  .app-switcher__select {
    width: 100%;
  }

  :deep(.app-switcher__select .el-select__wrapper) {
    min-height: 38px;
    border-radius: 12px;
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.92), transparent 55%),
      linear-gradient(135deg, rgb(248 250 252 / 0.98), rgb(241 245 249 / 0.94));
    box-shadow: inset 0 0 0 1px rgb(226 232 240 / 0.95);
  }
</style>
