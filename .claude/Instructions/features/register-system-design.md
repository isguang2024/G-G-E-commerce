# 注册体系设计文档（三层模型）

> 版本：v2 · 日期：2026-04-09 · 状态：设计已确认，待进入实施计划
>
> 目标：从第一天起就构建独立的 `account-portal` 认证中心 App，与 `platform-admin` 共享同一个后端进程和数据库，但在 `apps` / `AppHostBinding` / `MenuSpace` / 菜单 / 权限层面完全隔离。数据模型与后端判定逻辑按"未来还会有 `user-portal` / `merchant-console`"预留。
>
> **v2 变更点**：删除"模式 A 单 App"路径，直接按"模式 B 多 App 共享后端"实施。`account-portal` 作为独立 App 从第一期就建好，公开注册入口挂在它下面；`platform-admin` 完全不开放公开注册。

---

## 0. TL;DR

- **共享后端 + 多 App 架构**：一个 Go 后端进程、一个数据库，在 `apps` 表里同时存在 `platform-admin`（治理）和 `account-portal`（认证中心）两条记录，通过 `AppHostBinding` + `app_key` 实现路由、菜单、权限的完全隔离。
- **三层模型**：入口层（`register_entries`，归属 `account-portal`）→ 策略层（`register_policies`，`target_app_key` 解耦）→ 承载层（目标 App + MenuSpace + 功能包 + 角色）。
- **入口判定由服务端基于 `host + path` 命中 `register_entries`**，前端传来的 `register_source` 仅作辅助、不作真相。
- **`platform-admin` 完全不开放公开注册**。公开注册入口全部归属 `account-portal`；注册成功后按策略 `target_app_key` 分流到承载 App（第一期承载 App 仍是 `platform-admin`，MenuSpace 进 `self-service`；未来新增 `user-portal` 时只改策略表，不改模型）。
- **`account-portal` 自带最小 MenuSpace**，只包含注册、登录、邮箱验证、找回密码、邀请接受等公开页，不暴露任何业务菜单。
- **第一期落地范围**：数据模型 + `account-portal` App seed + 注册 seed + `/auth/register-context` + 改造 `/auth/register` + 后台三个管理页（入口 / 策略 / 注册记录）。

---

## 1. 仓库现状快照

调研于 2026-04-09 完成，关键结论：

| 能力 | 现状 | 第一期是否需要改动 |
|---|---|---|
| `App` / `app_key`（默认 `platform-admin`） | ✅ 已有，`models/app.go` | **新增 `account-portal` App 记录** |
| `AppHostBinding`（host+path 匹配） | ✅ 已有，`models/app.go` | **新增 `account-portal` 的 host/path 绑定** |
| `MenuSpace`（导航空间） | ✅ 已有，`models/menu_space.go`，默认 `default` | **新增 `account-portal` 的 `public` space + `platform-admin` 的 `self-service` space** |
| `FeaturePackage`（功能包） | ✅ 已有，`models/model.go` | 新增 `self_service.basic` |
| `Role`（角色） | ✅ 已有，`models/model.go` | 新增 `personal.self_user` |
| `Workspace` / 个人空间 lazy 创建 | ✅ `workspace/service.go:62` | 改为注册时即刻创建 |
| `users.register_source` / `invited_by` | ✅ 已有 | 复用 |
| `users.register_ip` / `email_verified_at` / `register_app_key` / `register_entry_code` / `register_policy_code` / `user_agent` / `agreement_version` / `invitation_code_id` | ❌ 缺失 | 第一期新增 |
| `/auth/register` handler | ✅ `api/handlers/auth.go:49`，service `modules/system/auth/service.go:128` | 改造 |
| `/auth/register-context` | ❌ 不存在 | 第一期新增 |
| OpenAPI spec（auth 域） | ✅ `api/openapi/domains/auth/paths.yaml` | 扩展 |
| 迁移：Goose + GORM AutoMigrate | ✅ `internal/pkg/database/migrations/` | 新增 `00006_register_system.sql` |
| 前端注册页 | ✅ `frontend/src/views/auth/register/index.vue`，当前 API 调用被注释 | 接入真实接口 |

