<template>
  <div class="fast-enter-page art-full-height">
    <AdminWorkspaceHero
      title="快捷应用管理"
      description="按真实展示维护顶部快捷入口。列表尽量精简，内部跳转优先从菜单树中选择，外链保留独立入口。"
      :metrics="summaryMetrics"
    >
      <div class="fast-enter-hero-actions">
        <ElSelect
          v-model="selectedAppKey"
          clearable
          filterable
          placeholder="选择 App"
          class="fast-enter-app-select"
          @change="handleManagedAppChange"
        >
          <ElOption
            v-for="item in appOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
        <ElButton type="primary" @click="saveConfig" v-ripple>保存配置</ElButton>
        <ElButton @click="restoreDraft" v-ripple>撤销未保存修改</ElButton>
        <ElButton type="danger" plain @click="resetToDefault" v-ripple>恢复默认</ElButton>
      </div>
    </AdminWorkspaceHero>

    <ElCard class="fast-enter-shell art-card-xs" shadow="never">
      <div class="fast-enter-shell__toolbar">
        <div class="fast-enter-shell__toolbar-main">
          <div class="fast-enter-shell__title">顶部展示控制</div>
          <p class="fast-enter-shell__note">配置尽量贴近实际展示，不在这里堆大表单。点击条目后再在抽屉里编辑细节。</p>
        </div>
      </div>

      <div class="fast-enter-shell__divider" />

      <div class="fast-enter-board">
        <ElCard class="fast-enter-panel art-card-xs" shadow="never">
          <header class="fast-enter-panel__header">
            <div>
              <div class="fast-enter-panel__title">快捷应用</div>
              <p class="fast-enter-panel__desc">左侧卡片区，适合放高频后台入口和外部工具。</p>
            </div>
            <ElButton type="primary" plain @click="openCreateDrawer('application')">新增应用</ElButton>
          </header>

          <div class="fast-enter-preview fast-enter-preview--apps">
            <button
              v-for="(item, index) in sortedApplications"
              :key="item.id"
              type="button"
              class="fast-enter-preview-card"
              @click="openEditDrawer('application', index)"
            >
              <div class="fast-enter-preview-card__icon">
                <ArtSvgIcon :icon="item.icon || 'ri:apps-2-line'" :style="{ color: item.iconColor || '#377dff' }" />
              </div>
              <div class="fast-enter-preview-card__content">
                <div class="fast-enter-preview-card__head">
                  <span class="fast-enter-preview-card__name">{{ item.name || '未命名应用' }}</span>
                  <span class="fast-enter-preview-card__order">#{{ item.order || index + 1 }}</span>
                </div>
                <p class="fast-enter-preview-card__desc">
                  {{ item.description || '补充一行说明，避免只有名字没有语义。' }}
                </p>
                <div class="fast-enter-preview-card__meta">
                  <span class="fast-enter-badge" :class="{ 'is-muted': item.enabled === false }">
                    {{ item.enabled === false ? '已停用' : '已启用' }}
                  </span>
                  <span class="fast-enter-badge fast-enter-badge--soft">
                    {{ resolveTargetLabel(item) }}
                  </span>
                </div>
              </div>
            </button>
          </div>
        </ElCard>

        <ElCard class="fast-enter-panel art-card-xs" shadow="never">
          <header class="fast-enter-panel__header">
            <div>
              <div class="fast-enter-panel__title">快捷链接</div>
              <p class="fast-enter-panel__desc">右侧轻量链接区，适合放帮助页、个人页或外部文档。</p>
            </div>
            <ElButton type="primary" plain @click="openCreateDrawer('quickLink')">新增链接</ElButton>
          </header>

          <div class="fast-enter-preview fast-enter-preview--links">
            <button
              v-for="(item, index) in sortedQuickLinks"
              :key="item.id"
              type="button"
              class="fast-enter-link-row"
              @click="openEditDrawer('quickLink', index)"
            >
              <div class="fast-enter-link-row__main">
                <span class="fast-enter-link-row__name">{{ item.name || '未命名链接' }}</span>
                <span class="fast-enter-link-row__target">{{ resolveTargetLabel(item) }}</span>
              </div>
              <div class="fast-enter-link-row__side">
                <span class="fast-enter-link-row__status">{{ item.enabled === false ? '停用' : '启用' }}</span>
                <ArtSvgIcon icon="ri:arrow-right-s-line" />
              </div>
            </button>
          </div>
        </ElCard>
      </div>
    </ElCard>

    <ElDrawer
      v-model="drawerVisible"
      :title="drawerTitle"
      size="520px"
      destroy-on-close
      append-to-body
      class="fast-enter-drawer"
    >
      <template v-if="drawerDraft">
        <ElCard class="fast-enter-drawer__summary art-card-xs" shadow="never">
          <div class="fast-enter-drawer__summary-icon" v-if="applicationDrawerDraft">
            <ArtSvgIcon
              :icon="applicationDrawerDraft.icon || 'ri:apps-2-line'"
              :style="{ color: applicationDrawerDraft.iconColor || '#377dff' }"
            />
          </div>
          <div class="fast-enter-drawer__summary-body">
            <div class="fast-enter-drawer__summary-title">{{ drawerDraft.name || '未命名入口' }}</div>
            <div class="fast-enter-drawer__summary-text">
              {{ drawerMode === 'application' ? '用于顶部左侧卡片区' : '用于顶部右侧轻量链接区' }}
            </div>
          </div>
        </ElCard>

        <div class="fast-enter-drawer__form">
          <ElFormItem label="名称">
            <ElInput v-model="drawerDraft.name" placeholder="例如 页面管理" />
          </ElFormItem>

          <ElFormItem label="排序">
            <ElInputNumber v-model="drawerDraft.order" :min="1" :max="999" style="width: 100%" />
          </ElFormItem>

          <ElFormItem label="启用状态">
            <ElSwitch v-model="drawerDraft.enabled" inline-prompt active-text="开" inactive-text="关" />
          </ElFormItem>

          <ElFormItem v-if="applicationDrawerDraft" label="描述">
            <ElInput
              v-model="applicationDrawerDraft.description"
              type="textarea"
              :rows="3"
              placeholder="简短说明这个快捷应用用于做什么"
            />
          </ElFormItem>

          <template v-if="applicationDrawerDraft">
            <ElFormItem label="图标">
              <ElInput v-model="applicationDrawerDraft.icon" placeholder="例如 ri:apps-2-line" />
            </ElFormItem>

            <ElFormItem label="图标色">
              <ElColorPicker v-model="applicationDrawerDraft.iconColor" show-alpha />
            </ElFormItem>
          </template>

          <ElFormItem label="跳转方式">
            <ElRadioGroup v-model="drawerTargetType">
              <ElRadioButton value="route">内部路由</ElRadioButton>
              <ElRadioButton value="path">内部路径</ElRadioButton>
              <ElRadioButton value="link">外部链接</ElRadioButton>
            </ElRadioGroup>
          </ElFormItem>

          <ElFormItem v-if="drawerTargetType === 'route'" label="内部路由">
            <ElTreeSelect
              v-model="drawerDraft.routeName"
              class="w-full"
              :data="menuRouteTree"
              node-key="value"
              clearable
              filterable
              check-strictly
              :default-expand-all="false"
              :expand-on-click-node="false"
              :render-after-expand="false"
              :props="treeSelectProps"
              placeholder="搜索菜单树并选择对应内部路由"
              @change="handleRouteChange"
            />
          </ElFormItem>

          <div v-if="drawerTargetType === 'route'" class="fast-enter-field-hint">
            <span>推荐：直接从菜单树选择内部路由。</span>
            <span>保存值是路由名，例如 `Console`、`User`、`PageManagement`。</span>
          </div>

          <ElFormItem v-else-if="drawerTargetType === 'path'" label="内部路径">
            <ElInput v-model="drawerDraft.link" placeholder="/system/user 或 /team/members" />
          </ElFormItem>

          <div v-if="drawerTargetType === 'path'" class="fast-enter-field-hint">
            <span>支持以 `/` 开头的站内路径。</span>
            <span>例如 `/system/user`、`/system/api-endpoint`、`/team/members`。</span>
          </div>

          <ElFormItem v-else label="外部链接">
            <ElInput v-model="drawerDraft.link" placeholder="https://..." />
          </ElFormItem>

          <div v-if="drawerTargetType === 'link'" class="fast-enter-field-hint">
            <span>仅支持完整外链地址。</span>
            <span>例如 `https://docs.example.com` 或 `http://127.0.0.1:9000`。</span>
          </div>

          <div class="fast-enter-drawer__hint">
            <span>当前支持 3 种写法：菜单树路由名、以 `/` 开头的内部路径、`http/https` 外部链接。</span>
            <span>内部跳转会跟随当前菜单权限和页面可访问结果自动过滤；外链不受菜单权限裁剪。</span>
            <span>消息模块建议只保留消息发送主入口和消息中心，模板、发送人、接收组、发送记录统一从消息页内导航进入。</span>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="fast-enter-drawer__footer">
          <ElButton v-if="drawerEditing" type="danger" plain @click="removeCurrentItem">删除</ElButton>
          <div class="fast-enter-drawer__footer-actions">
            <ElButton @click="drawerVisible = false">取消</ElButton>
            <ElButton type="primary" @click="applyDrawer">确定</ElButton>
          </div>
        </div>
      </template>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { storeToRefs } from 'pinia'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { fetchGetApps, fetchGetMenuTreeAll } from '@/api/system-manage'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import { getDefaultFastEnterConfig, useFastEnterStore } from '@/store/modules/fast-enter'
  import type { AppRouteRecord } from '@/types/router'
  import type { FastEnterApplication, FastEnterConfig, FastEnterQuickLink } from '@/types/config'

  defineOptions({ name: 'FastEnterManage' })

  type DrawerMode = 'application' | 'quickLink'
  type DrawerTargetType = 'route' | 'path' | 'link'

  interface RouteTreeOption {
    label: string
    value: string
    path: string
    children?: RouteTreeOption[]
  }

  const compactMessageWorkspaceRouteNames = new Set([
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

  type DrawerApplicationDraft = FastEnterApplication & { id: string }
  type DrawerQuickLinkDraft = FastEnterQuickLink & { id: string }
  type DrawerDraft = DrawerApplicationDraft | DrawerQuickLinkDraft

  const fastEnterStore = useFastEnterStore()
  const { config } = storeToRefs(fastEnterStore)
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )

  const cloneConfig = (value: FastEnterConfig): FastEnterConfig =>
    JSON.parse(JSON.stringify(value)) as FastEnterConfig

  const cloneDrawerDraft = (value: DrawerDraft): DrawerDraft =>
    JSON.parse(JSON.stringify(value)) as DrawerDraft

  const draft = reactive<FastEnterConfig>(cloneConfig(config.value))
  const persistedConfig = ref<FastEnterConfig>(cloneConfig(config.value))
  const menuRouteTree = ref<RouteTreeOption[]>([])
  const drawerVisible = ref(false)
  const drawerMode = ref<DrawerMode>('application')
  const drawerEditing = ref(false)
  const drawerIndex = ref(-1)
  const drawerDraft = ref<DrawerDraft | null>(null)
  const drawerTargetType = ref<DrawerTargetType>('route')

  const treeSelectProps = {
    label: 'label',
    value: 'value',
    children: 'children'
  }

  const sortedApplications = computed(() =>
    [...draft.applications].sort((a, b) => (a.order || 0) - (b.order || 0))
  )

  const sortedQuickLinks = computed(() =>
    [...draft.quickLinks].sort((a, b) => (a.order || 0) - (b.order || 0))
  )

  const summaryMetrics = computed(() => [
    { label: '当前 App', value: targetAppKey.value },
    { label: '快捷应用', value: draft.applications.length || 0 },
    { label: '快捷链接', value: draft.quickLinks.length || 0 },
    { label: '启用项', value: [...draft.applications, ...draft.quickLinks].filter((item) => item.enabled !== false).length }
  ])

  const loadMenuRouteTree = async () => {
    if (!targetAppKey.value) {
      menuRouteTree.value = []
      return
    }
    const menuTree = await fetchGetMenuTreeAll(undefined, targetAppKey.value)
    menuRouteTree.value = buildRouteTree(menuTree || [])
  }

  const loadAppOptions = async () => {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }

  const handleManagedAppChange = async (value?: string) => {
    await setManagedAppKey(`${value || ''}`.trim())
  }

  const drawerTitle = computed(() =>
    `${drawerEditing.value ? '编辑' : '新增'}${drawerMode.value === 'application' ? '快捷应用' : '快捷链接'}`
  )

  const applicationDrawerDraft = computed(() =>
    drawerMode.value === 'application' && drawerDraft.value ? (drawerDraft.value as DrawerApplicationDraft) : null
  )

  const buildRouteTree = (menus: AppRouteRecord[]): RouteTreeOption[] => {
    const result: RouteTreeOption[] = []

    for (const item of menus) {
      const routeName = typeof item.name === 'string' ? item.name : ''
      const titleSource = item.meta?.title ?? (typeof item.name === 'symbol' ? String(item.name) : item.name) ?? item.path ?? '未命名路由'
      const title = String(titleSource)
      const path = `${item.path || ''}`.trim()
      const children = Array.isArray(item.children) ? buildRouteTree(item.children) : []

      if (routeName && compactMessageWorkspaceRouteNames.has(routeName)) {
        if (children.length) {
          result.push(...children)
        }
        continue
      }

      if (!routeName && children.length === 0) {
        continue
      }

      result.push({
        label: path ? `${title} · ${path}` : title,
        value: routeName,
        path,
        children: children.length ? children : undefined
      })
    }

    return result
  }

  const createApplication = (): DrawerApplicationDraft => ({
    id: `app-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    name: '',
    description: '',
    icon: 'ri:apps-2-line',
    iconColor: '#377dff',
    enabled: true,
    order: draft.applications.length + 1,
    routeName: '',
    link: ''
  })

  const createQuickLink = (): DrawerQuickLinkDraft => ({
    id: `link-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    name: '',
    enabled: true,
    order: draft.quickLinks.length + 1,
    routeName: '',
    link: ''
  })

  const resolveTargetLabel = (item: FastEnterApplication | FastEnterQuickLink) => {
    if (item.routeName) {
      return `内部 · ${item.routeName}`
    }
    if (item.link?.startsWith('/')) {
      return `路径 · ${item.link}`
    }
    if (item.link) {
      return `外链 · ${item.link}`
    }
    return '未配置跳转'
  }

  const restoreDraft = () => {
    Object.assign(draft, cloneConfig(persistedConfig.value))
    fastEnterStore.replaceConfig(cloneConfig(persistedConfig.value))
    ElMessage.success('已恢复到当前已保存配置')
  }

  const saveConfig = async () => {
    await fastEnterStore.saveConfig(cloneConfig(draft))
    persistedConfig.value = cloneConfig(config.value)
    Object.assign(draft, cloneConfig(persistedConfig.value))
    ElMessage.success('快捷入口配置已保存')
  }

  const resetToDefault = async () => {
    await ElMessageBox.confirm('恢复默认后会覆盖当前快捷应用和快捷链接配置，是否继续？', '恢复默认', {
      type: 'warning'
    })
    Object.assign(draft, getDefaultFastEnterConfig())
    await fastEnterStore.saveConfig(cloneConfig(draft))
    persistedConfig.value = cloneConfig(config.value)
    Object.assign(draft, cloneConfig(persistedConfig.value))
    ElMessage.success('已恢复默认配置并同步到后台')
  }

  const syncDraftPreview = () => {
    fastEnterStore.replaceConfig(cloneConfig(draft))
  }

  const openCreateDrawer = (mode: DrawerMode) => {
    drawerMode.value = mode
    drawerEditing.value = false
    drawerIndex.value = -1
    drawerDraft.value = mode === 'application' ? createApplication() : createQuickLink()
    drawerTargetType.value = 'route'
    drawerVisible.value = true
  }

  const openEditDrawer = (mode: DrawerMode, index: number) => {
    drawerMode.value = mode
    drawerEditing.value = true
    drawerIndex.value = index
    const source = mode === 'application' ? sortedApplications.value[index] : sortedQuickLinks.value[index]
    drawerDraft.value = cloneDrawerDraft(source as DrawerDraft)
    drawerTargetType.value = source?.routeName ? 'route' : source?.link?.startsWith('/') ? 'path' : 'link'
    drawerVisible.value = true
  }

  const handleRouteChange = () => {
    if (drawerTargetType.value === 'route' && drawerDraft.value) {
      drawerDraft.value.link = ''
    }
  }

  watch(drawerTargetType, (value) => {
    if (!drawerDraft.value) return
    if (value === 'route') {
      drawerDraft.value.link = ''
    } else {
      drawerDraft.value.routeName = ''
    }
  })

  const applyDrawer = () => {
    if (!drawerDraft.value) return

    const payload = cloneDrawerDraft(drawerDraft.value)
    if (!payload.name.trim()) {
      ElMessage.warning('请先填写名称')
      return
    }

    if (drawerTargetType.value === 'route' && !`${payload.routeName || ''}`.trim()) {
      ElMessage.warning('请选择内部路由')
      return
    }

    if (drawerTargetType.value === 'path') {
      const internalPath = `${payload.link || ''}`.trim()
      if (!internalPath) {
        ElMessage.warning('请填写内部路径')
        return
      }
      if (!internalPath.startsWith('/')) {
        ElMessage.warning('内部路径必须以 / 开头')
        return
      }
    }

    if (drawerTargetType.value === 'link' && !`${payload.link || ''}`.trim()) {
      ElMessage.warning('请填写外部链接')
      return
    }

    if (drawerMode.value === 'application') {
      const items = [...draft.applications]
      if (drawerEditing.value && drawerIndex.value >= 0) {
        const current = sortedApplications.value[drawerIndex.value]
        const targetIndex = items.findIndex((item) => item.id === current?.id)
        if (targetIndex >= 0) {
          items[targetIndex] = payload as FastEnterApplication
        }
      } else {
        items.push(payload as FastEnterApplication)
      }
      draft.applications = items
    } else {
      const items = [...draft.quickLinks]
      if (drawerEditing.value && drawerIndex.value >= 0) {
        const current = sortedQuickLinks.value[drawerIndex.value]
        const targetIndex = items.findIndex((item) => item.id === current?.id)
        if (targetIndex >= 0) {
          items[targetIndex] = payload as FastEnterQuickLink
        }
      } else {
        items.push(payload as FastEnterQuickLink)
      }
      draft.quickLinks = items
    }

    syncDraftPreview()
    drawerVisible.value = false
    ElMessage.success(drawerEditing.value ? '已更新到未保存草稿' : '已加入未保存草稿')
  }

  const removeCurrentItem = async () => {
    if (!drawerDraft.value) return
    await ElMessageBox.confirm('删除后会从当前配置草稿中移除该入口，是否继续？', '删除入口', {
      type: 'warning'
    })

    if (drawerMode.value === 'application') {
      draft.applications = draft.applications.filter((item) => item.id !== drawerDraft.value?.id)
    } else {
      draft.quickLinks = draft.quickLinks.filter((item) => item.id !== drawerDraft.value?.id)
    }
    syncDraftPreview()
    drawerVisible.value = false
    ElMessage.success('已从未保存草稿中移除')
  }

  onMounted(async () => {
    selectedAppKey.value = targetAppKey.value
    await loadAppOptions().catch(() => {
      appList.value = []
    })
    await fastEnterStore.loadConfig(true)
    Object.assign(draft, cloneConfig(config.value))
    persistedConfig.value = cloneConfig(config.value)
    await loadMenuRouteTree()
  })

  onUnmounted(() => {
    fastEnterStore.replaceConfig(cloneConfig(persistedConfig.value))
  })

  watch(
    targetAppKey,
    () => {
      selectedAppKey.value = targetAppKey.value || ''
      loadMenuRouteTree().catch(() => {
        ElMessage.error('刷新当前 App 菜单树失败')
      })
    }
  )
</script>

<style scoped lang="scss">
  .fast-enter-page {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .fast-enter-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .fast-enter-app-select {
    width: 240px;
  }

  .fast-enter-shell {
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 18px;
  }

  .fast-enter-shell__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
  }

  .fast-enter-shell__divider {
    display: none;
  }

  .fast-enter-shell__toolbar-main {
    display: grid;
    gap: 6px;
  }

  .fast-enter-shell__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .fast-enter-shell__note {
    margin: 0;
    font-size: 12px;
    line-height: 1.65;
    color: #64748b;
  }

  .fast-enter-board {
    display: grid;
    grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.84fr);
    gap: 18px;
    min-height: 0;
    margin-top: 20px;
  }

  .fast-enter-panel {
    display: flex;
    flex-direction: column;
    gap: 0;
    min-height: 0;
    padding: 14px;
  }

  .fast-enter-panel__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 14px;
    padding-bottom: 0;
  }

  .fast-enter-panel__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .fast-enter-panel__desc {
    margin: 4px 0 0;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .fast-enter-preview {
    display: grid;
    gap: 14px;
    min-height: 0;
    margin-top: 20px;
  }

  .fast-enter-preview--apps {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .fast-enter-preview-card,
  .fast-enter-link-row {
    width: 100%;
    border: 1px solid rgb(226 232 240 / 0.82);
    border-radius: 16px;
    background: rgb(255 255 255 / 0.96);
    text-align: left;
    transition:
      transform 0.18s ease,
      border-color 0.18s ease,
      box-shadow 0.18s ease;
    cursor: pointer;
  }

  .fast-enter-preview-card:hover,
  .fast-enter-link-row:hover {
    border-color: rgb(59 130 246 / 0.28);
    box-shadow: 0 16px 30px rgb(15 23 42 / 0.08);
    transform: translateY(-1px);
  }

  .fast-enter-preview-card {
    display: grid;
    grid-template-columns: 52px minmax(0, 1fr);
    gap: 12px;
    padding: 13px 14px;
  }

  .fast-enter-preview-card__icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    border-radius: 16px;
    background: rgb(241 245 249 / 0.92);
    font-size: 20px;
  }

  .fast-enter-preview-card__content {
    display: grid;
    gap: 7px;
    min-width: 0;
  }

  .fast-enter-preview-card__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .fast-enter-preview-card__name,
  .fast-enter-link-row__name {
    min-width: 0;
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
  }

  .fast-enter-preview-card__order {
    flex: none;
    font-size: 11px;
    color: #94a3b8;
  }

  .fast-enter-preview-card__desc {
    margin: 0;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .fast-enter-preview-card__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .fast-enter-badge {
    display: inline-flex;
    align-items: center;
    height: 24px;
    padding: 0 10px;
    border-radius: 999px;
    background: rgb(15 118 110 / 0.12);
    color: #0f766e;
    font-size: 11px;
    font-weight: 600;
  }

  .fast-enter-badge.is-muted {
    background: rgb(148 163 184 / 0.14);
    color: #64748b;
  }

  .fast-enter-badge--soft {
    background: rgb(59 130 246 / 0.08);
    color: #2563eb;
  }

  .fast-enter-link-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 15px;
  }

  .fast-enter-link-row__main {
    display: grid;
    gap: 6px;
    min-width: 0;
  }

  .fast-enter-link-row__target {
    font-size: 12px;
    color: #64748b;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .fast-enter-link-row__side {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: #94a3b8;
    font-size: 12px;
  }

  .fast-enter-link-row__status {
    color: #475569;
  }

  .fast-enter-drawer__summary {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 14px 16px;
  }

  .fast-enter-drawer__summary-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 48px;
    border-radius: 16px;
    background: rgb(241 245 249 / 0.92);
    font-size: 20px;
  }

  .fast-enter-drawer__summary-body {
    display: grid;
    gap: 4px;
  }

  .fast-enter-drawer__summary-title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .fast-enter-drawer__summary-text {
    font-size: 12px;
    color: #64748b;
  }

  .fast-enter-drawer__form {
    display: grid;
    gap: 14px;
    margin-top: 16px;
  }

  .fast-enter-field-hint {
    display: grid;
    gap: 4px;
    margin-top: -6px;
    padding: 0 2px;
    color: #64748b;
    font-size: 12px;
    line-height: 1.6;
  }

  .fast-enter-drawer__hint {
    display: grid;
    gap: 4px;
    padding: 12px 14px;
    border-radius: 14px;
    background: rgb(248 250 252 / 0.96);
    color: #64748b;
    font-size: 12px;
    line-height: 1.6;
  }

  .fast-enter-drawer__footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
  }

  .fast-enter-drawer__footer-actions {
    display: flex;
    gap: 12px;
    margin-left: auto;
  }

  :deep(.fast-enter-drawer .el-drawer__body) {
    padding-top: 8px;
  }

  :deep(.fast-enter-drawer .el-form-item) {
    margin-bottom: 0;
  }

  :deep(.fast-enter-drawer .el-form-item__label) {
    font-size: 12px;
    color: #64748b;
  }

  @media (max-width: 1180px) {
    .fast-enter-board,
    .fast-enter-preview--apps {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .fast-enter-shell__toolbar,
    .fast-enter-panel__header,
    .fast-enter-drawer__footer {
      flex-direction: column;
      align-items: stretch;
    }

    .fast-enter-drawer__footer-actions {
      width: 100%;
    }
  }
</style>

