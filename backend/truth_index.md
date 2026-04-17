# backend/truth_index.md

> 后端开发真相详细索引。每个条目说明文档用途与触发场景，方便按场景快速定位。
>
> 说明：`docs/tmp/` 下的设计稿、Gap Analysis、交付稿都不是后端真相；若与本页条目冲突，以本页链接到的 `backend/Truth/*` 文档为准。

## 真相文档一览

| 文档 | 位置 | 触发场景 |
| --- | --- | --- |
| 后端开发指引 | [Truth/backend-guide.md](./Truth/backend-guide.md) | 新加 endpoint、sub-handler 拆分、反模式清单 |
| API 闭环流程 | [Truth/api-openapi-flow.md](./Truth/api-openapi-flow.md) | 新增/修改 API 的完整 14 步流程 |
| OpenAPI spec 编辑规则 | [Truth/openapi-spec-rules.md](./Truth/openapi-spec-rules.md) | 改 `api/openapi/` 下 spec、`x-permission-key` 协同 |
| 新模块完整接入 checklist | [Truth/new-module-checklist.md](./Truth/new-module-checklist.md) | 新增带管理页的系统模块（迁移/seed/权限/菜单/功能包）|
| 权限系统 | [Truth/permission-system.md](./Truth/permission-system.md) | 理解权限公式、三层数据结构、调试判权 |
| 数据库与迁移 | [Truth/database.md](./Truth/database.md) | migrate / dbreset / init-admin / seed 排障 |
| 常用命令 | [Truth/commands.md](./Truth/commands.md) | 启动、代码生成、检查命令速查 |
| 日志与审计 | [Truth/logging-spec.md](./Truth/logging-spec.md) | 日志上下文、Request ID、审计事件、`/telemetry/logs` |
| GitHub OAuth 配置 | [Truth/social-oauth-github.md](./Truth/social-oauth-github.md) | 配置社交登录的 env、回调地址、启用步骤 |
| 上传系统总览 | [Truth/upload-system/overview.md](./Truth/upload-system/overview.md) | 上传架构、ADR、API、Driver 扩展、运维 |
| 上传配置运维 | [Truth/upload-system/ops-guide.md](./Truth/upload-system/ops-guide.md) | Provider/Bucket/UploadKey/Rule 的配置示例与排障 |
| 上传扩展参数约束 | [Truth/upload-system/schema-registry.md](./Truth/upload-system/schema-registry.md) | 给 UploadKey/Rule 加 `extra_schema` 字段时的约束 |

## 按场景找文档

| 你要做什么 | 看哪份 |
| --- | --- |
| 加一个 API | `api-openapi-flow.md` → `openapi-spec-rules.md` → `backend-guide.md` |
| 加一个带管理页的模块 | `new-module-checklist.md` |
| 改权限键 / 权限模型 | `openapi-spec-rules.md`（协同清单）+ `permission-system.md` |
| 排查权限不生效 | `permission-system.md` |
| 排查上传失败 | `upload-system/overview.md`（故障排查表）+ `upload-system/ops-guide.md` |
| 跑数据库迁移 | `database.md` |
| 找某个命令 | `commands.md` |
| 接 GitHub 登录 | `social-oauth-github.md` |
| 看日志规范 | `logging-spec.md` |