---

## 2. 核心概念

| 概念 | 表达 | 职责 |
|---|---|---|
| **入口（Entry）** | `register_entries` | 承接公开访问入口（host+path / 邀请码链路 / 活动渠道），回答"从哪里进" |
| **策略（Policy）** | `register_policies` | 回答"注册后给什么"（目标 App、MenuSpace、首页、功能包、角色、开关） |
| **承载（Landing）** | 由策略计算得出的 `{target_app_key, target_navigation_space_key, target_home_path, workspace}` | 回答"用户最终进入哪里" |
| **来源（Source）** | `register_source` 字符串 | 审计维度：`self` / `invite` / `merchant_campaign` / `partner` |

**关键解耦**：`register_entries.app_key`（入口归属 App）与 `register_policies.target_app_key`（承载 App）是两个字段。当前单 App 时两者都是 `platform-admin`；未来 `account-portal` 上线后，入口侧切换，承载侧不动。

---

## 3. 部署架构

### 3.1 共享后端，多 App 并存

- **一个 Go 后端进程 + 一个 PostgreSQL 实例**，承载 `apps` 表里的多条 App 记录。
- App 之间通过 `app_key` 做逻辑隔离：菜单（`menus.app_key`）、导航空间（`menu_spaces.app_key`）、功能包（`feature_packages.app_key`）、入口绑定（`app_host_bindings.app_key`）、路由注解（`routes.app_key`，如有）全部按 `app_key` 过滤。
- **路由归属判定**：请求进入时，服务端根据 `host + path` 命中 `AppHostBinding`，得到当前请求的 `current_app_key`；后续鉴权、菜单加载、注册入口解析全部以此为准。
- 所有 App 共用同一套 `users` / `roles` / `workspaces` 表（用户是平台级资产，不按 App 切分）。

### 3.2 第一期的 App 清单

| `app_key` | 用途 | 是否开放公开注册 | 代表性 MenuSpace |
|---|---|---|---|
| `platform-admin` | 平台治理后台 | ❌ | `default`（现有） + `self-service`（新增，用于承载自注册用户的自助页） |
| `account-portal` | 认证中心（公开入口） | ✅ | `public`（新增，只含注册/登录/邮箱验证/找回密码/邀请接受） |

> 第一期**不新建** `user-portal`。自注册用户登录后，承载 App 仍是 `platform-admin`，但 MenuSpace 为 `self-service`——借助 App 内 MenuSpace 隔离，治理用户与自助用户虽共用一个业务 App 但看不到对方的菜单。未来要拆 `user-portal` 时只需新增 App 记录 + 改策略 `target_app_key`，无需迁移用户或权限。

### 3.3 双通道路由：`host` 与 `path` 同时生效

`AppHostBinding` 从第一天起同时支持 **host 绑定** 和 **path 前缀绑定**两种 `match_type`，两者都 seed 好，覆盖"生产子域名部署"和"本地/单域名部署"两种场景，互不冲突：

**Host 绑定**（生产典型形态）：

```
account.example.com/*                 → account-portal / public
admin.example.com/*                   → platform-admin / default
admin.example.com/self/*              → platform-admin / self-service
```

**Path 绑定**（本地开发 / 单域名部署）：

```
/account/*                            → account-portal / public
/self/*                               → platform-admin / self-service
/  或 /admin/*                        → platform-admin / default   （兜底）
```

**匹配优先级**（resolver 按此顺序遍历 `app_host_bindings`）：

1. `match_type = host` 且 host 精确匹配（优先匹配更精细的 path_pattern）。
2. `match_type = path` 且 path 前缀匹配（按前缀长度降序）。
3. 同级内按 `priority` 字段降序。
4. 全部未命中 → 默认兜底 App（当前为 `platform-admin` / `default`）。

> 同一套前端与后端二进制既能跑 `localhost:5173/account/auth/register`，也能跑 `https://account.example.com/auth/register`，由部署侧在 `app_host_bindings` 表里决定走哪条。

---

## 4. 数据模型

### 4.1 `users` 表扩展字段

