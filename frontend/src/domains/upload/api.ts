import { v5Client, unwrap } from '@/domains/governance/api/_shared'
import type { components } from '@/api/v5/schema'
import {
  getCurrentAuthWorkspaceId,
  getCurrentRuntimeBackendEntryURL,
  getHttpAccessToken
} from '@/utils/http/request-context'

export type MediaUploadResponse = components['schemas']['MediaUploadResponse']
export type MediaItem = components['schemas']['MediaItem']
export type MediaListResponse = components['schemas']['MediaListResponse']
export type MediaPrepareUploadRequest = components['schemas']['MediaPrepareUploadRequest']
export type MediaPrepareUploadResponse = components['schemas']['MediaPrepareUploadResponse']
export type MediaCompleteUploadRequest = components['schemas']['MediaCompleteUploadRequest']
export type MediaVisibleUploadKey = components['schemas']['MediaVisibleUploadKey']
export type MediaVisibleUploadRule = components['schemas']['MediaVisibleUploadRule']
export type MediaVisibleUploadKeyListResponse =
  components['schemas']['MediaVisibleUploadKeyListResponse']
// 上传模式开关：
//   - auto（默认）：调 /media/prepare，按服务端返回 mode 走 direct 或 relay；
//   - direct：调 /media/prepare，但若服务端给的是 relay 直接抛错（用于强制验证 STS 链路）；
//   - relay：跳过 prepare，直接走 /api/v1/media/upload 中转。
// 这条开关给前端自己用：开发联调、集成测试、或者临时禁用直传时不必改 UploadKey 配置。
export type MediaUploadMode = 'auto' | 'direct' | 'relay'

export type MediaUploadOptions = {
  key?: string
  rule?: string
  mode?: MediaUploadMode
  metadata?: Record<string, unknown>
  visibleKeys?: MediaVisibleUploadKey[]
  signal?: AbortSignal
  onProgress?: (percent: number, event: ProgressEvent<EventTarget>) => void
  onFallback?: (prepare: MediaPrepareUploadResponse) => void
  onResolved?: (plan: MediaUploadExecutionPlan) => void
  onCompleted?: (result: MediaUploadExecutionResult) => void
}
export type MediaUploadTarget = string | MediaUploadOptions | undefined
export type MediaResolvedUploadMode = 'direct' | 'relay'
export type MediaResolvedVisibleUploadTarget = {
  uploadKey?: MediaVisibleUploadKey
  rule?: MediaVisibleUploadRule
}
export type MediaUploadExecutionPlan = {
  requestedMode: MediaUploadMode
  resolvedMode: MediaResolvedUploadMode
  uploadKey?: string
  ruleKey?: string
  relayUrl?: string
  fallbackUsed: boolean
  prepare?: MediaPrepareUploadResponse
  visibleTarget?: MediaResolvedVisibleUploadTarget
}
export type MediaUploadExecutionResult = {
  media: MediaUploadResponse
  plan: MediaUploadExecutionPlan
}

function normalizeBackendBaseUrl(value?: string): string {
  const raw = `${value || ''}`.trim()
  if (!raw) return ''
  return raw.replace(/\/+$/, '')
}

function resolveUploadURL(path: string): string {
  const dynamicBase = normalizeBackendBaseUrl(getCurrentRuntimeBackendEntryURL())
  if (!dynamicBase) {
    return path
  }
  if (/^https?:\/\//i.test(dynamicBase)) {
    return `${dynamicBase}${path.startsWith('/') ? path : `/${path}`}`
  }
  const prefix = dynamicBase.startsWith('/') ? dynamicBase : `/${dynamicBase}`
  return `${prefix}${path.startsWith('/') ? path : `/${path}`}`
}

function buildUploadHeaders(): Headers {
  const headers = new Headers()
  const accessToken = getHttpAccessToken()
  if (accessToken) {
    headers.set(
      'Authorization',
      accessToken.startsWith('Bearer ') ? accessToken : `Bearer ${accessToken}`
    )
  }

  const authWorkspaceID = getCurrentAuthWorkspaceId()
  if (authWorkspaceID) {
    headers.set('X-Auth-Workspace-Id', authWorkspaceID)
  }

  return headers
}

