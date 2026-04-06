<template>
  <div class="message-manage-page art-full-height">
    <AdminWorkspaceHero :title="pageTitle" :description="pageDescription" :metrics="heroMetrics">
      <div class="message-manage-hero__actions">
        <ElButton @click="loadOptions" :loading="loading" v-ripple>刷新配置</ElButton>
        <ElButton
          type="primary"
          @click="submitDispatch"
          :loading="submitting"
          :disabled="!canDispatch"
          v-ripple
          >发送消息</ElButton
        >
      </div>
    </AdminWorkspaceHero>

    <div class="message-manage-nav">
      <MessageWorkspaceNav :scope="props.scope" current="dispatch" />
    </div>

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
                    v-for="item in filteredTemplateOptions"
                    :key="item.id"
                    :label="templateOptionLabel(item)"
                    :value="item.id"
                  />
                </ElSelect>
                <div class="field-hint">
                  {{
                    isCollaborationScope
                      ? '仅展示当前协作空间创建的模板，协作空间页不会混入平台模板。'
                      : '仅展示平台模板。'
                  }}
                </div>
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
                  <ElRadioButton
                    v-for="item in messageTypeOptions"
                    :key="item.value"
                    :value="item.value"
                  >
                    {{ item.label }}
                  </ElRadioButton>
                </ElRadioGroup>
              </ElFormItem>

              <ElFormItem label="消息优先级">
                <ElRadioGroup v-model="form.priority" class="message-manage-inline-options">
                  <ElRadioButton
                    v-for="item in priorityOptions"
                    :key="item.value"
                    :value="item.value"
                  >
                    {{ item.label }}
                  </ElRadioButton>
                </ElRadioGroup>
              </ElFormItem>
            </div>
          </section>

          <section class="message-manage-block">
            <div class="message-manage-block__header">
              <h4>接收对象</h4>
              <p>按主流消息公告后台习惯，先选对象，再补具体协作空间、用户或接收组。</p>
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
              <ElFormItem
                v-if="showTargetCollaborationWorkspaces"
                :label="targetCollaborationWorkspacesLabel"
              >
                <ElSelect
                  v-if="!isCollaborationScope"
                  v-model="form.targetCollaborationWorkspaceIds"
                  multiple
                  filterable
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="选择一个或多个协作空间"
                >
                  <ElOption
                    v-for="item in options.collaboration_workspaces"
                    :key="item.id"
                    :label="item.name"
                    :value="item.id"
                  />
                </ElSelect>
                <div v-else class="message-manage-fixed-target">
                  <strong>{{
                    options.current_collaboration_workspace_name ||
                    currentCollaborationWorkspaceName
                  }}</strong>
                  <span>协作空间上下文只允许向当前协作空间发送。</span>
                </div>
                <div class="field-hint">
                  {{
                    isCollaborationScope
                      ? '发送对象会自动绑定到当前协作空间，无需再额外选择。'
                      : '平台可选择多个目标协作空间，系统会按对象类型自动匹配成员。'
                  }}
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
                    :label="item.display_name"
                    :value="item.id"
                  />
                </ElSelect>
                <div class="field-hint">
                  {{
                    isCollaborationScope
                      ? '协作空间侧只会列出当前协作空间成员。'
                      : '平台侧可以直接按用户维度精确发送。'
                  }}
                </div>
              </ElFormItem>

              <ElFormItem
                v-if="showRecipientGroups"
                :label="
                  form.audience_type === 'role'
                    ? '角色接收组'
                    : form.audience_type === 'feature_package'
                      ? '功能包接收组'
                      : '接收组'
                "
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
                    :label="
                      item.estimated_count
                        ? `${item.name} · 约 ${item.estimated_count} 人`
                        : item.name
                    "
                    :value="item.id"
                  />
                </ElSelect>
                <div class="field-hint">
                  {{
                    form.audience_type === 'role'
                      ? '只会展开接收组里的角色规则，适合按个人空间角色或协作空间角色精准发送。'
                      : form.audience_type === 'feature_package'
                        ? '只会展开接收组里的功能包规则，适合按有效功能包命中成员。'
                        : '接收组可混合配置指定用户、协作空间成员、协作空间管理员、角色和功能包规则。'
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
              <ElInput
                v-model="form.title"
                maxlength="120"
                show-word-limit
                placeholder="例如：平台维护通知 / 协作空间待处理提醒"
              />
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
                <ElInput
                  v-model="form.biz_type"
                  maxlength="80"
                  placeholder="例如：personal_announcement / collaboration_notice"
                />
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
          <p>{{ templateOptionLabel(activeTemplate) }}</p>
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
    scope: 'personal' | 'collaboration'
  }>()

  const router = useRouter()
  const menuSpaceStore = useMenuSpaceStore()
  const loading = ref(false)
  const loadError = ref('')
  const submitting = ref(false)
  const formRef = ref()
  const {
    collaborationWorkspaceStore,
    isCollaborationScope,
    skipCollaborationWorkspaceHeader,
    currentCollaborationWorkspaceId,
    currentCollaborationWorkspaceName,
    currentWorkspaceName,
    currentWorkspaceLabel,
    ensureCollaborationWorkspaceContext,
    plainTextFromHtml
  } = useMessageWorkspace(props.scope)

  const options = reactive<Api.Message.DispatchOptions>({
    sender_scope: 'personal',
    current_collaboration_workspace_id: '',
    current_collaboration_workspace_name: '',
    sender_options: [],
    default_sender_id: '',
    audience_options: [],
    template_options: [],
    collaboration_workspaces: [],
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
    sender_scope: isCollaborationScope.value ? 'collaboration' : 'personal',
    current_collaboration_workspace_id: '',
    current_collaboration_workspace_name: '',
    sender_options: [],
    default_sender_id: '',
    audience_options: [],
    template_options: [],
    collaboration_workspaces: [],
    users: [],
    recipient_groups: [],
    roles: [],
    feature_packages: [],
    default_message_type: 'notice',
    default_audience_type: isCollaborationScope.value
      ? 'collaboration_workspace_users'
      : 'all_users',
    default_priority: 'normal',
    supports_external_link: true
  })

  const normalizeDispatchOptions = (
    payload?: Partial<Api.Message.DispatchOptions> | null
  ): Api.Message.DispatchOptions => {
    const base = createDefaultDispatchOptions()
    return {
      ...base,
      ...(payload || {}),
      sender_options: Array.isArray(payload?.sender_options)
        ? payload.sender_options
        : base.sender_options,
      audience_options: Array.isArray(payload?.audience_options)
        ? payload.audience_options
        : base.audience_options,
      template_options: Array.isArray(payload?.template_options)
        ? payload.template_options
        : base.template_options,
      collaboration_workspaces: Array.isArray(payload?.collaboration_workspaces)
        ? payload.collaboration_workspaces
        : base.collaboration_workspaces,
      users: Array.isArray(payload?.users) ? payload.users : base.users,
      recipient_groups: Array.isArray(payload?.recipient_groups)
        ? payload.recipient_groups
        : base.recipient_groups,
      roles: Array.isArray(payload?.roles) ? payload.roles : base.roles,
      feature_packages: Array.isArray(payload?.feature_packages)
        ? payload.feature_packages
        : base.feature_packages
    }
  }

  const form = reactive<
    Api.Message.DispatchParams & {
      sender_id: string
      template_id: string
      summary: string
      content: string
      action_type: string
      action_target: string
      biz_type: string
      expired_at: string
      targetCollaborationWorkspaceIds: string[]
    }
  >({
    sender_id: '',
    template_id: '',
    message_type: 'notice',
    audience_type: 'all_users',
    targetCollaborationWorkspaceIds: [],
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

  const effectiveCollaborationWorkspaceName = computed(
    () => options.current_collaboration_workspace_name || currentCollaborationWorkspaceName.value
  )

  const pageTitle = computed(() => (isCollaborationScope.value ? '协作空间消息发送' : '消息发送'))
  const pageDescription = computed(() =>
    isCollaborationScope.value
      ? `以 ${currentWorkspaceLabel.value} 视角给 ${effectiveCollaborationWorkspaceName.value} 发送通知、消息和待办，模板与发送记录都从这里进入。`
      : '统一给所有用户、协作空间管理员或指定协作空间成员发送站内通知、消息和待办，模板与发送记录都从这里进入。'
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

  const filteredTemplateOptions = computed(() =>
    (options.template_options || []).filter((item) =>
      isCollaborationScope.value
        ? item.owner_scope === 'collaboration'
        : item.owner_scope === 'personal'
    )
  )

  const activeTemplate = computed(
    () => filteredTemplateOptions.value.find((item) => item.id === form.template_id) || null
  )

  const activeSender = computed(
    () => options.sender_options.find((item) => item.id === form.sender_id) || null
  )

  const activeSenderDescription = computed(() => {
    if (activeSender.value?.description) return activeSender.value.description
    return isCollaborationScope.value
      ? '协作空间默认发送人为“协作空间”，也可以改成更具体的协作空间身份。'
      : '个人空间默认发送人为“个人空间”，也可以改成更具体的个人身份。'
  })

  const activeAudienceDescription = computed(
    () =>
      options.audience_options.find((item) => item.value === form.audience_type)?.description ||
      '请选择发送对象。'
  )

  const selectedAudienceLabel = computed(
    () =>
      options.audience_options.find((item) => item.value === form.audience_type)?.label ||
      '未选择对象'
  )

  const selectedSenderLabel = computed(() => activeSender.value?.name || '未选择发送人')

  const selectedMessageTypeLabel = computed(
    () => messageTypeOptions.find((item) => item.value === form.message_type)?.label || '通知'
  )

  const selectedPriorityLabel = computed(
    () => priorityOptions.find((item) => item.value === form.priority)?.label || '普通'
  )

  const senderScopeText = computed(() => {
    if (isCollaborationScope.value) {
      return `当前授权工作空间为 ${currentWorkspaceName.value}，默认协作空间视图为 ${effectiveCollaborationWorkspaceName.value}。`
    }
    return '当前以个人空间身份发送，可按所有用户、协作空间管理员或指定协作空间成员分发。'
  })

  const senderScopeBadge = computed(() =>
    isCollaborationScope.value ? '协作空间发信' : '个人空间发信'
  )

  const showTargetCollaborationWorkspaces = computed(() => form.audience_type !== 'all_users')
  const showTargetUsers = computed(() => form.audience_type === 'specified_users')
  const showRecipientGroups = computed(() =>
    ['recipient_group', 'role', 'feature_package'].includes(form.audience_type)
  )
  const targetCollaborationWorkspacesLabel = computed(() =>
    isCollaborationScope.value ? '目标协作空间' : '目标协作空间'
  )
  const targetUsersLabel = computed(() =>
    isCollaborationScope.value ? '协作空间成员' : '目标用户'
  )

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
    if (isCollaborationScope.value) {
      return effectiveCollaborationWorkspaceName.value
    }
    if (!form.targetCollaborationWorkspaceIds?.length) return '待选择协作空间'
    const names = options.collaboration_workspaces
      .filter((item) => form.targetCollaborationWorkspaceIds?.includes(item.id))
      .map((item) => item.name)
    return names.join('、') || '待选择协作空间'
  })

  const heroMetrics = computed(() => [
    { label: '可用发送人', value: (options.sender_options || []).length },
    { label: '可用模板', value: (options.template_options || []).length },
    { label: '可选对象', value: (options.audience_options || []).length },
    {
      label: isCollaborationScope.value ? '当前协作空间' : '目标协作空间',
      value: isCollaborationScope.value
        ? effectiveCollaborationWorkspaceName.value
        : options.collaboration_workspaces.length
    }
  ])
  const canDispatch = computed(() => !loadError.value && (options.sender_options || []).length > 0)

  const normalizeEditorValue = (value?: string) => {
    const target = `${value || ''}`.trim()
    return plainTextFromHtml(target) ? target : ''
  }

  const normalizeSummaryValue = (value?: string) => `${value || ''}`.trim()

  const wrapFallbackHtml = (value: string, fallback: string) => {
    const normalized = normalizeEditorValue(value)
    return normalized || `<p>${fallback}</p>`
  }
  const templateOptionLabel = (template: Api.Message.DispatchTemplateOption) => {
    if (isCollaborationScope.value) return `${template.name} · 协作空间模板`
    return `${template.name} · 个人空间模板`
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
    form.audience_type =
      options.default_audience_type ||
      (isCollaborationScope.value ? 'collaboration_workspace_users' : 'all_users')
    form.priority = options.default_priority || 'normal'
    form.targetCollaborationWorkspaceIds =
      isCollaborationScope.value && options.current_collaboration_workspace_id
        ? [options.current_collaboration_workspace_id]
        : []
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
      ensureCollaborationWorkspaceContext()
      if (isCollaborationScope.value) {
        await collaborationWorkspaceStore.loadMyCollaborationWorkspaces({
          preferredCollaborationWorkspaceId: currentCollaborationWorkspaceId.value || undefined
        })
      }
      Object.assign(options, createDefaultDispatchOptions())
      const data = await fetchGetMessageDispatchOptions({
        skipCollaborationWorkspaceHeader: skipCollaborationWorkspaceHeader.value
      })
      Object.assign(options, normalizeDispatchOptions(data))
      if (
        form.template_id &&
        !filteredTemplateOptions.value.some((item) => item.id === form.template_id)
      ) {
        form.template_id = ''
      }
      resetFormDefaults()
    } catch {
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
      form.targetCollaborationWorkspaceIds = []
      form.target_user_ids = []
      form.target_group_ids = []
      return
    }
    if (form.audience_type === 'specified_users') {
      form.targetCollaborationWorkspaceIds = []
      form.target_group_ids = []
      return
    }
    if (
      form.audience_type === 'recipient_group' ||
      form.audience_type === 'role' ||
      form.audience_type === 'feature_package'
    ) {
      form.targetCollaborationWorkspaceIds = []
      form.target_user_ids = []
      return
    }
    if (isCollaborationScope.value && options.current_collaboration_workspace_id) {
      form.targetCollaborationWorkspaceIds = [options.current_collaboration_workspace_id]
    }
    form.target_user_ids = []
    form.target_group_ids = []
  }

  const submitDispatch = async () => {
    if (isCollaborationScope.value && !currentCollaborationWorkspaceId.value) {
      await collaborationWorkspaceStore.loadMyCollaborationWorkspaces({
        preferredCollaborationWorkspaceId: currentCollaborationWorkspaceId.value || undefined
      })
    }
    if (!form.title.trim()) {
      ElMessage.warning('请先填写消息标题')
      return
    }
    if (!normalizeSummaryValue(form.summary) && !plainTextFromHtml(form.content)) {
      ElMessage.warning('请至少填写摘要或正文')
      return
    }
    if (
      !isCollaborationScope.value &&
      showTargetCollaborationWorkspaces.value &&
      !showTargetUsers.value &&
      !showRecipientGroups.value &&
      !form.targetCollaborationWorkspaceIds?.length
    ) {
      ElMessage.warning('请选择目标协作空间')
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
          target_collaboration_workspace_ids:
            showTargetCollaborationWorkspaces.value &&
            !showTargetUsers.value &&
            !showRecipientGroups.value
              ? form.targetCollaborationWorkspaceIds
              : [],
          target_user_ids: showTargetUsers.value ? form.target_user_ids : [],
          target_group_ids: showRecipientGroups.value ? form.target_group_ids : []
        },
        {
          skipCollaborationWorkspaceHeader: skipCollaborationWorkspaceHeader.value
        }
      )
      ElMessage.success(
        result.dispatch_status === 'queued'
          ? '消息已进入发送队列'
          : `发送成功，已投递 ${result.delivery_count} 人`
      )
      form.title = ''
      form.summary = ''
      form.content = ''
      form.biz_type = ''
      form.expired_at = ''
      form.template_id = ''
      if (!isCollaborationScope.value) {
        form.targetCollaborationWorkspaceIds = []
      }
      form.target_user_ids = []
      form.target_group_ids = []
      resetFormDefaults()
    } catch {
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
    gap: 16px;
    margin-top: 0;
    align-items: stretch;
  }

  .message-manage-nav {
    margin-top: 0;
  }

  .message-manage-inline-alert {
    margin-top: 0;
  }

  .message-manage-main,
  .message-manage-preview {
    display: flex;
    flex-direction: column;
    padding: 22px;
    border-radius: 20px;
    border: 1px solid var(--art-card-border);
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(249 250 251 / 0.94));
    box-shadow: var(--art-shadow-sm);
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
    font-weight: 750;
    letter-spacing: -0.03em;
    color: var(--art-text-strong);
  }

  .message-manage-section__header p {
    margin: 6px 0 0;
    font-size: 13px;
    line-height: 1.6;
    color: var(--art-text-muted);
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
    gap: 16px;
  }

  .message-manage-block {
    display: grid;
    gap: 12px;
    padding: 18px;
    border: 1px solid color-mix(in srgb, var(--art-card-border) 92%, white);
    border-radius: 18px;
    background: rgb(255 255 255 / 0.92);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.6);
  }

  .message-manage-block__header {
    display: grid;
    gap: 4px;
  }

  .message-manage-block__header h4 {
    margin: 0;
    font-size: 15px;
    font-weight: 750;
    color: var(--art-text-strong);
  }

  .message-manage-block__header p {
    margin: 0;
    font-size: 12px;
    line-height: 1.7;
    color: var(--art-text-muted);
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
    border: 1px solid color-mix(in srgb, var(--art-card-border) 92%, white);
    border-radius: 16px;
    background: rgb(249 250 251 / 0.96);
  }

  .message-manage-fixed-target strong {
    font-size: 14px;
    color: var(--art-text-strong);
  }

  .message-manage-fixed-target span,
  .field-hint {
    font-size: 12px;
    line-height: 1.6;
    color: var(--art-text-soft);
  }

  .message-manage-preview__card {
    padding: 18px;
    border: 1px solid color-mix(in srgb, var(--art-card-border) 92%, white);
    border-radius: 18px;
    background:
      radial-gradient(circle at top right, rgb(191 219 254 / 0.12), transparent 24%),
      linear-gradient(180deg, rgb(248 250 252 / 0.96), rgb(255 255 255 / 1));
    box-shadow: var(--art-shadow-sm);
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
    color: var(--art-text-muted);
  }

  .message-manage-preview__card h4 {
    margin: 12px 0 0;
    font-size: 20px;
    line-height: 1.5;
    color: var(--art-text-strong);
  }

  .message-manage-preview__summary {
    margin-top: 12px;
    padding: 14px 16px;
    border-radius: 16px;
    border: 1px solid color-mix(in srgb, var(--art-card-border) 88%, white);
    background: rgb(248 250 252 / 0.88);
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
    border: 1px solid color-mix(in srgb, var(--art-card-border) 88%, white);
  }

  .rich-text-content {
    font-size: 13px;
    line-height: 1.8;
    color: var(--art-text-base);
    word-break: break-word;
  }

  .rich-text-content :deep(p) {
    margin: 0 0 0.9em;
  }

  .rich-text-content :deep(p:last-child) {
    margin-bottom: 0;
  }

  .rich-text-content :deep(a) {
    color: var(--theme-color);
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
    border: 1px dashed color-mix(in srgb, var(--art-card-border) 92%, white);
    background: rgb(248 250 252 / 0.88);
  }

  .message-manage-preview__card {
    flex: 1;
  }

  .message-manage-preview__template-title {
    font-size: 12px;
    color: var(--art-text-soft);
  }

  .message-manage-preview__template p {
    margin: 8px 0 4px;
    font-size: 14px;
    font-weight: 600;
    color: var(--art-text-strong);
  }

  .message-manage-preview__template span {
    font-size: 12px;
    line-height: 1.6;
    color: var(--art-text-muted);
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
