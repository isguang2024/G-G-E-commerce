# GGE 5.0 开发手册

本目录是项目固定用法的集中参考，所有手册随代码同步维护。

## 目录

| 文件 | 说明 |
|------|------|
| [commands.md](commands.md) | 常用命令速查（含生成链路一览） |
| [add-endpoint.md](add-endpoint.md) | 新增一条后端接口（3 步完成） |
| [api-auto-registration.md](api-auto-registration.md) | 接口自动入库机制原理与用法 |
| [permission-system.md](permission-system.md) | 权限系统结构与调试 |
| [database.md](database.md) | 数据库迁移、重置、seed |
| [permission-audit.md](permission-audit.md) | 权限审计 |
| [social-oauth-github.md](social-oauth-github.md) | GitHub OAuth 社交登录配置 |
| [social-oauth-manual-checklist.md](social-oauth-manual-checklist.md) | 社交登录手工验证清单 |

### Spec 目录说明

[`backend/api/openapi/README.md`](../../backend/api/openapi/README.md) — OpenAPI 多文件结构、编辑规则、生成链路、错误码体系的完整说明。

---

## 阅读顺序

初次接手项目：`commands.md` → `add-endpoint.md` → `permission-system.md`

只做业务开发：`add-endpoint.md` 即可

排查权限问题：`permission-system.md`
