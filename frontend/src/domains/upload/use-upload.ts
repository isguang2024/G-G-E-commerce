import { ref } from 'vue'
import {
  deleteMedia,
  listMedia,
  uploadMediaWithPrepare,
  type MediaItem,
  type MediaUploadTarget,
  type MediaUploadResponse
} from './api'

export function useUpload() {
  const uploading = ref(false)
  const loading = ref(false)
  const error = ref('')

  async function submit(file: File, options: MediaUploadTarget = {}): Promise<MediaUploadResponse> {
    uploading.value = true
    error.value = ''
    try {
      return await uploadMediaWithPrepare(file, options)
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
    error,
    submit,
    fetchList,
    remove
  }
}
