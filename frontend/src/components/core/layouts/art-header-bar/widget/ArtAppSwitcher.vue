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
  import { fetchGetApps } from '@/domains/governance/api'
  import { useAppContextStore } from '@/domains/app-runtime/context'
  import { useUserStore } from '@/domains/auth/store'

  defineOptions({ name: 'ArtAppSwitcher' })
  withDefaults(defineProps<{ compact?: boolean }>(), {
    compact: false
  })

  const APP_SWITCHER_LAST_VISITED_KEY = 'gg:app-switcher:last-visited'

  const appContextStore = useAppContextStore()
  const userStore = useUserStore()
  const { isLogin } = storeToRefs(userStore)

  const loading = ref(false)
  const apps = ref<Api.SystemManage.AppItem[]>([])
  // 首屏 Layout 重挂载会导致 onMounted 发起的 /system/apps 被浏览器兜底 abort，
  // 这里用 AbortController + onBeforeUnmount 把取消收敛为已知事件，
  // 详见 docs/guides/dashboard-request-abort-rootcause.md
  let loadAppsController: AbortController | null = null

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

  const selectedValue = computed(() =>
    `${appContextStore.effectiveManagedAppKey || appContextStore.currentRuntimeAppKey || ''}`.trim()
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

  const shouldShowSwitcher = computed(() => {
    if (sortedApps.value.length <= 1) return false
    return appContextStore.supportsAppSwitchForApp(selectedValue.value)
  })

  const buildOptionLabel = (item: Api.SystemManage.AppItem) => {
    const name = `${item.name || item.appKey || ''}`.trim()
    const appKey = `${item.appKey || ''}`.trim()
    const entry = `${item.frontendEntryUrl || ''}`.trim()
    const suffix = entry ? ` · ${entry}` : ''
    return `${name}${appKey ? ` · ${appKey}` : ''}${suffix}`
  }

  const loadApps = async (force = false) => {
    if (!isLogin.value) {
      apps.value = []
      return
    }
    if (loading.value) return
    if (apps.value.length > 0 && !force) return
    loading.value = true
    // 复用 controller：如果上一次请求还没回来，重新 fire 时先取消旧的
    loadAppsController?.abort('app-switcher-refresh')
    const controller = new AbortController()
    loadAppsController = controller
    try {
      const res = await fetchGetApps({ signal: controller.signal })
      if (controller.signal.aborted) return
      apps.value = res.records || []
      for (const item of apps.value) {
        appContextStore.setAppProfile({
          appKey: item.appKey,
          authMode: item.authMode || '',
          capabilities: item.capabilities || {},
          meta: item.meta || {}
        })
      }
    } catch (error) {
      if (controller.signal.aborted) return
      throw error
    } finally {
      if (loadAppsController === controller) {
        loadAppsController = null
      }
      loading.value = false
    }
  }

  const handleVisibleChange = (visible: boolean) => {
    if (visible && isLogin.value) {
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
      await appContextStore.switchApp(targetApp)
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

  onBeforeUnmount(() => {
    loadAppsController?.abort('component-unmount')
    loadAppsController = null
  })

  watch(
    isLogin,
    (loggedIn) => {
      if (loggedIn) {
        void loadApps(true)
        return
      }
      apps.value = []
    },
    { flush: 'post' }
  )
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
