## 2026-04-16 backend 子目录文档按目录重构

### 本次改动
- 重写 `backend/internal/modules/README.md`，将上层说明收口为领域实现层总览，只保留 `system/` 与 `observability/` 两个当前主域入口。
- 重写 `backend/internal/modules/system/README.md`，按真实子目录重新列出 `apiendpoint`、`app`、`auth`、`workspace`、`space`、`register`、`social` 等当前有效模块，并补充目录间关系说明。
- 重写 `backend/internal/modules/system/collaborationworkspace/README.md` 与 `backend/internal/modules/system/permission/README.md`，删除旧的作用域叙事，改为“负责什么 / 不负责什么 / 相关目录”的目录级说明。
- 新增 `backend/internal/modules/observability/README.md`，为 `audit/` 和 `telemetry/` 补齐可观测性目录入口文档。
- 重写 `backend/api/openapi/README.md`，保留契约真相源、目录职责和生成链，删除与上层流程文档重复的长篇历史说明。

### 下次方向
- 若后续继续给 `backend/internal/modules/system/` 下的高频目录补 README，优先补 `workspace/`、`space/`、`page/`、`register/`，保持“就近写目录职责”的原则。
- 若 OpenAPI 目录中的 `paths/` 已彻底无消费，可在确认生成链和引用后再清理历史兼容目录，避免 README 长期保留“兼容结构”说明。

## 2026-04-16 文档导航收束与去重

### 本次改动
- 将仓库入口文档重新分层：`docs/INDEX.md` 作为唯一导航中枢，`docs/project-structure.md` 作为唯一详细结构说明，`docs/API_OPENAPI_FIXED_FLOW.md` 继续作为唯一 API 闭环流程说明。
- 将根 `README.md`、`backend/README.md`、`frontend/README.md`、`frontend/src/README.md`、`PROJECT_STRUCTURE.md` 收口为轻量入口文档，删除与主说明文档重复的大段语义内容。
- 重写 `docs/guides/README.md` 与 `docs/guides/add-endpoint.md`，让手册页只保留操作入口和差异化内容，不再重复维护一整套 OpenAPI 规则和流程说明。
- 将 `docs/guides/permission-audit.md` 移入 `docs/archive/permission-audit-report.md`，明确历史审计与长期手册分层，避免主手册继续混入一次性排查报告。

### 下次方向
- 继续审查 `docs/guides/commands.md`、`database.md`、`api-auto-registration.md` 与 `backend/api/openapi/README.md` 的交叉内容，必要时再做第二轮收口。
- 若后续新增长期文档，优先补 `docs/INDEX.md` 和对应主题主文档，避免再次出现“多份入口同时解释同一主题”的分散状态。

## 2026-04-14 App 治理配置结构化编辑

### 本次改动
- 将应用管理抽屉中的三段原始 JSON textarea 改为摘要卡片 + 配置入口，默认走结构化表单，高级 JSON 模式收敛到弹窗内部。
- 新增 `frontend/src/views/system/app/config-editor.ts`，统一处理 capabilities、env_profiles、feature_flags、sensitive_config 的 parse、serialize、summary 和历史兼容转换。
- 完成 CapabilitiesDialog、EnvFlagsDialog、GovernanceDialog，分别覆盖稳定能力字段、环境/Flag 配置、治理引用配置，并复用现有运行时与预检数据链路。
- 完成 `vue-tsc`、前端构建、后端 handlers/system-app 测试，以及 dev 预览登录页与历史 JSON 样例回放验证。

### 下次方向
- 补带登录态的真实页面联调，验证“进入应用管理页 -> 打开抽屉 -> 打开弹窗 -> 保存 -> 重新打开”的完整闭环。
- 单独排查生产预览根路由的压缩态运行时错误，确认是否为本仓库已有问题，并补最小复现。
- 若后续决定收紧 `SystemAppCapabilities` / `SystemMeta` 契约，再按 OpenAPI-first 主线推进 spec、生成链和前后端同步。

## 2026-04-14 编辑应用页精简为接入管理

### 本次改动
- 将“运行能力声明”收敛为“接入安全”，只保留 `capabilities.cors_origins` 与 `capabilities.csp` 两个已被平台真实消费的字段。
- 从编辑抽屉移除“多环境与 Feature Flag”“治理补充”两整块配置入口，避免后台继续承担业务 App 的内部运行配置。
- 删除接入安全弹窗里的多余 Tab 壳，只保留单屏表单与高级 JSON 兜底；保存时继续保留历史 `meta` 数据，不主动清洗旧字段。
- 调整治理总览和检查项文案，统一强调当前页面只负责 App 接入登记，不再暗示可在此配置运行能力、环境差异和内部治理细节。

