import { v5Client, unwrap, type V5Query, type V5RequestBody } from '@/domains/governance/api/_shared'
import type { components } from '@/api/v5/schema'

interface MessageRequestOptions {
  skipAuthWorkspaceHeader?: boolean
  skipCollaborationWorkspaceHeader?: boolean
}

type V5InboxItem = components['schemas']['InboxItem']
type V5InboxSummary = components['schemas']['InboxSummary']
type V5DispatchOptions = components['schemas']['MessageDispatchOptions']
type V5DispatchResult = components['schemas']['MessageDispatchResult']
type V5MessageTemplateItem = components['schemas']['MessageTemplateItem']
type V5MessageSenderItem = components['schemas']['MessageSenderItem']
type V5MessageRecipientGroupItem = components['schemas']['MessageRecipientGroupItem']
type V5DispatchRecordItem = components['schemas']['DispatchRecordItem']
type V5DispatchRecordDetail = components['schemas']['MessageDispatchRecord']

function normalizeInboxItem(item: V5InboxItem | undefined): Api.Message.InboxItem {
  return {
    id: item?.id || '',
    message_id: item?.message_id || '',
    box_type: (item?.box_type || 'notice') as Api.Message.BoxType,
    delivery_status: item?.delivery_status || 'unread',
    todo_status: item?.todo_status || '',
    read_at: item?.read_at || '',
    done_at: item?.done_at || '',
    last_action_at: item?.last_action_at || '',
    recipient_collaboration_workspace_id: item?.recipient_collaboration_workspace_id || '',
    title: item?.title || '',
    summary: item?.summary || '',
    content: item?.content || '',
    priority: item?.priority || '',
    action_type: item?.action_type || 'none',
    action_target: item?.action_target || '',
    message_type: (item?.message_type || item?.box_type || 'notice') as Api.Message.BoxType,
    biz_type: item?.biz_type || '',
    scope_type: item?.scope_type || '',
    scope_id: item?.scope_id || '',
    sender_type: item?.sender_type || '',
    sender_name_snapshot: item?.sender_name_snapshot || '',
    sender_avatar_snapshot: item?.sender_avatar_snapshot || '',
    sender_service_key: item?.sender_service_key || '',
    audience_type: item?.audience_type || '',
    audience_scope: item?.audience_scope || '',
    target_collaboration_workspace_id: item?.target_collaboration_workspace_id || '',
    published_at: item?.published_at || '',
    expired_at: item?.expired_at || '',
    created_at: item?.created_at || '',
    meta: item?.meta || {}
  }
}

function normalizeInboxSummary(item: V5InboxSummary | undefined): Api.Message.InboxSummary {
  return {
    unread_total: Number(item?.unread_total ?? 0),
    notice_count: Number(item?.notice_count ?? 0),
    message_count: Number(item?.message_count ?? 0),
    todo_count: Number(item?.todo_count ?? 0)
  }
}

function normalizeDispatchOptions(
  item: V5DispatchOptions | undefined
): Api.Message.DispatchOptions {
  return {
    sender_scope: item?.sender_scope || 'personal',
    current_collaboration_workspace_id: item?.current_collaboration_workspace_id || '',
    current_collaboration_workspace_name: item?.current_collaboration_workspace_name || '',
    sender_options: Array.isArray(item?.sender_options) ? item.sender_options : [],
    default_sender_id: item?.default_sender_id || '',
    audience_options: Array.isArray(item?.audience_options) ? item.audience_options : [],
    template_options: (Array.isArray(item?.template_options)
      ? item.template_options
      : []) as Api.Message.DispatchTemplateOption[],
    collaboration_workspaces: Array.isArray(item?.collaboration_workspaces)
      ? item.collaboration_workspaces
      : [],
    collaborationWorkspaces: Array.isArray(item?.collaboration_workspaces)
      ? item.collaboration_workspaces
      : [],
    users: Array.isArray(item?.users) ? item.users : [],
    recipient_groups: Array.isArray(item?.recipient_groups) ? item.recipient_groups : [],
    roles: Array.isArray(item?.roles) ? item.roles : [],
    feature_packages: Array.isArray(item?.feature_packages) ? item.feature_packages : [],
    default_message_type: (item?.default_message_type || 'notice') as Api.Message.BoxType,
    default_audience_type: item?.default_audience_type || 'all_users',
    default_priority: item?.default_priority || '',
    supports_external_link: Boolean(item?.supports_external_link)
  }
}

