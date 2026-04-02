# 权限系统总体说明

> 基线日期：2026-03-31。本文只描述当前正式权限架构。

## 当前结论

- 功能包负责业务开通。
- 权限键负责动作与 API 原子能力。
- 菜单负责导航入口。
- 受管页面负责非菜单直达访问单元。
- 平台与团队通过快照进入运行时读链。
- 导航与页面访问统一由后端一次编译、前端一次消费。

## 主数据

- 用户、角色、团队、成员关系：`users`、`roles`、`tenants`、`tenant_members`
- 权限与功能包：`permission_keys`、`feature_packages`、`feature_package_bundles`、`feature_package_keys`、`feature_package_menus`
- 菜单、页面、API：`menus`、`menu_manage_groups`、`ui_pages`、`page_space_bindings`、`api_endpoints`、`api_endpoint_permission_bindings`
- 运行时快照：`platform_user_access_snapshots`、`platform_role_access_snapshots`、`team_access_snapshots`、`team_role_access_snapshots`

## 上下文模型

- 只保留 `platform` 与 `team` 两个正式上下文。
- 无 `X-Tenant-ID` 时走平台上下文。
- 有 `X-Tenant-ID` 时走团队上下文。

## 正式授权模型

- 功能包决定候选能力范围。
- 平台侧和团队侧分别按各自快照与减法规则收口。
- 本轮不重写功能包、权限键、API 鉴权底座和快照刷新机制。

## 统一访问编译层

- 正式入口：`GET /api/v1/runtime/navigation`
- 后端一次性编译当前空间解析结果、菜单树、菜单入口路由、受管页面路由和动作权限摘要。
- 前端只做轻量归一化、动态路由注册和缺失路由重拉。

## 判断标准

- 看模块是否开通：先看功能包。
- 看入口是否展示：看菜单与菜单减法结果。
- 看受管页面是否可访问：看后端编译后的页面访问结果。
- 看动作或 API 是否允许：看权限键与快照。
