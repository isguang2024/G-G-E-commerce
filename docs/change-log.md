## 2026-04-17 参数管理语义收口与作用域解析增强

### 本次改动
- 将 `site-config` 这条能力线在用户可见层统一收口为“参数管理”，同步更新 OpenAPI 摘要、后端权限/菜单/模块种子、系统菜单标题以及管理页文案，不再继续使用“配置中心 / 站点配置”作为主语义。
- 参数管理 API 改为显式 `scope_type / scope_key` 模型：管理端列表与保存改走 `global / app / all` 作用域，运行时解析支持 `context / global / app`，并新增 `GET /site-configs/lookup` 单键解析接口。
- 这次又补了参数级 `fallback_policy`，默认 `inherit`，`strict` 表示当前作用域未命中时不回退到全局默认；前端管理页、OpenAPI、后端模型与服务逻辑都已同步。
- 全局参数的回退策略现在默认折叠隐藏，不再占基础表单和列表列位，需要时可在编辑弹层里展开高级项。
- 后端 `siteconfig` service/repository/cache 已支持按作用域解析与单键读取，前端 API/store/管理页也同步切到新契约，并在页面中明确提示“不同作用域可重复建同 key、全局也是作用域、不同 APP 可复用同 key”。
- 页面右上角的参数管理标签已经去掉，管理页只保留主标题和作用域说明，不再显示额外徽标。
- 系统菜单种子仍保留内建 seed key `SiteConfig` 稳定，避免生成第二套内建菜单记录；相关验证已覆盖 `go test`、`pnpm exec vue-tsc --noEmit` 和 `pnpm build`。

### 下次方向
- 若要让更多后端服务在组装响应时直接消费参数，下一步应从 `siteconfig.Service` 再抽一个只读 `Reader` 窄接口，并补 `GetString / GetBool / GetNumber / GetJSON` 这类 typed helper，避免各业务层手动解析 `MetaJSON`。
- 可继续在参数管理里补 APP 级默认参数模板、参数用途/敏感级别标记、环境级覆盖与批量导入导出，逐步把它从“参数 CRUD”升级成统一参数提供器，而不是继续在各模块硬编码默认值。

## 2026-04-17 docs/tmp 状态标签与上传文档入口收口

### 本次改动
- 为 `docs/tmp/upload-system-plan.md`、`upload-config-v2-gap-analysis.md`、`upload-config-v2-design.md`、`upload-config-v2-release.md` 补充了“文档状态”提示，明确它们是历史规划稿、阶段性分析稿、设计稿和交付稿，不再充当现行真相入口。
- 为 `docs/tmp/task-tree-skill-feedback.md` 补充单次反馈说明，并指向 `.learnings/ERRORS.md`，避免临时反馈文件继续承担长期知识库角色。
- 将 [backend/Truth/upload-system/overview.md](../backend/Truth/upload-system/overview.md) 的关联文档拆成“当前真相”和“历史材料”两组，明确上传域的正式入口与追溯入口。
- 在 [backend/truth_index.md](../backend/truth_index.md) 增加说明，明确 `docs/tmp/` 下的设计稿、Gap Analysis 和交付稿不是后端真相。
- 已再次校验排除 `frontend-platform/` 后的全部仓库文档，当前仍未发现站内坏链。

### 下次方向
- 若继续清理上传域文档，下一步应把 `docs/tmp` 里仍有长期价值的结论提炼进 `backend/Truth/upload-system/overview.md` 或 `ops-guide.md`，而不是让读者继续读完整历史稿。
- 可继续给 `frontend/Truth/*` 和 `backend/Truth/*` 补统一的“适用场景 / 现行状态 / 关联文档”头部模板，让索引页之外的单文档也能自解释。

## 2026-04-17 文档导航坏链修复与归档入口收口

### 本次改动
- 修复 `docs/tmp/INDEX.md`、`docs/tmp/_archived/guides-README.md`、`frontend/README.md`、`frontend/src/README.md`、`backend/Truth/api-openapi-flow.md`、`docs/change-log.md` 中的陈旧链接，全部改为当前真实路径。
- 给 `docs/tmp/_archived/GUIDELINES.md` 与 `docs/tmp/_archived/guides-README.md` 补充“已归档”提示，避免旧入口继续被误当成现行规范。
- 重新梳理 `docs/tmp/INDEX.md` 的导航语义，把旧时代的 `project-framework`、`guides/*`、`API_OPENAPI_FIXED_FLOW` 路径映射到当前 `AGENTS.md`、`README.md`、`backend/Truth/*`、`frontend/Truth/*` 文档体系。
- 已重新校验排除 `frontend-platform/` 后的全部仓库文档，当前未发现剩余站内坏链。

### 下次方向
- 继续收口 `docs/tmp/` 中仍有长期价值的设计稿与交付稿，把稳定结论迁回 `backend/Truth/` 或 `frontend/Truth/`，减少临时目录承担正式导航职责。
- 若后续继续整理文档，可再统一补一轮“文档状态标签”，明确区分现行真相、目录说明、临时稿和归档稿，降低误读成本。

## 2026-04-17 上传配置中心二期交付与审计收口