### 下次方向
- 带登录态走一次真实页面闭环，确认隐藏旧配置后不会影响已有 App 的编辑、保存和回显。
- 若后续确实需要统一治理某类字段，再按“平台真实消费 -> 明确真相源 -> 再开放配置”的顺序逐项恢复，不再一次性堆入后台表单。

## 2026-04-14 旧能力字段审计与收口

### 本次改动
- 审计 `capabilities` 与 `meta` 的真实消费链，确认 `env_profiles`、`feature_flags`、`supports_app_switch`、`supports_dynamic_routes`、`login_strategy`、`is_auth_center`、`cors_origins`、`csp` 仍在运行时或安全中间件使用，因此未删除。
- 新增前端能力清洗逻辑：编辑应用时会自动剥离已判定无消费的旧能力键，如 `routing.entry_mode`、`routing.route_prefix`、`routing.supports_public_runtime`、`runtime.kind`、`runtime.supports_worktab`、`navigation.*`、`integration.supports_broadcast_channel`、`auth.session_mode`。
- 收缩 `frontend/src/views/system/app/config-editor.ts`，移除已不再使用的环境配置、Feature Flag、治理补充 helper，只保留接入安全与旧能力清洗。
- 更新后端默认 App 能力模板与示例 seed，停止继续写入已无消费的旧能力字段。

### 下次方向
- 对现网已有 App 跑一次数据审计，确认数据库里还残留多少旧能力字段，再决定是否补一次性历史清洗。
- 如果后续要删除 `env_profiles` 或 `feature_flags`，必须先迁走 `frontend/src/domains/app-runtime/context.ts` 的消费链，不能只删管理页入口。

## 2026-04-14 旧能力字段历史清洗

### 本次改动
- 新增一次性历史修正迁移 [00017_cleanup_deprecated_app_capabilities.sql](../backend/internal/pkg/database/migrations/00017_cleanup_deprecated_app_capabilities.sql)，清理 `apps.capabilities` 中已确认无消费的旧键。
- 已在本地执行 `go run ./cmd/migrate`，数据库迁移版本已推进到 `17`，现有 3 个 App 的 `deprecated_keys` 已全部清零。
- 清洗后保留的字段仅包括仍在运行时或安全中间件中被真实消费的能力项，例如 `auth.login_strategy`、`auth.is_auth_center`、`runtime.supports_dynamic_routes`、`integration.supports_app_switch`、`cors_origins`、`csp`。

### 下次方向
- 继续审计 `demo-app` 上的 `managed_pages`、`runtime_navigation`、`app_switchable` 这类旧顶层键，确认是否仍有消费链，避免留下第二批历史包袱。
- 如果要进一步清理 `meta.env_profiles`、`meta.feature_flags`、`meta.sensitive_config`，先迁出对应运行时读取逻辑，再补历史修正。

## 2026-04-14 Demo App 旧顶层能力键清理

### 本次改动
- 审计确认 `demo-app` 上的 `managed_pages`、`runtime_navigation`、`app_switchable` 没有真实消费链；命中的 `managed_pages` 仅是导航 Manifest 返回字段，不是 `apps.capabilities` 的读取。
- 更新 [register_seed.go](../backend/internal/pkg/permissionseed/register_seed.go) 的 demo app seed，停止继续写入这 3 个旧顶层键。
- 新增一次性历史修正迁移 [00018_cleanup_legacy_demo_app_capabilities.sql](../backend/internal/pkg/database/migrations/00018_cleanup_legacy_demo_app_capabilities.sql)，从现有 `apps.capabilities` 中移除这 3 个顶层旧键。
- 已在本地执行迁移，数据库版本推进到 `18`，`demo-app` 当前 `capabilities` 只剩 `auth.is_auth_center` 和 `auth.login_strategy`。

### 下次方向
- 如果后续还要继续做减法，下一个目标应是评估 `meta.env_profiles`、`meta.feature_flags`、`meta.sensitive_config` 的真实收益和运行时耦合，而不是再回头扩张管理页表单。

## 2026-04-14 sensitive_config 收口

### 本次改动
- 审计确认 `meta.sensitive_config` 没有任何运行时消费链，仓库中只剩 `system/app/service` 的规范化与测试用例。
- 更新 [service.go](../backend/internal/modules/system/app/service.go) 的治理元数据规范化逻辑，保存应用时显式丢弃 `sensitive_config`，不再接受或保留该字段。
- 删除对应的后端校验分支与冗余 helper，并把测试改成校验 `sensitive_config` 会被忽略而不是继续入库。
- 新增一次性历史修正迁移 [00019_drop_sensitive_config_from_app_meta.sql](../backend/internal/pkg/database/migrations/00019_drop_sensitive_config_from_app_meta.sql)，从现有 `apps.meta` 中移除 `sensitive_config`。
- 已在本地执行迁移，数据库版本推进到 `19`，当前 3 个 App 的 `meta` 都不再包含 `sensitive_config`。