export async function uploadMedia(file: File): Promise<MediaUploadResponse> {
  return uploadMediaWithPrepare(file)
}

function normalizeUploadOptions(target: MediaUploadTarget): MediaUploadOptions {
  if (!target) {
    return {}
  }
  if (typeof target === 'string') {
    const raw = target.trim()
    if (!raw) {
      return {}
    }
    const [key, ...rest] = raw.split('.')
    const rule = rest.join('.').trim()
    return {
      key: key.trim() || undefined,
      rule: rule || undefined
    }
  }
  return {
    ...target,
    key: `${target.key || ''}`.trim() || undefined,
    rule: `${target.rule || ''}`.trim() || undefined,
    mode: target.mode || undefined,
    visibleKeys: Array.isArray(target.visibleKeys) ? target.visibleKeys : undefined
  }
}

function pickVisibleUploadKey(
  records: MediaVisibleUploadKey[],
  requestedKey?: string
): MediaVisibleUploadKey | undefined {
  if (requestedKey) {
    return records.find((item) => item.key === requestedKey)
  }
  if (records.length === 1) {
    return records[0]
  }
  return records.find((item) => item.key === 'default')
}

function pickVisibleUploadRule(
  uploadKey: MediaVisibleUploadKey | undefined,
  requestedRule?: string
): MediaVisibleUploadRule | undefined {
  if (!uploadKey || !Array.isArray(uploadKey.rules) || uploadKey.rules.length === 0) {
    return undefined
  }
  if (requestedRule) {
    return uploadKey.rules.find((item) => item.ruleKey === requestedRule)
  }
  if (uploadKey.defaultRuleKey) {
    const matched = uploadKey.rules.find((item) => item.ruleKey === uploadKey.defaultRuleKey)
    if (matched) return matched
  }
  return uploadKey.rules.find((item) => item.isDefault) || uploadKey.rules[0]
}

export function resolveVisibleMediaUploadTarget(
  records: MediaVisibleUploadKey[],
  target: MediaUploadTarget = {}
): MediaResolvedVisibleUploadTarget {
  const options = normalizeUploadOptions(target)
  const uploadKey = pickVisibleUploadKey(records, options.key)
  return {
    uploadKey,
    rule: pickVisibleUploadRule(uploadKey, options.rule)
  }
}

export function describeMediaUploadPlan(plan: MediaUploadExecutionPlan): string {
  const modeLabel =
    plan.resolvedMode === 'direct'
      ? '前端直传'
      : plan.fallbackUsed
        ? '后端中转（已回退）'
        : '后端中转'
  const parts = [modeLabel]
  if (plan.uploadKey) {
    parts.push(`UploadKey: ${plan.uploadKey}`)
  }
  if (plan.ruleKey) {
    parts.push(`Rule: ${plan.ruleKey}`)
  }
  return parts.join(' / ')
}

function xhrUpload(
  url: string,
  init: {
    method: string
    headers?: HeadersInit
    body: XMLHttpRequestBodyInit
    responseType?: XMLHttpRequestResponseType
    signal?: AbortSignal
    onProgress?: (percent: number, event: ProgressEvent<EventTarget>) => void
  }
): Promise<{ status: number; responseText: string }> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open(init.method, url, true)
    Object.entries(new Headers(init.headers || {})).forEach(([key, value]) => {
      xhr.setRequestHeader(key, value)
    })
    if (init.responseType) {
      xhr.responseType = init.responseType
    }
    xhr.onload = () => resolve({ status: xhr.status, responseText: xhr.responseText || '' })
    xhr.onerror = () => reject(new Error('上传失败'))
    xhr.onabort = () => reject(new DOMException('上传已取消', 'AbortError'))
    if (xhr.upload && init.onProgress) {
      xhr.upload.onprogress = (event) => {
        if (!event.lengthComputable || event.total <= 0) {
          init.onProgress?.(0, event)
          return
        }
        init.onProgress?.(Math.round((event.loaded / event.total) * 100), event)
      }
    }
    if (init.signal) {
      if (init.signal.aborted) {
        xhr.abort()
        return
      }
      init.signal.addEventListener('abort', () => xhr.abort(), { once: true })
    }
    xhr.send(init.body)
  })
}

