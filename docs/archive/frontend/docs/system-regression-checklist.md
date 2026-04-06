# 系统回归验收清单

> 更新时间：2026-03-31。默认以单域单菜单模式回归。

## 前置条件

- 后端：`go test ./...`
- 前端：`pnpm exec vue-tsc --noEmit`
- 演示数据：`go run ./cmd/init-demo`
- 固定账号：`platform_admin_demo`、`collaboration_workspace_admin_demo`、`member_demo`
- 固定空间：`default`、`ops`

## 菜单管理

- 默认空间过滤正确，不串到 `ops`
- 菜单分组失败时只显示页内提示或空态，不弹业务式 warning
- 菜单挂接页面后，页面管理侧能同步看到关系

## 页面管理

- 页面列表、候选、父链按当前空间过滤
- 独立页、挂接页、分组页都能正常保存
- 页面“访问”入口与运行时路由一致

## 菜单空间 / Host

- 默认空间在未配置额外 Host 时可正常运行
- `ops` 可初始化、可重新初始化、可查看状态
- 默认首页与 Host 冲突校验有效

## 消息系统

- 平台消息发送、消息中心、发送记录、模板、发送人、接收组主链可用
- 正文中的站内链接在消息中心和发送记录详情里都可跳转
- 平台消息与团队消息入口分开，交互结构一致

## 角色场景

- 平台管理员可进入 `#/system/menu`、`#/system/page`、`#/system/menu-space`、`#/system/message`、`#/workspace/inbox`
- 团队管理员可进入 `#/team/message`
- 普通成员只能验证菜单裁剪和平台收件，不应进入团队发信页