### 下次方向
- 继续审计 `meta.env_profiles` 与 `meta.feature_flags` 的实际运行价值；这两块仍有运行时依赖，不能像 `sensitive_config` 一样直接收口。

## 2026-04-14 shared_cookie Feature Flag 收口

### 本次改动
- 审计确认当前数据库内没有任何 App 继续配置 `meta.feature_flags.shared_cookie`，而前端仅剩 [auth-session.ts](../frontend/src/utils/http/auth-session.ts) 的兜底读取。
- 删除 `shared_cookie` 的 Feature Flag 兜底逻辑，`共享 Cookie` 会话模式现在只由 `auth_mode` / `login_strategy` 表达，不再允许通过通用 Flag 二次覆盖。
- 更新 [service_test.go](../backend/internal/modules/system/app/service_test.go) 的治理元数据测试样例与错误断言，避免继续把 `shared_cookie` 当作推荐 Flag。
- 新增一次性历史修正迁移 [00020_drop_shared_cookie_feature_flag.sql](../backend/internal/pkg/database/migrations/00020_drop_shared_cookie_feature_flag.sql)，从 `apps.meta.feature_flags` 中移除 `shared_cookie`，并在 `feature_flags` 变空时一并删除该对象。
- 已在本地执行迁移，数据库版本推进到 `20`；当前 3 个 App 的 `meta` 均已不再包含 `feature_flags`。

### 下次方向
- 继续区分 `feature_flags` 中“平台临时覆盖开关”和“业务 App 自己的内部 Flag”；若平台确认不再需要临时覆盖 `supports_app_switch` / `supports_dynamic_routes`，再整体下线 `meta.feature_flags`。

## 2026-04-14 feature_flags 整体下线与旧能力兜底复清

### 本次改动
- 审计确认 `meta.feature_flags` 在运行时只剩 `app_switcher` 和 `disable_dynamic_routes` 两个历史覆盖入口，且当前数据库中没有任何真实配置值，因此将应用切换与动态路由判定统一收回到 `capabilities.integration.supports_app_switch`、`capabilities.runtime.supports_dynamic_routes`。
- 删除前端运行时上下文中的通用 `isFeatureEnabledForApp` / `isHttpAppFeatureEnabled` 链路，避免继续把 `feature_flags` 当作平台级运行配置入口。[context.ts](../frontend/src/domains/app-runtime/context.ts) [request-context.ts](../frontend/src/utils/http/request-context.ts)
- 更新 [service.go](../backend/internal/modules/system/app/service.go) 的治理元数据规范化逻辑，保存应用时显式丢弃 `meta.feature_flags`，不再接收该字段；对应测试改为验证 `feature_flags` 会被忽略。
- 新增一次性历史修正迁移 [00021_cleanup_feature_flags_and_stale_app_capabilities.sql](../backend/internal/pkg/database/migrations/00021_cleanup_feature_flags_and_stale_app_capabilities.sql)，统一移除 `meta.feature_flags`，并对旧能力字段再做一次兜底清洗，避免 `platform-admin` 等历史 App 残留旧键。
- 已在本地执行迁移，数据库版本推进到 `21`，当前 3 个 App 的 `meta` 均不再包含 `feature_flags`；`platform-admin` 的 `capabilities` 也已回到精简后的最小集合。

### 下次方向
- 若继续做减法，下一步应评估 `meta.env_profiles` 是否仍值得由基座维护；这块还关联运行时入口覆盖，不能像 `feature_flags` 一样直接整体删除。

## 2026-04-14 env_profiles 下线

### 本次改动
- 审计确认 `meta.env_profiles` 在当前仓库只剩运行时入口 URL 的一层兜底读取，数据库中也已没有任何真实配置值，因此将运行时入口统一收回到顶层 `frontend_entry_url`、`backend_entry_url`、`health_check_url`。
- 删除 [context.ts](../frontend/src/domains/app-runtime/context.ts) 中基于 `env_profiles` 的运行环境推断与入口回退逻辑，避免基座继续承担 App 多环境配置中心职责。
- 更新 [service.go](../backend/internal/modules/system/app/service.go) 的治理元数据规范化逻辑，保存应用时显式丢弃 `meta.env_profiles`；对应测试改为验证 `env_profiles` 会被忽略。
- 新增一次性历史修正迁移 [00022_drop_env_profiles_from_app_meta.sql](../backend/internal/pkg/database/migrations/00022_drop_env_profiles_from_app_meta.sql)，从 `apps.meta` 中整体删除 `env_profiles`。
- 已在本地执行迁移，数据库版本推进到 `22`，当前 3 个 App 的 `meta` 仍全部为空对象。

