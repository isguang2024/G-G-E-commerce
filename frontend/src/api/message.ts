import { v5Client, unwrap } from '@/api/system-manage/_shared'

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
  return unwrap(v5Client.GET('/messages/inbox/summary', { params: { query: {} as any } })) as unknown as Promise<Api.Message.InboxSummary>
}

export function fetchGetInboxList(params: Api.Message.InboxQuery) {
  return unwrap(
    v5Client.GET('/messages/inbox', { params: { query: params as any } })
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
  const { error } = await v5Client.POST('/messages/inbox/read-all', {
    params: { query: { box_type: boxType || undefined } as any }
  })
  if (error) throw error
}

export async function fetchHandleInboxTodo(deliveryId: string, params: Api.Message.TodoActionParams) {
  const { error } = await v5Client.POST('/messages/inbox/{deliveryId}/todo-action', {
    params: { path: { deliveryId } },
    body: params as any
  })
  if (error) throw error
}

export function fetchGetMessageDispatchOptions(_options?: MessageRequestOptions) {
  return unwrap(
    v5Client.GET('/messages/dispatch/options', { params: { query: {} as any } })
  ) as unknown as Promise<Api.Message.DispatchOptions>
}

export function fetchDispatchMessage(
  params: Api.Message.DispatchParams,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.POST('/messages/dispatch', { body: params as any })
  ) as unknown as Promise<Api.Message.DispatchResult>
}

export function fetchGetMessageTemplateList(
  params: Api.Message.MessageTemplateQuery,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.GET('/messages/templates', { params: { query: params as any } })
  ) as unknown as Promise<Api.Message.MessageTemplateListResponse>
}

export function fetchCreateMessageTemplate(
  params: Api.Message.MessageTemplateSaveParams,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.POST('/messages/templates', { body: params as any })
  ) as unknown as Promise<Api.Message.MessageTemplateItem>
}

export function fetchUpdateMessageTemplate(
  templateId: string,
  params: Api.Message.MessageTemplateSaveParams,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.PUT('/messages/templates/{templateId}', {
      params: { path: { templateId } },
      body: params as any
    })
  ) as unknown as Promise<Api.Message.MessageTemplateItem>
}

export function fetchGetDispatchRecordList(
  params: Api.Message.DispatchRecordQuery,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.GET('/messages/records', { params: { query: params as any } })
  ) as unknown as Promise<Api.Message.DispatchRecordListResponse>
}

export function fetchGetDispatchRecordDetail(recordId: string, _options?: MessageRequestOptions) {
  return unwrap(
    v5Client.GET('/messages/records/{recordId}', { params: { path: { recordId } } })
  ) as unknown as Promise<Api.Message.DispatchRecordDetail>
}

export function fetchGetMessageSenderList(_options?: MessageRequestOptions) {
  return unwrap(
    v5Client.GET('/messages/senders', { params: { query: {} as any } })
  ) as unknown as Promise<Api.Message.MessageSenderListResponse>
}

export function fetchCreateMessageSender(
  params: Api.Message.MessageSenderSaveParams,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.POST('/messages/senders', { body: params as any })
  ) as unknown as Promise<Api.Message.MessageSenderItem>
}

export function fetchUpdateMessageSender(
  senderId: string,
  params: Api.Message.MessageSenderSaveParams,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.PUT('/messages/senders/{senderId}', {
      params: { path: { senderId } },
      body: params as any
    })
  ) as unknown as Promise<Api.Message.MessageSenderItem>
}

export function fetchGetMessageRecipientGroupList(_options?: MessageRequestOptions) {
  return unwrap(
    v5Client.GET('/messages/recipient-groups', { params: { query: {} as any } })
  ) as unknown as Promise<Api.Message.MessageRecipientGroupListResponse>
}

export function fetchCreateMessageRecipientGroup(
  params: Api.Message.MessageRecipientGroupSaveParams,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.POST('/messages/recipient-groups', { body: params as any })
  ) as unknown as Promise<Api.Message.MessageRecipientGroupItem>
}

export function fetchUpdateMessageRecipientGroup(
  groupId: string,
  params: Api.Message.MessageRecipientGroupSaveParams,
  _options?: MessageRequestOptions
) {
  return unwrap(
    v5Client.PUT('/messages/recipient-groups/{groupId}', {
      params: { path: { groupId } },
      body: params as any
    })
  ) as unknown as Promise<Api.Message.MessageRecipientGroupItem>
}
