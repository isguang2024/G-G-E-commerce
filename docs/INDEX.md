# 文档索引

这个索引只收当前仓库里真正要用的入口。目标是让人不需要翻目录猜路径，也能快速判断哪份文档负责什么。

## 0. 快速导航

- 做任务推进：[`V5_REFACTOR_TASKS.md`](V5_REFACTOR_TASKS.md)
- 改后端 API：[`../backend/api/openapi/README.md`](../backend/api/openapi/README.md)
- 改前端实现：[`../FRONTEND_GUIDELINE.md`](../FRONTEND_GUIDELINE.md)
- 查文档规范：[`GUIDELINES.md`](GUIDELINES.md)
- 查导航模板：[`guides/doc-navigation-template.md`](guides/doc-navigation-template.md)

## 1. 一级入口

| 文档 | 作用 |
| --- | --- |
| [`../README.md`](../README.md) | 根目录快速入口，先看仓库分工和阅读顺序 |
| [`README.md`](README.md) | `docs/` 目录总入口 |
| [`GUIDELINES.md`](GUIDELINES.md) | 文档与协作写作规则 |
| [`V5_REFACTOR_TASKS.md`](V5_REFACTOR_TASKS.md) | V5 当前任务总入口与阶段进度 |
| [`API_OPENAPI_FIXED_FLOW.md`](API_OPENAPI_FIXED_FLOW.md) | API / OpenAPI 固定闭环流程 |
| [`project-structure.md`](project-structure.md) | 当前有效代码结构和模块分工 |

## 2. 真相源

| 主题 | 入口 |
| --- | --- |
| 后端框架、模块边界、5.0 约束 | [`../PROJECT_FRAMEWORK.md`](../PROJECT_FRAMEWORK.md) |
| 前端实现规范、壳层和状态管理 | [`../FRONTEND_GUIDELINE.md`](../FRONTEND_GUIDELINE.md) |
| API 契约、生成链路、错误码 | [`../backend/api/openapi/README.md`](../backend/api/openapi/README.md) |
| V5 阶段任务和收口范围 | [`V5_REFACTOR_TASKS.md`](V5_REFACTOR_TASKS.md) |

## 3. 专题手册

| 文档 | 作用 |
| --- | --- |
| [`guides/README.md`](guides/README.md) | 专题手册入口 |
| [`guides/commands.md`](guides/commands.md) | 常用命令与生成链路速查 |
| [`guides/add-endpoint.md`](guides/add-endpoint.md) | 新增后端接口的标准步骤 |
| [`guides/api-auto-registration.md`](guides/api-auto-registration.md) | 接口自动入库机制 |
| [`guides/permission-system.md`](guides/permission-system.md) | 权限系统结构与调试 |
| [`guides/database.md`](guides/database.md) | 数据库迁移、重置、seed |
| [`guides/permission-audit.md`](guides/permission-audit.md) | 权限审计说明 |

## 4. 当前工作线

| 代码域 | 入口 |
| --- | --- |
| 后端实现 | [`../backend/`](../backend/) |
| 前端实现 | [`../frontend/`](../frontend/) |
| 功能/流程说明（迁移） | [`../.claude/Instructions/`](../.claude/Instructions/) |
| 阶段审计记录 | [`reports/`](reports/) |

## 5. 阅读顺序

### 新接手仓库

1. `../README.md`
2. `V5_REFACTOR_TASKS.md`
3. `API_OPENAPI_FIXED_FLOW.md`
4. `guides/README.md`
5. `project-structure.md`

### 做 API 变更

1. `../backend/api/openapi/README.md`
2. `API_OPENAPI_FIXED_FLOW.md`
3. `guides/add-endpoint.md`
4. `guides/api-auto-registration.md`

### 做前端联调

1. `../FRONTEND_GUIDELINE.md`
2. `../frontend/`
3. `V5_REFACTOR_TASKS.md`

## 6. 约定

- 如与 `AGENTS.md`、`PROJECT_FRAMEWORK.md`、`FRONTEND_GUIDELINE.md`、`backend/CLAUDE.md`、`docs/V5_REFACTOR_TASKS.md` 冲突，以这些真相源为准。
- 这里优先写“入口”和“职责”，不写重复的实现细节。
- 新增文档先判断它是总入口、专题手册还是阶段任务，再决定放哪一层。
- 如果是生成产物、契约产物或临时说明，不要塞进索引主干。
- 索引中的目录链接必须是仓库内真实存在路径。

## 7. 快速翻阅索引（按内容类别）

| 类别 | 入口 |
| --- | --- |
| 功能说明文档 | [`project-structure.md`](project-structure.md)、[`../.claude/Instructions/features/register-system-design.md`](../.claude/Instructions/features/register-system-design.md) |
| 流程介绍文档 | [`V5_REFACTOR_TASKS.md`](V5_REFACTOR_TASKS.md)、[`API_OPENAPI_FIXED_FLOW.md`](API_OPENAPI_FIXED_FLOW.md) |
| 架构设计文档 | [`../PROJECT_FRAMEWORK.md`](../PROJECT_FRAMEWORK.md)、[`../backend/README.md`](../backend/README.md) |
| API 文档 | [`../backend/api/openapi/README.md`](../backend/api/openapi/README.md)、[`guides/add-endpoint.md`](guides/add-endpoint.md) |
| 配置与部署文档 | [`guides/commands.md`](guides/commands.md)、[`guides/database.md`](guides/database.md) |