在现有字段基础上新增：

| 字段 | 类型 | 说明 |
|---|---|---|
| `register_app_key` | varchar(64) | 命中的入口归属 App；默认 `platform-admin` |
| `register_entry_code` | varchar(64) | 命中的入口编码；可空（无入口配置时） |
| `register_policy_code` | varchar(64) | 实际套用的策略编码 |
| `register_ip` | varchar(64) | 注册 IP（服务端捕获） |
| `register_user_agent` | varchar(512) | 注册 UA |
| `agreement_version` | varchar(32) | 同意的协议版本 |
| `email_verified_at` | timestamptz | 邮箱验证时间，null 表示未验证 |
| `invitation_code_id` | uuid | 邀请码 ID，FK 到未来的 `invitation_codes` |

> `register_source` 与 `invited_by` 已存在，复用。

### 4.2 `register_entries`（注册入口表）

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uuid (PK) | |
| `app_key` | varchar(64) | 入口归属 App |
| `entry_code` | varchar(64), UNIQUE | 入口编码，业务主键 |
| `name` | varchar(128) | 展示名 |
| `host` | varchar(128) | 匹配 host；空表示不限定 |
| `path_prefix` | varchar(256) | 匹配 path 前缀；空表示不限定 |
| `register_source` | varchar(32) | 命中该入口时写入 `users.register_source` |
| `policy_code` | varchar(64) | 关联的策略 |
| `status` | varchar(16) | `enabled` / `disabled` |
| `allow_public_register` | bool | 入口级开关（可覆盖策略默认值，便于一键关停） |
| `require_invite` | bool | |
| `require_email_verify` | bool | |
| `require_captcha` | bool | |
| `auto_login` | bool | 注册成功后是否直接签发 token |
| `sort_order` | int | 命中优先级 |
| `remark` | text | |
| `created_at` / `updated_at` / `deleted_at` | | |

**唯一约束**：`(host, path_prefix)` 在 `status=enabled` 时唯一（通过部分索引）。

**匹配算法**：`host` 精确匹配（空视为通配）；`path_prefix` 按最长前缀；再按 `sort_order` 降序；均不命中时 fallback 到 `entry_code = "default"` 的内置入口。

### 4.3 `register_policies`（注册策略表）

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uuid (PK) | |
| `app_key` | varchar(64) | 策略归属 App（通常同入口） |
| `policy_code` | varchar(64), UNIQUE | 策略编码 |
| `name` | varchar(128) | |
| `description` | text | |
| `target_app_key` | varchar(64) | **承载 App**（与 `app_key` 解耦） |
| `target_navigation_space_key` | varchar(64) | 目标 MenuSpace，例如 `self-service` |
| `target_home_path` | varchar(256) | 默认首页，例如 `/user-center` |
| `default_workspace_type` | varchar(32) | 默认创建的工作空间类型，一期固定 `personal` |
| `status` | varchar(16) | `enabled` / `disabled` |
| `welcome_message_template_key` | varchar(128) | 欢迎消息模板（可空） |
| `allow_public_register` | bool | 策略级默认开关 |
| `require_invite` / `require_email_verify` / `require_captcha` / `auto_login` | bool | 策略级默认 |
| `created_at` / `updated_at` / `deleted_at` | | |

> 入口上的同名字段为"显式覆盖"；未设置时继承策略。第一期采用"入口字段非 null 即覆盖"语义（`bool` 字段通过 `*bool` 区分）。

### 4.4 `register_policy_feature_packages`

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uuid (PK) | |
| `policy_code` | varchar(64) | FK → `register_policies.policy_code` |
| `package_id` | uuid | FK → `feature_packages.id` |
| `workspace_scope` | varchar(32) | `personal` / `collaboration` / `global` |
| `sort_order` | int | |

### 4.5 `register_policy_roles`

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uuid (PK) | |
| `policy_code` | varchar(64) | FK → `register_policies.policy_code` |
| `role_id` | uuid | FK → `roles.id` |
| `workspace_scope` | varchar(32) | |
| `sort_order` | int | |

### 4.6 迁移文件规划

