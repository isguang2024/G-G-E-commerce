import request from '@/utils/http'

const MESSAGE_INBOX_BASE = '/api/v1/messages/inbox'
const MESSAGE_DISPATCH_BASE = '/api/v1/messages/dispatch'
const MESSAGE_TEMPLATE_BASE = '/api/v1/messages/templates'
const MESSAGE_SENDER_BASE = '/api/v1/messages/senders'
const MESSAGE_RECIPIENT_GROUP_BASE = '/api/v1/messages/recipient-groups'
const MESSAGE_RECORD_BASE = '/api/v1/messages/records'

interface MessageRequestOptions {
  skipAuthWorkspaceHeader?: boolean
  skipCollaborationWorkspaceHeader?: boolean
}

export function fetchGetInboxSummary() {
  return request.get<Api.Message.InboxSummary>({
    url: `${MESSAGE_INBOX_BASE}/summary`
  })
}

export function fetchGetInboxList(params: Api.Message.InboxQuery) {
  return request.get<Api.Message.InboxListResponse>({
    url: MESSAGE_INBOX_BASE,
    params
  })
}

export function fetchGetInboxDetail(deliveryId: string) {
  return request.get<Api.Message.InboxDetail>({
    url: `${MESSAGE_INBOX_BASE}/${deliveryId}`
  })
}

export function fetchMarkInboxRead(deliveryId: string) {
  return request.post<void>({
    url: `${MESSAGE_INBOX_BASE}/${deliveryId}/read`
  })
}

export function fetchMarkInboxReadAll(boxType?: Api.Message.BoxType | '') {
  return request.post<void>({
    url: `${MESSAGE_INBOX_BASE}/read-all`,
    params: {
      box_type: boxType || undefined
    }
  })
}

export function fetchHandleInboxTodo(deliveryId: string, params: Api.Message.TodoActionParams) {
  return request.post<void>({
    url: `${MESSAGE_INBOX_BASE}/${deliveryId}/todo-action`,
    data: params
  })
}

export function fetchGetMessageDispatchOptions(options?: MessageRequestOptions) {
  return request.get<Api.Message.DispatchOptions>({
    url: `${MESSAGE_DISPATCH_BASE}/options`,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader
  })
}

export function fetchDispatchMessage(
  params: Api.Message.DispatchParams,
  options?: MessageRequestOptions
) {
  return request.post<Api.Message.DispatchResult>({
    url: MESSAGE_DISPATCH_BASE,
    data: params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader,
    showSuccessMessage: true
  })
}

export function fetchGetMessageTemplateList(
  params: Api.Message.MessageTemplateQuery,
  options?: MessageRequestOptions
) {
  return request.get<Api.Message.MessageTemplateListResponse>({
    url: MESSAGE_TEMPLATE_BASE,
    params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader
  })
}

export function fetchCreateMessageTemplate(
  params: Api.Message.MessageTemplateSaveParams,
  options?: MessageRequestOptions
) {
  return request.post<Api.Message.MessageTemplateItem>({
    url: MESSAGE_TEMPLATE_BASE,
    data: params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader,
    showSuccessMessage: true
  })
}

export function fetchUpdateMessageTemplate(
  templateId: string,
  params: Api.Message.MessageTemplateSaveParams,
  options?: MessageRequestOptions
) {
  return request.put<Api.Message.MessageTemplateItem>({
    url: `${MESSAGE_TEMPLATE_BASE}/${templateId}`,
    data: params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader,
    showSuccessMessage: true
  })
}

export function fetchGetDispatchRecordList(
  params: Api.Message.DispatchRecordQuery,
  options?: MessageRequestOptions
) {
  return request.get<Api.Message.DispatchRecordListResponse>({
    url: MESSAGE_RECORD_BASE,
    params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader
  })
}

export function fetchGetDispatchRecordDetail(recordId: string, options?: MessageRequestOptions) {
  return request.get<Api.Message.DispatchRecordDetail>({
    url: `${MESSAGE_RECORD_BASE}/${recordId}`,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader
  })
}

export function fetchGetMessageSenderList(options?: MessageRequestOptions) {
  return request.get<Api.Message.MessageSenderListResponse>({
    url: MESSAGE_SENDER_BASE,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader
  })
}

export function fetchCreateMessageSender(
  params: Api.Message.MessageSenderSaveParams,
  options?: MessageRequestOptions
) {
  return request.post<Api.Message.MessageSenderItem>({
    url: MESSAGE_SENDER_BASE,
    data: params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader,
    showSuccessMessage: true
  })
}

export function fetchUpdateMessageSender(
  senderId: string,
  params: Api.Message.MessageSenderSaveParams,
  options?: MessageRequestOptions
) {
  return request.put<Api.Message.MessageSenderItem>({
    url: `${MESSAGE_SENDER_BASE}/${senderId}`,
    data: params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader,
    showSuccessMessage: true
  })
}

export function fetchGetMessageRecipientGroupList(options?: MessageRequestOptions) {
  return request.get<Api.Message.MessageRecipientGroupListResponse>({
    url: MESSAGE_RECIPIENT_GROUP_BASE,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader
  })
}

export function fetchCreateMessageRecipientGroup(
  params: Api.Message.MessageRecipientGroupSaveParams,
  options?: MessageRequestOptions
) {
  return request.post<Api.Message.MessageRecipientGroupItem>({
    url: MESSAGE_RECIPIENT_GROUP_BASE,
    data: params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader,
    showSuccessMessage: true
  })
}

export function fetchUpdateMessageRecipientGroup(
  groupId: string,
  params: Api.Message.MessageRecipientGroupSaveParams,
  options?: MessageRequestOptions
) {
  return request.put<Api.Message.MessageRecipientGroupItem>({
    url: `${MESSAGE_RECIPIENT_GROUP_BASE}/${groupId}`,
    data: params,
    skipAuthWorkspaceHeader: options?.skipAuthWorkspaceHeader,
    skipCollaborationWorkspaceHeader: options?.skipCollaborationWorkspaceHeader,
    showSuccessMessage: true
  })
}