### 下次方向
- 如果继续做减法，下一步应评估是否还需要在运行时保留 `appMetaMap` 这层容器；目前它已不再承载环境配置和 Feature Flag。

## 2026-04-14 登录后菜单空白修复

### 本次改动
- 复现并确认问题不在后端导航接口：登录成功后 `runtime/navigation` 已返回完整菜单树，但前端仍会出现“已跳转到工作台、左侧菜单为空白”的状态。
- 在 [shared.ts](../frontend/src/domains/auth/flows/shared.ts) 中调整登录完成时序，登录前清理旧会话时不再额外挂起延迟路由重置，登录成功后改为立即同步重置导航运行时。
- 在 [store.ts](../frontend/src/domains/auth/store.ts) 中为 `clearSessionState` 增加可选 `resetRouterDelay`，避免登录链路复用登出清理逻辑时把新的菜单状态再次清掉。
- 在 [reset-handlers.ts](../frontend/src/domains/navigation/runtime/reset-handlers.ts) 中抽出 `resetRouterStateNow()`，把同步重置和延迟重置分开，减少登录初始化与定时清理互相串扰。

### 下次方向
- 再补一轮“切换账号 / 退出后立即重登 / 多标签页回流”的真实浏览器回归，确认没有其他会话广播或延迟任务继续影响菜单初始化。
- 若后续还出现偶发导航空白，优先补登录链路的运行时埋点，记录 `clearSessionState`、`resetRouterState`、菜单加载完成的先后顺序，避免再次靠人工快照定位。

## 2026-04-14 认证模板页面级文案恢复

### 本次改动
- 修正模板语义误判：继续删除全局 `texts`，但恢复 `pages.<scene>.texts` 作为登录页、注册页、找回密码页各自的唯一可编辑文案来源。
- 更新 [index.vue](../frontend/src/views/system/login-page-template/index.vue) 的 `pages` 面板，补回每页标题、副标题、主按钮文案，以及找回密码页的次按钮文案编辑项。
- 更新 [useAuthPageTemplate.ts](../frontend/src/domains/auth/useAuthPageTemplate.ts) 和三张认证页，使运行时只读取当前页面自己的 `texts`，不再依赖全局文案继承。
- 调整 [register_seed.go](../backend/internal/pkg/permissionseed/register_seed.go) 默认模板 seed，重新写入页面级默认文案；同时移除错误清理 `pages.*.texts` 的临时迁移 `00023_cleanup_login_page_template_texts.sql`。

### 下次方向
- 用真实浏览器再做一轮模板编辑联调，确认 `pages` 面板改文案后，右侧登录页 / 注册页 / 找回密码页预览都能即时反映。
- 如果后续还要扩展页面级文案，优先明确字段白名单，再决定是否补占位文案、链接文案等细项，避免重新膨胀成“全局 texts + 页面覆盖”双层模型。

## 2026-04-14 注册策略去掉归属 App 与 Workspace 类型

### 本次改动
- 从注册策略配置链中彻底删除 `app_key` 与 `default_workspace_type`：前端表单不再展示“所属 App Key”，策略列表也不再单独显示归属 App，只保留真正影响注册结果的去向和规则字段。
- 更新 [schemas.yaml](../backend/api/openapi/domains/system-register/schemas.yaml)、[schemas.yaml](../backend/api/openapi/domains/auth/schemas.yaml) 以及对应 handler / model 映射，确保请求、响应、`register-context` 和运行时结构都不再携带这两个字段。
- 新增迁移 [00023_drop_unused_register_policy_columns.sql](../backend/internal/pkg/database/migrations/00023_drop_unused_register_policy_columns.sql)，从 `register_policies` 表实际删除这两列；默认 seed 与设计文档也同步改成新的最小策略结构。
- 已完成 OpenAPI 生成链、前端 `gen:api`、`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit` 与 `go run ./cmd/migrate`，数据库版本已推进到 `23`。

### 下次方向
- 如果继续给注册策略做减法，下一步应评估 `target_home_path` 是否需要常驻表单，还是降级为高级项，避免“注册后去向”继续暴露过多实现细节。
- `register-entry` 与 `register-policy` 目前仍保留模板卡片式入口；如果后续要继续收紧，可以把“预设模板”再收成下拉或轻量向导。

