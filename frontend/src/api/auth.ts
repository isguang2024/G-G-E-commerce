import { v5Client } from '@/api/v5/client'

/**
 * 登录 — 走 v5 OpenAPI client。后端 ogen handler 直接返回裸 schema
 * （没有 {code,data,message} 信封），所以这里手动把响应映射回前端
 * 既有的 Api.Auth.LoginResponse 形状，避免一次性改动整个登录流程。
 */
export async function fetchLogin(params: Api.Auth.LoginParams) {
  const { data, error } = await v5Client.POST('/auth/login', {
    body: { username: params.username, password: params.password }
  })
  if (error || !data) {
    throw error || new Error('login failed')
  }
  return data as unknown as Api.Auth.LoginResponse
}

/**
 * 刷新 Token — v5 OpenAPI client。
 */
export async function fetchRefreshToken(refreshToken: string) {
  const { data, error } = await v5Client.POST('/auth/refresh', {
    body: { refresh_token: refreshToken }
  })
  if (error || !data) {
    throw error || new Error('refresh failed')
  }
  return data as unknown as Api.Auth.LoginResponse
}

/**
 * 获取当前登录账户信息 — v5 OpenAPI /auth/me。
 * 后端 handler 返回 gen.AuthMe（enriched：actions/roles/current_*），这里
 * 扁平化成前端既有的 Api.Auth.UserInfo 形状，避免连带改 store / guard。
 */
export async function fetchGetUserInfo(): Promise<Api.Auth.UserInfo> {
  const { data, error } = await v5Client.GET('/auth/me')
  if (error || !data) {
    throw error || new Error('get auth.me failed')
  }
  return {
    id: data.id,
    email: data.email ?? '',
    username: data.username,
    nickname: data.nickname ?? '',
    avatar_url: data.avatar_url ?? data.avatar ?? undefined,
    phone: data.phone ?? undefined,
    status: data.status ?? '',
    is_super_admin: data.is_super_admin,
    current_collaboration_workspace_id:
      data.current_collaboration_workspace_id ?? undefined,
    collaboration_workspace_id: data.collaboration_workspace_id ?? undefined,
    current_auth_workspace_id: data.current_auth_workspace_id ?? undefined,
    current_auth_workspace_type: data.current_auth_workspace_type ?? undefined,
    actions: data.actions ?? [],
    created_at: data.created_at ?? '',
    roles: (data.roles ?? []).map((r) => r.code)
  }
}
