import { computed, reactive, ref, shallowReactive } from 'vue'
import { defineStore } from 'pinia'
import {
  fetchGetInboxSummary,
  fetchGetInboxList,
  fetchGetInboxDetail,
  fetchMarkInboxRead,
  fetchMarkInboxReadAll,
  fetchHandleInboxTodo
} from '@/api/message'

type BoxType = Api.Message.BoxType
type InboxItem = Api.Message.InboxItem
type InboxSummary = Api.Message.InboxSummary

const boxTypes: BoxType[] = ['notice', 'message', 'todo']

function createSummary(): InboxSummary {
  return {
    unread_total: 0,
    notice_count: 0,
    message_count: 0,
    todo_count: 0
  }
}

export const useMessageStore = defineStore('messageStore', () => {
  const summary = ref<InboxSummary>(createSummary())
  const summaryLoaded = ref(false)
  const summaryLoading = ref(false)

  // shallowReactive：仅顶层 key 响应，避免对每条 InboxItem 做深度代理
  const previewMap = shallowReactive<Record<BoxType, InboxItem[]>>({
    notice: [],
    message: [],
    todo: []
  })
  const previewLoading = reactive<Record<BoxType, boolean>>({
    notice: false,
    message: false,
    todo: false
  })
  const previewLoaded = reactive<Record<BoxType, boolean>>({
    notice: false,
    message: false,
    todo: false
  })

  const hasUnread = computed(() => summary.value.unread_total > 0)

  const loadSummary = async (force = false, options?: { signal?: AbortSignal }) => {
    if (summaryLoading.value) return summary.value
    if (summaryLoaded.value && !force) return summary.value
    summaryLoading.value = true
    try {
      summary.value = await fetchGetInboxSummary({ signal: options?.signal })
      summaryLoaded.value = true
      return summary.value
    } finally {
      summaryLoading.value = false
    }
  }

  const loadPreview = async (boxType: BoxType, force = false) => {
    if (previewLoading[boxType]) return previewMap[boxType]
    if (previewLoaded[boxType] && !force) return previewMap[boxType]
    previewLoading[boxType] = true
    try {
      const response = await fetchGetInboxList({
        box_type: boxType,
        current: 1,
        size: 6
      })
      previewMap[boxType] = response.records || []
      previewLoaded[boxType] = true
      return previewMap[boxType]
    } finally {
      previewLoading[boxType] = false
    }
  }

  const loadPanelData = async (force = false) => {
    await Promise.all([
      loadSummary(force),
      ...boxTypes.map((boxType) => loadPreview(boxType, force))
    ])
  }

  const markRead = async (deliveryId: string) => {
    await fetchMarkInboxRead(deliveryId)
    await loadPanelData(true)
  }

  const markReadAll = async (boxType?: BoxType | '') => {
    await fetchMarkInboxReadAll(boxType)
    await loadPanelData(true)
  }

  const handleTodo = async (deliveryId: string, action: 'done' | 'ignored') => {
    await fetchHandleInboxTodo(deliveryId, { action })
    await loadPanelData(true)
  }

  const getDetail = async (deliveryId: string) => {
    return fetchGetInboxDetail(deliveryId)
  }

  const resetState = () => {
    summary.value = createSummary()
    summaryLoaded.value = false
    boxTypes.forEach((boxType) => {
      previewMap[boxType] = []
      previewLoading[boxType] = false
      previewLoaded[boxType] = false
    })
  }

  return {
    summary,
    summaryLoaded,
    summaryLoading,
    previewMap,
    previewLoading,
    previewLoaded,
    hasUnread,
    loadSummary,
    loadPreview,
    loadPanelData,
    markRead,
    markReadAll,
    handleTodo,
    getDetail,
    resetState
  }
})