### 本次改动
- 完成上传配置中心二期主链路交付，补齐 UploadKey / Rule 运行时字段、前端可见 UploadKey 接口、驱动差异化表单、自定义参数 schema 编辑器，以及前端上传 SDK 的可见目标解析与计划回传。
- 新增 [upload-config-v2-release.md](./tmp/upload-config-v2-release.md) 与 [upload-config-v2-ops-guide.md](../backend/Truth/upload-system/ops-guide.md)，把上线步骤、迁移兼容、回滚策略、安全审计结论、配置示例与运维排障集中收口。
- 更新 [backend/Truth/upload-system/overview.md](../backend/Truth/upload-system/overview.md) 文档导航，便于后续从设计、约束、交付、运维四个层次定位资料。
- 对权限与泄露面做了一轮代码链路审计，确认管理面密钥回显走脱敏值、Provider 写入走加密、前端可见 UploadKey 接口只暴露安全摘要字段。

### 下次方向
- 若要把上传能力开放给更广的业务前台，优先拆分 `system.media.manage`，把“上传使用”和“媒体管理”拆成不同权限，而不是继续用后台管理权限兜底。
- 为 `COS / S3` 补 driver registry、默认 extra 模板与管理页说明卡片，沿用当前分层 UI 和 schema registry，不再复制新页面架构。
- 若后续需要按 `extra_schema` 真正存储 UploadKey / Rule 自定义业务值，再补独立 `extra` 存值槽，避免把 schema 与 value 混在一起。

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

## 2026-04-17 字典描述与备注补齐

### 本次改动
- 为字典项新增 `description` 字段，补齐后端模型、OpenAPI schema、ogen 生成物与前端 `gen:api`，字典项现在可以正式保存备注并通过接口返回。
- 字典管理页补上展示与编辑链路：字典类型列表展示类型描述，字典项弹窗新增“备注”，字典项表格新增“备注”列；`DictSelect` 下拉项也会展示字典项说明。
- 内置字典种子补充了类型描述和各字典项说明，`register_source` 同步带上默认项语义；新库执行迁移后会直接拿到完整说明数据。
- 新增 `00033_dictionary_item_descriptions.sql`，为现有库补 `dict_items.description` 列，并把已有内置字典/字典项的描述回填到当前数据。

### 下次方向
- 如果要进一步利用这些说明，下一步可以在更多字典选择场景里增加 tooltip/帮助文案，而不只是下拉二级说明。
- `EnsureBuiltinDicts` 目前对已存在内置字典仍以“创建为主、更新靠迁移”为主；如果后续内置字典还会频繁演进，可以考虑把它升级为更稳妥的内置字典同步器。

## 2026-04-17 字典项单项保存与删除保护收口

### 本次改动
- 字典项管理不再依赖“整表保存”才能落库：新增单项创建、更新、删除接口，前端改为编辑即保存、停用/启用即保存，避免字典项数量变大后整表回写带来的性能和并发风险。
- 删除链路改成“先停用，后删除”，并在前端停用/删除动作上都加了二次确认；后端同步增加校验，未停用的字典项禁止删除。
- 为 `dict_items` 新增 `is_builtin` 标记与 `00034_dictionary_item_builtin_flags.sql` 迁移，内置字典项现在禁止删除，且编辑时禁止修改 `value`；已有内置项会被迁移脚本回填为内置。
- 已完成 OpenAPI bundle、ogen、`pnpm run gen:api`、`go test ./internal/api/handlers -count=1`、`pnpm exec vue-tsc --noEmit` 与 `go run ./cmd/migrate`，数据库版本已推进到 `34`。

### 下次方向
- 当前删除保护已经覆盖“内置不可删”和“未停用不可删”，下一步如果继续收口，应补“被业务数据引用的自定义字典项禁止删除”，避免删掉历史值后影响用户列表、注册来源等展示。
- 字典类型删除目前仍是整型级删除确认；如果后续这个模块会交给更多运营角色使用，建议把“类型停用”和“类型删除”也拆开，并补影响范围提示。

## 2026-04-17 字典页主从布局重构

### 本次改动
- 字典管理页改成“主区类型表格 + 固定宽度详情侧栏”的 master-detail 结构，类型搜索、筛选、服务端分页统一收口到主区，不再让右侧明细长期占据大面积空白。
- 右侧详情栏改为围绕当前选中类型展示配置摘要和字典项列表，字典项支持局部搜索、状态筛选、前端分页以及原有的新增、编辑、启停、删除动作。
- 保留原有类型和字典项 CRUD 接口，当前重构只调整交互组织方式，没有改后端契约；已执行 `pnpm exec vue-tsc --noEmit` 与 `pnpm run build` 验证通过。

### 下次方向
- 如果后续字典项数量继续增大，下一步应补后端字典项分页和筛选接口，把右侧的前端分页替换成真实服务端分页。
- 右侧详情栏现在更适合承载“最近修改时间、引用位置、使用中的表单/页面”等上下文信息，后续可以继续补运营诊断信息，而不是再把页面做回双大栏表格。

## 2026-04-17 字典页单层表格收敛

### 本次改动
- 将上一版“主表格 + 右侧详情栏”继续收敛为单层全宽表格，名称、编码、描述、状态、字典项数量和更新时间拆成独立列，避免信息继续堆在同一个单元格里。
- 详情查看与字典项维护改为表格展开行内完成，保留字典项搜索、筛选、分页和增删改状态切换能力，同时移除未再使用的旧双栏面板组件。
- 已再次执行 `pnpm exec vue-tsc --noEmit` 与 `pnpm run build`，当前前端联编与构建通过。

