# AGENTS.md

## 协作约束

- 使用中文沟通。
- 通过 Shell 读写文本必须显式 UTF-8，避免乱码。
- 当前仓库已完成 5.0 基线落地，日常工作以**增量开发 + 历史收口**为主；涉及架构、API、权限、迁移的任务，先核对 `docs/V5_REFACTOR_TASKS.md` 与外部基线文档《GGE_5.0_初始化架构文档.docx》。
- 当前协作文档真相源如下，其余说明若与它们冲突，一律按这里为准：
  - `AGENTS.md`
  - `PROJECT_FRAMEWORK.md`
  - `FRONTEND_GUIDELINE.md`
  - `docs/V5_REFACTOR_TASKS.md`

## 实施原则

- 搜索代码优先用 `rg`；若 Windows 终端编码导致结果异常，再回退到 PowerShell 的 `Get-Content` / `Select-String`，并显式指定 UTF-8。
- 数据库允许清空重建；迁移只负责一次性结构变更或历史数据修正，**默认数据走 seed / ensure 幂等逻辑**，不要把长期默认状态反复写进迁移链。
- 临时修复型迁移在目标状态达成后必须删除，不长期保留。
- 不手写已有成熟模块能解决的能力（路由、校验、API 文档、权限模型、迁移、DI、缓存）。新增依赖前先核对 V5 任务文档的技术栈定锤表。
- 新模块 / 新表 / 新接口在评审时必须显式回答：**是否带 tenant_id、是否在仓储层强制过滤**。
- API 一律走 OpenAPI-first：spec 即真相，先改 `backend/api/openapi/`，再刷新生成物，再写实现；不允许手写 router/dto 绕开 ogen 生成。
- `backend/api/gen/` 与 `frontend/src/api/v5/`（当前包含 `client.ts`、`types.ts`、`schema.d.ts`、`error-codes.ts`）属于生成产物，禁止手改生成文件本体；业务封装只能写在生成层之外。
- 权限判断一律走 `backend/internal/pkg/permission/evaluator`，不允许在 handler / service 内散写权限交集逻辑。
- 前端已接真实接口；任何后端契约变更必须同步更新 OpenAPI spec、后端生成代码与前端生成 client。

## API 变更固定步骤

- 每次**新增 API**或**修改 API 契约**后，必须按以下顺序执行，不允许跳步：
  1. 先修改 `backend/api/openapi/` 下的 spec，保持 OpenAPI 为唯一真相源。
  2. 执行 `bundle`，生成最新 `backend/api/openapi/dist/openapi.yaml`。
  3. 执行 `ogen`，刷新 `backend/api/gen/`。
  4. 执行前端 `pnpm run gen:api`，刷新 `frontend/src/api/v5/schema.d.ts`；若错误码或 client 辅助文件同步变化，一并检查 `frontend/src/api/v5/`。
  5. 基于最新生成物修正后端 handler / service 与前端调用代码，禁止继续依赖旧字段或旧签名。
  6. 执行 `go test ./internal/api/handlers -count=1`。
  7. 执行 `pnpm exec vue-tsc --noEmit`。
- 常规情况下优先执行上面的显式步骤；若需要一次性刷新后端 OpenAPI 生成链，可使用 `backend/update-openapi.bat`，但仍要补做前端 `pnpm run gen:api` 与联编校验。
- 若本次 API 变更同时影响权限点、默认数据或错误码，还必须继续执行对应生成步骤（如 `go run ./cmd/gen-permissions`）并检查受影响产物。
- 未完成上述生成、修正、校验前，不得判定“接口改造完成”，也不得开始依赖该接口继续开发下游功能。

## 协作技能位置（Claude Code 与 Codex 共用）

- 项目相关技能统一放在 `<repo>/.claude/skills/`，作为单一真相源，跟随 git。
- Claude Code 自动加载该目录。
- Codex 通过用户目录下的 junction `~/.codex/skills/gge` → `<repo>/.claude/skills/` 复用同一份；本仓库内运行的 codex 任务必须从 `gge/<skill-name>` 加载技能。
- 当仓库内技能与 codex 全局技能同名（典型：`change-wrapup`），**一律以仓库版 `gge/<skill-name>` 为准**，禁止使用全局同名技能。
- 跨项目通用技能（如 fluent-react-v9 等）才允许放 `~/.claude/` 或 `~/.codex/` 用户目录；项目相关技能不在用户目录维护副本，避免双份漂移。
- 仓库内技能的修改一律在 `<repo>/.claude/skills/` 直接改，不在用户目录改后再回灌。

## 风险动作纪律

- 删除文件、清库、force push、删除分支等不可逆动作必须先确认。
- 迁移失败、hook 失败时排查根因，禁止 `--no-verify` 或跳过校验。
