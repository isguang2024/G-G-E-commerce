# AGENTS.md

## 协作约束

- 使用中文沟通。
- 通过 Shell 读写文本必须显式 UTF-8，避免乱码。
- 当前协作文档真相源如下，其余说明若与它们冲突，一律按这里为准：
  - `AGENTS.md`
  - `docs/project-framework.md`
  - `docs/frontend-guideline.md`
  - `backend/CLAUDE.md`

## 实施原则

- 搜索代码优先用 `rg`；若 Windows 终端编码导致结果异常，再回退到 PowerShell 的 `Get-Content` / `Select-String`，并显式指定 UTF-8。
- 数据库允许清空重建；迁移只负责一次性结构变更或历史数据修正，**默认数据走 seed / ensure 幂等逻辑**，不要把长期默认状态反复写进迁移链。
- 只要本次改动涉及 migration，必须优先创建并落当前迁移，再继续后续实现；不要拖到收尾阶段补迁移，避免并行开发时新增迁移插队，导致编号、内容或目标状态被覆盖。
- 临时修复型迁移在目标状态达成后必须删除，不长期保留。
- 不手写已有成熟模块能解决的能力（路由、校验、API 文档、权限模型、迁移、DI、缓存）。新增依赖前先核对当前技术栈。
- 新模块 / 新表 / 新接口在评审时必须显式回答：**是否带 tenant_id、是否在仓储层强制过滤**。
- API 一律走 OpenAPI-first：spec 即真相，先改 `backend/api/openapi/`，再刷新生成物，再写实现；不允许手写 router/dto 绕开 ogen 生成。
- `backend/api/gen/` 与 `frontend/src/api/v5/`（当前包含 `client.ts`、`types.ts`、`schema.d.ts`、`error-codes.ts`）属于生成产物，禁止手改生成文件本体；业务封装只能写在生成层之外。
- 权限判断一律走 `backend/internal/pkg/permission/evaluator`，不允许在 handler / service 内散写权限交集逻辑。
- 前端已接真实接口；任何后端契约变更必须同步更新 OpenAPI spec、后端生成代码与前端生成 client。

## API 变更固定步骤

- 新能力默认按以下顺序推进，不允许倒序跳步：
  `model/domain → migration → seed/ensure → OpenAPI spec → bundle → lint → ogen → gen-permissions → restart backend → router/bridge check → sub-handler/service → frontend gen:api → frontend API 封装 → UI → build/test/browser verify`
- 其中：
  - 结构变更先落 `model + migration`
  - 长期默认数据走 `seed / ensure`，不反复写进 migration
  - API 契约只改 `backend/api/openapi/`
  - 生成物只通过 `bundle / ogen / gen-permissions / pnpm run gen:api` 刷新，不手改
  - `gen-permissions` 产出 `openapi_seed.json`，路由↔权限键映射在 `router` 初始化时读入一次、进程级缓存；合并/重命名/删除权限键后**必须重启后端**，否则旧映射继续生效，会出现"DB 已对齐但接口仍 403"的幻觉
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
  6. **重启正在运行的后端进程**，让 `openapi_seed.json` 的路由↔权限键映射重新加载；跳过这一步会出现"新键已生效但旧键仍被网关拦截"的假权限问题。
  7. 执行 `go test ./internal/api/handlers -count=1`。
  8. 执行 `pnpm exec vue-tsc --noEmit`。
- 常规情况下优先执行上面的显式步骤；若需要一次性刷新后端 OpenAPI 生成链，可使用 `backend/update-openapi.bat`，但仍要补做前端 `pnpm run gen:api` 与联编校验。
- 若本次 API 变更同时影响权限点、默认数据或错误码，还必须继续执行对应生成步骤（如 `go run ./cmd/gen-permissions`）并检查受影响产物。
- 未完成上述生成、修正、校验前，不得判定"接口改造完成"，也不得开始依赖该接口继续开发下游功能。

## 协作技能位置（Claude Code 与 Codex 共用）

- 技能统一按用户目录全局技能读取（`~/.claude/skills/`、`~/.codex/skills/`）。
  - 本仓库执行任务时，不再强制要求从 `<repo>/.claude/skills/` 或 `maben/<skill-name>` 加载。
- 当仓库内技能与全局技能同名时，以全局技能为准。
- 若需调整技能规则，优先修改全局技能，不要求回灌仓库副本。

## Handler 域拆分规范（God Handler Split）

`backend/internal/api/handlers/` 已完成"god handler"拆分。每个域有独立的 sub-handler：

### 目录结构约定

- `{domain}_handler.go` — 域 sub-handler 的结构体定义 + 构造函数（持有该域所需的最小依赖集）
- `{domain}.go` — 所有 receiver 为 `*{domain}APIHandler` 的 op 实现
- `workspace.go` — `APIHandler` 主结构体（嵌入所有域 sub-handler + workspace 自身方法 + `NewAPIHandler`）

### 扩展新域的步骤

1. 在 `{domain}_handler.go` 中定义 `{domain}APIHandler` struct 与 `new{Domain}APIHandler(...)` 构造函数；
2. 在 `{domain}.go` 中将所有 op 方法 receiver 改为 `*{domain}APIHandler`；
3. 在 `workspace.go` 的 `APIHandler` struct 中添加 `*{domain}APIHandler` 匿名嵌入；
4. 在 `NewAPIHandler` 中调用构造函数并赋值；
5. 用 `go build ./internal/api/handlers/...` 验证无歧义选择器错误。

### 深度规则（防止方法歧义）

- `handlerBase`（包含 `gen.UnimplementedHandler`）位于深度 2；
- 各域 sub-handler（`*dictionaryAPIHandler` 等）位于深度 1；
- Go 自动选取最浅深度，sub-handler 的 op 方法自动覆盖 UnimplementedHandler 的 stub。
- 若两个 sub-handler 都有同名方法会产生歧义编译错误：应重新检查域边界，确保每个 op 只归一个域。

### 迁移历史

现有 19 个域已完成拆分（见 `workspace.go` 嵌入列表）。测试文件中直接构建
`&APIHandler{...}` 时，应通过嵌入字段传入子 handler：
```go
h := &APIHandler{
    logPolicyAPIHandler: &logPolicyAPIHandler{policyRepo: repo, ...},
}
```

## 风险动作纪律

- 删除文件、清库、force push、删除分支等不可逆动作必须先确认。
- 迁移失败、hook 失败时排查根因，禁止 `--no-verify` 或跳过校验。