### 下次方向
- 如果字典项后续会超过当前展开行可舒适承载的数量，下一步应把字典项区域替换为服务端分页表格，而不是继续增加前端卡片数量。
- 若需要更强运营视角，可以在展开行里补“引用位置 / 最近修改人 / 使用页面”等信息，但不建议再恢复单独右侧大详情区。

## 2026-04-17 字典页双面板与 splitter 对齐

### 本次改动
- 字典页重新收口为“左侧类型面板 + 右侧字典项面板”的双栏布局，靠近当前组件库中的管理控制台样式，而不是继续使用单层展开表格。
- 两侧分页统一替换为项目内的 `WorkspacePagination`，使页码、尺寸切换和圆角样式与现有组件库保持一致。
- 在两个面板之间新增可拖动 splitter，桌面端可左右拖动调整左侧面板宽度，双击可恢复默认宽度；窄屏下自动回退为上下堆叠布局。
- 已执行 `pnpm exec vue-tsc --noEmit` 与 `pnpm run build`，前端联编和构建通过。

### 下次方向
- 如果要继续贴近截图中的控制台风格，下一步可以把左右面板的查询条、工具条和表头间距继续做统一 tokens 收敛，而不是单页局部微调。
- splitter 当前是页面内自实现；如果后续还有别的页面需要同类交互，建议抽成通用 `SplitPane` 组件，避免后面再各页重复实现。

## 2026-04-17 frontend-platform Vben 安装链修复

### 本次改动
- 对齐 `frontend-platform/.npmrc`，移除 `node-linker=hoisted`，补充 `hoist-workspace-packages=true` 以及 `@vben/*`、`@vben-core/*` 的 hoist 规则，避免 pnpm 10 下 Vben workspace 包在安装后不可见。
- 新增 `frontend-platform/scripts/setup-workspace-links.mjs`，在 `postinstall` 阶段自动把 `@vben` / `@vben-core` 工作区包挂到根 `node_modules`，保证 `@vben/tsconfig`、`@vben/vite-config` 等内部包能在 stub 构建前被解析到。
- `frontend-platform/package.json` 的 `postinstall` 改为先执行 workspace links setup，再执行 `pnpm -r run stub --if-present`；已验证 `corepack pnpm install`、`pnpm -F @vben/web-ele run typecheck`、`pnpm build:ele` 通过。

### 下次方向
- 最好在后续有空时用全新 `node_modules` 再走一次首次安装验证，确认从零 clone 的场景也能一次成功，而不是只在当前修复后的环境里通过。
- 如果后面还要启用 `frontend-platform` 下的其它 Vben app，可以直接复用这套自动链接脚本，不要再手工创建 junction。

## 2026-04-17 frontend-platform runtime navigation 菜单适配

### 本次改动
- 新增共享包 `frontend-platform/packages/runtime-navigation`，把后端 `/runtime/navigation` 的 `menu_tree + managed_pages` 统一转换成 Vben backend mode 可消费的动态路由，并补了查询参数和 `/index.vue` 组件别名辅助，给后续 `web-naive` / 新 APP 复用。
- `apps/web-ele` 已改为直接请求 `/runtime/navigation`，登录、`/auth/me`、refresh token 也切到当前后端真实契约；菜单拉取、权限码写入和页面组件解析不再依赖 Vben demo 的 `/menu/all`、`/user/info`、`/auth/codes`。
- `web-ele` 偏好配置切到 `backend` 模式并启用 refresh token，开发代理默认指向 `http://localhost:8080/api/v1`；已验证 `pnpm -F @vben/web-ele run typecheck`、`pnpm exec vitest run packages/runtime-navigation/src/__tests__/build-routes.test.ts --dom`、`pnpm -F @vben/web-ele run build` 通过。
- `pnpm build:ele` 仍会被 monorepo 里若干现存 UI-kit 打包依赖问题拦住，例如缺少 `vue-tsc` / `@tsdown/css` 链接以及 `menu-ui` 现有样式构建错误，这一轮未改 Vben 内核，因此没有顺手修这些基础包问题。

### 下次方向
- 下一步最值得做的是把 `runtime-navigation` 的 pageMap/layoutMap 接线抽成更明确的 app factory，让 `web-naive` 接入时只需要提供自己的页面资源映射，不再重复写 `access.ts` 粘合层。
- 如果要继续推进“菜单配置中心”迁移，应基于这次保留下来的扩展 meta，补运行时 badge / 内页 / 全屏页的 UI 展示校验，而不是再退回到纯 Vben 菜单字段模型。
- 当前 `pnpm build:ele` 的 turbo 链失败是 monorepo 依赖链接和现有 UI-kit 构建问题；后续如果要把 `frontend-platform` 当正式主线，需要单独清理这批基础包构建稳定性。 

## 2026-04-17 frontend-platform 登录主链迁移

