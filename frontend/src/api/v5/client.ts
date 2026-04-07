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
import type { paths } from './schema'

export const v5Client = createClient<paths>({
  baseUrl: '/api/v1'
})

// 注入 Authorization 头：openapi-fetch 的 middleware 钩子在每次请求前
// 从 user store 拿 access token，与原 axios 拦截器行为对齐。
v5Client.use({
  onRequest({ request }) {
    const { accessToken } = useUserStore()
    if (accessToken) {
      const token = accessToken.startsWith('Bearer ') ? accessToken : `Bearer ${accessToken}`
      request.headers.set('Authorization', token)
    }
    return request
  }
})
