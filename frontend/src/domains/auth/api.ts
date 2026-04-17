import { v5Client, SKIP_WORKSPACE_CONTEXT_HEADER } from '@/api/v5/client'
import { unwrap, createV5HttpError } from '@/domains/governance/api/_shared'
import { AUTH_PROTOCOL_VERSION } from '@/domains/auth/centralized-login'
import {
  getCurrentRuntimeAppKey,
  getEffectiveManagedAppKey,
  getHttpAccessToken
} from '@/utils/http/request-context'

// @compat-status: transition 认证主链已切 v5 接口，但返回值仍映射为旧 Api.Auth.* 形状。

const legacyCurrentWorkspaceIdKey = ['current', 'collaboration', 'workspace', 'id'].join('_')
const legacyWorkspaceRecordKey = ['collaboration', 'workspace', 'id'].join('_')

export async function fetchLogin(params: Api.Auth.LoginParams) {
  const { data, error, response } = await v5Client.POST('/auth/login', {
    body: {
      username: params.username,
      password: params.password,
      target_app_key: params.target_app_key,
      redirect_uri: params.redirect_uri,
      target_path: params.target_path,
      navigation_space_key: params.navigation_space_key,
      state: params.state,
      nonce: params.nonce,
      auth_protocol_version: params.auth_protocol_version || AUTH_PROTOCOL_VERSION
    },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw error ? createV5HttpError(error, response) : new Error('login failed')
  }
  return data as Api.Auth.LoginResponse
}

export async function fetchRegisterContext(host?: string, path?: string) {
  const { data, error, response } = await v5Client.GET('/auth/register-context', {
    params: { query: { host, path } },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw error ? createV5HttpError(error, response) : new Error('fetch register context failed')
  }
  return data
}

export async function fetchLoginPageContext(params?: {
  host?: string
  path?: string
  target_app_key?: string
  login_page_key?: string
  page_scene?: 'login' | 'register' | 'forget_password'
}) {
  const { data, error, response } = await v5Client.GET('/auth/login-page-context', {
    params: {
      query: {
        host: params?.host,
        path: params?.path,
        target_app_key: params?.target_app_key,
        login_page_key: params?.login_page_key,
        page_scene: params?.page_scene
      }
    },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw error ? createV5HttpError(error, response) : new Error('fetch login page context failed')
  }
  return data
}

export async function fetchRegister(body: {
  username: string
  password: string
  confirm_password?: string
  email?: string
  nickname?: string
  captcha_token?: string
  invitation_code?: string
  agreement_version?: string
  social_token?: string
  source_app_key?: string
  source_navigation_space_key?: string
  source_home_path?: string
}) {
  const { data, error, response } = await v5Client.POST('/auth/register', {
    body,
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw error ? createV5HttpError(error, response) : new Error('register failed')
  }
  return data
}

export async function fetchSocialTokenExchange(socialToken: string) {
  const { data, error, response } = await v5Client.POST('/auth/social/exchange', {
    body: { social_token: socialToken },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw error ? createV5HttpError(error, response) : new Error('social token exchange failed')
  }
  return data
}

export async function fetchRefreshToken(refreshToken: string) {
  const { data, error, response } = await v5Client.POST('/auth/refresh', {
    body: {
      refresh_token: refreshToken,
      client_app_key: getEffectiveManagedAppKey() || getCurrentRuntimeAppKey(),
      auth_protocol_version: AUTH_PROTOCOL_VERSION
    },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw error ? createV5HttpError(error, response) : new Error('refresh failed')
  }
  return data as Api.Auth.LoginResponse
}

export async function fetchExchangeAuthCallback(body: {
  code: string
  state: string
  nonce: string
  target_app_key: string
  redirect_uri: string
}) {
  const { data, error, response } = await v5Client.POST('/auth/callback/exchange', {
    body: {
      ...body,
      auth_protocol_version: AUTH_PROTOCOL_VERSION
    },
    headers: {
      [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
    }
  })
  if (error || !data) {
    throw error ? createV5HttpError(error, response) : new Error('exchange auth callback failed')
  }
  return data as Api.Auth.LoginResponse
}

export async function fetchLogout() {
  const accessToken = getHttpAccessToken()
  const headers: Record<string, string> = {
    [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
  }
  if (accessToken) {
    headers.Authorization = accessToken.startsWith('Bearer ')
      ? accessToken
      : `Bearer ${accessToken}`
  }
  const { error, response } = await v5Client.POST('/auth/logout', {
    headers
  })
  if (error) {
    throw createV5HttpError(error, response)
  }
}

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
    [legacyCurrentWorkspaceIdKey]: data.current_workspace_id ?? data.workspace_id ?? undefined,
    [legacyWorkspaceRecordKey]: data.workspace_id ?? undefined,
    current_auth_workspace_id: data.current_auth_workspace_id ?? undefined,
    current_auth_workspace_type: data.current_auth_workspace_type ?? undefined,
    actions: data.actions ?? [],
    created_at: data.created_at ?? '',
    roles: (data.roles ?? []).map((r) => r.code)
  }
}
