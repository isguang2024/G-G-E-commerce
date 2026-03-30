<template>
  <div class="message-manage-page art-full-height">
    <AdminWorkspaceHero
      :title="pageTitle"
      :description="pageDescription"
      :metrics="heroMetrics"
    >
      <div class="message-manage-hero__actions">
        <ElButton @click="loadOptions" :loading="loading" v-ripple>刷新配置</ElButton>
        <ElButton type="primary" @click="submitDispatch" :loading="submitting" :disabled="!canDispatch" v-ripple>发送消息</ElButton>
      </div>
    </AdminWorkspaceHero>

    <MessageWorkspaceNav :scope="props.scope" current="dispatch" />

    <ElAlert
      v-if="loadError"
      class="message-manage-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="message-manage-shell" v-loading="loading">
      <div class="message-manage-main art-card">
        <header class="message-manage-section__header">
          <div>
            <h3>发送配置</h3>
            <p>{{ senderScopeText }}</p>
          </div>
          <ElTag type="info" effect="plain">{{ senderScopeBadge }}</ElTag>
        </header>

        <ElForm ref="formRef" :model="form" label-position="top" class="message-manage-form">
          <section class="message-manage-block">
            <div class="message-manage-block__header">
              <h4>基础信息</h4>
              <p>先确定模板、发送人、消息类型和优先级，再进入发送对象和内容配置。</p>
            </div>
            <div class="message-manage-grid">
            <ElFormItem label="消息模板">
              <ElSelect
                v-model="form.template_id"
                clearable
                filterable
                placeholder="可选，选择后带入默认标题和富文本内容"
                @change="handleTemplateChange"
              >
                <ElOption
                  v-for="item in options.template_options"
                  :key="item.id"
                  :label="`${item.name} · ${item.owner_scope === 'team' ? '团队模板' : '平台模板'}`"
                  :value="item.id"
                />
              </ElSelect>
              <div class="field-hint">模板归属只决定模板来源，不改变本次实际的发送对象。</div>
            </ElFormItem>

            <ElFormItem label="发送人" prop="sender_id">
              <ElSelect v-model="form.sender_id" placeholder="请选择发送人">
                <ElOption
                  v-for="item in options.sender_options"
                  :key="item.id"
                  :label="item.is_default ? `${item.name} · 默认` : item.name"
                  :value="item.id"
                />
              </ElSelect>
              <div class="field-hint">{{ activeSenderDescription }}</div>
            </ElFormItem>

            <ElFormItem label="消息类型" prop="message_type">
              <ElRadioGroup v-model="form.message_type" class="message-manage-inline-options">
                <ElRadioButton v-for="item in messageTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>

            <ElFormItem label="消息优先级">
              <ElRadioGroup v-model="form.priority" class="message-manage-inline-options">
                <ElRadioButton v-for="item in priorityOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>
            </div>
          </section>

          <section class="message-manage-block">
            <div class="message-manage-block__header">
              <h4>接收对象</h4>
              <p>按主流消息公告后台习惯，先选对象，再补具体团队、用户或接收组。</p>
            </div>

            <ElFormItem label="发送对象" prop="audience_type">
              <ElSelect v-model="form.audience_type" @change="handleAudienceChange">
                <ElOption
                  v-for="item in options.audience_options"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
              <div class="field-hint">{{ activeAudienceDescription }}</div>
            </ElFormItem>

            <div class="message-manage-target-layout">
          <ElFormItem v-if="showTargetTeams" :label="targetTeamsLabel">
            <ElSelect
              v-if="!isTeamScope"
              v-model="form.target_tenant_ids"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择一个或多个团队"
            >
              <ElOption v-for="item in options.teams" :key="item.id" :label="item.name" :value="item.id" />
            </ElSelect>
            <div v-else class="message-manage-fixed-target">
              <strong>{{ options.current_tenant_name || '当前团队' }}</strong>
              <span>团队上下文只允许向当前团队发送。</span>
            </div>
            <div class="field-hint">
              {{ isTeamScope ? '发送对象会自动绑定到当前团队，无需再额外选择。' : '平台可选择多个目标团队，系统会按对象类型自动匹配成员。' }}
            </div>
          </ElFormItem>

          <ElFormItem v-if="showTargetUsers" :label="targetUsersLabel">
            <ElSelect
              v-model="form.target_user_ids"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择一个或多个用户"
            >
              <ElOption
                v-for="item in options.users"
                :key="item.id"
                :label="item.team_name ? `${item.display_name} · ${item.team_name}` : item.display_name"
                :value="item.id"
              />
            </ElSelect>
            <div class="field-hint">
              {{ isTeamScope ? '团队侧只会列出当前团队成员。' : '平台侧可以直接按用户维度精确发送。' }}
            </div>
          </ElFormItem>

          <ElFormItem
            v-if="showRecipientGroups"
            :label="form.audience_type === 'role' ? '角色接收组' : form.audience_type === 'feature_package' ? '功能包接收组' : '接收组'"
          >
            <ElSelect
              v-model="form.target_group_ids"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              placeholder="选择一个或多个接收组"
            >
              <ElOption
                v-for="item in options.recipient_groups"
                :key="item.id"
                :label="item.estimated_count ? `${item.name} · 约 ${item.estimated_count} 人` : item.name"
                :value="item.id"
              />
            </ElSelect>
            <div class="field-hint">
              {{
                form.audience_type === 'role'
                  ? '只会展开接收组里的角色规则，适合按平台角色或团队角色精准发送。'
                  : form.audience_type === 'feature_package'
                    ? '只会展开接收组里的功能包规则，适合按有效功能包命中成员。'
                : '接收组可混合配置指定用户、团队成员、团队管理员、角色和功能包规则。'
              }}
            </div>
          </ElFormItem>
            </div>
          </section>

          <section class="message-manage-block">
            <div class="message-manage-block__header">
              <h4>内容编辑</h4>
              <p>摘要使用普通文本输入，正文保留富文本，更符合公告和站内信后台的常见做法。</p>
            </div>
          <ElFormItem label="消息标题" prop="title">
            <ElInput v-model="form.title" maxlength="120" show-word-limit placeholder="例如：平台维护通知 / 团队待处理提醒" />
          </ElFormItem>

            <ElFormItem label="摘要">
              <ElInput
                v-model="form.summary"
                type="textarea"
                :rows="4"
                maxlength="200"
                show-word-limit
                resize="vertical"
                placeholder="顶部消息预览和消息列表优先展示这里的摘要，建议一句话说明重点。"
              />
              <div class="field-hint">摘要保持纯文本，更适合列表、副标题和公告摘要展示。</div>
            </ElFormItem>

            <ElFormItem label="正文">
              <ArtWangEditor
                v-model="form.content"
                height="320px"
                placeholder="正文支持富文本和超链接，相关跳转直接放在正文里即可。"
              />
              <div class="field-hint">内部链接、外部链接都放在正文里，不再单独配置动作按钮。</div>
            </ElFormItem>

          <div class="message-manage-grid">
            <ElFormItem label="业务分类">
              <ElInput v-model="form.biz_type" maxlength="80" placeholder="例如：platform_announcement / team_notice" />
            </ElFormItem>

            <ElFormItem label="失效时间">
              <ElDatePicker
                v-model="expiredAtValue"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ssZ"
                placeholder="可选，过期后不再展示"
              />
            </ElFormItem>
          </div>
          </section>
        </ElForm>
      </div>

      <aside class="message-manage-preview art-card">
        <header class="message-manage-section__header">
          <div>
            <h3>发送预览</h3>
            <p>这里模拟右上角消息面板和消息中心详情的最终展示效果。</p>
          </div>
        </header>

        <div class="message-manage-preview__card">
          <div class="message-manage-preview__eyebrow">
            <span>{{ selectedMessageTypeLabel }}</span>
            <span>{{ selectedPriorityLabel }}</span>
          </div>
          <h4>{{ form.title || '未填写标题' }}</h4>
          <div
            class="message-manage-preview__summary rich-text-content"
            v-html="previewSummaryHtml"
            @click="handlePreviewRichTextClick"
          ></div>
          <div class="message-manage-preview__meta">
            <ElTag effect="plain" type="info">{{ selectedSenderLabel }}</ElTag>
            <ElTag effect="plain">{{ selectedAudienceLabel }}</ElTag>
            <ElTag effect="plain" type="success">{{ receiverSummary }}</ElTag>
          </div>
          <div
            class="message-manage-preview__content rich-text-content"
            v-html="previewContentHtml"
            @click="handlePreviewRichTextClick"
          ></div>
        </div>

        <div class="message-manage-preview__template" v-if="activeTemplate">
          <div class="message-manage-preview__template-title">当前模板</div>
          <p>{{ activeTemplate.name }} · {{ activeTemplate.owner_scope === 'team' ? '团队模板' : '平台模板' }}</p>
          <span>{{ activeTemplate.description || '该模板未填写额外说明。' }}</span>
        </div>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import ArtWangEditor from '@/components/core/forms/art-wang-editor/index.vue'
  import MessageWorkspaceNav from '@/views/message/modules/message-workspace-nav.vue'
  import { fetchDispatchMessage, fetchGetMessageDispatchOptions } from '@/api/message'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { handleRichTextLinkNavigation } from '@/utils/navigation/rich-text'
  import { useMessageWorkspace } from '@/views/message/modules/useMessageWorkspace'

  defineOptions({ name: 'MessageDispatchConsole' })

  const props = defineProps<{
    scope: 'platform' | 'team'
  }>()

  const router = useRouter()
  const menuSpaceStore = useMenuSpaceStore()
  const loading = ref(false)
  const loadError = ref('')
  const submitting = ref(false)
  const formRef = ref()
  const { isTeamScope, skipTenantHeader, currentTeamName, ensureTeamContext, plainTextFromHtml } =
    useMessageWorkspace(props.scope)

  const options = reactive<Api.Message.DispatchOptions>({
    sender_scope: 'platform',
    current_tenant_id: '',
    current_tenant_name: '',
    sender_options: [],
    default_sender_id: '',
    audience_options: [],
    template_options: [],
    teams: [],
    users: [],
    recipient_groups: [],
    roles: [],
    feature_packages: [],
    default_message_type: 'notice',
    default_audience_type: 'all_users',
    default_priority: 'normal',
    supports_external_link: true
  })

  const createDefaultDispatchOptions = (): Api.Message.DispatchOptions => ({
    sender_scope: isTeamScope.value ? 'team' : 'platform',
    current_tenant_id: '',
    current_tenant_name: '',
    sender_options: [],
    default_sender_id: '',
    audience_options: [],
    template_options: [],
    teams: [],
    users: [],
    recipient_groups: [],
    roles: [],
    feature_packages: [],
    default_message_type: 'notice',
    default_audience_type: isTeamScope.value ? 'tenant_users' : 'all_users',
    default_priority: 'normal',
    supports_external_link: true
  })

  const form = reactive<Api.Message.DispatchParams & {
    sender_id: string
    template_id: string
    summary: string
    content: string
    action_type: string
    action_target: string
    biz_type: string
    expired_at: string
    target_tenant_ids: string[]
  }>({
    sender_id: '',
    template_id: '',
    message_type: 'notice',
    audience_type: 'all_users',
    target_tenant_ids: [],
    target_user_ids: [],
    target_group_ids: [],
    title: '',
    summary: '',
    content: '',
    priority: 'normal',
    action_type: 'none',
    action_target: '',
    biz_type: '',
    expired_at: ''
  })

  const effectiveTeamName = computed(() => options.current_tenant_name || currentTeamName.value)

  const pageTitle = computed(() => (isTeamScope.value ? '团队消息发送' : '消息发送'))
  const pageDescription = computed(() =>
    isTeamScope.value
      ? '以当前团队管理员身份给本团队成员发送通知、消息和待办，模板与发送记录都从这里进入。'
      : '统一给所有用户、团队管理员或指定团队成员发送站内通知、消息和待办，模板与发送记录都从这里进入。'
  )

  const messageTypeOptions = [
    { label: '通知', value: 'notice' },
    { label: '消息', value: 'message' },
    { label: '待办', value: 'todo' }
  ]

  const priorityOptions = [
    { label: '低', value: 'low' },
    { label: '普通', value: 'normal' },
    { label: '高', value: 'high' },
    { label: '紧急', value: 'urgent' }
  ]

  const expiredAtValue = computed({
    get: () => form.expired_at || '',
    set: (value: string) => {
      form.expired_at = value || ''
    }
  })

  const activeTemplate = computed(() =>
    options.template_options.find((item) => item.id === form.template_id) || null
  )

  const activeSender = computed(() =>
    options.sender_options.find((item) => item.id === form.sender_id) || null
  )

  const activeSenderDescription = computed(() => {
    if (activeSender.value?.description) return activeSender.value.description
    return isTeamScope.value
      ? '团队默认发送人为“团队”，也可以改成更具体的团队身份。'
      : '平台默认发送人为“平台”，也可以改成平台管理、平台空间等发信身份。'
  })

  const activeAudienceDescription = computed(
    () => options.audience_options.find((item) => item.value === form.audience_type)?.description || '请选择发送对象。'
  )

  const selectedAudienceLabel = computed(
    () => options.audience_options.find((item) => item.value === form.audience_type)?.label || '未选择对象'
  )

  const selectedSenderLabel = computed(() => activeSender.value?.name || '未选择发送人')

  const selectedMessageTypeLabel = computed(
    () => messageTypeOptions.find((item) => item.value === form.message_type)?.label || '通知'
  )

  const selectedPriorityLabel = computed(
    () => priorityOptions.find((item) => item.value === form.priority)?.label || '普通'
  )

  const senderScopeText = computed(() => {
    if (isTeamScope.value) {
      return `当前以团队管理员身份发送，默认团队为 ${effectiveTeamName.value}。`
    }
    return '当前以平台管理身份发送，可按所有用户、团队管理员或指定团队成员分发。'
  })

  const senderScopeBadge = computed(() => (isTeamScope.value ? '团队发信' : '平台发信'))

  const showTargetTeams = computed(() => form.audience_type !== 'all_users')
  const showTargetUsers = computed(() => form.audience_type === 'specified_users')
  const showRecipientGroups = computed(() =>
    ['recipient_group', 'role', 'feature_package'].includes(form.audience_type)
  )
  const targetTeamsLabel = computed(() => (isTeamScope.value ? '目标团队' : '目标团队'))
  const targetUsersLabel = computed(() => (isTeamScope.value ? '团队成员' : '目标用户'))

  const receiverSummary = computed(() => {
    if (form.audience_type === 'all_users') return '全部有效用户'
    if (form.audience_type === 'specified_users') {
      if (!form.target_user_ids?.length) return '待选择用户'
      const names = options.users
        .filter((item) => form.target_user_ids?.includes(item.id))
        .map((item) => item.display_name)
      return names.join('、') || '待选择用户'
    }
    if (form.audience_type === 'recipient_group') {
      if (!form.target_group_ids?.length) return '待选择接收组'
      const names = options.recipient_groups
        .filter((item) => form.target_group_ids?.includes(item.id))
        .map((item) => item.name)
      return names.join('、') || '待选择接收组'
    }
    if (form.audience_type === 'role') {
      if (!form.target_group_ids?.length) return '待选择含角色规则的接收组'
      const names = options.recipient_groups
        .filter((item) => form.target_group_ids?.includes(item.id))
        .map((item) => item.name)
      return names.join('、') || '待选择含角色规则的接收组'
    }
    if (form.audience_type === 'feature_package') {
      if (!form.target_group_ids?.length) return '待选择含功能包规则的接收组'
      const names = options.recipient_groups
        .filter((item) => form.target_group_ids?.includes(item.id))
        .map((item) => item.name)
      return names.join('、') || '待选择含功能包规则的接收组'
    }
    if (isTeamScope.value) {
      return effectiveTeamName.value
    }
    if (!form.target_tenant_ids?.length) return '待选择团队'
    const names = options.teams
      .filter((item) => form.target_tenant_ids?.includes(item.id))
      .map((item) => item.name)
    return names.join('、') || '待选择团队'
  })

  const heroMetrics = computed(() => [
    { label: '可用发送人', value: options.sender_options.length },
    { label: '可用模板', value: options.template_options.length },
    { label: '可选对象', value: options.audience_options.length },
    { label: isTeamScope.value ? '当前团队' : '目标团队', value: isTeamScope.value ? effectiveTeamName.value : options.teams.length }
  ])
  const canDispatch = computed(() => !loadError.value && options.sender_options.length > 0)

  const normalizeEditorValue = (value?: string) => {
    const target = `${value || ''}`.trim()
    return plainTextFromHtml(target) ? target : ''
  }

  const normalizeSummaryValue = (value?: string) => `${value || ''}`.trim()

  const wrapFallbackHtml = (value: string, fallback: string) => {
    const normalized = normalizeEditorValue(value)
    return normalized || `<p>${fallback}</p>`
  }

  const previewSummaryHtml = computed(() =>
    normalizeSummaryValue(form.summary)
      ? `<p>${normalizeSummaryValue(form.summary)}</p>`
      : '<p>摘要会显示在顶部消息预览和列表项副标题里。</p>'
  )

  const previewContentHtml = computed(() =>
    wrapFallbackHtml(form.content, '正文会显示在消息中心详情区域，链接也直接写在正文里。')
  )

  const resetFormDefaults = () => {
    form.sender_id = options.default_sender_id || options.sender_options[0]?.id || ''
    form.message_type = options.default_message_type || 'notice'
    form.audience_type = options.default_audience_type || (isTeamScope.value ? 'tenant_users' : 'all_users')
    form.priority = options.default_priority || 'normal'
    form.target_tenant_ids = isTeamScope.value && options.current_tenant_id ? [options.current_tenant_id] : []
    form.target_user_ids = []
    form.target_group_ids = []
    form.action_type = 'none'
    form.action_target = ''
  }

  const applyTemplate = (template: Api.Message.DispatchTemplateOption | null) => {
    if (!template) return
    form.message_type = template.message_type || form.message_type
    if (options.audience_options.some((item) => item.value === template.audience_type)) {
      form.audience_type = template.audience_type
    }
    if (!plainTextFromHtml(form.title)) {
      form.title = template.title_template || ''
    }
    if (!plainTextFromHtml(form.summary)) {
      form.summary = template.summary_template || ''
    }
    if (!plainTextFromHtml(form.content)) {
      form.content = template.content_template || ''
    }
  }

  const loadOptions = async () => {
    loading.value = true
    loadError.value = ''
    try {
      ensureTeamContext()
      const data = await fetchGetMessageDispatchOptions({
        skipTenantHeader: skipTenantHeader.value
      })
      Object.assign(options, data || {})
      resetFormDefaults()
    } catch (error) {
      Object.assign(options, createDefaultDispatchOptions())
      resetFormDefaults()
      loadError.value = '发信配置暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  const handleTemplateChange = () => {
    applyTemplate(activeTemplate.value)
  }

  const handleAudienceChange = () => {
    if (form.audience_type === 'all_users') {
      form.target_tenant_ids = []
      form.target_user_ids = []
      form.target_group_ids = []
      return
    }
    if (form.audience_type === 'specified_users') {
      form.target_tenant_ids = []
      form.target_group_ids = []
      return
    }
    if (form.audience_type === 'recipient_group' || form.audience_type === 'role' || form.audience_type === 'feature_package') {
      form.target_tenant_ids = []
      form.target_user_ids = []
      return
    }
    if (isTeamScope.value && options.current_tenant_id) {
      form.target_tenant_ids = [options.current_tenant_id]
    }
    form.target_user_ids = []
    form.target_group_ids = []
  }

  const submitDispatch = async () => {
    if (!form.title.trim()) {
      ElMessage.warning('请先填写消息标题')
      return
    }
    if (!normalizeSummaryValue(form.summary) && !plainTextFromHtml(form.content)) {
      ElMessage.warning('请至少填写摘要或正文')
      return
    }
    if (
      !isTeamScope.value &&
      showTargetTeams.value &&
      !showTargetUsers.value &&
      !showRecipientGroups.value &&
      !form.target_tenant_ids?.length
    ) {
      ElMessage.warning('请选择目标团队')
      return
    }
    if (showTargetUsers.value && !form.target_user_ids?.length) {
      ElMessage.warning('请选择目标用户')
      return
    }
    if (showRecipientGroups.value && !form.target_group_ids?.length) {
      ElMessage.warning(
        form.audience_type === 'role'
          ? '请选择包含角色规则的接收组'
          : form.audience_type === 'feature_package'
            ? '请选择包含功能包规则的接收组'
            : '请选择接收组'
      )
      return
    }

    submitting.value = true
    try {
      const result = await fetchDispatchMessage(
        {
          ...form,
          template_id: form.template_id || undefined,
          summary: normalizeSummaryValue(form.summary),
          content: normalizeEditorValue(form.content),
          action_type: 'none',
          action_target: '',
          target_tenant_ids:
            showTargetTeams.value && !showTargetUsers.value && !showRecipientGroups.value
              ? form.target_tenant_ids
              : [],
          target_user_ids: showTargetUsers.value ? form.target_user_ids : [],
          target_group_ids: showRecipientGroups.value ? form.target_group_ids : []
        },
        {
          skipTenantHeader: skipTenantHeader.value
        }
      )
      ElMessage.success(result.dispatch_status === 'queued' ? '消息已进入发送队列' : `发送成功，已投递 ${result.delivery_count} 人`)
      form.title = ''
      form.summary = ''
      form.content = ''
      form.biz_type = ''
      form.expired_at = ''
      form.template_id = ''
      if (!isTeamScope.value) {
        form.target_tenant_ids = []
      }
      form.target_user_ids = []
      form.target_group_ids = []
      resetFormDefaults()
    } catch (error) {
      ElMessage.error('发送消息失败')
    } finally {
      submitting.value = false
    }
  }

  const handlePreviewRichTextClick = async (event: MouseEvent) => {
    await handleRichTextLinkNavigation(event, {
      router,
      spaceResolver: menuSpaceStore
    })
  }

  onMounted(() => {
    loadOptions()
  })
</script>

<style scoped lang="scss">
  .message-manage-shell {
    display: grid;
    grid-template-columns: minmax(0, 1.7fr) minmax(320px, 0.9fr);
    gap: 18px;
    margin-top: 18px;
  }

  .message-manage-inline-alert {
    margin-top: 16px;
  }

  .message-manage-main,
  .message-manage-preview {
    padding: 20px 22px;
    border-radius: 20px;
  }

  .message-manage-section__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 18px;
  }

  .message-manage-section__header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #0f172a;
  }

  .message-manage-section__header p {
    margin: 6px 0 0;
    font-size: 13px;
    line-height: 1.6;
    color: #64748b;
  }

  .message-manage-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 16px;
  }

  .message-manage-hero__actions {
    display: flex;
    gap: 12px;
  }

  .message-manage-form {
    display: grid;
    gap: 18px;
  }

  .message-manage-block {
    display: grid;
    gap: 14px;
    padding: 18px;
    border: 1px solid rgb(226 232 240 / 0.88);
    border-radius: 20px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.92));
  }

  .message-manage-block__header {
    display: grid;
    gap: 4px;
  }

  .message-manage-block__header h4 {
    margin: 0;
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-manage-block__header p {
    margin: 0;
    font-size: 12px;
    line-height: 1.7;
    color: #64748b;
  }

  .message-manage-target-layout {
    display: grid;
    gap: 0 16px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .message-manage-inline-options {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .message-manage-fixed-target {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 14px 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 16px;
    background: linear-gradient(180deg, rgb(248 250 252 / 0.95), rgb(255 255 255 / 1));
  }

  .message-manage-fixed-target strong {
    font-size: 14px;
    color: #0f172a;
  }

  .message-manage-fixed-target span,
  .field-hint {
    font-size: 12px;
    line-height: 1.6;
    color: #94a3b8;
  }

  .message-manage-preview__card {
    padding: 18px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 18px;
    background: linear-gradient(180deg, rgb(248 250 252 / 0.95), rgb(255 255 255 / 1));
  }

  .message-manage-preview__eyebrow,
  .message-manage-preview__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .message-manage-preview__eyebrow {
    font-size: 12px;
    color: #64748b;
  }

  .message-manage-preview__card h4 {
    margin: 12px 0 0;
    font-size: 20px;
    line-height: 1.5;
    color: #0f172a;
  }

  .message-manage-preview__summary {
    margin-top: 10px;
    padding: 14px 16px;
    border-radius: 16px;
    background: rgb(248 250 252 / 0.92);
  }

  .message-manage-preview__meta {
    margin-top: 14px;
  }

  .message-manage-preview__content {
    margin-top: 16px;
    min-height: 200px;
    border-radius: 16px;
    background: rgb(255 255 255 / 0.88);
    padding: 16px;
  }

  .rich-text-content {
    font-size: 13px;
    line-height: 1.8;
    color: #334155;
    word-break: break-word;
  }

  .rich-text-content :deep(p) {
    margin: 0 0 0.9em;
  }

  .rich-text-content :deep(p:last-child) {
    margin-bottom: 0;
  }

  .rich-text-content :deep(a) {
    color: #2563eb;
    text-decoration: underline;
  }

  .rich-text-content :deep(ul),
  .rich-text-content :deep(ol) {
    padding-left: 1.3em;
    margin: 0 0 0.9em;
  }

  .message-manage-preview__template {
    margin-top: 16px;
    padding: 16px 18px;
    border-radius: 18px;
    background: rgb(248 250 252 / 0.9);
  }

  .message-manage-preview__template-title {
    font-size: 12px;
    color: #94a3b8;
  }

  .message-manage-preview__template p {
    margin: 8px 0 4px;
    font-size: 14px;
    font-weight: 600;
    color: #0f172a;
  }

  .message-manage-preview__template span {
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  @media (max-width: 1180px) {
    .message-manage-shell {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .message-manage-grid,
    .message-manage-target-layout {
      grid-template-columns: 1fr;
    }

    .message-manage-main,
    .message-manage-preview {
      padding: 18px;
    }

    .message-manage-section__header {
      flex-direction: column;
    }
  }
</style>
