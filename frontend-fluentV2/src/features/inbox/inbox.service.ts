import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchInboxDetail, fetchInboxList, fetchInboxSummary, handleInboxTodo, markInboxRead, markInboxReadAll } from '@/shared/api/modules/message.api'
import { queryKeys } from '@/shared/api/query-keys'

function serializeFilters(filters?: Record<string, unknown>) {
  return JSON.stringify(filters || {})
}

export function useInboxSummaryQuery() {
  return useQuery({
    queryKey: queryKeys.inbox.summary,
    queryFn: fetchInboxSummary,
    placeholderData: (previousData) => previousData,
  })
}

export function useInboxListQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.inbox.list(serializeFilters(filters)),
    queryFn: () => fetchInboxList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useInboxDetailQuery(deliveryId?: string | null) {
  return useQuery({
    queryKey: queryKeys.inbox.detail(deliveryId || ''),
    queryFn: () => fetchInboxDetail(deliveryId!),
    enabled: Boolean(deliveryId),
    placeholderData: (previousData) => previousData,
  })
}

export function useMarkInboxReadMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: markInboxRead,
    onSuccess: async (_result, deliveryId) => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.inbox.summary }),
        client.invalidateQueries({ queryKey: ['inbox', 'list'] }),
        client.invalidateQueries({ queryKey: queryKeys.inbox.detail(deliveryId) }),
      ])
    },
  })
}

export function useMarkInboxReadAllMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: markInboxReadAll,
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.inbox.summary }),
        client.invalidateQueries({ queryKey: ['inbox', 'list'] }),
      ])
    },
  })
}

export function useHandleInboxTodoMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ deliveryId, payload }: { deliveryId: string; payload: Record<string, unknown> }) =>
      handleInboxTodo(deliveryId, payload),
    onSuccess: async (_result, variables) => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.inbox.summary }),
        client.invalidateQueries({ queryKey: ['inbox', 'list'] }),
        client.invalidateQueries({ queryKey: queryKeys.inbox.detail(variables.deliveryId) }),
      ])
    },
  })
}