- `backend/internal/pkg/database/migrations/00006_register_system.sql`：新增三张新表 + users 扩展字段（DDL）。
- 其余实体（字段类型、索引细化）交给 GORM AutoMigrate 补充。

---

## 5. Seed 设计（第一期内置数据）

在 `backend/internal/pkg/permissionseed/seeds.go` 中扩展，或新增 `registerseed/seeds.go`：

### 5.0 新增 App：`account-portal`

```
app_key = "account-portal"
name = "认证中心"
space_mode = "single"           # 单 MenuSpace 模式
default_space_key = "public"
is_default = false
```

对应的 `AppHostBinding` seed（双通道，两条都建）：

```
# A. path 绑定（本地 / 单域名）
app_key = "account-portal"
match_type = "path"
host = ""
path_pattern = "/account"
priority = 100
status = "enabled"

# B. host 绑定（生产子域名，示例，默认 disabled 由运维启用）
app_key = "account-portal"
match_type = "host"
host = "account.example.com"
path_pattern = ""
priority = 200
status = "disabled"
```

同样地，`platform-admin` 的 `self-service` space 也新增一条 path 绑定 `/self` + 一条 host 绑定示例（disabled）。

### 5.1 新增 MenuSpace

**A. `account-portal` / `public`**（认证中心自带的公开页空间）

```
app_key = "account-portal"
space_key = "public"
name = "公开入口"
is_default = true
default_home_path = "/account/auth/login"
```

包含的菜单（仅公开页，无需登录）：
- 登录 `/account/auth/login`
- 注册 `/account/auth/register`
- 邮箱验证 `/account/auth/verify-email`（预留）
- 找回密码 `/account/auth/forgot-password`（预留）
- 邀请接受 `/account/invite/accept`（预留）

**B. `platform-admin` / `self-service`**（承载自注册用户登录后自助页）

```
app_key = "platform-admin"
space_key = "self-service"
name = "自助中心"
is_default = false
default_home_path = "/self/user-center"
```

### 5.2 新增 FeaturePackage：`self_service.basic`

- `app_key = "platform-admin"`
- `package_key = "self_service.basic"`
- `package_type = "self_service"`
- `context_type = "personal"`
- `workspace_scope = "personal"`
- `is_builtin = true`
- 包含的菜单（限定在 `self-service` space）：
  - 工作台首页 `/dashboard/home`
  - 个人中心 `/user-center`
  - 账号设置 `/user-center/settings`
  - 收件箱 `/inbox`
  - 我的协作空间 `/workspaces/mine`
  - 加入协作空间 `/workspaces/join`
  - 帮助与支持 `/help`
- 包含的权限键：`workspace.read`、`workspace.switch`、以及将来新增的 `self_service.*` 读权限。
- **绝不包含**：`user.*` / `role.*` / `permission.*` / 菜单管理 / 功能包管理 / API 管理 等治理权限。

### 5.3 新增 Role：`personal.self_user`

- `code = "personal.self_user"`
- `name = "个人自助用户"`
- `is_system = true`
- `app_keys = ["platform-admin"]`
- 绑定功能包：`self_service.basic`（`workspace_scope = "personal"`）

### 5.4 默认 `register_policies` 种子

```
policy_code = "default.self"
app_key = "account-portal"              # 策略归属认证中心
target_app_key = "platform-admin"       # 承载 App 仍是治理后端，但进 self-service space
target_navigation_space_key = "self-service"
target_home_path = "/self/user-center"
default_workspace_type = "personal"
status = "enabled"
allow_public_register = false           # 默认关闭，部署方到后台开启
require_invite = false
require_email_verify = false
require_captcha = false
auto_login = true
```

绑定：`register_policy_feature_packages` → `self_service.basic`；`register_policy_roles` → `personal.self_user`。

> 未来新增 `user-portal` App 时，只需把 `target_app_key` 改为 `user-portal`、`target_navigation_space_key` 改为该 App 的默认 space，**策略表以外的任何结构都不动**。

### 5.5 默认 `register_entries` 种子

