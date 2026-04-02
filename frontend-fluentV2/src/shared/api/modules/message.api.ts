import { requestData, type RequestDataConfig } from '@/shared/api/client'
import type {
  InboxMessageDetail,
  InboxSummary,
  InboxThread,
  MessageDispatchResult,
  MessageDispatchOptions,
  MessageDispatchPayload,
  MessageRecipientGroupTarget,
  MessageRecipientGroupRecord,
  MessageRecipientGroupSavePayload,
  MessageTimelineItem,
  MessageDeliveryRecord,
  MessageRecord,
  MessageRecordDetail,
  MessageSenderRecord,
  MessageSenderSavePayload,
  MessageTemplateRecord,
  MessageTemplateSavePayload,
} from '@/shared/types/message-center'

interface PaginationEnvelope<T> {
  current?: number
  total?: number
  size?: number
  records?: T[]
}

export type MessageScope = 'platform' | 'team'

export interface PaginatedResult<T> {
  current: number
  total: number
  size: number
  records: T[]
}

function toPaginatedResult<T>(input: PaginationEnvelope<T>): PaginatedResult<T> {
  return {
    current: Number(input.current || 1),
    total: Number(input.total || 0),
    size: Number(input.size || input.records?.length || 0),
    records: Array.isArray(input.records) ? input.records : [],
  }
}

function resolveScopeConfig(scope: MessageScope, config: RequestDataConfig): RequestDataConfig {
  return {
    ...config,
    tenantMode: scope === 'platform' ? 'none' : 'current',
  }
}

function normalizeText(value: unknown) {
  return `${value || ''}`.trim()
}

function normalizeInboxSummary(input: Record<string, unknown>): InboxSummary {
  return {
    unreadTotal: Number(input.unread_total || input.unreadTotal || 0),
    noticeCount: Number(input.notice_count || input.noticeCount || 0),
    messageCount: Number(input.message_count || input.messageCount || 0),
    todoCount: Number(input.todo_count || input.todoCount || 0),
  }
}

function normalizeInboxThread(input: Record<string, unknown>): InboxThread {
  return {
    id: normalizeText(input.id || input.delivery_id),
    title: normalizeText(input.title),
    summary: normalizeText(input.summary),
    messageType: normalizeText(input.message_type || input.messageType),
    boxType: normalizeText(input.box_type || input.boxType),
    read: Boolean(input.is_read ?? input.read),
    todoStatus: normalizeText(input.todo_status || input.todoStatus),
    senderName: normalizeText(input.sender_name || input.senderName),
    senderAvatarUrl: normalizeText(input.sender_avatar_url || input.senderAvatarUrl),
    sentAt: normalizeText(input.sent_at || input.sentAt),
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
    tenantName: normalizeText(input.tenant_name || input.tenantName),
  }
}

function normalizeInboxDetail(input: Record<string, unknown>): InboxMessageDetail {
  return {
    ...normalizeInboxThread(input),
    content: normalizeText(input.content),
    priority: normalizeText(input.priority || 'normal'),
    tenantName: normalizeText(input.tenant_name || input.tenantName),
    payload: ((input.payload || {}) as Record<string, unknown>) || {},
  }
}

function normalizeRecipientTarget(input: Record<string, unknown>): MessageRecipientGroupTarget {
  const userName = normalizeText(input.user_name || input.userName)
  const tenantName = normalizeText(input.tenant_name || input.tenantName)
  const roleName = normalizeText(input.role_name || input.roleName)
  const packageName = normalizeText(input.package_name || input.packageName)
  const targetType = normalizeText(input.target_type || input.targetType)
  const targetLabel = userName || tenantName || roleName || packageName || targetType || '未命名目标'

  return {
    id: normalizeText(input.id),
    targetType,
    targetLabel,
    targetValue:
      normalizeText(input.user_id || input.userId) ||
      normalizeText(input.tenant_id || input.tenantId) ||
      normalizeText(input.role_code || input.roleCode) ||
      normalizeText(input.package_key || input.packageKey),
    sortOrder: Number(input.sort_order || input.sortOrder || 0),
    meta: ((input.meta || {}) as Record<string, unknown>) || {},
  }
}

function normalizeAudienceSubject(input: Record<string, unknown>) {
  return {
    id: normalizeText(input.id || input.value),
    name: normalizeText(input.name || input.label),
    displayName: normalizeText(input.display_name || input.displayName) || undefined,
    description: normalizeText(input.description) || undefined,
    teamId: normalizeText(input.team_id || input.teamId) || null,
    teamName: normalizeText(input.team_name || input.teamName) || undefined,
  }
}

