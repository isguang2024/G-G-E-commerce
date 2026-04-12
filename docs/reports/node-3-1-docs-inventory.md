# 3.1 文档盘点与分类清单

生成时间：2026-04-12  
任务节点：`V5-FOLDER-OPTIMIZE/STAGE-DOCS-CLEANUP/1`  
扫描范围：`docs/`、仓库根目录文档、`.claude/`（排除 `.claude/worktrees/` 镜像）

## 分类规则

- `V5相关`：当前 5.0 基线实施与协作直接依赖，默认保留。
- `历史版本`：阶段性清理记录或一次性过程文档，作为 3.2 删除候选。
- `重复`：与现有真相源或报告内容重叠，作为 3.3 合并/删除候选。
- `待保留`：当前仍可能有用，但需在后续节点进一步确认归档策略。

## A. 根目录文档

| 文件 | 分类 | 说明 |
| --- | --- | --- |
| `AGENTS.md` | V5相关 | 项目协作真相源，必须保留。 |
| `PROJECT_FRAMEWORK.md` | V5相关 | 框架约束文档，必须保留。 |
| `FRONTEND_GUIDELINE.md` | V5相关 | 前端实施准则，必须保留。 |
| `README.md` | 待保留 | 根导航文档（当前未跟踪），建议后续纳入并补齐指向 `docs/INDEX.md`。 |

## B. docs/ 目录文档

### B1 V5相关（默认保留）

- `docs/V5_REFACTOR_TASKS.md`
- `docs/API_OPENAPI_FIXED_FLOW.md`
- `docs/GUIDELINES.md`
- `docs/INDEX.md`
- `docs/README.md`
- `docs/project-structure.md`
- `docs/register-setup-guide.md`
- `docs/register-system-design.md`
- `docs/guides/README.md`
- `docs/guides/add-endpoint.md`
- `docs/guides/api-auto-registration.md`
- `docs/guides/commands.md`
- `docs/guides/database.md`
- `docs/guides/permission-audit.md`
- `docs/guides/permission-system.md`

### B2 历史版本（3.2 删除候选）

- `docs/frontend-cleanup-p1ab-notes.md`
- `docs/frontend-cleanup-p2a-notes.md`
- `docs/frontend-cleanup-p2b-notes.md`
- `docs/frontend-cleanup-p3b-notes.md`
- `docs/multi-app-playwright-smoke.md`

### B3 重复（3.3 已处理）

- `docs/v5-folder-optimize-delete-list.md`  
  与当前任务树（3.x 节点）中的盘点动作重叠，已在 3.3 节点删除并由阶段报告承接。

### B4 待保留（后续确认）

- `docs/multi-app-hosting-foundation.md`  
  仍具设计参考价值，但与当前 V5 主线关联度需在 5.x 节点确认。
- `docs/reports/node-1-3-duplicate-audit.md`
- `docs/reports/node-1-4-utils-types-audit.md`
- `docs/reports/node-1-5-1-6-assets-deps-audit.md`
- `docs/reports/node-2-1-frontend-structure-audit.md`
- `docs/reports/node-2-2-backend-structure-audit.md`
- `docs/reports/node-2-3-2-4-config-output-audit.md`
- `docs/reports/node-2-6-import-verify.md`  
  以上为阶段审计产物，建议统一归档路径与命名规范后再决定是否保留全部明细文件。

## C. `.claude/` 文档

- `/.claude/skills/change-wrapup/SKILL.md`：`V5相关`（仓库内技能真相源，必须保留）

## 统计汇总

- V5相关：20
- 历史版本：5
- 重复：1
- 待保留：9
- 合计：35

## 下一步建议（对接 3.2 / 3.3）

1. 3.2 先处理 `历史版本` 5 个文档，删除前做一次引用检索确认无外链依赖。
2. 3.3 再处理 `重复` 与 `待保留`，按“合并到真相源 -> 删除冗余”的策略执行。
