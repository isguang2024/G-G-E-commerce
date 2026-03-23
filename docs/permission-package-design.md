# 权限与功能包设计现状

## 1. 目标
当前权限模型已经从“API 粒度动作权限”收敛为“业务功能权限 + 功能包 + 团队上下文”。

固定目标：
- 平台管理员负责平台级管理能力。
- 大部分业务功能默认运行在团队上下文。
- 平台通过功能包向团队开通能力。
- 团队内部再把能力分配给角色和成员。
- 单团队用户自动进入默认团队，多团队用户才显示切换。

## 2. 当前已完成

### 2.1 功能权限主模型
- 系统以 `permission_key` 为主授权单元。
- `scope` 已从权限主链移除。
- `category` 已删除。
- `resource_code/action_code` 仅保留兼容和 API 注册用途，不再是主模型。
- 前后端菜单、按钮、页面、运行时鉴权都以 `permission_key` 为核心。

示例：
- `system.menu.manage`
- `tenant.member.manage`
- `team.member.assign_role`
- `platform.package.manage`
- `platform.package.assign`

### 2.2 上下文模型
- 已明确支持 `platform` / `team` 两类上下文。
- `permission_actions` 已增加 `context_type`。
- `feature_packages` 已增加 `context_type`。
- `roles` 已增加 `tenant_id`，支持平台角色、基础团队角色、团队自定义角色并存。

### 2.3 功能包模型
已落地表结构：
- `feature_packages`
- `feature_package_actions`
- `team_feature_packages`
- `team_manual_action_permissions`

已落地能力：
- 功能包 CRUD。
- 功能包包含权限配置。
- 从团队侧给团队开通功能包。
- 从功能包侧查看和配置开通团队。
- 功能包列表返回权限数、团队数统计。

默认种子：
- `platform.system_admin`
- `platform.menu_admin`
- `platform.api_admin`
- `team.member_admin`

### 2.4 API 注册关系
当前正式关系：

`功能模块 -> 功能权限(permission_key) -> 多个 API`

`功能包 -> 多个功能权限(permission_key)`

约束：
- API 继续存在注册表。
- API 归属到功能权限。
- 功能包不直接包含 API。

### 2.5 团队边界链路
当前团队边界已经拆成三层：
- `team_feature_packages` + `feature_package_actions`：功能包展开来源。
- `team_manual_action_permissions`：团队手工补充权限。
- `tenant_action_permissions`：当前最终生效边界。

当前能力：
- 团队权限页可看到已开通功能包。
- 团队权限页可区分“功能包展开”和“额外补充”来源摘要。
- 团队成员个人权限覆盖只允许落在团队已开通能力内。
- 团队角色功能权限读取和保存已按团队边界收口。

### 2.6 团队角色模型
当前已完成：
- `roles.tenant_id` 已落库并已迁移。
- 平台角色查询默认只看 `tenant_id IS NULL`。
- 团队侧角色候选集为“基础团队角色 + 当前团队自定义角色”。
- 平台侧系统角色接口禁止直接维护团队自定义角色。
- 已提供当前团队角色 CRUD。
- 已提供当前团队角色菜单权限维护。
- 已提供当前团队角色功能权限维护。
- 前端已有“当前团队角色管理”页面。

## 3. 当前正式规则

### 3.1 权限永远是业务功能权限
系统最小授权单元是 `permission_key`。

禁止回退到：
- 每个 API 一条主权限。
- 每个 CRUD 一条主权限。
- `scope` 参与主权限判定。

### 3.2 功能包只包含功能权限
固定关系：
- `功能包 -> 多个 permission_key`
- `permission_key -> 多个 API`

禁止：
- `功能包 -> API`

### 3.3 平台包和团队包分离
功能包必须带 `context_type`：
- `platform`
- `team`

规则：
- 平台包用于平台管理侧。
- 团队包用于团队业务侧。
- 不允许把平台包直接当团队包使用。

### 3.4 平台与团队职责分离
平台负责：
- 管模块。
- 管功能权限。
- 管功能包。
- 管团队。
- 给团队开通团队包。

团队负责：
- 在已开通能力内使用业务。
- 给团队角色和成员分配能力。

### 3.5 单团队默认，多团队切换
产品规则：
- 没有团队：只能进入平台能力或引导能力。
- 只有一个团队：自动默认，不显示切换。
- 多个团队：显示切换。

## 4. 现有表的定位

### 4.1 正式模型表
- `permission_actions`：定义功能权限。
- `role_action_permissions`：角色拥有的功能权限。
- `user_action_permissions`：用户个人覆盖。
- `user_roles`：角色分配关系。
- `roles`：平台角色 / 基础团队角色 / 团队自定义角色定义。
- `feature_packages`：功能包定义。
- `feature_package_actions`：功能包包含的功能权限。
- `team_feature_packages`：团队已开通的功能包。
- `team_manual_action_permissions`：团队手工补充权限。

### 4.2 当前仍作为生效边界使用的表
- `tenant_action_permissions`

当前语义：
- 仍参与运行时团队边界判定。
- 当前由“功能包展开权限 + 团队手工补充权限”汇总写入。
- 不再承载团队原始手工配置来源。

长期目标：
- 逐步降级为团队边界展开缓存。

## 5. 新模块接入规范
新增模块时固定顺序：
1. 定义模块。
2. 定义 `permission_key`。
3. 明确 `context_type`。
4. 决定这些权限归属哪些功能包。
5. 最后注册 API 和页面。

推荐命名：
`领域.对象.能力`

示例：
- `product.series.manage`
- `product.series.publish`
- `channel.series.manage`
- `channel.series.sync`

推荐能力词：
- `manage`
- `view`
- `assign`
- `publish`
- `sync`
- `export`
- `import`

## 6. 当前未完成

### 6.1 `tenant_action_permissions` 还不是纯缓存
目标状态：
- 功能包决定团队具备哪些能力。
- 团队手工补充权限单独存储。
- `tenant_action_permissions` 只保存最终生效结果或缓存，并通过统一刷新链路重建。

当前虽然已拆出 `team_manual_action_permissions`，但运行时和部分管理逻辑仍直接依赖 `tenant_action_permissions` 作为最终判定依据。

### 6.2 团队角色与功能包边界的联动还可继续收口
当前已完成团队角色 CRUD 和菜单/功能权限维护，但仍可继续加强：
- 团队角色功能权限界面更直观地标记哪些能力来自已开通功能包。
- 团队成员分配角色界面继续强化“基础角色 / 团队自定义角色”语义。
- 团队角色如果后续需要数据权限，还缺独立模型。

### 6.3 团队内分配仍需继续收口到“仅可分配已开通能力”
目标状态：
- 团队内所有角色/成员分配都严格受功能包边界限制。

当前方向已经明确，成员个人权限覆盖和角色读取已收口，但前端交互和部分链路仍可继续加强。

## 7. 当前结论
这套权限模型已经进入稳定中期状态：
- 功能权限以 `permission_key` 为主。
- 权限支持 `platform/team` 上下文。
- 功能包模型已落库并已接入前后端。
- 功能包与团队开通链路已跑通。
- API 注册已经退回为元数据层，不再是权限主模型。
- `scope` 和旧分类模型已退出主链。

当前剩余的核心任务不是再改命名，而是继续把：
- 团队角色/成员分配。
- 运行时团队边界刷新。

完全收口到“功能包展开 + 手工补充 + 最终缓存”的三层模型。