function normalizeMessageTemplate(input: Record<string, unknown>): MessageTemplateRecord {
  return {
    id: normalizeText(input.id),
    templateKey: normalizeText(input.template_key || input.templateKey),
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    messageType: normalizeText(input.message_type || input.messageType),
    ownerScope: normalizeText(input.owner_scope || input.ownerScope),
    ownerTenantId: normalizeText(input.owner_tenant_id || input.ownerTenantId) || null,
    ownerTenantName: normalizeText(input.owner_tenant_name || input.ownerTenantName),
    audienceType: normalizeText(input.audience_type || input.audienceType),
    titleTemplate: normalizeText(input.title_template || input.titleTemplate),
    summaryTemplate: normalizeText(input.summary_template || input.summaryTemplate),
    contentTemplate: normalizeText(input.content_template || input.contentTemplate),
    status: normalizeText(input.status || 'normal'),
    editable: Boolean(input.editable ?? true),
    meta: ((input.meta || {}) as Record<string, unknown>) || {},
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeMessageSender(input: Record<string, unknown>): MessageSenderRecord {
  return {
    id: normalizeText(input.id),
    scopeType: normalizeText(input.scope_type || input.scopeType),
    scopeId: normalizeText(input.scope_id || input.scopeId) || null,
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    avatarUrl: normalizeText(input.avatar_url || input.avatarUrl),
    isDefault: Boolean(input.is_default ?? input.isDefault),
    status: normalizeText(input.status || 'normal'),
    editable: Boolean(input.editable ?? true),
    meta: ((input.meta || {}) as Record<string, unknown>) || {},
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeMessageRecipientGroup(input: Record<string, unknown>): MessageRecipientGroupRecord {
  return {
    id: normalizeText(input.id),
    groupKey: normalizeText(input.group_key || input.groupKey),
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    scopeType: normalizeText(input.scope_type || input.scopeType),
    scopeId: normalizeText(input.scope_id || input.scopeId) || null,
    ownerScope: normalizeText(input.owner_scope || input.ownerScope),
    ownerTenantId: normalizeText(input.owner_tenant_id || input.ownerTenantId) || null,
    ownerTenantName: normalizeText(input.owner_tenant_name || input.ownerTenantName),
    matchMode: normalizeText(input.match_mode || input.matchMode || 'manual'),
    status: normalizeText(input.status || 'normal'),
    editable: Boolean(input.editable ?? true),
    memberCount: Number(input.member_count || input.memberCount || 0),
    estimatedCount: Number(input.estimated_count || input.estimatedCount || input.member_count || input.memberCount || 0),
    targets: Array.isArray(input.targets)
      ? (input.targets as unknown[]).map((item) => normalizeRecipientTarget(item as Record<string, unknown>))
      : [],
    meta: ((input.meta || {}) as Record<string, unknown>) || {},
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeMessageRecord(input: Record<string, unknown>): MessageRecord {
  return {
    id: normalizeText(input.id || input.record_id),
    title: normalizeText(input.title),
    summary: normalizeText(input.summary),
    messageType: normalizeText(input.message_type || input.messageType),
    scopeType: normalizeText(input.scope_type || input.scopeType),
    scopeId: normalizeText(input.scope_id || input.scopeId) || null,
    targetTenantId: normalizeText(input.target_tenant_id || input.targetTenantId) || null,
    targetTenantName: normalizeText(input.target_tenant_name || input.targetTenantName),
    senderName: normalizeText(input.sender_name || input.senderName),
    templateName: normalizeText(input.template_name || input.templateName),
    audienceType: normalizeText(input.audience_type || input.audienceType),
    priority: normalizeText(input.priority || 'normal'),
    status: normalizeText(input.status || 'normal'),
    publishedAt: normalizeText(input.published_at || input.publishedAt),
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
    deliveryCount: Number(input.delivery_count || input.deliveryCount || 0),
    readCount: Number(input.read_count || input.readCount || 0),
    unreadCount: Number(input.unread_count || input.unreadCount || 0),
    pendingTodoCount: Number(input.pending_todo_count || input.pendingTodoCount || 0),
  }
}

function normalizeDelivery(input: Record<string, unknown>): MessageDeliveryRecord {
  return {
    id: normalizeText(input.id),
    recipientUserId: normalizeText(input.recipient_user_id || input.recipientUserId),
    recipientName: normalizeText(input.recipient_name || input.recipientName),
    recipientTeamId: normalizeText(input.recipient_team_id || input.recipientTeamId) || null,
    recipientTeamName: normalizeText(input.recipient_team_name || input.recipientTeamName),
    deliveryStatus: normalizeText(input.delivery_status || input.deliveryStatus || 'queued'),
    todoStatus: normalizeText(input.todo_status || input.todoStatus),
    readAt: normalizeText(input.read_at || input.readAt),
    doneAt: normalizeText(input.done_at || input.doneAt),
    lastActionAt: normalizeText(input.last_action_at || input.lastActionAt),
    sourceGroupId: normalizeText(input.source_group_id || input.sourceGroupId) || null,
    sourceGroupName: normalizeText(input.source_group_name || input.sourceGroupName),
    sourceRuleType: normalizeText(input.source_rule_type || input.sourceRuleType),
    sourceRuleLabel: normalizeText(input.source_rule_label || input.sourceRuleLabel),
    sourceTargetId: normalizeText(input.source_target_id || input.sourceTargetId) || null,
    sourceTargetType: normalizeText(input.source_target_type || input.sourceTargetType),
    sourceTargetValue: normalizeText(input.source_target_value || input.sourceTargetValue),
  }
}

function buildMessageTimeline(detail: MessageRecordDetail): MessageTimelineItem[] {
  const items: MessageTimelineItem[] = [
    {
      id: 'created',
      label: '创建时间',
      value: detail.createdAt || '未记录',
      tone: 'neutral',
    },
  ]

  if (detail.publishedAt) {
    items.push({
      id: 'published',
      label: '发布时间',
      value: detail.publishedAt,
      tone: 'brand',
    })
  }

  const firstReadAt = detail.deliveries.find((item) => item.readAt)?.readAt
  if (firstReadAt) {
    items.push({
      id: 'first-read',
      label: '首次已读',
      value: firstReadAt,
      tone: 'success',
    })
  }

  const latestActionAt = detail.deliveries.find((item) => item.lastActionAt)?.lastActionAt
  if (latestActionAt) {
    items.push({
      id: 'latest-action',
      label: '最近动作',
      value: latestActionAt,
      tone: 'warning',
    })
  }

  return items
}

function normalizeMessageRecordDetail(input: Record<string, unknown>): MessageRecordDetail {
  const detail: MessageRecordDetail = {
    ...normalizeMessageRecord(input),
    content: normalizeText(input.content),
    deliverySummary: [],
    timeline: [],
    deliveries: Array.isArray(input.deliveries)
      ? (input.deliveries as Record<string, unknown>[]).map((item) => normalizeDelivery(item))
      : [],
    payload: ((input.payload || {}) as Record<string, unknown>) || {},
  }

  detail.deliverySummary = [
    { label: '投递总数', value: `${detail.deliveryCount}` },
    { label: '已读', value: `${detail.readCount}` },
    { label: '未读', value: `${detail.unreadCount}` },
    { label: '待办处理中', value: `${detail.pendingTodoCount}` },
  ]
  detail.timeline = buildMessageTimeline(detail)

  return detail
}

function resolveAudienceLabel(audienceType: string) {
  switch (audienceType) {
    case 'all_users':
      return '全体用户'
    case 'tenant_admins':
      return '团队管理员'
    case 'tenant_users':
      return '当前团队成员'
    case 'specified_users':
      return '指定用户'
    case 'recipient_group':
      return '收件组'
    case 'role':
      return '角色规则'
    case 'feature_package':
      return '功能包规则'
    default:
      return audienceType || '未指定受众'
  }
}

function resolveAudienceSummary(payload: MessageDispatchPayload) {
  const label = resolveAudienceLabel(payload.audienceType)
  const targetCount = payload.targetIds.length

  switch (payload.audienceType) {
    case 'all_users':
      return label
    case 'tenant_admins':
    case 'tenant_users':
      return targetCount > 0 ? `${label} · ${targetCount} 个团队` : label
    case 'specified_users':
      return `${label} · ${targetCount} 个目标`
    case 'recipient_group':
    case 'role':
    case 'feature_package':
      return `${label} · ${targetCount} 个规则`
    default:
      return targetCount > 0 ? `${label} · ${targetCount} 个目标` : label
  }
}

function normalizeDispatchResult(
  input: Record<string, unknown>,
  payload: MessageDispatchPayload,
): MessageDispatchResult {
  const deliveryCount = Number(input.delivery_count || input.deliveryCount || 0)
  const targetCount = payload.targetIds.length
  return {
    messageId: normalizeText(input.message_id || input.messageId),
    recordId: normalizeText(input.message_id || input.messageId),
    dispatchStatus: normalizeText(input.dispatch_status || input.dispatchStatus || 'queued'),
    deliveryCount,
    title: payload.title,
    audienceSummary: resolveAudienceSummary(payload),
    messageType: payload.messageType,
    priority: payload.priority,
    targetCount,
  }
}

function buildDispatchRequestPayload(payload: MessageDispatchPayload) {
  const targetIds = payload.targetIds.filter(Boolean)

  return {
    sender_id: payload.senderId,
    template_id: payload.templateId || undefined,
    message_type: payload.messageType,
    audience_type: payload.audienceType,
    title: payload.title,
    summary: payload.summary,
    content: payload.content,
    target_tenant_ids:
      payload.audienceType === 'tenant_admins' || payload.audienceType === 'tenant_users'
        ? targetIds
        : undefined,
    target_user_ids: payload.audienceType === 'specified_users' ? targetIds : undefined,
    target_group_ids:
      payload.audienceType === 'recipient_group' ||
      payload.audienceType === 'role' ||
      payload.audienceType === 'feature_package'
        ? targetIds
        : undefined,
    priority: payload.priority,
  }
}

export async function fetchInboxSummary() {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/messages/inbox/summary',
  })

  return normalizeInboxSummary(result)
}

export async function fetchInboxList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/messages/inbox',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeInboxThread(item)),
  }
}

export async function fetchInboxDetail(deliveryId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/messages/inbox/${deliveryId}`,
  })

  return normalizeInboxDetail(result)
}

export async function markInboxRead(deliveryId: string) {
  await requestData<unknown>({
    method: 'POST',
    url: `/api/v1/messages/inbox/${deliveryId}/read`,
  })
}

export async function markInboxReadAll(boxType?: string) {
  await requestData<unknown>({
    method: 'POST',
    url: '/api/v1/messages/inbox/read-all',
    params: boxType ? { box_type: boxType } : undefined,
  })
}

export async function handleInboxTodo(deliveryId: string, payload: Record<string, unknown>) {
  await requestData<unknown>({
    method: 'POST',
    url: `/api/v1/messages/inbox/${deliveryId}/todo-action`,
    data: payload,
  })
}

export async function fetchMessageDispatchOptions(scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'GET',
    url: '/api/v1/messages/dispatch/options',
  }))

  return {
    senderScope: normalizeText(result.sender_scope || result.senderScope),
    currentTenantId: normalizeText(result.current_tenant_id || result.currentTenantId),
    currentTenantName: normalizeText(result.current_tenant_name || result.currentTenantName),
    senderOptions: Array.isArray(result.sender_options || result.senderOptions)
      ? ((result.sender_options || result.senderOptions) as unknown[]).map((item) =>
          normalizeMessageSender(item as Record<string, unknown>),
        )
      : [],
    defaultSenderId: normalizeText(result.default_sender_id || result.defaultSenderId),
    audienceOptions: Array.isArray(result.audience_options || result.audienceOptions)
      ? ((result.audience_options || result.audienceOptions) as unknown[]).map((item) => ({
          value: normalizeText((item as Record<string, unknown>).value),
          label: normalizeText((item as Record<string, unknown>).label),
          description: normalizeText((item as Record<string, unknown>).description),
        }))
      : [],
    templateOptions: Array.isArray(result.template_options || result.templateOptions)
      ? ((result.template_options || result.templateOptions) as unknown[]).map((item) =>
          normalizeMessageTemplate(item as Record<string, unknown>),
        )
      : [],
    teams: Array.isArray(result.teams)
      ? (result.teams as unknown[]).map((item) => normalizeAudienceSubject(item as Record<string, unknown>))
      : [],
    users: Array.isArray(result.users)
      ? (result.users as unknown[]).map((item) => normalizeAudienceSubject(item as Record<string, unknown>))
      : [],
    recipientGroups: Array.isArray(result.recipient_groups || result.recipientGroups)
      ? ((result.recipient_groups || result.recipientGroups) as unknown[]).map((item) =>
          normalizeMessageRecipientGroup(item as Record<string, unknown>),
        )
      : [],
    roles: Array.isArray(result.roles)
      ? (result.roles as unknown[]).map((item) => normalizeAudienceSubject(item as Record<string, unknown>))
      : [],
    featurePackages: Array.isArray(result.feature_packages || result.featurePackages)
      ? ((result.feature_packages || result.featurePackages) as unknown[]).map((item) =>
          normalizeAudienceSubject(item as Record<string, unknown>),
        )
      : [],
    defaultMessageType: normalizeText(result.default_message_type || result.defaultMessageType || 'notice'),
    defaultAudienceType: normalizeText(result.default_audience_type || result.defaultAudienceType || 'all_users'),
    defaultPriority: normalizeText(result.default_priority || result.defaultPriority || 'normal'),
    supportsExternalLink: Boolean(result.supports_external_link ?? result.supportsExternalLink),
  } satisfies MessageDispatchOptions
}

export async function dispatchMessage(payload: MessageDispatchPayload, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'POST',
    url: '/api/v1/messages/dispatch',
    data: buildDispatchRequestPayload(payload),
  }))

  return normalizeDispatchResult(result, payload)
}

export async function fetchMessageTemplateList(scope: MessageScope, params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>(resolveScopeConfig(scope, {
    method: 'GET',
    url: '/api/v1/messages/templates',
    params,
  }))

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeMessageTemplate(item)),
  }
}

export async function createMessageTemplate(payload: MessageTemplateSavePayload, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'POST',
    url: '/api/v1/messages/templates',
    data: {
      template_key: payload.templateKey,
      name: payload.name,
      description: payload.description,
      message_type: payload.messageType,
      audience_type: payload.audienceType,
      title_template: payload.titleTemplate,
      summary_template: payload.summaryTemplate,
      content_template: payload.contentTemplate,
      status: payload.status,
    },
  }))

  return normalizeMessageTemplate(result)
}

export async function updateMessageTemplate(templateId: string, payload: MessageTemplateSavePayload, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'PUT',
    url: `/api/v1/messages/templates/${templateId}`,
    data: {
      template_key: payload.templateKey,
      name: payload.name,
      description: payload.description,
      message_type: payload.messageType,
      audience_type: payload.audienceType,
      title_template: payload.titleTemplate,
      summary_template: payload.summaryTemplate,
      content_template: payload.contentTemplate,
      status: payload.status,
    },
  }))

  return normalizeMessageTemplate(result)
}

export async function fetchMessageSenderList(scope: MessageScope) {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
  }>(resolveScopeConfig(scope, {
    method: 'GET',
    url: '/api/v1/messages/senders',
  }))

  return Array.isArray(result.records)
    ? result.records.map((item) => normalizeMessageSender(item))
    : []
}

export async function createMessageSender(payload: MessageSenderSavePayload, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'POST',
    url: '/api/v1/messages/senders',
    data: {
      name: payload.name,
      description: payload.description,
      avatar_url: payload.avatarUrl || undefined,
      is_default: payload.isDefault,
      status: payload.status,
    },
  }))

  return normalizeMessageSender(result)
}

export async function updateMessageSender(senderId: string, payload: MessageSenderSavePayload, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'PUT',
    url: `/api/v1/messages/senders/${senderId}`,
    data: {
      name: payload.name,
      description: payload.description,
      avatar_url: payload.avatarUrl || undefined,
      is_default: payload.isDefault,
      status: payload.status,
    },
  }))

  return normalizeMessageSender(result)
}

export async function fetchMessageRecipientGroupList(scope: MessageScope) {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
  }>(resolveScopeConfig(scope, {
    method: 'GET',
    url: '/api/v1/messages/recipient-groups',
  }))

  return Array.isArray(result.records)
    ? result.records.map((item) => normalizeMessageRecipientGroup(item))
    : []
}

export async function createMessageRecipientGroup(payload: MessageRecipientGroupSavePayload, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'POST',
    url: '/api/v1/messages/recipient-groups',
    data: {
      group_key: payload.groupKey,
      name: payload.name,
      description: payload.description,
      status: payload.status,
    },
  }))

  return normalizeMessageRecipientGroup(result)
}

export async function updateMessageRecipientGroup(groupId: string, payload: MessageRecipientGroupSavePayload, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'PUT',
    url: `/api/v1/messages/recipient-groups/${groupId}`,
    data: {
      group_key: payload.groupKey,
      name: payload.name,
      description: payload.description,
      status: payload.status,
    },
  }))

  return normalizeMessageRecipientGroup(result)
}

export async function fetchMessageRecordList(scope: MessageScope, params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>> & {
    summary?: Record<string, unknown>
  }>(resolveScopeConfig(scope, {
    method: 'GET',
    url: '/api/v1/messages/records',
    params,
  }))

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeMessageRecord(item)),
    summary: ((result.summary || {}) as Record<string, unknown>) || {},
  }
}

export async function fetchMessageRecordDetail(recordId: string, scope: MessageScope) {
  const result = await requestData<Record<string, unknown>>(resolveScopeConfig(scope, {
    method: 'GET',
    url: `/api/v1/messages/records/${recordId}`,
  }))

  return normalizeMessageRecordDetail(result)
}
