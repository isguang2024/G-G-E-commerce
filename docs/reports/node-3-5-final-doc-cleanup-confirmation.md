# 3.5 文档删除清单与确认

生成时间：2026-04-12  
任务节点：`V5-FOLDER-OPTIMIZE/STAGE-DOCS-CLEANUP/5`

## 一、已执行删除清单

| 文件路径 | 删除原因 | 来源节点 |
| --- | --- | --- |
| `docs/frontend-cleanup-p1ab-notes.md` | `CLEANUP-V1` 历史笔记，已过期 | 3.2 |
| `docs/frontend-cleanup-p2a-notes.md` | `CLEANUP-V1` 历史笔记，已过期 | 3.2 |
| `docs/frontend-cleanup-p2b-notes.md` | `CLEANUP-V1` 历史笔记，已过期 | 3.2 |
| `docs/frontend-cleanup-p3b-notes.md` | `CLEANUP-V1` 历史笔记，已过期 | 3.2 |
| `docs/v5-folder-optimize-delete-list.md` | 与阶段盘点清单重复 | 3.3 |

## 二、待保留文档清单

| 文件路径 | 保留原因 |
| --- | --- |
| `docs/V5_REFACTOR_TASKS.md` | 当前任务真相源 |
| `docs/API_OPENAPI_FIXED_FLOW.md` | API 契约固定流程 |
| `docs/GUIDELINES.md` | 文档协作规范 |
| `docs/INDEX.md` | 文档总索引 |
| `docs/README.md` | docs 目录入口 |
| `docs/multi-app-playwright-smoke.md` | 多 APP smoke 基线仍在使用 |
| `docs/multi-app-hosting-foundation.md` | 设计参考文档，后续阶段再判定 |
| `docs/reports/*.md` | 阶段执行证据与审计记录 |
| `AGENTS.md`、`PROJECT_FRAMEWORK.md`、`FRONTEND_GUIDELINE.md`、`backend/CLAUDE.md` | 协作真相源 |

## 三、确认结论

1. 本阶段删除对象已限定为“历史过期 + 重复文档”，未触碰当前真相源。  
2. 删除后文档体系可通过 `docs/README.md` 与 `docs/INDEX.md` 正常导航。  
3. 阶段 3 的删除动作已完成，可进入下一阶段（文档导航模板与 docs 转型）。