function emitResolvedPlan(options: MediaUploadOptions, plan: MediaUploadExecutionPlan) {
  options.onResolved?.(plan)
}

function emitCompletedResult(options: MediaUploadOptions, result: MediaUploadExecutionResult) {
  options.onCompleted?.(result)
}

async function uploadMediaByRelay(
  file: File,
  target: MediaUploadTarget = {},
  relayPath = '/api/v1/media/upload'
): Promise<MediaUploadResponse> {
  const options = normalizeUploadOptions(target)
  const formData = new FormData()
  formData.set('file', file)
  if (options.key) {
    formData.set('key', options.key)
  }
  if (options.rule) {
    formData.set('rule', options.rule)
  }

  const response = await xhrUpload(resolveUploadURL(relayPath), {
    method: 'POST',
    body: formData,
    headers: buildUploadHeaders(),
    signal: options.signal,
    onProgress: options.onProgress
  })
  const payload = JSON.parse(response.responseText || '{}')
  if (response.status < 200 || response.status >= 300) {
    const message = `${payload?.message || payload?.error_message || '上传失败'}`.trim()
    throw new Error(message || '上传失败')
  }

  return payload as MediaUploadResponse
}

export async function prepareMediaUpload(
  file: File,
  target: MediaUploadTarget = {}
): Promise<MediaPrepareUploadResponse> {
  const options = normalizeUploadOptions(target)
  return unwrap(
    v5Client.POST('/media/prepare', {
      body: {
        key: options.key,
        rule: options.rule,
        filename: file.name,
        mimeType: file.type || undefined,
        size: file.size,
        checksum: undefined
      }
    })
  )
}

export async function listVisibleMediaUploadKeys(): Promise<MediaVisibleUploadKeyListResponse> {
  return unwrap(v5Client.GET('/media/upload-keys'))
}

export async function fetchVisibleMediaUploadKeys(): Promise<MediaVisibleUploadKey[]> {
  const response = await listVisibleMediaUploadKeys()
  return Array.isArray(response.records) ? response.records : []
}

export async function completeMediaUpload(
  file: File,
  prepare: MediaPrepareUploadResponse,
  target: MediaUploadTarget = {}
): Promise<MediaUploadResponse> {
  const options = normalizeUploadOptions(target)
  return unwrap(
    v5Client.POST('/media/complete', {
      body: {
        key: options.key,
        rule: options.rule,
        filename: prepare.filename || file.name,
        storageKey: prepare.storageKey,
        mimeType: prepare.contentType || file.type || undefined,
        size: file.size,
        checksum: undefined,
        etag: undefined
      }
    })
  )
}

async function uploadMediaDirect(
  file: File,
  prepare: MediaPrepareUploadResponse,
  target: MediaUploadTarget
): Promise<MediaUploadResponse> {
  const options = normalizeUploadOptions(target)
  const method = `${prepare.method || 'POST'}`.trim().toUpperCase()
  const url = `${prepare.url || ''}`.trim()
  if (!url) {
    throw new Error('直传地址缺失')
  }

  const requestHeaders = new Headers()
  Object.entries(prepare.headers || {}).forEach(([key, value]) => {
    if (typeof value === 'string' && value.trim()) {
      requestHeaders.set(key, value)
    }
  })

  let body: BodyInit
  if (method === 'PUT') {
    if (!requestHeaders.has('Content-Type') && (prepare.contentType || file.type)) {
      requestHeaders.set('Content-Type', prepare.contentType || file.type)
    }
    body = file
  } else {
    const formData = new FormData()
    Object.entries(prepare.form || {}).forEach(([key, value]) => {
      if (typeof value === 'string') {
        formData.append(key, value)
      }
    })
    formData.append('file', file)
    body = formData
  }

  const response = await xhrUpload(url, {
    method,
    headers: requestHeaders,
    body: body as XMLHttpRequestBodyInit,
    signal: options.signal,
    onProgress: options.onProgress
  })
  if (response.status < 200 || response.status >= 300) {
    throw new Error(`直传失败: ${response.status}`)
  }

  return completeMediaUpload(file, prepare, options)
}