```
entry_code = "default"
app_key = "account-portal"                    # 入口归属认证中心
name = "默认公开注册入口"
host = ""                                      # 通配
path_prefix = "/account/auth/register"
register_source = "self"
policy_code = "default.self"
status = "enabled"
allow_public_register = null                  # 继承策略
...其它开关均 null
sort_order = 0
```

---

## 6. 入口识别与策略解析

### 6.1 识别优先级

1. 邀请码链路命中邀请上下文（第一期未实现，预留钩子）。
2. `host + path_prefix` 命中 `register_entries`（`status=enabled`）。
3. 渠道参数 `?channel=xxx` 作为补充命中（第一期预留）。
4. 前端 `register_source` **仅用于审计日志辅助**，不参与命中。
5. 兜底：`entry_code = "default"`。

### 6.2 策略合并

`effective = merge(policy, entry_overrides)`：入口字段非 null 则覆盖策略同名字段。结果称为 **EffectiveRegisterContext**，贯穿 `/auth/register-context` 和 `/auth/register` 两个接口。

### 6.3 服务接口（内部）

新增 `backend/internal/modules/system/register/`：
- `resolver.go`：`ResolveEntry(ctx, host, path) (*Entry, *Policy, error)`
- `context.go`：`BuildContext(entry, policy) *EffectiveRegisterContext`
- `service.go`：`Register(ctx, req, effective) (*LoginResponse | *RegisterPendingResponse, error)`
- `landing.go`：`BuildLanding(user, effective) *LandingInfo`

---

## 7. 接口设计

### 7.1 `GET /auth/register-context`

**Query**：`?host=...&path=...`（前端自带当前 location）；服务端优先信任请求 host header。

**Response**：

```json
{
  "entry_code": "default",
  "entry_name": "默认公开注册入口",
  "entry_app_key": "platform-admin",
  "register_source": "self",
  "policy_code": "default.self",
  "allow_public_register": false,
  "require_invite": false,
  "require_email_verify": false,
  "require_captcha": false,
  "auto_login": true,
  "agreement_version": "v1",
  "brand_info": { "app_name": "GGE", "logo_url": "..." },
  "field_schema": {
    "username": { "required": true, "min": 3, "max": 32 },
    "password": { "required": true, "min": 8 },
    "email":    { "required": false },
    "nickname": { "required": false },
    "invitation_code": { "required": false }
  }
}
```

若 `allow_public_register = false`，前端渲染"当前入口未开放注册"。

### 7.2 `POST /auth/register`（改造）

**Request**：

```json
{
  "username": "...",
  "password": "...",
  "confirm_password": "...",
  "email": "...",
  "nickname": "...",
  "captcha_token": "...",
  "invitation_code": "...",
  "agreement_version": "v1"
}
```

**服务端处理顺序**：

1. `ResolveEntry(host, path)` → `(entry, policy)` → `effective`。
2. 校验 `effective.allow_public_register == true`，否则 403。
3. 按 `effective` 校验：captcha / invite / email_verify 要求。
4. 字段校验（TrimSpace、长度、字符白名单、保留词、唯一性、密码复杂度）。
5. 协议版本校验。
6. IP / 设备限流（第三期，第一期预留 hook）。
7. 创建 `users`，写入审计字段（`register_app_key` / `register_entry_code` / `register_policy_code` / `register_source` / `register_ip` / `register_user_agent` / `agreement_version`）。
8. **即刻**创建个人 workspace（改掉当前 lazy 模式），并写入 `workspace_members` owner。
9. 按 `register_policy_feature_packages` / `register_policy_roles` 绑定功能包与角色。
10. 若 `auto_login=true`，签发 token，返回 `LoginResponse + landing`；否则返回 `pending` 状态，提示去登录或验证邮箱。