### 本次改动
- `apps/web-ele` 的登录页已移除 Vben demo 账号选择和滑块校验，改为直接提交当前后端 `/auth/login` 所需的用户名密码与 `target_app_key`、`redirect_uri`、`state`、`nonce` 等 centralized-login 参数，并补上 `/auth/login-page-context` 文案上下文读取。
- 登录状态管理改为识别后端 `LoginResponse` 里的 `landing` 和 `callback` 语义：直接登录时会优先消费 `redirect` / `landing.home_path` / `target_path`，命中 `callback.redirect_to` 时会先缓存 `state/nonce` 上下文，再跳转中心登录回调。
- `web-ele` 新增 `/auth/callback` 页面和本地 centralized attempt 存储，回跳后可通过 `/auth/callback/exchange` 换取 token 并继续走现有的 `auth/me + 动态菜单` 链路，不再停留在 Vben demo 的“只认 access_token”模型。
- `web-ele` 的注册页已接入 `/auth/register-context` 和 `/auth/register`：会按后端策略动态显示邮箱、邀请码、验证码字段，提交后根据 `access_token` / `pending` 分流到自动登录或回登录页，并保留 `landing_url` / `landing_path` 这类后续跳转信息。
- 忘记密码页已从空提交表单改成明确的说明页：继续使用登录页模板上下文，但不再伪装成“可发送重置链接”的能力，而是直接告知当前后端尚未暴露重置密码 API，并引导返回登录。
- 已执行 `corepack pnpm -F @vben/web-ele run typecheck` 与 `corepack pnpm -F @vben/web-ele run build`，类型检查和生产构建通过。

### 下次方向
- 忘记密码页仍停留在基础壳子，而且后端当前只暴露了登录页模板上下文，没有看到对应提交流程；下一步应明确它是“仅展示入口”还是补真正的重置密码 API，避免继续保留空提交。
- 如果后续还要支持社交登录或 callback 失败后的更细 landing/space 恢复，建议把 `centralized-login` 这层 app 内适配继续抽成共享 `@gge/shared-access`，避免 `web-naive` 再重复一遍。

## 2026-04-17 frontend-platform runtime-access 抽取

### 本次改动
- 新增共享包 `frontend-platform/packages/runtime-access`，提供 `createAccessGenerator()` 和 `createRouterGuard()`，把 Vben app 的“动态菜单生成 + 通用进度条守卫 + 权限守卫”抽成可复用纯逻辑，不再把这套代码散写在每个 app 的 `router/access.ts`、`router/guard.ts` 里。
- `apps/web-ele` 已切到新包：`src/router/access.ts` 现在只负责注入菜单接口、toast、页面映射和布局映射，`src/router/guard.ts` 只负责注入 Vben store、登录路径、默认首页和 `authStore.fetchUserInfo()`。
- 这层抽取仍然保持 app 边界清晰：登录 store、菜单接口、`ElMessage`、`preferences` 等壳和 UI 单例没有下沉进共享包，后续 `web-naive` 只需改自己的注入，不用重写守卫主逻辑。

### 下次方向
- 当前 workspace 仍被安装链环境问题卡住，`vue-tsc` 缺 `@volar/typescript`、`vite` 缺 `rolldown`；下一步如果要正式验证 `runtime-access`，得先把 `frontend-platform` 的 `node_modules` 和 stub 构建恢复完整。
- 再往下值得做的是 Step 3，把 workspace/app 上下文收口成 `@gge/runtime-context`，这样登录、菜单、space 跳转就能继续去掉 `web-ele` 里的零散运行时参数处理。

## 2026-04-17 frontend-platform 安装链修复与正式验证

### 本次改动
- 定位到 `frontend-platform` 安装链的真实根因不是业务代码，而是当前 pnpm 生效配置里有 `symlink=false`，导致包虽然已经下载进 `node_modules/.pnpm`，但 `node_modules` 链接层没有建出来，最终引发 `tsdown` 找不到 `ansis`、`vue-tsc` 找不到 `@volar/typescript`、`vite` 找不到 `rolldown`。
- 在 [frontend-platform/.npmrc](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-platform/.npmrc) 明确补上 `symlink=true`，然后删除损坏的 `node_modules` 并重新执行 `corepack pnpm install`，`postinstall` 的 workspace link 和 `pnpm -r run stub --if-present` 已全部正常跑完。
- 同步把 `runtime-access` 收口到不直接依赖 `vue-router` 类型，改成最小 Router 抽象接口，避免源码直出 workspace 包在当前 monorepo 类型解析下再次被具体路由实现卡住。
- 已重新执行 `corepack pnpm install`、`corepack pnpm -F @vben/web-ele run typecheck`、`corepack pnpm -F @vben/web-ele run build`，三条链路均通过。

### 下次方向
- 当前 install 虽已恢复，但 `internal/vite-config` 仍有一组基于 Vite 8 的 peer warning（`vite-plugin-pwa`、`vite-plugin-vue-devtools` 及其下游插件）；这不影响当前安装和构建，但如果后续要把 `frontend-platform` 当长期主线，建议单独收口这组版本兼容性。
- 认证和菜单基础链路、共享请求层、共享 access 层现在都已落地，下一步最值得做的是继续抽 `@gge/runtime-context`，把 app/space/runtime 参数从 `web-ele` 再剥一层出去。

## 2026-04-17 frontend-platform Vite 版本兼容性清理

