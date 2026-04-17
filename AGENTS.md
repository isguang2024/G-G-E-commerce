# AGENTS.md

> 本仓库 AI 协作的唯一真相源。`CLAUDE.md` / `CODEX.md` 等文件只做指向本文件的链接中转。

## 真相源分层

遇到冲突一律按以下顺序为准：

1. **根真相源**：本文件（`AGENTS.md`）
2. **子项目真相源**：
   - `backend/truth.md` + `backend/Truth/*`
   - `frontend/truth.md` + `frontend/Truth/*`
   - `frontend-platform/truth.md` + `frontend-platform/Truth/*`
3. **项目说明**（非真相）：`docs/` 下的各端子目录
4. **目录级 README**：只作目录介绍，不作真相

## 协作约束

- 使用中文沟通。
- 通过 Shell 读写文本必须显式 UTF-8，避免乱码。
- 所有 README.md 只做"当前目录介绍"，不细说业务和代码构成，不作真相。
- 开发真相只放在对应子项目的 `Truth/` 文件夹，其他位置一律不放真相文档。
- 临时任务、临时记忆、阶段性文档一律进入 `docs/tmp/`，不进主干导航。

## 项目定位

- 定位：**通用 Admin 管理后台脚手架**。
- 主干职责：认证、权限、路由/页面注册、OpenAPI 契约治理、运行时上下文、多APP、多空间治理等通用底座。
- 业务落位原则：垂直业务能力以模块化方式接入，不反向污染底座抽象。

仓库主线：

- `backend/` — Go 1.25 + Gin + GORM + Postgres + Redis + Elasticsearch + ogen + goose + OpenTelemetry
- `frontend/` — Vue 3 + TypeScript + Vite + Element Plus + Pinia + openapi-fetch（已接真实接口，不走 mock）
- `frontend-platform/` — 基于 Vben Admin 的前端平台工作空间（独立 monorepo）开发中

## 核心语义

- **Workspace 空间**：系统唯一的空间主实体，分 `personal` / `collaboration` 两类型。
- **Personal 空间**：个人账号自身的空间，权限来源以账号角色为主。
- **Collaboration 空间**：协作空间，权限来源以成员角色为主，空间功能包给出权限上限。
- **Member 成员**：账号在 workspace 内的身份记录。
- **最终权限公式**：`空间功能包权限键和 成员角色权限键交集`。个人空间可直接退化为账号角色权限集合。
- **菜单 / 页面 / 权限键** 三段分离：菜单管导航，页面管路由，权限键管访问。`menu_space_key` 只是某 app 下的导航菜单视图，不参与权限计算。

## 当前非目标

- 不开放租户能力、不暴露租户管理界面、不做 schema 分片。
- 新增设计不得以"菜单反推权限 / mock 接口 / 手写权限规则"模式扩散。
- 不维护第二套后台前端、不重启第二个后端工程。

## 实施原则

- 搜索代码优先用 `rg`；若 Windows 终端编码导致结果异常，再回退到 PowerShell 的 `Get-Content` / `Select-String`，并显式指定 UTF-8。
- 数据库允许清空重建；迁移只负责一次性结构变更或历史数据修正，**默认数据走 seed / ensure 幂等逻辑**，不要把长期默认状态反复写进迁移链。
- 只要本次改动涉及 migration，必须优先创建并落当前迁移，再继续后续实现；不要拖到收尾阶段补迁移，避免并行开发时新增迁移插队，导致编号、内容或目标状态被覆盖。
- 临时修复型迁移在目标状态达成后必须删除，不长期保留。
- 不手写已有成熟模块能解决的能力（路由、校验、API 文档、权限模型、迁移、DI、缓存等）。新增依赖前先核对当前技术栈。
- API 一律走 OpenAPI-first：spec 即真相，先改 `backend/api/openapi/`，再刷新生成物，再写实现；不允许手写 router/dto 绕开 ogen 生成。
- `backend/api/gen/` 与 `frontend/src/api/v5/`（当前包含 `client.ts`、`types.ts`、`schema.d.ts`、`error-codes.ts`）属于生成产物，禁止手改生成文件本体；业务封装只能写在生成层之外。
- 权限判断一律走 `backend/internal/pkg/permission/evaluator`，不允许在 sub-handler / service 内散写权限交集逻辑。
- 前端已接真实接口；任何后端契约变更必须同步更新 OpenAPI spec、后端生成代码与前端生成 client。

