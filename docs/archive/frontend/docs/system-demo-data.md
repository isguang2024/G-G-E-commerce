# 系统最小演示数据

> 更新时间：2026-03-31。这里只保留系统回归需要的固定账号、空间和消息数据。

## 初始化命令

- 执行位置：`backend/`
- 初始化命令：`go run ./cmd/init-demo`
- 默认密码：`Demo123456`
- 默认运行模式：`single`
- 常用参数：`-password`、`-team-name`、`-space-key`、`-allow-production`

## 固定账号

- `platform_admin_demo`：平台管理员
- `collaboration_workspace_admin_demo`：团队管理员
- `member_demo`：普通成员

## 固定团队与菜单空间

- 默认团队：`演示团队`
- 非默认空间：`ops / 运营空间`
- 默认空间：`default`
- 初始化方式：每次执行 `init-demo` 都会从 `default` 复制菜单、页面与功能包菜单关联到 `ops`

## 消息演示数据

- 平台模板：`demo.wrapup.platform.notice`
- 团队模板：`demo.wrapup.collaboration_workspace.notice`
- 平台接收组：`演示平台接收组`
- 团队接收组：`演示团队接收组`
- 平台消息：`demo.system_wrapup.platform`
- 团队消息：`demo.system_wrapup.team`

## 最小联调顺序

1. `platform_admin_demo`：菜单、页面、空间、消息中心
2. `collaboration_workspace_admin_demo`：团队消息
3. `member_demo`：菜单裁剪、平台收件、团队入口拦截

## 使用约束

- 只用于非生产环境。
- 执行后会重置固定账号密码并刷新权限快照。
- 执行后会重建示例消息，避免重复积累。