### 本次改动
- 单独清理了 `frontend-platform` 里这组 Vite peer warning：定位后确认问题不在本地安装链，而是当前 catalog 把 `vite` 统一升到了 `8.0.8`，但 `vite-plugin-pwa@1.2.0`、`vite-plugin-inspect@11.3.3`、`vite-dev-rpc@1.1.0` 这条链的官方 peer 仍停留在 `vite 7`。
- 将 [frontend-platform/pnpm-workspace.yaml](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-platform/pnpm-workspace.yaml) 的 catalog `vite` 从 `^8.0.8` 调整为 `^7.1.12`，其余 `@vitejs/plugin-vue@6.0.6`、`@vitejs/plugin-vue-jsx@5.1.5`、`vite-plugin-vue-devtools@8.1.1` 保持不变，因为它们本身已支持 `vite 7`。
- 重新执行 `corepack pnpm install` 后，原先的 Vite peer warning 已消失，只剩 4 个与本次无关的 deprecated subdependency 提示；同时复验 `corepack pnpm -F @vben/web-ele run typecheck` 与 `corepack pnpm -F @vben/web-ele run build` 均通过。

### 下次方向
- 这次是按当前插件生态选择更稳的 `vite 7`，不是永久锁死；后续如果 `vite-plugin-pwa` / `vite-plugin-inspect` 官方补上 `vite 8` peer，再统一升回去会更干净。
- 如果要继续把 `frontend-platform` 当主线，下一轮值得清的是 install 输出里剩余那 4 个 deprecated subdependency，避免后面继续混入真正需要处理的警告。

## 2026-04-17 frontend-platform deprecated subdependency warning 清理

### 本次改动
- 继续追踪 install 输出里剩余的 4 个 deprecated subdependency，确认它们都来自 `@vben/vite-config` 里默认关闭但仍被静态安装的可选能力：`importmap` 依赖的 `@jspm/generator` / `cheerio`，以及 `PWA` / `Vite Devtools` 依赖链。
- 调整 [frontend-platform/internal/vite-config/package.json](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-platform/internal/vite-config/package.json)，移除 `@jspm/generator`、`cheerio`、`vite-plugin-pwa`、`vite-plugin-vue-devtools` 这 4 个静态依赖；同时改造 [plugins/index.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-platform/internal/vite-config/src/plugins/index.ts) 与 [plugins/importmap.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-platform/internal/vite-config/src/plugins/importmap.ts)，在真正启用对应功能时才动态加载可选包，并在缺包时抛出明确错误。
- 移除了 `vite-plugin-pwa` 的类型硬依赖，改为本地最小 `PwaPluginOptions` 结构，相关收口在 [typing.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-platform/internal/vite-config/src/typing.ts) 与 [options.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-platform/internal/vite-config/src/options.ts)。
- 重新执行 `corepack pnpm install` 后，这 4 个 deprecated warning 已全部消失；同时复验 `corepack pnpm -F @vben/web-ele run typecheck` 与 `corepack pnpm -F @vben/web-ele run build` 通过。

### 下次方向
- 当前默认路径已经不再把这些可选插件装进工作区；如果后续某个 app 真的要开启 `PWA`、`ImportMap CDN` 或 `Vite Devtools`，需要显式补装对应依赖，否则会收到这次新增的明确提示。
- 这轮收口的是默认安装体验，不是删除能力；后面如果要把这些可选插件做成更标准的 app 级依赖，可以继续往 `apps/*` 侧下沉，而不是再挂回 `@vben/vite-config` 的基础依赖。 

## 2026-04-17 workspace / collaboration / menu_space 边界修正

### 本次改动
- 修正 [docs/project-framework.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/project-framework.md) 的核心语义：删除 tenant 叙事，改成以 `workspace` 为唯一空间主实体，`personal` / `collaboration` 只作为类型分流。
- 修正 [backend/internal/modules/system/README.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/README.md) 与 [backend/internal/modules/system/collaborationworkspace/README.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/collaborationworkspace/README.md) 的目录职责说明，明确 `workspace/`、`collaborationworkspace/`、`space/` 的当前边界。

### 下次方向
- 如果后续继续清理命名，优先把 `space_key` 的文档口径统一成 `menu_space_key`，避免“空间”中文语义和导航域混淆。
- 若再扩展工作区模型说明，优先补 `workspace/` 与 `permission/` 的联动说明，避免把协作空间业务实体和鉴权主域再混写。

## 2026-04-17 workspace 单主域收口后端兜底修复

### 本次改动
- 修复了 [backend/internal/pkg/database/database.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/database/database.go)、[backend/internal/pkg/permissionseed/ensure.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/permissionseed/ensure.go)、[backend/internal/pkg/permissionseed/register_seed.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/permissionseed/register_seed.go) 与 [backend/cmd/migrate/main.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/migrate/main.go) 中对已删除 `collaboration_*` 旧表/旧列的索引、seed、迁移收尾逻辑，避免 `00035` 执行后继续命中旧 schema。
- 修复 [backend/internal/modules/system/user/repository.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/repository.go)、[backend/internal/modules/system/user/service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/service.go)、[backend/internal/modules/system/role/service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/role/service.go) 的角色链路：协作空间上下文只走 `workspace_role_bindings`，不再错误写回全局 `user_roles`。
- 修复 [backend/internal/modules/system/system/message_service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/system/message_service.go)、[backend/internal/modules/system/featurepackage/assign_service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/featurepackage/assign_service.go)、[backend/internal/pkg/workspacefeaturebinding/service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/workspacefeaturebinding/service.go) 等协作空间运行时查询，统一改为基于 `workspaces / workspace_members / workspace_role_bindings / workspace_feature_packages`。
- 已验证 `go build ./...`、`go test ./internal/api/handlers -count=1`、`go test ./internal/api/router -count=1` 通过，后端恢复到可编译、可基本回归状态。

