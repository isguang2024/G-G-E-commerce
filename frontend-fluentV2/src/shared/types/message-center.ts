export interface InboxSummary {
  unreadTotal: number
  noticeCount: number
  messageCount: number
  todoCount: number
}

export interface InboxThread {
  id: string
  title: string
  summary: string
  messageType: string
  boxType: string
  read: boolean
  todoStatus: string
  senderName: string
  senderAvatarUrl: string
  sentAt: string
  createdAt: string
  updatedAt: string
  tenantName: string
}

export interface InboxMessageDetail {
  id: string
  title: string
  summary: string
  content: string
  messageType: string
  boxType: string
  read: boolean
  todoStatus: string
  priority: string
  senderName: string
  senderAvatarUrl: string
  sentAt: string
  createdAt: string
  updatedAt: string
  tenantName: string
  payload: Record<string, unknown>
}

export interface MessageDetailField {
  label: string
  value: string
}

export interface MessageTimelineItem {
  id: string
  label: string
  value: string
  tone?: 'brand' | 'success' | 'warning' | 'danger' | 'neutral'
}

export interface DispatchOption {
  value: string
  label: string
  description: string
}

export interface DispatchSender {
  id: string
  name: string
  description: string
  avatarUrl: string
  isDefault: boolean
}

export interface DispatchAudienceSubject {
  id: string
  name: string
  displayName?: string
  description?: string
  teamId?: string | null
  teamName?: string
}

export interface MessageDispatchOptions {
  senderScope: string
  currentTenantId: string
  currentTenantName: string
  senderOptions: DispatchSender[]
  defaultSenderId: string
  audienceOptions: DispatchOption[]
  templateOptions: MessageTemplateRecord[]
  teams: DispatchAudienceSubject[]
  users: DispatchAudienceSubject[]
  recipientGroups: MessageRecipientGroupRecord[]
  roles: DispatchAudienceSubject[]
  featurePackages: DispatchAudienceSubject[]
  defaultMessageType: string
  defaultAudienceType: string
  defaultPriority: string
  supportsExternalLink: boolean
}

export interface MessageTemplateRecord {
  id: string
  templateKey: string
  name: string
  description: string
  messageType: string
  ownerScope: string
  ownerTenantId?: string | null
  ownerTenantName: string
  audienceType: string
  titleTemplate: string
  summaryTemplate: string
  contentTemplate: string
  status: string
  editable: boolean
  meta: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface MessageSenderRecord {
  id: string
  scopeType: string
  scopeId?: string | null
  name: string
  description: string
  avatarUrl: string
  isDefault: boolean
  status: string
  editable: boolean
  meta: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface MessageRecipientGroupRecord {
  id: string
  groupKey: string
  name: string
  description: string
  scopeType: string
  scopeId?: string | null
  ownerScope: string
  ownerTenantId?: string | null
  ownerTenantName: string
  matchMode: string
  status: string
  editable: boolean
  memberCount: number
  estimatedCount: number
  targets: MessageRecipientGroupTarget[]
  meta: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface MessageRecipientGroupTarget {
  id: string
  targetType: string
  targetLabel: string
  targetValue: string
  sortOrder: number
  meta: Record<string, unknown>
}

export interface MessageRecord {
  id: string
  title: string
  summary: string
  messageType: string
  scopeType: string
  scopeId?: string | null
  targetTenantId?: string | null
  targetTenantName: string
  senderName: string
  templateName: string
  audienceType: string
  priority: string
  status: string
  publishedAt: string
  createdAt: string
  updatedAt: string
  deliveryCount: number
  readCount: number
  unreadCount: number
  pendingTodoCount: number
}

export interface MessageDeliveryRecord {
  id: string
  recipientUserId: string
  recipientName: string
  recipientTeamId?: string | null
  recipientTeamName: string
  deliveryStatus: string
  todoStatus: string
  readAt: string
  doneAt: string
  lastActionAt: string
  sourceGroupId?: string | null
  sourceGroupName: string
  sourceRuleType: string
  sourceRuleLabel: string
  sourceTargetId?: string | null
  sourceTargetType: string
  sourceTargetValue: string
}

export interface MessageRecordDetail extends MessageRecord {
  content: string
  deliverySummary: MessageDetailField[]
  timeline: MessageTimelineItem[]
  deliveries: MessageDeliveryRecord[]
  payload: Record<string, unknown>
}

export interface MessageDispatchResult {
  messageId: string
  recordId: string
  dispatchStatus: string
  deliveryCount: number
  title: string
  audienceSummary: string
  messageType: string
  priority: string
  targetCount: number
}

export interface MessageTemplateSavePayload {
  templateKey: string
  name: string
  description: string
  messageType: string
  audienceType: string
  titleTemplate: string
  summaryTemplate: string
  contentTemplate: string
  status: string
}

export interface MessageSenderSavePayload {
  name: string
  description: string
  avatarUrl?: string
  isDefault: boolean
  status: string
}

export interface MessageRecipientGroupSavePayload {
  groupKey: string
  name: string
  description: string
  status: string
}

export interface MessageDispatchPayload {
  senderId: string
  templateId?: string
  messageType: string
  audienceType: string
  title: string
  summary: string
  content: string
  targetIds: string[]
  priority: string
}
