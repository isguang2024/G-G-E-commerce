# 专题手册

这里只保留仍然会被反复使用的高频操作手册。被主说明文档覆盖的旧手册、专项验收和排查报告统一放到 [../archive/README.md](../archive/README.md)。

## 开发手册

| 文档 | 适用场景 |
| --- | --- |
| [commands.md](commands.md) | 查命令、生成链和常用检查 |
| [database.md](database.md) | 查迁移、seed、数据库排障 |
| [permission-system.md](permission-system.md) | 理解权限模型与运行时判定 |
| [social-oauth-github.md](social-oauth-github.md) | 配置 GitHub OAuth |
| [frontend-observability-spec.md](frontend-observability-spec.md) | 前端可观测性与 `data-testid` 规范 |
| [logging-spec.md](logging-spec.md) | 日志与审计相关约束 |

## 如何选文档

- 只想知道流程顺序：先看 [../API_OPENAPI_FIXED_FLOW.md](../API_OPENAPI_FIXED_FLOW.md)
- 只想知道命令：先看 [commands.md](commands.md)
- 只改一个接口：先看 [../../backend/api/openapi/README.md](../../backend/api/openapi/README.md)
- 需要追溯历史审计或专项排查：去 [../archive/README.md](../archive/README.md)