### 下次方向
- 继续完成 `frontend/` 的 S7 收口，尤其是 `api/workspace.ts`、`workspace store`、请求头 `X-Collaboration-Workspace-Id` 与协作空间页面封装，避免前后端仍处于半双轨状态。
- 若要继续做更彻底的领域收敛，下一轮应单独规划消息域里的 `owner/target/recipient_collaboration_workspace_id` 兼容字段是否改名；这次先保证运行时不再依赖被删除的 legacy schema。

## 2026-04-17 菜单语义 menu_space / menu_space_key 收紧

### 本次改动
- 统一收紧菜单领域命名：菜单空间主语义改为 `menu_space`，相关字段统一改为 `menu_space_key`、`menu_space_keys`、`default_menu_space_key`，避免再和业务主域 `workspace space` 混用。
- 后端已同步修改模型、DTO、handler、service、运行时缓存、默认 seed 与 OpenAPI 真相源，并新增迁移 [backend/internal/pkg/database/migrations/00035_menu_space_key_cleanup.sql](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/database/migrations/00035_menu_space_key_cleanup.sql) 完成库表字段收口。
- 已按 OpenAPI-first 链路刷新生成物：重跑 `bundle`、`lint`、`ogen`、`gen-permissions` 与前端 `pnpm run gen:api`，前后端均切到新契约；前端读取侧保留旧字段 fallback 兼容，`navigation_space_key` 这一条 auth 协议线本轮未动。
- 已验证 `go test ./internal/api/handlers -count=1`、`go test ./internal/api/router -count=1`、`go test ./internal/modules/system/app ./internal/modules/system/space ./internal/modules/system/page ./internal/modules/system/navigation -count=1`、`pnpm run gen:api`、`pnpm build` 通过。

### 下次方向
- 继续清理文档与零散运行时里的旧称呼，把剩余“菜单叫 space”的表述统一收口，避免新代码再引入歧义。
- 如果后续要彻底去掉兼容分支，可以在一轮单独回收中移除前端对 `space_key` / `space_keys` / `default_space_key` 的 fallback，并同步做一次真实库数据巡检。

## 2026-04-17 菜单空间兼容兜底彻底回收

### 本次改动
- 继续把前端菜单域、页面域、系统治理页里的旧字段兼容彻底删掉：`spaceKey`、`spaceKeys`、`defaultSpaceKey` 不再作为菜单空间读写入口，统一只认 `menuSpaceKey`、`menuSpaceKeys`、`defaultMenuSpaceKey`。
- 收口了运行时菜单空间 store、系统 App/菜单空间/页面管理页、菜单弹窗和路由跳转链路；路由 meta、菜单树、页面暴露和 Host 绑定都已只走新命名，不再保留旧字段 fallback。
- 后端同时补掉了剩余注释和上下文口径里的旧 `space_key` 表述，菜单上下文统一回到 `menu_space_key`；认证协议里的 `navigation_space_key` 保持不动，因为它属于登录落点协议，不是菜单领域旧名兼容。
- 已复验 `pnpm exec vue-tsc --noEmit`、`pnpm build`、`go test ./internal/api/handlers ./internal/api/router ./internal/modules/system/navigation ./internal/modules/system/space ./internal/modules/system/page ./internal/modules/system/app -count=1` 全部通过。

### 下次方向
- 如果要继续深挖，可再做一轮纯重命名，把后端 service/repository 内部局部变量里的 `spaceKey` 也系统改为 `menuSpaceKey`，进一步减少阅读歧义；这轮先优先回收兼容分支与对外字段。
- 认证、注册、回跳链路里的 `navigation_space_key` 目前仍是独立协议语义；若后续也要统一命名，需要单独评估登录态回跳和第三方注册落点的兼容成本。

## 2026-04-17 workspace / collaboration 命名与角色建模分析补充

### 本次改动
- 补做了一轮前后端、OpenAPI、鉴权与权限链路的结构分析，确认 `collaboration_workspace` 不是一个可以全局替换的名字：主实体、主表、主 CRUD 应统一回 `workspace`，只有“当前协作态 / 协作边界 / 协作成员”这类运行时语义才应保留为 `collaboration`，历史桥接字段和旧表应视为待删除兼容层。
- 角色模型也一起核对了当前实现，现状是“单 `roles` 表 + `collaboration_workspace_id` 作用域列 + `user_roles` / `workspace_role_bindings` 双轨并存”。其中 [backend/internal/modules/system/models/model.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/models/model.go:99) 仍让 `Role`、`UserRole` 直接带 `collaboration_workspace_id`，而 [backend/internal/modules/system/models/workspace.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/models/workspace.go:56) 已经有更接近目标态的 `workspace_role_bindings`。
- 结论上不建议再新造一张“协作角色主表”。更合适的目标是保留单一 `roles` 定义表：全局角色就是无空间作用域的角色，协作角色就是绑定到某个 `collaboration` 类型 `workspace` 的角色；用户在空间内拿到角色，统一走 `workspace_role_bindings` 这类关联表，而不是继续把作用域塞进 `user_roles.collaboration_workspace_id`。
- 如果只考虑当前两种作用域，最省改动的落地是把 `roles.collaboration_workspace_id` 收口成 `roles.workspace_id`；如果考虑未来扩展性更强，也可以进一步拆成“`roles` 定义表 + `role_scopes/workspace_role_scopes` 作用域表 + `workspace_role_bindings` 赋权表”，但无论哪条都不建议再分裂出第二套 `collaboration_roles` 主表。

