import request from '@/utils/http'
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
 * 获取用户信息
 * @returns 用户信息
 */
export function fetchGetUserInfo() {
  return request.get<Api.Auth.UserInfo>({
    url: '/api/v1/user/info'
    // 自定义请求头
    // headers: {
    //   'X-Custom-Header': 'your-custom-value'
    // }
  })
}