function normalizeDispatchResult(item: V5DispatchResult | undefined): Api.Message.DispatchResult {
  return {
    message_id: item?.message_id || '',
    delivery_count: Number(item?.delivery_count ?? 0),
    dispatch_status: item?.dispatch_status || 'queued'
  }
}

function normalizeMessageTemplateItem(
  item: V5MessageTemplateItem | undefined
): Api.Message.MessageTemplateItem {
  return {
    id: item?.id || '',
    template_key: item?.template_key || '',
    name: item?.name || '',
    description: item?.description || '',
    message_type: (item?.message_type || 'notice') as Api.Message.BoxType,
    owner_scope: item?.owner_scope || 'personal',
    owner_collaboration_workspace_id: item?.owner_collaboration_workspace_id || '',
    owner_collaboration_workspace_name: item?.owner_collaboration_workspace_name || '',
    audience_type: item?.audience_type || 'all_users',
    title_template: item?.title_template || '',
    summary_template: item?.summary_template || '',
    content_template: item?.content_template || '',
    status: item?.status || 'normal',
    editable: Boolean(item?.editable ?? true),
    meta: item?.meta || {},
    created_at: item?.created_at || '',
    updated_at: item?.updated_at || ''
  }
}

function normalizeMessageSenderItem(
  item: V5MessageSenderItem | undefined
): Api.Message.MessageSenderItem {
  return {
    id: item?.id || '',
    scope_type: item?.scope_type || 'personal',
    scope_id: item?.scope_id || '',
    name: item?.name || '',
    description: item?.description || '',
    avatar_url: item?.avatar_url || '',
    is_default: Boolean(item?.is_default),
    status: item?.status || 'normal',
    editable: Boolean(item?.editable ?? true),
    meta: item?.meta || {},
    created_at: item?.created_at || '',
    updated_at: item?.updated_at || ''
  }
}

function normalizeMessageRecipientGroupItem(
  item: V5MessageRecipientGroupItem | undefined
): Api.Message.MessageRecipientGroupItem {
  return {
    id: item?.id || '',
    scope_type: item?.scope_type || 'personal',
    scope_id: item?.scope_id || '',
    name: item?.name || '',
    description: item?.description || '',
    match_mode: item?.match_mode || 'manual',
    status: item?.status || 'normal',
    editable: Boolean(item?.editable ?? true),
    estimated_count: Number(item?.estimated_count ?? 0),
    meta: item?.meta || {},
    targets: Array.isArray(item?.targets) ? item.targets : [],
    created_at: item?.created_at || '',
    updated_at: item?.updated_at || ''
  }
}

function normalizeDispatchRecordItem(
  item: V5DispatchRecordItem | undefined
): Api.Message.DispatchRecordItem {
  return {
    id: item?.id || '',
    title: item?.title || '',
    summary: item?.summary || '',
    content: item?.content || '',
    message_type: (item?.message_type || 'notice') as Api.Message.BoxType,
    audience_type: item?.audience_type || 'all_users',
    scope_type: item?.scope_type || 'personal',
    scope_id: item?.scope_id || '',
    target_collaboration_workspace_id: item?.target_collaboration_workspace_id || '',
    target_collaboration_workspace_name: item?.target_collaboration_workspace_name || '',
    sender_name: item?.sender_name || '',
    template_name: item?.template_name || '',
    priority: item?.priority || '',
    status: item?.status || '',
    published_at: item?.published_at || '',
    created_at: item?.created_at || '',
    delivery_count: Number(item?.delivery_count ?? 0),
    read_count: Number(item?.read_count ?? 0),
    unread_count: Number(item?.unread_count ?? 0),
    pending_todo_count: Number(item?.pending_todo_count ?? 0)
  }
}

function normalizeDispatchRecordDetail(
  item: V5DispatchRecordDetail | undefined
): Api.Message.DispatchRecordDetail {
  return {
    ...normalizeDispatchRecordItem(item),
    deliveries: Array.isArray(item?.deliveries) ? item.deliveries : []
  }
}