## 2026-04-16 上传系统基础设施与媒体上传链路落地

### 本次改动
- 新增上传系统 5 张核心表与模型：`storage_providers`、`storage_buckets`、`upload_keys`、`upload_key_rules`、`upload_records`，并接入 AutoMigrate 与默认 seed，形成可用的本地上传配置基座。
- 打通媒体上传后端链路：`/media/upload`、`/media`、`/media/{id}` 现在落到真实上传服务，文件写入 `data/uploads`，返回结构化媒体元数据，并通过 `/uploads` 静态路由对外访问。
- 补齐上传配置与密钥基础设施：新增 `upload` 配置段、环境变量映射、默认值与校验；实现 AES-256-GCM 版 `SecretCipher`、版本化密文前缀与解密缓存，为后续 Provider 密钥加密接入预留稳定接口。
- 前端新增统一上传 SDK 与 `useUpload`，WangEditor 不再手拼旧接口，改为走统一媒体上传入口；同时完成 OpenAPI bundle/ogen/前端 `gen:api` 刷新。
- 顺手修复仓库里阻塞全量测试的历史断层：路由桥接对账、权限上下文判定、导航测试桩接口漂移，保证当前分支 `go test ./...` 与前端构建均通过。

### 下次方向
- 如果要继续兑现任务树里的“配置中心”，下一步应补 Provider/Bucket/UploadKey/Rule 的管理 API 与管理端页面，而不是继续把默认配置硬编码在 seed 上。
- 直传、分片、处理管道、多云驱动（OSS/COS/S3）和 E2E/运维文档仍未落地；这些是独立阶段工作，应该按真实范围继续拆分推进，不应在任务树里一次性伪完成。

## 2026-04-16 上传系统阶段 3/4/5 第二轮收口

### 本次改动
- 补齐直传主链：新增 `/media/prepare`、`/media/complete` 契约与 handler，`UploadService` 支持 prepare/direct-complete 两段式流程，前端上传 SDK 改为统一走 `prepare -> direct/relay -> complete`。
- 前端 SDK 收口了字符串简写 key、对象参数、直传进度/取消与 WangEditor 集成；旧 `uploadMedia` 入口仍保留，作为迁移期兼容封装继续存在。
- 为阶段 3 新增自定义 driver 支撑物：最小模板代码、contract harness、自测以及 FTP/SFTP 占位示例文档，并把安全约束写入扩展文档。
- 本轮执行并通过 `go test ./internal/modules/system/upload/...`、`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit`、`pnpm run build`。

### 下次方向
- 阿里云 relay 节点还没达到 `>5MB multipart + abort + 并发上传` 的验收线，不能关；下一步要么补 multipart，要么收窄节点验收标准。
- 旧接口 deprecation 层还差后端旧路径与响应 header，当前只是保留了前端兼容入口；如果要完成该节点，必须把后端兼容层一并补齐。
- 阶段 3 还剩 Driver 指标/中间件、动态开关，以及腾讯云 / S3 兼容 driver；阶段 5 还剩管理端 UI 和 E2E，这些都是真正未完成的尾项。

## 2026-04-16 迁移与种子整理

### 本次改动
- 重构 `backend/cmd/migrate/main.go` 的执行顺序，把 `schema finalizers`、`default seeds`、`runtime sync` 三段拆开，复用统一任务执行器，避免迁移入口继续把结构修正、默认数据和运行时同步混在一起。
- 收敛上传默认 seed：`backend/internal/modules/system/upload/service.go` 改为内置上传场景声明式初始化，保留 `media.default`、`user.avatar`、`doc.attachment`、`editor.inline` 这些真实基线，不再用“示例”注释组织这段逻辑。
- 站点配置默认值改回脚手架语义：`site.name` 统一为 `MaBen Admin`，描述和版权文案去掉电商定制表述，保持与当前项目定位一致。
- 已完成 `go test ./...` 与 `pnpm exec vue-tsc --noEmit`，当前迁移入口和种子整理后编译通过。

### 下次方向
- `00001_permission_seed_baseline.sql` 仍然是历史补丁型基线迁移，后续如果确认线上状态已稳定，可以评估是否把这类“热修复回放”再归档说明，避免继续堆叠历史语义。
- 上传 seed 里 `doc.attachment` 目前还属于可用但未被前端直接消费的基线能力；如果后续脚手架确定不内置文档上传，可继续收缩到更小的默认面。
- 目前 `cmd/migrate` 仍承担部分运行时同步职责；如果后续继续收口，可以再把 OpenAPI/权限快照同步拆成更明确的启动阶段或独立命令。
