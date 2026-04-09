import {
  v5Client,
  unwrap,
  toV5Body,
  type V5Query,
  type V5RequestBody
} from '@/api/system-manage/_shared'

interface MessageRequestOptions {
  skipAuthWorkspaceHeader?: boolean
  skipCollaborationWorkspaceHeader?: boolean
}

// Note: skipAuthWorkspaceHeader/skipCollaborationWorkspaceHeader are legacy flags
// that were used to suppress header injection for specific calls. With v5Client
// the header middleware always injects based on store state; per-call escapes
// are no longer supported. The options arg is kept for signature compatibility
// but currently ignored.
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function _unused(_: MessageRequestOptions | undefined) {}

export function fetchGetInboxSummary() {
  return unwrap(v5Client.GET('/messages/inbox/summary')) as unknown as Promise<Api.Message.InboxSummary>
}

export function fetchGetInboxList(params: Api.Message.InboxQuery) {
  const query: V5Query<'/messages/inbox', 'get'> = params
  return unwrap(
    v5Client.GET('/messages/inbox', { params: { query } })
  ) as unknown as Promise<Api.Message.InboxListResponse>
}

export function fetchGetInboxDetail(deliveryId: string) {
  return unwrap(
    v5Client.GET('/messages/inbox/{deliveryId}', { params: { path: { deliveryId } } })
  ) as unknown as Promise<Api.Message.InboxDetail>
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

export async function fetchHandleInboxTodo(deliveryId: string, params: Api.Message.TodoActionParams) {
  const body: V5RequestBody<'/messages/inbox/{deliveryId}/todo-action', 'post'> = toV5Body(params)
  const { error } = await v5Client.POST('/messages/inbox/{deliveryId}/todo-action', {
    params: { path: { deliveryId } },
    body
  })
  if (error) throw error
}

export function fetchGetMessageDispatchOptions(_options?: MessageRequestOptions) {
  return unwrap(v5Client.GET('/messages/dispatch/options')) as unknown as Promise<Api.Message.DispatchOptions>
}

export function fetchDispatchMessage(
  params: Api.Message.DispatchParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/dispatch', 'post'> = toV5Body(params)
  return unwrap(
    v5Client.POST('/messages/dispatch', { body })
  ) as unknown as Promise<Api.Message.DispatchResult>
}

export function fetchGetMessageTemplateList(
  params: Api.Message.MessageTemplateQuery,
  _options?: MessageRequestOptions
) {
  void params
  return unwrap(
    v5Client.GET('/messages/templates')
  ) as unknown as Promise<Api.Message.MessageTemplateListResponse>
}

export function fetchCreateMessageTemplate(
  params: Api.Message.MessageTemplateSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/templates', 'post'> = toV5Body(params)
  return unwrap(
    v5Client.POST('/messages/templates', { body })
  ) as unknown as Promise<Api.Message.MessageTemplateItem>
}

export function fetchUpdateMessageTemplate(
  templateId: string,
  params: Api.Message.MessageTemplateSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/templates/{templateId}', 'put'> = toV5Body(params)
  return unwrap(
    v5Client.PUT('/messages/templates/{templateId}', {
      params: { path: { templateId } },
      body
    })
  ) as unknown as Promise<Api.Message.MessageTemplateItem>
}

export function fetchGetDispatchRecordList(
  params: Api.Message.DispatchRecordQuery,
  _options?: MessageRequestOptions
) {
  const query: V5Query<'/messages/records', 'get'> = params
  return unwrap(
    v5Client.GET('/messages/records', { params: { query } })
  ) as unknown as Promise<Api.Message.DispatchRecordListResponse>
}

export function fetchGetDispatchRecordDetail(recordId: string, _options?: MessageRequestOptions) {
  return unwrap(
    v5Client.GET('/messages/records/{recordId}', { params: { path: { recordId } } })
  ) as unknown as Promise<Api.Message.DispatchRecordDetail>
}

export function fetchGetMessageSenderList(_options?: MessageRequestOptions) {
  return unwrap(v5Client.GET('/messages/senders')) as unknown as Promise<Api.Message.MessageSenderListResponse>
}

export function fetchCreateMessageSender(
  params: Api.Message.MessageSenderSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/senders', 'post'> = toV5Body(params)
  return unwrap(
    v5Client.POST('/messages/senders', { body })
  ) as unknown as Promise<Api.Message.MessageSenderItem>
}

export function fetchUpdateMessageSender(
  senderId: string,
  params: Api.Message.MessageSenderSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/senders/{senderId}', 'put'> = toV5Body(params)
  return unwrap(
    v5Client.PUT('/messages/senders/{senderId}', {
      params: { path: { senderId } },
      body
    })
  ) as unknown as Promise<Api.Message.MessageSenderItem>
}

export function fetchGetMessageRecipientGroupList(_options?: MessageRequestOptions) {
  return unwrap(v5Client.GET('/messages/recipient-groups')) as unknown as Promise<Api.Message.MessageRecipientGroupListResponse>
}

export function fetchCreateMessageRecipientGroup(
  params: Api.Message.MessageRecipientGroupSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/recipient-groups', 'post'> = toV5Body(params)
  return unwrap(
    v5Client.POST('/messages/recipient-groups', { body })
  ) as unknown as Promise<Api.Message.MessageRecipientGroupItem>
}

export function fetchUpdateMessageRecipientGroup(
  groupId: string,
  params: Api.Message.MessageRecipientGroupSaveParams,
  _options?: MessageRequestOptions
) {
  const body: V5RequestBody<'/messages/recipient-groups/{groupId}', 'put'> = toV5Body(params)
  return unwrap(
    v5Client.PUT('/messages/recipient-groups/{groupId}', {
      params: { path: { groupId } },
      body
    })
  ) as unknown as Promise<Api.Message.MessageRecipientGroupItem>
}