## API 变更固定步骤

- 新能力默认按以下顺序推进，不允许倒序跳步：
  `model/domain → migration → seed/ensure → OpenAPI spec → bundle → lint → ogen → gen-permissions → router/bridge check → sub-handler/service → frontend gen:api → frontend API 封装 → UI → build/test/browser verify`
- 其中：
  - 结构变更先落 `model + migration`
  - 长期默认数据走 `seed / ensure`，不反复写进 migration
  - API 契约只改 `backend/api/openapi/`
  - 生成物只通过 `bundle / ogen / gen-permissions / pnpm run gen:api` 刷新，不手改
  - API 网关配置在本仓库内等价于 `OpenAPI 扩展字段 + gin middleware 分组`；/api/v1 下的路由由 OpenAPI seed 驱动，启动时自动从 `permissionseed.LoadOpenAPISeed()` 派生并挂到 Gin，新增/删除 operation 不需要改 `backend/internal/api/router/router.go`
  - `router/bridge check` 默认只做覆盖率核对：对齐 `go test ./internal/api/router -count=1` 即可，**只有非 OpenAPI 入口**（`/health`、`/uploads`、OAuth 回调、WebSocket 等）需要手动在 `router.go` 注册
  - `sub-handler/service` 指按 domain 拆分到 `internal/api/handlers/{domain}.go` + `{domain}_handler.go`，不允许回退到单一 god `APIHandler` 堆方法
  - sub-handler 写完不算完成，必须继续收口前端类型、前端封装、UI 联调和校验

- 每次**新增 API**或**修改 API 契约**后，必须按以下顺序执行，不允许跳步：
  1. 先修改 `backend/api/openapi/` 下的 spec，保持 OpenAPI 为唯一真相源。
  2. 执行 `bundle`，生成最新 `backend/api/openapi/dist/openapi.yaml`。
  3. 执行 `ogen`，刷新 `backend/api/gen/`。
  4. 执行前端 `pnpm run gen:api`，刷新 `frontend/src/api/v5/schema.d.ts`；若错误码或 client 辅助文件同步变化，一并检查 `frontend/src/api/v5/`。
  5. 基于最新生成物修正后端 sub-handler / service 与前端调用代码，禁止继续依赖旧字段或旧签名。
  6. 执行 `go test ./internal/api/handlers -count=1`。
  7. 执行 `pnpm exec vue-tsc --noEmit`。
- 常规情况下优先执行上面的显式步骤；若需要一次性刷新后端 OpenAPI 生成链，可使用 `backend/update-openapi.bat`，但仍要补做前端 `pnpm run gen:api` 与联编校验。
- 若本次 API 变更同时影响权限点、默认数据或错误码，还必须继续执行对应生成步骤（如 `go run ./cmd/gen-permissions`）并检查受影响产物。
- 未完成上述生成、修正、校验前，不得判定"接口改造完成"，也不得开始依赖该接口继续开发下游功能。

## 协作技能位置（Claude Code 与 Codex 共用）
-长任务要使用任务树技能
- 技能统一按用户目录全局技能读取（`~/.claude/skills/`、`~/.codex/skills/`）。
 - 本仓库执行任务时，不再强制要求从 `<repo>/.claude/skills/` 或 `maben/<skill-name>` 加载。
- 当仓库内技能与全局技能同名时，以全局技能为准。
- 若需调整技能规则，优先修改全局技能，不要求回灌仓库副本。

## Handler 域拆分（摘要）

- `backend/internal/api/handlers/` 已完成 "god handler" 拆分，19 个域按 domain 各自持有 sub-handler
- 新增 op 必须归到某个 `*{domain}APIHandler`，不允许堆回主 `APIHandler`
- 同名方法出现在两个 sub-handler 会触发歧义编译错误 → 重新划分域边界
- 目录约定、扩展新域步骤、嵌入深度规则、测试构造姿势详见 [backend/Truth/backend-guide.md](backend/Truth/backend-guide.md)