**Response（auto_login=true）**：

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 3600,
  "user": { ... },
  "landing": {
    "app_key": "platform-admin",
    "navigation_space_key": "self-service",
    "home_path": "/self/user-center"
  }
}
```

**Response（auto_login=false）**：

```json
{
  "status": "pending",
  "next_action": "login" | "verify_email",
  "message": "注册成功，请登录"
}
```

### 7.3 预留接口（第三期）

- `POST /auth/register/send-email-code`
- `POST /auth/register/verify-email`
- `GET  /auth/invite-context`

### 7.4 OpenAPI spec 改动

- `api/openapi/domains/auth/paths.yaml`：新增 `/auth/register-context`，改造 `/auth/register` 的 request/response schema。
- `api/openapi/domains/auth/schemas.yaml`：把 `RegisterRequest` / `LoginResponse` 从 `common.yaml` 中迁出，并新增 `RegisterContext` / `LandingInfo` / `RegisterPendingResponse`。
- `make api-bundle && make api-gen` 重新生成 client。

---

## 8. 默认权限原则

**最小可用权限**。自注册用户默认**只能**访问：

- 工作台首页、个人中心、账号设置、收件箱、我的协作空间、加入协作空间、帮助与支持。

**绝不默认开放**：用户管理、角色管理、菜单管理、页面管理、功能包管理、权限键管理、API 管理、App 管理、治理型消息模板管理。

通过"`self-service` MenuSpace + `self_service.basic` 功能包 + `personal.self_user` 角色"三件套自然实现：自注册用户登录后，前端根据 `landing.navigation_space_key = "self-service"` 渲染对应菜单树，治理菜单物理上不会被拉取。

---

## 9. 前端改动

### 9.0 前端 App 划分

前端采用**单 SPA + App 路由分区**策略，不拆仓库：

```
frontend/src/views/
  account-portal/          # 归属 account-portal App（公开）
    auth/
      login/               # 从 views/auth/login 迁入；旧路径 307 redirect
      register/            # 从 views/auth/register 迁入；旧路径 307 redirect
      verify-email/        # 预留
      forgot-password/     # 预留
    invite/                # 预留
  platform-admin/          # 归属 platform-admin App（治理 + self-service）
    (现有治理页)
    self-service/          # 新增：自注册用户登录后自助页
      user-center/
      inbox/
      workspaces/
