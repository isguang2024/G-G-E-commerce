# GGE 5.0 开发手册

本目录是项目固定用法的集中参考，所有手册随代码同步维护。

## 目录

| 文件 | 说明 |
|------|------|
| [add-endpoint.md](add-endpoint.md) | 新增一条后端接口（完整 7 步） |
| [api-auto-registration.md](api-auto-registration.md) | 接口自动入库机制原理与用法 |
| [commands.md](commands.md) | 常用命令速查 |
| [permission-system.md](permission-system.md) | 权限系统结构与调试 |
| [database.md](database.md) | 数据库迁移、重置、seed |

## 阅读顺序

初次接手项目：`commands.md` → `add-endpoint.md` → `permission-system.md`

只做业务开发：`add-endpoint.md` 即可

排查权限问题：`permission-system.md`
