# AGENTS.md

## 协作约束

- 使用中文沟通。
- 通过 Shell 读写文本必须显式 UTF-8，避免乱码。
- 仓库当前正在执行 5.0 重构，所有任务从 `docs/V5_REFACTOR_TASKS.md` 出发；权威设计基线为外部文档《GGE_5.0_初始化架构文档.docx》。
- 根目录有效协作文档只有三份，其余一律视为过期：
  - `AGENTS.md`
  - `PROJECT_FRAMEWORK.md`
  - `FRONTEND_GUIDELINE.md`

## 实施原则

- 代码搜索/修改不使用 rg。
- 数据库允许清空重建；迁移只负责一次性结构变更或历史数据修正，**默认数据走 seed / ensure 幂等逻辑**，不要把长期默认状态反复写进迁移链。
- 临时修复型迁移在目标状态达成后必须删除，不长期保留。
- 不手写已有成熟模块能解决的能力（路由、校验、API 文档、权限模型、迁移、DI、缓存）。新增依赖前先核对 V5 任务文档的技术栈定锤表。
- 新模块 / 新表 / 新接口在评审时必须显式回答：**是否带 tenant_id、是否在仓储层强制过滤**。
- API 一律走 OpenAPI-first：spec 即真相，先改 `api/openapi/`，再 `make gen`，再写实现；不允许手写 router/dto 绕开 ogen 生成。
- 权限判断一律走 `pkg/permission/evaluator`，不允许在 handler 内散写权限交集逻辑。
- 前端已接真实接口；任何后端契约变更必须同步更新 OpenAPI spec 与前端生成 client。

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
