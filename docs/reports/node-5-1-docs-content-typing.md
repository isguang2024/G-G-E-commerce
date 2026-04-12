# 5.1 docs 内容类型梳理与迁移基线

生成时间：2026-04-12  
任务节点：`V5-FOLDER-OPTIMIZE/STAGE-DOCS-MIGRATION/1`

## 现状分类

### A. 规范/流程（建议保留 docs）

- `docs/V5_REFACTOR_TASKS.md`
- `docs/API_OPENAPI_FIXED_FLOW.md`
- `docs/GUIDELINES.md`
- `docs/INDEX.md`
- `docs/README.md`
- `docs/project-structure.md`
- `docs/guides/*`

### B. 功能与方案说明（待 5.2 决定迁移）

- `docs/register-system-design.md`
- `docs/register-setup-guide.md`
- `docs/multi-app-hosting-foundation.md`
- `docs/multi-app-playwright-smoke.md`

### C. 执行审计与进度证据（建议保留 docs/reports）

- `docs/reports/*.md`（阶段节点执行记录与审查报告）

## 迁移可行性判断

- 当前不存在 `.claude/Instructions/` 目录，需在 5.2 决定是否创建并迁移 B 类文档。
- A 类与 C 类均属于当前 docs 主干职责，5.1 结论为“先保留不迁移”。

## 下一步建议（对接 5.2）

1. 先创建 `.claude/Instructions/` 目录结构（若确认执行迁移）。
2. 仅迁移 B 类中的“执行型说明”文档；架构真相源仍保留在 docs。
3. 更新 `docs/INDEX.md` 与 `docs/README.md` 的入口链接，避免迁移后断链。
