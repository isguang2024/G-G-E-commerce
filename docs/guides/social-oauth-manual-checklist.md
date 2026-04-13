# Social OAuth 手动验证清单

## 前置

- 后端已重启到最新代码。
- 前端 `5174` 已启动。
- GitHub OAuth App 回调地址配置正确。
- provider `github` 已启用。

## 用例 1：已绑定账号直接登录

1. 打开 `/account/auth/login`，点击 GitHub。
2. 完成 GitHub 授权后应进入 `/account/auth/social-callback`。
3. 页面自动登录并跳转目标页。

## 用例 2：未绑定账号走注册绑定

1. 使用未绑定 GitHub 账号授权。
2. 进入 `social-callback` 后应跳转 `/account/auth/register?social_token=...`。
3. 注册页显示社交来源提示，邮箱/用户名可预填。
4. 提交注册成功后，用户可登录且绑定关系写入 `user_social_accounts`。

## 用例 3：冲突账号

1. GitHub 邮箱与系统已有账号邮箱一致但未绑定。
2. 回调后应进入注册页并显示冲突提示，不应自动关联。

## 用例 4：能力禁用

1. 关闭公开注册或禁用 provider。
2. 登录/注册页社交入口应出现禁用提示（非可点击失败）。

## 错误检查

- 浏览器 Console 不应出现前端异常。
- `/api/v1/auth/social/exchange` 不应 404（若 404，优先确认后端是否已重启）。
