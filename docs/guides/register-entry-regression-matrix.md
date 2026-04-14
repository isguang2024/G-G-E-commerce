# 注册入口重构 — 回归场景矩阵

> AUTH-REG-ENTRY-REFACTOR Stage 7 验收用。逐项执行并记录 PASS/FAIL。

---

## A. 入口解析（Resolver）

| # | 场景 | 操作 | 预期结果 |
|---|------|------|----------|
| A1 | 命中自定义入口 | 访问配置了 host+path_prefix 的注册页 | GetRegisterContext 返回该入口的完整配置 |
| A2 | 未命中任何入口 | 访问未配置 host+path 的注册页 | Fallback 到 default 入口，返回默认配置 |
| A3 | 入口 status=disabled | 访问已禁用入口的路径 | 仍可解析但 allow_public_register=false 时阻止注册 |
| A4 | LoginPageKey 回退链 | 入口未配置 login_page_key | 依次回退：app capabilities → default 模板 |

## B. 注册流程

| # | 场景 | 操作 | 预期结果 |
|---|------|------|----------|
| B1 | auto_login=true 正常注册 | 提交注册表单 | 返回 access_token + landing，前端 finalizeAuthenticatedSession → gotoAfterAuth 跳转 |
| B2 | auto_login=false 待登录 | 提交注册表单 | 返回 pending=true + landing，前端编码 landing 参数跳转登录页 |
| B3 | 公开注册关闭 | allow_public_register=false 时提交注册 | 返回错误"公开注册未开启" |
| B4 | 需要邀请码 | require_invite=true 时不带邀请码 | 返回错误"当前入口需要邀请码" |
| B5 | 需要验证码 | require_captcha=true 时不带 captcha_token | 返回错误"请先完成人机验证" |
| B6 | 角色绑定 | 入口 role_codes=["personal.self_user"] | 注册成功后 user_roles 表有对应记录 |
| B7 | 功能包绑定 | 入口 feature_package_keys=["self_service.basic"] | 注册成功后 user_feature_packages 表有对应记录 |
| B8 | 策略快照冻结 | 注册成功后修改入口配置 | user.register_policy_snapshot 仍保留注册时刻的配置 |
| B9 | Social token 注册 | 携带 social_token 提交注册 | 注册成功后 social_auth_bindings 表有绑定记录 |

## C. 认证后跳转（PostAuthLanding 优先级链）

| # | 场景 | 操作 | 预期结果 |
|---|------|------|----------|
| C1 | 优先级 1: entry target_url | 入口配置 target_url=https://app.example.com/welcome | landing.url = 该 URL，前端 window.location.assign 跳转 |
| C2 | 优先级 2: entry target_app_key | 入口配置 target_app_key + home_path | landing 包含 app_key + home_path，前端 gotoAfterLogin 处理 |
| C3 | 优先级 3: source app 回源 | 请求携带 source_app_key + source_home_path | landing 使用 source 值（当 entry 未配 target） |
| C4 | 优先级 4: 空 landing | 入口和请求均无 target/source | landing 为空对象，前端 fallback 到 '/' 或 '/dashboard/console' |
| C5 | target_url 不安全 | 入口 target_url=javascript:alert(1) | 后端 applyEntryUpsert 拒绝保存；ResolvePostAuthLanding 跳过不安全 URL |
| C6 | target_url 为相对路径 | 入口 target_url=/welcome | 通过安全校验，landing.url = /welcome |

## D. Pending Intent 保留

| # | 场景 | 操作 | 预期结果 |
|---|------|------|----------|
| D1 | 编码 landing 到登录页 | auto_login=false 注册成功 | 跳转登录页 URL 携带 landing_url/landing_app_key/landing_space/landing_path |
| D2 | 登录页恢复 landing | 在携带 landing_* 参数的登录页登录 | buildLandingFromQuery 恢复 landing，gotoAfterAuth 使用恢复的 landing |
| D3 | 无 landing 参数 | 直接访问登录页登录 | buildLandingFromQuery 返回 undefined，使用 redirect 或 '/' |

## E. 5 条认证链路统一跳转

| # | 链路 | 操作 | 预期结果 |
|---|------|------|----------|
| E1 | 普通登录 | 用户名密码登录 | gotoAfterAuth(response.landing \|\| buildLandingFromQuery) |
| E2 | 注册自动登录 | auto_login=true 注册成功 | gotoAfterAuth(response.landing, router, '/dashboard/console') |
| E3 | 注册待登录 | auto_login=false 注册成功后手动登录 | 登录页 gotoAfterAuth(recovered landing) |
| E4 | Centralized callback | 跨 App SSO 回调 | gotoAfterAuth(mergedLanding, router, fallbackPath) |
| E5 | Social callback login | 社交登录已绑定账号 | gotoAfterAuth({home_path, navigation_space_key}) |

## F. Source Context 传递

| # | 场景 | 操作 | 预期结果 |
|---|------|------|----------|
| F1 | Social → Register | 社交回调跳注册页 | URL 携带 source_app_key/source_navigation_space_key/source_home_path |
| F2 | Register → Backend | 注册表单提交 | fetchRegister body 包含 source_app_key/space/path |
| F3 | Backend → Landing | 注册服务内部 | RegisterInput.Source* → PostAuthLandingInput.Source* → LandingInfo（优先级 3） |
| F4 | Login → Register (social) | 登录页 consumeSocialToken 非 login | 跳转注册页转发 source_app_key |

## G. 治理端 CRUD

| # | 场景 | 操作 | 预期结果 |
|---|------|------|----------|
| G1 | 创建入口 | 新建自定义入口（含 role_codes、feature_package_keys） | 创建成功，列表显示新入口 |
| G2 | 编辑入口 | 修改 target_app_key、auto_login 等 | 更新成功，GetRegisterContext 返回新值 |
| G3 | 删除入口 | 删除非保留入口 | 删除成功，列表不再显示 |
| G4 | 系统保留入口 — 禁删 | 尝试删除 default 入口 | 返回"系统保留入口不可删除" |
| G5 | 系统保留入口 — 禁改 code | 尝试修改 default 入口的 entry_code | 返回"系统保留入口不可修改 entry_code" |
| G6 | 系统保留入口 — 禁取消保留 | 尝试 is_system_reserved=false | 返回"系统保留入口不可取消保留标记" |
| G7 | target_url 安全校验 | 创建入口 target_url=data:text/html | 返回"target_url 不安全" |
| G8 | 模板预设创建入口 | 点击「新建入口」下拉选「邀请码注册」模板 | 抽屉打开并预填 allow_public_register=true, require_invite=true |
| G9 | 注册日志查看 | 过滤 entry_code 查询注册记录 | 显示对应记录含入口快照（policy_code 已移除） |

## H. 边界与异常

| # | 场景 | 操作 | 预期结果 |
|---|------|------|----------|
| H1 | 用户名重复 | 注册已存在的用户名 | 事务回滚，返回重复用户名错误 |
| H2 | 密码不一致 | confirm_password != password | 返回"两次密码不一致" |
| H3 | 空用户名/密码 | 提交空值 | 返回"用户名和密码必填" |
| H4 | normalizeRedirect 嵌套 | redirect=redirect%3D%2Fdashboard | 正确解析到 /dashboard |
| H5 | redirect 为登录页 | redirect=/auth/login | 规范化为 '/' |
| H6 | 角色 code 不存在 | 入口 role_codes 含无效 code | 注册成功但跳过不存在的角色（不报错） |

---

**执行方式**：优先通过后端 build + 前端 type-check 验证编译正确性，再逐项手动验证核心链路（B1-B2, C1-C4, E1-E5）。
