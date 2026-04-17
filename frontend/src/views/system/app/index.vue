<template>
  <div class="app-manage-page art-full-height" v-loading="loading">
    <AdminWorkspaceHero
      title="应用管理"
      description="以 App 为资源边界统一管理站点默认空间、Host 绑定与导航入口。"
      :metrics="summaryMetrics"
    >
      <div class="app-manage-hero-actions">
        <ElButton type="primary" @click="openAppDrawer()" v-ripple> 新增 App </ElButton>
        <ElButton :disabled="!selectedAppRecord" @click="openAppDrawer(selectedAppRecord)" v-ripple>
          编辑选中 App
        </ElButton>
        <ElButton :disabled="!selectedAppKey" @click="openEntryDialog()" v-ripple>
          新增入口绑定
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
              <div class="app-manage-panel__desc"
                >默认内置 App
                为平台管理后台；后续多前端产品继续沿用同一套用户、权限和导航体系。</div
              >
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
            :data-testid="'app-card'"
            :data-app-key="item.appKey"
            :data-status="item.status"
            :data-is-default="item.isDefault ? 'true' : 'false'"
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
                <span>{{ item.appKey }}</span>
                <span>·</span>
                <span>{{ item.spaceMode === 'multi' ? '多空间' : '单空间' }}</span>
                <span>·</span>
                <span>{{ authModeLabel(item.authMode) }}</span>
                <span>·</span>
                <span>默认 {{ displaySpaceLabel(item.defaultMenuSpaceKey) }}</span>
                <span>·</span>
                <span
                  >{{ item.menuSpaceCount || 0 }} 空间 / {{ item.menuCount || 0 }} 菜单 /
                  {{ item.pageCount || 0 }} 页面</span
                >
              </div>
              <p v-if="item.description" class="app-manage-item__desc">{{ item.description }}</p>
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
              <div class="app-manage-panel__title">APP 入口解析绑定</div>
              <div class="app-manage-panel__desc"
                >按 Host / 路径模式匹配进入 APP，未命中时退回默认
                App。支持精确域名、子域名通配、路径前缀和 host+path 组合。</div
              >
            </div>
            <div class="app-manage-panel__status">
              <ElTag v-if="currentAppRecord" effect="plain" type="warning" size="small">
                解析来源 {{ currentAppResolvedLabel }}
              </ElTag>
              <ElTag v-if="currentAppRequestHost" effect="plain" type="info" size="small">
                请求 Host {{ currentAppRequestHost }}
              </ElTag>
            </div>
          </div>
        </template>

        <!-- App 摘要行 -->
        <div v-if="selectedAppRecord" class="app-overview">
          <div class="app-overview__summary">
            <span>主 Host <strong>{{ selectedAppRecord.primaryHost || '未设置' }}</strong></span>
            <span>·</span>
            <span>默认空间 <strong>{{ displaySpaceLabel(selectedAppRecord.defaultMenuSpaceKey) }}</strong></span>
            <span>·</span>
            <span>前端入口 <strong>{{ selectedAppRecord.frontendEntryUrl || '继承当前地址' }}</strong></span>
          </div>
          <div class="app-overview__actions">
            <ElButton text @click="goToMenuManagement">菜单管理</ElButton>
            <ElButton text @click="goToPageManagement">页面管理</ElButton>
            <ElButton text @click="goToSpaceManagement()">高级空间配置</ElButton>
          </div>
        </div>

        <!-- 面板内切换导航 -->
        <div v-if="selectedAppRecord" class="app-panel-nav">
          <button
            class="app-panel-nav__item"
            :class="{ 'is-active': governanceTab === 'bindings' }"
            @click="governanceTab = 'bindings'"
          >
            入口规则
            <span class="app-panel-nav__badge">{{ hostBindings.length + spaceEntryBindings.length }}</span>
          </button>
          <button
            class="app-panel-nav__item"
            :class="{ 'is-active': governanceTab === 'overview' }"
            @click="governanceTab = 'overview'"
          >
            配置总览
            <ElTag size="small" effect="plain" :type="appReadinessTagType" style="margin-left:6px;vertical-align:middle">
              {{ appReadinessLabel }}
            </ElTag>
          </button>
          <button
            class="app-panel-nav__item"
            :class="{ 'is-active': governanceTab === 'dryrun' }"
            @click="governanceTab = 'dryrun'"
          >
            快速预检
          </button>
          <button
            class="app-panel-nav__item"
            :class="{ 'is-active': governanceTab === 'capabilities' }"
            @click="governanceTab = 'capabilities'"
          >
            接入声明
            <span v-if="capabilityHighlights.length" class="app-panel-nav__badge">{{ capabilityHighlights.length }}</span>
          </button>
        </div>

        <!-- Tab: 入口规则 -->
        <div v-if="selectedAppRecord && governanceTab === 'bindings'" class="app-tab-content">

          <!-- Level 1 -->
          <div class="app-rule-card">
            <div class="app-rule-card__head">
              <div class="app-rule-card__head-left">
                <span class="app-rule-card__level">L1</span>
                <div>
                  <div class="app-rule-card__title">APP 入口</div>
                  <div class="app-rule-card__desc">按 Host / 路径匹配请求归属的 App。精确域名 › 子域名通配 › 路径前缀，未命中退回默认 App。</div>
                </div>
              </div>
              <ElButton size="small" @click="openEntryDialog()">+ 新增</ElButton>
            </div>
            <div v-if="!hostBindings.length" class="app-rule-card__empty">
              暂无规则——缺少规则时此 App 无法独立解析域名或路径，所有请求退回默认 App。
            </div>
            <div
              v-for="item in hostBindings"
              :key="item.id || item.host + item.pathPattern"
              class="app-rule-row"
              @click="openEntryDialog(item)"
            >
              <div class="app-rule-row__body">
                <div class="app-rule-row__title-row">
                  <ElTag size="small" effect="plain">{{ matchTypeLabel(item.matchType) }}</ElTag>
                  <span class="app-rule-row__host">{{ describeEntryRule(item) }}</span>
                  <ElTag v-if="item.isPrimary" size="small" type="success" effect="plain">主</ElTag>
                  <ElTag size="small" :type="item.status === 'normal' ? 'info' : 'danger'" effect="plain">
                    {{ item.status === 'normal' ? '启用' : '停用' }}
                  </ElTag>
                </div>
                <div class="app-rule-row__meta">
                  <span>默认空间 {{ displaySpaceLabel(item.defaultMenuSpaceKey, selectedAppRecord?.defaultMenuSpaceKey) }}</span>
                  <span>优先级 {{ item.priority || 0 }}</span>
                  <span v-if="item.description">{{ item.description }}</span>
                </div>
              </div>
              <ElButton text type="danger" size="small" @click.stop="deleteEntry(item)">删除</ElButton>
            </div>
          </div>

          <!-- Level 2 -->
          <div v-if="isMultiSpaceApp" class="app-rule-card">
            <div class="app-rule-card__head">
              <div class="app-rule-card__head-left">
                <span class="app-rule-card__level app-rule-card__level--l2">L2</span>
                <div>
                  <div class="app-rule-card__title">菜单空间入口</div>
                  <div class="app-rule-card__desc">在 L1 命中后，进一步按路径决定进入哪个菜单空间。未命中按 App 默认空间进入。仅多空间模式生效。</div>
                </div>
              </div>
              <ElButton size="small" :disabled="!spaces.length" @click="openSpaceEntryDialog()">+ 新增</ElButton>
            </div>
            <div v-if="!spaceEntryBindings.length" class="app-rule-card__empty">
              暂无规则——所有请求均落入 App 默认空间。
            </div>
            <div
              v-for="item in spaceEntryBindings"
              :key="item.id || item.menuSpaceKey + item.host + item.pathPattern"
              class="app-rule-row"
              @click="openSpaceEntryDialog(item)"
            >
              <div class="app-rule-row__body">
                <div class="app-rule-row__title-row">
                  <ElTag size="small" effect="plain">{{ matchTypeLabel(item.matchType) }}</ElTag>
                  <span class="app-rule-row__host">{{ describeEntryRule(item) }}</span>
                  <ElTag size="small" type="warning" effect="plain">→ {{ item.spaceName || item.menuSpaceKey }}</ElTag>
                  <ElTag size="small" :type="item.status === 'normal' ? 'info' : 'danger'" effect="plain">
                    {{ item.status === 'normal' ? '启用' : '停用' }}
                  </ElTag>
                </div>
                <div class="app-rule-row__meta">
                  <span>优先级 {{ item.priority || 0 }}</span>
                  <span v-if="item.description">{{ item.description }}</span>
                </div>
              </div>
              <ElButton text type="danger" size="small" @click.stop="deleteSpaceEntry(item)">删除</ElButton>
            </div>
          </div>

        </div>

        <!-- Tab: 配置总览 -->
        <div v-if="selectedAppRecord && governanceTab === 'overview'" class="app-tab-content">
          <div class="app-tab-header">
            <div class="app-tab-header__title">配置总览</div>
            <div class="app-tab-header__desc">
              检查此 App 关键配置项的填写状态。每一项都会影响 App 的解析与接入质量，黄色表示缺少建议配置，不影响基本运行但会降低可观测性。
            </div>
          </div>
          <div class="app-governance-meta">
            <span>空间模式 {{ selectedAppRecord.spaceMode === 'multi' ? '多空间' : '单空间' }}</span>
            <span>认证 {{ authModeLabel(selectedAppRecord.authMode) }}</span>
            <span>APP 入口 {{ hostBindings.length || 0 }} 条</span>
            <span>空间入口 {{ spaceEntryBindings.length || 0 }} 条</span>
          </div>
          <div class="app-governance-checks app-governance-checks--full">
            <div
              v-for="item in appRegistrationChecks"
              :key="item.key"
              class="app-governance-check"
              :class="`is-${item.level}`"
            >
              <div class="app-governance-check__title">{{ item.title }}</div>
              <div class="app-governance-check__text">{{ item.text }}</div>
            </div>
          </div>
        </div>

        <!-- Tab: 快速预检 -->
        <div v-if="selectedAppRecord && governanceTab === 'dryrun'" class="app-tab-content">
          <div class="app-tab-header">
            <div class="app-tab-header__title">快速预检</div>
            <div class="app-tab-header__desc">
              当前 App 的实时解析快照。展示入口命中情况、首跳落点、Manifest 和健康探针状态，帮助快速定位接入问题。数据由后端聚合，不在本地推断。
            </div>
          </div>
          <div class="app-preview-list app-preview-list--full">
            <div v-for="item in appDryRunPreview" :key="item.label" class="app-preview-item">
              <div class="app-preview-item__label">{{ item.label }}</div>
              <code class="app-preview-item__value">{{ item.value }}</code>
              <div class="app-preview-item__hint">{{ item.hint }}</div>
            </div>
          </div>
        </div>

        <!-- Tab: 接入声明 -->
        <div v-if="selectedAppRecord && governanceTab === 'capabilities'" class="app-tab-content">
          <div class="app-tab-header">
            <div class="app-tab-header__title">接入声明</div>
            <div class="app-tab-header__desc">
              当前只保留接入相关的最小声明：认证接入方式和安全控制项。业务 App 自身的运行能力、环境配置和内部治理信息不再在这里编辑。
            </div>
          </div>
          <div class="app-capability-pills">
            <span v-for="item in capabilityHighlights" :key="item" class="app-capability-pill">{{ item }}</span>
            <span v-if="!capabilityHighlights.length" class="app-capability-pill is-soft">当前未登记额外接入安全项。</span>
          </div>
        </div>
      </ElCard>
    </section>

    <ElDrawer v-model="appDrawerVisible" :title="appDrawerTitle" size="50%" destroy-on-close>
      <ElForm ref="appFormRef" :model="appForm" :rules="appFormRules" label-position="top" class="app-drawer-form">

        <!-- ① 基础标识 -->
        <div class="app-form-card">
          <div class="app-form-card__header">
            <span class="app-form-card__title">基础标识</span>
            <span class="app-form-card__desc">
              App Key 创建后不可修改，是跨端导航、权限和 Host 解析的唯一主键。
            </span>
          </div>
          <div class="app-drawer-grid">
            <ElFormItem
              label="应用名称"
              prop="name"
              :error="appFieldErrors.name"
              :data-testid="'app-field-error'"
              :data-field="'name'"
              required
            >
              <ElInput v-model="appForm.name" placeholder="例如 平台管理后台" />
            </ElFormItem>
            <ElFormItem
              label="应用标识（app_key）"
              prop="app_key"
              :error="appFieldErrors.app_key"
              :data-testid="'app-field-error'"
              :data-field="'app_key'"
              required
            >
              <ElInput
                v-model="appForm.app_key"
                :disabled="Boolean(editingAppKey)"
                placeholder="例如 platform-admin"
              />
            </ElFormItem>
          </div>
          <ElFormItem label="说明">
            <ElInput
              v-model="appForm.description"
              type="textarea"
              :rows="2"
              placeholder="说明这个 App 面向哪个站点或后台产品"
            />
          </ElFormItem>
          <div class="app-drawer-grid">
            <ElFormItem label="默认 App">
              <div class="app-form-switch-row">
                <ElSwitch v-model="appForm.is_default" />
                <span class="app-form-switch-label">同一时间只有一个 App 为默认</span>
              </div>
            </ElFormItem>
            <ElFormItem label="状态">
              <ElSelect v-model="appForm.status" style="width: 100%">
                <ElOption label="启用" value="normal" />
                <ElOption label="停用" value="disabled" />
              </ElSelect>
            </ElFormItem>
          </div>
          <div v-if="!editingAppKey" class="app-form-tip">
            新建时系统自动为此 App 创建默认空间 <code>default</code>，无需手动选择。
          </div>
        </div>

        <!-- ② 空间配置 -->
        <div class="app-form-card">
          <div class="app-form-card__header">
            <span class="app-form-card__title">空间配置</span>
            <span class="app-form-card__desc">
              决定此 App 有几个导航空间。多空间支持按 Host 或路径分别进入不同空间。
            </span>
          </div>
          <div class="app-drawer-grid">
            <ElFormItem label="空间模式">
              <ElSelect v-model="appForm.space_mode" style="width: 100%">
                <ElOption label="单空间" value="single" />
                <ElOption label="多空间" value="multi" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem v-if="editingAppKey" label="默认菜单空间">
              <ElSelect
                v-model="defaultMenuSpaceKey"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
              >
                <ElOption
                  v-for="item in spaces"
                  :key="item.menuSpaceKey"
                  :label="`${item.name} · ${item.menuSpaceKey}`"
                  :value="item.menuSpaceKey"
                />
              </ElSelect>
            </ElFormItem>
          </div>
          <div v-if="editingAppKey" class="app-form-tip">
            默认空间决定未命中 Level 2 入口时落入哪个空间；空间首页请到"高级空间配置"调整
            <code>default_home_path</code>。
          </div>
        </div>

        <!-- ③ 认证配置 -->
        <div class="app-form-card">
          <div class="app-form-card__header">
            <span class="app-form-card__title">认证配置</span>
            <span class="app-form-card__desc">
              决定登录如何发生、是否复用认证中心会话，以及登录页面使用哪套 UI。
            </span>
          </div>
          <div class="app-drawer-grid">
            <ElFormItem label="认证模式">
              <ElSelect v-model="appForm.auth_mode" style="width: 100%">
                <ElOption label="继承当前 Host" value="inherit_host" />
                <ElOption label="共享 Cookie" value="shared_cookie" />
                <ElOption label="独立认证中心" value="centralized_login" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="SSO 策略">
              <ElSelect v-model="appAuthSsoMode" style="width: 100%">
                <ElOption label="participate — 复用中心会话" value="participate" />
                <ElOption label="reauth — 进入须重新认证" value="reauth" />
                <ElOption label="isolated — 完全不复用登录态" value="isolated" />
              </ElSelect>
            </ElFormItem>
          </div>

          <ElFormItem label="登录页模式">
            <ElSelect v-model="appAuthLoginUiMode" style="width: 100%">
              <ElOption label="auth_center_ui — 使用认证中心默认页（当前）" value="auth_center_ui" />
              <ElOption label="auth_center_custom — 使用认证中心自定义模板" value="auth_center_custom" />
              <ElOption label="local_ui — 业务 App 自建登录页" value="local_ui" />
            </ElSelect>
          </ElFormItem>

          <!-- 联动：auth_center_ui — 简要说明 -->
          <div v-if="appAuthLoginUiMode === 'auth_center_ui'" class="app-form-tip">
            使用认证中心默认页（<code>/account/auth/login</code>），模板 Key 由系统自动解析：
            URL query → 注册入口 → App 级 → <code>default</code>，无需在此额外配置。
          </div>

          <!-- 联动：auth_center_custom — 显示模板 Key 输入 -->
          <div v-if="appAuthLoginUiMode === 'auth_center_custom'" class="app-form-sub-card">
            <div class="app-form-sub-card__label">
              指定认证中心模板
            </div>
            <ElFormItem label="登录页模板 Key" style="margin-bottom: 0">
              <ElInput v-model="appAuthLoginPageKey" placeholder="例如 aurora" />
              <div class="app-form-hint">
                此处是 App 级兜底，优先级低于 URL query 和注册入口的模板 Key 设置。
                模板需在"认证页模板管理"中预先创建。
              </div>
            </ElFormItem>
          </div>

          <!-- 联动：local_ui — 警告说明 -->
          <ElAlert
            v-if="appAuthLoginUiMode === 'local_ui'"
            type="warning"
            :closable="false"
            show-icon
            title="local_ui 功能尚在建设中"
            description="此模式声明 App 使用自建登录页，认证中心登录页模板对本 App 不生效。当前路由守卫跳转逻辑待补全，实际行为与 auth_center_ui 一致，请知悉。"
            style="margin-top: 4px"
          />
        </div>

        <!-- ④ 运行入口与探针 -->
        <div class="app-form-card">
          <div class="app-form-card__header">
            <span class="app-form-card__title">运行入口与部署探针</span>
            <span class="app-form-card__desc">
              可公开的入口地址与健康检查地址。环境差异由各业务 App 自己维护，这里只登记平台接入所需的统一入口。
            </span>
          </div>
          <div class="app-drawer-grid">
            <ElFormItem label="前端入口地址">
              <ElInput
                v-model="appForm.frontend_entry_url"
                placeholder="/account 或 https://account.example.com"
              />
            </ElFormItem>
            <ElFormItem label="后端入口地址">
              <ElInput
                v-model="appForm.backend_entry_url"
                placeholder="/api 或 https://api.example.com"
              />
            </ElFormItem>
            <ElFormItem label="Manifest 地址">
              <ElInput
                v-model="appForm.manifest_url"
                placeholder="https://cdn.example.com/app/manifest.json"
              />
            </ElFormItem>
            <ElFormItem label="运行版本">
              <ElInput v-model="appForm.runtime_version" placeholder="2026.04.12-rc1" />
            </ElFormItem>
          </div>
          <ElFormItem label="健康检查地址" style="margin-bottom: 0">
            <ElInput v-model="appForm.health_check_url" placeholder="/health" />
          </ElFormItem>
        </div>

        <!-- ⑤ 接入安全 -->
        <div class="app-form-card">
          <div class="app-form-card__header">
            <span class="app-form-card__title">接入安全</span>
            <span class="app-form-card__desc">
              当前只保留平台已经真实消费的安全项：CORS 来源白名单与 CSP。其他运行能力、环境配置和治理细节先不在这里维护。
            </span>
          </div>
          <div class="app-config-entry">
            <div class="app-config-entry__summary">
              <ElTag
                v-for="item in capabilityDialogSummary"
                :key="item"
                size="small"
                effect="plain"
              >
                {{ item }}
              </ElTag>
              <span v-if="!capabilityDialogSummary.length" class="app-config-entry__empty">
                尚未登记额外安全项
              </span>
            </div>
            <div class="app-config-entry__actions">
              <ElButton @click="openCapabilityDialog()">配置接入安全</ElButton>
            </div>
          </div>
          <div v-if="capabilityRawError" class="app-config-entry__error">{{ capabilityRawError }}</div>
        </div>

      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="appDrawerVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingApp" @click="saveApp">保存</ElButton>
        </div>
      </template>
    </ElDrawer>

    <ElDialog
      v-model="capabilityDialogVisible"
      title="接入安全"
      width="720px"
      destroy-on-close
      append-to-body
    >
      <div class="app-config-dialog__intro">
        这里只维护平台已经真实消费的接入安全项。业务 App 自己的运行能力、环境差异和内部治理信息不在这里编辑。
      </div>
      <div class="app-config-panel">
        <div class="app-config-panel__title">CORS 来源白名单</div>
        <div class="app-config-panel__desc">
          对应 <code>capabilities.cors_origins</code>，会被动态安全中间件直接消费。
        </div>
        <div
          v-for="(item, index) in capabilityForm.security.corsOrigins"
          :key="`cors-${index}`"
          class="app-inline-row"
        >
          <ElInput
            v-model="capabilityForm.security.corsOrigins[index]"
            placeholder="https://app.example.com"
          />
          <ElButton text type="danger" @click="removeCapabilityCorsOrigin(index)">删除</ElButton>
        </div>
        <ElButton text @click="addCapabilityCorsOrigin()">+ 新增来源</ElButton>
        <ElFormItem label="CSP" style="margin-top: 12px; margin-bottom: 0">
          <ElInput
            v-model="capabilityForm.security.csp"
            type="textarea"
            :rows="4"
            placeholder="default-src 'self'; frame-ancestors 'self';"
          />
        </ElFormItem>
      </div>
      <details class="app-config-advanced">
        <summary>高级 JSON 模式</summary>
        <div class="app-config-advanced__body">
          <ElAlert
            v-if="capabilityRawError"
            type="warning"
            :closable="false"
            show-icon
            :title="capabilityRawError"
            class="app-config-advanced__alert"
          />
          <ElInput
            v-model="appCapabilitiesText"
            type="textarea"
            :rows="12"
            placeholder='{"cors_origins":["https://app.example.com"],"csp":"default-src self;"}'
          />
          <div class="app-config-advanced__actions">
            <ElButton @click="syncCapabilityFormFromRaw(true)">从 JSON 覆盖表单</ElButton>
            <ElButton text @click="syncCapabilityRawFromForm(true)">用表单结果回写 JSON</ElButton>
          </div>
        </div>
      </details>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="capabilityDialogVisible = false">取消</ElButton>
          <ElButton type="primary" @click="saveCapabilityDialog()">确认</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog
      v-model="entryDialogVisible"
      :title="entryDialogTitle"
      width="560px"
      destroy-on-close
      append-to-body
    >
      <ElForm ref="entryFormRef" :model="entryForm" label-position="top">
        <ElFormItem label="匹配类型" prop="match_type">
          <ElRadioGroup v-model="entryForm.match_type">
            <ElRadioButton value="host_exact">精确域名</ElRadioButton>
            <ElRadioButton value="host_suffix">子域名通配</ElRadioButton>
            <ElRadioButton value="path_prefix">路径模式</ElRadioButton>
            <ElRadioButton value="host_and_path">域名+路径</ElRadioButton>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem
          v-if="entryNeedsHost"
          label="Host"
          prop="host"
          :error="entryFieldErrors.host"
          :data-testid="'app-field-error'"
          :data-field="'host'"
        >
          <ElInput v-model="entryForm.host" :placeholder="entryHostPlaceholder" />
          <div
            v-if="entryFieldErrors.host"
            data-testid="host-conflict-reason"
            :data-field="'host'"
            class="app-form-hint"
            style="color: var(--el-color-danger)"
          >
            {{ entryFieldErrors.host }}
          </div>
        </ElFormItem>
        <ElFormItem
          v-if="entryNeedsPath"
          label="路径模式"
          prop="path_pattern"
          :error="entryFieldErrors.path_pattern"
          :data-testid="'app-field-error'"
          :data-field="'path_pattern'"
        >
          <ElInput v-model="entryForm.path_pattern" placeholder="例如 /admin/** 或 /shop/:id/**" />
          <div class="app-form-hint">
            支持
            <code>*</code>（单段通配）、<code>**</code>（多段通配）、<code>:name</code>（命名参数）
          </div>
        </ElFormItem>
            <ElFormItem label="默认菜单空间">
              <ElSelect
                v-model="entryDefaultMenuSpaceKey"
                filterable
                allow-create
                default-first-option
            style="width: 100%"
          >
            <ElOption
              v-for="item in spaces"
              :key="item.menuSpaceKey"
              :label="`${item.name} · ${item.menuSpaceKey}`"
              :value="item.menuSpaceKey"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput
            v-model="entryForm.description"
            type="textarea"
            :rows="2"
            placeholder="例如 平台治理入口 / 商家后台入口"
          />
        </ElFormItem>
        <div class="app-drawer-grid">
          <ElFormItem label="优先级">
            <ElInputNumber v-model="entryForm.priority" :min="0" :max="999" />
          </ElFormItem>
          <ElFormItem label="主绑定">
            <ElSwitch v-model="entryForm.is_primary" />
          </ElFormItem>
          <ElFormItem label="状态">
            <ElSelect v-model="entryForm.status">
              <ElOption label="启用" value="normal" />
              <ElOption label="停用" value="disabled" />
            </ElSelect>
          </ElFormItem>
        </div>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="entryDialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingHost" @click="saveEntryBinding">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog
      v-model="spaceEntryDialogVisible"
      :title="spaceEntryDialogTitle"
      width="560px"
      destroy-on-close
      append-to-body
    >
      <ElForm ref="spaceEntryFormRef" :model="spaceEntryForm" label-position="top">
        <ElFormItem label="目标菜单空间" prop="menu_space_key"
          :error="spaceEntryFieldErrors.menu_space_key"
          :data-testid="'app-field-error'"
          :data-field="'menu_space_key'">
          <ElSelect v-model="spaceEntryMenuSpaceKey" filterable style="width: 100%">
            <ElOption
              v-for="item in spaces"
              :key="item.menuSpaceKey"
              :label="`${item.name} · ${item.menuSpaceKey}`"
              :value="item.menuSpaceKey"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="匹配类型">
          <ElRadioGroup v-model="spaceEntryForm.match_type">
            <ElRadioButton value="host_exact">精确域名</ElRadioButton>
            <ElRadioButton value="host_suffix">子域名通配</ElRadioButton>
            <ElRadioButton value="path_prefix">路径模式</ElRadioButton>
            <ElRadioButton value="host_and_path">域名+路径</ElRadioButton>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem
          v-if="spaceEntryNeedsHost"
          label="Host"
          prop="host"
          :error="spaceEntryFieldErrors.host"
          :data-testid="'app-field-error'"
          :data-field="'host'"
        >
          <ElInput v-model="spaceEntryForm.host" :placeholder="spaceEntryHostPlaceholder" />
          <div
            v-if="spaceEntryFieldErrors.host"
            data-testid="host-conflict-reason"
            :data-field="'host'"
            class="app-form-hint"
            style="color: var(--el-color-danger)"
          >
            {{ spaceEntryFieldErrors.host }}
          </div>
        </ElFormItem>
        <ElFormItem
          v-if="spaceEntryNeedsPath"
          label="路径模式"
          prop="path_pattern"
          :error="spaceEntryFieldErrors.path_pattern"
          :data-testid="'app-field-error'"
          :data-field="'path_pattern'"
        >
          <ElInput v-model="spaceEntryForm.path_pattern" placeholder="例如 /a/** 或 /shop/:id" />
          <div class="app-form-hint">
            支持 <code>*</code> / <code>**</code> / <code>:name</code>，且必须落在 APP
            入口规则范围内。
          </div>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput
            v-model="spaceEntryForm.description"
            type="textarea"
            :rows="2"
            placeholder="例如 商家后台 /shop 路径进入 shop 空间"
          />
        </ElFormItem>
        <div class="app-drawer-grid">
          <ElFormItem label="优先级">
            <ElInputNumber v-model="spaceEntryForm.priority" :min="0" :max="999" />
          </ElFormItem>
          <ElFormItem label="主绑定">
            <ElSwitch v-model="spaceEntryForm.is_primary" />
          </ElFormItem>
          <ElFormItem label="状态">
            <ElSelect v-model="spaceEntryForm.status">
              <ElOption label="启用" value="normal" />
              <ElOption label="停用" value="disabled" />
            </ElSelect>
          </ElFormItem>
        </div>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="spaceEntryDialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingSpaceEntry" @click="saveSpaceEntryBinding"
            >保存</ElButton
          >
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { HttpError } from '@/utils/http/error'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useManagedAppScope } from '@/domains/app-runtime/useManagedAppScope'
  import {
    fetchGetApps,
    fetchGetAppHostBindings,
    fetchGetAppPreflight,
    fetchGetCurrentApp,
    fetchGetMenuSpaces,
    fetchSaveApp,
    fetchSaveAppHostBinding,
    fetchDeleteAppHostBinding,
    fetchGetMenuSpaceEntryBindings,
    fetchSaveMenuSpaceEntryBinding,
    fetchDeleteMenuSpaceEntryBinding
  } from '@/domains/governance/api'
  import {
    createCapabilityFormState,
    formatJsonObject,
    omitDeprecatedCapabilityFields,
    omitEditableCapabilitySections,
    parseJSONObjectText,
    pickEditableCapabilitySections,
    serializeCapabilityFormState,
    summarizeManagedCapabilities
  } from './config-editor'
  import type { CapabilityFormState } from './config-editor'

  defineOptions({ name: 'AppManage' })

  const router = useRouter()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const managedAppMissingText = '请先从 App 列表选择要管理的应用'
  const loading = ref(false)
  const loadError = ref('')
  const savingApp = ref(false)
  const savingHost = ref(false)
  const governanceTab = ref<'bindings' | 'overview' | 'dryrun' | 'capabilities'>('bindings')
  const apps = ref<Api.SystemManage.AppItem[]>([])
  const hostBindings = ref<Api.SystemManage.AppHostBindingItem[]>([])
  const spaceEntryBindings = ref<Api.SystemManage.MenuSpaceEntryBindingItem[]>([])
  const spaces = ref<Api.SystemManage.MenuSpaceItem[]>([])
  const currentApp = ref<Api.SystemManage.CurrentAppResponse>()
  const appPreflight = ref<Api.SystemManage.AppPreflightResponse>()
  const selectedAppKey = ref('')

  const appDrawerVisible = ref(false)
  const entryDialogVisible = ref(false)
  const spaceEntryDialogVisible = ref(false)
  const capabilityDialogVisible = ref(false)
  const savingSpaceEntry = ref(false)
  const editingAppKey = ref('')
  const editingEntryId = ref('')
  const editingSpaceEntryId = ref('')

  // ── 字段级错误回显：三个表单独立维护 fieldErrors；后端 Error.details.<field> 打进对应表
  // 规范：docs/guides/frontend-observability-spec.md §2.4。
  const appFormRef = ref<FormInstance>()
  const appFieldErrors = reactive<Record<string, string>>({})
  const entryFormRef = ref<FormInstance>()
  const entryFieldErrors = reactive<Record<string, string>>({})
  const spaceEntryFormRef = ref<FormInstance>()
  const spaceEntryFieldErrors = reactive<Record<string, string>>({})

  const appFormRules: FormRules = {
    app_key: [
      { required: true, message: '请输入应用标识', trigger: 'blur' },
      { pattern: /^[a-z0-9][a-z0-9-]*$/, message: '仅允许小写字母数字和短横线', trigger: 'blur' }
    ],
    name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }]
  }

  function clearFieldErrors(target: Record<string, string>) {
    for (const k of Object.keys(target)) delete target[k]
  }

  function applyBackendFieldErrors(target: Record<string, string>, e: unknown): boolean {
    if (!(e instanceof HttpError)) return false
    const data = (e.data || {}) as { details?: Record<string, string> }
    const details = data.details
    if (!details || typeof details !== 'object') return false
    let applied = false
    for (const [field, reason] of Object.entries(details)) {
      if (typeof reason === 'string') {
        target[field] = reason
        applied = true
      }
    }
    return applied
  }
  const appForm = reactive<Api.SystemManage.AppSaveParams>({
    app_key: '',
    name: '',
    description: '',
    space_mode: 'single',
    default_menu_space_key: '',
    auth_mode: 'inherit_host',
    frontend_entry_url: '',
    backend_entry_url: '',
    health_check_url: '',
    manifest_url: '',
    runtime_version: '',
    capabilities: {},
    is_default: false,
    status: 'normal',
    meta: {}
  })
  const appCapabilitiesText = ref('{}')
  const appCapabilitiesBase = ref<Record<string, unknown>>({})
  const appAuthSsoMode = ref<'participate' | 'reauth' | 'isolated'>('participate')
  const appAuthLoginUiMode = ref<'auth_center_ui' | 'auth_center_custom' | 'local_ui'>(
    'auth_center_ui'
  )
  const appAuthLoginPageKey = ref('default')
  const appMetaBase = ref<Record<string, unknown>>({})
  const capabilityRawError = ref('')
  const capabilityForm = reactive<CapabilityFormState>(createCapabilityFormState())
  type MatchType = Api.SystemManage.AppHostBindingSaveParams['match_type']

  const entryForm = reactive<Api.SystemManage.AppHostBindingSaveParams>({
    id: '',
    app_key: '',
    match_type: 'host_exact',
    host: '',
    path_pattern: '',
    priority: 0,
    default_menu_space_key: '',
    description: '',
    is_primary: false,
    status: 'normal',
    meta: {}
  })

  const spaceEntryForm = reactive<Api.SystemManage.MenuSpaceEntryBindingSaveParams>({
    id: '',
    app_key: '',
    menu_space_key: '',
    match_type: 'host_exact',
    host: '',
    path_pattern: '',
    priority: 0,
    description: '',
    is_primary: false,
    status: 'normal',
    meta: {}
  })
  const defaultMenuSpaceKey = computed({
    get: () => appForm.default_menu_space_key,
    set: (value: string) => {
      appForm.default_menu_space_key = value
    }
  })
  const entryDefaultMenuSpaceKey = computed({
    get: () => entryForm.default_menu_space_key,
    set: (value: string) => {
      entryForm.default_menu_space_key = value
    }
  })
  const spaceEntryMenuSpaceKey = computed({
    get: () => spaceEntryForm.menu_space_key,
    set: (value: string) => {
      spaceEntryForm.menu_space_key = value
    }
  })

  const matchTypeLabelMap: Record<string, string> = {
    host_exact: '精确域名',
    host_suffix: '子域名',
    path_prefix: '路径',
    host_and_path: '域名+路径'
  }
  const authModeLabelMap: Record<string, string> = {
    inherit_host: '继承 Host',
    shared_cookie: '共享 Cookie',
    centralized_login: '独立认证'
  }

  function matchTypeLabel(type?: string) {
    return matchTypeLabelMap[type || 'host_exact'] || type || ''
  }

  function authModeLabel(type?: string) {
    return authModeLabelMap[type || 'inherit_host'] || type || 'inherit_host'
  }

  function describeEntryRule(item: { matchType?: string; host?: string; pathPattern?: string }) {
    const host = item.host || ''
    const path = item.pathPattern || ''
    switch (item.matchType) {
      case 'host_suffix':
        return `*${host.startsWith('.') ? host : '.' + host}`
      case 'path_prefix':
        return path || '/'
      case 'host_and_path':
        return `${host}${path}`
      default:
        return host || path || '-'
    }
  }

  const isMultiSpaceApp = computed(() => selectedAppRecord.value?.spaceMode === 'multi')

  const entryNeedsHost = computed(() =>
    ['host_exact', 'host_suffix', 'host_and_path'].includes(`${entryForm.match_type}`)
  )
  const entryNeedsPath = computed(() =>
    ['path_prefix', 'host_and_path'].includes(`${entryForm.match_type}`)
  )
  const entryHostPlaceholder = computed(() =>
    entryForm.match_type === 'host_suffix' ? '例如 .example.com' : '例如 admin.example.com'
  )
  const entryDialogTitle = computed(() => (editingEntryId.value ? '编辑入口绑定' : '新增入口绑定'))

  const spaceEntryNeedsHost = computed(() =>
    ['host_exact', 'host_suffix', 'host_and_path'].includes(`${spaceEntryForm.match_type}`)
  )
  const spaceEntryNeedsPath = computed(() =>
    ['path_prefix', 'host_and_path'].includes(`${spaceEntryForm.match_type}`)
  )
  const spaceEntryHostPlaceholder = computed(() =>
    spaceEntryForm.match_type === 'host_suffix' ? '例如 .example.com' : '例如 shop.example.com'
  )
  const spaceEntryDialogTitle = computed(() =>
    editingSpaceEntryId.value ? '编辑菜单空间入口' : '新增菜单空间入口'
  )

  const currentAppRecord = computed(() => currentApp.value?.app)
  const currentAppRequestHost = computed(() => `${currentApp.value?.requestHost || ''}`.trim())
  const selectedAppRecord = computed(() =>
    apps.value.find((item) => item.appKey === selectedAppKey.value)
  )
  const currentAppResolvedLabel = computed(() => {
    switch (`${currentApp.value?.resolvedBy || ''}`.trim()) {
      case 'host_binding':
        return 'Host 绑定'
      case 'explicit':
        return '显式指定'
      case 'default_app':
        return '默认 App'
      default:
        return `${currentApp.value?.resolvedBy || '默认 App'}`
    }
  })
  const appDrawerTitle = computed(() => (editingAppKey.value ? '编辑应用' : '新增应用'))
  const summaryMetrics = computed(() => [
    { label: '应用数', value: apps.value.length || 0 },
    { label: '管理 App', value: selectedAppRecord.value?.name || selectedAppKey.value || '未选择' },
    { label: '菜单空间', value: selectedAppRecord.value?.menuSpaceCount || 0 },
    { label: 'Host 绑定', value: hostBindings.value.length || 0 }
  ])
  const appRegistrationChecks = computed(() => {
    const record = selectedAppRecord.value
    if (!record) return []
    const hasPrimaryBinding =
      Boolean(record.primaryHost) || hostBindings.value.some((item) => item.isPrimary)
    const hasFrontendEntry = Boolean(`${record.frontendEntryUrl || ''}`.trim())
    const hasBackendEntry = Boolean(`${record.backendEntryUrl || ''}`.trim())
    const hasHealthCheck = Boolean(`${record.healthCheckUrl || ''}`.trim())
    const hasManifest = Boolean(`${record.manifestUrl || ''}`.trim())
    const hasRuntimeVersion = Boolean(`${record.runtimeVersion || ''}`.trim())
    const probeStatus = `${record.probeStatus || ''}`.trim()
    const securitySummary = summarizeManagedCapabilities(
      (record.capabilities || {}) as Record<string, any>
    )
    const hasSecurityDeclarations = Boolean(securitySummary.length)
    const authMode = `${record.authMode || 'inherit_host'}`.trim()
    return [
      {
        key: 'binding',
        title: '入口绑定',
        level: hasPrimaryBinding ? 'success' : 'warning',
        text: hasPrimaryBinding
          ? `已配置主入口，当前共 ${hostBindings.value.length || 0} 条 APP 入口规则。`
          : '缺少主入口绑定，解析会退回默认 App。'
      },
      {
        key: 'frontend',
        title: '前端入口',
        level: hasFrontendEntry ? 'success' : 'warning',
        text: hasFrontendEntry
          ? `前端入口已声明为 ${record.frontendEntryUrl}。`
          : '未声明前端入口，将依赖当前地址或 host 规则推断。'
      },
      {
        key: 'backend',
        title: '后端接入',
        level: hasBackendEntry ? 'success' : 'info',
        text: hasBackendEntry
          ? `后端入口已声明为 ${record.backendEntryUrl}。`
          : '未声明后端入口，适合仍与主站共用 API 网关的场景。'
      },
      {
        key: 'manifest',
        title: '远端清单',
        level: hasManifest ? 'success' : 'info',
        text: hasManifest
          ? `已声明 manifest 地址 ${record.manifestUrl}。`
          : '未声明 manifest 地址，远端页来源仍需从页面契约补齐。'
      },
      {
        key: 'runtime-version',
        title: '运行版本',
        level: hasRuntimeVersion ? 'success' : 'info',
        text: hasRuntimeVersion
          ? `当前登记运行版本 ${record.runtimeVersion}。`
          : '未声明运行版本，治理台暂时无法直接比对远端接入版本。'
      },
      {
        key: 'health',
        title: '运行探针',
        level:
          probeStatus === 'healthy' ? 'success' : probeStatus === 'missing' ? 'info' : 'warning',
        text:
          probeStatus === 'healthy'
            ? `探针最近一次探测成功：${record.probeTarget || record.healthCheckUrl || '已探测'}，${record.probeMessage || '运行正常'}。`
            : probeStatus === 'missing'
              ? '未声明健康检查地址，平台暂时无法统一展示运行探针。'
              : `探针最近一次探测失败：${record.probeTarget || record.healthCheckUrl || '未命中目标'}，${record.probeMessage || '待排查'}。`
      },
      {
        key: 'auth',
        title: '认证治理',
        level: authMode === 'centralized_login' ? 'warning' : 'success',
        text:
          authMode === 'centralized_login'
            ? '当前为独立认证中心模式。如需统一治理回调域名或 issuer，再按平台能力单独收口，不在当前页面展开。'
            : `当前为 ${authModeLabel(authMode)} 模式，登录态可沿用现有主站协同链路。`
      },
      {
        key: 'capabilities',
        title: '接入安全',
        level: hasSecurityDeclarations ? 'success' : 'info',
        text: hasSecurityDeclarations
          ? `已登记 ${securitySummary.join('、')}。`
          : '当前未登记额外的 CORS 或 CSP 控制项，将沿用平台默认安全策略。'
      }
    ]
  })
  const appReadinessTagType = computed(() => {
    const backendLevel = `${appPreflight.value?.summary?.level || ''}`.trim()
    if (backendLevel === 'blocking' || backendLevel === 'warning') return 'warning'
    if (backendLevel === 'info') return 'info'
    const levels = appRegistrationChecks.value.map((item) => item.level)
    if (levels.includes('warning')) return 'warning'
    if (levels.includes('info')) return 'info'
    return 'success'
  })
  const appReadinessLabel = computed(() => {
    switch (appReadinessTagType.value) {
      case 'warning':
        return '需补治理信息'
      case 'info':
        return '基础可用'
      default:
        return '配置完整'
    }
  })
  const appDryRunPreview = computed(() => {
    if (appPreflight.value?.previewItems?.length) {
      return appPreflight.value.previewItems
    }
    const record = selectedAppRecord.value
    if (!record) return []
    const primaryBinding =
      hostBindings.value.find((item) => item.isPrimary) || hostBindings.value[0]
    const entryRule = primaryBinding ? describeEntryRule(primaryBinding) : '未配置'
    const landing = `${record.frontendEntryUrl || ''}`.trim() || '继承当前地址'
    const manifest = `${record.manifestUrl || ''}`.trim() || '未配置'
    const health = `${record.probeStatus || ''}`.trim()
      ? `${record.probeStatus}${record.probeTarget ? ` · ${record.probeTarget}` : ''}`
      : `${record.healthCheckUrl || ''}`.trim() || '未配置'
    return [
      {
        label: '入口命中',
        value: entryRule,
        hint: primaryBinding
          ? `按 ${matchTypeLabel(primaryBinding.matchType)} 规则进入 ${record.appKey}。`
          : '当前没有 APP 入口规则，只能依赖默认 App。'
      },
      {
        label: 'Manifest',
        value: manifest,
        hint:
          manifest === '未配置'
            ? '远端页治理仍会缺少统一清单来源。'
            : '远端模块、版本和页面入口应优先与该清单对齐。'
      },
      {
        label: '首跳落点',
        value: landing,
        hint:
          record.authMode === 'centralized_login'
            ? '登录前通常先进入认证中心，再根据 callback 回跳到这里。'
            : '登录后将以这个入口或当前 URL 作为首跳落点。'
      },
      {
        label: '健康探针',
        value: health,
        hint:
          health === '未配置'
            ? '暂时无法做统一探针聚合。'
            : record.probeMessage || '可作为联调和运行状态检查入口。'
      }
    ]
  })
  const capabilityHighlights = computed(() => {
    const record = selectedAppRecord.value
    if (!record) return []
    const highlights = [`认证 ${authModeLabel(record.authMode)}`]
    const capabilitySummary = summarizeManagedCapabilities(
      (record.capabilities || {}) as Record<string, any>
    )
    return [...highlights, ...capabilitySummary]
  })
  const capabilityDialogSummary = computed(() => {
    const parsed = safeParseJSONObject(appCapabilitiesText.value)
    return summarizeManagedCapabilities(parsed || {})
  })

  function replaceList<T>(target: T[], values: T[]) {
    target.splice(0, target.length, ...values)
  }

  function safeParseJSONObject(rawText: string) {
    try {
      return parseJSONObjectText(rawText, '配置')
    } catch {
      return undefined
    }
  }

  function assignCapabilityForm(next: CapabilityFormState) {
    replaceList(capabilityForm.security.corsOrigins, [...next.security.corsOrigins])
    capabilityForm.security.csp = next.security.csp
  }

  function syncCapabilityFormFromRaw(showMessage = false) {
    try {
      const parsed = parseJSONObjectText(appCapabilitiesText.value, '接入安全') as Record<string, any>
      assignCapabilityForm(createCapabilityFormState(parsed))
      capabilityRawError.value = ''
      if (showMessage) ElMessage.success('接入安全 JSON 已同步到结构化表单')
      return true
    } catch (error: any) {
      capabilityRawError.value = error?.message || '接入安全格式错误'
      if (showMessage) ElMessage.warning(capabilityRawError.value)
      return false
    }
  }

  function syncCapabilityRawFromForm(showMessage = false) {
    appCapabilitiesText.value = formatJsonObject(serializeCapabilityFormState(capabilityForm))
    capabilityRawError.value = ''
    if (showMessage) ElMessage.success('接入安全 JSON 已用表单结果更新')
  }

  function openCapabilityDialog() {
    syncCapabilityFormFromRaw()
    capabilityDialogVisible.value = true
  }

  function saveCapabilityDialog() {
    syncCapabilityRawFromForm()
    capabilityDialogVisible.value = false
  }

  function addCapabilityCorsOrigin() {
    capabilityForm.security.corsOrigins.push('')
  }

  function removeCapabilityCorsOrigin(index: number) {
    capabilityForm.security.corsOrigins.splice(index, 1)
  }

  function resolveAppKey(...candidates: Array<string | undefined | null>) {
    for (const candidate of candidates) {
      const normalized = `${candidate || ''}`.trim()
      if (normalized) {
        return normalized
      }
    }
    return ''
  }

  function resolveSpaceKey(...candidates: Array<string | undefined | null>) {
    for (const candidate of candidates) {
      const normalized = `${candidate || ''}`.trim()
      if (normalized) {
        return normalized
      }
    }
    return ''
  }

  function displaySpaceLabel(...candidates: Array<string | undefined | null>) {
    return resolveSpaceKey(...candidates) || '未设置'
  }

  function parseCapabilitiesText() {
    return parseJSONObjectText(appCapabilitiesText.value, '接入安全') as Record<string, unknown>
  }

  function extractAuthCapability(
    capabilities?: Record<string, any>
  ): { ssoMode: 'participate' | 'reauth' | 'isolated'; loginUiMode: 'auth_center_ui' | 'auth_center_custom' | 'local_ui'; loginPageKey: string } {
    const auth =
      capabilities && typeof capabilities.auth === 'object' && !Array.isArray(capabilities.auth)
        ? capabilities.auth
        : {}
    const rawSso = `${auth.sso_mode || auth.ssoMode || ''}`.trim()
    const rawLoginUi = `${auth.login_ui_mode || auth.loginUiMode || ''}`.trim()
    const ssoMode =
      rawSso === 'reauth' || rawSso === 'isolated' ? rawSso : 'participate'
    const loginUiMode =
      rawLoginUi === 'auth_center_custom' || rawLoginUi === 'local_ui'
        ? rawLoginUi
        : 'auth_center_ui'
    const loginPageKey = `${auth.login_page_key || auth.loginPageKey || ''}`.trim() || 'default'
    return { ssoMode, loginUiMode, loginPageKey }
  }

  function patchAuthCapability(
    capabilities: Record<string, any>,
    authConfig: { ssoMode: string; loginUiMode: string; loginPageKey: string }
  ) {
    const next = {
      ...capabilities,
      auth: {
        ...(capabilities?.auth || {}),
        sso_mode: authConfig.ssoMode,
        login_ui_mode: authConfig.loginUiMode,
        login_page_key: authConfig.loginPageKey
      }
    }
    return next
  }

  function normalizeMatchType(value?: string): MatchType {
    switch (value) {
      case 'host_exact':
      case 'host_suffix':
      case 'path_prefix':
      case 'host_and_path':
        return value
      default:
        return 'host_exact'
    }
  }

  async function loadSelectedAppContext(appKey: string) {
    const normalizedAppKey = resolveAppKey(appKey)
    if (!normalizedAppKey) {
      throw new Error('缺少 app 上下文')
    }
    selectedAppKey.value = normalizedAppKey
    await setManagedAppKey(normalizedAppKey)
    const [hostRes, spaceRes, entryRes, preflightRes] = await Promise.all([
      fetchGetAppHostBindings(normalizedAppKey),
      fetchGetMenuSpaces(normalizedAppKey),
      fetchGetMenuSpaceEntryBindings(normalizedAppKey).catch(() => ({
        records: [] as Api.SystemManage.MenuSpaceEntryBindingItem[]
      })),
      fetchGetAppPreflight(normalizedAppKey).catch(() => undefined)
    ])
    hostBindings.value = hostRes.records || []
    spaces.value = spaceRes.records || []
    spaceEntryBindings.value = entryRes.records || []
    appPreflight.value = preflightRes
  }

  async function loadData() {
    loading.value = true
    loadError.value = ''
    try {
      const [appsRes, currentRes] = await Promise.all([fetchGetApps(), fetchGetCurrentApp()])
      apps.value = appsRes.records || []
      currentApp.value = currentRes
      const nextAppKey = resolveAppKey(targetAppKey.value, selectedAppKey.value)
      if (!nextAppKey) {
        selectedAppKey.value = ''
        hostBindings.value = []
        spaces.value = []
        appPreflight.value = undefined
        loadError.value = managedAppMissingText
        return
      }
      await loadSelectedAppContext(nextAppKey)
    } catch (error: any) {
      apps.value = []
      hostBindings.value = []
      spaces.value = []
      spaceEntryBindings.value = []
      appPreflight.value = undefined
      loadError.value = error?.message || '应用数据暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  function resetAppForm() {
    editingAppKey.value = ''
    appForm.app_key = ''
    appForm.name = ''
    appForm.description = ''
    appForm.space_mode = 'single'
    appForm.default_menu_space_key = ''
    appForm.auth_mode = 'inherit_host'
    appForm.frontend_entry_url = ''
    appForm.backend_entry_url = ''
    appForm.health_check_url = ''
    appForm.manifest_url = ''
    appForm.runtime_version = ''
    appForm.capabilities = {}
    appForm.is_default = false
    appForm.status = 'normal'
    appForm.meta = {}
    appCapabilitiesBase.value = {}
    appCapabilitiesText.value = '{}'
    appAuthSsoMode.value = 'participate'
    appAuthLoginUiMode.value = 'auth_center_ui'
    appAuthLoginPageKey.value = 'default'
    appMetaBase.value = {}
    capabilityRawError.value = ''
    assignCapabilityForm(createCapabilityFormState())
  }

  function resetEntryForm() {
    editingEntryId.value = ''
    entryForm.id = ''
    entryForm.app_key = resolveAppKey(selectedAppKey.value)
    entryForm.match_type = 'host_exact'
    entryForm.host = ''
    entryForm.path_pattern = ''
    entryForm.priority = 0
    entryForm.default_menu_space_key = resolveSpaceKey(
      selectedAppRecord.value?.defaultMenuSpaceKey,
      selectedAppRecord.value?.defaultMenuSpaceKey
    )
    entryForm.description = ''
    entryForm.is_primary = false
    entryForm.status = 'normal'
    entryForm.meta = {}
  }

  function resetSpaceEntryForm() {
    editingSpaceEntryId.value = ''
    spaceEntryForm.id = ''
    spaceEntryForm.app_key = resolveAppKey(selectedAppKey.value)
    spaceEntryForm.menu_space_key = spaces.value[0]?.menuSpaceKey || ''
    spaceEntryForm.match_type = 'host_exact'
    spaceEntryForm.host = ''
    spaceEntryForm.path_pattern = ''
    spaceEntryForm.priority = 0
    spaceEntryForm.description = ''
    spaceEntryForm.is_primary = false
    spaceEntryForm.status = 'normal'
    spaceEntryForm.meta = {}
  }

  function openAppDrawer(item?: Api.SystemManage.AppItem) {
    resetAppForm()
    if (item) {
      editingAppKey.value = item.appKey
      appForm.app_key = item.appKey
      appForm.name = item.name
      appForm.description = item.description || ''
      appForm.space_mode = item.spaceMode === 'multi' ? 'multi' : 'single'
      appForm.default_menu_space_key = resolveSpaceKey(
        item.defaultMenuSpaceKey,
        item.defaultMenuSpaceKey
      )
      appForm.auth_mode = item.authMode || 'inherit_host'
      appForm.frontend_entry_url = item.frontendEntryUrl || ''
      appForm.backend_entry_url = item.backendEntryUrl || ''
      appForm.health_check_url = item.healthCheckUrl || ''
      appForm.manifest_url = item.manifestUrl || ''
      appForm.runtime_version = item.runtimeVersion || ''
      appForm.capabilities = item.capabilities || {}
      appForm.is_default = Boolean(item.isDefault)
      appForm.status = item.status || 'normal'
      appForm.meta = item.meta || {}
      const normalizedCapabilities = omitDeprecatedCapabilityFields(item.capabilities || {})
      appCapabilitiesBase.value = omitEditableCapabilitySections(normalizedCapabilities)
      appCapabilitiesText.value = formatJsonObject(
        pickEditableCapabilitySections(normalizedCapabilities)
      )
      const authCapability = extractAuthCapability(normalizedCapabilities)
      appAuthSsoMode.value = authCapability.ssoMode
      appAuthLoginUiMode.value = authCapability.loginUiMode
      appAuthLoginPageKey.value = authCapability.loginPageKey
      appMetaBase.value = item.meta || {}
      syncCapabilityFormFromRaw()
    }
    appDrawerVisible.value = true
  }

  function openEntryDialog(item?: Api.SystemManage.AppHostBindingItem) {
    resetEntryForm()
    if (item) {
      editingEntryId.value = item.id || ''
      entryForm.id = item.id || ''
      entryForm.app_key = item.appKey || selectedAppKey.value
      entryForm.match_type = normalizeMatchType(item.matchType)
      entryForm.host = item.host || ''
      entryForm.path_pattern = item.pathPattern || ''
      entryForm.priority = item.priority || 0
      entryForm.default_menu_space_key = resolveSpaceKey(
        item.defaultMenuSpaceKey,
        item.defaultMenuSpaceKey,
        selectedAppRecord.value?.defaultMenuSpaceKey
      )
      entryForm.description = item.description || ''
      entryForm.is_primary = Boolean(item.isPrimary)
      entryForm.status = item.status || 'normal'
      entryForm.meta = item.meta || {}
    }
    entryDialogVisible.value = true
  }

  function openSpaceEntryDialog(item?: Api.SystemManage.MenuSpaceEntryBindingItem) {
    resetSpaceEntryForm()
    if (item) {
      editingSpaceEntryId.value = item.id || ''
      spaceEntryForm.id = item.id || ''
      spaceEntryForm.app_key = item.appKey || selectedAppKey.value
      spaceEntryForm.menu_space_key = item.menuSpaceKey || ''
      spaceEntryForm.match_type = normalizeMatchType(item.matchType)
      spaceEntryForm.host = item.host || ''
      spaceEntryForm.path_pattern = item.pathPattern || ''
      spaceEntryForm.priority = item.priority || 0
      spaceEntryForm.description = item.description || ''
      spaceEntryForm.is_primary = Boolean(item.isPrimary)
      spaceEntryForm.status = item.status || 'normal'
      spaceEntryForm.meta = item.meta || {}
    }
    spaceEntryDialogVisible.value = true
  }

  async function saveApp() {
    clearFieldErrors(appFieldErrors)
    const valid = await appFormRef.value?.validate().catch(() => false)
    if (!valid) return
    if (!appForm.app_key.trim()) {
      appFieldErrors.app_key = '请输入应用标识'
      return
    }
    if (!appForm.name.trim()) {
      appFieldErrors.name = '请输入应用名称'
      return
    }
    let capabilities: Record<string, any>
    try {
      capabilities = omitDeprecatedCapabilityFields(
        patchAuthCapability(
          {
            ...appCapabilitiesBase.value,
            ...parseCapabilitiesText()
          },
          {
            ssoMode: appAuthSsoMode.value,
            loginUiMode: appAuthLoginUiMode.value,
            loginPageKey: `${appAuthLoginPageKey.value || ''}`.trim() || 'default'
          }
        )
      )
    } catch (error: any) {
      ElMessage.warning(error?.message || '应用治理配置格式错误')
      return
    }
    savingApp.value = true
    try {
      const payload: Api.SystemManage.AppSaveParams = {
        ...appForm,
        app_key: appForm.app_key.trim(),
        name: appForm.name.trim(),
        description: appForm.description?.trim() || '',
        space_mode: appForm.space_mode === 'multi' ? 'multi' : 'single',
        auth_mode: appForm.auth_mode || 'inherit_host',
        frontend_entry_url: `${appForm.frontend_entry_url || ''}`.trim(),
        backend_entry_url: `${appForm.backend_entry_url || ''}`.trim(),
        health_check_url: `${appForm.health_check_url || ''}`.trim(),
        manifest_url: `${appForm.manifest_url || ''}`.trim(),
        runtime_version: `${appForm.runtime_version || ''}`.trim(),
        capabilities,
        meta: { ...appMetaBase.value }
      }
      const nextDefaultSpaceKey = resolveSpaceKey(appForm.default_menu_space_key)
      if (editingAppKey.value && nextDefaultSpaceKey) {
        payload.default_menu_space_key = nextDefaultSpaceKey
      } else {
        delete payload.default_menu_space_key
      }
      const saved = await fetchSaveApp({
        ...payload
      })
      ElMessage.success('应用已保存')
      appDrawerVisible.value = false
      await setManagedAppKey(saved.appKey)
      selectedAppKey.value = saved.appKey
      await loadData()
    } catch (error: any) {
      if (applyBackendFieldErrors(appFieldErrors, error)) return
      ElMessage.error(error?.message || '应用保存失败')
    } finally {
      savingApp.value = false
    }
  }

  function validateEntryForm(form: { match_type?: string; host?: string; path_pattern?: string }) {
    const mt = form.match_type || 'host_exact'
    const host = (form.host || '').trim()
    const path = (form.path_pattern || '').trim()
    if (['host_exact', 'host_suffix'].includes(mt) && !host) {
      return 'Host 不能为空'
    }
    if (mt === 'path_prefix' && !path) {
      return '路径模式不能为空'
    }
    if (mt === 'host_and_path' && (!host || !path)) {
      return 'host_and_path 类型必须同时填写 Host 和路径'
    }
    return ''
  }

  async function saveEntryBinding() {
    clearFieldErrors(entryFieldErrors)
    if (!selectedAppKey.value) {
      ElMessage.warning('请先选择应用')
      return
    }
    const err = validateEntryForm(entryForm)
    if (err) {
      // 按字段精准标注，而不是整体 toast
      if (err.startsWith('Host')) entryFieldErrors.host = err
      else if (err.startsWith('路径模式')) entryFieldErrors.path_pattern = err
      else {
        entryFieldErrors.host = err
        entryFieldErrors.path_pattern = err
      }
      return
    }
    if (
      !resolveSpaceKey(
        entryForm.default_menu_space_key,
        selectedAppRecord.value?.defaultMenuSpaceKey,
        selectedAppRecord.value?.defaultMenuSpaceKey
      )
    ) {
      entryFieldErrors.default_menu_space_key = '请选择或填写默认空间'
      return
    }
    savingHost.value = true
    try {
      await fetchSaveAppHostBinding({
        ...entryForm,
        app_key: selectedAppKey.value,
        host: (entryForm.host || '').trim().toLowerCase(),
        path_pattern: (entryForm.path_pattern || '').trim(),
        default_menu_space_key: resolveSpaceKey(
          entryForm.default_menu_space_key,
          selectedAppRecord.value?.defaultMenuSpaceKey,
          selectedAppRecord.value?.defaultMenuSpaceKey
        ),
        description: entryForm.description?.trim() || ''
      })
      ElMessage.success('入口绑定已保存')
      entryDialogVisible.value = false
      await loadSelectedAppContext(selectedAppKey.value)
    } catch (error: any) {
      if (applyBackendFieldErrors(entryFieldErrors, error)) return
      ElMessage.error(error?.message || '入口绑定保存失败')
    } finally {
      savingHost.value = false
    }
  }

  async function deleteEntry(item: Api.SystemManage.AppHostBindingItem) {
    if (!item.id) return
    try {
      await fetchDeleteAppHostBinding(item.id, selectedAppKey.value)
      ElMessage.success('已删除')
      await loadSelectedAppContext(selectedAppKey.value)
    } catch (error: any) {
      ElMessage.error(error?.message || '删除失败')
    }
  }

  async function saveSpaceEntryBinding() {
    clearFieldErrors(spaceEntryFieldErrors)
    if (!selectedAppKey.value) {
      ElMessage.warning('请先选择应用')
      return
    }
    if (!spaceEntryForm.menu_space_key) {
      spaceEntryFieldErrors.menu_space_key = '请选择目标菜单空间'
      return
    }
    const err = validateEntryForm(spaceEntryForm)
    if (err) {
      if (err.startsWith('Host')) spaceEntryFieldErrors.host = err
      else if (err.startsWith('路径模式')) spaceEntryFieldErrors.path_pattern = err
      else {
        spaceEntryFieldErrors.host = err
        spaceEntryFieldErrors.path_pattern = err
      }
      return
    }
    savingSpaceEntry.value = true
    try {
      await fetchSaveMenuSpaceEntryBinding({
        ...spaceEntryForm,
        app_key: selectedAppKey.value,
        host: (spaceEntryForm.host || '').trim().toLowerCase(),
        path_pattern: (spaceEntryForm.path_pattern || '').trim(),
        description: spaceEntryForm.description?.trim() || ''
      })
      ElMessage.success('菜单空间入口绑定已保存')
      spaceEntryDialogVisible.value = false
      await loadSelectedAppContext(selectedAppKey.value)
    } catch (error: any) {
      if (applyBackendFieldErrors(spaceEntryFieldErrors, error)) return
      ElMessage.error(error?.message || '菜单空间入口绑定保存失败')
    } finally {
      savingSpaceEntry.value = false
    }
  }

  async function deleteSpaceEntry(item: Api.SystemManage.MenuSpaceEntryBindingItem) {
    if (!item.id) return
    try {
      await fetchDeleteMenuSpaceEntryBinding(item.id, selectedAppKey.value)
      ElMessage.success('已删除')
      await loadSelectedAppContext(selectedAppKey.value)
    } catch (error: any) {
      ElMessage.error(error?.message || '删除失败')
    }
  }

  function selectApp(appKey: string) {
    if (!appKey || appKey === selectedAppKey.value) return
    loadSelectedAppContext(appKey).catch((error: any) => {
      ElMessage.error(error?.message || '切换应用失败')
    })
  }

  async function goToMenuManagement() {
    if (selectedAppKey.value) {
      await loadSelectedAppContext(selectedAppKey.value)
    }
    router.push({ path: '/system/menu' })
  }

  async function goToPageManagement() {
    if (selectedAppKey.value) {
      await loadSelectedAppContext(selectedAppKey.value)
    }
    router.push({ path: '/system/page' })
  }

  async function goToSpaceManagement(appKey?: string) {
    const targetKey = appKey || selectedAppKey.value
    if (targetKey) {
      try {
        await loadSelectedAppContext(targetKey)
      } catch (error: any) {
        ElMessage.error(error?.message || '切换应用失败')
        return
      }
    }
    router.push({ path: '/system/menu-space' })
  }

  onMounted(() => {
    loadData()
  })

  watch(targetAppKey, (value) => {
    if (value && value !== selectedAppKey.value) {
      selectedAppKey.value = value
    } else if (!value) {
      selectedAppKey.value = ''
    }
  })
</script>

<style scoped lang="scss">
  .app-manage-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  /* ── 抽屉表单整体 ─────────────────────────── */
  .app-drawer-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  /* ── 分组卡片 ────────────────────────────── */
  .app-form-card {
    border: 1px solid var(--el-border-color-light);
    border-radius: 14px;
    padding: 18px 20px 6px;
    background: #fff;
  }

  .app-form-card--last {
    margin-bottom: 8px;
  }

  .app-form-card__header {
    margin-bottom: 16px;
  }

  .app-form-card__title {
    display: block;
    font-size: 13px;
    font-weight: 700;
    color: var(--art-text-gray-900);
    margin-bottom: 4px;
  }

  .app-form-card__desc {
    display: block;
    font-size: 12px;
    line-height: 1.65;
    color: var(--art-text-gray-500);
  }

  /* ── 卡片内嵌子块（条件联动区） ──────────────── */
  .app-form-sub-card {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    padding: 14px 16px 8px;
    background: color-mix(in srgb, var(--el-color-primary-light-9) 60%, white);
    margin-top: 4px;
    margin-bottom: 18px;
  }

  .app-form-sub-card__label {
    font-size: 12px;
    font-weight: 600;
    color: var(--el-color-primary);
    margin-bottom: 10px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .app-form-sub-card__label::before {
    content: '';
    display: inline-block;
    width: 3px;
    height: 12px;
    border-radius: 2px;
    background: var(--el-color-primary);
  }

  /* ── 提示条 ──────────────────────────────── */
  .app-form-tip {
    margin: -4px 0 18px;
    padding: 10px 12px;
    border-radius: 8px;
    background: color-mix(in srgb, var(--art-gray-100) 80%, white);
    color: var(--art-text-gray-600);
    font-size: 12px;
    line-height: 1.65;

    code {
      padding: 1px 5px;
      border-radius: 4px;
      background: var(--art-gray-200);
      font-family: 'JetBrains Mono', Consolas, monospace;
      font-size: 11px;
      color: var(--art-text-gray-800);
    }
  }

  /* ── Switch 行 ───────────────────────────── */
  .app-form-switch-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .app-form-switch-label {
    font-size: 12px;
    color: var(--art-text-gray-500);
    line-height: 1.4;
  }

  /* ── 原有 hint（保留兼容） ──────────────────── */
  .app-form-hint {
    margin: 4px 0 14px;
    color: var(--art-text-gray-600);
    font-size: 12px;
    line-height: 1.6;
  }

  .app-form-section {
    margin: 6px 0 12px;
  }

  .app-form-section__title {
    color: var(--art-text-gray-900);
    font-size: 13px;
    font-weight: 700;
  }

  .app-form-section__desc {
    margin-top: 4px;
    color: var(--art-text-gray-500);
    font-size: 12px;
    line-height: 1.6;
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

  .app-manage-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  /* ── 面板内切换导航 ───────────────────────── */
  .app-panel-nav {
    display: flex;
    gap: 4px;
    margin: 12px 0 0;
    padding: 4px;
    border-radius: 12px;
    background: var(--art-gray-100);
  }

  .app-panel-nav__item {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 7px 10px;
    border: none;
    border-radius: 9px;
    background: transparent;
    color: var(--art-text-gray-600);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.18s, color 0.18s;
    white-space: nowrap;

    &:hover {
      background: color-mix(in srgb, var(--art-gray-200) 80%, white);
      color: var(--art-text-gray-900);
    }

    &.is-active {
      background: #fff;
      color: var(--art-text-gray-900);
      font-weight: 600;
      box-shadow: 0 1px 4px rgb(0 0 0 / 8%);
    }
  }

  .app-panel-nav__badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    border-radius: 999px;
    background: var(--art-gray-200);
    color: var(--art-text-gray-600);
    font-size: 11px;
    font-weight: 600;
  }

  .app-panel-nav__item.is-active .app-panel-nav__badge {
    background: color-mix(in srgb, var(--art-primary) 15%, white);
    color: var(--art-primary);
  }

  /* ── Tab 内容区 ──────────────────────────── */
  .app-tab-content {
    margin-top: 14px;
  }

  .app-tab-header {
    margin-bottom: 14px;
  }

  .app-tab-header__title {
    font-size: 13px;
    font-weight: 700;
    color: var(--art-text-gray-900);
    margin-bottom: 4px;
  }

  .app-tab-header__desc {
    font-size: 12px;
    line-height: 1.65;
    color: var(--art-text-gray-500);
  }

  /* ── 配置总览 Full-width checks ─────────── */
  .app-governance-checks--full {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
    margin-top: 12px;
  }

  /* ── 快速预检 Full-width ──────────────────  */
  .app-preview-list--full {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  /* ── 入口规则 card 容器 ──────────────────── */
  .app-rule-card {
    border: 1px solid var(--el-border-color-light);
    border-radius: 14px;
    overflow: hidden;
    margin-top: 14px;
    background: #fff;
  }

  .app-rule-card__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 16px;
    background: color-mix(in srgb, var(--art-gray-100) 60%, white);
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .app-rule-card__head-left {
    display: flex;
    align-items: flex-start;
    gap: 10px;
  }

  .app-rule-card__level {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 8px;
    background: color-mix(in srgb, var(--el-color-primary) 14%, white);
    color: var(--el-color-primary);
    font-size: 11px;
    font-weight: 700;
    flex-shrink: 0;
    margin-top: 1px;
  }

  .app-rule-card__level--l2 {
    background: color-mix(in srgb, var(--el-color-warning) 14%, white);
    color: var(--el-color-warning-dark-2);
  }

  .app-rule-card__title {
    font-size: 13px;
    font-weight: 700;
    color: var(--art-text-gray-900);
    line-height: 1.4;
  }

  .app-rule-card__desc {
    margin-top: 3px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--art-text-gray-500);
  }

  .app-rule-card__empty {
    padding: 16px;
    font-size: 13px;
    color: var(--art-text-gray-400);
    line-height: 1.6;
  }

  /* ── 规则行 ──────────────────────────────── */
  .app-rule-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-top: 1px solid var(--el-border-color-lighter);
    cursor: pointer;
    transition: background 0.15s;

    &:hover {
      background: color-mix(in srgb, var(--el-color-primary-light-9) 50%, white);
    }
  }

  .app-rule-row__body {
    flex: 1;
    min-width: 0;
  }

  .app-rule-row__title-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 7px;
  }

  .app-rule-row__host {
    font-size: 15px;
    font-weight: 700;
    color: var(--art-text-gray-900);
  }

  .app-rule-row__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 5px;
    font-size: 12px;
    color: var(--art-text-gray-500);
  }

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
    transition:
      border-color 0.2s ease,
      box-shadow 0.2s ease,
      transform 0.2s ease;
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

  .app-overview__summary {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    font-size: 13px;
    color: var(--art-text-gray-600);

    strong {
      color: var(--art-text-gray-900);
      font-weight: 600;
    }
  }

  .app-overview__actions,
  .app-space-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
  }

  .app-governance-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
    margin-bottom: 16px;
  }

  .app-governance-card {
    border: 1px solid color-mix(in srgb, var(--art-border-color) 82%, white);
    border-radius: 16px;
    background: color-mix(in srgb, var(--art-main-bg-color) 92%, white);
    padding: 16px;
    min-height: 100%;
  }

  .app-governance-card__header {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: flex-start;
  }

  .app-governance-card__title {
    color: var(--art-text-gray-900);
    font-size: 14px;
    font-weight: 700;
  }

  .app-governance-card__desc {
    margin-top: 4px;
    color: var(--art-text-gray-500);
    font-size: 12px;
    line-height: 1.6;
  }

  .app-governance-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 12px;
    margin-top: 12px;
    color: var(--art-text-gray-600);
    font-size: 12px;
  }

  .app-governance-checks,
  .app-preview-list,
  .app-sensitive-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 12px;
  }

  .app-governance-check,
  .app-preview-item,
  .app-sensitive-item,
  .app-sensitive-note {
    border-radius: 12px;
    padding: 12px;
    background: color-mix(in srgb, var(--art-gray-100) 70%, white);
  }

  .app-governance-check.is-success {
    background: color-mix(in srgb, var(--el-color-success-light-9) 72%, white);
  }

  .app-governance-check.is-warning {
    background: color-mix(in srgb, var(--el-color-warning-light-9) 74%, white);
  }

  .app-governance-check__title,
  .app-preview-item__label,
  .app-sensitive-item__title,
  .app-sensitive-note__title {
    color: var(--art-text-gray-900);
    font-size: 13px;
    font-weight: 600;
  }

  .app-governance-check__text,
  .app-preview-item__hint,
  .app-sensitive-item__text,
  .app-sensitive-note__text {
    margin-top: 4px;
    color: var(--art-text-gray-600);
    font-size: 12px;
    line-height: 1.6;
  }

  .app-preview-item__value {
    display: inline-block;
    margin-top: 6px;
    color: var(--art-text-gray-900);
    font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
    font-size: 12px;
    line-height: 1.5;
    word-break: break-all;
  }

  .app-capability-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }

  .app-capability-pill {
    padding: 6px 10px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--art-primary) 10%, white);
    color: var(--art-text-gray-700);
    font-size: 12px;
  }

  .app-capability-pill.is-soft {
    background: var(--art-gray-100);
    color: var(--art-text-gray-500);
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

  .app-config-entry,
  .app-config-entry__summary,
  .app-config-entry__actions,
  .app-config-stack,
  .app-config-advanced__actions,
  .app-structured-card__actions {
    display: flex;
  }

  .app-config-entry {
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    padding: 14px 16px;
    border: 1px solid color-mix(in srgb, var(--art-border-color) 86%, white);
    border-radius: 14px;
    background: color-mix(in srgb, var(--art-main-bg-color) 94%, white);
  }

  .app-config-entry__summary,
  .app-config-stack {
    flex-direction: column;
    gap: 10px;
  }

  .app-config-entry__summary {
    flex: 1;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    flex-direction: row;
  }

  .app-config-entry__actions {
    flex: 0 0 auto;
    gap: 10px;
  }

  .app-config-entry__empty,
  .app-config-panel__desc {
    color: var(--art-text-gray-500);
    font-size: 12px;
    line-height: 1.6;
  }

  .app-config-entry__error {
    margin-top: 8px;
    color: var(--el-color-warning-dark-2);
    font-size: 12px;
    line-height: 1.6;
  }

  .app-config-dialog__intro {
    margin-bottom: 12px;
    color: var(--art-text-gray-600);
    font-size: 13px;
    line-height: 1.7;
  }

  .app-config-panel {
    padding: 4px 2px 0;
  }

  .app-config-panel__title {
    color: var(--art-text-gray-900);
    font-size: 14px;
    font-weight: 700;
  }

  .app-inline-row,
  .app-structured-kv {
    display: grid;
    gap: 10px;
    align-items: flex-start;
  }

  .app-inline-row {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .app-structured-card {
    border: 1px solid color-mix(in srgb, var(--art-border-color) 86%, white);
    border-radius: 14px;
    padding: 14px;
    background: color-mix(in srgb, var(--art-main-bg-color) 96%, white);
  }

  .app-structured-card__header {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: flex-start;
    margin-bottom: 12px;
  }

  .app-structured-card__body,
  .app-structured-card__actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .app-structured-kv {
    grid-template-columns: minmax(0, 1.2fr) 120px minmax(0, 1.4fr) auto;
  }

  .app-config-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
  }

  .app-config-advanced {
    margin-top: 16px;
    border: 1px dashed color-mix(in srgb, var(--art-border-color) 88%, white);
    border-radius: 14px;
    background: color-mix(in srgb, var(--art-main-bg-color) 96%, white);
  }

  .app-config-advanced > summary {
    cursor: pointer;
    list-style: none;
    padding: 12px 14px;
    color: var(--art-text-gray-700);
    font-size: 13px;
    font-weight: 600;
  }

  .app-config-advanced > summary::-webkit-details-marker {
    display: none;
  }

  .app-config-advanced__body {
    padding: 0 14px 14px;
  }

  .app-config-advanced__alert {
    margin-bottom: 12px;
  }

  .app-config-advanced__actions {
    justify-content: flex-end;
    gap: 10px;
    margin-top: 12px;
  }

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  :deep(.el-drawer) {
    min-width: 680px;
  }

  @media (max-width: 1200px) {
    .app-manage-board {
      grid-template-columns: 1fr;
    }

    .app-governance-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .app-drawer-grid {
      grid-template-columns: 1fr;
    }

    .app-config-grid,
    .app-structured-kv {
      grid-template-columns: 1fr;
    }

    .app-config-entry,
    .app-inline-row {
      grid-template-columns: 1fr;
      flex-direction: column;
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
