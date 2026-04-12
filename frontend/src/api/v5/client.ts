/**
 * v5 OpenAPI-first client.
 *
 * 这是 GGE 5.0 重构后的唯一前端 API 入口骨架。所有走 ogen 生成的后端
 * 端点（即 backend/api/openapi/openapi.yaml 中声明的 operation）从这里
 * 调用，类型由 schema.d.ts 自动派生，不再手写 axios 接口类型。
 *
 * 旧的 src/api/*.ts 在 Phase 5 的多个 PR 中被逐一替换。本文件先落骨架 +
 * 第一处真实接口替换（listMyWorkspaces），后续按域增加。
 */
import createClient from 'openapi-fetch'
import { useUserStore } from '@/store/modules/user'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { useCollaborationWorkspaceStore } from '@/store/modules/collaboration-workspace'
import { useAppContextStore } from '@/store/modules/app-context'
import { AUTH_PROTOCOL_VERSION } from '@/utils/auth/centralized-login'
import type { paths } from './schema'

export const v5Client = createClient<paths>({
  baseUrl: '/api/v1'
})

const SKIP_WORKSPACE_CONTEXT_HEADER = 'X-Skip-Workspace-Context'
const AUTH_RETRY_HEADER = 'X-Auth-Retry'
let refreshPromise: Promise<boolean> | null = null

function normalizeBackendBaseUrl(value?: string): string {
  const raw = `${value || ''}`.trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw)) {
    return raw.replace(/\/+$/, '')
  }
  const path = raw.startsWith('/') ? raw : `/${raw}`
  return path.replace(/\/+$/, '')
}

function rewriteRequestWithDynamicBase(request: Request, dynamicBaseUrl: string): Request {
  const current = new URL(request.url, window.location.origin)
  const apiPath = `${current.pathname}${current.search}`
  const fallbackPrefix = '/api/v1'

  if (/^https?:\/\//i.test(dynamicBaseUrl)) {
    const base = new URL(dynamicBaseUrl)
    const basePath = base.pathname.replace(/\/+$/, '')
    let suffix = apiPath
    if (basePath && suffix.startsWith(basePath)) {
      suffix = suffix.slice(basePath.length)
      if (!suffix.startsWith('/')) suffix = `/${suffix}`
    } else if (basePath === '' && suffix.startsWith(fallbackPrefix)) {
      suffix = suffix
    } else if (basePath !== '' && suffix.startsWith(fallbackPrefix) && !basePath.endsWith(fallbackPrefix)) {
      suffix = `${fallbackPrefix}${suffix.slice(fallbackPrefix.length)}`
    }
    const targetPath = `${basePath}${suffix}`.replace(/\/{2,}/g, '/')
    const nextURL = `${base.origin}${targetPath}${current.hash}`
    return new Request(nextURL, request)
  }

  const basePath = dynamicBaseUrl
  let suffix = apiPath
  if (basePath && suffix.startsWith(basePath)) {
    suffix = suffix.slice(basePath.length)
    if (!suffix.startsWith('/')) suffix = `/${suffix}`
  }
  const targetPath = `${basePath}${suffix}`.replace(/\/{2,}/g, '/')
  return new Request(`${targetPath}${current.hash}`, request)
}

function shouldUseSharedSessionMode(): boolean {
  const appContextStore = useAppContextStore()
  const appKey = appContextStore.effectiveManagedAppKey || appContextStore.currentRuntimeAppKey
  const authMode = `${appContextStore.resolveAppAuthMode(appKey) || ''}`.trim()
  if (authMode === 'shared_cookie') {
    return true
  }
  return appContextStore.isFeatureEnabledForApp(appKey, 'shared_cookie')
}

function shouldSkipRefreshAttempt(request: Request): boolean {
  const pathname = new URL(request.url, window.location.origin).pathname
  if (request.headers.get(AUTH_RETRY_HEADER) === '1') {
    return true
  }
  return (
    pathname.endsWith('/auth/login') ||
    pathname.endsWith('/auth/refresh') ||
    pathname.endsWith('/auth/logout') ||
    pathname.endsWith('/auth/callback/exchange')
  )
}

