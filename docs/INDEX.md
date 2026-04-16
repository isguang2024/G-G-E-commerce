# 文档索引

这是仓库唯一推荐的文档导航页。目标是先判断“该看哪一类文档”，再进入唯一主文档。

## 真相源

| 主题 | 文档 |
| --- | --- |
| 协作约束 | [../AGENTS.md](../AGENTS.md) |
| 项目边界与核心语义 | [project-framework.md](project-framework.md) |
| 前端实现规范 | [frontend-guideline.md](frontend-guideline.md) |
| 后端开发约束 | [../backend/CLAUDE.md](../backend/CLAUDE.md) |

## 按任务找文档

| 你要做什么 | 先看 |
| --- | --- |
| 理解仓库结构 | [project-structure.md](project-structure.md) |
| 新增或修改 API | [../backend/api/openapi/README.md](../backend/api/openapi/README.md) |
| 走完整 API 闭环 | [API_OPENAPI_FIXED_FLOW.md](API_OPENAPI_FIXED_FLOW.md) |
| 查命令 | [guides/commands.md](guides/commands.md) |
| 查数据库/迁移/seed | [guides/database.md](guides/database.md) |
| 查权限模型 | [guides/permission-system.md](guides/permission-system.md) |
| 新增系统模块 | [guides/new-module-checklist.md](guides/new-module-checklist.md) |
| 查专题手册 | [guides/README.md](guides/README.md) |

## 文档分层

| 层级 | 作用 | 代表文档 |
| --- | --- | --- |
| 真相源 | 定义约束与边界 | `AGENTS.md`、`docs/project-framework.md` |
| 主说明 | 解释一个主题的当前做法 | `project-structure.md`、`API_OPENAPI_FIXED_FLOW.md` |
| 手册 | 面向具体操作的步骤或速查 | `guides/*.md` |
| 归档 | 历史材料，只做追溯 | `archive/*.md` |

## 当前推荐阅读顺序

### 新接手仓库

1. [../README.md](../README.md)
2. [project-framework.md](project-framework.md)
3. [project-structure.md](project-structure.md)
4. [guides/README.md](guides/README.md)

### 做 API 变更

1. [../backend/api/openapi/README.md](../backend/api/openapi/README.md)
2. [API_OPENAPI_FIXED_FLOW.md](API_OPENAPI_FIXED_FLOW.md)
3. [guides/commands.md](guides/commands.md)
4. [guides/database.md](guides/database.md)

### 做前端联调

1. [frontend-guideline.md](frontend-guideline.md)
2. [../frontend/README.md](../frontend/README.md)
3. [project-structure.md](project-structure.md)
