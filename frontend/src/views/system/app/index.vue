<template>
  <div class="app-manage-page art-full-height" v-loading="loading">
    <AdminWorkspaceHero
      title="应用管理"
      description="以 App 为资源边界统一收口站点默认空间、Host 绑定和导航管理入口；菜单空间降级为当前 App 下的高级配置。"
      :metrics="summaryMetrics"
    >
      <div class="app-manage-hero-actions">
        <ElButton type="primary" @click="openAppDrawer(currentAppRecord || selectedAppRecord)" v-ripple>
          编辑当前 App
        </ElButton>
        <ElButton :disabled="!selectedAppKey" @click="openHostDrawer()" v-ripple>
          新增 Host 绑定
        </ElButton>
        <ElButton :disabled="!selectedAppKey" @click="goToMenuManagement" v-ripple>
          菜单管理
        </ElButton>
        <ElButton :disabled="!selectedAppKey" @click="goToPageManagement" v-ripple>
          页面管理
        </ElButton>
        <ElButton :disabled="!selectedAppKey" @click="goToSpaceManagement()" v-ripple>
          高级空间配置
        </ElButton>
      </div>
    </AdminWorkspaceHero>

    <ElAlert
      v-if="loadError"
      class="app-manage-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="app-manage-board">
      <ElCard class="app-manage-panel" shadow="never">
        <template #header>
          <div class="app-manage-panel__header">
            <div>
              <div class="app-manage-panel__title">App 列表</div>
              <div class="app-manage-panel__desc">默认内置 App 为平台管理后台；后续多前端产品继续沿用同一套用户、权限和导航体系。</div>
            </div>
          </div>
        </template>

        <div class="app-manage-list">
          <button
            v-for="item in apps"
            :key="item.appKey"
            type="button"
            class="app-manage-item"
            :class="{ 'is-current': selectedAppKey === item.appKey }"
            @click="selectApp(item.appKey)"
          >
            <div class="app-manage-item__main">
              <div class="app-manage-item__title-row">
                <span class="app-manage-item__title">{{ item.name }}</span>
                <ElTag v-if="item.isDefault" size="small" type="success" effect="plain">默认</ElTag>
                <ElTag
                  v-if="currentAppRecord?.appKey === item.appKey"
                  size="small"
                  type="warning"
                  effect="plain"
                >
                  当前解析
                </ElTag>
                <ElTag
                  size="small"
                  :type="item.status === 'normal' ? 'info' : 'danger'"
                  effect="plain"
                >
                  {{ item.status === 'normal' ? '启用' : '停用' }}
                </ElTag>
              </div>
              <div class="app-manage-item__meta">
                <span>标识 {{ item.appKey }}</span>
                <span>默认空间 {{ item.defaultSpaceKey || 'default' }}</span>
                <span>空间 {{ item.menuSpaceCount || 0 }}</span>
                <span>菜单 {{ item.menuCount || 0 }}</span>
                <span>页面 {{ item.pageCount || 0 }}</span>
                <span>Host {{ item.hostCount || 0 }}</span>
              </div>
              <p class="app-manage-item__desc">
                {{ item.description || '当前 App 未填写说明，建议补充站点职责、登录策略或业务边界。' }}
              </p>
            </div>
            <div class="app-manage-item__actions">
              <ElButton text type="primary" @click.stop="openAppDrawer(item)">编辑</ElButton>
              <ElButton text @click.stop="goToSpaceManagement(item.appKey)">空间配置</ElButton>
            </div>
          </button>
        </div>
      </ElCard>

      <ElCard class="app-manage-panel" shadow="never">
        <template #header>
          <div class="app-manage-panel__header">
            <div>
              <div class="app-manage-panel__title">当前 App 概览</div>
              <div class="app-manage-panel__desc">Host 命中和默认空间都会影响运行时导航编译结果，下面只展示当前选中 App 的有效配置。</div>
            </div>
            <div class="app-manage-panel__status">
              <ElTag effect="plain" type="info">当前查看 {{ selectedAppRecord?.name || selectedAppKey || '-' }}</ElTag>
              <ElTag v-if="currentAppRecord" effect="plain" type="warning">
                解析来源 {{ currentAppResolvedLabel }}
              </ElTag>
            </div>
          </div>
        </template>

        <div v-if="selectedAppRecord" class="app-overview">
          <div class="app-overview__grid">
            <div class="app-overview__item">
              <span class="app-overview__label">App 标识</span>
              <strong>{{ selectedAppRecord.appKey }}</strong>
            </div>
            <div class="app-overview__item">
              <span class="app-overview__label">默认空间</span>
              <strong>{{ selectedAppRecord.defaultSpaceKey || 'default' }}</strong>
            </div>
            <div class="app-overview__item">
              <span class="app-overview__label">主 Host</span>
              <strong>{{ selectedAppRecord.primaryHost || '未设置' }}</strong>
            </div>
            <div class="app-overview__item">
              <span class="app-overview__label">空间 / 菜单 / 页面</span>
              <strong>{{ selectedAppRecord.menuSpaceCount || 0 }} / {{ selectedAppRecord.menuCount || 0 }} / {{ selectedAppRecord.pageCount || 0 }}</strong>
            </div>
            <div class="app-overview__item">
              <span class="app-overview__label">请求 Host</span>
              <strong>{{ currentAppRequestHost || '未命中' }}</strong>
            </div>
            <div class="app-overview__item">
              <span class="app-overview__label">Host 绑定</span>
              <strong>{{ hostBindings.length }}</strong>
            </div>
          </div>
          <div class="app-overview__actions">
            <ElButton text @click="openHostDrawer()">新增 Host 绑定</ElButton>
            <ElButton text @click="goToMenuManagement">进入菜单管理</ElButton>
            <ElButton text @click="goToPageManagement">进入页面管理</ElButton>
            <ElButton text @click="goToSpaceManagement()">进入高级空间配置</ElButton>
          </div>
        </div>

        <div class="app-binding-list">
          <div v-if="!hostBindings.length" class="app-manage-empty">
            当前 App 还没有 Host 绑定。未命中 Host 时，系统会退回 App 默认空间。
          </div>
          <button
            v-for="item in hostBindings"
            :key="item.host"
            type="button"
            class="app-binding-item"
            @click="openHostDrawer(item)"
          >
            <div class="app-binding-item__main">
              <div class="app-binding-item__title-row">
                <span class="app-binding-item__host">{{ item.host }}</span>
                <ElTag v-if="item.isPrimary" size="small" type="success" effect="plain">主绑定</ElTag>
                <ElTag
                  size="small"
                  :type="item.status === 'normal' ? 'info' : 'danger'"
                  effect="plain"
                >
                  {{ item.status === 'normal' ? '启用' : '停用' }}
                </ElTag>
              </div>
              <div class="app-binding-item__meta">
                <span>默认空间 {{ item.defaultSpaceKey || selectedAppRecord?.defaultSpaceKey || 'default' }}</span>
                <span v-if="item.description">{{ item.description }}</span>
              </div>
            </div>
            <ArtSvgIcon icon="ri:arrow-right-s-line" />
          </button>
        </div>

        <div class="app-space-pills">
          <span class="app-space-pills__label">空间配置</span>
          <span
            v-for="item in spaces"
            :key="item.spaceKey"
            class="app-space-pill"
          >
            {{ item.name }} · {{ item.spaceKey }}
          </span>
          <span v-if="!spaces.length" class="app-space-pill is-soft">当前 App 暂无空间配置</span>
        </div>
      </ElCard>
    </section>

    <ElDrawer v-model="appDrawerVisible" :title="appDrawerTitle" size="520px" destroy-on-close>
      <ElForm :model="appForm" label-position="top">
        <ElFormItem label="应用名称">
          <ElInput v-model="appForm.name" placeholder="例如 平台管理后台" />
        </ElFormItem>
        <ElFormItem label="应用标识">
          <ElInput v-model="appForm.app_key" :disabled="Boolean(editingAppKey)" placeholder="例如 platform-admin" />
        </ElFormItem>
        <ElFormItem label="默认空间">
          <ElSelect v-model="appForm.default_space_key" filterable allow-create default-first-option style="width: 100%">
            <ElOption
              v-for="item in spaces"
              :key="item.spaceKey"
              :label="`${item.name} · ${item.spaceKey}`"
              :value="item.spaceKey"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput v-model="appForm.description" type="textarea" :rows="3" placeholder="说明这个 App 面向哪个站点或后台产品" />
        </ElFormItem>
        <div class="app-drawer-grid">
          <ElFormItem label="默认 App">
            <ElSwitch v-model="appForm.is_default" />
          </ElFormItem>
          <ElFormItem label="状态">
            <ElSelect v-model="appForm.status">
              <ElOption label="启用" value="normal" />
              <ElOption label="停用" value="disabled" />
            </ElSelect>
          </ElFormItem>
        </div>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="appDrawerVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingApp" @click="saveApp">保存</ElButton>
        </div>
      </template>
    </ElDrawer>

    <ElDrawer v-model="hostDrawerVisible" :title="hostDrawerTitle" size="520px" destroy-on-close>
      <ElForm :model="hostForm" label-position="top">
        <ElFormItem label="Host / 子域名">
          <ElInput v-model="hostForm.host" placeholder="例如 admin.example.com" />
        </ElFormItem>
        <ElFormItem label="默认空间">
          <ElSelect v-model="hostForm.default_space_key" filterable allow-create default-first-option style="width: 100%">
            <ElOption
              v-for="item in spaces"
              :key="item.spaceKey"
              :label="`${item.name} · ${item.spaceKey}`"
              :value="item.spaceKey"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput v-model="hostForm.description" type="textarea" :rows="3" placeholder="例如 平台治理入口 / 商家后台入口" />
        </ElFormItem>
        <div class="app-drawer-grid">
          <ElFormItem label="主绑定">
            <ElSwitch v-model="hostForm.is_primary" />
          </ElFormItem>
          <ElFormItem label="状态">
            <ElSelect v-model="hostForm.status">
              <ElOption label="启用" value="normal" />
              <ElOption label="停用" value="disabled" />
            </ElSelect>
          </ElFormItem>
        </div>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="hostDrawerVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingHost" @click="saveHostBinding">保存</ElButton>
        </div>
      </template>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import { useAppContextStore } from '@/store/modules/app-context'
  import {
    fetchGetApps,
    fetchGetAppHostBindings,
    fetchGetCurrentApp,
    fetchGetMenuSpaces,
    fetchSaveApp,
    fetchSaveAppHostBinding
  } from '@/api/system-manage'

  defineOptions({ name: 'AppManage' })

  const router = useRouter()
  const appContextStore = useAppContextStore()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const loading = ref(false)
  const loadError = ref('')
  const savingApp = ref(false)
  const savingHost = ref(false)
  const apps = ref<Api.SystemManage.AppItem[]>([])
  const hostBindings = ref<Api.SystemManage.AppHostBindingItem[]>([])
  const spaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const currentApp = ref<Api.SystemManage.CurrentAppResponse>()
  const selectedAppKey = ref('')

  const appDrawerVisible = ref(false)
  const hostDrawerVisible = ref(false)
  const editingAppKey = ref('')
  const editingHost = ref('')

  const appForm = reactive<Api.SystemManage.AppSaveParams>({
    app_key: '',
    name: '',
    description: '',
    default_space_key: 'default',
    is_default: false,
    status: 'normal',
    meta: {}
  })

  const hostForm = reactive<Api.SystemManage.AppHostBindingSaveParams>({
    app_key: '',
    host: '',
    default_space_key: 'default',
    description: '',
    is_primary: false,
    status: 'normal',
    meta: {}
  })

  const currentAppRecord = computed(() => currentApp.value?.app)
  const currentAppRequestHost = computed(() => `${currentApp.value?.requestHost || ''}`.trim())
  const selectedAppRecord = computed(
    () => apps.value.find((item) => item.appKey === selectedAppKey.value) || currentAppRecord.value
  )
  const currentAppResolvedLabel = computed(() => {
    switch (`${currentApp.value?.resolvedBy || ''}`.trim()) {
      case 'host_binding':
        return 'Host 绑定'
      case 'legacy_space_host_binding':
        return '旧空间 Host 绑定'
      case 'explicit':
        return '显式指定'
      case 'default_app':
        return '默认 App'
      default:
        return `${currentApp.value?.resolvedBy || '默认 App'}`
    }
  })
  const appDrawerTitle = computed(() => (editingAppKey.value ? '编辑应用' : '新增应用'))
  const hostDrawerTitle = computed(() => (editingHost.value ? '编辑 Host 绑定' : '新增 Host 绑定'))
  const summaryMetrics = computed(() => [
    { label: '应用数', value: apps.value.length || 0 },
    { label: '当前 App', value: currentAppRecord.value?.appKey || selectedAppKey.value || '-' },
    { label: '空间数', value: selectedAppRecord.value?.menuSpaceCount || 0 },
    { label: '菜单数', value: selectedAppRecord.value?.menuCount || 0 },
    { label: '页面数', value: selectedAppRecord.value?.pageCount || 0 },
    { label: 'Host 绑定', value: hostBindings.value.length || 0 }
  ])

  function resolveAppKey(...candidates: Array<string | undefined | null>) {
    for (const candidate of candidates) {
      const normalized = `${candidate || ''}`.trim()
      if (normalized) {
        return normalized
      }
    }
    return ''
  }

  async function loadSelectedAppContext(appKey: string) {
    const normalizedAppKey = resolveAppKey(appKey)
    if (!normalizedAppKey) {
      throw new Error('缺少 app 上下文')
    }
    selectedAppKey.value = normalizedAppKey
    appContextStore.setManagedAppKey(normalizedAppKey)
    await setManagedAppKey(normalizedAppKey)
    const [hostRes, spaceRes] = await Promise.all([
      fetchGetAppHostBindings(normalizedAppKey),
      fetchGetMenuSpaces(normalizedAppKey)
    ])
    hostBindings.value = hostRes.records || []
    spaces.value = spaceRes.records || []
  }

  async function loadData() {
    loading.value = true
    loadError.value = ''
    try {
      const [appsRes, currentRes] = await Promise.all([fetchGetApps(), fetchGetCurrentApp()])
      apps.value = appsRes.records || []
      currentApp.value = currentRes
      const nextAppKey = resolveAppKey(targetAppKey.value, selectedAppKey.value, currentRes?.app?.appKey, apps.value[0]?.appKey)
      if (!nextAppKey) {
        throw new Error('未找到可管理的 App')
      }
      await loadSelectedAppContext(nextAppKey)
    } catch (error: any) {
      apps.value = []
      hostBindings.value = []
      spaces.value = []
      loadError.value = error?.message || '应用数据暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  function resetAppForm() {
    editingAppKey.value = ''
    appForm.app_key = resolveAppKey(selectedAppKey.value, currentAppRecord.value?.appKey)
    appForm.name = ''
    appForm.description = ''
    appForm.default_space_key = selectedAppRecord.value?.defaultSpaceKey || spaces.value[0]?.spaceKey || 'default'
    appForm.is_default = false
    appForm.status = 'normal'
    appForm.meta = {}
  }

  function resetHostForm() {
    editingHost.value = ''
    hostForm.app_key = resolveAppKey(selectedAppKey.value, currentAppRecord.value?.appKey)
    hostForm.host = ''
    hostForm.default_space_key = selectedAppRecord.value?.defaultSpaceKey || spaces.value[0]?.spaceKey || 'default'
    hostForm.description = ''
    hostForm.is_primary = false
    hostForm.status = 'normal'
    hostForm.meta = {}
  }

  function openAppDrawer(item?: Api.SystemManage.AppItem) {
    resetAppForm()
    if (item) {
      editingAppKey.value = item.appKey
      appForm.app_key = item.appKey
      appForm.name = item.name
      appForm.description = item.description || ''
      appForm.default_space_key = item.defaultSpaceKey || 'default'
      appForm.is_default = Boolean(item.isDefault)
      appForm.status = item.status || 'normal'
      appForm.meta = item.meta || {}
    }
    appDrawerVisible.value = true
  }

  function openHostDrawer(item?: Api.SystemManage.AppHostBindingItem) {
    resetHostForm()
    if (item) {
      editingHost.value = item.host
      hostForm.app_key = item.appKey || selectedAppKey.value
      hostForm.host = item.host
      hostForm.default_space_key = item.defaultSpaceKey || selectedAppRecord.value?.defaultSpaceKey || 'default'
      hostForm.description = item.description || ''
      hostForm.is_primary = Boolean(item.isPrimary)
      hostForm.status = item.status || 'normal'
      hostForm.meta = item.meta || {}
    }
    hostDrawerVisible.value = true
  }

  async function saveApp() {
    if (!appForm.app_key.trim()) {
      ElMessage.warning('请输入应用标识')
      return
    }
    if (!appForm.name.trim()) {
      ElMessage.warning('请输入应用名称')
      return
    }
    savingApp.value = true
    try {
      const saved = await fetchSaveApp({
        ...appForm,
        app_key: appForm.app_key.trim(),
        name: appForm.name.trim(),
        description: appForm.description?.trim() || '',
        default_space_key: `${appForm.default_space_key || 'default'}`.trim() || 'default'
      })
      ElMessage.success('应用已保存')
      appDrawerVisible.value = false
      await setManagedAppKey(saved.appKey)
      selectedAppKey.value = saved.appKey
      await loadData()
    } catch (error: any) {
      ElMessage.error(error?.message || '应用保存失败')
    } finally {
      savingApp.value = false
    }
  }

  async function saveHostBinding() {
    if (!selectedAppKey.value) {
      ElMessage.warning('请先选择应用')
      return
    }
    if (!hostForm.host.trim()) {
      ElMessage.warning('请输入 Host')
      return
    }
    savingHost.value = true
    try {
      await fetchSaveAppHostBinding({
        ...hostForm,
        app_key: selectedAppKey.value,
        host: hostForm.host.trim().toLowerCase(),
        default_space_key: `${hostForm.default_space_key || selectedAppRecord.value?.defaultSpaceKey || 'default'}`.trim() || 'default',
        description: hostForm.description?.trim() || ''
      })
      ElMessage.success('Host 绑定已保存')
      hostDrawerVisible.value = false
      await loadSelectedAppContext(selectedAppKey.value)
      await loadData()
    } catch (error: any) {
      ElMessage.error(error?.message || 'Host 绑定保存失败')
    } finally {
      savingHost.value = false
    }
  }

  function selectApp(appKey: string) {
    if (!appKey || appKey === selectedAppKey.value) return
    loadSelectedAppContext(appKey).catch((error: any) => {
      ElMessage.error(error?.message || '切换应用失败')
    })
  }

  function goToMenuManagement() {
    router.push({ path: '/system/menu', query: { app_key: selectedAppKey.value } })
  }

  function goToPageManagement() {
    router.push({ path: '/system/page', query: { app_key: selectedAppKey.value } })
  }

  function goToSpaceManagement(appKey?: string) {
    router.push({ path: '/system/menu-space', query: { app_key: appKey || selectedAppKey.value } })
  }

  onMounted(() => {
    loadData()
  })

  watch(
    targetAppKey,
    (value) => {
      if (value && value !== selectedAppKey.value) {
        selectedAppKey.value = value
      }
    }
  )
</script>

<style scoped lang="scss">
  .app-manage-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .app-manage-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .app-manage-inline-alert {
    margin-top: -4px;
  }

  .app-manage-board {
    display: grid;
    grid-template-columns: minmax(360px, 1.1fr) minmax(420px, 1fr);
    gap: 16px;
    min-height: 0;
  }

  .app-manage-panel {
    min-height: 0;
  }

  .app-manage-panel__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .app-manage-panel__title {
    font-size: 16px;
    font-weight: 700;
    color: var(--art-text-gray-900);
  }

  .app-manage-panel__desc {
    margin-top: 6px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--art-text-gray-500);
  }

  .app-manage-panel__status {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .app-manage-list,
  .app-binding-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .app-manage-item,
  .app-binding-item {
    display: flex;
    width: 100%;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    border: 1px solid var(--art-border-color);
    border-radius: 16px;
    background: var(--art-main-bg-color);
    padding: 16px;
    text-align: left;
    transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
  }

  .app-manage-item:hover,
  .app-binding-item:hover,
  .app-manage-item.is-current {
    border-color: var(--art-primary);
    box-shadow: 0 12px 24px rgba(19, 45, 95, 0.08);
  }

  .app-manage-item__main,
  .app-binding-item__main {
    flex: 1 1 auto;
    min-width: 0;
  }

  .app-manage-item__title-row,
  .app-binding-item__title-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .app-manage-item__title,
  .app-binding-item__host {
    font-size: 16px;
    font-weight: 700;
    color: var(--art-text-gray-900);
  }

  .app-manage-item__meta,
  .app-binding-item__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 8px;
    font-size: 13px;
    color: var(--art-text-gray-500);
  }

  .app-manage-item__desc {
    margin-top: 10px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--art-text-gray-600);
  }

  .app-manage-item__actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 6px;
    flex: 0 0 auto;
  }

  .app-overview {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-bottom: 16px;
    padding: 16px;
    border-radius: 16px;
    background: linear-gradient(180deg, rgba(72, 120, 255, 0.06), rgba(72, 120, 255, 0.02));
  }

  .app-overview__grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
  }

  .app-overview__item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 12px 14px;
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.84);
  }

  .app-overview__label {
    font-size: 12px;
    color: var(--art-text-gray-500);
  }

  .app-overview__actions,
  .app-space-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
  }

  .app-space-pills__label {
    font-size: 13px;
    font-weight: 600;
    color: var(--art-text-gray-700);
  }

  .app-space-pill {
    padding: 6px 10px;
    border-radius: 999px;
    background: var(--art-gray-100);
    color: var(--art-text-gray-600);
    font-size: 12px;
  }

  .app-space-pill.is-soft,
  .app-manage-empty {
    color: var(--art-text-gray-500);
  }

  .app-drawer-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  @media (max-width: 1200px) {
    .app-manage-board {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .app-overview__grid,
    .app-drawer-grid {
      grid-template-columns: 1fr;
    }

    .app-manage-item,
    .app-binding-item {
      flex-direction: column;
      align-items: stretch;
    }

    .app-manage-item__actions {
      flex-direction: row;
      justify-content: flex-end;
    }
  }
</style>
