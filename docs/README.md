# 文档中枢

这个目录只保留当前有效的文档入口和专题索引。目标是让人能快速定位到"该看什么、该改什么、真相源在哪里"。

若与 `AGENTS.md`、`PROJECT_FRAMEWORK.md`、`FRONTEND_GUIDELINE.md`、`backend/CLAUDE.md` 冲突，以这些真相源为准。

## 快速入口

1. [`INDEX.md`](INDEX.md) - 文档总索引
2. [`GUIDELINES.md`](GUIDELINES.md) - 文档与协作规则
3. [`API_OPENAPI_FIXED_FLOW.md`](API_OPENAPI_FIXED_FLOW.md) - API 契约闭环流程
4. [`guides/README.md`](guides/README.md) - 专题手册入口
5. [`project-structure.md`](project-structure.md) - 代码结构与模块分工

## 面向不同角色

- 新接手仓库：先看 `INDEX.md`，再看 `project-structure.md`
- 做后端接口：先看 `backend/api/openapi/README.md`，再看 `guides/add-endpoint.md`
- 做前端页面：先看 `FRONTEND_GUIDELINE.md`，再回到 `frontend/`
- 做文档维护：先看 `GUIDELINES.md`

## 目录原则

- 真相源优先于说明文
- 指南优先于散落备注
- 专题文档优先链接到唯一入口，避免重复解释

## docs 目录规范

- `docs/` 保留以下内容：
  - 导航与规范（`INDEX.md`、`GUIDELINES.md`）
  - API 流程真相源（`API_OPENAPI_FIXED_FLOW.md`）
  - 代码结构说明（`project-structure.md`）
  - 专题开发手册（`guides/`）
  - 历史审计记录（`reports/`）
- 功能说明与流程操作手册放到 `.claude/Instructions/`。
- 若新增文档不属于上述类别，默认不放入 `docs/` 主干。
