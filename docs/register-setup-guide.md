# 注册体系配置与验证手册

## 目标

这份手册回答 4 个问题：

1. 注册页现在归谁管
2. 注册模板放在哪里配
3. 入口、策略、公开页应该按什么顺序配起来
4. 配完以后怎么验证是否真的生效

## 当前真相

- 公开认证页归属到 `account-portal` App。
- `account-portal` 负责承载以下公开页：
  - `/account/auth/login`
  - `/account/auth/register`
  - `/account/auth/forget-password`
- “注册模板”不单独建新模型，直接由“注册策略”承载。
- “注册入口”负责按 `host + path_prefix` 命中策略，并可对 `allow_public_register` 做入口级覆盖。

## 你应该先配什么

推荐顺序固定如下：

1. 先确认默认 seed 已存在
2. 再创建或编辑注册策略
3. 再让注册入口引用该策略
4. 最后访问公开注册页验证命中结果

原因：

- 策略定义的是业务结果，例如是否开放公开注册、注册后 landing 到哪里、需要哪些附加校验。
- 入口只负责把 URL 命中到策略。
- 公开注册页只是把当前命中的入口和策略渲染出来，不负责决定业务规则。

## 默认可用方案

默认 seed 提供了一条可直接验证的最小链路：

- App: `account-portal`
- 默认公开页空间: `public`
- 注册策略: `default.self`
- 注册入口: `default`
- 注册页路径: `/account/auth/register`
- 注册后目标 App: `platform-admin`
- 目标空间: `self-service`
- 目标首页: `/self/user-center`

如果只想先跑通一条最小可用链路，优先基于这套默认值验证。

## 注册策略怎么理解

注册策略就是模板，核心回答这几个问题：

- 是否允许公开注册
- 是否要求邀请码
- 是否要求邮箱验证
- 是否要求验证码 / 人机验证
- 注册成功后是否自动登录
- 注册后进入哪个 App / 空间 / 首页
- 默认绑定哪些角色和功能包

建议优先使用 3 类模板：

### 1. 默认公开注册

适用场景：

- 面向普通自助用户
- 注册成功后直接进入个人自助空间

建议值：

- `allow_public_register = true`
- `require_invite = false`
- `require_email_verify = false`
- `require_captcha = false`
- `auto_login = true`
- `target_app_key = platform-admin`
- `target_navigation_space_key = self-service`
- `target_home_path = /self/user-center`
- `role_codes = [personal.self_user]`
- `feature_package_keys = [self_service.basic]`

### 2. 邀请码注册

适用场景：

- 活动邀约
- 私域导入
- 不希望任何匿名访问者直接完成注册

建议值：

- 在默认公开注册基础上开启 `require_invite = true`
- 常见做法是 `auto_login = false`

### 3. 邮箱验证注册

适用场景：

- 对账户真实性要求更高
- 需要后续邮件链路

建议值：

- 在默认公开注册基础上开启 `require_email_verify = true`
- 若风险更高，可叠加 `require_captcha = true`

## 注册入口怎么理解

注册入口决定“哪个 URL 用哪套策略”。

常用字段：

- `app_key`
  - 公开认证页建议固定为 `account-portal`
- `host`
  - 留空表示任意域名
  - 如果使用子域名部署，可填 `account.example.com`
- `path_prefix`
  - 推荐默认值：`/account/auth/register`
  - 按前缀匹配，越具体越优先
- `policy_code`
  - 指向某个注册策略
- `allow_public_register`
  - 留空表示继承策略
  - 显式设置后，优先级高于策略同名开关

## 页面归属原则

公开注册页、登录页、找回密码页都应该归属 `account-portal`，不要再把它们长期维护成纯本地静态页。

原因：

- 这样页面管理、App 归属、入口命中和运行时公开路由是同一套真相。
- 否则会出现“文档说属于 account-portal，代码却在 staticRoutes 里硬编码”的治理断层。
- 后续新增公开认证页时，也可以继续复用 `account-portal` 的 `ui_pages` 管理方式。

## 怎么验证

每次配置后，至少按下面顺序验一遍：

1. 打开后台“注册策略”，确认策略已启用，landing 正确。
2. 打开后台“注册入口”，确认入口 URL、策略引用和公开注册覆盖值正确。
3. 访问 `/account/auth/register`。
4. 查看页面顶部上下文区：
   - 是否显示命中的入口 Code
   - 是否显示策略 Code
   - 是否显示注册来源
   - 是否显示注册后去向
   - 是否正确展示邮箱 / 邀请码 / 验证码字段
5. 提交注册，确认：
   - `auto_login=true` 时会直接进入目标首页
   - `auto_login=false` 时会回到登录页
6. 打开“注册记录”，检查：
   - 是否记录入口 Code
   - 是否记录策略 Code
   - 是否记录策略快照

## 排查顺序

如果页面提示“当前未开启公开注册”或字段不符合预期，按这个顺序排查：

1. 入口是否真的命中了当前 URL
2. 入口是否引用了预期策略
3. 入口级 `allow_public_register` 是否覆盖了策略
4. 策略是否启用了邀请码 / 邮箱验证 / 人机验证
5. `account-portal` 的公开页 seed 是否已执行

## 本次收口后的规则

- 公开认证页归属 `account-portal`
- 公开认证页通过页面管理运行时注册为 `public` 页面
- 注册模板统一由“注册策略”承载
- 后台页必须直接说明“先配策略，再配入口，最后访问 URL 验证”
