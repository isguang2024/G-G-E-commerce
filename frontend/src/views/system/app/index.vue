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
                <span>默认 {{ displaySpaceLabel(item.defaultSpaceKey) }}</span>
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

        <div v-if="selectedAppRecord" class="app-overview">
          <div class="app-overview__summary">
            <span
              >主 Host <strong>{{ selectedAppRecord.primaryHost || '未设置' }}</strong></span
            >
            <span>·</span>
            <span
              >默认空间
              <strong>{{ displaySpaceLabel(selectedAppRecord.defaultSpaceKey) }}</strong></span
            >
            <span>·</span>
            <span
              >前端入口
              <strong>{{ selectedAppRecord.frontendEntryUrl || '继承当前地址' }}</strong></span
            >
          </div>
          <div class="app-overview__actions">
            <ElButton text @click="goToMenuManagement">菜单管理</ElButton>
            <ElButton text @click="goToPageManagement">页面管理</ElButton>
            <ElButton text @click="goToSpaceManagement()">高级空间配置</ElButton>
          </div>
        </div>

        <div v-if="selectedAppRecord" class="app-governance-grid">
          <section class="app-governance-card">
            <div class="app-governance-card__header">
              <div>
                <div class="app-governance-card__title">注册中心字段总览</div>
                <div class="app-governance-card__desc"
                  >把入口、认证、运行能力与治理约束拆开看，避免都堆在 JSON 里。</div
                >
              </div>
              <ElTag size="small" effect="plain" :type="appReadinessTagType">
                {{ appReadinessLabel }}
              </ElTag>
            </div>
            <div class="app-governance-meta">
              <span
                >空间模式 {{ selectedAppRecord.spaceMode === 'multi' ? '多空间' : '单空间' }}</span
              >
              <span>认证 {{ authModeLabel(selectedAppRecord.authMode) }}</span>
              <span>入口绑定 {{ hostBindings.length || 0 }}</span>
              <span>空间入口 {{ spaceEntryBindings.length || 0 }}</span>
            </div>
            <div class="app-governance-checks">
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
          </section>

          <section class="app-governance-card">
            <div class="app-governance-card__header">
              <div>
                <div class="app-governance-card__title">接入预检查与本地预演</div>
                <div class="app-governance-card__desc"
                  >后端会聚合注册中心、入口绑定和认证配置输出 dry-run
                  结果，前端只负责展示，不再本地拼接结论。</div
                >
              </div>
            </div>
            <div class="app-preview-list">
              <div v-for="item in appDryRunPreview" :key="item.label" class="app-preview-item">
                <div class="app-preview-item__label">{{ item.label }}</div>
                <code class="app-preview-item__value">{{ item.value }}</code>
                <div class="app-preview-item__hint">{{ item.hint }}</div>
              </div>
            </div>
          </section>

          <section class="app-governance-card">
            <div class="app-governance-card__header">
              <div>
                <div class="app-governance-card__title">运行能力与敏感配置</div>
                <div class="app-governance-card__desc"
                  >能力声明适合描述
                  runtime/navigation/integration；密钥、证书、回调域名不要直接写在公开表单。</div
                >
              </div>
            </div>
            <div class="app-capability-pills">
              <span v-for="item in capabilityHighlights" :key="item" class="app-capability-pill">
                {{ item }}
              </span>
              <span v-if="!capabilityHighlights.length" class="app-capability-pill is-soft">
                暂无显式能力声明
              </span>
            </div>
            <div class="app-capability-pills app-capability-pills--meta">
              <span v-for="item in appMetaSections" :key="item" class="app-capability-pill is-soft">
                {{ item }}
              </span>
            </div>
            <div class="app-sensitive-list">
              <div
                v-for="item in sensitiveConfigHints"
                :key="item.title"
                class="app-sensitive-item"
              >
                <div class="app-sensitive-item__title">{{ item.title }}</div>
                <div class="app-sensitive-item__text">{{ item.text }}</div>
              </div>
            </div>
          </section>
        </div>

        <div class="app-binding-section">
          <div class="app-binding-section__header">
            <div class="app-binding-section__title">Level 1 · APP 入口</div>
            <ElButton size="small" type="primary" link @click="openEntryDialog()">+ 新增</ElButton>
          </div>
          <div class="app-binding-list">
            <div v-if="!hostBindings.length" class="app-manage-empty">
              暂无入口规则，未命中时系统将退回默认 App。
            </div>
            <div
              v-for="item in hostBindings"
              :key="item.id || item.host + item.pathPattern"
              class="app-binding-item"
            >
              <div class="app-binding-item__main" @click="openEntryDialog(item)">
                <div class="app-binding-item__title-row">
                  <ElTag size="small" effect="plain">{{ matchTypeLabel(item.matchType) }}</ElTag>
                  <span class="app-binding-item__host">{{ describeEntryRule(item) }}</span>
                  <ElTag v-if="item.isPrimary" size="small" type="success" effect="plain">主</ElTag>
                  <ElTag
                    size="small"
                    :type="item.status === 'normal' ? 'info' : 'danger'"
                    effect="plain"
                  >
                    {{ item.status === 'normal' ? '启用' : '停用' }}
                  </ElTag>
                </div>
                <div class="app-binding-item__meta">
                  <span
                    >默认空间
                    {{
                      displaySpaceLabel(item.defaultSpaceKey, selectedAppRecord?.defaultSpaceKey)
                    }}</span
                  >
                  <span>优先级 {{ item.priority || 0 }}</span>
                  <span v-if="item.description">{{ item.description }}</span>
                </div>
              </div>
              <ElButton text type="danger" size="small" @click.stop="deleteEntry(item)"
                >删除</ElButton
              >
            </div>
          </div>
        </div>

        <div v-if="isMultiSpaceApp" class="app-binding-section">
          <div class="app-binding-section__header">
            <div class="app-binding-section__title">Level 2 · 菜单空间入口</div>
            <ElButton
              size="small"
              type="primary"
              link
              :disabled="!spaces.length"
              @click="openSpaceEntryDialog()"
            >
              + 新增
            </ElButton>
          </div>
          <div class="app-binding-list">
            <div v-if="!spaceEntryBindings.length" class="app-manage-empty">
              暂无菜单空间入口规则，未命中时按 APP 默认空间进入。
            </div>
            <div
              v-for="item in spaceEntryBindings"
              :key="item.id || item.spaceKey + item.host + item.pathPattern"
              class="app-binding-item"
            >
              <div class="app-binding-item__main" @click="openSpaceEntryDialog(item)">
                <div class="app-binding-item__title-row">
                  <ElTag size="small" effect="plain">{{ matchTypeLabel(item.matchType) }}</ElTag>
                  <span class="app-binding-item__host">{{ describeEntryRule(item) }}</span>
                  <ElTag size="small" type="warning" effect="plain"
                    >→ {{ item.spaceName || item.spaceKey }}</ElTag
                  >
                  <ElTag
                    size="small"
                    :type="item.status === 'normal' ? 'info' : 'danger'"
                    effect="plain"
                  >
                    {{ item.status === 'normal' ? '启用' : '停用' }}
                  </ElTag>
                </div>
                <div class="app-binding-item__meta">
                  <span>优先级 {{ item.priority || 0 }}</span>
                  <span v-if="item.description">{{ item.description }}</span>
                </div>
              </div>
              <ElButton text type="danger" size="small" @click.stop="deleteSpaceEntry(item)"
                >删除</ElButton
              >
            </div>
          </div>
        </div>

        <div class="app-space-pills">
          <span class="app-space-pills__label">空间配置</span>
          <span v-for="item in spaces" :key="item.spaceKey" class="app-space-pill">
            {{ item.name }} · {{ item.spaceKey }}
          </span>
          <span v-if="!spaces.length" class="app-space-pill is-soft">当前 App 暂无空间配置</span>
        </div>
      </ElCard>
    </section>

    <ElDrawer v-model="appDrawerVisible" :title="appDrawerTitle" size="520px" destroy-on-close>
      <ElForm :model="appForm" label-position="top">
        <div class="app-form-section">
          <div class="app-form-section__title">基础标识</div>
          <div class="app-form-section__desc"
            >面向注册中心的稳定标识。`app_key` 创建后即作为跨端导航、权限和 host 解析主键。</div
          >
        </div>
        <ElFormItem label="应用名称">
          <ElInput v-model="appForm.name" placeholder="例如 平台管理后台" />
        </ElFormItem>
        <ElFormItem label="应用标识">
          <ElInput
            v-model="appForm.app_key"
            :disabled="Boolean(editingAppKey)"
            placeholder="例如 platform-admin"
          />
        </ElFormItem>
        <div v-if="!editingAppKey" class="app-form-hint">
          新建 App 时系统会自动创建当前 App 自己的默认空间 `default`，无需手动选择。
        </div>
        <div class="app-form-section">
          <div class="app-form-section__title">空间与认证</div>
          <div class="app-form-section__desc"
            >先决定多空间/单空间，再决定登录如何发生。登录方式决定是否需要单独的认证中心回跳治理。</div
          >
        </div>
        <ElFormItem v-if="editingAppKey" label="默认空间">
          <ElSelect
            v-model="appForm.default_space_key"
            filterable
            allow-create
            default-first-option
            style="width: 100%"
          >
            <ElOption
              v-for="item in spaces"
              :key="item.spaceKey"
              :label="`${item.name} · ${item.spaceKey}`"
              :value="item.spaceKey"
            />
          </ElSelect>
        </ElFormItem>
        <div v-if="editingAppKey" class="app-form-hint">
          这里控制 APP 首次解析落到哪个空间；空间内首页请到“高级空间配置”里调整
          <code>default_home_path</code>。
        </div>
        <ElFormItem label="空间模式">
          <ElSelect v-model="appForm.space_mode" style="width: 100%">
            <ElOption label="单空间" value="single" />
            <ElOption label="多空间" value="multi" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="认证模式">
          <ElSelect v-model="appForm.auth_mode" style="width: 100%">
            <ElOption label="继承当前 Host" value="inherit_host" />
            <ElOption label="共享 Cookie" value="shared_cookie" />
            <ElOption label="独立认证中心" value="centralized_login" />
          </ElSelect>
        </ElFormItem>
        <div class="app-form-section">
          <div class="app-form-section__title">运行入口与部署探针</div>
          <div class="app-form-section__desc"
            >这里放可公开的入口地址和健康检查地址。环境差异、feature flag 和部署 profile
            建议继续放进 `meta` 的环境分组，而不是散落在 description。</div
          >
        </div>
        <div class="app-drawer-grid">
          <ElFormItem label="前端入口地址">
            <ElInput
              v-model="appForm.frontend_entry_url"
              placeholder="例如 /account 或 https://account.example.com"
            />
          </ElFormItem>
          <ElFormItem label="后端入口地址">
            <ElInput
              v-model="appForm.backend_entry_url"
              placeholder="例如 /api 或 https://api.example.com"
            />
          </ElFormItem>
          <ElFormItem label="Manifest 地址">
            <ElInput
              v-model="appForm.manifest_url"
              placeholder="例如 https://cdn.example.com/app/manifest.json"
            />
          </ElFormItem>
          <ElFormItem label="运行版本">
            <ElInput v-model="appForm.runtime_version" placeholder="例如 2026.04.12-rc1" />
          </ElFormItem>
        </div>
        <ElFormItem label="健康检查地址">
          <ElInput v-model="appForm.health_check_url" placeholder="例如 /health" />
        </ElFormItem>
        <div class="app-form-section">
          <div class="app-form-section__title">运行能力声明</div>
          <div class="app-form-section__desc"
            >把前端能力、集成方式、feature flag 读取能力写成结构化 JSON，避免靠命名约定猜测。</div
          >
        </div>
        <ElFormItem label="运行能力声明">
          <ElInput
            v-model="appCapabilitiesText"
            type="textarea"
            :rows="8"
            placeholder='例如 {"runtime":{"supports_worktab":true}}'
          />
          <div class="app-form-hint">
            使用 JSON 对象描述 routing/runtime/navigation/integration 能力；这里不重复填写
            `space_mode`、`auth_mode` 之类顶层字段。
          </div>
        </ElFormItem>
        <div class="app-form-section">
          <div class="app-form-section__title">多环境与 Feature Flag</div>
          <div class="app-form-section__desc"
            >按 APP 维度维护环境 profile 与 feature flag。这里统一写结构化配置，避免同一应用的
            dev/test/prod 规则散落在说明字段。</div
          >
        </div>
        <ElFormItem label="环境配置（meta.env_profiles）">
          <ElInput
            v-model="appEnvProfilesText"
            type="textarea"
            :rows="6"
            placeholder='例如 {"dev":{"frontend_base":"http://127.0.0.1:5174"},"prod":{"frontend_base":"https://app.example.com"}}'
          />
          <div class="app-form-hint">
            推荐按环境名分组，记录入口、profile、外部系统键名等非敏感配置。
          </div>
        </ElFormItem>
        <ElFormItem label="Feature Flag（meta.feature_flags）">
          <ElInput
            v-model="appFeatureFlagsText"
            type="textarea"
            :rows="6"
            placeholder='例如 {"shared_cookie":{"default":false,"prod":true},"remote_manifest":{"default":false}}'
          />
          <div class="app-form-hint">
            推荐按 flag 名存储默认值与环境覆盖，避免只在前端代码里写死开关判断。
          </div>
        </ElFormItem>
        <div class="app-form-section">
          <div class="app-form-section__title">治理补充</div>
          <div class="app-form-section__desc"
            >说明里写产品边界；敏感配置只记录键名、归属和托管方式，不直接录入明文。</div
          >
        </div>
        <ElFormItem label="敏感配置引用（meta.sensitive_config）">
          <ElInput
            v-model="appSensitiveConfigText"
            type="textarea"
            :rows="7"
            placeholder='例如 {"secret_refs":["oidc/client-secret"],"certificate_refs":["gateway/tls-cert"],"callback_domains":["https://account.example.com/callback"]}'
          />
          <div class="app-form-hint">
            这里写引用键、证书别名、回调域名白名单与 issuer 说明，不写真实密钥内容。
          </div>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput
            v-model="appForm.description"
            type="textarea"
            :rows="3"
            placeholder="说明这个 App 面向哪个站点或后台产品"
          />
        </ElFormItem>
        <div class="app-sensitive-note">
          <div class="app-sensitive-note__title">敏感配置约束</div>
          <div class="app-sensitive-note__text">
            密钥、证书、回调域名白名单、第三方 client secret
            不要直接填在表单明文里；这里只保留引用键、用途和托管位置，实际值交给环境变量或密钥系统。
          </div>
        </div>
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

    <ElDialog
      v-model="entryDialogVisible"
      :title="entryDialogTitle"
      width="560px"
      destroy-on-close
      append-to-body
    >
      <ElForm :model="entryForm" label-position="top">
        <ElFormItem label="匹配类型">
          <ElRadioGroup v-model="entryForm.match_type">
            <ElRadioButton value="host_exact">精确域名</ElRadioButton>
            <ElRadioButton value="host_suffix">子域名通配</ElRadioButton>
            <ElRadioButton value="path_prefix">路径模式</ElRadioButton>
            <ElRadioButton value="host_and_path">域名+路径</ElRadioButton>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem v-if="entryNeedsHost" label="Host">
          <ElInput v-model="entryForm.host" :placeholder="entryHostPlaceholder" />
        </ElFormItem>
        <ElFormItem v-if="entryNeedsPath" label="路径模式">
          <ElInput v-model="entryForm.path_pattern" placeholder="例如 /admin/** 或 /shop/:id/**" />
          <div class="app-form-hint">
            支持
            <code>*</code>（单段通配）、<code>**</code>（多段通配）、<code>:name</code>（命名参数）
          </div>
        </ElFormItem>
        <ElFormItem label="默认空间">
          <ElSelect
            v-model="entryForm.default_space_key"
            filterable
            allow-create
            default-first-option
            style="width: 100%"
          >
            <ElOption
              v-for="item in spaces"
              :key="item.spaceKey"
              :label="`${item.name} · ${item.spaceKey}`"
              :value="item.spaceKey"
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
      <ElForm :model="spaceEntryForm" label-position="top">
        <ElFormItem label="目标菜单空间">
          <ElSelect v-model="spaceEntryForm.space_key" filterable style="width: 100%">
            <ElOption
              v-for="item in spaces"
              :key="item.spaceKey"
              :label="`${item.name} · ${item.spaceKey}`"
              :value="item.spaceKey"
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
        <ElFormItem v-if="spaceEntryNeedsHost" label="Host">
          <ElInput v-model="spaceEntryForm.host" :placeholder="spaceEntryHostPlaceholder" />
        </ElFormItem>
        <ElFormItem v-if="spaceEntryNeedsPath" label="路径模式">
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
  import { ElMessage } from 'element-plus'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
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
  } from '@/api/system-manage'

  defineOptions({ name: 'AppManage' })

  const router = useRouter()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()
  const managedAppMissingText = '请先从 App 列表选择要管理的应用'
  const loading = ref(false)
  const loadError = ref('')
  const savingApp = ref(false)
  const savingHost = ref(false)
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
  const savingSpaceEntry = ref(false)
  const editingAppKey = ref('')
  const editingEntryId = ref('')
  const editingSpaceEntryId = ref('')

  const appForm = reactive<Api.SystemManage.AppSaveParams>({
    app_key: '',
    name: '',
    description: '',
    space_mode: 'single',
    default_space_key: '',
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
  const appEnvProfilesText = ref('{}')
  const appFeatureFlagsText = ref('{}')
  const appSensitiveConfigText = ref('{}')
  const appMetaBase = ref<Record<string, unknown>>({})
  type MatchType = Api.SystemManage.AppHostBindingSaveParams['match_type']

  const entryForm = reactive<Api.SystemManage.AppHostBindingSaveParams>({
    id: '',
    app_key: '',
    match_type: 'host_exact',
    host: '',
    path_pattern: '',
    priority: 0,
    default_space_key: '',
    description: '',
    is_primary: false,
    status: 'normal',
    meta: {}
  })

  const spaceEntryForm = reactive<Api.SystemManage.MenuSpaceEntryBindingSaveParams>({
    id: '',
    app_key: '',
    space_key: '',
    match_type: 'host_exact',
    host: '',
    path_pattern: '',
    priority: 0,
    description: '',
    is_primary: false,
    status: 'normal',
    meta: {}
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
    const hasRuntimeCapabilities = Boolean(
      record.capabilities && Object.keys(record.capabilities).length > 0
    )
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
            ? '当前为独立认证中心模式，建议把回调域名白名单和 issuer 说明写入受控配置，而不是公开表单。'
            : `当前为 ${authModeLabel(authMode)} 模式，登录态可沿用现有主站协同链路。`
      },
      {
        key: 'capabilities',
        title: '能力声明',
        level: hasRuntimeCapabilities ? 'success' : 'info',
        text: hasRuntimeCapabilities
          ? `已声明 ${Object.keys(record.capabilities || {}).length} 个一级能力分组。`
          : '尚未声明 runtime / navigation / integration 能力，后续接入和 feature flag 只能靠约定。'
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
    const source = selectedAppRecord.value?.capabilities || {}
    const sections = Object.entries(source)
    return sections.slice(0, 8).map(([key, value]) => {
      if (value && typeof value === 'object' && !Array.isArray(value)) {
        return `${key} · ${Object.keys(value as Record<string, any>).length} 项`
      }
      if (Array.isArray(value)) {
        return `${key} · ${value.length} 项`
      }
      return `${key} · ${String(value)}`
    })
  })
  const appMetaSections = computed(() => {
    const meta = (selectedAppRecord.value?.meta || {}) as Record<string, any>
    const envProfiles = meta.env_profiles
    const featureFlags = meta.feature_flags
    const sensitiveConfig = meta.sensitive_config
    const envCount =
      envProfiles && typeof envProfiles === 'object' && !Array.isArray(envProfiles)
        ? Object.keys(envProfiles).length
        : 0
    const flagCount =
      featureFlags && typeof featureFlags === 'object' && !Array.isArray(featureFlags)
        ? Object.keys(featureFlags).length
        : 0
    const sensitiveCount =
      sensitiveConfig && typeof sensitiveConfig === 'object' && !Array.isArray(sensitiveConfig)
        ? Object.keys(sensitiveConfig).length
        : 0
    return [
      `环境 ${envCount || 0} 组`,
      `Flag ${flagCount || 0} 项`,
      `敏感治理 ${sensitiveCount || 0} 组`
    ]
  })
  const sensitiveConfigHints = computed(() => {
    const authMode = `${selectedAppRecord.value?.authMode || 'inherit_host'}`.trim()
    return [
      {
        title: '环境配置 / Feature Flag',
        text: '建议在 meta 中按 env/profile 分组记录键名或 profile 名称，不在 description 里散写 dev/test/prod 差异。'
      },
      {
        title: '密钥与证书',
        text: '只保存引用键、用途和轮换归属，真实值交给环境变量、Vault 或部署系统。'
      },
      {
        title: '回调域名与白名单',
        text:
          authMode === 'centralized_login'
            ? '独立认证中心模式必须由受控配置维护 callback / logout 白名单，前端表单只展示治理说明。'
            : '即使不是独立认证，也建议把第三方回调域名白名单交给部署侧管理。'
      },
      {
        title: '当前治理覆盖度',
        text: appMetaSections.value.join(' / ')
      }
    ]
  })

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

  function formatCapabilitiesText(value?: Record<string, any>) {
    try {
      return JSON.stringify(value && Object.keys(value).length ? value : {}, null, 2)
    } catch {
      return '{}'
    }
  }

  function pickMetaObject(value: Record<string, any> | undefined, key: string) {
    const section = value?.[key]
    if (!section || Array.isArray(section) || typeof section !== 'object') {
      return {}
    }
    return section as Record<string, any>
  }

  function omitGovernedMetaSections(value?: Record<string, any>) {
    const source = value && typeof value === 'object' && !Array.isArray(value) ? value : {}
    const next = { ...source }
    delete next.env_profiles
    delete next.feature_flags
    delete next.sensitive_config
    return next
  }

  function parseCapabilitiesText() {
    const raw = `${appCapabilitiesText.value || ''}`.trim()
    if (!raw) {
      return {}
    }
    const parsed = JSON.parse(raw)
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      throw new Error('运行能力声明必须是 JSON 对象')
    }
    return parsed as Record<string, unknown>
  }

  function parseMetaSectionText(rawText: string, label: string) {
    const raw = `${rawText || ''}`.trim()
    if (!raw) {
      return {}
    }
    const parsed = JSON.parse(raw)
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      throw new Error(`${label}必须是 JSON 对象`)
    }
    return parsed as Record<string, unknown>
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
    appForm.default_space_key = ''
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
    appCapabilitiesText.value = '{}'
    appEnvProfilesText.value = '{}'
    appFeatureFlagsText.value = '{}'
    appSensitiveConfigText.value = '{}'
    appMetaBase.value = {}
  }

  function resetEntryForm() {
    editingEntryId.value = ''
    entryForm.id = ''
    entryForm.app_key = resolveAppKey(selectedAppKey.value)
    entryForm.match_type = 'host_exact'
    entryForm.host = ''
    entryForm.path_pattern = ''
    entryForm.priority = 0
    entryForm.default_space_key = resolveSpaceKey(selectedAppRecord.value?.defaultSpaceKey)
    entryForm.description = ''
    entryForm.is_primary = false
    entryForm.status = 'normal'
    entryForm.meta = {}
  }

  function resetSpaceEntryForm() {
    editingSpaceEntryId.value = ''
    spaceEntryForm.id = ''
    spaceEntryForm.app_key = resolveAppKey(selectedAppKey.value)
    spaceEntryForm.space_key = spaces.value[0]?.spaceKey || ''
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
      appForm.default_space_key = resolveSpaceKey(item.defaultSpaceKey)
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
      appCapabilitiesText.value = formatCapabilitiesText(item.capabilities)
      appMetaBase.value = omitGovernedMetaSections(item.meta || {})
      appEnvProfilesText.value = formatCapabilitiesText(
        pickMetaObject(item.meta || {}, 'env_profiles')
      )
      appFeatureFlagsText.value = formatCapabilitiesText(
        pickMetaObject(item.meta || {}, 'feature_flags')
      )
      appSensitiveConfigText.value = formatCapabilitiesText(
        pickMetaObject(item.meta || {}, 'sensitive_config')
      )
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
      entryForm.default_space_key = resolveSpaceKey(
        item.defaultSpaceKey,
        selectedAppRecord.value?.defaultSpaceKey
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
      spaceEntryForm.space_key = item.spaceKey || ''
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
    if (!appForm.app_key.trim()) {
      ElMessage.warning('请输入应用标识')
      return
    }
    if (!appForm.name.trim()) {
      ElMessage.warning('请输入应用名称')
      return
    }
    let capabilities: Record<string, any>
    let envProfiles: Record<string, any>
    let featureFlags: Record<string, any>
    let sensitiveConfig: Record<string, any>
    try {
      capabilities = parseCapabilitiesText()
      envProfiles = parseMetaSectionText(appEnvProfilesText.value, '环境配置')
      featureFlags = parseMetaSectionText(appFeatureFlagsText.value, 'Feature Flag')
      sensitiveConfig = parseMetaSectionText(appSensitiveConfigText.value, '敏感配置引用')
    } catch (error: any) {
      ElMessage.warning(error?.message || '应用治理配置格式错误')
      return
    }
    savingApp.value = true
    try {
      const meta: Record<string, any> = { ...appMetaBase.value }
      if (Object.keys(envProfiles).length) meta.env_profiles = envProfiles
      if (Object.keys(featureFlags).length) meta.feature_flags = featureFlags
      if (Object.keys(sensitiveConfig).length) meta.sensitive_config = sensitiveConfig
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
        meta
      }
      const nextDefaultSpaceKey = resolveSpaceKey(appForm.default_space_key)
      if (editingAppKey.value && nextDefaultSpaceKey) {
        payload.default_space_key = nextDefaultSpaceKey
      } else {
        delete payload.default_space_key
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
    if (!selectedAppKey.value) {
      ElMessage.warning('请先选择应用')
      return
    }
    const err = validateEntryForm(entryForm)
    if (err) {
      ElMessage.warning(err)
      return
    }
    if (!resolveSpaceKey(entryForm.default_space_key, selectedAppRecord.value?.defaultSpaceKey)) {
      ElMessage.warning('请选择或填写默认空间')
      return
    }
    savingHost.value = true
    try {
      await fetchSaveAppHostBinding({
        ...entryForm,
        app_key: selectedAppKey.value,
        host: (entryForm.host || '').trim().toLowerCase(),
        path_pattern: (entryForm.path_pattern || '').trim(),
        default_space_key: resolveSpaceKey(
          entryForm.default_space_key,
          selectedAppRecord.value?.defaultSpaceKey
        ),
        description: entryForm.description?.trim() || ''
      })
      ElMessage.success('入口绑定已保存')
      entryDialogVisible.value = false
      await loadSelectedAppContext(selectedAppKey.value)
    } catch (error: any) {
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
    if (!selectedAppKey.value) {
      ElMessage.warning('请先选择应用')
      return
    }
    if (!spaceEntryForm.space_key) {
      ElMessage.warning('请选择目标菜单空间')
      return
    }
    const err = validateEntryForm(spaceEntryForm)
    if (err) {
      ElMessage.warning(err)
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

  .app-form-hint {
    margin: -2px 0 14px;
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

  .app-binding-section {
    margin-top: 16px;
  }

  .app-binding-section__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .app-binding-section__title {
    font-size: 13px;
    font-weight: 600;
    color: var(--art-text-gray-700);
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

  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
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
