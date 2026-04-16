import { ref } from 'vue'
import {
  deleteMedia,
  fetchVisibleMediaUploadKeys,
  listMedia,
  type MediaUploadExecutionPlan,
  resolveVisibleMediaUploadTarget,
  uploadMediaWithPlan,
  uploadMediaWithPrepare,
  type MediaItem,
  type MediaResolvedVisibleUploadTarget,
  type MediaUploadExecutionResult,
  type MediaUploadTarget,
  type MediaVisibleUploadKey,
  type MediaUploadResponse
} from './api'

export function useUpload() {
  const uploading = ref(false)
  const loading = ref(false)
  const visibleKeysLoading = ref(false)
  const error = ref('')
  const visibleKeys = ref<MediaVisibleUploadKey[]>([])
  const lastPlan = ref<MediaUploadExecutionPlan | null>(null)
  const lastResult = ref<MediaUploadExecutionResult | null>(null)

  async function submit(file: File, options: MediaUploadTarget = {}): Promise<MediaUploadResponse> {
    uploading.value = true
    error.value = ''
    lastPlan.value = null
    lastResult.value = null
    try {
      return await uploadMediaWithPrepare(file, options)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '上传失败'
      throw err
    } finally {
      uploading.value = false
    }
  }

  async function submitDetailed(
    file: File,
    options: MediaUploadTarget = {}
  ): Promise<MediaUploadExecutionResult> {
    uploading.value = true
    error.value = ''
    lastPlan.value = null
    lastResult.value = null
    try {
      const result =
        typeof options === 'string'
          ? await uploadMediaWithPlan(file, options)
          : await uploadMediaWithPlan(file, {
              ...options,
              onResolved: (plan) => {
                lastPlan.value = plan
                options.onResolved?.(plan)
              },
              onCompleted: (completed) => {
                lastResult.value = completed
                options.onCompleted?.(completed)
              }
            })
      if (!lastPlan.value) {
        lastPlan.value = result.plan
      }
      lastResult.value = result
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : '上传失败'
      throw err
    } finally {
      uploading.value = false
    }
  }

  async function fetchList(): Promise<MediaItem[]> {
    loading.value = true
    error.value = ''
    try {
      const response = await listMedia()
      return Array.isArray(response.records) ? response.records : []
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载媒体列表失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchVisibleKeys(): Promise<MediaVisibleUploadKey[]> {
    visibleKeysLoading.value = true
    error.value = ''
    try {
      const records = await fetchVisibleMediaUploadKeys()
      visibleKeys.value = records
      return records
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载上传场景失败'
      throw err
    } finally {
      visibleKeysLoading.value = false
    }
  }

  function resolveVisibleTarget(target: MediaUploadTarget = {}): MediaResolvedVisibleUploadTarget {
    return resolveVisibleMediaUploadTarget(visibleKeys.value, target)
  }

  async function remove(id: string) {
    error.value = ''
    try {
      await deleteMedia(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除媒体失败'
      throw err
    }
  }

  return {
    uploading,
    loading,
    visibleKeysLoading,
    error,
    visibleKeys,
    lastPlan,
    lastResult,
    submit,
    submitDetailed,
    fetchList,
    fetchVisibleKeys,
    resolveVisibleTarget,
    remove
  }
}