export async function fetchGetInboxSummary(options?: { signal?: AbortSignal }) {
  const res = await unwrap(v5Client.GET('/messages/inbox/summary', { signal: options?.signal }))
  return normalizeInboxSummary(res)
}

export async function fetchGetInboxList(params: Api.Message.InboxQuery) {
  const query: V5Query<'/messages/inbox', 'get'> = params
  const res = await unwrap(v5Client.GET('/messages/inbox', { params: { query } }))
  return {
    records: (res.records || []).map(normalizeInboxItem),
    total: Number(res.total || 0),
    current: res.current,
    size: res.size
  } as Api.Message.InboxListResponse
}

export async function fetchGetInboxDetail(deliveryId: string) {
  const res = await unwrap(
    v5Client.GET('/messages/inbox/{deliveryId}', { params: { path: { deliveryId } } })
  )
  return normalizeInboxItem(res)
}

export async function fetchMarkInboxRead(deliveryId: string) {
  const { error } = await v5Client.POST('/messages/inbox/{deliveryId}/read', {
    params: { path: { deliveryId } }
  })
  if (error) throw error
}

export async function fetchMarkInboxReadAll(boxType?: Api.Message.BoxType | '') {
  void boxType
  const { error } = await v5Client.POST('/messages/inbox/read-all')
  if (error) throw error
}

export async function fetchHandleInboxTodo(
  deliveryId: string,
  params: Api.Message.TodoActionParams
) {
  const body: V5RequestBody<'/messages/inbox/{deliveryId}/todo-action', 'post'> = {
    action: params.action
  }
  const { error } = await v5Client.POST('/messages/inbox/{deliveryId}/todo-action', {
    params: { path: { deliveryId } },
    body
  })
  if (error) throw error
}

export async function fetchGetMessageDispatchOptions(_options?: MessageRequestOptions) {
  const res = await unwrap(v5Client.GET('/messages/dispatch/options'))
  return normalizeDispatchOptions(res)
}

export async function fetchDispatchMessage(
  params: Api.Message.DispatchParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/dispatch', 'post'> = {
    sender_id: params.sender_id,
    template_id: params.template_id,
    template_key: params.template_key,
    message_type: params.message_type,
    audience_type: params.audience_type,
    target_collaboration_workspace_ids: params.target_collaboration_workspace_ids,
    target_user_ids: params.target_user_ids,
    target_group_ids: params.target_group_ids,
    title: params.title,
    summary: params.summary,
    content: params.content,
    priority: params.priority,
    action_type: params.action_type,
    action_target: params.action_target,
    biz_type: params.biz_type,
    expired_at: params.expired_at,
    dry_run: params.dry_run ?? undefined
  }
  const res = await unwrap(v5Client.POST('/messages/dispatch', { body }))
  return normalizeDispatchResult(res)
}

export async function fetchGetMessageTemplateList(
  params: Api.Message.MessageTemplateQuery,
  _options?: MessageRequestOptions
) {
  void params
  const res = await unwrap(v5Client.GET('/messages/templates'))
  return {
    records: (res.records || []).map(normalizeMessageTemplateItem),
    total: Number(res.total || 0),
    current: res.current,
    size: res.size
  } as Api.Message.MessageTemplateListResponse
}

export async function fetchCreateMessageTemplate(
  params: Api.Message.MessageTemplateSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/templates', 'post'> = {
    template_key: params.template_key,
    name: params.name,
    description: params.description,
    message_type: params.message_type,
    audience_type: params.audience_type,
    title_template: params.title_template,
    summary_template: params.summary_template,
    content_template: params.content_template,
    status: params.status
  }
  const res = await unwrap(v5Client.POST('/messages/templates', { body }))
  return normalizeMessageTemplateItem(res)
}

export async function fetchUpdateMessageTemplate(
  templateId: string,
  params: Api.Message.MessageTemplateSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/templates/{templateId}', 'put'> = {
    template_key: params.template_key,
    name: params.name,
    description: params.description,
    message_type: params.message_type,
    audience_type: params.audience_type,
    title_template: params.title_template,
    summary_template: params.summary_template,
    content_template: params.content_template,
    status: params.status
  }
  const res = await unwrap(
    v5Client.PUT('/messages/templates/{templateId}', {
      params: { path: { templateId } },
      body
    })
  )
  return normalizeMessageTemplateItem(res)
}