async function refreshSessionIfNeeded(): Promise<boolean> {
  if (refreshPromise) {
    return refreshPromise
  }
  refreshPromise = (async () => {
    const userStore = useUserStore()
    const appContextStore = useAppContextStore()
    const rawRefreshToken = `${userStore.refreshToken || ''}`.trim()
    if (!rawRefreshToken) {
      return false
    }
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
      },
      body: JSON.stringify({
        refresh_token: rawRefreshToken,
        client_app_key:
          appContextStore.effectiveManagedAppKey || appContextStore.currentRuntimeAppKey || '',
        auth_protocol_version: AUTH_PROTOCOL_VERSION
      })
    }).catch(() => null)
    if (!response?.ok) {
      return false
    }
    const payload = await response.json().catch(() => null)
    const accessToken = `${payload?.access_token || ''}`.trim()
    const refreshToken = `${payload?.refresh_token || ''}`.trim()
    if (!accessToken || !refreshToken) {
      return false
    }
    userStore.applySession({
      accessToken,
      refreshToken,
      isLogin: true
    })
    return true
  })().finally(() => {
    refreshPromise = null
  })
  return refreshPromise
}

// 注入 Authorization + 工作空间头：与原 axios 拦截器行为对齐。
// X-Auth-Workspace-Id: 当前鉴权工作空间（个人 / 协作）
// X-Collaboration-Workspace-Id: 仅在协作模式下注入
v5Client.use({
  onRequest({ request }) {
    let nextRequest = request
    const appContextStore = useAppContextStore()
    const dynamicBackendBaseUrl = normalizeBackendBaseUrl(
      appContextStore.currentRuntimeBackendEntryURL
    )
    if (dynamicBackendBaseUrl && typeof window !== 'undefined') {
      nextRequest = rewriteRequestWithDynamicBase(request, dynamicBackendBaseUrl)
    }

    const shouldSkipWorkspaceContext = request.headers.get(SKIP_WORKSPACE_CONTEXT_HEADER) === 'true'
    if (shouldSkipWorkspaceContext) {
      nextRequest.headers.delete(SKIP_WORKSPACE_CONTEXT_HEADER)
    }

    const { accessToken } = useUserStore()
    if (accessToken) {
      const token = accessToken.startsWith('Bearer ') ? accessToken : `Bearer ${accessToken}`
      nextRequest.headers.set('Authorization', token)
    }

    const { currentAuthWorkspaceId } = useWorkspaceStore()
    if (!shouldSkipWorkspaceContext && currentAuthWorkspaceId) {
      nextRequest.headers.set('X-Auth-Workspace-Id', currentAuthWorkspaceId)
    }

    const { currentCollaborationWorkspaceId, currentContextMode } =
      useCollaborationWorkspaceStore()
    if (
      !shouldSkipWorkspaceContext &&
      currentContextMode === 'collaboration' &&
      currentCollaborationWorkspaceId
    ) {
      nextRequest.headers.set('X-Collaboration-Workspace-Id', currentCollaborationWorkspaceId)
    }

    return nextRequest
  },
  async onResponse({ request, response }) {
    if (response.status !== 401 || typeof window === 'undefined') {
      return response
    }
    if (!shouldUseSharedSessionMode() || shouldSkipRefreshAttempt(request)) {
      return response
    }
    const refreshed = await refreshSessionIfNeeded()
    if (!refreshed) {
      return response
    }
    const { accessToken } = useUserStore()
    if (!accessToken) {
      return response
    }
    if (request.bodyUsed) {
      return response
    }
    const headers = new Headers(request.headers)
    headers.set('Authorization', accessToken.startsWith('Bearer ') ? accessToken : `Bearer ${accessToken}`)
    headers.set(AUTH_RETRY_HEADER, '1')
    return fetch(new Request(request, { headers }))
  }
})

export { SKIP_WORKSPACE_CONTEXT_HEADER }
