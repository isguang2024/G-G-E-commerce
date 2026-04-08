<template>
  <div class="menu-space-page art-full-height" v-loading="loading">
    <AdminWorkspaceHero
      title="空间布局高级配置"
      description="菜单定义与空间布局已经分层：这里负责空间列表、默认空间、Host 绑定与布局树入口，不再承担菜单定义本体维护。"
      :metrics="summaryMetrics"
    >
      <div class="menu-space-hero-actions">
        <ElSelect
          v-model="selectedAppKey"
          clearable
          filterable
          placeholder="选择 App"
          class="menu-space-app-select"
          @change="handleManagedAppChange"
        >
          <ElOption
            v-for="item in appOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
        <ElSelect v-model="spaceMode" class="menu-space-mode-select">
          <ElOption label="单空间模式" value="single" />
          <ElOption label="多空间模式" value="multi" />
        </ElSelect>
        <ElButton :loading="savingSpaceMode" @click="saveSpaceMode" v-ripple>保存模式</ElButton>
        <ElButton type="primary" @click="openSpaceDrawer()" v-ripple>新增空间</ElButton>
        <ElButton @click="openHostDrawer()" v-ripple>新增 Host 绑定</ElButton>
        <ElButton @click="loadData" v-ripple>刷新</ElButton>
      </div>
    </AdminWorkspaceHero>

    <ElAlert
      v-if="loadError"
      class="menu-space-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="menu-space-board">
      <ElCard class="menu-space-panel" shadow="never">
        <template #header>
          <div class="menu-space-panel__header">
            <div>
              <div class="menu-space-panel__title">菜单空间</div>
              <div class="menu-space-panel__desc"
                >这里负责当前 App 的空间列表、Host
                绑定、默认空间与布局入口，菜单定义本体不在此维护。</div
              >
            </div>
          </div>
        </template>

        <div class="menu-space-list">
          <button
            v-for="item in spaces"
            :key="item.spaceKey"
            type="button"
            class="menu-space-item"
            :class="{ 'is-current': currentSpaceKey === item.spaceKey }"
            @click="currentSpaceKey = item.spaceKey"
          >
            <div class="menu-space-item__main">
              <div class="menu-space-item__title-row">
                <span class="menu-space-item__title">{{ item.name }}</span>
                <ElTag v-if="item.isDefault" size="small" type="success" effect="plain">默认</ElTag>
                <ElTag
                  size="small"
                  :type="item.status === 'normal' ? 'info' : 'danger'"
                  effect="plain"
                >
                  {{ item.status === 'normal' ? '启用' : '停用' }}
                </ElTag>
              </div>
              <div class="menu-space-item__meta">
                <span>标识 {{ item.spaceKey }}</span>
                <span>首页 {{ item.defaultHomePath || '-' }}</span>
                <span>菜单 {{ item.menuCount || 0 }}</span>
                <span>独立页暴露 {{ item.pageCount || 0 }}</span>
                <span>空间权限 {{ getAccessModeLabel(item.accessMode) }}</span>
                <span>Host {{ item.hostCount || 0 }}</span>
              </div>
              <p class="menu-space-item__desc">
                {{ item.description || '当前空间未填写描述，建议补充业务边界或使用说明。' }}
              </p>
              <div class="menu-space-item__hosts">
                <span class="menu-space-chip" :class="{ 'is-soft': !isSpaceInitialized(item) }">
                  {{ isSpaceInitialized(item) ? '已初始化' : '待初始化' }}
                </span>
                <span v-for="host in item.hosts?.slice(0, 3)" :key="host" class="menu-space-chip">
                  {{ host }}
                </span>
                <span v-if="(item.hosts?.length || 0) > 3" class="menu-space-chip is-soft">
                  +{{ (item.hosts?.length || 0) - 3 }}
                </span>
              </div>
            </div>
            <div class="menu-space-item__actions">
              <ElButton text type="primary" @click.stop="openSpaceDrawer(item)">编辑</ElButton>
              <ElButton text @click.stop="goToMenuManagement(item.spaceKey)">空间布局</ElButton>
              <ElButton text @click.stop="goToPageManagement(item.spaceKey)">受管页面</ElButton>
              <ElButton
                v-if="!item.isDefault && !isSpaceInitialized(item)"
                text
                :loading="initializingSpaceKey === item.spaceKey"
                @click.stop="initializeSpace(item)"
              >
                初始化菜单树
              </ElButton>
              <ElButton v-else-if="!item.isDefault" text disabled> 已初始化 </ElButton>
              <ElButton text @click.stop="openHostDrawer(undefined, item.spaceKey)"
                >绑定 Host</ElButton
              >
            </div>
          </button>
        </div>
      </ElCard>

      <ElCard class="menu-space-panel" shadow="never">
        <template #header>
          <div class="menu-space-panel__header">
            <div>
              <div class="menu-space-panel__title">Host 绑定</div>
              <div class="menu-space-panel__desc"
                >可选配置。先解析 App，再决定空间绑定；未命中 Host
                时，不再额外兜底到固定默认空间。</div
              >
            </div>
            <div class="menu-space-panel__status">
              <ElTag effect="plain" type="info">当前解析 {{ currentSpaceLabel }}</ElTag>
              <ElTag effect="plain" :type="spaceModeTagType">模式 {{ spaceModeLabel }}</ElTag>
              <ElTag v-if="resolveByLabel" effect="plain" type="warning"
                >来源 {{ resolveByLabel }}</ElTag
              >
            </div>
          </div>
        </template>

        <div v-if="currentSpace" class="menu-space-overview">
          <div class="menu-space-overview__grid">
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">当前空间</span>
              <strong>{{ currentSpace.name }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">默认首页</span>
              <strong>{{ currentSpace.defaultHomePath || '未设置' }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">菜单 / 独立页暴露</span>
              <strong>{{ currentSpace.menuCount || 0 }} / {{ currentSpace.pageCount || 0 }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">初始化状态</span>
              <strong>{{ isSpaceInitialized(currentSpace) ? '已初始化' : '待初始化' }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">空间准入</span>
              <strong>{{ getAccessModeSummary(currentSpace) }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">解析来源</span>
              <strong>{{ resolveByLabel || '未返回' }}</strong>
            </div>
            <div class="menu-space-overview__item">
              <span class="menu-space-overview__label">请求 Host</span>
              <strong>{{ currentRequestHost || '未命中' }}</strong>
            </div>
          </div>
          <div class="menu-space-overview__actions">
            <ElButton text @click="goToMenuManagement(currentSpace.spaceKey)"
              >编辑当前空间布局</ElButton
            >
            <ElButton text @click="goToPageManagement(currentSpace.spaceKey)"
              >进入受管页面</ElButton
            >
            <ElButton
              v-if="!currentSpace.isDefault && isSpaceInitialized(currentSpace)"
              text
              type="danger"
              :loading="initializingSpaceKey === currentSpace.spaceKey"
              @click="reinitializeSpace(currentSpace)"
            >
              重新初始化
            </ElButton>
          </div>
        </div>

        <div class="menu-space-binding-list">
          <div v-if="!hostBindings.length" class="menu-space-empty">
            还没有 Host 绑定。当前依然按默认菜单空间运行，不影响现有架构。
          </div>
          <button
            v-for="item in hostBindings"
            :key="item.host"
            type="button"
            class="menu-space-binding"
            @click="openHostDrawer(item)"
          >
            <div class="menu-space-binding__main">
              <div class="menu-space-binding__title-row">
                <span class="menu-space-binding__host">{{ item.host }}</span>
                <ElTag
                  size="small"
                  :type="item.status === 'normal' ? 'success' : 'danger'"
                  effect="plain"
                >
                  {{ item.status === 'normal' ? '启用' : '停用' }}
                </ElTag>
              </div>
              <div class="menu-space-binding__meta">
                <span>菜单空间 {{ item.spaceName || item.spaceKey }}</span>
                <span v-if="item.isDefault">主绑定</span>
                <span>{{ getHostAuthModeLabel(item.authMode) }}</span>
                <span v-if="item.routePrefix">前缀 {{ item.routePrefix }}</span>
                <span v-if="item.description">{{ item.description }}</span>
              </div>
            </div>
            <ArtSvgIcon icon="ri:arrow-right-s-line" />
          </button>
        </div>
      </ElCard>
    </section>

    <ElDrawer v-model="spaceDrawerVisible" :title="spaceDrawerTitle" size="520px" destroy-on-close>
      <ElForm ref="spaceFormRef" :model="spaceForm" label-position="top">
        <ElFormItem label="空间名称">
          <ElInput v-model="spaceForm.name" placeholder="例如 默认菜单空间 / 平台运营空间" />
        </ElFormItem>
        <ElFormItem label="空间标识">
          <ElInput
            v-model="spaceForm.space_key"
            :disabled="spaceForm.is_default"
            placeholder="例如 default / personal / crm"
          />
        </ElFormItem>
        <ElFormItem label="默认首页">
          <ElSelect
            v-model="spaceForm.default_home_path"
            filterable
            allow-create
            clearable
            default-first-option
            style="width: 100%"
            placeholder="请选择或输入空间默认落地页"
          >
            <ElOption v-for="item in landingPathOptions" :key="item" :label="item" :value="item" />
          </ElSelect>
          <div class="field-hint">
            {{ landingPathHint }}
          </div>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput
            v-model="spaceForm.description"
            type="textarea"
            :rows="3"
            placeholder="说明这个菜单空间承载什么菜单树、默认入口与 Host 边界"
          />
        </ElFormItem>
        <ElFormItem label="空间权限">
          <ElSelect v-model="spaceForm.access_mode" style="width: 100%">
            <ElOption label="全部可进" value="all" />
            <ElOption label="仅个人空间管理员" value="personal_workspace_admin" />
            <ElOption label="仅协作空间管理员" value="collaboration_workspace_admin" />
            <ElOption label="指定空间角色码" value="role_codes" />
          </ElSelect>
          <div class="field-hint"
            >先决定谁有资格进入这个菜单空间，进入后菜单入口与受管页面都统一复用后端访问编译结果。</div
          >
        </ElFormItem>
        <ElFormItem v-if="spaceForm.access_mode === 'role_codes'" label="允许空间角色码">
          <ElInput
            v-model="allowedRoleCodesText"
            type="textarea"
            :rows="3"
            placeholder="多个角色码用英文逗号分隔，例如 admin, collaboration_workspace_admin, ops_manager"
          />
        </ElFormItem>
        <div class="menu-space-drawer-grid">
          <ElFormItem label="默认空间">
            <ElSwitch v-model="spaceForm.is_default" />
          </ElFormItem>
          <ElFormItem label="状态">
            <ElSelect v-model="spaceForm.status">
              <ElOption label="启用" value="normal" />
              <ElOption label="停用" value="disabled" />
            </ElSelect>
          </ElFormItem>
        </div>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="spaceDrawerVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingSpace" @click="saveSpace">保存</ElButton>
        </div>
      </template>
    </ElDrawer>

    <ElDrawer v-model="hostDrawerVisible" :title="hostDrawerTitle" size="520px" destroy-on-close>
      <ElForm ref="hostFormRef" :model="hostForm" label-position="top">
        <ElFormItem label="Host / 子域名">
          <ElInput
            v-model="hostForm.host"
            placeholder="例如 admin.example.com 或 collaboration_workspace.example.com"
          />
        </ElFormItem>
        <ElFormItem label="菜单空间">
          <ElSelect v-model="hostForm.space_key" filterable style="width: 100%">
            <ElOption
              v-for="item in spaceOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>
        <div class="menu-space-drawer-grid">
          <ElFormItem label="访问协议">
            <ElSelect v-model="hostForm.meta.scheme">
              <ElOption label="HTTPS" value="https" />
              <ElOption label="HTTP" value="http" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="路由前缀">
            <ElInput v-model="hostForm.meta.route_prefix" placeholder="例如 /admin，可留空" />
          </ElFormItem>
        </div>
        <ElFormItem label="认证策略">
          <ElSelect v-model="hostForm.meta.auth_mode" style="width: 100%">
            <ElOption label="沿用当前 Host 登录" value="inherit_host" />
            <ElOption label="统一登录 Host" value="centralized_login" />
            <ElOption label="共享 Cookie 域" value="shared_cookie" />
          </ElSelect>
          <div class="field-hint"
            >默认沿用当前 Host。后续真正启用多 Host 时，再按这里的策略决定登录入口和回跳方式。</div
          >
        </ElFormItem>
        <div v-if="hostForm.meta.auth_mode === 'centralized_login'" class="menu-space-drawer-grid">
          <ElFormItem label="统一登录 Host">
            <ElInput v-model="hostForm.meta.login_host" placeholder="例如 auth.example.com" />
          </ElFormItem>
          <ElFormItem label="登录回调 Host">
            <ElInput
              v-model="hostForm.meta.callback_host"
              placeholder="例如 admin.example.com，可留空默认当前 Host"
            />
          </ElFormItem>
        </div>
        <div v-if="hostForm.meta.auth_mode === 'shared_cookie'" class="menu-space-drawer-grid">
          <ElFormItem label="Cookie 作用域">
            <ElSelect v-model="hostForm.meta.cookie_scope_mode">
              <ElOption label="沿用默认" value="inherit" />
              <ElOption label="仅当前 Host" value="host_only" />
              <ElOption label="父域共享" value="parent_domain" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="Cookie 域">
            <ElInput
              v-model="hostForm.meta.cookie_domain"
              placeholder="例如 .example.com，可留空"
            />
          </ElFormItem>
        </div>
        <ElFormItem label="说明">
          <ElInput
            v-model="hostForm.description"
            type="textarea"
            :rows="3"
            placeholder="例如 个人空间治理入口 / 协作空间工作区入口"
          />
        </ElFormItem>
        <ElFormItem label="主绑定">
          <ElSwitch v-model="hostForm.is_default" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="hostForm.status">
            <ElOption label="启用" value="normal" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
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
  // 视图脚本：所有 reactive state、handler、watch、lifecycle 均在 useMenuSpacePage 中
  // 这里只做：1) 引入子组件；2) 调用 composable；3) 把返回值拉到 setup 作用域供模板访问。
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useMenuSpacePage } from './modules/use-menu-space-page'

  defineOptions({ name: 'MenuSpaceManage' })

  const {
    loading,
    loadError,
    selectedAppKey,
    savingSpace,
    savingHost,
    savingSpaceMode,
    initializingSpaceKey,
    spaces,
    hostBindings,
    currentSpaceKey,
    spaceMode,
    currentRequestHost,
    landingPathOptions,
    spaceDrawerVisible,
    hostDrawerVisible,
    spaceFormRef,
    hostFormRef,
    spaceForm,
    allowedRoleCodesText,
    hostForm,
    currentSpace,
    currentSpaceLabel,
    spaceModeLabel,
    spaceModeTagType,
    resolveByLabel,
    summaryMetrics,
    appOptions,
    spaceOptions,
    spaceDrawerTitle,
    hostDrawerTitle,
    landingPathHint,
    loadData,
    saveSpaceMode,
    openSpaceDrawer,
    openHostDrawer,
    saveSpace,
    saveHostBinding,
    handleManagedAppChange,
    initializeSpace,
    reinitializeSpace,
    goToMenuManagement,
    goToPageManagement,
    isSpaceInitialized,
    getAccessModeLabel,
    getAccessModeSummary,
    getHostAuthModeLabel
  } = useMenuSpacePage()
</script>

<style scoped lang="scss">
  .menu-space-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .menu-space-mode-select {
    width: clamp(132px, 18vw, 168px);
  }

  .menu-space-app-select {
    width: 240px;
  }

  .menu-space-inline-alert {
    margin-top: 0;
  }

  .menu-space-board {
    display: grid;
    grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
    gap: 16px;
    margin-top: 0;
  }

  .menu-space-panel {
    border-radius: 18px;
  }

  .menu-space-panel__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .menu-space-panel__status {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
  }

  .menu-space-panel__title {
    font-size: 18px;
    font-weight: 600;
    color: var(--art-gray-900);
  }

  .menu-space-panel__desc {
    margin-top: 6px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--art-gray-600);
  }

  .menu-space-list,
  .menu-space-binding-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .menu-space-item,
  .menu-space-binding {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    width: 100%;
    padding: 16px 18px;
    text-align: left;
    background: linear-gradient(180deg, #fbfcff 0%, #f5f8ff 100%);
    border: 1px solid rgba(55, 125, 255, 0.12);
    border-radius: 16px;
    transition: 0.2s ease;
  }

  .menu-space-item:hover,
  .menu-space-binding:hover,
  .menu-space-item.is-current {
    border-color: rgba(55, 125, 255, 0.32);
    box-shadow: 0 12px 30px rgba(55, 125, 255, 0.08);
  }

  .menu-space-item__main,
  .menu-space-binding__main {
    min-width: 0;
    flex: 1;
  }

  .menu-space-item__title-row,
  .menu-space-binding__title-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .menu-space-item__title,
  .menu-space-binding__host {
    font-size: 15px;
    font-weight: 600;
    color: var(--art-gray-900);
  }

  .menu-space-item__meta,
  .menu-space-binding__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 14px;
    margin-top: 8px;
    font-size: 12px;
    color: var(--art-gray-500);
  }

  .menu-space-item__desc {
    margin-top: 12px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--art-gray-700);
  }

  .menu-space-item__hosts {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }

  .menu-space-chip {
    padding: 4px 10px;
    font-size: 12px;
    color: #2156d8;
    background: rgba(55, 125, 255, 0.1);
    border-radius: 999px;
  }

  .menu-space-chip.is-soft {
    color: var(--art-gray-500);
    background: rgba(148, 163, 184, 0.12);
  }

  .menu-space-item__actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 4px;
  }

  .menu-space-empty {
    padding: 28px 16px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--art-gray-500);
    text-align: center;
    background: rgba(148, 163, 184, 0.08);
    border-radius: 16px;
  }

  .menu-space-overview {
    margin-bottom: 14px;
    padding: 14px 16px;
    background: rgba(55, 125, 255, 0.05);
    border: 1px solid rgba(55, 125, 255, 0.12);
    border-radius: 14px;
  }

  .menu-space-overview__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .menu-space-overview__item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .menu-space-overview__label {
    font-size: 12px;
    color: var(--art-gray-500);
  }

  .menu-space-overview__item strong {
    font-size: 14px;
    color: var(--art-gray-900);
    word-break: break-all;
  }

  .menu-space-overview__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }

  .menu-space-drawer-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
  }

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .field-hint {
    margin-top: 8px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--art-gray-500);
  }

  @media (max-width: 1080px) {
    .menu-space-board {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 640px) {
    .menu-space-drawer-grid {
      grid-template-columns: 1fr;
    }

    .menu-space-overview__grid {
      grid-template-columns: 1fr;
    }
  }
</style>
