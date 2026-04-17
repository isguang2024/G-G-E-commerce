# GitHub OAuth 社交登录配置

## 必填环境变量

后端读取顺序：`social_auth_providers` 表配置优先，其次环境变量回退。

- `GG_SOCIAL_GITHUB_CLIENT_ID`
- `GG_SOCIAL_GITHUB_CLIENT_SECRET`

## 回调地址

在 GitHub OAuth App 中配置回调地址：

- `http://<你的后端域名>/api/v1/auth/oauth/github/callback`

本地开发示例：

- `http://127.0.0.1:8080/api/v1/auth/oauth/github/callback`

## 启用步骤

1. 迁移并执行种子，确保存在 `github` provider。
2. 在系统配置中将 provider `enabled` 设为 `true`。
3. 填写 client_id / client_secret（DB 或环境变量）。
4. 登录页模板开启 `features.socialLogin` 并配置 `social.items` 中的 GitHub 入口。

## 路由说明

- 发起授权：`/api/v1/auth/oauth/github/authorize`
- OAuth 回调：`/api/v1/auth/oauth/github/callback`
- token 兑换：`/api/v1/auth/social/exchange`
- 前端中转页：`/account/auth/social-callback`

## 安全说明

- OAuth state 持久化并一次性消费，防重放。
- `social_token` 为短期 JWT，仅携带必要身份上下文，不存储 provider access_token。
