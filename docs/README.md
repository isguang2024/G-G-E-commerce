# 文档中枢

这个目录只保留当前有效的文档入口和专题索引。目标不是堆说明，而是让人能快速定位到“该看什么、该改什么、真相源在哪里”。

若与 `AGENTS.md`、`PROJECT_FRAMEWORK.md`、`FRONTEND_GUIDELINE.md`、`backend/CLAUDE.md`、`docs/V5_REFACTOR_TASKS.md` 冲突，以这些真相源为准。

## 快速入口

1. [`INDEX.md`](INDEX.md) - 文档总索引
2. [`GUIDELINES.md`](GUIDELINES.md) - 文档与协作规则
3. [`V5_REFACTOR_TASKS.md`](V5_REFACTOR_TASKS.md) - 当前阶段和任务入口
4. [`API_OPENAPI_FIXED_FLOW.md`](API_OPENAPI_FIXED_FLOW.md) - API 契约闭环流程
5. [`guides/README.md`](guides/README.md) - 专题手册入口

## 面向不同角色

- 新接手仓库：先看 `INDEX.md`，再看 `V5_REFACTOR_TASKS.md`
- 做后端接口：先看 `backend/api/openapi/README.md`，再看 `guides/add-endpoint.md`
- 做前端页面：先看 `FRONTEND_GUIDELINE.md`，再回到 `frontend/`
- 做文档收口：先看 `GUIDELINES.md`

## 目录原则

- 真相源优先于说明文
- 指南优先于散落备注
- 专题文档优先链接到唯一入口，避免重复解释

## 阶段 5 保留策略（docs 转型）

本轮转型后，`docs/` 目录按以下优先级保留：

1. 任务与进度：`V5_REFACTOR_TASKS.md`、`reports/`
2. 发布记录：`../CHANGELOG.md`
3. 真相源级流程与规范：`API_OPENAPI_FIXED_FLOW.md`、`GUIDELINES.md`、`INDEX.md`

说明：功能/流程说明文档已迁移到 `../.claude/Instructions/`，`docs` 保留导航与进度主干，不再承担执行型说明的主存放职责。

## docs 目录新规范（执行版）

- `docs/` 只保留三类内容：
  - 任务与进度（`V5_REFACTOR_TASKS.md`、`reports/`）
  - 导航与规范（`INDEX.md`、`GUIDELINES.md`）
  - API 流程真相源（`API_OPENAPI_FIXED_FLOW.md`）
- 功能说明与流程操作手册统一放到 `../.claude/Instructions/`。
- 新增执行记录统一落在 `docs/reports/`，命名遵循 `node-<阶段>-<节点>-<主题>.md`。
- 若新增文档不属于上述类别，默认不放入 `docs/` 主干。
