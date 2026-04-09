import { v5Client, SKIP_WORKSPACE_CONTEXT_HEADER } from '@/api/v5/client'
import { unwrap, createV5HttpError } from '@/api/system-manage/_shared'

/**
 * 登录 — 走 v5 OpenAPI client。后端 ogen handler 直接返回裸 schema
 * （没有 {code,data,message} 信封），所以这里手动把响应映射回前端
 * 既有的 Api.Auth.LoginResponse 形状，避免一次性改动整个登录流程。
 */
export async function fetchLogin(params: Api.Auth.LoginParams) {
  const { data, error, response } = await v5Client.POST('/auth/login', {
    body: { username: params.username, password: params.password },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw (error ? createV5HttpError(error, response) : new Error('login failed'))
  }
  return data as Api.Auth.LoginResponse
}

/**
 * 获取注册上下文：按当前 host+path 命中注册入口 + 合并策略，返回前端
 * 渲染注册页所需的开关（是否允许公开注册、是否需要邀请码 / 人机验证等）
 * 以及注册成功后的 landing 目标。
 */
export async function fetchRegisterContext(host?: string, path?: string) {
  const { data, error, response } = await v5Client.GET('/auth/register-context', {
    params: { query: { host, path } },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw (error ? createV5HttpError(error, response) : new Error('fetch register context failed'))
  }
  return data
}

/**
 * 注册账号 — 走 v5 OpenAPI /auth/register。返回 LoginResponse + landing。
 */
export async function fetchRegister(body: {
  username: string
  password: string
  confirm_password?: string
  email?: string
  nickname?: string
  captcha_token?: string
  invitation_code?: string
  agreement_version?: string
}) {
  const { data, error, response } = await v5Client.POST('/auth/register', {
    body,
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw (error ? createV5HttpError(error, response) : new Error('register failed'))
  }
  return data
}

/**
 * 刷新 Token — v5 OpenAPI client。
 */
export async function fetchRefreshToken(refreshToken: string) {
  const { data, error, response } = await v5Client.POST('/auth/refresh', {
    body: { refresh_token: refreshToken },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw (error ? createV5HttpError(error, response) : new Error('refresh failed'))
  }
  return data as Api.Auth.LoginResponse
}

/**
 * 获取当前登录账户信息 — v5 OpenAPI /auth/me。
 * 后端 handler 返回 gen.AuthMe（enriched：actions/roles/current_*），这里
 * 扁平化成前端既有的 Api.Auth.UserInfo 形状，避免连带改 store / guard。
 */
export async function fetchGetUserInfo(): Promise<Api.Auth.UserInfo> {
  const data = await unwrap(
    v5Client.GET('/auth/me', {
      headers: {
        [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
      }
    })
  )
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