export async function uploadMediaWithPrepare(
  file: File,
  target: MediaUploadTarget = {}
): Promise<MediaUploadResponse> {
  const result = await uploadMediaWithPlan(file, target)
  return result.media
}

export async function uploadMediaWithPlan(
  file: File,
  target: MediaUploadTarget = {}
): Promise<MediaUploadExecutionResult> {
  const options = normalizeUploadOptions(target)
  const mode: MediaUploadMode = options.mode || 'auto'
  const visibleTarget = Array.isArray(options.visibleKeys)
    ? resolveVisibleMediaUploadTarget(options.visibleKeys, options)
    : undefined

  // 强制中转：跳过 prepare，少一次 RTT。常用于禁用直传或快速回退。
  if (mode === 'relay') {
    const plan: MediaUploadExecutionPlan = {
      requestedMode: mode,
      resolvedMode: 'relay',
      uploadKey: options.key,
      ruleKey: options.rule,
      relayUrl: '/api/v1/media/upload',
      fallbackUsed: false,
      visibleTarget
    }
    emitResolvedPlan(options, plan)
    const media = await uploadMediaByRelay(file, options)
    const result: MediaUploadExecutionResult = {
      media,
      plan
    }
    emitCompletedResult(options, result)
    return result
  }

  const prepare = await prepareMediaUpload(file, options)
  if (prepare.mode === 'direct') {
    const plan: MediaUploadExecutionPlan = {
      requestedMode: mode,
      resolvedMode: 'direct',
      uploadKey: prepare.uploadKey || options.key,
      ruleKey: prepare.ruleKey || options.rule,
      fallbackUsed: prepare.fallbackUsed === true,
      prepare,
      visibleTarget
    }
    emitResolvedPlan(options, plan)
    const media = await uploadMediaDirect(file, prepare, options)
    const result: MediaUploadExecutionResult = {
      media,
      plan
    }
    emitCompletedResult(options, result)
    return result
  }

  // 强制直传但服务端给的是 relay：抛错而非静默回退，让调用方知道直传不可用
  // （UploadKey 配置或 Driver 能力限制导致），便于排查而不是看似成功了一次中转。
  if (mode === 'direct') {
    throw new Error('已请求强制直传（mode=direct），但服务端返回 relay 模式')
  }

  options.onFallback?.(prepare)
  const relayUrl = prepare.relayUrl || '/api/v1/media/upload'
  const plan: MediaUploadExecutionPlan = {
    requestedMode: mode,
    resolvedMode: 'relay',
    uploadKey: prepare.uploadKey || options.key,
    ruleKey: prepare.ruleKey || options.rule,
    relayUrl,
    fallbackUsed: prepare.fallbackUsed === true,
    prepare,
    visibleTarget
  }
  emitResolvedPlan(options, plan)
  const media = await uploadMediaByRelay(file, options, relayUrl)
  const result: MediaUploadExecutionResult = {
    media,
    plan
  }
  emitCompletedResult(options, result)
  return result
}

export async function listMedia(): Promise<MediaListResponse> {
  return unwrap(v5Client.GET('/media'))
}

export async function deleteMedia(id: string): Promise<void> {
  const { error } = await v5Client.DELETE('/media/{id}', {
    params: { path: { id } }
  })
  if (error) {
    throw error
  }
}