### 下次方向
- 真正动手时先做 migration 方案，不要先做代码 rename：优先决定 `roles.collaboration_workspace_id` 是直接改 `workspace_id`，还是抽成独立 scope 表；同时清退 `user_roles.collaboration_workspace_id`、`workspaces.collaboration_workspace_id`、JWT `collaboration_workspace_id` 与 `X-Collaboration-Workspace-Id`。
- OpenAPI 需要按两类重排：协作空间实体治理回 `workspace/workspaces`，当前协作态接口收口到 `collaboration/current/*`；对应前端再把 `collaboration-workspace.ts` / `collaboration-workspace` store 拆回 `workspace.ts` 与 `collaboration.ts` 两条主线。
- 权限真相和实现也要一并对齐，尤其是 [backend/internal/pkg/permission/evaluator/evaluator.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/permission/evaluator/evaluator.go:197) 这条仍然依赖 `workspaces.collaboration_workspace_id + user_roles.collaboration_workspace_id` 的旧链路，否则即便表名改完，运行时依旧会被旧模型锁死。

## 2026-04-17 workspace/collaboration 单主域与 role/message scope 泛化落地

### 本次改动
- 新增迁移 [backend/internal/pkg/database/migrations/00039_workspace_collaboration_scope_unification.sql](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/database/migrations/00039_workspace_collaboration_scope_unification.sql)，为 `roles` 补 `scope_type/scope_id`、新建 `role_scopes`，并把 message template / dispatch / delivery / recipient group target 的 scope 字段扩到 `global / personal / collaboration` 兼容模型，同时完成旧 collaboration 数据回填。
- 后端运行时已切主链：鉴权 middleware 与 JWT 改走 `auth_workspace_id + auth_workspace_type`，权限评估和平台/个人角色读取改为 `workspace_role_bindings + role_scopes`，消息服务补上 `global` 语义，非协作上下文默认以全局模板、全局发送人、全局消息记录运行。
- OpenAPI 与前端协议已同步收口：`/collaboration-workspaces*` 根路径改为 `/workspaces/collaboration*` 与 `/collaboration/current*` 两类，消息 schema 新增 `owner_scope_id / target_scope_type / recipient_scope_type` 等字段；前端请求头已删除 `X-Collaboration-Workspace-Id` 注入，消息模块系统侧统一改成 `global` scope。
- 已按生成链刷新 `backend/api/gen/`、`frontend/src/api/v5/schema.d.ts`、权限 seed 与前端错误码；并复验 `go test ./internal/api/handlers -count=1`、`go test ./internal/api/router -count=1`、`pnpm run gen:api`、`pnpm exec vue-tsc --noEmit`、`pnpm build` 通过。
- 继续收口后端低风险 legacy：`backend/internal/modules/system/user/repository.go` 中用户列表角色过滤、全局角色 fallback 加载与多处“无协作空间即全局角色”兜底查询，已统一切到 `role_scopes` 规则；`backend/internal/pkg/permissionseed/ensure.go` 与 `backend/internal/pkg/permissionseed/register_seed.go` 里的默认角色/自助角色查找也不再依赖 `roles.collaboration_workspace_id IS NULL`。
- 额外补跑了 `go test ./internal/pkg/permissionseed -count=1`，确认新的全局角色判定不会把 seed/ensure 链路打断；同时把本轮已完成节点和剩余风险回写到任务树 `tsk_01KPDZT17355P7TTRRKBEM`，避免后续继续推进时状态漂移。
- 本轮继续收口消息域 runtime：`backend/internal/modules/system/system/message_service.go` 中发送人、接收组、发送记录等协作上下文的 `scope_id` 已统一改用 workspace ID，不再把 `collaborationWorkspaceID` 直接写进新 scope 字段；模板/发送记录/投递明细/接收组目标项的空间名称解析也改成从 `workspaces` 读取。
- 协作成员类查询已优先切到 `workspace_members + workspaces`：包括消息投递对象、协作空间指定成员、派发用户列表，以及基于内建协作角色/功能包的 identity recipient 解析；并新增接收组目标保存时的 `target_scope_type/target_scope_id` 写入，避免新数据继续落在半新半旧状态。
- 已补跑 `go test ./internal/modules/system/system -count=1`、`go test ./internal/api/handlers -count=1`、`go test ./internal/api/router -count=1` 通过。当前消息域仍保留少量 legacy fallback：`user_roles.collaboration_workspace_id`、快照表上的 `collaboration_workspace_id`、以及若干对旧字段的展示兼容，这部分留到下一轮继续回收。

