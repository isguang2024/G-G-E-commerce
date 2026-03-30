# 系统最小演示数据

> 更新时间：2026-03-30。用于系统收尾回归、演示链路和浏览器联调，不作为正式生产初始化方案。

## 1. 初始化命令

- 执行位置：`backend/`
- 初始化命令：`go run ./cmd/init-demo`
- 默认密码：`Demo123456`
- 运行模式：默认 `GG_SPACE_MODE=single`（单空间运行，Host 解析保留但运行时统一回到 `default`）
- 可选参数：
  - `-password`：覆盖默认密码
  - `-team-name`：覆盖默认团队名称
  - `-space-key`：覆盖非默认菜单空间标识，默认 `ops`
  - `-allow-production`：仅在明确允许时跳过生产环境保护

## 2. 固定账号

### 2.1 平台管理员

- 用户名：`platform_admin_demo`
- 邮箱：`platform_admin_demo@gg.demo`
- 昵称：`平台演示管理员`
- 密码：`Demo123456`
- 角色：平台超级管理员
- 用途：菜单管理、页面管理、菜单空间、平台消息发送与发送记录回归

### 2.2 团队管理员

- 用户名：`team_admin_demo`
- 邮箱：`team_admin_demo@gg.demo`
- 昵称：`团队演示管理员`
- 密码：`Demo123456`
- 团队身份：`演示团队 / team_admin`
- 用途：团队消息发送、团队模板、团队发送记录、团队消息中心回归

### 2.3 普通成员

- 用户名：`member_demo`
- 邮箱：`member_demo@gg.demo`
- 昵称：`普通演示成员`
- 密码：`Demo123456`
- 团队身份：无固定团队归属
- 用途：菜单可见性、页面准入、平台消息中心收件回归

## 3. 固定团队与菜单空间

### 3.1 默认团队

- 团队名称：`演示团队`
- 团队 owner：`team_admin_demo`
- 成员：
  - `team_admin_demo`

### 3.2 非默认菜单空间

- 空间标识：`ops`
- 空间名称：`运营空间`
- 默认首页：`/dashboard/console`
- 准入规则：`仅平台管理员`
- 初始化方式：每次执行 `init-demo` 都会强制从 `default` 复制菜单、页面与功能包菜单关联到 `ops`，用于保留管理员全量配置空间。

### 3.3 默认空间基线（兜底）

- 空间标识：`default`
- 运行用途：所有运行时兜底空间（`single` 模式下始终使用）
- 菜单基线：恢复为默认全量系统菜单
- 页面基线：恢复为默认全量系统页面注册
- 目标：先保证本地调试可直接进入系统管理空间，后续再按需要做精简

## 4. 消息演示数据

### 4.1 发送人

- 平台：
  - `平台`（默认）
  - `平台管理`
- 团队：
  - `团队`（默认）
  - `团队管理`

### 4.2 模板

- 平台模板：`demo.wrapup.platform.notice`
- 团队模板：`demo.wrapup.team.notice`

### 4.3 接收组

- 平台接收组：`演示平台接收组`
  - 指定用户：`team_admin_demo`
  - 指定用户：`member_demo`
  - 角色规则：`admin`
- 团队接收组：`演示团队接收组`
  - 当前团队成员
  - 角色规则：`team_admin`
  - 功能包规则：`team.member_admin`

### 4.4 演示消息

- 平台消息业务标识：`demo.system_wrapup.platform`
  - 面向：`platform_admin_demo`、`member_demo`
  - 带内部链接：`#/system/page`、`/system/menu`、`#/workspace/inbox`
- 团队消息业务标识：`demo.system_wrapup.team`
  - 面向：`演示团队`
  - 通过团队接收组展开到 `team_admin_demo`

## 5. 推荐联调顺序

1. `platform_admin_demo`
   - 登录
   - 进入 `#/system/menu`
   - 进入 `#/system/page`
   - 进入 `#/system/menu-space`
   - 进入 `#/system/message`
   - 进入 `#/workspace/inbox`
2. `team_admin_demo`
   - 登录
   - 切到 `演示团队`
   - 进入 `#/team/message`
   - 查看团队消息中心内容
3. `member_demo`
   - 登录
   - 验证菜单裁剪
   - 打开消息中心，查看平台消息
   - 验证无法进入 `#/team/message`

## 6. 使用约束

- 此命令用于非生产环境，不用于正式数据初始化。
- 命令会重置上述演示账号密码，并刷新相关权限快照。
- 命令会清理并重建 `demo.system_wrapup.platform`、`demo.system_wrapup.team` 相关演示消息，避免重复积累。
- 当前演示策略为：团队管理员承担团队链路回归，普通成员只承担平台收件链路回归，不再共享团队成员身份。
- 若需要追加新的系统回归场景，应在此文档登记命名与用途，避免后续临时造数据。
