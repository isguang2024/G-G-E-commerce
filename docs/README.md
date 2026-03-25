# docs 文档索引

## 阅读顺序

1. `../PROJECT_FRAMEWORK.md`：项目主框架与执行清单（全局约束）。
2. `permission-overall-summary.md`：权限系统总体模型与现状。
3. `menu-page-management-design.md`：菜单、全局菜单、页面管理的正式设计方案。
4. `menu-page-management-implementation-plan.md`：菜单与页面管理的数据库、接口、守卫和迁移实施草案。
5. `change-log.md`：按时间追加的改动记录。
6. `../FRONTEND_GUIDELINE.md`：前端系统管理页统一规范。
7. `../AGENTS.md`：仓库协作规则与交付约束。

## 阅读路径

- 新人接手：先看 `../PROJECT_FRAMEWORK.md`，再看 `permission-overall-summary.md`、`menu-page-management-design.md` 与 `menu-page-management-implementation-plan.md`，最后看 `change-log.md` 了解最近演进。
- 日常开发：先看任务相关模块文档，再回看 `../FRONTEND_GUIDELINE.md` 与 `../AGENTS.md` 校对实现与交付约束。
- 评审回归：先看 `change-log.md` 最近记录，再核对 `permission-overall-summary.md` 与代码当前行为。

## 文档维护规则

- `docs/` 只保留长期有效文档与变更日志。
- 重复、过期、阶段性草案要合并后删除，避免多版本并存。
- 涉及核心链路（权限、API、迁移、上下文）变更时，必须同步更新对应文档。
- 文档与代码冲突时，以代码真实行为为准，随后补文档。
- 删除文档时，必须同步检查 `change-log.md` 与根目录文档中是否仍存在旧链接或旧名称引用。

## 本次收敛

- 已移除重复设计文档 `permission-package-design.md`，内容已并入现有权限总览与框架文档。