### 下次方向
- 当前仍保留一批 legacy `collaboration_workspace_id` 兼容字段和消息域内部查询，下一轮应继续清理 `message_service.go`、`user/repository.go` 中剩余的旧表/旧列依赖，再做真正的 schema 回收。
- 这轮没有同步重命名旧 permission key，`collaboration_workspace.manage` 等权限点仍在兼容运行；若要彻底完成命名收口，下一步应单独规划 permission key、seed、菜单页面 key 与前端路由命名的统一迁移。

## 2026-04-18 workspace 边界 canonical 第二阶段

### 本次改动
- 新增迁移 [backend/internal/pkg/database/migrations/00040_workspace_boundary_canonicalization.sql](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/database/migrations/00040_workspace_boundary_canonicalization.sql)，补齐 `workspace_blocked_menus`、`workspace_blocked_actions`、`workspace_access_snapshots`、`workspace_role_access_snapshots`，并回填 `workspace_members`、`workspace_feature_packages`、`workspace_role_bindings` 等 canonical 数据，开始把协作边界运行时从旧 `collaboration_workspace_*` 表链路脱开。
- 后端 runtime 已切一批主路径到 `workspace_id` 真相：`collaborationworkspaceboundary`、`appscope`、`permissionrefresh`、`workspacefeaturebinding`、`platformaccess`、`cmd/init-admin`、菜单清理与协作空间删除流程，现已优先读写 `workspace_*` 表，不再把旧协作阻断表和旧协作快照表当主存储。
- 消息域继续回收 legacy：`backend/internal/modules/system/system/message_service.go` 中平台角色收件人与协作角色/功能包收件人的主查询，已移除对 `user_roles.collaboration_workspace_id` 和 `collaboration_workspace_role_access_snapshots` 的依赖，改为 `workspace_role_bindings` 与 `workspace_role_access_snapshots`。
- 协议层也同步收口了一层：旧 JWT/APIKey middleware 现在会补齐 `auth_workspace_id/auth_workspace_type`，`router` 里的类型化 logger 改记 canonical workspace，上游 CORS 允许头移除了 `X-Collaboration-Workspace-Id`。已验证 `go test ./internal/pkg/collaborationworkspaceboundary ./internal/pkg/appscope ./internal/pkg/permissionrefresh ./internal/pkg/workspacefeaturebinding ./internal/modules/system/user ./internal/modules/system/menu -count=1`、`go test ./internal/pkg/platformaccess ./internal/modules/system/menu -count=1`、`go test ./internal/modules/system/system ./internal/api/handlers ./internal/api/router -count=1` 通过。

### 下次方向
- 继续清理 `backend/internal/api/middleware/app_context.go`、`page/service.go`、`page/runtime_cache.go`、`user/repository.go`、`cmd/migrate/main.go` 里剩余的 `collaboration_workspace_id` 查询与旧 consolidate 逻辑，把后端 runtime 里最后一批 legacy 判定链路收掉。
- 后端真相层稳定后，再整批推进 OpenAPI / permission key / 前端目录与路由命名重排；当前 `CollaborationWorkspace*` schema、`collaboration-workspace` 路径和前端页面目录还没开始系统迁移，下一轮要接上生成链一起处理。

## 2026-04-18 collaboration 导航 seed 合并与真实浏览器闭环

### 本次改动
- 在 [backend/cmd/migrate/main.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/migrate/main.go) 新增 `consolidateNavigationSeeds`，把数据库里残留的 `CollaborationWorkspace*` 菜单和 `collaboration_workspace.message.*` 页面种子自动折叠到 canonical `Collaboration*` 记录，并在 `go run ./cmd/migrate` 时幂等执行。
- 实际回收了旧 `menu_definitions` / `space_menu_placements` / `ui_pages` / `page_space_bindings` 中的 legacy 记录，浏览器失败根因“路由已是 `/collaboration/*`，但动态组件仍返回 `/collaboration-workspace/*`”已经消除。
- 已重新执行生成与校验链：`bundle`、`lint`、`ogen`、`gen-permissions`、前端 `pnpm run gen:api`、`go run ./cmd/migrate`、`go test ./internal/api/handlers ./internal/api/router ./internal/modules/system/permission ./cmd/migrate -count=1`、`pnpm exec vue-tsc --noEmit`、`pnpm build` 全部通过。Windows 环境没有 `make`，本轮改为显式执行同等生成命令。
- 真实浏览器回归已通过两轮：`frontend/e2e/tests/workspace-collaboration-pages.real.spec.ts` 与 `frontend/e2e/tests/high-risk-send.real.spec.ts` 共 4 个用例全部通过，覆盖 `/system/message`、`/collaboration/message`、`/collaboration/workspaces`、`/collaboration/members`、`/collaboration/roles` 等关键主路径。

### 下次方向
- 终态残留扫描仍有较多 legacy 命中，热点集中在 `backend/internal/modules/system/user/repository.go`、`backend/internal/modules/system/system/message_service.go`、`backend/internal/modules/system/collaborationworkspace/service.go`、`backend/internal/pkg/permissionkey/permissionkey.go`、`frontend/src/views/message/modules/message-dispatch-console.vue` 等文件；其中相当一部分仍是兼容旧表/旧字段的运行时桥接或旧类型名。
- 下一轮应按热点继续拆分收口：优先处理 `user/repository.go`、`message_service.go`、`permission/service.go`、`permissionkey.go` 与前端消息/治理 API 的 legacy fallback，再决定是否把剩余旧模型名进一步压到仅历史迁移和 changelog 可见。