```

路由层用路径前缀 `/account/*` 和 `/self/*` 区分，router guard 拉取菜单时按 `current_app_key + space_key` 过滤；子域名部署时前缀可为空（`account.xxx/auth/register` 和 `localhost/account/auth/register` 两种 URL 都指向同一个 Vue 路由组件）。**前端不硬编码 App 归属**：页面加载时调用 `/auth/current-app`（或复用 `/auth/register-context` 的 `entry_app_key` 字段）向后端询问"当前这个请求归哪个 App"，由后端基于 `host+path` 命中 `AppHostBinding` 返回结果。

**登录入口统一**：登录页归 `account-portal/public`，治理管理员和自注册用户共用同一个登录页。登录成功后由后端按用户身上的角色 / 功能包计算出 landing，自注册用户进 `platform-admin/self-service`，管理员进 `platform-admin/default`。同一个 token 在两个 App 之间通用。

**旧路径兼容**：`/auth/login` 与 `/auth/register` 做 307 redirect 到 `/account/auth/login` 与 `/account/auth/register`，避免破坏已分享的链接与书签。

### 9.1 注册页

- 挂在 `/account/auth/register`，归属 `account-portal` App。
- 打开时先调 `GET /auth/register-context`，按 `field_schema` 动态渲染字段，按 `allow_public_register` 决定是否禁用提交。
- 补充 email / nickname / invitation_code 字段（根据 `field_schema` 控制显示）。
- 提交后按响应中的 `landing` 跳转：`auto_login=true` 时跳到 `landing.home_path`（默认 `/self/user-center`，此时已切到 `platform-admin` 的 `self-service` space），`false` 时跳 `/account/auth/login`。
- `frontend/src/api/auth.ts` 新增 `registerContext()` 与 `register()` 封装。

### 9.2 后台管理页

在 `frontend/src/views/system/` 下新增：

1. **注册入口管理**（路由 `/system/register-entries`）：CRUD `register_entries`，挂在"应用管理"菜单下。
2. **注册策略管理**（路由 `/system/register-policies`）：CRUD `register_policies` + 绑定功能包 / 角色，挂在"系统管理"菜单下。
3. **注册记录列表**（路由 `/system/users/register-logs`）：基于 `users` 表审计字段做筛选视图，挂在"用户管理"菜单下。

对应后端接口（第一期纳入）：
- `GET/POST/PUT/DELETE /system/register-entries`
- `GET/POST/PUT/DELETE /system/register-policies`
- `POST /system/register-policies/:code/feature-packages`（绑定）
- `POST /system/register-policies/:code/roles`（绑定）
- `GET /system/users/register-logs`

均需 `x-permission-key` 声明，走现有 permissionseed 流程。

---

## 10. 校验设计

### 10.1 前端（即时反馈）

用户名长度/字符、密码复杂度、确认密码一致、协议勾选、邮箱格式、验证码非空、邀请码非空。

### 10.2 后端（最终裁决）

- 用户名 TrimSpace、长度（3–32）、字符白名单（`[a-zA-Z0-9_.-]`）、保留词（`admin` / `root` / `system` / `null` / …）。
- 邮箱 TrimSpace、RFC 格式、唯一性。
- 密码复杂度：长度 ≥ 8，至少包含字母 + 数字。
- 入口 / 策略 `status = enabled` 校验。
- 用户名 & 邮箱唯一性（已有）。
- Captcha / 邀请码 / 协议版本（按 `effective` 条件校验）。
- IP 频率限制（第三期；第一期预留 `register_ip` 字段）。

OpenAPI spec 作为契约真相，所有长度 / 格式 / 必填写入 schema。

---

## 11. 工作空间关系

- **自注册**：注册时**即刻**创建 `personal` workspace，写入 `workspace_members` owner。不默认创建任何协作空间。
- **邀请注册**（预留）：注册成功后按 `invitation_code_id` 将用户加入目标协作空间，角色由邀请规则决定。
- **商家注册**（未来）：注册后引导创建业务空间，不在注册瞬间直接授治理权限。

> 改动点：`authService.Register` 中需调用 `workspaceService.EnsurePersonalWorkspaceForUser`，而不是延迟到 `auth.me`。

---

## 12. 审计与可观测

第一期：
- `users` 表审计字段齐全。
- 失败注册通过结构化日志（已有 zap）记录：`entry_code`、`policy_code`、`host`、`path`、`ip`、`reason`。

第二期预留：独立 `register_audit_logs` 表（目前不新增，`users` 字段已足够）。

---

## 13. 实施顺序与拆分（第一期）

按以下 slice 推进，每个 slice 保证可编译、可测试：

| Slice | 内容 | 产出 |
|---|---|---|
| **S1 数据模型** | `00006_register_system.sql` + GORM entity + `users` 字段扩展 | 表结构 + `make migrate` 通过 |
| **S2 App Seed** | `account-portal` App 记录 + `AppHostBinding` + `public` MenuSpace + `self-service` MenuSpace + 两边菜单 | 两个 App 在 `apps` 表共存，host/path 路由生效 |
| **S3 注册 Seed** | `self_service.basic` FP / `personal.self_user` Role / `default.self` policy / `default` entry | `permissionseed` 启动注入 |
| **S4 Resolver & Context 接口** | `register/resolver.go` + `register/context.go` + OpenAPI `/auth/register-context` | 接口可调通，能按 `host+path` 命中入口 |
| **S5 Register 改造** | `authService.Register` 接入 `effective` + 即刻创建 personal workspace + 审计字段写入 + landing | `/auth/register` e2e 通过 |
| **S6 前端接入** | 迁移 `views/auth/register` → `views/account-portal/auth/register` + 新增 `views/platform-admin/self-service/*` + router guard 按 `app_key` 过滤菜单 + `register-context` 调用 + landing 跳转 | 前端 build 通过，`/account/auth/register` 可注册、跳 `/self/user-center` |
| **S7 后台管理页（入口/策略/记录）** | 三个页面 + 对应 CRUD 接口 + permissionseed 同步 | 后台可配置 |

第一期完成后，模式 A 即可投产。第二/三期按原设计文档第十六节推进。

---

## 14. 未来演进：新增 `user-portal` / `merchant-console`

第一期 `account-portal` 已独立，后续再拆业务承载 App 时路径如下，**零模型变更**：

1. 新建 `apps` 记录：`app_key = "user-portal"`，配置其 `AppHostBinding`（如 `user.example.com/*` 或 `/user/*`）。
2. 在 `user-portal` 下创建业务 MenuSpace、菜单、功能包、角色。
3. 更新 `register_policies.default.self`：`target_app_key` 从 `platform-admin` 改为 `user-portal`，`target_navigation_space_key` 改为 `user-portal` 的默认 space，`target_home_path` 改为该 App 下的首页。
4. （可选）把原本绑在 `self_service.basic` 上的菜单迁到 `user-portal`，或直接在 `user-portal` 下建新功能包并替换策略绑定。
5. 入口表 `register_entries` **不动**——公开入口仍归 `account-portal`，只是承载侧换了 App。
6. 前端新增 `frontend/src/views/user-portal/` 目录，挂 `/user/*` 路由。

同理可扩展 `merchant-console` 等 App。关键点：**`account-portal` 是永久的公开入口层，业务 App 只管承载**，两者通过 `register_policies.target_app_key` 字段松耦合。

---

## 15. 开放问题（待确认或第二期决定）

1. **`allow_public_register` 的覆盖语义**：第一期确定"入口字段 `*bool` 非 null 即覆盖"。更复杂的策略继承链（例如 App 级默认）留到第二期。
2. **邀请码表结构**：第一期只预留 `users.invitation_code_id` 字段（nullable，暂无外键），完整 `invitation_codes` 表在第三期设计。
3. **Captcha 后端实现**：第一期预留开关字段与校验 hook，具体实现（图形码 / turnstile / hcaptcha）在第三期决定。
4. **注册策略 → 菜单直接绑定**：第一期坚持"功能包 + 角色 + MenuSpace 自然决定菜单"，不新增策略→菜单直绑表。
5. **注册上下文缓存**：`register-context` 接口是否加 etag / 短 TTL 缓存？第一期直接读库；观测 QPS 后再决定。

---

## 16. 文件改动清单（供实施计划引用）

**新增：**
- `backend/internal/pkg/database/migrations/00006_register_system.sql`
- `backend/internal/modules/system/models/register.go`（`RegisterEntry` / `RegisterPolicy` / `RegisterPolicyFeaturePackage` / `RegisterPolicyRole`）
- `backend/internal/modules/system/register/{resolver,context,service,landing,repository}.go`
- `backend/internal/pkg/permissionseed/register_seed.go`（或合入 `seeds.go`）；含 `account-portal` App、两个 MenuSpace、菜单、功能包、角色、策略、入口种子
- `backend/api/openapi/domains/auth/schemas.yaml`（扩展）
- `frontend/src/views/account-portal/auth/register/`（从 `views/auth/register` 迁入并改造）
- `frontend/src/views/account-portal/auth/login/`（从 `views/auth/login` 迁入；原路径 redirect）
- `frontend/src/views/platform-admin/self-service/user-center/`
- `frontend/src/views/platform-admin/self-service/inbox/`
- `frontend/src/views/platform-admin/self-service/workspaces/`
- `frontend/src/views/system/register-entries/`
- `frontend/src/views/system/register-policies/`
- `frontend/src/views/system/users/register-logs.vue`

**修改：**
- `backend/internal/modules/system/models/model.go`（`User` 新增字段）
- `backend/internal/modules/system/auth/service.go`（`Register` 改造）
- `backend/internal/api/handlers/auth.go`（接入新字段 + 新增 `RegisterContext` handler）
- `backend/api/openapi/domains/auth/paths.yaml`（新增 `/auth/register-context`，改造 `/auth/register`）
- `backend/internal/modules/system/workspace/service.go`（`EnsurePersonalWorkspaceForUser` 可复用，无需改签名）
- `frontend/src/views/auth/register/index.vue`
- `frontend/src/api/auth.ts`

---

> 本文档经用户确认后进入 Plan 模式，产出精确到函数签名与字段类型的实施计划。
