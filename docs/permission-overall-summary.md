# 权限系统总体说明

> 现状基线：2026-03-27。本文只描述当前正式权限架构，不再保留历史分阶段推进记录。

## 1. 当前结论

当前权限体系已经收口为：

- 功能包负责业务开通
- 权限键负责动作与 API 的原子能力标识
- 菜单负责导航入口
- 页面负责访问单元
- 平台与团队分别通过快照进入运行时读链

一句话总结：
- 先开功能包，再由菜单、页面、权限键展开运行时结果，最后由角色、用户和团队边界做减法裁剪。

## 2. 当前正式主数据

### 2.1 用户、角色、团队

- 用户：`users`
- 角色：`roles`
- 团队：`tenants`
- 团队成员：`tenant_members`
- 用户角色关联：`user_roles`

### 2.2 权限与功能包

- 权限键主表：`permission_keys`
- 功能包主表：`feature_packages`
- 组合包关系：`feature_package_bundles`
- 功能包绑定权限键：`feature_package_keys`
- 功能包绑定菜单：`feature_package_menus`
- 平台/团队/角色/用户的功能包绑定：
  - `role_feature_packages`
  - `user_feature_packages`
  - `team_feature_packages`

### 2.3 菜单、页面、API

- 菜单：`menus`
- 菜单管理分组：`menu_manage_groups`
- 页面：`ui_pages`
- API 元数据：`api_endpoints`
- API 与权限键绑定：`api_endpoint_permission_bindings`

### 2.4 运行时快照

- 平台用户快照：`platform_user_access_snapshots`
- 平台角色快照：`platform_role_access_snapshots`
- 团队边界快照：`team_access_snapshots`
- 团队角色快照：`team_role_access_snapshots`

## 3. 上下文模型

系统只保留两个正式上下文：

- `platform`
- `team`

规则：
- 无 `X-Tenant-ID` 时走平台上下文
- 有 `X-Tenant-ID` 时走团队上下文
- 团队鉴权前必须确认成员关系与成员状态

功能包上下文规则：
- `platform` 包仅用于平台侧
- `team` 包仅用于团队侧
- `platform,team` 包可同时参与平台和团队链路

## 4. 正式授权模型

### 4.1 功能包

功能包分为：
- 基础包 `base`
- 组合包 `bundle`

规则：
- 基础包直接绑定菜单和权限键
- 组合包只负责组合基础包
- 组合包不直接绑定菜单和权限键

### 4.2 平台侧

平台侧最终能力来自：
- 平台角色功能包并集
- 平台用户直绑功能包
- 平台角色减法：
  - `role_hidden_menus`
  - `role_disabled_actions`
- 平台用户减法：
  - `user_hidden_menus`

运行时主读链：
- 平台菜单、动作、来源映射优先读取 `platform_user_access_snapshots`
- 平台角色候选与减法边界优先读取 `platform_role_access_snapshots`

### 4.3 团队侧

团队侧最终能力来自：
- 团队已开通功能包
- 团队边界减法：
  - `team_blocked_menus`
  - `team_blocked_actions`
- 团队角色减法：
  - `role_hidden_menus`
  - `role_disabled_actions`

运行时主读链：
- 团队边界优先读取 `team_access_snapshots`
- 团队角色优先读取 `team_role_access_snapshots`
- 团队用户最终权限由“成员有效性 + 团队边界 + 团队角色快照”共同决定

### 4.4 兼容层

当前仍保留但不再作为主链的兼容数据包括：
- `user_action_permissions`

它们只能作为兼容例外或审计来源，不能再扩张为主授权模型。

## 5. 权限键与 API

### 5.1 权限键

- 正式字段与正式语义已经收口到 `permission_keys`
- 正式权限标识统一使用点式 `permission_key`
- 历史冒号格式仅保留兼容输入能力，不再作为正式展示和正式存储口径

### 5.2 API 元数据

API 元数据当前主链为：
- 路由注册时声明元数据
- 自动同步到 `api_endpoints`
- 再通过 `api_endpoint_permission_bindings` 绑定权限键

运行时 API 请求顺序：
1. 先检查 API 是否被停用
2. 再按权限键做鉴权

因此 API 不只是业务接口，也是权限和边界的正式载体。

## 6. 菜单与页面在权限体系中的位置

- 菜单控制导航入口是否可见
- 页面控制访问路径、挂载关系、面包屑与访问模式
- 权限键控制动作/API 是否允许执行
- 功能包决定菜单与权限键的候选范围

禁止把以下概念混用：
- 菜单 = 页面
- 页面 = 权限
- API = 功能包

## 7. 快照与刷新策略

当前正式策略：
- 运行时优先读快照
- 写链成功后主动刷新受影响主体
- 快照缺失时允许自动重建并持久化

需要触发刷新的典型变更：
- 功能包绑定关系变化
- 组合包子包变化
- 菜单启停或菜单绑定变化
- 角色绑定包、角色隐藏菜单、角色禁用动作变化
- 用户直绑包、用户隐藏菜单变化
- 团队开包、团队边界变化
- 团队成员角色变化

核心要求：
- 不能只改主表，不刷快照
- 不能把快照缺失当成“天然无权限”

## 8. 新模块接入清单

新增模块时至少检查：

1. 是否新增了正式 `permission_keys`
2. 是否需要新增基础功能包或纳入组合包
3. 是否需要新增菜单入口
4. 是否需要新增页面管理记录
5. 是否需要新增 API 元数据与权限键绑定
6. 是否覆盖了相关快照刷新
7. 是否同步更新了专题文档

## 9. 当前非目标

- 不再恢复“用户个人功能权限直配”作为主链
- 不再恢复“菜单内页类型”作为页面管理主模型
- 不允许前端自行推导一套后端没有定义的权限语义

## 10. 当前应记住的判断标准

- 看模块是否开通：先看功能包
- 看入口是否展示：看菜单与菜单减法结果
- 看页面能否访问：看页面访问模式与挂载继承结果
- 看动作/API 是否允许：看权限键与快照

后续实现都应围绕这条主线继续推进。