export async function fetchGetDispatchRecordList(
  params: Api.Message.DispatchRecordQuery,
  _options?: MessageRequestOptions
) {
  const query: V5Query<'/messages/records', 'get'> = params
  const res = await unwrap(v5Client.GET('/messages/records', { params: { query } }))
  return {
    records: (res.records || []).map(normalizeDispatchRecordItem),
    total: Number(res.total || 0),
    current: res.current,
    size: res.size,
    summary: res.summary || {
      total_messages: 0,
      total_deliveries: 0,
      read_deliveries: 0,
      todo_messages: 0
    }
  } as Api.Message.DispatchRecordListResponse
}

export async function fetchGetDispatchRecordDetail(
  recordId: string,
  _options?: MessageRequestOptions
) {
  const res = await unwrap(
    v5Client.GET('/messages/records/{recordId}', { params: { path: { recordId } } })
  )
  return normalizeDispatchRecordDetail(res)
}

export async function fetchGetMessageSenderList(_options?: MessageRequestOptions) {
  const res = await unwrap(v5Client.GET('/messages/senders'))
  return {
    records: (res.records || []).map(normalizeMessageSenderItem)
  } as Api.Message.MessageSenderListResponse
}

export async function fetchCreateMessageSender(
  params: Api.Message.MessageSenderSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/senders', 'post'> = {
    name: params.name,
    description: params.description,
    avatar_url: params.avatar_url,
    is_default: params.is_default,
    status: params.status,
    meta: params.meta
  }
  const res = await unwrap(v5Client.POST('/messages/senders', { body }))
  return normalizeMessageSenderItem(res)
}

export async function fetchUpdateMessageSender(
  senderId: string,
  params: Api.Message.MessageSenderSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/senders/{senderId}', 'put'> = {
    name: params.name,
    description: params.description,
    avatar_url: params.avatar_url,
    is_default: params.is_default,
    status: params.status,
    meta: params.meta
  }
  const res = await unwrap(
    v5Client.PUT('/messages/senders/{senderId}', {
      params: { path: { senderId } },
      body
    })
  )
  return normalizeMessageSenderItem(res)
}

export async function fetchGetMessageRecipientGroupList(_options?: MessageRequestOptions) {
  const res = await unwrap(v5Client.GET('/messages/recipient-groups'))
  return {
    records: (res.records || []).map(normalizeMessageRecipientGroupItem)
  } as Api.Message.MessageRecipientGroupListResponse
}

export async function fetchCreateMessageRecipientGroup(
  params: Api.Message.MessageRecipientGroupSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/recipient-groups', 'post'> = {
    name: params.name,
    description: params.description,
    match_mode: params.match_mode,
    status: params.status,
    meta: params.meta,
    targets: (params.targets || []).map((item) => ({
      target_type: item.target_type,
      user_id: item.user_id,
      collaboration_workspace_id: item.collaboration_workspace_id,
      role_code: item.role_code,
      package_key: item.package_key,
      sort_order: item.sort_order,
      meta: item.meta
    }))
  }
  const res = await unwrap(v5Client.POST('/messages/recipient-groups', { body }))
  return normalizeMessageRecipientGroupItem(res)
}

export async function fetchUpdateMessageRecipientGroup(
  groupId: string,
  params: Api.Message.MessageRecipientGroupSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/recipient-groups/{groupId}', 'put'> = {
    name: params.name,
    description: params.description,
    match_mode: params.match_mode,
    status: params.status,
    meta: params.meta,
    targets: (params.targets || []).map((item) => ({
      target_type: item.target_type,
      user_id: item.user_id,
      collaboration_workspace_id: item.collaboration_workspace_id,
      role_code: item.role_code,
      package_key: item.package_key,
      sort_order: item.sort_order,
      meta: item.meta
    }))
  }
  const res = await unwrap(
    v5Client.PUT('/messages/recipient-groups/{groupId}', {
      params: { path: { groupId } },
      body
    })
  )
  return normalizeMessageRecipientGroupItem(res)
}
