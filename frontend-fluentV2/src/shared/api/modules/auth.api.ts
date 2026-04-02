import { requestData } from '@/shared/api/client'
import type { AuthSession, CurrentUser, CurrentUserRole } from '@/shared/types/auth'

interface BackendUserRole {
  id?: string
  code?: string
  name?: string
  description?: string
}

interface BackendLoginUser {
  id?: string
  username?: string
  nickname?: string
  email?: string
  avatar_url?: string
  phone?: string
  status?: string
  is_super_admin?: boolean
  roles?: BackendUserRole[]
}

interface BackendLoginResponse {
  access_token?: string
  refresh_token?: string
  expires_in?: number
  user?: BackendLoginUser
}

interface BackendCurrentUser extends BackendLoginUser {
  current_tenant_id?: string
  actions?: string[]
  created_at?: string
  updated_at?: string
}

function normalizeUserRole(input: BackendUserRole): CurrentUserRole {
  return {
    id: `${input?.id || ''}`.trim(),
    code: `${input?.code || ''}`.trim(),
    name: `${input?.name || ''}`.trim(),
    description: `${input?.description || ''}`.trim() || undefined,
  }
}

export function normalizeCurrentUser(input: BackendCurrentUser): CurrentUser {
  const username = `${input?.username || ''}`.trim()
  const nickname = `${input?.nickname || ''}`.trim()
  const roles = Array.isArray(input?.roles) ? input.roles.map(normalizeUserRole) : []
  const actions = Array.isArray(input?.actions)
    ? input.actions.map((item) => `${item || ''}`.trim()).filter(Boolean)
    : []

  return {
    id: `${input?.id || ''}`.trim(),
    username,
    displayName: nickname || username || `${input?.email || ''}`.trim() || '未命名用户',
    email: `${input?.email || ''}`.trim(),
    avatarUrl: `${input?.avatar_url || ''}`.trim(),
    phone: `${input?.phone || ''}`.trim(),
    status: `${input?.status || 'active'}`.trim(),
    isSuperAdmin: Boolean(input?.is_super_admin),
    currentTenantId: `${input?.current_tenant_id || ''}`.trim() || null,
    roles,
    actions,
    badges: [
      ...(Boolean(input?.is_super_admin) ? ['超级管理员'] : []),
      ...roles.map((item) => item.name).filter(Boolean),
    ],
  }
}

function normalizeSession(input: BackendLoginResponse): AuthSession {
  return {
    accessToken: `${input?.access_token || ''}`.trim(),
    refreshToken: `${input?.refresh_token || ''}`.trim(),
    expiresIn: Number(input?.expires_in || 0),
    issuedAt: Date.now(),
  }
}

export async function loginWithPassword(payload: { username: string; password: string }) {
  const result = await requestData<BackendLoginResponse>({
    method: 'POST',
    url: '/api/v1/auth/login',
    data: payload,
  })

  return {
    session: normalizeSession(result),
    loginUser: normalizeCurrentUser((result.user || {}) as BackendCurrentUser),
  }
}

export async function registerWithPassword(payload: {
  username: string
  password: string
  email?: string
  nickname?: string
}) {
  const result = await requestData<BackendLoginResponse>({
    method: 'POST',
    url: '/api/v1/auth/register',
    data: payload,
  })

  return {
    session: normalizeSession(result),
    loginUser: normalizeCurrentUser((result.user || {}) as BackendCurrentUser),
  }
}

export async function fetchCurrentUser() {
  const result = await requestData<BackendCurrentUser>({
    method: 'GET',
    url: '/api/v1/user/info',
  })

  return normalizeCurrentUser(result)
}

export async function refreshAuthSession(refreshToken: string) {
  const result = await requestData<BackendLoginResponse>({
    method: 'POST',
    url: '/api/v1/auth/refresh',
    data: {
      refresh_token: refreshToken,
    },
  })

  return normalizeSession(result)
}
