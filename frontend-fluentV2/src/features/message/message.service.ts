import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createMessageRecipientGroup,
  createMessageSender,
  createMessageTemplate,
  dispatchMessage,
  fetchMessageDispatchOptions,
  fetchMessageRecipientGroupList,
  fetchMessageRecordDetail,
  fetchMessageRecordList,
  fetchMessageSenderList,
  fetchMessageTemplateList,
  updateMessageRecipientGroup,
  updateMessageSender,
  updateMessageTemplate,
} from '@/shared/api/modules/message.api'
import { queryKeys } from '@/shared/api/query-keys'
import type {
  MessageDispatchPayload,
  MessageRecipientGroupSavePayload,
  MessageSenderSavePayload,
  MessageTemplateSavePayload,
} from '@/shared/types/message-center'

function serializeFilters(filters?: Record<string, unknown>) {
  return JSON.stringify(filters || {})
}

export function useMessageDispatchOptionsQuery() {
  return useQuery({
    queryKey: queryKeys.message.dispatchOptions('platform'),
    queryFn: () => fetchMessageDispatchOptions('platform'),
    placeholderData: (previousData) => previousData,
  })
}

export function useScopedMessageDispatchOptionsQuery(scope: 'platform' | 'team') {
  return useQuery({
    queryKey: queryKeys.message.dispatchOptions(scope),
    queryFn: () => fetchMessageDispatchOptions(scope),
    placeholderData: (previousData) => previousData,
  })
}

export function useMessageTemplatesQuery(scope: string, filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.message.templates(serializeFilters(filters), scope),
    queryFn: () => fetchMessageTemplateList(scope as 'platform' | 'team', filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useMessageSendersQuery(scope: string) {
  return useQuery({
    queryKey: queryKeys.message.senders(scope),
    queryFn: () => fetchMessageSenderList(scope as 'platform' | 'team'),
    placeholderData: (previousData) => previousData,
  })
}

export function useMessageRecipientGroupsQuery(scope: string) {
  return useQuery({
    queryKey: queryKeys.message.recipientGroups(scope),
    queryFn: () => fetchMessageRecipientGroupList(scope as 'platform' | 'team'),
    placeholderData: (previousData) => previousData,
  })
}

export function useMessageRecordsQuery(scope: string, filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.message.records(serializeFilters(filters), scope),
    queryFn: () => fetchMessageRecordList(scope as 'platform' | 'team', filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useMessageRecordDetailQuery(scope: string, recordId?: string | null) {
  return useQuery({
    queryKey: queryKeys.message.recordDetail(recordId || '', scope),
    queryFn: () => fetchMessageRecordDetail(recordId!, scope as 'platform' | 'team'),
    enabled: Boolean(recordId),
    placeholderData: (previousData) => previousData,
  })
}

function invalidateMessageQueries(client: ReturnType<typeof useQueryClient>, scope: string, recordId?: string) {
  return Promise.all([
    client.invalidateQueries({ queryKey: queryKeys.message.dispatchOptions(scope) }),
    client.invalidateQueries({ queryKey: ['message', 'templates', scope] }),
    client.invalidateQueries({ queryKey: queryKeys.message.senders(scope) }),
    client.invalidateQueries({ queryKey: queryKeys.message.recipientGroups(scope) }),
    client.invalidateQueries({ queryKey: ['message', 'records', scope] }),
    recordId ? client.invalidateQueries({ queryKey: queryKeys.message.recordDetail(recordId, scope) }) : Promise.resolve(),
  ])
}

export function useDispatchMessageMutation(scope: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: MessageDispatchPayload) => dispatchMessage(payload, scope as 'platform' | 'team'),
    onSuccess: async () => {
      await invalidateMessageQueries(client, scope)
    },
  })
}

export function useCreateMessageTemplateMutation(scope: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: MessageTemplateSavePayload) => createMessageTemplate(payload, scope as 'platform' | 'team'),
    onSuccess: async () => {
      await invalidateMessageQueries(client, scope)
    },
  })
}

export function useUpdateMessageTemplateMutation(scope: string, templateId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: MessageTemplateSavePayload) => updateMessageTemplate(templateId, payload, scope as 'platform' | 'team'),
    onSuccess: async () => {
      await invalidateMessageQueries(client, scope)
    },
  })
}

export function useCreateMessageSenderMutation(scope: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: MessageSenderSavePayload) => createMessageSender(payload, scope as 'platform' | 'team'),
    onSuccess: async () => {
      await invalidateMessageQueries(client, scope)
    },
  })
}

export function useUpdateMessageSenderMutation(scope: string, senderId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: MessageSenderSavePayload) => updateMessageSender(senderId, payload, scope as 'platform' | 'team'),
    onSuccess: async () => {
      await invalidateMessageQueries(client, scope)
    },
  })
}

export function useCreateMessageRecipientGroupMutation(scope: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: MessageRecipientGroupSavePayload) => createMessageRecipientGroup(payload, scope as 'platform' | 'team'),
    onSuccess: async () => {
      await invalidateMessageQueries(client, scope)
    },
  })
}

export function useUpdateMessageRecipientGroupMutation(scope: string, groupId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: MessageRecipientGroupSavePayload) => updateMessageRecipientGroup(groupId, payload, scope as 'platform' | 'team'),
    onSuccess: async () => {
      await invalidateMessageQueries(client, scope)
    },
  })
}
